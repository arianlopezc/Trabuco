package config

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
		Name:         "Model",
		Description:  "DTOs, Entities, Enums, Exceptions",
		Required:     true,
		Internal:     false,
		Dependencies: []string{},
	},
	{
		Name:         "Jobs",
		Description:  "Job request contracts for background processing",
		Required:     false,
		Internal:     true, // Auto-included when Worker is selected
		Dependencies: []string{"Model"},
	},
	{
		Name:         "SQLDatastore",
		Description:  "SQL repositories, Flyway migrations (PostgreSQL, MySQL) - exclusive with NoSQLDatastore",
		Required:     false,
		Internal:     false,
		Dependencies: []string{"Model"},
	},
	{
		Name:         "NoSQLDatastore",
		Description:  "NoSQL repositories (MongoDB, Redis) - exclusive with SQLDatastore",
		Required:     false,
		Internal:     false,
		Dependencies: []string{"Model"},
	},
	{
		Name:         "Shared",
		Description:  "Services, Circuit breaker, Utilities",
		Required:     false,
		Internal:     false,
		Dependencies: []string{"Model"},
	},
	{
		Name:         "API",
		Description:  "REST endpoints, Validation",
		Required:     false,
		Internal:     false,
		Dependencies: []string{"Model"},
	},
	{
		Name:         "Worker",
		Description:  "Background jobs (fire-and-forget, scheduled, delayed, batch)",
		Required:     false,
		Internal:     false,
		Dependencies: []string{"Model", "Jobs"}, // Jobs is auto-included; uses datastore for JobRunr persistence (defaults to PostgreSQL if none)
	},
	{
		Name:         "Events",
		Description:  "Event contracts for event-driven processing",
		Required:     false,
		Internal:     true, // Auto-included when EventConsumer is selected
		Dependencies: []string{"Model"},
	},
	{
		Name:         "EventConsumer",
		Description:  "Event listeners (Kafka, RabbitMQ)",
		Required:     false,
		Internal:     false,
		Dependencies: []string{"Model", "Events"}, // Events is auto-included
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
		if name == "SQLDatastore" {
			hasSQLDatastore = true
		}
		if name == "NoSQLDatastore" {
			hasNoSQLDatastore = true
		}
	}
	if hasSQLDatastore && hasNoSQLDatastore {
		return "SQLDatastore and NoSQLDatastore cannot be selected together. Choose one datastore type."
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
