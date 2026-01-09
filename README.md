# Protohackers

A collection of network protocol implementations in Go.

## Project Structure

```
.
├── cmd/              # Command-line applications
│   └── hello/        # Hello world example binary
├── internal/         # Private application and library code
└── pkg/              # Public library code
```

## Building

Build a specific binary:
```bash
go build -o bin/hello ./cmd/hello
```

Build all binaries:
```bash
go build -o bin/ ./cmd/...
```

## Running

Run the hello world example:
```bash
go run ./cmd/hello
```

Or build and run:
```bash
./bin/hello
```
