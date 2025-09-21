---
title: "zen version"
slug: "/cli/zen-version"
description: "CLI reference for zen version"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-21
keywords:
  - zen
  - cli
  - zen-version
---

## zen version

Display version information

### Synopsis

Display the version information for Zen CLI.

By default, shows just the version number. Use --build-options to see detailed
build information including Git commit, build date, Go version, and platform.

```
zen version [flags]
```

### Examples

```
  # Display simple version
  zen version

  # Display detailed build information
  zen version --build-options

  # Output as JSON for scripting
  zen version --build-options --output json

  # Output as YAML
  zen version --build-options --output yaml

  # Check version in scripts
  zen version --build-options --output json | jq -r '.version'
```

### Options

```
      --build-options   Show detailed build information
  -h, --help            help for version
  -o, --output string   Output format (text, json, yaml) (default "text")
```

### Options inherited from parent commands

```
  -c, --config string   Path to configuration file
      --dry-run         Show what would be executed without making changes
      --no-color        Disable colored output
  -v, --verbose         Enable verbose output
```

### SEE ALSO

* [zen](zen.md.md)	 - AI-Powered Productivity Suite

