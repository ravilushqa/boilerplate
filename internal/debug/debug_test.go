package debug_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/internal/debug"
	"github.com/ravilushqa/boilerplate/internal/version"
)

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil))

	return rec
}

func TestHandler(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())

	var ready atomic.Bool
	h := debug.NewHandler(reg, ready.Load)

	require.Equal(t, http.StatusOK, get(t, h, "/livez").Code)

	require.Equal(t, http.StatusServiceUnavailable, get(t, h, "/readyz").Code)
	ready.Store(true)
	require.Equal(t, http.StatusOK, get(t, h, "/readyz").Code)

	metrics := get(t, h, "/metrics")
	require.Equal(t, http.StatusOK, metrics.Code)
	require.Contains(t, metrics.Body.String(), "go_goroutines")

	rec := get(t, h, "/debug/version")
	require.Equal(t, http.StatusOK, rec.Code)

	var info version.Info
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &info))
	require.Equal(t, version.Get(), info)

	require.Equal(t, http.StatusOK, get(t, h, "/debug/pprof/").Code)
	require.Equal(t, http.StatusNotFound, get(t, h, "/nope").Code)
}
