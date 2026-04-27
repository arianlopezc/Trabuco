package vcs

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// gitTestRepo initializes a fresh repo in t.TempDir() with one commit so
// HEAD exists. Returns the directory path.
func gitTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v (%s)", args, err, string(out))
		}
	}
	run("init", "-q")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "test")
	run("config", "commit.gpgsign", "false")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "README.md")
	run("commit", "-q", "-m", "initial")
	return dir
}

func TestPhaseTagNames(t *testing.T) {
	if got := PhasePreTag(types.PhaseModel); got != "trabuco-migration-phase-2-pre" {
		t.Errorf("pre tag = %q", got)
	}
	if got := PhasePostTag(types.PhaseFinalization); got != "trabuco-migration-phase-13-post" {
		t.Errorf("post tag = %q", got)
	}
}

func TestIsRepo(t *testing.T) {
	dir := gitTestRepo(t)
	if !IsRepo(dir) {
		t.Error("IsRepo on git repo should be true")
	}
	if IsRepo(t.TempDir()) {
		t.Error("IsRepo on plain dir should be false")
	}
}

func TestIsClean_FreshRepo(t *testing.T) {
	dir := gitTestRepo(t)
	clean, err := IsClean(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !clean {
		t.Error("fresh repo should be clean")
	}
}

func TestIsClean_DirtyTracked(t *testing.T) {
	dir := gitTestRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	clean, err := IsClean(dir)
	if err != nil {
		t.Fatal(err)
	}
	if clean {
		t.Error("repo with modified tracked file should not be clean")
	}
}

func TestIsClean_UntrackedIgnored(t *testing.T) {
	dir := gitTestRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "newfile.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	clean, err := IsClean(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !clean {
		t.Error("untracked files should not make IsClean false")
	}
}

func TestCurrentBranch(t *testing.T) {
	dir := gitTestRepo(t)
	branch, err := CurrentBranch(dir)
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if branch != "main" && branch != "master" {
		t.Errorf("branch = %q, want main or master", branch)
	}
}

func TestCreateTag_Exists_Reset(t *testing.T) {
	dir := gitTestRepo(t)
	tag := PhasePreTag(types.PhaseAssessment)

	if err := CreateTag(dir, tag, "test tag", false); err != nil {
		t.Fatalf("CreateTag: %v", err)
	}
	if !TagExists(dir, tag) {
		t.Fatalf("tag %s should exist after CreateTag", tag)
	}

	// Add a commit so we can rollback.
	if err := os.WriteFile(filepath.Join(dir, "next.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CommitAll(dir, "second commit"); err != nil {
		t.Fatal(err)
	}

	headBefore, err := HEAD(dir)
	if err != nil {
		t.Fatal(err)
	}

	// Reset back to the tag.
	if err := ResetHard(dir, tag); err != nil {
		t.Fatalf("ResetHard: %v", err)
	}
	headAfter, err := HEAD(dir)
	if err != nil {
		t.Fatal(err)
	}
	if headAfter == headBefore {
		t.Error("HEAD should have moved after reset")
	}
}

func TestCommitAll_DiffPatch(t *testing.T) {
	dir := gitTestRepo(t)
	from, _ := HEAD(dir)
	if err := os.WriteFile(filepath.Join(dir, "added.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := CommitAll(dir, "add file"); err != nil {
		t.Fatal(err)
	}
	to, _ := HEAD(dir)
	patch, err := DiffPatch(dir, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if patch == "" {
		t.Error("DiffPatch should return non-empty patch")
	}
}
