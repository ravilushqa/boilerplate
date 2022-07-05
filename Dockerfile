FROM golang:1.18 as builder

WORKDIR /usr/src
COPY go.mod .
COPY go.sum .
RUN GOPROXY=${PROXY} go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

FROM alpine:latest

RUN apk update && apk add --no-cache ca-certificates tzdata
WORKDIR /usr/app
COPY --from=builder /usr/src/service .
CMD ["./app"]
