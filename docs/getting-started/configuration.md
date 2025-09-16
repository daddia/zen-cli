# Configuration

Zen CLI uses a hierarchical configuration system that supports multiple sources and formats.

## Configuration Hierarchy

Configuration is loaded and merged in the following order (highest priority first):

1. **Command-line flags** - Override all other settings
2. **Environment variables** - System-wide or session settings
3. **Project configuration** - `.zen/config.yaml` in project root
4. **User configuration** - `~/.config/zen/config.yaml`
5. **System configuration** - `/etc/zen/config.yaml`
6. **Default values** - Built-in defaults

## Configuration Files

### File Locations

```bash
# Project-level (highest priority for file-based config)
.zen/config.yaml

# User-level
~/.config/zen/config.yaml     # Linux/macOS
%APPDATA%\zen\config.yaml      # Windows

# System-level
/etc/zen/config.yaml           # Linux/macOS
C:\ProgramData\zen\config.yaml # Windows
```

### File Format

Zen supports YAML, JSON, and TOML configuration formats:

#### YAML (Recommended)
```yaml
# ~/.config/zen/config.yaml
provider: openai
log_level: info

openai:
  api_key: sk-...
  model: gpt-4
  temperature: 0.7
  max_tokens: 2000

workspace:
  auto_detect: true
  default_branch: main

features:
  auto_complete: true
  syntax_highlighting: true
  telemetry: false
```

#### JSON
```json
{
  "provider": "openai",
  "log_level": "info",
  "openai": {
    "api_key": "sk-...",
    "model": "gpt-4"
  }
}
```

#### TOML
```toml
provider = "openai"
log_level = "info"

[openai]
api_key = "sk-..."
model = "gpt-4"
```

## Configuration Management

### Setting Configuration

```bash
# Set a single value
zen config set provider anthropic
zen config set log_level debug

# Set nested values
zen config set openai.model gpt-4
zen config set workspace.auto_detect false

# Set from file
zen config import config.yaml
```

### Getting Configuration

```bash
# Get a specific value
zen config get provider
zen config get openai.model

# List all configuration
zen config list

# List with sources (shows where each value comes from)
zen config list --show-source

# Export configuration
zen config export > my-config.yaml
```

### Resetting Configuration

```bash
# Reset a specific value to default
zen config unset provider

# Reset all configuration
zen config reset

# Reset to factory defaults (removes all config files)
zen config reset --factory
```

## Environment Variables

All configuration options can be set via environment variables using the `ZEN_` prefix:

```bash
# Format: ZEN_<SECTION>_<KEY>
export ZEN_PROVIDER=openai
export ZEN_LOG_LEVEL=debug
export ZEN_OPENAI_API_KEY=sk-...
export ZEN_OPENAI_MODEL=gpt-4
export ZEN_WORKSPACE_AUTO_DETECT=true

# Special environment variables
export ZEN_CONFIG_DIR=/custom/config/path
export ZEN_CACHE_DIR=/custom/cache/path
export ZEN_DATA_DIR=/custom/data/path
```

## Configuration Options

### Core Settings

```yaml
# Provider selection
provider: openai|anthropic|azure|local

# Logging configuration
log_level: debug|info|warn|error
log_format: json|text|pretty
log_file: /path/to/logfile.log

# Output formatting
output_format: text|json|yaml
color_output: true
interactive: true

# Performance
parallel_workers: 4
timeout: 30s
retry_attempts: 3
```

### AI Provider Configuration

#### OpenAI
```yaml
openai:
  api_key: sk-...
  organization: org-...
  model: gpt-4|gpt-3.5-turbo
  temperature: 0.7
  max_tokens: 2000
  top_p: 1.0
  frequency_penalty: 0.0
  presence_penalty: 0.0
  timeout: 60s
  base_url: https://api.openai.com/v1  # For proxies or compatible APIs
```

#### Anthropic
```yaml
anthropic:
  api_key: sk-ant-...
  model: claude-3-opus|claude-3-sonnet|claude-2
  max_tokens: 4000
  temperature: 0.7
  timeout: 60s
```

#### Azure OpenAI
```yaml
azure:
  api_key: your-key
  endpoint: https://your-resource.openai.azure.com/
  deployment: your-deployment-name
  api_version: 2023-12-01-preview
  model: gpt-4
```

#### Local Models
```yaml
local:
  model_path: /path/to/model
  model_type: llama|gpt4all|ollama
  context_size: 2048
  gpu_layers: 0  # Number of layers to offload to GPU
  threads: 4
```

### Workspace Configuration

```yaml
workspace:
  # Project detection
  auto_detect: true
  project_root: .
  scan_depth: 3
  
  # Version control
  git_enabled: true
  default_branch: main
  commit_convention: conventional|angular|custom
  
  # File handling
  ignore_patterns:
    - "*.log"
    - "node_modules/**"
    - ".git/**"
  include_patterns:
    - "src/**/*.go"
    - "*.md"
```

### Integration Settings

```yaml
integrations:
  github:
    enabled: true
    token: ghp_...
    api_url: https://api.github.com
    default_owner: your-org
    
  jira:
    enabled: false
    url: https://company.atlassian.net
    email: user@company.com
    api_token: ...
    
  slack:
    enabled: false
    webhook_url: https://hooks.slack.com/...
    channel: "#dev"
    username: ZenBot
```

