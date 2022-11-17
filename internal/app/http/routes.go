package http

import "net/http"

func (s *Server) routes() {
	s.router.HandleFunc("/", s.handleRoot())
	s.router.HandleFunc("/greet", s.handleGreet()).Methods(http.MethodPost)
}
