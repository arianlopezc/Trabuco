package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/templates"
	"github.com/fatih/color"
)

// Version constants for dependencies added during module addition
const (
	JobRunrVersion           = "8.4.0"
	LogstashEncoderVersion   = "8.0"
	SpringDocVersion         = "2.7.0"
	JaCoCoVersion            = "0.8.12"
	SpringCloudAWSVersion    = "3.2.0"
	SpringCloudGCPVersion    = "5.8.0"
	LocalStackImageVersion   = "3.0"
	ConfluentKafkaVersion    = "7.5.0"
	MCPSDKVersion            = "0.17.2"
	EnforcerVersion          = "3.5.0"
	SpotlessVersion          = "2.44.4"
	ArchUnitVersion          = "1.4.0"
)

// Module, database, and broker constants are defined in config package
// Use config.ModuleModel, config.DatabasePostgreSQL, config.BrokerKafka, etc.

// ModuleAdder handles adding modules to an existing Trabuco project
type ModuleAdder struct {
	projectPath string
	metadata    *config.ProjectMetadata
	config      *config.ProjectConfig
	engine      *templates.Engine
	backup      *BackupManager
	version     string
}

// NewModuleAdder creates a new ModuleAdder
func NewModuleAdder(projectPath string, metadata *config.ProjectMetadata, version string, enableBackup bool) *ModuleAdder {
	cfg := metadata.ToProjectConfig()

	return &ModuleAdder{
		projectPath: projectPath,
		metadata:    metadata,
		config:      cfg,
		engine:      templates.NewEngine(),
		backup:      NewBackupManager(projectPath, enableBackup),
		version:     version,
	}
}

// Add adds a module and its dependencies to the project
func (a *ModuleAdder) Add(module string, database, nosqlDatabase, messageBroker string) error {
	green := color.New(color.FgGreen)

	// Validate module can be added
	if err := a.ValidateCanAdd(module); err != nil {
		return err
	}

	// Validate options for specific modules
	if err := a.validateOptions(module, database, nosqlDatabase, messageBroker); err != nil {
		return err
	}

	// Resolve dependencies
	dependencies := a.ResolveDependencies(module)
	// Create a new slice to avoid modifying the original dependencies slice
	allModules := make([]string, 0, len(dependencies)+1)
	allModules = append(allModules, dependencies...)
	allModules = append(allModules, module)

	// Backup existing files
	filesToBackup := GetFilesToBackup(module)
	if err := a.backup.BackupAll(filesToBackup); err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// Update config with new options
	a.updateConfig(module, database, nosqlDatabase, messageBroker)

	// Add each module
	for _, mod := range allModules {
		if err := a.addModule(mod); err != nil {
			// Restore on failure
			if restoreErr := a.backup.Restore(); restoreErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to restore backup: %v\n", restoreErr)
				a.backup.PrintRestoreInstructions()
			}
			return fmt.Errorf("failed to add %s: %w", mod, err)
		}
		green.Printf("  \u2713 Created %s module\n", mod)
	}

	// Update parent POM (modules and properties)
	if err := a.updateParentPOM(allModules, messageBroker); err != nil {
		return fmt.Errorf("failed to update parent POM: %w", err)
	}
	green.Println("  \u2713 Updated pom.xml")

	// Update docker-compose if needed
	if err := a.updateDockerCompose(module, database, nosqlDatabase, messageBroker); err != nil {
		return fmt.Errorf("failed to update docker-compose: %w", err)
	}

	// Update Model module if needed
	if err := a.updateModelModule(module); err != nil {
		return fmt.Errorf("failed to update Model module: %w", err)
	}

	// Update Shared module if needed (when adding datastore)
	if err := a.updateSharedModule(module); err != nil {
		return fmt.Errorf("failed to update Shared module: %w", err)
	}

	// Update API module if needed (to include new packages in ComponentScan)
	if err := a.updateAPIModule(module); err != nil {
		return fmt.Errorf("failed to update API module: %w", err)
	}

	// Update metadata
	a.updateMetadata(allModules, database, nosqlDatabase, messageBroker)
	if err := config.SaveMetadata(a.projectPath, a.metadata); err != nil {
		return fmt.Errorf("failed to update metadata: %w", err)
	}
	green.Println("  ✓ Updated .trabuco.json")

	// Regenerate documentation files (README.md and AI agent files)
	if err := a.regenerateDocs(); err != nil {
		return fmt.Errorf("failed to regenerate documentation: %w", err)
	}
	green.Println("  ✓ Updated documentation files")

	// Cleanup old backups first, then current backup after successful operation
	if err := a.backup.CleanupOldBackups(); err != nil {
		// Non-fatal warning
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup old backups: %v\n", err)
	}
	if err := a.backup.Cleanup(); err != nil {
		// Non-fatal warning
		fmt.Fprintf(os.Stderr, "Warning: failed to cleanup backup: %v\n", err)
	}

	return nil
}

// ValidateCanAdd checks if a module can be added
func (a *ModuleAdder) ValidateCanAdd(module string) error {
	// Check if module already exists
	if a.metadata.HasModule(module) {
		return fmt.Errorf("module %s already exists in this project", module)
	}

	// Check mutual exclusion
	if module == config.ModuleSQLDatastore && a.metadata.HasModule(config.ModuleNoSQLDatastore) {
		return fmt.Errorf("cannot add %s: %s already exists (mutually exclusive)", config.ModuleSQLDatastore, config.ModuleNoSQLDatastore)
	}
	if module == config.ModuleNoSQLDatastore && a.metadata.HasModule(config.ModuleSQLDatastore) {
		return fmt.Errorf("cannot add %s: %s already exists (mutually exclusive)", config.ModuleNoSQLDatastore, config.ModuleSQLDatastore)
	}

	// Check if valid module
	m := config.GetModule(module)
	if m == nil {
		return fmt.Errorf("unknown module: %s", module)
	}

	// Check if internal module
	if m.Internal {
		return fmt.Errorf("cannot add %s directly: it's automatically included", module)
	}

	return nil
}

