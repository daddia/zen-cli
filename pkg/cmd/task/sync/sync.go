package sync

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/daddia/zen/pkg/iostreams"
	"github.com/daddia/zen/pkg/task"
	"github.com/spf13/cobra"
)

// SyncOptions contains options for the task sync command
type SyncOptions struct {
	IO      *iostreams.IOStreams
	Factory *cmdutil.Factory

	Direction        string   // pull, push, bidirectional
	ConflictStrategy string   // local_wins, remote_wins, manual_review, timestamp
	Sources          []string // Specific sources to sync
	DryRun           bool
	Force            bool
	All              bool // Sync all tasks
}

// NewCmdTaskSync creates the task sync command
func NewCmdTaskSync(f *cmdutil.Factory) *cobra.Command {
	opts := &SyncOptions{
		IO:      f.IOStreams,
		Factory: f,
		DryRun:  f.DryRun,
	}

	cmd := &cobra.Command{
		Use:   "sync [task-id]",
		Short: "Synchronize tasks with external source systems",
		Long: `Synchronize task data between Zen and external source systems.

This command can:
- Pull latest data from external sources (jira, github, linear)
- Push local changes to external sources
- Perform bidirectional sync with conflict resolution
- Sync specific tasks or all tasks in the workspace

Sync directions:
- pull: Fetch latest data from external sources
- push: Send local changes to external sources  
- bidirectional: Two-way sync with conflict resolution

Conflict strategies:
- local_wins: Keep local changes, discard remote
- remote_wins: Accept remote changes, discard local
- timestamp: Use most recent timestamp
- manual_review: Create conflict records for manual resolution`,
		Example: heredoc.Doc(`
			# Sync specific task with external sources
			zen task sync ZEN-123

			# Pull latest data from all sources
			zen task sync ZEN-123 --direction pull

			# Push local changes to external sources
			zen task sync ZEN-123 --direction push

			# Bidirectional sync with timestamp conflict resolution
			zen task sync ZEN-123 --direction bidirectional --conflict-strategy timestamp

			# Sync all tasks in workspace
			zen task sync --all

			# Sync only with specific sources
			zen task sync ZEN-123 --sources jira,github

			# Dry run to see what would be synced
			zen task sync ZEN-123 --dry-run
		`),
		Args: func(cmd *cobra.Command, args []string) error {
			if opts.All && len(args) > 0 {
				return fmt.Errorf("cannot specify task ID when using --all flag")
			}
			if !opts.All && len(args) == 0 {
				return fmt.Errorf("task ID required unless using --all flag")
			}
			if !opts.All && len(args) > 1 {
				return fmt.Errorf("only one task ID allowed")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.All {
				return syncAllRun(opts)
			} else {
				taskID := args[0]
				return syncTaskRun(opts, taskID)
			}
		},
	}

	// Add flags
	cmd.Flags().StringVarP(&opts.Direction, "direction", "d", "bidirectional", "Sync direction (pull|push|bidirectional)")
	cmd.Flags().StringVar(&opts.ConflictStrategy, "conflict-strategy", "timestamp", "Conflict resolution strategy (local_wins|remote_wins|timestamp|manual_review)")
	cmd.Flags().StringSliceVar(&opts.Sources, "sources", nil, "Specific sources to sync (comma-separated)")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Force sync even if conflicts exist")
	cmd.Flags().BoolVar(&opts.All, "all", false, "Sync all tasks in workspace")

	return cmd
}

// syncTaskRun executes task synchronization for a specific task
func syncTaskRun(opts *SyncOptions, taskID string) error {
	ctx := context.Background()

	// Create task manager
	taskManager := task.NewManager(opts.Factory)

	// Validate sync direction
	direction, err := parseSyncDirection(opts.Direction)
	if err != nil {
		return fmt.Errorf("invalid sync direction: %w", err)
	}

	// Validate conflict strategy
	conflictStrategy, err := parseConflictStrategy(opts.ConflictStrategy)
	if err != nil {
		return fmt.Errorf("invalid conflict strategy: %w", err)
	}

	if opts.DryRun {
		fmt.Fprintf(opts.IO.Out, "%s Sync plan for task %s:\n",
			opts.IO.FormatSuccess(""),
			opts.IO.ColorBold(taskID))
		fmt.Fprintf(opts.IO.Out, "  %s Direction: %s\n",
			opts.IO.ColorNeutral("→"), opts.Direction)
		fmt.Fprintf(opts.IO.Out, "  %s Conflict Strategy: %s\n",
			opts.IO.ColorNeutral("→"), opts.ConflictStrategy)
		if len(opts.Sources) > 0 {
			fmt.Fprintf(opts.IO.Out, "  %s Sources: %v\n",
				opts.IO.ColorNeutral("→"), opts.Sources)
		}
		fmt.Fprintf(opts.IO.Out, "\n%s Run without --dry-run to execute sync\n",
			opts.IO.ColorInfo("ℹ"))
		return nil
	}

	// Execute sync
	syncOpts := &task.SyncOptions{
		Direction:        direction,
		ConflictStrategy: conflictStrategy,
		DryRun:           opts.DryRun,
		Force:            opts.Force,
		Sources:          opts.Sources,
	}

	fmt.Fprintf(opts.IO.Out, "%s Syncing task %s with external sources...\n",
		opts.IO.ColorInfo("ℹ"), taskID)

	result, err := taskManager.SyncTask(ctx, taskID, syncOpts)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	// Display results
	if result.Success {
		fmt.Fprintf(opts.IO.Out, "%s Sync completed successfully\n",
			opts.IO.FormatSuccess("✓"))
		fmt.Fprintf(opts.IO.Out, "  %s Task: %s\n",
			opts.IO.ColorNeutral("→"), result.TaskID)
		fmt.Fprintf(opts.IO.Out, "  %s Source: %s\n",
			opts.IO.ColorNeutral("→"), result.Source)
		fmt.Fprintf(opts.IO.Out, "  %s Direction: %s\n",
			opts.IO.ColorNeutral("→"), result.Direction)
		if len(result.ChangedFields) > 0 {
			fmt.Fprintf(opts.IO.Out, "  %s Changed fields: %v\n",
				opts.IO.ColorNeutral("→"), result.ChangedFields)
		}
		fmt.Fprintf(opts.IO.Out, "  %s Duration: %v\n",
			opts.IO.ColorNeutral("→"), result.Duration)
	} else {
		fmt.Fprintf(opts.IO.Out, "%s Sync failed: %s\n",
			opts.IO.FormatError("✗"), result.Error)
	}

	return nil
}

// syncAllRun executes synchronization for all tasks
func syncAllRun(opts *SyncOptions) error {
	ctx := context.Background()

	// Create task manager
	taskManager := task.NewManager(opts.Factory)

	// Validate sync direction
	direction, err := parseSyncDirection(opts.Direction)
	if err != nil {
		return fmt.Errorf("invalid sync direction: %w", err)
	}

	// Validate conflict strategy
	conflictStrategy, err := parseConflictStrategy(opts.ConflictStrategy)
	if err != nil {
		return fmt.Errorf("invalid conflict strategy: %w", err)
	}

	if opts.DryRun {
		fmt.Fprintf(opts.IO.Out, "%s Sync plan for all tasks:\n",
			opts.IO.FormatSuccess(""))
		fmt.Fprintf(opts.IO.Out, "  %s Direction: %s\n",
			opts.IO.ColorNeutral("→"), opts.Direction)
		fmt.Fprintf(opts.IO.Out, "  %s Conflict Strategy: %s\n",
			opts.IO.ColorNeutral("→"), opts.ConflictStrategy)
		fmt.Fprintf(opts.IO.Out, "\n%s Run without --dry-run to execute sync\n",
			opts.IO.ColorInfo("ℹ"))
		return nil
	}

	// Execute sync for all tasks
	syncOpts := &task.SyncOptions{
		Direction:        direction,
		ConflictStrategy: conflictStrategy,
		DryRun:           opts.DryRun,
		Force:            opts.Force,
		Sources:          opts.Sources,
	}

	fmt.Fprintf(opts.IO.Out, "%s Syncing all tasks with external sources...\n",
		opts.IO.ColorInfo("ℹ"))

	results, err := taskManager.SyncAllTasks(ctx, syncOpts)
	if err != nil {
		return fmt.Errorf("sync all failed: %w", err)
	}

	// Display results
	successful := 0
	failed := 0
	for _, result := range results {
		if result.Success {
			successful++
		} else {
			failed++
		}
	}

	fmt.Fprintf(opts.IO.Out, "%s Sync completed\n",
		opts.IO.FormatSuccess("✓"))
	fmt.Fprintf(opts.IO.Out, "  %s Total tasks: %d\n",
		opts.IO.ColorNeutral("→"), len(results))
	fmt.Fprintf(opts.IO.Out, "  %s Successful: %d\n",
		opts.IO.ColorNeutral("→"), successful)
	if failed > 0 {
		fmt.Fprintf(opts.IO.Out, "  %s Failed: %d\n",
			opts.IO.ColorWarning("!"), failed)
	}

	return nil
}

// Helper functions

func parseSyncDirection(direction string) (task.SyncDirection, error) {
	switch direction {
	case "pull":
		return task.SyncDirectionPull, nil
	case "push":
		return task.SyncDirectionPush, nil
	case "bidirectional":
		return task.SyncDirectionBidirectional, nil
	default:
		return "", fmt.Errorf("invalid direction '%s', must be one of: pull, push, bidirectional", direction)
	}
}

func parseConflictStrategy(strategy string) (task.ConflictStrategy, error) {
	switch strategy {
	case "local_wins":
		return task.ConflictStrategyLocalWins, nil
	case "remote_wins":
		return task.ConflictStrategyRemoteWins, nil
	case "timestamp":
		return task.ConflictStrategyTimestamp, nil
	case "manual_review":
		return task.ConflictStrategyManualReview, nil
	default:
		return "", fmt.Errorf("invalid strategy '%s', must be one of: local_wins, remote_wins, timestamp, manual_review", strategy)
	}
}
