# Template Engine API

Comprehensive template compilation and rendering system with asset integration.

## Overview

The Template Engine API provides a powerful system for:
- Loading templates from the Asset Library
- Compiling Go templates with custom functions
- Variable validation and type checking
- Caching compiled templates for performance
- Rendering templates with context and error handling

## Core Interfaces

### TemplateEngine Interface

```go
type TemplateEngine interface {
    LoadTemplate(ctx context.Context, name string) (*Template, error)
    CompileTemplate(ctx context.Context, name, content string, metadata *TemplateMetadata) (*Template, error)
    RenderTemplate(ctx context.Context, tmpl *Template, variables map[string]interface{}) (string, error)
    ListTemplates(ctx context.Context, filter TemplateFilter) ([]*TemplateInfo, error)
    ValidateVariables(ctx context.Context, tmpl *Template, variables map[string]interface{}) error
}
```

### Template Structure

```go
type Template struct {
    Name       string                 // Template name/identifier
    Content    string                 // Raw template content
    Compiled   *template.Template     // Compiled Go template
    Metadata   *TemplateMetadata      // Template metadata
    Variables  []VariableSpec         // Variable specifications
    CompiledAt time.Time             // Compilation timestamp
    Checksum   string                // Content checksum
}

type TemplateMetadata struct {
    Title       string                 `yaml:"title"`
    Description string                 `yaml:"description"`
    Version     string                 `yaml:"version"`
    Author      string                 `yaml:"author"`
    Tags        []string              `yaml:"tags"`
    Variables   []VariableSpec        `yaml:"variables"`
    Outputs     []OutputSpec          `yaml:"outputs"`
}
```

## Engine Configuration

### Config Structure

```go
type Config struct {
    CacheEnabled  bool          `json:"cache_enabled" yaml:"cache_enabled"`
    CacheTTL      time.Duration `json:"cache_ttl" yaml:"cache_ttl"`
    CacheSize     int           `json:"cache_size" yaml:"cache_size"`
    StrictMode    bool          `json:"strict_mode" yaml:"strict_mode"`
    EnableAI      bool          `json:"enable_ai" yaml:"enable_ai"`
    DefaultDelims struct {
        Left  string `json:"left" yaml:"left"`
        Right string `json:"right" yaml:"right"`
    } `json:"default_delims" yaml:"default_delims"`
    WorkspaceRoot string `json:"workspace_root" yaml:"workspace_root"`
}
```

### Creating Engine

```go
// Create with default configuration
engine := template.NewEngine(logger, assetClient, template.DefaultConfig())

// Create with custom configuration
config := template.Config{
    CacheEnabled:  true,
    CacheTTL:      1 * time.Hour,
    CacheSize:     200,
    StrictMode:    true,
    EnableAI:      false,
    WorkspaceRoot: "/path/to/workspace",
}
config.DefaultDelims.Left = "{{"
config.DefaultDelims.Right = "}}"

engine := template.NewEngine(logger, assetClient, config)
```

## Template Loading

### Loading from Asset Library

```go
// Load template by name
tmpl, err := engine.LoadTemplate(ctx, "task/index.md")
if err != nil {
    return fmt.Errorf("failed to load template: %w", err)
}

// Template is automatically cached if caching is enabled
fmt.Printf("Loaded template: %s (compiled at %v)\n", tmpl.Name, tmpl.CompiledAt)
```

### Template Discovery

```go
// List all templates
filter := template.TemplateFilter{}
templates, err := engine.ListTemplates(ctx, filter)
if err != nil {
    return fmt.Errorf("failed to list templates: %w", err)
}

// Filter by type
filter = template.TemplateFilter{
    Type: "task",
    Tags: []string{"workflow"},
}
taskTemplates, err := engine.ListTemplates(ctx, filter)
if err != nil {
    return fmt.Errorf("failed to list task templates: %w", err)
}
```

## Template Compilation

### Manual Compilation

```go
// Compile template from content
content := `# Task: {{.title}}

Description: {{.description}}

Created: {{now | formatTime "2006-01-02 15:04:05"}}
`

metadata := &template.TemplateMetadata{
    Title:       "Task Template",
    Description: "Basic task template",
    Variables: []template.VariableSpec{
        {Name: "title", Type: "string", Required: true},
        {Name: "description", Type: "string", Required: false, Default: "No description"},
    },
}

tmpl, err := engine.CompileTemplate(ctx, "custom-task", content, metadata)
if err != nil {
    return fmt.Errorf("failed to compile template: %w", err)
}
```

### Template Metadata

Templates can include YAML frontmatter with metadata:

```yaml
---
title: "Task Index Template"
description: "Creates index.md for task directories"
version: "1.0.0"
author: "Zen Team"
tags: ["task", "workflow", "index"]
variables:
  - name: "task_id"
    type: "string"
    required: true
    description: "Unique task identifier"
  - name: "title"
    type: "string" 
    required: true
    description: "Task title"
  - name: "type"
    type: "string"
    required: false
    default: "story"
    allowed_values: ["story", "bug", "epic", "spike", "task"]
  - name: "priority"
    type: "string"
    required: false
    default: "P2"
    pattern: "^P[0-3]$"
outputs:
  - path: "index.md"
    description: "Task overview and status"
---
# {{.title}}

**Task ID**: {{.task_id}}
**Type**: {{.type}}
**Priority**: {{.priority}}

## Description

{{.description | default "No description provided"}}

## Status

- [ ] Align: Define success criteria
- [ ] Discover: Gather evidence  
- [ ] Prioritize: Rank by value
- [ ] Design: Specify solution
- [ ] Build: Implement
- [ ] Ship: Deploy
- [ ] Learn: Measure outcomes
```

## Template Rendering

### Basic Rendering

```go
// Prepare variables
variables := map[string]interface{}{
    "task_id":     "PROJ-123",
    "title":       "Implement user authentication",
    "type":        "story",
    "priority":    "P1",
    "description": "Add OAuth2 authentication with GitHub",
}

// Render template
output, err := engine.RenderTemplate(ctx, tmpl, variables)
if err != nil {
    return fmt.Errorf("failed to render template: %w", err)
}

fmt.Println(output)
```

### Variable Validation

```go
// Validate variables before rendering
err := engine.ValidateVariables(ctx, tmpl, variables)
if err != nil {
    var validationErr *template.ValidationError
    if errors.As(err, &validationErr) {
        fmt.Printf("Validation failed: %s\n", validationErr.Message)
        for _, fieldErr := range validationErr.FieldErrors {
            fmt.Printf("  %s: %s\n", fieldErr.Field, fieldErr.Message)
        }
        return err
    }
    return fmt.Errorf("validation error: %w", err)
}
```

## Variable Specifications

### Variable Types

```go
type VariableSpec struct {
    Name          string      `yaml:"name"`
    Type          string      `yaml:"type"`           // string, int, bool, array, object
    Required      bool        `yaml:"required"`
    Default       interface{} `yaml:"default"`
    Description   string      `yaml:"description"`
    AllowedValues []string    `yaml:"allowed_values"`
    Pattern       string      `yaml:"pattern"`        // Regex pattern
    MinLength     int         `yaml:"min_length"`
    MaxLength     int         `yaml:"max_length"`
    Min           float64     `yaml:"min"`            // For numeric types
    Max           float64     `yaml:"max"`            // For numeric types
}
```

### Variable Examples

```yaml
variables:
  # String with validation
  - name: "email"
    type: "string"
    required: true
    pattern: "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"
    description: "User email address"
    
  # Enum with allowed values
  - name: "environment"
    type: "string"
    required: true
    allowed_values: ["development", "staging", "production"]
    default: "development"
    
  # Numeric with range
  - name: "port"
    type: "int"
    required: false
    default: 8080
    min: 1024
    max: 65535
    
  # Array type
  - name: "tags"
    type: "array"
    required: false
    default: []
    description: "List of tags"
    
  # Object type
  - name: "config"
    type: "object"
    required: false
    default: {}
    description: "Configuration object"
```

## Custom Functions

### Built-in Functions

The template engine includes Zen-specific functions:

```go
// Date and time functions
{{now}}                           // Current timestamp
{{now | formatTime "2006-01-02"}} // Formatted date
{{.created | timeAgo}}            // "2 hours ago"

// String functions
{{.title | title}}                // Title case
{{.text | truncate 50}}           // Truncate to 50 chars
{{.name | slug}}                  // Convert to slug
{{.content | indent 4}}           // Indent by 4 spaces

// File functions
{{readFile "README.md"}}          // Read file content
{{fileExists "config.yaml"}}      // Check file existence
{{fileName "/path/to/file.txt"}}  // Extract filename

// Workspace functions
{{workspaceRoot}}                 // Get workspace root
{{relativePath "/abs/path"}}      // Convert to relative path
{{joinPath "dir" "file.txt"}}     // Join path components

// Formatting functions
{{.data | toJSON}}                // Convert to JSON
{{.data | toYAML}}                // Convert to YAML
{{.number | formatNumber}}        // Format number with commas

// Conditional functions
{{.value | default "fallback"}}   // Default value
{{if .condition}}...{{end}}       // Conditional rendering
{{range .items}}...{{end}}        // Loop over items
```

### Custom Function Registration

```go
// Register custom functions
functions := template.NewFunctionRegistry(logger, workspaceRoot)

err := functions.Register("customFunc", func(input string) string {
    return strings.ToUpper(input)
})
if err != nil {
    return fmt.Errorf("failed to register function: %w", err)
}

// Use in templates
// {{.text | customFunc}}
```

## Caching

### Cache Management

```go
// Cache is automatically managed by the engine
// Templates are cached after compilation

// Check cache status
cacheStats := engine.GetCacheStats()
fmt.Printf("Cache hits: %d, misses: %d, size: %d\n", 
    cacheStats.Hits, cacheStats.Misses, cacheStats.Size)

// Clear cache
engine.ClearCache()

// Disable caching for specific operations
config := template.Config{CacheEnabled: false}
engine := template.NewEngine(logger, assetClient, config)
```

