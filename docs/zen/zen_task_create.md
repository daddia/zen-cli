---
title: "zen task create"
slug: "/cli/zen-task-create"
description: "CLI reference for zen task create"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-09-21
keywords:
  - zen
  - cli
  - zen-task-create
---

## zen task create

Create a new task with structured workflow

### Synopsis

Create a new task with structured workflow directories and templates.

This command creates a complete task structure in .zen/work/tasks/<task-id>/ including:
- index.md: Human-readable task overview and status
- manifest.yaml: Machine-readable metadata for automation
- .taskrc.yaml: Task-specific configuration
- Work-type directories: research/, spikes/, design/, execution/, outcomes/

The task follows the seven-stage Zenflow workflow:
1. Align: Define success criteria and stakeholder alignment
2. Discover: Gather evidence and validate assumptions
3. Prioritize: Rank work by value vs effort
4. Design: Specify implementation approach
5. Build: Deliver working software increment
6. Ship: Deploy safely to production
7. Learn: Measure outcomes and iterate

Task types determine the workflow focus:
- story: User-facing feature development with UX focus
- bug: Defect fixes with root cause analysis
- epic: Large initiatives requiring breakdown
- spike: Research and exploration with learning focus
- task: General work items with flexible structure

```
zen task create <task-id> [flags]
```

### Examples

```
# Create a user story for feature development
zen task create USER-123 --type story --title "User login with SSO"

# Create a bug fix task
zen task create BUG-456 --type bug --title "Fix memory leak in auth service"

# Create an epic for large initiative
zen task create EPIC-789 --type epic --title "Implement new payment system"

# Create a research spike
zen task create SPIKE-101 --type spike --title "Evaluate GraphQL vs REST"

# Create with additional metadata
zen task create PROJ-200 --type story --title "Dashboard redesign" --owner "jane.doe" --team "frontend"

```

### Options

```
  -h, --help              help for create
      --owner string      Task owner (optional, defaults to current user)
      --priority string   Task priority (P0|P1|P2|P3) (default "P2")
      --team string       Team name (optional)
      --title string      Task title (optional, will prompt if not provided)
  -t, --type string       Task type (story|bug|epic|spike|task)
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

* [zen task](zen-task.md.md)	 - Manage tasks and workflow

