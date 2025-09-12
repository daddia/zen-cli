# Zen CLI: Comprehensive Project Structure

*AI-Powered Product Lifecycle Productivity Platform*

## 📁 **Root Directory Structure**

```
zen/
├── .github/                    # GitHub workflows and templates
├── .vscode/                    # VS Code workspace configuration
├── build/                      # Build artifacts and packaging
├── cmd/                        # CLI entry points and commands
├── configs/                    # Configuration templates and examples
├── docs/                       # Documentation
├── examples/                   # Usage examples and tutorials
├── internal/                   # Private Go packages
├── pkg/                        # Public Go packages (APIs for plugins)
├── plugins/                    # Official plugins
├── scripts/                    # Build, development, and utility scripts
├── templates/                  # Built-in templates (migrated from existing)
├── test/                       # Integration and end-to-end tests
├── tools/                      # Development tools and generators
├── web/                        # Web UI assets (if needed)
├── .gitignore
├── .golangci.yml              # Linting configuration
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── CONTRIBUTING.md
├── Dockerfile
├── LICENSE
├── Makefile
├── README.md
├── go.mod
├── go.sum
└── goreleaser.yml             # Release automation
```

## 🎯 **Detailed Structure Breakdown**

### **1. Command Layer (`cmd/`)**

```
cmd/
├── zen/                       # Main CLI binary
│   ├── main.go               # Entry point
│   └── version.go            # Version information
├── plugins/                   # Plugin binaries (if separate)
│   ├── agent-custom/
│   └── integration-slack/
└── tools/                     # Development tools
    ├── template-validator/
    ├── prompt-optimizer/
    └── config-migrator/
```

**Key Files:**
```go
// cmd/zen/main.go
package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"
    
    "github.com/org/zen/internal/cli"
    "github.com/org/zen/internal/config"
    "github.com/org/zen/internal/logging"
)

func main() {
    ctx, cancel := signal.NotifyContext(context.Background(), 
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()
    
    cfg, err := config.Load()
    if err != nil {
        os.Exit(1)
    }
    
    logger := logging.New(cfg.LogLevel)
    
    if err := cli.Execute(ctx, cfg, logger); err != nil {
        logger.Error("execution failed", "error", err)
        os.Exit(1)
    }
}
```

### **2. Internal Packages (`internal/`)**

