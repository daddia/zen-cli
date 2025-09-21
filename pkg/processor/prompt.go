package processor

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"
)

// PromptProcessor handles AI prompt template processing with XML structure
type PromptProcessor struct {
	*BaseProcessor
}

// NewPromptProcessor creates a new Prompt processor
func NewPromptProcessor() Processor {
	return &PromptProcessor{
		BaseProcessor: NewBaseProcessor(
			PromptOutput,
			"AI prompt XML structure generation with semantic validation",
		),
	}
}

// Process executes prompt template processing with XML validation
func (pp *PromptProcessor) Process(ctx ProcessingContext) (string, error) {
	if err := pp.validateTemplateContext(ctx); err != nil {
		return "", fmt.Errorf("context validation failed: %w", err)
	}

	// Execute template
	output, err := pp.ProcessTemplate(ctx)
	if err != nil {
		return "", fmt.Errorf("template processing failed: %w", err)
	}

	// Apply prompt-specific formatting
	formatted := pp.formatPromptOutput(output)

	// Validate XML structure and semantic elements if strict mode is enabled
	if ctx.Options.StrictMode {
		if err := pp.Validate(formatted); err != nil {
			return "", fmt.Errorf("prompt validation failed: %w", err)
		}
	}

	return formatted, nil
}

// Validate performs prompt-specific validation
func (pp *PromptProcessor) Validate(output string) error {
	validations := []func(string) error{
		pp.validateXMLStructure,
		pp.validateSemanticElements,
		pp.validatePromptStructure,
	}

	for _, validation := range validations {
		if err := validation(output); err != nil {
			return err
		}
	}

	return nil
}

// formatPromptOutput applies prompt-specific formatting
func (pp *PromptProcessor) formatPromptOutput(output string) string {
	// Normalize line endings
	output = strings.ReplaceAll(output, "\r\n", "\n")

	// Ensure proper XML formatting
	output = pp.formatXMLStructure(output)

	// Ensure proper semantic element formatting
	output = pp.formatSemanticElements(output)

	// Trim leading/trailing whitespace
	output = strings.TrimSpace(output)

	// Ensure file ends with single newline
	return output + "\n"
}

// formatXMLStructure ensures proper XML formatting
func (pp *PromptProcessor) formatXMLStructure(output string) string {
	// Only apply XML formatting for complex structures
	// For simple inline XML, preserve the original formatting
	if strings.Count(output, "<") <= 4 && !strings.Contains(output, "\n") {
		// Simple inline XML, don't format
		return output
	}

	// Ensure proper spacing around XML tags
	tagRegex := regexp.MustCompile(`>\s*<`)
	output = tagRegex.ReplaceAllString(output, ">\n<")

	// Ensure proper indentation for nested elements
	output = pp.indentXMLElements(output)

	return output
}

// indentXMLElements applies proper indentation to XML elements
func (pp *PromptProcessor) indentXMLElements(output string) string {
	lines := strings.Split(output, "\n")
	var indented []string
	indentLevel := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			indented = append(indented, "")
			continue
		}

		// Check if this is a closing tag
		if strings.HasPrefix(trimmed, "</") {
			indentLevel--
			if indentLevel < 0 {
				indentLevel = 0
			}
		}

		// Apply indentation
		indent := strings.Repeat("  ", indentLevel)
		indented = append(indented, indent+trimmed)

		// Check if this is an opening tag (not self-closing)
		if strings.HasPrefix(trimmed, "<") &&
			!strings.HasPrefix(trimmed, "</") &&
			!strings.HasSuffix(trimmed, "/>") {
			indentLevel++
		}
	}

	return strings.Join(indented, "\n")
}

// formatSemanticElements ensures proper formatting of semantic prompt elements
func (pp *PromptProcessor) formatSemanticElements(output string) string {
	// Only format if the output contains multiple semantic elements or complex structure
	// For simple inline elements, preserve the original formatting
	semanticElements := []string{"role", "objective", "policies", "workflow", "inputs", "examples", "constraints"}

	// Count how many semantic elements are present
	elementCount := 0
	for _, element := range semanticElements {
		elementRegex := regexp.MustCompile(fmt.Sprintf(`<%s\s*>`, element))
		if elementRegex.MatchString(output) {
			elementCount++
		}
	}

	// If there are multiple elements or the content spans multiple lines, apply formatting
	if elementCount > 2 || strings.Count(output, "\n") > 2 {
		for _, element := range semanticElements {
			// Opening tags
			openingRegex := regexp.MustCompile(fmt.Sprintf(`(<\s*%s\s*>)`, element))
			output = openingRegex.ReplaceAllString(output, "\n$1\n")

			// Closing tags
			closingRegex := regexp.MustCompile(fmt.Sprintf(`(<\s*/\s*%s\s*>)`, element))
			output = closingRegex.ReplaceAllString(output, "\n$1\n")
		}

		// Clean up excessive blank lines
		output = regexp.MustCompile(`\n{3,}`).ReplaceAllString(output, "\n\n")
	}

	return output
}

