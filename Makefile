.PHONY: lint test test-coverage helm-install

lint:
	golangci-lint run --timeout=30m ./...

test:
	go test -race -count 1 ./...

test-coverage:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

helm-install:
	helm install boilerplate chart/ --values chart/values.yaml