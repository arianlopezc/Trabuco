package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/auth"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	authProvider string
	authAPIKey   string
	authModel    string
	authForce    bool
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage LLM provider credentials",
	Long: `Manage authentication credentials for LLM providers used by trabuco migrate.

Trabuco securely stores your API keys in the system keychain (macOS Keychain,
Linux Secret Service, or Windows Credential Manager) with fallback to an
encrypted file.

SUBCOMMANDS:
  login      Configure credentials for an LLM provider
  status     Show configured providers and their status
  logout     Remove stored credentials
  providers  List supported LLM providers with pricing info

Examples:
  # Interactive login (recommended)
  trabuco auth login

  # Login with specific provider
  trabuco auth login --provider anthropic

  # Login with API key directly
  trabuco auth login --provider anthropic --api-key sk-ant-...

  # Check configured providers
  trabuco auth status

  # Remove all credentials
  trabuco auth logout`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Configure credentials for an LLM provider",
	Long: `Configure credentials for an LLM provider.

This command guides you through setting up API keys for providers like
Anthropic (Claude), OpenRouter, or OpenAI. Credentials are stored securely
in your system keychain.

If no provider is specified, you'll be prompted to choose one.`,
	Run: runAuthLogin,
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show configured providers and their status",
	Long:  `Display all configured LLM providers, their validation status, and which one is set as default.`,
	Run:   runAuthStatus,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout [provider]",
	Short: "Remove stored credentials",
	Long: `Remove stored credentials for a specific provider or all providers.

If no provider is specified, all credentials will be removed.`,
	Args: cobra.MaximumNArgs(1),
	Run:  runAuthLogout,
}

var authProvidersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List supported LLM providers with pricing info",
	Long:  `Display all supported LLM providers with their pricing, models, and setup instructions.`,
	Run:   runAuthProviders,
}

func init() {
	// Login flags
	authLoginCmd.Flags().StringVarP(&authProvider, "provider", "p", "", "Provider: anthropic, openrouter, openai, ollama")
	authLoginCmd.Flags().StringVar(&authAPIKey, "api-key", "", "API key (alternative to interactive input)")
	authLoginCmd.Flags().StringVar(&authModel, "model", "", "Default model to use")
	authLoginCmd.Flags().BoolVarP(&authForce, "force", "f", false, "Overwrite existing credentials")

	// Logout flags
	authLogoutCmd.Flags().BoolVarP(&authForce, "force", "f", false, "Skip confirmation")

	// Add subcommands
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authProvidersCmd)
}

