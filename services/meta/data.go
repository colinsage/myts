package meta

import (
	"sort"
	"time"

	"github.com/gogo/protobuf/proto"
	"fmt"
	"github.com/influxdata/influxdb/services/meta"
	"github.com/colinsage/myts/services/meta/internal"
	"github.com/influxdata/influxdb"
	"github.com/influxdata/influxql"
)

// Data represents the top level collection of all metadata.
type Data struct {
	Data meta.Data

	MetaNodes []meta.NodeInfo
	DataNodes []meta.NodeInfo
	MaxNodeID       uint64
}

// Database returns a DatabaseInfo by the database name.
func (data *Data) Database(name string) *meta.DatabaseInfo {
	return data.Data.Database(name)
}

// CloneDatabases returns a copy of the DatabaseInfo.
func (data *Data) CloneDatabases() []meta.DatabaseInfo {
	return data.Data.CloneDatabases()
}


// CreateDatabase creates a new database.
// It returns an error if name is blank or if a database with the same name already exists.
func (data *Data) CreateDatabase(name string) error {
	return data.Data.CreateDatabase(name)
}

// DropDatabase removes a database by name. It does not return an error
// if the database cannot be found.
func (data *Data) DropDatabase(name string) error {
	return data.Data.DropDatabase(name)
}

// RetentionPolicy returns a retention policy for a database by name.
func (data *Data) RetentionPolicy(database, name string) (*meta.RetentionPolicyInfo, error) {
	return data.Data.RetentionPolicy(database, name)
}

// CreateRetentionPolicy creates a new retention policy on a database.
// It returns an error if name is blank or if the database does not exist.
func (data *Data) CreateRetentionPolicy(database string, rpi *meta.RetentionPolicyInfo, makeDefault bool) error {
	 return data.Data.CreateRetentionPolicy(database,rpi, makeDefault )
}

// DropRetentionPolicy removes a retention policy from a database by name.
func (data *Data) DropRetentionPolicy(database, name string) error {

	return data.Data.DropRetentionPolicy(database, name)
}

// UpdateRetentionPolicy updates an existing retention policy.
func (data *Data) UpdateRetentionPolicy(database, name string, rpu *meta.RetentionPolicyUpdate, makeDefault bool) error {
	 return data.Data.UpdateRetentionPolicy(database, name, rpu, makeDefault)
}

// DropShard removes a shard by ID.
//
// DropShard won't return an error if the shard can't be found, which
// allows the command to be re-run in the case that the meta store
// succeeds but a data node fails.
func (data *Data) DropShard(id uint64) {
	 data.Data.DropShard(id)
}

// ShardGroups returns a list of all shard groups on a database and retention policy.
func (data *Data) ShardGroups(database, policy string) ([]meta.ShardGroupInfo, error) {
	 return data.Data.ShardGroups(database, policy)
}

// ShardGroupsByTimeRange returns a list of all shard groups on a database and policy that may contain data
// for the specified time range. Shard groups are sorted by start time.
func (data *Data) ShardGroupsByTimeRange(database, policy string, tmin, tmax time.Time) ([]meta.ShardGroupInfo, error) {
	 return data.Data.ShardGroupsByTimeRange(database, policy, tmin, tmax)
}

// ShardGroupByTimestamp returns the shard group on a database and policy for a given timestamp.
func (data *Data) ShardGroupByTimestamp(database, policy string, timestamp time.Time) (*meta.ShardGroupInfo, error) {
	return data.Data.ShardGroupByTimestamp(database, policy, timestamp)
}

