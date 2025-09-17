package info

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCmdAssetsInfo(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsInfo(f)

	// Test command metadata
	assert.Equal(t, "info <asset-name>", cmd.Use)
	assert.Equal(t, "Show detailed information about an asset", cmd.Short)
	assert.Contains(t, cmd.Long, "Display detailed information")
	assert.Contains(t, cmd.Example, "zen assets info technical-spec")
}

func TestInfoCommandFlags(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsInfo(f)

	// Test that expected flags exist
	includeContentFlag := cmd.Flags().Lookup("include-content")
	require.NotNil(t, includeContentFlag)
	assert.Equal(t, "bool", includeContentFlag.Value.Type())
	assert.Equal(t, "false", includeContentFlag.DefValue)

	verifyFlag := cmd.Flags().Lookup("verify")
	require.NotNil(t, verifyFlag)
	assert.Equal(t, "bool", verifyFlag.Value.Type())
	assert.Equal(t, "true", verifyFlag.DefValue)
}

func TestInfoTextOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testAsset := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:        "technical-spec",
			Type:        assets.AssetTypeTemplate,
			Category:    "documentation",
			Description: "Comprehensive technical specification template",
			Tags:        []string{"documentation", "architecture", "planning"},
			Variables: []assets.Variable{
				{
					Name:        "FEATURE_NAME",
					Description: "Name of the feature",
					Required:    true,
					Type:        "string",
				},
				{
					Name:        "VERSION",
					Description: "Version number",
					Required:    false,
					Type:        "string",
					Default:     "1.0",
				},
			},
			Path:      "templates/technical-spec.md.template",
			UpdatedAt: time.Now().Add(-2 * time.Hour),
			Checksum:  "sha256:abcd1234567890",
		},
		Content:  "# Technical Specification - {{FEATURE_NAME}}\n\n**Version:** {{VERSION}}",
		Checksum: "sha256:abcd1234567890",
		Cached:   true,
		CacheAge: 7200, // 2 hours
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockInfoAssetClient{
			asset: testAsset,
		}, nil
	}

	cmd := NewCmdAssetsInfo(f)
	cmd.SetArgs([]string{"technical-spec"})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Check basic information
	assert.Contains(t, output, "technical-spec")
	assert.Contains(t, output, "template")
	assert.Contains(t, output, "documentation")
	assert.Contains(t, output, "Comprehensive technical specification template")

	// Check tags
	assert.Contains(t, output, "documentation, architecture, planning")

	// Check variables
	assert.Contains(t, output, "Template Variables")
	assert.Contains(t, output, "FEATURE_NAME")
	assert.Contains(t, output, "(required)")
	assert.Contains(t, output, "VERSION")
	assert.Contains(t, output, "(default: 1.0)")

	// Check file information
	assert.Contains(t, output, "templates/technical-spec.md.template")
	assert.Contains(t, output, "abcd1234567890...")

	// Check cache status
	assert.Contains(t, output, "Cached")
	assert.Contains(t, output, "2.0h")

	// Should not include content by default
	assert.NotContains(t, output, "# Technical Specification")

	// Should show content preview
	assert.Contains(t, output, "Content Preview")
}

func TestInfoWithContent(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testAsset := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:        "simple-template",
			Type:        assets.AssetTypeTemplate,
			Category:    "test",
			Description: "Simple test template",
		},
		Content: "Hello {{NAME}}!",
		Cached:  false,
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockInfoAssetClient{
			asset: testAsset,
		}, nil
	}

	cmd := NewCmdAssetsInfo(f)
	cmd.SetArgs([]string{"simple-template", "--include-content"})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Should include full content
	assert.Contains(t, output, "Content")
	assert.Contains(t, output, "Hello {{NAME}}!")

	// Should not show preview when full content is included
	assert.NotContains(t, output, "Content Preview")
}

func TestInfoJSONOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testAsset := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name:     "json-test",
			Type:     assets.AssetTypePrompt,
			Category: "test",
		},
		Content: "Test content",
		Cached:  true,
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockInfoAssetClient{
			asset: testAsset,
		}, nil
	}

	cmd := NewCmdAssetsInfo(f)
	cmd.SetArgs([]string{"json-test", "--output", "json"})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Parse JSON output
	var content assets.AssetContent
	err = json.Unmarshal([]byte(output), &content)
	require.NoError(t, err)

	assert.Equal(t, "json-test", content.Metadata.Name)
	assert.Equal(t, assets.AssetTypePrompt, content.Metadata.Type)
	assert.True(t, content.Cached)

	// Content should not be included by default
	assert.Empty(t, content.Content)
}

func TestInfoJSONWithContent(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testAsset := &assets.AssetContent{
		Metadata: assets.AssetMetadata{
			Name: "json-with-content",
			Type: assets.AssetTypePrompt,
		},
		Content: "Full content here",
		Cached:  true,
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockInfoAssetClient{
			asset: testAsset,
		}, nil
	}

	cmd := NewCmdAssetsInfo(f)
	cmd.SetArgs([]string{"json-with-content", "--output", "json", "--include-content"})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Parse JSON output
	var content assets.AssetContent
	err = json.Unmarshal([]byte(output), &content)
	require.NoError(t, err)

	// Content should be included
	assert.Equal(t, "Full content here", content.Content)
}

func TestInfoAssetNotFound(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	stderr := io.ErrOut
	f := cmdutil.NewTestFactory(io)

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockInfoAssetClient{
			notFoundError: true,
		}, nil
	}

	cmd := NewCmdAssetsInfo(f)
	cmd.SetArgs([]string{"nonexistent"})
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)

	err := cmd.Execute()
	require.Error(t, err)

	assert.Contains(t, err.Error(), "asset 'nonexistent' not found")
	assert.Contains(t, err.Error(), "zen assets list")
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes int
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatFileSize(tt.bytes)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		want     string
	}{
		{30 * time.Second, "30s"},
		{90 * time.Second, "2m"},
		{2*time.Hour + 30*time.Minute, "2.5h"},
		{25 * time.Hour, "1.0d"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.duration)
			assert.Equal(t, tt.want, got)
		})
	}
}

// Mock asset client for info testing
type mockInfoAssetClient struct {
	asset         *assets.AssetContent
	notFoundError bool
}

func (m *mockInfoAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	return &assets.AssetList{}, nil
}

func (m *mockInfoAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	if m.notFoundError {
		return nil, &assets.AssetClientError{
			Code:    assets.ErrorCodeAssetNotFound,
			Message: "asset not found",
		}
	}

	if m.asset != nil {
		return m.asset, nil
	}

	return &assets.AssetContent{}, nil
}

func (m *mockInfoAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	return &assets.SyncResult{Status: "success"}, nil
}

func (m *mockInfoAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	return &assets.CacheInfo{}, nil
}

func (m *mockInfoAssetClient) ClearCache(ctx context.Context) error {
	return nil
}

func (m *mockInfoAssetClient) Close() error {
	return nil
}
