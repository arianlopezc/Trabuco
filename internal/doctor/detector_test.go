package doctor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

func TestInferFromPOM(t *testing.T) {
	t.Run("infers metadata from valid Trabuco project", func(t *testing.T) {
		tempDir := createTrabucoProjectWithoutMetadata(t)
		defer os.RemoveAll(tempDir)

		metadata, err := InferFromPOM(tempDir)
		if err != nil {
			t.Fatalf("InferFromPOM failed: %v", err)
		}

		if metadata.GroupID != "com.example.test" {
			t.Errorf("Expected groupId 'com.example.test', got '%s'", metadata.GroupID)
		}
		if metadata.JavaVersion != "21" {
			t.Errorf("Expected javaVersion '21', got '%s'", metadata.JavaVersion)
		}
		if len(metadata.Modules) == 0 {
			t.Error("Expected at least one module")
		}
	})

	t.Run("fails for non-Trabuco project", func(t *testing.T) {
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

		_, err = InferFromPOM(tempDir)
		if err == nil {
			t.Error("Expected error for non-Trabuco project")
		}
	})
}

func TestParseParentPOM(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pom-parse-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("parses valid POM", func(t *testing.T) {
		pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
    <artifactId>test-parent</artifactId>
    <modules>
        <module>Model</module>
        <module>API</module>
    </modules>
    <properties>
        <maven.compiler.source>21</maven.compiler.source>
        <maven.compiler.target>21</maven.compiler.target>
    </properties>
</project>`
		pomPath := filepath.Join(tempDir, "pom.xml")
		if err := os.WriteFile(pomPath, []byte(pomContent), 0644); err != nil {
			t.Fatalf("Failed to write pom.xml: %v", err)
		}

		pom, err := ParseParentPOM(pomPath)
		if err != nil {
			t.Fatalf("ParseParentPOM failed: %v", err)
		}

		if pom.GroupID != "com.example" {
			t.Errorf("Expected groupId 'com.example', got '%s'", pom.GroupID)
		}
		if pom.ArtifactID != "test-parent" {
			t.Errorf("Expected artifactId 'test-parent', got '%s'", pom.ArtifactID)
		}
		if len(pom.Modules) != 2 {
			t.Errorf("Expected 2 modules, got %d", len(pom.Modules))
		}
		if pom.Properties.JavaSource != "21" {
			t.Errorf("Expected javaSource '21', got '%s'", pom.Properties.JavaSource)
		}
	})

	t.Run("fails with invalid XML", func(t *testing.T) {
		pomPath := filepath.Join(tempDir, "invalid.xml")
		if err := os.WriteFile(pomPath, []byte("not xml"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		_, err := ParseParentPOM(pomPath)
		if err == nil {
			t.Error("Expected error for invalid XML")
		}
	})

	t.Run("fails with missing file", func(t *testing.T) {
		_, err := ParseParentPOM(filepath.Join(tempDir, "nonexistent.xml"))
		if err == nil {
			t.Error("Expected error for missing file")
		}
	})
}

func TestGetModulesFromPOM(t *testing.T) {
	tempDir := createTestTrabucoProject(t)
	defer os.RemoveAll(tempDir)

	modules, err := GetModulesFromPOM(tempDir)
	if err != nil {
		t.Fatalf("GetModulesFromPOM failed: %v", err)
	}

	if len(modules) != 2 {
		t.Errorf("Expected 2 modules, got %d", len(modules))
	}

	hasModel := false
	hasAPI := false
	for _, m := range modules {
		if m == "Model" {
			hasModel = true
		}
		if m == "API" {
			hasAPI = true
		}
	}

	if !hasModel {
		t.Error("Expected Model module")
	}
	if !hasAPI {
		t.Error("Expected API module")
	}
}

func TestGetJavaVersionFromPOM(t *testing.T) {
	tempDir := createTestTrabucoProject(t)
	defer os.RemoveAll(tempDir)

	version, err := GetJavaVersionFromPOM(tempDir)
	if err != nil {
		t.Fatalf("GetJavaVersionFromPOM failed: %v", err)
	}

	if version != "21" {
		t.Errorf("Expected version '21', got '%s'", version)
	}
}

func TestGetGroupIDFromPOM(t *testing.T) {
	tempDir := createTestTrabucoProject(t)
	defer os.RemoveAll(tempDir)

	groupID, err := GetGroupIDFromPOM(tempDir)
	if err != nil {
		t.Fatalf("GetGroupIDFromPOM failed: %v", err)
	}

	if groupID != "com.example.test" {
		t.Errorf("Expected groupId 'com.example.test', got '%s'", groupID)
	}
}

func TestIsValidPOM(t *testing.T) {
	tempDir := createTestTrabucoProject(t)
	defer os.RemoveAll(tempDir)

	t.Run("valid POM", func(t *testing.T) {
		if !IsValidPOM(filepath.Join(tempDir, "pom.xml")) {
			t.Error("Expected POM to be valid")
		}
	})

	t.Run("invalid POM", func(t *testing.T) {
		invalidPath := filepath.Join(tempDir, "invalid.xml")
		if err := os.WriteFile(invalidPath, []byte("not xml"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		if IsValidPOM(invalidPath) {
			t.Error("Expected POM to be invalid")
		}
	})

	t.Run("missing POM", func(t *testing.T) {
		if IsValidPOM(filepath.Join(tempDir, "nonexistent.xml")) {
			t.Error("Expected POM to be invalid (missing)")
		}
	})
}

func TestHasRequiredPOMSections(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pom-sections-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("has all sections", func(t *testing.T) {
		pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <modules>
        <module>Model</module>
    </modules>
    <properties>
        <java.version>21</java.version>
    </properties>
</project>`
		if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
			t.Fatalf("Failed to write pom.xml: %v", err)
		}

		hasModules, hasProperties, err := HasRequiredPOMSections(tempDir)
		if err != nil {
			t.Fatalf("HasRequiredPOMSections failed: %v", err)
		}

		if !hasModules {
			t.Error("Expected hasModules to be true")
		}
		if !hasProperties {
			t.Error("Expected hasProperties to be true")
		}
	})

	t.Run("missing modules section", func(t *testing.T) {
		pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <properties>
        <java.version>21</java.version>
    </properties>
</project>`
		if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
			t.Fatalf("Failed to write pom.xml: %v", err)
		}

		hasModules, hasProperties, err := HasRequiredPOMSections(tempDir)
		if err != nil {
			t.Fatalf("HasRequiredPOMSections failed: %v", err)
		}

		if hasModules {
			t.Error("Expected hasModules to be false")
		}
		if !hasProperties {
			t.Error("Expected hasProperties to be true")
		}
	})
}

func TestParseModulePOM(t *testing.T) {
	tempDir := createTestTrabucoProject(t)
	defer os.RemoveAll(tempDir)

	info, err := ParseModulePOM(filepath.Join(tempDir, "Model", "pom.xml"))
	if err != nil {
		t.Fatalf("ParseModulePOM failed: %v", err)
	}

	if info.ArtifactID != "model" {
		t.Errorf("Expected artifactId 'model', got '%s'", info.ArtifactID)
	}
	if info.Parent.GroupID != "com.example.test" {
		t.Errorf("Expected parent groupId 'com.example.test', got '%s'", info.Parent.GroupID)
	}
}

func TestGetRequiredDockerServices(t *testing.T) {
	tests := []struct {
		name     string
		metadata *config.ProjectMetadata
		expected []string
	}{
		{
			name: "PostgreSQL datastore",
			metadata: &config.ProjectMetadata{
				Modules:  []string{"Model", "SQLDatastore"},
				Database: "postgresql",
			},
			expected: []string{"postgres"},
		},
		{
			name: "MySQL datastore",
			metadata: &config.ProjectMetadata{
				Modules:  []string{"Model", "SQLDatastore"},
				Database: "mysql",
			},
			expected: []string{"mysql"},
		},
		{
			name: "MongoDB datastore",
			metadata: &config.ProjectMetadata{
				Modules:       []string{"Model", "NoSQLDatastore"},
				NoSQLDatabase: "mongodb",
			},
			expected: []string{"mongodb"},
		},
		{
			name: "Redis datastore",
			metadata: &config.ProjectMetadata{
				Modules:       []string{"Model", "NoSQLDatastore"},
				NoSQLDatabase: "redis",
			},
			expected: []string{"redis"},
		},
		{
			name: "Kafka broker",
			metadata: &config.ProjectMetadata{
				Modules:       []string{"Model", "EventConsumer"},
				MessageBroker: "kafka",
			},
			expected: []string{"kafka"},
		},
		{
			name: "RabbitMQ broker",
			metadata: &config.ProjectMetadata{
				Modules:       []string{"Model", "EventConsumer"},
				MessageBroker: "rabbitmq",
			},
			expected: []string{"rabbitmq"},
		},
		{
			name: "SQS broker",
			metadata: &config.ProjectMetadata{
				Modules:       []string{"Model", "EventConsumer"},
				MessageBroker: "sqs",
			},
			expected: []string{"localstack"},
		},
		{
			name: "Pub/Sub broker",
			metadata: &config.ProjectMetadata{
				Modules:       []string{"Model", "EventConsumer"},
				MessageBroker: "pubsub",
			},
			expected: []string{"pubsub-emulator"},
		},
		{
			name: "Worker with no datastore needs postgres-jobrunr",
			metadata: &config.ProjectMetadata{
				Modules: []string{"Model", "Worker"},
			},
			expected: []string{"postgres-jobrunr"},
		},
		{
			name:     "No services needed",
			metadata: &config.ProjectMetadata{Modules: []string{"Model", "API"}},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			services := GetRequiredDockerServices(tt.metadata)

			if len(services) != len(tt.expected) {
				t.Errorf("Expected %d services, got %d: %v", len(tt.expected), len(services), services)
			}

			for _, expected := range tt.expected {
				found := false
				for _, s := range services {
					if s == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected service '%s' not found in %v", expected, services)
				}
			}
		})
	}
}

func TestParseDockerCompose(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "docker-compose-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("parses valid docker-compose", func(t *testing.T) {
		content := `services:
  postgres:
    image: postgres:16-alpine
  redis:
    image: redis:7-alpine
`
		composePath := filepath.Join(tempDir, "docker-compose.yml")
		if err := os.WriteFile(composePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write docker-compose.yml: %v", err)
		}

		dc, err := ParseDockerCompose(composePath)
		if err != nil {
			t.Fatalf("ParseDockerCompose failed: %v", err)
		}

		if len(dc.Services) != 2 {
			t.Errorf("Expected 2 services, got %d", len(dc.Services))
		}

		if _, ok := dc.Services["postgres"]; !ok {
			t.Error("Expected postgres service")
		}
		if _, ok := dc.Services["redis"]; !ok {
			t.Error("Expected redis service")
		}
	})

	t.Run("fails with invalid YAML", func(t *testing.T) {
		composePath := filepath.Join(tempDir, "invalid.yml")
		if err := os.WriteFile(composePath, []byte("not: valid: yaml: {{"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		_, err := ParseDockerCompose(composePath)
		if err == nil {
			t.Error("Expected error for invalid YAML")
		}
	})
}

func TestParseApplicationYAML(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "app-yaml-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("parses database config", func(t *testing.T) {
		content := `spring:
  datasource:
    url: jdbc:postgresql://localhost:5432/testdb
`
		yamlPath := filepath.Join(tempDir, "application.yml")
		if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write application.yml: %v", err)
		}

		appConfig, err := ParseApplicationYAML(yamlPath)
		if err != nil {
			t.Fatalf("ParseApplicationYAML failed: %v", err)
		}

		if appConfig.Spring.Datasource.URL == "" {
			t.Error("Expected datasource URL to be set")
		}
	})

	t.Run("parses Kafka config", func(t *testing.T) {
		content := `spring:
  kafka:
    bootstrap-servers: localhost:9092
`
		yamlPath := filepath.Join(tempDir, "kafka-app.yml")
		if err := os.WriteFile(yamlPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		appConfig, err := ParseApplicationYAML(yamlPath)
		if err != nil {
			t.Fatalf("ParseApplicationYAML failed: %v", err)
		}

		if appConfig.Spring.Kafka.BootstrapServers == "" {
			t.Error("Expected kafka bootstrap-servers to be set")
		}
	})
}
