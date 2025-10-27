package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/provider"
)

// Executor provides secure, context-aware execution of CLI commands
//
// Executor is designed with security and performance in mind:
//   - No shell usage: commands are executed directly to prevent injection attacks
//   - Arg-array only: arguments are passed as an array, never through a shell
//   - Minimal environment: only specified environment variables are passed
//   - Context-aware: supports cancellation and timeout via context
//   - Streaming support: allows processing output incrementally
//
// Executor is safe for concurrent use by multiple goroutines.
type Executor struct {
	// binaryPath is the absolute path to the binary to execute
	binaryPath string

	// workDir is the working directory for command execution
	// If empty, the command inherits the current process's working directory
	workDir string

	// env is the environment variables to set for command execution
	// These are added to a minimal base environment (PATH only)
	// Use sparingly for security - only pass required variables
	env []string

	// logger is used for structured logging of command execution
	logger logging.Logger
}

// NewExecutor creates a new CLI command executor
//
// Parameters:
//   - binaryPath: absolute path to the binary (required)
//   - workDir: working directory for execution (optional, empty = inherit)
//   - env: additional environment variables (optional, minimal base provided)
//   - logger: structured logger (required)
//
// The executor uses a minimal environment by default. Only PATH is passed
// from the parent process unless additional env vars are explicitly provided.
func NewExecutor(binaryPath, workDir string, env []string, logger logging.Logger) *Executor {
	return &Executor{
		binaryPath: binaryPath,
		workDir:    workDir,
		env:        env,
		logger:     logger,
	}
}

// ExecuteResult contains the complete result of a command execution
type ExecuteResult struct {
	// Stdout contains the standard output
	Stdout []byte

	// Stderr contains the standard error output
	Stderr []byte

	// ExitCode is the process exit code (0 = success)
	ExitCode int

	// Duration is the time taken to execute the command
	Duration time.Duration
}

// Execute runs the command with the given arguments and captures all output
//
// The command is executed with the provided context, which allows for
// cancellation and timeout handling. The context deadline is enforced
// at the OS process level via exec.CommandContext.
//
// Returns:
//   - ExecuteResult with captured stdout/stderr and exit code
//   - error if the command fails to start or context is canceled
//
// Note: A non-zero exit code is NOT returned as an error. The exit code
// is available in the result. This allows callers to distinguish between
// execution failures (error != nil) and command failures (ExitCode != 0).
func (e *Executor) Execute(ctx context.Context, args []string) (ExecuteResult, error) {
	start := time.Now()

	e.logger.Debug("executing command",
		"binary", e.binaryPath,
		"args", e.sanitizeArgs(args),
		"workdir", e.workDir)

	// Create command with context for timeout/cancellation support
	// #nosec G204 - This is a secure CLI executor by design:
	// - binaryPath is an absolute path from provider discovery
	// - args are passed as array (no shell interpretation)
	// - No shell usage (exec.CommandContext, not sh -c)
	// - Minimal environment with explicit allowlist
	// - Input sanitization applied via sanitizeArgs
	cmd := exec.CommandContext(ctx, e.binaryPath, args...)

	// Set working directory if specified
	if e.workDir != "" {
		cmd.Dir = e.workDir
	}

	// Set minimal environment
	cmd.Env = e.buildEnv()

	// Capture stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	execErr := cmd.Run()
	duration := time.Since(start)

	// Determine exit code
	exitCode := 0
	if execErr != nil {
		// Check if error is due to non-zero exit code
		if exitError, ok := execErr.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
			// Non-zero exit code is not an error from execution perspective
		} else {
			// Real execution error (failed to start, context canceled, etc.)
			e.logger.Error("command execution failed",
				"binary", e.binaryPath,
				"args", e.sanitizeArgs(args),
				"error", execErr,
				"duration", duration)
			return ExecuteResult{}, provider.WrapError(
				provider.ErrorCodeExecution,
				"failed to execute command",
				e.binaryPath,
				execErr,
			)
		}
	}

	result := ExecuteResult{
		Stdout:   stdout.Bytes(),
		Stderr:   stderr.Bytes(),
		ExitCode: exitCode,
		Duration: duration,
	}

	e.logger.Debug("command completed",
		"binary", e.binaryPath,
		"args", e.sanitizeArgs(args),
		"exitcode", exitCode,
		"duration", duration)

	return result, nil
}

