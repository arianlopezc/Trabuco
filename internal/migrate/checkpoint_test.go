package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCheckpointManager_Initialize(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	err := cm.Initialize(tempDir, filepath.Join(tempDir, "output"))

	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Check checkpoint directory exists
	checkpointDir := GetCheckpointDir(tempDir)
	if _, err := os.Stat(checkpointDir); os.IsNotExist(err) {
		t.Error("checkpoint directory was not created")
	}

	// Check current checkpoint is set
	if cm.Current() == nil {
		t.Error("current checkpoint is nil after initialization")
	}

	if cm.Current().SourcePath != tempDir {
		t.Errorf("SourcePath = %v, want %v", cm.Current().SourcePath, tempDir)
	}

	if cm.Current().Stage != StageDiscovery {
		t.Errorf("Stage = %v, want %v", cm.Current().Stage, StageDiscovery)
	}
}

func TestCheckpointManager_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()

	// Initialize and save
	cm := NewCheckpointManager(tempDir)
	err := cm.Initialize(tempDir, filepath.Join(tempDir, "output"))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Modify checkpoint
	cm.current.Stage = StageEntities
	cm.current.Status = StatusCompleted
	cm.current.Data["test_key"] = "test_value"

	err = cm.Save()
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Create new manager and load
	cm2 := NewCheckpointManager(tempDir)
	checkpoint, err := cm2.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if checkpoint == nil {
		t.Fatal("loaded checkpoint is nil")
	}

	if checkpoint.Stage != StageEntities {
		t.Errorf("Stage = %v, want %v", checkpoint.Stage, StageEntities)
	}

	if checkpoint.Status != StatusCompleted {
		t.Errorf("Status = %v, want %v", checkpoint.Status, StatusCompleted)
	}

	if checkpoint.Data["test_key"] != "test_value" {
		t.Errorf("Data[test_key] = %v, want test_value", checkpoint.Data["test_key"])
	}
}

func TestCheckpointManager_StageProgress(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	err := cm.Initialize(tempDir, filepath.Join(tempDir, "output"))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	// Start stage
	err = cm.StartStage(StageEntities)
	if err != nil {
		t.Fatalf("StartStage() error = %v", err)
	}

	if cm.Current().Stage != StageEntities {
		t.Errorf("Stage = %v, want %v", cm.Current().Stage, StageEntities)
	}

	if cm.Current().Status != StatusInProgress {
		t.Errorf("Status = %v, want %v", cm.Current().Status, StatusInProgress)
	}

	// Complete stage
	data := map[string]interface{}{"entities_count": 5}
	err = cm.CompleteStage(StageEntities, data)
	if err != nil {
		t.Fatalf("CompleteStage() error = %v", err)
	}

	if cm.Current().Status != StatusCompleted {
		t.Errorf("Status = %v, want %v", cm.Current().Status, StatusCompleted)
	}

	if cm.Current().Data["entities_count"] != 5 {
		t.Errorf("Data[entities_count] = %v, want 5", cm.Current().Data["entities_count"])
	}
}

func TestCheckpointManager_FailStage(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	err := cm.Initialize(tempDir, filepath.Join(tempDir, "output"))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	cm.StartStage(StageEntities)

	// Fail stage
	testErr := os.ErrNotExist // Using a standard error for testing
	err = cm.FailStage(StageEntities, testErr)
	if err != nil {
		t.Fatalf("FailStage() error = %v", err)
	}

	if cm.Current().Status != StatusFailed {
		t.Errorf("Status = %v, want %v", cm.Current().Status, StatusFailed)
	}

	if len(cm.Current().Errors) != 1 {
		t.Errorf("len(Errors) = %v, want 1", len(cm.Current().Errors))
	}
}

func TestCheckpointManager_AddAIDecision(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	err := cm.Initialize(tempDir, filepath.Join(tempDir, "output"))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	decision := AIDecision{
		Stage:           StageEntities,
		File:            "/path/to/User.java",
		Action:          "convert",
		PromptSummary:   "Convert User entity",
		ResponseSummary: "Generated User class...",
		TokensUsed: TokenUsage{
			InputTokens:  500,
			OutputTokens: 1000,
		},
	}

	cm.AddAIDecision(decision)

	if len(cm.Current().AIDecisions) != 1 {
		t.Errorf("len(AIDecisions) = %v, want 1", len(cm.Current().AIDecisions))
	}

	if cm.Current().TokensUsed.InputTokens != 500 {
		t.Errorf("TokensUsed.InputTokens = %v, want 500", cm.Current().TokensUsed.InputTokens)
	}

	if cm.Current().TokensUsed.OutputTokens != 1000 {
		t.Errorf("TokensUsed.OutputTokens = %v, want 1000", cm.Current().TokensUsed.OutputTokens)
	}

	// Add another decision
	decision2 := AIDecision{
		Stage: StageEntities,
		File:  "/path/to/Order.java",
		TokensUsed: TokenUsage{
			InputTokens:  300,
			OutputTokens: 600,
		},
	}

	cm.AddAIDecision(decision2)

	// Should accumulate tokens
	if cm.Current().TokensUsed.InputTokens != 800 {
		t.Errorf("TokensUsed.InputTokens = %v, want 800", cm.Current().TokensUsed.InputTokens)
	}

	if cm.Current().TokensUsed.OutputTokens != 1600 {
		t.Errorf("TokensUsed.OutputTokens = %v, want 1600", cm.Current().TokensUsed.OutputTokens)
	}
}

