---
title: zen-init
slug: /cli/zen-init
section: "CLI Reference"
man_section: 1
since: v0.0.0
aliases: []
see_also:
  - zen-config
  - zen-status
---

# zen init

```
zen init [flags]
```

Initialize a new Zen workspace in the current directory.

This command creates a `.zen/` directory structure and a `zen.yaml` configuration file with default settings based on your project type. It automatically detects common project types like Git repositories, Node.js, Go, Python, Rust, and Java projects.

The `.zen/` directory contains:
- Configuration files
- Cache directory  
- Log directory
- Templates directory (for future use)
- Backups directory

## Options

`-f`, `--force`

Overwrite existing configuration and create backup

`-c`, `--config <file>`

Path to configuration file (default: `zen.yaml`)

`-h`, `--help`

Show help for command

### Options inherited from parent commands

`-v`, `--verbose`

Enable verbose output

`--no-color`

Disable colored output

`-o`, `--output <format>`

Output format (text, json, yaml)

`--dry-run`

Show what would be executed without making changes

## Examples

```bash
# Initialize in current directory
$ zen init
Initialized empty Zen workspace in /Users/username/project/.zen/

# Reinitialize existing workspace (safe operation)
$ zen init
Reinitialized existing Zen workspace in /Users/username/project/.zen/

# Force reinitialize with backup of existing config
$ zen init --force
Reinitialized existing Zen workspace in /Users/username/project/.zen/
Configuration backup saved to .zen/backups/config-2024-01-15.yaml

# Initialize with custom config file location
$ zen init --config ./config/zen.yaml
Initialized empty Zen workspace in /Users/username/project/.zen/

# Initialize with verbose output to see project detection
$ zen init --verbose
Analyzing project in /Users/username/project...
Detected project type: Go module
Initialized empty Zen workspace in /Users/username/project/.zen/
```

## Exit Status

`0` - Success (workspace initialized or reinitialized)  
`1` - General error  
`2` - Invalid arguments  
`3` - Permission denied

## Notes

**Project Detection:**
- Automatically detects Git repositories, Node.js, Go, Python, Rust, and Java projects
- Creates optimized configuration based on detected project type
- Falls back to generic configuration if no specific type detected

**Directory Structure:**
The command creates the following structure:
```
.zen/
├── config.yaml      # Main configuration file
├── cache/           # Temporary cache files
├── logs/            # Command execution logs
├── templates/       # Template files (future use)
└── backups/         # Configuration backups
```

**Idempotent Operation:**
- Safe to run multiple times in the same directory
- Existing workspaces are reinitialized without errors
- Use `--force` to create backup and reset configuration
- Similar to `git init` behavior for familiar user experience

**Permissions:**
- Requires write permissions in the current directory
- Will attempt to create directories with mode `0755`
- Configuration files created with mode `0644`

## See Also

* [zen config](zen-config.md) - Manage configuration settings
* [zen status](zen-status.md) - Check workspace status
