package integration

import (
	"context"
	"testing"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Enhanced mock provider with new interface methods
type EnhancedMockProvider struct {
	mock.Mock
}

func (m *EnhancedMockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *EnhancedMockProvider) GetTaskData(ctx context.Context, externalID string) (*ExternalTaskData, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExternalTaskData), args.Error(1)
}

func (m *EnhancedMockProvider) CreateTask(ctx context.Context, taskData *ZenTaskData) (*ExternalTaskData, error) {
	args := m.Called(ctx, taskData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExternalTaskData), args.Error(1)
}

func (m *EnhancedMockProvider) UpdateTask(ctx context.Context, externalID string, taskData *ZenTaskData) (*ExternalTaskData, error) {
	args := m.Called(ctx, externalID, taskData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExternalTaskData), args.Error(1)
}

func (m *EnhancedMockProvider) SearchTasks(ctx context.Context, query map[string]interface{}) ([]*ExternalTaskData, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*ExternalTaskData), args.Error(1)
}

func (m *EnhancedMockProvider) ValidateConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *EnhancedMockProvider) GetFieldMapping() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func (m *EnhancedMockProvider) MapToZen(external *ExternalTaskData) (*ZenTaskData, error) {
	args := m.Called(external)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ZenTaskData), args.Error(1)
}

func (m *EnhancedMockProvider) MapToExternal(zen *ZenTaskData) (*ExternalTaskData, error) {
	args := m.Called(zen)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExternalTaskData), args.Error(1)
}

// New interface methods
func (m *EnhancedMockProvider) HealthCheck(ctx context.Context) (*ProviderHealth, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProviderHealth), args.Error(1)
}

func (m *EnhancedMockProvider) GetRateLimitInfo(ctx context.Context) (*RateLimitInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RateLimitInfo), args.Error(1)
}

func (m *EnhancedMockProvider) SupportsRealtime() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *EnhancedMockProvider) GetWebhookURL() string {
	args := m.Called()
	return args.String(0)
}

