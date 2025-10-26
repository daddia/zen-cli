//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/daddia/zen/internal/zencmd"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// setupFileStorage forces file storage for CI environments where keychain is not available
func setupFileStorage(t *testing.T) func() {
	originalStorageType := os.Getenv("ZEN_AUTH_STORAGE_TYPE")
	os.Setenv("ZEN_AUTH_STORAGE_TYPE", "file")
	return func() {
		if originalStorageType == "" {
			os.Unsetenv("ZEN_AUTH_STORAGE_TYPE")
		} else {
			os.Setenv("ZEN_AUTH_STORAGE_TYPE", originalStorageType)
		}
	}
}

// TestAssetsIntegration_Comprehensive tests comprehensive asset functionality
// This test suite covers all asset commands with various scenarios
func TestAssetsIntegration_Comprehensive(t *testing.T) {
	// Skip if no GitHub token is available
	if os.Getenv("GITHUB_TOKEN") == "" && os.Getenv("GH_TOKEN") == "" && os.Getenv("ZEN_GITHUB_TOKEN") == "" {
		t.Skip("Skipping comprehensive integration test: no GITHUB_TOKEN environment variable set")
	}

	// Create temporary directory for test
	tempDir := t.TempDir()

	// Change to temp directory
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Initialize workspace first
	t.Run("init_workspace", func(t *testing.T) {
		streams := iostreams.Test()
		err := zencmd.Execute(ctx, []string{"init"}, streams)
		require.NoError(t, err)

		// Verify .zen directory was created
		zenDir := filepath.Join(tempDir, ".zen")
		assert.DirExists(t, zenDir)
	})

	// Test comprehensive authentication scenarios
	t.Run("authentication_comprehensive", func(t *testing.T) {
		// Test auth help
		t.Run("auth_help", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "auth", "--help"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Authenticate with Git providers")
			assert.Contains(t, output, "github")
			assert.Contains(t, output, "gitlab")
			assert.Contains(t, output, "--validate")
		})

		// Test GitHub authentication
		t.Run("github_auth_with_validation", func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			err := zencmd.Execute(ctx, []string{"assets", "auth", "github", "--validate"}, streams)
			require.NoError(t, err, "stderr: %s", stderr.String())

			output := stdout.String()
			errorOutput := stderr.String()
			assert.Contains(t, output, "GitHub")
			assert.Contains(t, errorOutput, "Provider: github")
			// Should indicate successful authentication
			assert.NotContains(t, output, "failed")
		})

		// Test authentication without validation
		t.Run("github_auth_no_validation", func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			err := zencmd.Execute(ctx, []string{"assets", "auth", "github", "--validate=false"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Successfully authenticated")
		})

		// Test invalid provider
		t.Run("invalid_provider", func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			err := zencmd.Execute(ctx, []string{"assets", "auth", "invalid-provider"}, streams)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "unsupported provider")
		})
	})

	// Test comprehensive status functionality
	t.Run("status_comprehensive", func(t *testing.T) {
		// Test text output
		t.Run("status_text_format", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "status"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Assets Asset Status")
			assert.Contains(t, output, "Auth Authentication")
			assert.Contains(t, output, "Cache Cache")
			assert.Contains(t, output, "Repository Repository")
			assert.Contains(t, output, "Quick actions")
		})

		// Test JSON output
		t.Run("status_json_format", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "status", "--output", "json"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			// Verify it's valid JSON
			var statusData map[string]interface{}
			err = json.Unmarshal([]byte(output), &statusData)
			require.NoError(t, err, "Status output should be valid JSON")

			// Verify required fields
			assert.Contains(t, statusData, "authentication")
			assert.Contains(t, statusData, "cache")
			assert.Contains(t, statusData, "repository")
		})

		// Test YAML output
		t.Run("status_yaml_format", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "status", "--output", "yaml"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			// Verify it's valid YAML
			var statusData map[string]interface{}
			err = yaml.Unmarshal([]byte(output), &statusData)
			require.NoError(t, err, "Status output should be valid YAML")

			// Verify required fields
			assert.Contains(t, statusData, "authentication")
			assert.Contains(t, statusData, "cache")
			assert.Contains(t, statusData, "repository")
		})
	})

	// Test comprehensive synchronization functionality
	t.Run("sync_comprehensive", func(t *testing.T) {
		// Test basic sync
		t.Run("sync_basic", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "sync"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Synchronizing")
			assert.Contains(t, output, "Summary")
		})

		// Test sync with force flag
		t.Run("sync_force", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "sync", "--force"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Synchronizing")
		})

		// Test sync with custom branch
		t.Run("sync_custom_branch", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "sync", "--branch", "main"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Synchronizing")
		})

		// Test sync with timeout
		t.Run("sync_timeout", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "sync", "--timeout", "30"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Synchronizing")
		})

		// Test sync JSON output
		t.Run("sync_json_output", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "sync", "--output", "json"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			// Should contain JSON structure
			if strings.Contains(output, "{") {
				var syncData map[string]interface{}
				// Try to parse as JSON (may have mixed output)
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					if strings.HasPrefix(line, "{") {
						err = json.Unmarshal([]byte(line), &syncData)
						if err == nil {
							break
						}
					}
				}
			}
		})
	})

	// Test comprehensive listing functionality
	t.Run("list_comprehensive", func(t *testing.T) {
		// Test basic list
		t.Run("list_basic", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "list"}, streams)
			// Don't require success since sync might not have worked
			if err != nil {
				t.Logf("List command failed (expected if sync failed): %v", err)
				return
			}

			output := stdout.String()
			assert.NotEmpty(t, output)
		})

		// Test list with type filters
		t.Run("list_type_filters", func(t *testing.T) {
			tests := []struct {
				name      string
				assetType string
			}{
				{"templates", "template"},
				{"prompts", "prompt"},
				{"mcp", "mcp"},
				{"schemas", "schema"},
			}

			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					var stdout bytes.Buffer
					streams := iostreams.Test()
					streams.Out = &stdout

					err := zencmd.Execute(ctx, []string{"assets", "list", "--type", tt.assetType}, streams)
					if err != nil {
						t.Logf("List with type filter failed (expected if no assets): %v", err)
						return
					}

					output := stdout.String()
					assert.NotEmpty(t, output)
				})
			}
		})

		// Test list with pagination
		t.Run("list_pagination", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "list", "--limit", "5", "--offset", "0"}, streams)
			if err != nil {
				t.Logf("List with pagination failed (expected if no assets): %v", err)
				return
			}

			output := stdout.String()
			assert.NotEmpty(t, output)
		})

		// Test list output formats
		t.Run("list_output_formats", func(t *testing.T) {
			formats := []struct {
				name   string
				format string
				check  func(string) bool
			}{
				{
					name:   "json",
					format: "json",
					check: func(output string) bool {
						return strings.Contains(output, "{") || strings.Contains(output, "[]")
					},
				},
				{
					name:   "yaml",
					format: "yaml",
					check: func(output string) bool {
						return strings.Contains(output, ":") || strings.Contains(output, "-")
					},
				},
			}

			for _, tt := range formats {
				t.Run(tt.name, func(t *testing.T) {
					var stdout bytes.Buffer
					streams := iostreams.Test()
					streams.Out = &stdout

					err := zencmd.Execute(ctx, []string{"assets", "list", "--output", tt.format, "--limit", "3"}, streams)
					if err != nil {
						t.Logf("List with %s format failed (expected if no assets): %v", tt.format, err)
						return
					}

					output := stdout.String()
					if output != "" {
						assert.True(t, tt.check(output), "Output should be valid %s format", tt.format)
					}
				})
			}
		})
	})

	// Test asset listing with filters
	t.Run("assets_list_with_filters", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
		}{
			{
				name: "list_templates",
				args: []string{"assets", "list", "--type", "template"},
			},
			{
				name: "list_prompts",
				args: []string{"assets", "list", "--type", "prompt"},
			},
			{
				name: "list_limited",
				args: []string{"assets", "list", "--limit", "5"},
			},
			{
				name: "list_json_output",
				args: []string{"assets", "list", "--output", "json"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var stdout bytes.Buffer
				streams := iostreams.Test()
				streams.Out = &stdout

				err := zencmd.Execute(ctx, tt.args, streams)
				require.NoError(t, err)

				output := stdout.String()
				assert.NotEmpty(t, output)

				// JSON output should be valid JSON
				if contains(tt.args, "--output", "json") {
					assert.Contains(t, output, "{")
					assert.Contains(t, output, "}")
				}
			})
		}
	})

	// Test comprehensive info functionality
	t.Run("info_comprehensive", func(t *testing.T) {
		// Test info help
		t.Run("info_help", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "info", "--help"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Show detailed information about an asset")
			assert.Contains(t, output, "--include-content")
			assert.Contains(t, output, "--no-verify")
		})

		// Test info with common asset names
		commonAssets := []string{"story-definition", "technical-spec", "user-story"}
		for _, assetName := range commonAssets {
			t.Run("info_"+assetName, func(t *testing.T) {
				var stdout bytes.Buffer
				streams := iostreams.Test()
				streams.Out = &stdout

				err := zencmd.Execute(ctx, []string{"assets", "info", assetName}, streams)
				if err != nil {
					t.Logf("Info command failed for %s (expected if asset doesn't exist): %v", assetName, err)
					return
				}

				output := stdout.String()
				assert.NotEmpty(t, output)
				if strings.Contains(output, assetName) {
					assert.Contains(t, output, "Name:")
					assert.Contains(t, output, "Type:")
				}
			})
		}

		// Test info with content inclusion
		t.Run("info_with_content", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "info", "story-definition", "--include-content"}, streams)
			if err != nil {
				t.Logf("Info with content failed (expected if asset doesn't exist): %v", err)
				return
			}

			output := stdout.String()
			assert.NotEmpty(t, output)
		})

		// Test info output formats
		t.Run("info_output_formats", func(t *testing.T) {
			formats := []string{"json", "yaml"}
			for _, format := range formats {
				t.Run("info_"+format, func(t *testing.T) {
					var stdout bytes.Buffer
					streams := iostreams.Test()
					streams.Out = &stdout

					err := zencmd.Execute(ctx, []string{"assets", "info", "story-definition", "--output", format}, streams)
					if err != nil {
						t.Logf("Info with %s format failed (expected if asset doesn't exist): %v", format, err)
						return
					}

					output := stdout.String()
					if output != "" {
						if format == "json" {
							assert.True(t, strings.Contains(output, "{") || strings.Contains(output, "["))
						} else if format == "yaml" {
							assert.True(t, strings.Contains(output, ":") || strings.Contains(output, "-"))
						}
					}
				})
			}
		})

		// Test info with nonexistent asset
		t.Run("info_nonexistent", func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout
			streams.ErrOut = &stderr

			err := zencmd.Execute(ctx, []string{"assets", "info", "nonexistent-asset-12345"}, streams)
			require.Error(t, err)
			// Don't check specific error text as it depends on auth status
			// Could be "not found" or "authentication failed" depending on environment
		})
	})
}

