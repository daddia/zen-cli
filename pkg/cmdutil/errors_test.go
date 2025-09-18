package cmdutil

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExitCodes(t *testing.T) {
	assert.Equal(t, ExitCode(0), ExitOK)
	assert.Equal(t, ExitCode(1), ExitError)
	assert.Equal(t, ExitCode(2), ExitCancel)
	assert.Equal(t, ExitCode(4), ExitAuth)
}

func TestFlagError(t *testing.T) {
	baseErr := errors.New("invalid flag")
	flagErr := &FlagError{Err: baseErr}

	assert.Equal(t, "invalid flag", flagErr.Error())
}

func TestNoResultsError(t *testing.T) {
	t.Run("with custom message", func(t *testing.T) {
		err := NoResultsError{Message: "no items found"}
		assert.Equal(t, "no items found", err.Error())
	})

	t.Run("with default message", func(t *testing.T) {
		err := NoResultsError{}
		assert.Equal(t, "no results found", err.Error())
	})
}

func TestIsUserCancellation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "canceled error",
			err:      errors.New("canceled"),
			expected: true,
		},
		{
			name:     "interrupted error",
			err:      errors.New("interrupted"),
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("something else"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUserCancellation(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
