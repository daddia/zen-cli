# Zenflow Commands Reference

This reference documents all Zen CLI commands for the Zenflow workflow. Commands are organized by workflow stage and functionality.

## Command Structure

All Zenflow commands follow this pattern:

```bash
zen <stage> [subcommand] [options] [arguments]
```

### Global Options

These options work with all commands:

| Option | Description |
|--------|-------------|
| `--help, -h` | Show command help |
| `--verbose, -v` | Verbose output |
| `--quiet, -q` | Suppress output |
| `--format <type>` | Output format (json, yaml, table) |
| `--workspace <path>` | Specify workspace directory |
| `--config <file>` | Use specific config file |

## Stage Commands

### zen align - Strategic Alignment

Define what success looks like and why the work matters.

#### Basic Commands

```bash
# Initialize new initiative
zen align init <name>
zen align init "Payment System v2"

# Show current alignment
zen align status
zen align show
```

#### Metrics Management

```bash
# Add success metrics
zen align metrics --add <metric> --target <value>
zen align metrics --add "Conversion rate" --target "15% increase"
zen align metrics --add "Response time" --target "<200ms"

# List metrics
zen align metrics --list

# Update metric
zen align metrics --update "Conversion rate" --target "20% increase"

# Remove metric
zen align metrics --remove "Response time"
```

#### Stakeholder Management

```bash
# Add stakeholders
zen align stakeholders --add <name> --role <role>
zen align stakeholders --add "Jane Smith" --role "Product Owner"
zen align stakeholders --add "@security-team" --role "Reviewer"

# Request approval
zen align approve --request --from "Jane Smith"

# Record approval
zen align approve --record --from "Jane Smith" --status "approved"

# List stakeholders
zen align stakeholders --list
```

#### Constraints and Resources

```bash
# Set timeline
zen align constraints --timeline <duration>
zen align constraints --timeline "6 weeks"
zen align constraints --timeline "2024-Q2"

# Set budget
zen align constraints --budget <amount>
zen align constraints --budget "$250,000"

# Allocate team
zen align constraints --team <resources>
zen align constraints --team "3 engineers, 1 designer, 1 PM"
```

#### Risk Management

```bash
# Add risk
zen align risk --add <description> --severity <level> --mitigation <plan>
zen align risk --add "Third-party API dependency" \
  --severity "high" \
  --mitigation "Build fallback mechanism"

# List risks
zen align risk --list

# Update risk
zen align risk --update <id> --severity "medium"
```

#### Templates

```bash
# Use template
zen align --template <template-name>
zen align --template "feature-development"
zen align --template "bug-fix"
zen align --template "infrastructure"

# List templates
zen align templates --list
```

### zen discover - Research and Discovery

Gather evidence and insights to inform decisions.

#### Basic Commands

```bash
# Start discovery
zen discover start
zen discover init

# Show discovery status
zen discover status
zen discover summary
```

#### Research Management

```bash
# Add research finding
zen discover research --add --method <type> --finding <insight>
zen discover research --add \
  --method "user-interview" \
  --finding "Users want faster checkout"

# Research methods
zen discover research --add --method "user-interview"
zen discover research --add --method "survey"
zen discover research --add --method "data-analysis"
zen discover research --add --method "competitive-analysis"
zen discover research --add --method "usability-testing"

# List findings
zen discover research --list

# Tag findings
zen discover research --tag <id> --labels "critical,ux"
```

#### Constraint Documentation

```bash
# Add constraint
zen discover constraint --add --type <type> --description <desc>
zen discover constraint --add \
  --type "technical" \
  --description "Must support IE11"

# Constraint types
zen discover constraint --add --type "technical"
zen discover constraint --add --type "business"
zen discover constraint --add --type "legal"
zen discover constraint --add --type "security"

# List constraints
zen discover constraint --list
```

#### Assumption Tracking

