package create

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/MakeNowJust/heredoc"
	"github.com/daddia/zen/internal/logging"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/filesystem"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/types"
	"github.com/spf13/cobra"
)

// CreateOptions contains options for the task create command
type CreateOptions struct {
	IO               *iostreams.IOStreams
	WorkspaceManager func() (cmdutil.WorkspaceManager, error)
	TemplateEngine   func() (cmdutil.TemplateEngineInterface, error)

	TaskID   string
	TaskType string
	Title    string
	Owner    string
	Team     string
	Priority string
	DryRun   bool
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
		DryRun:           f.DryRun,
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
			# Create a user story for feature development
			zen task create USER-123 --type story --title "User login with SSO"

			# Create a bug fix task
			zen task create BUG-456 --type bug --title "Fix memory leak in auth service"

			# Create an epic for large initiative
			zen task create EPIC-789 --type epic --title "Implement new payment system"

			# Create a research spike
			zen task create SPIKE-101 --type spike --title "Evaluate GraphQL vs REST"

			# Create with additional metadata
			zen task create PROJ-200 --type story --title "Dashboard redesign" --owner "jane.doe" --team "frontend"
		`),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.TaskID = args[0]

			// Validate task ID format
			if err := validateTaskID(opts.TaskID); err != nil {
				return err
			}

			// Validate required flags
			if opts.TaskType == "" {
				return &types.Error{
					Code:    types.ErrorCodeInvalidInput,
					Message: "task type is required",
					Details: fmt.Sprintf("use --type with one of: %s", strings.Join(ValidTaskTypes(), ", ")),
				}
			}

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
	cmd.Flags().StringVarP(&opts.TaskType, "type", "t", "", fmt.Sprintf("Task type (%s)", strings.Join(ValidTaskTypes(), "|")))
	cmd.Flags().StringVar(&opts.Title, "title", "", "Task title (optional, will prompt if not provided)")
	cmd.Flags().StringVar(&opts.Owner, "owner", "", "Task owner (optional, defaults to current user)")
	cmd.Flags().StringVar(&opts.Team, "team", "", "Team name (optional)")
	cmd.Flags().StringVar(&opts.Priority, "priority", "P2", "Task priority (P0|P1|P2|P3)")

	// Mark required flags
	_ = cmd.MarkFlagRequired("type")

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

	// Create task directory structure using shared filesystem utilities
	logger := logging.NewBasic() // Use a basic logger for filesystem operations
	fsManager := filesystem.New(logger)
	if err := fsManager.CreateTaskDirectory(taskDir); err != nil {
		return fmt.Errorf("failed to create task directories: %w", err)
	}

	// Generate task files from templates
	if err := generateTaskFiles(ctx, opts, taskDir); err != nil {
		return fmt.Errorf("failed to generate task files: %w", err)
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
	templateEngine, err := opts.TemplateEngine()
	if err != nil {
		return fmt.Errorf("failed to get template engine: %w", err)
	}

	// Prepare template variables
	variables := buildTemplateVariables(opts)

	// Generate index.md from template
	if err := generateIndexFile(ctx, templateEngine, taskDir, variables); err != nil {
		return fmt.Errorf("failed to generate index.md: %w", err)
	}

	// Generate manifest.yaml from template
	if err := generateManifestFile(ctx, templateEngine, taskDir, variables); err != nil {
		return fmt.Errorf("failed to generate manifest.yaml: %w", err)
	}

	// Generate .taskrc.yaml from template
	if err := generateTaskrcFile(ctx, templateEngine, taskDir, variables); err != nil {
		return fmt.Errorf("failed to generate .taskrc.yaml: %w", err)
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

// generateIndexFile generates the index.md file from template
func generateIndexFile(ctx context.Context, engine cmdutil.TemplateEngineInterface, taskDir string, variables map[string]interface{}) error {
	// Try to load the template from assets
	tmpl, err := engine.LoadTemplate(ctx, "task-index.md")
	if err != nil {
		// If template not found, use a simple fallback
		return generateFallbackIndexFile(taskDir, variables)
	}

	// Render the template
	content, err := engine.RenderTemplate(ctx, tmpl, variables)
	if err != nil {
		return fmt.Errorf("failed to render index template: %w", err)
	}

	// Write to file
	indexPath := filepath.Join(taskDir, "index.md")
	return os.WriteFile(indexPath, []byte(content), 0644)
}

// generateFallbackIndexFile generates a simple index.md when template is not available
func generateFallbackIndexFile(taskDir string, variables map[string]interface{}) error {
	content := fmt.Sprintf("# %s: %s\n\n"+
		"**Status:** %s (1/7) · **Owner:** %s · **Team:** %s · **Type:** %s\n\n"+
		"## Overview\n\n"+
		"<!-- Provide a task overview -->\n\n"+
		"## Business Context\n\n"+
		"<!-- Describe the business context -->\n\n"+
		"## Current Progress\n\n"+
		"### Workflow Status\n"+
		"```\n"+
		"[→·······]\n"+
		"1:Align  2:Discover  3:Prioritize  4:Design  5:Build  6:Ship  7:Learn\n"+
		"```\n\n"+
		"### Current Stage: Align\n"+
		"**Progress:** 0%%\n\n"+
		"<!-- List current stage artifacts -->\n"+
		"- → Define success criteria and stakeholder alignment\n\n"+
		"### Completed Stages\n"+
		"<!-- List completed stages -->\n"+
		"- (None yet)\n\n"+
		"### Upcoming Stages\n"+
		"- · **Discover**: Gather evidence and validate assumptions (Est. 3-5 days)\n\n"+
		"## Success Criteria\n\n"+
		"### Business Goals\n"+
		"<!-- List business success criteria -->\n"+
		"- Define business value\n\n"+
		"### Technical Goals\n"+
		"<!-- List technical success criteria -->\n"+
		"- Define technical approach\n\n"+
		"### User Experience Goals\n"+
		"<!-- List UX success criteria -->\n"+
		"- Define user experience\n\n"+
		"## Key Artifacts\n\n"+
		"### Current Stage Artifacts\n"+
		"<!-- List current artifacts -->\n"+
		"- [manifest.yaml](manifest.yaml): Machine-readable task metadata\n"+
		"- [.taskrc.yaml](.taskrc.yaml): Task-specific configuration\n\n"+
		"### Key Deliverables\n"+
		"<!-- List key deliverables -->\n"+
		"- ⏳ **Task Definition**: Complete task overview and requirements\n\n"+
		"## Dependencies and Blockers\n\n"+
		"### Dependencies\n"+
		"<!-- List dependencies -->\n"+
		"- (None identified)\n\n"+
		"### Blockers\n"+
		"<!-- List current blockers -->\n"+
		"- (None identified)\n\n"+
		"## Risks\n\n"+
		"<!-- List risks -->\n"+
		"- **Low**: Standard task execution risk\n\n"+
		"## Team and Collaboration\n\n"+
		"### Core Team\n"+
		"<!-- List core team members -->\n"+
		"- **Owner**: %s\n\n"+
		"### Recent Activity\n"+
		"<!-- List recent activity -->\n"+
		"- **%s**: Task created\n\n"+
		"## Links and Resources\n\n"+
		"### External Systems\n"+
		"- **Jira:** (Not integrated)\n"+
		"- **Pull Requests:** (None yet)\n"+
		"- **Design:** (Not specified)\n"+
		"- **Documentation:** (Not specified)\n\n"+
		"### Quick Actions\n"+
		"- [View Manifest](manifest.yaml) - Machine-readable metadata\n"+
		"- [Task Config](.taskrc.yaml) - Task-specific configuration\n"+
		"- [Workflow Artifacts](./) - Stage-specific artifacts\n\n"+
		"## Next Steps\n\n"+
		"<!-- List next steps -->\n"+
		"1. Define task scope and requirements\n"+
		"2. Identify stakeholders and success criteria\n"+
		"3. Begin Discover stage when ready\n\n"+
		"---\n\n"+
		"*Last updated: %s | Next review: %s*\n\n"+
		"*This is the task index providing a human-readable overview. For machine-readable data, see [manifest.yaml](manifest.yaml).*\n",
		variables["TASK_ID"],
		variables["TASK_TITLE"],
		variables["CURRENT_STAGE"].(string),
		variables["OWNER_NAME"],
		variables["TEAM_NAME"],
		variables["TASK_TYPE"],
		variables["OWNER_NAME"],
		variables["last_updated"],
		variables["last_updated"],
		variables["next_review_date"],
	)

	indexPath := filepath.Join(taskDir, "index.md")
	return os.WriteFile(indexPath, []byte(content), 0644)
}

