package jira

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// JiraTime is a custom time type that handles Jira's various time formats
type JiraTime time.Time

// UnmarshalJSON handles Jira's time format variations
func (jt *JiraTime) UnmarshalJSON(data []byte) error {
	// Remove quotes
	s := strings.Trim(string(data), "\"")

	// Try different formats
	formats := []string{
		"2006-01-02T15:04:05.999-0700",  // RFC3339 with milliseconds
		"2006-01-02T15:04:05.999Z07:00", // RFC3339 with milliseconds and colon in timezone
		"2006-01-02T15:04:05.999Z0700",  // RFC3339 with milliseconds without colon
		"2006-01-02T15:04:05-0700",      // RFC3339 without milliseconds
		"2006-01-02T15:04:05Z07:00",     // RFC3339 with colon in timezone
		"2006-01-02T15:04:05Z0700",      // RFC3339 without colon
		time.RFC3339,                    // Standard RFC3339
		time.RFC3339Nano,                // RFC3339 with nanoseconds
	}

	var t time.Time
	var err error

	// Special handling for timezone formats like +1000
	if strings.Contains(s, "+") || strings.Contains(s, "-") {
		// Try to fix the timezone format
		// Convert +1000 to +10:00
		parts := strings.Split(s, "+")
		if len(parts) == 2 && len(parts[1]) == 4 {
			s = parts[0] + "+" + parts[1][:2] + ":" + parts[1][2:]
		} else {
			parts = strings.Split(s, "-")
			if len(parts) == 3 && len(parts[2]) == 4 { // date-time-timezone
				s = parts[0] + "-" + parts[1] + "-" + parts[2][:2] + ":" + parts[2][2:]
			}
		}
	}

	for _, format := range formats {
		t, err = time.Parse(format, s)
		if err == nil {
			*jt = JiraTime(t)
			return nil
		}
	}

	return err
}

// Time returns the underlying time.Time
func (jt JiraTime) Time() time.Time {
	return time.Time(jt)
}

// MarshalJSON converts back to JSON
func (jt JiraTime) MarshalJSON() ([]byte, error) {
	t := time.Time(jt)
	return t.MarshalJSON()
}

// FlexibleID handles IDs that can be either string or number in Jira responses
type FlexibleID string

// UnmarshalJSON handles both string and number IDs from Jira
func (f *FlexibleID) UnmarshalJSON(data []byte) error {
	// Try as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleID(s)
		return nil
	}

	// Try as number
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleID(fmt.Sprintf("%.0f", n))
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s as string or number", string(data))
}

// String returns the string representation
func (f FlexibleID) String() string {
	return string(f)
}

// JiraDescription handles both string and object description formats
type JiraDescription struct {
	Text   string
	Object map[string]interface{}
}

// UnmarshalJSON handles both string and ADF (Atlassian Document Format) descriptions
func (d *JiraDescription) UnmarshalJSON(data []byte) error {
	// Try as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		d.Text = s
		return nil
	}

	// Try as object (ADF format)
	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		d.Object = obj
		// Try to extract plain text from ADF
		d.Text = d.extractTextFromADF(obj)
		return nil
	}

	// If it's null, that's okay
	if string(data) == "null" {
		d.Text = ""
		return nil
	}

	return fmt.Errorf("cannot unmarshal description: %s", string(data))
}

// extractTextFromADF extracts plain text from Atlassian Document Format
func (d *JiraDescription) extractTextFromADF(adf map[string]interface{}) string {
	var text strings.Builder

	// Navigate the ADF structure to extract text
	if content, ok := adf["content"].([]interface{}); ok {
		for _, node := range content {
			d.extractNodeText(node, &text)
		}
	}

	return strings.TrimSpace(text.String())
}

// extractNodeText recursively extracts text from ADF nodes
func (d *JiraDescription) extractNodeText(node interface{}, text *strings.Builder) {
	if nodeMap, ok := node.(map[string]interface{}); ok {
		// Check for text node
		if nodeType, ok := nodeMap["type"].(string); ok && nodeType == "text" {
			if nodeText, ok := nodeMap["text"].(string); ok {
				text.WriteString(nodeText)
			}
		}

		// Check for paragraph or other container nodes
		if content, ok := nodeMap["content"].([]interface{}); ok {
			for _, child := range content {
				d.extractNodeText(child, text)
			}
			// Add newline after paragraph
			if nodeType, ok := nodeMap["type"].(string); ok && nodeType == "paragraph" {
				text.WriteString("\n")
			}
		}
	}
}

// MarshalJSON converts the description to JSON
func (d JiraDescription) MarshalJSON() ([]byte, error) {
	if d.Object != nil {
		return json.Marshal(d.Object)
	}
	return json.Marshal(d.Text)
}

// String returns the text representation
func (d JiraDescription) String() string {
	return d.Text
}

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
	Description JiraDescription `json:"description"`
	Created     JiraTime        `json:"created"`
	Updated     JiraTime        `json:"updated"`
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
	ID   FlexibleID `json:"id"`
	Name string     `json:"name"`
	Key  string     `json:"key"`
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
