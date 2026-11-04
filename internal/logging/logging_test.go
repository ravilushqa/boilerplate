package logging_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/internal/config"
	"github.com/ravilushqa/boilerplate/internal/logging"
)

func TestNew(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	l, err := logging.New(&buf, config.Log{Level: "warn", Format: "json"})
	require.NoError(t, err)

	l.Info("hidden")
	l.Warn("shown", "key", "value")

	var entry map[string]any
	require.NoError(t, json.Unmarshal(buf.Bytes(), &entry), "exactly one JSON line expected, got: %s", buf.String())
	require.Equal(t, "shown", entry["msg"])
	require.Equal(t, "value", entry["key"])
}

func TestNewInvalid(t *testing.T) {
	t.Parallel()

	_, err := logging.New(&bytes.Buffer{}, config.Log{Level: "loud", Format: "json"})
	require.Error(t, err)

	_, err = logging.New(&bytes.Buffer{}, config.Log{Level: "info", Format: "xml"})
	require.Error(t, err)
}
