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
	if err != nil {
		t.Fatal(err)
	}

	a, err := newApp(cfg, slog.New(slog.NewTextHandler(t.Output(), nil)))
	if err != nil {
		t.Fatal(err)
	}

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
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

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
		if code != http.StatusOK || strings.TrimSpace(body) != `{"message":"Hello, World!"}` {
			t.Errorf("POST /v1/greet = %d %s", code, body)
		}
	})

	t.Run("GRPC", func(t *testing.T) {
		conn, err := grpc.NewClient(a.grpcAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			t.Fatal(err)
		}

		defer conn.Close()

		resp, err := greeterv1.NewGreeterServiceClient(conn).Greet(t.Context(), greeterv1.GreetRequest_builder{Name: "World"}.Build())
		if err != nil {
			t.Fatal(err)
		}

		if got := resp.GetMessage(); got != "Hello, World!" {
			t.Errorf("Greet = %q", got)
		}
	})

	t.Run("Metrics", func(t *testing.T) {
		_, body := do(t, http.MethodGet, "http://"+a.debugAddr()+"/metrics", "")

		for _, want := range []string{
			`http_server_request_duration_seconds_count{code="200",method="POST",route="POST /v1/greet"} 1`,
			`grpc_server_handled_total{grpc_code="OK",grpc_method="Greet",grpc_service="greeter.v1.GreeterService",grpc_type="unary"} 1`,
		} {
			if !strings.Contains(body, want) {
				t.Errorf("/metrics does not contain %s", want)
			}
		}
	})

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("run() = %v, want nil after a clean shutdown", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("service did not stop in time")
	}
}

func TestNewAppPortInUse(t *testing.T) {
	t.Parallel()

	first := newTestApp(t)

	cfg, err := config.Parse([]string{"--http-addr=127.0.0.1:0", "--grpc-addr=" + first.grpcAddr(), "--debug-addr=127.0.0.1:0"})
	if err != nil {
		t.Fatal(err)
	}

	_, err = newApp(cfg, slog.New(slog.DiscardHandler))
	if err == nil || !strings.Contains(err.Error(), "grpc: listening on") {
		t.Errorf("newApp() error = %v, want a gRPC listen error", err)
	}
}

func TestNewLogger(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	l, err := newLogger(&buf, config.Log{Level: "warn", Format: "json"})
	if err != nil {
		t.Fatal(err)
	}

	l.Info("hidden")
	l.Warn("shown", "key", "value")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("want exactly one JSON line, got %q: %v", buf.String(), err)
	}

	if entry["msg"] != "shown" || entry["key"] != "value" {
		t.Errorf("unexpected entry: %v", entry)
	}
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

			if _, err := newLogger(io.Discard, tt.cfg); err == nil {
				t.Errorf("newLogger(%+v) succeeded, want an error", tt.cfg)
			}
		})
	}
}
