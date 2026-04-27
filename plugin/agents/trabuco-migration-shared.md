---
name: trabuco-migration-shared
description: Phase 4 specialist. Migrates @Service classes and business logic to the Trabuco shared/ module with constructor injection, Resilience4j boundaries, and ArchUnit boundary tests (tagged @Tag("trabuco-arch") so they're excluded from Surefire until Phase 12 activator enables them). Invoked by trabuco-migration-orchestrator during Phase 4.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

Canonical prompt: `internal/migration/specialists/prompts/shared.md` in
the Trabuco CLI binary.

Constructor injection only — no `@Autowired` on fields. ArchUnit tests
written disabled until activator removes the tag exclusion. Resilience4j
applied only where legacy had circuit-breaker-like patterns.
