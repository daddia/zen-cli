---
title: zen-<command>
slug: /cli/zen-<command>
section: "CLI Reference"
man_section: 1
since: v0.0.0
aliases: [ "zen <alias>", "zen <parent> <command>" ]
see_also:
  - zen-help
  - zen-<related-command>
---

# zen \<command>

```sh
zen <command> [subcommand] [flags] [args]
```

\<Explain what the command does, when to use it, any important context, and any side effects.>

## OPTIONS

<!-- Prefer terse bullets; group related flags. Keep descriptions action-led. -->

* `-h`, `--help` — Show help and exit.
* `-q`, `--quiet` — Suppress non-error output.
* `-v`, `--verbose` — Show detailed output and progress.
* `--format <text|json|yaml>` — Output format (default: `text`).
* `--no-color` — Disable colored output.
* \<add command-specific flags here>

## ARGUMENTS

* `<required>` — \<what it is / allowed values>.
* `[optional]` — \<what it is / default>.

## ENVIRONMENT

* `ZEN_CONFIG` — Path to default config file.
* `ZEN_TOKEN` — API token used when not provided via config.
* `HTTP_PROXY`, `HTTPS_PROXY` — Proxy settings honored by network calls.

## EXAMPLES

```sh
# Basic usage
zen <command> <required>

# Quiet mode with JSON output for scripting
zen <command> --quiet --format json > result.json

# Verbose output with progress indicators
zen <command> --verbose <required>

# Disable colors (useful for CI/CD or accessibility)
zen <command> --no-color <required>

# Use explicit config
ZEN_CONFIG=~/.zen/config.yaml zen <command>
```

## EXIT STATUS

`0` success; `1` general error; `2` invalid args; `3` auth error.

## NOTES

\<Performance caveats, pagination, rate limits, idempotency, retries, etc.>

**Output Formatting:**

* Uses color and Unicode symbols (✓, ✗, !, -, +) to indicate status
* Supports machine-readable output when piped or using `--format json`
* Progress indicators shown for long-running operations

**Accessibility:**

* Screen reader compatible with proper punctuation for pauses
* `--no-color` flag removes all color formatting

## SEE ALSO

* [zen help](zen-help.md)
* [zen \<related>](zen-<related>.md)
