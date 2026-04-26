# Trabuco CLI — Migration Feature Redesign Plan

**Status:** Approved 2026-04-26 with all open decisions ratified by the user. Updated 2026-04-26 (second pass) to add dependency-aware phasing — enforcement mechanisms are deferred to a new Phase 12 so legacy CI continues to work at every phase boundary during migration.
**Target release:** Trabuco CLI **1.10.0** (single release containing the deletion of the old migrate feature and the entire new feature).
**Iteration scope:** First iteration. Future iterations will refine based on real-world migration runs.

---

## 1. Goal and non-goals

### Goal

Receive a path to a user's existing repository and transform it **in place** (same git checkout, no new repository) into a Trabuco-shaped multi-module Maven project, driven by an orchestrator agent that interacts with the user and dispatches to specialized subagents at each phase, with explicit user approval between phases and structured blocker reporting when transformation isn't possible.

The feature is exposed two ways:
- **Plugin (Claude Code et al.)** — an agentic UX where the orchestrator is the `trabuco-migration-orchestrator` subagent and gates surface as natural-language exchanges.
- **CLI (`trabuco migrate`)** — a self-sufficient terminal flow that reaches the same Go handlers and the same Claude API specialists, with gates surfaced as interactive terminal prompts. No IDE required.

Both surfaces share the same Go-side handlers, the same `.trabuco-migration/` state directory, the same specialist prompts, and the same validation gates.

### Non-goals

- **Mechanical transforms.** No two migrations are the same. Every change is the product of analytical reasoning over the specific source. The methodology is **100% LLM-driven**; AST recipes, codemods, and OpenRewrite are explicitly out of scope. The validation funnel exists to *verify* LLM output, not to perform transformation.
- **Imagining infrastructure that doesn't exist.** The migration only transforms what is present in the source. It never adds CI/CD pipelines, Docker registries, monitoring stacks, or deployment targets that the source did not already have. Any "while we're at it" thinking is forbidden by architectural contract (see §4).
- **Out-of-band scaffolding.** Trabuco-standard scaffolding (parent POM, multi-module skeleton, AI tooling files, base Dockerfiles for selected modules) is generated only because the user explicitly requested those modules in the Phase 0 target config — not because Trabuco "normally" produces them. The target config is the source of authority for scaffolding.
- **Multi-repo orchestration.** One repo at a time. Multi-repo / monorepo splits are out of scope for this iteration.
- **Continuous migration.** The migration completes once; it is not a long-running background reconciliation process.

---

## 2. Methodology

The methodology is a **composition** of established patterns at different layers, with all transformation done by LLM specialists:

| Layer | Pattern | Source |
|---|---|---|
| Macro structure | **Branch by Abstraction along the module DAG.** Trabuco's module hierarchy is the abstraction. Migrate bottom-up: Model → Datastore → Shared → API/Worker/EventConsumer/AIAgent | Humble, Hammant, Fowler |
| Per-module structure | **Vertical slices by aggregate.** One entity end-to-end before next; never leave a module half-migrated | Shopify componentization |
| Transformation engine | **LLM specialists for every change.** Each specialist reasons over source + target conventions and produces a justified patch with cited source evidence | User decision (2026-04-26) |
| Verification (not transformation) | **Six-step validation funnel** per change: lex/parse → compile → ArchUnit → unit tests → Testcontainers integration → optional behavioral equivalence vs legacy | Google funnel + Feathers characterization tests |
| Discovery | **Pre-flight assessment** as a separate phase producing a written modernization-assessment artifact before any transformation begins | AWS Q, Microsoft .NET Modernization, IBM watsonx |
| Human-in-the-loop | **Interrupt + checkpoint** at each phase boundary, three actions — approve / edit-and-approve / reject — with per-phase git tags for atomic rollback | Cursor checkpoint, LangGraph `interrupt()` |
| Per-phase tracking | **State machine with burndown**: `applied` / `blocked` / `requires-decision` / `not-applicable` / `retained-legacy` | Sourcegraph Batch Changes |

