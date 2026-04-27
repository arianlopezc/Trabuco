// Package state owns the durable migration state stored under
// `.trabuco-migration/` in the user's repo. State is the single source of
// truth that lets the orchestrator (in CLI mode or plugin mode) suspend,
// resume, and reason about progress across phases.
package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// SchemaVersion is the on-disk state.json schema. Bumped when the layout
// changes incompatibly; the loader rejects unknown versions.
const SchemaVersion = 1

// MigrationDir is the working directory inside the user's repo.
const MigrationDir = ".trabuco-migration"

const (
	stateFile = "state.json"
	lockFile  = "lock.json"
)

// State is the full migration state persisted to state.json.
type State struct {
	SchemaVersion     int                                 `json:"schemaVersion"`
	TrabucoCLIVersion string                              `json:"trabucoCliVersion"`
	StartedAt         time.Time                           `json:"startedAt"`
	LastUpdatedAt     time.Time                           `json:"lastUpdatedAt"`
	SourceConfig      SourceConfig                        `json:"sourceConfig"`
	TargetConfig      TargetConfig                        `json:"targetConfig"`
	Phases            map[types.Phase]*PhaseRecord        `json:"phases"`
	Blockers          []BlockerRecord                     `json:"blockers"`
	Decisions         []DecisionRecord                    `json:"decisions"`
	RetainedLegacy    []string                            `json:"retainedLegacy"`
}

// SourceConfig captures what the assessor learned about the source repo.
// Filled by Phase 0; immutable thereafter.
type SourceConfig struct {
	BuildSystem    string   `json:"buildSystem"`    // maven | gradle | other
	Framework      string   `json:"framework"`      // spring-boot-2.x | spring-boot-3.x | quarkus | ...
	JavaVersion    string   `json:"javaVersion"`
	Persistence    string   `json:"persistence"`    // jpa | spring-data-jdbc | mongodb | none | ...
	Messaging      string   `json:"messaging"`      // kafka | rabbitmq | sqs | pubsub | none
	HasAIIntegration bool   `json:"hasAiIntegration"`
	CISystems      []string `json:"ciSystems"`
	TestFramework  string   `json:"testFramework"`
}

// TargetConfig is the user-approved Trabuco shape we are migrating into.
type TargetConfig struct {
	Modules       []string `json:"modules"`
	Database      string   `json:"database"`
	NoSQLDatabase string   `json:"nosqlDatabase,omitempty"`
	MessageBroker string   `json:"messageBroker,omitempty"`
	AIAgents      []string `json:"aiAgents,omitempty"`
	CIProvider    string   `json:"ciProvider,omitempty"`
	JavaVersion   string   `json:"javaVersion"`
}

// PhaseRecord is the runtime state of a single phase.
type PhaseRecord struct {
	State         types.PhaseStateLabel `json:"state"`
	StartedAt     *time.Time            `json:"startedAt,omitempty"`
	ApprovedAt    *time.Time            `json:"approvedAt,omitempty"`
	PreTag        string                `json:"preTag,omitempty"`
	PostTag       string                `json:"postTag,omitempty"`
	Reason        string                `json:"reason,omitempty"`
	SubAggregates map[string]types.PhaseStateLabel `json:"subAggregates,omitempty"`
	RetryCount    int                   `json:"retryCount,omitempty"`
}

// BlockerRecord is a recorded blocker with the user's resolution.
type BlockerRecord struct {
	Phase        types.Phase       `json:"phase"`
	Code         types.BlockerCode `json:"code"`
	File         string            `json:"file"`
	Note         string            `json:"note"`
	Alternatives []string          `json:"alternatives"`
	UserChoice   string            `json:"userChoice,omitempty"`
	RecordedAt   time.Time         `json:"recordedAt"`
}

// DecisionRecord is a user-recorded answer to a requires-decision item.
type DecisionRecord struct {
	ID         string      `json:"id"`
	Phase      types.Phase `json:"phase"`
	Question   string      `json:"question"`
	Choices    []string    `json:"choices"`
	Choice     string      `json:"choice"`
	DecidedAt  time.Time   `json:"decidedAt"`
}

