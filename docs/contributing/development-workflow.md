# Development Workflow

This guide covers the day-to-day development process for contributing to Zen.

## Workflow Overview

1. **Plan** - Understand what you're building
2. **Develop** - Write code following our standards
3. **Test** - Ensure quality and reliability
4. **Document** - Keep documentation current
5. **Submit** - Create pull request for review
6. **Iterate** - Address feedback and improve

## Branch Strategy

### Branch Naming

Use descriptive branch names with prefixes:

- `feat/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation updates
- `refactor/` - Code refactoring
- `test/` - Test additions or fixes
- `perf/` - Performance improvements
- `ci/` - CI/CD changes
- `chore/` - Maintenance tasks

Examples:
- `feat/add-claude-provider`
- `fix/config-validation-error`
- `docs/improve-testing-guide`

### Branch Management

```bash
# Start from updated main
git checkout main
git pull upstream main

# Create feature branch
git checkout -b feat/new-feature

# Keep branch updated
git fetch upstream
git rebase upstream/main
```

## Writing Code

### Code Standards

#### Go Conventions
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use meaningful variable names
- Keep functions small and focused
- Handle errors explicitly
- Add comments for exported functions

#### Project Patterns
- Use factory pattern for object creation
- Implement interfaces for testability
- Follow existing package structure
- Use dependency injection

### File Organization

```go
// Package comment describes the package purpose
package cmd

import (
    // Standard library first
    "context"
    "fmt"
    
    // Third party packages
    "github.com/spf13/cobra"
    
    // Internal packages last
    "github.com/zen/internal/config"
)

// Exported types and functions with comments
type Command struct {
    // fields
}

// NewCommand creates a new command instance
func NewCommand() *Command {
    // implementation
}

// private functions follow
func helperFunction() {
    // implementation
}
```

### Error Handling

```go
// Wrap errors with context
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}

// Use sentinel errors for known conditions
var ErrNotFound = errors.New("resource not found")

// Check specific errors
if errors.Is(err, ErrNotFound) {
    // handle not found case
}
```

## Making Changes

### Before Coding

1. **Understand the requirement**
   - Read issue thoroughly
   - Ask questions if unclear
   - Check acceptance criteria

2. **Review existing code**
   - Find similar patterns
   - Understand dependencies
   - Check for reusable components

3. **Plan your approach**
   - Break into small commits
   - Consider test strategy
   - Think about documentation

### During Development

#### Commit Guidelines

Follow Conventional Commits format:

```bash
# Format: <type>(<scope>): <subject>

feat(cli): add progress indicator for long operations
fix(config): resolve path resolution on Windows
docs(api): update authentication examples
test(workspace): add coverage for edge cases
```

#### Commit Best Practices

- Make atomic commits (one logical change)
- Write clear commit messages
- Reference issues in commits (`Fixes #123`)
- Sign commits if configured

```bash
# Stage specific files
git add -p

# Commit with message
git commit -m "fix(auth): handle token expiration correctly

- Add retry logic for expired tokens
- Update error messages for clarity
- Add test coverage for edge cases

Fixes #456"
```

### Code Quality Checks

Run these before committing:

```bash
# Format code
make fmt

# Run linter
make lint

# Run unit tests
make test-unit

# Check everything
make verify
```

## Testing Your Changes

### Test Requirements

- Write tests for new functionality
- Update tests for changed behavior
- Maintain or improve coverage
- Include edge cases

### Running Tests

```bash
# Quick unit tests
go test ./pkg/...

# With coverage
go test -cover ./...

# Specific package
go test -v ./pkg/cmd/init

# Integration tests
make test-integration

# Full test suite
make test-all
```

See [Testing Guide](testing.md) for detailed testing practices.

## Updating Documentation

### When to Update Docs

- New features or commands
- Changed behavior
- New configuration options
- Deprecated functionality
- Better examples needed

### Documentation Types

1. **Code Comments** - In-source documentation
2. **README Files** - Package and directory docs
3. **User Guides** - End-user documentation
4. **API Docs** - Generated from code
5. **ADRs** - Architecture decisions

### Generate Documentation

```bash
# Generate CLI docs
make docs

# Generate API docs
make api-docs

# Validate links
make docs-check
```

## Preparing for Review

### Pre-Submission Checklist

- [ ] Code follows style guidelines
- [ ] Tests pass locally
- [ ] Documentation updated
- [ ] Commits are clean and logical
- [ ] Branch is up to date with main
- [ ] No commented-out code
- [ ] No debug statements
- [ ] Sensitive data removed

### Final Verification

```bash
# Run complete check
make verify

# Check for common issues
go vet ./...
staticcheck ./...
golangci-lint run

# Ensure clean git status
git status
```

## Troubleshooting

### Common Issues

#### Merge Conflicts
```bash
# Update your branch
git fetch upstream
git rebase upstream/main

# Resolve conflicts manually
# Then continue
git rebase --continue
```

#### Failed Tests
```bash
# Run specific test with verbose output
go test -v -run TestSpecificFunction ./pkg/...

# Check test coverage gaps
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Linter Errors
```bash
# Auto-fix where possible
golangci-lint run --fix

# Check specific linter
golangci-lint run --disable-all --enable errcheck
```

## Best Practices

### Do's
- Keep PRs focused and small
- Write descriptive commit messages
- Test edge cases
- Update documentation
- Respond to feedback promptly
- Ask for help when stuck

### Don'ts
- Mix unrelated changes
- Ignore CI failures
- Skip tests
- Leave TODOs without issues
- Force push to shared branches
- Commit sensitive data

## Next Steps

- Submit your code for [Code Review](code-review.md)
- Learn about our [Testing](testing.md) practices
- Understand [Release Process](release-process.md)

---

Need help? Check our [FAQ](README.md#getting-help) or ask in discussions.
