# Zenflow Best Practices

This guide provides proven patterns and recommendations from successful Zenflow implementations. Follow these practices to maximize value and avoid common pitfalls.

## General Best Practices

### Start Small, Scale Gradually

#### Begin with Low-Risk Projects

Start your Zenflow journey with smaller, less critical projects:

```bash
# Good first projects
- Internal tool improvement
- Bug fix with clear scope  
- Small feature enhancement
- Technical debt reduction

# Avoid for first project
- Critical customer-facing feature
- Major architecture change
- Compliance-related work
- Multi-team dependencies
```

#### Progressive Adoption

```bash
# Week 1-2: Learn the basics
zen workflow create --tutorial "My First Workflow"
zen align init "Simple Feature"

# Week 3-4: Add quality gates
zen config gates --enable "basic"

# Week 5-6: Integrate tools
zen config integrations --add "jira,github"

# Week 7-8: Full workflow
zen config gates --enable "standard"
zen workflow create --production "Real Feature"
```

### Complete Stages Fully

#### Avoid Stage Skipping

Each stage builds critical context for the next:

```bash
# ❌ Bad: Skipping discovery
zen align init "New Feature"
zen design start  # Error: Discovery not complete

# ✅ Good: Complete each stage
zen align init "New Feature"
zen align approve --get
zen discover start
zen discover research --complete
zen design start
```

#### Definition of Done

Establish clear completion criteria:

```bash
# Configure stage completion requirements
zen config stages.align.done-criteria \
  --requires "metrics-defined,stakeholders-aligned,resources-allocated"

zen config stages.discover.done-criteria \
  --requires "research-complete,risks-documented,assumptions-logged"
```

### Measure Everything

#### Instrument Comprehensively

```bash
# Business metrics
zen metrics add --business \
  "conversion-rate,revenue-per-user,churn-rate"

# Technical metrics  
zen metrics add --technical \
  "latency-p99,error-rate,throughput"

# Process metrics
zen metrics add --process \
  "cycle-time,defect-rate,velocity"
```

#### Review Metrics Regularly

```bash
# Weekly metric review
zen metrics review --period "1w" --compare-baseline

# Monthly trend analysis
zen metrics trends --period "30d" --identify-patterns

# Quarterly outcomes assessment
zen metrics outcomes --quarter "Q2" --vs-objectives
```

## Stage-Specific Best Practices

### Align Stage

#### Write Clear Problem Statements

```bash
# ❌ Vague problem
"Users are unhappy with the app"

# ✅ Specific problem
"60% of users abandon checkout when shipping costs appear, 
resulting in $2M monthly lost revenue"
```

#### Define Measurable Success

```bash
# ❌ Unmeasurable success criteria
zen align metrics --add "Better user experience"

# ✅ Measurable success criteria
zen align metrics --add "Checkout completion rate" --target "75%" --by "Q2"
zen align metrics --add "Time to checkout" --target "<90 seconds"
```

#### Get Explicit Buy-in

```bash
# Document stakeholder approval
zen align stakeholders --add "Product VP" --role "Sponsor"
zen align stakeholders --add "Engineering Director" --role "Approver"
zen align approve --request --from "all"
zen align approve --record --evidence "meeting-notes.md"
```

### Discover Stage

#### Research Broadly

Gather evidence from multiple sources:

```bash
# User perspective
zen discover research --method "user-interviews" --count 15
zen discover research --method "surveys" --responses 500
zen discover research --method "analytics" --cohort "last-30-days"

# Technical perspective
zen discover technical --assessment "current-architecture"
zen discover technical --constraints "platform-limitations"

# Market perspective
zen discover market --competitors "top-5"
zen discover market --trends "industry-report-2024"
```

#### Document Assumptions

Make implicit assumptions explicit:

```bash
# Log all assumptions
zen discover assumption --add \
  "Users will trust social login" \
  --confidence "medium" \
  --test "A/B test with 5% traffic"

zen discover assumption --add \
  "API can handle 2x traffic" \
  --confidence "high" \
  --test "Load test before launch"
```

#### Identify Risks Early

```bash
# Categorize risks
zen discover risk --add "Regulatory" \
  --description "GDPR compliance for EU users" \
  --impact "high" \
  --mitigation "Legal review and data audit"

zen discover risk --add "Technical" \
  --description "Legacy system integration" \
  --impact "medium" \
  --mitigation "Build adapter layer"
```

