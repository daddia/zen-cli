# ADR-0024: File-Based Cache Architecture

## Status
Accepted

## Context

Zen CLI requires caching capabilities across multiple components:
- Asset management needs to cache templates, prompts, and manifests
- Configuration system could benefit from caching expensive operations
- Git operations could cache repository metadata and file contents
- Future features will likely need similar caching capabilities

Initially, a cache implementation was created specifically for the assets package, but this approach led to:
- Code duplication when other components needed caching
- Inconsistent caching behavior across different features
- Difficulty in testing and mocking cache operations
- Tight coupling between cache logic and asset-specific types

## Decision

We will implement a type-safe, file-based caching system in `pkg/cache/` that can be reused across all Zen components.

### Architecture Components

1. **Type-Safe Cache Interface** (`cache.Manager[T]`)
   - Type-safe operations using Go generics
   - Standard operations: Get, Put, Delete, Clear, GetInfo, Cleanup
   - Pluggable serialization strategies

2. **File-Based Implementation** (`cache.FileManager[T]`)
   - Persistent file-system storage
   - LRU eviction policy
   - TTL-based expiration
   - Thread-safe concurrent access
   - Automatic index management

3. **Serialization Strategy** (`cache.Serializer[T]`)
   - JSON serialization for complex types
   - String serialization for text data
   - Extensible for custom serialization needs

4. **Adapter Pattern** for Integration
   - Asset-specific adapter (`AssetCacheManager`)
   - Error code translation
   - Interface compatibility with existing code

5. **Factory Integration**
   - Available through `cmdutil.Factory.Cache(basePath)`
   - Easy access for all commands
   - Consistent configuration and setup

### Implementation Details

```go
// Generic cache interface
type Manager[T any] interface {
    Get(ctx context.Context, key string) (*Entry[T], error)
    Put(ctx context.Context, key string, data T, opts PutOptions) error
    Delete(ctx context.Context, key string) error
    Clear(ctx context.Context) error
    GetInfo(ctx context.Context) (*Info, error)
    Cleanup(ctx context.Context) error
    Close() error
}

// Usage in commands
func NewCmdExample(f *cmdutil.Factory) *cobra.Command {
    return &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            cache := f.Cache("~/.zen/cache/example")
            defer cache.Close()
            
            // Use cache for string data
            err := cache.Put(ctx, "key", "value", cache.PutOptions{TTL: time.Hour})
            return err
        },
    }
}

// Asset-specific adapter
type AssetCacheManager struct {
    cache cache.Manager[AssetContent]
}

func NewAssetCacheManager(basePath string, sizeMB int64, ttl time.Duration, logger logging.Logger) *AssetCacheManager {
    config := cache.Config{BasePath: basePath, SizeLimitMB: sizeMB, DefaultTTL: ttl}
    serializer := NewAssetContentSerializer()
    genericCache := cache.NewManager(config, logger, serializer)
    
    return &AssetCacheManager{cache: genericCache}
}
```

## Consequences

### Positive

1. **Reusability**: Any component can use caching with minimal setup
2. **Type Safety**: Go generics prevent runtime type errors
3. **Consistency**: Same caching behavior across all features
4. **Testability**: Generic cache can be thoroughly tested once
5. **Performance**: Optimized file-based storage with LRU and TTL
6. **Maintainability**: Single cache implementation to maintain
7. **Extensibility**: Easy to add new serialization strategies

### Negative

1. **Complexity**: Generic types add some complexity to the codebase
2. **Learning Curve**: Developers need to understand generics and adapters
3. **Migration**: Existing asset cache needed to be migrated
4. **Testing Overhead**: More comprehensive testing required for generic components

### Risks and Mitigations

**Risk**: Generic cache might not fit all use cases
**Mitigation**: Adapter pattern allows customization while reusing core functionality

**Risk**: Performance overhead from generics
**Mitigation**: Go generics compile to efficient code with minimal overhead

**Risk**: Increased complexity for simple use cases
**Mitigation**: Factory provides simple `f.Cache(path)` for common string caching

## Implementation Notes

### Directory Structure
```
pkg/cache/
├── cache.go                 # Generic interfaces and types
├── file.go                  # File-based implementation
├── serializer.go            # Serialization strategies
├── cache_test.go           # Interface tests
├── file_test.go            # Implementation tests
└── serializer_test.go      # Serialization tests
```

### Testing Strategy
- Generic cache tests achieve 85.9% coverage
- Both unit and integration tests included
- Error conditions and edge cases covered
- Thread safety validated with race detection
- Performance characteristics documented

### Usage Guidelines
1. Use `f.Cache(basePath)` for simple string caching in commands
2. Create specialized adapters for complex domain types
3. Always specify appropriate TTL for cached data
4. Use different base paths for different cache purposes
5. Handle cache errors gracefully with fallback strategies

## Related ADRs

- [ADR-0006: Factory Pattern](ADR-0006-factory-pattern.md) - Dependency injection strategy
- [ADR-0020: Library-First Development](ADR-0020-library-first.md) - Reusable component strategy
- [ADR-0003: Project Structure](ADR-0003-project-structure.md) - Package organization

## References

- Go Generics Best Practices
- Cache Implementation Patterns
- Adapter Pattern Documentation
- Zen CLI Design Guidelines
