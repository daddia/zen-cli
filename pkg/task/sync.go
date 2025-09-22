package task

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daddia/zen/pkg/templates"
)

// Source-specific data synchronization methods

// syncJiraSpecificData extracts rich Jira-specific data from raw response
func (m *Manager) syncJiraSpecificData(variables map[string]interface{}, rawData map[string]interface{}) {
	// Extract fields object
	fields, ok := rawData["fields"].(map[string]interface{})
	if !ok {
		return
	}

	// Extract assignee details
	if assignee, ok := fields["assignee"].(map[string]interface{}); ok {
		if displayName, ok := assignee["displayName"].(string); ok {
			variables["JIRA_ASSIGNEE"] = displayName
			variables["OWNER_NAME"] = displayName // Override with real assignee
		}
		if email, ok := assignee["emailAddress"].(string); ok {
			variables["JIRA_ASSIGNEE_EMAIL"] = email
			variables["OWNER_EMAIL"] = email // Override with real email
		}
		if accountId, ok := assignee["accountId"].(string); ok {
			variables["JIRA_ASSIGNEE_ID"] = accountId
		}
	}

	// Extract reporter details
	if reporter, ok := fields["reporter"].(map[string]interface{}); ok {
		if displayName, ok := reporter["displayName"].(string); ok {
			variables["JIRA_REPORTER"] = displayName
		}
		if email, ok := reporter["emailAddress"].(string); ok {
			variables["JIRA_REPORTER_EMAIL"] = email
		}
	}

	// Extract creator details
	if creator, ok := fields["creator"].(map[string]interface{}); ok {
		if displayName, ok := creator["displayName"].(string); ok {
			variables["JIRA_CREATOR"] = displayName
		}
		if email, ok := creator["emailAddress"].(string); ok {
			variables["JIRA_CREATOR_EMAIL"] = email
		}
	}

	// Extract project details
	if project, ok := fields["project"].(map[string]interface{}); ok {
		if name, ok := project["name"].(string); ok {
			variables["JIRA_PROJECT_NAME"] = name
			variables["TEAM_NAME"] = name // Override with real project name
		}
		if key, ok := project["key"].(string); ok {
			variables["JIRA_PROJECT_KEY"] = key
		}
		if projectType, ok := project["projectTypeKey"].(string); ok {
			variables["JIRA_PROJECT_TYPE"] = projectType
		}
	}

	// Extract status details
	if status, ok := fields["status"].(map[string]interface{}); ok {
		if name, ok := status["name"].(string); ok {
			variables["JIRA_STATUS"] = name
			variables["TASK_STATUS"] = m.mapExternalStatusToZenStatus(name, "jira")
		}
		if statusCategory, ok := status["statusCategory"].(map[string]interface{}); ok {
			if categoryName, ok := statusCategory["name"].(string); ok {
				variables["JIRA_STATUS_CATEGORY"] = categoryName
			}
		}
	}

	// Extract priority details
	if priority, ok := fields["priority"].(map[string]interface{}); ok {
		if name, ok := priority["name"].(string); ok {
			variables["JIRA_PRIORITY"] = name
			variables["PRIORITY"] = m.mapExternalPriorityToZenPriority(name, "jira")
		}
		if iconUrl, ok := priority["iconUrl"].(string); ok {
			variables["JIRA_PRIORITY_ICON"] = iconUrl
		}
	}

	// Extract issue type details
	if issueType, ok := fields["issuetype"].(map[string]interface{}); ok {
		if name, ok := issueType["name"].(string); ok {
			variables["JIRA_ISSUE_TYPE"] = name
			variables["TASK_TYPE"] = m.mapJiraIssueTypeToZenType(name)
		}
		if description, ok := issueType["description"].(string); ok {
			variables["JIRA_ISSUE_TYPE_DESCRIPTION"] = description
		}
		if iconUrl, ok := issueType["iconUrl"].(string); ok {
			variables["JIRA_ISSUE_TYPE_ICON"] = iconUrl
		}
	}

	// Extract description (rich text structure)
	if description, ok := fields["description"].(map[string]interface{}); ok {
		variables["JIRA_DESCRIPTION_RAW"] = description
		if plainText := m.extractPlainTextFromJiraDescription(description); plainText != "" {
			variables["JIRA_DESCRIPTION_PLAIN"] = plainText
		}
	}

	// Extract summary (title)
	if summary, ok := fields["summary"].(string); ok {
		variables["JIRA_SUMMARY"] = summary
		variables["TASK_TITLE"] = summary // Override with real title
	}

	// Extract labels
	if labels, ok := fields["labels"].([]interface{}); ok {
		labelStrings := make([]string, 0, len(labels))
		for _, label := range labels {
			if labelStr, ok := label.(string); ok {
				labelStrings = append(labelStrings, labelStr)
			}
		}
		if len(labelStrings) > 0 {
			variables["LABELS"] = labelStrings
			variables["JIRA_LABELS"] = labelStrings
		}
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
		if len(componentNames) > 0 {
			variables["JIRA_COMPONENTS"] = componentNames
		}
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
		if len(versionNames) > 0 {
			variables["JIRA_FIX_VERSIONS"] = versionNames
		}
	}

	// Extract time tracking
	if timeSpent, ok := fields["timespent"]; ok && timeSpent != nil {
		variables["JIRA_TIME_SPENT"] = timeSpent
	}
	if timeEstimate, ok := fields["timeestimate"]; ok && timeEstimate != nil {
		variables["JIRA_TIME_ESTIMATE"] = timeEstimate
	}

	// Extract issue links count
	if issueLinks, ok := fields["issuelinks"].([]interface{}); ok {
		variables["JIRA_LINKED_ISSUES_COUNT"] = len(issueLinks)
		if len(issueLinks) > 0 {
			variables["JIRA_ISSUE_LINKS"] = issueLinks
		}
	}

	// Extract subtasks count
	if subtasks, ok := fields["subtasks"].([]interface{}); ok {
		variables["JIRA_SUBTASKS_COUNT"] = len(subtasks)
		if len(subtasks) > 0 {
			variables["JIRA_SUBTASKS"] = subtasks
		}
	}

	// Extract custom fields
	customFields := make(map[string]interface{})
	for key, value := range fields {
		if strings.HasPrefix(key, "customfield_") && value != nil {
			customFields[key] = value
		}
	}
	if len(customFields) > 0 {
		variables["JIRA_CUSTOM_FIELDS"] = customFields
	}

	// Extract dates
	if created, ok := fields["created"].(string); ok {
		if createdTime, err := time.Parse(time.RFC3339, created); err == nil {
			variables["CREATED_DATE"] = createdTime.Format("2006-01-02")
			variables["JIRA_CREATED_RAW"] = created
			variables["JIRA_CREATED_FORMATTED"] = createdTime.Format("January 2, 2006")
		}
	}
	if updated, ok := fields["updated"].(string); ok {
		if updatedTime, err := time.Parse(time.RFC3339, updated); err == nil {
			variables["LAST_UPDATED"] = updatedTime.Format("2006-01-02 15:04:05")
			variables["JIRA_UPDATED_RAW"] = updated
			variables["JIRA_UPDATED_FORMATTED"] = updatedTime.Format("January 2, 2006")
		}
	}

	// Add Jira-specific URLs and references
	if key, ok := rawData["key"].(string); ok {
		variables["JIRA_KEY"] = key
		variables["JIRA_EXTERNAL_ID"] = key
	}
	if self, ok := rawData["self"].(string); ok {
		variables["JIRA_API_URL"] = self
	}
	if id, ok := rawData["id"].(string); ok {
		variables["JIRA_ID"] = id
	}
}

