# Design Patterns Quick Reference

This guide provides a quick reference to the design patterns used in Zen CLI. For comprehensive documentation, see the **[Architecture Design Patterns Guide](../architecture/patterns/design-patterns.md)**.

## Key Patterns for Contributors

### 1. Factory Pattern (Dependency Injection)
Used throughout for dependency management. All commands receive dependencies via factory.

```go
func NewCmdInit(f cmdutil.Factory) *cobra.Command {
    opts := &InitOptions{
        Factory: f,
        IO:      f.IOStreams(),
    }
    // ...
}
```

**See**: [Full Factory Pattern Documentation](../architecture/patterns/design-patterns.md#1-factory-pattern)

### 2. Command Pattern (CLI Structure)
Every CLI command follows Cobra's command pattern:

```go
cmd := &cobra.Command{
    Use:   "command",
    Short: "Brief description",
    RunE: func(cmd *cobra.Command, args []string) error {
        return runCommand(opts)
    },
}
```

**See**: [Command Pattern Details](../architecture/patterns/design-patterns.md#2-command-pattern)

### 3. Error Handling Pattern
Consistent error handling with wrapping and context:

```go
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

**See**: [Error Handling Best Practices](../architecture/patterns/design-patterns.md#best-practices)

## Quick Reference Table

| Pattern | Usage | Location | Details |
|---------|-------|----------|---------|
| Factory | Dependency injection | `pkg/cmd/factory/` | [Details](../architecture/patterns/design-patterns.md#1-factory-pattern) |
| Command | CLI commands | `pkg/cmd/*/` | [Details](../architecture/patterns/design-patterns.md#2-command-pattern) |
| Strategy | LLM providers | `internal/agents/providers/` | [Details](../architecture/patterns/design-patterns.md#3-strategy-pattern) |
| Observer | Event handling | `internal/workflow/` | [Details](../architecture/patterns/design-patterns.md#4-observer-pattern) |
| Repository | Data access | `internal/storage/` | [Details](../architecture/patterns/design-patterns.md#5-repository-pattern) |

## Common Development Patterns

### Creating a New Command

1. Use the factory pattern for dependencies
2. Follow the command structure template
3. Implement proper error handling
4. Add comprehensive tests

```go
// Standard command structure
type CommandOptions struct {
    Factory cmdutil.Factory
    IO      *iostreams.IOStreams
    // Command-specific fields
}

func NewCommand(f cmdutil.Factory) *cobra.Command {
    opts := &CommandOptions{
        Factory: f,
        IO:      f.IOStreams(),
    }
    
    cmd := &cobra.Command{
        Use:   "mycommand",
        Short: "Description",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runCommand(opts)
        },
    }
    
    return cmd
}
```

### Testing with Mocks

Use the factory pattern to inject test dependencies:

```go
func TestCommand(t *testing.T) {
    f := &cmdutil.TestFactory{
        IOStreams: iostreams.Test(),
        Config: func() (*config.Config, error) {
            return &config.Config{}, nil
        },
    }
    
    cmd := NewCommand(f)
    // Test command execution
}
```

## Best Practices Summary

1. **Use Dependency Injection** - Never create dependencies directly
2. **Follow Interface Segregation** - Keep interfaces small and focused  
3. **Handle Errors Consistently** - Wrap errors with context
4. **Write Testable Code** - Use interfaces and dependency injection
5. **Avoid Anti-Patterns** - No god objects, circular dependencies, or tight coupling

## Learn More

- **[Complete Design Patterns Guide](../architecture/patterns/design-patterns.md)** - Full documentation with diagrams
- **[Integration Patterns](../architecture/patterns/integration-patterns.md)** - External system integration
- **[Security Patterns](../architecture/patterns/security-patterns.md)** - Security best practices
- **[Architecture Decision Records](../architecture/decisions/register.md)** - Design decisions and rationale

---

*This is a quick reference. For detailed explanations, diagrams, and examples, see the [Architecture Design Patterns Guide](../architecture/patterns/design-patterns.md).*
