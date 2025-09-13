---
status: Accepted
date: 2025-09-12
decision-makers: Development Team, Architecture Team
consulted: DevOps Team, Security Team, UX Team
informed: Product Team, Support Team
---

# ADR-0004 - Configuration Management Strategy

## Context and Problem Statement

The Zen CLI requires a flexible and secure configuration management system that can handle multiple configuration sources, environment-specific settings, and user preferences while maintaining security and usability. The system must support development, testing, and production scenarios with appropriate defaults and validation.

Key requirements:
- Multiple configuration sources with clear precedence rules
- Environment-specific configuration profiles
- Secure handling of sensitive information (API keys, tokens)
- Schema validation with helpful error messages
- Hot-reload capability for development workflows
- Cross-platform configuration file locations
- Integration with CLI flags and environment variables
- Backward compatibility for configuration migrations

## Decision Drivers

* **Flexibility**: Support multiple configuration sources and formats
* **Security**: Secure handling of sensitive configuration data
* **User Experience**: Intuitive configuration with helpful defaults and error messages
* **Developer Experience**: Easy configuration management during development
* **Environment Support**: Different configurations for dev/test/prod environments
* **Integration**: Seamless integration with CLI framework (Cobra)
* **Validation**: Schema validation with clear error reporting
* **Maintainability**: Easy to extend and modify configuration structure

## Considered Options

* **Viper** - Popular Go configuration management library
* **Go-Config** - Lightweight configuration library
* **Custom YAML Parser** - Direct yaml.v3 usage with custom logic
* **Environment Variables Only** - Configuration through env vars exclusively
* **Configuration Files Only** - YAML/JSON files without environment integration

## Decision Outcome

Chosen option: **Viper v1.20.0+** with YAML primary format, because it provides comprehensive multi-source configuration management with excellent Cobra integration and robust validation capabilities.

### Consequences

**Good:**
- Comprehensive multi-source configuration with clear precedence
- Seamless integration with Cobra CLI framework
- Support for multiple formats (YAML, JSON, TOML, etc.)
- Automatic environment variable binding with prefix support
- Hot-reload capability for development workflows
- Cross-platform configuration file discovery
- Extensive validation and error reporting capabilities
- Active development and community support

**Bad:**
- Additional dependency (~3MB to binary size)
- Complex API for advanced use cases
- Potential for configuration complexity with multiple sources

**Neutral:**
- Opinionated about configuration structure and precedence
- Rich feature set may be overkill for simple configurations

### Confirmation

The decision has been validated through:
- Successful implementation of multi-source configuration loading
- Comprehensive validation with clear error messages
- Environment variable integration working across all platforms
- Hot-reload functionality supporting development workflows
- Security review confirming no sensitive data exposure
- Performance benchmarks showing minimal startup overhead

## Configuration Architecture

### 📋 **Configuration Sources (Precedence Order)**

1. **Command-line Flags** (Highest Priority)
   - Immediate overrides for any configuration value
   - Temporary settings for specific command executions

2. **Environment Variables** 
   - Prefixed with `ZEN_` (e.g., `ZEN_LOG_LEVEL=debug`)
   - Automatic snake_case to camelCase conversion
   - Support for nested configuration (e.g., `ZEN_CLI_VERBOSE=true`)

3. **Configuration Files**
   - Primary: `zen.yaml` in current directory
   - User: `~/.zen/config.yaml`
   - System: `/etc/zen/config.yaml` (Linux/macOS)
   - Format: YAML (primary), JSON, TOML supported

4. **Default Values** (Lowest Priority)
   - Sensible defaults for all configuration options
   - Development-friendly defaults

### 🏗️ **Configuration Structure**

```yaml
# Zen CLI Configuration
log_level: info                    # trace, debug, info, warn, error, fatal, panic
log_format: text                   # text, json

cli:
  no_color: false                  # Disable colored output
  verbose: false                   # Enable verbose output  
  output_format: text              # text, json, yaml

workspace:
  root: .                          # Workspace root directory
  config_file: zen.yaml            # Configuration file name

development:
  debug: false                     # Enable debug mode
  profile: false                   # Enable profiling

# Future configuration sections (planned):
# agents:                          # AI agent configuration
# integrations:                    # External system integrations  
# workflows:                       # Workflow templates and settings
# quality:                         # Quality gate configuration
```

### 🔒 **Security Considerations**

- **No Plaintext Secrets**: Sensitive values referenced via environment variables
- **Configuration Validation**: Schema validation prevents invalid configurations
- **File Permissions**: Configuration files use restrictive permissions (0644)
- **Environment Isolation**: Different configurations for different environments
- **Audit Logging**: Configuration changes logged for security auditing

