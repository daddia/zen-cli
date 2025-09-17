package list

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/daddia/zen/pkg/assets"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestNewCmdAssetsList(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsList(f)

	// Test command metadata
	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List available assets", cmd.Short)
	assert.Contains(t, cmd.Long, "List available assets with optional filtering")
	assert.Contains(t, cmd.Example, "zen assets list")
}

func TestListCommandFlags(t *testing.T) {
	io := iostreams.Test()
	f := cmdutil.NewTestFactory(io)

	cmd := NewCmdAssetsList(f)

	// Test that expected flags exist
	typeFlag := cmd.Flags().Lookup("type")
	require.NotNil(t, typeFlag)
	assert.Equal(t, "string", typeFlag.Value.Type())

	categoryFlag := cmd.Flags().Lookup("category")
	require.NotNil(t, categoryFlag)
	assert.Equal(t, "string", categoryFlag.Value.Type())

	tagsFlag := cmd.Flags().Lookup("tags")
	require.NotNil(t, tagsFlag)
	assert.Equal(t, "stringSlice", tagsFlag.Value.Type())

	limitFlag := cmd.Flags().Lookup("limit")
	require.NotNil(t, limitFlag)
	assert.Equal(t, "int", limitFlag.Value.Type())
	assert.Equal(t, "50", limitFlag.DefValue)

	offsetFlag := cmd.Flags().Lookup("offset")
	require.NotNil(t, offsetFlag)
	assert.Equal(t, "int", offsetFlag.Value.Type())
	assert.Equal(t, "0", offsetFlag.DefValue)
}

func TestListTextOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	// Create test assets
	testAssets := []assets.AssetMetadata{
		{
			Name:        "technical-spec",
			Type:        assets.AssetTypeTemplate,
			Category:    "documentation",
			Description: "Technical specification template",
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "user-story",
			Type:        assets.AssetTypeTemplate,
			Category:    "planning",
			Description: "User story template with acceptance criteria",
			UpdatedAt:   time.Now(),
		},
		{
			Name:        "code-review",
			Type:        assets.AssetTypePrompt,
			Category:    "quality",
			Description: "Code review prompt for AI analysis",
			UpdatedAt:   time.Now(),
		},
	}

	// Override the asset client
	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockListAssetClient{
			assets: testAssets,
		}, nil
	}

	cmd := NewCmdAssetsList(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Check table headers
	assert.Contains(t, output, "NAME")
	assert.Contains(t, output, "TYPE")
	assert.Contains(t, output, "CATEGORY")
	assert.Contains(t, output, "DESCRIPTION")

	// Check asset data
	assert.Contains(t, output, "technical-spec")
	assert.Contains(t, output, "user-story")
	assert.Contains(t, output, "code-review")
	assert.Contains(t, output, "documentation")
	assert.Contains(t, output, "planning")
	assert.Contains(t, output, "quality")

	// Check summary
	assert.Contains(t, output, "Total: 3 assets")
}

func TestListJSONOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testAssets := []assets.AssetMetadata{
		{
			Name:        "test-template",
			Type:        assets.AssetTypeTemplate,
			Category:    "test",
			Description: "Test template",
		},
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockListAssetClient{
			assets: testAssets,
		}, nil
	}

	cmd := NewCmdAssetsList(f)
	cmd.SetArgs([]string{"--output", "json"})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Parse JSON output
	var assetList assets.AssetList
	err = json.Unmarshal([]byte(output), &assetList)
	require.NoError(t, err)

	assert.Len(t, assetList.Assets, 1)
	assert.Equal(t, "test-template", assetList.Assets[0].Name)
	assert.Equal(t, assets.AssetTypeTemplate, assetList.Assets[0].Type)
	assert.Equal(t, "test", assetList.Assets[0].Category)
}

func TestListYAMLOutput(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	testAssets := []assets.AssetMetadata{
		{
			Name:        "yaml-test",
			Type:        assets.AssetTypePrompt,
			Category:    "testing",
			Description: "YAML test asset",
		},
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockListAssetClient{
			assets: testAssets,
		}, nil
	}

	cmd := NewCmdAssetsList(f)
	cmd.SetArgs([]string{"--output", "yaml"})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()

	// Parse YAML output
	var assetList assets.AssetList
	err = yaml.Unmarshal([]byte(output), &assetList)
	require.NoError(t, err)

	assert.Len(t, assetList.Assets, 1)
	assert.Equal(t, "yaml-test", assetList.Assets[0].Name)
}

