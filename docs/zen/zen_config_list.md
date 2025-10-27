---
title: "zen config list"
slug: "/cli/zen-config-list"
description: "CLI reference for zen config list"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-10-27
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

Configuration is organized by component:
- Core: log_level, log_format
- Assets: repository_url, branch, cache settings
- Workspace: root, zen_path
- CLI: no_color, verbose, output_format
- Development: debug, profile
- Task: source, sync, project_key
- Auth: storage_type, validation_timeout
- Cache: base_path, size_limit_mb
- Templates: cache_enabled, cache_ttl

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