// ResolveDependencies returns dependencies that need to be added
func (a *ModuleAdder) ResolveDependencies(module string) []string {
	var deps []string
	m := config.GetModule(module)
	if m == nil {
		return deps
	}

	for _, dep := range m.Dependencies {
		// Only add if not already present
		if !a.metadata.HasModule(dep) {
			deps = append(deps, dep)
		}
	}

	return deps
}

// DryRun returns what would be changed without actually making changes
func (a *ModuleAdder) DryRun(module string) *DryRunResult {
	result := &DryRunResult{
		Module:       module,
		Dependencies: a.ResolveDependencies(module),
	}

	// Files that would be created
	allModules := make([]string, 0, len(result.Dependencies)+1)
	allModules = append(allModules, result.Dependencies...)
	allModules = append(allModules, module)
	for _, mod := range allModules {
		result.FilesCreated = append(result.FilesCreated, a.getModuleFiles(mod)...)
	}

	// Files that would be modified
	result.FilesModified = append(result.FilesModified, "pom.xml", ".trabuco.json", "README.md")
	if needsDockerComposeUpdate(module) {
		result.FilesModified = append(result.FilesModified, "docker-compose.yml")
	}
	// Add AI agent files that are present
	for _, agent := range a.config.GetSelectedAIAgents() {
		result.FilesModified = append(result.FilesModified, agent.FilePath)
	}
	// Add CI workflow if configured
	if a.config.HasCIProvider("github") {
		result.FilesModified = append(result.FilesModified, ".github/workflows/ci.yml")
	}

	return result
}

// DryRunResult represents what would happen during an add operation
type DryRunResult struct {
	Module        string
	Dependencies  []string
	FilesCreated  []string
	FilesModified []string
}

// Print prints the dry run result
func (d *DryRunResult) Print() {
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	fmt.Println()
	cyan.Println("Dry Run Results:")
	fmt.Println()

	fmt.Printf("Module to add: %s\n", d.Module)
	if len(d.Dependencies) > 0 {
		fmt.Printf("Dependencies: %s\n", strings.Join(d.Dependencies, ", "))
	}

	fmt.Println()
	yellow.Println("Files that would be created:")
	for _, f := range d.FilesCreated {
		fmt.Printf("  + %s\n", f)
	}

	fmt.Println()
	yellow.Println("Files that would be modified:")
	for _, f := range d.FilesModified {
		fmt.Printf("  ~ %s\n", f)
	}
}

// validateOptions validates the options provided for a module
func (a *ModuleAdder) validateOptions(module, database, nosqlDatabase, messageBroker string) error {
	switch module {
	case config.ModuleSQLDatastore:
		if database != "" && database != config.DatabasePostgreSQL && database != config.DatabaseMySQL {
			return fmt.Errorf("invalid database type: %s (must be '%s' or '%s')", database, config.DatabasePostgreSQL, config.DatabaseMySQL)
		}
	case config.ModuleNoSQLDatastore:
		if nosqlDatabase != "" && nosqlDatabase != config.DatabaseMongoDB && nosqlDatabase != config.DatabaseRedis {
			return fmt.Errorf("invalid NoSQL database type: %s (must be '%s' or '%s')", nosqlDatabase, config.DatabaseMongoDB, config.DatabaseRedis)
		}
	case config.ModuleEventConsumer:
		if messageBroker != "" && messageBroker != config.BrokerKafka && messageBroker != config.BrokerRabbitMQ && messageBroker != config.BrokerSQS && messageBroker != config.BrokerPubSub {
			return fmt.Errorf("invalid message broker: %s (must be '%s', '%s', '%s', or '%s')", messageBroker, config.BrokerKafka, config.BrokerRabbitMQ, config.BrokerSQS, config.BrokerPubSub)
		}
	}
	return nil
}

// updateConfig updates the config with new options
func (a *ModuleAdder) updateConfig(module, database, nosqlDatabase, messageBroker string) {
	if module == config.ModuleSQLDatastore && database != "" {
		a.config.Database = database
	}
	if module == config.ModuleNoSQLDatastore && nosqlDatabase != "" {
		a.config.NoSQLDatabase = nosqlDatabase
	}
	if module == config.ModuleEventConsumer && messageBroker != "" {
		a.config.MessageBroker = messageBroker
	}

	// Add all modules that will be added
	deps := a.ResolveDependencies(module)
	for _, dep := range deps {
		if !a.config.HasModule(dep) {
			a.config.Modules = append(a.config.Modules, dep)
		}
	}
	if !a.config.HasModule(module) {
		a.config.Modules = append(a.config.Modules, module)
	}
}

// updateMetadata updates the metadata after adding modules
func (a *ModuleAdder) updateMetadata(modules []string, database, nosqlDatabase, messageBroker string) {
	for _, mod := range modules {
		a.metadata.AddModule(mod)
	}
	if database != "" {
		a.metadata.Database = database
	}
	if nosqlDatabase != "" {
		a.metadata.NoSQLDatabase = nosqlDatabase
	}
	if messageBroker != "" {
		a.metadata.MessageBroker = messageBroker
	}
	a.metadata.UpdateGeneratedAt()
}

// addModule adds a single module's files
func (a *ModuleAdder) addModule(module string) error {
	// Create directories
	if err := a.createModuleDirectories(module); err != nil {
		return err
	}

	// Generate module files using the existing generator logic
	gen := &Generator{
		config: a.config,
		engine: a.engine,
		outDir: a.projectPath,
	}

	return gen.generateModule(module)
}

