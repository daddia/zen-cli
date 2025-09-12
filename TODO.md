# TODO - Helix

*Last updated: 2025-09-01*

---

## Conventions

- **Priority:** P1 (urgent/critical), P2 (high), P3 (normal), P4 (low)
- **Status:** Not started · In progress · Blocked · In review · Done
- **Estimates:** S (≤2h), M (≤1d), L (≤3d), XL (>3d)
- **DoD (Definition of Done):** Tests updated · Docs updated · Review approved · Deployed/Released (as applicable)

---

## Backlog

### Sprint 1 (1–14 Sep 2025) — **Hello Temporal (golden path)** → Release v0.1.0

- [ ] **Temporal dev up (compose)** (P1, M): Stand up Temporal Server, Postgres, Jaeger via docker-compose; healthcheck docs.  
      DoD: Docs updated · Review approved
- [ ] **Orchestrator skeleton** (P1, L): `FeatureWorkflow` + one child workflow; Signals/Queries types; sample activity.  
      DoD: Tests updated · Docs updated · Review approved
- [ ] **Control Plane (FastAPI) minimal** (P1, L): `/workflows/start`, `/approvals/{signal}`; OIDC stub.  
      DoD: Tests updated · Docs updated · Deployed
- [ ] **Webhook Ingest stub** (P2, M): Accept Jira/SCM webhooks; log-only mapping to Signals.  
      DoD: Tests updated · Docs updated
- [ ] **Agent A — Discovery (v0)** (P1, L): Create ADR outline from Jira context; post back to Jira comment (mock Confluence ok).  
      DoD: Tests updated · Docs updated
- [ ] **Agent E — Code Review (v0)** (P1, L): PR size check (≤400 lines, excluding gen/lockfiles) + CODEOWNERS presence; structured findings.  
      DoD: Tests updated · Docs updated
- [ ] **CI pipeline bootstrap** (P1, L): fmt (ruff), lint (ruff), types (mypy), unit tests; required checks on main.  
      DoD: Deployed
- [ ] **Observability baseline** (P2, M): OTEL traces for orchestrator/control-plane/agents; Jaeger UI linked.  
      DoD: Docs updated

### Sprint 2 (15–28 Sep 2025) — **Gate the path (contracts & architecture)** → Release v0.2.0

- [ ] **Contracts-first wiring** (P1, L): buf for Protobuf; oasdiff for REST; non-breaking gate vs `main`.  
      DoD: Tests updated · Docs updated · Review approved
- [ ] **Agent C — Design & Contract (v1)** (P1, L): Generate/validate contracts; expand/contract plan output.  
      DoD: Tests updated · Docs updated
- [ ] **Agent F — Architecture Review (v0)** (P1, L): Policy-as-code (OPA) minimal checks for NFRs; auto/route to human.  
      DoD: Tests updated · Docs updated
- [ ] **Risk tiers config** (P2, M): `policy.yaml` with Low/Medium/High → which gates are mandatory.  
      DoD: Docs updated
- [ ] **PR guardrails** (P1, M): Per-diff coverage gate; enforce ≤400 lines; CODEOWNERS approval.  
      DoD: Deployed

### Sprint 3 (29 Sep–12 Oct 2025) — **Test & Security foundations** → Release v0.3.0

- [ ] **Agent G — Test & QA (v0)** (P1, L): Orchestrate unit/contract/integration smokes in ephemeral env; flake detector.  
      DoD: Tests updated · Docs updated
- [ ] **Ephemeral integration env** (P1, L): Compose profile for integration; seed fixtures; deterministic runs.  
      DoD: Deployed
- [ ] **Agent H — Security (v0)** (P1, L): SAST/SCA/secrets + Trivy image scan; CVE policy & waiver path.  
      DoD: Docs updated · Deployed
- [ ] **Audit log (append-only)** (P2, L): Record approvals, agent decisions, policy outcomes; 400-day retention.  
      DoD: Docs updated

### Sprint 4 (13–26 Oct 2025) — **Release & Verify** → Release v0.4.0

