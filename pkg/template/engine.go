package template

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"sync"
	"text/template"
	"time"

	"github.com/daddia/zen/internal/config"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/assets"
	"github.com/go-viper/mapstructure/v2"
)

// Engine implements the TemplateEngine interface
type Engine struct {
	logger    logging.Logger
	loader    TemplateLoader
	validator VariableValidator
	functions FunctionRegistry
	cache     CacheManager

	// Configuration
	config EngineConfig

	// Internal state (mutex for concurrent template operations)
	mu sync.RWMutex //nolint:unused // Used for concurrent safety in template operations
}

// EngineConfig configures the template engine
type EngineConfig struct {
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

// DefaultEngineConfig returns default template engine configuration
func DefaultEngineConfig() EngineConfig {
	cfg := EngineConfig{
		CacheEnabled:  true,
		CacheTTL:      30 * time.Minute,
		CacheSize:     100,
		StrictMode:    false,
		EnableAI:      false,
		WorkspaceRoot: ".",
	}
	cfg.DefaultDelims.Left = "{{"
	cfg.DefaultDelims.Right = "}}"
	return cfg
}

// Implement config.Configurable interface

// Validate validates the template engine configuration
func (c EngineConfig) Validate() error {
	if c.CacheSize < 0 {
		return fmt.Errorf("cache_size must be non-negative")
	}
	if c.CacheTTL < 0 {
		return fmt.Errorf("cache_ttl must be non-negative")
	}
	if c.DefaultDelims.Left == "" {
		return fmt.Errorf("left delimiter cannot be empty")
	}
	if c.DefaultDelims.Right == "" {
		return fmt.Errorf("right delimiter cannot be empty")
	}
	return nil
}

// Defaults returns a new EngineConfig with default values
func (c EngineConfig) Defaults() config.Configurable {
	return DefaultEngineConfig()
}

// ConfigParser implements config.ConfigParser[EngineConfig] interface
type ConfigParser struct{}

// Parse converts raw configuration data to EngineConfig
func (p ConfigParser) Parse(raw map[string]interface{}) (EngineConfig, error) {
	// Start with defaults to ensure all fields are properly initialized
	cfg := DefaultEngineConfig()

	// If raw data is empty, return defaults
	if len(raw) == 0 {
		return cfg, nil
	}

	// Use mapstructure to decode the raw map into our config struct
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &cfg,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		),
	})
	if err != nil {
		return cfg, fmt.Errorf("failed to create decoder: %w", err)
	}

	if err := decoder.Decode(raw); err != nil {
		return cfg, fmt.Errorf("failed to decode template config: %w", err)
	}

	return cfg, nil
}

// Section returns the configuration section name for templates
func (p ConfigParser) Section() string {
	return "templates"
}

// NewEngine creates a new template engine with dependencies
func NewEngine(
	logger logging.Logger,
	assetClient assets.AssetClientInterface,
	config EngineConfig,
) *Engine {
	// Initialize components
	loader := NewAssetLoader(assetClient, logger)
	validator := NewVariableValidator(logger)
	functions := NewFunctionRegistry(logger, config.WorkspaceRoot)
	cache := NewMemoryCache(config.CacheSize, config.CacheTTL, logger)

	// Set default delimiters if not specified
	if config.DefaultDelims.Left == "" {
		config.DefaultDelims.Left = "{{"
	}
	if config.DefaultDelims.Right == "" {
		config.DefaultDelims.Right = "}}"
	}

	engine := &Engine{
		logger:    logger,
		loader:    loader,
		validator: validator,
		functions: functions,
		cache:     cache,
		config:    config,
	}

	// Register Zen-specific functions
	if err := functions.RegisterZenFunctions(); err != nil {
		logger.Warn("failed to register Zen functions", "error", err)
	}

	return engine
}

