package create

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/task"
	"github.com/daddia/zen/pkg/templates"
	"github.com/daddia/zen/pkg/types"
	"github.com/spf13/cobra"
)

// CreateOptions contains options for the task create command
type CreateOptions struct {
	IO               *iostreams.IOStreams
	WorkspaceManager func() (cmdutil.WorkspaceManager, error)
	TemplateEngine   func() (cmdutil.TemplateEngineInterface, error)
	Factory          *cmdutil.Factory

	TaskID   string
	TaskType string
	Title    string
	Owner    string
	Team     string
	Priority string
	DryRun   bool
	Source   string // Source system to fetch task details from (jira, github, linear, etc.)

	// Task operations for external integrations
	TaskOps *task.Operations
}

// TaskType represents valid task types
type TaskType string

const (
	TaskTypeStory TaskType = "story"
	TaskTypeBug   TaskType = "bug"
	TaskTypeEpic  TaskType = "epic"
	TaskTypeSpike TaskType = "spike"
	TaskTypeTask  TaskType = "task"
)

// ValidTaskTypes returns all valid task types
func ValidTaskTypes() []string {
	return []string{
		string(TaskTypeStory),
		string(TaskTypeBug),
		string(TaskTypeEpic),
		string(TaskTypeSpike),
		string(TaskTypeTask),
	}
}

// IsValidTaskType checks if a task type is valid
func IsValidTaskType(taskType string) bool {
	for _, valid := range ValidTaskTypes() {
		if taskType == valid {
			return true
		}
	}
	return false
}

