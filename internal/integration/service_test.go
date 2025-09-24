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

// MockProvider implements IntegrationProvider for testing
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) GetTaskData(ctx context.Context, externalID string) (*ExternalTaskData, error) {
	args := m.Called(ctx, externalID)
	return args.Get(0).(*ExternalTaskData), args.Error(1)
}

func (m *MockProvider) CreateTask(ctx context.Context, taskData *ZenTaskData) (*ExternalTaskData, error) {
	args := m.Called(ctx, taskData)
	return args.Get(0).(*ExternalTaskData), args.Error(1)
}

func (m *MockProvider) UpdateTask(ctx context.Context, externalID string, taskData *ZenTaskData) (*ExternalTaskData, error) {
	args := m.Called(ctx, externalID, taskData)
	return args.Get(0).(*ExternalTaskData), args.Error(1)
}

func (m *MockProvider) SearchTasks(ctx context.Context, query map[string]interface{}) ([]*ExternalTaskData, error) {
	args := m.Called(ctx, query)
	return args.Get(0).([]*ExternalTaskData), args.Error(1)
}

func (m *MockProvider) ValidateConnection(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockProvider) GetFieldMapping() map[string]string {
	args := m.Called()
	return args.Get(0).(map[string]string)
}

func (m *MockProvider) MapToZen(external *ExternalTaskData) (*ZenTaskData, error) {
	args := m.Called(external)
	return args.Get(0).(*ZenTaskData), args.Error(1)
}

func (m *MockProvider) MapToExternal(zen *ZenTaskData) (*ExternalTaskData, error) {
	args := m.Called(zen)
	return args.Get(0).(*ExternalTaskData), args.Error(1)
}

// New interface methods for enhanced provider
func (m *MockProvider) HealthCheck(ctx context.Context) (*ProviderHealth, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProviderHealth), args.Error(1)
}

func (m *MockProvider) GetRateLimitInfo(ctx context.Context) (*RateLimitInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*RateLimitInfo), args.Error(1)
}

func (m *MockProvider) SupportsRealtime() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockProvider) GetWebhookURL() string {
	args := m.Called()
	return args.String(0)
}

// MockAuthManager implements auth.Manager for testing
type MockAuthManager struct {
	mock.Mock
}

func (m *MockAuthManager) Authenticate(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) GetCredentials(provider string) (string, error) {
	args := m.Called(provider)
	return args.String(0), args.Error(1)
}

func (m *MockAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
}

func (m *MockAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	args := m.Called(ctx, provider)
	return args.Bool(0)
}

func (m *MockAuthManager) ListProviders() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockAuthManager) DeleteCredentials(provider string) error {
	args := m.Called(provider)
	return args.Error(0)
}

func (m *MockAuthManager) GetProviderInfo(provider string) (*auth.ProviderInfo, error) {
	args := m.Called(provider)
	return args.Get(0).(*auth.ProviderInfo), args.Error(1)
}

