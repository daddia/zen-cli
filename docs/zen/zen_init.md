---
title: "zen init"
slug: "/cli/zen-init"
description: "CLI reference for zen init"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-10-27
keywords:
  - zen
  - cli
  - zen-init
---

## zen init

Initialize your new Zen workspace or reinitialize an existing one

### Synopsis

Initialize a new Zen workspace in the current directory.

This command creates a .zen/ directory structure and a zen.yaml configuration file
with default settings based on your project type. It automatically detects common
project types like Git repositories, Node.js, Go, Python, Rust, and Java projects.

Running 'zen init' in an existing workspace is safe and will reinitialize the workspace
without errors, similar to 'git init' behavior.

The .zen/ directory contains:
  - Configuration files
  - Library directory with manifest cache
  - Cache directory
  - Log directory
  - Templates directory (for future use)
  - Backups directory

If GitHub authentication is configured, zen init will automatically:
  - Set up the library infrastructure
  - Download the latest library manifest (if needed)
  - Make library available for immediate use

```
zen init [flags]
```

### Examples

```
  # Initialize in current directory (safe to run multiple times)
  zen init

  # Reinitialize existing workspace (safe operation like git init)
  zen init

  # Force reinitialize with backup of existing configuration
  zen init --force

  # Initialize with custom config file location
  zen init --config ./config/zen.yaml

  # Initialize with verbose output to see project detection
  zen init --verbose
```

### Options

```
  -c, --config string   Path to configuration file (default: zen.yaml)
  -f, --force           Overwrite existing configuration and create backup
  -h, --help            help for init
```

### Options inherited from parent commands

```
      --dry-run         Show what would be executed without making changes
      --no-color        Disable colored output
  -o, --output string   Output format (text, json, yaml) (default "text")
  -v, --verbose         Enable verbose output
```

### SEE ALSO

* [zen](zen.md.md)	 - AI-Powered Productivity Suite