func TestListWithFilters(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockListAssetClient{
			captureFilter: true,
		}, nil
	}

	cmd := NewCmdAssetsList(f)
	cmd.SetArgs([]string{
		"--type", "template",
		"--category", "documentation",
		"--tags", "ai,technical",
		"--limit", "25",
		"--offset", "10",
	})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	// The mock client should have captured the filter
	client, _ := f.AssetClient()
	mockClient := client.(*mockListAssetClient)

	assert.Equal(t, assets.AssetTypeTemplate, mockClient.lastFilter.Type)
	assert.Equal(t, "documentation", mockClient.lastFilter.Category)
	assert.Equal(t, []string{"ai", "technical"}, mockClient.lastFilter.Tags)
	assert.Equal(t, 25, mockClient.lastFilter.Limit)
	assert.Equal(t, 10, mockClient.lastFilter.Offset)
}

func TestListEmptyResults(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockListAssetClient{
			assets: []assets.AssetMetadata{}, // Empty list
		}, nil
	}

	cmd := NewCmdAssetsList(f)
	cmd.SetArgs([]string{})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "No assets found")
}

func TestParseAssetType(t *testing.T) {
	tests := []struct {
		input   string
		want    assets.AssetType
		wantErr bool
	}{
		{"template", assets.AssetTypeTemplate, false},
		{"prompt", assets.AssetTypePrompt, false},
		{"mcp", assets.AssetTypeMCP, false},
		{"schema", assets.AssetTypeSchema, false},
		{"invalid", "", true},
		{"", "", true},
		{"TEMPLATE", "", true}, // Case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseAssetType(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestHasFilters(t *testing.T) {
	tests := []struct {
		name   string
		filter assets.AssetFilter
		want   bool
	}{
		{
			name:   "no filters",
			filter: assets.AssetFilter{},
			want:   false,
		},
		{
			name: "type filter",
			filter: assets.AssetFilter{
				Type: assets.AssetTypeTemplate,
			},
			want: true,
		},
		{
			name: "category filter",
			filter: assets.AssetFilter{
				Category: "documentation",
			},
			want: true,
		},
		{
			name: "tags filter",
			filter: assets.AssetFilter{
				Tags: []string{"ai"},
			},
			want: true,
		},
		{
			name: "pagination only",
			filter: assets.AssetFilter{
				Limit:  10,
				Offset: 5,
			},
			want: false, // Pagination doesn't count as filtering
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasFilters(tt.filter)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestListPagination(t *testing.T) {
	io := iostreams.Test()
	stdout := io.Out
	f := cmdutil.NewTestFactory(io)

	// Create more assets than the limit
	var testAssets []assets.AssetMetadata
	for i := 0; i < 75; i++ {
		testAssets = append(testAssets, assets.AssetMetadata{
			Name:        strings.Repeat("a", i+1),
			Type:        assets.AssetTypeTemplate,
			Category:    "test",
			Description: "Test asset",
		})
	}

	f.AssetClient = func() (assets.AssetClientInterface, error) {
		return &mockListAssetClient{
			assets:  testAssets[:50], // Return first 50
			total:   75,              // But total is 75
			hasMore: true,
		}, nil
	}

	cmd := NewCmdAssetsList(f)
	cmd.SetArgs([]string{"--limit", "50"})
	cmd.SetOut(stdout)

	err := cmd.Execute()
	require.NoError(t, err)

	output := stdout.(*bytes.Buffer).String()
	assert.Contains(t, output, "showing 50 of 75")
	assert.Contains(t, output, "--offset 50")
}

// Mock asset client for list testing
type mockListAssetClient struct {
	assets        []assets.AssetMetadata
	total         int
	hasMore       bool
	captureFilter bool
	lastFilter    assets.AssetFilter
}

func (m *mockListAssetClient) ListAssets(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	if m.captureFilter {
		m.lastFilter = filter
	}

	total := m.total
	if total == 0 {
		total = len(m.assets)
	}

	return &assets.AssetList{
		Assets:  m.assets,
		Total:   total,
		HasMore: m.hasMore,
	}, nil
}

func (m *mockListAssetClient) GetAsset(ctx context.Context, name string, opts assets.GetAssetOptions) (*assets.AssetContent, error) {
	return &assets.AssetContent{}, nil
}

func (m *mockListAssetClient) SyncRepository(ctx context.Context, req assets.SyncRequest) (*assets.SyncResult, error) {
	return &assets.SyncResult{Status: "success"}, nil
}

func (m *mockListAssetClient) GetCacheInfo(ctx context.Context) (*assets.CacheInfo, error) {
	return &assets.CacheInfo{}, nil
}

func (m *mockListAssetClient) ClearCache(ctx context.Context) error {
	return nil
}

func (m *mockListAssetClient) Close() error {
	return nil
}
