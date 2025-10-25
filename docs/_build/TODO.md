# TODO - Zen CLI

_Last updated: 25 October, 2025_

## Mission: Strategy-to-Value in 90 Days

Build the **Foundation & Validation** phase that proves Zen's core value proposition: unified AI-powered control plane for product & engineering. Focus on achieving 3 pilot teams with sub-week time-to-first-value and 85%+ workflow completion rates within 30 days.

## Conventions

- **Story ID:** ZEN-### (sequential numbering)
- **Status:** Not started · In progress · Blocked · In review · Done
- **Priority:** P0 (Critical) · P1 (High) · P2 (Medium) · P3 (Low) · P4 (Lowest)
- **Estimates:** Story points (Fibonacci)
- **DoD:** Working in Zen project · Tests pass · Docs updated · Released

---

## NOW (Days 1-30): Foundation & Validation

**Objective**: Pilot teams achieve sub-week time-to-first-value with core workflows

**Success Metrics:**
- 3 pilot teams consistently activated weekly
- Average TTFV < 3 days (target: < 1 day)  
- 85%+ workflow completion rate (target: 90%+)
- Positive feedback and expansion requests

**Key Deliverables:**
- ✓ Core CLI with workspace initialization
- ✓ Opinionated workflow templates (7-stage Zenflow)
- ✓ Git/GitHub integration for version control
- ✓ Basic quality validation and error handling

---

## Critical Path Items (P0)

### **CLI Foundation** → Target: Week 1

- [x] **[ZEN-011] Template Generation Command**
  - **Deliverable**: `zen draft <activity>` renders template as plain format
  - **Acceptance**: Generate document template for 7-stage Zenflow
  - **Status**: Done - Template generation working with zen-assets integration
  - **Strategic Impact**: Enables rapid workflow template creation

- [ ] **[ZEN-012] Workspace Templates**  
  - **Deliverable**: Opinionated workspace initialization with 7-stage Zenflow templates
  - **Acceptance**: `zen init` creates structured workspace with quality gates
  - **Status**: Not started - Requires template system enhancement
  - **Strategic Impact**: Reduces TTFV to < 1 day target

- [ ] **[ZEN-013] Git Integration**
  - **Deliverable**: Version control integration for workspace and workflow tracking
  - **Acceptance**: 100% version control workflows with automated commit patterns
  - **Status**: Not started - Core CLI dependency
  - **Strategic Impact**: Enables workflow completion tracking

### **Quality Gates** → Target: Week 2

- [ ] **[ZEN-014] Basic Quality Gates**
  - **Deliverable**: Validation framework for workflow progression
  - **Acceptance**: 85%+ pass rate on quality checks (target: 90%+)
  - **Status**: Not started - Requires validation framework
  - **Strategic Impact**: Ensures workflow completion and quality standards

### **Pilot Validation** → Target: Week 3-4

- [ ] **[ZEN-015] Pilot Team Onboarding**
  - **Deliverable**: 3 pilot teams successfully onboarded with Zen workflows
  - **Acceptance**: Teams complete first workflow cycle within 3 days
  - **Status**: Not started - Requires CLI foundation completion
  - **Strategic Impact**: Validates core value proposition and TTFV targets

- [ ] **[ZEN-016] Workflow Completion Tracking**
  - **Deliverable**: Basic metrics collection for workflow completion rates
  - **Acceptance**: Track initiation vs. completion events for pilot teams
  - **Status**: Not started - Requires instrumentation
  - **Strategic Impact**: Measures progress toward 85%+ completion rate target

- [ ] **[ZEN-017] Feedback Collection System**
  - **Deliverable**: Structured feedback collection from pilot teams
  - **Acceptance**: Weekly feedback sessions with clear improvement priorities
  - **Status**: Not started - Requires pilot team engagement
  - **Strategic Impact**: Drives rapid iteration and product-market fit validation

---

## NEXT Phase (Days 31-60): Integration & Intelligence
*Moved to future roadmap - focus on Foundation & Validation first*

**Objective**: Teams eliminate tool fragmentation with AI-powered orchestration

**Key Initiatives (Future):**
- Tool Integrations (Jira, Slack, Teams)
- AI Orchestration (Multi-provider LLM)
- Workflow Automation (90% completion rate)
- Metric Collection (Real-time dashboard)

---

## LATER Phase (Days 61-90): Optimization & Scale  
*Moved to future roadmap - focus on Foundation & Validation first*

**Objective**: Demonstrate measurable productivity gains and scale readiness

**Key Initiatives (Future):**
- Advanced Quality Gates (95%+ pass rate)
- Performance Optimization (< 200ms p95 latency)
- Team Expansion (15 Weekly Activated Teams)
- Enterprise Features (RBAC + audit ready)

---

## Success Criteria

### 30-Day Milestone: "Foundation & Validation"
**Primary Outcome:** Prove core value proposition with pilot teams

**Success Metrics:**
- [x] Create task with `zen task create --type story`
- [x] Create task with Jira integration `zen task create --from jira`
- [ ] **3 pilot teams consistently activated weekly**
- [ ] **Average TTFV < 3 days** (target: < 1 day)
- [ ] **85%+ workflow completion rate** (target: 90%+)
- [ ] **Positive feedback and expansion requests**

**Risk Mitigation:**
- Daily check-ins with pilot teams
- Rapid iteration on onboarding friction
- Clear escalation path for blockers

### North Star Alignment
**Weekly Activated Teams Progress:**
- Day 30 Target: 3 teams (pilot validation)
- Day 60 Target: 8 teams (product-market fit signal)  
- Day 90 Target: 15 teams (scale readiness)

## Key Principles

1. **Strategy-to-Value Focus**: Every feature must advance our 90-day roadmap outcomes
2. **Pilot-Driven Development**: Build for real pilot teams, not hypothetical users
3. **TTFV Obsession**: Optimize relentlessly for sub-week time-to-first-value
4. **Quality Gates First**: Build validation into every workflow step
5. **Metrics-Driven**: Measure completion rates, pass rates, and team activation
6. **Foundation Before Features**: Nail core workflows before expanding capabilities

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

- [x] **[ZEN-009] Template Engine Core** - Go template engine with custom functions
- [x] **[ZEN-010] Task Creation Command** - `zen task create` with template-driven structure.

- [x] **[ZEN-025] Integration Services Layer** Plugin-based external integration architecture
- [x] **[ZEN-026] Task Integration** Task synchronization foundation
- [x] **[ZEN-027] Task Creation with Jira Integration** Complete Jira task creation workflow
- [x] **[ZEN-028] Code Quality & Build System** Refactoring analysis and build system fixes
