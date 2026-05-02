# Trabuco Security Audit — Web + Infra Domain

Web layer (controllers, error handling, security headers, CORS, SSE) and infrastructure (Docker, docker-compose, GitHub Actions CI, Maven, actuator exposure, OpenAPI surface, observability, dependency hygiene).

This file is the **detail reference** for the
`trabuco-security-audit-web-infra` specialist subagent. The orchestrator
loads this file, the master checklist (`./checklist.md`), and the
specialist's prompt, then dispatches the subagent against the project's
source tree.

**How to read each entry:**

- **`<F-...>` heading** — the stable check ID. Findings reference this ID.
- **Severity floor** — the orchestrator may not downgrade below this
  unless an explicit `[suppress: <reason>]` justification is recorded.
- **Taxonomy** — OWASP / API Security / LLM / ASVS / CWE / Trabuco-specific
  cross-references.
- **Where to look** — the file paths and line ranges in a Trabuco-generated
  project where this issue typically lands.
- **Evidence pattern** — the antipattern to grep for. Specialist subagents
  use this as the primary detection signal.
- **Why it matters** — concise explanation of the threat model.
- **Suggested fix** — the recommended remediation. Specialists include this
  in their finding records so operators don't have to think from scratch.

**Total checks in this domain: 42**
(3 Critical,
 9 High,
 16 Medium,
 14 Low,
 0 Informational)

---

## F-INFRA-05 — JobRunr dashboard defaults to **enabled** with **no authentication**; `JOBRUNR_DASHBOARD_ENABLED:true` ships as the factory default

**Severity floor:** Critical
**Taxonomy:** TRABUCO-006, MAVEN-007, OWASP-A05, OWASP-A07, CWE-306, CWE-1188

### Evidence pattern

```yaml
jobrunr:
  dashboard:
    enabled: ${JOBRUNR_DASHBOARD_ENABLED:true}
    port: ${JOBRUNR_DASHBOARD_PORT:8000}
```

### Why it matters

Anyone on the network can navigate to `http://worker-host:8000/dashboard`, see the full job queue (which leaks job parameters that often contain user IDs / payloads), and re-queue or delete arbitrary jobs. In `crud-sql-worker` and `full-fat-all-modules` the Worker container has direct DB access — the dashboard is effectively a remote-control panel. The TRABUCO-006 check is failed by default. The `Dockerfile` (Worker) `EXPOSE 8081 8082` does not list 8000, but Docker `EXPOSE` is documentation-only — the JobRunr dashboard is still reachable on `8000` inside the container and would be exposed by any `--publish-all` or any explicit port mapping. In `full-fat-all-modules` docker-compose does not include the Worker service at all, so this hits operators who deploy by hand without realizing.

### Suggested fix

Default `JOBRUNR_DASHBOARD_ENABLED` to `false`; if enabled, refuse to start unless `JOBRUNR_DASHBOARD_USERNAME` and `JOBRUNR_DASHBOARD_PASSWORD` are set (JobRunr has built-in basic auth); document explicit "use a reverse proxy" pattern; add a `BackgroundJobServerConfigurationCustomizer` that binds the dashboard to `127.0.0.1:8000`.

---

---

## F-INFRA-15 — Spring AI MCP server `enabled: true` by default in AIAgent application.yml; binds on the same port as REST and is `permitAll()`'d

**Severity floor:** Critical
**Taxonomy:** LLM-EXT-02, OWASP-A05, API-API2, CWE-1188

### Evidence pattern

```yaml
spring:
  ai:
    mcp:
      server:
        enabled: true
        name: FullFatAllModules AI Agent
        version: 1.0.0
```

### Why it matters

Even though F-AIAGENT-06 covers the runtime behavior, this finding is **infra-config-specific**: the YAML default is `enabled: true` with no companion `${MCP_ENABLED:false}` env knob to disable it. Operators have no obvious way to flip MCP off without editing source — and there is no `@ConditionalOnProperty` on `McpServerConfig.java` keyed to a Trabuco-namespace property. Compounded by F-INFRA-03 (MCP knobs not in `.env.example`), the operator cannot even discover the option exists.

### Suggested fix

Default `spring.ai.mcp.server.enabled: ${MCP_SERVER_ENABLED:false}`; add a `MCP_SERVER_ENABLED=false` line to `.env.example` with a `# SECURITY: enable only after wiring auth` comment; document that enabling MCP requires implementing an authenticator.

---

---

## F-WEB-01 — Public REST controllers ship with no method-level authorization (`@PreAuthorize` / `@RequireScope`) — even when JWT chain is activated, every endpoint is gated only by `anyRequest().authenticated()`

**Severity floor:** Critical
**Taxonomy:** OWASP-A01, API-API1, API-API5, ASVS-V4-2, CWE-862, TRABUCO-001

### Where to look

- `API/src/main/java/com/security/audit/api/controller/PlaceholderController.java:18-93` (CRUD on `/api/placeholders`)

### Evidence pattern

```java
// PlaceholderController.java:28-39 (crud-sql-worker)
@PostMapping
public ResponseEntity<ImmutablePlaceholderResponse> create(@Valid @RequestBody ImmutablePlaceholderRequest request) {
    var created = service.create(request);
    return ResponseEntity.status(HttpStatus.CREATED) ... ;
}
```

### Why it matters

When the operator follows the documented "flip `trabuco.auth.enabled=true`" path, *any* token the issuer signs unlocks every business endpoint — full CRUD on `/api/placeholders`, job submission on `/api/jobs/placeholder/process`, and RabbitMQ publish on `/api/events/placeholder`. There is no `@PreAuthorize`, no per-resource ownership predicate (BOLA), and no admin-vs-user split (BFLA). The Javadoc on `SecurityConfig` and `JwtAuthenticationConverter` explicitly tells the reader that scope-based `@PreAuthorize("hasAuthority('SCOPE_*')")` *should* be used, but the generated controllers have none — every CI run, every fresh project, every "authenticated" caller has full create/update/delete access. The crud-sql-worker `PlaceholderJobController` example is especially dangerous: a `Map<String, String>` request body triggers a JobRunr enqueue with `delaySeconds` from a query string, no scope check.

### Suggested fix

Emit `@PreAuthorize("hasAuthority(T(com.security.audit.model.auth.AuthorityScope).AUTHORITY_*)")` on every state-changing controller method in the templates; for `GET /{id}`, emit a sample `ownerId` predicate that the developer must replace. Add an `ArchUnit` test that fails the build if any `@RestController` method in `*-api` lacks a `@PreAuthorize` / `@PermitAll` annotation.

---

---

## F-INFRA-01 — No Maven Wrapper (`mvnw` / `.mvn/wrapper/`) generated; build is non-reproducible and unverifiable

**Severity floor:** High
**Taxonomy:** MAVEN-002, MAVEN-011, OWASP-A08, ASVS-V14-1, CWE-1104

### Why it matters

MAVEN-002 / OWASP-A08 require Maven Wrapper integrity (`wrapperSha256Sum`, `distributionSha256Sum`) so the developer's, CI's, and operator's Maven match. Without it: (a) every contributor uses whatever `mvn` is on their PATH — Maven version drift can silently change plugin behavior; (b) the `Dockerfile` builds use Maven from `maven:3-eclipse-temurin-21` (floating tag, no SHA digest), so a malicious or breaking Maven 3.x release in that image alters every container build; (c) `.gitignore` even contains `.mvn/wrapper/maven-wrapper.jar` exclusion as if a wrapper were expected — implying intent without delivery. Build reproducibility (MAVEN-011) is also unmet because `project.build.outputTimestamp` is not set in the parent POM.

### Suggested fix

Run `mvn wrapper:wrapper -Dmaven=3.9.x` once at template time, commit `.mvn/wrapper/maven-wrapper.properties` (with `wrapperSha256Sum`, `distributionSha256Sum`) and `mvnw`/`mvnw.cmd`; switch Dockerfiles to use `./mvnw` instead of `mvn`; add `<project.build.outputTimestamp>2026-01-01T00:00:00Z</project.build.outputTimestamp>` to the parent POM.

---

---

## F-INFRA-02 — `docker-compose.yml` ships hardcoded credentials and binds DB/broker/Kafka to `0.0.0.0` on host ports

**Severity floor:** High
**Taxonomy:** CI-005, CI-006, OWASP-A05, OWASP-A07, CWE-798, CWE-1188, ASVS-V14-1

### Evidence pattern

