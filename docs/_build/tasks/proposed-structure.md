# Proposed Structure

Here’s a **comprehensive, lifecycle-aware** structure for an individual task folder that cleanly accommodates Jira, Confluence, Figma, GitHub, testing, security, release, and ops artefacts — without turning the repo into a dumping ground.

Use plural container (`tasks/`) and a stable key per task (e.g. `JIRA-124/`). Keep **curated metadata** small and **raw system payloads** isolated.

```
repo-root/
└─ work/
   └─ tasks/
      └─ JIRA-124/
         ├─ index.md                      # Human landing page (what/why/scope/DoD/links)
         ├─ manifest.yaml                 # Curated, machine-readable metadata (schemaed)
         ├─ .taskrc.yaml                  # Task-local config (owners, labels, lint toggles)
         ├─ metadata/                     # Raw snapshots from external systems (read-only)
         │  ├─ jira.json                  # GET /rest/api/3/issue/JIRA-124 (full payload)
         │  ├─ github-prs.json            # PR list snapshot (ids, states)
         │  ├─ confluence.json            # Page/meta snapshot (if synced)
         │  └─ figma.json                 # File/key + nodes of interest (optional)
         ├─ docs/                         # Human docs (decisions, notes, research)
         │  ├─ meeting-notes/
         │  │  └─ 2025-09-13-kickoff.md
         │  ├─ research/
         │  │  ├─ user-interviews.md
         │  │  └─ references.md
         │  └─ adr/
         │     └─ ADR-0001-cache-invalidation.md
         ├─ design/                       # Design deliverables (links + exports)
         │  ├─ links.md                   # Figma files/frames, Confluence specs
         │  ├─ figma-exports/             # PNG/SVG/PDF exports (auto-generated)
         │  └─ tokens/                    # Design tokens snippets (JSON/Tailwind/etc.)
         ├─ specs/                        # Contracts: OpenAPI, JSON Schema, Protocol Buffers
         │  ├─ storefront.cart.v1.openapi.yaml
         │  ├─ cart-item.schema.json
         │  └─ events/
         │     └─ cart.item.published.schema.json
         ├─ code/                         # Spikes/PoCs local to this task (throwaway or merged later)
         │  ├─ spike-notes.md
         │  └─ examples/
         ├─ qa/                           # Test planning + evidence
         │  ├─ test-plan.md               # What, how, environments, acceptance criteria
         │  ├─ test-cases.csv             # Optional: tabular test cases
         │  ├─ results/                   # Screenshots, videos, junit, allure
         │  └─ accessibility/             # axe reports, manual checklists
         ├─ perf/                         # Performance test scripts + results
         │  ├─ k6/
         │  │  └─ cart_scenario.js
         │  └─ results/
         ├─ security/                     # Threat model + scans + approvals
         │  ├─ threat-model.md            # STRIDE/LINDDUN; DFDs if relevant
         │  ├─ checks.md                  # OWASP, authz review, secrets review
         │  └─ sbom/                      # CycloneDX/SPDX if applicable
         ├─ data/                         # Inputs/outputs used during the work
         │  ├─ raw/
         │  └─ processed/
         ├─ scripts/                      # Task-scoped helpers (sync, lint, gen)
         │  ├─ sync-jira.sh               # Writes metadata/jira.json; updates manifest.yaml
         │  ├─ export-figma.mjs           # Exports frames → design/figma-exports/
         │  └─ validate.sh                # Lints markdown/yaml/json/openapi
         ├─ reviews/                      # Review evidence, decisions, PR discussions
         │  ├─ review-notes.md
         │  └─ screenshots/
         ├─ approvals/                    # Sign-offs: design, security, privacy, release
         │  ├─ design-signoff.md
         │  ├─ security-approval.md
         │  └─ release-approval.md
         ├─ release/                      # Rollout, FFs, comms, runbooks-to-adopt
         │  ├─ feature-flag.md            # Flag name, owner, exposure plan, kill-switch
         │  ├─ rollout-plan.md            # Envs, % ramps, guardrails
         │  ├─ rollback.md                # Preconditions, steps, verification
         │  └─ comms.md                   # Stakeholder comms (exec, CS, marketing)
         ├─ ops/                          # Operational handover hooks
         │  ├─ runbook-diff.md            # Delta to existing runbooks
         │  ├─ alerts-dashboards.md       # Links to Datadog/Grafana; SLO impacts
         │  └─ post-deploy-checks.md      # Smoke/perf/accessibility checks
         ├─ artifacts/                    # Produced artefacts to keep (immutable)
         │  ├─ diagrams/                  # C4, sequence, flow
         │  ├─ exports/                   # PDFs, archives
         │  └─ recordings/                # Loom/MP4 (if policy allows)
         └─ links.md                      # Friendly link hub (Jira, PRs, Figma, Confluence, runs)
```

