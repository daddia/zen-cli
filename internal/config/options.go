package config

import (
	"fmt"
	"strings"
)

// ConfigOption represents a configuration option with validation
type ConfigOption struct {
	Key           string
	Description   string
	AllowedValues []string
	DefaultValue  string
	Type          string // "string", "bool", "int"
}

// Options contains configuration options organized by section
var Options = []ConfigOption{
	// Core settings
	{
		Key:           "core.config_dir",
		Description:   "Zen configuration directory",
		DefaultValue:  ".zen",
		Type:          "string",
	},
	{
		Key:           "core.token",
		Description:   "Zen authentication token",
		DefaultValue:  "",
		Type:          "string",
	},
	{
		Key:           "core.debug",
		Description:   "Enable debug mode",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "false",
		Type:          "bool",
	},
	{
		Key:           "core.log_level",
		Description:   "Set the logging level",
		AllowedValues: []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"},
		DefaultValue:  "info",
		Type:          "string",
	},
	{
		Key:           "core.log_format",
		Description:   "Set the logging format",
		AllowedValues: []string{"text", "json"},
		DefaultValue:  "text",
		Type:          "string",
	},
	// Project settings
	{
		Key:          "project.name",
		Description:  "Project name (defaults to directory name)",
		DefaultValue: "",
		Type:         "string",
	},
	{
		Key:          "project.path",
		Description:  "Project path",
		DefaultValue: "",
		Type:         "string",
	},
	// Task settings
	{
		Key:          "task.task_path",
		Description:  "Tasks directory path",
		DefaultValue: ".zen/tasks",
		Type:         "string",
	},
	{
		Key:           "task.task_source",
		Description:   "Task source system",
		AllowedValues: []string{"local", "jira", "github", "linear"},
		DefaultValue:  "local",
		Type:          "string",
	},
}

// GetCurrentValue returns the current value of a configuration option
func (opt ConfigOption) GetCurrentValue(cfg *Config) string {
	value := opt.getValueFromConfig(cfg)
	if value == "" {
		return opt.DefaultValue
	}
	return value
}

// getValueFromConfig extracts the current value from the configuration
func (opt ConfigOption) getValueFromConfig(cfg *Config) string {
	switch opt.Key {
	// Core settings
	case "core.config_dir":
		return cfg.Core.ConfigDir
	case "core.token":
		return cfg.Core.Token
	case "core.debug":
		return fmt.Sprintf("%t", cfg.Core.Debug)
	case "core.log_level":
		return cfg.Core.LogLevel
	case "core.log_format":
		return cfg.Core.LogFormat
	// Project settings
	case "project.name":
		return cfg.Project.Name
	case "project.path":
		return cfg.Project.Path
	// Task settings
	case "task.task_path":
		return cfg.Task.TaskPath
	case "task.task_source":
		return cfg.Task.TaskSource
	default:
		return ""
	}
}

// FindOption finds a configuration option by key
func FindOption(key string) (ConfigOption, bool) {
	for _, opt := range Options {
		if opt.Key == key {
			return opt, true
		}
	}
	return ConfigOption{}, false
}

// ValidateKey validates that a configuration key exists
func ValidateKey(key string) error {
	if _, found := FindOption(key); !found {
		availableKeys := make([]string, len(Options))
		for i, opt := range Options {
			availableKeys[i] = opt.Key
		}
		return fmt.Errorf("unknown configuration key %q. Available keys: %s",
			key, strings.Join(availableKeys, ", "))
	}
	return nil
}

// ValidateValue validates that a value is valid for the given key
func ValidateValue(key, value string) error {
	opt, found := FindOption(key)
	if !found {
		return fmt.Errorf("unknown configuration key: %s", key)
	}

	// If no allowed values specified, any value is valid
	if len(opt.AllowedValues) == 0 {
		return nil
	}

	// Check if value is in allowed values
	for _, allowed := range opt.AllowedValues {
		if value == allowed {
			return nil
		}
	}

	return &InvalidValueError{
		Key:         key,
		Value:       value,
		ValidValues: opt.AllowedValues,
	}
}

// InvalidValueError represents an invalid configuration value error
type InvalidValueError struct {
	Key         string
	Value       string
	ValidValues []string
}

func (e *InvalidValueError) Error() string {
	return fmt.Sprintf("invalid value %q for key %q. Valid values: %s",
		e.Value, e.Key, strings.Join(e.ValidValues, ", "))
}
