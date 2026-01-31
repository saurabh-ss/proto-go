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
BINARIES := $(patsubst cmd/%,$(BIN_DIR)/%,$(CMDS))
DARWIN_BINARIES := $(patsubst cmd/%,$(BIN_DIR)/darwin/%-arm64,$(CMDS))
LINUX_BINARIES := $(patsubst cmd/%,$(BIN_DIR)/linux/%,$(CMDS))

all: build ## Build binaries for current OS

build: $(BINARIES) ## Build binaries for current OS

$(BIN_DIR)/%: cmd/%/main.go
	@mkdir -p $(BIN_DIR)
	@echo "Building $*..."
	$(GOBUILD) -o $@ ./cmd/$*

build-darwin: $(DARWIN_BINARIES) ## Build binaries for macOS (ARM)

$(BIN_DIR)/darwin/%-arm64: cmd/%/main.go
	@mkdir -p $(BIN_DIR)/darwin
	@echo "Building $* for darwin/arm64..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $@ ./cmd/$*

build-linux: $(LINUX_BINARIES) ## Build binaries for Linux (x86-64)

$(BIN_DIR)/linux/%: cmd/%/main.go
	@mkdir -p $(BIN_DIR)/linux
	@echo "Building $* for linux/amd64..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $@ ./cmd/$*

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
