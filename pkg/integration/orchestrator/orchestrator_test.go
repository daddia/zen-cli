package orchestrator

import (
	"context"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/integration/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock plugin for testing
type mockPlugin struct {
	mock.Mock
}

func (m *mockPlugin) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockPlugin) Version() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockPlugin) Description() string {
	args := m.Called()
	return args.String(0)
}

func (m *mockPlugin) Initialize(ctx context.Context, config *plugin.PluginConfig) error {
	args := m.Called(ctx, config)
	return args.Error(0)
}

func (m *mockPlugin) Validate(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockPlugin) HealthCheck(ctx context.Context) (*plugin.PluginHealth, error) {
	args := m.Called(ctx)
	return args.Get(0).(*plugin.PluginHealth), args.Error(1)
}

func (m *mockPlugin) Shutdown(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *mockPlugin) FetchTask(ctx context.Context, externalID string, opts *plugin.FetchOptions) (*plugin.TaskData, error) {
	args := m.Called(ctx, externalID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plugin.TaskData), args.Error(1)
}

func (m *mockPlugin) CreateTask(ctx context.Context, taskData *plugin.TaskData, opts *plugin.CreateOptions) (*plugin.TaskData, error) {
	args := m.Called(ctx, taskData, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plugin.TaskData), args.Error(1)
}

func (m *mockPlugin) UpdateTask(ctx context.Context, externalID string, taskData *plugin.TaskData, opts *plugin.UpdateOptions) (*plugin.TaskData, error) {
	args := m.Called(ctx, externalID, taskData, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plugin.TaskData), args.Error(1)
}

func (m *mockPlugin) DeleteTask(ctx context.Context, externalID string, opts *plugin.DeleteOptions) error {
	args := m.Called(ctx, externalID, opts)
	return args.Error(0)
}

func (m *mockPlugin) SearchTasks(ctx context.Context, query *plugin.SearchQuery, opts *plugin.SearchOptions) ([]*plugin.TaskData, error) {
	args := m.Called(ctx, query, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*plugin.TaskData), args.Error(1)
}

func (m *mockPlugin) SyncTask(ctx context.Context, taskID string, opts *plugin.SyncOptions) (*plugin.SyncResult, error) {
	args := m.Called(ctx, taskID, opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plugin.SyncResult), args.Error(1)
}

func (m *mockPlugin) GetSyncMetadata(ctx context.Context, taskID string) (*plugin.SyncMetadata, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plugin.SyncMetadata), args.Error(1)
}

func (m *mockPlugin) MapToZen(ctx context.Context, externalData interface{}) (*plugin.TaskData, error) {
	args := m.Called(ctx, externalData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plugin.TaskData), args.Error(1)
}

func (m *mockPlugin) MapToExternal(ctx context.Context, zenData *plugin.TaskData) (interface{}, error) {
	args := m.Called(ctx, zenData)
	return args.Get(0), args.Error(1)
}

func (m *mockPlugin) GetFieldMapping() *plugin.FieldMappingConfig {
	args := m.Called()
	return args.Get(0).(*plugin.FieldMappingConfig)
}

func (m *mockPlugin) GetAuthConfig() *plugin.AuthConfig {
	args := m.Called()
	return args.Get(0).(*plugin.AuthConfig)
}

func (m *mockPlugin) GetRateLimitInfo(ctx context.Context) (*plugin.RateLimitInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*plugin.RateLimitInfo), args.Error(1)
}

func (m *mockPlugin) SupportsOperation(operation plugin.OperationType) bool {
	args := m.Called(operation)
	return args.Bool(0)
}

func TestOperationOrchestrator_ExecuteOperation(t *testing.T) {
	logger := logging.NewBasic()
	orchestrator := NewOperationOrchestrator(logger)

	// Create mock plugin
	mockPlug := &mockPlugin{}
	mockPlug.On("GetRateLimitInfo", mock.Anything).Return(&plugin.RateLimitInfo{
		Limit:     300,
		Remaining: 300,
		ResetTime: time.Now().Add(time.Hour),
	}, nil)

	// Register plugin
	err := orchestrator.RegisterPlugin("test-plugin", mockPlug)
	require.NoError(t, err)

	tests := []struct {
		name          string
		operation     *Operation
		expectSuccess bool
		expectError   bool
	}{
		{
			name: "successful fetch operation",
			operation: &Operation{
				ID:         "op-1",
				PluginName: "test-plugin",
				Type:       OperationTypeFetch,
				ExternalID: "TEST-123",
				Timeout:    30 * time.Second,
			},
			expectSuccess: true,
			expectError:   false,
		},
		{
			name: "successful create operation",
			operation: &Operation{
				ID:         "op-2",
				PluginName: "test-plugin",
				Type:       OperationTypeCreate,
				TaskData: &plugin.TaskData{
					Title: "New Task",
					Type:  "story",
				},
				Timeout: 30 * time.Second,
			},
			expectSuccess: true,
			expectError:   false,
		},
		{
			name: "unsupported operation type",
			operation: &Operation{
				ID:         "op-3",
				PluginName: "test-plugin",
				Type:       OperationType("unsupported"),
				Timeout:    30 * time.Second,
			},
			expectSuccess: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := orchestrator.ExecuteOperation(context.Background(), tt.operation)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expectSuccess, result.Success)
			assert.Equal(t, tt.operation.ID, result.OperationID)
			assert.Equal(t, tt.operation.PluginName, result.PluginName)
			assert.Equal(t, tt.operation.Type, result.Type)
		})
	}

	mockPlug.AssertExpectations(t)
}

