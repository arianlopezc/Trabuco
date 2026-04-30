# Module catalog — interpretation

This doc does NOT duplicate the module catalog — for that, read `trabuco://modules` (live, authoritative). This is the *interpretation layer*: which modules to reach for, which combinations are idiomatic, and which are code smells.

## The 8 selectable modules (Model is always auto-included)

| Module | What it ships | Pick when | Skip when |
|---|---|---|---|
| **Model** | DTOs, Entities, Enums, Exceptions (Immutables) | Always (auto) | Never skip |
| **SQLDatastore** | Spring Data JDBC, Flyway, PostgreSQL/MySQL | Relational data with clear entity shape | Pure cache, pure document store, or truly stateless |
| **NoSQLDatastore** | MongoDB or Redis repositories | Flexible schemas, key-value, caching | Relational needs, reporting queries |
| **Shared** | Services + Resilience4j circuit breaker | Any non-trivial business logic orchestration | One-file CRUD that's just `repo.findAll()` |
| **API** | Spring MVC REST + OpenAPI + correlation IDs + health + dormant OIDC Resource Server | HTTP frontend needed | Worker-only, pure event consumer, AI-only |
| **Worker** | JobRunr (background, scheduled, delayed, batch) | Async tasks, cron, retries, long-running jobs | Purely request/response; fire-and-forget over events |
| **EventConsumer** | Kafka/RabbitMQ/SQS/Pub/Sub listeners + `EventPublisher` | Event-driven, CQRS, replay semantics, cross-service choreography | Intra-service only; `@Async` is enough |
| **AIAgent** | Spring AI + guardrails + A2A + MCP server + knowledge base + dormant OIDC Resource Server (alongside legacy ApiKeyAuthFilter) | LLM-backed features, tool calling, agent-to-agent | Hardcoded ML inference, embeddings-only RAG |

**Internal modules** (auto-included, never select directly): `Jobs` (auto with Worker), `Events` (auto with EventConsumer).

## Idiomatic combinations

These are the patterns in `trabuco://patterns`. When a user's requirement maps cleanly to one, recommend it by name — don't reinvent the combo.

- `rest-api` → Model + SQLDatastore + Shared + API (Postgres default)
- `rest-api-nosql` → Model + NoSQLDatastore + Shared + API (Mongo default)
- `event-driven` → Model + SQLDatastore + Shared + API + EventConsumer (+ Kafka)
- `background-processing` → Model + SQLDatastore + Shared + API + Worker
- `full-stack-backend` → Model + SQLDatastore + Shared + API + Worker + EventConsumer
- `microservice-light` → Model + Shared + API (stateless)
- `worker-only` → Model + SQLDatastore + Shared + Worker (no HTTP)
- `ai-agent` → Model + AIAgent (often combined with API + SQLDatastore for state)

When the requirement genuinely doesn't map, use `suggest_architecture` — it scores patterns and returns a confidence. Don't guess combos.

## Non-obvious rules

- **SQLDatastore and NoSQLDatastore conflict** — you can pick one or neither, never both. If a user truly needs both, that's a decomposition signal: use `/trabuco:design-system` to split into two services.
- **Worker has NO hard dependency on SQLDatastore anymore** — JobRunr defaults to embedded storage if no datastore is selected. But in practice, always pair Worker with a datastore unless you're doing pure stateless scheduling.
- **EventConsumer pulls in Events automatically** — don't try to select Events directly; it's internal.
- **AIAgent is independent of API** — the agent exposes its own `/chat`, `/ask`, `/a2a`, `/mcp` endpoints via its own controllers. If the user wants their own REST endpoints alongside the agent, they need API too.
- **Shared is not technically required**, but projects without it tend to stuff business logic into controllers. Recommend Shared for anything beyond toy CRUD.

## Code smells in module selection

- Picking **every** module "to be safe" — Trabuco is opinionated on minimalism. Unused modules are dead weight.
- Picking **API + Worker + EventConsumer for a single-purpose service** — that's usually a sign of unresolved bounded contexts. Suggest decomposition.
- Picking **AIAgent for deterministic workflows** — LLMs are expensive and non-deterministic. If the logic is a state machine, use Worker + Shared.
- Picking **NoSQLDatastore for reporting-heavy workloads** — MongoDB aggregation pipelines are painful compared to SQL. Push back.

## What a module does NOT include

Every module has a `does_not_include` field — read it. Common surprises:

- API auto-includes OIDC Resource Server scaffolding (since 1.11) — generated dormant, activated at runtime via `trabuco.auth.enabled=true` + `OIDC_ISSUER_URI`. The runtime utilities live in Shared, so picking API auto-resolves Shared as a hard dependency. API does NOT include rate limiting (Bucket4j config is wired in `application.yml` but disabled by default), API versioning strategies, or identity-provider-side flows (login forms, token issuance — pair with a hosted IdP).
- SQLDatastore does NOT include JPA/Hibernate (Spring Data JDBC only) — users expecting `@OneToMany` will be surprised.
- Worker's dashboard is JobRunr's default; there's no custom monitoring UI.
- AIAgent does NOT include a frontend chat UI, vector DB, or fine-tuning — it's a backend agent framework.

When a user asks for something, check the relevant module's `does_not_include` against their ask. If it's on the excluded list, surface it BEFORE they generate.
