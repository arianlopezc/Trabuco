package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/auth"
	"github.com/arianlopezc/Trabuco/internal/migrate"
	"github.com/arianlopezc/Trabuco/internal/prompts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	migrateOutput       string
	migrateAPIKey       string
	migrateProvider     string
	migrateModel        string
	migrateDryRun       bool
	migrateInteractive  bool
	migrateResume       bool
	migrateRollback     bool
	migrateRollbackTo   string
	migrateIncludeTests bool
	migrateVerbose      bool
	migrateDebug        bool
	migrateSkipBuild    bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate [source-project]",
	Short: "[Experimental] Migrate an existing Java project to Trabuco structure using AI",
	Long: `[EXPERIMENTAL] Migrate an existing Spring Boot project to Trabuco's multi-module architecture.

NOTE: This command is experimental and under active development.
      Results should be reviewed and may require manual adjustments.

This command uses AI (Claude) to analyze your existing Java project and
intelligently migrate it to Trabuco's clean module structure:
  - Model: DTOs, Entities, Enums
  - SQLDatastore/NoSQLDatastore: Repositories, Migrations
  - Shared: Services, Utilities
  - API: REST Controllers
  - Worker: Background Jobs (if applicable)
  - EventConsumer: Event Listeners (if applicable)

MODES:
  Interactive (Guided):
    Run without arguments for a step-by-step guided experience:
      trabuco migrate

  Non-Interactive:
    Provide source path and flags for scripted/CI usage:
      trabuco migrate /path/to/legacy-app [flags]

AUTHENTICATION:
  Configure credentials with: trabuco auth login
  Or set environment variables: ANTHROPIC_API_KEY, OPENROUTER_API_KEY
  Or use --api-key flag for one-time use.

Examples:
  # Guided interactive migration (recommended for first-time users)
  trabuco migrate

  # Basic migration with source path
  trabuco migrate /path/to/legacy-app

  # Specify output directory
  trabuco migrate /path/to/legacy-app -o ./migrated-app

  # Use OpenRouter instead of Anthropic
  trabuco migrate --provider=openrouter /path/to/legacy-app

  # Dry run (analyze without generating files)
  trabuco migrate --dry-run /path/to/legacy-app

  # Resume interrupted migration
  trabuco migrate --resume /path/to/legacy-app

  # Verbose output
  trabuco migrate -v /path/to/legacy-app`,
	Args: cobra.MaximumNArgs(1),
	Run:  runMigrate,
}

func init() {
	migrateCmd.Flags().StringVarP(&migrateOutput, "output", "o", "", "Output directory (default: ./<project-name>-trabuco)")
	migrateCmd.Flags().StringVar(&migrateAPIKey, "api-key", "", "API key (or set ANTHROPIC_API_KEY/OPENROUTER_API_KEY env var)")
	migrateCmd.Flags().StringVar(&migrateProvider, "provider", "anthropic", "AI provider: anthropic, openrouter")
	migrateCmd.Flags().StringVar(&migrateModel, "model", "", "Model to use (default: claude-sonnet-4-5)")
	migrateCmd.Flags().BoolVar(&migrateDryRun, "dry-run", false, "Analyze only, don't generate files")
	migrateCmd.Flags().BoolVar(&migrateInteractive, "interactive", true, "Confirm each major decision")
	migrateCmd.Flags().BoolVar(&migrateResume, "resume", false, "Resume from last checkpoint")
	migrateCmd.Flags().BoolVar(&migrateRollback, "rollback", false, "Rollback migration completely")
	migrateCmd.Flags().StringVar(&migrateRollbackTo, "rollback-to", "", "Rollback to specific stage")
	migrateCmd.Flags().BoolVar(&migrateIncludeTests, "include-tests", false, "Migrate test files")
	migrateCmd.Flags().BoolVarP(&migrateVerbose, "verbose", "v", false, "Verbose output")
	migrateCmd.Flags().BoolVar(&migrateDebug, "debug", false, "Debug mode (save all AI interactions)")
	migrateCmd.Flags().BoolVar(&migrateSkipBuild, "skip-build", false, "Skip Maven build after migration")
}

