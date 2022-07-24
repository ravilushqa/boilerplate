package http

import "net/http"

func (s *server) routes() {
	s.router.HandleFunc("/greet", s.handleGreet()).Methods(http.MethodPost)
}
