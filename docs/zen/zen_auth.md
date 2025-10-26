---
title: "zen auth"
slug: "/cli/zen-auth"
description: "CLI reference for zen auth"
section: "CLI Reference"
man_section: 1
since: v0.0.0
date: 2025-10-26
keywords:
  - zen
  - cli
  - zen-auth
---

## zen auth

Authenticate with Git providers

### Synopsis

Authenticate with Git providers for secure access to repositories and services.

Supported providers:
- github: GitHub Personal Access Token authentication
- gitlab: GitLab Project Access Token authentication

Authentication tokens are stored securely using your operating system's
credential manager (Keychain on macOS, Credential Manager on Windows,
Secret Service on Linux).

The auth command provides a centralized authentication system used by all
Zen CLI components that require Git provider access, including asset management,
repository operations, and future integrations.

```
zen auth [provider] [flags]
```

### Examples

```
# Authenticate with GitHub (interactive)
zen auth github

# Authenticate with GitHub using a token
zen auth github --token ghp_your_token_here

# Authenticate with GitHub using a token file
zen auth github --token-file ~/.tokens/github

# Authenticate with GitLab
zen auth gitlab --token glpat_your_token_here

# List all authenticated providers
zen auth --list

# Validate existing credentials
zen auth github --validate

# Delete stored credentials
zen auth github --delete

# Use environment variable (Zen standard)
ZEN_GITHUB_TOKEN=ghp_token zen auth github

```

### Options

```
      --delete              Delete stored credentials for the provider
  -h, --help                help for auth
      --list                List all authenticated providers
      --token string        Authentication token (use environment variable for better security)
      --token-file string   Path to file containing authentication token
      --validate            Validate token after authentication (default true)
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

