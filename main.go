package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/ravilushqa/boilerplate/internal/app/grpc"
	"github.com/ravilushqa/boilerplate/internal/app/http"
	httpprovider "github.com/ravilushqa/boilerplate/providers/http"
	loggerprovider "github.com/ravilushqa/boilerplate/providers/logger"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string

	id, _ = os.Hostname()
)

func main() {
	// init dependencies
	cfg := newConfig()
	l := initLogger(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		s := <-sigCh
		l.Info("received signal", zap.String("signal", s.String()))
		cancel()
	}()

	systemHTTPServer := httpprovider.New(l, cfg.HTTPAddress, nil)
	r := mux.NewRouter()

	appHTTPServer := http.New(l, r, cfg.AppHTTPAddress)

	grpcServer := grpc.New(l, cfg.GRPCAddress)
	// run application
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return systemHTTPServer.Run(gctx)
	})
	g.Go(func() error {
		return appHTTPServer.Run(gctx)
	})
	g.Go(func() error {
		return grpcServer.Run(gctx)
	})
	if err := g.Wait(); err != nil {
		l.Error("run failed", zap.Error(err))
	}

	// cleanup
	defer func() {
		l.Info("graceful shutdown finished")
		_ = l.Sync() // https://github.com/uber-go/zap/issues/880
	}()
	l.Info("start gracefully shutdown...")
}

func initLogger(cfg *config) *zap.Logger {
	l, err := loggerprovider.New(cfg.Env, cfg.LogLevel)
	if err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	l = l.With(zap.String("service.id", id), zap.String("service.version", Version))
	return l
}
