package httpprovider

import (
	"context"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server struct for http server
type Server struct {
	srv *http.Server
}

// New creates server
func New(addr string, handler *http.ServeMux) *Server {
	if handler == nil {
		handler = http.NewServeMux()
	}

	handler.HandleFunc("/health-check", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"status": "ok"}`)
	})
	handler.Handle("/metrics", promhttp.Handler())

	handler.HandleFunc("/debug/pprof/", pprof.Index)
	handler.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	handler.HandleFunc("/debug/pprof/profile", pprof.Profile)
	handler.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	handler.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return &Server{
		srv: &http.Server{
			Addr:    addr,
			Handler: handler,
		},
	}
}

// Run runs server
func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		err := s.srv.Shutdown(ctx)
		if err != nil {
			return
		}
	}()
	if err := s.srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
