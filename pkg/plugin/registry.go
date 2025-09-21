package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/fs"
	"gopkg.in/yaml.v3"
)

// Registry manages plugin discovery, loading, and lifecycle
type Registry struct {
	logger     logging.Logger
	fsManager  fs.Manager
	pluginDirs []string
	plugins    map[string]*PluginInfo
	mu         sync.RWMutex

	// Plugin runtime management
	runtime       Runtime
	loadedPlugins map[string]LoadedPlugin
	runtimeMu     sync.RWMutex
}

// PluginInfo contains metadata about a discovered plugin
type PluginInfo struct {
	Manifest     *Manifest    `json:"manifest"`
	Path         string       `json:"path"`
	WASMPath     string       `json:"wasm_path"`
	Discovered   time.Time    `json:"discovered"`
	LastModified time.Time    `json:"last_modified"`
	Status       PluginStatus `json:"status"`
	Error        string       `json:"error,omitempty"`
}

// PluginStatus represents the current status of a plugin
type PluginStatus string

const (
	PluginStatusDiscovered PluginStatus = "discovered"
	PluginStatusLoaded     PluginStatus = "loaded"
	PluginStatusError      PluginStatus = "error"
	PluginStatusDisabled   PluginStatus = "disabled"
)

// LoadedPlugin represents a plugin that has been loaded into the runtime
type LoadedPlugin struct {
	Info     *PluginInfo    `json:"info"`
	Instance PluginInstance `json:"-"`
	LoadedAt time.Time      `json:"loaded_at"`
}

// PluginInstance represents a running plugin instance
type PluginInstance interface {
	// Execute calls a function in the plugin with the given arguments
	Execute(ctx context.Context, function string, args []byte) ([]byte, error)

	// Close cleans up the plugin instance
	Close() error

	// GetExports returns the list of exported functions
	GetExports() []string
}

// Runtime represents the plugin runtime environment
type Runtime interface {
	// LoadPlugin loads a plugin from the given WASM file
	LoadPlugin(ctx context.Context, wasmPath string, manifest *Manifest) (PluginInstance, error)

	// UnloadPlugin unloads a plugin instance
	UnloadPlugin(instance PluginInstance) error

	// GetCapabilities returns the runtime capabilities
	GetCapabilities() []string

	// Close shuts down the runtime
	Close() error
}

// NewRegistry creates a new plugin registry
func NewRegistry(logger logging.Logger, fsManager fs.Manager, pluginDirs []string, runtime Runtime) *Registry {
	return &Registry{
		logger:        logger,
		fsManager:     fsManager,
		pluginDirs:    pluginDirs,
		plugins:       make(map[string]*PluginInfo),
		runtime:       runtime,
		loadedPlugins: make(map[string]LoadedPlugin),
	}
}

// DiscoverPlugins scans configured directories for plugins
func (r *Registry) DiscoverPlugins(ctx context.Context) error {
	r.logger.Debug("starting plugin discovery", "directories", r.pluginDirs)

	discovered := make(map[string]*PluginInfo)

	for _, dir := range r.pluginDirs {
		// Expand home directory if needed
		expandedDir := r.expandPath(dir)

		// Check if directory exists
		if _, err := os.Stat(expandedDir); os.IsNotExist(err) {
			r.logger.Debug("plugin directory does not exist", "directory", expandedDir)
			continue
		}

		// Scan directory for plugins
		plugins, err := r.scanDirectory(ctx, expandedDir)
		if err != nil {
			r.logger.Warn("failed to scan plugin directory", "directory", expandedDir, "error", err)
			continue
		}

		// Merge discovered plugins
		for name, plugin := range plugins {
			if existing, exists := discovered[name]; exists {
				r.logger.Warn("duplicate plugin found",
					"name", name,
					"existing_path", existing.Path,
					"new_path", plugin.Path)
				// Keep the first one found (directory priority order)
				continue
			}
			discovered[name] = plugin
		}
	}

	// Update registry with discovered plugins
	r.mu.Lock()
	r.plugins = discovered
	r.mu.Unlock()

	r.logger.Info("plugin discovery completed", "count", len(discovered))

	return nil
}

