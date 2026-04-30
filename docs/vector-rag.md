# Vector RAG (Retrieval-Augmented Generation)

Trabuco can scaffold a complete vector-search retrieval layer for the
**AIAgent** module. Pass `--vector-store=<flavor>` at generation time
and the resulting project ships embedding configuration, ingestion
endpoints, similarity-search retrieval, and Spring AI's
`RetrievalAugmentationAdvisor` wired into the primary agent's chat
client. Three backends are supported:

| Flag value | Backend | Where it runs | Use when |
|---|---|---|---|
| `pgvector` | PGVector inside the application's Postgres | Same datastore, isolated `vector` schema | You already have PostgreSQL and want one less moving part |
| `qdrant`   | Qdrant standalone container or cloud      | Separate gRPC server (default port 6334) | You need maximum vector-search performance, willing to run another service |
| `mongodb`  | MongoDB Atlas Vector Search               | Atlas cluster (community Mongo cannot serve `$vectorSearch`) | You're already on Atlas; metadata-rich documents fit Mongo's model |

Without `--vector-store`, the AIAgent module ships with the default
`KeywordKnowledgeRetriever` — a pure in-memory token-overlap scorer
against the small static `KnowledgeBase.ENTRIES` list. Useful for
demos and smoke tests; not a serious retrieval system.

## Generating a project with RAG

```bash
# PGVector — auto-adds SQLDatastore, forces --database=postgresql
trabuco init --modules=Model,Shared,API,AIAgent --vector-store=pgvector

# Qdrant — no datastore module needed; standalone server
trabuco init --modules=Model,Shared,API,AIAgent --vector-store=qdrant

# MongoDB Atlas — works standalone OR alongside NoSQLDatastore=mongodb
trabuco init --modules=Model,Shared,API,AIAgent --vector-store=mongodb
```

The CLI applies cross-flag rules and prints user-visible notices when
it has to auto-add a module or coerce a database choice (e.g., picking
`pgvector` without `SQLDatastore` triggers an automatic add). Conflicts
that can't be auto-resolved (`pgvector` + `--database=mysql`,
`mongodb` + `--nosql-database=redis`, any vector store without
`AIAgent`) fail fast with a clear message — no half-configured
projects.

## Choosing a backend

**Pick PGVector when** you already have PostgreSQL in the stack and
want the smallest possible operational footprint. One database server,
one backup target, one connection pool. Good for collections up to a
few million rows; HNSW index keeps query latency in the low-ms range.
You give up the highest raw QPS you'd get from a dedicated vector
service, but you save an entire deployment surface.

**Pick Qdrant when** you have hard latency targets, expect to scale
beyond the single-Postgres comfort zone, or want first-class filtered
search (`SearchRequest.filterExpression(...)` against rich metadata).
Qdrant Cloud handles the operations; self-host with one
`docker run -p 6333:6333 -p 6334:6334 qdrant/qdrant`.

**Pick MongoDB Atlas when** you're already on Atlas and want one fewer
infrastructure concern — Atlas Vector Search is fully managed.
**Important**: `$vectorSearch` is Atlas-exclusive. Generated code
compiles and starts against community / self-hosted Mongo, but
queries return errors. The vector index itself is **not** created by
Spring AI or Trabuco — see below.

## What gets generated

For all three backends:

- `aiagent/knowledge/VectorKnowledgeRetriever.java` — implements
  `KnowledgeRetriever` via Spring AI's `VectorStore.similaritySearch`.
- `aiagent/knowledge/DocumentIngestionService.java` — chunks via
  `TokenTextSplitter` (1000-token chunks, 50-token overlap by default),
  embeds, stores.
- `aiagent/knowledge/EmbeddingService.java` — thin wrapper over the
  configured `EmbeddingModel`.
- `aiagent/protocol/IngestionController.java` — `POST /ingest` and
  `POST /ingest/batch` endpoints, both gated by `@RequireScope("partner")`
  (legacy API-key path) or swappable to `@PreAuthorize("hasAuthority('SCOPE_admin')")`
  when JWT auth is on.
