package cli

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	syncpkg "github.com/arianlopezc/Trabuco/internal/sync"
	"github.com/spf13/cobra"
)

var (
	syncApply bool
	syncJSON  bool
)

var syncCmd = &cobra.Command{
	Use:   "sync [PATH]",
	Short: "Add missing AI-tooling files to an existing Trabuco project",
	Long: `Bring a project's AI-tooling files up to date with what the installed
Trabuco CLI would generate for the same module and agent selection.

Sync is additive for almost every file: it creates files the current CLI
would generate that the project is missing, and never modifies existing
files. To refresh a file like CLAUDE.md with newer content, delete the
file and re-run sync.

The single exception is .gitignore, which is updated in place between
two Trabuco-managed marker comments. Lines outside those markers are
user-owned and untouched.

Scope: .ai/**, .claude/**, .cursor/**, .codex/**, .agents/**, .github/instructions/**,
.github/scripts/review-checks.sh, .github/skills/**, .github/workflows/copilot-setup-steps.yml,
.trabuco/review.config.json, CLAUDE.md, AGENTS.md, .github/copilot-instructions.md,
and the Trabuco-managed block in .gitignore.

Out of scope: Java source, POMs, Flyway migrations, application.yml,
docker-compose.yml, CI workflows (other than copilot-setup-steps.yml),
.env files, .run/ configs, README.md, .trabuco.json. These are either user-
owned or infrastructure; sync will never touch them.

Usage:
  trabuco sync              # dry-run — show what would be added
  trabuco sync --apply      # actually create missing files
  trabuco sync --json       # machine-readable plan (for CI or agents)`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncApply, "apply", false, "Apply the plan (create missing files). Without this flag, sync runs as dry-run.")
	syncCmd.Flags().BoolVar(&syncJSON, "json", false, "Emit the plan as JSON for machine consumption.")
}

func runSync(cmd *cobra.Command, args []string) error {
	projectPath := "."
	if len(args) == 1 {
		projectPath = args[0]
	}

	plan, err := syncpkg.Run(projectPath, Version, syncApply)
	if err != nil {
		return fmt.Errorf("sync failed: %w", err)
	}

	if syncJSON {
		return plan.WriteJSON(os.Stdout)
	}

	if err := plan.WritePretty(os.Stdout); err != nil {
		return err
	}

	// Human-readable apply confirmation, only when we actually wrote files.
	if syncApply && !plan.Blocked() && plan.HasWork() {
		green := color.New(color.FgGreen)
		fmt.Println()
		if len(plan.WouldAdd) > 0 {
			green.Printf("Added %d files.\n", len(plan.WouldAdd))
		}
		if len(plan.WouldUpdate) > 0 {
			green.Printf("Updated %d files (Trabuco-managed block).\n", len(plan.WouldUpdate))
		}
	} else if !syncApply && !plan.Blocked() && plan.HasWork() {
		fmt.Println()
		fmt.Println("Run `trabuco sync --apply` to create or update these files.")
	}

	if plan.Blocked() {
		os.Exit(1)
	}
	return nil
}
