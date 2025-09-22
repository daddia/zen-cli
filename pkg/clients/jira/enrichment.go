package jira

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Legacy Enrichment - DEPRECATED
// This file contains legacy enrichment functionality that will be replaced
// by the new standardized plugin architecture in data_mapping.go

// EnrichmentData contains enriched data extracted from Jira
// DEPRECATED: Use TaskData from plugin.go instead
type EnrichmentData struct {
	// Basic information
	Key     string
	ID      string
	URL     string
	Summary string

	// Issue details
	IssueType      string
	IssueTypeDesc  string
	Status         string
	StatusCategory string
	Priority       string
	PriorityIcon   string

	// People
	Assignee      string
	AssigneeEmail string
	AssigneeID    string
	Reporter      string
	ReporterEmail string
	Creator       string
	CreatorEmail  string

	// Project
	ProjectName string
	ProjectKey  string
	ProjectType string

	// Dates
	Created          time.Time
	Updated          time.Time
	CreatedFormatted string
	UpdatedFormatted string

	// Content
	Description      map[string]interface{}
	DescriptionPlain string

	// Organization
	Labels      []string
	Components  []string
	FixVersions []string

	// Time tracking
	TimeSpent    interface{}
	TimeEstimate interface{}

	// Relationships
	IssueLinks        []interface{}
	LinkedIssuesCount int
	Subtasks          []interface{}
	SubtasksCount     int

	// Custom fields
	CustomFields map[string]interface{}

	// Zen mappings
	ZenTaskType string
	ZenStatus   string
	ZenPriority string
}

// LoadJiraData loads and parses Jira data from a task's metadata file
func LoadJiraData(taskDir string) (map[string]interface{}, error) {
	jiraFilePath := filepath.Join(taskDir, "metadata", "jira.json")

	// Check if file exists
	if _, err := os.Stat(jiraFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("jira metadata file not found")
	}

	// Read and parse the file
	data, err := os.ReadFile(jiraFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read jira metadata: %w", err)
	}

	var jiraData map[string]interface{}
	if err := json.Unmarshal(data, &jiraData); err != nil {
		return nil, fmt.Errorf("failed to parse jira metadata: %w", err)
	}

	return jiraData, nil
}

