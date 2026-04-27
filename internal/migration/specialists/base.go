// Package specialists defines the contract every migration specialist
// implements. Each phase has one specialist (the test specialist runs
// cross-cutting). Specialists are stateless: their context is the
// assessment artifact, target config, and phase-specific input.
//
// Implementation note: in 1.10.0 specialists are LLM agents driven by the
// orchestrator. The Specialist interface lets us treat them uniformly in
// CLI mode (where the orchestrator is a Go state machine that calls the
// Anthropic API) and plugin mode (where the orchestrator subagent calls
// the same Go handlers via MCP tools).
package specialists

import (
	"context"

	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// Input is what the orchestrator hands to a specialist when invoking it.
// All specialists receive the same shape; phase-specific data lives in
// the assessment artifact and target config.
type Input struct {
	RepoRoot     string         `json:"repoRoot"`
	Phase        types.Phase    `json:"phase"`
	State        *state.State   `json:"state"`
	UserHint     string         `json:"userHint,omitempty"`     // present when re-running after edit-and-approve
	Aggregate    string         `json:"aggregate,omitempty"`    // for per-aggregate gate granularity
}

// Output is what a specialist returns. The orchestrator validates each
// item's source_evidence (where applicable) and runs the validation funnel
// before presenting the gate.
type Output struct {
	Phase     types.Phase           `json:"phase"`
	Items     []types.OutputItem    `json:"items"`
	Summary   string                `json:"summary"`
	Decisions []DecisionRequest     `json:"decisions,omitempty"`
}

// DecisionRequest is a question the user must answer before the phase can
// proceed. Each maps to a `requires_decision` item; surfacing them
// separately lets the orchestrator batch them at the gate.
type DecisionRequest struct {
	ID       string   `json:"id"`
	Question string   `json:"question"`
	Choices  []string `json:"choices"`
	Context  string   `json:"context,omitempty"`
}

// Specialist is the contract every specialist implements.
type Specialist interface {
	// Phase returns the phase this specialist handles.
	Phase() types.Phase

	// Name returns the specialist's canonical name (e.g., "model-specialist").
	// Used for logging, metrics, and subagent file naming.
	Name() string

	// Run produces an Output for the given Input. Implementations are
	// expected to call out to the LLM for analysis and patch generation.
	// Errors here are infrastructure errors (LLM unreachable, JSON parse
	// failure); domain errors flow through Output.Items.
	Run(ctx context.Context, in *Input) (*Output, error)
}

// Registry maps phases to specialists. The orchestrator dispatches via
// this registry. Specialists register themselves on package init.
type Registry struct {
	specialists map[types.Phase]Specialist
}

// NewRegistry returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{specialists: make(map[types.Phase]Specialist)}
}

// Register adds a specialist for its declared Phase. Panics if two
// specialists claim the same phase.
func (r *Registry) Register(s Specialist) {
	if _, exists := r.specialists[s.Phase()]; exists {
		panic("specialists: phase " + s.Phase().String() + " already has a registered specialist")
	}
	r.specialists[s.Phase()] = s
}

// Get returns the specialist for the given phase, or nil if none registered.
func (r *Registry) Get(p types.Phase) Specialist {
	return r.specialists[p]
}

// Phases returns all phases that have a specialist registered.
func (r *Registry) Phases() []types.Phase {
	phases := make([]types.Phase, 0, len(r.specialists))
	for p := range r.specialists {
		phases = append(phases, p)
	}
	return phases
}