```bash
# Log assumption
zen discover assumption --add <assumption> --test <method>
zen discover assumption --add \
  "Users will adopt new feature quickly" \
  --test "A/B test with 5% of users"

# Validate assumption
zen discover assumption --validate <id> --result <outcome>
zen discover assumption --validate 3 --result "confirmed"

# List assumptions
zen discover assumption --list --status "unvalidated"
```

#### Data Collection

```bash
# Import user feedback
zen discover import --source "support-tickets" --file tickets.csv
zen discover import --source "app-reviews" --file reviews.json

# Scrape market data
zen discover scrape --source "competitor-sites"
zen discover scrape --source "industry-reports"

# Analyze existing data
zen discover analyze --source "analytics" --period "30d"
```

### zen prioritize - Ranking and Planning

Rank work by value and allocate resources.

#### Basic Commands

```bash
# List items to prioritize
zen prioritize list
zen prioritize items

# Show current priorities
zen prioritize status
zen prioritize show --ranked
```

#### Scoring Methods

```bash
# Apply WSJF scoring
zen prioritize score --method wsjf
zen prioritize score --method wsjf \
  --business-value 8 \
  --time-criticality 6 \
  --risk-reduction 4 \
  --job-size 5

# Apply ICE scoring
zen prioritize score --method ice
zen prioritize score --method ice \
  --impact 8 \
  --confidence 7 \
  --ease 5

# Apply RICE scoring
zen prioritize score --method rice
zen prioritize score --method rice \
  --reach 10000 \
  --impact 3 \
  --confidence 80 \
  --effort 4

# Batch scoring
zen prioritize score --batch --file scores.csv
```

#### Capacity Planning

```bash
# Check capacity
zen prioritize capacity --check
zen prioritize capacity --check --team "mobile"

# Set capacity
zen prioritize capacity --set --sprint-weeks 2 --velocity 40
zen prioritize capacity --set --engineers 5 --days 10

# Capacity forecast
zen prioritize capacity --forecast --items 1,2,3,5,8
```

#### Dependency Management

```bash
# Map dependencies
zen prioritize dependencies --map
zen prioritize dependencies --add "Feature A" --requires "Feature B"

# Check dependency conflicts
zen prioritize dependencies --check

# Generate dependency graph
zen prioritize dependencies --visualize --output deps.png
```

#### Release Planning

```bash
# Create release
zen prioritize release --create <name> --target <date>
zen prioritize release --create "v2.0" --target "2024-06-01"

# Assign items to release
zen prioritize release --assign <items> --to <release>
zen prioritize release --assign 1,3,5 --to "v2.0"

# Show release plan
zen prioritize release --show "v2.0"

# Validate release feasibility
zen prioritize release --validate "v2.0"
```

### zen design - Specification and Architecture

Define what will be built and how.

#### Basic Commands

```bash
# Initialize design phase
zen design init
zen design start

# Show design status
zen design status
zen design progress
```

#### Contract Management

```bash
# Create API contract
zen design contract --create <name>
zen design contract --create "payment-api"

# Add endpoint
zen design contract --endpoint <method> <path>
zen design contract --endpoint "POST /payments/charge"
zen design contract --endpoint "GET /payments/:id"

# Define schemas
zen design contract --schema request.json --for "POST /payments/charge"
zen design contract --schema response.json --for "POST /payments/charge"

# Validate contract
zen design contract --validate
zen design contract --validate --strict

# Generate from OpenAPI
zen design contract --import openapi.yaml
```

#### UX Design

```bash
# Create wireframe
zen design ux --wireframe <name>
zen design ux --wireframe "checkout-flow"

# Create prototype
zen design ux --prototype <name>
zen design ux --prototype "payment-selection"

# User testing
zen design ux --test --users 5
zen design ux --test --method "think-aloud"

# Export designs
zen design ux --export --format "figma"
zen design ux --export --format "pdf" --output designs.pdf
```

#### Architecture Decisions

