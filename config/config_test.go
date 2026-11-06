package config_test

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/config"
)

func TestParseDefaults(t *testing.T) {
	cfg, err := config.Parse(nil)
	require.NoError(t, err)

	want := &config.Config{
		Env:             "development",
		Log:             config.Log{Level: "info", Format: "json"},
		HTTP:            config.HTTP{Addr: ":8080"},
		GRPC:            config.GRPC{Addr: ":50051"},
		Debug:           config.Debug{Addr: ":8081"},
		ShutdownTimeout: 15 * time.Second,
	}

	if diff := cmp.Diff(want, cfg); diff != "" {
		t.Errorf("Parse() mismatch (-want +got):\n%s", diff)
	}
}

func TestParseEnv(t *testing.T) {
	// t.Setenv forbids t.Parallel: the environment is process-wide.
	t.Setenv("ENV", "production")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("HTTP_ADDR", ":9090")
	t.Setenv("SHUTDOWN_TIMEOUT", "3s")

	cfg, err := config.Parse(nil)
	require.NoError(t, err)

	require.Equal(t, "production", cfg.Env)
	require.Equal(t, "debug", cfg.Log.Level)
	require.Equal(t, ":9090", cfg.HTTP.Addr)
	require.Equal(t, 3*time.Second, cfg.ShutdownTimeout)
}

func TestParseFlagsOverrideEnv(t *testing.T) {
	t.Setenv("GRPC_ADDR", ":1111")

	cfg, err := config.Parse([]string{"--grpc-addr=:2222", "--log-format=text"})
	require.NoError(t, err)

	require.Equal(t, ":2222", cfg.GRPC.Addr)
	require.Equal(t, "text", cfg.Log.Format)
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

			_, err := config.Parse(tt.args)
			require.Error(t, err)
		})
	}
}
