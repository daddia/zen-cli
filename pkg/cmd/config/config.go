package config

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/pkg/cmd/config/get"
	"github.com/daddia/zen/pkg/cmd/config/list"
	"github.com/daddia/zen/pkg/cmd/config/set"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewCmdConfig creates the config command with subcommands
func NewCmdConfig(f *cmdutil.Factory) *cobra.Command {
	longDoc := strings.Builder{}
	longDoc.WriteString("Display or change configuration settings for Zen CLI.\n\n")
	longDoc.WriteString("Current configuration options:\n")
	for _, co := range config.Options {
		longDoc.WriteString(fmt.Sprintf("- `%s`: %s", co.Key, co.Description))
		if len(co.AllowedValues) > 0 {
			longDoc.WriteString(fmt.Sprintf(" `{%s}`", strings.Join(co.AllowedValues, " | ")))
		}
		if co.DefaultValue != "" {
			longDoc.WriteString(fmt.Sprintf(" (default `%s`)", co.DefaultValue))
		}
		longDoc.WriteRune('\n')
	}

	cmd := &cobra.Command{
		Use:     "config <command>",
		Short:   "Manage configuration for Zen CLI",
		GroupID: "workspace",
		Long:    longDoc.String(),
		Example: `  # Display current configuration
  zen config

  # Get a specific configuration value
  zen config get log_level

  # Set a configuration value
  zen config set log_level debug

  # List all configuration with values
  zen config list

  # Output configuration as JSON
  zen config --output json

  # Use environment variables
  ZEN_LOG_LEVEL=debug zen config get log_level`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default behavior when no subcommand is provided - show current config
			return displayCurrentConfig(f, cmd)
		},
	}

	// Add subcommands
	cmd.AddCommand(get.NewCmdConfigGet(f, nil))
	cmd.AddCommand(set.NewCmdConfigSet(f, nil))
	cmd.AddCommand(list.NewCmdConfigList(f, nil))

	return cmd
}

// displayCurrentConfig shows the current configuration (default behavior)
func displayCurrentConfig(f *cmdutil.Factory, cmd *cobra.Command) error {
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

	// Display Core section
	if core, ok := configMap["core"].(map[string]interface{}); ok {
		fmt.Fprintln(out, iostreams.FormatBold("Core:"))
		if configDir, ok := core["config_dir"]; ok && configDir != "" {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Config Dir: %v\n", configDir), 1))
		}
		if token, ok := core["token"]; ok && token != "" {
			// Redact token for security
			fmt.Fprint(out, iostreams.Indent("Token:      [REDACTED]\n", 1))
		}
		if debug, ok := core["debug"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Debug:      %v\n", debug), 1))
		}
		if logLevel, ok := core["log_level"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Log Level:  %v\n", logLevel), 1))
		}
		if logFormat, ok := core["log_format"]; ok {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Log Format: %v\n", logFormat), 1))
		}
		fmt.Fprintln(out)
	}

	// Display Project section
	if project, ok := configMap["project"].(map[string]interface{}); ok {
		fmt.Fprintln(out, iostreams.FormatBold("Project:"))
		if name, ok := project["name"]; ok && name != "" {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Name: %v\n", name), 1))
		} else {
			fmt.Fprint(out, iostreams.Indent("Name: [auto-detected]\n", 1))
		}
		if path, ok := project["path"]; ok && path != "" {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Path: %v\n", path), 1))
		}
		fmt.Fprintln(out)
	}

	// Display Task section
	if task, ok := configMap["task"].(map[string]interface{}); ok {
		fmt.Fprintln(out, iostreams.FormatBold("Task:"))
		if taskPath, ok := task["task_path"]; ok && taskPath != "" {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Task Path:   %v\n", taskPath), 1))
		} else {
			fmt.Fprint(out, iostreams.Indent("Task Path:   .zen/tasks\n", 1))
		}
		if taskSource, ok := task["task_source"]; ok && taskSource != "" {
			fmt.Fprint(out, iostreams.Indent(fmt.Sprintf("Task Source: %v\n", taskSource), 1))
		} else {
			fmt.Fprint(out, iostreams.Indent("Task Source: local\n", 1))
		}
		fmt.Fprintln(out)
	}

	return nil
}
