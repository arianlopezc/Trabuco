---
name: migrate
description: Migrate an existing Java repository in place into a Trabuco-shaped multi-module project. Drives the 14-phase orchestrated flow with specialized subagents, dependency-aware phasing (legacy CI keeps working at every phase boundary), per-phase approval gates, and atomic rollback via git tags. Use when the user has an existing Spring Boot or other JVM project and wants it transformed into Trabuco's structure.
user-invocable: true
allowed-tools: [mcp__trabuco__migrate_assess, mcp__trabuco__migrate_skeleton, mcp__trabuco__migrate_module, mcp__trabuco__migrate_config, mcp__trabuco__migrate_deployment, mcp__trabuco__migrate_tests, mcp__trabuco__migrate_activate, mcp__trabuco__migrate_finalize, mcp__trabuco__migrate_status, mcp__trabuco__migrate_rollback, mcp__trabuco__migrate_decision, mcp__trabuco__migrate_resume, mcp__trabuco__get_project_info, Read, Glob, Grep]
argument-hint: "[/path/to/repo]"
---

# Migrate an existing Java project into a Trabuco-shaped project

This skill drives the 14-phase migration. It is the user-facing entry point.
The actual reasoning is delegated to the **`trabuco-migration-orchestrator`** subagent,
which coordinates 13 specialized subagents (one per migration phase).

## When to invoke

- User says "migrate this project to Trabuco" / "convert this codebase to Trabuco" /
  "trabuco migrate <path>" / similar.
- User has an existing JVM repository they want transformed into Trabuco's
  multi-module Maven structure with conventions intact.

## When NOT to invoke

- User wants a brand-new project → use `/trabuco:new-project` instead.
- User has a Trabuco project already and wants to extend it → use `/trabuco:add-module`
  or `/trabuco:extend`.
- Repository is non-JVM (Python, Go, JS/TS) → tell the user this is out of
  scope; Trabuco only migrates JVM code. Frontend in a monorepo is left untouched.

## Flow

1. **Hand off to the orchestrator subagent.** Invoke `trabuco-migration-orchestrator`
   with the user's argument (the repo path) plus their stated goals. The
   orchestrator:
   - Runs preflight (clean working tree, on a branch, repo has commits).
   - Runs `migrate_assess` to scan the source and produce
     `.trabuco-migration/assessment.json` (the no-out-of-scope contract).
   - Walks the user through 14 phases, presenting diffs and gates at each.
   - Records user decisions via `migrate_decision`.
   - Rolls back via `migrate_rollback` if the user rejects a phase.

2. **Do NOT do the work yourself.** This skill is a router. The orchestrator
   subagent owns the migration. Your only job here is to invoke the right
   subagent and surface its summary.

## Rules for the orchestrator subagent

- **No out-of-scope.** Every change must be grounded in evidence from
  `.trabuco-migration/assessment.json`. Specialists that propose unevidenced
  changes are auto-rejected by the orchestrator's diff-inspection layer.
- **Dependency-aware phasing.** Enforcement mechanisms (Maven Enforcer,
  Spotless, ArchUnit, Jacoco threshold) are deferred until Phase 12. Skeleton
  generates a "migration-mode" parent POM with skip flags. Legacy CI must keep
  working at every phase boundary.
- **No invented infrastructure.** The deployment phase only adapts what the
  legacy already has. Never adds CI/CD, monitoring, or deployment scaffolding
  that wasn't there.
- **Safe over fast.** Default gate granularity is per-phase; the user can
  request `--per-aggregate` for high-risk phases (Datastore, API).

See `docs/migration-guide.md` for the user-facing walkthrough of the phases, gates, and rollback model.