### Feature Flags

```yaml
features:
  # UI features
  auto_complete: true
  syntax_highlighting: true
  progress_bars: true
  
  # Behavior
  auto_save: true
  confirm_destructive: true
  dry_run_default: false
  
  # Telemetry
  telemetry: false
  crash_reports: false
  usage_stats: false
```

### Security Settings

```yaml
security:
  # API key storage
  keyring_enabled: true
  encryption_key: ${ZEN_ENCRYPTION_KEY}
  
  # Network
  ssl_verify: true
  proxy: http://proxy.company.com:8080
  no_proxy: localhost,127.0.0.1
  
  # Audit
  audit_log: true
  audit_file: ~/.zen/audit.log
```

## Profiles

Use profiles for different environments or contexts:

### Creating Profiles

```bash
# Create a new profile
zen config profile create production

# Copy from existing profile
zen config profile create staging --from production

# Set profile-specific values
zen config set log_level warn --profile production
zen config set dry_run_default true --profile staging
```

### Using Profiles

```bash
# Switch profiles
zen config profile use production

# Run command with specific profile
zen --profile staging workflow run deploy

# List profiles
zen config profile list

# Delete profile
zen config profile delete old-profile
```

### Profile Configuration Files

```yaml
# ~/.config/zen/profiles/production.yaml
extends: default
log_level: warn
features:
  dry_run_default: false
  confirm_destructive: true

# ~/.config/zen/profiles/development.yaml  
extends: default
log_level: debug
features:
  dry_run_default: true
  confirm_destructive: false
```

## Advanced Configuration

### Template Variables

Use variables in configuration:

```yaml
# Define variables
variables:
  project_name: ${PROJECT_NAME:-my-project}
  api_endpoint: https://api.${ENVIRONMENT:-dev}.example.com

# Use variables
openai:
  api_key: ${OPENAI_API_KEY}
  organization: ${OPENAI_ORG:-default-org}
```

### Conditional Configuration

```yaml
# Environment-based conditions
$if: ${ENVIRONMENT} == "production"
then:
  log_level: warn
  features:
    telemetry: true
else:
  log_level: debug
  features:
    telemetry: false
```

### Include External Configuration

```yaml
# Include other configuration files
$include:
  - ~/.config/zen/providers.yaml
  - ~/.config/zen/integrations.yaml
  
# Include with conditions
$include:
  - path: ~/.config/zen/prod.yaml
    if: ${ENVIRONMENT} == "production"
```

## Validation

### Configuration Schema

Zen validates configuration against a schema:

```bash
# Validate current configuration
zen config validate

# Validate specific file
zen config validate --file custom-config.yaml

# Show schema
zen config schema

# Generate example configuration
zen config example > example-config.yaml
```

### Common Validation Errors

```bash
# Invalid provider
Error: provider "invalid" not recognized
Valid options: openai, anthropic, azure, local

# Missing required field
Error: openai.api_key is required when provider is "openai"

# Type mismatch
Error: parallel_workers must be a number, got "four"
```

## Migration

### Migrating from Other Tools

```bash
# Import from GitHub CLI
zen config migrate --from gh-cli

# Import from environment variables
zen config migrate --from env

# Custom migration script
zen config migrate --script migrate.js
```

### Upgrading Configuration

```bash
# Check for deprecated options
zen config check

# Auto-upgrade configuration format
zen config upgrade

# Backup before upgrade
zen config backup
```

## Best Practices

### Security

1. **Never commit API keys** - Use environment variables or secure key storage
2. **Use keyring integration** - Enable `security.keyring_enabled`
3. **Rotate keys regularly** - Update API keys periodically
4. **Audit configuration** - Review `zen config list` regularly

### Organization

1. **Use profiles** - Separate development, staging, and production
2. **Project-specific config** - Keep project settings in `.zen/config.yaml`
3. **Share team settings** - Commit safe configuration to version control
4. **Document custom config** - Add comments to YAML files

### Performance

1. **Adjust workers** - Set `parallel_workers` based on CPU cores
2. **Configure timeouts** - Increase for slow networks
3. **Enable caching** - Use cache for repeated operations
4. **Minimize includes** - Avoid excessive configuration includes

## Troubleshooting

### Debug Configuration Issues

```bash
# Show configuration resolution
zen config debug

# Test specific configuration
zen config test --key openai.api_key

# Show configuration sources
zen config sources

# Reset problematic configuration
zen config reset --section integrations
```

### Common Problems

#### Configuration Not Loading
```bash
# Check file permissions
ls -la ~/.config/zen/config.yaml

# Verify file format
zen config validate --file config.yaml

# Check for syntax errors
yamllint ~/.config/zen/config.yaml
```

#### Environment Variables Not Working
```bash
# List recognized environment variables
zen config env

# Debug environment loading
ZEN_DEBUG=true zen config list
```

#### Profile Issues
```bash
# Verify active profile
zen config profile current

# Check profile path
ls ~/.config/zen/profiles/

# Reset profile
zen config profile reset
```

## Next Steps

- Explore [Advanced Workflows](workflows.md)
- Set up [Integrations](integrations.md)
- Configure [AI Agents](agents.md)
- Review [Security Best Practices](security.md)