## What goes where (quick rules)

* **Curated facts** → `manifest.yaml` (CI-validated, human-edited).
* **Raw system dumps** → `metadata/` (machine-written snapshots).
* **Design** (Figma, tokens, exports) → `design/` (keep sources linked, exports versioned).
* **Contracts** → `specs/` (OpenAPI/Schema/Event contracts; lint in CI).
* **Testing evidence** → `qa/`, `perf/`, `accessibility/` (attach results).
* **Security** (threat model, checks, SBOM) → `security/`.
* **Release plans & feature flags** → `release/`.
* **Ops deltas** → `ops/` (what changes in runbooks/alerts).
* **Immutable proof** (PDFs, diagrams, recordings) → `artifacts/`.
* **PR/Review chatter & screenshots** → `reviews/`.
* **Formal sign-offs** → `approvals/`.

---

## Templates (drop these straight in)

### `manifest.yaml`

Keep small, typed, and useful for dashboards. Validate via JSON Schema in CI.

```yaml
schema: 1
key: JIRA-124
title: "Cart API: add item-level promotions"
type: Story            # Story | Bug | Spike | Task
status: In Progress    # Proposed | In Progress | Blocked | Done
priority: High
owner:
  name: "Karla Smith"
  email: "karla@example.com"
squad: "Basket & Checkout"
labels: [storefront, bff, api, performance]
dates:
  created: 2025-09-12
  target: 2025-09-26
  last_updated: 2025-09-13
delivery:
  branch: "feature/JIRA-124-cart-item-promos"
  feature_flag: "cart.item_promotions"
  environments: ["dev", "staging", "prod"]
  dependencies: ["CENTRAL-API-42"]
success_criteria:
  - "OpenAPI published; SDKs regenerated"
  - "/cart p95 ≤ 200 ms @ 25 items"
  - "95% BFF cache hit for cart reads"
links:
  jira: "https://jira.example.com/browse/JIRA-124"
  epic: "JIRA-100"
  prs:
    - "https://github.com/org/repo/pull/987"
  confluence:
    - "https://confluence.example.com/x/ABC123"
  figma:
    - "https://www.figma.com/file/XYZ?node-id=123"
  dashboards:
    - "https://app.datadoghq.com/…"
risk:
  level: Medium
  notes: "Cache invalidation touchpoints; promo edge-cases"
```

### `.taskrc.yaml`

Local knobs for tooling and ownership.

```yaml
owners: ["@karla", "@jd"]
reviewers: ["@arch-alex", "@qa-sam"]
linters:
  markdownlint: true
  yamllint: true
  spectral_openapi: true
  jsonschema: true
policies:
  store_raw_metadata: true
  commit_large_binaries: false
generated:
  collapse_in_diff: ["metadata/*.json", "qa/results/**", "perf/results/**"]
```

### `index.md`

Short, human-friendly, front door.

