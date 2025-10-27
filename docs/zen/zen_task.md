---
title: "zen task"
slug: "/cli/zen-task"
description: "CLI reference for zen task"
section: "CLI Reference"
man_section: 1
since: v0.0.0
keywords:
  - zen
  - cli
---

## zen task

Manage tasks and workflow

### Synopsis

Manage tasks and workflow for Zen CLI.

Tasks are the core unit of work in Zen, following the seven-stage Zenflow
workflow: Align → Discover → Prioritize → Design → Build → Ship → Learn.

Each task creates a minimal directory in .zen/work/tasks/ with:
- index.md: Human-readable task overview
- manifest.yaml: Machine-readable metadata and workflow state
- .taskrc.yaml: Task-specific configuration
- .zenflow/: Workflow state tracking
- metadata/: External system snapshots

Work-type directories (research/, spikes/, design/, execution/, outcomes/)
are created on-demand when artifacts are added.

Tasks support different types:
- story: User-facing feature development
- bug: Defect fixes and corrections
- epic: Large initiatives spanning multiple tasks
- spike: Research and exploration work
- task: General work items

### Examples

```
  # Create a new story task
  zen task create PROJ-123 --type story

  # Create a bug fix task
  zen task create BUG-456 --type bug

  # Create an epic for large initiatives
  zen task create EPIC-789 --type epic

  # Create a research spike
  zen task create SPIKE-101 --type spike
```

### Options

```
  -h, --help   help for task
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
* [zen task create](zen-task-create.md.md)	 - Create a new task with structured workflow

