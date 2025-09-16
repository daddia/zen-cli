# Building Zen from Source

This guide provides detailed instructions for building Zen CLI from source code.

## Prerequisites

### Required Tools

- **Go 1.25+** - Required for building
- **Git 2.30+** - For cloning the repository  
- **Make** (optional) - For using build automation

### Verify Prerequisites

```bash
# Check Go version
go version

# Check Git version  
git version

# Check Make (optional)
make --version
```

If any tools are missing, install them:
- **Go**: Follow instructions on [the Go website](https://golang.org/doc/install)
- **Git**: Visit [Git downloads](https://git-scm.com/downloads)
- **Make**: Install via your package manager

## Building from Source

### 1. Clone the Repository

```bash
# Clone the official repository
git clone https://github.com/zen-org/zen.git
cd zen

# Or clone your fork
git clone https://github.com/YOUR_USERNAME/zen.git
cd zen
```

### 2. Build and Install

#### Unix-like Systems (Linux/macOS)

```bash
# Build using Make (recommended)
make build

# Install to /usr/local/bin (may require sudo)
sudo make install

# Or install to custom location
make install prefix=$HOME/.local

# Or build directly with Go
go build -o bin/zen ./cmd/zen
```

#### Windows

```powershell
# Build using the build script (recommended)
go run script\build.go

# Or use the batch script
.\script\build.bat

# Or build directly with Go
go build -o bin\zen.exe .\cmd\zen

# Add to PATH manually
# Copy zen.exe to a directory in your PATH, or
# Add the bin directory to your PATH environment variable
```

### 3. Verify Installation

```bash
# Unix-like systems
zen version

# Windows (if not in PATH)
.\bin\zen.exe version

# Expected output
# Zen CLI version 1.0.0
# Build: abc1234
# Date: 2024-03-15
```

## Cross-compiling binaries for different platforms

You can use any platform with Go installed to build a binary that is intended for another platform or CPU architecture. This is achieved by setting environment variables such as GOOS and GOARCH.

For example, to compile the `zen` binary for the 32-bit Raspberry Pi OS:

```sh
# on a Unix-like system:
$ GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 make clean bin/zen
```

```pwsh
# on Windows, pass environment variables as arguments to the build script:
> .\script\build.ps1 clean bin\zen GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0
# or
> go run script\build.go clean bin\zen GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0
```

Run `go tool dist list` to see all supported GOOS/GOARCH combinations.

### Build Optimization

Reduce binary size and add version information:

```bash
# Optimize for size
go build -ldflags="-s -w" -o bin/zen ./cmd/zen

# Include version information
go build -ldflags="-X main.version=1.0.0 -X main.commit=$(git rev-parse HEAD)" -o bin/zen ./cmd/zen

# Full production build
make build-prod
```

See [Go linker documentation](https://golang.org/cmd/link/) for all available flags.

## Development Builds

### Fast Development Build

```bash
# Quick build for testing
go build -o bin/zen ./cmd/zen

# Run without installing
./bin/zen --help
```

### Debug Build

```bash
# Build with debug symbols
go build -gcflags="all=-N -l" -o bin/zen ./cmd/zen

# Use with debugger
gdb ./bin/zen
# or
dlv exec ./bin/zen
```

## Troubleshooting

### Common Build Issues

#### Module Download Errors
```bash
# Clear module cache
go clean -modcache

# Download dependencies again
go mod download
```

#### Build Errors
```bash
# Ensure dependencies are up to date
go mod tidy

# Verify module integrity
go mod verify
```

#### Permission Errors on Install
```bash
# Use sudo for system-wide installation
sudo make install

# Or install to user directory
make install prefix=$HOME/.local
# Add $HOME/.local/bin to PATH
```

### Platform-Specific Issues

#### macOS Code Signing
```bash
# Remove quarantine attribute
xattr -d com.apple.quarantine ./bin/zen

# Or sign the binary
codesign -s - ./bin/zen
```

#### Windows Antivirus
Windows Defender may flag the new binary. Add an exception or wait for the scan to complete.

#### Linux SELinux
```bash
# Set correct context
chcon -t bin_t ./bin/zen
```

## Next Steps

- Return to the main [Installation Guide](../user-guide/installation.md)
- Continue with [Quick Start](../user-guide/quick-start.md)
- Set up your [Development Environment](../contributing/getting-started.md)
