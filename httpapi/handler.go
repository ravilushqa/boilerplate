// Package httpapi implements the public HTTP JSON API.
//
// Routing uses the standard library [http.ServeMux] with method and
// wildcard patterns. Handlers are thin: decode the request, call business
// logic, map errors to status codes, encode the response.
package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/ravilushqa/boilerplate/greeter"
)

// Greeter is the business logic the API depends on.
// It is declared here, where it is used, so tests can substitute it.
type Greeter interface {
	Greet(ctx context.Context, name string) (string, error)
}

// Deps are the dependencies of the HTTP API.
type Deps struct {
	Logger  *slog.Logger
	Metrics prometheus.Registerer
	Greeter Greeter
}

// NewHandler returns the root handler of the HTTP API with middlewares applied.
func NewHandler(d Deps) http.Handler {
	h := &handlers{
		l:       d.Logger.With("component", "http"),
		greeter: d.Greeter,
	}

	mux := http.NewServeMux()

	// All routes are declared here, in one place.
	mux.HandleFunc("GET /{$}", h.root)
	mux.HandleFunc("POST /v1/greet", h.greet)

	// Middlewares run top to bottom on the way in.
	return chain(mux,
		logRequests(h.l),
		instrument(d.Metrics),
		recoverPanics(h.l),
	)
}

type handlers struct {
	l       *slog.Logger
	greeter Greeter
}

func (h *handlers) root(w http.ResponseWriter, r *http.Request) {
	h.respond(w, r, http.StatusOK, map[string]string{"status": "ok"})
}

type greetRequest struct {
	Name string `json:"name"`
}

type greetResponse struct {
	Message string `json:"message"`
}

func (h *handlers) greet(w http.ResponseWriter, r *http.Request) {
	req, err := decode[greetRequest](w, r)
	if err != nil {
		h.respondError(w, r, http.StatusBadRequest, err)
		return
	}

	msg, err := h.greeter.Greet(r.Context(), req.Name)
	if err != nil {
		h.respondError(w, r, statusFor(err), err)
		return
	}

	h.respond(w, r, http.StatusOK, greetResponse{Message: msg})
}

// statusFor maps business errors to HTTP status codes.
func statusFor(err error) int {
	switch {
	case errors.Is(err, greeter.ErrEmptyName), errors.Is(err, greeter.ErrNameTooLong):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
