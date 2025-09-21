package integration

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"golang.org/x/time/rate"
)

// Service implements the main integration service
type Service struct {
	config    *config.IntegrationsConfig
	logger    logging.Logger
	auth      auth.Manager
	cache     cache.Manager[*TaskSyncRecord]
	providers map[string]IntegrationProvider
	mu        sync.RWMutex

	// Rate limiting and circuit breaker
	rateLimiters    map[string]*rate.Limiter
	circuitBreakers map[string]*CircuitBreaker
	healthStatus    map[string]*ProviderHealth
	healthMu        sync.RWMutex

	// Metrics and monitoring
	metrics *ServiceMetrics

	// Conflict resolution
	conflictStore map[string]*ConflictRecord
	conflictMu    sync.RWMutex
}

// ServiceMetrics contains service performance metrics
type ServiceMetrics struct {
	SyncOperations    int64
	SuccessfulSyncs   int64
	FailedSyncs       int64
	ConflictCount     int64
	AverageLatency    time.Duration
	LastOperationTime time.Time
	mu                sync.RWMutex
}

// CircuitBreaker implements a simple circuit breaker pattern
type CircuitBreaker struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	FailureCount     int
	LastFailureTime  time.Time
	State            CircuitBreakerState
	mu               sync.RWMutex
}

type CircuitBreakerState int

const (
	CircuitBreakerClosed CircuitBreakerState = iota
	CircuitBreakerOpen
	CircuitBreakerHalfOpen
)

// NewService creates a new integration service
func NewService(
	cfg *config.IntegrationsConfig,
	logger logging.Logger,
	authManager auth.Manager,
	cacheManager cache.Manager[*TaskSyncRecord],
) *Service {
	s := &Service{
		config:          cfg,
		logger:          logger,
		auth:            authManager,
		cache:           cacheManager,
		providers:       make(map[string]IntegrationProvider),
		rateLimiters:    make(map[string]*rate.Limiter),
		circuitBreakers: make(map[string]*CircuitBreaker),
		healthStatus:    make(map[string]*ProviderHealth),
		metrics:         &ServiceMetrics{},
		conflictStore:   make(map[string]*ConflictRecord),
	}

	// Start background health monitoring
	go s.startHealthMonitoring()

	return s
}

// GetProvider returns a provider by name
func (s *Service) GetProvider(name string) (IntegrationProvider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	provider, exists := s.providers[name]
	if !exists {
		return nil, fmt.Errorf("integration provider '%s' not found", name)
	}

	return provider, nil
}

