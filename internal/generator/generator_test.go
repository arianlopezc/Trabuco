package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

func TestGenerator_Generate_ModelOnly(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "test-project",
		GroupID:     "com.test.project",
		ArtifactID:  "test-project",
		JavaVersion: "21",
		Modules:     []string{"Model"},
		Database:    "",
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{
		"pom.xml",
		".gitignore",
		"README.md",
		"Model/pom.xml",
		"Model/src/main/java/com/test/project/model/ImmutableStyle.java",
		"Model/src/main/java/com/test/project/model/entities/Placeholder.java",
		"Model/src/main/java/com/test/project/model/dto/PlaceholderRequest.java",
		"Model/src/main/java/com/test/project/model/dto/PlaceholderResponse.java",
	}

	for _, file := range expectedFiles {
		path := filepath.Join("test-project", file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", path)
		}
	}

	// Verify CLAUDE.md is NOT generated when IncludeCLAUDEMD is false
	claudePath := filepath.Join("test-project", "CLAUDE.md")
	if _, err := os.Stat(claudePath); !os.IsNotExist(err) {
		t.Error("CLAUDE.md should not exist when IncludeCLAUDEMD is false")
	}
}

func TestGenerator_Generate_AllModules(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.platform",
		ArtifactID:  "my-platform",
		JavaVersion: "21",
		Modules:     []string{"Model", "Jobs", "SQLDatastore", "Shared", "API", "Worker"},
		Database:    "postgresql",
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify expected files exist
	expectedFiles := []string{
		// Root
		"pom.xml",
		".gitignore",
		".dockerignore",
		"README.md",
		// Model
		"Model/pom.xml",
		"Model/src/main/java/com/company/platform/model/ImmutableStyle.java",
		"Model/src/main/java/com/company/platform/model/entities/Placeholder.java",
		"Model/src/main/java/com/company/platform/model/dto/PlaceholderRequest.java",
		"Model/src/main/java/com/company/platform/model/dto/PlaceholderResponse.java",
		// Jobs module (job service for enqueueing)
		"Jobs/pom.xml",
		"Jobs/src/main/java/com/company/platform/jobs/PlaceholderJobService.java",
		// Job request schemas in Model module
		"Model/src/main/java/com/company/platform/model/jobs/PlaceholderJobRequest.java",
		"Model/src/main/java/com/company/platform/model/jobs/ProcessPlaceholderJobRequest.java",
		"Model/src/main/java/com/company/platform/model/jobs/ProcessPlaceholderJobRequestHandler.java",
		// SQLDatastore
		"SQLDatastore/pom.xml",
		"SQLDatastore/src/main/java/com/company/platform/sqldatastore/config/DatabaseConfig.java",
		"SQLDatastore/src/main/java/com/company/platform/sqldatastore/repository/PlaceholderRepository.java",
		"SQLDatastore/src/main/resources/db/migration/V1__baseline.sql",
		"SQLDatastore/src/test/java/com/company/platform/sqldatastore/repository/PlaceholderRepositoryTest.java",
		// Shared
		"Shared/pom.xml",
		"Shared/src/main/java/com/company/platform/shared/config/SharedConfig.java",
		"Shared/src/main/java/com/company/platform/shared/config/CircuitBreakerConfiguration.java",
		"Shared/src/main/java/com/company/platform/shared/service/PlaceholderService.java",
		"Shared/src/main/resources/application.yml",
		"Shared/src/test/java/com/company/platform/shared/service/PlaceholderServiceTest.java",
		// API
		"API/pom.xml",
		"API/Dockerfile",
		"API/src/main/java/com/company/platform/api/MyPlatformApiApplication.java",
		"API/src/main/java/com/company/platform/api/controller/HealthController.java",
		"API/src/main/java/com/company/platform/api/controller/PlaceholderController.java",
		"API/src/main/java/com/company/platform/api/controller/PlaceholderJobController.java",
		"API/src/main/java/com/company/platform/api/config/WebConfig.java",
		"API/src/main/java/com/company/platform/api/config/GlobalExceptionHandler.java",
		"API/src/main/java/com/company/platform/api/config/SecurityHeadersFilter.java",
		"API/src/main/resources/application.yml",
		"API/src/main/resources/logback-spring.xml",
		// Worker
		"Worker/pom.xml",
		"Worker/Dockerfile",
		"Worker/src/main/java/com/company/platform/worker/MyPlatformWorkerApplication.java",
		"Worker/src/main/java/com/company/platform/worker/config/JobRunrConfig.java",
		"Worker/src/main/java/com/company/platform/worker/config/RecurringJobsConfig.java",
		"Worker/src/main/java/com/company/platform/worker/handler/ProcessPlaceholderJobRequestHandler.java",
		"Worker/src/main/resources/application.yml",
		"Worker/src/main/resources/logback-spring.xml",
		"Worker/src/test/java/com/company/platform/worker/handler/ProcessPlaceholderJobRequestHandlerTest.java",
		// Run configurations
		".run/API.run.xml",
		".run/Worker.run.xml",
		// Docker
		"docker-compose.yml",
		".env.example",
	}

	for _, file := range expectedFiles {
		path := filepath.Join("my-platform", file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", path)
		}
	}
}

