//go:build integration

package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Round-trip tests prove the full "agent edits → Stop hook blocks → agent
// fixes → Stop hook passes" loop actually works, not just individual branches.
// This is what Phase 4 of the e2e plan asserts: the thing the user cares
// about — "the review system actually reviews and unblocks" — behaves
// correctly, deterministically, per-tool.
//
// The agent is simulated by direct file edits. LLM-in-the-loop would be
// non-deterministic and expensive; the edits we apply are the exact edits
// a correctly-operating reviewer-guided agent would make.

// revertEdit overwrites a file the test previously created via writeJava.
// Hooks consume working-tree diff, so a fresh write makes them see the new
// state on the next invocation. The intent-to-add from the original
// writeJava is already staged.
func revertEdit(t *testing.T, projectDir, rel, newContent string) {
	t.Helper()
	full := filepath.Join(projectDir, rel)
	if err := os.WriteFile(full, []byte(newContent), 0o644); err != nil {
		t.Fatalf("overwrite %s: %v", rel, err)
	}
}

// ─── Claude round-trip ──────────────────────────────────────────────────────

func TestStopHook_RoundTrip_Claude(t *testing.T) {
	dir := setupClaudeHook(t, projectCfg("Model", "API"))
	tr := writeTranscript(t, dir, true) // reviewer invoked — skip Layer 1

	// Step 1: inject a violation.
	bad := `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class RoundTrip {
  @Autowired
  private Thing t;
}
`
	writeJava(t, dir, "API/src/main/java/RoundTrip.java", bad)

	res1 := runClaudeHook(t, dir, stopInput("rt-claude", tr, false))
	if res1.Decision == nil || res1.Decision.Decision != "block" {
		t.Fatalf("step 1: expected block, got exit=%d decision=%+v",
			res1.ExitCode, res1.Decision)
	}
	if !strings.Contains(res1.Decision.Reason, "spring.field-injection") {
		t.Fatalf("step 1: block reason missing rule: %q", res1.Decision.Reason)
	}
	if got := readCycleCount(t, dir, "rt-claude"); got != 1 {
		t.Fatalf("step 1: cycle counter = %d, want 1", got)
	}

	// Step 2: agent "fixes" by switching to constructor injection.
	good := `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class RoundTrip {
  private final Thing t;
  @Autowired public RoundTrip(Thing t) { this.t = t; }
}
`
	revertEdit(t, dir, "API/src/main/java/RoundTrip.java", good)

	res2 := runClaudeHook(t, dir, stopInput("rt-claude", tr, true))
	if res2.ExitCode != 0 || res2.Decision != nil {
		t.Fatalf("step 2: fixed code should pass, got exit=%d decision=%+v\nstderr: %s",
			res2.ExitCode, res2.Decision, res2.Stderr)
	}
	// Cycle counter must not increment on pass.
	if got := readCycleCount(t, dir, "rt-claude"); got != 1 {
		t.Errorf("step 2: cycle counter = %d, want 1 (unchanged on pass)", got)
	}
}

// ─── Cursor round-trip ──────────────────────────────────────────────────────

func TestStopHook_RoundTrip_Cursor(t *testing.T) {
	dir := setupStopHook(t, "cursor", projectCfg("Model", "API"))

	writeJava(t, dir, "API/src/main/java/RT.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class RT {
  @Autowired
  private Thing t;
}
`)

	// Step 1: loop_count=0 → hook emits followup_message with findings.
	res1 := runStopHook(t, "cursor", dir, map[string]any{"loop_count": 0})
	d1 := parseCursorDecision(t, res1)
	if d1 == nil {
		t.Fatalf("step 1: expected followup_message, got empty stdout\nstderr: %s", res1.Stderr)
	}
	if !strings.Contains(d1.FollowupMessage, "spring.field-injection") {
		t.Fatalf("step 1: followup missing rule: %q", d1.FollowupMessage)
	}

	// Step 2: agent fixes, Cursor increments loop_count → hook sees the fix
	// and returns empty stdout (turn completes).
	revertEdit(t, dir, "API/src/main/java/RT.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class RT {
  private final Thing t;
  @Autowired public RT(Thing t) { this.t = t; }
}
`)
	res2 := runStopHook(t, "cursor", dir, map[string]any{"loop_count": 1})
	if strings.TrimSpace(res2.Stdout) != "" {
		t.Fatalf("step 2: fixed code should not emit followup, got: %q", res2.Stdout)
	}
}

// ─── Codex round-trip ───────────────────────────────────────────────────────

func TestStopHook_RoundTrip_Codex(t *testing.T) {
	dir := setupStopHook(t, "codex", projectCfg("Model", "API"))

	writeJava(t, dir, "API/src/main/java/RT.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class RT {
  @Autowired
  private Thing t;
}
`)

	// Step 1.
	res1 := runStopHook(t, "codex", dir,
		map[string]any{"turn_id": "rt-codex", "stop_hook_active": false})
	if strings.TrimSpace(res1.Stdout) == "" {
		t.Fatalf("step 1: expected block decision, got empty stdout\nstderr: %s", res1.Stderr)
	}
	if !strings.Contains(res1.Stdout, "spring.field-injection") {
		t.Fatalf("step 1: output missing rule: %q", res1.Stdout)
	}
	if !strings.Contains(res1.Stdout, `"decision":"block"`) {
		t.Fatalf("step 1: output missing Claude-schema decision: %q", res1.Stdout)
	}

	// Step 2: fix, loop continues via stop_hook_active=true.
	revertEdit(t, dir, "API/src/main/java/RT.java", `
package x;
import org.springframework.beans.factory.annotation.Autowired;
public class RT {
  private final Thing t;
  @Autowired public RT(Thing t) { this.t = t; }
}
`)
	res2 := runStopHook(t, "codex", dir,
		map[string]any{"turn_id": "rt-codex", "stop_hook_active": true})
	if strings.TrimSpace(res2.Stdout) != "" {
		t.Fatalf("step 2: fixed code should pass, got stdout: %q", res2.Stdout)
	}
}
