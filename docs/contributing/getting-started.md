# Getting Started

This guide helps you set up your development environment for contributing to Zen.

## Prerequisites

### Required Tools
- **Go 1.25+** - Primary development language
- **Git 2.30+** - Version control
- **Make** - Build automation (optional but recommended)
- **Docker** - For containerized testing (optional)

### Recommended IDE Setup
- **VS Code** with Go extension
- **GoLand** for JetBrains users
- **Neovim** with gopls for terminal enthusiasts

## Initial Setup

### 1. Fork and Clone

```bash
# Fork the repository on GitHub first, then:
git clone https://github.com/YOUR_USERNAME/zen.git
cd zen
git remote add upstream https://github.com/zen-org/zen.git
```

### 2. Install Dependencies

```bash
# Install Go dependencies
go mod download

# Verify installation
go mod verify
```

### 3. Environment Configuration

```bash
# Copy example environment file
cp .env.example .env

# Edit .env with your settings
# Required: LLM provider API keys if testing AI features
```

### 4. Build the Project

```bash
# Using Make (recommended)
make build

# Or directly with Go
go build -o bin/zen ./cmd/zen

# Verify the build
./bin/zen version
```

## Development Environment

### Directory Structure

```
zen/
├── cmd/           # Entry points
├── pkg/           # Public packages
├── internal/      # Private packages
├── docs/          # Documentation
├── test/          # Test suites
└── scripts/       # Build scripts
```

### Key Configuration Files

- `.env` - Local environment variables
- `go.mod` - Go module dependencies
- `Makefile` - Build and test commands
- `.golangci.yml` - Linter configuration

## Verify Your Setup

### Run Basic Commands

```bash
# Run unit tests
make test-unit

# Run linter
make lint

# Generate documentation
make docs

# Full verification
make verify
```

### Common Setup Issues

#### Go Module Errors
```bash
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
```

#### Build Failures
```bash
# Ensure Go version is correct
go version

# Update dependencies
go mod tidy
```

#### Permission Issues
```bash
# Make scripts executable
chmod +x scripts/*.sh
```

## IDE Configuration

### VS Code Settings

Create `.vscode/settings.json`:

```json
{
  "go.lintTool": "golangci-lint",
  "go.lintOnSave": "package",
  "go.formatTool": "goimports",
  "go.useLanguageServer": true
}
```

### Git Configuration

```bash
# Set up commit signing (optional but recommended)
git config user.name "Your Name"
git config user.email "your.email@example.com"

# Enable pre-commit hooks
git config core.hooksPath .githooks
```

## Quick Start Development

### Your First Contribution

1. **Find an Issue**
   - Look for issues labeled `good-first-issue`
   - Check no one is already assigned

2. **Create a Branch**
   ```bash
   git checkout -b fix/issue-description
   ```

3. **Make Changes**
   - Write your code
   - Add tests
   - Update documentation

4. **Verify Changes**
   ```bash
   make test-unit
   make lint
   ```

5. **Commit and Push**
   ```bash
   git add .
   git commit -m "fix: clear description of change"
   git push origin fix/issue-description
   ```

6. **Open Pull Request**
   - Use the PR template
   - Link the related issue
   - Wait for CI checks

## Development Tools

### Makefile Commands

```bash
make help         # Show all commands
make build        # Build binary
make test         # Run all tests
make lint         # Run linters
make fmt          # Format code
make docs         # Generate docs
make clean        # Clean build artifacts
```

### Debugging

```bash
# Run with debug logging
ZEN_LOG_LEVEL=debug ./bin/zen status

# Use delve debugger
dlv debug ./cmd/zen -- status
```

### Performance Profiling

```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

## Next Steps

- Read the [Development Workflow](development-workflow.md) guide
- Understand our [Architecture](architecture.md)
- Learn about [Testing](testing.md) practices
- Review [Code Review](code-review.md) standards

## Getting Help

- Check existing documentation first
- Search closed issues for similar problems
- Open a discussion for setup help
- Tag `@maintainers` for urgent blockers

---

Ready to code? Continue with the [Development Workflow](development-workflow.md).
