package app_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	greeterv1 "github.com/ravilushqa/boilerplate/api/greeter/v1"
	"github.com/ravilushqa/boilerplate/internal/app"
	"github.com/ravilushqa/boilerplate/internal/config"
)

// TestApp starts the whole service on random ports, calls every server
// and checks that it shuts down cleanly.
func TestApp(t *testing.T) {
	t.Parallel()

	cfg, err := config.Parse([]string{
		"--env=test",
		"--http-addr=127.0.0.1:0",
		"--grpc-addr=127.0.0.1:0",
		"--debug-addr=127.0.0.1:0",
		"--shutdown-timeout=5s",
	})
	require.NoError(t, err)

	a, err := app.New(cfg, slog.New(slog.NewTextHandler(t.Output(), nil)))
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)

	go func() { done <- a.Run(ctx) }()

	// Wait until the service reports readiness.
	require.Eventually(t, func() bool {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+a.DebugAddr()+"/readyz", http.NoBody)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return false
		}

		_ = resp.Body.Close()

		return resp.StatusCode == http.StatusOK
	}, 5*time.Second, 10*time.Millisecond)

	t.Run("HTTP", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, "http://"+a.HTTPAddr()+"/v1/greet", strings.NewReader(`{"name":"World"}`))
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.JSONEq(t, `{"message":"Hello, World!"}`, string(body))
	})

	t.Run("GRPC", func(t *testing.T) {
		conn, err := grpc.NewClient(a.GRPCAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)

		defer conn.Close()

		resp, err := greeterv1.NewGreeterServiceClient(conn).Greet(t.Context(), greeterv1.GreetRequest_builder{Name: "World"}.Build())
		require.NoError(t, err)
		require.Equal(t, "Hello, World!", resp.GetMessage())
	})

	t.Run("Metrics", func(t *testing.T) {
		req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+a.DebugAddr()+"/metrics", http.NoBody)
		require.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(body), `http_server_request_duration_seconds_count{code="200",method="POST",route="POST /v1/greet"} 1`)
		require.Contains(t, string(body), `grpc_server_handled_total{grpc_code="OK",grpc_method="Greet",grpc_service="greeter.v1.GreeterService",grpc_type="unary"} 1`)
	})

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(10 * time.Second):
		t.Fatal("service did not stop in time")
	}
}

func TestNewPortInUse(t *testing.T) {
	t.Parallel()

	cfg, err := config.Parse([]string{"--http-addr=127.0.0.1:0", "--grpc-addr=127.0.0.1:0", "--debug-addr=127.0.0.1:0"})
	require.NoError(t, err)

	first, err := app.New(cfg, slog.New(slog.DiscardHandler))
	require.NoError(t, err)

	cfg.GRPC.Addr = first.GRPCAddr()

	_, err = app.New(cfg, slog.New(slog.DiscardHandler))
	require.ErrorContains(t, err, "grpc: listen on")
}
