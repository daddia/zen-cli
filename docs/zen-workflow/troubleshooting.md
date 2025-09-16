# Zenflow Troubleshooting Guide

This guide helps you diagnose and resolve common issues with Zenflow. Find your problem below or use the diagnostic commands to identify issues.

## Quick Diagnostics

### Check System Health

```bash
# Overall system check
zen doctor
# ✓ CLI version: 0.2.1
# ✓ Workspace: configured
# ⚠ Integration: GitHub (token expired)
# ✓ Dependencies: all installed
# ✗ Quality gates: 2 failing

# Detailed diagnostics
zen doctor --verbose
zen doctor --component "integrations"
zen doctor --fix  # Attempt auto-fix
```

### Check Workflow Status

```bash
# Current workflow state
zen status --detailed
# Workflow: Payment Feature
# Stage: Build (blocked)
# Blockers: 2 failing quality gates
# - Test coverage below threshold (78% < 80%)
# - Security scan found vulnerabilities

# Get resolution suggestions
zen suggest --fix
```

## Common Issues and Solutions

### Installation and Setup Issues

#### Problem: Command 'zen' not found

**Symptoms:**
```bash
$ zen version
bash: zen: command not found
```

**Solutions:**
```bash
# Solution 1: Add to PATH
export PATH=$PATH:/usr/local/bin/zen
echo 'export PATH=$PATH:/usr/local/bin/zen' >> ~/.bashrc

# Solution 2: Reinstall
curl -sSL https://get.zen.dev | sh

# Solution 3: Use full path
/usr/local/bin/zen version
```

#### Problem: Workspace not initialized

**Symptoms:**
```bash
zen align init "Feature"
# Error: No workspace found. Run 'zen init' first
```

**Solutions:**
```bash
# Initialize workspace
zen init

# Or specify workspace path
zen --workspace /path/to/project align init "Feature"

# Set default workspace
export ZEN_WORKSPACE=/path/to/project
```

#### Problem: Configuration not found

**Symptoms:**
```bash
zen config show
# Error: Config file not found at .zen/config.yaml
```

**Solutions:**
```bash
# Create default config
zen config init

# Restore from backup
cp .zen/config.yaml.backup .zen/config.yaml

# Use template config
zen config --template "standard"
```

### Stage Progression Issues

#### Problem: Cannot progress to next stage

**Symptoms:**
```bash
zen design start
# Error: Cannot start Design stage
# Reason: Discover stage incomplete
# Missing: Risk documentation, test strategy
```

**Solutions:**
```bash
# Complete missing requirements
zen discover status --incomplete
# Missing items:
# - Risk assessment
# - Test strategy

# Complete requirements
zen discover risk --add "Performance risk" --mitigation "Load testing"
zen discover test-strategy --define

# Force progression (not recommended)
zen design start --force --reason "Urgent deadline"
```

#### Problem: Stage marked complete but isn't

**Symptoms:**
```bash
zen status
# Align: ✓ Complete
# But metrics aren't defined
```

**Solutions:**
```bash
# Reset stage status
zen stage reset "align"

# Revalidate stage
zen validate --stage "align" --strict

# Fix data inconsistency
zen repair --stage "align"
```

### Quality Gate Failures

#### Problem: Test coverage below threshold

**Symptoms:**
```bash
zen build validate
# ✗ Test coverage: 72% (minimum: 80%)
```

**Solutions:**
```bash
# Identify untested code
zen coverage report --uncovered
# Uncovered files:
# - src/utils/helpers.js (45%)
# - src/api/auth.js (60%)

# Generate test templates
zen test generate --for "src/utils/helpers.js"

# Temporarily lower threshold (with approval)
zen config gates.build.coverage --set 70 --temporary --approve "@tech-lead"

# Exclude files from coverage
echo "src/generated/*" >> .covignore
```

#### Problem: Security vulnerabilities detected

**Symptoms:**
```bash
zen ship security --scan
# ✗ 2 high severity vulnerabilities found
# - CVE-2024-1234 in dependency X
# - SQL injection risk in module Y
```

**Solutions:**
```bash
# Get detailed vulnerability info
zen security details "CVE-2024-1234"

# Update dependencies
zen deps update --security

# Apply security patches
zen security patch --auto

# Accept risk (with justification)
zen security accept-risk "CVE-2024-1234" \
  --reason "Not exploitable in our usage" \
  --approver "@security-team"
```

#### Problem: Linting errors blocking progress

**Symptoms:**
```bash
zen build validate
# ✗ Linting: 47 errors found
```

**Solutions:**
```bash
# Auto-fix where possible
zen lint --fix

# Show remaining issues
zen lint --errors-only

# Configure linting rules
zen config lint --rule "no-console" --level "warning"

# Ignore specific files
echo "src/debug.js" >> .lintignore
```

### Integration Issues

#### Problem: GitHub integration not working

**Symptoms:**
```bash
zen build review --create
# Error: GitHub API error: 401 Unauthorized
```

