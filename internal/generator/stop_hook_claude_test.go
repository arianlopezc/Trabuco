//go:build integration

package generator

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/templates"
)

// Claude Stop-hook adapter is a pure stdin → stdout + exit code function.
// Tests drive it with synthetic input and assert the JSON decision schema
// plus the cycle-counter side effects. If the adapter ever regresses its
// block schema (e.g., renames `decision` → `block`), every consuming agent
// silently fails open, so the contract-level tests here matter more than
// any other coverage in this package.

// claudeDecision is the JSON schema the Claude Code Stop hook emits.
type claudeDecision struct {
	Decision string `json:"decision"`
	Reason   string `json:"reason"`
}

// runResult captures what the hook did on one invocation.
type runResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Decision *claudeDecision // nil when stdout is empty (clean exit)
}

// setupClaudeHook renders the Claude Stop-hook adapter + review-checks.sh into
// a git-initialized tmp dir, returns the project path. Fixtures then modify
// the working tree and call runClaudeHook() to exercise one Stop event.
func setupClaudeHook(t *testing.T, cfg *config.ProjectConfig) string {
	t.Helper()
	dir := t.TempDir()
	engine := templates.NewEngine()

	renderInto := func(tmplPath, outRelPath string) {
		rendered, err := engine.Execute(tmplPath, cfg)
		if err != nil {
			t.Fatalf("render %s: %v", tmplPath, err)
		}
		full := filepath.Join(dir, outRelPath)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(rendered), 0o755); err != nil {
			t.Fatalf("write %s: %v", full, err)
		}
	}

	renderInto("claude/hooks/require-review.sh.tmpl", ".claude/hooks/require-review.sh")
	renderInto("github/scripts/review-checks.sh.tmpl", ".github/scripts/review-checks.sh")
	renderInto("trabuco/review.config.json.tmpl", ".trabuco/review.config.json")

	// Module source dirs — same pattern as review fixture harness.
	for _, mod := range cfg.Modules {
		for _, sub := range []string{"src/main/java", "src/test/java"} {
			_ = os.MkdirAll(filepath.Join(dir, mod, sub), 0o755)
		}
	}

	gitInit(t, dir)
	return dir
}

