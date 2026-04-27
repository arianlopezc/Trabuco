package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/arianlopezc/Trabuco/internal/migration/orchestrator"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
	"github.com/arianlopezc/Trabuco/internal/migration/vcs"

	// Specialist registrations (each milestone wires its specialists here):
	_ "github.com/arianlopezc/Trabuco/internal/migration/specialists/registry"
)

// migrateCmd is the new top-level command for the 1.10.0 migration feature.
// It orchestrates the 14-phase flow defined in docs/MIGRATION_REDESIGN_PLAN.md.
//
// Subcommands map 1:1 to the MCP tools so plugin and CLI mode are at parity.
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate an existing Java repo into a Trabuco-shaped project (in place)",
	Long: `Run the 14-phase Trabuco migration on an existing repository.

The migration transforms the user's repo in place into a Trabuco multi-module
Maven project: assessment, skeleton bootstrap, then per-module migrations
(Model, Datastore, Shared, API, Worker, EventConsumer, AIAgent), config,
deployment adaptation, test analysis, enforcement activation, and finalization.
Each phase ends with an approval gate (approve / edit-and-approve / reject).

Subcommands run individual phases. The typical end-to-end flow is:
  trabuco migrate assess    /path/to/repo
  trabuco migrate skeleton  /path/to/repo
  trabuco migrate module    /path/to/repo --module=model
  ...
  trabuco migrate finalize  /path/to/repo

Or run autopilot through every phase, gating at each:
  trabuco migrate run       /path/to/repo

State lives at .trabuco-migration/ inside the repo. Per-phase git tags
(trabuco-migration-phase-N-pre/post) provide atomic rollback boundaries.

See docs/MIGRATION_REDESIGN_PLAN.md for the full design.`,
}

func init() {
	migrateCmd.AddCommand(migrateAssessCmd)
	migrateCmd.AddCommand(migrateSkeletonCmd)
	migrateCmd.AddCommand(migrateModuleCmd)
	migrateCmd.AddCommand(migrateConfigCmd)
	migrateCmd.AddCommand(migrateDeploymentCmd)
	migrateCmd.AddCommand(migrateTestsCmd)
	migrateCmd.AddCommand(migrateActivateCmd)
	migrateCmd.AddCommand(migrateFinalizeCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateRollbackCmd)
	migrateCmd.AddCommand(migrateDecisionCmd)
	migrateCmd.AddCommand(migrateResumeCmd)
	migrateCmd.AddCommand(migrateRunCmd)

	rootCmd.AddCommand(migrateCmd)

	migrateModuleCmd.Flags().String("module", "", "Module to migrate (model|sqldatastore|nosqldatastore|shared|api|worker|eventconsumer|aiagent)")
	migrateRollbackCmd.Flags().Int("to-phase", -1, "Phase number to roll back to (0..13)")
	migrateDecisionCmd.Flags().String("id", "", "Decision ID to record")
	migrateDecisionCmd.Flags().String("choice", "", "Choice value")
}

// ---------- subcommand definitions ----------

var migrateAssessCmd = &cobra.Command{
	Use:   "assess <repo-path>",
	Short: "Phase 0 — Intake & Assessment (LLM scans the source, produces assessment.json)",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runPhase(cmd, args[0], types.PhaseAssessment) },
}

var migrateSkeletonCmd = &cobra.Command{
	Use:   "skeleton <repo-path>",
	Short: "Phase 1 — Skeleton bootstrap (multi-module structure, migration-mode parent POM, legacy/ wrap)",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runPhase(cmd, args[0], types.PhaseSkeleton) },
}

var migrateModuleCmd = &cobra.Command{
	Use:   "module <repo-path> --module=<name>",
	Short: "Phase 2-8 — Migrate a single module (Model, Datastore, Shared, API, Worker, EventConsumer, AIAgent)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modName, _ := cmd.Flags().GetString("module")
		phase, err := phaseForModuleName(modName)
		if err != nil {
			return err
		}
		return runPhase(cmd, args[0], phase)
	},
}

