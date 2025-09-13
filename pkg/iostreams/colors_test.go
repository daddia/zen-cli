package iostreams

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestColorFunctions(t *testing.T) {
	tests := []struct {
		name     string
		colorFn  func(*IOStreams, string) string
		input    string
		expected string
	}{
		{
			name:     "ColorSuccess with color enabled",
			colorFn:  (*IOStreams).ColorSuccess,
			input:    "success",
			expected: ColorGreen + "success" + ColorReset,
		},
		{
			name:     "ColorError with color enabled",
			colorFn:  (*IOStreams).ColorError,
			input:    "error",
			expected: ColorRed + "error" + ColorReset,
		},
		{
			name:     "ColorWarning with color enabled",
			colorFn:  (*IOStreams).ColorWarning,
			input:    "warning",
			expected: ColorYellow + "warning" + ColorReset,
		},
		{
			name:     "ColorInfo with color enabled",
			colorFn:  (*IOStreams).ColorInfo,
			input:    "info",
			expected: ColorBlue + "info" + ColorReset,
		},
		{
			name:     "ColorBold with color enabled",
			colorFn:  (*IOStreams).ColorBold,
			input:    "bold",
			expected: ColorBold + "bold" + ColorReset,
		},
		{
			name:     "ColorNeutral with color enabled",
			colorFn:  (*IOStreams).ColorNeutral,
			input:    "neutral",
			expected: ColorCyan + "neutral" + ColorReset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			streams := Test()
			streams.SetColorEnabled(true)
			result := tt.colorFn(streams, tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestColorFunctionsDisabled(t *testing.T) {
	streams := Test()
	streams.SetColorEnabled(false)

	tests := []struct {
		name    string
		colorFn func(*IOStreams, string) string
		input   string
	}{
		{"ColorSuccess", (*IOStreams).ColorSuccess, "success"},
		{"ColorError", (*IOStreams).ColorError, "error"},
		{"ColorWarning", (*IOStreams).ColorWarning, "warning"},
		{"ColorInfo", (*IOStreams).ColorInfo, "info"},
		{"ColorBold", (*IOStreams).ColorBold, "bold"},
		{"ColorNeutral", (*IOStreams).ColorNeutral, "neutral"},
	}

	for _, tt := range tests {
		t.Run(tt.name+" disabled", func(t *testing.T) {
			result := tt.colorFn(streams, tt.input)
			assert.Equal(t, tt.input, result, "Should return plain text when color is disabled")
		})
	}
}

func TestFormatFunctions(t *testing.T) {
	streams := Test()
	streams.SetColorEnabled(true)

	tests := []struct {
		name     string
		formatFn func(*IOStreams, string) string
		input    string
		symbol   string
		color    string
	}{
		{
			name:     "FormatSuccess",
			formatFn: (*IOStreams).FormatSuccess,
			input:    "operation completed",
			symbol:   SymbolSuccess,
			color:    ColorGreen,
		},
		{
			name:     "FormatError",
			formatFn: (*IOStreams).FormatError,
			input:    "operation failed",
			symbol:   SymbolFailure,
			color:    ColorRed,
		},
		{
			name:     "FormatWarning",
			formatFn: (*IOStreams).FormatWarning,
			input:    "warning message",
			symbol:   SymbolAlert,
			color:    ColorYellow,
		},
		{
			name:     "FormatNeutral",
			formatFn: (*IOStreams).FormatNeutral,
			input:    "neutral message",
			symbol:   SymbolNeutral,
			color:    ColorCyan,
		},
		{
			name:     "FormatChange",
			formatFn: (*IOStreams).FormatChange,
			input:    "change required",
			symbol:   SymbolChange,
			color:    ColorYellow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.formatFn(streams, tt.input)
			expected := tt.color + tt.symbol + " " + tt.input + ColorReset
			assert.Equal(t, expected, result)
		})
	}
}

func TestFormatFunctionsDisabled(t *testing.T) {
	streams := Test()
	streams.SetColorEnabled(false)

	tests := []struct {
		name     string
		formatFn func(*IOStreams, string) string
		input    string
		symbol   string
	}{
		{"FormatSuccess", (*IOStreams).FormatSuccess, "success", SymbolSuccess},
		{"FormatError", (*IOStreams).FormatError, "error", SymbolFailure},
		{"FormatWarning", (*IOStreams).FormatWarning, "warning", SymbolAlert},
		{"FormatNeutral", (*IOStreams).FormatNeutral, "neutral", SymbolNeutral},
		{"FormatChange", (*IOStreams).FormatChange, "change", SymbolChange},
	}

	for _, tt := range tests {
		t.Run(tt.name+" disabled", func(t *testing.T) {
			result := tt.formatFn(streams, tt.input)
			expected := tt.symbol + " " + tt.input
			assert.Equal(t, expected, result)
		})
	}
}

func TestHeaderFormatting(t *testing.T) {
	streams := Test()
	streams.SetColorEnabled(true)

	t.Run("FormatHeader", func(t *testing.T) {
		result := streams.FormatHeader("Test Header")
		expected := ColorBold + "Test Header" + ColorReset
		assert.Equal(t, expected, result)
	})

	t.Run("FormatSectionHeader", func(t *testing.T) {
		result := streams.FormatSectionHeader("Section")
		expected := ColorBold + "Section" + ColorReset + "\n" + ColorBold + "=======" + ColorReset
		assert.Equal(t, expected, result)
	})

	t.Run("FormatSubHeader", func(t *testing.T) {
		result := streams.FormatSubHeader("SubSection")
		expected := ColorBold + "SubSection" + ColorReset + "\n" + ColorCyan + "----------" + ColorReset
		assert.Equal(t, expected, result)
	})
}

func TestHeaderFormattingDisabled(t *testing.T) {
	streams := Test()
	streams.SetColorEnabled(false)

	t.Run("FormatHeader disabled", func(t *testing.T) {
		result := streams.FormatHeader("Test Header")
		assert.Equal(t, "Test Header", result)
	})

	t.Run("FormatSectionHeader disabled", func(t *testing.T) {
		result := streams.FormatSectionHeader("Section")
		expected := "Section\n======="
		assert.Equal(t, expected, result)
	})

	t.Run("FormatSubHeader disabled", func(t *testing.T) {
		result := streams.FormatSubHeader("SubSection")
		expected := "SubSection\n----------"
		assert.Equal(t, expected, result)
	})
}

func TestIndent(t *testing.T) {
	streams := Test()

	tests := []struct {
		name     string
		input    string
		level    int
		expected string
	}{
		{
			name:     "single line, level 1",
			input:    "test",
			level:    1,
			expected: "  test",
		},
		{
			name:     "single line, level 2",
			input:    "test",
			level:    2,
			expected: "    test",
		},
		{
			name:     "multi line, level 1",
			input:    "line1\nline2",
			level:    1,
			expected: "  line1\n  line2",
		},
		{
			name:     "empty lines preserved",
			input:    "line1\n\nline3",
			level:    1,
			expected: "  line1\n\n  line3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := streams.Indent(tt.input, tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatTable(t *testing.T) {
	streams := Test()
	streams.SetColorEnabled(true)

	headers := []string{"Name", "Status", "Type"}
	rows := [][]string{
		{"test1", "active", "service"},
		{"test2", "inactive", "job"},
	}

	result := streams.FormatTable(headers, rows)

	// Should contain headers with bold formatting (accounting for spacing)
	assert.Contains(t, result, ColorBold+"Name ")
	assert.Contains(t, result, ColorBold+"Status  ")
	assert.Contains(t, result, ColorBold+"Type   ")

	// Should contain row data
	assert.Contains(t, result, "test1")
	assert.Contains(t, result, "active")
	assert.Contains(t, result, "service")
}

func TestFormatMachineTable(t *testing.T) {
	streams := Test()

	headers := []string{"Name", "Status", "Type"}
	rows := [][]string{
		{"test1", "active", "service"},
		{"test2", "inactive", "job"},
	}

	result := streams.FormatMachineTable(headers, rows)

	// Should use tabs as delimiters
	assert.Contains(t, result, "test1\tactive\tservice")
	assert.Contains(t, result, "test2\tinactive\tjob")

	// Should not contain headers
	assert.NotContains(t, result, "Name")
	assert.NotContains(t, result, "Status")
	assert.NotContains(t, result, "Type")
}

func TestFormatStatus(t *testing.T) {
	streams := Test()
	streams.SetColorEnabled(true)

	t.Run("success status", func(t *testing.T) {
		result := streams.FormatStatus("Ready", true)
		expected := ColorGreen + SymbolSuccess + " Ready" + ColorReset
		assert.Equal(t, expected, result)
	})

	t.Run("failure status", func(t *testing.T) {
		result := streams.FormatStatus("Failed", false)
		expected := ColorRed + SymbolFailure + " Failed" + ColorReset
		assert.Equal(t, expected, result)
	})
}

func TestFormatBoolStatus(t *testing.T) {
	streams := Test()
	streams.SetColorEnabled(true)

	t.Run("true value", func(t *testing.T) {
		result := streams.FormatBoolStatus(true, "Enabled", "Disabled")
		expected := ColorGreen + SymbolSuccess + " Enabled" + ColorReset
		assert.Equal(t, expected, result)
	})

	t.Run("false value", func(t *testing.T) {
		result := streams.FormatBoolStatus(false, "Enabled", "Disabled")
		expected := ColorRed + SymbolFailure + " Disabled" + ColorReset
		assert.Equal(t, expected, result)
	})
}

func TestSymbolConstants(t *testing.T) {
	// Verify Unicode symbols match design guide
	assert.Equal(t, "✓", SymbolSuccess)
	assert.Equal(t, "-", SymbolNeutral)
	assert.Equal(t, "✗", SymbolFailure)
	assert.Equal(t, "!", SymbolAlert)
	assert.Equal(t, "+", SymbolChange)
}
