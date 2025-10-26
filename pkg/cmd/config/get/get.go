package get

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
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

// GetOptions contains options for the get command
type GetOptions struct {
	IO     *iostreams.IOStreams
	Config func() (*config.Config, error)
	Key    string
}

// NewCmdConfigGet creates the config get command
func NewCmdConfigGet(f *cmdutil.Factory, runF func(*GetOptions) error) *cobra.Command {
	opts := &GetOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Print the value of a given configuration key",
		Long: `Print the value of a configuration key.

Configuration keys use dot notation to access component values:
- log_level, log_format (core config)
- assets.repository_url, assets.branch
- workspace.root, workspace.zen_path
- cli.no_color, cli.verbose, cli.output_format
- development.debug, development.profile
- task.source, task.sync, task.project_key
- auth.storage_type, auth.validation_timeout
- cache.base_path, cache.size_limit_mb
- templates.cache_enabled, templates.cache_ttl`,
		Example: heredoc.Doc(`
			$ zen config get log_level
			$ zen config get assets.repository_url
			$ zen config get workspace.root
			$ zen config get cli.output_format
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Key = args[0]

			if runF != nil {
				return runF(opts)
			}

			return getRun(opts)
		},
	}

	return cmd
}

func getRun(opts *GetOptions) error {
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
		value, err := getCoreConfigValue(cfg, field)
		if err != nil {
			return fmt.Errorf("failed to get core config: %w", err)
		}
		fmt.Fprintln(opts.IO.Out, value)
		return nil
	}

	// Handle component config using standard APIs
	switch component {
	case "assets":
		return getComponentConfig(cfg, assets.ConfigParser{}, field, opts.IO)
	case "auth":
		return getComponentConfig(cfg, auth.ConfigParser{}, field, opts.IO)
	case "cache":
		return getComponentConfig(cfg, cache.ConfigParser{}, field, opts.IO)
	case "cli":
		return getComponentConfig(cfg, cli.ConfigParser{}, field, opts.IO)
	case "development":
		return getComponentConfig(cfg, development.ConfigParser{}, field, opts.IO)
	case "task":
		return getComponentConfig(cfg, task.ConfigParser{}, field, opts.IO)
	case "templates":
		return getComponentConfig(cfg, template.ConfigParser{}, field, opts.IO)
	case "workspace":
		return getComponentConfig(cfg, workspace.ConfigParser{}, field, opts.IO)
	default:
		return fmt.Errorf("unknown component: %s", component)
	}
}

// getComponentConfig gets a field value from a component configuration using the standard API
func getComponentConfig[T config.Configurable](cfg *config.Config, parser config.ConfigParser[T], field string, io *iostreams.IOStreams) error {
	// Get component config
	componentConfig, err := config.GetConfig(cfg, parser)
	if err != nil {
		return fmt.Errorf("failed to get %s config: %w", parser.Section(), err)
	}

	// Extract the field value
	value, err := extractFieldValue(componentConfig, field)
	if err != nil {
		return fmt.Errorf("failed to get field %s: %w", field, err)
	}

	fmt.Fprintln(io.Out, value)
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

// getCoreConfigValue gets a value from the core config
func getCoreConfigValue(cfg *config.Config, field string) (string, error) {
	switch field {
	case "log_level":
		return cfg.LogLevel, nil
	case "log_format":
		return cfg.LogFormat, nil
	default:
		return "", fmt.Errorf("unknown core config field: %s", field)
	}
}

// extractFieldValue extracts a field value from a configuration struct using reflection
func extractFieldValue(configStruct interface{}, fieldName string) (string, error) {
	v := reflect.ValueOf(configStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("config must be a struct, got %T", configStruct)
	}

	// Convert field name to struct field name (snake_case to PascalCase)
	structFieldName := toPascalCase(fieldName)

	field := v.FieldByName(structFieldName)
	if !field.IsValid() {
		return "", fmt.Errorf("field %s not found in config", fieldName)
	}

	return formatFieldValue(field), nil
}

// formatFieldValue formats a reflect.Value as a string for display
func formatFieldValue(field reflect.Value) string {
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Bool:
		return strconv.FormatBool(field.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			return time.Duration(field.Int()).String()
		}
		return strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64)
	case reflect.Struct:
		// Handle nested structs (like DefaultDelims in template config)
		if field.Type().Name() == "" { // Anonymous struct
			// For now, return a simple representation
			return fmt.Sprintf("%+v", field.Interface())
		}
		return fmt.Sprintf("%+v", field.Interface())
	default:
		return fmt.Sprintf("%v", field.Interface())
	}
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