// createModuleDirectories creates the directory structure for a module
// and tracks the module root directory for rollback purposes
func (a *ModuleAdder) createModuleDirectories(module string) error {
	packagePath := a.config.PackagePath()

	// Track the module root directory for rollback
	moduleRoot := filepath.Join(a.projectPath, module)
	if _, err := os.Stat(moduleRoot); os.IsNotExist(err) {
		// Only track if directory doesn't already exist
		a.backup.TrackCreatedDir(moduleRoot)
	}

	var dirs []string
	switch module {
	case config.ModuleModel:
		modelBase := filepath.Join(a.projectPath, config.ModuleModel, "src", "main", "java", packagePath, "model")
		dirs = []string{
			modelBase,
			filepath.Join(modelBase, "entities"),
			filepath.Join(modelBase, "dto"),
			filepath.Join(modelBase, "enums"),
			filepath.Join(modelBase, "exception"),
			filepath.Join(modelBase, "util"),
			filepath.Join(modelBase, "events"),
			filepath.Join(modelBase, "jobs"),
			filepath.Join(modelBase, "validation"),
		}
	case config.ModuleSQLDatastore:
		sqlBase := filepath.Join(a.projectPath, config.ModuleSQLDatastore, "src", "main", "java", packagePath, "sqldatastore")
		sqlTestBase := filepath.Join(a.projectPath, config.ModuleSQLDatastore, "src", "test", "java", packagePath, "sqldatastore")
		dirs = []string{
			filepath.Join(sqlBase, "config"),
			filepath.Join(sqlBase, "repository"),
			filepath.Join(a.projectPath, config.ModuleSQLDatastore, "src", "main", "resources", "db", "migration"),
			filepath.Join(sqlTestBase, "repository"),
		}
	case config.ModuleNoSQLDatastore:
		nosqlBase := filepath.Join(a.projectPath, config.ModuleNoSQLDatastore, "src", "main", "java", packagePath, "nosqldatastore")
		nosqlTestBase := filepath.Join(a.projectPath, config.ModuleNoSQLDatastore, "src", "test", "java", packagePath, "nosqldatastore")
		dirs = []string{
			filepath.Join(nosqlBase, "config"),
			filepath.Join(nosqlBase, "repository"),
			filepath.Join(a.projectPath, config.ModuleNoSQLDatastore, "src", "main", "resources"),
			filepath.Join(nosqlTestBase, "repository"),
		}
	case config.ModuleShared:
		sharedBase := filepath.Join(a.projectPath, config.ModuleShared, "src", "main", "java", packagePath, "shared")
		sharedTestBase := filepath.Join(a.projectPath, config.ModuleShared, "src", "test", "java", packagePath, "shared")
		dirs = []string{
			filepath.Join(sharedBase, "config"),
			filepath.Join(sharedBase, "service"),
			filepath.Join(a.projectPath, config.ModuleShared, "src", "main", "resources"),
			filepath.Join(sharedTestBase, "service"),
		}
	case config.ModuleAPI:
		apiBase := filepath.Join(a.projectPath, config.ModuleAPI, "src", "main", "java", packagePath, "api")
		dirs = []string{
			apiBase,
			filepath.Join(apiBase, "controller"),
			filepath.Join(apiBase, "config"),
			filepath.Join(a.projectPath, config.ModuleAPI, "src", "main", "resources"),
			filepath.Join(a.projectPath, ".run"),
		}
	case config.ModuleJobs:
		jobsBase := filepath.Join(a.projectPath, config.ModuleJobs, "src", "main", "java", packagePath, "jobs")
		dirs = []string{jobsBase}
	case config.ModuleWorker:
		workerBase := filepath.Join(a.projectPath, config.ModuleWorker, "src", "main", "java", packagePath, "worker")
		workerTestBase := filepath.Join(a.projectPath, config.ModuleWorker, "src", "test", "java", packagePath, "worker")
		dirs = []string{
			workerBase,
			filepath.Join(workerBase, "config"),
			filepath.Join(workerBase, "handler"),
			filepath.Join(a.projectPath, config.ModuleWorker, "src", "main", "resources"),
			filepath.Join(workerTestBase, "handler"),
			filepath.Join(a.projectPath, ".run"),
		}
	case config.ModuleEvents:
		eventsBase := filepath.Join(a.projectPath, config.ModuleEvents, "src", "main", "java", packagePath, "events")
		dirs = []string{
			eventsBase,
			filepath.Join(eventsBase, "config"),
		}
	case config.ModuleEventConsumer:
		ecBase := filepath.Join(a.projectPath, config.ModuleEventConsumer, "src", "main", "java", packagePath, "eventconsumer")
		ecTestBase := filepath.Join(a.projectPath, config.ModuleEventConsumer, "src", "test", "java", packagePath, "eventconsumer")
		dirs = []string{
			ecBase,
			filepath.Join(ecBase, "config"),
			filepath.Join(ecBase, "listener"),
			filepath.Join(a.projectPath, config.ModuleEventConsumer, "src", "main", "resources"),
			filepath.Join(ecTestBase, "listener"),
		}
	case config.ModuleMCP:
		mcpBase := filepath.Join(a.projectPath, config.ModuleMCP, "src", "main", "java", packagePath, "mcp")
		dirs = []string{
			mcpBase,
			filepath.Join(mcpBase, "tools"),
			filepath.Join(a.projectPath, config.ModuleMCP, "src", "main", "resources"),
		}
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// updateDockerCompose updates docker-compose.yml with required services
func (a *ModuleAdder) updateDockerCompose(module, database, nosqlDatabase, messageBroker string) error {
	if !needsDockerComposeUpdate(module) {
		return nil
	}

	composePath := filepath.Join(a.projectPath, "docker-compose.yml")
	updater, err := NewDockerComposeUpdater(composePath)
	if err != nil {
		return err
	}

	switch module {
	case config.ModuleSQLDatastore:
		if database == config.DatabasePostgreSQL && !updater.HasService("postgres") {
			// PostgreSQL supports hyphens in database names
			dbName := a.config.ProjectName
			// Use postgres/postgres credentials to match application.yml template defaults
			// Port 5433 to avoid conflicts with local PostgreSQL
			updater.AddService("postgres", GetPostgresService(
				"postgres",
				dbName,
				"postgres",
				"postgres",
				5433,
			))
			updater.AddVolume("postgres-data")
		} else if database == config.DatabaseMySQL && !updater.HasService("mysql") {
			// MySQL doesn't support hyphens in database names, use snake_case
			dbName := a.config.ProjectNameSnake()
			// Use root/root credentials to match application.yml template defaults
			updater.AddService("mysql", GetMySQLService("mysql", dbName, "root"))
			updater.AddVolume("mysql-data")
		}

	case config.ModuleNoSQLDatastore:
		// Use original project name to match application.yml template defaults
		dbName := a.config.ProjectName
		if nosqlDatabase == config.DatabaseMongoDB && !updater.HasService("mongodb") {
			// No authentication for local development (matches docker-compose template)
			updater.AddService("mongodb", GetMongoDBService("mongodb", dbName))
			updater.AddVolume("mongodb-data")
		} else if nosqlDatabase == config.DatabaseRedis && !updater.HasService("redis") {
			updater.AddService("redis", GetRedisService("redis"))
			updater.AddVolume("redis-data")
		}

	case config.ModuleWorker:
		// Check if Worker needs its own PostgreSQL for JobRunr
		if a.config.WorkerNeedsOwnPostgres() && !updater.HasService("postgres-jobrunr") {
			// Use original project name to match application.yml template defaults
			dbName := a.config.ProjectName + "_jobs"
			// Use postgres/postgres credentials to match application.yml template defaults
			// Port 5434 to differentiate from main database
			updater.AddService("postgres-jobrunr", GetPostgresService(
				"postgres-jobrunr",
				dbName,
				"postgres",
				"postgres",
				5434,
			))
			updater.AddVolume("postgres-jobrunr-data")
		}

	case config.ModuleEventConsumer:
		switch messageBroker {
		case config.BrokerKafka:
			if !updater.HasService("kafka") {
				kafka, zookeeper := GetKafkaService()
				updater.AddService("zookeeper", zookeeper)
				updater.AddService("kafka", kafka)
			}
		case config.BrokerRabbitMQ:
			if !updater.HasService("rabbitmq") {
				// Use guest/guest credentials to match application.yml template defaults
				updater.AddService("rabbitmq", GetRabbitMQService("guest", "guest"))
				updater.AddVolume("rabbitmq-data")
			}
		case config.BrokerSQS:
			if !updater.HasService("localstack") {
				updater.AddService("localstack", GetLocalStackService())
				// Also create the SQS init script
				if err := a.createSQSInitScript(); err != nil {
					return err
				}
			}
		case config.BrokerPubSub:
			if !updater.HasService("pubsub-emulator") {
				updater.AddService("pubsub-emulator", GetPubSubEmulatorService())
			}
		}
	}

	return updater.Save()
}

// updateParentPOM updates the parent pom.xml with modules and required properties/BOMs
func (a *ModuleAdder) updateParentPOM(modules []string, messageBroker string) error {
	pomPath := filepath.Join(a.projectPath, "pom.xml")
	updater, err := NewPOMUpdater(pomPath)
	if err != nil {
		return err
	}

	// Add modules
	for _, mod := range modules {
		if err := updater.AddModule(mod); err != nil {
			return fmt.Errorf("failed to add module %s: %w", mod, err)
		}
	}

	// Add required properties based on modules being added
	for _, mod := range modules {
		switch mod {
		case config.ModuleJobs, config.ModuleWorker:
			// JobRunr version property needed for Jobs and Worker modules
			if err := updater.AddProperty("jobrunr.version", JobRunrVersion); err != nil {
				return fmt.Errorf("failed to add jobrunr.version property: %w", err)
			}
			// logstash-logback-encoder for structured logging in Worker
			if mod == config.ModuleWorker {
				if err := updater.AddProperty("logstash-logback-encoder.version", LogstashEncoderVersion); err != nil {
					return fmt.Errorf("failed to add logstash-logback-encoder.version property: %w", err)
				}
			}
		case config.ModuleAPI:
			// logstash-logback-encoder for structured logging
			if err := updater.AddProperty("logstash-logback-encoder.version", LogstashEncoderVersion); err != nil {
				return fmt.Errorf("failed to add logstash-logback-encoder.version property: %w", err)
			}
			// springdoc for OpenAPI/Swagger
			if err := updater.AddProperty("springdoc.version", SpringDocVersion); err != nil {
				return fmt.Errorf("failed to add springdoc.version property: %w", err)
			}
			// jacoco for test coverage
			if err := updater.AddProperty("jacoco.version", JaCoCoVersion); err != nil {
				return fmt.Errorf("failed to add jacoco.version property: %w", err)
			}
		case config.ModuleEventConsumer:
			// logstash-logback-encoder for structured logging
			if err := updater.AddProperty("logstash-logback-encoder.version", LogstashEncoderVersion); err != nil {
				return fmt.Errorf("failed to add logstash-logback-encoder.version property: %w", err)
			}
		case config.ModuleMCP:
			// MCP SDK version for AI tool integration
			if err := updater.AddProperty("mcp-sdk.version", MCPSDKVersion); err != nil {
				return fmt.Errorf("failed to add mcp-sdk.version property: %w", err)
			}
		case config.ModuleShared:
			// Quality plugin versions (Enforcer, Spotless, ArchUnit)
			if err := updater.AddProperty("maven-enforcer.version", EnforcerVersion); err != nil {
				return fmt.Errorf("failed to add maven-enforcer.version property: %w", err)
			}
			if err := updater.AddProperty("spotless.version", SpotlessVersion); err != nil {
				return fmt.Errorf("failed to add spotless.version property: %w", err)
			}
			if err := updater.AddProperty("archunit.version", ArchUnitVersion); err != nil {
				return fmt.Errorf("failed to add archunit.version property: %w", err)
			}
		}
	}

	// Add required BOMs for message brokers
	if messageBroker == config.BrokerSQS {
		// Spring Cloud AWS BOM for SQS
		if err := updater.AddDependencyManagement(
			"io.awspring.cloud",
			"spring-cloud-aws-dependencies",
			SpringCloudAWSVersion,
			"pom",
			"import",
		); err != nil {
			return fmt.Errorf("failed to add Spring Cloud AWS BOM: %w", err)
		}
	} else if messageBroker == config.BrokerPubSub {
		// Spring Cloud GCP BOM for Pub/Sub
		if err := updater.AddDependencyManagement(
			"com.google.cloud",
			"spring-cloud-gcp-dependencies",
			SpringCloudGCPVersion,
			"pom",
			"import",
		); err != nil {
			return fmt.Errorf("failed to add Spring Cloud GCP BOM: %w", err)
		}
	}

	return updater.Save()
}

// createSQSInitScript creates the LocalStack SQS initialization script
func (a *ModuleAdder) createSQSInitScript() error {
	localstackDir := filepath.Join(a.projectPath, "localstack-init")
	scriptDir := filepath.Join(localstackDir, "ready.d")

	// Track the localstack-init directory for rollback if it doesn't exist
	if _, err := os.Stat(localstackDir); os.IsNotExist(err) {
		a.backup.TrackCreatedDir(localstackDir)
	}

	if err := os.MkdirAll(scriptDir, 0755); err != nil {
		return err
	}

	scriptPath := filepath.Join(scriptDir, "init-sqs.sh")
	// Queue name matches the application.yml template default: placeholder-events
	content := `#!/bin/bash
# Create SQS queues for local development
awslocal sqs create-queue --queue-name placeholder-events
echo "SQS queues created successfully"
`

	return os.WriteFile(scriptPath, []byte(content), 0755)
}

// updateModelModule adds new files to Model module when needed
func (a *ModuleAdder) updateModelModule(module string) error {
	gen := &Generator{
		config: a.config,
		engine: a.engine,
		outDir: a.projectPath,
	}

	switch module {
	case config.ModuleSQLDatastore:
		// Add spring-data-relational dependency to Model pom.xml
		modelPomPath := filepath.Join(a.projectPath, config.ModuleModel, "pom.xml")
		modelPom, err := NewPOMUpdater(modelPomPath)
		if err != nil {
			return fmt.Errorf("failed to read Model pom.xml: %w", err)
		}
		if err := modelPom.AddDependency("org.springframework.data", "spring-data-relational", ""); err != nil {
			return fmt.Errorf("failed to add spring-data-relational dependency to Model: %w", err)
		}
		if err := modelPom.Save(); err != nil {
			return fmt.Errorf("failed to save Model pom.xml: %w", err)
		}
		color.New(color.FgGreen).Println("  ✓ Added spring-data-relational dependency to Model")

		// Backup and regenerate Placeholder.java with SQL id field
		placeholderPath := gen.javaPath(config.ModuleModel, filepath.Join("entities", "Placeholder.java"))
		if err := a.backup.Backup(placeholderPath); err != nil {
			return fmt.Errorf("failed to backup Placeholder.java: %w", err)
		}
		if err := gen.writeTemplate(
			"java/model/entities/Placeholder.java.tmpl",
			placeholderPath,
		); err != nil {
			return err
		}
		color.New(color.FgGreen).Println("  ✓ Updated Placeholder.java with SQL id field")

		// Backup and regenerate PlaceholderResponse.java with SQL id field
		responsePath := gen.javaPath(config.ModuleModel, filepath.Join("dto", "PlaceholderResponse.java"))
		if err := a.backup.Backup(responsePath); err != nil {
			return fmt.Errorf("failed to backup PlaceholderResponse.java: %w", err)
		}
		if err := gen.writeTemplate(
			"java/model/dto/PlaceholderResponse.java.tmpl",
			responsePath,
		); err != nil {
			return err
		}
		color.New(color.FgGreen).Println("  ✓ Updated PlaceholderResponse.java with SQL id field")

		// Add PlaceholderRecord.java if not exists
		recordPath := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, filepath.Join("entities", "PlaceholderRecord.java")))
		if _, err := os.Stat(recordPath); os.IsNotExist(err) {
			if err := gen.writeTemplate(
				"java/model/entities/PlaceholderRecord.java.tmpl",
				gen.javaPath(config.ModuleModel, filepath.Join("entities", "PlaceholderRecord.java")),
			); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("  ✓ Added PlaceholderRecord.java to Model")
		}

	case config.ModuleNoSQLDatastore:
		// Add NoSQL dependency to Model pom.xml based on database type
		modelPomPath := filepath.Join(a.projectPath, config.ModuleModel, "pom.xml")
		modelPom, err := NewPOMUpdater(modelPomPath)
		if err != nil {
			return fmt.Errorf("failed to read Model pom.xml: %w", err)
		}

		nosqlDB := a.config.NoSQLDatabase
		switch nosqlDB {
		case config.DatabaseMongoDB:
			if err := modelPom.AddDependency("org.springframework.data", "spring-data-mongodb", ""); err != nil {
				return fmt.Errorf("failed to add spring-data-mongodb dependency to Model: %w", err)
			}
			color.New(color.FgGreen).Println("  ✓ Added spring-data-mongodb dependency to Model")
		case config.DatabaseRedis:
			if err := modelPom.AddDependency("org.springframework.data", "spring-data-redis", ""); err != nil {
				return fmt.Errorf("failed to add spring-data-redis dependency to Model: %w", err)
			}
			color.New(color.FgGreen).Println("  ✓ Added spring-data-redis dependency to Model")
		}

		if err := modelPom.Save(); err != nil {
			return fmt.Errorf("failed to save Model pom.xml: %w", err)
		}

		// Backup and regenerate Placeholder.java with NoSQL documentId field
		placeholderPath := gen.javaPath(config.ModuleModel, filepath.Join("entities", "Placeholder.java"))
		if err := a.backup.Backup(placeholderPath); err != nil {
			return fmt.Errorf("failed to backup Placeholder.java: %w", err)
		}
		if err := gen.writeTemplate(
			"java/model/entities/Placeholder.java.tmpl",
			placeholderPath,
		); err != nil {
			return err
		}
		color.New(color.FgGreen).Println("  ✓ Updated Placeholder.java with NoSQL documentId field")

		// Backup and regenerate PlaceholderResponse.java with NoSQL documentId field
		responsePath := gen.javaPath(config.ModuleModel, filepath.Join("dto", "PlaceholderResponse.java"))
		if err := a.backup.Backup(responsePath); err != nil {
			return fmt.Errorf("failed to backup PlaceholderResponse.java: %w", err)
		}
		if err := gen.writeTemplate(
			"java/model/dto/PlaceholderResponse.java.tmpl",
			responsePath,
		); err != nil {
			return err
		}
		color.New(color.FgGreen).Println("  ✓ Updated PlaceholderResponse.java with NoSQL documentId field")

		// Add PlaceholderDocument.java if not exists
		docPath := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, filepath.Join("entities", "PlaceholderDocument.java")))
		if _, err := os.Stat(docPath); os.IsNotExist(err) {
			if err := gen.writeTemplate(
				"java/model/entities/PlaceholderDocument.java.tmpl",
				gen.javaPath(config.ModuleModel, filepath.Join("entities", "PlaceholderDocument.java")),
			); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("  ✓ Added PlaceholderDocument.java to Model")
		}

	case config.ModuleWorker:
		// Add JobRunr dependency to Model pom.xml
		modelPomPath := filepath.Join(a.projectPath, config.ModuleModel, "pom.xml")
		modelPom, err := NewPOMUpdater(modelPomPath)
		if err != nil {
			return fmt.Errorf("failed to read Model pom.xml: %w", err)
		}
		if err := modelPom.AddDependency("org.jobrunr", "jobrunr", "${jobrunr.version}"); err != nil {
			return fmt.Errorf("failed to add JobRunr dependency to Model: %w", err)
		}
		if err := modelPom.Save(); err != nil {
			return fmt.Errorf("failed to save Model pom.xml: %w", err)
		}
		color.New(color.FgGreen).Println("  ✓ Added JobRunr dependency to Model")

		// Add job request files
		jobsDir := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, "jobs"))
		if err := os.MkdirAll(jobsDir, 0755); err != nil {
			return err
		}

		// PlaceholderJobRequest.java
		jobReqPath := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, filepath.Join("jobs", "PlaceholderJobRequest.java")))
		if _, err := os.Stat(jobReqPath); os.IsNotExist(err) {
			if err := gen.writeTemplate(
				"java/model/jobs/PlaceholderJobRequest.java.tmpl",
				gen.javaPath(config.ModuleModel, filepath.Join("jobs", "PlaceholderJobRequest.java")),
			); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("  ✓ Added PlaceholderJobRequest.java to Model")
		}

		// ProcessPlaceholderJobRequest.java
		processPath := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, filepath.Join("jobs", "ProcessPlaceholderJobRequest.java")))
		if _, err := os.Stat(processPath); os.IsNotExist(err) {
			if err := gen.writeTemplate(
				"java/model/jobs/ProcessPlaceholderJobRequest.java.tmpl",
				gen.javaPath(config.ModuleModel, filepath.Join("jobs", "ProcessPlaceholderJobRequest.java")),
			); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("  ✓ Added ProcessPlaceholderJobRequest.java to Model")
		}

		// ProcessPlaceholderJobRequestHandler.java (base class)
		handlerPath := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, filepath.Join("jobs", "ProcessPlaceholderJobRequestHandler.java")))
		if _, err := os.Stat(handlerPath); os.IsNotExist(err) {
			if err := gen.writeTemplate(
				"java/model/jobs/ProcessPlaceholderJobRequestHandler.java.tmpl",
				gen.javaPath(config.ModuleModel, filepath.Join("jobs", "ProcessPlaceholderJobRequestHandler.java")),
			); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("  ✓ Added ProcessPlaceholderJobRequestHandler.java to Model")
		}

	case config.ModuleEventConsumer:
		// Add event files
		eventsDir := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, "events"))
		if err := os.MkdirAll(eventsDir, 0755); err != nil {
			return err
		}

		// PlaceholderEvent.java
		eventPath := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, filepath.Join("events", "PlaceholderEvent.java")))
		if _, err := os.Stat(eventPath); os.IsNotExist(err) {
			if err := gen.writeTemplate(
				"java/model/events/PlaceholderEvent.java.tmpl",
				gen.javaPath(config.ModuleModel, filepath.Join("events", "PlaceholderEvent.java")),
			); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("  ✓ Added PlaceholderEvent.java to Model")
		}

		// PlaceholderCreatedEvent.java
		createdPath := filepath.Join(a.projectPath, gen.javaPath(config.ModuleModel, filepath.Join("events", "PlaceholderCreatedEvent.java")))
		if _, err := os.Stat(createdPath); os.IsNotExist(err) {
			if err := gen.writeTemplate(
				"java/model/events/PlaceholderCreatedEvent.java.tmpl",
				gen.javaPath(config.ModuleModel, filepath.Join("events", "PlaceholderCreatedEvent.java")),
			); err != nil {
				return err
			}
			color.New(color.FgGreen).Println("  ✓ Added PlaceholderCreatedEvent.java to Model")
		}
	}

	return nil
}

