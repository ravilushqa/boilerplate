.PHONY: lint test test-coverage helm-install protoc precommit

lint:
	golangci-lint run --timeout=30m ./...

test:
	go test -race -count 1 ./...

test-coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

helm-install:
	helm install boilerplate chart/ --values chart/values.yaml

protoc:
	protoc --go_out=. --go_opt=paths=source_relative \
        --go-grpc_out=. --go-grpc_opt=paths=source_relative \
        api/grpc.proto

precommit: lint test