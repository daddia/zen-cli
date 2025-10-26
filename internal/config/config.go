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

// Standard Configuration Interfaces

// Configurable is the standard interface that all configuration types must implement
type Configurable interface {
	// Validate validates the configuration and returns an error if invalid
	Validate() error

	// Defaults returns a new instance with default values
	Defaults() Configurable
}

// ConfigParser provides type-safe parsing for configuration sections
type ConfigParser[T Configurable] interface {
	// Parse converts raw configuration data to the typed configuration
	Parse(raw map[string]interface{}) (T, error)

	// Section returns the configuration section name (e.g., "assets", "templates")
	Section() string
}

// Manager provides the standard configuration management APIs
// Note: Generic methods cannot be in interfaces, so Manager is a concrete type alias
type Manager = Config

// Config represents the core application configuration and implements Manager interface
type Config struct {
	// Core application configuration
	LogLevel  string `mapstructure:"log_level" validate:"required,oneof=trace debug info warn error fatal panic"`
	LogFormat string `mapstructure:"log_format" validate:"required,oneof=text json"`

	// Temporary fields - to be removed when components are fully migrated
	Integrations IntegrationsConfig `mapstructure:"integrations"`
	Work         WorkConfig         `mapstructure:"work"`

	// Internal fields for configuration management
	viper      *viper.Viper `mapstructure:"-" json:"-" yaml:"-"`
	configFile string       `mapstructure:"-" json:"-" yaml:"-"`
	loadedFrom []string     `mapstructure:"-" json:"-" yaml:"-"`
}

// Temporary types - to be removed when integration component is created
type IntegrationsConfig struct {
	TaskSystem        string                               `mapstructure:"task_system"`
	SyncEnabled       bool                                 `mapstructure:"sync_enabled"`
	SyncFrequency     string                               `mapstructure:"sync_frequency"`
	PluginDirectories []string                             `mapstructure:"plugin_directories"`
	Providers         map[string]IntegrationProviderConfig `mapstructure:"providers"`
}

type IntegrationProviderConfig struct {
	URL           string                 `mapstructure:"url"`
	ProjectKey    string                 `mapstructure:"project_key"`
	Type          string                 `mapstructure:"type"`
	Credentials   string                 `mapstructure:"credentials"`
	Email         string                 `mapstructure:"email"`
	APIKey        string                 `mapstructure:"api_key"`
	FieldMapping  map[string]string      `mapstructure:"field_mapping"`
	SyncDirection string                 `mapstructure:"sync_direction"`
	Settings      map[string]interface{} `mapstructure:"settings"`
}

type WorkConfig struct {
	Tasks TasksConfig `mapstructure:"tasks"`
}

