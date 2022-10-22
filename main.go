package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gophermodz/http/httpinfra"
	"github.com/gorilla/mux"
	_ "go.uber.org/automaxprocs"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/ravilushqa/boilerplate/internal/app/grpc"
	"github.com/ravilushqa/boilerplate/internal/app/http"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Version is the version of the compiled software.
	Version string
	id, _   = os.Hostname()
)

func main() {
	// init dependencies
	cfg := newConfig()
	l := initLogger(cfg)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	infraHTTPServer, err := httpinfra.New(ctx, l, httpinfra.WithPort(cfg.InfraPort), httpinfra.WithVersion(Version))
	if err != nil {
		return
	}
	r := mux.NewRouter()

	appHTTPServer := http.New(l, r, cfg.AppHTTPAddress)

	grpcServer := grpc.New(l, cfg.GRPCAddress)
	// run application
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return infraHTTPServer.Run(gctx)
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
	lcfg := zap.NewProductionConfig()
	atom := zap.NewAtomicLevel()
	_ = atom.UnmarshalText([]byte(cfg.LogLevel))
	lcfg.Level = atom

	if cfg.Env == "development" || cfg.Env == "test" {
		lcfg = zap.NewDevelopmentConfig()
		lcfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	l, err := lcfg.Build(zap.Hooks())
	if err != nil {
		panic(fmt.Sprintf("failed to create logger: %v", err))
	}
	if err != nil {
		panic(fmt.Errorf("failed to init logger: %w", err))
	}
	l = l.With(zap.String("service.id", id), zap.String("service.version", Version))
	return l
}