// RegisterProvider registers a new integration provider
func (s *Service) RegisterProvider(provider IntegrationProvider) error {
	if provider == nil {
		return fmt.Errorf("provider cannot be nil")
	}

	name := provider.Name()
	if name == "" {
		return fmt.Errorf("provider name cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Initialize rate limiter for provider
	s.rateLimiters[name] = rate.NewLimiter(rate.Limit(10), 20) // 10 requests/second with burst of 20

	// Initialize circuit breaker for provider
	s.circuitBreakers[name] = &CircuitBreaker{
		FailureThreshold: 5,
		ResetTimeout:     30 * time.Second,
		State:            CircuitBreakerClosed,
	}

	// Initialize health status
	s.healthStatus[name] = &ProviderHealth{
		Provider:    name,
		Healthy:     true,
		LastChecked: time.Now(),
	}

	s.providers[name] = provider
	s.logger.Info("registered integration provider", "provider", name)

	return nil
}

// ListProviders returns all registered providers
func (s *Service) ListProviders() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	providers := make([]string, 0, len(s.providers))
	for name := range s.providers {
		providers = append(providers, name)
	}

	return providers
}

// IsConfigured checks if integration is properly configured
func (s *Service) IsConfigured() bool {
	if s.config == nil {
		return false
	}

	// Check if a task system is configured
	if s.config.TaskSystem == "" || s.config.TaskSystem == "none" {
		return false
	}

	// Check if the configured provider exists
	s.mu.RLock()
	_, exists := s.providers[s.config.TaskSystem]
	s.mu.RUnlock()

	return exists
}

// GetTaskSystem returns the configured task system of record
func (s *Service) GetTaskSystem() string {
	if s.config == nil {
		return ""
	}
	return s.config.TaskSystem
}

// IsSyncEnabled returns true if sync is enabled
func (s *Service) IsSyncEnabled() bool {
	if s.config == nil {
		return false
	}
	return s.config.SyncEnabled
}

// SyncTask synchronizes a single task with advanced error handling and conflict resolution
func (s *Service) SyncTask(ctx context.Context, taskID string, opts SyncOptions) (*SyncResult, error) {
	start := time.Now()
	correlationID := opts.CorrelationID
	if correlationID == "" {
		correlationID = s.generateCorrelationID()
	}

	s.logger.Debug("starting task sync",
		"task_id", taskID,
		"direction", opts.Direction,
		"correlation_id", correlationID)

	// Check if integration is configured
	if !s.IsConfigured() {
		return nil, s.createIntegrationError(ErrCodeConfigError, "integration not configured", "", taskID)
	}

	// Get the provider
	provider, err := s.GetProvider(s.config.TaskSystem)
	if err != nil {
		return nil, s.createIntegrationError(ErrCodeProviderError, fmt.Sprintf("failed to get provider: %v", err), s.config.TaskSystem, taskID)
	}

	// Check circuit breaker
	if !s.isCircuitBreakerClosed(s.config.TaskSystem) {
		return nil, s.createIntegrationError(ErrCodeProviderError, "provider circuit breaker is open", s.config.TaskSystem, taskID)
	}

	// Check rate limiting
	if !s.checkRateLimit(s.config.TaskSystem) {
		return nil, s.createIntegrationError(ErrCodeRateLimited, "rate limit exceeded", s.config.TaskSystem, taskID)
	}

	// Get sync record
	syncRecord, err := s.GetSyncRecord(ctx, taskID)
	if err != nil {
		return nil, s.createIntegrationError(ErrCodeInvalidData, fmt.Sprintf("sync record not found for task %s: %v", taskID, err), s.config.TaskSystem, taskID)
	}

	result := &SyncResult{
		TaskID:        taskID,
		Direction:     opts.Direction,
		ChangedFields: []string{},
		Conflicts:     []FieldConflict{},
		Timestamp:     time.Now(),
		CorrelationID: correlationID,
		Metadata:      make(map[string]interface{}),
	}

	// Set timeout context
	timeout := opts.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	syncCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Perform sync with retry logic
	retryCount := opts.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}

	err = s.retryWithBackoff(syncCtx, retryCount, func() error {
		return s.performSync(syncCtx, provider, syncRecord, result, opts)
	})

	// Update metrics
	s.updateMetrics(start, err == nil)

	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Retryable = s.isRetryableError(err)

		if integrationErr, ok := err.(IntegrationError); ok {
			result.ErrorCode = integrationErr.Code
		}

		// Record failure in circuit breaker
		s.recordFailure(s.config.TaskSystem)

		// Update sync record with error
		syncRecord.Status = SyncStatusError
		syncRecord.ErrorCount++
		syncRecord.LastError = err.Error()
		syncRecord.RetryAfter = s.calculateRetryAfter(syncRecord.ErrorCount)

		s.logger.Error("task sync failed",
			"task_id", taskID,
			"error", err,
			"correlation_id", correlationID,
			"retry_count", syncRecord.ErrorCount)
		return result, err
	}

	result.Success = true
	result.ExternalID = syncRecord.ExternalID
	result.Duration = time.Since(start)

	// Record success in circuit breaker
	s.recordSuccess(s.config.TaskSystem)

	// Update sync record with success
	syncRecord.Status = SyncStatusActive
	syncRecord.LastSyncTime = time.Now()
	syncRecord.ErrorCount = 0
	syncRecord.LastError = ""
	syncRecord.RetryAfter = nil
	syncRecord.Version++

	if err := s.UpdateSyncRecord(ctx, syncRecord); err != nil {
		s.logger.Warn("failed to update sync record", "task_id", taskID, "error", err)
	}

	s.logger.Info("task sync completed",
		"task_id", taskID,
		"external_id", syncRecord.ExternalID,
		"changed_fields", len(result.ChangedFields),
		"duration", result.Duration,
		"correlation_id", correlationID)

	return result, nil
}