- [ ] **Agent I — Release Manager (v0)** (P1, L): Build-once promote; canary rollout (dummy target); rollback on breach.  
      DoD: Tests updated · Docs updated · Deployed
- [ ] **Feature flags (OpenFeature)** (P1, M): Flag SDK + owner/expiry convention; sample gated endpoint.  
      DoD: Docs updated · Deployed
- [ ] **Agent J — Post-Deploy Verify (v0)** (P1, L): Guardrail checks (latency/error rate); final verdict Signal.  
      DoD: Tests updated · Docs updated
- [ ] **Runbooks & dashboards** (P2, M): Canary, rollback; delivery and reliability dashboards published.  
      DoD: Docs updated

### Sprint 5 (27 Oct–9 Nov 2025) — **Feedback loop & prioritisation** → Release v0.5.0

- [ ] **Agent K — Roadmap Feedback (v0)** (P1, L): Open follow-ups (bugs/iterations/tech debt) with preliminary WSJF.  
      DoD: Docs updated · Deployed
- [ ] **Agent B — Prioritisation (v0)** (P2, L): Rule-based WSJF/ICE; propose ordering; PM approval Signal.  
      DoD: Docs updated
- [ ] **Webhook mapping v1** (P2, M): Robust mappings Jira/SCM/CI → Signals; idempotency keys.  
      DoD: Tests updated · Docs updated
- [ ] **Delivery scorecard** (P2, M): DORA + flow metrics; weekly view.  
      DoD: Docs updated

### Sprint 6 (10–23 Nov 2025) — **Harden & prepare v1** → Release v1.0.0-rc1

- [ ] **SLAs & escalation automation** (P1, L): Time-boxed waits → Slack/email escalations → auto-Blocked transition.  
      DoD: Docs updated · Deployed
- [ ] **Compensation library** (P1, L): Standard compensations (close draft ADR, revert flag/label, reopen Jira).  
      DoD: Tests updated · Docs updated
- [ ] **RBAC for approvals** (P1, L): Control-plane roles (PM/TL/Architect/Security); audit trails.  
      DoD: Docs updated · Deployed
- [ ] **mTLS + Vault integration** (P1, L): mTLS for internal calls; secrets via Vault/KMS only.  
      DoD: Docs updated · Deployed
- [ ] **Policy pack expansion** (P2, L): Architecture/security gates extended; risk-tier tuning.  
      DoD: Docs updated

---

## Continuous Items (Track Throughout)

- [ ] **Performance:** Keep orchestrator overhead <200ms per step; agent p95 <2s; watch queue latency.
- [ ] **Quality:** Enforce PR guardrails; no net coverage decline; fix flaky tests within 48h.
- [ ] **Security:** Zero plaintext secrets; Critical/High CVEs require approved time-boxed waivers.
- [ ] **Dependencies:** Weekly dep scan & update cadence; track image base versions.
- [ ] **Documentation:** ADRs, runbooks, READMEs maintained alongside code.
- [ ] **Monitoring:** Dashboards for delivery, gate health, queue depth, rollback rate; alert on SLO breaches.

---

## Future Enhancement Candidates

### **ML Assist (opt-in)**
- [ ] **PR risk scoring (baseline ML)**: Lightweight model from code churn/ownership signals with deterministic fallback.
- [ ] **Backlog ranking (ML)**: Learn WSJF/ICE from historical outcomes; A/B guardrails.

### **Scalability & Multi-tenant**
- [ ] **Per-team namespaces & policy packs**: Isolate queues/policies; per-team dashboards.
- [ ] **Task queue strategy v2**: Dynamic sharding; HPA on queue depth.

### **Protocol & Contracts**
- [ ] **gRPC for hot paths**: Upgrade select agent calls to gRPC; HTTP remains default.
- [ ] **Schema registry**: Central registry for API/DB contract versions & deprecation windows.

### **Knowledge & Insights**
- [ ] **Artifact graph**: Link issues, PRs, ADRs, builds; enable richer agent context.
- [ ] **Cost guardrails**: Release agent checks for cost regressions before promotion.

---

## Completed
> Move finished items here to keep the active list clean.

<!-- None yet — populate as we deliver releases -->
