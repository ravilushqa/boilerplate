// Package logging builds the structured logger used by the whole service.
//
// The logger is created once in main and passed down explicitly;
// packages must not use the global slog functions.
package logging

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/ravilushqa/boilerplate/internal/config"
)

// New returns a logger that writes to w in the configured format and level.
func New(w io.Writer, cfg config.Log) (*slog.Logger, error) {
	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, fmt.Errorf("parse log level: %w", err)
	}

	opts := &slog.HandlerOptions{Level: level}

	var h slog.Handler

	switch cfg.Format {
	case "json":
		h = slog.NewJSONHandler(w, opts)
	case "text":
		h = slog.NewTextHandler(w, opts)
	default:
		return nil, fmt.Errorf("unknown log format %q", cfg.Format)
	}

	return slog.New(h), nil
}
