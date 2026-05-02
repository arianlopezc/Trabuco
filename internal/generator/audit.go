package generator

import "fmt"

// generateAuditArtifacts emits the security-audit toolchain into a generated
// project. The audit's source-of-truth is the canonical checklist (six
// markdown files in `.ai/security-audit/`); each coding-agent variant gets a
// thin entry-point file that points back at that shared checklist.
//
// File layout (relative to project root):
//
//	.ai/security-audit/checklist.md                  (master index — always)
//	.ai/security-audit/checklist-{auth,ai-surface,
//	    aiagent-java,data-events,web-infra}.md       (domain detail — always)
//
//	.claude/agents/security-audit-orchestrator.md    (if claude selected)
//	.claude/agents/security-audit-{auth,ai-surface,
//	    aiagent-java,data-events,web-infra}.md       (if claude selected)
//	.claude/skills/audit/SKILL.md                    (if claude selected)
//
//	.cursor/rules/security-audit.mdc                 (if cursor selected)
//	.github/instructions/security-audit.instructions.md   (if copilot selected)
//	.codex/security-audit.md                         (if codex selected)
//
// The shared `.ai/security-audit/` checklist is emitted regardless of which
// coding agents are selected — it's the data layer every variant reads.
// Without it, Cursor/Copilot/Codex have nothing to walk; with it, even
// projects that selected zero AI agents can be audited later by setting one
// up and running `trabuco sync`.
//
// Claude gets the orchestrator + 5 domain specialists because it's the only
// supported agent with first-class subagent dispatch. The non-Claude
// variants get a single guidance file each that walks the same checklist
// sequentially (per Q4=a in the toolchain plan).
func (g *Generator) generateAuditArtifacts() error {
	// Shared per-project checklist — emitted always.
	checklistFiles := []string{
		"checklist.md",
		"checklist-auth.md",
		"checklist-ai-surface.md",
		"checklist-aiagent-java.md",
		"checklist-data-events.md",
		"checklist-web-infra.md",
	}
	for _, name := range checklistFiles {
		src := "ai/security-audit/" + name + ".tmpl"
		dst := ".ai/security-audit/" + name
		if err := g.writeTemplate(src, dst); err != nil {
			return fmt.Errorf("failed to write audit checklist %s: %w", name, err)
		}
	}

	// Claude variant: full subagent system + skill entry.
	if g.config.HasAIAgent("claude") {
		claudeAgents := []string{
			"security-audit-orchestrator.md",
			"security-audit-auth.md",
			"security-audit-ai-surface.md",
			"security-audit-aiagent-java.md",
			"security-audit-data-events.md",
			"security-audit-web-infra.md",
		}
		for _, name := range claudeAgents {
			src := "claude/agents/" + name + ".tmpl"
			dst := ".claude/agents/" + name
			if err := g.writeTemplate(src, dst); err != nil {
				return fmt.Errorf("failed to write claude audit agent %s: %w", name, err)
			}
		}
		if err := g.writeTemplate("claude/skills/audit/SKILL.md.tmpl", ".claude/skills/audit/SKILL.md"); err != nil {
			return fmt.Errorf("failed to write claude audit skill: %w", err)
		}
	}

	// Cursor variant: single rule file pointing at the shared checklist.
	if g.config.HasAIAgent("cursor") {
		if err := g.writeTemplate("cursor/rules/security-audit.mdc.tmpl", ".cursor/rules/security-audit.mdc"); err != nil {
			return fmt.Errorf("failed to write cursor audit rule: %w", err)
		}
	}

	// Copilot variant: applyTo-globbed instructions file. Lives under
	// .github/instructions/ — Copilot's documented location for additional
	// instruction files (the existing `java.instructions.md` for this
	// project lives there too).
	if g.config.HasAIAgent("copilot") {
		if err := g.writeTemplate("copilot/instructions/security-audit.instructions.md.tmpl", ".github/instructions/security-audit.instructions.md"); err != nil {
			return fmt.Errorf("failed to write copilot audit instructions: %w", err)
		}
	}

	// Codex variant: AGENTS.md-style guidance file under `.codex/`.
	if g.config.HasAIAgent("codex") {
		if err := g.writeTemplate("codex/security-audit.md.tmpl", ".codex/security-audit.md"); err != nil {
			return fmt.Errorf("failed to write codex audit guidance: %w", err)
		}
	}

	return nil
}
