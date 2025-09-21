package processor

import (
	"text/template"
	"time"
)

// OutputType represents the detected output format for a template
type OutputType string

const (
	MarkdownOutput OutputType = "markdown"
	YAMLOutput     OutputType = "yaml"
	JSONOutput     OutputType = "json"
	XMLOutput      OutputType = "xml"
	PromptOutput   OutputType = "prompt"
	DockerOutput   OutputType = "dockerfile"
	OpenAPIOutput  OutputType = "openapi"
	DefaultOutput  OutputType = "text"
)

// TemplateInfo contains metadata about a detected template
type TemplateInfo struct {
	FilePath   string     `json:"file_path"`
	OutputType OutputType `json:"output_type"`
	FileName   string     `json:"file_name"`
	Extension  string     `json:"extension"`
}

// ProcessingContext holds all context needed for template processing
type ProcessingContext struct {
	TemplateInfo TemplateInfo           `json:"template_info"`
	Template     *template.Template     `json:"-"`
	Data         map[string]interface{} `json:"data"`
	Functions    template.FuncMap       `json:"-"`
	Options      ProcessingOptions      `json:"options"`
}

// ProcessingOptions controls template processing behavior
type ProcessingOptions struct {
	ValidateOutput  bool          `json:"validate_output"`
	StrictMode      bool          `json:"strict_mode"`
	TimeoutDuration time.Duration `json:"timeout_duration"`
	MaxMemoryMB     int           `json:"max_memory_mb"`
}

// ProcessingResult contains the result of template processing
type ProcessingResult struct {
	Output           string            `json:"output"`
	TemplateInfo     TemplateInfo      `json:"template_info"`
	ProcessingTimeMS int64             `json:"processing_time_ms"`
	Metadata         map[string]string `json:"metadata,omitempty"`
}
