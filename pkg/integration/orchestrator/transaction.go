package orchestrator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
)

// TransactionManager manages distributed transactions across plugin operations
type TransactionManager struct {
	logger       logging.Logger
	transactions map[string]*TransactionState
	mu           sync.RWMutex
}

// TransactionState represents the state of a transaction
type TransactionState struct {
	ID         string                 `json:"id"`
	Status     TransactionStatus      `json:"status"`
	Operations []string               `json:"operations"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    *time.Time             `json:"end_time,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// TransactionStatus represents the status of a transaction
type TransactionStatus string

const (
	TransactionStatusActive     TransactionStatus = "active"
	TransactionStatusCommitted  TransactionStatus = "committed"
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
	TransactionStatusFailed     TransactionStatus = "failed"
)

// CompensationManager manages compensation actions for failed operations
type CompensationManager struct {
	logger        logging.Logger
	compensations map[string]CompensationAction
	mu            sync.RWMutex
}

// CompensationAction represents a compensation action
type CompensationAction struct {
	OperationID  string                 `json:"operation_id"`
	Function     CompensationFunc       `json:"-"`
	Data         interface{}            `json:"data"`
	RegisteredAt time.Time              `json:"registered_at"`
	Executed     bool                   `json:"executed"`
	ExecutedAt   *time.Time             `json:"executed_at,omitempty"`
	Error        string                 `json:"error,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// NewTransactionManager creates a new transaction manager
func NewTransactionManager(logger logging.Logger) *TransactionManager {
	return &TransactionManager{
		logger:       logger,
		transactions: make(map[string]*TransactionState),
	}
}

// BeginTransaction starts a new transaction
func (tm *TransactionManager) BeginTransaction(ctx context.Context, transactionID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if transaction already exists
	if _, exists := tm.transactions[transactionID]; exists {
		return fmt.Errorf("transaction already exists: %s", transactionID)
	}

	// Create new transaction state
	transaction := &TransactionState{
		ID:         transactionID,
		Status:     TransactionStatusActive,
		Operations: make([]string, 0),
		StartTime:  time.Now(),
		Metadata:   make(map[string]interface{}),
	}

	tm.transactions[transactionID] = transaction

	tm.logger.Debug("transaction started", "id", transactionID)

	return nil
}

// CommitTransaction commits a transaction
func (tm *TransactionManager) CommitTransaction(ctx context.Context, transactionID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	transaction, exists := tm.transactions[transactionID]
	if !exists {
		return fmt.Errorf("transaction not found: %s", transactionID)
	}

	if transaction.Status != TransactionStatusActive {
		return fmt.Errorf("transaction not active: %s (status: %s)", transactionID, transaction.Status)
	}

	// Update transaction state
	now := time.Now()
	transaction.Status = TransactionStatusCommitted
	transaction.EndTime = &now

	tm.logger.Debug("transaction committed",
		"id", transactionID,
		"operations", len(transaction.Operations),
		"duration", time.Since(transaction.StartTime))

	return nil
}

// RollbackTransaction rolls back a transaction
func (tm *TransactionManager) RollbackTransaction(ctx context.Context, transactionID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	transaction, exists := tm.transactions[transactionID]
	if !exists {
		return fmt.Errorf("transaction not found: %s", transactionID)
	}

	if transaction.Status != TransactionStatusActive {
		return fmt.Errorf("transaction not active: %s (status: %s)", transactionID, transaction.Status)
	}

	// Update transaction state
	now := time.Now()
	transaction.Status = TransactionStatusRolledBack
	transaction.EndTime = &now

	tm.logger.Debug("transaction rolled back",
		"id", transactionID,
		"operations", len(transaction.Operations),
		"duration", time.Since(transaction.StartTime))

	return nil
}

// AddOperation adds an operation to a transaction
func (tm *TransactionManager) AddOperation(transactionID string, operation *Operation) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	transaction, exists := tm.transactions[transactionID]
	if !exists {
		return fmt.Errorf("transaction not found: %s", transactionID)
	}

	if transaction.Status != TransactionStatusActive {
		return fmt.Errorf("transaction not active: %s", transactionID)
	}

	transaction.Operations = append(transaction.Operations, operation.ID)

	tm.logger.Debug("operation added to transaction",
		"transaction_id", transactionID,
		"operation_id", operation.ID)

	return nil
}

// GetTransaction returns transaction state
func (tm *TransactionManager) GetTransaction(transactionID string) (*TransactionState, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transaction, exists := tm.transactions[transactionID]
	if !exists {
		return nil, fmt.Errorf("transaction not found: %s", transactionID)
	}

	return transaction, nil
}

// CleanupCompletedTransactions removes old completed transactions
func (tm *TransactionManager) CleanupCompletedTransactions(maxAge time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for id, transaction := range tm.transactions {
		if transaction.EndTime != nil && transaction.EndTime.Before(cutoff) {
			delete(tm.transactions, id)
			tm.logger.Debug("cleaned up old transaction", "id", id)
		}
	}
}

// NewCompensationManager creates a new compensation manager
func NewCompensationManager(logger logging.Logger) *CompensationManager {
	return &CompensationManager{
		logger:        logger,
		compensations: make(map[string]CompensationAction),
	}
}

// RegisterCompensation registers a compensation action
func (cm *CompensationManager) RegisterCompensation(operationID string, compensationFunc CompensationFunc) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	compensation := CompensationAction{
		OperationID:  operationID,
		Function:     compensationFunc,
		RegisteredAt: time.Now(),
		Executed:     false,
		Metadata:     make(map[string]interface{}),
	}

	cm.compensations[operationID] = compensation

	cm.logger.Debug("compensation registered", "operation_id", operationID)

	return nil
}

// ExecuteCompensation executes compensation for a failed operation
func (cm *CompensationManager) ExecuteCompensation(ctx context.Context, operationID string) error {
	cm.mu.Lock()
	compensation, exists := cm.compensations[operationID]
	if !exists {
		cm.mu.Unlock()
		return fmt.Errorf("compensation not found for operation: %s", operationID)
	}

	if compensation.Executed {
		cm.mu.Unlock()
		return fmt.Errorf("compensation already executed for operation: %s", operationID)
	}

	// Mark as executed to prevent concurrent execution
	compensation.Executed = true
	now := time.Now()
	compensation.ExecutedAt = &now
	cm.compensations[operationID] = compensation
	cm.mu.Unlock()

	cm.logger.Debug("executing compensation", "operation_id", operationID)

	// Execute compensation function
	if err := compensation.Function(ctx, compensation.Data); err != nil {
		// Update compensation with error
		cm.mu.Lock()
		compensation.Error = err.Error()
		cm.compensations[operationID] = compensation
		cm.mu.Unlock()

		cm.logger.Error("compensation execution failed",
			"operation_id", operationID,
			"error", err)

		return fmt.Errorf("compensation failed: %w", err)
	}

	cm.logger.Info("compensation executed successfully", "operation_id", operationID)

	return nil
}

// ClearCompensation clears compensation for a successful operation
func (cm *CompensationManager) ClearCompensation(operationID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	delete(cm.compensations, operationID)

	cm.logger.Debug("compensation cleared", "operation_id", operationID)

	return nil
}

// ListCompensations returns all registered compensations
func (cm *CompensationManager) ListCompensations() map[string]CompensationAction {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]CompensationAction)
	for id, compensation := range cm.compensations {
		result[id] = compensation
	}

	return result
}

// CleanupExecutedCompensations removes old executed compensations
func (cm *CompensationManager) CleanupExecutedCompensations(maxAge time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)

	for id, compensation := range cm.compensations {
		if compensation.Executed && compensation.ExecutedAt != nil && compensation.ExecutedAt.Before(cutoff) {
			delete(cm.compensations, id)
			cm.logger.Debug("cleaned up old compensation", "operation_id", id)
		}
	}
}