```yaml
postgres:
  environment:
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres   # literal, not ${DB_PASSWORD:-postgres}
  ports:
    - "5433:5432"                  # Docker default = 0.0.0.0:5433 → world-reachable on the host

rabbitmq:
  environment:
    RABBITMQ_DEFAULT_USER: guest
    RABBITMQ_DEFAULT_PASS: guest
  ports:
    - "5672:5672"
    - "15672:15672"  # management UI exposed unauthenticated to the host network

redis:
  ports: ["6379:6379"]   # NO password set; default unauthenticated
```

### Why it matters

Three compounding defects in one file:
1. **Hardcoded literals**, not `${VAR:-default}` substitution — operators who write a `.env` cannot override the password inside compose unless they also edit YAML. Defeats the documented "copy `.env.example` to `.env`" workflow (`.env.example` defines `DB_PASSWORD=postgres` but compose ignores it).
2. **`5433:5432` (etc.) maps to `0.0.0.0`** by default — on a developer laptop on a coffee-shop Wi-Fi or a VPS, every service is reachable from the LAN/internet with the trivial credentials above. CIS Docker Benchmark 5.7 / CI-005 require explicit `127.0.0.1:5433:5432`.
3. **Redis has no password set** in `nosql-redis-no-ai`. Default Redis configuration responds to anyone who can `TCP CONNECT 6379` with full unauthenticated `FLUSHALL`/`CONFIG GET *` access.
4. **RabbitMQ Management UI** (`15672`) and **Kafka broker** (`9092`) and **Zookeeper** (`2181`) are all exposed unauthenticated; any process on the host can produce/consume arbitrary messages.

Even though comments say "for local development only", the file is what gets `docker compose up`'d and the binding is identical regardless of profile. The "SECURITY NOTE" comment does not enforce anything.

### Suggested fix

Bind to `127.0.0.1:5433:5432` (and similar) in compose; switch envs to `${POSTGRES_PASSWORD:-postgres}` so `.env` overrides take effect; add `command: redis-server --requirepass ${REDIS_PASSWORD:-redis}` for redis; add `RABBITMQ_DEFAULT_PASS: ${RABBITMQ_PASSWORD:-guest}` and bind `15672` to localhost only; provide a `compose.prod.yml` that demands real env-file secrets.

---

---

## F-INFRA-06 — `management.endpoints.web.exposure.include` defaults expose `prometheus` and `metrics` on the application port; combined with secured-chain `permitAll()` (F-AUTH-07) anyone can scrape

**Severity floor:** High
**Taxonomy:** SPRING-001, OWASP-A05, ASVS-V8-1, CWE-200

### Evidence pattern

```yaml
management:
  endpoints:
    web:
      exposure:
        include: ${MANAGEMENT_ENDPOINTS:health,info,prometheus,metrics}
```

### Why it matters

SPRING-001 wants only `health` and `info` exposed publicly; everything else must be authenticated. The `/actuator/metrics` endpoint additionally lets `/actuator/metrics/jvm.threads.live` etc. be probed, which when combined with `/actuator/health` `show-details: when_authorized` (correctly defaulted) creates an asymmetry: `prometheus` shows everything `health` would withhold. Worker module exposes the same on `:8082` (Dockerfile `EXPOSE 8081 8082` — both ports advertised). EventConsumer exposes on `:8084`. None have authentication.

### Suggested fix

Default include to `health,info` only; require operator opt-in via env to add `prometheus,metrics`; remove `/actuator/prometheus` from `permitAll()` in the secured chain — instead require a `SCOPE_metrics:read` (or, better, expose actuator on a separate port bound to `127.0.0.1` and let a sidecar/proxy with auth scrape it); use `management.server.address: 127.0.0.1` for that port.

---

---

## F-WEB-02 — `IllegalArgumentException` handler echoes raw `ex.getMessage()` into ProblemDetail `detail` — exception message leakage path

**Severity floor:** High
**Taxonomy:** OWASP-A05, ASVS-V7-1, CWE-200, TRABUCO-002, SPRING-002

### Where to look

- `API/src/main/java/com/security/audit/api/config/GlobalExceptionHandler.java:115-126` (api-minimal — also identical in crud-sql-worker:226-237 and full-fat-all-modules:244-255)

### Evidence pattern

```java
@ExceptionHandler(IllegalArgumentException.class)
public ProblemDetail handleIllegalArgument(IllegalArgumentException ex, HttpServletRequest request) {
    logger.warn("Invalid argument: {}", ex.getMessage());
    ProblemDetail problem = ProblemDetail.forStatusAndDetail(
      HttpStatus.BAD_REQUEST,
      ex.getMessage() != null ? ex.getMessage() : "Invalid argument");  // <-- direct echo
    ...
}
```

### Why it matters

`IllegalArgumentException` is the "default 400" used everywhere in the codebase — `WebhookManager.deregister` throws `"Webhook not found: " + webhookId`, `JsonRpcRequest` parsing throws on enum coercion with raw values, `Document.builder().text(null)` throws Spring AI's "text must not be null", etc. Some of those messages are benign, but the contract is "every exception with an attacker-controlled substring in its message becomes a leaked 400 body." The Javadoc just two methods up promises *"the wire response never includes exception messages"* — the catch-all does sanitize, but `IllegalArgumentException` is a broader funnel than the catch-all and is hit first. Combined with the AIAgent `ResponseStatusException` handler that echoes `ex.getReason()`, an attacker probing scope errors can enumerate the tier hierarchy and the exact internal tier strings (`public`, `partner`, etc.) used elsewhere as authorization keys.

### Suggested fix

Replace the `ex.getMessage()` echo with a static `"Invalid argument provided"`; introduce an explicit `BadRequestException(String userSafeMessage)` that the codebase throws when the message *is* meant to be public, and only that subclass's message gets reflected. For the AIAgent `ResponseStatusException` handler, reduce the wire `detail` to "Insufficient privileges." and log the original reason server-side.

---

---

## F-WEB-03 — AIAgent ships no security-headers filter — `/ask`, `/chat`, `/ingest`, `/webhooks`, `/tasks/{id}/stream`, `/a2a` all respond without HSTS, X-Frame-Options, CSP, X-Content-Type-Options, Referrer-Policy, Permissions-Policy

**Severity floor:** High
**Taxonomy:** API-API8, SPRING-013, CWE-1021, ASVS-V14-1

### Where to look

- Confirmed-absent: there is no `SecurityHeadersFilter` under `aiagent-pgvector-rag/AIAgent/src/main/java/com/security/audit/aiagent/` (compare `api-minimal/API/src/main/java/com/security/audit/api/config/SecurityHeadersFilter.java` which the API module *does* ship).

### Why it matters

AIAgent endpoints ship streaming-text responses (`/tasks/{id}/stream` SSE), HTML-renderable JSON (chat output may contain LLM-emitted markdown/HTML, see LLM-05 in the AIAgent findings), and the `/.well-known/agent.json` discovery doc — none of which carry the same protective headers the API module gets. A reverse proxy that does not add HSTS upstream will leave AIAgent HTTPS connections downgrade-vulnerable; a browser that surfaces `/chat` output without CSP will execute embedded scripts. The pure-default Spring Security headers also do not include `Cache-Control: no-store` on the SSE stream, so an intermediate cache could capture a partner's task output.

### Suggested fix

Generate an AIAgent-specific `SecurityHeadersFilter` mirroring the API one (or extract the API filter into a Shared module bean both apps import); set `Cache-Control: no-store` explicitly on `StreamingController` and `WebhookController` responses; add a smoke test asserting CSP/HSTS headers on every response.

---

---

## F-WEB-04 — `WebhookController` has no SSRF allow-list on registered URLs; `WebhookManager.dispatch` posts to attacker-controlled URLs with `X-Webhook-Signature` carrying the partner's API-key-derived HMAC

**Severity floor:** High
**Taxonomy:** OWASP-A10, API-API7, CWE-918, OWASP-A04

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/protocol/WebhookController.java:27-40`

### Evidence pattern

```java
// WebhookManager.java:22, 57-62
private final WebClient webClient = WebClient.create();
...
webClient.post()
    .uri(wh.url())                 // <-- attacker-controlled; no allow-list
    .header("X-Webhook-Signature", "sha256=" + signature)
    .header("X-Webhook-ID", wh.webhookId())
    .bodyValue(payload)
    .retrieve()
    .toBodilessEntity()
    .subscribe(...);
