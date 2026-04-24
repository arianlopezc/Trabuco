package sync

import "testing"

func TestInJurisdiction_Allowed(t *testing.T) {
	cases := []string{
		// Top-level AI docs
		"CLAUDE.md",
		"AGENTS.md",
		// Prompts and task guides
		".ai/prompts/add-entity.md",
		".ai/prompts/JAVA_CODE_QUALITY.md",
		".ai/README.md",
		// Claude Code surface
		".claude/settings.json",
		".claude/rules/code-review.md",
		".claude/skills/review/SKILL.md",
		".claude/agents/code-reviewer.md",
		".claude/hooks/format.sh",
		".claude/HOOKS.md",
		// Cursor
		".cursor/rules/java.mdc",
		".cursor/hooks.json",
		".cursor/hooks/require-review.sh",
		".cursor/rules/add-entity.mdc",
		// Codex
		".codex/hooks.json",
		".codex/config.toml",
		".codex/hooks/require-review.sh",
		// Cross-tool skills (Codex fanout)
		".agents/skills/commit/SKILL.md",
		// Copilot
		".github/instructions/java.instructions.md",
		".github/skills/review/SKILL.md",
		".github/workflows/copilot-setup-steps.yml",
		".github/copilot-instructions.md",
		// Review infra
		".github/scripts/review-checks.sh",
		".trabuco/review.config.json",
	}
	for _, p := range cases {
		if !InJurisdiction(p) {
			t.Errorf("expected %q to be in jurisdiction", p)
		}
	}
}

func TestInJurisdiction_Forbidden(t *testing.T) {
	cases := []string{
		// Business code
		"Model/src/main/java/com/example/Entity.java",
		"API/src/main/java/com/example/Controller.java",
		"SQLDatastore/src/main/resources/db/migration/V1__init.sql",
		// POMs
		"pom.xml",
		"Model/pom.xml",
		// Runtime config
		"API/src/main/resources/application.yml",
		"Worker/src/main/resources/application.yml",
		// Infra
		"docker-compose.yml",
		".env",
		".env.example",
		".dockerignore",
		// CI (non-copilot)
		".github/workflows/ci.yml",
		".github/workflows/release.yml",
		// Project-level docs the user owns
		"README.md",
		// IntelliJ
		".run/API.run.xml",
		// Metadata (owned by doctor / init, not sync)
		".trabuco.json",
		// Session state, explicitly excluded
		".ai/checkpoint.json",
		// Git internals
		".git/HEAD",
		".gitignore",
	}
	for _, p := range cases {
		if InJurisdiction(p) {
			t.Errorf("expected %q to be OUT of jurisdiction", p)
		}
	}
}

func TestInJurisdiction_PathTraversal(t *testing.T) {
	cases := []string{
		"",
		".",
		"..",
		"../etc/passwd",
		"/etc/passwd",
		"/CLAUDE.md",
	}
	for _, p := range cases {
		if InJurisdiction(p) {
			t.Errorf("expected %q to be REJECTED (traversal/absolute)", p)
		}
	}
}
