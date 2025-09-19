# TODO - Zen CLI Golden Path

_Last updated: 20 September, 2025_

## Mission: Prove Zen's Value Through Direct Usage

Build a **v1 Golden Path** that demonstrates Zen's core value proposition: AI-powered product development workflow automation. Use Zen to build Zen itself, iterating rapidly based on real usage feedback.

## Conventions

- **Story ID:** ZEN-### (sequential numbering)
- **Status:** Not started · In progress · Blocked · In review · Done
- **Priority:** P0 (MVP Critical) · P1 (High) · P2 (Medium) · P3 (Low)
- **Estimates:** S/M/L/XL (T-shirt sizes)
- **DoD:** Working in Zen project · Tests pass · Docs updated · Released

---

## Phase 1: Golden Path Foundation

**Objective**: Build the minimum viable Golden Path to take a task from idea to shipped artifact using Zen on the Zen project itself.

---

## In Progress

### **Template Engine & Task Creation** → Target Release v0.3.0

- [ ] **[ZEN-009] Template Engine Core**
  - **Deliverable**: Go template engine with custom functions integrated with Asset Client
  - **Acceptance**: Can render templates with variables, conditionals, and Zen-specific extensions

- [ ] **[ZEN-010] Task Creation Command**
  - **Deliverable**: `zen task create --type [story|bug|epic]` with template-driven structure
  - **Acceptance**: Creates proper `.zen/work/tasks/{TASK-ID}/` structure with index.md and manifest.yaml
  - **Priority**: P0 (Critical for Golden Path)

- [ ] **[ZEN-011] AI Client Foundation**
  - **Deliverable**: Multi-provider LLM client (OpenAI, Anthropic, Azure) with strategy pattern
  - **Acceptance**: Unified interface for content enhancement, cost tracking, and provider switching
  - **Priority**: P0 (Core AI capability)

- [ ] **[ZEN-012] Content Generation Commands**
  - **Deliverable**: `zen write <template>` and `zen draft <template>` for content creation
  - **Acceptance**: Generate content from templates with AI enhancement and variable prompting
  - **Priority**: P1 (User workflow efficiency)

- [ ] **[ZEN-013] Template Viewing Command**
  - **Deliverable**: `zen template view <template>` renders template as plain markdown
  - **Acceptance**: Shows template structure and variables for user reference
  - **Priority**: P2 (User experience enhancement)

### **Work Management & Workflow** → Target Release v0.4.0

- [ ] **[ZEN-014] Task Structure Implementation**
  - **Deliverable**: Complete task directory structure with manifest.yaml
  - **Acceptance**: Tasks have workflow stages, metadata, and artifact directories

- [ ] **[ZEN-015] Task Management Commands**
  - **Deliverable**: `zen task list`, `zen task status`, `zen task show`
  - **Acceptance**: Can view and manage existing tasks

- [ ] **[ZEN-016] Workflow Stage Commands**
  - **Deliverable**: `zen align`, `zen discover`, `zen prioritize`, etc.
  - **Acceptance**: Can progress tasks through workflow stages

- [ ] **[ZEN-017] Stage Validation**
  - **Deliverable**: Basic quality gates for stage progression
  - **Acceptance**: Validates required artifacts before stage completion
  - **Priority**: P1

### **Golden Path Completion** → Target Release v0.5.0

- [ ] **[ZEN-018] Design Stage Templates**
  - **Deliverable**: Templates for API specs, technical designs, architecture
  - **Acceptance**: Can generate design artifacts with AI assistance

- [ ] **[ZEN-019] Build Stage Integration**
  - **Deliverable**: PR templates, commit messages, code scaffolding
  - **Acceptance**: Integrates with development workflow

- [ ] **[ZEN-020] Documentation Generation**
  - **Deliverable**: Auto-generate README, CHANGELOG from task artifacts
  - **Acceptance**: Documentation stays in sync with task progress
  - **Priority**: P1

- [ ] **[ZEN-021] End-to-End Validation**
  - **Deliverable**: Complete Golden Path validation
  - **Acceptance**: Ship a Zen feature built entirely with Zen

---

## Phase 2: Refinement & Expansion

**Objective**: Improve Golden Path based on usage, add essential integrations.

### **Workflow Enhancement** → Target Release v0.6.0

- [ ] **[ZEN-021] Context Awareness**
  - **Deliverable**: Commands understand current task and stage context
  - **Acceptance**: Reduced need for explicit task/stage specification
  - **Priority**: P1

- [ ] **[ZEN-022] Template Library Expansion**
  - **Deliverable**: Comprehensive templates based on real usage
  - **Acceptance**: Templates cover 80% of common scenarios
  - **Priority**: P1

- [ ] **[ZEN-023] AI Prompt Optimization**
  - **Deliverable**: Refined prompts for better AI assistance
  - **Acceptance**: Less manual editing of AI-generated content
  - **Priority**: P1

### **External Integration** → Target Release v0.7.0

- [ ] **[ZEN-024] GitHub Integration**
  - **Deliverable**: Link tasks with issues, PRs, commits
  - **Acceptance**: Bidirectional sync with GitHub
  - **Priority**: P1

- [ ] **[ZEN-025] Jira Integration**
  - **Deliverable**: Optional Jira synchronization
  - **Acceptance**: Tasks can sync with external project tracking
  - **Priority**: P2

- [ ] **[ZEN-026] Documentation Publishing**
  - **Deliverable**: Publish to Confluence, GitHub Wiki, etc.
  - **Acceptance**: Automated documentation distribution
  - **Priority**: P2

---

## Success Criteria

### Phase 1: Golden Path Proven
- [ ] Create task with `zen task create --type story`
- [ ] Progress through all 7 workflow stages
- [ ] Generate technical artifacts (ADRs, designs, docs)
- [ ] Ship complete Zen feature developed with Zen
- [ ] Measurable time-to-ship reduction

### Phase 2: Workflow Refined  
- [ ] 80% of new Zen features use Golden Path
- [ ] User satisfaction >4.0/5
- [ ] Template library covers common scenarios
- [ ] External integrations work seamlessly

## Key Principles

1. **Task-Centric**: Everything organized around tasks with types (story, bug, epic)
2. **Workflow-Driven**: All tasks follow the 7-stage Zen workflow
3. **Template-Powered**: Templates + AI for rapid artifact generation
4. **Dogfooding**: Every feature used to build Zen itself
5. **Value-First**: Each feature must improve productivity
6. **Real-World**: No toy examples - must work for actual development

---

## Complete

- [x] **[ZEN-001] Go Project Setup** - CLI framework with Cobra
- [x] **[ZEN-002] Core CLI Framework** - Root command, version, help
- [x] **[ZEN-003] Workspace Detection & Init** - `zen init` command
- [x] **[ZEN-004] Configuration Management** - Config loading system
- [x] **[ZEN-005] Testing & CI/CD** - GitHub Actions, releases, cross-platform builds

- [x] **[ZEN-006] Private Asset Repository** - Private `zen-assets` repo with manifest.yaml
- [x] **[ZEN-007] Git-based Asset Client** - Asset client with GitHub auth
- [x] **[ZEN-008] Asset Commands** - `zen assets auth`, `zen assets sync`, `zen assets list`
