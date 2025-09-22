package jira

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// Data Mapping and Transformation

// MapToZen converts external Jira data to Zen format
func (p *Plugin) MapToZen(ctx context.Context, externalData interface{}) (*PluginTaskData, error) {
	jiraIssue, ok := externalData.(*JiraIssue)
	if !ok {
		return nil, fmt.Errorf("expected JiraIssue, got %T", externalData)
	}

	return p.convertJiraIssueToTaskData(jiraIssue), nil
}

// MapToExternal converts Zen task data to external Jira format
func (p *Plugin) MapToExternal(ctx context.Context, zenData *PluginTaskData) (interface{}, error) {
	return p.convertTaskDataToJiraCreate(zenData), nil
}

// GetFieldMapping returns the field mapping configuration
func (p *Plugin) GetFieldMapping() *PluginFieldMappingConfig {
	return &PluginFieldMappingConfig{
		Mappings: []PluginFieldMapping{
			{
				ZenField:      "id",
				ExternalField: "key",
				Direction:     SyncDirectionBidirectional,
				Required:      true,
			},
			{
				ZenField:      "title",
				ExternalField: "fields.summary",
				Direction:     SyncDirectionBidirectional,
				Required:      true,
			},
			{
				ZenField:      "description",
				ExternalField: "fields.description",
				Direction:     SyncDirectionBidirectional,
				Required:      false,
			},
			{
				ZenField:      "status",
				ExternalField: "fields.status.name",
				Direction:     SyncDirectionBidirectional,
				Required:      true,
			},
			{
				ZenField:      "priority",
				ExternalField: "fields.priority.name",
				Direction:     SyncDirectionBidirectional,
				Required:      false,
			},
			{
				ZenField:      "assignee",
				ExternalField: "fields.assignee.displayName",
				Direction:     SyncDirectionBidirectional,
				Required:      false,
			},
			{
				ZenField:      "type",
				ExternalField: "fields.issuetype.name",
				Direction:     SyncDirectionBidirectional,
				Required:      true,
			},
		},
	}
}

// PluginFieldMappingConfig represents field mapping configuration for the plugin
type PluginFieldMappingConfig struct {
	Mappings   []PluginFieldMapping    `json:"mappings" yaml:"mappings"`
	Transforms []PluginFieldTransform  `json:"transforms" yaml:"transforms"`
	Validation []PluginFieldValidation `json:"validation" yaml:"validation"`
}

// PluginFieldMapping represents a field mapping between systems
type PluginFieldMapping struct {
	ZenField      string        `json:"zen_field" yaml:"zen_field" validate:"required"`
	ExternalField string        `json:"external_field" yaml:"external_field" validate:"required"`
	Direction     SyncDirection `json:"direction" yaml:"direction"`
	Required      bool          `json:"required" yaml:"required"`
	DefaultValue  interface{}   `json:"default_value,omitempty" yaml:"default_value,omitempty"`
}

// PluginFieldTransform represents a field transformation
type PluginFieldTransform struct {
	Field     string                 `json:"field" yaml:"field" validate:"required"`
	Type      TransformType          `json:"type" yaml:"type" validate:"required"`
	Config    map[string]interface{} `json:"config" yaml:"config"`
	Direction SyncDirection          `json:"direction" yaml:"direction"`
}

// PluginFieldValidation represents field validation rules
type PluginFieldValidation struct {
	Field string                 `json:"field" yaml:"field" validate:"required"`
	Rules map[string]interface{} `json:"rules" yaml:"rules"`
}

// TransformType represents transformation types
type TransformType string

const (
	TransformTypeMap      TransformType = "map"
	TransformTypeFormat   TransformType = "format"
	TransformTypeTemplate TransformType = "template"
	TransformTypeCustom   TransformType = "custom"
)

// Conversion methods

// convertJiraIssueToTaskData converts a Jira issue to standard task data
func (p *Plugin) convertJiraIssueToTaskData(issue *JiraIssue) *PluginTaskData {
	taskData := &PluginTaskData{
		ID:          issue.Key,
		ExternalID:  issue.Key,
		Title:       issue.Fields.Summary,
		Description: issue.Fields.Description,
		Status:      p.mapJiraStatusToZen(issue.Fields.Status.Name),
		Priority:    p.mapJiraPriorityToZen(issue.Fields.Priority.Name),
		Type:        p.mapJiraTypeToZen(issue.Fields.IssueType.Name),
		Assignee:    issue.Fields.Assignee.DisplayName,
		Created:     issue.Fields.Created,
		Updated:     issue.Fields.Updated,
		ExternalURL: p.buildJiraIssueURL(issue.Key),
		Metadata: map[string]interface{}{
			"external_system": "jira",
			"jira_id":         issue.ID,
			"project_key":     p.config.ProjectKey,
		},
		Version:  1,
		Checksum: p.generateChecksum(issue),
	}

	// Extract labels if present
	if len(issue.Fields.Labels) > 0 {
		taskData.Labels = issue.Fields.Labels
	}

	// Extract components if present
	if len(issue.Fields.Components) > 0 {
		components := make([]string, len(issue.Fields.Components))
		for i, comp := range issue.Fields.Components {
			components[i] = comp.Name
		}
		taskData.Components = components
	}

	return taskData
}

