---
status: Accepted
date: 2025-09-13
decision-makers: Development Team, Architecture Team
consulted: Security Team, DevOps Team, UX Team
informed: Product Team, Support Team, Community Contributors
---

# ADR-0018 - Configuration Validation Framework

## Context and Problem Statement

The Zen CLI configuration system requires comprehensive validation to ensure configuration correctness, security, and user experience across multiple configuration sources (files, environment variables, CLI flags). The validation framework must provide clear error messages, support complex validation rules, and integrate seamlessly with the multi-source configuration loading system established in ADR-0004.

Key requirements:
- Schema validation for all configuration structures with type safety
- Cross-field validation and dependency checking
- Clear, actionable error messages with suggestions for fixes
- Performance-optimized validation for CLI startup speed
- Extensible validation rules for plugin configurations
- Security validation to prevent configuration-based attacks
- Development-friendly validation with helpful debugging information
- Support for configuration migration and backward compatibility

## Decision Drivers

* **User Experience**: Clear, actionable error messages that help users fix configuration issues
* **Security**: Prevent configuration-based security vulnerabilities
* **Performance**: Fast validation that doesn't impact CLI startup time
* **Maintainability**: Clean, extensible validation code that's easy to modify
* **Type Safety**: Compile-time and runtime type checking for configuration
* **Extensibility**: Support for plugin-specific validation rules
* **Development Experience**: Helpful validation during development and testing
* **Backward Compatibility**: Support for configuration migrations and deprecations

## Considered Options

* **Struct Tags with Custom Validator** - Go struct tags with custom validation library
* **JSON Schema Validation** - External JSON schema with validation library
* **Custom Validation Framework** - Purpose-built validation system for Zen
* **Viper Built-in Validation** - Use Viper's native validation capabilities
* **Interface-Based Validation** - Go interfaces with validation methods

## Decision Outcome

Chosen option: **Struct Tags with Custom Validator**, because it provides the best balance of type safety, performance, maintainability, and integration with the existing Viper-based configuration system while supporting complex validation scenarios.

### Consequences

**Good:**
- Compile-time type safety with runtime validation
- Excellent performance with minimal overhead
- Clear validation rules co-located with struct definitions
- Extensible validation rules for custom scenarios
- Seamless integration with existing Viper configuration loading
- Rich error messages with field-specific context
- Support for complex cross-field validation

**Bad:**
- Custom validation library dependency and maintenance
- Learning curve for developers unfamiliar with struct tag validation
- Validation rules embedded in code rather than external schema
- Limited runtime schema modification capabilities

**Neutral:**
- Validation logic tightly coupled with Go structs
- Requires careful design for validation rule extensibility
- Balance needed between validation completeness and performance

### Confirmation

The decision will be validated through:
- Configuration validation performance benchmarks showing <5ms validation time
- User experience testing with common configuration errors
- Developer experience validation with validation rule implementation
- Security testing with malicious configuration inputs
- Integration testing with multi-source configuration loading
- Plugin configuration validation extensibility demonstration

## Pros and Cons of the Options

### Struct Tags with Custom Validator

**Good:**
- Type safety at compile time and runtime
- Excellent performance with minimal overhead
- Validation rules co-located with struct definitions
- Rich error messages with field context
- Extensible validation rules and custom validators
- Seamless Viper integration
- Support for complex cross-field validation

**Bad:**
- Custom validation library maintenance overhead
- Validation rules embedded in code
- Learning curve for struct tag syntax
- Limited runtime schema modification

**Neutral:**
- Tight coupling between validation and struct definitions
- Go-specific validation approach
- Requires careful extensibility design

### JSON Schema Validation

**Good:**
- Industry-standard schema validation approach
- External schema files separate from code
- Rich validation capabilities and ecosystem
- Language-agnostic schema definitions
- Excellent tooling and documentation support

**Bad:**
- Performance overhead from JSON marshaling/unmarshaling
- Additional complexity with external schema files
- Less type safety in Go code
- Schema-struct synchronization challenges
- Limited integration with Viper's native types

**Neutral:**
- Requires schema file management and versioning
- May be overkill for CLI configuration validation
- External dependency on JSON schema library

### Custom Validation Framework

**Good:**
- Complete control over validation logic and performance
- Optimized for Zen's specific configuration needs
- No external dependencies for validation
- Perfect integration with existing architecture

**Bad:**
- Significant development and maintenance effort
- Need to implement validation patterns from scratch
- Risk of bugs in custom validation logic
- Less community support and testing
- Reinventing well-solved problems

**Neutral:**
- Full flexibility but high implementation cost
- Requires expertise in validation system design
- May be justified for highly specialized needs

### Viper Built-in Validation

**Good:**
- Native integration with existing configuration system
- No additional dependencies or complexity
- Familiar API for developers using Viper
- Maintained as part of Viper library

