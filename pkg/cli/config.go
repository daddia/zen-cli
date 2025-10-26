package cli

import (
	"fmt"

	"github.com/daddia/zen/internal/config"
	"github.com/go-viper/mapstructure/v2"
)

// Config contains CLI-specific configuration
type Config struct {
	// Color output settings
	NoColor bool `yaml:"no_color" json:"no_color" mapstructure:"no_color"`

	// Verbose output
	Verbose bool `yaml:"verbose" json:"verbose" mapstructure:"verbose"`

	// Output format (text, json, yaml)
	OutputFormat string `yaml:"output_format" json:"output_format" mapstructure:"output_format"`
}

// DefaultConfig returns default CLI configuration
func DefaultConfig() Config {
	return Config{
		NoColor:      false,
		Verbose:      false,
		OutputFormat: "text",
	}
}

// Implement config.Configurable interface

// Validate validates the CLI configuration
func (c Config) Validate() error {
	validFormats := []string{"text", "json", "yaml"}
	validFormat := false
	for _, f := range validFormats {
		if c.OutputFormat == f {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid output_format: %s (must be one of: text, json, yaml)", c.OutputFormat)
	}

	return nil
}

// Defaults returns a new Config with default values
func (c Config) Defaults() config.Configurable {
	return DefaultConfig()
}

// ConfigParser implements config.ConfigParser[Config] interface
type ConfigParser struct{}

// Parse converts raw configuration data to Config
func (p ConfigParser) Parse(raw map[string]interface{}) (Config, error) {
	// Start with defaults to ensure all fields are properly initialized
	cfg := DefaultConfig()

	// If raw data is empty, return defaults
	if len(raw) == 0 {
		return cfg, nil
	}

	// Use mapstructure to decode the raw map into our config struct
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &cfg,
		WeaklyTypedInput: true,
	})
	if err != nil {
		return cfg, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(raw); err != nil {
		return cfg, fmt.Errorf("failed to decode CLI config: %w", err)
	}

	return cfg, nil
}

// Section returns the configuration section name for CLI
func (p ConfigParser) Section() string {
	return "cli"
}
