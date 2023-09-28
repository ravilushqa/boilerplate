package main

import (
	"net"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

const (
	httpAddress = ":8080"
	grpcAddress = ":50051"
	infraPort   = "8081"
)

func TestGracefullShutdown(t *testing.T) {
	defer goleak.VerifyNone(t)
	os.Args = []string{"--http-address", httpAddress, "--grpc-address", grpcAddress, "--infra-port", infraPort}

	done := make(chan struct{})
	go func() {
		<-done
		e := syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
		require.NoError(t, e)
	}()

	finished := make(chan struct{})
	go func() {
		main()
		close(finished)
	}()

	defer func() {
		close(done)
		<-finished
	}()

	waitPort(t, httpAddress)
	waitPort(t, grpcAddress)
}

func waitPort(t *testing.T, addr string) {
	t.Helper()
	for i := 0; i < 10; i++ {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			_ = conn.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatal("port is not open")
}
