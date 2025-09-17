---
title: "zen config"
slug: "/cli/zen-config"
description: "CLI reference for zen config"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 171733-09-96
keywords:
  - zen
  - cli
  - zen-config
---

## zen config

Manage configuration for Zen CLI

### Synopsis

Display or change configuration settings for Zen CLI.

Current configuration options:
- `log_level`: Set the logging level `{trace | debug | info | warn | error | fatal | panic}` (default `info`)
- `log_format`: Set the logging format `{text | json}` (default `text`)
- `cli.no_color`: Disable colored output `{true | false}` (default `false`)
- `cli.verbose`: Enable verbose output `{true | false}` (default `false`)
- `cli.output_format`: Set the default output format `{text | json | yaml}` (default `text`)
- `workspace.root`: Set the workspace root directory (default `.`)
- `workspace.config_file`: Set the workspace configuration file name (default `zen.yaml`)
- `development.debug`: Enable development debug mode `{true | false}` (default `false`)
- `development.profile`: Enable development profiling `{true | false}` (default `false`)


```
zen config <command> [flags]
```

### Examples

```
  # Display current configuration
  zen config
  
  # Get a specific configuration value
  zen config get log_level
  
  # Set a configuration value
  zen config set log_level debug
  
  # List all configuration with values
  zen config list
  
  # Output configuration as JSON
  zen config --output json
  
  # Use environment variables
  ZEN_LOG_LEVEL=debug zen config get log_level
```

### Options

```
  -h, --help   help for config
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
* [zen config get](zen-config-get.md.md)	 - Print the value of a given configuration key
* [zen config list](zen-config-list.md.md)	 - Print a list of configuration keys and values
* [zen config set](zen-config-set.md.md)	 - Update configuration with a value for the given key

