# Decisions

Short records of why the template is built this way. Add a new entry (do not
rewrite old ones) when a decision changes, so the reasoning stays traceable.

### Standard library first

`net/http` routing patterns (`"POST /v1/greet"`, `r.PathValue`), `log/slog`,
`errors.Join`, `testing` with `t.Context()`. Each dependency is a long-term
maintenance cost and something both humans and agents must learn.
Replaced: gorilla/mux, go-flags, tint, automaxprocs (Go sets `GOMAXPROCS` from
container CPU limits since 1.25).

### Business logic separate from transports

HTTP and gRPC handlers are adapters over the same business packages. This keeps
handlers trivial, lets the same feature be exposed over both protocols and
makes business rules testable without servers.

### Manual dependency injection in one place

All construction happens in `app.go`, next to `main.go`. No DI framework, no `init()`, no
globals: the dependency graph is plain Go code that is easy to read and change.

### Configuration with kong

One struct describes every option; tags give the flag, env var, default,
allowed values and help text. Invalid configuration fails at startup with a
clear message. Chosen over hand-written `flag` + `os.Getenv` for consistency and
over viper for simplicity.

### Separate debug server

Health probes, metrics and pprof listen on their own port so they are never
exposed through the public ingress and keep working while the public servers drain.

### Tools pinned with `go tool` in a separate module

`tools/go.mod` pins golangci-lint, buf, protoc plugins and govulncheck. `go tool
-modfile=tools/go.mod` builds and caches them, so the only prerequisites are Go
and make, every developer and CI use identical versions, and Renovate updates
them. A separate module keeps tool dependencies out of the service's `go.mod`.

### buf for protobuf, Opaque API

buf lints the API (`STANDARD`), detects breaking changes and drives generation
without a system `protoc`. Generated code uses the Opaque API recommended for
new code by the Go protobuf team.

### Make as the task runner

Available everywhere, known to everyone. `make` lists targets; `make check` is
the single definition of done for humans, agents and CI.

### AGENTS.md as the source of conventions

One file that both people and AI agents read (`CLAUDE.md` imports it). Rules
are explicit and checkable (`make check`), recipes show the expected shape of
common changes, so generated code stays consistent across services.

### Distroless, non-root image

`gcr.io/distroless/static-debian13:nonroot`: no shell or package manager,
minimal attack surface, CA certificates and tzdata included. The Helm chart
runs it with a read-only root filesystem and all capabilities dropped.

### Flat layout and plain `testing`, following the `go` skill

The service follows the [`go` skill by spf13](../.claude/skills/go/SKILL.md),
vendored in the repository so every agent and person uses the same guidance:
domain packages one level deep at the module root, `main.go` at the root for a
single binary, no `internal/` or `cmd/`; tests with the standard `testing`
package and `go-cmp` instead of testify. Replaced the earlier `cmd/server` +
`internal/` layout: for an application nobody imports, those directories only
added path depth.