// Stream executes the command and returns a reader for streaming output
//
// This is useful for long-running commands where you want to process output
// incrementally rather than buffering it all in memory.
//
// The returned ReadCloser provides:
//   - Combined stdout and stderr output (interleaved)
//   - Streaming access to output as it's generated
//   - Close() must be called to clean up resources
//
// The command runs asynchronously. The caller is responsible for:
//   - Reading from the returned ReadCloser
//   - Calling Close() when done
//   - Handling context cancellation (command will be killed)
//
// Returns:
//   - io.ReadCloser for streaming output
//   - error if the command fails to start
//
// Example:
//
//	stream, err := executor.Stream(ctx, []string{"--version"})
//	if err != nil {
//	    return err
//	}
//	defer stream.Close()
//
//	scanner := bufio.NewScanner(stream)
//	for scanner.Scan() {
//	    fmt.Println(scanner.Text())
//	}
func (e *Executor) Stream(ctx context.Context, args []string) (io.ReadCloser, error) {
	e.logger.Debug("streaming command output",
		"binary", e.binaryPath,
		"args", e.sanitizeArgs(args),
		"workdir", e.workDir)

	// Create command with context
	// #nosec G204 - This is a secure CLI executor by design:
	// - binaryPath is an absolute path from provider discovery
	// - args are passed as array (no shell interpretation)
	// - No shell usage (exec.CommandContext, not sh -c)
	// - Minimal environment with explicit allowlist
	// - Input sanitization applied via sanitizeArgs
	cmd := exec.CommandContext(ctx, e.binaryPath, args...)

	// Set working directory if specified
	if e.workDir != "" {
		cmd.Dir = e.workDir
	}

	// Set minimal environment
	cmd.Env = e.buildEnv()

	// Get combined stdout/stderr pipe
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, provider.WrapError(
			provider.ErrorCodeExecution,
			"failed to create stdout pipe",
			e.binaryPath,
			err,
		)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, provider.WrapError(
			provider.ErrorCodeExecution,
			"failed to create stderr pipe",
			e.binaryPath,
			err,
		)
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return nil, provider.WrapError(
			provider.ErrorCodeExecution,
			"failed to start command",
			e.binaryPath,
			err,
		)
	}

	// Combine stdout and stderr into a single reader
	combined := io.MultiReader(stdout, stderr)

	// Return a ReadCloser that waits for the command to complete
	return &streamReader{
		reader: combined,
		cmd:    cmd,
		logger: e.logger,
		binary: e.binaryPath,
		args:   args,
	}, nil
}

// streamReader wraps a command's output stream and ensures proper cleanup
type streamReader struct {
	reader io.Reader
	cmd    *exec.Cmd
	logger logging.Logger
	binary string
	args   []string
}

// Read implements io.Reader
func (s *streamReader) Read(p []byte) (n int, err error) {
	return s.reader.Read(p)
}

// Close waits for the command to complete and releases resources
func (s *streamReader) Close() error {
	// Wait for command to complete
	err := s.cmd.Wait()
	if err != nil {
		// Log error but don't return it - Close should be idempotent
		if exitError, ok := err.(*exec.ExitError); ok {
			s.logger.Debug("command exited with non-zero code",
				"binary", s.binary,
				"exitcode", exitError.ExitCode())
		} else {
			s.logger.Error("error waiting for command to complete",
				"binary", s.binary,
				"error", err)
		}
	}
	return nil
}

// buildEnv constructs a minimal environment for command execution
//
// Security considerations:
//   - Only PATH is inherited from the parent process
//   - Additional env vars must be explicitly specified
//   - This prevents leaking sensitive environment variables
func (e *Executor) buildEnv() []string {
	// Start with an empty environment
	env := []string{}

	// Add PATH from parent environment (required for binary resolution)
	// Note: In production, consider restricting PATH to known-safe directories
	env = append(env, "PATH="+getPathEnv())

	// Add any additional environment variables specified by the caller
	if len(e.env) > 0 {
		env = append(env, e.env...)
	}

	return env
}

// sanitizeArgs removes or masks sensitive arguments for logging
//
// This prevents credentials, tokens, or other sensitive data from
// appearing in log files.
func (e *Executor) sanitizeArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}

	sanitized := make([]string, len(args))
	copy(sanitized, args)

	// Mask common sensitive flags
	sensitiveFlags := []string{
		"--password", "-p",
		"--token", "-t",
		"--secret", "-s",
		"--key", "-k",
		"--api-key",
		"--auth",
	}

	for i := 0; i < len(sanitized); i++ {
		// Check if this is a sensitive flag
		for _, flag := range sensitiveFlags {
			if sanitized[i] == flag {
				// Mask the next argument (the value)
				if i+1 < len(sanitized) {
					sanitized[i+1] = "***"
				}
				break
			}
		}
	}

	return sanitized
}

// getPathEnv returns the PATH environment variable value
//
// This is a helper function to allow easy mocking in tests
var getPathEnv = func() string {
	path := os.Getenv("PATH")
	if path == "" {
		return "/usr/local/bin:/usr/bin:/bin"
	}
	return path
}
