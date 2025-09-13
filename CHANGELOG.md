# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/), and this project adheres to [Semantic Versioning](http://semver.org/).

---

## [Unreleased]

### Added
- **ZEN-003**: Enhanced `zen init` command with comprehensive workspace detection
  - Automatic project type detection (Git, Node.js, Go, Python, Rust, Java)
  - Creates `.zen/` directory structure similar to `.git/`
  - Generates project-specific configuration with detected metadata
  - Framework detection (React, Vue, Angular, Express, Maven, Gradle, etc.)
  - Language detection (TypeScript, JavaScript, Go, Python, Rust, Java)
  - Force flag (`--force`) with automatic backup creation
  - Automatic `.gitignore` updates
  - Comprehensive error handling with user-friendly messages

---

## [v0.1.0] - YYYY-MM-DD

### Added
- Describe new features or functionalities.
- Example: Added support for feature XYZ.

### Changed
- Detail updates to existing functionality.
- Example: Updated dependency ABC to version 1.2.3.

### Fixed
- List bug fixes.
- Example: Fixed issue with module loading on startup.

### Deprecated
- Mention any soon-to-be removed features.
- Example: Deprecated legacy API endpoints.

### Removed
- Document features that have been removed.
- Example: Removed support for legacy configuration files.

### Security
- Note any security-related changes.
- Example: Patched vulnerability in dependency DEF.
