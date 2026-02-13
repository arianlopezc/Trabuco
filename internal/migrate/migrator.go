package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/cache"
	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/utils"
	"github.com/fatih/color"
)

// Migrator orchestrates the migration process
type Migrator struct {
	provider   ai.Provider
	config     *Config
	checkpoint *CheckpointManager
	scanner    *ProjectScanner
	analyzer   *DependencyAnalyzer

	// Cost tracking
	costTracker *ai.CostTracker

	// Caching
	cache *cache.Cache

	// Discovered project information
	projectInfo *ProjectInfo

	// Colors for output
	cyan   *color.Color
	green  *color.Color
	yellow *color.Color
	red    *color.Color
}

// NewMigrator creates a new Migrator
func NewMigrator(provider ai.Provider, cfg *Config) *Migrator {
	// Get model from provider for cost tracking
	var model ai.Model
	if ap, ok := provider.(*ai.AnthropicProvider); ok {
		model = ap.GetModel()
	} else {
		model = ai.ModelClaudeSonnet // Default
	}

	m := &Migrator{
		provider:    provider,
		config:      cfg,
		checkpoint:  NewCheckpointManager(cfg.SourcePath),
		scanner:     NewProjectScanner(cfg.SourcePath),
		analyzer:    NewDependencyAnalyzer(),
		costTracker: ai.NewCostTracker(model),
		cache:       cache.NewCache(""),
		cyan:        color.New(color.FgCyan),
		green:       color.New(color.FgGreen),
		yellow:      color.New(color.FgYellow),
		red:         color.New(color.FgRed),
	}

	// Set up cost update callback for real-time display
	m.costTracker.SetUpdateCallback(func(update ai.CostUpdate) {
		if cfg.Verbose {
			fmt.Printf("    [%s: %s in / %s out = %s]\n",
				update.Phase,
				ai.FormatTokens(update.InputTokens),
				ai.FormatTokens(update.OutputTokens),
				ai.FormatCost(update.TotalCost))
		}
	})

	return m
}

// Run executes the migration workflow
func (m *Migrator) Run() error {
	ctx := context.Background()

	// Check for existing checkpoint
	if m.config.Resume {
		checkpoint, err := m.checkpoint.Load()
		if err != nil {
			return fmt.Errorf("failed to load checkpoint: %w", err)
		}
		if checkpoint == nil {
			return fmt.Errorf("no checkpoint found to resume from")
		}
		m.cyan.Printf("Resuming from stage: %s\n\n", checkpoint.Stage)
	} else if m.checkpoint.HasCheckpoint() {
		m.yellow.Println("Warning: A previous migration checkpoint exists.")
		m.yellow.Println("Use --resume to continue, or delete .trabuco-migrate/ to start fresh.")
		fmt.Println()
	}

	// Stage 1: Discovery
	if err := m.runDiscovery(ctx); err != nil {
		return err
	}

	// Stage 2: Dependency Analysis
	if err := m.runDependencyAnalysis(ctx); err != nil {
		return err
	}

	// If dry run, stop here and print summary
	if m.config.DryRun {
		m.printDryRunSummary()
		return nil
	}

	// Stage 3: Entity Extraction
	if err := m.runEntityExtraction(ctx); err != nil {
		return err
	}

	// Stage 4: Repository Migration
	if err := m.runRepositoryMigration(ctx); err != nil {
		return err
	}

	// Stage 5: Service Extraction
	if err := m.runServiceExtraction(ctx); err != nil {
		return err
	}

	// Stage 6: Controller Migration
	if err := m.runControllerMigration(ctx); err != nil {
		return err
	}

	// Stage 7: Jobs Migration (if applicable)
	if m.projectInfo.HasScheduledJobs {
		if err := m.runJobsMigration(ctx); err != nil {
			return err
		}
	}

	// Stage 8: Events Migration (if applicable)
	if m.projectInfo.HasEventListeners {
		if err := m.runEventsMigration(ctx); err != nil {
			return err
		}
	}

	// Stage 9: Configuration
	if err := m.runConfiguration(ctx); err != nil {
		return err
	}

	// Stage 10: Final Assembly
	if err := m.runFinalAssembly(ctx); err != nil {
		return err
	}

	// Build if requested
	if !m.config.SkipBuild {
		if err := m.buildProject(); err != nil {
			m.yellow.Printf("\nBuild failed: %v\n", err)
			m.yellow.Println("You can try building manually:")
			fmt.Printf("  cd %s && mvn clean install\n", m.config.OutputPath)
		}
	}

	// Cleanup checkpoint on success
	if err := m.checkpoint.Cleanup(); err != nil {
		m.yellow.Printf("Warning: failed to cleanup checkpoint: %v\n", err)
	}

	// Print cost summary
	fmt.Print(m.costTracker.GetSummary())

	return nil
}

