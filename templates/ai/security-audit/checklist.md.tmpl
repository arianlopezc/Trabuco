# Trabuco Security Audit — Master Checklist

This file is the **master index** for the Trabuco security audit. It lists
every check ID by domain, with severity and a one-line title. The
[`/audit`](./SKILL.md) skill loads this file plus the five domain-detail
files and dispatches the
[`trabuco-security-audit-orchestrator`](../../agents/trabuco-security-audit-orchestrator.md)
subagent.

**Scope.** Trabuco-generated Spring Boot 3.4.x / Java 21 / Maven multi-module
projects. The check evidence patterns assume Trabuco's module shape: Model,
SQLDatastore, NoSQLDatastore, Shared, API, Worker (JobRunr), EventConsumer
(Kafka / RabbitMQ / SQS / PubSub), AIAgent (Spring AI 1.0.5), with a
dormant OIDC JWT resource server, ApiKeyAuthFilter, ScopeEnforcer, RFC 7807
GlobalExceptionHandler, application.yml, Docker Compose, Flyway,
Testcontainers, and GitHub Actions CI.

## How the orchestrator uses this file

1. Reads the per-domain tables below to plan the dispatch.
2. Spawns one specialist subagent per domain in parallel
   (`trabuco-security-audit-{auth,ai-surface,aiagent-java,data-events,web-infra}`).
3. Each specialist loads only its `checklist-{domain}.md` (the detail file
   in this same directory) and the project's source tree, then walks every
   check in its domain.
4. Specialists return structured finding records keyed by the check IDs
   defined here.
5. Orchestrator merges, deduplicates, severity-sorts, writes
   `.ai/security-audit/findings.md`.

## Severity rubric

| Severity | Definition |
| -------- | ---------- |
| Critical | Default-on exploitable vulnerability or data-loss path. Block release. |
| High | Exploitable under common conditions; fix before next deploy. |
| Medium | Latent issue that becomes critical under composition (e.g., a perf foot-gun that a load test would expose). |
| Low | Hardening opportunity; fix during routine maintenance. |
| Informational | Defensible default with a documented trade-off; track but no immediate action. |

The orchestrator MUST NOT downgrade a finding below the severity floor
recorded for its check ID without an explicit `[suppress: <reason>]`
justification, which it surfaces in the report.

## Domain index

### Auth (28 checks) — see [`checklist-auth.md`](./checklist-auth.md)

