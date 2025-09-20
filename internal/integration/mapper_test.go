package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDataMapper(t *testing.T) {
	mapper := NewDataMapper()
	assert.NotNil(t, mapper)
}

func TestDataMapper_MapFields(t *testing.T) {
	mapper := NewDataMapper()

	source := map[string]interface{}{
		"key":     "PROJ-123",
		"summary": "Test Task",
		"status": map[string]interface{}{
			"name": "In Progress",
		},
		"priority": map[string]interface{}{
			"name": "High",
		},
	}

	mapping := map[string]string{
		"task_id":  "key",
		"title":    "summary",
		"status":   "status.name",
		"priority": "priority.name",
	}

	result, err := mapper.MapFields(source, mapping)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "PROJ-123", result["task_id"])
	assert.Equal(t, "Test Task", result["title"])
	assert.Equal(t, "In Progress", result["status"])
	assert.Equal(t, "High", result["priority"])
}

func TestDataMapper_MapFields_NilSource(t *testing.T) {
	mapper := NewDataMapper()

	mapping := map[string]string{
		"task_id": "key",
	}

	result, err := mapper.MapFields(nil, mapping)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "source data cannot be nil")
}

func TestDataMapper_MapFields_NilMapping(t *testing.T) {
	mapper := NewDataMapper()

	source := map[string]interface{}{
		"key": "PROJ-123",
	}

	result, err := mapper.MapFields(source, nil)

	assert.NoError(t, err)
	assert.Equal(t, source, result)
}

func TestDataMapper_ValidateMapping(t *testing.T) {
	mapper := NewDataMapper()

	tests := []struct {
		name        string
		mapping     map[string]string
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid mapping",
			mapping: map[string]string{
				"task_id": "key",
				"title":   "summary",
			},
			expectError: false,
		},
		{
			name:        "nil mapping",
			mapping:     nil,
			expectError: true,
			errorMsg:    "mapping cannot be nil",
		},
		{
			name: "missing required field task_id",
			mapping: map[string]string{
				"title": "summary",
			},
			expectError: true,
			errorMsg:    "required field 'task_id' not found",
		},
		{
			name: "missing required field title",
			mapping: map[string]string{
				"task_id": "key",
			},
			expectError: true,
			errorMsg:    "required field 'title' not found",
		},
		{
			name: "empty zen field",
			mapping: map[string]string{
				"task_id": "key",
				"title":   "summary",
				"":        "description",
			},
			expectError: true,
			errorMsg:    "zen field name cannot be empty",
		},
		{
			name: "empty external path",
			mapping: map[string]string{
				"task_id": "key",
				"title":   "",
			},
			expectError: true,
			errorMsg:    "external path cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapper.ValidateMapping(tt.mapping)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDataMapper_GetDefaultMapping(t *testing.T) {
	mapper := NewDataMapper()

	tests := []struct {
		provider string
		expected map[string]string
	}{
		{
			provider: "jira",
			expected: map[string]string{
				"task_id":     "key",
				"title":       "summary",
				"description": "description",
				"status":      "status.name",
				"priority":    "priority.name",
				"assignee":    "assignee.displayName",
				"created":     "created",
				"updated":     "updated",
			},
		},
		{
			provider: "github",
			expected: map[string]string{
				"task_id":     "number",
				"title":       "title",
				"description": "body",
				"status":      "state",
				"priority":    "labels.priority",
				"assignee":    "assignee.login",
				"created":     "created_at",
				"updated":     "updated_at",
			},
		},
		{
			provider: "unknown",
			expected: map[string]string{
				"task_id":     "id",
				"title":       "title",
				"description": "description",
				"status":      "status",
				"priority":    "priority",
				"assignee":    "assignee",
				"created":     "created",
				"updated":     "updated",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			result := mapper.GetDefaultMapping(tt.provider)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataMapper_getNestedValue(t *testing.T) {
	mapper := NewDataMapper()

	data := map[string]interface{}{
		"key": "PROJ-123",
		"status": map[string]interface{}{
			"name": "In Progress",
		},
		"user": map[string]interface{}{
			"profile": map[string]interface{}{
				"email": "test@example.com",
			},
		},
	}

	tests := []struct {
		name        string
		path        string
		expected    interface{}
		expectError bool
	}{
		{
			name:     "simple field",
			path:     "key",
			expected: "PROJ-123",
		},
		{
			name:     "nested field",
			path:     "status.name",
			expected: "In Progress",
		},
		{
			name:     "deeply nested field",
			path:     "user.profile.email",
			expected: "test@example.com",
		},
		{
			name:        "empty path",
			path:        "",
			expectError: true,
		},
		{
			name:        "nonexistent field",
			path:        "nonexistent",
			expectError: true,
		},
		{
			name:        "nonexistent nested field",
			path:        "status.nonexistent",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := mapper.getNestedValue(data, tt.path)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestDataMapper_ReverseMapping(t *testing.T) {
	mapper := NewDataMapper()

	original := map[string]string{
		"task_id": "key",
		"title":   "summary",
		"status":  "status.name",
	}

	expected := map[string]string{
		"key":         "task_id",
		"summary":     "title",
		"status.name": "status",
	}

	result := mapper.ReverseMapping(original)

	assert.Equal(t, expected, result)
}

func TestDataMapper_ReverseMapping_Nil(t *testing.T) {
	mapper := NewDataMapper()

	result := mapper.ReverseMapping(nil)

	assert.Nil(t, result)
}

func TestDataMapper_MergeFields(t *testing.T) {
	mapper := NewDataMapper()

	base := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	override := map[string]interface{}{
		"key2": "overridden_value2",
		"key3": "value3",
	}

	expected := map[string]interface{}{
		"key1": "value1",
		"key2": "overridden_value2",
		"key3": "value3",
	}

	result := mapper.MergeFields(base, override)

	assert.Equal(t, expected, result)
}

func TestDataMapper_MergeFields_NilMaps(t *testing.T) {
	mapper := NewDataMapper()

	tests := []struct {
		name     string
		base     map[string]interface{}
		override map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:     "both nil",
			base:     nil,
			override: nil,
			expected: map[string]interface{}{},
		},
		{
			name: "base nil",
			base: nil,
			override: map[string]interface{}{
				"key1": "value1",
			},
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
		{
			name: "override nil",
			base: map[string]interface{}{
				"key1": "value1",
			},
			override: nil,
			expected: map[string]interface{}{
				"key1": "value1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapper.MergeFields(tt.base, tt.override)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDataMapper_copyMap(t *testing.T) {
	mapper := NewDataMapper()

	original := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	copy := mapper.copyMap(original)

	// Should be equal but different instances
	assert.Equal(t, original, copy)

	// Modifying copy shouldn't affect original
	copy["key3"] = "value3"
	assert.NotContains(t, original, "key3")
	assert.Contains(t, copy, "key3")
}

func TestDataMapper_copyMap_Nil(t *testing.T) {
	mapper := NewDataMapper()

	result := mapper.copyMap(nil)

	assert.NotNil(t, result)
	assert.Equal(t, map[string]interface{}{}, result)
}