// SyncAllTasks synchronizes all tasks configured for sync
func (s *Service) SyncAllTasks(ctx context.Context, opts SyncOptions) ([]*SyncResult, error) {
	s.logger.Debug("starting sync all tasks", "direction", opts.Direction)

	syncRecords, err := s.ListSyncRecords(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list sync records: %w", err)
	}

	results := make([]*SyncResult, 0, len(syncRecords))

	for _, record := range syncRecords {
		result, err := s.SyncTask(ctx, record.TaskID, opts)
		if err != nil {
			s.logger.Error("failed to sync task", "task_id", record.TaskID, "error", err)
			// Continue with other tasks even if one fails
		}
		if result != nil {
			results = append(results, result)
		}
	}

	s.logger.Info("sync all tasks completed", "total", len(syncRecords), "results", len(results))

	return results, nil
}

// GetSyncRecord retrieves the sync record for a task
func (s *Service) GetSyncRecord(ctx context.Context, taskID string) (*TaskSyncRecord, error) {
	entry, err := s.cache.Get(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("sync record not found for task %s: %w", taskID, err)
	}

	return entry.Data, nil
}

// CreateSyncRecord creates a new sync record
func (s *Service) CreateSyncRecord(ctx context.Context, record *TaskSyncRecord) error {
	if record.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	// Check if record already exists
	if _, err := s.GetSyncRecord(ctx, record.TaskID); err == nil {
		return fmt.Errorf("sync record already exists for task %s", record.TaskID)
	}

	opts := cache.PutOptions{}
	if err := s.cache.Put(ctx, record.TaskID, record, opts); err != nil {
		return fmt.Errorf("failed to create sync record: %w", err)
	}

	s.logger.Debug("created sync record", "task_id", record.TaskID, "external_id", record.ExternalID)

	return nil
}

// UpdateSyncRecord updates an existing sync record with versioning
func (s *Service) UpdateSyncRecord(ctx context.Context, record *TaskSyncRecord) error {
	if record.TaskID == "" {
		return fmt.Errorf("task ID cannot be empty")
	}

	// Update timestamps and version
	record.UpdatedAt = time.Now()
	if record.CreatedAt.IsZero() {
		record.CreatedAt = record.UpdatedAt
	}

	opts := cache.PutOptions{}
	if err := s.cache.Put(ctx, record.TaskID, record, opts); err != nil {
		return fmt.Errorf("failed to update sync record: %w", err)
	}

	s.logger.Debug("updated sync record",
		"task_id", record.TaskID,
		"external_id", record.ExternalID,
		"version", record.Version,
		"status", record.Status)

	return nil
}

// DeleteSyncRecord deletes a sync record
func (s *Service) DeleteSyncRecord(ctx context.Context, taskID string) error {
	if err := s.cache.Delete(ctx, taskID); err != nil {
		return fmt.Errorf("failed to delete sync record: %w", err)
	}

	s.logger.Debug("deleted sync record", "task_id", taskID)

	return nil
}

// ListSyncRecords lists all sync records
func (s *Service) ListSyncRecords(ctx context.Context) ([]*TaskSyncRecord, error) {
	// Note: This is a simplified implementation
	// In a production system, you might want to use a different storage mechanism
	// that supports listing all records more efficiently

	// For now, we'll return an empty list as the cache interface doesn't support listing
	// This would need to be enhanced with a proper storage backend for sync records
	s.logger.Debug("listing sync records - simplified implementation")

	return []*TaskSyncRecord{}, nil
}