// CreateShardGroup creates a shard group on a database and policy for a given timestamp.
func (data *Data) CreateShardGroup(database, policy string, timestamp time.Time) error {
	//TODO maybe change
	// Ensure there are nodes in the metadata.
	if len(data.DataNodes) == 0 {
		return nil
	}

	// Find retention policy.
	rpi, err := data.RetentionPolicy(database, policy)
	if err != nil {
		return err
	} else if rpi == nil {
		return influxdb.ErrRetentionPolicyNotFound(policy)
	}

	// Verify that shard group doesn't already exist for this timestamp.
	if rpi.ShardGroupByTimestamp(timestamp) != nil {
		return nil
	}

	// Require at least one replica but no more replicas than nodes.
	replicaN := rpi.ReplicaN
	if replicaN == 0 {
		replicaN = 1
	} else if replicaN > len(data.DataNodes) {
		replicaN = len(data.DataNodes)
	}

	// Determine shard count by node count divided by replication factor.
	// This will ensure nodes will get distributed across nodes evenly and
	// replicated the correct number of times.
	shardN := len(data.DataNodes) / replicaN

	// Create the shard group.
	data.Data.MaxShardGroupID++
	sgi := meta.ShardGroupInfo{}
	sgi.ID = data.Data.MaxShardGroupID
	sgi.StartTime = timestamp.Truncate(rpi.ShardGroupDuration).UTC()
	sgi.EndTime = sgi.StartTime.Add(rpi.ShardGroupDuration).UTC()

	// Create shards on the group.
	sgi.Shards = make([]meta.ShardInfo, shardN)
	for i := range sgi.Shards {
		data.Data.MaxShardID++
		sgi.Shards[i] = meta.ShardInfo{ID: data.Data.MaxShardID}
	}

	// Assign data nodes to shards via round robin.
	// Start from a repeatably "random" place in the node list.
	nodeIndex := int(data.Data.Index % uint64(len(data.DataNodes)))
	for i := range sgi.Shards {
		si := &sgi.Shards[i]
		for j := 0; j < replicaN; j++ {
			nodeID := data.DataNodes[nodeIndex%len(data.DataNodes)].ID
			si.Owners = append(si.Owners, meta.ShardOwner{NodeID: nodeID})
			nodeIndex++
		}
	}

	// Retention policy has a new shard group, so update the policy. Shard
	// Groups must be stored in sorted order, as other parts of the system
	// assume this to be the case.
	rpi.ShardGroups = append(rpi.ShardGroups, sgi)
	sort.Sort(meta.ShardGroupInfos(rpi.ShardGroups))
	 return nil
}

// DeleteShardGroup removes a shard group from a database and retention policy by id.
func (data *Data) DeleteShardGroup(database, policy string, id uint64) error {
	 return data.Data.DeleteShardGroup(database, policy, id)
}

// CreateContinuousQuery adds a named continuous query to a database.
func (data *Data) CreateContinuousQuery(database, name, query string) error {
	 return data.Data.CreateContinuousQuery(database, name, query)
}

// DropContinuousQuery removes a continuous query.
func (data *Data) DropContinuousQuery(database, name string) error {
	return data.Data.DropContinuousQuery(database, name)
}


// CreateSubscription adds a named subscription to a database and retention policy.
func (data *Data) CreateSubscription(database, rp, name, mode string, destinations []string) error {
	 return data.Data.CreateSubscription(database, rp, name, mode, destinations )
}

// DropSubscription removes a subscription.
func (data *Data) DropSubscription(database, rp, name string) error {
	 return data.Data.DropSubscription(database, rp, name)
}

// User returns a user by username.
func (data *Data) User(username string) meta.User {
	return data.Data.User(username)
}

// CreateUser creates a new user.
func (data *Data) CreateUser(name, hash string, admin bool) error {
	 return data.Data.CreateUser(name, hash, admin)
}

// DropUser removes an existing user by name.
func (data *Data) DropUser(name string) error {
	 return data.Data.DropUser(name)
}

// UpdateUser updates the password hash of an existing user.
func (data *Data) UpdateUser(name, hash string) error {
	 return data.Data.UpdateUser(name, hash)
}

// CloneUsers returns a copy of the user infos.
func (data *Data) CloneUsers() []meta.UserInfo {
	 return data.Data.CloneUsers()
}

// SetPrivilege sets a privilege for a user on a database.
func (data *Data) SetPrivilege(name, database string, p influxql.Privilege) error {
	return data.Data.SetPrivilege(name, database, p)
}

// SetAdminPrivilege sets the admin privilege for a user.
func (data *Data) SetAdminPrivilege(name string, admin bool) error {
	 return data.Data.SetAdminPrivilege(name, admin)
}

// AdminUserExists returns true if an admin user exists.
func (data Data) AdminUserExists() bool {
	return data.Data.AdminUserExists()
}

// UserPrivileges gets the privileges for a user.
func (data *Data) UserPrivileges(name string) (map[string]influxql.Privilege, error) {
	 return data.Data.UserPrivileges(name)
}

// UserPrivilege gets the privilege for a user on a database.
func (data *Data) UserPrivilege(name, database string) (*influxql.Privilege, error) {
	 return data.Data.UserPrivilege(name, database)
}

