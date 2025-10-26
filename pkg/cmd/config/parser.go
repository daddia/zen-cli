package config

import (
	"fmt"
	"strings"
)

// parseConfigKey parses a configuration key into component and field parts
// Examples:
//   "assets.repository_url" → "assets", "repository_url"
//   "cli.no_color" → "cli", "no_color"
//   "log_level" → "core", "log_level"
func parseConfigKey(key string) (component, field string, err error) {
	if key == "" {
		return "", "", fmt.Errorf("config key cannot be empty")
	}

	parts := strings.SplitN(key, ".", 2)
	
	// Handle core config keys (no component prefix)
	if len(parts) == 1 {
		// Core config keys like "log_level", "log_format"
		coreKeys := map[string]bool{
			"log_level":  true,
			"log_format": true,
		}
		
		if coreKeys[key] {
			return "core", key, nil
		}
		
		return "", "", fmt.Errorf("invalid config key: %s (must be component.field or core key)", key)
	}
	
	// Component-specific keys
	component = parts[0]
	field = parts[1]
	
	if component == "" {
		return "", "", fmt.Errorf("component name cannot be empty in key: %s", key)
	}
	
	if field == "" {
		return "", "", fmt.Errorf("field name cannot be empty in key: %s", key)
	}
	
	return component, field, nil
}

// validateConfigKey validates that a configuration key is properly formatted
func validateConfigKey(key string) error {
	_, _, err := parseConfigKey(key)
	return err
}
