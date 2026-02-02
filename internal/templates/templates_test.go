package templates

import (
	"strings"
	"testing"

	"github.com/trabuco/trabuco/internal/config"
)

func TestNewEngine(t *testing.T) {
	engine := NewEngine()
	if engine == nil {
		t.Fatal("NewEngine() returned nil")
	}
	if engine.fs == nil {
		t.Error("Engine.fs is nil")
	}
	if engine.funcs == nil {
		t.Error("Engine.funcs is nil")
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-project", "MyProject"},
		{"simple", "Simple"},
		{"my-cool-app", "MyCoolApp"},
		{"already_snake", "AlreadySnake"},
		{"UPPER", "UPPER"},
		{"", ""},
	}

	for _, tt := range tests {
		result := toPascalCase(tt.input)
		if result != tt.expected {
			t.Errorf("toPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-project", "myProject"},
		{"simple", "simple"},
		{"my-cool-app", "myCoolApp"},
		{"already_snake", "alreadySnake"},
		{"", ""},
	}

	for _, tt := range tests {
		result := toCamelCase(tt.input)
		if result != tt.expected {
			t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestPackageToPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"com.company.project", "com/company/project"},
		{"org.example", "org/example"},
		{"single", "single"},
		{"", ""},
	}

	for _, tt := range tests {
		result := packageToPath(tt.input)
		if result != tt.expected {
			t.Errorf("packageToPath(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestInList(t *testing.T) {
	list := []string{"Model", "SQLDatastore", "API"}

	if !inList(list, "Model") {
		t.Error("inList should return true for 'Model'")
	}
	if !inList(list, "API") {
		t.Error("inList should return true for 'API'")
	}
	if inList(list, "Shared") {
		t.Error("inList should return false for 'Shared'")
	}
	if inList(nil, "Model") {
		t.Error("inList should return false for nil list")
	}
}

func TestFirstLast(t *testing.T) {
	list := []string{"a", "b", "c"}

	if first(list) != "a" {
		t.Errorf("first() = %q, want 'a'", first(list))
	}
	if last(list) != "c" {
		t.Errorf("last() = %q, want 'c'", last(list))
	}
	if first(nil) != "" {
		t.Error("first(nil) should return empty string")
	}
	if last(nil) != "" {
		t.Error("last(nil) should return empty string")
	}
}

func TestExecuteString(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "API"},
		Database:    "postgresql",
	}

	// Test basic template
	template := "Project: {{.ProjectName}}, Pascal: {{.ProjectNamePascal}}"
	result, err := engine.ExecuteString("test", template, cfg)
	if err != nil {
		t.Fatalf("ExecuteString failed: %v", err)
	}
	if result != "Project: my-platform, Pascal: MyPlatform" {
		t.Errorf("Unexpected result: %q", result)
	}

	// Test custom functions
	template = "Path: {{packagePath .GroupID}}, Camel: {{camelCase .ProjectName}}"
	result, err = engine.ExecuteString("test2", template, cfg)
	if err != nil {
		t.Fatalf("ExecuteString failed: %v", err)
	}
	if result != "Path: com/company/project, Camel: myPlatform" {
		t.Errorf("Unexpected result: %q", result)
	}

	// Test HasModule
	template = "{{if .HasModule \"Model\"}}has-model{{end}}"
	result, err = engine.ExecuteString("test3", template, cfg)
	if err != nil {
		t.Fatalf("ExecuteString failed: %v", err)
	}
	if result != "has-model" {
		t.Errorf("Expected 'has-model', got %q", result)
	}

	// Test conditional database
	template = "{{if eq .Database \"postgresql\"}}pg{{else}}other{{end}}"
	result, err = engine.ExecuteString("test4", template, cfg)
	if err != nil {
		t.Fatalf("ExecuteString failed: %v", err)
	}
	if result != "pg" {
		t.Errorf("Expected 'pg', got %q", result)
	}
}

func TestExecuteFromFile(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName:    "test-project",
		GroupID:        "com.example.test",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "Shared", "API"},
		Database:    "postgresql",
	}

	// Test README template
	result, err := engine.Execute("docs/README.md.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify content
	if !strings.Contains(result, "# test-project") {
		t.Error("README should contain project name as heading")
	}
	if !strings.Contains(result, "Java 21") {
		t.Error("README should contain Java version")
	}
	if !strings.Contains(result, "PostgreSQL") {
		t.Error("README should contain PostgreSQL for postgresql database")
	}
	if !strings.Contains(result, "Model/") {
		t.Error("README should list Model module")
	}
	if !strings.Contains(result, "API/") {
		t.Error("README should list API module")
	}
}

func TestExecuteGitignore(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "test-project",
	}

	result, err := engine.Execute("docs/gitignore.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify common patterns
	if !strings.Contains(result, "target/") {
		t.Error("gitignore should contain Maven target directory")
	}
	if !strings.Contains(result, ".idea/") {
		t.Error("gitignore should contain IntelliJ directory")
	}
	if !strings.Contains(result, ".env") {
		t.Error("gitignore should contain .env")
	}
}

func TestListTemplates(t *testing.T) {
	engine := NewEngine()

	templates, err := engine.ListTemplates("docs")
	if err != nil {
		t.Fatalf("ListTemplates failed: %v", err)
	}

	if len(templates) < 2 {
		t.Errorf("Expected at least 2 templates in docs/, got %d", len(templates))
	}

	// Check that we found expected templates
	foundReadme := false
	foundGitignore := false
	for _, tmpl := range templates {
		if strings.Contains(tmpl, "README.md.tmpl") {
			foundReadme = true
		}
		if strings.Contains(tmpl, "gitignore.tmpl") {
			foundGitignore = true
		}
	}

	if !foundReadme {
		t.Error("Expected to find README.md.tmpl in docs/")
	}
	if !foundGitignore {
		t.Error("Expected to find gitignore.tmpl in docs/")
	}
}

func TestTemplateExists(t *testing.T) {
	engine := NewEngine()

	if !engine.TemplateExists("docs/README.md.tmpl") {
		t.Error("TemplateExists should return true for existing template")
	}
	if engine.TemplateExists("nonexistent/template.tmpl") {
		t.Error("TemplateExists should return false for non-existing template")
	}
}

func TestConditionalModules(t *testing.T) {
	engine := NewEngine()

	// Test with only Model module
	cfg := &config.ProjectConfig{
		ProjectName: "minimal-app",
		GroupID:     "com.example",
		JavaVersion: "21",
		Modules:     []string{"Model"},
	}

	result, err := engine.Execute("docs/README.md.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should have Model
	if !strings.Contains(result, "Model/") {
		t.Error("README should contain Model module")
	}

	// Should NOT have other modules
	if strings.Contains(result, "SQLDatastore/") {
		t.Error("README should NOT contain SQLDatastore when not selected")
	}
	if strings.Contains(result, "API/") {
		t.Error("README should NOT contain API when not selected")
	}
	if strings.Contains(result, "PostgreSQL") {
		t.Error("README should NOT mention PostgreSQL when SQLDatastore not selected")
	}
}

// POM Template Tests

func TestParentPOM(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		ArtifactID:  "my-platform",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "API"},
	}

	result, err := engine.Execute("pom/parent.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify content
	if !strings.Contains(result, "<groupId>com.company.project</groupId>") {
		t.Error("Parent POM should contain groupId")
	}
	if !strings.Contains(result, "<artifactId>my-platform-parent</artifactId>") {
		t.Error("Parent POM should contain artifactId")
	}
	if !strings.Contains(result, "<maven.compiler.source>21</maven.compiler.source>") {
		t.Error("Parent POM should contain Java version")
	}
	if !strings.Contains(result, "<module>Model</module>") {
		t.Error("Parent POM should contain Model module")
	}
	if !strings.Contains(result, "<module>SQLDatastore</module>") {
		t.Error("Parent POM should contain SQLDatastore module")
	}
	if !strings.Contains(result, "<module>API</module>") {
		t.Error("Parent POM should contain API module")
	}
	// Shared not selected
	if strings.Contains(result, "<module>Shared</module>") {
		t.Error("Parent POM should NOT contain Shared module when not selected")
	}
}

func TestModelPOM(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model"},
	}

	result, err := engine.Execute("pom/model.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify Immutables
	if !strings.Contains(result, "org.immutables") {
		t.Error("Model POM should contain Immutables dependency")
	}
	// Verify Jackson
	if !strings.Contains(result, "jackson-databind") {
		t.Error("Model POM should contain Jackson dependency")
	}
	// Verify validation API
	if !strings.Contains(result, "jakarta.validation-api") {
		t.Error("Model POM should contain Jakarta Validation API")
	}
}

func TestSQLDatastorePOM_PostgreSQL(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore"},
		Database:    "postgresql",
	}

	result, err := engine.Execute("pom/sqldatastore.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify PostgreSQL driver
	if !strings.Contains(result, "<artifactId>postgresql</artifactId>") {
		t.Error("SQLDatastore POM should contain PostgreSQL driver")
	}
	// Verify Flyway PostgreSQL module
	if !strings.Contains(result, "flyway-database-postgresql") {
		t.Error("SQLDatastore POM should contain Flyway PostgreSQL module")
	}
	// Should NOT have MySQL
	if strings.Contains(result, "mysql-connector-j") {
		t.Error("SQLDatastore POM should NOT contain MySQL driver for postgresql")
	}
}

func TestSQLDatastorePOM_MySQL(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore"},
		Database:    "mysql",
	}

	result, err := engine.Execute("pom/sqldatastore.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify MySQL driver
	if !strings.Contains(result, "mysql-connector-j") {
		t.Error("SQLDatastore POM should contain MySQL driver")
	}
	// Verify Flyway MySQL module
	if !strings.Contains(result, "flyway-mysql") {
		t.Error("SQLDatastore POM should contain Flyway MySQL module")
	}
	// Should NOT have PostgreSQL
	if strings.Contains(result, "<artifactId>postgresql</artifactId>") {
		t.Error("SQLDatastore POM should NOT contain PostgreSQL driver for mysql")
	}
}

func TestSQLDatastorePOM_Generic(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore"},
		Database:    "generic",
	}

	result, err := engine.Execute("pom/sqldatastore.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify H2 fallback for generic
	if !strings.Contains(result, "h2") {
		t.Error("SQLDatastore POM should contain H2 for generic database")
	}
	// Should NOT have specific drivers
	if strings.Contains(result, "<artifactId>postgresql</artifactId>") {
		t.Error("SQLDatastore POM should NOT contain PostgreSQL driver for generic")
	}
	if strings.Contains(result, "mysql-connector-j") {
		t.Error("SQLDatastore POM should NOT contain MySQL driver for generic")
	}
}

func TestSharedPOM(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "Shared"},
	}

	result, err := engine.Execute("pom/shared.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify Resilience4j
	if !strings.Contains(result, "resilience4j-spring-boot3") {
		t.Error("Shared POM should contain Resilience4j dependency")
	}
	// Verify Model dependency
	if !strings.Contains(result, "<artifactId>Model</artifactId>") {
		t.Error("Shared POM should depend on Model")
	}
	// No SQLDatastore when not selected
	if strings.Contains(result, "<artifactId>SQLDatastore</artifactId>") {
		t.Error("Shared POM should NOT depend on SQLDatastore when not selected")
	}
}