// CloneUsers returns a copy of the user infos.
func (data *Data) CloneMetaNodes() []meta.NodeInfo {
	if len(data.MetaNodes) == 0 {
		return []meta.NodeInfo{}
	}
	metas := make([]meta.NodeInfo, len(data.MetaNodes))
	for i := range data.MetaNodes {
		metas[i]  = data.MetaNodes[i]
	}

	return metas
}

// CloneUsers returns a copy of the user infos.
func (data *Data) CloneDataNodes() []meta.NodeInfo {
	if len(data.DataNodes) == 0 {
		return []meta.NodeInfo{}
	}
	datas := make([]meta.NodeInfo, len(data.DataNodes))
	for i := range data.DataNodes {
		datas[i]  = data.DataNodes[i]
	}

	return datas
}

// Clone returns a copy of data with a new version.
func (data *Data) Clone() *Data {
	other := *data

	other.Data = *data.Data.Clone()

	other.DataNodes = data.CloneDataNodes()
	other.MetaNodes = data.CloneMetaNodes()

	return &other
}

// marshal serializes data to a protobuf representation.
func (data *Data) marshal() *internal.Data {
	pb := &internal.Data{
		Term:      proto.Uint64(data.Data.Term),
		Index:     proto.Uint64(data.Data.Index),
		ClusterID: proto.Uint64(data.Data.ClusterID),

		MaxShardGroupID: proto.Uint64(data.Data.MaxShardGroupID),
		MaxShardID:      proto.Uint64(data.Data.MaxShardID),

		// Need this for reverse compatibility
		MaxNodeID: proto.Uint64(0),
	}

	pb.Databases = make([]*internal.DatabaseInfo, len(data.Data.Databases))
	for i := range data.Data.Databases {
		pb.Databases[i] = marshalDatabaseInfo(data.Data.Databases[i])
	}


	if len(data.MetaNodes) > 0 {
		pb.MetaNodes = make([]*internal.NodeInfo, len(data.MetaNodes))
		for i := range data.MetaNodes {
			pb.MetaNodes[i] = marshalNodeInfo(data.MetaNodes[i])
		}
	}

	if len(data.DataNodes) > 0 {
		pb.DataNodes = make([]*internal.NodeInfo, len(data.DataNodes))
		for i := range data.DataNodes{
			pb.DataNodes[i] = marshalNodeInfo(data.DataNodes[i])
		}
	}

	//pb.Users = make([]*internal.UserInfo, len(data.Data.Users))
	//for i := range data.Data.Users {
	//	pb.Users[i] = data.Data.Users[i].MarshalBinary()
	//}

	pb.MaxNodeID = proto.Uint64(data.MaxNodeID)
	return pb
}

// unmarshal deserializes from a protobuf representation.
func (data *Data) unmarshal(pb *internal.Data) {
	data.Data.Term = pb.GetTerm()
	data.Data.Index = pb.GetIndex()
	data.Data.ClusterID = pb.GetClusterID()

	data.Data.MaxShardGroupID = pb.GetMaxShardGroupID()
	data.Data.MaxShardID = pb.GetMaxShardID()

	data.Data.Databases = make([]meta.DatabaseInfo, len(pb.GetDatabases()))
	for i, x := range pb.GetDatabases() {
		data.Data.Databases[i] = *unmarshalDatabaseInfo(x)
	}

	//data.Data.Users = make([]meta.UserInfo, len(pb.GetUsers()))
	//for i, x := range pb.GetUsers() {
	//	data.Data.Users[i].unmarshal(x)
	//}

	// Exhaustively determine if there is an admin user. The marshalled cache
	// value may not be correct.
	//data.adminUserExists = data.hasAdminUser()
	data.MetaNodes = make([]meta.NodeInfo, len(pb.GetMetaNodes()))
	for i, x := range pb.GetMetaNodes() {
		data.MetaNodes[i] = *unmarshalNodeInfo(x)
	}

	data.DataNodes = make([]meta.NodeInfo, len(pb.GetDataNodes()))
	for i, x := range pb.GetDataNodes() {
		data.DataNodes[i] = *unmarshalNodeInfo(x)
	}
}

// MarshalBinary encodes the metadata to a binary format.
func (data *Data) MarshalBinary() ([]byte, error) {
	return proto.Marshal(data.marshal())
}

