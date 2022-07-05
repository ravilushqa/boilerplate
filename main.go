package main

import (
	"context"
	"time"

	"github.com/neonxp/rutina"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/dig"
	"go.uber.org/zap"

	httpprovider "github.com/ravilushqa/boilerplate/providers/http"
	loggerprovider "github.com/ravilushqa/boilerplate/providers/logger"
)

func buildContainer() (*dig.Container, error) {
	constructors := []interface{}{
		newConfig,
		func(cfg *config) (*zap.Logger, error) {
			return loggerprovider.New(cfg.Env, cfg.LogLevel)
		},
		func(cfg *config) *httpprovider.Server {
			return httpprovider.New(cfg.HTTPAddress, nil)
		},
	}

	c := dig.New()
	for _, cf := range constructors {
		if err := c.Provide(cf); err != nil {
			return nil, err
		}
	}

	return c, c.Invoke(func(h *httpprovider.Server) {})
}

func main() {
	tl, _ := zap.NewProduction()
	container, err := buildContainer()
	if err != nil {
		tl.Fatal("cannot build depends", zap.Error(err))
	}

	r := rutina.New(rutina.WithErrChan())
	go func() {
		for err := range r.Errors() {
			tl.Error("runtime error", zap.Error(err))
		}
	}()

	err = container.Invoke(func(h *httpprovider.Server) {
		r.Go(h.Run)
		r.ListenOsSignals()
	})
	if err != nil {
		tl.Fatal("invoke failed", zap.Error(err))
	}

	if err = r.Wait(); err != nil {
		tl.Error("run failed", zap.Error(err))
	}

	err = container.Invoke(func(l *zap.Logger) error {
		defer func() { _ = l.Sync() }() // https://github.com/uber-go/zap/issues/880
		l.Info("start gracefully shutdown...")
		_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return nil
	})
	if err != nil {
		tl.Error("shutdown failed", zap.Error(err))
	}
}
