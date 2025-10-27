package cli

import (
	"context"
	"io"
	"testing"

	"github.com/daddia/zen/pkg/provider"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockCLIProvider is a test implementation of the CLIProvider interface
type MockCLIProvider struct {
	binaryPath string
	workDir    string
	env        []string
	infoFunc   func(ctx context.Context) (provider.Info, error)
	execFunc   func(ctx context.Context, op string, params map[string]any) (provider.Result, error)
	streamFunc func(ctx context.Context, op string, params map[string]any) (io.ReadCloser, error)
	argsFunc   func(op string, params map[string]any) ([]string, error)
}

func (m *MockCLIProvider) Info(ctx context.Context) (provider.Info, error) {
	if m.infoFunc != nil {
		return m.infoFunc(ctx)
	}
	return provider.Info{
		Name:       "mock",
		Kind:       provider.KindCLI,
		Version:    "1.0.0",
		Available:  true,
		BinaryPath: m.binaryPath,
	}, nil
}

func (m *MockCLIProvider) Execute(ctx context.Context, op string, params map[string]any) (provider.Result, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, op, params)
	}
	return provider.Result{ExitCode: 0}, nil
}

func (m *MockCLIProvider) Stream(ctx context.Context, op string, params map[string]any) (io.ReadCloser, error) {
	if m.streamFunc != nil {
		return m.streamFunc(ctx, op, params)
	}
	return io.NopCloser(nil), nil
}

func (m *MockCLIProvider) BinaryPath() string {
	return m.binaryPath
}

func (m *MockCLIProvider) ExecArgsFor(op string, params map[string]any) ([]string, error) {
	if m.argsFunc != nil {
		return m.argsFunc(op, params)
	}
	return []string{}, nil
}

func (m *MockCLIProvider) WorkDir() string {
	return m.workDir
}

func (m *MockCLIProvider) Env() []string {
	return m.env
}

func TestCLIProvider_Interface(t *testing.T) {
	t.Run("implements Provider interface", func(t *testing.T) {
		// Arrange
		var _ provider.Provider = (*MockCLIProvider)(nil)

		// This test ensures CLIProvider can be used as Provider
		mock := &MockCLIProvider{
			binaryPath: "/usr/bin/test",
			workDir:    "/tmp",
			env:        []string{"TEST=value"},
		}

		// Act & Assert
		info, err := mock.Info(context.Background())
		require.NoError(t, err)
		assert.Equal(t, "mock", info.Name)
		assert.Equal(t, provider.KindCLI, info.Kind)
	})

	t.Run("BinaryPath returns correct path", func(t *testing.T) {
		// Arrange
		expectedPath := "/usr/local/bin/git"
		mock := &MockCLIProvider{binaryPath: expectedPath}

		// Act
		path := mock.BinaryPath()

		// Assert
		assert.Equal(t, expectedPath, path)
	})

	t.Run("WorkDir returns correct directory", func(t *testing.T) {
		// Arrange
		expectedDir := "/home/user/repo"
		mock := &MockCLIProvider{workDir: expectedDir}

		// Act
		dir := mock.WorkDir()

		// Assert
		assert.Equal(t, expectedDir, dir)
	})

	t.Run("Env returns environment variables", func(t *testing.T) {
		// Arrange
		expectedEnv := []string{"GIT_AUTHOR=test", "PATH=/usr/bin"}
		mock := &MockCLIProvider{env: expectedEnv}

		// Act
		env := mock.Env()

		// Assert
		assert.Equal(t, expectedEnv, env)
	})

	t.Run("empty WorkDir", func(t *testing.T) {
		// Arrange
		mock := &MockCLIProvider{workDir: ""}

		// Act
		dir := mock.WorkDir()

		// Assert
		assert.Empty(t, dir)
	})

	t.Run("nil Env", func(t *testing.T) {
		// Arrange
		mock := &MockCLIProvider{env: nil}

		// Act
		env := mock.Env()

		// Assert
		assert.Nil(t, env)
	})
}

