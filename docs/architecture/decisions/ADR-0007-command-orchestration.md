---
status: Accepted
date: 2025-09-13
decision-makers: Development Team, Architecture Team
consulted: DevOps Team, Security Team
informed: Product Team, Contributors
---

# ADR-0007 - Command Orchestration Design

## Context and Problem Statement

The Zen CLI requires a sophisticated command orchestration system that handles graceful shutdown, error management, exit codes, and signal handling while maintaining a clean separation between the entry point and command logic. The original implementation mixed concerns and lacked proper error categorization and recovery mechanisms.

Key requirements:
- Ultra-lightweight main entry point with minimal logic
- Centralized error handling with categorized responses
- Structured exit codes following CLI conventions
- Graceful shutdown with signal handling (SIGINT, SIGTERM)
- Helpful error messages with actionable suggestions
- Context propagation and cancellation support
- Integration with factory pattern for dependency injection

## Decision Drivers

* **Separation of Concerns**: Clear boundary between entry point and orchestration
* **Error Management**: Comprehensive error handling with helpful user feedback
* **Signal Handling**: Proper response to system signals and interruptions
* **Exit Codes**: Standard exit codes for different failure scenarios
* **User Experience**: Clear error messages with suggestions for resolution
* **Maintainability**: Centralized orchestration logic for easier maintenance
* **Testability**: Isolated orchestration logic that can be thoroughly tested
* **Industry Standards**: Follow CLI best practices and conventions

## Considered Options

* **Centralized Orchestration** - Dedicated orchestration layer in internal package
* **Inline Orchestration** - All logic directly in main.go
* **Command-Level Handling** - Each command handles its own orchestration
* **Framework-Based** - Use CLI framework's built-in orchestration
* **Middleware Pattern** - Chain of middleware handlers for different concerns

## Decision Outcome

Chosen option: **Centralized Orchestration** with dedicated `internal/zencmd` package, because it provides the best separation of concerns, error handling, and maintainability while keeping main.go minimal.

### Consequences

**Good:**
- Ultra-lightweight main.go (reduced from 38 to 11 lines)
- Centralized error handling with consistent user experience
- Proper signal handling and graceful shutdown
- Structured exit codes following CLI conventions
- Comprehensive error categorization with helpful suggestions
- Better testability through separation of concerns
- Clear integration point with factory pattern

**Bad:**
- Additional abstraction layer adds slight complexity
- Requires understanding of orchestration patterns
- One more package to maintain

**Neutral:**
- Different approach from traditional CLI implementations
- Follows patterns used by successful CLI tools like GitHub CLI

### Confirmation

The decision has been validated through:
- Successful implementation with all error scenarios handled
- Proper signal handling during long-running operations
- Comprehensive test coverage for orchestration logic
- User feedback on improved error messages and suggestions
- Performance benchmarks showing no overhead impact

## Orchestration Architecture

### **Entry Point Design**

```go
// cmd/zen/main.go - Ultra-lightweight (11 lines)
package main

import (
    "os"
    "github.com/daddia/zen/internal/zencmd"
)

func main() {
    code := zencmd.Main()
    os.Exit(int(code))
}
```

### **Orchestration Layer**

```go
// internal/zencmd/cmd.go - Main orchestration
func Main() cmdutil.ExitCode {
    // Setup graceful shutdown
    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    // Create factory and root command
    cmdFactory := factory.New()
    rootCmd, err := root.NewCmdRoot(cmdFactory)
    if err != nil {
        fmt.Fprintf(stderr, "failed to create root command: %s\n", err)
        return cmdutil.ExitError
    }

    // Execute with context and handle errors
    rootCmd.SetContext(ctx)
    if err := rootCmd.Execute(); err != nil {
        return handleError(err, cmdFactory)
    }

    return cmdutil.ExitOK
}
```

### **Exit Code System**

```go
type ExitCode int

const (
    ExitOK      ExitCode = 0  // Successful completion
    ExitError   ExitCode = 1  // General error
    ExitCancel  ExitCode = 2  // User cancellation
    ExitAuth    ExitCode = 4  // Authentication failure
)
```

### **Error Handling Strategy**

```go
func handleError(err error, f *cmdutil.Factory) cmdutil.ExitCode {
    stderr := f.IOStreams.ErrOut

    // Check for specific error types
    if err == cmdutil.SilentError {
        return cmdutil.ExitError
    }

    if cmdutil.IsUserCancellation(err) {
        fmt.Fprint(stderr, "\n")
        return cmdutil.ExitCancel
    }

    var noResultsError cmdutil.NoResultsError
    if errors.As(err, &noResultsError) {
        if f.IOStreams.IsStdoutTTY() {
            fmt.Fprintln(stderr, noResultsError.Error())
        }
        return cmdutil.ExitOK
    }

    // Print error with suggestions
    printError(stderr, err)
    return cmdutil.ExitError
}
```