// runDiscovery scans the source project structure
func (m *Migrator) runDiscovery(ctx context.Context) error {
	m.printStageHeader("Discovery", "Scanning source project...")

	if err := m.checkpoint.StartStage(StageDiscovery); err != nil {
		return err
	}

	// Use pre-scanned project info if available (from guided flow)
	if m.config.PreScannedProjectInfo != nil {
		m.projectInfo = m.config.PreScannedProjectInfo
		m.green.Println("  Using pre-scanned project information...")
	} else {
		// Scan project
		projectInfo, err := m.scanner.Scan()
		if err != nil {
			m.checkpoint.FailStage(StageDiscovery, err)
			return fmt.Errorf("failed to scan project: %w", err)
		}
		m.projectInfo = projectInfo
	}

	// Display findings
	m.printDiscoverySummary(m.projectInfo)

	// Initialize checkpoint with output path
	if err := m.checkpoint.Initialize(m.config.SourcePath, m.config.OutputPath); err != nil {
		return err
	}

	// Save discovery data
	data := map[string]interface{}{
		"project_info": m.projectInfo,
	}

	return m.checkpoint.CompleteStage(StageDiscovery, data)
}

// runDependencyAnalysis analyzes and classifies dependencies
func (m *Migrator) runDependencyAnalysis(ctx context.Context) error {
	m.printStageHeader("Dependency Analysis", "Analyzing dependencies...")

	if err := m.checkpoint.StartStage(StageDependencies); err != nil {
		return err
	}

	// Use pre-analyzed dependencies if available (from guided flow)
	var report *DependencyReport
	if m.config.PreAnalyzedDependencies != nil {
		report = m.config.PreAnalyzedDependencies
		m.green.Println("  Using pre-analyzed dependency information...")
	} else {
		// Analyze dependencies
		report = m.analyzer.Analyze(m.projectInfo.Dependencies)
	}

	// Display report
	m.printDependencyReport(report)

	// If interactive and not from guided flow (which already confirmed), get user confirmation
	if m.config.Interactive && m.config.PreAnalyzedDependencies == nil && len(report.Replaceable) > 0 {
		if err := m.confirmDependencyReplacements(report); err != nil {
			m.checkpoint.FailStage(StageDependencies, err)
			return err
		}
	}

	// Check for blockers (skip if from guided flow which already showed these)
	if len(report.Unsupported) > 0 && m.config.Interactive && m.config.PreAnalyzedDependencies == nil {
		m.yellow.Println("\nSome dependencies are not supported and require manual migration.")
		fmt.Println("The migration will proceed, but you may need to handle these manually.")
		if !m.confirmContinue() {
			return fmt.Errorf("migration cancelled by user")
		}
	}

	// Estimate cost (skip if from guided flow which already showed this)
	if m.config.PreScannedProjectInfo == nil {
		m.printCostEstimate()

		if m.config.Interactive && !m.config.DryRun {
			if !m.confirmContinue() {
				return fmt.Errorf("migration cancelled by user")
			}
		}
	}

	data := map[string]interface{}{
		"dependency_report": report,
	}

	return m.checkpoint.CompleteStage(StageDependencies, data)
}

