# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/), and this project adheres to [Semantic Versioning](http://semver.org/).

---

## [Unreleased]

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