```bash
# Create ADR (Architecture Decision Record)
zen design adr --create <title>
zen design adr --create "Use microservices architecture"

# Add context and decision
zen design adr --context "Need to scale independently"
zen design adr --decision "Split into 5 microservices"
zen design adr --consequences "Increased operational complexity"

# List ADRs
zen design adr --list

# Review ADR
zen design adr --review <id> --status "approved"
```

#### Data Modeling

```bash
# Define schema
zen design schema --table <name> --fields <fields>
zen design schema --table "users" \
  --fields "id:uuid,email:string,created_at:timestamp"

# Add relationships
zen design schema --relationship "users" --has-many "orders"

# Generate migrations
zen design migration --generate
zen design migration --plan

# Validate schema
zen design schema --validate
```

### zen build - Implementation

Transform designs into working code.

#### Basic Commands

```bash
# Start building
zen build start
zen build init --feature <name>

# Show build status
zen build status
zen build progress
```

#### Code Generation

```bash
# Generate scaffolding
zen build scaffold --from-contract <contract>
zen build scaffold --from-contract "payment-api"

# Generate specific components
zen build generate --controller "PaymentController"
zen build generate --model "Payment"
zen build generate --test "PaymentTest"

# Generate from templates
zen build generate --template "rest-api"
```

#### Testing

```bash
# Run tests
zen build test
zen build test --unit
zen build test --integration
zen build test --e2e

# Run specific test
zen build test --file "payment_test.go"
zen build test --pattern "*Payment*"

# Coverage report
zen build test --coverage
zen build test --coverage --min 80
```

#### Code Quality

```bash
# Run linting
zen build lint
zen build lint --fix

# Security scan
zen build security-scan
zen build security-scan --severity "high,critical"

# Complexity analysis
zen build analyze --complexity
zen build analyze --duplication
```

#### Code Review

```bash
# Create pull request
zen build review --create --title <title>
zen build review --create \
  --title "Add payment processing" \
  --description "Implements Stripe integration" \
  --reviewers "@alice,@bob"

# Request review
zen build review --request --from "@senior-dev"

# Submit feedback
zen build review --comment "LGTM with minor suggestions"
zen build review --approve
zen build review --request-changes

# Check review status
zen build review --status
```

#### Validation

```bash
# Validate all quality gates
zen build validate
zen build validate --strict

# Validate specific aspects
zen build validate --tests
zen build validate --coverage
zen build validate --security
zen build validate --docs
```

### zen ship - Deployment

Deploy safely to production.

#### Basic Commands

```bash
# Start shipping process
zen ship start
zen ship init

# Show shipping status
zen ship status
zen ship checklist
```

#### Pre-deployment Validation

```bash
# Comprehensive validation
zen ship validate --comprehensive
zen ship validate --stage "production"

# Specific validations
zen ship validate --security
zen ship validate --performance
zen ship validate --accessibility
zen ship validate --compatibility
```

#### Security Assessment

```bash
# Run security scans
zen ship security --scan
zen ship security --scan --depth "full"
zen ship security --penetration-test

# Review vulnerabilities
zen ship security --report
zen ship security --accept-risk <vuln-id> --reason <text>
```

#### Performance Testing

```bash
# Load testing
zen ship performance --load-test --users 1000
zen ship performance --load-test --rps 500

# Stress testing
zen ship performance --stress-test
zen ship performance --stress-test --duration "30m"

# Performance report
zen ship performance --report
zen ship performance --compare --baseline "v1.0"
```

#### Deployment

```bash
# Deploy to environment
zen ship deploy --environment <env>
zen ship deploy --environment "staging"
zen ship deploy --environment "production"

# Canary deployment
zen ship deploy --canary --percentage 5
zen ship canary --expand --percentage 25
zen ship canary --expand --percentage 50
zen ship canary --promote

# Blue-green deployment
zen ship deploy --blue-green
zen ship deploy --switch-to "green"

# Feature flag deployment
zen ship deploy --feature-flag "new-checkout"
zen ship feature-flag --enable --percentage 10
zen ship feature-flag --enable --users "beta-testers"
```

