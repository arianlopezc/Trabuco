// Package orchestrator runs the migration's 14-phase state machine.
// One Orchestrator instance owns the migration for a single repo;
// it loads/saves state, dispatches to specialists, runs the validation
// funnel, presents gates, and manages git tags.
//
// In CLI mode the orchestrator is invoked by the trabuco migrate command
// and presents gates as terminal prompts. In plugin mode the orchestrator
// is the trabuco-migration-orchestrator subagent invoking the same handlers
// via MCP tools and presenting gates as natural-language exchanges.
package orchestrator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
	"github.com/arianlopezc/Trabuco/internal/migration/validation"
	"github.com/arianlopezc/Trabuco/internal/migration/vcs"
)

// Orchestrator owns one migration run.
type Orchestrator struct {
	repoRoot   string
	cliVersion string
	registry   *specialists.Registry
	gate       Gate
}

// Gate abstracts the user-approval surface. CLI mode supplies a terminal
// prompt implementation; plugin mode supplies a stub that defers to the
// subagent (which presents the gate via natural language and calls
// migrate_decision/migrate_rollback to record the choice).
type Gate interface {
	Present(ctx context.Context, phase types.Phase, out *specialists.Output) (types.GateAction, string, error)
}

// New constructs an Orchestrator. Specialists must be registered before
// running phases.
func New(repoRoot, cliVersion string, registry *specialists.Registry, gate Gate) *Orchestrator {
	return &Orchestrator{
		repoRoot:   repoRoot,
		cliVersion: cliVersion,
		registry:   registry,
		gate:       gate,
	}
}

// PreflightError is returned when a pre-Phase-0 hard gate fails.
type PreflightError struct{ Reason string }

func (e *PreflightError) Error() string { return "preflight: " + e.Reason }

// Preflight runs the pre-Phase-0 hard checks from §5 of the plan: clean
// working tree, on-branch (not detached), at least one commit, no prior
// .trabuco-migration/ unless explicitly resumable.
func (o *Orchestrator) Preflight() error {
	if !vcs.IsRepo(o.repoRoot) {
		return &PreflightError{"not a git repository"}
	}
	if !vcs.HasCommits(o.repoRoot) {
		return &PreflightError{"git repository has no commits"}
	}
	if _, err := vcs.CurrentBranch(o.repoRoot); err != nil {
		return &PreflightError{"detached HEAD — checkout a branch first"}
	}
	clean, err := vcs.IsClean(o.repoRoot)
	if err != nil {
		return &PreflightError{fmt.Sprintf("could not check working tree: %v", err)}
	}
	if !clean {
		return &PreflightError{"working tree has uncommitted changes — commit or stash first"}
	}
	return nil
}

// Init creates a fresh state.json under .trabuco-migration/. Refuses if
// state already exists; use Resume() to continue an existing run.
func (o *Orchestrator) Init(target state.TargetConfig) (*state.State, error) {
	if state.Exists(o.repoRoot) {
		return nil, errors.New("a migration is already in progress in this repo (use migrate resume to continue, or migrate rollback --to-phase=0 to restart)")
	}
	s := state.New(o.cliVersion)
	s.TargetConfig = target
	if err := state.Save(o.repoRoot, s); err != nil {
		return nil, err
	}
	return s, nil
}

// LoadState loads the current state.json.
func (o *Orchestrator) LoadState() (*state.State, error) {
	return state.Load(o.repoRoot)
}

// SaveState persists state.
func (o *Orchestrator) SaveState(s *state.State) error {
	return state.Save(o.repoRoot, s)
}

