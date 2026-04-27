package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arianlopezc/Trabuco/internal/migration/orchestrator"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	// Specialist registrations (each milestone wires its specialists here):
	_ "github.com/arianlopezc/Trabuco/internal/migration/specialists/registry"
)

// registerMigrationTools registers the 11 MCP tools that drive the 1.10
// migration feature. Each tool maps 1:1 to a CLI subcommand so plugin and
// CLI mode are at parity (per §11 of MIGRATION_REDESIGN_PLAN.md).
//
// In plugin mode the orchestrator is the trabuco-migration-orchestrator
// subagent; it calls these tools to drive phases and presents gates as
// natural-language exchanges. The MCP gate is a "stub" in this mode —
// the subagent owns the user dialogue and records the user's choice via
// migrate_decision.
func registerMigrationTools(s *server.MCPServer, version string) {
	registerMigrateAssess(s, version)
	registerMigrateSkeleton(s, version)
	registerMigrateModule(s, version)
	registerMigrateConfig(s, version)
	registerMigrateDeployment(s, version)
	registerMigrateTests(s, version)
	registerMigrateActivate(s, version)
	registerMigrateFinalize(s, version)
	registerMigrateStatus(s)
	registerMigrateRollback(s, version)
	registerMigrateDecision(s, version)
	registerMigrateResume(s, version)
}

// runPhaseTool is the shared handler that backs each phase-running tool.
func runPhaseTool(repoRoot, version string, phase types.Phase) (*mcp.CallToolResult, error) {
	abs, err := resolvePath(repoRoot)
	if err != nil {
		return toolError(fmt.Sprintf("resolve path: %v", err)), nil
	}
	o := orchestrator.New(abs, version, specialists.Default(), pluginGate{})

	if !state.Exists(abs) {
		if phase != types.PhaseAssessment {
			return toolError(fmt.Sprintf("no migration initialized at %s; call migrate_assess first", abs)), nil
		}
		if err := o.Preflight(); err != nil {
			return toolError(fmt.Sprintf("preflight: %v", err)), nil
		}
		if _, err := o.Init(state.TargetConfig{}); err != nil {
			return toolError(fmt.Sprintf("init: %v", err)), nil
		}
	}

	action, err := o.RunPhase(context.Background(), phase, "")
	if err != nil {
		return toolError(fmt.Sprintf("run phase %s: %v", phase, err)), nil
	}

	st, _ := o.Status()
	return toolJSON(map[string]any{
		"phase":  phase.String(),
		"action": string(action),
		"state":  st,
	})
}

// pluginGate is the no-op Gate for plugin mode. The orchestrator subagent
// owns the user dialogue; this gate auto-approves so the orchestrator can
// surface the diff to the user via natural language and call migrate_rollback
// or migrate_decision based on the user's response.
//
// CAUTION: do not use pluginGate in CLI mode — it will silently approve
// every phase. The CLI uses its own terminalGate instead.
type pluginGate struct{}

func (pluginGate) Present(ctx context.Context, phase types.Phase, out *specialists.Output) (types.GateAction, string, error) {
	// Plugin mode: the agent inspects state.json + phase-N-output.json,
	// presents the diff to the user in conversation, and decides via
	// migrate_decision / migrate_rollback what to do next. The phase is
	// auto-approved at the Go layer so the diff sticks; the agent is
	// responsible for calling migrate_rollback if the user rejects.
	return types.GateApprove, "", nil
}

// ---------- per-tool registrations ----------

func registerMigrateAssess(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_assess",
		mcp.WithDescription("Phase 0 — Run the assessor specialist on the user's existing repo. Produces .trabuco-migration/assessment.json (the no-out-of-scope contract for all later phases). Pre-flight checks: must be a git repo with at least one commit, clean working tree, on a branch."),
		mcp.WithString("repo_path", mcp.Description("Absolute path to the user's repository"), mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return runPhaseTool(req.GetString("repo_path", ""), version, types.PhaseAssessment)
	})
}

func registerMigrateSkeleton(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_skeleton",
		mcp.WithDescription("Phase 1 — Generate the Trabuco multi-module skeleton inside the user's repo (in migration mode: enforcement deferred). Wraps existing source in a legacy/ Maven module and verifies the project still builds."),
		mcp.WithString("repo_path", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return runPhaseTool(req.GetString("repo_path", ""), version, types.PhaseSkeleton)
	})
}

func registerMigrateModule(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_module",
		mcp.WithDescription("Phases 2-8 — Run a single module specialist. Module: model | sqldatastore | nosqldatastore | shared | api | worker | eventconsumer | aiagent."),
		mcp.WithString("repo_path", mcp.Required()),
		mcp.WithString("module", mcp.Description("Module to migrate"), mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		mod := req.GetString("module", "")
		var phase types.Phase
		switch mod {
		case "model":
			phase = types.PhaseModel
		case "sqldatastore", "nosqldatastore", "datastore":
			phase = types.PhaseDatastore
		case "shared":
			phase = types.PhaseShared
		case "api":
			phase = types.PhaseAPI
		case "worker":
			phase = types.PhaseWorker
		case "eventconsumer":
			phase = types.PhaseEventConsumer
		case "aiagent":
			phase = types.PhaseAIAgent
		default:
			return toolError(fmt.Sprintf("unknown module: %q", mod)), nil
		}
		return runPhaseTool(req.GetString("repo_path", ""), version, phase)
	})
}

