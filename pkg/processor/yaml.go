package processor

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAMLProcessor handles YAML template processing
type YAMLProcessor struct {
	*BaseProcessor
}

// NewYAMLProcessor creates a new YAML processor
func NewYAMLProcessor() Processor {
	return &YAMLProcessor{
		BaseProcessor: NewBaseProcessor(
			YAMLOutput,
			"YAML configuration file generation with syntax validation",
		),
	}
}

// Process executes YAML template processing with validation
func (yp *YAMLProcessor) Process(ctx ProcessingContext) (string, error) {
	if err := yp.validateTemplateContext(ctx); err != nil {
		return "", fmt.Errorf("context validation failed: %w", err)
	}

	// Execute template
	output, err := yp.ProcessTemplate(ctx)
	if err != nil {
		return "", fmt.Errorf("template processing failed: %w", err)
	}

	// Apply YAML-specific formatting
	formatted := yp.formatYAMLOutput(output)

	// Validate YAML syntax if strict mode is enabled
	if ctx.Options.StrictMode {
		if err := yp.Validate(formatted); err != nil {
			return "", fmt.Errorf("YAML validation failed: %w", err)
		}
	}

	return formatted, nil
}

// Validate performs YAML-specific validation
func (yp *YAMLProcessor) Validate(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("YAML output is empty")
	}

	// Parse YAML to validate syntax
	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(output), &yamlData); err != nil {
		return fmt.Errorf("invalid YAML syntax: %w", err)
	}

	// Additional YAML-specific validations
	if err := yp.validateYAMLStructure(output); err != nil {
		return err
	}

	return nil
}

// formatYAMLOutput applies YAML-specific formatting
func (yp *YAMLProcessor) formatYAMLOutput(output string) string {
	// Normalize line endings
	output = strings.ReplaceAll(output, "\r\n", "\n")

	// Remove trailing spaces
	lines := strings.Split(output, "\n")
	var formatted []string

	for _, line := range lines {
		formatted = append(formatted, strings.TrimRight(line, " \t"))
	}

	output = strings.Join(formatted, "\n")

	// Ensure consistent indentation (2 spaces is YAML standard)
	output = yp.normalizeIndentation(output)

	// Trim leading/trailing whitespace
	output = strings.TrimSpace(output)

	// Ensure file ends with single newline
	return output + "\n"
}

// normalizeIndentation ensures consistent YAML indentation
func (yp *YAMLProcessor) normalizeIndentation(output string) string {
	lines := strings.Split(output, "\n")
	var normalized []string

	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			normalized = append(normalized, "")
			continue
		}

		// Count leading spaces
		leadingSpaces := 0
		for _, char := range line {
			if char == ' ' {
				leadingSpaces++
			} else {
				break
			}
		}

		// Convert tabs to spaces if any
		line = strings.ReplaceAll(line, "\t", "  ")

		normalized = append(normalized, line)
	}

	return strings.Join(normalized, "\n")
}

// validateYAMLStructure validates YAML structure and conventions
func (yp *YAMLProcessor) validateYAMLStructure(output string) error {
	lines := strings.Split(output, "\n")

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for tabs (YAML should use spaces)
		if strings.Contains(line, "\t") {
			return fmt.Errorf("tabs found at line %d, YAML should use spaces for indentation", lineNum+1)
		}

		// Check for trailing spaces (common YAML issue)
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			return fmt.Errorf("trailing whitespace found at line %d", lineNum+1)
		}
	}

	return nil
}
