package prompts

import (
	"strings"
	"testing"
)

func TestGetAvailableModulesToAdd(t *testing.T) {
	tests := []struct {
		name            string
		existingModules []string
		expectedCount   int
		shouldContain   []string
		shouldNotContain []string
	}{
		{
			name:            "empty project has all modules available",
			existingModules: []string{},
			shouldContain:   []string{"SQLDatastore", "NoSQLDatastore", "Shared", "API", "Worker", "EventConsumer"},
			shouldNotContain: []string{"Model", "Jobs", "Events"}, // Model is required, Jobs/Events are internal
		},
		{
			name:            "project with Model only",
			existingModules: []string{"Model"},
			shouldContain:   []string{"SQLDatastore", "NoSQLDatastore", "Shared", "API", "Worker", "EventConsumer"},
			shouldNotContain: []string{"Model", "Jobs", "Events"},
		},
		{
			name:            "project with SQLDatastore excludes NoSQLDatastore",
			existingModules: []string{"Model", "SQLDatastore"},
			shouldContain:   []string{"Shared", "API", "Worker", "EventConsumer"},
			shouldNotContain: []string{"SQLDatastore", "NoSQLDatastore", "Model"},
		},
		{
			name:            "project with NoSQLDatastore excludes SQLDatastore",
			existingModules: []string{"Model", "NoSQLDatastore"},
			shouldContain:   []string{"Shared", "API", "Worker", "EventConsumer"},
			shouldNotContain: []string{"SQLDatastore", "NoSQLDatastore", "Model"},
		},
		{
			name:            "project with all modules",
			existingModules: []string{"Model", "SQLDatastore", "Shared", "API", "Worker", "EventConsumer", "Jobs", "Events"},
			expectedCount:   0,
			shouldContain:   []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			available := getAvailableModulesToAdd(tt.existingModules)

			if tt.expectedCount > 0 && len(available) != tt.expectedCount {
				t.Errorf("Expected %d modules, got %d: %v", tt.expectedCount, len(available), available)
			}

			for _, should := range tt.shouldContain {
				found := false
				for _, m := range available {
					if m == should {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected %s to be in available modules: %v", should, available)
				}
			}

			for _, shouldNot := range tt.shouldNotContain {
				for _, m := range available {
					if m == shouldNot {
						t.Errorf("Expected %s to NOT be in available modules: %v", shouldNot, available)
						break
					}
				}
			}
		})
	}
}

func TestBuildModuleOptions(t *testing.T) {
	t.Run("basic modules", func(t *testing.T) {
		available := []string{"SQLDatastore", "API"}
		existing := []string{"Model"}

		options := buildModuleOptions(available, existing)

		if len(options) != 2 {
			t.Errorf("Expected 2 options, got %d", len(options))
		}

		// Options should contain module name and description
		for _, opt := range options {
			if !strings.Contains(opt, " - ") {
				t.Errorf("Option should have description: %s", opt)
			}
		}
	})

	t.Run("Worker shows Jobs dependency", func(t *testing.T) {
		available := []string{"Worker"}
		existing := []string{"Model"}

		options := buildModuleOptions(available, existing)

		if len(options) != 1 {
			t.Fatalf("Expected 1 option, got %d", len(options))
		}

		if !strings.Contains(options[0], "will add Jobs") {
			t.Errorf("Worker option should mention Jobs dependency: %s", options[0])
		}
	})

	t.Run("Worker with Jobs existing doesn't show dependency", func(t *testing.T) {
		available := []string{"Worker"}
		existing := []string{"Model", "Jobs"}

		options := buildModuleOptions(available, existing)

		if len(options) != 1 {
			t.Fatalf("Expected 1 option, got %d", len(options))
		}

		if strings.Contains(options[0], "will add Jobs") {
			t.Errorf("Worker option should not mention Jobs when Jobs exists: %s", options[0])
		}
	})

	t.Run("EventConsumer shows Events dependency", func(t *testing.T) {
		available := []string{"EventConsumer"}
		existing := []string{"Model"}

		options := buildModuleOptions(available, existing)

		if len(options) != 1 {
			t.Fatalf("Expected 1 option, got %d", len(options))
		}

		if !strings.Contains(options[0], "will add Events") {
			t.Errorf("EventConsumer option should mention Events dependency: %s", options[0])
		}
	})
}

