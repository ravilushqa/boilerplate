package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gophermodz/http/httpinfra"
	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
	"github.com/lmittmann/tint"
	"go.uber.org/automaxprocs/maxprocs"
	"golang.org/x/sync/errgroup"

	"github.com/ravilushqa/boilerplate/internal/app/grpc"
	"github.com/ravilushqa/boilerplate/internal/app/http"
)

var (
	// Version is the version of the compiled software.
	Version string
	id, _   = os.Hostname()
)

var opts struct {
	Env         string `long:"env" env:"ENV" description:"Environment name" default:"development"`
	LogLevel    string `long:"log-level" env:"LOG_LEVEL" description:"Log level" default:"info"`
	HTTPAddress string `long:"http-address" env:"HTTP_ADDRESS" description:"HTTP address" default:":8080"`
	GRPCAddress string `long:"grpc-address" env:"GRPC_ADDRESS" description:"GRPC address" default:":50051"`
	InfraPort   int    `long:"infra-port" env:"INFRA_PORT" description:"Infra port" default:"8081"`
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	_, err := flags.Parse(&opts)
	if err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			panic(err)
		}
		return
	}

	l := initLogger()
	_, _ = maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
		l.Info(fmt.Sprintf(s, i...))
	}))

	if err = run(ctx, l); err != nil {
		l.Error("run failed", tint.Err(err))
	}
}

func run(ctx context.Context, l *slog.Logger) error {
	eg, ctx := errgroup.WithContext(ctx)

	// HTTP`
	r := mux.NewRouter()
	httpServer := http.New(l, r, opts.HTTPAddress)
	eg.Go(func() error {
		return httpServer.Run(ctx)
	})

	// GRPC
	grpcServer := grpc.New(l, opts.GRPCAddress)
	eg.Go(func() error {
		return grpcServer.Run(ctx)
	})

	// Infra
	infraServer := httpinfra.New(ctx, l, httpinfra.WithPort(opts.InfraPort), httpinfra.WithVersion(Version))
	eg.Go(func() error {
		return infraServer.Run(ctx)
	})

	return eg.Wait()
}

func initLogger() *slog.Logger {
	w := os.Stderr

	var handler slog.Handler
	handler = slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	if opts.Env == "development" || opts.Env == "test" {
		handler = tint.NewHandler(w, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		})
	}

	handler = handler.WithAttrs([]slog.Attr{slog.String("id", id), slog.String("version", Version)})

	slog.SetDefault(slog.New(handler))

	return slog.With("id", id, "version", Version)
}