// runEntityExtraction extracts and converts entities
func (m *Migrator) runEntityExtraction(ctx context.Context) error {
	m.printStageHeader("Entity Extraction", "Converting entities to Trabuco Model module...")

	if err := m.checkpoint.StartStage(StageEntities); err != nil {
		return err
	}

	// Start cost tracking phase
	m.costTracker.StartPhase("Entities")

	// Create output Model module structure
	if err := m.createModuleStructure(config.ModuleModel); err != nil {
		m.checkpoint.FailStage(StageEntities, err)
		return err
	}

	entities := m.projectInfo.Entities
	total := len(entities)
	converted := 0
	skipped := 0

	for i, entity := range entities {
		m.printProgress(i+1, total, entity.Name)

		// Use AI to convert entity
		result, err := m.convertEntity(ctx, entity)
		if err != nil {
			m.yellow.Printf("  Warning: failed to convert %s: %v\n", entity.Name, err)
			continue
		}

		// Check if entity was skipped (not a real entity)
		if result.Skip {
			skipped++
			if m.config.Verbose && len(result.Notes) > 0 {
				m.yellow.Printf("  Skipped: %s\n", result.Notes[0])
			}
			continue
		}

		// Write converted files
		if err := m.writeEntityFiles(result); err != nil {
			m.yellow.Printf("  Warning: failed to write %s: %v\n", entity.Name, err)
			continue
		}
		converted++
	}

	fmt.Println()
	if skipped > 0 {
		m.green.Printf("  ✓ Converted %d entities (%d skipped - not database entities)\n", converted, skipped)
	} else {
		m.green.Printf("  ✓ Converted %d entities\n", converted)
	}

	// End cost tracking phase
	m.costTracker.EndPhase()

	return m.checkpoint.CompleteStage(StageEntities, nil)
}

// runRepositoryMigration migrates repository interfaces
func (m *Migrator) runRepositoryMigration(ctx context.Context) error {
	m.printStageHeader("Repository Migration", "Converting repositories...")

	if err := m.checkpoint.StartStage(StageRepositories); err != nil {
		return err
	}

	// Start cost tracking phase
	m.costTracker.StartPhase("Repositories")

	// Determine datastore module
	datastoreModule := config.ModuleSQLDatastore
	if m.projectInfo.UsesNoSQL {
		datastoreModule = config.ModuleNoSQLDatastore
	}

	// Create datastore module structure
	if err := m.createModuleStructure(datastoreModule); err != nil {
		m.checkpoint.FailStage(StageRepositories, err)
		return err
	}

	repos := m.projectInfo.Repositories
	total := len(repos)

	for i, repo := range repos {
		m.printProgress(i+1, total, repo.Name)

		converted, err := m.convertRepository(ctx, repo)
		if err != nil {
			m.yellow.Printf("  Warning: failed to convert %s: %v\n", repo.Name, err)
			continue
		}

		if err := m.writeRepositoryFiles(converted, datastoreModule); err != nil {
			m.yellow.Printf("  Warning: failed to write %s: %v\n", repo.Name, err)
		}
	}

	// Generate Flyway migrations
	if !m.projectInfo.UsesNoSQL {
		if err := m.generateFlywayMigrations(ctx); err != nil {
			m.yellow.Printf("  Warning: failed to generate migrations: %v\n", err)
		}
	}

	fmt.Println()
	m.green.Printf("  ✓ Converted %d repositories\n", total)

	// End cost tracking phase
	m.costTracker.EndPhase()

	return m.checkpoint.CompleteStage(StageRepositories, nil)
}

// runServiceExtraction extracts and converts services
func (m *Migrator) runServiceExtraction(ctx context.Context) error {
	m.printStageHeader("Service Extraction", "Converting services to Shared module...")

	if err := m.checkpoint.StartStage(StageServices); err != nil {
		return err
	}

	// Start cost tracking phase
	m.costTracker.StartPhase("Services")

	if err := m.createModuleStructure(config.ModuleShared); err != nil {
		m.checkpoint.FailStage(StageServices, err)
		return err
	}

	services := m.projectInfo.Services
	total := len(services)

	for i, service := range services {
		m.printProgress(i+1, total, service.Name)

		converted, err := m.convertService(ctx, service)
		if err != nil {
			m.yellow.Printf("  Warning: failed to convert %s: %v\n", service.Name, err)
			continue
		}

		if err := m.writeServiceFiles(converted); err != nil {
			m.yellow.Printf("  Warning: failed to write %s: %v\n", service.Name, err)
		}
	}

	fmt.Println()
	m.green.Printf("  ✓ Converted %d services\n", total)

	// End cost tracking phase
	m.costTracker.EndPhase()

	return m.checkpoint.CompleteStage(StageServices, nil)
}