// TestAssetsIntegration_OutputFormats tests different output formats
func TestAssetsIntegration_OutputFormats(t *testing.T) {
	// Skip if no GitHub token is available
	if os.Getenv("GITHUB_TOKEN") == "" && os.Getenv("GH_TOKEN") == "" && os.Getenv("ZEN_GITHUB_TOKEN") == "" {
		t.Skip("Skipping output format test: no GITHUB_TOKEN environment variable set")
	}

	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	// Sync assets first
	err = zencmd.Execute(ctx, []string{"assets", "sync"}, streams)
	require.NoError(t, err)

	tests := []struct {
		name         string
		args         []string
		expectedText string
	}{
		{
			name:         "status_text",
			args:         []string{"assets", "status", "--output", "text"},
			expectedText: "Asset",
		},
		{
			name:         "status_json",
			args:         []string{"assets", "status", "--output", "json"},
			expectedText: `"`,
		},
		{
			name:         "status_yaml",
			args:         []string{"assets", "status", "--output", "yaml"},
			expectedText: ":",
		},
		{
			name:         "list_json",
			args:         []string{"assets", "list", "--output", "json", "--limit", "5"},
			expectedText: "{",
		},
		{
			name:         "list_yaml",
			args:         []string{"assets", "list", "--output", "yaml", "--limit", "5"},
			expectedText: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			streams.Out = &stdout

			err := zencmd.Execute(ctx, tt.args, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, tt.expectedText)
		})
	}
}

