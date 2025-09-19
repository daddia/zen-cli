---
title: "zen completion"
slug: "/cli/zen-completion"
description: "CLI reference for zen completion"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 191919-09-96
keywords:
  - zen
  - cli
  - zen-completion
---

## zen completion

Generate shell completion scripts

### Synopsis

Generate shell completion scripts for Zen CLI.

The completion script for each shell will be different. Please refer to your shell's
documentation on how to install completion scripts.

Examples:
  # Generate bash completion script
  zen completion bash > /usr/local/etc/bash_completion.d/zen

  # Generate zsh completion script
  zen completion zsh > "${fpath[1]}/_zen"

  # Generate fish completion script
  zen completion fish > ~/.config/fish/completions/zen.fish

  # Generate PowerShell completion script
  zen completion powershell > zen.ps1

```
zen completion [bash|zsh|fish|powershell] [flags]
```

### Examples

```
  zen zen completion
```

### Options

```
  -h, --help   help for completion
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