// runControllerMigration migrates REST controllers
func (m *Migrator) runControllerMigration(ctx context.Context) error {
	m.printStageHeader("Controller Migration", "Converting controllers to API module...")

	if err := m.checkpoint.StartStage(StageControllers); err != nil {
		return err
	}

	// Start cost tracking phase
	m.costTracker.StartPhase("Controllers")

	if err := m.createModuleStructure(config.ModuleAPI); err != nil {
		m.checkpoint.FailStage(StageControllers, err)
		return err
	}

	controllers := m.projectInfo.Controllers
	total := len(controllers)

	for i, controller := range controllers {
		m.printProgress(i+1, total, controller.Name)

		converted, err := m.convertController(ctx, controller)
		if err != nil {
			m.yellow.Printf("  Warning: failed to convert %s: %v\n", controller.Name, err)
			continue
		}

		if err := m.writeControllerFiles(converted); err != nil {
			m.yellow.Printf("  Warning: failed to write %s: %v\n", controller.Name, err)
		}
	}

	fmt.Println()
	m.green.Printf("  ✓ Converted %d controllers\n", total)

	// End cost tracking phase
	m.costTracker.EndPhase()

	return m.checkpoint.CompleteStage(StageControllers, nil)
}

// runJobsMigration migrates scheduled jobs
func (m *Migrator) runJobsMigration(ctx context.Context) error {
	m.printStageHeader("Jobs Migration", "Converting scheduled jobs to Worker module...")

	if err := m.checkpoint.StartStage(StageJobs); err != nil {
		return err
	}

	// Start cost tracking phase
	m.costTracker.StartPhase("Jobs")

	if err := m.createModuleStructure(config.ModuleWorker); err != nil {
		m.checkpoint.FailStage(StageJobs, err)
		return err
	}

	jobs := m.projectInfo.ScheduledJobs
	total := len(jobs)

	for i, job := range jobs {
		m.printProgress(i+1, total, job.Name)

		converted, err := m.convertJob(ctx, job)
		if err != nil {
			m.yellow.Printf("  Warning: failed to convert %s: %v\n", job.Name, err)
			continue
		}

		if err := m.writeJobFiles(converted); err != nil {
			m.yellow.Printf("  Warning: failed to write %s: %v\n", job.Name, err)
		}
	}

	fmt.Println()
	m.green.Printf("  ✓ Converted %d jobs\n", total)

	// End cost tracking phase
	m.costTracker.EndPhase()

	return m.checkpoint.CompleteStage(StageJobs, nil)
}

// runEventsMigration migrates event listeners
func (m *Migrator) runEventsMigration(ctx context.Context) error {
	m.printStageHeader("Events Migration", "Converting event listeners to EventConsumer module...")

	if err := m.checkpoint.StartStage(StageEvents); err != nil {
		return err
	}

	// Start cost tracking phase
	m.costTracker.StartPhase("Events")

	if err := m.createModuleStructure(config.ModuleEventConsumer); err != nil {
		m.checkpoint.FailStage(StageEvents, err)
		return err
	}

	listeners := m.projectInfo.EventListeners
	total := len(listeners)

	for i, listener := range listeners {
		m.printProgress(i+1, total, listener.Name)

		converted, err := m.convertEventListener(ctx, listener)
		if err != nil {
			m.yellow.Printf("  Warning: failed to convert %s: %v\n", listener.Name, err)
			continue
		}

		if err := m.writeEventListenerFiles(converted); err != nil {
			m.yellow.Printf("  Warning: failed to write %s: %v\n", listener.Name, err)
		}
	}

	fmt.Println()
	m.green.Printf("  ✓ Converted %d event listeners\n", total)

	// End cost tracking phase
	m.costTracker.EndPhase()

	return m.checkpoint.CompleteStage(StageEvents, nil)
}

// runConfiguration generates configuration files
func (m *Migrator) runConfiguration(ctx context.Context) error {
	m.printStageHeader("Configuration", "Generating configuration files...")

	if err := m.checkpoint.StartStage(StageConfiguration); err != nil {
		return err
	}

	// Generate docker-compose.yml
	if err := m.generateDockerCompose(); err != nil {
		m.yellow.Printf("  Warning: failed to generate docker-compose.yml: %v\n", err)
	} else {
		m.green.Println("  ✓ Generated docker-compose.yml")
	}

	// Generate .env.example
	if err := m.generateEnvExample(); err != nil {
		m.yellow.Printf("  Warning: failed to generate .env.example: %v\n", err)
	} else {
		m.green.Println("  ✓ Generated .env.example")
	}

	// Generate AI agent files
	if err := m.generateAIAgentFiles(); err != nil {
		m.yellow.Printf("  Warning: failed to generate AI agent files: %v\n", err)
	} else {
		m.green.Println("  ✓ Generated AI agent context files")
	}

	return m.checkpoint.CompleteStage(StageConfiguration, nil)
}

