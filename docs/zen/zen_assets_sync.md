---
title: "zen assets sync"
slug: "/cli/zen-assets-sync"
description: "CLI reference for zen assets sync"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-20
keywords:
  - zen
  - cli
  - zen-assets-sync
---

## zen assets sync

Synchronize assets with remote repository

### Synopsis

Synchronize the local asset metadata with the remote repository.

This command fetches the latest asset manifest (catalog) from the configured
Git repository and updates the local metadata cache. It does NOT download
actual asset content - assets are downloaded on-demand when requested.

What gets synchronized:
- Asset manifest (manifest.yaml) - lightweight metadata only
- Asset descriptions, categories, tags, and checksums
- Asset availability and version information

Actual asset content is downloaded only when you use 'zen assets get <name>'.
This keeps sync operations fast and minimizes network/disk usage.

The sync operation requires authentication with the Git provider.
Use 'zen assets auth' to configure authentication first.

```
zen assets sync [flags]
```

### Examples

```
  # Synchronize asset metadata (manifest only)
  zen assets sync

  # Force refresh of cached metadata
  zen assets sync --force

  # Sync from a specific branch
  zen assets sync --branch develop

  # Sync with custom timeout
  zen assets sync --timeout 60

  # Output sync results as JSON
  zen assets sync --output json

  # After sync, list available assets
  zen assets list

  # Download specific asset content on-demand
  zen assets get technical-spec
```

### Options

```
      --branch string   Branch to synchronize (default "main")
      --force           Force refresh of cached metadata
  -h, --help            help for sync
      --timeout int     Timeout in seconds for sync operation (default 60)
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

