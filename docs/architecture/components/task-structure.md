# Zenflow Task Structure Specification

## Overview

This specification defines how tasks and stories are organized within the Zenflow workflow to maximize Agile delivery while maintaining clear artifact management. The structure organizes work by **type** rather than **temporal sequence**, enabling flexible, iterative development while supporting the seven Zenflow stages.

For the technical implementation details, APIs, and component architecture that manages this task structure, see the [Work Management Technical Specification](work-management.md).

## Core Principles

### Agile-First Design

- Lightweight documentation that supports working software
- Flexible artifact creation based on value and necessity
- Cross-stage work types that don't constrain iteration
- Just-in-time elaboration of details


### Zenflow Integration

- Clear progression through Align → Discover → Prioritize → Design → Build → Ship → Learn
- Stage-agnostic artifact types that can be created when needed
- Quality gates based on outcomes, not documentation completeness
- Machine-readable workflow state tracking

## Task Structure

Each task is organized in a logical work-type hierarchy:

```sh
.zen/tasks/PROJ-123/
├─ .zenflow/                  # Zenflow state and progression tracking
├─ index.md                   # Story brief (required)
├─ manifest.yaml              # Machine-readable metadata
├─ .taskrc.yaml              # Task-specific configuration
├─ research/                  # Investigation and discovery work
├─ spikes/                    # Technical exploration and prototyping  
├─ design/                    # Specifications and planning artifacts
├─ execution/                 # Implementation evidence and results
├─ outcomes/                  # Learning, metrics, and retrospectives
└─ metadata/                  # External system snapshots
```

## Work Type Definitions


### Task/Story Brief (index.md)
The foundational document that remains lightweight and user-focused throughout the task lifecycle. Contains user story, acceptance criteria, success metrics, and scope boundaries. Required for all tasks.

### Machine-readable Metadata (manifest.yaml)
Rich YAML file containing detailed workflow progression, quality gates, integrations, and automation metadata. Drives workflow automation and external system synchronization. Required for all tasks.

### Task Configuration (.taskrc.yaml)
Task-specific configuration controlling automation behavior, quality standards, team settings, and integration preferences. Required for all tasks.

### Research Directory
Contains all investigative work including user research, requirements gathering, competitive analysis, and assumption validation. Created during Discover stage but can be updated throughout the task lifecycle as new insights emerge.

### Spikes Directory
Houses technical exploration, proof-of-concepts, feasibility studies, and prototyping work. Can be created at any Zenflow stage when technical uncertainty needs resolution. Emphasizes learning over perfect documentation.

### Design Directory  
Holds technical specifications, architecture decisions (ADRs), API contracts, and implementation planning. Created primarily during Design stage but can evolve during Build as implementation details emerge.

### Execution Directory
Contains implementation evidence including code reviews, test results, deployment logs, and monitoring setup. Populated during Build and Ship stages to provide audit trail of delivery quality.

### Outcomes Directory
Captures success measurements, user feedback, retrospectives, and learnings for future iterations. Primarily populated during Learn stage but can be updated throughout as insights emerge.

### Metadata Directory
Contains raw snapshots from external systems (Jira, GitHub, Figma, Confluence) for integration and synchronization purposes. Automatically managed by external sync processes.

## Zenflow Stage Mapping

The work types support all seven Zenflow stages without constraining when artifacts are created:

