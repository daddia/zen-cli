package templates

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed task/*.tmpl
var templateFiles embed.FS

// LocalTemplateLoader provides access to embedded template files
type LocalTemplateLoader struct {
	templates map[string]*template.Template
}

// NewLocalTemplateLoader creates a new local template loader
func NewLocalTemplateLoader() *LocalTemplateLoader {
	return &LocalTemplateLoader{
		templates: make(map[string]*template.Template),
	}
}

// LoadTemplate loads and compiles a template by name from the task/ subdirectory
func (l *LocalTemplateLoader) LoadTemplate(name string) (*template.Template, error) {
	// Check cache first
	if tmpl, exists := l.templates[name]; exists {
		return tmpl, nil
	}

	// Build template path in task/ subdirectory
	templatePath := filepath.Join("task", name)
	if filepath.Ext(name) != ".tmpl" {
		templatePath = filepath.Join("task", name+".tmpl")
	}

	content, err := templateFiles.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", templatePath, err)
	}

	// Compile template
	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to compile template %s: %w", name, err)
	}

	// Cache compiled template
	l.templates[name] = tmpl

	return tmpl, nil
}

// RenderTemplate renders a template with the given variables
func (l *LocalTemplateLoader) RenderTemplate(name string, variables map[string]interface{}) (string, error) {
	tmpl, err := l.LoadTemplate(name)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", name, err)
	}

	return buf.String(), nil
}
