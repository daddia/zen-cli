# Zen CLI Makefile
# AI-Powered Productivity Suite

.PHONY: help build build-all test test-unit test-integration test-e2e lint security deps clean install docker-build release dev-setup

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build parameters
BINARY_NAME=zen
BINARY_DIR=bin
COVERAGE_DIR=coverage

# Version information (can be overridden)
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME?=$(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
GO_VERSION?=$(shell go version | awk '{print $$3}')

# Build flags
LDFLAGS=-ldflags "\
	-X github.com/daddia/zen/pkg/cmd/factory.version=$(VERSION) \
	-X github.com/daddia/zen/pkg/cmd/factory.commit=$(COMMIT) \
	-X github.com/daddia/zen/pkg/cmd/factory.buildTime=$(BUILD_TIME)"

# Build tags
BUILD_TAGS?=

# Target platforms for cross-compilation
PLATFORMS=linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

help: ## Display this help screen
	@echo "Zen CLI Build System"
	@echo "===================="
	@echo ""
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Built:   $(BUILD_TIME)"

## Build targets

build: deps ## Build the zen binary for current platform
	@echo "Building zen binary..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -tags "$(BUILD_TAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) ./cmd/zen
	@echo "✅ Binary built: $(BINARY_DIR)/$(BINARY_NAME)"

build-all: deps ## Build binaries for all supported platforms
	@echo "Building zen binaries for all platforms..."
	@mkdir -p $(BINARY_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output=$(BINARY_DIR)/$(BINARY_NAME)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then output=$$output.exe; fi; \
		echo "Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -tags "$(BUILD_TAGS)" -o $$output ./cmd/zen; \
		if [ $$? -eq 0 ]; then \
			echo "✅ Built: $$output"; \
		else \
			echo "❌ Failed to build for $$os/$$arch"; \
			exit 1; \
		fi; \
	done
	@echo "✅ All binaries built successfully"

## Test targets

test: test-unit ## Run all tests (alias for test-unit)

test-unit: ## Run unit tests with coverage
	@echo "Running unit tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...
	@echo "✅ Unit tests completed"
	@echo "📊 Coverage report: $(COVERAGE_DIR)/coverage.out"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	$(GOTEST) -v -tags=integration ./test/integration/...
	@echo "✅ Integration tests completed"

test-e2e: build ## Run end-to-end tests
	@echo "Running end-to-end tests..."
	$(GOTEST) -v -tags=e2e ./test/e2e/...
	@echo "✅ End-to-end tests completed"

test-coverage: test-unit ## Generate HTML coverage report
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "✅ Coverage report generated: $(COVERAGE_DIR)/coverage.html"

## Quality targets

lint: ## Run linting checks
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
	else \
		echo "⚠️  golangci-lint not installed, running basic checks..."; \
		$(GOFMT) -d -s .; \
		$(GOVET) ./...; \
		echo "✅ Basic checks completed"; \
	fi

security: ## Run security analysis
	@echo "Running security analysis..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./...; \
		echo "✅ Security analysis completed"; \
	else \
		echo "⚠️  gosec not installed, skipping security analysis"; \
		echo "   Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) -w -s .
	@echo "Code formatted"

## Dependency targets

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✅ Dependencies updated"

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "✅ Dependencies verified"

deps-upgrade: ## Upgrade all dependencies
	@echo "Upgrading dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "✅ Dependencies upgraded"

## Utility targets

clean: ## Clean build artifacts and cache
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)/
	rm -rf $(COVERAGE_DIR)/
	rm -rf dist/
	@echo "✅ Clean completed"

install: build ## Install binary to system PATH
	@echo "Installing zen binary..."
	@if cp $(BINARY_DIR)/$(BINARY_NAME) /usr/local/bin/ 2>/dev/null; then \
		echo "✅ Installed to /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "❌ Cannot write to /usr/local/bin (try: sudo make install)"; \
		exit 1; \
	fi

uninstall: ## Remove binary from system PATH
	@echo "Uninstalling zen binary..."
	@if [ -f /usr/local/bin/$(BINARY_NAME) ]; then \
		rm /usr/local/bin/$(BINARY_NAME); \
		echo "✅ Uninstalled from /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "ℹ️  Binary not found in /usr/local/bin"; \
	fi

## Development targets

dev-setup: ## Setup development environment
	@echo "Setting up development environment..."
	@echo "Installing development tools..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	fi
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	fi
	@echo "✅ Development environment setup completed"

run: build ## Build and run zen with arguments (use ARGS="...")
	@echo "Running zen $(ARGS)..."
	./$(BINARY_DIR)/$(BINARY_NAME) $(ARGS)

debug: ## Build and run zen with debug logging
	@echo "Running zen in debug mode..."
	ZEN_DEBUG=true ./$(BINARY_DIR)/$(BINARY_NAME) $(ARGS)

## Docker targets

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t zen:$(VERSION) -t zen:latest .
	@echo "✅ Docker image built: zen:$(VERSION)"

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	docker run --rm -it zen:$(VERSION) $(ARGS)

## Release targets

release: ## Create release (requires goreleaser)
	@echo "Creating release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
		echo "✅ Release created"; \
	else \
		echo "❌ goreleaser not installed"; \
		echo "   Install from: https://goreleaser.com/install/"; \
		exit 1; \
	fi

release-snapshot: ## Create snapshot release
	@echo "Creating snapshot release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
		echo "✅ Snapshot release created"; \
	else \
		echo "❌ goreleaser not installed"; \
		exit 1; \
	fi

## Information targets

version: ## Display version information
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Built:   $(BUILD_TIME)"
	@echo "Go:      $(GO_VERSION)"

info: version ## Display build information
	@echo ""
	@echo "Build Configuration:"
	@echo "  Binary:     $(BINARY_NAME)"
	@echo "  Directory:  $(BINARY_DIR)"
	@echo "  Tags:       $(BUILD_TAGS)"
	@echo "  Platforms:  $(PLATFORMS)"
	@echo ""
	@echo "Paths:"
	@echo "  Binary:     $(BINARY_DIR)/$(BINARY_NAME)"
	@echo "  Coverage:   $(COVERAGE_DIR)/"
	@echo ""

check: lint security test ## Run all quality checks

all: clean deps check build ## Run full build pipeline

.DEFAULT_GOAL := help
