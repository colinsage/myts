package data

import (
	"io"
	"time"
	"math/rand"
	"net"

	"github.com/influxdata/influxdb/query"
	"github.com/influxdata/influxdb/tsdb"

	"github.com/influxdata/influxdb/services/meta"
	"github.com/influxdata/influxql"
	"context"
	"github.com/influxdata/influxdb/coordinator"
	"github.com/uber-go/zap"
	"encoding/json"
	"sync"
	"fmt"
)

// IteratorCreator is an interface that combines mapping fields and creating iterators.
type IteratorCreator interface {
	query.IteratorCreator
	influxql.FieldMapper
	io.Closer
}

// ShardMapper retrieves and maps shards into an IteratorCreator that can later be
// used for executing queries.
type ShardMapper interface {
	MapShards(sources influxql.Sources, opt *query.SelectOptions) (IteratorCreator, error)
}

type DistributedShardMapper struct {

	MetaClient interface {
		ShardGroupsByTimeRange(database, policy string, min, max time.Time) (a []meta.ShardGroupInfo, err error)
		DataNode(id uint64) (*meta.NodeInfo, error)
		ShardIsReadable(shardID, nodeID uint64) bool
	}

	TSDBStore interface {
		ShardGroup(ids []uint64) tsdb.ShardGroup
	}

	Node *meta.NodeInfo

	// Remote execution timeout
	Timeout time.Duration

	Logger zap.Logger
}

func (c *DistributedShardMapper)  MapShards(sources influxql.Sources, t influxql.TimeRange, opt query.SelectOptions) (query.ShardGroup, error) {
	ics := &ShardMappings{
		HasLocal: false,
	}

	ics.WithLogger(c.Logger)
	tmin := time.Unix(0, t.MinTimeNano())
	tmax := time.Unix(0, t.MaxTimeNano())
	if err := c.mapShards(ics, sources, tmin, tmax); err != nil {
		return nil, err
	}
	return ics, nil
}
func (c *DistributedShardMapper)  mapShards(mappings *ShardMappings, sources influxql.Sources, tmin, tmax time.Time) error {
	all := make(map[coordinator.Source]*meta.ShardGroupInfo)

	shardIDsByNodeID := make(map[uint64]map[coordinator.Source][]uint64)
	for _, s := range sources {
		switch s := s.(type) {
		case *influxql.Measurement:
			source := coordinator.Source{
				Database:        s.Database,
				RetentionPolicy: s.RetentionPolicy,
			}

			// Retrieve the list of shards for this database. This list of
			// shards is always the same regardless of which measurement we are
			// using.
			if _, ok := all[source]; !ok {
				groups, err := c.MetaClient.ShardGroupsByTimeRange(s.Database, s.RetentionPolicy, tmin, tmax)
				if err != nil {
					return err
				}

				if len(groups) == 0 {
					all[source] = nil
					continue
				}

				//shardIDs := make([]uint64, 0, len(groups[0].Shards)*len(groups))
				for _, g := range groups {
					for _, si := range g.Shards {
						var nodeID uint64
						found := false
						if si.OwnedBy(c.Node.ID) {
							nodeID = c.Node.ID
						} else if len(si.Owners) > 0 {
							nodeID = si.Owners[rand.Intn(len(si.Owners))].NodeID
						} else {
							// This should not occur but if the shard has no owners then
							// we don't want this to panic by trying to randomly select a node.
							continue
						}
						if c.MetaClient.ShardIsReadable(si.ID, nodeID) {
							found = true
						}else {
							inner:
							for i:=0; i< len(si.Owners) ;i++ {
								nodeID = si.Owners[i].NodeID
								if c.MetaClient.ShardIsReadable(si.ID, nodeID) {
									found = true
									break inner
								}

							}
						}

						if !found {
							c.Logger.Error(fmt.Sprintf("not found node for shard %d", si.ID))
							continue
						}

						ss := shardIDsByNodeID[nodeID]

						if ss == nil{
							ss = make(map[coordinator.Source][]uint64)
							shardIDsByNodeID[nodeID] = ss
						}
						shardIDsByNodeID[nodeID][source] = append(shardIDsByNodeID[nodeID][source], si.ID)
						//shardIDs = append(shardIDs, si.ID)
					}
				}
				all[source] = &groups[0]
			}
		case *influxql.SubQuery:
			if err := c.mapShards(mappings, s.Statement.Sources, tmin, tmax); err != nil {
				return err
			}
		}
	}

	if err := func() error {
		for nodeID, shardMap := range shardIDsByNodeID {
			// Sort shard IDs so we get more predicable execution.
			//sort.Sort(uint64Slice(shardIDs))

			// Create iterator creators from TSDB if local.
			if nodeID == c.Node.ID {
				mappings.HasLocal = true
				if mappings.Local == nil {
					mappings.Local = &coordinator.LocalShardMapping{
						ShardMap: make(map[coordinator.Source]tsdb.ShardGroup),
					}
				}
				for source, shardIDs := range shardMap{
					mappings.Local.ShardMap[source] = c.TSDBStore.ShardGroup(shardIDs)
				}
				continue
			}

			// Otherwise create iterator creator remotely.
			dialer := &NodeDialer{
				MetaClient: c.MetaClient,
				Timeout:    c.Timeout,
			}
			rsm := NewRemoteShardMapping(dialer, nodeID, shardMap)
			rsm.WithLogger(c.Logger)
			mappings.Remotes = append(mappings.Remotes, rsm)
		}

		return nil
	}(); err != nil {
		return err
	}
	return nil

}

