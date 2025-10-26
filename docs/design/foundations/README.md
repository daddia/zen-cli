# Design Foundations

Core design concepts and constraints that create a cohesive Terminal experience for Zen CLI users.

## Overview

Zen CLI follows established design principles to ensure consistency, usability, and accessibility across all commands and interactions. These foundations are implemented in code and enforced through design guidelines.

## Language Structure

### Command Pattern

All Zen commands follow this consistent structure:

```bash
zen <command> <subcommand> [value] [flags]
```

| Component | Description | Example |
|-----------|-------------|---------|
| **Command** | The object you want to interact with | `task`, `config`, `assets` |
| **Subcommand** | The action you want to take | `create`, `list`, `get`, `set` |
| **Value** | Arguments passed to commands/flags | `PROJ-123`, `"Task Title"` |
| **Flags** | Modifiers that change command behavior | `--verbose`, `--output json` |

### Language Guidelines

- **Use clear, unambiguous language** that cannot be misconstrued
- **Prefer shorter phrases** when appropriate and understood
- **Use understood shorthands** to save typing (`repo` vs `repository`)
- **Flags modify actions**, not separate commands

#### Examples

✓ **Do**: Use flags for action modifiers

```bash
zen task create PROJ-123 --title "New feature"
```

✗ **Don't**: Make modifiers separate commands

```bash
zen task title PROJ-123 "New feature"  # Avoid this
```

## Visual Hierarchy

### Typography

Everything in CLI is text, so hierarchy uses font weight and spacing:

- **Bold text** for headers and important information
- **Normal weight** for regular content
- **Monospace fonts** create inherent visual order
- **No italics** (limited terminal support)

### Spacing

Create hierarchy and visual rhythm using:

- **Line breaks** for section separation
- **Tables** for structured data
- **Indentation** (2 spaces per level) for nested content

✓ **Do**: Use space for legible output

```text
Workspace Status
================
✓ Configuration: Valid
  - Log level: info
  - Output format: text
✓ Assets: Ready
  - Library: Synced
  - Cache: 42MB
```

✗ **Don't**: Compress output without spacing

```text
Workspace Status
✓ Configuration: Valid
- Log level: info
- Output format: text
✓ Assets: Ready
- Library: Synced
- Cache: 42MB
```

## Color System

### ANSI Color Support

Zen uses the 8 basic ANSI colors for reliable terminal support:

| Color | Usage | Implementation |
|-------|-------|----------------|
| **Green** | Success, positive states | `ColorGreen` |
| **Red** | Errors, failures | `ColorRed` |
| **Yellow** | Warnings, alerts | `ColorYellow` |
| **Blue** | Information, neutral | `ColorBlue` |
| **Cyan** | Neutral content | `ColorCyan` |
| **Bold** | Headers, emphasis | `ColorBold` |

### Semantic Color Functions

```go
// Success states
iostreams.ColorSuccess("Operation completed")
iostreams.FormatSuccess("Task created")

// Error states  
iostreams.ColorError("Failed to connect")
iostreams.FormatError("Invalid configuration")

// Warnings
iostreams.ColorWarning("Deprecated feature")
iostreams.FormatWarning("Cache limit exceeded")

// Information
iostreams.ColorInfo("Processing...")
iostreams.FormatNeutral("Status: pending")
```

### Color Guidelines

- **Only enhance meaning**, don't communicate meaning through color alone
- **Support `--no-color` flag** for accessibility and scripting
- **Users can customize** the 8 basic colors (this is expected)
- **Avoid 256-color sequences** (unreliable support)

## Iconography

### Unicode Symbol System

Zen uses consistent Unicode symbols with semantic meaning:

| Symbol | Meaning | Usage | Code |
|--------|---------|-------|------|
| `✓` | Success | Completed operations | `SymbolSuccess` |
| `✗` | Failure | Failed operations | `SymbolFailure` |
| `!` | Alert | Warnings, important notices | `SymbolAlert` |
| `-` | Neutral | Status, list items | `SymbolNeutral` |
| `+` | Changes | Modifications, additions | `SymbolChange` |

### Symbol Guidelines

