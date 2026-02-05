package config

import (
	"reflect"
	"testing"
)

func TestResolveDependencies(t *testing.T) {
	tests := []struct {
		name     string
		selected []string
		expected []string
	}{
		{
			name:     "Model only",
			selected: []string{"Model"},
			expected: []string{"Model"},
		},
		{
			name:     "SQLDatastore adds Model",
			selected: []string{"SQLDatastore"},
			expected: []string{"Model", "SQLDatastore"},
		},
		{
			name:     "Shared only adds Model (NOT SQLDatastore)",
			selected: []string{"Shared"},
			expected: []string{"Model", "Shared"},
		},
		{
			name:     "API only adds Model (NOT Shared or SQLDatastore)",
			selected: []string{"API"},
			expected: []string{"Model", "API"},
		},
		{
			name:     "Shared + SQLDatastore",
			selected: []string{"Shared", "SQLDatastore"},
			expected: []string{"Model", "SQLDatastore", "Shared"},
		},
		{
			name:     "All modules explicitly selected",
			selected: []string{"Model", "SQLDatastore", "Shared", "API"},
			expected: []string{"Model", "SQLDatastore", "Shared", "API"},
		},
		{
			name:     "Empty selection still includes required (Model)",
			selected: []string{},
			expected: []string{"Model"},
		},
		{
			name:     "API + SQLDatastore (no Shared)",
			selected: []string{"API", "SQLDatastore"},
			expected: []string{"Model", "SQLDatastore", "API"},
		},
		{
			name:     "Worker adds Jobs dependency",
			selected: []string{"Worker", "SQLDatastore"},
			expected: []string{"Model", "Jobs", "SQLDatastore", "Worker"},
		},
		{
			name:     "Worker with NoSQLDatastore",
			selected: []string{"Worker", "NoSQLDatastore"},
			expected: []string{"Model", "Jobs", "NoSQLDatastore", "Worker"},
		},
		{
			name:     "All modules with Worker",
			selected: []string{"Model", "SQLDatastore", "Shared", "API", "Worker"},
			expected: []string{"Model", "Jobs", "SQLDatastore", "Shared", "API", "Worker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveDependencies(tt.selected)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ResolveDependencies(%v) = %v, want %v", tt.selected, result, tt.expected)
			}
		})
	}
}

func TestGetModule(t *testing.T) {
	// Test existing module
	model := GetModule("Model")
	if model == nil {
		t.Fatal("GetModule('Model') returned nil")
	}
	if model.Name != "Model" {
		t.Errorf("Expected name 'Model', got '%s'", model.Name)
	}
	if !model.Required {
		t.Error("Model should be required")
	}

	// Test Shared module - should only depend on Model
	shared := GetModule("Shared")
	if shared == nil {
		t.Fatal("GetModule('Shared') returned nil")
	}
	if len(shared.Dependencies) != 1 || shared.Dependencies[0] != "Model" {
		t.Errorf("Shared should only depend on Model, got: %v", shared.Dependencies)
	}

	// Test API module - should only depend on Model
	api := GetModule("API")
	if api == nil {
		t.Fatal("GetModule('API') returned nil")
	}
	if len(api.Dependencies) != 1 || api.Dependencies[0] != "Model" {
		t.Errorf("API should only depend on Model, got: %v", api.Dependencies)
	}

	// Test non-existing module
	unknown := GetModule("Unknown")
	if unknown != nil {
		t.Error("GetModule('Unknown') should return nil")
	}
}

func TestValidateModuleSelection(t *testing.T) {
	tests := []struct {
		name      string
		selected  []string
		wantError bool
	}{
		{
			name:      "Valid: all modules",
			selected:  []string{"Model", "SQLDatastore", "Shared", "API"},
			wantError: false,
		},
		{
			name:      "Valid: Model only",
			selected:  []string{"Model"},
			wantError: false,
		},
		{
			name:      "Valid: API only (Model auto-added)",
			selected:  []string{"API"},
			wantError: false,
		},
		{
			name:      "Valid: Shared without SQLDatastore",
			selected:  []string{"Shared"},
			wantError: false,
		},
		{
			name:      "Invalid: empty selection",
			selected:  []string{},
			wantError: true,
		},
		{
			name:      "Valid: Worker with SQLDatastore",
			selected:  []string{"Worker", "SQLDatastore"},
			wantError: false,
		},
		{
			name:      "Valid: Worker with NoSQLDatastore",
			selected:  []string{"Worker", "NoSQLDatastore"},
			wantError: false,
		},
		{
			name:      "Valid: Worker without datastore (defaults to PostgreSQL)",
			selected:  []string{"Worker"},
			wantError: false,
		},
		{
			name:      "Valid: Worker with Model only (defaults to PostgreSQL)",
			selected:  []string{"Model", "Worker"},
			wantError: false,
		},
		{
			name:      "Invalid: SQLDatastore and NoSQLDatastore together",
			selected:  []string{"SQLDatastore", "NoSQLDatastore"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := ValidateModuleSelection(tt.selected)
			hasError := errMsg != ""
			if hasError != tt.wantError {
				t.Errorf("ValidateModuleSelection(%v) error = %v, wantError %v", tt.selected, errMsg, tt.wantError)
			}
		})
	}
}

