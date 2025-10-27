---
title: "zen status"
slug: "/cli/zen-status"
description: "CLI reference for zen status"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-10-27
keywords:
  - zen
  - cli
  - zen-status
---

## zen status

Display workspace and system status

### Synopsis

Display comprehensive status information about your Zen workspace,
configuration, system environment, and available integrations.

This command provides a detailed overview of the current state of your Zen installation
and workspace, helping you troubleshoot issues and understand your environment.

```
zen status [flags]
```

### Examples

```
  # Display status overview
  zen status

  # Output status as JSON for scripting
  zen status --output json

  # Output status as YAML
  zen status --output yaml

  # Check status with verbose output
  zen status --verbose
```

### Options

```
  -h, --help   help for status
```

### Options inherited from parent commands

```
  -c, --config string   Path to configuration file
      --dry-run         Show what would be executed without making changes
      --no-color        Disable colored output
  -o, --output string   Output format (text, json, yaml) (default "text")
  -v, --verbose         Enable verbose output
```

### SEE ALSO

* [zen](zen.md.md)	 - AI-Powered Productivity Suite

