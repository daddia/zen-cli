# Configuration Management - Technical Specification

**Version:** 2.0  
**Author:** Configuration Architecture Team  
**Date:** 2025-10-26  
**Status:** Implementation Required

## Executive Summary

The current configuration management system exhibits architectural violations and tight coupling between modules. This specification defines a comprehensive refactor to implement proper separation of concerns, eliminate duplicate configuration types, and establish clean APIs for configuration access.

**Key Issues Identified:**
- Duplicate configuration types across modules (e.g., `config.AssetsConfig` vs `assets.AssetConfig`)
- Manual conversion logic in factory layer
- Tight coupling between config module and domain modules
- Violation of single responsibility principle
- No public APIs for configuration parsing

## CRITICAL DESIGN PRINCIPLE: Central Config Management

**! ARCHITECTURAL RULE: Only `internal/config` touches configuration files. ALL other components MUST use the config module.**

### The Central Config Management Pattern

```
┌─────────────────────────────────────────────────────────────┐
│                    CLI Commands                             │
│  (config get, config set, config list)                     │
├─────────────────────────────────────────────────────────────┤
│                    Factory Layer                            │
│  (Gets typed configs for components)                        │
├─────────────────────────────────────────────────────────────┤
│               CENTRAL CONFIG MODULE                         │
│  internal/config - ONLY component that touches files       │
│  • Reads/writes .zen/config                                │
│  • Manages Viper instance                                  │
│  • Provides GetConfig[T]() and SetConfig[T]() APIs         │
├─────────────────────────────────────────────────────────────┤
│              Component Config Types                         │
│  • workspace.Config (defines structure only)               │
│  • assets.Config (defines structure only)                  │
│  • task.Config (defines structure only)                    │
│  • NO FILE ACCESS - only type definitions                  │
├─────────────────────────────────────────────────────────────┤
│                  File System                                │
│  .zen/config - managed ONLY by internal/config             │
└─────────────────────────────────────────────────────────────┘
```

### What Each Layer Does

#### + Central Config Module (`internal/config`)
- **ONLY** component that reads/writes config files
- Manages Viper instance and file I/O
- Provides standard APIs: `GetConfig[T]()`, `SetConfig[T]()`
- Handles config source precedence (files, env vars, flags)
- Validates config file integrity

#### + Component Config Types (workspace, assets, task, etc.)
- Define their config structure (`type Config struct`)
- Implement standard interfaces (`Configurable`, `ConfigParser[T]`)
- Provide validation logic (`Validate()` method)
- Provide defaults (`Defaults()` method)
- **NEVER** touch config files
- **NEVER** import Viper
- **NEVER** know about file paths

#### + Factory Layer
- Uses central config: `config.GetConfig(cfg, workspace.ConfigParser{})`
- Passes typed config to components
- **NEVER** touches config files directly

#### + CLI Commands (`pkg/cmd/config`)
- Use central config APIs for all operations
- `config get`: Uses `config.GetConfig[T]()` with appropriate parser
- `config set`: Uses `config.SetConfig[T]()` with appropriate parser
- `config list`: Uses central config to enumerate available sections
- **NEVER** touch config files directly
- **NEVER** import Viper directly

### - VIOLATIONS - What Components Must NEVER Do

#### File System Violations
- **NEVER** import `github.com/spf13/viper`
- **NEVER** read config files with `os.ReadFile`, `ioutil.ReadFile`, etc.
- **NEVER** write config files with `os.WriteFile`, `ioutil.WriteFile`, etc.
- **NEVER** check config file existence with `os.Stat`
- **NEVER** create config directories with `os.MkdirAll`
- **NEVER** parse YAML/JSON config files directly
- **NEVER** know about `.zen/config` file paths
- **NEVER** hardcode config file names or paths

#### Architectural Violations
- **NEVER** bypass central config APIs
- **NEVER** duplicate config types across components
- **NEVER** implement custom config loading logic
- **NEVER** cache config data outside of central config
- **NEVER** directly access Viper instances

#### Examples of Violations to Avoid

```go
// - WRONG - Component touching config files
func (w *WorkspaceManager) loadConfig() error {
    data, err := os.ReadFile(".zen/config")  // VIOLATION!
    return yaml.Unmarshal(data, &w.config)   // VIOLATION!
}

// - WRONG - Component importing Viper
import "github.com/spf13/viper"  // VIOLATION!

// - WRONG - Component checking config file existence
if _, err := os.Stat(".zen/config"); err == nil {  // VIOLATION!
    // do something
}

// - WRONG - Config command touching files directly
func runSet(key, value string) error {
    data, err := os.ReadFile(".zen/config")  // VIOLATION!
    // modify data
    return os.WriteFile(".zen/config", data, 0644)  // VIOLATION!
}
```

#### + Correct Patterns

```go
// + CORRECT - Component defines config type only
type Config struct {
    Root string `yaml:"root"`
}

func (c Config) Validate() error { /* validation only */ }
func (c Config) Defaults() config.Configurable { /* defaults only */ }

// + CORRECT - Factory uses central config
workspaceConfig, err := config.GetConfig(cfg, workspace.ConfigParser{})
workspace := workspace.New(workspaceConfig, logger)

// + CORRECT - Config command uses central config
func runSet(f *cmdutil.Factory, key, value string) error {
    cfg, err := f.Config()  // Get central config
    // Use config.SetConfig() API
    return config.SetConfig(cfg, parser, updatedConfig)
}
```

## Goals and Non-Goals

### Goals
- Establish single source of truth for configuration schema
- Implement clean separation between config management and domain logic
- Provide type-safe configuration APIs for modules
- Eliminate duplicate configuration types
- Enable module-specific configuration validation
- Support dynamic configuration reloading
- Maintain backward compatibility for existing config files

