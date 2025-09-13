package iostreams

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSystem(t *testing.T) {
	streams := System()

	assert.Equal(t, os.Stdin, streams.In)
	assert.Equal(t, os.Stdout, streams.Out)
	assert.Equal(t, os.Stderr, streams.ErrOut)
}

func TestTest(t *testing.T) {
	streams := Test()

	assert.NotNil(t, streams.In)
	assert.NotNil(t, streams.Out)
	assert.NotNil(t, streams.ErrOut)

	// Test streams should not be TTYs
	assert.False(t, streams.IsStdinTTY())
	assert.False(t, streams.IsStdoutTTY())
	assert.False(t, streams.IsStderrTTY())
}

func TestColorEnabled(t *testing.T) {
	streams := Test()

	// Default is false for test streams
	assert.False(t, streams.ColorEnabled())

	// Can be enabled
	streams.SetColorEnabled(true)
	assert.True(t, streams.ColorEnabled())

	// Can be disabled
	streams.SetColorEnabled(false)
	assert.False(t, streams.ColorEnabled())
}

func TestCanPrompt(t *testing.T) {
	t.Run("test streams cannot prompt", func(t *testing.T) {
		streams := Test()
		assert.False(t, streams.CanPrompt())
	})

	t.Run("never prompt setting", func(t *testing.T) {
		streams := Test()
		streams.SetNeverPrompt(true)
		assert.False(t, streams.CanPrompt())
	})
}

func TestProgressWriter(t *testing.T) {
	t.Run("defaults to ErrOut", func(t *testing.T) {
		streams := Test()
		assert.Equal(t, streams.ErrOut, streams.ProgressWriter())
	})

	t.Run("can be overridden", func(t *testing.T) {
		streams := Test()
		customWriter := &bytes.Buffer{}
		streams.SetProgressWriter(customWriter)
		assert.Equal(t, customWriter, streams.ProgressWriter())
	})
}
