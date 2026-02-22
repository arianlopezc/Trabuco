package prompts

import (
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/fatih/color"
)

// AddPromptResult contains the result of add prompts
type AddPromptResult struct {
	Module        string
	Database      string
	NoSQLDatabase string
	MessageBroker string
	Confirmed     bool
}

// RunAddPrompts runs interactive prompts for the add command
func RunAddPrompts(existingModules []string) (*AddPromptResult, error) {
	result := &AddPromptResult{}

	// Get available modules (filter out already existing)
	availableModules := getAvailableModulesToAdd(existingModules)
	if len(availableModules) == 0 {
		return nil, fmt.Errorf("all modules are already present in this project")
	}

	// Build module options
	options := buildModuleOptions(availableModules, existingModules)

	// Module selection
	var selectedOption string
	if err := survey.AskOne(&survey.Select{
		Message: "Select module to add:",
		Options: options,
	}, &selectedOption); err != nil {
		return nil, err
	}

	// Extract module name from option (before the " - ")
	result.Module = strings.Split(selectedOption, " - ")[0]

	// Prompt for database if adding SQLDatastore
	if result.Module == config.ModuleSQLDatastore {
		if err := survey.AskOne(&survey.Select{
			Message: "SQL Database:",
			Options: []string{
				"PostgreSQL (Recommended)",
				"MySQL",
				"Generic (bring your own driver)",
			},
			Default: "PostgreSQL (Recommended)",
		}, &result.Database); err != nil {
			return nil, err
		}
		result.Database = normalizeDatabaseChoice(result.Database)
	}

	// Prompt for NoSQL database if adding NoSQLDatastore
	if result.Module == config.ModuleNoSQLDatastore {
		// Check if Worker exists in current modules
		hasWorker := containsModule(existingModules, config.ModuleWorker)
		if hasWorker {
			yellow := color.New(color.FgYellow)
			yellow.Println("\n⚠ Worker module exists: Redis support is deprecated in JobRunr 8+.")
			fmt.Println("  If you select Redis, JobRunr will continue using its current storage.")
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
		}, &result.NoSQLDatabase); err != nil {
			return nil, err
		}
		result.NoSQLDatabase = normalizeNoSQLDatabaseChoice(result.NoSQLDatabase)
	}

	// Prompt for message broker if adding EventConsumer
	if result.Module == config.ModuleEventConsumer {
		if err := survey.AskOne(&survey.Select{
			Message: "Message Broker:",
			Options: []string{
				"Kafka (Recommended - High throughput, partitioned)",
				"RabbitMQ (Traditional message queue)",
				"AWS SQS (Managed queue service)",
				"GCP Pub/Sub (Google Cloud messaging)",
			},
			Default: "Kafka (Recommended - High throughput, partitioned)",
		}, &result.MessageBroker); err != nil {
			return nil, err
		}
		result.MessageBroker = normalizeMessageBrokerChoice(result.MessageBroker)
	}

	result.Confirmed = true
	return result, nil
}

// PromptModuleSelection prompts for module selection only
func PromptModuleSelection(existingModules []string) (string, error) {
	availableModules := getAvailableModulesToAdd(existingModules)
	if len(availableModules) == 0 {
		return "", fmt.Errorf("all modules are already present in this project")
	}

	options := buildModuleOptions(availableModules, existingModules)

	var selectedOption string
	if err := survey.AskOne(&survey.Select{
		Message: "Select module to add:",
		Options: options,
	}, &selectedOption); err != nil {
		return "", err
	}

	return strings.Split(selectedOption, " - ")[0], nil
}

// PromptDatabase prompts for SQL database selection
func PromptDatabase() (string, error) {
	var database string
	if err := survey.AskOne(&survey.Select{
		Message: "SQL Database:",
		Options: []string{
			"PostgreSQL (Recommended)",
			"MySQL",
			"Generic (bring your own driver)",
		},
		Default: "PostgreSQL (Recommended)",
	}, &database); err != nil {
		return "", err
	}
	return normalizeDatabaseChoice(database), nil
}

// PromptNoSQLDatabase prompts for NoSQL database selection
func PromptNoSQLDatabase(hasWorker bool) (string, error) {
	if hasWorker {
		yellow := color.New(color.FgYellow)
		yellow.Println("\n⚠ Worker module exists: Redis support is deprecated in JobRunr 8+.")
		fmt.Println("  MongoDB is recommended for Worker + NoSQLDatastore.")
		fmt.Println()
	}

	var database string
	if err := survey.AskOne(&survey.Select{
		Message: "NoSQL Database:",
		Options: []string{
			"MongoDB (Recommended - Document store)",
			"Redis (Key-Value store)",
		},
		Default: "MongoDB (Recommended - Document store)",
	}, &database); err != nil {
		return "", err
	}
	return normalizeNoSQLDatabaseChoice(database), nil
}