// runFinalAssembly creates the final project structure
func (m *Migrator) runFinalAssembly(ctx context.Context) error {
	m.printStageHeader("Final Assembly", "Generating project files...")

	if err := m.checkpoint.StartStage(StageFinalAssembly); err != nil {
		return err
	}

	// Generate parent pom.xml
	if err := m.generateParentPOM(); err != nil {
		m.checkpoint.FailStage(StageFinalAssembly, err)
		return fmt.Errorf("failed to generate parent POM: %w", err)
	}
	m.green.Println("  ✓ Generated pom.xml")

	// Generate module pom.xml files
	if err := m.generateModulePOMs(); err != nil {
		m.checkpoint.FailStage(StageFinalAssembly, err)
		return fmt.Errorf("failed to generate module POMs: %w", err)
	}
	m.green.Println("  ✓ Generated module pom.xml files")

	// Generate README.md
	if err := m.generateREADME(); err != nil {
		m.yellow.Printf("  Warning: failed to generate README: %v\n", err)
	} else {
		m.green.Println("  ✓ Generated README.md")
	}

	// Generate .gitignore
	if err := m.generateGitignore(); err != nil {
		m.yellow.Printf("  Warning: failed to generate .gitignore: %v\n", err)
	} else {
		m.green.Println("  ✓ Generated .gitignore")
	}

	// Generate .trabuco.json metadata
	if err := m.generateMetadata(); err != nil {
		m.yellow.Printf("  Warning: failed to generate metadata: %v\n", err)
	} else {
		m.green.Println("  ✓ Generated .trabuco.json")
	}

	return m.checkpoint.CompleteStage(StageFinalAssembly, nil)
}

// HasCheckpoint returns true if a checkpoint exists
func (m *Migrator) HasCheckpoint() bool {
	return m.checkpoint.HasCheckpoint()
}

// Helper methods

func (m *Migrator) printStageHeader(name, description string) {
	fmt.Println()
	m.cyan.Printf("─── %s ───\n", name)
	fmt.Println(description)
	fmt.Println()
}

func (m *Migrator) printProgress(current, total int, name string) {
	fmt.Printf("  [%d/%d] %s\n", current, total, name)
}

func (m *Migrator) printDiscoverySummary(info *ProjectInfo) {
	fmt.Println()
	m.cyan.Println("Project Summary:")
	fmt.Printf("  Name:           %s\n", info.Name)
	fmt.Printf("  Type:           %s\n", info.ProjectType)
	fmt.Printf("  Java Version:   %s\n", info.JavaVersion)
	fmt.Printf("  Base Package:   %s\n", info.BasePackage)
	fmt.Println()

	m.cyan.Println("Detected Infrastructure:")
	// Database detection
	databaseDisplay := m.getDatabaseDisplayName(info.Database)
	if info.UsesNoSQL {
		fmt.Printf("  Database:       %s (NoSQL)\n", databaseDisplay)
	} else {
		fmt.Printf("  Database:       %s (SQL)\n", databaseDisplay)
	}
	// Message broker detection
	if info.MessageBroker != "" {
		brokerDisplay := info.MessageBroker
		if info.MessageBroker == "rabbitmq" {
			brokerDisplay = "RabbitMQ"
		} else if info.MessageBroker == "kafka" {
			brokerDisplay = "Kafka"
		}
		fmt.Printf("  Message Broker: %s\n", brokerDisplay)
	}
	// Cache detection
	if info.UsesRedis {
		fmt.Printf("  Cache:          Redis\n")
	}
	fmt.Println()

	m.cyan.Println("Source Structure:")
	fmt.Printf("  Entities:       %d classes\n", len(info.Entities))
	fmt.Printf("  Repositories:   %d interfaces\n", len(info.Repositories))
	fmt.Printf("  Services:       %d classes\n", len(info.Services))
	fmt.Printf("  Controllers:    %d classes\n", len(info.Controllers))
	if info.HasScheduledJobs {
		fmt.Printf("  Scheduled Jobs: %d classes\n", len(info.ScheduledJobs))
	}
	if info.HasEventListeners {
		fmt.Printf("  Event Listeners: %d classes\n", len(info.EventListeners))
	}
	fmt.Println()
}

