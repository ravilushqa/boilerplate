// Package debug implements the internal (non-public) HTTP server:
// health probes, Prometheus metrics, pprof and build info.
//
// Endpoints:
//
//	GET /livez          liveness probe: the process is up
//	GET /readyz         readiness probe: the service accepts traffic
//	GET /metrics        Prometheus metrics
//	GET /debug/version  build information as JSON
//	    /debug/pprof/   Go profiler
package debug

import (
	"encoding/json"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ravilushqa/boilerplate/version"
)

// NewHandler returns the debug server handler.
// ready reports whether the service is ready to accept traffic.
func NewHandler(reg *prometheus.Registry, ready func() bool) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		writeText(w, http.StatusOK, "ok")
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		if !ready() {
			writeText(w, http.StatusServiceUnavailable, "not ready")
			return
		}

		writeText(w, http.StatusOK, "ok")
	})

	mux.Handle("GET /metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	mux.HandleFunc("GET /debug/version", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(version.Current())
	})

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}

func writeText(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body + "\n"))
}
