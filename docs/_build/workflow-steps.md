
The workflow follows 12 distinct stages, each with dedicated agents and prompts:

1. **01-Discover** - Discovery Agent analyzes requirements and creates ADRs

2. **02-Prioritise** - Prioritisation Agent ranks work using WSJF/RICE/ICE frameworks
   - [prioritise.md](02-Prioritise/prioritise.md) - Backlog prioritization

3. **03-Design** - Design & Contract Agent creates technical specifications
   - [design.md](03-Design/design.md) - API contract design

4. **04-Architect** - Architecture Review Agent validates designs against NFRs
   - [architect-review.md](04-Architect/architect-review.md) - Architecture validation
   - [architect-adr.md](04-Architect/architect-adr.md) - ADR creation
   - [architect-solution.md](04-Architect/architect-solution.md) - Solution architecture

5. **05-Plan** - Planning & Scaffolding Agent breaks down work and creates scaffolds
   - [plan-scaffold.md](05-Plan/plan-scaffold.md) - Implementation planning

6. **06-Build** - Build stage with human-led development
   - [build.md](06-Build/build.md) - Production-quality implementation

7. **07-Code-Review** - Code Review Agent analyzes quality and compliance
   - [code-review.md](07-Code-Review/code-review.md) - Automated code review

8. **08-QA** - Test & QA Agent orchestrates comprehensive testing
   - [qa-test.md](08-QA/qa-test.md) - Multi-layer testing
   - [qa-solution-review.md](08-QA/qa-solution-review.md) - Architecture review

9. **09-Security** - Security & Compliance Agent scans for vulnerabilities
   - [security-compliance.md](09-Security/security-compliance.md) - Security assessment

10. **10-Release** - Release Manager Agent handles progressive deployments
    - [release-management.md](10-Release/release-management.md) - Release orchestration

11. **11-Post-Deploy** - Post-Deploy Verification Agent monitors production
    - [post-deploy-verification.md](11-Post-Deploy/post-deploy-verification.md) - Production validation

12. **12-Roadmap-Feedback** - Roadmap Feedback Agent analyzes outcomes



Discovery (Agent A) - Stakeholder/context scrape; draft ADR outline; risk/assumption log
Prioritisation (Agent B) - Compute WSJF/ICE from Jira fields; propose ranking
Design & Contracting (Agent C) - Generate contracts first (Protobuf/OpenAPI skeletons)
Architecture Review (Agent F) - NFR budgets, security model, policy checks
Scaffolding & Planning (Agent D) - Create repo/module scaffolds, CI jobs, feature flags
Build (Human-led, agent-assisted) - Short-lived branch development
Code Review (Agent E) - Static analysis, structured findings, human reviewer support
Test & QA (Agent G) - Test pyramid orchestration, coverage checks, flake detection
Security & Compliance (Agent H) - SAST/SCA/DAST, SBOM, policy validation
Release Management (Agent I) - Build promotion, canary deployment, guardrails
Post-Deploy Verification (Agent J) - KPI checks, monitoring, promotion decisions
Roadmap Feedback Loop (Agent K) - Create follow-ups, reprioritize, update forecasts
