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

| ID                                               | Title                                   | Status   | Date       | Supersedes | Related ADRs       |
| ------------------------------------------------ | --------------------------------------- | -------- | ---------- | ---------- | ------------------ |
| [ADR-0001](ADR-0001-language-choice.md)          | Language Choice for Zen CLI Platform    | Accepted | 2025-09-12 | —          | ADR-0002, ADR-0003 |
| [ADR-0002](ADR-0002-cli-framework.md)            | Cobra CLI Framework Selection           | Accepted | 2025-09-12 | —          | ADR-0001, ADR-0004 |
| [ADR-0003](ADR-0003-project-structure.md)        | Project Structure and Organization      | Accepted | 2025-09-12 | —          | ADR-0001, ADR-0002 |
| [ADR-0004](ADR-0004-configuration-management.md) | Configuration Management Strategy       | Accepted | 2025-09-12 | —          | ADR-0002, ADR-0005 |
| [ADR-0005](ADR-0005-structured-logging.md)       | Structured Logging Implementation       | Accepted | 2025-09-12 | —          | ADR-0004           |
| [ADR-0006](ADR-0006-factory-pattern.md)          | Factory Pattern Implementation          | Accepted | 2025-09-13 | —          | ADR-0003, ADR-0007 |
| [ADR-0007](ADR-0007-command-orchestration.md)    | Command Orchestration Design            | Accepted | 2025-09-13 | —          | ADR-0006, ADR-0002 |

## Planned Architecture Decisions

The following ADRs are planned for future implementation phases:

| ID                                               | Title                                   | Status   | Date       | Supersedes | Related ADRs       |
| ------------------------------------------------ | --------------------------------------- | -------- | ---------- | ---------- | ------------------ |
| [ADR-0008](ADR-0008-plugin-architecture.md)      | Plugin Architecture Design              | Proposed | TBD        | —          | ADR-0003, ADR-0006 |
| [ADR-0009](ADR-0009-agent-orchestration.md)      | AI Agent Orchestration                  | Proposed | TBD        | —          | ADR-0010           |
| [ADR-0010](ADR-0010-llm-abstraction.md)          | LLM Provider Abstraction                | Proposed | TBD        | —          | ADR-0009           |
| [ADR-0011](ADR-0011-workflow-management.md)      | Workflow State Management               | Proposed | TBD        | —          | ADR-0012           |
| [ADR-0012](ADR-0012-integration-architecture.md) | External Integration Architecture       | Proposed | TBD        | —          | ADR-0013           |
| [ADR-0013](ADR-0013-template-engine.md)          | Template Engine Design                  | Proposed | TBD        | —          | ADR-0014           |
| [ADR-0014](ADR-0014-quality-gates.md)            | Quality Gates Framework                 | Proposed | TBD        | —          | ADR-0015           |
| [ADR-0015](ADR-0015-security-model.md)           | Security Model Implementation           | Proposed | TBD        | —          | ADR-0004, ADR-0005 |
| [ADR-0016](ADR-0016-deployment-strategy.md)      | Deployment Distribution Strategy        | Proposed | TBD        | —          | ADR-0001, ADR-0003 |
| [ADR-0017](ADR-0017-observability-strategy.md)   | Observability Monitoring Strategy       | Proposed | TBD        | —          | ADR-0005           |

---

### Status legend

- **Proposed** — Under review, not yet adopted.
- **Accepted** — Adopted as the standard going forward.
- **Rejected** — Considered but not adopted.
- **Superseded** — Replaced by a newer ADR (link it in *Supersedes*).

---

## Creating a New ADR

Note. You must use template provided [adr-template.md](./adr-template.md) to create a new ADRs.

1. Copy the template to a new file with the naming convention: `ADR-{NUMBER}-{short-title}.md`

   ```bash
   cp docs/architecture/decisions/adr-template.md docs/architecture/decisions/adr/ADR-{####}-{short-title}.md
   ```

The filenames are following the pattern `ADR-####-short-title.md`.

- `ADR` prefix to identify a Architecture Decision Record (ADR).
- `####` is a consecutive number in sequence.
- `short-title` is maximum two words using dashes and lowercase.
- `.md` filetype is , because it is a Markdown file.

2. Complete all the sections of the ADR template.
  - **Status**: `Proposed`
  - **Context**: Why was this decision needed?
  - **Decision**: What was decided and why?
  - **Consequences**: What are the trade-offs?
  - **Alternatives**: What else was considered?

3. Submit for review via pull request (PR). Start with **Proposed**.

4. On acceptance:
   - Update the status to **Accepted**.
   - Add a row to this Architecture Decision Register (`decisions/register.md`) index.

---

## Resources

- [ADR GitHub Organization](https://adr.github.io/)
- [Markdown Architectural Decision Records (MADR)](https://adr.github.io/madr/)
- [ADR Tools](https://github.com/npryce/adr-tools)