```
internal/
├── cli/                       # CLI command implementations
│   ├── root.go               # Root command setup
│   ├── global/               # Global commands
│   │   ├── init.go
│   │   ├── config.go
│   │   ├── status.go
│   │   └── version.go
│   ├── product/              # Product management commands
│   │   ├── strategy.go       # Product strategy and roadmapping
│   │   ├── research.go       # Market research and user insights
│   │   ├── requirements.go   # Requirements and user story management
│   │   ├── analytics.go      # Product analytics and metrics
│   │   └── validation.go     # Product validation and testing
│   ├── workflow/             # 12-stage engineering workflow commands
│   │   ├── discover.go       # Stage 01: Requirements analysis
│   │   ├── prioritize.go     # Stage 02: Backlog prioritization
│   │   ├── design.go         # Stage 03: Technical design
│   │   ├── architect.go      # Stage 04: Architecture review
│   │   ├── plan.go           # Stage 05: Implementation planning
│   │   ├── build.go          # Stage 06: Code generation and development
│   │   ├── review.go         # Stage 07: Code review and quality
│   │   ├── test.go           # Stage 08: Testing and QA
│   │   ├── secure.go         # Stage 09: Security and compliance
│   │   ├── release.go        # Stage 10: Deployment and release
│   │   ├── verify.go         # Stage 11: Post-deployment verification
│   │   └── feedback.go       # Stage 12: Analytics and feedback
│   ├── integrations/         # External system commands
│   │   ├── jira.go           # Jira project management
│   │   ├── confluence.go     # Documentation publishing
│   │   ├── git.go            # Version control
│   │   ├── ci.go             # CI/CD systems
│   │   ├── analytics.go      # Analytics platforms (GA, Mixpanel)
│   │   ├── design.go         # Design tools (Figma, Sketch)
│   │   └── communication.go  # Slack, Teams, Discord
│   └── utilities/            # Utility commands
│       ├── template.go
│       ├── agent.go
│       └── workflow.go
├── agents/                    # AI agent orchestration
│   ├── client.go             # Multi-provider LLM client
│   ├── manager.go            # Agent lifecycle management
│   ├── registry.go           # Agent registry and discovery
│   ├── context.go            # Conversation context management
│   ├── cost.go               # Cost tracking and optimization
│   ├── providers/            # LLM provider implementations
│   │   ├── openai.go
│   │   ├── anthropic.go
│   │   ├── azure.go
│   │   └── local.go          # Local model support
│   └── prompts/              # Prompt management
│       ├── loader.go         # Template loading
│       ├── renderer.go       # Template rendering
│       ├── optimizer.go      # Prompt optimization
│       └── cache.go          # Prompt caching
├── config/                    # Configuration management
│   ├── config.go             # Main configuration struct
│   ├── loader.go             # Configuration loading
│   ├── validator.go          # Configuration validation
│   ├── workspace.go          # Workspace detection and setup
│   ├── integrations.go       # External system configuration
│   ├── agents.go             # AI agent configuration
│   └── migration.go          # Configuration migration
├── integrations/             # External system clients
│   ├── product/              # Product management integrations
│   │   ├── analytics/        # Analytics platforms
│   │   │   ├── google.go     # Google Analytics
│   │   │   ├── mixpanel.go   # Mixpanel
│   │   │   ├── amplitude.go  # Amplitude
│   │   │   └── segment.go    # Segment
│   │   ├── design/           # Design tool integrations
│   │   │   ├── figma.go      # Figma API
│   │   │   ├── sketch.go     # Sketch integration
│   │   │   └── invision.go   # InVision
│   │   ├── research/         # User research tools
│   │   │   ├── surveys.go    # Survey platforms
│   │   │   ├── feedback.go   # Feedback collection
│   │   │   └── usertesting.go # User testing platforms
│   │   └── crm/              # CRM integrations
│   │       ├── salesforce.go # Salesforce
│   │       ├── hubspot.go    # HubSpot
│   │       └── pipedrive.go  # Pipedrive
│   ├── engineering/          # Engineering platform integrations
│   │   ├── jira/             # Jira integration
│   │   │   ├── client.go     # API client
│   │   │   ├── types.go      # Data types
│   │   │   ├── operations.go # CRUD operations
│   │   │   ├── workflow.go   # Workflow automation
│   │   │   └── sync.go       # Synchronization
│   │   ├── confluence/       # Confluence integration
│   │   │   ├── client.go
│   │   │   ├── publisher.go  # Document publishing
│   │   │   ├── layout.go     # Layout management
│   │   │   └── converter.go  # Markdown conversion
│   │   ├── git/              # Git integration
│   │   │   ├── client.go
│   │   │   ├── hooks.go      # Git hooks
│   │   │   ├── workflow.go   # Git workflow automation
│   │   │   └── analysis.go   # Repository analysis
│   │   └── ci/               # CI/CD integrations
│   │       ├── github.go     # GitHub Actions
│   │       ├── gitlab.go     # GitLab CI
│   │       ├── jenkins.go    # Jenkins
│   │       └── generic.go    # Generic CI/CD
│   ├── communication/        # Communication platform integrations
│   │   ├── slack.go          # Slack integration
│   │   ├── teams.go          # Microsoft Teams
│   │   ├── discord.go        # Discord
│   │   └── email.go          # Email notifications
│   └── common/               # Common integration utilities
│       ├── auth.go           # Authentication
│       ├── retry.go          # Retry logic
│       ├── cache.go          # Response caching
│       └── rate_limit.go     # Rate limiting
├── quality/                   # Quality gates and enforcement
│   ├── gates.go              # Quality gate definitions
│   ├── enforcement.go        # Automated enforcement
│   ├── metrics.go            # Quality metrics collection
│   ├── rules/                # Quality rules
│   │   ├── code.go           # Code quality rules
│   │   ├── security.go       # Security rules
│   │   ├── performance.go    # Performance rules
│   │   └── documentation.go  # Documentation rules
│   └── reporters/            # Quality reporters
│       ├── console.go
│       ├── json.go
│       └── html.go
├── templates/                 # Template engine and management
│   ├── engine.go             # Template rendering engine
│   ├── registry.go           # Template registry and discovery
│   ├── validator.go          # Template validation
│   ├── functions.go          # Custom template functions
│   ├── loader.go             # Template loading
│   ├── cache.go              # Template caching
│   └── product.go            # Product-specific template functions
├── workflow/                  # Workflow state management
│   ├── state.go              # Workflow state tracking
│   ├── orchestrator.go       # Multi-stage orchestration
│   ├── hooks.go              # Pre/post stage hooks
│   ├── persistence.go        # State persistence
│   ├── visualization.go      # Workflow visualization
│   └── recovery.go           # Error recovery
├── storage/                   # Data storage layer
│   ├── sqlite.go             # SQLite implementation
│   ├── postgres.go           # PostgreSQL implementation
│   ├── memory.go             # In-memory storage
│   ├── migrations/           # Database migrations
│   └── models/               # Data models
├── logging/                   # Structured logging
│   ├── logger.go             # Main logger
│   ├── formatters.go         # Log formatters
│   ├── levels.go             # Log levels
│   └── outputs.go            # Output destinations
├── security/                  # Security utilities
│   ├── encryption.go         # Encryption/decryption
│   ├── secrets.go            # Secret management
│   ├── auth.go               # Authentication
│   └── audit.go              # Audit logging
└── utils/                     # Common utilities
    ├── files.go              # File operations
    ├── http.go               # HTTP utilities
    ├── json.go               # JSON utilities
    ├── strings.go            # String utilities
    └── validation.go         # Input validation
```