// TestGenerator_Generate_SelectedModulesMatchGenerated verifies that every module
// in the config produces its corresponding directory, and no extra module directories
// appear. This catches bugs where module selection maps to wrong modules.
func TestGenerator_Generate_SelectedModulesMatchGenerated(t *testing.T) {
	tests := []struct {
		name            string
		modules         []string
		database        string
		noSQLDatabase   string
		expectedDirs    []string
		notExpectedDirs []string
	}{
		{
			name:            "Worker selected produces Worker dir, not just Jobs",
			modules:         []string{"Model", "Jobs", "Worker"},
			expectedDirs:    []string{"Model", "Jobs", "Worker"},
			notExpectedDirs: []string{"API", "SQLDatastore", "NoSQLDatastore", "Shared"},
		},
		{
			name:            "API selected does not produce Worker",
			modules:         []string{"Model", "API"},
			expectedDirs:    []string{"Model", "API"},
			notExpectedDirs: []string{"Worker", "Jobs", "SQLDatastore", "NoSQLDatastore", "Shared"},
		},
		{
			name:            "NoSQLDatastore with MongoDB",
			modules:         []string{"Model", "NoSQLDatastore"},
			noSQLDatabase:   "mongodb",
			expectedDirs:    []string{"Model", "NoSQLDatastore"},
			notExpectedDirs: []string{"SQLDatastore", "Worker", "Jobs", "API", "Shared"},
		},
		{
			name:            "Full stack with Worker",
			modules:         []string{"Model", "Jobs", "SQLDatastore", "Shared", "API", "Worker"},
			database:        "postgresql",
			expectedDirs:    []string{"Model", "Jobs", "SQLDatastore", "Shared", "API", "Worker"},
			notExpectedDirs: []string{"NoSQLDatastore"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "trabuco-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			cfg := &config.ProjectConfig{
				ProjectName:   "test-project",
				GroupID:       "com.test.project",
				ArtifactID:    "test-project",
				JavaVersion:   "21",
				Modules:       tt.modules,
				Database:      tt.database,
				NoSQLDatabase: tt.noSQLDatabase,
			}

			gen, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create generator: %v", err)
			}

			if err := gen.Generate(); err != nil {
				t.Fatalf("Failed to generate project: %v", err)
			}

			for _, dir := range tt.expectedDirs {
				path := filepath.Join("test-project", dir)
				info, err := os.Stat(path)
				if os.IsNotExist(err) {
					t.Errorf("Expected directory %s to exist", dir)
				} else if !info.IsDir() {
					t.Errorf("Expected %s to be a directory", dir)
				}
			}

			for _, dir := range tt.notExpectedDirs {
				path := filepath.Join("test-project", dir)
				if _, err := os.Stat(path); !os.IsNotExist(err) {
					t.Errorf("Directory %s should NOT exist for this module selection", dir)
				}
			}
		})
	}
}

