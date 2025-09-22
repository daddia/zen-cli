package mapping

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/integration/plugin"
)

// DataMapper provides flexible field mapping with transformation functions and validation
type DataMapper struct {
	logger       logging.Logger
	transformers map[string]FieldTransformerInterface
	validators   map[string]MappingValidatorInterface
}

// DataMapperInterface defines the data mapper interface
type DataMapperInterface interface {
	// MapFields maps fields between external and Zen formats
	MapFields(ctx context.Context, sourceData map[string]interface{}, mapping *plugin.FieldMappingConfig, direction plugin.SyncDirection) (map[string]interface{}, error)

	// ValidateMapping validates a field mapping configuration
	ValidateMapping(mapping *plugin.FieldMappingConfig) error

	// GetDefaultMapping returns the default field mapping for a provider
	GetDefaultMapping(provider string) *plugin.FieldMappingConfig

	// RegisterTransformer registers a custom field transformer
	RegisterTransformer(name string, transformer FieldTransformerInterface) error

	// RegisterValidator registers a custom mapping validator
	RegisterValidator(name string, validator MappingValidatorInterface) error
}

// FieldTransformerInterface defines field transformation interface
type FieldTransformerInterface interface {
	// Transform transforms a field value
	Transform(value interface{}, config map[string]interface{}) (interface{}, error)

	// Validate validates transformation configuration
	Validate(config map[string]interface{}) error

	// Name returns the transformer name
	Name() string
}

// MappingValidatorInterface defines mapping validation interface
type MappingValidatorInterface interface {
	// Validate validates mapped data
	Validate(data map[string]interface{}, rules map[string]interface{}) error

	// Name returns the validator name
	Name() string
}

// MappingError represents a mapping error with field-level details
type MappingError struct {
	Message string       `json:"message"`
	Errors  []FieldError `json:"errors"`
}

func (e *MappingError) Error() string {
	return e.Message
}

