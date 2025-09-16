# Integration Patterns

## Overview

This document describes the patterns and strategies used for integrating Zen CLI with external systems, APIs, and services. These patterns ensure reliable, secure, and maintainable integrations.

## Core Integration Patterns

### 1. API Gateway Pattern

Centralized entry point for all external API calls with cross-cutting concerns.

```mermaid
graph LR
    subgraph "Zen CLI"
        Client[Integration Client]
        Gateway[API Gateway]
    end
    
    subgraph "Gateway Features"
        Auth[Authentication]
        RateLimit[Rate Limiting]
        Retry[Retry Logic]
        Cache[Response Cache]
        Circuit[Circuit Breaker]
    end
    
    subgraph "External APIs"
        GitHub[GitHub API]
        Jira[Jira API]
        Slack[Slack API]
    end
    
    Client --> Gateway
    Gateway --> Auth
    Auth --> RateLimit
    RateLimit --> Retry
    Retry --> Cache
    Cache --> Circuit
    
    Circuit --> GitHub
    Circuit --> Jira
    Circuit --> Slack
```

**Implementation:**
```go
type APIGateway struct {
    rateLimiter *RateLimiter
    cache       *ResponseCache
    circuit     *CircuitBreaker
    retry       *RetryManager
}

func (g *APIGateway) Call(request Request) (Response, error) {
    // Apply cross-cutting concerns
    if err := g.rateLimiter.Check(request); err != nil {
        return nil, err
    }
    
    if cached := g.cache.Get(request); cached != nil {
        return cached, nil
    }
    
    response, err := g.circuit.Execute(func() (Response, error) {
        return g.retry.Execute(request)
    })
    
    if err == nil {
        g.cache.Set(request, response)
    }
    
    return response, err
}
```

### 2. Circuit Breaker Pattern

Prevents cascading failures when external services are unavailable.

```mermaid
stateDiagram-v2
    [*] --> Closed
    Closed --> Open: Failure Threshold
    Open --> HalfOpen: Timeout
    HalfOpen --> Closed: Success
    HalfOpen --> Open: Failure
    
    state Closed {
        [*] --> Monitoring
        Monitoring --> Monitoring: Success
        Monitoring --> Counting: Failure
        Counting --> Monitoring: Reset
    }
    
    state Open {
        [*] --> Rejecting
        Rejecting --> Rejecting: Fast Fail
    }
    
    state HalfOpen {
        [*] --> Testing
        Testing --> Success: API Success
        Testing --> Failure: API Failure
    }
```

**Configuration:**
```go
type CircuitBreaker struct {
    failureThreshold uint32
    successThreshold uint32
    timeout          time.Duration
    maxHalfOpen      uint32
}

// Per-service configuration
var breakers = map[string]*CircuitBreaker{
    "github": {
        failureThreshold: 5,
        successThreshold: 2,
        timeout:          30 * time.Second,
        maxHalfOpen:      3,
    },
    "jira": {
        failureThreshold: 3,
        successThreshold: 1,
        timeout:          60 * time.Second,
        maxHalfOpen:      1,
    },
}
```

### 3. Retry Pattern with Backoff

Intelligent retry logic with exponential backoff for transient failures.

```mermaid
graph TD
    Request[API Request] --> Try[Try Request]
    Try --> Success{Success?}
    Success -->|Yes| Return[Return Response]
    Success -->|No| Retryable{Retryable?}
    Retryable -->|No| Fail[Return Error]
    Retryable -->|Yes| Count{Max Retries?}
    Count -->|Yes| Fail
    Count -->|No| Backoff[Wait Backoff]
    Backoff --> Try
    
    subgraph "Backoff Strategy"
        B1[1s] --> B2[2s] --> B4[4s] --> B8[8s]
    end
```

**Implementation:**
```go
type RetryStrategy struct {
    MaxRetries  int
    InitialWait time.Duration
    MaxWait     time.Duration
    Multiplier  float64
    Jitter      float64
}

func (r *RetryStrategy) Execute(fn func() error) error {
    wait := r.InitialWait
    
    for i := 0; i <= r.MaxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }
        
        if !isRetryable(err) {
            return err
        }
        
        if i < r.MaxRetries {
            jitter := time.Duration(rand.Float64() * r.Jitter * float64(wait))
            time.Sleep(wait + jitter)
            wait = time.Duration(float64(wait) * r.Multiplier)
            if wait > r.MaxWait {
                wait = r.MaxWait
            }
        }
    }
    
    return ErrMaxRetriesExceeded
}
```

### 4. Rate Limiting Pattern

Respect API rate limits and prevent quota exhaustion.

```mermaid
graph LR
    subgraph "Rate Limiter"
        TokenBucket[Token Bucket]
        SlidingWindow[Sliding Window]
        Adaptive[Adaptive Limiting]
    end
    
    subgraph "Rate Limit Strategies"
        PerAPI[Per API]
        PerEndpoint[Per Endpoint]
        PerUser[Per User]
    end
    
    Request --> TokenBucket
    TokenBucket --> Check{Tokens Available?}
    Check -->|Yes| Allow[Allow Request]
    Check -->|No| Queue{Queue?}
    Queue -->|Yes| Wait[Wait for Token]
    Queue -->|No| Reject[Reject Request]
```

**Token Bucket Implementation:**
```go
type TokenBucket struct {
    capacity    int64
    tokens      int64
    refillRate  time.Duration
    lastRefill  time.Time
    mu          sync.Mutex
}

func (tb *TokenBucket) Allow() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()
    
    tb.refill()
    
    if tb.tokens > 0 {
        tb.tokens--
        return true
    }
    
    return false
}
```

