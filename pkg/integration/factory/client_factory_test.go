package factory

import (
	"context"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/integration/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock auth manager for testing
type mockAuthManager struct {
	mock.Mock
}

func (m *mockAuthManager) GetCredentials(provider string) (string, error) {
	args := m.Called(provider)
	return args.String(0), args.Error(1)
}

func (m *mockAuthManager) Authenticate(ctx context.Context, provider string) error {
	args := m.Called(ctx, provider)
	return args.Error(0)
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

func createTestConfig() *config.Config {
	return &config.Config{
		Integrations: config.IntegrationsConfig{
			TaskSystem:  "jira",
			SyncEnabled: true,
			Providers: map[string]config.IntegrationProviderConfig{
				"jira": {
					URL:         "https://test.atlassian.net",
					ProjectKey:  "TEST",
					Type:        "basic",
					Credentials: "jira_creds",
				},
				"github": {
					URL:         "https://api.github.com",
					ProjectKey:  "owner/repo",
					Type:        "token",
					Credentials: "github_token",
				},
			},
		},
	}
}

func TestClientFactory_CreatePlugin(t *testing.T) {
	logger := logging.NewBasic()
	config := createTestConfig()
	authMgr := &mockAuthManager{}

	// Set up auth manager mock expectations
	authMgr.On("GetCredentials", "jira").Return("Basic dXNlcjpwYXNz", nil)

	factory := NewClientFactory(logger, config, authMgr, nil)

	tests := []struct {
		name          string
		provider      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "create jira plugin",
			provider:    "jira",
			expectError: false,
		},
		{
			name:        "create github plugin",
			provider:    "github",
			expectError: false,
		},
		{
			name:          "unsupported provider",
			provider:      "unsupported",
			expectError:   true,
			errorContains: "provider configuration not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := factory.CreatePlugin(context.Background(), tt.provider)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, plugin)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, plugin)
				assert.Equal(t, tt.provider, plugin.Name())
			}
		})
	}
}

func TestClientFactory_GetPlugin(t *testing.T) {
	logger := logging.NewBasic()
	config := createTestConfig()
	authMgr := &mockAuthManager{}

	// Set up auth manager mock expectations
	authMgr.On("GetCredentials", "jira").Return("Basic dXNlcjpwYXNz", nil)

	factory := NewClientFactory(logger, config, authMgr, nil)

	// Create a plugin first
	plugin, err := factory.CreatePlugin(context.Background(), "jira")
	require.NoError(t, err)
	require.NotNil(t, plugin)

	// Test getting existing plugin
	retrievedPlugin, err := factory.GetPlugin("jira")
	assert.NoError(t, err)
	assert.Equal(t, plugin, retrievedPlugin)

	// Test getting non-existent plugin
	_, err = factory.GetPlugin("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestClientFactory_ListPlugins(t *testing.T) {
	logger := logging.NewBasic()
	config := createTestConfig()
	authMgr := &mockAuthManager{}

	// Set up auth manager mock expectations
	authMgr.On("GetCredentials", "jira").Return("Basic dXNlcjpwYXNz", nil)

	factory := NewClientFactory(logger, config, authMgr, nil)

	// Initially no plugins
	plugins := factory.ListPlugins()
	assert.Empty(t, plugins)

	// Create some plugins
	_, err := factory.CreatePlugin(context.Background(), "jira")
	require.NoError(t, err)

	_, err = factory.CreatePlugin(context.Background(), "github")
	require.NoError(t, err)

	// Check list
	plugins = factory.ListPlugins()
	assert.Len(t, plugins, 2)
	assert.Contains(t, plugins, "jira")
	assert.Contains(t, plugins, "github")
}

func TestClientFactory_ValidatePlugin(t *testing.T) {
	logger := logging.NewBasic()
	config := createTestConfig()
	authMgr := &mockAuthManager{}

	factory := NewClientFactory(logger, config, authMgr, nil)

	validConfig := &plugin.PluginConfig{
		Name:    "jira",
		Version: "1.0.0",
		Enabled: true,
		BaseURL: "https://test.atlassian.net",
		Auth: &plugin.AuthConfig{
			Type:           plugin.AuthTypeBasic,
			CredentialsRef: "jira_creds",
		},
	}

	err := factory.ValidatePlugin(context.Background(), "jira", validConfig)
	assert.NoError(t, err)

	// Test invalid configuration
	invalidConfig := &plugin.PluginConfig{
		Name:    "jira",
		Version: "1.0.0",
		Enabled: true,
		// Missing BaseURL and Auth
	}

	err = factory.ValidatePlugin(context.Background(), "jira", invalidConfig)
	assert.Error(t, err)
}

func TestClientFactory_ShutdownPlugin(t *testing.T) {
	logger := logging.NewBasic()
	config := createTestConfig()
	authMgr := &mockAuthManager{}

	// Set up auth manager mock expectations
	authMgr.On("GetCredentials", "jira").Return("Basic dXNlcjpwYXNz", nil)

	factory := NewClientFactory(logger, config, authMgr, nil)

	// Create a plugin
	plugin, err := factory.CreatePlugin(context.Background(), "jira")
	require.NoError(t, err)
	require.NotNil(t, plugin)

	// Shutdown the plugin
	err = factory.ShutdownPlugin(context.Background(), "jira")
	assert.NoError(t, err)

	// Plugin should no longer be available
	_, err = factory.GetPlugin("jira")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "plugin not found")
}