```markdown
# JIRA-124 — Cart API: item-level promotions

**Status:** In Progress · **Owner:** Karla S. · **Squad:** Basket & Checkout

## Goal
Expose item-level promotions in `/cart` via BFF with ≤200 ms p95 and high cache hit rate.

## Scope
- Map `promo_items[]` from Central → `cart.items[].promotions[]`.
- Update OpenAPI (`storefront.cart.v1`) + JSON Schema (`cart-item.schema.json`).

## Definition of Done
- Contracts merged & published; SDKs regenerated.
- Perf target met in staging for 3 days.
- Accessibility checks pass for affected UI.
- Runbook + flag rollout/rollback documented.

👉 See `links.md` for Jira, PRs, Figma, Confluence.
```

### `design/links.md`

```markdown
# Design Links
- **Figma File:** <https://www.figma.com/file/XYZ?node-id=123>
  - Frames: `Cart/Item Promo`, `Cart/Empty State`
- **Design tokens:** `design/tokens/cart.promotions.tokens.json`
- **Confluence Spec:** <https://confluence.example.com/x/ABC123>
```

### `qa/test-plan.md`

```markdown
# Test Plan — JIRA-124
**Scope:** BFF `/cart` response structure; UI display of promotions; analytics event.

## Strategy
- Contract tests via OpenAPI (Dredd/Newman).
- E2E (Cypress) happy path + edge cases (no promos, multiple promos).
- Accessibility: axe + manual keyboard checks.

## Environments
- Dev, Staging (prod shadow traffic optional).

## Acceptance
- All tests pass; promo analytics event received; SLO unchanged.
```

### `security/threat-model.md`

```markdown
# Threat Model — JIRA-124
**Context:** New promo fields exposed via BFF.

- **Assets:** price/savings visibility; promo eligibility rules.
- **Entry points:** `/cart` (GET/POST) via BFF (KrakenD).
- **Concerns:** info disclosure of ineligible promos; cache poisoning.

**Mitigations**
- Authorisation check retained at Central; fields derived server-side.
- BFF cache key includes `auth` + `customer_segment`.
- Schemas validated at BFF ingress/egress.
```

### `release/feature-flag.md`

```markdown
# Feature Flag
name: cart.item_promotions
owner: karla
exposure: 0% → 10% → 50% → 100% (staging, then prod)
guardrails: error rate, p95 latency, conversion at cart
kill_switch: revert env var or config + purge BFF cache
```

### `links.md`

```markdown
# Links
- **Jira:** https://jira.example.com/browse/JIRA-124
- **Epic:** JIRA-100
- **PRs:** https://github.com/org/repo/pull/987
- **Confluence:** https://confluence.example.com/x/ABC123
- **Figma:** https://www.figma.com/file/XYZ?node-id=123
- **Builds:** GitHub Actions (workflow run): <…> · GitLab pipelines: <…>
- **Dashboards:** Datadog cart latency, BFF cache hit: <…>
```

---

## Conventions & hygiene

* **Names:** `YYYY-MM-DD-{purpose}-{slug}.md` for notes; ISO dates for sortability.
* **Binary exports:** keep in `design/figma-exports/` or `artifacts/exports/`, **never** in root.
* **Large/Generated files:** mark with `.gitattributes` as `linguist-generated` so diffs collapse.
* **Validation in CI (at task folder level):**

  * `yamllint` + JSON Schema validation for `manifest.yaml`
  * `spectral` (OpenAPI lint) for `specs/**/*.yaml`
  * `markdownlint` for `*.md`
  * Optional: verify Figma links resolve (headless check), Jira key exists.

---

## Minimal automation (works from repo root)

* `scripts/sync-jira.sh KEY` → writes `metadata/jira.json`, updates `manifest.yaml`.
* `scripts/export-figma.mjs KEY` → exports tagged frames → `design/figma-exports/`.
* `scripts/validate.sh KEY` → runs all linters/validators for this task.

---

This structure keeps a single task self-contained across **discovery → design → build → test → release → operate**, while making it dead-simple to automate and audit. If you want, I can generate a tiny **scaffold script** that creates this layout given a Jira key and title.
