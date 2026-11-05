package debug_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/ravilushqa/boilerplate/debug"
	"github.com/ravilushqa/boilerplate/version"
)

func get(t *testing.T, h http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, http.NoBody))

	return rec
}

func TestHandler(t *testing.T) {
	t.Parallel()

	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())

	var ready atomic.Bool
	h := debug.NewHandler(reg, ready.Load)

	checkStatus := func(path string, want int) {
		t.Helper()

		if got := get(t, h, path).Code; got != want {
			t.Errorf("GET %s: status %d, want %d", path, got, want)
		}
	}

	checkStatus("/livez", http.StatusOK)
	checkStatus("/readyz", http.StatusServiceUnavailable)
	ready.Store(true)
	checkStatus("/readyz", http.StatusOK)
	checkStatus("/debug/pprof/", http.StatusOK)
	checkStatus("/nope", http.StatusNotFound)

	if body := get(t, h, "/metrics").Body.String(); !strings.Contains(body, "go_goroutines") {
		t.Errorf("GET /metrics: no go_goroutines metric in:\n%s", body)
	}

	var info version.Info
	if err := json.Unmarshal(get(t, h, "/debug/version").Body.Bytes(), &info); err != nil {
		t.Fatalf("GET /debug/version: %v", err)
	}

	if info != version.Current() {
		t.Errorf("GET /debug/version = %+v, want %+v", info, version.Current())
	}
}
