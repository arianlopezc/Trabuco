package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Stage represents a migration stage
type Stage string

const (
	StageDiscovery        Stage = "discovery"
	StageDependencies     Stage = "dependencies"
	StageEntities         Stage = "entities"
	StageRepositories     Stage = "repositories"
	StageServices         Stage = "services"
	StageControllers      Stage = "controllers"
	StageJobs             Stage = "jobs"
	StageEvents           Stage = "events"
	StageConfiguration    Stage = "configuration"
	StageFinalAssembly    Stage = "final_assembly"
)

// StageOrder defines the order of migration stages
var StageOrder = []Stage{
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

// Status represents the status of a stage or checkpoint
type Status string

const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusSkipped    Status = "skipped"
)

// Checkpoint represents a saved migration state
type Checkpoint struct {
	// Stage is the current/last completed stage
	Stage Stage `json:"stage"`

	// Status of the checkpoint
	Status Status `json:"status"`

	// Timestamp when the checkpoint was created
	Timestamp time.Time `json:"timestamp"`

	// SourcePath is the original source project path
	SourcePath string `json:"source_path"`

	// OutputPath is the target output path
	OutputPath string `json:"output_path"`

	// Data contains stage-specific data
	Data map[string]interface{} `json:"data"`

	// AIDecisions tracks all AI decisions made
	AIDecisions []AIDecision `json:"ai_decisions"`

	// Errors encountered during this stage
	Errors []string `json:"errors,omitempty"`

	// TokensUsed tracks cumulative token usage
	TokensUsed TokenUsage `json:"tokens_used"`

	// EstimatedCost tracks cumulative cost
	EstimatedCost float64 `json:"estimated_cost"`
}

