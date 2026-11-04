// Package grpcapi implements the public gRPC API.
//
// Handlers are thin: they validate transport-level input, call business
// logic and map its errors to gRPC status codes. No business logic here.
package grpcapi

import (
	"context"
	"errors"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	prom "github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	greeterv1 "github.com/ravilushqa/boilerplate/api/greeter/v1"
	"github.com/ravilushqa/boilerplate/internal/greeter"
)

// Greeter is the business logic the API depends on.
// It is declared here, where it is used, so tests can substitute it.
type Greeter interface {
	Greet(ctx context.Context, name string) (string, error)
}

// Deps are the dependencies of the gRPC server.
type Deps struct {
	Logger  *slog.Logger
	Metrics prom.Registerer
	Health  *health.Server
	Greeter Greeter
}

// NewServer returns a gRPC server with all services and interceptors registered.
// The caller owns its lifecycle (Serve / GracefulStop).
func NewServer(d Deps) *grpc.Server {
	metrics := prometheus.NewServerMetrics(prometheus.WithServerHandlingTimeHistogram())
	d.Metrics.MustRegister(metrics)

	l := d.Logger.With("component", "grpc")
	logger := logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...) //nolint:sloglint // adapter: messages come from the interceptor
	})
	logOpts := []logging.Option{logging.WithLogOnEvents(logging.FinishCall)}

	recoveryOpts := []recovery.Option{
		recovery.WithRecoveryHandlerContext(func(ctx context.Context, p any) error {
			l.ErrorContext(ctx, "panic recovered", "panic", p)
			return status.Error(codes.Internal, "internal error")
		}),
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			metrics.UnaryServerInterceptor(),
			logging.UnaryServerInterceptor(logger, logOpts...),
			recovery.UnaryServerInterceptor(recoveryOpts...),
		),
		grpc.ChainStreamInterceptor(
			metrics.StreamServerInterceptor(),
			logging.StreamServerInterceptor(logger, logOpts...),
			recovery.StreamServerInterceptor(recoveryOpts...),
		),
	)

	greeterv1.RegisterGreeterServiceServer(srv, &greeterServer{greeter: d.Greeter})
	healthpb.RegisterHealthServer(srv, d.Health)
	reflection.Register(srv)

	metrics.InitializeMetrics(srv)

	return srv
}

// greeterServer implements greeterv1.GreeterServiceServer.
type greeterServer struct {
	greeterv1.UnimplementedGreeterServiceServer

	greeter Greeter
}

func (s *greeterServer) Greet(ctx context.Context, req *greeterv1.GreetRequest) (*greeterv1.GreetResponse, error) {
	msg, err := s.greeter.Greet(ctx, req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	return greeterv1.GreetResponse_builder{Message: msg}.Build(), nil
}

// toStatus maps business errors to gRPC status errors.
// Unknown errors become codes.Internal without leaking details to clients.
func toStatus(err error) error {
	switch {
	case errors.Is(err, greeter.ErrEmptyName), errors.Is(err, greeter.ErrNameTooLong):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
