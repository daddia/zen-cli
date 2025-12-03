package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindOption(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		found bool
	}{
		{
			name:  "valid core.log_level",
			key:   "core.log_level",
			found: true,
		},
		{
			name:  "valid core.log_format",
			key:   "core.log_format",
			found: true,
		},
		{
			name:  "valid task.task_source",
			key:   "task.task_source",
			found: true,
		},
		{
			name:  "invalid key",
			key:   "invalid.key",
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt, found := FindOption(tt.key)
			assert.Equal(t, tt.found, found)
			if found {
				assert.Equal(t, tt.key, opt.Key)
				assert.NotEmpty(t, opt.Description)
			}
		})
	}
}

func TestValidateKey(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		wantError bool
	}{
		{
			name:      "valid core.log_level",
			key:       "core.log_level",
			wantError: false,
		},
		{
			name:      "valid core.log_format",
			key:       "core.log_format",
			wantError: false,
		},
		{
			name:      "valid task.task_source",
			key:       "task.task_source",
			wantError: false,
		},
		{
			name:      "invalid key",
			key:       "invalid.key",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateKey(tt.key)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateValue(t *testing.T) {
	tests := []struct {
		name      string
		key       string
		value     string
		wantError bool
	}{
		{
			name:      "valid core.log_level",
			key:       "core.log_level",
			value:     "debug",
			wantError: false,
		},
		{
			name:      "invalid core.log_level",
			key:       "core.log_level",
			value:     "invalid",
			wantError: true,
		},
		{
			name:      "valid core.log_format",
			key:       "core.log_format",
			value:     "json",
			wantError: false,
		},
		{
			name:      "invalid core.log_format",
			key:       "core.log_format",
			value:     "invalid",
			wantError: true,
		},
		{
			name:      "valid task.task_source",
			key:       "task.task_source",
			value:     "jira",
			wantError: false,
		},
		{
			name:      "invalid task.task_source",
			key:       "task.task_source",
			value:     "invalid",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateValue(tt.key, tt.value)
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestGetCurrentValue(t *testing.T) {
	cfg := &Config{
		Core: CoreConfig{
			LogLevel:  "debug",
			LogFormat: "json",
			Debug:     true,
		},
		Task: TaskConfig{
			TaskPath:   ".zen/tasks",
			TaskSource: "jira",
		},
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "core.log_level",
			key:      "core.log_level",
			expected: "debug",
		},
		{
			name:     "core.log_format",
			key:      "core.log_format",
			expected: "json",
		},
		{
			name:     "core.debug",
			key:      "core.debug",
			expected: "true",
		},
		{
			name:     "task.task_path",
			key:      "task.task_path",
			expected: ".zen/tasks",
		},
		{
			name:     "task.task_source",
			key:      "task.task_source",
			expected: "jira",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt, found := FindOption(tt.key)
			require.True(t, found)

			value := opt.GetCurrentValue(cfg)
			assert.Equal(t, tt.expected, value)
		})
	}
}
