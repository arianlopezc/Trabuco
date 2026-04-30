<p align="center">
  <pre align="center">
 ████████╗██████╗  █████╗ ██████╗ ██╗   ██╗ ██████╗ ██████╗
 ╚══██╔══╝██╔══██╗██╔══██╗██╔══██╗██║   ██║██╔════╝██╔═══██╗
    ██║   ██████╔╝███████║██████╔╝██║   ██║██║     ██║   ██║
    ██║   ██╔══██╗██╔══██║██╔══██╗██║   ██║██║     ██║   ██║
    ██║   ██║  ██║██║  ██║██████╔╝╚██████╔╝╚██████╗╚██████╔╝
    ╚═╝   ╚═╝  ╚═╝╚═╝  ╚═╝╚═════╝  ╚═════╝  ╚═════╝ ╚═════╝
  </pre>
</p>

<h3 align="center">Generate production-ready Java projects — with the AI context your coding agents need built in.</h3>

Starting a new Java project is slow — even with a coding agent helping. You set up Maven modules, wire databases and tests, copy boilerplate from that one project that "mostly works," and then spend the first hour explaining every architectural convention to Claude Code, Cursor, Copilot, or Codex from scratch. Hours pass before you write a single line of business logic. Trabuco exists because your time — and your agent's context window — is better spent building features, not fighting configuration or re-teaching conventions.

