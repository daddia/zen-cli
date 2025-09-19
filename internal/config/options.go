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
	{
		Key:           "templates.cache_enabled",
		Description:   "Enable template compilation caching",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "true",
		Type:          "bool",
	},
	{
		Key:          "templates.cache_ttl",
		Description:  "Template cache TTL duration",
		DefaultValue: "30m",
		Type:         "string",
	},
	{
		Key:          "templates.cache_size",
		Description:  "Maximum number of templates to cache",
		DefaultValue: "100",
		Type:         "int",
	},
	{
		Key:           "templates.strict_mode",
		Description:   "Enable strict mode (error on missing variables)",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "false",
		Type:          "bool",
	},
	{
		Key:           "templates.enable_ai",
		Description:   "Enable AI enhancement features",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "false",
		Type:          "bool",
	},
	{
		Key:          "templates.left_delim",
		Description:  "Left template delimiter",
		DefaultValue: "{{",
		Type:         "string",
	},
	{
		Key:          "templates.right_delim",
		Description:  "Right template delimiter",
		DefaultValue: "}}",
		Type:         "string",
	},
	// Assets configuration options
	{
		Key:          "assets.repository_url",
		Description:  "Asset repository URL",
		DefaultValue: "https://github.com/daddia/zen-assets.git",
		Type:         "string",
	},
	{
		Key:          "assets.branch",
		Description:  "Asset repository branch",
		DefaultValue: "main",
		Type:         "string",
	},
	{
		Key:           "assets.auth_provider",
		Description:   "Authentication provider for assets",
		AllowedValues: []string{"github", "gitlab"},
		DefaultValue:  "github",
		Type:          "string",
	},
	{
		Key:          "assets.cache_path",
		Description:  "Local cache path for assets",
		DefaultValue: "~/.zen/assets",
		Type:         "string",
	},
	{
		Key:          "assets.cache_size_mb",
		Description:  "Maximum cache size in MB",
		DefaultValue: "100",
		Type:         "int",
	},
	{
		Key:          "assets.sync_timeout_seconds",
		Description:  "Sync timeout in seconds",
		DefaultValue: "30",
		Type:         "int",
	},
	{
		Key:           "assets.integrity_checks_enabled",
		Description:   "Enable integrity checks for assets",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "true",
		Type:          "bool",
	},
	{
		Key:           "assets.prefetch_enabled",
		Description:   "Enable prefetching of assets",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "true",
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

	current := reflect.ValueOf(cfg).Elem()

	for i, part := range parts {
		// Handle nested structs
		switch part {
		case "cli":
			current = current.FieldByName("CLI")
		case "workspace":
			current = current.FieldByName("Workspace")
		case "assets":
			current = current.FieldByName("Assets")
		case "templates":
			current = current.FieldByName("Templates")
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
				case "repository_url":
					current = current.FieldByName("RepositoryURL")
				case "auth_provider":
					current = current.FieldByName("AuthProvider")
				case "cache_path":
					current = current.FieldByName("CachePath")
				case "cache_size_mb":
					current = current.FieldByName("CacheSizeMB")
				case "sync_timeout_seconds":
					current = current.FieldByName("SyncTimeoutSeconds")
				case "integrity_checks_enabled":
					current = current.FieldByName("IntegrityChecksEnabled")
				case "prefetch_enabled":
					current = current.FieldByName("PrefetchEnabled")
				case "cache_enabled":
					current = current.FieldByName("CacheEnabled")
				case "cache_ttl":
					current = current.FieldByName("CacheTTL")
				case "cache_size":
					current = current.FieldByName("CacheSize")
				case "strict_mode":
					current = current.FieldByName("StrictMode")
				case "enable_ai":
					current = current.FieldByName("EnableAI")
				case "left_delim":
					current = current.FieldByName("LeftDelim")
				case "right_delim":
					current = current.FieldByName("RightDelim")
				default:
					current = current.FieldByName(fieldName)
				}
			}
		}

		if !current.IsValid() {
			return ""
		}
	}

	// Handle pointer types (used for optional bool fields)
	if current.Kind() == reflect.Ptr {
		if current.IsNil() {
			return ""
		}
		current = current.Elem()
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
