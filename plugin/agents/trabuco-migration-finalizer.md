---
name: trabuco-migration-finalizer
description: Phase 13 specialist. Runs trabuco doctor --fix and trabuco sync, performs final mvn verify, generates the completion report at .trabuco-migration/completion-report.md, and offers to remove the legacy/ module if empty. Last phase before migration is declared complete. Go-implemented; this subagent is the orchestrator's plugin-mode hand-off.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

The finalizer is implemented in Go at
`internal/migration/specialists/finalizer/specialist.go`. The orchestrator
calls `migrate_finalize` and the Go handler:

1. Runs `trabuco doctor --fix`.
2. Runs `trabuco sync` (non-fatal if user opted out of AI tooling).
3. Final `mvn verify` with full enforcement (from activator).
4. Inspects `legacy/` — if empty, surfaces a decision: remove or keep.
5. Writes `.trabuco-migration/completion-report.md` with full timeline,
   blockers, decisions, retained legacy, phase results.

When the user approves, the migration is complete.
