# Trabuco Security Audit — Data + Events Domain

Persistence (Flyway, JDBC, HikariCP, NoSQL drivers) and messaging (Kafka, RabbitMQ, SQS, Pub/Sub) — schema validation, idempotency, deserialization, credential handling, TLS, and consumer hardening.

This file is the **detail reference** for the
`trabuco-security-audit-data-events` specialist subagent. The orchestrator
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

**Total checks in this domain: 44**
(2 Critical,
 11 High,
 18 Medium,
 5 Low,
 2 Informational)

---

## F-DATA-08 — Vector `documents` table has no `tenant_id` / `caller_id` / `authority_scope` column; `VectorKnowledgeRetriever.retrieve` builds `SearchRequest` without filter

**Severity floor:** Critical
**Taxonomy:** TRABUCO-004, OWASP-A01, API-API1, CWE-639

### Where to look

`aiagent-pgvector-rag/AIAgent/src/main/resources/db/vector-migration/V1__create_vector_schema.sql:35-40`, `aiagent-pgvector-rag/AIAgent/src/main/java/com/security/audit/aiagent/knowledge/VectorKnowledgeRetriever.java:41-47`

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

The DDL has no partition key tying a document to a tenant, caller, or authority scope; the retrieval helper builds `SearchRequest` with neither `filterExpression` nor any per-caller predicate. Any document ingested by partner A is retrieval-visible to a `/chat` from partner B. The auth.md F-AUTH-25 already noted ingestion's metadata blob is unsanitized — combined here, partner A can inject `metadata.tenant=B` and steer cross-tenant retrieval. There is also no integration test asserting cross-tenant queries return empty (per checklist's verification requirement). aiagent.md surfaces F-AIAGENT-03 from the prompt-injection angle; this entry anchors the same defect to the data-schema and retrieval-API layers per TRABUCO-004.

### Suggested fix

Add `tenant_id UUID NOT NULL` (or `authority_scope TEXT`) as a non-null column with an index; emit `FilterExpressionBuilder.eq("tenant_id", caller.tenantId())` in `VectorKnowledgeRetriever.retrieve`; ship a Testcontainers test asserting `partner-A` query returns 0 rows from `partner-B` ingest.

---

---

## F-EVENTS-02 — RabbitMQ listener consumes `PlaceholderEvent` with a default `Jackson2JsonMessageConverter` and **no** `DefaultJackson2JavaTypeMapper.setTrustedPackages(...)` configured — Spring-AMQP's default trusted-package list is `["*"]`, meaning any sender can specify any FQCN in the `__TypeId__` header

**Severity floor:** Critical

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/eventconsumer/src/main/java/com/security/audit/eventconsumer/config/RabbitConfig.java:41-44`, `/tmp/trabuco-secaudit/full-fat-all-modules/EventConsumer/src/main/java/com/security/audit/eventconsumer/config/RabbitConfig.java:41-44` (duplicate path; both ship), and same listener at `eventconsumer/listener/PlaceholderEventListener.java:34-37`

### Evidence pattern

```java
@Bean
public Jackson2JsonMessageConverter jsonMessageConverter(ObjectMapper objectMapper) {
  return new Jackson2JsonMessageConverter(objectMapper);
}
```

---

## F-DATA-01 — Flyway migration runs as the same Postgres user as the runtime app (super/owner privileges leak into runtime)

**Severity floor:** High
**Taxonomy:** TRABUCO-009, CWE-269, ASVS-V14-1

### Where to look

`*/SQLDatastore/src/main/resources/application.yml:5-27`, `aiagent-pgvector-rag/AIAgent/src/main/java/com/security/audit/aiagent/config/VectorFlywayConfig.java:43-54`

### Evidence pattern

```yaml
spring:
  datasource:
    username: ${DB_USERNAME:postgres}
    password: ${DB_PASSWORD:postgres}
  flyway:
    enabled: ${FLYWAY_ENABLED:true}
    locations: classpath:db/migration
    baseline-on-migrate: ${FLYWAY_BASELINE_ON_MIGRATE:false}
    validate-on-migrate: ${FLYWAY_VALIDATE:true}