// UnmarshalBinary decodes the object from a binary format.
func (data *Data) UnmarshalBinary(buf []byte) error {
	var pb internal.Data
	if err := proto.Unmarshal(buf, &pb); err != nil {
		return err
	}

	//tb,_ := json.Marshal(pb)
	//fmt.Println(string(tb[:]))

	data.unmarshal(&pb)
	return nil
}

// hasAdminUser exhaustively checks for the presence of at least one admin
// user.
func (data *Data) hasAdminUser() bool {
	for _, u := range data.Data.Users {
		if u.Admin {
			return true
		}
	}
	return false
}


// DataNode returns a node by id.
func (data *Data) DataNode(id uint64) *meta.NodeInfo {
	for i := range data.DataNodes {
		if data.DataNodes[i].ID == id {
			return &data.DataNodes[i]
		}
	}
	return nil
}

// CreateDataNode adds a node to the metadata.
func (data *Data) CreateDataNode(host, tcpHost string) error {
	// Ensure a node with the same host doesn't already exist.
	for _, n := range data.DataNodes {
		if n.TCPHost == tcpHost {
			return ErrNodeExists
		}
	}

	// If an existing meta node exists with the same TCPHost address,
	// then these nodes are actually the same so re-use the existing ID
	var existingID uint64
	for _, n := range data.MetaNodes {
		if n.TCPHost == tcpHost {
			existingID = n.ID
			break
		}
	}

	// We didn't find an existing node, so assign it a new node ID
	if existingID == 0 {
		data.MaxNodeID++
		existingID = data.MaxNodeID
	}

	// Append new node.
	data.DataNodes = append(data.DataNodes, meta.NodeInfo{
		ID:      existingID,
		Host:    host,
		TCPHost: tcpHost,
	})
	sort.Sort(meta.NodeInfos(data.DataNodes))

	return nil
}

// setDataNode adds a data node with a pre-specified nodeID.
// this should only be used when the cluster is upgrading from 0.9 to 0.10
func (data *Data) setDataNode(nodeID uint64, host, tcpHost string) error {
	// Ensure a node with the same host doesn't already exist.
	for _, n := range data.DataNodes {
		if n.Host == host {
			return ErrNodeExists
		}
	}

	// Append new node.
	data.DataNodes = append(data.DataNodes, meta.NodeInfo{
		ID:      nodeID,
		Host:    host,
		TCPHost: tcpHost,
	})

	return nil
}

// DeleteDataNode removes a node from the Meta store.
//
// If necessary, DeleteDataNode reassigns ownership of any shards that
// would otherwise become orphaned by the removal of the node from the
// cluster.
func (data *Data) DeleteDataNode(id uint64) error {
	var nodes []meta.NodeInfo

	// Remove the data node from the store's list.
	for _, n := range data.DataNodes {
		if n.ID != id {
			nodes = append(nodes, n)
		}
	}

	if len(nodes) == len(data.DataNodes) {
		return ErrNodeNotFound
	}
	data.DataNodes = nodes

	// Remove node id from all shard infos
	for di, d := range data.Data.Databases {
		for ri, rp := range d.RetentionPolicies {
			for sgi, sg := range rp.ShardGroups {
				var (
					nodeOwnerFreqs = make(map[int]int)
					orphanedShards []meta.ShardInfo
				)
				// Look through all shards in the shard group and
				// determine (1) if a shard no longer has any owners
				// (orphaned); (2) if all shards in the shard group
				// are orphaned; and (3) the number of shards in this
				// group owned by each data node in the cluster.
				for si, s := range sg.Shards {
					// Track of how many shards in the group are
					// owned by each data node in the cluster.
					var nodeIdx = -1
					for i, owner := range s.Owners {
						if owner.NodeID == id {
							nodeIdx = i
						}
						nodeOwnerFreqs[int(owner.NodeID)]++
					}

					if nodeIdx > -1 {
						// Data node owns shard, so relinquish ownership
						// and set new owners on the shard.
						s.Owners = append(s.Owners[:nodeIdx], s.Owners[nodeIdx+1:]...)
						data.Data.Databases[di].RetentionPolicies[ri].ShardGroups[sgi].Shards[si].Owners = s.Owners
					}

					// Shard no longer owned. Will need reassigning
					// an owner.
					if len(s.Owners) == 0 {
						orphanedShards = append(orphanedShards, s)
					}
				}

				// Mark the shard group as deleted if it has no shards,
				// or all of its shards are orphaned.
				if len(sg.Shards) == 0 || len(orphanedShards) == len(sg.Shards) {
					data.Data.Databases[di].RetentionPolicies[ri].ShardGroups[sgi].DeletedAt = time.Now().UTC()
					continue
				}

				// Reassign any orphaned shards. Delete the node we're
				// dropping from the list of potential new owners.
				delete(nodeOwnerFreqs, int(id))

				for _, orphan := range orphanedShards {
					newOwnerID, err := newShardOwner(orphan, nodeOwnerFreqs)
					if err != nil {
						return err
					}

					for si, s := range sg.Shards {
						if s.ID == orphan.ID {
							sg.Shards[si].Owners = append(sg.Shards[si].Owners, meta.ShardOwner{NodeID: newOwnerID})
							data.Data.Databases[di].RetentionPolicies[ri].ShardGroups[sgi].Shards = sg.Shards
							break
						}
					}

				}
			}
		}
	}
	return nil
}

