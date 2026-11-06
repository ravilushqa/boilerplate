// Package config defines every runtime option of the service.
//
// Options are declared once, as struct fields with tags, and can be set
// with a command-line flag or an environment variable (flags win).
// Run the binary with --help to see the generated reference.
//
// To add an option: add a field with `default`, `env` and `help` tags,
// then pass the value down from internal/app. Never read os.Getenv elsewhere.
package config

import (
	"time"

	"github.com/alecthomas/kong"

	"github.com/ravilushqa/boilerplate/version"
)

// Config is the root configuration.
type Config struct {
	Version kong.VersionFlag `help:"Print version information and exit."`

	Env string `default:"development" enum:"development,test,production" env:"ENV" help:"Deployment environment: ${enum}."`

	Log   Log   `embed:"" envprefix:"LOG_"   prefix:"log-"`
	HTTP  HTTP  `embed:"" envprefix:"HTTP_"  prefix:"http-"`
	GRPC  GRPC  `embed:"" envprefix:"GRPC_"  prefix:"grpc-"`
	Debug Debug `embed:"" envprefix:"DEBUG_" prefix:"debug-"`

	ShutdownTimeout time.Duration `default:"15s" env:"SHUTDOWN_TIMEOUT" help:"How long to wait for in-flight requests on shutdown."`
}

// Log configures the logger.
type Log struct {
	Level  string `default:"info" enum:"debug,info,warn,error" env:"LEVEL"  help:"Minimal log level: ${enum}."`
	Format string `default:"json" enum:"json,text"             env:"FORMAT" help:"Log format: ${enum}."`
}

// HTTP configures the public HTTP API server.
type HTTP struct {
	Addr string `default:":8080" env:"ADDR" help:"HTTP API listen address."`
}

// GRPC configures the public gRPC API server.
type GRPC struct {
	Addr string `default:":50051" env:"ADDR" help:"gRPC API listen address."`
}

// Debug configures the internal server with health checks, metrics and pprof.
// It must not be exposed publicly.
type Debug struct {
	Addr string `default:":8081" env:"ADDR" help:"Debug server listen address (health, metrics, pprof)."`
}

// Parse builds the configuration from command-line args and the environment.
//
// For --help and --version it prints to stdout and exits the process,
// unless options override that (tests do, with kong.Exit).
func Parse(args []string, options ...kong.Option) (*Config, error) {
	var cfg Config

	opts := []kong.Option{
		kong.Description("HTTP and gRPC service."),
		kong.Vars{"version": version.Current().String()},
	}

	parser, err := kong.New(&cfg, append(opts, options...)...)
	if err != nil {
		return nil, err
	}

	if _, err = parser.Parse(args); err != nil {
		return nil, err
	}

	return &cfg, nil
}
