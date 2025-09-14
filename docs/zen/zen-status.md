---
title: zen-status
slug: /cli/zen-status
section: "CLI Reference"
man_section: 1
since: v0.0.0
aliases: []
see_also:
  - zen-init
  - zen-config
---

# zen status

```
zen status [flags]
```

Display comprehensive status information about your Zen workspace, configuration, system environment, and available integrations.

This command provides a detailed overview of the current state of your Zen installation and workspace, helping you troubleshoot issues and understand your environment.

## Options

`-h`, `--help`

Show help for command

### Options inherited from parent commands

`-v`, `--verbose`

Enable verbose output

`--no-color`

Disable colored output

`-o`, `--output <format>`

Output format (text, json, yaml)

## Examples

```bash
# Display status overview
$ zen status
Zen CLI Status

Workspace:
  Status:      ✓ Ready
  Path:        /Users/username/project
  Config File: .zen/config.yaml

Configuration:
  Status:    ✓ Loaded
  Source:    .zen/config.yaml
  Log Level: info

System:
  OS:           darwin
  Architecture: arm64
  Go Version:   go1.23.1
  CPU Cores:    8

Integrations:
  Available: [jira confluence git slack]
  Active:    []

# Output as JSON for scripting
$ zen status --output json
{
  "workspace": {
    "initialized": true,
    "path": "/Users/username/project",
    "config_file": ".zen/config.yaml"
  },
  "configuration": {
    "loaded": true,
    "source": ".zen/config.yaml",
    "log_level": "info"
  },
  "system": {
    "os": "darwin",
    "architecture": "arm64",
    "go_version": "go1.23.1",
    "num_cpu": 8
  },
  "integrations": {
    "available": ["jira", "confluence", "git", "slack"],
    "active": []
  }
}

# Output as YAML
$ zen status --output yaml
workspace:
  initialized: true
  path: /Users/username/project
  config_file: .zen/config.yaml
configuration:
  loaded: true
  source: .zen/config.yaml
  log_level: info
system:
  os: darwin
  architecture: arm64
  go_version: go1.23.1
  num_cpu: 8
integrations:
  available:
    - jira
    - confluence
    - git
    - slack
  active: []

# Check status when not initialized
$ zen status
Zen CLI Status

Workspace:
  Status:      ✗ Not Initialized
  Path:        /Users/username/project
  Config File: 

Configuration:
  Status:    ✗ Not Loaded
  Source:    defaults
  Log Level: info

System:
  OS:           darwin
  Architecture: arm64
  Go Version:   go1.23.1
  CPU Cores:    8

Integrations:
  Available: [jira confluence git slack]
  Active:    []
```

## Status Information

### Workspace Status

**Initialized**
- ✓ Ready: Workspace is properly initialized with `.zen/` directory
- ✗ Not Initialized: No workspace found, run `zen init` to create one

**Path**
- Current working directory or workspace root path
- Shows where Zen is looking for configuration and workspace files

**Config File**
- Path to the active configuration file
- Empty if no configuration file is found

### Configuration Status

**Loaded**
- ✓ Loaded: Configuration successfully loaded from file or defaults
- ✗ Not Loaded: Configuration could not be loaded (file errors, etc.)

**Source**
- Shows where configuration was loaded from:
  - `zen.yaml` - Local workspace configuration
  - `.zen/config.yaml` - Local workspace configuration directory
  - `~/.zen/config.yaml` - User home configuration
  - `defaults` - Built-in default values

**Log Level**
- Current logging verbosity level
- Affects how much detail is shown in command output

### System Information

**OS**
- Operating system (darwin, linux, windows)

**Architecture**
- CPU architecture (amd64, arm64, etc.)

**Go Version**
- Version of Go runtime used to build the CLI

**CPU Cores**
- Number of available CPU cores for parallel operations

### Integration Status

**Available**
- List of integrations that Zen can work with
- Currently includes: jira, confluence, git, slack

**Active**
- List of integrations currently configured and enabled
- Empty list indicates no integrations are configured yet

## Exit Status

`0` - Success  
`1` - General error  
`2` - Invalid arguments

## Notes

**Troubleshooting:**
- Use this command to diagnose workspace and configuration issues
- Check if workspace is properly initialized before running other commands
- Verify configuration source and values are as expected

**Machine-Readable Output:**
- JSON and YAML formats provide structured data for scripts
- Useful for monitoring and automation tools
- Status booleans are consistently represented as `true`/`false`

**Performance:**
- Command runs quickly as it only reads local files and system information
- No network calls or external dependencies required
- Safe to run frequently in scripts or monitoring

## See Also

* [zen init](zen-init.md) - Initialize a workspace
* [zen config](zen-config.md) - Manage configuration
* [zen config list](zen-config-list.md) - List configuration values
