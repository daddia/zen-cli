package template

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/daddia/zen/internal/logging"
)

// DefaultVariableValidator implements VariableValidator interface
type DefaultVariableValidator struct {
	logger logging.Logger
}

// NewVariableValidator creates a new variable validator
func NewVariableValidator(logger logging.Logger) *DefaultVariableValidator {
	return &DefaultVariableValidator{
		logger: logger,
	}
}

// ValidateRequired validates that all required variables are present
func (v *DefaultVariableValidator) ValidateRequired(variables map[string]interface{}, specs []VariableSpec) []ValidationError {
	var errors []ValidationError

	for _, spec := range specs {
		if spec.Required {
			if value, exists := variables[spec.Name]; !exists || v.isEmpty(value) {
				errors = append(errors, ValidationError{
					Variable: spec.Name,
					Message:  fmt.Sprintf("required variable '%s' is missing or empty", spec.Name),
					Code:     string(ErrorCodeVariableRequired),
				})
			}
		}
	}

	v.logger.Debug("required variable validation completed", "errors", len(errors))
	return errors
}

// ValidateTypes validates variable types against specifications
func (v *DefaultVariableValidator) ValidateTypes(variables map[string]interface{}, specs []VariableSpec) []ValidationError {
	var errors []ValidationError

	for _, spec := range specs {
		value, exists := variables[spec.Name]
		if !exists {
			continue // Skip missing variables (handled by ValidateRequired)
		}

		if !v.isValidType(value, spec.Type) {
			actualType := v.getTypeName(value)
			errors = append(errors, ValidationError{
				Variable: spec.Name,
				Message:  fmt.Sprintf("variable '%s' has invalid type: expected %s, got %s", spec.Name, spec.Type, actualType),
				Code:     string(ErrorCodeVariableInvalid),
				Value:    fmt.Sprintf("%v", value),
			})
		}
	}

	v.logger.Debug("type validation completed", "errors", len(errors))
	return errors
}

// ValidateConstraints validates variable constraints (regex, ranges, etc.)
func (v *DefaultVariableValidator) ValidateConstraints(variables map[string]interface{}, specs []VariableSpec) []ValidationError {
	var errors []ValidationError

	for _, spec := range specs {
		value, exists := variables[spec.Name]
		if !exists {
			continue // Skip missing variables
		}

		if spec.Validation != "" {
			if err := v.validateConstraint(value, spec.Validation, spec.Name); err != nil {
				errors = append(errors, *err)
			}
		}
	}

	v.logger.Debug("constraint validation completed", "errors", len(errors))
	return errors
}

// ApplyDefaults applies default values for missing variables
func (v *DefaultVariableValidator) ApplyDefaults(variables map[string]interface{}, specs []VariableSpec) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing variables
	for k, val := range variables {
		result[k] = val
	}

	// Apply defaults for missing variables
	defaultsApplied := 0
	for _, spec := range specs {
		if _, exists := result[spec.Name]; !exists && spec.Default != nil {
			result[spec.Name] = spec.Default
			defaultsApplied++
		}
	}

	v.logger.Debug("defaults applied", "count", defaultsApplied)
	return result
}

// isEmpty checks if a value is empty
func (v *DefaultVariableValidator) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	val := reflect.ValueOf(value)
	switch val.Kind() {
	case reflect.String:
		return strings.TrimSpace(val.String()) == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return val.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return val.IsNil()
	default:
		return false
	}
}

// isValidType checks if a value matches the expected type
func (v *DefaultVariableValidator) isValidType(value interface{}, expectedType string) bool {
	if value == nil {
		return expectedType == "any" || expectedType == "interface{}"
	}

	actualType := v.getTypeName(value)

	// Handle type aliases and common variations
	switch strings.ToLower(expectedType) {
	case "string", "str":
		return actualType == "string"
	case "int", "integer":
		return actualType == "int" || actualType == "int64" || actualType == "int32"
	case "float", "float64", "number":
		return actualType == "float64" || actualType == "float32"
	case "bool", "boolean":
		return actualType == "bool"
	case "slice", "array", "list":
		return strings.Contains(actualType, "[]")
	case "map", "object":
		return strings.Contains(actualType, "map[")
	case "any", "interface{}", "interface":
		return true
	default:
		return actualType == expectedType
	}
}

