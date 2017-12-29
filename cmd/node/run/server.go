package run

import (
	"net"

	"fmt"

	"time"
	"log"
	"github.com/influxdata/influxdb/coordinator"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/influxdata/influxdb/monitor"
	"github.com/influxdata/influxdb/models"
	"github.com/influxdata/influxdb/services/retention"
	"github.com/influxdata/influxdb/services/httpd"
	"github.com/influxdata/influxdb/services/meta"

	// Initialize the engine & index packages
	_ "github.com/influxdata/influxdb/tsdb/engine"
	_ "github.com/influxdata/influxdb/tsdb/index"
	"github.com/influxdata/influxdb/tcp"

	thisMeta "github.com/colinsage/myts/services/meta"
	"github.com/colinsage/myts/services/data"


	"github.com/influxdata/influxdb/query"
	"github.com/uber-go/zap"
	"os"
	"io"
	"strconv"
)


var startTime time.Time

func init() {
	startTime = time.Now().UTC()
}

// BuildInfo represents the build details for the server code.
type BuildInfo struct {
	Version string
	Commit  string
	Branch  string
	Time    string
}

type Server struct {
	buildInfo BuildInfo

	err     chan error
	closing chan struct{}

	BindAddress string
	Listener    net.Listener

	Logger zap.Logger

	MetaClient  *thisMeta.Client
	MetaService *thisMeta.Service

	Cluster *data.Service

	Services []Service

	TSDBStore     *tsdb.Store
	QueryExecutor *query.QueryExecutor
	PointsWriter  *data.PointsWriter
	ShardWriter   *data.ShardWriter

	// joinPeers are the metaservers specified at run time to join this server to
	joinPeers []string

	// metaUseTLS specifies if we should use a TLS connection to the meta servers
	metaUseTLS bool

	// httpAPIAddr is the host:port combination for the main HTTP API for querying and writing data
	httpAPIAddr string

	// httpUseTLS specifies if we should use a TLS connection to the http servers
	httpUseTLS bool

	// tcpAddr is the host:port combination for the TCP listener that services mux onto
	tcpAddr string

	config *Config

	Monitor *monitor.Monitor

	Node *meta.NodeInfo
}

// Statistics returns statistics for the services running in the Server.
func (s *Server) Statistics(tags map[string]string) []models.Statistic {
	var statistics []models.Statistic
	statistics = append(statistics, s.QueryExecutor.Statistics(tags)...)
	statistics = append(statistics, s.TSDBStore.Statistics(tags)...)
	statistics = append(statistics, s.PointsWriter.Statistics(tags)...)
	for _, srv := range s.Services {
		if m, ok := srv.(monitor.Reporter); ok {
			statistics = append(statistics, m.Statistics(tags)...)
		}
	}
	return statistics
}

func NewServer(c *Config, buildInfo *BuildInfo) (*Server, error) {
	s := &Server{
		buildInfo: *buildInfo,
		err:       make(chan error),
		closing:   make(chan struct{}),
		Logger: zap.New(
			zap.NewTextEncoder(),
			zap.Output(os.Stderr),
		),

		MetaClient: thisMeta.NewClient(),

		joinPeers:         c.Meta.JoinPeers,
		metaUseTLS:        c.Meta.HTTPSEnabled,

		BindAddress:  c.BindAddress,

		config: c,
	}

	//s.Monitor = monitor.New(s, c.Monitor)
	//s.config.registerDiagnostics(s.Monitor)



	s.Node = &meta.NodeInfo{}
	if c.Global.MetaEnabled{
		s.MetaService = thisMeta.NewService(c.Meta)
		s.MetaService.Version = s.buildInfo.Version
	}

	if c.Global.DataEnabled {
		s.TSDBStore = tsdb.NewStore(c.Data.Dir)
		s.TSDBStore.EngineOptions.Config = c.Data

		// Copy TSDB configuration.
		s.TSDBStore.EngineOptions.EngineVersion = c.Data.Engine
		s.TSDBStore.EngineOptions.IndexVersion = c.Data.Index


		s.ShardWriter = data.NewShardWriter(data.DefaultWriteTimeout, data.DefaultMaxRemoteWriteConnections)
		s.ShardWriter.MetaClient = s.MetaClient


		// Initialize points writer.
		s.PointsWriter = data.NewPointsWriter()
		s.PointsWriter.WriteTimeout = time.Duration(c.DataExt.WriteTimeout)
		s.PointsWriter.TSDBStore = s.TSDBStore
		s.PointsWriter.Node = s.Node
		s.PointsWriter.ShardWriter = s.ShardWriter

		s.QueryExecutor = query.NewQueryExecutor()
		s.QueryExecutor.StatementExecutor = &coordinator.StatementExecutor{
			MetaClient:  s.MetaClient,
			TaskManager: s.QueryExecutor.TaskManager,
			TSDBStore:   coordinator.LocalTSDBStore{Store: s.TSDBStore},
			ShardMapper: &data.DistributedShardMapper{
				MetaClient: s.MetaClient,
				TSDBStore:  coordinator.LocalTSDBStore{Store: s.TSDBStore},
				Node: s.Node,
				Timeout: data.DefaultWriteTimeout, //TODO
			},
			Monitor:           s.Monitor,
			PointsWriter:      s.PointsWriter,
			MaxSelectPointN:   c.DataExt.MaxSelectPointN,
			MaxSelectSeriesN:  c.DataExt.MaxSelectSeriesN,
			MaxSelectBucketsN: c.DataExt.MaxSelectBucketsN,
		}
		s.QueryExecutor.TaskManager.QueryTimeout = time.Duration(c.DataExt.QueryTimeout)
		s.QueryExecutor.TaskManager.LogQueriesAfter = time.Duration(c.DataExt.LogQueriesAfter)
		s.QueryExecutor.TaskManager.MaxConcurrentQueries = c.DataExt.MaxConcurrentQueries

		s.Cluster = data.NewService(s.config.DataExt)
		s.Cluster.TSDBStore = s.TSDBStore
		s.Cluster.MetaClient = s.MetaClient
		s.Cluster.Config = s.config.DataExt


		// Initialize the monitor
		//s.Monitor.Version = s.buildInfo.Version
		//s.Monitor.Commit = s.buildInfo.Commit
		//s.Monitor.Branch = s.buildInfo.Branch
		//s.Monitor.BuildTime = s.buildInfo.Time
		//s.Monitor.PointsWriter = (*monitorPointsWriter)(s.PointsWriter)
	}

	//tsdb.NewInmemIndex = func(name string) (interface{}, error) { return nil, nil }
	return s, nil
}

