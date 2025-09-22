package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/integration/plugin"
	"github.com/daddia/zen/pkg/integration/resilience"
)

// OperationOrchestrator coordinates plugin operations with consistent error handling and retry logic
type OperationOrchestrator struct {
	logger          logging.Logger
	circuitBreakers map[string]*resilience.CircuitBreaker
	rateLimiters    map[string]*resilience.RateLimiter
	transactionMgr  TransactionManagerInterface
	compensationMgr CompensationInterface
	metrics         *OrchestratorMetrics
	mu              sync.RWMutex
}

// OperationOrchestratorInterface defines the orchestrator interface
type OperationOrchestratorInterface interface {
	// ExecuteOperation executes a plugin operation with resilience patterns
	ExecuteOperation(ctx context.Context, operation *Operation) (*OperationResult, error)

	// ExecuteTransaction executes multiple operations as a transaction
	ExecuteTransaction(ctx context.Context, transaction *Transaction) (*TransactionResult, error)

	// GetOperationMetrics returns operation metrics
	GetOperationMetrics() *OrchestratorMetrics

	// RegisterPlugin registers a plugin with the orchestrator
	RegisterPlugin(pluginName string, plugin plugin.IntegrationPluginInterface) error

	// UnregisterPlugin unregisters a plugin
	UnregisterPlugin(pluginName string) error
}

// TransactionManagerInterface defines transaction management interface
type TransactionManagerInterface interface {
	// BeginTransaction starts a new transaction
	BeginTransaction(ctx context.Context, transactionID string) error

	// CommitTransaction commits a transaction
	CommitTransaction(ctx context.Context, transactionID string) error

	// RollbackTransaction rolls back a transaction
	RollbackTransaction(ctx context.Context, transactionID string) error

	// AddOperation adds an operation to a transaction
	AddOperation(transactionID string, operation *Operation) error
}

// CompensationInterface defines compensation logic interface
type CompensationInterface interface {
	// RegisterCompensation registers a compensation action
	RegisterCompensation(operationID string, compensationFunc CompensationFunc) error

	// ExecuteCompensation executes compensation for a failed operation
	ExecuteCompensation(ctx context.Context, operationID string) error

	// ClearCompensation clears compensation for a successful operation
	ClearCompensation(operationID string) error
}