type ShardMappings struct{
	Local *coordinator.LocalShardMapping
	Remotes []*RemoteShardMapping
	HasLocal bool

	Logger zap.Logger
}

func (sm *ShardMappings) WithLogger(logger zap.Logger){
	sm.Logger = logger.With(zap.String("service", "shard-mapping"))
}

func (sm *ShardMappings) FieldDimensions(m *influxql.Measurement) (fields map[string]influxql.DataType, dimensions map[string]struct{}, err error) {
	//TODO
	fields = make(map[string]influxql.DataType)
	dimensions = make(map[string]struct{})

	var wg sync.WaitGroup

	if sm.HasLocal {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f,d,err := sm.Local.FieldDimensions(m)
			if err == nil{
				for k, typ := range f {
					fields[k] = typ
				}
				for k := range d {
					dimensions[k] = struct{}{}
				}
			}else{
				sm.Logger.Error(err.Error())
			}
		}()

	}

	for _, remote := range sm.Remotes {
		wg.Add(1)

		go func(){
			defer wg.Done()
			f,d,err := remote.FieldDimensions(m)
			if err == nil{
				for k, typ := range f {
					fields[k] = typ
				}
				for k := range d {
					dimensions[k] = struct{}{}
				}
			}else{
				sm.Logger.Error(err.Error())
			}
		}()

	}
	wg.Wait()
	return fields,dimensions,nil
}

func (sm *ShardMappings) MapType(m *influxql.Measurement, field string) influxql.DataType {
	//TODO
	if sm.Local != nil {
		mtype := sm.Local.MapType(m, field)
		if mtype != influxql.Unknown {
			return mtype
		}
	}
	for _,remote := range sm.Remotes {
		mtype := remote.MapType(m,field)
		if mtype != influxql.Unknown {
			return mtype
		}
	}
	return influxql.Unknown
}

func (sm *ShardMappings) CreateIterator(ctx context.Context, m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
	//TODO
	itrs := make([]query.Iterator, 0, 100)
	var wg sync.WaitGroup
	mu := new(sync.RWMutex)
	for _,remote := range sm.Remotes {
		wg.Add(1)
		go func(){
			defer wg.Done()
			input, err := remote.CreateIterator(ctx, m, opt)
			if err == nil {
				if input != nil {
					mu.Lock()
					itrs = append(itrs, input )
					mu.Unlock()
				}
			}else{
				sm.Logger.Error("error:"+ err.Error())
			}
		}()

	}

	if sm.HasLocal == true {
		input, err := sm.Local.CreateIterator(ctx, m, opt)
		if err == nil {
			mu.Lock()
			itrs = append(itrs, input )
			mu.Unlock()
		}else{
			sm.Logger.Error(err.Error())
		}
	}

	wg.Wait()
	return query.Iterators(itrs).Merge(opt)
}

// Determines the potential cost for creating an iterator.
func (sm *ShardMappings) IteratorCost(m *influxql.Measurement, opt query.IteratorOptions) (query.IteratorCost, error) {
	var costs query.IteratorCost
	var wg sync.WaitGroup
	mu := new(sync.RWMutex)

	if sm.HasLocal {
		cost, err := sm.Local.IteratorCost( m, opt)
		if err == nil {
			costs = costs.Combine(cost)
		} else{
			return costs, err
		}
	}

	for _,remote := range sm.Remotes {
		wg.Add(1)
		go func(){
			defer wg.Done()
			cost, err := remote.IteratorCost(m, opt)
			if err == nil {
				mu.Lock()
				costs = costs.Combine(cost)
				mu.Unlock()
			}else {
				sm.Logger.Error(err.Error())
			}
		}()
	}
	wg.Wait()

	return costs, nil
}


func (sm *ShardMappings) Close() error {
	return nil
}

// ShardMapper maps data sources to a list of shard information.
type RemoteShardMapping struct {
	dialer   *NodeDialer
	nodeID   uint64
	shardMap map[coordinator.Source][]uint64
	//shardIDs []uint64

	Logger zap.Logger
}

func NewRemoteShardMapping(dialer *NodeDialer, nodeID uint64, shardMap map[coordinator.Source][]uint64) *RemoteShardMapping{
	return &RemoteShardMapping {
		dialer: dialer,
		nodeID:nodeID,
		shardMap:shardMap,
	}
}