- **Enhance meaning**, don't rely on symbols alone
- **Consider font support** - users have varying Unicode support
- **Consistent usage** across all commands
- **Fallback gracefully** when symbols aren't supported

#### Examples

✓ **Do**: Use appropriate symbols for context
```bash
✓ Task created successfully
✗ Failed to connect to Jira
! Configuration file not found
```

✗ **Don't**: Use wrong symbols for context
```bash
✓ Failed to connect to Jira  # Wrong symbol for failure
✗ Task created successfully  # Wrong symbol for success
```

## Output Formats

### Human-Readable Output (Default)

Optimized for terminal interaction:

- **Colors and formatting** for visual hierarchy
- **Symbols** for quick status recognition
- **Proper spacing** and indentation
- **Table layouts** for structured data
- **Headers and sections** for organization

### Machine-Readable Output

Optimized for scripting and automation:

- **No colors or styling** (`--no-color` implied)
- **Tab-delimited columns** for `cut` compatibility
- **No headers** in table output
- **Exact values** without truncation
- **Consistent format** across commands

#### Example Comparison

**Human output** (`zen status`):

```text
Workspace Status
================
✓ Configuration    Valid
✓ Assets          Ready (42MB cached)
! Integration     Jira connection timeout
```

**Machine output** (`zen status --output json`):

```json
{
  "configuration": {"status": "valid"},
  "assets": {"status": "ready", "cache_size": "42MB"},
  "integration": {"status": "error", "message": "Jira connection timeout"}
}
```

## Accessibility

### Screen Reader Support

- **Use punctuation** for pauses: periods (`.`), commas (`,`), colons (`:`)
- **Descriptive text** alongside symbols and colors
- **Logical reading order** with proper hierarchy
- **Alternative text** for visual elements

### Keyboard Navigation

- **Standard shortcuts** work in all terminals
- **No custom key bindings** that conflict with terminal/shell
- **Tab completion** support via `zen completion`

## Scriptability

### Design for Automation

- **Create flags for interactive features** (`--yes`, `--force`)
- **Clear language and defaults** for all options
- **Consistent exit codes** (0=success, 1=error, 2=cancel)
- **Machine output formats** for parsing

### Environment Variables

Support standard environment variables:

- `NO_COLOR` - Disable colors
- `ZEN_*` - Zen-specific configuration
- `EDITOR` - Text editor preference
- `PAGER` - Output paging preference

## Customization Support

### User Environment Awareness

- **Shell customization** - prompts, aliases, PATH
- **Terminal customization** - fonts, colors, shortcuts
- **OS customization** - language, accessibility settings

### Zen Customization

- **Configuration** via `zen config`
- **Aliases** for common commands
- **Environment variables** for preferences
- **Output formats** for different use cases

## Implementation

### Code Integration

Design foundations are implemented in:

- **`pkg/iostreams/colors.go`** - Color and symbol system
- **`pkg/cmd/root/root.go`** - Global flags and behavior
- **Command implementations** - Consistent patterns across all commands

### Validation

- **Automated tests** verify design consistency
- **Linting rules** enforce formatting standards
- **Code review** ensures adherence to guidelines

## Examples

### Status Command Output

```bash
$ zen status
Workspace Status
================
✓ Configuration: Valid
  - Log level: info
  - Output format: text
  - Config file: zen.yaml

✓ Assets: Ready
  - Library: Synced (2 hours ago)
  - Cache: 42MB (1,247 files)
  - Repository: github.com/zen-org/library

! Integration: Partial
  - Jira: Connected (PROJ)
  - GitHub: Authentication required
  - Slack: Not configured
```

### Error Handling

```bash
$ zen config get invalid_key
✗ Unknown configuration key "invalid_key"

Available keys: log_level, log_format

Try: zen config list
```

### Machine Output

```bash
$ zen assets list --output json | jq '.[] | select(.type=="template")'
```

## See Also

- **[Components](../components.md)** - UI component guidelines
- **[Getting Started](../getting-started.md)** - Design process
- **[Command Reference](../../zen/)** - Auto-generated command docs
- **[Architecture](../../architecture/)** - Technical implementation