// updateSharedModule updates the Shared module when adding a datastore
// This adds the datastore as a dependency and regenerates PlaceholderService
func (a *ModuleAdder) updateSharedModule(module string) error {
	// Only update Shared for datastore modules, and only if Shared exists
	if module != config.ModuleSQLDatastore && module != config.ModuleNoSQLDatastore {
		return nil
	}

	if !a.metadata.HasModule(config.ModuleShared) {
		return nil
	}

	gen := &Generator{
		config: a.config,
		engine: a.engine,
		outDir: a.projectPath,
	}

	// Add datastore dependency to Shared pom.xml
	sharedPomPath := filepath.Join(a.projectPath, config.ModuleShared, "pom.xml")
	sharedPom, err := NewPOMUpdater(sharedPomPath)
	if err != nil {
		return fmt.Errorf("failed to read Shared pom.xml: %w", err)
	}

	if err := sharedPom.AddDependency("${project.groupId}", module, "${project.version}"); err != nil {
		return fmt.Errorf("failed to add %s dependency to Shared: %w", module, err)
	}
	if err := sharedPom.Save(); err != nil {
		return fmt.Errorf("failed to save Shared pom.xml: %w", err)
	}
	color.New(color.FgGreen).Printf("  ✓ Added %s dependency to Shared\n", module)

	// Backup and regenerate PlaceholderService.java
	servicePath := gen.javaPath(config.ModuleShared, filepath.Join("service", "PlaceholderService.java"))
	if err := a.backup.Backup(servicePath); err != nil {
		return fmt.Errorf("failed to backup PlaceholderService.java: %w", err)
	}
	if err := gen.writeTemplate(
		"java/shared/service/PlaceholderService.java.tmpl",
		servicePath,
	); err != nil {
		return err
	}
	color.New(color.FgGreen).Println("  ✓ Updated PlaceholderService.java to use repository")

	// Backup and regenerate PlaceholderServiceTest.java
	testPath := gen.javaPath(config.ModuleShared, filepath.Join("service", "PlaceholderServiceTest.java"))
	// Put test in test directory
	testPath = strings.Replace(testPath, "/main/", "/test/", 1)
	if err := a.backup.Backup(testPath); err != nil {
		// Test file might not exist, that's OK
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to backup PlaceholderServiceTest.java: %w", err)
		}
	}
	if err := gen.writeTemplate(
		"java/shared/test/PlaceholderServiceTest.java.tmpl",
		testPath,
	); err != nil {
		return err
	}
	color.New(color.FgGreen).Println("  ✓ Updated PlaceholderServiceTest.java")

	return nil
}