func registerMigrateConfig(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_config",
		mcp.WithDescription("Phase 9 — Author per-module application.yml, OpenTelemetry, structured logging."),
		mcp.WithString("repo_path", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return runPhaseTool(req.GetString("repo_path", ""), version, types.PhaseConfiguration)
	})
}

func registerMigrateDeployment(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_deployment",
		mcp.WithDescription("Phase 10 — Adapt legacy CI/CD files to the new multi-module structure. STRICT no-out-of-scope: only updates files that already exist in the legacy repo. Never adds CI/CD that wasn't there."),
		mcp.WithString("repo_path", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return runPhaseTool(req.GetString("repo_path", ""), version, types.PhaseDeployment)
	})
}

func registerMigrateTests(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_tests",
		mcp.WithDescription("Phase 11 — Per-test analysis: KEEP / ADAPT / DISCARD / CHARACTERIZE-FIRST decisions on every test in the source."),
		mcp.WithString("repo_path", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return runPhaseTool(req.GetString("repo_path", ""), version, types.PhaseTests)
	})
}

func registerMigrateActivate(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_activate",
		mcp.WithDescription("Phase 12 — Enforcement activation. Flips Maven Enforcer / Spotless / ArchUnit / Jacoco threshold from skip to enforce, runs spotless:apply, then full mvn verify to confirm the project passes its own enforcement. This is the only phase where the validation funnel runs at full strength."),
		mcp.WithString("repo_path", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return runPhaseTool(req.GetString("repo_path", ""), version, types.PhaseActivation)
	})
}

func registerMigrateFinalize(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_finalize",
		mcp.WithDescription("Phase 13 — Finalization. Runs trabuco doctor --fix and trabuco sync, generates the completion report, and (if user opts) removes the legacy/ module."),
		mcp.WithString("repo_path", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return runPhaseTool(req.GetString("repo_path", ""), version, types.PhaseFinalization)
	})
}

func registerMigrateStatus(s *server.MCPServer) {
	tool := mcp.NewTool("migrate_status",
		mcp.WithDescription("Returns the current migration state from .trabuco-migration/state.json: which phases completed, which blockers were hit, what decisions are pending."),
		mcp.WithString("repo_path", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		abs, err := resolvePath(req.GetString("repo_path", ""))
		if err != nil {
			return toolError(fmt.Sprintf("resolve path: %v", err)), nil
		}
		st, err := state.Load(abs)
		if err != nil {
			return toolError(fmt.Sprintf("no migration in progress at %s: %v", abs, err)), nil
		}
		data, _ := json.MarshalIndent(st, "", "  ")
		return mcp.NewToolResultText(string(data)), nil
	})
}

func registerMigrateRollback(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_rollback",
		mcp.WithDescription("Roll back the migration to the pre-tag of phase N (0..13). Destructive: git resets working tree to the tag and clears phases >= N from state.json."),
		mcp.WithString("repo_path", mcp.Required()),
		mcp.WithNumber("to_phase", mcp.Description("Phase number 0..13 to roll back to"), mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		abs, err := resolvePath(req.GetString("repo_path", ""))
		if err != nil {
			return toolError(fmt.Sprintf("resolve path: %v", err)), nil
		}
		toPhase := int(req.GetFloat("to_phase", -1))
		if toPhase < 0 || toPhase > 13 {
			return toolError("to_phase must be 0..13"), nil
		}
		o := orchestrator.New(abs, version, specialists.Default(), pluginGate{})
		if err := o.Rollback(types.Phase(toPhase)); err != nil {
			return toolError(fmt.Sprintf("rollback: %v", err)), nil
		}
		return toolJSON(map[string]string{"status": "rolled_back", "to_phase": types.Phase(toPhase).String()})
	})
}

func registerMigrateDecision(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_decision",
		mcp.WithDescription("Record the user's choice for a requires-decision item. Called by the orchestrator subagent after asking the user."),
		mcp.WithString("repo_path", mcp.Required()),
		mcp.WithString("decision_id", mcp.Required()),
		mcp.WithString("choice", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		abs, err := resolvePath(req.GetString("repo_path", ""))
		if err != nil {
			return toolError(fmt.Sprintf("resolve path: %v", err)), nil
		}
		o := orchestrator.New(abs, version, specialists.Default(), pluginGate{})
		err = o.RecordDecision(state.DecisionRecord{
			ID:     req.GetString("decision_id", ""),
			Choice: req.GetString("choice", ""),
		})
		if err != nil {
			return toolError(fmt.Sprintf("record decision: %v", err)), nil
		}
		return toolJSON(map[string]string{"status": "recorded"})
	})
}

func registerMigrateResume(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_resume",
		mcp.WithDescription("Resume the migration from the most recent in_progress or failed phase. Useful after a crash or after a long delay."),
		mcp.WithString("repo_path", mcp.Required()),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		abs, err := resolvePath(req.GetString("repo_path", ""))
		if err != nil {
			return toolError(fmt.Sprintf("resolve path: %v", err)), nil
		}
		st, err := state.Load(abs)
		if err != nil {
			return toolError(fmt.Sprintf("no migration in progress: %v", err)), nil
		}
		for _, p := range types.AllPhases() {
			rec := st.Phases[p]
			if rec.State == types.PhaseInProgress || rec.State == types.PhaseFailed {
				return runPhaseTool(abs, version, p)
			}
		}
		return toolJSON(map[string]string{"status": "nothing_to_resume"})
	})
}
