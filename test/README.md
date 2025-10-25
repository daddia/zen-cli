# Zen CLI Testing Strategy

This directory contains the testing infrastructure for the Zen CLI, following Go best practices and the test pyramid approach.

## Test Organization

### Test Types Distribution
- **70% Unit Tests** - Fast, isolated, co-located with source code
- **20% Integration Tests** - Component interactions in `test/integration/`
- **10% End-to-End Tests** - Complete user workflows in `test/e2e/`

### Directory Structure

```
test/
├── README.md           # This file - testing overview
├── e2e/               # End-to-end tests (build tag: e2e)
│   ├── setup.go       # Test environment setup
│   ├── *_test.go      # E2E test files
│   └── README.md      # E2E-specific documentation
├── integration/       # Integration tests (build tag: integration)
│   └── *_test.go      # Integration test files
└── fixtures/          # Shared test data and fixtures
    └── assets/        # Asset-related test data

# Unit tests are co-located with source code:
pkg/cmd/status/
├── status.go
├── status_test.go     # Unit tests for status command
└── testdata/          # Test fixtures for status package
```

## Running Tests

### All Tests
```bash
make test              # Run complete test suite
make test-all          # Run unit + integration + e2e
```

### By Type
```bash
make test-unit         # Unit tests only (fast)
make test-integration  # Integration tests
make test-e2e          # End-to-end tests
```

### Specific Categories
```bash
make test-e2e-core     # Core commands e2e tests
make test-e2e-workspace # Workspace operations e2e tests
make test-e2e-assets   # Assets management e2e tests
```

### Coverage
```bash
make test-coverage     # Generate HTML coverage report
make test-coverage-ci  # CI coverage validation (80% minimum)
```

## Test Standards

### Unit Tests
- **Location**: Co-located with source code (`*_test.go`)
- **Package**: Use `package_test` for external API testing
- **Coverage**: ≥90% for business logic, ≥80% overall
- **Speed**: <30 seconds total execution time
- **Dependencies**: Mock all external dependencies

### Integration Tests
- **Location**: `test/integration/`
- **Build Tag**: `//go:build integration`
- **Coverage**: Component interactions and real dependencies
- **Speed**: <60 seconds total execution time
- **Dependencies**: Real services (database, APIs) where needed

### End-to-End Tests
- **Location**: `test/e2e/`
- **Build Tag**: `//go:build e2e`
- **Coverage**: Complete user workflows
- **Speed**: <300 seconds total execution time
- **Environment**: Isolated test environment (`../zen-test`)

## Test Utilities

### Shared Fixtures
- `test/fixtures/` - Common test data across test types
- `pkg/*/testdata/` - Package-specific test fixtures

### Test Helpers
- `test/e2e/setup.go` - E2E environment management
- `pkg/cmdutil/factory.go` - Test factory for CLI commands
- `pkg/iostreams/iostreams.go` - Test streams for output capture

## Best Practices

### Writing Tests
1. **Follow AAA Pattern**: Arrange, Act, Assert
2. **Use Table-Driven Tests**: For multiple scenarios
3. **Test Error Conditions**: Not just happy paths
4. **Use Descriptive Names**: Clearly state scenario and expectation
5. **Keep Tests Independent**: No shared state between tests

### Test Data
1. **Use `testdata/` Directories**: For package-specific fixtures
2. **Keep Data Minimal**: Only what's needed for the test
3. **Use Golden Files**: For complex output comparisons
4. **Clean Up Resources**: Use `t.Cleanup()` or `defer`

### Mocking
1. **Mock External Dependencies**: Network, filesystem, databases
2. **Use Interface-Based Mocking**: For better testability
3. **Verify Mock Expectations**: Ensure mocks were called correctly
4. **Keep Mocks Simple**: Focus on the behavior being tested

## Quality Gates

All tests must pass these quality gates:

- ✅ **Coverage**: ≥80% overall, ≥90% business logic
- ✅ **Performance**: Unit tests <30s, Integration <60s, E2E <300s
- ✅ **Race Conditions**: All tests pass with `-race` flag
- ✅ **Isolation**: Tests can run in any order
- ✅ **Cleanup**: No test artifacts left behind

## Continuous Integration

Tests run automatically on:
- Every pull request
- Every commit to main branch
- Nightly for full regression testing

See `.github/workflows/test.yml` for CI configuration.

## Troubleshooting

### Common Issues
1. **Test Timeouts**: Increase timeout or optimize test performance
2. **Flaky Tests**: Usually indicates shared state or timing issues
3. **Coverage Failures**: Add tests for uncovered code paths
4. **Race Conditions**: Use proper synchronization or avoid shared state

### Debug Commands
```bash
go test -v ./...                    # Verbose output
go test -race ./...                 # Race condition detection
go test -timeout=60s ./...          # Custom timeout
go test -run TestSpecificTest ./... # Run specific test
```

For more detailed testing guidance, see:
- [Testing Guide](../docs/contributing/testing.md)
- [Unit Test Standards](../.cursor/rules/test-unit.mdc)
- [Test Standards](../.cursor/rules/test-standards.mdc)
