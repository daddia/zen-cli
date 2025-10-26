---
title: "zen assets list"
slug: "/cli/zen-assets-list"
description: "CLI reference for zen assets list"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-10-26
keywords:
  - zen
  - cli
  - zen-assets-list
---

## zen assets list

List available assets

### Synopsis

List available activities with optional filtering.

This command reads from the local manifest file (.zen/library/manifest.yaml) for fast,
offline activity discovery. The manifest contains metadata about all available activities
without storing the actual content locally.

Activities can be filtered by category and tags. Results are paginated
to handle large activity repositories efficiently.

Each activity represents a workflow step with associated templates and prompts
for generating documentation, code, or configurations.

The list command works offline using the local manifest. Use 'zen assets sync'
to update the manifest with the latest activities from the repository.

```
zen assets list [flags]
```

### Examples

```
  # List all activities
  zen assets list

  # List activities by type
  zen assets list --type template

  # List activities in a specific category
  zen assets list --category development

  # List activities with specific tags
  zen assets list --tags api,design

  # Combine filters
  zen assets list --type template --category planning --tags strategy

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
      --type string       Filter by asset type (template|prompt|mcp|schema)
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

