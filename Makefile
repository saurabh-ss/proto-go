.PHONY: all build build-darwin build-linux build-all clean run test fmt vet lint help

# Binary output directory
BIN_DIR := bin

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Detect current OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
	GOOS_DEFAULT := darwin
else
	GOOS_DEFAULT := linux
endif

# Find all cmd directories
CMDS := $(wildcard cmd/*)
CMD_NAMES := $(notdir $(CMDS))

all: build ## Build binaries for current OS

build: ## Build binaries for current OS
	@mkdir -p $(BIN_DIR)
	@for cmd in $(CMD_NAMES); do \
		echo "Building $$cmd for $(GOOS_DEFAULT)..."; \
		$(GOBUILD) -o $(BIN_DIR)/$$cmd ./cmd/$$cmd; \
	done

build-darwin: ## Build binaries for macOS (Intel and ARM)
	@mkdir -p $(BIN_DIR)/darwin
	@for cmd in $(CMD_NAMES); do \
		echo "Building $$cmd for darwin/amd64..."; \
		GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BIN_DIR)/darwin/$$cmd-amd64 ./cmd/$$cmd; \
		echo "Building $$cmd for darwin/arm64..."; \
		GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BIN_DIR)/darwin/$$cmd-arm64 ./cmd/$$cmd; \
	done

build-linux: ## Build binaries for Linux (Intel and ARM)
	@mkdir -p $(BIN_DIR)/linux
	@for cmd in $(CMD_NAMES); do \
		echo "Building $$cmd for linux/amd64..."; \
		GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BIN_DIR)/linux/$$cmd-amd64 ./cmd/$$cmd; \
		echo "Building $$cmd for linux/arm64..."; \
		GOOS=linux GOARCH=arm64 $(GOBUILD) -o $(BIN_DIR)/linux/$$cmd-arm64 ./cmd/$$cmd; \
	done

build-all: build-darwin build-linux ## Build binaries for all platforms (macOS and Linux)

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	$(GOCLEAN)

clean-logs: ## Remove log files
	@echo "Cleaning logs..."
	@rm -rf logs

test: ## Run tests
	$(GOTEST) -v ./...

fmt: ## Format code
	$(GOFMT) ./...

vet: ## Run go vet
	$(GOVET) ./...

lint: fmt vet ## Run formatters and linters

help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'
