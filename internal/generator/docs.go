package generator

import (
	"fmt"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// generateDocs generates all documentation files
func (g *Generator) generateDocs() error {
	// Generate .gitignore
	if err := g.writeTemplate("docs/gitignore.tmpl", ".gitignore"); err != nil {
		return err
	}

	// Generate README.md
	if err := g.writeTemplate("docs/README.md.tmpl", "README.md"); err != nil {
		return err
	}

	// Generate AGENTS.md cross-tool baseline first (when any AI agent is selected).
	// This is written before per-agent files so that Codex (which uses AGENTS.md as its
	// primary context file) can overwrite it with the full content from CLAUDE.md.tmpl.
	if g.config.HasAnyAIAgent() {
		if err := g.writeTemplate("docs/AGENTS.md.tmpl", "AGENTS.md"); err != nil {
			return err
		}
	}

	// Generate AI agent context files for each selected agent.
	// All agents use the same template content (CLAUDE.md.tmpl), just different file paths,
	// prompts directories, and optional frontmatter per agent conventions.
	// Codex is skipped here — it uses the concise baseline AGENTS.md (written above) because
	// Codex's official guidance recommends short, focused AGENTS.md files (20-30 core lines).
	for _, agent := range g.config.GetSelectedAIAgents() {
		if agent.ID == "codex" {
			continue // Codex uses the baseline AGENTS.md, not the verbose CLAUDE.md
		}

		promptsDir := ".ai/prompts"
		frontmatter := ""

		switch agent.ID {
		case "claude":
			promptsDir = ".claude/rules"
		case "cursor":
			// Cursor .mdc files require YAML frontmatter with alwaysApply
			frontmatter = "description: Project architecture and coding standards\nalwaysApply: true\n"
		}

		data := &templateData{
			ProjectConfig: g.config,
			PromptsDir:    promptsDir,
			TaskGuidesDir: ".ai/prompts",
			Frontmatter:   frontmatter,
			Agent:         agent.ID,
		}
		if err := g.writeTemplateWithData("docs/CLAUDE.md.tmpl", agent.FilePath, data); err != nil {
			return err
		}
	}

	// Generate .ai directory with prompts and checkpoint (only if any AI agent is selected)
	if g.config.HasAnyAIAgent() {
		if err := g.generateAIDirectory(); err != nil {
			return err
		}
	}

	// Generate docker-compose.yml and .env.example when a runtime module needs a datastore
	if g.config.NeedsDockerCompose() {
		if err := g.writeTemplate("docker/docker-compose.yml.tmpl", "docker-compose.yml"); err != nil {
			return err
		}
		if err := g.writeTemplate("docker/env.example.tmpl", ".env.example"); err != nil {
			return err
		}
	}

	// Generate LocalStack init script for SQS
	if g.config.UsesSQS() {
		if err := g.writeTemplateExecutable("docker/localstack-init/ready.d/init-sqs.sh.tmpl", "localstack-init/ready.d/init-sqs.sh"); err != nil {
			return err
		}
	}

	// Generate .dockerignore when API or Worker is selected
	if g.config.HasModule(config.ModuleAPI) || g.config.HasModule(config.ModuleWorker) {
		if err := g.writeTemplate("docker/dockerignore.tmpl", ".dockerignore"); err != nil {
			return err
		}
	}

	// Generate CI workflow when a CI provider is configured
	if g.config.HasCIProvider("github") {
		if err := g.writeTemplate("github/workflows/ci.yml.tmpl", ".github/workflows/ci.yml"); err != nil {
			return err
		}
		// Emit the deterministic review check script when review is enabled.
		// Mirrors what the code-reviewer / performance-reviewer subagents do,
		// as a CI gate. Script is executable and the CI workflow invokes it.
		if g.config.ReviewEnabled() {
			if err := g.writeTemplateExecutable("github/scripts/review-checks.sh.tmpl", ".github/scripts/review-checks.sh"); err != nil {
				return err
			}
		}
	}

	// Generate Claude Code specific files when Claude is selected
	if g.config.HasAIAgent("claude") {
		if err := g.generateClaudeCodeFiles(); err != nil {
			return err
		}
	}

	// Generate Cursor specific files when Cursor is selected
	if g.config.HasAIAgent("cursor") {
		if err := g.generateCursorFiles(); err != nil {
			return err
		}
	}

	// Generate Copilot specific files when Copilot is selected
	if g.config.HasAIAgent("copilot") {
		if err := g.generateCopilotFiles(); err != nil {
			return err
		}
	}

	// Generate Codex specific files when Codex is selected
	if g.config.HasAIAgent("codex") {
		if err := g.generateCodexFiles(); err != nil {
			return err
		}
	}

	return nil
}

// generateAIDirectory generates the .ai directory with prompts and checkpoint
func (g *Generator) generateAIDirectory() error {
	// Prompt templates use {{.PromptsDir}} for cross-references between files.
	// In .ai/prompts/ (shared cross-agent location), references point to .ai/prompts/.
	aiData := &templateData{
		ProjectConfig: g.config,
		PromptsDir:    ".ai/prompts",
	}

	// Generate .ai/README.md
	if err := g.writeTemplate("ai/README.md.tmpl", ".ai/README.md"); err != nil {
		return err
	}

	// Generate .ai/checkpoint.json
	if err := g.writeTemplate("ai/checkpoint.json.tmpl", ".ai/checkpoint.json"); err != nil {
		return err
	}

	// Generate code quality specification (always - this is the core quality guide)
	if err := g.writeTemplateWithData("ai/prompts/JAVA_CODE_QUALITY.md.tmpl", ".ai/prompts/JAVA_CODE_QUALITY.md", aiData); err != nil {
		return err
	}

	// Generate code review guide (always - for proactive self-review)
	if err := g.writeTemplateWithData("ai/prompts/code-review.md.tmpl", ".ai/prompts/code-review.md", aiData); err != nil {
		return err
	}

	// Generate .ai/prompts/add-entity.md (always, if Model module exists)
	if g.config.HasModule(config.ModuleModel) {
		if err := g.writeTemplateWithData("ai/prompts/add-entity.md.tmpl", ".ai/prompts/add-entity.md", aiData); err != nil {
			return err
		}
	}

	// Generate .ai/prompts/add-endpoint.md (only if API module exists)
	if g.config.HasModule(config.ModuleAPI) {
		if err := g.writeTemplateWithData("ai/prompts/add-endpoint.md.tmpl", ".ai/prompts/add-endpoint.md", aiData); err != nil {
			return err
		}
	}

	// Generate .ai/prompts/add-job.md (only if Worker module exists)
	if g.config.HasModule(config.ModuleWorker) {
		if err := g.writeTemplateWithData("ai/prompts/add-job.md.tmpl", ".ai/prompts/add-job.md", aiData); err != nil {
			return err
		}
	}

	// Generate .ai/prompts/add-event.md (only if EventConsumer module exists)
	if g.config.HasModule(config.ModuleEventConsumer) {
		if err := g.writeTemplateWithData("ai/prompts/add-event.md.tmpl", ".ai/prompts/add-event.md", aiData); err != nil {
			return err
		}
	}

	// AI Agent prompts (only if AIAgent module exists)
	if g.config.HasModule(config.ModuleAIAgent) {
		aiAgentPrompts := []struct{ tmpl, out string }{
			{"ai/prompts/add-tool.md.tmpl", ".ai/prompts/add-tool.md"},
			{"ai/prompts/add-skill.md.tmpl", ".ai/prompts/add-skill.md"},
			{"ai/prompts/add-guardrail-rule.md.tmpl", ".ai/prompts/add-guardrail-rule.md"},
			{"ai/prompts/add-knowledge-entry.md.tmpl", ".ai/prompts/add-knowledge-entry.md"},
		}
		for _, p := range aiAgentPrompts {
			if err := g.writeTemplateWithData(p.tmpl, p.out, aiData); err != nil {
				return fmt.Errorf("failed to generate %s: %w", p.out, err)
			}
		}
	}

	// Shared module prompts
	if g.config.HasModule(config.ModuleShared) {
		if err := g.writeTemplateWithData(
			"ai/prompts/add-service.md.tmpl",
			".ai/prompts/add-service.md",
			aiData,
		); err != nil {
			return fmt.Errorf("failed to generate add-service.md: %w", err)
		}
	}

	// SQLDatastore module prompts
	if g.config.HasModule(config.ModuleSQLDatastore) {
		sqlPrompts := []struct{ tmpl, out string }{
			{"ai/prompts/add-repository-method.md.tmpl", ".ai/prompts/add-repository-method.md"},
			{"ai/prompts/add-migration.md.tmpl", ".ai/prompts/add-migration.md"},
		}
		for _, p := range sqlPrompts {
			if err := g.writeTemplateWithData(p.tmpl, p.out, aiData); err != nil {
				return fmt.Errorf("failed to generate %s: %w", p.out, err)
			}
		}
	}

	// Generate .ai/prompts/extending-the-project.md (always — guides adding auth, caching, etc.)
	if err := g.writeTemplateWithData("ai/prompts/extending-the-project.md.tmpl", ".ai/prompts/extending-the-project.md", aiData); err != nil {
		return err
	}

	// Generate .ai/prompts/testing-guide.md (always — comprehensive testing playbook)
	if err := g.writeTemplateWithData("ai/prompts/testing-guide.md.tmpl", ".ai/prompts/testing-guide.md", aiData); err != nil {
		return err
	}

	return nil
}

// generateClaudeCodeFiles generates Claude Code specific configuration files
func (g *Generator) generateClaudeCodeFiles() error {
	// Generate .claude/settings.json with hooks and permissions
	if err := g.writeTemplate("claude/settings.json.tmpl", ".claude/settings.json"); err != nil {
		return err
	}

	// Generate .claude/skills/ directory with skill templates
	// Each skill must be in its own directory as SKILL.md
	if err := g.writeTemplate("claude/skills/commit.md.tmpl", ".claude/skills/commit/SKILL.md"); err != nil {
		return err
	}

	if err := g.writeTemplate("claude/skills/pr.md.tmpl", ".claude/skills/pr/SKILL.md"); err != nil {
		return err
	}

	if err := g.writeTemplate("claude/skills/review.md.tmpl", ".claude/skills/review/SKILL.md"); err != nil {
		return err
	}

	// Generate path-scoped rules to .claude/rules/ (Claude Code's official auto-discovery location).
	// Rules include `paths:` frontmatter so they only load when matching files are accessed,
	// keeping context budget efficient instead of loading 1000+ lines at every session start.
	// Task playbooks (add-entity, add-endpoint, etc.) are NOT placed in rules — they live
	// only in .ai/prompts/ and are referenced from the main CLAUDE.md file.
	javaRuleData := &templateData{
		ProjectConfig: g.config,
		PromptsDir:    ".claude/rules",
		RulePaths:     `  - "**/*.java"`,
	}

	if err := g.writeTemplateWithData("ai/prompts/JAVA_CODE_QUALITY.md.tmpl", ".claude/rules/JAVA_CODE_QUALITY.md", javaRuleData); err != nil {
		return err
	}

	if err := g.writeTemplateWithData("ai/prompts/code-review.md.tmpl", ".claude/rules/code-review.md", javaRuleData); err != nil {
		return err
	}

	testRuleData := &templateData{
		ProjectConfig: g.config,
		PromptsDir:    ".claude/rules",
		RulePaths:     "  - \"**/*Test.java\"\n  - \"**/*Tests.java\"\n  - \"**/*IT.java\"",
	}

	if err := g.writeTemplateWithData("ai/prompts/testing-guide.md.tmpl", ".claude/rules/testing-guide.md", testRuleData); err != nil {
		return err
	}

	// Review subagents, hooks, and module-gated review skills. Emission is gated
	// on ReviewConfig.Mode — "off" short-circuits inside generateReviewArtifacts.
	if err := g.generateReviewArtifacts(); err != nil {
		return err
	}

	return nil
}

// generateCursorFiles generates Cursor specific configuration files
func (g *Generator) generateCursorFiles() error {
	// Generate .cursor/rules/java.mdc with Java coding rules
	if err := g.writeTemplate("cursor/rules/java.mdc.tmpl", ".cursor/rules/java.mdc"); err != nil {
		return err
	}

	// Generate .cursor/hooks.json for auto-formatting
	if err := g.writeTemplate("cursor/hooks.json.tmpl", ".cursor/hooks.json"); err != nil {
		return err
	}

	return nil
}

// generateCopilotFiles generates GitHub Copilot specific configuration files
func (g *Generator) generateCopilotFiles() error {
	// Generate .github/workflows/copilot-setup-steps.yml for cloud coding agent
	if err := g.writeTemplate("copilot/copilot-setup-steps.yml.tmpl", ".github/workflows/copilot-setup-steps.yml"); err != nil {
		return err
	}

	// Generate .github/instructions/java.instructions.md with scoped Java rules
	if err := g.writeTemplate("copilot/instructions/java.instructions.md.tmpl", ".github/instructions/java.instructions.md"); err != nil {
		return err
	}

	return nil
}

// generateCodexFiles generates Codex specific configuration files
func (g *Generator) generateCodexFiles() error {
	// Generate .codex/hooks.json with auto-formatting hooks
	if err := g.writeTemplate("codex/hooks.json.tmpl", ".codex/hooks.json"); err != nil {
		return err
	}

	// Generate .codex/config.toml with hooks feature flag and MCP config
	if err := g.writeTemplate("codex/config.toml.tmpl", ".codex/config.toml"); err != nil {
		return err
	}

	return nil
}

// generateMetadata generates the .trabuco.json metadata file
func (g *Generator) generateMetadata(version string) error {
	metadata := config.NewMetadataFromConfig(g.config, version)
	return config.SaveMetadata(g.outDir, metadata)
}
