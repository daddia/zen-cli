# Zen CLI Testing Guide

## Test Pyramid Implementation

This project implements a comprehensive test pyramid following the 70/20/10 distribution:

- **70% Unit Tests**: Fast, isolated tests for business logic
- **20% Integration Tests**: Component interaction validation  
- **10% End-to-End Tests**: Critical user journey verification

## Test Structure

```
test/
├── integration/          # Integration tests (build tag: integration)
│   └── cli_integration_test.go
├── e2e/                 # End-to-end tests (build tag: e2e)
│   └── user_journeys_test.go
└── testdata/            # Test fixtures and data

internal/
├── config/
│   └── config_test.go   # Unit tests for configuration
├── logging/
│   └── logger_test.go   # Unit tests for logging
├── workspace/
│   └── workspace_test.go # Unit tests for workspace
└── zencmd/
    └── cmd_test.go      # Unit tests for command execution

pkg/
├── cmd/
│   ├── init/
│   │   └── init_test.go # Unit tests for init command
│   ├── root/
│   │   └── root_test.go # Unit tests for root command
│   ├── status/
│   │   └── status_test.go # Unit tests for status command
│   └── version/
│       └── version_test.go # Unit tests for version command
├── cmdutil/
│   └── factory_test.go  # Unit tests for command utilities
├── errors/
│   └── errors_test.go   # Unit tests for error handling
├── iostreams/
│   └── iostreams_test.go # Unit tests for I/O streams
└── types/
    └── common_test.go   # Unit tests for common types
```

## Running Tests

### Complete Test Pyramid
```bash
make test-pyramid
```

### Individual Test Types
```bash
# Unit tests (70% of suite) - Target: <30s execution
make test-unit

# Integration tests (20% of suite) - Target: <1min execution  
make test-integration

# End-to-end tests (10% of suite) - Target: <2min execution
make test-e2e
```

### Development Testing
```bash
# Fast unit tests (no race detection)
make test-unit-fast

# Watch for changes and run tests
make test-watch

# Run tests with maximum parallelization
make test-parallel

# Run only short tests
make test-short
```

### Coverage and Quality
```bash
# Generate coverage report
make test-coverage

# Coverage analysis with targets validation
make test-coverage-report

# CI coverage validation (strict)
make test-coverage-ci

# Race condition detection
make test-race

# Benchmark tests
make test-benchmarks
```

## Coverage Targets

### Overall Coverage Goals
- **≥90% coverage** for business logic packages (`internal/`, core `pkg/` packages)
- **≥80% overall coverage** across the entire codebase
- **Zero flakiness** - all tests must be deterministic and repeatable

### Coverage by Package Type
- **Business Logic** (`internal/config`, `internal/logging`, `internal/workspace`): ≥90%
- **Command Packages** (`pkg/cmd/*`): ≥90% 
- **Utility Packages** (`pkg/cmdutil`, `pkg/errors`, `pkg/types`): ≥90%
- **I/O and Streams** (`pkg/iostreams`): ≥90%

## Test Categories

### Unit Tests (70%)
Fast, isolated tests that execute in <30 seconds total:

- **Business Logic**: Configuration management, logging, workspace operations
- **Command Logic**: CLI command implementations with mocked dependencies
- **Utilities**: Error handling, type definitions, I/O abstractions
- **Characteristics**: No external dependencies, mocked filesystem/network

### Integration Tests (20%)
Component interaction tests that execute in <1 minute total:

- **CLI Workflows**: Complete command execution with real filesystem
- **Configuration Loading**: Multi-source config with precedence testing
- **Workspace Operations**: Project detection and initialization
- **Error Scenarios**: Error handling across component boundaries

### End-to-End Tests (10%)
Critical user journey tests that execute in <2 minutes total:

- **User Workflows**: `init` → `config` → `status` complete journeys
- **Cross-platform**: Binary execution on Linux, macOS, Windows
- **Output Formats**: JSON, YAML, text output validation
- **Error Handling**: Real-world error scenarios and recovery

## Test Quality Standards

### Deterministic Tests
- All tests must pass consistently across runs
- No reliance on timing, random values, or external state
- Proper cleanup of temporary resources
- Isolated test execution (no shared state)

