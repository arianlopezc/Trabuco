// Package llm provides a shared specialist implementation that drives
// LLM-only specialists (M2-M10). Each specialist supplies its own system
// prompt and an optional input builder; the runner handles the LLM call,
// JSON parsing, and validation common to all specialists.
//
// Design: the system prompt does the reasoning work. The Go side packages
// the assessment + state into a user message, asks for a JSON response
// matching specialists.Output, and validates the result against the no-out-
// of-scope contract before returning to the orchestrator.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// Spec configures one LLM-driven specialist. All M2-M10 specialists are
// built by passing one of these to New().
type Spec struct {
	// Phase is the phase this specialist owns.
	Phase types.Phase

	// Name is the canonical specialist name (e.g., "assessor", "model").
	Name string

	// SystemPrompt is the full system prompt for the specialist. In
	// practice this is the body of plugin/agents/trabuco-migration-{name}.md
	// loaded into the Go binary at compile time via embed.
	SystemPrompt string

	// MaxTokens caps the LLM response length. Defaults to a per-phase
	// reasonable value if zero.
	MaxTokens int

	// BuildUserPrompt is an optional override that lets a specialist
	// customize what goes into the user message. Default: a JSON dump of
	// the input augmented with the assessment artifact.
	BuildUserPrompt func(in *specialists.Input) (string, error)
}

// Specialist is the LLM-driven implementation of specialists.Specialist.
type Specialist struct {
	spec     Spec
	provider ai.Provider
}

// New constructs a Specialist from a Spec. The provider is resolved lazily
// on first Run() so importers don't need to wire up auth at registration
// time.
func New(spec Spec) *Specialist {
	return &Specialist{spec: spec}
}

// Phase implements specialists.Specialist.
func (s *Specialist) Phase() types.Phase { return s.spec.Phase }

// Name implements specialists.Specialist.
func (s *Specialist) Name() string { return s.spec.Name }

// Run implements specialists.Specialist. Builds the prompt, calls the LLM,
// parses the JSON output, and returns it.
func (s *Specialist) Run(ctx context.Context, in *specialists.Input) (*specialists.Output, error) {
	if s.provider == nil {
		p, err := defaultProvider()
		if err != nil {
			return nil, fmt.Errorf("no LLM provider available (run 'trabuco auth login' first): %w", err)
		}
		s.provider = p
	}

	user, err := s.buildUserPrompt(in)
	if err != nil {
		return nil, err
	}

	maxTokens := s.spec.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8000
	}

	req := &ai.AnalysisRequest{
		SystemPrompt: s.spec.SystemPrompt + "\n\n" + outputContract,
		UserPrompt:   user,
		MaxTokens:    maxTokens,
		Temperature:  0.2, // mostly-deterministic; prompts demand JSON
	}
	resp, err := s.provider.Analyze(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM call: %w", err)
	}

	out, err := parseOutput(resp.Content, s.spec.Phase)
	if err != nil {
		return nil, fmt.Errorf("parse LLM output: %w (content: %s)", err, truncate(resp.Content, 1000))
	}
	return out, nil
}

// buildUserPrompt is the default implementation; specialists can override
// via Spec.BuildUserPrompt.
func (s *Specialist) buildUserPrompt(in *specialists.Input) (string, error) {
	if s.spec.BuildUserPrompt != nil {
		return s.spec.BuildUserPrompt(in)
	}
	return DefaultUserPrompt(in)
}