### Non-Goals
- Change existing config file formats or locations
- Modify Viper as the core configuration engine
- Break existing CLI functionality
- Implement distributed configuration management

## Requirements

### Functional Requirements
- **FR-1**: Each module defines its own configuration types
  - Priority: High
  - Acceptance Criteria: No duplicate config types across modules
- **FR-2**: Config module provides parsing APIs for modules
  - Priority: High  
  - Acceptance Criteria: Clean `ParseConfig[T]()` API available
- **FR-3**: Module-specific configuration validation
  - Priority: Medium
  - Acceptance Criteria: Each module validates its own config
- **FR-4**: Dynamic configuration updates
  - Priority: Low
  - Acceptance Criteria: Support config reloading without restart

### Non-Functional Requirements
- **NFR-1**: Configuration loading performance
  - Category: Performance
  - Target: P95 ≤ 10ms for config loading
  - Measurement: Benchmark tests
- **NFR-2**: Memory efficiency
  - Category: Resource Usage
  - Target: ≤ 1MB memory overhead for config management
  - Measurement: Memory profiling
- **NFR-3**: Type safety
  - Category: Code Quality
  - Target: 100% compile-time type checking
  - Measurement: Static analysis

## System Architecture

### High-Level Design

The new configuration management system implements a **layered architecture** with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                     CLI Layer                               │
├─────────────────────────────────────────────────────────────┤
│                   Factory Layer                             │
├─────────────────────────────────────────────────────────────┤
│  Config Management Layer (Viper + Parsing APIs)            │
├─────────────────────────────────────────────────────────────┤
│  Component/Module-Specific Config Types (Assets, Templates, etc.)    │
├─────────────────────────────────────────────────────────────┤
│                File System Layer                            │
└─────────────────────────────────────────────────────────────┘
```

### Component Architecture

#### Configuration Manager
- **Purpose:** Central configuration loading and parsing coordination
- **Technology:** Go + Viper v1.25.3+
- **Interfaces:** `ConfigManager`, `ConfigParser[T]`
- **Dependencies:** Viper, file system

#### Config Parsers
- **Purpose:** Standard configuration parsing and validation for any component
- **Technology:** Go generics for type safety
- **Interfaces:** `ConfigParser[T]`, `Configurable`
- **Dependencies:** Config Manager

#### Configuration Sources
- **Purpose:** Hierarchical configuration source management
- **Technology:** Viper source management
- **Interfaces:** Viper's built-in interfaces
- **Dependencies:** File system, environment

### Data Architecture

#### Data Models

##### Core Configuration
```go
type Config struct {
    // Core application settings
    LogLevel  string `mapstructure:"log_level"`
    LogFormat string `mapstructure:"log_format"`
    
    // CLI settings
    CLI CLIConfig `mapstructure:"cli"`
    
    // Workspace settings
    Workspace WorkspaceConfig `mapstructure:"workspace"`
    
    // Raw module configurations (untyped)
    Modules map[string]interface{} `mapstructure:"modules"`
    
    // Internal fields
    viper      *viper.Viper
    configFile string
    loadedFrom []string
}
```

##### Standard Configuration Interfaces
```go
// Core interface that all configuration types must implement
type Configurable interface {
    Validate() error
    Defaults() Configurable
}

// Generic parser interface for type-safe configuration parsing
type ConfigParser[T Configurable] interface {
    Parse(raw map[string]interface{}) (T, error)
    Section() string // Returns the config section name (e.g., "assets", "templates")
}

// Config manager provides standard APIs for any component
type Manager interface {
    GetConfig[T Configurable](parser ConfigParser[T]) (T, error)
    SetConfig[T Configurable](parser ConfigParser[T], config T) error
    WatchConfig[T Configurable](parser ConfigParser[T], callback func(T)) error
}
```

#### Data Flow
1. **Load**: Viper loads raw configuration from sources
2. **Parse**: Config manager unmarshals core config
3. **Extract**: Components request their config sections via standard APIs
4. **Convert**: Type-safe parsers convert raw data to component types
5. **Validate**: Component-specific validation rules applied

### API Design

#### GetConfig[T] API
- **Purpose:** Retrieve typed configuration for any component
- **Request:**
  ```go
  func (m *Manager) GetConfig[T Configurable](
      parser ConfigParser[T]
  ) (T, error)
  ```
- **Response:**
  ```go
  type ConfigResponse[T] struct {
      Config T      `json:"config"`
      Source string `json:"source"`
      Valid  bool   `json:"valid"`
  }
  ```
- **Error Codes:** `ErrSectionNotFound`, `ErrInvalidConfig`, `ErrParseError`

#### SetConfig[T] API
- **Purpose:** Update typed configuration for any component
- **Request:**
  ```go
  func (m *Manager) SetConfig[T Configurable](
      parser ConfigParser[T],
      config T
  ) error
  ```
- **Response:** Success/error status
- **Error Codes:** `ErrValidationFailed`, `ErrWriteError`

## Config Command Integration

### CRITICAL: How Config Commands Work

The `pkg/cmd/config` commands are **consumers** of the central config management system. They must follow the same pattern as all other components.

#### Config Command Architecture

```go
// pkg/cmd/config/get/get.go
func runGet(f *cmdutil.Factory, key string) error {
    // 1. Get central config manager
    cfg, err := f.Config()
    if err != nil {
        return err
    }
    
    // 2. Parse key to determine component and field
    component, field := parseConfigKey(key) // e.g., "assets.repository_url" -> "assets", "repository_url"
    
    // 3. Use appropriate component parser
    switch component {
    case "assets":
        assetConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
        if err != nil {
            return err
        }
        return displayConfigValue(field, assetConfig)
    case "workspace":
        workspaceConfig, err := config.GetConfig(cfg, workspace.ConfigParser{})
        if err != nil {
            return err
        }
        return displayConfigValue(field, workspaceConfig)
    // ... other components
    }
}

