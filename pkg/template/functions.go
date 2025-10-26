package template

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/daddia/zen/internal/logging"
)

// DefaultFunctionRegistry implements FunctionRegistry interface
type DefaultFunctionRegistry struct {
	logger        logging.Logger
	workspaceRoot string
	functions     template.FuncMap
}

// NewFunctionRegistry creates a new function registry
func NewFunctionRegistry(logger logging.Logger, workspaceRoot string) *DefaultFunctionRegistry {
	registry := &DefaultFunctionRegistry{
		logger:        logger,
		workspaceRoot: workspaceRoot,
		functions:     make(template.FuncMap),
	}

	// Register standard functions
	registry.registerStandardFunctions()

	return registry
}

// RegisterFunction registers a custom template function
func (r *DefaultFunctionRegistry) RegisterFunction(name string, fn interface{}) error {
	if name == "" {
		return fmt.Errorf("function name cannot be empty")
	}

	if fn == nil {
		return fmt.Errorf("function cannot be nil")
	}

	r.functions[name] = fn
	r.logger.Debug("registered template function", "name", name)
	return nil
}

// GetFunctions returns all registered functions as template.FuncMap
func (r *DefaultFunctionRegistry) GetFunctions() template.FuncMap {
	// Return a copy to prevent external modification
	functions := make(template.FuncMap)
	for name, fn := range r.functions {
		functions[name] = fn
	}
	return functions
}

// RegisterZenFunctions registers Zen-specific template functions
func (r *DefaultFunctionRegistry) RegisterZenFunctions() error {
	zenFunctions := map[string]interface{}{
		// Task and ID functions
		"taskID":      r.generateTaskID,
		"taskIDShort": r.generateShortTaskID,
		"randomID":    r.generateRandomID,

		// Date and time functions
		"now":         r.now,
		"today":       r.today,
		"tomorrow":    r.tomorrow,
		"formatDate":  r.formatDate,
		"formatTime":  r.formatTime,
		"addDays":     r.addDays,
		"workingDays": r.workingDays,

		// Workflow stage functions
		"zenflowStages":    r.zenflowStages,
		"stageNumber":      r.stageNumber,
		"stageName":        r.stageName,
		"nextStage":        r.nextStage,
		"prevStage":        r.prevStage,
		"isStageCompleted": r.isStageCompleted,

		// Path and file functions
		"workspacePath": r.workspacePath,
		"relativePath":  r.relativePath,
		"joinPath":      r.joinPath,
		"fileName":      r.fileName,
		"fileExt":       r.fileExt,
		"dirName":       r.dirName,

		// String manipulation functions
		"camelCase":  r.camelCase,
		"pascalCase": r.pascalCase,
		"snakeCase":  r.snakeCase,
		"kebabCase":  r.kebabCase,
		"titleCase":  r.titleCase,
		"slugify":    r.slugify,

		// Formatting functions
		"indent":   r.indent,
		"dedent":   r.dedent,
		"wrap":     r.wrap,
		"truncate": r.truncate,
		"pad":      r.pad,

		// Collection functions
		"join":      r.join,
		"split":     r.split,
		"contains":  r.contains,
		"hasPrefix": r.hasPrefix,
		"hasSuffix": r.hasSuffix,
		"replace":   r.replace,

		// Conditional functions
		"default":  r.defaultValue,
		"coalesce": r.coalesce,
		"ternary":  r.ternary,

		// Zen metadata functions
		"zenVersion":   r.zenVersion,
		"zenWorkspace": r.zenWorkspace,
		"zenConfig":    r.zenConfig,
	}

	for name, fn := range zenFunctions {
		if err := r.RegisterFunction(name, fn); err != nil {
			return fmt.Errorf("failed to register Zen function '%s': %w", name, err)
		}
	}

	r.logger.Info("registered Zen template functions", "count", len(zenFunctions))
	return nil
}

// registerStandardFunctions registers standard utility functions
func (r *DefaultFunctionRegistry) registerStandardFunctions() {
	standardFunctions := map[string]interface{}{
		// String functions
		"upper":     strings.ToUpper,
		"lower":     strings.ToLower,
		"trim":      strings.TrimSpace,
		"trimLeft":  strings.TrimLeft,
		"trimRight": strings.TrimRight,

		// Math functions (basic)
		"add": r.add,
		"sub": r.subtract,
		"mul": r.multiply,
		"div": r.divide,
		"mod": r.modulo,

		// Type conversion
		"toString": r.toString,
		"toInt":    r.toInt,
	}

	for name, fn := range standardFunctions {
		r.functions[name] = fn
	}
}

