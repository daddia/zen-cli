---
title: "zen assets list"
slug: "/cli/zen-assets-list"
description: "CLI reference for zen assets list"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-19
keywords:
  - zen
  - cli
  - zen-assets-list
---

## zen assets list

List available assets

### Synopsis

List available assets with optional filtering.

Assets can be filtered by type, category, and tags. Results are paginated
to handle large asset repositories efficiently.

Asset Types:
- template: Reusable content templates with variables
- prompt: AI prompts for various tasks
- mcp: Model Context Protocol definitions
- schema: JSON/YAML schemas for validation

The list command works offline using cached asset metadata. Use 'zen assets sync'
to update the cache with the latest assets from the repository.

```
zen assets list [flags]
```

### Examples

```
  # List all assets
  zen assets list

  # List only templates
  zen assets list --type template

  # List assets in a specific category
  zen assets list --category documentation

  # List assets with specific tags
  zen assets list --tags ai,technical

  # Combine filters
  zen assets list --type prompt --category planning --tags sprint

  # Limit results and use pagination
  zen assets list --limit 10 --offset 20

  # Output as JSON
  zen assets list --output json
```

### Options

```
      --category string   Filter by category
  -h, --help              help for list
      --limit int         Maximum number of results (default 50)
      --offset int        Number of results to skip
      --tags strings      Filter by tags (comma-separated)
      --type string       Filter by asset type (template, prompt, mcp, schema)
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