func TestContainsModule(t *testing.T) {
	modules := []string{"Model", "API", "SQLDatastore"}

	tests := []struct {
		name     string
		expected bool
	}{
		{"Model", true},
		{"API", true},
		{"SQLDatastore", true},
		{"Worker", false},
		{"NoSQLDatastore", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := containsModule(modules, tt.name); got != tt.expected {
				t.Errorf("containsModule(%s) = %v, want %v", tt.name, got, tt.expected)
			}
		})
	}
}

func TestGetModuleDependencies(t *testing.T) {
	tests := []struct {
		name            string
		module          string
		existingModules []string
		expectedDeps    []string
	}{
		{
			name:            "Worker adds Jobs",
			module:          "Worker",
			existingModules: []string{"Model"},
			expectedDeps:    []string{"Jobs"},
		},
		{
			name:            "Worker with Jobs already exists",
			module:          "Worker",
			existingModules: []string{"Model", "Jobs"},
			expectedDeps:    []string{},
		},
		{
			name:            "EventConsumer adds Events",
			module:          "EventConsumer",
			existingModules: []string{"Model"},
			expectedDeps:    []string{"Events"},
		},
		{
			name:            "EventConsumer with Events already exists",
			module:          "EventConsumer",
			existingModules: []string{"Model", "Events"},
			expectedDeps:    []string{},
		},
		{
			name:            "SQLDatastore has no additional deps",
			module:          "SQLDatastore",
			existingModules: []string{"Model"},
			expectedDeps:    []string{},
		},
		{
			name:            "API has no additional deps",
			module:          "API",
			existingModules: []string{"Model"},
			expectedDeps:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := GetModuleDependencies(tt.module, tt.existingModules)

			if len(deps) != len(tt.expectedDeps) {
				t.Errorf("Expected %d dependencies, got %d: %v", len(tt.expectedDeps), len(deps), deps)
			}

			for _, expected := range tt.expectedDeps {
				found := false
				for _, dep := range deps {
					if dep == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected dependency %s not found in %v", expected, deps)
				}
			}
		})
	}
}

func TestValidateModuleCanBeAdded(t *testing.T) {
	tests := []struct {
		name            string
		module          string
		existingModules []string
		wantError       bool
		errorContains   string
	}{
		{
			name:            "can add SQLDatastore",
			module:          "SQLDatastore",
			existingModules: []string{"Model"},
			wantError:       false,
		},
		{
			name:            "cannot add existing module",
			module:          "Model",
			existingModules: []string{"Model"},
			wantError:       true,
			errorContains:   "already exists",
		},
		{
			name:            "cannot add SQLDatastore when NoSQLDatastore exists",
			module:          "SQLDatastore",
			existingModules: []string{"Model", "NoSQLDatastore"},
			wantError:       true,
			errorContains:   "mutually exclusive",
		},
		{
			name:            "cannot add NoSQLDatastore when SQLDatastore exists",
			module:          "NoSQLDatastore",
			existingModules: []string{"Model", "SQLDatastore"},
			wantError:       true,
			errorContains:   "mutually exclusive",
		},
		{
			name:            "cannot add unknown module",
			module:          "UnknownModule",
			existingModules: []string{"Model"},
			wantError:       true,
			errorContains:   "unknown module",
		},
		{
			name:            "cannot add internal module Jobs directly",
			module:          "Jobs",
			existingModules: []string{"Model"},
			wantError:       true,
			errorContains:   "automatically included",
		},
		{
			name:            "cannot add internal module Events directly",
			module:          "Events",
			existingModules: []string{"Model"},
			wantError:       true,
			errorContains:   "automatically included",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateModuleCanBeAdded(tt.module, tt.existingModules)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error containing '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestGetParentModule(t *testing.T) {
	tests := []struct {
		internal string
		expected string
	}{
		{"Jobs", "Worker"},
		{"Events", "EventConsumer"},
		{"Model", ""},
		{"API", ""},
		{"Unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.internal, func(t *testing.T) {
			if got := getParentModule(tt.internal); got != tt.expected {
				t.Errorf("getParentModule(%s) = %s, want %s", tt.internal, got, tt.expected)
			}
		})
	}
}
