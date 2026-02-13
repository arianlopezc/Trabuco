package prompts

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"

	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/auth"
	"github.com/arianlopezc/Trabuco/internal/migrate"
)

// MigrateConfig holds the configuration gathered from migrate prompts
type MigrateConfig struct {
	SourcePath   string
	OutputPath   string
	Provider     string
	APIKey       string
	Model        string
	DryRun       bool
	Interactive  bool
	IncludeTests bool
	SkipBuild    bool

	// ProjectInfo holds the pre-scanned project information to avoid duplicate scanning
	ProjectInfo *migrate.ProjectInfo

	// DependencyReport holds the pre-analyzed dependency report
	DependencyReport *migrate.DependencyReport
}

// RunMigratePrompts runs the interactive prompts for the migrate command
func RunMigratePrompts() (*MigrateConfig, error) {
	cfg := &MigrateConfig{
		Interactive: true, // Default to interactive
	}

	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)

	// 1. Source project path
	if err := survey.AskOne(&survey.Input{
		Message: "Path to your existing Java project:",
		Help:    "The directory containing pom.xml of your legacy Spring Boot project",
	}, &cfg.SourcePath, survey.WithValidator(validateSourcePath)); err != nil {
		return nil, err
	}

	// Make path absolute
	absPath, err := filepath.Abs(cfg.SourcePath)
	if err != nil {
		return nil, fmt.Errorf("invalid path: %w", err)
	}
	cfg.SourcePath = absPath

	// 2. Scan the project and show summary
	cyan.Println("\nScanning project...")
	scanner := migrate.NewProjectScanner(cfg.SourcePath)
	projectInfo, err := scanner.Scan()
	if err != nil {
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	// Store projectInfo in config to avoid duplicate scanning
	cfg.ProjectInfo = projectInfo

	// Display discovered info
	fmt.Println()
	yellow.Println("─────────────────────────────────────────")
	yellow.Println("  Discovered Project")
	yellow.Println("─────────────────────────────────────────")
	fmt.Printf("  Name:           %s\n", projectInfo.Name)
	fmt.Printf("  Type:           %s\n", projectInfo.ProjectType)
	fmt.Printf("  Group ID:       %s\n", projectInfo.GroupID)
	fmt.Printf("  Java Version:   %s\n", projectInfo.JavaVersion)
	fmt.Println()

	// Detected infrastructure
	cyan.Println("  Infrastructure:")
	databaseDisplay := getDatabaseDisplayName(projectInfo.Database)
	if projectInfo.UsesNoSQL {
		fmt.Printf("    Database:       %s (NoSQL)\n", databaseDisplay)
	} else {
		fmt.Printf("    Database:       %s (SQL)\n", databaseDisplay)
	}
	if projectInfo.MessageBroker != "" {
		brokerDisplay := projectInfo.MessageBroker
		if projectInfo.MessageBroker == "rabbitmq" {
			brokerDisplay = "RabbitMQ"
		} else if projectInfo.MessageBroker == "kafka" {
			brokerDisplay = "Kafka"
		}
		fmt.Printf("    Message Broker: %s\n", brokerDisplay)
	}
	if projectInfo.UsesRedis {
		fmt.Printf("    Cache:          Redis\n")
	}
	fmt.Println()

	// Source structure
	cyan.Println("  Source Structure:")
	fmt.Printf("    Entities:       %d\n", len(projectInfo.Entities))
	fmt.Printf("    Repositories:   %d\n", len(projectInfo.Repositories))
	fmt.Printf("    Services:       %d\n", len(projectInfo.Services))
	fmt.Printf("    Controllers:    %d\n", len(projectInfo.Controllers))
	if projectInfo.HasScheduledJobs {
		fmt.Printf("    Scheduled Jobs: %d\n", len(projectInfo.ScheduledJobs))
	}
	if projectInfo.HasEventListeners {
		fmt.Printf("    Event Listeners: %d\n", len(projectInfo.EventListeners))
	}
	yellow.Println("─────────────────────────────────────────")
	fmt.Println()

	// Java version warning
	red := color.New(color.FgRed)
	if projectInfo.JavaVersion != "" && projectInfo.JavaVersion < "17" {
		yellow.Println("⚠ Warning: Java version " + projectInfo.JavaVersion + " detected.")
		yellow.Println("  Trabuco targets Java 17+. Some features may need adjustment.")
		fmt.Println()
	}

	// Empty project warning
	totalClasses := len(projectInfo.Entities) + len(projectInfo.Repositories) +
		len(projectInfo.Services) + len(projectInfo.Controllers)
	if totalClasses == 0 {
		red.Println("⚠ Warning: No entities, repositories, services, or controllers found.")
		red.Println("  Please verify this is the correct project directory.")
		fmt.Println()
	}

	// Dependency analysis preview
	analyzer := migrate.NewDependencyAnalyzer()
	depReport := analyzer.Analyze(projectInfo.Dependencies)
	cfg.DependencyReport = depReport

	if len(depReport.Replaceable) > 0 || len(depReport.Unsupported) > 0 {
		fmt.Println()
		yellow.Println("─────────────────────────────────────────")
		yellow.Println("  Dependency Analysis")
		yellow.Println("─────────────────────────────────────────")
		if len(depReport.Compatible) > 0 {
			fmt.Printf("  ✓ Compatible:   %d dependencies\n", len(depReport.Compatible))
		}
		if len(depReport.Replaceable) > 0 {
			yellow.Printf("  ⚠ Replaceable:  %d dependencies\n", len(depReport.Replaceable))
			for _, dep := range depReport.Replaceable {
				fmt.Printf("    • %s → %s\n", dep.Source, dep.TrabucoAlternative)
			}
		}
		if len(depReport.Unsupported) > 0 {
			red.Printf("  ✗ Unsupported:  %d dependencies (manual work needed)\n", len(depReport.Unsupported))
			for _, dep := range depReport.Unsupported {
				fmt.Printf("    • %s\n", dep)
			}
		}
		yellow.Println("─────────────────────────────────────────")
		fmt.Println()
	}

	// 3. Confirm project info is correct
	var confirmProject bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Does this look correct?",
		Default: true,
	}, &confirmProject); err != nil {
		return nil, err
	}
	if !confirmProject {
		return nil, fmt.Errorf("migration cancelled - please verify your project structure")
	}

	// 4. Output directory - default to current directory + project name
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	cfg.OutputPath = filepath.Join(cwd, projectInfo.Name+"-trabuco")

	// Show where output will be generated
	cyan.Printf("\nOutput will be generated in: %s\n", cfg.OutputPath)

	// Check if output exists
	if _, err := os.Stat(cfg.OutputPath); err == nil {
		var overwrite bool
		if err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Directory '%s' already exists. Overwrite?", cfg.OutputPath),
			Default: false,
		}, &overwrite); err != nil {
			return nil, err
		}
		if !overwrite {
			return nil, fmt.Errorf("migration cancelled - output directory exists")
		}
	}

	// Check if current directory is writable
	testFile := filepath.Join(cwd, ".trabuco-write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return nil, fmt.Errorf("current directory is not writable: %w", err)
	}
	os.Remove(testFile) // Clean up test file

	// 5. Check for existing credentials
	green := color.New(color.FgGreen)
	manager, err := auth.NewManager()
	if err != nil {
		yellow.Printf("Warning: Could not load credential manager: %v\n", err)
	}

	// Check if we have stored credentials or environment variables
	var hasStoredCredentials bool
	var storedProvider auth.Provider
	var storedAPIKey string

	if manager != nil {
		if cred, err := manager.GetCredentialWithFallback(""); err == nil {
			hasStoredCredentials = true
			storedProvider = cred.Provider
			storedAPIKey = cred.APIKey
		}
	}

	if hasStoredCredentials {
		// Show stored credential info
		providerInfo := auth.SupportedProviders[storedProvider]
		maskedKey := maskAPIKey(storedAPIKey)

		fmt.Println()
		yellow.Println("─────────────────────────────────────────")
		yellow.Println("  Configured Credentials")
		yellow.Println("─────────────────────────────────────────")
		fmt.Printf("  Provider: %s\n", providerInfo.Name)
		fmt.Printf("  API Key:  %s\n", maskedKey)
		yellow.Println("─────────────────────────────────────────")

		var useStored bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Use these credentials?",
			Default: true,
		}, &useStored); err != nil {
			return nil, err
		}

		if useStored {
			cfg.Provider = string(storedProvider)
			cfg.APIKey = storedAPIKey

			// Validate the stored key
			cyan.Print("Validating API key... ")
			if err := validateAPIKeyWithProvider(cfg.Provider, cfg.APIKey); err != nil {
				fmt.Println()
				red.Printf("✗ API key validation failed: %v\n", err)
				yellow.Println("Please run 'trabuco auth login' to update your credentials.")
				return nil, fmt.Errorf("invalid API key: %w", err)
			}
			green.Println("✓")
		} else {
			hasStoredCredentials = false // Fall through to provider selection
		}
	}

	if !hasStoredCredentials {
		// 5a. AI Provider selection (all 4 providers)
		providerOptions := []string{
			"Anthropic (Claude) - Recommended for best quality",
			"OpenRouter - Multi-model gateway",
			"OpenAI (GPT-4o)",
			"Ollama (Local) - Free, self-hosted",
		}
		var providerChoice string
		if err := survey.AskOne(&survey.Select{
			Message: "AI Provider:",
			Options: providerOptions,
			Default: providerOptions[0],
			Help:    "Select your preferred LLM provider. Configure credentials with 'trabuco auth login' for persistent storage.",
		}, &providerChoice); err != nil {
			return nil, err
		}

		switch {
		case strings.HasPrefix(providerChoice, "Anthropic"):
			cfg.Provider = "anthropic"
		case strings.HasPrefix(providerChoice, "OpenRouter"):
			cfg.Provider = "openrouter"
		case strings.HasPrefix(providerChoice, "OpenAI"):
			cfg.Provider = "openai"
		case strings.HasPrefix(providerChoice, "Ollama"):
			cfg.Provider = "ollama"
		}

		// 5b. API Key (check env first, then prompt)
		selectedProvider := auth.Provider(cfg.Provider)
		providerInfo := auth.SupportedProviders[selectedProvider]

		// Skip API key for Ollama
		if !providerInfo.RequiresKey {
			cyan.Println("\nOllama doesn't require an API key.")
			fmt.Printf("Make sure Ollama is running at: %s\n", providerInfo.BaseURL)
			cfg.APIKey = "" // No key needed
		} else {
			// Check environment variable
			envKey := os.Getenv(providerInfo.EnvVar)

			if envKey != "" {
				maskedKey := maskAPIKey(envKey)
				var useEnvKey bool
				if err := survey.AskOne(&survey.Confirm{
					Message: fmt.Sprintf("Use API key from %s? (%s)", providerInfo.EnvVar, maskedKey),
					Default: true,
				}, &useEnvKey); err != nil {
					return nil, err
				}
				if useEnvKey {
					cfg.APIKey = envKey
				}
			}

			if cfg.APIKey == "" {
				// Prompt for API key
				fmt.Printf("\nGet your API key at: %s\n\n", providerInfo.DocumentURL)

				if err := survey.AskOne(&survey.Password{
					Message: fmt.Sprintf("API Key (%s):", providerInfo.EnvVar),
					Help:    "Your API key will be used for this session. Run 'trabuco auth login' to store it securely.",
				}, &cfg.APIKey, survey.WithValidator(validateAPIKey)); err != nil {
					return nil, err
				}
			}

			// Validate API key
			cyan.Print("Validating API key... ")
			if err := validateAPIKeyWithProvider(cfg.Provider, cfg.APIKey); err != nil {
				fmt.Println()
				red.Printf("✗ API key validation failed: %v\n", err)
				return nil, fmt.Errorf("invalid API key: %w", err)
			}
			green.Println("✓")

			// Offer to save credentials
			var saveCredentials bool
			if err := survey.AskOne(&survey.Confirm{
				Message: "Save these credentials for future use?",
				Default: true,
				Help:    "Credentials will be stored securely in your system keychain.",
			}, &saveCredentials); err != nil {
				return nil, err
			}

			if saveCredentials && manager != nil {
				cred := &auth.Credential{
					Provider: selectedProvider,
					APIKey:   cfg.APIKey,
				}
				if err := manager.SetCredential(cred, false); err != nil {
					yellow.Printf("Warning: Could not save credentials: %v\n", err)
				} else {
					green.Println("✓ Credentials saved to system keychain")
				}
			}
		}
	}

	// 7. Model selection (based on provider)
	var modelOptions []string
	var modelMap map[string]string

	switch cfg.Provider {
	case "anthropic", "openrouter":
		modelOptions = []string{
			"Claude Sonnet 4.5 (Recommended - Best balance of speed and quality)",
			"Claude Haiku 4.5 (Faster, lower cost)",
			"Claude Opus 4.6 (Highest quality, slower)",
		}
		modelMap = map[string]string{
			"Claude Sonnet 4.5": "claude-sonnet-4-5",
			"Claude Haiku 4.5":  "claude-haiku-4-5",
			"Claude Opus 4.6":   "claude-opus-4-6",
		}
	case "openai":
		modelOptions = []string{
			"GPT-4o (Recommended - Latest flagship model)",
			"GPT-4o-mini (Faster, lower cost)",
		}
		modelMap = map[string]string{
			"GPT-4o":      "gpt-4o",
			"GPT-4o-mini": "gpt-4o-mini",
		}
	case "ollama":
		modelOptions = []string{
			"Llama 3.3 (Recommended - Best open source model)",
			"CodeLlama (Optimized for code)",
			"Mistral (Fast and efficient)",
		}
		modelMap = map[string]string{
			"Llama 3.3":  "llama3.3",
			"CodeLlama":  "codellama",
			"Mistral":    "mistral",
		}
	default:
		// Fallback to Claude models
		modelOptions = []string{
			"Claude Sonnet 4.5 (Recommended)",
		}
		modelMap = map[string]string{
			"Claude Sonnet 4.5": "claude-sonnet-4-5",
		}
	}

	var modelChoice string
	if err := survey.AskOne(&survey.Select{
		Message: "AI Model:",
		Options: modelOptions,
		Default: modelOptions[0],
		Help:    "Select the model to use for code migration.",
	}, &modelChoice); err != nil {
		return nil, err
	}

	// Extract model name prefix for lookup
	for prefix, modelID := range modelMap {
		if strings.HasPrefix(modelChoice, prefix) {
			cfg.Model = modelID
			break
		}
	}

	// Show model info
	if cfg.Provider == "anthropic" || cfg.Provider == "openrouter" {
		yellow.Println("\nℹ Using model alias (auto-updates to latest version)")
		fmt.Printf("  To pin a specific version, use: --model=claude-sonnet-4-5-20250929\n")
	} else if cfg.Provider == "ollama" {
		yellow.Println("\nℹ Using local Ollama model (no API costs)")
		fmt.Println("  Make sure the model is pulled: ollama pull " + cfg.Model)
	}

	// 8. Additional options
	fmt.Println()
	yellow.Println("Migration Options")

	var dryRun bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Dry run? (Analyze only, don't generate files)",
		Default: false,
	}, &dryRun); err != nil {
		return nil, err
	}
	cfg.DryRun = dryRun

	if !dryRun {
		var includeTests bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Include test files in migration?",
			Default: false,
			Help:    "Migrate test classes along with source files",
		}, &includeTests); err != nil {
			return nil, err
		}
		cfg.IncludeTests = includeTests

		var skipBuild bool
		if err := survey.AskOne(&survey.Confirm{
			Message: "Skip Maven build after migration?",
			Default: false,
		}, &skipBuild); err != nil {
			return nil, err
		}
		cfg.SkipBuild = skipBuild
	}

	// 9. Show cost estimate
	fmt.Println()
	totalFiles := len(projectInfo.Entities) + len(projectInfo.Repositories) +
		len(projectInfo.Services) + len(projectInfo.Controllers) +
		len(projectInfo.ScheduledJobs) + len(projectInfo.EventListeners)

	// Rough estimate: ~500 input tokens per file, ~1000 output tokens
	estInputTokens := totalFiles * 500
	estOutputTokens := totalFiles * 1000

	// Pricing (per 1M tokens) - based on latest pricing
	var inputCost, outputCost float64
	switch {
	case strings.Contains(cfg.Model, "sonnet"):
		inputCost = 3.00
		outputCost = 15.00
	case strings.Contains(cfg.Model, "haiku"):
		inputCost = 1.00
		outputCost = 5.00
	case strings.Contains(cfg.Model, "opus"):
		inputCost = 5.00
		outputCost = 25.00
	case strings.Contains(cfg.Model, "gpt-4o-mini"):
		inputCost = 0.15
		outputCost = 0.60
	case strings.Contains(cfg.Model, "gpt-4o"):
		inputCost = 2.50
		outputCost = 10.00
	case cfg.Provider == "ollama":
		inputCost = 0.00
		outputCost = 0.00
	default:
		// Default to sonnet pricing
		inputCost = 3.00
		outputCost = 15.00
	}

	yellow.Println("─────────────────────────────────────────")
	yellow.Println("  Cost Estimate")
	yellow.Println("─────────────────────────────────────────")
	fmt.Printf("  Files to process: %d\n", totalFiles)
	fmt.Printf("  Est. tokens:      ~%d\n", estInputTokens+estOutputTokens)

	if cfg.Provider == "ollama" {
		green.Println("  Est. cost:        $0.00 (local model)")
	} else {
		estCost := (float64(estInputTokens)*inputCost + float64(estOutputTokens)*outputCost) / 1_000_000
		fmt.Printf("  Est. cost:        $%.2f - $%.2f\n", estCost*0.8, estCost*1.5)
	}
	yellow.Println("─────────────────────────────────────────")
	fmt.Println()

	// 10. Final confirmation
	var proceed bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Proceed with migration?",
		Default: true,
	}, &proceed); err != nil {
		return nil, err
	}
	if !proceed {
		return nil, fmt.Errorf("migration cancelled by user")
	}

	return cfg, nil
}

