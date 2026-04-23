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
	// CLAUDE.md should reference .claude/rules/ for always-loaded rules (quality, review, testing)
	if !contains(contentStr, ".claude/rules/") {
		t.Error("CLAUDE.md should reference .claude/rules/ for always-loaded rules")
	}
	// CLAUDE.md should ALSO reference .ai/prompts/ for on-demand task guides (add-entity, add-tool, etc.)
	if !contains(contentStr, ".ai/prompts/") {
		t.Error("CLAUDE.md should reference .ai/prompts/ for on-demand task guides")
	}

	// Verify .claude/rules/ files are generated (only path-scoped rules, not task playbooks)
	rulesDir := filepath.Join("claude-test", ".claude", "rules")
	expectedRules := []string{
		"JAVA_CODE_QUALITY.md",
		"code-review.md",
		"testing-guide.md",
	}
	for _, rule := range expectedRules {
		rulePath := filepath.Join(rulesDir, rule)
		if _, err := os.Stat(rulePath); os.IsNotExist(err) {
			t.Errorf(".claude/rules/%s should exist when Claude Code is selected", rule)
		}
	}

	// Verify task playbooks are NOT in .claude/rules/ (they belong only in .ai/prompts/)
	unexpectedRules := []string{
		"extending-the-project.md",
		"add-entity.md",
		"add-endpoint.md",
		"add-job.md",
		"add-event.md",
	}
	for _, rule := range unexpectedRules {
		rulePath := filepath.Join(rulesDir, rule)
		if _, err := os.Stat(rulePath); err == nil {
			t.Errorf(".claude/rules/%s should NOT exist (task playbooks belong in .ai/prompts/ only)", rule)
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

func TestGenerator_Generate_GlobalExceptionHandler_DataIntegrity(t *testing.T) {
	tests := []struct {
		name             string
		modules          []string
		database         string
		noSQLDatabase    string
		shouldContain    bool
	}{
		{
			name:          "SQLDatastore includes DataIntegrity handlers",
			modules:       []string{"Model", "SQLDatastore", "API"},
			database:      "postgresql",
			shouldContain: true,
		},
		{
			name:          "NoSQLDatastore includes DataIntegrity handlers",
			modules:       []string{"Model", "NoSQLDatastore", "API"},
			noSQLDatabase: "mongodb",
			shouldContain: true,
		},
		{
			name:          "API-only omits DataIntegrity handlers",
			modules:       []string{"Model", "API"},
			shouldContain: false,
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
				ProjectName:   "integrity-test",
				GroupID:       "com.test",
				ArtifactID:    "integrity-test",
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

			handlerPath := filepath.Join("integrity-test", "API", "src", "main", "java",
				"com", "test", "api", "config", "GlobalExceptionHandler.java")
			data, err := os.ReadFile(handlerPath)
			if err != nil {
				t.Fatalf("Failed to read GlobalExceptionHandler.java: %v", err)
			}
			src := string(data)

			markers := []string{
				"DataIntegrityViolationException",
				"DuplicateKeyException",
				"HttpStatus.CONFLICT",
			}
			for _, m := range markers {
				present := contains(src, m)
				if tt.shouldContain && !present {
					t.Errorf("expected GlobalExceptionHandler to contain %q", m)
				}
				if !tt.shouldContain && present {
					t.Errorf("expected GlobalExceptionHandler to NOT contain %q", m)
				}
			}
		})
	}
}

func TestGenerator_Generate_DatastorePerformancePatterns(t *testing.T) {
	tests := []struct {
		name          string
		modules       []string
		database      string
		noSQLDatabase string
		checks        map[string][]string // relative path -> required substrings
	}{
		{
			name:     "SQL repo exposes batch + drain + bounded-update methods",
			modules:  []string{"Model", "SQLDatastore", "Shared"},
			database: "postgresql",
			checks: map[string][]string{
				"perf-test/SQLDatastore/src/main/java/com/test/sqldatastore/repository/PlaceholderRepository.java": {
					"findAllByIdIn",
					"findPage",
					"updateDescriptionBatchWithLimit",
					"FOR UPDATE SKIP LOCKED",
				},
				"perf-test/Shared/src/main/java/com/test/shared/service/PlaceholderService.java": {
					"DEFAULT_IN_CHUNK_SIZE",
					"processAllBatched",
					"findByIds",
					"chunked",
				},
			},
		},
		{
			name:     "MySQL uses ORDER BY LIMIT form (no FOR UPDATE SKIP LOCKED)",
			modules:  []string{"Model", "SQLDatastore", "Shared"},
			database: "mysql",
			checks: map[string][]string{
				"perf-test/SQLDatastore/src/main/java/com/test/sqldatastore/repository/PlaceholderRepository.java": {
					"updateDescriptionBatchWithLimit",
					"ORDER BY id",
					"LIMIT :limit",
				},
			},
		},
		{
			name:          "Mongo repo exposes findAllByIdIn + keyset method",
			modules:       []string{"Model", "NoSQLDatastore", "Shared"},
			noSQLDatabase: "mongodb",
			checks: map[string][]string{
				"perf-test/NoSQLDatastore/src/main/java/com/test/nosqldatastore/repository/PlaceholderDocumentRepository.java": {
					"findAllByIdIn",
					"findByIdGreaterThanOrderByIdAsc",
					"Limit limit",
				},
				"perf-test/Shared/src/main/java/com/test/shared/service/PlaceholderService.java": {
					"processAllBatched",
					"findByIds",
					"chunked",
				},
			},
		},
		{
			name:     "JAVA_CODE_QUALITY.md ships §5.5 when a datastore is selected",
			modules:  []string{"Model", "SQLDatastore", "API"},
			database: "postgresql",
			checks: map[string][]string{
				"perf-test/.ai/prompts/JAVA_CODE_QUALITY.md": {
					"### 5.5 Datastore Performance",
					"Keyset Drain Loop",
					"chunked",
					"ESR",
				},
				"perf-test/AGENTS.md": {
					"Keyset Drain Loop",
				},
				"perf-test/CLAUDE.md": {
					"findAllByIdIn",
				},
			},
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
				ProjectName:     "perf-test",
				GroupID:         "com.test",
				ArtifactID:      "perf-test",
				JavaVersion:     "21",
				Modules:         tt.modules,
				Database:        tt.database,
				NoSQLDatabase:   tt.noSQLDatabase,
				IncludeCLAUDEMD: true,
			}

			gen, err := New(cfg)
			if err != nil {
				t.Fatalf("Failed to create generator: %v", err)
			}
			if err := gen.Generate(); err != nil {
				t.Fatalf("Failed to generate project: %v", err)
			}

			for path, needles := range tt.checks {
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read %s: %v", path, err)
				}
				src := string(data)
				for _, needle := range needles {
					if !contains(src, needle) {
						t.Errorf("%s: expected substring %q not found", path, needle)
					}
				}
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

func TestGenerator_Generate_AIAgentModule(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "ai-test",
		GroupID:     "com.test.aitest",
		ArtifactID:  "ai-test",
		JavaVersion: "21",
		Modules:     []string{"Model", "Shared", "AIAgent"},
		Database:    "",
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	expectedFiles := []string{
		// Root
		"pom.xml",
		// AIAgent POM
		"AIAgent/pom.xml",
		"AIAgent/Dockerfile",
		// Application class
		"AIAgent/src/main/java/com/test/aitest/aiagent/AiTestAIAgentApplication.java",
		// Config
		"AIAgent/src/main/java/com/test/aitest/aiagent/config/ChatClientConfig.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/config/McpServerConfig.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/config/WebConfig.java",
		// Security
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/CallerIdentity.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/CallerContext.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/ApiKeyAuthFilter.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/ScopeEnforcer.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/RequireScope.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/ScopeInterceptor.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/RateLimiter.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/InputGuardrailAdvisor.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/OutputGuardrailAdvisor.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/security/CorrelationIdFilter.java",
		// Tools
		"AIAgent/src/main/java/com/test/aitest/aiagent/tool/PlaceholderTools.java",
		// Agents
		"AIAgent/src/main/java/com/test/aitest/aiagent/agent/PrimaryAgent.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/agent/SpecialistAgent.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/agent/SpecialistAgentTool.java",
		// Brain
		"AIAgent/src/main/java/com/test/aitest/aiagent/brain/MemoryEntry.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/brain/Scratchpad.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/brain/ReflectionDecision.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/brain/ReflectionService.java",
		// Knowledge
		"AIAgent/src/main/java/com/test/aitest/aiagent/knowledge/KnowledgeBase.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/knowledge/KnowledgeTools.java",
		// Protocol
		"AIAgent/src/main/java/com/test/aitest/aiagent/protocol/AgentRestController.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/protocol/A2AController.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/protocol/DiscoveryController.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/protocol/StreamingController.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/protocol/WebhookController.java",
		// Task
		"AIAgent/src/main/java/com/test/aitest/aiagent/task/TaskRecord.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/task/TaskEvent.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/task/TaskManager.java",
		// Event
		"AIAgent/src/main/java/com/test/aitest/aiagent/event/WebhookRegistration.java",
		"AIAgent/src/main/java/com/test/aitest/aiagent/event/WebhookManager.java",
		// Model DTOs (generated into Model module)
		"Model/src/main/java/com/test/aitest/model/dto/JsonRpcRequest.java",
		"Model/src/main/java/com/test/aitest/model/dto/JsonRpcResponse.java",
		"Model/src/main/java/com/test/aitest/model/dto/ChatRequest.java",
		"Model/src/main/java/com/test/aitest/model/dto/ChatResponse.java",
		"Model/src/main/java/com/test/aitest/model/dto/AskRequest.java",
		"Model/src/main/java/com/test/aitest/model/dto/AskResponse.java",
		"Model/src/main/java/com/test/aitest/model/dto/WebhookRegisterRequest.java",
		// Resources
		"AIAgent/src/main/resources/application.yml",
		"AIAgent/src/main/resources/logback-spring.xml",
		"AIAgent/src/main/resources/.well-known/agent.json",
		// Tests
		"AIAgent/src/test/java/com/test/aitest/aiagent/security/CallerIdentityTest.java",
		"AIAgent/src/test/java/com/test/aitest/aiagent/security/ScopeEnforcerTest.java",
		"AIAgent/src/test/java/com/test/aitest/aiagent/security/RateLimiterTest.java",
		"AIAgent/src/test/java/com/test/aitest/aiagent/security/OutputGuardrailTest.java",
		"AIAgent/src/test/java/com/test/aitest/aiagent/security/CorrelationIdFilterTest.java",
		"AIAgent/src/test/java/com/test/aitest/aiagent/brain/ScratchpadTest.java",
		"AIAgent/src/test/java/com/test/aitest/aiagent/brain/ReflectionDecisionTest.java",
		"AIAgent/src/test/java/com/test/aitest/aiagent/tool/PlaceholderToolsTest.java",
		"AIAgent/src/test/java/com/test/aitest/aiagent/task/TaskManagerTest.java",
		// IntelliJ run config
		".run/AIAgent.run.xml",
	}

	for _, file := range expectedFiles {
		path := filepath.Join("ai-test", file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file %s to exist", path)
		}
	}
}

func TestGenerator_Generate_AIAgentOnly_NoShared(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "ai-minimal",
		GroupID:     "com.test.minimal",
		ArtifactID:  "ai-minimal",
		JavaVersion: "21",
		Modules:     []string{"Model", "AIAgent"},
		Database:    "",
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify AIAgent exists but Shared does not
	aiPath := filepath.Join("ai-minimal", "AIAgent")
	if _, err := os.Stat(aiPath); os.IsNotExist(err) {
		t.Error("Expected AIAgent directory to exist")
	}
	sharedPath := filepath.Join("ai-minimal", "Shared")
	if _, err := os.Stat(sharedPath); !os.IsNotExist(err) {
		t.Error("Shared directory should NOT exist when not selected")
	}

	// Verify ComponentScan does NOT include shared package
	appFile := filepath.Join("ai-minimal", "AIAgent", "src", "main", "java", "com", "test", "minimal", "aiagent", "AiMinimalAIAgentApplication.java")
	content, err := os.ReadFile(appFile)
	if err != nil {
		t.Fatalf("Failed to read application file: %v", err)
	}
	if contains(string(content), "com.test.minimal.shared") {
		t.Error("Application class should NOT reference shared package when Shared is not selected")
	}
}

func TestGenerator_Generate_AIAgentModuleSelection(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "ai-select",
		GroupID:     "com.test.aiselect",
		ArtifactID:  "ai-select",
		JavaVersion: "21",
		Modules:     []string{"Model", "AIAgent"},
		Database:    "",
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	expectedDirs := []string{"Model", "AIAgent"}
	notExpectedDirs := []string{"API", "Worker", "Jobs", "SQLDatastore", "NoSQLDatastore", "Shared", "EventConsumer", "Events"}

	for _, dir := range expectedDirs {
		path := filepath.Join("ai-select", dir)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			t.Errorf("Expected directory %s to exist", dir)
		} else if !info.IsDir() {
			t.Errorf("Expected %s to be a directory", dir)
		}
	}

	for _, dir := range notExpectedDirs {
		path := filepath.Join("ai-select", dir)
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("Directory %s should NOT exist when only Model+AIAgent selected", dir)
		}
	}
}

func TestGenerator_Generate_AIAgentParentPOM_SpringAI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "ai-pom",
		GroupID:     "com.test.aipom",
		ArtifactID:  "ai-pom",
		JavaVersion: "21",
		Modules:     []string{"Model", "AIAgent"},
		Database:    "",
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify parent POM includes Spring AI BOM
	parentPom, err := os.ReadFile(filepath.Join("ai-pom", "pom.xml"))
	if err != nil {
		t.Fatalf("Failed to read parent pom.xml: %v", err)
	}
	pomContent := string(parentPom)
	if !contains(pomContent, "spring-ai.version") {
		t.Error("Parent POM should include spring-ai.version property when AIAgent is selected")
	}
	if !contains(pomContent, "spring-ai-bom") {
		t.Error("Parent POM should include spring-ai-bom dependency when AIAgent is selected")
	}
}

