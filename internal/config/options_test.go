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
			name:  "valid log_level",
			key:   "log_level",
			found: true,
		},
		{
			name:  "valid log_format",
			key:   "log_format",
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
			name:      "valid log_level",
			key:       "log_level",
			wantError: false,
		},
		{
			name:      "valid log_format",
			key:       "log_format",
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
			name:      "valid log_level",
			key:       "log_level",
			value:     "debug",
			wantError: false,
		},
		{
			name:      "invalid log_level",
			key:       "log_level",
			value:     "invalid",
			wantError: true,
		},
		{
			name:      "valid log_format",
			key:       "log_format",
			value:     "json",
			wantError: false,
		},
		{
			name:      "invalid log_format",
			key:       "log_format",
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
		LogLevel:  "debug",
		LogFormat: "json",
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "log_level",
			key:      "log_level",
			expected: "debug",
		},
		{
			name:     "log_format",
			key:      "log_format",
			expected: "json",
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
