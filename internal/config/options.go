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

// Options contains core configuration options only
// Component-specific options are now handled by each component
var Options = []ConfigOption{
	{
		Key:           "log_level",
		Description:   "Set the logging level",
		AllowedValues: []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"},
		DefaultValue:  "info",
		Type:          "string",
	},
	{
		Key:           "log_format",
		Description:   "Set the logging format",
		AllowedValues: []string{"text", "json"},
		DefaultValue:  "text",
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
	case "log_level":
		return cfg.LogLevel
	case "log_format":
		return cfg.LogFormat
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
