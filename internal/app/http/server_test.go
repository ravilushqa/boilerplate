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

func Test_server_handleGreet(t *testing.T) {
	type request struct {
		Name string `json:"name"`
	}
	type response struct {
		Greeting string `json:"greeting,omitempty"`
		Error    string `json:"error,omitempty"`
	}
	tests := []struct {
		name    string
		request request
		code    int
		want    response
	}{
		{
			name:    "success",
			request: request{Name: "Ravilushqa"},
			code:    http.StatusOK,
			want:    response{Greeting: "Hello Ravilushqa"},
		},
		{
			name:    "empty name",
			code:    http.StatusBadRequest,
			request: request{Name: ""},
			want:    response{Error: "name is required"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := New(zap.NewNop(), mux.NewRouter(), "")

			w := httptest.NewRecorder()

			var buf bytes.Buffer
			err := json.NewEncoder(&buf).Encode(tt.request)
			require.NoError(t, err)
			r := httptest.NewRequest(http.MethodPost, "/greet", &buf)

			srv.handleGreet()(w, r)

			require.Equal(t, tt.code, w.Code)
			var resp response
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}
