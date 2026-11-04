package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// maxBodyBytes limits the size of request bodies.
const maxBodyBytes = 1 << 20 // 1 MiB

// errorResponse is the body of every non-2xx response.
type errorResponse struct {
	Error string `json:"error"`
}

// decode reads a JSON request body into T.
// Unknown fields and trailing data are rejected.
func decode[T any](w http.ResponseWriter, r *http.Request) (T, error) {
	var v T

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()

	if err := dec.Decode(&v); err != nil {
		return v, fmt.Errorf("invalid JSON body: %w", err)
	}

	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return v, errors.New("invalid JSON body: unexpected data after JSON object")
	}

	return v, nil
}

// respond writes v as a JSON response with the given status code.
func (h *handlers) respond(w http.ResponseWriter, r *http.Request, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.l.WarnContext(r.Context(), "failed to write response", "error", err)
	}
}

// respondError writes an error response. Details of 5xx errors are logged
// but not exposed to clients.
func (h *handlers) respondError(w http.ResponseWriter, r *http.Request, status int, err error) {
	msg := err.Error()

	if status >= http.StatusInternalServerError {
		h.l.ErrorContext(r.Context(), "request failed", "error", err)
		msg = http.StatusText(status)
	}

	h.respond(w, r, status, errorResponse{Error: msg})
}