// getDatabaseDisplayName returns a user-friendly name for the database type
func (m *Migrator) getDatabaseDisplayName(database string) string {
	switch database {
	case "postgresql":
		return "PostgreSQL"
	case "mysql":
		return "MySQL"
	case "mongodb":
		return "MongoDB"
	case "oracle":
		return "Oracle"
	case "sqlserver":
		return "SQL Server"
	case "h2":
		return "H2"
	default:
		return database
	}
}

func (m *Migrator) printDependencyReport(report *DependencyReport) {
	fmt.Println()
	m.cyan.Println("Dependency Analysis:")

	if len(report.Compatible) > 0 {
		m.green.Printf("  ✓ Compatible: %d dependencies\n", len(report.Compatible))
	}

	if len(report.Replaceable) > 0 {
		m.yellow.Printf("  ⚠ Replaceable: %d dependencies\n", len(report.Replaceable))
		for _, dep := range report.Replaceable {
			fmt.Printf("    - %s → %s\n", dep.Source, dep.TrabucoAlternative)
		}
	}

	if len(report.Unsupported) > 0 {
		m.red.Printf("  ✗ Unsupported: %d dependencies\n", len(report.Unsupported))
		for _, dep := range report.Unsupported {
			fmt.Printf("    - %s\n", dep)
		}
	}

	fmt.Println()
}

func (m *Migrator) printCostEstimate() {
	// Estimate tokens based on project size
	totalFiles := len(m.projectInfo.Entities) + len(m.projectInfo.Repositories) +
		len(m.projectInfo.Services) + len(m.projectInfo.Controllers)

	// Rough estimate: 500 tokens per file input, 1000 tokens output
	estimatedInputTokens := totalFiles * 500
	estimatedOutputTokens := totalFiles * 1000

	estimatedCost := m.provider.EstimateCost(estimatedInputTokens, estimatedOutputTokens)

	fmt.Println()
	m.cyan.Println("Estimated Migration Cost:")
	fmt.Printf("  Files to process: %d\n", totalFiles)
	fmt.Printf("  Est. tokens:      ~%d\n", estimatedInputTokens+estimatedOutputTokens)
	fmt.Printf("  Est. cost:        $%.2f - $%.2f\n", estimatedCost*0.8, estimatedCost*1.5)
	fmt.Println()
}

func (m *Migrator) printDryRunSummary() {
	fmt.Println()
	m.cyan.Println("═══════════════════════════════════════")
	m.cyan.Println("         DRY RUN SUMMARY")
	m.cyan.Println("═══════════════════════════════════════")
	fmt.Println()

	fmt.Println("Migration would create:")
	fmt.Printf("  Output directory: %s\n", m.config.OutputPath)
	fmt.Println()

	fmt.Println("Modules to create:")
	fmt.Println("  - Model")
	if m.projectInfo.UsesNoSQL {
		fmt.Println("  - NoSQLDatastore")
	} else {
		fmt.Println("  - SQLDatastore")
	}
	fmt.Println("  - Shared")
	fmt.Println("  - API")
	if m.projectInfo.HasScheduledJobs {
		fmt.Println("  - Worker")
	}
	if m.projectInfo.HasEventListeners {
		fmt.Println("  - EventConsumer")
	}
	fmt.Println()

	fmt.Println("Files to migrate:")
	fmt.Printf("  Entities:     %d → Model/src/main/java/.../model/entities/\n", len(m.projectInfo.Entities))
	fmt.Printf("  Repositories: %d → [SQL/NoSQL]Datastore/src/main/java/.../repository/\n", len(m.projectInfo.Repositories))
	fmt.Printf("  Services:     %d → Shared/src/main/java/.../shared/service/\n", len(m.projectInfo.Services))
	fmt.Printf("  Controllers:  %d → API/src/main/java/.../api/controller/\n", len(m.projectInfo.Controllers))

	if m.projectInfo.HasScheduledJobs {
		fmt.Printf("  Jobs:         %d → Worker/src/main/java/.../worker/handler/\n", len(m.projectInfo.ScheduledJobs))
	}
	if m.projectInfo.HasEventListeners {
		fmt.Printf("  Listeners:    %d → EventConsumer/src/main/java/.../eventconsumer/listener/\n", len(m.projectInfo.EventListeners))
	}
	fmt.Println()
}