// ExtractEnrichmentData extracts structured enrichment data from raw Jira response
func ExtractEnrichmentData(jiraData map[string]interface{}) *EnrichmentData {
	enrichment := &EnrichmentData{}

	// Extract basic information
	if key, ok := jiraData["key"].(string); ok {
		enrichment.Key = key
	}
	if id, ok := jiraData["id"].(string); ok {
		enrichment.ID = id
	}
	if self, ok := jiraData["self"].(string); ok {
		enrichment.URL = self
	}

	fields, ok := jiraData["fields"].(map[string]interface{})
	if !ok {
		return enrichment
	}

	// Extract summary
	if summary, ok := fields["summary"].(string); ok {
		enrichment.Summary = summary
	}

	// Extract issue type
	if issueType, ok := fields["issuetype"].(map[string]interface{}); ok {
		if name, ok := issueType["name"].(string); ok {
			enrichment.IssueType = name
			enrichment.ZenTaskType = mapJiraIssueTypeToZen(name)
		}
		if description, ok := issueType["description"].(string); ok {
			enrichment.IssueTypeDesc = description
		}
	}

	// Extract status
	if status, ok := fields["status"].(map[string]interface{}); ok {
		if name, ok := status["name"].(string); ok {
			enrichment.Status = name
			enrichment.ZenStatus = mapJiraStatusToZen(name)
		}
		if statusCategory, ok := status["statusCategory"].(map[string]interface{}); ok {
			if categoryName, ok := statusCategory["name"].(string); ok {
				enrichment.StatusCategory = categoryName
			}
		}
	}

	// Extract priority
	if priority, ok := fields["priority"].(map[string]interface{}); ok {
		if name, ok := priority["name"].(string); ok {
			enrichment.Priority = name
			enrichment.ZenPriority = mapJiraPriorityToZen(name)
		}
		if iconUrl, ok := priority["iconUrl"].(string); ok {
			enrichment.PriorityIcon = iconUrl
		}
	}

	// Extract people
	if assignee, ok := fields["assignee"].(map[string]interface{}); ok {
		if displayName, ok := assignee["displayName"].(string); ok {
			enrichment.Assignee = displayName
		}
		if emailAddress, ok := assignee["emailAddress"].(string); ok {
			enrichment.AssigneeEmail = emailAddress
		}
		if accountId, ok := assignee["accountId"].(string); ok {
			enrichment.AssigneeID = accountId
		}
	}

	if reporter, ok := fields["reporter"].(map[string]interface{}); ok {
		if displayName, ok := reporter["displayName"].(string); ok {
			enrichment.Reporter = displayName
		}
		if emailAddress, ok := reporter["emailAddress"].(string); ok {
			enrichment.ReporterEmail = emailAddress
		}
	}

	if creator, ok := fields["creator"].(map[string]interface{}); ok {
		if displayName, ok := creator["displayName"].(string); ok {
			enrichment.Creator = displayName
		}
		if emailAddress, ok := creator["emailAddress"].(string); ok {
			enrichment.CreatorEmail = emailAddress
		}
	}

	// Extract project
	if project, ok := fields["project"].(map[string]interface{}); ok {
		if name, ok := project["name"].(string); ok {
			enrichment.ProjectName = name
		}
		if key, ok := project["key"].(string); ok {
			enrichment.ProjectKey = key
		}
		if projectType, ok := project["projectTypeKey"].(string); ok {
			enrichment.ProjectType = projectType
		}
	}

	// Extract dates
	if created, ok := fields["created"].(string); ok {
		if createdTime, err := parseJiraTimestamp(created); err == nil {
			enrichment.Created = createdTime
			enrichment.CreatedFormatted = createdTime.Format("January 2, 2006")
		}
	}
	if updated, ok := fields["updated"].(string); ok {
		if updatedTime, err := parseJiraTimestamp(updated); err == nil {
			enrichment.Updated = updatedTime
			enrichment.UpdatedFormatted = updatedTime.Format("January 2, 2006")
		}
	}

	// Extract description
	if description, ok := fields["description"].(map[string]interface{}); ok {
		enrichment.Description = description
		enrichment.DescriptionPlain = extractPlainTextFromJiraDescription(description)
	}

	// Extract labels
	if labels, ok := fields["labels"].([]interface{}); ok {
		labelStrings := make([]string, 0, len(labels))
		for _, label := range labels {
			if labelStr, ok := label.(string); ok {
				labelStrings = append(labelStrings, labelStr)
			}
		}
		enrichment.Labels = labelStrings
	}

	// Extract components
	if components, ok := fields["components"].([]interface{}); ok {
		componentNames := make([]string, 0, len(components))
		for _, component := range components {
			if comp, ok := component.(map[string]interface{}); ok {
				if name, ok := comp["name"].(string); ok {
					componentNames = append(componentNames, name)
				}
			}
		}
		enrichment.Components = componentNames
	}

	// Extract fix versions
	if fixVersions, ok := fields["fixVersions"].([]interface{}); ok {
		versionNames := make([]string, 0, len(fixVersions))
		for _, version := range fixVersions {
			if ver, ok := version.(map[string]interface{}); ok {
				if name, ok := ver["name"].(string); ok {
					versionNames = append(versionNames, name)
				}
			}
		}
		enrichment.FixVersions = versionNames
	}

	// Extract time tracking
	if timeSpent, ok := fields["timespent"]; ok && timeSpent != nil {
		enrichment.TimeSpent = timeSpent
	}
	if timeEstimate, ok := fields["timeestimate"]; ok && timeEstimate != nil {
		enrichment.TimeEstimate = timeEstimate
	}

	// Extract issue links
	if issueLinks, ok := fields["issuelinks"].([]interface{}); ok {
		enrichment.IssueLinks = issueLinks
		enrichment.LinkedIssuesCount = len(issueLinks)
	}

	// Extract subtasks
	if subtasks, ok := fields["subtasks"].([]interface{}); ok {
		enrichment.Subtasks = subtasks
		enrichment.SubtasksCount = len(subtasks)
	}

	// Extract custom fields
	customFields := make(map[string]interface{})
	for key, value := range fields {
		if strings.HasPrefix(key, "customfield_") && value != nil {
			customFields[key] = value
		}
	}
	enrichment.CustomFields = customFields

	return enrichment
}

// EnrichTemplateVariables enriches template variables with Jira data using field mappings
func EnrichTemplateVariables(variables map[string]interface{}, jiraData map[string]interface{}) {
	ApplyFieldMappings(variables, jiraData)
}

// Helper functions for mapping Jira values to Zen values

func mapJiraIssueTypeToZen(issueType string) string {
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

func mapJiraStatusToZen(status string) string {
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

func mapJiraPriorityToZen(priority string) string {
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

func parseJiraTimestamp(timestamp string) (time.Time, error) {
	formats := []string{
		"2006-01-02T15:04:05.000-0700", // Jira format: 2025-09-21T22:24:41.965+1000
		"2006-01-02T15:04:05-0700",     // Without milliseconds
		time.RFC3339,                   // Standard format
		time.RFC3339Nano,               // With nanoseconds
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timestamp); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestamp)
}

func extractPlainTextFromJiraDescription(description map[string]interface{}) string {
	content, ok := description["content"].([]interface{})
	if !ok {
		return ""
	}

	var plainText strings.Builder
	for _, block := range content {
		if blockMap, ok := block.(map[string]interface{}); ok {
			if blockContent, ok := blockMap["content"].([]interface{}); ok {
				for _, textNode := range blockContent {
					if textMap, ok := textNode.(map[string]interface{}); ok {
						if text, ok := textMap["text"].(string); ok {
							plainText.WriteString(text)
						}
					}
				}
				plainText.WriteString("\n")
			}
		}
	}

	return strings.TrimSpace(plainText.String())
}
