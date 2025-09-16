# Zen CLI Documentation

This directory contains auto-generated CLI documentation for the Zen CLI.

## Documentation Structure

Files are automatically generated from the Cobra command definitions:
- `index.md` - Auto-generated index of all commands
- `zen.md` - Main command reference
- `zen_config.md` - Config command group
- `zen_config_get.md` - Config get subcommand
- `zen_config_list.md` - Config list subcommand
- `zen_config_set.md` - Config set subcommand
- `zen_init.md` - Init command
- `zen_status.md` - Status command
- `zen_version.md` - Version command
- etc.

These files are generated using the Cobra documentation generator and stay in sync with the actual command implementations.

## Generating Documentation

To regenerate the auto-generated documentation:

```bash
# Using Make (if available)
make docs

# Or directly with Go
go run internal/tools/docgen/main.go -out ./docs/zen -format markdown -frontmatter

# Generate all formats
make docs-all  # Markdown, Man pages, and ReStructuredText

# Clean generated docs (preserves this README.md)
make docs-clean
```

## Documentation Features

### LLM-Optimized Structure
The auto-generated documentation follows best practices for LLM consumption:
- One command per file for easy chunking
- Auto-generated index file with command organization
- Structured front matter with metadata
- Clear sections (Synopsis, Examples, Options, etc.)
- Stable, predictable filenames
- Rich examples for each command

### Index File
The `index.md` file is automatically generated and provides:
- Organized command listing (Core, Future, Shell Completion)
- Direct links to all command documentation
- Subcommand listings for commands with children
- Update timestamps

### Front Matter
Each generated file includes YAML front matter suitable for static site generators:
- `title` - Command name
- `slug` - URL-friendly identifier
- `description` - Brief description for search/SEO
- `section` - Documentation section
- `man_section` - Man page section number
- `date` - Generation date
- `keywords` - Search keywords

### Multiple Output Formats
The documentation generator supports:
- **Markdown** - For documentation sites and GitHub
- **Man pages** - For Unix/Linux manual system
- **ReStructuredText** - For Sphinx documentation

## Integration with CI/CD

The documentation generation can be integrated into CI/CD pipelines:

```bash
# Check if docs are up-to-date
make docs-check

# This will fail if documentation needs regeneration
# Perfect for CI checks to ensure docs stay in sync
```

## Benefits

1. **Always in Sync** - Documentation automatically reflects command changes
2. **Auto-Generated Index** - Index file automatically created with proper organization
3. **No Manual Updates** - Examples and flags update automatically
4. **Consistent Format** - All commands follow the same structure
5. **LLM-Ready** - Optimized for AI tool consumption
6. **Multiple Formats** - Support different documentation systems
7. **Version Control** - Track documentation changes with code changes