package create

import (
	"context"
	"fmt"
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
	FromJira bool // Flag to fetch task details from Jira

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

			# Create task from existing Jira issue (type and details fetched from Jira)
			zen task create ZEN-123 --jira
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.TaskID = args[0]

			// Validate task ID format
			if err := validateTaskID(opts.TaskID); err != nil {
				return err
			}

			// Apply type rules:
			// 1. Type is always optional
			// 2. Default to "story" if not set
			// 3. External system (Jira) will override type
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
	cmd.Flags().BoolVar(&opts.FromJira, "jira", false, "Fetch task details from Jira using the task ID")

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

	// If --jira flag is used, fetch task details from Jira first
	if opts.FromJira {
		if err := fetchTaskFromJira(ctx, opts); err != nil {
			return fmt.Errorf("failed to fetch from Jira: %w", err)
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

	// Generate task files from templates
	if err := generateTaskFiles(ctx, opts, taskDir); err != nil {
		return fmt.Errorf("failed to generate task files: %w", err)
	}

	// Create Jira metadata if task was fetched from Jira
	if opts.FromJira {
		if err := opts.TaskOps.CreateJiraMetadataFromTask(taskDir, opts.TaskID); err != nil {
			// Log warning but don't fail task creation
			fmt.Fprintf(opts.IO.Out, "%s Failed to create Jira metadata: %v\n",
				opts.IO.FormatWarning("!"), err)
		}
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

	// Prepare template variables
	variables := buildTemplateVariables(opts)

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
func buildTemplateVariables(opts *CreateOptions) map[string]interface{} {
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

	return map[string]interface{}{
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

// fetchTaskFromJira fetches task details from Jira and populates the CreateOptions
func fetchTaskFromJira(ctx context.Context, opts *CreateOptions) error {
	// Use task operations to fetch from Jira
	jiraData, err := opts.TaskOps.FetchFromJira(ctx, opts.TaskID)
	if err != nil {
		return err
	}

	// Update options with Jira data (External SoR overrides)
	opts.Title = jiraData.Title
	opts.TaskType = jiraData.Type // Jira overrides type
	if jiraData.Assignee != "" {
		opts.Owner = jiraData.Assignee
	}

	return nil
}