// Operation represents a plugin operation
type Operation struct {
	ID          string                 `json:"id"`
	PluginName  string                 `json:"plugin_name"`
	Type        OperationType          `json:"type"`
	ExternalID  string                 `json:"external_id,omitempty"`
	TaskData    *plugin.TaskData       `json:"task_data,omitempty"`
	Query       *plugin.SearchQuery    `json:"query,omitempty"`
	Options     map[string]interface{} `json:"options"`
	Timeout     time.Duration          `json:"timeout"`
	RetryPolicy *RetryPolicy           `json:"retry_policy,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OperationType represents the type of operation
type OperationType string

const (
	OperationTypeFetch  OperationType = "fetch"
	OperationTypeCreate OperationType = "create"
	OperationTypeUpdate OperationType = "update"
	OperationTypeDelete OperationType = "delete"
	OperationTypeSearch OperationType = "search"
	OperationTypeSync   OperationType = "sync"
)

// OperationResult represents the result of an operation
type OperationResult struct {
	Success     bool                   `json:"success"`
	OperationID string                 `json:"operation_id"`
	PluginName  string                 `json:"plugin_name"`
	Type        OperationType          `json:"type"`
	Data        interface{}            `json:"data,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ErrorCode   string                 `json:"error_code,omitempty"`
	Retryable   bool                   `json:"retryable,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Timestamp   time.Time              `json:"timestamp"`
	RetryCount  int                    `json:"retry_count"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// Transaction represents a group of operations executed together
type Transaction struct {
	ID             string                 `json:"id"`
	Operations     []*Operation           `json:"operations"`
	Timeout        time.Duration          `json:"timeout"`
	IsolationLevel string                 `json:"isolation_level"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// TransactionResult represents the result of a transaction
type TransactionResult struct {
	Success       bool               `json:"success"`
	TransactionID string             `json:"transaction_id"`
	Results       []*OperationResult `json:"results"`
	Duration      time.Duration      `json:"duration"`
	Timestamp     time.Time          `json:"timestamp"`
	Error         string             `json:"error,omitempty"`
	RolledBack    bool               `json:"rolled_back"`
}

// RetryPolicy defines retry behavior for operations
type RetryPolicy struct {
	MaxRetries      int           `json:"max_retries"`
	BaseDelay       time.Duration `json:"base_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffStrategy string        `json:"backoff_strategy"` // exponential, linear, fixed
	RetryableErrors []string      `json:"retryable_errors"`
}

// CompensationFunc represents a compensation function
type CompensationFunc func(ctx context.Context, operationData interface{}) error

// OrchestratorMetrics contains orchestrator performance metrics
type OrchestratorMetrics struct {
	TotalOperations     int64         `json:"total_operations"`
	SuccessfulOps       int64         `json:"successful_operations"`
	FailedOps           int64         `json:"failed_operations"`
	RetriedOps          int64         `json:"retried_operations"`
	AverageLatency      time.Duration `json:"average_latency"`
	CircuitBreakerTrips int64         `json:"circuit_breaker_trips"`
	RateLimitHits       int64         `json:"rate_limit_hits"`
	LastOperationTime   time.Time     `json:"last_operation_time"`
	mu                  sync.RWMutex
}

// NewOperationOrchestrator creates a new operation orchestrator
func NewOperationOrchestrator(logger logging.Logger) *OperationOrchestrator {
	return &OperationOrchestrator{
		logger:          logger,
		circuitBreakers: make(map[string]*resilience.CircuitBreaker),
		rateLimiters:    make(map[string]*resilience.RateLimiter),
		transactionMgr:  NewTransactionManager(logger),
		compensationMgr: NewCompensationManager(logger),
		metrics:         &OrchestratorMetrics{},
	}
}

// ExecuteOperation executes a plugin operation with resilience patterns
func (o *OperationOrchestrator) ExecuteOperation(ctx context.Context, operation *Operation) (*OperationResult, error) {
	start := time.Now()

	result := &OperationResult{
		OperationID: operation.ID,
		PluginName:  operation.PluginName,
		Type:        operation.Type,
		Timestamp:   time.Now(),
	}

	o.logger.Debug("executing operation",
		"id", operation.ID,
		"plugin", operation.PluginName,
		"type", operation.Type)

	// Check circuit breaker
	if !o.isCircuitBreakerClosed(operation.PluginName) {
		err := fmt.Errorf("circuit breaker is open for plugin: %s", operation.PluginName)
		result.Success = false
		result.Error = err.Error()
		result.ErrorCode = "CIRCUIT_BREAKER_OPEN"
		result.Duration = time.Since(start)
		return result, err
	}

	// Check rate limiter
	if !o.checkRateLimit(operation.PluginName) {
		err := fmt.Errorf("rate limit exceeded for plugin: %s", operation.PluginName)
		result.Success = false
		result.Error = err.Error()
		result.ErrorCode = "RATE_LIMITED"
		result.Duration = time.Since(start)
		o.updateMetrics(false, true, false)
		return result, err
	}

	// Execute operation with retry logic
	retryPolicy := operation.RetryPolicy
	if retryPolicy == nil {
		retryPolicy = &RetryPolicy{
			MaxRetries:      3,
			BaseDelay:       time.Second,
			MaxDelay:        30 * time.Second,
			BackoffStrategy: "exponential",
		}
	}

	var lastError error
	for attempt := 0; attempt <= retryPolicy.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			result.Success = false
			result.Error = ctx.Err().Error()
			result.ErrorCode = "CONTEXT_CANCELLED"
			result.Duration = time.Since(start)
			return result, ctx.Err()
		default:
		}

		// Execute the actual operation
		operationResult, err := o.executePluginOperation(ctx, operation)
		if err == nil {
			// Success
			result.Success = true
			result.Data = operationResult
			result.Duration = time.Since(start)
			result.RetryCount = attempt

			o.recordSuccess(operation.PluginName)
			o.updateMetrics(true, false, attempt > 0)

			return result, nil
		}

		lastError = err
		result.RetryCount = attempt

		// Check if error is retryable
		if !o.isRetryableError(err, retryPolicy) || attempt == retryPolicy.MaxRetries {
			break
		}

		// Calculate backoff delay
		delay := o.calculateBackoffDelay(attempt, retryPolicy)
		o.logger.Debug("retrying operation after delay",
			"operation_id", operation.ID,
			"attempt", attempt+1,
			"delay", delay,
			"error", err)

		// Wait before retry
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			result.Success = false
			result.Error = ctx.Err().Error()
			result.ErrorCode = "CONTEXT_CANCELLED"
			result.Duration = time.Since(start)
			return result, ctx.Err()
		case <-timer.C:
			// Continue to next attempt
		}
	}

	// All retries failed
	result.Success = false
	result.Error = lastError.Error()
	result.Retryable = o.isRetryableError(lastError, retryPolicy)
	result.Duration = time.Since(start)

	o.recordFailure(operation.PluginName)
	o.updateMetrics(false, false, result.RetryCount > 0)

	return result, lastError
}

