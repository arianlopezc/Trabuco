//go:build integration

package generator

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/templates"
)

// The Cursor + Codex adapters share 90% of their scaffolding with the Claude
// adapter (the test fixtures — git init, review-checks.sh, touchJava, etc.).
// The contracts are explicitly different:
//
//   Cursor:  {"followup_message": "..."}    auto-submits as next user turn
//            cycle cap via Cursor's native loop_count passed in stdin
//   Codex:   {"decision":"block","reason":"..."}    same schema as Claude
//            cycle cap via .trabuco/review-state/codex-<turn_id>.count
//
// Regressions in either direction (Codex emitting followup_message, Cursor
// emitting decision:block, either adapter forgetting kill switches) silently
// fail open in production. These tests are the contract.

// cursorDecision mirrors Cursor's Stop-hook response schema.
type cursorDecision struct {
	FollowupMessage string `json:"followup_message"`
}

// setupStopHook renders the requested Stop-hook adapter alongside the shared
// review-checks.sh and config into a git-initialized project dir. `adapter`
// selects which tool: "cursor" or "codex".
func setupStopHook(t *testing.T, adapter string, cfg *config.ProjectConfig) string {
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

	switch adapter {
	case "cursor":
		renderInto("cursor/hooks/require-review.sh.tmpl", ".cursor/hooks/require-review.sh")
	case "codex":
		renderInto("codex/hooks/require-review.sh.tmpl", ".codex/hooks/require-review.sh")
	default:
		t.Fatalf("unknown adapter: %s", adapter)
	}
	renderInto("github/scripts/review-checks.sh.tmpl", ".github/scripts/review-checks.sh")
	renderInto("trabuco/review.config.json.tmpl", ".trabuco/review.config.json")

	for _, mod := range cfg.Modules {
		for _, sub := range []string{"src/main/java", "src/test/java"} {
			_ = os.MkdirAll(filepath.Join(dir, mod, sub), 0o755)
		}
	}
	gitInit(t, dir)
	return dir
}

// runStopHook invokes the given adapter's script with stdin + env. Returns
// the raw exit/stdout/stderr; caller parses the tool-specific schema.
func runStopHook(t *testing.T, adapter, projectDir string, stdin any, env ...string) runResult {
	t.Helper()

	payload, err := json.Marshal(stdin)
	if err != nil {
		t.Fatalf("marshal stdin: %v", err)
	}

	var scriptPath string
	switch adapter {
	case "cursor":
		scriptPath = filepath.Join(projectDir, ".cursor", "hooks", "require-review.sh")
	case "codex":
		scriptPath = filepath.Join(projectDir, ".codex", "hooks", "require-review.sh")
	default:
		t.Fatalf("unknown adapter: %s", adapter)
	}

	cmd := exec.Command("bash", scriptPath)
	cmd.Dir = projectDir
	cmd.Stdin = strings.NewReader(string(payload))

	// Workspace env vars match what the respective tools pass in real use.
	baseEnv := os.Environ()
	switch adapter {
	case "cursor":
		baseEnv = append(baseEnv, "CURSOR_WORKSPACE_DIR="+projectDir)
	case "codex":
		baseEnv = append(baseEnv, "CODEX_WORKSPACE_DIR="+projectDir)
	}
	cmd.Env = append(baseEnv, env...)

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	exit := 0
	if ee, ok := err.(*exec.ExitError); ok {
		exit = ee.ExitCode()
	} else if err != nil {
		t.Fatalf("run %s hook: %v\nstdout: %s\nstderr: %s", adapter, err, stdout.String(), stderr.String())
	}
	return runResult{ExitCode: exit, Stdout: stdout.String(), Stderr: stderr.String()}
}

// ─── Cursor ──────────────────────────────────────────────────────────────────

// parseCursorDecision asserts the stdout is a valid Cursor followup_message
// response and returns the parsed struct. Fails the test on malformed output.
func parseCursorDecision(t *testing.T, res runResult) *cursorDecision {
	t.Helper()
	if strings.TrimSpace(res.Stdout) == "" {
		return nil
	}
	var d cursorDecision
	if err := json.Unmarshal([]byte(res.Stdout), &d); err != nil {
		t.Fatalf("cursor stdout is not a followup_message JSON: %v\nraw: %q", err, res.Stdout)
	}
	// Cursor's schema explicitly forbids Claude-style `decision`/`reason`
	// fields — check they aren't present.
	var any map[string]any
	_ = json.Unmarshal([]byte(res.Stdout), &any)
	if _, hasDecision := any["decision"]; hasDecision {
		t.Errorf("cursor output must not carry `decision` key (Claude schema leaked): %q", res.Stdout)
	}
	return &d
}

func TestCursorStopHook_KillSwitch_EnvVar(t *testing.T) {
	dir := setupStopHook(t, "cursor", projectCfg("Model", "API"))
	touchJava(t, dir, "API/src/main/java/Bad.java")

	res := runStopHook(t, "cursor", dir,
		map[string]any{"loop_count": 0},
		"TRABUCO_REVIEW_HOOK=off")
	if res.ExitCode != 0 || strings.TrimSpace(res.Stdout) != "" {
		t.Fatalf("kill switch should exit silently, got exit=%d stdout=%q",
			res.ExitCode, res.Stdout)
	}
}

func TestCursorStopHook_CleanDiff_EmptyResponse(t *testing.T) {
	dir := setupStopHook(t, "cursor", projectCfg("Model", "API"))
	// No java changes → hook exits silently so Cursor completes the turn.
	res := runStopHook(t, "cursor", dir, map[string]any{"loop_count": 0})
	if res.ExitCode != 0 || strings.TrimSpace(res.Stdout) != "" {
		t.Fatalf("clean diff should exit silently, got exit=%d stdout=%q",
			res.ExitCode, res.Stdout)
	}
}

