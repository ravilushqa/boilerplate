// Command boilerplate runs the service: the public HTTP and gRPC APIs and
// the internal debug server (health checks, metrics, pprof).
//
// main.go is the wiring point: it parses configuration, builds the logger
// and the dependencies (see app.go) and runs them until a signal arrives.
// No business logic lives in package main.
//
// Run with --help to list every option; each one can also be set through
// an environment variable.
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/ravilushqa/boilerplate/config"
	"github.com/ravilushqa/boilerplate/version"
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

	logger, err := newLogger(os.Stdout, cfg.Log)
	if err != nil {
		return err
	}

	logger = logger.With("version", version.Current().Version, "env", cfg.Env)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := newApp(cfg, logger)
	if err != nil {
		return err
	}

	return a.run(ctx)
}

// newLogger returns a logger that writes to w in the configured format and level.
// It is created once here and passed down explicitly; packages never use
// the global slog functions.
func newLogger(w io.Writer, cfg config.Log) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("parsing log level: %w", err)
	}

	opts := &slog.HandlerOptions{Level: level}

	switch cfg.Format {
	case "json":
		return slog.New(slog.NewJSONHandler(w, opts)), nil
	case "text":
		return slog.New(slog.NewTextHandler(w, opts)), nil
	default:
		return nil, fmt.Errorf("unknown log format %q", cfg.Format)
	}
}
