package iostreams

import (
	"bytes"
	"io"
	"os"

	"github.com/mattn/go-isatty"
)

// IOStreams provides access to io streams
type IOStreams struct {
	In     io.ReadCloser
	Out    io.Writer
	ErrOut io.Writer

	colorEnabled   bool
	neverPrompt    bool
	progressWriter io.Writer
}

// System returns IOStreams connected to os.Stdin, os.Stdout, and os.Stderr
func System() *IOStreams {
	stdoutIsTTY := isatty.IsTerminal(os.Stdout.Fd())
	stderrIsTTY := isatty.IsTerminal(os.Stderr.Fd())

	return &IOStreams{
		In:           os.Stdin,
		Out:          os.Stdout,
		ErrOut:       os.Stderr,
		colorEnabled: stdoutIsTTY && stderrIsTTY && os.Getenv("NO_COLOR") == "",
	}
}

// Test returns IOStreams suitable for testing
func Test() *IOStreams {
	return &IOStreams{
		In:     io.NopCloser(&bytes.Buffer{}),
		Out:    &bytes.Buffer{},
		ErrOut: &bytes.Buffer{},
	}
}

// ColorEnabled returns true if color output is enabled
func (s *IOStreams) ColorEnabled() bool {
	return s.colorEnabled
}

// SetColorEnabled sets whether color output is enabled
func (s *IOStreams) SetColorEnabled(enabled bool) {
	s.colorEnabled = enabled
}

// IsStdinTTY returns true if stdin is a terminal
func (s *IOStreams) IsStdinTTY() bool {
	if f, ok := s.In.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}

// IsStdoutTTY returns true if stdout is a terminal
func (s *IOStreams) IsStdoutTTY() bool {
	if f, ok := s.Out.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}

// IsStderrTTY returns true if stderr is a terminal
func (s *IOStreams) IsStderrTTY() bool {
	if f, ok := s.ErrOut.(*os.File); ok {
		return isatty.IsTerminal(f.Fd())
	}
	return false
}

// SetNeverPrompt sets whether to never prompt for input
func (s *IOStreams) SetNeverPrompt(never bool) {
	s.neverPrompt = never
}

// CanPrompt returns true if prompting is allowed
func (s *IOStreams) CanPrompt() bool {
	return !s.neverPrompt && s.IsStdinTTY() && s.IsStdoutTTY()
}

// ProgressWriter returns the writer for progress output
func (s *IOStreams) ProgressWriter() io.Writer {
	if s.progressWriter != nil {
		return s.progressWriter
	}
	return s.ErrOut
}

// SetProgressWriter sets the writer for progress output
func (s *IOStreams) SetProgressWriter(w io.Writer) {
	s.progressWriter = w
}
