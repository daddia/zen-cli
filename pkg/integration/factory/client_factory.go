package factory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/clients/jira"
	"github.com/daddia/zen/pkg/integration/plugin"
	pluginpkg "github.com/daddia/zen/pkg/plugin"
)

// ClientFactory creates and manages integration plugin instances
type ClientFactory struct {
	logger   logging.Logger
	config   *config.Config
	authMgr  auth.Manager
	registry PluginRegistryInterface
	plugins  map[string]plugin.IntegrationPluginInterface
	mu       sync.RWMutex
}

// ClientFactoryInterface defines the client factory interface
type ClientFactoryInterface interface {
	// CreatePlugin creates a new plugin instance
	CreatePlugin(ctx context.Context, providerName string) (plugin.IntegrationPluginInterface, error)

	// GetPlugin returns an existing plugin instance
	GetPlugin(providerName string) (plugin.IntegrationPluginInterface, error)

	// ListPlugins returns all available plugins
	ListPlugins() []string

	// ConfigurePlugin configures a plugin with provider-specific settings
	ConfigurePlugin(ctx context.Context, providerName string, config *plugin.PluginConfig) error

	// ValidatePlugin validates a plugin configuration
	ValidatePlugin(ctx context.Context, providerName string, config *plugin.PluginConfig) error

	// ShutdownPlugin shuts down a specific plugin
	ShutdownPlugin(ctx context.Context, providerName string) error

	// ShutdownAll shuts down all plugins
	ShutdownAll(ctx context.Context) error
}

// PluginConfiguratorInterface defines plugin configuration interface
type PluginConfiguratorInterface interface {
	// LoadConfiguration loads plugin configuration from various sources
	LoadConfiguration(providerName string) (*plugin.PluginConfig, error)

	// ValidateConfiguration validates plugin configuration
	ValidateConfiguration(config *plugin.PluginConfig) error

	// ApplyDefaults applies default values to configuration
	ApplyDefaults(config *plugin.PluginConfig) error
}

// DependencyInjectorInterface defines dependency injection interface
type DependencyInjectorInterface interface {
	// InjectDependencies injects required dependencies into plugin
	InjectDependencies(plugin plugin.IntegrationPluginInterface, config *plugin.PluginConfig) error

	// ResolveDependencies resolves plugin dependencies
	ResolveDependencies(pluginName string) (map[string]interface{}, error)
}

// PluginRegistryInterface defines the plugin registry interface
type PluginRegistryInterface interface {
	GetPlugin(pluginName string) (*pluginpkg.PluginInfo, error)
	ListPlugins() map[string]*pluginpkg.PluginInfo
	LoadPlugin(ctx context.Context, pluginName string) error
}

// NewClientFactory creates a new client factory
func NewClientFactory(
	logger logging.Logger,
	config *config.Config,
	authMgr auth.Manager,
	registry PluginRegistryInterface,
) *ClientFactory {
	return &ClientFactory{
		logger:   logger,
		config:   config,
		authMgr:  authMgr,
		registry: registry,
		plugins:  make(map[string]plugin.IntegrationPluginInterface),
	}
}

// CreatePlugin creates a new plugin instance
func (f *ClientFactory) CreatePlugin(ctx context.Context, providerName string) (plugin.IntegrationPluginInterface, error) {
	f.logger.Debug("creating plugin instance", "provider", providerName)

	// Check if plugin already exists
	f.mu.RLock()
	if existing, exists := f.plugins[providerName]; exists {
		f.mu.RUnlock()
		return existing, nil
	}
	f.mu.RUnlock()

	// Load plugin configuration
	pluginConfig, err := f.loadPluginConfiguration(providerName)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin configuration: %w", err)
	}

	// Validate configuration
	if err := f.validatePluginConfiguration(pluginConfig); err != nil {
		return nil, fmt.Errorf("invalid plugin configuration: %w", err)
	}

	// Create plugin instance based on type
	var pluginInstance plugin.IntegrationPluginInterface

	// For now, we'll create instances based on provider name
	// In the future, this would use the plugin registry to load WASM plugins
	switch providerName {
	case "jira":
		pluginInstance = f.createJiraPlugin(pluginConfig)
	case "github":
		pluginInstance = f.createGitHubPlugin(pluginConfig)
	case "linear":
		pluginInstance = f.createLinearPlugin(pluginConfig)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Initialize plugin
	if err := pluginInstance.Initialize(ctx, pluginConfig); err != nil {
		return nil, fmt.Errorf("failed to initialize plugin: %w", err)
	}

	// Store plugin instance
	f.mu.Lock()
	f.plugins[providerName] = pluginInstance
	f.mu.Unlock()

	f.logger.Info("plugin instance created successfully", "provider", providerName)

	return pluginInstance, nil
}