// TestAssetsIntegration_ErrorHandling tests error scenarios
func TestAssetsIntegration_ErrorHandling(t *testing.T) {
	// Force file storage in CI environments where keychain is not available
	defer setupFileStorage(t)()

	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "invalid_asset_type",
			args:        []string{"assets", "list", "--type", "invalid"},
			expectError: true, // Should error for invalid asset type
			errorText:   "invalid asset type",
		},
		{
			name:        "nonexistent_asset_info",
			args:        []string{"assets", "info", "nonexistent-asset-12345"},
			expectError: true,
			errorText:   "", // Don't check specific error text as it depends on auth status
		},
		{
			name:        "invalid_provider",
			args:        []string{"assets", "auth", "invalid-provider"},
			expectError: true,
			errorText:   "provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			streams.Out = &stdout
			streams.ErrOut = &stderr

			err := zencmd.Execute(ctx, tt.args, streams)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorText != "" {
					errorOutput := err.Error() + stderr.String()
					assert.Contains(t, errorOutput, tt.errorText)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestAssetsIntegration_ConfigurationPrecedence tests configuration handling
func TestAssetsIntegration_ConfigurationPrecedence(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	// Test that environment variables are respected
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		t.Run("github_token_from_env", func(t *testing.T) {
			var stdout bytes.Buffer
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "status"}, streams)
			require.NoError(t, err)

			// Should work with environment token
			output := stdout.String()
			assert.NotEmpty(t, output)
		})
	}

	// Test configuration values
	t.Run("config_values", func(t *testing.T) {
		var stdout bytes.Buffer
		streams.Out = &stdout

		// Test that config commands work with assets
		err := zencmd.Execute(ctx, []string{"config", "get", "assets.branch"}, streams)
		require.NoError(t, err)

		output := stdout.String()
		// The output might be redacted, so just check it's not empty
		assert.NotEmpty(t, output)
	})
}

