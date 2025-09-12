---
# Optional metadata
project_key: <PROJECT>
issue_key: <PROJECT-123>
summary: <Summary>
parent_key: <PROJECT-467> # from Jira
component: <e.g. Storefront | Webster | BFF | Central>
status: <e.g To Do, In Progress, Blocked, In Review, Done>
priority: <e.g. Critial (P1) | High (P2) | Medium (P3) | Low (P4) | Lowest (P5)>
story_points: <Story points estimate>
owner: <assignee>
stakeholders: [<product>, <design>, <architecture>, <QA>, <ops>]
sprint: <e.g. Sprint 12>
risk_level: <Low|Medium|High>
labels: [story, storefront, bff]
created: <YYYY-MM-DD>
---

# Story Title
Short, outcome-oriented. Imperative voice. (e.g. “Support guest checkout with card vaulting”)

## 1. User Story
As a **<user/persona>**, I want **<capability/behaviour>** so that **<outcome/benefit>**.

## 2. Context & Problem
- What’s the customer problem? Why now?
- Links to discovery, research, prior incidents, or constraints.

## 3. Business Value & Success Measures

**Value:** <revenue, conversion, CSAT, cost saving, risk reduction>

**Success metrics (MUST):**
- <e.g. checkout conversion +1.5pp, p95 latency < 200ms>

**Guardrails (SHOULD):**
- <e.g. error rate ≤ 0.5%, zero P1 regressions>

## 4. Scope

**In-scope (MUST):**
- In scope item…

**Out-of-scope (WON’T for this story):**
- Out of scope item…

## 5. Acceptance Criteria (Gherkin)

**Scenario: <happy path>**
- **Given** …
- **When** …
- **Then** …

**Scenario: <validation/error>**
- **Given** …
- **When** …
- **Then** …

**Scenario: <edge case>**
- **Given** …
- **When** …
- **Then** …

> Notes: Prefer explicit, testable behaviours. Avoid ambiguous terms (“fast”, “soon”). Include errors, empty states, retries, timeouts, and accessibility behaviours.

## 6. UX / Content

- **Designs:** <Figma link(s)>

- **Content rules:** tone, microcopy, validation messages.

- **Accessibility (MUST):** WCAG 2.2 AA; keyboard/ SR flows documented.

## 7. Data, Telemetry & Events

- **Tracking plan (MUST):** events, properties, naming (owner: analytics)
- **Data model impacts:** schema changes, PII handling, retention.
- **Experimentation (IF ANY):** flag name, variants, exposure, success metrics.

## 8. API & Contract Changes

- **Endpoints touched/new:** `GET /…`, `POST /…`
- **Request/response diffs:** link to OpenAPI/Buf schema MR.
- **Backward compatibility:** versioning/tag strategy, deprecation plan.
- **Contract tests (MUST):** where & how they run in CI.

## 9. Non-Functional Requirements (Budgets)

| Area            | Requirement (MUST unless stated)                     |
|-----------------|------------------------------------------------------|
| Performance     | p95 < 200ms BFF; p95 < 350ms page TTFB (unauth); cache strategy defined |
| Availability    | ≥ 99.9% for affected path; graceful degradation      |
| Security        | OWASP controls; authN/Z rules; secrets handling      |
| Privacy         | PII classes; DSR readiness; data residency           |
| Accessibility   | WCAG 2.2 AA; focus order; contrast; ARIA            |
| SEO             | Indexing rules; canonical; structured data (if UI)   |
| International   | i18n/l10n notes (if applicable)                      |

## 10. Operational Readiness

**Observability (MUST):**
- logs, traces, metrics; dashboard link; alert thresholds.

**Feature flag (SHOULD):**
-`<flag_name>`; rollout plan; kill-switch behaviour.

**Runbook:**
- link with symptoms, checks, remediation, rollback.

**Rate limiting / circuit breaking:**
- defined if calling flaky upstreams.

## 11. Dependencies & Constraints

**Internal dependencies:**
- list internal services

**External dependencies:**
- list external, third-parties, etc.

**Environment constraints:**
- list environments, credentials, data sets

**Approval requirements:**
- list approvals, reviews, etc.

## 12. Risks & Assumptions

**Risks:**
- <risk> → **Mitigation:** <plan>

**Assumptions:**
- <what must hold true to deliver>

## 13. Test Approach

- **Unit:** coverage target & key cases
- **Contract:** provider/consumer tests & fixtures
- **Integration/E2E:** environments, test data, negative paths
- **Accessibility & visual:** tooling (e.g. axe), critical flows
- **Performance:** load profile, thresholds, test plan

## 14. Rollout & Backout

**Rollout plan:**
- dev → staging → prod (progressive exposure, %/cohorts)

**Backout plan (MUST):**
- How to disable/revert safely; data cleanup steps.

**Comms:**
- Who to notify on change & failure.

## 15. Estimation & Tracking

**Story Points:** <N>

**Time tracking (optional):**
- Original Estimate <h>
- Remaining <h>

**Related issues:**
- Sub-tasks, bugs, spikes

## 16. Links & Traceability

- ADR(s): <links>
- MR(s): <links>
- Confluence / Docs: <links>

---

## Definition of Ready (DoR) – checklist

- [ ] Clear user story & value articulated
- [ ] Acceptance Criteria complete & testable (incl. errors/edges)
- [ ] Designs/content available or consciously deferred
- [ ] NFR budgets agreed (perf, security, accessibility, etc.)
- [ ] Dependencies identified & not blocking
- [ ] Data/tracking plan agreed
- [ ] Test approach agreed
- [ ] Estimation aligned with team norms

## Definition of Done (DoD) – checklist

- [ ] Code merged with green CI (lint, tests, security scans)
- [ ] Contract tests passing; schema/SDKs updated
- [ ] A/Cs verified (QA or developer-in-test evidence)
- [ ] Observability in place; dashboards & alerts updated
- [ ] Docs/runbook updated; changelog entry added
- [ ] Feature flagged & rolled out per plan (or safely disabled)
- [ ] No P1/P2 regressions in affected areas
