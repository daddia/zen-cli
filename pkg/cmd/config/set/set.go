package set

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// SetOptions contains options for the set command
type SetOptions struct {
	IO     *iostreams.IOStreams
	Config func() (*config.Config, error)
	Key    string
	Value  string
}

// NewCmdConfigSet creates the config set command
func NewCmdConfigSet(f *cmdutil.Factory, runF func(*SetOptions) error) *cobra.Command {
	opts := &SetOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Update configuration with a value for the given key",
		Long: `Set a configuration value for the given key.

Configuration keys use dot notation to access nested values:
- log_level (trace, debug, info, warn, error, fatal, panic)
- log_format (text, json)
- cli.no_color (true, false)
- cli.verbose (true, false)
- cli.output_format (text, json, yaml)
- workspace.root (directory path)
- workspace.config_file (filename)
- development.debug (true, false)
- development.profile (true, false)

The configuration is saved to the first available location:
1. .zen/config.yaml (current directory)
2. ~/.zen/config.yaml (user home directory)`,
		Example: heredoc.Doc(`
			$ zen config set log_level debug
			$ zen config set cli.output_format json
			$ zen config set cli.no_color true
			$ zen config set workspace.root /path/to/workspace
		`),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Key = args[0]
			opts.Value = args[1]

			if runF != nil {
				return runF(opts)
			}

			return setRun(opts)
		},
	}

	return cmd
}

func setRun(opts *SetOptions) error {
	// Validate key
	if err := config.ValidateKey(opts.Key); err != nil {
		warning := opts.IO.FormatWarning("warning:")
		fmt.Fprintf(opts.IO.ErrOut, "%s %s\n", warning, err.Error())
	}

	// Validate value
	if err := config.ValidateValue(opts.Key, opts.Value); err != nil {
		if invalidValueErr, ok := err.(*config.InvalidValueError); ok {
			var values []string
			for _, v := range invalidValueErr.ValidValues {
				values = append(values, fmt.Sprintf("'%s'", v))
			}
			return fmt.Errorf("failed to set %q to %q: valid values are %s",
				opts.Key, opts.Value, strings.Join(values, ", "))
		}
		return fmt.Errorf("failed to validate value: %w", err)
	}

	// Load current configuration to get the config file path
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Determine config file path
	configPath := cfg.GetConfigFile()
	if configPath == "" {
		// No config file exists, create one in .zen directory
		zenDir := ".zen"
		if err := os.MkdirAll(zenDir, 0755); err != nil {
			// Try user home directory
			if home, homeErr := os.UserHomeDir(); homeErr == nil {
				zenDir = filepath.Join(home, ".zen")
				if err := os.MkdirAll(zenDir, 0755); err != nil {
					return fmt.Errorf("failed to create config directory: %w", err)
				}
			} else {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
		}
		configPath = filepath.Join(zenDir, "config.yaml")
	}

	// Load existing config file or create new structure
	var configData map[string]interface{}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configData = make(map[string]interface{})
	} else {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		if err := yaml.Unmarshal(data, &configData); err != nil {
			return fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	// Set the value in the config data
	if err := setNestedValue(configData, opts.Key, opts.Value); err != nil {
		return fmt.Errorf("failed to set configuration value: %w", err)
	}

	// Write the updated configuration
	data, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Fprintf(opts.IO.Out, "✓ Set %s to %q\n", opts.Key, opts.Value)
	fmt.Fprintf(opts.IO.Out, "Configuration saved to %s\n", configPath)

	return nil
}

// setNestedValue sets a value in nested map structure using dot notation
func setNestedValue(data map[string]interface{}, key, value string) error {
	parts := strings.Split(key, ".")
	current := data

	// Navigate to the parent of the target key
	for i, part := range parts[:len(parts)-1] {
		if _, exists := current[part]; !exists {
			current[part] = make(map[string]interface{})
		}

		nested, ok := current[part].(map[string]interface{})
		if !ok {
			return fmt.Errorf("configuration key path %q conflicts with existing non-object value at %q",
				key, strings.Join(parts[:i+1], "."))
		}
		current = nested
	}

	// Set the final value
	finalKey := parts[len(parts)-1]

	// Convert string values to appropriate types
	opt, found := config.FindOption(key)
	if found && opt.Type == "bool" {
		switch strings.ToLower(value) {
		case "true", "yes", "1":
			current[finalKey] = true
		case "false", "no", "0":
			current[finalKey] = false
		default:
			return fmt.Errorf("invalid boolean value %q for key %q", value, key)
		}
	} else {
		current[finalKey] = value
	}

	return nil
}
