package template

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"gopkg.in/yaml.v3"
)

// AssetLoader implements TemplateLoader using Asset Client
type AssetLoader struct {
	assetClient assets.AssetClientInterface
	logger      logging.Logger
}

// NewAssetLoader creates a new asset loader
func NewAssetLoader(assetClient assets.AssetClientInterface, logger logging.Logger) *AssetLoader {
	return &AssetLoader{
		assetClient: assetClient,
		logger:      logger,
	}
}

// LoadByName loads a template by name from Asset Client
func (l *AssetLoader) LoadByName(ctx context.Context, name string) (*assets.AssetContent, error) {
	l.logger.Debug("loading template by name", "name", name)

	opts := assets.GetAssetOptions{
		IncludeMetadata: true,
		VerifyIntegrity: true,
		UseCache:        true,
	}

	content, err := l.assetClient.GetAsset(ctx, name, opts)
	if err != nil {
		l.logger.Debug("failed to load template", "name", name, "error", err)
		return nil, fmt.Errorf("failed to load template '%s': %w", name, err)
	}

	l.logger.Debug("template loaded successfully", "name", name, "size", len(content.Content))
	return content, nil
}

// LoadByCategory loads templates by category
func (l *AssetLoader) LoadByCategory(ctx context.Context, category string) ([]*assets.AssetContent, error) {
	l.logger.Debug("loading templates by category", "category", category)

	filter := assets.AssetFilter{
		Type:     assets.AssetTypeTemplate,
		Category: category,
	}

	assetList, err := l.assetClient.ListAssets(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list templates by category '%s': %w", category, err)
	}

	var templates []*assets.AssetContent
	for _, asset := range assetList.Assets {
		opts := assets.GetAssetOptions{
			IncludeMetadata: true,
			VerifyIntegrity: true,
			UseCache:        true,
		}

		content, err := l.assetClient.GetAsset(ctx, asset.Name, opts)
		if err != nil {
			l.logger.Warn("failed to load template", "name", asset.Name, "error", err)
			continue
		}

		templates = append(templates, content)
	}

	l.logger.Debug("templates loaded by category", "category", category, "count", len(templates))
	return templates, nil
}

// ListAvailable lists available templates
func (l *AssetLoader) ListAvailable(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error) {
	l.logger.Debug("listing available templates", "filter", filter)

	// Ensure we're only looking for templates
	filter.Type = assets.AssetTypeTemplate

	assetList, err := l.assetClient.ListAssets(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list available templates: %w", err)
	}

	l.logger.Debug("templates listed successfully", "count", len(assetList.Assets), "total", assetList.Total)
	return assetList, nil
}

// GetMetadata extracts metadata from template content
func (l *AssetLoader) GetMetadata(ctx context.Context, content string) (*TemplateMetadata, error) {
	l.logger.Debug("extracting metadata from template content", "size", len(content))

	metadata := &TemplateMetadata{}

	// Look for YAML frontmatter
	if frontmatter := l.extractFrontmatter(content); frontmatter != "" {
		if err := yaml.Unmarshal([]byte(frontmatter), metadata); err != nil {
			l.logger.Debug("failed to parse YAML frontmatter", "error", err)
			// Continue with comment-based extraction
		} else {
			l.logger.Debug("metadata extracted from YAML frontmatter")
			return metadata, nil
		}
	}

	// Look for metadata in comments
	l.extractCommentMetadata(content, metadata)

	l.logger.Debug("metadata extraction completed", "name", metadata.Name)
	return metadata, nil
}