```

### Why it matters

Two SSRF attack lines:
1. **Outbound SSRF on dispatch.** Default `WebClient.create()` follows redirects and accepts arbitrary schemes (file:, http to RFC1918 ranges, http to `169.254.169.254` AWS metadata, etc., depending on JDK). Combined with the publicly-known `partner-secret-key` (F-AIAGENT-01), an attacker registers a webhook pointing at internal IPs and waits for `dispatch()` to be called by domain code (`WebhookManager.dispatch(eventType, data)` is invoked from event publishers in the agent runtime). The agent then sends the *internal* event payload (potentially containing partner data and tool output) to the attacker-controlled URL.
2. **Header-leak side-channel.** Because the dispatch sets `X-Webhook-Signature: sha256=<hmac of payload with apiKey>`, an attacker who captures one delivery learns enough to verify their guess of the API key offline (the apiKey is the literal `partner-secret-key` string from F-AUTH-03 here, but in any future deployment that uses real keys this is a leak primitive).
3. **Storage primitive.** `register()` stores webhooks in a `ConcurrentHashMap` with no per-tenant cap; an attacker with the partner key can register thousands.

### Suggested fix

Validate `url()` against `^https://` with a host allow-list (configurable via `agent.webhooks.allowed-hosts`); resolve and reject RFC 1918 + link-local + loopback addresses *after* DNS resolution (TOCTOU-safe — pin the resolved IP and use it in the `WebClient` request); cap webhook count per tenant; disable redirect-following; add a per-tenant rate limit; gate `@RequireScope("partner")` behind a real, server-issued credential.

---

---

## F-WEB-05 — `IngestionController` accepts unvalidated, unbounded `Map<String, Object> metadata` straight into the vector store — payload size and shape are uncapped

**Severity floor:** High
**Taxonomy:** API-API3, API-API4, API-API6, ASVS-V5-1, CWE-770, OWASP-LLM-04

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/protocol/IngestionController.java:63-95`

### Evidence pattern

```java
@PostMapping
@RequireScope("partner")
public ResponseEntity<Map<String, Object>> ingest(@RequestBody IngestRequest req) {       // <-- no @Valid
    ingestionService.ingest(req.text(), req.metadata());
    ...
}

@PostMapping("/batch")
@RequireScope("partner")
public ResponseEntity<Map<String, Object>> ingestBatch(@RequestBody List<IngestRequest> requests) {  // <-- no @Valid, no size cap
    ...
}

public record IngestRequest(String text, Map<String, Object> metadata) {}   // <-- no validation
```

### Why it matters

- `text` is `null`-allowed (no `@NotBlank`, no `@Size`); a request with a 100 MB string lands in the vector store and is split-and-embedded — every chunk costs an embedding API call. With the publicly-known `partner-secret-key` (F-AIAGENT-01) this is denial-of-wallet.
- `metadata` is a free-form map that flows into Spring AI `Document.metadata()` and is persisted alongside the vector. Polymorphic deserialization is *not* the main risk (Jackson default typing isn't enabled), but the field accepts arbitrarily nested JSON; a `{"a": {"a": {"a": ...}}}` 50-deep payload triggers the default Jackson `StreamReadConstraints` 1000-deep limit but every level allocates objects and the deserializer copies the map. No `JsonNode` size cap is configured.
- `ingestBatch` accepts `List<IngestRequest>` with no element-count cap.
- The `Map<String, Object>` shape is also a *response* mass-assignment vector — the controller builds responses with `Map.of("status", "ingested", "chunks", "split-and-stored")` (a literal string!), so changes to internal state could leak through similar literal-builder paths in adjacent endpoints. (See F-WEB-09 for the broader DTO shape complaint.)

### Suggested fix

Replace `IngestRequest` with an `Immutables`-style DTO carrying `@NotBlank @Size(max = N) String text()` and `@Size(max = 32) @Valid Map<String, @Size(max = 256) String> metadata()` (or a typed metadata record); add `@Size(max = 100) @Valid` on the batch list; configure `spring.servlet.multipart.max-request-size: 1MB` and `server.tomcat.max-http-form-post-size: 1MB`; gate ingestion behind a real `SCOPE_rag:write` JWT scope (cross-ref F-AIAGENT-02).

---

---

## F-WEB-06 — `PlaceholderJobController` and `EventController` accept untyped `Map<String, String>` request bodies — no DTO, no `@Valid`, no field constraints

**Severity floor:** High
**Taxonomy:** API-API3, ASVS-V5-1, CWE-20

### Where to look

- `API/src/main/java/com/security/audit/api/controller/PlaceholderJobController.java:44-50, 64-73` (crud-sql-worker, full-fat-all-modules)

### Evidence pattern

```java
// PlaceholderJobController.java:44-50
@PostMapping("/process")
public ResponseEntity<Map<String, String>> enqueue(@RequestBody Map<String, String> body) {
    String message = body.getOrDefault("message", "default");
    placeholderJobService.processAsync(message);
    ...
}