// syncGitHubSpecificData extracts GitHub-specific data from raw response
func (m *Manager) syncGitHubSpecificData(variables map[string]interface{}, rawData map[string]interface{}) {
	// Extract GitHub issue/PR specific fields
	if number, ok := rawData["number"].(float64); ok {
		variables["GITHUB_NUMBER"] = int(number)
		variables["GITHUB_EXTERNAL_ID"] = fmt.Sprintf("%.0f", number)
	}

	if htmlURL, ok := rawData["html_url"].(string); ok {
		variables["GITHUB_URL"] = htmlURL
	}

	if state, ok := rawData["state"].(string); ok {
		variables["GITHUB_STATE"] = state
		variables["TASK_STATUS"] = m.mapExternalStatusToZenStatus(state, "github")
	}

	if user, ok := rawData["user"].(map[string]interface{}); ok {
		if login, ok := user["login"].(string); ok {
			variables["GITHUB_AUTHOR"] = login
			variables["GITHUB_USERNAME"] = login
		}
	}

	if assignees, ok := rawData["assignees"].([]interface{}); ok && len(assignees) > 0 {
		if assignee, ok := assignees[0].(map[string]interface{}); ok {
			if login, ok := assignee["login"].(string); ok {
				variables["GITHUB_ASSIGNEE"] = login
				variables["OWNER_NAME"] = login
			}
		}
	}

	if labels, ok := rawData["labels"].([]interface{}); ok {
		labelNames := make([]string, 0, len(labels))
		for _, label := range labels {
			if labelMap, ok := label.(map[string]interface{}); ok {
				if name, ok := labelMap["name"].(string); ok {
					labelNames = append(labelNames, name)
				}
			}
		}
		if len(labelNames) > 0 {
			variables["LABELS"] = labelNames
			variables["GITHUB_LABELS"] = labelNames
		}
	}
}