// updateAPIModule updates the API module when adding modules that need ComponentScan
// This regenerates Application.java to include new packages in the scan
func (a *ModuleAdder) updateAPIModule(module string) error {
	// Only update API for modules that need to be scanned
	scannableModules := []string{
		config.ModuleSQLDatastore,
		config.ModuleNoSQLDatastore,
		config.ModuleEvents,
		config.ModuleJobs,
	}

	needsUpdate := false
	for _, m := range scannableModules {
		if module == m {
			needsUpdate = true
			break
		}
	}

	if !needsUpdate {
		return nil
	}

	if !a.metadata.HasModule(config.ModuleAPI) {
		return nil
	}

	gen := &Generator{
		config: a.config,
		engine: a.engine,
		outDir: a.projectPath,
	}

	// Backup and regenerate Application.java
	appPath := gen.javaPath(config.ModuleAPI, a.config.ProjectNamePascal()+"ApiApplication.java")
	if err := a.backup.Backup(appPath); err != nil {
		return fmt.Errorf("failed to backup API Application.java: %w", err)
	}
	if err := gen.writeTemplate(
		"java/api/Application.java.tmpl",
		appPath,
	); err != nil {
		return err
	}
	color.New(color.FgGreen).Println("  ✓ Updated API Application.java with ComponentScan")

	// Backup and regenerate application.yml (includes datasource config conditionally)
	ymlPath := filepath.Join(config.ModuleAPI, "src", "main", "resources", "application.yml")
	if err := a.backup.Backup(ymlPath); err != nil {
		return fmt.Errorf("failed to backup API application.yml: %w", err)
	}
	if err := gen.writeTemplate(
		"java/api/resources/application.yml.tmpl",
		ymlPath,
	); err != nil {
		return err
	}
	color.New(color.FgGreen).Println("  ✓ Updated API application.yml with database config")

	return nil
}

