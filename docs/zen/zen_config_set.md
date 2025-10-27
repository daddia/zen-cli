---
title: "zen config set"
slug: "/cli/zen-config-set"
description: "CLI reference for zen config set"
section: "CLI Reference"
man_section: 1
since: v0.0.0
keywords:
  - zen
  - cli
---

## zen config set

Update configuration with a value for the given key

### Synopsis

Set a configuration value for the given key.

Configuration keys use dot notation to access nested values:
- log_level (trace, debug, info, warn, error, fatal, panic)
- log_format (text, json)
- cli.no_color (true, false)
- cli.verbose (true, false)
- cli.output_format (text, json, yaml)
- workspace.root (directory path)
- workspace.config_file (filename)
- development.debug (true, false)
- development.profile (true, false)

The configuration is saved to the first available location:
1. .zen/config.yaml (current directory)
2. ~/.zen/config.yaml (user home directory)

```
zen config set <key> <value> [flags]
```

### Examples

```
$ zen config set log_level debug
$ zen config set cli.output_format json
$ zen config set cli.no_color true
$ zen config set workspace.root /path/to/workspace

```

### Options

```
  -h, --help   help for set
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