// executePluginOperation executes the actual plugin operation
func (o *OperationOrchestrator) executePluginOperation(ctx context.Context, operation *Operation) (interface{}, error) {
	// This is a placeholder implementation
	// In the actual implementation, this would:
	// 1. Get the plugin instance from the factory
	// 2. Call the appropriate plugin method based on operation type
	// 3. Handle the response and errors

	switch operation.Type {
	case OperationTypeFetch:
		// Would call plugin.FetchTask()
		return map[string]interface{}{
			"id":    operation.ExternalID,
			"title": "Fetched Task",
			"type":  "task",
		}, nil

	case OperationTypeCreate:
		// Would call plugin.CreateTask()
		return map[string]interface{}{
			"id":    "NEW-123",
			"title": operation.TaskData.Title,
			"type":  operation.TaskData.Type,
		}, nil

	case OperationTypeUpdate:
		// Would call plugin.UpdateTask()
		return map[string]interface{}{
			"id":    operation.ExternalID,
			"title": operation.TaskData.Title,
			"type":  operation.TaskData.Type,
		}, nil

	case OperationTypeDelete:
		// Would call plugin.DeleteTask()
		return nil, nil

	case OperationTypeSearch:
		// Would call plugin.SearchTasks()
		return []*plugin.TaskData{}, nil

	case OperationTypeSync:
		// Would call plugin.SyncTask()
		return &plugin.SyncResult{
			Success: true,
			TaskID:  operation.TaskData.ID,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported operation type: %s", operation.Type)
	}
}

// ExecuteTransaction executes multiple operations as a transaction
func (o *OperationOrchestrator) ExecuteTransaction(ctx context.Context, transaction *Transaction) (*TransactionResult, error) {
	start := time.Now()

	result := &TransactionResult{
		TransactionID: transaction.ID,
		Results:       make([]*OperationResult, 0, len(transaction.Operations)),
		Timestamp:     time.Now(),
	}

	o.logger.Debug("executing transaction",
		"id", transaction.ID,
		"operations", len(transaction.Operations))

	// Begin transaction
	if err := o.transactionMgr.BeginTransaction(ctx, transaction.ID); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to begin transaction: %v", err)
		result.Duration = time.Since(start)
		return result, err
	}

	// Execute operations sequentially
	var compensationOps []string
	for _, op := range transaction.Operations {
		opResult, err := o.ExecuteOperation(ctx, op)
		result.Results = append(result.Results, opResult)

		if err != nil {
			// Operation failed - rollback transaction
			o.logger.Warn("operation failed in transaction",
				"transaction_id", transaction.ID,
				"operation_id", op.ID,
				"error", err)

			// Execute compensations for completed operations
			for i := len(compensationOps) - 1; i >= 0; i-- {
				if compErr := o.compensationMgr.ExecuteCompensation(ctx, compensationOps[i]); compErr != nil {
					o.logger.Error("compensation failed",
						"operation_id", compensationOps[i],
						"error", compErr)
				}
			}

			// Rollback transaction
			if rollbackErr := o.transactionMgr.RollbackTransaction(ctx, transaction.ID); rollbackErr != nil {
				o.logger.Error("transaction rollback failed",
					"transaction_id", transaction.ID,
					"error", rollbackErr)
			}

			result.Success = false
			result.Error = err.Error()
			result.RolledBack = true
			result.Duration = time.Since(start)
			return result, err
		}

		// Register compensation for successful operation
		compensationOps = append(compensationOps, op.ID)
	}

	// Commit transaction
	if err := o.transactionMgr.CommitTransaction(ctx, transaction.ID); err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("failed to commit transaction: %v", err)
		result.Duration = time.Since(start)
		return result, err
	}

	// Clear compensations for committed transaction
	for _, opID := range compensationOps {
		o.compensationMgr.ClearCompensation(opID)
	}

	result.Success = true
	result.Duration = time.Since(start)

	o.logger.Info("transaction completed successfully",
		"transaction_id", transaction.ID,
		"operations", len(transaction.Operations),
		"duration", result.Duration)

	return result, nil
}

// GetOperationMetrics returns operation metrics
func (o *OperationOrchestrator) GetOperationMetrics() *OrchestratorMetrics {
	o.metrics.mu.RLock()
	defer o.metrics.mu.RUnlock()

	// Return a copy to avoid concurrent access issues
	return &OrchestratorMetrics{
		TotalOperations:     o.metrics.TotalOperations,
		SuccessfulOps:       o.metrics.SuccessfulOps,
		FailedOps:           o.metrics.FailedOps,
		RetriedOps:          o.metrics.RetriedOps,
		AverageLatency:      o.metrics.AverageLatency,
		CircuitBreakerTrips: o.metrics.CircuitBreakerTrips,
		RateLimitHits:       o.metrics.RateLimitHits,
		LastOperationTime:   o.metrics.LastOperationTime,
	}
}

