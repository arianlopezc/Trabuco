package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

func TestProjectStructureCheck(t *testing.T) {
	check := NewProjectStructureCheck()

	t.Run("passes when pom.xml exists", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("fails when pom.xml missing", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "empty-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityError {
			t.Errorf("Expected ERROR, got %s", result.Status)
		}
	})
}

func TestTrabucoProjectCheck(t *testing.T) {
	check := NewTrabucoProjectCheck()

	t.Run("passes with .trabuco.json", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("passes with Model module structure", func(t *testing.T) {
		tempDir := createTrabucoProjectWithoutMetadata(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("fails without Model module", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "non-trabuco-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		// Create pom.xml without Model module
		pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
    <artifactId>test</artifactId>
    <modules>
        <module>SomeModule</module>
    </modules>
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
    </properties>
</project>`
		if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
			t.Fatalf("Failed to write pom.xml: %v", err)
		}

		result := check.Check(tempDir, nil)
		if result.Status != SeverityError {
			t.Errorf("Expected ERROR, got %s", result.Status)
		}
	})
}

func TestMetadataExistsCheck(t *testing.T) {
	check := NewMetadataExistsCheck()

	t.Run("passes when metadata exists", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s", result.Status)
		}
	})

	t.Run("warns when metadata missing", func(t *testing.T) {
		tempDir := createTrabucoProjectWithoutMetadata(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityWarn {
			t.Errorf("Expected WARN, got %s", result.Status)
		}
		if !result.CanAutoFix {
			t.Error("Expected CanAutoFix to be true")
		}
	})
}

func TestMetadataValidCheck(t *testing.T) {
	check := NewMetadataValidCheck()

	t.Run("passes with valid metadata", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("skips when metadata doesn't exist", func(t *testing.T) {
		tempDir := createTrabucoProjectWithoutMetadata(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS (skip), got %s", result.Status)
		}
	})

	t.Run("fails with invalid JSON", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Corrupt the metadata file
		if err := os.WriteFile(filepath.Join(tempDir, ".trabuco.json"), []byte("invalid json"), 0644); err != nil {
			t.Fatalf("Failed to corrupt metadata: %v", err)
		}

		result := check.Check(tempDir, nil)
		if result.Status != SeverityError {
			t.Errorf("Expected ERROR, got %s", result.Status)
		}
	})

	t.Run("fails with missing required fields", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Write metadata with missing fields
		if err := os.WriteFile(filepath.Join(tempDir, ".trabuco.json"), []byte(`{"version":"1.0.0"}`), 0644); err != nil {
			t.Fatalf("Failed to write metadata: %v", err)
		}

		result := check.Check(tempDir, nil)
		if result.Status != SeverityError {
			t.Errorf("Expected ERROR, got %s", result.Status)
		}
	})
}

func TestMetadataSyncCheck(t *testing.T) {
	check := NewMetadataSyncCheck()

	t.Run("passes when in sync", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		meta, _ := config.LoadMetadata(tempDir)
		result := check.Check(tempDir, meta)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("warns when out of sync", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Modify POM to add a module not in metadata
		pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example.test</groupId>
    <artifactId>test-project-parent</artifactId>
    <modules>
        <module>Model</module>
        <module>API</module>
        <module>NewModule</module>
    </modules>
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
    </properties>
</project>`
		if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
			t.Fatalf("Failed to write pom.xml: %v", err)
		}

		meta, _ := config.LoadMetadata(tempDir)
		result := check.Check(tempDir, meta)
		if result.Status != SeverityWarn {
			t.Errorf("Expected WARN, got %s", result.Status)
		}
		if !result.CanAutoFix {
			t.Error("Expected CanAutoFix to be true")
		}
	})
}

func TestParentPOMValidCheck(t *testing.T) {
	check := NewParentPOMValidCheck()

	t.Run("passes with valid POM", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("fails without modules section", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "no-modules-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tempDir)

		pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
    <artifactId>test</artifactId>
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
    </properties>
</project>`
		if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
			t.Fatalf("Failed to write pom.xml: %v", err)
		}

		result := check.Check(tempDir, nil)
		if result.Status != SeverityError {
			t.Errorf("Expected ERROR, got %s", result.Status)
		}
	})
}

func TestModulePOMsExistCheck(t *testing.T) {
	check := NewModulePOMsExistCheck()

	t.Run("passes when all module POMs exist", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("fails when module POM missing", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Remove API pom.xml
		if err := os.Remove(filepath.Join(tempDir, "API", "pom.xml")); err != nil {
			t.Fatalf("Failed to remove API pom.xml: %v", err)
		}

		result := check.Check(tempDir, nil)
		if result.Status != SeverityError {
			t.Errorf("Expected ERROR, got %s", result.Status)
		}
	})
}

func TestModuleDirsExistCheck(t *testing.T) {
	check := NewModuleDirsExistCheck()

	t.Run("passes when all module dirs exist", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("fails when module dir missing", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Remove API directory entirely
		if err := os.RemoveAll(filepath.Join(tempDir, "API")); err != nil {
			t.Fatalf("Failed to remove API directory: %v", err)
		}

		result := check.Check(tempDir, nil)
		if result.Status != SeverityError {
			t.Errorf("Expected ERROR, got %s", result.Status)
		}
	})
}

func TestJavaVersionConsistentCheck(t *testing.T) {
	check := NewJavaVersionConsistentCheck()

	t.Run("passes with consistent version", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})
}

func TestGroupIDConsistentCheck(t *testing.T) {
	check := NewGroupIDConsistentCheck()

	t.Run("passes with consistent group ID", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("warns with inconsistent group ID", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Change API pom.xml to have different groupId
		apiPomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.different.group</groupId>
    <artifactId>api</artifactId>
    <parent>
        <groupId>com.example.test</groupId>
        <artifactId>test-project-parent</artifactId>
        <version>1.0-SNAPSHOT</version>
    </parent>
</project>`
		if err := os.WriteFile(filepath.Join(tempDir, "API", "pom.xml"), []byte(apiPomContent), 0644); err != nil {
			t.Fatalf("Failed to write API pom.xml: %v", err)
		}

		result := check.Check(tempDir, nil)
		if result.Status != SeverityWarn {
			t.Errorf("Expected WARN, got %s", result.Status)
		}
	})
}

func TestDockerComposeSyncCheck(t *testing.T) {
	check := NewDockerComposeSyncCheck()

	t.Run("passes when docker-compose not needed", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Project without datastore doesn't need docker-compose
		meta, _ := config.LoadMetadata(tempDir)
		result := check.Check(tempDir, meta)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})

	t.Run("warns when docker-compose missing but needed", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Add SQLDatastore to metadata
		meta, _ := config.LoadMetadata(tempDir)
		meta.Modules = append(meta.Modules, "SQLDatastore")
		meta.Database = "postgresql"
		if err := config.SaveMetadata(tempDir, meta); err != nil {
			t.Fatalf("Failed to save metadata: %v", err)
		}

		result := check.Check(tempDir, meta)
		if result.Status != SeverityWarn {
			t.Errorf("Expected WARN, got %s", result.Status)
		}
	})

	t.Run("passes when docker-compose exists with required services", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		// Add SQLDatastore to metadata
		meta, _ := config.LoadMetadata(tempDir)
		meta.Modules = append(meta.Modules, "SQLDatastore")
		meta.Database = "postgresql"
		if err := config.SaveMetadata(tempDir, meta); err != nil {
			t.Fatalf("Failed to save metadata: %v", err)
		}

		// Create docker-compose.yml with postgres service
		dockerCompose := `services:
  postgres:
    image: postgres:16-alpine
`
		if err := os.WriteFile(filepath.Join(tempDir, "docker-compose.yml"), []byte(dockerCompose), 0644); err != nil {
			t.Fatalf("Failed to write docker-compose.yml: %v", err)
		}

		result := check.Check(tempDir, meta)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})
}

