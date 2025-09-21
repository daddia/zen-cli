# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/), and this project adheres to [Semantic Versioning](http://semver.org/).

---

## [Unreleased]

---

## [v0.6.0] - 2025-09-21

### Added
- **Assets Library Migration**: Migrated assets directory from `.zen/assets` to `.zen/library` for improved user experience
  - Updated `zen init` command to create `.zen/library` directory structure
  - Modified all asset management commands to use new library location
  - Updated documentation and help text to reflect library terminology
- **Remote Manifest Location Update**: Updated manifest fetching to use new remote repository structure
  - Changed manifest path from root `manifest.yaml` to `assets/manifest.yaml` in remote repository
  - Updated both HTTP API and Git CLI manifest fetching methods
  - Enhanced error handling for new manifest location
- **Activity-Based Manifest Structure**: Complete redesign of manifest structure from asset types to workflow activities
  - Migrated from asset-type organization (templates, prompts, mcp, schemas) to activity-based organization
  - Updated manifest parser to handle activities with commands, workflow stages, and use cases
  - Enhanced AssetMetadata with command and output file fields for activity support
- **Enhanced Assets List Display**: Improved table format for better usability
  - New table format: `NAME | COMMAND | DESCRIPTION | OUTPUT FORMAT`
  - Alphabetical sorting by activity name for consistent ordering
  - Command display with backticks for CLI command clarity
  - Removed output file column for cleaner interface
  - Enhanced type filtering to work with activity-based structure

### Changed
- Workspace directory structure: `.zen/assets` → `.zen/library`
- Remote manifest location: `manifest.yaml` → `assets/manifest.yaml`
- Manifest organization: Asset types → Workflow activities
- Assets list table: Added sorting, removed output file column, enhanced command display
- Help text and documentation updated to use "library" and "activity" terminology

### Removed
- Backward compatibility for old `.zen/assets` directory structure
- Support for old manifest structure with separate asset type arrays
- Output file column from assets list table for cleaner display

### Technical
- **Parser Redesign**: Complete rewrite of YAML manifest parser for activity structure
- **Type System**: Enhanced AssetMetadata with activity-specific fields (Command, OutputFile)
- **Test Coverage**: Updated all tests for new structure, maintained >80% coverage
- **API Compatibility**: Maintained existing CLI interface while updating internal structure

---

## [v0.5.0] - 2025-09-20

### Added
- **Integration Services Layer**: Complete plugin-based external integration architecture for seamless task synchronization
  - Configuration-driven integration system with support for multiple external platforms
  - `integrations.task_system` configuration for selecting task system of record (jira, github, monday, asana, none)
  - `integrations.sync_enabled` and `integrations.sync_frequency` controls for synchronization behavior
  - Plugin directory discovery system for extensible provider architecture
  - Factory pattern integration with dependency injection for integration components
- **Jira Integration Provider**: Production-ready Jira Cloud integration with comprehensive API support
  - Complete Jira REST API v3 integration with authentication via email + API token
  - Bidirectional field mapping between Zen tasks and Jira issues
  - Task data retrieval, creation, and updates with proper error handling
  - Status and priority mapping with configurable field transformations
  - Connection validation and credential management integration
  - Search functionality for finding existing Jira issues by task ID
- **External System Authentication**: Enhanced authentication system with Jira provider support
  - Added Jira provider configuration to existing multi-provider auth system
  - Support for Basic Auth with email + API token for Jira Cloud
  - Seamless integration with existing secure credential storage (Keychain/File/Memory)
  - Environment variable and configuration file support for credentials
- **Task Creation Integration Hooks**: Non-breaking enhancement to task creation workflow
  - Automatic detection of configured external integration systems
  - Integration status reporting during task creation
  - Graceful handling of integration failures without breaking task creation
  - Metadata directory utilization for external system data storage
- **Data Mapping Framework**: Flexible data transformation system for external system integration
  - Configurable field mapping with dot notation for nested field access
  - Bidirectional data transformation between Zen and external formats
  - Validation framework for mapping configurations
  - Default mappings for common external systems (Jira, GitHub)
  - Type-safe data conversion with reflection-based struct mapping

### Changed
- Enhanced configuration system with new `IntegrationsConfig` section
- Extended authentication system to support additional provider types
- Improved factory dependency injection to include integration management
- Updated task creation workflow with optional integration synchronization
- Enhanced documentation with integration configuration options

