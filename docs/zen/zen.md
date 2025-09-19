---
title: "zen"
slug: "/cli/zen"
description: "CLI reference for zen"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 191919-09-96
keywords:
  - zen
  - cli
  - zen
---

## zen

AI-Powered Productivity Suite

### Synopsis

Zen CLI - AI-Powered Productivity Suite

Zen is a unified command-line interface that revolutionizes productivity across
the entire product lifecycle. By orchestrating intelligent workflows for both
product management and engineering teams, Zen eliminates context switching,
automates repetitive tasks, and ensures consistent quality delivery from
ideation to production.

Key Features:
  ✓ Product Management Excellence - Market research, strategy, and roadmap planning
  ✓ Engineering Workflow Automation - 12-stage development workflow automation
  ✓ AI-First Intelligence - Multi-provider LLM support with context-aware automation
  ✓ Comprehensive Integrations - Product tools, engineering platforms, and communication

Getting Started:
  zen init          Initialize a new workspace
  zen config        Configure Zen settings
  zen status        Check workspace status
  zen assets        Manage assets and templates
  zen --help        Show detailed help for any command

Documentation: https://zen.dev/docs
Report Issues:  https://github.com/daddia/zen/issues

### Examples

```
  # Initialize a new workspace
  zen init

  # Check workspace status
  zen status

  # Configure Zen settings
  zen config set log_level debug

  # Display version information
  zen version

  # Get help for any command
  zen <command> --help

  # Generate shell completion
  zen completion bash > /usr/local/etc/bash_completion.d/zen
```

### Options

```
  -c, --config string   Path to configuration file
      --dry-run         Show what would be executed without making changes
  -h, --help            help for zen
      --no-color        Disable colored output
  -o, --output string   Output format (text, json, yaml) (default "text")
  -v, --verbose         Enable verbose output
```

### SEE ALSO

* [zen agents](zen-agents.md.md)	 - AI agent management
* [zen assets](zen-assets.md.md)	 - Manage assets and templates
* [zen completion](zen-completion.md.md)	 - Generate shell completion scripts
* [zen config](zen-config.md.md)	 - Manage configuration for Zen CLI
* [zen init](zen-init.md.md)	 - Initialize a new Zen workspace
* [zen integrations](zen-integrations.md.md)	 - Manage external integrations
* [zen product](zen-product.md.md)	 - Product management commands
* [zen status](zen-status.md.md)	 - Display workspace and system status
* [zen version](zen-version.md.md)	 - Display version information
* [zen workflow](zen-workflow.md.md)	 - Manage engineering workflows

