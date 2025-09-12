# Architecture Decision Register - Zen CLI

Below are all the proposed Architecture Decision Records (ADRs) for the Zen AI-powered productivity suite.

ADRs document important architectural decisions which govern the project's design and development.

## What is an ADR?

An Architecture Decision Record captures a single architectural decision and its rationale. Each ADR describes:

- The context and problem statement
- The decision made
- The consequences of that decision
- Alternatives that were considered

---

## Foundation Architecture Decisions

The following ADRs document the architectural decisions made todate:

| ID                                                                         | Title                                                               | Status   | Date       | Supersedes | Related ADRs       |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------- | -------- | ---------- | ---------- | ------------------ |
| [ADR-0001](ADR-0001-go-language-choice.md)                                | Go Language Choice for Zen CLI Platform                            | Accepted | 2025-09-12 | —          | ADR-0002, ADR-0003 |
| [ADR-0002](ADR-0002-cobra-cli-framework.md)                               | Cobra CLI Framework Selection                                       | Accepted | 2025-09-12 | —          | ADR-0001, ADR-0004 |
| [ADR-0003](ADR-0003-project-structure-organization.md)                    | Project Structure and Organization                                  | Accepted | 2025-09-12 | —          | ADR-0001, ADR-0002 |
| [ADR-0004](ADR-0004-configuration-management-strategy.md)                 | Configuration Management Strategy                                   | Accepted | 2025-09-12 | —          | ADR-0002, ADR-0005 |
| [ADR-0005](ADR-0005-structured-logging-implementation.md)                 | Structured Logging Implementation                                   | Accepted | 2025-09-12 | —          | ADR-0004          |

## Planned Architecture Decisions

The following ADRs are planned for future implementation phases:

| ID                                                                         | Title                                                               | Status   | Date       | Supersedes | Related ADRs       |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------- | -------- | ---------- | ---------- | ------------------ |
| [ADR-0006](ADR-0006-plugin-architecture-design.md)                        | Plugin Architecture and Extension System                           | Proposed | TBD        | —          | ADR-0003          |
| [ADR-0007](ADR-0007-ai-agent-orchestration.md)                            | AI Agent Orchestration and Multi-Provider Support                  | Proposed | TBD        | —          | ADR-0008          |
| [ADR-0008](ADR-0008-llm-provider-abstraction.md)                          | LLM Provider Abstraction and Cost Management                       | Proposed | TBD        | —          | ADR-0007          |
| [ADR-0009](ADR-0009-workflow-state-management.md)                         | Workflow State Management and Persistence                          | Proposed | TBD        | —          | ADR-0010          |
| [ADR-0010](ADR-0010-integration-client-architecture.md)                   | External Integration Client Architecture                            | Proposed | TBD        | —          | ADR-0011          |
| [ADR-0011](ADR-0011-template-engine-design.md)                            | Template Engine and Content Generation                             | Proposed | TBD        | —          | ADR-0012          |
| [ADR-0012](ADR-0012-quality-gates-framework.md)                           | Quality Gates Framework and Automation                             | Proposed | TBD        | —          | ADR-0013          |
| [ADR-0013](ADR-0013-security-model-implementation.md)                     | Security Model and Threat Mitigation                               | Proposed | TBD        | —          | ADR-0004, ADR-0005 |
| [ADR-0014](ADR-0014-deployment-distribution-strategy.md)                  | Deployment and Distribution Strategy                                | Proposed | TBD        | —          | ADR-0001, ADR-0003 |
| [ADR-0015](ADR-0015-observability-monitoring-strategy.md)                 | Observability and Monitoring Strategy                              | Proposed | TBD        | —          | ADR-0005          |


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
