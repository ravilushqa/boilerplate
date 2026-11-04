package config_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/internal/config"
)

func TestParseDefaults(t *testing.T) {
	cfg, err := config.Parse(nil)
	require.NoError(t, err)

	require.Equal(t, "development", cfg.Env)
	require.Equal(t, config.Log{Level: "info", Format: "json"}, cfg.Log)
	require.Equal(t, ":8080", cfg.HTTP.Addr)
	require.Equal(t, ":50051", cfg.GRPC.Addr)
	require.Equal(t, ":8081", cfg.Debug.Addr)
	require.Equal(t, 15*time.Second, cfg.ShutdownTimeout)
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
	for name, args := range map[string][]string{
		"UnknownEnv":      {"--env=staging"},
		"UnknownLevel":    {"--log-level=verbose"},
		"BadDuration":     {"--shutdown-timeout=soon"},
		"UnknownFlag":     {"--no-such-flag"},
		"PositionalInput": {"extra"},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := config.Parse(args)
			require.Error(t, err)
		})
	}
}
