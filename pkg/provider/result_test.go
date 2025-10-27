package provider

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestResult_Success(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{
			name: "CLI success - exit code 0",
			result: Result{
				ExitCode: 0,
				Stdout:   []byte("success"),
			},
			expected: true,
		},
		{
			name: "CLI failure - exit code 1",
			result: Result{
				ExitCode: 1,
				Stderr:   []byte("error"),
			},
			expected: false,
		},
		{
			name: "API success - HTTP 200",
			result: Result{
				ExitCode: 200,
				Body:     []byte(`{"status": "ok"}`),
			},
			expected: true,
		},
		{
			name: "API success - HTTP 201",
			result: Result{
				ExitCode: 201,
				Body:     []byte(`{"id": "123"}`),
			},
			expected: true,
		},
		{
			name: "API client error - HTTP 400",
			result: Result{
				ExitCode: 400,
				Body:     []byte(`{"error": "bad request"}`),
			},
			expected: false,
		},
		{
			name: "API server error - HTTP 500",
			result: Result{
				ExitCode: 500,
				Body:     []byte(`{"error": "internal server error"}`),
			},
			expected: false,
		},
		{
			name: "API redirect - HTTP 302",
			result: Result{
				ExitCode: 302,
				Headers:  map[string]string{"Location": "https://example.com"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			success := tt.result.Success()

			// Assert
			assert.Equal(t, tt.expected, success)
		})
	}
}

func TestResult_IsClientError(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{
			name:     "HTTP 400 - bad request",
			result:   Result{ExitCode: 400},
			expected: true,
		},
		{
			name:     "HTTP 404 - not found",
			result:   Result{ExitCode: 404},
			expected: true,
		},
		{
			name:     "HTTP 403 - forbidden",
			result:   Result{ExitCode: 403},
			expected: true,
		},
		{
			name:     "HTTP 200 - success",
			result:   Result{ExitCode: 200},
			expected: false,
		},
		{
			name:     "HTTP 500 - server error",
			result:   Result{ExitCode: 500},
			expected: false,
		},
		{
			name:     "CLI exit code 1",
			result:   Result{ExitCode: 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isClientError := tt.result.IsClientError()

			// Assert
			assert.Equal(t, tt.expected, isClientError)
		})
	}
}

func TestResult_IsServerError(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected bool
	}{
		{
			name:     "HTTP 500 - internal server error",
			result:   Result{ExitCode: 500},
			expected: true,
		},
		{
			name:     "HTTP 502 - bad gateway",
			result:   Result{ExitCode: 502},
			expected: true,
		},
		{
			name:     "HTTP 503 - service unavailable",
			result:   Result{ExitCode: 503},
			expected: true,
		},
		{
			name:     "HTTP 200 - success",
			result:   Result{ExitCode: 200},
			expected: false,
		},
		{
			name:     "HTTP 400 - client error",
			result:   Result{ExitCode: 400},
			expected: false,
		},
		{
			name:     "CLI exit code 1",
			result:   Result{ExitCode: 1},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			isServerError := tt.result.IsServerError()

			// Assert
			assert.Equal(t, tt.expected, isServerError)
		})
	}
}

