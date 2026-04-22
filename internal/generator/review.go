package generator

import (
	"fmt"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// generateReviewArtifacts emits the subagents, hooks, skills, and config that
// implement the post-turn code review automation. Gated on the Claude agent
// being selected and Review.Mode != "off".
//
// File layout (all relative to project root):
//
//	.claude/agents/code-reviewer.md               (always when enabled)
//	.claude/agents/performance-reviewer.md        (if SQLDatastore || NoSQLDatastore)
//	.claude/agents/prompt-reviewer.md             (if AIAgent)
//	.claude/hooks/format.sh                       (always when enabled)
//	.claude/hooks/require-review.sh               (only when Mode == "full")
//	.claude/HOOKS.md                              (always when enabled)
//	.claude/skills/review-performance/SKILL.md    (if datastore)
//	.claude/skills/review-prompts/SKILL.md        (if AIAgent)
//	.trabuco/review.config.json                   (always when enabled — runtime kill switch source)
//
// The per-session Stop-hook guard only emits when Mode == "full". In "minimal",
// we emit the subagents and skills but skip the hook — users get description-based
// auto-delegation plus the CLAUDE.md directive, without any blocking behavior.
// GenerateReviewArtifactsOnly is the public entry point used by the
// `trabuco review install` command. It delegates to the same internal
// function used by the init-time generation pipeline. The caller is
// responsible for ensuring g.outDir is the existing project root.
func (g *Generator) GenerateReviewArtifactsOnly() error {
	return g.generateReviewArtifacts()
}

func (g *Generator) generateReviewArtifacts() error {
	if !g.config.ReviewEnabled() {
		return nil
	}

	// Claude-specific artifacts: subagents, skills, hooks, HOOKS.md. Emitted
	// only when Claude is among the selected AI agents.
	if g.config.HasAIAgent("claude") {
		if err := g.writeTemplate("claude/agents/code-reviewer.md.tmpl", ".claude/agents/code-reviewer.md"); err != nil {
			return fmt.Errorf("failed to write code-reviewer agent: %w", err)
		}
		if g.config.HasAnyDatastore() {
			if err := g.writeTemplate("claude/agents/performance-reviewer.md.tmpl", ".claude/agents/performance-reviewer.md"); err != nil {
				return fmt.Errorf("failed to write performance-reviewer agent: %w", err)
			}
		}
		if g.config.HasModule(config.ModuleAIAgent) {
			if err := g.writeTemplate("claude/agents/prompt-reviewer.md.tmpl", ".claude/agents/prompt-reviewer.md"); err != nil {
				return fmt.Errorf("failed to write prompt-reviewer agent: %w", err)
			}
		}
		if err := g.writeTemplateExecutable("claude/hooks/format.sh.tmpl", ".claude/hooks/format.sh"); err != nil {
			return fmt.Errorf("failed to write format.sh: %w", err)
		}
		if g.config.ReviewEmitsStopHook() {
			if err := g.writeTemplateExecutable("claude/hooks/require-review.sh.tmpl", ".claude/hooks/require-review.sh"); err != nil {
				return fmt.Errorf("failed to write require-review.sh: %w", err)
			}
		}
		if err := g.writeTemplate("claude/HOOKS.md.tmpl", ".claude/HOOKS.md"); err != nil {
			return fmt.Errorf("failed to write HOOKS.md: %w", err)
		}
		if g.config.HasAnyDatastore() {
			if err := g.writeTemplate("claude/skills/review-performance.md.tmpl", ".claude/skills/review-performance/SKILL.md"); err != nil {
				return fmt.Errorf("failed to write review-performance skill: %w", err)
			}
		}
		if g.config.HasModule(config.ModuleAIAgent) {
			if err := g.writeTemplate("claude/skills/review-prompts.md.tmpl", ".claude/skills/review-prompts/SKILL.md"); err != nil {
				return fmt.Errorf("failed to write review-prompts skill: %w", err)
			}
		}
	}

	// Cross-tool Stop-hook adapters. Each adapter shares the same deterministic
	// enforcement mechanism (via .github/scripts/review-checks.sh) but formats
	// output for its tool's stop-hook schema. Emit when mode=full and the tool
	// is selected. Independent of whether Claude is also selected.
	if g.config.ReviewEmitsStopHook() {
		if g.config.HasAIAgent("codex") {
			if err := g.writeTemplateExecutable("codex/hooks/require-review.sh.tmpl", ".codex/hooks/require-review.sh"); err != nil {
				return fmt.Errorf("failed to write codex require-review.sh: %w", err)
			}
		}
		if g.config.HasAIAgent("cursor") {
			if err := g.writeTemplateExecutable("cursor/hooks/require-review.sh.tmpl", ".cursor/hooks/require-review.sh"); err != nil {
				return fmt.Errorf("failed to write cursor require-review.sh: %w", err)
			}
		}
	}

	// Runtime config — the kill-switch source read by every tool's adapter.
	if err := g.writeTemplate("trabuco/review.config.json.tmpl", ".trabuco/review.config.json"); err != nil {
		return fmt.Errorf("failed to write review.config.json: %w", err)
	}

	return nil
}
