# Installation Documentation

This directory contains detailed installation instructions for specific scenarios.

## Primary Installation Guide

For most users, see the main **[Installation Guide](../user-guide/installation.md)** which covers:
- Package manager installation (Homebrew, Scoop, Snap)
- Pre-built binary installation
- Docker installation
- Basic source installation

## Specialized Installation Guides

### Building from Source

**[install-source.md](install-source.md)** - Detailed instructions for building Zen from source code, including:
- Development environment setup
- Cross-compilation for different platforms
- Custom build configurations
- Troubleshooting build issues

### Platform-Specific Guides

Additional platform-specific installation guides (when available):
- `install-linux.md` - Linux distribution-specific instructions
- `install-macos.md` - macOS-specific setup and signing
- `install-windows.md` - Windows-specific configuration
- `install-docker.md` - Container deployment details

## Quick Installation

For the fastest installation path:

```bash
# macOS/Linux with Homebrew
brew install zen

# Windows with Scoop
scoop install zen

# Linux with Snap
sudo snap install zen
```

## Installation Support

If you encounter installation issues:

1. Check the [Troubleshooting section](../user-guide/installation.md#troubleshooting) in the main guide
2. Search [existing issues](https://github.com/zen-org/zen/issues)
3. Open a [new issue](https://github.com/zen-org/zen/issues/new) with:
   - Your operating system and version
   - Installation method attempted
   - Complete error messages
   - Output of relevant commands

## Contributing

To improve installation documentation:
- Report unclear instructions via issues
- Submit pull requests with improvements
- Share platform-specific tips in discussions

---

Return to [User Guide](../user-guide/README.md) | [Documentation Index](../README.md)
