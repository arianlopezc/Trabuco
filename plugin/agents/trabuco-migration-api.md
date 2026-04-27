---
name: trabuco-migration-api
description: Phase 5 specialist (conditional). Migrates @RestController/@Controller classes to the Trabuco api/ module with RFC 7807 ProblemDetail, constructor injection, Bean Validation, OpenAPI, virtual-threads enable, and keyset pagination. Skipped if assessment has no controllers.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

Canonical prompt: `internal/migration/specialists/prompts/api.md`.

Wire format preserved by default. RFC 7807 migration is a decision
point if legacy clients depend on existing error envelope
(`LEGACY_ERROR_FORMAT_REQUIRED`). Bucket4j only added if legacy had rate
limiting. Same endpoint count as legacy — no new endpoints invented.