func TestOperationOrchestrator_ExecuteTransaction(t *testing.T) {
	logger := logging.NewBasic()
	orchestrator := NewOperationOrchestrator(logger)

	// Create mock plugin
	mockPlug := &mockPlugin{}
	mockPlug.On("GetRateLimitInfo", mock.Anything).Return(&plugin.RateLimitInfo{
		Limit:     300,
		Remaining: 300,
		ResetTime: time.Now().Add(time.Hour),
	}, nil)

	// Register plugin
	err := orchestrator.RegisterPlugin("test-plugin", mockPlug)
	require.NoError(t, err)

	transaction := &Transaction{
		ID: "txn-1",
		Operations: []*Operation{
			{
				ID:         "op-1",
				PluginName: "test-plugin",
				Type:       OperationTypeFetch,
				ExternalID: "TEST-123",
				Timeout:    30 * time.Second,
			},
			{
				ID:         "op-2",
				PluginName: "test-plugin",
				Type:       OperationTypeCreate,
				TaskData: &plugin.TaskData{
					Title: "New Task",
					Type:  "story",
				},
				Timeout: 30 * time.Second,
			},
		},
		Timeout: 60 * time.Second,
	}

	result, err := orchestrator.ExecuteTransaction(context.Background(), transaction)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Equal(t, "txn-1", result.TransactionID)
	assert.Len(t, result.Results, 2)

	mockPlug.AssertExpectations(t)
}

func TestOperationOrchestrator_GetMetrics(t *testing.T) {
	logger := logging.NewBasic()
	orchestrator := NewOperationOrchestrator(logger)

	// Get initial metrics
	metrics := orchestrator.GetOperationMetrics()
	assert.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.TotalOperations)
	assert.Equal(t, int64(0), metrics.SuccessfulOps)
	assert.Equal(t, int64(0), metrics.FailedOps)

	// Execute some operations to update metrics
	mockPlug := &mockPlugin{}
	mockPlug.On("GetRateLimitInfo", mock.Anything).Return(&plugin.RateLimitInfo{
		Limit: 300,
	}, nil)

	err := orchestrator.RegisterPlugin("test-plugin", mockPlug)
	require.NoError(t, err)

	operation := &Operation{
		ID:         "op-1",
		PluginName: "test-plugin",
		Type:       OperationTypeFetch,
		ExternalID: "TEST-123",
		Timeout:    30 * time.Second,
	}

	_, err = orchestrator.ExecuteOperation(context.Background(), operation)
	assert.NoError(t, err)

	// Check updated metrics
	metrics = orchestrator.GetOperationMetrics()
	assert.Equal(t, int64(1), metrics.TotalOperations)
	assert.Equal(t, int64(1), metrics.SuccessfulOps)

	mockPlug.AssertExpectations(t)
}

func TestRetryPolicy_CalculateBackoffDelay(t *testing.T) {
	logger := logging.NewBasic()
	orchestrator := NewOperationOrchestrator(logger)

	tests := []struct {
		name     string
		policy   *RetryPolicy
		attempt  int
		expected time.Duration
	}{
		{
			name: "exponential backoff",
			policy: &RetryPolicy{
				BaseDelay:       time.Second,
				MaxDelay:        30 * time.Second,
				BackoffStrategy: "exponential",
			},
			attempt:  2,
			expected: 4 * time.Second, // 1s * 2^2
		},
		{
			name: "linear backoff",
			policy: &RetryPolicy{
				BaseDelay:       time.Second,
				MaxDelay:        30 * time.Second,
				BackoffStrategy: "linear",
			},
			attempt:  2,
			expected: 3 * time.Second, // 1s * (2+1)
		},
		{
			name: "fixed backoff",
			policy: &RetryPolicy{
				BaseDelay:       5 * time.Second,
				MaxDelay:        30 * time.Second,
				BackoffStrategy: "fixed",
			},
			attempt:  5,
			expected: 5 * time.Second, // Always 5s
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			delay := orchestrator.calculateBackoffDelay(tt.attempt, tt.policy)
			assert.Equal(t, tt.expected, delay)
		})
	}
}
