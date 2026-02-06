package prompts

import (
	"fmt"
	"regexp"
	"strconv"
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
	if cfg.HasModule("SQLDatastore") {
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
	if cfg.HasModule("NoSQLDatastore") {
		// Show warning if Worker is selected - Redis has limited support
		if cfg.HasModule("Worker") {
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
		if cfg.HasModule("Worker") && cfg.NoSQLDatabase == "redis" {
			yellow := color.New(color.FgYellow)
			yellow.Println("\n⚠ Redis selected with Worker: JobRunr will use PostgreSQL for job storage.")
			fmt.Println("  A separate PostgreSQL instance will be added to docker-compose.yml.")
			fmt.Println()
		}
	}

	// 7. Message Broker (only if EventConsumer is selected)
	if cfg.HasModule("EventConsumer") {
		if err := survey.AskOne(&survey.Select{
			Message: "Message Broker:",
			Options: []string{
				"Kafka (Recommended - High throughput, partitioned)",
				"RabbitMQ (Traditional message queue)",
			},
			Default: "Kafka (Recommended - High throughput, partitioned)",
		}, &cfg.MessageBroker); err != nil {
			return nil, err
		}
		// Normalize message broker value
		cfg.MessageBroker = normalizeMessageBrokerChoice(cfg.MessageBroker)
	}

	// 8. CLAUDE.md (AI assistant context file)
	if err := survey.AskOne(&survey.Confirm{
		Message: "Generate CLAUDE.md?",
		Default: false,
		Help:    "Creates a context file for AI assistants (Claude Code) with project-specific commands and conventions.",
	}, &cfg.IncludeCLAUDEMD); err != nil {
		return nil, err
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
		if name == "SQLDatastore" {
			hasSQLDatastore = true
		}
		if name == "NoSQLDatastore" {
			hasNoSQLDatastore = true
		}
	}
	if hasSQLDatastore && hasNoSQLDatastore {
		return fmt.Errorf("SQLDatastore and NoSQLDatastore cannot be selected together")
	}

	return nil
}

func normalizeDatabaseChoice(choice string) string {
	switch {
	case strings.HasPrefix(choice, "PostgreSQL"):
		return "postgresql"
	case strings.HasPrefix(choice, "MySQL"):
		return "mysql"
	default:
		return "generic"
	}
}

func normalizeNoSQLDatabaseChoice(choice string) string {
	switch {
	case strings.HasPrefix(choice, "MongoDB"):
		return "mongodb"
	case strings.HasPrefix(choice, "Redis"):
		return "redis"
	default:
		return "mongodb"
	}
}

func normalizeMessageBrokerChoice(choice string) string {
	switch {
	case strings.HasPrefix(choice, "Kafka"):
		return "kafka"
	case strings.HasPrefix(choice, "RabbitMQ"):
		return "rabbitmq"
	default:
		return "kafka"
	}
}

// promptJavaVersion prompts for Java version with detection status
func promptJavaVersion(detection *java.DetectionResult) (version string, detected bool, err error) {
	// Build options with detection status
	options := buildJavaOptions(detection)

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Java version:",
		Options: options,
		Default: options[0], // First option is recommended
	}, &selected); err != nil {
		return "", false, err
	}

	// Extract version number from selection
	version = strings.Split(selected, " ")[0]
	versionInt, _ := strconv.Atoi(version)
	detected = detection.IsVersionDetected(versionInt)

	// If version not detected, show confirmation prompt
	if !detected {
		confirmed, err := confirmUndetectedVersion(version, detection)
		if err != nil {
			return "", false, err
		}
		if !confirmed {
			// User chose not to proceed, re-prompt with detected versions only
			return promptJavaVersionDetectedOnly(detection)
		}
	}

	return version, detected, nil
}

// buildJavaOptions creates the Java version options with detection status
func buildJavaOptions(detection *java.DetectionResult) []string {
	type versionOption struct {
		version     int
		label       string
		recommended bool
	}

	// Define available versions with descriptions
	allVersions := []versionOption{
		{21, "21 (LTS until 2031 - Recommended)", true},
		{25, "25 (Latest LTS)", false},
		{17, "17 (LTS - Minimum supported)", false},
	}

	var options []string
	var hasRecommended bool

	// Add versions with detection status
	for _, v := range allVersions {
		// Only show 17 if it's detected and 21/25 are not detected
		if v.version == 17 {
			if !detection.IsVersionDetected(17) {
				continue
			}
			if detection.IsVersionDetected(21) || detection.IsVersionDetected(25) {
				continue
			}
		}

		status := "[not detected]"
		if detection.IsVersionDetected(v.version) {
			status = "[detected]"
		}
		option := fmt.Sprintf("%d %s %s", v.version, strings.TrimPrefix(v.label, fmt.Sprintf("%d ", v.version)), status)
		options = append(options, option)

		if v.recommended && detection.IsVersionDetected(v.version) {
			hasRecommended = true
		}
	}

	// If recommended version (21) is not detected but another is, reorder to put detected first
	if !hasRecommended && len(options) > 0 {
		// Find first detected version and move to front
		for i, opt := range options {
			if strings.Contains(opt, "[detected]") {
				// Move to front
				options = append([]string{opt}, append(options[:i], options[i+1:]...)...)
				break
			}
		}
	}

	return options
}

// confirmUndetectedVersion asks user to confirm using an undetected Java version
func confirmUndetectedVersion(version string, detection *java.DetectionResult) (bool, error) {
	yellow := color.New(color.FgYellow)
	yellow.Printf("\n\u26a0 Java %s was not detected on your system.\n\n", version)

	detectedVersions := detection.GetDetectedVersions()
	if len(detectedVersions) > 0 {
		fmt.Printf("Detected compatible versions: %s\n\n", java.FormatDetectedVersions(detectedVersions))
	} else {
		fmt.Println("No compatible Java versions detected.")
	}

	var confirmed bool
	if err := survey.AskOne(&survey.Confirm{
		Message: fmt.Sprintf("Continue with Java %s anyway?", version),
		Default: false,
	}, &confirmed); err != nil {
		return false, err
	}

	return confirmed, nil
}

// promptJavaVersionDetectedOnly prompts for Java version showing only detected versions
func promptJavaVersionDetectedOnly(detection *java.DetectionResult) (string, bool, error) {
	detectedVersions := detection.GetDetectedVersions()
	if len(detectedVersions) == 0 {
		// No detected versions, fall back to 21
		return "21", false, nil
	}

	var options []string
	for _, v := range detectedVersions {
		label := strconv.Itoa(v)
		switch v {
		case 21:
			label = "21 (LTS until 2031 - Recommended)"
		case 25:
			label = "25 (Latest LTS)"
		case 17:
			label = "17 (LTS - Minimum supported)"
		}
		options = append(options, label)
	}

	var selected string
	if err := survey.AskOne(&survey.Select{
		Message: "Select a detected Java version:",
		Options: options,
	}, &selected); err != nil {
		return "", false, err
	}

	version := strings.Split(selected, " ")[0]
	return version, true, nil
}