// pkg/cmd/config/set/set.go
func runSet(f *cmdutil.Factory, key, value string) error {
    // 1. Get central config manager
    cfg, err := f.Config()
    if err != nil {
        return err
    }
    
    // 2. Parse key to determine component and field
    component, field := parseConfigKey(key)
    
    // 3. Get current config, update field, set back
    switch component {
    case "assets":
        assetConfig, err := config.GetConfig(cfg, assets.ConfigParser{})
        if err != nil {
            return err
        }
        
        // Update the specific field
        updatedConfig := updateConfigField(assetConfig, field, value)
        
        // Set back using central config
        return config.SetConfig(cfg, assets.ConfigParser{}, updatedConfig)
    // ... other components
    }
}
```

#### Key Principles for Config Commands

1. **No Direct File Access**: Config commands use `config.GetConfig[T]()` and `config.SetConfig[T]()`
2. **Component Awareness**: Commands know which parser to use for each component
3. **Type Safety**: All config access is type-safe through component parsers
4. **Central Validation**: All validation happens through component `Validate()` methods

### Config Command Flow

```
zen config get assets.repository_url
    ↓
pkg/cmd/config/get
    ↓
config.GetConfig(cfg, assets.ConfigParser{})
    ↓
internal/config reads .zen/config via Viper
    ↓
assets.ConfigParser.Parse() converts to assets.Config
    ↓
Display assets.Config.RepositoryURL
```

```
zen config set assets.repository_url "https://github.com/new/repo.git"
    ↓
pkg/cmd/config/set
    ↓
config.GetConfig(cfg, assets.ConfigParser{}) // Get current config
    ↓
Update assets.Config.RepositoryURL field
    ↓
config.SetConfig(cfg, assets.ConfigParser{}, updatedConfig)
    ↓
internal/config writes .zen/config via Viper
```

### Config Command Implementation Requirements

#### Component Registry Pattern
Config commands must maintain a registry of available components and their parsers:

```go
// pkg/cmd/config/registry.go
type ComponentRegistry struct {
    parsers map[string]interface{} // component name -> parser instance
}

func NewComponentRegistry() *ComponentRegistry {
    return &ComponentRegistry{
        parsers: map[string]interface{}{
            "assets":    assets.ConfigParser{},
            "workspace": workspace.ConfigParser{},
            "task":      task.ConfigParser{},
            "auth":      auth.ConfigParser{},
            "cache":     cache.ConfigParser{},
            // Add new components here
        },
    }
}

func (r *ComponentRegistry) GetParser(component string) (interface{}, bool) {
    parser, exists := r.parsers[component]
    return parser, exists
}

func (r *ComponentRegistry) ListComponents() []string {
    components := make([]string, 0, len(r.parsers))
    for component := range r.parsers {
        components = append(components, component)
    }
    return components
}
```

#### Config Get Command Implementation

```go
// pkg/cmd/config/get/get.go
func runGet(f *cmdutil.Factory, key string) error {
    cfg, err := f.Config()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    component, field, err := parseConfigKey(key)
    if err != nil {
        return fmt.Errorf("invalid config key %s: %w", key, err)
    }
    
    registry := config.NewComponentRegistry()
    
    switch component {
    case "assets":
        return getComponentConfig(cfg, assets.ConfigParser{}, field)
    case "workspace":
        return getComponentConfig(cfg, workspace.ConfigParser{}, field)
    case "task":
        return getComponentConfig(cfg, task.ConfigParser{}, field)
    // ... other components
    default:
        return fmt.Errorf("unknown component: %s", component)
    }
}

func getComponentConfig[T config.Configurable](cfg *config.Config, parser config.ConfigParser[T], field string) error {
    componentConfig, err := config.GetConfig(cfg, parser)
    if err != nil {
        return fmt.Errorf("failed to get %s config: %w", parser.Section(), err)
    }
    
    value, err := extractFieldValue(componentConfig, field)
    if err != nil {
        return fmt.Errorf("failed to get field %s: %w", field, err)
    }
    
    fmt.Println(value)
    return nil
}
```

#### Config Set Command Implementation

```go
// pkg/cmd/config/set/set.go
func runSet(f *cmdutil.Factory, key, value string) error {
    cfg, err := f.Config()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    component, field, err := parseConfigKey(key)
    if err != nil {
        return fmt.Errorf("invalid config key %s: %w", key, err)
    }
    
    switch component {
    case "assets":
        return setComponentConfig(cfg, assets.ConfigParser{}, field, value)
    case "workspace":
        return setComponentConfig(cfg, workspace.ConfigParser{}, field, value)
    case "task":
        return setComponentConfig(cfg, task.ConfigParser{}, field, value)
    // ... other components
    default:
        return fmt.Errorf("unknown component: %s", component)
    }
}