func TestClientFactory_ShutdownAll(t *testing.T) {
	logger := logging.NewBasic()
	config := createTestConfig()
	authMgr := &mockAuthManager{}

	// Set up auth manager mock expectations
	authMgr.On("GetCredentials", "jira").Return("Basic dXNlcjpwYXNz", nil)

	factory := NewClientFactory(logger, config, authMgr, nil)

	// Create multiple plugins
	_, err := factory.CreatePlugin(context.Background(), "jira")
	require.NoError(t, err)

	_, err = factory.CreatePlugin(context.Background(), "github")
	require.NoError(t, err)

	// Verify plugins exist
	plugins := factory.ListPlugins()
	assert.Len(t, plugins, 2)

	// Shutdown all plugins
	err = factory.ShutdownAll(context.Background())
	assert.NoError(t, err)

	// No plugins should remain
	plugins = factory.ListPlugins()
	assert.Empty(t, plugins)
}

func TestJiraPluginAdapter_Interface(t *testing.T) {
	config := &plugin.PluginConfig{
		Name:    "jira",
		Version: "1.0.0",
		BaseURL: "https://test.atlassian.net",
		Auth: &plugin.AuthConfig{
			Type:           plugin.AuthTypeBasic,
			CredentialsRef: "jira",
		},
		Settings: map[string]interface{}{
			"project_key": "TEST",
		},
	}

	logger := logging.NewBasic()
	authMgr := &mockAuthManager{}

	// Set up auth manager mock expectations
	authMgr.On("GetCredentials", "jira").Return("Basic dXNlcjpwYXNz", nil)

	adapter := &JiraPluginAdapter{
		config:  config,
		logger:  logger,
		authMgr: authMgr,
	}

	// Test interface compliance
	var _ plugin.IntegrationPluginInterface = adapter

	// Test basic methods
	assert.Equal(t, "jira", adapter.Name())
	assert.Equal(t, "1.0.0", adapter.Version())
	assert.NotEmpty(t, adapter.Description())

	// Test lifecycle methods
	err := adapter.Initialize(context.Background(), config)
	assert.NoError(t, err)

	err = adapter.Validate(context.Background())
	assert.NoError(t, err)

	health, err := adapter.HealthCheck(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, health)
	assert.Equal(t, "jira", health.Provider)

	err = adapter.Shutdown(context.Background())
	assert.NoError(t, err)

	// Test operation support
	assert.True(t, adapter.SupportsOperation(plugin.OperationTypeFetch))
	assert.True(t, adapter.SupportsOperation(plugin.OperationTypeCreate))

	// Test configuration methods
	authConfig := adapter.GetAuthConfig()
	assert.NotNil(t, authConfig)
	assert.Equal(t, plugin.AuthTypeBasic, authConfig.Type)

	rateLimitInfo, err := adapter.GetRateLimitInfo(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, rateLimitInfo)

	fieldMapping := adapter.GetFieldMapping()
	assert.NotNil(t, fieldMapping)
}