// LoadTemplate loads a template by name from Asset Client
func (e *Engine) LoadTemplate(ctx context.Context, name string) (*Template, error) {
	e.logger.Debug("loading template", "name", name)

	// Check cache first if enabled
	if e.config.CacheEnabled {
		if cached, found := e.cache.Get(name); found {
			e.logger.Debug("template found in cache", "name", name)
			return cached, nil
		}
	}

	// Load from asset client
	assetContent, err := e.loader.LoadByName(ctx, name)
	if err != nil {
		return nil, &TemplateEngineError{
			Code:    ErrorCodeAssetClientError,
			Message: fmt.Sprintf("failed to load template '%s': %v", name, err),
			Details: err,
		}
	}

	// Extract metadata from asset
	metadata, err := e.extractMetadata(assetContent)
	if err != nil {
		return nil, &TemplateEngineError{
			Code:    ErrorCodeConfigurationError,
			Message: fmt.Sprintf("failed to extract template metadata: %v", err),
			Details: err,
		}
	}

	// Compile template
	tmpl, err := e.CompileTemplate(ctx, name, assetContent.Content, metadata)
	if err != nil {
		return nil, err
	}

	// Cache compiled template if enabled
	if e.config.CacheEnabled {
		if err := e.cache.Set(name, tmpl); err != nil {
			e.logger.Warn("failed to cache template", "name", name, "error", err)
		}
	}

	return tmpl, nil
}

// CompileTemplate compiles a template string into a Template
func (e *Engine) CompileTemplate(ctx context.Context, name, content string, metadata *TemplateMetadata) (*Template, error) {
	e.logger.Debug("compiling template", "name", name, "size", len(content))

	// Create new Go template
	goTmpl := template.New(name)

	// Set custom delimiters if specified
	goTmpl = goTmpl.Delims(e.config.DefaultDelims.Left, e.config.DefaultDelims.Right)

	// Add custom functions
	goTmpl = goTmpl.Funcs(e.functions.GetFunctions())

	// Set missing key behavior based on strict mode
	if e.config.StrictMode {
		goTmpl = goTmpl.Option("missingkey=error")
	} else {
		goTmpl = goTmpl.Option("missingkey=zero")
	}

	// Parse template content
	compiled, err := goTmpl.Parse(content)
	if err != nil {
		return nil, &TemplateEngineError{
			Code:    ErrorCodeCompilationFailed,
			Message: fmt.Sprintf("failed to compile template '%s': %v", name, err),
			Details: err,
		}
	}

	// Calculate checksum
	checksum := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(content)))

	// Create Template instance
	tmpl := &Template{
		Name:       name,
		Content:    content,
		Compiled:   compiled,
		Metadata:   metadata,
		Variables:  metadata.Variables,
		CompiledAt: time.Now(),
		Checksum:   checksum,
	}

	e.logger.Debug("template compiled successfully", "name", name, "checksum", checksum)
	return tmpl, nil
}

// RenderTemplate renders a template with provided variables
func (e *Engine) RenderTemplate(ctx context.Context, tmpl *Template, variables map[string]interface{}) (string, error) {
	e.logger.Debug("rendering template", "name", tmpl.Name, "variables", len(variables))

	// Validate variables if template has specifications
	if len(tmpl.Variables) > 0 {
		if err := e.ValidateVariables(ctx, tmpl, variables); err != nil {
			return "", err
		}
	}

	// Apply default values for missing variables
	enrichedVariables := e.validator.ApplyDefaults(variables, tmpl.Variables)

	// Create render context
	renderCtx := &RenderContext{
		Variables:     enrichedVariables,
		Functions:     e.functions.GetFunctions(),
		Metadata:      tmpl.Metadata,
		WorkspaceRoot: e.config.WorkspaceRoot,
		Options: RenderOptions{
			StrictVariables: e.config.StrictMode,
			EnableAI:        e.config.EnableAI,
			ValidateOutput:  true,
			Delims: struct {
				Left  string `json:"left"`
				Right string `json:"right"`
			}{
				Left:  e.config.DefaultDelims.Left,
				Right: e.config.DefaultDelims.Right,
			},
		},
	}

	// Add render context to variables for template access
	enrichedVariables["__context"] = renderCtx

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Compiled.Execute(&buf, enrichedVariables); err != nil {
		return "", &TemplateEngineError{
			Code:    ErrorCodeRenderingFailed,
			Message: fmt.Sprintf("failed to render template '%s': %v", tmpl.Name, err),
			Details: err,
		}
	}

	result := buf.String()
	e.logger.Debug("template rendered successfully", "name", tmpl.Name, "output_size", len(result))

	return result, nil
}

