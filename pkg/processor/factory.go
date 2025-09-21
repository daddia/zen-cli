package processor

import (
	"fmt"
)

// Processor defines the interface for processing templates of specific types
type Processor interface {
	Process(ctx ProcessingContext) (string, error)
	Validate(output string) error
	GetSupportedType() OutputType
	GetDescription() string
}

// Factory creates processors for different output types
type Factory interface {
	CreateProcessor(outputType OutputType) (Processor, error)
	GetSupportedTypes() []OutputType
	RegisterProcessor(outputType OutputType, processor Processor)
}

// factory implements Factory with type-specific processor creation
type factory struct {
	processors map[OutputType]func() Processor
}

// NewFactory creates a new processor factory with default processors
func NewFactory() Factory {
	f := &factory{
		processors: make(map[OutputType]func() Processor),
	}
	f.registerDefaultProcessors()
	return f
}

// CreateProcessor creates a processor for the specified output type
func (f *factory) CreateProcessor(outputType OutputType) (Processor, error) {
	if processorFunc, exists := f.processors[outputType]; exists {
		return processorFunc(), nil
	}

	return nil, fmt.Errorf("no processor available for output type: %s", outputType)
}

// GetSupportedTypes returns all supported output types
func (f *factory) GetSupportedTypes() []OutputType {
	var types []OutputType
	for outputType := range f.processors {
		types = append(types, outputType)
	}
	return types
}

// RegisterProcessor registers a custom processor for an output type
func (f *factory) RegisterProcessor(outputType OutputType, processor Processor) {
	f.processors[outputType] = func() Processor {
		return processor
	}
}

// registerDefaultProcessors registers all default processors
func (f *factory) registerDefaultProcessors() {
	// Register Markdown processor
	f.processors[MarkdownOutput] = NewMarkdownProcessor

	// Register YAML processor
	f.processors[YAMLOutput] = NewYAMLProcessor

	// Register JSON processor
	f.processors[JSONOutput] = NewJSONProcessor

	// Register XML processor
	f.processors[XMLOutput] = NewXMLProcessor

	// Register Prompt processor
	f.processors[PromptOutput] = NewPromptProcessor

	// Register Docker processor
	f.processors[DockerOutput] = NewDockerProcessor

	// Register OpenAPI processor
	f.processors[OpenAPIOutput] = NewOpenAPIProcessor

	// Register Default processor
	f.processors[DefaultOutput] = NewDefaultProcessor
}