// newShardOwner sets the owner of the provided shard to the data node
// that currently owns the fewest number of shards. If multiple nodes
// own the same (fewest) number of shards, then one of those nodes
// becomes the new shard owner.
func newShardOwner(s meta.ShardInfo, ownerFreqs map[int]int) (uint64, error) {
	var (
		minId   = -1
		minFreq int
	)

	for id, freq := range ownerFreqs {
		if minId == -1 || freq < minFreq {
			minId, minFreq = int(id), freq
		}
	}

	if minId < 0 {
		return 0, fmt.Errorf("cannot reassign shard %d due to lack of data nodes", s.ID)
	}

	// Update the shard owner frequencies and set the new owner on the
	// shard.
	ownerFreqs[minId]++
	return uint64(minId), nil
}

// MetaNode returns a node by id.
func (data *Data) MetaNode(id uint64) *meta.NodeInfo {
	for i := range data.MetaNodes {
		if data.MetaNodes[i].ID == id {
			return &data.MetaNodes[i]
		}
	}
	return nil
}

// CreateMetaNode will add a new meta node to the metastore
func (data *Data) CreateMetaNode(httpAddr, tcpAddr string) error {
	// Ensure a node with the same host doesn't already exist.
	for _, n := range data.MetaNodes {
		if n.Host == httpAddr {
			return ErrNodeExists
		}
	}

	// If an existing data node exists with the same TCPHost address,
	// then these nodes are actually the same so re-use the existing ID
	var existingID uint64
	for _, n := range data.DataNodes {
		if n.TCPHost == tcpAddr {
			existingID = n.ID
			break
		}
	}

	// We didn't find and existing data node ID, so assign a new ID
	// to this meta node.
	if existingID == 0 {
		data.MaxNodeID++
		existingID = data.MaxNodeID
	}

	// Append new node.
	data.MetaNodes = append(data.MetaNodes, meta.NodeInfo{
		ID:      existingID,
		Host:    httpAddr,
		TCPHost: tcpAddr,
	})

	sort.Sort(meta.NodeInfos(data.MetaNodes))
	return nil
}

// SetMetaNode will update the information for the single meta
// node or create a new metanode. If there are more than 1 meta
// nodes already, an error will be returned
func (data *Data) SetMetaNode(httpAddr, tcpAddr string) error {
	if len(data.MetaNodes) > 1 {
		return fmt.Errorf("can't set meta node when there are more than 1 in the metastore")
	}

	if len(data.MetaNodes) == 0 {
		return data.CreateMetaNode(httpAddr, tcpAddr)
	}

	data.MetaNodes[0].Host = httpAddr
	data.MetaNodes[0].TCPHost = tcpAddr

	return nil
}

// DeleteMetaNode will remove the meta node from the store
func (data *Data) DeleteMetaNode(id uint64) error {
	// Node has to be larger than 0 to be real
	if id == 0 {
		return ErrNodeIDRequired
	}

	var nodes []meta.NodeInfo
	for _, n := range data.MetaNodes {
		if n.ID == id {
			continue
		}
		nodes = append(nodes, n)
	}

	if len(nodes) == len(data.MetaNodes) {
		return ErrNodeNotFound
	}

	data.MetaNodes = nodes
	return nil
}


