---
title: "zen assets sync"
slug: "/cli/zen-assets-sync"
description: "CLI reference for zen assets sync"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-19
keywords:
  - zen
  - cli
  - zen-assets-sync
---

## zen assets sync

Synchronize assets with remote repository

### Synopsis

Synchronize the local asset cache with the remote repository.

This command fetches the latest assets from the configured Git repository
and updates the local cache. It performs authentication, downloads new or
updated assets, and removes assets that no longer exist in the repository.

Synchronization modes:
- Normal sync: Incremental update (git pull)
- Force sync: Full re-download (git clone --force)
- Shallow sync: Download only the latest commit (faster)

The sync operation requires authentication with the Git provider.
Use 'zen assets auth' to configure authentication first.

```
zen assets sync [flags]
```

### Examples

```
  # Normal synchronization
  zen assets sync

  # Force a complete re-synchronization
  zen assets sync --force

  # Shallow sync (faster, latest commit only)
  zen assets sync --shallow

  # Sync from a specific branch
  zen assets sync --branch develop

  # Sync with custom timeout
  zen assets sync --timeout 600

  # Output sync results as JSON
  zen assets sync --output json
```

### Options

```
      --branch string   Branch to synchronize (default "main")
      --force           Force complete re-synchronization
  -h, --help            help for sync
      --shallow         Perform shallow sync (latest commit only)
      --timeout int     Timeout in seconds for sync operation (default 300)
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