// getTypeName returns the type name of a value
func (v *DefaultVariableValidator) getTypeName(value interface{}) string {
	if value == nil {
		return "nil"
	}
	return reflect.TypeOf(value).String()
}

// validateConstraint validates a single constraint
func (v *DefaultVariableValidator) validateConstraint(value interface{}, constraint, varName string) *ValidationError {
	// Parse constraint format: type:rule
	// Examples:
	// - regex:^[a-zA-Z0-9]+$
	// - range:1-100
	// - length:5-50
	// - enum:option1,option2,option3

	parts := strings.SplitN(constraint, ":", 2)
	if len(parts) != 2 {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("invalid constraint format: %s", constraint),
			Code:     string(ErrorCodeVariableInvalid),
		}
	}

	constraintType := strings.ToLower(strings.TrimSpace(parts[0]))
	rule := strings.TrimSpace(parts[1])

	switch constraintType {
	case "regex", "regexp":
		return v.validateRegex(value, rule, varName)
	case "range":
		return v.validateRange(value, rule, varName)
	case "length", "len":
		return v.validateLength(value, rule, varName)
	case "enum", "oneof":
		return v.validateEnum(value, rule, varName)
	case "min":
		return v.validateMin(value, rule, varName)
	case "max":
		return v.validateMax(value, rule, varName)
	default:
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("unsupported constraint type: %s", constraintType),
			Code:     string(ErrorCodeVariableInvalid),
		}
	}
}

// validateRegex validates a value against a regular expression
func (v *DefaultVariableValidator) validateRegex(value interface{}, pattern, varName string) *ValidationError {
	str := fmt.Sprintf("%v", value)

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("invalid regex pattern: %s", pattern),
			Code:     string(ErrorCodeVariableInvalid),
		}
	}

	if !regex.MatchString(str) {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("variable '%s' does not match pattern '%s'", varName, pattern),
			Code:     string(ErrorCodeVariableInvalid),
			Value:    str,
		}
	}

	return nil
}

// validateRange validates a numeric value is within a range
func (v *DefaultVariableValidator) validateRange(value interface{}, rule, varName string) *ValidationError {
	// Parse range: min-max
	parts := strings.Split(rule, "-")
	if len(parts) != 2 {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("invalid range format: %s (expected min-max)", rule),
			Code:     string(ErrorCodeVariableInvalid),
		}
	}

	min, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("invalid range minimum: %s", parts[0]),
			Code:     string(ErrorCodeVariableInvalid),
		}
	}

	max, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("invalid range maximum: %s", parts[1]),
			Code:     string(ErrorCodeVariableInvalid),
		}
	}

	// Convert value to float64
	var numValue float64
	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			numValue = parsed
		} else {
			return &ValidationError{
				Variable: varName,
				Message:  fmt.Sprintf("cannot validate range for non-numeric value: %v", value),
				Code:     string(ErrorCodeVariableInvalid),
				Value:    fmt.Sprintf("%v", value),
			}
		}
	default:
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("cannot validate range for non-numeric value: %v", value),
			Code:     string(ErrorCodeVariableInvalid),
			Value:    fmt.Sprintf("%v", value),
		}
	}

	if numValue < min || numValue > max {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("variable '%s' value %v is outside range %v-%v", varName, numValue, min, max),
			Code:     string(ErrorCodeVariableInvalid),
			Value:    fmt.Sprintf("%v", value),
		}
	}

	return nil
}

