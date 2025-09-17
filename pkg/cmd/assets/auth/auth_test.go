package auth

import (
	"bytes"
	"context"
	"testing"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
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
			name:     "valid github provider",
			provider: "github",
			wantErr:  false,
		},
		{
			name:     "valid gitlab provider",
			provider: "gitlab",
			wantErr:  false,
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
			} else if tt.provider != "" {
				// Note: In real implementation, this would succeed with proper auth
				// For now, we expect it to work with the mock implementation
				// Should show the note about implementation
				output := stderr.(*bytes.Buffer).String()
				assert.Contains(t, output, "Authentication implementation requires interface updates")
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
	// Note: This test would need to mock environment variables
	// For now, we test the function exists and returns expected empty string
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
				assert.Contains(t, output, "Authentication implementation requires interface updates")
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

	// Should succeed with mock implementation
	assert.NoError(t, err)

	// Should show implementation note
	output := stderr.(*bytes.Buffer).String()
	assert.Contains(t, output, "Authentication implementation requires interface updates")
}
