# Migrating an existing Spring Boot project to Trabuco

Trabuco 1.10 ships an orchestrated, 14-phase migration that transforms an
existing Spring Boot 2.x or 3.x repository in place into a Trabuco-shaped
multi-module Maven project. The migration is LLM-driven: each phase is
owned by a specialist that proposes structured changes you approve at a
gate before they're committed.

This guide covers what the migration does, how to run it, and how to
recover when something goes sideways.

## Prerequisites

The migration mutates your repository. Before you start:

1. **Anthropic API access.** Each phase invokes Claude. Set
   `ANTHROPIC_API_KEY` or run `trabuco auth login` once.
2. **Git.** Clean working tree, on a branch (not detached), at least one
   commit. The migration creates per-phase git tags
   (`trabuco-migration-phase-N-pre/post`) and uses them as rollback
   anchors. **Do not migrate against your only copy of the source —
   work on a branch you can throw away.**
3. **Maven source.** Trabuco 1.10 supports Maven (`pom.xml`). Gradle
   sources are not currently supported; running on Gradle produces a
   `NON_MAVEN_BUILD_SYSTEM` blocker at Phase 0 with conversion guidance.
4. **JDK matching the project's target.** The generated code targets a
   specific Java version (17, 21, 25). The build runtime —
   whatever `java -version` reports — must match. Mismatches surface
   plugin failures (ArchUnit, Spotless, Enforcer) that look obscure.
   Trabuco runs a preflight check and refuses to start a phase if
   `java -version` doesn't match `state.targetConfig.javaVersion`.

## The 14 phases

| # | Phase | What runs | Code changes? |
|---|-------|-----------|---------------|
| 0 | Assessment | Catalogs every artifact in your source | No (writes assessment.json) |
| 1 | Skeleton | Generates the multi-module skeleton in migration mode (enforcement OFF), wraps your source in `legacy/` | Yes |
| 2 | Model | Migrates entities/DTOs/events to `model/` as records or Immutables | Yes |
| 3 | Datastore | Migrates repositories to Spring Data JDBC + Flyway (or Spring Data MongoDB), drops FK constraints, swaps offset → keyset pagination | Yes |
| 4 | Shared | Migrates services to `shared/` with constructor injection + Resilience4j; emits ArchUnit boundary tests (tagged `trabuco-arch`) | Yes |
| 5 | API | Migrates controllers to `api/` with RFC 7807 ProblemDetail | Yes |
| 6 | Worker | Migrates `@Scheduled` jobs to JobRunr `JobRequestHandler` | Yes |
| 7 | EventConsumer | Migrates `@KafkaListener` / `@RabbitListener` etc. to `eventconsumer/` | Yes |
| 8 | AIAgent | Migrates AI integration if any (Spring AI, LangChain4j) | Yes |
| 9 | Configuration | Splits `application.properties`/yml per module, adds OpenTelemetry, replaces hardcoded credentials with env vars | Yes |
| 10 | Deployment | Adapts the legacy CI/CD workflows to the multi-module structure. **Strict: never invents a pipeline** | Yes |
| 11 | Tests | Per-test KEEP / ADAPT / DISCARD / CHARACTERIZE-FIRST decisions | Yes |
| 12 | Activation | Flips Maven Enforcer / Spotless / ArchUnit / Jacoco threshold from skip to enforce, runs spotless:apply, then full `mvn verify` | Yes |
| 13 | Finalization | Runs `trabuco doctor --fix` and `trabuco sync`, writes the completion report, optionally removes `legacy/` | Yes |

Phases 7 and 8 self-skip with `not_applicable` when the source has no
listeners / AI integration. Phase 10 self-skips when there's no CI in
the repo.

The "deferred enforcement" design — Maven Enforcer, Spotless, ArchUnit, and
the Jacoco coverage threshold all start the migration in skipped form — means
your legacy CI keeps working at every phase boundary during the migration.
Enforcement only flips ON at Phase 12.

## Running the migration

### Step by step (recommended for the first run)