### **3. Public APIs (`pkg/`)**

```
pkg/
├── client/                    # Go client library
│   ├── client.go             # Main client
│   ├── workflow.go           # Workflow client
│   ├── agents.go             # Agent client
│   └── types.go              # Public types
├── plugin/                    # Plugin SDK
│   ├── sdk.go                # Plugin SDK interface
│   ├── agent.go              # Agent plugin interface
│   ├── integration.go        # Integration plugin interface
│   ├── template.go           # Template plugin interface
│   └── examples/             # Plugin examples
├── types/                     # Shared types and interfaces
│   ├── workflow.go           # Workflow types
│   ├── agent.go              # Agent types
│   ├── integration.go        # Integration types
│   ├── template.go           # Template types
│   ├── quality.go            # Quality types
│   └── common.go             # Common types
└── errors/                    # Error types and handling
    ├── errors.go             # Error definitions
    ├── codes.go              # Error codes
    └── handlers.go           # Error handlers
```

### **4. Plugin System (`plugins/`)**

```
plugins/
├── agents/                    # Custom agent plugins
│   ├── domain-expert/        # Domain-specific agents
│   │   ├── main.go
│   │   ├── plugin.yaml
│   │   └── README.md
│   ├── code-reviewer/        # Specialized code reviewers
│   └── security-scanner/     # Security-focused agents
├── integrations/             # Integration plugins
│   ├── slack/                # Slack integration
│   ├── teams/                # Microsoft Teams
│   ├── datadog/              # Datadog monitoring
│   └── pagerduty/            # PagerDuty alerting
├── templates/                # Template plugins
│   ├── enterprise/           # Enterprise templates
│   ├── microservices/        # Microservice templates
│   └── mobile/               # Mobile app templates
└── quality/                  # Quality gate plugins
    ├── sonarqube/            # SonarQube integration
    ├── veracode/             # Veracode security
    └── lighthouse/           # Lighthouse performance
```

### **5. Templates (`templates/`)**

```
templates/
├── workflows/                 # Workflow templates
│   ├── story-definition.yaml # Migrated story template
│   ├── adr-template.yaml     # Architecture Decision Record
│   ├── feature-design.yaml   # Feature design document
│   ├── runbook.yaml          # Operational runbook
│   └── security-policy.yaml  # Security policy
├── prompts/                  # AI prompt templates (migrated)
│   ├── 01-discover/
│   │   ├── discover.yaml
│   │   ├── overview.yaml
│   │   └── define-story.yaml
│   ├── 02-prioritize/
│   │   └── prioritize.yaml
│   ├── 03-design/
│   │   ├── design.yaml
│   │   └── technical-design.yaml
│   ├── 04-architect/
│   │   ├── architect-review.yaml
│   │   ├── architect-adr.yaml
│   │   └── architect-solution.yaml
│   ├── 05-plan/
│   │   ├── plan-scaffold.yaml
│   │   └── sprint-planning.yaml
│   ├── 06-build/
│   │   ├── code-generation.yaml
│   │   ├── refactoring.yaml
│   │   └── performance-optimization.yaml
│   ├── 07-review/
│   │   └── code-review.yaml
│   ├── 08-test/
│   │   ├── qa-test.yaml
│   │   ├── qa-solution-review.yaml
│   │   └── unit-tests.yaml
│   ├── 09-secure/
│   │   └── security-compliance.yaml
│   ├── 10-release/
│   │   └── release-management.yaml
│   ├── 11-verify/
│   │   └── post-deploy-verification.yaml
│   ├── 12-feedback/
│   │   └── roadmap-feedback.yaml
│   └── documentation/
│       ├── doc-writer.yaml
│       └── knowledge-management.yaml
├── scaffolds/                # Project scaffolding templates
│   ├── go-service/
│   ├── react-app/
│   ├── python-api/
│   └── terraform-module/
└── configs/                  # Configuration templates
    ├── zen.yaml.template
    ├── agents.yaml.template
    ├── integrations.yaml.template
    └── product.yaml.template
```

