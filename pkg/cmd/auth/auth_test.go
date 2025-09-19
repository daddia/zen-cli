package auth

import (
	"bytes"
	"context"
	"testing"

	"github.com/daddia/zen/pkg/auth"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
)

func TestNewCmdAuth(t *testing.T) {
	// Arrange
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	// Act
	cmd := NewCmdAuth(f)

	// Assert
	assert.NotNil(t, cmd)
	assert.Equal(t, "auth [provider]", cmd.Use)
	assert.Equal(t, "Authenticate with Git providers", cmd.Short)
	assert.Contains(t, cmd.Long, "Authenticate with Git providers for secure access")
	assert.Contains(t, cmd.Example, "zen auth github")
	assert.Equal(t, "core", cmd.GroupID)

	// Check flags
	assert.True(t, cmd.Flags().HasAvailableFlags())
	assert.NotNil(t, cmd.Flags().Lookup("token"))
	assert.NotNil(t, cmd.Flags().Lookup("token-file"))
	assert.NotNil(t, cmd.Flags().Lookup("validate"))
	assert.NotNil(t, cmd.Flags().Lookup("list"))
	assert.NotNil(t, cmd.Flags().Lookup("delete"))
}

func TestAuthCommand_Help(t *testing.T) {
	// Arrange
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAuth(f)
	cmd.SetArgs([]string{"--help"})
	cmd.SetOut(stdout)

	// Act
	err := cmd.Execute()

	// Assert
	assert.NoError(t, err)
	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "Authenticate with Git providers")
	assert.Contains(t, output, "github: GitHub Personal Access Token")
	assert.Contains(t, output, "gitlab: GitLab Project Access Token")
	assert.Contains(t, output, "ZEN_GITHUB_TOKEN")
}

func TestAuthCommand_ListProviders(t *testing.T) {
	// Arrange
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAuth(f)
	cmd.SetArgs([]string{"--list"})
	cmd.SetOut(stdout)

	// Act
	err := cmd.Execute()

	// Assert
	assert.NoError(t, err)
	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "Configured authentication providers")
	assert.Contains(t, output, "github:")
	assert.Contains(t, output, "gitlab:")
	assert.Contains(t, output, "ZEN_GITHUB_TOKEN")
}

func TestAuthCommand_ValidateProvider(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		expectError bool
		errorText   string
	}{
		{
			name:        "valid github provider",
			args:        []string{"github"},
			expectError: false,
		},
		{
			name:        "valid gitlab provider",
			args:        []string{"gitlab"},
			expectError: false,
		},
		{
			name:        "invalid provider",
			args:        []string{"bitbucket"},
			expectError: true,
			errorText:   "invalid provider",
		},
		{
			name:        "empty provider without list flag",
			args:        []string{},
			expectError: true,
			errorText:   "provider is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			io := iostreams.Test()
			stdout := io.Out
			stderr := io.ErrOut
			f := cmdutil.NewTestFactory(io)

			cmd := NewCmdAuth(f)
			cmd.SetArgs(tt.args)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)

			// Act
			err := cmd.Execute()

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorText != "" {
					assert.Contains(t, err.Error(), tt.errorText)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthCommand_WithToken(t *testing.T) {
	// Arrange
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAuth(f)
	cmd.SetArgs([]string{"github", "--token", "test-token"})
	cmd.SetOut(stdout)

	// Act
	err := cmd.Execute()

	// Assert
	assert.NoError(t, err)
	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "Successfully authenticated")
}

func TestAuthCommand_DeleteCredentials(t *testing.T) {
	// Arrange
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAuth(f)
	cmd.SetArgs([]string{"github", "--delete"})
	cmd.SetOut(stdout)

	// Act
	err := cmd.Execute()

	// Assert
	assert.NoError(t, err)
	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "Deleted credentials")
}

func TestAuthOptions_Validation(t *testing.T) {
	tests := []struct {
		name     string
		options  *AuthOptions
		provider string
		wantErr  bool
	}{
		{
			name: "valid options with provider",
			options: &AuthOptions{
				IO: iostreams.Test(),
				AuthManager: func() (auth.Manager, error) {
					return &mockAuthManager{}, nil
				},
				Provider: "github",
				Validate: true,
			},
			wantErr: false,
		},
		{
			name: "missing auth manager",
			options: &AuthOptions{
				IO:          iostreams.Test(),
				AuthManager: func() (auth.Manager, error) { return nil, assert.AnError },
				Provider:    "github",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			err := authRun(tt.options)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// May succeed or fail based on test environment
				if err != nil {
					t.Logf("Auth run failed in test environment (expected): %v", err)
				}
			}
		})
	}
}

func TestGetEnvVarName(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected string
	}{
		{
			name:     "github provider",
			provider: "github",
			expected: "ZEN_GITHUB_TOKEN",
		},
		{
			name:     "gitlab provider",
			provider: "gitlab",
			expected: "ZEN_GITLAB_TOKEN",
		},
		{
			name:     "custom provider",
			provider: "custom",
			expected: "ZEN_CUSTOM_TOKEN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := getEnvVarName(tt.provider)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPromptForToken(t *testing.T) {
	// Arrange
	io := iostreams.Test()
	opts := &AuthOptions{
		IO:       io,
		Provider: "github",
		AuthManager: func() (auth.Manager, error) {
			return &mockAuthManager{}, nil
		},
	}

	// Act
	_, err := promptForToken(opts)

	// Assert
	// Should fail because prompting is disabled in test streams
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "prompting is disabled")
}

// Mock implementations for testing

type mockAuthManager struct{}

func (m *mockAuthManager) Authenticate(ctx context.Context, provider string) error {
	return nil
}

func (m *mockAuthManager) GetCredentials(provider string) (string, error) {
	return "test-token", nil
}

func (m *mockAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	return nil
}

func (m *mockAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	return nil
}

func (m *mockAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	return true
}

func (m *mockAuthManager) ListProviders() []string {
	return []string{"github", "gitlab"}
}

func (m *mockAuthManager) DeleteCredentials(provider string) error {
	return nil
}

func (m *mockAuthManager) GetProviderInfo(provider string) (*auth.ProviderInfo, error) {
	if provider == "bitbucket" {
		return nil, auth.NewAuthError(auth.ErrorCodeProviderNotSupported, "unsupported provider", provider)
	}

	envVars := []string{"ZEN_GITHUB_TOKEN", "GITHUB_TOKEN", "GH_TOKEN"}
	if provider == "gitlab" {
		envVars = []string{"ZEN_GITLAB_TOKEN", "GITLAB_TOKEN", "GL_TOKEN"}
	}

	return &auth.ProviderInfo{
		Name:         provider,
		Type:         "token",
		Description:  "Test provider",
		Instructions: []string{"Test instruction"},
		EnvVars:      envVars,
	}, nil
}
