package iostreams

import (
	"fmt"
	"strings"
)

// ANSI color codes following the design guide's 8 basic colors
const (
	// Basic ANSI colors (design guide compliant)
	ColorReset  = "\033[0m"
	ColorBold   = "\033[1m"
	
	// Foreground colors
	ColorBlack   = "\033[30m"
	ColorRed     = "\033[31m"
	ColorGreen   = "\033[32m"
	ColorYellow  = "\033[33m"
	ColorBlue    = "\033[34m"
	ColorMagenta = "\033[35m"
	ColorCyan    = "\033[36m"
	ColorWhite   = "\033[37m"
	
	// Bright variants (less reliable but available)
	ColorBrightBlack   = "\033[90m"
	ColorBrightRed     = "\033[91m"
	ColorBrightGreen   = "\033[92m"
	ColorBrightYellow  = "\033[93m"
	ColorBrightBlue    = "\033[94m"
	ColorBrightMagenta = "\033[95m"
	ColorBrightCyan    = "\033[96m"
	ColorBrightWhite   = "\033[97m"
)

// Unicode symbols following the design guide
const (
	SymbolSuccess = "✓"  // Success
	SymbolNeutral = "-"  // Neutral
	SymbolFailure = "✗"  // Failure
	SymbolAlert   = "!"  // Alert
	SymbolChange  = "+"  // Changes requested
)

// ColorFunc represents a function that applies color to text
type ColorFunc func(string) string

// Color functions for semantic meaning (following design guide)
func (s *IOStreams) ColorSuccess(text string) string {
	if !s.ColorEnabled() {
		return text
	}
	return ColorGreen + text + ColorReset
}

func (s *IOStreams) ColorError(text string) string {
	if !s.ColorEnabled() {
		return text
	}
	return ColorRed + text + ColorReset
}

func (s *IOStreams) ColorWarning(text string) string {
	if !s.ColorEnabled() {
		return text
	}
	return ColorYellow + text + ColorReset
}

func (s *IOStreams) ColorInfo(text string) string {
	if !s.ColorEnabled() {
		return text
	}
	return ColorBlue + text + ColorReset
}

func (s *IOStreams) ColorBold(text string) string {
	if !s.ColorEnabled() {
		return text
	}
	return ColorBold + text + ColorReset
}

func (s *IOStreams) ColorNeutral(text string) string {
	if !s.ColorEnabled() {
		return text
	}
	return ColorCyan + text + ColorReset
}

// Status formatting functions with symbols and colors
func (s *IOStreams) FormatSuccess(text string) string {
	if s.ColorEnabled() {
		return s.ColorSuccess(SymbolSuccess + " " + text)
	}
	return SymbolSuccess + " " + text
}

func (s *IOStreams) FormatError(text string) string {
	if s.ColorEnabled() {
		return s.ColorError(SymbolFailure + " " + text)
	}
	return SymbolFailure + " " + text
}

func (s *IOStreams) FormatWarning(text string) string {
	if s.ColorEnabled() {
		return s.ColorWarning(SymbolAlert + " " + text)
	}
	return SymbolAlert + " " + text
}

func (s *IOStreams) FormatNeutral(text string) string {
	if s.ColorEnabled() {
		return s.ColorNeutral(SymbolNeutral + " " + text)
	}
	return SymbolNeutral + " " + text
}

func (s *IOStreams) FormatChange(text string) string {
	if s.ColorEnabled() {
		return s.ColorWarning(SymbolChange + " " + text)
	}
	return SymbolChange + " " + text
}

// Header formatting for consistent typography
func (s *IOStreams) FormatHeader(text string) string {
	if s.ColorEnabled() {
		return s.ColorBold(text)
	}
	return text
}

func (s *IOStreams) FormatSectionHeader(text string) string {
	header := s.FormatHeader(text)
	underline := strings.Repeat("=", len(text))
	if s.ColorEnabled() {
		underline = s.ColorBold(underline)
	}
	return header + "\n" + underline
}

func (s *IOStreams) FormatSubHeader(text string) string {
	header := s.FormatHeader(text)
	underline := strings.Repeat("-", len(text))
	if s.ColorEnabled() {
		underline = s.ColorNeutral(underline)
	}
	return header + "\n" + underline
}

// Indentation helpers for proper spacing
func (s *IOStreams) Indent(text string, level int) string {
	indent := strings.Repeat("  ", level) // 2 spaces per level
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

// Table formatting helpers
func (s *IOStreams) FormatTable(headers []string, rows [][]string) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}
	
	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}
	
	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}
	
	var result strings.Builder
	
	// Format headers
	for i, header := range headers {
		if i > 0 {
			result.WriteString("  ") // Tab-like spacing
		}
		result.WriteString(s.FormatBold(fmt.Sprintf("%-*s", colWidths[i], header)))
	}
	result.WriteString("\n")
	
	// Format rows
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				result.WriteString("  ") // Tab-like spacing
			}
			if i < len(colWidths) {
				result.WriteString(fmt.Sprintf("%-*s", colWidths[i], cell))
			} else {
				result.WriteString(cell)
			}
		}
		result.WriteString("\n")
	}
	
	return result.String()
}

// Machine-readable output (for scriptability)
func (s *IOStreams) FormatMachineTable(headers []string, rows [][]string) string {
	if len(rows) == 0 {
		return ""
	}
	
	var result strings.Builder
	
	// No headers in machine output
	// Use tabs as delimiters for cut compatibility
	for _, row := range rows {
		for i, cell := range row {
			if i > 0 {
				result.WriteString("\t")
			}
			result.WriteString(cell)
		}
		result.WriteString("\n")
	}
	
	return result.String()
}

// Format status with appropriate color and symbol
func (s *IOStreams) FormatStatus(status string, isSuccess bool) string {
	if isSuccess {
		return s.FormatSuccess(status)
	}
	return s.FormatError(status)
}

// Format boolean status
func (s *IOStreams) FormatBoolStatus(value bool, trueText, falseText string) string {
	if value {
		return s.FormatSuccess(trueText)
	}
	return s.FormatError(falseText)
}

// Helper for consistent bold formatting
func (s *IOStreams) FormatBold(text string) string {
	return s.ColorBold(text)
}
