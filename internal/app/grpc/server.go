package grpc

import (
	"context"
	"log/slog"
	"net"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcrecovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpcprometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/ravilushqa/boilerplate/api"
)

type Server struct {
	api.GreeterServer
	l    *slog.Logger
	addr string
}

func New(l *slog.Logger, addr string) *Server {
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
			grpcrecovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpcmiddleware.ChainUnaryServer(
			grpcprometheus.UnaryServerInterceptor,
			grpcrecovery.UnaryServerInterceptor(),
		)),
	)
	grpcprometheus.EnableHandlingTimeHistogram()

	api.RegisterGreeterServer(grpcSrv, s)

	reflection.Register(grpcSrv)

	go func() {
		<-ctx.Done()
		grpcSrv.GracefulStop()
		s.l.Info("[GRPC] server stopping", slog.String("addr", s.addr))
	}()

	s.l.Info("[GRPC] server listening", slog.String("addr", s.addr))

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
