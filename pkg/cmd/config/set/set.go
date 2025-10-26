package set

import (
	"fmt"
	"strings"

	"github.com/MakeNowJust/heredoc"
	"github.com/daddia/zen/internal/config"
	configcmd "github.com/daddia/zen/pkg/cmd/config"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
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
	component, field, err := configcmd.parseConfigKey(opts.Key)
	if err != nil {
		return fmt.Errorf("invalid config key %s: %w", opts.Key, err)
	}

	// Get central config manager
	cfg, err := opts.Config()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create component registry
	registry := configcmd.NewComponentRegistry()

	// Handle core config separately
	if component == "core" {
		if err := configcmd.setCoreConfigValue(cfg, field, opts.Value); err != nil {
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
	updatedConfig, err := configcmd.updateConfigField(componentConfig, field, value)
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
