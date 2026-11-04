# syntax=docker/dockerfile:1

# Build stage: runs on the build machine's platform and cross-compiles,
# which is much faster than emulating the target platform.
FROM --platform=$BUILDPLATFORM golang:1.27 AS build

WORKDIR /src

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=bind,source=go.mod,target=go.mod \
    --mount=type=bind,source=go.sum,target=go.sum \
    go mod download

ARG TARGETOS TARGETARCH
ARG VERSION=dev

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=bind,target=. \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath \
      -ldflags "-s -w -X github.com/ravilushqa/boilerplate/internal/version.version=${VERSION}" \
      -o /out/server ./cmd/server

# Runtime stage: no shell, no package manager, non-root user.
# Includes CA certificates and timezone data.
FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=build /out/server /server

USER nonroot:nonroot

# 8080: HTTP API, 50051: gRPC API, 8081: debug (health, metrics, pprof)
EXPOSE 8080 50051 8081

ENTRYPOINT ["/server"]