// syncLinearSpecificData extracts Linear-specific data from raw response
func (m *Manager) syncLinearSpecificData(variables map[string]interface{}, rawData map[string]interface{}) {
	// Extract Linear issue specific fields
	if identifier, ok := rawData["identifier"].(string); ok {
		variables["LINEAR_IDENTIFIER"] = identifier
		variables["LINEAR_EXTERNAL_ID"] = identifier
	}

	if url, ok := rawData["url"].(string); ok {
		variables["LINEAR_URL"] = url
	}

	if state, ok := rawData["state"].(map[string]interface{}); ok {
		if name, ok := state["name"].(string); ok {
			variables["LINEAR_STATE"] = name
			variables["TASK_STATUS"] = m.mapExternalStatusToZenStatus(name, "linear")
		}
	}

	if priority, ok := rawData["priority"].(float64); ok {
		variables["LINEAR_PRIORITY"] = int(priority)
		variables["PRIORITY"] = m.mapLinearPriorityToZenPriority(int(priority))
	}

	if assignee, ok := rawData["assignee"].(map[string]interface{}); ok {
		if displayName, ok := assignee["displayName"].(string); ok {
			variables["LINEAR_ASSIGNEE"] = displayName
			variables["OWNER_NAME"] = displayName
		}
		if email, ok := assignee["email"].(string); ok {
			variables["LINEAR_ASSIGNEE_EMAIL"] = email
			variables["OWNER_EMAIL"] = email
		}
	}

	if team, ok := rawData["team"].(map[string]interface{}); ok {
		if name, ok := team["name"].(string); ok {
			variables["LINEAR_TEAM"] = name
			variables["TEAM_NAME"] = name
		}
	}
}

// mapJiraIssueTypeToZenType maps Jira issue types to Zen task types
func (m *Manager) mapJiraIssueTypeToZenType(issueType string) string {
	switch strings.ToLower(issueType) {
	case "story", "user story":
		return "story"
	case "bug", "defect":
		return "bug"
	case "epic", "initiative":
		return "epic"
	case "spike", "research":
		return "spike"
	case "task", "sub-task", "subtask":
		return "task"
	default:
		return "task"
	}
}

// mapLinearPriorityToZenPriority maps Linear priority numbers to Zen priority
func (m *Manager) mapLinearPriorityToZenPriority(priority int) string {
	switch priority {
	case 4:
		return "P0" // Urgent
	case 3:
		return "P1" // High
	case 2:
		return "P2" // Medium
	case 1:
		return "P3" // Low
	case 0:
		return "P3" // No priority
	default:
		return "P2"
	}
}

