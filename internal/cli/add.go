package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/doctor"
	"github.com/arianlopezc/Trabuco/internal/generator"
	"github.com/arianlopezc/Trabuco/internal/prompts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	addDatabase      string
	addNoSQLDatabase string
	addMessageBroker string
	addDryRun        bool
	addNoBackup      bool
	addSkipDoctor    bool
	addSkipBuild     bool
)

var addCmd = &cobra.Command{
	Use:   "add [module]",
	Short: "Add a module to an existing Trabuco project",
	Long: `Add a module to an existing Trabuco project.

The add command validates your project health first (using doctor),
then adds the specified module along with any required dependencies.

Available modules:
  SQLDatastore    - SQL repositories, Flyway migrations
  NoSQLDatastore  - NoSQL repositories (MongoDB, Redis)
  Shared          - Services, Circuit breaker
  API             - REST endpoints
  Worker          - Background jobs (JobRunr)
  EventConsumer   - Event listeners (Kafka, RabbitMQ, SQS, Pub/Sub)
  MCP             - MCP server for AI tool integration

Examples:
  trabuco add SQLDatastore
  trabuco add SQLDatastore --database=postgresql
  trabuco add EventConsumer --message-broker=kafka
  trabuco add Worker --dry-run
  trabuco add                    # Interactive mode`,
	Run: runAdd,
}

func init() {
	addCmd.Flags().StringVar(&addDatabase, "database", "", "SQL database type: postgresql, mysql, generic")
	addCmd.Flags().StringVar(&addNoSQLDatabase, "nosql-database", "", "NoSQL database type: mongodb, redis")
	addCmd.Flags().StringVar(&addMessageBroker, "message-broker", "", "Message broker: kafka, rabbitmq, sqs, pubsub")
	addCmd.Flags().BoolVar(&addDryRun, "dry-run", false, "Show what would change without making changes")
	addCmd.Flags().BoolVar(&addNoBackup, "no-backup", false, "Skip creating backup (not recommended)")
	addCmd.Flags().BoolVar(&addSkipDoctor, "skip-doctor", false, "Skip doctor validation (not recommended)")
	addCmd.Flags().BoolVar(&addSkipBuild, "skip-build", false, "Skip running 'mvn clean install' after adding module")
}

