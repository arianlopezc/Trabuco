// Package vcs wraps git operations the migration needs: per-phase pre/post
// tags for atomic rollback, patch generation, and pre-flight cleanliness
// checks. All work is done via the `git` binary on PATH; no third-party Go
// libraries — keeps the dependency surface tight and behavior predictable.
package vcs

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// TagPrefix is the prefix used for all migration phase tags. Tag format is
// `trabuco-migration-phase-{N}-pre` and `trabuco-migration-phase-{N}-post`.
const TagPrefix = "trabuco-migration-phase"

// PhasePreTag returns the canonical "pre" tag name for a phase.
func PhasePreTag(p types.Phase) string {
	return fmt.Sprintf("%s-%d-pre", TagPrefix, int(p))
}

// PhasePostTag returns the canonical "post" tag name for a phase.
func PhasePostTag(p types.Phase) string {
	return fmt.Sprintf("%s-%d-post", TagPrefix, int(p))
}

// IsRepo reports whether dir is a git repository.
func IsRepo(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// IsClean reports whether the working tree is clean (no uncommitted changes
// to tracked files; untracked files don't count).
func IsClean(dir string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain", "--untracked-files=no")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return len(strings.TrimSpace(string(out))) == 0, nil
}

// CurrentBranch returns the current branch name. Returns an error if HEAD
// is detached.
func CurrentBranch(dir string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--short", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("detached HEAD or not on a branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// HasCommits reports whether the repo has at least one commit.
func HasCommits(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "HEAD")
	cmd.Dir = dir
	return cmd.Run() == nil
}

// CreateTag creates an annotated tag pointing at HEAD. If force is true,
// an existing tag with the same name is overwritten.
func CreateTag(dir, name, message string, force bool) error {
	args := []string{"tag", "-a", name, "-m", message}
	if force {
		args = append(args, "-f")
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git tag %s: %w (%s)", name, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// TagExists reports whether a tag with the given name exists.
func TagExists(dir, name string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", "refs/tags/"+name)
	cmd.Dir = dir
	return cmd.Run() == nil
}

// ResetHard resets the working tree to the given ref. This is destructive —
// only the orchestrator's rollback path should call it, after the user
// approves rollback at a phase gate.
func ResetHard(dir, ref string) error {
	cmd := exec.Command("git", "reset", "--hard", ref)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git reset --hard %s: %w (%s)", ref, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// CommitAll stages everything (including untracked files) and creates a
// commit. The orchestrator commits at phase boundaries; specialists never
// commit directly.
func CommitAll(dir, message string) error {
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = dir
	if out, err := addCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("git add: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	commitCmd := exec.Command("git", "commit", "-m", message, "--allow-empty")
	commitCmd.Dir = dir
	out, err := commitCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git commit: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// DiffPatch returns a unified-diff patch between two refs. Used to write
// phase-{N}-diff.patch.
func DiffPatch(dir, fromRef, toRef string) (string, error) {
	cmd := exec.Command("git", "diff", fromRef+".."+toRef)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git diff: %w", err)
	}
	return string(out), nil
}

// HEAD returns the current HEAD commit SHA.
func HEAD(dir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