func TestSharedPOM_WithSQLDatastore(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "Shared"},
		Database:    "postgresql",
	}

	result, err := engine.Execute("pom/shared.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should have SQLDatastore dependency when selected
	if !strings.Contains(result, "<artifactId>SQLDatastore</artifactId>") {
		t.Error("Shared POM should depend on SQLDatastore when selected")
	}
}

func TestAPIPOM(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
	}

	result, err := engine.Execute("pom/api.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify Spring Boot web
	if !strings.Contains(result, "spring-boot-starter-web") {
		t.Error("API POM should contain Spring Boot Web")
	}
	// Verify validation
	if !strings.Contains(result, "spring-boot-starter-validation") {
		t.Error("API POM should contain Spring Boot Validation")
	}
	// Verify Spring Boot plugin
	if !strings.Contains(result, "spring-boot-maven-plugin") {
		t.Error("API POM should contain Spring Boot Maven Plugin")
	}
	// Verify main class
	if !strings.Contains(result, "com.company.project.api.MyPlatformApiApplication") {
		t.Error("API POM should contain correct main class")
	}
}

func TestAPIPOM_WithAllModules(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "Shared", "API"},
		Database:    "postgresql",
	}

	result, err := engine.Execute("pom/api.xml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Should have all module dependencies
	if !strings.Contains(result, "<artifactId>Model</artifactId>") {
		t.Error("API POM should depend on Model")
	}
	if !strings.Contains(result, "<artifactId>SQLDatastore</artifactId>") {
		t.Error("API POM should depend on SQLDatastore when selected")
	}
	if !strings.Contains(result, "<artifactId>Shared</artifactId>") {
		t.Error("API POM should depend on Shared when selected")
	}
}