func TestService_EnhancedSyncTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		opts           SyncOptions
		setupMocks     func(*EnhancedMockProvider, *mockAuthManager, *mockCache)
		expectedError  bool
		expectedResult func(*testing.T, *SyncResult)
	}{
		{
			name:   "successful pull sync with conflict detection",
			taskID: "task-123",
			opts: SyncOptions{
				Direction:        SyncDirectionPull,
				ConflictStrategy: ConflictStrategyTimestamp,
				DryRun:           false,
				ForceSync:        false,
				Timeout:          30 * time.Second,
				RetryCount:       3,
			},
			setupMocks: func(provider *EnhancedMockProvider, auth *mockAuthManager, cache *mockCache) {
				// Mock sync record
				syncRecord := &TaskSyncRecord{
					TaskID:           "task-123",
					ExternalID:       "PROJ-456",
					ExternalSystem:   "jira",
					SyncDirection:    SyncDirectionPull,
					FieldMappings:    map[string]string{"title": "summary"},
					ConflictStrategy: ConflictStrategyTimestamp,
					Status:           SyncStatusActive,
				}
				
				cacheEntry := &cache.Entry[*TaskSyncRecord]{
					Data: syncRecord,
				}
				cache.On("Get", mock.Anything, "task-123").Return(cacheEntry, nil)
				cache.On("Put", mock.Anything, "task-123", mock.AnythingOfType("*integration.TaskSyncRecord"), mock.Anything).Return(nil)
				
				// Mock external task data
				externalData := &ExternalTaskData{
					ID:          "PROJ-456",
					Title:       "Enhanced Task Title",
					Description: "Task description",
					Status:      "In Progress",
					Priority:    "High",
					Updated:     time.Now(),
				}
				provider.On("GetTaskData", mock.Anything, "PROJ-456").Return(externalData, nil)
				
				// Mock mapping
				zenData := &ZenTaskData{
					ID:          "task-123",
					Title:       "Enhanced Task Title",
					Description: "Task description",
					Status:      "in_progress",
					Priority:    "high",
					Updated:     time.Now(),
				}
				provider.On("MapToZen", externalData).Return(zenData, nil)
				provider.On("Name").Return("jira")
			},
			expectedError: false,
			expectedResult: func(t *testing.T, result *SyncResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "task-123", result.TaskID)
				assert.Equal(t, "PROJ-456", result.ExternalID)
				assert.Equal(t, SyncDirectionPull, result.Direction)
				assert.NotEmpty(t, result.CorrelationID)
				assert.Greater(t, result.Duration, time.Duration(0))
			},
		},
		{
			name:   "sync with rate limiting",
			taskID: "task-456",
			opts: SyncOptions{
				Direction:        SyncDirectionPush,
				ConflictStrategy: ConflictStrategyLocalWins,
				RetryCount:       1,
			},
			setupMocks: func(provider *EnhancedMockProvider, auth *mockAuthManager, cache *mockCache) {
				// Mock sync record
				syncRecord := &TaskSyncRecord{
					TaskID:        "task-456",
					ExternalID:    "PROJ-789",
					ExternalSystem: "jira",
					Status:        SyncStatusActive,
				}
				
				cacheEntry := &cache.Entry[*TaskSyncRecord]{
					Data: syncRecord,
				}
				cache.On("Get", mock.Anything, "task-456").Return(cacheEntry, nil)
				
				provider.On("Name").Return("jira")
				provider.On("MapToExternal", mock.AnythingOfType("*integration.ZenTaskData")).Return(&ExternalTaskData{ID: "PROJ-789"}, nil)
				provider.On("UpdateTask", mock.Anything, "PROJ-789", mock.AnythingOfType("*integration.ZenTaskData")).Return(&ExternalTaskData{ID: "PROJ-789"}, nil)
			},
			expectedError: false,
			expectedResult: func(t *testing.T, result *SyncResult) {
				assert.True(t, result.Success)
				assert.Equal(t, "task-456", result.TaskID)
			},
		},
		{
			name:   "sync with circuit breaker open",
			taskID: "task-789",
			opts: SyncOptions{
				Direction: SyncDirectionPull,
			},
			setupMocks: func(provider *EnhancedMockProvider, auth *mockAuthManager, cache *mockCache) {
				// Mock sync record
				syncRecord := &TaskSyncRecord{
					TaskID:     "task-789",
					ExternalID: "PROJ-999",
					Status:     SyncStatusActive,
				}
				
				cacheEntry := &cache.Entry[*TaskSyncRecord]{
					Data: syncRecord,
				}
				cache.On("Get", mock.Anything, "task-789").Return(cacheEntry, nil)
				
				provider.On("Name").Return("jira")
			},
			expectedError: true,
			expectedResult: func(t *testing.T, result *SyncResult) {
				assert.False(t, result.Success)
				assert.Contains(t, result.Error, "circuit breaker")
				assert.Equal(t, ErrCodeProviderError, result.ErrorCode)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			provider := &EnhancedMockProvider{}
			authManager := &mockAuthManager{}
			cacheManager := &mockCache{}
			
			// Setup mocks
			tt.setupMocks(provider, authManager, cacheManager)
			
			// Create service
			cfg := &config.IntegrationsConfig{
				TaskSystem:  "jira",
				SyncEnabled: true,
			}
			logger := logging.NewBasic()
			service := NewService(cfg, logger, authManager, cacheManager)
			
			// Register provider
			err := service.RegisterProvider(provider)
			require.NoError(t, err)
			
			// For circuit breaker test, manually open the circuit breaker
			if tt.name == "sync with circuit breaker open" {
				service.mu.Lock()
				if cb, exists := service.circuitBreakers["jira"]; exists {
					cb.mu.Lock()
					cb.State = CircuitBreakerOpen
					cb.FailureCount = 10
					cb.LastFailureTime = time.Now()
					cb.mu.Unlock()
				}
				service.mu.Unlock()
			}
			
			// Execute sync
			result, err := service.SyncTask(context.Background(), tt.taskID, tt.opts)
			
			// Verify results
			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			
			if result != nil && tt.expectedResult != nil {
				tt.expectedResult(t, result)
			}
			
			// Verify mock expectations
			provider.AssertExpectations(t)
			authManager.AssertExpectations(t)
			cacheManager.AssertExpectations(t)
		})
	}
}