## Implementation Details

### 📦 **Core Components**

```go
// Configuration structure
type Config struct {
    LogLevel  string `mapstructure:"log_level" validate:"required,oneof=trace debug info warn error fatal panic"`
    LogFormat string `mapstructure:"log_format" validate:"required,oneof=text json"`
    
    CLI         CLIConfig         `mapstructure:"cli"`
    Workspace   WorkspaceConfig   `mapstructure:"workspace"`
    Development DevelopmentConfig `mapstructure:"development"`
}

// Configuration loading with validation
func Load() (*Config, error) {
    v := viper.New()
    setDefaults(v)
    configureViper(v)
    
    if err := v.ReadInConfig(); err != nil {
        // Handle missing config file gracefully
    }
    
    var config Config
    if err := v.Unmarshal(&config); err != nil {
        return nil, fmt.Errorf("failed to unmarshal config: %w", err)
    }
    
    return &config, validate(&config)
}
```

### 🎯 **Configuration Discovery**

1. **Current Directory**: `./zen.yaml`
2. **User Home**: `~/.zen/config.yaml`
3. **XDG Config**: `$XDG_CONFIG_HOME/zen/config.yaml`
4. **System Config**: `/etc/zen/config.yaml` (Unix)

### ✅ **Validation Strategy**

- **Schema Validation**: Comprehensive validation rules for all configuration options
- **Type Safety**: Strong typing with automatic conversion and validation
- **Error Messages**: Clear, actionable error messages with suggestions
- **Default Validation**: Ensure defaults are always valid
- **Migration Support**: Automatic migration from older configuration versions

## Pros and Cons of the Options

### Viper

**Good:**
- Comprehensive multi-source configuration management
- Excellent Cobra integration with automatic flag binding
- Support for multiple configuration formats
- Automatic environment variable binding
- Hot-reload capability for development
- Cross-platform file discovery
- Active community and regular updates
- Extensive documentation and examples

**Bad:**
- Larger dependency footprint
- Complex API for advanced scenarios
- Can be overkill for simple configurations
- Learning curve for advanced features

**Neutral:**
- Opinionated about configuration precedence
- Rich feature set requires understanding of best practices

### Go-Config

**Good:**
- Lightweight and simple API
- Good performance characteristics
- Minimal dependencies

**Bad:**
- Limited multi-source support
- No built-in Cobra integration
- Less comprehensive validation
- Smaller community and ecosystem
- Manual environment variable handling

**Neutral:**
- Simpler but less feature-rich
- Requires more custom implementation

### Custom YAML Parser

**Good:**
- Complete control over configuration logic
- No external dependencies for core functionality
- Optimized for specific use case

**Bad:**
- Significant development and maintenance effort
- Need to implement multi-source logic from scratch
- No built-in validation framework
- Error handling complexity
- No CLI integration helpers

**Neutral:**
- Full flexibility but high implementation cost
- Requires expertise in configuration management patterns

### Environment Variables Only

**Good:**
- Simple and secure approach
- No configuration files to manage
- Easy container and cloud deployment

**Bad:**
- Limited structure for complex configurations
- Poor user experience for many settings
- No configuration file benefits (comments, structure)
- Difficult to manage large configurations

**Neutral:**
- Suitable for simple applications only
- May not scale with feature growth

### Configuration Files Only

**Good:**
- Clear configuration structure
- Easy to version control and share
- Good for complex configurations

**Bad:**
- No runtime environment customization
- Security concerns with sensitive data
- No CLI flag integration
- Poor container deployment experience

**Neutral:**
- Traditional approach with known limitations
- May require additional tooling for deployment

## More Information

**Configuration Examples:**

```yaml
# Development Configuration
log_level: debug
log_format: text
cli:
  verbose: true
  output_format: text
development:
  debug: true
```

```yaml
# Production Configuration  
log_level: info
log_format: json
cli:
  no_color: true
  output_format: json
```

**Environment Variable Examples:**
```bash
# Override log level
export ZEN_LOG_LEVEL=debug

# Enable verbose mode
export ZEN_CLI_VERBOSE=true

# Set output format
export ZEN_CLI_OUTPUT_FORMAT=json
```

**Related ADRs:**
- ADR-0001: Go Language Choice
- ADR-0002: Cobra CLI Framework Selection  
- ADR-0003: Project Structure and Organization
- ADR-0005: Logging Strategy and Implementation

**References:**
- [Viper Documentation](https://github.com/spf13/viper)
- [12-Factor App Configuration](https://12factor.net/config)
- [Configuration Management Best Practices](https://kubernetes.io/docs/concepts/configuration/)
- [YAML Specification](https://yaml.org/spec/)
