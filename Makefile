.PHONY: build run init-tools lint test test-coverage helm-install protoc precommit help
VERSION="v0.0.0"

# if git is available and we are in a git repo, use git describe to get the version
ifneq ($(shell which git),)
	ifneq ($(shell git rev-parse --is-inside-work-tree 2>/dev/null),)
		VERSION=$(shell git describe --tags --always)
	endif
endif

# build
build:
	mkdir -p bin/ && go build -ldflags "-X main.Version=$(VERSION)" -o ./bin/app .

# run application
run: build
	./bin/app

# run this once to install tools required for development.
init-tools:
	cd tools && \
	go mod tidy && \
	go mod verify && \
	go generate -x -tags "tools"

# run golangci-lint
lint: init-tools
	./bin/golangci-lint run --timeout=30m ./...

# run go test
test:
	go test -race -count 1 ./...

# run go test with coverage
test-coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

# install helm
helm-install:
	helm install boilerplate chart/ --values chart/values.yaml

# generate api proto
protoc:
	protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/grpc.proto

# precommit command. run lint, test
precommit: lint test

# show help
help:
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