func setComponentConfig[T config.Configurable](cfg *config.Config, parser config.ConfigParser[T], field, value string) error {
    // Get current config
    componentConfig, err := config.GetConfig(cfg, parser)
    if err != nil {
        return fmt.Errorf("failed to get %s config: %w", parser.Section(), err)
    }
    
    // Update the field
    updatedConfig, err := updateConfigField(componentConfig, field, value)
    if err != nil {
        return fmt.Errorf("failed to update field %s: %w", field, err)
    }
    
    // Set back using central config
    if err := config.SetConfig(cfg, parser, updatedConfig); err != nil {
        return fmt.Errorf("failed to save %s config: %w", parser.Section(), err)
    }
    
    fmt.Printf("✓ Set %s.%s = %s\n", parser.Section(), field, value)
    return nil
}
```

#### Config List Command Implementation

```go
// pkg/cmd/config/list/list.go
func runList(f *cmdutil.Factory) error {
    cfg, err := f.Config()
    if err != nil {
        return fmt.Errorf("failed to load config: %w", err)
    }
    
    registry := config.NewComponentRegistry()
    
    for _, component := range registry.ListComponents() {
        switch component {
        case "assets":
            listComponentConfig(cfg, assets.ConfigParser{})
        case "workspace":
            listComponentConfig(cfg, workspace.ConfigParser{})
        case "task":
            listComponentConfig(cfg, task.ConfigParser{})
        // ... other components
        }
    }
    
    return nil
}

func listComponentConfig[T config.Configurable](cfg *config.Config, parser config.ConfigParser[T]) {
    componentConfig, err := config.GetConfig(cfg, parser)
    if err != nil {
        fmt.Printf("Error loading %s config: %v\n", parser.Section(), err)
        return
    }
    
    fmt.Printf("[%s]\n", parser.Section())
    displayConfigStruct(componentConfig)
    fmt.Println()
}
```

## Implementation Details

### Technology Stack
- **Configuration Engine**: Viper v1.20.0+ (battle-tested, feature-rich)
  - Justification: Proven in production, supports multiple sources
- **Type System**: Go 1.25+ generics for type-safe parsing
  - Justification: Compile-time type safety, eliminates runtime errors
- **Validation**: Custom validation interfaces per module
  - Justification: Domain-specific validation rules

### Algorithms and Logic

#### Configuration Parsing Algorithm
- **Purpose:** Type-safe conversion from raw config to any component type
- **Complexity:** O(1) per component (constant time lookup and conversion)
- **Implementation:**
  ```go
  func (m *Manager) GetConfig[T Configurable](
      parser ConfigParser[T]
  ) (T, error) {
      // 1. Extract section using parser's section name
      sectionData := m.viper.GetStringMap(parser.Section())
      
      // 2. Type-safe parsing
      config, err := parser.Parse(sectionData)
      if err != nil {
          // Return defaults on parse error
          return config.Defaults().(T), fmt.Errorf("parse error: %w", err)
      }
      
      // 3. Validation
      if err := config.Validate(); err != nil {
          return config, fmt.Errorf("validation error: %w", err)
      }
      
      return config, nil
  }
  ```
- **Performance:** Sub-millisecond parsing per component

### Security Design
- **Configuration Isolation**: Each component only accesses its own config section
  - Risk Mitigation: Prevents cross-component configuration leakage
  - Validation: Interface-based access control
- **Sensitive Data Handling**: Component-specific redaction rules
  - Risk Mitigation: Domain-aware sensitive field detection
  - Validation: Unit tests for redaction logic
- **File Access Control**: Only config module touches config files
  - Risk Mitigation: Single point of file system access
  - Validation: Architecture tests prevent violations

### Error Handling Strategy

Comprehensive error handling with typed errors and contextual information:

#### Error Types
- **ConfigNotFoundError**: Configuration section missing
  - Handling: Return component defaults
  - User Message: "Using default configuration for {component}"
  - Logging: Debug level
- **ParseError**: Invalid configuration format
  - Handling: Fail fast with detailed error
  - User Message: "Invalid configuration in {section}: {details}"
  - Logging: Error level with context
- **ValidationError**: Configuration values invalid
  - Handling: Fail fast with validation details
  - User Message: "Configuration validation failed: {field} {error}"
  - Logging: Error level with field context

## Performance & Scalability

### Performance Requirements
- **Config Loading**: P95 ≤ 10ms
  - Current: ~5ms (measured)
  - Measurement: Benchmark tests with realistic config files
- **Component Config Parsing**: P95 ≤ 1ms per component
  - Current: Not measured
  - Measurement: Per-component parsing benchmarks
- **Memory Usage**: ≤ 1MB total for config management
  - Current: ~500KB
  - Measurement: Memory profiling

### Scalability Design
Configuration system scales with number of components through:
- **O(1) component lookup**: Hash map-based component config access
- **Lazy parsing**: Component configs parsed only when requested
- **Caching**: Parsed configs cached until file changes

### Load Testing Requirements
- **Concurrent Access**: 1000 concurrent config reads
  - Load: 1000 goroutines accessing config simultaneously
  - Duration: 30 seconds
  - Success Criteria: No race conditions, consistent results
- **File Watching**: 100 config file changes per second
  - Load: Rapid file modifications
  - Duration: 60 seconds  
  - Success Criteria: All changes detected and processed

### Caching Strategy
- **Parsed Config Cache**: In-memory caching of parsed component configs
  - Technology: sync.Map for concurrent access
  - TTL: Until file modification detected
  - Invalidation: File system watcher triggers

## Database Design

### Schema Changes
No database changes required - configuration is file-based.

### Configuration File Schema
```yaml
# .zen/config
log_level: info
log_format: text

cli:
  no_color: false
  verbose: false
  output_format: text

workspace:
  root: .
  config_file: config

# Component-specific configurations
assets:
  repository_url: https://github.com/daddia/zen-assets.git
  branch: main
  auth_provider: github
  cache_path: ~/.zen/library
  cache_size_mb: 100
  sync_timeout_seconds: 30
  integrity_checks_enabled: true
  prefetch_enabled: true
  
templates:
  cache_enabled: true
  cache_ttl: 30m
  cache_size: 100
  strict_mode: false
  enable_ai: false
  left_delim: "{{"
  right_delim: "}}"
