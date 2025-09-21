package workspace

// WorkspaceDirectoryConfig defines the directory structure for a Zen workspace
type WorkspaceDirectoryConfig struct {
	// Essential directories created during workspace initialization
	EssentialDirectories []DirectorySpec `json:"essential_directories" yaml:"essential_directories"`

	// Task-specific directories created on demand
	TaskDirectories []DirectorySpec `json:"task_directories" yaml:"task_directories"`

	// Work-type directories created on demand
	WorkTypeDirectories []DirectorySpec `json:"work_type_directories" yaml:"work_type_directories"`
}

// DirectorySpec defines a directory specification
type DirectorySpec struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Permissions uint32 `json:"permissions" yaml:"permissions"` // Octal permissions (e.g., 0755)
}

// DefaultWorkspaceDirectoryConfig returns the default workspace directory configuration
func DefaultWorkspaceDirectoryConfig() WorkspaceDirectoryConfig {
	return WorkspaceDirectoryConfig{
		EssentialDirectories: []DirectorySpec{
			{
				Name:        "library",
				Description: "Library cache and manifest storage",
				Permissions: 0755,
			},
			{
				Name:        "cache",
				Description: "CLI caches and temporary data",
				Permissions: 0755,
			},
			{
				Name:        "logs",
				Description: "CLI run logs and sync traces",
				Permissions: 0755,
			},
			{
				Name:        "work",
				Description: "Work directory (tasks created here)",
				Permissions: 0755,
			},
			{
				Name:        "metadata",
				Description: "External system integration data",
				Permissions: 0755,
			},
		},
		TaskDirectories: []DirectorySpec{
			{
				Name:        ".zenflow",
				Description: "Zenflow state tracking",
				Permissions: 0755,
			},
			{
				Name:        "metadata",
				Description: "External system snapshots",
				Permissions: 0755,
			},
		},
		WorkTypeDirectories: []DirectorySpec{
			{
				Name:        "research",
				Description: "Investigation and discovery work",
				Permissions: 0755,
			},
			{
				Name:        "spikes",
				Description: "Technical exploration and prototyping",
				Permissions: 0755,
			},
			{
				Name:        "design",
				Description: "Specifications and planning artifacts",
				Permissions: 0755,
			},
			{
				Name:        "execution",
				Description: "Implementation evidence and results",
				Permissions: 0755,
			},
			{
				Name:        "outcomes",
				Description: "Learning, metrics, and retrospectives",
				Permissions: 0755,
			},
		},
	}
}