func TestResult_Output(t *testing.T) {
	tests := []struct {
		name     string
		result   Result
		expected []byte
	}{
		{
			name: "CLI result - returns stdout",
			result: Result{
				ExitCode: 0,
				Stdout:   []byte("command output"),
				Stderr:   []byte(""),
			},
			expected: []byte("command output"),
		},
		{
			name: "API result - returns body",
			result: Result{
				ExitCode: 200,
				Body:     []byte(`{"data": "value"}`),
			},
			expected: []byte(`{"data": "value"}`),
		},
		{
			name: "both body and stdout - prefers body",
			result: Result{
				ExitCode: 200,
				Stdout:   []byte("stdout content"),
				Body:     []byte("body content"),
			},
			expected: []byte("body content"),
		},
		{
			name: "empty result",
			result: Result{
				ExitCode: 0,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			output := tt.result.Output()

			// Assert
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestResult_CLIProvider(t *testing.T) {
	t.Run("successful CLI execution", func(t *testing.T) {
		// Arrange
		result := Result{
			ExitCode: 0,
			Stdout:   []byte("Cloning into 'repo'...\ndone."),
			Stderr:   []byte(""),
			Duration: 2 * time.Second,
		}

		// Assert
		assert.True(t, result.Success())
		assert.False(t, result.IsClientError())
		assert.False(t, result.IsServerError())
		assert.Equal(t, []byte("Cloning into 'repo'...\ndone."), result.Output())
		assert.Equal(t, 2*time.Second, result.Duration)
	})

	t.Run("failed CLI execution", func(t *testing.T) {
		// Arrange
		result := Result{
			ExitCode: 1,
			Stdout:   []byte(""),
			Stderr:   []byte("fatal: repository not found"),
			Duration: 100 * time.Millisecond,
		}

		// Assert
		assert.False(t, result.Success())
		assert.Equal(t, []byte(""), result.Output())
		assert.Equal(t, []byte("fatal: repository not found"), result.Stderr)
	})
}

func TestResult_APIProvider(t *testing.T) {
	t.Run("successful API request", func(t *testing.T) {
		// Arrange
		result := Result{
			ExitCode: 200,
			Body:     []byte(`{"items": [{"id": 1}, {"id": 2}]}`),
			Headers: map[string]string{
				"Content-Type":          "application/json",
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Remaining": "4999",
			},
			Meta: map[string]any{
				"rate_limit_remaining": 4999,
				"rate_limit_reset":     1234567890,
			},
			Duration: 150 * time.Millisecond,
		}

		// Assert
		assert.True(t, result.Success())
		assert.False(t, result.IsClientError())
		assert.False(t, result.IsServerError())
		assert.Equal(t, []byte(`{"items": [{"id": 1}, {"id": 2}]}`), result.Output())
		assert.Equal(t, "application/json", result.Headers["Content-Type"])
		assert.Equal(t, 4999, result.Meta["rate_limit_remaining"])
	})

	t.Run("API not found error", func(t *testing.T) {
		// Arrange
		result := Result{
			ExitCode: 404,
			Body:     []byte(`{"message": "Not Found"}`),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Duration: 50 * time.Millisecond,
		}

		// Assert
		assert.False(t, result.Success())
		assert.True(t, result.IsClientError())
		assert.False(t, result.IsServerError())
		assert.Equal(t, []byte(`{"message": "Not Found"}`), result.Output())
	})

	t.Run("API server error", func(t *testing.T) {
		// Arrange
		result := Result{
			ExitCode: 500,
			Body:     []byte(`{"message": "Internal Server Error"}`),
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Duration: 5 * time.Second,
		}

		// Assert
		assert.False(t, result.Success())
		assert.False(t, result.IsClientError())
		assert.True(t, result.IsServerError())
	})

	t.Run("API rate limit with metadata", func(t *testing.T) {
		// Arrange
		result := Result{
			ExitCode: 429,
			Body:     []byte(`{"message": "API rate limit exceeded"}`),
			Headers: map[string]string{
				"X-RateLimit-Remaining": "0",
				"Retry-After":           "3600",
			},
			Meta: map[string]any{
				"rate_limit_remaining": 0,
				"retry_after":          3600,
			},
			Duration: 100 * time.Millisecond,
		}

		// Assert
		assert.False(t, result.Success())
		assert.True(t, result.IsClientError())
		assert.Equal(t, 0, result.Meta["rate_limit_remaining"])
		assert.Equal(t, 3600, result.Meta["retry_after"])
	})
}

func TestResult_EdgeCases(t *testing.T) {
	t.Run("empty result", func(t *testing.T) {
		// Arrange
		result := Result{}

		// Assert
		assert.True(t, result.Success()) // ExitCode 0 is success
		assert.False(t, result.IsClientError())
		assert.False(t, result.IsServerError())
		assert.Nil(t, result.Output())
	})

	t.Run("result with only metadata", func(t *testing.T) {
		// Arrange
		result := Result{
			ExitCode: 200,
			Meta: map[string]any{
				"cached":    true,
				"cache_age": 3600,
			},
			Duration: 5 * time.Millisecond,
		}

		// Assert
		assert.True(t, result.Success())
		assert.True(t, result.Meta["cached"].(bool))
		assert.Equal(t, 3600, result.Meta["cache_age"])
	})

	t.Run("large output", func(t *testing.T) {
		// Arrange
		largeOutput := make([]byte, 1024*1024) // 1MB
		for i := range largeOutput {
			largeOutput[i] = byte(i % 256)
		}

		result := Result{
			ExitCode: 0,
			Stdout:   largeOutput,
			Duration: 10 * time.Second,
		}

		// Assert
		assert.True(t, result.Success())
		assert.Equal(t, 1024*1024, len(result.Output()))
	})
}