**Why LLM-only.** Mechanical recipes appear safe but conceal reasoning. "Convert OFFSET to keyset" requires knowing which column is monotonic. "Convert javax to jakarta" requires knowing whether a transitive dep still has javax-only versions. "Convert constructor injection" requires untangling circular dependencies. Each of these decisions is per-codebase. Treating any of them as mechanical creates a class of silent failures that recipe-based migration tools accept (Google's paper concedes ~50% time savings *in exchange for* a manual triage queue). For Trabuco's "safe over fast" target, mechanical recipes are net-negative — they trade reasoning for speed in a context where reasoning is the whole point.

**The validation funnel is verification, not transformation.** It catches LLM mistakes after the fact: a patch that doesn't compile, breaks ArchUnit boundaries, regresses tests, or violates module conventions is rejected and fed back to the same specialist with the failure context for retry. Specialists that fail repeatedly are escalated as `requires-decision` items.

---

## 3. Architecture

```
                        ┌──────────────────────────────────┐
                        │  trabuco-migration-orchestrator  │  (sole user-facing agent in plugin mode)
                        │   - owns state.json              │
                        │   - dispatches to specialists    │  (in CLI mode: orchestrator is a Go state
                        │   - presents diffs at gates      │   machine that calls the same specialists
                        │   - manages git tags             │   via Claude API)
                        └──────────────────────────────────┘
                                        │
       ┌─────────────────┬──────────────┼──────────────┬─────────────────┐
       ▼                 ▼              ▼              ▼                 ▼
  Assessor    Skeleton    Model     Datastore    Shared    API   Worker   EventConsumer  AIAgent  Config  Deployment  Test  Finalizer
  (Ph 0)      (Ph 1)      (Ph 2)    (Ph 3)       (Ph 4)    (Ph 5) (Ph 6)  (Ph 7)         (Ph 8)   (Ph 9)  (Ph 10)    (Ph 11) (Ph 12)
                                                                                                                       ▲
                                                                                              ┌────────────────────────┘
                                                                                              │   cross-cutting:
                                                                                              │   runs at every phase
                                                                                              │   to capture
                                                                                              │   characterization
                                                                                              │   tests + analyze
                                                                                              │   per-test fate
                                                                                              │
                                                                                          (test specialist)
```

**Specialist contract.** Every specialist:

1. Receives a task scope (which subset of the assessment artifact it should consider) plus the current state.json and the phase-specific target config.
2. Reads only the assessment artifact for source evidence — it does **not** re-scan the raw source. (The assessor is the only specialist that scans raw source.)
3. Produces output items, each in one of five states:
   - `applied` — patch ready, with `source_evidence: { file, lines, content_hash }` proving the change is grounded
   - `blocked` — fixed reason code (§7) with concrete file/line context and at least one alternative
   - `requires-decision` — multiple legal Trabuco-shaped outcomes; user must pick
   - `not-applicable` — nothing in the assessment matches this specialist's scope; phase is a no-op
   - `retained-legacy` — user explicitly accepted that this artifact stays in the `legacy/` module
4. Writes its output to `.trabuco-migration/phase-{N}-output.json` and stops. The orchestrator runs the validation funnel and presents the gate.

---

## 4. The "no out-of-scope" enforcement contract

This is the most important architectural rule in the design. It is enforced at four layers, not just by prompting.

**Layer 1 — Specialist prompt invariant.**
Every specialist's system prompt opens with this verbatim text:

> You are forbidden from proposing any change not directly grounded in evidence from the assessment artifact at `.trabuco-migration/assessment.json`. You may not introduce new files, configuration, workflows, dependencies, or features that do not correspond to an existing artifact in the source. If your scope yields no matching evidence, output `not-applicable` for the entire phase. You will be penalized for introducing scope.

**Layer 2 — Source evidence on every change.**
Every `applied` patch must carry a `source_evidence` field:
```json
{
  "file": "src/main/java/com/legacy/UserController.java",
  "lines": "12-58",
  "content_hash": "sha256:..."
}
```
The orchestrator validates: (a) the file exists in the source, (b) the line range exists, (c) the content at that range hashes to the claimed value. Any of these failing → patch rejected with `EVIDENCE_INVALID`.

**Layer 3 — Assessment artifact as the contract.**
The Phase 0 assessor produces a structured catalog of every artifact in the source the migration may touch (entities, controllers, services, listeners, jobs, configs, deployment files, etc.). All later specialists read assessment.json — never raw source. New scope additions require the orchestrator to re-invoke the assessor with explicit user approval, and the resulting delta is recorded as a separate decision in state.json.

**Layer 4 — Orchestrator-level diff inspection.**
Before showing the user a phase diff, the orchestrator runs a check pass:
- Every introduced file must be either (a) Trabuco-standard scaffolding requested in target config, or (b) a transformation of an artifact in assessment.json with traceable source_evidence
- Every modified file must correspond to an assessment.json entry
- Diffs that fail are auto-rejected with the reason surfaced to the user

**The `not-applicable` state is critical.** It is the explicit happy path for "this phase has nothing to do." If a source has no CI/CD files, the deployment specialist outputs `not-applicable` and the phase is skipped. If a source has no async work, the Worker phase is `not-applicable`. The orchestrator presents these as "Phase X — skipped (nothing to migrate)" rather than producing empty diffs or speculative additions.

**Implication for skeleton-builder (Phase 1).** Trabuco-standard scaffolding (parent POM, .gitignore, .editorconfig, AI tooling files, module Dockerfiles for chosen modules) is generated based on the **target config from Phase 0**, not "because Trabuco normally generates it." The user picks modules, CI provider, AI agents in Phase 0; Phase 1 generates exactly what was selected. If the user's target config specifies no CI provider, no `.github/workflows/` is created. The deployment specialist (Phase 10) then has no scaffolding to compete with — it only adapts the legacy CI/CD if the legacy had any. Both ends are constrained by source/target evidence.

---

## 5. Dependency-aware phasing

The migration runs over many phases. Each phase adds new code, new modules, or new configuration. **A phase must never break the project's ability to commit, push, build, or run CI** — even if the user pushes mid-migration.

This is not just a convenience principle; it shapes the phase ordering. Trabuco's standard generation includes enforcement mechanisms (Maven Enforcer, Spotless format check, ArchUnit module-boundary tests) that exist precisely to **catch** code that violates the conventions. During migration, the legacy code violates these conventions by definition — that's why we're migrating it. If those mechanisms are active during migration, every push fails CI.

**The principle:** *enforcement of conformance is enabled only after conformance has been achieved.*

### 5.1 Classification of generated mechanisms

| Mechanism | Active during migration? | Reason |
|---|---|---|
| Multi-module structure, parent POM (structural) | Yes | Multi-module Maven handles empty modules trivially |
| Module Dockerfiles, docker-compose.yml | Yes | Passive; not invoked unless user runs them |
| `.gitignore`, `.editorconfig`, `mvnw`, `.mvn/` | Yes | Passive |
| `.trabuco.json` metadata | Yes | Passive read-only file |
| AI tooling rule files (`.claude/rules/*.md`) | Yes | Guidance for AI agents, not build enforcement |
| `application.yml`, `logback-spring.xml` | Yes | Configuration, not enforcement |
| Flyway migrations | Yes | Only run when datastore is exercised |
| Per-module unit + integration tests | Yes (per-module) | Validation funnel ensures they pass at every checkpoint |
| **Maven Enforcer plugin** (`bannedDependencies`, `dependencyConvergence`, `requireMavenVersion`, `requireJavaVersion`) | **No — deferred** | Fails on legacy `javax.*`, JUnit 4, version conflicts |
| **Spotless format check** | **No — deferred** | Legacy code isn't Google Java Format |
| **ArchUnit module-boundary tests** | **No — deferred** (written disabled) | Legacy code violates Trabuco's module boundaries |
| **Jacoco coverage threshold** | **No — deferred** | Partial migration → low coverage |
| **CI enforcement steps** (e.g., `mvn spotless:check`) | **No — deferred** | Strict CI checks would fail mid-migration |
| **`.claude/settings.json` enforcement hooks** | **No — deferred** (or migration-friendly variant) | Could fire on every file edit during migration |

### 5.2 The deferred-enforcement mechanism

The skeleton-builder generates a **"migration-mode" parent POM** that is structurally identical to Trabuco's standard parent POM, with these specific differences:

- `maven-enforcer-plugin` block has `<skip>true</skip>` in its execution configuration, OR is omitted entirely
- `spotless-maven-plugin` block has `<skip>true</skip>`, OR is omitted entirely
- Jacoco's `check` execution (the threshold-enforcing one) is omitted; the `prepare-agent` and `report` executions remain so coverage data is still collected

**ArchUnit tests** in the Shared module are written with a JUnit 5 tag (`@Tag("trabuco-arch")`). Surefire is configured during migration to exclude this tag. The activation phase removes the exclusion, enabling the tests.

**`.claude/settings.json` hooks** are generated in a migration-friendly variant: file-create and file-modify hooks that would block edits are commented out or set to advisory mode. The `stop` hook (which runs after generation) is disabled. The activation phase swaps in the production-mode settings.

### 5.3 The activation phase (new)

A new **Phase 12 — Enforcement activation** is inserted before finalization. Its sole purpose is to flip the deferred mechanisms from off to on, and to verify everything still passes once they're on. The activation specialist:

1. Runs `mvn spotless:apply` to format all code (so the subsequent check passes)
2. Removes `<skip>true</skip>` from Maven Enforcer in parent POM (or adds the plugin if it was omitted)
3. Removes `<skip>true</skip>` from Spotless in parent POM
4. Adds Jacoco threshold-enforcement execution
5. Removes the `trabuco-arch` tag exclusion from Surefire so ArchUnit tests run
6. Swaps `.claude/settings.json` migration-mode → production-mode
7. If a CI workflow exists and was adapted in Phase 10, optionally adds enforcement steps (`mvn spotless:check`, full `mvn verify`) — only if the legacy CI already had a "lint" or "format" stage; otherwise leaves CI unchanged (no out-of-scope additions)
8. Configures Maven Enforcer to skip the `legacy/` module if it still exists (so retained legacy doesn't trigger enforcer)
9. Runs full `mvn verify` end-to-end with all enforcement on
10. If any failures surface, presents them with concrete remediation; user can choose to fix-and-retry, accept-with-caveats, or reject the activation
11. Approval gate: user reviews the activation diff and confirms

This phase is gated like every other phase. The user can reject and the activation rolls back to migration-mode. They can then continue iterating on the code.

### 5.4 Legacy CI continues to work

The legacy CI was designed for the legacy code structure. After Phase 1 (which wraps legacy in a `legacy/` Maven module), the legacy CI's `mvn verify` at root sees a multi-module project. New empty modules pass trivially (zero source files); only the legacy module has content; the legacy module continues to build and test. Migration-mode parent POM has enforcer/spotless skipped, so they don't fire. The build passes.

As migration proceeds, new modules accumulate code. The validation funnel ensures each module's unit and integration tests pass before the orchestrator commits the phase. So at every phase boundary, `mvn verify` at root passes.

The deployment specialist (Phase 10) updates legacy CI for **path changes only** — if the legacy CI did `cd webapp && mvn package` and webapp moved to `legacy/webapp`, the CI is updated to the new path. No enforcement steps added in Phase 10. The enforcement specialist (Phase 12) is the only one allowed to add enforcement steps to CI, and only if the legacy CI already had a stage of that nature.

### 5.5 Implementing-time discipline

For us as we develop the migration:

- The skeleton-builder template produces both **migration-mode** and **production-mode** parent POMs from the same Trabuco template by passing a `migrationMode: true` flag. This keeps the templates DRY — there is no separate "migration parent POM template."
- The activation specialist's primary responsibility is the migration→production POM swap, plus the corresponding swaps in `.claude/settings.json` and Surefire test exclusions.
- Validation funnel rules during migration phases skip enforcer/spotless/archunit checks (because they're deliberately off). The funnel explicitly verifies they're still off in `migration-mode` and on in `production-mode`. The activation phase is the only place where the funnel cares about enforcement passing.

---

## 6. Phase plan

```
Phase 0  — Intake & Assessment              ─┐
Phase 1  — Skeleton bootstrap (migration-    │
           mode parent POM, enforcement OFF) │
Phase 2  — Model                             │
Phase 3  — Datastore                         │
Phase 4  — Shared (services; ArchUnit tests  │
           written but tagged disabled)      │
Phase 5  — API           (conditional)       ├─ each ends with an APPROVAL GATE
Phase 6  — Worker        (conditional)       │  (approve / edit-and-approve / reject)
Phase 7  — EventConsumer (conditional)       │
Phase 8  — AIAgent       (conditional)       │
Phase 9  — Configuration                     │
Phase 10 — Deployment    (conditional)       │  ← legacy CI/CD path adaptation only,
           — path adaptations only           │    no enforcement steps added
Phase 11 — Test analysis & migration         │  ← per-test decision: keep/adapt/discard/characterize
Phase 12 — Enforcement activation            │  ← flips Maven Enforcer, Spotless, ArchUnit,
           (the deferred mechanisms turn on) │    Jacoco threshold, settings.json hooks ON
Phase 13 — Finalization                     ─┘
```

**Conditional phases** are skipped automatically when their precondition isn't met:
- Phase 5 skipped if no API module selected
- Phase 6 skipped if no Worker module selected
- Phase 7 skipped if no EventConsumer module selected
- Phase 8 skipped if no AIAgent module selected
- Phase 10 skipped if `not-applicable` (no legacy CI/CD found)
- Phase 12 always runs (enforcement activation is non-optional; even minimal projects need it)

**Gate behavior at every phase:**
- **Approve** — orchestrator commits the phase, tags `trabuco-migration-phase-{N}-post`, advances state.json
- **Edit-and-approve** — user supplies guidance ("the OrderQueue belongs in Worker, not Shared"); orchestrator re-runs the relevant specialist with that hint; new diff presented
- **Reject** — orchestrator resets to `trabuco-migration-phase-{N}-pre` tag, marks phase as `user-rejected` in state.json, offers retry-with-different-config or abort

**Gate granularity is configurable.** Default is one gate per phase. `--per-aggregate` flag breaks Phases 2–8 into one gate per aggregate (per entity, per controller, per service domain). The "safe over fast" mode used by default is `--per-aggregate` for Phases 3 (datastore) and 5 (API), where the blast radius of a wrong decision is highest, but per-phase for the rest. We can iterate on these defaults after real-world runs.

**Pre-Phase 0 hard preconditions** (orchestrator refuses to start if any fail):
- Repo is a git repo with at least one commit
- Working tree is clean (no uncommitted changes)
- Branch is checked out (no detached HEAD)
- No prior `.trabuco-migration/` from an aborted run, OR user explicitly chose to resume/restart
- Trabuco CLI auth is configured (`trabuco auth status` returns valid)
- Source repo size is bounded (warning if >100k LOC; abort if >500k LOC pending future iteration)

---

## 7. Specialist roster

Each specialist is a separately-defined LLM agent (subagent file in plugin mode, prompt template in CLI mode) with narrow scope and tool restrictions. Specialists are **stateless** — their full context comes from the assessment artifact, the target config, and the phase-specific input. They write their output and stop.

### 7.1 `trabuco-migration-assessor` (Phase 0)
**Scope.** Read the source repository top-to-bottom. Catalog every artifact relevant to migration. Identify the source's build system, framework, runtime, persistence, broker, deployment infrastructure. Classify each artifact for downstream specialists. Identify blockers up front.
**Tools allowed.** Read, Glob, Grep. No Write, no Edit, no Bash beyond inspection.
**Output.** `.trabuco-migration/assessment.json` — structured index of:
- Build system (maven | gradle | other)
- Framework (spring-boot-2.x | spring-boot-3.x | quarkus | micronaut | helidon | jaxrs | servlet | non-jvm | mixed)
- Java version detected
- Persistence (jpa | spring-data-jdbc | jdbc-template | mybatis | mongodb | redis | none) + per-aggregate entity catalog
- Web layer (spring-mvc | webflux | jaxrs | none) + per-controller catalog
- Async work (scheduled-annotation | async-annotation | quartz | other | none) + per-job catalog
- Messaging (kafka | rabbitmq | sqs | pubsub | jms | none) + per-listener and per-publisher catalog
- AI/LLM integration (yes/no, with framework if yes)
- CI/CD systems detected (github-actions | gitlab-ci | jenkins | circleci | azure-pipelines | travis | argo | flux | helm | k8s-manifests | terraform | other | none) with file paths
- Test framework (junit-4 | junit-5 | spock | testng) + per-test catalog
- Hardcoded secrets (mandatory blocker if found)
- Recommended Trabuco target config based on findings

### 7.2 `trabuco-migration-skeleton-builder` (Phase 1)
**Scope.** Take the user-approved target config and generate the Trabuco multi-module skeleton inside the existing repo in **migration mode** — enforcement mechanisms are deferred (see §5). Wrap existing source in a `legacy/` Maven module (the **adjacent strategy**, ratified). Verify the legacy module compiles unchanged.
**Tools allowed.** Write, Edit, Read, Bash (`mvn compile`).
**Behavior.**
- Generates **migration-mode parent `pom.xml`** — Trabuco-standard parent POM with these specific deferrals:
  - `maven-enforcer-plugin` execution has `<skip>true</skip>`
  - `spotless-maven-plugin` execution has `<skip>true</skip>`
  - Jacoco's `check` execution (threshold-enforcing) is omitted; `prepare-agent` and `report` remain
  - Surefire is configured to exclude tests tagged `trabuco-arch`
- Generates `.gitignore`, `.editorconfig`, `mvnw`, `.mvn/`, `.trabuco.json`
- Generates AI tooling files (`.claude/`, `.cursor/`, `.github/copilot-instructions.md`, etc.) only for AI agents requested in target config — **in migration-mode form**: `.claude/settings.json` hooks that could block edits are commented out or set to advisory
- Generates module skeletons (Model, SQLDatastore, Shared, API, Worker, EventConsumer, AIAgent) **only for modules in target config**
- Generates module Dockerfiles only for app modules selected
- Generates `docker-compose.yml` only if dev-time services are required (database/broker selected)
- Wraps existing source as `legacy/pom.xml` + moves files; runs `mvn verify` at root (legacy + empty new modules) to verify everything compiles and the user's pre-existing tests in legacy still pass
- Does **not** generate CI workflows — that is the deployment specialist's job in Phase 10 if and only if the legacy already has CI/CD

### 7.3 `trabuco-migration-model-specialist` (Phase 2)
**Scope.** Migrate entities, DTOs, events, job-request types from `legacy/` to the Model module. Apply Immutables (`@Value.Immutable`), Trabuco naming conventions, sub-package layout (entities/, dtos/, events/, jobs/).
**Per-aggregate vertical slice.** One entity at a time: write the Model-module equivalent, mark legacy as `@Deprecated` with adapter, run characterization tests via test specialist, run `mvn compile + tests`.
**Decisions surfaced.** Naming when source naming doesn't match Trabuco conventions. Sub-package placement when ambiguous.

### 7.4 `trabuco-migration-datastore-specialist` (Phase 3)
**Scope.** Migrate persistence (JdbcTemplate, EntityManager, JPA repos, MongoTemplate) to Trabuco's Spring Data JDBC + Flyway + keyset pagination + HikariCP, dropping FK constraints.
**Adjacency adapter.** During transition, Trabuco repositories may delegate to legacy data sources via adapter classes; cutover happens once tests pass on the new implementation.
**Decisions surfaced.** FK retention vs application-level integrity check. Keyset cursor column selection per aggregate. Liquibase → Flyway translation vs retention.

### 7.5 `trabuco-migration-shared-specialist` (Phase 4)
**Scope.** Migrate `@Service` and business logic to the Shared module. Constructor injection (replacing field/setter injection). Resilience4j boundaries where the legacy used circuit-breaker-like patterns. Add ArchUnit tests enforcing module boundaries — **tagged `@Tag("trabuco-arch")` so they are excluded by Surefire during migration**. The activation phase (Phase 12) removes the exclusion.
**Decisions surfaced.** Splitting stateful service classes. Cross-domain service ownership.

### 7.6 `trabuco-migration-api-specialist` (Phase 5, conditional)
**Scope.** Migrate `@RestController` / `@Controller` to the API module. RFC 7807 ProblemDetail (replacing bespoke error envelopes). Bean Validation. SpringDoc OpenAPI. Bucket4j (only if rate limiting was already present). Security headers filter, CorrelationIdFilter. Virtual threads enabled in `application.yml`. Keyset pagination on collection endpoints.
**Decisions surfaced.** Wire-format compatibility (legacy clients depending on existing error format). Endpoint naming when source uses non-Trabuco conventions.

### 7.7 `trabuco-migration-worker-specialist` (Phase 6, conditional)
**Scope.** Migrate `@Scheduled` / `@Async` / queue listeners that were doing async work to JobRunr handlers + RecurringJobsConfig.
**Decisions surfaced.** Existing job retry/concurrency settings → JobRunr equivalents.

### 7.8 `trabuco-migration-events-specialist` (Phase 7, conditional)
**Scope.** Migrate `@KafkaListener` / `@RabbitListener` / SQS / PubSub listeners to the EventConsumer module + Events publisher pattern. Idempotent processing.
**Decisions surfaced.** Existing consumer-group / queue topology preservation.

### 7.9 `trabuco-migration-aiagent-specialist` (Phase 8, conditional)
**Scope.** If the legacy has LLM/agent integration, migrate to Trabuco's AIAgent module: Spring AI 1.0.5, MCP server, A2A protocol, tools, knowledge, guardrails. Most migrations will skip this phase.
**Decisions surfaced.** Tool taxonomy. Knowledge schema mapping.

### 7.10 `trabuco-migration-config-specialist` (Phase 9)
**Scope.** Author `application.yml` per module per profile from the legacy property sources. Add OpenTelemetry blocks. Author `logback-spring.xml` for structured logging. Establish env var conventions.
**Decisions surfaced.** Profile mapping (legacy `dev/test/prod` → Trabuco profiles). External property source preservation (Vault, Consul, etc.).

### 7.11 `trabuco-migration-deployment-specialist` (Phase 10, conditional)

**Scope.** Adapt the legacy CI/CD pipeline files to fit the new multi-module Trabuco structure. **The specialist does not invent deployments. It does not generate workflows that the legacy did not have.** It only updates the existing pipeline files (or `not-applicable` if there are none).

**Scope of adaptation per legacy CI/CD file:**
- Update Maven build commands to reference the multi-module structure (`mvn -pl :api package` instead of single-module `mvn package` where appropriate)
- Update Docker image build paths to the new module locations (`api/Dockerfile` instead of root `Dockerfile`)
- Update test command invocations to point at the new module test directories
- **Preserve verbatim**: triggers (push to main, PR, tag, schedule), env vars, secrets references, deploy targets (staging, prod, etc.), runner labels (self-hosted, ubuntu-latest), step ordering, conditional logic, branch protection assumptions, deploy-on-merge patterns, manual approval steps, environment-protection rules

**Hard rules (architecturally enforced):**
- The specialist may not introduce new pipeline stages, jobs, or steps
- The specialist may not change pipeline providers (Jenkins → GitHub Actions is forbidden even if "more modern")
- The specialist may not add observability, monitoring, alerting, canary, blue/green, GitOps, Helm chart, k8s manifest, or any other infrastructure-as-code that wasn't in the legacy
- The specialist may not add security scanning, lint, format, or any other check that wasn't there
- If the legacy has multiple deploy targets (staging + prod, multi-region), all are preserved as-is
- If the legacy uses a specific Java version in CI different from Trabuco's target Java version, this is flagged as `JAVA_VERSION_MISMATCH_CI` requiring user decision (keep CI version or unify with Trabuco target)

**Files in scope** (only those present in legacy):
- `.github/workflows/*.yml`
- `.gitlab-ci.yml`
- `Jenkinsfile`, `Jenkinsfile.*`
- `.circleci/config.yml`
- `azure-pipelines.yml`
- `.travis.yml`
- `argocd/`, `argo-cd/`, `flux/`
- `charts/`, `helm/`
- `k8s/`, `kubernetes/`, `manifests/`
- `terraform/`, `pulumi/`
- `Dockerfile.*`, `docker-compose.prod.yml`, `compose.prod.yml`
- Heroku `app.json`, fly.io `fly.toml`, Render `render.yaml`
- AWS CDK / CloudFormation / SAM templates

**Edge cases:**
- Legacy CI references a script in `scripts/` that has hardcoded paths — flagged for user review (specialist does not modify untracked scripts)
- Legacy CI builds Docker images with a single root `Dockerfile` but Trabuco target has multi-module Dockerfiles — flagged as `DOCKERFILE_GRANULARITY_CHANGE` requiring user decision
- Legacy CI deploys a single jar but Trabuco target has multiple deployable apps — flagged as `DEPLOYMENT_TOPOLOGY_CHANGE` requiring user decision
- Legacy CI has a "release" job that publishes to a registry — preserved verbatim with paths updated
- Legacy CI has manual approval gates — preserved verbatim
- Legacy CI uses self-hosted runners with specific labels — preserved verbatim

**Output.** Patches to existing CI/CD files only. Never new files. If the legacy has no CI/CD whatsoever, the specialist outputs `not-applicable` and Phase 10 is skipped entirely. The user can manually add CI/CD later — that is explicitly out of migration scope.

### 7.12 `trabuco-migration-test-specialist` (Phase 11 + cross-cutting)

**Scope (cross-cutting at every phase).** Before any specialist transforms code in a phase, the test specialist captures **characterization tests** (Feathers) over the existing legacy behavior. These tests pin observable behavior so transformation can be verified.

**Scope (Phase 11, dedicated).** Per-test analysis of every test in the source repo, with one of four decisions:

- **KEEP** — the test is still valid in the new structure, perhaps with import/annotation adjustments. Test specialist applies the minimal adjustments needed.
- **ADAPT** — the test is fundamentally about the right thing but needs significant rewriting (legacy `@SpringBootTest` → sliced `@WebMvcTest`/`@DataJdbcTest`; legacy MockMvc → Testcontainers integration). Test specialist proposes the adapted version with a side-by-side comparison.
- **DISCARD** — the test was testing legacy behavior that no longer exists in the new structure (e.g., a test for an FK constraint we dropped, or a test for a bespoke error envelope we replaced with ProblemDetail). Test specialist proposes deletion with explicit justification citing what no longer exists.
- **CHARACTERIZE-FIRST** — there's no test for a piece of legacy behavior we are about to migrate. Test specialist authors a characterization test against legacy *before* the transformation, runs it, and uses it as the verification basis.

Each per-test decision is presented for user approval. Default gate is per-phase batch; `--per-test` flag presents each individually for high-stakes migrations.

**Decisions surfaced.** Every ADAPT and DISCARD requires user approval. KEEP and CHARACTERIZE-FIRST can be auto-approved unless `--per-test` is set.

### 7.13 `trabuco-migration-activator` (Phase 12) — NEW

**Scope.** Flip the deferred enforcement mechanisms from off to on, format any code that needs formatting first, and verify the project still passes with full enforcement. This specialist exists because Trabuco's standard generation includes mechanisms (Maven Enforcer, Spotless, ArchUnit, Jacoco threshold) that would break the build if enabled mid-migration; they are deliberately deferred until now.

**Behavior, in order:**
1. Run `mvn spotless:apply` to format all code in the workspace (so the subsequent `spotless:check` passes)
2. In the parent POM: remove `<skip>true</skip>` from `maven-enforcer-plugin` (or add the plugin if it was omitted) with full Trabuco-standard rule set (`bannedDependencies`, `dependencyConvergence`, `requireMavenVersion`, `requireJavaVersion`)
3. Configure Maven Enforcer to skip the `legacy/` module if it still exists (so retained legacy code doesn't trigger enforcer)
4. In the parent POM: remove `<skip>true</skip>` from `spotless-maven-plugin`
5. In the parent POM: add Jacoco's threshold-enforcing `check` execution
6. Remove the `trabuco-arch` tag exclusion from Surefire so ArchUnit tests run
7. Swap `.claude/settings.json` from migration-mode → production-mode (enable hooks)
8. **Optional CI step** — only if the legacy CI already had a "lint" or "format" stage, add `mvn spotless:check` and `mvn verify` (with full enforcement) to the equivalent step. **Never adds CI stages that the legacy didn't have.** This is the same no-out-of-scope rule as Phase 10.
9. Run full `mvn verify` end-to-end with all enforcement on
10. If failures surface, classifies them by reason code (`COMPILE_FAILED`, `ENFORCER_VIOLATION`, `SPOTLESS_VIOLATION`, `ARCHUNIT_VIOLATED`, `COVERAGE_BELOW_THRESHOLD`) and presents them with concrete remediation. User can fix-and-retry, accept-with-caveats (skip a single rule), or reject the activation
11. Approval gate: user reviews the activation diff and confirms

**Why a separate phase, not folded into finalizer.** Activation has user-visible decisions (which enforcer rules to keep, whether to add a CI lint step) that finalizer's "verify and clean up" scope shouldn't be making. Separation keeps each gate's responsibility clear.

**Tools allowed.** Edit, Read, Bash (`mvn` family).

### 7.14 `trabuco-migration-finalizer` (Phase 13)
**Scope.** Verify all migration phases completed and the activation phase succeeded. Remove the residual `legacy/` module if empty (or retain it with `@Deprecated` markers if user opted to preserve unmigrated artifacts). Run `trabuco doctor --fix`. Run `trabuco sync`. Run final `mvn verify`. Generate the migration completion report.
**Output.** `.trabuco-migration/completion-report.md` summarizing: what was migrated, what was retained as legacy, what blockers remain, what decisions were made, what tests were kept/adapted/discarded, what activation rules are in force.

---

## 8. Blocker reason codes (fixed enum)

Specialists must classify every blocker into this enum. New codes require an explicit code change. The enum is the contract between specialists and the orchestrator.

**Schema / data model**
- `FK_REQUIRED` — schema depends on FK cascade or referential integrity
- `OFFSET_PAGINATION_INCOMPATIBLE` — API uses offset pagination, no natural keyset ordering exists
- `STATEFUL_DTO` — DTO has setters or mutable fields used by external callers
- `COMPOSITE_PK_NO_NATURAL_ORDER`
- `MUTABLE_ENTITY_GRAPH` — JPA entity-graph traversal Spring Data JDBC can't replicate
- `EMBEDDED_DB_DIALECT` — H2-specific SQL that won't run on Postgres Testcontainer

**Coupling / runtime**
- `STATIC_GLOBAL_STATE`
- `APPCONTEXT_LOOKUP`
- `SERVICELOADER`
- `FIELD_INJECTION_COMPLEX` — circular dependencies that resist constructor-injection rewrite
- `THREADLOCAL_LIFECYCLE` — manual ThreadLocal lifecycle that breaks under virtual threads
- `NON_VIRTUAL_THREAD_SAFE` — `synchronized` blocks around I/O that pin virtual threads
- `BLOCKING_REACTIVE_MIX` — WebFlux + blocking via shared state

**Build / framework**
- `GRADLE_PARENT_AS_ARTIFACT`
- `BUILD_PLUGIN_NOT_PORTABLE`
- `JAVA_VERSION_INCOMPATIBLE`
- `NON_JAKARTA_DEP_NO_REPLACEMENT`
- `NON_SPRING_FRAMEWORK` — Quarkus, Micronaut, Helidon, etc.

**Wire format / contract**
- `LEGACY_ERROR_FORMAT_REQUIRED`
- `BESPOKE_AUTH_PROTOCOL`
- `BINARY_PROTOCOL`

**Tests**
- `POWERMOCK_LEGACY`
- `MISSING_CHARACTERIZATION_BASIS`
- `BROAD_TEST_SUITE_SLOW`
- `SPOCK_TESTS`

**Repo shape**
- `NON_JVM_CODE_SUBSTANTIAL`
- `MULTI_LANGUAGE_BUILD`
- `KOTLIN_PARTIAL`
- `SECRET_IN_SOURCE` — hardcoded credentials (mandatory blocker; no override)

**Deployment / CI-CD** (new for the deployment specialist)
- `DOCKERFILE_GRANULARITY_CHANGE` — single-Dockerfile legacy vs multi-module Trabuco
- `DEPLOYMENT_TOPOLOGY_CHANGE` — single-app legacy vs multi-app Trabuco
- `JAVA_VERSION_MISMATCH_CI` — CI builds with a Java version different from Trabuco target
- `EXTERNAL_SCRIPT_REFERENCED` — CI references a script with hardcoded paths the specialist won't modify
- `DEPLOY_TARGET_UNRESOLVABLE` — CI references a deploy target whose new equivalent can't be inferred

**Validation funnel failures**
- `COMPILE_FAILED`
- `ARCHUNIT_VIOLATED`
- `TESTS_REGRESSED`
- `EVIDENCE_INVALID` — source_evidence in the patch doesn't match actual source (out-of-scope guard)

**Activation-phase failures** (only surfaced during Phase 12)
- `ENFORCER_VIOLATION` — Maven Enforcer caught a banned dependency or version conflict that survived migration
- `SPOTLESS_VIOLATION` — code didn't format cleanly even after `spotless:apply` (rare, indicates a Spotless rule we didn't account for)
- `COVERAGE_BELOW_THRESHOLD` — Jacoco threshold not met after migration; user can lower threshold or add tests

For each blocker, the specialist must propose at least one alternative. If the user refuses every alternative, the artifact is marked `retained-legacy` (stays in the `legacy/` module forever) or the phase is aborted, depending on user choice.

---

## 9. Edge case catalog

**Source codebase shapes**
- Single-module Maven → default path
- Multi-module Maven → adjacent strategy: keep their modules as legacy, add Trabuco modules alongside, migrate one-by-one (user decision ratified)
- Gradle → assessor flags + asks user whether to convert build system before migrating, or migrate as-is leaving Gradle in `legacy/`. Conversion of Gradle to Maven is out of scope for Phase 1; if the user wants it, they do it before migration starts.
- Spring Boot 2.x → migration includes implicit Spring Boot 3 + jakarta upgrade as part of code transformation (LLM-driven, not OpenRewrite)
- Quarkus / Micronaut / Helidon → flagged `NON_SPRING_FRAMEWORK`; user warned heavy effort; partial migration only
- Plain JAX-RS / Servlet → flagged `NON_SPRING_FRAMEWORK`; translatable but expensive
- Mixed Java + Kotlin → Kotlin retained in legacy; only Java migrated; flag `KOTLIN_PARTIAL`
- Backend + frontend monorepo → frontend retained untouched; only Java migrated

**Configuration shapes**
- `application.properties` → translated to `application.yml`
- Multiple Spring profiles → preserved per-profile
- External property sources (Vault, Consul) → mechanism preserved; flagged for user verification
- Hardcoded credentials → `SECRET_IN_SOURCE` mandatory blocker

**Database / persistence shapes**
- Existing Flyway → Trabuco continues from highest version
- Liquibase → user choice: translate or retain
- `ddl-auto=update` → mandatory blocker; explicit migrations required first
- No persistence → SQLDatastore not generated
- Sharded / multi-tenant → flagged `MULTI_TENANT_COMPLEX`

**Test shapes**
- H2 embedded → migrate to Postgres Testcontainer; SQL dialect mismatches flagged
- `@SpringBootTest` for everything → split where feasible (test specialist's per-test analysis)
- PowerMock → `POWERMOCK_LEGACY` blocker
- Spock → retained as-is, flagged

**Repo state preconditions** (orchestrator hard gate)
- Uncommitted changes → require commit/stash
- Detached HEAD → require branch
- No git → require `git init`
- Submodules → flag, decision required
- Existing `.trabuco-migration/` → resume vs fresh-start prompt
- Remote diverged → warn, offer pull/rebase

**Failure recovery**
- Specialist crash mid-phase → state.json marks `in_progress`; orchestrator offers resume or rollback
- Build fails after phase apply → automatic rollback to phase pre-tag, structured failure report
- User rejects 3 times → escalate to "manual phase" with structured report; no further auto-transformation

**Deployment shapes** (specific to Phase 10)
- No CI/CD at all → `not-applicable`, phase skipped
- Multiple CI providers (legacy migration in progress) → all retained, paths updated in each
- CI references untracked scripts → flag `EXTERNAL_SCRIPT_REFERENCED`
- Single Dockerfile in legacy, multi-module target → `DOCKERFILE_GRANULARITY_CHANGE` decision
- Deploy targets that no longer make sense (deploy-monolith → deploy-services) → `DEPLOYMENT_TOPOLOGY_CHANGE` decision

---

## 10. Validation funnel

Every LLM-generated change passes through this funnel before the user sees it. The funnel is the verification mechanism that backstops LLM-only transformation.

```
1. Lex/parse           ← `javac --no-output` or tree-sitter check on changed files
2. Source-evidence     ← orchestrator validates source_evidence field against actual source
3. Compile             ← `mvn compile -pl :{affected-modules}`
4. ArchUnit            ← module-boundary tests must pass (DEFERRED — see below)
5. Unit tests          ← pinned characterization tests + new tests pass
6. Integration tests   ← Testcontainers-backed slice tests pass
[7 — optional]        ← behavioral equivalence (Scientist-style) vs legacy (off by default)
```

**Steps 1–3 and 5–6 run inside the orchestrator before showing the user.** Failures are not surfaced as blockers — they are auto-fed back to the originating specialist with the failure trace for retry. Three retry attempts then escalate to `requires-decision`.

**Step 4 (ArchUnit) is deferred during migration phases** (Phases 1–11). ArchUnit tests are written into the Shared module by Phase 4 but tagged `@Tag("trabuco-arch")` and excluded from Surefire. The funnel does not run ArchUnit during migration phases because the modules are still being assembled and would violate the module-boundary tests by definition. The activation phase (Phase 12) is the only place ArchUnit is fully exercised — that's its purpose.

**Maven Enforcer and Spotless are similarly deferred.** They are not in the funnel during migration phases because they are skipped in the migration-mode parent POM (see §5). The activation phase is the gate where they run for the first time.

**Step 7** is opt-in via `--behavioral-equivalence` flag. It runs old and new on the same input set and asserts equivalence. Useful for high-risk phases (datastore cutover) but expensive.

The funnel is **not transformation**. It does not modify code. It only accepts or rejects LLM-produced patches.

**The activation phase is the funnel's "full-strength" run.** Before activation, the funnel verifies the project compiles and tests pass with enforcement off. After activation, the funnel re-runs with enforcement on (Spotless, Maven Enforcer, ArchUnit, Jacoco threshold all active). This is the moment of truth: the project must pass its own enforcement rules. Any failures route to the activator's user-visible decision flow rather than auto-retry.

---

## 11. State management & rollback

**`.trabuco-migration/` directory in the user's repo:**
```
.trabuco-migration/
├── state.json              # current phase, approvals log, blockers, decisions
├── assessment.json         # Phase 0 output — the contract for all later phases
├── lock.json               # prevents concurrent runs
├── phase-{N}-input.json    # input given to specialist for phase N
├── phase-{N}-output.json   # specialist's structured output
├── phase-{N}-diff.patch    # git-format patch for phase N
├── phase-{N}-report.md     # human-readable phase summary
├── phase-{N}-blockers.json # blockers surfaced in phase N
├── decisions/
│   └── {decision_id}.json  # user's recorded choice on requires-decision items
└── completion-report.md    # written by finalizer at the end
```

`.trabuco-migration/` should be added to the migrated project's `.gitignore`. It is local working state, not committed alongside the migration result.

**`state.json` schema (v1):**
```json
{
  "schemaVersion": 1,
  "trabucoCliVersion": "1.10.0",
  "startedAt": "2026-04-26T...",
  "lastUpdatedAt": "2026-04-26T...",
  "sourceConfig": {
    "buildSystem": "maven",
    "framework": "spring-boot-2.7",
    "javaVersion": "11",
    "persistence": "spring-data-jpa",
    "messaging": "kafka",
    "ciSystems": ["github-actions"]
  },
  "targetConfig": {
    "modules": ["Model", "SQLDatastore", "Shared", "API", "Worker"],
    "database": "postgresql",
    "messageBroker": "kafka",
    "aiAgents": ["claude"],
    "ciProvider": "github",
    "javaVersion": "21"
  },
  "phases": {
    "0": { "state": "completed", "approvedAt": "...", "preTag": "...", "postTag": "..." },
    "1": { "state": "completed", "approvedAt": "...", "preTag": "...", "postTag": "..." },
    "2": { "state": "in_progress", "preTag": "...", "subAggregates": { "User": "completed", "Order": "in_progress" } },
    "10": { "state": "not-applicable", "reason": "no CI/CD detected in source" }
  },
  "blockers": [
    { "phase": 3, "code": "FK_REQUIRED", "file": "...", "alternatives": [...], "userChoice": "drop-fk-app-level-check" }
  ],
  "decisions": [
    { "id": "d-001", "phase": 2, "question": "...", "choice": "...", "decidedAt": "..." }
  ],
  "retainedLegacy": [ "src/main/java/com/old/SpecialClass.java" ]
}
```

**Per-phase git tags** (`trabuco-migration-phase-{N}-pre`, `trabuco-migration-phase-{N}-post`) provide atomic rollback boundaries.

**Rollback semantics:**
- `migrate rollback --to-phase=N` → `git reset --hard trabuco-migration-phase-{N}-pre`, updates state.json, removes phases (N+1...current), keeps assessment.json
- Soft rollback (rebuild diff but don't apply) → write a clean patch file; user can hand-edit before reapplying

**Resume semantics:**
- `migrate resume` → reads state.json, identifies the phase marked `in_progress`, picks up from the most recent checkpoint within that phase

---

## 12. CLI ↔ plugin parity

The migration is exposed two ways. Both share Go-side handlers, state directory, specialist prompts, and validation gates.

**Plugin mode (Claude Code, Cursor, etc.):**
- User invokes `/trabuco:migrate` skill
- Skill activates `trabuco-migration-orchestrator` subagent
- Orchestrator subagent calls MCP tools (`migrate_assess`, `migrate_skeleton`, `migrate_module`, etc.) for state operations and specialist invocations
- Approval gates surface as natural-language messages from the orchestrator agent
- Diffs are presented in chat grouped by Trabuco module
- User approves/edits/rejects via natural language; orchestrator parses intent and acts

**CLI mode (`trabuco migrate <repo-path>`):**
- User runs `trabuco migrate /path/to/repo` from terminal
- A Go state machine acts as the orchestrator
- Same MCP tool handlers are invoked (now as Go function calls instead of MCP RPC)
- Specialist work is dispatched via the Anthropic SDK using the user's authenticated API key (`trabuco auth status` must pass)
- Approval gates surface as interactive terminal prompts (with full diff display, action menu)
- All state is identical to plugin mode — `.trabuco-migration/` is identical

**Shared between modes:**
- Go `internal/migration/` package (renamed from `internal/migrate/`, fresh implementation)
- MCP tool handlers in `internal/mcp/migration_tools.go`
- Specialist prompts in `internal/migration/prompts/`
- Validation funnel in `internal/migration/validation/`
- State management in `internal/migration/state/`
- Git tag and patch helpers in `internal/migration/vcs/`

**Parity invariant.** A migration started in CLI mode can be resumed in plugin mode and vice versa. State is the source of truth; the UX is just two ways to drive the same state.

---

## 13. MCP tool surface

| Tool | Purpose |
|---|---|
| `migrate_assess` | Run Phase 0; produces assessment.json |
| `migrate_skeleton` | Run Phase 1; generates target scaffolding (migration mode, enforcement deferred) + wraps source in `legacy/` |
| `migrate_module` | Run a module phase; takes `module: model|sqldatastore|nosqldatastore|shared|api|worker|eventconsumer|aiagent` |
| `migrate_config` | Run Phase 9 |
| `migrate_deployment` | Run Phase 10 (path adaptations only, no enforcement) |
| `migrate_tests` | Run Phase 11 |
| `migrate_activate` | Run Phase 12; flips deferred enforcement on, runs full `mvn verify` |
| `migrate_finalize` | Run Phase 13 |
| `migrate_status` | Returns current state.json |
| `migrate_rollback` | Reverts to a phase pre-tag |
| `migrate_decision` | Records user's choice on a `requires-decision` item |
| `migrate_resume` | Resumes from last checkpoint after crash/exit |

Each tool's handler is in Go, called by both MCP server and CLI command. None of the tools run longer than ~60s of wall time — long-running specialist work is async with progress polled via `migrate_status`.

---

## 14. Deletion checklist (old migrate feature)

Audit shows the old feature is **5,977 lines of Go alone**, plus extensive plugin and documentation references. Deletion happens **before** new feature work begins (user decision ratified).

**Delete entirely (files):**
- `internal/migrate/` — entire directory, 15 Go files
- `internal/cli/migrate.go`
- `internal/prompts/migrate_prompts.go`
- `internal/prompts/migrate_test.go`
- `plugin/skills/migrate/SKILL.md`
- `plugin/agents/trabuco-migration-expert.md`
- `plugin/hooks/post-migrate-project.sh`

**Modify (remove migrate references):**
- `internal/cli/root.go` — drop migrate cobra registration
- `internal/mcp/tools.go` — drop `registerMigrateProject` AND `registerScanProject` (user decision: remove scan_project entirely)
- `plugin/hooks/hooks.json` — drop the `mcp__trabuco__migrate_project` matcher
- `plugin/.mcp.json` — update tool count and description
- `.claude-plugin/marketplace.json` — update plugin description
- `plugin/.claude-plugin/plugin.json` — update description
- `README.md` — strip the entire AI-Powered Migration section (lines ~242–331); strip migrate references at lines 22, 35, 82, 176, 179, 672
- `internal/cli/auth.go` line 27 — update help text
- `internal/cli/mcp.go` line 38 — update help text
- `docs/CDK_EXTENSION_DESIGN.md` line 1716 — remove
- `plugin/docs/when-not-to-use.md` line 53 — remove migrate cost-warning
- `plugin/skills/new-project/SKILL.md` — drop fallback reference to scan_project
- `plugin/agents/trabuco-architect.md` — drop scan_project reference
- `plugin/skills/extend/SKILL.md` — drop scan_project reference
- `plugin/skills/add-module/SKILL.md` — drop scan_project reference

**`scan_project` removal — second-order impact.**
The skills `extend`, `add-module`, `new-project`, and the `trabuco-architect` agent currently reference `scan_project` as a fallback "look at this project" mechanism. Removing `scan_project` means these references must be replaced. **Resolution:** these skills/agents should rely on the existing `get_project_info` tool instead (which already returns `.trabuco.json` data), with `Read`/`Glob`/`Grep` for ad-hoc inspection. No new tool is added to fill the gap. If a real need surfaces in practice, we add a purpose-built tool in a later iteration.

**Version bump:** `1.9.4` → `1.10.0`. The deletion + redesign ship as one release (user decision ratified).

---

## 15. Release sequencing — single 1.10.0

The work ships as one cohesive release. Internally we develop in milestones for engineering discipline, but these are **not user-facing**. The user sees one release: 1.10.0 with the migration redesign.

**Internal development milestones:**

| Milestone | Scope |
|---|---|
| **M0 — Deletion** | Remove all old migrate feature artifacts. Add a temporary CLI message "migration is being redesigned" if `trabuco migrate` is invoked. Verify `mvn build` and `go test ./...` pass. |
| **M1 — Foundations** | New `internal/migration/` package skeleton. State.json schema + serializers. Git tag/patch helpers. CLI command stub. MCP tool stubs. Validation funnel framework (no specialists yet). |
| **M2 — Phase 0 (Assessor)** | Assessor specialist + `migrate_assess` tool. Assessment artifact schema. Pre-flight precondition checks. End-to-end: user can run `trabuco migrate assess /path` and get the assessment. |
| **M3 — Phase 1 (Skeleton)** | Skeleton-builder specialist + `migrate_skeleton` tool. Adjacent-strategy `legacy/` wrapping. Target-config-driven scaffolding (only what user requested). End-to-end on a fixture project. |
| **M4 — Phases 2-3 (Model + Datastore)** | Model and Datastore specialists. Per-aggregate vertical slice. Validation funnel fully wired. Test specialist's characterization-test capture for these phases. |
| **M5 — Phases 4-7 (Shared, API, Worker, EventConsumer)** | Shared, API, Worker, Events specialists. Conditional skip when phase precondition isn't met. |
| **M6 — Phases 8-9 (AIAgent + Configuration)** | AIAgent specialist (mostly skipped in practice). Config specialist with profile mapping. |
| **M7 — Phase 10 (Deployment)** | Deployment specialist with strict no-out-of-scope rules. Per-CI-system handlers (GitHub Actions, GitLab CI, Jenkins, Circle, Azure, etc.). `not-applicable` happy path for no-CI sources. |
| **M8 — Phase 11 (Test analysis)** | Test specialist's per-test analysis (KEEP/ADAPT/DISCARD/CHARACTERIZE-FIRST). Cross-cutting characterization-test capture refined. |
| **M9 — Phase 12 (Activator)** | Activator specialist. Migration-mode → production-mode parent POM swap logic. `mvn spotless:apply` integration. Surefire tag exclusion removal. ArchUnit re-enable. `.claude/settings.json` mode swap. Activation-phase failure classification (`ENFORCER_VIOLATION`, `SPOTLESS_VIOLATION`, `COVERAGE_BELOW_THRESHOLD`). |
| **M10 — Phase 13 (Finalizer)** | Finalizer specialist. `trabuco doctor --fix` integration. `trabuco sync` integration. Completion report generation. End-to-end on three fixture projects (Spring Boot 2.7 monolith, Spring Boot 3 multi-module, Quarkus negative-case). Verification that legacy CI continues to work at every phase boundary up through activation. |
| **M11 — Documentation & release** | README rewrite. Plugin manifest update. CLI help text. Skill and subagent docs. Version bump to 1.10.0. Public release. |

**Fixture projects** for end-to-end integration testing across milestones:
1. **Simple-monolith.** Spring Boot 2.7, single Maven module, JPA + Postgres, `@RestController`, `@Service`, `@Scheduled`. GitHub Actions with single deploy. ~10k LOC. The default happy path.
2. **Mid-monolith.** Spring Boot 3.0, single Maven module, JPA + Postgres + Kafka listeners, JobRunr-like async. GitLab CI with two deploy targets. ~30k LOC. Tests the broker and worker specialists.
3. **Negative-case.** Quarkus 3, multi-module, JAX-RS endpoints. Jenkins pipeline. The migration should `requires-decision` cleanly and present blockers without crashing.

Real-world migrations on user repos start happening only after M10. Until then we're working on the controlled fixtures.

---

## 16. Open items for follow-up iterations

These are explicitly **out of scope** for 1.10.0 and noted for a future iteration:

- Multi-repo / monorepo splits (one repo per migration in 1.10.0)
- Gradle → Maven build-system conversion as a pre-phase (1.10.0 leaves Gradle in `legacy/`)
- Quarkus / Micronaut full migration support (1.10.0 surfaces as `NON_SPRING_FRAMEWORK` blocker)
- `scan_project` replacement tool, if the skills/agents demonstrate need
- Behavioral-equivalence Scientist-style validation (Step 7 of funnel) — opt-in only in 1.10.0
- AI Agent specialist for non-Claude legacy LLM frameworks (LangChain Java, LlamaIndex, etc.)
- Real-time progress streaming during specialist invocations (1.10.0 polls `migrate_status`)
- Migration report visualization (1.10.0 emits markdown only)

---

## Decisions ratified by user (2026-04-26)

1. **Multi-module source strategy**: adjacent (B). Existing modules retained as legacy; Trabuco modules added alongside; migrate one-by-one.
2. **`scan_project` MCP tool**: removed entirely; downstream skills updated to use `get_project_info` + Read/Glob/Grep.
3. **Deletion vs. parallel transition**: delete the old feature first, in the same release as the new one. No deprecation window.
4. **Transformation engine**: 100% LLM-driven. No OpenRewrite, no codemods, no AST recipes. Validation funnel is verification only.
5. **Speed vs. safety**: safe over fast. Default gate granularity is per-phase; per-aggregate enabled by default for high-risk phases (Datastore, API).
6. **Test handling**: per-test analysis with KEEP/ADAPT/DISCARD/CHARACTERIZE-FIRST decisions.
7. **Plan persistence**: this document. Living, iterated.
8. **Release**: single 1.10.0 holding everything.

Plus from this same iteration:
9. **CI/CD specialist exists** (Phase 10) and is constrained to transforming legacy CI/CD only — never inventing pipelines.
10. **CLI parity** with the plugin: `trabuco migrate` is a fully self-sufficient terminal flow that reaches the same handlers and specialists.

Plus from the dependency-aware-phasing iteration (2026-04-26, second pass):

11. **Dependency-aware phasing**: enforcement of conformance is enabled only after conformance is achieved. Maven Enforcer, Spotless, ArchUnit module tests, Jacoco threshold, and `.claude/settings.json` enforcement hooks are all deferred during migration phases (1–11) and turned on by a new dedicated **Phase 12 — Enforcement activation** before finalization. The skeleton-builder generates a "migration-mode" parent POM that is byte-for-byte identical to Trabuco's standard parent POM except that those mechanisms have `<skip>true</skip>`. Total phases now 14 (0–13).
12. **Legacy CI continuity**: at every phase boundary throughout migration, the legacy CI workflow continues to work. The validation funnel ensures `mvn verify` passes (with enforcement off) at every commit point, so a user who pushes mid-migration does not break their CI.

---

*End of plan v2. Iterations to follow as we move through development milestones.*
