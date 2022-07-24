.PHONY: lint test helm-install
lint:
	golangci-lint run --timeout=30m ./...

test:
	go test -race -count 1 -covermode=atomic ./...

helm-install:
	helm install boilerplate chart/ --values chart/values.yaml