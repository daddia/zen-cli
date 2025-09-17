package internal

import (
	"fmt"
	"os"
	"strings"

	"github.com/daddia/zen/pkg/iostreams"
)

// ColorScheme provides color formatting functions for asset commands
type ColorScheme struct {
	io *iostreams.IOStreams
}

// NewColorScheme creates a new color scheme
func NewColorScheme(io *iostreams.IOStreams) *ColorScheme {
	return &ColorScheme{io: io}
}

// Bold formats text as bold
func (cs *ColorScheme) Bold(text string) string {
	return cs.io.ColorBold(text)
}

// Green formats text in green (success)
func (cs *ColorScheme) Green(text string) string {
	return cs.io.ColorSuccess(text)
}

// Red formats text in red (error)
func (cs *ColorScheme) Red(text string) string {
	return cs.io.ColorError(text)
}

// Yellow formats text in yellow (warning)
func (cs *ColorScheme) Yellow(text string) string {
	return cs.io.ColorWarning(text)
}

// Blue formats text in blue (info)
func (cs *ColorScheme) Blue(text string) string {
	return cs.io.ColorInfo(text)
}

// Magenta formats text in magenta
func (cs *ColorScheme) Magenta(text string) string {
	if cs.io.ColorEnabled() {
		return "\033[35m" + text + "\033[0m"
	}
	return text
}

// Gray formats text in gray (neutral/secondary)
func (cs *ColorScheme) Gray(text string) string {
	return cs.io.ColorNeutral(text)
}

// SuccessIcon returns a success icon
func (cs *ColorScheme) SuccessIcon() string {
	return cs.Green("✓")
}

// ErrorIcon returns an error icon
func (cs *ColorScheme) ErrorIcon() string {
	return cs.Red("✗")
}

// WarningIcon returns a warning icon
func (cs *ColorScheme) WarningIcon() string {
	return cs.Yellow("⚠")
}

// InfoIcon returns an info icon
func (cs *ColorScheme) InfoIcon() string {
	return cs.Blue("ℹ")
}

// PromptForPassword prompts the user for a password input
func PromptForPassword(io *iostreams.IOStreams, prompt string) (string, error) {
	if !io.CanPrompt() {
		return "", fmt.Errorf("prompting is disabled")
	}

	fmt.Fprint(io.Out, prompt+": ")

	// For now, use a simple approach - in a real implementation,
	// this would use a proper terminal library to hide input
	var password string
	_, err := fmt.Fscanln(io.In, &password)
	if err != nil {
		return "", err
	}

	return password, nil
}

// GetEnvVar gets an environment variable
func GetEnvVar(name string) string {
	return os.Getenv(name)
}

// Capitalize capitalizes the first letter of a string
func Capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + strings.ToLower(s[1:])
}
