# Documentation Maintenance Guide

Guidelines for maintaining professional-grade documentation for Zen CLI.

## Overview

This guide ensures documentation remains current, accurate, and consumer-ready as the codebase evolves. It establishes processes for documentation updates, quality assurance, and release preparation.

## Maintenance Principles

### 1. Documentation-as-Code

- **Version controlled** - All documentation lives in the repository
- **Reviewed like code** - Documentation changes go through PR review
- **Automated validation** - Linting and link checking in CI/CD
- **Release synchronized** - Documentation versions match software releases

### 2. Single Source of Truth

- **Auto-generated content** - Command documentation generated from code
- **Manual documentation** - Architecture, user guides, design system
- **Clear boundaries** - Know what's generated vs. manually maintained
- **Consistent updates** - Keep manual docs synchronized with code changes

### 3. User-Centric Quality

- **Consumer-grade standards** - Documentation ready for public consumption
- **Multiple audiences** - End users, developers, architects, product managers
- **Progressive disclosure** - Start simple, add complexity gradually
- **Task-oriented** - Focus on what users want to accomplish

## Content Categories

### Auto-Generated Documentation

**Location**: `docs/zen/`
**Source**: Command implementations in `pkg/cmd/`
**Maintenance**: Automatic via build process

**Content includes**:
- Command reference (`zen_*.md`)
- Flag documentation
- Usage examples
- Help text

**Process**:
1. Update command implementations
2. Run documentation generation
3. Review generated output
4. Commit generated files

### Architecture Documentation

**Location**: `docs/architecture/`
**Source**: Manual maintenance
**Maintenance**: Update with significant architectural changes

**Content includes**:
- System overview and principles
- Component descriptions
- Architecture Decision Records (ADRs)
- Design patterns and integration guides

**Update triggers**:
- New components or services
- Significant refactoring
- Technology stack changes
- Integration pattern changes

### User Documentation

**Location**: `docs/user-guide/`, `docs/getting-started/`
**Source**: Manual maintenance
**Maintenance**: Update with feature changes

**Content includes**:
- Installation and setup guides
- Workflow documentation
- Configuration reference
- Troubleshooting guides

**Update triggers**:
- New features or commands
- Configuration changes
- Workflow modifications
- Integration updates

### Design System

**Location**: `docs/design/`
**Source**: Manual maintenance synchronized with implementation
**Maintenance**: Update with UI/UX changes

**Content includes**:
- Design foundations and principles
- Color and typography systems
- Component guidelines
- Implementation examples

**Update triggers**:
- CLI output format changes
- Color or symbol updates
- New formatting functions
- Accessibility improvements

## Maintenance Workflows

### Regular Maintenance (Weekly)

```bash
# 1. Check for broken links
make docs-lint

# 2. Validate markdown formatting
markdownlint docs/

# 3. Update auto-generated docs if needed
make docs-generate

# 4. Review recent code changes for doc impacts
git log --since="1 week ago" --oneline pkg/ internal/
```

### Feature Development Workflow

When adding new features:

1. **Code implementation** with inline documentation
2. **Update relevant user guides** with new workflows
3. **Add architecture documentation** if new components
4. **Update design system** if UI changes
5. **Generate command docs** if new commands
6. **Review and test** all documentation changes

### Release Preparation

Before each release:

1. **Regenerate command documentation**
2. **Update version references** in documentation
3. **Review and update** getting started guides
4. **Validate all links** and references
5. **Test installation instructions** on clean systems
6. **Update changelog** with documentation changes

## Quality Assurance

### Automated Checks

**Linting** (`.markdownlint.yml`):
```yaml
# Enforce consistent formatting
MD013: false  # Allow long lines for code examples
MD033: false  # Allow inline HTML for tables
MD041: false  # Allow missing H1 in fragments
```

**Link validation**:
```bash
# Check internal links
make docs-check-links

# Validate external references
make docs-validate-external
```

**Spell checking**:
```bash
# Run spell check on documentation
make docs-spell-check
```

### Manual Review Checklist

For each documentation update:

