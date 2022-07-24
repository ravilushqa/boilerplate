package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	addr = ":8080"
)

func Test_server(t *testing.T) {
	h := New(zap.NewNop(), mux.NewRouter(), addr)
	srv := httptest.NewServer(h)
	defer srv.Close()
	t.Run("greet", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			resp, err := http.Post(srv.URL+"/greet", "application/json", bytes.NewBuffer([]byte(`{"name":"Ravilushqa"}`)))
			require.NoError(t, err)

			require.Equal(t, http.StatusOK, resp.StatusCode)
			var respBody struct {
				Greeting string `json:"greeting"`
			}
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)
			require.Equal(t, "Hello Ravilushqa", respBody.Greeting)
		})
		t.Run("failure", func(t *testing.T) {
			resp, err := http.Post(srv.URL+"/greet", "application/json", bytes.NewBuffer([]byte(`{"name":""}`)))
			require.NoError(t, err)

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			var respBody struct {
				Error string `json:"error"`
			}
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)
			require.Equal(t, "name is required", respBody.Error)
		})
	})

	t.Run("not-found", func(t *testing.T) {
		resp, err := http.Get(srv.URL + "/not-found")
		require.NoError(t, err)
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
