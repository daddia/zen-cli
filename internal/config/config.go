package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// Logging configuration
	LogLevel  string `mapstructure:"log_level" validate:"required,oneof=trace debug info warn error fatal panic"`
	LogFormat string `mapstructure:"log_format" validate:"required,oneof=text json"`

	// CLI configuration
	CLI CLIConfig `mapstructure:"cli"`

	// Workspace configuration
	Workspace WorkspaceConfig `mapstructure:"workspace"`

	// Assets configuration
	Assets AssetsConfig `mapstructure:"assets"`

	// Templates configuration
	Templates TemplatesConfig `mapstructure:"templates"`

	// Development configuration
	Development DevelopmentConfig `mapstructure:"development"`

	// Work configuration
	Work WorkConfig `mapstructure:"work"`

	// Integration configuration
	Integrations IntegrationsConfig `mapstructure:"integrations"`

	// Provider configurations
	Providers map[string]ProviderConfig `mapstructure:"providers"`

	// Internal fields for configuration management
	viper      *viper.Viper `mapstructure:"-" json:"-" yaml:"-"`
	configFile string       `mapstructure:"-" json:"-" yaml:"-"`
	loadedFrom []string     `mapstructure:"-" json:"-" yaml:"-"`
}

// CLIConfig contains CLI-specific configuration
type CLIConfig struct {
	// Color output settings
	NoColor bool `mapstructure:"no_color"`

	// Verbose output
	Verbose bool `mapstructure:"verbose"`

	// Output format (text, json, yaml)
	OutputFormat string `mapstructure:"output_format" validate:"required,oneof=text json yaml"`
}

// WorkspaceConfig contains workspace-specific configuration
type WorkspaceConfig struct {
	// Root directory for workspace detection
	Root string `mapstructure:"root"`

	// Zen directory path relative to workspace root
	ZenPath string `mapstructure:"zen_path"`

	// Configuration file name
	ConfigFile string `mapstructure:"config_file"`
}

// AssetsConfig contains asset repository configuration
type AssetsConfig struct {
	// Repository URL (defaults to official zen-assets repository)
	RepositoryURL string `mapstructure:"repository_url"`

	// Branch to use (defaults to "main")
	Branch string `mapstructure:"branch"`

	// Authentication provider (github, gitlab)
	AuthProvider string `mapstructure:"auth_provider"`

	// Cache settings
	CachePath   string `mapstructure:"cache_path"`
	CacheSizeMB int64  `mapstructure:"cache_size_mb"`

	// Sync timeout in seconds
	SyncTimeoutSeconds int `mapstructure:"sync_timeout_seconds"`

	// Feature flags
	IntegrityChecksEnabled bool `mapstructure:"integrity_checks_enabled"`
	PrefetchEnabled        bool `mapstructure:"prefetch_enabled"`
}

// Redacted returns a copy of the AssetsConfig with sensitive fields redacted
func (a AssetsConfig) Redacted() AssetsConfig {
	redacted := a
	redacted.RepositoryURL = RedactSensitiveValue("repository_url", a.RepositoryURL)
	return redacted
}

// TemplatesConfig contains template engine configuration
type TemplatesConfig struct {
	// Enable template compilation caching
	CacheEnabled *bool `mapstructure:"cache_enabled"`

	// Cache TTL duration (e.g., "30m", "1h")
	CacheTTL string `mapstructure:"cache_ttl"`

	// Maximum number of templates to cache
	CacheSize int `mapstructure:"cache_size"`

	// Enable strict mode (error on missing variables)
	StrictMode *bool `mapstructure:"strict_mode"`

	// Enable AI enhancement features
	EnableAI *bool `mapstructure:"enable_ai"`

	// Custom template delimiters
	LeftDelim  string `mapstructure:"left_delim"`
	RightDelim string `mapstructure:"right_delim"`
}

// DevelopmentConfig contains development-specific settings
type DevelopmentConfig struct {
	// Enable debug mode
	Debug bool `mapstructure:"debug"`

	// Enable profiling
	Profile bool `mapstructure:"profile"`
}

// WorkConfig contains work-related configuration
type WorkConfig struct {
	// Tasks configuration
	Tasks TasksConfig `mapstructure:"tasks"`
}

// TasksConfig contains task-specific configuration
type TasksConfig struct {
	// Task source system (jira, github, linear, monday, asana, local, none)
	Source string `mapstructure:"source" validate:"oneof=jira github linear monday asana local none ''"`

	// Sync frequency (hourly, daily, manual, none)
	Sync string `mapstructure:"sync" validate:"oneof=hourly daily manual none ''"`

	// Project key or identifier for tasks
	ProjectKey string `mapstructure:"project_key"`
}

