package provider

import (
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKind_Constants(t *testing.T) {
	// Verify Kind constants are properly defined
	assert.Equal(t, Kind("cli"), KindCLI)
	assert.Equal(t, Kind("api"), KindAPI)
}

func TestInfo_Structure(t *testing.T) {
	tests := []struct {
		name string
		info Info
	}{
		{
			name: "cli provider info",
			info: Info{
				Name:       "git",
				Kind:       KindCLI,
				Version:    "2.39.0",
				Available:  true,
				Reason:     "",
				BinaryPath: "/usr/bin/git",
				BaseURL:    "",
				Capabilities: map[string]bool{
					"git.clone":  true,
					"git.commit": true,
					"git.push":   true,
				},
			},
		},
		{
			name: "api provider info",
			info: Info{
				Name:       "github",
				Kind:       KindAPI,
				Version:    "2022-11-28",
				Available:  true,
				Reason:     "",
				BinaryPath: "",
				BaseURL:    "https://api.github.com",
				Capabilities: map[string]bool{
					"pull_request.list":   true,
					"pull_request.create": true,
					"issue.list":          true,
				},
			},
		},
		{
			name: "unavailable provider info",
			info: Info{
				Name:         "terraform",
				Kind:         KindCLI,
				Version:      "",
				Available:    false,
				Reason:       "binary not found in PATH",
				BinaryPath:   "",
				BaseURL:      "",
				Capabilities: map[string]bool{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify all fields are accessible
			assert.NotEmpty(t, tt.info.Name)
			assert.NotEmpty(t, tt.info.Kind)

			if tt.info.Available {
				assert.Empty(t, tt.info.Reason)
			} else {
				assert.NotEmpty(t, tt.info.Reason)
			}

			// Verify kind-specific fields
			if tt.info.Kind == KindCLI && tt.info.Available {
				assert.NotEmpty(t, tt.info.BinaryPath)
				assert.Empty(t, tt.info.BaseURL)
			} else if tt.info.Kind == KindAPI && tt.info.Available {
				assert.Empty(t, tt.info.BinaryPath)
				assert.NotEmpty(t, tt.info.BaseURL)
			}
		})
	}
}

// MockProvider is a test implementation of the Provider interface
type MockProvider struct {
	InfoFunc    func(ctx context.Context) (Info, error)
	ExecuteFunc func(ctx context.Context, op string, params map[string]any) (Result, error)
	StreamFunc  func(ctx context.Context, op string, params map[string]any) (io.ReadCloser, error)
}

func (m *MockProvider) Info(ctx context.Context) (Info, error) {
	if m.InfoFunc != nil {
		return m.InfoFunc(ctx)
	}
	return Info{}, nil
}

func (m *MockProvider) Execute(ctx context.Context, op string, params map[string]any) (Result, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(ctx, op, params)
	}
	return Result{}, nil
}

func (m *MockProvider) Stream(ctx context.Context, op string, params map[string]any) (io.ReadCloser, error) {
	if m.StreamFunc != nil {
		return m.StreamFunc(ctx, op, params)
	}
	return io.NopCloser(nil), nil
}

func TestProvider_Interface(t *testing.T) {
	t.Run("info method", func(t *testing.T) {
		// Arrange
		expectedInfo := Info{
			Name:      "test-provider",
			Kind:      KindCLI,
			Version:   "1.0.0",
			Available: true,
		}

		provider := &MockProvider{
			InfoFunc: func(ctx context.Context) (Info, error) {
				return expectedInfo, nil
			},
		}

		// Act
		info, err := provider.Info(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedInfo.Name, info.Name)
		assert.Equal(t, expectedInfo.Kind, info.Kind)
		assert.Equal(t, expectedInfo.Version, info.Version)
		assert.Equal(t, expectedInfo.Available, info.Available)
	})

	t.Run("execute method", func(t *testing.T) {
		// Arrange
		expectedResult := Result{
			ExitCode: 0,
			Stdout:   []byte("success"),
		}

		provider := &MockProvider{
			ExecuteFunc: func(ctx context.Context, op string, params map[string]any) (Result, error) {
				assert.Equal(t, "test.op", op)
				assert.Equal(t, "value", params["key"])
				return expectedResult, nil
			},
		}

		// Act
		result, err := provider.Execute(context.Background(), "test.op", map[string]any{"key": "value"})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, expectedResult.ExitCode, result.ExitCode)
		assert.Equal(t, expectedResult.Stdout, result.Stdout)
	})

	t.Run("stream method", func(t *testing.T) {
		// Arrange
		provider := &MockProvider{
			StreamFunc: func(ctx context.Context, op string, params map[string]any) (io.ReadCloser, error) {
				assert.Equal(t, "test.stream", op)
				return io.NopCloser(nil), nil
			},
		}

		// Act
		stream, err := provider.Stream(context.Background(), "test.stream", map[string]any{})

		// Assert
		require.NoError(t, err)
		assert.NotNil(t, stream)
		defer stream.Close()
	})
}

func TestProvider_ContextCancellation(t *testing.T) {
	t.Run("context canceled during info", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		provider := &MockProvider{
			InfoFunc: func(ctx context.Context) (Info, error) {
				// Simulate context check
				if ctx.Err() != nil {
					return Info{}, ctx.Err()
				}
				return Info{}, nil
			},
		}

		// Act
		_, err := provider.Info(ctx)

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("context canceled during execute", func(t *testing.T) {
		// Arrange
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		provider := &MockProvider{
			ExecuteFunc: func(ctx context.Context, op string, params map[string]any) (Result, error) {
				// Simulate context check
				if ctx.Err() != nil {
					return Result{}, ctx.Err()
				}
				return Result{}, nil
			},
		}

		// Act
		_, err := provider.Execute(ctx, "test.op", map[string]any{})

		// Assert
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestProvider_ErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		setupProvider func() Provider
		operation     string
		params        map[string]any
		expectError   bool
	}{
		{
			name: "invalid operation",
			setupProvider: func() Provider {
				return &MockProvider{
					ExecuteFunc: func(ctx context.Context, op string, params map[string]any) (Result, error) {
						return Result{}, ErrInvalidOp("test-provider", op)
					},
				}
			},
			operation:   "unsupported.op",
			params:      map[string]any{},
			expectError: true,
		},
		{
			name: "execution failure",
			setupProvider: func() Provider {
				return &MockProvider{
					ExecuteFunc: func(ctx context.Context, op string, params map[string]any) (Result, error) {
						return Result{}, ErrExecution("test-provider", op, "command failed", nil)
					},
				}
			},
			operation:   "test.op",
			params:      map[string]any{},
			expectError: true,
		},
		{
			name: "successful execution",
			setupProvider: func() Provider {
				return &MockProvider{
					ExecuteFunc: func(ctx context.Context, op string, params map[string]any) (Result, error) {
						return Result{ExitCode: 0}, nil
					},
				}
			},
			operation:   "test.op",
			params:      map[string]any{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			provider := tt.setupProvider()

			// Act
			result, err := provider.Execute(context.Background(), tt.operation, tt.params)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}