// Java Source Template Tests

func TestModelTemplates(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model"},
	}

	// Test ImmutableStyle
	result, err := engine.Execute("java/model/ImmutableStyle.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("ImmutableStyle template failed: %v", err)
	}
	if !strings.Contains(result, "package com.company.project.model;") {
		t.Error("ImmutableStyle should have correct package")
	}
	if !strings.Contains(result, "@Value.Style") {
		t.Error("ImmutableStyle should contain @Value.Style annotation")
	}

	// Test Placeholder entity (Immutable interface)
	result, err = engine.Execute("java/model/entities/Placeholder.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("Placeholder entity template failed: %v", err)
	}
	if !strings.Contains(result, "@Value.Immutable") {
		t.Error("Placeholder should have @Value.Immutable annotation")
	}
	if !strings.Contains(result, "public interface Placeholder") {
		t.Error("Placeholder should be an interface")
	}
	// Placeholder uses builder pattern, no factory methods
	if !strings.Contains(result, "Long id()") {
		t.Error("Placeholder should have id() method")
	}

	// Test PlaceholderRecord (database record)
	result, err = engine.Execute("java/model/entities/PlaceholderRecord.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("PlaceholderRecord template failed: %v", err)
	}
	if !strings.Contains(result, "@Table(\"placeholders\")") {
		t.Error("PlaceholderRecord should have @Table annotation")
	}
	if !strings.Contains(result, "public record PlaceholderRecord") {
		t.Error("PlaceholderRecord should be a record")
	}

	// Test PlaceholderRequest DTO
	result, err = engine.Execute("java/model/dto/PlaceholderRequest.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("PlaceholderRequest template failed: %v", err)
	}
	if !strings.Contains(result, "@Value.Immutable") {
		t.Error("PlaceholderRequest should have @Value.Immutable annotation")
	}
	if !strings.Contains(result, "public interface PlaceholderRequest") {
		t.Error("PlaceholderRequest should be an interface")
	}
	if !strings.Contains(result, "@NotBlank") {
		t.Error("PlaceholderRequest should have validation annotations")
	}

	// Test PlaceholderResponse DTO
	result, err = engine.Execute("java/model/dto/PlaceholderResponse.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("PlaceholderResponse template failed: %v", err)
	}
	if !strings.Contains(result, "@Value.Immutable") {
		t.Error("PlaceholderResponse should have @Value.Immutable annotation")
	}
	// PlaceholderResponse uses builder pattern, no factory methods
	if !strings.Contains(result, "Long id()") {
		t.Error("PlaceholderResponse should have id() method")
	}
}

