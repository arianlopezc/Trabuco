package config

// Module name constants - use these instead of string literals
const (
	ModuleModel          = "Model"
	ModuleJobs           = "Jobs"
	ModuleSQLDatastore   = "SQLDatastore"
	ModuleNoSQLDatastore = "NoSQLDatastore"
	ModuleShared         = "Shared"
	ModuleAPI            = "API"
	ModuleWorker         = "Worker"
	ModuleEvents         = "Events"
	ModuleEventConsumer  = "EventConsumer"
	ModuleMCP            = "MCP"
)

// Database type constants
const (
	DatabasePostgreSQL = "postgresql"
	DatabaseMySQL      = "mysql"
	DatabaseMongoDB    = "mongodb"
	DatabaseRedis      = "redis"
)

// Message broker constants
const (
	BrokerKafka    = "kafka"
	BrokerRabbitMQ = "rabbitmq"
	BrokerSQS      = "sqs"
	BrokerPubSub   = "pubsub"
)

// Module represents a project module with its metadata
type Module struct {
	Name         string   // Module name (e.g., "Model")
	Description  string   // Short description for CLI display
	Required     bool     // If true, cannot be deselected
	Internal     bool     // If true, not shown in CLI prompts (auto-included when needed)
	Dependencies []string // Names of modules this depends on (only Model is a real dependency)
}

// ModuleRegistry contains all available modules
// NOTE: Only Model is a hard dependency for other modules.
// Other modules can work independently - if a module is selected without
// another, the generated code will have TODO placeholders instead of
// actual integrations.
var ModuleRegistry = []Module{
	{
		Name:         ModuleModel,
		Description:  "DTOs, Entities, Enums, Exceptions",
		Required:     true,
		Internal:     false,
		Dependencies: []string{},
	},
	{
		Name:         ModuleJobs,
		Description:  "Job request contracts for background processing",
		Required:     false,
		Internal:     true, // Auto-included when Worker is selected
		Dependencies: []string{ModuleModel},
	},
	{
		Name:         ModuleSQLDatastore,
		Description:  "SQL repositories, Flyway migrations (PostgreSQL, MySQL) - exclusive with NoSQLDatastore",
		Required:     false,
		Internal:     false,
		Dependencies: []string{ModuleModel},
	},
	{
		Name:         ModuleNoSQLDatastore,
		Description:  "NoSQL repositories (MongoDB, Redis) - exclusive with SQLDatastore",
		Required:     false,
		Internal:     false,
		Dependencies: []string{ModuleModel},
	},
	{
		Name:         ModuleShared,
		Description:  "Services, Circuit breaker, Utilities",
		Required:     false,
		Internal:     false,
		Dependencies: []string{ModuleModel},
	},
	{
		Name:         ModuleAPI,
		Description:  "REST endpoints, Validation",
		Required:     false,
		Internal:     false,
		Dependencies: []string{ModuleModel},
	},
	{
		Name:         ModuleWorker,
		Description:  "Background jobs (fire-and-forget, scheduled, delayed, batch)",
		Required:     false,
		Internal:     false,
		Dependencies: []string{ModuleModel, ModuleJobs}, // Jobs is auto-included; uses datastore for JobRunr persistence (defaults to PostgreSQL if none)
	},
	{
		Name:         ModuleEvents,
		Description:  "Event contracts for event-driven processing",
		Required:     false,
		Internal:     true, // Auto-included when EventConsumer is selected
		Dependencies: []string{ModuleModel},
	},
	{
		Name:         ModuleEventConsumer,
		Description:  "Event listeners (Kafka, RabbitMQ)",
		Required:     false,
		Internal:     false,
		Dependencies: []string{ModuleModel, ModuleEvents}, // Events is auto-included
	},
	{
		Name:         ModuleMCP,
		Description:  "MCP server for AI tool integration (build, test, introspection)",
		Required:     false,
		Internal:     false,
		Dependencies: []string{ModuleModel},
	},
}

// GetModule returns a module by name, or nil if not found
func GetModule(name string) *Module {
	for i := range ModuleRegistry {
		if ModuleRegistry[i].Name == name {
			return &ModuleRegistry[i]
		}
	}
	return nil
}

// GetModuleNames returns all module names
func GetModuleNames() []string {
	names := make([]string, len(ModuleRegistry))
	for i, m := range ModuleRegistry {
		names[i] = m.Name
	}
	return names
}

// GetRequiredModules returns names of all required modules
func GetRequiredModules() []string {
	var required []string
	for _, m := range ModuleRegistry {
		if m.Required {
			required = append(required, m.Name)
		}
	}
	return required
}

// ResolveDependencies takes a list of selected modules and returns
// the full list including all required dependencies, in correct order.
// NOTE: Only Model is a true dependency - other modules are optional
// and independent of each other.
func ResolveDependencies(selected []string) []string {
	// Build a set of selected modules
	selectedSet := make(map[string]bool)
	for _, name := range selected {
		selectedSet[name] = true
	}

	// Add Model dependency for any selected module
	for _, name := range selected {
		module := GetModule(name)
		if module != nil {
			for _, dep := range module.Dependencies {
				selectedSet[dep] = true
			}
		}
	}

	// Always include required modules (Model)
	for _, m := range ModuleRegistry {
		if m.Required {
			selectedSet[m.Name] = true
		}
	}

	// Return in registry order
	var result []string
	for _, m := range ModuleRegistry {
		if selectedSet[m.Name] {
			result = append(result, m.Name)
		}
	}

	return result
}

// ValidateModuleSelection checks if a selection is valid
// Returns an error message or empty string if valid
func ValidateModuleSelection(selected []string) string {
	if len(selected) == 0 {
		return "At least one module must be selected"
	}

	// Check for mutually exclusive modules: SQLDatastore and NoSQLDatastore cannot be selected together
	hasSQLDatastore := false
	hasNoSQLDatastore := false
	for _, name := range selected {
		if name == ModuleSQLDatastore {
			hasSQLDatastore = true
		}
		if name == ModuleNoSQLDatastore {
			hasNoSQLDatastore = true
		}
	}
	if hasSQLDatastore && hasNoSQLDatastore {
		return ModuleSQLDatastore + " and " + ModuleNoSQLDatastore + " cannot be selected together. Choose one datastore type."
	}

	// Note: Worker no longer requires a datastore module - it defaults to PostgreSQL if none selected

	// Check that all required modules are included after resolution
	resolved := ResolveDependencies(selected)
	for _, m := range ModuleRegistry {
		if m.Required {
			found := false
			for _, name := range resolved {
				if name == m.Name {
					found = true
					break
				}
			}
			if !found {
				return m.Name + " is required and must be selected"
			}
		}
	}

	return ""
}

// GetModuleDisplayOptions returns formatted strings for CLI display
// Internal modules are excluded from display
func GetModuleDisplayOptions() []string {
	var options []string
	for _, m := range ModuleRegistry {
		if m.Internal {
			continue // Skip internal modules
		}
		suffix := ""
		if m.Required {
			suffix = " (required)"
		}
		options = append(options, m.Name+" - "+m.Description+suffix)
	}
	return options
}

// GetSelectableModules returns names of modules that can be selected by users
func GetSelectableModules() []string {
	var modules []string
	for _, m := range ModuleRegistry {
		if !m.Internal {
			modules = append(modules, m.Name)
		}
	}
	return modules
}