// EventController.java:53-58
@PostMapping("/placeholder")
public ResponseEntity<Map<String, String>> publishPlaceholderEvent(
    @RequestBody Map<String, String> request) {
    String name = request.getOrDefault("name", "Unnamed");
    ...
}
```

### Why it matters

The `Map<String, String>` binding is the canonical "skip Bean Validation" pattern. There is no length limit on `message` / `name`, no character set check, no `@NotBlank`, no shape validation. Any payload that fits in the request size limit (which itself is unset, see F-WEB-05) produces a JobRunr job (`placeholderJobService.processAsync(message)`) or a RabbitMQ event (`PlaceholderCreatedEvent.create(entityId, name)`); the worker / consumer trusts the field without re-validation. A `message` containing 5 MB of newlines becomes a 5 MB job payload persisted in the JobRunr table; an attacker bomb-spams the worker queue. The `processAt(message, runAt)` variant additionally accepts `delaySeconds` from `@RequestParam` with no bound — `delaySeconds=Long.MAX_VALUE` is happily accepted, scheduling a job for the year 292278994.

### Suggested fix

Replace `Map<String, String>` with `@Value.Immutable @Validated EnqueueJobRequest` (`@NotBlank @Size(max=1024) String message()`); add `@Min(0) @Max(86400)` on `delaySeconds`; emit `@Validated` on the controller class so `@RequestParam` constraints fire.

---

---

## F-WEB-16 — `A2AController` maps `/a2a` with no `@RequireScope` and accepts arbitrary `JsonRpcRequest` bodies — JSON-RPC dispatcher is unauthenticated at the HTTP layer; only individual `tasks/*` cases call `ScopeEnforcer.requireScope`

**Severity floor:** High
**Taxonomy:** TRABUCO-003, API-API1, API-API5, OWASP-A01, CWE-862

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/protocol/A2AController.java:43-63`

### Evidence pattern

```java
@PostMapping("/a2a")           // <-- no @RequireScope, no @PreAuthorize on the controller method
public JsonRpcResponse handle(@RequestBody JsonRpcRequest body) {     // <-- no @Valid
    var caller = CallerContext.get();
    rateLimiter.checkRateLimit(caller);
    ...
    return switch (method) {
        case "tasks/send" -> handleTasksSend(rpcId, params, caller);
        case "tasks/chat" -> handleTasksChat(rpcId, params, caller);
        case "tasks/get"  -> handleTasksGet(rpcId, params);     // <-- NO scope check
        default -> JsonRpcResponse.error(rpcId, -32601, "Unknown method: " + method);
    };
}

private JsonRpcResponse handleTasksGet(String rpcId, Map<String, Object> params) {
    String taskId = (String) params.get("task_id");
    ...
    var task = taskManager.getTask(taskId);    // <-- BOLA: no caller-vs-owner check
```

### Why it matters

- The dispatcher itself has no authorization annotation — anonymous callers reach `handle()` and only the per-method branches enforce scope (`tasks/send` and `tasks/chat` call `ScopeEnforcer.requireScope("public", caller)`; `tasks/get` does not).
- `tasks/get` returns *any* task by ID — full BOLA. A caller who knows or guesses a `taskId` (UUID-like, but emitted in `tasks/send` responses to other callers) reads someone else's task result, including LLM output and tool invocation results.
- The `params` is `Map<String, Object>` (no DTO, no `@Valid`), so unbounded depth, no field length, no `@NotNull` — same shape complaint as F-WEB-05/06.
- The `JsonRpcRequest` DTO is also bound without `@Valid` — even if validation existed, it wouldn't fire.

### Suggested fix

Add `@RequireScope("public")` (or `@PreAuthorize`) to the `/a2a` controller method; add an explicit `ScopeEnforcer.requireScope("public", caller)` at the top of `handleTasksGet`; verify the caller's identity owns the task (e.g., add an `ownerKeyHash` field on `Task` and compare against `caller.keyHash()`); replace `Map<String, Object> params` with typed records for each method shape; add `@Valid` on `@RequestBody JsonRpcRequest`.

---

---

## F-INFRA-03 — `.env.example` files omit every secret the code actually needs (`OIDC_ISSUER_URI`, `OIDC_AUDIENCE`, `ANTHROPIC_API_KEY`, `QDRANT_API_KEY`, `TRABUCO_AUTH_ENABLED`, `JOBRUNR_DASHBOARD_*`)

**Severity floor:** Medium
**Taxonomy:** ASVS-V14-1, ASVS-V6-1, CWE-1188, OWASP-A05

### Why it matters

Operators discover the variables by stack-trace ("AuthenticationServiceException: issuer must not be empty") or by `grep -r '\${' src/main/resources/`. More importantly: the .env.example does not document **which variables enable security features**. A developer using `.env.example` as a checklist will never know that `TRABUCO_AUTH_ENABLED=true` and `OIDC_ISSUER_URI=https://...` must both be set — the dormant-auth design (F-AUTH-01) compounds this. Same for `JOBRUNR_DASHBOARD_ENABLED=false` to close that hole on the Worker (see F-INFRA-05).

### Suggested fix

Generate `.env.example` from a single source-of-truth schema that includes commented-out lines for every `${VAR}` referenced by any module's application.yml, with a one-line note marking each `# SECURITY-CRITICAL` variable.

---

---

## F-INFRA-04 — `springdoc.api-docs.enabled` and `springdoc.swagger-ui.enabled` default to `true` with no profile guard; production deploys leak full OpenAPI spec

**Severity floor:** Medium
**Taxonomy:** SPRING-006, OWASP-A05, API-API9, CWE-1188, TRABUCO-008

### Evidence pattern

```yaml
springdoc:
  api-docs:
    enabled: ${SPRINGDOC_ENABLED:true}
    path: /api-docs
  swagger-ui:
    enabled: ${SWAGGER_UI_ENABLED:true}
    path: /swagger-ui.html
```

### Why it matters

SPRING-006 wants Swagger off in prod or auth-gated. The `SPRINGDOC_ENABLED` / `SWAGGER_UI_ENABLED` env knobs exist but: (a) the default is `true`, so the operator must remember to flip them; (b) even if flipped, `/api-docs` is a separate flag from `/swagger-ui`, so half-disabling is easy; (c) the `permitAll()` in SecurityConfig is hardcoded — turning Swagger off does not remove the bypass for any future endpoint a dev mounts at `/swagger-ui/**`.

### Suggested fix

Default both flags to `false`; opt-in via a profile (`@Profile({"local","dev"})` on a `SwaggerConfig` bean); when enabled, gate behind a dedicated `SCOPE_swagger:read` rather than `permitAll()`; configure `springdoc.paths-to-exclude=/internal/**,/admin/**`.

---

---

## F-INFRA-07 — `Dockerfile`s use `sh -c` ENTRYPOINT (signal-loss + arg-injection surface) and bake `JAVA_OPTS` as a single string

**Severity floor:** Medium
**Taxonomy:** CI-001 (informational — non-root **is** set, good), CWE-78 (mitigated — no user input flows to ENTRYPOINT, but pattern is risky), Docker best practice

### Evidence pattern

```dockerfile
ENV JAVA_OPTS="-XX:+UseContainerSupport -XX:MaxRAMPercentage=75.0"
ENTRYPOINT ["sh", "-c", "java $JAVA_OPTS -jar app.jar"]
```

### Why it matters

1. **PID-1 / signal handling:** `sh -c` becomes PID 1; SIGTERM from `docker stop` / Kubernetes goes to `sh`, not the JVM. Spring's `server.shutdown: graceful` and `spring.lifecycle.timeout-per-shutdown-phase: 30s` (configured in every application.yml) **do not run** because the JVM never gets the signal — `sh` exits, the JVM is killed. Graceful shutdown is silently broken in containers.
2. **Variable expansion at runtime:** `$JAVA_OPTS` is expanded by `sh` from whatever `JAVA_OPTS` is set to in the container env. Anyone able to set env vars on the running container (e.g., a Kubernetes downstream operator who can `kubectl set env`) can append arbitrary JVM flags including `-Djavax.net.ssl.trustStore=/tmp/evil` or `-Dorg.apache.tomcat.util.http.ServerCookie.ALWAYS_ADD_EXPIRES=true`. Less of an attack and more of a footgun, but the signal loss is real.
3. The `sh -c` form negates the benefit of `eclipse-temurin:21-jre-alpine` (`alpine` already ships `busybox sh` so no new attack surface, but in distroless images this would error).

### Suggested fix

Either use exec form with literal flags `ENTRYPOINT ["java", "-XX:+UseContainerSupport", "-XX:MaxRAMPercentage=75.0", "-jar", "app.jar"]`, or use `JAVA_TOOL_OPTIONS` env var which the JVM reads automatically (no `sh` wrapper needed), or use `tini`/`dumb-init` as PID 1 if a wrapper is genuinely required.

---

---

## F-INFRA-08 — Build-stage base image `maven:3-eclipse-temurin-21` is a floating tag (no digest, no minor pin)

**Severity floor:** Medium
**Taxonomy:** CI-002, OWASP-A06, OWASP-A08, CWE-1104, MAVEN-005

### Why it matters

Reproducible-build / supply-chain integrity (OWASP-A08, CI-002, CWE-1104) is impossible. A new Maven release tomorrow could break the build, change plugin behavior, or — worst case — be compromised at the registry. The runtime image floats across **JDK security patches**: a CVE in an older 21.0.x patch could ship if the image hasn't been re-pulled, while a freshly pulled image fixes it — making behavior depend on `docker pull` cache state. There is no `--platform=linux/amd64` pinning either, so multi-arch builds can be inconsistent.

### Suggested fix

Pin to a specific minor + digest: `FROM maven:3.9.10-eclipse-temurin-21@sha256:abcd...`, refresh during release cycles via Renovate/Dependabot. At minimum pin the minor: `maven:3.9-eclipse-temurin-21` and `eclipse-temurin:21.0.5_11-jre-alpine`. Consider distroless `gcr.io/distroless/java21-debian12` for the runtime stage.

---

---

## F-INFRA-09 — `actions/checkout@v4` and `actions/setup-java@v4` not pinned to a SHA in CI

**Severity floor:** Medium
**Taxonomy:** CI-008, OWASP-A08, CWE-1104, CWE-829

### Where to look

`/tmp/trabuco-secaudit/api-with-ci/.github/workflows/ci.yml:34, 37, 62`

### Evidence pattern

```yaml
- uses: actions/checkout@v4
- uses: actions/setup-java@v4
  with:
    java-version: '21'
    distribution: 'temurin'
    cache: 'maven'
- uses: actions/checkout@v4
  with:
    fetch-depth: 0
```

### Why it matters

GitHub's hardening guide (CI-008, OWASP-A08) states third-party actions should be pinned to a 40-char SHA. `actions/*` are first-party and considered lower-risk, but the principle still applies — the `v4` tag is mutable and a maintainer compromise (or accidentally bad release) can ship through. The workflow runs `mvn` against secrets it inherits from the `pull_request` event scope; while it does **not** use `pull_request_target` (good — CI-007 is met), an actions compromise still has full access to the workflow's environment. A SHA-pinned form would be `actions/checkout@v4.1.7@b4ffde65...`.

### Suggested fix

Pin all uses to SHAs (e.g., `actions/checkout@b4ffde65f46336ab88eb53be808477a3936bae11 # v4.1.7`), and configure Renovate/Dependabot to update with PRs that include the SHA.

---

---

## F-INFRA-10 — CI workflow has no top-level `permissions:` block; defaults to permissive workflow token scopes

**Severity floor:** Medium
**Taxonomy:** CI-009, OWASP-A05, CWE-269

### Where to look

`/tmp/trabuco-secaudit/api-with-ci/.github/workflows/ci.yml` (entire file — no `permissions:` declaration)

### Why it matters

A compromised `mvn` plugin, a malicious test payload, or an injected step can write to the repo using `GITHUB_TOKEN` if the repo defaults are permissive. Following least-privilege (CI-009), every workflow should declare `permissions: contents: read` (or stricter) at the top level, with per-job overrides only where needed.

### Suggested fix

Add at top of file:
```yaml
permissions:
  contents: read
  pull-requests: read
```

---

---

## F-INFRA-11 — CI workflow runs against PR head with a privileged `postgres` service container — `pull_request` event uses fork code with shared service network

**Severity floor:** Medium
**Taxonomy:** CI-007, CI-014, OWASP-A08

### Where to look

`/tmp/trabuco-secaudit/api-with-ci/.github/workflows/ci.yml:6-7, 12-26`

### Evidence pattern

```yaml
on:
  pull_request:
    branches: [main]
jobs:
  build:
    services:
      postgres:
        image: postgres:15-alpine
        env:
          POSTGRES_PASSWORD: postgres
        ports:
          - 5433:5432
```

### Why it matters

Less of an issue on GitHub-hosted runners (ephemeral, no secrets, no access to your infra), but if the user moves to **self-hosted runners** (CI-014) — which they will when builds get slow — the PR Java code runs with full host access. There is no `if: github.event.pull_request.head.repo.full_name == github.repository` guard or fork-detection. There's no Docker-socket mount today (good), but if Testcontainers tests are added (and the inventory shows `RepositoryIntegrationTest` already uses Testcontainers), the runner will need Docker socket access — at which point fork PRs could escape.

### Suggested fix

Add `if: github.event.pull_request.head.repo.full_name == github.repository || github.event_name == 'push'` to gate fork PRs from sensitive jobs; document that the workflow assumes GitHub-hosted runners; add a separate, manually-approved workflow for fork CI if needed.

---

---

## F-INFRA-14 — `Dockerfile`s skip tests at build time (`-DskipTests`) — security/integration tests never run before image is published

**Severity floor:** Medium
**Taxonomy:** CI-013, OWASP-A06, ASVS-V14-1

### Why it matters

OWASP-A06 / ASVS-V14-1 require that the deployable artifact passes its own tests. Without CI and with `-DskipTests`, an image can ship with broken security tests (e.g., the `AuthEndToEndTest` that should catch missing audience validation, F-AUTH-02). For archetypes that do not generate CI, the developer's local `mvn package` is the only line of defense — they may run it without `-DskipTests` locally, but the *image they push* never gets tested. This makes images lower-trust than source.

### Suggested fix

Move tests into the build stage but allow a `--build-arg SKIP_TESTS=true` opt-out for fast iteration; add a separate "test image" target in CI; or default-emit `.github/workflows/ci.yml` for **every** archetype rather than only api-with-ci.

---

---

## F-INFRA-16 — `logback-spring.xml` does not redact `Authorization`, `X-API-Key`, `Cookie`, or JWT tokens from request/response logs

**Severity floor:** Medium
**Taxonomy:** SPRING-010, OWASP-A09, ASVS-V8-1, ASVS-V7-2, CWE-532

### Why it matters

SPRING-010 / ASVS-V8-1 / OWASP-A09 require redaction of well-known secret-bearing headers and parameters. The Trabuco template ships with a generic pattern and a hope that no developer ever logs sensitive data — the inverse of defense-in-depth. The `logback-spring.xml` is identical across all archetypes; none redacts `Authorization`, `X-API-Key`, `Set-Cookie`, or `apiKey` (the latter being literally the header used by `ApiKeyAuthFilter`).

### Suggested fix

Add a Logback `<replace>` filter or a custom `MaskingPatternLayout` that replaces matches of `(?i)(authorization|x-api-key|cookie|set-cookie|api[_-]?key)\s*[:=]\s*\S+` with `***REDACTED***` before encoding; configure `LogstashEncoder` `<excludeMdcKeyName>` for any MDC keys that hold tokens.

---

---

## F-INFRA-19 — `bucket4j-spring-boot-starter 0.12.7` is a 3rd-party non-Spring artifact pinned in parent POM; no SBOM, no SCA workflow

**Severity floor:** Medium
**Taxonomy:** MAVEN-004, OWASP-A06, CI-013, CWE-1104

### Why it matters

MAVEN-004 / CI-013 / OWASP-A06 require SCA in CI. Bucket4j, JobRunr (`8.4.0`), Resilience4j (`2.2.0`), Logstash-logback-encoder (`8.0`), Springdoc (`2.7.0`), Spring AI (`1.0.5`), and the OpenTelemetry instrumentation (`2.11.0` BOM with `1.45.0-alpha` and `1.29.0-alpha` semconv overrides) are all pinned but not monitored. **Alpha-versioned** dependencies (the OTel semconv `1.29.0-alpha` and `opentelemetry-api-incubator 1.45.0-alpha`) are explicitly opt-in to "things may break" — but no Renovate config refreshes them. CVE feeds for these libs would need to be checked manually.

### Suggested fix

Generate `.github/dependabot.yml` with weekly Maven updates and grouping; add an OWASP `dependency-check-maven` plugin in CI (skipped locally); add `osv-scanner` as a workflow step.

---

---

## F-INFRA-25 — `application.yml`'s `spring.profiles.active: local` default in every module — production deploys silently inherit `local` profile if `SPRING_PROFILES_ACTIVE` is unset

**Severity floor:** Medium
**Taxonomy:** ASVS-V14-1, OWASP-A05, SPRING-006, CWE-1188

### Why it matters

Defense-in-depth: a production deploy that "just works" because the local profile defaults are sane is a brittle accident. ASVS-V14-1 wants explicit profile separation. The current setup has no `application-prod.yml` template — operators must hand-craft prod config. Combined with the dormant-auth default (F-AUTH-01) and the JobRunr-dashboard-on (F-INFRA-05), a "deploy and forget" workflow yields a wide-open service.

### Suggested fix

Generate an `application-prod.yml` per module that sets the security-critical opposites (`trabuco.auth.enabled: true`, `JOBRUNR_DASHBOARD_ENABLED: false`, `SPRINGDOC_ENABLED: false`, `LOG_LEVEL: INFO`, `MANAGEMENT_ENDPOINTS: health,info`); fail fast on missing OIDC config when `prod` profile is active; document which profile is required.

---

## Summary

---

## F-WEB-07 — `server.compression.enabled=true` is on by default; `application/json` payloads compressed despite carrying secrets/PII (BREACH/CRIME class)

**Severity floor:** Medium
**Taxonomy:** OWASP-A02, API-API8, ASVS-V9-1, OWASP-A05

### Where to look

- `API/src/main/resources/application.yml:4-7` (api-minimal — identical text in crud-sql-worker, full-fat-all-modules)

### Evidence pattern

```yaml
server:
  port: ${SERVER_PORT:8080}
  shutdown: graceful
  compression:
    enabled: true
    mime-types: application/json,application/xml,text/html,text/xml,text/plain,application/javascript,text/css
    min-response-size: 1024
```

### Why it matters

BREACH/CRIME-style attacks exploit compression-length oracles: when a response contains attacker-influenced text *and* a server-issued secret (CSRF token, session ID, JWT in body, API key in error message), the size of the gzip'd response leaks information byte-by-byte. The Trabuco baseline ships with compression *on*, applies it to `application/json` — which carries every API response, including ProblemDetail bodies that may include `ex.getMessage()` (F-WEB-02) and `Map<String, Object>` echoes of request inputs (F-WEB-09). The AIAgent `/chat` endpoint reflects the user's request into the response; an attacker can mount a same-site BREACH oracle if cookies are ever introduced on the same origin (the templates are stateless today but the future migration target — sister oauth2-login services — would be vulnerable). Spring Boot's default is `enabled: false`; Trabuco *flips* it on.

### Suggested fix

Default `server.compression.enabled: false`; if compression is required for bandwidth, restrict to `text/css,application/javascript,text/html` (static assets only) and exclude `application/json`; or rely on a CDN/edge proxy where the secret-bearing response never gets compressed twice.

---

---

## F-WEB-08 — Swagger UI and OpenAPI spec served on `/swagger-ui.html` and `/api-docs` by default with no profile guard; default-permitted in the oauth2 chain

**Severity floor:** Medium
**Taxonomy:** OWASP-A05, SPRING-006, API-API9, CWE-1188, TRABUCO-008

### Where to look

- `API/src/main/resources/application.yml:163-172` (api-minimal — identical in other archetypes)

### Evidence pattern

```yaml
# application.yml:163-172
springdoc:
  api-docs:
    enabled: ${SPRINGDOC_ENABLED:true}
    path: /api-docs
  swagger-ui:
    enabled: ${SWAGGER_UI_ENABLED:true}
    path: /swagger-ui.html
```

### Why it matters

OpenAPI is enabled by *default* (`matchIfMissing` semantics via `${SPRINGDOC_ENABLED:true}`) and `permitAll()`'d in the oauth2 chain — even when the operator turns auth on, the spec endpoint stays public. Any future controller (admin endpoints, internal endpoints) is auto-published in the spec. There is no `paths-to-exclude` baseline for `/internal/**` or `/admin/**` — Trabuco's TRABUCO-008 check explicitly calls this out and the generated config does not satisfy it. In production this means: anyone reading `/api-docs` learns the entire surface (routes, parameter shapes, status codes), which materially shortens the attacker's recon. Combined with F-WEB-01 (no `@PreAuthorize`), the spec acts as a free authorization map of which endpoints are interesting.

### Suggested fix

Default `SPRINGDOC_ENABLED=false` and `SWAGGER_UI_ENABLED=false`; provide a `dev`/`local` profile override that flips them on; add `springdoc.paths-to-exclude=/internal/**,/admin/**` baseline; require `@PreAuthorize("hasAuthority('SCOPE_admin')")` on the spec endpoints when enabled in non-local profiles.

---

---

## F-WEB-09 — Controllers return raw `Map<String, Object>` and reflect request inputs back in responses — uncontrolled response shape, hard to audit for excessive data exposure

**Severity floor:** Medium
**Taxonomy:** API-API3, ASVS-V5-2, CWE-200

### Where to look

- `API/src/main/java/com/security/audit/api/controller/HealthController.java:21-27`

### Evidence pattern

```java
// PlaceholderJobController.java:48-50
return ResponseEntity.status(HttpStatus.ACCEPTED)
    .body(Map.of("status", "enqueued", "message", message));   // <-- echoes request input

// WebhookController.java:34-39
return Map.of(
    "webhook_id", reg.webhookId(),
    "url", reg.url(),                  // <-- echoes attacker-controlled URL
    "events", reg.events(),
    "status", reg.status()
);
```

### Why it matters

Two compounding effects:
1. **Reflection of request inputs** (`message`, `url`, `name`) in the response body widens the BREACH/CRIME oracle from F-WEB-07 — any caller can choose what gets compressed alongside any future server-issued secret.
2. **No typed contract** for the response shape — adding a sensitive field to a domain model later (e.g., `webhook.apiKeyHash`, `placeholder.ownerId`) flows into the response by accident if a builder picks it up. Immutable response DTOs (which `PlaceholderResponse` *does* use) are the antidote, but Trabuco emits them only for the placeholder CRUD path; everywhere else it's `Map.of(...)`.

### Suggested fix

Generate response DTOs in the same Immutables style as `PlaceholderResponse` for every controller; add an ArchUnit test that fails on `Map<...>` return types from `@RestController` methods; never reflect request inputs into the response — use server-generated correlation IDs instead.

---

---

## F-WEB-11 — AIAgent `WebConfig` registers `ScopeInterceptor` with no path filter and no order — runs on every request, including unmapped 404s, OPTIONS preflight, actuator, and `.well-known/agent.json`

**Severity floor:** Medium
**Taxonomy:** API-API8, ASVS-V13-1, OWASP-A05

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/config/WebConfig.java:23-26`

### Evidence pattern

```java
// WebConfig.java
@Override
public void addInterceptors(InterceptorRegistry registry) {
    registry.addInterceptor(scopeInterceptor);     // <-- no addPathPatterns / excludePathPatterns
}

// ScopeInterceptor.java
@Override
public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
    if (handler instanceof HandlerMethod method) {
        RequireScope annotation = method.getMethodAnnotation(RequireScope.class);
        if (annotation != null) {
            ScopeEnforcer.requireScope(annotation.value());
        }
    }
    return true;
}
```

### Why it matters

The interceptor itself fails *open* — `ScopeEnforcer.requireScope` only fires when a `@RequireScope` annotation is present (`if (annotation != null)`). That means *every* controller method must remember to annotate, and any new endpoint added later defaults to no scope check. The pattern is the inverse of Spring Security's "default deny": annotate to *require*, default to *allow*. Combined with F-WEB-01, an AIAgent controller author who forgets `@RequireScope` ships an endpoint reachable by anonymous callers (when `trabuco.auth.enabled=false`, the default). The interceptor also cannot run before the SecurityFilterChain (it's an MVC interceptor, after the dispatcher), so when JWT validation is on, scope decisions are split across two layers (`@PreAuthorize` for JWT, `@RequireScope` for legacy keys) with no warning when only one is wired — a controller carrying *only* `@RequireScope("partner")` and *no* `@PreAuthorize` runs *fully open* under the JWT chain (the JWT chain only checks `anyRequest().authenticated()`; the interceptor checks `@RequireScope` against `CallerContext` which is *empty* when the JWT path is in use).

### Suggested fix

Make `ScopeInterceptor` fail closed — require *some* annotation (`@RequireScope` or `@PermitAnonymous`) on every `HandlerMethod` and 500 if neither is present; merge the JWT and legacy-key authority models so a single `@PreAuthorize` works in both modes; add an `ArchUnit` test that every `@RestController` method declares its authorization stance.

---

---

## F-WEB-17 — `StreamingController` SSE emitter uses a 60s timeout but the `taskManager.subscribe` callback runs synchronously on every event with no per-tenant emitter cap; one partner can pin SSE connections to exhaust Tomcat threads

**Severity floor:** Medium
**Taxonomy:** API-API4, CWE-770, OWASP-A04

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/protocol/StreamingController.java:29-79`

### Evidence pattern

```java
@GetMapping("/tasks/{taskId}/stream")
@RequireScope("partner")
public SseEmitter streamTask(@PathVariable String taskId) {
    rateLimiter.checkRateLimit(CallerContext.get());     // per-call rate, not concurrent-emitter cap
    ...
    SseEmitter emitter = new SseEmitter(60_000L);        // 60s
    taskManager.subscribe(taskId, event -> { ... });
    return emitter;
}
```

### Why it matters

- BOLA: same as `tasks/get` in F-WEB-16 — partner-tier caller subscribes to *anyone's* task ID and watches the SSE feed.
- Resource exhaustion: 60s emitter timeouts × N concurrent calls × Tomcat connector limits → DoS budget. Virtual threads (which the YAML enables) mitigate carrier-thread exhaustion, but each emitter still pins a `SseEmitter` and a `taskManager.subscribe` listener, both reachable from `CallerContext` if not properly cleaned up on disconnect.
- Outbound SSE through a proxy without `X-Accel-Buffering: no` may be buffered, breaking the streaming contract silently.

### Suggested fix

Add an owner check (`task.ownerKeyHash != caller.keyHash()` → 404, not 403, to avoid task-existence enumeration); set a per-caller emitter limit; emit `Cache-Control: no-store` and `X-Accel-Buffering: no` headers; add an explicit `taskManager.unsubscribe(taskId, listener)` in `emitter.onCompletion`/`onTimeout` callbacks (the current code does not register them).

---

## False positives considered

- **`server.error.include-stacktrace` / `include-message` / `include-binding-errors` not configured.** Spring Boot 3 defaults are `never` for all three; `grep -rn "include-stacktrace\|include-message\|include-binding-errors"` across the audit tree finds zero matches, so the framework defaults stand. Not a finding.
- **`spring.mvc.problemdetails.enabled: true`** in the API yaml. This is correct — it activates RFC 7807 globally and matches the `GlobalExceptionHandler` design. Not a finding.
- **`management.endpoints.web.exposure.include: "*"`.** The default *is* a curated list `health,info,prometheus,metrics` (not `*`). `health.show-details: when_authorized` is also set. Actuator surface is acceptable. Not a finding (`/env`, `/heapdump`, `/threaddump`, `/loggers` are not exposed by default).
- **`management.info.env.enabled: false`.** Explicitly set to false; environment leakage via `/actuator/info` is closed. Not a finding.
- **`@Value`-based CSP `default-src 'self'`.** It is permissive but the API does not serve HTML; the policy is effectively a backstop for any future `text/html` response. Not a finding *per se* — covered indirectly by F-WEB-03 (AIAgent has no headers at all).
- **CSRF disabled.** Correct for stateless JWT/API-key flows; raised as low-severity F-WEB-13 only because of latent risk (Swagger UI, future oauth2-login).
- **`/actuator/prometheus` permitAll.** Metrics endpoints are conventionally scraped from a trusted network; they do not contain per-request data. Not flagged here, though operators should still front them with network policy.
- **`@Valid` is *present* on the API `PlaceholderController`.** Confirmed via inspection — `@Valid @RequestBody ImmutablePlaceholderRequest`. The DTO has `@NotBlank @Size` constraints. So the placeholder CRUD path is correct on the validation axis; the gap is on the typed-DTO axis only for the *other* controllers (F-WEB-06).
- **Default `MultipartConfig` without `max-file-size`.** No controller in the audited surface declares `MultipartFile` parameters; multipart is a non-issue here. (If document-upload is added later, F-WEB-05 captures the size cap concern at the JSON-body layer.)
- **Outbound `WebClient.create()` without TLS pinning.** Beyond the WebhookManager case (F-WEB-04, where it is the load-bearing issue), the codebase otherwise uses Spring AI's pre-configured clients — those are not `@RestController`-reachable directly. Not a separate web-layer finding.

---

## Summary

---

## F-INFRA-12 — `review-checks.sh` uses `set -u` only — missing `set -e -o pipefail`; subshell `find` failures are silently swallowed

**Severity floor:** Low
**Taxonomy:** CWE-754, CWE-755, CI-011 (defensive scripting)

### Where to look

`/tmp/trabuco-secaudit/api-with-ci/.github/scripts/review-checks.sh:16`

### Why it matters

A broken CI gate that silently passes is worse than no gate. The script is the deterministic-rules layer that mirrors the code-reviewer subagent (its raison d'être). If `git fetch` fails on PR builds (transient network, branch protection change), the gate runs in `--scope=all` mode unintentionally, potentially catching pre-existing violations and failing PRs unrelated to their changes — or running with empty file set and passing PRs that should fail.

### Suggested fix

`set -euo pipefail` at top; explicitly handle expected failures with `if ! cmd; then handle_error; fi` instead of `cmd || true`; emit a CI annotation when scope falls back to `all` because `diff` mode failed.

---

---

## F-INFRA-13 — `review-checks.sh` interpolates `$GITHUB_BASE_REF` directly into shell commands without sanitization

**Severity floor:** Low
**Taxonomy:** CI-011, CWE-78, CWE-88

### Where to look

`/tmp/trabuco-secaudit/api-with-ci/.github/scripts/review-checks.sh:41, 73 (CI side: ci.yml:73)`

### Evidence pattern

```yaml
# ci.yml:73
git fetch origin "${GITHUB_BASE_REF}":"refs/remotes/origin/${GITHUB_BASE_REF}" --quiet || true
```

### Why it matters

GitHub branch names theoretically allow some refspec-confusing characters; the current quoting prevents shell injection but not refspec confusion. Defensive CI scripts should validate `GITHUB_BASE_REF` matches `^[A-Za-z0-9._/-]+$` before use.

### Suggested fix

Add early validation: `[[ "$GITHUB_BASE_REF" =~ ^[A-Za-z0-9._/-]+$ ]] || { echo "invalid base ref"; exit 1; }`.

---

---

## F-INFRA-17 — `info.env.enabled: false` is set on API but **not** on AIAgent / Worker / EventConsumer — `/actuator/info` may leak environment variables

**Severity floor:** Low
**Taxonomy:** SPRING-001, OWASP-A05, CWE-200

### Why it matters

Inconsistent hardening across modules invites a future regression. Either every module asserts the value, or none of them do (and the team relies on Boot defaults). Mixed posture is the worst of both.

### Suggested fix

Move `management.info.env.enabled: false` and `management.info.java.enabled: true` (allow-list approach) into a `application-shared.yml` or a `shared/application.yml` profile fragment so every module inherits the same actuator hardening.

---

---

## F-INFRA-18 — `Dockerfile`s ship logback-config-source files (`*.xml`) and `application.yml` with potentially sensitive defaults baked into the layer; no `LOG_LEVEL` override path documented

**Severity floor:** Low
**Taxonomy:** CI-004 (informational), ASVS-V14-1

### Why it matters

ASVS-V14-1 wants different profiles per environment; the default-active profile being `local` rather than refusing to start without an explicit profile sets up production deployments to silently use dev settings.

### Suggested fix

Default `spring.profiles.active: ${SPRING_PROFILES_ACTIVE}` (no fallback) so the JVM fails fast with a missing-property error on prod; or set the Dockerfile `ENV SPRING_PROFILES_ACTIVE=prod` so containers inherit prod defaults; default `LOG_LEVEL` to `INFO`; ship a `Dockerfile.dev` for local with overrides.

---

---

## F-INFRA-20 — Spring AI alpha modules pinned (`opentelemetry-semconv 1.29.0-alpha`, `opentelemetry-api-incubator 1.45.0-alpha`) — explicitly unstable but treated as production deps

**Severity floor:** Low
**Taxonomy:** MAVEN-001, OWASP-A06, LLM-03

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/pom.xml:108-117` and `/tmp/trabuco-secaudit/api-minimal/pom.xml:92-101`

### Evidence pattern

```xml
<dependency>
    <groupId>io.opentelemetry.semconv</groupId>
    <artifactId>opentelemetry-semconv</artifactId>
    <version>1.29.0-alpha</version>
</dependency>
<dependency>
    <groupId>io.opentelemetry</groupId>
    <artifactId>opentelemetry-api-incubator</artifactId>
    <version>1.45.0-alpha</version>
</dependency>
```

### Why it matters

MAVEN-001 (pinned) is satisfied, but pinning to `-alpha` means accepting breaking-change risk. The comment says these are "alpha-version cycles that the instrumentation BOM does not pin" — the workaround is to pin them ourselves. Acceptable for now, but this is **infra-debt** that should be tracked: the moment the OTel BOM stabilizes these modules, they should move out of explicit `<dependency>` declarations to BOM-managed.

### Suggested fix

Add a comment noting the upgrade trigger ("remove this when otel-instrumentation-bom 3.x stabilizes semconv"); track an issue.

---

---

## F-INFRA-21 — `Dockerfile`'s `RUN mvn dependency:resolve ... 2>/dev/null || true` swallows resolution errors — silent partial-cache state

**Severity floor:** Low
**Taxonomy:** OWASP-A08, CI-013, CWE-754

### Why it matters

OWASP-A08 / supply chain wants deterministic dependency resolution. The `|| true` gives no signal that something was wrong.

### Suggested fix

Drop the `|| true`; rely on Docker layer caching (the `COPY pom.xml*` step + a properly failing `dependency:go-offline` is the canonical pattern).

---

---

## F-INFRA-22 — Spring AI `spring-ai-rag` dependency excludes `javax.validation` but does not pin `spring-ai-bom` consistently; mixed-version risk if a sub-module overrides

**Severity floor:** Low
**Taxonomy:** MAVEN-006, MAVEN-001, OWASP-A06, LLM-03

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/AIAgent/pom.xml:99-110`

### Why it matters

This is mostly OK — pin via BOM, exclude legacy. The minor concern is that the BOM is imported in parent dependencyManagement but a consumer module could override individual artifacts; a `mvn dependency:tree` discipline isn't enforced. Spring AI 1.0.5 has had advisories around prompt-injection in `RetrievalAugmentationAdvisor` (covered in AIAgent agent's findings) — keeping the BOM pinned to 1.0.5 means missing 1.0.x patches if/when they ship.

### Suggested fix

Confirm via `mvn dependency:tree -Dverbose` that no module overrides spring-ai versions; subscribe to https://github.com/spring-projects/spring-ai/security/advisories; auto-update via Dependabot once configured (F-INFRA-19).

---

---

## F-INFRA-23 — `enforcer:enforce` runs `dependencyConvergence` but explicitly **allows** legacy `javax.validation:validation-api` — JCache exception is fine; Bean Validation is a real risk

**Severity floor:** Low
**Taxonomy:** MAVEN-009, JAVA-001 (related), OWASP-A08

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/pom.xml:244-256` (and analogous in every parent POM)

### Evidence pattern

```xml
<bannedDependencies>
    <excludes>
        <exclude>javax.*:*</exclude>
        <exclude>junit:junit</exclude>
    </excludes>
    <includes>
        <include>javax.cache:cache-api</include>
        <include>javax.validation:validation-api</include>
    </includes>
</bannedDependencies>
```

### Why it matters

A legacy JAR on the classpath isn't always unused — when the Jakarta and javax artifacts both have `javax.validation.constraints.NotNull` and `jakarta.validation.constraints.NotNull` classes, downstream code can accidentally import the wrong one (IDE auto-import). Hibernate Validator 8.x ignores javax.* annotations entirely, so a misimported `javax.validation.constraints.NotNull` is **silently a no-op** — failing closed in security-sensitive validation paths. ASVS-V5-1 wants input validation; an annotation that compiles but doesn't validate is the worst kind of silent failure.

### Suggested fix

Use Maven `analyze` plugin to confirm `javax.validation` is genuinely unreferenced in compile-time imports; if confirmed unused, escalate the `<include>` to a TODO with a target date; consider an Spotbugs/Checkstyle rule banning `import javax.validation.*` from main code.

---

---

## F-INFRA-24 — `.dockerignore` excludes `.git/` but not `*.iml`/`.idea/`/test sources — image bloat and potential leak of IDE files

**Severity floor:** Low
**Taxonomy:** CI-004 (informational)

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/.dockerignore`

### Why it matters

Minor: build-context bloat slows CI and leaks test-only sources/configs into the build cache, which is shipped as an OCI layer when the build cache is shared.

### Suggested fix

Add `**/src/test/` to `.dockerignore` (with a comment noting that test compilation is skipped via `-DskipTests`); add `**/*.iml`, `.run/`.

---

---

## F-WEB-10 — `/health` controller mapped with bare `@RequestMapping("/health")` — no method whitelist, exposes TRACE/OPTIONS/HEAD; collides with actuator semantics

**Severity floor:** Low
**Taxonomy:** SPRING-015, API-API8, CWE-749, ASVS-V13-1

### Where to look

`API/src/main/java/com/security/audit/api/controller/HealthController.java:16-28` (api-minimal — identical in crud-sql-worker, full-fat-all-modules)

### Evidence pattern

```java
@RestController
@RequestMapping("/health")        // <-- no method= argument
public class HealthController {
    @GetMapping
    public ResponseEntity<Map<String, Object>> health() { ... }
}
```

### Why it matters

Two minor issues:
1. `@RequestMapping` without `method=` propagates whatever the inner `@GetMapping` declares, so this specific case is fine — but the *class-level* `@RequestMapping("/health")` advertises the path to Spring's mapping registry, and `RequestMappingHandlerMapping.allowedMethods` includes `OPTIONS`/`HEAD` automatically. Spring 6 still rejects `TRACE` by default unless `mvc.tomcat.allow-trace` is set, so this is mostly hygiene.
2. The path overlap with actuator (`/actuator/health/**` permitAll vs. `/health` not permitAll) is confusing at best; a load balancer pointed at `/health` after `trabuco.auth.enabled=true` returns 401, masking the failure mode as a real outage when the LB starts marking the host unhealthy.

### Suggested fix

Remove the redundant `HealthController` — the actuator already exposes `/actuator/health/{liveness,readiness}` and is permitAll'd. If a custom controller is wanted, use `@GetMapping("/health")` directly (no class-level `@RequestMapping`), and add `/health` to the permitAll list to match actuator.

---

---

## F-WEB-12 — RFC 7807 `type` URI uses unstable `urn:problem-type:*` placeholder — cannot be dereferenced, breaks RFC §3.1 ("a URI that, when dereferenced, provides documentation")

**Severity floor:** Low
**Taxonomy:** TRABUCO-002, API-API8 (compliance), ASVS-V7-1

### Where to look

- `API/src/main/java/com/security/audit/api/config/GlobalExceptionHandler.java` — every `setType(URI.create("urn:problem-type:*"))` call (api-minimal:59,79,106,123,141; equivalents in crud-sql-worker and full-fat).

### Why it matters

RFC 7807 §3.1 expects `type` to be a "URI reference \[RFC3986\] that … when dereferenced, it SHOULD provide human-readable documentation for the problem type." `urn:problem-type:forbidden` cannot be dereferenced; HTTP clients that follow `type` URIs (e.g., spec-driven SDK error-mapping libraries) get nothing useful. Worse, leaving the placeholder in production publicly advertises that the project shipped without taking the "replace this" path the comment requested — it is a "developer never finished hardening" tell. Severity is Low because no exploit follows directly, but it is a meaningful compliance gap and Trabuco's own checklist enumerates it as TRABUCO-002.

### Suggested fix

Source the `type` URI base from `${trabuco.problem.type-base:https://example.com/problems}` and 500-fail on boot if the value still ends in `example.com`. Or rename `urn:problem-type:*` to `https://docs.<service>.example.com/errors/<slug>` at template-emit time using the project's package name as a slug.

---

---

## F-WEB-13 — CSRF disabled in the permit-all chain (default) and the oauth2 chain — correct for stateless, but `OpenApiSecurityConfig` registers a `bearerAuth` scheme implying browser usage; no SameSite cookie hardening, no documented stance on browser-based callers

**Severity floor:** Low
**Taxonomy:** CWE-352, SPRING-008, ASVS-V13-1

### Where to look

- `API/src/main/java/com/security/audit/api/config/security/SecurityConfig.java:69, 98` — `csrf(AbstractHttpConfigurer::disable)` in both chains.

### Evidence pattern

```java
return http
    .csrf(AbstractHttpConfigurer::disable)
    .cors(cors -> {})
    .sessionManagement(s -> s.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
    ...
```

### Why it matters

The `csrf(disable)` decision is correct for the documented stateless+JWT+API-key story, but two latent risks:
1. The Swagger UI flow is browser-based and posts to the API. If the operator later adds an oauth2-login redirect for Swagger (a common request), the CSRF disable becomes incorrect *and silently exploitable* — the developer would have to remember to flip it back.
2. No baseline `server.servlet.session.cookie.same-site=Strict` / `secure=true` is set, so any accidental session creation issues a non-Secure, no-SameSite cookie. The oauth2 chain's `STATELESS` policy prevents Spring Security from creating sessions, but a controller that calls `request.getSession(true)` directly would not be blocked.

### Suggested fix

Add `server.servlet.session.cookie.{secure,same-site,http-only}=true,Strict,true` to `application.yml`; document the CSRF-off invariant in `SecurityConfig` Javadoc with a "do not flip back without re-enabling CSRF" warning; add a startup check that fails if any `HttpSession` bean is wired.

---

---

## F-WEB-14 — Default CORS `allowedOrigins` includes `http://localhost:8080` — same-origin as the API itself, irrelevant for browsers, and the comma-split string handling silently masks misconfiguration

**Severity floor:** Low
**Taxonomy:** SPRING-007, API-API8, OWASP-A05

### Where to look

- `API/src/main/java/com/security/audit/api/config/WebConfig.java:17-39` (api-minimal — identical in crud-sql-worker, full-fat-all-modules)

### Evidence pattern

```java
// API WebConfig.java:17, 33-39
@Value("${cors.allowed-origins:http://localhost:3000,http://localhost:8080}")
private String allowedOrigins;
...
registry.addMapping("/api/**")
    .allowedOriginPatterns(allowedOrigins.split(","))   // <-- pattern matcher
    ...
    .allowCredentials(allowCredentials)                  // false by default

// AIAgent WebConfig.java:30-35  -- different shape
registry.addMapping("/**")                               // <-- ALL paths, not just /api/**
    .allowedOrigins(allowedOrigins.split(","))           // <-- exact match (no patterns)
    .allowedMethods("GET","POST","PUT","DELETE","OPTIONS")
    .allowedHeaders("Content-Type","Authorization","X-Correlation-ID")
    .allowCredentials(false)
    .maxAge(3600);
```

### Why it matters

- The API uses `allowedOriginPatterns` (Spring's wildcard-aware variant) but the AIAgent uses `allowedOrigins` (exact match) — inconsistent across modules; a developer who copies the AIAgent config into the API silently loses pattern support.
- `http://localhost:8080` is the default *server* port; allowing CORS from the same origin is a nop for browsers (same-origin requests aren't CORS), but signals "we don't really know what we're allow-listing." More importantly, `allowedOrigins.split(",")` does not trim whitespace — `CORS_ALLOWED_ORIGINS="https://app.example.com, https://api.example.com"` (with a space) silently fails the second origin's match.
- The AIAgent variant maps `/**` while the API maps `/api/**` — the AIAgent therefore enables CORS on `/.well-known/agent.json`, `/actuator/*`, and `/mcp` (when MCP server is enabled). This is moot for `allowCredentials=false`, but the moment the operator flips `CORS_ALLOW_CREDENTIALS=true` for, say, a partner UI, the actuator becomes credentials-eligible cross-origin.

### Suggested fix

Drop `localhost:8080` from defaults; in `WebConfig`, trim whitespace from each split element; in the AIAgent config, restrict `addMapping` to the explicit business path prefixes (`/ask`, `/chat`, `/ingest/**`, `/webhooks/**`, `/tasks/**`); use the same `allowedOriginPatterns` API in both modules.

---

---

## F-WEB-15 — `AgentExceptionHandler` does not handle `MethodArgumentNotValidException` / `MissingServletRequestParameterException` despite extending `ResponseEntityExceptionHandler` — `Map<...>` fallthrough leaks the framework's default response shape

**Severity floor:** Low
**Taxonomy:** ASVS-V7-1, ASVS-V5-1, CWE-200

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/config/AgentExceptionHandler.java:51-127`

### Why it matters

Less of a leak, more of a coverage gap — `@Validated @PathVariable String webhookId` triggering a constraint violation falls into the catch-all `Exception` handler in the AIAgent and emits `"An unexpected error occurred."` with status 500, instead of the API module's helpful 400 with the violations map. The user can't tell why their request failed; debug-time noise is higher; HTTP status semantics are wrong.

### Suggested fix

Mirror the API's `handleConstraintViolation` in `AgentExceptionHandler`; add coverage for `ResponseStatusException` reasons that should *not* be reflected (cross-ref F-WEB-02 — split into "safe message" vs "leaked message" classes).

---

---

