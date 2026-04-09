package prompts

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/java"
	"github.com/fatih/color"
)

// projectNameRegex validates project names: lowercase, alphanumeric, hyphens allowed (not at start/end)
var projectNameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

// groupIDRegex validates group IDs: valid Java package format
var groupIDRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$`)

// RunPrompts runs the interactive prompts and returns a ProjectConfig
func RunPrompts() (*config.ProjectConfig, error) {
	cfg := &config.ProjectConfig{}

	// 1. Project name
	if err := survey.AskOne(&survey.Input{
		Message: "Project name:",
		Help:    "Lowercase, alphanumeric, hyphens allowed (e.g., my-platform)",
	}, &cfg.ProjectName, survey.WithValidator(validateProjectName)); err != nil {
		return nil, err
	}
	cfg.ArtifactID = cfg.ProjectName

	// 2. Group ID
	defaultGroupID := "com.company." + strings.ReplaceAll(cfg.ProjectName, "-", "")
	if err := survey.AskOne(&survey.Input{
		Message: "Group ID:",
		Default: defaultGroupID,
		Help:    "Java package format (e.g., com.company.project)",
	}, &cfg.GroupID, survey.WithValidator(validateGroupID)); err != nil {
		return nil, err
	}

	// 3. Module selection
	moduleOptions := config.GetModuleDisplayOptions()
	var selectedIndices []int
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Select modules to include:",
		Options: moduleOptions,
		Default: []int{0}, // Only Model selected by default
		Help:    "Use space to toggle, enter to confirm. At least one module is required.",
	}, &selectedIndices, survey.WithValidator(validateModuleSelection)); err != nil {
		return nil, err
	}

	// Convert indices to module names (use selectable modules to match display options)
	selectableModules := config.GetSelectableModules()
	selectedModules := make([]string, len(selectedIndices))
	for i, idx := range selectedIndices {
		selectedModules[i] = selectableModules[idx]
	}

	// Resolve dependencies (only adds Model if not selected)
	cfg.Modules = config.ResolveDependencies(selectedModules)

	// 4. Java version with detection
	javaDetection := java.Detect()
	javaVersion, javaDetected, err := promptJavaVersion(javaDetection)
	if err != nil {
		return nil, err
	}
	cfg.JavaVersion = javaVersion
	cfg.JavaVersionDetected = javaDetected

	// 5. SQL Database (only if SQLDatastore is selected)
	if cfg.HasModule(config.ModuleSQLDatastore) {
		if err := survey.AskOne(&survey.Select{
			Message: "SQL Database:",
			Options: []string{
				"PostgreSQL (Recommended)",
				"MySQL",
				"Generic (bring your own driver)",
			},
			Default: "PostgreSQL (Recommended)",
		}, &cfg.Database); err != nil {
			return nil, err
		}
		// Normalize database value
		cfg.Database = normalizeDatabaseChoice(cfg.Database)
	}

	// 6. NoSQL Database (only if NoSQLDatastore is selected)
	if cfg.HasModule(config.ModuleNoSQLDatastore) {
		// Show warning if Worker is selected - Redis has limited support
		if cfg.HasModule(config.ModuleWorker) {
			yellow := color.New(color.FgYellow)
			yellow.Println("\n⚠ Worker module note: Redis support is deprecated in JobRunr 8+.")
			fmt.Println("  If you select Redis, JobRunr will use PostgreSQL for job storage.")
			fmt.Println("  MongoDB is recommended for Worker + NoSQLDatastore.")
			fmt.Println()
		}

		if err := survey.AskOne(&survey.Select{
			Message: "NoSQL Database:",
			Options: []string{
				"MongoDB (Recommended - Document store)",
				"Redis (Key-Value store)",
			},
			Default: "MongoDB (Recommended - Document store)",
		}, &cfg.NoSQLDatabase); err != nil {
			return nil, err
		}
		// Normalize NoSQL database value
		cfg.NoSQLDatabase = normalizeNoSQLDatabaseChoice(cfg.NoSQLDatabase)

		// Additional warning if Redis was selected with Worker
		if cfg.HasModule(config.ModuleWorker) && cfg.NoSQLDatabase == config.DatabaseRedis {
			yellow := color.New(color.FgYellow)
			yellow.Println("\n⚠ Redis selected with Worker: JobRunr will use PostgreSQL for job storage.")
			fmt.Println("  A separate PostgreSQL instance will be added to docker-compose.yml.")
			fmt.Println()
		}
	}

	// 7. Message Broker (only if EventConsumer is selected)
	if cfg.HasModule(config.ModuleEventConsumer) {
		if err := survey.AskOne(&survey.Select{
			Message: "Message Broker:",
			Options: []string{
				"Kafka (Recommended - High throughput, partitioned)",
				"RabbitMQ (Traditional message queue)",
				"AWS SQS (Managed queue service)",
				"GCP Pub/Sub (Google Cloud messaging)",
			},
			Default: "Kafka (Recommended - High throughput, partitioned)",
		}, &cfg.MessageBroker); err != nil {
			return nil, err
		}
		// Normalize message broker value
		cfg.MessageBroker = normalizeMessageBrokerChoice(cfg.MessageBroker)
	}

	// 8. AI coding agent context files
	agentOptions := config.GetAIAgentDisplayOptions()
	var selectedAgentIndices []int
	if err := survey.AskOne(&survey.MultiSelect{
		Message: "Generate AI agent context files:",
		Options: agentOptions,
		Default: []int{}, // None selected by default
		Help:    "Creates context files with project-specific commands and conventions for AI coding assistants.",
	}, &selectedAgentIndices); err != nil {
		return nil, err
	}

	// Convert indices to agent IDs
	allAgents := config.GetAvailableAIAgents()
	cfg.AIAgents = make([]string, len(selectedAgentIndices))
	for i, idx := range selectedAgentIndices {
		cfg.AIAgents[i] = allAgents[idx].ID
	}

	// 9. CI/CD provider
	ciOptions := append(config.GetCIProviderDisplayOptions(), "None - Skip CI configuration")
	var ciChoice string
	if err := survey.AskOne(&survey.Select{
		Message: "CI/CD provider:",
		Options: ciOptions,
		Default: "None - Skip CI configuration",
		Help:    "Generate a CI workflow for your repository.",
	}, &ciChoice); err != nil {
		return nil, err
	}
	// Extract provider ID if not "None"
	if !strings.HasPrefix(ciChoice, "None") {
		providers := config.GetAvailableCIProviders()
		for _, p := range providers {
			if strings.HasPrefix(ciChoice, p.Name) {
				cfg.CIProvider = p.ID
				break
			}
		}
	}

	return cfg, nil
}

// Validators

func validateProjectName(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input")
	}
	if str == "" {
		return fmt.Errorf("project name is required")
	}
	if !projectNameRegex.MatchString(str) {
		return fmt.Errorf("must be lowercase, alphanumeric, hyphens allowed (not at start/end)")
	}
	return nil
}

func validateGroupID(val interface{}) error {
	str, ok := val.(string)
	if !ok {
		return fmt.Errorf("invalid input")
	}
	if str == "" {
		return fmt.Errorf("group ID is required")
	}
	if !groupIDRegex.MatchString(str) {
		return fmt.Errorf("must be valid Java package (e.g., com.company.project)")
	}
	return nil
}

func validateModuleSelection(val interface{}) error {
	var selectedIndices []int

	// Handle different survey response types
	switch v := val.(type) {
	case []survey.OptionAnswer:
		if len(v) == 0 {
			return fmt.Errorf("at least one module must be selected")
		}
		for _, opt := range v {
			selectedIndices = append(selectedIndices, opt.Index)
		}
	case []int:
		if len(v) == 0 {
			return fmt.Errorf("at least one module must be selected")
		}
		selectedIndices = v
	default:
		return fmt.Errorf("invalid selection")
	}

	// Convert indices to module names and check for conflicts
	// Use selectable modules (excludes internal) to match display option indices
	selectableModules := config.GetSelectableModules()
	var selectedModules []string
	for _, idx := range selectedIndices {
		if idx < len(selectableModules) {
			selectedModules = append(selectedModules, selectableModules[idx])
		}
	}

	// Check for SQLDatastore + NoSQLDatastore conflict
	hasSQLDatastore := false
	hasNoSQLDatastore := false
	for _, name := range selectedModules {
		if name == config.ModuleSQLDatastore {
			hasSQLDatastore = true
		}
		if name == config.ModuleNoSQLDatastore {
			hasNoSQLDatastore = true
		}
	}
	if hasSQLDatastore && hasNoSQLDatastore {
		return fmt.Errorf("%s and %s cannot be selected together", config.ModuleSQLDatastore, config.ModuleNoSQLDatastore)
	}

	return nil
}

func normalizeDatabaseChoice(choice string) string {
	switch {
	case strings.HasPrefix(choice, "PostgreSQL"):
		return config.DatabasePostgreSQL
	case strings.HasPrefix(choice, "MySQL"):
		return config.DatabaseMySQL
	default:
		return "generic"
	}
}

func normalizeNoSQLDatabaseChoice(choice string) string {
	switch {
	case strings.HasPrefix(choice, "MongoDB"):
		return config.DatabaseMongoDB
	case strings.HasPrefix(choice, "Redis"):
		return config.DatabaseRedis
	default:
		return config.DatabaseMongoDB
	}
}

func normalizeMessageBrokerChoice(choice string) string {
	switch {
	case strings.HasPrefix(choice, "Kafka"):
		return config.BrokerKafka
	case strings.HasPrefix(choice, "RabbitMQ"):
		return config.BrokerRabbitMQ
	case strings.HasPrefix(choice, "AWS SQS"):
		return config.BrokerSQS
	case strings.HasPrefix(choice, "GCP Pub/Sub"):
		return config.BrokerPubSub
	default:
		return config.BrokerKafka
	}
}

// promptJavaVersion prompts for Java version showing only detected supported versions.
// If no supported version is detected, returns an error asking the user to install one.
func promptJavaVersion(detection *java.DetectionResult) (version string, detected bool, err error) {
	options := buildJavaOptions(detection)
	if len(options) == 0 {
		red := color.New(color.FgRed)
		red.Println("\nNo supported Java version detected on your system.")
		fmt.Println()
		fmt.Printf("Trabuco requires Java %d or later. Supported versions: %s\n",
			java.MinSupportedVersion, java.FormatDetectedVersions(java.SupportedVersions))
		fmt.Println()
		fmt.Println("Install one of the supported versions and try again:")
		fmt.Println("  • SDKMAN:  sdk install java 21-tem")
		fmt.Println("  • Homebrew: brew install openjdk@21")
		fmt.Println("  • Manual:  https://adoptium.net/")
		return "", false, fmt.Errorf("no supported Java version detected (minimum: %d)", java.MinSupportedVersion)
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Java version:",
		Options: options,
		Default: options[0],
	}, &selected); err != nil {
		return "", false, err
	}

	version = strings.Split(selected, " ")[0]
	return version, true, nil
}

// buildJavaOptions creates Java version options from detected supported versions only.
// Returns options sorted by preference: 21 (recommended) first, then others descending.
func buildJavaOptions(detection *java.DetectionResult) []string {
	// Labels for known versions
	labels := map[int]string{
		21: "LTS until 2031 — Recommended",
		25: "LTS",
		26: "Latest",
	}

	// Collect detected versions that are in our supported list
	var detected []int
	for _, v := range java.SupportedVersions {
		if detection.IsVersionDetected(v) {
			detected = append(detected, v)
		}
	}

	// Build options — put 21 first if present (recommended), then others descending
	var options []string
	for _, v := range detected {
		label := labels[v]
		if label == "" {
			label = "Supported"
		}
		options = append(options, fmt.Sprintf("%d (%s)", v, label))
	}

	return options
}
