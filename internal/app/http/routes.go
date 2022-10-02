package http

import "net/http"

func (s *Server) routes() {
	s.router.HandleFunc("/greet", s.handleGreet()).Methods(http.MethodPost)
}
