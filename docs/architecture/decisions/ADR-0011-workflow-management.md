---
status: Proposed
date: 2025-09-16
decision-makers: Development Team, Architecture Team, Product Team
consulted: Engineering Leadership, Platform Engineering Team
informed: DevOps Team, QA Team
---

# ADR-0011 - Workflow State Management

## Context and Problem Statement

The Zen CLI platform orchestrates complex multi-stage workflows across the product lifecycle, including a 12-stage engineering workflow (Discover → Prioritize → Design → Architect → Plan → Build → Review → Test → Secure → Release → Validate → Feedback) and product management workflows. These workflows require persistent state management, progress tracking, rollback capabilities, and coordination between human decisions and AI-powered automation. The state management system must support concurrent workflows, branching scenarios, and integration with external systems while maintaining consistency and auditability.

## Decision Drivers

* **Workflow Complexity**: Support for 12-stage engineering workflows with conditional branching and parallel execution
* **State Persistence**: Reliable state storage with transaction support and crash recovery capabilities
* **Concurrency**: Support for multiple concurrent workflows without conflicts or race conditions
* **Integration**: Seamless state synchronization with external systems (Jira, GitHub, CI/CD pipelines)
* **Auditability**: Complete audit trail of state changes and decision points for compliance and debugging
* **Performance**: Fast state queries and updates to maintain responsive CLI interactions
* **Rollback Support**: Ability to revert to previous workflow states and handle failed operations gracefully

## Considered Options

1. **Event Sourcing with Aggregate Roots**
2. **State Machine Pattern with Persistent Storage**
3. **Database-Centric Workflow Engine** 
4. **File-Based Workflow State with Git Integration**

## Decision Outcome

Chosen option: "State Machine Pattern with Persistent Storage", because it provides the optimal balance of simplicity, performance, and reliability for CLI-based workflows. State machines naturally model workflow progression with clear state transitions, while persistent storage ensures durability and enables audit trails without the complexity of event sourcing.

### Consequences

**Good:**
- Clear state transition semantics enable predictable workflow behavior and easy debugging
- Persistent storage provides durability and crash recovery without complex event replay logic  
- State machine pattern naturally supports conditional branching and parallel workflow paths
- Simple model enables fast CLI operations and responsive user interactions

**Bad:**
- State snapshots consume more storage compared to event-only approaches
- Limited ability to reconstruct historical workflow decisions without additional event logging
- Potential for state inconsistencies if concurrent updates are not properly synchronized

### Confirmation

Integration tests validating workflow state persistence across CLI restart scenarios, performance benchmarks demonstrating <100ms state update latency, and audit trail verification showing complete workflow progression history.

## Pros and Cons of the Options

### Event Sourcing with Aggregate Roots

An event-driven architecture that stores all workflow changes as immutable events, with current state derived by replaying events from an aggregate root.

**Good:**
- Complete audit trail of all workflow decisions and state changes
- Natural support for workflow replay and debugging scenarios
- Excellent consistency guarantees and concurrent update handling
- Can reconstruct any historical workflow state from event history

**Neutral:**
- Complex implementation requiring deep understanding of event sourcing patterns

**Bad:**
- High implementation complexity and steep learning curve for development team
- Performance overhead from event replay for current state reconstruction
- Complex schema evolution and event versioning requirements

### State Machine Pattern with Persistent Storage

A finite state machine that models workflow states and transitions, with current workflow state persisted to durable storage for crash recovery.

**Good:**
- Simple, well-understood pattern with clear state transition semantics
- Fast state queries and updates suitable for responsive CLI interactions
- Natural support for workflow validation and conditional state transitions
- Easy to implement rollback through state snapshots and checkpointing

**Bad:**
- Limited audit trail without additional event logging infrastructure
- Risk of state inconsistencies under concurrent access scenarios
- Storage overhead from maintaining complete state snapshots

### Database-Centric Workflow Engine

A workflow system built around a relational database that manages workflow state, transitions, and metadata through SQL operations and transactions.

**Good:**
- Robust transactional guarantees and concurrent access handling
- Rich query capabilities for workflow analytics and reporting
- Excellent integration with external business systems and APIs
- Scalable architecture suitable for enterprise multi-user scenarios

**Bad:**
- Complex deployment requiring database management and migration handling
- Performance overhead and network latency for simple CLI state operations
- Over-engineered solution for single-user CLI workflow scenarios

### File-Based Workflow State with Git Integration

A file-based system that stores workflow state in version-controlled files, leveraging Git for versioning, branching, and conflict resolution.

**Good:**
- Natural integration with development workflows and version control
- Human-readable state files enable easy debugging and manual intervention
- Built-in versioning and branching capabilities through Git primitives
- Zero external dependencies suitable for simple deployment scenarios

**Bad:**
- Poor concurrent access handling and risk of merge conflicts
- Limited transactional guarantees and potential for corrupted state files
- Complex to implement atomic updates and rollback across multiple state files

## More Information

- Related ADRs: [ADR-0009](ADR-0009-agent-orchestration.md), [ADR-0012](ADR-0012-integration-architecture.md)
- Implementation Location: `internal/workflow/`
- State Management Documentation: 12-stage engineering workflow specification
- Follow-ups: Workflow analytics and reporting, external state synchronization patterns
