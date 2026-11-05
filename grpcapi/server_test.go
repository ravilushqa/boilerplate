package grpcapi_test

import (
	"context"
	"log/slog"
	"net"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	greeterv1 "github.com/ravilushqa/boilerplate/api/greeter/v1"
	"github.com/ravilushqa/boilerplate/greeter"
	"github.com/ravilushqa/boilerplate/grpcapi"
)

// greeterFunc adapts a function to the grpcapi.Greeter interface.
type greeterFunc func(ctx context.Context, name string) (string, error)

func (f greeterFunc) Greet(ctx context.Context, name string) (string, error) { return f(ctx, name) }

// dial starts an in-memory gRPC server and returns a client connection to it.
func dial(t *testing.T, g grpcapi.Greeter) *grpc.ClientConn {
	t.Helper()

	srv := grpcapi.NewServer(grpcapi.Deps{
		Logger:  slog.New(slog.NewTextHandler(t.Output(), nil)),
		Metrics: prometheus.NewRegistry(),
		Health:  health.NewServer(),
		Greeter: g,
	})

	lis := bufconn.Listen(1 << 20)

	go func() { _ = srv.Serve(lis) }()

	t.Cleanup(srv.Stop)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { _ = conn.Close() })

	return conn
}

func TestGreet(t *testing.T) {
	t.Parallel()

	client := greeterv1.NewGreeterServiceClient(dial(t, &greeter.Service{}))

	tests := []struct {
		name     string
		input    string
		want     string
		wantCode codes.Code
	}{
		{name: "ok", input: "World", want: "Hello, World!", wantCode: codes.OK},
		{name: "empty name", input: "", wantCode: codes.InvalidArgument},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.Greet(t.Context(), greeterv1.GreetRequest_builder{Name: tt.input}.Build())
			if code := status.Code(err); code != tt.wantCode {
				t.Fatalf("Greet(%q) code = %v, want %v (err: %v)", tt.input, code, tt.wantCode, err)
			}

			if got := resp.GetMessage(); got != tt.want {
				t.Errorf("Greet(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGreetPanic(t *testing.T) {
	t.Parallel()

	client := greeterv1.NewGreeterServiceClient(dial(t, greeterFunc(func(context.Context, string) (string, error) {
		panic("boom")
	})))

	_, err := client.Greet(t.Context(), greeterv1.GreetRequest_builder{Name: "x"}.Build())
	if code := status.Code(err); code != codes.Internal {
		t.Errorf("code = %v, want %v", code, codes.Internal)
	}
}

func TestHealth(t *testing.T) {
	t.Parallel()

	resp, err := healthpb.NewHealthClient(dial(t, &greeter.Service{})).Check(t.Context(), &healthpb.HealthCheckRequest{})
	if err != nil {
		t.Fatal(err)
	}

	if got := resp.GetStatus(); got != healthpb.HealthCheckResponse_SERVING {
		t.Errorf("status = %v, want SERVING", got)
	}
}
