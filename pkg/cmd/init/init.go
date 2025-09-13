package init

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jonathandaddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// NewCmdInit creates the init command
func NewCmdInit(f *cmdutil.Factory) *cobra.Command {
	var force bool
	var configFile string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new Zen workspace",
		Long: `Initialize a new Zen workspace in the current directory.

This command creates a zen.yaml configuration file with default settings
and sets up the necessary directory structure for your project.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get current directory
			cwd, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			// Determine config file path
			if configFile == "" {
				configFile = filepath.Join(cwd, "zen.yaml")
			}

			// Check if config already exists
			if _, err := os.Stat(configFile); err == nil && !force {
				return fmt.Errorf("configuration file already exists at %s (use --force to overwrite)", configFile)
			}

			// Create default configuration
			defaultConfig := map[string]interface{}{
				"version": "1.0",
				"workspace": map[string]interface{}{
					"root": cwd,
					"name": filepath.Base(cwd),
				},
				"logging": map[string]interface{}{
					"level":  "info",
					"format": "text",
				},
				"cli": map[string]interface{}{
					"output_format": "text",
					"no_color":      false,
				},
			}

			// Marshal to YAML
			data, err := yaml.Marshal(defaultConfig)
			if err != nil {
				return fmt.Errorf("failed to marshal configuration: %w", err)
			}

			// Add header comment
			configContent := []byte(`# Zen CLI Configuration
# AI-Powered Product Lifecycle Productivity Platform
# 
# For more information, visit: https://github.com/jonathandaddia/zen

`)
			configContent = append(configContent, data...)

			// Write configuration file
			if err := os.WriteFile(configFile, configContent, 0644); err != nil {
				return fmt.Errorf("failed to write configuration file: %w", err)
			}

			// Success message
			fmt.Fprintf(f.IOStreams.Out, "✅ Zen workspace initialized successfully!\n")
			fmt.Fprintf(f.IOStreams.Out, "📁 Configuration file: %s\n", configFile)
			fmt.Fprintln(f.IOStreams.Out, "\n🚀 Get started with 'zen status' to check your workspace")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing configuration")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "Path to configuration file")

	return cmd
}
