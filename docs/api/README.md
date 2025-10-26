# Zen CLI API Documentation

Comprehensive API reference for Zen CLI internal packages and public interfaces.

## Overview

Zen CLI is built with a clean architecture separating concerns into distinct layers:

- **Command Layer** (`pkg/cmd/`) - CLI commands and user interface
- **Core Services** (`internal/`) - Business logic and core functionality  
- **Integration Layer** (`pkg/clients/`, `internal/integration/`) - External system connectivity
- **Utilities** (`pkg/`) - Shared utilities and common functionality

## API Categories

### Core APIs

#### Configuration Management
- **[Config API](config.md)** - Multi-source configuration with type safety
- **[Workspace API](workspace.md)** - Workspace initialization and management
- **[Logging API](logging.md)** - Structured logging interface

#### Template System
- **[Template Engine API](template-engine.md)** - Template compilation and rendering
- **[Asset Client API](asset-client.md)** - Asset library management

#### Authentication & Security
- **[Auth API](auth.md)** - Secure credential management
- **[Security API](security.md)** - Security utilities and validation

### Integration APIs

#### External Systems
- **[Jira Client API](jira-client.md)** - Jira integration and task synchronization
- **[Git Client API](git-client.md)** - Git repository operations
- **[HTTP Client API](http-client.md)** - Generic HTTP client utilities

#### Plugin System
- **[Plugin API](plugin.md)** - Plugin architecture and runtime
- **[Integration Service API](integration-service.md)** - Provider management

### Utility APIs

#### I/O and Formatting
- **[IOStreams API](iostreams.md)** - Terminal I/O abstraction with color support
- **[Colors API](colors.md)** - Terminal color and formatting utilities

#### File System
- **[File System API](filesystem.md)** - File operations and management
- **[Cache API](cache.md)** - Caching utilities and serialization

#### Error Handling
- **[Error API](errors.md)** - Structured error handling and reporting
- **[Command Utilities API](cmdutil.md)** - CLI command utilities

## Package Structure

```
pkg/
├── assets/          # Asset library management
├── auth/            # Authentication and credential storage
├── cache/           # Caching utilities
├── cli/             # CLI configuration
├── clients/         # External system clients
├── cmd/             # Command implementations
├── cmdutil/         # Command utilities
├── errors/          # Error handling
├── fs/              # File system operations
├── integration/     # Integration orchestration
├── iostreams/       # I/O abstraction
├── plugin/          # Plugin system
├── processor/       # Data processing
├── task/            # Task management
├── template/        # Template engine
├── templates/       # Template storage
└── types/           # Common types

internal/
├── config/          # Configuration management
├── development/     # Development utilities
├── integration/     # Integration services
├── logging/         # Logging implementation
├── providers/       # External providers
├── tools/           # Internal tools
├── workspace/       # Workspace management
└── zencmd/          # CLI entry point
```

## Design Principles

### Interface-Based Design
All major components expose interfaces for testability and modularity:

```go
type ConfigManager interface {
    Load() (*Config, error)
    GetConfig[T Configurable](parser ConfigParser[T]) (T, error)
    SetConfig[T Configurable](parser ConfigParser[T], config T) error
}

type TemplateEngine interface {
    LoadTemplate(ctx context.Context, name string) (*Template, error)
    RenderTemplate(ctx context.Context, tmpl *Template, variables map[string]interface{}) (string, error)
}
```

### Error Handling
Consistent error handling with structured error types:

```go
type ZenError struct {
    Code    ErrorCode
    Message string
    Details error
    Context map[string]interface{}
}
```

### Configuration
Type-safe configuration with validation:

```go
type Configurable interface {
    Validate() error
    Defaults() Configurable
}

type ConfigParser[T Configurable] interface {
    Parse(raw map[string]interface{}) (T, error)
    Section() string
}
```

### Dependency Injection
Factory pattern for clean dependency management:

```go
type Factory struct {
    IOStreams        *iostreams.IOStreams
    Config          func() (*config.Config, error)
    Logger          logging.Logger
    WorkspaceManager func() (WorkspaceManager, error)
    TemplateEngine   func() (TemplateEngine, error)
}
```

## Usage Examples

### Basic Configuration
```go
// Load configuration
cfg, err := config.Load()
if err != nil {
    return err
}

// Get typed configuration
cliConfig, err := config.GetConfig(cfg, cli.ConfigParser{})
if err != nil {
    return err
}
```

### Template Rendering
```go
// Create template engine
engine := template.NewEngine(logger, assetClient, template.DefaultConfig())

// Load and render template
tmpl, err := engine.LoadTemplate(ctx, "task/index.md")
if err != nil {
    return err
}

output, err := engine.RenderTemplate(ctx, tmpl, variables)
if err != nil {
    return err
}
```

### Integration Client
```go
// Create Jira client
client := jira.NewClient(integrationManager, logger)

// Fetch task data
taskData, err := client.FetchTask(ctx, jira.FetchTaskOptions{
    TaskID:     "PROJ-123",
    IncludeRaw: true,
})
if err != nil {
    return err
}
```

## Testing

### Test Utilities
Zen provides comprehensive test utilities for API testing:

```go
// Test factories
factory := testutil.NewTestFactory()

// Mock implementations
mockAuth := &testutil.MockAuthManager{}
mockAssets := &testutil.MockAssetClient{}

// Test helpers
testutil.AssertNoError(t, err)
testutil.AssertEqual(t, expected, actual)
```

### Integration Tests
Integration tests validate API interactions:

```go
func TestJiraIntegration(t *testing.T) {
    // Setup test environment
    server := testutil.NewMockJiraServer()
    defer server.Close()
    
    // Test client operations
    client := jira.NewClient(integrationManager, logger)
    taskData, err := client.FetchTask(ctx, opts)
    
    assert.NoError(t, err)
    assert.Equal(t, "PROJ-123", taskData.ID)
}
```

## Versioning and Compatibility

### API Stability
- **Public APIs** (`pkg/`) follow semantic versioning
- **Internal APIs** (`internal/`) may change between versions
- **Breaking changes** are documented in CHANGELOG.md

### Deprecation Policy
- Deprecated APIs are marked with `// Deprecated:` comments
- Deprecated APIs are supported for at least one major version
- Migration guides are provided for breaking changes

## Contributing

### Adding New APIs
1. Define interfaces in appropriate `pkg/` package
2. Implement in `internal/` package if business logic
3. Add comprehensive tests with >80% coverage
4. Document public APIs with examples
5. Update this API documentation

### API Design Guidelines
- Use interfaces for public APIs
- Follow Go naming conventions
- Provide context.Context for cancellation
- Return structured errors with context
- Support configuration via interfaces
- Include comprehensive examples

## See Also

- **[Architecture Overview](../architecture/README.md)** - System architecture
- **[Contributing Guide](../contributing/README.md)** - Development process
- **[Command Reference](../zen/)** - CLI command documentation
- **[User Guide](../user-guide/README.md)** - End-user documentation
