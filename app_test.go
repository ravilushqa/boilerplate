package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/ravilushqa/boilerplate/config"
)

// newTestApp builds the service on random local ports.
func newTestApp(t *testing.T) *app {
	t.Helper()

	cfg, err := config.Parse([]string{
		"--env=test",
		"--http-addr=127.0.0.1:0",
		"--grpc-addr=127.0.0.1:0",
		"--debug-addr=127.0.0.1:0",
		"--shutdown-timeout=5s",
	})
	require.NoError(t, err)

	a, err := newApp(cfg, slog.New(slog.NewTextHandler(t.Output(), nil)))
	require.NoError(t, err)

	// Release ports of an app that never ran (after run, they are already closed).
	t.Cleanup(func() {
		_ = a.httpLis.Close()
		_ = a.grpcLis.Close()
		_ = a.debugLis.Close()
	})

	return a
}

// do sends a request and returns the status code and body.
func do(t *testing.T, method, url, body string) (int, string) {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), method, url, strings.NewReader(body))
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, string(b)
}

// TestApp starts the whole service, calls every server and checks that it
// shuts down cleanly. Listeners are bound in newApp, so requests queue in
// the backlog until the servers start: no polling is needed.
func TestApp(t *testing.T) {
	t.Parallel()

	a := newTestApp(t)

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)

	go func() { done <- a.run(ctx) }()

	t.Run("HTTP", func(t *testing.T) {
		code, body := do(t, http.MethodPost, "http://"+a.httpAddr()+"/v1/greet", `{"name":"World"}`)
		require.Equal(t, http.StatusOK, code)
		require.JSONEq(t, `{"message":"Hello, World!"}`, body)
	})

	t.Run("GRPC", func(t *testing.T) {
		conn, err := grpc.NewClient(a.grpcAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)

		defer conn.Close()

		resp, err := greeterv1.NewGreeterServiceClient(conn).Greet(t.Context(), greeterv1.GreetRequest_builder{Name: "World"}.Build())
		require.NoError(t, err)
		require.Equal(t, "Hello, World!", resp.GetMessage())
	})

	t.Run("Metrics", func(t *testing.T) {
		_, body := do(t, http.MethodGet, "http://"+a.debugAddr()+"/metrics", "")

		require.Contains(t, body, `http_server_request_duration_seconds_count{code="200",method="POST",route="POST /v1/greet"} 1`)
		require.Contains(t, body, `grpc_server_handled_total{grpc_code="OK",grpc_method="Greet",grpc_service="greeter.v1.GreeterService",grpc_type="unary"} 1`)
	})

	cancel()

	select {
	case err := <-done:
		require.NoError(t, err, "run() must return nil after a clean shutdown")
	case <-time.After(10 * time.Second):
		t.Fatal("service did not stop in time")
	}
}

func TestNewAppPortInUse(t *testing.T) {
	t.Parallel()

	first := newTestApp(t)

	cfg, err := config.Parse([]string{"--http-addr=127.0.0.1:0", "--grpc-addr=" + first.grpcAddr(), "--debug-addr=127.0.0.1:0"})
	require.NoError(t, err)

	_, err = newApp(cfg, slog.New(slog.DiscardHandler))
	require.ErrorContains(t, err, "grpc: listening on")
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	l, err := newLogger(&buf, config.Log{Level: "warn", Format: "json"})
	require.NoError(t, err)

	l.Info("hidden")
	l.Warn("shown", "key", "value")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry), "want exactly one JSON line, got %q", buf.String())
	require.Equal(t, "shown", entry["msg"])
	require.Equal(t, "value", entry["key"])
}

func TestNewLoggerInvalid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  config.Log
	}{
		{name: "unknown level", cfg: config.Log{Level: "loud", Format: "json"}},
		{name: "unknown format", cfg: config.Log{Level: "info", Format: "xml"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := newLogger(io.Discard, tt.cfg)
			require.Error(t, err)
		})
	}
}
