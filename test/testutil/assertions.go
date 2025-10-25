package testutil

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// AssertValidJSON checks if a string is valid JSON
func AssertValidJSON(t *testing.T, jsonStr string, msgAndArgs ...interface{}) {
	t.Helper()

	var js json.RawMessage
	err := json.Unmarshal([]byte(jsonStr), &js)
	require.NoError(t, err, msgAndArgs...)
}

// AssertValidYAML checks if a string is valid YAML
func AssertValidYAML(t *testing.T, yamlStr string, msgAndArgs ...interface{}) {
	t.Helper()

	var data interface{}
	err := yaml.Unmarshal([]byte(yamlStr), &data)
	require.NoError(t, err, msgAndArgs...)
}

// AssertContainsAll checks if text contains all required substrings
func AssertContainsAll(t *testing.T, text string, required []string, msgAndArgs ...interface{}) {
	t.Helper()

	for _, req := range required {
		assert.Contains(t, text, req, msgAndArgs...)
	}
}

// AssertNotContainsAny checks if text doesn't contain any forbidden substrings
func AssertNotContainsAny(t *testing.T, text string, forbidden []string, msgAndArgs ...interface{}) {
	t.Helper()

	for _, forb := range forbidden {
		assert.NotContains(t, text, forb, msgAndArgs...)
	}
}

// AssertExitCode checks command exit code with helpful error message
func AssertExitCode(t *testing.T, expected, actual int, stdout, stderr string, msgAndArgs ...interface{}) {
	t.Helper()

	if expected != actual {
		t.Errorf("Expected exit code %d, got %d\nStdout: %s\nStderr: %s",
			expected, actual, stdout, stderr)
		if len(msgAndArgs) > 0 {
			t.Errorf("Additional info: %v", msgAndArgs)
		}
	}
}

// AssertCommandSuccess checks that a command succeeded (exit code 0)
func AssertCommandSuccess(t *testing.T, exitCode int, stdout, stderr string, msgAndArgs ...interface{}) {
	t.Helper()

	if exitCode != 0 {
		t.Errorf("Command should have succeeded (exit code 0), got %d\nStdout: %s\nStderr: %s",
			exitCode, stdout, stderr)
		if len(msgAndArgs) > 0 {
			t.Errorf("Additional info: %v", msgAndArgs)
		}
	}
}

// AssertCommandFailure checks that a command failed (non-zero exit code)
func AssertCommandFailure(t *testing.T, exitCode int, stdout, stderr string, msgAndArgs ...interface{}) {
	t.Helper()

	if exitCode == 0 {
		t.Errorf("Command should have failed (non-zero exit code), got 0\nStdout: %s\nStderr: %s",
			stdout, stderr)
		if len(msgAndArgs) > 0 {
			t.Errorf("Additional info: %v", msgAndArgs)
		}
	}
}

// AssertZenDesignCompliance checks output follows Zen design guidelines
func AssertZenDesignCompliance(t *testing.T, output string) {
	t.Helper()

	// Check for proper symbols
	if strings.Contains(output, "success") || strings.Contains(output, "completed") {
		assert.Contains(t, output, "✓", "Success messages should use ✓ symbol")
	}

	if strings.Contains(output, "error") || strings.Contains(output, "failed") {
		assert.Contains(t, output, "✗", "Error messages should use ✗ symbol")
	}

	if strings.Contains(output, "warning") || strings.Contains(output, "alert") {
		assert.Contains(t, output, "!", "Warning messages should use ! symbol")
	}

	// Check no emojis are used
	emojiPatterns := []string{"😀", "😃", "😄", "😁", "😆", "🎉", "🚀", "💡", "⚠️", "❌", "✅"}
	AssertNotContainsAny(t, output, emojiPatterns, "Output should not contain emojis")
}
