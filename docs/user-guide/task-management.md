# Task Management Guide

Master Zen's structured task management system with Zenflow methodology.

## Overview

Zen CLI provides a comprehensive task management system that implements the Zenflow methodology - a 7-stage unified workflow that standardizes how teams move from strategy to shipped value.

## Zenflow Methodology

### The Seven Stages

1. **Align** - Define success criteria and stakeholder alignment
2. **Discover** - Gather evidence and validate assumptions
3. **Prioritize** - Rank work by value vs effort
4. **Design** - Specify implementation approach
5. **Build** - Deliver working software increment
6. **Ship** - Deploy safely to production
7. **Learn** - Measure outcomes and iterate

### Task Types

Each task type optimizes the workflow for different kinds of work:

- **Story** - User-facing feature development with UX focus
- **Bug** - Defect fixes with root cause analysis
- **Epic** - Large initiatives requiring breakdown
- **Spike** - Research and exploration with learning focus
- **Task** - General work items with flexible structure

## Creating Tasks

### Basic Task Creation

```bash
# Create a basic task
zen task create PROJ-123 --title "Implement user authentication"

# Specify task type
zen task create BUG-456 --type bug --title "Fix memory leak in auth service"

# Add metadata
zen task create FEAT-789 \
  --title "Dashboard redesign" \
  --type story \
  --owner "jane.doe" \
  --team "frontend" \
  --priority P1
```

### Task from External Systems

```bash
# Create from Jira issue
zen task create PROJ-123 --from jira

# Create from GitHub issue (planned)
zen task create GH-456 --from github

# Create from Linear issue (planned)
zen task create LIN-789 --from linear

# Create local-only task
zen task create LOCAL-123 --from local
```

### Interactive Task Creation

```bash
# Interactive mode (prompts for details)
zen task create PROJ-123

# Zen will prompt for:
# - Task title
# - Task type (story, bug, epic, spike, task)
# - Priority (P0, P1, P2, P3)
# - Owner (optional)
# - Team (optional)
```

## Task Structure

### Directory Layout

When you create a task, Zen generates this structure:

```
.zen/tasks/PROJ-123/
├── index.md              # Human-readable task overview
├── manifest.yaml         # Machine-readable metadata
├── .taskrc.yaml         # Task-specific configuration
├── metadata/            # Integration metadata
│   ├── jira.json       # Jira sync data (if applicable)
│   └── github.json     # GitHub sync data (if applicable)
├── research/           # Stage 1-2: Align & Discover
│   ├── requirements.md
│   ├── stakeholders.md
│   └── assumptions.md
├── spikes/             # Stage 2-3: Discover & Prioritize
│   ├── technical-spike.md
│   └── user-research.md
├── design/             # Stage 4: Design
│   ├── architecture.md
│   ├── api-spec.md
│   └── mockups/
├── execution/          # Stage 5: Build
│   ├── implementation-plan.md
│   ├── code-review.md
│   └── testing.md
└── outcomes/           # Stage 6-7: Ship & Learn
    ├── deployment.md
    ├── metrics.md
    └── retrospective.md
```

### Task Files

#### index.md - Task Overview

```markdown
# Implement User Authentication

**Task ID**: PROJ-123
**Type**: Story
**Priority**: P1
**Owner**: jane.doe
**Team**: backend
**Status**: In Progress

## Description

Implement OAuth2 authentication system with support for GitHub and Google providers.

## Acceptance Criteria

- [ ] Users can sign in with GitHub
- [ ] Users can sign in with Google
- [ ] JWT tokens are properly managed
- [ ] User sessions persist across browser restarts
- [ ] Proper error handling for auth failures

## Zenflow Progress

- [x] **Align**: Success criteria defined
- [x] **Discover**: Requirements gathered
- [ ] **Prioritize**: Value assessment complete
- [ ] **Design**: Technical specification ready
- [ ] **Build**: Implementation complete
- [ ] **Ship**: Deployed to production
- [ ] **Learn**: Metrics collected and analyzed

## Links

- Jira Issue: [PROJ-123](https://company.atlassian.net/browse/PROJ-123)
- Design Doc: [design/architecture.md](design/architecture.md)
- Implementation: [execution/implementation-plan.md](execution/implementation-plan.md)
```

#### manifest.yaml - Machine-Readable Metadata

