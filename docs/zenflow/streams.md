# Zenflow Streams

Streams are specialized implementations of the seven-stage workflow optimized for specific types of work and team configurations. Each stream maintains the core Zenflow structure while adapting activities and emphasis to its purpose.

## Understanding Streams

### What Are Streams?

Streams are workflow variations that:

- **Maintain consistency** - Same seven stages, different focus
- **Optimize for context** - Tailored activities for specific work types
- **Enable specialization** - Teams focus on their strengths
- **Support handoffs** - Clear integration points between streams

### Core Principle

**Same Verbs, Different Focus**

All streams use the same seven stages but emphasize different aspects:

| Stage | I2D Focus | C2M Focus | D2S Focus |
|-------|-----------|-----------|-----------|
| Align | User problem | Technical approach | Operational goals |
| Discover | User research | Code analysis | System assessment |
| Prioritize | Feature value | Technical debt | Performance issues |
| Design | UX/UI design | Code architecture | Infrastructure design |
| Build | Prototypes | Implementation | Automation |
| Ship | User testing | Code deployment | Production release |
| Learn | User feedback | Code metrics | System metrics |

## Available Streams

### I2D - Idea to Delivery

Transform ideas into validated, production-ready specifications.

#### Purpose

Focus on product discovery, user research, and design validation before engineering implementation.

#### Team Composition

- Product Manager (Lead)
- UX/UI Designer
- User Researcher
- Engineering Lead (Advisory)
- Data Analyst

#### Workflow Emphasis

```bash
# I2D stream initialization
zen workflow create --stream i2d "New Feature Discovery"

# Stage emphasis
zen align    # Define user problem and success metrics
zen discover # Deep user research and market analysis
zen design   # Create detailed UX specifications
zen build    # Build prototypes for validation
```

#### Key Activities

**Align Stage**
- Define user problems and pain points
- Establish user success metrics
- Identify target user segments

**Discover Stage**
- Conduct user interviews and surveys
- Analyze user behavior data
- Research competitive solutions
- Map user journeys

**Design Stage**
- Create wireframes and mockups
- Design user flows
- Build interactive prototypes
- Define acceptance criteria

**Build Stage**
- Create functional prototypes
- Implement proof of concepts
- Prepare handoff documentation

#### Artifacts

- User research reports
- Journey maps
- Wireframes and prototypes
- Product specifications
- Acceptance criteria

#### Success Metrics

```bash
# I2D specific metrics
zen metrics i2d
# Experiment velocity: 3 per week
# Validation success rate: 72%
# Time to specification: 8 days average
# Design debt accumulated: 2 items
```

#### Example I2D Cycle

```bash
# Week 1: Problem definition
zen align init "Improve checkout conversion"
zen align metrics --add "Cart abandonment" --target "-30%"

# Week 2: User research
zen discover research --method "user-interview" --participants 15
zen discover research --method "session-replay" --sessions 100
zen discover finding "Users confused by shipping options"

# Week 3: Design solutions
zen design ux --wireframe "simplified-checkout"
zen design prototype --interactive "checkout-flow-v2"
zen design test --users 10

# Week 4: Validate and handoff
zen build prototype --functional
zen ship test --with-users 50
zen learn analyze --conversion-impact
```

### C2M - Code to Market

Transform specifications into high-quality, production-ready code.

#### Purpose

Focus on engineering implementation, code quality, and technical excellence.

#### Team Composition

- Engineering Lead
- Senior Engineers
- Software Engineers
- QA Engineers
- DevOps Engineer

#### Workflow Emphasis

```bash
# C2M stream initialization
zen workflow create --stream c2m "Payment System Implementation"

# Stage emphasis
zen design   # Technical architecture and API design
zen build    # Code implementation and testing
zen ship     # Deployment preparation
zen learn    # Code quality metrics
```

#### Key Activities

**Design Stage**
- Design system architecture
- Define API contracts
- Plan database schemas
- Create technical specifications

**Build Stage**
- Implement features
- Write comprehensive tests
- Conduct code reviews
- Update documentation

**Ship Stage**
- Run integration tests
- Perform security scans
- Prepare deployment packages
- Configure feature flags