// GetPlugin returns an existing plugin instance
func (f *ClientFactory) GetPlugin(providerName string) (plugin.IntegrationPluginInterface, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	pluginInstance, exists := f.plugins[providerName]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", providerName)
	}

	return pluginInstance, nil
}

// ListPlugins returns all available plugins
func (f *ClientFactory) ListPlugins() []string {
	f.mu.RLock()
	defer f.mu.RUnlock()

	plugins := make([]string, 0, len(f.plugins))
	for name := range f.plugins {
		plugins = append(plugins, name)
	}

	return plugins
}

// ConfigurePlugin configures a plugin with provider-specific settings
func (f *ClientFactory) ConfigurePlugin(ctx context.Context, providerName string, config *plugin.PluginConfig) error {
	pluginInstance, err := f.GetPlugin(providerName)
	if err != nil {
		return fmt.Errorf("plugin not found: %w", err)
	}

	// Validate new configuration
	if err := f.validatePluginConfiguration(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Re-initialize plugin with new configuration
	if err := pluginInstance.Initialize(ctx, config); err != nil {
		return fmt.Errorf("failed to reconfigure plugin: %w", err)
	}

	f.logger.Info("plugin reconfigured successfully", "provider", providerName)

	return nil
}

// ValidatePlugin validates a plugin configuration
func (f *ClientFactory) ValidatePlugin(ctx context.Context, providerName string, config *plugin.PluginConfig) error {
	// Validate configuration structure
	if err := f.validatePluginConfiguration(config); err != nil {
		return err
	}

	// Create temporary plugin instance for validation
	var pluginInstance plugin.IntegrationPluginInterface

	switch providerName {
	case "jira":
		pluginInstance = f.createJiraPlugin(config)
	case "github":
		pluginInstance = f.createGitHubPlugin(config)
	case "linear":
		pluginInstance = f.createLinearPlugin(config)
	default:
		return fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Validate plugin
	return pluginInstance.Validate(ctx)
}

// ShutdownPlugin shuts down a specific plugin
func (f *ClientFactory) ShutdownPlugin(ctx context.Context, providerName string) error {
	f.mu.Lock()
	pluginInstance, exists := f.plugins[providerName]
	if exists {
		delete(f.plugins, providerName)
	}
	f.mu.Unlock()

	if !exists {
		return fmt.Errorf("plugin not found: %s", providerName)
	}

	if err := pluginInstance.Shutdown(ctx); err != nil {
		f.logger.Warn("plugin shutdown failed", "provider", providerName, "error", err)
		return err
	}

	f.logger.Info("plugin shutdown completed", "provider", providerName)

	return nil
}

// ShutdownAll shuts down all plugins
func (f *ClientFactory) ShutdownAll(ctx context.Context) error {
	f.mu.Lock()
	plugins := make(map[string]plugin.IntegrationPluginInterface)
	for name, instance := range f.plugins {
		plugins[name] = instance
	}
	f.plugins = make(map[string]plugin.IntegrationPluginInterface)
	f.mu.Unlock()

	var lastError error
	for name, instance := range plugins {
		if err := instance.Shutdown(ctx); err != nil {
			f.logger.Warn("plugin shutdown failed", "provider", name, "error", err)
			lastError = err
		}
	}

	f.logger.Info("all plugins shutdown completed")

	return lastError
}

// Helper methods

// loadPluginConfiguration loads configuration for a specific provider
func (f *ClientFactory) loadPluginConfiguration(providerName string) (*plugin.PluginConfig, error) {
	// Get provider configuration from main config
	providerConfig, exists := f.config.Integrations.Providers[providerName]
	if !exists {
		return nil, fmt.Errorf("provider configuration not found: %s", providerName)
	}

	// Convert to plugin configuration
	pluginConfig := &plugin.PluginConfig{
		Name:       providerName,
		Version:    "1.0.0", // Default version
		Enabled:    true,
		BaseURL:    providerConfig.URL,
		Timeout:    30 * time.Second, // Default timeout
		MaxRetries: 3,                // Default retries
		Auth: &plugin.AuthConfig{
			Type:           plugin.AuthType(providerConfig.Type),
			CredentialsRef: providerName, // Use provider name as credentials reference
			CustomFields: map[string]string{
				"email":   providerConfig.Email,
				"api_key": providerConfig.APIKey,
			},
		},
		Settings: make(map[string]interface{}),
	}

	// Apply provider-specific settings
	if providerConfig.ProjectKey != "" {
		pluginConfig.Settings["project_key"] = providerConfig.ProjectKey
	}

	return pluginConfig, nil
}

// validatePluginConfiguration validates plugin configuration
func (f *ClientFactory) validatePluginConfiguration(config *plugin.PluginConfig) error {
	if config.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if config.BaseURL == "" {
		return fmt.Errorf("base URL is required")
	}

	if config.Auth == nil {
		return fmt.Errorf("authentication configuration is required")
	}

	if config.Auth.CredentialsRef == "" {
		return fmt.Errorf("credentials reference is required")
	}

	return nil
}

// Plugin creation methods - these would be replaced by WASM plugin loading in the future

// createJiraPlugin creates a Jira plugin instance
func (f *ClientFactory) createJiraPlugin(config *plugin.PluginConfig) plugin.IntegrationPluginInterface {
	adapter := &JiraPluginAdapter{
		config:  config,
		logger:  f.logger,
		authMgr: f.authMgr,
	}

	// Initialize the adapter (this will create the real Jira plugin)
	if err := adapter.Initialize(context.Background(), config); err != nil {
		f.logger.Warn("failed to initialize Jira plugin adapter", "error", err)
	}

	return adapter
}

// createGitHubPlugin creates a GitHub plugin instance
func (f *ClientFactory) createGitHubPlugin(config *plugin.PluginConfig) plugin.IntegrationPluginInterface {
	// This is a placeholder for future GitHub plugin implementation
	return &GitHubPluginAdapter{
		config:  config,
		logger:  f.logger,
		authMgr: f.authMgr,
	}
}

// createLinearPlugin creates a Linear plugin instance
func (f *ClientFactory) createLinearPlugin(config *plugin.PluginConfig) plugin.IntegrationPluginInterface {
	// This is a placeholder for future Linear plugin implementation
	return &LinearPluginAdapter{
		config:  config,
		logger:  f.logger,
		authMgr: f.authMgr,
	}
}

// Plugin adapters - these would be replaced by actual WASM plugins

// JiraPluginAdapter adapts the existing Jira client to the plugin interface
type JiraPluginAdapter struct {
	config     *plugin.PluginConfig
	logger     logging.Logger
	authMgr    auth.Manager
	jiraPlugin *jira.Plugin // Real Jira plugin implementation
}

// Implement plugin interface methods for Jira adapter
func (j *JiraPluginAdapter) Name() string        { return "jira" }
func (j *JiraPluginAdapter) Version() string     { return j.config.Version }
func (j *JiraPluginAdapter) Description() string { return "Jira integration plugin adapter" }

func (j *JiraPluginAdapter) Initialize(ctx context.Context, config *plugin.PluginConfig) error {
	j.config = config

	// Create the real Jira plugin
	jiraConfig := &jira.PluginConfig{
		Name:       config.Name,
		Version:    config.Version,
		Enabled:    config.Enabled,
		BaseURL:    config.BaseURL,
		ProjectKey: getProjectKeyFromSettings(config.Settings), // Extract project key from settings
		Timeout:    config.Timeout,
		MaxRetries: config.MaxRetries,
		Headers:    config.Headers,
		Auth: &jira.AuthConfig{
			Type:           jira.AuthType(config.Auth.Type),
			CredentialsRef: config.Auth.CredentialsRef,
			CustomFields:   config.Auth.CustomFields, // Pass through the custom fields
		},
		Settings: config.Settings,
	}

	j.jiraPlugin = jira.NewPlugin(jiraConfig, j.logger, j.authMgr)

	// Initialize the real plugin
	return j.jiraPlugin.Initialize(ctx, jiraConfig)
}

func (j *JiraPluginAdapter) Validate(ctx context.Context) error {
	// Basic validation - would be enhanced with actual Jira connectivity
	return nil
}

func (j *JiraPluginAdapter) HealthCheck(ctx context.Context) (*plugin.PluginHealth, error) {
	return &plugin.PluginHealth{
		Provider:    j.Name(),
		Healthy:     true,
		LastChecked: time.Now(),
	}, nil
}

func (j *JiraPluginAdapter) Shutdown(ctx context.Context) error {
	return nil
}

func (j *JiraPluginAdapter) FetchTask(ctx context.Context, externalID string, opts *plugin.FetchOptions) (*plugin.TaskData, error) {
	j.logger.Debug("fetching task from Jira adapter", "external_id", externalID)

	// Use the real Jira plugin
	if j.jiraPlugin == nil {
		return nil, fmt.Errorf("Jira plugin not initialized")
	}

	jiraTaskData, err := j.jiraPlugin.FetchTask(ctx, externalID, &jira.FetchOptions{
		IncludeRaw: opts.IncludeRaw,
		Timeout:    opts.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch from Jira: %w", err)
	}

	// Convert jira.PluginTaskData to plugin.TaskData
	return j.convertJiraTaskDataToPluginTaskData(jiraTaskData), nil
}

// convertJiraTaskDataToPluginTaskData converts Jira plugin data to standard plugin data
func (j *JiraPluginAdapter) convertJiraTaskDataToPluginTaskData(jiraData *jira.PluginTaskData) *plugin.TaskData {
	return &plugin.TaskData{
		ID:          jiraData.ID,
		ExternalID:  jiraData.ExternalID,
		Title:       jiraData.Title,
		Description: jiraData.Description,
		Status:      jiraData.Status,
		Priority:    jiraData.Priority,
		Type:        jiraData.Type,
		Owner:       jiraData.Owner,
		Assignee:    jiraData.Assignee,
		Team:        jiraData.Team,
		Created:     jiraData.Created,
		Updated:     jiraData.Updated,
		DueDate:     jiraData.DueDate,
		Labels:      jiraData.Labels,
		Tags:        jiraData.Tags,
		Components:  jiraData.Components,
		ExternalURL: jiraData.ExternalURL,
		RawData:     jiraData.RawData,
		Metadata:    jiraData.Metadata,
		Version:     jiraData.Version,
		Checksum:    jiraData.Checksum,
	}
}

func (j *JiraPluginAdapter) CreateTask(ctx context.Context, taskData *plugin.TaskData, opts *plugin.CreateOptions) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (j *JiraPluginAdapter) UpdateTask(ctx context.Context, externalID string, taskData *plugin.TaskData, opts *plugin.UpdateOptions) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (j *JiraPluginAdapter) DeleteTask(ctx context.Context, externalID string, opts *plugin.DeleteOptions) error {
	return fmt.Errorf("not implemented")
}

func (j *JiraPluginAdapter) SearchTasks(ctx context.Context, query *plugin.SearchQuery, opts *plugin.SearchOptions) ([]*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (j *JiraPluginAdapter) SyncTask(ctx context.Context, taskID string, opts *plugin.SyncOptions) (*plugin.SyncResult, error) {
	return nil, fmt.Errorf("not implemented")
}

func (j *JiraPluginAdapter) GetSyncMetadata(ctx context.Context, taskID string) (*plugin.SyncMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}

func (j *JiraPluginAdapter) MapToZen(ctx context.Context, externalData interface{}) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}

func (j *JiraPluginAdapter) MapToExternal(ctx context.Context, zenData *plugin.TaskData) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}

func (j *JiraPluginAdapter) GetFieldMapping() *plugin.FieldMappingConfig {
	return &plugin.FieldMappingConfig{}
}

func (j *JiraPluginAdapter) GetAuthConfig() *plugin.AuthConfig {
	return j.config.Auth
}

func (j *JiraPluginAdapter) GetRateLimitInfo(ctx context.Context) (*plugin.RateLimitInfo, error) {
	return &plugin.RateLimitInfo{
		Limit:     300,
		Remaining: 300,
		ResetTime: time.Now().Add(time.Hour),
	}, nil
}

func (j *JiraPluginAdapter) SupportsOperation(operation plugin.OperationType) bool {
	return true
}

// Placeholder adapters for other providers

type GitHubPluginAdapter struct {
	config  *plugin.PluginConfig
	logger  logging.Logger
	authMgr auth.Manager
}

// Implement basic interface methods for GitHub adapter
func (g *GitHubPluginAdapter) Name() string        { return "github" }
func (g *GitHubPluginAdapter) Version() string     { return g.config.Version }
func (g *GitHubPluginAdapter) Description() string { return "GitHub integration plugin adapter" }
func (g *GitHubPluginAdapter) Initialize(ctx context.Context, config *plugin.PluginConfig) error {
	return nil
}
func (g *GitHubPluginAdapter) Validate(ctx context.Context) error { return nil }
func (g *GitHubPluginAdapter) HealthCheck(ctx context.Context) (*plugin.PluginHealth, error) {
	return &plugin.PluginHealth{Provider: g.Name(), Healthy: true, LastChecked: time.Now()}, nil
}
func (g *GitHubPluginAdapter) Shutdown(ctx context.Context) error { return nil }
func (g *GitHubPluginAdapter) FetchTask(ctx context.Context, externalID string, opts *plugin.FetchOptions) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) CreateTask(ctx context.Context, taskData *plugin.TaskData, opts *plugin.CreateOptions) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) UpdateTask(ctx context.Context, externalID string, taskData *plugin.TaskData, opts *plugin.UpdateOptions) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) DeleteTask(ctx context.Context, externalID string, opts *plugin.DeleteOptions) error {
	return fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) SearchTasks(ctx context.Context, query *plugin.SearchQuery, opts *plugin.SearchOptions) ([]*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) SyncTask(ctx context.Context, taskID string, opts *plugin.SyncOptions) (*plugin.SyncResult, error) {
	return nil, fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) GetSyncMetadata(ctx context.Context, taskID string) (*plugin.SyncMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) MapToZen(ctx context.Context, externalData interface{}) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) MapToExternal(ctx context.Context, zenData *plugin.TaskData) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (g *GitHubPluginAdapter) GetFieldMapping() *plugin.FieldMappingConfig {
	return &plugin.FieldMappingConfig{}
}
func (g *GitHubPluginAdapter) GetAuthConfig() *plugin.AuthConfig { return g.config.Auth }
func (g *GitHubPluginAdapter) GetRateLimitInfo(ctx context.Context) (*plugin.RateLimitInfo, error) {
	return &plugin.RateLimitInfo{}, nil
}
func (g *GitHubPluginAdapter) SupportsOperation(operation plugin.OperationType) bool { return false }

