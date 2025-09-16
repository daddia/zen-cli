# Key Design Patterns

This guide explains Zen's architecture and key design patterns used throughout the codebase.

## System Architecture

### High-Level Design

```
┌─────────────────────────────────────────┐
│            CLI Interface                │
│           (Zen Commands)                │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│           Command Layer                 │
│      (Business Logic & Validation)      │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│           Core Services                 │
│  (Config, Workspace, Logging, Errors)   │
└─────────────────────────────────────────┘
                    │
┌─────────────────────────────────────────┐
│          Integration Layer              │
│   (LLM, Git, External APIs, Plugins)    │
└─────────────────────────────────────────┘
```

### Package Structure

```
zen/
├── cmd/zen/          # Zen CLI binary entry point
├── pkg/              # Public packages
│   ├── cmd/          # Command implementations
│   ├── cmdutil/      # Command utilities
│   ├── errors/       # Error handling
│   ├── iostreams/    # I/O abstractions
│   └── types/        # Common types
└── internal/         # Private packages
    ├── config/       # Configuration management
    ├── logging/      # Structured logging
    ├── workspace/    # Project management
    └── zencmd/       # Command execution
```

## Design Patterns

### Factory Pattern

We use factories to create configured instances:

```go
// Factory provides dependencies for commands
type Factory interface {
    IOStreams() *iostreams.IOStreams
    Config() *config.Config
    Logger() *logging.Logger
}

// Commands receive factory
func NewCmdInit(f cmdutil.Factory) *cobra.Command {
    opts := &InitOptions{
        Factory: f,
        IO:      f.IOStreams(),
    }
    // ...
}
```

### Command Pattern

Each CLI command follows this structure:

```go
type CommandOptions struct {
    Factory cmdutil.Factory
    IO      *iostreams.IOStreams
    
    // Command-specific fields
    Path    string
    Config  string
}

func NewCommand(f cmdutil.Factory) *cobra.Command {
    opts := &CommandOptions{
        Factory: f,
        IO:      f.IOStreams(),
    }
    
    cmd := &cobra.Command{
        Use:   "command",
        Short: "Brief description",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runCommand(opts)
        },
    }
    
    // Add flags
    cmd.Flags().StringVar(&opts.Path, "path", "", "Path description")
    
    return cmd
}

func runCommand(opts *CommandOptions) error {
    // Implementation
}
```

### Dependency Injection

Dependencies are injected rather than created:

```go
// Bad - Hard dependency
func Process() error {
    logger := logging.New()  // Creates dependency
    config := config.Load()  // Creates dependency
}

// Good - Injected dependencies
func Process(logger *logging.Logger, cfg *config.Config) error {
    // Use provided dependencies
}
```

### Interface Segregation

Define minimal interfaces for dependencies:

```go
// Specific interface for what we need
type ConfigReader interface {
    Get(key string) string
    GetBool(key string) bool
}

// Not the entire config implementation
func ProcessWithConfig(cfg ConfigReader) error {
    if cfg.GetBool("feature.enabled") {
        // Process
    }
}
```

## Core Components

### Configuration Management

Configuration follows precedence order:

1. Command-line flags (highest)
2. Environment variables
3. Configuration file
4. Default values (lowest)

```go
// Configuration loading
cfg := config.New()
cfg.LoadFile("zen.yaml")
cfg.LoadEnv()
cfg.LoadFlags(cmd)
```

### Error Handling

Consistent error handling across the codebase:

```go
// Sentinel errors for known conditions
var (
    ErrNotFound = errors.New("not found")
    ErrInvalid  = errors.New("invalid input")
)

// Wrapped errors with context
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Error types for rich information
type ValidationError struct {
    Field   string
    Message string
}
```

### Logging

Structured logging with levels:

```go
logger.Debug("Processing request", 
    "user", userID,
    "action", action)

logger.Info("Operation completed",
    "duration", time.Since(start))

logger.Error("Operation failed",
    "error", err,
    "retry", retryCount)
```

### I/O Streams

Abstracted I/O for testability:

```go
type IOStreams struct {
    In     io.ReadCloser
    Out    io.Writer
    ErrOut io.Writer
    
    // Terminal detection
    IsTerminal    bool
    IsInteractive bool
    
    // Color support
    ColorEnabled bool
}
```

## Plugin Architecture

### Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Init(context.Context) error
    Execute(context.Context, []string) error
    Close() error
}
```

### Plugin Loading

```go
// Discover plugins
plugins := plugin.Discover("~/.zen/plugins")

