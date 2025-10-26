---
title: "zen config get"
slug: "/cli/zen-config-get"
description: "CLI reference for zen config get"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-10-26
keywords:
  - zen
  - cli
  - zen-config-get
---

## zen config get

Print the value of a given configuration key

### Synopsis

Print the value of a configuration key.

Configuration keys use dot notation to access component values:
- log_level, log_format (core config)
- assets.repository_url, assets.branch
- workspace.root, workspace.zen_path
- cli.no_color, cli.verbose, cli.output_format
- development.debug, development.profile
- task.source, task.sync, task.project_key
- auth.storage_type, auth.validation_timeout
- cache.base_path, cache.size_limit_mb
- templates.cache_enabled, templates.cache_ttl

```
zen config get <key> [flags]
```

### Examples

```
$ zen config get log_level
$ zen config get assets.repository_url
$ zen config get workspace.root
$ zen config get cli.output_format

```

### Options

```
  -h, --help   help for get
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

