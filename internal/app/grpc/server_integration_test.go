package grpc

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

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

	// Wait for server to be ready by attempting to connect a few times
	var cc *grpc.ClientConn
	var err error
	for i := 0; i < 10; i++ { // Increased retries slightly
		cc, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	require.NoError(t, err, "failed to connect to gRPC server after multiple retries")
	defer cc.Close()

	t.Run("greet", func(t *testing.T) {
		c := api.NewGreeterClient(cc)
		resp, err := c.Greet(ctx, &api.GreetRequest{Name: "World"})
		require.NoError(t, err)

		require.Equal(t, "Hello World", resp.Message)
	})
}