### **6. Configuration (`configs/`)**

```
configs/
├── examples/                  # Example configurations
│   ├── basic.yaml            # Basic configuration
│   ├── enterprise.yaml       # Enterprise setup
│   ├── development.yaml      # Development environment
│   └── ci-cd.yaml           # CI/CD integration
├── schemas/                  # Configuration schemas
│   ├── zen.schema.json
│   ├── agent.schema.json
│   ├── integration.schema.json
│   └── product.schema.json
└── migrations/               # Configuration migrations
    ├── v1-to-v2.yaml
    └── v2-to-v3.yaml
```

### **7. Documentation (`docs/`)**

```
docs/
├── architecture/             # Architecture documentation
│   ├── overview.md
│   ├── decisions/            # ADRs
│   │   ├── 001-monorepo.md
│   │   ├── 002-go-choice.md
│   │   └── 003-plugin-system.md
│   └── diagrams/             # Architecture diagrams
├── user/                     # User documentation
│   ├── installation.md
│   ├── quick-start.md
│   ├── configuration.md
│   ├── workflows/            # Workflow guides
│   │   ├── discover.md
│   │   ├── design.md
│   │   └── ...
│   └── integrations/         # Integration guides
│       ├── jira.md
│       ├── confluence.md
│       └── git.md
├── developer/                # Developer documentation
│   ├── contributing.md
│   ├── plugin-development.md
│   ├── api-reference.md
│   └── testing.md
├── operations/               # Operations documentation
│   ├── deployment.md
│   ├── monitoring.md
│   ├── troubleshooting.md
│   └── security.md
└── examples/                 # Example usage
    ├── basic-workflow.md
    ├── enterprise-setup.md
    └── custom-agents.md
```

### **8. Testing (`test/`)**

```
test/
├── integration/              # Integration tests
│   ├── workflow_test.go
│   ├── agents_test.go
│   ├── jira_test.go
│   └── confluence_test.go
├── e2e/                     # End-to-end tests
│   ├── full_workflow_test.go
│   ├── plugin_system_test.go
│   └── migration_test.go
├── fixtures/                # Test fixtures
│   ├── configs/
│   ├── templates/
│   └── data/
├── mocks/                   # Mock implementations
│   ├── agent_mock.go
│   ├── jira_mock.go
│   └── git_mock.go
└── utils/                   # Test utilities
    ├── helpers.go
    ├── containers.go        # Testcontainers setup
    └── fixtures.go
```

### **9. Build & Deployment (`build/`)**

```
build/
├── ci/                      # CI/CD configurations
│   ├── github-actions/
│   ├── gitlab-ci/
│   └── jenkins/
├── docker/                  # Docker configurations
│   ├── Dockerfile
│   ├── Dockerfile.dev
│   └── docker-compose.yml
├── packages/                # Package configurations
│   ├── deb/                 # Debian packages
│   ├── rpm/                 # RPM packages
│   ├── msi/                 # Windows installer
│   └── homebrew/            # Homebrew formula
└── scripts/                 # Build scripts
    ├── build.sh
    ├── test.sh
    ├── release.sh
    └── package.sh
```

### **10. Development Tools (`tools/`)**

```
tools/
├── generators/              # Code generators
│   ├── plugin-scaffold/
│   ├── integration-client/
│   └── template-validator/
├── migrators/              # Migration tools
│   ├── config-migrator/
│   ├── template-migrator/
│   └── data-migrator/
└── dev-setup/              # Development setup
    ├── install-deps.sh
    ├── setup-env.sh
    └── pre-commit-hooks/
```

## 🔧 **Key Configuration Files**

