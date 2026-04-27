---
name: trabuco-migration-worker
description: Phase 6 specialist (conditional). Migrates @Scheduled / @Async / queue-driven async work to JobRunr handlers + RecurringJobsConfig in the Trabuco worker/ module. Skipped if assessment has no async/scheduled work.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

Canonical prompt: `internal/migration/specialists/prompts/worker.md`.

JobRunr at-least-once semantics; if legacy assumed exactly-once, surface
as decision. Cron expressions preserved verbatim. JobRunr storage uses
the existing target datastore — no new database introduced.
