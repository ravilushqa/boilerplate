package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const (
	addr    = "http://localhost:8080"
	srvAddr = ":8080"
)

func Test_server_integration(t *testing.T) {
	s := New(zap.NewNop(), mux.NewRouter(), srvAddr)
	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := s.Run(ctx)
		require.NoError(t, err)
	}()

	defer func() {
		cancel()
		wg.Wait()
	}()

	// sleep for a second to let the server start
	time.Sleep(time.Second)

	t.Run("greet", func(t *testing.T) {
		t.Run("success", func(t *testing.T) {
			resp, err := http.Post(addr+"/greet", "application/json", bytes.NewBuffer([]byte(`{"name":"World"}`)))
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusOK, resp.StatusCode)
			var respBody struct {
				Greeting string `json:"greeting"`
			}
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)
			require.Equal(t, "Hello World", respBody.Greeting)
		})
		t.Run("failure", func(t *testing.T) {
			resp, err := http.Post(addr+"/greet", "application/json", bytes.NewBuffer([]byte(`{"name":""}`)))
			require.NoError(t, err)
			defer resp.Body.Close()

			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			var respBody struct {
				Error string `json:"error"`
			}
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, "name is required", respBody.Error)
		})
	})

	t.Run("not-found", func(t *testing.T) {
		resp, err := http.Get(addr + "/not-found")
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
