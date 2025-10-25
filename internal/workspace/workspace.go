package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/fs"
)

// Manager implements workspace operations
type Manager struct {
	root       string
	configFile string
	logger     logging.Logger
	fsManager  *fs.Manager
}

// New creates a new workspace manager
func New(root, configFile string, logger logging.Logger) *Manager {
	// Use the config file path provided by the Config module
	// If none provided, use default
	if configFile == "" {
		configFile = filepath.Join(root, ".zen", "config")
	}

	return &Manager{
		root:       root,
		configFile: configFile,
		logger:     logger,
		fsManager:  fs.New(logger),
	}
}

// ProjectType represents different project types that can be detected
type ProjectType string

const (
	ProjectTypeUnknown ProjectType = "unknown"
	ProjectTypeGit     ProjectType = "git"
	ProjectTypeNodeJS  ProjectType = "nodejs"
	ProjectTypeGo      ProjectType = "go"
	ProjectTypePython  ProjectType = "python"
	ProjectTypeRust    ProjectType = "rust"
	ProjectTypeJava    ProjectType = "java"
)

// ProjectInfo contains detected project information
type ProjectInfo struct {
	Type        ProjectType       `json:"type" yaml:"type"`
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description,omitempty" yaml:"description,omitempty"`
	Version     string            `json:"version,omitempty" yaml:"version,omitempty"`
	Language    string            `json:"language,omitempty" yaml:"language,omitempty"`
	Framework   string            `json:"framework,omitempty" yaml:"framework,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}

// WorkspaceConfig represents the full workspace configuration
type WorkspaceConfig struct {
	Version   string      `json:"version" yaml:"version"`
	Workspace Workspace   `json:"workspace" yaml:"workspace"`
	Project   ProjectInfo `json:"project" yaml:"project"`
	Logging   LogConfig   `json:"logging" yaml:"logging"`
	CLI       CLIConfig   `json:"cli" yaml:"cli"`
}

// Workspace contains workspace-specific settings
type Workspace struct {
	Root        string    `json:"root" yaml:"root"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description,omitempty" yaml:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
}

// LogConfig contains logging configuration
type LogConfig struct {
	Level  string `json:"level" yaml:"level"`
	Format string `json:"format" yaml:"format"`
}

// CLIConfig contains CLI configuration
type CLIConfig struct {
	OutputFormat string `json:"output_format" yaml:"output_format"`
	NoColor      bool   `json:"no_color" yaml:"no_color"`
}

// Status represents workspace status
type Status struct {
	Initialized  bool        `json:"initialized" yaml:"initialized"`
	ConfigPath   string      `json:"config_path" yaml:"config_path"`
	Root         string      `json:"root" yaml:"root"`
	ZenDirectory string      `json:"zen_directory" yaml:"zen_directory"`
	Project      ProjectInfo `json:"project" yaml:"project"`
}

// Root returns the workspace root directory
func (m *Manager) Root() string {
	return m.root
}

// ConfigFile returns the configuration file path
func (m *Manager) ConfigFile() string {
	return m.configFile
}

// ZenDirectory returns the .zen directory path
func (m *Manager) ZenDirectory() string {
	return filepath.Join(m.root, ".zen")
}

