# boilerplate

<!-- template:start -->
> **Template for Go services.** Production-ready layout with HTTP and gRPC APIs,
> health checks, metrics, graceful shutdown, a Helm chart and CI, written to
> be easy to read and extend for humans and AI agents alike.
>
> ```sh
> go install github.com/ravilushqa/boilerplate/cmd/boilerplate@latest
> boilerplate new github.com/acme/billing
> cd billing && git init && make generate && make check
> ```
>
> See [docs/template.md](docs/template.md) for how the generator works and how
> to maintain the template. This block is removed from generated projects.
<!-- template:end -->

A Go service exposing an HTTP JSON API and a gRPC API.

## Quick start

Requirements: Go (version in `go.mod`) and make. Docker is optional.

```sh
make run          # start locally with human-readable logs
make              # list all tasks
make check        # format, lint, test: run before every commit
```

Try it:

```sh
curl -s localhost:8080/v1/greet -d '{"name":"World"}'
# {"message":"Hello, World!"}

grpcurl -plaintext -d '{"name":"World"}' localhost:50051 greeter.v1.GreeterService/Greet

curl -s localhost:8081/readyz        # readiness probe
curl -s localhost:8081/metrics       # Prometheus metrics
curl -s localhost:8081/debug/version # build info
```

## Configuration

Every option is a flag and an environment variable. `go run ./cmd/server --help`
prints the full reference; [.env.example](.env.example) lists the defaults.

| Variable | Default | Description |
|---|---|---|
| `ENV` | `development` | `development`, `test` or `production` |
| `LOG_LEVEL` | `info` | `debug`, `info`, `warn`, `error` |
| `LOG_FORMAT` | `json` | `json` or `text` |
| `HTTP_ADDR` | `:8080` | HTTP API listen address |
| `GRPC_ADDR` | `:50051` | gRPC API listen address |
| `DEBUG_ADDR` | `:8081` | health, metrics and pprof; keep it internal |
| `SHUTDOWN_TIMEOUT` | `15s` | time to drain in-flight requests on shutdown |

## Deployment

```sh
make docker                          # build the image
helm upgrade --install boilerplate chart/ \
  --set image.repository=<registry>/boilerplate --set image.tag=<version>
```

The container is distroless and runs as non-root. Kubernetes probes use
`/livez` and `/readyz` on the debug port; enable `serviceMonitor` in
`chart/values.yaml` if you run the Prometheus Operator.

## Documentation

- [AGENTS.md](AGENTS.md): commands, layout, rules and conventions. Start here, whether you are a person or an AI agent.
- [docs/architecture.md](docs/architecture.md): how the service works.
- [docs/recipes.md](docs/recipes.md): adding endpoints, gRPC methods, options, dependencies.
- [docs/decisions.md](docs/decisions.md): why it is built this way.
