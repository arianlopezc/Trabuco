// Package registry collects all specialist registrations into a single
// place that the CLI and MCP layers import. Each milestone (M2-M10) adds
// its specialist's blank-import here, which triggers the specialist's
// init() to call specialists.Default().Register(...).
//
// In M1 (foundations) no specialists are registered yet — calls to
// orchestrator.RunPhase() return "no specialist registered for phase X"
// until that phase's milestone lands. This is intentional and lets the
// foundation ship cleanly without requiring all specialists to exist.
package registry

import (
	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/activator"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/assessor"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/finalizer"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/llm"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/prompts"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists/skeleton"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// init() registers every specialist with specialists.Default(). Each
// milestone (M2-M11) appends its specialist registration here.
func init() {
	r := specialists.Default()

	// M2: assessor (Phase 0) — has dedicated Go logic for assessment.json.
	r.Register(assessor.New())

	// M3: skeleton-builder (Phase 1) — Go-driven scaffolding.
	r.Register(skeleton.New())

	// M4-M10: pure LLM-driven specialists. Each gets its embedded prompt
	// from the prompts package and a phase identifier; the shared
	// llm.Specialist handles the API call, JSON parsing, and validation.
	r.Register(llm.New(llm.Spec{Phase: types.PhaseModel, Name: "model", SystemPrompt: prompts.Model, MaxTokens: 16000}))
	r.Register(llm.New(llm.Spec{Phase: types.PhaseDatastore, Name: "datastore", SystemPrompt: prompts.Datastore, MaxTokens: 16000}))
	r.Register(llm.New(llm.Spec{Phase: types.PhaseShared, Name: "shared", SystemPrompt: prompts.Shared, MaxTokens: 16000}))
	r.Register(llm.New(llm.Spec{Phase: types.PhaseAPI, Name: "api", SystemPrompt: prompts.API, MaxTokens: 16000}))
	r.Register(llm.New(llm.Spec{Phase: types.PhaseWorker, Name: "worker", SystemPrompt: prompts.Worker, MaxTokens: 12000}))
	r.Register(llm.New(llm.Spec{Phase: types.PhaseEventConsumer, Name: "eventconsumer", SystemPrompt: prompts.EventConsumer, MaxTokens: 12000}))
	r.Register(llm.New(llm.Spec{Phase: types.PhaseAIAgent, Name: "aiagent", SystemPrompt: prompts.AIAgent, MaxTokens: 12000}))
	r.Register(llm.New(llm.Spec{Phase: types.PhaseConfiguration, Name: "config", SystemPrompt: prompts.Config, MaxTokens: 8000}))
	r.Register(llm.New(llm.Spec{Phase: types.PhaseDeployment, Name: "deployment", SystemPrompt: prompts.Deployment, MaxTokens: 8000}))

	r.Register(llm.New(llm.Spec{Phase: types.PhaseTests, Name: "tests", SystemPrompt: prompts.Tests, MaxTokens: 16000}))

	// M9: activator (Phase 12) — Go-driven parent-POM rewrite + mvn invocations.
	r.Register(activator.New())

	// M10: finalizer (Phase 13) — Go-driven doctor/sync + completion report.
	r.Register(finalizer.New())
}