// FieldError represents an error with a specific field
type FieldError struct {
	Field   string      `json:"field"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

// NewDataMapper creates a new data mapper
func NewDataMapper(logger logging.Logger) *DataMapper {
	dm := &DataMapper{
		logger:       logger,
		transformers: make(map[string]FieldTransformerInterface),
		validators:   make(map[string]MappingValidatorInterface),
	}

	// Register built-in transformers
	dm.registerBuiltinTransformers()

	// Register built-in validators
	dm.registerBuiltinValidators()

	return dm
}

// MapFields maps fields between external and Zen formats
func (dm *DataMapper) MapFields(ctx context.Context, sourceData map[string]interface{}, mapping *plugin.FieldMappingConfig, direction plugin.SyncDirection) (map[string]interface{}, error) {
	var errors []FieldError
	targetData := make(map[string]interface{})

	dm.logger.Debug("mapping fields", "direction", direction, "mappings", len(mapping.Mappings))

	// Apply field mappings
	for _, fieldMapping := range mapping.Mappings {
		// Check direction compatibility
		if !dm.supportsDirection(fieldMapping.Direction, direction) {
			continue
		}

		// Extract source value
		sourceValue := dm.getNestedValue(sourceData, fieldMapping.ExternalField)

		// Handle required fields
		if sourceValue == nil && fieldMapping.Required {
			errors = append(errors, FieldError{
				Field:   fieldMapping.ZenField,
				Code:    "REQUIRED_FIELD_MISSING",
				Message: fmt.Sprintf("Required field %s is missing", fieldMapping.ExternalField),
			})
			continue
		}

		// Apply default value if needed
		if sourceValue == nil && fieldMapping.DefaultValue != nil {
			sourceValue = fieldMapping.DefaultValue
		}

		// Apply transformations
		transformedValue := sourceValue
		for _, transform := range mapping.Transforms {
			if transform.Field == fieldMapping.ZenField && dm.supportsDirection(transform.Direction, direction) {
				var err error
				transformedValue, err = dm.applyTransform(transformedValue, transform)
				if err != nil {
					errors = append(errors, FieldError{
						Field:   fieldMapping.ZenField,
						Code:    "TRANSFORM_FAILED",
						Message: fmt.Sprintf("Transform failed: %v", err),
						Value:   sourceValue,
					})
					continue
				}
			}
		}

		// Set target field
		if transformedValue != nil {
			dm.setNestedValue(targetData, fieldMapping.ZenField, transformedValue)
		}
	}

	// Apply validation rules
	for _, validation := range mapping.Validation {
		if err := dm.applyValidation(targetData, validation); err != nil {
			errors = append(errors, FieldError{
				Field:   validation.Field,
				Code:    "VALIDATION_FAILED",
				Message: fmt.Sprintf("Validation failed: %v", err),
			})
		}
	}

	// Return results
	if len(errors) > 0 {
		return targetData, &MappingError{
			Message: "Data mapping completed with errors",
			Errors:  errors,
		}
	}

	dm.logger.Debug("field mapping completed successfully", "target_fields", len(targetData))

	return targetData, nil
}

// ValidateMapping validates a field mapping configuration
func (dm *DataMapper) ValidateMapping(mapping *plugin.FieldMappingConfig) error {
	// Validate mappings
	for _, fieldMapping := range mapping.Mappings {
		if fieldMapping.ZenField == "" {
			return fmt.Errorf("zen_field is required")
		}
		if fieldMapping.ExternalField == "" {
			return fmt.Errorf("external_field is required for zen_field: %s", fieldMapping.ZenField)
		}
	}

	// Validate transforms
	for _, transform := range mapping.Transforms {
		if transform.Field == "" {
			return fmt.Errorf("field is required for transform")
		}
		if transform.Type == "" {
			return fmt.Errorf("type is required for transform field: %s", transform.Field)
		}

		// Validate transformer exists
		transformer, exists := dm.transformers[string(transform.Type)]
		if !exists {
			return fmt.Errorf("unknown transformer type: %s", transform.Type)
		}

		// Validate transformer configuration
		if err := transformer.Validate(transform.Config); err != nil {
			return fmt.Errorf("invalid transformer config for field %s: %w", transform.Field, err)
		}
	}

	// Validate validation rules
	for _, validation := range mapping.Validation {
		if validation.Field == "" {
			return fmt.Errorf("field is required for validation")
		}
	}

	return nil
}

// GetDefaultMapping returns the default field mapping for a provider
func (dm *DataMapper) GetDefaultMapping(provider string) *plugin.FieldMappingConfig {
	switch provider {
	case "jira":
		return dm.getJiraDefaultMapping()
	case "github":
		return dm.getGitHubDefaultMapping()
	case "linear":
		return dm.getLinearDefaultMapping()
	default:
		return dm.getGenericDefaultMapping()
	}
}

// RegisterTransformer registers a custom field transformer
func (dm *DataMapper) RegisterTransformer(name string, transformer FieldTransformerInterface) error {
	dm.transformers[name] = transformer
	dm.logger.Debug("registered field transformer", "name", name)
	return nil
}

// RegisterValidator registers a custom mapping validator
func (dm *DataMapper) RegisterValidator(name string, validator MappingValidatorInterface) error {
	dm.validators[name] = validator
	dm.logger.Debug("registered mapping validator", "name", name)
	return nil
}

// Helper methods

// supportsDirection checks if a mapping supports the given sync direction
func (dm *DataMapper) supportsDirection(mappingDirection, syncDirection plugin.SyncDirection) bool {
	if mappingDirection == plugin.SyncDirectionBidirectional {
		return true
	}
	return mappingDirection == syncDirection
}

// getNestedValue extracts a nested value using dot notation
func (dm *DataMapper) getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			return current[part]
		}

		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

// setNestedValue sets a nested value using dot notation
func (dm *DataMapper) setNestedValue(data map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - set the value
			current[part] = value
			return
		}

		// Navigate deeper, creating maps as needed
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			next := make(map[string]interface{})
			current[part] = next
			current = next
		}
	}
}

// applyTransform applies a field transformation
func (dm *DataMapper) applyTransform(value interface{}, transform plugin.FieldTransform) (interface{}, error) {
	transformer, exists := dm.transformers[string(transform.Type)]
	if !exists {
		return value, fmt.Errorf("unknown transformer: %s", transform.Type)
	}

	return transformer.Transform(value, transform.Config)
}

// applyValidation applies validation rules to data
func (dm *DataMapper) applyValidation(data map[string]interface{}, validation plugin.FieldValidation) error {
	// For now, implement basic validation
	// In the future, this would use registered validators

	value, exists := data[validation.Field]
	if !exists {
		if required, ok := validation.Rules["required"].(bool); ok && required {
			return fmt.Errorf("required field missing: %s", validation.Field)
		}
		return nil
	}

	// Type validation
	if expectedType, ok := validation.Rules["type"].(string); ok {
		if !dm.validateType(value, expectedType) {
			return fmt.Errorf("field %s has invalid type, expected %s", validation.Field, expectedType)
		}
	}

	// Range validation for numeric fields
	if minVal, ok := validation.Rules["min"]; ok {
		if !dm.validateMin(value, minVal) {
			return fmt.Errorf("field %s is below minimum value %v", validation.Field, minVal)
		}
	}

	if maxVal, ok := validation.Rules["max"]; ok {
		if !dm.validateMax(value, maxVal) {
			return fmt.Errorf("field %s exceeds maximum value %v", validation.Field, maxVal)
		}
	}

	// Pattern validation for string fields
	if pattern, ok := validation.Rules["pattern"].(string); ok {
		if !dm.validatePattern(value, pattern) {
			return fmt.Errorf("field %s does not match pattern %s", validation.Field, pattern)
		}
	}

	return nil
}

// Built-in transformers

func (dm *DataMapper) registerBuiltinTransformers() {
	dm.transformers["map"] = &MapTransformer{}
	dm.transformers["format"] = &FormatTransformer{}
	dm.transformers["template"] = &TemplateTransformer{}
	dm.transformers["custom"] = &CustomTransformer{}
}

func (dm *DataMapper) registerBuiltinValidators() {
	dm.validators["required"] = &RequiredValidator{}
	dm.validators["type"] = &TypeValidator{}
	dm.validators["range"] = &RangeValidator{}
	dm.validators["pattern"] = &PatternValidator{}
}

// Validation helper methods

func (dm *DataMapper) validateType(value interface{}, expectedType string) bool {
	valueType := reflect.TypeOf(value).Kind().String()
	return strings.EqualFold(valueType, expectedType)
}

func (dm *DataMapper) validateMin(value interface{}, minVal interface{}) bool {
	switch v := value.(type) {
	case int, int32, int64:
		if minInt, ok := minVal.(int); ok {
			return reflect.ValueOf(v).Int() >= int64(minInt)
		}
	case float32, float64:
		if minFloat, ok := minVal.(float64); ok {
			return reflect.ValueOf(v).Float() >= minFloat
		}
	case string:
		if minLen, ok := minVal.(int); ok {
			return len(v) >= minLen
		}
	}
	return true
}

func (dm *DataMapper) validateMax(value interface{}, maxVal interface{}) bool {
	switch v := value.(type) {
	case int, int32, int64:
		if maxInt, ok := maxVal.(int); ok {
			return reflect.ValueOf(v).Int() <= int64(maxInt)
		}
	case float32, float64:
		if maxFloat, ok := maxVal.(float64); ok {
			return reflect.ValueOf(v).Float() <= maxFloat
		}
	case string:
		if maxLen, ok := maxVal.(int); ok {
			return len(v) <= maxLen
		}
	}
	return true
}

func (dm *DataMapper) validatePattern(value interface{}, pattern string) bool {
	if str, ok := value.(string); ok {
		if regex, err := regexp.Compile(pattern); err == nil {
			return regex.MatchString(str)
		}
	}
	return false
}

// Default mappings for different providers

func (dm *DataMapper) getJiraDefaultMapping() *plugin.FieldMappingConfig {
	return &plugin.FieldMappingConfig{
		Mappings: []plugin.FieldMapping{
			{ZenField: "id", ExternalField: "key", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "title", ExternalField: "fields.summary", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "description", ExternalField: "fields.description", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "status", ExternalField: "fields.status.name", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "priority", ExternalField: "fields.priority.name", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "assignee", ExternalField: "fields.assignee.displayName", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "type", ExternalField: "fields.issuetype.name", Direction: plugin.SyncDirectionBidirectional, Required: true},
		},
		Transforms: []plugin.FieldTransform{
			{
				Field:     "status",
				Type:      plugin.TransformTypeMap,
				Direction: plugin.SyncDirectionBidirectional,
				Config: map[string]interface{}{
					"mappings": map[string]string{
						"To Do": "proposed", "In Progress": "in_progress", "Done": "completed",
						"proposed": "To Do", "in_progress": "In Progress", "completed": "Done",
					},
				},
			},
		},
	}
}

func (dm *DataMapper) getGitHubDefaultMapping() *plugin.FieldMappingConfig {
	return &plugin.FieldMappingConfig{
		Mappings: []plugin.FieldMapping{
			{ZenField: "id", ExternalField: "number", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "title", ExternalField: "title", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "description", ExternalField: "body", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "status", ExternalField: "state", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "assignee", ExternalField: "assignee.login", Direction: plugin.SyncDirectionBidirectional},
		},
		Transforms: []plugin.FieldTransform{
			{
				Field:     "status",
				Type:      plugin.TransformTypeMap,
				Direction: plugin.SyncDirectionBidirectional,
				Config: map[string]interface{}{
					"mappings": map[string]string{
						"open": "proposed", "closed": "completed",
						"proposed": "open", "completed": "closed",
					},
				},
			},
		},
	}
}

func (dm *DataMapper) getLinearDefaultMapping() *plugin.FieldMappingConfig {
	return &plugin.FieldMappingConfig{
		Mappings: []plugin.FieldMapping{
			{ZenField: "id", ExternalField: "identifier", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "title", ExternalField: "title", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "description", ExternalField: "description", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "status", ExternalField: "state.name", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "priority", ExternalField: "priority", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "assignee", ExternalField: "assignee.displayName", Direction: plugin.SyncDirectionBidirectional},
		},
	}
}

func (dm *DataMapper) getGenericDefaultMapping() *plugin.FieldMappingConfig {
	return &plugin.FieldMappingConfig{
		Mappings: []plugin.FieldMapping{
			{ZenField: "id", ExternalField: "id", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "title", ExternalField: "title", Direction: plugin.SyncDirectionBidirectional, Required: true},
			{ZenField: "description", ExternalField: "description", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "status", ExternalField: "status", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "priority", ExternalField: "priority", Direction: plugin.SyncDirectionBidirectional},
			{ZenField: "assignee", ExternalField: "assignee", Direction: plugin.SyncDirectionBidirectional},
		},
	}
}

// Built-in transformers

// MapTransformer transforms values using a mapping table
type MapTransformer struct{}

func (mt *MapTransformer) Name() string { return "map" }

func (mt *MapTransformer) Transform(value interface{}, config map[string]interface{}) (interface{}, error) {
	mappings, ok := config["mappings"].(map[string]interface{})
	if !ok {
		return value, fmt.Errorf("mappings configuration required")
	}

	valueStr := fmt.Sprintf("%v", value)
	if mapped, exists := mappings[valueStr]; exists {
		return mapped, nil
	}

	// Return original value if no mapping found
	return value, nil
}

func (mt *MapTransformer) Validate(config map[string]interface{}) error {
	if _, ok := config["mappings"]; !ok {
		return fmt.Errorf("mappings configuration required")
	}
	return nil
}

// FormatTransformer formats values using format strings
type FormatTransformer struct{}

func (ft *FormatTransformer) Name() string { return "format" }

func (ft *FormatTransformer) Transform(value interface{}, config map[string]interface{}) (interface{}, error) {
	format, ok := config["format"].(string)
	if !ok {
		return value, fmt.Errorf("format configuration required")
	}

	return fmt.Sprintf(format, value), nil
}

func (ft *FormatTransformer) Validate(config map[string]interface{}) error {
	if _, ok := config["format"]; !ok {
		return fmt.Errorf("format configuration required")
	}
	return nil
}

// TemplateTransformer transforms values using Go templates
type TemplateTransformer struct{}

func (tt *TemplateTransformer) Name() string { return "template" }

func (tt *TemplateTransformer) Transform(value interface{}, config map[string]interface{}) (interface{}, error) {
	// Placeholder implementation
	// In production, this would use the existing template engine
	return value, nil
}

func (tt *TemplateTransformer) Validate(config map[string]interface{}) error {
	if _, ok := config["template"]; !ok {
		return fmt.Errorf("template configuration required")
	}
	return nil
}

// CustomTransformer allows custom transformation logic
type CustomTransformer struct{}

func (ct *CustomTransformer) Name() string { return "custom" }

func (ct *CustomTransformer) Transform(value interface{}, config map[string]interface{}) (interface{}, error) {
	// Custom transformations would be implemented here
	// For now, return the value unchanged
	return value, nil
}

func (ct *CustomTransformer) Validate(config map[string]interface{}) error {
	return nil
}

// Built-in validators

// RequiredValidator validates required fields
type RequiredValidator struct{}

func (rv *RequiredValidator) Name() string { return "required" }

func (rv *RequiredValidator) Validate(data map[string]interface{}, rules map[string]interface{}) error {
	for field, ruleValue := range rules {
		if required, ok := ruleValue.(bool); ok && required {
			if _, exists := data[field]; !exists {
				return fmt.Errorf("required field missing: %s", field)
			}
		}
	}
	return nil
}

// TypeValidator validates field types
type TypeValidator struct{}

func (tv *TypeValidator) Name() string { return "type" }

func (tv *TypeValidator) Validate(data map[string]interface{}, rules map[string]interface{}) error {
	for field, expectedType := range rules {
		if value, exists := data[field]; exists {
			if typeStr, ok := expectedType.(string); ok {
				valueType := reflect.TypeOf(value).Kind().String()
				if !strings.EqualFold(valueType, typeStr) {
					return fmt.Errorf("field %s has type %s, expected %s", field, valueType, typeStr)
				}
			}
		}
	}
	return nil
}

// RangeValidator validates numeric ranges
type RangeValidator struct{}

func (rv *RangeValidator) Name() string { return "range" }

func (rv *RangeValidator) Validate(data map[string]interface{}, rules map[string]interface{}) error {
	// Implementation would validate numeric ranges
	return nil
}

// PatternValidator validates string patterns
type PatternValidator struct{}

func (pv *PatternValidator) Name() string { return "pattern" }

func (pv *PatternValidator) Validate(data map[string]interface{}, rules map[string]interface{}) error {
	// Implementation would validate regex patterns
	return nil
}
