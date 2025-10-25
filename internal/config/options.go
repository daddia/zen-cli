package config

import (
	"fmt"
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
		DefaultValue: "config",
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
		DefaultValue: "~/.zen/library",
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
	// Work configuration options
	{
		Key:           "work.tasks.source",
		Description:   "Task source system",
		AllowedValues: []string{"jira", "github", "linear", "monday", "asana", "local", "none", ""},
		DefaultValue:  "local",
		Type:          "string",
	},
	{
		Key:           "work.tasks.sync",
		Description:   "Task synchronization frequency",
		AllowedValues: []string{"hourly", "daily", "manual", "none", ""},
		DefaultValue:  "manual",
		Type:          "string",
	},
	{
		Key:           "work.tasks.project_key",
		Description:   "Project key or identifier for tasks",
		AllowedValues: []string{},
		DefaultValue:  "",
		Type:          "string",
	},

	// Provider configuration options
	{
		Key:           "providers.jira.type",
		Description:   "Jira provider type",
		AllowedValues: []string{"jira"},
		DefaultValue:  "jira",
		Type:          "string",
	},
	{
		Key:           "providers.jira.url",
		Description:   "Jira server URL",
		AllowedValues: []string{},
		DefaultValue:  "",
		Type:          "string",
	},
	{
		Key:           "providers.jira.email",
		Description:   "Jira user email for authentication",
		AllowedValues: []string{},
		DefaultValue:  "",
		Type:          "string",
	},
	{
		Key:           "providers.jira.api_token",
		Description:   "Jira API token for authentication",
		AllowedValues: []string{},
		DefaultValue:  "",
		Type:          "string",
	},
	{
		Key:           "providers.github.type",
		Description:   "GitHub provider type",
		AllowedValues: []string{"github"},
		DefaultValue:  "github",
		Type:          "string",
	},
	{
		Key:           "providers.github.url",
		Description:   "GitHub API URL",
		AllowedValues: []string{},
		DefaultValue:  "https://api.github.com",
		Type:          "string",
	},
	{
		Key:           "providers.linear.type",
		Description:   "Linear provider type",
		AllowedValues: []string{"linear"},
		DefaultValue:  "linear",
		Type:          "string",
	},
	{
		Key:           "providers.linear.url",
		Description:   "Linear API URL",
		AllowedValues: []string{},
		DefaultValue:  "https://api.linear.app",
		Type:          "string",
	},

	// Integration configuration options (legacy - kept for backward compatibility)
	{
		Key:           "integrations.task_system",
		Description:   "Task system of record for external integration (deprecated: use work.tasks.source)",
		AllowedValues: []string{"jira", "github", "monday", "asana", "none", ""},
		DefaultValue:  "",
		Type:          "string",
	},
	{
		Key:           "integrations.sync_enabled",
		Description:   "Enable task synchronization with external systems",
		AllowedValues: []string{"true", "false"},
		DefaultValue:  "false",
		Type:          "bool",
	},
	{
		Key:           "integrations.sync_frequency",
		Description:   "Frequency of automatic synchronization",
		AllowedValues: []string{"hourly", "daily", "manual", ""},
		DefaultValue:  "manual",
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

// getValueFromConfig extracts the value from the config struct using simple field access
func (opt ConfigOption) getValueFromConfig(cfg *Config) string {
	parts := strings.Split(opt.Key, ".")

	// Handle top-level fields
	if len(parts) == 1 {
		switch parts[0] {
		case "log_level":
			return cfg.LogLevel
		case "log_format":
			return cfg.LogFormat
		default:
			return ""
		}
	}

	// Handle nested fields
	if len(parts) == 2 {
		switch parts[0] {
		case "cli":
			return opt.getCLIValue(cfg, parts[1])
		case "workspace":
			return opt.getWorkspaceValue(cfg, parts[1])
		case "assets":
			return opt.getAssetsValue(cfg, parts[1])
		case "templates":
			return opt.getTemplatesValue(cfg, parts[1])
		case "development":
			return opt.getDevelopmentValue(cfg, parts[1])
		default:
			return ""
		}
	}

	// Handle deeper nesting (work.tasks.*, providers.*.*, etc.)
	if len(parts) == 3 {
		switch parts[0] {
		case "work":
			if parts[1] == "tasks" {
				return opt.getTasksValue(cfg, parts[2])
			}
		case "providers":
			return opt.getProviderValue(cfg, parts[1], parts[2])
		}
	}

	return ""
}

func (opt ConfigOption) getCLIValue(cfg *Config, field string) string {
	switch field {
	case "no_color":
		return strconv.FormatBool(cfg.CLI.NoColor)
	case "verbose":
		return strconv.FormatBool(cfg.CLI.Verbose)
	case "output_format":
		return cfg.CLI.OutputFormat
	default:
		return ""
	}
}

func (opt ConfigOption) getWorkspaceValue(cfg *Config, field string) string {
	switch field {
	case "root":
		return cfg.Workspace.Root
	case "zen_path":
		return cfg.Workspace.ZenPath
	case "config_file":
		return cfg.Workspace.ConfigFile
	default:
		return ""
	}
}

func (opt ConfigOption) getAssetsValue(cfg *Config, field string) string {
	switch field {
	case "repository_url":
		return cfg.Assets.RepositoryURL
	case "branch":
		return cfg.Assets.Branch
	case "auth_provider":
		return cfg.Assets.AuthProvider
	case "cache_path":
		return cfg.Assets.CachePath
	case "cache_size_mb":
		return strconv.FormatInt(cfg.Assets.CacheSizeMB, 10)
	case "sync_timeout_seconds":
		return strconv.Itoa(cfg.Assets.SyncTimeoutSeconds)
	case "integrity_checks_enabled":
		return strconv.FormatBool(cfg.Assets.IntegrityChecksEnabled)
	case "prefetch_enabled":
		return strconv.FormatBool(cfg.Assets.PrefetchEnabled)
	default:
		return ""
	}
}

func (opt ConfigOption) getTemplatesValue(cfg *Config, field string) string {
	switch field {
	case "cache_enabled":
		if cfg.Templates.CacheEnabled != nil {
			return strconv.FormatBool(*cfg.Templates.CacheEnabled)
		}
		return ""
	case "cache_ttl":
		return cfg.Templates.CacheTTL
	case "cache_size":
		return strconv.Itoa(cfg.Templates.CacheSize)
	case "strict_mode":
		if cfg.Templates.StrictMode != nil {
			return strconv.FormatBool(*cfg.Templates.StrictMode)
		}
		return ""
	case "enable_ai":
		if cfg.Templates.EnableAI != nil {
			return strconv.FormatBool(*cfg.Templates.EnableAI)
		}
		return ""
	case "left_delim":
		return cfg.Templates.LeftDelim
	case "right_delim":
		return cfg.Templates.RightDelim
	default:
		return ""
	}
}

func (opt ConfigOption) getDevelopmentValue(cfg *Config, field string) string {
	switch field {
	case "debug":
		return strconv.FormatBool(cfg.Development.Debug)
	case "profile":
		return strconv.FormatBool(cfg.Development.Profile)
	default:
		return ""
	}
}

func (opt ConfigOption) getTasksValue(cfg *Config, field string) string {
	switch field {
	case "source":
		return cfg.Work.Tasks.Source
	case "sync":
		return cfg.Work.Tasks.Sync
	case "project_key":
		return cfg.Work.Tasks.ProjectKey
	default:
		return ""
	}
}

func (opt ConfigOption) getProviderValue(cfg *Config, provider, field string) string {
	if cfg.Providers == nil {
		return ""
	}

	providerConfig, exists := cfg.Providers[provider]
	if !exists {
		return ""
	}

	switch field {
	case "type":
		return providerConfig.Type
	case "url":
		return providerConfig.URL
	case "email":
		return providerConfig.Email
	case "api_token":
		return providerConfig.APIToken
	default:
		return ""
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