// DefaultUserPrompt is the baseline user-prompt structure: phase context +
// state.json + assessment (if present). Specialists with phase-specific
// inputs override BuildUserPrompt to add more.
func DefaultUserPrompt(in *specialists.Input) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Phase: %d (%s)\n", int(in.Phase), in.Phase)
	fmt.Fprintf(&b, "Repo root: %s\n\n", in.RepoRoot)
	if in.UserHint != "" {
		fmt.Fprintf(&b, "User guidance from previous gate (apply this in your output):\n%s\n\n", in.UserHint)
	}
	if in.Aggregate != "" {
		fmt.Fprintf(&b, "Restrict scope to aggregate: %s\n\n", in.Aggregate)
	}

	stateJSON, err := json.MarshalIndent(in.State, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal state: %w", err)
	}
	fmt.Fprintf(&b, "Current state.json:\n```json\n%s\n```\n\n", stateJSON)

	if in.Phase != types.PhaseAssessment {
		// Assessment artifact should exist by now; include for grounding.
		assessPath := state.AssessmentPath(in.RepoRoot)
		if data, err := readFileBest(assessPath); err == nil {
			fmt.Fprintf(&b, "Assessment artifact (.trabuco-migration/assessment.json):\n```json\n%s\n```\n\n", data)
		}
	}

	return b.String(), nil
}

// outputContract is appended to every system prompt; it tells the LLM
// exactly what JSON shape to return so the Go side can parse without LLM-
// specific tooling. The shape mirrors specialists.Output verbatim.
const outputContract = `

# Output contract

You MUST respond with a single JSON object matching this Go struct exactly.
No surrounding prose, no Markdown fences, no commentary — just the JSON.

` + "```json" + `
{
  "phase": <number 0-13>,
  "summary": "<one-paragraph summary of what you did or why nothing was done>",
  "items": [
    {
      "id": "<unique id within this phase>",
      "state": "applied" | "blocked" | "requires_decision" | "not_applicable" | "retained_legacy",
      "description": "<what this item represents>",
      "source_evidence": {
        "file": "<path relative to repo root>",
        "lines": "<start-end, e.g. 12-58>",
        "content_hash": "<sha256 of the byte range, hex>"
      },
      "patch": "<unified-diff format, only when state=applied>",
      "blocker_code": "<one of the fixed BlockerCode enum values, only when state=blocked>",
      "blocker_note": "<concrete file:line context, only when state=blocked>",
      "alternatives": ["<alternative 1>", "<alternative 2>"],
      "question": "<the decision question, only when state=requires_decision>",
      "choices": ["<choice 1>", "<choice 2>"],
      "reason": "<why this is not_applicable, only when state=not_applicable>"
    }
  ],
  "decisions": [
    { "id": "<...>", "question": "<...>", "choices": ["<...>"], "context": "<...>" }
  ]
}
` + "```" + `

Constraints:
- Every item with state=applied MUST include source_evidence pointing at
  REAL file paths and line ranges in the source repo. You will be REJECTED
  if source_evidence is fabricated or doesn't match the actual file content.
- BlockerCode MUST be one of the fixed enum values from the plan. Inventing
  new codes will cause the orchestrator to reject your output.
- If your phase has no work to do (no source artifacts match its scope),
  return ONE item with state=not_applicable and an explanation in 'reason'.
- Do NOT propose changes that aren't grounded in source evidence. The
  orchestrator's diff-inspection layer will reject scope creep.
`

func parseOutput(raw string, expectedPhase types.Phase) (*specialists.Output, error) {
	// Some models wrap responses in markdown fences despite instructions.
	clean := stripMarkdownFences(raw)

	var out specialists.Output
	if err := json.Unmarshal([]byte(clean), &out); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if out.Phase != expectedPhase {
		return nil, fmt.Errorf("phase mismatch: got %d, want %d", out.Phase, expectedPhase)
	}

	// Validate every blocker code is in the fixed enum.
	for _, item := range out.Items {
		if item.State == types.ItemBlocked && !item.BlockerCode.IsKnown() {
			return nil, fmt.Errorf("item %s: blocker_code %q not in fixed enum", item.ID, item.BlockerCode)
		}
	}

	return &out, nil
}

func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if i := strings.LastIndex(s, "```"); i != -1 {
			s = s[:i]
		}
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if i := strings.LastIndex(s, "```"); i != -1 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