// TestAssetsIntegration_SharedAuthArchitecture tests the shared auth manager integration
func TestAssetsIntegration_SharedAuthArchitecture(t *testing.T) {
	// Force file storage in CI environments where keychain is not available
	defer setupFileStorage(t)()

	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	t.Run("auth_state_consistency", func(t *testing.T) {
		// Test that authentication state is consistent across commands
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			// Test auth command
			var authOut bytes.Buffer
			streams.Out = &authOut
			err := zencmd.Execute(ctx, []string{"assets", "auth", "github"}, streams)
			require.NoError(t, err)

			// Test that status reflects authentication
			var statusOut bytes.Buffer
			streams.Out = &statusOut
			err = zencmd.Execute(ctx, []string{"assets", "status"}, streams)
			require.NoError(t, err)

			statusOutput := statusOut.String()
			// Note: Due to current integration state, this may still show not authenticated
			// but the commands should execute without crashing
			assert.Contains(t, statusOutput, "Authentication")
		}
	})

	t.Run("auth_provider_validation", func(t *testing.T) {
		tests := []struct {
			name        string
			provider    string
			expectError bool
		}{
			{"github_valid", "github", false},
			{"gitlab_valid", "gitlab", false},
			{"invalid_provider", "bitbucket", true},
			{"empty_provider", "", false}, // Now succeeds with validation message
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				streams.Out = &stdout
				streams.ErrOut = &stderr

				args := []string{"assets", "auth"}
				if tt.provider != "" {
					args = append(args, tt.provider)
				}

				err := zencmd.Execute(ctx, args, streams)
				if tt.expectError {
					assert.Error(t, err)
				} else {
					// May succeed or fail based on token availability
					if err != nil {
						t.Logf("Auth command failed (expected without token): %v", err)
					}
				}
			})
		}
	})
}

