package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/generator"
)

// TestRun_Blockers_NotATrabucoProject verifies sync refuses to proceed on a
// directory that has no .trabuco.json. This is the first line of defense:
// even if a user points sync at /tmp or an unrelated project, we don't
// guess, we refuse.
func TestRun_Blockers_NotATrabucoProject(t *testing.T) {
	tmp := t.TempDir()

	plan, err := Run(tmp, "test", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !plan.Blocked() {
		t.Fatalf("expected plan to be blocked for non-Trabuco dir")
	}
	if len(plan.WouldAdd) != 0 {
		t.Errorf("blocked plan should never list additions: %v", plan.WouldAdd)
	}
}

// TestRun_RoundTrip_FreshInit_IsNoOp is the core correctness guarantee for
// sync: running sync against a freshly generated project must be a no-op.
// If this ever fails, init and sync have diverged — the generator
// produces a file the sync path isn't expecting, or vice versa. Either way,
// it means a user's first sync after generation would see spurious drift.
func TestRun_RoundTrip_FreshInit_IsNoOp(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "roundtrip-project")

	cfg := &config.ProjectConfig{
		ProjectName: "roundtrip",
		GroupID:     "com.example.roundtrip",
		ArtifactID:  "roundtrip",
		JavaVersion: "21",
		Modules: []string{
			config.ModuleModel,
			config.ModuleSQLDatastore,
			config.ModuleShared,
			config.ModuleAPI,
		},
		Database: config.DatabasePostgreSQL,
		AIAgents: []string{"claude", "cursor", "codex", "copilot"},
		Review:   config.ReviewConfig{Mode: config.ReviewModeFull},
	}

	gen, err := generator.NewWithVersionAt(cfg, "test-cli", projectDir)
	if err != nil {
		t.Fatalf("generator init: %v", err)
	}
	// The generator expects to create the project directory itself, so we
	// must NOT pre-create projectDir.
	if err := gen.Generate(); err != nil {
		t.Fatalf("init: %v", err)
	}

	plan, err := Run(projectDir, "test-cli", false)
	if err != nil {
		t.Fatalf("sync run: %v", err)
	}
	if plan.Blocked() {
		t.Fatalf("fresh project should not be blocked: %v", plan.Blockers)
	}
	if plan.HasWork() {
		t.Fatalf("fresh init → sync must be a no-op, but plan adds: %v", plan.WouldAdd)
	}
}

