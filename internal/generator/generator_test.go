package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/trabuco/trabuco/internal/config"
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
		Modules:     []string{"Model", "SQLDatastore", "Shared", "API"},
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
		"README.md",
		// Model
		"Model/pom.xml",
		"Model/src/main/java/com/company/platform/model/ImmutableStyle.java",
		"Model/src/main/java/com/company/platform/model/entities/Placeholder.java",
		"Model/src/main/java/com/company/platform/model/dto/PlaceholderRequest.java",
		"Model/src/main/java/com/company/platform/model/dto/PlaceholderResponse.java",
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
		"API/src/main/java/com/company/platform/api/MyPlatformApiApplication.java",
		"API/src/main/java/com/company/platform/api/controller/HealthController.java",
		"API/src/main/java/com/company/platform/api/controller/PlaceholderController.java",
		"API/src/main/java/com/company/platform/api/config/WebConfig.java",
		"API/src/main/java/com/company/platform/api/config/GlobalExceptionHandler.java",
		"API/src/main/java/com/company/platform/api/config/SecurityHeadersFilter.java",
		"API/src/main/resources/application.yml",
		// Run configuration
		".run/API.run.xml",
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
	if !contains(contentStr, "flyway") {
		t.Error("CLAUDE.md should contain Flyway reference when SQLDatastore is selected")
	}
	// Should contain API-specific content
	if !contains(contentStr, "spring-boot:run") {
		t.Error("CLAUDE.md should contain spring-boot:run when API is selected")
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
