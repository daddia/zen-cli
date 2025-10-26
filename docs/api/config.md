# Configuration API

Type-safe configuration management with multi-source loading and validation.

## Overview

The Configuration API provides a robust system for managing application configuration with:
- Multi-source loading (CLI flags, environment variables, config files, defaults)
- Type-safe configuration with validation
- Component-specific configuration sections
- Clear precedence rules and error handling

## Core Interfaces

### Configurable Interface

All configuration types must implement the `Configurable` interface:

```go
type Configurable interface {
    // Validate validates the configuration and returns an error if invalid
    Validate() error
    
    // Defaults returns a new instance with default values
    Defaults() Configurable
}
```

### ConfigParser Interface

Component-specific parsers implement `ConfigParser[T]`:

```go
type ConfigParser[T Configurable] interface {
    // Parse converts raw configuration data to the typed configuration
    Parse(raw map[string]interface{}) (T, error)
    
    // Section returns the configuration section name
    Section() string
}
```

## Core Configuration

### Config Structure

```go
type Config struct {
    // Core application configuration
    LogLevel  string `mapstructure:"log_level" validate:"required,oneof=trace debug info warn error fatal panic"`
    LogFormat string `mapstructure:"log_format" validate:"required,oneof=text json"`
    
    // Legacy fields (to be migrated)
    Integrations IntegrationsConfig `mapstructure:"integrations"`
    Work         WorkConfig         `mapstructure:"work"`
}
```

### Loading Configuration

#### Basic Loading
```go
// Load with defaults
cfg, err := config.Load()
if err != nil {
    return fmt.Errorf("failed to load config: %w", err)
}
```

#### Loading with Command Context
```go
// Load with CLI flag binding
cfg, err := config.LoadWithCommand(cmd)
if err != nil {
    return fmt.Errorf("failed to load config with command: %w", err)
}
```

#### Loading with Options
```go
opts := config.LoadOptions{
    Command:    cmd,
    ConfigFile: "/path/to/config.yaml",
}

cfg, err := config.LoadWithOptions(opts)
if err != nil {
    return fmt.Errorf("failed to load config with options: %w", err)
}
```

## Type-Safe Configuration

### Getting Component Configuration

```go
// Get CLI configuration
cliConfig, err := config.GetConfig(cfg, cli.ConfigParser{})
if err != nil {
    return fmt.Errorf("failed to get CLI config: %w", err)
}

// Get assets configuration  
assetsConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
if err != nil {
    return fmt.Errorf("failed to get assets config: %w", err)
}
```

### Setting Component Configuration

```go
// Update configuration
newConfig := assets.Config{
    CacheEnabled: true,
    CacheTTL:     30 * time.Minute,
}

err := config.SetConfig(cfg, assets.ConfigParser{}, newConfig)
if err != nil {
    return fmt.Errorf("failed to set assets config: %w", err)
}
```

## Configuration Sources

### Precedence Order
1. **CLI Flags** - Highest priority
2. **Environment Variables** - `ZEN_*` prefix
3. **Configuration Files** - `zen.yaml`, `.zen/config/`
4. **Default Values** - Lowest priority

### Environment Variables
```bash
# Core configuration
export ZEN_LOG_LEVEL=debug
export ZEN_LOG_FORMAT=json
export ZEN_NO_COLOR=true

# Component configuration
export ZEN_ASSETS_CACHE_ENABLED=true
export ZEN_ASSETS_CACHE_TTL=1800
```

### Configuration Files

#### zen.yaml
```yaml
# Core configuration
log_level: info
log_format: text

# Component configurations
assets:
  cache_enabled: true
  cache_ttl: 30m
  repository_url: "https://github.com/zen-org/library"

cli:
  output_format: text
  no_color: false
  verbose: false

# Legacy configuration (to be migrated)
integrations:
  providers:
    jira:
      url: "https://company.atlassian.net"
      project_key: "PROJ"
```

## Component Configuration Examples

### CLI Configuration

```go
type Config struct {
    OutputFormat string `json:"output_format" yaml:"output_format"`
    NoColor      bool   `json:"no_color" yaml:"no_color"`
    Verbose      bool   `json:"verbose" yaml:"verbose"`
}

func (c Config) Validate() error {
    validFormats := []string{"text", "json", "yaml"}
    for _, format := range validFormats {
        if c.OutputFormat == format {
            return nil
        }
    }
    return fmt.Errorf("invalid output format: %s", c.OutputFormat)
}

func (c Config) Defaults() config.Configurable {
    return Config{
        OutputFormat: "text",
        NoColor:      false,
        Verbose:      false,
    }
}

type ConfigParser struct{}

func (p ConfigParser) Parse(raw map[string]interface{}) (Config, error) {
    var cfg Config
    if err := mapstructure.Decode(raw, &cfg); err != nil {
        return cfg, fmt.Errorf("failed to parse CLI config: %w", err)
    }
    return cfg, nil
}

func (p ConfigParser) Section() string {
    return "cli"
}
```

