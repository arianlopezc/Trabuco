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