```

### Why it matters

Trabuco never templates `spring.flyway.user` / `spring.flyway.password` distinct from `spring.datasource.*`. The runtime application therefore connects with the same role that performs `CREATE EXTENSION`, `CREATE SCHEMA`, `CREATE TABLE`, `CREATE INDEX`, and (with the dual-Flyway design) creates a brand-new schema. A SQL-injection or RCE that lands inside the JDBC session immediately inherits DDL/DROP power; principle-of-least-privilege is unenforceable.

### Suggested fix

Template a separate `spring.flyway.user` / `spring.flyway.password` (defaulted to the same role for local dev) and document creating a runtime DML-only role; in `VectorFlywayConfig`, accept an alternate datasource bean.

---

---

## F-DATA-02 — `postgres / postgres` shipped as default DB credentials in env, yml, docker-compose AND .env.example

**Severity floor:** High
**Taxonomy:** CWE-798, CWE-1188, ASVS-V14-1, OWASP-A05

### Where to look

`*/SQLDatastore/src/main/resources/application.yml:8-9`, `*/docker-compose.yml:14-17`, `*/.env.example:10-11`, `*/Worker/src/main/resources/application.yml:25-29`

### Evidence pattern

```yaml
# application.yml
username: ${DB_USERNAME:postgres}
password: ${DB_PASSWORD:postgres}
# docker-compose.yml
POSTGRES_USER: postgres
POSTGRES_PASSWORD: postgres
# .env.example
DB_USERNAME=postgres
DB_PASSWORD=postgres
```

### Why it matters

Three layered defaults all converge on the trivial `postgres/postgres` credential. The Worker module's `application.yml` hardcodes the literal `postgres/postgres` with no env-var indirection at all (line 26-28 — `username: postgres`, `password: postgres`). Any operator who skips `.env` overrides ships these to prod; the comment says "for local dev only" but `${DB_PASSWORD:postgres}` makes "no env override" silently equal to "default password". Detection by static scanners is not triggered because the yaml looks parameterized.

### Suggested fix

Replace defaults with `${DB_PASSWORD:?DB_PASSWORD must be set}` (Spring required-property syntax) so the app refuses to boot without an explicit value; remove hardcoded creds from Worker yaml.

---

---

## F-DATA-03 — Postgres `sslmode=prefer` default — silently downgrades to plaintext when server doesn't offer TLS

**Severity floor:** High
**Taxonomy:** ASVS-V9-1, OWASP-A02, CWE-319

### Where to look

`*/SQLDatastore/src/main/resources/application.yml:7`, `*/AIAgent/src/main/resources/application.yml:87`

### Evidence pattern

```yaml
url: ${DB_URL:jdbc:postgresql://${DB_HOST:localhost}:${DB_PORT:5433}/${DB_NAME:crud-sql-worker}?sslmode=${DB_SSL_MODE:prefer}}
```

### Why it matters

`sslmode=prefer` is documented by the PostgreSQL JDBC driver to attempt TLS but transparently fall back to plaintext if the server replies that TLS is unavailable. There is no certificate verification (`verify-full` / `verify-ca`). An MITM that strips the StartSSL exchange or a misconfigured PgBouncer in front of the prod cluster yields a silent plaintext session carrying credentials and row data. The `# SECURITY: Always use SSL in production` comment doesn't change the actual default value.

### Suggested fix

Default `DB_SSL_MODE=verify-full` (or `require` at minimum) and require an explicit `DB_SSL_ROOT_CERT` env var; document local-dev override as `sslmode=disable`.

---

---

## F-DATA-06 — Redis ships with no AUTH, no TLS, optional-password commented out

**Severity floor:** High
**Taxonomy:** CWE-798, CWE-319, ASVS-V9-1, OWASP-A05

### Where to look

`nosql-redis-no-ai/NoSQLDatastore/src/main/resources/application.yml:1-19`, `nosql-redis-no-ai/docker-compose.yml:11-22`

### Evidence pattern

```yaml
spring:
  data:
    redis:
      host: ${REDIS_HOST:localhost}
      port: ${REDIS_PORT:6379}
      # Optional password (uncomment if needed)
      # password: ${REDIS_PASSWORD:}
      timeout: 2000ms
```

### Why it matters

Generated docker-compose binds 6379 to the host with no authentication; the matching application config has the password line commented out. An operator copying this to staging exposes an unauthenticated Redis with FLUSHALL/CONFIG SET available — Redis is a documented RCE vector via `CONFIG SET dir` + `CONFIG SET dbfilename`. Spring's `application.yml` also has no `ssl: enabled: true` toggle, so even setting REDIS_PASSWORD travels in cleartext.

### Suggested fix

Uncomment `password: ${REDIS_PASSWORD:?REDIS_PASSWORD must be set}` by default; add `requirepass` and `--protected-mode yes` to the docker-compose service; template `spring.data.redis.ssl.enabled` and a TLS truststore.

---

---

## F-DATA-07 — Redis `RedisTemplate` uses `GenericJackson2JsonRedisSerializer` — ships polymorphic `@class` markers (Jackson-default-typing-equivalent gadget chain risk)

**Severity floor:** High
**Taxonomy:** ASVS-V5-3, OWASP-A08, CWE-502

### Where to look

`nosql-redis-no-ai/NoSQLDatastore/src/main/java/com/security/audit/nosqldatastore/config/NoSQLConfig.java:31-41`

### Evidence pattern

```java
@Bean
public RedisTemplate<String, Object> redisTemplate(RedisConnectionFactory connectionFactory, ObjectMapper redisObjectMapper) {
  RedisTemplate<String, Object> template = new RedisTemplate<>();
  ...
  template.setValueSerializer(new GenericJackson2JsonRedisSerializer(redisObjectMapper));
  template.setHashValueSerializer(new GenericJackson2JsonRedisSerializer(redisObjectMapper));
  template.afterPropertiesSet();
  return template;
}
```

### Why it matters

`GenericJackson2JsonRedisSerializer` writes a Jackson `@class` type marker into every value and reads it back via `ObjectMapper.activateDefaultTyping`-equivalent semantics. Combined with the unauthenticated Redis (F-DATA-06), an attacker who can write a key (FLUSHDB → SET) injects a serialized `@class` referencing any class on the classpath with a single-arg-string constructor or property setter — the same class of CVE-2017-15095 / CVE-2019-12086 gadgets that have plagued Jackson default-typing in the past. Even with auth, a co-tenant or compromised microservice on the Redis becomes an RCE vector.

### Suggested fix

Replace with `Jackson2JsonRedisSerializer<>(YourSpecificType.class)` and a per-cache typed bean, or use `StringRedisSerializer` + manual JSON via a context-typed mapper; explicitly disallow polymorphic deserialization.

---

---

## F-DATA-09 — `IngestRequest` (record) has no validation; `metadata: Map<String,Object>` accepted unbounded into vector store

**Severity floor:** High
**Taxonomy:** ASVS-V5-1, OWASP-A04, API-API4, CWE-20, CWE-770

### Where to look

`aiagent-pgvector-rag/AIAgent/src/main/java/com/security/audit/aiagent/protocol/IngestionController.java:63-95`, `aiagent-pgvector-rag/AIAgent/src/main/java/com/security/audit/aiagent/knowledge/DocumentIngestionService.java:50-69`

### Evidence pattern

```java
@PostMapping
@RequireScope("partner")
public ResponseEntity<Map<String, Object>> ingest(@RequestBody IngestRequest req) {
    ingestionService.ingest(req.text(), req.metadata());
    ...
}
public record IngestRequest(String text, Map<String, Object> metadata) {}
```

### Why it matters

The endpoint annotation lacks `@Valid`, the record fields lack `@NotBlank`, `@Size`, `@Pattern`. Any `text` (1 byte → 100 MB), any `metadata` keys/values pass through to `vectorStore.add(chunks)`. Combined with no `spring.servlet.multipart.max-file-size`, no `server.tomcat.max-http-form-post-size`, the endpoint accepts arbitrarily large bodies. The metadata `Map<String, Object>` becomes JSONB rows in the `documents` table without key allow-listing — a partner can inject `metadata.embedding_attack_prompt` or metadata so large it exceeds Postgres's TOAST page limit. aiagent.md F-AIAGENT-14 surfaces the same DTO issue from a prompt-injection angle; this entry anchors it to ASVS-V5-1 / mass-assignment / table-bloat from the data-DTO angle.

### Suggested fix

Add `@Valid` on the parameter; add `@NotBlank @Size(max=200_000)` on `text`, `@Size(max=20)` on the metadata Map, key-allow-listing or fixed-shape DTO for metadata; cap multipart and JSON body size in YAML.

---

---

## F-DATA-12 — Worker module hardcodes `username: postgres` / `password: postgres` (no env-var indirection) — `application.yml` overrides the SQLDatastore yaml's parameterized version

**Severity floor:** High
**Taxonomy:** CWE-798, ASVS-V14-1, OWASP-A05

### Where to look

`crud-sql-worker/Worker/src/main/resources/application.yml:25-29`, `full-fat-all-modules/Worker/src/main/resources/application.yml:25-29`

### Evidence pattern

```yaml
datasource:
  url: jdbc:postgresql://localhost:5433/crud-sql-worker
  username: postgres
  password: postgres
  driver-class-name: org.postgresql.Driver
```

### Why it matters

The Worker module's yaml is concatenated with the Shared/SQLDatastore yamls at runtime via Spring's profile resolution. Because Worker hardcodes literal values (no `${DB_USERNAME:postgres}` syntax), the operator's `DB_USERNAME` / `DB_PASSWORD` env vars set on the API process are silently ignored on the Worker process — operators discover this only when the Worker fails auth in prod, and the comment "override via SPRING_DATASOURCE_*" depends on the operator setting *those* specific env vars and not the `DB_*` ones documented elsewhere. Net effect: divergent credential sourcing between API and Worker, and Worker traffics literal `postgres/postgres` if the SPRING_DATASOURCE_* envs aren't set.

### Suggested fix

Use `${DB_URL:jdbc:postgresql://localhost:5433/...}`, `${DB_USERNAME:?required}`, `${DB_PASSWORD:?required}` consistently; OR delete the duplicated `datasource:` block and import the SQLDatastore yaml.

---

---

## F-EVENTS-04 — Inbound event records have **zero** `jakarta.validation` constraints; payload size, character class, and required fields are entirely unbounded

**Severity floor:** High

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/Model/src/main/java/com/security/audit/model/events/PlaceholderCreatedEvent.java:18-23`, `/tmp/trabuco-secaudit/events-kafka-consumer/Model/src/main/java/com/security/audit/model/events/PlaceholderCreatedEvent.java:18-23`, `/tmp/trabuco-secaudit/full-fat-all-modules/Model/src/main/java/com/security/audit/model/jobs/ProcessPlaceholderJobRequest.java:31`

### Evidence pattern

```java
public record PlaceholderCreatedEvent(
  @JsonProperty("eventId") String eventId,
  @JsonProperty("occurredAt") Instant occurredAt,
  @JsonProperty("placeholderId") String placeholderId,
  @JsonProperty("name") String name
) implements PlaceholderEvent { ... }
```

---

## F-EVENTS-05 — No idempotency / dedup at the consumer; the inventory template promises "Consumers should use this to detect duplicate events" but the listener does not — every retry, every replay, every at-least-once redelivery re-runs business logic

**Severity floor:** High

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/eventconsumer/src/main/java/com/security/audit/eventconsumer/listener/PlaceholderEventListener.java:46-64`, `/tmp/trabuco-secaudit/full-fat-all-modules/eventconsumer/src/main/java/com/security/audit/eventconsumer/listener/PlaceholderEventListener.java:34-43`, `Model/.../PlaceholderEvent.java:42-45` (the eventId contract)

---

## F-EVENTS-09 — Kafka broker config: `bootstrap-servers` defaults to `localhost:9092` with **PLAINTEXT** protocol; docker-compose advertises `PLAINTEXT://localhost:9092` and enables `KAFKA_AUTO_CREATE_TOPICS_ENABLE: true`. There is no SASL/SSL, no per-consumer-group ACL, and no schema-registry stamping

**Severity floor:** High

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/eventconsumer/src/main/resources/application.yml:16-21`, `/tmp/trabuco-secaudit/events-kafka-consumer/docker-compose.yml:43-58`

### Evidence pattern

```yaml
# application.yml
kafka:
  bootstrap-servers: ${KAFKA_BOOTSTRAP_SERVERS:localhost:9092}
  consumer:
    group-id: ${KAFKA_CONSUMER_GROUP:events-kafka-consumer-consumers}
```

---

## F-EVENTS-10 — RabbitMQ ships `guest:guest` as the AMQP credential default with no SSL — duplicate of F-INFRA / F-DATA findings but specific to AMQP listener: `guest` is locked to localhost-only by RabbitMQ, so production deploys that bind to non-localhost silently fail-open with no creds

**Severity floor:** High

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/eventconsumer/src/main/resources/application.yml:16-21`, `/tmp/trabuco-secaudit/full-fat-all-modules/.env.example:27-32`, `docker-compose.yml:30-38`

### Evidence pattern

```yaml
rabbitmq:
  host: ${RABBITMQ_HOST:localhost}
  port: ${RABBITMQ_PORT:5672}
  username: ${RABBITMQ_USERNAME:guest}
  password: ${RABBITMQ_PASSWORD:guest}
```

---

## F-DATA-04 — HikariCP `leak-detection-threshold: 0` — connection leaks never surface

**Severity floor:** Medium
**Taxonomy:** ASVS-V11-1, OWASP-A04, CWE-770

### Where to look

`*/SQLDatastore/src/main/resources/application.yml:20`

### Evidence pattern

```yaml
hikari:
  ...
  leak-detection-threshold: ${DB_LEAK_DETECTION:0}
```

### Why it matters

`0` disables Hikari's leak detector. A repository or service that fails to close its connection (`try-with-resources` missed, exception path not draining the cursor, JdbcTemplate lifecycle anomaly) silently exhausts the pool until `connection-timeout` (20 s default). Combined with `maximum-pool-size: 10` and no rate limiting (per F-AUTH-13), an operator only learns of the problem at the outage. Hikari recommends 2000–5000 ms in production.

### Suggested fix

Default `DB_LEAK_DETECTION=2000` (2 s) and document tuning above 5 s for known long transactions.

---

---

## F-DATA-05 — AIAgent HikariCP `maximum-pool-size: 5` shared with Spring AI's PgVectorStore — vector-search saturation = denial of all DB access

**Severity floor:** Medium
**Taxonomy:** OWASP-A04, ASVS-V11-1, CWE-770

### Where to look

`aiagent-pgvector-rag/AIAgent/src/main/resources/application.yml:91-97`

### Evidence pattern

```yaml
hikari:
  pool-name: AiagentPgvectorRagAgentPool
  maximum-pool-size: ${DB_POOL_SIZE:5}
  minimum-idle: ${DB_POOL_MIN_IDLE:2}
```

### Why it matters

The AIAgent process shares a single `DataSource` between the application's JDBC repositories, the primary `Flyway` bean, the secondary `flywayVector` bean (F-DATA-01), and Spring AI's `PgVectorStore` for similarity search. With a 5-connection pool and `/ingest`, `/ingest/batch`, `/ask`, `/chat` all colliding on it, an unauthenticated burst of similarity searches (per F-AUTH-08, ingestion is reachable without true auth) starves every other DB-backed feature including the actuator's DB health check. There is no isolation between agent traffic and domain traffic.

### Suggested fix

Either raise the default (10–20) and add a Bucket4j-style rate limit per caller, or template a second HikariCP pool just for the vector store.

---

---

## F-DATA-10 — Spring Data JDBC `@Query` `searchByName` ILIKE accepts wildcards in `:search` — LIKE-pattern injection causes full-scan availability hit

**Severity floor:** Medium
**Taxonomy:** OWASP-A04, API-API4, CWE-770

### Where to look

`*/SQLDatastore/src/main/java/com/security/audit/sqldatastore/repository/PlaceholderRepository.java:38-40`

### Evidence pattern

```java
@Query("SELECT * FROM placeholders WHERE name ILIKE '%' || :search || '%'")
List<PlaceholderRecord> searchByName(@Param("search") String search);
```

### Why it matters

While parameter binding is safe vs. classic SQL injection, the query never escapes `%` and `_` inside `:search`. A caller passing `%` matches every row; `___` matches every 3-letter row. Combined with the `findAll`-style `getAll()` controller path returning the whole result set, this is a built-in DoS amplifier. Even with the index on `name`, ILIKE with leading-`%` cannot use a btree, so every search forces a sequential scan; an attacker iterating thousands of `%`-prefixed searches saturates the DB.

### Suggested fix

Either escape `%` and `_` in the bound parameter (`replace(:search,'%','\%')`) or accept a constrained `:prefix` parameter and use `name ILIKE :prefix || '%'` with a `text_pattern_ops` index; document the trade-off in the placeholder comment.

---

---

## F-DATA-11 — `PlaceholderRecord`/`PlaceholderDocument` are bound directly via Spring Data — entity serves as both DTO and persistence row for tests / events; mass-assignment via `@RequestBody` shape change waits one refactor away

**Severity floor:** Medium
**Taxonomy:** API-API3, OWASP-A04, CWE-915

### Where to look

`*/Model/src/main/java/com/security/audit/model/entities/PlaceholderRecord.java:17-44`, `nosql-redis-no-ai/Model/src/main/java/com/security/audit/model/entities/PlaceholderDocument.java:18-36`

### Evidence pattern

```java
@Table("placeholders")
public record PlaceholderRecord(
  @Id @Nullable Long id,
  String name,
  @Nullable String description,
  @Nullable Instant createdAt,
  @Nullable Instant updatedAt
) { ... }
```

### Why it matters

The current `PlaceholderController` correctly accepts `ImmutablePlaceholderRequest` (DTO) and re-builds responses, so today's wiring is safe. But the templates reuse the record across DTO-shaped paths (events, jobs, test fixtures) and the codebase comment "Replace this with your actual database record classes" explicitly invites the developer to extend it. Once a future endpoint binds `@RequestBody PlaceholderRecord`, mass-assignment reaches `id`, `createdAt`, `updatedAt` — fields that are server-controlled today. There is no `@JsonIgnore` / `@JsonProperty(access=READ_ONLY)` on those columns to make the safer path the default.

### Suggested fix

Annotate `id`, `createdAt`, `updatedAt` with `@JsonProperty(access = READ_ONLY)` and add a `code-reviewer` skill rule that flags `@RequestBody *Record` / `@RequestBody *Document` patterns.

---

---

## F-DATA-13 — Flyway `validate-on-migrate` defaulted true but `clean-disabled` not set — `flyway:clean` reachable in any environment

**Severity floor:** Medium
**Taxonomy:** OWASP-A05, ASVS-V14-1, CWE-269

### Where to look

`*/SQLDatastore/src/main/resources/application.yml:22-27`

### Evidence pattern

```yaml
flyway:
  enabled: ${FLYWAY_ENABLED:true}
  locations: classpath:db/migration
  baseline-on-migrate: ${FLYWAY_BASELINE_ON_MIGRATE:false}
  validate-on-migrate: ${FLYWAY_VALIDATE:true}
```

### Why it matters

`spring.flyway.clean-disabled` is missing. From Flyway 9 onwards the default flipped to `true`, but Spring Boot 3.x exposes `clean-disabled` through `FlywayProperties` and a misconfigured CI/CD or operator's `mvn flyway:clean` can drop every table the runtime DB role can `DROP` — combined with F-DATA-01 (runtime user ≡ Flyway user with DROP), one stray Maven goal wipes prod. A defense-in-depth template would set `clean-disabled: true` explicitly.

### Suggested fix

Add `clean-disabled: ${FLYWAY_CLEAN_DISABLED:true}` and document opt-out only for ephemeral test environments.

---

---

## F-DATA-14 — `VectorFlywayConfig` runs `createSchemas(true)` + `baselineOnMigrate(true)` against a shared DataSource — silent skipped-baseline on partial state

**Severity floor:** Medium
**Taxonomy:** TRABUCO-009, OWASP-A04, ASVS-V14-1

### Where to look

`aiagent-pgvector-rag/AIAgent/src/main/java/com/security/audit/aiagent/config/VectorFlywayConfig.java:43-54`

### Evidence pattern

```java
return Flyway.configure()
    .dataSource(dataSource)
    .schemas("vector")
    .defaultSchema("vector")
    .table("flyway_schema_history_vector")
    .locations("classpath:db/vector-migration")
    .baselineOnMigrate(true)
    .createSchemas(true)
    .load();
```

### Why it matters

`baselineOnMigrate(true)` is deliberately enabled to handle the "schema exists but no Flyway history" case. In production, this means: if an operator manually creates the `vector` schema between deploys (e.g., to grant permissions), the next deploy silently baselines whatever state is present and never reapplies the V1 migration — leaving rows ingested under a missing index, or with the wrong vector dimension. Combined with the `IF NOT EXISTS` guard on `documents`, schema drift is absorbed silently rather than failing fast.

### Suggested fix

Default `baselineOnMigrate(false)` and require explicit operator opt-in via env var; explicitly fail fast if `vector.documents` exists with mismatched column types.

---

---

## F-DATA-15 — pgvector schema migration uses `CREATE EXTENSION` — fails closed only if runtime user is non-superuser, but Trabuco generated user IS superuser (per F-DATA-01)

**Severity floor:** Medium
**Taxonomy:** CWE-269, TRABUCO-009

### Where to look

`aiagent-pgvector-rag/AIAgent/src/main/resources/db/vector-migration/V1__create_vector_schema.sql:21`

### Evidence pattern

```sql
CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public;
```

### Why it matters

`CREATE EXTENSION` requires superuser (or a member of `pg_database_owner` in PG 13+ for trusted extensions). pgvector is *not* a trusted extension on most managed Postgres products (RDS, Cloud SQL, Heroku) — meaning the migration only succeeds when the runtime DB user is superuser. F-DATA-01 already noted that Trabuco uses one user for both Flyway and runtime, so the runtime DB session has superuser rights too: `pg_read_server_files`, `COPY ... PROGRAM`, `CREATE LANGUAGE`. A SQL-injection elsewhere becomes RCE-on-DB-host with no further escalation needed.

### Suggested fix

Document operator pre-creation of the `vector` extension by a DBA, then have the migration assume it exists (`SELECT 1 FROM pg_extension WHERE extname='vector'` and abort migration if missing). The runtime user can then be DML-only.

---

---

## F-DATA-16 — Qdrant API key default `QDRANT_API_KEY:` empty, `QDRANT_USE_TLS:false`, port 6334 over plaintext gRPC

**Severity floor:** Medium
**Taxonomy:** ASVS-V9-1, CWE-319, ASVS-V2-2, OWASP-A05

### Where to look

`full-fat-all-modules/AIAgent/src/main/resources/application.yml:74-80`

### Evidence pattern

```yaml
vectorstore:
  qdrant:
    host: ${QDRANT_HOST:localhost}
    port: ${QDRANT_PORT:6334}
    api-key: ${QDRANT_API_KEY:}
    collection-name: ${QDRANT_COLLECTION:documents}
    use-tls: ${QDRANT_USE_TLS:false}
    initialize-schema: true
```

### Why it matters

Defaults set the API key to empty (Qdrant accepts unauthenticated requests when its own `service.api_key` is also empty — which the local-dev image is) and TLS off. `initialize-schema: true` means the agent process auto-creates the collection at startup with whatever `dimensions` the embedder picks, eliding any operator-mediated provisioning step. There is no equivalent of the `vector` schema isolation that pgvector ships with — the entire collection is shared across all callers (same as F-DATA-08, but no JSONB metadata column to even add filters to without payload-design changes).

### Suggested fix

Default `QDRANT_API_KEY` to a required env (`${... :?QDRANT_API_KEY required}`), default `QDRANT_USE_TLS=true`, and `initialize-schema=false` with a one-shot init script that creates the collection with a `tenant` payload field for filtering.

---

---

## F-DATA-17 — Spring Data Redis `@RedisHash("placeholder")` uses a flat hard-coded key namespace — no prefix per-tenant or per-environment, key collisions across deployments sharing a Redis

**Severity floor:** Medium
**Taxonomy:** OWASP-A04, TRABUCO-004, CWE-639

### Where to look

`nosql-redis-no-ai/Model/src/main/java/com/security/audit/model/entities/PlaceholderDocument.java:18-31`

### Evidence pattern

```java
@RedisHash("placeholder")
public record PlaceholderDocument(
  @Id String id,
  @Indexed String name,
  ...
) { ... }
```

### Why it matters

Spring Data Redis stores under `placeholder:<id>` with index `placeholder:name:<value>`. The keyspace has no per-tenant or per-app prefix. Two services pointed at the same Redis (e.g., a multi-stage cluster reusing one Redis to save infra cost — F-DATA-06 makes that more attractive because the Redis is unauthenticated) collide silently: service A's id `42` overwrites service B's id `42`. Combined with no `tenant_id` field on the record, BOLA via id-collision is one mis-config away.

### Suggested fix

Set `@RedisHash(value = "${spring.application.name}:placeholder", timeToLive = 3600)`; add a `tenantId` `@Indexed` field and require it on every read/write.

---

---

## F-DATA-18 — `findAll()` returned by repository iterators with no caller / tenant filter (NoSQL Redis path) — entire collection enumerable on a single GET

**Severity floor:** Medium
**Taxonomy:** API-API4, OWASP-A04, CWE-770

### Where to look

`nosql-redis-no-ai/Shared/src/main/java/com/security/audit/shared/service/PlaceholderService.java:80-92`

### Evidence pattern

```java
@CircuitBreaker(name = "default")
public List<ImmutablePlaceholder> findAll() {
  // trabuco-allow: perf.unbounded-scan (placeholder demo — replace with cursor or keyset drain)
  return StreamSupport.stream(repository.findAll().spliterator(), false)
    ...
}
```

### Why it matters

Redis `findAll` does a `SCAN` across the entire `placeholder:*` keyspace and then a `MGET`-equivalent. With no caller filter, no rate limit (per F-AUTH-13), and the controller exposing it on `GET /api/placeholders`, a single request enumerates the full dataset of every tenant. The `trabuco-allow` suppression normalizes the antipattern in the template.

### Suggested fix

Replace `findAll()` with `findByTenantIdAndIdGreaterThanOrderById(tenant, afterId, PageRequest.of(0, 50))` and remove the suppression; cap the `topK` server-side.

---

---

## F-DATA-19 — No request body / multipart size cap globally — `IngestionController` and any future `MultipartFile` uploader are unbounded

**Severity floor:** Medium
**Taxonomy:** ASVS-V12-1, ASVS-V11-1, API-API4, CWE-770, CWE-434

### Where to look

all `*/API/src/main/resources/application.yml`, all `*/AIAgent/src/main/resources/application.yml`

### Why it matters

Spring Boot's defaults are 1 MB / 10 MB (multipart) but `server.max-http-request-header-size` is 8 KB (fine) and JSON bodies are governed by `server.tomcat.max-swallow-size=2 MB` only after parsing has begun. The IngestionController accepts `@RequestBody List<IngestRequest>` (F-DATA-09) — that JSON path bypasses multipart limits entirely and is bounded only by Tomcat's `maxPostSize` (default 2 MB but flipped to no-limit for non-form bodies). A 200 MB JSON `text` payload is parsed into heap, embedded, chunked, and stored. No `MultipartFile` is currently used, but the moment one is added (e.g., file-based ingestion noted in `IngestionController` doc comment line 73-75), the absence of templates makes "no limit" the silent default.

### Suggested fix

Template `spring.servlet.multipart.max-file-size: 10MB`, `spring.servlet.multipart.max-request-size: 50MB`, and a custom `WebServerFactoryCustomizer` capping JSON body size; add a per-record `@Size(max = 100_000)` on `IngestRequest.text`.

---

---

## F-EVENTS-03 — Sealed-interface `switch` in listeners is non-exhaustive (no `default` branch) — a deserialized event for an unmapped subtype silently falls through and is `ack`-ed; combined with `JsonTypeInfo` lookup, this is a poison-pill that drops messages unobserved

**Severity floor:** Medium

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/eventconsumer/src/main/java/com/security/audit/eventconsumer/listener/PlaceholderEventListener.java:60-64`, `/tmp/trabuco-secaudit/full-fat-all-modules/eventconsumer/src/main/java/com/security/audit/eventconsumer/listener/PlaceholderEventListener.java:39-43`

### Evidence pattern

```java
switch (event) {
  case PlaceholderCreatedEvent created -> handleCreated(created);
  // Add more cases as event types grow
}
```

---

## F-EVENTS-06 — DLT/DLQ end-state is "log + TODO": no alerting, no replay tooling, no quota — a poison-pill silently fills the DLQ until the broker hits its disk limit

**Severity floor:** Medium

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/eventconsumer/src/main/java/com/security/audit/eventconsumer/listener/PlaceholderEventListener.java:73-78`, RabbitMQ DLQ at `/tmp/trabuco-secaudit/full-fat-all-modules/eventconsumer/src/main/java/com/security/audit/eventconsumer/config/RabbitConfig.java:80-87` (declared but no consumer)

### Evidence pattern

```java
@DltHandler
public void handleDlt(PlaceholderEvent event, @Header(KafkaHeaders.RECEIVED_TOPIC) String topic) {
  logger.error("Event sent to DLT: ...");
  // TODO: Add alerting, store for manual review, etc.
}
```

---

## F-EVENTS-07 — Kafka producer in `Events/EventPublisher.java` does not block / await ack, has no idempotent-producer setting, no transactional-id, and no per-record retry policy

**Severity floor:** Medium

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/Events/src/main/java/com/security/audit/events/EventPublisher.java:51-55`, `application.yml:22-25` (producer config)

### Evidence pattern

```java
public void publish(PlaceholderEvent event) {
  logger.info("Publishing event to Kafka: ...");
  kafkaTemplate.send(placeholderTopic, event.eventId(), event);   // returns CompletableFuture, never awaited
}
```

---

## F-EVENTS-08 — Rabbit producer publishes via fanout exchange with **empty routing key**, **no `MessageDeliveryMode.PERSISTENT`** explicitly set, no publisher-confirms enabled, and no return-on-unroutable callback — silent message drops are possible on broker restart or queue absence

**Severity floor:** Medium

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/events/src/main/java/com/security/audit/events/EventPublisher.java:51-55`, `/tmp/trabuco-secaudit/full-fat-all-modules/Events/src/main/java/com/security/audit/events/config/RabbitConfig.java:26-33`

### Evidence pattern

```java
rabbitTemplate.convertAndSend(placeholderExchange, "", event);  // routing key ""
```

---

## F-EVENTS-11 — Listener concurrency hard-coded high (Kafka `setConcurrency(3)`, Rabbit `setMaxConcurrentConsumers(10)`) but with **no `setMaxPollRecords`, no `setQueueSize`, no `setPrefetchCount` cap** — single misbehaving handler can fan out arbitrary in-flight DB connections

**Severity floor:** Medium

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/eventconsumer/src/main/java/com/security/audit/eventconsumer/config/KafkaConfig.java:54-62`, `/tmp/trabuco-secaudit/full-fat-all-modules/eventconsumer/src/main/java/com/security/audit/eventconsumer/config/RabbitConfig.java:48-59`

### Evidence pattern

```java
factory.setConcurrency(3);            // KafkaConfig
```

---

## F-EVENTS-13 — `JobRunr` `default-number-of-retries: 10` with exponential backoff; no max-retry cap on individual handler types; failing-fast handlers cannot opt out

**Severity floor:** Medium

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/worker/src/main/resources/application.yml:85-87`, `/tmp/trabuco-secaudit/crud-sql-worker/worker/src/main/resources/application.yml:85-87`

### Evidence pattern

```yaml
jobs:
  default-number-of-retries: 10
```

---

## F-EVENTS-16 — Schema versioning is absent: the `@type` discriminator is the **simple class name** (not FQCN, not a versioned name); a refactor that renames `PlaceholderCreatedEvent` will silently drop in-flight events

**Severity floor:** Medium

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/Model/src/main/java/com/security/audit/model/events/PlaceholderEvent.java:30-37`

### Evidence pattern

```java
@JsonSubTypes({
  @JsonSubTypes.Type(value = PlaceholderCreatedEvent.class, name = "PlaceholderCreatedEvent")
})
```

---

## F-DATA-20 — `documents.metadata` stored as freeform JSONB with no key allowlist or size cap — JSONB bloat / TOAST exhaustion path

**Severity floor:** Low
**Taxonomy:** OWASP-A04, ASVS-V11-1, CWE-770

### Where to look

`aiagent-pgvector-rag/AIAgent/src/main/resources/db/vector-migration/V1__create_vector_schema.sql:36-40`, `aiagent-pgvector-rag/AIAgent/src/main/java/com/security/audit/aiagent/knowledge/DocumentIngestionService.java:50-57`

### Evidence pattern

```sql
metadata  JSONB,
```

### Why it matters

PostgreSQL JSONB has a 1 GB hard limit per value but the GIN index on `metadata` (line 56-57 of the migration) makes write amplification severe — each ingestion with attacker-chosen keys causes index bloat that VACUUM must reclaim. The lack of a key allowlist means an attacker who passes ingestion (per F-AUTH-08) can submit metadata with thousands of unique keys, growing the GIN index without bound.

### Suggested fix

Constrain metadata at the application layer (allow-list e.g. `source`, `version`, `tenant_id`, `ingested_at`); add `CHECK (jsonb_object_keys(metadata) ...)` constraint or pre-validation.

---

---

## F-DATA-21 — `Spring Data JDBC` repository uses `Modifying @Query` with `FOR UPDATE SKIP LOCKED` but no caller-scope predicate — bulk update can be triggered cross-tenant once auth ships

**Severity floor:** Low
**Taxonomy:** API-API1, OWASP-A01, CWE-639

### Where to look

`*/SQLDatastore/src/main/java/com/security/audit/sqldatastore/repository/PlaceholderRepository.java:65-79`

### Evidence pattern

```java
@Modifying
@Query("""
    UPDATE placeholders SET description = :neu
    WHERE id IN (
      SELECT id FROM placeholders
      WHERE description = :old
      ORDER BY id
      LIMIT :limit
      FOR UPDATE SKIP LOCKED
    )
    """)
int updateDescriptionBatchWithLimit(...);
```

### Why it matters

No `tenant_id` predicate; if the placeholder ever evolves into a multi-tenant table (which the Trabuco doc trail in JAVA_CODE_QUALITY.md encourages), this query updates rows across all tenants based on a globally-matched description string. The pattern propagates as the recommended bulk-DML idiom in `add-repository-method` skills, so future repositories inherit the omission.

### Suggested fix

In the placeholder template, add `AND tenant_id = :tenantId`; document tenant predicate as a required clause in the skill body.

---

---

## F-DATA-22 — Spring Boot OAuth2 resource-server `oauth2 client_id`-style exposure: actuator `/actuator/prometheus`, `/actuator/metrics`, `/actuator/health` all reach the database health probe — DB credentials test path leaks runtime user

**Severity floor:** Low
**Taxonomy:** OWASP-A05, ASVS-V8-1

### Where to look

`*/API/src/main/resources/application.yml:144-174`

### Evidence pattern

```yaml
management:
  endpoints:
    web:
      exposure:
        include: ${MANAGEMENT_ENDPOINTS:health,info,prometheus,metrics}
  endpoint:
    health:
      show-details: ${MANAGEMENT_HEALTH_DETAILS:when_authorized}
  health:
    db:
      enabled: true
```

### Why it matters

With `permitAllFilterChain` active by default (per auth.md F-AUTH-01), `/actuator/health` is reachable anonymously. With `MANAGEMENT_HEALTH_DETAILS` defaulted to `when_authorized`, the unauthenticated path returns generic UP/DOWN — but `/actuator/prometheus` and `/actuator/metrics` are exposed and unauthenticated. The Hikari pool name (`CrudSqlWorkerPool`) leaks via Prometheus tags, indirectly disclosing the project name + DB user via metric labels like `hikaricp_connections_active{pool="CrudSqlWorkerPool"}`. Combined with F-DATA-01, a fingerprinter knows the username and pool size.

### Suggested fix

Default `MANAGEMENT_ENDPOINTS=health,info` in API yaml; route `/actuator/prometheus` to a secondary management port (already done for Worker on `:8082`) and require an internal scope.

---

---

## F-EVENTS-18 — Kafka listener relies on `KAFKA_AUTO_CREATE_TOPICS_ENABLE=true` (compose default) yet `@RetryableTopic(autoCreateTopics = "false")` — startup race: the DLT topic does not exist on first boot and the first failure throws, masking the real error

**Severity floor:** Low

### Where to look

`PlaceholderEventListener.java:50` (`autoCreateTopics = "false"`), `docker-compose.yml:54` (`KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"`)

---

## F-EVENTS-19 — `events-kafka-consumer` has Spring Kafka producer config in EventConsumer's `application.yml:22-25` even though the EventConsumer module does not publish events — a dev who later wires a publish call will use the default `JsonSerializer` with no `addTypeInfo=false`, leaking class FQCN to Kafka headers

**Severity floor:** Low

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/eventconsumer/src/main/resources/application.yml:22-25`

### Evidence pattern

```yaml
producer:
  key-serializer: org.apache.kafka.common.serialization.StringSerializer
  value-serializer: org.springframework.kafka.support.serializer.JsonSerializer
```

---

## F-DATA-23 — `application-postgres.yml`-style profile variants not emitted; Postgres-specific defaults baked into the SQLDatastore base yaml — switching `Database=mysql` is silently broken

**Severity floor:** Informational
**Taxonomy:** OWASP-A05, ASVS-V14-1, TRABUCO-009

### Where to look

`*/SQLDatastore/src/main/resources/application.yml:7-10`

### Why it matters

Inventory says variants `application-postgres.yml` / `application-mysql.yml` are gated by `Database=="..."`, but the *SQLDatastore* yaml hardcodes Postgres-specific keys without a profile guard. If the operator selects MySQL, the `sslmode=prefer` URL parameter (Postgres-only) silently breaks JDBC parsing, and `org.postgresql.Driver` won't load. Migration scripts (`BIGSERIAL`) are also Postgres-specific. This is more correctness than security but a misconfigured driver fallback is a known DoS path.

### Suggested fix

Move driver class name and SSL mode into `application-postgres.yml` / `application-mysql.yml` profile; make the SQLDatastore yaml database-agnostic.

---

---

## F-DATA-24 — `flyway_schema_history` and `flyway_schema_history_vector` tables readable by runtime user — schema-version disclosure via app SQL injection or read-side leak

**Severity floor:** Informational
**Taxonomy:** OWASP-A05, ASVS-V8-1, CWE-200

### Where to look

`aiagent-pgvector-rag/AIAgent/src/main/java/com/security/audit/aiagent/config/VectorFlywayConfig.java:49`, `*/SQLDatastore/src/main/resources/application.yml`

### Why it matters

Schema version + checksum disclosure speeds CVE-mapping for an attacker who lands a SELECT primitive. With per-version migration descriptions ("create_vector_schema", "add_users_table"), the attacker maps the data model without reading the actual application source. Mitigations are typically `REVOKE SELECT ON flyway_schema_history FROM app_runtime` after migrations complete — Trabuco never templates that.

### Suggested fix

Document a post-migration `REVOKE` step and ship a Flyway callback that runs it; or move the history table to a separate, runtime-unreadable schema.

---

## False positives considered

- **`@Query` `searchByName` SQL injection (F-DATA-10 angle):** Spring Data JDBC parameter binding via `:search` is genuinely safe vs. SQL injection. Filed only as LIKE-pattern injection / DoS, not Critical SQLi.
- **`PlaceholderRepository.findAllByIdIn(Collection<Long>)` IN-clause expansion:** Spring Data JDBC expands `(:ids)` via `JdbcTemplate.NamedParameterJdbcOperations` parameter binding — each value is bound, no concatenation. Safe.
- **`PlaceholderRecord` mass assignment today:** Currently mediated by the controller's separate `ImmutablePlaceholderRequest`. Filed only as a future-refactor risk (F-DATA-11) because the templates invite developers down the antipattern.
- **Spring Data Redis `@Indexed` field secondary-index injection:** The `name` field is indexed via Redis `SADD placeholder:name:<value>` — but Spring Data Redis serializes the value through `StringRedisSerializer` after Spring's bean-property binding. No key-injection vector via repository derivation `findByName(String)`.
- **`PgVectorStore` SQL-injection via `metadata.filterExpression`:** Spring AI builds parameterized filters via `FilterExpressionBuilder` — strings are bound as JDBC parameters. Filed only as the absence-of-filter problem (F-DATA-08), not SQLi.
- **`GenericJackson2JsonRedisSerializer` polymorphism (F-DATA-07) — caveat:** Recent Spring Data Redis (3.x) uses an internal allow-list (`AllowAllNullClassLoader` removed; explicit type-validator). The risk is reduced from "any class" to "any class on the classpath that the validator allows." Still material because the allow-list is permissive and not under the operator's control.
- **`@Testcontainers(disabledWithoutDocker=true)` — TRABUCO-010 `Testcontainers Resource Bleed`:** Generated tests do not mount host paths, do not share state across PRs, do not pin reusable containers with secrets. Negative finding.
- **Mongo BSON injection / Mongo NoSQL key injection (CWE-89 NoSQL flavour):** No MongoDB target generated in the audit set (`NoSQLDatabase=mongodb` not selected); no `MongoTemplate` / `Criteria.where` / `new Query` in any reviewed archetype. Out of scope for this run.
- **`MultipartFile` directory-traversal (CWE-22, ASVS-V12-1):** Searched all five archetypes; no `MultipartFile` import or `Files.copy` / `Paths.get` user-supplied filename usage. Filed only as the prerequisite "no body limits" prep step (F-DATA-19) for when a developer adds it.
- **`@JsonAutoDetect` mass assignment via Jackson visibility:** No `@JsonAutoDetect` annotation present anywhere. Templates use Immutables-generated `@JsonDeserialize(as = ImmutableX.class)` which is constructor-bound and excludes fields not in the interface.

## Summary

24 data/persistence findings filed against Trabuco-CLI v1.11.0 across SQLDatastore, NoSQLDatastore, AIAgent vector schema, Flyway, HikariCP, Redis, and DTO surfaces.

| Severity | Count |
|---|---|
| Critical | 1 |
| High | 7 |
| Medium | 11 |
| Low | 3 |
| Informational | 2 |

---

## F-EVENTS-01 — Kafka `JsonDeserializer.TRUSTED_PACKAGES` set to a directory that already imports cross-module DTOs; combined with `@JsonTypeInfo(use=NAME) + @JsonSubTypes` on a sealed interface this is the documented Spring-Kafka deserialization-RCE pattern

**Severity floor:** High (Critical if a future event subtype references a non-record type with a Jackson-reachable setter/factory accepting `Object`)

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/eventconsumer/src/main/java/com/security/audit/eventconsumer/config/KafkaConfig.java:38-39`, plus `/tmp/trabuco-secaudit/events-kafka-consumer/Model/src/main/java/com/security/audit/model/events/PlaceholderEvent.java:30-37`

### Evidence pattern

```java
// KafkaConfig.java
props.put(JsonDeserializer.TRUSTED_PACKAGES, "com.security.audit.model.events");
props.put(JsonDeserializer.USE_TYPE_INFO_HEADERS, true);
```

---

## F-EVENTS-12 — RecurringJobsConfig allows free-form CRON strings with no validation; combined with the `JobScheduler` bean being available to any module, a future controller that accepts cron from the HTTP body has no schedule-injection guard

**Severity floor:** Medium (Latent — current generated code uses static `Cron.daily(...)` examples; the risk arises when developers add a config-driven schedule)

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/worker/src/main/java/com/security/audit/worker/config/RecurringJobsConfig.java:18-58`, `/tmp/trabuco-secaudit/crud-sql-worker/worker/src/main/java/com/security/audit/worker/config/RecurringJobsConfig.java:18-58`

---

## F-EVENTS-14 — `ProcessPlaceholderJobRequest` is JSON-serialized via JobRunr's default Jackson config and stored in `jobrunr_jobs.scheduledAt`/`jobAsJson` columns — a malicious enqueue (when an unauthenticated path exists, see web.md) can persist arbitrary JSON, and JobRunr deserializes it via `runtime.getJobActivator()` reflective dispatch

**Severity floor:** Medium (depends on F-AUTH-16 / web.md gating but it is the path)

### Where to look

`/tmp/trabuco-secaudit/full-fat-all-modules/Model/src/main/java/com/security/audit/model/jobs/ProcessPlaceholderJobRequestHandler.java:17-28`, `/tmp/trabuco-secaudit/full-fat-all-modules/Jobs/src/main/java/com/security/audit/jobs/PlaceholderJobService.java:31-55`

---

## F-EVENTS-15 — Producer adds **no message signing** or HMAC over event payload — a peer service that compromises the broker can rewrite events in flight; consumer has no integrity check

**Severity floor:** Medium (Defense-in-depth; broker is the trust boundary, so the practical risk depends on broker hygiene)

### Where to look

All `EventPublisher.java` and `PlaceholderEventListener.java` files

---

## F-EVENTS-17 — Event payloads logged at INFO with full `eventId` and `name` — sensitive fields would leak; `LOG_LEVEL=DEBUG` default in EventConsumer yaml exposes much more

**Severity floor:** Low (Today's payload is benign; the risk is structural — once developers add payload fields like `email`, `userId`, `apiKey`, they will log them)

### Where to look

`/tmp/trabuco-secaudit/events-kafka-consumer/eventconsumer/src/main/resources/application.yml:80` (`com.security.audit: ${LOG_LEVEL:DEBUG}`), `PlaceholderEventListener.java:57-58, 88-89` (logs `event.toString()`-style data)

---

## F-EVENTS-20 — `EventPublisher` in `Events/` is a `@Service` with no transactional outbox pattern; if the publisher commits the DB transaction first then the broker write fails, business state and event stream diverge — security-relevant for audit events

**Severity floor:** Low (correctness/audit issue; flagged because event-driven audit logs are a security control)

### Where to look

`EventPublisher.java` for both Kafka & Rabbit; `EventController.java:53-69` (publishes after-the-fact)

---

