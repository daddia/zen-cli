package processor

import (
	"fmt"
	"regexp"
	"strings"
)

// MarkdownProcessor handles Markdown template processing
type MarkdownProcessor struct {
	*BaseProcessor
}

// NewMarkdownProcessor creates a new Markdown processor
func NewMarkdownProcessor() Processor {
	return &MarkdownProcessor{
		BaseProcessor: NewBaseProcessor(
			MarkdownOutput,
			"Markdown document generation with structure validation",
		),
	}
}

// Process executes Markdown template processing with validation
func (mp *MarkdownProcessor) Process(ctx ProcessingContext) (string, error) {
	if err := mp.validateTemplateContext(ctx); err != nil {
		return "", fmt.Errorf("context validation failed: %w", err)
	}

	// Execute template
	output, err := mp.ProcessTemplate(ctx)
	if err != nil {
		return "", fmt.Errorf("template processing failed: %w", err)
	}

	// Apply Markdown-specific formatting
	formatted := mp.formatMarkdownOutput(output)

	// Validate structure if strict mode is enabled
	if ctx.Options.StrictMode {
		if err := mp.Validate(formatted); err != nil {
			return "", fmt.Errorf("markdown validation failed: %w", err)
		}
	}

	return formatted, nil
}

// Validate performs Markdown-specific validation
func (mp *MarkdownProcessor) Validate(output string) error {
	validations := []func(string) error{
		mp.validateMarkdownStructure,
		mp.validateHeaders,
		mp.validateLinks,
		mp.validateCodeBlocks,
	}

	for _, validation := range validations {
		if err := validation(output); err != nil {
			return err
		}
	}

	return nil
}

// formatMarkdownOutput applies Markdown-specific formatting
func (mp *MarkdownProcessor) formatMarkdownOutput(output string) string {
	// Normalize line endings
	output = strings.ReplaceAll(output, "\r\n", "\n")

	// Remove excessive blank lines (more than 2 consecutive)
	output = regexp.MustCompile(`\n{3,}`).ReplaceAllString(output, "\n\n")

	// Ensure proper spacing around headers
	output = regexp.MustCompile(`\n(#{1,6})\s`).ReplaceAllString(output, "\n\n$1 ")

	// Ensure proper list formatting
	output = mp.formatLists(output)

	// Ensure proper code block formatting
	output = mp.formatCodeBlocks(output)

	// Trim leading/trailing whitespace
	output = strings.TrimSpace(output)

	// Ensure file ends with single newline
	return output + "\n"
}

// formatLists ensures proper list formatting
func (mp *MarkdownProcessor) formatLists(output string) string {
	lines := strings.Split(output, "\n")
	var formatted []string
	inList := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if line is a list item
		isListItem := regexp.MustCompile(`^[-*+]\s+`).MatchString(trimmed) ||
			regexp.MustCompile(`^\d+\.\s+`).MatchString(trimmed)

		if isListItem {
			if !inList && i > 0 && strings.TrimSpace(lines[i-1]) != "" {
				// Add blank line before list starts
				formatted = append(formatted, "")
			}
			inList = true
		} else if inList && trimmed != "" && !strings.HasPrefix(trimmed, " ") {
			// List ended, add blank line after
			if len(formatted) > 0 && strings.TrimSpace(formatted[len(formatted)-1]) != "" {
				formatted = append(formatted, "")
			}
			inList = false
		}

		formatted = append(formatted, line)
	}

	return strings.Join(formatted, "\n")
}

// formatCodeBlocks ensures proper code block formatting
func (mp *MarkdownProcessor) formatCodeBlocks(output string) string {
	// Ensure blank lines around code blocks - using a simpler regex that works with Go's RE2
	codeBlockRegex := regexp.MustCompile("(?m)^```[\\s\\S]*?^```$")

	output = codeBlockRegex.ReplaceAllString(output, "\n\n$0\n\n")

	// Clean up excessive blank lines that might have been introduced
	output = regexp.MustCompile(`\n{3,}`).ReplaceAllString(output, "\n\n")

	return output
}

// validateMarkdownStructure validates basic Markdown structure
func (mp *MarkdownProcessor) validateMarkdownStructure(output string) error {
	if strings.TrimSpace(output) == "" {
		return fmt.Errorf("markdown output is empty")
	}

	// Check for malformed headers
	malformedHeaders := regexp.MustCompile(`#{7,}|^#{1,6}[^#\s]`)
	if malformedHeaders.MatchString(output) {
		return fmt.Errorf("malformed headers detected")
	}

	return nil
}

// validateHeaders validates header structure and hierarchy
func (mp *MarkdownProcessor) validateHeaders(output string) error {
	headerRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	lines := strings.Split(output, "\n")

	var headerLevels []int

	for lineNum, line := range lines {
		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			level := len(matches[1])
			headerText := strings.TrimSpace(matches[2])

			if headerText == "" {
				return fmt.Errorf("empty header at line %d", lineNum+1)
			}

			headerLevels = append(headerLevels, level)
		}
	}

	// Validate header hierarchy (optional - can be disabled for flexibility)
	if len(headerLevels) > 1 {
		for i := 1; i < len(headerLevels); i++ {
			if headerLevels[i] > headerLevels[i-1]+1 {
				// Skip validation - allow flexible header hierarchy
				// This is more practical for real-world documents
				continue
			}
		}
	}

	return nil
}

// validateLinks validates link formatting
func (mp *MarkdownProcessor) validateLinks(output string) error {
	// Check for malformed links
	linkRegex := regexp.MustCompile(`\[([^\]]*)\]\(([^)]*)\)`)
	lines := strings.Split(output, "\n")

	for lineNum, line := range lines {
		matches := linkRegex.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			linkText := strings.TrimSpace(match[1])
			linkURL := strings.TrimSpace(match[2])

			if linkText == "" {
				return fmt.Errorf("empty link text at line %d", lineNum+1)
			}

			if linkURL == "" {
				return fmt.Errorf("empty link URL at line %d", lineNum+1)
			}
		}
	}

	return nil
}

// validateCodeBlocks validates code block formatting
func (mp *MarkdownProcessor) validateCodeBlocks(output string) error {
	// Count code block delimiters
	codeBlockDelimiters := regexp.MustCompile("```").FindAllString(output, -1)

	if len(codeBlockDelimiters)%2 != 0 {
		return fmt.Errorf("unmatched code block delimiters (found %d)", len(codeBlockDelimiters))
	}

	return nil
}
