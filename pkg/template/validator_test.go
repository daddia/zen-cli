package template

import (
	"testing"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewVariableValidator(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	assert.NotNil(t, validator)
	assert.Equal(t, logger, validator.logger)
}

func TestVariableValidator_ValidateRequired(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	specs := []VariableSpec{
		{
			Name:     "required_string",
			Required: true,
			Type:     "string",
		},
		{
			Name:     "optional_string",
			Required: false,
			Type:     "string",
		},
		{
			Name:     "required_int",
			Required: true,
			Type:     "int",
		},
	}

	tests := []struct {
		name           string
		variables      map[string]interface{}
		expectedErrors int
	}{
		{
			name: "all required variables present",
			variables: map[string]interface{}{
				"required_string": "test",
				"required_int":    42,
				"optional_string": "optional",
			},
			expectedErrors: 0,
		},
		{
			name: "missing one required variable",
			variables: map[string]interface{}{
				"required_string": "test",
				"optional_string": "optional",
			},
			expectedErrors: 1,
		},
		{
			name: "missing all required variables",
			variables: map[string]interface{}{
				"optional_string": "optional",
			},
			expectedErrors: 2,
		},
		{
			name: "empty string treated as missing",
			variables: map[string]interface{}{
				"required_string": "",
				"required_int":    42,
			},
			expectedErrors: 1,
		},
		{
			name: "nil value treated as missing",
			variables: map[string]interface{}{
				"required_string": nil,
				"required_int":    42,
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateRequired(tt.variables, specs)
			assert.Len(t, errors, tt.expectedErrors)

			for _, err := range errors {
				assert.Equal(t, string(ErrorCodeVariableRequired), err.Code)
				assert.Contains(t, err.Message, "required variable")
			}
		})
	}
}

func TestVariableValidator_ValidateTypes(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	specs := []VariableSpec{
		{Name: "string_var", Type: "string"},
		{Name: "int_var", Type: "int"},
		{Name: "bool_var", Type: "bool"},
		{Name: "float_var", Type: "float64"},
		{Name: "any_var", Type: "any"},
	}

	tests := []struct {
		name           string
		variables      map[string]interface{}
		expectedErrors int
	}{
		{
			name: "all correct types",
			variables: map[string]interface{}{
				"string_var": "test",
				"int_var":    42,
				"bool_var":   true,
				"float_var":  3.14,
				"any_var":    "anything",
			},
			expectedErrors: 0,
		},
		{
			name: "wrong string type",
			variables: map[string]interface{}{
				"string_var": 42,
			},
			expectedErrors: 1,
		},
		{
			name: "wrong int type",
			variables: map[string]interface{}{
				"int_var": "not an int",
			},
			expectedErrors: 1,
		},
		{
			name: "wrong bool type",
			variables: map[string]interface{}{
				"bool_var": "not a bool",
			},
			expectedErrors: 1,
		},
		{
			name: "any type accepts anything",
			variables: map[string]interface{}{
				"any_var": 42,
			},
			expectedErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateTypes(tt.variables, specs)
			assert.Len(t, errors, tt.expectedErrors)

			for _, err := range errors {
				assert.Equal(t, string(ErrorCodeVariableInvalid), err.Code)
				assert.Contains(t, err.Message, "invalid type")
			}
		})
	}
}

