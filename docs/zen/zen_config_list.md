---
title: "zen config list"
slug: "/cli/zen-config-list"
description: "CLI reference for zen config list"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-22
keywords:
  - zen
  - cli
  - zen-config-list
---

## zen config list

Print a list of configuration keys and values

### Synopsis

List all configuration keys and their current values.

This shows the effective configuration after loading from files,
environment variables, and command-line flags.

```
zen config list [flags]
```

### Examples

```
  zen zen config list
```

### Options

```
  -h, --help   help for list
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

* [zen config](zen-config.md.md)	 - Manage configuration for Zen CLI

