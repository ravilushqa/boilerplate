package httpapi

import (
	"log/slog"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// middleware wraps a handler with additional behavior.
type middleware func(http.Handler) http.Handler

// chain applies middlewares so that the first one is the outermost.
func chain(h http.Handler, mws ...middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}

	return h
}

// statusRecorder captures the response status code.
type statusRecorder struct {
	http.ResponseWriter

	status int
}

func (rec *statusRecorder) WriteHeader(code int) {
	if rec.status == 0 {
		rec.status = code
	}

	rec.ResponseWriter.WriteHeader(code)
}

func (rec *statusRecorder) Write(b []byte) (int, error) {
	if rec.status == 0 {
		rec.status = http.StatusOK
	}

	return rec.ResponseWriter.Write(b)
}

// Unwrap lets [http.ResponseController] reach the underlying writer.
func (rec *statusRecorder) Unwrap() http.ResponseWriter {
	return rec.ResponseWriter
}

// route returns the matched ServeMux pattern, which has low cardinality
// and is therefore safe to use in logs and metric labels.
func route(r *http.Request) string {
	if r.Pattern == "" {
		return "unmatched"
	}

	return r.Pattern
}

// logRequests logs every request after it is served.
func logRequests(l *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			l.InfoContext(r.Context(), "request served",
				"method", r.Method,
				"path", r.URL.Path,
				"route", route(r),
				"status", rec.status,
				"duration", time.Since(start),
				"remote_addr", r.RemoteAddr,
			)
		})
	}
}

// instrument records request duration metrics.
func instrument(reg prometheus.Registerer) middleware {
	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "http_server_request_duration_seconds",
		Help:    "Duration of HTTP API requests.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "route", "code"})
	reg.MustRegister(duration)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w}

			next.ServeHTTP(rec, r)

			duration.WithLabelValues(r.Method, route(r), strconv.Itoa(rec.status)).Observe(time.Since(start).Seconds())
		})
	}
}

// recoverPanics turns a panic in a handler into a 500 response.
func recoverPanics(l *slog.Logger) middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				p := recover()
				if p == nil {
					return
				}

				if p == http.ErrAbortHandler { //nolint:errorlint // sentinel compared as documented by net/http
					panic(p)
				}

				l.ErrorContext(r.Context(), "panic recovered", "panic", p, "route", route(r))
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}()

			next.ServeHTTP(w, r)
		})
	}
}