// scanDirectory scans a single directory for plugins
func (r *Registry) scanDirectory(ctx context.Context, dir string) (map[string]*PluginInfo, error) {
	plugins := make(map[string]*PluginInfo)

	// Walk the directory tree
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip non-directories and hidden directories
		if !info.IsDir() || strings.HasPrefix(info.Name(), ".") {
			return nil
		}

		// Look for manifest.yaml in each subdirectory
		manifestPath := filepath.Join(path, "manifest.yaml")
		if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
			return nil // No manifest, skip this directory
		}

		// Parse plugin manifest
		plugin, err := r.parsePlugin(path, manifestPath)
		if err != nil {
			r.logger.Warn("failed to parse plugin", "path", path, "error", err)
			// Create error plugin entry
			plugins[filepath.Base(path)] = &PluginInfo{
				Path:         path,
				Discovered:   time.Now(),
				LastModified: info.ModTime(),
				Status:       PluginStatusError,
				Error:        err.Error(),
			}
			return nil
		}

		plugins[plugin.Manifest.Plugin.Name] = plugin
		r.logger.Debug("discovered plugin", "name", plugin.Manifest.Plugin.Name, "path", path)

		return nil
	})

	return plugins, err
}

// parsePlugin parses a plugin from its manifest file
func (r *Registry) parsePlugin(pluginDir, manifestPath string) (*PluginInfo, error) {
	// Read manifest file
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse manifest YAML
	var manifest Manifest
	if err := yaml.Unmarshal(manifestData, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest YAML: %w", err)
	}

	// Validate manifest
	if err := r.validateManifest(&manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest: %w", err)
	}

	// Resolve WASM file path
	wasmPath := filepath.Join(pluginDir, manifest.Runtime.WASMFile)
	if _, err := os.Stat(wasmPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("WASM file not found: %s", wasmPath)
	}

	// Get file modification time
	wasmInfo, err := os.Stat(wasmPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat WASM file: %w", err)
	}

	plugin := &PluginInfo{
		Manifest:     &manifest,
		Path:         pluginDir,
		WASMPath:     wasmPath,
		Discovered:   time.Now(),
		LastModified: wasmInfo.ModTime(),
		Status:       PluginStatusDiscovered,
	}

	return plugin, nil
}

// validateManifest validates a plugin manifest
func (r *Registry) validateManifest(manifest *Manifest) error {
	if manifest.SchemaVersion == "" {
		return fmt.Errorf("missing schema_version")
	}

	if manifest.Plugin.Name == "" {
		return fmt.Errorf("missing plugin name")
	}

	if manifest.Plugin.Version == "" {
		return fmt.Errorf("missing plugin version")
	}

	if manifest.Runtime.WASMFile == "" {
		return fmt.Errorf("missing WASM file")
	}

	// Validate capabilities
	validCapabilities := map[string]bool{
		"task_sync":        true,
		"field_mapping":    true,
		"webhook_support":  true,
		"real_time_sync":   true,
		"batch_operations": true,
	}

	for _, capability := range manifest.Capabilities {
		if !validCapabilities[capability] {
			return fmt.Errorf("invalid capability: %s", capability)
		}
	}

	// Validate API requirements
	validAPIRequirements := map[string]bool{
		"http_client":       true,
		"credential_access": true,
		"logging":           true,
		"config_access":     true,
		"task_access":       true,
	}

	for _, requirement := range manifest.APIRequirements {
		if !validAPIRequirements[requirement] {
			return fmt.Errorf("invalid API requirement: %s", requirement)
		}
	}

	return nil
}

// LoadPlugin loads a plugin into the runtime
func (r *Registry) LoadPlugin(ctx context.Context, pluginName string) error {
	r.mu.RLock()
	pluginInfo, exists := r.plugins[pluginName]
	r.mu.RUnlock()

	if !exists {
		return fmt.Errorf("plugin not found: %s", pluginName)
	}

	if pluginInfo.Status != PluginStatusDiscovered {
		return fmt.Errorf("plugin cannot be loaded, status: %s", pluginInfo.Status)
	}

	r.logger.Debug("loading plugin", "name", pluginName, "path", pluginInfo.WASMPath)

	// Load plugin into runtime
	instance, err := r.runtime.LoadPlugin(ctx, pluginInfo.WASMPath, pluginInfo.Manifest)
	if err != nil {
		// Update plugin status to error
		r.mu.Lock()
		pluginInfo.Status = PluginStatusError
		pluginInfo.Error = err.Error()
		r.mu.Unlock()

		return fmt.Errorf("failed to load plugin: %w", err)
	}

	// Store loaded plugin
	r.runtimeMu.Lock()
	r.loadedPlugins[pluginName] = LoadedPlugin{
		Info:     pluginInfo,
		Instance: instance,
		LoadedAt: time.Now(),
	}
	r.runtimeMu.Unlock()

	// Update plugin status
	r.mu.Lock()
	pluginInfo.Status = PluginStatusLoaded
	pluginInfo.Error = ""
	r.mu.Unlock()

	r.logger.Info("plugin loaded successfully", "name", pluginName)

	return nil
}