func TestCLIProvider_ExecArgsFor(t *testing.T) {
	t.Run("maps operation to arguments", func(t *testing.T) {
		// Arrange
		mock := &MockCLIProvider{
			argsFunc: func(op string, params map[string]any) ([]string, error) {
				switch op {
				case "test.echo":
					return []string{"echo", params["message"].(string)}, nil
				default:
					return nil, provider.ErrInvalidOp("test", op)
				}
			},
		}

		// Act
		args, err := mock.ExecArgsFor("test.echo", map[string]any{"message": "hello"})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, []string{"echo", "hello"}, args)
	})

	t.Run("returns error for invalid operation", func(t *testing.T) {
		// Arrange
		mock := &MockCLIProvider{
			argsFunc: func(op string, params map[string]any) ([]string, error) {
				return nil, provider.ErrInvalidOp("test", op)
			},
		}

		// Act
		_, err := mock.ExecArgsFor("invalid.op", map[string]any{})

		// Assert
		require.Error(t, err)
		assert.True(t, provider.IsProviderError(err))
		assert.Equal(t, provider.ErrorCodeInvalidOp, provider.GetErrorCode(err))
	})

	t.Run("validates required parameters", func(t *testing.T) {
		// Arrange
		mock := &MockCLIProvider{
			argsFunc: func(op string, params map[string]any) ([]string, error) {
				if op == "test.clone" {
					url, ok := params["url"].(string)
					if !ok || url == "" {
						return nil, provider.NewError(
							provider.ErrorCodeExecution,
							"url parameter required",
							"test",
						)
					}
					return []string{"clone", url}, nil
				}
				return nil, provider.ErrInvalidOp("test", op)
			},
		}

		// Act
		_, err := mock.ExecArgsFor("test.clone", map[string]any{})

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "url parameter required")
	})

	t.Run("handles complex parameter mapping", func(t *testing.T) {
		// Arrange
		mock := &MockCLIProvider{
			argsFunc: func(op string, params map[string]any) ([]string, error) {
				if op == "test.command" {
					args := []string{"command"}

					if flag, ok := params["flag"].(bool); ok && flag {
						args = append(args, "--flag")
					}

					if value, ok := params["value"].(string); ok && value != "" {
						args = append(args, "--value", value)
					}

					if items, ok := params["items"].([]string); ok {
						args = append(args, items...)
					}

					return args, nil
				}
				return nil, provider.ErrInvalidOp("test", op)
			},
		}

		// Act
		args, err := mock.ExecArgsFor("test.command", map[string]any{
			"flag":  true,
			"value": "test",
			"items": []string{"a", "b", "c"},
		})

		// Assert
		require.NoError(t, err)
		assert.Equal(t, []string{"command", "--flag", "--value", "test", "a", "b", "c"}, args)
	})
}

func TestCLIProvider_Integration(t *testing.T) {
	t.Run("full provider workflow", func(t *testing.T) {
		// Arrange
		mock := &MockCLIProvider{
			binaryPath: "/usr/bin/test",
			workDir:    "/tmp/workdir",
			env:        []string{"TEST_ENV=value"},
			infoFunc: func(ctx context.Context) (provider.Info, error) {
				return provider.Info{
					Name:       "test",
					Kind:       provider.KindCLI,
					Version:    "1.2.3",
					Available:  true,
					BinaryPath: "/usr/bin/test",
					Capabilities: map[string]bool{
						"test.echo":    true,
						"test.command": true,
					},
				}, nil
			},
			execFunc: func(ctx context.Context, op string, params map[string]any) (provider.Result, error) {
				return provider.Result{
					ExitCode: 0,
					Stdout:   []byte("test output"),
				}, nil
			},
			argsFunc: func(op string, params map[string]any) ([]string, error) {
				if op == "test.echo" {
					return []string{"echo", params["message"].(string)}, nil
				}
				return nil, provider.ErrInvalidOp("test", op)
			},
		}

		// Act - Get info
		info, err := mock.Info(context.Background())
		require.NoError(t, err)

		// Assert info
		assert.Equal(t, "test", info.Name)
		assert.Equal(t, provider.KindCLI, info.Kind)
		assert.Equal(t, "1.2.3", info.Version)
		assert.True(t, info.Available)
		assert.Equal(t, "/usr/bin/test", info.BinaryPath)
		assert.True(t, info.Capabilities["test.echo"])

		// Act - Execute operation
		result, err := mock.Execute(context.Background(), "test.echo", map[string]any{"message": "hello"})
		require.NoError(t, err)

		// Assert execution
		assert.Equal(t, 0, result.ExitCode)
		assert.Equal(t, []byte("test output"), result.Stdout)

		// Act - Get args for operation
		args, err := mock.ExecArgsFor("test.echo", map[string]any{"message": "hello"})
		require.NoError(t, err)

		// Assert args
		assert.Equal(t, []string{"echo", "hello"}, args)

		// Act - Get CLI-specific properties
		assert.Equal(t, "/usr/bin/test", mock.BinaryPath())
		assert.Equal(t, "/tmp/workdir", mock.WorkDir())
		assert.Equal(t, []string{"TEST_ENV=value"}, mock.Env())
	})
}

func TestCLIProvider_TypeAssertion(t *testing.T) {
	t.Run("can type assert from Provider to CLIProvider", func(t *testing.T) {
		// Arrange
		var baseProvider provider.Provider = &MockCLIProvider{
			binaryPath: "/usr/bin/git",
		}

		// Act
		cliProvider, ok := baseProvider.(CLIProvider)

		// Assert
		assert.True(t, ok)
		assert.NotNil(t, cliProvider)
		assert.Equal(t, "/usr/bin/git", cliProvider.BinaryPath())
	})

	t.Run("can use Provider methods on CLIProvider", func(t *testing.T) {
		// Arrange
		cliProvider := &MockCLIProvider{
			binaryPath: "/usr/bin/test",
		}

		// Act - Use Provider interface methods
		info, err := cliProvider.Info(context.Background())

		// Assert
		require.NoError(t, err)
		assert.Equal(t, "mock", info.Name)
		assert.Equal(t, provider.KindCLI, info.Kind)
	})
}
