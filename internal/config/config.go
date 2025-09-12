package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	// Logging configuration
	LogLevel  string `mapstructure:"log_level"`
	LogFormat string `mapstructure:"log_format"`

	// CLI configuration
	CLI CLIConfig `mapstructure:"cli"`

	// Workspace configuration
	Workspace WorkspaceConfig `mapstructure:"workspace"`

	// Development configuration
	Development DevelopmentConfig `mapstructure:"development"`
}

// CLIConfig contains CLI-specific configuration
type CLIConfig struct {
	// Color output settings
	NoColor bool `mapstructure:"no_color"`

	// Verbose output
	Verbose bool `mapstructure:"verbose"`

	// Output format (text, json, yaml)
	OutputFormat string `mapstructure:"output_format"`
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

// Load loads configuration from various sources
func Load() (*Config, error) {
	v := viper.New()

	// Set configuration defaults
	setDefaults(v)

	// Set configuration name and paths
	v.SetConfigName("zen")
	v.SetConfigType("yaml")

	// Add configuration paths
	if home, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(filepath.Join(home, ".zen"))
	}
	v.AddConfigPath(".")
	v.AddConfigPath("./configs")

	// Read environment variables
	v.SetEnvPrefix("ZEN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Read configuration file (optional)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults
	}

	// Unmarshal configuration
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(&config)

	// Validate configuration
	if err := validate(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

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

// validate validates the configuration
func validate(config *Config) error {
	// Validate log level
	validLogLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
	if !contains(validLogLevels, config.LogLevel) {
		return fmt.Errorf("invalid log level: %s (valid options: %s)",
			config.LogLevel, strings.Join(validLogLevels, ", "))
	}

	// Validate log format
	validLogFormats := []string{"text", "json"}
	if !contains(validLogFormats, config.LogFormat) {
		return fmt.Errorf("invalid log format: %s (valid options: %s)",
			config.LogFormat, strings.Join(validLogFormats, ", "))
	}

	// Validate output format
	validOutputFormats := []string{"text", "json", "yaml"}
	if !contains(validOutputFormats, config.CLI.OutputFormat) {
		return fmt.Errorf("invalid output format: %s (valid options: %s)",
			config.CLI.OutputFormat, strings.Join(validOutputFormats, ", "))
	}

	return nil
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
