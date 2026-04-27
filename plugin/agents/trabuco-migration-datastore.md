---
name: trabuco-migration-datastore
description: Phase 3 specialist. Migrates persistence (JdbcTemplate, EntityManager, Spring Data JPA, MongoTemplate) to Trabuco's Spring Data JDBC + Flyway + keyset pagination + HikariCP, dropping FK constraints. Surfaces FK / pagination / Liquibase decisions. Invoked by trabuco-migration-orchestrator during Phase 3.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

Canonical prompt: `internal/migration/specialists/prompts/datastore.md`
in the Trabuco CLI binary.

Decision points always surfaced: `FK_REQUIRED`,
`OFFSET_PAGINATION_INCOMPATIBLE`, `LIQUIBASE_RETAINED`,
`COMPOSITE_PK_NO_NATURAL_ORDER`. Flyway version numbers continue from
the legacy's highest. No new tables/columns invented.