**Solutions:**
```bash
# Check token status
zen integrate github --test
# Token expired

# Update token
zen integrate github --token "ghp_xxxxxxxxxxxx"

# Or use environment variable
export ZEN_GITHUB_TOKEN="ghp_xxxxxxxxxxxx"

# Verify permissions
zen integrate github --check-permissions
# Need: repo, read:org, write:issues
```

#### Problem: Jira sync failing

**Symptoms:**
```bash
zen integrate jira --sync
# Error: Failed to sync with Jira
# Status mapping not found
```

**Solutions:**
```bash
# Configure status mapping
zen integrate jira --map-status \
  --zen "align" --jira "Discovery" \
  --zen "build" --jira "In Progress"

# Test connection
zen integrate jira --test

# Re-authenticate
zen integrate jira --auth --email "user@company.com"

# Debug sync issues
zen integrate jira --debug --verbose
```

### Performance Issues

#### Problem: Commands running slowly

**Symptoms:**
```bash
time zen status
# real    0m15.234s  # Too slow!
```

**Solutions:**
```bash
# Profile slow operations
zen profile --command "status"
# Bottleneck: metrics calculation (12s)

# Enable caching
zen config cache --enable
zen config cache --ttl 300  # 5 minutes

# Optimize database
zen db optimize

# Disable expensive features
zen config features --disable "real-time-metrics"
```

#### Problem: High memory usage

**Symptoms:**
```bash
# System becomes unresponsive during zen operations
top
# zen process using 4GB RAM
```

**Solutions:**
```bash
# Limit memory usage
zen config limits --memory "1GB"

# Clear cache
zen cache clear

# Reduce batch sizes
zen config batch-size --set 100

# Use streaming mode for large operations
zen report --streaming
```

### Data and State Issues

#### Problem: Corrupted workflow state

**Symptoms:**
```bash
zen status
# Error: Invalid workflow state
# State file corrupted
```

**Solutions:**
```bash
# Restore from backup
zen restore --latest
zen restore --timestamp "2024-01-15T10:30:00"

# Rebuild state from history
zen repair --rebuild-state

# Reset to last known good state
zen reset --to-last-valid

# Start fresh (last resort)
zen reset --hard --confirm
```

#### Problem: Lost work or data

**Symptoms:**
```bash
# Work from yesterday is missing
zen discover research --list
# No results (but had 10 items yesterday)
```

**Solutions:**
```bash
# Check backups
zen backup list
zen backup restore --date "yesterday"

# Review audit log
zen audit --date "yesterday" --filter "discover"

# Recover from version control
git log -- .zen/
git checkout <commit> -- .zen/data/

# Enable auto-backup going forward
zen config backup --auto --interval "1h"
```

### Deployment Issues

#### Problem: Deployment rollback not working

**Symptoms:**
```bash
zen ship rollback
# Error: Cannot rollback, no previous version found
```

**Solutions:**
```bash
# Check rollback history
zen ship rollback --list

# Manual rollback
zen ship deploy --version "v1.2.3"

# Fix rollback configuration
zen config rollback --retention 5
zen config rollback --test

# Emergency rollback
zen ship rollback --emergency --force
```

#### Problem: Canary deployment stuck

**Symptoms:**
```bash
zen ship canary --status
# Canary at 5% for 6 hours (expected: 1 hour)
```

**Solutions:**
```bash
# Check canary metrics
zen ship canary --metrics
# Error rate slightly elevated but below threshold

# Force progression
zen ship canary --progress --force

# Or rollback canary
zen ship canary --abort

# Adjust canary criteria
zen config canary --success-criteria \
  --error-rate "<0.5%" \
  --duration "30m"
```

## Debugging Commands

### Enable Debug Mode

```bash
# Verbose output
zen --verbose <command>
zen --debug <command>

# Trace mode (very detailed)
zen --trace align init "Feature"

# Log to file
zen --log-file debug.log <command>
```

### Analyze Logs

```bash
# View recent logs
zen logs --tail 100
zen logs --since "1 hour ago"

# Filter logs
zen logs --level "error"
zen logs --component "quality-gates"
zen logs --stage "build"

# Export logs
zen logs export --format "json" --output "zen-logs.json"
```

### Check Dependencies

```bash
# Verify all dependencies
zen deps check
# git: ✓ 2.34.1
# docker: ✓ 20.10.12
# node: ✗ Not found (required for some features)

# Install missing dependencies
zen deps install --missing

# Update dependencies
zen deps update --all
```

## Error Messages Explained

### Authentication Errors

```bash
"Error: Authentication failed"
# Token expired or invalid
# Solution: Re-authenticate
zen auth login

"Error: Insufficient permissions"
# User lacks required role
# Solution: Request permissions
zen auth request-role "developer"
```

### Validation Errors

```bash
"Error: Validation failed: missing required field 'metrics'"
# Stage requirements not met
# Solution: Complete requirements
zen align metrics --add "Success rate" --target "95%"

"Error: Invalid configuration"
# Config file syntax error
# Solution: Validate and fix
zen config validate
zen config repair
```