```

## Integration Design

### External Integrations
- **Viper**: Configuration loading and source management
  - Protocol: Direct library integration
  - Authentication: N/A
  - Rate Limits: File system I/O limits
  - Error Handling: Viper error wrapping

### Component Responsibilities Matrix

| Component | Owns Config Type | Implements Interfaces | Touches Files | Uses Central Config |
|-----------|------------------|----------------------|---------------|-------------------|
| `internal/config` | Core config only | N/A | + YES - ONLY component | N/A |
| `internal/workspace` | `workspace.Config` | + YES | - NEVER | + Via factory |
| `pkg/assets` | `assets.Config` | + YES | - NEVER | + Via factory |
| `pkg/task` | `task.Config` | + YES | - NEVER | + Via factory |
| `pkg/auth` | `auth.Config` | + YES | - NEVER | + Via factory |
| `pkg/cmd/config` | None | N/A | - NEVER | + Direct API calls |
| `pkg/cmd/factory` | None | N/A | - NEVER | + Gets configs for components |

### Component Integration Pattern
Each component implements the standard `Configurable` interface:

```go
// In pkg/assets/config.go
type Config struct {
    RepositoryURL string `yaml:"repository_url"`
    Branch        string `yaml:"branch"`
    // ... other fields
}

// Implement Configurable interface
func (c Config) Validate() error {
    if c.RepositoryURL == "" {
        return errors.New("repository_url is required")
    }
    return nil
}

func (c Config) Defaults() Configurable {
    return Config{
        RepositoryURL: "https://github.com/daddia/zen-assets.git",
        Branch:        "main",
        // ... other defaults
    }
}

// Implement ConfigParser interface
type Parser struct{}

func (p Parser) Section() string {
    return "assets"
}

func (p Parser) Parse(raw map[string]interface{}) (Config, error) {
    var config Config
    if err := mapstructure.Decode(raw, &config); err != nil {
        return config, err
    }
    return config, nil
}

