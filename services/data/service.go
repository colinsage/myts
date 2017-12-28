package data

import (
	"encoding"
	"encoding/binary"

	"fmt"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/influxdata/influxdb/query"
	"github.com/influxdata/influxdb/services/meta"
	"github.com/influxdata/influxdb/tsdb"
	"github.com/uber-go/zap"
	"github.com/influxdata/influxql"
	"context"
	"github.com/influxdata/influxdb/models"
)

// MaxMessageSize defines how large a message can be before we reject it
const MaxMessageSize = 1024 * 1024 * 1024 // 1GB

// MuxHeader is the header byte used in the TCP mux.
const MuxHeader = 2

// Statistics maintained by the cluster package
const (
	writeShardReq       = "writeShardReq"
	writeShardPointsReq = "writeShardPointsReq"
	writeShardFail      = "writeShardFail"

	createIteratorReq  = "createIteratorReq"
	createIteratorResp = "createIteratorResp"

	fieldDimensionsReq  = "fieldDimensionsReq"
	fieldDimensionsResp = "fieldDimensionsResp"

	mapTypeReq  = "mapTypeReq"
	mapTypeResp = "mapTypeResp"
)

type ServiceStatistics struct {
	WriteShardReq           int64
	WriteShardPointsReq     int64
	WriteShardFail          int64
	CreateIteratorReq       int64
	CreateIteratorResp      int64
	FieldDimensionsReq       int64
	FieldDimensionsResp       int64
	MapTypeReq           int64
	MapTypeResp         int64

}

// Service processes data received over raw TCP connections.
type Service struct {
	mu sync.RWMutex

	wg      sync.WaitGroup
	closing chan struct{}

	Listener net.Listener

	MetaClient interface {
		ShardOwner(shardID uint64) (string, string, *meta.ShardGroupInfo)
	}

	TSDBStore interface {
		ShardGroup(ids []uint64) tsdb.ShardGroup
		WriteToShard(shardID uint64, points []models.Point) error

		DeleteDatabase(name string) error
		DeleteRetentionPolicy(database, name string) error
		DeleteMeasurement(database, name string) error
		DeleteSeries(database string, sources []influxql.Source, condition influxql.Expr) error
		CreateShard(database, retentionPolicy string, shardID uint64, enabled bool) error

	}

	Config  *Config

	Logger  zap.Logger
	stats   *ServiceStatistics
}

// NewService returns a new instance of Service.
func NewService(c *Config) *Service {
	return &Service{
		closing: make(chan struct{}),
		Logger:  zap.New(zap.NullEncoder()),
		stats:   &ServiceStatistics{},
		Config:  c,
	}
}

// Open opens the network listener and begins serving requests.
func (s *Service) Open() error {

	//listener, err := net.Listen("tcp", s.Config.BindAddress)
	//if err != nil {
	//	return err
	//}
	//s.Listener = listener

	fmt.Println("Starting cluster service")
	// Begin serving conections.
	s.wg.Add(1)
	go s.serve()

	return nil
}

// SetLogger sets the internal logger to the logger passed in.
func (s *Service) WithLogger(log zap.Logger) {
	s.Logger = log.With(zap.String("service", "write"))
}

// serve accepts connections from the listener and handles them.
func (s *Service) serve() {
	defer s.wg.Done()

	for {
		// Check if the service is shutting down.
		select {
		case <-s.closing:
			return
		default:
		}

		// Accept the next connection.
		conn, err := s.Listener.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "connection closed") {
				s.Logger.Info(fmt.Sprintf("cluster service accept error: %s", err))
				return
			}
			s.Logger.Info(fmt.Sprintf("accept error: %s", err))
			continue
		}

		// Delegate connection handling to a separate goroutine.
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConn(conn)
		}()
	}
}

// Close shuts down the listener and waits for all connections to finish.
func (s *Service) Close() error {
	if s.Listener != nil {
		s.Listener.Close()
	}

	// Shut down all handlers.
	close(s.closing)
	s.wg.Wait()

	return nil
}

