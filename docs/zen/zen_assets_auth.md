---
title: "zen assets auth"
slug: "/cli/zen-assets-auth"
description: "CLI reference for zen assets auth"
section: "CLI Reference"
man_section: 1
since: v0.0.0
keywords:
  - zen
  - cli
---

## zen assets auth

Authenticate with Git providers for asset access

### Synopsis

Authenticate with Git providers to access private asset repositories.

Supported providers:
- github: GitHub Personal Access Token authentication
- gitlab: GitLab Project Access Token authentication

Authentication tokens are stored securely using your operating system's
credential manager (Keychain on macOS, Credential Manager on Windows,
Secret Service on Linux).

For GitHub:
1. Go to Settings > Developer settings > Personal access tokens
2. Generate a new token with 'repo' scope for private repositories
3. Use the token with this command

For GitLab:
1. Go to your project > Settings > Access Tokens
2. Create a project access token with 'read_repository' scope
3. Use the token with this command

```
zen assets auth [provider] [flags]
```

### Examples

```
# Authenticate with GitHub (interactive)
zen assets auth github

# Authenticate with GitHub using a token file
zen assets auth github --token-file ~/.tokens/github

# Authenticate with GitLab using environment variable
GITLAB_TOKEN=glpat-xxx zen assets auth gitlab

# Validate existing credentials
zen assets auth --validate

```

### Options

```
  -h, --help                help for auth
      --token string        Authentication token (not recommended for security)
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

* [zen assets](zen-assets.md.md)	 - Manage assets and templates