// AIDecision records an AI analysis decision
type AIDecision struct {
	// Stage where the decision was made
	Stage Stage `json:"stage"`

	// Timestamp of the decision
	Timestamp time.Time `json:"timestamp"`

	// File being analyzed
	File string `json:"file"`

	// Action taken (create, modify, delete, skip)
	Action string `json:"action"`

	// Prompt sent to AI (truncated for storage)
	PromptSummary string `json:"prompt_summary"`

	// Response summary (truncated)
	ResponseSummary string `json:"response_summary"`

	// TokensUsed for this decision
	TokensUsed TokenUsage `json:"tokens_used"`

	// RequiredReview if human review was needed
	RequiredReview bool `json:"required_review"`
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// CheckpointManager handles checkpoint persistence
type CheckpointManager struct {
	checkpointDir string
	current       *Checkpoint
}

// NewCheckpointManager creates a new checkpoint manager
func NewCheckpointManager(sourcePath string) *CheckpointManager {
	return &CheckpointManager{
		checkpointDir: GetCheckpointDir(sourcePath),
	}
}

// GetCheckpointDir returns the checkpoint directory for a source project
func GetCheckpointDir(sourcePath string) string {
	return filepath.Join(sourcePath, ".trabuco-migrate")
}

// Initialize creates the checkpoint directory and initial checkpoint
func (cm *CheckpointManager) Initialize(sourcePath, outputPath string) error {
	// Create checkpoint directory
	if err := os.MkdirAll(cm.checkpointDir, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	// Create initial checkpoint
	cm.current = &Checkpoint{
		Stage:      StageDiscovery,
		Status:     StatusPending,
		Timestamp:  time.Now(),
		SourcePath: sourcePath,
		OutputPath: outputPath,
		Data:       make(map[string]interface{}),
	}

	return cm.Save()
}

// Load loads the latest checkpoint
func (cm *CheckpointManager) Load() (*Checkpoint, error) {
	checkpointFile := filepath.Join(cm.checkpointDir, "checkpoint.json")

	data, err := os.ReadFile(checkpointFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No checkpoint exists
		}
		return nil, fmt.Errorf("failed to read checkpoint: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	cm.current = &checkpoint
	return &checkpoint, nil
}

// Save persists the current checkpoint
func (cm *CheckpointManager) Save() error {
	if cm.current == nil {
		return fmt.Errorf("no checkpoint to save")
	}

	cm.current.Timestamp = time.Now()

	data, err := json.MarshalIndent(cm.current, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	// Ensure checkpoint directory exists
	if err := os.MkdirAll(cm.checkpointDir, 0755); err != nil {
		return fmt.Errorf("failed to create checkpoint directory: %w", err)
	}

	checkpointFile := filepath.Join(cm.checkpointDir, "checkpoint.json")
	if err := os.WriteFile(checkpointFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}

	// Also save stage-specific checkpoint
	stageFile := filepath.Join(cm.checkpointDir, fmt.Sprintf("%s.json", cm.current.Stage))
	if err := os.WriteFile(stageFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write stage checkpoint: %w", err)
	}

	return nil
}

// Current returns the current checkpoint
func (cm *CheckpointManager) Current() *Checkpoint {
	return cm.current
}

// StartStage marks a stage as in progress
func (cm *CheckpointManager) StartStage(stage Stage) error {
	if cm.current == nil {
		cm.current = &Checkpoint{
			Data: make(map[string]interface{}),
		}
	}

	cm.current.Stage = stage
	cm.current.Status = StatusInProgress
	cm.current.Errors = nil

	return cm.Save()
}

// CompleteStage marks a stage as completed
func (cm *CheckpointManager) CompleteStage(stage Stage, data map[string]interface{}) error {
	if cm.current == nil {
		return fmt.Errorf("no checkpoint initialized")
	}

	cm.current.Stage = stage
	cm.current.Status = StatusCompleted

	// Merge data
	for k, v := range data {
		cm.current.Data[k] = v
	}

	return cm.Save()
}

// FailStage marks a stage as failed
func (cm *CheckpointManager) FailStage(stage Stage, err error) error {
	if cm.current == nil {
		return fmt.Errorf("no checkpoint initialized")
	}

	cm.current.Stage = stage
	cm.current.Status = StatusFailed
	cm.current.Errors = append(cm.current.Errors, err.Error())

	return cm.Save()
}

// AddAIDecision records an AI decision
func (cm *CheckpointManager) AddAIDecision(decision AIDecision) {
	if cm.current == nil {
		return
	}

	decision.Timestamp = time.Now()
	cm.current.AIDecisions = append(cm.current.AIDecisions, decision)

	// Update token usage
	cm.current.TokensUsed.InputTokens += decision.TokensUsed.InputTokens
	cm.current.TokensUsed.OutputTokens += decision.TokensUsed.OutputTokens
}

// UpdateCost updates the estimated cost
func (cm *CheckpointManager) UpdateCost(cost float64) {
	if cm.current != nil {
		cm.current.EstimatedCost += cost
	}
}

// GetLastCompletedStage returns the last successfully completed stage
func (cm *CheckpointManager) GetLastCompletedStage() Stage {
	if cm.current == nil || cm.current.Status != StatusCompleted {
		return ""
	}
	return cm.current.Stage
}

// GetNextStage returns the next stage to execute
func (cm *CheckpointManager) GetNextStage() Stage {
	lastCompleted := cm.GetLastCompletedStage()

	if lastCompleted == "" {
		return StageDiscovery
	}

	for i, stage := range StageOrder {
		if stage == lastCompleted && i < len(StageOrder)-1 {
			return StageOrder[i+1]
		}
	}

	return "" // All stages completed
}

// HasCheckpoint returns true if a checkpoint exists
func (cm *CheckpointManager) HasCheckpoint() bool {
	checkpointFile := filepath.Join(cm.checkpointDir, "checkpoint.json")
	_, err := os.Stat(checkpointFile)
	return err == nil
}

// Cleanup removes the checkpoint directory
func (cm *CheckpointManager) Cleanup() error {
	return os.RemoveAll(cm.checkpointDir)
}

// RollbackToStage rolls back to a specific stage
func RollbackToStage(checkpointDir string, stageName string) error {
	stage := Stage(stageName)

	// Load the stage-specific checkpoint
	stageFile := filepath.Join(checkpointDir, fmt.Sprintf("%s.json", stage))
	data, err := os.ReadFile(stageFile)
	if err != nil {
		return fmt.Errorf("no checkpoint found for stage %s: %w", stage, err)
	}

	// Restore as current checkpoint
	checkpointFile := filepath.Join(checkpointDir, "checkpoint.json")
	if err := os.WriteFile(checkpointFile, data, 0644); err != nil {
		return fmt.Errorf("failed to restore checkpoint: %w", err)
	}

	return nil
}

// RollbackAll removes all migration artifacts
func RollbackAll(checkpointDir string) error {
	// Load checkpoint to get output path
	checkpointFile := filepath.Join(checkpointDir, "checkpoint.json")
	data, err := os.ReadFile(checkpointFile)
	if err != nil {
		return fmt.Errorf("no checkpoint found: %w", err)
	}

	var checkpoint Checkpoint
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return fmt.Errorf("failed to parse checkpoint: %w", err)
	}

	// Remove output directory
	if checkpoint.OutputPath != "" {
		if err := os.RemoveAll(checkpoint.OutputPath); err != nil {
			return fmt.Errorf("failed to remove output directory: %w", err)
		}
	}

	// Remove checkpoint directory
	return os.RemoveAll(checkpointDir)
}
