package grpcapi_test

import (
	"context"
	"log/slog"
	"net"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	greeterv1 "github.com/ravilushqa/boilerplate/api/greeter/v1"
	"github.com/ravilushqa/boilerplate/internal/greeter"
	"github.com/ravilushqa/boilerplate/internal/grpcapi"
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
	require.NoError(t, err)

	t.Cleanup(func() { require.NoError(t, conn.Close()) })

	return conn
}

func TestGreet(t *testing.T) {
	t.Parallel()

	client := greeterv1.NewGreeterServiceClient(dial(t, greeter.New()))

	for name, tc := range map[string]struct {
		input    string
		want     string
		wantCode codes.Code
	}{
		"OK":    {input: "World", want: "Hello, World!", wantCode: codes.OK},
		"Empty": {input: "", wantCode: codes.InvalidArgument},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			resp, err := client.Greet(t.Context(), greeterv1.GreetRequest_builder{Name: tc.input}.Build())
			require.Equal(t, tc.wantCode, status.Code(err), "%v", err)

			if tc.wantCode == codes.OK {
				require.Equal(t, tc.want, resp.GetMessage())
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
	require.Equal(t, codes.Internal, status.Code(err))
}

func TestHealth(t *testing.T) {
	t.Parallel()

	resp, err := healthpb.NewHealthClient(dial(t, greeter.New())).Check(t.Context(), &healthpb.HealthCheckRequest{})
	require.NoError(t, err)
	require.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.GetStatus())
}