// Load and initialize
for _, p := range plugins {
    if err := p.Init(ctx); err != nil {
        logger.Warn("Plugin init failed", 
            "plugin", p.Name(),
            "error", err)
    }
}
```

## LLM Integration

### Provider Abstraction

```go
type LLMProvider interface {
    Complete(ctx context.Context, prompt string) (string, error)
    Stream(ctx context.Context, prompt string) (<-chan string, error)
    Models() []string
}

// Multiple provider support
type ProviderRegistry struct {
    providers map[string]LLMProvider
    default   string
}
```

### Prompt Management

```go
type Prompt struct {
    System  string
    User    string
    Context map[string]interface{}
}

// Template-based prompts
template := prompt.Load("code-review")
rendered := template.Render(context)
```

## Workflow Engine

### Workflow Definition

```go
type Workflow struct {
    Name  string
    Steps []Step
}

type Step struct {
    Name     string
    Type     StepType
    Action   func(context.Context) error
    Requires []string  // Dependencies
}
```

### Workflow Execution

```go
engine := workflow.NewEngine()
wf := workflow.Load("development")
result := engine.Execute(ctx, wf)
```

## Testing Architecture

### Test Helpers

```go
// Test factory for commands
func NewTestFactory() *TestFactory {
    streams := iostreams.Test()
    return &TestFactory{
        streams: streams,
        config:  config.TestConfig(),
    }
}

// Golden file testing
func TestCommand(t *testing.T) {
    actual := runCommand()
    golden := filepath.Join("testdata", "expected.txt")
    assert.Equal(t, golden, actual)
}
```

### Mock Services

```go
type MockLLMProvider struct {
    mock.Mock
}

func (m *MockLLMProvider) Complete(ctx context.Context, prompt string) (string, error) {
    args := m.Called(ctx, prompt)
    return args.String(0), args.Error(1)
}
```

## Security Considerations

### Input Validation

```go
// Validate all user input
func ValidatePath(path string) error {
    if !filepath.IsAbs(path) {
        path = filepath.Clean(path)
    }
    
    if strings.Contains(path, "..") {
        return ErrInvalidPath
    }
    
    return nil
}
```

### Secret Management

```go
// Never log secrets
type Config struct {
    APIKey string `json:"-"` // Excluded from JSON
}

// Use environment variables
apiKey := os.Getenv("ZEN_API_KEY")
if apiKey == "" {
    return errors.New("API key required")
}
```

## Performance Patterns

### Concurrent Processing

```go
// Worker pool pattern
func ProcessItems(items []Item) error {
    workers := runtime.NumCPU()
    ch := make(chan Item, len(items))
    
    var wg sync.WaitGroup
    for i := 0; i < workers; i++ {
        wg.Add(1)
        go worker(ch, &wg)
    }
    
    for _, item := range items {
        ch <- item
    }
    close(ch)
    
    wg.Wait()
    return nil
}
```

### Caching

```go
type Cache struct {
    mu    sync.RWMutex
    items map[string]interface{}
    ttl   time.Duration
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    
    item, found := c.items[key]
    return item, found
}
```

## Best Practices

### Code Organization

- Keep packages focused and cohesive
- Avoid circular dependencies
- Use internal/ for private code
- Export only necessary types

### Error Design

- Create sentinel errors for known conditions
- Wrap errors with context
- Use error types for rich information
- Handle errors at appropriate level

### Concurrency

- Prefer channels over shared memory
- Use context for cancellation
- Avoid goroutine leaks
- Test concurrent code with race detector

### Testing

- Write tests alongside code
- Use table-driven tests
- Mock external dependencies
- Test error conditions

## Architecture Decision Records

Key architectural decisions are documented in ADRs:

- [ADR-0001](../architecture/decisions/ADR-0001-language-choice.md) - Go as primary language
- [ADR-0002](../architecture/decisions/ADR-0002-cli-framework.md) - Cobra for CLI
- [ADR-0006](../architecture/decisions/ADR-0006-factory-pattern.md) - Factory pattern
- [ADR-0010](../architecture/decisions/ADR-0010-llm-abstraction.md) - LLM abstraction

See [Architecture Decisions](../architecture/decisions/) for complete list.

## Future Considerations

### Planned Improvements

- GraphQL API support
- Real-time collaboration
- Distributed execution
- Cloud-native deployment

### Technical Debt

- Migrate legacy configuration
- Improve error messages
- Add telemetry support
- Enhance plugin discovery

## Next Steps

- Review [Development Workflow](development-workflow.md)
- Understand [Testing](testing.md) practices
- Check [Code Review](code-review.md) standards

---

Questions about architecture? Review ADRs or discuss with maintainers.