### Prioritize Stage

#### Use Consistent Scoring

Apply the same method across all items:

```bash
# Configure default method
zen config prioritize.method --default "wsjf"

# Score all items
zen prioritize score --batch --consistent
```

#### Consider Hidden Costs

Include all effort, not just development:

```bash
# ❌ Incomplete effort estimate
zen prioritize estimate "Feature X" --dev-hours 40

# ✅ Complete effort estimate
zen prioritize estimate "Feature X" \
  --dev-hours 40 \
  --qa-hours 16 \
  --design-hours 20 \
  --deployment-hours 8 \
  --documentation-hours 4
```

#### Map Dependencies

Identify and sequence dependencies:

```bash
# Map technical dependencies
zen prioritize dependencies --map
# Feature A → requires → Authentication System
# Feature B → requires → Feature A
# Feature C → independent

# Optimize sequence
zen prioritize sequence --optimize
# Recommended order: Auth System → A → B, C (parallel)
```

### Design Stage

#### Contracts First

Define interfaces before implementation:

```bash
# Start with API contract
zen design contract --create "user-service"
zen design contract --endpoint "POST /users"
zen design contract --validate

# Then design internals
zen design architecture --components
zen design database --schema
```

#### Design for Testability

```bash
# Include test scenarios in design
zen design test-scenarios --add \
  "Happy path: successful user creation"
  
zen design test-scenarios --add \
  "Error case: duplicate email"
  
zen design test-scenarios --add \
  "Edge case: maximum field lengths"
```

#### Document Decisions

```bash
# Create ADRs for key decisions
zen design adr --create \
  "Use event-driven architecture" \
  --context "Need to decouple services" \
  --decision "Implement event bus with Kafka" \
  --consequences "Added operational complexity"
```

### Build Stage

#### Small, Focused Changes

```bash
# ❌ Large pull request
zen build pr --stats
# Files changed: 47
# Lines added: 2,847
# Review time estimate: 4 hours

# ✅ Small pull requests
zen build pr --stats
# Files changed: 5
# Lines added: 127
# Review time estimate: 20 minutes
```

#### Test-Driven Development

```bash
# Write tests first
zen build test --write "UserService.createUser"

# Then implement
zen build implement "UserService.createUser"

# Verify coverage
zen build coverage --check
# Coverage: 92% ✓
```

#### Continuous Integration

```bash
# Run tests on every commit
zen build ci --configure \
  --on "push,pull-request" \
  --tests "unit,integration" \
  --quality "lint,security"
```

### Ship Stage

#### Progressive Deployment

Never go from 0 to 100%:

```bash
# Gradual rollout
zen ship deploy --canary --percentage 1
# Monitor for 1 hour
zen ship deploy --canary --percentage 5
# Monitor for 4 hours
zen ship deploy --canary --percentage 25
# Monitor for 1 day
zen ship deploy --canary --percentage 50
# Monitor for 1 day
zen ship deploy --production
```

#### Monitor Everything

```bash
# Application metrics
zen ship monitor --metrics \
  "error-rate,latency,throughput"

# Business metrics
zen ship monitor --business \
  "conversion,revenue,user-activity"

# Infrastructure metrics
zen ship monitor --infrastructure \
  "cpu,memory,disk,network"
```

#### Prepare for Rollback

```bash
# Test rollback procedure
zen ship rollback --test --dry-run

# Document rollback triggers
zen ship rollback --triggers \
  --error-rate ">1%" \
  --latency-p99 ">500ms" \
  --revenue-drop ">5%"
```

### Learn Stage

#### Analyze Objectively

Base conclusions on data:

```bash
# ❌ Subjective assessment
"The feature seems to be working well"

# ✅ Data-driven assessment
zen learn analyze
# Conversion rate: +12% (p<0.01, significant)
# User complaints: -8%
# Performance impact: +20ms latency (acceptable)
```

#### Share Learnings

```bash
# Document insights
zen learn document --insight \
  "Mobile users convert 3x when Apple Pay is offered"

# Share with organization
zen learn share --channels "slack,wiki,email"
zen learn present --to "all-hands-meeting"
```

#### Close the Loop

```bash
# Create follow-up actions
zen learn actions --create \
  "Optimize mobile checkout based on learnings"

# Update next cycle priorities
zen prioritize update --based-on "learnings"
```