func (rsm *RemoteShardMapping) WithLogger(logger zap.Logger){
	rsm.Logger = logger.With(zap.String("service", "remote-shard"))
}
func (rsm *RemoteShardMapping) IteratorCost( m *influxql.Measurement,opt query.IteratorOptions) (query.IteratorCost, error) {
	conn, err := rsm.dialer.DialNode(rsm.nodeID)
	if err != nil {
		return query.IteratorCost{}, err
	}
	defer conn.Close()
	source := coordinator.Source{
		Database:        m.Database,
		RetentionPolicy: m.RetentionPolicy,
	}

	sg := rsm.shardMap[source]
	if sg == nil {
		return query.IteratorCost{}, nil
	}

	// Write request.
	if err := EncodeTLV(conn, iteratorCostRequestMessage, &IteratorCostRequest{
		ShardIDs: sg,
		Measurement: m,
		IteratorOptions: &opt,
	}); err != nil {
		return query.IteratorCost{}, err
	}

	// Read the response.
	var resp IteratorCostResponse
	if _, err := DecodeTLV(conn, &resp); err != nil {
		return query.IteratorCost{}, err
	}

	if len(resp.Error) != 0 {
		return resp.IteratorCost, RpcError{
			message: resp.Error,
		}
	}

	rlt, _ := json.Marshal(resp)
	rsm.Logger.Info(string(rlt[:]))

	return resp.IteratorCost, nil

}
// CreateIterator creates a remote streaming iterator.
func (rsm *RemoteShardMapping) CreateIterator(ctx context.Context, m *influxql.Measurement,opt query.IteratorOptions) (query.Iterator, error) {
	conn, err := rsm.dialer.DialNode(rsm.nodeID)
	if err != nil {
		return nil, err
	}

	source := coordinator.Source{
		Database:        m.Database,
		RetentionPolicy: m.RetentionPolicy,
	}

	if err := func() error {
		// Write request.
		if err := EncodeTLV(conn, createIteratorRequestMessage, &CreateIteratorRequest{
			Measurement: m,
			ShardIDs: rsm.shardMap[source],
			Opt:      opt,
		}); err != nil {
			return err
		}

		// Read the response.
		var resp CreateIteratorResponse
		if _, err := DecodeTLV(conn, &resp); err != nil {
			return err
		} else if resp.Err != nil {
			return err
		}

		return nil
	}(); err != nil {
		conn.Close()
		return nil, err
	}

	t := influxql.Integer
	if opt.Expr != nil {
		name := opt.Expr.(*influxql.Call).Name
		switch name {
		case "count":
			t = influxql.Integer
		default:
			t = influxql.Float
		}
	}

	return query.NewReaderIterator(ctx, conn, t, query.IteratorStats{} ), nil
}

// FieldDimensions returns the unique fields and dimensions across a list of sources.
func (rsm *RemoteShardMapping) FieldDimensions(m *influxql.Measurement) (fields map[string]influxql.DataType, dimensions map[string]struct{}, err error) {
	conn, err := rsm.dialer.DialNode(rsm.nodeID)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()

	source := coordinator.Source{
		Database:        m.Database,
		RetentionPolicy: m.RetentionPolicy,
	}

	// Write request.
	if err := EncodeTLV(conn, fieldDimensionsRequestMessage, &FieldDimensionsRequest{
		Measurement: m.Name,
		ShardIDs: rsm.shardMap[source],
		//Sources:  sources,
	}); err != nil {
		return nil, nil, err
	}

	// Read the response.
	var resp FieldDimensionsResponse
	if _, err := DecodeTLV(conn, &resp); err != nil {
		return nil, nil, err
	}
	return resp.Fields, resp.Dimensions, resp.Err
}

func (rsm *RemoteShardMapping) MapType (m *influxql.Measurement, field string) influxql.DataType {
	//TODO
	conn, err := rsm.dialer.DialNode(rsm.nodeID)
	if err != nil {
		return influxql.Unknown
	}
	defer conn.Close()
	source := coordinator.Source{
		Database:        m.Database,
		RetentionPolicy: m.RetentionPolicy,
	}

	sg := rsm.shardMap[source]
	if sg == nil {
		return influxql.Unknown
	}

	// Write request.
	if err := EncodeTLV(conn, mapTypeRequestMessage, &MapTypeRequest{
		ShardIDs: sg,
		Measurement: m.Name,
		Field: field,
	}); err != nil {
		return influxql.Unknown
	}

	// Read the response.
	var resp MapTypeResponse
	if _, err := DecodeTLV(conn, &resp); err != nil {
		return influxql.Unknown
	}

	t := resp.Type
	return influxql.DataType(int(t))
}

func (rsm *RemoteShardMapping) Close() error {
	return nil
}

// NodeDialer dials connections to a given node.
type NodeDialer struct {
	MetaClient interface {
		DataNode(id uint64) (*meta.NodeInfo, error)
	}
	Timeout    time.Duration
}

// DialNode returns a connection to a node.
func (d *NodeDialer) DialNode(nodeID uint64) (net.Conn, error) {
	ni, err := d.MetaClient.DataNode(nodeID)
	if err != nil {
		return nil, err
	}

	conn, err := net.Dial("tcp", ni.TCPHost)
	if err != nil {
		return nil, err
	}
//	conn.SetDeadline(time.Now().Add(d.Timeout))

	// Write the cluster multiplexing header byte
	if _, err := conn.Write([]byte{MuxHeader}); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}


