package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewCmdAssetsSync(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsSync(f)

	// Test command metadata
	assert.Equal(t, "sync", cmd.Use)
	assert.Equal(t, "Synchronize assets with remote repository", cmd.Short)
	assert.Contains(t, cmd.Long, "Synchronize the local asset metadata")
	assert.Contains(t, cmd.Example, "zen assets sync")
}

func TestSyncCommandFlags(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsSync(f)

	// Test that expected flags exist
	forceFlag := cmd.Flags().Lookup("force")
	require.NotNil(t, forceFlag)
	assert.Equal(t, "bool", forceFlag.Value.Type())
	assert.Equal(t, "false", forceFlag.DefValue)

	branchFlag := cmd.Flags().Lookup("branch")
	require.NotNil(t, branchFlag)
	assert.Equal(t, "string", branchFlag.Value.Type())
	assert.Equal(t, "main", branchFlag.DefValue)

	timeoutFlag := cmd.Flags().Lookup("timeout")
	require.NotNil(t, timeoutFlag)
	assert.Equal(t, "int", timeoutFlag.Value.Type())
	assert.Equal(t, "60", timeoutFlag.DefValue)
}

func TestSyncSuccessfulTextOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testResult := &assets.SyncResult{
		Status:        "success",
		DurationMS:    3200,
		AssetsAdded:   2,
		AssetsUpdated: 5,
		AssetsRemoved: 0,
		CacheSizeMB:   15.2,
		LastSync:      time.Now(),
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockSyncAssetClient{
			result: testResult,
		}, nil
	}

	cmd := NewCmdAssetsSync(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Check success message
	assert.Contains(t, output, "Sync completed successfully")

	// Check statistics
	assert.Contains(t, output, "Added: 2 assets")
	assert.Contains(t, output, "Updated: 5 assets")
	assert.NotContains(t, output, "Removed") // Should not show 0 removals

	// Check summary
	assert.Contains(t, output, "Cache size: 15.2 MB")
	assert.Contains(t, output, "Duration: 3.2s")
}

func TestSyncWithErrorTextOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	stderr := io.ErrOut
	f := cmdutil.NewTestFactory(io)

	testResult := &assets.SyncResult{
		Status: "error",
		Error:  "Network connection failed",
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockSyncAssetClient{
			result: testResult,
		}, nil
	}

	cmd := NewCmdAssetsSync(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err := cmd.Execute()
	require.Error(t, err)

	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "Sync failed")
	assert.Contains(t, output, "Network connection failed")
}

func TestSyncPartialTextOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testResult := &assets.SyncResult{
		Status:      "partial",
		Error:       "Some assets could not be updated",
		DurationMS:  2500,
		CacheSizeMB: 12.8,
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockSyncAssetClient{
			result: testResult,
		}, nil
	}

	cmd := NewCmdAssetsSync(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "Sync completed with warnings")
	assert.Contains(t, output, "Some assets could not be updated")
}

func TestSyncJSONOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testResult := &assets.SyncResult{
		Status:        "success",
		DurationMS:    1500,
		AssetsAdded:   1,
		AssetsUpdated: 3,
		AssetsRemoved: 1,
		CacheSizeMB:   8.5,
		LastSync:      time.Now(),
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockSyncAssetClient{
			result: testResult,
		}, nil
	}

	// Create parent command to inherit persistent flags
	parentCmd := &cobra.Command{Use: "assets"}
	parentCmd.PersistentFlags().StringP("output", "o", "text", "Output format")

	cmd := NewCmdAssetsSync(f)
	parentCmd.AddCommand(cmd)
	parentCmd.SetArgs([]string{"sync", "--output", "json"})
	parentCmd.SetOut(stdout)

	err := parentCmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Parse JSON output
	var result assets.SyncResult
	err = json.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, int64(1500), result.DurationMS)
	assert.Equal(t, 1, result.AssetsAdded)
	assert.Equal(t, 3, result.AssetsUpdated)
	assert.Equal(t, 1, result.AssetsRemoved)
	assert.Equal(t, 8.5, result.CacheSizeMB)
}

func TestSyncYAMLOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testResult := &assets.SyncResult{
		Status:      "success",
		DurationMS:  1000,
		CacheSizeMB: 5.0,
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockSyncAssetClient{
			result: testResult,
		}, nil
	}

	// Create parent command to inherit persistent flags
	parentCmd := &cobra.Command{Use: "assets"}
	parentCmd.PersistentFlags().StringP("output", "o", "text", "Output format")

	cmd := NewCmdAssetsSync(f)
	parentCmd.AddCommand(cmd)
	parentCmd.SetArgs([]string{"sync", "--output", "yaml"})
	parentCmd.SetOut(stdout)

	err := parentCmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Parse YAML output
	var result assets.SyncResult
	err = yaml.Unmarshal([]byte(output), &result)
	require.NoError(t, err)

	assert.Equal(t, "success", result.Status)
	assert.Equal(t, int64(1000), result.DurationMS)
	assert.Equal(t, 5.0, result.CacheSizeMB)
}

func TestSyncWithFlags(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	mockClient := &mockSyncAssetClient{
		captureRequest: true,
		result: &assets.SyncResult{
			Status: "success",
		},
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return mockClient, nil
	}

	cmd := NewCmdAssetsSync(f)
	cmd.SetArgs([]string{
		"--force",
		"--branch", "develop",
		"--timeout", "600",
	})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	// The mock client should have captured the sync request
	assert.True(t, mockClient.lastRequest.Force)
	assert.Equal(t, "develop", mockClient.lastRequest.Branch)
}

func TestSyncAuthenticationError(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	stderr := io.ErrOut
	f := cmdutil.NewTestFactory(io)

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockSyncAssetClient{
			authError: true,
		}, nil
	}

	cmd := NewCmdAssetsSync(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err := cmd.Execute()
	require.Error(t, err)

	assert.Contains(t, err.Error(), "authentication failed")
	assert.Contains(t, err.Error(), "zen assets auth")
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{500 * time.Millisecond, "500ms"},
		{1500 * time.Millisecond, "1.5s"},
		{90 * time.Second, "1.5m"},
		{2*time.Hour + 30*time.Minute, "2.5h"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Mock asset client for sync testing
type mockSyncAssetClient struct {
	result         *assets.SyncResult
	captureRequest bool
	lastRequest    assets.SyncRequest
	authError      bool
}

func (m *mockSyncAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	return &assets.AssetList{}, nil
}

func (m *mockSyncAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	return &assets.AssetContent{}, nil
}

func (m *mockSyncAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	if m.authError {
		return nil, &assets.AssetClientError{
			Code:    assets.ErrorCodeAuthenticationFailed,
			Message: "authentication failed",
		}
	}

	if m.captureRequest {
		m.lastRequest = req
	}

	if m.result != nil {
		return m.result, nil
	}

	return &assets.SyncResult{Status: "success"}, nil
}

func (m *mockSyncAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	return &assets.CacheInfo{}, nil
}

func (m *mockSyncAssetClient) ClearCache(ctx context.Context) error {
	return nil
}

func (m *mockSyncAssetClient) Close() error {
	return nil
}