// migrate
//
//// marshal serializes to a protobuf representation.
func  marshalDatabaseInfo(di meta.DatabaseInfo) *internal.DatabaseInfo {
	pb := &internal.DatabaseInfo{}
	pb.Name = proto.String(di.Name)
	pb.DefaultRetentionPolicy = proto.String(di.DefaultRetentionPolicy)

	pb.RetentionPolicies = make([]*internal.RetentionPolicyInfo, len(di.RetentionPolicies))
	for i := range di.RetentionPolicies {
		pb.RetentionPolicies[i] = marshalRetentionPolicyInfo(&di.RetentionPolicies[i])
	}

	//pb.ContinuousQueries = make([]*internal.ContinuousQueryInfo, len(di.ContinuousQueries))
	//for i := range di.ContinuousQueries {
	//	pb.ContinuousQueries[i] = di.ContinuousQueries[i].marshal()
	//}
	return pb
}

// unmarshal deserializes from a protobuf representation.
func  unmarshalDatabaseInfo(pb *internal.DatabaseInfo) (di *meta.DatabaseInfo) {
	di = &meta.DatabaseInfo{}
	di.Name = pb.GetName()
	di.DefaultRetentionPolicy = pb.GetDefaultRetentionPolicy()

	if len(pb.GetRetentionPolicies()) > 0 {
		di.RetentionPolicies = make([]meta.RetentionPolicyInfo, len(pb.GetRetentionPolicies()))
		for i, x := range pb.GetRetentionPolicies() {
			di.RetentionPolicies[i] = *unmarshalRetentionPolicyInfo(x)
		}
	}

	//if len(pb.GetContinuousQueries()) > 0 {
	//	di.ContinuousQueries = make([]meta.ContinuousQueryInfo, len(pb.GetContinuousQueries()))
	//	for i, x := range pb.GetContinuousQueries() {
	//		di.ContinuousQueries[i].unmarshal(x)
	//	}
	//}

	return di
}

// marshal serializes to a protobuf representation.
func  marshalRetentionPolicyInfo(rpi *meta.RetentionPolicyInfo) *internal.RetentionPolicyInfo {
	pb := &internal.RetentionPolicyInfo{
		Name:               proto.String(rpi.Name),
		ReplicaN:           proto.Uint32(uint32(rpi.ReplicaN)),
		Duration:           proto.Int64(int64(rpi.Duration)),
		ShardGroupDuration: proto.Int64(int64(rpi.ShardGroupDuration)),
	}

	pb.ShardGroups = make([]*internal.ShardGroupInfo, len(rpi.ShardGroups))
	for i, sgi := range rpi.ShardGroups {
		pb.ShardGroups[i] = marshalShardGroupInfo(&sgi)
	}

	//pb.Subscriptions = make([]*internal.SubscriptionInfo, len(rpi.Subscriptions))
	//for i, sub := range rpi.Subscriptions {
	//	pb.Subscriptions[i] = sub.marshal()
	//}

	return pb
}

// unmarshal deserializes from a protobuf representation.
func unmarshalRetentionPolicyInfo(pb *internal.RetentionPolicyInfo)(rpi *meta.RetentionPolicyInfo)  {
	rpi = &meta.RetentionPolicyInfo{}
	rpi.Name = pb.GetName()
	rpi.ReplicaN = int(pb.GetReplicaN())
	rpi.Duration = time.Duration(pb.GetDuration())
	rpi.ShardGroupDuration = time.Duration(pb.GetShardGroupDuration())

	if len(pb.GetShardGroups()) > 0 {
		rpi.ShardGroups = make([]meta.ShardGroupInfo, len(pb.GetShardGroups()))
		for i, x := range pb.GetShardGroups() {
			rpi.ShardGroups[i]= *unmarshalShardGroupInfo(x)
		}
	}
	//if len(pb.GetSubscriptions()) > 0 {
	//	rpi.Subscriptions = make([]meta.SubscriptionInfo, len(pb.GetSubscriptions()))
	//	for i, x := range pb.GetSubscriptions() {
	//		rpi.Subscriptions[i].unmarshal(x)
	//	}
	//}

	return rpi
}

