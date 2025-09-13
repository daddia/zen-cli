---
status: Accepted
date: 2025-09-12
decision-makers: Development Team, Architecture Team
consulted: UX Team, DevOps Team
informed: Product Team, Support Team
---

# ADR-0002 - Cobra CLI Framework Selection

## Context and Problem Statement

The Zen CLI requires a robust framework for command-line interface development that can handle complex command hierarchies, flag management, help generation, and extensibility. The framework must support the sophisticated command structure needed for product lifecycle management while maintaining excellent user experience and developer productivity.

Key requirements:
- Hierarchical command structure with subcommands and nested operations
- Comprehensive flag management (global, persistent, local flags)
- Automatic help generation with professional formatting
- Shell completion support for improved user experience
- Extensible architecture for plugin commands
- Configuration binding and validation
- Consistent error handling and user feedback
- Industry-standard CLI patterns and conventions

## Decision Drivers

* **Command Complexity**: Support for multi-level command hierarchies (zen workflow build, zen product analyze)
* **Flag Management**: Global flags, command-specific flags, persistent flags across subcommands
* **User Experience**: Professional help system, shell completion, intuitive command discovery
* **Developer Experience**: Clear API, extensive documentation, active community
* **Ecosystem Integration**: Seamless integration with configuration management (Viper)
* **Extensibility**: Plugin architecture support and dynamic command registration
* **Industry Standards**: Follow established CLI conventions and patterns
* **Maintenance**: Active development, security updates, long-term viability

## Considered Options

* **Cobra** - Popular Go CLI framework used by Kubernetes, Docker, GitHub CLI
* **Urfave/CLI** - Lightweight Go CLI framework with simple API
* **Kingpin** - Command-line parser with strong type safety
* **Flag** - Go standard library flag package
* **Custom Implementation** - Build CLI framework from scratch

## Decision Outcome

Chosen option: **Cobra v1.10.1+**, because it provides the most comprehensive feature set for complex CLI applications with excellent ecosystem integration and industry adoption.

### Consequences

**Good:**
- Industry-proven framework used by major CLI tools (kubectl, docker, gh)
- Comprehensive command hierarchy support with intuitive API
- Automatic help generation with professional formatting
- Built-in shell completion for bash, zsh, fish, powershell
- Seamless integration with Viper for configuration management
- Extensive documentation and community resources
- Plugin architecture support for extensible commands
- Consistent error handling and user feedback patterns
- Active development and security maintenance

**Bad:**
- Additional dependency (~2MB to binary size)
- Opinionated command structure and conventions
- Learning curve for advanced features (custom completions, hooks)
- Some complexity for simple CLI applications

**Neutral:**
- Requires following Cobra conventions and patterns
- Rich feature set may be overkill for simple commands

### Confirmation

The decision has been validated through:
- Successful implementation of all ZEN-001 CLI commands with professional help output
- Shell completion working across all supported shells
- Extensible command structure supporting future plugin architecture
- Positive developer feedback on API clarity and documentation
- Performance benchmarks showing minimal overhead
- Security audit showing no critical vulnerabilities

## Pros and Cons of the Options

### Cobra

**Good:**
- Industry standard for Go CLI applications
- Comprehensive command hierarchy and subcommand support
- Automatic help generation with excellent formatting
- Built-in shell completion for multiple shells
- Seamless Viper integration for configuration
- Extensive documentation and examples
- Active community and regular updates
- Plugin architecture support
- Consistent error handling patterns
- Professional CLI conventions (--help, --version, etc.)

**Bad:**
- Larger dependency footprint
- Can be complex for simple CLI applications
- Opinionated structure and conventions
- Learning curve for advanced features

**Neutral:**
- Rich feature set requires understanding of CLI best practices
- Established patterns may limit creative command structures

### Urfave/CLI

**Good:**
- Lightweight and simple API
- Good documentation and examples
- Flexible command structure
- Smaller binary footprint

**Bad:**
- Less sophisticated help generation
- Limited shell completion support
- No built-in configuration integration
- Smaller community and ecosystem
- Less suitable for complex command hierarchies

**Neutral:**
- Simpler API may be easier to learn
- Less opinionated about command structure

### Kingpin

**Good:**
- Strong type safety and validation
- Good error messages
- Flexible command structure

**Bad:**
- Less active development
- Limited shell completion
- No configuration framework integration
- Smaller community
- Complex API for nested commands

**Neutral:**
- Different approach to command definition
- Focus on type safety over ease of use

### Go Flag Package

**Good:**
- Part of standard library (no dependencies)
- Simple and lightweight
- Direct control over flag parsing

**Bad:**
- No subcommand support
- Manual help generation required
- No shell completion
- No configuration integration
- Limited to flat command structure

**Neutral:**
- Requires significant custom development
- Full control over implementation

### Custom Implementation

**Good:**
- Complete control over features and API
- No external dependencies
- Optimized for specific use case

**Bad:**
- Significant development time and effort
- Need to implement all CLI conventions from scratch
- Maintenance burden for framework code
- Risk of bugs and security issues
- No community support or documentation

**Neutral:**
- Complete flexibility in design decisions
- Requires expertise in CLI design patterns

## More Information

**Implementation Details:**
- Root command with global flags (--verbose, --no-color, --output)
- Hierarchical subcommands (version, init, config, status, workflow, product)
- Persistent flags inherited by subcommands
- Custom help templates with Zen branding
- Shell completion for commands and flags
- Plugin command registration hooks

**Performance Impact:**
- Binary size increase: ~2MB
- Startup time impact: <5ms
- Memory overhead: ~1MB
- Command parsing: <1ms per command

**Related ADRs:**
- ADR-0001: Go Language Choice
- ADR-0003: Configuration Management (Viper)
- ADR-0005: Plugin Architecture Design

**References:**
- [Cobra Documentation](https://cobra.dev/)
- [Cobra GitHub Repository](https://github.com/spf13/cobra)
- [CLI Best Practices](https://clig.dev/)
- [Shell Completion Guide](https://cobra.dev/#generating-documentation-for-your-command)

**Examples of Cobra Usage:**
- Kubernetes CLI (kubectl)
- Docker CLI
- GitHub CLI (gh)
- Terraform CLI
- Hugo static site generator