// RunPhase executes a single phase end-to-end: tags pre-state, invokes the
// specialist, validates output, presents the gate, commits or rolls back.
// Returns the user's gate action so the caller can decide whether to
// proceed to the next phase.
func (o *Orchestrator) RunPhase(ctx context.Context, phase types.Phase, hint string) (types.GateAction, error) {
	if err := state.AcquireLock(o.repoRoot, "cli"); err != nil {
		return "", err
	}
	defer state.ReleaseLock(o.repoRoot)

	s, err := o.LoadState()
	if err != nil {
		return "", err
	}
	rec := s.Phases[phase]

	// Idempotency: if phase is already completed, return without rerunning
	// unless the user is doing edit-and-approve (signaled by non-empty hint).
	if rec.State == types.PhaseCompleted && hint == "" {
		return types.GateApprove, nil
	}

	specialist := o.registry.Get(phase)
	if specialist == nil {
		return "", fmt.Errorf("no specialist registered for phase %s (this is a bug — milestone for that phase isn't shipped yet)", phase)
	}

	// Tag the pre-state so we can roll back atomically.
	preTag := vcs.PhasePreTag(phase)
	if !vcs.TagExists(o.repoRoot, preTag) {
		if err := vcs.CreateTag(o.repoRoot, preTag, fmt.Sprintf("trabuco migration: pre-%s", phase), false); err != nil {
			return "", fmt.Errorf("create pre-tag: %w", err)
		}
	}
	rec.State = types.PhaseInProgress
	rec.PreTag = preTag
	now := time.Now().UTC()
	rec.StartedAt = &now
	if err := o.SaveState(s); err != nil {
		return "", err
	}

	// Invoke the specialist.
	in := &specialists.Input{
		RepoRoot: o.repoRoot,
		Phase:    phase,
		State:    s,
		UserHint: hint,
	}
	if err := writeJSON(state.PhaseInputPath(o.repoRoot, phase), in); err != nil {
		return "", fmt.Errorf("write phase input: %w", err)
	}

	out, err := specialist.Run(ctx, in)
	if err != nil {
		rec.State = types.PhaseFailed
		_ = o.SaveState(s)
		return "", fmt.Errorf("specialist %s failed: %w", specialist.Name(), err)
	}
	if err := writeJSON(state.PhaseOutputPath(o.repoRoot, phase), out); err != nil {
		return "", fmt.Errorf("write phase output: %w", err)
	}

	// Handle the not-applicable happy path before validation.
	if isNotApplicable(out) {
		rec.State = types.PhaseNotApplicable
		rec.Reason = notApplicableReason(out)
		approvedAt := time.Now().UTC()
		rec.ApprovedAt = &approvedAt
		if err := o.SaveState(s); err != nil {
			return "", err
		}
		return types.GateApprove, nil
	}

	// Verify source evidence on every applied item (out-of-scope guard).
	for _, item := range out.Items {
		if item.State != types.ItemApplied || item.SourceEvidence == nil {
			continue
		}
		if err := validation.VerifyEvidence(o.repoRoot, item.SourceEvidence); err != nil {
			return "", fmt.Errorf("specialist %s emitted invalid source_evidence on item %s: %w", specialist.Name(), item.ID, err)
		}
	}

	// Apply file writes from each applied item. Specialists declare
	// the changes; the orchestrator materializes them. Rollback to
	// pre-tag if the validation funnel later fails.
	if err := applyFileWrites(o.repoRoot, out); err != nil {
		_ = vcs.ResetHard(o.repoRoot, preTag)
		rec.State = types.PhaseFailed
		_ = o.SaveState(s)
		return "", fmt.Errorf("apply file writes: %w (rolled back to %s)", err, preTag)
	}

	// Run the validation funnel (compile + tests). ArchUnit deferred
	// during migration phases.
	mode := validation.ModeMigration
	if phase == types.PhaseActivation {
		mode = validation.ModeActivation
	}
	res := validation.Run(o.repoRoot, mode, affectedModules(out))
	if !res.Passed {
		// Auto-rollback to pre-tag, surface failure as a blocker.
		_ = vcs.ResetHard(o.repoRoot, preTag)
		s.Blockers = append(s.Blockers, state.BlockerRecord{
			Phase:      phase,
			Code:       res.BlockerCode,
			Note:       fmt.Sprintf("validation funnel: %s\n\n%s", res.FailedStep, truncate(res.FailureLog, 4000)),
			RecordedAt: time.Now().UTC(),
		})
		rec.State = types.PhaseFailed
		_ = o.SaveState(s)
		return "", fmt.Errorf("validation funnel failed at %s (%s); state rolled back to %s", res.FailedStep, res.BlockerCode, preTag)
	}

	// Present the gate.
	action, editHint, err := o.gate.Present(ctx, phase, out)
	if err != nil {
		return "", err
	}

	switch action {
	case types.GateApprove:
		// Commit and tag post-state.
		commitMsg := fmt.Sprintf("trabuco migration: %s\n\n%s", phase, out.Summary)
		if err := vcs.CommitAll(o.repoRoot, commitMsg); err != nil {
			return "", fmt.Errorf("commit phase: %w", err)
		}
		postTag := vcs.PhasePostTag(phase)
		if err := vcs.CreateTag(o.repoRoot, postTag, fmt.Sprintf("trabuco migration: post-%s", phase), true); err != nil {
			return "", fmt.Errorf("create post-tag: %w", err)
		}
		rec.PostTag = postTag
		rec.State = types.PhaseCompleted
		approvedAt := time.Now().UTC()
		rec.ApprovedAt = &approvedAt
		if err := o.SaveState(s); err != nil {
			return "", err
		}
		// Write the human-readable phase report.
		_ = writeReport(state.PhaseReportPath(o.repoRoot, phase), phase, out, &res)
		return types.GateApprove, nil

	case types.GateEditAndApprove:
		// Roll back, then re-run with the user's hint.
		_ = vcs.ResetHard(o.repoRoot, preTag)
		rec.State = types.PhasePending
		rec.RetryCount++
		_ = o.SaveState(s)
		return o.RunPhase(ctx, phase, editHint)

	case types.GateReject:
		_ = vcs.ResetHard(o.repoRoot, preTag)
		rec.State = types.PhaseUserRejected
		_ = o.SaveState(s)
		return types.GateReject, nil
	}

	return "", fmt.Errorf("unknown gate action: %s", action)
}

