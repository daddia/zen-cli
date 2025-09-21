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
	f.processors[MarkdownOutput] = func() Processor {
		return NewMarkdownProcessor()
	}

	// Register YAML processor
	f.processors[YAMLOutput] = func() Processor {
		return NewYAMLProcessor()
	}

	// Register JSON processor
	f.processors[JSONOutput] = func() Processor {
		return NewJSONProcessor()
	}

	// Register XML processor
	f.processors[XMLOutput] = func() Processor {
		return NewXMLProcessor()
	}

	// Register Prompt processor
	f.processors[PromptOutput] = func() Processor {
		return NewPromptProcessor()
	}

	// Register Docker processor
	f.processors[DockerOutput] = func() Processor {
		return NewDockerProcessor()
	}

	// Register OpenAPI processor
	f.processors[OpenAPIOutput] = func() Processor {
		return NewOpenAPIProcessor()
	}

	// Register Default processor
	f.processors[DefaultOutput] = func() Processor {
		return NewDefaultProcessor()
	}
}
