# Every development task is a make target. Run `make` to list them.
# Tools are pinned in tools/go.mod and run with `go tool`: only Go and make are required.

MODULE  := $(shell go list -m)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X $(MODULE)/internal/version.version=$(VERSION)
TOOL    := go tool -modfile=tools/go.mod
IMAGE   ?= boilerplate:local

.DEFAULT_GOAL := help

##@ Development

.PHONY: run
run: ## Run the service locally with human-readable logs
	go run ./cmd/server --log-format=text --log-level=debug

.PHONY: generate
generate: ## Generate code (protobuf, go:generate)
	$(TOOL) buf generate
	go generate ./...

.PHONY: fmt
fmt: ## Format code
	$(TOOL) golangci-lint fmt ./...
	$(TOOL) buf format -w

.PHONY: tidy
tidy: ## Tidy go.mod files
	go mod tidy
	cd tools && go mod tidy

##@ Verification

.PHONY: lint
lint: ## Run linters
	$(TOOL) golangci-lint run ./...
	$(TOOL) buf lint

.PHONY: test
test: ## Run tests with the race detector, write coverage.out
	go test -race -shuffle=on -count=1 -covermode=atomic -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

.PHONY: cover
cover: test ## Open the coverage report in a browser
	go tool cover -html=coverage.out

.PHONY: breaking
breaking: ## Check protobuf API for breaking changes against main
	$(TOOL) buf breaking --against '.git#branch=main'

.PHONY: vuln
vuln: ## Check dependencies for known vulnerabilities
	$(TOOL) govulncheck ./...

.PHONY: check
check: generate fmt tidy lint test ## Run everything CI runs; must pass before every commit

##@ Build

.PHONY: build
build: ## Build the binary into bin/server
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/server ./cmd/server

.PHONY: docker
docker: ## Build the container image
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .

.PHONY: up
up: ## Run the service with Docker Compose
	docker compose up --build

.PHONY: helm-lint
helm-lint: ## Lint and render the Helm chart (requires helm)
	helm lint chart
	helm template boilerplate chart > /dev/null

##@ Help

.PHONY: help
help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} \
		/^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 } \
		/^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) }' $(MAKEFILE_LIST)
