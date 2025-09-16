# Installation

Zen CLI can be installed using package managers, pre-built binaries, or from source code.

## System Requirements

### Minimum Requirements
- **Operating System**: Windows 10+, macOS 11+, or Linux (Ubuntu 20.04+, RHEL 8+, or equivalent)
- **Memory**: 4 GB RAM
- **Disk Space**: 500 MB available space
- **Network**: Internet connection for AI features

### Optional Requirements
- **Go 1.23+**: Only needed for building from source
- **Docker**: For containerized deployments
- **Git**: For version control integration

## Recommended Installation

Choose the installation method that best suits your environment.

### Package Managers

#### macOS and Linux (Homebrew)

```bash
# Install
brew install zen

# Verify installation
zen version
```

#### Windows (Scoop)

```powershell
# Install
scoop install zen

# Verify installation
zen version
```

#### Linux (Snap)

```bash
# Install
sudo snap install zen

# Verify installation
zen version
```

### Pre-built Binaries

Download the latest release for your platform:

1. Visit the [releases page](https://github.com/zen-org/zen/releases)
2. Download the appropriate archive for your system:
   - `zen-darwin-amd64.tar.gz` - macOS Intel
   - `zen-darwin-arm64.tar.gz` - macOS Apple Silicon
   - `zen-linux-amd64.tar.gz` - Linux x64
   - `zen-linux-arm64.tar.gz` - Linux ARM
   - `zen-windows-amd64.zip` - Windows x64

3. Extract the archive:

   **macOS/Linux:**
   ```bash
   tar -xzf zen-*.tar.gz
   sudo mv zen /usr/local/bin/
   chmod +x /usr/local/bin/zen
   ```

   **Windows:**
   ```powershell
   # Extract the zip file
   Expand-Archive zen-windows-amd64.zip -DestinationPath C:\zen
   
   # Add to PATH
   [Environment]::SetEnvironmentVariable(
       "Path",
       $env:Path + ";C:\zen",
       [EnvironmentVariableTarget]::User
   )
   ```

4. Verify the installation:
   ```bash
   zen version
   ```

### Docker

Run Zen in a container without installing locally:

```bash
# Run latest version
docker run --rm -it zen:latest

# Run with local directory mounted
docker run --rm -it -v $(pwd):/workspace zen:latest

# Create alias for convenience
alias zen='docker run --rm -it -v $(pwd):/workspace zen:latest'
```

### Install from Source

For developers or users who need the latest development version:

```bash
# Requirements: Go 1.23+
go install github.com/zen-org/zen/cmd/zen@latest

# Or clone and build
git clone https://github.com/zen-org/zen.git
cd zen
make install
```

For detailed source installation instructions, see [Building from Source](../installation/install-source.md).

## Post-Installation Setup

### Verify Installation

Confirm Zen is correctly installed:

```bash
# Check version
zen version

# View available commands
zen --help

# Run system check
zen status
```

### Shell Completion

Enable command completion for your shell:

#### Bash
```bash
# Add to ~/.bashrc
source <(zen completion bash)
```

#### Zsh
```bash
# Add to ~/.zshrc
source <(zen completion zsh)
```

#### Fish
```bash
# Add to ~/.config/fish/config.fish
zen completion fish | source
```

#### PowerShell
```powershell
# Add to $PROFILE
zen completion powershell | Out-String | Invoke-Expression
```

### Environment Variables

Configure Zen using environment variables:

```bash
# Set default configuration directory
export ZEN_CONFIG_DIR="$HOME/.config/zen"

# Set log level
export ZEN_LOG_LEVEL="info"

# Set API keys for AI providers
export ZEN_OPENAI_API_KEY="your-key-here"
export ZEN_ANTHROPIC_API_KEY="your-key-here"
```

## Updating Zen

### Package Managers

```bash
# Homebrew
brew upgrade zen

# Scoop
scoop update zen

# Snap
sudo snap refresh zen
```

### Binary Updates

```bash
# Check current version
zen version

# Download and replace binary
# Follow installation steps with new version
```

### Docker Updates

```bash
# Pull latest image
docker pull zen:latest
```

## Uninstallation

### Package Managers

```bash
# Homebrew
brew uninstall zen

# Scoop
scoop uninstall zen

# Snap
sudo snap remove zen
```

### Manual Uninstallation

```bash
# Remove binary
sudo rm /usr/local/bin/zen

# Remove configuration (optional)
rm -rf ~/.config/zen
rm -rf ~/.zen
```

## Troubleshooting

### Common Issues

#### Command Not Found
Ensure the installation directory is in your PATH:

```bash
# Check PATH
echo $PATH

# Add to PATH temporarily
export PATH=$PATH:/usr/local/bin

# Add permanently (bash)
echo 'export PATH=$PATH:/usr/local/bin' >> ~/.bashrc
source ~/.bashrc
```

#### Permission Denied
Fix executable permissions:

```bash
chmod +x /usr/local/bin/zen
```

#### Version Mismatch
Remove old versions before installing:

```bash
which -a zen  # Find all installations
# Remove unwanted versions
```

#### SSL/TLS Errors
Update certificates:

```bash
# macOS
brew install ca-certificates

# Linux
sudo apt-get update && sudo apt-get install ca-certificates
```

### Getting Help

- Check the [Quick Start Guide](quick-start.md)
- Review [Configuration](configuration.md) options
- Search [existing issues](https://github.com/zen-org/zen/issues)
- Open a [new issue](https://github.com/zen-org/zen/issues/new)

## Platform-Specific Notes

### macOS
- Gatekeeper may block the first run. Allow in System Preferences > Security & Privacy
- Apple Silicon users should use the arm64 version for best performance

### Windows
- Run as Administrator for first-time setup if installing system-wide
- Windows Defender may scan the binary on first run

### Linux
- SELinux users may need to adjust contexts
- AppArmor profiles may need updating for snap installations

## Next Steps

- Follow the [Quick Start Guide](quick-start.md) to begin using Zen
- Configure Zen for your environment with the [Configuration Guide](configuration.md)
- Explore available commands with `zen --help`
