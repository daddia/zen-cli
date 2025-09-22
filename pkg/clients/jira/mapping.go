package jira

import (
	"strings"
)

// Legacy Field Mapping - DEPRECATED
// This file contains legacy field mapping functionality that will be replaced
// by the new standardized plugin architecture in data_mapping.go

// FieldMapping defines how Jira fields map to Zen template variables
// DEPRECATED: Use FieldMappingConfig from data_mapping.go instead
type FieldMapping struct {
	// Source field path in Jira response (dot notation for nested fields)
	JiraField string

	// Target template variable name
	ZenVariable string

	// Optional transformation function
	Transform func(interface{}) interface{}

	// Whether this field is required
	Required bool
}

// GetFieldMappings returns the complete mapping of Jira fields to Zen template variables
func GetFieldMappings() []FieldMapping {
	return []FieldMapping{
		// Basic task information
		{
			JiraField:   "key",
			ZenVariable: "TASK_ID",
			Required:    true,
		},
		{
			JiraField:   "fields.summary",
			ZenVariable: "TASK_TITLE",
			Required:    true,
		},
		{
			JiraField:   "fields.issuetype.name",
			ZenVariable: "TASK_TYPE",
			Transform:   mapIssueTypeToZen,
			Required:    true,
		},
		{
			JiraField:   "fields.status.name",
			ZenVariable: "TASK_STATUS",
			Transform:   mapStatusToZen,
		},
		{
			JiraField:   "fields.priority.name",
			ZenVariable: "PRIORITY",
			Transform:   mapPriorityToZen,
		},

		// Ownership and team
		{
			JiraField:   "fields.assignee.displayName",
			ZenVariable: "OWNER_NAME",
		},
		{
			JiraField:   "fields.assignee.emailAddress",
			ZenVariable: "OWNER_EMAIL",
		},
		{
			JiraField:   "fields.assignee.displayName",
			ZenVariable: "GITHUB_USERNAME",
			Transform:   extractUsernameFromDisplayName,
		},
		{
			JiraField:   "fields.project.name",
			ZenVariable: "TEAM_NAME",
		},

		// Dates
		{
			JiraField:   "fields.created",
			ZenVariable: "JIRA_CREATED_RAW",
		},
		{
			JiraField:   "fields.updated",
			ZenVariable: "JIRA_UPDATED_RAW",
		},

		// Additional Jira-specific fields for enrichment
		{
			JiraField:   "fields.creator.displayName",
			ZenVariable: "JIRA_CREATOR",
		},
		{
			JiraField:   "fields.creator.emailAddress",
			ZenVariable: "JIRA_CREATOR_EMAIL",
		},
		{
			JiraField:   "fields.reporter.displayName",
			ZenVariable: "JIRA_REPORTER",
		},
		{
			JiraField:   "fields.reporter.emailAddress",
			ZenVariable: "JIRA_REPORTER_EMAIL",
		},
		{
			JiraField:   "fields.project.key",
			ZenVariable: "JIRA_PROJECT_KEY",
		},
		{
			JiraField:   "fields.project.projectTypeKey",
			ZenVariable: "JIRA_PROJECT_TYPE",
		},
		{
			JiraField:   "fields.issuetype.description",
			ZenVariable: "JIRA_ISSUE_TYPE_DESCRIPTION",
		},
		{
			JiraField:   "fields.status.statusCategory.name",
			ZenVariable: "JIRA_STATUS_CATEGORY",
		},
		{
			JiraField:   "fields.priority.iconUrl",
			ZenVariable: "JIRA_PRIORITY_ICON",
		},
		{
			JiraField:   "self",
			ZenVariable: "JIRA_URL",
		},
		{
			JiraField:   "fields.description",
			ZenVariable: "JIRA_DESCRIPTION_RAW",
		},
		{
			JiraField:   "fields.labels",
			ZenVariable: "LABELS",
			Transform:   mergeWithExistingLabels,
		},
		{
			JiraField:   "fields.components",
			ZenVariable: "JIRA_COMPONENTS",
			Transform:   extractComponentNames,
		},
		{
			JiraField:   "fields.fixVersions",
			ZenVariable: "JIRA_FIX_VERSIONS",
			Transform:   extractVersionNames,
		},
		{
			JiraField:   "fields.timespent",
			ZenVariable: "JIRA_TIME_SPENT",
		},
		{
			JiraField:   "fields.timeestimate",
			ZenVariable: "JIRA_TIME_ESTIMATE",
		},
		{
			JiraField:   "fields.issuelinks",
			ZenVariable: "JIRA_LINKED_ISSUES_COUNT",
			Transform:   countArrayItems,
		},
		{
			JiraField:   "fields.subtasks",
			ZenVariable: "JIRA_SUBTASKS_COUNT",
			Transform:   countArrayItems,
		},
	}
}

