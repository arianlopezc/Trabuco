# Trabuco Security Audit — AIAgent + Java Platform Domain

AIAgent runtime (Spring AI 1.0.5, RAG, tool dispatch, A2A protocol, vector store, guardrails, MCP exposure) plus Java-platform gotchas (deserialization, regex DoS, HTTP client hardening, virtual-thread context).

This file is the **detail reference** for the
`trabuco-security-audit-aiagent-java` specialist subagent. The orchestrator
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
(7 Critical,
 9 High,
 15 Medium,
 11 Low,
 0 Informational)

---

## F-AIAGENT-01 — Hardcoded API keys in `ApiKeyAuthFilter` ship as the default auth path

**Severity floor:** Critical
**Taxonomy:** ASVS-V2-2, OWASP-A07, CWE-798, CWE-256, ASVS-V10-1, TRABUCO-001

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/security/ApiKeyAuthFilter.java:40-49,71-83`

### Evidence pattern

```java
@Component
@ConditionalOnProperty(name = "app.aiagent.api-key.enabled", havingValue = "true", matchIfMissing = true)
public class ApiKeyAuthFilter extends OncePerRequestFilter {
    private static final Map<String, KeyEntry> API_KEYS = Map.of(
        "public-read-key", new KeyEntry("public", "public-read-key"),
        "partner-secret-key", new KeyEntry("partner", "partner-secret-key")
    );
    ...
    String key = authHeader.substring(7).trim();
    KeyEntry entry = API_KEYS.get(key);  // plaintext compare via HashMap.get
```

### Why it matters

The filter is `matchIfMissing=true` and the SecurityFilterChain default is `permitAll()` (`AgentSecurityConfig.agentPermitAllFilterChain`), so out of the box every Trabuco AIAgent project accepts the literal string `partner-secret-key` as a partner-tier credential — granting `@RequireScope("partner")` access to `/ingest`, `/ingest/batch`, `/webhooks`, `/tasks/{id}/stream`. Plaintext `Map.get` comparison is also vulnerable to timing side-channel and these strings are checked into VCS; key rotation requires a code release. The companion `application.yml` `agent.auth.keys` block is ignored — only the in-class `Map` is consulted.

### Suggested fix

Default `app.aiagent.api-key.enabled=false`, require an explicit, hashed-at-rest configuration source, and use `MessageDigest.isEqual` for comparison.

---

---

## F-AIAGENT-02 — `DocumentIngestionService` REST endpoint authorized only by hardcoded API key

**Severity floor:** Critical
**Taxonomy:** TRABUCO-005, OWASP-LLM-04, OWASP-A04, API-API6, CWE-862

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/protocol/IngestionController.java:63-88`

### Evidence pattern

```java
@PostMapping
@RequireScope("partner")
public ResponseEntity<Map<String, Object>> ingest(@RequestBody IngestRequest req) {
    ingestionService.ingest(req.text(), req.metadata());
```

### Why it matters

`@RequireScope("partner")` resolves through `ScopeInterceptor` → `ScopeEnforcer` → in-memory `API_KEYS` map (F-01). With the publicly-known `partner-secret-key`, any unauthenticated attacker can poison the RAG vector store with arbitrary content (LLM-04 data poisoning). Combined with the `RetrievalAugmentationAdvisor` in `PrimaryAgent`, those poisoned chunks are then injected into the system prompt of every chat (LLM-01 indirect prompt injection). There is no rate limit on `/ingest` (RateLimiter not invoked), no per-byte/per-doc cap, and no source allow-list.

### Suggested fix

Gate ingestion on a JWT scope (`SCOPE_rag:write`), add `RateLimiter.checkRateLimit`, enforce per-tenant byte/doc quotas, and require the operator to explicitly enable the endpoint.

---

---

## F-AIAGENT-03 — Vector store has no tenant / authority-scope partition key

**Severity floor:** Critical
**Taxonomy:** TRABUCO-004, OWASP-LLM-08, API-API1, CWE-639

### Where to look

- `AIAgent/src/main/resources/db/vector-migration/V1__create_vector_schema.sql:35-40`

### Evidence pattern

```sql
CREATE TABLE IF NOT EXISTS documents (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    content   TEXT NOT NULL,
    metadata  JSONB,
    embedding vector(384)
);
```

### Why it matters

No `tenant_id` column, no composite index, no `FilterExpressionBuilder` filter on `similaritySearch` and no metadata enforcement on `vectorStore.add(chunks)`. Every caller can read every document any other caller (or partner key) has ingested. Same applies to the Qdrant variant in the full-fat archetype: the collection is `documents` shared globally, no payload filter. This is the canonical LLM-08 / TRABUCO-004 failure.

### Suggested fix

Emit `tenant_id` (or `authority_scope`) column + composite HNSW index; have `DocumentIngestionService` stamp `CallerContext` identity on every chunk; require `VectorKnowledgeRetriever` to add a `tenantId == ?` filter from the resolved caller before search.

---

---

## F-AIAGENT-04 — `RetrievalAugmentationAdvisor` injects raw RAG content into prompt with no provenance fence

**Severity floor:** Critical
**Taxonomy:** OWASP-LLM-01, OWASP-LLM-04, CWE-1039

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/agent/PrimaryAgent.java:80-97`

### Evidence pattern

```java
vectorStore.ifPresent(vs -> builder.defaultAdvisors(buildRagAdvisor(vs)));
...
private static Advisor buildRagAdvisor(VectorStore vectorStore) {
    return RetrievalAugmentationAdvisor.builder()
        .documentRetriever(VectorStoreDocumentRetriever.builder()
            .vectorStore(vectorStore).topK(4).build())
        .build();
}
```

### Why it matters

Combined with F-02 / F-03, anyone who can ingest can rewrite the agent's system prompt for everyone (indirect prompt injection). Even when ingestion is gated, the advisor uses Spring AI's default prompt template that concatenates retrieved text directly into the user/system message with no untrusted-content delimiters, no per-document source label, and no instruction to ignore embedded instructions. This violates LLM-01.

### Suggested fix

Supply a custom `PromptTemplate` to `RetrievalAugmentationAdvisor` that wraps each retrieved chunk in `<untrusted_document source="...">` fences and tells the model that text inside fences is data, not instructions; sanitize/strip role tokens before injection.

---

---

## F-AIAGENT-06 — MCP server exposed at `/mcp` with no authentication boundary

**Severity floor:** Critical
**Taxonomy:** LLM-EXT-02, OWASP-LLM-06, API-API2, TRABUCO-003, ASVS-V4-1

### Where to look

- `AIAgent/src/main/resources/application.yml:43-48`

### Evidence pattern

```yaml
spring:
  ai:
    mcp:
      server:
        enabled: true
        name: AiagentPgvectorRag AI Agent
```

### Why it matters

Spring AI MCP server auto-discovers every `@Tool` bean (`PlaceholderTools`, `KnowledgeTools`, `SpecialistAgentTool`) and exposes them at `/mcp` over Streamable HTTP. With the default chain (`permitAll`) and the API-key filter passing through anonymous (since `/mcp` doesn't require a `@RequireScope`), any unauthenticated client on the network can call those tools at attacker-chosen pace. With `trabuco.auth.enabled=true` the chain authenticates, but the JWT scope check (`SCOPE_*`) is not wired into the auto-registered tool methods. Comment in the file even instructs `claude mcp add --transport http aiagent-pgvector-rag-agent http://localhost:8080/mcp` — that URL is intended public.

### Suggested fix

Bind MCP server to localhost by default; require an explicit auth interceptor on `/mcp`; wire `@PreAuthorize` on every `@Tool` method or on the MCP transport adapter; add per-tool scope mapping documented in the skill.

---

---

## F-AIAGENT-07 — A2A JSON-RPC `/a2a` endpoint accepts anonymous callers; no `@RequireScope` annotation

**Severity floor:** Critical
**Taxonomy:** TRABUCO-003, OWASP-LLM-06, API-API2, API-API5, CWE-862

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/protocol/A2AController.java:43-91`

### Evidence pattern

```java
@PostMapping("/a2a")
public JsonRpcResponse handle(@RequestBody JsonRpcRequest body) {
    var caller = CallerContext.get();   // returns anonymous if no/invalid bearer
    rateLimiter.checkRateLimit(caller);
    ...
    case "ask_question" -> {
        ScopeEnforcer.requireScope("public", caller);
```

### Why it matters

No `@RequireScope` on the controller method, only an in-method scope check on individual skills (and `tasks/get` has no scope check at all — line 111-126). Anonymous callers get the 10-rps tier (F-13). `tasks/get` lets anyone enumerate task IDs and read results that may contain sensitive LLM responses generated for other callers (BOLA — task IDs are random UUIDs but the controller never ties tasks to caller). Inter-agent calls are not mutually authenticated; the spec calls for caller-identity verification (TRABUCO-003) but identity is whatever bearer token the request carries (or anonymous).

### Suggested fix

Require `@RequireScope("public")` (or stronger) on the controller; bind tasks to caller-keyhash on submit and check on `tasks/get`; require mutual TLS or signed agent identity for inter-agent calls.

---

---

## F-JAVA-01 — Spring AMQP `Jackson2JsonMessageConverter` accepts `__TypeId__` header without trusted-packages allow-list (deserialization gadget — JAVA-001)

**Severity floor:** Critical
**Taxonomy:** JAVA-001, OWASP-A08, ASVS-V8-3, CWE-502, CWE-915

### Evidence pattern

```java
// EventConsumer
@Bean
public Jackson2JsonMessageConverter jsonMessageConverter(ObjectMapper objectMapper) {
  return new Jackson2JsonMessageConverter(objectMapper);
}
// Events
@Bean
public Jackson2JsonMessageConverter jackson2JsonMessageConverter(ObjectMapper objectMapper) {
  return new Jackson2JsonMessageConverter(objectMapper);
}
```

### Why it matters

`Jackson2JsonMessageConverter` reads the AMQP `__TypeId__` header and instantiates the Java class named there via Jackson's `ObjectMapper.readValue(..., Class)` — this is functionally equivalent to enabling Jackson default typing on inbound messages. The default `DefaultClassMapper` accepts **any** class on the classpath that Jackson can deserialize; `setTrustedPackages` is not invoked, no `BasicPolymorphicTypeValidator` is wired. A publisher able to put a message on the `placeholder-events` queue (or any other queue this listener serves) can set `__TypeId__` to a known gadget class — `org.springframework.context.support.ClassPathXmlApplicationContext`, `com.zaxxer.hikari.HikariConfig`, or any class on the classpath with a setter that fetches a remote URL — and trigger remote-class-loading / SSRF / RCE the moment the consumer deserializes. Spring AMQP's documentation explicitly recommends `converter.setClassMapper(new DefaultClassMapper { setTrustedPackages(...) })` or the new `Jackson2JsonMessageConverter#setAllowedListPatterns`, neither of which is configured. The presence of `@JsonSubTypes`-restricted sealed interface in `PlaceholderEvent` does **not** help because the converter consults `__TypeId__` first and bypasses the sealed-interface allow-list entirely.

### Suggested fix

In both `RabbitConfig` files, configure
```java
DefaultClassMapper mapper = new DefaultClassMapper();
mapper.setTrustedPackages("com.security.audit.model.events");
converter.setClassMapper(mapper);
```
or set `converter.setAllowedListPatterns(List.of("com\\.security\\.audit\\.model\\.events\\..*"))`. Crucially, also set `converter.setDefaultType(PlaceholderEvent.class)` so consumers fall back to the sealed-interface allow-list when no `__TypeId__` is supplied.

---

---

## F-AIAGENT-05 — `KnowledgeTools.askQuestion` `@Tool` returns raw retrieved text to the LLM

**Severity floor:** High
**Taxonomy:** OWASP-LLM-01, OWASP-LLM-05, OWASP-LLM-04

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/knowledge/KnowledgeTools.java:39-51`

### Evidence pattern

```java
@Tool(description = "Ask a natural language question about ... services, policies, hours, pricing, and more")
public String askQuestion(@ToolParam(...) String query) {
    List<Document> docs = retriever.retrieve(query, TOP_K);
    return docs.stream().map(Document::getText).collect(Collectors.joining("\n\n"));
}
```

### Why it matters

Tool output flows back through Spring AI as a tool result that the model treats as content; raw poisoned chunks (from F-02/03/04) are concatenated unfenced and re-enter the model's context window with no provenance. Combined with `RateLimiter` not being invoked at the tool boundary, this is also an inexpensive abuse vector.

### Suggested fix

Wrap each chunk with source/provenance fences and length-cap the joined output; consider emitting structured JSON with `text` + `source_id` so the model can be instructed to treat `text` as data only.

---

---

## F-AIAGENT-08 — Webhook URL is not validated against an allow-list (SSRF)

**Severity floor:** High
**Taxonomy:** OWASP-A10, API-API7, CWE-918, JAVA-011

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/protocol/WebhookController.java:27-40`

### Evidence pattern

```java
@NotBlank(message = "URL is required")
String url();
...
webClient.post().uri(wh.url()).bodyValue(payload).retrieve().toBodilessEntity().subscribe(...)
```

### Why it matters

Any caller able to invoke `/webhooks` with the partner key (F-01) can register `http://169.254.169.254/latest/meta-data/`, `http://localhost:8081/admin`, or any internal-network host. Every dispatched event is then POSTed to that URL with a signature header. The DTO validates only `@NotBlank`, never URL shape, scheme, or destination IP. `WebClient.create()` follows redirects by default (JAVA-011), enabling attacker-controlled DNS rebinding into private ranges.

### Suggested fix

Validate the URL is `https://` and resolves to a non-private IP at registration *and* dispatch time; maintain an operator-controlled domain allow-list; disable redirect-follow on the WebClient.

---

---

## F-AIAGENT-09 — Webhook signing key is the caller's API key label, not a separate signing secret

**Severity floor:** High
**Taxonomy:** OWASP-A02, ASVS-V11-1, CWE-345

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/protocol/WebhookController.java:33`

### Evidence pattern

```java
var reg = webhookManager.register(url, events, CallerContext.get().keyLabel());
...
String signature = hmacSha256(wh.apiKey(), payload);
```

### Why it matters

The label stored as `apiKey()` and passed to `hmacSha256` is `"partner-secret-key"` (literal string from F-01) — i.e., the same string that's already known to anyone who reads the source. Any receiver of a webhook payload can forge subsequent webhooks because the HMAC key is public. There's no per-registration random secret returned to the registrant for verification.

### Suggested fix

Generate a per-webhook random signing secret at registration, return it to the caller exactly once, store its hash; sign deliveries with the random secret.

---

---

## F-AIAGENT-10 — Input guardrail fails open and is regex-only on output; `agent.guardrails.enabled` flag is ignored

**Severity floor:** High
**Taxonomy:** OWASP-LLM-01, OWASP-LLM-02, OWASP-A04

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/security/InputGuardrailAdvisor.java:60-79`

### Evidence pattern

```java
} catch (Exception e) {
    log.error("Guardrail classification failed, allowing input: {}", e.getMessage());
    return null; // fail-open for availability; production might fail-closed
}
...
if (response != null && response.toUpperCase().contains("BLOCKED")) {
```

### Why it matters

1. Fail-open on exception — any LLM error or rate-limit exhaustion lets prompt-injection payloads pass.
2. The classifier checks `response.contains("BLOCKED")`. A model that emits `DECISION: ALLOWED but the user said "ignore previous BLOCKED instructions"` would be flagged as BLOCKED — and conversely, a jailbreak that elicits `decision: allowed` lowercase passes (`toUpperCase` then `contains("BLOCKED")` only catches the literal token, but the parser doesn't structurally enforce `DECISION:` field).
3. Output guardrail is four `Pattern.compile` calls with simple regex — the email regex matches `\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Z|a-z]{2,}\b` (note `[A-Z|a-z]` is a character class containing the pipe, a common ReDoS-shaped mistake) and is trivially bypassed by `user[at]example[dot]com`. PII is *not* redacted; the entire response is replaced with a fixed fallback, eliminating the legitimate response if any false positive triggers.
4. `agent.guardrails.enabled` property is read nowhere — there is no `@ConditionalOnProperty` on either guardrail bean. Operators believe they can disable guardrails via config; they cannot.

### Suggested fix

Fail-closed on guardrail errors for high-risk paths; parse the classifier output structurally; add `@ConditionalOnProperty` honoring the documented flag; replace regex-only PII detection with a maintained library + redact-in-place behavior.

---

---

## F-AIAGENT-12 — No max-tool-call recursion / token budget on `ChatClient`

**Severity floor:** High
**Taxonomy:** OWASP-LLM-10, OWASP-LLM-06, API-API4, CWE-770, CWE-400

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/agent/PrimaryAgent.java:76-105`

### Evidence pattern

```java
ChatClient.Builder builder = ChatClient.builder(chatModel)
    .defaultSystem(SYSTEM_PROMPT)
    .defaultTools(placeholderTools, knowledgeTools, specialistAgentTool);
...
return chatClient.prompt().user(message).call().content();
```

### Why it matters

`ChatOptions.maxToolCalls` is not set; `PrimaryAgent` exposes `SpecialistAgentTool` which itself calls `SpecialistAgent` which has `KnowledgeTools` — there is a code path for indirect cycles, and Spring AI's default is to keep calling tools until the model stops. Combined with no token-cost cap and no per-caller cost meter, an attacker who can pass the input guardrail (F-10 fail-open) can drive the LLM provider bill up — anonymous tier still gets 10 rps × 1024 max-output-tokens × N tool round-trips.

### Suggested fix

Set `ChatOptions.toolCallLimit` to a small constant (e.g., 5); add per-caller token-budget meter and quota; cap `topK`-driven retrieval bytes.

---

---

## F-AIAGENT-16 — Default OAuth2 resource-server YAML has empty audience and no validator

**Severity floor:** High
**Taxonomy:** ASVS-V2-1, OWASP-A07, CWE-345

### Where to look

- `AIAgent/src/main/resources/application.yml:37-42`

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

When operators set `OIDC_ISSUER_URI` but forget `OIDC_AUDIENCE` (the YAML default is empty, which Spring reads as "no audience required"), every JWT signed by the configured issuer is accepted regardless of intended `aud` — including tokens issued for entirely different services in the same realm (Auth0/Okta tenant token confusion). The `JwtAuthenticationConverter` does no further validation. There is no custom `OAuth2TokenValidator` adding `JwtClaimValidator(AUD)`.

### Suggested fix

Make audience validation mandatory when issuer is set; fail boot if audience is empty; document the cross-service token-confusion risk.

---

---

## F-JAVA-02 — A2A JSON-RPC envelope deserializes user-controlled `params` as `Map<String, Object>`, then unchecked-casts to `Map<String, Object>` and `String` at every handler — Jackson default-typing-equivalent risk via Spring AI tool inputs

**Severity floor:** High
**Taxonomy:** JAVA-001, OWASP-A08, OWASP-A03, CWE-502, CWE-704, TRABUCO-003

### Evidence pattern

```java
// JsonRpcRequest.java
@Value.Default
default Map<String, Object> params() { return Map.of(); }

// A2AController.java
private JsonRpcResponse handleTasksSend(String rpcId, Map<String, Object> params, ...) {
    String skill = (String) params.get("skill");                       // ClassCastException → 500
    @SuppressWarnings("unchecked")
    var input = (Map<String, Object>) params.getOrDefault("input", Map.of());
    ...
    String query = (String) input.get("query");                         // unchecked
    String taskId = taskManager.submitTask(() -> Map.of("answer", knowledgeTools.askQuestion(query)));
}
```

### Why it matters

`Map<String, Object>` is Jackson's untyped bag — values are deserialized to whatever shape the JSON has (`String`, `Number`, `Boolean`, `List`, `Map`). The controller then unchecked-casts these objects, so a malicious caller sending `{"skill":"ask_question","input":{"query":{"@class":"..."}}}` yields a `LinkedHashMap`, not a `String`, producing a `ClassCastException` — but more importantly the value flows into `knowledgeTools.askQuestion(query)` whose signature is `String`, meaning the cast happens at the call site. If the caller crafts `query` as a list/map, Jackson stores it as a `LinkedHashMap` and the cast `(String) input.get("query")` throws — bypassing the rate-limit-after-skill-check intent and surfacing in `GlobalExceptionHandler`. Worse: future additions of `case "fetch_url"` style skills that pass `params` straight to a tool will deserialize attacker-controlled JSON shapes into untyped sinks. Spring AI `@Tool` methods with `Map<String, Object>` argument types extend this gadget — the agent loop converts LLM output to the same untyped map, so an attacker who can prompt-inject the LLM can shape `params` arbitrarily.

### Suggested fix

Replace the `Map<String, Object> params()` bag with a `JsonNode` (so handlers explicitly call `.path("skill").asText()` and `.path("input").path("query").asText()`, surfacing typed null/missing as 400) or a sealed interface of skill-specific param records (`AskQuestionParams`, `FetchUrlParams`) wired by `@JsonTypeInfo`. Also configure `ObjectMapper.configure(DeserializationFeature.FAIL_ON_TRAILING_TOKENS, true)` and set `StreamReadConstraints` (see F-JAVA-08).

---

---

## F-JAVA-03 — `WebClient.create()` in `WebhookManager` has no timeouts, no redirect bound, no proxy hardening, no maximum response size — JDK-default redirect follow into private IPs (JAVA-008, JAVA-011)

**Severity floor:** High
**Taxonomy:** JAVA-008, JAVA-011, OWASP-A10, ASVS-V12-6, CWE-918, CWE-295, CWE-770

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/event/WebhookManager.java:22,57-68`

### Evidence pattern

```java
private final WebClient webClient = WebClient.create();          // line 22 — zero config
...
webClient.post()
    .uri(wh.url())
    .header("Content-Type", "application/json")
    ...
    .bodyValue(payload)
    .retrieve()
    .toBodilessEntity()
    .subscribe(...);                                              // line 57-68
```

### Why it matters

`WebClient.create()` returns a Reactor-Netty client with: (a) no read/connect/response timeout — a slowloris webhook target ties up agent resources indefinitely; (b) auto-follow of HTTP 30x redirects to any host (JDK 21 reactor-netty default is `HttpClient.followRedirect(true)` when configured by Spring Boot's auto-config; this raw-built client uses Reactor's defaults — `followRedirect=false`, but the `WebClient.create()` flavor used here does NOT pin that explicitly, and Spring Boot's `WebClientCustomizer` runs only on `WebClient.Builder` injection, not on `WebClient.create()`); (c) no enforced TLS protocol/cipher floor — relies entirely on JVM defaults; (d) no `responseTimeout` so reads block until the OS socket idle-timeout. Combined with F-AIAGENT-08 (URL allow-list missing — covered by aiagent.md), the webhook delivery path becomes an unbounded SSRF-and-resource-exhaustion sink. Even if the URL is validated at registration, DNS-rebinding rotates the resolved IP between registration and dispatch, sending the signed payload to `169.254.169.254` or `localhost:8080` (the API itself) on the second call. The `.subscribe(...)` callback discards both success and error logs aside from `wh.url()` — there's no audit of which downstream IP the request actually hit.

### Suggested fix

```java
HttpClient httpClient = HttpClient.create()
    .responseTimeout(Duration.ofSeconds(5))
    .option(ChannelOption.CONNECT_TIMEOUT_MILLIS, 3000)
    .followRedirect(false)                          // pin explicitly
    .resolver(DefaultAddressResolverGroup.INSTANCE) // or custom IP allow-list resolver
    .secure(spec -> spec.sslContext(/* TLSv1.3 floor, server-name verification */));
this.webClient = WebClient.builder()
    .clientConnector(new ReactorClientHttpConnector(httpClient))
    .codecs(c -> c.defaultCodecs().maxInMemorySize(64 * 1024))
    .build();
```
plus IP allow-list validation per dispatch (re-resolve and compare vs. expected CIDR ranges).

---

---

## F-JAVA-05 — `TaskManager.tasks` and `WebhookManager.webhooks` are unbounded `ConcurrentHashMap` — DoS via memory exhaustion (no eviction, no per-tenant cap)

**Severity floor:** High
**Taxonomy:** JAVA-012, OWASP-A04, API-API4, CWE-400, CWE-770

### Evidence pattern

```java
// TaskManager
private final ConcurrentHashMap<String, TaskRecord> tasks = new ConcurrentHashMap<>();
private final ConcurrentHashMap<String, List<Consumer<TaskEvent>>> subscribers = new ConcurrentHashMap<>();
public String submitTask(Callable<Map<String, Object>> work) {
    String taskId = "TASK-" + UUID.randomUUID();
    tasks.put(taskId, record);                          // never evicted
    ...
}
public void reset() { tasks.clear(); subscribers.clear(); }   // only test-time hook

// WebhookManager
private final ConcurrentHashMap<String, WebhookRegistration> webhooks = new ConcurrentHashMap<>();
public WebhookRegistration register(String url, ...) {
    var reg = ImmutableWebhookRegistration.builder().webhookId("WH-" + UUID.randomUUID())...build();
    webhooks.put(reg.webhookId(), reg);                 // never evicted
}

// RateLimiter
private final ConcurrentHashMap<String, Deque<Long>> store = new ConcurrentHashMap<>();
// keys never removed even after window expires
```

### Why it matters

All three caches grow unbounded:
- `TaskManager.tasks` — a `partner`-tier caller authenticated with the bake-in API key (F-AUTH-03 / F-AIAGENT-01) calls `/a2a` `tasks/send` 200×/minute (the partner rate-limit ceiling) for a year ⇒ ~100M `TaskRecord` instances + their `Map<String,Object>` results held in heap. No size cap, no LRU, no time-based eviction. `reset()` is package-private test-only.
- `WebhookManager.webhooks` — same story; no cap on per-caller registrations or globally. Combined with the unbounded webhook URL (F-AIAGENT-08), an attacker inflates webhooks to OOM the JVM.
- `RateLimiter.store` — keyed by `caller.keyHash()`. The `anonymous` tier is rate-limited but uses a single shared key (`"anonymous"`), so this isn't an attacker-controlled key. However when JWT auth is on (`trabuco.auth.enabled=true`) the key would become per-subject; without a cleanup of empty deques (`while (!window.isEmpty() ...)` removes entries from a deque but never removes the deque from the map), the map grows forever.

### Suggested fix

Replace each with `Caffeine.newBuilder().expireAfterWrite(...).maximumSize(...)` or a Redis-backed store (matching the EventConsumer/data audit's recommendation). For `TaskManager`, expire completed tasks after 1 h; for `WebhookManager`, cap registrations per tenant (e.g., 10) and persist to a real datastore so restarts don't lose state. For `RateLimiter`, evict empty deques on each check.

---

---

## F-AIAGENT-13 — Rate limiter keyed by caller key-hash; anonymous callers share one bucket

**Severity floor:** Medium
**Taxonomy:** API-API4, OWASP-LLM-10, ASVS-V2-3, CWE-770

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/security/RateLimiter.java:23-42`

### Evidence pattern

```java
private final ConcurrentHashMap<String, Deque<Long>> store = new ConcurrentHashMap<>();
...
Deque<Long> window = store.computeIfAbsent(caller.keyHash(), k -> new ConcurrentLinkedDeque<>());
```

### Why it matters

All anonymous callers share the same string key `"anonymous"` so their requests aggregate into a single 10-req/min bucket — but that's a global cap, not per-IP, and every legitimate anonymous caller can DoS the whole anonymous pool by exhausting it. Conversely, an attacker can flood under different IPs without ever getting rate-limited because the bucket key has no IP component. There's also no eviction on the in-memory `ConcurrentHashMap` — long-lived process accumulates one entry per distinct key-hash forever (CWE-400).

### Suggested fix

Key the rate-limiter on `(remoteAddr, caller.keyHash())`; add a Caffeine-cache-bounded backing store; consider Bucket4j with a Redis backend.

---

---

## F-AIAGENT-14 — `IngestionController` and `WebhookController` accept `@RequestBody` without `@Valid`; metadata blob is unbounded

**Severity floor:** Medium
**Taxonomy:** ASVS-V5-1, API-API4, CWE-20, CWE-770, OWASP-A04

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/protocol/IngestionController.java:65,77`

### Evidence pattern

```java
public ResponseEntity<Map<String, Object>> ingest(@RequestBody IngestRequest req) {
...
public ResponseEntity<Map<String, Object>> ingestBatch(@RequestBody List<IngestRequest> requests) {
...
public record IngestRequest(String text, Map<String, Object> metadata) {}
```

### Why it matters

No `@Valid`, no `@NotBlank`/`@Size` on `text`, no per-batch cap on `requests.size()`, no max bytes on `text`, no schema check on the arbitrary `Map<String, Object>` metadata. A malicious caller can submit a 100 MB string, or millions of batch entries, or metadata containing reserved keys that collide with vector-store tenant fields.

### Suggested fix

Add `@Valid` and a constraint-annotated DTO; cap batch size; restrict metadata keys to a documented allow-list; set `spring.servlet.multipart.max-request-size`.

---

---

## F-AIAGENT-15 — `ResponseStatusException.getReason()` echoed back in ProblemDetail `detail`

**Severity floor:** Medium
**Taxonomy:** TRABUCO-002, ASVS-V7-1, OWASP-A09, CWE-209, CWE-200

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/config/AgentExceptionHandler.java:62-85`

### Evidence pattern

```java
problem.setDetail(ex.getReason() != null ? ex.getReason() : "Request rejected");
...
problem.setDetail(ex.getMessage() != null ? ex.getMessage() : "Invalid argument");
```

### Why it matters

`ScopeEnforcer.requireScope` raises `ResponseStatusException(FORBIDDEN, "This action requires partner access. Your tier: anonymous")` — the response detail leaks the tier the caller currently holds, helping an attacker confirm their key was accepted but at a lower tier (vs. a generic 403 "forbidden"). `IllegalArgumentException` handler echoes `ex.getMessage()` raw — any internal detail thrown from tool code (e.g., DB layer messages, regex errors) becomes part of the public response. Catch-all 500 handler is properly sanitized; the per-exception ones are not.

### Suggested fix

Map `ResponseStatusException` and `IllegalArgumentException` to fixed, sanitized strings; log the original for support correlation.

---

---

## F-AIAGENT-17 — Default permit-all SecurityFilterChain leaves `/actuator/prometheus`, `/actuator/metrics`, `/.well-known/agent.json`, `/capabilities`, `/mcp`, `/a2a` open even when JWT auth is on

**Severity floor:** Medium
**Taxonomy:** SPRING-001, OWASP-A05, ASVS-V4-1, API-API9, TRABUCO-008

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/config/security/AgentSecurityConfig.java:88-95,111`

### Evidence pattern

```java
.requestMatchers("/actuator/health/**", "/actuator/info", "/.well-known/agent.json").permitAll()
.anyRequest().authenticated()
```

### Why it matters

With `trabuco.auth.enabled=true` (the recommended state) the JWT chain authenticates `/mcp` and `/a2a` — but `prometheus` and `metrics` are exposed publicly because they're under `/actuator/**` which the chain authenticates by default — *but* prometheus tooling commonly cannot send a JWT, so operators frequently widen the matcher. Conversely, `DiscoveryController.@GetMapping("/capabilities")` enumerates internal scope names ("public", "partner") and tool descriptions to anonymous callers — useful for an attacker mapping the tool surface (LLM-EXT-01).

### Suggested fix

Add `permitAll` for `/actuator/health/**` only and require IP allow-list / scrape-token for prometheus; reduce `/capabilities` to authenticated callers or strip scope metadata.

---

---

## F-AIAGENT-19 — ONNX embedding model downloaded over network with no checksum / signature verification

**Severity floor:** Medium
**Taxonomy:** OWASP-LLM-03, OWASP-A08, MAVEN-001, CWE-494

### Where to look

- `AIAgent/pom.xml:128-131`

### Evidence pattern

```xml
<dependency>
    <groupId>org.springframework.ai</groupId>
    <artifactId>spring-ai-starter-model-transformers</artifactId>
</dependency>
```

### Why it matters

`spring-ai-starter-model-transformers` (via DJL) downloads the ONNX model from a remote URL on first boot with no operator-controlled checksum. Anyone able to MITM the build VM (or the first prod boot before the cache is populated) can substitute a malicious embedding model. This is the canonical LLM-03 supply-chain failure for embedding models.

### Suggested fix

Pin the ONNX model artifact via a checked-in path or Maven repository; verify SHA-256 at load.

---

---

## F-AIAGENT-21 — `Scratchpad` and reflection memory have no caller isolation; static `KnowledgeBase` is process-global

**Severity floor:** Medium
**Taxonomy:** OWASP-LLM-08, API-API1, CWE-639

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/brain/Scratchpad.java:9-58`

### Evidence pattern

```java
public class Scratchpad {
    private final List<MemoryEntry> entries = new ArrayList<>();
```

### Why it matters

`Scratchpad` is not a Spring bean by default and is not yet shared, so the immediate cross-caller leak is latent — but the API exposes no caller binding, so the moment a consumer makes it `@Component` (the natural step) every caller sees every other caller's memory entries. `KnowledgeBase.ENTRIES` is `static` and `KeywordKnowledgeRetriever` operates on it directly with no tenant filter; the same data is returned to every caller regardless of authority. For the keyword (no-vector-store) path this is the equivalent of TRABUCO-004 missing tenant isolation.

### Suggested fix

Bind `Scratchpad` to a caller key when promoting to a bean; replace `KnowledgeBase` static list with a per-tenant store gated on `CallerContext`.

---

---

## F-AIAGENT-22 — `TaskManager` retains tasks indefinitely; no caller binding on `getTask` or `subscribe`

**Severity floor:** Medium
**Taxonomy:** API-API1, OWASP-LLM-02, CWE-639, CWE-400

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/task/TaskManager.java:14-75`

### Evidence pattern

```java
private final ConcurrentHashMap<String, TaskRecord> tasks = new ConcurrentHashMap<>();
...
public TaskRecord getTask(String taskId) { return tasks.get(taskId); }
public void subscribe(String taskId, Consumer<TaskEvent> listener) { ... }
```

### Why it matters

Anyone who knows or guesses a task ID (UUID — random but enumerable in monitoring/logs) can read task results. `A2AController.handleTasksGet` and `StreamingController.streamTask` look up by ID and never check that `CallerContext.get()` originally submitted the task. Task results may include LLM responses with PII or proprietary content. There's also no eviction — the map grows forever.

### Suggested fix

Stamp each task with submitter's `keyHash`; require requester to match on read/stream; add TTL eviction.

---

---

## F-JAVA-04 — `OutputGuardrailAdvisor` PII regex has malformed character class `[A-Z|a-z]` (matches the literal pipe), uses non-anchored `\b` patterns, and replaces the **entire** response on any match — bypass + denial-of-utility

**Severity floor:** Medium
**Taxonomy:** JAVA-006, OWASP-A04, ASVS-V5-1, CWE-1333, CWE-697

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/security/OutputGuardrailAdvisor.java:17-21,30-44`

### Evidence pattern

```java
private static final List<Pattern> PII_PATTERNS = List.of(
    Pattern.compile("\\b[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\\.[A-Z|a-z]{2,}\\b"), // email — note [A-Z|a-z]
    Pattern.compile("\\b\\d{3}[-.]?\\d{2}[-.]?\\d{4}\\b"),                       // SSN
    Pattern.compile("\\b\\d{4}[- ]?\\d{4}[- ]?\\d{4}[- ]?\\d{4}\\b"),           // credit card
    Pattern.compile("\\b(?:\\+1)?\\(?\\d{3}\\)?[-. ]?\\d{3}[-. ]?\\d{4}\\b")     // phone
);
...
for (Pattern pattern : PII_PATTERNS) {
    if (pattern.matcher(output).find()) {
        return SAFE_FALLBACK;                                                      // entire response dropped
    }
}
```

### Why it matters

1. **Char-class typo:** `[A-Z|a-z]` is a character class containing the literal pipe `|`; the intent was `[A-Za-z]`. Side effect: `foo@example.com|something` matches but `foo@example.museum` (5-letter TLD with no pipe character class needed) still matches because `a-z` covers it — so the typo passes test cases but signals copy-paste / unreviewed regex.
2. **Bypass via Unicode:** Email, SSN, credit-card, phone patterns are ASCII-only. A response like `john[at]example.com` or `social: 123-45-6789` (replace any digit with the U+FF11 fullwidth digit) matches none of them. This is a known LLM PII-redaction bypass — guardrails of this shape provide *false confidence*.
3. **Phone pattern is a ReDoS amplifier on adversarial input** because of nested optional groups `(?:\\+1)?\\(?\\d{3}\\)?[-. ]?\\d{3}[-. ]?\\d{4}` — input like `+1(123) 456-7890.....-..` makes the engine backtrack across the trailing `[-. ]?` alternations. Java's NFA regex engine is not catastrophic on this exact shape but the cost is super-linear.
4. **Denial of utility:** Any single `find()` hit replaces the entire LLM response with `SAFE_FALLBACK`. The chosen patterns false-positive on common product-content (1-800 phone numbers, customer-service email links, order numbers in `\d{4}-\d{4}-\d{4}-\d{4}` format), so the agent produces the canned apology instead of the real answer for benign queries — pushing operators to disable guardrails entirely, which is worse.

### Suggested fix

Replace ad-hoc regex with `org.apache.commons.text` or a vetted PII detector (Microsoft Presidio over a tool API), redact matches in place rather than dropping the whole response, anchor patterns or use Unicode-aware classes, and unit-test against an adversarial corpus including fullwidth/lookalike digits and obfuscation (`[at]`, `[dot]`).

---

---

## F-JAVA-06 — `WebhookManager.dispatch` HMAC uses `wh.apiKey()` (the literal partner-secret-key string) and re-instantiates `Mac` per-call without `MessageDigest.isEqual`-style timing safety; `HmacSHA256` initialization in a hot loop with no key-derivation

**Severity floor:** Medium
**Taxonomy:** OWASP-A02, ASVS-V6-2, CWE-330, CWE-916

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/event/WebhookManager.java:75-84`

### Evidence pattern

```java
private String hmacSha256(String key, String data) {
    try {
        Mac mac = Mac.getInstance("HmacSHA256");
        mac.init(new SecretKeySpec(key.getBytes(StandardCharsets.UTF_8), "HmacSHA256"));
        byte[] hash = mac.doFinal(data.getBytes(StandardCharsets.UTF_8));
        return HexFormat.of().formatHex(hash);
    } catch (Exception e) {
        throw new RuntimeException("HMAC computation failed", e);
    }
}
```

### Why it matters

The HMAC primitive itself is correct, but **the key is not a secret** — it's the literal `keyLabel()` from the bake-in API key map (`"partner-secret-key"` per F-AIAGENT-01 / F-AUTH-03), so the HMAC tag is forgeable by anyone who reads the source. This is documented in F-AIAGENT-09 (aiagent.md) at the *use site*. The platform-level concerns this finding adds:
- `Mac.getInstance("HmacSHA256")` is invoked per dispatch (no caching) — costs ~50 µs each call; not security-critical alone but interacts with F-JAVA-05 unbounded fan-out.
- Verification uses byte-level string comparison via the receiver's `hex` (caller code, not generated) — no `MessageDigest.isEqual` recommendation in the generated docs/headers, no example of constant-time comparison shipped to teach the integrator.
- No nonce / replay protection in the signed envelope (only `event` + `timestamp` + `data`). Receiver cannot detect replay without storing all seen `X-Webhook-ID` values.

### Suggested fix

Generate a per-webhook 256-bit `SecureRandom` secret at registration, return it once (never store plaintext), persist `argon2id` hash; sign with that secret. Include a monotonic `nonce` in the canonical-string body. Ship a receiver-side example using `MessageDigest.isEqual`.

---

---

## F-JAVA-07 — `ApiKeyAuthFilter.sha256()` correctly uses SHA-256 but the hash is computed AFTER plaintext lookup and never compared with `MessageDigest.isEqual`; no `SecureRandom` is used anywhere for token-shaped values (only `UUID.randomUUID`, which is sufficient for non-secret IDs but is the wrong primitive if any future code mistakes the value for a secret)

**Severity floor:** Medium
**Taxonomy:** OWASP-A02, OWASP-A07, ASVS-V2-2, ASVS-V6-3, CWE-208, CWE-330

### Evidence pattern

```java
// ApiKeyAuthFilter
String key = authHeader.substring(7).trim();
KeyEntry entry = API_KEYS.get(key);                  // line 71 — Map.get; String.equals; timing-vulnerable
...
String keyHash = sha256(key);                        // computed AFTER successful lookup, used only for logging
```

### Why it matters

- The `sha256()` helper exists, indicating intent to hash — but it's invoked *after* the plaintext map lookup that already performed the equality check. An attacker times the lookup, not the hash. F-AUTH-* covers the timing aspect at the auth layer; the platform-level note here is that the `MessageDigest.getInstance("SHA-256")` import (line 16) is technically unused for security purposes — a static-analysis tool grepping for `MessageDigest.isEqual` finds zero usages across the entire codebase, confirming no constant-time comparison is performed for any secret-shaped value.
- `UUID.randomUUID()` in Java is backed by `SecureRandom` internally (per JDK source) — so the use sites for correlation IDs, task IDs, webhook IDs, and entity IDs are *cryptographically reasonable* even though the API surface is `java.util.UUID`. **However,** if a developer follows the existing pattern when introducing a new secret (e.g., a per-webhook signing secret per F-JAVA-06's recommendation), the obvious "copy what the codebase does" path is `UUID.randomUUID().toString()` — which is 122 bits of entropy, less than the 256-bit secret demanded by webhook signing. There is no shipped utility (`Tokens.newSecret()` with `SecureRandom + Base64`) and no skill telling the integrator to use one.

### Suggested fix

Add a `Shared/security/SecureTokens.java` helper (`SecureRandom.nextBytes(new byte[32])` → `Base64.getUrlEncoder().withoutPadding().encodeToString(...)`); document its use in the `add-tool`/`add-endpoint` skills. Replace the unused `MessageDigest` import in `ApiKeyAuthFilter` with the lookup-by-hash pattern (auth.md owns the active fix).

---

---

## F-JAVA-08 — Inline `new ObjectMapper()` in two AIAgent components bypasses the Spring-managed mapper that has Boot's hardening; no `StreamReadConstraints` / `maxNestingDepth` / `maxDocumentLength` set anywhere — JSON-bomb DoS (JAVA-010)

**Severity floor:** Medium
**Taxonomy:** JAVA-010, OWASP-A04, API-API4, ASVS-V13-1, CWE-776, CWE-400

### Evidence pattern

```java
public class StreamingController {
    ...
    private final ObjectMapper objectMapper = new ObjectMapper();        // not the Spring bean
}

public class WebhookManager {
    ...
    private final ObjectMapper objectMapper = new ObjectMapper();
}
```

### Why it matters

1. **Inline mapper bypass**: `new ObjectMapper()` is a fresh mapper without any of the Spring-Boot-applied modules (`Jdk8Module`, `JavaTimeModule`, `ParameterNamesModule`) or `Jackson2ObjectMapperBuilder` customizations. Anywhere a custom `Jackson2ObjectMapperBuilderCustomizer` is added (e.g., to register a `StreamReadConstraints` or `BasicPolymorphicTypeValidator`), these inline mappers don't get it. They will silently behave differently from the autowired mapper — stamping out a copy of `LocalDateTime` as `[2026, 5, 1, ...]` (default-array form) instead of ISO-8601, etc.
2. **No depth/length limits configured anywhere**: `grep -rn "StreamReadConstraints\|maxNestingDepth\|maxDocumentLength"` returns zero results. Jackson 2.16+ ships defaults of `maxNestingDepth=1000`, `maxDocumentLength=-1` (unlimited), `maxStringLength=20_000_000`. A `POST /a2a` body of `{"a":{"a":{...500 deep}}}` is permitted by default — combined with virtual threads (carrier-pinning is unlikely here but parsing CPU is not) plus the `Map<String,Object>` untyped bag from F-JAVA-02, an attacker exhausts CPU + heap with one request. No request-body size limit (`spring.servlet.multipart.max-request-size` — applies to multipart only — is unset; `server.tomcat.max-http-form-post-size` not configured; `server.tomcat.max-swallow-size` not configured).
3. **`@JsonInclude` not configured globally**: response DTOs that bind `Map<String, Object>` (e.g., `IngestionController.ingest` returns `ResponseEntity<Map<String, Object>>`) will serialize null values, leaking stack-trace-like keys with `null` values when the contract changes.

### Suggested fix

Inject the Spring `ObjectMapper` via constructor in both classes; add a `@Configuration` bean:
```java
@Bean
public Jackson2ObjectMapperBuilderCustomizer hardenJackson() {
    return b -> b.factory(JsonFactory.builder()
        .streamReadConstraints(StreamReadConstraints.builder()
            .maxNestingDepth(64)
            .maxDocumentLength(1_000_000)
            .maxStringLength(100_000)
            .build())
        .build());
}
```
Configure `server.tomcat.max-swallow-size=2MB` in `application.yml`.

---

---

## F-JAVA-09 — Virtual threads enabled (`spring.threads.virtual.enabled=true`) with raw `ThreadLocal`-based `CallerContext`/`RequestContextHolder` and no `ScopedValue` migration — context leak/loss across carrier-thread reuse for any code path that doesn't use the existing `try/finally` clear (compounds F-AUTH-14)

**Severity floor:** Medium
**Taxonomy:** JAVA-013, OWASP-A09, ASVS-V8-1, CWE-488, CWE-668

### Evidence pattern

```java
// TaskManager — does NOT capture CallerContext
private final ExecutorService executor = Executors.newVirtualThreadPerTaskExecutor();
public String submitTask(Callable<Map<String, Object>> work) {
    String taskId = "TASK-" + UUID.randomUUID();
    var record = new TaskRecord(taskId);
    tasks.put(taskId, record);
    executor.submit(() -> {
        record.setStatus("working");
        notify(taskId, "working", null, null);              // CallerContext.get() here ⇒ anonymous
        try {
            Map<String, Object> result = work.call();        // any tool call inside loses caller identity
        } catch (Exception e) { ... }
    });
    return taskId;
}
```

### Why it matters

The agent's task submission boundary explicitly enters a virtual-thread executor without capturing `CallerContext`. Anywhere the `Callable<Map<String, Object>> work` argument transitively reads `CallerContext.get()` — e.g., a tool that needs to respect partner-tier rate limits, a vector-store query that filters by tenant, an A2A onward call — receives `CallerIdentity.anonymous()`. Two distinct failure modes:
1. **Loss:** anonymous identity ⇒ tool denies legitimate work or, worse, uses a default permissive path (silent privilege downgrade that confuses operators auditing logs).
2. **Leak:** if any code path mounts the same virtual thread on a carrier that previously ran a different request and *forgets to clear*, identity carries over. F-AUTH-14 documents the JWT-side leak; the platform-level concern is that `TaskManager` does not even have an `AuthScope` — there's no place to clear because there's no place to set, so the leak is from the *prior* HTTP request's filter context if the carrier thread was its carrier.
3. **`MDC` propagation:** `CorrelationIdFilter` writes `MDC.put("correlationId", ...)` then `MDC.remove(...)` in finally. With virtual threads, MDC is per-thread (Logback uses `InheritableThreadLocal` with strategy `LogbackMDCAdapter`), so the executor-submitted Callable runs without the correlation ID — log lines from `TaskManager.notify` come through with `[correlationId-]`. Auditing a partner's task lifecycle is broken.

### Suggested fix

- Capture `CallerContext`/`MDC` at submission and restore inside the lambda:
```java
CallerIdentity caller = CallerContext.get();
Map<String, String> mdc = MDC.getCopyOfContextMap();
executor.submit(() -> {
    CallerContext.set(caller);
    if (mdc != null) MDC.setContextMap(mdc);
    try { ... } finally { CallerContext.clear(); MDC.clear(); }
});
```
- Long-term: migrate to `ScopedValue<CallerIdentity>` (Java 21 preview / Java 25 final) which is virtual-thread-safe by design; the `RequestContextHolder` Javadoc already notes this migration but no code does it.

---

---

## F-JAVA-10 — `ResponseStatusException` reflects user-controlled input back into RFC 7807 `detail` (`"Task not found: " + taskId`, `"Unknown skill: " + skill`, `"Unknown method: " + method`, `"Webhook not found: " + webhookId`) — combined with the absence of HTML-escape on the detail field, partial reflective injection

**Severity floor:** Medium
**Taxonomy:** OWASP-A03, ASVS-V5-2, ASVS-V7-1, CWE-79, CWE-209, TRABUCO-002

### Evidence pattern

```java
// A2AController
return JsonRpcResponse.error(rpcId, -32601, "Unknown method: " + method);
return JsonRpcResponse.error(rpcId, -32601, "Unknown skill: " + skill);
throw new ResponseStatusException(HttpStatus.NOT_FOUND, "Task not found: " + taskId);

// WebhookManager
public void deregister(String webhookId) {
    if (webhooks.remove(webhookId) == null) {
        throw new IllegalArgumentException("Webhook not found: " + webhookId);
    }
}
// WebhookController
catch (IllegalArgumentException e) {
    throw new ResponseStatusException(HttpStatus.NOT_FOUND, e.getMessage());     // forwards user-controlled
}
```

### Why it matters

Spring's `ResponseStatusException` populates `ProblemDetail.detail` with the `reason`, which is then serialized as `application/problem+json`. The `detail` field is reflected verbatim with no HTML escape — RFC 7807 doesn't require it because the content type is JSON, but consumers (browser-tab error pages, log dashboards, status-page widgets) frequently render the JSON body as HTML or pass the detail string through a Markdown/HTML pipeline. An attacker calling `POST /a2a {"method":"<script>alert(1)</script>",...}` gets `detail: "Unknown method: <script>alert(1)</script>"` echoed back. Combined with `Content-Type: application/problem+json` not always preserved by intermediaries, this is a low-risk reflected-XSS / log-forging vector. F-AUTH (auth.md) covers ProblemDetail leakage of stack traces; this is the orthogonal *user-input-reflection* concern.

### Suggested fix

Either truncate to a known shape (`"Task not found"` with no ID echo — log the ID with the correlation ID instead) or HTML-escape the reflected portion via `HtmlUtils.htmlEscape` before insertion, even though the response is JSON, to defend against downstream rendering.

---

---

## F-JAVA-12 — `InputGuardrailAdvisor.classify` is fail-open on exception, uses `String.format` to inject user input into a prompt (prompt injection), and `userInput.substring(0, Math.min(50, userInput.length()))` may split a UTF-16 surrogate pair → `StringIndexOutOfBoundsException` on emoji input → fail-open path triggered

**Severity floor:** Medium
**Taxonomy:** JAVA-006, JAVA-013, OWASP-A04, LLM-01, ASVS-V5-1, CWE-755, CWE-1335

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/security/InputGuardrailAdvisor.java:60-79`

### Evidence pattern

```java
public String classify(String userInput) {
    try {
        String response = this.chatClient.prompt()
            .user(CLASSIFICATION_PROMPT.formatted(userInput))         // user input into prompt template
            .call()
            .content();
        if (response != null && response.toUpperCase().contains("BLOCKED")) {
            log.warn("Input blocked by guardrail: {}", userInput.substring(0, Math.min(50, userInput.length())));
            ...
        }
        return null; // allowed
    } catch (Exception e) {
        log.error("Guardrail classification failed, allowing input: {}", e.getMessage());
        return null; // fail-open for availability; production might fail-closed
    }
}
```

### Why it matters

1. **Fail-open:** any exception (LLM rate-limit, network blip, JSON-parse fail in the upstream Spring AI client) silently bypasses the guardrail. The comment acknowledges "production might fail-closed" but ships the wrong default. An attacker who can induce a single LLM error (e.g., by sending a 100k-char input that exceeds the max tokens) bypasses the guardrail entirely.
2. **Prompt injection delivery:** `CLASSIFICATION_PROMPT.formatted(userInput)` is a textbook prompt-injection delivery vector — covered semantically by aiagent.md's LLM section, but the platform mechanism is `String.format`/`.formatted` with no escape, no fence, no role separation. (aiagent.md owns the LLM-side prompt-engineering finding.)
3. **`substring` surrogate-pair UTF-16 hazard:** `userInput.substring(0, 50)` cuts at the 50th UTF-16 code unit. If the 50th and 51st code units form a surrogate pair (any character above U+FFFF — emoji, many CJK extensions), the cut leaves an unpaired high-surrogate. `String.toString` is fine, but logging encoders (Logstash JSON encoder used in non-local profiles) can throw or emit malformed JSON when the surrogate cannot be UTF-8 encoded. The encoder failure raises an exception inside the `log.warn` call, which propagates up — though wrapped in the `catch (Exception e)` it sends control flow into the fail-open branch, **flipping a BLOCKED classification into an ALLOWED one**.
4. **`response.toUpperCase()`** — locale-sensitive (Turkish locale's dotless-i flips `i ⇒ İ` ≠ `I`); `BLOCKED` could fail to match if the JVM default locale is `tr-TR`. Use `toUpperCase(Locale.ROOT)`.

### Suggested fix

Fail-closed by default (`return "Input classification temporarily unavailable; try again."` on exception), use a UTF-16-aware truncation (`userInput.codePoints().limit(50).collect(...)`), use `.toUpperCase(Locale.ROOT)`, separate user input from the system prompt with a structured tool-call instead of string concat (aiagent.md owns the prompt fix).

---

---

## F-JAVA-13 — `ChatClient` and the underlying Spring AI Anthropic / Transformer / PgVector clients are configured purely via `application.yml` properties — there's no shipped HTTP-client hardening (TLS floor, timeouts, retry-with-jitter) and no test that proxies through a corporate forward proxy; defaults take effect silently

**Severity floor:** Medium
**Taxonomy:** JAVA-008, JAVA-011, OWASP-A02, OWASP-A06, LLM-03, ASVS-V12-6, CWE-295, CWE-1188

### Evidence pattern

```yaml
spring:
  ai:
    anthropic:
      api-key: ${ANTHROPIC_API_KEY:}
      chat:
        options:
          model: ${AI_MODEL:claude-sonnet-4-20250514}
          max-tokens: ${AI_MAX_TOKENS:1024}
    retry:
      max-attempts: ${AI_RETRY_MAX_ATTEMPTS:2}
```

### Why it matters

- Spring AI's Anthropic client uses `RestClient` underneath (via `AnthropicApi`), which itself uses `JdkClientHttpRequestFactory` (Java 11 `HttpClient`). The `HttpClient` is built with default redirect policy `NEVER` (good), but no connect/read/total timeout — if the Anthropic API hangs, the call hangs until the OS socket idle timeout (~minutes). With virtual threads, hung threads are cheap, but back-pressure on the agent loop and rate-limit-token consumption is real.
- No `Retry` jitter / max-delay configured. `spring.ai.retry.max-attempts=2` is shipped but `spring.ai.retry.backoff.initial-interval`, `multiplier`, `max-interval`, `jitter` are all defaulted. Default jitter is 0 → thundering herd on Anthropic outage.
- No test or doc that the user's outbound-proxy / TLS pinning / corporate CA bundle is wired. A user behind a Zscaler/MITM proxy needs `-Djavax.net.ssl.trustStore` plus `HTTP_PROXY` env vars; without instructions, the obvious fix some teams will reach for is `TrustAll` (JAVA-008). No `TrustAll` is present *in the generated code* (good — false-positive considered below), but the absence of explicit guidance leads operators to add it.
- The transformer ONNX model downloads `intfloat/e5-small-v2` from HuggingFace at runtime (per `application.yml` comment), no checksum / SBOM / pinned hash. Spring AI's downloader uses `~/.djl.ai/` and trusts whatever HF returns. **LLM-03 supply-chain** material — covered semantically by aiagent.md but the JVM-platform expression is "no Maven-side dep, runtime resource fetch over HTTPS without integrity check."

### Suggested fix

Ship a `RestClientCustomizer` that pins TLSv1.3, sets `Duration.ofSeconds(30)` connect/`Duration.ofSeconds(120)` read timeouts, and a `RetryTemplate` with `ExponentialRandomBackOffPolicy` (jitter). Document `HTTPS_PROXY` / corporate CA. Pin the embedding model digest (`spring.ai.embedding.transformer.onnx.model-uri` to a hash-anchored CDN) or download it at build time via a Maven plugin.

---

---

## F-AIAGENT-11 — System prompts include the project name and a developer TODO list visible to the model

**Severity floor:** Low
**Taxonomy:** OWASP-LLM-07, LLM-EXT-01, CWE-200

### Where to look

- `AIAgent/src/main/java/com/security/audit/aiagent/agent/PrimaryAgent.java:49-67`

### Evidence pattern

```java
private static final String SYSTEM_PROMPT = """
    You are the AiagentPgvectorRag AI assistant.
    ...
    TODO: Customize this system prompt for your domain:
    - Describe what the agent can help with
    - List the main capabilities and constraints
    - Define the agent's personality and tone
    - Specify any rules (e.g., topics to decline, data the agent should never fabricate)
    Rules:
    - Only discuss topics relevant to your domain. Politely decline off-topic requests.
    ...
    """;
```

### Why it matters

The system prompts ship to production unmodified by default (the TODO is a placeholder). System prompts are extractable via prompt injection (LLM-07), and what they currently leak is mostly low-impact (project name, existence of guardrails, existence of a specialist agent). However the `InputGuardrailAdvisor` prompt enumerates the exact bypass tokens it's looking for ("ignore previous instructions", "system:") — telling an attacker which strings to obfuscate. The reflection prompt enumerates the action tokens (`RETRY`, `ESCALATE`, `GIVE_UP`).

### Suggested fix

Treat guardrail/reflection prompts as security-sensitive (avoid enumerating bypass tokens verbatim); replace TODO scaffolding with a runtime startup check that warns if the default prompt is unchanged.

---

---

## F-AIAGENT-18 — CORS `allowedOrigins` parsed by `String.split(",")` with no trim/validation

**Severity floor:** Low
**Taxonomy:** SPRING-007, API-API8, OWASP-A05

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/config/WebConfig.java:16-35`

### Evidence pattern

```java
@Value("${cors.allowed-origins:http://localhost:3000,http://localhost:8080}")
private String allowedOrigins;
...
.allowedOrigins(allowedOrigins.split(","))
```

### Why it matters

`allowCredentials=false` so the worst-case wildcard issue is not present, but a misconfigured `${CORS_ALLOWED_ORIGINS:*}` env var would be accepted as `["*"]`. The default values include `http://` (non-TLS) origins. `Authorization` is in `allowedHeaders`, so a dev who flips `allow-credentials=true` (env-var override is exposed) immediately exposes credentialed cross-origin reads from any HTTP origin.

### Suggested fix

Validate allowed-origins entries reject `*` when credentials true; reject non-`https` origins outside `localhost`; trim whitespace.

---

---

## F-AIAGENT-20 — `AgentRestController` `chat`/`ask` endpoints fall back to a static reply when LLM is missing — bypasses guardrails silently

**Severity floor:** Low
**Taxonomy:** OWASP-LLM-09, ASVS-V10-1

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/protocol/AgentRestController.java:52-77`

### Evidence pattern

```java
if (primaryAgent != null) { ... }
return ImmutableChatResponse.builder().response("Agent brain requires ANTHROPIC_API_KEY.").blocked(false).build();
```

### Why it matters

The boolean `blocked` is set to `false` when the response is the placeholder string — the client cannot distinguish "request was processed and is safe" from "request was silently no-op'd because the API key isn't set". A monitoring system that checks `blocked` cannot detect the misconfiguration. More importantly, when the API key is set but `inputGuardrail` is null (e.g., `@Qualifier` mismatch silently fails), the controller proceeds without guardrails — there is no boot-time assertion that all expected components are wired.

### Suggested fix

Return 503 when `primaryAgent == null`; assert at boot that guardrails are present when `agent.guardrails.enabled=true`.

---

---

## F-AIAGENT-23 — `agent.guardrails.enabled` configuration is documented but never read

**Severity floor:** Low
**Taxonomy:** OWASP-A04, ASVS-V14-1, OWASP-A05

### Where to look

`AIAgent/src/main/resources/application.yml:130-131` vs. guardrail bean declarations.

### Evidence pattern

```yaml
guardrails:
  enabled: ${GUARDRAILS_ENABLED:true}
```

### Why it matters

Operators changing `GUARDRAILS_ENABLED=false` expect guardrails off; they remain on (or, more dangerously, an operator who flips this thinking it's a kill-switch ships to prod believing they have a way to disable guardrails for incident response — they don't). This is a configuration-drift / "fake setting" defect.

### Suggested fix

Honor the property with `@ConditionalOnProperty(name = "agent.guardrails.enabled", havingValue = "true", matchIfMissing = true)`; or remove from application.yml.

---

---

## F-AIAGENT-24 — `ApiKeyAuthFilter` sends 403 for unknown keys instead of 401; doesn't differentiate "missing" from "invalid"

**Severity floor:** Low
**Taxonomy:** API-API2, ASVS-V2-2, CWE-287

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/security/ApiKeyAuthFilter.java:65-76`

### Evidence pattern

```java
if (!authHeader.toLowerCase().startsWith("bearer ")) {
    sendError(response, HttpStatus.FORBIDDEN, "Invalid API key");
...
if (entry == null) {
    sendError(response, HttpStatus.FORBIDDEN, "Invalid API key");
```

### Why it matters

`403 Forbidden` is the wrong status for authentication failure (RFC 9110: 401 for missing/invalid credentials, 403 for valid credentials lacking permission). Auth metrics that watch 401 won't fire; existing monitoring dashboards expect the standard semantic. There's also no `WWW-Authenticate` response header.

### Suggested fix

Return 401 with `WWW-Authenticate: Bearer realm=...` for unknown keys.

---

---

## F-AIAGENT-25 — `agent.json` discovery file advertises the agent's URL and skill list publicly

**Severity floor:** Low
**Taxonomy:** API-API9, LLM-EXT-01, TRABUCO-008

### Where to look

`AIAgent/src/main/resources/.well-known/agent.json`

### Evidence pattern

```json
{
  "url": "http://localhost:8080",
  "skills": [{"id":"list_items"}, {"id":"get_item_detail"}, {"id":"check_availability"}, {"id":"ask_question"}, {"id":"chat"}],
  "authentication": {"schemes": ["bearer"]}
}
```

### Why it matters

Served by `WebConfig.addResourceHandlers` from `classpath:.well-known/` and explicitly `permitAll`'d in `AgentSecurityConfig` even when JWT auth is on. The hardcoded `http://localhost:8080` URL is a templating defect; the published skill list lets unauthenticated attackers map the tool surface.

### Suggested fix

Make `url` resolve from `server.address`/`spring.application.name` at runtime; consider gating the file behind authentication for non-public agents.

---

---

## F-AIAGENT-26 — Anthropic API key default is empty; no fail-fast at boot when the agent is expected to function

**Severity floor:** Low
**Taxonomy:** ASVS-V14-1, ASVS-V10-1, OWASP-A05

### Where to look

`AIAgent/src/main/resources/application.yml:49-50`

### Evidence pattern

```yaml
anthropic:
  api-key: ${ANTHROPIC_API_KEY:}
```

### Why it matters

Default empty key means Spring AI auto-configures a `ChatModel` bean that will fail on first call (silent degradation to F-20 fallback string). Combined with `InputGuardrailAdvisor` using the same model, guardrails also silently degrade. There's no explicit boot-time check.

### Suggested fix

Fail boot with a clear message when `ANTHROPIC_API_KEY` is empty unless an explicit "demo mode" flag is set.

---

## False positives considered

- **`InputGuardrailAdvisor.classify` `String.format`** — `String.format("%s", userInput)` is interpolation, not SQL/SpEL injection; the result is text passed to the model, not executed code. Already covered as part of F-04 (prompt template fences) and F-10 (fail-open).
- **`OutputGuardrailAdvisor` regex character class `[A-Z|a-z]`** — looked like ReDoS but the regex has bounded quantifiers; the `|` is a literal in the character class (a known false-positive pattern but not a runtime vulnerability beyond the matching imprecision already noted in F-10).
- **`@CircuitBreaker` masking auth errors** — circuit breaker's fallback returns a generic string but does not mask 4xx auth failures because Spring Security throws before the controller method is invoked; not a finding.
- **`VectorStore.add` accepting `Document.metadata`** — looked like a SQL/JSON injection vector, but Spring AI's PgVectorStore parameter-binds JSONB metadata with the JDBC driver. Filed only the missing tenant column (F-03), not a separate injection finding.
- **`A2AController.handle` switch on `method`** — JSON-RPC method dispatch is allowlisted; unknown methods return -32601. Not a finding.
- **`ResponseStatusException(HttpStatus.NOT_FOUND, "Task not found: " + taskId)` in `StreamingController`** — leaks supplied taskId back, but taskId is the caller's own input; standard echo, not an information leak. Wider authz issue is captured in F-22.
- **`MessageDigest.getInstance("SHA-256")`** for hashing the API key in `ApiKeyAuthFilter.sha256` — the hash is fine; the upstream comparison happens *before* the hash is computed (against plaintext map), so the hash adds nothing. Already covered by F-01.
- **`@RequestParam` SpEL injection** — none of the controllers use SpEL on request input; `@PreAuthorize` strings are static. Not a finding.
- **`ObjectMapper` default typing** — only `new ObjectMapper()` (default disabled typing) is used; not a Jackson polymorphic finding.

---

## Summary

26 findings recorded across the AIAgent module's LLM surface. Both archetypes share identical Java code; differences are only in `application.yml` for the vector store backend and parent POM module list, so all findings apply to both targets.

---

## F-JAVA-11 — Default `X-XSS-Protection: 1; mode=block` header is OWASP-deprecated and re-introduces side-channel info-leak in older browsers — modern guidance is `0` or omit

**Severity floor:** Low
**Taxonomy:** OWASP-A05, API-API8, ASVS-V14-4, CWE-1021, CWE-1188

### Where to look

`API/src/main/java/com/security/audit/api/config/SecurityHeadersFilter.java:34-35,57-58`

### Evidence pattern

```java
@Value("${security.headers.x-xss-protection:1; mode=block}")
private String xXssProtection;
...
httpResponse.setHeader("X-XSS-Protection", xXssProtection);
```

### Why it matters

Per OWASP Cheat Sheet (Secure Headers Project) and Mozilla MDN, `X-XSS-Protection: 1; mode=block` is *worse than not setting the header at all*: it triggers a buggy reflective filter in legacy IE/Edge that has been used as a side-channel oracle for cross-origin information disclosure. All modern browsers ignore the header; legacy browsers that respect `1; mode=block` are vulnerable. The current OWASP recommendation is `X-XSS-Protection: 0` (explicitly disable) and rely on a strict CSP. The template-default is the wrong direction.

### Suggested fix

Change the default in `SecurityHeadersFilter` to `0` and let the operator override only if they specifically need to support a legacy-browser environment. Update web.md / docs/security-headers.md to match.

---

---

## F-JAVA-14 — `Pattern.compile` calls produce static patterns from constants only (no user-controlled regex), but `RateLimiter.checkRateLimit` uses `System.currentTimeMillis()` and a per-call linear scan of the deque — wall-clock-skew + linear scan makes the limiter unreliable and itself a CPU vector under burst

**Severity floor:** Low
**Taxonomy:** JAVA-006, JAVA-018, OWASP-A04, ASVS-V11-1, CWE-1333, CWE-400

### Where to look

`AIAgent/src/main/java/com/security/audit/aiagent/security/RateLimiter.java:25-42`

### Evidence pattern

```java
public void checkRateLimit(CallerIdentity caller) {
    int limit = LIMITS.getOrDefault(caller.tier(), 10);
    long now = System.currentTimeMillis();                     // wall clock, not monotonic
    long windowStart = now - 60_000;
    Deque<Long> window = store.computeIfAbsent(caller.keyHash(), k -> new ConcurrentLinkedDeque<>());
    while (!window.isEmpty() && window.peekFirst() < windowStart) {
        window.pollFirst();
    }
    if (window.size() >= limit) { throw new ResponseStatusException(HttpStatus.TOO_MANY_REQUESTS, ...); }
    window.addLast(now);
}
```

### Why it matters

- `System.currentTimeMillis()` is wall-clock, subject to NTP step adjustment. A 30-s clock-step backward causes every cached entry to look "in the future," `peekFirst() < windowStart` is false for everything, and `window.size() >= limit` rapidly trips on the next 200 partner calls — denying legitimate traffic. Use `System.nanoTime()` (monotonic) or a monotonic millisecond clock.
- `ConcurrentLinkedDeque.size()` is **O(n)** per Javadoc — every check walks the entire window. At 200 RPM (partner tier ceiling) the deque holds up to 200 longs; per-call cost is ~µs and dominated by the size scan. Not a critical hazard, but a pathological burst (200 in 1 ms) makes every subsequent check do 200 size-scans before the first eviction.
- No multi-instance coordination — running two AIAgent replicas behind a load balancer doubles the effective rate; the in-memory deque is per-JVM. Document or move to Bucket4j-on-Hazelcast / Redis (the codebase already pins `bucket4j.version=0.12.7` in the parent POM but no module imports it).

### Suggested fix

Replace the deque with `Bucket4j.builder().addLimit(...)`, pre-bind `bucket4j-redis` for shared rate state, keep monotonic time, document the per-replica caveat.

---

---

## F-JAVA-15 — `JobRequestHandler` base/override pattern (Model defines no-op base, Worker `@Component` extends and overrides) relies on JobRunr's Jackson-backed serialization of `ProcessPlaceholderJobRequest` — sealed-interface hierarchy plus future `@JsonTypeInfo(use=Id.CLASS)` would be a JAVA-001 step away

**Severity floor:** Low
**Taxonomy:** JAVA-001, JAVA-003, OWASP-A08, CWE-502, CWE-915

### Evidence pattern

```java
// Model
public sealed interface PlaceholderJobRequest extends JobRequest permits ProcessPlaceholderJobRequest { }
public record ProcessPlaceholderJobRequest(String message) implements PlaceholderJobRequest {
    @Override public Class<ProcessPlaceholderJobRequestHandler> getJobRequestHandler() {
        return ProcessPlaceholderJobRequestHandler.class;
    }
}
public class ProcessPlaceholderJobRequestHandler implements JobRequestHandler<ProcessPlaceholderJobRequest> {
    @Override public void run(ProcessPlaceholderJobRequest request) { /* no-op */ }
}
```

### Why it matters

JobRunr 8.x serializes job arguments via Jackson and stores the class FQN (`org.jobrunr.utils.mapper.jackson.JacksonJsonMapper`) in the `jobAsJson` column. On dequeue, the Jackson default-typing-equivalent path reads the FQN and instantiates — same gadget surface as Spring AMQP F-JAVA-01. The shipped sealed-interface design is *currently safe* because:
- only `record ProcessPlaceholderJobRequest(String message)` is permitted,
- `String message` is the only field,
- JobRunr's `JacksonJsonMapper` does not enable global default typing.

But the *pattern shipped to integrators* is "add a new record implementing `PlaceholderJobRequest` and a new handler in Worker." A maintainer who later writes `public record FetchAndIngestRequest(URI source, Map<String, Object> hints) implements PlaceholderJobRequest` introduces a polymorphic field (`Map<String, Object>` → JAVA-001) plus an SSRF (`URI source` is dispatched in the worker without validation; the existing skill `add-job` does not warn about this) — the *gate* is operator discipline, not a framework constraint. Additionally, an attacker with database write access (less far-fetched in JobRunr-managed Postgres if the Postgres user is reused) can poison `jobAsJson` to instantiate any class Jackson can resolve.

### Suggested fix

Document the constraint in `add-job` skill — "job request fields must be primitives, records, or fully-typed; never `Map<String,Object>` or `Object`." Add a JobRunr `JsonMapper` configuration that wires a `BasicPolymorphicTypeValidator` allowing only `com.security.audit.model.jobs.*`. Add an integration test that asserts a malicious `jobAsJson` row fails deserialization rather than executing.

---

---

## F-JAVA-16 — No Bean Validation custom regex constraints; built-in `@Size` / `@NotBlank` on `ChatRequest`/`AskRequest` are safe but `@Pattern` is reachable from `add-endpoint` skill — no warning that `@Pattern(regexp = userSupplied)` is forbidden, and Hibernate Validator 8.x ships a non-timeout-bounded regex engine

**Severity floor:** Low
**Taxonomy:** JAVA-006, OWASP-A04, ASVS-V5-1, CWE-1333

### Evidence pattern

```java
public interface ChatRequest {
    @NotBlank(message = "Message is required")
    @Size(max = 10000, message = "Message must not exceed 10,000 characters")
    String message();
}
```

### Why it matters

Current state is safe — Hibernate Validator 8.x (Spring Boot 3.4.2 BOM) does not implement regex timeout, but there are no user-controllable regex inputs in the current generated code. The platform-level concern is preventive: the skill `add-endpoint` doesn't say "if you add `@Pattern`, the regex string must be a compile-time constant; never `@Pattern(regexp = "${user.supplied.regex:}")` — the resulting `Pattern.compile` is unbounded and `(a+)+b`-style adversarial inputs cause catastrophic backtracking with no timeout."

### Suggested fix

Add a paragraph to the `add-endpoint` and `code-quality-guide` skills warning against user-controlled regex; consider linking `re2j` if cross-tenant regex is genuinely needed; document `Pattern.compile(...).matcher(...).results()` with a `CompletableFuture.orTimeout` wrapper if a timeout is required.

---

## False positives considered

- **`new ObjectMapper()` is not Jackson default-typing.** The two inline mappers (`StreamingController.java:22`, `WebhookManager.java:23`) are constructed with no calls to `enableDefaultTyping()` / `activateDefaultTyping()`. They are *unsafe* in the senses listed in F-JAVA-08 (depth limits, Spring-Boot module loss) but **not** the JAVA-001 polymorphic gadget. Logged in F-JAVA-08 with reduced severity rather than as JAVA-001.
- **`@JsonTypeInfo` on `PlaceholderEvent` is *not* JAVA-001.** It pairs `Id.NAME` with an explicit `@JsonSubTypes` allow-list (lines 30-37 of `PlaceholderEvent.java`). Jackson restricts polymorphism to the listed sub-types — same shape Jackson recommends. **However,** see F-JAVA-01: this allow-list is bypassed by the AMQP `__TypeId__` header path which consults `DefaultClassMapper` *before* the sealed-interface deserializer.
- **`UUID.randomUUID()` for correlation/task/webhook IDs is not insecure.** Java's `UUID.randomUUID()` is backed by `SecureRandom` (`SecureRandom.next` for the 122 random bits). Sufficient for non-secret identifiers. Flagged in F-JAVA-07 only as a *teaching-by-example* concern when integrators copy the pattern for actual secrets.
- **No `ObjectInputStream` / `readObject` / RMI / JMX endpoints anywhere.** `grep -rn "ObjectInputStream\|readObject\|RMI\|JMX"` returns no hits in user-scanned packages. JAVA-003 (Java native serialization) is genuinely absent.
- **No XML parsing surface.** No `DocumentBuilderFactory`, `SAXParserFactory`, `TransformerFactory`, `JAXBContext`, `XMLInputFactory`, or `XMLReader` usages anywhere. JAVA-002 / JAVA-010 (XML billion-laughs) is genuinely absent — Trabuco emits JSON-only controllers.
- **No JNDI / Logback `JndiLookup`.** No `InitialContext`, `JndiTemplate`, `ctx.lookup`, or `jndi:` literal anywhere. Logback configuration files contain no JNDI receivers and no socket-receiver. Logback 1.5.x via Spring Boot 3.4.2 BOM is post-CVE-2023-6378. JAVA-004 is genuinely absent.
- **No SnakeYAML usages.** `new Yaml(...)` not present in any module. Spring Boot 3.4.2 includes SnakeYAML 2.x by transitive dep, with `SafeConstructor` as the default. JAVA-005 is genuinely absent.
- **No TrustAll SSL / `X509TrustManager` / `HostnameVerifier` overrides.** `grep -rn "TrustManager\|HostnameVerifier\|trustAll\|InsecureTrustManager\|HttpsURLConnection\.setDefault"` returns zero hits in generated code. JAVA-008 (TrustAll) is genuinely absent. (F-JAVA-13 is the *related* concern of "no positive TLS hardening shipped, so an integrator might add TrustAll later.")
- **No ZIP/TAR extraction.** `grep -rn "ZipInputStream\|ZipFile\|TarInputStream\|getNextEntry"` returns zero hits. JAVA-009 (Zip Slip) is genuinely absent — `DocumentIngestionService.ingest(String text, ...)` accepts plain text only.
- **No `Runtime.exec` / `ProcessBuilder`.** Genuinely absent in generated Java code. (`localstack-init.sh` is shell, not JVM.)
- **No SpEL evaluation of user input.** `@PreAuthorize` strings are static literals (`hasAuthority('SCOPE_admin')`), not concatenated. No `SpelExpressionParser` / `StandardEvaluationContext` usage. `@Value` strings are config-property literals with safe defaults. JAVA-017 (Spring4Shell-class) is genuinely absent.
- **`OutputGuardrailAdvisor` regex ReDoS is not catastrophic.** Tested mentally and against the corpus: the four patterns are non-nested-quantifier and Java's NFA engine handles them without exponential blowup. F-JAVA-04 is **bypass + denial-of-utility**, not ReDoS.
- **`SecurityHeadersFilter` runs at `Ordered.HIGHEST_PRECEDENCE`** so headers are written even on responses generated by Spring Security's auth filter (401/403). This is *intentional* and correct — protects error pages. Not a finding.
- **`Logback-spring.xml` patterns** include `%X{correlationId:-}` MDC reference. This is a server-generated UUID, not user-controlled. No `%X{...}` over any auth header. Not a finding.

## Summary

---

