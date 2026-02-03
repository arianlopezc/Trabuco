package cli

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/generator"
	"github.com/arianlopezc/Trabuco/internal/java"
	"github.com/arianlopezc/Trabuco/internal/prompts"
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
	flagJavaVersion   string
	flagIncludeClaude bool
	flagStrict        bool
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
	initCmd.Flags().StringVar(&flagModules, "modules", "", "Comma-separated modules: Model,SQLDatastore,NoSQLDatastore,Shared,API (SQLDatastore and NoSQLDatastore are mutually exclusive)")
	initCmd.Flags().StringVar(&flagDatabase, "database", "postgresql", "SQL database type: postgresql, mysql, none (non-interactive)")
	initCmd.Flags().StringVar(&flagNoSQLDatabase, "nosql-database", "mongodb", "NoSQL database type: mongodb, redis (non-interactive)")
	initCmd.Flags().StringVar(&flagJavaVersion, "java-version", "21", "Java version: 17, 21, or 25 (non-interactive)")
	initCmd.Flags().BoolVar(&flagIncludeClaude, "include-claude", true, "Include CLAUDE.md file (non-interactive)")
	initCmd.Flags().BoolVar(&flagStrict, "strict", false, "Fail if specified Java version is not detected (non-interactive)")
}

func runInit(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	cyan.Println("\n╔════════════════════════════════════════╗")
	cyan.Println("║   Trabuco - Java Project Generator     ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()

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
			IncludeCLAUDEMD:     flagIncludeClaude,
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
	if cfg.HasModule("SQLDatastore") {
		fmt.Printf("  SQL DB:     %s\n", cfg.Database)
	}
	if cfg.HasModule("NoSQLDatastore") {
		fmt.Printf("  NoSQL DB:   %s\n", cfg.NoSQLDatabase)
	}
	if cfg.IncludeCLAUDEMD {
		fmt.Printf("  CLAUDE.md:  Yes\n")
	}
	yellow.Println("─────────────────────────────────────────")
	fmt.Println()

	// Generate project
	gen, err := generator.New(cfg)
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

	fmt.Println("Next steps:")
	fmt.Printf("  cd %s\n", cfg.ProjectName)
	fmt.Printf("  mvn clean install\n")
	if cfg.HasModule("API") {
		fmt.Printf("  cd API && mvn spring-boot:run\n")
	}
}
