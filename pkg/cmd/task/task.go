package task

import (
	"github.com/daddia/zen/pkg/cmd/task/create"
	"github.com/daddia/zen/pkg/cmdutil"
	"github.com/spf13/cobra"
)

// NewCmdTask creates the task command with subcommands
func NewCmdTask(f *cmdutil.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task <command>",
		Short: "Manage tasks and workflow",
		Long: `Manage tasks and workflow for Zen CLI.

Tasks are the core unit of work in Zen, following the seven-stage Zenflow
workflow: Align → Discover → Prioritize → Design → Build → Ship → Learn.

Each task creates a minimal directory in .zen/tasks/ with:
- index.md: Human-readable task overview
- manifest.yaml: Machine-readable metadata and workflow state
- .taskrc.yaml: Task-specific configuration
- .zenflow/: Workflow state tracking
- metadata/: External system snapshots

Work-type directories (research/, spikes/, design/, execution/, outcomes/)
are created on-demand when artifacts are added.

Tasks support different types:
- story: User-facing feature development
- bug: Defect fixes and corrections
- epic: Large initiatives spanning multiple tasks
- spike: Research and exploration work
- task: General work items`,
		Example: `  # Create a new story task
  zen task create PROJ-123 --type story

  # Create a bug fix task
  zen task create BUG-456 --type bug

  # Create an epic for large initiatives
  zen task create EPIC-789 --type epic

  # Create a research spike
  zen task create SPIKE-101 --type spike`,
		GroupID: "core",
	}

	// Add subcommands
	cmd.AddCommand(create.NewCmdTaskCreate(f))

	return cmd
}