// Usage in factory:
assetConfig, err := configManager.GetConfig(assets.Parser{})
```

## Testing Strategy

### Test Coverage
- **Unit Tests**: 95% coverage target
  - Framework: testify
  - Strategy: Test each parser and validator independently
- **Integration Tests**: 90% coverage target
  - Framework: testify + real config files
  - Strategy: Test full config loading and module integration
- **Architecture Tests**: 100% coverage target
  - Framework: Custom architecture tests
  - Strategy: Verify no direct file access violations

### Test Cases

#### TC-001: Module Config Parsing
- **Type:** Unit Test
- **Priority:** High
- **Preconditions:** Valid module config section exists
- **Test Steps:** 
  1. Load raw config with module section
  2. Parse using module parser
  3. Validate parsed config
- **Expected Result:** Correctly typed and validated module config

#### TC-002: Missing Module Config
- **Type:** Unit Test
- **Priority:** High
- **Preconditions:** Module config section missing from file
- **Test Steps:**
  1. Load config without module section
  2. Request module config
  3. Verify defaults returned
- **Expected Result:** Module defaults returned, no errors

#### TC-003: Invalid Module Config
- **Type:** Unit Test
- **Priority:** High
- **Preconditions:** Invalid values in module config
- **Test Steps:**
  1. Load config with invalid module values
  2. Parse using module parser
  3. Verify validation fails
- **Expected Result:** Validation error with specific field details

### Test Automation
Automated testing integrated into CI/CD pipeline with:
- Pre-commit hooks for config validation
- Architecture tests preventing violations
- Performance regression detection

### Test Data Management
- **Config Fixtures**: Standardized test configuration files
  - Generation: Template-based generation for different scenarios
  - Privacy: No sensitive data in test configs
  - Lifecycle: Version controlled, cleaned up after tests

## Deployment & Operations

### Deployment Strategy
Configuration changes deployed through:
1. **Code Deployment**: New parsers and validators
2. **Config Migration**: Automated migration of existing configs
3. **Validation**: Pre-deployment config validation
4. **Rollback**: Automatic rollback on validation failures

### Environment Configuration
- **Development**: Local `.zen/config` files
  - Configuration: Debug logging, relaxed validation
  - Data: Test data and mock services
  - Monitoring: Local logging only
- **Production**: System-wide `/etc/zen/config`
  - Configuration: Production logging, strict validation
  - Data: Production data sources
  - Monitoring: Full observability stack

### Feature Flags
- **config_validation**: Enable strict configuration validation
  - Default State: Enabled
  - Rollout Strategy: Immediate (safety feature)

### Monitoring & Observability
- **config_load_duration**: Configuration loading time
  - Type: Histogram
  - Alert Threshold: P95 > 50ms
  - Dashboard: Configuration Performance
- **config_parse_errors**: Configuration parsing error rate
  - Type: Counter
  - Alert Threshold: > 1% error rate
  - Dashboard: Configuration Health
- **config_validation_failures**: Validation failure rate per module
  - Type: Counter per module
  - Alert Threshold: > 5% failure rate
  - Dashboard: Configuration Quality

### Operational Runbooks
- [Config Loading Failures](./runbooks/config-loading.md): Troubleshooting config load issues
- [Module Config Migration](./runbooks/config-migration.md): Migrating module configurations
- [Performance Debugging](./runbooks/config-performance.md): Debugging config performance issues

## Risk Assessment

### Technical Risks
- **Breaking Changes**: Refactoring may break existing integrations
  - Probability: Medium
  - Impact: High
  - Mitigation: Comprehensive testing, gradual rollout, feature flags
- **Performance Regression**: New parsing layer may add overhead
  - Probability: Low
  - Impact: Medium
  - Mitigation: Benchmark tests, performance monitoring
- **Type Safety Violations**: Generic parsing may lose type safety
  - Probability: Low
  - Impact: High
  - Mitigation: Compile-time type checking, comprehensive tests

### Performance Risks
- **Memory Overhead**: Additional parsing layer may increase memory usage
  - Mitigation: Lazy parsing, efficient caching, memory profiling
  - Monitoring: Memory usage metrics per module
- **CPU Overhead**: Generic parsing may be slower than direct access
  - Mitigation: Benchmark-driven optimization, caching
  - Monitoring: CPU usage during config operations

### Security Risks
- **Configuration Injection**: Malicious config values could affect modules
  - Impact: Medium (limited to config values)
  - Controls: Input validation, type checking, sandboxed parsing
- **Sensitive Data Exposure**: Module configs may contain secrets
  - Impact: High (credential exposure)
  - Controls: Module-specific redaction, secure defaults

## Dependencies

### Internal Dependencies
- **Viper Library**: Core configuration engine
  - Owner: Config Management Team
  - Timeline: Already integrated
  - Risk: Low (stable, mature library)
- **Config Interfaces**: Each module/component must implement config interfaces
  - Risk: Medium (requires coordination across teams)

### External Dependencies
- **File System**: Configuration file storage
  - SLA: Local file system reliability
  - Fallback: In-memory defaults
  - Contact: Infrastructure team
- **Environment Variables**: Runtime configuration overrides
  - SLA: OS environment reliability
  - Fallback: File-based configuration
  - Contact: Platform team

### Library Dependencies
- **github.com/spf13/viper**: v1.20.0+
  - Purpose: Configuration loading and source management
  - License: MIT (compatible)
  - Alternatives: koanf, envconfig (less feature-rich)
- **github.com/mitchellh/mapstructure**: v1.5.0+ (Viper dependency)
  - Purpose: Raw map to struct conversion
  - License: MIT (compatible)
  - Alternatives: Built into Viper

## Migration and Rollback

### Migration Plan

1. **Phase 1: Interface Definition**
   - Description: Define Config interfaces and parsing APIs
   - Validation: Interface contracts compile and test
   - Rollback: Remove interfaces, no impact

2. **Phase 2: Module Implementation**
   - Description: Implement config types in each module
   - Validation: All modules have config parsers
   - Rollback: Remove module config types, use factory conversion

3. **Phase 3: Factory Refactor** (Week 3)
   - Description: Update factory to use new parsing APIs
   - Validation: All factory tests pass
   - Rollback: Restore manual conversion logic

4. **Phase 4: Config Module Cleanup** (Week 4)
   - Description: Remove duplicate config types from config module
   - Validation: No duplicate types remain
   - Rollback: Restore original config types

### Rollback Strategy
**Automated rollback triggers:**
- Configuration parsing errors > 5%
- Module initialization failures > 1%
- Performance regression > 20%

**Rollback procedure:**
1. Feature flag disable (`modular_config=false`)
2. Restore factory conversion logic
3. Validate system functionality
4. Monitor for 24 hours

### Compatibility Matrix
| Component | Current Version | New Version | Compatibility |
|-----------|----------------|-------------|---------------|
| Config Module | v1.0 (monolithic) | v2.0 (modular) | Breaking changes |
| Assets Module | v1.0 (factory conversion) | v2.0 (self-contained) | Compatible |
| Template Module | v1.0 (factory conversion) | v2.0 (self-contained) | Compatible |
| Factory Layer | v1.0 (manual conversion) | v2.0 (API-based) | Breaking changes |

## Timeline and Milestones

### Implementation Timeline
| Phase | Duration | Deliverables | Owner |
|-------|----------|--------------|-------|
| Interface Design | 1 week | ModuleConfig interfaces, parsing APIs | Config Team |
| Module Implementation | 2 weeks | Config types in assets, templates, etc. | Module Teams |
| Factory Refactor | 1 week | Updated factory using new APIs | Factory Team |
| Testing & Validation | 1 week | Comprehensive test suite | QA Team |
| Documentation | 1 week | Updated docs and migration guides | Tech Writing |

### Critical Milestones
- **Interface Freeze**: Week 1 end
  - Success Criteria: All interfaces defined and approved
  - Dependencies: Architecture review completion
- **Module Migration Complete**: Week 3 end
  - Success Criteria: All modules using self-contained configs
  - Dependencies: Module team coordination
- **Production Deployment**: Week 6 end
  - Success Criteria: Zero-downtime deployment successful
  - Dependencies: Testing completion, rollback procedures tested

## Success Criteria

### Technical Success
- **Zero Duplicate Config Types**: No duplicate configuration types across modules
  - Target: 0 duplicates
  - Validation: Static analysis, architecture tests
- **Type Safety**: 100% compile-time type checking for config access
  - Target: 0 runtime type errors
  - Validation: Comprehensive test suite
- **Performance Maintained**: Config loading performance unchanged
  - Target: P95 ≤ 10ms (current: ~5ms)
  - Validation: Benchmark comparison

### Performance Success
- **Memory Efficiency**: Memory usage for config management
  - Target: ≤ 1MB total overhead
  - Baseline: ~500KB current
  - Measurement: Memory profiling
- **CPU Efficiency**: CPU overhead for config parsing
  - Target: ≤ 1ms per module parse
  - Baseline: Not currently measured
  - Measurement: CPU profiling

### User Success
- **Developer Experience**: Ease of adding new config options
  - Target: ≤ 5 minutes to add new config field
  - Method: Developer surveys and timing studies
  - Timeline: Post-implementation measurement
- **Configuration Errors**: Reduction in config-related issues
  - Target: 50% reduction in config-related bugs
  - Method: Issue tracking analysis
  - Timeline: 3 months post-deployment

## Implementation Plan

### Current State Analysis
**Problems Identified:**
1. **Tight Coupling**: Config module contains domain-specific types
2. **Code Duplication**: `config.AssetsConfig` vs `assets.AssetConfig`
3. **Manual Conversion**: Factory layer does manual type conversion
4. **No Validation**: No module-specific validation
5. **Poor Separation**: Config concerns mixed with domain logic

### Proposed Architecture

```go
// New standard interface architecture

