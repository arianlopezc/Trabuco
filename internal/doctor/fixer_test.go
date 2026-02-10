package doctor

import (
	"os"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

func TestFixer(t *testing.T) {
	t.Run("fix metadata exists check", func(t *testing.T) {
		tempDir := createTrabucoProjectWithoutMetadata(t)
		defer os.RemoveAll(tempDir)

		// Run doctor - should find missing metadata warning
		doc := New(tempDir, "1.0.0")
		result, err := doc.Run()
		if err != nil {
			t.Fatalf("Doctor run failed: %v", err)
		}

		// Find the metadata exists check
		var metadataCheck *CheckResult
		for i, check := range result.Checks {
			if check.ID == "METADATA_EXISTS" {
				metadataCheck = &result.Checks[i]
				break
			}
		}

		if metadataCheck == nil {
			t.Fatal("Could not find METADATA_EXISTS check")
		}

		if metadataCheck.Status != SeverityWarn {
			t.Errorf("Expected WARN status, got %s", metadataCheck.Status)
		}

		// Fix it
		fixer := NewFixer(tempDir, nil)
		fixResult := fixer.Fix(*metadataCheck)

		if !fixResult.Success {
			t.Errorf("Fix failed: %s", fixResult.Error)
		}

		// Verify metadata now exists
		if !config.MetadataExists(tempDir) {
			t.Error("Metadata should exist after fix")
		}

		// Run doctor again - should pass now
		result2, err := doc.Run()
		if err != nil {
			t.Fatalf("Second doctor run failed: %v", err)
		}

		for _, check := range result2.Checks {
			if check.ID == "METADATA_EXISTS" && check.Status != SeverityPass {
				t.Errorf("Expected PASS after fix, got %s", check.Status)
			}
		}
	})
}

func TestFixerFixAll(t *testing.T) {
	tempDir := createTrabucoProjectWithoutMetadata(t)
	defer os.RemoveAll(tempDir)

	// Run doctor
	doc := New(tempDir, "1.0.0")
	result, err := doc.Run()
	if err != nil {
		t.Fatalf("Doctor run failed: %v", err)
	}

	// Get fixable checks
	fixable := result.GetFixableChecks()
	if len(fixable) == 0 {
		t.Skip("No fixable checks found")
	}

	// Fix all
	fixer := NewFixer(tempDir, nil)
	fixResults := fixer.FixAll(result)

	// Verify we got results for all fixable checks
	if len(fixResults) != len(fixable) {
		t.Errorf("Expected %d fix results, got %d", len(fixable), len(fixResults))
	}

	// Count successes
	successCount := 0
	for _, r := range fixResults {
		if r.Success {
			successCount++
		}
	}

	if successCount == 0 {
		t.Error("Expected at least one successful fix")
	}
}

func TestRegenerateMetadata(t *testing.T) {
	tempDir := createTrabucoProjectWithoutMetadata(t)
	defer os.RemoveAll(tempDir)

	// Verify no metadata
	if config.MetadataExists(tempDir) {
		t.Fatal("Metadata should not exist initially")
	}

	// Regenerate
	metadata, err := RegenerateMetadata(tempDir)
	if err != nil {
		t.Fatalf("RegenerateMetadata failed: %v", err)
	}

	// Verify metadata was created
	if !config.MetadataExists(tempDir) {
		t.Error("Metadata should exist after regeneration")
	}

	// Verify content
	if metadata.GroupID != "com.example.test" {
		t.Errorf("Expected groupId 'com.example.test', got '%s'", metadata.GroupID)
	}
}

func TestSyncMetadataWithPOM(t *testing.T) {
	tempDir := createTestTrabucoProject(t)
	defer os.RemoveAll(tempDir)

	// Load current metadata
	metadata, err := config.LoadMetadata(tempDir)
	if err != nil {
		t.Fatalf("Failed to load metadata: %v", err)
	}

	// Modify metadata to be out of sync
	metadata.Modules = []string{"Model"} // Missing API

	// Sync
	if err := SyncMetadataWithPOM(tempDir, metadata); err != nil {
		t.Fatalf("SyncMetadataWithPOM failed: %v", err)
	}

	// Reload and verify
	reloaded, err := config.LoadMetadata(tempDir)
	if err != nil {
		t.Fatalf("Failed to reload metadata: %v", err)
	}

	if len(reloaded.Modules) != 2 {
		t.Errorf("Expected 2 modules after sync, got %d", len(reloaded.Modules))
	}
}

func TestDoctorRunAndFix(t *testing.T) {
	tempDir := createTrabucoProjectWithoutMetadata(t)
	defer os.RemoveAll(tempDir)

	doc := New(tempDir, "1.0.0")

	// Run and fix
	result, fixResults, err := doc.RunAndFix()
	if err != nil {
		t.Fatalf("RunAndFix failed: %v", err)
	}

	// Should have some fix results (at least metadata regeneration)
	if len(fixResults) == 0 {
		t.Log("No fixes needed - project might already be healthy")
	}

	// Result should be from the second run (after fixes)
	// All fixable issues should be resolved
	for _, check := range result.Checks {
		if check.CanAutoFix && check.Status != SeverityPass {
			t.Errorf("Check %s should be fixed but has status %s", check.ID, check.Status)
		}
	}
}

func TestDoctorRunCategory(t *testing.T) {
	tempDir := createTestTrabucoProject(t)
	defer os.RemoveAll(tempDir)

	doc := New(tempDir, "1.0.0")

	t.Run("structure category", func(t *testing.T) {
		result, err := doc.RunCategory("structure")
		if err != nil {
			t.Fatalf("RunCategory failed: %v", err)
		}

		// Should only have structure checks
		for _, check := range result.Checks {
			// Find the checker to get its category
			for _, checker := range GetAllChecks() {
				if checker.ID() == check.ID {
					if checker.Category() != "structure" {
						t.Errorf("Expected structure category for check %s", check.ID)
					}
					break
				}
			}
		}
	})

	t.Run("metadata category", func(t *testing.T) {
		result, err := doc.RunCategory("metadata")
		if err != nil {
			t.Fatalf("RunCategory failed: %v", err)
		}

		if len(result.Checks) == 0 {
			t.Error("Expected at least one metadata check")
		}
	})

	t.Run("invalid category runs all checks", func(t *testing.T) {
		result, err := doc.RunCategory("invalid")
		if err != nil {
			t.Fatalf("RunCategory failed: %v", err)
		}

		// Should run all checks for invalid category
		allChecks := GetAllChecks()
		if len(result.Checks) != len(allChecks) {
			t.Errorf("Expected %d checks for invalid category, got %d", len(allChecks), len(result.Checks))
		}
	})
}

func TestDoctorNewWithChecks(t *testing.T) {
	tempDir := createTestTrabucoProject(t)
	defer os.RemoveAll(tempDir)

	// Create doctor with only one check
	checks := []Checker{NewProjectStructureCheck()}
	doc := NewWithChecks(tempDir, "1.0.0", checks)

	result, err := doc.Run()
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	if len(result.Checks) != 1 {
		t.Errorf("Expected 1 check, got %d", len(result.Checks))
	}

	if result.Checks[0].ID != "PROJECT_STRUCTURE" {
		t.Errorf("Expected PROJECT_STRUCTURE check, got %s", result.Checks[0].ID)
	}
}

func TestValidationError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "broken-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	doc := New(tempDir, "1.0.0")
	err = doc.Validate()

	if err == nil {
		t.Fatal("Expected validation error")
	}

	validationErr, ok := err.(*ValidationError)
	if !ok {
		t.Fatalf("Expected ValidationError, got %T", err)
	}

	if validationErr.Result == nil {
		t.Error("ValidationError should have Result")
	}

	if validationErr.Error() != "project validation failed" {
		t.Errorf("Unexpected error message: %s", validationErr.Error())
	}
}
