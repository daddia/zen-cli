---
title: "zen assets info"
slug: "/cli/zen-assets-info"
description: "CLI reference for zen assets info"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-21
keywords:
  - zen
  - cli
  - zen-assets-info
---

## zen assets info

Show detailed information about an asset

### Synopsis

Display detailed information about a specific asset.

This command shows comprehensive metadata about an asset including:
- Basic information (name, type, description)
- Categorization (category, tags)
- Template variables (for template assets)
- File information (size, checksum, last updated)
- Cache status and integrity

The asset content can optionally be included in the output using
the --include-content flag.

```
zen assets info <asset-name> [flags]
```

### Examples

```
  # Show asset information
  zen assets info technical-spec

  # Include the asset content in output
  zen assets info user-story --include-content

  # Output as JSON with content
  zen assets info technical-spec --output json --include-content

  # Skip integrity verification for faster response
  zen assets info large-template --no-verify
```

### Options

```
  -h, --help              help for info
      --include-content   Include asset content in output
      --verify            Verify asset integrity (default true)
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

