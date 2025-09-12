package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/jonathandaddia/zen/internal/config"
	"github.com/jonathandaddia/zen/internal/logging"
	"gopkg.in/yaml.v3"
)

func TestNewVersionCommand(t *testing.T) {
	cfg := &config.Config{
		CLI: config.CLIConfig{
			OutputFormat: "text",
		},
	}
	logger := logging.NewBasic()

	cmd := newVersionCommand(cfg, logger)

	if cmd.Use != "version" {
		t.Errorf("Use = %q, want %q", cmd.Use, "version")
	}

	if !strings.Contains(cmd.Short, "version information") {
		t.Errorf("Short description doesn't contain expected text")
	}
}

func TestVersionCommandText(t *testing.T) {
	cfg := &config.Config{
		CLI: config.CLIConfig{
			OutputFormat: "text",
		},
	}
	logger := logging.NewBasic()

	cmd := newVersionCommand(cfg, logger)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Version command error = %v", err)
	}

	output := buf.String()

	expectedContent := []string{
		"zen version",
		"commit:",
		"built:",
		"go version:",
		"platform:",
	}

	for _, expected := range expectedContent {
		if !strings.Contains(output, expected) {
			t.Errorf("Version output missing expected content: %q", expected)
		}
	}
}

func TestVersionCommandJSON(t *testing.T) {
	cfg := &config.Config{
		CLI: config.CLIConfig{
			OutputFormat: "json",
		},
	}
	logger := logging.NewBasic()

	cmd := newVersionCommand(cfg, logger)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Version command error = %v", err)
	}

	output := buf.String()

	// Parse JSON output
	var versionInfo VersionInfo
	if err := json.Unmarshal([]byte(output), &versionInfo); err != nil {
		t.Errorf("Failed to parse JSON output: %v", err)
		return
	}

	// Check required fields
	if versionInfo.Version == "" {
		t.Error("Version field is empty")
	}

	if versionInfo.GoVersion == "" {
		t.Error("GoVersion field is empty")
	}

	if versionInfo.Platform == "" {
		t.Error("Platform field is empty")
	}
}

func TestVersionCommandYAML(t *testing.T) {
	cfg := &config.Config{
		CLI: config.CLIConfig{
			OutputFormat: "yaml",
		},
	}
	logger := logging.NewBasic()

	cmd := newVersionCommand(cfg, logger)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Errorf("Version command error = %v", err)
	}

	output := buf.String()

	// Parse YAML output
	var versionInfo VersionInfo
	if err := yaml.Unmarshal([]byte(output), &versionInfo); err != nil {
		t.Errorf("Failed to parse YAML output: %v", err)
		return
	}

	// Check required fields
	if versionInfo.Version == "" {
		t.Error("Version field is empty")
	}

	if versionInfo.GoVersion == "" {
		t.Error("GoVersion field is empty")
	}

	if versionInfo.Platform == "" {
		t.Error("Platform field is empty")
	}
}

func TestVersionCommandShort(t *testing.T) {
	cfg := &config.Config{
		CLI: config.CLIConfig{
			OutputFormat: "text",
		},
	}
	logger := logging.NewBasic()

	cmd := newVersionCommand(cfg, logger)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	// Set short flag
	cmd.SetArgs([]string{"--short"})

	err := cmd.Execute()
	if err != nil {
		t.Errorf("Version command with --short error = %v", err)
	}

	output := strings.TrimSpace(buf.String())

	// Should only contain the version number
	if output == "" {
		t.Errorf("Short version output is empty. Full output: %q", buf.String())
	}

	// For dev builds, the output should be "dev"
	if output != "dev" && output != version {
		t.Errorf("Unexpected short version output: %q (expected 'dev' or %q)", output, version)
	}

	// Should not contain extra information
	if strings.Contains(output, "commit:") || strings.Contains(output, "built:") {
		t.Errorf("Short version output contains extra information: %q", output)
	}
}

func TestVersionInfo(t *testing.T) {
	// Test the VersionInfo struct
	info := VersionInfo{
		Version:   "1.0.0",
		Commit:    "abc123",
		BuildTime: "2023-01-01T00:00:00Z",
		GoVersion: "go1.25.0",
		Platform:  "linux/amd64",
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(info)
	if err != nil {
		t.Errorf("Failed to marshal VersionInfo to JSON: %v", err)
	}

	var unmarshaled VersionInfo
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal VersionInfo from JSON: %v", err)
	}

	if unmarshaled != info {
		t.Errorf("JSON round-trip failed: got %+v, want %+v", unmarshaled, info)
	}

	// Test YAML marshaling
	yamlData, err := yaml.Marshal(info)
	if err != nil {
		t.Errorf("Failed to marshal VersionInfo to YAML: %v", err)
	}

	var yamlUnmarshaled VersionInfo
	if err := yaml.Unmarshal(yamlData, &yamlUnmarshaled); err != nil {
		t.Errorf("Failed to unmarshal VersionInfo from YAML: %v", err)
	}

	if yamlUnmarshaled != info {
		t.Errorf("YAML round-trip failed: got %+v, want %+v", yamlUnmarshaled, info)
	}
}