// UnloadPlugin unloads a plugin from the runtime
func (r *Registry) UnloadPlugin(pluginName string) error {
	r.runtimeMu.Lock()
	loaded, exists := r.loadedPlugins[pluginName]
	if exists {
		delete(r.loadedPlugins, pluginName)
	}
	r.runtimeMu.Unlock()

	if !exists {
		return fmt.Errorf("plugin not loaded: %s", pluginName)
	}

	// Close plugin instance
	if err := r.runtime.UnloadPlugin(loaded.Instance); err != nil {
		r.logger.Warn("failed to cleanly unload plugin", "name", pluginName, "error", err)
	}

	// Update plugin status
	r.mu.Lock()
	if pluginInfo, exists := r.plugins[pluginName]; exists {
		pluginInfo.Status = PluginStatusDiscovered
		pluginInfo.Error = ""
	}
	r.mu.Unlock()

	r.logger.Info("plugin unloaded", "name", pluginName)

	return nil
}

// GetPlugin returns information about a plugin
func (r *Registry) GetPlugin(pluginName string) (*PluginInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin not found: %s", pluginName)
	}

	return plugin, nil
}

// ListPlugins returns all discovered plugins
func (r *Registry) ListPlugins() map[string]*PluginInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]*PluginInfo)
	for name, plugin := range r.plugins {
		result[name] = plugin
	}

	return result
}

// GetLoadedPlugin returns a loaded plugin instance
func (r *Registry) GetLoadedPlugin(pluginName string) (*LoadedPlugin, error) {
	r.runtimeMu.RLock()
	defer r.runtimeMu.RUnlock()

	loaded, exists := r.loadedPlugins[pluginName]
	if !exists {
		return nil, fmt.Errorf("plugin not loaded: %s", pluginName)
	}

	return &loaded, nil
}

// ListLoadedPlugins returns all loaded plugins
func (r *Registry) ListLoadedPlugins() map[string]LoadedPlugin {
	r.runtimeMu.RLock()
	defer r.runtimeMu.RUnlock()

	// Create a copy to avoid concurrent access issues
	result := make(map[string]LoadedPlugin)
	for name, loaded := range r.loadedPlugins {
		result[name] = loaded
	}

	return result
}

// ReloadPlugin reloads a plugin (unload + load)
func (r *Registry) ReloadPlugin(ctx context.Context, pluginName string) error {
	// Unload if loaded
	if _, err := r.GetLoadedPlugin(pluginName); err == nil {
		if err := r.UnloadPlugin(pluginName); err != nil {
			return fmt.Errorf("failed to unload plugin: %w", err)
		}
	}

	// Load plugin
	return r.LoadPlugin(ctx, pluginName)
}

// Close shuts down the plugin registry and runtime
func (r *Registry) Close() error {
	r.logger.Debug("shutting down plugin registry")

	// Unload all plugins
	r.runtimeMu.Lock()
	for name := range r.loadedPlugins {
		if err := r.UnloadPlugin(name); err != nil {
			r.logger.Warn("failed to unload plugin during shutdown", "name", name, "error", err)
		}
	}
	r.runtimeMu.Unlock()

	// Close runtime
	if r.runtime != nil {
		if err := r.runtime.Close(); err != nil {
			r.logger.Warn("failed to close plugin runtime", "error", err)
		}
	}

	r.logger.Info("plugin registry shutdown completed")

	return nil
}

// expandPath expands ~ in file paths
func (r *Registry) expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// RefreshPlugins re-discovers plugins from configured directories
func (r *Registry) RefreshPlugins(ctx context.Context) error {
	r.logger.Debug("refreshing plugin registry")

	// Store currently loaded plugins
	r.runtimeMu.RLock()
	currentlyLoaded := make(map[string]bool)
	for name := range r.loadedPlugins {
		currentlyLoaded[name] = true
	}
	r.runtimeMu.RUnlock()

	// Re-discover plugins
	if err := r.DiscoverPlugins(ctx); err != nil {
		return fmt.Errorf("failed to refresh plugins: %w", err)
	}

	// Reload previously loaded plugins if they still exist
	for name := range currentlyLoaded {
		if _, exists := r.plugins[name]; exists {
			if err := r.ReloadPlugin(ctx, name); err != nil {
				r.logger.Warn("failed to reload plugin during refresh", "name", name, "error", err)
			}
		} else {
			r.logger.Info("previously loaded plugin no longer exists", "name", name)
		}
	}

	r.logger.Info("plugin registry refresh completed")

	return nil
}