### 5. Webhook Pattern

Asynchronous event handling from external systems.

```mermaid
sequenceDiagram
    participant External as External System
    participant Webhook as Webhook Endpoint
    participant Validator as Validator
    participant Queue as Event Queue
    participant Processor as Processor
    
    External->>Webhook: POST /webhooks/github
    Webhook->>Validator: Validate signature
    Validator-->>Webhook: Valid
    Webhook->>Queue: Enqueue event
    Webhook-->>External: 200 OK
    
    Queue->>Processor: Dequeue event
    Processor->>Processor: Process event
    Processor->>Queue: Mark complete
```

### 6. Polling Pattern

For systems without webhook support.

```mermaid
graph TD
    subgraph "Polling Strategy"
        Start[Start Polling] --> Interval{Check Interval}
        Interval --> Poll[Poll API]
        Poll --> Changes{Changes?}
        Changes -->|Yes| Process[Process Changes]
        Changes -->|No| Wait[Wait]
        Process --> UpdateCursor[Update Cursor]
        Wait --> Interval
        UpdateCursor --> Interval
    end
    
    subgraph "Optimization"
        Adaptive[Adaptive Polling]
        LongPoll[Long Polling]
        DeltaSync[Delta Sync]
    end
```

### 7. Bulk Operation Pattern

Optimize API calls through batching.

```go
type BulkProcessor struct {
    batchSize    int
    flushInterval time.Duration
    buffer       []Request
}

func (bp *BulkProcessor) Add(request Request) {
    bp.buffer = append(bp.buffer, request)
    
    if len(bp.buffer) >= bp.batchSize {
        bp.flush()
    }
}

func (bp *BulkProcessor) flush() {
    if len(bp.buffer) == 0 {
        return
    }
    
    batch := bp.buffer
    bp.buffer = nil
    
    // Process batch
    bp.processBatch(batch)
}
```

### 8. Cache-Aside Pattern

Intelligent caching for external API responses.

```mermaid
graph TD
    Request[Request] --> Cache{In Cache?}
    Cache -->|Yes| CheckTTL{Valid TTL?}
    CheckTTL -->|Yes| ReturnCache[Return Cached]
    CheckTTL -->|No| Fetch
    Cache -->|No| Fetch[Fetch from API]
    Fetch --> Store[Store in Cache]
    Store --> Return[Return Response]
```

**Cache Strategy:**
```go
type CacheEntry struct {
    Data      interface{}
    ExpiresAt time.Time
    ETag      string
}

type CacheStrategy struct {
    TTL           time.Duration
    MaxSize       int
    EvictionPolicy string // LRU, LFU, FIFO
}

// Different TTLs for different data types
var cacheConfigs = map[string]CacheStrategy{
    "user_info":     {TTL: 1 * time.Hour},
    "project_list":  {TTL: 5 * time.Minute},
    "issue_details": {TTL: 30 * time.Second},
}
```

### 9. Authentication Patterns

#### OAuth 2.0 Flow
```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Browser
    participant AuthServer
    participant API
    
    User->>CLI: zen auth login
    CLI->>Browser: Open auth URL
    Browser->>AuthServer: Authorize
    AuthServer->>Browser: Redirect with code
    Browser->>CLI: Return code
    CLI->>AuthServer: Exchange for token
    AuthServer->>CLI: Access token
    CLI->>API: API call with token
```

#### API Key Management
```go
type CredentialStore interface {
    Store(service string, credential Credential) error
    Retrieve(service string) (Credential, error)
    Delete(service string) error
}

// Platform-specific implementations
type KeychainStore struct{}  // macOS
type CredManStore struct{}   // Windows
type SecretService struct{} // Linux
```

### 10. Error Recovery Patterns

#### Graceful Degradation
```go
func GetUserInfo(id string) (*User, error) {
    // Try primary source
    user, err := api.GetUser(id)
    if err == nil {
        return user, nil
    }
    
    // Fallback to cache
    if cached := cache.GetUser(id); cached != nil {
        log.Warn("Using cached user data due to API error")
        return cached, nil
    }
    
    // Fallback to local data
    if local := db.GetUser(id); local != nil {
        log.Warn("Using local user data due to API error")
        return local, nil
    }
    
    return nil, err
}
```

## Integration Testing Patterns

### Mock Service Pattern
```go
type MockGitHubAPI struct {
    mock.Mock
}

func (m *MockGitHubAPI) CreatePullRequest(pr PullRequest) (*PullRequest, error) {
    args := m.Called(pr)
    return args.Get(0).(*PullRequest), args.Error(1)
}
```

### Contract Testing
```go
func TestGitHubContract(t *testing.T) {
    // Test against contract, not implementation
    client := NewGitHubClient()
    
    // Verify contract
    assert.Implements(t, (*VersionControl)(nil), client)
    
    // Test contract behavior
    pr, err := client.CreatePullRequest(testPR)
    assert.NoError(t, err)
    assert.NotEmpty(t, pr.ID)
}
```

## Best Practices

1. **Idempotency**: Make operations idempotent where possible
2. **Pagination**: Always handle paginated responses
3. **Timeouts**: Set appropriate timeouts for all external calls
4. **Monitoring**: Log and metric all integration points
5. **Documentation**: Keep API documentation current
6. **Versioning**: Handle API version changes gracefully
7. **Testing**: Comprehensive integration tests with mocks
8. **Security**: Never log sensitive data (tokens, keys)
9. **Error Context**: Provide detailed error context
10. **Graceful Degradation**: Have fallback strategies
