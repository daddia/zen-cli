package config

import (
	"encoding/json"
	"fmt"

	"github.com/daddia/zen/pkg/cmdutil"
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
				return displayTextConfig(f.IOStreams.Out, cfg, f.IOStreams)
			}
		},
	}

	return cmd
}

// displayTextConfig displays configuration in human-readable text format following design guide
func displayTextConfig(out interface{ Write([]byte) (int, error) }, cfg interface{}, iostreams interface {
	FormatSectionHeader(string) string
	FormatBold(string) string
	Indent(string, int) string
}) error {
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

	// Main header following design guide typography
	fmt.Fprintln(out, iostreams.FormatSectionHeader("Zen Configuration"))
	fmt.Fprintln(out)

	// Display configuration sections with proper formatting and indentation
	if logging, ok := configMap["LogLevel"]; ok {
		fmt.Fprintln(out, iostreams.FormatBold("Logging:"))
		fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Level:  %v\n", logging), 1))
		if format, ok := configMap["LogFormat"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Format: %v\n", format), 1))
		}
		fmt.Fprintln(out)
	}

	if cli, ok := configMap["CLI"].(map[string]interface{}); ok {
		fmt.Fprintln(out, iostreams.FormatBold("CLI:"))
		if noColor, ok := cli["NoColor"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("No Color:      %v\n", noColor), 1))
		}
		if verbose, ok := cli["Verbose"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Verbose:       %v\n", verbose), 1))
		}
		if outputFormat, ok := cli["OutputFormat"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Output Format: %v\n", outputFormat), 1))
		}
		fmt.Fprintln(out)
	}

	if workspace, ok := configMap["Workspace"].(map[string]interface{}); ok {
		fmt.Fprintln(out, iostreams.FormatBold("Workspace:"))
		if root, ok := workspace["Root"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Root:        %v\n", root), 1))
		}
		if configFile, ok := workspace["ConfigFile"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Config File: %v\n", configFile), 1))
		}
		fmt.Fprintln(out)
	}

	if dev, ok := configMap["Development"].(map[string]interface{}); ok {
		fmt.Fprintln(out, iostreams.FormatBold("Development:"))
		if debug, ok := dev["Debug"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Debug:   %v\n", debug), 1))
		}
		if profile, ok := dev["Profile"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Profile: %v\n", profile), 1))
		}
	}

	return nil
}
