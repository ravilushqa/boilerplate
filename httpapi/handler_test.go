package httpapi_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"github.com/ravilushqa/boilerplate/greeter"
	"github.com/ravilushqa/boilerplate/httpapi"
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

func serve(t *testing.T, h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body)))

	return rec
}

func TestHandler(t *testing.T) {
	t.Parallel()

	h, _ := newHandler(t, &greeter.Service{})

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
		wantBody   string // empty: not checked
	}{
		{"root", http.MethodGet, "/", "", http.StatusOK, `{"status":"ok"}`},
		{"greet", http.MethodPost, "/v1/greet", `{"name":"World"}`, http.StatusOK, `{"message":"Hello, World!"}`},
		{"greet empty name", http.MethodPost, "/v1/greet", `{"name":""}`, http.StatusBadRequest, `{"error":"name must not be empty"}`},
		{"greet bad JSON", http.MethodPost, "/v1/greet", `{"name":`, http.StatusBadRequest, ""},
		{"greet unknown field", http.MethodPost, "/v1/greet", `{"nam":"x"}`, http.StatusBadRequest, ""},
		{"greet trailing data", http.MethodPost, "/v1/greet", `{"name":"a"}{}`, http.StatusBadRequest, ""},
		{"wrong method", http.MethodGet, "/v1/greet", "", http.StatusMethodNotAllowed, ""},
		{"not found", http.MethodGet, "/nope", "", http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			rec := serve(t, h, tt.method, tt.path, tt.body)
			require.Equal(t, tt.wantStatus, rec.Code, rec.Body.String())

			if tt.wantBody != "" {
				require.JSONEq(t, tt.wantBody, rec.Body.String())
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

	rec := serve(t, h, http.MethodPost, "/v1/greet", `{"name":"x"}`)

	require.Equal(t, http.StatusInternalServerError, rec.Code)
	require.JSONEq(t, `{"error":"Internal Server Error"}`, rec.Body.String(), "internal details must not leak")
}

func TestHandlerPanic(t *testing.T) {
	t.Parallel()

	h, _ := newHandler(t, greeterFunc(func(context.Context, string) (string, error) {
		panic("boom")
	}))

	rec := serve(t, h, http.MethodPost, "/v1/greet", `{"name":"x"}`)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestHandlerMetrics(t *testing.T) {
	t.Parallel()

	h, reg := newHandler(t, &greeter.Service{})
	serve(t, h, http.MethodPost, "/v1/greet", `{"name":"x"}`)

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
	want := map[string]string{"method": "POST", "route": "POST /v1/greet", "code": "200"}
	if diff := cmp.Diff(want, labels); diff != "" {
		t.Errorf("labels mismatch (-want +got):\n%s", diff)
	}
}