```bash
# Phase 0 — produces .trabuco-migration/assessment.json. Review it
# carefully; it's the no-out-of-scope contract for every later phase.
trabuco migrate assess /path/to/your/repo

# Phase 1 — multi-module skeleton; legacy/ now wraps your source
trabuco migrate skeleton /path/to/your/repo

# Phases 2-8 — module specialists
trabuco migrate module /path/to/your/repo --module=model
trabuco migrate module /path/to/your/repo --module=sqldatastore   # or nosqldatastore
trabuco migrate module /path/to/your/repo --module=shared
trabuco migrate module /path/to/your/repo --module=api
trabuco migrate module /path/to/your/repo --module=worker
trabuco migrate module /path/to/your/repo --module=eventconsumer
trabuco migrate module /path/to/your/repo --module=aiagent

# Phases 9-13
trabuco migrate config     /path/to/your/repo
trabuco migrate deployment /path/to/your/repo
trabuco migrate tests      /path/to/your/repo
trabuco migrate activate   /path/to/your/repo
trabuco migrate finalize   /path/to/your/repo
```

Each invocation runs one phase, presents the changes at a gate, and
asks `[a]pprove / [e]dit and approve / [r]eject?`.

### Autopilot

```bash
trabuco migrate run /path/to/your/repo
```

Same sequence, gating at every phase. Use this once you trust the
output of the per-phase form.

### Inspecting state

```bash
trabuco migrate status /path/to/your/repo
```

## Gates: approve, edit, reject

At every phase the orchestrator presents a summary plus a list of
items. Each item has a state:

- **applied** — change ready to commit (with `file_writes` + source
  evidence)
- **blocked** — specialist hit a `BlockerCode`. You must fix and re-run,
  or accept the alternative the specialist offers.
- **requires_decision** — specialist needs a user choice (FK constraints
  vs app-level integrity, offset vs keyset pagination, etc.). Use
  `trabuco migrate decision` to record the choice, then re-run the
  phase.
- **not_applicable** — phase has nothing to do (e.g. EventConsumer with
  no listeners).
- **retained_legacy** — kept in `legacy/` rather than migrated, by user
  decision.

Your three gate options:

- **`a` (approve)** — commit the changes, tag `trabuco-migration-phase-N-post`, advance.
- **`e` (edit and approve)** — provide a hint; the orchestrator rolls
  back to `pre`, re-runs the specialist with your hint as `UserHint`.
- **`r` (reject)** — roll back and stop.

## Decisions

When a specialist emits `requires_decision`, record the choice:

```bash
trabuco migrate decision /path/to/your/repo \
  --id=fk-constraint-decision \
  --choice="Drop FK constraints"
```

Decisions persist in `state.json`. On the next phase invocation, every
recorded decision for that phase is appended to the specialist's
`UserHint`, so the LLM applies your choice without re-asking. To
override an earlier decision, run `migrate decision` again with the
same id — it replaces in place.

## Rollback

Per-phase rollback to the `pre` tag:

```bash
trabuco migrate rollback /path/to/your/repo --to-phase=3
```

This `git reset --hard`'s to `trabuco-migration-phase-3-pre`, runs
`git clean -fd` to remove untracked files from the failed phase, and
clears phases ≥ 3 from `state.json`. Decisions you've already
recorded in `state.json` survive.

## Troubleshooting

**`[JAVA_VERSION_MISMATCH_RUNTIME] build runtime mismatch`** — the
`java` on PATH doesn't match `state.targetConfig.javaVersion`. Set
`JAVA_HOME` (`/usr/libexec/java_home -v 21` on macOS), use Maven
Toolchains, or run from a JDK-pinned container.

**`COMPILE_FAILED`** — usually one of: missing dependency in a
module's `pom.xml` (the LLM forgot to update it), wrong package import
(LLM hallucinated the package of a class produced by an earlier
phase), or version mismatch in a `<parent>` block. Inspect
`.trabuco-migration/phase-N-{specialist}-raw.txt` for the LLM's raw
output, then either:
- run the phase again (LLMs vary across runs and may produce a clean
  output the second time), or
- use `migrate <phase> --to-phase=N` then re-run with `e` (edit and
  approve) and a hint describing the fix.

**`ENFORCER_VIOLATION`, `SPOTLESS_VIOLATION`, `ARCHUNIT_VIOLATED`** —
real violations Phase 12 (activation) detected. Fix the underlying
issue in the codebase, commit, and re-run `trabuco migrate activate`.
The violation log is preserved in the blocker note inside `state.json`.

