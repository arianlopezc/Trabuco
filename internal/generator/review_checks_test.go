//go:build integration

package generator

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/templates"
)

// These tests exercise the generated `.github/scripts/review-checks.sh`
// against fixture Java sources. The script is the single source of
// deterministic review truth — it's what the Stop-hook adapters (Claude,
// Cursor, Codex) execute to decide whether to block. If a rule is silently
// broken in the template, every consumer of it lies about code quality.
//
// Run with: go test -tags=integration -run TestReviewChecks ./internal/generator/...

// reviewFinding is one parsed `::error file=X,line=Y,title=[rule] ...::snippet`
// annotation emitted by review-checks.sh.
type reviewFinding struct {
	Rule    string
	File    string // path relative to the project root
	Line    int
	Snippet string
}

// reviewAnnotation matches a GitHub Actions error annotation the way
// review-checks.sh emits them. The title field is `[rule-id] Human-readable`
// and the body after the final `::` is the snippet (usually the offending
// line). Non-greedy `.*?` is required because titles legitimately contain
// single colons (e.g., "project policy: indexed columns only").
var reviewAnnotation = regexp.MustCompile(
	`^::error file=([^,]+),line=(\d+),title=\[([^\]]+)\].*?::(.*)$`,
)

// setupReviewProject writes a rendered review-checks.sh and the directory
// skeleton for each module. Returns the project root. Caller drops fixture
// Java files into the returned path using writeJava().
func setupReviewProject(t *testing.T, cfg *config.ProjectConfig) string {
	t.Helper()

	projectDir := t.TempDir()

	// Render the review-checks.sh template with the given config.
	engine := templates.NewEngine()
	rendered, err := engine.Execute("github/scripts/review-checks.sh.tmpl", cfg)
	if err != nil {
		t.Fatalf("render review-checks.sh: %v", err)
	}
	scriptPath := filepath.Join(projectDir, ".github", "scripts", "review-checks.sh")
	if err := os.MkdirAll(filepath.Dir(scriptPath), 0o755); err != nil {
		t.Fatalf("mkdir scripts: %v", err)
	}
	if err := os.WriteFile(scriptPath, []byte(rendered), 0o755); err != nil {
		t.Fatalf("write review-checks.sh: %v", err)
	}

	// Make module source directories so fixture writes don't have to MkdirAll
	// each time. The checks key off `/src/main/java/` and `/src/test/java/`
	// path substrings.
	for _, mod := range cfg.Modules {
		for _, sub := range []string{"src/main/java", "src/test/java"} {
			_ = os.MkdirAll(filepath.Join(projectDir, mod, sub), 0o755)
		}
	}
	// SQL migration dir for SQLDatastore.
	if cfg.HasModule("SQLDatastore") {
		_ = os.MkdirAll(
			filepath.Join(projectDir, "SQLDatastore", "src", "main", "resources", "db", "migration"),
			0o755,
		)
	}

	// Initialize a git repo so `git ls-files` inside the script has something
	// to enumerate. Without this the script falls back to `find`, which is
	// slower and has slightly different semantics.
	gitInit(t, projectDir)

	return projectDir
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"add", "-A"},
		{"commit", "-q", "--allow-empty", "-m", "init"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, out)
		}
	}
}

// writeJava writes a Java source file at relPath under projectDir and stages
// it so git ls-files / git diff scopes see it. `add -N` (intent-to-add) keeps
// the file visible to `git diff --name-only HEAD` for --scope=local without
// requiring a new commit per fixture.
func writeJava(t *testing.T, projectDir, relPath, content string) {
	t.Helper()
	full := filepath.Join(projectDir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatalf("mkdir for %s: %v", relPath, err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", relPath, err)
	}
	cmd := exec.Command("git", "add", "-N", relPath)
	cmd.Dir = projectDir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add -N %s: %v\n%s", relPath, err, out)
	}
}

// writeSQL is the migration-file equivalent of writeJava; added separately to
// make intent at the call-site obvious.
func writeSQL(t *testing.T, projectDir, relPath, content string) {
	t.Helper()
	writeJava(t, projectDir, relPath, content) // same mechanics, different caller semantics
}

