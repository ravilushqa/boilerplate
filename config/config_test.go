package config_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/ravilushqa/boilerplate/config"
)

// parse calls config.Parse and fails the test on error.
func parse(t *testing.T, args ...string) *config.Config {
	t.Helper()

	cfg, err := config.Parse(args)
	if err != nil {
		t.Fatalf("Parse(%q): %v", args, err)
	}

	return cfg
}

// ignoreVersion skips the --version flag field, which has no value to compare.
var ignoreVersion = cmpopts.IgnoreFields(config.Config{}, "Version")

func TestParseDefaults(t *testing.T) {
	want := &config.Config{
		Env:             "development",
		Log:             config.Log{Level: "info", Format: "json"},
		HTTP:            config.HTTP{Addr: ":8080"},
		GRPC:            config.GRPC{Addr: ":50051"},
		Debug:           config.Debug{Addr: ":8081"},
		ShutdownTimeout: 15 * time.Second,
	}

	if diff := cmp.Diff(want, parse(t), ignoreVersion); diff != "" {
		t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
	}
}

func TestParseEnv(t *testing.T) {
	// t.Setenv forbids t.Parallel: the environment is process-wide.
	t.Setenv("ENV", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")

	cfg := parse(t)

	if cfg.Env != "production" || cfg.Log.Level != "debug" || cfg.HTTP.Addr != ":9090" || cfg.ShutdownTimeout != 3*time.Second {
		t.Errorf("environment not applied: %+v", cfg)
	}
}

func TestParseFlagsOverrideEnv(t *testing.T) {
	t.Setenv("GRPC_ADDR", ":1111")

	cfg := parse(t, "--grpc-addr=:2222", "--log-format=text")

	if cfg.GRPC.Addr != ":2222" {
		t.Errorf("GRPC.Addr = %q, want %q", cfg.GRPC.Addr, ":2222")
	}

	if cfg.Log.Format != "text" {
		t.Errorf("Log.Format = %q, want %q", cfg.Log.Format, "text")
	}
}

func TestParseInvalid(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "unknown env", args: []string{"--env=staging"}},
		{name: "unknown level", args: []string{"--log-level=verbose"}},
		{name: "bad duration", args: []string{"--shutdown-timeout=soon"}},
		{name: "unknown flag", args: []string{"--no-such-flag"}},
		{name: "positional argument", args: []string{"extra"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if _, err := config.Parse(tt.args); err == nil {
				t.Errorf("Parse(%q) succeeded, want an error", tt.args)
			}
		})
	}
}