// extractPlainTextFromJiraDescription extracts plain text from Jira's rich text description
func (m *Manager) extractPlainTextFromJiraDescription(description map[string]interface{}) string {
	content, ok := description["content"].([]interface{})
	if !ok {
		return ""
	}

	var plainText strings.Builder
	for _, block := range content {
		if blockMap, ok := block.(map[string]interface{}); ok {
			if blockType, ok := blockMap["type"].(string); ok {
				switch blockType {
				case "paragraph":
					if blockContent, ok := blockMap["content"].([]interface{}); ok {
						for _, textNode := range blockContent {
							if textMap, ok := textNode.(map[string]interface{}); ok {
								if text, ok := textMap["text"].(string); ok {
									plainText.WriteString(text)
								}
							}
						}
						plainText.WriteString("\n\n")
					}

				case "heading":
					if attrs, ok := blockMap["attrs"].(map[string]interface{}); ok {
						if level, ok := attrs["level"].(float64); ok {
							plainText.WriteString(strings.Repeat("#", int(level)) + " ")
						}
					}
					if blockContent, ok := blockMap["content"].([]interface{}); ok {
						for _, textNode := range blockContent {
							if textMap, ok := textNode.(map[string]interface{}); ok {
								if text, ok := textMap["text"].(string); ok {
									plainText.WriteString(text)
								}
							}
						}
						plainText.WriteString("\n\n")
					}

				case "bulletList", "orderedList":
					if blockContent, ok := blockMap["content"].([]interface{}); ok {
						for i, listItem := range blockContent {
							if itemMap, ok := listItem.(map[string]interface{}); ok {
								if itemContent, ok := itemMap["content"].([]interface{}); ok {
									if blockType == "orderedList" {
										plainText.WriteString(fmt.Sprintf("%d. ", i+1))
									} else {
										plainText.WriteString("- ")
									}
									for _, paragraph := range itemContent {
										if paraMap, ok := paragraph.(map[string]interface{}); ok {
											if paraContent, ok := paraMap["content"].([]interface{}); ok {
												for _, textNode := range paraContent {
													if textMap, ok := textNode.(map[string]interface{}); ok {
														if text, ok := textMap["text"].(string); ok {
															plainText.WriteString(text)
														}
													}
												}
											}
										}
									}
									plainText.WriteString("\n")
								}
							}
						}
						plainText.WriteString("\n")
					}

				case "codeBlock":
					if blockContent, ok := blockMap["content"].([]interface{}); ok {
						plainText.WriteString("```\n")
						for _, textNode := range blockContent {
							if textMap, ok := textNode.(map[string]interface{}); ok {
								if text, ok := textMap["text"].(string); ok {
									plainText.WriteString(text)
								}
							}
						}
						plainText.WriteString("\n```\n\n")
					}
				}
			}
		}
	}

	return strings.TrimSpace(plainText.String())
}

// saveSourceMetadata saves source metadata to the task directory
func (m *Manager) saveSourceMetadata(taskDir string, sourceData *TaskData, source string) error {
	// Create metadata directory
	metadataDir := filepath.Join(taskDir, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Create comprehensive metadata structure
	metadata := map[string]interface{}{
		"external_system": source,
		"external_id":     sourceData.ExternalID,
		"fetched_at":      time.Now().Format(time.RFC3339),
		"sync_enabled":    true,
		"last_sync":       time.Now().Format(time.RFC3339),
		"plugin_version":  "1.0.0",
		"sync_direction":  "bidirectional",

		// Synced task data
		"task_data": map[string]interface{}{
			"id":           sourceData.ID,
			"external_id":  sourceData.ExternalID,
			"title":        sourceData.Title,
			"description":  sourceData.Description,
			"type":         sourceData.Type,
			"status":       sourceData.Status,
			"priority":     sourceData.Priority,
			"assignee":     sourceData.Assignee,
			"owner":        sourceData.Owner,
			"team":         sourceData.Team,
			"created":      sourceData.Created.Format(time.RFC3339),
			"updated":      sourceData.Updated.Format(time.RFC3339),
			"external_url": sourceData.ExternalURL,
			"labels":       sourceData.Labels,
			"components":   sourceData.Components,
		},

		// Source metadata
		"source_metadata": sourceData.Metadata,

		// Raw response data for future processing
		"raw_data": sourceData.RawData,
	}

	// Marshal to JSON with pretty formatting
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal %s metadata: %w", source, err)
	}

	// Write to source-specific metadata file
	metadataFilePath := filepath.Join(metadataDir, fmt.Sprintf("%s.json", source))
	if err := os.WriteFile(metadataFilePath, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write %s metadata file: %w", source, err)
	}

	return nil
}

// generateFileFromTemplate generates a file from a template
func (m *Manager) generateFileFromTemplate(loader *templates.LocalTemplateLoader, templateName, taskDir, fileName string, variables map[string]interface{}) error {
	content, err := loader.RenderTemplate(templateName, variables)
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", templateName, err)
	}

	filePath := filepath.Join(taskDir, fileName)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", fileName, err)
	}

	return nil
}

