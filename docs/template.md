# Maintaining the template

This file exists only in the template repository (it is listed in
`.templateignore`).

## Creating a service

```sh
go install github.com/ravilushqa/boilerplate/cmd/boilerplate@latest

boilerplate new github.com/acme/billing          # creates ./billing
boilerplate new github.com/acme/billing --force  # overwrite without asking (CI, AI agents)
boilerplate new github.com/acme/billing -b dev   # from another branch of the template
LAYOUT_REPO=git@github.com:acme/template.git boilerplate new github.com/acme/billing

cd billing
git init
make generate   # regenerate protobuf code for the new module path
make check
```

`boilerplate new` is non-interactive unless the target directory exists and
stdin is a terminal, so agents and scripts can run it safely.

## What the generator does

1. Clones (or updates) the template into `~/.boilerplate/repo/...`.
2. Copies it into the new directory, without `.git`, replacing the template
   module path (`github.com/ravilushqa/boilerplate`) with the new one in every
   file except generated `*.pb.go` (their embedded descriptors are binary;
   `make generate` refreshes them).
3. Removes paths listed in `.templateignore` (this CLI, template-only
   workflows and docs).
4. Replaces the word `boilerplate` with the service name in non-Go files
   (Helm chart, Makefile, compose, docs).
5. Removes blocks between `<!-- template:start -->` and `<!-- template:end -->`
   in Markdown files.

`--nomod` adds the service into the current module instead (monorepo mode),
skipping module-level files.

## Rules for template changes

- Everything in the template must work right after generation: the `CLI`
  workflow generates a project from the current commit and runs `make check`
  and `make build` in it.
- Do not write the word `boilerplate` where it must not be renamed, and keep
  Go code free of it (Go code is renamed through the module path only).
- Template-only files go to `.templateignore`; template-only text in Markdown
  goes between `template:start` / `template:end` markers.
- Keep [AGENTS.md](../AGENTS.md) in sync with the code: it is the contract
  that keeps generated services consistent.
