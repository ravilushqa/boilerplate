# Go Service Boilerplate

This repository serves as a boilerplate for creating robust Go services. It provides a structured starting point with common features needed for modern applications, including HTTP and gRPC servers, structured logging, configuration management, and more.

This project is built using **Go 1.24**.

## Project Structure

The project is organized as follows:

-   `api/`: Contains Protocol Buffer (protobuf) definitions for gRPC services.
-   `chart/`: Helm chart for deploying the application to Kubernetes.
-   `cmd/boilerplate/`: A command-line tool related to this boilerplate (details might need further exploration - potentially for generating new projects or managing this one).
-   `internal/`: Houses the internal application logic for the main service.
    -   `internal/app/grpc/`: gRPC server implementation.
    -   `internal/app/http/`: HTTP server implementation, including routes and middlewares.
-   `tools/`: Go files for installing development tools.
-   `main.go`: The main entry point for the Go service.
-   `Makefile`: Provides convenient targets for common development tasks like building, running, testing, and linting.
-   `Dockerfile`: For building a container image of the service.
-   `.golangci.yml`: Configuration for golangci-lint.

## Prerequisites

Before you begin, ensure you have the following installed:

-   **Go 1.24** or later.
-   **Make**: For using the Makefile targets.
-   **Docker**: (Optional) For containerization and potentially some build processes.
-   **Protobuf Compiler (`protoc`)**: Required for generating Go code from `.proto` files if you modify them (`make protoc`).

## Development

### Building the Service

To build the service, run:

```bash
make build
```

This command compiles the application (from `main.go`) and places the resulting binary at `./bin/app`. The build process also injects version information.

### Running the Service

To run the service after building it:

```bash
make run
```

Alternatively, you can run the binary directly:

```bash
./bin/app [flags]
```

The service starts with the following default ports:
-   **HTTP**: `:8080`
-   **gRPC**: `:50051`
-   **Infra/Metrics**: `:8081`

You can use the `--help` flag to see available command-line options:
```bash
./bin/app --help
```

### Configuration

The service can be configured using command-line flags or environment variables. Key options include:

-   `--env` / `ENV`: Environment name (e.g., `development`, `production`). Default: `development`.
-   `--log-level` / `LOG_LEVEL`: Log level (e.g., `debug`, `info`, `warn`, `error`). Default: `info`.
-   `--http-address` / `HTTP_ADDRESS`: HTTP server address. Default: `:8080`.
-   `--grpc-address` / `GRPC_ADDRESS`: gRPC server address. Default: `:50051`.
-   `--infra-port` / `INFRA_PORT`: Port for the infrastructure/metrics server. Default: `8081`.

### Running Tests

To run all tests:

```bash
make test
```
This command executes tests with race detection enabled.

To run tests with coverage analysis:

```bash
make test-coverage
```
This will generate a `coverage.txt` file. You can view the HTML report using:
```bash
go tool cover -html=coverage.txt
```

### Linting

To lint the codebase using `golangci-lint` (as configured in `.golangci.yml`):

```bash
make lint
```
This requires development tools to be initialized first (see `make init-tools`).

### Generating Protobuf Code

If you modify the `.proto` files in the `api/` directory, you need to regenerate the Go code:

```bash
make protoc
```

### Makefile Help

To see a list of all available Makefile targets and their descriptions:

```bash
make help
```

## Boilerplate CLI Tool

The `cmd/boilerplate/` directory contains a command-line tool. Based on its name, it might be used to generate new service projects from this boilerplate or manage aspects of this project.

To build this CLI tool (example):
```bash
go build -o ./bin/boilerplate-cli ./cmd/boilerplate/main.go
```
Then run it:
```bash
./bin/boilerplate-cli --help
```
(Note: The exact functionality and usage of this CLI tool should be verified by inspecting its source code and documentation if available.)

## License

This project is licensed under the terms of the LICENSE file.