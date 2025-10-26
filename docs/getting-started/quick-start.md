# Quick Start

Get up and running with Zen CLI in 5 minutes.

## Prerequisites

Before starting, ensure you have:
- Zen CLI installed ([Installation Guide](installation.md))
- An API key from at least one AI provider (OpenAI, Anthropic, or Azure)
- A project directory ready for initialization

## Initialize Your First Project

### Step 1: Create a New Project

```bash
# Create and navigate to project directory
mkdir my-project
cd my-project

# Initialize Zen
zen init
```

This command will:
- Detect project type and structure
- Create `.zen` configuration directory
- Set up default workflow templates
- Initialize version control integration

### Step 2: Configure External Integrations (Optional)

Set up integrations with external systems:

```bash
# Configure Jira integration
zen config set integrations.providers.jira.url "https://company.atlassian.net"
zen config set integrations.providers.jira.project_key "PROJ"

# Authenticate with external systems
zen auth setup jira

# Configure GitHub integration (if needed)
zen config set integrations.providers.github.owner "company"
zen config set integrations.providers.github.repo "project"
zen auth setup github
```

Verify configuration:

```bash
zen config list
```

### Step 3: Check Project Status

Ensure everything is configured correctly:

```bash
zen status
```

Expected output:
```
Workspace Status
================
✓ Configuration: Valid
  - Log level: info
  - Output format: text

✓ Assets: Ready
  - Library: Synced
  - Cache: 42MB

✓ Integration: Configured
  - Jira: Connected (PROJ)
  - GitHub: Not configured
```

## Core Workflows

### Task Management

Create and manage tasks with structured workflows:

```bash
# Create a new task
zen task create PROJ-123 --title "Implement user authentication" --type story

# Create task from external system
zen task create JIRA-456 --from jira

# Create with additional metadata
zen task create FEAT-789 --title "Dashboard redesign" --owner "jane.doe" --team "frontend"
```

### Asset Library

Access and manage templates and assets:

```bash
# Authenticate with asset library
zen assets auth

# List available assets
zen assets list

# Get asset information
zen assets info template/task-index

# Sync asset library
zen assets sync
```

### Configuration Management

Manage Zen configuration:

```bash
# View current configuration
zen config list

# Set configuration values
zen config set log_level debug
zen config set integrations.providers.jira.url "https://company.atlassian.net"

# Get specific values
zen config get log_level
```

## Essential Commands

### Configuration Management

```bash
# View all configuration
zen config list

# Set a configuration value
zen config set <key> <value>

# Get a specific value
zen config get <key>

# Reset to defaults
zen config reset
```

### Project Information

```bash
# Show project status
zen status

# Display version
zen version

# Get help
zen help
zen help <command>
```

### Template Management

```bash
# List available templates
zen templates list

# Create from template
zen templates apply <template-name>

# Create custom template
zen templates create --from-workflow
```

## Working with AI Agents

### Interactive Mode

Start an interactive session:

```bash
# Launch interactive mode
zen agents chat

# With specific context
zen agents chat --context "src/"

# With specific agent
zen agents chat --agent architect
```

### Batch Operations

Process multiple tasks:

```bash
# Batch code review
zen agents review src/**/*.go

# Batch documentation
zen agents document --all-functions

# Batch refactoring
zen agents refactor --pattern singleton
```

## Example: Complete Feature Development

Walk through developing a new feature:

```bash
# 1. Define the feature
zen product create "Payment Processing"

# 2. Design architecture
zen workflow run architect --feature payment

# 3. Generate implementation
zen code generate "Payment service with Stripe integration"

# 4. Create tests
zen code test src/payment/

# 5. Review code
zen agents review src/payment/

# 6. Generate documentation
zen agents document src/payment/
```

## Integration Examples

### GitHub Integration

```bash
# Initialize GitHub integration
zen integrations add github

# Create pull request
zen workflow run pr-create --title "Add payment processing"

# Run CI checks
zen workflow run ci-check
```

### Jira Integration

```bash
# Connect to Jira
zen integrations add jira --url "company.atlassian.net"

# Sync tasks
zen integrations sync jira

# Create issue from feature
zen product sync --to jira
```

## Tips for Success

### Best Practices

1. **Start Small**: Begin with simple commands before complex workflows
2. **Use Templates**: Leverage pre-built templates for common tasks
3. **Version Control**: Always use Zen with Git for better tracking
4. **Regular Updates**: Keep Zen updated for latest features

### Productivity Tips

```bash
# Set up aliases for common commands
alias zs='zen status'
alias zw='zen workflow run'
alias zc='zen code generate'

# Use shell completion
source <(zen completion bash)  # or zsh/fish

# Export common configurations
export ZEN_DEFAULT_PROVIDER=openai
export ZEN_LOG_LEVEL=info
```

### Debugging

Enable verbose output for troubleshooting:

```bash
# Debug mode
zen --debug <command>

# Verbose logging
ZEN_LOG_LEVEL=debug zen <command>

# Dry run mode
zen workflow run development --dry-run
```

## Common Use Cases

### Starting a New Feature

```bash
zen product create "Feature Name"
zen workflow run development
```

### Daily Development

```bash
zen status
zen agents chat --context .
zen code generate "Implementation details"
```

### Code Review Preparation

```bash
zen agents review --changes
zen agents document --changes
zen workflow run pre-commit
```

### Production Deployment

```bash
zen workflow run test-all
zen workflow run security-scan
zen workflow run deploy --env production
```

## Getting Help

### Built-in Help

```bash
# General help
zen help

# Command-specific help
zen help <command>
zen <command> --help

# List all commands
zen help --all
```

### Resources

- [Configuration Guide](configuration.md) - Detailed configuration options
- [Command Reference](../zen/index.md) - Complete command documentation
- [Workflows Guide](workflows.md) - Creating custom workflows
- [Troubleshooting](troubleshooting.md) - Common issues and solutions

### Community

- [GitHub Issues](https://github.com/zen-org/zen/issues) - Report bugs or request features
- [Discussions](https://github.com/zen-org/zen/discussions) - Ask questions and share tips
- [Discord](https://discord.gg/zen) - Real-time community support

## Next Steps

Now that you have Zen running:

1. Explore the [Configuration Guide](configuration.md) for advanced setup
2. Review [Available Commands](../zen/index.md) for full capabilities
3. Create [Custom Workflows](workflows.md) for your team
4. Set up [Integrations](integrations.md) with your tools

Ready to boost your productivity? Start with `zen workflow run development` and let Zen guide you through the development process.