func runAdd(cmd *cobra.Command, args []string) {
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)
	cyan := color.New(color.FgCyan)

	// Get current working directory
	projectPath, err := os.Getwd()
	if err != nil {
		red.Fprintf(os.Stderr, "Error: could not get current directory: %v\n", err)
		os.Exit(1)
	}

	// Step 1: Run doctor (unless skipped)
	if !addSkipDoctor {
		doc := doctor.New(projectPath, Version)
		result, err := doc.Run()
		if err != nil {
			red.Fprintf(os.Stderr, "Error running doctor: %v\n", err)
			os.Exit(1)
		}

		if result.HasErrors() {
			fmt.Println()
			red.Println("Project has errors that must be fixed first.")
			fmt.Println("Run 'trabuco doctor' for details.")
			os.Exit(1)
		}

		if result.HasWarnings() {
			fmt.Println()
			yellow.Println("Project has warnings (continuing anyway):")
			result.PrintWarnings()
			fmt.Println()
		}
	}

	// Step 2: Load/detect project metadata
	metadata, err := doctor.GetProjectMetadata(projectPath)
	if err != nil {
		red.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Show detected project info
	fmt.Println()
	cyan.Printf("Detected Trabuco project: ")
	fmt.Println(metadata.ProjectName)
	cyan.Printf("  Modules: ")
	fmt.Println(strings.Join(metadata.Modules, ", "))
	if metadata.Database != "" {
		cyan.Printf("  Database: ")
		fmt.Println(metadata.Database)
	}
	if metadata.NoSQLDatabase != "" {
		cyan.Printf("  NoSQL Database: ")
		fmt.Println(metadata.NoSQLDatabase)
	}
	if metadata.MessageBroker != "" {
		cyan.Printf("  Message Broker: ")
		fmt.Println(metadata.MessageBroker)
	}
	cyan.Printf("  Java: ")
	fmt.Println(metadata.JavaVersion)
	fmt.Println()

	// Step 3: Determine which module to add
	var module string
	if len(args) > 0 {
		module = args[0]
	} else {
		// Interactive mode
		module, err = prompts.PromptModuleSelection(metadata.Modules)
		if err != nil {
			if err.Error() == "all modules are already present in this project" {
				green.Println("All available modules are already present in this project.")
				os.Exit(0)
			}
			red.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Step 4: Validate module can be added
	if err := prompts.ValidateModuleCanBeAdded(module, metadata.Modules); err != nil {
		red.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Step 5: Get module-specific options
	database := addDatabase
	nosqlDatabase := addNoSQLDatabase
	messageBroker := addMessageBroker

	if module == config.ModuleSQLDatastore && database == "" {
		database, err = prompts.PromptDatabase()
		if err != nil {
			red.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if module == config.ModuleNoSQLDatastore && nosqlDatabase == "" {
		hasWorker := false
		for _, m := range metadata.Modules {
			if m == config.ModuleWorker {
				hasWorker = true
				break
			}
		}
		nosqlDatabase, err = prompts.PromptNoSQLDatabase(hasWorker)
		if err != nil {
			red.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	if module == config.ModuleEventConsumer && messageBroker == "" {
		messageBroker, err = prompts.PromptMessageBroker()
		if err != nil {
			red.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	}

	// Step 6: Create module adder
	adder := generator.NewModuleAdder(projectPath, metadata, Version, !addNoBackup)

	// Step 7: Dry run if requested
	if addDryRun {
		result := adder.DryRun(module)
		result.Print()
		fmt.Println()
		yellow.Println("This is a dry run. No changes were made.")
		os.Exit(0)
	}

	// Step 8: Show what will be added
	dependencies := adder.ResolveDependencies(module)
	fmt.Printf("Adding %s module", module)
	if len(dependencies) > 0 {
		fmt.Printf(" (with dependencies: %s)", strings.Join(dependencies, ", "))
	}
	fmt.Println("...")
	fmt.Println()

	// Step 9: Add the module
	if err := adder.Add(module, database, nosqlDatabase, messageBroker); err != nil {
		red.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	// Step 10: Success message
	fmt.Println()
	green.Println("✓ Module added successfully!")
	fmt.Println()

	// Step 11: Run Maven build unless skipped
	if addSkipBuild {
		fmt.Println("Skipping Maven build (--skip-build flag).")
		fmt.Println()
		cyan.Println("Next steps:")
		fmt.Println("  mvn clean install")
		if needsDocker(module, database, nosqlDatabase, messageBroker) {
			fmt.Println("  docker-compose up -d")
		}
	} else {
		// Run Maven build
		if err := runMavenBuild(projectPath); err != nil {
			yellow.Printf("\nMaven build failed: %v\n", err)
			fmt.Println("You can try running it manually:")
			fmt.Println("  mvn clean install")
			fmt.Println()
		} else {
			green.Println("✓ Maven build completed successfully!")
			fmt.Println()
		}
		if needsDocker(module, database, nosqlDatabase, messageBroker) {
			cyan.Println("Next steps:")
			fmt.Println("  docker-compose up -d")
		}
	}

	// Show MCP server info when MCP module is added
	if module == config.ModuleMCP {
		fmt.Println()
		cyan.Println("MCP Server:")
		fmt.Println("  JAR: MCP/target/MCP-1.0-SNAPSHOT.jar")
		fmt.Println()
		fmt.Println("  Pre-configured for:")
		fmt.Println("    • Claude Code  → .mcp.json")
		fmt.Println("    • Cursor       → .cursor/mcp.json")
		fmt.Println("    • VS Code      → .vscode/mcp.json")
		fmt.Println()
		fmt.Println("  See MCP/README.md for setup instructions.")
	}
}

// needsDocker returns true if the module requires docker services
func needsDocker(module, database, nosqlDatabase, messageBroker string) bool {
	switch module {
	case config.ModuleSQLDatastore:
		return database == config.DatabasePostgreSQL || database == config.DatabaseMySQL
	case config.ModuleNoSQLDatastore:
		return nosqlDatabase != ""
	case config.ModuleEventConsumer:
		return messageBroker != ""
	case config.ModuleWorker:
		return true // Always needs something for JobRunr
	default:
		return false
	}
}

// GetAvailableModulesToAdd returns the list of modules that can be added
// Used by shell completion
func GetAvailableModulesToAdd() []string {
	projectPath, err := os.Getwd()
	if err != nil {
		return config.GetSelectableModules()
	}

	metadata, err := doctor.GetProjectMetadata(projectPath)
	if err != nil {
		return config.GetSelectableModules()
	}

	existingSet := make(map[string]bool)
	for _, m := range metadata.Modules {
		existingSet[m] = true
	}

	// Check for mutual exclusion
	hasSQLDatastore := existingSet[config.ModuleSQLDatastore]
	hasNoSQLDatastore := existingSet[config.ModuleNoSQLDatastore]

	var available []string
	for _, m := range config.ModuleRegistry {
		if m.Internal || m.Required {
			continue
		}
		if existingSet[m.Name] {
			continue
		}
		if m.Name == config.ModuleSQLDatastore && hasNoSQLDatastore {
			continue
		}
		if m.Name == config.ModuleNoSQLDatastore && hasSQLDatastore {
			continue
		}
		available = append(available, m.Name)
	}

	return available
}
