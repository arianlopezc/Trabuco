package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// Skill-emission tests lock in the single-catalog → four-tool fanout. A
// regression here (wrong frontmatter, missing module gate, name/dir
// mismatch on Copilot's strict rule) silently degrades discoverability in
// generated projects — users' /add-entity stops auto-completing, Copilot
// refuses to load the skill, etc.

func generateWithSkills(t *testing.T, cfg *config.ProjectConfig) string {
	t.Helper()
	tempDir := t.TempDir()
	oldWd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	t.Cleanup(func() { _ = os.Chdir(oldWd) })

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return filepath.Join(tempDir, cfg.ProjectName)
}

func readAll(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(b)
}

// A helper assertion: the skill file exists at the expected path and its
// frontmatter contains expected substrings. Keeps the per-tool test cases
// concise while asserting the real compliance surface.
func assertSkillPresent(t *testing.T, projectDir, toolPath, _ string, expectedFrontmatter []string) {
	t.Helper()
	full := filepath.Join(projectDir, toolPath)
	if _, err := os.Stat(full); err != nil {
		t.Fatalf("expected skill at %s: %v", toolPath, err)
	}
	content := readAll(t, full)
	// Split frontmatter from body: content starts "---\n<fm>\n---\n\n<body>".
	parts := strings.SplitN(content, "---\n", 3)
	if len(parts) < 3 {
		t.Fatalf("%s: not a well-formed frontmatter/body file:\n%s", toolPath, content)
	}
	frontmatter := parts[1]
	body := parts[2]

	for _, want := range expectedFrontmatter {
		if !strings.Contains(frontmatter, want) {
			t.Errorf("%s frontmatter missing %q:\n%s", toolPath, want, frontmatter)
		}
	}
	if strings.TrimSpace(body) == "" {
		t.Errorf("%s has empty body", toolPath)
	}
}

func assertSkillAbsent(t *testing.T, projectDir, toolPath string) {
	t.Helper()
	full := filepath.Join(projectDir, toolPath)
	if _, err := os.Stat(full); err == nil {
		t.Errorf("expected NO skill at %s, but it was emitted", toolPath)
	}
}

// ─── Universal skills emit for all 4 agents ──────────────────────────────

