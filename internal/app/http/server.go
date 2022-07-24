package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var (
	errNameRequired = errors.New("name is required")
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type server struct {
	l      *zap.Logger
	router *mux.Router
	srv    *http.Server
}

func New(l *zap.Logger, router *mux.Router, addr string) *server {
	s := &server{l: l, router: router}
	s.routes()
	s.srv = &http.Server{
		Addr:         addr,
		Handler:      s,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	return s
}

func (s *server) Run(ctx context.Context) error {
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

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) handleGreet() http.HandlerFunc {
	type request struct {
		Name string
	}
	type response struct {
		Greeting string `json:"greeting"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var req request
		if err := s.decode(w, r, &req); err != nil {
			s.l.Error("failed to decode request", zap.Error(err))
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

func (s *server) respond(w http.ResponseWriter, _ *http.Request, status int, data interface{}) {
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			s.l.Error("failed to encode response", zap.Error(err))
		}
	}
}

func (s *server) decode(_ http.ResponseWriter, r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