func TestGetSelectableModules(t *testing.T) {
	modules := GetSelectableModules()

	// Jobs should NOT be in the list (it's internal)
	for _, m := range modules {
		if m == "Jobs" {
			t.Error("Jobs should not be in selectable modules (it's internal)")
		}
	}

	// Model, SQLDatastore, NoSQLDatastore, Shared, API, Worker should be in the list
	expected := []string{"Model", "SQLDatastore", "NoSQLDatastore", "Shared", "API", "Worker"}
	for _, exp := range expected {
		found := false
		for _, m := range modules {
			if m == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %s to be in selectable modules", exp)
		}
	}
}

func TestJobsModuleIsInternal(t *testing.T) {
	jobs := GetModule("Jobs")
	if jobs == nil {
		t.Fatal("Jobs module should exist")
	}
	if !jobs.Internal {
		t.Error("Jobs module should be marked as Internal")
	}
}

func TestWorkerDependsOnJobs(t *testing.T) {
	worker := GetModule("Worker")
	if worker == nil {
		t.Fatal("Worker module should exist")
	}

	hasJobsDep := false
	for _, dep := range worker.Dependencies {
		if dep == "Jobs" {
			hasJobsDep = true
			break
		}
	}
	if !hasJobsDep {
		t.Error("Worker module should have Jobs as a dependency")
	}
}

func TestProjectConfig_PackagePath(t *testing.T) {
	cfg := &ProjectConfig{GroupID: "com.company.project"}
	expected := "com/company/project"
	if got := cfg.PackagePath(); got != expected {
		t.Errorf("PackagePath() = %v, want %v", got, expected)
	}
}

func TestProjectConfig_ProjectNamePascal(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{"my-platform", "MyPlatform"},
		{"simple", "Simple"},
		{"my-cool-app", "MyCoolApp"},
		{"already_snake", "AlreadySnake"},
	}

	for _, tt := range tests {
		cfg := &ProjectConfig{ProjectName: tt.name}
		if got := cfg.ProjectNamePascal(); got != tt.expected {
			t.Errorf("ProjectNamePascal(%s) = %v, want %v", tt.name, got, tt.expected)
		}
	}
}

func TestProjectConfig_HasModule(t *testing.T) {
	cfg := &ProjectConfig{
		Modules: []string{"Model", "SQLDatastore", "API"},
	}

	if !cfg.HasModule("Model") {
		t.Error("HasModule('Model') should return true")
	}
	if !cfg.HasModule("API") {
		t.Error("HasModule('API') should return true")
	}
	if cfg.HasModule("Shared") {
		t.Error("HasModule('Shared') should return false")
	}
}

func TestProjectConfig_HasAllModules(t *testing.T) {
	cfg := &ProjectConfig{
		Modules: []string{"Model", "SQLDatastore", "Shared", "API"},
	}

	if !cfg.HasAllModules("Model", "API") {
		t.Error("HasAllModules('Model', 'API') should return true")
	}
	if !cfg.HasAllModules("Model", "SQLDatastore", "Shared") {
		t.Error("HasAllModules('Model', 'SQLDatastore', 'Shared') should return true")
	}

	cfg2 := &ProjectConfig{
		Modules: []string{"Model", "API"},
	}
	if cfg2.HasAllModules("Model", "SQLDatastore") {
		t.Error("HasAllModules('Model', 'SQLDatastore') should return false when SQLDatastore not selected")
	}
}

