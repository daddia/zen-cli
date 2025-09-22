package resilience

import (
	"fmt"
	"sync"
	"time"
)

// CircuitBreaker implements the circuit breaker pattern for resilient external API calls
type CircuitBreaker struct {
	config        *CircuitBreakerConfig
	state         CircuitBreakerState
	failureCount  int
	lastFailure   time.Time
	halfOpenCalls int
	mu            sync.RWMutex
}

// CircuitBreakerInterface defines the circuit breaker interface
type CircuitBreakerInterface interface {
	// IsCallAllowed checks if a call is allowed
	IsCallAllowed() bool

	// Execute executes an operation with circuit breaker protection
	Execute(operation func() (interface{}, error)) (interface{}, error)

	// RecordSuccess records a successful operation
	RecordSuccess()

	// RecordFailure records a failed operation
	RecordFailure()

	// GetState returns the current circuit breaker state
	GetState() CircuitBreakerState

	// GetMetrics returns circuit breaker metrics
	GetMetrics() *CircuitBreakerMetrics

	// Reset resets the circuit breaker to closed state
	Reset()

	// IsOpen returns true if the circuit breaker is open
	IsOpen() bool

	// IsClosed returns true if the circuit breaker is closed
	IsClosed() bool

	// IsHalfOpen returns true if the circuit breaker is half-open
	IsHalfOpen() bool
}

// CircuitBreakerConfig contains circuit breaker configuration
type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold" yaml:"failure_threshold"`
	ResetTimeout     time.Duration `json:"reset_timeout" yaml:"reset_timeout"`
	HalfOpenMaxCalls int           `json:"half_open_max_calls" yaml:"half_open_max_calls"`
	Name             string        `json:"name" yaml:"name"`
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	CircuitBreakerClosed   CircuitBreakerState = "closed"
	CircuitBreakerOpen     CircuitBreakerState = "open"
	CircuitBreakerHalfOpen CircuitBreakerState = "half_open"
)

// CircuitBreakerMetrics contains circuit breaker performance metrics
type CircuitBreakerMetrics struct {
	State              CircuitBreakerState `json:"state"`
	FailureCount       int                 `json:"failure_count"`
	SuccessCount       int64               `json:"success_count"`
	TotalCalls         int64               `json:"total_calls"`
	LastFailure        *time.Time          `json:"last_failure,omitempty"`
	LastStateChange    time.Time           `json:"last_state_change"`
	TimeInCurrentState time.Duration       `json:"time_in_current_state"`
	TripsCount         int64               `json:"trips_count"`
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config.FailureThreshold <= 0 {
		config.FailureThreshold = 5
	}
	if config.ResetTimeout <= 0 {
		config.ResetTimeout = 30 * time.Second
	}
	if config.HalfOpenMaxCalls <= 0 {
		config.HalfOpenMaxCalls = 3
	}

	return &CircuitBreaker{
		config: config,
		state:  CircuitBreakerClosed,
	}
}

// IsCallAllowed checks if a call is allowed based on circuit breaker state
func (cb *CircuitBreaker) IsCallAllowed() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case CircuitBreakerClosed:
		return true

	case CircuitBreakerOpen:
		// Check if we should transition to half-open
		if time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
			// Transition to half-open will be handled by the next call
			return true
		}
		return false

	case CircuitBreakerHalfOpen:
		// Allow limited calls in half-open state
		return cb.halfOpenCalls < cb.config.HalfOpenMaxCalls

	default:
		return false
	}
}

// Execute executes an operation with circuit breaker protection
func (cb *CircuitBreaker) Execute(operation func() (interface{}, error)) (interface{}, error) {
	// Check if call is allowed
	if !cb.IsCallAllowed() {
		return nil, fmt.Errorf("circuit breaker is open")
	}

	// Handle state transitions
	cb.mu.Lock()
	if cb.state == CircuitBreakerOpen && time.Since(cb.lastFailure) >= cb.config.ResetTimeout {
		cb.state = CircuitBreakerHalfOpen
		cb.halfOpenCalls = 0
	}

	if cb.state == CircuitBreakerHalfOpen {
		cb.halfOpenCalls++
	}
	cb.mu.Unlock()

	// Execute operation
	result, err := operation()

	// Record result
	if err != nil {
		cb.RecordFailure()
		return nil, err
	} else {
		cb.RecordSuccess()
		return result, nil
	}
}

// RecordSuccess records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitBreakerClosed:
		// Reset failure count on success
		cb.failureCount = 0

	case CircuitBreakerHalfOpen:
		// If we've had enough successful calls, close the circuit
		if cb.halfOpenCalls >= cb.config.HalfOpenMaxCalls {
			cb.state = CircuitBreakerClosed
			cb.failureCount = 0
			cb.halfOpenCalls = 0
		}
	}
}

// RecordFailure records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailure = time.Now()

	switch cb.state {
	case CircuitBreakerClosed:
		if cb.failureCount >= cb.config.FailureThreshold {
			cb.state = CircuitBreakerOpen
		}

	case CircuitBreakerHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.state = CircuitBreakerOpen
		cb.halfOpenCalls = 0
	}
}

// GetState returns the current circuit breaker state
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() *CircuitBreakerMetrics {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	metrics := &CircuitBreakerMetrics{
		State:        cb.state,
		FailureCount: cb.failureCount,
	}

	if !cb.lastFailure.IsZero() {
		metrics.LastFailure = &cb.lastFailure
	}

	return metrics
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.state = CircuitBreakerClosed
	cb.failureCount = 0
	cb.halfOpenCalls = 0
	cb.lastFailure = time.Time{}
}

// IsOpen returns true if the circuit breaker is open
func (cb *CircuitBreaker) IsOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == CircuitBreakerOpen
}

// IsClosed returns true if the circuit breaker is closed
func (cb *CircuitBreaker) IsClosed() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == CircuitBreakerClosed
}

// IsHalfOpen returns true if the circuit breaker is half-open
func (cb *CircuitBreaker) IsHalfOpen() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state == CircuitBreakerHalfOpen
}
