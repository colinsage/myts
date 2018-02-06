package meta

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/influxdata/influxdb"

	"github.com/gogo/protobuf/proto"
	"github.com/hashicorp/raft"
	"github.com/colinsage/myts/services/meta/internal"
	"github.com/influxdata/influxdb/services/meta"
	"github.com/uber-go/zap"
	"github.com/colinsage/myts/utils"
	"google.golang.org/grpc"

	pb "github.com/colinsage/myts/grpc/proto"

	"context"
)

// Retention policy settings.
const (
	autoCreateRetentionPolicyName   = "default"
	autoCreateRetentionPolicyPeriod = 0

	// maxAutoCreatedRetentionPolicyReplicaN is the maximum replication factor that will
	// be set for auto-created retention policies.
	maxAutoCreatedRetentionPolicyReplicaN = 2

	defaultReplicaN = 2
)

// Raft configuration.
const (
	raftListenerStartupTimeout = time.Second
)

type store struct {
	mu      sync.RWMutex
	closing chan struct{}

	config      *Config
	data        *Data
	raftState   *raftState
	dataChanged chan struct{}
	path        string
	opened      bool
	logger      zap.Logger

	raftAddr string
	httpAddr string

	node *influxdb.Node
}

// newStore will create a new metastore with the passed in config
func newStore(c *Config, httpAddr, raftAddr string) *store {
	internalData := meta.Data{
		Index: 1,
	}
	s := store{
		data: &Data{
			Data: internalData,
		},
		closing:     make(chan struct{}),
		dataChanged: make(chan struct{}),
		path:        c.Dir,
		config:      c,
		httpAddr:    httpAddr,
		raftAddr:    raftAddr,
		logger:  zap.New(zap.NullEncoder()),
	}

	return &s
}

func (s *store) WithLogger(log zap.Logger) {
	s.logger = log.With(zap.String("service", "meta_store"))
}
// open opens and initializes the raft store.
func (s *store) open(raftln net.Listener) error {
	s.logger.Info(fmt.Sprintf("Using data dir: %v", s.path))

	joinPeers, err := s.filterAddr(s.config.JoinPeers, s.httpAddr)
	if err != nil {
		return err
	}
	joinPeers = s.config.JoinPeers

	var initializePeers []string
	if len(joinPeers) > 0 {
		c := NewClient()
		c.SetMetaServers(joinPeers)
		c.SetTLS(s.config.HTTPSEnabled)
		for {
			peers := c.peers()
			if !Peers(peers).Contains(s.raftAddr) {
				peers = append(peers, s.raftAddr)
			}

			s.logger.Info(fmt.Sprintf("len : %d, %d \r\n" ,len(s.config.JoinPeers),len(peers)))
			if len(s.config.JoinPeers)-len(peers) == 0 {
				initializePeers = peers
				break
			}

			if len(peers) > len(s.config.JoinPeers) {
				s.logger.Info(fmt.Sprintf("waiting for join peers to match config specified. found %v, config specified %v", peers, s.config.JoinPeers))
			} else {
				s.logger.Info(fmt.Sprintf("Waiting for %d join peers.  Have %v. Asking nodes: %v", len(s.config.JoinPeers)-len(peers), peers, joinPeers))
			}
			time.Sleep(time.Second)
		}
	}

	if err := s.setOpen(); err != nil {
		return err
	}

	// Create the root directory if it doesn't already exist.
	if err := os.MkdirAll(s.path, 0777); err != nil {
		return fmt.Errorf("mkdir all: %s", err)
	}

	// Open the raft store.
	if err := s.openRaft(initializePeers, raftln); err != nil {
		return fmt.Errorf("raft: %s", err)
	}

	s.logger.Info(fmt.Sprintf("open raft done. %d", len(joinPeers)))
	if len(joinPeers) > 0 {
		c := NewClient()
		c.SetMetaServers(joinPeers)
		c.SetTLS(s.config.HTTPSEnabled)
		if err := c.Open(); err != nil {
			return err
		}
		defer c.Close()

		_, err := c.JoinMetaServer(s.httpAddr, s.raftAddr)
		if err != nil {
			return err
		}
	}

	// Wait for a leader to be elected so we know the raft log is loaded
	// and up to date
	if err := s.waitForLeader(0); err != nil {
		return err
	}

	// Make sure this server is in the list of metanodes
	peers, err := s.raftState.peers()
	if err != nil {
		return err
	}
	s.logger.Info(fmt.Sprintf("peers count= %d", len(peers)))
	if len(peers) <= 1 {
		// we have to loop here because if the hostname has changed
		// raft will take a little bit to normalize so that this host
		// will be marked as the leader
		for {
			err := s.setMetaNode(s.httpAddr, s.raftAddr)
			if err == nil {
				break
			}else{
				s.logger.Error(err.Error())
			}
			time.Sleep(100 * time.Millisecond)
		}
	}

	go func() { s.rebalance(s.closing) }()
	s.logger.Info("open store done")
	return nil
}