// pullFromExternal pulls data from external system to Zen
func (s *Service) pullFromExternal(ctx context.Context, provider IntegrationProvider, record *TaskSyncRecord, result *SyncResult, opts SyncOptions) error {
	if record.ExternalID == "" {
		return fmt.Errorf("external ID not set for task %s", record.TaskID)
	}

	// Get external data
	externalData, err := provider.GetTaskData(ctx, record.ExternalID)
	if err != nil {
		return fmt.Errorf("failed to get external task data: %w", err)
	}

	// Convert to Zen format
	zenData, err := provider.MapToZen(externalData)
	if err != nil {
		return fmt.Errorf("failed to map external data to Zen format: %w", err)
	}

	// TODO: Update Zen task with external data
	// This would integrate with the task management system

	s.logger.Debug("pulled data from external system",
		"task_id", record.TaskID,
		"external_id", record.ExternalID,
		"title", zenData.Title)

	return nil
}

// pushToExternal pushes data from Zen to external system
func (s *Service) pushToExternal(ctx context.Context, provider IntegrationProvider, record *TaskSyncRecord, result *SyncResult, opts SyncOptions) error {
	// TODO: Get Zen task data
	// This would integrate with the task management system

	// For now, create a placeholder Zen task data
	zenData := &ZenTaskData{
		ID:          record.TaskID,
		Title:       "Sample Task",
		Description: "Sample task description",
		Status:      "In Progress",
		Priority:    "Medium",
	}

	// Convert to external format
	_, err := provider.MapToExternal(zenData)
	if err != nil {
		return fmt.Errorf("failed to map Zen data to external format: %w", err)
	}

	// Update or create external task
	var updatedExternal *ExternalTaskData
	if record.ExternalID != "" {
		updatedExternal, err = provider.UpdateTask(ctx, record.ExternalID, zenData)
	} else {
		updatedExternal, err = provider.CreateTask(ctx, zenData)
		if err == nil {
			record.ExternalID = updatedExternal.ID
		}
	}

	if err != nil {
		return fmt.Errorf("failed to update external task: %w", err)
	}

	s.logger.Debug("pushed data to external system",
		"task_id", record.TaskID,
		"external_id", updatedExternal.ID)

	return nil
}

// bidirectionalSync performs bidirectional synchronization with conflict detection and resolution
func (s *Service) bidirectionalSync(ctx context.Context, provider IntegrationProvider, record *TaskSyncRecord, result *SyncResult, opts SyncOptions) error {
	// Get both local and external data for comparison
	externalData, err := provider.GetTaskData(ctx, record.ExternalID)
	if err != nil {
		return s.createIntegrationError(ErrCodeProviderError, fmt.Sprintf("failed to get external data: %v", err), provider.Name(), record.TaskID)
	}

	// TODO: Get Zen task data - this would integrate with the task management system
	// For now, create placeholder data
	zenData := &ZenTaskData{
		ID:          record.TaskID,
		Title:       "Sample Task",
		Description: "Sample task description",
		Status:      "In Progress",
		Priority:    "Medium",
		Updated:     time.Now().Add(-1 * time.Hour), // Simulate last update
	}

	// Detect conflicts
	conflicts, err := s.detectConflicts(zenData, externalData, record.FieldMappings)
	if err != nil {
		return s.createIntegrationError(ErrCodeSyncConflict, fmt.Sprintf("failed to detect conflicts: %v", err), provider.Name(), record.TaskID)
	}

	if len(conflicts) > 0 {
		result.Conflicts = conflicts
		s.metrics.mu.Lock()
		s.metrics.ConflictCount++
		s.metrics.mu.Unlock()

		// Handle conflicts based on strategy
		if err := s.resolveConflicts(ctx, record, conflicts, opts.ConflictStrategy); err != nil {
			return err
		}
	}

	// Perform sync based on conflict resolution
	if opts.ConflictStrategy != ConflictStrategyManualReview || len(conflicts) == 0 {
		// Determine sync direction based on timestamps and conflict resolution
		if s.shouldPullFirst(zenData, externalData, conflicts, opts.ConflictStrategy) {
			if err := s.pullFromExternal(ctx, provider, record, result, opts); err != nil {
				s.logger.Warn("failed to pull during bidirectional sync", "error", err)
			}
		} else {
			if err := s.pushToExternal(ctx, provider, record, result, opts); err != nil {
				s.logger.Warn("failed to push during bidirectional sync", "error", err)
			}
		}
	}

	return nil
}