### Network Errors

```bash
"Error: Connection timeout"
# Network or firewall issue
# Solution: Check connectivity
zen network test
zen config proxy --set "http://proxy:8080"

"Error: API rate limit exceeded"
# Too many requests
# Solution: Wait or upgrade
zen status --rate-limit
zen config rate-limit --increase
```

## Getting Help

### Self-Service Resources

```bash
# Built-in help
zen help
zen help <command>
zen help --examples

# Show command examples
zen examples align
zen examples --scenario "hotfix"

# Search documentation
zen docs search "quality gates"
zen docs --online
```

### Community Support

```bash
# Report issues
zen feedback --bug \
  --description "Deploy command fails" \
  --logs-attached

# Request features
zen feedback --feature \
  --description "Add support for GitLab"

# Join community
zen community --join
# Slack: https://zenflow.slack.com
# Forum: https://community.zen.dev
# GitHub: https://github.com/zenflow/zen
```

### Professional Support

```bash
# Check support status
zen support status
# Plan: Professional
# Tickets remaining: 8
# Response time: 4 hours

# Create support ticket
zen support ticket \
  --priority "high" \
  --issue "Production deployment blocked"

# Schedule call
zen support call --schedule
```

## Prevention Best Practices

### Regular Maintenance

```bash
# Weekly maintenance
zen maintenance weekly
- Clear old cache
- Optimize database
- Update dependencies
- Backup data

# Monthly health check
zen health check --comprehensive
zen health report --email "team@company.com"
```

### Monitoring Setup

```bash
# Configure alerts
zen monitor alerts \
  --workflow-stuck ">2h" \
  --quality-gate-failures ">3" \
  --integration-errors "any"

# Dashboard setup
zen dashboard create \
  --metrics "cycle-time,quality-score,velocity" \
  --refresh "5m"
```

### Backup Strategy

```bash
# Automated backups
zen backup schedule --daily --retain 30
zen backup schedule --weekly --retain 12
zen backup test --restore --verify

# Before risky operations
zen backup create --label "before-upgrade"
zen upgrade
# If issues occur:
zen backup restore --label "before-upgrade"
```

## Recovery Procedures

### Disaster Recovery

```bash
# Complete system failure
# 1. Restore from backup
zen restore --emergency --source "backup-server"

# 2. Verify integrity
zen verify --all

# 3. Resync integrations
zen integrate sync --all

# 4. Notify team
zen notify team --message "System restored"
```

### Data Recovery

```bash
# Accidental deletion
# 1. Check recycle bin
zen recover --recent-deletes

# 2. Restore from backup
zen backup restore --selective --data "discover/*"

# 3. Rebuild from sources
zen rebuild --from-integrations
```

### State Recovery

```bash
# Inconsistent state
# 1. Export current state
zen state export --backup

# 2. Reset to known good state
zen state reset --to-checkpoint

# 3. Replay recent actions
zen replay --from "checkpoint" --to "now"
```

## Performance Tuning

### Optimize for Speed

```bash
# Database optimization
zen db optimize --vacuum --reindex

# Cache configuration
zen config cache \
  --size "512MB" \
  --strategy "lru" \
  --ttl 600

# Parallel processing
zen config parallel \
  --workers 4 \
  --batch-size 100
```

### Reduce Resource Usage

```bash
# Limit resource consumption
zen config limits \
  --cpu "2 cores" \
  --memory "2GB" \
  --disk-io "100MB/s"

# Disable unnecessary features
zen features disable \
  --real-time-sync \
  --auto-suggest \
  --preview-generation
```

## FAQ

### Why is my workflow stuck?

Check for:
1. Incomplete quality gates: `zen validate --current`
2. Missing approvals: `zen approvals --pending`
3. Integration issues: `zen integrate test --all`
4. Resource constraints: `zen capacity check`

### How do I undo an action?

```bash
# Undo last action
zen undo

# Undo specific action
zen history
zen undo --action-id "abc123"

# Revert to timestamp
zen revert --to "2024-01-15T10:00:00"
```

### Can I skip a stage?

Not recommended, but possible with approval:
```bash
zen skip --stage "discover" \
  --reason "Re-implementation of existing feature" \
  --approver "@tech-lead"
```

### How do I migrate from another tool?

```bash
# Import data
zen import --from "jira" --project "PROJ"
zen import --from "github" --repo "org/repo"

# Map workflows
zen migrate map \
  --source-workflow "old-process" \
  --target-workflow "zenflow"

# Validate migration
zen migrate validate --dry-run
zen migrate execute
```

## Summary

Most Zenflow issues can be resolved by:

1. **Running diagnostics**: `zen doctor`
2. **Checking status**: `zen status --detailed`
3. **Reviewing logs**: `zen logs --recent`
4. **Validating configuration**: `zen config validate`
5. **Updating components**: `zen update`

If issues persist:
- Check documentation: `zen docs`
- Search community forums: https://community.zen.dev
- Contact support: `zen support ticket`

Remember: Regular maintenance and backups prevent most issues.
