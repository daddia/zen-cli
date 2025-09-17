# Installing Zen on Windows

This guide provides detailed instructions for installing Zen CLI on Windows systems.

## System Requirements

### Minimum Requirements
- **Windows Version**: Windows 10 version 1809 or later, Windows 11, or Windows Server 2019+
- **Architecture**: x64 (AMD64) or ARM64
- **Memory**: 4 GB RAM
- **Disk Space**: 500 MB available space
- **Network**: Internet connection for AI features

### Recommended Requirements
- **PowerShell**: 5.1 or PowerShell 7+ for optimal experience
- **Windows Terminal**: For enhanced command-line experience
- **Git for Windows**: For version control integration

## Installation Methods

### Method 1: Package Manager (Recommended)

#### Scoop (Recommended)

Scoop is a command-line installer for Windows that makes managing software easy.

**Install Scoop first (if not already installed):**
```powershell
# Set execution policy
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Install Scoop
Invoke-RestMethod -Uri https://get.scoop.sh | Invoke-Expression
```

**Install Zen:**
```powershell
# Install Zen
scoop install zen

# Verify installation
zen version
```

**Update Zen:**
```powershell
# Update to latest version
scoop update zen
```

#### Chocolatey

If you prefer Chocolatey:

```powershell
# Install Zen
choco install zen

# Verify installation
zen version

# Update Zen
choco upgrade zen
```

#### WinGet

Using Windows Package Manager:

```powershell
# Install Zen
winget install zen

# Verify installation
zen version

# Update Zen
winget upgrade zen
```

### Method 2: Pre-built Binary

#### Download and Install

