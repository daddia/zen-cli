# Getting Started with Zenflow

This guide walks you through your first complete Zenflow cycle, from initial strategy alignment to production deployment and outcome measurement.

## Prerequisites

### Required Setup

Before starting, ensure you have:

1. **Zen CLI installed and configured**
   ```bash
   zen version
   # Should output: zen version 0.1.0 or higher
   ```

2. **Workspace initialized**
   ```bash
   zen init
   # Creates .zen directory with configuration
   ```

3. **Tool integrations configured** (optional but recommended)
   ```bash
   zen config integrations
   # Connect to Jira, GitHub, Figma, etc.
   ```

### Team Preparation

- Identify key stakeholders for your initiative
- Gather any existing research or documentation
- Allocate time for the full workflow cycle (typically 2-4 weeks)

## Tutorial: Your First Zenflow Cycle

We'll build a simple user authentication feature to demonstrate the complete workflow.

### Stage 1: Align - Define Success

Start by establishing what you're building and why it matters.

```bash
# Initialize a new initiative
zen align init "User Authentication"

# Define success metrics
zen align metrics --add "User registration rate" --target "50% increase"
zen align metrics --add "Login success rate" --target "95%"

# Set timeline and resources
zen align constraints --timeline "4 weeks" --team "2 engineers, 1 designer"
```

**What happens:**
- Creates a PR/FAQ document template
- Establishes measurable success criteria
- Documents resource allocation

**Key deliverable:** Approved strategy document with clear success metrics

### Stage 2: Discover - Gather Insights

Research user needs and technical constraints.

```bash
# Start discovery process
zen discover start

# Capture user research findings
zen discover research --source "user-interviews" --finding "Users want social login options"

# Document technical constraints
zen discover technical --constraint "Must integrate with existing user database"

# Log risks and assumptions
zen discover risk --add "Third-party OAuth provider dependency" --severity "medium"
```

**What happens:**
- Aggregates research from multiple sources
- Documents constraints and dependencies
- Creates risk register with mitigation plans

**Key deliverable:** Discovery brief with validated requirements

### Stage 3: Prioritize - Rank Features

Determine what to build first based on value and effort.

```bash
# List discovered features
zen prioritize list
# Output: 
# 1. Email/password authentication
# 2. Social login (Google, GitHub)
# 3. Two-factor authentication
# 4. Password reset flow

# Apply prioritization framework
zen prioritize rank --method ice
# Prompts for Impact, Confidence, Ease scores

# Check capacity
zen prioritize capacity --check
# Validates against team availability

# Assign to release
zen prioritize release --assign "v1.0" --items 1,2,4
```

**What happens:**
- Applies systematic prioritization scoring
- Validates delivery capacity
- Creates release plan

**Key deliverable:** Ranked backlog with release assignments

### Stage 4: Design - Specify Solution

Create detailed specifications for what you'll build.

```bash
# Generate API contract
zen design contract --init "auth-api"

# Define endpoints
zen design contract --add-endpoint "POST /auth/register"
zen design contract --add-endpoint "POST /auth/login"
zen design contract --add-endpoint "POST /auth/reset-password"

# Create UX flows
zen design ux --wireframe "registration-flow"
zen design ux --wireframe "login-flow"

# Plan data migrations
zen design migration --plan "add-oauth-providers-table"
```

**What happens:**
- Creates OpenAPI specification
- Generates wireframes and user flows
- Documents architecture decisions
- Plans database changes

**Key deliverable:** Complete technical and UX specifications

### Stage 5: Build - Implement Solution

Generate code scaffolding and implement features.

```bash
# Generate implementation from contracts
zen build scaffold --from-contract "auth-api"

# Start development
zen build start --feature "user-registration"

# Run tests locally
zen build test --local
# All tests pass ✓

# Submit for review
zen build review --submit

# Check quality gates
zen build validate
# ✓ Test coverage: 85%
# ✓ Linting: Pass
# ✓ Security scan: No issues
```

**What happens:**
- Generates boilerplate code from specifications
- Enforces code quality standards
- Runs automated testing
- Manages code review process

**Key deliverable:** Feature-complete, tested code

### Stage 6: Ship - Deploy Safely

Deploy to production with confidence.

```bash
# Run comprehensive validation
zen ship validate --comprehensive
# ✓ All tests pass
# ✓ Security scan clean
# ✓ Performance within budget

# Deploy canary release
zen ship canary --percentage 5
# Deploying to 5% of users...

# Monitor canary metrics
zen ship monitor --duration "2 hours"
# ✓ No errors detected
# ✓ Performance stable
# ✓ User metrics positive

# Promote to full production
zen ship promote --to-production
# Rolling out to 100% of users...

# Verify deployment
zen ship verify
# ✓ All systems operational
```

