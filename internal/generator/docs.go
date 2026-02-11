package generator

import (
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

	// Generate AI agent context files for each selected agent
	// All agents use the same template content (CLAUDE.md.tmpl), just different file paths
	// The writeTemplate method handles parent directory creation automatically
	for _, agent := range g.config.GetSelectedAIAgents() {
		if err := g.writeTemplate("docs/CLAUDE.md.tmpl", agent.FilePath); err != nil {
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

	// Generate MCP configuration files when MCP module is selected
	if g.config.HasModule(config.ModuleMCP) {
		// Claude Code: .mcp.json (project root)
		if err := g.writeTemplate("docs/mcp.json.tmpl", ".mcp.json"); err != nil {
			return err
		}

		// Cursor: .cursor/mcp.json
		if err := g.writeTemplate("docs/cursor-mcp.json.tmpl", ".cursor/mcp.json"); err != nil {
			return err
		}

		// VS Code / GitHub Copilot: .vscode/mcp.json
		if err := g.writeTemplate("docs/vscode-mcp.json.tmpl", ".vscode/mcp.json"); err != nil {
			return err
		}

		// MCP README with setup instructions for all agents
		if err := g.writeTemplate("docs/MCP-README.md.tmpl", "MCP/README.md"); err != nil {
			return err
		}
	}

	return nil
}

// generateAIDirectory generates the .ai directory with prompts and checkpoint
func (g *Generator) generateAIDirectory() error {
	// Generate .ai/README.md
	if err := g.writeTemplate("ai/README.md.tmpl", ".ai/README.md"); err != nil {
		return err
	}

	// Generate .ai/checkpoint.json
	if err := g.writeTemplate("ai/checkpoint.json.tmpl", ".ai/checkpoint.json"); err != nil {
		return err
	}

	// Generate .ai/prompts/add-entity.md (always, if Model module exists)
	if g.config.HasModule(config.ModuleModel) {
		if err := g.writeTemplate("ai/prompts/add-entity.md.tmpl", ".ai/prompts/add-entity.md"); err != nil {
			return err
		}
	}

	// Generate .ai/prompts/add-endpoint.md (only if API module exists)
	if g.config.HasModule(config.ModuleAPI) {
		if err := g.writeTemplate("ai/prompts/add-endpoint.md.tmpl", ".ai/prompts/add-endpoint.md"); err != nil {
			return err
		}
	}

	// Generate .ai/prompts/add-job.md (only if Worker module exists)
	if g.config.HasModule(config.ModuleWorker) {
		if err := g.writeTemplate("ai/prompts/add-job.md.tmpl", ".ai/prompts/add-job.md"); err != nil {
			return err
		}
	}

	// Generate .ai/prompts/add-event.md (only if EventConsumer module exists)
	if g.config.HasModule(config.ModuleEventConsumer) {
		if err := g.writeTemplate("ai/prompts/add-event.md.tmpl", ".ai/prompts/add-event.md"); err != nil {
			return err
		}
	}

	return nil
}

// generateMetadata generates the .trabuco.json metadata file
func (g *Generator) generateMetadata(version string) error {
	metadata := config.NewMetadataFromConfig(g.config, version)
	return config.SaveMetadata(g.outDir, metadata)
}