// runReviewChecks runs the generated script with --scope=<scope> against the
// given project dir and returns (findings, rawOutput). findings are parsed
// from the `::error` annotations the script emits.
func runReviewChecks(t *testing.T, projectDir, scope string) ([]reviewFinding, string) {
	t.Helper()
	scriptPath := filepath.Join(projectDir, ".github", "scripts", "review-checks.sh")
	args := []string{scriptPath}
	if scope != "" {
		args = append(args, "--scope="+scope)
	}
	cmd := exec.Command("bash", args...)
	cmd.Dir = projectDir
	out, err := cmd.CombinedOutput()
	// Exit status 1 means findings; 0 means clean. Other codes (2, segfault, …)
	// mean the script itself is broken — fail the test with the raw output so
	// the regression is obvious.
	if exitErr, ok := err.(*exec.ExitError); ok {
		code := exitErr.ExitCode()
		if code != 0 && code != 1 {
			t.Fatalf("review-checks.sh exited with %d:\n%s", code, out)
		}
	} else if err != nil {
		t.Fatalf("review-checks.sh failed to execute: %v\n%s", err, out)
	}

	var findings []reviewFinding
	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	scanner.Buffer(make([]byte, 1<<20), 1<<20)
	for scanner.Scan() {
		m := reviewAnnotation.FindStringSubmatch(scanner.Text())
		if m == nil {
			continue
		}
		var line int
		_, _ = fmt.Sscanf(m[2], "%d", &line)
		// Normalize file path: strip leading "./" the script prepends in some
		// fallback paths, and make relative to projectDir if absolute.
		file := strings.TrimPrefix(m[1], projectDir+string(filepath.Separator))
		file = strings.TrimPrefix(file, "./")
		findings = append(findings, reviewFinding{
			Rule:    m[3],
			File:    file,
			Line:    line,
			Snippet: m[4],
		})
	}
	return findings, string(out)
}

// assertFinding fails unless there is exactly one finding with the given rule
// in the given file. File match is suffix-based so tests don't couple to the
// concrete tmp dir layout.
func assertFinding(t *testing.T, findings []reviewFinding, rule, fileSuffix string) {
	t.Helper()
	var matched []reviewFinding
	for _, f := range findings {
		if f.Rule == rule && strings.HasSuffix(f.File, fileSuffix) {
			matched = append(matched, f)
		}
	}
	if len(matched) == 0 {
		t.Fatalf("expected rule %q in file %q, got findings: %+v", rule, fileSuffix, findings)
	}
	if len(matched) > 1 {
		t.Fatalf("expected exactly one %q finding in %q, got %d: %+v",
			rule, fileSuffix, len(matched), matched)
	}
}

// assertNoFinding fails if any finding with the given rule exists against any
// file whose path ends with fileSuffix. Use for negative/suppression cases.
func assertNoFinding(t *testing.T, findings []reviewFinding, rule, fileSuffix string) {
	t.Helper()
	for _, f := range findings {
		if f.Rule == rule && strings.HasSuffix(f.File, fileSuffix) {
			t.Fatalf("expected no %q finding in %q, but got: %+v", rule, fileSuffix, f)
		}
	}
}

// projectCfg is a shorthand for the common fixture configs.
func projectCfg(modules ...string) *config.ProjectConfig {
	return &config.ProjectConfig{
		ProjectName: "fixture",
		GroupID:     "com.trabuco.fixture",
		ArtifactID:  "fixture",
		JavaVersion: "21",
		Modules:     modules,
		Database:    "postgresql",
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
	}
}

// TestReviewChecks_Harness_Sanity is a smoke test that proves the harness
// itself works: render script, write a clean file, run, expect zero findings.
func TestReviewChecks_Harness_Sanity(t *testing.T) {
	dir := setupReviewProject(t, projectCfg("Model", "API"))
	writeJava(t, dir, "API/src/main/java/com/trabuco/fixture/Clean.java", `
package com.trabuco.fixture;

public class Clean {
  private final int value = 42;
  public int value() { return value; }
}
`)
	findings, out := runReviewChecks(t, dir, "all")
	if len(findings) != 0 {
		t.Fatalf("clean fixture produced findings:\n%s\n%+v", out, findings)
	}
}
