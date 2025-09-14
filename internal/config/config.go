package config

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Development configuration
	Development DevelopmentConfig `mapstructure:"development"`

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

	// Configuration file name
	ConfigFile string `mapstructure:"config_file"`
}

// DevelopmentConfig contains development-specific settings
type DevelopmentConfig struct {
	// Enable debug mode
	Debug bool `mapstructure:"debug"`

	// Enable profiling
	Profile bool `mapstructure:"profile"`
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

// LoadWithOptions loads configuration with the provided options
func LoadWithOptions(opts LoadOptions) (*Config, error) {
	start := time.Now()
	v := viper.New()

	// Set configuration defaults
	setDefaults(v)

	// Configure file discovery
	configureFileDiscovery(v, opts.ConfigFile)

	// Configure environment variables
	configureEnvironment(v)

	// Bind CLI flags if command is provided
	if opts.Command != nil {
		if err := bindFlags(v, opts.Command); err != nil {
			return nil, errors.Wrap(err, "failed to bind CLI flags")
		}
	}

	// Track configuration sources
	var loadedFrom []string
	configFile := ""

	// Read configuration file (optional)
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

	// Check for environment variables
	if hasEnvVars() {
		loadedFrom = append(loadedFrom, "environment")
	}

	// Check for CLI flags
	if opts.Command != nil && hasFlagOverrides(opts.Command) {
		loadedFrom = append(loadedFrom, "flags")
	}

	// Unmarshal configuration
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

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Logging defaults
	v.SetDefault("log_level", "info")
	v.SetDefault("log_format", "text")

	// CLI defaults
	v.SetDefault("cli.no_color", false)
	v.SetDefault("cli.verbose", false)
	v.SetDefault("cli.output_format", "text")

	// Workspace defaults
	v.SetDefault("workspace.root", ".")
	v.SetDefault("workspace.config_file", "zen.yaml")

	// Development defaults
	v.SetDefault("development.debug", false)
	v.SetDefault("development.profile", false)
}

// configureFileDiscovery sets up configuration file discovery paths
func configureFileDiscovery(v *viper.Viper, configFile string) {
	// If specific config file is provided, use it
	if configFile != "" {
		v.SetConfigFile(configFile)
		return
	}

	// Set configuration name and type
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	// Add configuration paths in precedence order
	// 1. Current directory .zen/config.yaml
	v.AddConfigPath("./.zen")

	// 2. User home directory ~/.zen/config.yaml
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".zen"))
	}

	// 3. XDG config directory (Linux/macOS)
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		v.AddConfigPath(filepath.Join(xdgConfig, "zen"))
	} else if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".config", "zen"))
	}

	// 4. System config directory (Unix)
	v.AddConfigPath("/etc/zen")

	// 5. Backwards compatibility
	v.AddConfigPath("./configs")
}

// configureEnvironment sets up environment variable handling
func configureEnvironment(v *viper.Viper) {
	// Set environment prefix
	v.SetEnvPrefix("ZEN")

	// Replace dots with underscores for nested config
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Enable automatic environment variable reading
	v.AutomaticEnv()
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

// hasEnvVars checks if any ZEN_ environment variables are set
func hasEnvVars() bool {
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "ZEN_") {
			return true
		}
	}
	return false
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

// IsSensitiveField checks if a field contains sensitive information that should be redacted
func IsSensitiveField(fieldName string) bool {
	sensitive := []string{
		"api_key", "token", "secret", "password", "key",
		"auth", "credential", "private", "cert", "pem",
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

// LoadDefaults returns a configuration with default values
func LoadDefaults() *Config {
	cfg := &Config{
		LogLevel:  "info",
		LogFormat: "text",
		CLI: CLIConfig{
			NoColor:      false,
			Verbose:      false,
			OutputFormat: "text",
		},
		Workspace: WorkspaceConfig{
			Root:       ".",
			ConfigFile: "zen.yaml",
		},
		Development: DevelopmentConfig{
			Debug:   false,
			Profile: false,
		},
	}
	cfg.loadedFrom = []string{"defaults"}
	return cfg
}
