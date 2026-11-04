package httpapi_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/internal/greeter"
	"github.com/ravilushqa/boilerplate/internal/httpapi"
)

// greeterFunc adapts a function to the httpapi.Greeter interface.
type greeterFunc func(ctx context.Context, name string) (string, error)

func (f greeterFunc) Greet(ctx context.Context, name string) (string, error) { return f(ctx, name) }

func newHandler(t *testing.T, g httpapi.Greeter) (http.Handler, *prometheus.Registry) {
	t.Helper()

	reg := prometheus.NewRegistry()
	h := httpapi.NewHandler(httpapi.Deps{
		Logger:  slog.New(slog.NewTextHandler(t.Output(), nil)),
		Metrics: reg,
		Greeter: g,
	})

	return h, reg
}

func TestHandler(t *testing.T) {
	t.Parallel()

	h, _ := newHandler(t, greeter.New())

	for name, tc := range map[string]struct {
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string
	}{
		"Root":          {http.MethodGet, "/", "", http.StatusOK, `{"status":"ok"}`},
		"Greet":         {http.MethodPost, "/v1/greet", `{"name":"World"}`, http.StatusOK, `{"message":"Hello, World!"}`},
		"GreetEmpty":    {http.MethodPost, "/v1/greet", `{"name":""}`, http.StatusBadRequest, `{"error":"name must not be empty"}`},
		"GreetBadJSON":  {http.MethodPost, "/v1/greet", `{"name":`, http.StatusBadRequest, ""},
		"GreetUnknown":  {http.MethodPost, "/v1/greet", `{"nam":"x"}`, http.StatusBadRequest, ""},
		"GreetTrailing": {http.MethodPost, "/v1/greet", `{"name":"a"}{}`, http.StatusBadRequest, ""},
		"WrongMethod":   {http.MethodGet, "/v1/greet", "", http.StatusMethodNotAllowed, ""},
		"NotFound":      {http.MethodGet, "/nope", "", http.StatusNotFound, ""},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, strings.NewReader(tc.body))
			rec := httptest.NewRecorder()

			h.ServeHTTP(rec, req)

			require.Equal(t, tc.wantStatus, rec.Code)

			if tc.wantBody != "" {
				require.JSONEq(t, tc.wantBody, rec.Body.String())
				require.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			}
		})
	}
}

func TestHandlerInternalError(t *testing.T) {
	t.Parallel()

	h, _ := newHandler(t, greeterFunc(func(context.Context, string) (string, error) {
		return "", errors.New("database is down")
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/greet", strings.NewReader(`{"name":"x"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.JSONEq(t, `{"error":"Internal Server Error"}`, rec.Body.String(), "internal details must not leak")
}

func TestHandlerPanic(t *testing.T) {
	t.Parallel()

	h, _ := newHandler(t, greeterFunc(func(context.Context, string) (string, error) {
		panic("boom")
	}))

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/greet", strings.NewReader(`{"name":"x"}`))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandlerMetrics(t *testing.T) {
	t.Parallel()

	h, reg := newHandler(t, greeter.New())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/greet", strings.NewReader(`{"name":"x"}`))
	h.ServeHTTP(httptest.NewRecorder(), req)

	mfs, err := reg.Gather()
	require.NoError(t, err)
	require.Len(t, mfs, 1)
	require.Equal(t, "http_server_request_duration_seconds", mfs[0].GetName())
	require.Len(t, mfs[0].GetMetric(), 1)

	labels := map[string]string{}
	for _, lp := range mfs[0].GetMetric()[0].GetLabel() {
		labels[lp.GetName()] = lp.GetValue()
	}

	// The route label is the ServeMux pattern, not the raw path: low cardinality.
	require.Equal(t, map[string]string{"method": "POST", "route": "POST /v1/greet", "code": "200"}, labels)
}
