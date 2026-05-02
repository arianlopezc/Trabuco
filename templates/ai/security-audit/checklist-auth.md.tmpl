# Trabuco Security Audit — Auth Domain

Authentication, authorization, identity propagation, scope enforcement, session/CSRF, JWT validation, API-key handling, and rate limiting.

This file is the **detail reference** for the
`trabuco-security-audit-auth` specialist subagent. The orchestrator
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

**Total checks in this domain: 28**
(8 Critical,
 8 High,
 8 Medium,
 3 Low,
 1 Informational)

---

## F-AUTH-01 — Default `permitAllFilterChain` ships open by default; `matchIfMissing=true` makes "no auth" the silent factory state

**Severity floor:** Critical
**Taxonomy:** OWASP-A05, ASVS-V4-1, API-API2, CWE-1188, TRABUCO-001

### Where to look

`API/src/main/java/com/security/audit/api/config/security/SecurityConfig.java:94-103` (and identical `AIAgent/.../AgentSecurityConfig.java:105-113`)

### Evidence pattern

```java
@Bean
@ConditionalOnProperty(value = "trabuco.auth.enabled", havingValue = "false", matchIfMissing = true)
public SecurityFilterChain permitAllFilterChain(HttpSecurity http) throws Exception {
    return http
        .csrf(AbstractHttpConfigurer::disable)
        .cors(cors -> {})
        .sessionManagement(s -> s.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
        .authorizeHttpRequests(auth -> auth.anyRequest().permitAll())
        .build();
}
```

### Why it matters

The default-emitted `SecurityFilterChain` is `anyRequest().permitAll()` with CSRF disabled, and is selected when `trabuco.auth.enabled` is missing — i.e., out-of-the-box. A developer who runs `mvn spring-boot:run` and immediately deploys gets a service with no authentication, no authorization, and no CSRF protection on every endpoint, including state-changing ones. The "ship dormant" pattern violates ASVS V4.1 default-deny: dormant auth code is not the same as default-deny posture.

### Suggested fix

Make the secured chain the default (`havingValue = "true", matchIfMissing = true`), or fail fast on startup unless the operator explicitly opts into `trabuco.auth.enabled=false` for local dev.

---

---

## F-AUTH-02 — JWT audience validation never enforced — empty default `${OIDC_AUDIENCE:}` silently disables aud claim check

**Severity floor:** Critical
**Taxonomy:** OWASP-A07, ASVS-V2-1, API-API2, CWE-287, CWE-345

### Where to look

`API/src/main/resources/application.yml:53-58` (and `AIAgent/.../application.yml:37-42`)

### Evidence pattern

```yaml
spring:
  security:
    oauth2:
      resourceserver:
        jwt:
          issuer-uri: ${OIDC_ISSUER_URI:}
          audiences: ${OIDC_AUDIENCE:}
```

### Why it matters

The `audiences:` property defaults to an empty string. Spring Boot's `OAuth2ResourceServerProperties.Jwt.audiences` is a `List<String>`; an empty string yields either an empty list or `[""]`, neither of which adds a meaningful audience validator. Combined with `SecurityIntegrationTest`/`AuthEndToEndTest` (`SignedJwtTestSupport.java:96-118`) which mint tokens with NO `aud` claim and pass — the test suite gives false confidence that audience binding is in place. Any token intended for a sister service (signed by the same IdP, different audience) will be accepted. Token confusion / cross-service replay is unprotected.

### Suggested fix

Refuse to start when `trabuco.auth.enabled=true` and `OIDC_AUDIENCE` is empty; add an explicit `JwtClaimValidator<List<String>>("aud", aud -> aud != null && aud.contains(expected))` and an `AuthEndToEndTest#wrongAudience_returns401` regression.

---

---

## F-AUTH-03 — Hardcoded API keys in source (`public-read-key`, `partner-secret-key`) — both committed to VCS

