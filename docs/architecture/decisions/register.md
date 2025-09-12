# Architecture Decision Register - Helix

Below are all the proposed Architecture Decision Records (ADRs) for the helix agentic engineering framework.

ADRs document important architectural decisions which govern the project's design and development.

## What is an ADR?

An Architecture Decision Record captures a single architectural decision and its rationale. Each ADR describes:

- The context and problem statement
- The decision made
- The consequences of that decision
- Alternatives that were considered

---

All entries are **Proposed** with today’s date; update status as each ADR is reviewed.

| ID                                                                         | Title                                                               | Status   | Date       | Supersedes | Related ADRs       |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------- | -------- | ---------- | ---------- | ------------------ |
| [ADR-0001](ADR-0001-adopt-temporal-for-orchestration.md)                   | Adopt Temporal for durable orchestration                            | Proposed | 2025-09-01 | —          | ADR-0002, ADR-0005 |
| [ADR-0002](ADR-0002-python-3-12-fastapi-as-primary-runtime.md)             | Python 3.12 + FastAPI as primary runtime                            | Proposed | 2025-09-01 | —          | ADR-0014, ADR-0022 |
| [ADR-0003](ADR-0003-monorepo-structure-and-packaging.md)                   | Monorepo with Poetry workspaces; libs/services/agents layout        | Proposed | 2025-09-01 | —          | ADR-0014, ADR-0028 |
| [ADR-0004](ADR-0004-agents-as-microservices-with-single-responsibility.md) | Agents as stateless microservices (single-responsibility)           | Proposed | 2025-09-01 | —          | ADR-0015, ADR-0016 |
| [ADR-0005](ADR-0005-featureworkflow-per-jira-story.md)                     | One `FeatureWorkflow` per Jira Story; child workflows per stage     | Proposed | 2025-09-01 | —          | ADR-0001, ADR-0017 |
| [ADR-0006](ADR-0006-contracts-first-protobuf-openapi.md)                   | Contracts-first (Protobuf + OpenAPI) with buf/oasdiff gates         | Proposed | 2025-09-01 | —          | ADR-0015, ADR-0020 |
| [ADR-0007](ADR-0007-policy-as-code-with-opa-conftest.md)                   | Policy-as-code with OPA/Conftest for gates                          | Proposed | 2025-09-01 | —          | ADR-0015, ADR-0024 |
| [ADR-0008](ADR-0008-human-in-the-loop-approval-signals.md)                 | Human-in-the-loop approvals via Temporal Signals                    | Proposed | 2025-09-01 | —          | ADR-0005, ADR-0017 |
| [ADR-0009](ADR-0009-risk-tiers-drive-gates.md)                             | Risk tiers (Low/Medium/High) determine mandatory gates              | Proposed | 2025-09-01 | —          | ADR-0007, ADR-0017 |
| [ADR-0010](ADR-0010-kubernetes-deployment-model.md)                        | Kubernetes deployment with per-env namespaces                       | Proposed | 2025-09-01 | —          | ADR-0021, ADR-0023 |
| [ADR-0011](ADR-0011-secrets-management-with-vault-kms.md)                  | Secrets management via Vault/KMS; no plaintext secrets              | Proposed | 2025-09-01 | —          | ADR-0021, ADR-0024 |
| [ADR-0012](ADR-0012-observability-with-opentelemetry.md)                   | Observability: OpenTelemetry for traces/metrics/logs                | Proposed | 2025-09-01 | —          | ADR-0023, ADR-0026 |
| [ADR-0013](ADR-0013-structured-logging-and-correlation-ids.md)             | Structured JSON logging + correlation IDs (`jiraKey`,`workflowId`)  | Proposed | 2025-09-01 | —          | ADR-0012           |
| [ADR-0014](ADR-0014-tooling-poetry-ruff-mypy-pytest.md)                    | Tooling: Poetry, Ruff, mypy (strict), pytest (+async)               | Proposed | 2025-09-01 | —          | ADR-0003           |
| [ADR-0015](ADR-0015-gateway-gating-and-ci-checks.md)                       | Mandatory CI gates (format/lint/types/tests/contracts/security)     | Proposed | 2025-09-01 | —          | ADR-0006, ADR-0007 |
| [ADR-0016](ADR-0016-http-vs-grpc-for-agent-apis.md)                        | Transport: HTTP/JSON for agents initially; gRPC optional later      | Proposed | 2025-09-01 | —          | ADR-0004, ADR-0020 |
| [ADR-0017](ADR-0017-approval-slas-and-escalations.md)                      | Approval SLAs & escalations (Signals, timeouts, retries)            | Proposed | 2025-09-01 | —          | ADR-0008, ADR-0009 |
| [ADR-0018](ADR-0018-audit-log-and-retention.md)                            | Append-only audit log; ≥400 days retention                          | Proposed | 2025-09-01 | —          | ADR-0024           |
| [ADR-0019](ADR-0019-feature-flags-strategy.md)                             | Feature flags for user-visible changes; owner + expiry              | Proposed | 2025-09-01 | —          | ADR-0021           |
| [ADR-0020](ADR-0020-schema-migrations-expand-contract.md)                  | DB & API schema migrations via expand/contract                      | Proposed | 2025-09-01 | —          | ADR-0006, ADR-0016 |
| [ADR-0021](ADR-0021-build-once-promote-slsa-sbom.md)                       | Build-once promote; signed images + SBOM (SLSA-style)               | Proposed | 2025-09-01 | —          | ADR-0010, ADR-0011 |
| [ADR-0022](ADR-0022-clients-for-integrations.md)                           | Dedicated clients for Jira/SCM/Confluence/Flags/Sourcegraph         | Proposed | 2025-09-01 | —          | ADR-0002           |
| [ADR-0023](ADR-0023-observability-dashboards-and-slos.md)                  | Delivery & reliability dashboards; SLO/error-budget reviews         | Proposed | 2025-09-01 | —          | ADR-0012, ADR-0013 |
| [ADR-0024](ADR-0024-security-scanning-and-cve-policy.md)                   | Security scanning policy (SAST/SCA/IaC/DAST); CVE waiver rules      | Proposed | 2025-09-01 | —          | ADR-0011, ADR-0018 |
| [ADR-0025](ADR-0025-webhook-ingest-to-signal-mapping.md)                   | Webhook Ingest: map Jira/SCM/CI events to Signals                   | Proposed | 2025-09-01 | —          | ADR-0005, ADR-0022 |
| [ADR-0026](ADR-0026-pr-size-and-coverage-guardrails.md)                    | PR guardrails (≤400 lines; per-diff coverage; CODEOWNERS)           | Proposed | 2025-09-01 | —          | ADR-0015           |
| [ADR-0027](ADR-0027-environment-configuration-strategy.md)                 | Config strategy: pydantic settings + env layering                   | Proposed | 2025-09-01 | —          | ADR-0010           |
| [ADR-0028](ADR-0028-library-boundaries-and-sharing.md)                     | Shared libraries boundaries (`helix-common`, `helix-clients`, etc.) | Proposed | 2025-09-01 | —          | ADR-0003, ADR-0014 |
| [ADR-0029](ADR-0029-ml-in-agents-optional-baselines.md)                    | ML in agents as optional baselines; deterministic fallback          | Proposed | 2025-09-01 | —          | ADR-0002, ADR-0030 |
| [ADR-0030](ADR-0030-model-artifacts-versioning-and-governance.md)          | Model artifacts versioning, provenance & rollout controls           | Proposed | 2025-09-01 | —          | ADR-0029           |
| [ADR-0031](ADR-0031-error-handling-retries-and-compensations.md)           | Error handling: bounded retries + SAGA compensations                | Proposed | 2025-09-01 | —          | ADR-0001, ADR-0005 |
| [ADR-0032](ADR-0032-access-control-and-rbac-for-approvals.md)              | Access control/RBAC for approvals in Control Plane                  | Proposed | 2025-09-01 | —          | ADR-0008, ADR-0011 |
| [ADR-0033](ADR-0033-transport-security-and-mtls-internal.md)               | Transport security: TLS everywhere; mTLS for internal calls         | Proposed | 2025-09-01 | —          | ADR-0011, ADR-0024 |
| [ADR-0034](ADR-0034-incident-and-rollback-strategy.md)                     | Incident & rollback strategy; automated rollback triggers           | Proposed | 2025-09-01 | —          | ADR-0010, ADR-0021 |
| [ADR-0035](ADR-0035-data-privacy-and-pii-handling.md)                      | Data privacy: no PII in logs; redaction & sampling rules            | Proposed | 2025-09-01 | —          | ADR-0011, ADR-0013 |
| [ADR-0036](ADR-0036-temporal-task-queue-strategy.md)                       | Task queue strategy (per-stage/per-agent), scaling & isolation      | Proposed | 2025-09-01 | —          | ADR-0001, ADR-0010 |
| [ADR-0037](ADR-0037-release-strategy-canary-vs-blue-green.md)              | Release strategy: canary for agents; blue/green for orchestrator    | Proposed | 2025-09-01 | —          | ADR-0021, ADR-0034 |
| [ADR-0038](ADR-0038-telemetry-ids-and-event-schema.md)                     | Telemetry/event schema and ID conventions                           | Proposed | 2025-09-01 | —          | ADR-0012, ADR-0013 |
| [ADR-0039](ADR-0039-artifact-storage-and-retention.md)                     | Artifact storage (reports/SBOMs) and retention policy               | Proposed | 2025-09-01 | —          | ADR-0021, ADR-0018 |
| [ADR-0040](ADR-0040-architecture-docs-adr-governance.md)                   | Architecture docs & ADR governance (templates, review cadence)      | Proposed | 2025-09-01 | —          | ADR-0003           |


