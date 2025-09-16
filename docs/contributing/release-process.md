# Release Process

This guide explains how Zen releases are created, versioned, and distributed.

## Versioning Strategy

We follow [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH

1.2.3
│ │ └── Patch: Bug fixes
│ └──── Minor: New features (backward compatible)
└────── Major: Breaking changes
```

### Version Bumps

- **Patch (1.0.0 → 1.0.1)**
  - Bug fixes
  - Security patches
  - Documentation updates
  
- **Minor (1.0.0 → 1.1.0)**
  - New features
  - New commands
  - Performance improvements
  
- **Major (1.0.0 → 2.0.0)**
  - Breaking API changes
  - Removed features
  - Major architecture changes

## Release Cycle

### Release Schedule

- **Major**: Annually (as needed)
- **Minor**: Monthly
- **Patch**: As needed (security/critical fixes)

### Release Branches

```bash
main          # Development branch
├── release/v1.2.x  # Release branch
└── tags/
    ├── v1.2.0     # Release tags
    ├── v1.2.1
    └── v1.2.2
```

## Release Process

### 1. Preparation

#### Create Release Branch
```bash
# Branch from main for minor/major releases
git checkout main
git pull origin main
git checkout -b release/v1.2.x

# For patches, branch from release branch
git checkout release/v1.1.x
git checkout -b fix/security-patch
```

#### Update Version

Update version in:
- `internal/version/version.go`
- `CHANGELOG.md`
- Documentation references

#### Update Changelog

Follow Keep a Changelog format:

```markdown
## [1.2.0] - 2024-03-15

### Added
- New feature X (#123)
- Command Y support (#456)

### Changed
- Improved performance of Z (#789)

### Fixed
- Bug in configuration loading (#234)

### Security
- Updated dependency to fix CVE-XXXX-XXXX
```

### 2. Testing

#### Pre-release Testing

```bash
# Full test suite
make test-all

# Build all platforms
make build-all

# Smoke tests
./bin/zen version
./bin/zen init test-project
./bin/zen status
```

#### Release Candidate

For major/minor releases:

```bash
# Tag release candidate
git tag -a v1.2.0-rc.1 -m "Release candidate 1 for v1.2.0"
git push origin v1.2.0-rc.1

# Test RC in staging environment
# Gather feedback from early adopters
```

### 3. Release

#### Create Release Tag

```bash
# Tag the release
git tag -a v1.2.0 -m "Release version 1.2.0"

# Push tag
git push origin v1.2.0
```

#### GitHub Release

1. Go to GitHub Releases page
2. Click "Draft a new release"
3. Select the tag
4. Set as pre-release if applicable
5. Generate release notes
6. Upload artifacts
7. Publish release

### 4. Distribution

#### Binary Distribution

Automated via GitHub Actions:

```yaml
# .github/workflows/release.yml
on:
  push:
    tags:
      - 'v*'
```

Platforms supported:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

#### Package Managers

```bash
# Homebrew (macOS/Linux)
brew install zen

# Snap (Linux)
snap install zen

# Scoop (Windows)
scoop install zen

# Go install
go install github.com/zen-org/zen@latest
```

### 5. Post-Release

#### Announcement

- Update website/documentation
- Post release notes
- Notify users via:
  - GitHub Releases
  - Discord/Slack
  - Twitter/Social media
  - Email newsletter

#### Monitoring

- Monitor issue tracker
- Check download metrics
- Gather user feedback
- Track adoption rates

## Hotfix Process

For critical fixes:

### 1. Create Hotfix Branch
```bash
git checkout -b hotfix/v1.2.1 tags/v1.2.0
```

### 2. Apply Fix
```bash
# Make minimal changes
git commit -m "fix: critical security issue"
```

### 3. Fast-track Release
```bash
# Tag immediately
git tag -a v1.2.1 -m "Hotfix release v1.2.1"
git push origin v1.2.1
```

### 4. Backport to Main
```bash
git checkout main
git cherry-pick <commit-sha>
```

## Release Artifacts

### Build Artifacts

Generated artifacts:
```
dist/
├── zen-darwin-amd64.tar.gz
├── zen-darwin-arm64.tar.gz
├── zen-linux-amd64.tar.gz
├── zen-linux-arm64.tar.gz
├── zen-windows-amd64.zip
├── checksums.txt
└── checksums.txt.sig
```

### Docker Images

```bash
# Build and push Docker images
docker build -t zen:v1.2.0 .
docker tag zen:v1.2.0 zen:latest
docker push zen:v1.2.0
docker push zen:latest
```

## Automation

### CI/CD Pipeline

Release automation via GitHub Actions:

```yaml
name: Release
on:
  push:
    tags: ['v*']

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make build-all
      - uses: goreleaser/goreleaser-action@v4
```

### GoReleaser Configuration

```yaml
# .goreleaser.yml
builds:
  - binary: zen
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

release:
  github:
    owner: zen-org
    name: zen
```

## Deprecation Policy

### Deprecation Process

1. **Announce** - Note in release notes
2. **Warn** - Show runtime warnings
3. **Maintain** - Support for 2 minor versions
4. **Remove** - Delete in next major version

### Deprecation Notice

```go
// Deprecated: Use NewConfig instead. Will be removed in v2.0.0
func OldConfig() *Config {
    log.Warn("OldConfig is deprecated, use NewConfig")
    return NewConfig()
}
```

## Security Releases

### Security Process

1. Receive vulnerability report
2. Confirm and assess severity
3. Develop fix in private
4. Coordinate disclosure
5. Release patch
6. Public announcement

### Security Advisory

```markdown
# Security Advisory: ZEN-2024-001

## Summary
Brief description of vulnerability

## Affected Versions
- v1.0.0 - v1.2.0

## Fixed Versions
- v1.2.1

## Mitigation
Steps to mitigate if unable to upgrade
```

## Release Checklist

### Pre-release
- [ ] Version bumped
- [ ] CHANGELOG updated
- [ ] Documentation updated
- [ ] Tests passing
- [ ] Release notes drafted

### Release
- [ ] Tag created and pushed
- [ ] GitHub Release published
- [ ] Binaries uploaded
- [ ] Docker images pushed

### Post-release
- [ ] Website updated
- [ ] Announcements sent
- [ ] Package managers updated
- [ ] Monitoring in place

## Troubleshooting

### Common Issues

#### Failed Builds
```bash
# Check build locally
make build-all

# Verify Go version
go version
```

#### Missing Artifacts
```bash
# Rebuild with goreleaser
goreleaser release --snapshot --clean
```

#### Tag Issues
```bash
# Delete local tag
git tag -d v1.2.0

# Delete remote tag
git push --delete origin v1.2.0

# Recreate tag
git tag -a v1.2.0 -m "Release v1.2.0"
```

## Next Steps

- Review [Architecture](architecture.md) for system design
- Understand [Development Workflow](development-workflow.md)
- Check [Testing](testing.md) requirements

---

Questions? Contact release coordinators or check release documentation.
