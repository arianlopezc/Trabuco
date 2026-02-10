package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

func TestDetectProject(t *testing.T) {
	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "trabuco-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("not a Maven project", func(t *testing.T) {
		emptyDir, err := os.MkdirTemp("", "empty-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(emptyDir)

		_, err = DetectProject(emptyDir)
		if err == nil {
			t.Error("Expected error for non-Maven project")
		}
	})

	t.Run("Maven project but not Trabuco", func(t *testing.T) {
		nonTrabucoDir, err := os.MkdirTemp("", "maven-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(nonTrabucoDir)

		// Create a basic pom.xml without Model module
		pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example</groupId>
    <artifactId>test</artifactId>
    <version>1.0-SNAPSHOT</version>
    <modules>
        <module>SomeModule</module>
    </modules>
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
    </properties>
</project>`
		if err := os.WriteFile(filepath.Join(nonTrabucoDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
			t.Fatalf("Failed to write pom.xml: %v", err)
		}

		_, err = DetectProject(nonTrabucoDir)
		if err == nil {
			t.Error("Expected error for non-Trabuco project")
		}
	})

	t.Run("Trabuco project with metadata", func(t *testing.T) {
		trabucoDir := createTestTrabucoProject(t)
		defer os.RemoveAll(trabucoDir)

		metadata, err := DetectProject(trabucoDir)
		if err != nil {
			t.Fatalf("Failed to detect Trabuco project: %v", err)
		}

		if metadata.ProjectName != "test-project" {
			t.Errorf("Expected project name 'test-project', got '%s'", metadata.ProjectName)
		}
		if metadata.GroupID != "com.example.test" {
			t.Errorf("Expected group ID 'com.example.test', got '%s'", metadata.GroupID)
		}
	})
}

func TestDoctorRun(t *testing.T) {
	trabucoDir := createTestTrabucoProject(t)
	defer os.RemoveAll(trabucoDir)

	doc := New(trabucoDir, "1.0.0")
	result, err := doc.Run()
	if err != nil {
		t.Fatalf("Doctor run failed: %v", err)
	}

	if result.Project != "test-project" {
		t.Errorf("Expected project 'test-project', got '%s'", result.Project)
	}

	// All checks should pass for a valid project
	for _, check := range result.Checks {
		if check.Status == SeverityError {
			t.Errorf("Check %s failed: %s", check.ID, check.Message)
		}
	}
}

func TestDoctorValidate(t *testing.T) {
	t.Run("healthy project", func(t *testing.T) {
		trabucoDir := createTestTrabucoProject(t)
		defer os.RemoveAll(trabucoDir)

		doc := New(trabucoDir, "1.0.0")
		err := doc.Validate()
		if err != nil {
			t.Errorf("Expected no error for healthy project, got: %v", err)
		}
	})

	t.Run("broken project", func(t *testing.T) {
		brokenDir, err := os.MkdirTemp("", "broken-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(brokenDir)

		doc := New(brokenDir, "1.0.0")
		err = doc.Validate()
		if err == nil {
			t.Error("Expected error for broken project")
		}
	})
}

func TestQuickCheck(t *testing.T) {
	t.Run("valid project", func(t *testing.T) {
		trabucoDir := createTestTrabucoProject(t)
		defer os.RemoveAll(trabucoDir)

		metadata, err := QuickCheck(trabucoDir)
		if err != nil {
			t.Fatalf("QuickCheck failed: %v", err)
		}

		if metadata.ProjectName != "test-project" {
			t.Errorf("Expected project name 'test-project', got '%s'", metadata.ProjectName)
		}
	})

	t.Run("not a Maven project", func(t *testing.T) {
		emptyDir, err := os.MkdirTemp("", "empty-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(emptyDir)

		_, err = QuickCheck(emptyDir)
		if _, ok := err.(*NotMavenProjectError); !ok {
			t.Errorf("Expected NotMavenProjectError, got: %v", err)
		}
	})
}

func TestDoctorResult(t *testing.T) {
	result := &DoctorResult{
		Project:  "test",
		Location: "/tmp/test",
		Checks: []CheckResult{
			{ID: "CHECK1", Status: SeverityPass},
			{ID: "CHECK2", Status: SeverityWarn, CanAutoFix: true},
			{ID: "CHECK3", Status: SeverityError},
		},
	}
	result.ComputeSummary()

	if !result.HasErrors() {
		t.Error("Expected HasErrors() to be true")
	}

	if !result.HasWarnings() {
		t.Error("Expected HasWarnings() to be true")
	}

	if result.IsHealthy() {
		t.Error("Expected IsHealthy() to be false")
	}

	if result.Summary.Passed != 1 {
		t.Errorf("Expected 1 passed, got %d", result.Summary.Passed)
	}
	if result.Summary.Warnings != 1 {
		t.Errorf("Expected 1 warning, got %d", result.Summary.Warnings)
	}
	if result.Summary.Errors != 1 {
		t.Errorf("Expected 1 error, got %d", result.Summary.Errors)
	}

	fixable := result.GetFixableChecks()
	if len(fixable) != 1 {
		t.Errorf("Expected 1 fixable check, got %d", len(fixable))
	}
}

func TestSeverity(t *testing.T) {
	tests := []struct {
		severity Severity
		expected string
	}{
		{SeverityPass, "PASS"},
		{SeverityWarn, "WARN"},
		{SeverityError, "ERROR"},
	}

	for _, tt := range tests {
		if got := tt.severity.String(); got != tt.expected {
			t.Errorf("Severity.String() = %s, want %s", got, tt.expected)
		}
	}
}

// Helper function to create a test Trabuco project
func createTestTrabucoProject(t *testing.T) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "trabuco-project-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create parent pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.example.test</groupId>
    <artifactId>test-project-parent</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>pom</packaging>
    <modules>
        <module>Model</module>
        <module>API</module>
    </modules>
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
    </properties>
    <dependencyManagement>
        <dependencies>
        </dependencies>
    </dependencyManagement>
</project>`
	if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to write pom.xml: %v", err)
	}

	// Create .trabuco.json
	metadata := &config.ProjectMetadata{
		Version:     "1.0.0",
		GeneratedAt: "2024-01-01T00:00:00Z",
		ProjectName: "test-project",
		GroupID:     "com.example.test",
		ArtifactID:  "test-project",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
	}
	if err := config.SaveMetadata(tempDir, metadata); err != nil {
		t.Fatalf("Failed to save metadata: %v", err)
	}

	// Create Model module structure
	modelPath := filepath.Join(tempDir, "Model", "src", "main", "java", "com", "example", "test", "model")
	if err := os.MkdirAll(modelPath, 0755); err != nil {
		t.Fatalf("Failed to create Model directory: %v", err)
	}

	// Create Model pom.xml
	modelPomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>com.example.test</groupId>
        <artifactId>test-project-parent</artifactId>
        <version>1.0-SNAPSHOT</version>
    </parent>
    <artifactId>model</artifactId>
</project>`
	if err := os.WriteFile(filepath.Join(tempDir, "Model", "pom.xml"), []byte(modelPomContent), 0644); err != nil {
		t.Fatalf("Failed to write Model pom.xml: %v", err)
	}

	// Create API module structure
	apiPath := filepath.Join(tempDir, "API", "src", "main", "java", "com", "example", "test", "api")
	if err := os.MkdirAll(apiPath, 0755); err != nil {
		t.Fatalf("Failed to create API directory: %v", err)
	}

	// Create API pom.xml
	apiPomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <modelVersion>4.0.0</modelVersion>
    <parent>
        <groupId>com.example.test</groupId>
        <artifactId>test-project-parent</artifactId>
        <version>1.0-SNAPSHOT</version>
    </parent>
    <artifactId>api</artifactId>
</project>`
	if err := os.WriteFile(filepath.Join(tempDir, "API", "pom.xml"), []byte(apiPomContent), 0644); err != nil {
		t.Fatalf("Failed to write API pom.xml: %v", err)
	}

	return tempDir
}
