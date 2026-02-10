package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

func TestModuleAdderValidateCanAdd(t *testing.T) {
	metadata := &config.ProjectMetadata{
		ProjectName: "test-project",
		GroupID:     "com.example.test",
		Modules:     []string{"Model", "API"},
	}

	adder := NewModuleAdder("/tmp/test", metadata, "1.0.0", false)

	tests := []struct {
		name      string
		module    string
		wantError bool
	}{
		{"can add SQLDatastore", "SQLDatastore", false},
		{"can add Worker", "Worker", false},
		{"cannot add existing module", "API", true},
		{"cannot add unknown module", "UnknownModule", true},
		{"cannot add internal module", "Jobs", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := adder.ValidateCanAdd(tt.module)
			if tt.wantError && err == nil {
				t.Error("Expected error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestModuleAdderMutualExclusion(t *testing.T) {
	t.Run("cannot add NoSQLDatastore when SQLDatastore exists", func(t *testing.T) {
		metadata := &config.ProjectMetadata{
			ProjectName: "test-project",
			GroupID:     "com.example.test",
			Modules:     []string{"Model", "SQLDatastore"},
		}

		adder := NewModuleAdder("/tmp/test", metadata, "1.0.0", false)
		err := adder.ValidateCanAdd("NoSQLDatastore")
		if err == nil {
			t.Error("Expected error for mutual exclusion")
		}
		if !strings.Contains(err.Error(), "mutually exclusive") {
			t.Errorf("Expected mutual exclusion error, got: %v", err)
		}
	})

	t.Run("cannot add SQLDatastore when NoSQLDatastore exists", func(t *testing.T) {
		metadata := &config.ProjectMetadata{
			ProjectName: "test-project",
			GroupID:     "com.example.test",
			Modules:     []string{"Model", "NoSQLDatastore"},
		}

		adder := NewModuleAdder("/tmp/test", metadata, "1.0.0", false)
		err := adder.ValidateCanAdd("SQLDatastore")
		if err == nil {
			t.Error("Expected error for mutual exclusion")
		}
		if !strings.Contains(err.Error(), "mutually exclusive") {
			t.Errorf("Expected mutual exclusion error, got: %v", err)
		}
	})
}

func TestModuleAdderResolveDependencies(t *testing.T) {
	tests := []struct {
		name            string
		existingModules []string
		moduleToAdd     string
		expectedDeps    []string
	}{
		{
			name:            "Worker adds Jobs",
			existingModules: []string{"Model"},
			moduleToAdd:     "Worker",
			expectedDeps:    []string{"Jobs"},
		},
		{
			name:            "EventConsumer adds Events",
			existingModules: []string{"Model"},
			moduleToAdd:     "EventConsumer",
			expectedDeps:    []string{"Events"},
		},
		{
			name:            "Worker with Jobs already exists",
			existingModules: []string{"Model", "Jobs"},
			moduleToAdd:     "Worker",
			expectedDeps:    []string{},
		},
		{
			name:            "SQLDatastore adds nothing",
			existingModules: []string{"Model"},
			moduleToAdd:     "SQLDatastore",
			expectedDeps:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := &config.ProjectMetadata{
				ProjectName: "test-project",
				GroupID:     "com.example.test",
				Modules:     tt.existingModules,
			}

			adder := NewModuleAdder("/tmp/test", metadata, "1.0.0", false)
			deps := adder.ResolveDependencies(tt.moduleToAdd)

			if len(deps) != len(tt.expectedDeps) {
				t.Errorf("Expected %d dependencies, got %d", len(tt.expectedDeps), len(deps))
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

func TestModuleAdderDryRun(t *testing.T) {
	metadata := &config.ProjectMetadata{
		ProjectName: "test-project",
		GroupID:     "com.example.test",
		ArtifactID:  "test-project",
		Modules:     []string{"Model", "API"},
	}

	adder := NewModuleAdder("/tmp/test", metadata, "1.0.0", false)

	t.Run("dry run SQLDatastore", func(t *testing.T) {
		result := adder.DryRun("SQLDatastore")

		if result.Module != "SQLDatastore" {
			t.Errorf("Expected module 'SQLDatastore', got '%s'", result.Module)
		}

		// Should have pom.xml in modified files
		hasPOM := false
		for _, f := range result.FilesModified {
			if f == "pom.xml" {
				hasPOM = true
				break
			}
		}
		if !hasPOM {
			t.Error("Expected pom.xml in modified files")
		}

		// Should have SQLDatastore files in created files
		hasSQLFiles := false
		for _, f := range result.FilesCreated {
			if strings.HasPrefix(f, "SQLDatastore") {
				hasSQLFiles = true
				break
			}
		}
		if !hasSQLFiles {
			t.Error("Expected SQLDatastore files in created files")
		}
	})

	t.Run("dry run Worker", func(t *testing.T) {
		result := adder.DryRun("Worker")

		if result.Module != "Worker" {
			t.Errorf("Expected module 'Worker', got '%s'", result.Module)
		}

		// Should include Jobs as dependency
		hasJobs := false
		for _, dep := range result.Dependencies {
			if dep == "Jobs" {
				hasJobs = true
				break
			}
		}
		if !hasJobs {
			t.Error("Expected Jobs in dependencies")
		}
	})
}

func TestBackupManager(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "backup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testContent := "original content"
	testFile := "test.txt"
	if err := os.WriteFile(filepath.Join(tempDir, testFile), []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("backup and restore", func(t *testing.T) {
		backup := NewBackupManager(tempDir, true)

		// Backup the file
		if err := backup.Backup(testFile); err != nil {
			t.Fatalf("Backup failed: %v", err)
		}

		// Verify backup was created
		if !backup.HasBackups() {
			t.Error("Expected HasBackups() to be true")
		}

		// Modify the original file
		newContent := "modified content"
		if err := os.WriteFile(filepath.Join(tempDir, testFile), []byte(newContent), 0644); err != nil {
			t.Fatalf("Failed to modify test file: %v", err)
		}

		// Restore
		if err := backup.Restore(); err != nil {
			t.Fatalf("Restore failed: %v", err)
		}

		// Verify content was restored
		content, err := os.ReadFile(filepath.Join(tempDir, testFile))
		if err != nil {
			t.Fatalf("Failed to read restored file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("Expected '%s', got '%s'", testContent, string(content))
		}
	})

	t.Run("disabled backup", func(t *testing.T) {
		backup := NewBackupManager(tempDir, false)

		if err := backup.Backup(testFile); err != nil {
			t.Errorf("Backup should succeed when disabled: %v", err)
		}

		if backup.HasBackups() {
			t.Error("Expected no backups when disabled")
		}
	})

	t.Run("backup nonexistent file", func(t *testing.T) {
		backup := NewBackupManager(tempDir, true)

		if err := backup.Backup("nonexistent.txt"); err != nil {
			t.Errorf("Backup of nonexistent file should succeed: %v", err)
		}

		if backup.HasBackups() {
			t.Error("Expected no backups for nonexistent file")
		}
	})

	t.Run("cleanup removes backup and parent directory if empty", func(t *testing.T) {
		// Create a fresh temp dir for this test
		cleanupDir, err := os.MkdirTemp("", "cleanup-test-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(cleanupDir)

		// Create a test file to backup
		if err := os.WriteFile(filepath.Join(cleanupDir, "test.txt"), []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		backup := NewBackupManager(cleanupDir, true)

		// Create a backup
		if err := backup.Backup("test.txt"); err != nil {
			t.Fatalf("Backup failed: %v", err)
		}

		// Verify backup directory exists
		backupRoot := filepath.Join(cleanupDir, BackupDirName)
		if _, err := os.Stat(backupRoot); os.IsNotExist(err) {
			t.Fatal("Backup directory should exist after backup")
		}

		// Cleanup
		if err := backup.Cleanup(); err != nil {
			t.Fatalf("Cleanup failed: %v", err)
		}

		// Verify .trabuco-backup directory is removed (since it's now empty)
		if _, err := os.Stat(backupRoot); !os.IsNotExist(err) {
			t.Error("Backup root directory should be removed after cleanup")
		}
	})
}

func TestPOMUpdater(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pom-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
    <modules>
        <module>Model</module>
        <module>API</module>
    </modules>
    <properties>
        <java.version>21</java.version>
    </properties>
    <dependencyManagement>
        <dependencies>
        </dependencies>
    </dependencyManagement>
</project>`

	pomPath := filepath.Join(tempDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte(pomContent), 0644); err != nil {
		t.Fatalf("Failed to write pom.xml: %v", err)
	}

	t.Run("add module", func(t *testing.T) {
		updater, err := NewPOMUpdater(pomPath)
		if err != nil {
			t.Fatalf("Failed to create updater: %v", err)
		}

		if err := updater.AddModule("SQLDatastore"); err != nil {
			t.Fatalf("Failed to add module: %v", err)
		}

		if err := updater.Save(); err != nil {
			t.Fatalf("Failed to save: %v", err)
		}

		// Read back and verify
		content, err := os.ReadFile(pomPath)
		if err != nil {
			t.Fatalf("Failed to read pom: %v", err)
		}

		if !strings.Contains(string(content), "<module>SQLDatastore</module>") {
			t.Error("Module not found in POM")
		}
	})

	t.Run("add property", func(t *testing.T) {
		updater, err := NewPOMUpdater(pomPath)
		if err != nil {
			t.Fatalf("Failed to create updater: %v", err)
		}

		if err := updater.AddProperty("new.property", "value"); err != nil {
			t.Fatalf("Failed to add property: %v", err)
		}

		if err := updater.Save(); err != nil {
			t.Fatalf("Failed to save: %v", err)
		}

		content, err := os.ReadFile(pomPath)
		if err != nil {
			t.Fatalf("Failed to read pom: %v", err)
		}

		if !strings.Contains(string(content), "<new.property>value</new.property>") {
			t.Error("Property not found in POM")
		}
	})

	t.Run("add existing module is no-op", func(t *testing.T) {
		updater, err := NewPOMUpdater(pomPath)
		if err != nil {
			t.Fatalf("Failed to create updater: %v", err)
		}

		if err := updater.AddModule("Model"); err != nil {
			t.Fatalf("Adding existing module should not fail: %v", err)
		}
	})
}

func TestGetFilesToBackup(t *testing.T) {
	tests := []struct {
		module   string
		expected []string
	}{
		{
			module:   "SQLDatastore",
			expected: []string{"pom.xml", ".trabuco.json", "docker-compose.yml", ".env.example", "README.md", "CLAUDE.md", "Model/pom.xml"},
		},
		{
			module:   "API",
			expected: []string{"pom.xml", ".trabuco.json", "README.md", "CLAUDE.md"},
		},
		{
			module:   "Worker",
			expected: []string{"pom.xml", ".trabuco.json", "docker-compose.yml", "README.md", "Model/pom.xml"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.module, func(t *testing.T) {
			files := GetFilesToBackup(tt.module)

			for _, expected := range tt.expected {
				found := false
				for _, f := range files {
					if f == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected %s in files to backup", expected)
				}
			}
		})
	}
}