// New creates an empty State seeded with default values.
func New(cliVersion string) *State {
	now := time.Now().UTC()
	phases := make(map[types.Phase]*PhaseRecord, len(types.AllPhases()))
	for _, p := range types.AllPhases() {
		phases[p] = &PhaseRecord{State: types.PhasePending}
	}
	return &State{
		SchemaVersion:     SchemaVersion,
		TrabucoCLIVersion: cliVersion,
		StartedAt:         now,
		LastUpdatedAt:     now,
		Phases:            phases,
		Blockers:          []BlockerRecord{},
		Decisions:         []DecisionRecord{},
		RetainedLegacy:    []string{},
	}
}

// MigrationDirPath returns the path to .trabuco-migration/ inside repoRoot.
func MigrationDirPath(repoRoot string) string {
	return filepath.Join(repoRoot, MigrationDir)
}

// StatePath returns the path to state.json inside repoRoot.
func StatePath(repoRoot string) string {
	return filepath.Join(MigrationDirPath(repoRoot), stateFile)
}

// LockPath returns the path to lock.json inside repoRoot.
func LockPath(repoRoot string) string {
	return filepath.Join(MigrationDirPath(repoRoot), lockFile)
}

// Exists reports whether a migration state directory already exists in
// repoRoot. Used by the orchestrator's pre-Phase-0 check to detect a prior
// aborted run.
func Exists(repoRoot string) bool {
	_, err := os.Stat(StatePath(repoRoot))
	return err == nil
}

// Load reads state.json from repoRoot. Errors if the file doesn't exist or
// the schema version is incompatible.
func Load(repoRoot string) (*State, error) {
	path := StatePath(repoRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read state.json: %w", err)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse state.json: %w", err)
	}
	if s.SchemaVersion != SchemaVersion {
		return nil, fmt.Errorf("state.json schema version %d not supported (this CLI expects %d)", s.SchemaVersion, SchemaVersion)
	}
	if s.Phases == nil {
		s.Phases = make(map[types.Phase]*PhaseRecord)
	}
	return &s, nil
}

// Save writes state to disk atomically (write tmp, fsync, rename).
// Updates LastUpdatedAt.
func Save(repoRoot string, s *State) error {
	if err := os.MkdirAll(MigrationDirPath(repoRoot), 0o755); err != nil {
		return fmt.Errorf("create migration dir: %w", err)
	}
	s.LastUpdatedAt = time.Now().UTC()
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state.json: %w", err)
	}
	path := StatePath(repoRoot)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write state.json.tmp: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("rename state.json.tmp: %w", err)
	}
	return nil
}

// PhaseInputPath returns the path to phase-N-input.json (specialist input).
func PhaseInputPath(repoRoot string, phase types.Phase) string {
	return filepath.Join(MigrationDirPath(repoRoot), fmt.Sprintf("phase-%d-input.json", int(phase)))
}

// PhaseOutputPath returns the path to phase-N-output.json (specialist output).
func PhaseOutputPath(repoRoot string, phase types.Phase) string {
	return filepath.Join(MigrationDirPath(repoRoot), fmt.Sprintf("phase-%d-output.json", int(phase)))
}

// PhaseDiffPath returns the path to phase-N-diff.patch.
func PhaseDiffPath(repoRoot string, phase types.Phase) string {
	return filepath.Join(MigrationDirPath(repoRoot), fmt.Sprintf("phase-%d-diff.patch", int(phase)))
}

// PhaseReportPath returns the path to phase-N-report.md (human summary).
func PhaseReportPath(repoRoot string, phase types.Phase) string {
	return filepath.Join(MigrationDirPath(repoRoot), fmt.Sprintf("phase-%d-report.md", int(phase)))
}

// AssessmentPath returns the path to assessment.json (Phase 0 output, the
// no-out-of-scope contract).
func AssessmentPath(repoRoot string) string {
	return filepath.Join(MigrationDirPath(repoRoot), "assessment.json")
}

// CompletionReportPath returns the path to the final completion-report.md.
func CompletionReportPath(repoRoot string) string {
	return filepath.Join(MigrationDirPath(repoRoot), "completion-report.md")
}