### **Makefile**
```makefile
# Zen CLI Makefile
.PHONY: help build test clean install

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build parameters
BINARY_NAME=zen
BINARY_UNIX=$(BINARY_NAME)_unix
BINARY_WINDOWS=$(BINARY_NAME).exe

# Version information
VERSION?=$(shell git describe --tags --always --dirty)
COMMIT?=$(shell git rev-parse --short HEAD)
BUILD_TIME?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')

LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildTime=$(BUILD_TIME)"

help: ## Display this help screen
	@grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	$(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/zen

build-all: ## Build for all platforms
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-linux-amd64 ./cmd/zen
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-darwin-amd64 ./cmd/zen
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o bin/$(BINARY_NAME)-windows-amd64.exe ./cmd/zen

test: ## Run tests
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-integration: ## Run integration tests
	$(GOTEST) -v -tags=integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	$(GOTEST) -v -tags=e2e ./test/e2e/...

lint: ## Run linter
	golangci-lint run

security: ## Run security scan
	gosec ./...

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

clean: ## Clean build artifacts
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out

install: build ## Install binary
	cp bin/$(BINARY_NAME) /usr/local/bin/

docker-build: ## Build Docker image
	docker build -t zen:$(VERSION) .

release: ## Create release
	goreleaser release --rm-dist
```

### **go.mod**
```go
module github.com/org/zen

go 1.25

require (
    github.com/spf13/cobra v1.7.0
    github.com/spf13/viper v1.16.0
    github.com/stretchr/testify v1.8.4
    github.com/golang-migrate/migrate/v4 v4.16.2
    github.com/mattn/go-sqlite3 v1.14.17
    github.com/lib/pq v1.10.9
    github.com/go-resty/resty/v2 v2.7.0
    github.com/hashicorp/go-plugin v1.4.10
    github.com/sirupsen/logrus v1.9.3
    github.com/google/uuid v1.3.0
    github.com/mitchellh/mapstructure v1.5.0
    github.com/pkg/errors v0.9.1
    github.com/golang/mock v1.6.0
    github.com/testcontainers/testcontainers-go v0.21.0
)

require (
    // Indirect dependencies...
)
```

### **.golangci.yml**
```yaml
# Zen CLI Linting Configuration
run:
  timeout: 5m
  modules-download-mode: readonly

linters-settings:
  gocyclo:
    min-complexity: 15
  goconst:
    min-len: 3
    min-occurrences: 3
  gocritic:
    enabled-tags:
      - diagnostic
      - experimental
      - opinionated
      - performance
      - style
  govet:
    check-shadowing: true
  misspell:
    locale: US
  nolintlint:
    allow-leading-space: true
    require-explanation: false
    require-specific: false

linters:
  enable:
    - bodyclose
    - deadcode
    - depguard
    - dogsled
    - dupl
    - errcheck
    - exportloopref
    - exhaustive
    - funlen
    - gochecknoinits
    - goconst
    - gocritic
    - gocyclo
    - gofmt
    - goimports
    - gomnd
    - goprintffuncname
    - gosec
    - gosimple
    - govet
    - ineffassign
    - interfacer
    - lll
    - misspell
    - nakedret
    - noctx
    - nolintlint
    - rowserrcheck
    - staticcheck
    - structcheck
    - stylecheck
    - typecheck
    - unconvert
    - unparam
    - unused
    - varcheck
    - whitespace

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - gomnd
        - funlen
        - gocyclo
```

### **Dockerfile**
```dockerfile
# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o zen ./cmd/zen

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates git
WORKDIR /root/

COPY --from=builder /app/zen .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/configs ./configs

CMD ["./zen"]
```

## 🚀 **Getting Started Commands**

```bash
# Initialize new Zen project
mkdir zen && cd zen
go mod init github.com/org/zen

# Create basic structure
mkdir -p cmd/zen internal/{cli,agents,config,workflow,product} pkg/{client,types} templates

# Install dependencies
go get github.com/spf13/cobra
go get github.com/spf13/viper

# Create basic main.go
cat > cmd/zen/main.go << 'EOF'
package main

import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
)

func main() {
    if err := rootCmd.Execute(); err != nil {
        fmt.Fprintln(os.Stderr, err)
        os.Exit(1)
    }
}

var rootCmd = &cobra.Command{
    Use:   "zen",
    Short: "Zen - AI-powered product lifecycle productivity platform",
    Long:  "A unified CLI for orchestrating AI-powered workflows across the entire product lifecycle",
}
EOF

# Build and test
make build
./bin/zen --help
```

This comprehensive project structure provides:

1. **Clear Separation of Concerns** - Commands, business logic, and public APIs
2. **Scalable Architecture** - Room for growth while maintaining organization  
3. **Plugin Extensibility** - Well-defined plugin interfaces and SDK
4. **Comprehensive Testing** - Unit, integration, and E2E test support
5. **Production Ready** - Build automation, security, and deployment configs
6. **Developer Friendly** - Clear documentation and development tools

The structure follows Go best practices while accommodating the specific needs of the Zen AI-powered product lifecycle productivity platform, supporting both product management and engineering workflows.
