---
title: zen-config
slug: /cli/zen-config
section: "CLI Reference"
man_section: 1
since: v0.0.0
aliases: []
see_also:
  - zen-init
  - zen-status
---

# zen config

## Synopsis

```sh
zen config [flags]
zen config list [flags]
zen config get <key> [flags]
zen config set <key> <value> [flags]
```

## Description

You can query/set/replace/unset configuration options with this command. The name is actually the configuration key, and the value will be escaped appropriately.

Configuration is read from multiple sources in the following order:
1. Command-line options (highest precedence)
2. Environment variables with `ZEN_` prefix
3. Configuration files in order:
   - `.zen/config.yaml` (workspace-specific)
   - `~/.zen/config.yaml` (user-specific)
4. Default values (lowest precedence)

When no action is specified, the command displays the current configuration in human-readable format.

## Subcommands

`list`

List all variables set in config file, along with their values.

`get <key>`

Get the value for a given key. Returns error code 1 if the key was not found.

`set <key> <value>`

Set the configuration option to the given value.

## Options

`-h`, `--help`

Show help for command

### Options inherited from parent commands

`-v`, `--verbose`

Enable verbose output and show configuration source

`--no-color`

Disable colored output

`-o`, `--output <format>`

Output format (text, json, yaml)

## Configuration Variables

### Core Configuration

`log_level`

Controls the verbosity of logging output. Must be one of: `trace`, `debug`, `info`, `warn`, `error`, `fatal`, `panic`. Default is `info`.

`log_format`

Format for log output. Must be one of: `text`, `json`. Default is `text`.

### CLI Behavior

`cli.no_color`

When set to `true`, disables colored output. Can be set via `ZEN_CLI_NO_COLOR` environment variable. Default is `false`.

`cli.verbose`

When set to `true`, enables verbose output by default. Can be set via `ZEN_CLI_VERBOSE` environment variable. Default is `false`.

`cli.output_format`

Default output format for commands. Must be one of: `text`, `json`, `yaml`. Can be set via `ZEN_CLI_OUTPUT_FORMAT` environment variable. Default is `text`.

### Workspace Configuration

`workspace.root`

Path to the workspace root directory. Can be set via `ZEN_WORKSPACE_ROOT` environment variable.

`workspace.config_file`

Name of the workspace configuration file. Can be set via `ZEN_WORKSPACE_CONFIG_FILE` environment variable. Default is `zen.yaml`.

### Development Options

`development.debug`

When set to `true`, enables debug mode with additional diagnostic information. Can be set via `ZEN_DEVELOPMENT_DEBUG` environment variable. Default is `false`.

`development.profile`

When set to `true`, enables performance profiling. Can be set via `ZEN_DEVELOPMENT_PROFILE` environment variable. Default is `false`.

## Configuration Files

Configuration files are written in YAML format. The configuration file will be created automatically when setting values if it doesn't exist.

The file structure follows the key hierarchy:

```yaml
log_level: info
log_format: text
cli:
  no_color: false
  verbose: false
  output_format: text
workspace:
  root: /path/to/workspace
  config_file: zen.yaml
development:
  debug: false
  profile: false
```

## Environment Variables

All configuration variables can be overridden using environment variables. The environment variable name is constructed by:
1. Adding the `ZEN_` prefix
2. Converting to uppercase
3. Replacing dots with underscores

Examples:
- `log_level` → `ZEN_LOG_LEVEL`
- `cli.no_color` → `ZEN_CLI_NO_COLOR`
- `workspace.root` → `ZEN_WORKSPACE_ROOT`

## Examples

```bash
# Display current configuration
$ zen config
Zen CLI Configuration

log_level: info
log_format: text
cli:
  no_color: false
  verbose: false
  output_format: text
...

# Get a specific configuration value
$ zen config get log_level
info

# Get nested configuration value
$ zen config get cli.output_format
text

# Set log level to debug
$ zen config set log_level debug
✓ Set log_level to "debug"
Configuration saved to .zen/config.yaml

# Set nested configuration value
$ zen config set cli.no_color true
✓ Set cli.no_color to "true"
Configuration saved to .zen/config.yaml

# List all configuration keys and values
$ zen config list
log_level=debug
log_format=text
cli.no_color=true
cli.verbose=false
cli.output_format=text
workspace.root=/Users/username/project
workspace.config_file=zen.yaml
development.debug=false
development.profile=false

# List with verbose output showing sources
$ zen config list --verbose
log_level=debug (from: config file)
log_format=text (from: default)
cli.no_color=true (from: config file)
cli.verbose=false (from: default)
cli.output_format=text (from: default)
workspace.root=/Users/username/project (from: workspace)
workspace.config_file=zen.yaml (from: default)
development.debug=false (from: default)
development.profile=false (from: default)

# Use environment variables
$ ZEN_LOG_LEVEL=debug zen config get log_level
debug

# Output as JSON for scripting
$ zen config list --output json
{
  "log_level": "info",
  "log_format": "text",
  "cli": {
    "no_color": false,
    "verbose": false,
    "output_format": "text"
  },
  "workspace": {
    "root": "/Users/username/project",
    "config_file": "zen.yaml"
  },
  "development": {
    "debug": false,
    "profile": false
  }
}

# Output as YAML
$ zen config list --output yaml
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

`0` Success

`1` Invalid arguments or general error

`2` Configuration key not found (for `get` subcommand)

`3` Invalid configuration value

`4` Permission denied writing configuration file

## Files

`.zen/config.yaml`

Workspace-specific configuration file

`~/.zen/config.yaml`

User-specific configuration file

## Notes

**Value Types:**
- Boolean values accept: `true`, `false`, `yes`, `no`, `1`, `0` (case-insensitive)
- String values are used as-is
- Enum values are validated against allowed options

**Security:**
- Sensitive values (API tokens, passwords) are automatically redacted in output
- Configuration files are created with appropriate permissions (0644)
- Directory permissions are set to 0755

**Validation:**
- All configuration keys and values are validated before saving
- Invalid keys show available options
- Invalid values show valid choices for enum types
- Type validation ensures boolean values are properly formatted

**Precedence:**
- Command-line flags override all other sources
- Environment variables override configuration files
- Configuration files are checked in order (workspace, then user)
- Default values are used when no other source provides a value

## See Also

* [zen init](zen-init.md) - Initialize workspace
* [zen status](zen-status.md) - Display workspace status
