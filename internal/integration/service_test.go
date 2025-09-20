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
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "jira",
		SyncEnabled: true,
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
			cfg := &config.IntegrationsConfig{
				TaskSystem:  tt.taskSystem,
				SyncEnabled: true,
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
	t.Skip("Skipping sync task test - full implementation pending")
}

func TestService_SyncTask_NotConfigured(t *testing.T) {
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "",
		SyncEnabled: false,
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
	cfg := &config.IntegrationsConfig{
		TaskSystem:  "jira",
		SyncEnabled: true,
	}
	return createTestServiceWithConfig(t, cfg)
}

func createTestServiceWithConfig(t *testing.T, cfg *config.IntegrationsConfig) *Service {
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