func (s *store) setOpen() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check if store has already been opened.
	if s.opened {
		return ErrStoreOpen
	}
	s.opened = true
	return nil
}

// peers returns the raft peers known to this store
func (s *store) peers() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.raftState == nil {
		s.logger.Error("[ERROR] rafte state is nil")
		return []string{s.raftAddr}
	}
	peers, err := s.raftState.peers()
	if err != nil || peers == nil || len(peers) == 0 {
		return []string{s.raftAddr}
	}
	return peers
}

func (s *store) filterAddr(addrs []string, filter string) ([]string, error) {
	host, port, err := net.SplitHostPort(filter)
	if err != nil {
		return nil, err
	}

	ip, err := net.ResolveIPAddr("ip", host)
	if err != nil {
		return nil, err
	}

	var joinPeers []string
	for _, addr := range addrs {
		joinHost, joinPort, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		joinIp, err := net.ResolveIPAddr("ip", joinHost)
		if err != nil {
			return nil, err
		}

		// Don't allow joining ourselves
		if ip.String() == joinIp.String() && port == joinPort {
			continue
		}
		joinPeers = append(joinPeers, addr)
	}
	return joinPeers, nil
}

func (s *store) openRaft(initializePeers []string, raftln net.Listener) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	rs := newRaftState(s.config, s.raftAddr)
	rs.WithLogger(s.logger)

	rs.path = s.path

	if err := rs.open(s, raftln, initializePeers); err != nil {
		return err
	}
	s.raftState = rs

	return nil
}

func (s *store) close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.closing:
		// already closed
		return nil
	default:
		close(s.closing)
		return s.raftState.close()
	}
}

func (s *store) snapshot() (*Data, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Clone(), nil
}

// afterIndex returns a channel that will be closed to signal
// the caller when an updated snapshot is available.
func (s *store) afterIndex(index uint64) <-chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index < s.data.Data.Index {
		// Client needs update so return a closed channel.
		ch := make(chan struct{})
		close(ch)
		return ch
	}

	return s.dataChanged
}

// WaitForLeader sleeps until a leader is found or a timeout occurs.
// timeout == 0 means to wait forever.
func (s *store) waitForLeader(timeout time.Duration) error {
	// Begin timeout timer.
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	// Continually check for leader until timeout.
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	s.logger.Info("waitForLeader start")
	for {
		select {
		case <-s.closing:
			return errors.New("closing")
		case <-timer.C:
			if timeout != 0 {
				return errors.New("timeout")
			}
		case <-ticker.C:
			if s.leader() != "" {
				s.logger.Info("waitForLeader done")
				return nil
			}
		}
	}
}

// isLeader returns true if the store is currently the leader.
func (s *store) isLeader() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.raftState == nil {
		return false
	}
	return s.raftState.raft.State() == raft.Leader
}

// leader returns what the store thinks is the current leader. An empty
// string indicates no leader exists.
func (s *store) leader() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.raftState == nil || s.raftState.raft == nil {
		return ""
	}

	s.logger.Info("query leader. " + string(s.raftState.raft.Leader()))
	return string(s.raftState.raft.Leader())
}

// leaderHTTP returns the HTTP API connection info for the meta node
// that is the raft leader
func (s *store) leaderHTTP() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.raftState == nil {
		return ""
	}
	l := s.raftState.raft.Leader()

	for _, n := range s.data.MetaNodes {
		if n.TCPHost == string(l) {
			return n.Host
		}
	}

	return ""
}

// otherMetaServersHTTP will return the HTTP bind addresses of the other
// meta servers in the cluster
func (s *store) otherMetaServersHTTP() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var a []string
	for _, n := range s.data.MetaNodes {
		if n.TCPHost != s.raftAddr {
			a = append(a, n.Host)
		}
	}
	return a
}