// runClaudeHook invokes .claude/hooks/require-review.sh with the given stdin
// payload and environment. Env pairs are "KEY=VAL" strings appended to the
// process environment. CLAUDE_PROJECT_DIR is set automatically.
func runClaudeHook(t *testing.T, projectDir string, stdin any, env ...string) runResult {
	t.Helper()

	payload, err := json.Marshal(stdin)
	if err != nil {
		t.Fatalf("marshal stdin: %v", err)
	}

	cmd := exec.Command("bash", filepath.Join(projectDir, ".claude", "hooks", "require-review.sh"))
	cmd.Dir = projectDir
	cmd.Stdin = strings.NewReader(string(payload))
	cmd.Env = append(os.Environ(),
		"CLAUDE_PROJECT_DIR="+projectDir,
	)
	cmd.Env = append(cmd.Env, env...)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	exit := 0
	if ee, ok := err.(*exec.ExitError); ok {
		exit = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run hook: %v\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	res := runResult{
		ExitCode: exit,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
	// The hook uses `set -u` and `exit 0` for kill-switch paths, producing
	// empty stdout. Only attempt to parse JSON when stdout is non-empty.
	if strings.TrimSpace(res.Stdout) != "" {
		var d claudeDecision
		if err := json.Unmarshal([]byte(res.Stdout), &d); err != nil {
			t.Fatalf("stdout is not a claude decision JSON: %v\nraw: %q", err, res.Stdout)
		}
		res.Decision = &d
	}
	return res
}

// touchJava makes a Java file appear in the working tree so the hook sees a
// "code changed" state. intent-to-add only — no commit — matches what an
// editing agent's in-session working tree looks like.
func touchJava(t *testing.T, dir, rel string) {
	t.Helper()
	writeJava(t, dir, rel, `package x; class C {}`)
}

// writeTranscript writes a minimal Claude transcript file containing (or not)
// a `code-reviewer` subagent marker. Returns the absolute path.
func writeTranscript(t *testing.T, dir string, reviewerInvoked bool) string {
	t.Helper()
	path := filepath.Join(dir, "transcript.jsonl")
	var body string
	if reviewerInvoked {
		// The hook greps for either `"subagent_type":"code-reviewer"` or
		// `"name":"code-reviewer"`. Use the subagent_type variant — that's
		// what Task tool invocations look like in real transcripts.
		body = `{"type":"tool_use","subagent_type":"code-reviewer"}` + "\n"
	} else {
		body = `{"type":"text","content":"just a chat message"}` + "\n"
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write transcript: %v", err)
	}
	return path
}

// stopInput composes a realistic Claude Stop-event stdin payload.
func stopInput(sessionID, transcriptPath string, stopActive bool) map[string]any {
	return map[string]any{
		"stop_hook_active": stopActive,
		"session_id":       sessionID,
		"transcript_path":  transcriptPath,
	}
}

// readCycleCount returns the current cycle counter for the given session
// (0 if the file doesn't exist). The cycle counter is how the adapter
// enforces the 3-iteration cap.
func readCycleCount(t *testing.T, dir, sessionID string) int {
	t.Helper()
	path := filepath.Join(dir, ".trabuco", "review-state", sessionID+".count")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return 0
	}
	if err != nil {
		t.Fatalf("read cycle file %s: %v", path, err)
	}
	var n int
	_, _ = fmt.Sscanf(strings.TrimSpace(string(data)), "%d", &n)
	return n
}

// ─── Kill-switch branches ────────────────────────────────────────────────────

func TestClaudeStopHook_KillSwitch_EnvVar(t *testing.T) {
	dir := setupClaudeHook(t, projectCfg("Model", "API"))
	touchJava(t, dir, "API/src/main/java/Bad.java") // would trigger if hook ran
	tr := writeTranscript(t, dir, false)            // reviewer NOT invoked

	res := runClaudeHook(t, dir, stopInput("s1", tr, false), "TRABUCO_REVIEW_HOOK=off")

	if res.ExitCode != 0 {
		t.Fatalf("expected exit 0 with kill switch, got %d\nstderr: %s", res.ExitCode, res.Stderr)
	}
	if res.Decision != nil {
		t.Fatalf("expected no decision output, got: %+v", res.Decision)
	}
}

func TestClaudeStopHook_KillSwitch_ConfigDisabled(t *testing.T) {
	dir := setupClaudeHook(t, projectCfg("Model", "API"))
	// Flip enabled=false in the runtime config.
	cfgPath := filepath.Join(dir, ".trabuco", "review.config.json")
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("parse config: %v", err)
	}
	m["enabled"] = false
	out, _ := json.Marshal(m)
	if err := os.WriteFile(cfgPath, out, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	touchJava(t, dir, "API/src/main/java/Bad.java")
	tr := writeTranscript(t, dir, false)

	res := runClaudeHook(t, dir, stopInput("s2", tr, false))
	if res.ExitCode != 0 || res.Decision != nil {
		t.Fatalf("disabled config should pass silently, got exit=%d decision=%+v",
			res.ExitCode, res.Decision)
	}
}

// ─── No-op branches (nothing to review) ──────────────────────────────────────

func TestClaudeStopHook_CleanWorkingTree_PassesSilently(t *testing.T) {
	dir := setupClaudeHook(t, projectCfg("Model", "API"))
	// No touchJava call — working tree is clean. Hook should short-circuit.
	tr := writeTranscript(t, dir, false)

	res := runClaudeHook(t, dir, stopInput("s3", tr, false))
	if res.ExitCode != 0 || res.Decision != nil {
		t.Fatalf("clean tree should pass, got exit=%d decision=%+v",
			res.ExitCode, res.Decision)
	}
}

// ─── Layer 1: subagent-not-invoked block ─────────────────────────────────────

func TestClaudeStopHook_Layer1_SubagentNotInvoked_Blocks(t *testing.T) {
	dir := setupClaudeHook(t, projectCfg("Model", "API"))
	touchJava(t, dir, "API/src/main/java/Edit.java")
	tr := writeTranscript(t, dir, false) // reviewer NOT invoked

	res := runClaudeHook(t, dir, stopInput("s4", tr, false))

	if res.Decision == nil {
		t.Fatalf("expected block decision, got exit=%d stdout=%q",
			res.ExitCode, res.Stdout)
	}
	if res.Decision.Decision != "block" {
		t.Errorf("decision = %q, want block", res.Decision.Decision)
	}
	if !strings.Contains(res.Decision.Reason, "code-reviewer") {
		t.Errorf("reason missing code-reviewer mention: %q", res.Decision.Reason)
	}
	if got := readCycleCount(t, dir, "s4"); got != 1 {
		t.Errorf("cycle counter = %d, want 1 after first block", got)
	}
}

// ─── Layer 1: subagent invoked passes into Layer 2 ──────────────────────────

func TestClaudeStopHook_Layer1_SubagentInvoked_RunsLayer2(t *testing.T) {
	dir := setupClaudeHook(t, projectCfg("Model", "API"))
	// Write a fixture that would NOT trigger any rule in review-checks.sh —
	// so Layer 1 passes and Layer 2 finds nothing → clean exit.
	writeJava(t, dir, "API/src/main/java/Clean.java",
		`package x; public class Clean { private final int v = 1; public int v() { return v; } }`)
	tr := writeTranscript(t, dir, true) // reviewer invoked

	res := runClaudeHook(t, dir, stopInput("s5", tr, false))
	if res.ExitCode != 0 || res.Decision != nil {
		t.Fatalf("reviewer-seen + clean code should pass, got exit=%d decision=%+v",
			res.ExitCode, res.Decision)
	}
}

// ─── Layer 2: deterministic findings block ──────────────────────────────────

func TestClaudeStopHook_Layer2_FindingsBlockWithCycleCounter(t *testing.T) {
	dir := setupClaudeHook(t, projectCfg("Model", "API"))
	// Field injection triggers spring.field-injection.
	writeJava(t, dir, "API/src/main/java/Bad.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Bad {
  @Autowired
  private Thing t;
}
`)
	tr := writeTranscript(t, dir, true) // reviewer IS invoked (skips Layer 1)

	res := runClaudeHook(t, dir, stopInput("s6", tr, false))
	if res.Decision == nil {
		t.Fatalf("expected block decision, got exit=%d stdout=%q stderr=%q",
			res.ExitCode, res.Stdout, res.Stderr)
	}
	if res.Decision.Decision != "block" {
		t.Errorf("decision = %q, want block", res.Decision.Decision)
	}
	// Reason should embed the finding line + a cycle indicator.
	if !strings.Contains(res.Decision.Reason, "spring.field-injection") {
		t.Errorf("reason missing rule id: %q", res.Decision.Reason)
	}
	if !strings.Contains(res.Decision.Reason, "cycle 1/3") {
		t.Errorf("reason missing cycle counter: %q", res.Decision.Reason)
	}
	if got := readCycleCount(t, dir, "s6"); got != 1 {
		t.Errorf("cycle counter = %d, want 1", got)
	}
}

// ─── Cycle cap: advisory mode after 3 cycles ─────────────────────────────────

func TestClaudeStopHook_CycleCap_SwitchesToAdvisory(t *testing.T) {
	dir := setupClaudeHook(t, projectCfg("Model", "API"))
	writeJava(t, dir, "API/src/main/java/Bad.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Bad {
  @Autowired
  private Thing t;
}
`)
	tr := writeTranscript(t, dir, true)

	// Pre-load the cycle counter to 3 — next invocation should advisory-out.
	stateDir := filepath.Join(dir, ".trabuco", "review-state")
	_ = os.MkdirAll(stateDir, 0o755)
	countFile := filepath.Join(stateDir, "s7.count")
	if err := os.WriteFile(countFile, []byte("3"), 0o644); err != nil {
		t.Fatalf("seed cycle file: %v", err)
	}

	res := runClaudeHook(t, dir, stopInput("s7", tr, false))
	if res.ExitCode != 0 {
		t.Fatalf("cap should switch to advisory (exit 0), got exit=%d", res.ExitCode)
	}
	if res.Decision != nil {
		t.Fatalf("cap should not emit a block decision, got: %+v", res.Decision)
	}
	if !strings.Contains(res.Stderr, "Cycle cap reached") {
		t.Errorf("expected 'Cycle cap reached' in stderr, got: %q", res.Stderr)
	}
	if !strings.Contains(res.Stderr, "spring.field-injection") {
		t.Errorf("advisory stderr missing rule: %q", res.Stderr)
	}
}