Trabuco is a command-line tool — and a [Claude Code plugin](#claude-code-plugin) — that generates both halves of a modern Java codebase: a complete, production-ready multi-module Maven project *and* the AI context that teaches coding agents how to work in it. Run `trabuco init` (or inside Claude Code, type `/trabuco:new-project` and describe what you need in plain English), answer a few prompts (or pass flags for automation), and you get a fully wired Spring Boot codebase alongside task-specific prompts, quality specifications, per-agent rule files, and workflow hooks already configured for Claude Code, Codex, Cursor, and GitHub Copilot. No templates to download, no manual setup, and no session spent bootstrapping your agent's understanding of the project.

The generated code is production-grade by default. Spring Boot 3.4 with Spring Data JDBC (no JPA surprises), Flyway migrations, Testcontainers for real integration tests, Resilience4j circuit breakers, Google Java Format enforced by Spotless, ArchUnit rules that fail the build on layer violations, correlation-ID tracing, Prometheus metrics, OpenAPI + Swagger UI, and a global exception handler with sanitized responses. PostgreSQL, MySQL, MongoDB, or Redis — all configured with Docker Compose. JobRunr for background jobs; Kafka, RabbitMQ, AWS SQS, or GCP Pub/Sub for event-driven processing. The modular layout — **Model**, **SQLDatastore** / **NoSQLDatastore**, **Shared**, **API**, **Worker**, **EventConsumer** — has clean compile-time boundaries so `API` physically cannot import `Worker` code. Every opinion is deliberate: keyset pagination, no foreign-key constraints, Immutables at module boundaries, constructor injection only, bulk-bounded writes.

Alongside the code, Trabuco lays down an AI collaboration layer that the major coding agents load natively. The `.ai/prompts/` directory ships task-specific guides (`add-entity`, `add-endpoint`, `add-service`, `add-event`, `add-job`, `add-tool`) plus `JAVA_CODE_QUALITY.md` — an authoritative specification covering architecture boundaries, exception handling, datastore performance (bulk I/O, keyset drain loops, denormalization), and testing standards. Per-agent rule files — `CLAUDE.md` for Claude Code, `AGENTS.md` for Codex, `.cursor/rules/java.mdc` for Cursor, `.github/instructions/java.instructions.md` for Copilot — wire those conventions into each tool's native discovery. Claude also gets `.claude/skills/` for commit, PR, and review workflows; Codex and Cursor get hooks; Copilot gets setup steps. Every architectural convention lives in two places: enforced by the generated code and explained to the agents that will extend it.

Trabuco goes further. The **AI Agent module** ships production scaffolding for building AI agents *themselves* on Spring AI: tool calling, LLM input/output guardrails, multi-agent orchestration, knowledge-base integration, MCP server endpoint, the A2A (Agent-to-Agent) protocol, streaming responses, webhook callbacks, scope-based authorization, rate limiting, and correlation-ID tracing — wired, testable, and ready from the first commit. Trabuco itself is also an MCP server: run `trabuco mcp` and your coding agents invoke scaffolding operations natively (`init_project`, `add_module`, `design_system`, `suggest_architecture`).

> **Migrating an existing Spring Boot project** is supported in **1.10** via `trabuco migrate`: an orchestrated 14-phase migration that transforms a Maven-based Spring Boot 2.x or 3.x repo in place into a Trabuco-shaped multi-module project. Each phase is owned by a specialist that proposes structured changes you approve at a gate before they're committed; per-phase git tags provide atomic rollback. Start with `trabuco migrate assess /path/to/repo`. Full guide: [`docs/migration-guide.md`](docs/migration-guide.md).

## Install

```bash
# Recommended: use via npx (no install needed)
claude mcp add trabuco -- npx -y trabuco-mcp

# Or install the CLI globally
npm install -g trabuco-mcp                                       # via npm
go install github.com/arianlopezc/Trabuco/cmd/trabuco@latest     # via Go
curl -sSL https://raw.githubusercontent.com/arianlopezc/Trabuco/main/scripts/install.sh | bash  # via shell
```

For per-agent MCP setup (Cursor, Copilot, Codex), offline tarballs, and binary fallback, see [`docs/manual.md#installation`](docs/manual.md#installation).

## Quick start

```bash
# Interactive — Trabuco asks what you need
trabuco init

# Or all flags
trabuco init --name=myapp --group-id=com.company.myapp \
  --modules=Model,SQLDatastore,Shared,API --database=postgresql

cd myapp
mvn clean install
cd API && mvn spring-boot:run
```

Your API runs at `http://localhost:8080`. For Worker, EventConsumer, AIAgent, and CI examples, see [`docs/manual.md#quick-start`](docs/manual.md#quick-start).

## Claude Code plugin

Trabuco is published in [Anthropic's community plugin marketplace](https://github.com/anthropics/claude-plugins-community). Inside Claude Code:

```text
/plugin marketplace add anthropics/claude-plugins-community
/plugin install trabuco@claude-community
```

You can also browse the catalog at [claude.com/plugins](https://claude.com/plugins/), or pin to this repo directly (e.g. for unreleased changes) with `/plugin marketplace add arianlopezc/Trabuco` then `/plugin install trabuco@trabuco-marketplace`.

The plugin ships **8 skills** (`/trabuco:new-project`, `/trabuco:design-system`, `/trabuco:add-module`, `/trabuco:extend`, `/trabuco:migrate`, `/trabuco:doctor`, `/trabuco:suggest`, `/trabuco:sync`), **17 specialist subagents** (architect, AI-agent expert, migration orchestrator, 14 migration phase specialists), and an **MCP server** with 25 tools. Full plugin docs: [`plugin/README.md`](plugin/README.md).

The plugin requires the `trabuco` CLI on PATH — install it first (or after; the SessionStart hook detects missing binaries and tells the assistant exactly what to do).

## Highlights

- **Multi-module Maven** — clean compile-time boundaries between Model, SQLDatastore/NoSQLDatastore, Shared, API, Worker, EventConsumer, AIAgent.
- **Spring Boot 3.4 + Java 21/25** — Spring Data JDBC (no JPA), Flyway migrations, virtual threads on by default, Testcontainers for real integration tests.
- **OIDC Resource Server scaffolding (auto-generated for API/AIAgent)** — Spring Security 6 dual `SecurityFilterChain`, JWT validation, scope-mapped authorities, RFC 7807 ProblemDetail handlers, RSA-signed e2e tests. **Ships dormant** — flip `trabuco.auth.enabled=true` and set `OIDC_ISSUER_URI` to validate tokens from Keycloak / Auth0 / Okta / Cognito / generic OIDC. No CLI flag, no half-installed projects. Full guide: [`docs/auth.md`](docs/auth.md).
- **Production observability** — RFC 7807 Problem Details, OpenTelemetry auto-instrumentation, Prometheus metrics, correlation IDs, health probes.
- **AI Agent module** — Spring AI with tools, LLM guardrails, MCP server, A2A protocol, multi-agent orchestration, knowledge base, webhooks. The OIDC chain coexists with the legacy `ApiKeyAuthFilter` (governed by an independent property) for incremental migration.
- **AI collaboration layer** — `.ai/prompts/` task guides, `JAVA_CODE_QUALITY.md` spec, per-agent rule files for Claude Code / Codex / Cursor / GitHub Copilot.
- **Migration of existing repos** — `trabuco migrate` runs a 14-phase orchestrated flow with per-phase approval gates and atomic rollback.
- **CI that knows your modules** — `--ci github` emits a workflow with the exact Docker services your modules need.
- **Stay in sync with new releases** — `trabuco sync --apply` adds AI-tooling files added in newer CLI versions, additively.

For the full feature list, see [`docs/manual.md#features`](docs/manual.md#features).

## Documentation

- **[`docs/manual.md`](docs/manual.md)** — full CLI reference (installation, modules, MCP server, code quality, configuration options, every flag).
- **[`docs/migration-guide.md`](docs/migration-guide.md)** — migrating an existing Spring Boot 2.x/3.x project.
- **[`docs/README.md`](docs/README.md)** — full docs index, including design context (philosophy, pattern recipes, when not to use, limitations).
- **[`plugin/README.md`](plugin/README.md)** — Claude Code plugin: skills, subagents, hooks, MCP server.

## Requirements

- **Java 21+** (21, 25, or 26 — Trabuco auto-detects installed versions)
- **Maven 3.8+**
- **Docker** (for Testcontainers and local development)

## Contributing

Issues and pull requests welcome.

## License

MIT — do whatever you want with it.

---

**Built with Trabuco.** Now go build something amazing.
