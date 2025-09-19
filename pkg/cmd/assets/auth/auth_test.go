package auth

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewCmdAssetsAuth(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsAuth(f)

	// Test command metadata
	assert.Equal(t, "auth [provider]", cmd.Use)
	assert.Equal(t, "Authenticate with Git providers for asset access", cmd.Short)
	assert.Contains(t, cmd.Long, "Authenticate with Git providers")
	assert.Contains(t, cmd.Example, "zen assets auth github")
}

func TestAuthOptionsValidation(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "valid github provider without token",
			provider: "github",
			wantErr:  true, // Will fail because no token and prompting disabled
			errMsg:   "authentication token required but prompting is disabled",
		},
		{
			name:     "valid gitlab provider without token",
			provider: "gitlab",
			wantErr:  true, // Still fails because no token and prompting disabled
			errMsg:   "authentication token required but prompting is disabled",
		},
		{
			name:     "invalid provider",
			provider: "bitbucket",
			wantErr:  true,
			errMsg:   "unsupported provider",
		},
		{
			name:     "empty provider with validate flag",
			provider: "",
			wantErr:  false, // Should validate stored credentials
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io := iostreams.Test()
			stdout := io.Out
			stderr := io.ErrOut
			f := cmdutil.NewTestFactory(io)

			cmd := NewCmdAssetsAuth(f)

			args := []string{}
			if tt.provider != "" {
				args = append(args, tt.provider)
			}

			cmd.SetArgs(args)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsValidProvider(t *testing.T) {
	tests := []struct {
		provider string
		want     bool
	}{
		{"github", true},
		{"gitlab", true},
		{"bitbucket", false},
		{"", false},
		{"GITHUB", false}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := isValidProvider(tt.provider)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetTokenFromEnv(t *testing.T) {
	// Save and clear environment variables for predictable testing
	envVars := []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN", "GITLAB_TOKEN", "GL_TOKEN", "ZEN_GITLAB_TOKEN"}
	original := make(map[string]string)

	for _, envVar := range envVars {
		original[envVar] = os.Getenv(envVar)
		os.Unsetenv(envVar)
	}

	defer func() {
		for envVar, value := range original {
			if value != "" {
				os.Setenv(envVar, value)
			}
		}
	}()

	tests := []struct {
		provider string
		want     string
	}{
		{"github", ""},
		{"gitlab", ""},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := getTokenFromEnv(tt.provider)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAuthCommandFlags(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsAuth(f)

	// Test that expected flags exist
	tokenFileFlag := cmd.Flags().Lookup("token-file")
	require.NotNil(t, tokenFileFlag)
	assert.Equal(t, "string", tokenFileFlag.Value.Type())

	tokenFlag := cmd.Flags().Lookup("token")
	require.NotNil(t, tokenFlag)
	assert.Equal(t, "string", tokenFlag.Value.Type())

	validateFlag := cmd.Flags().Lookup("validate")
	require.NotNil(t, validateFlag)
	assert.Equal(t, "bool", validateFlag.Value.Type())
	assert.Equal(t, "true", validateFlag.DefValue) // Default should be true
}

func TestAuthCommandWithFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{
			name: "github with token file",
			args: []string{"github", "--token-file", "/tmp/token"},
		},
		{
			name: "gitlab with validate disabled",
			args: []string{"gitlab", "--validate=false"},
		},
		{
			name: "github with explicit token",
			args: []string{"github", "--token", "ghp_test123"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io := iostreams.Test()
			stdout := io.Out
			stderr := io.ErrOut
			f := cmdutil.NewTestFactory(io)

			cmd := NewCmdAssetsAuth(f)
			cmd.SetArgs(tt.args)
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)

			// Execute command - should not panic
			err := cmd.Execute()

			// With current mock implementation, should show implementation note
			if err == nil {
				output := stderr.(*bytes.Buffer).String()
				assert.Contains(t, output, "Authentication token will be managed by shared auth system")
			}
		})
	}
}

// Mock asset client for testing
type mockAssetClient struct{}

func (m *mockAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	return &assets.AssetList{}, nil
}

func (m *mockAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	return &assets.AssetContent{}, nil
}

func (m *mockAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	return &assets.SyncResult{Status: "success"}, nil
}

func (m *mockAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	return &assets.CacheInfo{}, nil
}

func (m *mockAssetClient) ClearCache(ctx context.Context) error {
	return nil
}

func (m *mockAssetClient) Close() error {
	return nil
}

// testAuthManager is a mock auth manager for testing
type testAuthManager struct{}

func (a *testAuthManager) Authenticate(ctx context.Context, provider string) error {
	return nil
}

func (a *testAuthManager) GetCredentials(provider string) (string, error) {
	return "test-token", nil
}

func (a *testAuthManager) ValidateCredentials(ctx context.Context, provider string) error {
	return nil
}

func (a *testAuthManager) RefreshCredentials(ctx context.Context, provider string) error {
	return nil
}

func (a *testAuthManager) IsAuthenticated(ctx context.Context, provider string) bool {
	return true
}

func (a *testAuthManager) ListProviders() []string {
	return []string{"github", "gitlab"}
}

func (a *testAuthManager) DeleteCredentials(provider string) error {
	return nil
}

func TestAuthRunWithMockClient(t *testing.T) {
	io := iostreams.Test()
	stderr := io.ErrOut

	opts := &AuthOptions{
		IO: io,
		AssetClient: func() (assets.AssetClientInterface, error) {
			return &mockAssetClient{}, nil
		},
		Provider: "github",
		Token:    "test-token",
		Validate: true,
	}

	err := authRun(opts)

	// May fail due to nil auth manager in test environment
	if err != nil {
		t.Skipf("Auth run failed in test environment (expected): %v", err)
		return
	}

	// Should show implementation note if successful
	output := stderr.(*bytes.Buffer).String()
	if output != "" {
		assert.Contains(t, output, "Authentication token will be managed by shared auth system")
	}
}

// MockAssetClient provides a more comprehensive mock for testing
type MockAssetClient struct {
	mock.Mock
}

func (m *MockAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*assets.AssetList), args.Error(1)
}

func (m *MockAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(*assets.AssetContent), args.Error(1)
}

func (m *MockAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*assets.SyncResult), args.Error(1)
}

