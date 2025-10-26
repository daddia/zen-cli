package workspace

import (
	"fmt"

	"github.com/daddia/zen/internal/config"
	"github.com/go-viper/mapstructure/v2"
)

// Config contains workspace-specific configuration
type Config struct {
	// Root directory for workspace detection
	Root string `mapstructure:"root" yaml:"root" json:"root"`

	// Zen directory path relative to workspace root
	ZenPath string `mapstructure:"zen_path" yaml:"zen_path" json:"zen_path"`
}

// DefaultConfig returns default workspace configuration
func DefaultConfig() Config {
	return Config{
		Root:    ".",
		ZenPath: ".zen",
	}
}

// Implement config.Configurable interface

// Validate validates the workspace configuration
func (c Config) Validate() error {
	if c.Root == "" {
		return fmt.Errorf("root is required")
	}
	if c.ZenPath == "" {
		return fmt.Errorf("zen_path is required")
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
		return cfg, fmt.Errorf("failed to decode workspace config: %w", err)
	}

	return cfg, nil
}

// Section returns the configuration section name for workspace
func (p ConfigParser) Section() string {
	return "workspace"
}