// TestRun_Apply_RestoresDeletedFile removes one AI-tooling file from a freshly
// generated project, runs sync with --apply, and verifies the file is
// restored with the same content the generator would produce. It also
// verifies that business files are untouched (jurisdiction boundary).
func TestRun_Apply_RestoresDeletedFile(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "apply-project")

	cfg := &config.ProjectConfig{
		ProjectName: "apply",
		GroupID:     "com.example.apply",
		ArtifactID:  "apply",
		JavaVersion: "21",
		Modules: []string{
			config.ModuleModel,
			config.ModuleSQLDatastore,
			config.ModuleShared,
			config.ModuleAPI,
		},
		Database: config.DatabasePostgreSQL,
		AIAgents: []string{"claude"},
		Review:   config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	gen, _ := generator.NewWithVersionAt(cfg, "test-cli", projectDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Capture expected content and snapshot a couple of business-file mtimes
	// so we can verify sync didn't touch them.
	deleted := filepath.Join(projectDir, ".claude", "skills", "commit", "SKILL.md")
	expectedContent, err := os.ReadFile(deleted)
	if err != nil {
		t.Fatalf("read before delete: %v", err)
	}
	pomInfo, err := os.Stat(filepath.Join(projectDir, "pom.xml"))
	if err != nil {
		t.Fatalf("stat pom: %v", err)
	}
	pomMTime := pomInfo.ModTime()

	if err := os.Remove(deleted); err != nil {
		t.Fatalf("delete target: %v", err)
	}

	// Dry-run first. Confirm the plan surfaces exactly this file.
	plan, err := Run(projectDir, "test-cli", false)
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if plan.Blocked() {
		t.Fatalf("unexpected blockers: %v", plan.Blockers)
	}
	if len(plan.WouldAdd) != 1 || plan.WouldAdd[0] != ".claude/skills/commit/SKILL.md" {
		t.Fatalf("dry-run plan should list the deleted file, got: %v", plan.WouldAdd)
	}

	// Apply. File should be restored with identical content.
	appliedPlan, err := Run(projectDir, "test-cli", true)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(appliedPlan.WouldAdd) != 1 {
		t.Fatalf("applied plan should reflect the one file it added: %v", appliedPlan.WouldAdd)
	}
	got, err := os.ReadFile(deleted)
	if err != nil {
		t.Fatalf("read restored: %v", err)
	}
	if string(got) != string(expectedContent) {
		t.Fatalf("restored content differs from original (len %d vs %d)", len(got), len(expectedContent))
	}

	// Jurisdiction boundary check: business files untouched.
	pomInfo2, err := os.Stat(filepath.Join(projectDir, "pom.xml"))
	if err != nil {
		t.Fatalf("stat pom after: %v", err)
	}
	if !pomInfo2.ModTime().Equal(pomMTime) {
		t.Errorf("pom.xml mtime changed (expected untouched)")
	}

	// Second run must be a no-op (idempotency).
	plan2, err := Run(projectDir, "test-cli", false)
	if err != nil {
		t.Fatalf("idempotency check: %v", err)
	}
	if plan2.HasWork() {
		t.Errorf("expected idempotent no-op, got: %v", plan2.WouldAdd)
	}
}

// TestRun_JurisdictionBoundary_InOutOfJurisdictionList verifies that every
// business-code path the generator produces is correctly filtered as
// out-of-jurisdiction, never appearing in the actionable add-list.
func TestRun_JurisdictionBoundary_InOutOfJurisdictionList(t *testing.T) {
	tmp := t.TempDir()
	projectDir := filepath.Join(tmp, "boundary-project")

	cfg := &config.ProjectConfig{
		ProjectName: "boundary",
		GroupID:     "com.example.boundary",
		ArtifactID:  "boundary",
		JavaVersion: "21",
		Modules: []string{
			config.ModuleModel,
			config.ModuleSQLDatastore,
			config.ModuleShared,
			config.ModuleAPI,
		},
		Database: config.DatabasePostgreSQL,
		AIAgents: []string{"claude"},
		Review:   config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	gen, _ := generator.NewWithVersionAt(cfg, "test-cli", projectDir)
	if err := gen.Generate(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Wipe the project of every file except .trabuco.json so sync sees every
	// generator-produced file as "missing." The filter should route business
	// files to OutOfJurisdiction and AI files to WouldAdd.
	wipeAllExceptMetadata(t, projectDir)

	plan, err := Run(projectDir, "test-cli", false)
	if err != nil {
		t.Fatalf("sync run: %v", err)
	}

	// Paths that MUST appear in out_of_jurisdiction. If any of these ever
	// end up in WouldAdd, sync has become unsafe.
	forbiddenInAdd := []string{
		"pom.xml",
		"Model/pom.xml",
		"Model/src/main/java",
		"API/src/main/java",
		"API/src/main/resources/application.yml",
		"SQLDatastore/src/main/resources/db/migration",
		"docker-compose.yml",
		".env.example",
		".gitignore",
		".run/",
		"README.md",
	}
	for _, needle := range forbiddenInAdd {
		for _, add := range plan.WouldAdd {
			if strings.HasPrefix(add, needle) {
				t.Errorf("business path leaked into WouldAdd: %s (matched %s)", add, needle)
			}
		}
	}

	// Sanity: AI paths DID make it into WouldAdd. If this list is empty the
	// test is meaningless.
	if len(plan.WouldAdd) == 0 {
		t.Fatalf("expected wiped project to surface AI additions, got none")
	}
}

// TestPlan_JSONRoundTrip verifies the JSON output shape is stable and
// round-trips through json.Unmarshal — this is the contract for CI and
// agent consumers of sync --json.
func TestPlan_JSONRoundTrip(t *testing.T) {
	plan := &Plan{
		ProjectPath:    "/tmp/demo",
		CLIVersion:     "v1.9.3",
		StampedVersion: "v1.8.0",
		Modules:        []string{"Model", "API"},
		AIAgents:       []string{"claude"},
		WouldAdd:       []string{".claude/skills/review/SKILL.md"},
		AlreadyPresent: []string{"CLAUDE.md"},
	}
	buf := &strings.Builder{}
	if err := plan.WriteJSON(buf); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	var decoded Plan
	if err := json.Unmarshal([]byte(buf.String()), &decoded); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if decoded.ProjectPath != plan.ProjectPath || len(decoded.WouldAdd) != 1 {
		t.Errorf("round-trip mismatch: %+v", decoded)
	}
}

// wipeAllExceptMetadata deletes every file under projectDir except
// .trabuco.json. Used by the jurisdiction test to force sync to see
// EVERY file the generator produces as missing.
func wipeAllExceptMetadata(t *testing.T, projectDir string) {
	t.Helper()
	err := filepath.Walk(projectDir, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(projectDir, p)
		if filepath.ToSlash(rel) == ".trabuco.json" {
			return nil
		}
		return os.Remove(p)
	})
	if err != nil {
		t.Fatalf("wipe: %v", err)
	}
}
