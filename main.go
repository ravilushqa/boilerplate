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
	_, err := flags.Parse(&opts)
	if err != nil {
		if fe, ok := err.(*flags.Error); ok && fe.Type == flags.ErrHelp {
			// Help message requested, standard exit
			return
		}
		// For other flag errors, log and exit
		slog.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	l := initLogger()
	slog.SetDefault(l) // Ensure default logger is set early

	_, _ = maxprocs.Set(maxprocs.Logger(func(s string, i ...interface{}) {
		l.Info(fmt.Sprintf(s, i...))
	}))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err = run(ctx, l); err != nil {
		l.Error("application run failed", "error", err)
		os.Exit(1)
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

	// Default handler using JSON format
	logLevel := new(slog.LevelVar)
	switch opts.LogLevel {
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "info":
		logLevel.Set(slog.LevelInfo)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo) // Default to Info
	}

	handlerOpts := &slog.HandlerOptions{Level: logLevel}
	handler = slog.NewJSONHandler(os.Stdout, handlerOpts)

	// Check if environment is development or test
	if opts.Env == "development" || opts.Env == "test" {
		// Override handler with tint handler for development/test
		// Ensure tint also respects the LogLevel opt
		tintOptions := &tint.Options{
			Level:      logLevel,
			TimeFormat: time.Kitchen,
		}
		if opts.LogLevel == "debug" { // tint has more verbose debug if not specified
			tintOptions.Level = slog.LevelDebug
		}
		handler = tint.NewHandler(w, tintOptions)
	}

	// Create logger with common attributes
	l := slog.New(handler).With(
		slog.String("id", id),
		slog.String("version", Version),
		slog.String("env", opts.Env),
	)

	// Set the default logger
	slog.SetDefault(l)

	return l
}
