# Zen CLI User Guide

Complete guide to using Zen CLI for product and engineering teams.

## Quick Start

### Installation and Setup

1. **Install Zen CLI** following the [Installation Guide](../getting-started/installation.md)
2. **Initialize your workspace**:

   ```bash
   cd your-project
   zen init
   ```

3. **Check status**:

   ```bash
   zen status
   ```

### First Steps

```bash
# Configure authentication (if using external integrations)
zen auth

# View configuration
zen config list

# Create your first task
zen task create PROJ-123 --title "Setup project infrastructure"

# Explore available assets
zen assets list
```

## Core Workflows

### Workspace Management

#### Setting Up a New Project

```bash
# Navigate to your project directory
cd /path/to/your/project

# Initialize Zen workspace (detects project type automatically)
zen init

# Verify setup
zen status
```

**What happens during initialization:**
- Creates `.zen/` directory structure
- Generates `zen.yaml` configuration file
- Detects project type (Go, Node.js, Python, etc.)
- Sets up asset library infrastructure
- Configures logging and caching

#### Project Type Detection

Zen automatically detects and configures for:

- **Git repositories** - Integrates with Git workflows
- **Go projects** - Detects `go.mod`, configures Go toolchain
- **Node.js projects** - Detects `package.json`, configures npm/yarn
- **Python projects** - Detects `requirements.txt`, `pyproject.toml`
- **Rust projects** - Detects `Cargo.toml`, configures Cargo
- **Java projects** - Detects `pom.xml`, `build.gradle`

### Configuration Management

#### Viewing Configuration

```bash
# List all configuration
zen config list

# Get specific value
zen config get log_level

# View in different formats
zen config list --output json
zen config list --output yaml
```

#### Setting Configuration

```bash
# Set log level
zen config set log_level debug

# Set output format preference
zen config set cli.output_format json

# Configure external integrations
zen config set integrations.providers.jira.url "https://company.atlassian.net"
zen config set integrations.providers.jira.project_key "PROJ"
```

#### Configuration Sources

Configuration is loaded in order of precedence:

1. **Command-line flags** - `--config`, `--verbose`, etc.
2. **Environment variables** - `ZEN_LOG_LEVEL`, `ZEN_NO_COLOR`, etc.
3. **Configuration files** - `zen.yaml`, `.zen/config/`
4. **Default values** - Built-in sensible defaults

### Task Management

#### Creating Tasks

```bash
# Basic task creation
zen task create PROJ-123 --title "Implement user authentication"

# Specify task type
zen task create BUG-456 --type bug --title "Fix memory leak in auth service"

# Add metadata
zen task create FEAT-789 \
  --title "Dashboard redesign" \
  --type story \
  --owner "jane.doe" \
  --team "frontend" \
  --priority high
```

#### Task Types and Workflows

Each task type optimizes the Zenflow workflow:

- **`story`** - User-facing features with UX focus
- **`bug`** - Defect fixes with root cause analysis
- **`epic`** - Large initiatives requiring breakdown
- **`spike`** - Research and exploration
- **`task`** - General work items

#### External Integration

```bash
# Sync with Jira
zen task create JIRA-123 --from jira

# Sync with GitHub Issues
zen task create GH-456 --from github

# Local-only task
zen task create LOCAL-789 --from local
```

### Asset Library Management

#### Authentication Setup

```bash
# Authenticate with GitHub (for asset library access)
zen assets auth

# Check authentication status
zen assets status
```

#### Browsing Assets

```bash
# List all available assets
zen assets list

# Filter by type
zen assets list --type template
zen assets list --type workflow

# Search assets
zen assets list --filter "authentication"

# Get detailed information
zen assets info template/auth-flow
```

#### Syncing Assets

```bash
# Sync asset library
zen assets sync

# Force refresh
zen assets sync --force

# Sync specific assets
zen assets sync --filter "template/*"
```

### Authentication Management

#### Setting Up Authentication

```bash
# Interactive authentication setup
zen auth

# Check authentication status
zen auth status

# Configure specific providers
zen auth setup jira
zen auth setup github
zen auth setup slack
```

#### Managing Credentials

Zen securely stores credentials using:
- **macOS**: Keychain
- **Windows**: Credential Manager  
- **Linux**: Secret Service (libsecret)