func TestCheckpointManager_UpdateCost(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	err := cm.Initialize(tempDir, filepath.Join(tempDir, "output"))
	if err != nil {
		t.Fatalf("Initialize() error = %v", err)
	}

	cm.UpdateCost(0.05)
	cm.UpdateCost(0.03)

	if cm.Current().EstimatedCost != 0.08 {
		t.Errorf("EstimatedCost = %v, want 0.08", cm.Current().EstimatedCost)
	}
}

func TestCheckpointManager_HasCheckpoint(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)

	// Before initialization
	if cm.HasCheckpoint() {
		t.Error("HasCheckpoint() should be false before initialization")
	}

	// After initialization
	cm.Initialize(tempDir, filepath.Join(tempDir, "output"))

	if !cm.HasCheckpoint() {
		t.Error("HasCheckpoint() should be true after initialization")
	}
}

func TestCheckpointManager_Cleanup(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	cm.Initialize(tempDir, filepath.Join(tempDir, "output"))

	// Verify checkpoint exists
	if !cm.HasCheckpoint() {
		t.Fatal("checkpoint should exist after initialization")
	}

	// Cleanup
	err := cm.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	// Verify checkpoint is gone
	checkpointDir := GetCheckpointDir(tempDir)
	if _, err := os.Stat(checkpointDir); !os.IsNotExist(err) {
		t.Error("checkpoint directory should be removed after cleanup")
	}
}

func TestCheckpointManager_GetNextStage(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	cm.Initialize(tempDir, filepath.Join(tempDir, "output"))

	// Initially should return Discovery (first stage)
	if cm.GetNextStage() != StageDiscovery {
		t.Errorf("GetNextStage() = %v, want %v", cm.GetNextStage(), StageDiscovery)
	}

	// After completing Discovery
	cm.StartStage(StageDiscovery)
	cm.CompleteStage(StageDiscovery, nil)

	// Should return Dependencies (second stage)
	if cm.GetNextStage() != StageDependencies {
		t.Errorf("GetNextStage() = %v, want %v", cm.GetNextStage(), StageDependencies)
	}
}

func TestCheckpointManager_GetLastCompletedStage(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	cm.Initialize(tempDir, filepath.Join(tempDir, "output"))

	// Initially should be empty
	if cm.GetLastCompletedStage() != "" {
		t.Errorf("GetLastCompletedStage() = %v, want empty", cm.GetLastCompletedStage())
	}

	// After completing a stage
	cm.StartStage(StageDiscovery)
	cm.CompleteStage(StageDiscovery, nil)

	if cm.GetLastCompletedStage() != StageDiscovery {
		t.Errorf("GetLastCompletedStage() = %v, want %v", cm.GetLastCompletedStage(), StageDiscovery)
	}
}

func TestRollbackToStage(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	cm.Initialize(tempDir, filepath.Join(tempDir, "output"))

	// Complete first two stages
	cm.StartStage(StageDiscovery)
	cm.CompleteStage(StageDiscovery, map[string]interface{}{"discovery": true})

	cm.StartStage(StageDependencies)
	cm.CompleteStage(StageDependencies, map[string]interface{}{"dependencies": true})

	// Now rollback to Discovery
	checkpointDir := GetCheckpointDir(tempDir)
	err := RollbackToStage(checkpointDir, string(StageDiscovery))
	if err != nil {
		t.Fatalf("RollbackToStage() error = %v", err)
	}

	// Load and verify
	cm2 := NewCheckpointManager(tempDir)
	checkpoint, err := cm2.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if checkpoint.Stage != StageDiscovery {
		t.Errorf("Stage = %v, want %v after rollback", checkpoint.Stage, StageDiscovery)
	}
}

func TestStageOrder(t *testing.T) {
	expectedOrder := []Stage{
		StageDiscovery,
		StageDependencies,
		StageEntities,
		StageRepositories,
		StageServices,
		StageControllers,
		StageJobs,
		StageEvents,
		StageConfiguration,
		StageFinalAssembly,
	}

	if len(StageOrder) != len(expectedOrder) {
		t.Errorf("len(StageOrder) = %v, want %v", len(StageOrder), len(expectedOrder))
	}

	for i, stage := range expectedOrder {
		if StageOrder[i] != stage {
			t.Errorf("StageOrder[%d] = %v, want %v", i, StageOrder[i], stage)
		}
	}
}

func TestCheckpointManager_LoadNonExistent(t *testing.T) {
	tempDir := t.TempDir()

	cm := NewCheckpointManager(tempDir)
	checkpoint, err := cm.Load()

	if err != nil {
		t.Fatalf("Load() should not error for non-existent checkpoint, got %v", err)
	}

	if checkpoint != nil {
		t.Error("Load() should return nil for non-existent checkpoint")
	}
}