- [ ] **Accuracy** - Information matches current implementation
- [ ] **Completeness** - All necessary information included
- [ ] **Clarity** - Language is clear and unambiguous
- [ ] **Consistency** - Follows established patterns and style
- [ ] **Examples** - Working examples provided where helpful
- [ ] **Links** - All internal and external links work
- [ ] **Formatting** - Proper markdown and consistent styling
- [ ] **Accessibility** - Screen reader friendly structure

### User Testing

Quarterly user testing:

1. **New user onboarding** - Test installation and quick start
2. **Feature discovery** - Verify users can find and use features
3. **Troubleshooting** - Test error scenarios and help content
4. **Advanced workflows** - Validate complex use cases

## Content Standards

### Writing Style

- **Clear and concise** - Avoid unnecessary complexity
- **Active voice** - Use active rather than passive voice
- **Present tense** - Write in present tense
- **Second person** - Address the user directly ("you")
- **Consistent terminology** - Use the same terms throughout

### Code Examples

- **Working examples** - All code examples must work
- **Complete context** - Provide necessary setup/context
- **Expected output** - Show what users should expect
- **Error cases** - Include common error scenarios
- **Copy-pasteable** - Format for easy copying

### Visual Elements

- **Consistent formatting** - Use established patterns
- **Semantic markup** - Use appropriate markdown elements
- **Tables for structure** - Organize complex information
- **Code blocks** - Proper syntax highlighting
- **Callouts** - Use for important information

## Update Triggers

### Code Changes That Require Documentation Updates

**New commands or subcommands**:
- Update user guide workflows
- Add to command reference (auto-generated)
- Update getting started if relevant

**Configuration changes**:
- Update configuration reference
- Update user guide examples
- Update troubleshooting if needed

**Integration changes**:
- Update architecture documentation
- Update user guide integration sections
- Update authentication guides

**Output format changes**:
- Update design system documentation
- Update user guide examples
- Update scripting examples

### External Changes

**Dependency updates**:
- Update installation requirements
- Update compatibility information
- Test and update examples

**Platform support changes**:
- Update installation instructions
- Update platform-specific documentation
- Update troubleshooting guides

## Tools and Automation

### Documentation Generation

```bash
# Generate command documentation
make docs-generate

# Build documentation site
make docs-build

# Serve documentation locally
make docs-serve
```

### Validation Tools

```bash
# Lint markdown files
make docs-lint

# Check for broken links
make docs-check-links

# Validate code examples
make docs-test-examples

# Full documentation validation
make docs-validate
```

### CI/CD Integration

Documentation validation in continuous integration:

```yaml
# .github/workflows/docs.yml
name: Documentation
on: [push, pull_request]
jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Lint documentation
        run: make docs-lint
      - name: Check links
        run: make docs-check-links
      - name: Test examples
        run: make docs-test-examples
```

## Metrics and Monitoring

### Documentation Health Metrics

- **Coverage** - Percentage of features documented
- **Freshness** - Time since last update for each section
- **Accuracy** - Number of reported documentation issues
- **Usability** - User feedback and support ticket analysis

### Regular Reviews

**Monthly**:
- Review documentation metrics
- Analyze user feedback and support tickets
- Identify gaps or outdated content
- Plan updates and improvements

**Quarterly**:
- Comprehensive documentation audit
- User experience testing
- Architecture documentation review
- Process improvement evaluation

## Troubleshooting

### Common Issues

**Broken links after refactoring**:
```bash
# Find and fix broken internal links
make docs-check-links
# Update references in affected files
```

**Outdated examples**:
```bash
# Test all code examples
make docs-test-examples
# Update failing examples
```

**Inconsistent formatting**:
```bash
# Run markdown linter
make docs-lint
# Fix formatting issues
```

### Getting Help

- **Documentation issues** - Create GitHub issue with `documentation` label
- **Style questions** - Refer to this maintenance guide
- **Technical writing** - Consult team technical writer
- **User feedback** - Monitor GitHub discussions and support channels

## See Also

- **[Contributing Guide](../contributing/README.md)** - General contribution guidelines
- **[Design System](../design/README.md)** - Documentation formatting standards
- **[Architecture](../architecture/README.md)** - Technical documentation structure
- **[User Guide](../user-guide/README.md)** - Example of user-focused documentation
