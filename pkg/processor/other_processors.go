package processor

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// JSONProcessor handles JSON template processing
type JSONProcessor struct {
	*BaseProcessor
}

// NewJSONProcessor creates a new JSON processor
func NewJSONProcessor() Processor {
	return &JSONProcessor{
		BaseProcessor: NewBaseProcessor(
			JSONOutput,
			"JSON data file generation with schema validation",
		),
	}
}

// Process executes JSON template processing
func (jp *JSONProcessor) Process(ctx ProcessingContext) (string, error) {
	if err := jp.validateTemplateContext(ctx); err != nil {
		return "", fmt.Errorf("context validation failed: %w", err)
	}

	// Execute template
	output, err := jp.ProcessTemplate(ctx)
	if err != nil {
		return "", fmt.Errorf("template processing failed: %w", err)
	}

	// Apply JSON-specific formatting
	formatted := jp.formatJSONOutput(output)

	// Validate JSON syntax if strict mode is enabled
	if ctx.Options.StrictMode {
		if err := jp.Validate(formatted); err != nil {
			return "", fmt.Errorf("JSON validation failed: %w", err)
		}
	}

	return formatted, nil
}

// Validate performs JSON-specific validation
func (jp *JSONProcessor) Validate(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("JSON output is empty")
	}

	// Parse JSON to validate syntax
	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err != nil {
		return fmt.Errorf("invalid JSON syntax: %w", err)
	}

	return nil
}

// formatJSONOutput applies JSON-specific formatting
func (jp *JSONProcessor) formatJSONOutput(output string) string {
	output = strings.TrimSpace(output)

	// Try to format JSON with proper indentation
	var jsonData interface{}
	if err := json.Unmarshal([]byte(output), &jsonData); err == nil {
		if formatted, err := json.MarshalIndent(jsonData, "", "  "); err == nil {
			return string(formatted) + "\n"
		}
	}

	return output + "\n"
}

// XMLProcessor handles XML template processing
type XMLProcessor struct {
	*BaseProcessor
}

// NewXMLProcessor creates a new XML processor
func NewXMLProcessor() Processor {
	return &XMLProcessor{
		BaseProcessor: NewBaseProcessor(
			XMLOutput,
			"XML document generation with structure validation",
		),
	}
}

// Process executes XML template processing
func (xp *XMLProcessor) Process(ctx ProcessingContext) (string, error) {
	if err := xp.validateTemplateContext(ctx); err != nil {
		return "", fmt.Errorf("context validation failed: %w", err)
	}

	// Execute template
	output, err := xp.ProcessTemplate(ctx)
	if err != nil {
		return "", fmt.Errorf("template processing failed: %w", err)
	}

	// Apply XML-specific formatting
	formatted := xp.formatXMLOutput(output)

	// Validate XML syntax if strict mode is enabled
	if ctx.Options.StrictMode {
		if err := xp.Validate(formatted); err != nil {
			return "", fmt.Errorf("XML validation failed: %w", err)
		}
	}

	return formatted, nil
}

// Validate performs XML-specific validation
func (xp *XMLProcessor) Validate(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("XML output is empty")
	}

	// Parse XML to validate syntax
	var xmlData interface{}
	if err := xml.Unmarshal([]byte(output), &xmlData); err != nil {
		return fmt.Errorf("invalid XML syntax: %w", err)
	}

	return nil
}

// formatXMLOutput applies XML-specific formatting
func (xp *XMLProcessor) formatXMLOutput(output string) string {
	output = strings.TrimSpace(output)
	return output + "\n"
}

// DockerProcessor handles Dockerfile template processing
type DockerProcessor struct {
	*BaseProcessor
}

// NewDockerProcessor creates a new Docker processor
func NewDockerProcessor() Processor {
	return &DockerProcessor{
		BaseProcessor: NewBaseProcessor(
			DockerOutput,
			"Dockerfile generation with syntax and best practice validation",
		),
	}
}

// Process executes Dockerfile template processing
func (dp *DockerProcessor) Process(ctx ProcessingContext) (string, error) {
	if err := dp.validateTemplateContext(ctx); err != nil {
		return "", fmt.Errorf("context validation failed: %w", err)
	}

	// Execute template
	output, err := dp.ProcessTemplate(ctx)
	if err != nil {
		return "", fmt.Errorf("template processing failed: %w", err)
	}

	// Apply Dockerfile-specific formatting
	formatted := dp.formatDockerOutput(output)

	// Validate Dockerfile syntax if strict mode is enabled
	if ctx.Options.StrictMode {
		if err := dp.Validate(formatted); err != nil {
			return "", fmt.Errorf("dockerfile validation failed: %w", err)
		}
	}

	return formatted, nil
}