// Err returns an error channel that multiplexes all out of band errors received from all services.
func (s *Server) Err() <-chan error { return s.err }

// Open opens the meta and data store and all services.
func (s *Server) Open() error {
	// Start profiling, if set.
	//startProfile(s.CPUProfile, s.MemProfile)

	// Open shared TCP connection.
	s.Logger.Info("bind_addr:"+ strconv.Itoa(len(s.config.BindAddress)))
	//TODO
	if s.config.Global.DataEnabled  && len(s.config.BindAddress) != 0{
		tcpAddr := s.config.Global.Hostname + s.config.BindAddress
		ln, err := net.Listen("tcp", tcpAddr)
		s.Logger.Info("listen tcp at : " +  tcpAddr)

		if err != nil {
			return fmt.Errorf("listen: %s", err)
		}
		s.Listener = ln

		// Multiplex listener.
		mux := tcp.NewMux()
		go mux.Serve(ln)

		s.Cluster.Listener = mux.Listen(data.MuxHeader)
		s.config.DataExt.BindAddress = tcpAddr
	}

	if s.MetaService != nil {
		//s.MetaService.RaftListener = mux.Listen(thisMeta.MuxHeader)
		// Open meta service.
		if s.config.Meta.LoggingEnabled {
			s.MetaService.WithLogger(s.Logger)
		}
		if err := s.MetaService.Open(); err != nil {
			return fmt.Errorf("open meta service: %s", err)
		}
		go s.monitorErrorChan(s.MetaService.Err())
	}


	// initialize MetaClient.
	if err := s.initializeMetaClient(); err != nil {
		return err
	}

	// Configure logging for all services and clients.
	if s.config.Meta.LoggingEnabled {
		s.MetaClient.WithLogger(s.Logger)
	}

	if s.config.Global.DataEnabled {

		s.TSDBStore.WithLogger(s.Logger)
		if s.config.Data.QueryLogEnabled {
			s.QueryExecutor.WithLogger(s.Logger)
		}

		s.PointsWriter.WithLogger(s.Logger)
		s.ShardWriter.WithLogger(s.Logger)

		// Append services.
		//s.appendMonitorService()
		//s.appendPrecreatorService(s.config.Precreator)

		s.config.HTTPD.BindAddress = s.config.Global.Hostname + s.config.HTTPD.BindAddress
		s.appendHTTPDService(s.config.HTTPD)
		s.appendRetentionPolicyService(s.config.Retention)

		s.PointsWriter.MetaClient = s.MetaClient
		//s.Monitor.MetaClient = s.MetaClient

		for _, svc := range s.Services {
			svc.WithLogger(s.Logger)
		}


		// Open TSDB store.
		if err := s.TSDBStore.Open(); err != nil {
			return fmt.Errorf("open tsdb store: %s", err)
		}

		// Open the points writer service
		if err := s.PointsWriter.Open(); err != nil {
			return fmt.Errorf("open points writer: %s", err)
		}


		if err := s.Cluster.Open(); err != nil {
			return fmt.Errorf("open cluster: %s", err)
		}

		for _, service := range s.Services {
			if err := service.Open(); err != nil {
				return fmt.Errorf("open service: %s", err)
			}
		}

	}

	return nil
}

func (s *Server) appendMonitorService() {
	s.Services = append(s.Services, s.Monitor)
}

