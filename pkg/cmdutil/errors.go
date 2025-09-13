package cmdutil

import "errors"

// ExitCode represents CLI exit codes
type ExitCode int

const (
	// ExitOK indicates successful completion
	ExitOK ExitCode = 0
	// ExitError indicates a general error
	ExitError ExitCode = 1
	// ExitCancel indicates user cancellation
	ExitCancel ExitCode = 2
	// ExitAuth indicates authentication failure
	ExitAuth ExitCode = 4
)

// Common errors
var (
	// SilentError is returned when an error should not be displayed
	SilentError = errors.New("silent error")
	// PendingError indicates an operation is pending
	PendingError = errors.New("pending error")
)

// FlagError represents a command flag error
type FlagError struct {
	Err error
}

func (e *FlagError) Error() string {
	return e.Err.Error()
}

// NoResultsError indicates no results were found
type NoResultsError struct {
	Message string
}

func (e NoResultsError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "no results found"
}

// IsUserCancellation checks if an error represents user cancellation
func IsUserCancellation(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "cancelled" || err.Error() == "interrupted"
}