// Validate performs Dockerfile-specific validation
func (dp *DockerProcessor) Validate(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("dockerfile output is empty")
	}

	// Check for required FROM instruction
	fromRegex := regexp.MustCompile(`(?i)^\s*FROM\s+`)
	lines := strings.Split(output, "\n")
	hasFrom := false

	for _, line := range lines {
		if strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			if fromRegex.MatchString(line) {
				hasFrom = true
				break
			} else {
				return fmt.Errorf("dockerfile must start with FROM instruction")
			}
		}
	}

	if !hasFrom {
		return fmt.Errorf("dockerfile is missing FROM instruction")
	}

	return nil
}

// formatDockerOutput applies Dockerfile-specific formatting
func (dp *DockerProcessor) formatDockerOutput(output string) string {
	lines := strings.Split(output, "\n")
	var formatted []string

	for _, line := range lines {
		// Trim trailing whitespace
		line = strings.TrimRight(line, " \t")
		formatted = append(formatted, line)
	}

	output = strings.Join(formatted, "\n")
	output = strings.TrimSpace(output)
	return output + "\n"
}

// OpenAPIProcessor handles OpenAPI specification template processing
type OpenAPIProcessor struct {
	*BaseProcessor
}

// NewOpenAPIProcessor creates a new OpenAPI processor
func NewOpenAPIProcessor() Processor {
	return &OpenAPIProcessor{
		BaseProcessor: NewBaseProcessor(
			OpenAPIOutput,
			"OpenAPI specification generation with schema validation",
		),
	}
}

// Process executes OpenAPI template processing
func (op *OpenAPIProcessor) Process(ctx ProcessingContext) (string, error) {
	if err := op.validateTemplateContext(ctx); err != nil {
		return "", fmt.Errorf("context validation failed: %w", err)
	}

	// Execute template
	output, err := op.ProcessTemplate(ctx)
	if err != nil {
		return "", fmt.Errorf("template processing failed: %w", err)
	}

	// Apply OpenAPI-specific formatting
	formatted := op.formatOpenAPIOutput(output)

	// Validate OpenAPI syntax if strict mode is enabled
	if ctx.Options.StrictMode {
		if err := op.Validate(formatted); err != nil {
			return "", fmt.Errorf("OpenAPI validation failed: %w", err)
		}
	}

	return formatted, nil
}

// Validate performs OpenAPI-specific validation
func (op *OpenAPIProcessor) Validate(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("OpenAPI output is empty")
	}

	// Parse as YAML first (most OpenAPI specs are YAML)
	var spec map[string]interface{}
	if err := yaml.Unmarshal([]byte(output), &spec); err != nil {
		return fmt.Errorf("invalid OpenAPI YAML syntax: %w", err)
	}

	// Check for required OpenAPI fields
	if _, exists := spec["openapi"]; !exists {
		return fmt.Errorf("missing required 'openapi' field")
	}

	if _, exists := spec["info"]; !exists {
		return fmt.Errorf("missing required 'info' field")
	}

	return nil
}

// formatOpenAPIOutput applies OpenAPI-specific formatting
func (op *OpenAPIProcessor) formatOpenAPIOutput(output string) string {
	// Use YAML formatting since most OpenAPI specs are YAML
	lines := strings.Split(output, "\n")
	var formatted []string

	for _, line := range lines {
		formatted = append(formatted, strings.TrimRight(line, " \t"))
	}

	output = strings.Join(formatted, "\n")
	output = strings.TrimSpace(output)
	return output + "\n"
}

// DefaultProcessor handles generic template processing
type DefaultProcessor struct {
	*BaseProcessor
}

// NewDefaultProcessor creates a new default processor
func NewDefaultProcessor() Processor {
	return &DefaultProcessor{
		BaseProcessor: NewBaseProcessor(
			DefaultOutput,
			"Generic template processing with basic validation",
		),
	}
}

// Process executes generic template processing
func (dp *DefaultProcessor) Process(ctx ProcessingContext) (string, error) {
	if err := dp.validateTemplateContext(ctx); err != nil {
		return "", fmt.Errorf("context validation failed: %w", err)
	}

	// Execute template
	output, err := dp.ProcessTemplate(ctx)
	if err != nil {
		return "", fmt.Errorf("template processing failed: %w", err)
	}

	// Apply basic formatting
	formatted := dp.formatDefaultOutput(output)

	return formatted, nil
}

// Validate performs basic validation
func (dp *DefaultProcessor) Validate(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("output is empty")
	}
	return nil
}

// formatDefaultOutput applies basic formatting
func (dp *DefaultProcessor) formatDefaultOutput(output string) string {
	output = strings.ReplaceAll(output, "\r\n", "\n")
	output = strings.TrimSpace(output)
	return output + "\n"
}