### Cache Configuration

```go
type CacheConfig struct {
    Enabled bool          `json:"enabled"`
    TTL     time.Duration `json:"ttl"`
    Size    int           `json:"size"`
}

// Configure caching
config := template.Config{
    CacheEnabled: true,
    CacheTTL:     1 * time.Hour,    // Templates expire after 1 hour
    CacheSize:    500,              // Maximum 500 templates in cache
}
```

## Error Handling

### Template Engine Errors

```go
type TemplateEngineError struct {
    Code    ErrorCode
    Message string
    Details error
}

// Error codes
const (
    ErrorCodeTemplateNotFound     ErrorCode = "TEMPLATE_NOT_FOUND"
    ErrorCodeCompilationFailed    ErrorCode = "COMPILATION_FAILED"
    ErrorCodeRenderingFailed      ErrorCode = "RENDERING_FAILED"
    ErrorCodeValidationFailed     ErrorCode = "VALIDATION_FAILED"
    ErrorCodeAssetClientError     ErrorCode = "ASSET_CLIENT_ERROR"
    ErrorCodeConfigurationError   ErrorCode = "CONFIGURATION_ERROR"
)
```

### Error Handling Examples

```go
tmpl, err := engine.LoadTemplate(ctx, "nonexistent")
if err != nil {
    var engineErr *template.TemplateEngineError
    if errors.As(err, &engineErr) {
        switch engineErr.Code {
        case template.ErrorCodeTemplateNotFound:
            fmt.Printf("Template not found: %s\n", engineErr.Message)
        case template.ErrorCodeAssetClientError:
            fmt.Printf("Asset client error: %s\n", engineErr.Message)
        default:
            fmt.Printf("Template engine error: %s\n", engineErr.Message)
        }
        return err
    }
    return fmt.Errorf("unexpected error: %w", err)
}
```

## Advanced Usage

### Render Context

```go
type RenderContext struct {
    Variables     map[string]interface{}
    Functions     template.FuncMap
    Metadata      *TemplateMetadata
    WorkspaceRoot string
    Options       RenderOptions
}

// Access render context in templates
// {{.__context.WorkspaceRoot}}
// {{.__context.Metadata.Title}}
```

### Template Inheritance

```go
// Base template
baseContent := `
{{define "header"}}# {{.title}}{{end}}
{{define "footer"}}---
Generated by Zen CLI
{{end}}
`

// Child template
childContent := `
{{template "header" .}}

Content goes here...

{{template "footer" .}}
`
```

### Conditional Rendering

```go
// Template with conditions
content := `
{{if .debug}}
Debug mode enabled
{{end}}

{{if eq .environment "production"}}
Production configuration
{{else}}
Development configuration  
{{end}}

{{range .items}}
- {{.name}}: {{.value}}
{{end}}
`
```

## Testing

### Template Testing

```go
func TestTemplateRendering(t *testing.T) {
    // Create test engine
    logger := testutil.NewTestLogger()
    assetClient := testutil.NewMockAssetClient()
    engine := template.NewEngine(logger, assetClient, template.DefaultConfig())
    
    // Test template compilation
    content := "Hello {{.name}}!"
    tmpl, err := engine.CompileTemplate(ctx, "test", content, nil)
    assert.NoError(t, err)
    assert.NotNil(t, tmpl)
    
    // Test rendering
    variables := map[string]interface{}{"name": "World"}
    output, err := engine.RenderTemplate(ctx, tmpl, variables)
    assert.NoError(t, err)
    assert.Equal(t, "Hello World!", output)
}
```

### Mock Asset Client

```go
type MockAssetClient struct {
    templates map[string]*assets.AssetContent
}

func (m *MockAssetClient) GetAsset(ctx context.Context, name string) (*assets.AssetContent, error) {
    if content, exists := m.templates[name]; exists {
        return content, nil
    }
    return nil, assets.ErrAssetNotFound
}
```

## Best Practices

1. **Use metadata** - Define variables and outputs in template metadata
2. **Validate variables** - Always validate variables before rendering
3. **Handle errors gracefully** - Provide meaningful error messages
4. **Cache templates** - Enable caching for better performance
5. **Use custom functions** - Leverage built-in functions for common operations
6. **Test templates** - Write tests for template compilation and rendering
7. **Document templates** - Include clear descriptions and examples

## Migration Guide

### From Simple Templates

```go
// Old way (simple string replacement)
content := strings.ReplaceAll(template, "{{NAME}}", name)

// New way (template engine)
tmpl, err := engine.CompileTemplate(ctx, "simple", template, nil)
if err != nil {
    return err
}

variables := map[string]interface{}{"NAME": name}
content, err := engine.RenderTemplate(ctx, tmpl, variables)
if err != nil {
    return err
}
```

## See Also

- **[Asset Client API](asset-client.md)** - Asset library integration
- **[Template Examples](../templates/)** - Template examples and patterns
- **[Custom Functions](template-functions.md)** - Complete function reference
- **[User Guide](../user-guide/README.md#templates)** - Template usage guide