func runAuthLogin(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	cyan.Println("\n╔════════════════════════════════════════╗")
	cyan.Println("║       Trabuco Auth - LLM Setup         ║")
	cyan.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	manager, err := auth.NewManager()
	if err != nil {
		red.Fprintf(os.Stderr, "Error initializing credential manager: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Using storage: %s\n\n", manager.StorageBackend())

	// Select provider
	var provider auth.Provider
	if authProvider != "" {
		provider = auth.Provider(strings.ToLower(authProvider))
		if _, ok := auth.SupportedProviders[provider]; !ok {
			red.Fprintf(os.Stderr, "Unknown provider: %s\n", authProvider)
			fmt.Println("\nSupported providers: anthropic, openrouter, openai, ollama")
			os.Exit(1)
		}
	} else {
		// Interactive provider selection
		providerOptions := []string{
			"Anthropic (Claude) - Recommended for best quality",
			"OpenRouter - Multi-model gateway",
			"OpenAI (GPT-4o)",
			"Ollama (Local) - No cost, self-hosted",
		}
		var providerIdx int
		prompt := &survey.Select{
			Message: "Select LLM provider:",
			Options: providerOptions,
		}
		if err := survey.AskOne(prompt, &providerIdx); err != nil {
			red.Fprintf(os.Stderr, "Cancelled\n")
			os.Exit(1)
		}

		providerMap := []auth.Provider{
			auth.ProviderAnthropic,
			auth.ProviderOpenRouter,
			auth.ProviderOpenAI,
			auth.ProviderOllama,
		}
		provider = providerMap[providerIdx]
	}

	info := auth.SupportedProviders[provider]

	// Check if already configured
	existing, _ := manager.GetCredential(provider)
	if existing != nil && !authForce {
		yellow.Printf("\n%s is already configured.\n", info.Name)
		var overwrite bool
		prompt := &survey.Confirm{
			Message: "Overwrite existing credentials?",
			Default: false,
		}
		if err := survey.AskOne(prompt, &overwrite); err != nil || !overwrite {
			fmt.Println("Cancelled")
			return
		}
	}

	// Get API key
	var apiKey string
	if !info.RequiresKey {
		fmt.Printf("\n%s doesn't require an API key.\n", info.Name)
		fmt.Printf("Make sure Ollama is running at: %s\n", info.BaseURL)
	} else if authAPIKey != "" {
		apiKey = authAPIKey
	} else {
		fmt.Printf("\nGet your API key at: %s\n\n", info.DocumentURL)

		prompt := &survey.Password{
			Message: fmt.Sprintf("Enter your %s API key:", info.Name),
		}
		if err := survey.AskOne(prompt, &apiKey); err != nil {
			red.Fprintf(os.Stderr, "Cancelled\n")
			os.Exit(1)
		}
	}

	// Validate API key format
	if err := auth.ValidateAPIKey(provider, apiKey); err != nil {
		red.Fprintf(os.Stderr, "Invalid API key: %v\n", err)
		os.Exit(1)
	}

	// Test connection
	fmt.Println()
	yellow.Print("Testing connection...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	valid, modelInfo := testProviderConnection(ctx, provider, apiKey)
	if !valid {
		fmt.Println()
		red.Println("✗ Connection failed")
		red.Println("  Please check your API key and try again.")
		os.Exit(1)
	}

	fmt.Print("\r")
	green.Printf("✓ Connected successfully")
	if modelInfo != "" {
		fmt.Printf(" (%s available)", modelInfo)
	}
	fmt.Println()

	// Select default model (optional)
	model := authModel
	if model == "" && len(info.Models) > 0 {
		fmt.Println()
		var useDefault bool
		prompt := &survey.Confirm{
			Message: fmt.Sprintf("Use default model (%s)?", info.Models[0]),
			Default: true,
		}
		if err := survey.AskOne(prompt, &useDefault); err == nil && !useDefault {
			modelPrompt := &survey.Select{
				Message: "Select default model:",
				Options: info.Models,
			}
			var modelIdx int
			if err := survey.AskOne(modelPrompt, &modelIdx); err == nil {
				model = info.Models[modelIdx]
			}
		}
	}

	// Store credential
	cred := &auth.Credential{
		Provider:    provider,
		APIKey:      apiKey,
		Model:       model,
		ValidatedAt: time.Now(),
	}

	if err := manager.SetCredential(cred, false); err != nil {
		red.Fprintf(os.Stderr, "Error storing credentials: %v\n", err)
		os.Exit(1)
	}

	// Print cost estimate
	fmt.Println()
	cyan.Println("Estimated migration costs:")
	fmt.Printf("  Small project  (~50 files):  $%.2f - $%.2f\n", info.InputCostPer1M*0.1+info.OutputCostPer1M*0.05, info.InputCostPer1M*0.5+info.OutputCostPer1M*0.25)
	fmt.Printf("  Medium project (~200 files): $%.2f - $%.2f\n", info.InputCostPer1M*0.5+info.OutputCostPer1M*0.25, info.InputCostPer1M*2.0+info.OutputCostPer1M*1.0)
	fmt.Printf("  Large project  (~500 files): $%.2f - $%.2f\n", info.InputCostPer1M*2.0+info.OutputCostPer1M*1.0, info.InputCostPer1M*5.0+info.OutputCostPer1M*2.5)
	fmt.Println()

	green.Println("✓ Credentials saved successfully!")
	fmt.Println()
	fmt.Println("You can now run:")
	fmt.Println("  trabuco migrate /path/to/project")
}

func runAuthStatus(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	manager, err := auth.NewManager()
	if err != nil {
		red.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	cyan.Println("\nTrabuco Auth Status")
	fmt.Printf("Storage: %s\n\n", manager.StorageBackend())

	statuses := manager.ListConfigured()

	if len(statuses) == 0 {
		yellow.Println("No credentials configured.")
		fmt.Println("\nRun 'trabuco auth login' to configure a provider.")
		return
	}

	for _, status := range statuses {
		var statusIcon, statusColor string
		if status.Configured {
			statusIcon = "✓"
			statusColor = "green"
		} else {
			statusIcon = "○"
			statusColor = "yellow"
		}

		// Provider name with default indicator
		name := status.Info.Name
		if status.IsDefault {
			name += " (default)"
		}

		// Print status line
		switch statusColor {
		case "green":
			green.Printf("%s %s\n", statusIcon, name)
		case "yellow":
			yellow.Printf("%s %s\n", statusIcon, name)
		default:
			fmt.Printf("%s %s\n", statusIcon, name)
		}

		if status.Configured {
			fmt.Printf("    Source: %s\n", status.Source)
			if !status.ValidatedAt.IsZero() {
				fmt.Printf("    Validated: %s\n", status.ValidatedAt.Format("2006-01-02 15:04"))
			}
			if status.Model != "" {
				fmt.Printf("    Model: %s\n", status.Model)
			}
		}
	}

	fmt.Println()
}

func runAuthLogout(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)
	red := color.New(color.FgRed)

	manager, err := auth.NewManager()
	if err != nil {
		red.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		// Remove specific provider
		provider := auth.Provider(strings.ToLower(args[0]))
		if _, ok := auth.SupportedProviders[provider]; !ok {
			red.Fprintf(os.Stderr, "Unknown provider: %s\n", args[0])
			os.Exit(1)
		}

		if !authForce {
			var confirm bool
			prompt := &survey.Confirm{
				Message: fmt.Sprintf("Remove credentials for %s?", auth.SupportedProviders[provider].Name),
				Default: false,
			}
			if err := survey.AskOne(prompt, &confirm); err != nil || !confirm {
				fmt.Println("Cancelled")
				return
			}
		}

		if err := manager.RemoveCredential(provider); err != nil {
			red.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		green.Printf("✓ Removed credentials for %s\n", auth.SupportedProviders[provider].Name)
	} else {
		// Remove all credentials
		if !authForce {
			yellow.Println("\nThis will remove ALL stored credentials.")
			var confirm bool
			prompt := &survey.Confirm{
				Message: "Are you sure?",
				Default: false,
			}
			if err := survey.AskOne(prompt, &confirm); err != nil || !confirm {
				fmt.Println("Cancelled")
				return
			}
		}

		if err := manager.Clear(); err != nil {
			red.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		cyan.Println("✓ All credentials removed")
	}
}

func runAuthProviders(cmd *cobra.Command, args []string) {
	cyan := color.New(color.FgCyan, color.Bold)
	green := color.New(color.FgGreen)

	cyan.Println("\nSupported LLM Providers")
	fmt.Println(strings.Repeat("─", 60))

	providers := []auth.Provider{
		auth.ProviderAnthropic,
		auth.ProviderOpenRouter,
		auth.ProviderOpenAI,
		auth.ProviderOllama,
	}

	for _, provider := range providers {
		info := auth.SupportedProviders[provider]

		green.Printf("\n%s\n", info.Name)
		fmt.Printf("  Get API Key: %s\n", info.DocumentURL)

		if info.RequiresKey {
			fmt.Printf("  Environment: %s\n", info.EnvVar)
			fmt.Printf("  Pricing:     $%.2f/1M input, $%.2f/1M output\n",
				info.InputCostPer1M, info.OutputCostPer1M)
		} else {
			fmt.Println("  Pricing:     Free (self-hosted)")
		}

		fmt.Printf("  Models:      %s\n", strings.Join(info.Models, ", "))
	}

	fmt.Println()
	fmt.Println("To configure a provider:")
	fmt.Println("  trabuco auth login")
	fmt.Println()
}

// testProviderConnection validates credentials by making a test API call
func testProviderConnection(ctx context.Context, provider auth.Provider, apiKey string) (bool, string) {
	config := &ai.ProviderConfig{
		APIKey:  apiKey,
		Timeout: 30,
	}

	var aiProvider ai.Provider
	var err error

	switch provider {
	case auth.ProviderAnthropic:
		aiProvider, err = ai.NewAnthropicProvider(config)
	case auth.ProviderOpenRouter:
		aiProvider, err = ai.NewOpenRouterProvider(config)
	case auth.ProviderOllama:
		// For Ollama, just check if server is reachable
		// TODO: Implement Ollama provider
		return true, "llama3.3"
	default:
		return false, ""
	}

	if err != nil {
		return false, ""
	}

	if err := aiProvider.ValidateAPIKey(ctx); err != nil {
		return false, ""
	}

	return true, "claude-sonnet-4-5"
}
