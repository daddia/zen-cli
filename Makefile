# Zen CLI Makefile
# AI-Powered Productivity Suite
#
# Design Guidelines:
# - Use ✓ for success messages (green when colored)
# - Use ✗ for failure messages (red when colored)
# - Use ! for alerts/warnings (yellow when colored)
# - Use - for neutral messages (default color)
# - Use proper indentation for hierarchical output
# - Use sentence case for consistency with Zen brand

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

# Test parameters
COVERAGE_DIR=coverage
COVERAGE_THRESHOLD=70
BUSINESS_COVERAGE_THRESHOLD=90

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
	@GOOS=$${GOOS:-$$(go env GOOS)}; \
	GOARCH=$${GOARCH:-$$(go env GOARCH)}; \
	OUTPUT=$(BINARY_DIR)/$(BINARY_NAME); \
	if [ "$$GOOS" = "windows" ]; then OUTPUT=$$OUTPUT.exe; fi; \
	$(GOBUILD) $(LDFLAGS) -tags "$(BUILD_TAGS)" -o $$OUTPUT ./cmd/zen; \
	echo "✓ Binary built: $$OUTPUT"

build-all: deps ## Build binaries for all supported platforms
	@echo "Building zen binaries for all platforms..."
	@mkdir -p $(BINARY_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output=$(BINARY_DIR)/$(BINARY_NAME)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then output=$$output.exe; fi; \
		echo "  Building for $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch $(GOBUILD) $(LDFLAGS) -tags "$(BUILD_TAGS)" -o $$output ./cmd/zen; \
		if [ $$? -eq 0 ]; then \
			echo "  ✓ Built: $$output"; \
		else \
			echo "  ✗ Failed to build for $$os/$$arch"; \
			exit 1; \
		fi; \
	done
	@echo "✓ All binaries built successfully"

## Test targets

test: test-all ## Run complete test (unit + integration + e2e)

test-all: ## Run all tests with proper distribution
	@echo "Running all tests..."
	@echo ""
	@$(MAKE) test-unit
	@$(MAKE) test-integration
	@$(MAKE) test-e2e
	@echo ""
	@echo "✓ All tests completed"
	@$(MAKE) test-coverage-report

test-unit: ## Run unit tests with coverage (70% of test suite)
	@echo "Running unit tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic \
		-timeout=30s \
		./internal/... ./pkg/...
	@echo "✓ Unit tests completed (target: <30s execution)"
	@echo "- Coverage report: $(COVERAGE_DIR)/coverage.out"

test-integration: ## Run integration tests (20% of test suite)
	@echo "🔗 Running integration tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -tags=integration -timeout=60s \
		-coverprofile=$(COVERAGE_DIR)/integration-coverage.out \
		-covermode=atomic \
		./test/integration/...
	@echo "✓ Integration tests completed (target: <1min execution)"

test-e2e: build ## Run end-to-end tests (10% of test suite)
	@echo "Running end-to-end tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -tags=e2e -timeout=120s \
		./test/e2e/...
	@echo "✓ End-to-end tests completed (target: <2min execution)"

test-unit-fast: ## Run unit tests without race detection (for development)
	@echo "Running fast unit tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -coverprofile=$(COVERAGE_DIR)/coverage-fast.out -covermode=atomic ./internal/... ./pkg/...

test-coverage: test-unit ## Generate HTML coverage report
	@echo "📈 Generating coverage report..."
	$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "✓ Coverage report generated: $(COVERAGE_DIR)/coverage.html"

test-coverage-report: ## Generate comprehensive coverage report with targets
	@echo "Coverage analysis:"
	@if [ -f $(COVERAGE_DIR)/coverage.out ]; then \
		COVERAGE=$$($(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
		echo "  Overall coverage: $$COVERAGE%"; \
		if [ $$(echo "$$COVERAGE >= $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
			echo "  ✓ Meets $(COVERAGE_THRESHOLD)% overall target"; \
		else \
			echo "  ✗ Below $(COVERAGE_THRESHOLD)% overall target"; \
		fi; \
		BUSINESS_COVERAGE=$$($(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | grep -E "(internal/|pkg/)" | awk '{sum += $$3; count++} END {if (count > 0) print sum/count; else print 0}' | sed 's/%//'); \
		if [ $$(echo "$$BUSINESS_COVERAGE >= 90" | bc -l) -eq 1 ]; then \
			echo "  ✓ Business logic meets 90% target"; \
		else \
			echo "  ✗ Business logic below 90% target"; \
		fi; \
	else \
		echo "  ✗ No coverage data found - run 'make test-unit' first"; \
	fi

test-coverage-ci: ## Generate coverage report for CI with strict validation
	@echo "CI coverage validation..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./internal/... ./pkg/...
	@COVERAGE=$$($(GOCMD) tool cover -func=$(COVERAGE_DIR)/coverage.out | tail -1 | awk '{print $$3}' | sed 's/%//'); \
	echo "Overall coverage: $$COVERAGE%"; \
	if [ $$(echo "$$COVERAGE < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "✗ Coverage $$COVERAGE% is below required $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi; \
	echo "✓ Coverage target met: $$COVERAGE% >= $(COVERAGE_THRESHOLD)%"

test-watch: ## Watch for changes and run unit tests
	@echo "Watching for changes..."
	@if command -v fswatch >/dev/null 2>&1; then \
		fswatch -o . -e ".*" -i "\\.go$$" | xargs -n1 -I{} make test-unit-fast; \
	else \
		echo "Install fswatch for file watching: brew install fswatch"; \
	fi

test-parallel: ## Run tests with maximum parallelization
	@echo "Running parallel tests..."
	@$(GOTEST) -v -race -parallel 4 -timeout=60s ./internal/... ./pkg/...

test-benchmarks: ## Run benchmark tests for performance validation
	@echo "Running benchmark tests..."
	@mkdir -p $(COVERAGE_DIR)
	$(GOTEST) -bench=. -benchmem -benchtime=1s ./internal/... ./pkg/... > $(COVERAGE_DIR)/benchmarks.txt 2>&1 || \
		(echo "✗ Benchmark tests failed. Output:" && cat $(COVERAGE_DIR)/benchmarks.txt && exit 1)
	@echo "✓ Benchmark results: $(COVERAGE_DIR)/benchmarks.txt"

test-race: ## Run tests with race detection only
	@echo "Running race condition tests..."
	$(GOTEST) -race -timeout=60s ./internal/... ./pkg/...

test-verbose: ## Run tests with verbose output
	@echo "Running verbose tests..."
	$(GOTEST) -v -race ./internal/... ./pkg/...

test-short: ## Run short tests only (skip long-running tests)
	@echo "Running short tests..."
	$(GOTEST) -short ./internal/... ./pkg/...

## Quality targets

lint: ## Run linting checks
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --timeout=5m; \
		echo "✓ Linter completed"; \
	else \
		echo "! golangci-lint not installed, running basic checks..."; \
		$(GOFMT) -d -s .; \
		$(GOVET) ./...; \
		echo "✓ Basic checks completed"; \
	fi

security: ## Run security analysis
	@echo "Running security analysis..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec -quiet ./...; \
		echo "✓ Security analysis completed"; \
	else \
		echo "! gosec not installed, skipping security analysis"; \
		echo "  Install with: go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest"; \
	fi

fmt: ## Format Go code
	@echo "Formatting code..."
	$(GOFMT) -w -s .
	@echo "✓ Code formatted"

## Dependency targets

deps: ## Download and tidy dependencies
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "✓ Dependencies updated"

deps-verify: ## Verify dependencies
	@echo "Verifying dependencies..."
	$(GOMOD) verify
	@echo "✓ Dependencies verified"

deps-upgrade: ## Upgrade all dependencies
	@echo "Upgrading dependencies..."
	$(GOGET) -u ./...
	$(GOMOD) tidy
	@echo "✓ Dependencies upgraded"

## Utility targets

clean: docs-clean ## Clean build artifacts and cache
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BINARY_DIR)/
	rm -rf $(COVERAGE_DIR)/
	rm -rf dist/
	@echo "✓ Clean completed"

install: build ## Install binary to system PATH
	@echo "Installing zen binary..."
	@if cp $(BINARY_DIR)/$(BINARY_NAME) /usr/local/bin/ 2>/dev/null; then \
		echo "✓ Installed to /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "✗ Cannot write to /usr/local/bin (try: sudo make install)"; \
		exit 1; \
	fi

uninstall: ## Remove binary from system PATH
	@echo "Uninstalling zen binary..."
	@if [ -f /usr/local/bin/$(BINARY_NAME) ]; then \
		rm /usr/local/bin/$(BINARY_NAME); \
		echo "✓ Uninstalled from /usr/local/bin/$(BINARY_NAME)"; \
	else \
		echo "- Binary not found in /usr/local/bin"; \
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
	@echo "✓ Development environment setup completed"

run: build ## Build and run zen with arguments (use ARGS="...")
	@echo "Running zen $(ARGS)..."
	./$(BINARY_DIR)/$(BINARY_NAME) $(ARGS)

debug: ## Build and run zen with debug logging
	@echo "Running zen in debug mode..."
	ZEN_DEBUG=true ./$(BINARY_DIR)/$(BINARY_NAME) $(ARGS)

## Documentation targets

docs: docs-markdown ## Generate all documentation formats

docs-markdown: ## Generate Markdown documentation
	@echo "Generating Markdown documentation..."
	@go run internal/tools/docgen/main.go \
		-out ./docs/zen \
		-format markdown \
		-frontmatter
	@echo "✓ Markdown documentation generated in docs/zen/"

docs-man: ## Generate Man page documentation
	@echo "Generating Man pages..."
	@go run internal/tools/docgen/main.go \
		-out ./man \
		-format man
	@echo "✓ Man pages generated in man/"

docs-rest: ## Generate ReStructuredText documentation
	@echo "Generating ReStructuredText documentation..."
	@go run internal/tools/docgen/main.go \
		-out ./docs/rest \
		-format rest
	@echo "✓ ReStructuredText documentation generated in docs/rest/"

docs-all: docs-markdown docs-man docs-rest ## Generate all documentation formats
	@echo "✓ All documentation formats generated"

docs-check: docs-markdown ## Regenerate docs and check for changes
	@echo "Checking if documentation is up-to-date..."
	@if git diff --exit-code docs/zen/; then \
		echo "✓ Documentation is up-to-date"; \
	else \
		echo "✗ Documentation needs to be regenerated"; \
		echo "Run 'make docs' to update documentation"; \
		exit 1; \
	fi

docs-clean: ## Remove generated documentation
	@echo "Cleaning generated documentation..."
	@rm -f docs/zen/zen_*.md docs/zen/zen.md docs/zen/index.md
	@rm -rf man/ docs/rest/
	@echo "✓ Generated documentation removed (README.md preserved)"

## Docker targets

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t zen:$(VERSION) -t zen:latest .
	@echo "✓ Docker image built: zen:$(VERSION)"

docker-run: docker-build ## Build and run Docker container
	@echo "Running Docker container..."
	docker run --rm -it zen:$(VERSION) $(ARGS)

## Release targets

release: ## Create release (requires goreleaser)
	@echo "Creating release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --clean; \
		echo "✓ Release created"; \
	else \
		echo "✗ goreleaser not installed"; \
		echo "  Install from: https://goreleaser.com/install/"; \
		exit 1; \
	fi

release-snapshot: ## Create snapshot release
	@echo "Creating snapshot release..."
	@if command -v goreleaser >/dev/null 2>&1; then \
		goreleaser release --snapshot --clean; \
		echo "✓ Snapshot release created"; \
	else \
		echo "✗ goreleaser not installed"; \
		exit 1; \
	fi

## Information targets

version: ## Display version information
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "Built:   $(BUILD_TIME)"
	@echo "Go:      $(GO_VERSION)"

binary-name: ## Display the binary name for current platform
	@GOOS=$${GOOS:-$$(go env GOOS)}; \
	OUTPUT=$(BINARY_NAME); \
	if [ "$$GOOS" = "windows" ]; then OUTPUT=$$OUTPUT.exe; fi; \
	echo "$$OUTPUT"

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

## Validation targets

ci-validate: ## Run CI validation pipeline (strict mode)
	@echo "Running CI validation pipeline..."
	@echo "===================================="
	@echo ""
	@$(MAKE) deps
	@$(MAKE) fmt
	@if ! git diff --exit-code; then \
		echo "✗ Code formatting changes detected"; \
		echo "  Run 'make fmt' and commit changes"; \
		exit 1; \
	fi
	@$(MAKE) lint
	@$(MAKE) security
	@$(MAKE) test-coverage-ci
	@$(MAKE) test-integration
	@$(MAKE) test-e2e
	@$(MAKE) test-race
	@$(MAKE) test-benchmarks
	@$(MAKE) docs-check
	@$(MAKE) build-all
	@$(GOMOD) verify
	@echo "✓ CI validation completed successfully"

ci-validate-no-docs: ## Run CI validation pipeline without documentation check
	@echo "Running CI validation pipeline (no docs)..."
	@echo "============================================="
	@echo ""
	@$(MAKE) deps
	@$(MAKE) fmt
	@if ! git diff --exit-code; then \
		echo "✗ Code formatting changes detected"; \
		echo "  Run 'make fmt' and commit changes"; \
		exit 1; \
	fi
	@$(MAKE) lint
	@$(MAKE) security
	@$(MAKE) test-coverage-ci
	@$(MAKE) test-integration
	@$(MAKE) test-e2e
	@$(MAKE) test-race
	@$(MAKE) test-benchmarks
	@$(MAKE) build-all
	@$(GOMOD) verify
	@echo "✓ CI validation completed successfully (docs check skipped)"

validate-fast: ## Run fast validation (unit tests + linting only)
	@echo "Running fast validation..."
	@echo "============================"
	@$(MAKE) deps
	@$(MAKE) fmt
	@$(MAKE) lint
	@$(MAKE) test-unit
	@$(MAKE) build
	@echo "Fast validation completed"

all: clean deps check build ## Run full build pipeline

.DEFAULT_GOAL := help
