---
status: Accepted
date: 2025-09-12
decision-makers: Development Team, Architecture Team
consulted: DevOps Team, Security Team
informed: Product Team, Contributors
---

# ADR-0003 - Project Structure and Organization

## Context and Problem Statement

The Zen CLI project requires a well-organized directory structure that supports maintainability, testability, security, and future extensibility. The structure must follow Go best practices while accommodating the specific needs of a complex CLI application with multiple components, plugins, and integrations.

Key requirements:
- Clear separation between public APIs and internal implementation
- Support for plugin architecture and extensibility
- Organized test structure with different test types
- Security through private package isolation
- Developer-friendly navigation and understanding
- CI/CD pipeline integration
- Documentation and configuration organization

## Decision Drivers

* **Go Best Practices**: Follow established Go project layout conventions
* **Security**: Isolate internal implementation from public APIs
* **Maintainability**: Clear component boundaries and responsibilities
* **Testability**: Organized test structure supporting different test types
* **Extensibility**: Plugin system and future component additions
* **Developer Experience**: Intuitive navigation and clear module boundaries
* **CI/CD Integration**: Support for automated build and deployment processes
* **Documentation**: Comprehensive documentation organization

## Considered Options

* **Standard Go Project Layout** - Following golang-standards/project-layout
* **Flat Structure** - All packages in root directory
* **Domain-Driven Structure** - Organized by business domains
* **Layer-Based Structure** - Organized by technical layers
* **Microservice Structure** - Separate repositories per component

## Decision Outcome

Chosen option: **Standard Go Project Layout** with domain-specific adaptations, because it provides the best balance of Go conventions, security, and maintainability for a complex CLI application.

### Consequences

**Good:**
- Follows established Go community standards and best practices
- Clear separation between public (`pkg/`) and private (`internal/`) APIs
- Organized structure supports team collaboration and onboarding
- Plugin system naturally fits into the layout
- CI/CD pipelines can easily target specific components
- Security through package visibility controls
- Scalable structure supporting future growth

**Bad:**
- More complex than flat structure for simple components
- Requires discipline to maintain boundaries
- Some redundancy in directory nesting

**Neutral:**
- Opinionated structure may not fit all use cases
- Requires understanding of Go package visibility rules

### Confirmation

The decision has been validated through:
- Successful implementation of ZEN-001 with clean component boundaries
- Clear separation between CLI commands and business logic
- Plugin architecture foundation established
- Developer feedback on code navigation and understanding
- CI/CD pipeline successfully targeting specific components
- Security review confirming proper API isolation

## Project Structure

```
zen/
├── cmd/                        # Main Applications
│   └── zen/                   # CLI Binary
│       ├── main.go           # Application entry point
│       └── version.go        # Build-time version info
├── internal/                   # Private Application Code
│   ├── cli/                   # Command Layer
│   │   ├── root.go           # Root command and global flags
│   │   ├── version.go        # Version command
│   │   ├── init.go           # Workspace initialization
│   │   ├── config.go         # Configuration management
│   │   └── status.go         # Workspace status
│   ├── config/               # Configuration Management
│   │   └── config.go         # Loading, validation, defaults
│   ├── logging/              # Logging Infrastructure
│   │   └── logger.go         # Structured logging interface
│   ├── agents/               # AI Agent Orchestration (future)
│   ├── workflow/             # Workflow State Management (future)
│   ├── integrations/         # External System Clients (future)
│   ├── templates/            # Template Engine (future)
│   ├── quality/              # Quality Gates (future)
│   └── storage/              # Data Persistence (future)
├── pkg/                        # Public Library Code
│   ├── types/                # Common Type Definitions
│   │   └── common.go         # Shared types and constants
│   ├── errors/               # Error Handling
│   │   └── errors.go         # Error types and utilities
│   └── client/               # Go Client Library (future)
├── plugins/                    # Plugin System (future)
│   ├── agents/               # Custom AI Agents
│   ├── integrations/         # External Integrations
│   └── templates/            # Template Extensions
├── configs/                    # Configuration Files
│   └── zen.example.yaml      # Example configuration
├── docs/                       # Documentation
│   ├── architecture/         # Architecture Documentation
│   ├── cli-structure.md      # Project structure guide
│   ├── overview.md           # Product overview
│   └── roadmap.md           # Development roadmap
├── scripts/                    # Build and Development Scripts (future)
├── test/                       # Additional Test Data (future)
│   ├── fixtures/             # Test fixtures
│   ├── integration/          # Integration tests
│   └── e2e/                  # End-to-end tests
├── .github/                    # GitHub-specific Files
│   └── workflows/            # CI/CD Workflows
├── build/                      # Build Artifacts (future)
├── tools/                      # Development Tools (future)
├── Makefile                    # Build Automation
├── Dockerfile                  # Container Definition
├── go.mod                      # Go Module Definition
├── go.sum                      # Go Module Checksums
└── README.md                   # Project Documentation
```