// DetectProject analyzes the current directory to determine project type and details
func (m *Manager) DetectProject() ProjectInfo {
	m.logger.Debug("Detecting project type", map[string]interface{}{
		"root": m.root,
	})

	var detectedTypes []ProjectType
	info := ProjectInfo{
		Type:     ProjectTypeUnknown,
		Name:     filepath.Base(m.root),
		Metadata: make(map[string]string),
	}

	// Check for Git repository
	if m.isGitRepository() {
		detectedTypes = append(detectedTypes, ProjectTypeGit)
		if gitInfo := m.detectGitInfo(); gitInfo != nil {
			if gitInfo.Name != "" {
				info.Name = gitInfo.Name
			}
			if gitInfo.Description != "" {
				info.Description = gitInfo.Description
			}
			for k, v := range gitInfo.Metadata {
				info.Metadata[k] = v
			}
		}
	}

	// Check for Node.js project
	if nodeInfo := m.detectNodeJS(); nodeInfo != nil {
		detectedTypes = append(detectedTypes, ProjectTypeNodeJS)
		info.Type = ProjectTypeNodeJS
		info.Language = nodeInfo.Language
		if nodeInfo.Name != "" {
			info.Name = nodeInfo.Name
		}
		if nodeInfo.Description != "" {
			info.Description = nodeInfo.Description
		}
		if nodeInfo.Version != "" {
			info.Version = nodeInfo.Version
		}
		if nodeInfo.Framework != "" {
			info.Framework = nodeInfo.Framework
		}
		for k, v := range nodeInfo.Metadata {
			info.Metadata[k] = v
		}
	}

	// Check for Go project
	if goInfo := m.detectGo(); goInfo != nil {
		detectedTypes = append(detectedTypes, ProjectTypeGo)
		info.Type = ProjectTypeGo
		info.Language = "go"
		if goInfo.Name != "" {
			info.Name = goInfo.Name
		}
		if goInfo.Version != "" {
			info.Version = goInfo.Version
		}
		for k, v := range goInfo.Metadata {
			info.Metadata[k] = v
		}
	}

	// Check for Python project
	if pythonInfo := m.detectPython(); pythonInfo != nil {
		detectedTypes = append(detectedTypes, ProjectTypePython)
		info.Type = ProjectTypePython
		info.Language = "python"
		if pythonInfo.Name != "" {
			info.Name = pythonInfo.Name
		}
		if pythonInfo.Description != "" {
			info.Description = pythonInfo.Description
		}
		if pythonInfo.Version != "" {
			info.Version = pythonInfo.Version
		}
		for k, v := range pythonInfo.Metadata {
			info.Metadata[k] = v
		}
	}

	// Check for Rust project
	if rustInfo := m.detectRust(); rustInfo != nil {
		detectedTypes = append(detectedTypes, ProjectTypeRust)
		info.Type = ProjectTypeRust
		info.Language = "rust"
		if rustInfo.Name != "" {
			info.Name = rustInfo.Name
		}
		if rustInfo.Description != "" {
			info.Description = rustInfo.Description
		}
		if rustInfo.Version != "" {
			info.Version = rustInfo.Version
		}
		for k, v := range rustInfo.Metadata {
			info.Metadata[k] = v
		}
	}

	// Check for Java project
	if javaInfo := m.detectJava(); javaInfo != nil {
		detectedTypes = append(detectedTypes, ProjectTypeJava)
		info.Type = ProjectTypeJava
		info.Language = "java"
		if javaInfo.Name != "" {
			info.Name = javaInfo.Name
		}
		if javaInfo.Description != "" {
			info.Description = javaInfo.Description
		}
		if javaInfo.Version != "" {
			info.Version = javaInfo.Version
		}
		if javaInfo.Framework != "" {
			info.Framework = javaInfo.Framework
		}
		for k, v := range javaInfo.Metadata {
			info.Metadata[k] = v
		}
	}

	// If multiple types detected, prioritize based on specificity
	if len(detectedTypes) > 1 {
		// Priority: Language-specific > Git
		for _, t := range []ProjectType{ProjectTypeNodeJS, ProjectTypeGo, ProjectTypePython, ProjectTypeRust, ProjectTypeJava} {
			for _, detected := range detectedTypes {
				if detected == t {
					info.Type = t
					break
				}
			}
			if info.Type != ProjectTypeUnknown {
				break
			}
		}
	}

	m.logger.Debug("Project detection completed", map[string]interface{}{
		"type":           info.Type,
		"name":           info.Name,
		"detected_types": detectedTypes,
	})

	return info
}

