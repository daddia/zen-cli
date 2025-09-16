# Quality Gates

Quality gates are automated and manual checkpoints that ensure each Zenflow stage meets defined standards before progression. They enforce consistency, reduce risk, and maintain alignment with business objectives.

## Overview

### What Are Quality Gates?

Quality gates are validation checkpoints that:

- **Prevent progression** until criteria are met
- **Enforce standards** consistently across teams
- **Reduce risk** by catching issues early
- **Maintain quality** throughout development
- **Provide transparency** on readiness status

### Core Principles

1. **Fail Fast, Fix Early** - Identify issues close to their source
2. **Automated First** - Prefer automated checks over manual reviews
3. **Measurable Criteria** - Use objective, quantifiable standards
4. **Progressive Rigor** - Increase strictness approaching production

## Gate Types

### Automated Gates

Automated gates run without human intervention:

```bash
# Example automated checks
zen validate --stage build
# ✓ Code coverage: 85% (threshold: 80%)
# ✓ Linting: No issues
# ✓ Security scan: Clean
# ✓ Tests: 156/156 passing
```

**Common Automated Gates:**
- Code quality and style checks
- Test coverage thresholds
- Security vulnerability scans
- Performance benchmarks
- Documentation completeness
- API contract validation

### Manual Gates

Manual gates require human review:

```bash
# Example manual reviews
zen review request --type "architecture"
# Waiting for review from: @senior-architect
# Status: In Review
```

**Common Manual Gates:**
- Stakeholder approval
- Architecture review
- UX design validation
- Code review
- Business sign-off
- Risk assessment

### Hybrid Gates

Some gates combine automation with manual review:

```bash
# Automated analysis + manual decision
zen gate evaluate --type "security"
# Automated scan: 2 medium risks found
# Manual review required for risk acceptance
```

## Stage-Specific Gates

Each stage has unique quality requirements:

### Stage 1: Align - Leadership Sign-off

**Purpose:** Ensure strategic alignment before tactical work

**Automated Checks:**
```bash
zen align validate
# ✓ PR/FAQ document exists
# ✓ Success metrics defined
# ✓ Timeline specified
# ✓ Budget allocated
```

**Manual Reviews:**
- Business case approval
- Technical feasibility review
- Resource commitment confirmation

**Gate Criteria:**
| Criterion | Type | Requirement |
|-----------|------|-------------|
| Problem statement | Manual | Clear and quantified |
| Success metrics | Automated | Measurable with targets |
| Stakeholder list | Automated | Complete with roles |
| Resource plan | Manual | Approved and realistic |
| Risk assessment | Hybrid | Identified with mitigation |

### Stage 2: Discover - Risk Capture

**Purpose:** Ensure comprehensive understanding before design

**Automated Checks:**
```bash
zen discover validate
# ✓ Research artifacts: 12 found
# ✓ Stakeholder coverage: 100%
# ✓ Assumptions logged: 8
# ✓ Risks documented: 5
```

**Manual Reviews:**
- Research quality assessment
- Assumption prioritization
- Risk mitigation planning

**Gate Criteria:**
| Criterion | Type | Requirement |
|-----------|------|-------------|
| Stakeholder interviews | Automated | All groups covered |
| User research | Manual | Insights documented |
| Technical constraints | Automated | Identified and validated |
| Assumptions | Automated | Explicit with tests |
| Risk register | Hybrid | Complete with strategies |

### Stage 3: Prioritize - Capacity Validation

**Purpose:** Ensure realistic planning and resource allocation

**Automated Checks:**
```bash
zen prioritize validate
# ✓ All items scored
# ✓ Capacity calculation: 95% utilized
# ✓ Dependencies mapped
# ✓ Release plan created
```

**Manual Reviews:**
- Priority alignment with strategy
- Capacity realism check
- Dependency risk assessment

**Gate Criteria:**
| Criterion | Type | Requirement |
|-----------|------|-------------|
| Prioritization scores | Automated | Method consistently applied |
| Effort estimates | Hybrid | Provided and validated |
| Capacity check | Automated | Within available resources |
| Dependencies | Automated | Identified and sequenced |
| Stakeholder alignment | Manual | Confirmed on priorities |

### Stage 4: Design - Specification Review

**Purpose:** Ensure complete, implementable specifications

**Automated Checks:**
```bash
zen design validate
# ✓ API contracts: Valid OpenAPI 3.0
# ✓ Schema validation: Pass
# ✓ Breaking changes: None detected
# ✓ Design coverage: 100%
```

**Manual Reviews:**
- Architecture decision review
- UX design validation
- Security design assessment

**Gate Criteria:**
| Criterion | Type | Requirement |
|-----------|------|-------------|
| API contracts | Automated | Complete and valid |
| UX designs | Manual | User-tested and approved |
| Architecture | Manual | Documented with ADRs |
| Data models | Automated | Support all use cases |
| Migration plan | Hybrid | Feasible and tested |

### Stage 5: Build - Code Quality

**Purpose:** Ensure implementation meets quality standards

