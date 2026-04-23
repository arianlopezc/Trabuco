# Trabuco philosophy

Trabuco is an opinionated Java/Spring Boot scaffolder. It is NOT a generic starter. Its choices express a specific view of how production Java services should be built. When you advise users, reason from these principles — not from generic Spring best-practice articles.

## Opinionated by design

Trabuco makes the choices most teams spend weeks debating and then sets them in stone so every generated project inherits the same shape. If a user pushes against a convention, that's the right time to ask *why* — usually there's a reason, occasionally there's a blind spot to log.

**Hard conventions** (enforced by the generated review tooling, not merely suggested):

- **Spring Boot 3.4.2 + Java 25** — no backports, no Boot 2.x compatibility.
- **Maven multi-module** — Gradle is not supported; Ant is not supported. The `init_project` tool emits Maven-only.
- **Constructor injection only** — no `@Autowired` fields, no setter injection. The review subagents flag this.
- **Immutables + builder pattern** — all DTOs/entities are immutable records or Immutables-generated classes. No Lombok.
- **Keyset pagination** — offset pagination is banned in generated code. The `perf.offset-pagination` review rule flags it.
- **No FK constraints in Flyway migrations** — denormalization is preferred; joins happen in the service layer. The review tooling flags FK usage.
- **Testcontainers for integration tests, never `@MockBean` of the database** — the `test.no-db-mock` review rule catches this. Reason: we've shipped bugs where mocked tests passed but real migrations failed in prod.
- **Never `Thread.sleep` in tests** — use Awaitility or Testcontainers wait strategies.

## Modules over micro-features

Trabuco's generation unit is the *module*, not the feature. A module is a coherent Maven module with a specific role (persistence, API, worker, events, AI). Each module ships with:

- A fixed dependency shape (what it depends on, what it conflicts with).
- A fixed boundary (what it does, what it explicitly does NOT include).
- A set of skills inside the generated project (`/add-entity`, `/add-endpoint`, etc.) that know how to extend it correctly.

Never recommend "add a little bit of X to the service module." Recommend a module or don't.

## AI-native, not AI-bolted-on

The AIAgent module is a first-class module, not an afterthought. It ships:

- Spring AI `@Tool` methods exposed via both HTTP and MCP (`/mcp` endpoint on localhost:8080).
- A separate-LLM guardrail classifier (NOT inline in the main prompt).
- A2A protocol endpoints so the generated agent can talk to other agents.
- A keyword-based knowledge base (NOT a vector DB — Trabuco doesn't ship pgvector).
- `ScopeEnforcer` + `RateLimiter` + `CallerIdentityFilter` security chain.

The **generated project itself** surfaces skills to coding agents — `/add-tool`, `/add-guardrail-rule`, `/add-a2a-skill`, `/add-knowledge-entry`. This closes the loop: the coding agent that built the project can extend it using project-local skills that know the conventions.

## The two-layer contract with coding agents

There are TWO places skills live, and they serve different purposes:

1. **This plugin's skills** (`/trabuco:new-project`, `/trabuco:add-module`, etc.) — for working on the *host* project: generating it, extending it, migrating to it. These call MCP tools.
2. **Skills inside the generated project** (`/add-entity`, `/add-endpoint`, `/add-tool`, etc.) — for working *inside* a generated Trabuco project after it exists. These don't need MCP; they know the conventions by reading the project's own CLAUDE.md and the review-checks tooling.

Never confuse the two. When a user says "I need to add a new endpoint to my Trabuco project," the right answer is "use the `/add-endpoint` skill inside the project," not "use an MCP tool from the plugin."

## Build review into the generation

Every Trabuco project ships a `review-checks.sh` script and a set of review-subagents that catch violations of the conventions above. A generated project is a *self-correcting* project — the same opinions that drove its creation continue to enforce themselves as code evolves.

When you advise users, remind them: if something feels like a workaround for a Trabuco convention, the review layer will flag it. Lean into the convention instead.
