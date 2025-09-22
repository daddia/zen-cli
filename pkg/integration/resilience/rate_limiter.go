package resilience

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting for external API calls
type RateLimiter struct {
	config     *RateLimiterConfig
	tokens     int64
	lastRefill time.Time
	mu         sync.Mutex
}

// RateLimiterInterface defines the rate limiter interface
type RateLimiterInterface interface {
	// Allow checks if a request is allowed
	Allow() bool

	// WaitForToken waits for a token to become available
	WaitForToken(ctx context.Context) error

	// GetTokens returns the current number of available tokens
	GetTokens() int64

	// GetMetrics returns rate limiter metrics
	GetMetrics() *RateLimiterMetrics

	// Reset resets the rate limiter
	Reset()

	// UpdateConfig updates the rate limiter configuration
	UpdateConfig(config *RateLimiterConfig)
}

// RateLimiterConfig contains rate limiter configuration
type RateLimiterConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute" yaml:"requests_per_minute"`
	BurstSize         int           `json:"burst_size" yaml:"burst_size"`
	RefillInterval    time.Duration `json:"refill_interval" yaml:"refill_interval"`
	Name              string        `json:"name" yaml:"name"`
}

// RateLimiterMetrics contains rate limiter performance metrics
type RateLimiterMetrics struct {
	RequestsPerMinute int       `json:"requests_per_minute"`
	BurstSize         int       `json:"burst_size"`
	AvailableTokens   int64     `json:"available_tokens"`
	LastRefill        time.Time `json:"last_refill"`
	TotalRequests     int64     `json:"total_requests"`
	AllowedRequests   int64     `json:"allowed_requests"`
	RejectedRequests  int64     `json:"rejected_requests"`
}

// ResilienceInterface combines circuit breaker and rate limiter interfaces
type ResilienceInterface interface {
	// Circuit breaker methods
	IsCallAllowed() bool
	Execute(operation func() (interface{}, error)) (interface{}, error)
	RecordSuccess()
	RecordFailure()
	GetState() CircuitBreakerState
	Reset()
	IsOpen() bool
	IsClosed() bool
	IsHalfOpen() bool

	// Rate limiter methods
	Allow() bool
	WaitForToken(ctx context.Context) error
	GetTokens() int64
	UpdateConfig(config *RateLimiterConfig)
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(config *RateLimiterConfig) *RateLimiter {
	if config.RequestsPerMinute <= 0 {
		config.RequestsPerMinute = 60 // Default 1 request per second
	}
	if config.BurstSize <= 0 {
		config.BurstSize = 10 // Default burst of 10
	}
	if config.RefillInterval <= 0 {
		config.RefillInterval = time.Second // Default refill every second
	}

	return &RateLimiter{
		config:     config,
		tokens:     int64(config.BurstSize), // Start with full bucket
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed (non-blocking)
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()

	if rl.tokens > 0 {
		rl.tokens--
		return true
	}

	return false
}

// WaitForToken waits for a token to become available (blocking)
func (rl *RateLimiter) WaitForToken(ctx context.Context) error {
	for {
		if rl.Allow() {
			return nil
		}

		// Calculate wait time until next token
		rl.mu.Lock()
		refillRate := float64(rl.config.RequestsPerMinute) / 60.0 // tokens per second
		waitTime := time.Duration(1.0/refillRate) * time.Second
		rl.mu.Unlock()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Continue loop to check again
		}
	}
}

// GetTokens returns the current number of available tokens
func (rl *RateLimiter) GetTokens() int64 {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.refillTokens()
	return rl.tokens
}

// GetMetrics returns rate limiter metrics
func (rl *RateLimiter) GetMetrics() *RateLimiterMetrics {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return &RateLimiterMetrics{
		RequestsPerMinute: rl.config.RequestsPerMinute,
		BurstSize:         rl.config.BurstSize,
		AvailableTokens:   rl.tokens,
		LastRefill:        rl.lastRefill,
		// Note: Total/Allowed/Rejected requests would need additional tracking
	}
}

// Reset resets the rate limiter
func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.tokens = int64(rl.config.BurstSize)
	rl.lastRefill = time.Now()
}

// UpdateConfig updates the rate limiter configuration
func (rl *RateLimiter) UpdateConfig(config *RateLimiterConfig) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.config = config
	// Adjust tokens if burst size changed
	if rl.tokens > int64(config.BurstSize) {
		rl.tokens = int64(config.BurstSize)
	}
}

// refillTokens adds tokens to the bucket based on elapsed time
func (rl *RateLimiter) refillTokens() {
	now := time.Now()
	elapsed := now.Sub(rl.lastRefill)

	// Calculate tokens to add based on elapsed time
	refillRate := float64(rl.config.RequestsPerMinute) / 60.0 // tokens per second
	tokensToAdd := int64(elapsed.Seconds() * refillRate)

	if tokensToAdd > 0 {
		rl.tokens = min(int64(rl.config.BurstSize), rl.tokens+tokensToAdd)
		rl.lastRefill = now
	}
}

// min returns the minimum of two int64 values
func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

// ResilienceManager combines circuit breaker and rate limiter functionality
type ResilienceManager struct {
	circuitBreaker *CircuitBreaker
	rateLimiter    *RateLimiter
	name           string
}

// NewResilienceManager creates a new resilience manager
func NewResilienceManager(name string, cbConfig *CircuitBreakerConfig, rlConfig *RateLimiterConfig) *ResilienceManager {
	return &ResilienceManager{
		circuitBreaker: NewCircuitBreaker(cbConfig),
		rateLimiter:    NewRateLimiter(rlConfig),
		name:           name,
	}
}

// ExecuteWithResilience executes an operation with both circuit breaker and rate limiting
func (rm *ResilienceManager) ExecuteWithResilience(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	// Check rate limit first
	if !rm.rateLimiter.Allow() {
		return nil, fmt.Errorf("rate limit exceeded for %s", rm.name)
	}

	// Execute with circuit breaker protection
	return rm.circuitBreaker.Execute(operation)
}

// ExecuteWithWait executes an operation, waiting for rate limit if necessary
func (rm *ResilienceManager) ExecuteWithWait(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {
	// Wait for rate limit token
	if err := rm.rateLimiter.WaitForToken(ctx); err != nil {
		return nil, fmt.Errorf("failed to acquire rate limit token: %w", err)
	}

	// Execute with circuit breaker protection
	return rm.circuitBreaker.Execute(operation)
}

// GetCircuitBreakerMetrics returns circuit breaker metrics
func (rm *ResilienceManager) GetCircuitBreakerMetrics() *CircuitBreakerMetrics {
	return rm.circuitBreaker.GetMetrics()
}

// GetRateLimiterMetrics returns rate limiter metrics
func (rm *ResilienceManager) GetRateLimiterMetrics() *RateLimiterMetrics {
	return rm.rateLimiter.GetMetrics()
}

// IsHealthy returns true if both circuit breaker and rate limiter are healthy
func (rm *ResilienceManager) IsHealthy() bool {
	return rm.circuitBreaker.IsClosed() && rm.rateLimiter.GetTokens() > 0
}

// Reset resets both circuit breaker and rate limiter
func (rm *ResilienceManager) Reset() {
	rm.circuitBreaker.Reset()
	rm.rateLimiter.Reset()
}