type TasksConfig struct {
	Source     string `mapstructure:"source"`
	Sync       string `mapstructure:"sync"`
	ProjectKey string `mapstructure:"project_key"`
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

	// Validate core configuration
	if err := validateCore(&config); err != nil {
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

// applyEnvOverrides applies environment variable overrides for core config only
func applyEnvOverrides(config *Config) {
	// Check ZEN_DEBUG for development mode (affects core log level)
	if os.Getenv("ZEN_DEBUG") == "true" {
		config.LogLevel = "debug"
	}
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

// Manager Interface Implementation

// GetConfig retrieves typed configuration for a component using the standard interface
func GetConfig[T Configurable](c *Config, parser ConfigParser[T]) (T, error) {
	// Extract the raw configuration section
	sectionData := c.viper.GetStringMap(parser.Section())

	// If section doesn't exist or is empty, return defaults
	if len(sectionData) == 0 {
		// Parse empty data to get a zero value, then get its defaults
		config, err := parser.Parse(map[string]interface{}{})
		if err != nil {
			// If parsing empty data fails, create zero value and get defaults
			var zero T
			defaults := zero.Defaults()
			if typed, ok := defaults.(T); ok {
				return typed, nil
			}
			return zero, fmt.Errorf("failed to get defaults for section %s", parser.Section())
		}

		// Get defaults from the parsed (empty) config
		defaults := config.Defaults()
		if typed, ok := defaults.(T); ok {
			return typed, nil
		}
		return config, fmt.Errorf("failed to get defaults for section %s", parser.Section())
	}

	// Parse the raw data using the component's parser
	config, err := parser.Parse(sectionData)
	if err != nil {
		// On parse error, return defaults
		var zero T
		defaults := zero.Defaults()
		if typed, ok := defaults.(T); ok {
			return typed, fmt.Errorf("failed to parse config section %s: %w", parser.Section(), err)
		}
		return zero, fmt.Errorf("failed to parse config section %s: %w", parser.Section(), err)
	}

	// Validate the parsed configuration
	if err := config.Validate(); err != nil {
		return config, fmt.Errorf("config validation failed for section %s: %w", parser.Section(), err)
	}

	return config, nil
}

// SetConfig updates typed configuration for a component using the standard interface
func SetConfig[T Configurable](c *Config, parser ConfigParser[T], config T) error {
	// Validate the configuration before setting
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed for section %s: %w", parser.Section(), err)
	}

	// Convert config to map for Viper
	sectionKey := parser.Section()

	// Set the entire section in Viper
	c.viper.Set(sectionKey, config)

	// Write to local config file (.zen/config)
	// Use absolute path to ensure we write to the correct location
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Find the project root by looking for go.mod
	projectRoot := workingDir
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			// Reached filesystem root, use current directory
			projectRoot = workingDir
			break
		}
		projectRoot = parent
	}

	configDir := filepath.Join(projectRoot, ".zen")
	configPath := filepath.Join(configDir, "config")

	// Ensure .zen directory exists with secure permissions
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write the configuration directly to the file
	if err := c.viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Create a fresh Viper instance to reload the configuration
	// This avoids issues with cached state in the existing instance
	freshViper := viper.New()
	freshViper.SetConfigName("config")
	freshViper.SetConfigType("yaml")
	freshViper.AddConfigPath(configDir)

	if err := freshViper.ReadInConfig(); err != nil {
		return fmt.Errorf("failed to reload config after write: %w", err)
	}

	// Replace the existing Viper instance with the fresh one
	c.viper = freshViper

	return nil
}

// GetConfigFile returns the path to the configuration file that was loaded
func (c *Config) GetConfigFile() string {
	return c.configFile
}

// GetLoadedSources returns the sources from which configuration was loaded
func (c *Config) GetLoadedSources() []string {
	return c.loadedFrom
}

// GetViper returns the underlying Viper instance for debugging
func (c *Config) GetViper() *viper.Viper {
	return c.viper
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
	// Use absolute path to ensure we write to the correct location
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Find the project root by looking for go.mod
	projectRoot := workingDir
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			// Reached filesystem root, use current directory
			projectRoot = workingDir
			break
		}
		projectRoot = parent
	}

	configDir := filepath.Join(projectRoot, ".zen")
	configPath := filepath.Join(configDir, "config")

	// Ensure .zen directory exists with secure permissions
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Set config type before writing
	c.viper.SetConfigType("yaml")

	// Write the configuration directly to the file
	if err := c.viper.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Redacted returns a copy of the Config with sensitive fields redacted for display
func (c *Config) Redacted() Config {
	redacted := *c
	// Note: Component-specific redaction is now handled by each component's config
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

// validateCore validates the core configuration fields only
func validateCore(config *Config) error {
	// Validate log level
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	valid := false
	for _, level := range validLogLevels {
		if config.LogLevel == level {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log_level: %s (must be one of: %s)",
			config.LogLevel, strings.Join(validLogLevels, ", "))
	}

	// Validate log format
	validLogFormats := []string{"text", "json"}
	valid = false
	for _, format := range validLogFormats {
		if config.LogFormat == format {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid log_format: %s (must be one of: %s)",
			config.LogFormat, strings.Join(validLogFormats, ", "))
	}

	return nil
}
