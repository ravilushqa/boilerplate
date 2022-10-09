package http

import (
	"bytes"
	"net/http"
	"testing"

	"github.com/gorilla/mux"

	tests "github.com/ravilushqa/boilerplate/pkg/tests/httptest"
	loggerprovider "github.com/ravilushqa/boilerplate/providers/logger"
)

func Test_server(t *testing.T) {
	l, err := loggerprovider.New("test", "debug")
	if err != nil {
		return
	}
	h := New(l, mux.NewRouter(), "")
	t.Run("greet", func(t *testing.T) {
		scenarios := []tests.APIScenario{
			{
				Name:            "success",
				Method:          http.MethodPost,
				URL:             "/greet",
				Body:            bytes.NewReader([]byte(`{"name":"World"}`)),
				ExpectedStatus:  http.StatusOK,
				ExpectedContent: []string{`{"greeting":"Hello World"}`},
				Handler:         h,
			},
			{
				Name:            "empty name",
				Method:          http.MethodPost,
				URL:             "/greet",
				Body:            bytes.NewReader([]byte(`{"name":""}`)),
				ExpectedStatus:  http.StatusBadRequest,
				ExpectedContent: []string{`{"error":"name is required"}`},
				Handler:         h,
			},
			{
				Name:           "wrong method",
				Method:         http.MethodGet,
				URL:            "/greet",
				Body:           bytes.NewReader([]byte(`{"name":"World"}`)),
				ExpectedStatus: http.StatusMethodNotAllowed,
				Handler:        h,
			},
			{
				Name:            "wrong url",
				Method:          http.MethodPost,
				URL:             "/greet1",
				Body:            bytes.NewReader([]byte(`{"name":"World"}`)),
				ExpectedStatus:  http.StatusNotFound,
				ExpectedContent: []string{`404 page not found`},
				Handler:         h,
			},
		}
		for _, scenario := range scenarios {
			scenario.Test(t)
		}
	})
}
