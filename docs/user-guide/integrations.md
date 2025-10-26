# Integration Guide

Connect Zen CLI with your existing tools and workflows for seamless productivity.

## Overview

Zen CLI integrates with popular development and project management tools to create a unified workflow. This guide covers setting up and using these integrations effectively.

## Supported Integrations

### Project Management
- **[Jira](#jira-integration)** - Issue tracking and task management
- **GitHub Issues** (planned) - GitHub issue integration
- **Linear** (planned) - Modern issue tracking

### Version Control
- **[Git](#git-integration)** - Local repository operations
- **GitHub** (planned) - Repository and pull request management
- **GitLab** (planned) - GitLab integration

### Communication
- **Slack** (planned) - Team notifications and updates

## Jira Integration

### Setup

#### 1. Configure Jira Connection

```bash
# Set Jira instance URL
zen config set integrations.providers.jira.url "https://company.atlassian.net"

# Set project key
zen config set integrations.providers.jira.project_key "PROJ"

# Optional: Configure field mapping
zen config set integrations.providers.jira.field_mapping.priority "customfield_10001"
```

#### 2. Authenticate

```bash
# Interactive authentication setup
zen auth setup jira

# Check authentication status
zen auth status jira
```

#### 3. Verify Connection

```bash
# Test connection
zen status

# Should show:
# ✓ Integration: Configured
#   - Jira: Connected (PROJ)
```

### Usage

#### Creating Tasks from Jira Issues

```bash
# Create task from Jira issue
zen task create PROJ-123 --from jira

# This will:
# - Fetch issue details from Jira
# - Create local task structure
# - Generate metadata file with sync information
# - Set up Zenflow directories
```

#### Task Synchronization

```bash
# Check sync status
zen assets status

# Manual sync (if needed)
zen task sync PROJ-123

# Sync all tasks
zen task sync --all
```

#### Task Structure with Jira Integration

When you create a task from Jira, Zen creates this structure:

```
.zen/work/tasks/PROJ-123/
├── index.md              # Human-readable overview
├── manifest.yaml         # Machine-readable metadata
├── .taskrc.yaml         # Task-specific configuration
├── metadata/
│   └── jira.json        # Jira sync metadata
├── research/            # Align & Discover stages
├── spikes/              # Discover & Prioritize stages
├── design/              # Design stage
├── execution/           # Build stage
└── outcomes/            # Ship & Learn stages
```

### Configuration Options

#### Basic Configuration

```yaml
# zen.yaml
integrations:
  providers:
    jira:
      url: "https://company.atlassian.net"
      project_key: "PROJ"
      sync_direction: "pull"  # pull, push, bidirectional
      sync_enabled: true
```

#### Advanced Configuration

```yaml
integrations:
  providers:
    jira:
      url: "https://company.atlassian.net"
      project_key: "PROJ"
      type: "cloud"  # cloud or server
      
      # Field mapping for custom fields
      field_mapping:
        story_points: "customfield_10002"
        epic_link: "customfield_10014"
        sprint: "customfield_10020"
        
      # Sync configuration
      sync_direction: "bidirectional"
      sync_enabled: true
      sync_frequency: "15m"
      
      # Issue type mapping
      type_mapping:
        "Story": "story"
        "Bug": "bug"
        "Epic": "epic"
        "Task": "task"
        "Spike": "spike"
```

### Troubleshooting

#### Authentication Issues

```bash
# Check authentication status
zen auth status jira

# Re-authenticate if needed
zen auth setup jira --force

# Clear stored credentials
zen auth logout jira
```

#### Connection Problems

```bash
# Test connection with verbose output
zen --verbose status

# Check configuration
zen config get integrations.providers.jira.url
zen config get integrations.providers.jira.project_key
```

#### Common Error Messages

**"Authentication required for Jira integration"**
```bash
# Solution: Set up authentication
zen auth setup jira
```

**"Jira connection timeout"**
```bash
# Solution: Check network and URL
zen config get integrations.providers.jira.url
# Verify URL is accessible
```

**"Project key not found"**
```bash
# Solution: Verify project key
zen config set integrations.providers.jira.project_key "CORRECT-KEY"
```

## Git Integration

### Setup

Git integration is automatically configured when you run `zen init` in a Git repository.

#### Manual Configuration

```bash
# Initialize in existing Git repo
cd /path/to/git/repo
zen init

# Zen will automatically detect:
# - Repository URL
# - Current branch
# - Remote configuration
```

### Usage

#### Repository Information

```bash
# View repository status
zen status

# Shows Git information:
# ✓ Project: my-project (git)
#   - Repository: https://github.com/user/repo
#   - Branch: main
#   - Status: Clean
```

#### Integration with Tasks

```bash
# Create task with Git context
zen task create FEAT-123 --title "New feature"

# Task metadata includes Git information:
# - Current branch
# - Repository URL
# - Commit hash (if any)
```

### Git Workflow Integration

#### Branch Management

```bash
# Create task-specific branch (planned feature)
zen task create FEAT-123 --create-branch

# This would:
# - Create task structure
# - Create Git branch: feature/FEAT-123
# - Switch to new branch
```

#### Commit Integration

```bash
# Generate commit messages from task context (planned)
zen task commit FEAT-123

# Would generate commit message like:
# "feat(FEAT-123): Implement user authentication
# 
# - Add OAuth2 integration
# - Update user model
# - Add authentication middleware"
```

## Authentication Management

### Overview

Zen securely stores credentials using platform-specific secure storage:

- **macOS**: Keychain Services
- **Windows**: Windows Credential Manager
- **Linux**: Secret Service (libsecret)

### Authentication Commands

#### Setup Authentication

```bash
# Interactive setup for specific provider
zen auth setup jira
zen auth setup github
zen auth setup slack

# Setup all configured providers
zen auth setup --all
```

#### Check Authentication Status

```bash
# Check all providers
zen auth status

# Check specific provider
zen auth status jira

# Example output:
# Authentication Status
# =====================
# ✓ Jira: Authenticated (expires in 7 days)
# ✗ GitHub: Not authenticated
# - Slack: Not configured
```

#### Manage Credentials

```bash
# Remove stored credentials
zen auth logout jira
zen auth logout --all

# Refresh expired credentials
zen auth refresh jira
zen auth refresh --all
```

### Security Best Practices

1. **Never commit credentials** - Use `zen auth` for secure storage
2. **Regular credential rotation** - Refresh tokens periodically
3. **Minimal permissions** - Use least-privilege access tokens
4. **Monitor access** - Check authentication status regularly

## Configuration Management

### Integration Configuration

#### Global Configuration

```yaml
# zen.yaml
integrations:
  sync_enabled: true
  sync_frequency: "15m"
  plugin_directories:
    - ".zen/plugins"
    - "/usr/local/zen/plugins"
  
  providers:
    jira:
      url: "https://company.atlassian.net"
      project_key: "PROJ"
      type: "cloud"
      sync_direction: "pull"
      
    github:
      owner: "company"
      repo: "project"
      sync_direction: "bidirectional"
```

#### Environment Variables

```bash
# Jira configuration
export ZEN_JIRA_URL="https://company.atlassian.net"
export ZEN_JIRA_PROJECT_KEY="PROJ"

# GitHub configuration
export ZEN_GITHUB_OWNER="company"
export ZEN_GITHUB_REPO="project"

# General integration settings
export ZEN_INTEGRATION_SYNC_ENABLED=true
export ZEN_INTEGRATION_SYNC_FREQUENCY="30m"
```

#### Task-Specific Configuration

```yaml
# .zen/work/tasks/PROJ-123/.taskrc.yaml
task:
  id: "PROJ-123"
  type: "story"
  
integration:
  jira:
    sync_enabled: true
    last_sync: "2025-10-26T10:30:00Z"
    external_id: "PROJ-123"
    
  github:
    sync_enabled: false
    pull_request: null
```

## Advanced Usage

### Custom Field Mapping

Map Jira custom fields to Zen task properties:

```yaml
integrations:
  providers:
    jira:
      field_mapping:
        # Map custom fields by ID
        story_points: "customfield_10002"
        epic_link: "customfield_10014"
        sprint: "customfield_10020"
        team: "customfield_10030"
        
        # Map standard fields
        priority: "priority"
        assignee: "assignee"
        status: "status"
```

### Sync Strategies

#### Pull-Only (Default)

```yaml
sync_direction: "pull"
```
- Fetches data from external system
- Updates local task information
- No changes sent back to external system

#### Push-Only

```yaml
sync_direction: "push"
```
- Sends local changes to external system
- Updates external issues/tasks
- No data fetched from external system

#### Bidirectional

```yaml
sync_direction: "bidirectional"
```
- Synchronizes in both directions
- Resolves conflicts using last-modified wins
- Requires careful conflict resolution

### Webhook Integration (Planned)

```yaml
integrations:
  webhooks:
    enabled: true
    port: 8080
    endpoints:
      jira: "/webhooks/jira"
      github: "/webhooks/github"
```

## Troubleshooting

### Common Issues

#### Integration Not Working

1. **Check configuration**:
   ```bash
   zen config list | grep integrations
   ```

2. **Verify authentication**:
   ```bash
   zen auth status
   ```

3. **Test connection**:
   ```bash
   zen status --verbose
   ```

#### Sync Problems

1. **Check sync status**:
   ```bash
   zen assets status
   ```

2. **Manual sync**:
   ```bash
   zen task sync PROJ-123 --force
   ```

3. **Clear sync metadata**:
   ```bash
   rm .zen/work/tasks/PROJ-123/metadata/jira.json
   zen task sync PROJ-123
   ```

#### Performance Issues

1. **Reduce sync frequency**:
   ```bash
   zen config set integrations.sync_frequency "1h"
   ```

2. **Disable unused integrations**:
   ```bash
   zen config set integrations.providers.unused.sync_enabled false
   ```

3. **Clear cache**:
   ```bash
   zen assets sync --force
   ```

### Debug Mode

Enable verbose logging for troubleshooting:

```bash
# Enable debug logging
zen config set log_level debug

# Run commands with verbose output
zen --verbose task create PROJ-123 --from jira

# Check logs
tail -f .zen/logs/zen.log
```

## Best Practices

### Integration Setup

1. **Start simple** - Configure one integration at a time
2. **Test thoroughly** - Verify each integration before adding more
3. **Document configuration** - Keep track of custom field mappings
4. **Regular maintenance** - Update credentials and test connections

### Task Management

1. **Consistent naming** - Use clear, descriptive task IDs
2. **Regular syncing** - Keep tasks synchronized with external systems
3. **Conflict resolution** - Handle sync conflicts promptly
4. **Backup metadata** - Version control task metadata files

### Security

1. **Secure credentials** - Always use `zen auth` for credential storage
2. **Minimal permissions** - Use least-privilege API tokens
3. **Regular rotation** - Refresh credentials periodically
4. **Monitor access** - Review authentication logs regularly

## Migration Guide

### From Manual Workflows

```bash
# Before: Manual issue tracking
# - Create Jira issues manually
# - Copy information to local files
# - Update status in multiple places

# After: Integrated workflow
zen task create PROJ-123 --from jira  # Automatic sync
zen status                            # Unified status view
```

### From Other Tools

```bash
# Export existing task data
zen task export --format json > tasks-backup.json

# Import to new integration
zen task import --from jira --file tasks-backup.json
```

## See Also

- **[Task Management](task-management.md)** - Task workflow guide
- **[Configuration Guide](../getting-started/configuration.md)** - Configuration reference
- **[Authentication API](../api/auth.md)** - Authentication API reference
- **[Jira Client API](../api/jira-client.md)** - Jira integration API