**Automated Checks:**
```bash
zen build validate
# ✓ Unit test coverage: 87%
# ✓ Integration tests: Pass
# ✓ Linting: No violations
# ✓ Security scan: Clean
# ✓ Documentation: Updated
```

**Manual Reviews:**
- Code review approval
- Design compliance check
- Documentation review

**Gate Criteria:**
| Criterion | Type | Requirement |
|-----------|------|-------------|
| Test coverage | Automated | >80% with meaningful tests |
| Code quality | Automated | Linting and standards pass |
| Security | Automated | No high/critical issues |
| Code review | Manual | Approved by peers |
| Documentation | Hybrid | Accurate and complete |

### Stage 6: Ship - Production Readiness

**Purpose:** Ensure safe, reliable production deployment

**Automated Checks:**
```bash
zen ship validate
# ✓ All tests: 423/423 passing
# ✓ Security: No vulnerabilities
# ✓ Performance: Within budgets
# ✓ Accessibility: WCAG 2.1 AA compliant
# ✓ Monitoring: Configured
```

**Manual Reviews:**
- Deployment approval
- Risk acceptance
- Go/no-go decision

**Gate Criteria:**
| Criterion | Type | Requirement |
|-----------|------|-------------|
| Test suites | Automated | All passing |
| Security assessment | Automated | Acceptable risk level |
| Performance | Automated | Meets budgets |
| Canary metrics | Hybrid | Positive or neutral |
| Rollback plan | Manual | Tested and ready |

### Stage 7: Learn - Outcome Validation

**Purpose:** Ensure learnings drive improvement

**Automated Checks:**
```bash
zen learn validate
# ✓ Metrics collected: 15/15
# ✓ Statistical significance: Yes
# ✓ Feedback processed: 127 items
# ✓ Report generated
```

**Manual Reviews:**
- Business impact assessment
- Learning synthesis
- Next iteration planning

**Gate Criteria:**
| Criterion | Type | Requirement |
|-----------|------|-------------|
| Success metrics | Automated | Measured with significance |
| User feedback | Automated | Collected and analyzed |
| Business impact | Manual | Quantified and validated |
| Learnings | Manual | Documented and actionable |
| Next priorities | Manual | Defined and justified |

## Gate Configuration

### Setting Thresholds

Configure quality gate thresholds:

```bash
# Set test coverage requirement
zen config gates.build.test-coverage --min 85

# Set performance budget
zen config gates.ship.response-time --max "200ms"

# Set security policy
zen config gates.ship.vulnerabilities.critical --max 0
zen config gates.ship.vulnerabilities.high --max 2
```

### Custom Gates

Add custom gates for specific needs:

```bash
# Define custom gate
zen gate create "license-compliance" \
  --stage build \
  --type automated \
  --script "scripts/check-licenses.sh"

# Configure gate criteria
zen gate configure "license-compliance" \
  --fail-on "GPL,AGPL" \
  --warn-on "LGPL"
```

### Override Mechanism

Override gates with proper authorization:

```bash
# Request override (requires approval)
zen gate override --stage ship \
  --gate "security-scan" \
  --reason "Known false positive in dependency" \
  --approver "@security-lead"

# Emergency override (logged and audited)
zen gate override --emergency \
  --stage ship \
  --reason "Critical hotfix for production issue"
```

## Gate Automation

### CI/CD Integration

Integrate gates into your pipeline:

```yaml
# Example GitHub Actions integration
- name: Validate Build Quality Gates
  run: |
    zen build validate --strict
    if [ $? -ne 0 ]; then
      echo "Quality gates failed"
      exit 1
    fi
```

### Pre-commit Hooks

Enforce gates before code commit:

```bash
# .git/hooks/pre-commit
#!/bin/bash
zen validate --local --stage build
if [ $? -ne 0 ]; then
  echo "Local quality gates failed. Fix issues before committing."
  exit 1
fi
```

### Monitoring and Alerts

Set up gate monitoring:

```bash
# Configure alerts
zen monitor gates --alert-on-failure
zen monitor gates --slack-webhook "$WEBHOOK_URL"

# View gate metrics
zen metrics gates --period "30d"
# Pass rate: 94%
# Average resolution time: 2.3 hours
# Most common failure: Test coverage
```

## Gate Enforcement Strategies

### Strict Mode

No exceptions, all gates must pass:

```bash
zen config enforcement --mode strict
# All gates are mandatory
# No overrides allowed
# Recommended for: Critical systems
```

### Standard Mode

Normal enforcement with override capability:

```bash
zen config enforcement --mode standard
# Gates enforced by default
# Overrides require approval
# Recommended for: Most projects
```

### Lenient Mode

Gates as guidelines with warnings:

```bash
zen config enforcement --mode lenient
# Gates generate warnings
# Progression allowed with acknowledgment
# Recommended for: Prototypes, experiments
```

## Progressive Gate Enhancement

### Maturity Levels

Gates evolve with team maturity:

#### Level 1: Basic
- Test coverage > 60%
- No critical security issues
- Manual code review

#### Level 2: Intermediate
- Test coverage > 70%
- No high security issues
- Automated linting
- Performance monitoring