func TestGenerator_Generate_DirectoryExists(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	// Create a directory that would conflict
	os.MkdirAll("existing-project", 0755)

	cfg := &config.ProjectConfig{
		ProjectName: "existing-project",
		GroupID:     "com.test",
		ArtifactID:  "existing-project",
		JavaVersion: "21",
		Modules:     []string{"Model"},
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	err = gen.Generate()
	if err == nil {
		t.Error("Expected error when directory already exists")
	}
}

func TestGenerator_Generate_APIWithoutWorker_NoJobController(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "api-no-worker",
		GroupID:     "com.test.apinoworker",
		ArtifactID:  "api-no-worker",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// PlaceholderJobController should NOT exist when Worker is not selected
	jobControllerPath := filepath.Join("api-no-worker", "API", "src", "main", "java",
		"com", "test", "apinoworker", "api", "controller", "PlaceholderJobController.java")
	if _, err := os.Stat(jobControllerPath); !os.IsNotExist(err) {
		t.Error("PlaceholderJobController.java should NOT exist when Worker module is not selected")
	}

	// PlaceholderController should still exist
	controllerPath := filepath.Join("api-no-worker", "API", "src", "main", "java",
		"com", "test", "apinoworker", "api", "controller", "PlaceholderController.java")
	if _, err := os.Stat(controllerPath); os.IsNotExist(err) {
		t.Error("PlaceholderController.java should exist")
	}
}

func TestGenerator_Generate_WithCLAUDEMD(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Change to temp directory
	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName:    "claude-test",
		GroupID:        "com.test",
		ArtifactID:     "claude-test",
		JavaVersion:    "21",
		Modules:        []string{"Model", "SQLDatastore", "API"},
		Database:       "postgresql",
		IncludeCLAUDEMD: true,
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify CLAUDE.md exists
	claudePath := filepath.Join("claude-test", "CLAUDE.md")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Error("CLAUDE.md should exist when IncludeCLAUDEMD is true")
	}

	// Verify CLAUDE.md contains module-specific content
	content, err := os.ReadFile(claudePath)
	if err != nil {
		t.Fatalf("Failed to read CLAUDE.md: %v", err)
	}

	contentStr := string(content)
	// Should contain SQLDatastore-specific content
	if !contains(contentStr, "Flyway") {
		t.Error("CLAUDE.md should contain Flyway reference when SQLDatastore is selected")
	}
	// Should contain API-specific content
	if !contains(contentStr, "GlobalExceptionHandler") {
		t.Error("CLAUDE.md should contain GlobalExceptionHandler when API is selected")
	}
	// CLAUDE.md should reference .claude/rules/ (not .ai/prompts/) for Claude Code
	if !contains(contentStr, ".claude/rules/") {
		t.Error("CLAUDE.md should reference .claude/rules/ for Claude Code agent")
	}
	if contains(contentStr, ".ai/prompts/") {
		t.Error("CLAUDE.md should NOT reference .ai/prompts/ for Claude Code agent")
	}

	// Verify .claude/rules/ files are generated
	rulesDir := filepath.Join("claude-test", ".claude", "rules")
	expectedRules := []string{
		"JAVA_CODE_QUALITY.md",
		"code-review.md",
		"testing-guide.md",
		"extending-the-project.md",
		"add-entity.md",
		"add-endpoint.md",
	}
	for _, rule := range expectedRules {
		rulePath := filepath.Join(rulesDir, rule)
		if _, err := os.Stat(rulePath); os.IsNotExist(err) {
			t.Errorf(".claude/rules/%s should exist when Claude Code is selected", rule)
		}
	}

	// Verify .claude/settings.json exists
	settingsPath := filepath.Join("claude-test", ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); os.IsNotExist(err) {
		t.Error(".claude/settings.json should exist when Claude Code is selected")
	}
}

func TestGenerator_Generate_DatabaseSpecificContent(t *testing.T) {
	tests := []struct {
		name     string
		database string
		contains string
	}{
		{"PostgreSQL", "postgresql", "org.postgresql.Driver"},
		{"MySQL", "mysql", "com.mysql.cj.jdbc.Driver"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "trabuco-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tempDir)

			oldWd, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(oldWd)

			cfg := &config.ProjectConfig{
				ProjectName: "db-test",
				GroupID:     "com.test",
				ArtifactID:  "db-test",
				JavaVersion: "21",
				Modules:     []string{"Model", "SQLDatastore", "API"},
				Database:    tt.database,
			}

			gen, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create generator: %v", err)
			}

			if err := gen.Generate(); err != nil {
				t.Fatalf("Failed to generate project: %v", err)
			}

			// Read application.yml to verify database-specific content
			appYaml, err := os.ReadFile(filepath.Join("db-test", "API", "src", "main", "resources", "application.yml"))
			if err != nil {
				t.Fatalf("Failed to read application.yml: %v", err)
			}

			if !contains(string(appYaml), tt.contains) {
				t.Errorf("application.yml should contain %s for %s database", tt.contains, tt.database)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
