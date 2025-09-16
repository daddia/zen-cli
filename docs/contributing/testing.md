# Testing Guide

A practical guide to writing and running tests for Zen.

## Testing Philosophy

We follow the test pyramid approach:
- **70% Unit Tests** - Fast, isolated, focused
- **20% Integration Tests** - Component interactions
- **10% End-to-End Tests** - Critical user journeys

## Writing Tests

### Test Structure

All test files follow the `*_test.go` naming convention and live alongside the code they test.

```go
func TestFunctionName(t *testing.T) {
    // Arrange - Set up test data
    input := "test data"
    expected := "expected result"
    
    // Act - Execute the function
    result := FunctionToTest(input)
    
    // Assert - Verify the result
    if result != expected {
        t.Errorf("got %v, want %v", result, expected)
    }
}
```

### Table-Driven Tests

Use table-driven tests for multiple scenarios:

```go
func TestCalculate(t *testing.T) {
    tests := []struct {
        name     string
        input    int
        expected int
    }{
        {"positive", 5, 10},
        {"negative", -5, -10},
        {"zero", 0, 0},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Calculate(tt.input)
            if result != tt.expected {
                t.Errorf("Calculate(%d) = %d; want %d", 
                    tt.input, result, tt.expected)
            }
        })
    }
}
```

### Testing Commands

Use the test factory for CLI commands:

```go
func TestStatusCommand(t *testing.T) {
    // Create test streams
    streams := iostreams.Test()
    factory := cmdutil.NewTestFactory(streams)
    
    // Execute command
    cmd := NewCmdStatus(factory)
    cmd.SetArgs([]string{"--json"})
    
    err := cmd.Execute()
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    
    // Verify output
    out := streams.Out.String()
    if !strings.Contains(out, "status") {
        t.Errorf("expected status in output, got: %s", out)
    }
}
```

## Running Tests

### Quick Commands

```bash
# Run all unit tests
make test-unit

# Run specific package tests
go test ./pkg/cmd/init

# Run specific test function
go test -run TestInit ./pkg/cmd/init

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...
```

### Test Categories

#### Unit Tests
Fast, isolated tests that mock external dependencies:

```bash
# Run unit tests (no build tags)
make test-unit

# Target: <30 seconds total execution
```

#### Integration Tests
Test component interactions with real dependencies:

```bash
# Run integration tests
make test-integration

# Or with build tag
go test -tags=integration ./test/integration
```

#### End-to-End Tests
Test complete user workflows:

```bash
# Run E2E tests
make test-e2e

# Or with build tag
go test -tags=e2e ./test/e2e
```

## Test Coverage

### Coverage Standards

- **Business Logic**: ≥90% coverage
- **Commands**: ≥90% coverage
- **Overall**: ≥80% coverage

### Checking Coverage

```bash
# Generate coverage report
make test-coverage

# View HTML report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Check specific package
go test -cover ./pkg/cmd/init
```

## Test Utilities

### Mock Objects

Create mocks for external dependencies:

```go
type MockClient struct {
    Response string
    Error    error
}

func (m *MockClient) Get() (string, error) {
    return m.Response, m.Error
}
```

### Test Helpers

Common test utilities:

```go
// TempDir creates a temporary directory
func TempDir(t *testing.T) string {
    t.Helper()
    return t.TempDir()
}

// Golden compares output to golden file
func Golden(t *testing.T, actual, golden string) {
    t.Helper()
    // comparison logic
}
```

### Test Fixtures

Store test data in `testdata/` directories:

```
pkg/cmd/init/
├── init.go
├── init_test.go
└── testdata/
    ├── config.yaml
    └── expected.json
```

## Best Practices

### Do's

- Write tests alongside implementation
- Use descriptive test names
- Test both success and failure cases
- Keep tests independent
- Use `t.Helper()` in test utilities
- Clean up resources with `t.Cleanup()`

### Don'ts

- Don't test implementation details
- Don't use real external services
- Don't rely on test execution order
- Don't ignore flaky tests
- Don't skip error checking

## Common Testing Patterns

### Testing Errors

```go
func TestErrorHandling(t *testing.T) {
    _, err := FunctionThatErrors()
    if err == nil {
        t.Fatal("expected error, got nil")
    }
    
    if !strings.Contains(err.Error(), "expected") {
        t.Errorf("error = %v; want containing 'expected'", err)
    }
}
```

### Testing with Context

```go
func TestWithContext(t *testing.T) {
    ctx, cancel := context.WithTimeout(
        context.Background(), 
        100*time.Millisecond,
    )
    defer cancel()
    
    err := FunctionWithContext(ctx)
    if err != nil {
        t.Errorf("unexpected error: %v", err)
    }
}
```

### Testing Concurrent Code

```go
func TestConcurrent(t *testing.T) {
    var wg sync.WaitGroup
    results := make([]int, 10)
    
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(index int) {
            defer wg.Done()
            results[index] = Process(index)
        }(i)
    }
    
    wg.Wait()
    // verify results
}
```

## Debugging Tests

### Verbose Output

```bash
# See detailed test output
go test -v ./pkg/...

# See test names only
go test -v -run=nothingmatches ./pkg/...
```

### Race Detection

```bash
# Check for race conditions
go test -race ./...
make test-race
```

### Test Caching

```bash
# Clear test cache
go clean -testcache

# Disable cache for single run
go test -count=1 ./...
```

## Performance Testing

### Benchmarks

```go
func BenchmarkFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        FunctionToBenchmark()
    }
}
```

Run benchmarks:

```bash
# Run benchmarks
go test -bench=. ./pkg/...

# Compare benchmarks
go test -bench=. -benchmem ./pkg/...
```

## Next Steps

- Review [Code Review](code-review.md) process
- Understand [Architecture](architecture.md) patterns
- Learn about [Documentation](documentation.md) standards

---

Questions? Check existing tests for examples or ask in discussions.