// Helper methods for enhanced functionality

// generateCorrelationID generates a unique correlation ID for tracking operations
func (s *Service) generateCorrelationID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// createIntegrationError creates a standardized integration error
func (s *Service) createIntegrationError(code, message, provider, taskID string) IntegrationError {
	return IntegrationError{
		Code:      code,
		Message:   message,
		Provider:  provider,
		TaskID:    taskID,
		Timestamp: time.Now(),
		Retryable: s.isRetryableErrorCode(code),
	}
}

// isRetryableErrorCode determines if an error code represents a retryable error
func (s *Service) isRetryableErrorCode(code string) bool {
	retryableCodes := map[string]bool{
		ErrCodeRateLimited:   true,
		ErrCodeNetworkError:  true,
		ErrCodeTimeoutError:  true,
		ErrCodeProviderError: true,
	}
	return retryableCodes[code]
}

// isRetryableError determines if an error is retryable
func (s *Service) isRetryableError(err error) bool {
	if integrationErr, ok := err.(IntegrationError); ok {
		return integrationErr.Retryable
	}
	return false
}

// calculateRetryAfter calculates the next retry time based on error count
func (s *Service) calculateRetryAfter(errorCount int) *time.Time {
	// Exponential backoff: 1s, 2s, 4s, 8s, 16s, max 5 minutes
	backoffSeconds := 1 << errorCount
	if backoffSeconds > 300 { // Max 5 minutes
		backoffSeconds = 300
	}
	retryTime := time.Now().Add(time.Duration(backoffSeconds) * time.Second)
	return &retryTime
}

// checkRateLimit checks if the provider is within rate limits
func (s *Service) checkRateLimit(provider string) bool {
	s.mu.RLock()
	limiter, exists := s.rateLimiters[provider]
	s.mu.RUnlock()

	if !exists {
		return true // No rate limiting configured
	}

	return limiter.Allow()
}