## Team Best Practices

### Cross-Functional Collaboration

#### Regular Sync Points

```bash
# Configure team sync points
zen team sync --schedule \
  --daily "standup,15min" \
  --weekly "planning,1hr" \
  --biweekly "retrospective,1hr"
```

#### Clear Ownership

```bash
# Assign stage owners
zen team assign --stage "align" --owner "@product-manager"
zen team assign --stage "discover" --owner "@ux-researcher"
zen team assign --stage "build" --owner "@tech-lead"
zen team assign --stage "ship" --owner "@sre-lead"
```

#### Shared Context

```bash
# Maintain shared documentation
zen docs update --after-each-stage
zen docs publish --to "confluence"

# Record decisions
zen decisions record --in "decision-log"
zen decisions share --with "team"
```

### Remote Team Practices

#### Async-First Communication

```bash
# Document everything
zen communicate --prefer "written"
zen meetings record --always
zen decisions document --immediately

# Use async reviews
zen review --async --deadline "48hrs"
```

#### Time Zone Awareness

```bash
# Configure team time zones
zen team timezone --set \
  --member "alice" --tz "PST" \
  --member "bob" --tz "EST" \
  --member "carol" --tz "GMT"

# Schedule considerately
zen schedule --respect-timezones
zen schedule --core-hours "9am-12pm PST"
```

## Tool Integration Best Practices

### Version Control Integration

```bash
# Link commits to workflow
git commit -m "feat: implement payment processing [zen:build:payment-feature]"

# Automate with hooks
# .git/hooks/prepare-commit-msg
#!/bin/bash
ZEN_CONTEXT=$(zen context current)
echo "[zen:$ZEN_CONTEXT]" >> $1
```

### Project Management Integration

```bash
# Sync with Jira
zen integrate jira --sync-bidirectional
zen integrate jira --map \
  --zen-stage "align" --jira-status "Discovery" \
  --zen-stage "build" --jira-status "In Progress" \
  --zen-stage "ship" --jira-status "Testing"
```

### CI/CD Integration

```bash
# GitHub Actions example
- name: Zenflow Validation
  run: |
    zen validate --stage ${{ env.ZEN_STAGE }}
    zen quality-gates --check
    zen metrics --capture
```

## Common Patterns

### Feature Development Pattern

Standard pattern for new features:

```bash
# 1. Strategic alignment (1-2 days)
zen align init "New Feature"
zen align metrics --define
zen align approve --get

# 2. Discovery sprint (1 week)
zen discover research --user
zen discover technical --assessment
zen discover risks --identify

# 3. Design phase (3-4 days)
zen prioritize score --items
zen design contracts --create
zen design ux --wireframes

# 4. Implementation sprints (2-3 weeks)
zen build scaffold --generate
zen build implement --tdd
zen build review --continuous

# 5. Deployment (2-3 days)
zen ship validate --comprehensive
zen ship deploy --progressive
zen ship monitor --continuous

# 6. Learning cycle (ongoing)
zen learn metrics --collect
zen learn analyze --weekly
zen learn iterate --plan
```

### Hotfix Pattern

Expedited pattern for critical fixes:

```bash
# 1. Emergency alignment (30 min)
zen align init --emergency "Critical Bug Fix"
zen align approve --emergency-protocol

# 2. Fast validation (2 hrs)
zen build fix --implement
zen build test --regression
zen build review --expedited

# 3. Rapid deployment (1 hr)
zen ship validate --critical-only
zen ship deploy --hotfix
zen ship monitor --heightened

# 4. Post-mortem (next day)
zen learn incident --analyze
zen learn prevent --plan
```

### Experiment Pattern

Pattern for A/B tests and experiments:

```bash
# 1. Hypothesis formation
zen align hypothesis "New checkout increases conversion"
zen align metrics --experiment "conversion-rate"

# 2. Experiment design
zen discover sample-size --calculate
zen design experiment --treatment "new-checkout"
zen design experiment --control "existing-checkout"

# 3. Implementation
zen build feature-flag "experiment-checkout"
zen ship deploy --experiment --split 50/50

# 4. Analysis
zen learn experiment --analyze --significance 0.05
zen learn decision --ship-or-revert
```

## Anti-Patterns to Avoid