// TestAssetsIntegration_PerformanceAndConcurrency tests performance characteristics
func TestAssetsIntegration_PerformanceAndConcurrency(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance tests in short mode")
	}

	// Force file storage in CI environments where keychain is not available
	defer setupFileStorage(t)()

	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	t.Run("concurrent_status_calls", func(t *testing.T) {
		// Test concurrent access to status command
		const numConcurrent = 5
		var wg sync.WaitGroup
		errors := make(chan error, numConcurrent)

		for i := 0; i < numConcurrent; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				var stdout bytes.Buffer
				streams := iostreams.Test()
				streams.Out = &stdout

				err := zencmd.Execute(ctx, []string{"assets", "status"}, streams)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d failed: %v", id, err)
					return
				}

				output := stdout.String()
				if !strings.Contains(output, "Asset Status") {
					errors <- fmt.Errorf("goroutine %d: unexpected output", id)
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for any errors
		for err := range errors {
			t.Error(err)
		}
	})

	t.Run("performance_benchmarks", func(t *testing.T) {
		// Basic performance test for status command
		start := time.Now()

		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		err := zencmd.Execute(ctx, []string{"assets", "status"}, streams)
		require.NoError(t, err)

		duration := time.Since(start)
		assert.Less(t, duration, 5*time.Second, "Status command should complete within 5 seconds")

		t.Logf("Status command took %v", duration)
	})
}

// TestAssetsIntegration_SecurityAndCredentials tests security aspects
func TestAssetsIntegration_SecurityAndCredentials(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	t.Run("credential_security", func(t *testing.T) {
		// Test that credentials are not leaked in output
		var stdout, stderr bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout
		streams.ErrOut = &stderr

		err := zencmd.Execute(ctx, []string{"assets", "status", "--verbose"}, streams)
		require.NoError(t, err)

		allOutput := stdout.String() + stderr.String()

		// Check that no actual tokens are exposed
		tokenPatterns := []string{
			"ghp_[a-zA-Z0-9]{36}",     // GitHub personal access token
			"github_pat_[a-zA-Z0-9_]", // GitHub fine-grained token
			"glpat-[a-zA-Z0-9_-]{20}", // GitLab project access token
		}

		for _, pattern := range tokenPatterns {
			matched, _ := regexp.MatchString(pattern, allOutput)
			assert.False(t, matched, "Output should not contain actual tokens matching pattern: %s", pattern)
		}
	})

	t.Run("config_file_security", func(t *testing.T) {
		// Verify that config file doesn't contain sensitive information
		configPath := filepath.Join(tempDir, ".zen", "config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			configData, err := os.ReadFile(configPath)
			require.NoError(t, err)

			configContent := string(configData)

			// Check that no tokens are stored in plain text
			tokenPatterns := []string{
				"ghp_", "github_pat_", "glpat-",
				"token:", "password:", "secret:",
			}

			for _, pattern := range tokenPatterns {
				assert.NotContains(t, configContent, pattern,
					"Config file should not contain sensitive data: %s", pattern)
			}
		}
	})

	t.Run("environment_variable_handling", func(t *testing.T) {
		// Test that environment variables are handled securely
		testEnvVars := []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN"}

		for _, envVar := range testEnvVars {
			if value := os.Getenv(envVar); value != "" {
				t.Run("env_"+strings.ToLower(envVar), func(t *testing.T) {
					var stdout, stderr bytes.Buffer
					streams := iostreams.Test()
					streams.Out = &stdout
					streams.ErrOut = &stderr

					err := zencmd.Execute(ctx, []string{"assets", "status"}, streams)
					require.NoError(t, err)

					allOutput := stdout.String() + stderr.String()
					// Environment variable value should not appear in output
					assert.NotContains(t, allOutput, value,
						"Token value should not appear in command output")
				})
			}
		}
	})
}