// isCircuitBreakerClosed checks if the circuit breaker is closed for the provider
func (s *Service) isCircuitBreakerClosed(provider string) bool {
	s.mu.RLock()
	cb, exists := s.circuitBreakers[provider]
	s.mu.RUnlock()

	if !exists {
		return true // No circuit breaker configured
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.State {
	case CircuitBreakerClosed:
		return true
	case CircuitBreakerOpen:
		// Check if we should transition to half-open
		if time.Since(cb.LastFailureTime) > cb.ResetTimeout {
			cb.State = CircuitBreakerHalfOpen
			return true
		}
		return false
	case CircuitBreakerHalfOpen:
		return true
	default:
		return false
	}
}

// recordFailure records a failure for circuit breaker tracking
func (s *Service) recordFailure(provider string) {
	s.mu.RLock()
	cb, exists := s.circuitBreakers[provider]
	s.mu.RUnlock()

	if !exists {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.FailureCount++
	cb.LastFailureTime = time.Now()

	if cb.FailureCount >= cb.FailureThreshold {
		cb.State = CircuitBreakerOpen
		s.logger.Warn("circuit breaker opened for provider", "provider", provider, "failures", cb.FailureCount)
	}
}

// recordSuccess records a success for circuit breaker tracking
func (s *Service) recordSuccess(provider string) {
	s.mu.RLock()
	cb, exists := s.circuitBreakers[provider]
	s.mu.RUnlock()

	if !exists {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.FailureCount = 0
	if cb.State == CircuitBreakerHalfOpen {
		cb.State = CircuitBreakerClosed
		s.logger.Info("circuit breaker closed for provider", "provider", provider)
	}
}

// updateMetrics updates service performance metrics
func (s *Service) updateMetrics(startTime time.Time, success bool) {
	s.metrics.mu.Lock()
	defer s.metrics.mu.Unlock()

	s.metrics.SyncOperations++
	if success {
		s.metrics.SuccessfulSyncs++
	} else {
		s.metrics.FailedSyncs++
	}

	duration := time.Since(startTime)
	// Simple moving average for latency
	if s.metrics.SyncOperations == 1 {
		s.metrics.AverageLatency = duration
	} else {
		s.metrics.AverageLatency = (s.metrics.AverageLatency + duration) / 2
	}

	s.metrics.LastOperationTime = time.Now()
}

// retryWithBackoff performs an operation with exponential backoff retry logic
func (s *Service) retryWithBackoff(ctx context.Context, maxRetries int, operation func() error) error {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if it's not a retryable error
		if !s.isRetryableError(err) {
			return err
		}

		// Don't wait after the last attempt
		if attempt == maxRetries {
			break
		}

		// Exponential backoff with jitter
		backoffTime := time.Duration(1<<attempt) * time.Second
		if backoffTime > 30*time.Second {
			backoffTime = 30 * time.Second
		}

		s.logger.Debug("retrying operation after backoff",
			"attempt", attempt+1,
			"backoff", backoffTime,
			"error", err)

		timer := time.NewTimer(backoffTime)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
			// Continue to next attempt
		}
	}

	return lastErr
}

// performSync performs the actual synchronization operation
func (s *Service) performSync(ctx context.Context, provider IntegrationProvider, record *TaskSyncRecord, result *SyncResult, opts SyncOptions) error {
	switch opts.Direction {
	case SyncDirectionPull:
		return s.pullFromExternal(ctx, provider, record, result, opts)
	case SyncDirectionPush:
		return s.pushToExternal(ctx, provider, record, result, opts)
	case SyncDirectionBidirectional:
		return s.bidirectionalSync(ctx, provider, record, result, opts)
	default:
		return s.createIntegrationError(ErrCodeInvalidData, fmt.Sprintf("unsupported sync direction: %s", opts.Direction), provider.Name(), record.TaskID)
	}
}

// detectConflicts compares local and external data to find conflicts
func (s *Service) detectConflicts(zenData *ZenTaskData, externalData *ExternalTaskData, fieldMappings map[string]string) ([]FieldConflict, error) {
	var conflicts []FieldConflict

	// Compare mapped fields
	for zenField, externalField := range fieldMappings {
		zenValue := s.getFieldValue(zenData, zenField)
		externalValue := s.getFieldValue(externalData, externalField)

		// Skip if values are the same
		if s.valuesEqual(zenValue, externalValue) {
			continue
		}

		// Create conflict record
		conflict := FieldConflict{
			Field:             zenField,
			ZenValue:          zenValue,
			ExternalValue:     externalValue,
			ZenTimestamp:      zenData.Updated,
			ExternalTimestamp: externalData.Updated,
		}

		conflicts = append(conflicts, conflict)
	}

	return conflicts, nil
}

// resolveConflicts resolves conflicts based on the specified strategy
func (s *Service) resolveConflicts(ctx context.Context, record *TaskSyncRecord, conflicts []FieldConflict, strategy ConflictStrategy) error {
	switch strategy {
	case ConflictStrategyLocalWins:
		// Keep Zen data, no action needed
		return nil

	case ConflictStrategyRemoteWins:
		// Accept external data, no action needed (will be handled in sync)
		return nil

	case ConflictStrategyTimestamp:
		// Resolve based on timestamps - handled in shouldPullFirst
		return nil

	case ConflictStrategyManualReview:
		// Store conflicts for manual resolution
		return s.storeConflictRecord(record.TaskID, conflicts)

	default:
		return s.createIntegrationError(ErrCodeInvalidData, fmt.Sprintf("unsupported conflict strategy: %s", strategy), "", record.TaskID)
	}
}

// storeConflictRecord stores a conflict record for manual resolution
func (s *Service) storeConflictRecord(taskID string, conflicts []FieldConflict) error {
	s.conflictMu.Lock()
	defer s.conflictMu.Unlock()

	conflictRecord := &ConflictRecord{
		ID:        s.generateCorrelationID(),
		TaskID:    taskID,
		Conflicts: conflicts,
		CreatedAt: time.Now(),
		Status:    ConflictStatusPending,
	}

	s.conflictStore[taskID] = conflictRecord
	s.logger.Info("stored conflict record for manual resolution", "task_id", taskID, "conflicts", len(conflicts))

	return nil
}

// shouldPullFirst determines whether to pull or push first based on data comparison
func (s *Service) shouldPullFirst(zenData *ZenTaskData, externalData *ExternalTaskData, conflicts []FieldConflict, strategy ConflictStrategy) bool {
	switch strategy {
	case ConflictStrategyRemoteWins:
		return true // Always pull first
	case ConflictStrategyLocalWins:
		return false // Always push first
	case ConflictStrategyTimestamp:
		// Pull if external data is newer
		return externalData.Updated.After(zenData.Updated)
	default:
		// Default to pulling first
		return true
	}
}

// getFieldValue extracts a field value from a data structure using reflection-like access
func (s *Service) getFieldValue(data interface{}, field string) interface{} {
	// This is a simplified implementation
	// In production, you'd want more sophisticated field access
	switch d := data.(type) {
	case *ZenTaskData:
		switch field {
		case "title":
			return d.Title
		case "description":
			return d.Description
		case "status":
			return d.Status
		case "priority":
			return d.Priority
		case "owner":
			return d.Owner
		}
	case *ExternalTaskData:
		switch field {
		case "title", "summary":
			return d.Title
		case "description":
			return d.Description
		case "status":
			return d.Status
		case "priority":
			return d.Priority
		case "assignee":
			return d.Assignee
		}
	}
	return nil
}

// valuesEqual compares two values for equality
func (s *Service) valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// startHealthMonitoring starts background health monitoring for providers
func (s *Service) startHealthMonitoring() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.performHealthChecks()
	}
}

