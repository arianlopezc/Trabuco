# Trabuco limitations

For the live list, read `trabuco://limitations` — it's the authoritative source and includes workarounds for each entry. This doc interprets the limitations: which ones are load-bearing (they shape architecture decisions), which ones are just missing features, and how to talk about them honestly.

## Load-bearing limitations — these shape architecture

These aren't gaps to plug; they're *choices*. If a requirement centers on one of these, Trabuco may not be the right tool, or the user needs to plan around the absence.

### No frontend

Backend-only. No React, no Vue, no Thymeleaf, no HTMX. The only exception: the AIAgent exposes backend HTTP endpoints (`/chat`, `/ask`) that a frontend can call.

**How to talk about it:** "Trabuco is a backend scaffolder. For a UI, pair it with a separate frontend project (Next.js, Vaadin, whatever). Don't expect Trabuco to help with the frontend."

### No vector DB / semantic RAG in AIAgent

AIAgent ships keyword-scored knowledge base retrieval. That's it. Embeddings, pgvector, Pinecone, Weaviate — all absent. If a user wants "proper" RAG with semantic search, they're adding that manually.

**How to talk about it:** "AIAgent's knowledge base is keyword-based — fast and free but not semantic. If you need embedding-based retrieval, Spring AI's RetrievalAugmentationAdvisor integrates with 20+ vector stores; you'd wire it in the AIAgent module after generation."

### No Kubernetes / IaC / deployment

Trabuco's `docker-compose.yml` is for **local dev only**. No K8s manifests, no Helm, no Terraform, no CloudFormation.

**How to talk about it:** "Deployment is outside Trabuco's scope. You get a runnable JAR and a local-dev Docker Compose. Production deployment — whatever your team's platform is — is your problem."

### No GraphQL, no gRPC, no WebSockets

REST API only. These aren't on the roadmap.

**How to talk about it:** "If your service fundamentally speaks GraphQL/gRPC/WebSocket, Trabuco isn't the right scaffolder. You can add the libraries post-generation, but you'll be fighting the shape."

## Feature-gap limitations — plug with manual work

These are limitations in scope, not philosophy. A motivated user can add them without fighting Trabuco.

- **No rate limiting outside AIAgent** — add Bucket4j manually if needed in API module. AIAgent already has it.
- **No API versioning strategy** — implement path- or header-based versioning yourself in the API module.
- **Limited multi-tenancy** — outside the AIAgent vector store, tenant isolation is your problem. AIAgent's RAG path (`DocumentIngestionService`, `VectorKnowledgeRetriever`) does enforce tenant isolation by default since 1.12: every chunk is server-side stamped with `metadata.tenant_id = caller.tenantId()` and every retrieval filters by it. Default tenancy is one-tenant-per-credential (`tenantId = keyHash`); JWT-mode deployments override via the `tenant_id` claim. SQL/NoSQL Datastore and Worker modules do not have built-in tenant scoping — you must add it yourself.
- **Placeholder entities and migrations** — every generated project ships a `PlaceholderEntity` and a trivial migration. Users must replace these with real domain objects.
- **No custom business logic** — obviously. The Shared module gives you the *structure* for services; the code is yours.
- **No model training / fine-tuning in AIAgent** — Spring AI connects to *trained* models via API. Training is done elsewhere (Hugging Face, SageMaker, Vertex AI).

## No-longer-limitations — shipped in 1.11

These were limitations in 1.10 and earlier. Resolved in 1.11; mention them only if a user is on an old version.

- **Authentication / authorization (resource-server side)** — Whenever API or AIAgent is selected (since 1.11), Trabuco emits a complete Spring Security 6 OAuth2 Resource Server scaffolding. Dual `SecurityFilterChain` beans (JWT chain when `trabuco.auth.enabled=true`; open chain when `=false`), `JwtAuthenticationConverter` populating `RequestContextHolder`, RFC 7807 ProblemDetail handlers for 401/403, RSA-signed e2e test utilities, regression backstop for both modes. Provider-agnostic: works with Keycloak / Auth0 / Okta / Cognito / generic OIDC. Universal data types (`IdentityClaims`, `AuthorityScope`, `AuthenticatedRequest`) in Model; cross-module utilities (`RequestContextHolder`, `JwtClaimsExtractor`, `AuthScope`, `DefaultAuthContextPropagator`) in Shared; filter chain + tests in API and AIAgent. **The generated app refuses to boot** when `trabuco.auth.enabled` is unset — `SecurityConfig#validateAuthDecisionMade` enforces an explicit choice between `true` (auth on, requires `OIDC_ISSUER_URI`) and `false` (open chain — local dev or, for AIAgent, the legacy API-key path only). AIAgent's legacy `ApiKeyAuthFilter` coexists, governed by its own `app.aiagent.api-key.enabled` property (default on). Full guide: `docs/auth.md`.

## Not-limitations — things users *think* are missing but aren't

- **"No JPA"** — intentional. Spring Data JDBC is simpler and avoids lazy-loading footguns. Don't call this a limitation; call it a choice.
- **"No FK constraints in migrations"** — intentional. Trabuco's review tooling will flag you for adding them. Referential integrity is a service-layer concern.
- **"Offset pagination isn't supported"** — intentional. Keyset pagination is the only blessed approach.
- **"No Lombok"** — intentional. Immutables gives stronger guarantees and plays nicer with records.
- **"No dev Docker Compose for production"** — intentional. The compose file is for local dev only, by design.

## When to explicitly say "Trabuco isn't the right tool"

Say this when:

- The project's **core identity** is something Trabuco explicitly excludes (e.g., "a GraphQL API server," "a React SPA with minimal backend," "a gRPC service mesh component").
- The user wants a **different framework family** — Quarkus, Micronaut, Dropwizard, Ktor, Helidon. Trabuco is Spring Boot-only.
- The project needs **Boot 2.x** compatibility — Trabuco targets 3.4.2 only.
- The project needs **non-JVM languages** — Trabuco is JVM-only.
- **Auth IS the project** (e.g., building an OAuth server / IdP) — Trabuco scaffolds resource-server-side validation only (auto-emitted with API/AIAgent, requires explicit `trabuco.auth.enabled=true` at runtime). It does NOT generate identity-provider-side code (login forms, token issuance, MFA, password reset, user management). Use a hosted IdP for the producer side.

Being honest about fit saves the user from generating a project they'll fight for weeks. Don't oversell.

## When a limitation is acceptable

Most limitations are acceptable if the user knows about them upfront. The failure mode is: generate → spend two days → discover Trabuco doesn't do X → blame the tool. Prevent that by surfacing the relevant limitation BEFORE `init_project` is called. The `suggest` skill and the `trabuco-architect` subagent should both cross-check requirements against `trabuco://limitations`.
