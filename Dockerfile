FROM golang:1.21 as builder

WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN GOPROXY=${PROXY} go mod download
COPY . .
RUN CGO_ENABLED=0 make build

FROM alpine:latest

RUN apk update && apk add --no-cache ca-certificates tzdata
COPY --from=builder /src/bin /app
WORKDIR /app
CMD ["./app"]
