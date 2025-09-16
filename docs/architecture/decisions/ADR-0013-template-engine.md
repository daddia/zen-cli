---
status: Accepted
date: 2025-09-16
decision-makers: Development Team, Architecture Team
consulted: Product Team, Technical Writing Team
informed: Engineering Leadership, UX Team
---

# ADR-0013 - Template Engine Design

## Context and Problem Statement

The Zen CLI platform requires a sophisticated template engine to generate dynamic content across multiple domains including user stories, ADRs, API contracts, code scaffolding, documentation, and workflow artifacts. The template system must support variable substitution, conditional logic, iteration, template inheritance, and integration with AI-generated content while maintaining high performance for CLI operations. The engine needs to handle diverse output formats (Markdown, YAML, JSON, code files) and enable template validation, testing, and extensible template registries.

## Decision Drivers

* **Output Diversity**: Support for multiple output formats including Markdown, YAML, JSON, Go, TypeScript, and custom formats
* **Template Complexity**: Conditional logic, loops, template inheritance, and composition for sophisticated content generation
* **Performance**: Fast template rendering suitable for interactive CLI usage with <100ms rendering times
* **AI Integration**: Seamless integration with AI-generated content and dynamic variable substitution
* **Extensibility**: Template registry system enabling custom templates and third-party template sharing
* **Validation**: Template syntax validation, output validation, and comprehensive error reporting
* **Maintainability**: Clear template syntax that is readable and maintainable by non-developers

## Considered Options

1. **Go Templates with Custom Extensions**
2. **Handlebars.js with Go Implementation** 
3. **Jinja2-Style Template Engine in Go**
4. **Custom DSL Template Language**

## Decision Outcome

Chosen option: "Go Templates with Custom Extensions", because it leverages the built-in Go template engine's proven performance and security while adding domain-specific extensions for Zen CLI's content generation needs. This approach minimizes external dependencies while providing the necessary features for complex template scenarios.

### Consequences

**Good:**
- Built-in Go templates provide excellent performance and memory efficiency for CLI usage
- Zero external dependencies align with single-binary distribution requirements
- Custom extensions enable domain-specific functionality for product and engineering workflows  
- Strong security model with automatic HTML escaping and injection protection

**Bad:**
- Go template syntax may be less familiar to non-developers compared to Handlebars or Jinja2
- Limited built-in functionality requires more custom extension development
- Less mature ecosystem compared to established template engines in other languages

### Confirmation

Performance benchmarks demonstrating <50ms rendering times for typical templates, template validation covering all custom extensions, security audit of template execution sandbox, and user experience testing with product managers creating custom templates.

## Pros and Cons of the Options

### Go Templates with Custom Extensions

Using Go's built-in text/template and html/template packages extended with custom functions specific to Zen CLI's content generation needs.

**Good:**
- Excellent performance with no external dependencies and built-in security features
- Natural integration with Go codebase and existing development workflows
- Extensible architecture allowing custom functions and domain-specific template features
- Strong security model with sandboxed execution and automatic escaping

**Neutral:**
- Moderate syntax complexity requiring documentation and training for template authors

**Bad:**
- Less expressive syntax compared to modern template engines like Jinja2 or Handlebars
- Limited built-in functionality requiring custom extension development for advanced features

### Handlebars.js with Go Implementation

Implementing a Handlebars-compatible template engine in Go, providing familiar JavaScript-style template syntax with helpers and partials support.

**Good:**
- Familiar syntax for developers experienced with JavaScript and web development
- Rich ecosystem of helpers and extensions available from JavaScript community
- Excellent support for template inheritance, partials, and composition patterns

**Bad:**
- Requires external dependency or Go implementation of Handlebars parser and runtime
- Performance overhead from JavaScript-compatible syntax parsing and evaluation
- Security concerns with JavaScript-style expression evaluation and helper execution

### Jinja2-Style Template Engine in Go

Building a template engine that mimics Jinja2's syntax and functionality, providing Python-style template features like filters, macros, and template inheritance.

**Good:**
- Very expressive template syntax with rich built-in functionality and filters
- Excellent support for template inheritance, macros, and advanced composition
- Familiar to developers with Python/Flask experience and Django background

**Bad:**
- Requires implementing complex template parser and runtime in Go from scratch
- Significant development and maintenance overhead for custom implementation
- Risk of incompatibility issues and performance problems compared to native Jinja2

### Custom DSL Template Language

Creating a domain-specific language tailored specifically for Zen CLI's content generation patterns, with syntax optimized for product and engineering workflows.

**Good:**
- Complete control over template syntax and functionality optimized for Zen CLI use cases
- Can design optimal integration points with AI content generation and workflow systems
- Minimal syntax tailored specifically for product and engineering content generation

**Bad:**
- Massive implementation effort requiring full parser, compiler, and runtime development
- No existing ecosystem or community knowledge about custom template language
- High risk of design mistakes and compatibility issues discovered after implementation

## More Information

- Related ADRs: [ADR-0009](ADR-0009-agent-orchestration.md), [ADR-0012](ADR-0012-integration-architecture.md)
- Implementation Location: `internal/templates/`, `pkg/templates/`
- Template Documentation: Go text/template package, template extension framework
- Follow-ups: Template registry and marketplace, AI-assisted template generation, template validation framework
