package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/generator"
	"github.com/arianlopezc/Trabuco/internal/java"
	"github.com/arianlopezc/Trabuco/internal/prompts"
	"github.com/arianlopezc/Trabuco/internal/utils"
)

// Validation patterns for non-interactive mode
var (
	projectNameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	groupIDRegex     = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$`)
)

// Non-interactive mode flags
var (
	flagProjectName   string
	flagGroupID       string
	flagModules       string
	flagDatabase      string
	flagNoSQLDatabase string
	flagMessageBroker string
	flagJavaVersion   string
	flagAIAgents      string
	flagIncludeClaude bool // Deprecated: use flagAIAgents instead
	flagStrict        bool
	flagSkipBuild     bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Java project",
	Long: `Initialize a new Java multi-module Maven project.

This command will interactively prompt you for:
  - Project name
  - Group ID (e.g., com.company.project)
  - Modules to include
  - Database type (if SQLDatastore selected)

Note: SQLDatastore and NoSQLDatastore cannot be selected together.

For non-interactive mode, provide all required flags:
  trabuco init --name=myproject --group-id=com.company.project --modules=Model,SQLDatastore --database=postgresql`,
	Run: runInit,
}

func init() {
	initCmd.Flags().StringVar(&flagProjectName, "name", "", "Project name (non-interactive)")
	initCmd.Flags().StringVar(&flagGroupID, "group-id", "", "Group ID, e.g., com.company.project (non-interactive)")
	initCmd.Flags().StringVar(&flagModules, "modules", "", "Comma-separated modules: Model,SQLDatastore,NoSQLDatastore,Shared,API,EventConsumer (SQLDatastore and NoSQLDatastore are mutually exclusive)")
	initCmd.Flags().StringVar(&flagDatabase, "database", "postgresql", "SQL database type: postgresql, mysql, none (non-interactive)")
	initCmd.Flags().StringVar(&flagNoSQLDatabase, "nosql-database", "mongodb", "NoSQL database type: mongodb, redis (non-interactive)")
	initCmd.Flags().StringVar(&flagMessageBroker, "message-broker", "kafka", "Message broker type: kafka, rabbitmq, sqs, pubsub (non-interactive, only used when EventConsumer is selected)")
	initCmd.Flags().StringVar(&flagJavaVersion, "java-version", "21", "Java version: 17, 21, or 25 (non-interactive)")
	initCmd.Flags().StringVar(&flagAIAgents, "ai-agents", "", "Comma-separated AI agents: claude,cursor,copilot,windsurf,cline (non-interactive)")
	initCmd.Flags().BoolVar(&flagIncludeClaude, "include-claude", false, "Deprecated: use --ai-agents=claude instead")
	initCmd.Flags().BoolVar(&flagStrict, "strict", false, "Fail if specified Java version is not detected (non-interactive)")
	initCmd.Flags().BoolVar(&flagSkipBuild, "skip-build", false, "Skip running 'mvn clean install' after generation")
}

func runInit(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	cyan.Println("\n╔════════════════════════════════════════╗")
	cyan.Println("║   Trabuco - Java Project Generator     ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// Validate Docker is running (required for Testcontainers and local development)
	dockerStatus := utils.CheckDocker()
	if !dockerStatus.Running {
		color.Red("Error: Docker is required but not available.\n")
		if !dockerStatus.Installed {
			color.Red("       Docker is not installed. Please install Docker Desktop.\n")
		} else {
			color.Red("       %s\n", dockerStatus.Error)
		}
		color.Yellow("\nDocker is required for:\n")
		color.Yellow("  - Running integration tests (Testcontainers)\n")
		color.Yellow("  - Local development with docker-compose\n")
		fmt.Println()
		return
	}

	var cfg *config.ProjectConfig
	var err error

	// Check if non-interactive mode (flags provided)
	if flagProjectName != "" && flagGroupID != "" && flagModules != "" {
		// Non-interactive mode

		// Validate project name
		if !projectNameRegex.MatchString(flagProjectName) {
			color.Red("\nError: Invalid project name '%s'. Must be lowercase, alphanumeric, hyphens allowed (not at start/end).\n", flagProjectName)
			return
		}

		// Validate group ID
		if !groupIDRegex.MatchString(flagGroupID) {
			color.Red("\nError: Invalid group ID '%s'. Must be valid Java package format (e.g., com.company.project).\n", flagGroupID)
			return
		}

		// Validate Java version
		if flagJavaVersion != "17" && flagJavaVersion != "21" && flagJavaVersion != "25" {
			color.Red("\nError: Invalid Java version '%s'. Must be 17, 21, or 25.\n", flagJavaVersion)
			return
		}

		// Java version detection for non-interactive mode
		javaDetection := java.Detect()
		javaVersionInt, _ := strconv.Atoi(flagJavaVersion)
		javaVersionDetected := javaDetection.IsVersionDetected(javaVersionInt)

		if !javaVersionDetected {
			detectedVersions := javaDetection.GetDetectedVersions()
			if flagStrict {
				color.Red("\nError: Java %s not detected (--strict mode).\n", flagJavaVersion)
				if len(detectedVersions) > 0 {
					fmt.Fprintf(os.Stderr, "Detected versions: [%s]\n", java.FormatDetectedVersions(detectedVersions))
				} else {
					fmt.Fprintln(os.Stderr, "No compatible Java versions detected.")
				}
				os.Exit(1)
			}
			// Non-strict mode: warn but continue
			yellow.Fprintf(os.Stderr, "\nWarning: Java %s not detected.", flagJavaVersion)
			if len(detectedVersions) > 0 {
				fmt.Fprintf(os.Stderr, " Detected: [%s]", java.FormatDetectedVersions(detectedVersions))
			}
			fmt.Fprintln(os.Stderr)
		}

		// Validate database type
		validDatabases := map[string]bool{"postgresql": true, "mysql": true, "none": true, "generic": true, "": true}
		if !validDatabases[flagDatabase] {
			color.Red("\nError: Invalid database type '%s'. Must be postgresql, mysql, or none.\n", flagDatabase)
			return
		}

		// Validate NoSQL database type
		validNoSQLDatabases := map[string]bool{"mongodb": true, "redis": true, "": true}
		if !validNoSQLDatabases[flagNoSQLDatabase] {
			color.Red("\nError: Invalid NoSQL database type '%s'. Must be mongodb or redis.\n", flagNoSQLDatabase)
			return
		}

		// Validate message broker type
		validMessageBrokers := map[string]bool{"kafka": true, "rabbitmq": true, "sqs": true, "pubsub": true, "": true}
		if !validMessageBrokers[flagMessageBroker] {
			color.Red("\nError: Invalid message broker type '%s'. Must be kafka, rabbitmq, sqs, or pubsub.\n", flagMessageBroker)
			return
		}

		// Parse and validate AI agents
		var aiAgents []string
		if flagAIAgents != "" {
			validAgents := make(map[string]bool)
			for _, id := range config.GetAIAgentIDs() {
				validAgents[id] = true
			}
			agents := strings.Split(flagAIAgents, ",")
			for _, agent := range agents {
				agent = strings.TrimSpace(strings.ToLower(agent))
				if agent == "" {
					continue
				}
				if !validAgents[agent] {
					color.Red("\nError: Invalid AI agent '%s'. Valid options: %s\n", agent, strings.Join(config.GetAIAgentIDs(), ", "))
					return
				}
				aiAgents = append(aiAgents, agent)
			}
		}

		// Handle deprecated --include-claude flag
		if flagIncludeClaude {
			hasClaudeInList := false
			for _, a := range aiAgents {
				if a == "claude" {
					hasClaudeInList = true
					break
				}
			}
			if !hasClaudeInList {
				aiAgents = append(aiAgents, "claude")
			}
			yellow.Fprintf(os.Stderr, "\nWarning: --include-claude is deprecated. Use --ai-agents=claude instead.\n\n")
		}

		modules := strings.Split(flagModules, ",")
		for i := range modules {
			modules[i] = strings.TrimSpace(modules[i])
		}

		// Validate module selection
		if validationErr := config.ValidateModuleSelection(modules); validationErr != "" {
			color.Red("\nError: %s\n", validationErr)
			return
		}

		// Resolve dependencies (auto-include Jobs when Worker is selected, etc.)
		resolvedModules := config.ResolveDependencies(modules)

		cfg = &config.ProjectConfig{
			ProjectName:         flagProjectName,
			GroupID:             flagGroupID,
			ArtifactID:          flagProjectName,
			JavaVersion:         flagJavaVersion,
			JavaVersionDetected: javaVersionDetected,
			Modules:             resolvedModules,
			Database:            flagDatabase,
			NoSQLDatabase:       flagNoSQLDatabase,
			MessageBroker:       flagMessageBroker,
			AIAgents:            aiAgents,
		}

		// Warn about Redis + Worker combination
		if cfg.ShowRedisWorkerWarning() {
			yellow.Fprintf(os.Stderr, "\nWarning: Redis support is deprecated in JobRunr 8+.\n")
			fmt.Fprintln(os.Stderr, "  JobRunr will use PostgreSQL for job storage instead.")
			fmt.Fprintln(os.Stderr, "  A separate PostgreSQL instance will be added to docker-compose.yml.")
			fmt.Fprintln(os.Stderr)
		}

		fmt.Println("Running in non-interactive mode...")
	} else {
		// Interactive mode - run prompts
		cfg, err = prompts.RunPrompts()
		if err != nil {
			color.Red("\nError: %v\n", err)
			return
		}
	}

	// Display summary
	fmt.Println()
	yellow.Println("─────────────────────────────────────────")
	yellow.Println("  Project Summary")
	yellow.Println("─────────────────────────────────────────")
	fmt.Printf("  Project:    %s\n", cfg.ProjectName)
	fmt.Printf("  Group ID:   %s\n", cfg.GroupID)
	fmt.Printf("  Java:       %s\n", cfg.JavaVersion)
	fmt.Printf("  Modules:    %s\n", strings.Join(cfg.Modules, ", "))
	if cfg.HasModule(config.ModuleSQLDatastore) {
		fmt.Printf("  SQL DB:     %s\n", cfg.Database)
	}
	if cfg.HasModule(config.ModuleNoSQLDatastore) {
		fmt.Printf("  NoSQL DB:   %s\n", cfg.NoSQLDatabase)
	}
	if cfg.HasModule(config.ModuleWorker) {
		storageType := cfg.JobRunrStorageType()
		storageInfo := storageType
		if cfg.WorkerUsesPostgresFallback() {
			storageInfo = "postgresql (fallback)"
		}
		fmt.Printf("  JobRunr:    %s\n", storageInfo)
	}
	if cfg.HasModule(config.ModuleEventConsumer) {
		fmt.Printf("  Broker:     %s\n", cfg.MessageBroker)
	}
	if cfg.HasAnyAIAgent() {
		selectedAgents := cfg.GetSelectedAIAgents()
		agentNames := make([]string, len(selectedAgents))
		for i, a := range selectedAgents {
			agentNames[i] = a.Name
		}
		fmt.Printf("  AI Agents:  %s\n", strings.Join(agentNames, ", "))
	}
	yellow.Println("─────────────────────────────────────────")
	fmt.Println()

	// Generate project
	gen, err := generator.NewWithVersion(cfg, Version)
	if err != nil {
		color.Red("\nError: %v\n", err)
		return
	}

	if err := gen.Generate(); err != nil {
		color.Red("\nError: %v\n", err)
		return
	}

	// Success message
	fmt.Println()
	green.Println("✓ Project generated successfully!")
	fmt.Println()

	// Reminder if Java version was not detected
	if !cfg.JavaVersionDetected {
		yellow.Printf("Reminder: Java %s was not detected. Install before building.\n\n", cfg.JavaVersion)
	}

	// Reminder about PostgreSQL fallback for Worker + Redis
	if cfg.ShowRedisWorkerWarning() {
		yellow.Println("Note: Worker module uses PostgreSQL for job storage (Redis is deprecated in JobRunr 8+).")
		fmt.Println("      Start docker-compose to run the PostgreSQL instance for JobRunr.")
		fmt.Println()
	}

	// Run Maven build unless skipped or Java not detected
	projectDir := filepath.Join(".", cfg.ProjectName)
	if flagSkipBuild {
		fmt.Println("Skipping Maven build (--skip-build flag).")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  cd %s\n", cfg.ProjectName)
		fmt.Printf("  mvn clean install\n")
	} else if !cfg.JavaVersionDetected {
		fmt.Println("Skipping Maven build (Java not detected).")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Printf("  cd %s\n", cfg.ProjectName)
		fmt.Printf("  mvn clean install\n")
	} else {
		// Run Maven build
		if err := runMavenBuild(projectDir); err != nil {
			yellow.Printf("\nMaven build failed: %v\n", err)
			fmt.Println("You can try running it manually:")
			fmt.Printf("  cd %s && mvn clean install\n", cfg.ProjectName)
			fmt.Println()
		} else {
			green.Println("✓ Maven build completed successfully!")
			fmt.Println()
		}
	}

	// Show how to run the application
	if cfg.HasModule(config.ModuleAPI) {
		fmt.Println("To run the API:")
		fmt.Printf("  cd %s/%s && mvn spring-boot:run\n", cfg.ProjectName, config.ModuleAPI)
	}
	if cfg.HasModule(config.ModuleMCP) {
		fmt.Println()
		cyan.Println("MCP Server:")
		fmt.Printf("  JAR: %s/MCP/target/MCP-1.0-SNAPSHOT.jar\n", cfg.ProjectName)
		fmt.Println()
		fmt.Println("  Pre-configured for:")
		fmt.Println("    • Claude Code  → .mcp.json")
		fmt.Println("    • Cursor       → .cursor/mcp.json")
		fmt.Println("    • VS Code      → .vscode/mcp.json")
		fmt.Println()
		fmt.Printf("  See %s/MCP/README.md for setup instructions.\n", cfg.ProjectName)
	}
}

// runMavenBuild executes 'mvn clean install -DskipTests' in the given directory
func runMavenBuild(projectDir string) error {
	cyan := color.New(color.FgCyan)

	cyan.Println("Building project with Maven...")
	fmt.Println()

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

	cmd := exec.Command("mvn", "clean", "install", "-DskipTests", "-q")
	cmd.Dir = projectDir

	// Capture output for error reporting
	output, err := cmd.CombinedOutput()

	// Stop spinner
	done <- true
	time.Sleep(50 * time.Millisecond) // Allow spinner goroutine to clean up

	if err != nil {
		fmt.Printf("\r                                                    \r") // Clear line
		if len(output) > 0 {
			// Show last 20 lines of output on error
			lines := strings.Split(string(output), "\n")
			start := 0
			if len(lines) > 20 {
				start = len(lines) - 20
			}
			fmt.Println("\nMaven output:")
			for _, line := range lines[start:] {
				if line != "" {
					fmt.Printf("  %s\n", line)
				}
			}
		}
		return err
	}

	fmt.Printf("\r                                                    \r") // Clear line
	return nil
}
