---
title: "zen assets"
slug: "/cli/zen-assets"
description: "CLI reference for zen assets"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-19
keywords:
  - zen
  - cli
  - zen-assets
---

## zen assets

Manage assets and templates

### Synopsis

Manage assets and templates for Zen CLI.

Assets include templates, prompts, and other reusable content stored in
Git repositories. This command provides authentication, discovery, and
synchronization capabilities for asset management.

Authentication:
  Assets are stored in Git repositories that may require authentication.
  Use 'zen assets auth' to configure authentication with GitHub or GitLab.

Discovery:
  List available assets with filtering and search capabilities.
  Get detailed information about specific assets including metadata.

Synchronization:
  Keep your local asset cache synchronized with remote repositories.
  Assets are cached locally for fast access and offline usage.

### Examples

```
  # Configure authentication with GitHub
  zen assets auth github

  # Check authentication and cache status
  zen assets status

  # List all available assets
  zen assets list

  # List only template assets
  zen assets list --type template

  # Get detailed information about a specific asset
  zen assets info technical-spec

  # Synchronize with remote repository
  zen assets sync

  # Force a full synchronization
  zen assets sync --force
```

### Options

```
  -h, --help   help for assets
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
* [zen assets auth](zen-assets-auth.md.md)	 - Authenticate with Git providers for asset access
* [zen assets info](zen-assets-info.md.md)	 - Show detailed information about an asset
* [zen assets list](zen-assets-list.md.md)	 - List available assets
* [zen assets status](zen-assets-status.md.md)	 - Show authentication and cache status
* [zen assets sync](zen-assets-sync.md.md)	 - Synchronize assets with remote repository

