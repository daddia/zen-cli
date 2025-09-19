---
status: Accepted
date: 2025-09-19
decision-makers: Development Team, Architecture Team
consulted: Template Engine Team, Infrastructure Team
informed: Engineering Leadership, Security Team
---

# ADR-0025 - Manifest-Only Asset Architecture

## Context and Problem Statement

The initial asset management design (ZEN-007) planned to sync and cache all assets locally to `~/.zen/cache/assets/` for offline access. However, this approach has several drawbacks:

- **Storage Overhead**: Full asset caching consumes significant local disk space
- **Sync Complexity**: Managing TTL, invalidation, and synchronization for all assets
- **Network Efficiency**: Downloading unused assets wastes bandwidth
- **Cache Management**: Complex cleanup and maintenance of cached content

The system needs to balance performance, storage efficiency, and simplicity while maintaining secure access to private assets and providing good user experience.

## Decision Drivers

- **Storage Efficiency**: Minimize local disk usage for asset management
- **Network Optimization**: Fetch only assets that are actually used
- **Simplicity**: Reduce complexity of cache management and synchronization
- **Performance**: Maintain fast asset discovery and reasonable fetch times
- **Security**: Preserve secure access to private repositories
- **User Experience**: Provide responsive asset listing and usage

## Decision Outcome

Chosen option: "Manifest-only synchronization with dynamic asset fetching", because it optimizes storage usage, reduces synchronization complexity, and provides better network efficiency while maintaining all security and performance requirements.

### Architecture Changes

**Manifest Storage**:
- Sync only `manifest.yaml` to `.zen/assets/` directory
- Manifest contains metadata for all available assets
- Local manifest enables fast asset discovery and listing

**Dynamic Asset Access**:
- Fetch assets on-demand from remote repository
- Session-based caching for fetched assets (temporary storage)
- Use Git CLI for authenticated remote access

**Performance Targets**:
- Manifest access: <100ms (local file system)
- Dynamic asset fetch: <5s (network + Git operations) - first access
- Cached asset access: <50ms (local cache) - subsequent access
- Asset listing: <50ms (manifest parsing)

**Session-Based Caching Strategy**:
- Assets cached temporarily during CLI session for reuse
- Cache TTL: Session lifetime (cleared on CLI exit or timeout)
- Cache location: `~/.zen/cache/assets/` using generic cache infrastructure
- Cache key: Asset ID from manifest for consistent lookup

### Implementation Details

```
.zen/
├── assets/
│   ├── manifest.yaml          # Synced from remote repository
│   └── .git-credentials       # Authentication cache
└── cache/
    └── assets/                # Session-based asset cache
        ├── templates/         # Cached template files
        ├── prompts/           # Cached prompt files
        └── .cache-index       # Cache metadata and TTL
```

**Manifest Structure**:
```yaml
version: "1.0"
repository: "https://github.com/org/zen-assets"
last_sync: "2025-09-19T10:30:00Z"
assets:
  templates:
    - id: "task-template"
      path: "templates/task.md"
      type: "template"
      description: "Standard task template"
    - id: "adr-template"  
      path: "templates/adr.md"
      type: "template"
      description: "Architecture Decision Record template"
  prompts:
    - id: "code-review"
      path: "prompts/code-review.txt"
      type: "prompt"
      description: "Code review assistant prompt"
```

### Consequences

**Good:**

- **Storage Efficiency**: Minimal persistent storage (manifest only)
- **Network Optimization**: Only download assets when needed, with session reuse
- **Simplified Sync**: Single manifest file synchronization
- **Performance Balance**: Fast access for repeated use within session
- **Reduced Complexity**: Simple session-based cache vs complex persistent cache

**Bad:**

- **Network Dependency**: Requires network access for first asset fetch
- **Session Boundary**: Cache cleared between sessions, requiring refetch
- **Temporary Storage**: Session cache uses disk space during active use

**Neutral:**

- **Security Model**: Same authentication requirements as full caching approach
- **Git Integration**: Still uses Git CLI wrapper per ADR-0023

### Migration Impact

**From Previous Design**:
- Update ZEN-007 acceptance criteria: manifest sync vs full asset sync
- Update ZEN-008 commands: `zen assets list` reads from local manifest
- Update Template Engine integration: dynamic fetching vs cached access
- Remove asset cache management complexity from generic cache system

**Implementation Changes**:
- Asset Client: Implement manifest sync + dynamic fetch
- Commands: Update `zen assets sync` to sync manifest only
- Template Engine: Add dynamic asset resolution
- Error Handling: Add network failure graceful degradation

## Pros and Cons of the Options

### Manifest-Only Architecture (Chosen)

**Good:**
- Minimal storage footprint with maximum efficiency
- Always up-to-date assets without cache invalidation complexity
- Simplified synchronization with single manifest file
- Network-efficient by fetching only used assets
- Eliminates complex cache management and TTL logic

**Neutral:**
- Maintains same authentication and security model
- Compatible with existing Git CLI wrapper approach

**Bad:**
- Network dependency for asset usage operations
- Higher latency for asset access (5s vs 100ms cached)
- Potential for repeated fetches of same assets

### Full Asset Caching (Previous Design)

**Good:**
- Fast offline access to all cached assets
- No network dependency after initial sync
- Predictable performance characteristics

**Bad:**
- Significant storage overhead for unused assets
- Complex cache management with TTL and invalidation
- Bandwidth waste downloading unused assets
- Stale cache issues requiring synchronization logic

### Hybrid Approach (Not Chosen)

Cache frequently used assets while fetching others dynamically.

**Good:**
- Balances storage efficiency with performance
- Reduces network dependency for common assets

**Bad:**
- Adds complexity of cache management decisions
- Difficult to predict which assets to cache
- Still requires complex cache invalidation logic
- Increases implementation and testing complexity

## Implementation Plan

### Phase 1: Manifest Infrastructure
1. Update Asset Client to sync manifest only
2. Implement manifest parsing and validation
3. Update `zen assets sync` command behavior
4. Update `zen assets list` to read from manifest

### Phase 2: Dynamic Fetching with Session Cache
1. Implement dynamic asset fetching via Git CLI
2. Integrate session-based caching using generic cache infrastructure
3. Add cache TTL management for session lifetime
4. Add network error handling and retry logic
5. Update Template Engine for dynamic asset resolution with cache lookup
6. Add performance monitoring and logging

### Phase 3: Integration & Testing
1. Update integration tests for new architecture
2. Performance testing for manifest access and asset fetch
3. Security testing for authentication flows
4. Cross-platform testing for Git CLI integration

## Related ADRs

- [ADR-0023: Git Library Selection](ADR-0023-git-library-selection.md) - Git CLI wrapper approach
- [ADR-0024: Generic Cache Architecture](ADR-0024-generic-cache-architecture.md) - Cache system (manifest caching)
- [ADR-0013: Template Engine Design](ADR-0013-template-engine.md) - Asset integration point
- [ADR-0020: Library-First Development](ADR-0020-library-first.md) - Simplicity principle

## References

- Asset Management Requirements: ZEN-006, ZEN-007, ZEN-008
- Performance Targets: <100ms manifest, <5s asset fetch
- Storage Efficiency: Minimal local footprint
- Network Optimization: Fetch-on-demand pattern