// Initialize creates the workspace structure and configuration
func (m *Manager) Initialize(force bool) error {
	m.logger.Debug("Initializing workspace", map[string]interface{}{
		"root":  m.root,
		"force": force,
	})

	// Check if .zen directory already exists
	zenDirExists := false
	if _, err := os.Stat(m.ZenDirectory()); err == nil {
		zenDirExists = true
	}

	// If already initialized and not forcing, just reinitialize (idempotent behavior)
	if zenDirExists && !force {
		m.logger.Debug("Workspace already initialized, reinitializing", map[string]interface{}{
			"zen_directory": m.ZenDirectory(),
		})
		// Don't return an error - just proceed with reinitialization
	}

	// Create .zen directory structure
	if err := m.createWorkspace(); err != nil {
		return fmt.Errorf("failed to create .zen directory: %w", err)
	}

	// Detect project information
	projectInfo := m.DetectProject()

	// Update .gitignore if it exists
	if err := m.updateGitignore(); err != nil {
		m.logger.Warn("Failed to update .gitignore", map[string]interface{}{
			"error": err.Error(),
		})
	}

	m.logger.Debug("Workspace initialized successfully", map[string]interface{}{
		"config_file": m.configFile,
		"zen_dir":     m.ZenDirectory(),
		"project":     projectInfo.Type,
	})

	return nil
}

// Status returns the current workspace status
func (m *Manager) Status() (Status, error) {
	status := Status{
		Root:         m.root,
		ConfigPath:   m.configFile,
		ZenDirectory: m.ZenDirectory(),
	}

	// Check if .zen directory exists (workspace initialization is based on directory, not config file)
	if _, err := os.Stat(m.ZenDirectory()); err == nil {
		status.Initialized = true
	} else {
		status.Initialized = false
	}

	// Detect project information
	status.Project = m.DetectProject()

	return status, nil
}

// createWorkspace creates the workspace directory structure
func (m *Manager) createWorkspace() error {
	m.logger.Debug("Creating workspace structure", "zen_dir", m.ZenDirectory())

	// Get workspace directory configuration
	config := DefaultWorkspaceDirectoryConfig()

	// Create main .zen directory
	if err := m.fsManager.CreateDirectory(m.ZenDirectory(), 0755); err != nil {
		return fmt.Errorf("failed to create .zen directory: %w", err)
	}

	// Create essential subdirectories
	var dirs []string
	for _, dirSpec := range config.EssentialDirectories {
		dirs = append(dirs, dirSpec.Name)
	}

	if err := m.fsManager.CreateDirectories(m.ZenDirectory(), dirs, 0755); err != nil {
		return fmt.Errorf("failed to create essential directories: %w", err)
	}

	return nil
}

// CreateTaskDirectory creates the minimal task directory structure
func (m *Manager) CreateTaskDirectory(taskDir string) error {
	m.logger.Debug("Creating task directory structure", "task_dir", taskDir)

	// Get workspace directory configuration
	config := DefaultWorkspaceDirectoryConfig()

	// Create main task directory
	if err := m.fsManager.CreateDirectory(taskDir, 0755); err != nil {
		return fmt.Errorf("failed to create task directory: %w", err)
	}

	// Create task subdirectories
	var dirs []string
	for _, dirSpec := range config.TaskDirectories {
		dirs = append(dirs, dirSpec.Name)
	}

	if err := m.fsManager.CreateDirectories(taskDir, dirs, 0755); err != nil {
		return fmt.Errorf("failed to create task directories: %w", err)
	}

	return nil
}

// CreateWorkTypeDirectory creates a work-type directory on demand
func (m *Manager) CreateWorkTypeDirectory(taskDir, workType string) error {
	dirPath := filepath.Join(taskDir, workType)
	m.logger.Debug("Creating work-type directory", "path", dirPath)

	if err := m.fsManager.CreateDirectory(dirPath, 0755); err != nil {
		return fmt.Errorf("failed to create work-type directory %s: %w", dirPath, err)
	}

	return nil
}

// GetWorkTypeDirectories returns the list of work-type directories that can be created on demand
func (m *Manager) GetWorkTypeDirectories() []string {
	config := DefaultWorkspaceDirectoryConfig()
	var dirs []string
	for _, dirSpec := range config.WorkTypeDirectories {
		dirs = append(dirs, dirSpec.Name)
	}
	return dirs
}