// ApplyFieldMappings applies Jira field mappings to template variables
func ApplyFieldMappings(variables map[string]interface{}, jiraData map[string]interface{}) {
	mappings := GetFieldMappings()

	for _, mapping := range mappings {
		value := getNestedValue(jiraData, mapping.JiraField)

		// Skip if value is nil and field is not required
		if value == nil && !mapping.Required {
			continue
		}

		// Apply transformation if provided
		if mapping.Transform != nil {
			value = mapping.Transform(value)
		}

		// Set the variable if we have a value
		if value != nil {
			variables[mapping.ZenVariable] = value
		}
	}

	// Add integration flags
	variables["JIRA_INTEGRATION"] = true
	variables["EXTERNAL_SYSTEM"] = "jira"
	variables["SYNC_ENABLED"] = true

	// Extract and add custom fields
	if fields, ok := jiraData["fields"].(map[string]interface{}); ok {
		customFields := make(map[string]interface{})
		for key, value := range fields {
			if strings.HasPrefix(key, "customfield_") && value != nil {
				customFields[key] = value
			}
		}
		if len(customFields) > 0 {
			variables["JIRA_CUSTOM_FIELDS"] = customFields
		}
	}
}

// Helper function to get nested values using dot notation
func getNestedValue(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			// Last part - return the value
			return current[part]
		}

		// Navigate deeper
		if next, ok := current[part].(map[string]interface{}); ok {
			current = next
		} else {
			return nil
		}
	}

	return nil
}

// Transformation functions

func mapIssueTypeToZen(value interface{}) interface{} {
	if issueType, ok := value.(string); ok {
		switch strings.ToLower(issueType) {
		case "story", "user story":
			return "story"
		case "bug", "defect":
			return "bug"
		case "epic", "initiative":
			return "epic"
		case "spike", "research":
			return "spike"
		default:
			return "task"
		}
	}
	return "task"
}

func mapStatusToZen(value interface{}) interface{} {
	if status, ok := value.(string); ok {
		switch strings.ToLower(status) {
		case "to do", "open", "new":
			return "proposed"
		case "in progress", "doing":
			return "in_progress"
		case "done", "closed", "resolved":
			return "completed"
		case "blocked":
			return "blocked"
		default:
			return "proposed"
		}
	}
	return "proposed"
}

func mapPriorityToZen(value interface{}) interface{} {
	if priority, ok := value.(string); ok {
		switch strings.ToLower(priority) {
		case "highest", "critical":
			return "P0"
		case "high":
			return "P1"
		case "medium":
			return "P2"
		case "low", "lowest":
			return "P3"
		default:
			return "P2"
		}
	}
	return "P2"
}

func extractUsernameFromDisplayName(value interface{}) interface{} {
	if displayName, ok := value.(string); ok {
		// Convert "Jonathan Daddia" to "jonathandaddia"
		username := strings.ToLower(strings.ReplaceAll(displayName, " ", ""))
		return username
	}
	return value
}

func mergeWithExistingLabels(value interface{}) interface{} {
	if labels, ok := value.([]interface{}); ok {
		labelStrings := make([]string, 0, len(labels))
		for _, label := range labels {
			if labelStr, ok := label.(string); ok {
				labelStrings = append(labelStrings, labelStr)
			}
		}
		return labelStrings
	}
	return []string{}
}

func extractComponentNames(value interface{}) interface{} {
	if components, ok := value.([]interface{}); ok {
		componentNames := make([]string, 0, len(components))
		for _, component := range components {
			if comp, ok := component.(map[string]interface{}); ok {
				if name, ok := comp["name"].(string); ok {
					componentNames = append(componentNames, name)
				}
			}
		}
		return componentNames
	}
	return []string{}
}

func extractVersionNames(value interface{}) interface{} {
	if versions, ok := value.([]interface{}); ok {
		versionNames := make([]string, 0, len(versions))
		for _, version := range versions {
			if ver, ok := version.(map[string]interface{}); ok {
				if name, ok := ver["name"].(string); ok {
					versionNames = append(versionNames, name)
				}
			}
		}
		return versionNames
	}
	return []string{}
}

func countArrayItems(value interface{}) interface{} {
	if array, ok := value.([]interface{}); ok {
		return len(array)
	}
	return 0
}