**Bad:**
- Limited validation capabilities
- Basic error messages without rich context
- No support for complex cross-field validation
- Limited extensibility for custom validation rules
- Performance characteristics tied to Viper internals

**Neutral:**
- Suitable only for simple validation scenarios
- May require supplementation with custom validation
- Dependent on Viper library development priorities

### Interface-Based Validation

**Good:**
- Go-idiomatic validation approach
- Type safety and compile-time checking
- Flexible validation method implementations
- No external dependencies required

**Bad:**
- Verbose validation code for each struct
- Limited validation rule reusability
- Manual error message construction
- No declarative validation syntax
- Complex cross-field validation implementation

**Neutral:**
- Requires significant boilerplate code
- May be suitable for simple validation scenarios
- Integration complexity with configuration loading

## More Information

**Validation Framework Components:**

**1. Core Validation Engine:**
```go
// Validation struct tags with custom validators
type Config struct {
    LogLevel  string `validate:"required,oneof=trace debug info warn error fatal panic" json:"log_level"`
    LogFormat string `validate:"required,oneof=text json" json:"log_format"`
    
    CLI CLIConfig `validate:"required" json:"cli"`
}

type CLIConfig struct {
    OutputFormat string `validate:"required,oneof=text json yaml" json:"output_format"`
    Verbose      bool   `json:"verbose"`
    NoColor      bool   `json:"no_color"`
}
```

**2. Custom Validators:**
- **Path Validator**: Validates file paths and prevents directory traversal
- **URL Validator**: Validates URLs with protocol and security checks  
- **Duration Validator**: Validates time durations with min/max constraints
- **Regex Validator**: Validates strings against regular expressions
- **Cross-Field Validator**: Validates dependencies between configuration fields

**3. Error Message Framework:**
```go
type ValidationError struct {
    Field   string `json:"field"`
    Value   string `json:"value"`
    Rule    string `json:"rule"`
    Message string `json:"message"`
    Suggestion string `json:"suggestion,omitempty"`
}

// Example error output:
// Error: Invalid configuration in log_level
// Value: "invalid"
// Rule: oneof=trace debug info warn error fatal panic
// Message: log_level must be one of: trace, debug, info, warn, error, fatal, panic
// Suggestion: Try setting log_level to "info" for general use
```

**4. Security Validation:**
- Input sanitization for all string fields
- Path traversal prevention for file paths
- URL scheme validation for external endpoints
- Regular expression DoS prevention
- Configuration size limits and depth restrictions

**5. Plugin Configuration Validation:**
```go
type PluginConfig struct {
    Name        string            `validate:"required,alphanum" json:"name"`
    Version     string            `validate:"required,semver" json:"version"`
    Permissions []string          `validate:"dive,oneof=filesystem network ai" json:"permissions"`
    Config      map[string]interface{} `validate:"max_depth=3" json:"config"`
}
```

**6. Configuration Migration Support:**
- Deprecated field warnings with migration suggestions
- Automatic value migration for renamed fields
- Version-specific validation rules
- Backward compatibility validation

**Validation Rules Library:**
- **required**: Field must be present and non-empty
- **oneof**: Value must be one of specified options
- **min/max**: Numeric range validation
- **alphanum**: Alphanumeric characters only
- **semver**: Semantic version format validation
- **path**: Valid file path with security checks
- **url**: Valid URL with protocol validation
- **duration**: Valid time duration format
- **dive**: Validate slice/map elements
- **cross_field**: Cross-field dependency validation

**Performance Optimizations:**
- Validation rule compilation and caching
- Early validation termination on first error (optional)
- Parallel validation for independent fields
- Validation result caching for repeated configurations

**Integration with Configuration Loading:**
```go
func Load() (*Config, error) {
    // Load configuration from multiple sources (Viper)
    config, err := loadFromSources()
    if err != nil {
        return nil, err
    }
    
    // Validate loaded configuration
    if err := validate(config); err != nil {
        return nil, &ConfigValidationError{
            Source: getConfigSource(),
            Errors: err.ValidationErrors,
        }
    }
    
    return config, nil
}
```

**Related ADRs:**
- ADR-0004: Configuration Management Strategy
- ADR-0015: Security Model Implementation
- ADR-0005: Structured Logging Implementation

**References:**
- [Go Validator Library](https://github.com/go-playground/validator)
- [JSON Schema Specification](https://json-schema.org/)
- [Configuration Validation Best Practices](https://12factor.net/config)
- [Input Validation Security Guide](https://owasp.org/www-project-proactive-controls/v3/en/c5-validate-inputs)

**Follow-ups:**
- Validation rule documentation and examples
- Configuration validation testing framework
- Migration tool for configuration upgrades
- Plugin configuration validation SDK
