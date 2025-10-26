package set

import (
	"fmt"
	"strings"

	"reflect"
	"strconv"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/development"
	"github.com/daddia/zen/internal/workspace"
	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cache"
	"github.com/daddia/zen/pkg/cli"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/task"
	"github.com/daddia/zen/pkg/template"
	"github.com/spf13/cobra"
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
	// Parse the config key to determine component and field
	component, field, err := parseConfigKey(opts.Key)
	if err != nil {
		return fmt.Errorf("invalid config key %s: %w", opts.Key, err)
	}

	// Get central config manager
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Handle core config separately
	if component == "core" {
		if err := setCoreConfigValue(cfg, field, opts.Value); err != nil {
			return fmt.Errorf("failed to set core config: %w", err)
		}

		// Write core config using SetValue (legacy method for core config)
		if err := cfg.SetValue(opts.Key, opts.Value); err != nil {
			return fmt.Errorf("failed to save core config: %w", err)
		}

		fmt.Fprintf(opts.IO.Out, "✓ Set %s to %q\n", opts.Key, opts.Value)
		return nil
	}

	// Handle component config using standard APIs
	switch component {
	case "assets":
		return setComponentConfig(cfg, assets.ConfigParser{}, field, opts.Value, opts.IO)
	case "auth":
		return setComponentConfig(cfg, auth.ConfigParser{}, field, opts.Value, opts.IO)
	case "cache":
		return setComponentConfig(cfg, cache.ConfigParser{}, field, opts.Value, opts.IO)
	case "cli":
		return setComponentConfig(cfg, cli.ConfigParser{}, field, opts.Value, opts.IO)
	case "development":
		return setComponentConfig(cfg, development.ConfigParser{}, field, opts.Value, opts.IO)
	case "task":
		return setComponentConfig(cfg, task.ConfigParser{}, field, opts.Value, opts.IO)
	case "templates":
		return setComponentConfig(cfg, template.ConfigParser{}, field, opts.Value, opts.IO)
	case "workspace":
		return setComponentConfig(cfg, workspace.ConfigParser{}, field, opts.Value, opts.IO)
	default:
		return fmt.Errorf("unknown component: %s", component)
	}
}

// setComponentConfig sets a field in a component configuration using the standard API
func setComponentConfig[T config.Configurable](cfg *config.Config, parser config.ConfigParser[T], field, value string, io *iostreams.IOStreams) error {
	// Get current component config
	componentConfig, err := config.GetConfig(cfg, parser)
	if err != nil {
		return fmt.Errorf("failed to get %s config: %w", parser.Section(), err)
	}

	// Update the field
	updatedConfig, err := updateConfigField(componentConfig, field, value)
	if err != nil {
		return fmt.Errorf("failed to update field %s: %w", field, err)
	}

	// Cast back to the correct type
	typedConfig, ok := updatedConfig.(T)
	if !ok {
		return fmt.Errorf("failed to cast updated config to correct type")
	}

	// Set back using central config
	if err := config.SetConfig(cfg, parser, typedConfig); err != nil {
		return fmt.Errorf("failed to save %s config: %w", parser.Section(), err)
	}

	fmt.Fprintf(io.Out, "✓ Set %s.%s to %q\n", parser.Section(), field, value)
	return nil
}

// parseConfigKey parses a configuration key into component and field parts
func parseConfigKey(key string) (component, field string, err error) {
	if key == "" {
		return "", "", fmt.Errorf("config key cannot be empty")
	}

	parts := strings.SplitN(key, ".", 2)

	// Handle core config keys (no component prefix)
	if len(parts) == 1 {
		coreKeys := map[string]bool{
			"log_level":  true,
			"log_format": true,
		}

		if coreKeys[key] {
			return "core", key, nil
		}

		return "", "", fmt.Errorf("invalid config key: %s (must be component.field or core key)", key)
	}

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

// setCoreConfigValue sets a value in the core config
func setCoreConfigValue(cfg *config.Config, field, value string) error {
	switch field {
	case "log_level":
		validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
		valid := false
		for _, level := range validLevels {
			if value == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log_level: %s (must be one of: %s)", value, strings.Join(validLevels, ", "))
		}
		cfg.LogLevel = value
	case "log_format":
		validFormats := []string{"text", "json"}
		valid := false
		for _, format := range validFormats {
			if value == format {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log_format: %s (must be one of: %s)", value, strings.Join(validFormats, ", "))
		}
		cfg.LogFormat = value
	default:
		return fmt.Errorf("unknown core config field: %s", field)
	}

	return nil
}

// updateConfigField updates a field in a configuration struct and returns the updated struct
func updateConfigField(configStruct interface{}, fieldName, value string) (interface{}, error) {
	// Create a copy of the struct
	v := reflect.ValueOf(configStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config must be a struct, got %T", configStruct)
	}

	// Create a new struct of the same type
	newStruct := reflect.New(v.Type()).Elem()
	newStruct.Set(v) // Copy all fields

	// Convert field name to struct field name
	structFieldName := toPascalCase(fieldName)

	field := newStruct.FieldByName(structFieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found in config", fieldName)
	}

	if !field.CanSet() {
		return nil, fmt.Errorf("field %s cannot be set", fieldName)
	}

	// Set the field value based on its type
	if err := setFieldValue(field, value); err != nil {
		return nil, fmt.Errorf("failed to set field %s: %w", fieldName, err)
	}

	return newStruct.Interface(), nil
}

// setFieldValue sets a reflect.Value from a string value
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		field.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration value: %s", value)
			}
			field.SetInt(int64(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value: %s", value)
			}
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value: %s", value)
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(floatVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}

	return nil
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}