// Task and ID functions

func (r *DefaultFunctionRegistry) generateTaskID(prefix string) string {
	timestamp := time.Now().Format("060102")
	randomBytes := make([]byte, 2)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp only if random fails
		return fmt.Sprintf("%s-%s", prefix, timestamp)
	}
	random := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s-%s", prefix, timestamp, strings.ToUpper(random))
}

func (r *DefaultFunctionRegistry) generateShortTaskID(prefix string) string {
	randomBytes := make([]byte, 2)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("%s-%d", prefix, time.Now().Unix()%10000)
	}
	random := hex.EncodeToString(randomBytes)
	return fmt.Sprintf("%s-%s", prefix, strings.ToUpper(random))
}

func (r *DefaultFunctionRegistry) generateRandomID(length int) string {
	if length <= 0 {
		length = 8
	}
	bytes := make([]byte, (length+1)/2)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID if random fails
		return fmt.Sprintf("ID%d", time.Now().UnixNano()%1000000)
	}
	id := hex.EncodeToString(bytes)
	if len(id) > length {
		id = id[:length]
	}
	return strings.ToUpper(id)
}

// Date and time functions

func (r *DefaultFunctionRegistry) now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func (r *DefaultFunctionRegistry) today() string {
	return time.Now().Format("2006-01-02")
}

func (r *DefaultFunctionRegistry) tomorrow() string {
	return time.Now().AddDate(0, 0, 1).Format("2006-01-02")
}

func (r *DefaultFunctionRegistry) formatDate(date interface{}, layout string) string {
	var t time.Time
	switch v := date.(type) {
	case time.Time:
		t = v
	case string:
		parsed, err := time.Parse("2006-01-02", v)
		if err != nil {
			return v // Return original if parsing fails
		}
		t = parsed
	default:
		return fmt.Sprintf("%v", date)
	}
	return t.Format(layout)
}

func (r *DefaultFunctionRegistry) formatTime(t time.Time, layout string) string {
	return t.Format(layout)
}

func (r *DefaultFunctionRegistry) addDays(date string, days int) string {
	parsed, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return parsed.AddDate(0, 0, days).Format("2006-01-02")
}

func (r *DefaultFunctionRegistry) workingDays(start, end string) int {
	startDate, err1 := time.Parse("2006-01-02", start)
	endDate, err2 := time.Parse("2006-01-02", end)
	if err1 != nil || err2 != nil {
		return 0
	}

	days := 0
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		weekday := d.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			days++
		}
	}
	return days
}

// Workflow stage functions

func (r *DefaultFunctionRegistry) zenflowStages() []map[string]interface{} {
	stages := []map[string]interface{}{
		{"number": 1, "id": "01-align", "name": "Align"},
		{"number": 2, "id": "02-discover", "name": "Discover"},
		{"number": 3, "id": "03-prioritize", "name": "Prioritize"},
		{"number": 4, "id": "04-design", "name": "Design"},
		{"number": 5, "id": "05-build", "name": "Build"},
		{"number": 6, "id": "06-ship", "name": "Ship"},
		{"number": 7, "id": "07-learn", "name": "Learn"},
	}
	return stages
}

func (r *DefaultFunctionRegistry) stageNumber(stageID string) int {
	stages := map[string]int{
		"01-align":      1,
		"02-discover":   2,
		"03-prioritize": 3,
		"04-design":     4,
		"05-build":      5,
		"06-ship":       6,
		"07-learn":      7,
	}
	if num, exists := stages[stageID]; exists {
		return num
	}
	return 0
}

func (r *DefaultFunctionRegistry) stageName(stageID string) string {
	stages := map[string]string{
		"01-align":      "Align",
		"02-discover":   "Discover",
		"03-prioritize": "Prioritize",
		"04-design":     "Design",
		"05-build":      "Build",
		"06-ship":       "Ship",
		"07-learn":      "Learn",
	}
	if name, exists := stages[stageID]; exists {
		return name
	}
	return stageID
}

