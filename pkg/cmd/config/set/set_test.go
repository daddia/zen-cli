package set

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestSetRun(t *testing.T) {
	tests := []struct {
		name            string
		key             string
		value           string
		expectedError   string
		checkConfigFile bool
	}{
		{
			name:            "set valid log_level",
			key:             "log_level",
			value:           "debug",
			checkConfigFile: true,
		},
		{
			name:            "set valid boolean",
			key:             "cli.verbose",
			value:           "true",
			checkConfigFile: true,
		},
		{
			name:            "set workspace root",
			key:             "workspace.root",
			value:           "/custom/path",
			checkConfigFile: true,
		},
		{
			name:          "invalid key",
			key:           "invalid.key",
			value:         "value",
			expectedError: "unknown configuration key",
		},
		{
			name:          "invalid value for enum",
			key:           "log_level",
			value:         "invalid",
			expectedError: "valid values are",
		},
		{
			name:          "invalid boolean value",
			key:           "cli.verbose",
			value:         "maybe",
			expectedError: "valid values are",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for config
			tempDir := t.TempDir()

			// Change to temp directory
			oldWd, err := os.Getwd()
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.Chdir(oldWd))
			}()
			require.NoError(t, os.Chdir(tempDir))

			ios := iostreams.Test()
			out := ios.Out.(*bytes.Buffer)
			errOut := ios.ErrOut.(*bytes.Buffer)

			testConfig := &config.Config{
				LogLevel:  "info",
				LogFormat: "text",
			}

			opts := &SetOptions{
				IO:    ios,
				Key:   tt.key,
				Value: tt.value,
				Config: func() (*config.Config, error) {
					return testConfig, nil
				},
			}

			err = setRun(opts)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				if strings.Contains(errOut.String(), "warning") {
					// Key validation warning is acceptable
					require.NoError(t, err)
				} else {
					require.NoError(t, err)
				}

				// Check success message
				output := out.String()
				assert.Contains(t, output, "✓ Set "+tt.key+" to")
				assert.Contains(t, output, "Configuration saved to")

				if tt.checkConfigFile {
					// Verify config file was created and contains the value
					configPath := filepath.Join(".zen", "config.yaml")
					require.FileExists(t, configPath)

					data, err := os.ReadFile(configPath)
					require.NoError(t, err)

					var configData map[string]interface{}
					err = yaml.Unmarshal(data, &configData)
					require.NoError(t, err)

					// Verify the value was set correctly
					value := getNestedValue(configData, tt.key)
					if tt.key == "cli.verbose" && tt.value == "true" {
						assert.Equal(t, true, value)
					} else {
						assert.Equal(t, tt.value, value)
					}
				}
			}
		})
	}
}

func TestSetNestedValue(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		value         string
		expectedError string
	}{
		{
			name:  "simple key",
			key:   "log_level",
			value: "debug",
		},
		{
			name:  "nested key",
			key:   "cli.verbose",
			value: "true",
		},
		{
			name:  "deep nested key",
			key:   "workspace.config_file",
			value: "custom.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := make(map[string]interface{})

			err := setNestedValue(data, tt.key, tt.value)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)

				// Verify the value was set
				value := getNestedValue(data, tt.key)
				if strings.Contains(tt.key, "verbose") && tt.value == "true" {
					assert.Equal(t, true, value)
				} else {
					assert.Equal(t, tt.value, value)
				}
			}
		})
	}
}

// Helper function to get nested values for testing
func getNestedValue(data map[string]interface{}, key string) interface{} {
	parts := strings.Split(key, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			return current[part]
		}

		if nested, ok := current[part].(map[string]interface{}); ok {
			current = nested
		} else {
			return nil
		}
	}

	return nil
}
