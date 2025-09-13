package config

import (
	"encoding/json"
	"fmt"

	"github.com/jonathandaddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewCmdConfig creates the config command
func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Display current configuration",
		Long: `Display the current Zen configuration settings.

This shows the effective configuration after loading from files,
environment variables, and command-line flags.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get configuration
			cfg, err := f.Config()
			if err != nil {
				return fmt.Errorf("failed to load configuration: %w", err)
			}

			// Get output format from parent command if available
			outputFormat := "text"
			if cmd.Parent() != nil && cmd.Parent().PersistentFlags().Changed("output") {
				if val, err := cmd.Parent().PersistentFlags().GetString("output"); err == nil {
					outputFormat = val
				}
			}

			switch outputFormat {
			case "json":
				encoder := json.NewEncoder(f.IOStreams.Out)
				encoder.SetIndent("", "  ")
				return encoder.Encode(cfg)

			case "yaml":
				encoder := yaml.NewEncoder(f.IOStreams.Out)
				defer encoder.Close()
				return encoder.Encode(cfg)

			default:
				return displayTextConfig(f.IOStreams.Out, cfg)
			}
		},
	}

	return cmd
}

// displayTextConfig displays configuration in human-readable text format
func displayTextConfig(out interface{ Write([]byte) (int, error) }, cfg interface{}) error {
	// Type assert to get the actual config
	configMap, ok := cfg.(map[string]interface{})
	if !ok {
		// Try to convert via JSON marshaling
		data, err := json.Marshal(cfg)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(data, &configMap); err != nil {
			return err
		}
	}

	fmt.Fprintln(out, "Zen Configuration")
	fmt.Fprintln(out, "=================")
	fmt.Fprintln(out)

	// Display configuration sections
	if logging, ok := configMap["LogLevel"]; ok {
		fmt.Fprintln(out, "Logging:")
		fmt.Fprintf(out, "  Level:  %v\n", logging)
		if format, ok := configMap["LogFormat"]; ok {
			fmt.Fprintf(out, "  Format: %v\n", format)
		}
		fmt.Fprintln(out)
	}

	if cli, ok := configMap["CLI"].(map[string]interface{}); ok {
		fmt.Fprintln(out, "CLI:")
		if noColor, ok := cli["NoColor"]; ok {
			fmt.Fprintf(out, "  No Color:      %v\n", noColor)
		}
		if verbose, ok := cli["Verbose"]; ok {
			fmt.Fprintf(out, "  Verbose:       %v\n", verbose)
		}
		if outputFormat, ok := cli["OutputFormat"]; ok {
			fmt.Fprintf(out, "  Output Format: %v\n", outputFormat)
		}
		fmt.Fprintln(out)
	}

	if workspace, ok := configMap["Workspace"].(map[string]interface{}); ok {
		fmt.Fprintln(out, "Workspace:")
		if root, ok := workspace["Root"]; ok {
			fmt.Fprintf(out, "  Root:        %v\n", root)
		}
		if configFile, ok := workspace["ConfigFile"]; ok {
			fmt.Fprintf(out, "  Config File: %v\n", configFile)
		}
		fmt.Fprintln(out)
	}

	if dev, ok := configMap["Development"].(map[string]interface{}); ok {
		fmt.Fprintln(out, "Development:")
		if debug, ok := dev["Debug"]; ok {
			fmt.Fprintf(out, "  Debug:   %v\n", debug)
		}
		if profile, ok := dev["Profile"]; ok {
			fmt.Fprintf(out, "  Profile: %v\n", profile)
		}
	}

	return nil
}