// 1. Config module provides standard interfaces and manager
package config

// Standard interfaces that all components implement
type Configurable interface {
    Validate() error
    Defaults() Configurable
}

type ConfigParser[T Configurable] interface {
    Parse(raw map[string]interface{}) (T, error)
    Section() string
}

type Manager struct {
    viper *viper.Viper
}

func (m *Manager) GetConfig[T Configurable](parser ConfigParser[T]) (T, error) {
    raw := m.viper.GetStringMap(parser.Section())
    return parser.Parse(raw)
}

// 2. Each component implements standard interfaces
package assets

type Config struct {
    RepositoryURL string `yaml:"repository_url"`
    Branch        string `yaml:"branch"`
    // ... other fields
}

func (c Config) Validate() error { /* component validation */ }
func (c Config) Defaults() Configurable { /* component defaults */ }

type Parser struct{}
func (p Parser) Section() string { return "assets" }
func (p Parser) Parse(raw map[string]interface{}) (Config, error) { /* parsing */ }

// 3. Factory uses standard APIs
package factory

func assetClientFunc(f *cmdutil.Factory) func() (assets.AssetClientInterface, error) {
    return func() (assets.AssetClientInterface, error) {
        configManager, err := f.ConfigManager()
        if err != nil {
            return nil, err
        }
        
        // Clean standard API call
        assetConfig, err := configManager.GetConfig(assets.Parser{})
        if err != nil {
            return nil, err
        }
        
        return assets.NewClient(assetConfig, logger), nil
    }
}
```

### Benefits of New Architecture
1. **Standard Interfaces**: All components use the same configuration pattern
2. **Type Safety**: Compile-time guarantees for config access
3. **Extensibility**: Easy to add new components and config options
4. **Testability**: Component configs can be tested independently
5. **Maintainability**: Clear ownership and boundaries
6. **Consistency**: Uniform configuration approach across all components

## Appendices

### Glossary
- **Configurable**: Standard interface that all configuration types must implement
- **ConfigParser**: Generic interface for parsing raw config to typed structs
- **Config Manager**: Central coordinator for configuration loading and parsing
- **Component Boundary**: Clear separation between config management and domain logic

### References
- [Viper Documentation](https://github.com/spf13/viper)
- [Go Configuration Best Practices](https://peter.bourgon.org/go-best-practices-2016/#configuration)
- [Twelve-Factor App Config](https://12factor.net/config)
- [Clean Architecture Patterns](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

### Architecture Decision Records
- [ADR-025: Standard Configuration Interfaces](../decisions/ADR-025-standard-config.md) - Unified config interface design
- [ADR-026: Generic Config Parsing](../decisions/ADR-026-generic-parsing.md) - Type-safe configuration APIs
- [ADR-027: Component Config Ownership](../decisions/ADR-027-config-ownership.md) - Component responsibility boundaries

---

**Specification Status:** Draft  
**Review Status:** Pending Architecture Review  
**Approval Date:** TBD

## Immediate Action Items

### High Priority
1. **Remove Duplicate Types**: Eliminate `config.AssetsConfig`, `config.TemplatesConfig`, etc.
2. **Define Standard Interfaces**: Create `Configurable` and `ConfigParser[T]` interfaces
3. **Implement Generic APIs**: Add `GetConfig[T]()` and `SetConfig[T]()` methods

### Medium Priority
1. **Component Migration**: Move config types to respective components
2. **Factory Refactor**: Update factory to use standard APIs
3. **Validation Implementation**: Add component-specific validation

### Low Priority
1. **Performance Optimization**: Optimize parsing performance
2. **Dynamic Reloading**: Add config file watching
3. **Advanced Features**: Configuration templating, environment-specific overrides

## Summary: Architectural Boundaries

### + The Golden Rule
**ONLY `internal/config` touches configuration files. Every other component is a consumer.**

### - Component Checklist

Before implementing any component config:

- [ ] + Component defines its own `Config` struct
- [ ] + Component implements `Configurable` interface (`Validate()`, `Defaults()`)
- [ ] + Component implements `ConfigParser[T]` interface (`Parse()`, `Section()`)
- [ ] - Component does NOT import `viper`
- [ ] - Component does NOT touch any files
- [ ] - Component does NOT know about `.zen/config` paths
- [ ] + Factory gets config via `config.GetConfig(cfg, component.ConfigParser{})`
- [ ] + Config commands use `config.GetConfig[T]()` and `config.SetConfig[T]()`

### ! Enforcement

**Architecture tests MUST verify:**
1. No component except `internal/config` imports `viper`
2. No component except `internal/config` uses file I/O for config
3. No component hardcodes `.zen/config` paths
4. All components implement required interfaces
5. Config commands use central APIs only

## File Organization Standards

### Standard Config File Naming Convention

Each component MUST follow the standard file naming pattern for configuration:

#### Primary Pattern: `config.go`
```
pkg/assets/config.go          # Asset configuration types and interfaces
pkg/auth/config.go            # Auth configuration types and interfaces  
pkg/cache/config.go           # Cache configuration types and interfaces
pkg/task/config.go            # Task configuration types and interfaces
internal/workspace/config.go  # Workspace configuration types and interfaces
```

#### File Structure Template
Each component's `config.go` file MUST contain:

```go
// pkg/component/config.go
package component

