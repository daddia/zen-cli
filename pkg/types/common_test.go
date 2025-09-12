package types

import (
	"encoding/json"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      Error
		expected string
	}{
		{
			name: "error without details",
			err: Error{
				Code:    ErrorCodeInvalidInput,
				Message: "invalid input provided",
			},
			expected: "invalid input provided",
		},
		{
			name: "error with details",
			err: Error{
				Code:    ErrorCodeInvalidConfig,
				Message: "configuration is invalid",
				Details: "missing required field 'name'",
			},
			expected: "configuration is invalid: missing required field 'name'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestErrorJSON(t *testing.T) {
	err := Error{
		Code:    ErrorCodeNotFound,
		Message: "resource not found",
		Details: "user with ID 123 not found",
	}

	// Test JSON marshaling
	jsonData, marshalErr := json.Marshal(err)
	if marshalErr != nil {
		t.Errorf("Failed to marshal Error to JSON: %v", marshalErr)
	}

	// Test JSON unmarshaling
	var unmarshaled Error
	if unmarshalErr := json.Unmarshal(jsonData, &unmarshaled); unmarshalErr != nil {
		t.Errorf("Failed to unmarshal Error from JSON: %v", unmarshalErr)
	}

	if unmarshaled != err {
		t.Errorf("JSON round-trip failed: got %+v, want %+v", unmarshaled, err)
	}
}

func TestMetadata(t *testing.T) {
	metadata := Metadata{
		ID:   "test-id",
		Name: "test-name",
		Labels: map[string]string{
			"env":     "test",
			"version": "1.0.0",
		},
	}

	// Test JSON marshaling
	jsonData, err := json.Marshal(metadata)
	if err != nil {
		t.Errorf("Failed to marshal Metadata to JSON: %v", err)
	}

	var unmarshaled Metadata
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal Metadata from JSON: %v", err)
	}

	// Check basic fields
	if unmarshaled.ID != metadata.ID {
		t.Errorf("ID mismatch: got %q, want %q", unmarshaled.ID, metadata.ID)
	}

	if unmarshaled.Name != metadata.Name {
		t.Errorf("Name mismatch: got %q, want %q", unmarshaled.Name, metadata.Name)
	}

	// Check labels
	if len(unmarshaled.Labels) != len(metadata.Labels) {
		t.Errorf("Labels length mismatch: got %d, want %d", len(unmarshaled.Labels), len(metadata.Labels))
	}

	for key, value := range metadata.Labels {
		if unmarshaled.Labels[key] != value {
			t.Errorf("Label %q mismatch: got %q, want %q", key, unmarshaled.Labels[key], value)
		}
	}
}

func TestResult(t *testing.T) {
	// Test successful result
	successResult := Result{
		Status:  StatusCompleted,
		Message: "operation completed successfully",
		Data:    json.RawMessage(`{"key": "value"}`),
	}

	jsonData, err := json.Marshal(successResult)
	if err != nil {
		t.Errorf("Failed to marshal successful Result to JSON: %v", err)
	}

	var unmarshaled Result
	if err := json.Unmarshal(jsonData, &unmarshaled); err != nil {
		t.Errorf("Failed to unmarshal successful Result from JSON: %v", err)
	}

	if unmarshaled.Status != successResult.Status {
		t.Errorf("Status mismatch: got %q, want %q", unmarshaled.Status, successResult.Status)
	}

	// Test error result
	errorResult := Result{
		Status:  StatusFailed,
		Message: "operation failed",
		Error: &Error{
			Code:    ErrorCodeInvalidInput,
			Message: "invalid input",
		},
	}

	jsonData, err = json.Marshal(errorResult)
	if err != nil {
		t.Errorf("Failed to marshal error Result to JSON: %v", err)
	}

	var errorUnmarshaled Result
	if err := json.Unmarshal(jsonData, &errorUnmarshaled); err != nil {
		t.Errorf("Failed to unmarshal error Result from JSON: %v", err)
	}

	if errorUnmarshaled.Status != errorResult.Status {
		t.Errorf("Status mismatch: got %q, want %q", errorUnmarshaled.Status, errorResult.Status)
	}

	if errorUnmarshaled.Error == nil {
		t.Error("Error field is nil")
	} else if errorUnmarshaled.Error.Code != errorResult.Error.Code {
		t.Errorf("Error code mismatch: got %q, want %q", errorUnmarshaled.Error.Code, errorResult.Error.Code)
	}
}

func TestConstants(t *testing.T) {
	// Test ErrorCode constants
	errorCodes := []ErrorCode{
		ErrorCodeUnknown,
		ErrorCodeInvalidInput,
		ErrorCodeNotFound,
		ErrorCodeAlreadyExists,
		ErrorCodePermissionDenied,
		ErrorCodeTimeout,
		ErrorCodeInvalidConfig,
		ErrorCodeConfigNotFound,
		ErrorCodeWorkspaceNotInit,
		ErrorCodeInvalidWorkspace,
	}

	for _, code := range errorCodes {
		if string(code) == "" {
			t.Errorf("ErrorCode constant is empty")
		}
	}

	// Test Status constants
	statuses := []Status{
		StatusPending,
		StatusRunning,
		StatusCompleted,
		StatusFailed,
		StatusCancelled,
	}

	for _, status := range statuses {
		if string(status) == "" {
			t.Errorf("Status constant is empty")
		}
	}

	// Test Priority constants
	priorities := []Priority{
		PriorityLow,
		PriorityMedium,
		PriorityHigh,
		PriorityCritical,
	}

	for _, priority := range priorities {
		if string(priority) == "" {
			t.Errorf("Priority constant is empty")
		}
	}

	// Test OutputFormat constants
	formats := []OutputFormat{
		OutputFormatText,
		OutputFormatJSON,
		OutputFormatYAML,
	}

	for _, format := range formats {
		if string(format) == "" {
			t.Errorf("OutputFormat constant is empty")
		}
	}

	// Test LogLevel constants
	levels := []LogLevel{
		LogLevelTrace,
		LogLevelDebug,
		LogLevelInfo,
		LogLevelWarn,
		LogLevelError,
		LogLevelFatal,
		LogLevelPanic,
	}

	for _, level := range levels {
		if string(level) == "" {
			t.Errorf("LogLevel constant is empty")
		}
	}
}