// updateGitignore adds .zen directory to .gitignore if it exists
func (m *Manager) updateGitignore() error {
	gitignorePath := filepath.Join(m.root, ".gitignore")

	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return nil // No .gitignore file
	}

	// Read existing .gitignore
	content, err := os.ReadFile(gitignorePath)
	if err != nil {
		return err
	}

	contentStr := string(content)

	// Check if .zen is already ignored
	lines := strings.Split(contentStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == ".zen/" || line == ".zen" {
			return nil // Already present
		}
	}

	// Add .zen to .gitignore
	if !strings.HasSuffix(contentStr, "\n") && len(contentStr) > 0 {
		contentStr += "\n"
	}
	contentStr += "\n# Zen CLI workspace directory\n.zen/\n"

	return os.WriteFile(gitignorePath, []byte(contentStr), 0644)
}

// Helper functions for project detection

func (m *Manager) isGitRepository() bool {
	gitDir := filepath.Join(m.root, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir()
	}
	return false
}

func (m *Manager) detectGitInfo() *ProjectInfo {
	// Try to read git config or remote origin
	// This is a simplified implementation
	info := &ProjectInfo{
		Metadata: make(map[string]string),
	}

	// Try to get remote origin URL
	gitConfigPath := filepath.Join(m.root, ".git", "config")
	if data, err := os.ReadFile(gitConfigPath); err == nil {
		content := string(data)
		switch {
		case strings.Contains(content, "github.com"):
			info.Metadata["git_provider"] = "github"
		case strings.Contains(content, "gitlab.com"):
			info.Metadata["git_provider"] = "gitlab"
		case strings.Contains(content, "bitbucket.org"):
			info.Metadata["git_provider"] = "bitbucket"
		}
	}

	return info
}

func (m *Manager) detectNodeJS() *ProjectInfo {
	packagePath := filepath.Join(m.root, "package.json")
	if _, err := os.Stat(packagePath); os.IsNotExist(err) {
		return nil
	}

	// Parse package.json
	data, err := os.ReadFile(packagePath)
	if err != nil {
		return nil
	}

	var pkg struct {
		Name            string            `json:"name"`
		Version         string            `json:"version"`
		Description     string            `json:"description"`
		Scripts         map[string]string `json:"scripts"`
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}

	info := &ProjectInfo{
		Name:        pkg.Name,
		Version:     pkg.Version,
		Description: pkg.Description,
		Language:    "javascript", // Default to JavaScript
		Metadata:    make(map[string]string),
	}

	// Detect framework
	if _, hasReact := pkg.Dependencies["react"]; hasReact {
		info.Framework = "react"
	} else if _, hasVue := pkg.Dependencies["vue"]; hasVue {
		info.Framework = "vue"
	} else if _, hasAngular := pkg.Dependencies["@angular/core"]; hasAngular {
		info.Framework = "angular"
	} else if _, hasExpress := pkg.Dependencies["express"]; hasExpress {
		info.Framework = "express"
	} else if _, hasNext := pkg.Dependencies["next"]; hasNext {
		info.Framework = "next.js"
	}

	// Check for TypeScript
	if _, hasTS := pkg.Dependencies["typescript"]; hasTS {
		info.Language = "typescript"
	} else if _, hasTS := pkg.DevDependencies["typescript"]; hasTS {
		info.Language = "typescript"
	}

	return info
}

func (m *Manager) detectGo() *ProjectInfo {
	goModPath := filepath.Join(m.root, "go.mod")
	if _, err := os.Stat(goModPath); os.IsNotExist(err) {
		return nil
	}

	// Parse go.mod
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return nil
	}

	lines := strings.Split(string(data), "\n")
	info := &ProjectInfo{
		Metadata: make(map[string]string),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			moduleName := strings.TrimPrefix(line, "module ")
			info.Name = filepath.Base(moduleName)
			info.Metadata["go_module"] = moduleName
		} else if strings.HasPrefix(line, "go ") {
			version := strings.TrimPrefix(line, "go ")
			info.Version = version
			info.Metadata["go_version"] = version
		}
	}

	return info
}

