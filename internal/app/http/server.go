package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gorilla/mux"
	"github.com/lmittmann/tint"

	"github.com/ravilushqa/boilerplate/internal/app/http/middlewares"
)

var errNameRequired = errors.New("name is required")

type ErrorResponse struct {
	Error string `json:"error"`
}

type Server struct {
	l      *slog.Logger
	router *mux.Router
	srv    *http.Server
}

func New(l *slog.Logger, router *mux.Router, addr string) *Server {
	s := &Server{l: l, router: router}
	s.routes()
	s.router.Use(middlewares.NewLogging(l))
	s.srv = &http.Server{
		Addr:         addr,
		Handler:      s,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	return s
}

func (s *Server) Run(ctx context.Context) error {
	stopc := make(chan struct{})
	context.AfterFunc(ctx, func() {
		defer close(stopc)
		s.l.Info("[HTTP] server stopping", slog.String("addr", s.srv.Addr))
		if err := s.srv.Shutdown(ctx); err != nil {
			s.l.Error("[HTTP] server shutdown error", tint.Err(err))
		}
	})
	s.l.Info("[HTTP] server listening", slog.String("addr", s.srv.Addr))
	if err := s.srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	<-stopc
	return nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *Server) handleRoot() http.HandlerFunc {
	type response struct {
		Message  string `json:"message"`
		Hostname string `json:"hostname"`
		MaxProcs int    `json:"max_procs"`
		NumCPU   int    `json:"num_cpu"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		hostname, _ := os.Hostname()
		s.respond(w, r, http.StatusOK, response{
			Message:  "Hello World",
			Hostname: hostname,
			MaxProcs: runtime.GOMAXPROCS(0),
			NumCPU:   runtime.NumCPU(),
		})
	}
}

func (s *Server) handleGreet() http.HandlerFunc {
	type request struct {
		Name string
	}
	type response struct {
		Greeting string `json:"greeting"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := s.decode(w, r, &req); err != nil {
			s.l.Error("failed to decode request", tint.Err(err))
			s.respond(w, r, http.StatusBadRequest, nil)
			return
		}

		if req.Name == "" {
			s.respond(w, r, http.StatusBadRequest, ErrorResponse{Error: errNameRequired.Error()})
			return
		}

		s.respond(w, r, http.StatusOK, response{Greeting: "Hello " + req.Name})
	}
}

func (s *Server) respond(w http.ResponseWriter, _ *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			s.l.Error("failed to encode response", tint.Err(err))
		}
	}
}

func (s *Server) decode(_ http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