- `aiagent/config/KnowledgeBeansConfiguration.java` — wires the
  retriever and ingestion service as Spring `@Bean`s with
  `@Primary` on the vector path so it wins autowiring against the
  always-present `KeywordKnowledgeRetriever`.
- `application.yml` blocks for `spring.ai.embedding.transformer.*` and
  the chosen `spring.ai.vectorstore.*`.
- The `spring-ai-starter-model-transformers` dependency (default
  embedding model — 384-dim ONNX, no API key) plus the matching
  `spring-ai-starter-vector-store-*` artifact.

PGVector additionally generates:
- `aiagent/config/VectorFlywayConfig.java` — second Flyway bean for
  the dedicated `vector` schema, with its own history table
  (`flyway_schema_history_vector`). Domain Flyway is untouched.
- `aiagent/src/main/resources/db/vector-migration/V1__create_vector_schema.sql` —
  PGVector extension, `documents` table (UUID, content, JSONB metadata,
  `vector(384)`), HNSW index for cosine similarity, GIN index on
  metadata for filter pushdown.

PGVector with `SQLDatastore=postgresql` also gets the `VectorRagIntegrationTest` —
boots a real `pgvector/pgvector:pg16` Testcontainer, ingests sample
documents, runs similarity queries, asserts the round-trip works. Skipped
gracefully when Docker is unavailable.

## PGVector setup

The default. Almost nothing to do — boot the project and the Flyway
migration runs against the application's Postgres. Two notes worth
calling out:

**The pgvector extension lives in `public`, not `vector`.** This is
deliberate. `VectorFlywayConfig` runs migrations with
`defaultSchema=vector`, which sets Flyway's search_path so that
unqualified DDL targets the `vector` schema. By default
`CREATE EXTENSION vector` would land the type *and* its similarity
operators (`<=>`, `<->`, `<#>`) in the `vector` schema — and Spring
AI's PgVectorStore emits queries like
`SELECT … FROM vector.documents WHERE embedding <=> ? …` without
qualifying the operator, so it must be reachable on the connection's
default search_path. The migration uses
`CREATE EXTENSION IF NOT EXISTS vector WITH SCHEMA public` to put the
operators in `public` (always reachable) while keeping the
`documents` table isolated in `vector`.

**Embedding dimensions are pinned at 384** in V1. That matches the
default ONNX transformer (`intfloat/e5-small-v2`). If you swap the
embedding provider — see *Embedding providers* below — change the
`vector(384)` column type in the migration to match the new
dimensionality **before** the first ingestion. PostgreSQL will not let
you shrink or grow the type once rows exist.

## Qdrant setup

Local development:

```bash
docker run -d -p 6333:6333 -p 6334:6334 qdrant/qdrant
```