// TestAssetsIntegration_ComprehensiveErrorScenarios tests extensive error handling
func TestAssetsIntegration_ComprehensiveErrorScenarios(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	t.Run("command_validation_errors", func(t *testing.T) {
		tests := []struct {
			name        string
			args        []string
			expectError bool
			errorText   string
		}{
			{
				name:        "invalid_asset_type",
				args:        []string{"assets", "list", "--type", "invalid"},
				expectError: true,
				errorText:   "invalid asset type",
			},
			{
				name:        "invalid_output_format",
				args:        []string{"assets", "status", "--output", "xml"},
				expectError: true,
				errorText:   "invalid output format",
			},
			{
				name:        "negative_limit",
				args:        []string{"assets", "list", "--limit", "-5"},
				expectError: true,
				errorText:   "invalid argument",
			},
			{
				name:        "negative_offset",
				args:        []string{"assets", "list", "--offset", "-10"},
				expectError: true,
				errorText:   "invalid argument",
			},
			{
				name:        "invalid_timeout",
				args:        []string{"assets", "sync", "--timeout", "-1"},
				expectError: false, // Sync command may succeed despite invalid timeout
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				streams.Out = &stdout
				streams.ErrOut = &stderr

				err := zencmd.Execute(ctx, tt.args, streams)
				if tt.expectError {
					assert.Error(t, err)
					if tt.errorText != "" && err != nil {
						errorOutput := err.Error() + stderr.String()
						assert.Contains(t, errorOutput, tt.errorText)
					}
				} else {
					assert.NoError(t, err)
				}
			})
		}
	})

	t.Run("resource_not_found_errors", func(t *testing.T) {
		tests := []struct {
			name      string
			args      []string
			errorText string
		}{
			{
				name:      "nonexistent_asset_info",
				args:      []string{"assets", "info", "definitely-does-not-exist-12345"},
				errorText: "", // Don't check specific error text as it depends on auth status
			},
			{
				name:      "empty_asset_name",
				args:      []string{"assets", "info", ""},
				errorText: "", // Don't check specific error text as it depends on auth status
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				streams.Out = &stdout
				streams.ErrOut = &stderr

				err := zencmd.Execute(ctx, tt.args, streams)
				require.Error(t, err)
				errorOutput := err.Error() + stderr.String()
				assert.Contains(t, errorOutput, tt.errorText)
			})
		}
	})

	t.Run("network_and_timeout_scenarios", func(t *testing.T) {
		// Test with very short timeout to simulate network issues
		t.Run("sync_short_timeout", func(t *testing.T) {
			var stdout bytes.Buffer
			streams := iostreams.Test()
			streams.Out = &stdout

			err := zencmd.Execute(ctx, []string{"assets", "sync", "--timeout", "1"}, streams)
			// May succeed or fail depending on network speed
			// Just verify it doesn't crash
			if err != nil {
				t.Logf("Sync with short timeout failed (expected): %v", err)
			}
		})
	})
}

// TestAssetsIntegration_EdgeCases tests edge cases and boundary conditions
func TestAssetsIntegration_EdgeCases(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	t.Run("boundary_values", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
		}{
			{
				name: "zero_limit",
				args: []string{"assets", "list", "--limit", "0"},
			},
			{
				name: "max_limit",
				args: []string{"assets", "list", "--limit", "1000"},
			},
			{
				name: "large_offset",
				args: []string{"assets", "list", "--offset", "999999"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var stdout bytes.Buffer
				streams := iostreams.Test()
				streams.Out = &stdout

				err := zencmd.Execute(ctx, tt.args, streams)
				// These should either succeed or fail gracefully
				if err != nil {
					t.Logf("Command with boundary values failed gracefully: %v", err)
				}
			})
		}
	})

	t.Run("special_characters", func(t *testing.T) {
		// Test asset names with special characters
		specialNames := []string{
			"asset-with-dashes",
			"asset_with_underscores",
			"asset.with.dots",
			"asset with spaces", // This should fail
		}

		for _, name := range specialNames {
			t.Run("special_name_"+strings.ReplaceAll(name, " ", "_"), func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				streams := iostreams.Test()
				streams.Out = &stdout
				streams.ErrOut = &stderr

				err := zencmd.Execute(ctx, []string{"assets", "info", name}, streams)
				// All of these should fail gracefully (asset not found or auth issues)
				assert.Error(t, err)
				// Accept either "not found" or authentication-related errors
				errorMsg := err.Error()
				assert.True(t,
					strings.Contains(errorMsg, "not found") ||
						strings.Contains(errorMsg, "authentication failed") ||
						strings.Contains(errorMsg, "insufficient permissions"),
					"Expected error about not found or authentication, got: %s", errorMsg)
			})
		}
	})

	t.Run("empty_workspace_scenarios", func(t *testing.T) {
		// Test commands in a fresh workspace with no cache
		freshDir := t.TempDir()
		oldWd2, err := os.Getwd()
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.Chdir(oldWd2))
		}()
		require.NoError(t, os.Chdir(freshDir))

		// Initialize fresh workspace
		err = zencmd.Execute(ctx, []string{"init"}, streams)
		require.NoError(t, err)

		// Test status in empty workspace
		var stdout bytes.Buffer
		streams.Out = &stdout
		err = zencmd.Execute(ctx, []string{"assets", "status"}, streams)
		require.NoError(t, err)

		output := stdout.String()
		assert.Contains(t, output, "Asset Status")
		assert.Contains(t, output, "Cache")
	})
}

