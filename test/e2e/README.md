# End-to-End (E2E) Tests

This directory contains end-to-end tests for the Zen CLI that test the complete user experience from command execution to output verification.

## Test Structure

The e2e tests are organized into separate files based on functionality:

### Core Test Files

- **`setup.go`** - Test environment setup and teardown utilities
- **`helpers.go`** - Common test helper functions and utilities
- **`main_test.go`** - Test main function for e2e test coordination

### Test Categories

#### 1. Core Commands (`core_commands_test.go`)
Tests basic zen commands that work without workspace initialization:
- `zen status` (not initialized behavior)
- `zen help` / `zen --help`
- `zen version` (text, JSON, YAML formats)
- Global flags (`--verbose`, `--no-color`, etc.)
- Error handling (invalid commands, invalid flags)

#### 2. Workspace Operations (`workspace_test.go`)
Tests zen workspace initialization and configuration:
- `zen status` (before and after initialization)
- `zen init` (first time, idempotent, force)
- `zen config` (list, get, set operations)
- Project type detection (Go, Node.js, Python, Git)
- Configuration validation

#### 3. Assets Management (`assets_test.go`)
Tests zen assets commands:
- `zen assets status` (text, JSON, YAML formats)
- `zen assets sync` (normal and dry-run modes)
- `zen assets list` (with various filters)
- Authentication and error scenarios
- Complete assets workflow

#### 4. User Journeys (`user_journeys_test.go`)
Tests complete user workflows:
- Critical path: init → config → status
- Force initialization scenarios
- Output format testing
- Global flags functionality
- Error scenarios

## Test Environment

### Setup and Teardown

Each test uses the new setup/teardown system:

```go
func TestExample(t *testing.T) {
    env := SetupTestEnvironment(t)
    defer TeardownTestEnvironment(t, env)
    
    workspaceDir := env.CreateTestWorkspace(t, "test-name")
    defer env.CleanupTestWorkspace(t, workspaceDir)
    
    // Test implementation
}
```

### Test Directory Structure

Tests create a `zen-test` directory outside of the `zen-cli` project:

```
/Users/jonathandaddia/Projects/
├── zen-cli/                    # Main project
└── zen-test/                   # E2E test workspace (created during tests)
    ├── core-commands-test/     # Individual test workspaces
    ├── workspace-init-test/
    ├── assets-test/
    └── ...
```

### Binary Management

- Each test environment builds its own zen binary
- Binaries are created in temporary directories
- Automatic cleanup after test completion
- Isolated from development builds

## Running E2E Tests

### Prerequisites

1. Go 1.25+ installed
2. Zen CLI project built successfully
3. Network access (for assets sync tests)

### Running Tests

```bash
# Run all e2e tests
go test -tags=e2e ./test/e2e/

# Run specific test category
go test -tags=e2e ./test/e2e/ -run TestE2E_CoreCommands
go test -tags=e2e ./test/e2e/ -run TestE2E_WorkspaceInitialization
go test -tags=e2e ./test/e2e/ -run TestE2E_AssetsCommands

# Run with verbose output
go test -tags=e2e -v ./test/e2e/

# Run specific test
go test -tags=e2e ./test/e2e/ -run TestE2E_CoreCommands/zen_status_not_initialized
```

### Test Output

Tests provide detailed logging:
- Command execution details
- Exit codes and output
- Environment setup information
- Cleanup status

Example output:
```
=== RUN   TestE2E_CoreCommands/zen_status_not_initialized
    setup.go:45: Test environment setup complete:
    setup.go:46:   Project Root: /Users/.../zen-cli
    setup.go:47:   Test Root: /Users/.../zen-test
    setup.go:48:   Zen Binary: /tmp/zen-e2e-binary-123/zen-e2e
    helpers.go:35: Command: zen [status]
    helpers.go:36: Exit Code: 1
    helpers.go:40: Stderr: ✗ Not Initialized: Not a zen workspace (or any of the parent directories): .zen
--- PASS: TestE2E_CoreCommands/zen_status_not_initialized (0.02s)
```

## Test Design Principles

### 1. Isolation
- Each test creates its own workspace
- No shared state between tests
- Clean environment for each test run

### 2. Realistic Scenarios
- Tests use actual zen binary (not mocked)
- Real filesystem operations
- Authentic command-line experience

### 3. Comprehensive Coverage
- Happy path and error scenarios
- Different output formats
- Various command combinations
- Edge cases and boundary conditions

### 4. Maintainability
- Clear test organization
- Reusable helper functions
- Descriptive test names
- Comprehensive assertions

### 5. Performance
- Parallel test execution where possible
- Efficient setup/teardown
- Minimal test duration
- Resource cleanup

## Adding New Tests

### 1. Choose the Right File
- **Core commands**: Add to `core_commands_test.go`
- **Workspace operations**: Add to `workspace_test.go`
- **Assets functionality**: Add to `assets_test.go`
- **Complete workflows**: Add to `user_journeys_test.go`

### 2. Follow the Pattern

```go
func TestE2E_NewFeature(t *testing.T) {
    env := SetupTestEnvironment(t)
    defer TeardownTestEnvironment(t, env)
    
    workspaceDir := env.CreateTestWorkspace(t, "new-feature-test")
    defer env.CleanupTestWorkspace(t, workspaceDir)
    
    t.Run("specific_scenario", func(t *testing.T) {
        result := env.RunZenCommand(t, workspaceDir, "command", "args")
        result.RequireSuccess(t, "command should succeed")
        
        assert.Contains(t, result.Stdout, "expected output")
    })
}
```

### 3. Test Guidelines
- Use descriptive test and subtest names
- Test both success and failure scenarios
- Verify exit codes and output content
- Include edge cases and boundary conditions
- Add comments for complex test logic

## Troubleshooting

### Common Issues

1. **Binary build failures**: Check Go version and dependencies
2. **Permission errors**: Ensure write access to parent directory
3. **Timeout errors**: Increase timeout for slow operations
4. **Cleanup failures**: Check for file locks or permissions

### Debug Tips

1. Use `-v` flag for verbose output
2. Check test logs for command details
3. Manually inspect `zen-test` directory during failures
4. Run individual tests to isolate issues

### Environment Variables

- `ZEN_E2E_DEBUG=1`: Enable debug logging
- `ZEN_E2E_KEEP_DIRS=1`: Don't cleanup test directories
- `ZEN_E2E_TIMEOUT=60s`: Custom command timeout
