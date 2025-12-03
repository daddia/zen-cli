package list

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

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

// ListOptions contains options for the list command
type ListOptions struct {
	IO     *iostreams.IOStreams
	Config func() (*config.Config, error)
}

// NewCmdConfigList creates the config list command
func NewCmdConfigList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		IO:     f.IOStreams,
		Config: f.Config,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Print a list of configuration keys and values",
		Aliases: []string{"ls"},
		Long: `List all configuration keys and their current values.

This shows the effective configuration after loading from files,
environment variables, and command-line flags.

Configuration is organized by component:
- Core: log_level, log_format
- Assets: repository_url, branch, cache settings
- Workspace: root, zen_path
- CLI: no_color, verbose, output_format
- Development: debug, profile
- Task: source, sync, project_key
- Auth: storage_type, validation_timeout
- Cache: base_path, size_limit_mb
- Templates: cache_enabled, cache_ttl`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			if runF != nil {
				return runF(opts)
			}

			return listRun(opts)
		},
	}

	return cmd
}

func listRun(opts *ListOptions) error {
	// Get central config manager
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// List available components
	components := []string{"assets", "auth", "cache", "cli", "development", "task", "templates", "workspace"}

	// Display core configuration first
	fmt.Fprintln(opts.IO.Out, "[core]")
	fmt.Fprintf(opts.IO.Out, "log_level = %s\n", cfg.Core.LogLevel)
	fmt.Fprintf(opts.IO.Out, "log_format = %s\n", cfg.Core.LogFormat)
	fmt.Fprintln(opts.IO.Out)

	// Display each component configuration
	for _, component := range components {
		switch component {
		case "assets":
			listComponentConfig(cfg, assets.ConfigParser{}, opts.IO)
		case "auth":
			listComponentConfig(cfg, auth.ConfigParser{}, opts.IO)
		case "cache":
			listComponentConfig(cfg, cache.ConfigParser{}, opts.IO)
		case "cli":
			listComponentConfig(cfg, cli.ConfigParser{}, opts.IO)
		case "development":
			listComponentConfig(cfg, development.ConfigParser{}, opts.IO)
		case "task":
			listComponentConfig(cfg, task.ConfigParser{}, opts.IO)
		case "templates":
			listComponentConfig(cfg, template.ConfigParser{}, opts.IO)
		case "workspace":
			listComponentConfig(cfg, workspace.ConfigParser{}, opts.IO)
		}
	}

	return nil
}

// listComponentConfig lists configuration for a specific component
func listComponentConfig[T config.Configurable](cfg *config.Config, parser config.ConfigParser[T], io *iostreams.IOStreams) {
	componentConfig, err := config.GetConfig(cfg, parser)
	if err != nil {
		fmt.Fprintf(io.ErrOut, "Error loading %s config: %v\n", parser.Section(), err)
		return
	}

	fmt.Fprintf(io.Out, "[%s]\n", parser.Section())
	displayConfigStruct(componentConfig, io)
	fmt.Fprintln(io.Out)
}

// displayConfigStruct displays a configuration struct in a readable format
func displayConfigStruct(configStruct interface{}, io *iostreams.IOStreams) {
	v := reflect.ValueOf(configStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		fmt.Fprintf(io.Out, "%v\n", configStruct)
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Get the field name from yaml tag or use struct field name
		fieldName := fieldType.Name
		if yamlTag := fieldType.Tag.Get("yaml"); yamlTag != "" {
			fieldName = strings.Split(yamlTag, ",")[0]
		}

		// Skip unexported fields
		if !fieldType.IsExported() {
			continue
		}

		value := formatFieldValue(field)
		fmt.Fprintf(io.Out, "%s = %s\n", fieldName, value)
	}
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
