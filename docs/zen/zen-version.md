---
title: zen-version
slug: /cli/zen-version
section: "CLI Reference"
man_section: 1
since: v0.0.0
aliases: []
see_also:
  - zen-status
---

# zen version

```
zen version [flags]
```

Display the version, build information, and platform details for Zen CLI.

This command shows comprehensive version information including the release version, build details, Git commit hash, build date, Go version used for compilation, and target platform.

## Options

`-o`, `--output <format>`

Output format (text, json, yaml)

`-h`, `--help`

Show help for command

### Options inherited from parent commands

`-v`, `--verbose`

Enable verbose output

`--no-color`

Disable colored output

## Examples

```bash
# Display version information
$ zen version
zen version v0.1.0
Build: 2024-01-15T10:30:00Z
Commit: a1b2c3d4e5f6789012345678901234567890abcd
Built: 2024-01-15T10:30:00Z
Go: go1.23.1
Platform: darwin/arm64

# Output as JSON for scripting
$ zen version --output json
{
  "version": "v0.1.0",
  "git_commit": "a1b2c3d4e5f6789012345678901234567890abcd",
  "build_date": "2024-01-15T10:30:00Z",
  "go_version": "go1.23.1",
  "platform": "darwin/arm64"
}

# Output as YAML
$ zen version --output yaml
version: v0.1.0
git_commit: a1b2c3d4e5f6789012345678901234567890abcd
build_date: "2024-01-15T10:30:00Z"
go_version: go1.23.1
platform: darwin/arm64

# Check version in scripts
$ zen version --output json | jq -r '.version'
v0.1.0
```

## Version Information

### Version

The semantic version of the Zen CLI release (e.g., `v0.1.0`, `v1.2.3`)

### Git Commit

The full SHA hash of the Git commit used to build this version. Useful for:
- Identifying exact source code version
- Tracking down specific builds
- Debugging and support purposes

### Build Date

ISO 8601 timestamp of when the binary was compiled. Shows:
- When the build was created
- Useful for determining build freshness
- Helps with support and debugging

### Go Version

Version of the Go compiler used to build the binary:
- Shows Go runtime compatibility
- Important for debugging Go-specific issues
- Indicates supported language features

### Platform

Target platform in `OS/Architecture` format:
- `darwin/arm64` - macOS on Apple Silicon
- `darwin/amd64` - macOS on Intel
- `linux/amd64` - Linux on x86_64
- `linux/arm64` - Linux on ARM64
- `windows/amd64` - Windows on x86_64

## Exit Status

`0` - Success  
`1` - General error  
`2` - Invalid arguments

## Notes

**Build Information:**
- All builds include complete version metadata
- Official releases have clean version numbers (e.g., `v1.0.0`)
- Development builds may include commit suffixes (e.g., `v1.0.0-dev.a1b2c3d`)

**Machine-Readable Output:**
- JSON format provides structured data for automation
- YAML format for configuration management tools
- Consistent field names across output formats

**Version Checking:**
- Use in scripts to verify minimum version requirements
- Compare against latest releases for update notifications
- Include in bug reports and support requests

**Platform Detection:**
- Platform string matches Go's `GOOS/GOARCH` format
- Useful for determining OS-specific behavior
- Helps with platform-specific troubleshooting

## See Also

* [zen status](zen-status.md) - Display system and workspace status