var migrateConfigCmd = &cobra.Command{
	Use:   "config <repo-path>",
	Short: "Phase 9 — Configuration (per-module application.yml, OpenTelemetry, structured logging)",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runPhase(cmd, args[0], types.PhaseConfiguration) },
}

var migrateDeploymentCmd = &cobra.Command{
	Use:   "deployment <repo-path>",
	Short: "Phase 10 — Adapt legacy CI/CD to multi-module structure (legacy CI only; never invents pipelines)",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runPhase(cmd, args[0], types.PhaseDeployment) },
}

var migrateTestsCmd = &cobra.Command{
	Use:   "tests <repo-path>",
	Short: "Phase 11 — Per-test analysis: keep/adapt/discard/characterize",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runPhase(cmd, args[0], types.PhaseTests) },
}

var migrateActivateCmd = &cobra.Command{
	Use:   "activate <repo-path>",
	Short: "Phase 12 — Enforcement activation (flips Maven Enforcer / Spotless / ArchUnit / Jacoco threshold ON)",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runPhase(cmd, args[0], types.PhaseActivation) },
}

var migrateFinalizeCmd = &cobra.Command{
	Use:   "finalize <repo-path>",
	Short: "Phase 13 — Finalization (doctor, sync, completion report)",
	Args:  cobra.ExactArgs(1),
	RunE:  func(cmd *cobra.Command, args []string) error { return runPhase(cmd, args[0], types.PhaseFinalization) },
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status <repo-path>",
	Short: "Show current migration state",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := absRepoPath(args[0])
		if err != nil {
			return err
		}
		s, err := state.Load(repoRoot)
		if err != nil {
			return fmt.Errorf("no migration in progress at %s: %w", repoRoot, err)
		}
		return printStatus(s)
	},
}

var migrateRollbackCmd = &cobra.Command{
	Use:   "rollback <repo-path> --to-phase=<N>",
	Short: "Roll back to the pre-tag of phase N",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		toPhase, _ := cmd.Flags().GetInt("to-phase")
		if toPhase < 0 || toPhase > 13 {
			return fmt.Errorf("--to-phase must be 0..13")
		}
		repoRoot, err := absRepoPath(args[0])
		if err != nil {
			return err
		}
		o := newOrch(repoRoot)
		return o.Rollback(types.Phase(toPhase))
	},
}

var migrateDecisionCmd = &cobra.Command{
	Use:   "decision <repo-path> --id=<id> --choice=<value>",
	Short: "Record a user choice for a requires-decision item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, _ := cmd.Flags().GetString("id")
		choice, _ := cmd.Flags().GetString("choice")
		if id == "" || choice == "" {
			return fmt.Errorf("--id and --choice are required")
		}
		repoRoot, err := absRepoPath(args[0])
		if err != nil {
			return err
		}
		o := newOrch(repoRoot)
		return o.RecordDecision(state.DecisionRecord{ID: id, Choice: choice})
	},
}

var migrateResumeCmd = &cobra.Command{
	Use:   "resume <repo-path>",
	Short: "Resume from the most recent in-progress phase",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := absRepoPath(args[0])
		if err != nil {
			return err
		}
		s, err := state.Load(repoRoot)
		if err != nil {
			return err
		}
		for _, p := range types.AllPhases() {
			rec := s.Phases[p]
			if rec.State == types.PhaseInProgress || rec.State == types.PhaseFailed {
				fmt.Printf("Resuming phase %s (was %s)\n", p, rec.State)
				return runPhase(cmd, repoRoot, p)
			}
		}
		fmt.Println("No in-progress phase to resume.")
		return nil
	},
}

var migrateRunCmd = &cobra.Command{
	Use:   "run <repo-path>",
	Short: "Run all phases sequentially, gating at each (use --auto-approve to skip gates — DANGEROUS)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		repoRoot, err := absRepoPath(args[0])
		if err != nil {
			return err
		}
		o := newOrch(repoRoot)
		ctx := context.Background()
		for _, p := range types.AllPhases() {
			fmt.Printf("\n=== Phase %d (%s) ===\n", int(p), p)
			action, err := o.RunPhase(ctx, p, "")
			if err != nil {
				return err
			}
			if action == types.GateReject {
				fmt.Printf("Phase %s rejected; halting migration.\n", p)
				return nil
			}
		}
		fmt.Println("\nMigration complete. See .trabuco-migration/completion-report.md")
		return nil
	},
}