type LinearPluginAdapter struct {
	config  *plugin.PluginConfig
	logger  logging.Logger
	authMgr auth.Manager
}

// Implement basic interface methods for Linear adapter
func (l *LinearPluginAdapter) Name() string        { return "linear" }
func (l *LinearPluginAdapter) Version() string     { return l.config.Version }
func (l *LinearPluginAdapter) Description() string { return "Linear integration plugin adapter" }
func (l *LinearPluginAdapter) Initialize(ctx context.Context, config *plugin.PluginConfig) error {
	return nil
}
func (l *LinearPluginAdapter) Validate(ctx context.Context) error { return nil }
func (l *LinearPluginAdapter) HealthCheck(ctx context.Context) (*plugin.PluginHealth, error) {
	return &plugin.PluginHealth{Provider: l.Name(), Healthy: true, LastChecked: time.Now()}, nil
}
func (l *LinearPluginAdapter) Shutdown(ctx context.Context) error { return nil }
func (l *LinearPluginAdapter) FetchTask(ctx context.Context, externalID string, opts *plugin.FetchOptions) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) CreateTask(ctx context.Context, taskData *plugin.TaskData, opts *plugin.CreateOptions) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) UpdateTask(ctx context.Context, externalID string, taskData *plugin.TaskData, opts *plugin.UpdateOptions) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) DeleteTask(ctx context.Context, externalID string, opts *plugin.DeleteOptions) error {
	return fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) SearchTasks(ctx context.Context, query *plugin.SearchQuery, opts *plugin.SearchOptions) ([]*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) SyncTask(ctx context.Context, taskID string, opts *plugin.SyncOptions) (*plugin.SyncResult, error) {
	return nil, fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) GetSyncMetadata(ctx context.Context, taskID string) (*plugin.SyncMetadata, error) {
	return nil, fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) MapToZen(ctx context.Context, externalData interface{}) (*plugin.TaskData, error) {
	return nil, fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) MapToExternal(ctx context.Context, zenData *plugin.TaskData) (interface{}, error) {
	return nil, fmt.Errorf("not implemented")
}
func (l *LinearPluginAdapter) GetFieldMapping() *plugin.FieldMappingConfig {
	return &plugin.FieldMappingConfig{}
}
func (l *LinearPluginAdapter) GetAuthConfig() *plugin.AuthConfig { return l.config.Auth }
func (l *LinearPluginAdapter) GetRateLimitInfo(ctx context.Context) (*plugin.RateLimitInfo, error) {
	return &plugin.RateLimitInfo{}, nil
}
func (l *LinearPluginAdapter) SupportsOperation(operation plugin.OperationType) bool { return false }

// Helper functions

// getProjectKeyFromSettings extracts project key from plugin settings
func getProjectKeyFromSettings(settings map[string]interface{}) string {
	if projectKey, ok := settings["project_key"].(string); ok {
		return projectKey
	}
	return ""
}