```bash
# View authentication status
zen auth status

# Remove stored credentials
zen auth logout jira
zen auth logout --all
```

## Advanced Usage

### Output Formats and Scripting

#### Machine-Readable Output

```bash
# JSON output for scripting
zen status --output json | jq '.workspace.status'

# YAML output for configuration
zen config list --output yaml > backup-config.yaml

# Tab-delimited for processing
zen assets list --output text | cut -f1,3
```

#### Environment Variables

```bash
# Disable colors for scripts
export ZEN_NO_COLOR=1

# Set default output format
export ZEN_OUTPUT_FORMAT=json

# Configure logging
export ZEN_LOG_LEVEL=debug
export ZEN_LOG_FORMAT=json
```

### Integration Workflows

#### Jira Integration

```bash
# Configure Jira connection
zen config set integrations.providers.jira.url "https://company.atlassian.net"
zen config set integrations.providers.jira.project_key "PROJ"

# Authenticate
zen auth setup jira

# Create synced tasks
zen task create PROJ-123 --from jira
```

#### GitHub Integration

```bash
# Configure GitHub integration
zen config set integrations.providers.github.owner "company"
zen config set integrations.providers.github.repo "project"

# Authenticate
zen auth setup github

# Create synced tasks
zen task create GH-456 --from github
```

### Troubleshooting

#### Common Issues

**Workspace not initialized:**
```bash
✗ Workspace not found in current directory

Run 'zen init' to initialize a workspace
```

**Authentication required:**
```bash
✗ Authentication required for Jira integration

Run 'zen auth setup jira' to configure authentication
```

**Configuration errors:**
```bash
✗ Invalid configuration: log_level must be one of: trace, debug, info, warn, error

Available values: trace, debug, info, warn, error, fatal, panic
```

#### Debug Mode

```bash
# Enable verbose output
zen --verbose status

# Enable debug logging
zen config set log_level debug
zen status

# Dry run mode
zen task create PROJ-123 --title "Test task" --dry-run
```

#### Getting Help

```bash
# General help
zen --help

# Command-specific help
zen task --help
zen task create --help

# Version and build information
zen version
```

## Best Practices

### Workspace Organization

- **One workspace per project** - Initialize Zen in your project root
- **Consistent naming** - Use clear, descriptive task IDs
- **Regular syncing** - Keep asset library updated with `zen assets sync`
- **Configuration management** - Use version control for `zen.yaml`

### Task Management

- **Descriptive titles** - Make task purposes clear
- **Appropriate types** - Choose the right task type for workflow optimization
- **External sync** - Use integration when working with teams
- **Regular updates** - Keep task status current

### Configuration

- **Environment-specific** - Use different configs for dev/staging/prod
- **Security** - Never commit credentials, use `zen auth`
- **Documentation** - Comment configuration choices in `zen.yaml`
- **Backup** - Export configuration with `zen config list --output yaml`

### Integration

- **Authentication first** - Set up auth before using integrations
- **Test connections** - Verify with `zen status` after configuration
- **Gradual adoption** - Start with one integration, expand gradually
- **Monitor status** - Regular `zen status` checks for integration health

## Examples

### Complete Project Setup

```bash
# 1. Initialize new project
mkdir my-project && cd my-project
git init
zen init

# 2. Configure integrations
zen config set integrations.providers.jira.url "https://company.atlassian.net"
zen config set integrations.providers.jira.project_key "MYPROJ"

# 3. Set up authentication
zen auth setup jira
zen assets auth

# 4. Verify setup
zen status

# 5. Create first task
zen task create MYPROJ-1 --title "Project setup and configuration" --type task

# 6. Sync assets
zen assets sync
```

### Daily Workflow

```bash
# Morning: Check status and sync
zen status
zen assets sync

# Create tasks as needed
zen task create MYPROJ-42 --title "Implement user login" --type story --from jira

# Check configuration when needed
zen config get integrations.providers.jira.project_key

# End of day: Verify everything is synced
zen status
```

## See Also

- **[Getting Started](../getting-started/README.md)** - Installation and basic setup
- **[Configuration Reference](../getting-started/configuration.md)** - Complete configuration options
- **[Command Reference](../zen/)** - Auto-generated command documentation
- **[Zenflow Guide](../zenflow/README.md)** - Workflow methodology
- **[Architecture](../architecture/README.md)** - Technical details
