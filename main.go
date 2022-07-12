package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	httpprovider "github.com/ravilushqa/boilerplate/providers/http"
	loggerprovider "github.com/ravilushqa/boilerplate/providers/logger"
)

func main() {
	// init dependencies
	cfg := newConfig()
	l, err := loggerprovider.New(cfg.Env, cfg.LogLevel)
	if err != nil {
		l.Fatal("failed to create logger", zap.Error(err))
	}
	systemHTTP := httpprovider.New(cfg.HTTPAddress, nil)

	// run application
	g, gctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return listenOsSignals(gctx)
	})
	g.Go(func() error {
		l.Info("starting http server")
		return systemHTTP.Run(gctx)
	})
	if err := g.Wait(); err != nil {
		l.Error("run failed", zap.Error(err))
	}

	// cleanup
	defer func() {
		_ = l.Sync() // https://github.com/uber-go/zap/issues/880
		l.Info("graceful shutdown finished")
	}()
	l.Info("start gracefully shutdown...")
}

func listenOsSignals(ctx context.Context) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		return nil
	case s := <-sigCh:
		return fmt.Errorf("received signal %s", s)
	}
}
