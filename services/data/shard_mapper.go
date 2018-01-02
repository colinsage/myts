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
						if si.OwnedBy(c.Node.ID) {
							nodeID = c.Node.ID
						} else if len(si.Owners) > 0 {
							nodeID = si.Owners[rand.Intn(len(si.Owners))].NodeID
						} else {
							// This should not occur but if the shard has no owners then
							// we don't want this to panic by trying to randomly select a node.
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
			mappings.Remotes = append(mappings.Remotes, NewRemoteShardMapping(dialer, nodeID, shardMap))
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

	if sm.Local != nil {
		f,d,err := sm.Local.FieldDimensions(m)
		if err == nil{
			for k, typ := range f {
				fields[k] = typ
			}
			for k := range d {
				dimensions[k] = struct{}{}
			}
		}else{
			return nil,nil,err
		}
	}
	for _, remote := range sm.Remotes {
		f,d,err := remote.FieldDimensions(m)
		if err == nil{
			for k, typ := range f {
				fields[k] = typ
			}
			for k := range d {
				dimensions[k] = struct{}{}
			}
		}else{
			//TODO
			return nil,nil,err
		}
	}
	return fields,dimensions,nil
}

func (sm *ShardMappings) MapType(m *influxql.Measurement, field string) influxql.DataType {
	//TODO
	if sm.Local != nil {
		input := sm.Local.MapType(m, field)
		if input != influxql.Unknown {
			return input
		}
	}
	for _,remote := range sm.Remotes {
		input := remote.MapType(m,field)
		if input != influxql.Unknown {
			return input
		}
	}
	return influxql.Unknown
}

func (sm *ShardMappings) CreateIterator(ctx context.Context, m *influxql.Measurement, opt query.IteratorOptions) (query.Iterator, error) {
	//TODO
	itrs := make([]query.Iterator, 0, 100)

	for _,remote := range sm.Remotes {
		input, err := remote.CreateIterator(ctx, m, opt)
		if err == nil {
			if input != nil {
				itrs = append(itrs, input )
			}
		}else{
			sm.Logger.Error("error:"+ err.Error())
		}
	}

	if sm.HasLocal == true {
		input, err := sm.Local.CreateIterator(ctx, m, opt)
		if err == nil {
			itrs = append(itrs, input )
		}else{
			sm.Logger.Error(err.Error())
		}
	}

	return query.Iterators(itrs).Merge(opt)
}

// Determines the potential cost for creating an iterator.
func (sm *ShardMappings) IteratorCost(m *influxql.Measurement, opt query.IteratorOptions) (query.IteratorCost, error) {
	var costs query.IteratorCost
	if sm.Local != nil {
		cost, err := sm.Local.IteratorCost( m, opt)
		if err == nil {
			costs.Combine(cost)
		} else{
			return cost, err
		}
	}

	for _,remote := range sm.Remotes {
		cost, err := remote.IteratorCost(m, opt)
		if err == nil {
			costs.Combine(cost)
		}else {
			return cost, err
		}
	}

	return costs,nil
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
}

func NewRemoteShardMapping(dialer *NodeDialer, nodeID uint64, shardMap map[coordinator.Source][]uint64) *RemoteShardMapping{
	return &RemoteShardMapping {
		dialer: dialer,
		nodeID:nodeID,
		shardMap:shardMap,
	}
}

func (ic *RemoteShardMapping) IteratorCost( m *influxql.Measurement,opt query.IteratorOptions) (query.IteratorCost, error) {

	//TODO
	return query.IteratorCost{}, nil
}
// CreateIterator creates a remote streaming iterator.
func (ic *RemoteShardMapping) CreateIterator(ctx context.Context, m *influxql.Measurement,opt query.IteratorOptions) (query.Iterator, error) {
	conn, err := ic.dialer.DialNode(ic.nodeID)
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
			ShardIDs: ic.shardMap[source],
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

//func GetPointType(point query.Point) influxql.DataType {
//	t := influxql.Unknown
//	switch pt := point.(type) {
//		case *query.FloatPoint:
//			t = influxql.Float
//		case *query.IntegerPoint:
//			t = influxql.Integer
//		case *query.StringPoint:
//			t = influxql.String
//		case *query.BooleanPoint:
//			t = influxql.Boolean
//		default:
//			fmt.Printf("unexpected type %T\n", pt)
//	}
//
//	return t
//}
// FieldDimensions returns the unique fields and dimensions across a list of sources.
func (ic *RemoteShardMapping) FieldDimensions(m *influxql.Measurement) (fields map[string]influxql.DataType, dimensions map[string]struct{}, err error) {
	conn, err := ic.dialer.DialNode(ic.nodeID)
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
		ShardIDs: ic.shardMap[source],
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

func (ic *RemoteShardMapping) MapType (m *influxql.Measurement, field string) influxql.DataType {
	//TODO
	conn, err := ic.dialer.DialNode(ic.nodeID)
	if err != nil {
		return influxql.Unknown
	}
	defer conn.Close()
	source := coordinator.Source{
		Database:        m.Database,
		RetentionPolicy: m.RetentionPolicy,
	}

	sg := ic.shardMap[source]
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

func (ic *RemoteShardMapping)Close() error {
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