func (m *Manager) detectPython() *ProjectInfo {
	// Check for various Python project files
	files := []string{"setup.py", "pyproject.toml", "requirements.txt", "Pipfile", "poetry.lock"}

	for _, file := range files {
		if _, err := os.Stat(filepath.Join(m.root, file)); err == nil {
			info := &ProjectInfo{
				Name:     filepath.Base(m.root),
				Metadata: make(map[string]string),
			}
			info.Metadata["python_project_file"] = file

			// Try to parse setup.py or pyproject.toml for more details
			if file == "pyproject.toml" {
				if pyInfo := m.parsePyprojectToml(); pyInfo != nil {
					if pyInfo.Name != "" {
						info.Name = pyInfo.Name
					}
					if pyInfo.Description != "" {
						info.Description = pyInfo.Description
					}
					if pyInfo.Version != "" {
						info.Version = pyInfo.Version
					}
				}
			}

			return info
		}
	}

	return nil
}

func (m *Manager) detectRust() *ProjectInfo {
	cargoPath := filepath.Join(m.root, "Cargo.toml")
	if _, err := os.Stat(cargoPath); os.IsNotExist(err) {
		return nil
	}

	// This is a simplified TOML parser - in production you'd use a proper TOML library
	data, err := os.ReadFile(cargoPath)
	if err != nil {
		return nil
	}

	info := &ProjectInfo{
		Metadata: make(map[string]string),
	}

	lines := strings.Split(string(data), "\n")
	inPackageSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "[package]" {
			inPackageSection = true
			continue
		}
		if strings.HasPrefix(line, "[") && line != "[package]" {
			inPackageSection = false
			continue
		}

		if inPackageSection {
			switch {
			case strings.HasPrefix(line, "name = "):
				name := strings.Trim(strings.TrimPrefix(line, "name = "), `"`)
				info.Name = name
			case strings.HasPrefix(line, "version = "):
				version := strings.Trim(strings.TrimPrefix(line, "version = "), `"`)
				info.Version = version
			case strings.HasPrefix(line, "description = "):
				description := strings.Trim(strings.TrimPrefix(line, "description = "), `"`)
				info.Description = description
			}
		}
	}

	return info
}

func (m *Manager) detectJava() *ProjectInfo {
	// Check for Maven or Gradle project files
	files := []string{"pom.xml", "build.gradle", "build.gradle.kts"}

	for _, file := range files {
		if _, err := os.Stat(filepath.Join(m.root, file)); err == nil {
			info := &ProjectInfo{
				Name:     filepath.Base(m.root),
				Metadata: make(map[string]string),
			}

			if file == "pom.xml" {
				info.Framework = "maven"
				info.Metadata["build_tool"] = "maven"
			} else if strings.Contains(file, "gradle") {
				info.Framework = "gradle"
				info.Metadata["build_tool"] = "gradle"
			}

			return info
		}
	}

	return nil
}

func (m *Manager) parsePyprojectToml() *ProjectInfo {
	// This is a simplified implementation - in production you'd use a proper TOML parser
	pyprojectPath := filepath.Join(m.root, "pyproject.toml")
	data, err := os.ReadFile(pyprojectPath)
	if err != nil {
		return nil
	}

	info := &ProjectInfo{}
	lines := strings.Split(string(data), "\n")
	inProjectSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "[tool.poetry]" || line == "[project]" {
			inProjectSection = true
			continue
		}
		if strings.HasPrefix(line, "[") && !strings.Contains(line, "project") && !strings.Contains(line, "poetry") {
			inProjectSection = false
			continue
		}

		if inProjectSection {
			switch {
			case strings.HasPrefix(line, "name = "):
				name := strings.Trim(strings.TrimPrefix(line, "name = "), `"`)
				info.Name = name
			case strings.HasPrefix(line, "version = "):
				version := strings.Trim(strings.TrimPrefix(line, "version = "), `"`)
				info.Version = version
			case strings.HasPrefix(line, "description = "):
				description := strings.Trim(strings.TrimPrefix(line, "description = "), `"`)
				info.Description = description
			}
		}
	}

	return info
}