### Technical
- **Architecture**: Leverages 70% of existing Zen infrastructure (auth, config, caching, logging, templates)
- **Performance**: Plugin-ready architecture with lazy loading and caching strategies
- **Security**: Secure credential management and sandboxed execution preparation
- **Testing**: Comprehensive test coverage with unit, integration, and mock testing
- **Extensibility**: Clean provider interface for adding new external system integrations
- **Compatibility**: Cross-platform support maintained (Linux, macOS, Windows)

### Foundation for Future WASM Plugins
- Plugin-based architecture designed for WebAssembly runtime integration
- Host API interfaces prepared for secure plugin-to-Zen communication
- Provider abstraction ready for WASM plugin conversion
- Security model foundation for capability-based permissions

---

## [v0.4.0] - 2025-09-20

### Added
- **Task Management System**: Complete task creation and workflow management with template-driven structure
  - `zen task create` command with support for multiple task types (story, bug, epic, spike, task)
  - Structured task directory creation in `.zen/work/tasks/{TASK-ID}/`
  - Template-driven file generation with index.md, manifest.yaml, and .taskrc.yaml
  - Task ID validation with support for alphanumeric and hyphenated formats
  - Priority levels (P0-P3) with P2 as default
  - Owner and team assignment capabilities
  - Interactive title prompting when not provided via flag
- **Task Directory Structure**: Comprehensive workspace organization for task artifacts
  - Main task directory with metadata files
  - Work-type subdirectories created on-demand (research/, spikes/, design/, execution/, outcomes/)
  - .zenflow/ directory for workflow state tracking
  - metadata/ directory for external system snapshots
- **Task Validation System**: Robust input validation and error handling
  - Task ID format validation (minimum 3 characters, alphanumeric with hyphens/underscores)
  - Task type validation against supported types
  - Priority validation with clear error messages
  - Workspace initialization checks before task creation
  - Duplicate task detection and prevention
- **Template Integration**: Seamless integration with the Template Engine for task file generation
  - Template-driven content generation with variable substitution
  - Fallback file generation when templates are unavailable
  - Support for custom task templates via Asset Client
  - Dynamic variable building with task metadata

### Changed
- Enhanced Factory pattern to include task management components
- Improved filesystem utilities with task-specific directory creation
- Extended error handling with task-specific error types and messages
- Updated CLI help system with comprehensive task management examples

### Technical
- Comprehensive test coverage: 89.5% for task creation package
- Performance-optimized directory creation with proper error handling
- Thread-safe task operations with workspace validation
- Integration with existing Asset Client and Template Engine systems
- Proper cleanup and rollback on task creation failures

---

## [v0.3.0] - 2025-09-20

### Added
- **Template Engine Core**: Comprehensive Go template engine with Asset Client integration and Zen-specific extensions
  - Native Go text/template engine with custom function registry (45+ functions)
  - Seamless Asset Client integration for template loading from remote repositories
  - Template compilation caching with configurable TTL and LRU eviction
  - Variable validation with type checking, constraints (regex, ranges, enums), and default values
  - YAML frontmatter and comment-based metadata extraction
  - Factory pattern integration with dependency injection
- **Zen-Specific Template Functions**: Extensive function library for workflow automation
  - **Task Functions**: `taskID`, `taskIDShort`, `randomID` for unique identifier generation
  - **Workflow Functions**: `zenflowStages`, `stageNumber`, `stageName`, `nextStage`, `prevStage` for Zenflow integration
  - **Date/Time Functions**: `now`, `today`, `tomorrow`, `formatDate`, `addDays`, `workingDays` for temporal operations
  - **String Manipulation**: `camelCase`, `pascalCase`, `snakeCase`, `kebabCase`, `slugify`, `titleCase` for naming conventions
  - **Path Functions**: `workspacePath`, `relativePath`, `joinPath`, `fileName`, `fileExt`, `dirName` for file operations
  - **Formatting Functions**: `indent`, `dedent`, `wrap`, `truncate`, `pad` for content formatting
  - **Collection Functions**: `join`, `split`, `contains`, `hasPrefix`, `hasSuffix`, `replace` for data manipulation
  - **Conditional Functions**: `default`, `coalesce`, `ternary` for logic operations
  - **Math Functions**: `add`, `sub`, `mul`, `div`, `mod` for calculations