`6334` is the gRPC port (Spring AI's default); `6333` is the HTTP
dashboard. The generated `application.yml` defaults to
`localhost:6334`, no API key, no TLS. Spring AI auto-creates the
collection on first use (`initialize-schema=true`); collection vector
size is inferred from the active `EmbeddingModel`.

Qdrant Cloud:

```yaml
# Override via env, no code change
QDRANT_HOST=your-cluster.qdrant.io
QDRANT_API_KEY=eyJ...
QDRANT_USE_TLS=true
QDRANT_COLLECTION=production-knowledge
```

## MongoDB Atlas setup

This is the one backend that requires manual work outside the
generated project. Atlas Vector Search indexes are **not** part of the
collection schema and **cannot** be created by Spring AI's auto-config
— `initialize-schema=false` in the generated `application.yml`
reflects this. You must create the index in the Atlas UI / Atlas CLI /
Atlas Admin API before the application can serve retrieval queries.

The minimum index definition:

```json
{
  "fields": [
    {
      "type": "vector",
      "path": "embedding",
      "numDimensions": 384,
      "similarity": "cosine"
    },
    {
      "type": "filter",
      "path": "metadata.topic"
    }
  ]
}
```

- `path` matches `spring.ai.vectorstore.mongodb.path-name` in
  `application.yml` (default: `embedding`).
- `numDimensions` matches the active embedding model. If you swap
  providers, recreate the index with the new dimensionality.
- `similarity` should be `cosine` for the normalized embeddings the
  default transformer emits. Switch to `dotProduct` or `euclidean`
  only if your embedding provider demands it.
- `type: "filter"` entries enable `SearchRequest.filterExpression()`
  pushdown. Add one per metadata field you'll filter on.

The index name in `application.yml` defaults to `vector_index` —
change `MONGODB_VECTOR_INDEX` if your Atlas index has a different
name. Indexing is asynchronous; expect a few minutes after creation
before queries return results.

The generator emits a fallback `spring.data.mongodb.uri` when
`--vector-store=mongodb` is selected without `NoSQLDatastore=mongodb`
in the modules. When both are selected, the NoSQL block above the
vector-store block provides the connection (no duplicate-key YAML).

## Embedding providers

The default is the `spring-ai-starter-model-transformers` starter —
pure-Java ONNX runtime running `intfloat/e5-small-v2`, 384-dim. No
API key required, ~80 MB model downloads once and caches under
`~/.djl.ai/`. Acceptable for FAQ / internal-knowledge retrieval; not
state-of-the-art for high-stakes use.

To swap to OpenAI / Bedrock / Vertex AI:

1. Replace `spring-ai-starter-model-transformers` in
   `aiagent/pom.xml` with one of:
   - `spring-ai-starter-model-openai`
   - `spring-ai-starter-model-bedrock`
   - `spring-ai-starter-model-vertex-ai-embedding`
2. Configure the matching properties (`spring.ai.openai.api-key`,
   `spring.ai.bedrock.aws.region`, etc.).
3. Update embedding dimensionality wherever it appears:
   - PGVector: the `vector(N)` column type in
     `V1__create_vector_schema.sql` and
     `spring.ai.vectorstore.pgvector.dimensions` in `application.yml`.
   - Qdrant: collection is auto-recreated; just delete and re-ingest.
   - Mongo Atlas: recreate the index with the new `numDimensions`.

Common dimensionalities: OpenAI `text-embedding-3-small` is 1536,
`text-embedding-3-large` is 3072; Bedrock Titan v2 is 1024; Vertex
`textembedding-gecko@003` is 768.

No code changes are needed — `VectorKnowledgeRetriever` and
`DocumentIngestionService` resolve the active `EmbeddingModel` from
the Spring context, regardless of which starter provides it.

## How retrieval works

Two complementary surfaces:

**Tool-based retrieval (`KnowledgeTools.@Tool askKnowledge`)** — the
LLM explicitly invokes the tool when it decides it needs domain
knowledge. The tool delegates to the wired `KnowledgeRetriever` (vector
or keyword), serializes hits as a string, and returns them. This is
the path the LLM controls.

**Ambient RAG (`RetrievalAugmentationAdvisor`)** — wired on the
`PrimaryAgent`'s `ChatClient` only when a `VectorStore` bean is
present. Every user message triggers a similarity search; the top-K
hits (default `topK=4`) prepend the prompt as context **before** the
LLM call. The LLM doesn't need to ask — context is supplied
automatically. Particularly useful for domain disambiguation
("our product X means …") and grounding ("don't fabricate;
here are relevant docs").

Both paths use the same `VectorStore`, so ingestions are visible to
both. The advisor is provider-agnostic — it works with any
`VectorStore` Spring AI knows about.

## Why `@Bean` instead of `@Component` + `@ConditionalOnBean`

Phase B's first cut wired `VectorKnowledgeRetriever` as
`@Component @ConditionalOnBean(VectorStore.class)`. This silently
failed: Spring evaluates `@ConditionalOnBean` on user-scanned
components **before** the auto-configuration phase that registers the
`VectorStore` bean. Tried `@AutoConfiguration(afterName = "…PgVectorStoreAutoConfiguration")`
to delay the evaluation; still failed because the `OnBeanCondition`
evaluator doesn't honour the auto-config processing order in the way
the docs suggest.

The fix is generator-time conditionality:
- `KeywordKnowledgeRetriever` carries `@Component` (always wired
  whenever AIAgent is present).
- `KnowledgeBeansConfiguration` is emitted **only when** `--vector-store`
  is set, as a regular `@Configuration` whose `@Bean` methods take
  `VectorStore` as a parameter. Plain Spring DI, no conditionals.
- `@Primary` on the vector retriever resolves the autowiring
  tie-break against the keyword bean.

When Trabuco emits the vector files, the matching Spring AI starter
is also in the pom, so the `VectorStore` bean is guaranteed to exist
at runtime. We don't need to gate on it — we know it's there.

## Ingesting documents

Programmatic API (inside the application):

```java
@Autowired
DocumentIngestionService ingest;

ingest.ingest(
    "Long-form text describing the topic …",
    Map.of("topic", "billing", "version", "v3"));
```

REST endpoint:

```bash
curl -X POST http://localhost:8080/ingest \
  -H "Content-Type: application/json" \
  -H "X-Api-Key: $PARTNER_KEY" \
  -d '{
    "id": "doc-001",
    "content": "Long-form text describing the topic …",
    "metadata": {"topic": "billing"}
  }'
```

The endpoint is gated by `@RequireScope("partner")` (the legacy
API-key path). When you flip `trabuco.auth.enabled=true` to switch to
JWT, swap to `@PreAuthorize("hasAuthority('SCOPE_admin')")` — see the
comment block in `IngestionController.java`.

`TokenTextSplitter` chunks long inputs automatically (default 1000
tokens with 50-token overlap, sentence-aware). Override via the
`DocumentIngestionService` constructor when you need finer control.

## Testing

When the project is generated with `--vector-store=pgvector` and
`SQLDatastore=postgresql`, Trabuco emits `VectorRagIntegrationTest` —
a `@SpringBootTest` that boots a `pgvector/pgvector:pg16`
Testcontainer, ingests three documents, queries, and asserts the
round-trip works. Three test cases cover:

1. The `@Primary` autowiring picks `VectorKnowledgeRetriever` over
   the keyword fallback.
2. Ingestion + retrieval round-trips end-to-end (Flyway migrations,
   embedding, INSERT, SELECT with `<=>` cosine operator).
3. Queries with no semantic match return a non-null list (don't
   throw).

The test is annotated `@Testcontainers(disabledWithoutDocker = true)`
so CI runners without Docker continue to pass the rest of the suite.
First run downloads the ~80 MB ONNX model and the ~150 MB pgvector
image; subsequent runs reuse the caches.

The test deliberately does **not** assert specific top-K ordering —
that depends on the embedding model's semantic precision, which
Trabuco doesn't control, and pinning to a particular ranking makes
the test flaky across model revisions.

Qdrant and Mongo Atlas integration tests are deferred — Qdrant has a
Testcontainer image but isn't yet wired in the generator; Mongo Atlas
cannot be containerized (vector search is Atlas-exclusive).

## Troubleshooting

**`ERROR: operator does not exist: vector <=> vector`** — the
extension landed in the wrong schema. Confirm V1 was generated with
`WITH SCHEMA public`. If you bootstrapped against an existing database
that had the extension in another schema, drop and recreate it
(`DROP EXTENSION vector CASCADE; CREATE EXTENSION vector WITH SCHEMA public;`)
or set the connection's search_path to include the schema where it
lives.

**`expected N dims, got M`** — embedding dimensionality mismatch.
You swapped embedding providers without updating the schema. Drop
the table (or the collection / index), update `vector(N)` /
`numDimensions`, re-ingest.

**`No qualifying bean of type DocumentIngestionService`** — you've
got `--vector-store` set but the matching starter isn't on the
classpath, or you copied `IngestionController` into a project that
wasn't generated with a vector store. Re-run the generator or pull
the starter dep manually.

**MongoDB Atlas queries return empty even with rows present** — the
vector index isn't created or is still building. Atlas indexing is
asynchronous; check the index status in the Atlas UI before
debugging the application.

**Qdrant context fails to start in tests** — the auto-config tries
to instantiate the gRPC client at startup, which is fine, but the
test environment may not have Qdrant running. Use the
`@Testcontainers` pattern from `VectorRagIntegrationTest` (with the
appropriate Qdrant container image) once we add it, or skip the
context-loading test in CI when Qdrant isn't available.