#### Monitoring

```bash
# Monitor deployment
zen ship monitor
zen ship monitor --duration "2h"
zen ship monitor --metrics "error-rate,latency,throughput"

# Set up alerts
zen ship monitor --alert --threshold "error-rate>0.1%"
zen ship monitor --alert --oncall "@ops-team"

# Health checks
zen ship health
zen ship health --continuous
```

#### Rollback

```bash
# Rollback deployment
zen ship rollback
zen ship rollback --reason "High error rate detected"
zen ship rollback --to-version "v1.9.8"

# Emergency rollback
zen ship rollback --emergency
zen ship rollback --emergency --skip-checks
```

### zen learn - Outcome Analysis

Measure results and plan improvements.

#### Basic Commands

```bash
# Start learning phase
zen learn start
zen learn init

# Show learning status
zen learn status
zen learn summary
```

#### Metrics Collection

```bash
# Collect metrics
zen learn metrics --collect
zen learn metrics --collect --period "7d"
zen learn metrics --collect --from "2024-01-01" --to "2024-01-31"

# Specific metrics
zen learn metrics --business
zen learn metrics --technical
zen learn metrics --user-experience
```

#### Analysis

```bash
# Analyze outcomes
zen learn analyze
zen learn analyze --compare-baseline
zen learn analyze --statistical-significance

# Deep dive into specific metric
zen learn analyze --metric "conversion-rate"
zen learn analyze --metric "page-load-time" --percentile "p95"

# Correlation analysis
zen learn correlate --metrics "feature-usage,retention"
```

#### Feedback Management

```bash
# Collect feedback
zen learn feedback --collect
zen learn feedback --collect --source "support"
zen learn feedback --collect --source "reviews"
zen learn feedback --collect --source "surveys"

# Synthesize feedback
zen learn feedback --synthesize
zen learn feedback --themes
zen learn feedback --sentiment

# Export feedback
zen learn feedback --export --format "csv"
```

#### Documentation

```bash
# Document learnings
zen learn document --insights
zen learn document --lesson "Clear CTAs increase conversion"
zen learn document --technical "Caching reduces load by 40%"

# Generate reports
zen learn report
zen learn report --format "pdf" --output "outcomes.pdf"
zen learn report --executive-summary
```

#### Iteration Planning

```bash
# Generate recommendations
zen learn recommend
zen learn recommend --priority "high"

# Create follow-up items
zen learn iterate --create-items
zen learn iterate --backlog

# Plan next cycle
zen learn next --initiatives
zen learn next --experiments
```

## Workflow Commands

### zen workflow - Workflow Management

Manage multiple concurrent workflows.

```bash
# List workflows
zen workflow list
zen workflow list --active
zen workflow list --completed

# Switch workflow
zen workflow switch <name>
zen workflow switch "Payment System"

# Create workflow
zen workflow create <name>
zen workflow create "Search Feature"

# Archive workflow
zen workflow archive <name>
zen workflow archive "Legacy Migration"

# Workflow status
zen workflow status
zen workflow status --detailed
```

### zen status - Overall Status

Check current position in workflow.

```bash
# Show current status
zen status
zen status --detailed
zen status --stage "build"

# Progress visualization
zen status --progress
zen status --timeline
zen status --burndown
```

### zen validate - Validation

Run quality gate validations.

```bash
# Validate current stage
zen validate
zen validate --stage <stage>
zen validate --stage "design"

# Validate all stages
zen validate --all
zen validate --all --stop-on-failure

# Specific validations
zen validate --quality-gates
zen validate --dependencies
zen validate --resources
```

## Utility Commands

### zen config - Configuration

Manage Zen configuration.

```bash
# Show configuration
zen config show
zen config show --section "gates"

# Set configuration
zen config set <key> <value>
zen config set gates.build.coverage 85
zen config set team.size 5

# Integration configuration
zen config integrations
zen config integrations --add "jira"
zen config integrations --test "github"
```

### zen templates - Template Management

