.PHONY: lint test test-coverage helm-install

ALL_COVERAGE_MOD_DIRS := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)


lint:
	golangci-lint run --timeout=30m ./...

test:
	go test -race -count 1 ./...

COVERAGE_MODE    = atomic
COVERAGE_PROFILE = coverage.out
test-coverage:
	echo $(ALL_COVERAGE_MOD_DIRS)
	@set -e; \
	printf "" > coverage.txt; \
	for dir in $(ALL_COVERAGE_MOD_DIRS); do \
	  echo "go test -coverpkg=./... -covermode=$(COVERAGE_MODE) -coverprofile="$(COVERAGE_PROFILE)" $${dir}/..."; \
	  (cd "$${dir}" && \
	    go list ./... \
	    | grep -v third_party \
	    | xargs go test -coverpkg=./... -covermode=$(COVERAGE_MODE) -coverprofile="$(COVERAGE_PROFILE)" && \
	  go tool cover -html=coverage.out -o coverage.html); \
	  [ -f "$${dir}/coverage.out" ] && cat "$${dir}/coverage.out" >> coverage.txt; \
	done; \
	sed -i.bak -e '2,$$ { /^mode: /d; }' coverage.txt

helm-install:
	helm install boilerplate chart/ --values chart/values.yaml