// Rollback resets to the pre-tag for the given phase and clears later
// phases from state.json.
func (o *Orchestrator) Rollback(toPhase types.Phase) error {
	if err := state.AcquireLock(o.repoRoot, "cli"); err != nil {
		return err
	}
	defer state.ReleaseLock(o.repoRoot)

	s, err := o.LoadState()
	if err != nil {
		return err
	}
	rec, ok := s.Phases[toPhase]
	if !ok || rec.PreTag == "" {
		return fmt.Errorf("phase %s has no pre-tag to roll back to", toPhase)
	}
	if err := vcs.ResetHard(o.repoRoot, rec.PreTag); err != nil {
		return err
	}
	// Clear all phases >= toPhase.
	for _, p := range types.AllPhases() {
		if int(p) >= int(toPhase) {
			s.Phases[p] = &state.PhaseRecord{State: types.PhasePending}
		}
	}
	return o.SaveState(s)
}

// RecordDecision persists a user's answer to a requires-decision item.
func (o *Orchestrator) RecordDecision(d state.DecisionRecord) error {
	if err := state.AcquireLock(o.repoRoot, "cli"); err != nil {
		return err
	}
	defer state.ReleaseLock(o.repoRoot)

	s, err := o.LoadState()
	if err != nil {
		return err
	}
	d.DecidedAt = time.Now().UTC()
	s.Decisions = append(s.Decisions, d)
	return o.SaveState(s)
}

// Status returns the current state for inspection (used by the migrate_status
// MCP tool and `trabuco migrate status` CLI command).
func (o *Orchestrator) Status() (*state.State, error) {
	return o.LoadState()
}

// ---------- helpers ----------

func isNotApplicable(out *specialists.Output) bool {
	if len(out.Items) == 0 {
		return true
	}
	for _, item := range out.Items {
		if item.State != types.ItemNotApplicable {
			return false
		}
	}
	return true
}

func notApplicableReason(out *specialists.Output) string {
	for _, item := range out.Items {
		if item.State == types.ItemNotApplicable && item.Reason != "" {
			return item.Reason
		}
	}
	return out.Summary
}

func affectedModules(out *specialists.Output) []string {
	seen := make(map[string]struct{})
	var modules []string
	for _, item := range out.Items {
		if item.SourceEvidence == nil {
			continue
		}
		// First path segment is typically the module name in multi-module Maven.
		// "model/src/main/java/..." → "model"
		dir, _ := filepath.Split(item.SourceEvidence.File)
		if dir == "" {
			continue
		}
		parts := filepath.SplitList(filepath.Clean(dir))
		_ = parts // noop; resolution below
		// Use the first directory component.
		i := 0
		for i < len(item.SourceEvidence.File) && item.SourceEvidence.File[i] != '/' {
			i++
		}
		if i > 0 && i < len(item.SourceEvidence.File) {
			mod := item.SourceEvidence.File[:i]
			if _, dup := seen[mod]; !dup {
				seen[mod] = struct{}{}
				modules = append(modules, mod)
			}
		}
	}
	return modules
}

func writeJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func writeReport(path string, phase types.Phase, out *specialists.Output, res *validation.Result) error {
	body := fmt.Sprintf("# Phase %d — %s\n\n%s\n\n## Items\n\n", int(phase), phase, out.Summary)
	for _, item := range out.Items {
		body += fmt.Sprintf("- **[%s]** %s\n", item.State, item.Description)
		if item.BlockerCode != "" {
			body += fmt.Sprintf("  - blocker: `%s` — %s\n", item.BlockerCode, item.BlockerNote)
		}
	}
	body += fmt.Sprintf("\n## Validation\n\nPassed: %v (in %s)\n", res.Passed, res.Duration)
	return os.WriteFile(path, []byte(body), 0o644)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
