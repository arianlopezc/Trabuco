package assessor

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/arianlopezc/Trabuco/internal/migration/scanner"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/llm"
	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

//go:embed prompt.md
var rawPrompt string

//go:embed schema.go
var schemaSource string

// systemPrompt is the assessor's prompt with the canonical Assessment Go
// struct appended so the LLM has the exact field names and types. Without
// this the LLM hallucinates field names (e.g., "scheduledJobs" vs "jobs",
// "broker" vs "messageBroker") and json.Unmarshal silently rejects them.
var systemPrompt = rawPrompt + "\n\n## Canonical Assessment schema (source of truth)\n\n" +
	"This is the EXACT Go struct your `patch` JSON must unmarshal into. " +
	"Field names are the JSON tags (after the `json:` tag). Use them verbatim — " +
	"a typo means the field is silently dropped and your assessment is treated " +
	"as malformed.\n\n```go\n" + schemaSource + "\n```\n"

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
	var lastUnmarshalErr error
	for _, item := range out.Items {
		if item.State == types.ItemApplied && item.Patch != "" {
			var a Assessment
			if err := json.Unmarshal([]byte(item.Patch), &a); err != nil {
				lastUnmarshalErr = fmt.Errorf("item %q: %w", item.ID, err)
				continue
			}
			assessment = &a
			break
		}
	}
	if assessment == nil {
		if lastUnmarshalErr != nil {
			return nil, fmt.Errorf("assessor's Assessment JSON did not match schema: %w", lastUnmarshalErr)
		}
		return nil, fmt.Errorf("assessor produced no applied item with non-empty patch (got %d items)", len(out.Items))
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

// buildPrompt is the assessor-specific user prompt. The Go side scans
// the source repo with internal/migration/scanner, then bundles the
// structured snapshot into the prompt. The LLM categorizes — it does NOT
// file-walk on its own (no tool-calling in our current architecture).
func (s *Specialist) buildPrompt(in *specialists.Input) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Phase: %d (assessment)\n", int(in.Phase))
	fmt.Fprintf(&b, "Repo root: %s\n\n", in.RepoRoot)
	if in.UserHint != "" {
		fmt.Fprintf(&b, "User guidance from previous assessment iteration:\n%s\n\n", in.UserHint)
	}
	fmt.Fprintf(&b, "User-supplied target config (may be empty — recommend if so):\n%v\n\n", in.State.TargetConfig)

	// Pre-scan the source repository so the LLM has concrete data.
	snap, err := scanner.Scan(in.RepoRoot)
	if err != nil {
		return "", fmt.Errorf("source pre-scan: %w", err)
	}

	fmt.Fprintf(&b, "# Source repo pre-scan\n\n")
	fmt.Fprintf(&b, "Build system: %s\n", snap.BuildSystem)
	fmt.Fprintf(&b, "Java files: %d\n", len(snap.JavaFiles))
	fmt.Fprintf(&b, "Kotlin files: %d\n", len(snap.KotlinFiles))
	fmt.Fprintf(&b, "Non-JVM files (sample): ")
	if len(snap.NonJVMFiles) > 5 {
		fmt.Fprintf(&b, "%v (and %d more)\n", snap.NonJVMFiles[:5], len(snap.NonJVMFiles)-5)
	} else {
		fmt.Fprintf(&b, "%v\n", snap.NonJVMFiles)
	}
	fmt.Fprintf(&b, "Config files: %v\n", snap.ConfigFiles)
	fmt.Fprintf(&b, "Migration files: %v\n", snap.MigrationFiles)
	fmt.Fprintf(&b, "Dockerfiles: %v\n", snap.Dockerfiles)

	if snap.RootPOM != "" {
		fmt.Fprintf(&b, "\n## Root pom.xml\n\n```xml\n%s\n```\n", truncatePOM(snap.RootPOM))
	} else if snap.RootBuild != "" {
		fmt.Fprintf(&b, "\n## Root build file\n\n```\n%s\n```\n", truncatePOM(snap.RootBuild))
	}

	if len(snap.CIFiles) > 0 {
		fmt.Fprintf(&b, "\n## CI/CD files detected\n")
		for _, c := range snap.CIFiles {
			fmt.Fprintf(&b, "- %s (%s)\n", c.Path, c.Provider)
		}
	} else {
		fmt.Fprintf(&b, "\n## CI/CD files\n(none detected — Phase 10 deployment specialist will be not_applicable)\n")
	}

	if len(snap.DeploymentFiles) > 0 {
		fmt.Fprintf(&b, "\n## Deployment files\n")
		for _, d := range snap.DeploymentFiles {
			fmt.Fprintf(&b, "- %s (%s)\n", d.Path, d.Kind)
		}
	}

	if len(snap.ConfigFileContents) > 0 {
		fmt.Fprintf(&b, "\n## Configuration file contents (inspect for hardcoded credentials)\n\n")
		for _, c := range snap.ConfigFileContents {
			fmt.Fprintf(&b, "### %s\n```\n%s\n```\n\n", c.Path, c.Content)
		}
	}

	// List Java files with their coarse signatures. This is the
	// catalog the LLM categorizes from.
	if len(snap.JavaFiles) > 0 {
		fmt.Fprintf(&b, "\n## Java files (path, package, class, annotations, signals)\n\n")
		buf, _ := json.MarshalIndent(snap.JavaFiles, "", "  ")
		fmt.Fprintf(&b, "```json\n%s\n```\n", string(buf))
	}

	b.WriteString(`

# Your task

Use the pre-scan data above (which is authoritative — you are NOT
expected to file-walk on your own) to produce a complete Assessment
catalog. You MUST:

1. Use the pre-scan's BuildSystem, RootPOM/RootBuild, and Java file
   annotations to classify the source. The pre-scan is authoritative —
   if it lists 12 Java files with @RestController, you have 12
   controllers, not "an API layer".
2. From the per-file annotations + signals, populate entities,
   repositories, controllers, services, jobs, listeners, publishers,
   tests in the Assessment struct.
3. From the pre-scan's CIFiles and DeploymentFiles, populate
   ciSystems and deploymentFiles.
4. Detect hardcoded credentials by inspecting Java files with
   "hardcoded-credential-suspect" signals + reading the relevant
   config file content from the prompt; emit secretsInSource.
5. Determine feasibility and recommend a Trabuco TargetConfig.
6. List top-level blocker codes from the fixed enum.

Output: emit ONE OutputItem with state="applied", description="initial
assessment", and patch=<JSON-stringified Assessment struct>. The Go side
parses patch and persists it to .trabuco-migration/assessment.json.

Source evidence is OPTIONAL for the assessor item — assessment.json is
not a code change. You can omit source_evidence.

If the pre-scan returned 0 Java files (highly unusual), emit state="blocked"
with a generic explanation.
`)

	return b.String(), nil
}

// truncatePOM caps very long pom.xml content so the prompt stays under
// the model's context. Most POMs are well under 10k chars.
func truncatePOM(s string) string {
	const max = 10000
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n\n... (POM truncated for prompt size)"
}
