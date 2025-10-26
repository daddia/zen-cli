# Zenflow Stages

This guide provides detailed documentation for each of the seven Zenflow stages. Each stage builds on the previous one to create a complete product development cycle.

## Stage Overview

| Stage | Purpose | Duration | Primary Team |
|-------|---------|----------|--------------|
| [Align](#align) | Define success | 1-3 days | Product, Leadership |
| [Discover](#discover) | Gather evidence | 3-5 days | Product, Design, Engineering |
| [Prioritize](#prioritize) | Rank by value | 1-2 days | Product, Engineering |
| [Design](#design) | Specify solution | 3-5 days | Design, Engineering |
| [Build](#build) | Implement | 5-10 days | Engineering |
| [Ship](#ship) | Deploy safely | 2-3 days | Engineering, Operations |
| [Learn](#learn) | Measure outcomes | 2-3 days | Product, Analytics |

## Align

### Purpose

Frame the problem, define success metrics, and establish constraints before any tactical work begins. This stage ensures everyone understands what success looks like and why the work matters.

### Key Activities

#### Product Management
- Define the business problem and opportunity size
- Research market context and competitive landscape
- Establish success metrics with measurement plans
- Identify stakeholders and decision makers

#### Engineering
- Assess technical feasibility
- Identify architectural implications
- Estimate resource requirements

### Commands

```bash
# Initialize strategy document
zen align init "<initiative-name>"

# Define success metrics
zen align metrics --add "<metric>" --target "<value>"

# Set constraints
zen align constraints --timeline "<duration>" --budget "<amount>"

# Get stakeholder approval
zen align approve --stakeholder "<name>"
```

### Example Workflow

```bash
# Start a new payment feature initiative
zen align init "International Payments"

# Define what success looks like
zen align metrics --add "Payment success rate" --target "99%"
zen align metrics --add "Processing time" --target "<2 seconds"
zen align metrics --add "Market expansion" --target "5 new countries"

# Document constraints
zen align constraints --timeline "Q2 2024" --budget "$500k"
zen align constraints --team "3 engineers, 1 designer, 1 PM"

# Capture risks early
zen align risk --add "Regulatory compliance varies by country" --impact "high"
```

### Artifacts Produced

- **PR/FAQ Document** - Press release and frequently asked questions
- **Success Metrics** - Measurable targets with baselines
- **Resource Plan** - Team, timeline, and budget allocation
- **Risk Register** - Known risks with mitigation strategies

### Quality Gate Checklist

- [ ] Business problem clearly articulated
- [ ] Success metrics defined and measurable
- [ ] Stakeholders identified and aligned
- [ ] Resources and timeline realistic
- [ ] Technical feasibility confirmed
- [ ] Leadership approval obtained

### Common Pitfalls

- **Vague metrics**: "Improve user experience" → "Reduce checkout time by 30%"
- **Solution jumping**: Starting with "We need a mobile app" instead of "Users can't pay on mobile"
- **Missing stakeholders**: Forgetting legal, security, or operations teams

## Discover

### Purpose

Gather comprehensive signals from users, stakeholders, market, and technical systems to inform solution design. Base decisions on evidence rather than assumptions.

### Key Activities

#### Product Management
- Conduct stakeholder interviews
- Analyze user feedback and support data
- Research market conditions
- Document assumptions

#### Design
- Run user research sessions
- Analyze existing user journeys
- Identify design patterns
- Plan prototypes

#### Engineering
- Assess technical constraints
- Research technology options
- Identify dependencies
- Document technical risks

### Commands

```bash
# Start discovery
zen discover start

# Capture research findings
zen discover research --method "user-interview" --finding "<insight>"
zen discover research --method "data-analysis" --finding "<pattern>"

# Document constraints
zen discover constraint --type "technical" --description "<limitation>"
zen discover constraint --type "business" --description "<requirement>"

# Log assumptions
zen discover assumption --add "<assumption>" --test-method "<validation>"
```

### Example Workflow

```bash
# Begin discovery phase
zen discover start

# User research findings
zen discover research --method "user-interview" \
  --finding "60% of users abandon cart at payment step"

zen discover research --method "survey" \
  --finding "Users want PayPal and Apple Pay options"

# Technical discoveries
zen discover technical --system "payment-gateway" \
  --finding "Current gateway doesn't support multi-currency"

# Market research
zen discover market --competitor "CompetitorX" \
  --finding "Offers one-click checkout for returning users"

# Key assumptions to test
zen discover assumption --add "Users will trust new payment providers" \
  --test-method "A/B test with small user segment"
```

### Artifacts Produced

- **Discovery Brief** - Synthesized research findings
- **User Journey Maps** - Current and desired user flows
- **Technical Analysis** - Constraints and opportunities
- **Assumption Log** - Hypotheses to validate
- **Test Strategy** - Validation approach for assumptions

### Quality Gate Checklist

- [ ] All stakeholder groups interviewed
- [ ] User research findings documented
- [ ] Technical constraints identified
- [ ] Market context understood
- [ ] Assumptions explicitly stated
- [ ] Risk mitigation planned
- [ ] Test strategy defined

### Common Pitfalls

- **Confirmation bias**: Only researching to prove existing beliefs
- **Analysis paralysis**: Endless research without decision points
- **Technical tunnel vision**: Ignoring user and business perspectives

## Prioritize

### Purpose

Rank solution options by impact, effort, and risk. Ensure resource allocation aligns with strategic objectives and team capacity.

### Key Activities

#### Product Management
- Apply prioritization frameworks (WSJF, ICE, RICE)
- Map business value for each option
- Coordinate stakeholder alignment
- Define release strategy

#### Engineering
- Provide effort estimates
- Identify dependencies
- Assess technical risk
- Plan capacity allocation

### Commands

```bash
# List items to prioritize
zen prioritize list

# Apply prioritization framework
zen prioritize score --method "ice"  # or wsjf, rice
zen prioritize score --item "Feature A" --impact 8 --confidence 7 --ease 5

# Check team capacity
zen prioritize capacity --team "mobile" --check

# Create release plan
zen prioritize release --create "v2.0" --target "2024-Q2"
zen prioritize release --assign "Feature A" --to "v2.0"
```

### Example Workflow

```bash
# View all discovered items
zen prioritize list
# 1. Multi-currency support
# 2. PayPal integration
# 3. Apple Pay integration
# 4. One-click checkout
# 5. Saved payment methods

# Score using ICE method
zen prioritize score --method "ice"
# Item 1: Impact=9, Confidence=8, Ease=3 → Score: 6.7
# Item 2: Impact=7, Confidence=9, Ease=7 → Score: 7.7
# Item 3: Impact=6, Confidence=9, Ease=8 → Score: 7.7
# Item 4: Impact=8, Confidence=5, Ease=4 → Score: 5.7
# Item 5: Impact=7, Confidence=8, Ease=6 → Score: 7.0

# Check capacity
zen prioritize capacity --sprint-weeks 4 --team-size 3
# Available: 12 person-weeks

# Assign to release
zen prioritize release --create "Payment v1" --items 2,3,5
# Total effort: 11 person-weeks ✓
```

### Prioritization Methods

#### WSJF (Weighted Shortest Job First)
```
Score = (Business Value + Time Criticality + Risk Reduction) / Job Size
```

#### ICE (Impact, Confidence, Ease)
```
Score = (Impact × Confidence × Ease) / 3
```

#### RICE (Reach, Impact, Confidence, Effort)
```
Score = (Reach × Impact × Confidence) / Effort
```

### Artifacts Produced

- **Ranked Backlog** - Prioritized list with scores
- **Effort Estimates** - Size and complexity assessments
- **Capacity Plan** - Resource allocation
- **Dependency Map** - Technical and business dependencies
- **Release Plan** - Milestone and delivery schedule

### Quality Gate Checklist

- [ ] Prioritization method applied consistently
- [ ] Effort estimates provided
- [ ] Capacity validated
- [ ] Dependencies identified
- [ ] Release plan created
- [ ] Stakeholder alignment confirmed

### Common Pitfalls

- **HiPPO decisions**: Highest paid person's opinion overrides data
- **Underestimating effort**: Forgetting testing, deployment, documentation
- **Ignoring dependencies**: Not considering prerequisite work

## Design

### Purpose

Define what will be built and how it will behave through detailed specifications. Create the blueprint for implementation.

### Key Activities

#### Design
- Create wireframes and prototypes
- Design user interfaces
- Ensure accessibility standards
- Validate with users

#### Engineering
- Define API contracts
- Design system architecture
- Plan data models
- Document decisions

### Commands

```bash
# Initialize design specifications
zen design init

# Create API contracts
zen design contract --create "user-api"
zen design contract --endpoint "POST /users" --request-schema "user-create.json"

# Design user experience
zen design ux --wireframe "checkout-flow"
zen design ux --prototype "payment-selection"

# Document architecture decisions
zen design adr --create "Use event-driven architecture for payments"

# Plan migrations
zen design migration --plan "add_payment_providers_table"
```

### Example Workflow

```bash
# Start design phase
zen design init

# Define API contracts
zen design contract --create "payment-api"
zen design contract --endpoint "POST /payments/charge" \
  --request "charge-request.yaml" \
  --response "charge-response.yaml"

zen design contract --endpoint "GET /payments/methods" \
  --response "payment-methods.yaml"

# Create UX designs
zen design ux --wireframe "payment-selection-screen"
zen design ux --interaction "saved-card-selection"
zen design ux --validate --with-users 5

# Architecture decisions
zen design adr --create "ADR-001: Use Stripe for payment processing"
zen design adr --rationale "Best multi-currency support and PCI compliance"

# Data model
zen design schema --table "payment_methods" \
  --fields "id,user_id,type,token,is_default"
```

### Artifacts Produced

- **API Specifications** - OpenAPI/GraphQL schemas
- **UX Designs** - Wireframes and prototypes
- **Architecture Decisions** - ADRs with rationale
- **Data Models** - Database schemas
- **Migration Plans** - Database change scripts

### Quality Gate Checklist

- [ ] API contracts complete and validated
- [ ] UX designs tested with users
- [ ] Architecture decisions documented
- [ ] Data models support all use cases
- [ ] Migration strategy defined
- [ ] Security requirements addressed
- [ ] Performance budgets set

### Common Pitfalls

- **Over-engineering**: Building for hypothetical future needs
- **Under-specifying**: Missing edge cases in contracts
- **Design debt**: Not updating design system components

## Build

### Purpose

Transform specifications into working software through implementation, testing, and review. Create high-quality, maintainable code.

### Key Activities

#### Engineering
- Generate code from contracts
- Implement features
- Write tests
- Conduct code reviews
- Update documentation

### Commands

```bash
# Generate scaffolding
zen build scaffold --from-contract "payment-api"

# Start feature development
zen build start --feature "payment-processing"

# Run tests
zen build test --unit
zen build test --integration
zen build test --e2e

# Submit for review
zen build review --create --branch "feature/payments"

# Validate quality gates
zen build validate
```

### Example Workflow

```bash
# Generate initial code
zen build scaffold --from-contract "payment-api"
# Generated: controllers/payment_controller.go
# Generated: models/payment.go
# Generated: tests/payment_test.go

# Implement feature
zen build start --feature "stripe-integration"

# Test thoroughly
zen build test --unit
# ✓ 48 unit tests pass
# Coverage: 87%

zen build test --integration
# ✓ 12 integration tests pass

# Check quality
zen build lint
# ✓ No issues found

zen build security-scan
# ✓ No vulnerabilities detected

# Submit for review
zen build review --create \
  --title "Add Stripe payment processing" \
  --reviewers "@alice,@bob"

# Monitor review status
zen build review --status
# Approved by @alice ✓
# Changes requested by @bob
```

### Development Practices

#### Test Coverage Requirements
- Unit tests: 80% minimum
- Integration tests: Critical paths
- E2E tests: User workflows

#### Code Quality Standards
- Linting rules enforced
- Security scanning required
- Performance benchmarks met

#### Review Guidelines
- Two approvals for critical code
- One approval for non-critical
- Automated checks must pass

### Artifacts Produced

- **Working Code** - Implemented features
- **Test Suites** - Comprehensive test coverage
- **Documentation** - API docs and guides
- **Review Records** - Code review history
- **Feature Flags** - Rollout configuration

### Quality Gate Checklist

- [ ] All tests passing
- [ ] Code coverage meets threshold
- [ ] Linting checks pass
- [ ] Security scan clean
- [ ] Documentation updated
- [ ] Code review approved
- [ ] Feature flags configured

### Common Pitfalls

- **Large pull requests**: Hard to review, higher risk
- **Insufficient testing**: Bugs escape to production
- **Documentation drift**: Code changes without doc updates

## Ship

### Purpose

Deploy software safely to production through comprehensive validation, controlled rollout, and monitoring. Ensure quality and reliability.

### Key Activities

#### Engineering
- Run comprehensive tests
- Perform security scans
- Execute performance tests
- Deploy progressively
- Monitor metrics

#### Operations
- Validate infrastructure
- Configure monitoring
- Prepare rollback plans
- Coordinate deployment

### Commands

```bash
# Validate readiness
zen ship validate --comprehensive

# Security assessment
zen ship security --scan --depth full

# Performance testing
zen ship performance --load-test --users 1000

# Deploy canary
zen ship deploy --canary --percentage 5

# Monitor deployment
zen ship monitor --metrics --duration "2h"

# Promote or rollback
zen ship promote --to-production
zen ship rollback --reason "<issue>"
```

### Example Workflow

```bash
# Pre-deployment validation
zen ship validate --comprehensive
# ✓ All unit tests pass (248/248)
# ✓ All integration tests pass (45/45)
# ✓ All E2E tests pass (12/12)
# ✓ Security scan: No critical issues
# ✓ Performance: Within budgets

# Deploy to staging
zen ship deploy --environment staging
# Deployment successful
# URL: https://staging.app.com

# Verify staging
zen ship verify --environment staging
# ✓ Health checks passing
# ✓ Smoke tests successful

# Start canary deployment
zen ship deploy --canary --percentage 5
# Deploying to 5% of production traffic

# Monitor canary
zen ship monitor --duration "4h"
# Hour 1: ✓ All metrics normal
# Hour 2: ✓ No errors detected
# Hour 3: ✓ Performance stable
# Hour 4: ✓ User feedback positive

# Gradual rollout
zen ship promote --percentage 25
# Expanded to 25% of traffic

zen ship promote --percentage 50
# Expanded to 50% of traffic

zen ship promote --to-production
# Full production deployment complete
```

### Deployment Strategies

#### Canary Release
- Start with 1-5% of traffic
- Monitor key metrics
- Gradually increase percentage
- Quick rollback if issues

#### Blue-Green
- Deploy to parallel environment
- Test thoroughly
- Switch traffic instantly
- Keep old version ready

#### Feature Flags
- Deploy code dark
- Enable for select users
- Monitor and iterate
- Full rollout when ready

### Artifacts Produced

- **Test Results** - Comprehensive validation
- **Security Report** - Vulnerability assessment
- **Performance Report** - Load test results
- **Deployment Log** - Rollout history
- **Monitoring Dashboard** - Real-time metrics

### Quality Gate Checklist

- [ ] All automated tests pass
- [ ] Security requirements met
- [ ] Performance budgets achieved
- [ ] Accessibility validated
- [ ] Canary metrics positive
- [ ] Rollback plan tested
- [ ] Monitoring active

### Common Pitfalls

- **Skipping canary phase**: Going directly to 100% is risky
- **Insufficient monitoring**: Not detecting issues quickly
- **No rollback plan**: Unable to recover from problems

## Learn

### Purpose

Close the feedback loop by measuring outcomes, synthesizing learnings, and planning improvements. Transform data into actionable insights.

### Key Activities

#### Product Management
- Analyze success metrics
- Synthesize user feedback
- Calculate business impact
- Plan next iterations

#### Analytics
- Correlate metrics
- Statistical analysis
- Generate insights
- Recommend improvements

### Commands

```bash
# Collect metrics
zen learn metrics --collect --period "2 weeks"

# Analyze outcomes
zen learn analyze --metrics --baseline

# Gather feedback
zen learn feedback --collect --channels "all"
zen learn feedback --synthesize

# Document learnings
zen learn document --insights

# Plan next iteration
zen learn recommend --next-actions
```

### Example Workflow

```bash
# Collect post-deployment metrics
zen learn metrics --collect --period "14 days"

# Analyze against targets
zen learn analyze
# Payment success rate: 98.5% (Target: 99%) ⚠
# Processing time: 1.8s average (Target: <2s) ✓
# Market expansion: 3 countries active (Target: 5) ⚠

# Deep dive into issues
zen learn investigate --metric "payment-success-rate"
# Finding: 1.5% failures due to network timeouts
# Recommendation: Implement retry logic

# User feedback
zen learn feedback --synthesize
# Positive: Fast checkout process (87% satisfaction)
# Negative: Confusing error messages (23% mentioned)
# Request: More payment options (45% requested)

# Business impact
zen learn impact --calculate
# Revenue increase: +$125k/month
# Support tickets: -15%
# User retention: +5%

# Document and share
zen learn document \
  --title "Payment System v1 Outcomes" \
  --learnings 5 \
  --recommendations 3

# Plan improvements
zen learn iterate --create "Payment System v1.1"
# Priority 1: Add retry logic for network failures
# Priority 2: Improve error messaging
# Priority 3: Add cryptocurrency payment option
```

### Analysis Methods

#### Statistical Analysis
- Significance testing
- Cohort analysis
- Regression analysis
- Trend identification

#### Qualitative Analysis
- Thematic analysis
- Sentiment scoring
- Journey mapping
- Pain point identification

### Artifacts Produced

- **Outcome Report** - Metrics vs targets
- **User Insights** - Feedback synthesis
- **Technical Analysis** - Performance review
- **Business Impact** - ROI calculation
- **Recommendations** - Next priorities

### Quality Gate Checklist

- [ ] Metrics measured and analyzed
- [ ] Statistical significance validated
- [ ] User feedback synthesized
- [ ] Business impact quantified
- [ ] Learnings documented
- [ ] Next actions prioritized
- [ ] Stakeholders informed

### Common Pitfalls

- **Vanity metrics**: Measuring what looks good vs what matters
- **Incomplete analysis**: Missing important data sources
- **No action**: Insights without follow-through

## Stage Transitions

### Moving Between Stages

Each stage must be complete before progressing:

```bash
# Check current status
zen status
# Current: Design (85% complete)
# Remaining: Contract validation, UX testing

# Complete remaining items
zen design contract --validate
zen design ux --test --users 5

# Progress to next stage
zen progress --to build
# ✓ Design stage complete
# → Entering Build stage
```

### Blocking Conditions

The workflow prevents skipping stages:

```bash
# Trying to ship without building
zen ship deploy
# Error: Build stage incomplete
# Required: Code review approval, test coverage > 80%
```

### Parallel Workflows

Teams can run multiple initiatives simultaneously:

```bash
# View all workflows
zen workflow list --active
# 1. Payment System [Ship]
# 2. Search Feature [Design]
# 3. Mobile App [Discover]

# Switch context
zen workflow switch "Search Feature"
# Switched to: Search Feature (Design stage)
```

## Best Practices

### General Guidelines

1. **Complete stages fully** - Don't rush to the next stage
2. **Document decisions** - Future you will thank present you
3. **Automate checks** - Let tools enforce quality
4. **Measure everything** - Data drives better decisions
5. **Iterate quickly** - Small cycles reduce risk

### Role-Specific Tips

#### For Product Managers
- Start with clear, measurable goals
- Involve stakeholders early and often
- Base priorities on data, not opinions
- Close the loop with outcome measurement

#### For Designers
- Test with real users, not assumptions
- Document design decisions
- Consider accessibility from the start
- Collaborate closely with engineering

#### For Engineers
- Generate code from specifications
- Write tests before implementation
- Keep pull requests small
- Automate everything possible

#### For Analytics Teams
- Define metrics before building
- Instrument comprehensively
- Validate data accuracy
- Share insights proactively

## Summary

The seven Zenflow stages create a complete product development cycle:

1. **Align** - Know what success looks like
2. **Discover** - Understand the problem space
3. **Prioritize** - Focus on highest value
4. **Design** - Specify the solution
5. **Build** - Implement with quality
6. **Ship** - Deploy safely
7. **Learn** - Measure and improve

Each stage builds on the previous one, with quality gates ensuring standards are met. By following this structured approach, teams can deliver valuable products predictably and efficiently.