import (
    "fmt"
    "github.com/daddia/zen/internal/config"
    "github.com/go-viper/mapstructure/v2"
)

// Config represents component-specific configuration
type Config struct {
    Field1 string `yaml:"field1" json:"field1" mapstructure:"field1"`
    Field2 int    `yaml:"field2" json:"field2" mapstructure:"field2"`
    // ... other fields
}

// DefaultConfig returns default component configuration
func DefaultConfig() Config {
    return Config{
        Field1: "default_value",
        Field2: 42,
        // ... other defaults
    }
}

// Implement config.Configurable interface

// Validate validates the component configuration
func (c Config) Validate() error {
    if c.Field1 == "" {
        return fmt.Errorf("field1 is required")
    }
    if c.Field2 <= 0 {
        return fmt.Errorf("field2 must be positive")
    }
    return nil
}

// Defaults returns a new Config with default values
func (c Config) Defaults() config.Configurable {
    return DefaultConfig()
}

// ConfigParser implements config.ConfigParser[Config] interface
type ConfigParser struct{}

// Parse converts raw configuration data to Config
func (p ConfigParser) Parse(raw map[string]interface{}) (Config, error) {
    // Start with defaults to ensure all fields are properly initialized
    cfg := DefaultConfig()
    
    // If raw data is empty, return defaults
    if len(raw) == 0 {
        return cfg, nil
    }
    
    // Use mapstructure to decode the raw map into our config struct
    decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
        Result:           &cfg,
        WeaklyTypedInput: true,
        DecodeHook: mapstructure.ComposeDecodeHookFunc(
            mapstructure.StringToTimeDurationHookFunc(),
        ),
    })
    if err != nil {
        return cfg, fmt.Errorf("failed to create decoder: %w", err)
    }
    
    if err := decoder.Decode(raw); err != nil {
        return cfg, fmt.Errorf("failed to decode component config: %w", err)
    }
    
    return cfg, nil
}

// Section returns the configuration section name
func (p ConfigParser) Section() string {
    return "component"  // Replace with actual component name
}
```

### Alternative Patterns (When Appropriate)

#### Pattern 2: Separate Files for Complex Components
For components with extensive configuration:

```
pkg/component/
├── config.go           # Main config types and interfaces
├── config_defaults.go  # Default configurations and presets
├── config_validation.go # Complex validation logic
└── config_test.go      # Configuration tests
```

#### Pattern 3: Embedded in Main File (Simple Components)
For very simple components with minimal config:

```go
// pkg/simple/simple.go
package simple

// Config can be embedded in main file if very simple
type Config struct {
    Enabled bool `yaml:"enabled"`
}

// ... implement interfaces inline
```

### File Naming Rules

#### ✓ MUST Follow
- **Primary file**: `config.go` (preferred)
- **Test file**: `config_test.go`
- **Package name**: Same as directory name
- **Config type**: `Config` (not `ComponentConfig` or `ComponentSettings`)
- **Parser type**: `ConfigParser` (not `ComponentParser` or `Parser`)

#### ✓ MAY Use (for complex components)
- `config_defaults.go` - Complex default configurations
- `config_validation.go` - Extensive validation logic
- `config_types.go` - Additional config-related types

#### - MUST NOT Use
- `settings.go`, `options.go`, `params.go` - Use `config.go`
- `configuration.go` - Too verbose, use `config.go`
- `cfg.go` - Too abbreviated, use `config.go`
- Component-prefixed names like `asset_config.go` - Redundant with package name

### Import Standards

#### Required Imports
```go
import (
    "fmt"                                    // For error messages
    "github.com/daddia/zen/internal/config"  // For interfaces
    "github.com/go-viper/mapstructure/v2"    // For parsing (if needed)
)
```

#### Conditional Imports
```go
import (
    "time"     // If config has time.Duration fields
    "strconv"  // If config has string conversion needs
    "strings"  // If config has string manipulation
)
```

#### Forbidden Imports
```go
import (
    "github.com/spf13/viper"     // VIOLATION! Only internal/config uses Viper
    "gopkg.in/yaml.v3"          // VIOLATION! Use central config APIs
    "encoding/json"             // VIOLATION! Use central config APIs
    "os"                        // VIOLATION! (unless for non-config file operations)
    "io/ioutil"                 // VIOLATION! No direct file I/O
)
```

### Documentation Standards

Each `config.go` file MUST include:

```go
// Package documentation
// Package component provides [component functionality].
// 
// Configuration is managed through the standard config.Configurable interface.
// Use config.GetConfig(cfg, component.ConfigParser{}) to access configuration.
package component

// Config represents [component] configuration.
// 
// This type implements config.Configurable interface and defines all
// configuration options available for the [component] component.
//
// Example usage:
//   cfg, err := config.GetConfig(configManager, component.ConfigParser{})
//   if err != nil {
//       return err
//   }
//   // Use cfg.Field1, cfg.Field2, etc.
type Config struct {
    // Field documentation with validation rules
    Field1 string `yaml:"field1" json:"field1" mapstructure:"field1"`
}
```

### Testing Standards

Each component MUST include config tests in `config_test.go`:

```go
// pkg/component/config_test.go
package component_test

import (
    "testing"
    "github.com/daddia/zen/pkg/component"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestConfig_Validate(t *testing.T) {
    // Test validation logic
}

func TestConfig_Defaults(t *testing.T) {
    // Test default values
}

func TestConfigParser_Parse(t *testing.T) {
    // Test parsing logic
}

func TestConfigParser_Section(t *testing.T) {
    // Test section name
}
```

**The current configuration system requires immediate architectural refactoring to eliminate tight coupling and establish proper separation of concerns.**
