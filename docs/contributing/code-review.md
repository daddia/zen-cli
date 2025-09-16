# Code Review Process

This guide covers how to submit pull requests and conduct code reviews for Zen.

## Pull Request Process

### Before Opening a PR

#### Pre-flight Checklist
- [ ] Tests pass locally (`make test-unit`)
- [ ] Code is formatted (`make fmt`)
- [ ] Linter passes (`make lint`)
- [ ] Documentation updated
- [ ] Commits are logical and clean
- [ ] Branch is up-to-date with main

#### PR Size Guidelines

Keep PRs focused and reviewable:
- **Ideal**: < 400 lines changed
- **Acceptable**: < 800 lines changed
- **Requires justification**: > 800 lines

Split large changes into multiple PRs when possible.

### Creating a Pull Request

#### PR Title

Follow Conventional Commits format:

```
feat(cli): add progress indicators for long operations
fix(config): resolve Windows path handling issue
docs(api): update authentication examples
```

#### PR Description Template

```markdown
## Description
Brief summary of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change)
- [ ] New feature (non-breaking change)
- [ ] Breaking change
- [ ] Documentation update

## Changes Made
- List specific changes
- Highlight important decisions
- Note any side effects

## Testing
- [ ] Unit tests pass
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] Tests added/updated
- [ ] Breaking changes documented

Fixes #(issue number)
```

### PR Best Practices

#### Do's
- Link related issues
- Provide context for reviewers
- Include screenshots for UI changes
- Respond to feedback promptly
- Keep discussion professional
- Update PR based on feedback

#### Don'ts
- Don't mix unrelated changes
- Don't ignore CI failures
- Don't force push after review starts
- Don't merge without approval
- Don't leave unresolved comments

## Review Process

### For Authors

#### Preparing for Review

1. **Self-review first**
   - Check your own diff
   - Remove debug code
   - Verify no secrets included

2. **Provide context**
   - Explain non-obvious decisions
   - Link to relevant documentation
   - Highlight areas needing attention

3. **Make it easy to review**
   - Order commits logically
   - Keep related changes together
   - Separate refactoring from features

#### Responding to Feedback

```markdown
# Acknowledge feedback
Thanks for catching that! Fixed in commit abc123.

# Explain decisions
I chose this approach because...

# Ask for clarification
Could you elaborate on the performance concern?

# Propose alternatives
Would it be better if we...?
```

### For Reviewers

#### Review Priorities

1. **Correctness** - Does it work as intended?
2. **Security** - Are there vulnerabilities?
3. **Performance** - Will it scale?
4. **Maintainability** - Is it easy to understand?
5. **Style** - Does it follow conventions?

#### Review Checklist

##### Architecture
- [ ] Follows established patterns
- [ ] No unnecessary complexity
- [ ] Proper abstraction levels
- [ ] Clear separation of concerns

##### Code Quality
- [ ] Clear variable/function names
- [ ] No code duplication
- [ ] Appropriate error handling
- [ ] Adequate logging

##### Testing
- [ ] Tests cover new functionality
- [ ] Edge cases considered
- [ ] No flaky tests introduced
- [ ] Coverage maintained or improved

##### Documentation
- [ ] Code comments where needed
- [ ] API documentation updated
- [ ] User documentation current
- [ ] ADRs for significant changes

#### Providing Feedback

##### Constructive Comments

Good feedback is:
- **Specific** - Point to exact issues
- **Actionable** - Suggest improvements
- **Respectful** - Focus on code, not person

```markdown
# Good
"This function could be simplified using the existing helper in pkg/utils"

# Poor
"This code is messy"

# Good
"Consider handling the error case where config is nil"

# Poor
"Wrong error handling"
```

##### Comment Types

Use conventional prefixes:

```markdown
# Must be fixed before merge
🔴 REQUIRED: Security issue - SQL injection vulnerability

# Should be addressed
🟡 SUGGESTION: Consider extracting this to a separate function

# Optional improvement
🟢 NITPICK: Typo in comment

# Just a comment
💭 QUESTION: Why did you choose this approach?

# Positive feedback
👍 PRAISE: Great test coverage!
```

## CI/CD Integration

### Required Checks

All PRs must pass:
- **Build** - Code compiles
- **Tests** - All tests pass
- **Lint** - No linter errors
- **Coverage** - Minimum thresholds met
- **Security** - No vulnerabilities
- **Docs** - Documentation builds

### Handling CI Failures

```bash
# Check specific failure
# Click "Details" link in PR

# Reproduce locally
make test-unit
make lint
make build

# Fix and push
git commit -am "fix: address CI failures"
git push
```

## Approval and Merge

### Approval Requirements

- At least 1 approval from maintainer
- All CI checks passing
- All comments resolved
- No merge conflicts

### Merge Strategy

We use **squash and merge** to keep history clean:

1. Ensure PR title follows conventions
2. Clean up commit message
3. Include co-authors if applicable
4. Delete branch after merge

### Post-Merge

After your PR is merged:
1. Delete your local branch
2. Pull latest main
3. Celebrate your contribution!

## Review Etiquette

### For Everyone

- Be respectful and professional
- Assume positive intent
- Focus on the code, not the person
- Acknowledge good work
- Be patient with new contributors

### Response Time Expectations

- **Initial review**: Within 2 business days
- **Follow-up reviews**: Within 1 business day
- **Critical fixes**: ASAP (tag maintainers)

### Handling Disagreements

1. Discuss in PR comments
2. Provide evidence/examples
3. Escalate to maintainers if needed
4. Consider ADR for significant decisions

## Special Scenarios

### Breaking Changes

For breaking changes:
1. Discuss in issue first
2. Document migration path
3. Update CHANGELOG
4. Consider deprecation period

### Security Fixes

For security issues:
1. Don't disclose details publicly
2. Contact maintainers privately
3. Use draft PR if needed
4. Fast-track review process

### Documentation-Only Changes

Documentation PRs:
- Can skip some CI checks
- Still need review for accuracy
- Consider impact on users

### Refactoring

For refactoring PRs:
- Separate from feature changes
- Maintain exact functionality
- Include comprehensive tests
- Explain motivation

## Tools and Automation

### PR Commands

Available bot commands:

```
/retest - Re-run failed CI checks
/approve - Approve PR (maintainers)
/lgtm - Looks good to me
/hold - Prevent automatic merge
```

### Review Tools

Recommended tools:
- GitHub PR interface
- VS Code GitHub Pull Requests extension
- gh CLI for command line
- Refined GitHub browser extension

## Next Steps

- Learn about [Release Process](release-process.md)
- Understand [Architecture](architecture.md)
- Review [Testing](testing.md) requirements

---

Questions? Tag `@maintainers` in your PR or open a discussion.