// ListTemplates lists available templates matching the filter
func (e *Engine) ListTemplates(ctx context.Context, filter TemplateFilter) (*TemplateList, error) {
	e.logger.Debug("listing templates", "filter", filter)

	// Convert to asset filter
	assetFilter := assets.AssetFilter{
		Type:     assets.AssetTypeTemplate,
		Category: filter.Category,
		Tags:     filter.Tags,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	}

	// Get assets from loader
	assetList, err := e.loader.ListAvailable(ctx, assetFilter)
	if err != nil {
		return nil, &TemplateEngineError{
			Code:    ErrorCodeAssetClientError,
			Message: fmt.Sprintf("failed to list templates: %v", err),
			Details: err,
		}
	}

	// Convert asset metadata to template metadata
	templates := make([]TemplateMetadata, 0, len(assetList.Assets))
	for _, asset := range assetList.Assets {
		tmplMetadata := TemplateMetadata{
			Name:        asset.Name,
			Description: asset.Description,
			Category:    asset.Category,
			Tags:        asset.Tags,
			UpdatedAt:   asset.UpdatedAt,
			Variables:   convertAssetVariables(asset.Variables),
		}
		templates = append(templates, tmplMetadata)
	}

	return &TemplateList{
		Templates: templates,
		Total:     assetList.Total,
		HasMore:   assetList.HasMore,
		Filter:    filter,
	}, nil
}

// ValidateVariables validates template variables against template requirements
func (e *Engine) ValidateVariables(ctx context.Context, tmpl *Template, variables map[string]interface{}) error {
	e.logger.Debug("validating template variables", "template", tmpl.Name, "variables", len(variables))

	var allErrors []ValidationError

	// Check required variables
	requiredErrors := e.validator.ValidateRequired(variables, tmpl.Variables)
	allErrors = append(allErrors, requiredErrors...)

	// Check variable types
	typeErrors := e.validator.ValidateTypes(variables, tmpl.Variables)
	allErrors = append(allErrors, typeErrors...)

	// Check constraints (regex, ranges, etc.)
	constraintErrors := e.validator.ValidateConstraints(variables, tmpl.Variables)
	allErrors = append(allErrors, constraintErrors...)

	if len(allErrors) > 0 {
		return &TemplateEngineError{
			Code:    ErrorCodeValidationFailed,
			Message: fmt.Sprintf("template variable validation failed: %d errors", len(allErrors)),
			Details: ValidationResult{
				Valid:  false,
				Errors: allErrors,
			},
		}
	}

	e.logger.Debug("template variables validated successfully", "template", tmpl.Name)
	return nil
}

// GetFunctions returns the available template functions
func (e *Engine) GetFunctions() template.FuncMap {
	return e.functions.GetFunctions()
}

// extractMetadata extracts template metadata from asset content
func (e *Engine) extractMetadata(asset *assets.AssetContent) (*TemplateMetadata, error) {
	// Use asset metadata as base
	metadata := &TemplateMetadata{
		Name:        asset.Metadata.Name,
		Description: asset.Metadata.Description,
		Category:    asset.Metadata.Category,
		Tags:        asset.Metadata.Tags,
		UpdatedAt:   asset.Metadata.UpdatedAt,
		Variables:   convertAssetVariables(asset.Metadata.Variables),
	}

	// Try to extract additional metadata from template content
	if contentMetadata, err := e.loader.GetMetadata(context.Background(), asset.Content); err == nil {
		// Merge content metadata with asset metadata
		if contentMetadata.Version != "" {
			metadata.Version = contentMetadata.Version
		}
		if contentMetadata.Author != "" {
			metadata.Author = contentMetadata.Author
		}
		if !contentMetadata.CreatedAt.IsZero() {
			metadata.CreatedAt = contentMetadata.CreatedAt
		}
		if len(contentMetadata.Variables) > 0 {
			// Content metadata variables take precedence
			metadata.Variables = contentMetadata.Variables
		}
	}

	return metadata, nil
}

// convertAssetVariables converts asset variables to template variable specs
func convertAssetVariables(assetVars []assets.Variable) []VariableSpec {
	vars := make([]VariableSpec, len(assetVars))
	for i, v := range assetVars {
		vars[i] = VariableSpec{
			Name:        v.Name,
			Description: v.Description,
			Type:        v.Type,
			Required:    v.Required,
			Default:     v.Default,
		}
	}
	return vars
}
