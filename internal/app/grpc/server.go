package grpc

import (
	"context"
	"net"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/ravilushqa/boilerplate/api"
)

type Server struct {
	api.GreeterServer
	l    *zap.Logger
	addr string
}

func New(l *zap.Logger, addr string) *Server {
	return &Server{l: l, addr: addr}
}

func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	grpcSrv := grpc.NewServer(
		grpc.StreamInterceptor(grpcmiddleware.ChainStreamServer(
			grpcprometheus.StreamServerInterceptor,
			grpczap.StreamServerInterceptor(s.l.Named("grpc_stream")),
			grpcrecovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(
			grpcprometheus.UnaryServerInterceptor,
			grpczap.UnaryServerInterceptor(s.l.Named("grpc_unary")),
			grpcrecovery.UnaryServerInterceptor(),
		)),
	)
	grpcprometheus.EnableHandlingTimeHistogram()

	api.RegisterGreeterServer(grpcSrv, s)

	reflection.Register(grpcSrv)

	go func() {
		<-ctx.Done()
		grpcSrv.GracefulStop()
		s.l.Info("[GRPC] server stopping", zap.String("addr", s.addr))
	}()

	s.l.Info("[GRPC] server listening", zap.String("addr", s.addr))

	return grpcSrv.Serve(lis)
}

func (s *Server) Greet(_ context.Context, r *api.GreetRequest) (*api.GreetResponse, error) {
	if r.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name cannot be empty")
	}
	return &api.GreetResponse{
		Message: "Hello " + r.Name,
	}, nil
}
