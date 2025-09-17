---
title: "zen assets status"
slug: "/cli/zen-assets-status"
description: "CLI reference for zen assets status"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 171733-09-96
keywords:
  - zen
  - cli
  - zen-assets-status
---

## zen assets status

Show authentication and cache status

### Synopsis

Display the current status of asset authentication, cache, and repository.

This command shows:
- Authentication status for configured Git providers
- Local cache status including size and hit ratio
- Repository synchronization status
- Asset availability (online/offline mode)

The status information helps troubleshoot authentication issues,
monitor cache performance, and understand asset availability.

```
zen assets status [flags]
```

### Examples

```
  # Show status in default text format
  zen assets status

  # Show status in JSON format
  zen assets status --output json

  # Show status in YAML format  
  zen assets status --output yaml
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

* [zen assets](zen-assets.md.md)	 - Manage assets and templates