#### Level 3: Advanced
- Test coverage > 80%
- No medium security issues
- Mutation testing
- Accessibility validation
- Automated documentation

#### Level 4: Expert
- Test coverage > 90%
- Property-based testing
- Formal verification
- Chaos engineering
- Full automation

### Gradual Introduction

Implement gates progressively:

```bash
# Week 1: Start with critical gates
zen gates enable --level critical
# - Security scanning
# - Test execution
# - Build success

# Week 2: Add quality gates
zen gates enable --level quality
# + Code coverage
# + Linting
# + Documentation

# Week 3: Add performance gates
zen gates enable --level performance
# + Response time
# + Resource usage
# + Load testing

# Week 4: Full enforcement
zen gates enable --all
```

## Gate Metrics and Reporting

### Dashboard

View gate status and trends:

```bash
zen dashboard gates
# Current Status: 
# ├─ Align: ✓ Complete
# ├─ Discover: ✓ Complete
# ├─ Prioritize: ✓ Complete
# ├─ Design: ✓ Complete
# ├─ Build: ⚠ In Progress (2 gates pending)
# ├─ Ship: ○ Not Started
# └─ Learn: ○ Not Started
```

### Reports

Generate gate compliance reports:

```bash
# Weekly report
zen report gates --period "1w" --format pdf
# Report generated: gates-report-2024-w12.pdf

# Team metrics
zen report gates --by-team
# Team A: 96% pass rate
# Team B: 89% pass rate
# Team C: 94% pass rate

# Failure analysis
zen report gates --failures --analyze
# Top failure reasons:
# 1. Insufficient test coverage (34%)
# 2. Linting violations (28%)
# 3. Security vulnerabilities (18%)
```

### Historical Analysis

Track gate performance over time:

```bash
zen analyze gates --historical
# Q1: 87% pass rate, 3.2h avg resolution
# Q2: 91% pass rate, 2.8h avg resolution
# Q3: 94% pass rate, 2.3h avg resolution
# Trend: Improving ↑
```

## Troubleshooting Gates

### Common Issues

#### Gate Keeps Failing

```bash
# Diagnose specific gate
zen diagnose gate "test-coverage"
# Current: 78%
# Required: 80%
# Uncovered files:
# - src/utils/helpers.js (45%)
# - src/api/auth.js (62%)
```

#### Gates Taking Too Long

```bash
# Profile gate performance
zen profile gates
# Security scan: 12m 34s ⚠ (target: <10m)
# Test suite: 8m 12s ✓
# Linting: 45s ✓

# Optimize slow gates
zen optimize gate "security-scan" --parallel
```

#### False Positives

```bash
# Mark false positive
zen gate false-positive \
  --gate "security-scan" \
  --issue "CVE-2024-1234" \
  --reason "Not applicable to our usage"

# Configure suppressions
zen gate suppress --rule "no-console" --file "debug.js"
```

### Getting Help

When gates block progress:

1. **Check details**: `zen gate details <gate-name>`
2. **View suggestions**: `zen gate suggest-fix <gate-name>`
3. **Request help**: `zen gate help-request <gate-name>`
4. **Schedule review**: `zen gate review-request <gate-name>`

## Best Practices

### Do's

✅ **Start simple** - Add gates gradually
✅ **Automate first** - Reduce manual review burden
✅ **Document overrides** - Always explain exceptions
✅ **Monitor trends** - Track gate performance
✅ **Iterate thresholds** - Adjust based on team capability

### Don'ts

❌ **Skip gates for speed** - Technical debt compounds
❌ **Set unrealistic thresholds** - Causes frustration
❌ **Ignore failures** - Gates exist for a reason
❌ **Override without review** - Defeats the purpose
❌ **Add too many gates** - Balance quality with velocity

### Tips for Teams

#### For New Teams
- Start with essential gates only
- Focus on automation over manual review
- Set achievable initial thresholds
- Increase rigor gradually

#### For Mature Teams
- Implement comprehensive gate coverage
- Optimize gate performance
- Use gates for continuous improvement
- Share gate configurations across projects

#### For Distributed Teams
- Emphasize automated gates
- Document gate requirements clearly
- Use async review processes
- Maintain gate dashboards

## Gate Evolution

### Continuous Improvement

Gates should evolve based on:

- **Team feedback** - Regular retrospectives on gate effectiveness
- **Incident analysis** - Add gates to prevent repeat issues
- **Industry standards** - Adopt best practices
- **Tool capabilities** - Leverage new automation

### Future Enhancements

Planned gate improvements:

- **AI-powered reviews** - Intelligent code analysis
- **Predictive gates** - Forecast likely failures
- **Auto-remediation** - Fix common issues automatically
- **Cross-project learning** - Share gate insights

## Summary

Quality gates are essential for maintaining standards throughout Zenflow:

- **Protect quality** - Prevent issues from progressing
- **Enforce standards** - Consistent across all teams
- **Reduce risk** - Catch problems early
- **Provide transparency** - Clear readiness status
- **Drive improvement** - Metrics identify areas to enhance

Start with basic gates, automate where possible, and progressively enhance your quality standards as your team matures.