// regenerateDocs regenerates README.md and AI agent context files
// This is called after adding a module to update documentation with new module info
func (a *ModuleAdder) regenerateDocs() error {
	gen := &Generator{
		config: a.config,
		engine: a.engine,
		outDir: a.projectPath,
	}

	// Regenerate README.md
	if err := gen.writeTemplate("docs/README.md.tmpl", "README.md"); err != nil {
		return fmt.Errorf("failed to regenerate README.md: %w", err)
	}

	// Regenerate AI agent context files for each selected agent
	// All agents use the same template content (CLAUDE.md.tmpl), just different file paths
	for _, agent := range a.config.GetSelectedAIAgents() {
		if err := gen.writeTemplate("docs/CLAUDE.md.tmpl", agent.FilePath); err != nil {
			return fmt.Errorf("failed to regenerate %s: %w", agent.FilePath, err)
		}
	}

	// Regenerate AGENTS.md cross-tool baseline
	if a.config.HasAnyAIAgent() {
		if err := gen.writeTemplate("docs/AGENTS.md.tmpl", "AGENTS.md"); err != nil {
			return fmt.Errorf("failed to regenerate AGENTS.md: %w", err)
		}
	}

	// Update .ai directory with new prompts if any AI agent is selected
	if a.config.HasAnyAIAgent() {
		if err := gen.generateAIDirectory(); err != nil {
			return fmt.Errorf("failed to update .ai directory: %w", err)
		}
	}

	// Generate MCP configuration files when MCP module is selected
	if a.config.HasModule(config.ModuleMCP) {
		// Claude Code: .mcp.json (project root)
		if err := gen.writeTemplate("docs/mcp.json.tmpl", ".mcp.json"); err != nil {
			return fmt.Errorf("failed to generate .mcp.json: %w", err)
		}

		// Cursor: .cursor/mcp.json
		if err := gen.writeTemplate("docs/cursor-mcp.json.tmpl", ".cursor/mcp.json"); err != nil {
			return fmt.Errorf("failed to generate .cursor/mcp.json: %w", err)
		}

		// VS Code / GitHub Copilot: .vscode/mcp.json
		if err := gen.writeTemplate("docs/vscode-mcp.json.tmpl", ".vscode/mcp.json"); err != nil {
			return fmt.Errorf("failed to generate .vscode/mcp.json: %w", err)
		}

		// MCP README with setup instructions for all agents
		if err := gen.writeTemplate("docs/MCP-README.md.tmpl", "MCP/README.md"); err != nil {
			return fmt.Errorf("failed to generate MCP/README.md: %w", err)
		}
	}

	// Regenerate CI workflow when a CI provider is configured
	if a.config.HasCIProvider("github") {
		if err := gen.writeTemplate("github/workflows/ci.yml.tmpl", ".github/workflows/ci.yml"); err != nil {
			return fmt.Errorf("failed to regenerate CI workflow: %w", err)
		}
	}

	// Regenerate agent-specific files
	if a.config.HasAIAgent("claude") {
		if err := gen.generateClaudeCodeFiles(); err != nil {
			return fmt.Errorf("failed to regenerate Claude Code files: %w", err)
		}
	}
	if a.config.HasAIAgent("cursor") {
		if err := gen.generateCursorFiles(); err != nil {
			return fmt.Errorf("failed to regenerate Cursor files: %w", err)
		}
	}
	if a.config.HasAIAgent("copilot") {
		if err := gen.generateCopilotFiles(); err != nil {
			return fmt.Errorf("failed to regenerate Copilot files: %w", err)
		}
	}
	if a.config.HasAIAgent("windsurf") {
		if err := gen.generateWindsurfFiles(); err != nil {
			return fmt.Errorf("failed to regenerate Windsurf files: %w", err)
		}
	}
	if a.config.HasAIAgent("cline") {
		if err := gen.generateClineFiles(); err != nil {
			return fmt.Errorf("failed to regenerate Cline files: %w", err)
		}
	}

	return nil
}

