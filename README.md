# Protohackers

A collection of network protocol implementations in Go.

## Project Structure

```
.
├── cmd/              # Command-line applications
│   ├── smoke/        # Smoke test TCP echo server (port 10001)
│   └── prime/        # Prime number checking server (port 50001)
├── internal/         # Private application and library code
│   └── logger/       # Logging utility
├── logs/             # Application log files (generated)
└── pkg/              # Public library code
```

## Building

Build a specific binary:
```bash
make build
```

Or use the Makefile targets:
```bash
make all      # Build all binaries
make clean    # Remove build artifacts
make clean-logs  # Remove log files
```

## Running

Run the smoke test server:
```bash
./bin/smoke
```

Run the prime number server:
```bash
./bin/prime
```

Or build and run directly:
```bash
go run ./cmd/smoke
go run ./cmd/prime
```

## Logging

All applications automatically log to timestamped files in the `logs/` directory. Logs are written to both the console (stdout) and log files for easy debugging. Each application creates its own log file with the format:

```
logs/<app-name>_<timestamp>.log
```

Example: `logs/smoke_2026-01-08_22-16-24.log`
