package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock runtime for testing
type mockRuntime struct {
	mock.Mock
}

func (m *mockRuntime) LoadPlugin(ctx context.Context, wasmPath string, manifest *Manifest) (PluginInstance, error) {
	args := m.Called(ctx, wasmPath, manifest)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(PluginInstance), args.Error(1)
}

func (m *mockRuntime) UnloadPlugin(instance PluginInstance) error {
	args := m.Called(instance)
	return args.Error(0)
}

func (m *mockRuntime) GetCapabilities() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *mockRuntime) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Mock plugin instance for testing
type mockPluginInstance struct {
	mock.Mock
	name string
}

func (m *mockPluginInstance) Execute(ctx context.Context, function string, args []byte) ([]byte, error) {
	mockArgs := m.Called(ctx, function, args)
	if mockArgs.Get(0) == nil {
		return nil, mockArgs.Error(1)
	}
	return mockArgs.Get(0).([]byte), mockArgs.Error(1)
}

func (m *mockPluginInstance) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockPluginInstance) GetExports() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestRegistry_DiscoverPlugins(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	
	// Create test plugin directory
	pluginDir := filepath.Join(tempDir, "test-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)
	
	// Create test manifest
	manifest := `schema_version: "1.0"
plugin:
  name: "test-plugin"
  version: "1.0.0"
  description: "Test plugin"
  author: "Test Author"
capabilities:
  - "task_sync"
  - "field_mapping"
runtime:
  wasm_file: "plugin.wasm"
  memory_limit: "10MB"
  execution_timeout: "30s"
api_requirements:
  - "http_client"
  - "credential_access"
security:
  permissions:
    - "network.http.outbound"
    - "config.read"
  checksum: "abc123"
configuration_schema:
  server_url:
    type: "string"
    required: true
    description: "Server URL"
`
	
	manifestPath := filepath.Join(pluginDir, "manifest.yaml")
	err = os.WriteFile(manifestPath, []byte(manifest), 0644)
	require.NoError(t, err)
	
	// Create dummy WASM file
	wasmPath := filepath.Join(pluginDir, "plugin.wasm")
	err = os.WriteFile(wasmPath, []byte("dummy wasm content"), 0644)
	require.NoError(t, err)
	
	// Create registry
	logger := logging.NewBasic()
	fsManager := fs.NewManager(logger)
	runtime := &mockRuntime{}
	runtime.On("GetCapabilities").Return([]string{"wasm_execution"})
	
	registry := NewRegistry(logger, fsManager, []string{tempDir}, runtime)
	
	// Discover plugins
	err = registry.DiscoverPlugins(context.Background())
	require.NoError(t, err)
	
	// Verify plugin was discovered
	plugins := registry.ListPlugins()
	assert.Len(t, plugins, 1)
	assert.Contains(t, plugins, "test-plugin")
	
	plugin := plugins["test-plugin"]
	assert.Equal(t, "test-plugin", plugin.Manifest.Plugin.Name)
	assert.Equal(t, "1.0.0", plugin.Manifest.Plugin.Version)
	assert.Equal(t, PluginStatusDiscovered, plugin.Status)
	assert.Equal(t, wasmPath, plugin.WASMPath)
}

func TestRegistry_LoadPlugin(t *testing.T) {
	// Create temporary directory structure
	tempDir := t.TempDir()
	pluginDir := filepath.Join(tempDir, "test-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)
	
	// Create test files
	manifest := `schema_version: "1.0"
plugin:
  name: "test-plugin"
  version: "1.0.0"
  description: "Test plugin"
capabilities:
  - "task_sync"
runtime:
  wasm_file: "plugin.wasm"
  memory_limit: "10MB"
  execution_timeout: "30s"
api_requirements:
  - "http_client"
security:
  permissions:
    - "network.http.outbound"
  checksum: "abc123"
`
	
	err = os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), []byte(manifest), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(filepath.Join(pluginDir, "plugin.wasm"), []byte("dummy wasm"), 0644)
	require.NoError(t, err)
	
	// Create mocks
	logger := logging.NewBasic()
	fsManager := fs.NewManager(logger)
	runtime := &mockRuntime{}
	instance := &mockPluginInstance{name: "test-plugin"}
	
	runtime.On("LoadPlugin", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*plugin.Manifest")).Return(instance, nil)
	runtime.On("GetCapabilities").Return([]string{"wasm_execution"})
	
	registry := NewRegistry(logger, fsManager, []string{tempDir}, runtime)
	
	// Discover plugins first
	err = registry.DiscoverPlugins(context.Background())
	require.NoError(t, err)
	
	// Load plugin
	err = registry.LoadPlugin(context.Background(), "test-plugin")
	require.NoError(t, err)
	
	// Verify plugin is loaded
	loaded, err := registry.GetLoadedPlugin("test-plugin")
	require.NoError(t, err)
	assert.Equal(t, "test-plugin", loaded.Info.Manifest.Plugin.Name)
	assert.Equal(t, instance, loaded.Instance)
	
	// Verify plugin status updated
	pluginInfo, err := registry.GetPlugin("test-plugin")
	require.NoError(t, err)
	assert.Equal(t, PluginStatusLoaded, pluginInfo.Status)
	
	runtime.AssertExpectations(t)
}

func TestRegistry_UnloadPlugin(t *testing.T) {
	logger := logging.NewBasic()
	fsManager := fs.NewManager(logger)
	runtime := &mockRuntime{}
	instance := &mockPluginInstance{name: "test-plugin"}
	
	runtime.On("UnloadPlugin", instance).Return(nil)
	runtime.On("GetCapabilities").Return([]string{"wasm_execution"})
	
	registry := NewRegistry(logger, fsManager, []string{}, runtime)
	
	// Manually add a loaded plugin
	pluginInfo := &PluginInfo{
		Manifest: &Manifest{
			Plugin: PluginMetadata{Name: "test-plugin"},
		},
		Status: PluginStatusLoaded,
	}
	
	registry.mu.Lock()
	registry.plugins["test-plugin"] = pluginInfo
	registry.mu.Unlock()
	
	registry.runtimeMu.Lock()
	registry.loadedPlugins["test-plugin"] = LoadedPlugin{
		Info:     pluginInfo,
		Instance: instance,
		LoadedAt: time.Now(),
	}
	registry.runtimeMu.Unlock()
	
	// Unload plugin
	err := registry.UnloadPlugin("test-plugin")
	require.NoError(t, err)
	
	// Verify plugin is unloaded
	_, err = registry.GetLoadedPlugin("test-plugin")
	assert.Error(t, err)
	
	// Verify plugin status updated
	pluginInfo, err = registry.GetPlugin("test-plugin")
	require.NoError(t, err)
	assert.Equal(t, PluginStatusDiscovered, pluginInfo.Status)
	
	runtime.AssertExpectations(t)
}

func TestRegistry_ValidateManifest(t *testing.T) {
	logger := logging.NewBasic()
	fsManager := fs.NewManager(logger)
	runtime := &mockRuntime{}
	registry := NewRegistry(logger, fsManager, []string{}, runtime)
	
	tests := []struct {
		name     string
		manifest Manifest
		wantErr  bool
	}{
		{
			name: "valid manifest",
			manifest: Manifest{
				SchemaVersion: "1.0",
				Plugin: PluginMetadata{
					Name:    "test-plugin",
					Version: "1.0.0",
				},
				Runtime: RuntimeConfig{
					WASMFile: "plugin.wasm",
				},
				Capabilities: []string{"task_sync"},
				APIRequirements: []string{"http_client"},
			},
			wantErr: false,
		},
		{
			name: "missing plugin name",
			manifest: Manifest{
				SchemaVersion: "1.0",
				Plugin: PluginMetadata{
					Version: "1.0.0",
				},
				Runtime: RuntimeConfig{
					WASMFile: "plugin.wasm",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid capability",
			manifest: Manifest{
				SchemaVersion: "1.0",
				Plugin: PluginMetadata{
					Name:    "test-plugin",
					Version: "1.0.0",
				},
				Runtime: RuntimeConfig{
					WASMFile: "plugin.wasm",
				},
				Capabilities: []string{"invalid_capability"},
			},
			wantErr: true,
		},
		{
			name: "invalid API requirement",
			manifest: Manifest{
				SchemaVersion: "1.0",
				Plugin: PluginMetadata{
					Name:    "test-plugin",
					Version: "1.0.0",
				},
				Runtime: RuntimeConfig{
					WASMFile: "plugin.wasm",
				},
				APIRequirements: []string{"invalid_api"},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.validateManifest(&tt.manifest)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistry_RefreshPlugins(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()
	
	logger := logging.NewBasic()
	fsManager := fs.NewManager(logger)
	runtime := &mockRuntime{}
	runtime.On("GetCapabilities").Return([]string{"wasm_execution"})
	
	registry := NewRegistry(logger, fsManager, []string{tempDir}, runtime)
	
	// Initial discovery should find no plugins
	err := registry.DiscoverPlugins(context.Background())
	require.NoError(t, err)
	assert.Len(t, registry.ListPlugins(), 0)
	
	// Create a plugin
	pluginDir := filepath.Join(tempDir, "new-plugin")
	err = os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)
	
	manifest := `schema_version: "1.0"
plugin:
  name: "new-plugin"
  version: "1.0.0"
  description: "New plugin"
capabilities:
  - "task_sync"
runtime:
  wasm_file: "plugin.wasm"
security:
  permissions:
    - "network.http.outbound"
  checksum: "def456"
`
	
	err = os.WriteFile(filepath.Join(pluginDir, "manifest.yaml"), []byte(manifest), 0644)
	require.NoError(t, err)
	
	err = os.WriteFile(filepath.Join(pluginDir, "plugin.wasm"), []byte("dummy"), 0644)
	require.NoError(t, err)
	
	// Refresh should discover the new plugin
	err = registry.RefreshPlugins(context.Background())
	require.NoError(t, err)
	
	plugins := registry.ListPlugins()
	assert.Len(t, plugins, 1)
	assert.Contains(t, plugins, "new-plugin")
}

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		name        string
		permissions []string
		wantErr     bool
	}{
		{
			name: "valid permissions",
			permissions: []string{
				string(PermissionNetworkHTTPOutbound),
				string(PermissionConfigRead),
				string(PermissionLogging),
			},
			wantErr: false,
		},
		{
			name: "invalid permission",
			permissions: []string{
				string(PermissionNetworkHTTPOutbound),
				"invalid.permission",
			},
			wantErr: true,
		},
		{
			name:        "empty permissions",
			permissions: []string{},
			wantErr:     false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePermissions(tt.permissions)
			if tt.wantErr {
				assert.Error(t, err)
				assert.IsType(t, &PluginError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCapabilities(t *testing.T) {
	tests := []struct {
		name         string
		capabilities []string
		wantErr      bool
	}{
		{
			name: "valid capabilities",
			capabilities: []string{
				string(CapabilityTaskSync),
				string(CapabilityFieldMapping),
				string(CapabilityWebhookSupport),
			},
			wantErr: false,
		},
		{
			name: "invalid capability",
			capabilities: []string{
				string(CapabilityTaskSync),
				"invalid_capability",
			},
			wantErr: true,
		},
		{
			name:         "empty capabilities",
			capabilities: []string{},
			wantErr:      false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCapabilities(tt.capabilities)
			if tt.wantErr {
				assert.Error(t, err)
				assert.IsType(t, &PluginError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDefaultResourceLimits(t *testing.T) {
	limits := DefaultResourceLimits()
	
	assert.Equal(t, 10, limits.MaxMemoryMB)
	assert.Equal(t, 30*time.Second, limits.MaxExecutionTime)
	assert.Equal(t, 50.0, limits.MaxCPUPercent)
	assert.Equal(t, 1, limits.MaxInstances)
}

func TestPluginError(t *testing.T) {
	err := &PluginError{
		Code:      ErrCodePluginLoadFailed,
		Message:   "Plugin load failed",
		Plugin:    "test-plugin",
		Timestamp: time.Now(),
	}
	
	assert.Equal(t, "Plugin load failed", err.Error())
	assert.Equal(t, ErrCodePluginLoadFailed, err.Code)
	assert.Equal(t, "test-plugin", err.Plugin)
}

func TestRegistry_ErrorHandling(t *testing.T) {
	// Create temporary directory with invalid plugin
	tempDir := t.TempDir()
	pluginDir := filepath.Join(tempDir, "invalid-plugin")
	err := os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)
	
	// Create invalid manifest (missing required fields)
	invalidManifest := `schema_version: "1.0"
plugin:
  description: "Invalid plugin - missing name"
`
	
	manifestPath := filepath.Join(pluginDir, "manifest.yaml")
	err = os.WriteFile(manifestPath, []byte(invalidManifest), 0644)
	require.NoError(t, err)
	
	// Create registry
	logger := logging.NewBasic()
	fsManager := fs.NewManager(logger)
	runtime := &mockRuntime{}
	runtime.On("GetCapabilities").Return([]string{"wasm_execution"})
	
	registry := NewRegistry(logger, fsManager, []string{tempDir}, runtime)
	
	// Discovery should handle invalid plugin gracefully
	err = registry.DiscoverPlugins(context.Background())
	require.NoError(t, err)
	
	// Should have discovered the plugin but marked it as error
	plugins := registry.ListPlugins()
	assert.Len(t, plugins, 1)
	
	// Plugin should be in error state
	plugin := plugins["invalid-plugin"]
	assert.Equal(t, PluginStatusError, plugin.Status)
	assert.NotEmpty(t, plugin.Error)
}

func TestRegistry_ConcurrentAccess(t *testing.T) {
	logger := logging.NewBasic()
	fsManager := fs.NewManager(logger)
	runtime := &mockRuntime{}
	runtime.On("GetCapabilities").Return([]string{"wasm_execution"})
	
	registry := NewRegistry(logger, fsManager, []string{}, runtime)
	
	// Test concurrent access to plugin registry
	done := make(chan bool, 10)
	
	// Start multiple goroutines accessing the registry
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// Simulate plugin operations
			plugins := registry.ListPlugins()
			_ = plugins
			
			loadedPlugins := registry.ListLoadedPlugins()
			_ = loadedPlugins
			
			// Try to get non-existent plugin
			_, err := registry.GetPlugin("non-existent")
			assert.Error(t, err)
		}(i)
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestRegistry_Close(t *testing.T) {
	logger := logging.NewBasic()
	fsManager := fs.NewManager(logger)
	runtime := &mockRuntime{}
	instance := &mockPluginInstance{name: "test-plugin"}
	
	runtime.On("UnloadPlugin", instance).Return(nil)
	runtime.On("Close").Return(nil)
	runtime.On("GetCapabilities").Return([]string{"wasm_execution"})
	
	registry := NewRegistry(logger, fsManager, []string{}, runtime)
	
	// Add a loaded plugin
	registry.runtimeMu.Lock()
	registry.loadedPlugins["test-plugin"] = LoadedPlugin{
		Instance: instance,
		LoadedAt: time.Now(),
	}
	registry.runtimeMu.Unlock()
	
	// Close registry
	err := registry.Close()
	require.NoError(t, err)
	
	// Verify all plugins were unloaded and runtime was closed
	runtime.AssertExpectations(t)
}
