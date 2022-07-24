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

	"github.com/ravilushqa/boilerplate/internal/app/http"
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
	r := mux.NewRouter()

	appHTTP := http.New(l, r, cfg.AppHTTPAddress)
	// run application
	g, gctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return listenOsSignals(gctx)
	})
	g.Go(func() error {
		l.Info("starting system http server")
		return systemHTTP.Run(gctx)
	})
	g.Go(func() error {
		l.Info("starting app http server")
		return appHTTP.Run(gctx)
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
