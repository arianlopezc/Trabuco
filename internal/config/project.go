package config

import "github.com/arianlopezc/Trabuco/internal/utils"

// ProjectConfig holds all configuration for a generated project
type ProjectConfig struct {
	// Basic info
	ProjectName string // e.g., "my-platform"
	GroupID     string // e.g., "com.company.project"
	ArtifactID  string // e.g., "my-platform" (usually same as ProjectName)

	// Java
	JavaVersion         string // "17", "21", or "25"
	JavaVersionDetected bool   // Whether the selected Java version was detected on the system

	// Modules
	Modules []string // e.g., ["Model", "SQLDatastore", "NoSQLDatastore", "Shared", "API"]

	// SQL Database (only if SQLDatastore selected)
	Database string // "postgresql", "mysql", or "generic"

	// NoSQL Database (only if NoSQLDatastore selected)
	NoSQLDatabase string // "mongodb" or "redis"

	// Documentation
	IncludeCLAUDEMD bool // Generate CLAUDE.md file
}

// Derived helper methods

// PackagePath returns the group ID as a file path (e.g., "com/company/project")
func (c *ProjectConfig) PackagePath() string {
	path := ""
	for _, ch := range c.GroupID {
		if ch == '.' {
			path += "/"
		} else {
			path += string(ch)
		}
	}
	return path
}

// ProjectNamePascal returns the project name in PascalCase (e.g., "MyPlatform")
func (c *ProjectConfig) ProjectNamePascal() string {
	return utils.ToPascalCase(c.ProjectName)
}

// ProjectNameCamel returns the project name in camelCase (e.g., "myPlatform")
func (c *ProjectConfig) ProjectNameCamel() string {
	return utils.ToCamelCase(c.ProjectName)
}

// ProjectNameSnake returns the project name in snake_case (e.g., "my_platform")
func (c *ProjectConfig) ProjectNameSnake() string {
	result := ""
	for _, ch := range c.ProjectName {
		if ch == '-' {
			result += "_"
		} else {
			result += string(ch)
		}
	}
	return result
}

// HasModule checks if a specific module is included
func (c *ProjectConfig) HasModule(name string) bool {
	for _, m := range c.Modules {
		if m == name {
			return true
		}
	}
	return false
}

// HasAllModules checks if all specified modules are included
func (c *ProjectConfig) HasAllModules(names ...string) bool {
	for _, name := range names {
		if !c.HasModule(name) {
			return false
		}
	}
	return true
}

// HasAnyDatastore checks if any datastore module is included
func (c *ProjectConfig) HasAnyDatastore() bool {
	return c.HasModule("SQLDatastore") || c.HasModule("NoSQLDatastore")
}

// HasBothDatastores checks if both datastore modules are included
func (c *ProjectConfig) HasBothDatastores() bool {
	return c.HasModule("SQLDatastore") && c.HasModule("NoSQLDatastore")
}
