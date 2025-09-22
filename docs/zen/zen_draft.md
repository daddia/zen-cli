---
title: "zen draft"
slug: "/cli/zen-draft"
description: "CLI reference for zen draft"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-22
keywords:
  - zen
  - cli
  - zen-draft
---

## zen draft

Generate document templates with task data

### Synopsis

Generate document templates populated with task manifest data.

This command fetches Go templates from the zen-assets repository and processes
them using data from the current task's manifest.yaml file.

Examples:
  # Generate a feature specification
  zen draft feature-spec

  # Preview template before generating
  zen draft user-story --preview

  # Force overwrite existing file
  zen draft epic --force

  # Generate to custom path
  zen draft roadmap --output ./custom/path/

```
zen draft <activity> [flags]
```

### Examples

```
  zen zen draft
```

### Options

```
      --force           Overwrite existing files
  -h, --help            help for draft
      --output string   Custom output directory
      --preview         Preview template without generating
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

