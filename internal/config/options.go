package config

import (
	"fmt"
	"reflect"
	"strconv"
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

// Options contains all available configuration options
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
	{
		Key:           "cli.no_color",
		Description:   "Disable colored output",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "false",
		Type:          "bool",
	},
	{
		Key:           "cli.verbose",
		Description:   "Enable verbose output",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "false",
		Type:          "bool",
	},
	{
		Key:           "cli.output_format",
		Description:   "Set the default output format",
		AllowedValues: []string{"text", "json", "yaml"},
		DefaultValue:  "text",
		Type:          "string",
	},
	{
		Key:          "workspace.root",
		Description:  "Set the workspace root directory",
		DefaultValue: ".",
		Type:         "string",
	},
	{
		Key:          "workspace.config_file",
		Description:  "Set the workspace configuration file name",
		DefaultValue: "zen.yaml",
		Type:         "string",
	},
	{
		Key:           "development.debug",
		Description:   "Enable development debug mode",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "false",
		Type:          "bool",
	},
	{
		Key:           "development.profile",
		Description:   "Enable development profiling",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "false",
		Type:          "bool",
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

// getValueFromConfig extracts the value from the config struct using reflection
func (opt ConfigOption) getValueFromConfig(cfg *Config) string {
	parts := strings.Split(opt.Key, ".")

	var current reflect.Value = reflect.ValueOf(cfg).Elem()

	for i, part := range parts {
		// Handle nested structs
		switch part {
		case "cli":
			current = current.FieldByName("CLI")
		case "workspace":
			current = current.FieldByName("Workspace")
		case "development":
			current = current.FieldByName("Development")
		default:
			// Convert snake_case to PascalCase for struct field names
			fieldName := toPascalCase(part)
			if i == 0 {
				// Top-level fields
				switch part {
				case "log_level":
					current = current.FieldByName("LogLevel")
				case "log_format":
					current = current.FieldByName("LogFormat")
				default:
					current = current.FieldByName(fieldName)
				}
			} else {
				// Nested fields
				switch part {
				case "no_color":
					current = current.FieldByName("NoColor")
				case "output_format":
					current = current.FieldByName("OutputFormat")
				case "config_file":
					current = current.FieldByName("ConfigFile")
				default:
					current = current.FieldByName(fieldName)
				}
			}
		}

		if !current.IsValid() {
			return ""
		}
	}

	// Convert the value to string
	switch current.Kind() {
	case reflect.String:
		return current.String()
	case reflect.Bool:
		return strconv.FormatBool(current.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(current.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(current.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(current.Float(), 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", current.Interface())
	}
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// FindOption finds a configuration option by key
func FindOption(key string) (*ConfigOption, bool) {
	for _, opt := range Options {
		if opt.Key == key {
			return &opt, true
		}
	}
	return nil, false
}

// ValidateKey checks if a configuration key is valid
func ValidateKey(key string) error {
	_, found := FindOption(key)
	if !found {
		return fmt.Errorf("unknown configuration key: %s", key)
	}
	return nil
}

// ValidateValue validates a value for a given configuration key
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
	return fmt.Sprintf("invalid value %q for key %q (valid values: %s)",
		e.Value, e.Key, strings.Join(e.ValidValues, ", "))
}