func TestVariableValidator_ValidateConstraints(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	specs := []VariableSpec{
		{
			Name:       "regex_var",
			Type:       "string",
			Validation: "regex:^[a-zA-Z0-9]+$",
		},
		{
			Name:       "range_var",
			Type:       "int",
			Validation: "range:1-100",
		},
		{
			Name:       "length_var",
			Type:       "string",
			Validation: "length:5-10",
		},
		{
			Name:       "enum_var",
			Type:       "string",
			Validation: "enum:red,green,blue",
		},
		{
			Name:       "min_var",
			Type:       "int",
			Validation: "min:10",
		},
		{
			Name:       "max_var",
			Type:       "int",
			Validation: "max:100",
		},
	}

	tests := []struct {
		name           string
		variables      map[string]interface{}
		expectedErrors int
	}{
		{
			name: "all constraints satisfied",
			variables: map[string]interface{}{
				"regex_var":  "test123",
				"range_var":  50,
				"length_var": "hello",
				"enum_var":   "red",
				"min_var":    20,
				"max_var":    80,
			},
			expectedErrors: 0,
		},
		{
			name: "regex constraint violated",
			variables: map[string]interface{}{
				"regex_var": "test-123!",
			},
			expectedErrors: 1,
		},
		{
			name: "range constraint violated",
			variables: map[string]interface{}{
				"range_var": 150,
			},
			expectedErrors: 1,
		},
		{
			name: "length constraint violated",
			variables: map[string]interface{}{
				"length_var": "hi",
			},
			expectedErrors: 1,
		},
		{
			name: "enum constraint violated",
			variables: map[string]interface{}{
				"enum_var": "yellow",
			},
			expectedErrors: 1,
		},
		{
			name: "min constraint violated",
			variables: map[string]interface{}{
				"min_var": 5,
			},
			expectedErrors: 1,
		},
		{
			name: "max constraint violated",
			variables: map[string]interface{}{
				"max_var": 150,
			},
			expectedErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateConstraints(tt.variables, specs)
			assert.Len(t, errors, tt.expectedErrors)

			for _, err := range errors {
				assert.Equal(t, string(ErrorCodeVariableInvalid), err.Code)
			}
		})
	}
}