| Stage | Align | Discover | Prioritize | Design | Build | Ship | Learn |
|-------|-------|----------|------------|--------|-------|------|-------|
| **Goal** | Define success criteria and stakeholder alignment | Gather evidence and validate assumptions | Rank work by value vs effort | Specify implementation approach | Deliver working software increment | Deploy safely to production | Measure outcomes and iterate |
| **Inputs** | Business requirements, stakeholder needs | **Story brief**, user hypotheses | **Research findings**, spike results | **Priority decisions**, technical constraints | **Technical specifications**, implementation plan | **Working software**, deployment strategy | **Production deployment**, success criteria |
| **Key Activities** | Story definition, success metrics, stakeholder alignment | User research, requirements gathering, feasibility validation | WSJF/RICE scoring, value analysis, effort estimation | Technical planning, architecture decisions, contract definition | Code development, testing, peer review | Deployment execution, monitoring setup, rollback preparation | Metrics collection, user feedback, retrospective analysis |
| **Required Outputs** | **index.md** (story brief) |  |  |  |  |  |  |
| **Primary Outputs** | | user-interviews, requirements | manifest.yaml (priority rationale) | technical-spec, adrs | code-reviews, test-results | deployment-logs, monitoring | metrics, retrospective |
| **Supporting Outputs** | assumptions, initial-exploration | feasibility-study, competitive-analysis | effort-estimates | architecture-exploration, api-contracts | implementation-poc, performance-tests | rollback-plan, deployment-validation | user-feedback, next-iteration |

## Quality Gates

Quality progression focuses on working software and user outcomes rather than documentation completeness:

### Stage Progression Criteria

- Align: Story brief complete with measurable success criteria
- Discover: Sufficient evidence to make informed design decisions
- Prioritize: Clear value proposition and effort understanding
- Design: Implementation approach agreed and technically feasible
- Build: Working software increment meets acceptance criteria
- Ship: Production deployment successful with monitoring active
- Learn: Success metrics captured and learnings documented


### Artifact Requirements

- Minimal viable documentation at each stage
- Optional artifacts created only when they add clear value
- Cross-stage flexibility allowing work type creation as needed
- Evidence-based progression rather than checklist completion

## Template Organization

Templates support just-in-time artifact creation:

```bash
.zen/tasks/_templates/
├─ .zenflow/                  # Zenflow state tracking templates
├─ index.md                   # Story brief template
├─ manifest.yaml              # Machine-readable metadata template
├─ .taskrc.yaml              # Task configuration template
├─ research/                  # Investigation templates
├─ spikes/                    # Exploration templates  
├─ design/                    # Specification templates
├─ execution/                 # Implementation templates
├─ outcomes/                  # Learning templates
└─ metadata/                  # External system integration templates
```

## Integration Points

### External Tool Synchronization

- Jira status mapping to Zenflow stages
- GitHub branch and PR integration with execution artifacts
- Figma design handoff automation with design specifications


### CLI Usage Examples

#### Task Management Commands
- `zen task create [TASK-ID] --type story/task` creates new task with full structure
- `zen task list` shows all tasks in workspace
- `zen task status [TASK-ID]` shows current stage and progress (current task if no ID)
- `zen task progress [TASK-ID]` advances to next Zenflow stage with validation
- `zen task config [TASK-ID]` manages task-specific settings

#### Content Creation Commands (Git-like)
- `zen write requirements` creates requirements document (auto-detects work type: research/)
- `zen write technical-spec` creates technical specification (auto-detects work type: design/)
- `zen write user-interviews` creates user interview document (auto-detects work type: research/)
- `zen write performance-test` creates performance test spike (auto-detects work type: spikes/)
- `zen write retrospective` creates retrospective document (auto-detects work type: outcomes/)

#### Context-Aware Commands
- `zen PROJ-123 write requirements` creates document in specific task
- `zen PROJ-123 status` shows status of specific task
- `zen write api-contract --type design` explicitly specify work type if needed

#### Smart Auto-Detection Examples
- `zen write requirements` → `research/requirements.md`
- `zen write user-research` → `research/user-research.md`
- `zen write competitive-analysis` → `research/competitive-analysis.md`
- `zen write feasibility-study` → `spikes/feasibility-study.md`
- `zen write architecture-spike` → `spikes/architecture-spike.md`
- `zen write technical-spec` → `design/technical-spec.md`
- `zen write api-contracts` → `design/api-contracts.yaml`
- `zen write test-plan` → `execution/test-plan.md`
- `zen write deployment-logs` → `execution/deployment-logs.md`
- `zen write metrics-analysis` → `outcomes/metrics-analysis.md`
- `zen write lessons-learned` → `outcomes/lessons-learned.md`  