// handleConn services an individual TCP connection.
func (s *Service) handleConn(conn net.Conn) {
	// Ensure connection is closed when service is closed.
	closing := make(chan struct{})
	defer close(closing)
	go func() {
		select {
		case <-closing:
		case <-s.closing:
		}
		conn.Close()
	}()

	fmt.Printf("accept remote connection from %v\n", conn.RemoteAddr())
	//s.Logger.Info(fmt.Sprintf("accept remote connection from %v\n", conn.RemoteAddr()))
	defer func() {
		s.Logger.Info(fmt.Sprintf("close remote connection from %v\n", conn.RemoteAddr()))
	}()
	for {
		// Read type-length-value.
		typ, err := ReadType(conn)
		if err != nil {
			if strings.HasSuffix(err.Error(), "EOF") {
				return
			}
			s.Logger.Info(fmt.Sprint("unable to read type: %s", err))
			return
		}

		// Delegate message processing by type.
		switch typ {
		case writeShardRequestMessage:
			buf, err := ReadLV(conn)
			if err != nil {
				fmt.Println("unable to read length-value: %s", err)
				//s.Logger.Info(fmt.Sprint("unable to read length-value: %s", err))
				return
			}

			fmt.Println("read req sucess")
			err = s.processWriteShardRequest(buf)
			if err != nil {
				s.Logger.Info(fmt.Sprint("process write shard error: %s", err))
			}
			fmt.Println("write remote points sucess")
			s.writeShardResponse(conn, err)
		case executeStatementRequestMessage:
			buf, err := ReadLV(conn)
			if err != nil {
				s.Logger.Info(fmt.Sprint("unable to read length-value: %s", err))
				return
			}

			err = s.processExecuteStatementRequest(buf)
			if err != nil {
				s.Logger.Info(fmt.Sprint("process execute statement error: %s", err))
			}
			s.writeShardResponse(conn, err)
		case createIteratorRequestMessage:
			s.processCreateIteratorRequest(conn)
			return
		case fieldDimensionsRequestMessage:
			s.processFieldDimensionsRequest(conn)
			return
		case mapTypeRequestMessage:
			s.processMapTypeRequest(conn)
			return
		//case seriesKeysRequestMessage:
		//	s.processSeriesKeysRequest(conn)
		//	return
		default:
			s.Logger.Info(fmt.Sprintf("cluster service message type not found: %d", typ))
		}
	}
}

func (s *Service) processExecuteStatementRequest(buf []byte) error {
	// Unmarshal the request.
	var req ExecuteStatementRequest
	if err := req.UnmarshalBinary(buf); err != nil {
		return err
	}

	// Parse the InfluxQL statement.
	stmt, err := influxql.ParseStatement(req.Statement())
	if err != nil {
		return err
	}

	return s.executeStatement(stmt, req.Database())
}

func (s *Service) executeStatement(stmt influxql.Statement, database string) error {
	switch t := stmt.(type) {
	case *influxql.DropDatabaseStatement:
		return s.TSDBStore.DeleteDatabase(t.Name)
	case *influxql.DropMeasurementStatement:
		return s.TSDBStore.DeleteMeasurement(database, t.Name)
	case *influxql.DropSeriesStatement:
		return s.TSDBStore.DeleteSeries(database, t.Sources, t.Condition)
	case *influxql.DropRetentionPolicyStatement:
		return s.TSDBStore.DeleteRetentionPolicy(database, t.Name)
	default:
		return fmt.Errorf("%q should not be executed across a cluster", stmt.String())
	}
}

func (s *Service) processWriteShardRequest(buf []byte) error {
	// Build request
	var req WriteShardRequest
	if err := req.UnmarshalBinary(buf); err != nil {
		return err
	}
	fmt.Println("unmarshal sucess.")
	points := req.Points()
	err := s.TSDBStore.WriteToShard(req.ShardID(), points)

	// We may have received a write for a shard that we don't have locally because the
	// sending node may have just created the shard (via the metastore) and the write
	// arrived before the local store could create the shard.  In this case, we need
	// to check the metastore to determine what database and retention policy this
	// shard should reside within.
	if err == tsdb.ErrShardNotFound {
		db, rp := req.Database(), req.RetentionPolicy()
		if db == "" || rp == "" {
			s.Logger.Info(fmt.Sprintf("drop write request: shard=%d. no database or rentention policy received", req.ShardID()))
			return nil
		}

		err = s.TSDBStore.CreateShard(req.Database(), req.RetentionPolicy(), req.ShardID(), true)
		if err != nil {
			return fmt.Errorf("create shard %d: %s", req.ShardID(), err)
		}

		err = s.TSDBStore.WriteToShard(req.ShardID(), points)
		if err != nil {
			return fmt.Errorf("write shard %d: %s", req.ShardID(), err)
		}
	}

	if err != nil {
		return fmt.Errorf("write shard %d: %s", req.ShardID(), err)
	}

	return nil
}

