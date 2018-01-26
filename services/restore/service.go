package restore

import (
	"time"

	"github.com/influxdata/influxdb/tsdb"
	"github.com/uber-go/zap"
	"google.golang.org/grpc"
	"github.com/colinsage/myts/grpc/proto"
	"io"
	"fmt"
	"context"
	"github.com/influxdata/influxdb/services/meta"
	"github.com/colinsage/myts/utils"
)

// Service manages the listener for the snapshot endpoint.
type Service struct {

	Node *meta.NodeInfo

	MetaClient interface {
		ShardInfo(shardID uint64) meta.ShardInfo
		DataNode(id uint64) (*meta.NodeInfo, error)
	}
	TSDBStore *tsdb.Store

	Logger   zap.Logger

}

// NewService returns a new instance of Service.
func NewService() *Service {
	return &Service{
		Logger: zap.New(zap.NullEncoder()),
	}
}

// Open starts the service.
func (s *Service) Register(server *grpc.Server) error{
	s.Logger.Info("register snapshot service")

	proto.RegisterRestoreServiceServer(server, s)

	return nil
}

// WithLogger sets the logger on the service.
func (s *Service) WithLogger(log zap.Logger) {
	s.Logger = log.With(zap.String("service", "restore"))
}

func (s *Service) ShardRestore(cxt context.Context, req *proto.ShardRestoreRequest) (*proto.ShardRestoreResponse, error){

	now := time.Now()

	var remote string
	// 1. get backup to local
	shardInfo := s.MetaClient.ShardInfo(req.ShardID)
	for _, owner := range shardInfo.Owners {
		if owner.NodeID != s.Node.ID {
			remoteNode,_ := s.MetaClient.DataNode(owner.NodeID)
			remote = utils.GetAdminHost(remoteNode.TCPHost)
		}
	}
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(remote, opts...)
	if err != nil {
		fmt.Println("dial,",err)
	}
	defer conn.Close()

	client := proto.NewRestoreServiceClient(conn)

	backupReq := &proto.ShardSnapshotRequest{
		Database: req.Database,
		RetentionPolicy: req.RetentionPolicy,
		ShardID: req.ShardID,
		Since: now.Unix(),
	}
	stream , _ := client.ShardBackup(context.Background(), backupReq)

	r, w := io.Pipe()
	defer r.Close()
	defer w.Close()

	go func(){
		for {
			resp ,err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				s.Logger.Error("read stream,", zap.String("err",err.Error()))
			}

			n, e := w.Write(resp.FileBlock)
			if e!= nil {
				s.Logger.Error("write file,", zap.String("err",err.Error()))
			}else if n==0 {
				s.Logger.Warn("write file 0 bytes!")
			}

		}
	}()
	// load to store
	resp := &proto.ShardRestoreResponse{
		Success: true,
	}
	if err :=s.TSDBStore.RestoreShard(req.ShardID, r); err!= nil{
		s.Logger.Error(err.Error())
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	return resp, nil
}

func (s *Service) ShardBackup(req *proto.ShardSnapshotRequest, srv proto.RestoreService_ShardBackupServer) error{
	//TODO 1. time convert 2.write target choose
	since:= time.Unix(req.Since,0)
	id := req.ShardID

	shard := s.TSDBStore.Shard(id)
	if shard == nil {
		return fmt.Errorf("shard %d doesn't exist on this server", id)
	}

	r, w := io.Pipe()
	defer r.Close()

	go func(){
		defer w.Close()
		if err := s.TSDBStore.BackupShard(req.ShardID, since, w); err != nil {
			s.Logger.Error(err.Error())
		}
	}()

	buf := make ([]byte, 4096)

	for {
		n, err  := r.Read(buf)
		if n == 0 || err == io.EOF{
			break
		}
		srv.Send(&proto.ShardSnapshotResponse{
			FileBlock: buf[:n],
		})
	}

	return nil
}