// RegisterPlugin registers a plugin with the orchestrator
func (o *OperationOrchestrator) RegisterPlugin(pluginName string, pluginInstance plugin.IntegrationPluginInterface) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Initialize circuit breaker for plugin
	o.circuitBreakers[pluginName] = resilience.NewCircuitBreaker(&resilience.CircuitBreakerConfig{
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		HalfOpenMaxCalls: 3,
	})

	// Initialize rate limiter for plugin
	rateLimitInfo, err := pluginInstance.GetRateLimitInfo(context.Background())
	if err != nil {
		// Use default rate limits
		rateLimitInfo = &plugin.RateLimitInfo{
			Limit: 300, // Default 300 requests per hour
		}
	}

	o.rateLimiters[pluginName] = resilience.NewRateLimiter(&resilience.RateLimiterConfig{
		RequestsPerMinute: rateLimitInfo.Limit / 60, // Convert hourly to per minute
		BurstSize:         10,
	})

	o.logger.Info("plugin registered with orchestrator", "plugin", pluginName)

	return nil
}

// UnregisterPlugin unregisters a plugin
func (o *OperationOrchestrator) UnregisterPlugin(pluginName string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	delete(o.circuitBreakers, pluginName)
	delete(o.rateLimiters, pluginName)

	o.logger.Info("plugin unregistered from orchestrator", "plugin", pluginName)

	return nil
}

// Helper methods

// isCircuitBreakerClosed checks if the circuit breaker is closed for a plugin
func (o *OperationOrchestrator) isCircuitBreakerClosed(pluginName string) bool {
	o.mu.RLock()
	cb, exists := o.circuitBreakers[pluginName]
	o.mu.RUnlock()

	if !exists {
		return true // No circuit breaker configured
	}

	return cb.IsCallAllowed()
}

// checkRateLimit checks if the plugin is within rate limits
func (o *OperationOrchestrator) checkRateLimit(pluginName string) bool {
	o.mu.RLock()
	rl, exists := o.rateLimiters[pluginName]
	o.mu.RUnlock()

	if !exists {
		return true // No rate limiting configured
	}

	return rl.Allow()
}

// recordSuccess records a successful operation for circuit breaker
func (o *OperationOrchestrator) recordSuccess(pluginName string) {
	o.mu.RLock()
	cb, exists := o.circuitBreakers[pluginName]
	o.mu.RUnlock()

	if exists {
		cb.RecordSuccess()
	}
}

// recordFailure records a failed operation for circuit breaker
func (o *OperationOrchestrator) recordFailure(pluginName string) {
	o.mu.RLock()
	cb, exists := o.circuitBreakers[pluginName]
	o.mu.RUnlock()

	if exists {
		cb.RecordFailure()

		if cb.IsOpen() {
			o.metrics.mu.Lock()
			o.metrics.CircuitBreakerTrips++
			o.metrics.mu.Unlock()
		}
	}
}

// isRetryableError determines if an error is retryable based on policy
func (o *OperationOrchestrator) isRetryableError(err error, policy *RetryPolicy) bool {
	errorStr := err.Error()

	// Check if error is in retryable list
	for _, retryableError := range policy.RetryableErrors {
		if errorStr == retryableError {
			return true
		}
	}

	// Default retryable errors
	retryableErrors := []string{
		"RATE_LIMITED",
		"NETWORK_ERROR",
		"TIMEOUT_ERROR",
		"SERVER_ERROR",
		"SERVICE_UNAVAILABLE",
	}

	for _, retryableError := range retryableErrors {
		if errorStr == retryableError {
			return true
		}
	}

	return false
}

// calculateBackoffDelay calculates the delay for retry attempts
func (o *OperationOrchestrator) calculateBackoffDelay(attempt int, policy *RetryPolicy) time.Duration {
	switch policy.BackoffStrategy {
	case "exponential":
		delay := policy.BaseDelay * time.Duration(1<<attempt)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
		return delay

	case "linear":
		delay := policy.BaseDelay * time.Duration(attempt+1)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
		return delay

	case "fixed":
		return policy.BaseDelay

	default:
		// Default to exponential
		delay := policy.BaseDelay * time.Duration(1<<attempt)
		if delay > policy.MaxDelay {
			delay = policy.MaxDelay
		}
		return delay
	}
}

// updateMetrics updates orchestrator metrics
func (o *OperationOrchestrator) updateMetrics(success, rateLimited, retried bool) {
	o.metrics.mu.Lock()
	defer o.metrics.mu.Unlock()

	o.metrics.TotalOperations++
	if success {
		o.metrics.SuccessfulOps++
	} else {
		o.metrics.FailedOps++
	}
	if rateLimited {
		o.metrics.RateLimitHits++
	}
	if retried {
		o.metrics.RetriedOps++
	}
	o.metrics.LastOperationTime = time.Now()
}
