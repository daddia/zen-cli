package template

import (
	"testing"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewFunctionRegistry(t *testing.T) {
	logger := logging.NewBasic()
	workspaceRoot := "/test/workspace"

	registry := NewFunctionRegistry(logger, workspaceRoot)

	assert.NotNil(t, registry)
	assert.Equal(t, workspaceRoot, registry.workspaceRoot)
	assert.NotNil(t, registry.functions)

	// Should have standard functions registered
	functions := registry.GetFunctions()
	assert.Contains(t, functions, "upper")
	assert.Contains(t, functions, "lower")
	assert.Contains(t, functions, "add")
}

func TestFunctionRegistry_RegisterFunction(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")

	// Test successful registration
	err := registry.RegisterFunction("testFunc", func() string { return "test" })
	assert.NoError(t, err)

	functions := registry.GetFunctions()
	assert.Contains(t, functions, "testFunc")

	// Test error cases
	err = registry.RegisterFunction("", func() string { return "test" })
	assert.Error(t, err)

	err = registry.RegisterFunction("nilFunc", nil)
	assert.Error(t, err)
}

func TestFunctionRegistry_RegisterZenFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test/workspace")

	err := registry.RegisterZenFunctions()
	assert.NoError(t, err)

	functions := registry.GetFunctions()

	// Test Zen-specific functions are registered
	zenFunctions := []string{
		"taskID", "taskIDShort", "randomID",
		"now", "today", "tomorrow", "formatDate",
		"zenflowStages", "stageNumber", "stageName",
		"workspacePath", "relativePath", "joinPath",
		"camelCase", "pascalCase", "snakeCase", "kebabCase",
		"indent", "dedent", "wrap", "truncate",
		"join", "split", "contains", "hasPrefix",
		"default", "coalesce", "ternary",
		"zenVersion", "zenWorkspace",
	}

	for _, funcName := range zenFunctions {
		assert.Contains(t, functions, funcName, "Zen function %s should be registered", funcName)
	}
}

func TestTaskIDFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test taskID function
	taskIDFunc := functions["taskID"].(func(string) string)
	result := taskIDFunc("TEST")
	assert.Contains(t, result, "TEST-")
	assert.True(t, len(result) >= 12) // TEST- + date + - + random (length can vary)

	// Test taskIDShort function
	taskIDShortFunc := functions["taskIDShort"].(func(string) string)
	result = taskIDShortFunc("TEST")
	assert.Contains(t, result, "TEST-")
	assert.Len(t, result, 9) // TEST- + 4 chars (random)

	// Test randomID function
	randomIDFunc := functions["randomID"].(func(int) string)
	result = randomIDFunc(8)
	assert.Len(t, result, 8)
	assert.Regexp(t, "^[A-F0-9]+$", result)
}

func TestDateTimeFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test now function
	nowFunc := functions["now"].(func() string)
	result := nowFunc()
	assert.Regexp(t, `^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, result)

	// Test today function
	todayFunc := functions["today"].(func() string)
	result = todayFunc()
	assert.Equal(t, time.Now().Format("2006-01-02"), result)

	// Test tomorrow function
	tomorrowFunc := functions["tomorrow"].(func() string)
	result = tomorrowFunc()
	expected := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	assert.Equal(t, expected, result)

	// Test formatDate function
	formatDateFunc := functions["formatDate"].(func(interface{}, string) string)
	testTime := time.Date(2025, 9, 19, 0, 0, 0, 0, time.UTC)
	result = formatDateFunc(testTime, "2006/01/02")
	assert.Equal(t, "2025/09/19", result)

	// Test addDays function
	addDaysFunc := functions["addDays"].(func(string, int) string)
	result = addDaysFunc("2025-09-19", 5)
	assert.Equal(t, "2025-09-24", result)

	// Test workingDays function
	workingDaysFunc := functions["workingDays"].(func(string, string) int)
	days := workingDaysFunc("2025-09-15", "2025-09-19") // Mon to Fri
	assert.Equal(t, 5, days)
}

func TestWorkflowFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test zenflowStages function
	zenflowStagesFunc := functions["zenflowStages"].(func() []map[string]interface{})
	stages := zenflowStagesFunc()
	assert.Len(t, stages, 7)
	assert.Equal(t, "01-align", stages[0]["id"])
	assert.Equal(t, "Align", stages[0]["name"])
	assert.Equal(t, 1, stages[0]["number"])

	// Test stageNumber function
	stageNumberFunc := functions["stageNumber"].(func(string) int)
	assert.Equal(t, 1, stageNumberFunc("01-align"))
	assert.Equal(t, 4, stageNumberFunc("04-design"))
	assert.Equal(t, 0, stageNumberFunc("invalid-stage"))

	// Test stageName function
	stageNameFunc := functions["stageName"].(func(string) string)
	assert.Equal(t, "Align", stageNameFunc("01-align"))
	assert.Equal(t, "Design", stageNameFunc("04-design"))
	assert.Equal(t, "invalid-stage", stageNameFunc("invalid-stage"))

	// Test nextStage function
	nextStageFunc := functions["nextStage"].(func(string) string)
	assert.Equal(t, "02-discover", nextStageFunc("01-align"))
	assert.Equal(t, "07-learn", nextStageFunc("06-ship"))
	assert.Equal(t, "07-learn", nextStageFunc("07-learn")) // Last stage

	// Test prevStage function
	prevStageFunc := functions["prevStage"].(func(string) string)
	assert.Equal(t, "01-align", prevStageFunc("02-discover"))
	assert.Equal(t, "06-ship", prevStageFunc("07-learn"))
	assert.Equal(t, "01-align", prevStageFunc("01-align")) // First stage

	// Test isStageCompleted function
	isStageCompletedFunc := functions["isStageCompleted"].(func(string, []string) bool)
	completedStages := []string{"01-align", "02-discover", "03-prioritize"}
	assert.True(t, isStageCompletedFunc("01-align", completedStages))
	assert.True(t, isStageCompletedFunc("02-discover", completedStages))
	assert.False(t, isStageCompletedFunc("04-design", completedStages))
}

func TestPathFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test/workspace")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test workspacePath function
	workspacePathFunc := functions["workspacePath"].(func(string) string)
	result := workspacePathFunc("subdir/file.txt")
	assert.Equal(t, "/test/workspace/subdir/file.txt", result)

	// Test joinPath function
	joinPathFunc := functions["joinPath"].(func(...string) string)
	result = joinPathFunc("dir1", "dir2", "file.txt")
	assert.Equal(t, "dir1/dir2/file.txt", result)

	// Test fileName function
	fileNameFunc := functions["fileName"].(func(string) string)
	result = fileNameFunc("/path/to/file.txt")
	assert.Equal(t, "file.txt", result)

	// Test fileExt function
	fileExtFunc := functions["fileExt"].(func(string) string)
	result = fileExtFunc("/path/to/file.txt")
	assert.Equal(t, ".txt", result)

	// Test dirName function
	dirNameFunc := functions["dirName"].(func(string) string)
	result = dirNameFunc("/path/to/file.txt")
	assert.Equal(t, "/path/to", result)
}

func TestStringManipulationFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	tests := []struct {
		funcName string
		input    string
		expected string
	}{
		{"camelCase", "hello world", "helloWorld"},
		{"camelCase", "hello-world-test", "helloWorldTest"},
		{"pascalCase", "hello world", "HelloWorld"},
		{"pascalCase", "hello-world-test", "HelloWorldTest"},
		{"snakeCase", "hello world", "hello_world"},
		{"snakeCase", "hello-world", "hello_world"},
		{"kebabCase", "hello world", "hello-world"},
		{"kebabCase", "hello_world", "hello-world"},
		{"titleCase", "hello world", "Hello World"},
		{"slugify", "Hello World!", "hello-world"},
	}

	for _, tt := range tests {
		t.Run(tt.funcName, func(t *testing.T) {
			fn := functions[tt.funcName].(func(string) string)
			result := fn(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormattingFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test indent function
	indentFunc := functions["indent"].(func(string, int) string)
	result := indentFunc("line1\nline2", 2)
	assert.Equal(t, "  line1\n  line2", result)

	// Test truncate function
	truncateFunc := functions["truncate"].(func(string, int) string)
	result = truncateFunc("hello world", 8)
	assert.Equal(t, "hello...", result)

	// Test pad function
	padFunc := functions["pad"].(func(string, int, string) string)
	result = padFunc("test", 8, " ")
	assert.Equal(t, "test    ", result)

	// Test wrap function
	wrapFunc := functions["wrap"].(func(string, int) string)
	result = wrapFunc("hello world test", 10)
	assert.Equal(t, "hello\nworld test", result)
}

func TestCollectionFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test join function
	joinFunc := functions["join"].(func([]string, string) string)
	result := joinFunc([]string{"a", "b", "c"}, ",")
	assert.Equal(t, "a,b,c", result)

	// Test split function
	splitFunc := functions["split"].(func(string, string) []string)
	result2 := splitFunc("a,b,c", ",")
	assert.Equal(t, []string{"a", "b", "c"}, result2)

	// Test contains function
	containsFunc := functions["contains"].(func(string, string) bool)
	assert.True(t, containsFunc("hello world", "world"))
	assert.False(t, containsFunc("hello world", "test"))

	// Test hasPrefix function
	hasPrefixFunc := functions["hasPrefix"].(func(string, string) bool)
	assert.True(t, hasPrefixFunc("hello world", "hello"))
	assert.False(t, hasPrefixFunc("hello world", "world"))

	// Test hasSuffix function
	hasSuffixFunc := functions["hasSuffix"].(func(string, string) bool)
	assert.True(t, hasSuffixFunc("hello world", "world"))
	assert.False(t, hasSuffixFunc("hello world", "hello"))

	// Test replace function
	replaceFunc := functions["replace"].(func(string, string, string) string)
	result = replaceFunc("hello world", "world", "universe")
	assert.Equal(t, "hello universe", result)
}

func TestConditionalFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test default function
	defaultFunc := functions["default"].(func(interface{}, interface{}) interface{})
	assert.Equal(t, "default", defaultFunc("", "default"))
	assert.Equal(t, "default", defaultFunc(nil, "default"))
	assert.Equal(t, "value", defaultFunc("value", "default"))

	// Test coalesce function
	coalesceFunc := functions["coalesce"].(func(...interface{}) interface{})
	assert.Equal(t, "first", coalesceFunc("first", "second"))
	assert.Equal(t, "second", coalesceFunc("", "second"))
	assert.Equal(t, "third", coalesceFunc("", nil, "third"))

	// Test ternary function
	ternaryFunc := functions["ternary"].(func(bool, interface{}, interface{}) interface{})
	assert.Equal(t, "true", ternaryFunc(true, "true", "false"))
	assert.Equal(t, "false", ternaryFunc(false, "true", "false"))
}

func TestMathFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test add function
	addFunc := functions["add"].(func(interface{}, interface{}) interface{})
	assert.Equal(t, 5.0, addFunc(2, 3))
	assert.Equal(t, 5.0, addFunc(2.5, 2.5))

	// Test subtract function
	subFunc := functions["sub"].(func(interface{}, interface{}) interface{})
	assert.Equal(t, 1.0, subFunc(3, 2))

	// Test multiply function
	mulFunc := functions["mul"].(func(interface{}, interface{}) interface{})
	assert.Equal(t, 6.0, mulFunc(2, 3))

	// Test divide function
	divFunc := functions["div"].(func(interface{}, interface{}) interface{})
	assert.Equal(t, 2.0, divFunc(6, 3))
	assert.Equal(t, 0.0, divFunc(6, 0)) // Division by zero

	// Test modulo function
	modFunc := functions["mod"].(func(interface{}, interface{}) interface{})
	assert.Equal(t, 1, modFunc(5, 2))
	assert.Equal(t, 0, modFunc(5, 0)) // Modulo by zero
}

func TestTypeConversionFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test toString function
	toStringFunc := functions["toString"].(func(interface{}) string)
	assert.Equal(t, "123", toStringFunc(123))
	assert.Equal(t, "true", toStringFunc(true))
	assert.Equal(t, "hello", toStringFunc("hello"))

	// Test toInt function
	toIntFunc := functions["toInt"].(func(interface{}) int)
	assert.Equal(t, 123, toIntFunc(123))
	assert.Equal(t, 123, toIntFunc(123.7))
	assert.Equal(t, 123, toIntFunc("123"))
	assert.Equal(t, 0, toIntFunc("invalid"))
}

func TestZenMetadataFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test/workspace")
	registry.RegisterZenFunctions()

	functions := registry.GetFunctions()

	// Test zenVersion function
	zenVersionFunc := functions["zenVersion"].(func() string)
	result := zenVersionFunc()
	assert.Equal(t, "0.3.0", result)

	// Test zenWorkspace function
	zenWorkspaceFunc := functions["zenWorkspace"].(func() string)
	result = zenWorkspaceFunc()
	assert.Equal(t, "/test/workspace", result)

	// Test zenConfig function
	zenConfigFunc := functions["zenConfig"].(func() map[string]interface{})
	config := zenConfigFunc()
	assert.Equal(t, "/test/workspace", config["workspace_root"])
	assert.Equal(t, "0.3.0", config["version"])
}

func TestStandardFunctions(t *testing.T) {
	logger := logging.NewBasic()
	registry := NewFunctionRegistry(logger, "/test")

	functions := registry.GetFunctions()

	// Test standard string functions
	upperFunc := functions["upper"].(func(string) string)
	assert.Equal(t, "HELLO", upperFunc("hello"))

	lowerFunc := functions["lower"].(func(string) string)
	assert.Equal(t, "hello", lowerFunc("HELLO"))

	trimFunc := functions["trim"].(func(string) string)
	assert.Equal(t, "hello", trimFunc("  hello  "))
}