func (r *DefaultFunctionRegistry) nextStage(currentStage string) string {
	stageOrder := []string{
		"01-align", "02-discover", "03-prioritize", "04-design",
		"05-build", "06-ship", "07-learn",
	}

	for i, stage := range stageOrder {
		if stage == currentStage && i < len(stageOrder)-1 {
			return stageOrder[i+1]
		}
	}
	return currentStage
}

func (r *DefaultFunctionRegistry) prevStage(currentStage string) string {
	stageOrder := []string{
		"01-align", "02-discover", "03-prioritize", "04-design",
		"05-build", "06-ship", "07-learn",
	}

	for i, stage := range stageOrder {
		if stage == currentStage && i > 0 {
			return stageOrder[i-1]
		}
	}
	return currentStage
}

func (r *DefaultFunctionRegistry) isStageCompleted(stage string, completedStages []string) bool {
	for _, completed := range completedStages {
		if completed == stage {
			return true
		}
	}
	return false
}

// Path and file functions

func (r *DefaultFunctionRegistry) workspacePath(path string) string {
	return filepath.Join(r.workspaceRoot, path)
}

func (r *DefaultFunctionRegistry) relativePath(path string) string {
	if r.workspaceRoot == "" {
		return path
	}
	rel, err := filepath.Rel(r.workspaceRoot, path)
	if err != nil {
		return path
	}
	return rel
}

func (r *DefaultFunctionRegistry) joinPath(parts ...string) string {
	return filepath.Join(parts...)
}

func (r *DefaultFunctionRegistry) fileName(path string) string {
	return filepath.Base(path)
}

func (r *DefaultFunctionRegistry) fileExt(path string) string {
	return filepath.Ext(path)
}

func (r *DefaultFunctionRegistry) dirName(path string) string {
	return filepath.Dir(path)
}

// String manipulation functions

func (r *DefaultFunctionRegistry) camelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_' || r == '-'
	})
	if len(words) == 0 {
		return s
	}
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		if len(words[i]) > 0 {
			result += strings.ToUpper(words[i][:1]) + strings.ToLower(words[i][1:])
		}
	}
	return result
}

func (r *DefaultFunctionRegistry) pascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_' || r == '-'
	})
	var result string
	for _, word := range words {
		if len(word) > 0 {
			result += strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return result
}

func (r *DefaultFunctionRegistry) snakeCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '-'
	})
	return strings.ToLower(strings.Join(words, "_"))
}

func (r *DefaultFunctionRegistry) kebabCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == '_'
	})
	return strings.ToLower(strings.Join(words, "-"))
}

func (r *DefaultFunctionRegistry) titleCase(s string) string {
	// Simple title case implementation to replace deprecated strings.Title
	words := strings.Fields(strings.ToLower(s))
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}
	return strings.Join(words, " ")
}

func (r *DefaultFunctionRegistry) slugify(s string) string {
	// Convert to lowercase and replace non-alphanumeric with hyphens
	var result strings.Builder
	for _, r := range strings.ToLower(s) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			result.WriteRune(r)
		} else if result.Len() > 0 && result.String()[result.Len()-1:] != "-" {
			result.WriteRune('-')
		}
	}
	return strings.Trim(result.String(), "-")
}

// Formatting functions

func (r *DefaultFunctionRegistry) indent(s string, spaces int) string {
	if spaces <= 0 {
		return s
	}
	prefix := strings.Repeat(" ", spaces)
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = prefix + line
		}
	}
	return strings.Join(lines, "\n")
}

func (r *DefaultFunctionRegistry) dedent(s string) string {
	lines := strings.Split(s, "\n")
	if len(lines) == 0 {
		return s
	}

	// Find minimum indentation (excluding empty lines)
	minIndent := -1
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		indent := 0
		for _, r := range line {
			switch r {
			case ' ':
				indent++
			case '\t':
				indent += 4
			default:
				goto breakLoop
			}
		}
	breakLoop:
		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent <= 0 {
		return s
	}

	// Remove minimum indentation from all lines
	for i, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		removed := 0
		for j, r := range line {
			if removed >= minIndent {
				lines[i] = line[j:]
				break
			}
			switch r {
			case ' ':
				removed++
			case '\t':
				removed += 4
			default:
				goto breakInnerLoop
			}
		}
	breakInnerLoop:
	}

	return strings.Join(lines, "\n")
}