- **Template Configuration System**: Complete configuration integration for template engine behavior
  - Template compilation caching controls (`templates.cache_enabled`, `templates.cache_ttl`, `templates.cache_size`)
  - Strict mode validation (`templates.strict_mode`) for error handling
  - AI enhancement preparation (`templates.enable_ai`) for future LLM integration
  - Custom delimiter configuration (`templates.left_delim`, `templates.right_delim`) for template syntax
- **Go Template Syntax Support**: Pure Go template syntax with advanced features
  - Variable substitution: `{{.VARIABLE_NAME}}`
  - Conditional rendering: `{{if .condition}}content{{end}}`
  - Loop iteration: `{{range .items}}{{.name}}{{end}}`
  - Function calls: `{{taskID "ZEN"}}`, `{{today}}`, `{{camelCase .title}}`
  - Pipeline operations: `{{.text | trim | upper}}`
  - Template composition: `{{define "template"}}{{template "template" .}}`

### Changed
- Enhanced Factory dependency injection to include Template Engine
- Extended configuration system with templates section
- Improved error handling with Template Engine specific error types
- Updated documentation generation to include template configuration options

### Technical
- Template compilation performance: P95 ≤ 10ms for templates under 50KB
- Memory-efficient caching with configurable size limits and TTL
- Thread-safe concurrent template operations
- Comprehensive test coverage: 84.2% for template package, 69.6% overall
- Security-focused design with input validation and sandboxed execution

---

## [v0.2.0] - 2025-09-19

### Added
- **Asset Management System**: Complete asset repository integration with manifest-driven architecture for private Git repositories
  - Private asset repository support with `zen-assets` integration
  - Manifest-only synchronization strategy for optimal storage efficiency
  - Dynamic asset fetching with session-based caching
  - Asset discovery and listing commands (`zen assets list`, `zen assets sync`)
- **Authentication Component**: Centralized, secure authentication system for Git provider integrations
  - Multi-provider support (GitHub, GitLab) with Personal Access Tokens
  - OS-native credential storage (Keychain, Credential Manager, Secret Service)
  - Encrypted file storage fallback with AES-256-GCM encryption
  - Secure credential management with automatic validation and refresh
  - Authentication commands (`zen assets auth`) for provider setup
- **File-System Cache Component**: Type-safe, persistent caching infrastructure
  - Generic cache manager with Go generics for type safety
  - File-based storage with LRU eviction policies and TTL expiration
  - Thread-safe concurrent access with optimized performance
  - Pluggable serialization strategies (JSON, string)
  - Factory integration for consistent access across components

### Changed
- Enhanced asset management with secure private repository access
- Improved credential handling with multiple storage backend support
- Streamlined caching architecture with type-safe operations

### Security
- Implemented OS-native credential storage for maximum security
- Added encrypted credential persistence with secure defaults
- Enhanced input validation for all authentication operations
- Secure Git provider authentication with proper token management

---

## [v0.1.0] - 2025-09-14

### Added
- **Core CLI Framework**: Implemented root command with Cobra framework and comprehensive help system
- **Workspace Initialization**: Added `zen init` command with automatic project type detection (Git, Node.js, Go, Python, Rust, Java)
- **Configuration Management**: Multi-source configuration system with Viper (CLI flags > environment > config file > defaults)
- **Version Command**: Added `zen version` with detailed build information and multiple output formats (text, JSON, YAML)
- **Status Command**: Added `zen status` for workspace health checking and system information
- **Configuration Commands**: Added `zen config get/set/list` for managing configuration keys with validation
- **Structured Logging**: Implemented logrus-based logging with security-aware field filtering
- **Cross-Platform Support**: Native binaries for Linux (amd64/arm64), macOS (amd64/arm64), and Windows (amd64)
- **Docker Support**: Multi-stage Dockerfile for containerized deployments
- **Comprehensive Testing**: Unit, integration, and end-to-end tests with >80% coverage
- **CI/CD Pipeline**: GitHub Actions workflow with automated testing, linting, and security scanning
- **GoReleaser Configuration**: Automated release pipeline with checksums and multi-platform builds

### Changed
- Enhanced error handling with user-friendly messages and actionable suggestions
- Improved CLI help text with consistent formatting and examples
- Streamlined project structure following Go best practices

### Security
- Implemented secure defaults with no hardcoded secrets
- Added input validation at all boundaries
- Configured Dependabot for automated security updates
- Added security scanning in CI pipeline