func TestService_ConflictResolution(t *testing.T) {
	tests := []struct {
		name     string
		strategy ConflictStrategy
		conflicts []FieldConflict
		expected  func(*testing.T, error)
	}{
		{
			name:     "local wins strategy",
			strategy: ConflictStrategyLocalWins,
			conflicts: []FieldConflict{
				{
					Field:         "title",
					ZenValue:      "Local Title",
					ExternalValue: "External Title",
				},
			},
			expected: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "timestamp strategy with external newer",
			strategy: ConflictStrategyTimestamp,
			conflicts: []FieldConflict{
				{
					Field:             "description",
					ZenValue:          "Old description",
					ExternalValue:     "New description",
					ZenTimestamp:      time.Now().Add(-1 * time.Hour),
					ExternalTimestamp: time.Now(),
				},
			},
			expected: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name:     "manual review strategy",
			strategy: ConflictStrategyManualReview,
			conflicts: []FieldConflict{
				{
					Field:         "status",
					ZenValue:      "in_progress",
					ExternalValue: "done",
				},
			},
			expected: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service
			cfg := &config.IntegrationsConfig{
				TaskSystem:  "test",
				SyncEnabled: true,
			}
			logger := logging.NewBasic()
			authManager := &mockAuthManager{}
			cacheManager := &mockCache{}
			
			service := NewService(cfg, logger, authManager, cacheManager)
			
			// Create test sync record
			record := &TaskSyncRecord{
				TaskID:     "test-task",
				ExternalID: "EXT-123",
			}
			
			// Test conflict resolution
			err := service.resolveConflicts(context.Background(), record, tt.conflicts, tt.strategy)
			tt.expected(t, err)
			
			// For manual review, check if conflict was stored
			if tt.strategy == ConflictStrategyManualReview {
				service.conflictMu.RLock()
				_, exists := service.conflictStore["test-task"]
				service.conflictMu.RUnlock()
				assert.True(t, exists, "conflict should be stored for manual review")
			}
		})
	}
}

func TestService_CircuitBreaker(t *testing.T) {
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "test",
		SyncEnabled: true,
	}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	cacheManager := &mockCache{}
	
	service := NewService(cfg, logger, authManager, cacheManager)
	
	provider := &EnhancedMockProvider{}
	provider.On("Name").Return("test")
	
	err := service.RegisterProvider(provider)
	require.NoError(t, err)
	
	providerName := "test"
	
	// Initially, circuit breaker should be closed
	assert.True(t, service.isCircuitBreakerClosed(providerName))
	
	// Record failures to open circuit breaker
	for i := 0; i < 5; i++ {
		service.recordFailure(providerName)
	}
	
	// Circuit breaker should now be open
	assert.False(t, service.isCircuitBreakerClosed(providerName))
	
	// Wait for reset timeout and check half-open state
	service.mu.Lock()
	if cb, exists := service.circuitBreakers[providerName]; exists {
		cb.mu.Lock()
		cb.LastFailureTime = time.Now().Add(-35 * time.Second) // Simulate timeout
		cb.mu.Unlock()
	}
	service.mu.Unlock()
	
	// Should transition to half-open
	assert.True(t, service.isCircuitBreakerClosed(providerName))
	
	// Record success to close circuit breaker
	service.recordSuccess(providerName)
	
	// Should be fully closed now
	assert.True(t, service.isCircuitBreakerClosed(providerName))
}

func TestService_RateLimiting(t *testing.T) {
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "test",
		SyncEnabled: true,
	}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	cacheManager := &mockCache{}
	
	service := NewService(cfg, logger, authManager, cacheManager)
	
	provider := &EnhancedMockProvider{}
	provider.On("Name").Return("test")
	
	err := service.RegisterProvider(provider)
	require.NoError(t, err)
	
	providerName := "test"
	
	// Initially, should allow requests
	assert.True(t, service.checkRateLimit(providerName))
	
	// Exhaust rate limit
	service.mu.Lock()
	if limiter, exists := service.rateLimiters[providerName]; exists {
		// Consume all tokens
		for i := 0; i < 25; i++ { // Burst size + some extra
			limiter.Allow()
		}
	}
	service.mu.Unlock()
	
	// Should now be rate limited
	assert.False(t, service.checkRateLimit(providerName))
}

func TestService_HealthMonitoring(t *testing.T) {
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "test",
		SyncEnabled: true,
	}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	cacheManager := &mockCache{}
	
	service := NewService(cfg, logger, authManager, cacheManager)
	
	provider := &EnhancedMockProvider{}
	provider.On("Name").Return("test")
	
	// Mock health check
	healthStatus := &ProviderHealth{
		Provider:    "test",
		Healthy:     true,
		LastChecked: time.Now(),
	}
	provider.On("HealthCheck", mock.Anything).Return(healthStatus, nil)
	
	err := service.RegisterProvider(provider)
	require.NoError(t, err)
	
	// Perform health check
	service.checkProviderHealth("test", provider)
	
	// Get health status
	health, err := service.GetProviderHealth(context.Background(), "test")
	require.NoError(t, err)
	assert.True(t, health.Healthy)
	assert.Equal(t, "test", health.Provider)
	
	// Get all provider health
	allHealth, err := service.GetAllProviderHealth(context.Background())
	require.NoError(t, err)
	assert.Contains(t, allHealth, "test")
	
	provider.AssertExpectations(t)
}