func (s *Service) writeShardResponse(w io.Writer, e error) {
	// Build response.
	var resp WriteShardResponse
	if e != nil {
		resp.SetCode(1)
		resp.SetMessage(e.Error())
	} else {
		resp.SetCode(0)
	}

	// Marshal response to binary.
	buf, err := resp.MarshalBinary()
	if err != nil {
		s.Logger.Info(fmt.Sprintf("error marshalling shard response: %s", err))
		return
	}

	// Write to connection.
	if err := WriteTLV(w, writeShardResponseMessage, buf); err != nil {
		s.Logger.Info(fmt.Sprintf("write shard response error: %s", err))
	}
	fmt.Println("write to write points resp sucess.")
}

func (s *Service) processCreateIteratorRequest(conn net.Conn) {
	defer conn.Close()

	var itr query.Iterator
	if err := func() error {
		// Parse request.
		var req CreateIteratorRequest
		if err := DecodeLV(conn, &req); err != nil {
			return err
		}

		sg := s.TSDBStore.ShardGroup(req.ShardIDs)

		i, err := sg.CreateIterator(context.Background(), req.Measurement, req.Opt)
		if err != nil {
			return err
		}
		itr = i
		return nil
	}(); err != nil {
		itr.Close()
		s.Logger.Info(fmt.Sprintf("error reading CreateIterator request: %s", err))
		EncodeTLV(conn, createIteratorResponseMessage, &CreateIteratorResponse{Err: err})
		return
	}

	// Encode success response.
	if err := EncodeTLV(conn, createIteratorResponseMessage, &CreateIteratorResponse{}); err != nil {
		s.Logger.Info(fmt.Sprintf("error writing CreateIterator response: %s", err))
		return
	}

	// Exit if no iterator was produced.
	if itr == nil {
		return
	}

	// Stream iterator to connection.
	if err := query.NewIteratorEncoder(conn).EncodeIterator(itr); err != nil {
		s.Logger.Info(fmt.Sprintf("error encoding CreateIterator iterator: %s", err))
		return
	}
}

func (s *Service) processFieldDimensionsRequest(conn net.Conn) {
	fields := make(map[string]influxql.DataType)
	dimensions := make(map[string]struct{})
	if err := func() error {
		// Parse request.
		var req FieldDimensionsRequest
		if err := DecodeLV(conn, &req); err != nil {
			return err
		}

		// Collect iterator creators for each shard.
		//ics := make([]influxql.IteratorCreator, 0, len(req.ShardIDs))
		//for _, shardID := range req.ShardIDs {
		//	ic := s.TSDBStore.ShardIteratorCreator(shardID)
		//	if ic == nil {
		//		return nil
		//	}
		//	ics = append(ics, ic)
		//}
		sg := s.TSDBStore.ShardGroup(req.ShardIDs)
		// Generate a single iterator from all shards.
		//f, d, err := influxql.IteratorCreators(ics).FieldDimensions(req.Sources)
		measurements := []string{req.Measurement}
		f, d, err := sg.FieldDimensions(measurements)
		if err != nil {
			return err
		}
		fields, dimensions = f, d

		return nil
	}(); err != nil {
		s.Logger.Info(fmt.Sprintf("error reading FieldDimensions request: %s", err))
		EncodeTLV(conn, fieldDimensionsResponseMessage, &FieldDimensionsResponse{Err: err})
		return
	}

	// Encode success response.
	if err := EncodeTLV(conn, fieldDimensionsResponseMessage, &FieldDimensionsResponse{
		Fields:     fields,
		Dimensions: dimensions,
	}); err != nil {
		s.Logger.Info(fmt.Sprintf("error writing FieldDimensions response: %s", err))
		return
	}
}

func (s *Service) processMapTypeRequest(conn net.Conn) {
	mapType := influxql.Unknown
	if err := func() error {
		// Parse request.
		var req MapTypeRequest
		if err := DecodeLV(conn, &req); err != nil {
			return err
		}

		sg := s.TSDBStore.ShardGroup(req.ShardIDs)
		t := sg.MapType(req.Measurement, req.Field)
		if mapType.LessThan(t) {
			mapType = t
		}

		return nil
	}(); err != nil {
		s.Logger.Info(fmt.Sprintf("error reading FieldDimensions request: %s", err))
		EncodeTLV(conn, mapTypeResponseMessage, &MapTypeResponse{Err: err})
		return
	}

	// Encode success response.
	if err := EncodeTLV(conn, mapTypeResponseMessage, &MapTypeResponse{
		Type: int(mapType),
	}); err != nil {
		s.Logger.Info(fmt.Sprintf("error writing FieldDimensions response: %s", err))
		return
	}
}
//func (s *Service) processSeriesKeysRequest(conn net.Conn) {
//	var seriesList influxql.SeriesList
//	if err := func() error {
//		// Parse request.
//		var req SeriesKeysRequest
//		if err := DecodeLV(conn, &req); err != nil {
//			return err
//		}
//
//		// Collect iterator creators for each shard.
//		ics := make([]influxql.IteratorCreator, 0, len(req.ShardIDs))
//		for _, shardID := range req.ShardIDs {
//			ic := s.TSDBStore.ShardIteratorCreator(shardID)
//			if ic == nil {
//				return nil
//			}
//			ics = append(ics, ic)
//		}
//
//		// Generate a single iterator from all shards.
//		a, err := influxql.IteratorCreators(ics).SeriesKeys(req.Opt)
//		if err != nil {
//			return err
//		}
//		seriesList = a
//
//		return nil
//	}(); err != nil {
//		s.Logger.Printf("error reading SeriesKeys request: %s", err)
//		EncodeTLV(conn, seriesKeysResponseMessage, &SeriesKeysResponse{Err: err})
//		return
//	}
//
//	// Encode success response.
//	if err := EncodeTLV(conn, seriesKeysResponseMessage, &SeriesKeysResponse{
//		SeriesList: seriesList,
//	}); err != nil {
//		s.Logger.Printf("error writing SeriesKeys response: %s", err)
//		return
//	}
//}