### Assets Configuration

```go
type Config struct {
    CacheEnabled    bool          `json:"cache_enabled" yaml:"cache_enabled"`
    CacheTTL        time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
    RepositoryURL   string        `json:"repository_url" yaml:"repository_url"`
    AuthRequired    bool          `json:"auth_required" yaml:"auth_required"`
}

func (c Config) Validate() error {
    if c.CacheTTL < 0 {
        return fmt.Errorf("cache TTL cannot be negative")
    }
    if c.RepositoryURL == "" {
        return fmt.Errorf("repository URL is required")
    }
    return nil
}

func (c Config) Defaults() config.Configurable {
    return Config{
        CacheEnabled:  true,
        CacheTTL:      30 * time.Minute,
        RepositoryURL: "https://github.com/zen-org/library",
        AuthRequired:  true,
    }
}
```

## Configuration Validation

### Built-in Validation

```go
// Validation occurs automatically during GetConfig
cfg, err := config.GetConfig(baseConfig, parser)
if err != nil {
    if config.IsValidationError(err) {
        // Handle validation error
        fmt.Printf("Configuration validation failed: %v\n", err)
        return err
    }
    // Handle other errors
    return fmt.Errorf("config error: %w", err)
}
```

### Custom Validation

```go
func (c MyConfig) Validate() error {
    var errors []string
    
    if c.Port < 1 || c.Port > 65535 {
        errors = append(errors, "port must be between 1 and 65535")
    }
    
    if c.Host == "" {
        errors = append(errors, "host is required")
    }
    
    if len(errors) > 0 {
        return fmt.Errorf("validation failed: %s", strings.Join(errors, ", "))
    }
    
    return nil
}
```

## Error Handling

### Configuration Errors

```go
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error for field '%s': %s (value: %v)", 
        e.Field, e.Message, e.Value)
}

// Check for validation errors
if config.IsValidationError(err) {
    validationErr := err.(*config.ValidationError)
    fmt.Printf("Field: %s, Message: %s\n", validationErr.Field, validationErr.Message)
}
```

### Loading Errors

```go
cfg, err := config.Load()
if err != nil {
    switch {
    case errors.Is(err, config.ErrConfigFileNotFound):
        // Config file not found - use defaults
        cfg = config.DefaultConfig()
    case errors.Is(err, config.ErrInvalidFormat):
        // Invalid config file format
        return fmt.Errorf("config file format error: %w", err)
    default:
        // Other loading error
        return fmt.Errorf("failed to load config: %w", err)
    }
}
```

## Advanced Usage

### Configuration Watching

```go
// Watch for configuration changes (future feature)
watcher, err := config.NewWatcher(cfg)
if err != nil {
    return err
}

go func() {
    for change := range watcher.Changes() {
        fmt.Printf("Configuration changed: %s\n", change.Key)
        // Reload component configuration
    }
}()
```

### Configuration Profiles

```yaml
# zen.yaml with profiles
default: &default
  log_level: info
  log_format: text

development:
  <<: *default
  log_level: debug
  
production:
  <<: *default
  log_level: warn
  log_format: json
```

```go
// Load profile-specific configuration
cfg, err := config.LoadWithProfile("development")
if err != nil {
    return err
}
```

## Testing

### Test Configuration

```go
func TestConfigLoading(t *testing.T) {
    // Create test configuration
    testConfig := config.Config{
        LogLevel:  "debug",
        LogFormat: "json",
    }
    
    // Test validation
    err := testConfig.Validate()
    assert.NoError(t, err)
    
    // Test defaults
    defaults := testConfig.Defaults()
    assert.NotNil(t, defaults)
}
```

### Mock Configuration

```go
type MockConfigParser struct {
    config MyConfig
    err    error
}

func (m MockConfigParser) Parse(raw map[string]interface{}) (MyConfig, error) {
    return m.config, m.err
}

func (m MockConfigParser) Section() string {
    return "test"
}
```

## Migration Guide

### From Legacy Configuration

```go
// Old way (deprecated)
jiraURL := cfg.Integrations.Providers["jira"].URL

// New way (recommended)
integrationsConfig, err := config.GetConfig(cfg, integrations.ConfigParser{})
if err != nil {
    return err
}
jiraURL := integrationsConfig.Providers["jira"].URL
```

## Best Practices

1. **Always validate configuration** - Implement comprehensive validation
2. **Provide sensible defaults** - Ensure the application works out of the box
3. **Use environment variables** - Support configuration via environment
4. **Document configuration options** - Provide clear documentation
5. **Test configuration loading** - Test all configuration scenarios
6. **Handle errors gracefully** - Provide helpful error messages

## See Also

- **[CLI Configuration](cli.md)** - CLI-specific configuration
- **[Assets Configuration](asset-client.md#configuration)** - Asset library configuration
- **[Workspace Configuration](workspace.md#configuration)** - Workspace settings
- **[Environment Variables](../user-guide/README.md#environment-variables)** - Complete variable reference