**`NON_MAVEN_BUILD_SYSTEM`** — Trabuco 1.10 is Maven-only. Convert
your project (`gradle init --type pom` for Gradle) and re-run.

## State on disk

Everything Trabuco writes lives under `.trabuco-migration/` at your
repo root:

```
.trabuco-migration/
├── state.json                    — current migration state
├── assessment.json               — Phase 0's catalog
├── phase-N-input.json            — what each specialist saw
├── phase-N-output.json           — what each specialist returned
├── phase-N-{name}-raw.txt        — raw LLM response (debug)
├── phase-N-report.md             — human-readable phase summary
├── completion-report.md          — Phase 13 final summary
└── lock.json                     — single-writer lock
```

The directory is `.gitignore`d during the migration. After Phase 13
you can delete it.

## Plugin / MCP usage

Every CLI command has an MCP equivalent registered by `trabuco mcp`,
so the same migration is drivable from Claude Code:

| CLI | MCP tool |
|-----|----------|
| `trabuco migrate assess` | `migrate_assess` |
| `trabuco migrate skeleton` | `migrate_skeleton` |
| `trabuco migrate module --module=X` | `migrate_module` (`module=X`) |
| `trabuco migrate config` | `migrate_config` |
| `trabuco migrate deployment` | `migrate_deployment` |
| `trabuco migrate tests` | `migrate_tests` |
| `trabuco migrate activate` | `migrate_activate` |
| `trabuco migrate finalize` | `migrate_finalize` |
| `trabuco migrate status` | `migrate_status` |
| `trabuco migrate rollback --to-phase=N` | `migrate_rollback` (`to_phase=N`) |
| `trabuco migrate decision --id=X --choice=Y` | `migrate_decision` |
| `trabuco migrate resume` | `migrate_resume` |

In plugin mode the gate is delegated to a subagent
(`trabuco-migration-orchestrator`) that presents each phase's diff in
the conversation and calls `migrate_decision` / `migrate_rollback`
based on the user's response.

## Test fixtures

Trabuco ships fixtures under `testdata/migration-fixtures/` you can
use to validate the feature end-to-end before pointing it at your own
repo:

- `spring-boot-27-monolith` — JPA + REST + scheduled job + GitHub
  Actions CI; exercises every blocker in the catalog.
- `spring-boot-27-mongo` — `@Document` entities + `MongoRepository`;
  exercises the NoSQLDatastore path.
- `spring-boot-27-kafka` — `@KafkaListener` consumer + KafkaTemplate
  publisher; exercises EventConsumer.
- `spring-boot-27-gradle` — Gradle source; exercises the
  `NON_MAVEN_BUILD_SYSTEM` blocker.

Copy any of them somewhere temporary, `git init && git add -A && git commit`,
then run the migration against the copy.

## Design principles

The migration is built around a few load-bearing ideas:

- **No-out-of-scope contract.** Phase 0 produces a structured assessment
  (`.trabuco-migration/assessment.json`) cataloging every file, controller,
  entity, listener, scheduled task, and CI artifact. Later phases are
  forbidden from inventing artifacts that aren't in that catalog — if the
  assessment shows no Kafka listeners, Phase 7 self-skips. This is what
  prevents specialists from "while we're at it" hallucinations.
- **Deferred enforcement.** Maven Enforcer, Spotless, ArchUnit, and the
  Jacoco coverage threshold all start the migration skipped. Your legacy
  CI keeps passing at every phase boundary. Phase 12 flips them ON in one
  place, all at once, after the structural migration is complete.
- **Per-phase atomic rollback.** Every phase brackets its work between two
  git tags (`trabuco-migration-phase-N-pre`, `-post`). Rejecting a phase
  resets to `-pre`; the orchestrator never leaves you in a half-applied
  state.
- **Gates over autopilot.** The orchestrator stops at every phase boundary
  and waits for explicit `approve`, `edit-and-approve`, or `reject`.
  Specialists propose; the user disposes.
- **Plugin/CLI parity.** The same Go specialists run whether you invoke
  `trabuco migrate` from the CLI or `/trabuco:migrate` from the Claude
  Code plugin. The plugin's orchestrator subagent is a thin conversational
  shell over the CLI's tooling.