// ReadTLV reads a type-length-value record from r.
func ReadTLV(r io.Reader) (byte, []byte, error) {
	typ, err := ReadType(r)
	if err != nil {
		return 0, nil, err
	}

	buf, err := ReadLV(r)
	if err != nil {
		return 0, nil, err
	}
	return typ, buf, err
}

// ReadType reads the type from a TLV record.
func ReadType(r io.Reader) (byte, error) {
	var typ [1]byte
	if _, err := io.ReadFull(r, typ[:]); err != nil {
		return 0, fmt.Errorf("read message type: %s", err)
	}
	return typ[0], nil
}

// ReadLV reads the length-value from a TLV record.
func ReadLV(r io.Reader) ([]byte, error) {
	// Read the size of the message.
	var sz int64
	if err := binary.Read(r, binary.BigEndian, &sz); err != nil {
		return nil, fmt.Errorf("read message size: %s", err)
	}

	if sz >= MaxMessageSize {
		return nil, fmt.Errorf("max message size of %d exceeded: %d", MaxMessageSize, sz)
	}

	// Read the value.
	buf := make([]byte, sz)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("read message value: %s", err)
	}

	return buf, nil
}

// WriteTLV writes a type-length-value record to w.
func WriteTLV(w io.Writer, typ byte, buf []byte) error {
	if err := WriteType(w, typ); err != nil {
		return err
	}
	if err := WriteLV(w, buf); err != nil {
		return err
	}
	return nil
}

// WriteType writes the type in a TLV record to w.
func WriteType(w io.Writer, typ byte) error {
	if _, err := w.Write([]byte{typ}); err != nil {
		return fmt.Errorf("write message type: %s", err)
	}
	return nil
}

// WriteLV writes the length-value in a TLV record to w.
func WriteLV(w io.Writer, buf []byte) error {
	// Write the size of the message.
	if err := binary.Write(w, binary.BigEndian, int64(len(buf))); err != nil {
		return fmt.Errorf("write message size: %s", err)
	}

	// Write the value.
	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("write message value: %s", err)
	}
	return nil
}

// EncodeTLV encodes v to a binary format and writes the record-length-value record to w.
func EncodeTLV(w io.Writer, typ byte, v encoding.BinaryMarshaler) error {
	if err := WriteType(w, typ); err != nil {
		return err
	}
	if err := EncodeLV(w, v); err != nil {
		return err
	}
	return nil
}

// EncodeLV encodes v to a binary format and writes the length-value record to w.
func EncodeLV(w io.Writer, v encoding.BinaryMarshaler) error {
	buf, err := v.MarshalBinary()
	if err != nil {
		return err
	}

	if err := WriteLV(w, buf); err != nil {
		return err
	}
	return nil
}

// DecodeTLV reads the type-length-value record from r and unmarshals it into v.
func DecodeTLV(r io.Reader, v encoding.BinaryUnmarshaler) (typ byte, err error) {
	typ, err = ReadType(r)
	if err != nil {
		return 0, err
	}
	if err := DecodeLV(r, v); err != nil {
		return 0, err
	}
	return typ, nil
}

// DecodeLV reads the length-value record from r and unmarshals it into v.
func DecodeLV(r io.Reader, v encoding.BinaryUnmarshaler) error {
	buf, err := ReadLV(r)
	if err != nil {
		return err
	}

	if err := v.UnmarshalBinary(buf); err != nil {
		return err
	}
	return nil
}
