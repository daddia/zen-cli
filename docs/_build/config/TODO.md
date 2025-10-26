# Configuration Management Refactor - Comprehensive TODO

**Version:** 1.0  
**Date:** 2025-10-26  
**Status:** Implementation Required  
**Priority:** P0 - Critical Architecture Fix

## Executive Summary

This document provides a comprehensive TODO list for refactoring the configuration management system across the entire Zen CLI platform. The refactor eliminates architectural violations, establishes central config management, and implements standard interfaces across all components.

**Scope:** Complete configuration system refactor with NO backward compatibility.

## PROGRESS UPDATE
**Phase 1: COMPLETED** - Core config module cleaned up
**Phase 2: 80% COMPLETED** - Major components migrated (workspace, task, CLI, development)
**Phase 4: 90% COMPLETED** - Config commands rewritten (fixing import cycles)
**Critical Violations: FIXED** - No more direct file I/O in config commands

---

## Phase 1: Core Config Module Cleanup (Week 1)

### 1.1 Remove Duplicate Config Types from internal/config

**Priority:** P0 - Critical  
**Estimated Effort:** 2 days  
**Dependencies:** None

- [x] **TASK-001**: Remove `WorkspaceConfig` from `internal/config/config.go`
  - File: `internal/config/config.go:78-87`
  - Action: Delete type definition and references
  - Impact: Forces workspace component to own its config
  - **Status: COMPLETED** - WorkspaceConfig removed from Config struct

- [x] **TASK-002**: Remove `WorkConfig` and `TasksConfig` from `internal/config/config.go`
  - File: `internal/config/config.go:99-114`
  - Action: Delete type definitions
  - Impact: Forces task component to own its config
  - **Status: COMPLETED** - WorkConfig and TasksConfig removed

- [x] **TASK-003**: Remove `ProviderConfig` from `internal/config/config.go`
  - File: `internal/config/config.go:117-132`
  - Action: Delete type definition
  - Impact: Forces provider component to own its config
  - **Status: COMPLETED** - ProviderConfig removed

- [x] **TASK-004**: Remove `IntegrationsConfig` and `IntegrationProviderConfig` from `internal/config/config.go`
  - File: `internal/config/config.go:135-180`
  - Action: Delete type definitions
  - Impact: Forces integration component to own its config
  - **Status: COMPLETED** - IntegrationsConfig and IntegrationProviderConfig removed

- [x] **TASK-005**: Remove default functions from `internal/config/config.go`
  - Files: `DefaultWorkConfig()`, `DefaultProvidersConfig()`, `DefaultIntegrationsConfig()`
  - Action: Delete functions (components will provide their own defaults)
  - Impact: Eliminates central knowledge of component defaults
  - **Status: COMPLETED** - Default functions removed

- [x] **TASK-006**: Clean Config struct to only contain core application settings
  - File: `internal/config/config.go:42-63`
  - Action: Remove `Workspace`, `Integrations` fields, keep only `LogLevel`, `LogFormat`, `CLI`, `Development`
  - Impact: Config struct becomes lightweight and focused
  - **Status: COMPLETED** - Config struct contains only core settings

### 1.2 Update Config Loading to Flat Structure

**Priority:** P0 - Critical  
**Estimated Effort:** 1 day  
**Dependencies:** TASK-001 to TASK-006

- [x] **TASK-007**: Update Viper configuration to expect flat structure
  - File: `internal/config/config.go:configureSimpleFileDiscovery()`
  - Action: Configure Viper to read flat YAML structure (assets:, workspace:, etc.)
  - Impact: Config file structure changes from nested to flat
  - **Status: COMPLETED** - Viper already configured for flat structure

- [x] **TASK-008**: Update setDefaults to only set core defaults
  - File: `internal/config/config.go:setDefaults()`
  - Action: Remove component-specific defaults from Options
  - Impact: Components handle their own defaults
  - **Status: COMPLETED** - setDefaults now only handles core options

- [x] **TASK-009**: Update validation to only validate core config
  - File: `internal/config/config.go:validate()`
  - Action: Remove component-specific validation
  - Impact: Components handle their own validation
  - **Status: COMPLETED** - validateCore() now only validates log_level and log_format

### 1.3 Fix Config Options System

**Priority:** P1 - High  
**Estimated Effort:** 1 day  
**Dependencies:** TASK-001 to TASK-006