// PromptMessageBroker prompts for message broker selection
func PromptMessageBroker() (string, error) {
	var broker string
	if err := survey.AskOne(&survey.Select{
		Message: "Message Broker:",
		Options: []string{
			"Kafka (Recommended - High throughput, partitioned)",
			"RabbitMQ (Traditional message queue)",
			"AWS SQS (Managed queue service)",
			"GCP Pub/Sub (Google Cloud messaging)",
		},
		Default: "Kafka (Recommended - High throughput, partitioned)",
	}, &broker); err != nil {
		return "", err
	}
	return normalizeMessageBrokerChoice(broker), nil
}

// ConfirmAdd asks for confirmation before adding module
func ConfirmAdd(module string, dependencies []string) (bool, error) {
	message := fmt.Sprintf("Add %s module?", module)
	if len(dependencies) > 0 {
		message = fmt.Sprintf("Add %s module (will also add: %s)?", module, strings.Join(dependencies, ", "))
	}

	var confirmed bool
	if err := survey.AskOne(&survey.Confirm{
		Message: message,
		Default: true,
	}, &confirmed); err != nil {
		return false, err
	}
	return confirmed, nil
}

// getAvailableModulesToAdd returns modules that can be added
func getAvailableModulesToAdd(existingModules []string) []string {
	existingSet := make(map[string]bool)
	for _, m := range existingModules {
		existingSet[m] = true
	}

	// Check for mutual exclusion
	hasSQLDatastore := existingSet[config.ModuleSQLDatastore]
	hasNoSQLDatastore := existingSet[config.ModuleNoSQLDatastore]

	var available []string
	for _, m := range config.ModuleRegistry {
		// Skip internal modules (Jobs, Events) - they're auto-included
		if m.Internal {
			continue
		}

		// Skip already existing modules
		if existingSet[m.Name] {
			continue
		}

		// Skip required modules (Model is always present)
		if m.Required {
			continue
		}

		// Handle mutual exclusion
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

// buildModuleOptions creates display options for module selection
func buildModuleOptions(available []string, existing []string) []string {
	existingSet := make(map[string]bool)
	for _, m := range existing {
		existingSet[m] = true
	}

	var options []string
	for _, name := range available {
		m := config.GetModule(name)
		if m == nil {
			continue
		}

		desc := m.Description

		// Add notes for modules with auto-includes
		switch name {
		case config.ModuleWorker:
			if !existingSet[config.ModuleJobs] {
				desc += " (will add " + config.ModuleJobs + ")"
			}
		case config.ModuleEventConsumer:
			if !existingSet[config.ModuleEvents] {
				desc += " (will add " + config.ModuleEvents + ")"
			}
		}

		options = append(options, fmt.Sprintf("%s - %s", name, desc))
	}

	return options
}

// containsModule checks if a module name is in the list
func containsModule(modules []string, name string) bool {
	for _, m := range modules {
		if m == name {
			return true
		}
	}
	return false
}

// GetModuleDependencies returns the dependencies that will be added with a module
func GetModuleDependencies(module string, existingModules []string) []string {
	existingSet := make(map[string]bool)
	for _, m := range existingModules {
		existingSet[m] = true
	}

	var deps []string
	m := config.GetModule(module)
	if m != nil {
		for _, dep := range m.Dependencies {
			if !existingSet[dep] {
				deps = append(deps, dep)
			}
		}
	}

	return deps
}

// ValidateModuleCanBeAdded checks if a module can be added to the project
func ValidateModuleCanBeAdded(module string, existingModules []string) error {
	existingSet := make(map[string]bool)
	for _, m := range existingModules {
		existingSet[m] = true
	}

	// Check if already exists
	if existingSet[module] {
		return fmt.Errorf("module %s already exists in this project", module)
	}

	// Check mutual exclusion
	if module == config.ModuleSQLDatastore && existingSet[config.ModuleNoSQLDatastore] {
		return fmt.Errorf("cannot add %s: %s already exists (mutually exclusive)", config.ModuleSQLDatastore, config.ModuleNoSQLDatastore)
	}
	if module == config.ModuleNoSQLDatastore && existingSet[config.ModuleSQLDatastore] {
		return fmt.Errorf("cannot add %s: %s already exists (mutually exclusive)", config.ModuleNoSQLDatastore, config.ModuleSQLDatastore)
	}

	// Check if it's a valid module
	m := config.GetModule(module)
	if m == nil {
		return fmt.Errorf("unknown module: %s", module)
	}

	// Check if it's an internal module
	if m.Internal {
		return fmt.Errorf("cannot add %s directly: it's automatically included with %s", module, getParentModule(module))
	}

	return nil
}

// PromptCIProvider asks if the user wants to add a CI workflow
func PromptCIProvider() (string, error) {
	var addCI bool
	if err := survey.AskOne(&survey.Confirm{
		Message: "Would you like to add a GitHub Actions CI workflow?",
		Default: false,
	}, &addCI); err != nil {
		return "", err
	}
	if addCI {
		return "github", nil
	}
	return "", nil
}

// getParentModule returns the parent module that auto-includes an internal module
func getParentModule(internal string) string {
	switch internal {
	case config.ModuleJobs:
		return config.ModuleWorker
	case config.ModuleEvents:
		return config.ModuleEventConsumer
	default:
		return ""
	}
}
