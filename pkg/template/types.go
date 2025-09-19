package template

import (
	"context"
	"text/template"
	"time"

	"github.com/daddia/zen/pkg/assets"
)

// TemplateEngine defines the interface for template operations
type TemplateEngine interface {
	// LoadTemplate loads a template by name from Asset Client
	LoadTemplate(ctx context.Context, name string) (*Template, error)

	// RenderTemplate renders a template with provided variables
	RenderTemplate(ctx context.Context, tmpl *Template, variables map[string]interface{}) (string, error)

	// ListTemplates lists available templates matching the filter
	ListTemplates(ctx context.Context, filter TemplateFilter) (*TemplateList, error)

	// ValidateVariables validates template variables against template requirements
	ValidateVariables(ctx context.Context, tmpl *Template, variables map[string]interface{}) error

	// CompileTemplate compiles a template string into a Template
	CompileTemplate(ctx context.Context, name, content string, metadata *TemplateMetadata) (*Template, error)

	// GetFunctions returns the available template functions
	GetFunctions() template.FuncMap
}

// Template represents a compiled template with metadata
type Template struct {
	Name       string             `json:"name"`
	Content    string             `json:"content"`
	Compiled   *template.Template `json:"-"`
	Metadata   *TemplateMetadata  `json:"metadata"`
	Variables  []VariableSpec     `json:"variables"`
	CompiledAt time.Time          `json:"compiled_at"`
	Checksum   string             `json:"checksum"`
}

// TemplateMetadata contains template metadata
type TemplateMetadata struct {
	Name        string         `json:"name" yaml:"name"`
	Description string         `json:"description" yaml:"description"`
	Category    string         `json:"category" yaml:"category"`
	Tags        []string       `json:"tags" yaml:"tags"`
	Version     string         `json:"version" yaml:"version"`
	Author      string         `json:"author,omitempty" yaml:"author,omitempty"`
	CreatedAt   time.Time      `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" yaml:"updated_at"`
	Variables   []VariableSpec `json:"variables" yaml:"variables"`
	Extends     string         `json:"extends,omitempty" yaml:"extends,omitempty"`
	Includes    []string       `json:"includes,omitempty" yaml:"includes,omitempty"`
}

// VariableSpec defines a template variable specification
type VariableSpec struct {
	Name        string      `json:"name" yaml:"name"`
	Description string      `json:"description" yaml:"description"`
	Type        string      `json:"type" yaml:"type"`
	Required    bool        `json:"required" yaml:"required"`
	Default     interface{} `json:"default,omitempty" yaml:"default,omitempty"`
	Validation  string      `json:"validation,omitempty" yaml:"validation,omitempty"`
	Examples    []string    `json:"examples,omitempty" yaml:"examples,omitempty"`
}

// RenderContext provides context for template rendering
type RenderContext struct {
	Variables     map[string]interface{} `json:"variables"`
	Functions     template.FuncMap       `json:"-"`
	Metadata      *TemplateMetadata      `json:"metadata"`
	Options       RenderOptions          `json:"options"`
	WorkspaceRoot string                 `json:"workspace_root"`
	TaskID        string                 `json:"task_id,omitempty"`
}

// RenderOptions configures template rendering behavior
type RenderOptions struct {
	StrictVariables bool   `json:"strict_variables"`
	MissingKey      string `json:"missing_key"` // "error", "zero", "invalid"
	Delims          struct {
		Left  string `json:"left"`
		Right string `json:"right"`
	} `json:"delims"`
	EnableAI       bool `json:"enable_ai"`
	ValidateOutput bool `json:"validate_output"`
}

// TemplateFilter represents filtering options for template queries
type TemplateFilter struct {
	Category string   `json:"category,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	Type     string   `json:"type,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
}

// TemplateList represents a paginated list of templates
type TemplateList struct {
	Templates []TemplateMetadata `json:"templates"`
	Total     int                `json:"total"`
	HasMore   bool               `json:"has_more"`
	Filter    TemplateFilter     `json:"filter"`
}

