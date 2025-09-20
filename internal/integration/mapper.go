package integration

import (
	"fmt"
	"reflect"
	"strings"
)

// DataMapper implements the DataMapperInterface
type DataMapper struct{}

// NewDataMapper creates a new data mapper
func NewDataMapper() *DataMapper {
	return &DataMapper{}
}

// MapFields maps fields between external and Zen formats using the provided mapping
func (m *DataMapper) MapFields(source map[string]interface{}, mapping map[string]string) (map[string]interface{}, error) {
	if source == nil {
		return nil, fmt.Errorf("source data cannot be nil")
	}

	if mapping == nil {
		return source, nil // Return source as-is if no mapping provided
	}

	result := make(map[string]interface{})

	for zenField, externalPath := range mapping {
		value, err := m.getNestedValue(source, externalPath)
		if err != nil {
			// Log the error but continue with other fields
			continue
		}
		result[zenField] = value
	}

	return result, nil
}

// ValidateMapping validates a field mapping configuration
func (m *DataMapper) ValidateMapping(mapping map[string]string) error {
	if mapping == nil {
		return fmt.Errorf("mapping cannot be nil")
	}

	requiredFields := []string{"task_id", "title"}
	for _, field := range requiredFields {
		if _, exists := mapping[field]; !exists {
			return fmt.Errorf("required field '%s' not found in mapping", field)
		}
	}

	// Validate mapping paths
	for zenField, externalPath := range mapping {
		if zenField == "" {
			return fmt.Errorf("zen field name cannot be empty")
		}
		if externalPath == "" {
			return fmt.Errorf("external path cannot be empty for field '%s'", zenField)
		}
	}

	return nil
}

// GetDefaultMapping returns the default field mapping for a provider
func (m *DataMapper) GetDefaultMapping(provider string) map[string]string {
	switch strings.ToLower(provider) {
	case "jira":
		return map[string]string{
			"task_id":     "key",
			"title":       "summary",
			"description": "description",
			"status":      "status.name",
			"priority":    "priority.name",
			"assignee":    "assignee.displayName",
			"created":     "created",
			"updated":     "updated",
		}
	case "github":
		return map[string]string{
			"task_id":     "number",
			"title":       "title",
			"description": "body",
			"status":      "state",
			"priority":    "labels.priority",
			"assignee":    "assignee.login",
			"created":     "created_at",
			"updated":     "updated_at",
		}
	default:
		// Generic mapping
		return map[string]string{
			"task_id":     "id",
			"title":       "title",
			"description": "description",
			"status":      "status",
			"priority":    "priority",
			"assignee":    "assignee",
			"created":     "created",
			"updated":     "updated",
		}
	}
}

// getNestedValue retrieves a nested value from a map using dot notation
func (m *DataMapper) getNestedValue(data map[string]interface{}, path string) (interface{}, error) {
	if path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if current == nil {
			return nil, fmt.Errorf("nil value encountered at path segment '%s'", part)
		}

		// Handle the last part
		if i == len(parts)-1 {
			if value, exists := current[part]; exists {
				return value, nil
			}
			return nil, fmt.Errorf("field '%s' not found", part)
		}

		// Navigate to the next level
		if value, exists := current[part]; exists {
			// Check if the value is a map for further navigation
			if nestedMap, ok := value.(map[string]interface{}); ok {
				current = nestedMap
			} else {
				// Try to convert to map using reflection
				if converted := m.convertToMap(value); converted != nil {
					current = converted
				} else {
					return nil, fmt.Errorf("cannot navigate through non-map value at '%s'", part)
				}
			}
		} else {
			return nil, fmt.Errorf("field '%s' not found in path '%s'", part, path)
		}
	}

	return nil, fmt.Errorf("unexpected end of path navigation")
}

// convertToMap tries to convert various types to map[string]interface{}
func (m *DataMapper) convertToMap(value interface{}) map[string]interface{} {
	if value == nil {
		return nil
	}

	// Already a map
	if mapValue, ok := value.(map[string]interface{}); ok {
		return mapValue
	}

	// Use reflection to handle struct types
	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !fieldValue.CanInterface() {
			continue
		}

		// Use json tag if available, otherwise use field name
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			if parts := strings.Split(jsonTag, ","); len(parts) > 0 && parts[0] != "" {
				fieldName = parts[0]
			}
		}

		// Convert field name to lowercase for consistency
		fieldName = strings.ToLower(fieldName)

		result[fieldName] = fieldValue.Interface()
	}

	return result
}

// ReverseMapping creates a reverse mapping (external -> zen becomes zen -> external)
func (m *DataMapper) ReverseMapping(mapping map[string]string) map[string]string {
	if mapping == nil {
		return nil
	}

	reversed := make(map[string]string)
	for zenField, externalPath := range mapping {
		reversed[externalPath] = zenField
	}

	return reversed
}

// MergeFields merges two field maps, with the second map taking precedence
func (m *DataMapper) MergeFields(base, override map[string]interface{}) map[string]interface{} {
	if base == nil && override == nil {
		return make(map[string]interface{})
	}
	if base == nil {
		return m.copyMap(override)
	}
	if override == nil {
		return m.copyMap(base)
	}

	result := m.copyMap(base)
	for k, v := range override {
		result[k] = v
	}

	return result
}

// copyMap creates a shallow copy of a map
func (m *DataMapper) copyMap(original map[string]interface{}) map[string]interface{} {
	if original == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{}, len(original))
	for k, v := range original {
		result[k] = v
	}

	return result
}
