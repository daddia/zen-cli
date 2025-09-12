package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		format    string
		wantLevel logrus.Level
	}{
		{
			name:      "debug level text format",
			level:     "debug",
			format:    "text",
			wantLevel: logrus.DebugLevel,
		},
		{
			name:      "info level json format",
			level:     "info",
			format:    "json",
			wantLevel: logrus.InfoLevel,
		},
		{
			name:      "invalid level defaults to info",
			level:     "invalid",
			format:    "text",
			wantLevel: logrus.InfoLevel,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := New(tt.level, tt.format)

			// Type assertion to check internal implementation
			if ll, ok := logger.(*LogrusLogger); ok {
				if ll.logger.Level != tt.wantLevel {
					t.Errorf("New() level = %v, want %v", ll.logger.Level, tt.wantLevel)
				}
			} else {
				t.Errorf("New() returned wrong type")
			}
		})
	}
}

func TestNewBasic(t *testing.T) {
	logger := NewBasic()

	if logger == nil {
		t.Error("NewBasic() returned nil")
	}

	// Type assertion to check internal implementation
	if ll, ok := logger.(*LogrusLogger); ok {
		if ll.logger.Level != logrus.InfoLevel {
			t.Errorf("NewBasic() level = %v, want %v", ll.logger.Level, logrus.InfoLevel)
		}
	}
}

func TestNewWithOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithOutput("info", "text", &buf)

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("NewWithOutput() did not write to custom output")
	}
}

func TestLogrusLogger_LogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithOutput("debug", "json", &buf)

	tests := []struct {
		name    string
		logFunc func(string, ...interface{})
		message string
		level   string
	}{
		{
			name:    "debug message",
			logFunc: logger.Debug,
			message: "debug test",
			level:   "debug",
		},
		{
			name:    "info message",
			logFunc: logger.Info,
			message: "info test",
			level:   "info",
		},
		{
			name:    "warn message",
			logFunc: logger.Warn,
			message: "warn test",
			level:   "warning",
		},
		{
			name:    "error message",
			logFunc: logger.Error,
			message: "error test",
			level:   "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc(tt.message, "key", "value")

			output := buf.String()
			if output == "" {
				t.Error("No output generated")
				return
			}

			// Parse JSON output
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
				t.Errorf("Failed to parse JSON log: %v", err)
				return
			}

			// Check message
			if msg, ok := logEntry["msg"]; !ok || msg != tt.message {
				t.Errorf("Expected message %q, got %q", tt.message, msg)
			}

			// Check level
			if level, ok := logEntry["level"]; !ok || level != tt.level {
				t.Errorf("Expected level %q, got %q", tt.level, level)
			}

			// Check custom field
			if key, ok := logEntry["key"]; !ok || key != "value" {
				t.Errorf("Expected key=value field, got key=%q", key)
			}
		})
	}
}

func TestLogrusLogger_WithField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithOutput("info", "json", &buf)

	fieldLogger := logger.WithField("component", "test")
	fieldLogger.Info("test message")

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Failed to parse JSON log: %v", err)
		return
	}

	if component, ok := logEntry["component"]; !ok || component != "test" {
		t.Errorf("Expected component=test field, got component=%q", component)
	}
}

func TestLogrusLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewWithOutput("info", "json", &buf)

	fields := map[string]interface{}{
		"component": "test",
		"version":   "1.0.0",
		"count":     42,
	}

	fieldLogger := logger.WithFields(fields)
	fieldLogger.Info("test message")

	output := buf.String()
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Errorf("Failed to parse JSON log: %v", err)
		return
	}

	for key, expectedValue := range fields {
		if actualValue, ok := logEntry[key]; !ok {
			t.Errorf("Expected field %q not found", key)
		} else {
			// Handle type conversion for numbers
			if key == "count" {
				if actualFloat, ok := actualValue.(float64); ok {
					actualValue = int(actualFloat)
				}
			}
			if actualValue != expectedValue {
				t.Errorf("Expected %s=%v, got %s=%v", key, expectedValue, key, actualValue)
			}
		}
	}
}

func TestParseFields(t *testing.T) {
	tests := []struct {
		name     string
		fields   []interface{}
		expected map[string]interface{}
	}{
		{
			name:     "empty fields",
			fields:   []interface{}{},
			expected: map[string]interface{}{},
		},
		{
			name:     "single key-value pair",
			fields:   []interface{}{"key", "value"},
			expected: map[string]interface{}{"key": "value"},
		},
		{
			name:     "multiple key-value pairs",
			fields:   []interface{}{"key1", "value1", "key2", 42},
			expected: map[string]interface{}{"key1": "value1", "key2": 42},
		},
		{
			name:     "odd number of fields (last ignored)",
			fields:   []interface{}{"key1", "value1", "key2"},
			expected: map[string]interface{}{"key1": "value1"},
		},
		{
			name:     "non-string key ignored",
			fields:   []interface{}{123, "value", "key2", "value2"},
			expected: map[string]interface{}{"key2": "value2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFields(tt.fields...)

			if len(result) != len(tt.expected) {
				t.Errorf("parseFields() length = %d, want %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("parseFields() missing key %q", key)
				} else if actualValue != expectedValue {
					t.Errorf("parseFields() %s = %v, want %v", key, actualValue, expectedValue)
				}
			}
		})
	}
}