// Validators

func validateSourcePath(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input")
	}
	if str == "" {
		return fmt.Errorf("path is required")
	}

	// Check if directory exists
	info, err := os.Stat(str)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory does not exist")
	}
	if err != nil {
		return fmt.Errorf("cannot access directory: %v", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory")
	}

	// Check for pom.xml
	pomPath := filepath.Join(str, "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return fmt.Errorf("no pom.xml found - is this a Maven project?")
	}

	return nil
}

func validateAPIKey(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input")
	}
	if str == "" {
		return fmt.Errorf("API key is required")
	}
	if len(str) < 10 {
		return fmt.Errorf("API key seems too short")
	}
	return nil
}

// maskAPIKey returns a masked version of the API key for display
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "..." + key[len(key)-4:]
}

// validateAPIKeyWithProvider tests the API key by making a minimal request
func validateAPIKeyWithProvider(provider, apiKey string) error {
	cfg := &ai.ProviderConfig{
		APIKey:  apiKey,
		Timeout: 30, // 30 second timeout for validation
	}

	var p ai.Provider
	var err error

	switch provider {
	case "anthropic":
		p, err = ai.NewAnthropicProvider(cfg)
	case "openrouter":
		p, err = ai.NewOpenRouterProvider(cfg)
	case "openai":
		// TODO: Implement OpenAI provider
		// For now, just do basic validation
		if apiKey == "" || len(apiKey) < 10 {
			return fmt.Errorf("invalid API key format")
		}
		return nil
	case "ollama":
		// Ollama doesn't need API key validation
		// TODO: Check if Ollama server is running
		return nil
	default:
		p, err = ai.NewAnthropicProvider(cfg)
	}

	if err != nil {
		return err
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return p.ValidateAPIKey(ctx)
}

// getDatabaseDisplayName returns a user-friendly name for the database type
func getDatabaseDisplayName(database string) string {
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
		if database == "" {
			return "PostgreSQL (default)"
		}
		return database
	}
}