Work with templates.

```bash
# List templates
zen templates list
zen templates list --category "feature"

# Show template
zen templates show <name>
zen templates show "microservice"

# Create from template
zen templates use <template> --name <project>
zen templates use "rest-api" --name "user-service"

# Create custom template
zen templates create <name> --from-current
```

### zen report - Reporting

Generate various reports.

```bash
# Generate report
zen report
zen report --type "workflow"
zen report --type "quality"
zen report --type "velocity"

# Report options
zen report --format "html"
zen report --period "30d"
zen report --team "mobile"
zen report --output "report.pdf"
```

### zen help - Help System

Get help on commands.

```bash
# General help
zen help
zen --help

# Command help
zen help <command>
zen help align
zen help build test

# Show examples
zen help --examples
zen help align --examples
```

## Command Shortcuts

Common command aliases for efficiency:

```bash
# Stage shortcuts
zen a  → zen align
zen d  → zen discover
zen p  → zen prioritize
zen de → zen design
zen b  → zen build
zen s  → zen ship
zen l  → zen learn

# Common operations
zen start      → Start current stage
zen next       → Progress to next stage
zen check      → Validate current stage
zen undo       → Rollback last action
```

## Environment Variables

Configure Zen behavior with environment variables:

```bash
# Workspace configuration
export ZEN_WORKSPACE="/path/to/workspace"
export ZEN_CONFIG="/path/to/config.yaml"

# Integration tokens
export ZEN_GITHUB_TOKEN="ghp_xxxxx"
export ZEN_JIRA_TOKEN="xxxxx"
export ZEN_FIGMA_TOKEN="xxxxx"

# Behavior modification
export ZEN_AUTO_VALIDATE="true"
export ZEN_STRICT_MODE="true"
export ZEN_VERBOSE="true"
```

## Exit Codes

Understanding command exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Validation failure |
| 3 | Quality gate failure |
| 4 | Missing prerequisites |
| 5 | Permission denied |
| 6 | Resource not found |
| 7 | Timeout |
| 8 | Conflict |

## Interactive Mode

Run Zen in interactive mode:

```bash
# Start interactive session
zen interactive
zen i

# Interactive commands
> align init "New Feature"
> discover start
> prioritize score --method ice
> exit
```

## Scripting

Use Zen in scripts:

```bash
#!/bin/bash
# Example: Automated validation script

zen validate --stage build || exit 1
zen test --coverage --min 80 || exit 1
zen security-scan --fail-on "high,critical" || exit 1

echo "All validations passed!"
```

## Command Completion

Enable shell completion:

```bash
# Bash
zen completion bash > /etc/bash_completion.d/zen

# Zsh
zen completion zsh > "${fpath[1]}/_zen"

# Fish
zen completion fish > ~/.config/fish/completions/zen.fish
```

## Best Practices

### Command Usage Tips

1. **Use verbose mode for debugging**: `zen --verbose <command>`
2. **Pipe to other tools**: `zen status --format json | jq`
3. **Combine commands**: `zen validate && zen ship deploy`
4. **Use templates for consistency**: `zen align --template feature`
5. **Check before progressing**: `zen validate --stage <next-stage>`

### Common Patterns

```bash
# Full cycle automation
zen align init "Feature" && \
zen discover start && \
zen prioritize score --method ice && \
zen design contract --create && \
zen build scaffold && \
zen ship deploy --canary && \
zen learn analyze

# Validation before progression
zen validate || { echo "Fix issues before continuing"; exit 1; }

# Conditional deployment
if zen ship validate --comprehensive; then
  zen ship deploy --production
else
  zen ship deploy --staging
fi
```

## Summary

The Zen CLI provides comprehensive commands for every stage of the Zenflow workflow. Commands are designed to be:

- **Intuitive** - Natural language-like syntax
- **Composable** - Combine for complex operations
- **Scriptable** - Automate workflows
- **Integrated** - Work with existing tools

Master these commands to streamline your product development workflow.