// generateManifestFile generates the manifest.yaml file from template
func generateManifestFile(ctx context.Context, engine cmdutil.TemplateEngineInterface, taskDir string, variables map[string]interface{}) error {
	// Try to load the template from assets
	tmpl, err := engine.LoadTemplate(ctx, "manifest.yaml")
	if err != nil {
		// If template not found, use a simple fallback
		return generateFallbackManifestFile(taskDir, variables)
	}

	// Render the template
	content, err := engine.RenderTemplate(ctx, tmpl, variables)
	if err != nil {
		return fmt.Errorf("failed to render manifest template: %w", err)
	}

	// Write to file
	manifestPath := filepath.Join(taskDir, "manifest.yaml")
	return os.WriteFile(manifestPath, []byte(content), 0644)
}

// generateFallbackManifestFile generates a simple manifest.yaml when template is not available
func generateFallbackManifestFile(taskDir string, variables map[string]interface{}) error {
	content := fmt.Sprintf(`# Task Manifest - Machine-readable metadata for workflow automation
schema_version: "1.0"

# Basic task information
task:
  id: "%s"
  title: "%s"
  type: "%s"
  status: "%s"
  priority: "%s"
  size: "%s"
  points: %v

# Ownership and team
owner:
  name: "%s"
  email: "%s"
  github: "%s"

team:
  name: "%s"
  stream: "%s"
  members: []

# Important dates
dates:
  created: "%s"
  started: null
  target: "%s"
  completed: null
  last_updated: "%s"

# Zen workflow state
workflow:
  current_stage: "%s"
  completed_stages: []

  # Detailed stage progress
  stages:
    01-align:
      name: "Align"
      status: "not_started"
      progress: 0
      started: null
      completed: null
      artifacts: []
    02-discover:
      name: "Discover"
      status: "not_started"
      progress: 0
      started: null
      completed: null
      artifacts: []
    03-prioritize:
      name: "Prioritize"
      status: "not_started"
      progress: 0
      started: null
      completed: null
      artifacts: []
    04-design:
      name: "Design"
      status: "not_started"
      progress: 0
      started: null
      completed: null
      artifacts: []
    05-build:
      name: "Build"
      status: "not_started"
      progress: 0
      started: null
      completed: null
      artifacts: []
    06-ship:
      name: "Ship"
      status: "not_started"
      progress: 0
      started: null
      completed: null
      artifacts: []
    07-learn:
      name: "Learn"
      status: "not_started"
      progress: 0
      started: null
      completed: null
      artifacts: []

# Quality gates
quality_gates: {}

# Success criteria
success_criteria:
  business:
    - "Define business value"
  technical:
    - "Define technical approach"
  user_experience:
    - "Define user experience"

# External integrations
integrations: {}

# Dependencies and relationships
dependencies:
  upstream: []
  downstream: []

# Risk assessment
risk:
  level: "low"
  factors: []

# Metrics and measurements
metrics:
  velocity:
    planned: 1.0
    actual: null
  cycle_time:
    days: null
  defect_rate: null
  custom: {}

# Automation configuration
automation:
  sync_enabled: false
  quality_gates_enforced: true
  notifications:
    channels: []
    events: []
  agents:
    enabled: []
    context_sharing: false

# Labels and tags for search/filter
labels:
  - "%s"

tags:
  - "%s"
  - "%s"

# Custom fields for organization-specific needs
custom_fields: {}
`,
		variables["TASK_ID"],
		variables["TASK_TITLE"],
		variables["TASK_TYPE"],
		variables["TASK_STATUS"],
		variables["PRIORITY"],
		variables["SIZE"],
		variables["STORY_POINTS"],
		variables["OWNER_NAME"],
		variables["OWNER_EMAIL"],
		variables["GITHUB_USERNAME"],
		variables["TEAM_NAME"],
		variables["STREAM_TYPE"],
		variables["CREATED_DATE"],
		variables["TARGET_DATE"],
		variables["LAST_UPDATED"],
		variables["CURRENT_STAGE"],
		variables["TASK_TYPE"],
		variables["TASK_TYPE"],
		variables["TEAM_NAME"],
	)

	manifestPath := filepath.Join(taskDir, "manifest.yaml")
	return os.WriteFile(manifestPath, []byte(content), 0644)
}