// ProviderConfig contains provider-specific configuration
type ProviderConfig struct {
	// Provider type (jira, github, linear, etc.)
	Type string `mapstructure:"type"`

	// Server URL for the provider
	URL string `mapstructure:"url"`

	// Email for authentication (if required)
	Email string `mapstructure:"email"`

	// API key/token for authentication
	APIToken string `mapstructure:"api_token"`

	// Additional provider-specific settings
	Settings map[string]interface{} `mapstructure:"settings"`
}

// IntegrationsConfig contains external integration configuration
type IntegrationsConfig struct {
	// Task system of record (jira, github, monday, asana, none)
	TaskSystem string `mapstructure:"task_system" validate:"oneof=jira github monday asana none ''"`

	// Enable synchronization
	SyncEnabled bool `mapstructure:"sync_enabled"`

	// Sync frequency (hourly, daily, manual)
	SyncFrequency string `mapstructure:"sync_frequency" validate:"oneof=hourly daily manual ''"`

	// Plugin directories for discovery
	PluginDirectories []string `mapstructure:"plugin_directories"`

	// Provider-specific configurations
	Providers map[string]IntegrationProviderConfig `mapstructure:"providers"`
}

// IntegrationProviderConfig contains provider-specific integration settings
type IntegrationProviderConfig struct {
	// Server URL for the integration
	URL string `mapstructure:"url"`

	// Project key or identifier
	ProjectKey string `mapstructure:"project_key"`

	// Authentication type (basic, oauth2, token)
	Type string `mapstructure:"type" validate:"oneof=basic oauth2 token ''"`

	// Credentials reference for auth system
	Credentials string `mapstructure:"credentials"`

	// Email for authentication (if required)
	Email string `mapstructure:"email"`

	// API key/token for authentication
	APIKey string `mapstructure:"api_key"`

	// Field mappings for data synchronization
	FieldMapping map[string]string `mapstructure:"field_mapping"`

	// Sync direction (pull, push, bidirectional)
	SyncDirection string `mapstructure:"sync_direction" validate:"oneof=pull push bidirectional ''"`

	// Additional provider-specific settings
	Settings map[string]interface{} `mapstructure:"settings"`
}

// DefaultWorkConfig returns default work configuration
func DefaultWorkConfig() WorkConfig {
	return WorkConfig{
		Tasks: TasksConfig{
			Source:     "local",
			Sync:       "manual",
			ProjectKey: "",
		},
	}
}

// DefaultProvidersConfig returns default providers configuration
func DefaultProvidersConfig() map[string]ProviderConfig {
	return map[string]ProviderConfig{
		"jira": {
			Type: "jira",
			URL:  "",
		},
		"github": {
			Type: "github",
			URL:  "https://api.github.com",
		},
		"linear": {
			Type: "linear",
			URL:  "https://api.linear.app",
		},
	}
}

// DefaultIntegrationsConfig returns default integration configuration
func DefaultIntegrationsConfig() IntegrationsConfig {
	return IntegrationsConfig{
		TaskSystem:    "",
		SyncEnabled:   false,
		SyncFrequency: "manual",
		PluginDirectories: []string{
			"~/.zen/plugins",
			".zen/plugins",
		},
		Providers: map[string]IntegrationProviderConfig{
			"jira": {
				Type:          "token",
				Credentials:   "jira",
				SyncDirection: "bidirectional",
				FieldMapping: map[string]string{
					"task_id":     "key",
					"title":       "summary",
					"status":      "status.name",
					"priority":    "priority.name",
					"assignee":    "assignee.displayName",
					"created":     "created",
					"updated":     "updated",
					"description": "description",
				},
				Settings: make(map[string]interface{}),
			},
		},
	}
}

// Load loads configuration from various sources with precedence handling
func Load() (*Config, error) {
	return LoadWithOptions(LoadOptions{})
}

// LoadWithCommand loads configuration and binds CLI flags from the provided command
func LoadWithCommand(cmd *cobra.Command) (*Config, error) {
	return LoadWithOptions(LoadOptions{
		Command: cmd,
	})
}

// LoadOptions provides configuration loading options
type LoadOptions struct {
	Command    *cobra.Command
	ConfigFile string
}

