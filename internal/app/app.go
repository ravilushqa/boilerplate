// Package app wires the service together and manages its lifecycle.
//
// This is the only place where dependencies are constructed and connected
// (manual dependency injection). To add a dependency such as a database:
// create it in [New], pass it to the packages that need it, and release it
// in [App.Run] after the servers are stopped.
package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	"github.com/ravilushqa/boilerplate/internal/config"
	"github.com/ravilushqa/boilerplate/internal/debug"
	"github.com/ravilushqa/boilerplate/internal/greeter"
	"github.com/ravilushqa/boilerplate/internal/grpcapi"
	"github.com/ravilushqa/boilerplate/internal/httpapi"
)

// App is the assembled service.
type App struct {
	cfg *config.Config
	l   *slog.Logger

	ready  atomic.Bool
	health *health.Server

	httpSrv  *http.Server
	httpLis  net.Listener
	grpcSrv  *grpc.Server
	grpcLis  net.Listener
	debugSrv *http.Server
	debugLis net.Listener
}

// New constructs all dependencies and binds listeners, so that configuration
// and port errors are reported before the service starts.
func New(cfg *config.Config, l *slog.Logger) (*App, error) {
	a := &App{
		cfg:    cfg,
		l:      l,
		health: health.NewServer(),
	}

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewBuildInfoCollector(),
	)

	// Business logic.
	greeterSvc := greeter.New()

	// Transports.
	a.httpSrv = newHTTPServer(httpapi.NewHandler(httpapi.Deps{
		Logger:  l,
		Metrics: reg,
		Greeter: greeterSvc,
	}), 30*time.Second)

	a.grpcSrv = grpcapi.NewServer(grpcapi.Deps{
		Logger:  l,
		Metrics: reg,
		Health:  a.health,
		Greeter: greeterSvc,
	})

	// pprof profiles and traces stream for up to 30s by default, hence the longer write timeout.
	a.debugSrv = newHTTPServer(debug.NewHandler(reg, a.ready.Load), 2*time.Minute)

	var err error

	if a.httpLis, err = listen("http", cfg.HTTP.Addr); err != nil {
		return nil, err
	}

	if a.grpcLis, err = listen("grpc", cfg.GRPC.Addr); err != nil {
		return nil, errors.Join(err, a.httpLis.Close())
	}

	if a.debugLis, err = listen("debug", cfg.Debug.Addr); err != nil {
		return nil, errors.Join(err, a.httpLis.Close(), a.grpcLis.Close())
	}

	return a, nil
}

// HTTPAddr returns the address the HTTP API listens on.
func (a *App) HTTPAddr() string { return a.httpLis.Addr().String() }

// GRPCAddr returns the address the gRPC API listens on.
func (a *App) GRPCAddr() string { return a.grpcLis.Addr().String() }

// DebugAddr returns the address the debug server listens on.
func (a *App) DebugAddr() string { return a.debugLis.Addr().String() }

// Run serves until ctx is canceled or a server fails, then shuts down gracefully.
// It returns nil after a clean shutdown.
func (a *App) Run(ctx context.Context) error {
	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error { return serveHTTP(a.debugSrv, a.debugLis) })
	g.Go(func() error { return serveHTTP(a.httpSrv, a.httpLis) })
	g.Go(func() error {
		if err := a.grpcSrv.Serve(a.grpcLis); err != nil {
			return fmt.Errorf("grpc server: %w", err)
		}

		return nil
	})

	a.ready.Store(true)
	a.l.InfoContext(ctx, "service started",
		"http_addr", a.HTTPAddr(),
		"grpc_addr", a.GRPCAddr(),
		"debug_addr", a.DebugAddr(),
	)

	g.Go(func() error {
		<-gctx.Done()

		// Shutdown must not be interrupted by the already canceled parent context.
		ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), a.cfg.ShutdownTimeout)
		defer cancel()

		return a.shutdown(ctx)
	})

	return g.Wait()
}

// shutdown stops accepting traffic, waits for in-flight requests and
// releases resources. The debug server is stopped last so that metrics
// and probes stay available while draining.
func (a *App) shutdown(ctx context.Context) error {
	a.l.InfoContext(ctx, "shutting down", "timeout", a.cfg.ShutdownTimeout)

	a.ready.Store(false)
	a.health.Shutdown()

	var errs []error

	if err := a.httpSrv.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("http shutdown: %w", err))
	}

	stopped := make(chan struct{})
	go func() {
		a.grpcSrv.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
	case <-ctx.Done():
		a.grpcSrv.Stop()
		errs = append(errs, fmt.Errorf("grpc shutdown: %w", ctx.Err()))
	}

	if err := a.debugSrv.Shutdown(ctx); err != nil {
		errs = append(errs, fmt.Errorf("debug shutdown: %w", err))
	}

	// Release other resources (database pools, clients, ...) here.

	a.l.InfoContext(ctx, "service stopped")

	return errors.Join(errs...)
}

func newHTTPServer(h http.Handler, writeTimeout time.Duration) *http.Server {
	return &http.Server{
		Handler:           h,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       2 * time.Minute,
	}
}

func listen(name, addr string) (net.Listener, error) {
	lis, err := (&net.ListenConfig{}).Listen(context.Background(), "tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("%s: listen on %q: %w", name, addr, err)
	}

	return lis, nil
}

func serveHTTP(srv *http.Server, lis net.Listener) error {
	if err := srv.Serve(lis); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("http server on %s: %w", lis.Addr(), err)
	}

	return nil
}
