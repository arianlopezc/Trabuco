package prompts

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/trabuco/trabuco/internal/config"
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

	// Convert indices to module names
	moduleNames := config.GetModuleNames()
	selectedModules := make([]string, len(selectedIndices))
	for i, idx := range selectedIndices {
		selectedModules[i] = moduleNames[idx]
	}

	// Resolve dependencies (only adds Model if not selected)
	cfg.Modules = config.ResolveDependencies(selectedModules)

	// 4. Java version
	if err := survey.AskOne(&survey.Select{
		Message: "Java version:",
		Options: []string{"21 (Recommended - LTS until 2031)", "25 (Latest LTS)"},
		Default: "21 (Recommended - LTS until 2031)",
	}, &cfg.JavaVersion); err != nil {
		return nil, err
	}
	// Extract just the version number
	cfg.JavaVersion = strings.Split(cfg.JavaVersion, " ")[0]

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
	}

	// 7. CLAUDE.md (AI assistant context file)
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
	moduleNames := config.GetModuleNames()
	var selectedModules []string
	for _, idx := range selectedIndices {
		if idx < len(moduleNames) {
			selectedModules = append(selectedModules, moduleNames[idx])
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