func (m *Migrator) confirmContinue() bool {
	fmt.Print("Continue? [y/N] ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(strings.TrimSpace(response)) == "y"
}

func (m *Migrator) confirmDependencyReplacements(report *DependencyReport) error {
	for _, dep := range report.Replaceable {
		fmt.Println()
		m.yellow.Printf("Dependency: %s\n", dep.Source)
		fmt.Printf("  Trabuco uses: %s\n", dep.TrabucoAlternative)
		fmt.Printf("  Impact: %s\n", dep.MigrationImpact)
		fmt.Print("  [A]ccept / [S]kip / [C]ancel: ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		switch response {
		case "a", "accept", "":
			dep.Accepted = true
		case "s", "skip":
			dep.Accepted = false
		case "c", "cancel":
			return fmt.Errorf("migration cancelled by user")
		}
	}
	return nil
}

func (m *Migrator) createModuleStructure(module string) error {
	packagePath := strings.ReplaceAll(m.projectInfo.GroupID, ".", string(filepath.Separator))

	// Determine subpackage based on module
	var subPackage string
	switch module {
	case config.ModuleModel:
		subPackage = "model"
	case config.ModuleSQLDatastore:
		subPackage = "sqldatastore"
	case config.ModuleNoSQLDatastore:
		subPackage = "nosqldatastore"
	case config.ModuleShared:
		subPackage = "shared"
	case config.ModuleAPI:
		subPackage = "api"
	case config.ModuleWorker:
		subPackage = "worker"
	case config.ModuleEventConsumer:
		subPackage = "eventconsumer"
	default:
		subPackage = strings.ToLower(module)
	}

	// Create main source directory
	mainJavaDir := filepath.Join(m.config.OutputPath, module, "src", "main", "java", packagePath, subPackage)
	if err := os.MkdirAll(mainJavaDir, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Create main resources directory
	mainResourcesDir := filepath.Join(m.config.OutputPath, module, "src", "main", "resources")
	if err := os.MkdirAll(mainResourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create resources directory: %w", err)
	}

	// Create test directories
	testJavaDir := filepath.Join(m.config.OutputPath, module, "src", "test", "java", packagePath, subPackage)
	if err := os.MkdirAll(testJavaDir, 0755); err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}

	testResourcesDir := filepath.Join(m.config.OutputPath, module, "src", "test", "resources")
	if err := os.MkdirAll(testResourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create test resources directory: %w", err)
	}

	return nil
}

func (m *Migrator) buildProject() error {
	m.cyan.Println("\nBuilding project with Maven...")

	// Create spinner animation
	done := make(chan bool)
	go func() {
		frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
		i := 0
		for {
			select {
			case <-done:
				fmt.Print("\r")
				return
			default:
				fmt.Printf("\r  %s Running mvn clean install -DskipTests...", frames[i%len(frames)])
				i++
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Run Maven build
	err := utils.RunMavenBuild(m.config.OutputPath)

	done <- true
	time.Sleep(50 * time.Millisecond)

	if err != nil {
		return err
	}

	fmt.Printf("\r                                                    \r")
	m.green.Println("✓ Maven build completed successfully!")

	return nil
}

// Placeholder methods - implemented in separate files:
// - entity_migration.go: convertEntity, writeEntityFiles, generateFlywayMigrations
// - service_migration.go: convertRepository, writeRepositoryFiles, convertService, writeServiceFiles
// - controller_migration.go: convertController, writeControllerFiles
// - job_migration.go: convertJob, writeJobFiles
// - event_migration.go: convertEventListener, writeEventListenerFiles
// - config_generation.go: generateDockerCompose, generateEnvExample, generateAIAgentFiles,
//                         generateParentPOM, generateREADME, generateGitignore, generateMetadata

// Converted types
type ConvertedEntity struct {
	Name            string
	EntityCode      string
	DTOCode         string
	ResponseCode    string
	FlywayMigration string
	Notes           []string
	Skip            bool // True if this class should be skipped (not a real entity)
}

type ConvertedRepository struct {
	Name           string
	RepositoryCode string
}

type ConvertedService struct {
	Name        string
	ServiceCode string
	TestCode    string
}

type ConvertedController struct {
	Name           string
	ControllerCode string
}

type ConvertedJob struct {
	Name            string
	JobRequestCode  string
	JobHandlerCode  string
}

type ConvertedEventListener struct {
	Name         string
	EventCode    string
	ListenerCode string
}