### Process Anti-Patterns

#### ❌ Waterfall in Disguise

```bash
# Bad: Batching all stages
zen align init "Q2 Features" # 10 features
zen discover --all-features  # 3 weeks
zen design --all-features    # 3 weeks
zen build --all-features     # 6 weeks
# Result: 3 months before any value delivered
```

#### ✅ Iterative Delivery

```bash
# Good: Small cycles
zen align init "Feature 1"  # 1 feature
zen workflow complete        # 2 weeks
zen align init "Feature 2"   # Next feature
# Result: Value every 2 weeks
```

### Technical Anti-Patterns

#### ❌ Big Bang Deployments

```bash
# Bad: Massive release
zen ship deploy --changes 147 --to-production
# High risk, hard to debug issues
```

#### ✅ Incremental Releases

```bash
# Good: Small, frequent releases
zen ship deploy --changes 5 --feature-flagged
# Lower risk, easy rollback
```

### Team Anti-Patterns

#### ❌ Hero Culture

```bash
# Bad: Single person owns everything
zen assign --all-stages --to "@superhero"
# Bottleneck, burnout risk
```

#### ✅ Distributed Ownership

```bash
# Good: Team ownership
zen assign --by-expertise
zen backup --assign --for-each-role
# Resilient, sustainable
```

## Optimization Tips

### Performance Optimization

```bash
# Parallelize where possible
zen discover research --parallel \
  --user-interviews \
  --data-analysis \
  --competitive-research

# Cache expensive operations
zen config cache --enable \
  --metrics-calculation \
  --report-generation

# Optimize slow commands
zen profile --identify-bottlenecks
zen optimize --command "ship validate"
```

### Workflow Optimization

```bash
# Identify bottlenecks
zen analyze workflow --bottlenecks
# Bottleneck: Code review (avg 8 hours)

# Implement improvements
zen optimize code-review \
  --auto-assign-reviewers \
  --parallel-reviews \
  --size-limits

# Measure improvement
zen metrics compare --before-after
# Review time: 8 hrs → 2 hrs (-75%)
```

## Scaling Zenflow

### From Team to Organization

#### Start with Pilot Team

```bash
# Phase 1: Single team
zen rollout --pilot-team "mobile"
zen support --dedicated --weeks 4
zen success --measure --share

# Phase 2: Early adopters
zen rollout --teams "web,api"
zen champions --identify --train

# Phase 3: Organization-wide
zen rollout --all-teams
zen governance --establish
zen metrics --standardize
```

#### Establish Center of Excellence

```bash
# Create CoE
zen coe create \
  --members "senior-engineers,architects" \
  --charter "standards,training,support"

# CoE activities
zen coe standards --define
zen coe training --deliver
zen coe support --provide
zen coe innovation --drive
```

### Customization vs Standardization

Find the right balance:

```bash
# Standardize core workflow
zen standards core --mandatory \
  --stages "all" \
  --quality-gates "security,testing"

# Allow team customization
zen customize allow \
  --metrics "team-specific" \
  --tools "preferred" \
  --thresholds "adjusted"
```

## Success Metrics

### Leading Indicators

Track early signs of success:

```bash
zen metrics leading
# Stage completion rate: 95%
# Quality gate pass rate: 88%
# Team satisfaction: 8.2/10
```

### Lagging Indicators

Measure actual outcomes:

```bash
zen metrics lagging
# Cycle time: -40%
# Defect rate: -60%
# Deployment frequency: +300%
# MTTR: -50%
```

### Continuous Improvement

```bash
# Regular retrospectives
zen retro schedule --biweekly
zen retro analyze --trends
zen retro actions --implement

# Quarterly reviews
zen review quarterly \
  --assess-metrics \
  --gather-feedback \
  --plan-improvements
```

## Summary

Key principles for Zenflow success:

1. **Start small** - Build confidence with low-risk projects
2. **Complete stages** - Don't skip, each stage adds value
3. **Measure outcomes** - Data drives better decisions
4. **Iterate quickly** - Small cycles reduce risk
5. **Collaborate actively** - Cross-functional teams win
6. **Automate repetitively** - Let tools handle routine tasks
7. **Learn continuously** - Every cycle improves the next
8. **Share knowledge** - Team learning accelerates progress

Remember: Zenflow is a journey, not a destination. Continuous improvement is the goal.
