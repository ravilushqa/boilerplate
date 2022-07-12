lint:
	golangci-lint run --timeout=30m ./...

test:
	go test --race ./...

helm-install:
	helm install boilerplate chart/ --values chart/values.yaml