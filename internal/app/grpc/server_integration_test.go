package grpc

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ravilushqa/boilerplate/api"
)

const (
	addr = ":50051"
)

func TestServer(t *testing.T) {
	s := New(slog.Default(), addr)
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

	t.Run("greet", func(t *testing.T) {
		cc, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		require.NoError(t, err)
		defer cc.Close()

		c := api.NewGreeterClient(cc)
		resp, err := c.Greet(ctx, &api.GreetRequest{Name: "World"})
		require.NoError(t, err)

		require.Equal(t, "Hello World", resp.Message)
	})
}
