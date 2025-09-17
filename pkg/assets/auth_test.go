package assets

import (
	"context"
	"os"
	"testing"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenAuthProvider(t *testing.T) {
	logger := logging.NewBasic()
	auth := NewTokenAuthProvider(logger)

	assert.NotNil(t, auth)
	assert.NotNil(t, auth.logger)
	assert.NotNil(t, auth.tokens)
}

func TestTokenAuthProvider_GetCredentials_FromEnv(t *testing.T) {
	logger := logging.NewBasic()
	auth := NewTokenAuthProvider(logger)

	// Set test environment variable
	testToken := "test-token-123"
	os.Setenv("GITHUB_TOKEN", testToken)
	defer os.Unsetenv("GITHUB_TOKEN")

	// Execute
	token, err := auth.GetCredentials("github")

	// Verify
	require.NoError(t, err)
	assert.Equal(t, testToken, token)
}

func TestTokenAuthProvider_GetCredentials_NotFound(t *testing.T) {
	logger := logging.NewBasic()
	auth := NewTokenAuthProvider(logger)

	// Ensure no environment variables are set
	for _, envVar := range []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN"} {
		os.Unsetenv(envVar)
	}

	// Execute
	token, err := auth.GetCredentials("github")

	// Verify
	assert.Empty(t, token)
	assert.Error(t, err)

	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeAuthenticationFailed, assetErr.Code)
}

func TestTokenAuthProvider_GetEnvVarNames(t *testing.T) {
	logger := logging.NewBasic()
	auth := NewTokenAuthProvider(logger)

	tests := []struct {
		provider string
		expected []string
	}{
		{
			provider: "github",
			expected: []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN"},
		},
		{
			provider: "gitlab",
			expected: []string{"GITLAB_TOKEN", "GL_TOKEN", "ZEN_GITLAB_TOKEN"},
		},
		{
			provider: "custom",
			expected: []string{"ZEN_CUSTOM_TOKEN"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result := auth.getEnvVarNames(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenAuthProvider_GetTokenInstructions(t *testing.T) {
	logger := logging.NewBasic()
	auth := NewTokenAuthProvider(logger)

	tests := []struct {
		provider string
	}{
		{provider: "github"},
		{provider: "gitlab"},
		{provider: "custom"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			instructions := auth.getTokenInstructions(tt.provider)
			assert.NotNil(t, instructions)

			// Verify it's a map with expected fields
			if instructionMap, ok := instructions.(map[string]interface{}); ok {
				assert.Contains(t, instructionMap, "message")
				assert.Contains(t, instructionMap, "env_vars")
			}
		})
	}
}

func TestTokenAuthProvider_RefreshCredentials(t *testing.T) {
	logger := logging.NewBasic()
	auth := NewTokenAuthProvider(logger)
	ctx := context.Background()

	// Execute
	err := auth.RefreshCredentials(ctx, "github")

	// Verify
	assert.Error(t, err)

	var assetErr *AssetClientError
	assert.ErrorAs(t, err, &assetErr)
	assert.Equal(t, ErrorCodeAuthenticationFailed, assetErr.Code)
	assert.Contains(t, assetErr.Message, "refresh not supported")
}

// Integration test with environment variables
func TestTokenAuthProvider_Integration_GitHub(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Skip if no token is available
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("GITHUB_TOKEN not set, skipping integration test")
	}

	logger := logging.NewBasic()
	auth := NewTokenAuthProvider(logger)
	ctx := context.Background()

	// Execute authentication
	err := auth.Authenticate(ctx, "github")

	// Verify (this will make a real API call)
	if err != nil {
		// Check if it's a rate limiting error (acceptable in tests)
		var assetErr *AssetClientError
		if assert.ErrorAs(t, err, &assetErr) {
			if assetErr.Code == ErrorCodeRateLimited {
				t.Skip("GitHub API rate limited, skipping integration test")
			}
		}
		t.Errorf("Authentication failed: %v", err)
	}
}
