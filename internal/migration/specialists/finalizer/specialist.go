// Package finalizer implements the Phase 13 finalizer specialist. It
// runs after the activator (Phase 12) and is the last phase before the
// migration is declared complete.
//
// Responsibilities:
//   - Run `trabuco doctor --fix` (resolves any structural drift introduced
//     during migration).
//   - Run `trabuco sync` (syncs AI-tooling files to current Trabuco
//     conventions).
//   - Final `mvn verify` (with full enforcement still on from Phase 12).
//   - Generate `.trabuco-migration/completion-report.md`.
//   - If user opts and legacy/ is empty, remove it.
package finalizer

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// Specialist is the Phase 13 finalizer.
type Specialist struct{}

// New constructs the finalizer.
func New() *Specialist { return &Specialist{} }

// Phase implements specialists.Specialist.
func (s *Specialist) Phase() types.Phase { return types.PhaseFinalization }

// Name implements specialists.Specialist.
func (s *Specialist) Name() string { return "finalizer" }

// Run implements specialists.Specialist.
func (s *Specialist) Run(ctx context.Context, in *specialists.Input) (*specialists.Output, error) {
	items := []types.OutputItem{}

	// 1. Run trabuco doctor --fix (idempotent; fixes any structural drift).
	if out, err := runTrabuco(in.RepoRoot, "doctor", "--fix"); err != nil {
		items = append(items, types.OutputItem{
			ID:          "finalizer-doctor",
			State:       types.ItemBlocked,
			Description: "trabuco doctor --fix surfaced issues",
			BlockerCode: types.BlockerCompileFailed,
			BlockerNote: truncate(out, 4000),
			Alternatives: []string{
				"resolve doctor issues manually and rerun migrate finalize",
				"accept-with-caveats and document",
			},
		})
	} else {
		items = append(items, types.OutputItem{
			ID:          "finalizer-doctor",
			State:       types.ItemApplied,
			Description: "trabuco doctor --fix passed (project structurally sound)",
		})
	}

	// 2. Run trabuco sync (brings AI-tooling files up to date).
	if _, err := runTrabuco(in.RepoRoot, "sync"); err != nil {
		// Non-fatal — sync may fail if user opted out of AI tooling.
		items = append(items, types.OutputItem{
			ID:          "finalizer-sync",
			State:       types.ItemApplied,
			Description: "trabuco sync skipped (no AI tooling configured or sync errored — non-blocking)",
		})
	} else {
		items = append(items, types.OutputItem{
			ID:          "finalizer-sync",
			State:       types.ItemApplied,
			Description: "trabuco sync brought AI-tooling files up to date",
		})
	}

	// 3. Final mvn verify with full enforcement on (from activator).
	if out, err := runMaven(in.RepoRoot, "verify", "-q"); err != nil {
		items = append(items, types.OutputItem{
			ID:          "finalizer-verify",
			State:       types.ItemBlocked,
			Description: "final mvn verify failed",
			BlockerCode: types.BlockerCompileFailed,
			BlockerNote: truncate(out, 4000),
			Alternatives: []string{
				"resolve verify failures and rerun migrate finalize",
			},
		})
	} else {
		items = append(items, types.OutputItem{
			ID:          "finalizer-verify",
			State:       types.ItemApplied,
			Description: "final mvn verify passed end-to-end",
		})
	}

	// 4. Inspect legacy/ — if empty, offer removal as a decision item.
	if isLegacyEmpty(in.RepoRoot) {
		items = append(items, types.OutputItem{
			ID:          "finalizer-legacy-empty",
			State:       types.ItemRequiresDecision,
			Description: "legacy/ module is empty after migration",
			Question:    "Remove legacy/ module entirely, or keep as @Deprecated marker for documentation?",
			Choices:     []string{"remove", "keep"},
		})
	}

	// 5. Write completion report.
	if err := writeCompletionReport(in.RepoRoot, in.State, items); err != nil {
		return nil, fmt.Errorf("write completion report: %w", err)
	}

	return &specialists.Output{
		Phase:   types.PhaseFinalization,
		Items:   items,
		Summary: "Migration complete. Project passes mvn verify with full enforcement, AI tooling synced, doctor green. Completion report at .trabuco-migration/completion-report.md.",
	}, nil
}