## Component Organization Principles

### 🎯 **`cmd/` Directory**
- Contains main applications and their entry points
- Each subdirectory represents a separate binary
- Minimal business logic - delegates to `internal/` packages
- Build-time information and version handling

### 🔒 **`internal/` Directory**
- Private application code not importable by external packages
- Core business logic and implementation details
- Organized by functional domains (cli, config, logging, etc.)
- Security boundary preventing external access

### 📦 **`pkg/` Directory**
- Public library code that can be imported by external applications
- Stable APIs with backward compatibility considerations
- Common types, utilities, and client libraries
- Well-documented public interfaces

### 🔌 **`plugins/` Directory**
- Plugin system for extensibility
- Organized by plugin types (agents, integrations, templates)
- Isolated from core application code
- Dynamic loading and registration support

### 📚 **`docs/` Directory**
- Comprehensive project documentation
- Architecture decisions and technical specifications
- User guides and API documentation
- Organized by audience (users, developers, operators)

## Pros and Cons of the Options

### Standard Go Project Layout

**Good:**
- Follows established Go community conventions
- Clear separation of public and private APIs
- Security through package visibility
- Scalable structure for complex applications
- CI/CD pipeline integration
- Plugin system support
- Developer familiarity

**Bad:**
- More complex than simple flat structure
- Requires understanding of Go package rules
- Some directory nesting overhead

**Neutral:**
- Opinionated about organization
- May be overkill for very simple projects

### Flat Structure

**Good:**
- Simple and easy to understand
- Minimal directory nesting
- Fast navigation for small projects

**Bad:**
- No API visibility controls
- Poor scalability for complex projects
- Mixing of concerns
- No plugin system support
- Security concerns with exposed internals

**Neutral:**
- Suitable only for simple applications
- Requires external documentation for organization

### Domain-Driven Structure

**Good:**
- Organized by business domains
- Clear feature boundaries
- Good for large teams

**Bad:**
- May not align with Go package conventions
- Cross-cutting concerns difficult to organize
- Potential for circular dependencies

**Neutral:**
- Requires deep domain understanding
- May change as business evolves

### Layer-Based Structure

**Good:**
- Clear technical separation
- Easy to understand for developers
- Good for traditional architectures

**Bad:**
- Can lead to anemic domain models
- Cross-layer dependencies
- Not idiomatic for Go
- Difficult to maintain boundaries

**Neutral:**
- Familiar to developers from other languages
- May not leverage Go's strengths

## More Information

**Implementation Guidelines:**
- Each `internal/` package should have a single, well-defined responsibility
- Public APIs in `pkg/` must maintain backward compatibility
- Tests should be co-located with the code they test
- Plugin interfaces should be defined in `pkg/` for external use
- Configuration and documentation should be easily discoverable

**Security Considerations:**
- Internal packages are not accessible to external importers
- Sensitive logic is protected by package boundaries
- Plugin system provides controlled extension points
- Configuration files exclude sensitive defaults

**Related ADRs:**
- ADR-0001: Go Language Choice
- ADR-0002: Cobra CLI Framework Selection
- ADR-0004: Configuration Management Strategy

**References:**
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Package Organization](https://go.dev/doc/code)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go Project Structure Best Practices](https://peter.bourgon.org/go-best-practices-2016/)