func TestNewService(t *testing.T) {
	cfg := &config.Config{
		Work: config.WorkConfig{
			Tasks: config.TasksConfig{
				Source: "jira",
				Sync:   "daily",
			},
		},
		Integrations: config.IntegrationsConfig{
			SyncEnabled: true,
		},
	}
	logger := logging.NewBasic()
	authManager := &MockAuthManager{}
	cacheConfig := cache.Config{
		BasePath:    "/tmp/test",
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	serializer := cache.NewJSONSerializer[*TaskSyncRecord]()
	cacheManager := cache.NewManager(cacheConfig, logger, serializer)

	service := NewService(cfg, logger, authManager, cacheManager)

	assert.NotNil(t, service)
	assert.Equal(t, cfg, service.config)
	assert.Equal(t, logger, service.logger)
	assert.Equal(t, authManager, service.auth)
}

func TestService_RegisterProvider(t *testing.T) {
	service := createTestService(t)
	provider := &MockProvider{}
	provider.On("Name").Return("test")

	err := service.RegisterProvider(provider)

	assert.NoError(t, err)
	assert.Contains(t, service.ListProviders(), "test")
}

func TestService_RegisterProvider_NilProvider(t *testing.T) {
	service := createTestService(t)

	err := service.RegisterProvider(nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "provider cannot be nil")
}

func TestService_GetProvider(t *testing.T) {
	service := createTestService(t)
	provider := &MockProvider{}
	provider.On("Name").Return("test")

	// Register provider
	err := service.RegisterProvider(provider)
	assert.NoError(t, err)

	// Get provider
	retrieved, err := service.GetProvider("test")
	assert.NoError(t, err)
	assert.Equal(t, provider, retrieved)
}

func TestService_GetProvider_NotFound(t *testing.T) {
	service := createTestService(t)

	provider, err := service.GetProvider("nonexistent")

	assert.Error(t, err)
	assert.Nil(t, provider)
	assert.Contains(t, err.Error(), "not found")
}

func TestService_IsConfigured(t *testing.T) {
	tests := []struct {
		name           string
		taskSystem     string
		hasProvider    bool
		expectedResult bool
	}{
		{
			name:           "configured with provider",
			taskSystem:     "jira",
			hasProvider:    true,
			expectedResult: true,
		},
		{
			name:           "configured without provider",
			taskSystem:     "jira",
			hasProvider:    false,
			expectedResult: false,
		},
		{
			name:           "not configured",
			taskSystem:     "",
			hasProvider:    false,
			expectedResult: false,
		},
		{
			name:           "none task system",
			taskSystem:     "none",
			hasProvider:    false,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				Work: config.WorkConfig{
					Tasks: config.TasksConfig{
						Source: tt.taskSystem,
						Sync:   "daily",
					},
				},
				Integrations: config.IntegrationsConfig{
					SyncEnabled: true,
				},
			}
			service := createTestServiceWithConfig(t, cfg)

			if tt.hasProvider {
				provider := &MockProvider{}
				provider.On("Name").Return(tt.taskSystem)
				err := service.RegisterProvider(provider)
				assert.NoError(t, err)
			}

			result := service.IsConfigured()
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestService_SyncTask(t *testing.T) {
	tests := []struct {
		name           string
		taskID         string
		opts           SyncOptions
		setupMocks     func(*MockProvider, *MockAuthManager, cache.Manager[*TaskSyncRecord])
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
			setupMocks: func(provider *MockProvider, auth *MockAuthManager, cacheManager cache.Manager[*TaskSyncRecord]) {
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

				// Store sync record in cache
				cacheManager.Put(context.Background(), "task-123", syncRecord, cache.PutOptions{})

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create service with real cache
			cfg := &config.Config{
				Work: config.WorkConfig{
					Tasks: config.TasksConfig{
						Source: "jira",
						Sync:   "daily",
					},
				},
				Integrations: config.IntegrationsConfig{
					SyncEnabled: true,
				},
			}
			logger := logging.NewBasic()
			authManager := &MockAuthManager{}
			cacheConfig := cache.Config{
				BasePath:    "/tmp/test-" + t.Name(),
				SizeLimitMB: 10,
				DefaultTTL:  time.Hour,
			}
			serializer := cache.NewJSONSerializer[*TaskSyncRecord]()
			cacheManager := cache.NewManager(cacheConfig, logger, serializer)

			service := NewService(cfg, logger, authManager, cacheManager)

			// Create and setup provider
			provider := &MockProvider{}
			tt.setupMocks(provider, authManager, cacheManager)

			// Register provider
			err := service.RegisterProvider(provider)
			require.NoError(t, err)

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
		})
	}
}

func TestService_SyncTask_NotConfigured(t *testing.T) {
	cfg := &config.Config{
		Work: config.WorkConfig{
			Tasks: config.TasksConfig{
				Source: "",
				Sync:   "manual",
			},
		},
		Integrations: config.IntegrationsConfig{
			SyncEnabled: false,
		},
	}
	service := createTestServiceWithConfig(t, cfg)

	opts := SyncOptions{
		Direction: SyncDirectionPull,
	}
	result, err := service.SyncTask(context.Background(), "TASK-1", opts)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not configured")
}

func createTestService(t *testing.T) *Service {
	cfg := &config.Config{
		Work: config.WorkConfig{
			Tasks: config.TasksConfig{
				Source: "jira",
				Sync:   "daily",
			},
		},
		Integrations: config.IntegrationsConfig{
			SyncEnabled: true,
		},
	}
	return createTestServiceWithConfig(t, cfg)
}

func createTestServiceWithConfig(t *testing.T, cfg *config.Config) *Service {
	logger := logging.NewBasic()
	authManager := &MockAuthManager{}
	cacheConfig := cache.Config{
		BasePath:    "/tmp/test-" + t.Name(),
		SizeLimitMB: 10,
		DefaultTTL:  time.Hour,
	}
	serializer := cache.NewJSONSerializer[*TaskSyncRecord]()
	cacheManager := cache.NewManager(cacheConfig, logger, serializer)

	return NewService(cfg, logger, authManager, cacheManager)
}

// Enhanced functionality tests

func TestService_ConflictResolution(t *testing.T) {
	tests := []struct {
		name      string
		strategy  ConflictStrategy
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
			cfg := &config.Config{
				Work: config.WorkConfig{
					Tasks: config.TasksConfig{
						Source: "test",
						Sync:   "daily",
					},
				},
				Integrations: config.IntegrationsConfig{
					SyncEnabled: true,
				},
			}
			service := createTestServiceWithConfig(t, cfg)

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
	service := createTestService(t)

	provider := &MockProvider{}
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
	service := createTestService(t)

	provider := &MockProvider{}
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
	service := createTestService(t)

	provider := &MockProvider{}
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
	service := createTestService(t)

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
	service := createTestService(t)

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
	service := createTestService(t)

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
