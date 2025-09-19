---
status: Accepted
date: 2025-09-16
decision-makers: Development Team, Architecture Team
consulted: Documentation Team, DevOps Team
informed: Product Team, Engineering Leadership
---

# ADR-0022 - Automatic Documentation Generation

## Context

The Zen CLI requires comprehensive, accurate, and up-to-date command-line documentation. Manual documentation maintenance is error-prone, time-consuming, and often falls out of sync with the actual command implementations. 

Following ADR-0021 (Cobra CLI Framework Maximization), we need to leverage Cobra's built-in documentation generation capabilities to ensure documentation accuracy and reduce maintenance burden while optimizing for modern consumption patterns including LLM ingestion.

## Decision Drivers

* **Documentation Accuracy**: Ensure documentation always reflects actual command behavior
* **Developer Productivity**: Eliminate manual documentation updates for command changes
* **LLM Optimization**: Structure documentation for effective AI tool consumption
* **Maintenance Efficiency**: Reduce documentation maintenance burden
* **Multiple Output Formats**: Support various documentation systems and platforms
* **CI/CD Integration**: Enable automated documentation validation in pipelines

## Decision Outcome

Implement automatic documentation generation using Cobra's built-in `github.com/spf13/cobra/doc` package, generating LLM-optimized Markdown with optional Man pages and ReStructuredText formats.

Implementation includes:
* Custom documentation generator tool at `internal/tools/docgen/main.go`
* Markdown generation with YAML front matter for static site generators
* One file per command for optimal chunking and search
* Rich examples automatically extracted from command definitions
* Stable, reproducible output without timestamps (unless requested)
* Integration with Makefile and CI/CD workflows

### Consequences

**Good:**
- Documentation always stays in sync with command implementations
- Zero manual effort for basic documentation updates
- Consistent structure across all command documentation
- LLM-optimized format with clear sections and examples
- Multiple output formats from single source of truth
- Reduced documentation errors and inconsistencies
- Faster release cycles with automated documentation

**Bad:**
- Requires discipline in writing good command descriptions and examples
- Limited customization compared to fully manual documentation
- Generated documentation may lack nuanced explanations
- Additional build dependency (cobra/doc package)

### Confirmation

Documentation generation successfully implemented and tested:
- All commands generate proper Markdown with front matter
- Examples properly extracted and formatted
- Multiple format support validated (Markdown, Man, ReST)
- CI/CD integration ready with `make docs-check`
- LLM-friendly structure with one command per file

## Implementation Details

### Documentation Generator (`internal/tools/docgen/main.go`)

```go
// Key features:
- Configurable output directory (default: ./docs/zen)
- Multiple format support (markdown, man, rest)
- Optional YAML front matter for static sites
- Stable output without timestamps (reproducible builds)
- Enhanced command preparation for better examples
```

### Makefile Integration

```makefile
docs: docs-markdown              # Default to Markdown generation
docs-markdown:                   # Generate Markdown with front matter
docs-man:                        # Generate Man pages
docs-rest:                      # Generate ReStructuredText
docs-all:                       # Generate all formats
docs-check:                     # Verify docs are up-to-date
docs-clean:                     # Remove generated documentation
```

### File Naming Convention

- Auto-generated files: `zen_*.md` (underscores)
- Manual documentation: `zen-*.md` (hyphens)
- This allows coexistence of both documentation types

### Front Matter Structure

```yaml
---
title: "command name"
slug: "/cli/command-name"
description: "CLI reference for command"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-16
keywords:
  - zen
  - cli
  - command
---
```

## Best Practices

### Command Definition Requirements

To generate high-quality documentation, commands must include:

1. **Clear `Short` description** - One-line command purpose
2. **Comprehensive `Long` description** - Detailed explanation with context
3. **Rich `Example` field** - Multiple realistic usage examples
4. **Complete flag descriptions** - Purpose and valid values for each flag

### Example Enhancement

```go
cmd := &cobra.Command{
    Use:   "config",
    Short: "Manage configuration for Zen CLI",
    Long: `Detailed description explaining the command's purpose,
how it works, and when to use it.`,
    Example: `  # Display current configuration
  zen config
  
  # Set a configuration value
  zen config set log_level debug
  
  # Get a specific value
  zen config get log_level`,
}
```

### CI/CD Integration

```yaml
# GitHub Actions example
- name: Check documentation
  run: make docs-check
```

## More Information

- Related ADRs: [ADR-0021](ADR-0021-cobra-maximization.md) (Cobra Maximization), [ADR-0002](ADR-0002-cli-framework.md) (CLI Framework)
- Cobra Documentation Guide: https://cobra.dev/docs/how-to-guides/clis-for-llms/
- Implementation: `internal/tools/docgen/main.go`
- Generated Documentation: `docs/zen/`
- Follow-ups: Integrate with documentation site, add to release process