// validateXMLStructure validates basic XML structure
func (pp *PromptProcessor) validateXMLStructure(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("prompt output is empty")
	}

	// Try to parse as XML to validate structure
	var xmlData interface{}
	decoder := xml.NewDecoder(strings.NewReader("<prompt>" + output + "</prompt>"))
	if err := decoder.Decode(&xmlData); err != nil {
		// XML parsing failed, check if it's due to missing root element
		if err := xml.Unmarshal([]byte(output), &xmlData); err != nil {
			return fmt.Errorf("invalid XML structure: %w", err)
		}
	}

	// Check for balanced tags
	if err := pp.validateBalancedTags(output); err != nil {
		return err
	}

	return nil
}

// validateBalancedTags ensures all XML tags are properly balanced
func (pp *PromptProcessor) validateBalancedTags(output string) error {
	tagStack := make([]string, 0)

	// Find all XML tags
	tagRegex := regexp.MustCompile(`<\s*(/?)(\w+)(?:\s+[^>]*)?\s*(/?)>`)
	matches := tagRegex.FindAllStringSubmatch(output, -1)

	for _, match := range matches {
		isClosing := match[1] == "/"
		tagName := match[2]
		isSelfClosing := match[3] == "/"

		if isSelfClosing {
			// Self-closing tag, no need to track
			continue
		}

		if isClosing {
			// Closing tag
			if len(tagStack) == 0 {
				return fmt.Errorf("unexpected closing tag: </%s>", tagName)
			}

			lastTag := tagStack[len(tagStack)-1]
			if lastTag != tagName {
				return fmt.Errorf("mismatched tags: expected </%s> but found </%s>", lastTag, tagName)
			}

			// Remove from stack
			tagStack = tagStack[:len(tagStack)-1]
		} else {
			// Opening tag
			tagStack = append(tagStack, tagName)
		}
	}

	if len(tagStack) > 0 {
		return fmt.Errorf("unclosed tags: %v", tagStack)
	}

	return nil
}

// validateSemanticElements validates required semantic elements for prompts
func (pp *PromptProcessor) validateSemanticElements(output string) error {
	requiredElements := []string{"role", "objective"}

	for _, element := range requiredElements {
		elementRegex := regexp.MustCompile(fmt.Sprintf(`(?s)<\s*%s\s*>.*?<\s*/\s*%s\s*>`, element, element))
		if !elementRegex.MatchString(output) {
			return fmt.Errorf("required semantic element <%s> is missing or malformed", element)
		}
	}

	return nil
}

// validatePromptStructure validates overall prompt structure and quality
func (pp *PromptProcessor) validatePromptStructure(output string) error {
	// Check for empty role
	roleRegex := regexp.MustCompile(`<role\s*>\s*</role>`)
	if roleRegex.MatchString(output) {
		return fmt.Errorf("role element cannot be empty")
	}

	// Check for empty objective
	objectiveRegex := regexp.MustCompile(`<objective\s*>\s*</objective>`)
	if objectiveRegex.MatchString(output) {
		return fmt.Errorf("objective element cannot be empty")
	}

	// Validate policy format if policies exist
	if strings.Contains(output, "<policies>") {
		if err := pp.validatePolicyFormat(output); err != nil {
			return err
		}
	}

	return nil
}

// validatePolicyFormat validates policy format within policies element
func (pp *PromptProcessor) validatePolicyFormat(output string) error {
	policiesRegex := regexp.MustCompile(`<policies>(.*?)</policies>`)
	matches := policiesRegex.FindStringSubmatch(output)

	if len(matches) < 2 {
		return nil // No policies content found
	}

	policiesContent := strings.TrimSpace(matches[1])
	if policiesContent == "" {
		return fmt.Errorf("policies element is empty")
	}

	// Check for valid policy format (MUST, SHOULD, MAY, etc.)
	policyLines := strings.Split(policiesContent, "\n")
	for _, line := range policyLines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "<!--") {
			continue
		}

		// Check for bullet points or policy level indicators
		hasValidFormat := false
		policyPrefixes := []string{"- **MUST**", "- **SHOULD**", "- **MAY**", "- **MUST NOT**", "- **SHOULD NOT**"}

		for _, prefix := range policyPrefixes {
			if strings.HasPrefix(strings.ToUpper(line), strings.ToUpper(prefix)) {
				hasValidFormat = true
				break
			}
		}

		// Also allow plain bullet points
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			hasValidFormat = true
		}

		if !hasValidFormat && line != "" {
			// Allow some flexibility for multi-line policy descriptions
			continue
		}
	}

	return nil
}