func TestService_MetricsTracking(t *testing.T) {
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "test",
		SyncEnabled: true,
	}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	cacheManager := &mockCache{}
	
	service := NewService(cfg, logger, authManager, cacheManager)
	
	// Initial metrics should be zero
	assert.Equal(t, int64(0), service.metrics.SyncOperations)
	assert.Equal(t, int64(0), service.metrics.SuccessfulSyncs)
	assert.Equal(t, int64(0), service.metrics.FailedSyncs)
	
	// Update metrics with successful operation
	start := time.Now()
	time.Sleep(10 * time.Millisecond) // Simulate operation time
	service.updateMetrics(start, true)
	
	assert.Equal(t, int64(1), service.metrics.SyncOperations)
	assert.Equal(t, int64(1), service.metrics.SuccessfulSyncs)
	assert.Equal(t, int64(0), service.metrics.FailedSyncs)
	assert.Greater(t, service.metrics.AverageLatency, time.Duration(0))
	
	// Update metrics with failed operation
	service.updateMetrics(start, false)
	
	assert.Equal(t, int64(2), service.metrics.SyncOperations)
	assert.Equal(t, int64(1), service.metrics.SuccessfulSyncs)
	assert.Equal(t, int64(1), service.metrics.FailedSyncs)
}

func TestService_ErrorHandling(t *testing.T) {
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "test",
		SyncEnabled: true,
	}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	cacheManager := &mockCache{}
	
	service := NewService(cfg, logger, authManager, cacheManager)
	
	// Test error creation
	err := service.createIntegrationError(ErrCodeProviderError, "test error", "test-provider", "task-123")
	
	assert.Equal(t, ErrCodeProviderError, err.Code)
	assert.Equal(t, "test error", err.Message)
	assert.Equal(t, "test-provider", err.Provider)
	assert.Equal(t, "task-123", err.TaskID)
	assert.True(t, err.Retryable)
	
	// Test retryable error detection
	assert.True(t, service.isRetryableError(err))
	
	// Test non-retryable error
	nonRetryableErr := service.createIntegrationError(ErrCodeInvalidData, "invalid data", "test", "task")
	assert.False(t, service.isRetryableError(nonRetryableErr))
}

func TestService_RetryLogic(t *testing.T) {
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "test",
		SyncEnabled: true,
	}
	logger := logging.NewBasic()
	authManager := &mockAuthManager{}
	cacheManager := &mockCache{}
	
	service := NewService(cfg, logger, authManager, cacheManager)
	
	// Test successful operation (no retries needed)
	callCount := 0
	err := service.retryWithBackoff(context.Background(), 3, func() error {
		callCount++
		return nil
	})
	
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount)
	
	// Test operation that succeeds after retries
	callCount = 0
	err = service.retryWithBackoff(context.Background(), 3, func() error {
		callCount++
		if callCount < 3 {
			return service.createIntegrationError(ErrCodeNetworkError, "network error", "test", "task")
		}
		return nil
	})
	
	assert.NoError(t, err)
	assert.Equal(t, 3, callCount)
	
	// Test non-retryable error
	callCount = 0
	err = service.retryWithBackoff(context.Background(), 3, func() error {
		callCount++
		return service.createIntegrationError(ErrCodeInvalidData, "invalid data", "test", "task")
	})
	
	assert.Error(t, err)
	assert.Equal(t, 1, callCount) // Should not retry
}

// Mock implementations for testing

type mockAuthManager struct {
	mock.Mock
}

func (m *mockAuthManager) Authenticate(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *mockAuthManager) GetCredentials(provider string) (string, error) {
	args := m.Called(provider)
	return args.String(0), args.Error(1)
}

func (m *mockAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *mockAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *mockAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	args := m.Called(ctx, provider)
	return args.Bool(0)
}

func (m *mockAuthManager) ListProviders() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *mockAuthManager) DeleteCredentials(provider string) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *mockAuthManager) GetProviderInfo(provider string) (*auth.ProviderInfo, error) {
	args := m.Called(provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.ProviderInfo), args.Error(1)
}

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (*cache.Entry[*TaskSyncRecord], error) {
	args := m.Called(ctx, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cache.Entry[*TaskSyncRecord]), args.Error(1)
}

func (m *mockCache) Put(ctx context.Context, key string, data *TaskSyncRecord, opts cache.PutOptions) error {
	args := m.Called(ctx, key, data, opts)
	return args.Error(0)
}

func (m *mockCache) Delete(ctx context.Context, key string) error {
	args := m.Called(ctx, key)
	return args.Error(0)
}

func (m *mockCache) Clear(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockCache) Stats() cache.Stats {
	args := m.Called()
	return args.Get(0).(cache.Stats)
}

func (m *mockCache) Close() error {
	args := m.Called()
	return args.Error(0)
}