### Performance Standards
- **Unit tests**: Individual test <1s, total suite <30s
- **Integration tests**: Individual test <5s, total suite <60s  
- **End-to-end tests**: Individual test <30s, total suite <120s

### Test Data Management
- Use `t.TempDir()` for temporary directories
- Store test fixtures in `testdata/` directories
- Keep test data minimal and focused
- Use table-driven tests for multiple scenarios

## CI/CD Integration

### GitHub Actions Workflow
The test pyramid is enforced in CI with the following stages:

1. **Unit Tests**: Fast execution with coverage validation
2. **Integration Tests**: Component interaction validation
3. **E2E Tests**: Cross-platform user journey testing
4. **Performance Tests**: Benchmark regression detection
5. **Test Pyramid Validation**: Distribution ratio enforcement

### Quality Gates
All the following must pass for merge approval:

- ✅ All tests pass with zero flakiness
- ✅ Coverage targets met (90%/80%)
- ✅ Test pyramid distribution maintained (70/20/10)
- ✅ Performance benchmarks within thresholds
- ✅ No race conditions detected
- ✅ Security scans pass

## Writing New Tests

### Unit Test Template
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        {
            name:     "invalid input",
            input:    invalidInput,
            expected: OutputType{},
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := FunctionName(tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Integration Test Template
```go
//go:build integration

func TestIntegrationScenario(t *testing.T) {
    // Setup test environment
    tempDir := t.TempDir()
    
    // Execute integration scenario
    // ... test implementation
    
    // Verify results
    // ... assertions
}
```

### E2E Test Template
```go
//go:build e2e

func TestE2E_UserJourney(t *testing.T) {
    // Setup: Build binary, prepare environment
    
    // Execute: Run real CLI commands
    
    // Verify: Check outputs and side effects
}
```

## Test Utilities

### Mock Factory
Use `cmdutil.NewTestFactory()` for creating test dependencies:

```go
streams := iostreams.Test()
factory := cmdutil.NewTestFactory(streams)
```

### Command Testing
Use the provided `Execute` function for testing commands:

```go
ctx := context.Background()
err := Execute(ctx, []string{"status"}, streams)
require.NoError(t, err)
```

### Assertions
- Use `testify/require` for critical assertions that should stop the test
- Use `testify/assert` for non-critical assertions that allow test continuation
- Always include meaningful error messages in assertions

## Debugging Tests

### Verbose Output
```bash
make test-verbose
```

### Individual Test Execution
```bash
# Run specific test
go test -v -run TestSpecificTest ./pkg/cmd/status

# Run with race detection
go test -race -run TestSpecificTest ./pkg/cmd/status

# Run integration tests
go test -tags=integration -v ./test/integration

# Run E2E tests  
go test -tags=e2e -v ./test/e2e
```

### Coverage Analysis
```bash
# Generate detailed coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Function-level coverage
go tool cover -func=coverage.out
```

## Best Practices

### Test Organization
- Group related tests in the same file
- Use descriptive test names that explain the scenario
- Follow the AAA pattern: Arrange, Act, Assert
- Keep tests focused and atomic

### Mock Usage
- Mock external dependencies (filesystem, network, databases)
- Use dependency injection to enable mocking
- Keep mocks simple and behavior-focused
- Verify mock expectations where appropriate

### Error Testing
- Test both success and failure paths
- Verify specific error types and messages
- Test error propagation across layers
- Include edge cases and boundary conditions

### Performance Considerations
- Use `testing.Short()` to skip long-running tests when appropriate
- Implement benchmark tests for performance-critical code
- Monitor test execution times and optimize slow tests
- Use parallel test execution where safe (`t.Parallel()`)

## Troubleshooting

### Common Issues
- **Flaky tests**: Usually caused by timing, concurrency, or external dependencies
- **Slow tests**: Often due to unnecessary I/O or lack of parallelization
- **Coverage gaps**: Missing edge cases or error condition testing
- **Build failures**: Import cycles, missing dependencies, or syntax errors

### Solutions
- Use `t.TempDir()` instead of hardcoded paths
- Mock external dependencies consistently
- Add proper cleanup with `defer` statements
- Use build constraints for integration and E2E tests
- Run tests with race detection during development

For more detailed information, see the [Architecture Decision Records](architecture/decisions/) and [Quality Gates Framework](architecture/decisions/ADR-0014-quality-gates.md).
