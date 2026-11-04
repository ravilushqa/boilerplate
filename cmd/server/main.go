// Command server runs the service: the public HTTP and gRPC APIs and the
// internal debug server (health checks, metrics, pprof).
//
// Run `server --help` to list every option; each one can also be set
// through an environment variable.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ravilushqa/boilerplate/internal/app"
	"github.com/ravilushqa/boilerplate/internal/config"
	"github.com/ravilushqa/boilerplate/internal/logging"
	"github.com/ravilushqa/boilerplate/internal/version"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	cfg, err := config.Parse(args)
	if err != nil {
		return err
	}

	logger, err := logging.New(os.Stdout, cfg.Log)
	if err != nil {
		return err
	}

	logger = logger.With("version", version.Get().Version, "env", cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New(cfg, logger)
	if err != nil {
		return err
	}

	return a.Run(ctx)
}
