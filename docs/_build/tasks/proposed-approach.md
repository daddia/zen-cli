# Proposed Approach

## **Key Strengths of This Approach**

### 1. **Clear Separation of Concerns**
Instead of cramming everything into one massive story template, you now have:
- **`index.md`** - Clean, focused story (like the Atlassian AI-simplified version)
- **Stage-specific folders** - Technical details where they belong
- **Curated vs. Raw data** - Human-readable vs. system dumps

### 2. **Perfect Workflow Alignment**
This structure maps beautifully to your 12 agentic stages:

| Workflow Stage | Task Folder | Agent Responsibility |
|----------------|-------------|---------------------|
| 01-Discover | `index.md` + `manifest.yaml` | Story definition & metadata |
| 02-Prioritise | `manifest.yaml` (priority, risk) | WSJF/RICE scoring |
| 03-Design | `design/` + `specs/` | Contracts & design artifacts |
| 04-Architect | `docs/adr/` + `security/threat-model.md` | Architecture decisions |
| 05-Plan | `code/` (spikes) + `scripts/` | Planning & scaffolding |
| 06-Build | `code/` → actual repos | Development work |
| 07-Code-Review | `reviews/` | Code review evidence |
| 08-QA | `qa/` + `perf/` | Test plans & results |
| 09-Security | `security/` | Threat models & scans |
| 10-Release | `release/` | Feature flags & rollout |
| 11-Post-Deploy | `ops/` | Operational handover |
| 12-Roadmap-Feedback | `artifacts/` | Outcomes & learnings |

### 3. **Automation-Ready**
The structure supports your agentic workflow with:
- **Machine-readable metadata** (`manifest.yaml`)
- **Sync scripts** for external systems  
- **Validation in CI** for quality gates
- **Template-driven** artifact generation

## **Recommendations for Integration**

### 1. **Enhance Agent Integration**
Add agent-specific markers to the structure:

```yaml
# manifest.yaml additions
workflow:
  current_stage: "03-Design"
  completed_stages: ["01-Discover", "02-Prioritise"]
  agent_outputs:
    "01-Discover": 
      - "index.md"
      - "docs/meeting-notes/2025-09-13-kickoff.md"
    "03-Design":
      - "specs/storefront.cart.v1.openapi.yaml" 
      - "design/links.md"
```

### 2. **Stage-Specific Validation Rules**
Create `.taskrc.yaml` extensions for workflow stages:

```yaml
# .taskrc.yaml
workflow:
  stage_03_design:
    required_artifacts: ["specs/*.yaml", "design/links.md"]
    validation: ["spectral", "figma_link_check"]
  stage_08_qa:
    required_artifacts: ["qa/test-plan.md", "qa/results/"]
    validation: ["test_coverage_check"]
```

### 3. **Template Standardization**
Your existing templates should become the defaults:

[1 tool called]

Perfect! Your existing templates map to the new structure:

```bash
# Current → New Structure Mapping
docs/templates/feature-brief.md → work/tasks/_templates/index.md
docs/templates/adr.md → work/tasks/_templates/docs/adr/ADR-XXXX.md
docs/templates/design-doc.md → work/tasks/_templates/specs/README.md
docs/templates/review-checklist.md → work/tasks/_templates/reviews/review-checklist.md
```

### 4. **DoR/DoD Integration**
Make DoR/DoD stage-aware:

```yaml
# manifest.yaml
quality_gates:
  stage_01_discover:
    dor: ["story_defined", "acceptance_criteria", "success_metrics"]
    dod: ["stakeholder_signoff", "manifest_complete"]
  stage_03_design:
    dor: ["contracts_drafted", "figma_linked"]  
    dod: ["openapi_valid", "design_approved"]
  # ... etc
```

## **Migration Strategy**

### 1. **Immediate (Current Story)**
- Convert `store-9.md` to use the simple `index.md` format
- Create `work/tasks/STORE-9/` folder with current artifacts
- Use this as your pilot for the new structure

### 2. **Template Creation**
- Generate scaffold templates under `work/tasks/_templates/`
- Create the automation scripts (`scripts/sync-jira.sh`, etc.)
- Set up CI validation for the task structure

### 3. **Agent Prompt Updates** 
- Update your 12-stage agent prompts to reference the task folder structure
- Each agent should know which folders it reads from and writes to
- Create handoff protocols between stages

## **Specific Improvements to Proposed Structure**

### 1. **Add Agent Tracking**
```bash
work/tasks/STORE-9/
├─ .workflow/                    # Agent state and handoffs
│  ├─ stage-history.yaml        # Which agents ran when
│  ├─ handoffs/                 # Agent-to-agent communication
│  └─ validation-results.yaml   # Quality gate results
```

### 2. **Better Integration Hooks**
```yaml
# .taskrc.yaml additions
integrations:
  jira:
    auto_sync: true
    status_mapping:
      "In Progress": "03-Design"
      "Code Review": "07-Code-Review"
  github:
    pr_template: "reviews/pr-template.md"
  figma:
    auto_export: ["design/figma-exports/"]
```

## **Bottom Line**

This proposed structure is **vastly superior** to the complex story template approach. It:

✅ **Solves the original problem** - No more bloated stories  
✅ **Enables your agentic workflow** - Each stage has clear inputs/outputs  
✅ **Supports automation** - Machine-readable metadata + validation  
✅ **Scales with complexity** - Rich artifact management without story bloat  
✅ **Maintains traceability** - Links between Jira, GitHub, Figma, etc.

I strongly recommend adopting this structure. It's exactly what you need to make your 12-stage agentic workflow truly effective while keeping individual stories focused and actionable.
