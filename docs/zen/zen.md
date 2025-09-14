---
title: zen
slug: /cli/zen
section: "CLI Reference"
man_section: 1
since: v0.0.0
aliases: []
see_also:
  - zen-init
  - zen-config
  - zen-status
  - zen-version
---

# zen

```
zen <command> [flags]
```

Zen CLI - AI-Powered Productivity Suite

Zen is a unified command-line interface that revolutionizes productivity across the entire product lifecycle. By orchestrating intelligent workflows for both product management and engineering teams, Zen eliminates context switching, automates repetitive tasks, and ensures consistent quality delivery from ideation to production.

## Core Commands

`init`

Initialize a new Zen workspace

`config`

Manage configuration settings

`status`

Display workspace and system status

`version`

Display version information

## Product Management Commands

`product` *(coming soon)*

Product management commands

## Engineering Workflow Commands

`workflow` *(coming soon)*

Manage engineering workflows

## Additional Commands

`integrations` *(coming soon)*

Manage external integrations

`templates` *(coming soon)*

Template management

`agents` *(coming soon)*

AI agent management

`completion`

Generate shell completion scripts

## Global Options

`-v`, `--verbose`

Enable verbose output

`--no-color`

Disable colored output

`-o`, `--output <format>`

Output format (text, json, yaml)

`-c`, `--config <file>`

Path to configuration file

`--dry-run`

Show what would be executed without making changes

`-h`, `--help`

Show help for command

## Key Features

### ✓ Product Management Excellence
- Market research and competitive analysis
- Strategic planning and roadmap development
- Stakeholder communication and alignment

### ✓ Engineering Workflow Automation
- 12-stage development workflow automation
- Code generation and review assistance
- Quality assurance and testing integration

### ✓ AI-First Intelligence
- Multi-provider LLM support (OpenAI, Anthropic, etc.)
- Context-aware automation and suggestions
- Intelligent task prioritization and routing

### ✓ Comprehensive Integrations
- Product tools (Jira, Confluence, Notion)
- Engineering platforms (GitHub, GitLab, Jenkins)
- Communication tools (Slack, Teams, Discord)

## Getting Started

```bash
# Initialize a new workspace
$ zen init
Initialized empty Zen workspace in /Users/username/project/.zen/

# Check status
$ zen status
Zen CLI Status
...

# Configure settings
$ zen config set log_level debug
✓ Set log_level to "debug"

# Get help for any command
$ zen <command> --help
```

## Examples

```bash
# Initialize workspace with verbose output
$ zen init --verbose
Analyzing project in /Users/username/project...
Detected project type: Go module
Initialized empty Zen workspace in /Users/username/project/.zen/

# Check comprehensive status
$ zen status
Zen CLI Status

Workspace:
  Status:      ✓ Ready
  Path:        /Users/username/project
  Config File: .zen/config.yaml
...

# Configure for JSON output by default
$ zen config set cli.output_format json
✓ Set cli.output_format to "json"

# Use environment variables
$ ZEN_LOG_LEVEL=debug zen status
...

# Generate shell completion
$ zen completion bash > /usr/local/etc/bash_completion.d/zen
$ zen completion zsh > "${fpath[1]}/_zen"
```

## Configuration

Zen CLI supports configuration through multiple sources in order of precedence:

1. **Command-line flags** (highest priority)
2. **Environment variables** (with `ZEN_` prefix)
3. **Configuration files**
   - `.zen/config.yaml` (workspace-specific)
   - `~/.zen/config.yaml` (user-wide)
4. **Default values** (lowest priority)

### Environment Variables

All configuration keys can be set via environment variables with the `ZEN_` prefix:

- `ZEN_LOG_LEVEL=debug`
- `ZEN_CLI_NO_COLOR=true`
- `ZEN_CLI_OUTPUT_FORMAT=json`
- `ZEN_CONFIG=/path/to/config.yaml`

### Configuration Files

Configuration files use YAML format:

```yaml
log_level: info
log_format: text
cli:
  no_color: false
  verbose: false
  output_format: text
workspace:
  root: /Users/username/project
  config_file: zen.yaml
development:
  debug: false
  profile: false
```

## Exit Status

`0` - Success  
`1` - General error  
`2` - Invalid arguments  
`3` - Authentication/authorization error  
`4` - Resource not found  
`5` - Configuration error

## Notes

**Workspace Initialization:**
- Always run `zen init` in your project directory first
- Creates `.zen/` directory with configuration and cache
- Detects project type and applies appropriate defaults

**Command Discovery:**
- Use `zen --help` to see all available commands
- Each command has detailed help: `zen <command> --help`
- Many commands coming soon - check roadmap for updates

**Output Formats:**
- Human-readable text output by default
- JSON/YAML formats for scripting and automation
- Colors and Unicode symbols enhance readability (disable with `--no-color`)

**Integration Ready:**
- Designed for CI/CD pipeline integration
- Supports machine-readable output formats
- Environment variable configuration for containerized environments

## Documentation

- **Website:** https://zen.dev/docs
- **API Reference:** https://zen.dev/api
- **Examples:** https://zen.dev/examples
- **GitHub:** https://github.com/daddia/zen

## Support

- **Issues:** https://github.com/daddia/zen/issues
- **Discussions:** https://github.com/daddia/zen/discussions
- **Documentation:** https://zen.dev/docs

## See Also

* [zen init](zen-init.md) - Initialize a workspace
* [zen config](zen-config.md) - Manage configuration
* [zen status](zen-status.md) - Check workspace status
* [zen version](zen-version.md) - Display version information