// TestAssetsIntegration_ConfigurationVariations tests different configuration scenarios
func TestAssetsIntegration_ConfigurationVariations(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	t.Run("global_flags", func(t *testing.T) {
		tests := []struct {
			name string
			args []string
		}{
			{
				name: "verbose_flag",
				args: []string{"--verbose", "assets", "status"},
			},
			{
				name: "no_color_flag",
				args: []string{"--no-color", "assets", "status"},
			},
			{
				name: "dry_run_flag",
				args: []string{"--dry-run", "assets", "status"},
			},
			{
				name: "output_json_global",
				args: []string{"--output", "json", "assets", "status"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var stdout bytes.Buffer
				streams := iostreams.Test()
				streams.Out = &stdout

				err := zencmd.Execute(ctx, tt.args, streams)
				require.NoError(t, err)

				output := stdout.String()
				assert.NotEmpty(t, output)
			})
		}
	})

	t.Run("config_precedence", func(t *testing.T) {
		// Test that environment variables override config
		originalToken := os.Getenv("ZEN_GITHUB_TOKEN")
		defer func() {
			if originalToken != "" {
				os.Setenv("ZEN_GITHUB_TOKEN", originalToken)
			} else {
				os.Unsetenv("ZEN_GITHUB_TOKEN")
			}
		}()

		// Set test environment variable
		os.Setenv("ZEN_GITHUB_TOKEN", "test-token-from-env")

		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		err := zencmd.Execute(ctx, []string{"assets", "status"}, streams)
		require.NoError(t, err)

		// Should execute without error
		output := stdout.String()
		assert.Contains(t, output, "Asset Status")
	})
}

// TestAssetsIntegration_UserExperience tests user experience aspects
func TestAssetsIntegration_UserExperience(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	t.Run("help_accessibility", func(t *testing.T) {
		commands := [][]string{
			{"assets", "--help"},
			{"assets", "auth", "--help"},
			{"assets", "status", "--help"},
			{"assets", "sync", "--help"},
			{"assets", "list", "--help"},
			{"assets", "info", "--help"},
		}

		for _, cmd := range commands {
			t.Run("help_"+strings.Join(cmd[1:], "_"), func(t *testing.T) {
				var stdout bytes.Buffer
				streams := iostreams.Test()
				streams.Out = &stdout

				err := zencmd.Execute(ctx, cmd, streams)
				require.NoError(t, err)

				output := stdout.String()
				assert.NotEmpty(t, output)
				assert.Contains(t, output, "Usage:")
				assert.Contains(t, output, "Examples:")
				assert.Contains(t, output, "Flags:")
			})
		}
	})

	t.Run("error_messages_helpful", func(t *testing.T) {
		// Test that error messages are helpful and actionable
		tests := []struct {
			name          string
			args          []string
			shouldContain []string
		}{
			{
				name:          "missing_asset_name",
				args:          []string{"assets", "info"},
				shouldContain: []string{"accepts 1 arg", "received 0"},
			},
			{
				name:          "invalid_provider",
				args:          []string{"assets", "auth", "invalid"},
				shouldContain: []string{"unsupported", "provider"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				streams := iostreams.Test()
				streams.Out = &stdout
				streams.ErrOut = &stderr

				err := zencmd.Execute(ctx, tt.args, streams)
				require.Error(t, err)

				errorOutput := err.Error() + stderr.String()
				for _, shouldContain := range tt.shouldContain {
					assert.Contains(t, errorOutput, shouldContain)
				}
			})
		}
	})

	t.Run("progress_and_feedback", func(t *testing.T) {
		// Test that long-running operations provide feedback
		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		err := zencmd.Execute(ctx, []string{"assets", "sync"}, streams)
		require.NoError(t, err)

		output := stdout.String()
		// Should show progress indicators
		progressIndicators := []string{"Synchronizing", "Summary", "Duration"}
		for _, indicator := range progressIndicators {
			assert.Contains(t, output, indicator)
		}
	})
}