```yaml
# Task metadata for automation and tooling
task:
  id: "PROJ-123"
  title: "Implement user authentication"
  type: "story"
  priority: "P1"
  status: "in_progress"
  owner: "jane.doe"
  team: "backend"
  created: "2025-10-26T10:00:00Z"
  updated: "2025-10-26T14:30:00Z"

# Zenflow stage tracking
zenflow:
  current_stage: "discover"
  stages:
    align:
      status: "completed"
      completed_at: "2025-10-26T11:00:00Z"
    discover:
      status: "in_progress"
      started_at: "2025-10-26T11:30:00Z"
    prioritize:
      status: "pending"
    design:
      status: "pending"
    build:
      status: "pending"
    ship:
      status: "pending"
    learn:
      status: "pending"

# External system integration
integration:
  jira:
    external_id: "PROJ-123"
    last_sync: "2025-10-26T14:30:00Z"
    sync_enabled: true
  github:
    pull_request: null
    branch: "feature/PROJ-123"
    sync_enabled: false

# Task relationships
relationships:
  parent: null
  children: []
  blocks: []
  blocked_by: []
  related: ["PROJ-124", "PROJ-125"]

# Labels and categorization
labels:
  - "authentication"
  - "oauth2"
  - "security"
  
components:
  - "auth-service"
  - "user-api"
  
# Effort estimation
estimation:
  story_points: 8
  hours_estimated: 32
  hours_actual: null
  complexity: "medium"
```

#### .taskrc.yaml - Task Configuration

```yaml
# Task-specific configuration
task:
  id: "PROJ-123"
  workspace_root: "../../../.."
  
# Template configuration
templates:
  enabled: true
  auto_generate: true
  custom_templates:
    - "auth-service-template"
    - "api-documentation-template"

# Integration settings
integration:
  jira:
    sync_enabled: true
    sync_frequency: "15m"
    field_mapping:
      status: "status"
      priority: "priority"
      assignee: "owner"
  
  github:
    sync_enabled: false
    auto_create_pr: false
    branch_naming: "feature/{{.task_id}}"

# Workflow automation
automation:
  stage_transitions:
    auto_advance: false
    require_approval: true
  
  notifications:
    slack_channel: "#backend-team"
    email_updates: true
  
  quality_gates:
    code_review_required: true
    tests_required: true
    documentation_required: true

# Custom fields
custom:
  business_value: "high"
  technical_risk: "medium"
  user_impact: "high"
```

## Task Lifecycle

### Stage Progression

Tasks progress through Zenflow stages with clear gates and deliverables:

#### 1. Align Stage

```bash
# Create task and define success criteria
zen task create PROJ-123 --title "User Authentication"

# Generated files:
# - research/requirements.md
# - research/stakeholders.md
# - research/success-criteria.md
```

**Deliverables:**
- Clear problem statement
- Success criteria defined
- Stakeholder alignment documented
- Business value articulated

#### 2. Discover Stage

```bash
# Gather evidence and validate assumptions
# Files automatically created:
# - research/user-research.md
# - research/technical-constraints.md
# - spikes/feasibility-study.md
```

**Deliverables:**
- User research completed
- Technical constraints identified
- Assumptions validated
- Risk assessment documented

#### 3. Prioritize Stage

```bash
# Rank work by value vs effort
# Files automatically created:
# - spikes/value-assessment.md
# - spikes/effort-estimation.md
```

**Deliverables:**
- Value proposition quantified
- Effort estimation completed
- Priority ranking justified
- Resource allocation planned

#### 4. Design Stage

```bash
# Specify implementation approach
# Files automatically created:
# - design/architecture.md
# - design/api-specification.md
# - design/user-interface.md
```

**Deliverables:**
- Technical architecture defined
- API specifications documented
- User interface designed
- Implementation plan created

#### 5. Build Stage

```bash
# Deliver working software increment
# Files automatically created:
# - execution/implementation-plan.md
# - execution/testing-strategy.md
# - execution/code-review.md
```

**Deliverables:**
- Working software increment
- Comprehensive test coverage
- Code review completed
- Documentation updated

#### 6. Ship Stage

```bash
# Deploy safely to production
# Files automatically created:
# - outcomes/deployment-plan.md
# - outcomes/rollback-strategy.md
# - outcomes/monitoring.md
```

