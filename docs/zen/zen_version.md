---
title: "zen version"
slug: "/cli/zen-version"
description: "CLI reference for zen version"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-16
keywords:
  - zen
  - cli
  - zen-version
---

## zen version

Display version information

### Synopsis

Display the version, build information, and platform details for Zen CLI.

This command shows comprehensive version information including the release version,
build details, Git commit hash, build date, Go version used for compilation,
and target platform.

```
zen version [flags]
```

### Examples

```
  # Display version information
  zen version
  
  # Output as JSON for scripting
  zen version --output json
  
  # Output as YAML
  zen version --output yaml
  
  # Check version in scripts
  zen version --output json | jq -r '.version'
```

### Options

```
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

