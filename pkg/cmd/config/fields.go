package config

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/daddia/zen/internal/config"
)

// extractFieldValue extracts a field value from a configuration struct using reflection
func extractFieldValue(configStruct interface{}, fieldName string) (string, error) {
	v := reflect.ValueOf(configStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return "", fmt.Errorf("config must be a struct, got %T", configStruct)
	}
	
	// Convert field name to struct field name (snake_case to PascalCase)
	structFieldName := toPascalCase(fieldName)
	
	field := v.FieldByName(structFieldName)
	if !field.IsValid() {
		return "", fmt.Errorf("field %s not found in config", fieldName)
	}
	
	return formatFieldValue(field), nil
}

// updateConfigField updates a field in a configuration struct and returns the updated struct
func updateConfigField(configStruct interface{}, fieldName, value string) (interface{}, error) {
	// Create a copy of the struct
	v := reflect.ValueOf(configStruct)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("config must be a struct, got %T", configStruct)
	}
	
	// Create a new struct of the same type
	newStruct := reflect.New(v.Type()).Elem()
	newStruct.Set(v) // Copy all fields
	
	// Convert field name to struct field name
	structFieldName := toPascalCase(fieldName)
	
	field := newStruct.FieldByName(structFieldName)
	if !field.IsValid() {
		return nil, fmt.Errorf("field %s not found in config", fieldName)
	}
	
	if !field.CanSet() {
		return nil, fmt.Errorf("field %s cannot be set", fieldName)
	}
	
	// Set the field value based on its type
	if err := setFieldValue(field, value); err != nil {
		return nil, fmt.Errorf("failed to set field %s: %w", fieldName, err)
	}
	
	return newStruct.Interface(), nil
}

// formatFieldValue formats a reflect.Value as a string for display
func formatFieldValue(field reflect.Value) string {
	switch field.Kind() {
	case reflect.String:
		return field.String()
	case reflect.Bool:
		return strconv.FormatBool(field.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			return time.Duration(field.Int()).String()
		}
		return strconv.FormatInt(field.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(field.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(field.Float(), 'f', -1, 64)
	case reflect.Struct:
		// Handle nested structs (like DefaultDelims in template config)
		if field.Type().Name() == "" { // Anonymous struct
			// For now, return a simple representation
			return fmt.Sprintf("%+v", field.Interface())
		}
		return fmt.Sprintf("%+v", field.Interface())
	default:
		return fmt.Sprintf("%v", field.Interface())
	}
}

// setFieldValue sets a reflect.Value from a string value
func setFieldValue(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return fmt.Errorf("invalid boolean value: %s", value)
		}
		field.SetBool(boolVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if field.Type() == reflect.TypeOf(time.Duration(0)) {
			duration, err := time.ParseDuration(value)
			if err != nil {
				return fmt.Errorf("invalid duration value: %s", value)
			}
			field.SetInt(int64(duration))
		} else {
			intVal, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid integer value: %s", value)
			}
			field.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid unsigned integer value: %s", value)
		}
		field.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid float value: %s", value)
		}
		field.SetFloat(floatVal)
	default:
		return fmt.Errorf("unsupported field type: %s", field.Kind())
	}
	
	return nil
}

// toPascalCase converts snake_case to PascalCase
func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}

// getCoreConfigValue gets a value from the core config
func getCoreConfigValue(cfg *config.Config, field string) (string, error) {
	switch field {
	case "log_level":
		return cfg.LogLevel, nil
	case "log_format":
		return cfg.LogFormat, nil
	default:
		return "", fmt.Errorf("unknown core config field: %s", field)
	}
}

// setCoreConfigValue sets a value in the core config
func setCoreConfigValue(cfg *config.Config, field, value string) error {
	switch field {
	case "log_level":
		// Validate log level
		validLevels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}
		valid := false
		for _, level := range validLevels {
			if value == level {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log_level: %s (must be one of: %s)", value, strings.Join(validLevels, ", "))
		}
		cfg.LogLevel = value
	case "log_format":
		// Validate log format
		validFormats := []string{"text", "json"}
		valid := false
		for _, format := range validFormats {
			if value == format {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid log_format: %s (must be one of: %s)", value, strings.Join(validFormats, ", "))
		}
		cfg.LogFormat = value
	default:
		return fmt.Errorf("unknown core config field: %s", field)
	}
	
	return nil
}
