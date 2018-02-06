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
	s.Logger.Info(fmt.Sprintf("got restore request for shard %d, from node %d",
		     req.ShardID, req.GetFromNodeId()))

	fromNode, _ := s.MetaClient.DataNode(req.FromNodeId)
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(utils.GetAdminHost(fromNode.TCPHost), opts...)
	if err != nil {
		fmt.Println("dial,",err)
	}
	defer conn.Close()

	client := proto.NewRestoreServiceClient(conn)

	backupReq := &proto.ShardSnapshotRequest{
		Database: req.Database,
		RetentionPolicy: req.RetentionPolicy,
		ShardID: req.ShardID,
		Since: 0,
	}
	stream , _ := client.ShardBackup(context.Background(), backupReq)

	r, w := io.Pipe()
	defer r.Close()
	go func(){
		defer w.Close()
		for {
			resp ,err := stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				s.Logger.Error("read stream interupted,", zap.String("err",err.Error()))
				break
			}

			n, e := w.Write(resp.FileBlock)
		//	s.Logger.Info(fmt.Sprintf("write to reader pipe, %d ", n))
			if e != nil {
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
	shard := s.TSDBStore.Shard(req.ShardID)
	if shard == nil {
		s.TSDBStore.CreateShard(req.Database, req.RetentionPolicy,req.ShardID,true)
	}
	if err :=s.TSDBStore.RestoreShard(req.ShardID, r); err!= nil{
		//TODO modify restore or use inmem index
		//if err.Error() == "can only read from tsm file" {
		//	s.Logger.Error(fmt.Sprintf("restore read backup error, %s", err.Error()))
		//	return resp, nil
		//}
		s.Logger.Error(fmt.Sprintf("restore read backup error, %s", err.Error()))
		resp.Success = false
		resp.Message = err.Error()
		return resp, err
	}
	return resp, nil
}

func (s *Service) ShardBackup(req *proto.ShardSnapshotRequest, srv proto.RestoreService_ShardBackupServer) error{
	//TODO 1. time convert 2.write target choose
	s.Logger.Info(fmt.Sprintf("got backup request for shard %d, since %d", req.ShardID, req.Since))
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
			s.Logger.Error(fmt.Sprintf("backup error, %s", err.Error()))
		}
	}()

	buf := make ([]byte, 4096)
	length := 0
	for {
		n, err  := r.Read(buf)
		if n == 0 || err == io.EOF{
			break
		}

		//s.Logger.Info(fmt.Sprintf("read %d, content=%v", n, buf[:n]))
		srv.Send(&proto.ShardSnapshotResponse{
			FileBlock: buf[:n],
		})

		length += n
	}

	s.Logger.Info(fmt.Sprintf("got backup request for shard %d is done. size=%d", req.ShardID, length))
	return nil
}



