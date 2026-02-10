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

	return nil
}

// generateMetadata generates the .trabuco.json metadata file
func (g *Generator) generateMetadata(version string) error {
	metadata := config.NewMetadataFromConfig(g.config, version)
	return config.SaveMetadata(g.outDir, metadata)
}