// LoadWithOptions loads configuration with the provided options using Viper as core engine
func LoadWithOptions(opts LoadOptions) (*Config, error) {
	start := time.Now()
	v := viper.New()

	// Set configuration defaults from Options
	setDefaults(v)

	// Configure fixed file paths
	configureSimpleFileDiscovery(v, opts.ConfigFile)

	// Bind CLI flags if command is provided
	if opts.Command != nil {
		if err := bindFlags(v, opts.Command); err != nil {
			return nil, errors.Wrap(err, "failed to bind CLI flags")
		}
	}

	// Track configuration sources
	var loadedFrom []string
	configFile := ""

	// Read configuration file (optional) - Viper handles the precedence
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, errors.Wrap(err, "failed to read config file")
		}
		// Config file not found is OK, we'll use defaults
		loadedFrom = append(loadedFrom, "defaults")
	} else {
		configFile = v.ConfigFileUsed()
		loadedFrom = append(loadedFrom, fmt.Sprintf("file:%s", configFile))
	}

	// Check for CLI flags
	if opts.Command != nil && hasFlagOverrides(opts.Command) {
		loadedFrom = append(loadedFrom, "flags")
	}

	// Unmarshal configuration using Viper
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal config")
	}

	// Store internal fields
	config.viper = v
	config.configFile = configFile
	config.loadedFrom = loadedFrom

	// Apply environment variable overrides (for non-Viper handled cases)
	applyEnvOverrides(&config)

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, errors.Wrap(err, "configuration validation failed")
	}

	// Log configuration loading performance
	loadDuration := time.Since(start)
	// Note: Performance logging is handled by the factory layer
	// to avoid circular dependencies with the logger
	_ = loadDuration // Suppress unused variable warning

	return &config, nil
}

// setDefaults sets default configuration values from Options using Viper
func setDefaults(v *viper.Viper) {
	// Use Options as single source of truth for defaults
	for _, opt := range Options {
		// Convert string defaults to appropriate types
		switch opt.Type {
		case "bool":
			if opt.DefaultValue == "true" {
				v.SetDefault(opt.Key, true)
			} else {
				v.SetDefault(opt.Key, false)
			}
		case "int":
			if val, err := strconv.Atoi(opt.DefaultValue); err == nil {
				v.SetDefault(opt.Key, val)
			}
		default: // "string"
			v.SetDefault(opt.Key, opt.DefaultValue)
		}
	}
}

// configureSimpleFileDiscovery sets up the fixed configuration file paths
func configureSimpleFileDiscovery(v *viper.Viper, configFile string) {
	// If specific config file is provided, use it
	if configFile != "" {
		v.SetConfigFile(configFile)
		return
	}

	// Set configuration name and type
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add fixed configuration paths:
	// Local/Project configuration (.zen/config)
	v.AddConfigPath("./.zen")

	// Global/User configuration (~/.zen/config) - PRIMARY LOCATION ONLY
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".zen"))
	}

	// System configuration (/etc/zen/config or C:\ProgramData\zen\config)
	if runtime.GOOS == "windows" {
		if programData := os.Getenv("ProgramData"); programData != "" {
			v.AddConfigPath(filepath.Join(programData, "zen"))
		}
	} else {
		v.AddConfigPath("/etc/zen")
	}
}

// bindFlags binds CLI flags to Viper configuration
func bindFlags(v *viper.Viper, cmd *cobra.Command) error {
	// Get the root command to access persistent flags
	rootCmd := cmd.Root()
	if rootCmd == nil {
		rootCmd = cmd
	}

	// Bind persistent flags if they exist
	if rootCmd.PersistentFlags() != nil {
		if flag := rootCmd.PersistentFlags().Lookup("verbose"); flag != nil {
			if err := v.BindPFlag("cli.verbose", flag); err != nil {
				return err
			}
		}
		if flag := rootCmd.PersistentFlags().Lookup("no-color"); flag != nil {
			if err := v.BindPFlag("cli.no_color", flag); err != nil {
				return err
			}
		}
		if flag := rootCmd.PersistentFlags().Lookup("output"); flag != nil {
			if err := v.BindPFlag("cli.output_format", flag); err != nil {
				return err
			}
		}
	}

	return nil
}

// hasFlagOverrides checks if any CLI flags are set
func hasFlagOverrides(cmd *cobra.Command) bool {
	rootCmd := cmd.Root()
	if rootCmd == nil {
		rootCmd = cmd
	}

	if rootCmd.PersistentFlags() == nil {
		return false
	}

	// Check if any relevant flags are changed
	flags := []string{"verbose", "no-color", "output", "config"}
	for _, flag := range flags {
		if rootCmd.PersistentFlags().Changed(flag) {
			return true
		}
	}

	return false
}

// applyEnvOverrides applies environment variable overrides
func applyEnvOverrides(config *Config) {
	// Check NO_COLOR environment variable (standard)
	if os.Getenv("NO_COLOR") != "" {
		config.CLI.NoColor = true
	}

	// Check ZEN_DEBUG for development mode
	if os.Getenv("ZEN_DEBUG") == "true" {
		config.Development.Debug = true
		config.LogLevel = "debug"
	}
}