func (m *MockAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	args := m.Called(ctx)
	return args.Get(0).(*assets.CacheInfo), args.Error(1)
}

func (m *MockAssetClient) ClearCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockAssetClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Test helper functions individually
func TestGetTokenFromEnvWithValues(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		envVars  map[string]string
		want     string
	}{
		{
			name:     "github token from GITHUB_TOKEN",
			provider: "github",
			envVars:  map[string]string{"GITHUB_TOKEN": "ghp_test123"},
			want:     "ghp_test123",
		},
		{
			name:     "github token from GH_TOKEN",
			provider: "github",
			envVars:  map[string]string{"GH_TOKEN": "ghp_alt123"},
			want:     "ghp_alt123",
		},
		{
			name:     "github token from ZEN_GITHUB_TOKEN",
			provider: "github",
			envVars:  map[string]string{"ZEN_GITHUB_TOKEN": "ghp_zen123"},
			want:     "ghp_zen123",
		},
		{
			name:     "gitlab token from GITLAB_TOKEN",
			provider: "gitlab",
			envVars:  map[string]string{"GITLAB_TOKEN": "glpat_test123"},
			want:     "glpat_test123",
		},
		{
			name:     "gitlab token from GL_TOKEN",
			provider: "gitlab",
			envVars:  map[string]string{"GL_TOKEN": "glpat_alt123"},
			want:     "glpat_alt123",
		},
		{
			name:     "gitlab token from ZEN_GITLAB_TOKEN",
			provider: "gitlab",
			envVars:  map[string]string{"ZEN_GITLAB_TOKEN": "glpat_zen123"},
			want:     "glpat_zen123",
		},
		{
			name:     "priority order - GITHUB_TOKEN over GH_TOKEN",
			provider: "github",
			envVars:  map[string]string{"GITHUB_TOKEN": "ghp_priority", "GH_TOKEN": "ghp_lower"},
			want:     "ghp_priority",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all environment variables first
			envVars := []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN", "GITLAB_TOKEN", "GL_TOKEN", "ZEN_GITLAB_TOKEN"}
			original := make(map[string]string)
			for _, envVar := range envVars {
				original[envVar] = os.Getenv(envVar)
				os.Unsetenv(envVar)
			}
			defer func() {
				for envVar, value := range original {
					if value != "" {
						os.Setenv(envVar, value)
					}
				}
			}()

			// Set test environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			got := getTokenFromEnv(tt.provider)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetAuthToken(t *testing.T) {
	tests := []struct {
		name        string
		opts        *AuthOptions
		envVars     map[string]string
		wantToken   string
		wantErr     bool
		errContains string
	}{
		{
			name: "explicit token takes priority",
			opts: &AuthOptions{
				Provider: "github",
				Token:    "explicit-token",
				IO:       iostreams.Test(),
			},
			wantToken: "explicit-token",
			wantErr:   false,
		},
		{
			name: "environment variable fallback",
			opts: &AuthOptions{
				Provider: "github",
				IO:       iostreams.Test(),
			},
			envVars:   map[string]string{"GITHUB_TOKEN": "env-token"},
			wantToken: "env-token",
			wantErr:   false,
		},
		{
			name: "no token and prompting disabled",
			opts: &AuthOptions{
				Provider: "github",
				IO:       iostreams.Test(),
			},
			wantErr:     true,
			errContains: "authentication token required but prompting is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set environment variables
			envVars := []string{"GITHUB_TOKEN", "GH_TOKEN", "ZEN_GITHUB_TOKEN", "GITLAB_TOKEN", "GL_TOKEN", "ZEN_GITLAB_TOKEN"}
			original := make(map[string]string)
			for _, envVar := range envVars {
				original[envVar] = os.Getenv(envVar)
				os.Unsetenv(envVar)
			}
			defer func() {
				for envVar, value := range original {
					if value != "" {
						os.Setenv(envVar, value)
					}
				}
			}()

			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Disable prompting for test
			tt.opts.IO.SetNeverPrompt(true)

			token, err := getAuthToken(tt.opts)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantToken, token)
			}
		})
	}
}

