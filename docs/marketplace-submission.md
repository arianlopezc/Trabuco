# Submitting Trabuco to the Anthropic plugin marketplace

This file is what you paste / reference when submitting Trabuco to
Anthropic's official Claude Code plugin directory at
[`anthropics/claude-plugins-official`](https://github.com/anthropics/claude-plugins-official).

The submission goes through one of two in-app forms (you must be logged
in to whichever Anthropic surface you use most):

- **From Claude.ai**: <https://claude.ai/settings/plugins/submit>
- **From Console**:   <https://platform.claude.com/plugins/submit>

Both forms feed into the same review pipeline. Pick whichever you're
already authenticated to. The directory is curated; reviews are manual.

---

## 1. Pre-submission checklist

Run through this before opening either form. Everything below is true on
`main` at commit `88c151d` (or later) and tag `v1.10.1`.

### Repository signals

- [x] Repo is public: <https://github.com/arianlopezc/Trabuco>
- [x] `LICENSE` file at repo root (MIT)
- [x] `README.md` at repo root explains what the project does
- [x] At least one tagged release: `v1.10.1`
- [x] Release artifacts attached (binaries, plugin tarball, MCPB bundles, install.sh)
- [x] CI green on `main` (18/18 checks on the most recent push)

### Plugin structure

- [x] `plugin/.claude-plugin/plugin.json` with `name`, `version`,
      `description`, `author`, `homepage`, `repository`, `license`,
      `keywords`
- [x] `plugin/README.md` explaining what the plugin ships, how to
      install, prerequisites, compatibility
- [x] `plugin/.mcp.json` with the MCP server configuration
- [x] `plugin/agents/`, `plugin/skills/`, `plugin/hooks/` populated
- [x] All hooks executable (`chmod +x`) and tolerant of the CLI not
      being installed (no crashes; helpful message instead)

### Discoverability signals

- [x] Plugin registered with the MCP Registry
      (`io.github.arianlopezc/trabuco`, published by the release pipeline)
- [x] CLI published to npm (`@arianlopezc/trabuco-cli`)
- [x] Self-hosted marketplace works:
      `/plugin marketplace add arianlopezc/Trabuco` → `/plugin install trabuco@trabuco-marketplace`

### Quality signals

- [x] No stale documentation referencing pre-1.10 redesign or removed
      beta tools (`migrate_project`, `scan_project`)
- [x] No references to non-existent install paths (e.g. Homebrew formula)
- [x] Smoke-tested end-to-end on four fixtures
      (`testdata/migration-fixtures/spring-boot-27-{monolith,mongo,kafka,gradle}`)

---

## 2. The marketplace.json entry to propose

Anthropic's reviewers will add Trabuco to the `plugins` array of
[`anthropics/claude-plugins-official/.claude-plugin/marketplace.json`](https://github.com/anthropics/claude-plugins-official/blob/main/.claude-plugin/marketplace.json).
The entry should look like this — paste into the form's "what should the
listing contain" field, or reference this file:

```json
{
  "name": "trabuco",
  "description": "Production-ready Java/Spring Boot scaffolding and orchestrated migration of legacy projects for AI coding agents. Generates multi-module Maven projects with opinionated conventions (keyset pagination, constructor injection, Testcontainers, ArchUnit boundaries, RFC 7807) plus the AI collaboration layer that teaches Claude Code the project's architecture. Migrates existing Spring Boot 2.x/3.x repos in place via /trabuco:migrate — a 14-phase orchestrated flow with per-phase approval gates and atomic git-tag rollback.",
  "author": {
    "name": "Arian Lopez",
    "url": "https://github.com/arianlopezc"
  },
  "category": "development",
  "source": {
    "source": "git-subdir",
    "url": "https://github.com/arianlopezc/Trabuco.git",
    "path": "plugin",
    "ref": "v1.10.1"
  },
  "homepage": "https://github.com/arianlopezc/Trabuco",
  "keywords": [
    "java",
    "spring-boot",
    "maven",
    "scaffolding",
    "migration",
    "mcp"
  ]
}
```

Notes for the reviewer about the entry:

- **`source.source: "git-subdir"`** — same model Stripe uses. Trabuco's
  plugin lives under `plugin/` in the main monorepo (CLI source + plugin
  + tests share one repo). Cloning the whole repo for plugin install is
  unnecessary; `git-subdir` references just the plugin path.
- **`source.ref: "v1.10.1"`** — pinned at a release tag for the first
  submission so the marketplace ships exactly what was reviewed. Future
  releases (v1.10.2, v1.11, …) will require a one-line PR to bump
  `ref`. Happy to switch to `ref: "main"` (rolling release) if Anthropic
  prefers the Stripe model.
- **`category: "development"`** — same as `quarkus-agent`, `terraform`,
  `postman`, `sourcegraph`. Closest peer is `quarkus-agent` (Quarkus is
  another JVM framework with a similar scaffolding mandate).

---

## 3. Submission-form content

The exact field labels on the in-app form aren't documented publicly, so
the items below cover everything any reasonable plugin-directory form
would ask. Paste each into the matching field; if a field doesn't exist
on the form, drop the item.

### Plugin name

```
trabuco
```

### One-line tagline (≤ 100 chars)

```
Production-ready Java/Spring Boot scaffolding and 14-phase legacy-project migration for AI coding agents.
```

### Full description

```
Trabuco generates production-ready Java multi-module Maven projects and migrates existing Spring Boot 2.x/3.x repos into the same shape. It exposes both halves to Claude Code: 8 skills (/trabuco:new-project, /trabuco:design-system, /trabuco:add-module, /trabuco:extend, /trabuco:migrate, /trabuco:doctor, /trabuco:suggest, /trabuco:sync), 17 subagents (architect + AI-agent expert + a 15-agent migration team), and a 25-tool MCP server.

The generated projects are opinionated and production-grade by default: Spring Boot 3 with Spring Data JDBC (no JPA), Flyway migrations, Testcontainers, Resilience4j circuit breakers, Google Java Format enforced by Spotless, ArchUnit boundary tests, RFC 7807 Problem Details, virtual threads, OpenTelemetry tracing, and the modular layout (Model / SQLDatastore / NoSQLDatastore / Shared / API / Worker / EventConsumer / AIAgent) with compile-time boundaries. A first-class AI Agent module ships scaffolding for tool calling, guardrails, MCP server endpoints, A2A protocol, and knowledge-base integration on Spring AI.

The /trabuco:migrate skill drives a 14-phase orchestrated migration of legacy Spring Boot codebases: assessment → skeleton → model → datastore → shared → api → worker → eventconsumer → aiagent → config → deployment → tests → activation → finalization. Every phase is gated for approval, validated against compile + tests, and protected by per-phase git tags so any phase can be atomically rolled back. Maven-only at v1.10; Gradle is detected and surfaced as a clear blocker with conversion guidance.

Alongside the code, Trabuco lays down per-agent rule files (CLAUDE.md, AGENTS.md, .cursor/rules, .github/instructions) that teach the major coding agents the project's architecture so they can extend it correctly without re-explaining conventions every session.
```

### Category

```
development
```

### Keywords / tags

```
java, spring-boot, maven, scaffolding, migration, mcp, ai-agent, microservices
```

### Repository URL

```
https://github.com/arianlopezc/Trabuco
```

### Homepage

```
https://github.com/arianlopezc/Trabuco
```

### Plugin version (current)

```
1.10.1
```

### License

```
MIT
```

### Maintainer

```
Name:   Arian Lopez
GitHub: https://github.com/arianlopezc
Repo:   https://github.com/arianlopezc/Trabuco
```

### Install instructions for the listing

```
1. Install the trabuco CLI (the plugin's MCP server requires it):
   curl -sSL https://github.com/arianlopezc/Trabuco/releases/latest/download/install.sh | bash

2. From Claude Code, install the plugin:
   /plugin install trabuco@claude-plugins-official
```

### Prerequisites the listing should call out

- `trabuco` CLI on PATH (≥ 1.8.0; latest 1.10.1)
- For the migration feature specifically: a Maven-based Spring Boot
  2.x/3.x project, JDK matching the project's target Java version,
  clean working tree on a branch, and an Anthropic API key for the
  per-phase LLM calls

### Security / safety posture

The plugin itself contains no credentials, no telemetry, no network
calls. The MCP server is the `trabuco` CLI run with `mcp` subcommand;
the CLI is open-source Go, builds reproducibly from the public repo
through a public GitHub Actions workflow, and is signed-and-shipped via
GitHub Releases. The `/trabuco:migrate` skill is destructive — it
writes inside a user's repository — but every change is gated for user
approval, every phase tags pre-state for atomic rollback, and the
preflight refuses to run on a dirty working tree.

### Why this belongs in the directory

Java-on-Spring-Boot is one of the largest professional development
populations in the world; the directory currently has `quarkus-agent`
(MCP-only, focused on Dev Mode) but nothing that produces full
production-grade Spring Boot projects with the AI-collaboration layer
that lets coding agents extend them safely. Trabuco fills that gap and
does so opinionatedly enough that one CLI command yields a project that
passes its own enforcement (Spotless, ArchUnit, Enforcer, Jacoco
threshold).

The 14-phase migration is the unique offering: there's no other plugin
in the directory that drives a structured legacy-to-modern migration
with per-phase gates and rollback.

---

## 4. Step-by-step submission

1. Confirm the repository is at tag `v1.10.1` on `main`:
   ```
   git -C /path/to/Trabuco describe --tags --exact-match origin/main
   ```
   should print `v1.10.1`.

2. Open the submission form:
   - Claude.ai users: <https://claude.ai/settings/plugins/submit>
   - Console users:   <https://platform.claude.com/plugins/submit>

3. Fill the form using the content in §3 above. For any field the form
   asks but isn't covered above, link back to the plugin README:
   <https://github.com/arianlopezc/Trabuco/blob/main/plugin/README.md>

4. If the form lets you attach the proposed `marketplace.json` entry,
   paste the JSON from §2 verbatim.

5. Submit. The directory README says "External plugins must meet
   quality and security standards for approval" — there's no published
   SLA, but expect review on the order of days. While you wait, the
   self-hosted marketplace path (`/plugin marketplace add arianlopezc/Trabuco`)
   keeps working.

6. **If accepted**, Anthropic will open a PR against
   `anthropics/claude-plugins-official` adding the marketplace.json
   entry. Once merged, users can install with:
   ```
   /plugin install trabuco@claude-plugins-official
   ```
   You'll want to update `plugin/README.md` and the main repo `README.md`
   to mention the official-marketplace install path as the recommended
   route. Also consider wiring `plugin-hints` so the `trabuco` CLI itself
   suggests the installation when run inside Claude Code (see
   <https://code.claude.com/docs/en/plugin-hints>).

7. **If rejected or revisions are requested**, the form's response will
   include feedback. Common revision asks based on what reviewers tend
   to flag:
   - Tighten the description (the current one above is long for a
     listing summary; have a 200-char fallback ready)
   - Add a screenshot or short demo video (no current Trabuco asset for
     this — would need to record one)
   - Clarify the Anthropic-API-key cost model for the migration feature
     in the listing

---

## 5. Post-listing maintenance

When Trabuco cuts a new release (e.g. v1.10.2, v1.11.0):

1. The release pipeline already publishes the new artifacts and
   tarball.
2. Open a 1-line PR against `anthropics/claude-plugins-official`
   bumping `source.ref` from the old tag to the new one. Reference the
   release notes in the PR description.

If you switch to `source.ref: "main"` (rolling release) once the plugin
is well-trusted, this maintenance step goes away — every commit on
`main` becomes the live version.

---

## 6. Useful links

- Anthropic plugin docs: <https://code.claude.com/docs/en/plugins>
- Plugin reference (full schema): <https://code.claude.com/docs/en/plugins-reference>
- Submission docs section: <https://code.claude.com/docs/en/plugins#submit-your-plugin-to-the-official-marketplace>
- Plugin hints (recommend-from-CLI): <https://code.claude.com/docs/en/plugin-hints>
- Official marketplace repo: <https://github.com/anthropics/claude-plugins-official>
- Closest peer (Quarkus Agent): <https://github.com/quarkusio/quarkus-agent-mcp>