// ---------- helpers ----------

func runPhase(cmd *cobra.Command, repoArg string, phase types.Phase) error {
	repoRoot, err := absRepoPath(repoArg)
	if err != nil {
		return err
	}
	o := newOrch(repoRoot)
	if !state.Exists(repoRoot) {
		// Auto-init at first phase only.
		if phase != types.PhaseAssessment {
			return fmt.Errorf("no migration initialized; run 'trabuco migrate assess %s' first", repoRoot)
		}
		// Preflight before initial init.
		if err := o.Preflight(); err != nil {
			return err
		}
		// Bootstrap with empty target config; the assessor recommends one
		// and the skeleton phase reads the user-approved config from
		// state.json.
		if _, err := o.Init(state.TargetConfig{}); err != nil {
			return err
		}
	}
	action, err := o.RunPhase(cmd.Context(), phase, "")
	if err != nil {
		return err
	}
	fmt.Printf("Phase %s: %s\n", phase, action)
	return nil
}

func absRepoPath(p string) (string, error) {
	abs, err := filepath.Abs(p)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(abs); err != nil {
		return "", err
	}
	if !vcs.IsRepo(abs) {
		return "", fmt.Errorf("%s is not a git repository", abs)
	}
	return abs, nil
}

func newOrch(repoRoot string) *orchestrator.Orchestrator {
	return orchestrator.New(repoRoot, Version, specialists.Default(), terminalGate{})
}

func phaseForModuleName(name string) (types.Phase, error) {
	switch strings.ToLower(name) {
	case "model":
		return types.PhaseModel, nil
	case "sqldatastore", "nosqldatastore", "datastore":
		return types.PhaseDatastore, nil
	case "shared":
		return types.PhaseShared, nil
	case "api":
		return types.PhaseAPI, nil
	case "worker":
		return types.PhaseWorker, nil
	case "eventconsumer":
		return types.PhaseEventConsumer, nil
	case "aiagent":
		return types.PhaseAIAgent, nil
	default:
		return 0, fmt.Errorf("unknown --module=%q (must be one of: model, sqldatastore, nosqldatastore, shared, api, worker, eventconsumer, aiagent)", name)
	}
}

func printStatus(s *state.State) error {
	out, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	_, _ = os.Stdout.Write(out)
	_, _ = os.Stdout.WriteString("\n")
	return nil
}

// terminalGate is the CLI-mode Gate implementation: presents the diff
// and approval prompt as terminal output and reads a line of stdin for
// the user's choice.
type terminalGate struct{}

func (terminalGate) Present(ctx context.Context, phase types.Phase, out *specialists.Output) (types.GateAction, string, error) {
	fmt.Printf("\n--- Phase %d (%s) summary ---\n", int(phase), phase)
	fmt.Println(out.Summary)
	fmt.Printf("\nItems: %d\n", len(out.Items))
	for _, item := range out.Items {
		fmt.Printf("  [%s] %s\n", item.State, item.Description)
	}
	if len(out.Decisions) > 0 {
		fmt.Println("\nDecisions required:")
		for _, d := range out.Decisions {
			fmt.Printf("  - %s: %s\n    choices: %s\n", d.ID, d.Question, strings.Join(d.Choices, ", "))
		}
	}
	fmt.Print("\n[a]pprove / [e]dit and approve / [r]eject? ")
	var choice string
	if _, err := fmt.Scanln(&choice); err != nil {
		return "", "", err
	}
	switch strings.ToLower(strings.TrimSpace(choice)) {
	case "a", "approve":
		return types.GateApprove, "", nil
	case "e", "edit":
		fmt.Print("Provide guidance for the specialist: ")
		var hint string
		if _, err := fmt.Scanln(&hint); err != nil {
			return "", "", err
		}
		return types.GateEditAndApprove, hint, nil
	case "r", "reject":
		return types.GateReject, "", nil
	default:
		return "", "", fmt.Errorf("unrecognized choice %q", choice)
	}
}

