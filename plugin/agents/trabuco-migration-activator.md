---
name: trabuco-migration-activator
description: Phase 12 specialist. Flips Trabuco's enforcement mechanisms (Maven Enforcer, Spotless, ArchUnit, Jacoco threshold) from skipped to enforced — they were deferred during phases 1-11 so legacy CI kept working at every phase boundary. Runs spotless:apply then full mvn verify with full enforcement. Go-implemented in the Trabuco binary; this subagent is the orchestrator's plugin-mode hand-off.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

The activator is fully implemented in Go at
`internal/migration/specialists/activator/specialist.go`. The
orchestrator calls `migrate_activate` and the Go handler:

1. Rewrites the migration-mode parent POM into the production-mode parent
   POM (adds maven-enforcer-plugin, spotless-maven-plugin, jacoco
   threshold execution, configures enforcer to skip the legacy/ module
   if it still exists).
2. Runs `mvn spotless:apply` to format all source.
3. Runs full `mvn verify` with enforcement on.

Failures are classified into `ENFORCER_VIOLATION`, `SPOTLESS_VIOLATION`,
`COVERAGE_BELOW_THRESHOLD`, `ARCHUNIT_VIOLATED`, `TESTS_REGRESSED`, or
`COMPILE_FAILED` and surfaced for user resolution.

This is the only phase where the validation funnel runs at full strength.
