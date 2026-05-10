// Package architecture codifies Trabuco's module dependency hierarchy as
// structured data. It is the single source of truth that every AI surface
// (CLAUDE.md, AGENTS.md, the trabuco-planner subagent, MCP server
// instructions) renders from. New modules added to internal/config/modules.go
// MUST also be placed in a layer here so that plan-stratification guidance
// stays correct without prose drift.
//
// The hierarchy enables module-stratified plan generation: agents working
// in a Trabuco-generated project structure multi-module changes as ordered
// stages, one per affected module layer, in dependency order. This package
// provides the data the agents reason over.
package architecture

import (
	"strconv"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// Layer is one rung in the dependency hierarchy. Layers are ordered by
// Index (Foundation = 0, then Contracts, Persistence, BusinessLogic,
// Edge). Modules within a layer are siblings — independent of each other.
type Layer struct {
	Index       int       // 0..N, monotonic; lower depends-on never include higher
	Name        string    // e.g. "Foundation", "Contracts", "Persistence", "BusinessLogic", "Edge"
	Description string    // one-line layer purpose, suitable for rendered docs
	Modules     []Module  // modules at this layer in canonical display order
}

// Module is a node in the hierarchy paired with prose detail used to
// render guidance. Field order intentionally narrows from identity →
// purpose → location → role-in-plans.
type Module struct {
	Name         string   // matches config.ModuleX constant
	Layer        int      // back-reference to Layer.Index
	DependsOn    []string // module names from lower layers
	Purpose      string   // single-sentence: what lives in this module
	SubPackages  []string // canonical sub-packages emitted by Trabuco
	Patterns     []string // representative file types
	Internal     bool     // mirrors config.Module.Internal — auto-included with parent
	AutoIncluded string   // module name that drags this one in (e.g. Worker → Jobs)
	StageHint    string   // role this module plays in stratified plans
}

// hierarchy is the canonical full dependency graph in layer order. It
// includes every Trabuco module — runtime selection determines which
// rows render in any given project.
var hierarchy = []Layer{
	{
		Index:       0,
		Name:        "Foundation",
		Description: "Domain types with no inter-module dependencies. Always present.",
		Modules: []Module{
			{
				Name:        config.ModuleModel,
				Layer:       0,
				DependsOn:   nil,
				Purpose:     "Entities, DTOs, enums, and domain types using Immutables.",
				SubPackages: []string{"entities", "dto"},
				Patterns:    []string{"Immutables interfaces", "Spring Data JDBC records / Mongo documents"},
				StageHint:   "Stage 1 — always first when a feature touches data shapes.",
			},
		},
	},
	{
		Index:       1,
		Name:        "Contracts",
		Description: "Data contracts between producers and consumers. Auto-included with their consumer modules so any caller can publish jobs/events without depending on the executor (Worker / EventConsumer).",
		Modules: []Module{
			{
				Name:         config.ModuleJobs,
				Layer:        1,
				DependsOn:    []string{config.ModuleModel},
				Purpose:      "JobRequest record types + abstract handler base classes for JobRunr.",
				SubPackages:  []string{"jobs"},
				Patterns:     []string{"JobRequest records", "JobRequestHandler<T> base classes"},
				Internal:     true,
				AutoIncluded: config.ModuleWorker,
				StageHint:    "Stage 2 — only when a job feature is added. Decoupled from Worker so services that ENQUEUE jobs depend only on Jobs, not the executor.",
			},
			{
				Name:         config.ModuleEvents,
				Layer:        1,
				DependsOn:    []string{config.ModuleModel},
				Purpose:      "Sealed event hierarchies + EventPublisher abstraction.",
				SubPackages:  []string{"events"},
				Patterns:     []string{"Sealed event records", "EventPublisher interface"},
				Internal:     true,
				AutoIncluded: config.ModuleEventConsumer,
				StageHint:    "Stage 2 — only when an event feature is added. Services that PUBLISH events depend only on Events, not on EventConsumer.",
			},
		},
	},
	{
		Index:       2,
		Name:        "Persistence",
		Description: "Database access layer. SQLDatastore and NoSQLDatastore are mutually exclusive in standard projects.",
		Modules: []Module{
			{
				Name:        config.ModuleSQLDatastore,
				Layer:       2,
				DependsOn:   []string{config.ModuleModel},
				Purpose:     "Spring Data JDBC repositories + Flyway migrations.",
				SubPackages: []string{"repository", "db/migration"},
				Patterns:    []string{"CrudRepository<Record, Long>", "V{N}__name.sql files"},
				StageHint:   "Stage 3 — schema + repository. Migration files belong here.",
			},
			{
				Name:        config.ModuleNoSQLDatastore,
				Layer:       2,
				DependsOn:   []string{config.ModuleModel},
				Purpose:     "Spring Data MongoDB repositories (or Redis abstractions).",
				SubPackages: []string{"repository"},
				Patterns:    []string{"MongoRepository<Document, String>", "@Indexed annotations"},
				StageHint:   "Stage 3 — document repositories. No Flyway; schema is implicit.",
			},
		},
	},
	{
		Index:       3,
		Name:        "BusinessLogic",
		Description: "Service layer. Holds @Service classes, circuit breakers, and cross-module auth utilities (RequestContextHolder, JwtClaimsExtractor).",
		Modules: []Module{
			{
				Name:        config.ModuleShared,
				Layer:       3,
				DependsOn:   []string{config.ModuleModel},
				Purpose:     "Business logic services + auth runtime utilities.",
				SubPackages: []string{"service", "auth"},
				Patterns:    []string{"@Service classes", "@CircuitBreaker boundaries", "AuthScope / RequestContextHolder helpers"},
				StageHint:   "Stage 4 — orchestration between persistence and edges. Wraps datastore calls in circuit breakers.",
			},
		},
	},
	{
		Index:       4,
		Name:        "Edge",
		Description: "Runtime modules that consume everything below. Independent of each other — multiple edges may be present and addressed as parallel stages within the same layer.",
		Modules: []Module{
			{
				Name:        config.ModuleAPI,
				Layer:       4,
				DependsOn:   []string{config.ModuleModel, config.ModuleShared},
				Purpose:     "REST controllers, GlobalExceptionHandler, OpenAPI, dormant OIDC Resource Server.",
				SubPackages: []string{"controller", "config"},
				Patterns:    []string{"@RestController classes", "RFC 7807 ProblemDetail responses"},
				StageHint:   "Stage 5 — HTTP boundary. Always last for features that surface via REST.",
			},
			{
				Name:        config.ModuleWorker,
				Layer:       4,
				DependsOn:   []string{config.ModuleModel, config.ModuleJobs},
				Purpose:     "JobRunr concrete handlers — the @Component subclasses that override the abstract handlers from Jobs.",
				SubPackages: []string{"handler", "config"},
				Patterns:    []string{"@Component JobRequestHandler subclasses", "RecurringJobsConfig"},
				StageHint:   "Stage 5 — job execution. Depends on Stage 2 (Jobs contracts) being complete first.",
			},
			{
				Name:        config.ModuleEventConsumer,
				Layer:       4,
				DependsOn:   []string{config.ModuleModel, config.ModuleEvents},
				Purpose:     "Kafka / RabbitMQ / SQS / Pub-Sub listeners with idempotency and DLT/DLQ wiring.",
				SubPackages: []string{"listener", "config"},
				Patterns:    []string{"@KafkaListener / @RabbitListener / @SqsListener / @ServiceActivator", "IdempotencyTracker.checkAndMark"},
				StageHint:   "Stage 5 — event consumption. Depends on Stage 2 (Events contracts) being complete first.",
			},
			{
				Name:        config.ModuleAIAgent,
				Layer:       4,
				DependsOn:   []string{config.ModuleModel, config.ModuleShared},
				Purpose:     "Spring AI agents, A2A protocol, MCP server, knowledge retrieval (vector RAG), guardrails.",
				SubPackages: []string{"agent", "tool", "guardrail", "knowledge", "protocol"},
				Patterns:    []string{"PrimaryAgent / SpecialistAgent classes", "@Tool methods", "DocumentRetriever implementations"},
				StageHint:   "Stage 5 — AI agent surface. Independent of API/Worker/EventConsumer.",
			},
		},
	},
}

// Hierarchy is the project-scoped view of the layer graph: only the
// layers and modules that actually exist in the given module set.
// Construct via FilterFor(modules).
type Hierarchy struct {
	Layers []Layer // filtered, in Index order
}

// FilterFor returns a Hierarchy containing only the layers/modules that
// appear in the supplied list. The mutually-exclusive pair
// SQLDatastore/NoSQLDatastore is naturally filtered: if neither is
// selected, the Persistence layer is empty and elided.
//
// Layers with zero present modules are dropped from the result so that
// rendered guidance never shows empty stages.
func FilterFor(modules []string) Hierarchy {
	present := make(map[string]bool, len(modules))
	for _, m := range modules {
		present[m] = true
	}

	var out []Layer
	for _, layer := range hierarchy {
		var kept []Module
		for _, mod := range layer.Modules {
			if present[mod.Name] {
				kept = append(kept, mod)
			}
		}
		if len(kept) == 0 {
			continue
		}
		out = append(out, Layer{
			Index:       layer.Index,
			Name:        layer.Name,
			Description: layer.Description,
			Modules:     kept,
		})
	}
	return Hierarchy{Layers: out}
}

// All returns the unfiltered hierarchy — every layer, every module —
// for use by docs that catalog all possibilities (deep reference doc,
// trabuco-planner subagent prose).
func All() Hierarchy {
	out := make([]Layer, len(hierarchy))
	copy(out, hierarchy)
	return Hierarchy{Layers: out}
}

// StagesFor returns the ordered list of stage labels (e.g. "Stage 1 —
// Model", "Stage 2 — Jobs", ...) for a project with the given modules.
// One stage per layer; layers with multiple sibling modules collapse
// into a single stage with all sibling names listed (e.g. "Stage 5 —
// API | Worker | EventConsumer").
//
// Used by templates to render the canonical stage outline in CLAUDE.md
// and AGENTS.md without each template having to re-derive it.
func StagesFor(modules []string) []string {
	h := FilterFor(modules)
	stages := make([]string, 0, len(h.Layers))
	stageNum := 1
	for _, layer := range h.Layers {
		stages = append(stages, formatStage(stageNum, layer))
		stageNum++
	}
	return stages
}

func formatStage(num int, layer Layer) string {
	if len(layer.Modules) == 1 {
		return formatStageLabel(num, layer.Modules[0].Name)
	}
	names := make([]string, len(layer.Modules))
	for i, m := range layer.Modules {
		names[i] = m.Name
	}
	return formatStageLabel(num, joinPipe(names))
}

func formatStageLabel(num int, suffix string) string {
	return "Stage " + strconv.Itoa(num) + " — " + suffix
}

func joinPipe(names []string) string {
	return strings.Join(names, " | ")
}