**Learn Stage**
- Analyze code metrics
- Review performance data
- Identify technical debt
- Plan refactoring

#### Artifacts

- Technical design documents
- API specifications
- Source code
- Test suites
- Deployment packages

#### Success Metrics

```bash
# C2M specific metrics
zen metrics c2m
# Velocity: 45 story points/sprint
# Code coverage: 87%
# Defect escape rate: 0.3%
# Pull request cycle time: 4.2 hours
```

#### Example C2M Cycle

```bash
# Sprint 1: Technical design
zen design contract --create "payment-api"
zen design architecture --document "microservices"
zen design schema --define "payment_tables"

# Sprint 2-3: Implementation
zen build feature --start "stripe-integration"
zen build test --write --coverage-target 85
zen build review --submit --reviewers "@senior-team"

# Sprint 4: Quality and deployment
zen ship validate --all-tests
zen ship security --scan --fix-critical
zen ship package --create --version "2.1.0"
zen learn metrics --code-quality
```

### D2S - Deploy to Scale

Deploy and scale systems with operational excellence.

#### Purpose

Focus on production deployment, system reliability, and operational efficiency.

#### Team Composition

- SRE Lead
- Site Reliability Engineers
- DevOps Engineers
- Infrastructure Engineers
- Security Engineer

#### Workflow Emphasis

```bash
# D2S stream initialization
zen workflow create --stream d2s "Production Scaling Initiative"

# Stage emphasis
zen align    # Define SLOs and operational goals
zen discover # Assess current system state
zen design   # Infrastructure architecture
zen ship     # Production deployment
zen learn    # System metrics and optimization
```

#### Key Activities

**Align Stage**
- Define Service Level Objectives (SLOs)
- Establish operational goals
- Set performance budgets

**Discover Stage**
- Analyze system performance
- Identify bottlenecks
- Assess infrastructure costs
- Review security posture

**Design Stage**
- Design infrastructure architecture
- Plan scaling strategies
- Define monitoring approach
- Create disaster recovery plans

**Ship Stage**
- Execute deployments
- Configure monitoring
- Implement auto-scaling
- Validate rollback procedures

**Learn Stage**
- Analyze system metrics
- Review incident reports
- Calculate SLO achievement
- Optimize infrastructure

#### Artifacts

- Infrastructure as Code
- Deployment manifests
- Monitoring dashboards
- Runbooks
- Post-mortem reports

#### Success Metrics

```bash
# D2S specific metrics
zen metrics d2s
# Deployment frequency: 12 per day
# MTTR: 8 minutes
# SLO achievement: 99.95%
# Infrastructure cost: -15% MoM
# Incident rate: 0.3 per week
```

#### Example D2S Cycle

```bash
# Week 1: Operational planning
zen align slo --define "API latency" --target "p99 < 100ms"
zen align slo --define "Availability" --target "99.99%"

# Week 2: System assessment
zen discover performance --analyze --period "30d"
zen discover bottlenecks --identify
zen discover costs --analyze --breakdown

# Week 3: Infrastructure design
zen design infrastructure --architecture "multi-region"
zen design scaling --policy "cpu > 70%"
zen design monitoring --dashboards

# Week 4: Deployment and monitoring
zen ship deploy --blue-green
zen ship monitor --slo-tracking
zen learn incident --analyze --last-30d
```

## Stream Integration

### Handoff Points

Streams connect at specific stages for seamless collaboration.

#### I2D → C2M Handoff

Occurs at Build stage:

```bash
# I2D completes specification
zen i2d build complete --artifacts "specs,wireframes,contracts"

# C2M receives handoff
zen c2m receive --from i2d --validate
# ✓ API contracts complete
# ✓ Acceptance criteria defined
# ✓ UX designs approved
# ✓ Test scenarios documented

# C2M begins implementation
zen c2m build start --from-specs
```

**Handoff Checklist:**
- [ ] Product specifications complete
- [ ] API contracts defined
- [ ] UX designs finalized
- [ ] Acceptance criteria clear
- [ ] Test scenarios documented
- [ ] Dependencies identified

#### C2M → D2S Handoff

Occurs at Ship stage:

```bash
# C2M prepares deployment package
zen c2m ship prepare --package "v2.1.0"

# D2S receives package
zen d2s receive --from c2m --package "v2.1.0"
# ✓ All tests passing
# ✓ Security scan clean
# ✓ Performance validated
# ✓ Documentation complete

# D2S deploys to production
zen d2s ship deploy --canary
```

**Handoff Checklist:**
- [ ] Code tested and reviewed
- [ ] Deployment package created
- [ ] Configuration documented
- [ ] Monitoring requirements defined
- [ ] Rollback procedures tested
- [ ] Operational runbooks updated

### Parallel Execution

Teams can run multiple streams simultaneously:

```bash
# View active streams
zen streams list
# I2D: Next Quarter Planning [Discover]
# C2M: Current Sprint [Build]
# D2S: Production Hotfix [Ship]

# Coordinate across streams
zen streams coordinate --sync-point "weekly-standup"
```

### Cross-Stream Communication

```bash
# I2D requests technical feasibility
zen i2d request --to c2m --type "feasibility" \
  --feature "real-time-notifications"

# C2M responds with assessment
zen c2m respond --to i2d --feasibility "possible" \
  --effort "3 sprints" --risks "websocket-scaling"

# D2S provides operational constraints
zen d2s inform --constraint "deployment-freeze" \
  --dates "2024-12-20 to 2025-01-03"
```

## Stream Selection

### Choosing the Right Stream

Use this decision tree to select appropriate streams:

```
Is this primarily about:
├─ Understanding users and defining products?
│  └─ Use I2D Stream
├─ Building and testing code?
│  └─ Use C2M Stream
└─ Deploying and operating systems?
   └─ Use D2S Stream
```

### Stream Combinations

Common patterns for different scenarios:

#### New Feature Development
```
I2D → C2M → D2S
Full cycle from idea to production
```

#### Bug Fix
```
C2M → D2S
Direct implementation and deployment
```

#### Performance Optimization
```
D2S → C2M → D2S
Identify issue, implement fix, deploy
```

#### User Research
```
I2D only
Discovery without immediate implementation
```

## Stream Customization

### Adapting Streams

Customize streams for your organization:

```bash
# Create custom stream
zen stream create --name "mobile" \
  --base "c2m" \
  --focus "mobile-development"

# Configure stream stages
zen stream configure "mobile" \
  --stage "design" --emphasis "responsive-ui"
  --stage "build" --tools "react-native,flutter"
  --stage "ship" --platforms "ios,android"
```

### Industry-Specific Streams

#### Healthcare Stream
```bash
zen stream create --name "healthcare" \
  --compliance "hipaa" \
  --validation "fda" \
  --security "enhanced"
```

#### Financial Services Stream
```bash
zen stream create --name "fintech" \
  --compliance "pci-dss,sox" \
  --audit "required" \
  --encryption "mandatory"
```

#### E-commerce Stream
```bash
zen stream create --name "ecommerce" \
  --focus "conversion,performance" \
  --testing "a/b-required" \
  --analytics "enhanced"
```

## Stream Metrics

### Performance Tracking

Monitor stream effectiveness:

```bash
# Stream velocity
zen metrics stream --velocity
# I2D: 2 experiments/week
# C2M: 40 points/sprint
# D2S: 8 deployments/day

# Stream quality
zen metrics stream --quality
# I2D: 75% validation success
# C2M: 0.5% defect rate
# D2S: 99.97% availability

# Stream efficiency
zen metrics stream --efficiency
# I2D: 6 day cycle time
# C2M: 3.5 hour PR cycle
# D2S: 12 minute MTTR
```

### Cross-Stream Metrics

```bash
# End-to-end metrics
zen metrics e2e
# Idea to production: 28 days average
# Handoff success rate: 94%
# Rework due to handoff: 8%
```

### Stream Health

```bash
# Check stream health
zen health stream --all
# I2D: ✓ Healthy (all metrics green)
# C2M: ⚠ Warning (velocity below target)
# D2S: ✓ Healthy (SLOs met)

# Get recommendations
zen recommend stream --improve "c2m"
# 1. Reduce PR size (current avg: 450 lines)
# 2. Increase test automation (manual: 30%)
# 3. Parallelize code reviews
```