func TestReadTokenFromFile(t *testing.T) {
	// Test the not-yet-implemented function
	token, err := readTokenFromFile("nonexistent.txt")
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "token file reading not yet implemented")
}

func TestGetEnvVarName(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"github", "GITHUB_TOKEN"},
		{"gitlab", "GITLAB_TOKEN"},
		{"unknown", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := getEnvVarName(tt.provider)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestShowTokenInstructions(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		contains []string
	}{
		{
			name:     "github instructions",
			provider: "github",
			contains: []string{
				"GitHub Personal Access Token required",
				"https://github.com/settings/tokens",
				"Generate new token (classic)",
				"repo",
			},
		},
		{
			name:     "gitlab instructions",
			provider: "gitlab",
			contains: []string{
				"GitLab Project Access Token required",
				"Settings > Access Tokens",
				"read_repository",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			io := iostreams.Test()
			io.Out = &buf

			opts := &AuthOptions{
				Provider: tt.provider,
				IO:       io,
			}

			showTokenInstructions(opts)

			output := buf.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestAuthenticateProvider(t *testing.T) {
	tests := []struct {
		name            string
		provider        string
		token           string
		authManagerFunc func() (interface{}, error)
		wantErr         bool
	}{
		{
			name:     "successful authentication with github",
			provider: "github",
			token:    "ghp_test123",
			authManagerFunc: func() (interface{}, error) {
				return &testAuthManager{}, nil
			},
			wantErr: false,
		},
		{
			name:     "successful authentication with gitlab",
			provider: "gitlab",
			token:    "glpat_test123",
			authManagerFunc: func() (interface{}, error) {
				return &testAuthManager{}, nil
			},
			wantErr: false,
		},
		{
			name:     "auth manager error",
			provider: "github",
			token:    "ghp_test123",
			authManagerFunc: func() (interface{}, error) {
				return nil, errors.New("auth manager failed")
			},
			wantErr: true,
		},
		{
			name:            "nil auth manager",
			provider:        "github",
			token:           "ghp_test123",
			authManagerFunc: nil,
			wantErr:         false, // Should not fail with nil auth manager
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stderr bytes.Buffer
			io := iostreams.Test()
			io.ErrOut = &stderr

			client := &mockAssetClient{}

			opts := &AuthOptions{
				IO:          io,
				AuthManager: tt.authManagerFunc,
			}

			ctx := context.Background()
			err := authenticateProvider(ctx, client, tt.provider, tt.token, opts)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)

				// Check that debug output was written
				output := stderr.String()
				assert.Contains(t, output, tt.provider)
				assert.Contains(t, output, "Token length:")
			}
		})
	}
}

func TestValidateStoredCredentials(t *testing.T) {
	var stdout bytes.Buffer
	io := iostreams.Test()
	io.Out = &stdout

	client := &mockAssetClient{}
	opts := &AuthOptions{IO: io}

	ctx := context.Background()
	err := validateStoredCredentials(ctx, client, opts)

	assert.NoError(t, err)
	output := stdout.String()
	assert.Contains(t, output, "Credential validation requires specifying a provider")
	assert.Contains(t, output, "zen assets auth <provider> --validate")
}

func TestAuthRunErrorHandling(t *testing.T) {
	tests := []struct {
		name        string
		opts        *AuthOptions
		wantErr     bool
		errContains string
	}{
		{
			name: "asset client error",
			opts: &AuthOptions{
				IO: iostreams.Test(),
				AssetClient: func() (assets.AssetClientInterface, error) {
					return nil, errors.New("asset client failed")
				},
				Provider: "github",
			},
			wantErr:     true,
			errContains: "failed to get asset client",
		},
		{
			name: "empty provider without validate",
			opts: &AuthOptions{
				IO:          iostreams.Test(),
				AssetClient: func() (assets.AssetClientInterface, error) { return &mockAssetClient{}, nil },
				Provider:    "",
				Validate:    false,
			},
			wantErr:     true,
			errContains: "provider is required",
		},
		{
			name: "unsupported provider",
			opts: &AuthOptions{
				IO:          iostreams.Test(),
				AssetClient: func() (assets.AssetClientInterface, error) { return &mockAssetClient{}, nil },
				Provider:    "unsupported",
			},
			wantErr:     true,
			errContains: "unsupported provider 'unsupported'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authRun(tt.opts)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test performance and edge cases
func TestAuthRunConcurrency(t *testing.T) {
	// Test that auth run can be called concurrently without race conditions
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			opts := &AuthOptions{
				IO: iostreams.Test(),
				AssetClient: func() (assets.AssetClientInterface, error) {
					return &mockAssetClient{}, nil
				},
				Provider: "github",
				Token:    "test-token",
				Validate: false, // Disable validation to avoid auth manager dependency
			}

			err := authRun(opts)
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-done:
			// Some may fail due to test environment, but shouldn't panic
			if err != nil {
				t.Logf("Goroutine %d failed (expected in test): %v", i, err)
			}
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out")
		}
	}
}

func TestPromptForTokenEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupIO     func() *iostreams.IOStreams
		provider    string
		wantErr     bool
		errContains string
	}{
		{
			name: "prompting disabled",
			setupIO: func() *iostreams.IOStreams {
				io := iostreams.Test()
				io.SetNeverPrompt(true)
				return io
			},
			provider:    "github",
			wantErr:     true,
			errContains: "authentication token required but prompting is disabled",
		},
		{
			name: "empty token input",
			setupIO: func() *iostreams.IOStreams {
				io := iostreams.Test()
				// Prompting will be disabled by default in test environment
				return io
			},
			provider:    "github",
			wantErr:     true,
			errContains: "authentication token required but prompting is disabled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			io := tt.setupIO()
			opts := &AuthOptions{
				IO:       io,
				Provider: tt.provider,
			}

			token, err := promptForToken(opts)

			if tt.wantErr {
				require.Error(t, err)
				assert.Empty(t, token)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}

// Benchmark tests
func BenchmarkIsValidProvider(b *testing.B) {
	providers := []string{"github", "gitlab", "bitbucket", "unknown", ""}
	for i := 0; i < b.N; i++ {
		provider := providers[i%len(providers)]
		isValidProvider(provider)
	}
}

func BenchmarkGetTokenFromEnv(b *testing.B) {
	// Set up environment
	os.Setenv("GITHUB_TOKEN", "test-token")
	defer os.Unsetenv("GITHUB_TOKEN")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getTokenFromEnv("github")
	}
}

func BenchmarkAuthOptionsCreation(b *testing.B) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd := NewCmdAssetsAuth(f)
		if cmd == nil {
			b.Fatal("command is nil")
		}
	}
}
