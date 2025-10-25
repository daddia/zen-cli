# Contributing to Zen CLI Tests

This guide explains how to add and maintain tests in the Zen CLI project.

## Adding New Tests

### 1. Unit Tests (70% of test suite)

**Location**: Co-located with source code
**When**: Testing individual functions/methods in isolation

```go
// In pkg/cmd/status/status_test.go
package status_test

import (
    "testing"
    "github.com/daddia/zen/pkg/cmd/status"
    "github.com/daddia/zen/test/testutil"
)

func TestStatusCommand(t *testing.T) {
    // Use shared test utilities
    workspace := testutil.CreateTempWorkspace(t)
    
    // Test implementation
}
```

### 2. Integration Tests (20% of test suite)

**Location**: `test/integration/`
**When**: Testing component interactions with real dependencies

```go
// In test/integration/new_feature_test.go
//go:build integration

package integration

import (
    "testing"
    "github.com/daddia/zen/test/testutil"
)

func TestFeatureIntegration(t *testing.T) {
    // Use shared fixtures
    config := testutil.LoadFixture(t, "configs/complete.yaml")
    
    // Test implementation
}
```

### 3. End-to-End Tests (10% of test suite)

**Location**: `test/e2e/`
**When**: Testing complete user workflows

```go
// In test/e2e/new_workflow_test.go
//go:build e2e

package e2e

func TestE2E_NewWorkflow(t *testing.T) {
    env := SetupTestEnvironment(t)
    defer TeardownTestEnvironment(t, env)
    
    workspaceDir := env.CreateTestWorkspace(t, "new-workflow-test")
    defer env.CleanupTestWorkspace(t, workspaceDir)
    
    // Test implementation
}
```

## Using Shared Test Utilities

### Golden Files
```go
import "github.com/daddia/zen/test/testutil"

func TestOutput(t *testing.T) {
    actual := generateOutput()
    testutil.Golden(t, actual, "test/fixtures/golden/expected.txt")
}

// Update golden files when output changes:
// go test -update-golden ./...
```

### Workspace Helpers
```go
import "github.com/daddia/zen/test/testutil"

func TestWorkspace(t *testing.T) {
    // Create temporary workspace with basic structure
    workspace := testutil.CreateTempWorkspace(t)
    
    // Create project files for testing
    testutil.CreateProjectFiles(t, workspace, "go")
}
```

### Custom Assertions
```go
import "github.com/daddia/zen/test/testutil"

func TestZenOutput(t *testing.T) {
    output := runZenCommand()
    
    // Check Zen design compliance
    testutil.AssertZenDesignCompliance(t, output)
    
    // Check JSON validity
    testutil.AssertValidJSON(t, output)
}
```

## Test Data Management

### Shared Fixtures
Place common test data in `test/fixtures/`:
```
test/fixtures/
├── configs/           # Configuration files
│   ├── minimal.yaml   # Basic config
│   └── complete.yaml  # Full config with all options
├── golden/            # Expected output files
│   ├── status_ready.txt
│   └── version_json.json
└── workspaces/        # Sample workspace structures
    ├── go-project/
    └── node-project/
```

### Package-Specific Data
Use `testdata/` directories in packages:
```
pkg/cmd/status/
├── status.go
├── status_test.go
└── testdata/
    ├── config.yaml    # Status-specific test config
    └── output.json    # Expected status output
```

## Test Categories by File Location

### Choose the Right Location

| Test Type | Location | When to Use |
|-----------|----------|-------------|
| **Unit** | `pkg/*/` | Testing individual functions, mocking dependencies |
| **Integration** | `test/integration/` | Testing component interactions, real dependencies |
| **E2E** | `test/e2e/` | Testing complete user workflows, CLI behavior |

### Examples

```go
// Unit Test - pkg/cmd/status/status_test.go
func TestFormatStatus(t *testing.T) {
    // Test individual function with mocked dependencies
}

// Integration Test - test/integration/config_integration_test.go
func TestConfigWithRealFiles(t *testing.T) {
    // Test config loading with real file system
}

// E2E Test - test/e2e/status_workflow_test.go
func TestE2E_StatusWorkflow(t *testing.T) {
    // Test complete zen status command workflow
}
```

## Performance Guidelines

### Test Execution Targets
- **Unit Tests**: <30 seconds total
- **Integration Tests**: <60 seconds total  
- **E2E Tests**: <300 seconds total

### Optimization Tips
1. **Use `t.Parallel()`** for independent tests
2. **Mock expensive operations** in unit tests
3. **Reuse test environments** where possible
4. **Use `testing.Short()`** for quick test runs

## Quality Standards

### Coverage Requirements
- **Overall**: ≥80% code coverage
- **Business Logic**: ≥90% coverage
- **Critical Paths**: 100% coverage (error handling, security)

### Test Quality
- **Independence**: Tests can run in any order
- **Cleanup**: No test artifacts left behind
- **Descriptive**: Clear test names and error messages
- **Comprehensive**: Happy path, edge cases, and error conditions

## Common Patterns

### Table-Driven Tests
```go
func TestFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    InputType
        expected OutputType
        wantErr  bool
    }{
        {"happy_path", validInput, expectedOutput, false},
        {"error_case", invalidInput, nil, true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### CLI Command Testing
```go
func TestCLICommand(t *testing.T) {
    streams := iostreams.Test()
    factory := cmdutil.NewTestFactory(streams)
    
    cmd := NewCommand(factory)
    cmd.SetArgs([]string{"--flag", "value"})
    
    err := cmd.Execute()
    require.NoError(t, err)
    
    output := streams.Out.String()
    assert.Contains(t, output, "expected")
}
```

### Error Testing
```go
func TestErrorHandling(t *testing.T) {
    result, err := functionThatShouldFail()
    
    require.Error(t, err)
    assert.Contains(t, err.Error(), "expected error message")
    assert.Nil(t, result)
}
```

## Debugging Tests

### Useful Commands
```bash
# Run specific test with verbose output
go test -v -run TestSpecificFunction ./pkg/cmd/status

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -cover ./...

# Update golden files
go test -update-golden ./...

# Run only fast tests
go test -short ./...
```

### Test Debugging Tips
1. **Use `t.Logf()`** for debug output
2. **Check test artifacts** in temp directories
3. **Run tests individually** to isolate issues
4. **Use `t.FailNow()`** to stop on first failure

## Maintenance

### Regular Tasks
1. **Update golden files** when output format changes
2. **Review test coverage** monthly
3. **Clean up unused fixtures** and test data
4. **Update test documentation** when adding new patterns

### When Refactoring
1. **Update affected tests** immediately
2. **Maintain test coverage** during refactoring
3. **Update shared utilities** if interfaces change
4. **Run full test suite** before merging changes

For more information, see:
- [Testing Guide](../docs/contributing/testing.md)
- [Test Standards](../.cursor/rules/test-standards.mdc)
- [Unit Test Standards](../.cursor/rules/test-unit.mdc)