- [x] **TASK-010**: Remove component-specific options from `internal/config/options.go`
  - File: `internal/config/options.go:Options` array
  - Action: Remove assets, templates, work, provider, integration options
  - Impact: Options only contain core application settings
  - **Status: COMPLETED** - Options array now contains only core log_level and log_format

- [x] **TASK-011**: Remove component-specific getter functions from `internal/config/options.go`
  - Functions: `getAssetsValue()`, `getTemplatesValue()`, `getTasksValue()`, `getProviderValue()`
  - Action: Delete functions (they create circular dependencies)
  - Impact: Config commands must use component parsers directly
  - **Status: COMPLETED** - Component-specific getter functions removed

---

## Phase 2: Component Config Migration (Week 2)

### 2.1 Workspace Component Config

**Priority:** P0 - Critical  
**Estimated Effort:** 1 day  
**Dependencies:** TASK-001

- [x] **TASK-012**: Create `internal/workspace/config.go`
  - Action: Create new file with `Config`, `ConfigParser`, standard interfaces
  - Content: Move `WorkspaceConfig` from `internal/config`
  - Validation: Root, ZenPath validation (removed ConfigFile - workspace doesn't need it)
  - Defaults: Root=".", ZenPath=".zen"
  - **Status: COMPLETED** - Workspace config created with proper separation of concerns

- [x] **TASK-013**: Remove config file interactions from `internal/workspace/workspace.go`
  - Current violations: Direct config file path handling in `New()` function
  - Action: Remove all config file path logic, accept config via parameter
  - Impact: Workspace component becomes config-agnostic
  - **Status: COMPLETED** - Workspace component is now config-agnostic

- [x] **TASK-014**: Update workspace Manager constructor
  - File: `internal/workspace/workspace.go:New()`
  - Action: Change signature to `New(config Config, logger logging.Logger)`
  - Impact: Workspace receives typed config from factory
  - **Status: COMPLETED** - Constructor properly accepts typed workspace.Config

### 2.2 Task Component Config

**Priority:** P0 - Critical  
**Estimated Effort:** 1 day  
**Dependencies:** TASK-002

- [x] **TASK-015**: Create `pkg/task/config.go`
  - Action: Create new file with task configuration types
  - Content: Move `WorkConfig`, `TasksConfig` from `internal/config`
  - Validation: Source validation, sync validation, project key validation
  - Defaults: Source="local", Sync="manual", ProjectKey=""
  - **Status: COMPLETED** - Task config interfaces implemented

- [x] **TASK-016**: Implement task config interfaces
  - Types: `Config`, `ConfigParser`
  - Interfaces: `Configurable`, `ConfigParser[Config]`
  - Section: "task" (not "work" - simplify naming)
  - **Status: COMPLETED** - Task interfaces implemented with "task" section

### 2.3 Integration Component Config

**Priority:** P1 - High  
**Estimated Effort:** 2 days  
**Dependencies:** TASK-004

- [ ] **TASK-017**: Create `internal/integration/config.go`
  - Action: Create new file with integration configuration types
  - Content: Move `IntegrationsConfig`, `IntegrationProviderConfig` from `internal/config`
  - Validation: TaskSystem validation, sync frequency validation
  - Defaults: TaskSystem="", SyncEnabled=false, SyncFrequency="manual"

- [ ] **TASK-018**: Create provider config component
  - Location: `pkg/providers/config.go` (new package)
  - Content: Move `ProviderConfig` from `internal/config`
  - Validation: Type validation, URL validation
  - Defaults: Component-specific defaults for jira, github, linear

### 2.4 CLI Component Config

**Priority:** P1 - High  
**Estimated Effort:** 0.5 days  
**Dependencies:** None

- [x] **TASK-019**: Create `pkg/cli/config.go`
  - Action: Create new file for CLI configuration
  - Content: Move `CLIConfig` from `internal/config`
  - Validation: OutputFormat validation (text, json, yaml)
  - Defaults: NoColor=false, Verbose=false, OutputFormat="text"
  - **Status: COMPLETED** - CLI config interfaces implemented

### 2.5 Development Component Config

**Priority:** P2 - Medium  
**Estimated Effort:** 0.5 days  
**Dependencies:** None

- [x] **TASK-020**: Create `internal/development/config.go`
  - Action: Create new file for development configuration
  - Content: Move `DevelopmentConfig` from `internal/config`
  - Validation: Basic boolean validation
  - Defaults: Debug=false, Profile=false
  - **Status: COMPLETED** - Development config interfaces implemented

---

## Phase 3: Factory Layer Refactor (Week 3)

### 3.1 Update Factory to Use Standard APIs

**Priority:** P0 - Critical  
**Estimated Effort:** 2 days  
**Dependencies:** Phase 2 completion

- [x] **TASK-021**: Update workspace factory integration
  - File: `pkg/cmd/factory/default.go:workspaceFunc()`
  - Action: Use `config.GetConfig(cfg, workspace.ConfigParser{})` 
  - Change: Pass typed config to `workspace.New(workspaceConfig, logger)`
  - **Status: COMPLETED** - Factory now uses central config API for workspace

- [x] **TASK-022**: Update auth factory integration
  - File: `pkg/cmd/factory/default.go:authFunc()`
  - Action: Already implemented, verify it works correctly
  - Validation: Ensure auth config parsing works
  - **Status: COMPLETED** - Auth factory already uses config.GetConfig() API

- [ ] **TASK-023**: Update integration factory integration
  - File: `pkg/cmd/factory/default.go:integrationFunc()`
  - Action: Use `config.GetConfig(cfg, integration.ConfigParser{})`
  - Challenge: Integration component needs to be created first

- [ ] **TASK-024**: Remove unused factory functions
  - Functions: Any remaining manual conversion functions
  - Action: Delete functions that manually convert between config types
  - Impact: Clean factory code using only standard APIs

---

## Phase 4: Config Command Refactor (Week 4)

### 4.1 Major Architectural Violations in Config Commands

**Priority:** P0 - Critical  
**Estimated Effort:** 3 days  
**Dependencies:** Phase 2 completion

- [x] **TASK-025**: Fix config set command violations
  - File: `pkg/cmd/config/set/set.go:setRun()`
  - Current violations:
    - Direct file I/O: `os.MkdirAll()`, `os.ReadFile()`, `os.WriteFile()`
    - Direct YAML parsing: `yaml.Unmarshal()`, `yaml.Marshal()`
    - Hardcoded paths: `".zen/config"`
  - Action: Replace with `config.SetConfig[T]()` API calls
  - Impact: Config set becomes a consumer of central config
  - **Status: COMPLETED** - Config set now uses central config APIs exclusively

- [x] **TASK-026**: Fix config get command violations
  - File: `pkg/cmd/config/get/get.go:getRun()`
  - Current violations: Uses old Options system instead of component parsers
  - Action: Replace with `config.GetConfig[T]()` API calls
  - Impact: Config get uses type-safe component access
  - **Status: COMPLETED** - Config get now uses component parsers and central APIs

- [x] **TASK-027**: Fix config list command violations
  - File: `pkg/cmd/config/list/list.go:listRun()`
  - Current violations: Uses old Options system
  - Action: Replace with component registry and `config.GetConfig[T]()` calls
  - Impact: Config list shows actual component configurations
  - **Status: COMPLETED** - Config list now uses component registry and central APIs

### 4.2 Implement Component Registry for Config Commands

**Priority:** P0 - Critical  
**Estimated Effort:** 1 day  
**Dependencies:** Phase 2 completion

- [x] **TASK-028**: Create component registry
  - File: `pkg/cmd/config/registry.go` (new file)
  - Content: Registry mapping component names to parsers
  - Components: assets, workspace, task, auth, cache, cli, development, integration, providers
  - Impact: Config commands can dynamically handle all components
  - **Status: COMPLETED** - Component registry implemented with all current components

- [x] **TASK-029**: Implement config key parsing
  - File: `pkg/cmd/config/parser.go` (new file)
  - Function: `parseConfigKey(key string) (component, field string, error)`
  - Examples: "assets.repository_url" → "assets", "repository_url"
  - Impact: Config commands can route to appropriate components
  - **Status: COMPLETED** - Config key parsing implemented

- [x] **TASK-030**: Implement field extraction utilities
  - File: `pkg/cmd/config/fields.go` (new file)
  - Functions: `extractFieldValue()`, `updateConfigField()`
  - Purpose: Generic field access using reflection or type switching
  - Impact: Config commands can access any component field
  - **Status: COMPLETED** - Field utilities implemented with reflection

### 4.3 Rewrite Config Commands

**Priority:** P0 - Critical  
**Estimated Effort:** 2 days  
**Dependencies:** TASK-028 to TASK-030

- [x] **TASK-031**: Rewrite config get command
  - File: `pkg/cmd/config/get/get.go`
  - Pattern: Use component registry → get parser → call `config.GetConfig[T]()` → extract field
  - Remove: All direct config access, Options system usage
  - Add: Type-safe component access
  - **Status: COMPLETED** - Config get completely rewritten with component-based approach

- [x] **TASK-032**: Rewrite config set command
  - File: `pkg/cmd/config/set/set.go`
  - Pattern: Use component registry → get parser → call `config.GetConfig[T]()` → update field → call `config.SetConfig[T]()`
  - Remove: All direct file I/O, YAML parsing, hardcoded paths
  - Add: Type-safe component updates
  - **Status: COMPLETED** - Config set completely rewritten with central APIs

- [x] **TASK-033**: Rewrite config list command
  - File: `pkg/cmd/config/list/list.go`
  - Pattern: Iterate through component registry → call `config.GetConfig[T]()` for each → display
  - Remove: Options system usage
  - Add: Dynamic component discovery and display
  - **Status: COMPLETED** - Config list completely rewritten with component registry

---

## Phase 5: Integration and Testing (Week 5)

### 5.1 Factory Integration Updates

**Priority:** P0 - Critical  
**Estimated Effort:** 1 day  
**Dependencies:** Phase 3 completion

- [ ] **TASK-034**: Update factory workspace integration
  - File: `pkg/cmd/factory/default.go:workspaceFunc()`
  - Action: Use `config.GetConfig(cfg, workspace.ConfigParser{})`
  - Validation: Workspace manager receives typed config

- [ ] **TASK-035**: Update factory task integration
  - File: `pkg/cmd/factory/default.go` (add taskFunc if needed)
  - Action: Use `config.GetConfig(cfg, task.ConfigParser{})`
  - Validation: Task manager receives typed config

- [ ] **TASK-036**: Update factory integration manager
  - File: `pkg/cmd/factory/default.go:integrationFunc()`
  - Action: Use `config.GetConfig(cfg, integration.ConfigParser{})`
  - Validation: Integration manager receives typed config

### 5.2 Test Suite Updates

**Priority:** P1 - High  
**Estimated Effort:** 2 days  
**Dependencies:** Phase 4 completion

- [x] **TASK-037**: Update config module tests
  - File: `internal/config/config_test.go`
  - Action: Remove tests for deleted config types
  - Add: Tests for new standard interface APIs
  - Validation: All config module tests pass
  - **Status: COMPLETED** - Config module tests updated and passing

- [ ] **TASK-038**: Update config command tests
  - Files: `pkg/cmd/config/*/test.go`
  - Action: Update tests to use new component-based approach
  - Remove: Tests that expect direct file manipulation
  - Add: Tests that verify central config API usage

- [ ] **TASK-039**: Create component config tests
  - Files: `*/config_test.go` for each component
  - Action: Test `Validate()`, `Defaults()`, `Parse()`, `Section()` methods
  - Coverage: 95% test coverage for all config implementations

- [ ] **TASK-040**: Create integration tests (`test/integration`)
  - File: `test/integration/config_integration_test.go` (enhance existing)
  - Action: Test full config flow from file → component → factory
  - Validation: End-to-end config system works

### 5.3 Performance and Security Validation

**Priority:** P1 - High  
**Estimated Effort:** 1 day  
**Dependencies:** Phase 4 completion

- [ ] **TASK-041**: Validate performance requirements (`test/performance`)
  - Requirement: P95 ≤ 10ms for config loading
  - Requirement: P95 ≤ 1ms for component config parsing
  - File: `test/performance/config_performance_test.go` (enhance existing)
  - Validation: Benchmark tests pass performance thresholds

- [ ] **TASK-042**: Security validation
  - Action: Verify sensitive data redaction works per component
  - Test: Each component's config redacts sensitive fields
  - Validation: No sensitive data in logs or output

---

## Phase 6: Component-Specific Implementation (Week 6)

### 6.1 Workspace Component

**Priority:** P0 - Critical  
**Estimated Effort:** 1 day  
**Dependencies:** TASK-012

- [x] **TASK-043**: Implement workspace config interfaces
  - File: `internal/workspace/config.go`
  - Status: Partially implemented, needs completion
  - Action: Ensure all interfaces work correctly
  - Validation: Workspace config parsing works
  - **Status: COMPLETED** - Workspace config interfaces fully implemented

- [x] **TASK-044**: Update workspace manager to use typed config
  - File: `internal/workspace/workspace.go`
  - Action: Update constructor to accept `workspace.Config`
  - Remove: All config file path logic
  - Impact: Workspace becomes config-agnostic
  - **Status: COMPLETED** - Workspace manager now uses typed config exclusively

### 6.2 Task Component

**Priority:** P0 - Critical  
**Estimated Effort:** 1 day  
**Dependencies:** TASK-015

- [x] **TASK-045**: Implement task config interfaces
  - File: `pkg/task/config.go` (new file)
  - Content: Task configuration types and interfaces
  - Validation: Source, sync, project key validation
  - Impact: Task component owns its configuration
  - **Status: COMPLETED** - Task config interfaces fully implemented

- [ ] **TASK-046**: Update task manager to use typed config
  - File: `pkg/task/manager.go`
  - Action: Update constructor to accept `task.Config`
  - Remove: Any direct config access
  - Impact: Task manager becomes config-agnostic

### 6.3 Integration Component

**Priority:** P1 - High  
**Estimated Effort:** 2 days  
**Dependencies:** TASK-017

- [ ] **TASK-047**: Implement integration config interfaces
  - File: `internal/integration/config.go` (new file)
  - Content: Integration configuration types and interfaces
  - Validation: Task system, sync frequency validation
  - Impact: Integration component owns its configuration

- [ ] **TASK-048**: Update integration service to use typed config
  - File: `internal/integration/service.go`
  - Action: Update constructor to accept `integration.Config`
  - Remove: Direct config struct access
  - Impact: Integration service becomes config-agnostic

### 6.4 Provider Component

**Priority:** P1 - High  
**Estimated Effort:** 1 day  
**Dependencies:** TASK-018

- [ ] **TASK-049**: Create provider config component
  - File: `pkg/providers/config.go` (new file)
  - Content: Provider configuration types and interfaces
  - Validation: Type, URL validation
  - Impact: Provider configuration is self-contained

- [ ] **TASK-050**: Update provider implementations
  - Files: `internal/providers/*/` 
  - Action: Update to use typed provider config
  - Remove: Direct config access
  - Impact: Providers become config-agnostic

### 6.5 CLI Component

**Priority:** P2 - Medium  
**Estimated Effort:** 0.5 days  
**Dependencies:** TASK-019

- [x] **TASK-051**: Implement CLI config interfaces
  - File: `pkg/cli/config.go` (new file)
  - Content: CLI configuration types and interfaces
  - Validation: Output format validation
  - Impact: CLI configuration is self-contained
  - **Status: COMPLETED** - CLI config interfaces implemented

- [x] **TASK-052**: Update iostreams to use typed config
  - File: `pkg/cmd/factory/default.go:ioStreams()`
  - Action: Use `config.GetConfig(cfg, cli.ConfigParser{})`
  - Impact: IOStreams uses typed CLI config
  - **Status: COMPLETED** - IOStreams now uses cli.ConfigParser{}

### 6.6 Development Component

**Priority:** P3 - Low  
**Estimated Effort:** 0.5 days  
**Dependencies:** TASK-020

- [x] **TASK-053**: Implement development config interfaces
  - File: `internal/development/config.go` (new file)
  - Content: Development configuration types and interfaces
  - Validation: Basic boolean validation
  - Impact: Development configuration is self-contained
  - **Status: COMPLETED** - Development config interfaces implemented

---

## Phase 7: Architecture Compliance and Validation (Week 7)

### 7.1 Architecture Tests

**Priority:** P0 - Critical  
**Estimated Effort:** 1 day  
**Dependencies:** All previous phases

- [ ] **TASK-054**: Create architecture compliance tests
  - File: `test/architecture/config_test.go` (new file)
  - Tests:
    - No component except `internal/config` imports `viper`
    - No component uses file I/O for config operations
    - No component hardcodes `.zen/config` paths
    - All components implement required interfaces
  - Impact: Prevents future architectural violations

- [ ] **TASK-055**: Create config integration tests
  - File: `test/integration/config_integration_test.go` (new file)
  - Tests:
    - Full config flow: file → central config → component → factory
    - Config commands work with all components
    - Performance requirements are met
  - Impact: Validates entire config system works together

### 7.2 Documentation Updates

**Priority:** P1 - High  
**Estimated Effort:** 1 day  
**Dependencies:** Implementation completion

- [ ] **TASK-056**: Update component documentation
  - Files: Each component's README or doc files
  - Action: Document new config interfaces and usage
  - Content: How to access component config via factory
  - Impact: Developers understand new config system

- [ ] **TASK-057**: Update config command documentation
  - Files: `docs/zen/zen_config*.md`
  - Action: Update examples to use new component.field syntax
  - Content: Document all available config keys per component
  - Impact: Users understand new config structure

### 7.3 Migration and Cleanup

**Priority:** P1 - High  
**Estimated Effort:** 1 day  
**Dependencies:** All implementation complete

- [ ] **TASK-058**: Create config migration utility
  - File: `internal/config/migrate.go` (new file)
  - Purpose: Migrate old config files to new flat structure
  - Action: Convert nested config to flat component structure
  - Impact: Existing users can migrate seamlessly

- [ ] **TASK-059**: Remove deprecated code
  - Files: Any remaining old config code
  - Action: Remove unused types, functions, imports
  - Validation: No dead code remains
  - Impact: Clean codebase

---

## Phase 8: Validation and Performance (Week 8)

### 8.1 End-to-End Validation

**Priority:** P0 - Critical  
**Estimated Effort:** 2 days  
**Dependencies:** All previous phases

- [ ] **TASK-060**: Validate all CLI commands work
  - Commands: `zen init`, `zen config get`, `zen config set`, `zen config list`, `zen status`
  - Action: Test each command with new config system
  - Validation: No regressions, all functionality preserved
  - Impact: User experience is maintained

- [ ] **TASK-061**: Validate all factory integrations work
  - Components: Assets, Auth, Template, Cache, Workspace, Task, Integration
  - Action: Test each component receives correct typed config
  - Validation: All components initialize correctly
  - Impact: All functionality works with new config system

- [ ] **TASK-062**: Performance regression testing
  - Metrics: Config loading time, component parsing time, memory usage
  - Baseline: Current performance measurements
  - Target: No regression, meet P95 requirements
  - Impact: New system is as fast or faster

### 8.2 Security and Compliance

**Priority:** P1 - High  
**Estimated Effort:** 1 day  
**Dependencies:** Implementation completion

- [ ] **TASK-063**: Security audit
  - Action: Verify no sensitive data leakage in new system
  - Test: Component-specific redaction works
  - Validation: Security requirements met
  - Impact: System is secure

- [ ] **TASK-064**: Compliance validation
  - Action: Verify new system meets all architectural requirements
  - Test: All components follow standard interfaces
  - Validation: Technical specification fully implemented
  - Impact: System meets design requirements

---

## Critical Path Analysis

### Must Complete First (Blockers)
1. **TASK-001 to TASK-006**: Core config cleanup (enables everything else)
2. **TASK-025**: Fix config set violations (critical user functionality)
3. **TASK-012**: Workspace config migration (most used component)

### High Risk Items
- **TASK-025**: Config set command has major violations
- **TASK-023**: Integration factory needs integration component first
- **TASK-060**: End-to-end validation may reveal integration issues

### Dependencies
- Phase 2 must complete before Phase 3 (factory needs component configs)
- Phase 3 must complete before Phase 4 (config commands need factory)
- All phases must complete before Phase 8 (validation needs everything)

---

## Success Criteria

### Technical Success
- [ ] Zero duplicate config types across components
- [ ] All components implement standard interfaces
- [ ] Only `internal/config` touches configuration files
- [ ] Config commands use central APIs exclusively
- [ ] Performance requirements met (P95 ≤ 10ms config loading)

### User Success
- [ ] All existing CLI functionality preserved
- [ ] Config commands work with new component structure
- [ ] No breaking changes to user workflows
- [ ] Clear error messages for invalid configurations

### Architecture Success
- [ ] Clean separation of concerns
- [ ] Type-safe configuration access
- [ ] Extensible design for new components
- [ ] No architectural violations remain

---

## Risk Mitigation

### High Risk
- **Config Command Rewrite**: Extensive testing required
- **Factory Integration**: Complex dependency chain
- **Component Migration**: Potential for missed references

### Mitigation Strategies
- Implement comprehensive test suite before refactoring
- Use feature flags for gradual rollout
- Create rollback procedures for each phase
- Daily validation of critical user workflows

---

**Total Estimated Effort:** 64 tasks across 8 weeks  
**Critical Path:** 8 weeks with parallel execution  
**Risk Level:** High (major architectural change)  
**Success Probability:** High (with proper testing and validation)