// marshal serializes to a protobuf representation.
func marshalShardGroupInfo(sgi *meta.ShardGroupInfo)  *internal.ShardGroupInfo {
	pb := &internal.ShardGroupInfo{
		ID:        proto.Uint64(sgi.ID),
		StartTime: proto.Int64(meta.MarshalTime(sgi.StartTime)),
		EndTime:   proto.Int64(meta.MarshalTime(sgi.EndTime)),
		DeletedAt: proto.Int64(meta.MarshalTime(sgi.DeletedAt)),
	}

	if !sgi.TruncatedAt.IsZero() {
		pb.TruncatedAt = proto.Int64(meta.MarshalTime(sgi.TruncatedAt))
	}

	pb.Shards = make([]*internal.ShardInfo, len(sgi.Shards))
	for i := range sgi.Shards {
		pb.Shards[i] = marshalShardInfo(sgi.Shards[i])
	}

	return pb
}

// unmarshal deserializes from a protobuf representation.
func  unmarshalShardGroupInfo(pb *internal.ShardGroupInfo) (sgi *meta.ShardGroupInfo) {
	sgi = &meta.ShardGroupInfo{}
	sgi.ID = pb.GetID()
	if i := pb.GetStartTime(); i == 0 {
		sgi.StartTime = time.Unix(0, 0).UTC()
	} else {
		sgi.StartTime = meta.UnmarshalTime(i)
	}
	if i := pb.GetEndTime(); i == 0 {
		sgi.EndTime = time.Unix(0, 0).UTC()
	} else {
		sgi.EndTime = meta.UnmarshalTime(i)
	}
	sgi.DeletedAt = meta.UnmarshalTime(pb.GetDeletedAt())

	if pb != nil && pb.TruncatedAt != nil {
		sgi.TruncatedAt = meta.UnmarshalTime(pb.GetTruncatedAt())
	}

	if len(pb.GetShards()) > 0 {
		sgi.Shards = make([]meta.ShardInfo, len(pb.GetShards()))
		for i, x := range pb.GetShards() {
			sgi.Shards[i]= *unmarshalShardInfo(x)
		}
	}

	return sgi
}

// marshal serializes to a protobuf representation.
func  marshalShardInfo(si meta.ShardInfo) *internal.ShardInfo {
	pb := &internal.ShardInfo{
		ID: proto.Uint64(si.ID),
	}

	pb.Owners = make([]*internal.ShardOwner, len(si.Owners))
	for i := range si.Owners {
		pb.Owners[i] = marshalShardOwner(si.Owners[i])
	}

	return pb
}

func unmarshalShardInfo(pb *internal.ShardInfo) (si *meta.ShardInfo) {

	si = &meta.ShardInfo{}
	si.ID = pb.GetID()

	// If deprecated "OwnerIDs" exists then convert it to "Owners" format.
	if len(pb.GetOwnerIDs()) > 0 {
		si.Owners = make([]meta.ShardOwner, len(pb.GetOwnerIDs()))
		for i, x := range pb.GetOwnerIDs() {
			si.Owners[i]= *unmarshalShardOwner(&internal.ShardOwner{
				NodeID: proto.Uint64(x),
			})
		}
	} else if len(pb.GetOwners()) > 0 {
		si.Owners = make([]meta.ShardOwner, len(pb.GetOwners()))
		for i, x := range pb.GetOwners() {
			si.Owners[i]= *unmarshalShardOwner(x)
		}
	}
	return si
}


// marshal serializes to a protobuf representation.
func marshalShardOwner(so meta.ShardOwner)  *internal.ShardOwner {
	return &internal.ShardOwner{
		NodeID: proto.Uint64(so.NodeID),
	}
}

// unmarshal deserializes from a protobuf representation.
func  unmarshalShardOwner(pb *internal.ShardOwner) (so *meta.ShardOwner) {
	so = &meta.ShardOwner{}
	so.NodeID = pb.GetNodeID()
	return so
}

// marshal serializes to a protobuf representation.
func marshalNodeInfo(so meta.NodeInfo)  *internal.NodeInfo {
	return &internal.NodeInfo{
		ID: proto.Uint64(so.ID),
		Host: proto.String(so.Host),
		TCPHost: proto.String(so.TCPHost),

	}
}


// unmarshal deserializes from a protobuf representation.
func  unmarshalNodeInfo(pb *internal.NodeInfo) (so *meta.NodeInfo) {
	so = &meta.NodeInfo{}
	so.ID = pb.GetID()
	so.TCPHost = pb.GetTCPHost()
	so.Host = pb.GetHost()
	return so
}


func (data *Data) user(username string) *meta.UserInfo {
	for i := range data.Data.Users {
		if data.Data.Users[i].Name == username {
			return &data.Data.Users[i]
		}
	}
	return nil
}