// NewCmdTaskCreate creates the task create command
func NewCmdTaskCreate(f *cmdutil.Factory) *cobra.Command {
	opts := &CreateOptions{
		IO:               f.IOStreams,
		WorkspaceManager: f.WorkspaceManager,
		TemplateEngine:   f.TemplateEngine,
		Factory:          f,
		DryRun:           f.DryRun,
		TaskOps:          task.NewOperations(f),
	}

	cmd := &cobra.Command{
		Use:   "create <task-id>",
		Short: "Create a new task with structured workflow",
		Long: `Create a new task with structured workflow directories and templates.

This command creates a complete task structure in .zen/work/tasks/<task-id>/ including:
- index.md: Human-readable task overview and status
- manifest.yaml: Machine-readable metadata for automation
- .taskrc.yaml: Task-specific configuration
- Work-type directories: research/, spikes/, design/, execution/, outcomes/

The task follows the seven-stage Zenflow workflow:
1. Align: Define success criteria and stakeholder alignment
2. Discover: Gather evidence and validate assumptions
3. Prioritize: Rank work by value vs effort
4. Design: Specify implementation approach
5. Build: Deliver working software increment
6. Ship: Deploy safely to production
7. Learn: Measure outcomes and iterate

Task types determine the workflow focus:
- story: User-facing feature development with UX focus
- bug: Defect fixes with root cause analysis
- epic: Large initiatives requiring breakdown
- spike: Research and exploration with learning focus
- task: General work items with flexible structure`,
		Example: heredoc.Doc(`
			# Create a user story (type defaults to story)
			zen task create USER-123 --title "User login with SSO"

			# Create a bug fix task
			zen task create BUG-456 --type bug --title "Fix memory leak in auth service"

			# Create an epic for large initiative
			zen task create EPIC-789 --type epic --title "Implement new payment system"

			# Create a research spike
			zen task create SPIKE-101 --type spike --title "Evaluate GraphQL vs REST"

			# Create with additional metadata
			zen task create PROJ-200 --title "Dashboard redesign" --owner "jane.doe" --team "frontend"

			# Create task from existing external source (type and details fetched from source)
			zen task create ZEN-123 --from jira
			zen task create GH-456 --from github
			zen task create LIN-789 --from linear
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.TaskID = args[0]

			// Validate task ID format
			if err := validateTaskID(opts.TaskID); err != nil {
				return err
			}

			// Task type rules:
			// 1. Type is optional
			// 2. Default to "story" if not set
			// 3. External source system (Jira) overrides type
			if opts.TaskType == "" {
				opts.TaskType = "story" // Default to story
			}

			// Validate task type if provided
			if !IsValidTaskType(opts.TaskType) {
				return &types.Error{
					Code:    types.ErrorCodeInvalidInput,
					Message: fmt.Sprintf("invalid task type '%s'", opts.TaskType),
					Details: fmt.Sprintf("valid types are: %s", strings.Join(ValidTaskTypes(), ", ")),
				}
			}

			return createRun(opts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.TaskType, "type", "t", "", fmt.Sprintf("Task type (%s, defaults to story)", strings.Join(ValidTaskTypes(), "|")))
	cmd.Flags().StringVar(&opts.Title, "title", "", "Task title (optional, will prompt if not provided)")
	cmd.Flags().StringVar(&opts.Owner, "owner", "", "Task owner (optional, defaults to current user)")
	cmd.Flags().StringVar(&opts.Team, "team", "", "Team name (optional)")
	cmd.Flags().StringVar(&opts.Priority, "priority", "P2", "Task priority (P0|P1|P2|P3)")
	cmd.Flags().StringVar(&opts.Source, "from", "", "Fetch task details from source system (jira, github, linear, etc.)")

	// No required flags - type is optional with default

	return cmd
}

// validateTaskID validates the task ID format
func validateTaskID(taskID string) error {
	if taskID == "" {
		return &types.Error{
			Code:    types.ErrorCodeInvalidInput,
			Message: "task ID cannot be empty",
		}
	}

	// Check for basic format requirements
	if len(taskID) < 3 {
		return &types.Error{
			Code:    types.ErrorCodeInvalidInput,
			Message: "task ID must be at least 3 characters long",
		}
	}

	// Check for invalid characters
	if strings.ContainsAny(taskID, " /\\:*?\"<>|") {
		return &types.Error{
			Code:    types.ErrorCodeInvalidInput,
			Message: "task ID contains invalid characters",
			Details: "task ID cannot contain spaces or filesystem-reserved characters",
		}
	}

	return nil
}

// createRun executes the task creation
func createRun(opts *CreateOptions) error {
	ctx := context.Background()

	// If --from flag is used, fetch task details from external source first
	if opts.Source != "" {
		if err := fetchTaskFromSource(ctx, opts); err != nil {
			return fmt.Errorf("failed to fetch from %s: %w", opts.Source, err)
		}
	}

	// Check workspace initialization
	wm, err := opts.WorkspaceManager()
	if err != nil {
		return fmt.Errorf("failed to get workspace manager: %w", err)
	}

	status, err := wm.Status()
	if err != nil {
		return fmt.Errorf("failed to get workspace status: %w", err)
	}

	if !status.Initialized {
		return &types.Error{
			Code:    types.ErrorCodeWorkspaceNotInit,
			Message: "workspace not initialized",
			Details: "run 'zen init' to initialize a workspace first",
		}
	}

	// Create task directory structure
	taskDir := filepath.Join(status.Root, ".zen", "work", "tasks", opts.TaskID)

	// Check if task already exists
	if _, err := os.Stat(taskDir); err == nil {
		return &types.Error{
			Code:    types.ErrorCodeAlreadyExists,
			Message: fmt.Sprintf("task '%s' already exists", opts.TaskID),
			Details: fmt.Sprintf("task directory exists at: %s", taskDir),
		}
	}

	if opts.DryRun {
		fmt.Fprintf(opts.IO.Out, "%s Task creation plan for %s:\n",
			opts.IO.FormatSuccess(""),
			opts.IO.ColorBold(opts.TaskID))
		fmt.Fprintf(opts.IO.Out, "  %s Type: %s\n",
			opts.IO.ColorNeutral("→"), opts.TaskType)
		fmt.Fprintf(opts.IO.Out, "  %s Directory: %s\n",
			opts.IO.ColorNeutral("→"), taskDir)
		fmt.Fprintf(opts.IO.Out, "  %s Templates: index.md, manifest.yaml, .taskrc.yaml\n",
			opts.IO.ColorNeutral("→"))
		fmt.Fprintf(opts.IO.Out, "  %s Work directories: research/, spikes/, design/, execution/, outcomes/\n",
			opts.IO.ColorNeutral("→"))

		fmt.Fprintf(opts.IO.Out, "\n%s Run without --dry-run to create the task\n",
			opts.IO.ColorInfo("ℹ"))
		return nil
	}

	// Create task directory structure using workspace manager
	ws, err := opts.Factory.WorkspaceManager()
	if err != nil {
		return fmt.Errorf("failed to get workspace manager: %w", err)
	}

	if err := ws.CreateTaskDirectory(taskDir); err != nil {
		return fmt.Errorf("failed to create task directories: %w", err)
	}

	// Create external system metadata first if task was fetched from external source
	if opts.Source != "" {
		// Save raw response from external provider using config credentials
		if err := saveRawResponse(opts, taskDir, opts.TaskID, opts.Source); err != nil {
			// Log warning but don't fail task creation
			fmt.Fprintf(opts.IO.Out, "%s Failed to create %s metadata: %v\n",
				opts.IO.FormatWarning("!"), opts.Source, err)
		} else {
			fmt.Fprintf(opts.IO.Out, "%s Created %s metadata: %s\n",
				opts.IO.FormatSuccess("✓"), opts.Source, fmt.Sprintf("metadata/%s.json", opts.Source))
		}
	}

	// Generate task files from templates (now with external source data available)
	if err := generateTaskFiles(ctx, opts, taskDir); err != nil {
		return fmt.Errorf("failed to generate task files: %w", err)
	}

	// Try to sync with external system if configured
	if err := tryIntegrationSync(ctx, opts); err != nil {
		// Log the error but don't fail the task creation
		fmt.Fprintf(opts.IO.Out, "%s Integration sync failed: %v\n",
			opts.IO.FormatWarning("!"), err)
	}

	// Success output
	fmt.Fprintf(opts.IO.Out, "%s Created task %s\n",
		opts.IO.FormatSuccess(""),
		opts.IO.ColorBold(opts.TaskID))
	fmt.Fprintf(opts.IO.Out, "  %s Type: %s\n",
		opts.IO.ColorNeutral("→"), opts.TaskType)
	fmt.Fprintf(opts.IO.Out, "  %s Location: %s\n",
		opts.IO.ColorNeutral("→"), taskDir)

	// Show next steps
	fmt.Fprintf(opts.IO.Out, "\n%s\n",
		opts.IO.ColorBold("Next steps:"))
	fmt.Fprintf(opts.IO.Out, "  1. Edit %s to define the task\n",
		opts.IO.ColorNeutral("index.md"))
	fmt.Fprintf(opts.IO.Out, "  2. Configure automation in %s\n",
		opts.IO.ColorNeutral(".taskrc.yaml"))
	fmt.Fprintf(opts.IO.Out, "  3. Start workflow: %s\n",
		opts.IO.ColorNeutral(fmt.Sprintf("zen %s align", opts.TaskID)))

	return nil
}

// generateTaskFiles generates the task files from templates
func generateTaskFiles(ctx context.Context, opts *CreateOptions, taskDir string) error {
	// Create local template loader
	templateLoader := templates.NewLocalTemplateLoader()

	// Prepare template variables (enriched with Jira data if available)
	variables := buildTemplateVariables(opts, taskDir)

	// Generate index.md from template
	if err := generateFileFromTemplate(templateLoader, "index.md", taskDir, "index.md", variables); err != nil {
		return fmt.Errorf("failed to generate index.md: %w", err)
	}

	// Generate manifest.yaml from template
	if err := generateFileFromTemplate(templateLoader, "manifest.yaml", taskDir, "manifest.yaml", variables); err != nil {
		return fmt.Errorf("failed to generate manifest.yaml: %w", err)
	}

	// Generate .taskrc.yaml from template
	if err := generateFileFromTemplate(templateLoader, "taskrc.yaml", taskDir, ".taskrc.yaml", variables); err != nil {
		return fmt.Errorf("failed to generate .taskrc.yaml: %w", err)
	}

	return nil
}

// generateFileFromTemplate generates a file from a template
func generateFileFromTemplate(loader *templates.LocalTemplateLoader, templateName, taskDir, fileName string, variables map[string]interface{}) error {
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

// buildTemplateVariables builds the template variables map
func buildTemplateVariables(opts *CreateOptions, taskDir string) map[string]interface{} {
	now := time.Now()

	// Default values
	title := opts.Title
	if title == "" {
		title = fmt.Sprintf("New %s task", opts.TaskType)
	}

	owner := opts.Owner
	if owner == "" {
		if user := os.Getenv("USER"); user != "" {
			owner = user
		} else {
			owner = "unknown"
		}
	}

	team := opts.Team
	if team == "" {
		team = "default"
	}

	priority := opts.Priority
	if priority == "" {
		priority = "P2"
	}

	// Build workflow stages
	workflowStages := []map[string]interface{}{
		{
			"id":             "01-align",
			"name":           "Align",
			"status":         "not_started",
			"progress":       0,
			"started_date":   nil,
			"completed_date": nil,
			"artifacts":      []interface{}{},
		},
		{
			"id":             "02-discover",
			"name":           "Discover",
			"status":         "not_started",
			"progress":       0,
			"started_date":   nil,
			"completed_date": nil,
			"artifacts":      []interface{}{},
		},
		{
			"id":             "03-prioritize",
			"name":           "Prioritize",
			"status":         "not_started",
			"progress":       0,
			"started_date":   nil,
			"completed_date": nil,
			"artifacts":      []interface{}{},
		},
		{
			"id":             "04-design",
			"name":           "Design",
			"status":         "not_started",
			"progress":       0,
			"started_date":   nil,
			"completed_date": nil,
			"artifacts":      []interface{}{},
		},
		{
			"id":             "05-build",
			"name":           "Build",
			"status":         "not_started",
			"progress":       0,
			"started_date":   nil,
			"completed_date": nil,
			"artifacts":      []interface{}{},
		},
		{
			"id":             "06-ship",
			"name":           "Ship",
			"status":         "not_started",
			"progress":       0,
			"started_date":   nil,
			"completed_date": nil,
			"artifacts":      []interface{}{},
		},
		{
			"id":             "07-learn",
			"name":           "Learn",
			"status":         "not_started",
			"progress":       0,
			"started_date":   nil,
			"completed_date": nil,
			"artifacts":      []interface{}{},
		},
	}

	// Create base template variables
	variables := map[string]interface{}{
		// Basic task information
		"TASK_ID":      opts.TaskID,
		"TASK_TITLE":   title,
		"TASK_TYPE":    opts.TaskType,
		"TASK_STATUS":  "proposed",
		"PRIORITY":     priority,
		"SIZE":         "M",
		"STORY_POINTS": 3,

		// Ownership and team
		"OWNER_NAME":      owner,
		"OWNER_EMAIL":     fmt.Sprintf("%s@company.com", owner),
		"GITHUB_USERNAME": owner,
		"TEAM_NAME":       team,
		"STREAM_TYPE":     "i2d",
		"TEAM_MEMBERS":    []interface{}{},

		// Dates
		"CREATED_DATE":   now.Format("2006-01-02"),
		"STARTED_DATE":   nil,
		"TARGET_DATE":    now.AddDate(0, 0, 14).Format("2006-01-02"), // 2 weeks from now
		"COMPLETED_DATE": nil,
		"LAST_UPDATED":   now.Format("2006-01-02 15:04:05"),

		// Workflow
		"CURRENT_STAGE":         "01-align",
		"COMPLETED_STAGES_LIST": []string{},
		"WORKFLOW_STAGES":       workflowStages,

		// Quality gates
		"QUALITY_GATES": []interface{}{},

		// Success criteria
		"BUSINESS_CRITERIA":  []string{"Define business value"},
		"TECHNICAL_CRITERIA": []string{"Define technical approach"},
		"UX_CRITERIA":        []string{"Define user experience"},

		// Integrations (all disabled by default)
		"JIRA_INTEGRATION":       false,
		"GITHUB_INTEGRATION":     false,
		"FIGMA_INTEGRATION":      false,
		"CONFLUENCE_INTEGRATION": false,

		// Dependencies
		"UPSTREAM_DEPS":   []interface{}{},
		"DOWNSTREAM_DEPS": []interface{}{},

		// Risk
		"RISK_LEVEL":   "low",
		"RISK_FACTORS": []interface{}{},

		// Metrics
		"PLANNED_VELOCITY": 1.0,
		"ACTUAL_VELOCITY":  nil,
		"CYCLE_TIME_DAYS":  nil,
		"DEFECT_RATE":      nil,
		"CUSTOM_METRICS":   []interface{}{},

		// Automation
		"SYNC_ENABLED":           false,
		"QUALITY_GATES_ENFORCED": true,
		"NOTIFICATION_EVENTS":    []string{},
		"ENABLED_AGENTS":         []string{},
		"CONTEXT_SHARING":        false,

		// Labels and tags
		"LABELS": []string{opts.TaskType},
		"TAGS":   []string{opts.TaskType, team},

		// Custom fields
		"CUSTOM_FIELDS": []interface{}{},

		// Task configuration defaults
		"SYNC_FREQUENCY":         "daily",
		"SYNC_DIRECTION":         "bidirectional",
		"SYNC_SYSTEMS":           []interface{}{},
		"AUTO_PROGRESS":          false,
		"VALIDATE_ON_TRANSITION": true,
		"GENERATE_ARTIFACTS":     true,
		"REQUIRED_APPROVALS":     []interface{}{},
		"AGENTS_ENABLED":         false,
		"AI_AGENTS":              []interface{}{},
		"AUTO_ENHANCE":           false,

		// Quality standards
		"CODE_COVERAGE_THRESHOLD":     80,
		"LINTING_ENABLED":             true,
		"COMPLEXITY_THRESHOLD":        10,
		"SECURITY_SCAN_REQUIRED":      true,
		"VULNERABILITY_THRESHOLD":     "medium",
		"COMPLIANCE_STANDARDS":        []string{},
		"RESPONSE_TIME_P95":           500,
		"THROUGHPUT_MINIMUM":          100,
		"ERROR_RATE_MAXIMUM":          0.01,
		"WCAG_LEVEL":                  "AA",
		"ACCESSIBILITY_TESTING":       true,
		"MANUAL_ACCESSIBILITY_REVIEW": false,

		// Notifications
		"NOTIFICATION_CHANNELS":           []interface{}{},
		"STAGE_COMPLETION_RECIPIENTS":     []string{owner},
		"QUALITY_GATE_FAILURE_RECIPIENTS": []string{owner},
		"BLOCKER_ADDED_RECIPIENTS":        []string{owner},
		"TASK_COMPLETED_RECIPIENTS":       []string{owner},

		// Templates
		"TEMPLATE_PREFERENCES":    []interface{}{},
		"TEMPLATE_VARIABLES":      []interface{}{},
		"AUTO_GENERATE_ARTIFACTS": []interface{}{},

		// Environments
		"ENVIRONMENTS": []interface{}{},

		// Feature flags
		"FEATURE_FLAGS": []interface{}{},

		// Hooks
		"PRE_STAGE_HOOKS":  []interface{}{},
		"POST_STAGE_HOOKS": []interface{}{},
		"CUSTOM_COMMANDS":  []interface{}{},

		// Reporting
		"STATUS_REPORTS_ENABLED":   false,
		"STATUS_REPORT_FREQUENCY":  "weekly",
		"STATUS_REPORT_RECIPIENTS": []string{owner},
		"TRACK_VELOCITY":           true,
		"TRACK_CYCLE_TIME":         true,
		"TRACK_DEFECTS":            true,

		// Resources
		"TEAM_ALLOCATION":  []interface{}{},
		"BUDGET_ALLOCATED": 0,
		"BUDGET_SPENT":     0,
		"BUDGET_CURRENCY":  "USD",

		// Custom config
		"CUSTOM_CONFIG": []interface{}{},

		// Additional template variables for index.md
		"current_stage_name":     "Align",
		"stage_number":           "1",
		"owner_name":             owner,
		"team_name":              team,
		"stream_type":            "i2d",
		"current_stage_progress": "0",
		"last_updated":           now.Format("January 2, 2006"),
		"next_review_date":       now.AddDate(0, 0, 7).Format("January 2, 2006"),
	}

	// Enrich with external source data if available using field mappings
	if opts.Source != "" {
		if sourceData, err := loadSourceData(taskDir, opts.Source); err == nil {
			enrichTemplateVariables(variables, sourceData, opts.Source)
		}
	}

	return variables
}

// tryIntegrationSync attempts to sync the newly created task with external systems
func tryIntegrationSync(ctx context.Context, opts *CreateOptions) error {
	// Check if factory has integration manager (may not be available in tests)
	if opts.Factory == nil || opts.Factory.IntegrationManager == nil {
		// Integration not available, skip silently
		return nil
	}

	// Get integration manager from factory
	integrationManager, err := opts.Factory.IntegrationManager()
	if err != nil {
		// Integration manager not available, skip silently
		return nil
	}

	// Check if integration is configured and enabled
	if !integrationManager.IsConfigured() || !integrationManager.IsSyncEnabled() {
		// Integration not configured, skip silently
		return nil
	}

	taskSystem := integrationManager.GetTaskSystem()
	fmt.Fprintf(opts.IO.Out, "%s Syncing with %s...\n",
		opts.IO.ColorInfo("ℹ"), taskSystem)

	// For now, just indicate that integration is configured and ready
	fmt.Fprintf(opts.IO.Out, "%s Integration with %s is configured and ready\n",
		opts.IO.FormatSuccess("✓"), taskSystem)
	fmt.Fprintf(opts.IO.Out, "  %s Task can be synced when sync commands are implemented\n",
		opts.IO.ColorNeutral("→"))
	fmt.Fprintf(opts.IO.Out, "  %s External ID will be: %s-XX (created in %s)\n",
		opts.IO.ColorNeutral("→"), opts.TaskID, taskSystem)

	return nil
}

// fetchTaskFromSource fetches task details from any external source and populates the CreateOptions
func fetchTaskFromSource(ctx context.Context, opts *CreateOptions) error {
	// Use task operations to fetch from external source
	taskData, err := opts.TaskOps.FetchFromSource(ctx, opts.TaskID, opts.Source)
	if err != nil {
		return err
	}

	// Update options with external source data (External SoR overrides)
	opts.Title = taskData.Title
	opts.TaskType = taskData.Type // External source overrides type
	if taskData.Assignee != "" {
		opts.Owner = taskData.Assignee
	}
	if taskData.Team != "" {
		opts.Team = taskData.Team
	}
	if taskData.Priority != "" {
		opts.Priority = taskData.Priority
	}

	return nil
}

// loadSourceData loads and parses source data from a task's metadata file
func loadSourceData(taskDir string, source string) (map[string]interface{}, error) {
	sourceFilePath := filepath.Join(taskDir, "metadata", fmt.Sprintf("%s.json", source))

	// Check if file exists
	if _, err := os.Stat(sourceFilePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s metadata file not found", source)
	}

	// Read and parse the file
	data, err := os.ReadFile(sourceFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s metadata: %w", source, err)
	}

	var sourceData map[string]interface{}
	if err := json.Unmarshal(data, &sourceData); err != nil {
		return nil, fmt.Errorf("failed to parse %s metadata: %w", source, err)
	}

	return sourceData, nil
}

// enrichTemplateVariables enriches template variables with external source data
func enrichTemplateVariables(variables map[string]interface{}, sourceData map[string]interface{}, source string) {
	// Set integration flags
	variables[fmt.Sprintf("%s_INTEGRATION", strings.ToUpper(source))] = true
	variables["EXTERNAL_SYSTEM"] = source
	variables["SYNC_ENABLED"] = true

	// Extract task data if available
	if taskDataInterface, ok := sourceData["task_data"]; ok {
		if taskData, ok := taskDataInterface.(map[string]interface{}); ok {
			// Map common fields
			if title, ok := taskData["title"].(string); ok {
				variables["TASK_TITLE"] = title
			}
			if taskType, ok := taskData["type"].(string); ok {
				variables["TASK_TYPE"] = taskType
			}
			if status, ok := taskData["status"].(string); ok {
				variables["TASK_STATUS"] = status
			}
			if priority, ok := taskData["priority"].(string); ok {
				variables["PRIORITY"] = priority
			}
			if assignee, ok := taskData["assignee"].(string); ok {
				variables["OWNER_NAME"] = assignee
			}
			if team, ok := taskData["team"].(string); ok {
				variables["TEAM_NAME"] = team
			}
			if externalURL, ok := taskData["external_url"].(string); ok {
				variables[fmt.Sprintf("%s_URL", strings.ToUpper(source))] = externalURL
			}
			if labels, ok := taskData["labels"].([]interface{}); ok {
				labelStrings := make([]string, 0, len(labels))
				for _, label := range labels {
					if labelStr, ok := label.(string); ok {
						labelStrings = append(labelStrings, labelStr)
					}
				}
				variables["LABELS"] = labelStrings
			}
		}
	}

	// Add source-specific prefixed variables
	if rawData, ok := sourceData["raw_data"]; ok {
		variables[fmt.Sprintf("%s_RAW_DATA", strings.ToUpper(source))] = rawData
	}
	if externalID, ok := sourceData["external_id"].(string); ok {
		variables[fmt.Sprintf("%s_EXTERNAL_ID", strings.ToUpper(source))] = externalID
	}
}

// saveRawResponse makes a direct API call to any provider and saves the raw response
func saveRawResponse(opts *CreateOptions, taskDir, taskID, provider string) error {
	// Get configuration to access provider settings
	cfg, err := opts.Factory.Config()
	if err != nil {
		return fmt.Errorf("failed to get config: %w", err)
	}

	// Get provider configuration
	providerConfig, exists := cfg.Integrations.Providers[provider]
	if !exists {
		return fmt.Errorf("provider %s not configured", provider)
	}

	// Build the API URL based on provider
	var url string
	switch provider {
	case "jira":
		url = fmt.Sprintf("%s/rest/api/3/issue/%s", providerConfig.URL, taskID)
	case "github":
		url = fmt.Sprintf("%s/repos/%s/issues/%s", providerConfig.URL, providerConfig.ProjectKey, taskID)
	case "gitlab":
		url = fmt.Sprintf("%s/projects/%s/issues/%s", providerConfig.URL, providerConfig.ProjectKey, taskID)
	default:
		return fmt.Errorf("unsupported provider: %s", provider)
	}

	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication based on provider configuration
	switch provider {
	case "jira":
		if providerConfig.Email != "" && providerConfig.APIKey != "" {
			// Use Basic Auth for Jira
			auth := providerConfig.Email + ":" + providerConfig.APIKey
			encoded := base64.StdEncoding.EncodeToString([]byte(auth))
			req.Header.Set("Authorization", "Basic "+encoded)
		}
	case "github":
		if providerConfig.APIKey != "" {
			req.Header.Set("Authorization", "Bearer "+providerConfig.APIKey)
		}
	case "gitlab":
		if providerConfig.APIKey != "" {
			req.Header.Set("Private-Token", providerConfig.APIKey)
		}
	}

	req.Header.Set("Accept", "application/json")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Create metadata directory
	metadataDir := filepath.Join(taskDir, "metadata")
	if err := os.MkdirAll(metadataDir, 0755); err != nil {
		return fmt.Errorf("failed to create metadata directory: %w", err)
	}

	// Save raw response
	metadataFilePath := filepath.Join(metadataDir, fmt.Sprintf("%s.json", provider))
	if err := os.WriteFile(metadataFilePath, body, 0644); err != nil {
		return fmt.Errorf("failed to write %s metadata file: %w", provider, err)
	}

	return nil
}