func TestSkills_AllAgentsEmit_UniversalSkill(t *testing.T) {
	// /add-entity is universal (no RequiredModule gate). Expect it at every
	// agent's skill location plus a Cursor rule.
	cfg := &config.ProjectConfig{
		ProjectName: "skilltest",
		GroupID:     "com.t.skilltest",
		ArtifactID:  "skilltest",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		AIAgents:    []string{"claude", "codex", "copilot", "cursor"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	dir := generateWithSkills(t, cfg)

	// Claude: name + description + user-invocable + allowed-tools + paths
	assertSkillPresent(t, dir, ".claude/skills/add-entity/SKILL.md", "add-entity",
		[]string{"name: add-entity", "user-invocable: true", "allowed-tools:"})

	// Codex: name + description + allow_implicit_invocation (distinct field)
	assertSkillPresent(t, dir, ".agents/skills/add-entity/SKILL.md", "add-entity",
		[]string{"name: add-entity", "allow_implicit_invocation: true"})

	// Copilot: name MUST match directory — that's the whole point of the
	// strict name rule. Body is present.
	assertSkillPresent(t, dir, ".github/skills/add-entity/SKILL.md", "add-entity",
		[]string{"name: add-entity"})

	// Cursor: description + globs + alwaysApply. Name field is absent (rules
	// don't have `name:`); body still present.
	cursorPath := filepath.Join(dir, ".cursor/rules/add-entity.mdc")
	if _, err := os.Stat(cursorPath); err != nil {
		t.Fatalf("cursor rule missing: %v", err)
	}
	cursorBody := readAll(t, cursorPath)
	for _, want := range []string{"description:", "globs:", "alwaysApply: false"} {
		if !strings.Contains(cursorBody, want) {
			t.Errorf("cursor rule missing %q", want)
		}
	}
}

// ─── Module gating ──────────────────────────────────────────────────────

func TestSkills_ModuleGate_AIAgentOnly(t *testing.T) {
	// /add-tool, /add-guardrail-rule, /add-knowledge-entry, /add-a2a-skill
	// require AIAgent. A minimal project (no AIAgent) must NOT emit them.
	cfg := &config.ProjectConfig{
		ProjectName: "gatetest",
		GroupID:     "com.t.gatetest",
		ArtifactID:  "gatetest",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		AIAgents:    []string{"claude"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	dir := generateWithSkills(t, cfg)

	for _, name := range []string{"add-tool", "add-guardrail-rule", "add-knowledge-entry", "add-a2a-skill"} {
		assertSkillAbsent(t, dir, ".claude/skills/"+name+"/SKILL.md")
	}
	// Universal skills still emit.
	assertSkillPresent(t, dir, ".claude/skills/add-entity/SKILL.md", "add-entity",
		[]string{"name: add-entity"})
}

func TestSkills_ModuleGate_Datastore(t *testing.T) {
	// /review-performance requires any datastore. No datastore → absent.
	cfg := &config.ProjectConfig{
		ProjectName: "nods",
		GroupID:     "com.t.nods",
		ArtifactID:  "nods",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		AIAgents:    []string{"claude"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	dir := generateWithSkills(t, cfg)
	assertSkillAbsent(t, dir, ".claude/skills/review-performance/SKILL.md")
	assertSkillAbsent(t, dir, ".claude/skills/add-migration/SKILL.md")
	assertSkillAbsent(t, dir, ".claude/skills/add-repository-method/SKILL.md")
}

func TestSkills_ModuleGate_WithDatastoreEnabled(t *testing.T) {
	cfg := &config.ProjectConfig{
		ProjectName: "withds",
		GroupID:     "com.t.withds",
		ArtifactID:  "withds",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "API"},
		Database:    "postgresql",
		AIAgents:    []string{"claude"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	dir := generateWithSkills(t, cfg)
	assertSkillPresent(t, dir, ".claude/skills/review-performance/SKILL.md", "review-performance",
		[]string{"name: review-performance"})
	assertSkillPresent(t, dir, ".claude/skills/add-migration/SKILL.md", "add-migration",
		[]string{"name: add-migration"})
	assertSkillPresent(t, dir, ".claude/skills/add-repository-method/SKILL.md", "add-repository-method",
		[]string{"name: add-repository-method"})
}

// ─── Per-tool filtering ─────────────────────────────────────────────────

func TestSkills_OnlySelectedAgentsEmit(t *testing.T) {
	// Only Claude selected → no .agents/, no .github/skills/, no .cursor/
	// rules should appear.
	cfg := &config.ProjectConfig{
		ProjectName: "claudeonly",
		GroupID:     "com.t.claudeonly",
		ArtifactID:  "claudeonly",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		AIAgents:    []string{"claude"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	dir := generateWithSkills(t, cfg)

	if _, err := os.Stat(filepath.Join(dir, ".agents/skills")); err == nil {
		t.Error(".agents/skills should not exist when codex is not selected")
	}
	if _, err := os.Stat(filepath.Join(dir, ".github/skills")); err == nil {
		t.Error(".github/skills should not exist when copilot is not selected")
	}
	// Note: .cursor/ may exist from other generation steps; skill fanout
	// specifically adds rules under .cursor/rules with add-* filenames.
	if _, err := os.Stat(filepath.Join(dir, ".cursor/rules/add-entity.mdc")); err == nil {
		t.Error("cursor rule should not exist when cursor is not selected")
	}
}

// ─── Cursor port filtering ──────────────────────────────────────────────

func TestSkills_CursorSkipsWorkflowInvocations(t *testing.T) {
	// Cursor has no skills primitive and its rules activate on globs, not
	// user invocation. /commit, /pr, /review are workflow-invocations that
	// don't translate — they must NOT emit a .cursor/rules/<name>.mdc.
	cfg := &config.ProjectConfig{
		ProjectName: "curtest",
		GroupID:     "com.t.curtest",
		ArtifactID:  "curtest",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		AIAgents:    []string{"cursor"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	dir := generateWithSkills(t, cfg)

	for _, name := range []string{"commit", "pr", "review"} {
		p := filepath.Join(dir, ".cursor/rules/"+name+".mdc")
		if _, err := os.Stat(p); err == nil {
			t.Errorf("workflow-invocation skill %q should NOT emit as cursor rule", name)
		}
	}
	// But add-entity does port — glob-activation makes sense.
	if _, err := os.Stat(filepath.Join(dir, ".cursor/rules/add-entity.mdc")); err != nil {
		t.Errorf("add-entity SHOULD port to cursor: %v", err)
	}
}

// ─── Frontmatter compliance ─────────────────────────────────────────────

func TestSkills_CopilotNameMatchesDirectory(t *testing.T) {
	// Copilot is strict: the frontmatter `name:` MUST equal the containing
	// directory name, or the skill silently fails to load.
	cfg := &config.ProjectConfig{
		ProjectName: "cpname",
		GroupID:     "com.t.cpname",
		ArtifactID:  "cpname",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "API"},
		Database:    "postgresql",
		AIAgents:    []string{"copilot"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
	}
	dir := generateWithSkills(t, cfg)

	skillsDir := filepath.Join(dir, ".github/skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("read skills dir: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("no skills emitted for copilot")
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirName := e.Name()
		skillFile := filepath.Join(skillsDir, dirName, "SKILL.md")
		content := readAll(t, skillFile)
		want := "name: " + dirName
		if !strings.Contains(content, want) {
			t.Errorf("skill %s: frontmatter name does not match directory (expected %q in content)",
				dirName, want)
		}
	}
}

// ─── Review kill-switch ─────────────────────────────────────────────────

func TestSkills_ReviewOff_EmitsNothing(t *testing.T) {
	cfg := &config.ProjectConfig{
		ProjectName: "reviewoff",
		GroupID:     "com.t.reviewoff",
		ArtifactID:  "reviewoff",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		AIAgents:    []string{"claude"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeOff},
	}
	dir := generateWithSkills(t, cfg)
	// None of the review-family skills should emit. The security-audit
	// skill is intentionally independent of review.mode (it's a
	// security gate, not a turn-scoped review automation), so its
	// directory is allowed.
	reviewSkills := []string{
		"review", "review-performance", "review-prompts",
		"add-entity", "add-endpoint", "add-service",
		"add-migration", "add-repository-method", "add-test",
		"add-tool", "add-event", "add-job", "add-knowledge-entry",
		"add-guardrail-rule", "add-a2a-skill", "pr", "commit",
	}
	for _, name := range reviewSkills {
		p := filepath.Join(dir, ".claude/skills", name, "SKILL.md")
		if _, err := os.Stat(p); err == nil {
			t.Errorf("review skill %q should not emit when review is off (found: %s)", name, p)
		}
	}
}
