# Deployment View

## Overview

The Deployment view shows how Zen CLI is distributed, installed, and operated across different environments and platforms. This view addresses operational concerns and deployment scenarios.

## Deployment Architecture

```mermaid
graph TB
    subgraph "Distribution Channels"
        GH[GitHub Releases]
        Brew[Homebrew]
        Docker[Docker Hub]
        Go[Go Install]
        Manual[Manual Download]
    end
    
    subgraph "Target Platforms"
        subgraph "Linux"
            L64[Linux amd64]
            LARM[Linux arm64]
        end
        
        subgraph "macOS"
            M64[macOS amd64]
            MARM[macOS arm64]
        end
        
        subgraph "Windows"
            W64[Windows amd64]
        end
    end
    
    subgraph "Runtime Environment"
        Local[Local Machine]
        Container[Container]
        CI[CI/CD Pipeline]
        Cloud[Cloud Shell]
    end
    
    GH --> L64 & LARM & M64 & MARM & W64
    Brew --> M64 & MARM
    Docker --> Container
    Go --> Local
    Manual --> Local
    
    L64 & LARM --> Local & Container & CI & Cloud
    M64 & MARM --> Local
    W64 --> Local
```

## Distribution Models

### 1. Single Binary Distribution

```mermaid
graph LR
    subgraph "Build Process"
        Source[Go Source] --> Build[go build]
        Build --> Binary[zen binary]
        Binary --> Compress[Compression]
        Compress --> Archive[zen-platform.tar.gz]
    end
    
    subgraph "Embedded Assets"
        Templates[Templates]
        Schemas[Schemas]
        Defaults[Default Configs]
    end
    
    Templates & Schemas & Defaults --> Build
```

**Characteristics:**
- No runtime dependencies required
- All assets embedded at compile time
- Single file to distribute and manage
- ~45MB compressed, ~15MB after compression

### 2. Container Distribution

```dockerfile
# Multi-stage Docker build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o zen cmd/zen/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=builder /app/zen /usr/local/bin/
ENTRYPOINT ["zen"]
```

**Container Architecture:**
```mermaid
graph TD
    subgraph "Container Layers"
        Base[Alpine Base ~5MB]
        Certs[CA Certificates ~1MB]
        Binary[Zen Binary ~45MB]
        Config[Config Volume]
    end
    
    Base --> Certs --> Binary --> Config
```

### 3. Package Manager Distribution

#### Homebrew (macOS/Linux)
```bash
brew tap zen-cli/zen
brew install zen
```

#### Go Install
```bash
go install github.com/zen-org/zen/cmd/zen@latest
```

## Installation Locations

```mermaid
graph TB
    subgraph "System Paths"
        Bin[/usr/local/bin/zen<br/>Executable]
        Config[~/.zen/<br/>User Config]
        Cache[~/.cache/zen/<br/>Cache Data]
    end
    
    subgraph "Project Paths"
        Workspace[./zen.yaml<br/>Project Config]
        Generated[./generated/<br/>Output Files]
        Logs[./.zen/logs/<br/>Local Logs]
    end
    
    Bin --> Config
    Bin --> Workspace
    Config --> Cache
    Workspace --> Generated
    Workspace --> Logs
```

## Environment Configuration

### Development Environment
```mermaid
graph LR
    subgraph "Dev Setup"
        IDE[IDE/Editor] --> Source[Source Code]
        Source --> Build[Local Build]
        Build --> Test[Test Execution]
        Test --> Debug[Debug Mode]
    end
    
    subgraph "Dev Config"
        EnvVars[ZEN_LOG_LEVEL=debug<br/>ZEN_DEV_MODE=true]
        LocalConfig[zen.dev.yaml]
    end
    
    Build --> EnvVars
    Build --> LocalConfig
```

### Production Environment
```mermaid
graph LR
    subgraph "Prod Setup"
        Binary[Zen Binary]
        ProdConfig[Production Config]
        Secrets[Secret Manager]
    end
    
    subgraph "Monitoring"
        Logs[Log Aggregation]
        Metrics[Metrics Collection]
        Alerts[Alert System]
    end
    
    Binary --> ProdConfig
    Binary --> Secrets
    Binary --> Logs
    Binary --> Metrics
    Metrics --> Alerts
```

## CI/CD Integration

### Pipeline Deployment
```mermaid
sequenceDiagram
    participant Dev
    participant Git
    participant CI
    participant Registry
    participant Prod
    
    Dev->>Git: Push code
    Git->>CI: Trigger build
    CI->>CI: Run tests
    CI->>CI: Build binaries
    CI->>CI: Security scan
    CI->>Registry: Push artifacts
    Registry->>Prod: Deploy
    Prod->>Prod: Health check
```

### CI/CD Environments
```yaml
# GitHub Actions example
name: Deploy Zen CLI
on:
  push:
    tags: ['v*']

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu, macos, windows]
        arch: [amd64, arm64]
    steps:
      - name: Build
        run: |
          GOOS=${{ matrix.os }}
          GOARCH=${{ matrix.arch }}
          go build -o zen-${{ matrix.os }}-${{ matrix.arch }}
```

## Security Considerations

### Deployment Security
```mermaid
graph TB
    subgraph "Build Security"
        Scan[Dependency Scanning]
        Sign[Binary Signing]
        SBOM[SBOM Generation]
    end
    
    subgraph "Runtime Security"
        Perms[File Permissions<br/>0755 binary, 0600 config]
        Secrets[Secret Management<br/>No hardcoded values]
        TLS[TLS Communication<br/>External services]
    end
    
    subgraph "Update Security"
        Verify[Signature Verification]
        Rollback[Rollback Capability]
        Audit[Update Audit Log]
    end
    
    Scan --> Sign --> SBOM
    Perms --> Secrets --> TLS
    Verify --> Rollback --> Audit
```

## Platform-Specific Considerations

### Linux Deployment
- System service integration (systemd)
- Package manager integration (apt, yum, dnf)
- XDG base directory compliance
- SELinux/AppArmor profiles

### macOS Deployment
- Code signing requirements
- Gatekeeper compliance
- Keychain integration for secrets
- Homebrew formula maintenance

### Windows Deployment
- Code signing with Authenticode
- Windows Credential Store integration
- PowerShell completion support
- Chocolatey package maintenance

## Monitoring and Operations

### Health Checks
```mermaid
graph LR
    subgraph "Health Endpoints"
        Version[zen version<br/>Build info]
        Status[zen status<br/>System health]
        Validate[zen validate<br/>Config check]
    end
    
    subgraph "Monitoring Integration"
        Prometheus[Prometheus Metrics]
        Grafana[Grafana Dashboards]
        AlertManager[Alert Rules]
    end
    
    Version & Status & Validate --> Prometheus
    Prometheus --> Grafana
    Prometheus --> AlertManager
```

### Operational Metrics
- Binary version and build info
- Configuration validation status
- Integration connectivity
- API rate limit status
- Error rates and types
- Performance metrics (latency, throughput)

## Update Strategy

### Auto-Update Flow
```mermaid
stateDiagram-v2
    [*] --> CheckVersion
    CheckVersion --> NewAvailable: Update Found
    CheckVersion --> Current: Up-to-date
    NewAvailable --> Download
    Download --> Verify
    Verify --> Backup
    Backup --> Replace
    Replace --> Restart
    Restart --> [*]
    Current --> [*]
```

### Rolling Updates
- Backward compatibility guarantee
- Configuration migration support
- Graceful shutdown and restart
- Rollback capability
- Update notification system