func TestProjectConfig_JobRunrStorageType(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ProjectConfig
		expected string
	}{
		{
			name: "SQLDatastore PostgreSQL",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "Worker"},
				Database: "postgresql",
			},
			expected: "sql",
		},
		{
			name: "SQLDatastore MySQL",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "Worker"},
				Database: "mysql",
			},
			expected: "sql",
		},
		{
			name: "NoSQLDatastore MongoDB",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "mongodb",
			},
			expected: "mongodb",
		},
		{
			name: "NoSQLDatastore Redis - fallback to SQL",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "redis",
			},
			expected: "sql",
		},
		{
			name: "Worker only - fallback to SQL",
			cfg: &ProjectConfig{
				Modules: []string{"Model", "Worker"},
			},
			expected: "sql",
		},
		{
			name: "No Worker - empty",
			cfg: &ProjectConfig{
				Modules: []string{"Model", "SQLDatastore"},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.JobRunrStorageType(); got != tt.expected {
				t.Errorf("JobRunrStorageType() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProjectConfig_JobRunrSqlDatabase(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ProjectConfig
		expected string
	}{
		{
			name: "SQLDatastore PostgreSQL",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "Worker"},
				Database: "postgresql",
			},
			expected: "postgresql",
		},
		{
			name: "SQLDatastore MySQL",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "Worker"},
				Database: "mysql",
			},
			expected: "mysql",
		},
		{
			name: "NoSQLDatastore Redis - defaults to postgresql",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "redis",
			},
			expected: "postgresql",
		},
		{
			name: "Worker only - defaults to postgresql",
			cfg: &ProjectConfig{
				Modules: []string{"Model", "Worker"},
			},
			expected: "postgresql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.JobRunrSqlDatabase(); got != tt.expected {
				t.Errorf("JobRunrSqlDatabase() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProjectConfig_WorkerUsesPostgresFallback(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ProjectConfig
		expected bool
	}{
		{
			name: "Worker with SQLDatastore - no fallback",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "Worker"},
				Database: "postgresql",
			},
			expected: false,
		},
		{
			name: "Worker with MongoDB - no fallback",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "mongodb",
			},
			expected: false,
		},
		{
			name: "Worker with Redis - fallback",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "redis",
			},
			expected: true,
		},
		{
			name: "No Worker - no fallback",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore"},
				NoSQLDatabase: "redis",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.WorkerUsesPostgresFallback(); got != tt.expected {
				t.Errorf("WorkerUsesPostgresFallback() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProjectConfig_WorkerNeedsOwnPostgres(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ProjectConfig
		expected bool
	}{
		{
			name: "Worker with SQLDatastore - no own postgres",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "Worker"},
				Database: "postgresql",
			},
			expected: false,
		},
		{
			name: "Worker with MongoDB - no own postgres",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "mongodb",
			},
			expected: false,
		},
		{
			name: "Worker with Redis - needs own postgres",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "redis",
			},
			expected: true,
		},
		{
			name: "No Worker with Redis - no own postgres",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore"},
				NoSQLDatabase: "redis",
			},
			expected: false,
		},
		{
			name: "Worker with no datastore - needs own postgres",
			cfg: &ProjectConfig{
				Modules: []string{"Model", "Jobs", "Worker"},
			},
			expected: true,
		},
		{
			name: "No Worker no datastore - no own postgres",
			cfg: &ProjectConfig{
				Modules: []string{"Model"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.WorkerNeedsOwnPostgres(); got != tt.expected {
				t.Errorf("WorkerNeedsOwnPostgres() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestProjectConfig_NeedsDockerCompose(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ProjectConfig
		expected bool
	}{
		{
			name: "API with SQLDatastore - needs docker",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "API"},
				Database: "postgresql",
			},
			expected: true,
		},
		{
			name: "API with NoSQLDatastore - needs docker",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "API"},
				NoSQLDatabase: "mongodb",
			},
			expected: true,
		},
		{
			name: "Worker with SQLDatastore no API - needs docker",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "Jobs", "Worker"},
				Database: "postgresql",
			},
			expected: true,
		},
		{
			name: "Worker with Redis - needs docker (postgres-jobrunr)",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Jobs", "Worker"},
				NoSQLDatabase: "redis",
			},
			expected: true,
		},
		{
			name: "Worker with no datastore - needs docker (postgres-jobrunr)",
			cfg: &ProjectConfig{
				Modules: []string{"Model", "Jobs", "Worker"},
			},
			expected: true,
		},
		{
			name: "Model only - no docker",
			cfg: &ProjectConfig{
				Modules: []string{"Model"},
			},
			expected: false,
		},
		{
			name: "SQLDatastore without runtime - no docker",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore"},
				Database: "postgresql",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.NeedsDockerCompose(); got != tt.expected {
				t.Errorf("NeedsDockerCompose() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// TestDisplayOptionsAlignWithSelectableModules verifies that the display options
// shown in the CLI multi-select are indexed consistently with GetSelectableModules().
// A mismatch here means users select one module but get a different one — the exact
// bug that caused Worker selection to produce Jobs instead.
func TestDisplayOptionsAlignWithSelectableModules(t *testing.T) {
	displayOptions := GetModuleDisplayOptions()
	selectableModules := GetSelectableModules()

	if len(displayOptions) != len(selectableModules) {
		t.Fatalf("Display options count (%d) != selectable modules count (%d)",
			len(displayOptions), len(selectableModules))
	}

	// Each display option must start with its corresponding selectable module name
	for i, moduleName := range selectableModules {
		displayOpt := displayOptions[i]
		expectedPrefix := moduleName + " - "
		if moduleName == "Model" {
			// Model has " (required)" suffix in description but still starts with "Model - "
		}
		if len(displayOpt) < len(expectedPrefix) || displayOpt[:len(expectedPrefix)] != expectedPrefix {
			t.Errorf("Display option[%d] = %q, expected to start with %q (module %q)",
				i, displayOpt, expectedPrefix, moduleName)
		}
	}

	// Verify internal modules are excluded from both
	for _, m := range ModuleRegistry {
		if !m.Internal {
			continue
		}
		for _, name := range selectableModules {
			if name == m.Name {
				t.Errorf("Internal module %q should not be in selectable modules", m.Name)
			}
		}
		for _, opt := range displayOptions {
			if len(opt) >= len(m.Name) && opt[:len(m.Name)] == m.Name {
				t.Errorf("Internal module %q should not appear in display options", m.Name)
			}
		}
	}
}

// TestModuleSelectionSimulation simulates the full CLI selection → resolution pipeline
// for every selectable module to ensure the selected module always ends up in the
// resolved list. This catches index-mapping bugs between display and module names.
func TestModuleSelectionSimulation(t *testing.T) {
	displayOptions := GetModuleDisplayOptions()
	selectableModules := GetSelectableModules()

	for displayIdx, moduleName := range selectableModules {
		t.Run(moduleName, func(t *testing.T) {
			// Simulate: user picks this display index
			if displayIdx >= len(displayOptions) {
				t.Fatalf("Display index %d out of range (only %d options)", displayIdx, len(displayOptions))
			}

			// Map display index to module name (this is what prompts.go does)
			selected := selectableModules[displayIdx]
			if selected != moduleName {
				t.Fatalf("Index %d mapped to %q, expected %q", displayIdx, selected, moduleName)
			}

			// Resolve dependencies
			resolved := ResolveDependencies([]string{selected})

			// The selected module must be in the resolved list
			found := false
			for _, m := range resolved {
				if m == moduleName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Module %q was selected but not found in resolved list: %v", moduleName, resolved)
			}

			// Model must always be present (it's required)
			modelFound := false
			for _, m := range resolved {
				if m == "Model" {
					modelFound = true
					break
				}
			}
			if !modelFound {
				t.Errorf("Model should always be in resolved list, got: %v", resolved)
			}

			// Worker must auto-include Jobs
			if moduleName == "Worker" {
				jobsFound := false
				for _, m := range resolved {
					if m == "Jobs" {
						jobsFound = true
						break
					}
				}
				if !jobsFound {
					t.Errorf("Worker should auto-include Jobs, resolved: %v", resolved)
				}
			}
		})
	}
}

func TestProjectConfig_ShowRedisWorkerWarning(t *testing.T) {
	tests := []struct {
		name     string
		cfg      *ProjectConfig
		expected bool
	}{
		{
			name: "Worker with Redis - show warning",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "redis",
			},
			expected: true,
		},
		{
			name: "Worker with MongoDB - no warning",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore", "Worker"},
				NoSQLDatabase: "mongodb",
			},
			expected: false,
		},
		{
			name: "Worker with SQLDatastore - no warning",
			cfg: &ProjectConfig{
				Modules:  []string{"Model", "SQLDatastore", "Worker"},
				Database: "postgresql",
			},
			expected: false,
		},
		{
			name: "No Worker with Redis - no warning",
			cfg: &ProjectConfig{
				Modules:       []string{"Model", "NoSQLDatastore"},
				NoSQLDatabase: "redis",
			},
			expected: false,
		},
		{
			name: "Worker only (no datastore) - no warning",
			cfg: &ProjectConfig{
				Modules: []string{"Model", "Worker"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cfg.ShowRedisWorkerWarning(); got != tt.expected {
				t.Errorf("ShowRedisWorkerWarning() = %v, want %v", got, tt.expected)
			}
		})
	}
}
