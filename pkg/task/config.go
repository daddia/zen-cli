package task

import (
	"fmt"

	"github.com/daddia/zen/internal/config"
	"github.com/go-viper/mapstructure/v2"
)

// Config represents task management configuration
type Config struct {
	// Task source system (jira, github, linear, monday, asana, local, none)
	Source string `yaml:"source" json:"source" mapstructure:"source"`

	// Sync frequency (hourly, daily, manual, none)
	Sync string `yaml:"sync" json:"sync" mapstructure:"sync"`

	// Project key or identifier for tasks
	ProjectKey string `yaml:"project_key" json:"project_key" mapstructure:"project_key"`
}

// DefaultConfig returns default task configuration
func DefaultConfig() Config {
	return Config{
		Source:     "local",
		Sync:       "manual",
		ProjectKey: "",
	}
}

// Implement config.Configurable interface

// Validate validates the task configuration
func (c Config) Validate() error {
	validSources := []string{"jira", "github", "linear", "monday", "asana", "local", "none", ""}
	validSource := false
	for _, s := range validSources {
		if c.Source == s {
			validSource = true
			break
		}
	}
	if !validSource {
		return fmt.Errorf("invalid source: %s (must be one of: jira, github, linear, monday, asana, local, none)", c.Source)
	}

	validSyncs := []string{"hourly", "daily", "manual", "none", ""}
	validSync := false
	for _, s := range validSyncs {
		if c.Sync == s {
			validSync = true
			break
		}
	}
	if !validSync {
		return fmt.Errorf("invalid sync: %s (must be one of: hourly, daily, manual, none)", c.Sync)
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
		return cfg, fmt.Errorf("failed to decode task config: %w", err)
	}

	return cfg, nil
}

// Section returns the configuration section name for tasks
func (p ConfigParser) Section() string {
	return "task"
}