## Best Practices

### Stream-Specific Tips

#### I2D Best Practices
1. **User-first thinking** - Always validate with real users
2. **Rapid experimentation** - Fail fast and iterate
3. **Clear specifications** - Remove ambiguity for handoffs
4. **Data-driven decisions** - Base choices on evidence

#### C2M Best Practices
1. **Small increments** - Ship small, review often
2. **Test obsession** - Comprehensive automated testing
3. **Code quality** - Maintain high standards consistently
4. **Documentation** - Keep docs in sync with code

#### D2S Best Practices
1. **Automation everything** - Manual processes don't scale
2. **Monitor proactively** - Detect issues before users
3. **Practice failures** - Regular disaster recovery drills
4. **Cost awareness** - Optimize infrastructure spending

### Handoff Excellence

#### Clear Contracts
```bash
# Define handoff contract
zen handoff define --from "i2d" --to "c2m" \
  --requires "specs,designs,criteria" \
  --validates "completeness,clarity"
```

#### Validation Gates
```bash
# Automated handoff validation
zen handoff validate --from "i2d" --to "c2m"
# ✓ All required artifacts present
# ✓ Acceptance criteria measurable
# ✓ Dependencies documented
# ✓ No blocking issues
```

#### Feedback Loops
```bash
# Provide feedback to upstream
zen feedback send --to "i2d" \
  --issue "Ambiguous acceptance criteria for feature X"
  
# Track feedback resolution
zen feedback track --status
# Issue #123: Resolved
# Issue #124: In Progress
```

## Advanced Stream Patterns

### Dynamic Streams

Adapt streams based on project characteristics:

```bash
# Risk-based stream selection
zen stream select --risk-level "high"
# Recommended: I2D → C2M → D2S (full validation)

zen stream select --risk-level "low"
# Recommended: C2M → D2S (skip discovery)
```

### Hybrid Streams

Combine stream characteristics:

```bash
# Create hybrid stream
zen stream create --hybrid \
  --name "rapid-feature" \
  --combine "i2d:discover,c2m:build,d2s:ship"
```

### Stream Orchestration

Coordinate multiple streams:

```bash
# Orchestrate parallel streams
zen orchestrate start --streams "i2d,c2m,d2s"
zen orchestrate sync --checkpoint "sprint-end"
zen orchestrate report --consolidated
```

## Troubleshooting Streams

### Common Issues

#### Handoff Delays
```bash
# Diagnose handoff issues
zen diagnose handoff --from "i2d" --to "c2m"
# Issue: Missing API documentation
# Impact: 2 day delay
# Resolution: Add to handoff checklist
```

#### Stream Bottlenecks
```bash
# Identify bottlenecks
zen analyze stream --bottlenecks "c2m"
# Bottleneck: Code review (avg 8 hours)
# Suggestion: Add more reviewers
# Suggestion: Reduce PR size
```

#### Quality Issues
```bash
# Analyze quality problems
zen analyze quality --stream "i2d"
# Issue: Low validation success (45%)
# Root cause: Insufficient user research
# Recommendation: Increase sample size
```

## Future Streams

### Planned Streams

#### S2R - Support to Resolution
Focus on customer support and issue resolution

#### M2I - Market to Insights
Market research and competitive intelligence

#### T2P - Threat to Protection
Security and compliance management

### Stream Evolution

Streams evolve based on:
- Team feedback
- Success metrics
- Industry trends
- Tool capabilities

```bash
# Suggest stream improvements
zen improve stream --suggest "c2m"
# 1. Integrate AI code review
# 2. Add mutation testing
# 3. Implement canary analysis
```

## Summary

Zenflow Streams provide specialized implementations while maintaining consistency:

- **I2D Stream** - Product discovery and validation
- **C2M Stream** - Engineering implementation
- **D2S Stream** - Production operations

Key principles:
- Same seven stages, different emphasis
- Clear handoff points between streams
- Parallel execution for efficiency
- Continuous improvement based on metrics

Choose the right stream for your work type and let specialization drive excellence while maintaining workflow consistency.
