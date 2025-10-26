package config

// ComponentRegistry manages available configuration components
type ComponentRegistry struct {
	components map[string]ComponentInfo
}

// ComponentInfo contains information about a configuration component
type ComponentInfo struct {
	Name        string
	Description string
}

// NewComponentRegistry creates a new component registry with all available components
func NewComponentRegistry() *ComponentRegistry {
	return &ComponentRegistry{
		components: map[string]ComponentInfo{
			"assets": {
				Name:        "assets",
				Description: "Asset repository configuration",
			},
			"auth": {
				Name:        "auth",
				Description: "Authentication configuration",
			},
			"cache": {
				Name:        "cache",
				Description: "Cache configuration",
			},
			"cli": {
				Name:        "cli",
				Description: "CLI interface configuration",
			},
			"development": {
				Name:        "development",
				Description: "Development settings",
			},
			"task": {
				Name:        "task",
				Description: "Task management configuration",
			},
			"templates": {
				Name:        "templates",
				Description: "Template engine configuration",
			},
			"workspace": {
				Name:        "workspace",
				Description: "Workspace configuration",
			},
		},
	}
}

// GetComponent returns component info for the given component name
func (r *ComponentRegistry) GetComponent(name string) (ComponentInfo, bool) {
	info, exists := r.components[name]
	return info, exists
}

// ListComponents returns all available component names
func (r *ComponentRegistry) ListComponents() []string {
	components := make([]string, 0, len(r.components))
	for name := range r.components {
		components = append(components, name)
	}
	return components
}

// IsValidComponent checks if a component name is valid
func (r *ComponentRegistry) IsValidComponent(name string) bool {
	_, exists := r.components[name]
	return exists
}
