package assets

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAssetClientError_Error(t *testing.T) {
	err := AssetClientError{
		Code:    ErrorCodeAssetNotFound,
		Message: "asset not found",
	}

	assert.Equal(t, "asset not found", err.Error())
}

func TestDefaultAssetConfig(t *testing.T) {
	config := DefaultAssetConfig()

	assert.Equal(t, "main", config.Branch)
	assert.Equal(t, "~/.zen/cache/assets", config.CachePath)
	assert.Equal(t, int64(100), config.CacheSizeMB)
	assert.Equal(t, 24*time.Hour, config.DefaultTTL)
	assert.Equal(t, "github", config.AuthProvider)
	assert.Equal(t, 30, config.SyncTimeoutSeconds)
	assert.Equal(t, 3, config.MaxConcurrentOps)
	assert.True(t, config.IntegrityChecksEnabled)
	assert.True(t, config.PrefetchEnabled)
}

func TestAssetType_Constants(t *testing.T) {
	assert.Equal(t, AssetType("template"), AssetTypeTemplate)
	assert.Equal(t, AssetType("prompt"), AssetTypePrompt)
	assert.Equal(t, AssetType("mcp"), AssetTypeMCP)
	assert.Equal(t, AssetType("schema"), AssetTypeSchema)
}

func TestAssetErrorCode_Constants(t *testing.T) {
	assert.Equal(t, AssetErrorCode("asset_not_found"), ErrorCodeAssetNotFound)
	assert.Equal(t, AssetErrorCode("authentication_failed"), ErrorCodeAuthenticationFailed)
	assert.Equal(t, AssetErrorCode("network_error"), ErrorCodeNetworkError)
	assert.Equal(t, AssetErrorCode("cache_error"), ErrorCodeCacheError)
	assert.Equal(t, AssetErrorCode("integrity_error"), ErrorCodeIntegrityError)
	assert.Equal(t, AssetErrorCode("rate_limited"), ErrorCodeRateLimited)
	assert.Equal(t, AssetErrorCode("repository_error"), ErrorCodeRepositoryError)
	assert.Equal(t, AssetErrorCode("configuration_error"), ErrorCodeConfigurationError)
}

func TestAssetMetadata_Structure(t *testing.T) {
	metadata := AssetMetadata{
		Name:        "test-asset",
		Type:        AssetTypeTemplate,
		Description: "Test asset",
		Format:      "markdown",
		Category:    "test",
		Tags:        []string{"test", "example"},
		Variables: []Variable{
			{
				Name:        "TEST_VAR",
				Description: "Test variable",
				Required:    true,
				Type:        "string",
			},
		},
		Checksum:  "sha256:test123",
		Path:      "templates/test.md.template",
		UpdatedAt: time.Now(),
	}

	assert.Equal(t, "test-asset", metadata.Name)
	assert.Equal(t, AssetTypeTemplate, metadata.Type)
	assert.Equal(t, "Test asset", metadata.Description)
	assert.Equal(t, "markdown", metadata.Format)
	assert.Equal(t, "test", metadata.Category)
	assert.Len(t, metadata.Tags, 2)
	assert.Len(t, metadata.Variables, 1)
	assert.Equal(t, "TEST_VAR", metadata.Variables[0].Name)
	assert.True(t, metadata.Variables[0].Required)
}

func TestAssetFilter_Structure(t *testing.T) {
	filter := AssetFilter{
		Type:     AssetTypeTemplate,
		Category: "planning",
		Tags:     []string{"strategy", "alignment"},
		Limit:    50,
		Offset:   0,
	}

	assert.Equal(t, AssetTypeTemplate, filter.Type)
	assert.Equal(t, "planning", filter.Category)
	assert.Len(t, filter.Tags, 2)
	assert.Equal(t, 50, filter.Limit)
	assert.Equal(t, 0, filter.Offset)
}

func TestSyncRequest_Structure(t *testing.T) {
	req := SyncRequest{
		Force:   false,
		Shallow: true,
		Branch:  "main",
	}

	assert.False(t, req.Force)
	assert.True(t, req.Shallow)
	assert.Equal(t, "main", req.Branch)
}

func TestGetAssetOptions_Structure(t *testing.T) {
	opts := GetAssetOptions{
		IncludeMetadata: true,
		VerifyIntegrity: true,
		UseCache:        true,
	}

	assert.True(t, opts.IncludeMetadata)
	assert.True(t, opts.VerifyIntegrity)
	assert.True(t, opts.UseCache)
}
