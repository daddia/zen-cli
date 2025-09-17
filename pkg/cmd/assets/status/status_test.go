package status

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

func TestNewCmdAssetsStatus(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsStatus(f)

	// Test command metadata
	assert.Equal(t, "status", cmd.Use)
	assert.Equal(t, "Show authentication and cache status", cmd.Short)
	assert.Contains(t, cmd.Long, "Display the current status")
	assert.Contains(t, cmd.Example, "zen assets status")
}

func TestStatusTextOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	// Override the asset client to return test data
	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockStatusAssetClient{
			cacheInfo: &assets.CacheInfo{
				TotalSize:     1024 * 1024 * 15, // 15 MB
				AssetCount:    42,
				LastSync:      time.Now().Add(-2 * time.Hour),
				CacheHitRatio: 0.85,
			},
		}, nil
	}

	cmd := NewCmdAssetsStatus(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Check for expected sections
	assert.Contains(t, output, "Asset Status")
	assert.Contains(t, output, "Authentication")
	assert.Contains(t, output, "Cache")
	assert.Contains(t, output, "Repository")

	// Check cache information
	assert.Contains(t, output, "15.0 MB")
	assert.Contains(t, output, "42 cached")
	assert.Contains(t, output, "85.0%")
}

func TestStatusJSONOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	// Override the asset client
	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockStatusAssetClient{
			cacheInfo: &assets.CacheInfo{
				TotalSize:     1024 * 1024 * 10,
				AssetCount:    25,
				LastSync:      time.Now().Add(-1 * time.Hour),
				CacheHitRatio: 0.75,
			},
		}, nil
	}

	// Create parent command to inherit persistent flags
	parentCmd := &cobra.Command{Use: "assets"}
	parentCmd.PersistentFlags().StringP("output", "o", "text", "Output format")

	cmd := NewCmdAssetsStatus(f)
	parentCmd.AddCommand(cmd)
	parentCmd.SetArgs([]string{"status", "--output", "json"})
	parentCmd.SetOut(stdout)

	err := parentCmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Parse JSON output
	var status StatusInfo
	err = json.Unmarshal([]byte(output), &status)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "healthy", status.Cache.Status)
	assert.Equal(t, 10.0, status.Cache.SizeMB)
	assert.Equal(t, 25, status.Cache.AssetCount)
	assert.Equal(t, 0.75, status.Cache.HitRatio)

	assert.Equal(t, "github", status.Authentication.Provider)
	assert.False(t, status.Authentication.Authenticated)
	assert.Equal(t, "unknown", status.Authentication.Status)
}

func TestStatusYAMLOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	// Override the asset client
	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockStatusAssetClient{
			cacheInfo: &assets.CacheInfo{
				TotalSize:     1024 * 1024 * 5,
				AssetCount:    10,
				LastSync:      time.Now().Add(-30 * time.Minute),
				CacheHitRatio: 0.90,
			},
		}, nil
	}

	// Create parent command to inherit persistent flags
	parentCmd := &cobra.Command{Use: "assets"}
	parentCmd.PersistentFlags().StringP("output", "o", "text", "Output format")

	cmd := NewCmdAssetsStatus(f)
	parentCmd.AddCommand(cmd)
	parentCmd.SetArgs([]string{"status", "--output", "yaml"})
	parentCmd.SetOut(stdout)

	err := parentCmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Parse YAML output
	var status StatusInfo
	err = yaml.Unmarshal([]byte(output), &status)
	require.NoError(t, err)

	// Verify structure
	assert.Equal(t, "healthy", status.Cache.Status)
	assert.Equal(t, 5.0, status.Cache.SizeMB)
	assert.Equal(t, 10, status.Cache.AssetCount)
	assert.Equal(t, 0.90, status.Cache.HitRatio)
}

func TestStatusWithCacheError(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	// Override the asset client to return error
	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockStatusAssetClient{
			cacheError: assert.AnError,
		}, nil
	}

	cmd := NewCmdAssetsStatus(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err) // Should not fail completely

	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "Unavailable")
}

func TestFormatTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "zero time",
			time: time.Time{},
			want: "never",
		},
		{
			name: "just now",
			time: now.Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "minutes ago",
			time: now.Add(-5 * time.Minute),
			want: "5 minutes ago",
		},
		{
			name: "one minute ago",
			time: now.Add(-1 * time.Minute),
			want: "1 minute ago",
		},
		{
			name: "hours ago",
			time: now.Add(-3 * time.Hour),
			want: "3 hours ago",
		},
		{
			name: "one hour ago",
			time: now.Add(-1 * time.Hour),
			want: "1 hour ago",
		},
		{
			name: "days ago",
			time: now.Add(-2 * 24 * time.Hour),
			want: "2 days ago",
		},
		{
			name: "one day ago",
			time: now.Add(-1 * 24 * time.Hour),
			want: "1 day ago",
		},
		{
			name: "weeks ago",
			time: now.Add(-10 * 24 * time.Hour),
			want: "2006-01-02 15:04", // Should use absolute format
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTime(tt.time)
			if tt.name == "weeks ago" {
				// For absolute format, just check it contains expected pattern
				assert.Regexp(t, `\d{4}-\d{2}-\d{2} \d{2}:\d{2}`, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// Mock asset client for status testing
type mockStatusAssetClient struct {
	cacheInfo  *assets.CacheInfo
	cacheError error
}

func (m *mockStatusAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	return &assets.AssetList{}, nil
}

func (m *mockStatusAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	return &assets.AssetContent{}, nil
}

func (m *mockStatusAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	return &assets.SyncResult{Status: "success"}, nil
}

func (m *mockStatusAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	if m.cacheError != nil {
		return nil, m.cacheError
	}
	if m.cacheInfo != nil {
		return m.cacheInfo, nil
	}
	return &assets.CacheInfo{}, nil
}

func (m *mockStatusAssetClient) ClearCache(ctx context.Context) error {
	return nil
}

func (m *mockStatusAssetClient) Close() error {
	return nil
}
