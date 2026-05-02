# Architecture pattern recipes

Concrete playbooks for common requirements. Each recipe names the modules, the infra defaults, the follow-up skills inside the generated project, and the tradeoff you're accepting.

For live pattern data, read `trabuco://patterns`. This doc is the *playbook* — what to do once a pattern is chosen.

## Authentication note (applies to every recipe with API or AIAgent)

Every recipe below that includes API or AIAgent gets OIDC Resource Server scaffolding **auto-generated**: `IdentityClaims` / `AuthorityScope` / `AuthenticatedRequest` in Model, `JwtClaimsExtractor` / `RequestContextHolder` / `AuthContextPropagator` in Shared, dual `SecurityFilterChain` (JWT chain + open chain for explicit local-dev opt-out) and RFC 7807 ProblemDetail handlers in API/AIAgent. **The app refuses to boot** until `trabuco.auth.enabled` is set explicitly to `true` or `false` (no implicit default — `SecurityConfig#validateAuthDecisionMade` enforces this so no project silently ships with neither chain wired). For any deployed environment set all three: `trabuco.auth.enabled=true`, `OIDC_ISSUER_URI=<your IdP discovery URL>`, `OIDC_AUDIENCE=<your service's API identifier>` — the audience requirement closes the silent-empty-default token-confusion vector. For local development without an IdP set `trabuco.auth.enabled=false`. Per-provider config: see `docs/auth.md`.

---

## "CRUD REST API over a relational database"

**Pattern:** `rest-api`
**Modules:** Model, SQLDatastore, Shared, API
**Infra:** PostgreSQL (default), no broker, no worker

**After generation, do this:**
1. `/add-entity` — replace placeholder entities with real domain objects.
2. `/add-endpoint` — wire REST endpoints, with keyset pagination for list endpoints (not offset).
3. `/add-migration` — write Flyway migrations. **No FK constraints** — enforce referential integrity in the service layer.
4. Run `review-checks.sh` before committing. For a deeper security pass before merging a security-relevant PR, run `/audit` (or `/trabuco:audit`) — full 173-check sweep across auth, AI surface, AIAgent + Java platform, data + events, web + infra.

**Tradeoff:** No async work. If a request needs to kick off long-running processing, you'll later have to add Worker (via `/trabuco:add-module`).

---

## "REST API with background processing"

**Pattern:** `background-processing`
**Modules:** Model, SQLDatastore, Shared, API, Worker (+ auto Jobs)
**Infra:** PostgreSQL, JobRunr dashboard at `/jobs`

**After generation, do this:**
1. Entities + migrations as above.
2. `/add-job` — create a job request (in Jobs module) and a handler (in Worker module).
3. `/add-endpoint` — the endpoint enqueues the job via JobRunr and returns the job ID.

**Tradeoff:** Worker uses the SQL database as the job store. High-throughput job workloads will contend with your domain tables. If that becomes a problem, split the DB.

---

## "Event-driven microservice"

**Pattern:** `event-driven`
**Modules:** Model, SQLDatastore, Shared, API, EventConsumer (+ auto Events)
**Infra:** PostgreSQL + Kafka (default; RabbitMQ/SQS/Pub-Sub supported)

**After generation, do this:**
1. Define event contracts as sealed interfaces in the Events module.
2. `/add-event-handler` — register a listener in EventConsumer.
3. Publish events via the `EventPublisher` bean from Shared services.

**Tradeoff:** Requires a running broker. For local dev, Trabuco's docker-compose spins one up — but production broker ops is your problem. Kafka gives durable replay; RabbitMQ is simpler but no replay.

---

## "AI agent that exposes tools to other agents"

**Pattern:** `ai-agent` (often + API + SQLDatastore for state)
**Modules:** Model, AIAgent, (optional: API, SQLDatastore)
**Infra:** `ANTHROPIC_API_KEY` required; no broker; no worker unless you add async tools

**After generation, do this:**
1. `/add-tool` — expose a business action as an `@Tool`. Keep it small — one verb per tool.
2. `/add-guardrail-rule` — add ALLOW/BLOCK rules with positive/negative examples. Remember: guardrails are a **separate LLM call**, not inline.
3. `/add-a2a-skill` — if this agent needs to be callable by OTHER agents, register A2A skills in the agent card.
4. `/add-knowledge-entry` — for FAQ-like responses, use the keyword knowledge base (token-free).

**Tradeoff:** Keyword knowledge retrieval is NOT semantic RAG. If you need embedding-based retrieval, you'll need to add pgvector or a vector DB yourself — Trabuco doesn't ship it.

---

## "Stateless aggregation / gateway service"

**Pattern:** `microservice-light`
**Modules:** Model, Shared, API
**Infra:** No DB, no broker

**After generation, do this:**
1. Wire external HTTP clients in Shared (use Resilience4j circuit breaker).
2. `/add-endpoint` — compose downstream calls.

**Tradeoff:** No local state means no rate limiting (Trabuco doesn't ship rate limiting outside AIAgent) and no caching. If you need either, you need to add NoSQLDatastore (Redis) or a different design.

---

## "Headless batch processor / ETL"

**Pattern:** `worker-only`
**Modules:** Model, SQLDatastore, Shared, Worker
**Infra:** PostgreSQL, no HTTP

**After generation, do this:**
1. Schedule recurring jobs in Worker via JobRunr.
2. Build services in Shared that orchestrate extract/transform/load.

**Tradeoff:** No HTTP means no external triggers without adding API. For triggering from a queue, pair with EventConsumer instead.

---

## "Multi-service system" — when to use `design-system`

If the requirement involves multiple services with different bounded contexts, different scaling profiles, or team-ownership boundaries, prefer `/trabuco:design-system` over a single project.

The `design_system` MCP tool returns a topology: service definitions, inter-service contracts (events, REST calls), and a recommended module per service. Then `generate_workspace` emits the whole workspace (multiple Maven projects side by side).

**Resist decomposition when:** (a) the user has a team of one, (b) the bounded contexts aren't clearly separate, (c) there's no independent deployment lifecycle. A single multi-module project is almost always the right first step.

---

## Pattern chooser cheat sheet

- User says "REST + database" → `rest-api`
- User says "REST + async" → `background-processing`
- User says "events", "Kafka", "CQRS" → `event-driven`
- User says "AI", "chatbot", "agent", "LLM" → `ai-agent`
- User says "gateway", "proxy", "stateless" → `microservice-light`
- User says "ETL", "batch", "processor", "no API" → `worker-only`
- User wants "everything" → `full-stack-backend` (but push back — usually too much)
- Requirements span multiple distinct services → `design-system`

When in doubt, call `suggest_architecture` — it returns a confidence score. Low confidence means the requirement is ambiguous; ask clarifying questions instead of guessing.
