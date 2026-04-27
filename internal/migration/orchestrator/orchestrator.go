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
	"strings"
	"time"

	"github.com/arianlopezc/Trabuco/internal/java"
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

	// Preflight: build runtime must match the project's target Java
	// version. mvn will use whatever java is on PATH; if it doesn't
	// match the target, plugin failures look mysterious (ArchUnit
	// rejecting class file major versions, Spotless misbehaving, etc.).
	// Skip on Phase 0 because targetConfig.javaVersion isn't set yet.
	if err := preflightRuntimeJava(s); err != nil {
		return "", err
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

	// Compose UserHint: explicit edit-and-approve hint takes priority,
	// but always append pending decisions for this phase so the
	// specialist can apply user choices on a re-run after `migrate
	// decision`. Without this the LLM has to dig through state.json's
	// decisions array on its own — error-prone.
	userHint := hint
	if dh := pendingDecisionHint(s, phase); dh != "" {
		if userHint != "" {
			userHint = userHint + "\n\n" + dh
		} else {
			userHint = dh
		}
	}

	// Invoke the specialist.
	in := &specialists.Input{
		RepoRoot: o.repoRoot,
		Phase:    phase,
		State:    s,
		UserHint: userHint,
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
	// during migration phases. We skip the funnel when no item declared
	// any file_writes — Phase 0 (assessor) produces only the assessment
	// catalog, and a phase whose items are all blocked / requires_decision
	// hasn't changed code, so there's nothing to compile-check.
	// Phase 12 (activation) runs through regardless because the activator
	// itself rewrites the parent POM in-process; its file changes happen
	// on disk before this point.
	res := validation.Result{Passed: true}
	if phase == types.PhaseActivation || hasFileWrites(out) {
		mode := validation.ModeMigration
		if phase == types.PhaseActivation {
			mode = validation.ModeActivation
		}
		res = validation.Run(o.repoRoot, mode, affectedModules(o.repoRoot, out))
	}
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
// It scans every phase-N-output.json for an item whose ID matches d.ID
// so it can backfill Phase / Question / Choices from the originating
// requires_decision item — the user only needs to supply id + choice on
// the CLI. If a decision with the same ID already exists it's replaced
// (re-running `migrate decision` should idempotently update, not duplicate).
func (o *Orchestrator) RecordDecision(d state.DecisionRecord) error {
	if err := state.AcquireLock(o.repoRoot, "cli"); err != nil {
		return err
	}
	defer state.ReleaseLock(o.repoRoot)

	s, err := o.LoadState()
	if err != nil {
		return err
	}
	if d.Phase == 0 || d.Question == "" {
		if found := lookupDecisionContext(o.repoRoot, d.ID); found != nil {
			if d.Phase == 0 {
				d.Phase = found.Phase
			}
			if d.Question == "" {
				d.Question = found.Question
			}
			if len(d.Choices) == 0 {
				d.Choices = found.Choices
			}
		}
	}
	d.DecidedAt = time.Now().UTC()

	// Replace existing record with the same ID; otherwise append.
	replaced := false
	for i, prev := range s.Decisions {
		if prev.ID == d.ID {
			s.Decisions[i] = d
			replaced = true
			break
		}
	}
	if !replaced {
		s.Decisions = append(s.Decisions, d)
	}
	return o.SaveState(s)
}

// lookupDecisionContext searches every phase-N-output.json under
// .trabuco-migration/ for an item with the given ID and returns its
// originating phase + question + choices. Returns nil if not found.
func lookupDecisionContext(repoRoot, id string) *state.DecisionRecord {
	for _, p := range types.AllPhases() {
		path := state.PhaseOutputPath(repoRoot, p)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var out specialists.Output
		if err := json.Unmarshal(data, &out); err != nil {
			continue
		}
		for _, item := range out.Items {
			if item.ID != id {
				continue
			}
			return &state.DecisionRecord{
				Phase:    p,
				Question: item.Question,
				Choices:  item.Choices,
			}
		}
		for _, d := range out.Decisions {
			if d.ID == id {
				return &state.DecisionRecord{
					Phase:    p,
					Question: d.Question,
					Choices:  d.Choices,
				}
			}
		}
	}
	return nil
}

// pendingDecisionHint composes a UserHint string from every recorded
// decision attached to the given phase. Returns "" if no decisions
// match the phase. This is what RunPhase passes into the specialist
// when it re-attempts a previously-blocked or requires_decision phase.
func pendingDecisionHint(s *state.State, phase types.Phase) string {
	var lines []string
	for _, d := range s.Decisions {
		if d.Phase != phase || d.Choice == "" {
			continue
		}
		line := fmt.Sprintf("- decision %q: user chose %q", d.ID, d.Choice)
		if d.Question != "" {
			line += fmt.Sprintf(" (question: %s)", d.Question)
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		return ""
	}
	return "The user has recorded decisions for previous requires_decision items in this phase. Apply each choice in your next output:\n" + strings.Join(lines, "\n")
}

// Status returns the current state for inspection (used by the migrate_status
// MCP tool and `trabuco migrate status` CLI command).
func (o *Orchestrator) Status() (*state.State, error) {
	return o.LoadState()
}

// ---------- helpers ----------

// preflightRuntimeJava checks that `java` on PATH matches the project's
// target Java version. Returns nil when targetConfig.javaVersion is
// unset (Phase 0 hasn't run yet) or when the major versions match;
// returns a JAVA_VERSION_MISMATCH_RUNTIME-tagged error otherwise. The
// error message lists the standard remediations (JAVA_HOME, install
// matching JDK, Maven Toolchains).
func preflightRuntimeJava(s *state.State) error {
	want := s.TargetConfig.JavaVersion
	if want == "" {
		return nil
	}
	wantMajor, _, err := java.ParseVersion(want)
	if err != nil || wantMajor == 0 {
		// Target value is unparseable; don't gate on a malformed config.
		return nil
	}
	gotMajor, gotFull, err := java.RuntimeJavaMajor()
	if err != nil {
		return fmt.Errorf("[%s] cannot probe java runtime: %w — install a JDK %d that matches the project's target, or set JAVA_HOME",
			types.BlockerJavaVersionMismatchRuntime, err, wantMajor)
	}
	if gotMajor == wantMajor {
		return nil
	}
	return fmt.Errorf(
		"[%s] build runtime mismatch: java -version reports %d (%q) but the project's target is Java %d.\n"+
			"  mvn uses the JDK on PATH, and ArchUnit/Spotless/Enforcer plugins can fail mysteriously when run on a JDK newer than the target.\n"+
			"  Fix any of:\n"+
			"    - install JDK %d and set JAVA_HOME to point at it (recommended for local dev)\n"+
			"    - configure Maven Toolchains so the build always uses JDK %d regardless of PATH\n"+
			"    - run trabuco from a container/CI image pinned to JDK %d (matches what the generated CI workflow does)",
		types.BlockerJavaVersionMismatchRuntime, gotMajor, gotFull, wantMajor, wantMajor, wantMajor, wantMajor,
	)
}

// hasFileWrites reports whether any item declared file_writes — used
// to decide whether the validation funnel needs to run.
func hasFileWrites(out *specialists.Output) bool {
	for _, item := range out.Items {
		if len(item.FileWrites) > 0 {
			return true
		}
	}
	return false
}

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

// affectedModules returns the first-directory-segment of every path the
// specialist touched (both source_evidence and file_writes), filtered to
// only those segments that are actually Maven modules (have a pom.xml).
// This is the list the validation funnel scopes its compile/test runs
// to. We MUST include file_writes paths as well: a phase that creates
// new files in `model/` without source_evidence in `model/` would
// otherwise have its new module skipped by `mvn -pl legacy -am`. And we
// MUST filter to real modules: a deployment phase that writes
// `.github/workflows/ci.yml` produces a path-segment of `.github`,
// which Maven would reject as a module ("Could not find the selected
// project in the reactor: :.github").
func affectedModules(repoRoot string, out *specialists.Output) []string {
	seen := make(map[string]struct{})
	var modules []string
	add := func(p string) {
		i := 0
		for i < len(p) && p[i] != '/' {
			i++
		}
		if i == 0 || i >= len(p) {
			return
		}
		mod := p[:i]
		if _, dup := seen[mod]; dup {
			return
		}
		// Real-module gate: must have a pom.xml under repoRoot/mod/.
		if _, err := os.Stat(filepath.Join(repoRoot, mod, "pom.xml")); err != nil {
			return
		}
		seen[mod] = struct{}{}
		modules = append(modules, mod)
	}
	for _, item := range out.Items {
		if item.SourceEvidence != nil {
			add(item.SourceEvidence.File)
		}
		for _, w := range item.FileWrites {
			add(w.Path)
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