**Deliverables:**
- Production deployment completed
- Monitoring and alerting configured
- Rollback strategy tested
- Performance metrics baseline

#### 7. Learn Stage

```bash
# Measure outcomes and iterate
# Files automatically created:
# - outcomes/metrics-analysis.md
# - outcomes/user-feedback.md
# - outcomes/retrospective.md
```

**Deliverables:**
- Success metrics measured
- User feedback collected
- Lessons learned documented
- Next iteration planned

### Stage Transitions

```bash
# Check current stage
zen task status PROJ-123

# Advance to next stage (with validation)
zen task advance PROJ-123

# Jump to specific stage (if allowed)
zen task stage PROJ-123 design

# Rollback to previous stage
zen task rollback PROJ-123
```

## Task Management Commands

### Listing Tasks

```bash
# List all tasks
zen task list

# Filter by status
zen task list --status in_progress
zen task list --status completed

# Filter by type
zen task list --type story
zen task list --type bug

# Filter by owner
zen task list --owner jane.doe

# Filter by team
zen task list --team backend

# Complex filtering
zen task list --type story --priority P1 --status in_progress
```

### Task Information

```bash
# Show task details
zen task show PROJ-123

# Show task status
zen task status PROJ-123

# Show task history
zen task history PROJ-123

# Show task relationships
zen task relationships PROJ-123
```

### Task Updates

```bash
# Update task properties
zen task update PROJ-123 --status in_progress
zen task update PROJ-123 --priority P1
zen task update PROJ-123 --owner john.doe

# Add labels
zen task label PROJ-123 add authentication security

# Remove labels
zen task label PROJ-123 remove legacy

# Update description
zen task update PROJ-123 --description "Updated requirements based on user feedback"
```

### Task Relationships

```bash
# Link related tasks
zen task link PROJ-123 PROJ-124 --relation related

# Create parent-child relationship
zen task link PROJ-100 PROJ-123 --relation parent

# Create blocking relationship
zen task link PROJ-123 PROJ-125 --relation blocks

# Remove relationship
zen task unlink PROJ-123 PROJ-124
```

## Task Templates

### Using Templates

```bash
# Create task with specific template
zen task create PROJ-123 --template story-template

# List available templates
zen task templates list

# Show template details
zen task templates show story-template
```

### Custom Templates

Create custom templates in `.zen/templates/task/`:

```yaml
# .zen/templates/task/api-story.yaml
name: "API Story Template"
description: "Template for API development stories"
type: "story"

files:
  - path: "design/api-specification.md"
    template: |
      # API Specification: {{.title}}
      
      ## Endpoints
      
      ### {{.endpoint_method}} {{.endpoint_path}}
      
      **Description**: {{.endpoint_description}}
      
      **Request**:
      ```json
      {{.request_example}}
      ```
      
      **Response**:
      ```json
      {{.response_example}}
      ```
  
  - path: "execution/api-tests.md"
    template: |
      # API Tests: {{.title}}
      
      ## Test Cases
      
      - [ ] Happy path test
      - [ ] Error handling test
      - [ ] Authentication test
      - [ ] Validation test

variables:
  - name: "endpoint_method"
    type: "string"
    required: true
    allowed_values: ["GET", "POST", "PUT", "DELETE", "PATCH"]
  - name: "endpoint_path"
    type: "string"
    required: true
  - name: "endpoint_description"
    type: "string"
    required: true
  - name: "request_example"
    type: "string"
    required: false
    default: "{}"
  - name: "response_example"
    type: "string"
    required: false
    default: "{}"
```

## Integration with External Systems

### Jira Integration

```bash
# Create task from Jira issue
zen task create PROJ-123 --from jira

# Sync task with Jira
zen task sync PROJ-123

# Push local changes to Jira
zen task push PROJ-123

# Pull updates from Jira
zen task pull PROJ-123
```

### GitHub Integration (Planned)

```bash
# Create task from GitHub issue
zen task create GH-456 --from github

# Create pull request for task
zen task pr PROJ-123 --title "Implement user authentication"

# Link task to existing PR
zen task link PROJ-123 --pr 123
```

## Automation and Workflows

### Quality Gates

Configure automatic quality gates:

```yaml
# .taskrc.yaml
automation:
  quality_gates:
    align:
      - requirements_documented: true
      - stakeholders_identified: true
    
    discover:
      - user_research_completed: true
      - technical_constraints_identified: true
    
    design:
      - architecture_reviewed: true
      - api_specification_complete: true
    
    build:
      - code_review_approved: true
      - tests_passing: true
      - documentation_updated: true
    
    ship:
      - deployment_successful: true
      - monitoring_configured: true
    
    learn:
      - metrics_collected: true
      - feedback_analyzed: true
```

### Notifications

```yaml
# .taskrc.yaml
automation:
  notifications:
    stage_transitions:
      slack:
        channel: "#team-backend"
        message: "Task {{.task_id}} advanced to {{.stage}}"
    
    quality_gate_failures:
      email:
        recipients: ["{{.owner}}", "team-lead@company.com"]
        subject: "Quality gate failed for {{.task_id}}"
```

## Reporting and Analytics

### Task Metrics

```bash
# Show task metrics
zen task metrics

# Example output:
# Task Metrics
# =============
# Total tasks: 45
# Completed: 23 (51%)
# In progress: 15 (33%)
# Blocked: 3 (7%)
# Not started: 4 (9%)
# 
# Average cycle time: 8.5 days
# Average lead time: 12.3 days
# 
# By Type:
# - Stories: 28 (62%)
# - Bugs: 12 (27%)
# - Spikes: 5 (11%)
```

### Stage Analysis

```bash
# Analyze stage progression
zen task stages

# Example output:
# Stage Analysis
# ==============
# Align → Discover: 2.1 days avg
# Discover → Prioritize: 1.3 days avg
# Prioritize → Design: 0.8 days avg
# Design → Build: 4.2 days avg
# Build → Ship: 1.9 days avg
# Ship → Learn: 3.1 days avg
# 
# Bottlenecks:
# 1. Design → Build (4.2 days)
# 2. Align → Discover (2.1 days)
```

### Team Performance

```bash
# Team performance metrics
zen task team-metrics --team backend

# Individual performance
zen task user-metrics --user jane.doe
```

## Best Practices

### Task Creation

1. **Use descriptive titles** - Make the task purpose immediately clear
2. **Choose appropriate types** - Select the type that best fits the work
3. **Set realistic priorities** - Use P0-P3 system consistently
4. **Define acceptance criteria** - Be specific about what "done" means
5. **Link related tasks** - Show dependencies and relationships

### Task Management

1. **Regular updates** - Keep task status current
2. **Document decisions** - Record important choices and rationale
3. **Use templates** - Leverage templates for consistency
4. **Stage progression** - Don't skip stages without good reason
5. **Quality gates** - Ensure deliverables meet standards before advancing

### Team Collaboration

1. **Clear ownership** - Assign tasks to specific individuals
2. **Team visibility** - Use team filters and dashboards
3. **Communication** - Update stakeholders on progress
4. **Knowledge sharing** - Document lessons learned
5. **Retrospectives** - Regular team retrospectives on task management

## Troubleshooting

### Common Issues

#### Task Creation Fails

```bash
# Check workspace initialization
zen status

# Verify configuration
zen config list | grep work

# Check permissions
ls -la .zen/tasks/
```

#### Sync Problems

```bash
# Check integration status
zen auth status

# Manual sync
zen task sync PROJ-123 --force

# Clear sync metadata
rm .zen/tasks/PROJ-123/metadata/jira.json
```

#### Template Issues

```bash
# List available templates
zen task templates list

# Validate template
zen task templates validate custom-template

# Reset to default templates
zen task templates reset
```

## Migration Guide

### From Other Task Management Tools

```bash
# Export existing tasks
zen task export --format json > tasks-backup.json

# Import from external system
zen task import --from jira --project PROJ
zen task import --from github --repo user/repo
zen task import --file tasks-backup.json
```

### Upgrading Task Structure

```bash
# Upgrade task structure to latest version
zen task upgrade PROJ-123

# Upgrade all tasks
zen task upgrade --all

# Backup before upgrade
zen task backup --all
```

## See Also

- **[Zenflow Guide](../zenflow/README.md)** - Complete Zenflow methodology
- **[Integration Guide](integrations.md)** - External system integration
- **[Template Guide](templates.md)** - Template system documentation
- **[Task API](../api/task.md)** - Task management API reference
