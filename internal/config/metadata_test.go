package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadMetadata(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "metadata-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("loads valid metadata", func(t *testing.T) {
		content := `{
  "version": "1.0.0",
  "generatedAt": "2024-01-01T00:00:00Z",
  "projectName": "test-project",
  "groupId": "com.example.test",
  "artifactId": "test-project",
  "javaVersion": "21",
  "modules": ["Model", "API"],
  "database": "postgresql",
  "aiAgents": ["claude"]
}`
		metadataPath := filepath.Join(tempDir, MetadataFileName)
		if err := os.WriteFile(metadataPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write metadata: %v", err)
		}

		meta, err := LoadMetadata(tempDir)
		if err != nil {
			t.Fatalf("LoadMetadata failed: %v", err)
		}

		if meta.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", meta.Version)
		}
		if meta.ProjectName != "test-project" {
			t.Errorf("Expected projectName 'test-project', got '%s'", meta.ProjectName)
		}
		if meta.GroupID != "com.example.test" {
			t.Errorf("Expected groupId 'com.example.test', got '%s'", meta.GroupID)
		}
		if meta.JavaVersion != "21" {
			t.Errorf("Expected javaVersion '21', got '%s'", meta.JavaVersion)
		}
		if len(meta.Modules) != 2 {
			t.Errorf("Expected 2 modules, got %d", len(meta.Modules))
		}
		if meta.Database != "postgresql" {
			t.Errorf("Expected database 'postgresql', got '%s'", meta.Database)
		}
	})

	t.Run("fails with missing file", func(t *testing.T) {
		emptyDir, err := os.MkdirTemp("", "empty-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(emptyDir)

		_, err = LoadMetadata(emptyDir)
		if err == nil {
			t.Error("Expected error for missing file")
		}
	})

	t.Run("fails with invalid JSON", func(t *testing.T) {
		invalidDir, err := os.MkdirTemp("", "invalid-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(invalidDir)

		if err := os.WriteFile(filepath.Join(invalidDir, MetadataFileName), []byte("not json"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		_, err = LoadMetadata(invalidDir)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestSaveMetadata(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "save-metadata-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	meta := &ProjectMetadata{
		Version:       "1.0.0",
		GeneratedAt:   "2024-01-01T00:00:00Z",
		ProjectName:   "test-project",
		GroupID:       "com.example.test",
		ArtifactID:    "test-project",
		JavaVersion:   "21",
		Modules:       []string{"Model", "API"},
		Database:      "postgresql",
		NoSQLDatabase: "",
		MessageBroker: "kafka",
		AIAgents:      []string{"claude", "cursor"},
	}

	if err := SaveMetadata(tempDir, meta); err != nil {
		t.Fatalf("SaveMetadata failed: %v", err)
	}

	// Verify file exists
	metadataPath := filepath.Join(tempDir, MetadataFileName)
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		t.Error("Metadata file should exist after save")
	}

	// Load and verify content
	loaded, err := LoadMetadata(tempDir)
	if err != nil {
		t.Fatalf("Failed to load saved metadata: %v", err)
	}

	if loaded.ProjectName != meta.ProjectName {
		t.Errorf("Expected projectName '%s', got '%s'", meta.ProjectName, loaded.ProjectName)
	}
	if loaded.MessageBroker != meta.MessageBroker {
		t.Errorf("Expected messageBroker '%s', got '%s'", meta.MessageBroker, loaded.MessageBroker)
	}
	if len(loaded.AIAgents) != len(meta.AIAgents) {
		t.Errorf("Expected %d AI agents, got %d", len(meta.AIAgents), len(loaded.AIAgents))
	}
}

func TestMetadataExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "exists-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("returns false when file doesn't exist", func(t *testing.T) {
		if MetadataExists(tempDir) {
			t.Error("Expected false for non-existent metadata")
		}
	})

	t.Run("returns true when file exists", func(t *testing.T) {
		if err := os.WriteFile(filepath.Join(tempDir, MetadataFileName), []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		if !MetadataExists(tempDir) {
			t.Error("Expected true for existing metadata")
		}
	})
}

func TestNewMetadataFromConfig(t *testing.T) {
	cfg := &ProjectConfig{
		ProjectName:   "my-project",
		GroupID:       "com.example.myproject",
		ArtifactID:    "my-project",
		JavaVersion:   "21",
		Modules:       []string{"Model", "SQLDatastore", "API"},
		Database:      "postgresql",
		NoSQLDatabase: "",
		MessageBroker: "kafka",
		AIAgents:      []string{"claude"},
	}

	meta := NewMetadataFromConfig(cfg, "2.0.0")

	if meta.Version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", meta.Version)
	}
	if meta.GeneratedAt == "" {
		t.Error("Expected generatedAt to be set")
	}
	if meta.ProjectName != cfg.ProjectName {
		t.Errorf("Expected projectName '%s', got '%s'", cfg.ProjectName, meta.ProjectName)
	}
	if meta.GroupID != cfg.GroupID {
		t.Errorf("Expected groupId '%s', got '%s'", cfg.GroupID, meta.GroupID)
	}
	if meta.Database != cfg.Database {
		t.Errorf("Expected database '%s', got '%s'", cfg.Database, meta.Database)
	}
	if len(meta.Modules) != len(cfg.Modules) {
		t.Errorf("Expected %d modules, got %d", len(cfg.Modules), len(meta.Modules))
	}
}

func TestToProjectConfig(t *testing.T) {
	meta := &ProjectMetadata{
		Version:       "1.0.0",
		GeneratedAt:   "2024-01-01T00:00:00Z",
		ProjectName:   "test-project",
		GroupID:       "com.example.test",
		ArtifactID:    "test-project",
		JavaVersion:   "21",
		Modules:       []string{"Model", "API", "Worker"},
		Database:      "postgresql",
		NoSQLDatabase: "",
		MessageBroker: "kafka",
		AIAgents:      []string{"claude"},
	}

	cfg := meta.ToProjectConfig()

	if cfg.ProjectName != meta.ProjectName {
		t.Errorf("Expected projectName '%s', got '%s'", meta.ProjectName, cfg.ProjectName)
	}
	if cfg.GroupID != meta.GroupID {
		t.Errorf("Expected groupId '%s', got '%s'", meta.GroupID, cfg.GroupID)
	}
	if cfg.ArtifactID != meta.ArtifactID {
		t.Errorf("Expected artifactId '%s', got '%s'", meta.ArtifactID, cfg.ArtifactID)
	}
	if cfg.JavaVersion != meta.JavaVersion {
		t.Errorf("Expected javaVersion '%s', got '%s'", meta.JavaVersion, cfg.JavaVersion)
	}
	if cfg.Database != meta.Database {
		t.Errorf("Expected database '%s', got '%s'", meta.Database, cfg.Database)
	}
	if cfg.MessageBroker != meta.MessageBroker {
		t.Errorf("Expected messageBroker '%s', got '%s'", meta.MessageBroker, cfg.MessageBroker)
	}
	if len(cfg.Modules) != len(meta.Modules) {
		t.Errorf("Expected %d modules, got %d", len(meta.Modules), len(cfg.Modules))
	}
}

func TestMetadataHasModule(t *testing.T) {
	meta := &ProjectMetadata{
		Modules: []string{"Model", "API", "SQLDatastore"},
	}

	tests := []struct {
		module   string
		expected bool
	}{
		{"Model", true},
		{"API", true},
		{"SQLDatastore", true},
		{"Worker", false},
		{"Events", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			if got := meta.HasModule(tt.module); got != tt.expected {
				t.Errorf("HasModule(%s) = %v, want %v", tt.module, got, tt.expected)
			}
		})
	}
}