func TestCursorStopHook_FindingsEmitFollowupMessage(t *testing.T) {
	dir := setupStopHook(t, "cursor", projectCfg("Model", "API"))
	writeJava(t, dir, "API/src/main/java/Bad.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Bad {
  @Autowired
  private Thing t;
}
`)
	res := runStopHook(t, "cursor", dir, map[string]any{"loop_count": 0})
	d := parseCursorDecision(t, res)
	if d == nil {
		t.Fatalf("expected followup_message response, got empty stdout\nstderr: %s", res.Stderr)
	}
	if !strings.Contains(d.FollowupMessage, "spring.field-injection") {
		t.Errorf("followup missing rule: %q", d.FollowupMessage)
	}
	if !strings.Contains(d.FollowupMessage, "cycle 1/3") {
		t.Errorf("followup missing cycle indicator: %q", d.FollowupMessage)
	}
}

func TestCursorStopHook_NativeLoopCountCap(t *testing.T) {
	// Cursor enforces loop_count natively — we honor it by emitting empty
	// stdout (advisory) once loop_count >= MAX so Cursor stops auto-looping.
	dir := setupStopHook(t, "cursor", projectCfg("Model", "API"))
	writeJava(t, dir, "API/src/main/java/Bad.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Bad {
  @Autowired
  private Thing t;
}
`)
	res := runStopHook(t, "cursor", dir, map[string]any{"loop_count": 3})
	if strings.TrimSpace(res.Stdout) != "" {
		t.Fatalf("at loop_count=MAX, hook should not emit followup (advisory only), got: %q",
			res.Stdout)
	}
	if !strings.Contains(res.Stderr, "Review cycle cap") {
		t.Errorf("advisory stderr missing cap notice: %q", res.Stderr)
	}
}

// ─── Codex ───────────────────────────────────────────────────────────────────

func TestCodexStopHook_KillSwitch_EnvVar(t *testing.T) {
	dir := setupStopHook(t, "codex", projectCfg("Model", "API"))
	touchJava(t, dir, "API/src/main/java/Bad.java")

	res := runStopHook(t, "codex", dir,
		map[string]any{"turn_id": "t1", "stop_hook_active": false},
		"TRABUCO_REVIEW_HOOK=off")
	if res.ExitCode != 0 || strings.TrimSpace(res.Stdout) != "" {
		t.Fatalf("kill switch should exit silently, got exit=%d stdout=%q",
			res.ExitCode, res.Stdout)
	}
}

func TestCodexStopHook_CleanDiff_Passes(t *testing.T) {
	dir := setupStopHook(t, "codex", projectCfg("Model", "API"))
	res := runStopHook(t, "codex", dir,
		map[string]any{"turn_id": "t2", "stop_hook_active": false})
	if res.ExitCode != 0 || strings.TrimSpace(res.Stdout) != "" {
		t.Fatalf("clean diff should pass, got exit=%d stdout=%q",
			res.ExitCode, res.Stdout)
	}
}

func TestCodexStopHook_FindingsEmitClaudeSchema(t *testing.T) {
	dir := setupStopHook(t, "codex", projectCfg("Model", "API"))
	writeJava(t, dir, "API/src/main/java/Bad.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Bad {
  @Autowired
  private Thing t;
}
`)
	res := runStopHook(t, "codex", dir,
		map[string]any{"turn_id": "t3", "stop_hook_active": false})

	if strings.TrimSpace(res.Stdout) == "" {
		t.Fatalf("expected block decision, got empty stdout\nstderr: %s", res.Stderr)
	}
	var d claudeDecision // same schema as Claude
	if err := json.Unmarshal([]byte(res.Stdout), &d); err != nil {
		t.Fatalf("codex stdout is not a decision/reason JSON: %v\nraw: %q", err, res.Stdout)
	}
	if d.Decision != "block" {
		t.Errorf("decision = %q, want block", d.Decision)
	}
	if !strings.Contains(d.Reason, "spring.field-injection") {
		t.Errorf("reason missing rule: %q", d.Reason)
	}
	if !strings.Contains(d.Reason, "cycle 1/3") {
		t.Errorf("reason missing cycle indicator: %q", d.Reason)
	}
	// Codex must not accidentally use Cursor's schema.
	var raw map[string]any
	_ = json.Unmarshal([]byte(res.Stdout), &raw)
	if _, has := raw["followup_message"]; has {
		t.Errorf("codex output must not carry `followup_message` key (cursor schema leaked): %q", res.Stdout)
	}
}

func TestCodexStopHook_CycleCap_Advisory(t *testing.T) {
	dir := setupStopHook(t, "codex", projectCfg("Model", "API"))
	writeJava(t, dir, "API/src/main/java/Bad.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class Bad {
  @Autowired
  private Thing t;
}
`)
	// Seed the codex-specific counter file to the cap.
	stateDir := filepath.Join(dir, ".trabuco", "review-state")
	_ = os.MkdirAll(stateDir, 0o755)
	if err := os.WriteFile(filepath.Join(stateDir, "codex-t4.count"), []byte("3"), 0o644); err != nil {
		t.Fatalf("seed counter: %v", err)
	}

	res := runStopHook(t, "codex", dir,
		map[string]any{"turn_id": "t4", "stop_hook_active": false})
	if res.ExitCode != 0 {
		t.Fatalf("cap should exit 0 (advisory), got exit=%d", res.ExitCode)
	}
	if strings.TrimSpace(res.Stdout) != "" {
		t.Fatalf("cap should suppress block, got stdout: %q", res.Stdout)
	}
	if !strings.Contains(res.Stderr, "Cycle cap reached") {
		t.Errorf("advisory stderr missing cap notice: %q", res.Stderr)
	}
}
