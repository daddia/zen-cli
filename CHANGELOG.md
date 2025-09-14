# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/), and this project adheres to [Semantic Versioning](http://semver.org/).

---

## [Unreleased]

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