// extractFrontmatter extracts YAML frontmatter from template content
func (l *AssetLoader) extractFrontmatter(content string) string {
	// Look for YAML frontmatter pattern: ---\n...yaml...\n---
	frontmatterRegex := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	matches := frontmatterRegex.FindStringSubmatch(content)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

// extractCommentMetadata extracts metadata from template comments
func (l *AssetLoader) extractCommentMetadata(content string, metadata *TemplateMetadata) {
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip non-comment lines
		if !strings.HasPrefix(line, "#") && !strings.HasPrefix(line, "//") {
			continue
		}

		// Remove comment prefix
		line = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(line, "#"), "//"))

		// Parse metadata fields
		switch {
		case strings.HasPrefix(line, "@name:"):
			metadata.Name = strings.TrimSpace(strings.TrimPrefix(line, "@name:"))
		case strings.HasPrefix(line, "@description:"):
			metadata.Description = strings.TrimSpace(strings.TrimPrefix(line, "@description:"))
		case strings.HasPrefix(line, "@category:"):
			metadata.Category = strings.TrimSpace(strings.TrimPrefix(line, "@category:"))
		case strings.HasPrefix(line, "@version:"):
			metadata.Version = strings.TrimSpace(strings.TrimPrefix(line, "@version:"))
		case strings.HasPrefix(line, "@author:"):
			metadata.Author = strings.TrimSpace(strings.TrimPrefix(line, "@author:"))
		case strings.HasPrefix(line, "@tags:"):
			tagsStr := strings.TrimSpace(strings.TrimPrefix(line, "@tags:"))
			if tagsStr != "" {
				tags := strings.Split(tagsStr, ",")
				for i, tag := range tags {
					tags[i] = strings.TrimSpace(tag)
				}
				metadata.Tags = tags
			}
		case strings.HasPrefix(line, "@created:"):
			createdStr := strings.TrimSpace(strings.TrimPrefix(line, "@created:"))
			if created, err := time.Parse("2006-01-02", createdStr); err == nil {
				metadata.CreatedAt = created
			}
		case strings.HasPrefix(line, "@updated:"):
			updatedStr := strings.TrimSpace(strings.TrimPrefix(line, "@updated:"))
			if updated, err := time.Parse("2006-01-02", updatedStr); err == nil {
				metadata.UpdatedAt = updated
			}
		case strings.HasPrefix(line, "@variable:"):
			// Parse variable specification: @variable: name:type:required:description
			varStr := strings.TrimSpace(strings.TrimPrefix(line, "@variable:"))
			if varSpec := l.parseVariableSpec(varStr); varSpec != nil {
				metadata.Variables = append(metadata.Variables, *varSpec)
			}
		}
	}
}

// parseVariableSpec parses a variable specification string
func (l *AssetLoader) parseVariableSpec(spec string) *VariableSpec {
	// Format: name:type:required:description or name:type:required:default:description
	parts := strings.Split(spec, ":")
	if len(parts) < 3 {
		return nil
	}

	varSpec := &VariableSpec{
		Name: strings.TrimSpace(parts[0]),
		Type: strings.TrimSpace(parts[1]),
	}

	// Parse required flag
	if required := strings.TrimSpace(parts[2]); required == "true" || required == "required" {
		varSpec.Required = true
	}

	// Handle description (last part)
	if len(parts) >= 4 {
		if len(parts) == 4 {
			// name:type:required:description
			varSpec.Description = strings.TrimSpace(parts[3])
		} else if len(parts) >= 5 {
			// name:type:required:default:description
			varSpec.Default = strings.TrimSpace(parts[3])
			varSpec.Description = strings.TrimSpace(strings.Join(parts[4:], ":"))
		}
	}

	return varSpec
}

// MetadataExtractor provides utilities for extracting metadata from various sources
type MetadataExtractor struct {
	logger logging.Logger
}

// NewMetadataExtractor creates a new metadata extractor
func NewMetadataExtractor(logger logging.Logger) *MetadataExtractor {
	return &MetadataExtractor{
		logger: logger,
	}
}

// ExtractFromAsset extracts metadata from an asset
func (e *MetadataExtractor) ExtractFromAsset(asset *assets.AssetContent) *TemplateMetadata {
	metadata := &TemplateMetadata{
		Name:        asset.Metadata.Name,
		Description: asset.Metadata.Description,
		Category:    asset.Metadata.Category,
		Tags:        asset.Metadata.Tags,
		UpdatedAt:   asset.Metadata.UpdatedAt,
		Variables:   convertAssetVariables(asset.Metadata.Variables),
	}

	return metadata
}

// MergeMetadata merges multiple metadata sources with precedence
func (e *MetadataExtractor) MergeMetadata(base, override *TemplateMetadata) *TemplateMetadata {
	merged := *base

	// Override takes precedence for non-empty values
	if override.Name != "" {
		merged.Name = override.Name
	}
	if override.Description != "" {
		merged.Description = override.Description
	}
	if override.Category != "" {
		merged.Category = override.Category
	}
	if override.Version != "" {
		merged.Version = override.Version
	}
	if override.Author != "" {
		merged.Author = override.Author
	}
	if len(override.Tags) > 0 {
		merged.Tags = override.Tags
	}
	if !override.CreatedAt.IsZero() {
		merged.CreatedAt = override.CreatedAt
	}
	if !override.UpdatedAt.IsZero() {
		merged.UpdatedAt = override.UpdatedAt
	}
	if len(override.Variables) > 0 {
		merged.Variables = override.Variables
	}

	return &merged
}