// ValidationResult represents the result of variable validation
type ValidationResult struct {
	Valid    bool                `json:"valid"`
	Errors   []ValidationError   `json:"errors,omitempty"`
	Warnings []ValidationWarning `json:"warnings,omitempty"`
	Missing  []string            `json:"missing,omitempty"`
	Extra    []string            `json:"extra,omitempty"`
}

// ValidationError represents a validation error
type ValidationError struct {
	Variable string `json:"variable"`
	Message  string `json:"message"`
	Code     string `json:"code"`
	Value    string `json:"value,omitempty"`
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Variable string `json:"variable"`
	Message  string `json:"message"`
	Code     string `json:"code"`
	Value    string `json:"value,omitempty"`
}

// TemplateEngineError represents template engine errors
type TemplateEngineError struct {
	Code    TemplateErrorCode `json:"code"`
	Message string            `json:"message"`
	Details interface{}       `json:"details,omitempty"`
}

func (e *TemplateEngineError) Error() string {
	return e.Message
}

// TemplateErrorCode represents template error codes
type TemplateErrorCode string

const (
	ErrorCodeTemplateNotFound   TemplateErrorCode = "template_not_found"
	ErrorCodeCompilationFailed  TemplateErrorCode = "compilation_failed"
	ErrorCodeRenderingFailed    TemplateErrorCode = "rendering_failed"
	ErrorCodeValidationFailed   TemplateErrorCode = "validation_failed"
	ErrorCodeVariableRequired   TemplateErrorCode = "variable_required"
	ErrorCodeVariableInvalid    TemplateErrorCode = "variable_invalid"
	ErrorCodeAssetClientError   TemplateErrorCode = "asset_client_error"
	ErrorCodeCacheError         TemplateErrorCode = "cache_error"
	ErrorCodeConfigurationError TemplateErrorCode = "configuration_error"
)

// TemplateLoader defines the interface for template loading operations
type TemplateLoader interface {
	// LoadByName loads a template by name from Asset Client
	LoadByName(ctx context.Context, name string) (*assets.AssetContent, error)

	// LoadByCategory loads templates by category
	LoadByCategory(ctx context.Context, category string) ([]*assets.AssetContent, error)

	// ListAvailable lists available templates
	ListAvailable(ctx context.Context, filter assets.AssetFilter) (*assets.AssetList, error)

	// GetMetadata extracts metadata from template content
	GetMetadata(ctx context.Context, content string) (*TemplateMetadata, error)
}

// VariableValidator defines the interface for variable validation
type VariableValidator interface {
	// ValidateRequired validates that all required variables are present
	ValidateRequired(variables map[string]interface{}, specs []VariableSpec) []ValidationError

	// ValidateTypes validates variable types against specifications
	ValidateTypes(variables map[string]interface{}, specs []VariableSpec) []ValidationError

	// ValidateConstraints validates variable constraints (regex, ranges, etc.)
	ValidateConstraints(variables map[string]interface{}, specs []VariableSpec) []ValidationError

	// ApplyDefaults applies default values for missing variables
	ApplyDefaults(variables map[string]interface{}, specs []VariableSpec) map[string]interface{}
}

// FunctionRegistry defines the interface for custom template functions
type FunctionRegistry interface {
	// RegisterFunction registers a custom template function
	RegisterFunction(name string, fn interface{}) error

	// GetFunctions returns all registered functions as template.FuncMap
	GetFunctions() template.FuncMap

	// RegisterZenFunctions registers Zen-specific template functions
	RegisterZenFunctions() error
}

// CacheManager defines the interface for template caching
type CacheManager interface {
	// Get retrieves a compiled template from cache
	Get(key string) (*Template, bool)

	// Set stores a compiled template in cache
	Set(key string, tmpl *Template) error

	// Delete removes a template from cache
	Delete(key string) error

	// Clear clears all cached templates
	Clear() error

	// Stats returns cache statistics
	Stats() CacheStats
}

// CacheStats represents cache statistics
type CacheStats struct {
	Size     int     `json:"size"`
	Hits     int64   `json:"hits"`
	Misses   int64   `json:"misses"`
	HitRatio float64 `json:"hit_ratio"`
}