// generateTaskrcFile generates the .taskrc.yaml file from template
func generateTaskrcFile(ctx context.Context, engine cmdutil.TemplateEngineInterface, taskDir string, variables map[string]interface{}) error {
	// Try to load the template from assets
	tmpl, err := engine.LoadTemplate(ctx, "taskrc.yaml")
	if err != nil {
		// If template not found, use a simple fallback
		return generateFallbackTaskrcFile(taskDir, variables)
	}

	// Render the template
	content, err := engine.RenderTemplate(ctx, tmpl, variables)
	if err != nil {
		return fmt.Errorf("failed to render taskrc template: %w", err)
	}

	// Write to file
	taskrcPath := filepath.Join(taskDir, ".taskrc.yaml")
	return os.WriteFile(taskrcPath, []byte(content), 0644)
}

// generateFallbackTaskrcFile generates a simple .taskrc.yaml when template is not available
func generateFallbackTaskrcFile(taskDir string, variables map[string]interface{}) error {
	content := fmt.Sprintf(`# Task Configuration - .taskrc.yaml
# Task-specific configuration and automation settings

# Task identification
task:
  id: "%s"
  type: "%s"

# Automation settings
automation:
  # External system synchronization
  sync:
    enabled: false
    frequency: "daily"
    direction: "bidirectional"
    systems: []

  # Workflow automation
  workflow:
    auto_progress: false
    validate_on_transition: true
    generate_artifacts: true
    required_approvals: {}

  # AI agent configuration
  agents:
    enabled: false
    agents: []
    context_sharing: false
    auto_enhance: false

# Quality standards
quality:
  # Code quality
  code:
    coverage_threshold: 80
    linting_enabled: true
    complexity_threshold: 10

  # Security requirements
  security:
    scan_required: true
    vulnerability_threshold: "medium"
    compliance_standards: []

  # Performance requirements
  performance:
    response_time_p95: 500
    throughput_minimum: 100
    error_rate_maximum: 0.01

  # Accessibility requirements
  accessibility:
    wcag_level: "AA"
    automated_testing: true
    manual_review_required: false

# Notification settings
notifications:
  # Notification channels
  channels: []

  # Notification recipients by event
  recipients:
    stage_completion:
      - "%s"
    quality_gate_failure:
      - "%s"
    blocker_added:
      - "%s"
    task_completed:
      - "%s"

# Template preferences
templates:
  # Preferred templates by artifact type
  preferences: {}

  # Template variables
  variables: {}

  # Auto-generation settings
  auto_generate: []

# Environment configuration
environments: {}

# Feature flags
feature_flags: []

# Custom scripts and hooks
hooks:
  # Pre-stage hooks
  pre_stage: {}

  # Post-stage hooks
  post_stage: {}

  # Custom commands
  commands: []

# Reporting configuration
reporting:
  # Status report generation
  status_reports:
    enabled: false
    frequency: "weekly"
    recipients:
      - "%s"

  # Metrics tracking
  metrics:
    track_velocity: true
    track_cycle_time: true
    track_defects: true
    custom_metrics: []

# Resource allocation
resources:
  # Team allocation
  team_allocation: []

  # Budget tracking
  budget:
    allocated: 0
    spent: 0
    currency: "USD"

# Custom configuration
custom: {}
`,
		variables["TASK_ID"],
		variables["TASK_TYPE"],
		variables["OWNER_NAME"],
		variables["OWNER_NAME"],
		variables["OWNER_NAME"],
		variables["OWNER_NAME"],
		variables["OWNER_NAME"],
	)

	taskrcPath := filepath.Join(taskDir, ".taskrc.yaml")
	return os.WriteFile(taskrcPath, []byte(content), 0644)
}