func runMigrate(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	// Print header
	cyan.Println("\n╔════════════════════════════════════════╗")
	cyan.Println("║   Trabuco Migrate - AI Code Migration  ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	// Experimental warning (can be suppressed via env var for CI/CD)
	if os.Getenv("TRABUCO_ACKNOWLEDGE_EXPERIMENTAL") != "true" {
		yellow.Println("⚠️  EXPERIMENTAL FEATURE")
		yellow.Println("   This command is under active development.")
		yellow.Println("   Results may require manual review and adjustment.")
		fmt.Println()
		yellow.Println("   Set TRABUCO_ACKNOWLEDGE_EXPERIMENTAL=true to suppress this warning.")
		fmt.Println()
	}

	var absSourcePath string
	var err error

	// Track if we have pre-scanned project info from guided flow
	var preScannedProjectInfo *migrate.ProjectInfo
	var preAnalyzedDependencies *migrate.DependencyReport

	// Check if we're in guided mode (no source path provided)
	if len(args) == 0 {
		// Run guided interactive prompts
		guidedConfig, err := prompts.RunMigratePrompts()
		if err != nil {
			red.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		// Transfer guided config to CLI variables
		absSourcePath = guidedConfig.SourcePath
		migrateOutput = guidedConfig.OutputPath
		migrateProvider = guidedConfig.Provider
		migrateAPIKey = guidedConfig.APIKey
		migrateModel = guidedConfig.Model
		migrateDryRun = guidedConfig.DryRun
		migrateInteractive = guidedConfig.Interactive
		migrateIncludeTests = guidedConfig.IncludeTests
		migrateSkipBuild = guidedConfig.SkipBuild

		// Pass pre-scanned data to avoid duplicate work
		preScannedProjectInfo = guidedConfig.ProjectInfo
		preAnalyzedDependencies = guidedConfig.DependencyReport
	} else {
		// Non-interactive mode: use provided source path
		sourcePath := args[0]

		// Resolve absolute path
		absSourcePath, err = filepath.Abs(sourcePath)
		if err != nil {
			red.Fprintf(os.Stderr, "Error: invalid source path: %v\n", err)
			os.Exit(1)
		}

		// Verify source exists
		if _, err := os.Stat(absSourcePath); os.IsNotExist(err) {
			red.Fprintf(os.Stderr, "Error: source project not found: %s\n", absSourcePath)
			os.Exit(1)
		}
	}

	// Handle rollback
	if migrateRollback || migrateRollbackTo != "" {
		handleRollback(absSourcePath)
		return
	}

	// Create AI provider using credential manager
	provider, providerName, err := createAIProviderWithAuth()
	if err != nil {
		red.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Println()
		yellow.Println("No API credentials found.")
		fmt.Println()
		fmt.Println("To configure credentials, run:")
		fmt.Println("  trabuco auth login")
		fmt.Println()
		fmt.Println("Or set environment variables:")
		fmt.Println("  export ANTHROPIC_API_KEY=sk-ant-...")
		fmt.Println("  export OPENROUTER_API_KEY=sk-or-...")
		fmt.Println()
		fmt.Println("Get an API key:")
		fmt.Println("  - Anthropic: https://console.anthropic.com/settings/keys")
		fmt.Println("  - OpenRouter: https://openrouter.ai/keys")
		os.Exit(1)
	}

	green.Printf("Using AI provider: %s\n", providerName)
	fmt.Println()

	// Create migration config
	config := &migrate.Config{
		SourcePath:              absSourcePath,
		OutputPath:              migrateOutput,
		DryRun:                  migrateDryRun,
		Interactive:             migrateInteractive,
		Resume:                  migrateResume,
		IncludeTests:            migrateIncludeTests,
		Verbose:                 migrateVerbose,
		Debug:                   migrateDebug,
		SkipBuild:               migrateSkipBuild,
		TrabucoVersion:          Version,
		PreScannedProjectInfo:   preScannedProjectInfo,
		PreAnalyzedDependencies: preAnalyzedDependencies,
	}

	// Determine output path if not specified
	if config.OutputPath == "" {
		projectName := filepath.Base(absSourcePath)
		config.OutputPath = filepath.Join(".", projectName+"-trabuco")
	}

	// Resolve output path
	config.OutputPath, err = filepath.Abs(config.OutputPath)
	if err != nil {
		red.Fprintf(os.Stderr, "Error: invalid output path: %v\n", err)
		os.Exit(1)
	}

	// Create migrator
	migrator := migrate.NewMigrator(provider, config)

	// Run migration
	if err := migrator.Run(); err != nil {
		red.Fprintf(os.Stderr, "\nMigration failed: %v\n", err)

		// Check if we have a checkpoint to resume from
		if migrator.HasCheckpoint() {
			fmt.Println()
			yellow.Println("A checkpoint was saved. You can resume with:")
			fmt.Printf("  trabuco migrate --resume %s\n", absSourcePath)
		}

		os.Exit(1)
	}

	// Success
	fmt.Println()
	green.Println("✓ Migration completed successfully!")
	fmt.Println()

	fmt.Printf("Output: %s\n", config.OutputPath)
	fmt.Println()

	if !config.SkipBuild && !config.DryRun {
		fmt.Println("The project has been built. To run it:")
		fmt.Printf("  cd %s\n", config.OutputPath)
		fmt.Println("  docker-compose up -d  # Start dependencies")
		fmt.Println("  cd API && mvn spring-boot:run")
	} else if config.DryRun {
		yellow.Println("This was a dry run. No files were generated.")
		fmt.Println("Run without --dry-run to perform the migration.")
	} else {
		fmt.Println("Next steps:")
		fmt.Printf("  cd %s\n", config.OutputPath)
		fmt.Println("  mvn clean install")
	}
}

func createAIProviderWithAuth() (ai.Provider, string, error) {
	// If API key was provided directly via flag, use it
	if migrateAPIKey != "" {
		config := &ai.ProviderConfig{
			APIKey: migrateAPIKey,
			Model:  migrateModel,
		}

		providerType := ai.ProviderType(strings.ToLower(migrateProvider))
		switch providerType {
		case ai.ProviderTypeAnthropic:
			provider, err := ai.NewAnthropicProvider(config)
			return provider, "Anthropic (Claude)", err
		case ai.ProviderTypeOpenRouter:
			provider, err := ai.NewOpenRouterProvider(config)
			return provider, "OpenRouter", err
		default:
			provider, err := ai.AutoDetectProvider(config)
			if provider != nil {
				return provider, provider.Name(), err
			}
			return nil, "", err
		}
	}

	// Use credential manager to get credentials
	manager, err := auth.NewManager()
	if err != nil {
		return nil, "", fmt.Errorf("failed to initialize credential manager: %w", err)
	}

	// Determine preferred provider from flag
	var preferredProvider auth.Provider
	switch strings.ToLower(migrateProvider) {
	case "anthropic":
		preferredProvider = auth.ProviderAnthropic
	case "openrouter":
		preferredProvider = auth.ProviderOpenRouter
	case "openai":
		preferredProvider = auth.ProviderOpenAI
	}

	// Get credentials with fallback
	cred, err := manager.GetCredentialWithFallback(preferredProvider)
	if err != nil {
		return nil, "", err
	}

	// Create provider config
	config := &ai.ProviderConfig{
		APIKey: cred.APIKey,
		Model:  migrateModel,
	}

	// Use model from credential if not specified
	if config.Model == "" && cred.Model != "" {
		config.Model = cred.Model
	}

	// Create the appropriate provider
	providerInfo := auth.SupportedProviders[cred.Provider]
	switch cred.Provider {
	case auth.ProviderAnthropic:
		provider, err := ai.NewAnthropicProvider(config)
		return provider, providerInfo.Name, err
	case auth.ProviderOpenRouter:
		provider, err := ai.NewOpenRouterProvider(config)
		return provider, providerInfo.Name, err
	default:
		// Try auto-detection for unsupported providers
		provider, err := ai.AutoDetectProvider(config)
		if provider != nil {
			return provider, providerInfo.Name, err
		}
		return nil, "", fmt.Errorf("unsupported provider: %s", cred.Provider)
	}
}

func handleRollback(sourcePath string) {
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	checkpointDir := migrate.GetCheckpointDir(sourcePath)

	if migrateRollbackTo != "" {
		cyan.Printf("Rolling back to stage: %s\n", migrateRollbackTo)
		if err := migrate.RollbackToStage(checkpointDir, migrateRollbackTo); err != nil {
			red.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
			os.Exit(1)
		}
		yellow.Println("Rollback complete. You can resume migration with --resume flag.")
	} else {
		cyan.Println("Rolling back migration completely...")
		if err := migrate.RollbackAll(checkpointDir); err != nil {
			red.Fprintf(os.Stderr, "Rollback failed: %v\n", err)
			os.Exit(1)
		}
		yellow.Println("Migration rolled back. Output directory has been removed.")
	}
}