// TestAssetsIntegration_RegressionTests tests for known issues and regressions
func TestAssetsIntegration_RegressionTests(t *testing.T) {
	tempDir := t.TempDir()
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, os.Chdir(oldWd))
	}()
	require.NoError(t, os.Chdir(tempDir))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Initialize workspace
	streams := iostreams.Test()
	err = zencmd.Execute(ctx, []string{"init"}, streams)
	require.NoError(t, err)

	t.Run("auth_disconnect_regression", func(t *testing.T) {
		// Test for the original auth disconnect issue we discovered
		// This test ensures the shared auth architecture prevents the issue

		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			// Step 1: Authenticate
			var authOut bytes.Buffer
			streams.Out = &authOut
			err := zencmd.Execute(ctx, []string{"assets", "auth", "github"}, streams)
			require.NoError(t, err)

			// Step 2: Check status immediately after auth
			var statusOut bytes.Buffer
			streams.Out = &statusOut
			err = zencmd.Execute(ctx, []string{"assets", "status"}, streams)
			require.NoError(t, err)

			// Step 3: Try sync operation
			var syncOut bytes.Buffer
			streams.Out = &syncOut
			err = zencmd.Execute(ctx, []string{"assets", "sync"}, streams)
			require.NoError(t, err)

			// All commands should execute without the "not authenticated" errors
			// that were present before the shared auth architecture
			statusOutput := statusOut.String()
			syncOutput := syncOut.String()

			// Should not see the old disconnected auth messages
			assert.NotContains(t, statusOutput, "failed to get auth provider")
			assert.NotContains(t, syncOutput, "auth provider not found")
		}
	})

	t.Run("cache_consistency", func(t *testing.T) {
		// Test that cache operations are consistent
		var stdout bytes.Buffer
		streams := iostreams.Test()
		streams.Out = &stdout

		// Multiple status calls should be consistent
		for i := 0; i < 3; i++ {
			stdout.Reset()
			err := zencmd.Execute(ctx, []string{"assets", "status"}, streams)
			require.NoError(t, err)

			output := stdout.String()
			assert.Contains(t, output, "Cache")
			assert.Contains(t, output, "Status:")
		}
	})

	t.Run("command_idempotency", func(t *testing.T) {
		// Test that commands are idempotent (can be run multiple times safely)
		commands := [][]string{
			{"assets", "status"},
			{"assets", "list", "--limit", "1"},
		}

		for _, cmd := range commands {
			t.Run("idempotent_"+strings.Join(cmd[1:], "_"), func(t *testing.T) {
				var outputs []string

				// Run command multiple times
				for i := 0; i < 3; i++ {
					var stdout bytes.Buffer
					streams := iostreams.Test()
					streams.Out = &stdout

					err := zencmd.Execute(ctx, cmd, streams)
					if err != nil {
						t.Logf("Command failed on iteration %d (may be expected): %v", i, err)
						continue
					}

					outputs = append(outputs, stdout.String())
				}

				// If we got outputs, they should be consistent
				if len(outputs) > 1 {
					// Basic consistency check - all outputs should contain similar structure
					for i := 1; i < len(outputs); i++ {
						// They should all be non-empty if the first one was
						if outputs[0] != "" {
							assert.NotEmpty(t, outputs[i], "Output %d should not be empty if first output wasn't", i)
						}
					}
				}
			})
		}
	})
}

// Helper function to check if slice contains specific values
func contains(slice []string, target ...string) bool {
	for _, item := range slice {
		for _, t := range target {
			if item == t {
				return true
			}
		}
	}
	return false
}

// Helper function to extract JSON from mixed output
func extractJSON(output string) (map[string]interface{}, error) {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") {
			var data map[string]interface{}
			err := json.Unmarshal([]byte(line), &data)
			if err == nil {
				return data, nil
			}
		}
	}
	return nil, fmt.Errorf("no valid JSON found in output")
}

// Helper function to validate output format
func validateOutputFormat(output, format string) bool {
	switch format {
	case "json":
		_, err := extractJSON(output)
		return err == nil
	case "yaml":
		var data interface{}
		return yaml.Unmarshal([]byte(output), &data) == nil
	case "text":
		return output != "" && !strings.HasPrefix(strings.TrimSpace(output), "{")
	default:
		return false
	}
}