// getModuleFiles returns the list of files that would be created for a module
func (a *ModuleAdder) getModuleFiles(module string) []string {
	var files []string
	packagePath := a.config.PackagePath()

	switch module {
	case config.ModuleModel:
		base := filepath.Join(config.ModuleModel, "src", "main", "java", packagePath, "model")
		files = append(files,
			filepath.Join(config.ModuleModel, "pom.xml"),
			filepath.Join(base, "ImmutableStyle.java"),
			filepath.Join(base, "entities", "Placeholder.java"),
			filepath.Join(base, "dto", "PlaceholderRequest.java"),
			filepath.Join(base, "dto", "PlaceholderResponse.java"),
		)

	case config.ModuleSQLDatastore:
		base := filepath.Join(config.ModuleSQLDatastore, "src", "main", "java", packagePath, "sqldatastore")
		files = append(files,
			filepath.Join(config.ModuleSQLDatastore, "pom.xml"),
			filepath.Join(base, "config", "DatabaseConfig.java"),
			filepath.Join(base, "repository", "PlaceholderRepository.java"),
			filepath.Join(config.ModuleSQLDatastore, "src", "main", "resources", "db", "migration", "V1__baseline.sql"),
			filepath.Join(config.ModuleSQLDatastore, "src", "main", "resources", "application.yml"),
		)

	case config.ModuleNoSQLDatastore:
		base := filepath.Join(config.ModuleNoSQLDatastore, "src", "main", "java", packagePath, "nosqldatastore")
		files = append(files,
			filepath.Join(config.ModuleNoSQLDatastore, "pom.xml"),
			filepath.Join(base, "config", "NoSQLConfig.java"),
			filepath.Join(base, "repository", "PlaceholderDocumentRepository.java"),
			filepath.Join(config.ModuleNoSQLDatastore, "src", "main", "resources", "application.yml"),
		)

	case config.ModuleShared:
		base := filepath.Join(config.ModuleShared, "src", "main", "java", packagePath, "shared")
		testBase := filepath.Join(config.ModuleShared, "src", "test", "java", packagePath, "shared")
		files = append(files,
			filepath.Join(config.ModuleShared, "pom.xml"),
			filepath.Join(base, "config", "SharedConfig.java"),
			filepath.Join(base, "config", "CircuitBreakerConfiguration.java"),
			filepath.Join(base, "service", "PlaceholderService.java"),
			filepath.Join(config.ModuleShared, "src", "main", "resources", "application.yml"),
			filepath.Join(testBase, "ArchitectureTest.java"),
		)

	case config.ModuleAPI:
		base := filepath.Join(config.ModuleAPI, "src", "main", "java", packagePath, "api")
		files = append(files,
			filepath.Join(config.ModuleAPI, "pom.xml"),
			filepath.Join(base, a.config.ProjectNamePascal()+"ApiApplication.java"),
			filepath.Join(base, "controller", "HealthController.java"),
			filepath.Join(base, "controller", "PlaceholderController.java"),
			filepath.Join(base, "config", "WebConfig.java"),
			filepath.Join(base, "config", "GlobalExceptionHandler.java"),
			filepath.Join(base, "config", "SecurityHeadersFilter.java"),
			filepath.Join(config.ModuleAPI, "src", "main", "resources", "application.yml"),
			filepath.Join(config.ModuleAPI, "Dockerfile"),
		)

	case config.ModuleJobs:
		base := filepath.Join(config.ModuleJobs, "src", "main", "java", packagePath, "jobs")
		files = append(files,
			filepath.Join(config.ModuleJobs, "pom.xml"),
			filepath.Join(base, "PlaceholderJobService.java"),
		)

	case config.ModuleWorker:
		base := filepath.Join(config.ModuleWorker, "src", "main", "java", packagePath, "worker")
		files = append(files,
			filepath.Join(config.ModuleWorker, "pom.xml"),
			filepath.Join(base, a.config.ProjectNamePascal()+"WorkerApplication.java"),
			filepath.Join(base, "config", "JobRunrConfig.java"),
			filepath.Join(base, "config", "RecurringJobsConfig.java"),
			filepath.Join(base, "handler", "ProcessPlaceholderJobRequestHandler.java"),
			filepath.Join(config.ModuleWorker, "src", "main", "resources", "application.yml"),
			filepath.Join(config.ModuleWorker, "Dockerfile"),
		)

	case config.ModuleEvents:
		base := filepath.Join(config.ModuleEvents, "src", "main", "java", packagePath, "events")
		files = append(files,
			filepath.Join(config.ModuleEvents, "pom.xml"),
			filepath.Join(base, "EventPublisher.java"),
		)

	case config.ModuleEventConsumer:
		base := filepath.Join(config.ModuleEventConsumer, "src", "main", "java", packagePath, "eventconsumer")
		files = append(files,
			filepath.Join(config.ModuleEventConsumer, "pom.xml"),
			filepath.Join(base, a.config.ProjectNamePascal()+"EventConsumerApplication.java"),
			filepath.Join(base, "listener", "PlaceholderEventListener.java"),
			filepath.Join(config.ModuleEventConsumer, "src", "main", "resources", "application.yml"),
			filepath.Join(config.ModuleEventConsumer, "Dockerfile"),
		)

	case config.ModuleMCP:
		base := filepath.Join(config.ModuleMCP, "src", "main", "java", packagePath, "mcp")
		files = append(files,
			filepath.Join(config.ModuleMCP, "pom.xml"),
			filepath.Join(base, "McpServerApplication.java"),
			filepath.Join(base, "tools", "BuildTools.java"),
			filepath.Join(base, "tools", "TestTools.java"),
			filepath.Join(base, "tools", "ProjectTools.java"),
			filepath.Join(base, "tools", "QualityTools.java"),
			filepath.Join(base, "tools", "ReviewTools.java"),
			filepath.Join(config.ModuleMCP, "src", "main", "resources", "logback.xml"),
			filepath.Join(config.ModuleMCP, "README.md"),
			".mcp.json",
			".cursor/mcp.json",
			".vscode/mcp.json",
		)
	}

	return files
}