1. **Download the latest release:**
   - Visit [https://github.com/zen-org/zen/releases](https://github.com/zen-org/zen/releases)
   - Download `zen-windows-amd64.zip` (or `zen-windows-arm64.zip` for ARM64)

2. **Extract and install:**

   **Option A: System-wide installation (requires Administrator)**
   ```powershell
   # Run PowerShell as Administrator
   # Extract to Program Files
   Expand-Archive -Path "zen-windows-amd64.zip" -DestinationPath "C:\Program Files\zen"
   
   # Add to system PATH
   $env:Path += ";C:\Program Files\zen"
   [Environment]::SetEnvironmentVariable("Path", $env:Path, [EnvironmentVariableTarget]::Machine)
   ```

   **Option B: User installation (no Administrator required)**
   ```powershell
   # Create user directory
   New-Item -ItemType Directory -Path "$env:USERPROFILE\.zen" -Force
   
   # Extract binary
   Expand-Archive -Path "zen-windows-amd64.zip" -DestinationPath "$env:USERPROFILE\.zen"
   
   # Add to user PATH
   $userPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
   [Environment]::SetEnvironmentVariable("Path", "$userPath;$env:USERPROFILE\.zen", [EnvironmentVariableTarget]::User)
   
   # Refresh current session
   $env:Path += ";$env:USERPROFILE\.zen"
   ```

3. **Verify installation:**
   ```powershell
   # Restart PowerShell or Command Prompt
   zen version
   ```

### Method 3: Build from Source

For developers or users needing the latest development version:

#### Prerequisites
- **Go 1.25+**: Download from [https://golang.org/dl/](https://golang.org/dl/)
- **Git for Windows**: Download from [https://git-scm.com/download/win](https://git-scm.com/download/win)
- **Make for Windows** (optional): Install via Chocolatey with `choco install make`

#### Build Steps

```powershell
# Clone repository
git clone https://github.com/zen-org/zen.git
cd zen

# Build using the build script
go run script\build.go

# Or use the batch script
.\script\build.bat

# Or build directly with Go
go build -o bin\zen.exe .\cmd\zen

# Install to user directory
New-Item -ItemType Directory -Path "$env:USERPROFILE\.zen" -Force
Copy-Item "bin\zen.exe" -Destination "$env:USERPROFILE\.zen\zen.exe"

# Add to PATH (if not already added)
$userPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
if ($userPath -notlike "*$env:USERPROFILE\.zen*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$env:USERPROFILE\.zen", [EnvironmentVariableTarget]::User)
}
```

For detailed source build instructions, see [Building from Source](install-source.md).

### Method 4: Docker

Run Zen in a container without local installation:

```powershell
# Install Docker Desktop for Windows first
# Then run Zen in container

# Run latest version
docker run --rm -it zen:latest

# Run with local directory mounted (PowerShell)
docker run --rm -it -v ${PWD}:/workspace zen:latest

# Run with local directory mounted (Command Prompt)
docker run --rm -it -v %cd%:/workspace zen:latest

# Create PowerShell alias for convenience
Set-Alias zen 'docker run --rm -it -v ${PWD}:/workspace zen:latest'
```

## Post-Installation Setup

### Verify Installation

```powershell
# Check version
zen version

# View available commands
zen --help

# Run system check
zen status
```

### Shell Completion

#### PowerShell

Add completion to your PowerShell profile:

```powershell
# Find your profile location
$PROFILE

# Create profile if it doesn't exist
if (!(Test-Path -Path $PROFILE)) {
    New-Item -ItemType File -Path $PROFILE -Force
}

# Add completion to profile
Add-Content -Path $PROFILE -Value "zen completion powershell | Out-String | Invoke-Expression"

# Reload profile
. $PROFILE
```

#### Command Prompt

For Command Prompt users, consider switching to PowerShell or Windows Terminal for better experience.

### Environment Variables

Configure Zen using environment variables:

```powershell
# Set configuration directory
[Environment]::SetEnvironmentVariable("ZEN_CONFIG_DIR", "$env:APPDATA\zen", [EnvironmentVariableTarget]::User)

# Set log level
[Environment]::SetEnvironmentVariable("ZEN_LOG_LEVEL", "info", [EnvironmentVariableTarget]::User)

# Set API keys for AI providers (replace with your actual keys)
[Environment]::SetEnvironmentVariable("ZEN_OPENAI_API_KEY", "your-openai-key-here", [EnvironmentVariableTarget]::User)
[Environment]::SetEnvironmentVariable("ZEN_ANTHROPIC_API_KEY", "your-anthropic-key-here", [EnvironmentVariableTarget]::User)
```

### Windows Terminal Integration

For the best experience, use Windows Terminal:

1. Install from Microsoft Store or GitHub
2. Configure a Zen profile (optional):
   ```json
   {
       "name": "Zen CLI",
       "commandline": "powershell.exe -NoExit -Command \"zen\"",
       "icon": "path/to/zen/icon.png"
   }
   ```

## Troubleshooting

### Common Issues

#### Command Not Found

**Problem:** `'zen' is not recognized as an internal or external command`

**Solutions:**
```powershell
# Check if zen.exe exists in expected location
Test-Path "$env:USERPROFILE\.zen\zen.exe"
Test-Path "C:\Program Files\zen\zen.exe"

# Check current PATH
$env:Path -split ';'

# Add to PATH temporarily
$env:Path += ";$env:USERPROFILE\.zen"

# Add to PATH permanently (user level)
$userPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
[Environment]::SetEnvironmentVariable("Path", "$userPath;$env:USERPROFILE\.zen", [EnvironmentVariableTarget]::User)

# Restart PowerShell/Command Prompt
```

#### Execution Policy Restrictions

**Problem:** PowerShell execution policy prevents running scripts

**Solution:**
```powershell
# Check current policy
Get-ExecutionPolicy

# Set policy for current user (recommended)
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser

# Or set for current session only
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process
```

#### Windows Defender / Antivirus Warnings

**Problem:** Antivirus software flags the zen.exe binary

**Solutions:**
1. **Wait for scan to complete** - First-time scans are normal
2. **Add exclusion:**
   - Open Windows Security
   - Go to Virus & threat protection
   - Add an exclusion for the zen installation directory
3. **Download from official releases only** to avoid false positives

#### Permission Denied Errors

**Problem:** Access denied when installing system-wide

**Solutions:**
```powershell
# Run PowerShell as Administrator
# Right-click PowerShell icon → "Run as Administrator"

# Or use user-level installation instead
# Follow "Option B" in the binary installation method
```

#### SSL/TLS Certificate Errors

**Problem:** SSL certificate verification failures

**Solutions:**
```powershell
# Update Windows
# Install latest updates through Windows Update

# Update certificates
certlm.msc  # Open certificate manager and refresh

# Or temporarily bypass (not recommended for production)
[Environment]::SetEnvironmentVariable("ZEN_SKIP_TLS_VERIFY", "true", [EnvironmentVariableTarget]::User)
```

#### Proxy Configuration

**Problem:** Corporate proxy blocking connections

**Solutions:**
```powershell
# Set proxy environment variables
[Environment]::SetEnvironmentVariable("HTTP_PROXY", "http://proxy.company.com:8080", [EnvironmentVariableTarget]::User)
[Environment]::SetEnvironmentVariable("HTTPS_PROXY", "http://proxy.company.com:8080", [EnvironmentVariableTarget]::User)

# Or configure in zen config
zen config set proxy.http "http://proxy.company.com:8080"
zen config set proxy.https "http://proxy.company.com:8080"
```

### Windows-Specific Considerations

#### Windows Subsystem for Linux (WSL)

If you're using WSL, you can install the Linux version:

```bash
# In WSL terminal
curl -fsSL https://install.zen.sh | sh
# or
wget -qO- https://install.zen.sh | sh
```

#### File Path Length Limitations

Windows has path length limitations. If you encounter issues:

1. **Enable long path support:**
   ```powershell
   # Run as Administrator
   New-ItemProperty -Path "HKLM:\SYSTEM\CurrentControlSet\Control\FileSystem" -Name "LongPathsEnabled" -Value 1 -PropertyType DWORD -Force
   ```

2. **Use shorter installation paths**

#### Performance Considerations

- **Windows Defender Real-time Protection** may slow first runs
- **Consider SSD storage** for better performance
- **Close unnecessary background applications** during intensive operations

## Uninstallation

### Package Managers

```powershell
# Scoop
scoop uninstall zen

# Chocolatey
choco uninstall zen

# WinGet
winget uninstall zen
```

### Manual Uninstallation

```powershell
# Remove binary
Remove-Item -Path "$env:USERPROFILE\.zen\zen.exe" -Force
# or
Remove-Item -Path "C:\Program Files\zen\zen.exe" -Force

# Remove from PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
$newPath = $userPath -replace ";$env:USERPROFILE\\\.zen", ""
[Environment]::SetEnvironmentVariable("Path", $newPath, [EnvironmentVariableTarget]::User)

# Remove configuration (optional)
Remove-Item -Path "$env:APPDATA\zen" -Recurse -Force
Remove-Item -Path "$env:USERPROFILE\.zen" -Recurse -Force
```

## Getting Help

If you encounter issues not covered here:

1. **Check the main [Installation Guide](../getting-started/installation.md)**
2. **Review [Windows-specific troubleshooting](../getting-started/installation.md#windows)**
3. **Search [existing issues](https://github.com/zen-org/zen/issues)**
4. **Open a [new issue](https://github.com/zen-org/zen/issues/new)** with:
   - Windows version (`winver` command output)
   - PowerShell version (`$PSVersionTable.PSVersion`)
   - Installation method attempted
   - Complete error messages
   - Output of `zen version` (if zen is partially working)

## Next Steps

- **[Quick Start Guide](../getting-started/quick-start.md)** - Learn basic Zen commands
- **[Configuration Guide](../getting-started/configuration.md)** - Customize Zen for your workflow
- **[Windows Terminal Setup](../getting-started/terminal-setup.md)** - Optimize your terminal experience

---

Return to [Installation Documentation](README.md) | [Getting Started](../getting-started/README.md)
