package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gophermodz/http/httpinfra"
	"github.com/gorilla/mux"
	"github.com/jessevdk/go-flags"
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

	if err := run(ctx, l); err != nil {
		l.Fatal("run failed", zap.Error(err))
	}
	_ = l.Sync()
}

func run(ctx context.Context, l *zap.Logger) error {
	eg, ctx := errgroup.WithContext(ctx)

	// HTTP
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

func initLogger() *zap.Logger {
	lcfg := zap.NewProductionConfig()
	atom := zap.NewAtomicLevel()
	_ = atom.UnmarshalText([]byte(opts.LogLevel))
	lcfg.Level = atom

	if opts.Env == "development" || opts.Env == "test" {
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
	l = l.With(zap.String("id", id), zap.String("version", Version))
	return l
}