**What happens:**
- Runs all quality gates
- Deploys incrementally with monitoring
- Validates production health
- Enables quick rollback if needed

**Key deliverable:** Successfully deployed features in production

### Stage 7: Learn - Measure Outcomes

Analyze results and plan improvements.

```bash
# Collect metrics
zen learn metrics --collect --period "1 week"

# Analyze outcomes
zen learn analyze
# User registration rate: +67% (exceeded target!)
# Login success rate: 93% (slightly below target)

# Gather feedback
zen learn feedback --synthesize
# Key insight: Users confused by password requirements

# Document learnings
zen learn document --lesson "Clearer password validation improves success rate"

# Plan next iteration
zen learn iterate --recommend
# Suggested: Improve password validation UX
# Suggested: Add password strength meter
```

**What happens:**
- Measures success metrics against targets
- Synthesizes user feedback
- Documents lessons learned
- Recommends next actions

**Key deliverable:** Outcome report with actionable insights

## Understanding Workflow Progression

### Sequential Flow

Each stage builds on the previous one:

```
Align → Discover → Prioritize → Design → Build → Ship → Learn
  ↑                                                           ↓
  └───────────────── Continuous Improvement ─────────────────┘
```

### Quality Gates

You cannot skip stages - quality gates enforce completion:

```bash
# Attempting to build without design
zen build start
# Error: Design stage not complete. Run 'zen design' first.

# Check current stage
zen status
# Current stage: Design (75% complete)
# Next: Complete UX wireframes
```

### Parallel Work

While stages are sequential, teams can work on multiple initiatives:

```bash
# List active workflows
zen workflow list
# 1. User Authentication [Ship stage]
# 2. Shopping Cart [Design stage]
# 3. Search Feature [Discover stage]

# Switch context
zen workflow switch "Shopping Cart"
```

## Common Workflows by Role

### Product Manager Workflow

Focus on stages 1-3 and 7:

```bash
zen align            # Define strategy and success metrics
zen discover         # Conduct user research
zen prioritize       # Rank features by value
zen learn            # Analyze outcomes
```

### Designer Workflow

Focus on stages 2-4:

```bash
zen discover --ux    # User research and testing
zen design --ux      # Create wireframes and prototypes
zen build --review   # Validate implementation
```

### Engineer Workflow

Focus on stages 4-6:

```bash
zen design --technical  # Create API contracts
zen build              # Implement features
zen ship               # Deploy to production
```

### Analytics Workflow

Involved in stages 1, 6, and 7:

```bash
zen align --metrics     # Define success metrics
zen ship --monitor      # Track deployment metrics
zen learn --analyze     # Measure outcomes
```

## Tips for Success

### Start Small

Begin with a low-risk feature to learn the workflow:
- Choose something with clear success metrics
- Limit scope to 1-2 week cycles initially
- Focus on completing all stages rather than perfection

### Use Templates

Leverage built-in templates to accelerate work:

```bash
# List available templates
zen templates list

# Use a template
zen align --template "feature-development"
```

### Automate Repetitive Tasks

Create aliases for common command combinations:

```bash
# Add to your shell configuration
alias zen-start='zen align && zen discover'
alias zen-develop='zen build && zen test --local'
alias zen-deploy='zen ship validate && zen ship canary'
```

### Monitor Progress

Track your workflow progress regularly:

```bash
# Check overall status
zen status --detailed

# View stage completion
zen progress

# Generate workflow report
zen report --format markdown > workflow-report.md
```

## Next Steps

Now that you've completed your first Zenflow cycle:

1. **Deep dive into stages** - Read the [Stages Guide](stages.md) for detailed documentation
2. **Learn the commands** - Explore the [Command Reference](commands.md)
3. **Understand quality gates** - Review [Quality Gates](quality-gates.md) documentation
4. **Explore streams** - Learn about [Workflow Streams](streams.md) for specialized implementations
5. **Apply best practices** - Read [Best Practices](best-practices.md) from successful teams

## Getting Help

If you encounter issues:

1. Check the [Troubleshooting Guide](troubleshooting.md)
2. Run `zen help [command]` for command-specific help
3. Visit the [Community Forum](https://community.zen.dev)
4. Report bugs via `zen feedback --bug`

Remember: Zenflow is designed to help you ship better products faster. Start with the basics, then customize as you learn what works for your team.