// index returns the current store index.
func (s *store) index() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Data.Index
}

// apply applies a command to raft.
func (s *store) apply(b []byte) error {
	if s.raftState == nil {
		return fmt.Errorf("store not open")
	}
	return s.raftState.apply(b)
}

// join adds a new server to the metaservice and raft
func (s *store) join(n *meta.NodeInfo) (*meta.NodeInfo, error) {
	s.mu.RLock()
	if s.raftState == nil {
		s.mu.RUnlock()
		return nil, fmt.Errorf("store not open")
	}
	if err := s.raftState.addPeer(n.TCPHost); err != nil {
		s.mu.RUnlock()
		return nil, err
	}
	s.mu.RUnlock()

	if err := s.createMetaNode(n.Host, n.TCPHost); err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, node := range s.data.MetaNodes {
		if node.TCPHost == n.TCPHost && node.Host == n.Host {
			return &node, nil
		}
	}
	return nil, ErrNodeNotFound
}

// leave removes a server from the metaservice and raft
func (s *store) leave(n *meta.NodeInfo) error {
	return s.raftState.removePeer(n.TCPHost)
}

// createMetaNode is used by the join command to create the metanode int
// the metastore
func (s *store) createMetaNode(addr, raftAddr string) error {
	val := &internal.CreateMetaNodeCommand{
		HTTPAddr: proto.String(addr),
		TCPAddr:  proto.String(raftAddr),
		Rand:     proto.Uint64(uint64(rand.Int63())),
	}
	t := internal.Command_CreateMetaNodeCommand
	cmd := &internal.Command{Type: &t}
	if err := proto.SetExtension(cmd, internal.E_CreateMetaNodeCommand_Command, val); err != nil {
		panic(err)
	}

	b, err := proto.Marshal(cmd)
	if err != nil {
		return err
	}

	return s.apply(b)
}

// setMetaNode is used when the raft group has only a single peer. It will
// either create a metanode or update the information for the one metanode
// that is there. It's used because hostnames can change
func (s *store) setMetaNode(addr, raftAddr string) error {
	val := &internal.SetMetaNodeCommand{
		HTTPAddr: proto.String(addr),
		TCPAddr:  proto.String(raftAddr),
		Rand:     proto.Uint64(uint64(rand.Int63())),
	}
	t := internal.Command_SetMetaNodeCommand
	cmd := &internal.Command{Type: &t}
	if err := proto.SetExtension(cmd, internal.E_SetMetaNodeCommand_Command, val); err != nil {
		panic(err)
	}

	b, err := proto.Marshal(cmd)
	if err != nil {
		return err
	}
	s.logger.Info("set meta node apply begin")
	return s.apply(b)
}

// timer task to check shard rebalance

func (s *store) rebalance( closing <-chan struct{}) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-closing:
			return
		case <-t.C:
			isNeed, tasks := s.needRebalance()
			if isNeed {
				s.logger.Info(fmt.Sprintf("need reblance for %d shards", len(tasks)))
				go func() {
					s.doRebalance(tasks)
				}()
			}
		}
	}
}

type ShardRestoreTask struct {
	Database string
	RetentionPolicy string
	ShardID uint64
	CurrentOwners []meta.ShardOwner
	NewOwners []meta.ShardOwner
	NewOwnerAddress map[uint64]string
}