// convertTaskDataToJiraCreate converts task data to Jira create request
func (p *Plugin) convertTaskDataToJiraCreate(taskData *PluginTaskData) *JiraCreateRequest {
	return &JiraCreateRequest{
		Fields: JiraCreateFields{
			Project: JiraProject{
				Key: p.config.ProjectKey,
			},
			Summary:     taskData.Title,
			Description: taskData.Description,
			IssueType: JiraIssueType{
				Name: p.mapZenTypeToJira(taskData.Type),
			},
			Priority: &JiraPriority{
				Name: p.mapZenPriorityToJira(taskData.Priority),
			},
			Labels: taskData.Labels,
		},
	}
}

// convertTaskDataToJiraUpdate converts task data to Jira update request
func (p *Plugin) convertTaskDataToJiraUpdate(taskData *PluginTaskData) *JiraUpdateRequest {
	fields := make(map[string]interface{})

	if taskData.Title != "" {
		fields["summary"] = taskData.Title
	}
	if taskData.Description != "" {
		fields["description"] = taskData.Description
	}
	if taskData.Status != "" {
		fields["status"] = map[string]interface{}{
			"name": p.mapZenStatusToJira(taskData.Status),
		}
	}
	if taskData.Priority != "" {
		fields["priority"] = map[string]interface{}{
			"name": p.mapZenPriorityToJira(taskData.Priority),
		}
	}
	if len(taskData.Labels) > 0 {
		fields["labels"] = taskData.Labels
	}

	return &JiraUpdateRequest{
		Fields: fields,
	}
}

// Mapping functions

func (p *Plugin) mapJiraStatusToZen(jiraStatus string) string {
	statusMap := map[string]string{
		"To Do":       "proposed",
		"Open":        "proposed",
		"New":         "proposed",
		"In Progress": "in_progress",
		"Doing":       "in_progress",
		"Done":        "completed",
		"Closed":      "completed",
		"Resolved":    "completed",
		"Blocked":     "blocked",
	}

	if zenStatus, ok := statusMap[jiraStatus]; ok {
		return zenStatus
	}

	// Default mapping for unknown statuses
	return strings.ToLower(strings.ReplaceAll(jiraStatus, " ", "_"))
}

func (p *Plugin) mapZenStatusToJira(zenStatus string) string {
	statusMap := map[string]string{
		"proposed":    "To Do",
		"in_progress": "In Progress",
		"completed":   "Done",
		"blocked":     "Blocked",
	}

	if jiraStatus, ok := statusMap[zenStatus]; ok {
		return jiraStatus
	}

	// Default mapping for unknown statuses
	return zenStatus
}

func (p *Plugin) mapJiraPriorityToZen(jiraPriority string) string {
	priorityMap := map[string]string{
		"Highest":  "P0",
		"Critical": "P0",
		"High":     "P1",
		"Medium":   "P2",
		"Low":      "P3",
		"Lowest":   "P3",
	}

	if zenPriority, ok := priorityMap[jiraPriority]; ok {
		return zenPriority
	}

	return "P2" // Default to medium priority
}

func (p *Plugin) mapZenPriorityToJira(zenPriority string) string {
	priorityMap := map[string]string{
		"P0": "Highest",
		"P1": "High",
		"P2": "Medium",
		"P3": "Low",
	}

	if jiraPriority, ok := priorityMap[zenPriority]; ok {
		return jiraPriority
	}

	return "Medium" // Default to medium priority
}

func (p *Plugin) mapJiraTypeToZen(jiraType string) string {
	typeMap := map[string]string{
		"Story":      "story",
		"User Story": "story",
		"Bug":        "bug",
		"Defect":     "bug",
		"Epic":       "epic",
		"Initiative": "epic",
		"Task":       "task",
		"Sub-task":   "task",
		"Subtask":    "task",
		"Spike":      "spike",
		"Research":   "spike",
	}

	if zenType, ok := typeMap[jiraType]; ok {
		return zenType
	}

	return "task" // Default to task type
}

func (p *Plugin) mapZenTypeToJira(zenType string) string {
	typeMap := map[string]string{
		"story": "Story",
		"bug":   "Bug",
		"epic":  "Epic",
		"task":  "Task",
		"spike": "Spike",
	}

	if jiraType, ok := typeMap[zenType]; ok {
		return jiraType
	}

	return "Task" // Default to task type
}

// Helper methods

func (p *Plugin) buildJiraIssueURL(issueKey string) string {
	baseURL := strings.TrimRight(p.config.BaseURL, "/")
	return fmt.Sprintf("%s/browse/%s", baseURL, issueKey)
}

func (p *Plugin) generateChecksum(issue *JiraIssue) string {
	// Simple checksum based on key fields
	// In production, this would be a proper hash
	return fmt.Sprintf("%s-%s-%d",
		issue.Key,
		issue.Fields.Updated.Format(time.RFC3339),
		len(issue.Fields.Summary))
}
