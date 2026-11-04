# AGENTS.md

Guide for AI coding agents and humans working on this service.
Read it before changing code. Keep it short and up to date: when a rule here
changes, change this file in the same commit.

## What this is

A Go service with a public **HTTP JSON API**, a public **gRPC API** and an
internal **debug server** (health probes, Prometheus metrics, pprof).
It ships as a distroless container and a Helm chart.

## Commands

Only Go and `make` are required; every other tool is pinned in `tools/go.mod`
and run through `go tool`.

| Task | Command |
|---|---|
| List all targets | `make` |
| Run locally | `make run` |
| Regenerate code after editing `.proto` | `make generate` |
| Format | `make fmt` |
| Lint | `make lint` |
| Test (race detector, coverage) | `make test` |
| One package / one test | `go test -race ./internal/httpapi -run TestHandler` |
| **Everything CI runs** | `make check` |
| Build the binary | `make build` → `bin/server` |
| Build / run the container | `make docker`, `make up` |

**Definition of done:** `make check` passes and `git status` shows no
unexpected changes (CI fails if generated, formatted or tidied files are not
committed).

## Layout

```
cmd/server/          main: parse config, build logger, run app. Nothing else.
internal/
  app/               wiring (manual DI) and lifecycle: listeners, graceful shutdown
  config/            all options: struct tags → flags + env vars
  greeter/           example business logic (transport-agnostic)
  httpapi/           HTTP transport: routes, handlers, middlewares, JSON helpers
  grpcapi/           gRPC transport: service implementations, interceptors
  debug/             /livez /readyz /metrics /debug/pprof /debug/version
  logging/           slog logger construction
  version/           build info (ldflags + runtime/debug)
api/greeter/v1/      protobuf API: *.proto (source) and *.pb.go (generated)
tools/               go.mod pinning dev tools (golangci-lint, buf, protoc-gen-*, govulncheck)
chart/               Helm chart
docs/                architecture, decisions, how-to recipes
```

## Architecture rules

- Dependencies point inward: `cmd → app → {httpapi, grpcapi, debug} → business packages`.
  Business packages (like `greeter`) never import transports, `config` or `app`.
- **Business logic lives in its own package**, not in handlers. Handlers only
  decode input, call business logic, map errors and encode output.
- Business packages return **sentinel errors** (`var ErrX = errors.New(...)`);
  each transport maps them to status codes in one function
  (`statusFor` in httpapi, `toStatus` in grpcapi). Unknown errors become 500 /
  `codes.Internal` without leaking details.
- Transports depend on **small interfaces declared where they are used**
  (`httpapi.Greeter`), so tests can pass fakes.
- **All wiring happens in `internal/app`.** No `init()`, no package-level
  mutable state, no service locators.
- **Configuration only through `internal/config`.** Do not call `os.Getenv` elsewhere.
- **Logging:** `*slog.Logger` is injected; use `l.InfoContext(ctx, "static message", "snake_case_key", value)`.
  Never use the global `slog.Info` & co. or `log`.
- `context.Context` is the first parameter of every function that does I/O.
- Wrap errors with context: `fmt.Errorf("load user %d: %w", id, err)`.

## Conventions

- Go version: see `go.mod`. Use modern stdlib: `net/http` routing patterns,
  `log/slog`, `errors.Join`, `t.Context()`, `for range n`, `slices`/`maps`.
- Protobuf: Opaque API (`GetX()`, `X_builder{...}.Build()`); never edit `*.pb.go`.
- HTTP routes are declared in one place: `httpapi.NewHandler`. Version public
  paths (`/v1/...`).
- Tests: package `foo_test` (black-box), table-driven with
  `map[string]struct{...}`, `t.Parallel()`, `github.com/stretchr/testify/require`
  (not `assert`), `t.Context()`, in-memory transports (`httptest`, `bufconn`).
- Every exported identifier has a doc comment; every package has a package comment
  explaining its role.
- Lint suppressions need a reason: `//nolint:gosec // why`.

## Recipes

Step-by-step guides are in [docs/recipes.md](docs/recipes.md):
add an HTTP endpoint, add a gRPC method, add a config option,
add a business package, add a dependency such as a database.

## Boundaries

- Do not edit generated files (`*.pb.go`); edit `.proto` and run `make generate`.
- Do not edit `go.sum` by hand; run `make tidy`.
- Do not add a dependency when the standard library is enough. New
  dependencies need a reason in the commit message.
- Do not change public API contracts (`/v1` paths, JSON fields, `.proto`
  fields) incompatibly; run `make breaking` for protobuf.
- Do not disable linters or skip tests to make CI green.
- Never commit secrets; local overrides go to `.env` (git-ignored).

More context: [docs/architecture.md](docs/architecture.md) (how it works),
[docs/decisions.md](docs/decisions.md) (why it is built this way).
