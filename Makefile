.PHONY: all build clean run test fmt vet lint help

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

# Find all cmd directories
CMDS := $(wildcard cmd/*)
BINARIES := $(patsubst cmd/%,$(BIN_DIR)/%,$(CMDS))

all: build ## Build all binaries

build: $(BINARIES) ## Build all binaries

$(BIN_DIR)/%: cmd/%/main.go
	@mkdir -p $(BIN_DIR)
	@echo "Building $*..."
	$(GOBUILD) -o $@ ./cmd/$*

clean: ## Remove build artifacts
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR)
	$(GOCLEAN)

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