func TestVariableValidator_ApplyDefaults(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	specs := []VariableSpec{
		{
			Name:    "with_default",
			Type:    "string",
			Default: "default_value",
		},
		{
			Name: "without_default",
			Type: "string",
		},
		{
			Name:    "int_default",
			Type:    "int",
			Default: 42,
		},
	}

	tests := []struct {
		name      string
		variables map[string]interface{}
		expected  map[string]interface{}
	}{
		{
			name: "apply defaults for missing variables",
			variables: map[string]interface{}{
				"existing_var": "existing",
			},
			expected: map[string]interface{}{
				"existing_var": "existing",
				"with_default": "default_value",
				"int_default":  42,
			},
		},
		{
			name: "don't override existing variables",
			variables: map[string]interface{}{
				"with_default": "custom_value",
				"int_default":  100,
			},
			expected: map[string]interface{}{
				"with_default": "custom_value",
				"int_default":  100,
			},
		},
		{
			name:      "empty variables map",
			variables: map[string]interface{}{},
			expected: map[string]interface{}{
				"with_default": "default_value",
				"int_default":  42,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ApplyDefaults(tt.variables, specs)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVariableValidator_IsEmpty(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"nil value", nil, true},
		{"empty string", "", true},
		{"whitespace string", "   ", true},
		{"non-empty string", "test", false},
		{"empty slice", []string{}, true},
		{"non-empty slice", []string{"item"}, false},
		{"empty map", map[string]string{}, true},
		{"non-empty map", map[string]string{"key": "value"}, false},
		{"zero int", 0, false},
		{"non-zero int", 42, false},
		{"false bool", false, false},
		{"true bool", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isEmpty(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVariableValidator_IsValidType(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	tests := []struct {
		name         string
		value        interface{}
		expectedType string
		expected     bool
	}{
		{"string matches string", "test", "string", true},
		{"string matches str", "test", "str", true},
		{"int matches int", 42, "int", true},
		{"int matches integer", 42, "integer", true},
		{"float64 matches float", 3.14, "float", true},
		{"float64 matches number", 3.14, "number", true},
		{"bool matches bool", true, "bool", true},
		{"bool matches boolean", true, "boolean", true},
		{"slice matches array", []string{"a", "b"}, "array", true},
		{"map matches object", map[string]string{"key": "value"}, "object", true},
		{"anything matches any", "anything", "any", true},
		{"nil matches any", nil, "any", true},
		{"wrong type", "string", "int", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.isValidType(tt.value, tt.expectedType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestVariableValidator_ValidateRegex(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	tests := []struct {
		name    string
		value   interface{}
		pattern string
		wantErr bool
	}{
		{"valid pattern match", "test123", "^[a-zA-Z0-9]+$", false},
		{"invalid pattern match", "test-123", "^[a-zA-Z0-9]+$", true},
		{"invalid regex pattern", "test", "[", true},
		{"number converted to string", 123, "^[0-9]+$", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRegex(tt.value, tt.pattern, "test_var")
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestVariableValidator_ValidateRange(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	tests := []struct {
		name    string
		value   interface{}
		rule    string
		wantErr bool
	}{
		{"valid int in range", 50, "1-100", false},
		{"valid float in range", 50.5, "1-100", false},
		{"int below range", 0, "1-100", true},
		{"int above range", 150, "1-100", true},
		{"string number in range", "50", "1-100", false},
		{"invalid string", "not a number", "1-100", true},
		{"invalid range format", 50, "1-", true},
		{"invalid range minimum", 50, "abc-100", true},
		{"invalid range maximum", 50, "1-abc", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateRange(tt.value, tt.rule, "test_var")
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestVariableValidator_ValidateLength(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	tests := []struct {
		name    string
		value   interface{}
		rule    string
		wantErr bool
	}{
		{"string length in range", "hello", "3-10", false},
		{"string length exact", "hello", "5", false},
		{"string too short", "hi", "3-10", true},
		{"string too long", "very long string", "3-10", true},
		{"slice length valid", []string{"a", "b", "c"}, "2-5", false},
		{"slice length invalid", []string{"a"}, "2-5", true},
		{"invalid length format", "test", "3-", true},
		{"non-string non-slice", 123, "3-10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateLength(tt.value, tt.rule, "test_var")
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestVariableValidator_ValidateEnum(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	tests := []struct {
		name    string
		value   interface{}
		rule    string
		wantErr bool
	}{
		{"valid enum value", "red", "red,green,blue", false},
		{"invalid enum value", "yellow", "red,green,blue", true},
		{"number enum", 1, "1,2,3", false},
		{"boolean enum", true, "true,false", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateEnum(tt.value, tt.rule, "test_var")
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestVariableValidator_ValidateMinMax(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	// Test min validation
	tests := []struct {
		name    string
		value   interface{}
		rule    string
		wantErr bool
	}{
		{"valid min", 50, "10", false},
		{"invalid min", 5, "10", true},
		{"string number valid min", "50", "10", false},
		{"invalid min rule", 50, "abc", true},
		{"non-numeric value", "not a number", "10", true},
	}

	for _, tt := range tests {
		t.Run("min_"+tt.name, func(t *testing.T) {
			err := validator.validateMin(tt.value, tt.rule, "test_var")
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})

		// Test max validation with correct logic
		t.Run("max_"+tt.name, func(t *testing.T) {
			err := validator.validateMax(tt.value, tt.rule, "test_var")
			// For max, we expect different logic than min
			switch tt.name {
			case "valid min":
				assert.NotNil(t, err) // 50 > 10, so max validation should fail
			case "invalid min":
				assert.Nil(t, err) // 5 <= 10, so max validation should pass
			case "string number valid min":
				assert.NotNil(t, err) // "50" > 10, so max validation should fail
			default:
				if tt.wantErr {
					assert.NotNil(t, err)
				} else {
					assert.Nil(t, err)
				}
			}
		})
	}
}

func TestVariableValidator_ValidateConstraint(t *testing.T) {
	logger := logging.NewBasic()
	validator := NewVariableValidator(logger)

	tests := []struct {
		name       string
		value      interface{}
		constraint string
		wantErr    bool
	}{
		{"valid regex constraint", "test123", "regex:^[a-zA-Z0-9]+$", false},
		{"valid range constraint", 50, "range:1-100", false},
		{"valid length constraint", "hello", "length:3-10", false},
		{"valid enum constraint", "red", "enum:red,green,blue", false},
		{"valid min constraint", 50, "min:10", false},
		{"valid max constraint", 50, "max:100", false},
		{"invalid constraint format", "test", "invalid", true},
		{"unsupported constraint type", "test", "unknown:value", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateConstraint(tt.value, tt.constraint, "test_var")
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