// validateLength validates the length of a string or collection
func (v *DefaultVariableValidator) validateLength(value interface{}, rule, varName string) *ValidationError {
	// Parse length: min-max or exact
	var min, max int
	var err error

	if strings.Contains(rule, "-") {
		parts := strings.Split(rule, "-")
		if len(parts) != 2 {
			return &ValidationError{
				Variable: varName,
				Message:  fmt.Sprintf("invalid length format: %s", rule),
				Code:     string(ErrorCodeVariableInvalid),
			}
		}
		min, err = strconv.Atoi(parts[0])
		if err != nil {
			return &ValidationError{
				Variable: varName,
				Message:  fmt.Sprintf("invalid length minimum: %s", parts[0]),
				Code:     string(ErrorCodeVariableInvalid),
			}
		}
		max, err = strconv.Atoi(parts[1])
		if err != nil {
			return &ValidationError{
				Variable: varName,
				Message:  fmt.Sprintf("invalid length maximum: %s", parts[1]),
				Code:     string(ErrorCodeVariableInvalid),
			}
		}
	} else {
		exact, err := strconv.Atoi(rule)
		if err != nil {
			return &ValidationError{
				Variable: varName,
				Message:  fmt.Sprintf("invalid length value: %s", rule),
				Code:     string(ErrorCodeVariableInvalid),
			}
		}
		min = exact
		max = exact
	}

	// Get length of value
	var length int
	switch v := value.(type) {
	case string:
		length = len(v)
	case []interface{}:
		length = len(v)
	default:
		val := reflect.ValueOf(value)
		if val.Kind() == reflect.Slice || val.Kind() == reflect.Array || val.Kind() == reflect.Map {
			length = val.Len()
		} else {
			return &ValidationError{
				Variable: varName,
				Message:  fmt.Sprintf("cannot validate length for type: %T", value),
				Code:     string(ErrorCodeVariableInvalid),
				Value:    fmt.Sprintf("%v", value),
			}
		}
	}

	if length < min || length > max {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("variable '%s' length %d is outside range %d-%d", varName, length, min, max),
			Code:     string(ErrorCodeVariableInvalid),
			Value:    fmt.Sprintf("%v", value),
		}
	}

	return nil
}

// validateEnum validates a value is one of the allowed options
func (v *DefaultVariableValidator) validateEnum(value interface{}, rule, varName string) *ValidationError {
	options := strings.Split(rule, ",")
	for i, option := range options {
		options[i] = strings.TrimSpace(option)
	}

	valueStr := fmt.Sprintf("%v", value)
	for _, option := range options {
		if valueStr == option {
			return nil
		}
	}

	return &ValidationError{
		Variable: varName,
		Message:  fmt.Sprintf("variable '%s' value '%s' is not one of allowed options: %s", varName, valueStr, strings.Join(options, ", ")),
		Code:     string(ErrorCodeVariableInvalid),
		Value:    valueStr,
	}
}

// validateMin validates a value is greater than or equal to minimum
func (v *DefaultVariableValidator) validateMin(value interface{}, rule, varName string) *ValidationError {
	min, err := strconv.ParseFloat(rule, 64)
	if err != nil {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("invalid minimum value: %s", rule),
			Code:     string(ErrorCodeVariableInvalid),
		}
	}

	// Convert value to float64
	var numValue float64
	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			numValue = parsed
		} else {
			return &ValidationError{
				Variable: varName,
				Message:  fmt.Sprintf("cannot validate minimum for non-numeric value: %v", value),
				Code:     string(ErrorCodeVariableInvalid),
				Value:    fmt.Sprintf("%v", value),
			}
		}
	default:
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("cannot validate minimum for non-numeric value: %v", value),
			Code:     string(ErrorCodeVariableInvalid),
			Value:    fmt.Sprintf("%v", value),
		}
	}

	if numValue < min {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("variable '%s' value %v is less than minimum %v", varName, numValue, min),
			Code:     string(ErrorCodeVariableInvalid),
			Value:    fmt.Sprintf("%v", value),
		}
	}

	return nil
}

// validateMax validates a value is less than or equal to maximum
func (v *DefaultVariableValidator) validateMax(value interface{}, rule, varName string) *ValidationError {
	max, err := strconv.ParseFloat(rule, 64)
	if err != nil {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("invalid maximum value: %s", rule),
			Code:     string(ErrorCodeVariableInvalid),
		}
	}

	// Convert value to float64
	var numValue float64
	switch v := value.(type) {
	case int:
		numValue = float64(v)
	case int64:
		numValue = float64(v)
	case float64:
		numValue = v
	case float32:
		numValue = float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			numValue = parsed
		} else {
			return &ValidationError{
				Variable: varName,
				Message:  fmt.Sprintf("cannot validate maximum for non-numeric value: %v", value),
				Code:     string(ErrorCodeVariableInvalid),
				Value:    fmt.Sprintf("%v", value),
			}
		}
	default:
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("cannot validate maximum for non-numeric value: %v", value),
			Code:     string(ErrorCodeVariableInvalid),
			Value:    fmt.Sprintf("%v", value),
		}
	}

	if numValue > max {
		return &ValidationError{
			Variable: varName,
			Message:  fmt.Sprintf("variable '%s' value %v is greater than maximum %v", varName, numValue, max),
			Code:     string(ErrorCodeVariableInvalid),
			Value:    fmt.Sprintf("%v", value),
		}
	}

	return nil
}
