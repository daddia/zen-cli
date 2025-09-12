package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "zen.yaml")

	// Create a test config file
	configContent := `
log_level: debug
log_format: json
cli:
  no_color: true
  verbose: true
  output_format: yaml
workspace:
  root: /test
  config_file: test.yaml
development:
  debug: true
  profile: true
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// Change to temp directory so config is found
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Load configuration
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test loaded values
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}

	if cfg.LogFormat != "json" {
		t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, "json")
	}

	if !cfg.CLI.NoColor {
		t.Error("CLI.NoColor = false, want true")
	}

	if !cfg.CLI.Verbose {
		t.Error("CLI.Verbose = false, want true")
	}

	if cfg.CLI.OutputFormat != "yaml" {
		t.Errorf("CLI.OutputFormat = %q, want %q", cfg.CLI.OutputFormat, "yaml")
	}

	if cfg.Workspace.Root != "/test" {
		t.Errorf("Workspace.Root = %q, want %q", cfg.Workspace.Root, "/test")
	}

	if cfg.Workspace.ConfigFile != "test.yaml" {
		t.Errorf("Workspace.ConfigFile = %q, want %q", cfg.Workspace.ConfigFile, "test.yaml")
	}

	if !cfg.Development.Debug {
		t.Error("Development.Debug = false, want true")
	}

	if !cfg.Development.Profile {
		t.Error("Development.Profile = false, want true")
	}
}

func TestLoadDefaults(t *testing.T) {
	// Create a temporary directory with no config file
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldWd); err != nil {
			t.Errorf("Failed to restore working directory: %v", err)
		}
	}()

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Load configuration (should use defaults)
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Test default values
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}

	if cfg.LogFormat != "text" {
		t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, "text")
	}

	if cfg.CLI.NoColor {
		t.Error("CLI.NoColor = true, want false")
	}

	if cfg.CLI.Verbose {
		t.Error("CLI.Verbose = true, want false")
	}

	if cfg.CLI.OutputFormat != "text" {
		t.Errorf("CLI.OutputFormat = %q, want %q", cfg.CLI.OutputFormat, "text")
	}

	if cfg.Workspace.Root != "." {
		t.Errorf("Workspace.Root = %q, want %q", cfg.Workspace.Root, ".")
	}

	if cfg.Workspace.ConfigFile != "zen.yaml" {
		t.Errorf("Workspace.ConfigFile = %q, want %q", cfg.Workspace.ConfigFile, "zen.yaml")
	}

	if cfg.Development.Debug {
		t.Error("Development.Debug = true, want false")
	}

	if cfg.Development.Profile {
		t.Error("Development.Profile = true, want false")
	}
}

func TestApplyEnvOverrides(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		expected func(*Config) bool
	}{
		{
			name: "NO_COLOR environment variable",
			envVars: map[string]string{
				"NO_COLOR": "1",
			},
			expected: func(cfg *Config) bool {
				return cfg.CLI.NoColor
			},
		},
		{
			name: "ZEN_DEBUG environment variable",
			envVars: map[string]string{
				"ZEN_DEBUG": "true",
			},
			expected: func(cfg *Config) bool {
				return cfg.Development.Debug && cfg.LogLevel == "debug"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				if err := os.Setenv(key, value); err != nil {
					t.Fatalf("Failed to set env var %s: %v", key, err)
				}
				defer func(k string) {
					if err := os.Unsetenv(k); err != nil {
						t.Errorf("Failed to unset env var %s: %v", k, err)
					}
				}(key)
			}

			cfg := &Config{
				LogLevel:  "info",
				LogFormat: "text",
				CLI: CLIConfig{
					NoColor:      false,
					Verbose:      false,
					OutputFormat: "text",
				},
				Workspace: WorkspaceConfig{
					Root:       ".",
					ConfigFile: "zen.yaml",
				},
				Development: DevelopmentConfig{
					Debug:   false,
					Profile: false,
				},
			}

			applyEnvOverrides(cfg)

			if !tt.expected(cfg) {
				t.Errorf("Environment override not applied correctly")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		wantError bool
	}{
		{
			name: "valid config",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "text",
				CLI: CLIConfig{
					OutputFormat: "text",
				},
			},
			wantError: false,
		},
		{
			name: "invalid log level",
			config: &Config{
				LogLevel:  "invalid",
				LogFormat: "text",
				CLI: CLIConfig{
					OutputFormat: "text",
				},
			},
			wantError: true,
		},
		{
			name: "invalid log format",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "invalid",
				CLI: CLIConfig{
					OutputFormat: "text",
				},
			},
			wantError: true,
		},
		{
			name: "invalid output format",
			config: &Config{
				LogLevel:  "info",
				LogFormat: "text",
				CLI: CLIConfig{
					OutputFormat: "invalid",
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.config)
			if (err != nil) != tt.wantError {
				t.Errorf("validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		item     string
		expected bool
	}{
		{
			name:     "item exists",
			slice:    []string{"a", "b", "c"},
			item:     "b",
			expected: true,
		},
		{
			name:     "item does not exist",
			slice:    []string{"a", "b", "c"},
			item:     "d",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			item:     "a",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := contains(tt.slice, tt.item)
			if result != tt.expected {
				t.Errorf("contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}