func (s *store) needRebalance() (bool, []ShardRestoreTask){
	if !s.isLeader() {
		return false, nil
	}

	curData := s.data.Clone()

	if len(curData.DataNodes) == 0 {
		return false, nil
	}
	shardRestoreTask := make([]ShardRestoreTask, 0)
	nodeIndex := int(curData.Data.Index % uint64(len(curData.DataNodes)))
	for _, dbi := range curData.Data.Databases {
		for _, rpi := range dbi.RetentionPolicies {
			for _, sgi := range rpi.ShardGroups {
				for _, si := range sgi.Shards {
					if len(si.Owners) < defaultReplicaN {
						task := ShardRestoreTask{
							Database: dbi.Name,
							RetentionPolicy: rpi.Name,
						}
						task.ShardID = si.ID
						task.CurrentOwners = si.Owners
						var newOwners []meta.ShardOwner
						nodeID := curData.DataNodes[nodeIndex%len(curData.DataNodes)].ID

						need := defaultReplicaN - len(si.Owners)
						maxChooseCount := 5
						choose:
						for need > 0 {
							nodeIndex++
							maxChooseCount--

							if maxChooseCount < 0 {
								break choose
							}
							exist := false
							for _, owner := range si.Owners {
								if nodeID == owner.NodeID {
									exist = true
								}
							}
							if !exist {
								newOwners = append(newOwners, meta.ShardOwner{NodeID: nodeID})
								need--
								for _, datanode := range curData.DataNodes {
									if datanode.ID == nodeID {
										if task.NewOwnerAddress == nil {
											task.NewOwnerAddress = make(map[uint64]string)
										}
										task.NewOwnerAddress[nodeID] = utils.GetAdminHost(datanode.TCPHost)
									}
								}
							}
						}
						task.NewOwners = newOwners
						shardRestoreTask = append(shardRestoreTask, task)
						
						nodeID = curData.DataNodes[nodeIndex%len(curData.DataNodes)].ID

					}
				}
			}
		}
	}
	if len(shardRestoreTask) > 0{
		return true, shardRestoreTask
	}
	return false, nil
}


// 1. change meta
// 2. send restore command
func (s *store) doRebalance(tasks []ShardRestoreTask) {
	for _, task := range tasks {
		// apply
		if len(task.CurrentOwners) == 0 {
			s.logger.Warn(fmt.Sprintf("shard %d current owners is zero.", task.ShardID))
			continue
		}
		s.logger.Info(fmt.Sprintf("restore shard %d from %d to %v ",
			task.ShardID, task.CurrentOwners[0].NodeID, task.NewOwners))
		var newOwnerNodeIDs []uint64
		for _, owner := range task.NewOwners{
			newOwnerNodeIDs = append(newOwnerNodeIDs, owner.NodeID)
		}

		c := &internal.UpdateShardCommand{
			ID: proto.Uint64(task.ShardID),
			NewOwnerNodeIDs:  newOwnerNodeIDs,
		}
		typ := internal.Command_UpdateShardCommand
		cmd := &internal.Command{ Type: &typ}
		if err := proto.SetExtension(cmd, internal.E_UpdateShardCommand_Command, c); err != nil {
			panic(err)
		}

		b, err := proto.Marshal(cmd)
		if err != nil {
			s.logger.Error(err.Error())
		}
		s.apply(b)

		for _, newOwner := range task.NewOwners {
			s.restoreShard(task.Database, task.RetentionPolicy,
				task.NewOwnerAddress[newOwner.NodeID],
				task.ShardID, task.CurrentOwners[0].NodeID,
				newOwner.NodeID)
		}
		s.logger.Info("do rebalance done.")

	}
}

func (s *store) restoreShard(database, retentionPolicy, address string, shardID, from, to uint64){
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		fmt.Println("dial,",err)
	}
	defer conn.Close()

	client := pb.NewRestoreServiceClient(conn)

	restoreReq := &pb.ShardRestoreRequest{
		Database: database,
		RetentionPolicy: retentionPolicy,
		ShardID: shardID,
		FromNodeId: from,
	}

	resp ,error  := client.ShardRestore(context.Background(), restoreReq)

	if error != nil {
		s.logger.Error(fmt.Sprintf("shard retore for %s failed, error: %s",
				address ,error.Error()))
		return
	}

	if resp.Success {
		// update shard state
		status := &internal.ShardOwnerStatus{
			OwnerNodeID: proto.Uint64(to),
			Status: proto.Int32(int32(ShardWriteRead)),
		}

		statusBytes, _ := proto.Marshal(status)
		c := &internal.UpdateShardCommand{
			ID: proto.Uint64(shardID),
			NewShardStatus: statusBytes,
		}
		typ := internal.Command_UpdateShardCommand
		cmd := &internal.Command{ Type: &typ}
		if err := proto.SetExtension(cmd, internal.E_UpdateShardCommand_Command, c); err != nil {
			panic(err)
		}

		b, err := proto.Marshal(cmd)
		if err != nil {
			s.logger.Error(err.Error())
			return
		}
		s.apply(b)

	}else{
		s.logger.Warn(fmt.Sprintf("restore failed. %s", resp.Message))
	}

}

