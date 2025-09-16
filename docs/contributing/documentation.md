# Documentation Standards

Clear documentation is essential for Zen's success. This guide covers how to write and maintain documentation.

## Documentation Types

### Code Documentation

#### Package Comments
Every package needs a clear description:

```go
// Package config provides configuration management for Zen.
// It handles loading, validation, and merging of configuration
// from multiple sources including files, environment variables,
// and command-line flags.
package config
```

#### Function Comments
Export functions require documentation:

```go
// LoadConfig reads configuration from the specified path
// and returns a validated Config instance.
// It returns an error if the file cannot be read or parsed.
func LoadConfig(path string) (*Config, error) {
    // implementation
}
```

#### Inline Comments
Use inline comments for complex logic:

```go
// Calculate weighted score using RICE framework
// (Reach * Impact * Confidence) / Effort
score := (r.Reach * r.Impact * r.Confidence) / r.Effort
```

### User Documentation

#### Command Documentation
Each command needs clear help text:

```go
cmd := &cobra.Command{
    Use:   "init [path]",
    Short: "Initialize a new Zen project",
    Long: `Initialize creates a new Zen project in the specified directory.
    
It sets up the project structure, configuration files, and 
initializes version control integration.`,
    Example: `  # Initialize in current directory
  zen init
  
  # Initialize in specific directory
  zen init ./my-project`,
}
```

#### README Files
Each major directory needs a README explaining its purpose:

```markdown
# Package Name

Brief description of what this package does.

## Usage

How to use this package.

## Examples

Common usage examples.
```

### API Documentation

#### OpenAPI/Swagger
Document REST APIs using OpenAPI annotations:

```go
// @Summary Get project status
// @Description Returns the current status of the project
// @Tags project
// @Accept json
// @Produce json
// @Success 200 {object} Status
// @Router /status [get]
func GetStatus(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

## Writing Style

### General Guidelines

- Use clear, simple language
- Write in active voice
- Keep sentences short
- Define acronyms on first use
- Use consistent terminology

### Technical Writing

#### Good Example
"The configuration loader reads settings from multiple sources in order of precedence."

#### Poor Example
"Settings are read by the configuration loader from sources with precedence."

### Code Examples

Include working examples that users can copy:

```markdown
## Configuration Example

Create a `zen.yaml` file:

\```yaml
project:
  name: my-app
  version: 1.0.0
settings:
  log_level: debug
\```

Load the configuration:

\```bash
zen init --config zen.yaml
\```
```

## Markdown Standards

### Headings

Use hierarchical heading structure:

```markdown
# Page Title

## Main Section

### Subsection

#### Detail Point
```

### Lists

Use lists for clarity:

```markdown
Requirements:
- Go 1.25 or higher
- Git 2.30 or higher
- 4GB RAM minimum

Steps:
1. Install prerequisites
2. Clone repository
3. Run setup script
```

### Code Blocks

Always specify language for syntax highlighting:

````markdown
```go
func main() {
    fmt.Println("Hello, Zen!")
}
```

```bash
make build
./bin/zen --help
```
````

### Tables

Use tables for structured data:

```markdown
| Command | Description | Example |
|---------|-------------|---------|
| init    | Initialize project | `zen init` |
| status  | Show status | `zen status` |
```

## Documentation Generation

### CLI Documentation

Auto-generate CLI docs from Cobra commands:

```bash
# Generate markdown documentation
make docs

# Generate man pages
make docs-man

# Check documentation is current
make docs-check
```

### API Documentation

Generate API documentation:

```bash
# Generate OpenAPI spec
make api-docs

# Serve documentation locally
make docs-serve
```

## Architecture Decision Records (ADRs)

Document significant decisions using ADRs:

```markdown
# ADR-NNNN: Title

## Status
Accepted/Rejected/Deprecated/Superseded

## Context
What is the issue we're facing?

## Decision
What have we decided to do?

## Consequences
What are the results of this decision?
```

Create new ADR:

```bash
# Copy template
cp docs/architecture/decisions/adr-template.md \
   docs/architecture/decisions/ADR-NNNN-your-title.md

# Edit and fill in details
```

## Documentation Maintenance

### When to Update

Update documentation when you:
- Add new features
- Change existing behavior
- Fix bugs that affect usage
- Deprecate functionality
- Improve examples

### Documentation Review

Before submitting PR, verify:
- [ ] Code comments are accurate
- [ ] README files are updated
- [ ] CLI help text is clear
- [ ] Examples work as shown
- [ ] Links are valid
- [ ] Formatting is consistent

### Common Issues

#### Outdated Examples
Test all examples in documentation:

```bash
# Extract and test code examples
make test-docs
```

#### Broken Links
Check for broken links:

```bash
# Validate all documentation links
make docs-links-check
```

#### Missing Documentation
Ensure all exports are documented:

```bash
# Check for missing comments
golint ./...
```

## Documentation Tools

### Linters

Use documentation linters:

```bash
# Markdown linter
markdownlint docs/

# Vale for prose
vale docs/
```

### Spell Checking

Check spelling:

```bash
# Run spell checker
make spell-check

# Add words to dictionary
echo "newword" >> .spelling
```

## Best Practices

### Do's

- Document the "why" not just "what"
- Include examples for complex features
- Keep documentation close to code
- Update docs with code changes
- Use diagrams for complex concepts
- Test documentation examples

### Don'ts

- Don't document obvious things
- Don't duplicate information
- Don't use jargon without explanation
- Don't leave TODOs in documentation
- Don't commit generated documentation

## Documentation Templates

### Feature Documentation

```markdown
# Feature Name

## Overview
Brief description of the feature.

## Use Cases
- Use case 1
- Use case 2

## Configuration
How to enable and configure.

## Examples
Working examples.

## Troubleshooting
Common issues and solutions.
```

### Troubleshooting Guide

```markdown
# Troubleshooting: [Topic]

## Symptom
What the user observes.

## Cause
Why this happens.

## Solution
How to fix it.

## Prevention
How to avoid it.
```

## Next Steps

- Review [Architecture](architecture.md) for system design
- Understand [Code Review](code-review.md) process
- Learn about [Release Process](release-process.md)

---

Questions about documentation? Check existing docs or ask maintainers.