func TestCrossModuleDepsCheck(t *testing.T) {
	check := NewCrossModuleDepsCheck()

	t.Run("passes with valid module POMs", func(t *testing.T) {
		tempDir := createTestTrabucoProject(t)
		defer os.RemoveAll(tempDir)

		result := check.Check(tempDir, nil)
		if result.Status != SeverityPass {
			t.Errorf("Expected PASS, got %s: %s", result.Status, result.Message)
		}
	})
}

func TestGetAllChecks(t *testing.T) {
	checks := GetAllChecks()

	expectedCount := 12
	if len(checks) != expectedCount {
		t.Errorf("Expected %d checks, got %d", expectedCount, len(checks))
	}

	// Verify all checks have unique IDs
	ids := make(map[string]bool)
	for _, check := range checks {
		if ids[check.ID()] {
			t.Errorf("Duplicate check ID: %s", check.ID())
		}
		ids[check.ID()] = true
	}
}

func TestGetChecksByCategory(t *testing.T) {
	t.Run("structure category", func(t *testing.T) {
		checks := GetChecksByCategory("structure")
		if len(checks) == 0 {
			t.Error("Expected at least one structure check")
		}
		for _, check := range checks {
			if check.Category() != "structure" {
				t.Errorf("Expected category 'structure', got '%s'", check.Category())
			}
		}
	})

	t.Run("metadata category", func(t *testing.T) {
		checks := GetChecksByCategory("metadata")
		if len(checks) == 0 {
			t.Error("Expected at least one metadata check")
		}
		for _, check := range checks {
			if check.Category() != "metadata" {
				t.Errorf("Expected category 'metadata', got '%s'", check.Category())
			}
		}
	})

	t.Run("consistency category", func(t *testing.T) {
		checks := GetChecksByCategory("consistency")
		if len(checks) == 0 {
			t.Error("Expected at least one consistency check")
		}
		for _, check := range checks {
			if check.Category() != "consistency" {
				t.Errorf("Expected category 'consistency', got '%s'", check.Category())
			}
		}
	})

	t.Run("invalid category returns empty", func(t *testing.T) {
		checks := GetChecksByCategory("invalid")
		if len(checks) != 0 {
			t.Errorf("Expected 0 checks for invalid category, got %d", len(checks))
		}
	})
}

// Helper function to create a Trabuco project without .trabuco.json
func createTrabucoProjectWithoutMetadata(t *testing.T) string {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "trabuco-no-meta-*")
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
    </modules>
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
    </properties>
</project>`
	if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to write pom.xml: %v", err)
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

	return tempDir
}
