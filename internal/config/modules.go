package config

// Module represents a project module with its metadata
type Module struct {
	Name         string   // Module name (e.g., "Model")
	Description  string   // Short description for CLI display
	Required     bool     // If true, cannot be deselected
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
		Dependencies: []string{},
	},
	{
		Name:         "SQLDatastore",
		Description:  "Repositories, Flyway migrations",
		Required:     false,
		Dependencies: []string{"Model"}, // Only Model is required
	},
	{
		Name:         "Shared",
		Description:  "Services, Circuit breaker, Utilities",
		Required:     false,
		Dependencies: []string{"Model"}, // Only Model is required (NOT SQLDatastore)
	},
	{
		Name:         "API",
		Description:  "REST endpoints, Validation",
		Required:     false,
		Dependencies: []string{"Model"}, // Only Model is required (NOT Shared or SQLDatastore)
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
func GetModuleDisplayOptions() []string {
	options := make([]string, len(ModuleRegistry))
	for i, m := range ModuleRegistry {
		suffix := ""
		if m.Required {
			suffix = " (required)"
		}
		options[i] = m.Name + " - " + m.Description + suffix
	}
	return options
}

func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