func TestGenerator_Generate_NoAIAgent_NoSpringAI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "no-ai",
		GroupID:     "com.test.noai",
		ArtifactID:  "no-ai",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		Database:    "",
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	// Verify parent POM does NOT include Spring AI when AIAgent is not selected
	parentPom, err := os.ReadFile(filepath.Join("no-ai", "pom.xml"))
	if err != nil {
		t.Fatalf("Failed to read parent pom.xml: %v", err)
	}
	pomContent := string(parentPom)
	if contains(pomContent, "spring-ai-bom") {
		t.Error("Parent POM should NOT include spring-ai-bom when AIAgent is not selected")
	}
}

// Stop-hook adapters invoke .github/scripts/review-checks.sh for Layer 2.
// The script used to be emitted only when --ci=github was selected, which
// meant hooks silently degraded to Layer 1 in projects without CI opt-in —
// a real bug we caught during e2e validation. This test locks in the fix:
// review-checks.sh must be emitted whenever review is enabled, regardless
// of CI provider choice.
func TestGenerator_Generate_ReviewScriptEmittedWithoutCI(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	cfg := &config.ProjectConfig{
		ProjectName: "review-no-ci",
		GroupID:     "com.test.reviewnoci",
		ArtifactID:  "review-no-ci",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		AIAgents:    []string{"claude"},
		Review:      config.ReviewConfig{Mode: config.ReviewModeFull},
		// NOTE: CIProvider intentionally empty — user did not opt into GitHub CI.
	}

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	scriptPath := filepath.Join("review-no-ci", ".github", "scripts", "review-checks.sh")
	info, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		t.Fatal("review-checks.sh should be emitted when review is enabled, regardless of CI provider")
	}
	if err != nil {
		t.Fatalf("stat review-checks.sh: %v", err)
	}
	// Must be executable — Stop hooks invoke it with bash, but the adapters
	// themselves check for the -x bit before running.
	if info.Mode()&0o111 == 0 {
		t.Errorf("review-checks.sh should be executable, got mode %v", info.Mode())
	}

	// And conversely: ci.yml must NOT be emitted, since user didn't opt in.
	ciPath := filepath.Join("review-no-ci", ".github", "workflows", "ci.yml")
	if _, err := os.Stat(ciPath); err == nil {
		t.Errorf("ci.yml should NOT be emitted without CIProvider='github'")
	}
}
