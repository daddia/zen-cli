package processor

import (
	"bytes"
	"fmt"
)

// BaseProcessor provides common functionality for all processors
type BaseProcessor struct {
	outputType  OutputType
	description string
}

// NewBaseProcessor creates a new base processor
func NewBaseProcessor(outputType OutputType, description string) *BaseProcessor {
	return &BaseProcessor{
		outputType:  outputType,
		description: description,
	}
}

// ProcessTemplate executes the template and returns the output
func (bp *BaseProcessor) ProcessTemplate(ctx ProcessingContext) (string, error) {
	if ctx.Template == nil {
		return "", fmt.Errorf("template is nil")
	}

	var buf bytes.Buffer
	if err := ctx.Template.Execute(&buf, ctx.Data); err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	return buf.String(), nil
}

// GetSupportedType returns the supported output type
func (bp *BaseProcessor) GetSupportedType() OutputType {
	return bp.outputType
}

// GetDescription returns the processor description
func (bp *BaseProcessor) GetDescription() string {
	return bp.description
}

// validateTemplateContext performs basic validation of processing context
func (bp *BaseProcessor) validateTemplateContext(ctx ProcessingContext) error {
	if ctx.Template == nil {
		return fmt.Errorf("template is required")
	}

	if ctx.Data == nil {
		return fmt.Errorf("data context is required")
	}

	return nil
}