**Severity floor:** Critical
**Taxonomy:** OWASP-A07, ASVS-V2-2, ASVS-V6-1, CWE-798, CWE-256

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/security/ApiKeyAuthFilter.java:44-49`

### Evidence pattern

```java
// TODO: Replace with a configuration bean or externalized config for production use.
// These are placeholder keys for development only.
private static final Map<String, KeyEntry> API_KEYS = Map.of(
    "public-read-key", new KeyEntry("public", "public-read-key"),
    "partner-secret-key", new KeyEntry("partner", "partner-secret-key")
);
```

### Why it matters

Two human-readable keys are baked into the compiled jar and into every developer's git history. The filter is `@ConditionalOnProperty(... matchIfMissing = true)` so it is **on by default**. A freshly generated AIAgent service accepts `Authorization: Bearer partner-secret-key` from any internet caller and grants the `partner` tier — which in `IngestionController.java`, `WebhookController.java`, `StreamingController.java` is the maximum tier and authorizes RAG ingestion, webhook registration to arbitrary URLs, and SSE streaming. The "TODO" comment does not block production deployment.

### Suggested fix

Remove the `Map.of(...)` literal entirely; require the operator to inject a `@ConfigurationProperties("agent.auth")` bean populated from environment / secret manager; refuse to boot the filter if no keys are configured.

---

---

## F-AUTH-08 — `IngestionController` and other AIAgent controllers rely on `@RequireScope` — bypassed entirely if `app.aiagent.api-key.enabled=false`

**Severity floor:** Critical
**Taxonomy:** OWASP-A01, API-API5, ASVS-V4-2, CWE-862, TRABUCO-001, TRABUCO-005

### Where to look

`AIAgent/.../protocol/IngestionController.java:64-78`, `AIAgent/.../protocol/StreamingController.java:30-31`, `AIAgent/.../protocol/WebhookController.java:28-43`

### Evidence pattern

```java
@PostMapping
@RequireScope("partner")
public ResponseEntity<Map<String, Object>> ingest(@RequestBody IngestRequest req) {
    ingestionService.ingest(req.text(), req.metadata());
    ...
}
```

### Why it matters

`@RequireScope` is enforced by `ScopeInterceptor` reading `CallerContext`, populated only by `ApiKeyAuthFilter`. If an operator sets `app.aiagent.api-key.enabled=false` (the documented "JWT-only target state" per AgentSecurityConfig.java:42), `ApiKeyAuthFilter` is removed, `CallerContext` is always anonymous, `@RequireScope("partner")` will reject every request — but the operator is supposed to "replace each `@RequireScope` with `@PreAuthorize` ... pick one" (per IngestionController Javadoc lines 37-43). This is a manual, error-prone migration with NO compile-time check and NO test coverage. A partial migration leaves admin-only endpoints (`/ingest`, `/webhooks`) callable by anonymous users (when JWT is on but the controller author forgot to swap the annotation, the request is authenticated by JWT but ScopeInterceptor sees anonymous CallerContext and either denies or — depending on misconfig — allows). This violates default-deny.

### Suggested fix

Replace `@RequireScope` with `@PreAuthorize("hasAuthority(...)")` everywhere; have `ScopeInterceptor` derive `CallerIdentity` from `SecurityContextHolder` so JWT and API-key paths feed the same enforcement; or fail loudly at startup when both auth modes are misaligned.

---

---

## F-AUTH-10 — A2A endpoint `/a2a` accepts all comers — no JSON-RPC-level authentication, no mutual auth between agents

**Severity floor:** Critical
**Taxonomy:** OWASP-A01, API-API1, API-API5, TRABUCO-003, CWE-287, LLM-06

### Where to look

`AIAgent/.../protocol/A2AController.java:43-91`

### Evidence pattern

```java
@PostMapping("/a2a")
public JsonRpcResponse handle(@RequestBody JsonRpcRequest body) {
    var caller = CallerContext.get();
    rateLimiter.checkRateLimit(caller);
    ...
    return switch (method) {
        case "tasks/send" -> handleTasksSend(rpcId, params, caller);
        case "tasks/chat" -> handleTasksChat(rpcId, params, caller);
        case "tasks/get" -> handleTasksGet(rpcId, params);
        ...
    };
}
```

### Why it matters

No class-level/method-level `@RequireScope` or `@PreAuthorize` on `/a2a`. `tasks/get` doesn't even pass a caller into the handler — a remote agent (or anonymous internet caller) can poll arbitrary `task_id` values by guessing UUIDs and read other tenants' task results (BOLA + cross-tenant leak). Inside `tasks/send` only `ask_question` checks `ScopeEnforcer.requireScope("public", caller)` — but `caller` is from `CallerContext.get()` which yields anonymous when no API-key chain ran (i.e., the JWT-only mode). A2A is documented in TRABUCO-003 as needing mutual auth; there is none.

### Suggested fix

Add `@PreAuthorize("hasAuthority('SCOPE_a2a:invoke')")` at the class level; require an mTLS/agent-token pair; enforce scope on `tasks/get` and verify `caller` owns the queried task.

---

---

## F-AUTH-17 — `PlaceholderController` (full-fat) has no `@PreAuthorize` and no ownership check on `getById(id)` — BOLA template

**Severity floor:** Critical
**Taxonomy:** OWASP-A01, API-API1, ASVS-V4-2, CWE-639, CWE-862

### Where to look

`full-fat-all-modules/API/.../controller/PlaceholderController.java:42-52` and `:86-92`

### Evidence pattern

```java
@GetMapping("/{id}")
public ResponseEntity<ImmutablePlaceholderResponse> getById(@PathVariable Long id) {
  return service.findById(id)
    .map(p -> ResponseEntity.ok(ImmutablePlaceholderResponse.builder()
      ...
@DeleteMapping("/{id}")
public ResponseEntity<Void> delete(@PathVariable Long id) {
  if (service.delete(id)) {
    return ResponseEntity.noContent().build();
  }
  ...
}
```

### Why it matters

Even when `trabuco.auth.enabled=true`, this controller emits no `@PreAuthorize`, performs no caller-vs-resource ownership check, and exposes raw `Long id` enumerable IDs. A user with the lowest valid scope can read/update/delete any placeholder across all tenants. This is the canonical BOLA pattern, generated by Trabuco as the **starter example** developers will copy when building real CRUD. The placeholder ships without `@PreAuthorize` even though `@EnableMethodSecurity` is on, so developers learn "no annotations needed" and replicate the bug.

### Suggested fix

Generate placeholder with `@PreAuthorize("hasAuthority('SCOPE_*')")` and an explicit owner predicate (e.g., `@PostAuthorize("returnObject?.body?.tenantId == authentication.principal.tenantId")`); add a comment "// SECURITY: replace with your tenant predicate".

---

---

## F-AUTH-18 — `EventController` and `PlaceholderJobController` (full-fat) have no auth — accept any payload and enqueue work, even when JWT chain is active

**Severity floor:** Critical
**Taxonomy:** OWASP-A01, API-API5, ASVS-V4-2, TRABUCO-007

### Where to look

`full-fat-all-modules/API/.../controller/EventController.java:53-70`, `full-fat-all-modules/API/.../controller/PlaceholderJobController.java:44-73`

### Evidence pattern

```java
@PostMapping("/placeholder")
public ResponseEntity<Map<String, String>> publishPlaceholderEvent(
    @RequestBody Map<String, String> request) {
    ...
    eventPublisher.publish(event);
    ...
}
```

### Why it matters

Both controllers POST onto async work pipelines (broker / JobRunr) with no `@PreAuthorize`. A JWT-authenticated user with the lowest scope (or anonymous, when permit-all is active) can flood the event bus, enqueue arbitrary jobs (delaySeconds is unbounded — see `schedule(...)`), and exhaust Worker/EventConsumer capacity. Combined with F-AUTH-15 (no identity propagation), the receiving handlers can't even tell who triggered the flood.

### Suggested fix

Add `@PreAuthorize("hasAuthority('SCOPE_jobs:write')")` and `SCOPE_events:publish`; cap `delaySeconds`; record originating identity in the event/job payload.

---

---

## F-AUTH-23 — `tasks/get` in A2AController returns task results without verifying caller owns the task — task IDs are UUIDs but enumerable

**Severity floor:** Critical
**Taxonomy:** OWASP-A01, API-API1, CWE-639, CWE-200

### Where to look

`AIAgent/.../protocol/A2AController.java:111-126`

### Evidence pattern

```java
private JsonRpcResponse handleTasksGet(String rpcId, Map<String, Object> params) {
    String taskId = (String) params.get("task_id");
    ...
    var task = taskManager.getTask(taskId);
    ...
    return JsonRpcResponse.success(rpcId, Map.of(
        "task_id", task.getTaskId(),
        "status", task.getStatus(),
        "result", task.getResult() != null ? task.getResult() : "",
        ...
    ));
}
```

### Why it matters

`handleTasksGet` does not receive a `caller`, does not consult `CallerContext`, and does not check that `task.callerKeyHash == caller.keyHash`. Any caller (anonymous in the default config — see F-AUTH-01) can poll any task ID and read another caller's chat output, which may include sensitive RAG-retrieved content. UUIDs are not authorization. This is the textbook BOLA pattern across tenant boundaries.

### Suggested fix

Tag tasks with caller identity at submission; reject `tasks/get` when caller does not match task owner; enforce `@PreAuthorize` on the method.

---

---

## F-AUTH-04 — `application.yml`'s `agent.auth.keys` config is dead — filter ignores it

**Severity floor:** High
**Taxonomy:** OWASP-A05, ASVS-V14-1, CWE-1188

### Where to look

`AIAgent/src/main/resources/application.yml:117-125` versus `AIAgent/.../ApiKeyAuthFilter.java:46-49`

### Evidence pattern

```yaml
agent:
  auth:
    keys:
      public-read-key:
        tier: public
        label: public-read-key
      partner-secret-key:
        tier: partner
        label: partner-secret-key
```

### Why it matters

No `@ConfigurationProperties` class binds `agent.auth.keys`; no bean reads it. An operator who edits this YAML to rotate keys will see zero behavioural change — the hardcoded `Map.of(...)` in the filter is what's actually consulted. This silently breaks key rotation and creates a false sense of "I changed the config so I'm safe". Combine with F-AUTH-03 and rotation is impossible without a code edit + redeploy.

### Suggested fix

Either delete the YAML stanza or wire a `@ConfigurationProperties("agent.auth")` and load `API_KEYS` from it at startup.

---

---

## F-AUTH-05 — API-key comparison via `Map.get(plaintext)` — hash-table lookup, not constant-time comparison

**Severity floor:** High
**Taxonomy:** OWASP-A07, ASVS-V2-2, CWE-208 (timing), CWE-916

### Where to look

`AIAgent/.../ApiKeyAuthFilter.java:70-76`

### Evidence pattern

```java
String key = authHeader.substring(7).trim();
KeyEntry entry = API_KEYS.get(key);

if (entry == null) {
    sendError(response, HttpStatus.FORBIDDEN, "Invalid API key");
    return;
}
```

### Why it matters

`HashMap.get` uses `String.hashCode()` and `String.equals()`, both of which short-circuit on first character mismatch. Combined with stable hash seeding inside a single JVM, this leaks information through timing about prefix matches and bucket distribution. Even though the keys are also stored in plaintext in the JAR (F-AUTH-03), a properly engineered API-key path would compute `MessageDigest.isEqual(sha256(supplied), sha256(stored))` on every candidate or use an HMAC over a server secret. Additionally, keys are stored plaintext at rest (in source); the `sha256()` helper at line 96-106 is computed AFTER lookup and used only for the log/identity hash, never for comparison.

### Suggested fix

Hash supplied keys, look up by hash, and use `MessageDigest.isEqual` for the final comparison.

---

---

## F-AUTH-06 — `ApiKeyAuthFilter` is a `@Component` bean, not registered with `HttpSecurity` — runs OUTSIDE the SecurityFilterChain

**Severity floor:** High
**Taxonomy:** OWASP-A01, API-API2, ASVS-V4-1, CWE-285

### Where to look

`AIAgent/.../ApiKeyAuthFilter.java:40-42` and `AIAgent/.../config/security/AgentSecurityConfig.java:80-95`

### Evidence pattern

```java
@Component
@ConditionalOnProperty(name = "app.aiagent.api-key.enabled", havingValue = "true", matchIfMissing = true)
public class ApiKeyAuthFilter extends OncePerRequestFilter {
```

### Why it matters

Spring Boot auto-registers any `OncePerRequestFilter` `@Component` as a generic servlet filter via `FilterRegistrationBean`. It does NOT participate in the Spring Security `SecurityFilterChain`. When `trabuco.auth.enabled=true` (the JWT chain), the JWT-AuthenticationEntryPoint is triggered for unauthenticated requests **before** the API-key filter sets `CallerContext` — meaning the documented "hybrid mode" (both creds accepted) does not actually accept API keys when JWT chain is active. Conversely, when the permit-all chain is active, ApiKeyAuthFilter runs but produces no `Authentication` in `SecurityContextHolder`, so `@PreAuthorize` checks on the JWT path won't see the API-key tier as an authority. The "matrix" in `AgentSecurityConfig`'s Javadoc (lines 36-44) is aspirational, not actual.

### Suggested fix

Either (a) register `ApiKeyAuthFilter` via `http.addFilterBefore(...)` inside the agentOauth2FilterChain when both flags are on and have it set a real `Authentication`, or (b) explicitly document and enforce mutual exclusion via a fail-fast `@PostConstruct` check.

---

---

## F-AUTH-07 — `actuator/prometheus` is `permitAll()` even when JWT chain is active

**Severity floor:** High
**Taxonomy:** OWASP-A05, SPRING-001, CWE-200, ASVS-V8-1

### Where to look

`API/src/main/java/com/security/audit/api/config/security/SecurityConfig.java:76-86`

### Evidence pattern

```java
.authorizeHttpRequests(auth -> auth
    .requestMatchers(
        "/actuator/health/**",
        "/actuator/info",
        "/actuator/prometheus",
        "/swagger-ui/**",
        "/v3/api-docs/**",
        "/api-docs/**",
        "/error"
    ).permitAll()
    .anyRequest().authenticated())
```

### Why it matters

`/actuator/prometheus` is exposed via `MANAGEMENT_ENDPOINTS` defaulting to `health,info,prometheus,metrics` (application.yml:135). The JWT chain explicitly permits scrape access without auth — leaking memory pressure, request rates, error rates, JVM heap sizes, datasource pool occupancy, and circuit breaker state to any unauthenticated caller. Same issue for `swagger-ui/**` and `/v3/api-docs/**` — production endpoints (and any operations annotated `@Hidden`'d via `@SecurityRequirement`) become publicly enumerable.

### Suggested fix

Drop `/actuator/prometheus`, `/swagger-ui/**`, `/api-docs/**`, `/v3/api-docs/**` from `permitAll()`; gate behind a separate `MANAGEMENT_PORT` chain or require basic-auth/scrape token.

---

---

## F-AUTH-09 — `ScopeEnforcer` uses string-based tier comparison with hardcoded ladder; `tierLevel("public")=1`, `tierLevel("partner")=2`, all unknowns silently `0`

**Severity floor:** High
**Taxonomy:** OWASP-A01, API-API5, ASVS-V4-2, CWE-863, CWE-285

### Where to look

`AIAgent/.../security/CallerIdentity.java:22-32` and `ScopeEnforcer.java:10-17`

### Evidence pattern

```java
default boolean isAtLeast(String requiredTier) {
    return tierLevel(tier()) >= tierLevel(requiredTier);
}
private static int tierLevel(String tier) {
    return switch (tier) {
        case "partner" -> 2;
        case "public" -> 1;
        default -> 0;
    };
}
```

### Why it matters

`requireScope("admin", caller)` resolves `tierLevel("admin") -> 0`; for any caller whose tier is `"public"` or `"partner"`, `tierLevel(caller.tier()) >= 0` is true, so **any "admin" requirement is satisfied** by any non-anonymous caller. Misuse vector: a developer adds `@RequireScope("admin")` to a destructive operation, expecting admin-only access; everyone with `partner-secret-key` can call it. There is no compile-time enum, no validator, no test asserting `requireScope("admin")` rejects partner.

### Suggested fix

Use an `enum Tier { ANONYMOUS, PUBLIC, PARTNER, ADMIN }`; fail fast at startup if `@RequireScope` references an unknown value; add tests for each tier × required pair.

---

---

## F-AUTH-11 — `RateLimiter` keys windows by `keyHash` — anonymous callers SHARE one window across the whole internet

**Severity floor:** High
**Taxonomy:** OWASP-A04, API-API4, ASVS-V2-3, CWE-770

### Where to look

`AIAgent/.../security/RateLimiter.java:25-42` cross `CallerIdentity.anonymous()` (`CallerIdentity.java:34-40`)

### Evidence pattern

```java
static CallerIdentity anonymous() {
    return ImmutableCallerIdentity.builder()
        .tier("anonymous")
        .keyHash("anonymous")
        ...
}
// RateLimiter
Deque<Long> window = store.computeIfAbsent(caller.keyHash(), k -> new ConcurrentLinkedDeque<>());
```

### Why it matters

Every anonymous request shares `keyHash="anonymous"` and therefore shares a single 60-second / 10-request bucket. One bad actor exhausts the limit for every other anonymous caller (DoS amplification and a deliberate "lockout" attack on legitimate anonymous traffic). Worse, the entire internet is on one bucket → at scale, legitimate traffic is permanently throttled. This is also a brute-force protection failure (ASVS V2.3): an attacker probing the API-key list can never be tracked per-IP.

### Suggested fix

For anonymous identities, key the window by the source IP (`request.getRemoteAddr()` or `X-Forwarded-For` after trusted-proxy validation); for authenticated identities, key by sub/keyHash.

---

---

## F-AUTH-15 — `AuthScope` documented as the safe set/clear pattern but Worker/EventConsumer modules in `full-fat-all-modules` do NOT call it — propagation is unwired

**Severity floor:** High
**Taxonomy:** OWASP-A01, ASVS-V4-2, CWE-862, TRABUCO-001

### Where to look

`Shared/.../auth/AuthScope.java:36-63` plus absence in `Worker/src/main/java/...`, `EventConsumer/src/main/java/...`

### Evidence pattern

```bash
$ grep -rn "AuthScope\|RequestContextHolder\|IdentityClaims\|AuthContextPropagator" \
    /tmp/trabuco-secaudit/full-fat-all-modules/Worker/ \
    /tmp/trabuco-secaudit/full-fat-all-modules/EventConsumer/ | grep -v test
(no output)
```

### Why it matters

The Shared module ships `AuthContextPropagator`, `DefaultAuthContextPropagator`, `AuthScope`, and a documented contract for serializing identity into broker headers. None of the generated EventConsumer listeners (Kafka/RabbitMQ/SQS/Pub-Sub) or Worker JobRunr handlers actually call these utilities. So when an authenticated user enqueues a job through `PlaceholderJobController`, the Worker handler runs with anonymous identity; downstream services that read `RequestContextHolder` get anonymous claims, masking who triggered the work and bypassing per-tenant authorization in service-layer code. The "identity flows through async boundaries" promise (AuthContextPropagator.java:9-16) is documented but unwired.

### Suggested fix

Generate Worker/EventConsumer listener templates that wrap handler entry points with `try (var scope = AuthScope.set(propagator.fromHeaders(brokerHeaders).orElse(IdentityClaims.anonymous()))) { ... }`.

---

---

## F-AUTH-25 — `IngestionController` accepts `Map<String, Object>` metadata directly into vector-store payload — no schema, no key allowlist, mass-assignment / prompt-injection vector

**Severity floor:** High
**Taxonomy:** API-API3, ASVS-V5-1, LLM-04, TRABUCO-005

### Where to look

`AIAgent/.../protocol/IngestionController.java:64-95` and downstream `DocumentIngestionService.java:50-57`

### Evidence pattern

```java
public record IngestRequest(String text, Map<String, Object> metadata) {}
...
ingestionService.ingest(req.text(), req.metadata());
```

### Why it matters

Beyond auth-only — but this is in scope for the auth domain because `@RequireScope("partner")` is the *only* gate, and once past that gate the caller can stuff arbitrary fields into vector metadata. With an attacker-controlled `tenant_id` field in metadata, a partner-tier caller can write documents tagged as a different tenant's, then query them back in the JWT-scoped retrieval. Tenant isolation in the vector store relies on metadata predicates that this endpoint allows the attacker to forge.

### Suggested fix

Use a constrained DTO with a fixed key set; never let request-supplied metadata write `tenant_id` / `user_id` / `authority` keys.

---

---

## F-AUTH-12 — `permitAllFilterChain` disables CSRF for stateless mode but still permits cookie-bearing requests via CORS — no defense if any future endpoint becomes stateful

**Severity floor:** Medium
**Taxonomy:** SPRING-008, CWE-352, OWASP-A05

### Where to look

`API/.../SecurityConfig.java:96-103`, `AIAgent/.../AgentSecurityConfig.java:107-113`

### Evidence pattern

```java
return http
    .csrf(AbstractHttpConfigurer::disable)
    ...
    .authorizeHttpRequests(auth -> auth.anyRequest().permitAll())
    .build();
```

### Why it matters

CSRF disabled globally on the permit-all chain is fine *only* while the app is purely stateless. The moment a developer adds session state or cookie auth and forgets to flip `trabuco.auth.enabled=true`, the open chain is still active and CSRF is off. The dormant configuration encourages "I'll deal with CSRF when I add auth" thinking; ASVS V4 demands defence in depth.

### Suggested fix

Keep CSRF off only inside the JWT chain; on the open chain, prefer Spring's defaults (CSRF on, with explicit ignored matchers).

---

---

## F-AUTH-13 — JWT signature algorithm not constrained — accepts whatever `JwtDecoders.fromIssuerLocation` allows; no `JwsAlgorithm` whitelist

**Severity floor:** Medium
**Taxonomy:** OWASP-A07, ASVS-V2-1, CWE-345, CWE-287

### Where to look

`API/.../SecurityConfig.java:67-87` (no algorithm restriction passed to `.jwt(...)`); `AIAgent/.../AgentSecurityConfig.java:80-95`

### Evidence pattern

```java
.oauth2ResourceServer(oauth2 -> oauth2
    .jwt(jwt -> jwt.jwtAuthenticationConverter(jwtConverter))
    .authenticationEntryPoint(authProblemHandler)
    .accessDeniedHandler(authProblemHandler))
```

### Why it matters

`NimbusJwtDecoder` derived from `issuer-uri` accepts whatever algorithms the JWKS advertises. If the IdP rotates its JWKS to include an HS256 entry, the decoder will accept HS256 tokens — opening symmetric-key attacks against an issuer that originally did RS256 only. Best practice: pin to a list (e.g., RS256/ES256/EdDSA) and reject `none`/`HS*` explicitly. There is no test (`SignedJwtTestSupport.java` only generates RS256) for algorithm-confusion attacks.

### Suggested fix

Configure `JwtDecoder` with `.jwsAlgorithms(algs -> algs.add(SignatureAlgorithm.RS256))` or set `spring.security.oauth2.resourceserver.jwt.jws-algorithms: [RS256, ES256]`.

---

---

## F-AUTH-14 — `RequestContextHolder` uses raw `ThreadLocal` with documented "always clear in finally" — no filter actually does this in JWT path

**Severity floor:** Medium
**Taxonomy:** ASVS-V8-1, CWE-200, JAVA-013, OWASP-A09

### Where to look

`Shared/.../auth/RequestContextHolder.java:23-49` and `API/.../security/JwtAuthenticationConverter.java:39-48`

### Evidence pattern

```java
// JwtAuthenticationConverter - note: NO finally{ clear() }
@Override
public AbstractAuthenticationToken convert(Jwt jwt) {
    IdentityClaims claims = extractor.extract(jwt);
    RequestContextHolder.set(claims);            // sets ThreadLocal
    ...
    return new JwtAuthenticationToken(jwt, authorities, claims.subject());
}
```

### Why it matters

The converter writes to `RequestContextHolder` at JWT validation time but no filter pairs that with `RequestContextHolder.clear()` in `finally`. With virtual threads enabled (`spring.threads.virtual.enabled=true` in application.yml) and Tomcat carrier-thread reuse, identity from request N can leak into request N+1 when the virtual thread that ran request N is parked and a new virtual thread mounts the same carrier. The Javadoc explicitly warns about this at lines 14-16 — but the code does not implement what the Javadoc demands. `ApiKeyAuthFilter` does clear `CallerContext` in a finally (line 86), but the equivalent for JWT path / `RequestContextHolder` is missing.

### Suggested fix

Add a Spring Security `OncePerRequestFilter` or `HandlerInterceptor` that wraps the request and calls `RequestContextHolder.clear()` (and `MDC.clear()`) in `finally`.

---

---

## F-AUTH-16 — `JwtAuthenticationConverter` derives authorities only from `scope`/`scp` claim — ignores `roles`, `groups`, `realm_access.roles` (Keycloak), `cognito:groups` (Cognito)

**Severity floor:** Medium
**Taxonomy:** OWASP-A07, ASVS-V4-2, CWE-285

### Where to look

`API/.../security/JwtAuthenticationConverter.java:39-48` + `Shared/.../auth/DefaultJwtClaimsExtractor.java:39-65`

### Evidence pattern

```java
// DefaultJwtClaimsExtractor only reads "scope" or "scp"
private Set<String> scopes(Jwt jwt) {
    Object raw = jwt.getClaim("scope");
    if (raw == null) {
        raw = jwt.getClaim("scp");
    }
    ...
}
```

### Why it matters

Real-world deployments using Keycloak put roles in `realm_access.roles`; Cognito uses `cognito:groups`; Auth0 uses namespaced claims; some IdPs combine scopes and roles. The default extractor silently produces an empty authority set for these tokens, so `@PreAuthorize("hasAuthority('SCOPE_admin')")` fails-closed — but a developer who notices "no auth working" may flip `permitAll()` rather than write a custom extractor (and the `application.yml` doc at lines 51-52 says "Auth0 and Cognito additionally need a custom JwtClaimsExtractor bean" — easy to miss). The likely outcome under load is misconfiguration that makes endpoints public.

### Suggested fix

Provide opt-in extractors for Keycloak / Cognito / Auth0; fail loudly at startup if a token validates but produces no authorities and `trabuco.auth.enabled=true`.

---

---

## F-AUTH-22 — `agentPermitAllFilterChain` configures `cors()` only on the API SecurityConfig, NOT on the AIAgent agent chain — AIAgent CORS is purely from `WebConfig` (MVC layer), bypassed when permit-all chain accepts a CORS preflight

**Severity floor:** Medium
**Taxonomy:** OWASP-A05, API-API8, SPRING-007, CWE-346

### Where to look

`AIAgent/.../config/security/AgentSecurityConfig.java:80-113` (no `.cors(...)`); `AIAgent/.../config/WebConfig.java:29-36`

### Evidence pattern

```java
// AgentSecurityConfig — NO .cors(...)
public SecurityFilterChain agentOauth2FilterChain(HttpSecurity http) throws Exception {
    return http
        .csrf(AbstractHttpConfigurer::disable)
        .sessionManagement(...)
        .oauth2ResourceServer(...)
        .authorizeHttpRequests(...)
        .build();
}
```

### Why it matters

Without `.cors(...)` in the SecurityFilterChain, Spring Security's CorsFilter is not registered for the Security pipeline. Spring MVC CORS via `WebMvcConfigurer.addCorsMappings` happens at the handler-mapping layer, AFTER security has already accepted/rejected the preflight. The result is inconsistent enforcement: with `trabuco.auth.enabled=true`, OPTIONS preflights to AIAgent endpoints will get a 401 (no token) and never reach the WebConfig CORS rules — breaking browser clients and signaling a misconfiguration. With permit-all, CORS works but only because every request is allowed.

### Suggested fix

Add `.cors(cors -> cors.configurationSource(corsConfigurationSource))` to `agentOauth2FilterChain` and centralize CORS via a `CorsConfigurationSource` bean.

---

---

## F-AUTH-26 — Two SecurityFilterChains gated by `@ConditionalOnProperty` are mutually exclusive ONLY because of strict matching — ambiguous values like `TRUE`, `Yes`, `1` deactivate BOTH chains, leaving Spring Security default (HTTP Basic, all endpoints)

**Severity floor:** Medium
**Taxonomy:** OWASP-A05, API-API8, CWE-1188, ASVS-V14-1, TRABUCO-001

### Where to look

`API/.../security/SecurityConfig.java:66,95`; `AIAgent/.../security/AgentSecurityConfig.java:79,106`

### Evidence pattern

```java
@ConditionalOnProperty(value = "trabuco.auth.enabled", havingValue = "true")            // case-sensitive
@ConditionalOnProperty(value = "trabuco.auth.enabled", havingValue = "false", matchIfMissing = true)  // case-sensitive
```

### Why it matters

`@ConditionalOnProperty(havingValue = "true")` is case-sensitive on the value (per Spring's `OnPropertyCondition`). If an operator sets `TRABUCO_AUTH_ENABLED=TRUE` (env vars are commonly uppercase), Spring matches neither `havingValue="true"` (literal lowercase) nor `havingValue="false"` (matchIfMissing requires the property to be missing OR exactly `false`, not `TRUE`). Result: NO `SecurityFilterChain` bean is declared. Spring Boot's auto-config falls back to the default `springSecurityFilterChain` bean which … actually, with `spring-boot-starter-security` it auto-configures HTTP Basic with a generated user/password printed to stdout. Surprise: the dual-chain pattern depends on a value never being a typo.

### Suggested fix

Use a `@ConditionalOnExpression("#{environment['trabuco.auth.enabled'] == null || environment['trabuco.auth.enabled'].toLowerCase() == 'false'}")` pair, OR add a fail-fast assertion at startup that the property parses to a boolean.

---

---

## F-AUTH-27 — `JwtAuthenticationConverter.convert()` mutates `RequestContextHolder` as a side effect — Spring Security calls `convert()` more than once during context propagation in async flows

**Severity floor:** Medium
**Taxonomy:** ASVS-V8-1, JAVA-013, CWE-200

### Where to look

`API/.../JwtAuthenticationConverter.java:38-48` and `AIAgent/.../JwtAuthenticationConverter.java:40-50`

### Evidence pattern

```java
@Override
public AbstractAuthenticationToken convert(Jwt jwt) {
    IdentityClaims claims = extractor.extract(jwt);
    RequestContextHolder.set(claims);    // SIDE EFFECT in a "Converter"
    ...
}
```

### Why it matters

A `Converter<Jwt, AbstractAuthenticationToken>` is contractually pure — Spring Security caches/recomputes it on multiple paths (sync auth, reactive backfill, security context propagation in `DelegatingSecurityContextRunnable` for `@Async`/`CompletableFuture`). Side-effecting into `RequestContextHolder` from a Converter is an anti-pattern: the ThreadLocal is set on the *Spring Security thread*, not the request thread when context is propagated. This makes context dependent on which thread won the race and contributes to the leak issue in F-AUTH-14.

### Suggested fix

Move the `RequestContextHolder.set(...)` into a dedicated `OncePerRequestFilter` after authentication, with paired `clear()` in `finally`.

---

---

## F-AUTH-28 — No test coverage for "audience mismatch returns 401" or "wrong issuer returns 401" — `AuthEndToEndTest` only checks signature, expiry, scope

**Severity floor:** Medium
**Taxonomy:** ASVS-V14-1, ASVS-V2-1, OWASP-A07

### Where to look

`API/src/test/java/com/security/audit/api/config/security/AuthEndToEndTest.java:88-159`, `SignedJwtTestSupport.java:96-118`

### Evidence pattern

```java
JWTClaimsSet claims = new JWTClaimsSet.Builder()
    .subject(subject)
    .issuer(DEFAULT_ISSUER)
    .issueTime(...)
    .expirationTime(...)
    .jwtID(UUID.randomUUID().toString())
    .claim("scope", scopes.length == 0 ? "" : String.join(" ", scopes))
    .build();
// ^^ No `aud` claim minted, no test for audience-mismatch rejection
```

### Why it matters

Because no test exists, the audience-validation gap (F-AUTH-02) was never going to be detected by CI. ASVS V14.1 requires that security-critical configuration be validated by tests; the test scaffolding gives a false signal.

### Suggested fix

Add `audienceMismatchReturns401` and `wrongIssuerReturns401` cases to `AuthEndToEndTest`.

---

## False positives considered

- **F-AUTH-FP-01: "ApiKeyAuthFilter compares plaintext keys with `String.equals` (timing attack)."** Looked initially like the canonical timing-attack pattern, but the lookup is `Map.get(plaintext)` — a hashmap probe, not a linear `equals` chain. Still flagged as F-AUTH-05 because hashmap timing leaks bucket distribution and equals-on-collision is non-constant, but reframed from "trivial timing attack" to "weaker than constant-time" with realistic severity (High, not Critical).

- **F-AUTH-FP-02: "JWT `tenantId` claim trusted blindly (multi-tenant bypass)."** `DefaultJwtClaimsExtractor` reads `tenant_id` as an optional claim (line 33). Looked like a tenant-confusion vector, but the extractor only deserializes the claim — actual tenant enforcement (e.g., vector-store filter) is a downstream concern outside the auth domain. Forwarded to RAG/multi-tenant agent.

- **F-AUTH-FP-03: "`SecurityConfig` doesn't configure HSTS / X-Frame-Options."** These ARE configured by `SecurityHeadersFilter` at `Order(HIGHEST_PRECEDENCE)` and run on every response (api-minimal/API/.../config/SecurityHeadersFilter.java). Not a finding.

- **F-AUTH-FP-04: "`stateless` session policy missing on permit-all chain."** Both chains explicitly set `SessionCreationPolicy.STATELESS` (lines 71/100 of SecurityConfig.java). ASVS V3.1 satisfied.

- **F-AUTH-FP-05: "`@EnableMethodSecurity` not configured."** Present on both `SecurityConfig` and `AgentSecurityConfig` (line 48 / line 61). ASVS V4.2 satisfied at the *config* level — but most controllers still lack `@PreAuthorize` (covered in F-AUTH-17 and F-AUTH-18).

- **F-AUTH-FP-06: "`AuthScope` ThreadLocal cleanup absent."** `AuthScope` itself implements `AutoCloseable` correctly (line 56-62) and is *the* recommended pattern. The actual leak is in JWT-side code that doesn't *use* AuthScope (covered in F-AUTH-14).

- **F-AUTH-FP-07: "JWT path supports `none` algorithm."** `NimbusJwtDecoder.fromIssuerLocation` rejects `none` by default (Nimbus JWT library). Not a current bug — flagged as part of F-AUTH-13 only because the algorithm whitelist isn't *explicitly* pinned, leaving the door open if JWKS rotates.

---

---

## F-AUTH-19 — `AuthProblemDetailHandler` returns generic 401 — does NOT distinguish "expired" vs "missing" vs "forged" — useful but loses one specific safety property

**Severity floor:** Low
**Taxonomy:** ASVS-V7-1, OWASP-A09, CWE-209

### Where to look

`API/.../security/AuthProblemDetailHandler.java:43-52`

### Evidence pattern

```java
ProblemDetail problem = ProblemDetail.forStatus(HttpStatus.UNAUTHORIZED);
problem.setType(UNAUTHORIZED_TYPE);
problem.setTitle("Unauthorized");
problem.setDetail("Bearer token is missing or invalid.");
```

### Why it matters

Sanitization is good (no exception messages). But the lack of `WWW-Authenticate` header (`Bearer realm="...", error="invalid_token", error_description="..."`) means standards-compliant clients (e.g., curl with --negotiate, Spring's OAuth2 client auto-refresh) can't distinguish "must re-auth" from "lost token" from "scope insufficient". This is borderline — some shops want this opacity for security. Flagging at Low for documentation visibility.

### Suggested fix

Emit `WWW-Authenticate: Bearer error="invalid_token"` for 401 and `Bearer error="insufficient_scope"` for 403.

---

---

## F-AUTH-21 — `WWW-Authenticate` header missing on 401; `application/problem+json` body but no scheme advertisement

**Severity floor:** Low
**Taxonomy:** API-API2, ASVS-V7-1, CWE-287

### Where to look

`API/.../security/AuthProblemDetailHandler.java:65-70` (overlaps with F-AUTH-19)

### Evidence pattern

```java
private void write(HttpServletResponse response, HttpStatus status, ProblemDetail problem) throws IOException {
    response.setStatus(status.value());
    response.setContentType(MediaType.APPLICATION_PROBLEM_JSON_VALUE);
    response.setCharacterEncoding(StandardCharsets.UTF_8.name());
    response.getWriter().write(objectMapper.writeValueAsString(problem));
}
```

### Why it matters

Resource-server best practice is `WWW-Authenticate: Bearer realm="<name>"`; absence is a minor RFC 6750 deviation that confuses standard clients. Combined with F-AUTH-19.

### Suggested fix

Set `WWW-Authenticate` header before writing the body.

---

---

## F-AUTH-24 — `cors.allow-credentials=false` default is fine, but `WebConfig` (AIAgent) hardcodes `allowCredentials(false)` — operator setting `CORS_ALLOW_CREDENTIALS=true` is silently ignored

**Severity floor:** Low
**Taxonomy:** API-API8, SPRING-007, CWE-1188

### Where to look

`AIAgent/.../config/WebConfig.java:29-36`

### Evidence pattern

```java
@Override
public void addCorsMappings(CorsRegistry registry) {
    registry.addMapping("/**")
            .allowedOrigins(allowedOrigins.split(","))
            .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
            .allowedHeaders("Content-Type", "Authorization", "X-Correlation-ID")
            .allowCredentials(false)
            .maxAge(3600);
}
```

### Why it matters

Hardcoded `allowCredentials(false)` is the *safe* default but ignores `cors.allow-credentials` from `application.yml`. Inconsistent with the API module's `WebConfig` which honors `${cors.allow-credentials}`. Operator confusion → may cause downstream issue when an authenticated SPA needs `withCredentials: true`.

### Suggested fix

Read from `${cors.allow-credentials:false}` like the API module does, and emit a startup log when `true` to flag the wider attack surface.

---

---

## F-AUTH-20 — `bearer ` token check is case-insensitive (`toLowerCase().startsWith`) — accepts `BEARER`, `bEaReR`, etc.; non-issue but indicates loose parsing

**Severity floor:** Informational
**Taxonomy:** API-API2, CWE-20

### Where to look

`AIAgent/.../security/ApiKeyAuthFilter.java:65-67`

### Evidence pattern

```java
if (!authHeader.toLowerCase().startsWith("bearer ")) {
    sendError(response, HttpStatus.FORBIDDEN, "Invalid API key");
    return;
}
```

### Why it matters

Per RFC 6750 §2.1 the scheme is case-insensitive — so the implementation is correct. But `toLowerCase()` on the entire header (not just the scheme prefix) means the locale-default lowercase is applied to the secret too, which then `substring(7).trim()` re-extracts from the original. Defensive but harmless. Mention here for completeness.

### Suggested fix

Use `Locale.ROOT` for safety: `authHeader.toLowerCase(Locale.ROOT).startsWith("bearer ")`.

---

---

