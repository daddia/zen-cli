package jira

import (
	"time"
)

// Jira API Types

// JiraIssue represents a Jira issue response
type JiraIssue struct {
	ID     string     `json:"id"`
	Key    string     `json:"key"`
	Self   string     `json:"self"`
	Fields JiraFields `json:"fields"`
}

// JiraFields represents Jira issue fields
type JiraFields struct {
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	Created     time.Time       `json:"created"`
	Updated     time.Time       `json:"updated"`
	Status      JiraStatus      `json:"status"`
	Priority    JiraPriority    `json:"priority"`
	Assignee    JiraUser        `json:"assignee"`
	Reporter    JiraUser        `json:"reporter"`
	Creator     JiraUser        `json:"creator"`
	IssueType   JiraIssueType   `json:"issuetype"`
	Project     JiraProject     `json:"project"`
	Labels      []string        `json:"labels"`
	Components  []JiraComponent `json:"components"`
	FixVersions []JiraVersion   `json:"fixVersions"`
}

// JiraStatus represents Jira status
type JiraStatus struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	StatusCategory JiraStatusCategory `json:"statusCategory"`
}

// JiraStatusCategory represents Jira status category
type JiraStatusCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

// JiraPriority represents Jira priority
type JiraPriority struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	IconURL string `json:"iconUrl"`
}

// JiraUser represents a Jira user
type JiraUser struct {
	AccountID    string `json:"accountId"`
	DisplayName  string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
	Active       bool   `json:"active"`
}

// JiraIssueType represents Jira issue type
type JiraIssueType struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
}

// JiraProject represents a Jira project
type JiraProject struct {
	ID             string `json:"id"`
	Key            string `json:"key"`
	Name           string `json:"name"`
	ProjectTypeKey string `json:"projectTypeKey"`
}

// JiraComponent represents a Jira component
type JiraComponent struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// JiraVersion represents a Jira version
type JiraVersion struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Released    bool   `json:"released"`
}

// JiraCreateRequest represents a Jira issue creation request
type JiraCreateRequest struct {
	Fields JiraCreateFields `json:"fields"`
}

// JiraCreateFields represents fields for creating a Jira issue
type JiraCreateFields struct {
	Project     JiraProject     `json:"project"`
	Summary     string          `json:"summary"`
	Description string          `json:"description"`
	IssueType   JiraIssueType   `json:"issuetype"`
	Priority    *JiraPriority   `json:"priority,omitempty"`
	Assignee    *JiraUser       `json:"assignee,omitempty"`
	Labels      []string        `json:"labels,omitempty"`
	Components  []JiraComponent `json:"components,omitempty"`
}

// JiraUpdateRequest represents a Jira issue update request
type JiraUpdateRequest struct {
	Fields map[string]interface{} `json:"fields"`
}

// JiraCreateResponse represents the response from creating a Jira issue
type JiraCreateResponse struct {
	ID   string `json:"id"`
	Key  string `json:"key"`
	Self string `json:"self"`
}

// JiraSearchResponse represents a Jira search response
type JiraSearchResponse struct {
	Expand     string      `json:"expand"`
	StartAt    int         `json:"startAt"`
	MaxResults int         `json:"maxResults"`
	Total      int         `json:"total"`
	Issues     []JiraIssue `json:"issues"`
}

// JiraSearchRequest represents a Jira search request
type JiraSearchRequest struct {
	JQL        string   `json:"jql"`
	StartAt    int      `json:"startAt"`
	MaxResults int      `json:"maxResults"`
	Fields     []string `json:"fields,omitempty"`
}

// JiraErrorResponse represents a Jira error response
type JiraErrorResponse struct {
	ErrorMessages   []string          `json:"errorMessages"`
	Errors          map[string]string `json:"errors"`
	WarningMessages []string          `json:"warningMessages,omitempty"`
}

// JiraTransition represents a Jira issue transition
type JiraTransition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	To   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"to"`
}

// JiraTransitionsResponse represents the response from getting available transitions
type JiraTransitionsResponse struct {
	Expand      string           `json:"expand"`
	Transitions []JiraTransition `json:"transitions"`
}

// JiraTransitionRequest represents a request to transition a Jira issue
type JiraTransitionRequest struct {
	Transition struct {
		ID string `json:"id"`
	} `json:"transition"`
	Fields map[string]interface{} `json:"fields,omitempty"`
}

// JiraServerInfo represents Jira server information
type JiraServerInfo struct {
	Version     string    `json:"version"`
	BuildNumber int       `json:"buildNumber"`
	BuildDate   time.Time `json:"buildDate"`
	ServerTime  time.Time `json:"serverTime"`
	ScmInfo     string    `json:"scmInfo"`
	ServerTitle string    `json:"serverTitle"`
}

// JiraWebhook represents a Jira webhook configuration
type JiraWebhook struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Events      []string `json:"events"`
	Filters     []string `json:"filters,omitempty"`
	ExcludeBody bool     `json:"excludeBody"`
}

// JiraWebhookEvent represents a Jira webhook event
type JiraWebhookEvent struct {
	Timestamp          int64     `json:"timestamp"`
	WebhookEvent       string    `json:"webhookEvent"`
	IssueEventTypeName string    `json:"issue_event_type_name,omitempty"`
	User               JiraUser  `json:"user"`
	Issue              JiraIssue `json:"issue"`
	Changelog          struct {
		ID    string `json:"id"`
		Items []struct {
			Field      string `json:"field"`
			FieldType  string `json:"fieldtype"`
			From       string `json:"from"`
			FromString string `json:"fromString"`
			To         string `json:"to"`
			ToString   string `json:"toString"`
		} `json:"items"`
	} `json:"changelog,omitempty"`
}
