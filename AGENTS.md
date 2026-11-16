# AGENTS.md

Guide for AI coding agents and humans working on this service.
Read it before changing code. Keep it short and up to date: when a rule here
changes, change this file in the same commit.

## What this is

A Go service with a public **HTTP JSON API**, a public **gRPC API** and an
internal **debug server** (health probes, Prometheus metrics, pprof).
It ships as a distroless container and a Helm chart.

## Go guidance

General Go style comes from the **`go` skill** vendored in
[.claude/skills/go/SKILL.md](.claude/skills/go/SKILL.md) (by spf13, MIT).
Claude Code loads it automatically; other agents and humans: read it before
writing Go. The skill is the authority on Go style; this file only adds
decisions specific to this project and must not contradict it, except for
deliberate exceptions recorded in [docs/decisions.md](docs/decisions.md)
(currently: packages under `internal/`). If you find any other contradiction,
change this file to match the skill (never edit the skill).

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
| One package / one test | `go test -race ./internal/httpserver -run TestHandler` |
| **Everything CI runs** | `make check` |
| Build the binary | `make build` → `bin/boilerplate` |
| Build / run the container | `make docker`, `make up` |

**Definition of done:** `make check` passes and `git status` shows no
unexpected changes (CI fails if generated, formatted or tidied files are not
committed).

## Layout

Go packages live in `internal/`, one level deep: one package per domain,
named after what it does. The root holds only the entry point and non-Go files.

```
main.go                entry point: parse config, build logger, run. Nothing else.
app.go                 wiring (manual DI) and lifecycle: listeners, graceful shutdown
internal/config/       all options: struct tags → flags + env vars
internal/greeter/      example business logic (transport-agnostic)
internal/httpserver/   HTTP transport: routes, handlers, middlewares, JSON helpers
internal/grpcserver/   gRPC transport: service implementations, interceptors
internal/debugserver/  /livez /readyz /metrics /debug/pprof /debug/version
internal/version/      build info (ldflags + runtime/debug)
api/greeter/v1/        protobuf API: *.proto (source) and *.pb.go (generated)
tools/                 go.mod pinning dev tools (golangci-lint, buf, protoc-gen-*, govulncheck)
chart/                 Helm chart
docs/                  architecture, decisions, how-to recipes
```

`internal/` is a deliberate exception to the `go` skill, which avoids it for
applications; see [docs/decisions.md](docs/decisions.md). No nesting inside
`internal/`, no `pkg/`, no `cmd/` for the service (single binary), and no
layer packages (`service/`, `repository/`, `domain/`) or `utils/`/`common/`.

## Architecture rules

- Imports point one way: `main → {httpserver, grpcserver, debugserver} → business packages`.
  Business packages (like `greeter`) never import transports or `config`.
- **Business logic lives in its own package**, not in handlers. Handlers only
  decode input, call business logic, map errors and encode output.
- Business packages return **sentinel errors** (`var ErrX = errors.New(...)`);
  each transport maps them to status codes in one function
  (`statusFor` in httpserver, `toStatus` in grpcserver). Unknown errors become 500 /
  `codes.Internal` without leaking details.
- Transports depend on **small interfaces declared where they are used**
  (`httpserver.Greeter`), so tests can pass fakes.
- **All wiring happens in `app.go`.** No `init()`, no package-level mutable
  state, no service locators.
- **Configuration only through `config`.** Do not call `os.Getenv` elsewhere.
- **Logging:** `*slog.Logger` is injected; use `l.InfoContext(ctx, "static message", "snake_case_key", value)`.
  Never use the global `slog.Info` & co. or `log`. Never log an error and also return it.
- Wrap errors with what was being done: `fmt.Errorf("loading user %d: %w", id, err)`.

## Conventions

- Go version: see `go.mod`; only use features it allows (the skill has the table).
- Protobuf: Opaque API (`GetX()`, `X_builder{...}.Build()`); never edit `*.pb.go`.
- HTTP routes are declared in one place: `httpserver.NewHandler`. Version public
  paths (`/v1/...`).
- Tests: table-driven over a slice of structs with `t.Run(tt.name, ...)`;
  assertions with `github.com/stretchr/testify/require` (not `assert`, not
  `mock`); `cmp.Diff` from `github.com/google/go-cmp/cmp` for complex structs
  and maps; hand-written fakes; `t.Context()`; in-memory transports
  (`httptest`, `bufconn`); `synctest` instead of sleeps for timing.
  Packages are tested black-box (`package foo_test`); `package main` from `app_test.go`.
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
- Do not edit vendored skills in `.claude/skills/`; run `make skills-update`.
- Do not add a dependency when the standard library is enough. New
  dependencies need a reason in the commit message.
- Do not change public API contracts (`/v1` paths, JSON fields, `.proto`
  fields) incompatibly; run `make breaking` for protobuf.
- Do not disable linters or skip tests to make CI green.
- Never commit secrets; local overrides go to `.env` (git-ignored).

More context: [docs/architecture.md](docs/architecture.md) (how it works),
[docs/decisions.md](docs/decisions.md) (why it is built this way).