func TestMetadataAddModule(t *testing.T) {
	meta := &ProjectMetadata{
		Modules: []string{"Model", "API"},
	}

	t.Run("adds new module", func(t *testing.T) {
		meta.AddModule("SQLDatastore")
		if !meta.HasModule("SQLDatastore") {
			t.Error("Expected SQLDatastore to be added")
		}
		if len(meta.Modules) != 3 {
			t.Errorf("Expected 3 modules, got %d", len(meta.Modules))
		}
	})

	t.Run("doesn't duplicate existing module", func(t *testing.T) {
		initialLen := len(meta.Modules)
		meta.AddModule("Model") // Already exists
		if len(meta.Modules) != initialLen {
			t.Errorf("Module count changed from %d to %d", initialLen, len(meta.Modules))
		}
	})
}

func TestMetadataUpdateGeneratedAt(t *testing.T) {
	meta := &ProjectMetadata{
		GeneratedAt: "2020-01-01T00:00:00Z",
	}

	oldTime := meta.GeneratedAt
	meta.UpdateGeneratedAt()

	if meta.GeneratedAt == oldTime {
		t.Error("Expected generatedAt to be updated")
	}

	// Should be a valid RFC3339 timestamp
	if len(meta.GeneratedAt) < 20 {
		t.Errorf("GeneratedAt doesn't look like RFC3339: %s", meta.GeneratedAt)
	}
}

func TestMetadataFileName(t *testing.T) {
	if MetadataFileName != ".trabuco.json" {
		t.Errorf("Expected MetadataFileName to be '.trabuco.json', got '%s'", MetadataFileName)
	}
}