// performHealthChecks performs health checks on all registered providers
func (s *Service) performHealthChecks() {
	s.mu.RLock()
	providers := make(map[string]IntegrationProvider)
	for name, provider := range s.providers {
		providers[name] = provider
	}
	s.mu.RUnlock()

	for name, provider := range providers {
		go s.checkProviderHealth(name, provider)
	}
}

// checkProviderHealth performs a health check on a single provider
func (s *Service) checkProviderHealth(name string, provider IntegrationProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	start := time.Now()
	health, err := provider.HealthCheck(ctx)
	responseTime := time.Since(start)

	s.healthMu.Lock()
	defer s.healthMu.Unlock()

	if err != nil {
		s.healthStatus[name] = &ProviderHealth{
			Provider:     name,
			Healthy:      false,
			LastChecked:  time.Now(),
			ResponseTime: responseTime,
			ErrorCount:   s.healthStatus[name].ErrorCount + 1,
			LastError:    err.Error(),
		}
		s.logger.Warn("provider health check failed", "provider", name, "error", err)
	} else if health != nil {
		s.healthStatus[name] = health
		s.healthStatus[name].ResponseTime = responseTime
		s.healthStatus[name].LastChecked = time.Now()
	}
}

// GetProviderHealth returns health status for a provider
func (s *Service) GetProviderHealth(ctx context.Context, provider string) (*ProviderHealth, error) {
	s.healthMu.RLock()
	defer s.healthMu.RUnlock()

	health, exists := s.healthStatus[provider]
	if !exists {
		return nil, s.createIntegrationError(ErrCodeProviderError, fmt.Sprintf("provider '%s' not found", provider), provider, "")
	}

	return health, nil
}

// GetAllProviderHealth returns health status for all providers
func (s *Service) GetAllProviderHealth(ctx context.Context) (map[string]*ProviderHealth, error) {
	s.healthMu.RLock()
	defer s.healthMu.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]*ProviderHealth)
	for name, health := range s.healthStatus {
		result[name] = health
	}

	return result, nil
}