---

### Status legend

- **Proposed** — Under review, not yet adopted.
- **Accepted** — Adopted as the standard going forward.
- **Rejected** — Considered but not adopted.
- **Superseded** — Replaced by a newer ADR (link it in *Supersedes*).

---

## Creating a New ADR

Note. You must use template provided [adr-template.md](./adr-template.md) to create a new ADRs.

1. Copy the template to a new file with the naming convention: `ADR-{NUMBER}-{title-with-dashes}.md`

   ```bash
   cp docs/architecture/decisions/adr-template.md docs/architecture/decisions/adr/ADR-{####}-{short-title}.md
   ```

The filenames are following the pattern `ADR-####-title-with-dashes.md`.

- The prefix is `ADR` to identify a Architecture Decision Record (ADR).
- `####` is a consecutive number in sequence.
- The title is stored using dashes and lowercase.
- The filetype is `.md`, because it is a Markdown file.

1. Complete all the sections of the ADR template.
   - **Status**: `Proposed`
   - **Context**: Why was this decision needed?
   - **Decision**: What was decided and why?
   - **Consequences**: What are the trade-offs?
   - **Alternatives**: What else was considered?

1. Submit for review via pull request (PR). Start with **Proposed**.

1. On acceptance:
   - Update the status to **Accepted**.
   - Add a row to this Architecture Decision Register (`decisions/register.md`) index.

---

## Resources

- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
- [ADR Tools](https://github.com/npryce/adr-tools)
- [ADR GitHub Organization](https://adr.github.io/)