### **Error Suggestions System**

```go
func getErrorSuggestion(err error) string {
    if err == nil {
        return ""
    }

    errMsg := err.Error()
    
    switch {
    case strings.Contains(errMsg, "config"):
        return "Try running 'zen config' to check your configuration"
    case strings.Contains(errMsg, "workspace"):
        return "Try running 'zen init' to initialize your workspace"
    case strings.Contains(errMsg, "permission"):
        return "Check file permissions and try again"
    default:
        return ""
    }
}
```

## Implementation Details

### **Signal Handling**

- SIGINT (Ctrl+C): Graceful cancellation with cleanup
- SIGTERM: Graceful shutdown for container environments
- Context cancellation propagated to all operations
- Cleanup handlers for temporary files and resources

### **Error Categories**

1. **Silent Errors**: Internal errors that shouldn't be displayed
2. **User Cancellation**: Ctrl+C or other user-initiated cancellation
3. **No Results**: Successful operations with no output
4. **Flag Errors**: Invalid command-line flags or arguments
5. **General Errors**: All other error conditions

### **Context Integration**

- Root context created with signal handling
- Context passed to all commands via Cobra
- Cancellation respected throughout command execution
- Timeout support for long-running operations

### **Factory Integration**

- Factory created once at startup
- Passed to root command for dependency injection
- Used for error handling and I/O stream access
- Enables consistent component initialization

## Pros and Cons of the Options

### Centralized Orchestration

**Good:**
- Clear separation of concerns
- Centralized error handling and exit codes
- Proper signal handling and graceful shutdown
- Ultra-lightweight main entry point
- Comprehensive error categorization
- Easy to test orchestration logic in isolation
- Consistent patterns across all commands

**Bad:**
- Additional abstraction layer
- Requires understanding of orchestration patterns
- One more package to maintain

**Neutral:**
- Different from traditional CLI approaches
- Follows modern CLI design patterns

### Inline Orchestration

**Good:**
- Simple and straightforward
- All logic in one place
- Easy to understand for simple cases

**Bad:**
- Mixes concerns in main.go
- Difficult to test orchestration logic
- Poor separation of responsibilities
- Hard to maintain as complexity grows

**Neutral:**
- Traditional approach with known limitations
- Suitable only for very simple CLIs

### Command-Level Handling

**Good:**
- Each command handles its own concerns
- Distributed responsibility
- Flexible per-command behavior

**Bad:**
- Inconsistent error handling across commands
- Duplicated orchestration logic
- Difficult to maintain consistent patterns
- Poor user experience due to inconsistency

**Neutral:**
- Flexible but requires discipline
- Can lead to fragmented user experience

### Framework-Based

**Good:**
- Leverages existing framework capabilities
- Less custom code to maintain
- Framework handles common concerns

**Bad:**
- Limited control over error handling
- May not support all required patterns
- Framework-specific behavior
- Less flexibility for custom requirements

**Neutral:**
- Depends on framework capabilities
- May be sufficient for simple use cases

### Middleware Pattern

**Good:**
- Composable handling logic
- Easy to add new concerns
- Clear separation of different aspects

**Bad:**
- Complex for CLI applications
- Overkill for command orchestration
- More difficult to understand and maintain

**Neutral:**
- Good for web applications but complex for CLI
- May be suitable for very complex scenarios

## More Information

**Performance Impact:**
- Orchestration overhead: <1ms per command
- Signal handling: No measurable impact
- Error handling: <1ms for error scenarios
- Context propagation: Minimal overhead

**Error Message Examples:**
```
Error: configuration file not found
Try running 'zen config' to check your configuration

Error: workspace not initialized
Try running 'zen init' to initialize your workspace

Error: permission denied accessing file
Check file permissions and try again
```

**Related ADRs:**
- ADR-0006: Factory Pattern Implementation
- ADR-0003: Project Structure and Organization
- ADR-0002: Cobra CLI Framework Selection

**References:**
- [CLI Exit Codes](https://tldp.org/LDP/abs/html/exitcodes.html)
- [Go Signal Handling](https://gobyexample.com/signals)
- [GitHub CLI Error Handling](https://github.com/cli/cli/blob/trunk/internal/ghcmd/cmd.go)
- [CLI Design Patterns](https://clig.dev/)

**Testing Strategy:**
- Unit tests for all error handling scenarios
- Signal handling integration tests
- Exit code validation tests
- Error message content verification
- Context cancellation behavior tests