// validate validates the configuration with comprehensive error messages
func validate(config *Config) error {
	// Validate log level
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLogLevels, config.LogLevel) {
		return &ValidationError{
			Field:        "log_level",
			Value:        config.LogLevel,
			ValidOptions: validLogLevels,
			Message:      "invalid log level",
		}
	}

	// Validate log format
	validLogFormats := []string{"text", "json"}
	if !contains(validLogFormats, config.LogFormat) {
		return &ValidationError{
			Field:        "log_format",
			Value:        config.LogFormat,
			ValidOptions: validLogFormats,
			Message:      "invalid log format",
		}
	}

	// Validate output format
	validOutputFormats := []string{"text", "json", "yaml"}
	if !contains(validOutputFormats, config.CLI.OutputFormat) {
		return &ValidationError{
			Field:        "cli.output_format",
			Value:        config.CLI.OutputFormat,
			ValidOptions: validOutputFormats,
			Message:      "invalid output format",
		}
	}

	// Validate workspace root exists and is accessible (only for non-default values)
	if config.Workspace.Root != "" && config.Workspace.Root != "." {
		if info, err := os.Stat(config.Workspace.Root); err != nil {
			if os.IsNotExist(err) {
				return &ValidationError{
					Field:   "workspace.root",
					Value:   config.Workspace.Root,
					Message: "workspace root directory does not exist",
				}
			}
			return errors.Wrap(err, "failed to access workspace root")
		} else if !info.IsDir() {
			return &ValidationError{
				Field:   "workspace.root",
				Value:   config.Workspace.Root,
				Message: "workspace root must be a directory",
			}
		}
	}

	return nil
}

// ValidationError represents a configuration validation error with helpful context
type ValidationError struct {
	Field        string
	Value        string
	ValidOptions []string
	Message      string
}

func (e *ValidationError) Error() string {
	if len(e.ValidOptions) > 0 {
		return fmt.Sprintf("%s: %s (got %q, valid options: %s)",
			e.Field, e.Message, e.Value, strings.Join(e.ValidOptions, ", "))
	}
	return fmt.Sprintf("%s: %s (got %q)", e.Field, e.Message, e.Value)
}

// IsValidationError checks if an error is a ValidationError
func IsValidationError(err error) bool {
	_, ok := err.(*ValidationError)
	return ok
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// GetConfigFile returns the path to the configuration file that was loaded
func (c *Config) GetConfigFile() string {
	return c.configFile
}

// GetLoadedSources returns the sources from which configuration was loaded
func (c *Config) GetLoadedSources() []string {
	return c.loadedFrom
}

// SetValue sets a configuration value and writes it to the local config file
func (c *Config) SetValue(key, value string) error {
	// Validate the key and value
	if err := ValidateKey(key); err != nil {
		return err
	}
	if err := ValidateValue(key, value); err != nil {
		return err
	}

	// Set the value in Viper
	c.viper.Set(key, value)

	// Write to local config file (.zen/config)
	configPath := ".zen/config"

	// Ensure .zen directory exists
	if err := os.MkdirAll(".zen", 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the configuration
	if err := c.viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Redacted returns a copy of the Config with sensitive fields redacted for display
func (c *Config) Redacted() Config {
	redacted := *c
	redacted.Assets = c.Assets.Redacted()
	return redacted
}

// IsSensitiveField checks if a field contains sensitive information that should be redacted
func IsSensitiveField(fieldName string) bool {
	sensitive := []string{
		"api_key", "token", "secret", "password", "key",
		"auth", "credential", "private", "cert", "pem",
		"repository_url", "repo", // Asset repository URLs should be obfuscated
	}

	fieldLower := strings.ToLower(fieldName)
	for _, s := range sensitive {
		if strings.Contains(fieldLower, s) {
			return true
		}
	}
	return false
}

// RedactSensitiveValue redacts sensitive values for logging/display
func RedactSensitiveValue(fieldName, value string) string {
	if IsSensitiveField(fieldName) {
		if len(value) <= 4 {
			return "***"
		}
		return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
	}
	return value
}

// LoadDefaults returns a configuration with default values from Options
func LoadDefaults() *Config {
	// Create a temporary viper instance to get defaults
	v := viper.New()
	setDefaults(v)

	// Unmarshal into config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		// This should not happen if Options are properly defined
		panic(fmt.Sprintf("failed to unmarshal default configuration: %v", err))
	}

	cfg.viper = v
	cfg.loadedFrom = []string{"defaults"}
	return &cfg
}
