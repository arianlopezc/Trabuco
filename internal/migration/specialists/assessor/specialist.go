package assessor

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/llm"
	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

//go:embed prompt.md
var systemPrompt string

// Specialist is the Phase 0 assessor.
//
// Unlike phases 1-13 which produce code patches, the assessor produces ONE
// deliverable: assessment.json. Its OutputItem is a single applied item
// whose patch is the assessment file write.
type Specialist struct {
	llm *llm.Specialist
}

// New constructs the assessor specialist.
func New() *Specialist {
	s := &Specialist{}
	s.llm = llm.New(llm.Spec{
		Phase:           types.PhaseAssessment,
		Name:            "assessor",
		SystemPrompt:    systemPrompt,
		MaxTokens:       16000,
		BuildUserPrompt: s.buildPrompt,
	})
	return s
}

// Phase implements specialists.Specialist.
func (s *Specialist) Phase() types.Phase { return types.PhaseAssessment }

// Name implements specialists.Specialist.
func (s *Specialist) Name() string { return "assessor" }

// Run implements specialists.Specialist. The assessor is unique: it
// produces assessment.json AND a single OutputItem so the orchestrator
// can present the assessment as a gate to the user.
func (s *Specialist) Run(ctx context.Context, in *specialists.Input) (*specialists.Output, error) {
	out, err := s.llm.Run(ctx, in)
	if err != nil {
		return nil, err
	}

	// The assessor is allowed (uniquely) to embed an Assessment struct in
	// the OutputItem.Patch field as JSON. We extract and persist it to
	// .trabuco-migration/assessment.json.
	var assessment *Assessment
	for _, item := range out.Items {
		if item.State == types.ItemApplied && item.Patch != "" {
			var a Assessment
			if err := json.Unmarshal([]byte(item.Patch), &a); err == nil {
				assessment = &a
				break
			}
		}
	}
	if assessment == nil {
		return nil, fmt.Errorf("assessor produced no parsable Assessment in any applied item")
	}
	assessment.GeneratedAt = time.Now().UTC().Format(time.RFC3339)

	path := state.AssessmentPath(in.RepoRoot)
	if err := Save(path, assessment); err != nil {
		return nil, fmt.Errorf("save assessment.json: %w", err)
	}

	// Also seed state.SourceConfig from the assessment so downstream
	// phases have it without re-reading assessment.json.
	in.State.SourceConfig.BuildSystem = assessment.BuildSystem
	in.State.SourceConfig.Framework = assessment.Framework
	in.State.SourceConfig.JavaVersion = assessment.JavaVersion
	in.State.SourceConfig.Persistence = assessment.Persistence
	in.State.SourceConfig.Messaging = assessment.Messaging
	in.State.SourceConfig.HasAIIntegration = assessment.HasAIIntegration
	in.State.SourceConfig.TestFramework = assessment.TestFramework
	for _, ci := range assessment.CISystems {
		in.State.SourceConfig.CISystems = append(in.State.SourceConfig.CISystems, ci.System)
	}

	// If user hasn't picked a TargetConfig yet, seed it with the assessor's
	// recommendation. The orchestrator's gate lets the user override.
	if len(in.State.TargetConfig.Modules) == 0 {
		in.State.TargetConfig.Modules = assessment.RecommendedTarget.Modules
		in.State.TargetConfig.Database = assessment.RecommendedTarget.Database
		in.State.TargetConfig.NoSQLDatabase = assessment.RecommendedTarget.NoSQLDatabase
		in.State.TargetConfig.MessageBroker = assessment.RecommendedTarget.MessageBroker
		in.State.TargetConfig.AIAgents = assessment.RecommendedTarget.AIAgents
		in.State.TargetConfig.CIProvider = assessment.RecommendedTarget.CIProvider
		in.State.TargetConfig.JavaVersion = assessment.RecommendedTarget.JavaVersion
	}
	if err := state.Save(in.RepoRoot, in.State); err != nil {
		return nil, fmt.Errorf("save state: %w", err)
	}

	return out, nil
}

// buildPrompt is the assessor-specific user prompt. The assessor is the
// only specialist that scans raw source; its prompt asks for direct
// inspection of pom.xml, package layout, and key Java files.
func (s *Specialist) buildPrompt(in *specialists.Input) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Phase: %d (assessment)\n", int(in.Phase))
	fmt.Fprintf(&b, "Repo root: %s\n\n", in.RepoRoot)
	if in.UserHint != "" {
		fmt.Fprintf(&b, "User guidance from previous assessment iteration:\n%s\n\n", in.UserHint)
	}
	fmt.Fprintf(&b, "User-supplied target config (may be empty — recommend if so):\n%v\n\n", in.State.TargetConfig)

	b.WriteString(`# Your task

Inspect the repository at the path above and produce a complete Assessment
catalog. You MUST:

1. Examine pom.xml or build.gradle to determine the build system, Java
   version, Spring Boot version (or alternative framework).
2. Walk the source tree and catalog every entity, repository, controller,
   service, scheduled job, message listener, message publisher, test class,
   config file, and CI/CD file.
3. Detect hardcoded credentials in source — emit them in secretsInSource.
4. Detect AI/LLM integration libraries and frameworks.
5. Determine feasibility (green / yellow / red) and a recommended Trabuco
   TargetConfig based on what you found.
6. List any top-level blocker codes from the fixed enum in the plan.

Output: emit a single OutputItem with state="applied", description="initial
assessment", and patch=<JSON-stringified Assessment struct>. The Go side
will extract the Assessment from patch and persist it to
.trabuco-migration/assessment.json.

Source evidence is OPTIONAL for the assessor (since assessment.json is the
result, not a code change). You can omit source_evidence on the single
applied item.

If you can't access source files for some reason (sandboxed environment,
permissions issue), emit state="blocked" with blocker_code="ASSESSMENT_FAILED"
in the blocker_note (this code is not in the fixed enum but is a meta-blocker
for orchestrator escalation).
`)

	return b.String(), nil
}
