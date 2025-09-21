package plugin

import (
	"context"
	"fmt"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/auth"
)

// SimpleRuntime implements a basic runtime for testing without WASM complexity
type SimpleRuntime struct {
	logger    logging.Logger
	auth      auth.Manager
	instances map[string]*SimpleInstance
}

// SimpleInstance represents a simple plugin instance for testing
type SimpleInstance struct {
	name      string
	manifest  *Manifest
	loadedAt  time.Time
	functions map[string]func([]byte) ([]byte, error)
}

// NewSimpleRuntime creates a new simple runtime for testing
func NewSimpleRuntime(logger logging.Logger, auth auth.Manager) *SimpleRuntime {
	return &SimpleRuntime{
		logger:    logger,
		auth:      auth,
		instances: make(map[string]*SimpleInstance),
	}
}

// LoadPlugin loads a plugin (simplified implementation)
func (r *SimpleRuntime) LoadPlugin(ctx context.Context, wasmPath string, manifest *Manifest) (PluginInstance, error) {
	r.logger.Debug("loading simple plugin", "path", wasmPath, "plugin", manifest.Plugin.Name)

	instance := &SimpleInstance{
		name:     manifest.Plugin.Name,
		manifest: manifest,
		loadedAt: time.Now(),
		functions: map[string]func([]byte) ([]byte, error){
			"plugin_init":         func([]byte) ([]byte, error) { return []byte("ok"), nil },
			"plugin_cleanup":      func([]byte) ([]byte, error) { return []byte("ok"), nil },
			"get_task_data":       func([]byte) ([]byte, error) { return []byte(`{"id":"test"}`), nil },
			"create_task":         func([]byte) ([]byte, error) { return []byte(`{"id":"created"}`), nil },
			"update_task":         func([]byte) ([]byte, error) { return []byte(`{"id":"updated"}`), nil },
			"search_tasks":        func([]byte) ([]byte, error) { return []byte(`[]`), nil },
			"map_to_zen":          func([]byte) ([]byte, error) { return []byte(`{"id":"zen"}`), nil },
			"map_to_external":     func([]byte) ([]byte, error) { return []byte(`{"id":"external"}`), nil },
			"validate_connection": func([]byte) ([]byte, error) { return []byte("ok"), nil },
			"health_check":        func([]byte) ([]byte, error) { return []byte(`{"healthy":true}`), nil },
		},
	}

	r.instances[manifest.Plugin.Name] = instance

	r.logger.Info("simple plugin loaded successfully", "plugin", manifest.Plugin.Name)

	return instance, nil
}

// UnloadPlugin unloads a plugin instance
func (r *SimpleRuntime) UnloadPlugin(instance PluginInstance) error {
	simpleInstance, ok := instance.(*SimpleInstance)
	if !ok {
		return fmt.Errorf("invalid plugin instance type")
	}

	delete(r.instances, simpleInstance.name)
	r.logger.Debug("simple plugin unloaded", "plugin", simpleInstance.name)

	return nil
}

// GetCapabilities returns the runtime capabilities
func (r *SimpleRuntime) GetCapabilities() []string {
	return []string{
		"simple_execution",
		"basic_isolation",
		"function_calls",
	}
}

// Close shuts down the runtime
func (r *SimpleRuntime) Close() error {
	r.logger.Debug("shutting down simple runtime")

	for name := range r.instances {
		delete(r.instances, name)
	}

	r.logger.Info("simple runtime shutdown completed")
	return nil
}

// Execute calls a function in the simple plugin instance
func (instance *SimpleInstance) Execute(ctx context.Context, function string, args []byte) ([]byte, error) {
	if fn, exists := instance.functions[function]; exists {
		return fn(args)
	}

	return nil, &PluginError{
		Code:      ErrCodePluginExecutionFailed,
		Message:   fmt.Sprintf("function not found: %s", function),
		Plugin:    instance.name,
		Function:  function,
		Timestamp: time.Now(),
	}
}

// Close cleans up the simple plugin instance
func (instance *SimpleInstance) Close() error {
	// Nothing to clean up for simple instance
	return nil
}

// GetExports returns the list of exported functions
func (instance *SimpleInstance) GetExports() []string {
	exports := make([]string, 0, len(instance.functions))
	for name := range instance.functions {
		exports = append(exports, name)
	}
	return exports
}
