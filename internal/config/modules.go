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
	Name           string   // Module name (e.g., "Model")
	Description    string   // Short description for CLI display
	UseCase        string   // Business-level: when would you pick this module
	WhenToUse      string   // Natural language triggers (e.g., "user says 'I need a database'")
	DoesNotInclude string   // Explicit boundaries for this module
	Required       bool     // If true, cannot be deselected
	Internal       bool     // If true, not shown in CLI prompts (auto-included when needed)
	Dependencies   []string // Names of modules this depends on (only Model is a real dependency)
	ConflictsWith  []string // Explicit mutual exclusions
}

// ModuleRegistry contains all available modules
// NOTE: Only Model is a hard dependency for other modules.
// Other modules can work independently - if a module is selected without
// another, the generated code will have TODO placeholders instead of
// actual integrations.
var ModuleRegistry = []Module{
	{
		Name:           ModuleModel,
		Description:    "DTOs, Entities, Enums, Exceptions",
		UseCase:        "Core data structures shared across all modules. Always included — contains domain entities, DTOs, enums, and exception types using Immutables.",
		WhenToUse:      "Always included automatically. Contains the data model that all other modules depend on.",
		DoesNotInclude: "Does not include persistence logic, API endpoints, or business rules — only data definitions",
		Required:       true,
		Internal:       false,
		Dependencies:   []string{},
		ConflictsWith:  []string{},
	},
	{
		Name:           ModuleJobs,
		Description:    "Job request contracts for background processing",
		UseCase:        "Defines job request data contracts for background processing. Auto-included when Worker is selected.",
		WhenToUse:      "Automatically included when Worker module is selected. Not selected directly.",
		DoesNotInclude: "Does not include job execution logic — only request contracts",
		Required:       false,
		Internal:       true,
		Dependencies:   []string{ModuleModel},
		ConflictsWith:  []string{},
	},
	{
		Name:           ModuleSQLDatastore,
		Description:    "SQL repositories, Flyway migrations (PostgreSQL, MySQL)",
		UseCase:        "Adds relational database persistence with Spring Data JDBC. Choose when storing structured data with relationships (users, orders, products).",
		WhenToUse:      "User mentions: database, PostgreSQL, MySQL, SQL, relational, tables, migrations, persistence, CRUD",
		DoesNotInclude: "Does not include JPA/Hibernate, database schema design, seed data, or connection pooling tuning",
		Required:       false,
		Internal:       false,
		Dependencies:   []string{ModuleModel},
		ConflictsWith:  []string{ModuleNoSQLDatastore},
	},
	{
		Name:           ModuleNoSQLDatastore,
		Description:    "NoSQL repositories (MongoDB, Redis)",
		UseCase:        "Adds NoSQL database persistence. Choose MongoDB for flexible document storage or Redis for key-value/caching.",
		WhenToUse:      "User mentions: MongoDB, Redis, NoSQL, document store, cache, key-value",
		DoesNotInclude: "Does not include data modeling, indexing strategies, or cache eviction policies",
		Required:       false,
		Internal:       false,
		Dependencies:   []string{ModuleModel},
		ConflictsWith:  []string{ModuleSQLDatastore},
	},
	{
		Name:           ModuleShared,
		Description:    "Services, Circuit breaker, Utilities",
		UseCase:        "Business logic layer with circuit breaker support (Resilience4j). Houses service classes that orchestrate between datastores and external systems.",
		WhenToUse:      "User mentions: service layer, business logic, circuit breaker, shared utilities, resilience",
		DoesNotInclude: "Does not include custom business logic implementation — provides the structure and patterns only",
		Required:       false,
		Internal:       false,
		Dependencies:   []string{ModuleModel},
		ConflictsWith:  []string{},
	},
	{
		Name:           ModuleAPI,
		Description:    "REST endpoints, Validation",
		UseCase:        "Adds a Spring Boot REST API with OpenAPI documentation, CORS, correlation IDs, and health endpoints. Choose when building HTTP services.",
		WhenToUse:      "User mentions: REST, API, endpoints, HTTP, web service, backend, server, controller, Swagger",
		DoesNotInclude: "Does not include authentication, authorization, rate limiting, API versioning, pagination helpers, or GraphQL",
		Required:       false,
		Internal:       false,
		Dependencies:   []string{ModuleModel},
		ConflictsWith:  []string{},
	},
	{
		Name:           ModuleWorker,
		Description:    "Background jobs (fire-and-forget, scheduled, delayed, batch)",
		UseCase:        "Adds JobRunr background job processing with a dashboard. Choose when you need async tasks, scheduled jobs, or batch processing.",
		WhenToUse:      "User mentions: background jobs, worker, async tasks, scheduled, cron, batch, queue processing, fire-and-forget, delayed",
		DoesNotInclude: "Does not include job orchestration/workflow engines, saga pattern, or job monitoring beyond JobRunr dashboard",
		Required:       false,
		Internal:       false,
		Dependencies:   []string{ModuleModel, ModuleJobs},
		ConflictsWith:  []string{},
	},
	{
		Name:           ModuleEvents,
		Description:    "Event contracts for event-driven processing",
		UseCase:        "Defines event contracts using sealed interfaces and provides EventPublisher. Auto-included when EventConsumer is selected.",
		WhenToUse:      "Automatically included when EventConsumer module is selected. Not selected directly.",
		DoesNotInclude: "Does not include event processing logic — only contracts and publisher",
		Required:       false,
		Internal:       true,
		Dependencies:   []string{ModuleModel},
		ConflictsWith:  []string{},
	},
	{
		Name:           ModuleEventConsumer,
		Description:    "Event listeners (Kafka, RabbitMQ, SQS, Pub/Sub)",
		UseCase:        "Adds event-driven message processing. Choose when building async workflows, event sourcing, or decoupled microservices.",
		WhenToUse:      "User mentions: events, messages, queue, async processing, Kafka, RabbitMQ, SQS, Pub/Sub, event-driven, CQRS, streaming",
		DoesNotInclude: "Does not include event sourcing framework, saga orchestration, or schema registry",
		Required:       false,
		Internal:       false,
		Dependencies:   []string{ModuleModel, ModuleEvents},
		ConflictsWith:  []string{},
	},
	{
		Name:           ModuleMCP,
		Description:    "MCP server for AI tool integration (build, test, introspection)",
		UseCase:        "Adds an MCP (Model Context Protocol) server that exposes project build, test, and code review tools to AI coding assistants.",
		WhenToUse:      "User mentions: MCP, AI tools, AI integration, coding assistant tooling",
		DoesNotInclude: "Does not include custom MCP tools — provides build, test, format, and review tools only",
		Required:       false,
		Internal:       false,
		Dependencies:   []string{ModuleModel},
		ConflictsWith:  []string{},
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