func (r *DefaultFunctionRegistry) wrap(s string, width int) string {
	if width <= 0 {
		return s
	}

	words := strings.Fields(s)
	if len(words) == 0 {
		return s
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		switch {
		case currentLine.Len() == 0:
			currentLine.WriteString(word)
		case currentLine.Len()+1+len(word) <= width:
			currentLine.WriteString(" " + word)
		default:
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		}
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

func (r *DefaultFunctionRegistry) truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

func (r *DefaultFunctionRegistry) pad(s string, length int, padStr string) string {
	if len(s) >= length {
		return s
	}
	if padStr == "" {
		padStr = " "
	}
	padding := strings.Repeat(padStr, (length-len(s)+len(padStr)-1)/len(padStr))
	return s + padding[:length-len(s)]
}

// Collection functions

func (r *DefaultFunctionRegistry) join(items []string, separator string) string {
	return strings.Join(items, separator)
}

func (r *DefaultFunctionRegistry) split(s string, separator string) []string {
	return strings.Split(s, separator)
}

func (r *DefaultFunctionRegistry) contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func (r *DefaultFunctionRegistry) hasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

func (r *DefaultFunctionRegistry) hasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

func (r *DefaultFunctionRegistry) replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// Conditional functions

func (r *DefaultFunctionRegistry) defaultValue(value, defaultVal interface{}) interface{} {
	if value == nil || value == "" {
		return defaultVal
	}
	return value
}

func (r *DefaultFunctionRegistry) coalesce(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil && value != "" {
			return value
		}
	}
	return nil
}

func (r *DefaultFunctionRegistry) ternary(condition bool, trueVal, falseVal interface{}) interface{} {
	if condition {
		return trueVal
	}
	return falseVal
}

// Math functions

func (r *DefaultFunctionRegistry) add(a, b interface{}) interface{} {
	return r.mathOp(a, b, func(x, y float64) float64 { return x + y })
}

func (r *DefaultFunctionRegistry) subtract(a, b interface{}) interface{} {
	return r.mathOp(a, b, func(x, y float64) float64 { return x - y })
}

func (r *DefaultFunctionRegistry) multiply(a, b interface{}) interface{} {
	return r.mathOp(a, b, func(x, y float64) float64 { return x * y })
}

func (r *DefaultFunctionRegistry) divide(a, b interface{}) interface{} {
	return r.mathOp(a, b, func(x, y float64) float64 {
		if y == 0 {
			return 0
		}
		return x / y
	})
}

func (r *DefaultFunctionRegistry) modulo(a, b interface{}) interface{} {
	aInt, aOk := a.(int)
	bInt, bOk := b.(int)
	if aOk && bOk && bInt != 0 {
		return aInt % bInt
	}
	return 0
}

func (r *DefaultFunctionRegistry) mathOp(a, b interface{}, op func(float64, float64) float64) interface{} {
	aFloat := r.toFloat64(a)
	bFloat := r.toFloat64(b)
	return op(aFloat, bFloat)
}

// Type conversion functions

func (r *DefaultFunctionRegistry) toString(v interface{}) string {
	return fmt.Sprintf("%v", v)
}

func (r *DefaultFunctionRegistry) toInt(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		if i, err := fmt.Sscanf(val, "%d", new(int)); err == nil && i == 1 {
			var result int
			if _, err := fmt.Sscanf(val, "%d", &result); err == nil {
				return result
			}
		}
	}
	return 0
}

func (r *DefaultFunctionRegistry) toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case float64:
		return val
	case float32:
		return float64(val)
	case string:
		if f, err := fmt.Sscanf(val, "%f", new(float64)); err == nil && f == 1 {
			var result float64
			if _, err := fmt.Sscanf(val, "%f", &result); err == nil {
				return result
			}
		}
	}
	return 0
}

// Zen metadata functions

func (r *DefaultFunctionRegistry) zenVersion() string {
	return "0.3.0" // This should be injected from build info
}

func (r *DefaultFunctionRegistry) zenWorkspace() string {
	return r.workspaceRoot
}

func (r *DefaultFunctionRegistry) zenConfig() map[string]interface{} {
	// This would normally load from actual config
	return map[string]interface{}{
		"workspace_root": r.workspaceRoot,
		"version":        "0.3.0",
	}
}
