package grpc

import (
	"github.com/uber-go/zap"
	"net"
	"google.golang.org/grpc"
)

type Service struct {
    Services []RpcService
    Address  string
    server  *grpc.Server
	Logger   zap.Logger
}

type RpcService interface {
	Register(s *grpc.Server) error
	WithLogger(log zap.Logger)
}

// NewService returns a new instance of Service.
func NewService() *Service {
	return &Service{
		Logger: zap.New(zap.NullEncoder()),
	}
}


// Open starts the service.
func (s *Service) Open() error {
	s.Logger.Info("Starting grpc service")

	grpcServer := grpc.NewServer()

	//register
	for _, srv := range s.Services {
		srv.Register(grpcServer)
		srv.WithLogger(s.Logger)
	}

	ln, err := net.Listen("tcp", s.Address)
	if err != nil {
		s.Logger.Error(err.Error())
	}

	s.server = grpcServer
	go grpcServer.Serve(ln)
	return nil
}

// Close implements the Service interface.
func (s *Service) Close() error {
	s.server.GracefulStop()
	return nil
}

// WithLogger sets the logger on the service.
func (s *Service) WithLogger(log zap.Logger) {
	s.Logger = log.With(zap.String("service", "grpc"))
}