### Automation Hooks

- YAML-based quality gate validation before stage progression
- Template instantiation with task context and configuration
- Cross-tool artifact linking and synchronization via metadata/
- Workflow state tracking in manifest.yaml and .zenflow/
- External system sync automation based on .taskrc.yaml settings

## Migration Strategy

### Immediate Implementation

- Convert existing stories to index.md format
- Create task directory structure with manifest.yaml and .taskrc.yaml for active work
- Begin using research/ and spikes/ for current investigations
- Set up metadata/ directories for external system integration

### Template Development

- Create minimal viable templates for each work type including configuration
- Establish CLI commands for task creation with full structure
- Configure YAML-based quality gate validations for stage progression
- Develop manifest.yaml and .taskrc.yaml templates with rich metadata

### Tool Integration

- Configure Jira synchronization with manifest.yaml workflow tracking
- Set up GitHub integration for execution artifacts with metadata/ snapshots
- Establish Figma handoff automation with .taskrc.yaml configuration
- Implement external system sync based on task-specific settings

## Success Metrics

### Agile Delivery Indicators

- Reduced time from story creation to working software
- Increased frequency of user feedback incorporation
- Higher percentage of delivered features meeting success criteria
- Faster identification and resolution of technical risks through spikes

### Workflow Effectiveness

- Clear progression through Zenflow stages without bottlenecks
- Appropriate artifact creation based on value and necessity
- Cross-functional team collaboration through shared work types
- Continuous improvement through outcomes capture and application

---

## Example Comprehensive Task Structure

Example of a comprehsive structure of a task - for reference only:

```sh
.zen/tasks/PROJ-123/
├─ .zenflow/                 # Zenflow state and progression tracking
├─ index.md                   # Story brief (required)
├─ manifest.yaml              # Machine-readable metadata (required)
├─ .taskrc.yaml              # Task-specific configuration (required)
├─ research/                  # Discovery & Investigation
│  ├─ user-research.md        # User interviews, surveys
│  ├─ requirements.md         # Gathered requirements
│  ├─ competitive-analysis.md # Market research
│  └─ assumptions.md          # Hypotheses to validate
├─ spikes/                    # Technical Exploration and Prototyping 
│  ├─ architecture-spike.md   # Technical feasibility
│  ├─ performance-poc.md      # Performance prototypes  
│  ├─ integration-test.md     # 3rd party API tests
│  └─ security-review.md      # Security considerations 
├─ design/                    # Specifications and planning artifacts
│  ├─ technical-spec.md       # Implementation approach
│  ├─ api-contracts.yaml      # API specifications
│  ├─ wireframes/             # UI/UX designs
│  ├─ adr/                    # Architecture decisions
│  └─ implementation-plan.md  # Development roadmap
├─ execution/                 # Implementation Artifacts and Results
│  ├─ code-reviews/           # Review evidence
│  ├─ test-results/           # QA artifacts
│  ├─ deployment-logs/        # Release evidence
│  └─ monitoring/             # Production health
├─ outcomes/                  # Results, Learning, and Retrospectives
│  ├─ metrics.md             #   Success measurements
│  ├─ retrospective.md       #   What we learned
│  ├─ user-feedback.md       #   Post-release insights
│  └─ next-iteration.md      #   Future improvements
└─ metadata/                  # External system snapshots
   ├─ jira.json              #   Jira issue data
   ├─ github.json            #   GitHub PR and commit data
   ├─ figma.json             #   Figma design metadata
   └─ confluence.json        #   Confluence documentation links
```

Note: The above are examples only and will be different for each Task.
