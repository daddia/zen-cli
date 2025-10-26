package development

import (
	"fmt"

	"github.com/daddia/zen/internal/config"
	"github.com/go-viper/mapstructure/v2"
)

// Config contains development-specific settings
type Config struct {
	// Enable debug mode
	Debug bool `yaml:"debug" json:"debug" mapstructure:"debug"`

	// Enable profiling
	Profile bool `yaml:"profile" json:"profile" mapstructure:"profile"`
}

// DefaultConfig returns default development configuration
func DefaultConfig() Config {
	return Config{
		Debug:   false,
		Profile: false,
	}
}

// Implement config.Configurable interface

// Validate validates the development configuration
func (c Config) Validate() error {
	// No specific validation needed for boolean fields
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
		return cfg, fmt.Errorf("failed to decode development config: %w", err)
	}

	return cfg, nil
}

// Section returns the configuration section name for development
func (p ConfigParser) Section() string {
	return "development"
}