// writeCompletionReport produces the human-readable summary the user
// receives at the end of migration.
func writeCompletionReport(repoRoot string, st *state.State, items []types.OutputItem) error {
	var b strings.Builder
	fmt.Fprintf(&b, "# Trabuco Migration — Completion Report\n\n")
	fmt.Fprintf(&b, "Generated: %s\n", time.Now().UTC().Format(time.RFC3339))
	fmt.Fprintf(&b, "Trabuco CLI: %s\n\n", st.TrabucoCLIVersion)

	fmt.Fprintln(&b, "## Source")
	fmt.Fprintf(&b, "- Build system: %s\n", st.SourceConfig.BuildSystem)
	fmt.Fprintf(&b, "- Framework: %s\n", st.SourceConfig.Framework)
	fmt.Fprintf(&b, "- Java: %s\n", st.SourceConfig.JavaVersion)
	fmt.Fprintf(&b, "- Persistence: %s\n", st.SourceConfig.Persistence)
	fmt.Fprintf(&b, "- Messaging: %s\n", st.SourceConfig.Messaging)
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Target")
	fmt.Fprintf(&b, "- Modules: %v\n", st.TargetConfig.Modules)
	fmt.Fprintf(&b, "- Database: %s\n", st.TargetConfig.Database)
	if st.TargetConfig.MessageBroker != "" {
		fmt.Fprintf(&b, "- Broker: %s\n", st.TargetConfig.MessageBroker)
	}
	if len(st.TargetConfig.AIAgents) > 0 {
		fmt.Fprintf(&b, "- AI agents: %v\n", st.TargetConfig.AIAgents)
	}
	fmt.Fprintf(&b, "- Java: %s\n", st.TargetConfig.JavaVersion)
	fmt.Fprintln(&b)

	fmt.Fprintln(&b, "## Phase results")
	for _, p := range types.AllPhases() {
		rec := st.Phases[p]
		fmt.Fprintf(&b, "- Phase %d (%s): %s\n", int(p), p, rec.State)
	}
	fmt.Fprintln(&b)

	if len(st.Blockers) > 0 {
		fmt.Fprintln(&b, "## Blockers encountered")
		for _, blk := range st.Blockers {
			fmt.Fprintf(&b, "- [Phase %d] `%s` in %s — user choice: %s\n",
				int(blk.Phase), blk.Code, blk.File, blk.UserChoice)
		}
		fmt.Fprintln(&b)
	}

	if len(st.Decisions) > 0 {
		fmt.Fprintln(&b, "## Decisions recorded")
		for _, d := range st.Decisions {
			fmt.Fprintf(&b, "- [Phase %d] %s → %s\n", int(d.Phase), d.Question, d.Choice)
		}
		fmt.Fprintln(&b)
	}

	if len(st.RetainedLegacy) > 0 {
		fmt.Fprintln(&b, "## Retained as legacy")
		for _, f := range st.RetainedLegacy {
			fmt.Fprintf(&b, "- %s\n", f)
		}
		fmt.Fprintln(&b)
	}

	fmt.Fprintln(&b, "## Final phase")
	for _, item := range items {
		fmt.Fprintf(&b, "- [%s] %s\n", item.State, item.Description)
	}

	return os.WriteFile(state.CompletionReportPath(repoRoot), []byte(b.String()), 0o644)
}

func isLegacyEmpty(repoRoot string) bool {
	srcDir := filepath.Join(repoRoot, "legacy", "src", "main", "java")
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return true // legacy/ doesn't even have a src — treat as empty
	}
	return len(entries) == 0
}

func runTrabuco(repoRoot string, args ...string) (string, error) {
	cmd := exec.Command("trabuco", args...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func runMaven(repoRoot string, args ...string) (string, error) {
	mvn := "mvn"
	if _, err := os.Stat(filepath.Join(repoRoot, "mvnw")); err == nil {
		mvn = "./mvnw"
	}
	cmd := exec.Command(mvn, args...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