func TestSQLDatastoreTemplates(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore"},
		Database:    "postgresql",
	}

	// Test DatabaseConfig
	result, err := engine.Execute("java/sqldatastore/config/DatabaseConfig.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("DatabaseConfig template failed: %v", err)
	}
	if !strings.Contains(result, "@Configuration") {
		t.Error("DatabaseConfig should have @Configuration")
	}
	if !strings.Contains(result, "@EnableJdbcRepositories") {
		t.Error("DatabaseConfig should have @EnableJdbcRepositories")
	}

	// Test application.yml for SQLDatastore
	result, err = engine.Execute("java/sqldatastore/resources/application.yml.tmpl", cfg)
	if err != nil {
		t.Fatalf("SQLDatastore application.yml template failed: %v", err)
	}
	if !strings.Contains(result, "jdbc:postgresql") {
		t.Error("application.yml should have PostgreSQL JDBC URL")
	}
	if !strings.Contains(result, "hikari:") {
		t.Error("application.yml should have HikariCP configuration")
	}
	if !strings.Contains(result, "flyway:") {
		t.Error("application.yml should have Flyway configuration")
	}

	// Test PlaceholderRepository
	result, err = engine.Execute("java/sqldatastore/repository/PlaceholderRepository.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("PlaceholderRepository template failed: %v", err)
	}
	if !strings.Contains(result, "extends CrudRepository") {
		t.Error("PlaceholderRepository should extend CrudRepository")
	}

	// Test migration
	result, err = engine.Execute("java/sqldatastore/migration/V1__baseline.sql.tmpl", cfg)
	if err != nil {
		t.Fatalf("Migration template failed: %v", err)
	}
	if !strings.Contains(result, "CREATE TABLE placeholders") {
		t.Error("Migration should create placeholders table")
	}
}

