# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/), and this project adheres to [Semantic Versioning](http://semver.org/).

---

## [Unreleased]

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