| ID | Severity | Title |
| -- | -------- | ----- |
| F-AUTH-01 | Critical | Default permitAllFilterChain ships open by default; matchIfMissing=true makes "no auth" the silen... |
| F-AUTH-02 | Critical | JWT audience validation never enforced — empty default ${OIDC_AUDIENCE:} silently disables aud cl... |
| F-AUTH-03 | Critical | Hardcoded API keys in source (public-read-key, partner-secret-key) — both committed to VCS |
| F-AUTH-08 | Critical | IngestionController and other AIAgent controllers rely on @RequireScope — bypassed entirely if ap... |
| F-AUTH-10 | Critical | A2A endpoint /a2a accepts all comers — no JSON-RPC-level authentication, no mutual auth between a... |
| F-AUTH-17 | Critical | PlaceholderController (full-fat) has no @PreAuthorize and no ownership check on getById(id) — BOL... |
| F-AUTH-18 | Critical | EventController and PlaceholderJobController (full-fat) have no auth — accept any payload and enq... |
| F-AUTH-23 | Critical | tasks/get in A2AController returns task results without verifying caller owns the task — task IDs... |
| F-AUTH-04 | High | application.yml's agent.auth.keys config is dead — filter ignores it |
| F-AUTH-05 | High | API-key comparison via Map.get(plaintext) — hash-table lookup, not constant-time comparison |
| F-AUTH-06 | High | ApiKeyAuthFilter is a @Component bean, not registered with HttpSecurity — runs OUTSIDE the Securi... |
| F-AUTH-07 | High | actuator/prometheus is permitAll() even when JWT chain is active |
| F-AUTH-09 | High | ScopeEnforcer uses string-based tier comparison with hardcoded ladder; tierLevel("public")=1, tie... |
| F-AUTH-11 | High | RateLimiter keys windows by keyHash — anonymous callers SHARE one window across the whole internet |
| F-AUTH-15 | High | AuthScope documented as the safe set/clear pattern but Worker/EventConsumer modules in full-fat-a... |
| F-AUTH-25 | High | IngestionController accepts Map<String, Object> metadata directly into vector-store payload — no ... |
| F-AUTH-12 | Medium | permitAllFilterChain disables CSRF for stateless mode but still permits cookie-bearing requests v... |
| F-AUTH-13 | Medium | JWT signature algorithm not constrained — accepts whatever JwtDecoders.fromIssuerLocation allows;... |
| F-AUTH-14 | Medium | RequestContextHolder uses raw ThreadLocal with documented "always clear in finally" — no filter a... |
| F-AUTH-16 | Medium | JwtAuthenticationConverter derives authorities only from scope/scp claim — ignores roles, groups,... |
| F-AUTH-22 | Medium | agentPermitAllFilterChain configures cors() only on the API SecurityConfig, NOT on the AIAgent ag... |
| F-AUTH-26 | Medium | Two SecurityFilterChains gated by @ConditionalOnProperty are mutually exclusive ONLY because of s... |
| F-AUTH-27 | Medium | JwtAuthenticationConverter.convert() mutates RequestContextHolder as a side effect — Spring Secur... |
| F-AUTH-28 | Medium | No test coverage for "audience mismatch returns 401" or "wrong issuer returns 401" — AuthEndToEnd... |
| F-AUTH-19 | Low | AuthProblemDetailHandler returns generic 401 — does NOT distinguish "expired" vs "missing" vs "fo... |
| F-AUTH-21 | Low | WWW-Authenticate header missing on 401; application/problem+json body but no scheme advertisement |
| F-AUTH-24 | Low | cors.allow-credentials=false default is fine, but WebConfig (AIAgent) hardcodes allowCredentials(... |
| F-AUTH-20 | Informational | bearer  token check is case-insensitive (toLowerCase().startsWith) — accepts BEARER, bEaReR, etc.... |

### AI Surface (17 checks) — see [`checklist-ai-surface.md`](./checklist-ai-surface.md)

| ID | Severity | Title |
| -- | -------- | ----- |
| F-AI-01 | High | Generated CLAUDE.md claims project does NOT include auth, contradicting dormant-by-default scaffo... |
| F-AI-02 | High | Generated extending-the-project.md walks users through ADDING Spring Security from scratch, while... |
| F-AI-03 | High | Generated add-a2a-skill / add-knowledge-entry / add-guardrail-rule curl examples hardcode the wel... |
| F-AI-04 | High | Plugin docs and skills repeatedly tell users "every endpoint is open … fine for local dev, do not... |
| F-AI-05 | High | app.aiagent.api-key.enabled=true (default) framed as "backward compatibility" without warning see... |
| F-AI-06 | Medium | Stop-hook block messages teach the agent the kill switches and the suppression directive |
| F-AI-07 | Medium | Migration tests specialist allows @MockBean JwtDecoder and MockJwtFactory-driven SecurityContext ... |
| F-AI-08 | Medium | Migration assessor classifies hardcoded secrets as "blocker" but config specialist treats env-var... |
| F-AI-09 | Medium | trabuco-ai-agent-expert MCP prompt teaches that guardrails save tokens as the *primary* benefit, ... |
| F-AI-10 | Medium | Generated add-tool.md allows-and-trusts LLM-synthesized inputs without mandatory parameter type b... |
| F-AI-11 | Medium | trabuco-migration-deployment specialist forbids security additions while migrating CI ("do not ad... |
| F-AI-12 | Low | trabuco-migration-orchestrator allows the user to type "approve" without per-file diff inspection |
| F-AI-13 | Low | Plugin SessionStart hook emits "MCP tools available" without warning the user MCP tools accept ar... |
| F-AI-14 | Low | Migration-mode parent POM ships with enforcement deferred for 11 phases and the orchestrator does... |
| F-AI-15 | Low | trabuco_expert MCP prompt's POST-GENERATION STEPS omit auth activation |
| F-AI-16 | Low | init_project MCP tool description claims auth scaffolding is dormant "by default" without surfaci... |
| F-AI-17 | Low | Generated test patterns recommend @MockBean of JwtDecoder with no warning that MockJwtFactory iss... |

### AIAgent + Java Platform (42 checks) — see [`checklist-aiagent-java.md`](./checklist-aiagent-java.md)

| ID | Severity | Title |
| -- | -------- | ----- |
| F-AIAGENT-01 | Critical | Hardcoded API keys in ApiKeyAuthFilter ship as the default auth path |
| F-AIAGENT-02 | Critical | DocumentIngestionService REST endpoint authorized only by hardcoded API key |
| F-AIAGENT-03 | Critical | Vector store has no tenant / authority-scope partition key |
| F-AIAGENT-04 | Critical | RetrievalAugmentationAdvisor injects raw RAG content into prompt with no provenance fence |
| F-AIAGENT-06 | Critical | MCP server exposed at /mcp with no authentication boundary |
| F-AIAGENT-07 | Critical | A2A JSON-RPC /a2a endpoint accepts anonymous callers; no @RequireScope annotation |
| F-JAVA-01 | Critical | Spring AMQP Jackson2JsonMessageConverter accepts __TypeId__ header without trusted-packages allow... |
| F-AIAGENT-05 | High | KnowledgeTools.askQuestion @Tool returns raw retrieved text to the LLM |
| F-AIAGENT-08 | High | Webhook URL is not validated against an allow-list (SSRF) |
| F-AIAGENT-09 | High | Webhook signing key is the caller's API key label, not a separate signing secret |
| F-AIAGENT-10 | High | Input guardrail fails open and is regex-only on output; agent.guardrails.enabled flag is ignored |
| F-AIAGENT-12 | High | No max-tool-call recursion / token budget on ChatClient |
| F-AIAGENT-16 | High | Default OAuth2 resource-server YAML has empty audience and no validator |
| F-JAVA-02 | High | A2A JSON-RPC envelope deserializes user-controlled params as Map<String, Object>, then unchecked-... |
| F-JAVA-03 | High | WebClient.create() in WebhookManager has no timeouts, no redirect bound, no proxy hardening, no m... |
| F-JAVA-05 | High | TaskManager.tasks and WebhookManager.webhooks are unbounded ConcurrentHashMap — DoS via memory ex... |
| F-AIAGENT-13 | Medium | Rate limiter keyed by caller key-hash; anonymous callers share one bucket |
| F-AIAGENT-14 | Medium | IngestionController and WebhookController accept @RequestBody without @Valid; metadata blob is un... |
| F-AIAGENT-15 | Medium | ResponseStatusException.getReason() echoed back in ProblemDetail detail |
| F-AIAGENT-17 | Medium | Default permit-all SecurityFilterChain leaves /actuator/prometheus, /actuator/metrics, /.well-kno... |
| F-AIAGENT-19 | Medium | ONNX embedding model downloaded over network with no checksum / signature verification |
| F-AIAGENT-21 | Medium | Scratchpad and reflection memory have no caller isolation; static KnowledgeBase is process-global |
| F-AIAGENT-22 | Medium | TaskManager retains tasks indefinitely; no caller binding on getTask or subscribe |
| F-JAVA-04 | Medium | OutputGuardrailAdvisor PII regex has malformed character class [A-Z\|a-z] (matches the literal pip... |
| F-JAVA-06 | Medium | WebhookManager.dispatch HMAC uses wh.apiKey() (the literal partner-secret-key string) and re-inst... |
| F-JAVA-07 | Medium | ApiKeyAuthFilter.sha256() correctly uses SHA-256 but the hash is computed AFTER plaintext lookup ... |
| F-JAVA-08 | Medium | Inline new ObjectMapper() in two AIAgent components bypasses the Spring-managed mapper that has B... |
| F-JAVA-09 | Medium | Virtual threads enabled (spring.threads.virtual.enabled=true) with raw ThreadLocal-based CallerCo... |
| F-JAVA-10 | Medium | ResponseStatusException reflects user-controlled input back into RFC 7807 detail ("Task not found... |
| F-JAVA-12 | Medium | InputGuardrailAdvisor.classify is fail-open on exception, uses String.format to inject user input... |
| F-JAVA-13 | Medium | ChatClient and the underlying Spring AI Anthropic / Transformer / PgVector clients are configured... |
| F-AIAGENT-11 | Low | System prompts include the project name and a developer TODO list visible to the model |
| F-AIAGENT-18 | Low | CORS allowedOrigins parsed by String.split(",") with no trim/validation |
| F-AIAGENT-20 | Low | AgentRestController chat/ask endpoints fall back to a static reply when LLM is missing — bypasses... |
| F-AIAGENT-23 | Low | agent.guardrails.enabled configuration is documented but never read |
| F-AIAGENT-24 | Low | ApiKeyAuthFilter sends 403 for unknown keys instead of 401; doesn't differentiate "missing" from ... |
| F-AIAGENT-25 | Low | agent.json discovery file advertises the agent's URL and skill list publicly |
| F-AIAGENT-26 | Low | Anthropic API key default is empty; no fail-fast at boot when the agent is expected to function |
| F-JAVA-11 | Low | Default X-XSS-Protection: 1; mode=block header is OWASP-deprecated and re-introduces side-channel... |
| F-JAVA-14 | Low | Pattern.compile calls produce static patterns from constants only (no user-controlled regex), but... |
| F-JAVA-15 | Low | JobRequestHandler base/override pattern (Model defines no-op base, Worker @Component extends and ... |
| F-JAVA-16 | Low | No Bean Validation custom regex constraints; built-in @Size / @NotBlank on ChatRequest/AskRequest... |

### Data + Events (44 checks) — see [`checklist-data-events.md`](./checklist-data-events.md)

| ID | Severity | Title |
| -- | -------- | ----- |
| F-DATA-08 | Critical | Vector documents table has no tenant_id / caller_id / authority_scope column; VectorKnowledgeRetr... |
| F-EVENTS-02 | Critical | RabbitMQ listener consumes PlaceholderEvent with a default Jackson2JsonMessageConverter and **no*... |
| F-DATA-01 | High | Flyway migration runs as the same Postgres user as the runtime app (super/owner privileges leak i... |
| F-DATA-02 | High | postgres / postgres shipped as default DB credentials in env, yml, docker-compose AND .env.example |
| F-DATA-03 | High | Postgres sslmode=prefer default — silently downgrades to plaintext when server doesn't offer TLS |
| F-DATA-06 | High | Redis ships with no AUTH, no TLS, optional-password commented out |
| F-DATA-07 | High | Redis RedisTemplate uses GenericJackson2JsonRedisSerializer — ships polymorphic @class markers (J... |
| F-DATA-09 | High | IngestRequest (record) has no validation; metadata: Map<String,Object> accepted unbounded into ve... |
| F-DATA-12 | High | Worker module hardcodes username: postgres / password: postgres (no env-var indirection) — applic... |
| F-EVENTS-01 | High | Kafka JsonDeserializer.TRUSTED_PACKAGES set to a directory that already imports cross-module DTOs... |
| F-EVENTS-04 | High | Inbound event records have **zero** jakarta.validation constraints; payload size, character class... |
| F-EVENTS-05 | High | No idempotency / dedup at the consumer; the inventory template promises "Consumers should use thi... |
| F-EVENTS-09 | High | Kafka broker config: bootstrap-servers defaults to localhost:9092 with **PLAINTEXT** protocol; do... |
| F-EVENTS-10 | High | RabbitMQ ships guest:guest as the AMQP credential default with no SSL — duplicate of F-INFRA / F-... |
| F-DATA-04 | Medium | HikariCP leak-detection-threshold: 0 — connection leaks never surface |
| F-DATA-05 | Medium | AIAgent HikariCP maximum-pool-size: 5 shared with Spring AI's PgVectorStore — vector-search satur... |
| F-DATA-10 | Medium | Spring Data JDBC @Query searchByName ILIKE accepts wildcards in :search — LIKE-pattern injection ... |
| F-DATA-11 | Medium | PlaceholderRecord/PlaceholderDocument are bound directly via Spring Data — entity serves as both ... |
| F-DATA-13 | Medium | Flyway validate-on-migrate defaulted true but clean-disabled not set — flyway:clean reachable in ... |
| F-DATA-14 | Medium | VectorFlywayConfig runs createSchemas(true) + baselineOnMigrate(true) against a shared DataSource... |
| F-DATA-15 | Medium | pgvector schema migration uses CREATE EXTENSION — fails closed only if runtime user is non-superu... |
| F-DATA-16 | Medium | Qdrant API key default QDRANT_API_KEY: empty, QDRANT_USE_TLS:false, port 6334 over plaintext gRPC |
| F-DATA-17 | Medium | Spring Data Redis @RedisHash("placeholder") uses a flat hard-coded key namespace — no prefix per-... |
| F-DATA-18 | Medium | findAll() returned by repository iterators with no caller / tenant filter (NoSQL Redis path) — en... |
| F-DATA-19 | Medium | No request body / multipart size cap globally — IngestionController and any future MultipartFile ... |
| F-EVENTS-03 | Medium | Sealed-interface switch in listeners is non-exhaustive (no default branch) — a deserialized event... |
| F-EVENTS-06 | Medium | DLT/DLQ end-state is "log + TODO": no alerting, no replay tooling, no quota — a poison-pill silen... |
| F-EVENTS-07 | Medium | Kafka producer in Events/EventPublisher.java does not block / await ack, has no idempotent-produc... |
| F-EVENTS-08 | Medium | Rabbit producer publishes via fanout exchange with **empty routing key**, **no MessageDeliveryMod... |
| F-EVENTS-11 | Medium | Listener concurrency hard-coded high (Kafka setConcurrency(3), Rabbit setMaxConcurrentConsumers(1... |
| F-EVENTS-12 | Medium | RecurringJobsConfig allows free-form CRON strings with no validation; combined with the JobSchedu... |
| F-EVENTS-13 | Medium | JobRunr default-number-of-retries: 10 with exponential backoff; no max-retry cap on individual ha... |
| F-EVENTS-14 | Medium | ProcessPlaceholderJobRequest is JSON-serialized via JobRunr's default Jackson config and stored i... |
| F-EVENTS-15 | Medium | Producer adds **no message signing** or HMAC over event payload — a peer service that compromises... |
| F-EVENTS-16 | Medium | Schema versioning is absent: the @type discriminator is the **simple class name** (not FQCN, not ... |
| F-DATA-20 | Low | documents.metadata stored as freeform JSONB with no key allowlist or size cap — JSONB bloat / TOA... |
| F-DATA-21 | Low | Spring Data JDBC repository uses Modifying @Query with FOR UPDATE SKIP LOCKED but no caller-scope... |
| F-DATA-22 | Low | Spring Boot OAuth2 resource-server oauth2 client_id-style exposure: actuator /actuator/prometheus... |
| F-EVENTS-17 | Low | Event payloads logged at INFO with full eventId and name — sensitive fields would leak; LOG_LEVEL... |
| F-EVENTS-18 | Low | Kafka listener relies on KAFKA_AUTO_CREATE_TOPICS_ENABLE=true (compose default) yet @RetryableTop... |
| F-EVENTS-19 | Low | events-kafka-consumer has Spring Kafka producer config in EventConsumer's application.yml:22-25 e... |
| F-EVENTS-20 | Low | EventPublisher in Events/ is a @Service with no transactional outbox pattern; if the publisher co... |
| F-DATA-23 | Informational | application-postgres.yml-style profile variants not emitted; Postgres-specific defaults baked int... |
| F-DATA-24 | Informational | flyway_schema_history and flyway_schema_history_vector tables readable by runtime user — schema-v... |

### Web + Infra (42 checks) — see [`checklist-web-infra.md`](./checklist-web-infra.md)

| ID | Severity | Title |
| -- | -------- | ----- |
| F-INFRA-05 | Critical | JobRunr dashboard defaults to **enabled** with **no authentication**; JOBRUNR_DASHBOARD_ENABLED:t... |
| F-INFRA-15 | Critical | Spring AI MCP server enabled: true by default in AIAgent application.yml; binds on the same port ... |
| F-WEB-01 | Critical | Public REST controllers ship with no method-level authorization (@PreAuthorize / @RequireScope) —... |
| F-INFRA-01 | High | No Maven Wrapper (mvnw / .mvn/wrapper/) generated; build is non-reproducible and unverifiable |
| F-INFRA-02 | High | docker-compose.yml ships hardcoded credentials and binds DB/broker/Kafka to 0.0.0.0 on host ports |
| F-INFRA-06 | High | management.endpoints.web.exposure.include defaults expose prometheus and metrics on the applicati... |
| F-WEB-02 | High | IllegalArgumentException handler echoes raw ex.getMessage() into ProblemDetail detail — exception... |
| F-WEB-03 | High | AIAgent ships no security-headers filter — /ask, /chat, /ingest, /webhooks, /tasks/{id}/stream, /... |
| F-WEB-04 | High | WebhookController has no SSRF allow-list on registered URLs; WebhookManager.dispatch posts to att... |
| F-WEB-05 | High | IngestionController accepts unvalidated, unbounded Map<String, Object> metadata straight into the... |
| F-WEB-06 | High | PlaceholderJobController and EventController accept untyped Map<String, String> request bodies — ... |
| F-WEB-16 | High | A2AController maps /a2a with no @RequireScope and accepts arbitrary JsonRpcRequest bodies — JSON-... |
| F-INFRA-03 | Medium | .env.example files omit every secret the code actually needs (OIDC_ISSUER_URI, OIDC_AUDIENCE, ANT... |
| F-INFRA-04 | Medium | springdoc.api-docs.enabled and springdoc.swagger-ui.enabled default to true with no profile guard... |
| F-INFRA-07 | Medium | Dockerfiles use sh -c ENTRYPOINT (signal-loss + arg-injection surface) and bake JAVA_OPTS as a si... |
| F-INFRA-08 | Medium | Build-stage base image maven:3-eclipse-temurin-21 is a floating tag (no digest, no minor pin) |
| F-INFRA-09 | Medium | actions/checkout@v4 and actions/setup-java@v4 not pinned to a SHA in CI |
| F-INFRA-10 | Medium | CI workflow has no top-level permissions: block; defaults to permissive workflow token scopes |
| F-INFRA-11 | Medium | CI workflow runs against PR head with a privileged postgres service container — pull_request even... |
| F-INFRA-14 | Medium | Dockerfiles skip tests at build time (-DskipTests) — security/integration tests never run before ... |
| F-INFRA-16 | Medium | logback-spring.xml does not redact Authorization, X-API-Key, Cookie, or JWT tokens from request/r... |
| F-INFRA-19 | Medium | bucket4j-spring-boot-starter 0.12.7 is a 3rd-party non-Spring artifact pinned in parent POM; no S... |
| F-INFRA-25 | Medium | application.yml's spring.profiles.active: local default in every module — production deploys sile... |
| F-WEB-07 | Medium | server.compression.enabled=true is on by default; application/json payloads compressed despite ca... |
| F-WEB-08 | Medium | Swagger UI and OpenAPI spec served on /swagger-ui.html and /api-docs by default with no profile g... |
| F-WEB-09 | Medium | Controllers return raw Map<String, Object> and reflect request inputs back in responses — uncontr... |
| F-WEB-11 | Medium | AIAgent WebConfig registers ScopeInterceptor with no path filter and no order — runs on every req... |
| F-WEB-17 | Medium | StreamingController SSE emitter uses a 60s timeout but the taskManager.subscribe callback runs sy... |
| F-INFRA-12 | Low | review-checks.sh uses set -u only — missing set -e -o pipefail; subshell find failures are silent... |
| F-INFRA-13 | Low | review-checks.sh interpolates $GITHUB_BASE_REF directly into shell commands without sanitization |
| F-INFRA-17 | Low | info.env.enabled: false is set on API but **not** on AIAgent / Worker / EventConsumer — /actuator... |
| F-INFRA-18 | Low | Dockerfiles ship logback-config-source files (*.xml) and application.yml with potentially sensiti... |
| F-INFRA-20 | Low | Spring AI alpha modules pinned (opentelemetry-semconv 1.29.0-alpha, opentelemetry-api-incubator 1... |
| F-INFRA-21 | Low | Dockerfile's RUN mvn dependency:resolve ... 2>/dev/null \|\| true swallows resolution errors — sile... |
| F-INFRA-22 | Low | Spring AI spring-ai-rag dependency excludes javax.validation but does not pin spring-ai-bom consi... |
| F-INFRA-23 | Low | enforcer:enforce runs dependencyConvergence but explicitly **allows** legacy javax.validation:val... |
| F-INFRA-24 | Low | .dockerignore excludes .git/ but not *.iml/.idea//test sources — image bloat and potential leak o... |
| F-WEB-10 | Low | /health controller mapped with bare @RequestMapping("/health") — no method whitelist, exposes TRA... |
| F-WEB-12 | Low | RFC 7807 type URI uses unstable urn:problem-type:* placeholder — cannot be dereferenced, breaks R... |
| F-WEB-13 | Low | CSRF disabled in the permit-all chain (default) and the oauth2 chain — correct for stateless, but... |
| F-WEB-14 | Low | Default CORS allowedOrigins includes http://localhost:8080 — same-origin as the API itself, irrel... |
| F-WEB-15 | Low | AgentExceptionHandler does not handle MethodArgumentNotValidException / MissingServletRequestPara... |


## Taxonomy cross-reference

This section maps standard taxonomies (OWASP / API / LLM / ASVS / CWE) onto
the check IDs above. Each Trabuco check is mapped to one or more taxonomic
classifications in its detail file. The summary below is for navigation —
when a finding fires, the `OWASP:` / `ASVS:` / `CWE:` lines in the detail
file are authoritative.

### OWASP Top 10 (2021)

| Code | Class | Most-relevant Trabuco checks |
| ---- | ----- | --------------------------- |
| A01 | Broken Access Control | F-AUTH-09, F-AUTH-17, F-AUTH-18, F-AUTH-23, F-WEB-01 |
| A02 | Cryptographic Failures | F-AUTH-05, F-DATA-02, F-DATA-03, F-INFRA-04, F-JAVA-08 |
| A03 | Injection | F-DATA-09, F-EVENTS-01, F-EVENTS-02, F-JAVA-01 |
| A04 | Insecure Design | F-AIAGENT-04, F-AIAGENT-05, F-AIAGENT-06, F-AI-04 |
| A05 | Security Misconfiguration | F-AUTH-01, F-AUTH-04, F-AUTH-26, F-INFRA-04, F-INFRA-06, F-WEB-03 |
| A06 | Vulnerable & Outdated Components | F-INFRA-18..22 |
| A07 | Identification & Authentication Failures | F-AUTH-01, F-AUTH-02, F-AUTH-03, F-AUTH-13, F-AUTH-19 |
| A08 | Software & Data Integrity Failures | F-EVENTS-01, F-EVENTS-02, F-INFRA-01, F-JAVA-01 |
| A09 | Security Logging & Monitoring Failures | F-AUTH-19, F-DATA-15, F-INFRA-15 |
| A10 | Server-Side Request Forgery | F-WEB-04, F-AIAGENT-13, F-JAVA-03 |

### OWASP API Security Top 10 (2023)

| Code | Class | Most-relevant Trabuco checks |
| ---- | ----- | --------------------------- |
| API1 | Broken Object Level Authorization (BOLA) | F-AUTH-17, F-AUTH-23, F-AIAGENT-03 |
| API2 | Broken Authentication | F-AUTH-01..05, F-AUTH-13, F-AUTH-26 |
| API3 | Broken Object Property Level Authorization | F-AUTH-25, F-WEB-05, F-WEB-06 |
| API4 | Unrestricted Resource Consumption | F-AUTH-11, F-WEB-17, F-AIAGENT-08, F-DATA-19 |
| API5 | Broken Function Level Authorization | F-AUTH-08, F-AUTH-18, F-WEB-01 |
| API6 | Unrestricted Access to Sensitive Business Flows | F-AIAGENT-02, F-AIAGENT-13 |
| API7 | Server-Side Request Forgery | F-WEB-04 |
| API8 | Security Misconfiguration | F-INFRA-04, F-INFRA-06, F-WEB-03 |
| API9 | Improper Inventory Management | F-INFRA-04, F-INFRA-08 |
| API10 | Unsafe Consumption of APIs | F-AIAGENT-04, F-JAVA-03 |

### OWASP LLM Top 10 (2025)

| Code | Class | Most-relevant Trabuco checks |
| ---- | ----- | --------------------------- |
| LLM01 | Prompt Injection | F-AIAGENT-04, F-AIAGENT-05, F-AIAGENT-06 |
| LLM02 | Sensitive Information Disclosure | F-JAVA-04, F-AIAGENT-15 |
| LLM03 | Supply Chain | F-INFRA-18..22 |
| LLM04 | Data and Model Poisoning | F-AIAGENT-02, F-AIAGENT-03 |
| LLM05 | Improper Output Handling | F-AIAGENT-05, F-JAVA-04 |
| LLM06 | Excessive Agency | F-AIAGENT-13, F-WEB-04 |
| LLM07 | System Prompt Leakage | F-AI-09, F-AIAGENT-15 |
| LLM08 | Vector and Embedding Weaknesses | F-AIAGENT-03 |
| LLM09 | Misinformation / Overreliance | F-AIAGENT-05 |
| LLM10 | Unbounded Consumption | F-AUTH-11, F-WEB-17, F-AIAGENT-08 |

### ASVS (subset, v4.0.3)

| Code | Class | Most-relevant Trabuco checks |
| ---- | ----- | --------------------------- |
| V1.x | Architecture / Threat Model | F-AUTH-01, F-AIAGENT-01..05 |
| V2.x | Authentication | F-AUTH-02..05, F-AUTH-13 |
| V3.x | Session Management | F-AUTH-12, F-AUTH-27 |
| V4.x | Access Control | F-AUTH-08, F-AUTH-17, F-WEB-01 |
| V5.x | Validation, Sanitization, Encoding | F-EVENTS-04, F-WEB-05, F-WEB-06 |
| V6.x | Cryptography at Rest | F-AUTH-05, F-DATA-08 |
| V7.x | Error Handling & Logging | F-WEB-02, F-INFRA-15 |
| V8.x | Data Protection | F-DATA-15, F-DATA-16 |
| V9.x | Communication Security | F-DATA-03, F-INFRA-04 |
| V10.x | Malicious Code | F-INFRA-01..03 |
| V11.x | Business Logic | F-AUTH-11, F-WEB-17 |
| V12.x | File / Resource | F-WEB-05, F-AIAGENT-02 |
| V13.x | Web Service / API | F-WEB-01..06 |
| V14.x | Build & Deployment | F-INFRA-01..15 |

### CWE (selected)

| Code | Title | Trabuco checks |
| ---- | ----- | -------------- |
| CWE-79 | Cross-Site Scripting | F-WEB-09 |
| CWE-89 | SQL Injection | F-DATA-09 |
| CWE-200 | Information Exposure | F-WEB-02, F-INFRA-15 |
| CWE-269 | Improper Privilege Management | F-DATA-01 |
| CWE-285 | Improper Authorization | F-AUTH-08, F-AUTH-17 |
| CWE-287 | Improper Authentication | F-AUTH-01..05 |
| CWE-352 | CSRF | F-AUTH-12 |
| CWE-434 | Unrestricted File Upload | F-WEB-05 |
| CWE-502 | Deserialization of Untrusted Data | F-EVENTS-01, F-JAVA-01 |
| CWE-639 | Authorization Bypass via User-Controlled Key | F-AUTH-23 |
| CWE-770 | Allocation of Resources Without Limits | F-JAVA-05, F-WEB-17 |
| CWE-798 | Hardcoded Credentials | F-AUTH-03, F-DATA-02 |
| CWE-862 | Missing Authorization | F-AUTH-08, F-AUTH-17, F-WEB-01 |
| CWE-863 | Incorrect Authorization | F-AUTH-09, F-AUTH-26 |
| CWE-918 | Server-Side Request Forgery | F-WEB-04, F-JAVA-03 |
| CWE-1188 | Insecure Default Initialization | F-AUTH-01, F-AUTH-26 |

## Operator extensions

To add project-specific rules without forking this file, create
`.ai/security-audit/checklist-local.md` in the target project's root. The
orchestrator reads it alongside the canonical files. Schema is identical
to the per-domain detail files. `trabuco sync` never overwrites this file.

## References

- OWASP Top 10 (2021): https://owasp.org/Top10/
- OWASP API Security Top 10 (2023): https://owasp.org/API-Security/
- OWASP LLM Top 10 (2025): https://genai.owasp.org/llm-top-10/
- OWASP ASVS v4.0.3: https://owasp.org/www-project-application-security-verification-standard/
- CWE Top 25 (2024): https://cwe.mitre.org/top25/