func TestSharedTemplates(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "Shared"},
		Database:    "postgresql",
	}

	// Test CircuitBreakerConfiguration
	result, err := engine.Execute("java/shared/config/CircuitBreakerConfiguration.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("CircuitBreakerConfiguration template failed: %v", err)
	}
	if !strings.Contains(result, "CircuitBreakerRegistry") {
		t.Error("CircuitBreakerConfiguration should use CircuitBreakerRegistry")
	}
	if !strings.Contains(result, "@Configuration") {
		t.Error("CircuitBreakerConfiguration should have @Configuration annotation")
	}

	// Test application.yml for Shared module
	result, err = engine.Execute("java/shared/resources/application.yml.tmpl", cfg)
	if err != nil {
		t.Fatalf("Shared application.yml template failed: %v", err)
	}
	if !strings.Contains(result, "resilience4j:") {
		t.Error("Shared application.yml should have resilience4j configuration")
	}
	if !strings.Contains(result, "failure-rate-threshold:") {
		t.Error("Shared application.yml should have failure-rate-threshold")
	}
	if !strings.Contains(result, "CB_FAILURE_RATE_THRESHOLD") {
		t.Error("Shared application.yml should have environment variable override for failure rate")
	}

	// Test PlaceholderService with SQLDatastore
	result, err = engine.Execute("java/shared/service/PlaceholderService.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("PlaceholderService template failed: %v", err)
	}
	if !strings.Contains(result, "PlaceholderRepository") {
		t.Error("PlaceholderService should use PlaceholderRepository when SQLDatastore included")
	}
	if !strings.Contains(result, "@CircuitBreaker") {
		t.Error("PlaceholderService should have CircuitBreaker annotation")
	}
}

func TestSharedTemplates_WithoutSQLDatastore(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "Shared"},
	}

	// Test PlaceholderService without any datastore
	result, err := engine.Execute("java/shared/service/PlaceholderService.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("PlaceholderService template failed: %v", err)
	}
	if !strings.Contains(result, "TODO: No datastore module included") {
		t.Error("PlaceholderService should have TODO when no datastore module included")
	}
}

func TestAPITemplates(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "Shared", "API"},
		Database:    "postgresql",
	}

	// Test Application
	result, err := engine.Execute("java/api/Application.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("Application template failed: %v", err)
	}
	if !strings.Contains(result, "MyPlatformApiApplication") {
		t.Error("Application should have correct class name")
	}
	if !strings.Contains(result, "@SpringBootApplication") {
		t.Error("Application should have @SpringBootApplication")
	}
	if !strings.Contains(result, "com.company.project.shared") {
		t.Error("Application should scan shared package when included")
	}

	// Test HealthController
	result, err = engine.Execute("java/api/controller/HealthController.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("HealthController template failed: %v", err)
	}
	if !strings.Contains(result, "@GetMapping") {
		t.Error("HealthController should have @GetMapping")
	}

	// Test PlaceholderController with Shared
	result, err = engine.Execute("java/api/controller/PlaceholderController.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("PlaceholderController template failed: %v", err)
	}
	if !strings.Contains(result, "PlaceholderService") {
		t.Error("PlaceholderController should use PlaceholderService when Shared included")
	}

	// Test application.yml
	result, err = engine.Execute("java/api/resources/application.yml.tmpl", cfg)
	if err != nil {
		t.Fatalf("application.yml template failed: %v", err)
	}
	if !strings.Contains(result, "jdbc:postgresql") {
		t.Error("application.yml should have PostgreSQL URL")
	}
	if !strings.Contains(result, "resilience4j") {
		t.Error("application.yml should have resilience4j config when Shared included")
	}
}

func TestAPITemplates_MinimalModules(t *testing.T) {
	engine := NewEngine()

	cfg := &config.ProjectConfig{
		ProjectName: "my-platform",
		GroupID:     "com.company.project",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
	}

	// Test Application without other modules
	result, err := engine.Execute("java/api/Application.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("Application template failed: %v", err)
	}
	if strings.Contains(result, "com.company.project.shared") {
		t.Error("Application should NOT scan shared package when not included")
	}

	// Test PlaceholderController without Shared or SQLDatastore
	result, err = engine.Execute("java/api/controller/PlaceholderController.java.tmpl", cfg)
	if err != nil {
		t.Fatalf("PlaceholderController template failed: %v", err)
	}
	if !strings.Contains(result, "NOT_IMPLEMENTED") {
		t.Error("PlaceholderController should return NOT_IMPLEMENTED when no data modules")
	}

	// Test application.yml without SQLDatastore
	result, err = engine.Execute("java/api/resources/application.yml.tmpl", cfg)
	if err != nil {
		t.Fatalf("application.yml template failed: %v", err)
	}
	if strings.Contains(result, "datasource") {
		t.Error("application.yml should NOT have datasource when SQLDatastore not included")
	}
}