// loadTaskFromManifest loads a task from its manifest file
func (m *Manager) loadTaskFromManifest(taskID string) (*Task, error) {
	// Get workspace path
	ws, err := m.factory.WorkspaceManager()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace manager: %w", err)
	}

	status, err := ws.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace status: %w", err)
	}

	taskDir := filepath.Join(status.Root, ".zen", "work", "tasks", taskID)
	manifestPath := filepath.Join(taskDir, "manifest.yaml")

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("task manifest not found: %s", taskID)
	}

	// Read and parse manifest
	_, err = os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	// Parse YAML manifest (simplified - would use proper YAML parsing)
	// For now, create a basic task structure
	task := &Task{
		ID:            taskID,
		WorkspacePath: taskDir,
		IndexPath:     filepath.Join(taskDir, "index.md"),
		ManifestPath:  manifestPath,
		MetadataPath:  filepath.Join(taskDir, "metadata"),
		Sources:       make(map[string]*TaskSource),
		Metadata:      make(map[string]interface{}),
	}

	// Load source metadata
	if err := m.loadTaskSources(task); err != nil {
		m.logger.Warn("failed to load task sources", "task_id", taskID, "error", err)
	}

	return task, nil
}

// loadTaskSources loads external source information for a task
func (m *Manager) loadTaskSources(task *Task) error {
	metadataDir := task.MetadataPath

	// Check if metadata directory exists
	if _, err := os.Stat(metadataDir); os.IsNotExist(err) {
		return nil // No metadata directory
	}

	// Scan for source metadata files
	entries, err := os.ReadDir(metadataDir)
	if err != nil {
		return fmt.Errorf("failed to read metadata directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		// Extract source name from filename
		source := strings.TrimSuffix(entry.Name(), ".json")

		// Load source metadata
		sourceMetadataPath := filepath.Join(metadataDir, entry.Name())
		sourceData, err := os.ReadFile(sourceMetadataPath)
		if err != nil {
			m.logger.Warn("failed to read source metadata", "source", source, "error", err)
			continue
		}

		var metadata map[string]interface{}
		if err := json.Unmarshal(sourceData, &metadata); err != nil {
			m.logger.Warn("failed to parse source metadata", "source", source, "error", err)
			continue
		}

		// Create task source
		taskSource := &TaskSource{
			System:        source,
			SyncEnabled:   true,
			SyncDirection: "bidirectional",
			Metadata:      metadata,
		}

		if externalID, ok := metadata["external_id"].(string); ok {
			taskSource.ExternalID = externalID
		}
		if lastSyncStr, ok := metadata["last_sync"].(string); ok {
			if lastSync, err := time.Parse(time.RFC3339, lastSyncStr); err == nil {
				taskSource.LastSync = lastSync
			}
		}

		task.Sources[source] = taskSource
	}

	return nil
}

// saveTask saves task data back to the manifest and metadata
func (m *Manager) saveTask(ctx context.Context, task *Task) error {
	// Update task metadata
	task.Updated = time.Now()

	// Save to manifest file (simplified - would use proper YAML generation)
	// For now, just update the metadata files

	// Save source metadata
	for source, sourceInfo := range task.Sources {
		if err := m.updateSourceMetadata(task.MetadataPath, source, sourceInfo); err != nil {
			m.logger.Warn("failed to update source metadata", "source", source, "error", err)
		}
	}

	return nil
}

// updateSourceMetadata updates source metadata file
func (m *Manager) updateSourceMetadata(metadataDir, source string, sourceInfo *TaskSource) error {
	sourceMetadataPath := filepath.Join(metadataDir, fmt.Sprintf("%s.json", source))

	// Read existing metadata
	var metadata map[string]interface{}
	if data, err := os.ReadFile(sourceMetadataPath); err == nil {
		json.Unmarshal(data, &metadata)
	} else {
		metadata = make(map[string]interface{})
	}

	// Update sync information
	metadata["last_sync"] = sourceInfo.LastSync.Format(time.RFC3339)
	metadata["sync_enabled"] = sourceInfo.SyncEnabled
	metadata["sync_direction"] = sourceInfo.SyncDirection

	// Write back to file
	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	return os.WriteFile(sourceMetadataPath, jsonData, 0644)
}
