package meta

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hashicorp/raft"
	"github.com/hashicorp/raft-boltdb"
	"strconv"
)

// Raft configuration.
const (
	raftLogCacheSize      = 512
	raftSnapshotsRetained = 2
	raftTransportMaxPool  = 3
	raftTransportTimeout  = 10 * time.Second
)

// raftState is a consensus strategy that uses a local raft implementation for
// consensus operations.
type raftState struct {
	wg        sync.WaitGroup
	config    *Config
	closing   chan struct{}
	raft      *raft.Raft
	transport *raft.NetworkTransport
	raftStore *raftboltdb.BoltStore
	ln        net.Listener
	addr      string
	logger    *log.Logger
	path      string
}

func newRaftState(c *Config, addr string) *raftState {
	return &raftState{
		config: c,
		addr:   addr,
	}
}

func (r *raftState) open(s *store, ln net.Listener, initializePeers []string) error {
	r.ln = ln
	r.closing = make(chan struct{})

	// Setup raft configuration.
	config := raft.DefaultConfig()
	config.LogOutput = ioutil.Discard

	if r.config.ClusterTracing {
		config.Logger = r.logger
	}
	//config.HeartbeatTimeout = time.Duration(r.config.HeartbeatTimeout)
	//config.ElectionTimeout = time.Duration(r.config.ElectionTimeout)
	//config.LeaderLeaseTimeout = time.Duration(r.config.LeaderLeaseTimeout)
	//config.CommitTimeout = time.Duration(r.config.CommitTimeout)
	// Since we actually never call `removePeer` this is safe.
	// If in the future we decide to call remove peer we have to re-evaluate how to handle this
	//config.ShutdownOnRemove = false

	// Build raft layer to multiplex listener.
	//r.raftLayer = newRaftLayer(r.addr, r.ln)

	// Create a transport layer
	trans, error := raft.NewTCPTransport(r.config.BindAddress,nil, 3, 1*time.Second, os.Stdout)
	if error != nil {
		r.logger.Println(error.Error())
	}

	r.transport = trans

	// Create the log store and stable store.
	store, err := raftboltdb.NewBoltStore(filepath.Join(r.path, "raft.db"))
	if err != nil {
		return fmt.Errorf("new bolt store: %s", err)
	}
	r.raftStore = store

	lastIndex,_ := store.LastIndex()
	l := new(raft.Log)
	store.GetLog(lastIndex, l)
	log.Println("meta dir:" + strconv.FormatUint(lastIndex,10) + " "+ string(l.Data[:]))
	// Create the snapshot store.
	snapshots, err := raft.NewFileSnapshotStore(r.path, raftSnapshotsRetained, os.Stderr)
	if err != nil {
		return fmt.Errorf("file snapshot store: %s", err)
	}

	config.LocalID = raft.ServerID(r.config.BindAddress)
	//config.StartAsLeader = true
	config.Logger = r.logger

	var configuration raft.Configuration

	for _ , peer := range initializePeers{
		configuration.Servers = append(configuration.Servers, raft.Server{
			Suffrage: raft.Voter,
			ID:       raft.ServerID(peer),
			Address:  raft.ServerAddress(peer),
		})
	}


	if err := raft.BootstrapCluster(config, store, store, snapshots, r.transport, configuration); err != nil {
		log.Printf("[ERR] BootstrapCluster failed: %v \r\n", err)
	}

	// Create raft log.
	ra, err := raft.NewRaft(config, (*storeFSM)(s), store, store, snapshots, r.transport)
	if err != nil {
		return fmt.Errorf("new raft: %s", err)
	}
	r.raft = ra

	log.Println("new raft done.")
	r.wg.Add(1)
	go r.logLeaderChanges()

	return nil
}

func (r *raftState) logLeaderChanges() {
	defer r.wg.Done()
	// Logs our current state (Node at 1.2.3.4:8088 [Follower])
	r.logger.Printf(r.raft.String())
	for {
		select {
		case <-r.closing:
			return
		case <-r.raft.LeaderCh():
			peers, err := r.peers()
			if err != nil {
				r.logger.Printf("failed to lookup peers: %v", err)
			}
			r.logger.Printf("%v. peers=%v", r.raft.String(), peers)
		}
	}
}

func (r *raftState) close() error {
	if r == nil {
		return nil
	}
	if r.closing != nil {
		close(r.closing)
	}
	r.wg.Wait()

	if r.transport != nil {
		r.transport.Close()
		r.transport = nil
	}

	// Shutdown raft.
	if r.raft != nil {
		if err := r.raft.Shutdown().Error(); err != nil {
			return err
		}
		r.raft = nil
	}

	if r.raftStore != nil {
		r.raftStore.Close()
		r.raftStore = nil
	}

	return nil
}

// apply applies a serialized command to the raft log.
func (r *raftState) apply(b []byte) error {
	// Apply to raft log.
	f := r.raft.Apply(b, time.Second*5)
	if err := f.Error(); err != nil {
		log.Printf("raft state in apply error. "+ err.Error())
		return err
	}
	log.Printf("raft state in apply done" )
	// Return response if it's an error.
	// No other non-nil objects should be returned.
	resp := f.Response()
	if err, ok := resp.(error); ok {
		return err
	}
	if resp != nil {
		panic(fmt.Sprintf("unexpected response: %#v", resp))
	}

	return nil
}

func (r *raftState) lastIndex() uint64 {
	return r.raft.LastIndex()
}

func (r *raftState) snapshot() error {
	future := r.raft.Snapshot()
	return future.Error()
}

// addPeer adds addr to the list of peers in the cluster.
func (r *raftState) addPeer(addr string) error {
	peers,_ := r.peers()

	for _, p := range peers {
		if addr == p {
			return nil
		}
	}

	if fut := r.raft.AddVoter(raft.ServerID(addr), raft.ServerAddress(addr),0,0); fut.Error() != nil {
		return fut.Error()
	}
	return nil
}

// removePeer removes addr from the list of peers in the cluster.
func (r *raftState) removePeer(addr string) error {
	// Only do this on the leader
	if !r.isLeader() {
		return raft.ErrNotLeader
	}

	peers,_ := r.peers()

	var exists bool
	for _, p := range peers {
		if addr == p {
			exists = true
			break
		}
	}

	if !exists {
		return nil
	}

	if fut := r.raft.RemovePeer(raft.ServerAddress(addr)); fut.Error() != nil {
		return fut.Error()
	}
	return nil
}

func (r *raftState) peers() ([]string, error) {
	f := r.raft.GetConfiguration()
	if err := f.Error(); err != nil {
		log.Printf("raft state in apply , error" )
		return nil, err
	}
	peers := f.Configuration().Servers

	//js,_ := json.Marshal(peers)
	log.Println("raft status: " + r.raft.String())

	var p  []string
	for _, s := range peers {
		p = append(p, string(s.Address))
	}
	return p, nil

}

func (r *raftState) leader() string {
	if r.raft == nil {
		return ""
	}

	return string(r.raft.Leader())
}

func (r *raftState) isLeader() bool {
	if r.raft == nil {
		return false
	}
	return r.raft.State() == raft.Leader
}





