---
status: Proposed
date: 2025-09-16
decision-makers: Development Team, Architecture Team
consulted: Template Engine Team, Infrastructure Team
informed: Engineering Leadership, Security Team
---

# ADR-0023 - Git Library Selection for Asset Client

## Context and Problem Statement

The Zen CLI requires a Git-based asset client (ZEN-007) to authenticate with Git providers, fetch asset manifests, and dynamically access private repository assets.

The asset client must support GitHub and GitLab authentication, provide fast manifest synchronization, and efficiently access assets on-demand from remote repositories. The implementation needs to balance performance, security, maintainability, and feature completeness while integrating seamlessly with the existing Zen CLI architecture.

Key requirements include:

- Private repository authentication and access
- Manifest synchronization to `.zen/assets/` directory
- Dynamic asset fetching without local storage
- Integration with Template Engine interface
- Support for GitHub PAT and GitLab tokens
- Minimal external dependencies per Library-First principles
- Performance targets: <100ms manifest access, <5s asset fetch

## Decision Drivers

- **Performance Requirements**: Manifest access <100ms, dynamic asset fetch <5s
- **Authentication Support**: GitHub Personal Access Tokens, GitLab Project Access Tokens
- **External Dependencies**: Minimize dependencies per Library-First development approach (ADR-0020)
- **Feature Completeness**: Support for clone, pull, fetch operations with authentication
- **Security**: Secure credential handling and storage per Security Model (ADR-0015)
- **Maintainability**: Clear error handling, debugging support, and code maintainability
- **Integration**: Seamless integration with existing Factory pattern (ADR-0006) and Template Engine (ADR-0013)
- **Platform Support**: Cross-platform compatibility (macOS, Linux, Windows)

## Considered Options

- **Pure Go git library (go-git/go-git)**
- **Git CLI wrapper with subprocess execution**
- **libgit2 Go bindings (git2go)**
- **Do nothing - direct file system access only**

## Decision Outcome

Chosen option: "Git CLI wrapper with subprocess execution", because it provides the most reliable and feature-complete Git implementation with proven authentication patterns from GitHub CLI reference, maintains compatibility with all Git features, and aligns with the Library-First principle by leveraging the battle-tested Git CLI rather than reimplementing Git operations.

### Consequences

**Good:**

- Leverages full Git CLI feature set and receives automatic updates
- Proven authentication patterns from GitHub CLI provide security best practices
- Subprocess isolation provides better error handling and debugging
- Eliminates risk of incomplete Git feature support
- Reduces maintenance burden compared to pure Go implementations

**Bad:**

- Introduces external dependency on Git CLI installation
- Subprocess overhead may impact performance (mitigated by caching)
- Requires parsing CLI output and handling process errors

### Confirmation

Implementation will be validated through:

- Performance benchmarks meeting <100ms manifest access and <5s dynamic asset fetch targets
- Authentication testing with GitHub and GitLab using PAT/Project tokens
- Cross-platform testing on macOS, Linux, and Windows
- Integration testing with Template Engine interface
- Error handling validation for network failures and authentication errors

## Pros and Cons of the Options

### Pure Go git library (go-git/go-git)

A native Go implementation of Git protocol and operations that provides Git functionality without requiring the Git CLI to be installed on the system.

**Good:**

- Eliminates external Git CLI dependency, simplifying deployment and distribution
- Provides native Go API integration with type-safe interfaces
- Potentially better performance for simple operations due to direct library calls
- Self-contained solution that works in any Go environment

**Neutral:**

- Active development and community support with regular updates
- Moderate learning curve for Git-specific API patterns

**Bad:**

- Incomplete feature parity with Git CLI, missing some advanced Git operations
- Authentication implementation complexity requiring custom OAuth and token handling
- Potential lag behind official Git feature updates and security patches
- Memory usage concerns for large repositories due to in-memory object handling

### Git CLI wrapper with subprocess execution

A wrapper approach that executes Git commands through the system's installed Git CLI, parsing outputs and handling process communication for Git operations.

**Good:**

- Complete Git feature set with automatic updates as Git CLI evolves
- Proven authentication patterns from GitHub CLI reference implementation
- Subprocess isolation provides better error handling and debugging capabilities
- Leverages battle-tested Git implementation with full ecosystem compatibility
- Familiar Git semantics for debugging and troubleshooting operations

**Neutral:**

- Moderate implementation complexity for subprocess handling and output parsing

**Bad:**

- External dependency on Git CLI installation across all target platforms
- Subprocess overhead and output parsing requirements impacting performance
- Potential platform-specific behavior differences in Git CLI implementations

### libgit2 Go bindings (git2go)

Go bindings for the libgit2 C library, providing comprehensive Git functionality through CGO bindings to the native libgit2 implementation.

**Good:**

- Native performance through optimized C library implementation
- Comprehensive Git feature support matching Git CLI capabilities
- No external Git CLI dependency, using compiled C library instead
- Mature library with extensive Git operation support

**Neutral:**

- Established library with community support and documentation
- Requires CGO knowledge for advanced customization

**Bad:**

- CGO dependency complicates cross-compilation and deployment scenarios
- C library integration complexity and debugging challenges across platforms
- Potential memory management issues requiring careful resource cleanup
- Authentication implementation complexity similar to pure Go solutions
- Violates Library-First principle (ADR-0020) preference for Go-native solutions

### Do nothing - direct file system access only

Access assets directly from the local file system without any Git integration, requiring manual asset management and distribution.

**Good:**

- Simplest implementation with no Git dependencies or complexity
- Fastest performance for local access with direct file system operations
- No external dependencies or network requirements

**Neutral:**

- Eliminates authentication complexity by avoiding remote repositories

**Bad:**

- No private repository support, failing core ZEN-007 requirements
- No synchronization capabilities for asset updates and distribution
- Manual asset management burden for users and deployment processes
- Incompatible with ZEN-007 requirements for Git-based asset client

## More Information

**Related ADRs:**

- [ADR-0006](ADR-0006-factory-pattern.md): Factory Pattern Implementation (dependency injection)
- [ADR-0013](ADR-0013-template-engine.md): Template Engine Design (asset client integration)
- [ADR-0015](ADR-0015-security-model.md): Security Model Implementation (credential handling)
- [ADR-0020](ADR-0020-library-first.md): Library-First Development Approach (dependency philosophy)

**Implementation References:**

- GitHub CLI authentication patterns: `/zen-reference/gh-cli/git/client.go`
- Credential helper implementation: `AuthenticatedCommand` pattern
- Error handling patterns: `GitError` type and exit code handling

**Performance Considerations:**

- Subprocess overhead mitigated by manifest caching and selective asset fetching
- Git CLI optimization flags: `--depth=1` for shallow clones
- Credential caching to minimize authentication overhead

**Security Considerations:**

- Credential helper pattern for secure token handling
- Environment variable support for CI/CD workflows
- No credential logging or exposure in error messages

Note: This decision focuses on Git operations only. Manifest storage and dynamic asset fetching implementation are separate concerns addressed in the asset client architecture.