func (s *Server) appendRetentionPolicyService(c retention.Config) {
	if !c.Enabled {
		return
	}
	srv := retention.NewService(c)
	srv.MetaClient = s.MetaClient
	srv.TSDBStore = s.TSDBStore
	s.Services = append(s.Services, srv)
}

func (s *Server) appendHTTPDService(c httpd.Config) {
	if !c.Enabled {
		return
	}
	srv := httpd.NewService(c)
	srv.Handler.MetaClient = s.MetaClient
	//srv.Handler.QueryAuthorizer = thisMeta.NewQueryAuthorizer(s.MetaClient)
	//srv.Handler.WriteAuthorizer = thisMeta.NewWriteAuthorizer(s.MetaClient)
	srv.Handler.QueryExecutor = s.QueryExecutor
	srv.Handler.Monitor = s.Monitor
	srv.Handler.PointsWriter = s.PointsWriter
	srv.Handler.Version = s.buildInfo.Version

	s.Services = append(s.Services, srv)
}

// Close shuts down the meta and data stores and all services.
func (s *Server) Close() error {
	//stopProfile()

	// Close the listener first to stop any new connections
	if s.Listener != nil {
		s.Listener.Close()
	}

	// Finally close the meta-store since everything else depends on it
	if s.MetaService != nil {
		s.MetaService.Close()
	}
	if s.MetaClient != nil {
		s.MetaClient.Close()
	}

	close(s.closing)
	return nil
}

// monitorErrorChan reads an error channel and resends it through the server.
func (s *Server) monitorErrorChan(ch <-chan error) {
	for {
		select {
		case err, ok := <-ch:
			if !ok {
				return
			}
			s.err <- err
		case <-s.closing:
			return
		}
	}
}

// initializeMetaClient will set the MetaClient and join the node to the cluster if needed
func (s *Server) initializeMetaClient() error {
	 //It's the first time starting up and we need to either join
	 //the cluster or initialize this node as the first member
	if len(s.joinPeers) == 0 {
		// start up a new single node cluster
		if s.MetaService == nil {
			return fmt.Errorf("server not set to join existing cluster must run also as a meta node")
		}
		s.MetaClient.SetMetaServers([]string{s.MetaService.HTTPAddr()})
		s.MetaClient.SetTLS(s.metaUseTLS)
	} else {
		// join this node to the cluster
		s.MetaClient.SetMetaServers(s.joinPeers)
		s.MetaClient.SetTLS(s.metaUseTLS)
	}
	if err := s.MetaClient.Open(); err != nil {
		return err
	}

	if s.config.Global.DataEnabled{
		//if _, err := s.MetaClient.DataNode(s.Node.ID); err == nil {
		//	return nil
		//}

		httpAddr := s.config.Global.Hostname + s.config.HTTPD.BindAddress
		tcpAddr := s.config.Global.Hostname + s.config.BindAddress
		node, err := s.MetaClient.CreateDataNode(httpAddr, tcpAddr)
		for err != nil {
			log.Printf("Unable to create data node. retry in 1s: %s", err.Error())
			time.Sleep(time.Second)
			_, err = s.MetaClient.CreateDataNode(httpAddr, tcpAddr)
		}

		log.Println(node.ID, node.Host, node.TCPHost)
		s.Node.TCPHost = node.TCPHost
		s.Node.Host = node.Host
		s.Node.ID = node.ID

	}
	return nil
}

// HTTPAddr returns the HTTP address used by other nodes for HTTP queries and writes.
func (s *Server) HTTPAddr() string {
	return s.remoteAddr(s.httpAPIAddr)
}

// TCPAddr returns the TCP address used by other nodes for cluster communication.
func (s *Server) TCPAddr() string {
	return s.remoteAddr(s.tcpAddr)
}

func (s *Server) remoteAddr(addr string) string {
	hostname := s.config.Global.Hostname
	if hostname == "" {
		hostname = thisMeta.DefaultHostname
	}
	remote, err := thisMeta.DefaultHost(hostname, addr)
	if err != nil {
		return addr
	}
	return remote
}

// MetaServers returns the meta node HTTP addresses used by this server.
func (s *Server) MetaServers() []string {
	return s.MetaClient.MetaServers()
}

// Service represents a service attached to the server.
type Service interface {
	WithLogger(log zap.Logger)
	Open() error
	Close() error
}


type tcpaddr struct{ host string }

func (a *tcpaddr) Network() string { return "tcp" }
func (a *tcpaddr) String() string  { return a.host }


type monitorPointsWriter data.PointsWriter

func (pw *monitorPointsWriter) WritePoints(database, retentionPolicy string, points models.Points) error {
	return (*data.PointsWriter)(pw).WritePointsPrivileged(database, retentionPolicy, models.ConsistencyLevelAny, points)
}

// SetLogOutput sets the logger used for all messages. It must not be called
// after the Open method has been called.
func (s *Server) SetLogOutput(w io.Writer) {
	s.Logger = zap.New(zap.NewTextEncoder(), zap.Output(zap.AddSync(w)))
}