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

> **Migration of legacy projects** is being redesigned for the upcoming **1.10.0** release. The previous beta `trabuco migrate` command and `migrate_project` / `scan_project` MCP tools have been removed. The new feature — orchestrator agent + specialized subagents, dependency-aware phasing, in-place transformation — is captured in [`docs/MIGRATION_REDESIGN_PLAN.md`](docs/MIGRATION_REDESIGN_PLAN.md).

## Table of Contents

- [Features](#features)
- [Installation](#installation)
  - [Claude Code Plugin](#claude-code-plugin)
- [Quick Start](#quick-start)
- [Managing Existing Projects](#managing-existing-projects)
  - [Project Health Check](#project-health-check)
  - [Adding Modules](#adding-modules)
  - [Syncing AI Tooling](#syncing-ai-tooling)
- [CLI MCP Server](#cli-mcp-server)
  - [Configuration](#configuration)
  - [Available Tools](#available-tools)
- [Generated Project Structure](#generated-project-structure)
- [Modules](#modules)
  - [Model](#model)
  - [SQLDatastore](#sqldatastore)
  - [NoSQLDatastore](#nosqldatastore)
  - [Shared](#shared)
  - [API](#api)
  - [Jobs](#jobs)
  - [Worker](#worker)
  - [Events](#events)
  - [EventConsumer](#eventconsumer)
  - [AI Agent](#ai-agent)
- [Code Quality & Architecture](#code-quality--architecture)
  - [Auto-Formatting](#auto-formatting)
  - [Architecture Tests](#architecture-tests)
  - [AI Task Prompts](#ai-task-prompts)
  - [Code Review Workflow](#code-review-workflow)
- [CI/CD](#cicd)
  - [GitHub Actions](#github-actions)
- [Observability](#observability)
  - [Metrics](#metrics)
  - [API Documentation](#api-documentation)
  - [Request Tracing](#request-tracing)
  - [Health Checks](#health-checks)
  - [Test Coverage](#test-coverage)
- [Configuration Options](#configuration-options)
  - [Available Modules](#available-modules)
  - [Java Version Detection](#java-version-detection)
  - [AI Coding Agents](#ai-coding-agents)
  - [CI/CD Provider](#cicd-provider)
- [Tech Stack](#tech-stack)
- [Local Development](#local-development)
- [Requirements](#requirements)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Claude Code plugin** — Drive scaffolding, extension, and architecture advice conversationally via `/trabuco:*` skills and specialist subagents. One install: `/plugin marketplace add arianlopezc/Trabuco`. See [Claude Code Plugin](#claude-code-plugin) below.
- **Multi-module Maven structure** — Clean separation between Model, Data, Services, API, Worker, and EventConsumer
- **Incremental module addition** — Start minimal and add modules as you need them with `trabuco add`
- **Project health checks** — Validate project structure and consistency with `trabuco doctor`
- **Immutables everywhere** — Type-safe, immutable DTOs and entities with builder pattern
- **Spring Boot 3.4** — Latest LTS with Spring Data JDBC (not JPA — no magic, no surprises)
- **SQL databases** — PostgreSQL/MySQL support with Flyway migrations out of the box
- **NoSQL databases** — MongoDB/Redis support with Spring Data repositories
- **Background jobs** — JobRunr for fire-and-forget, delayed, recurring, and batch jobs
- **Event-driven messaging** — Kafka, RabbitMQ, AWS SQS, or GCP Pub/Sub with type-safe event contracts
- **Testcontainers 2.x** — Real database tests that actually work with Docker Desktop
- **Circuit breakers** — Resilience4j configured and ready to use
- **Prometheus metrics** — Micrometer with `/actuator/prometheus` endpoint
- **API documentation** — OpenAPI 3.0 with Swagger UI at `/swagger-ui.html`
- **Correlation IDs** — Request tracing with `X-Correlation-ID` header
- **Health probes** — Kubernetes-ready readiness and liveness endpoints
- **Test coverage** — JaCoCo reports for code coverage
- **Docker Compose** — Local development stack included
- **IntelliJ run configs** — Just open and run
- **GitHub Actions CI** — Opt-in CI workflow that adapts to your modules with `--ci github`
- **Virtual threads on by default** — Project Loom enabled in Spring Boot 3.4 (`spring.threads.virtual.enabled=true`); I/O-heavy services scale concurrency without code changes
- **RFC 7807 Problem Details** — every error response is `application/problem+json` via Spring 6's `ProblemDetail`
- **Bucket4j rate limiting** — declarative per-IP / per-key limits via `application.yml`, ships off by default with sample policy
- **OpenTelemetry observability** — auto-instruments Spring MVC, JDBC, Kafka/RabbitMQ, JobRunr; traces export to stdout in dev, OTLP in prod
- **Code quality enforcement** — Google Java Format (Spotless), Maven Enforcer, and auto-formatting hooks
- **Architecture tests** — ArchUnit rules enforce constructor injection, layer boundaries, and no cyclic dependencies
- **AI Agent module** — Production-ready AI agent with Spring AI: tool calling, LLM guardrails, multi-agent orchestration, MCP server, A2A protocol, and knowledge base
- **AI-friendly** — Generates context files, coding rules, quality specs, and task prompts for Claude, Cursor, GitHub Copilot, and Codex
- **CLI MCP server** — `trabuco mcp` exposes all CLI functionality as structured tools for AI coding agents

## Installation

Pick any method below to install the CLI. If you also want the [Claude Code Plugin](#claude-code-plugin), install the CLI first — the plugin spawns `trabuco mcp` from your `PATH` and does not download the binary itself.

### npx (recommended for MCP use)

No installation needed — just reference `trabuco-mcp` in your AI agent config and npx handles everything:

```bash
claude mcp add trabuco -- npx -y trabuco-mcp
```

See [CLI MCP Server > Configuration](#configuration) for all agent configs.

### npm

```bash
npm install -g trabuco-mcp
```

Installs the MCP server wrapper globally. It downloads the correct Trabuco CLI binary for your platform on install.

### From GitHub

```bash
curl -sSL https://raw.githubusercontent.com/arianlopezc/Trabuco/main/scripts/install.sh | bash
```

### Using Go

```bash
go install github.com/arianlopezc/Trabuco/cmd/trabuco@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your PATH.

### Claude Code Plugin

If you use Claude Code, install the Trabuco plugin to drive the CLI conversationally — native Claude Code skills, specialist subagents, and grounded architecture advice instead of flag memorization or raw MCP tool names.

**The plugin requires the `trabuco` CLI.** It is metadata only — skills, subagents, hooks, and an MCP server declaration. It does NOT download, bundle, or install the `trabuco` binary. When Claude Code loads the plugin, it spawns `trabuco mcp` from your shell `PATH`. If the binary is missing, the MCP server fails to start and every `/trabuco:*` skill errors out.

**Install order:**

1. **Install the CLI** using any method above (`npx`, `npm`, `curl | bash`, or `go install`).
2. **Verify it's on PATH:**
   ```bash
   trabuco version
   ```
3. **Install the plugin:**
   ```
   /plugin marketplace add arianlopezc/Trabuco
   /plugin install trabuco@trabuco-marketplace
   ```
4. **Restart your Claude Code session.** MCP servers and hooks are loaded on session start; they don't activate mid-session.

**If you installed the plugin first, no harm done.** Install the CLI, then restart Claude Code. On the next SessionStart, the plugin's hook detects the binary and activates normally.

**The plugin will tell you if the CLI is missing.** Its SessionStart hook runs `command -v trabuco` on boot. If it fails, the hook injects a message into the assistant's context explaining what's missing and how to install it — the assistant can then walk you through installation if you try a `/trabuco:*` skill before the binary is present.

**What the plugin ships:**

| Layer | Contents |
|---|---|
| **Skills** (6) | `/trabuco:new-project`, `/trabuco:design-system`, `/trabuco:add-module`, `/trabuco:extend`, `/trabuco:doctor`, `/trabuco:suggest` |
| **Specialist subagents** (2) | `trabuco-architect` (architecture reasoning), `trabuco-ai-agent-expert` (AIAgent module deep dive) |
| **Grounding docs** | Trabuco philosophy, module catalog interpretation, pattern recipes, limitations, when-not-to-use — so the assistant won't recommend what Trabuco can't deliver |
| **Hooks** | Session-start binary detection, post-tool-use next-steps printers for `init_project` and `generate_workspace` |
| **MCP server** | All 12 CLI tools + 4 expert prompts + 3 reference resources available natively inside Claude Code |

**Manual install from a release tarball** (offline or restricted environments) — each Trabuco release attaches `trabuco-plugin-vX.Y.Z.zip` as an asset. Download, extract, then:

```
/plugin marketplace add /path/to/extracted/trabuco-plugin-vX.Y.Z
/plugin install trabuco@trabuco-marketplace
```

**Troubleshooting — `/trabuco:*` skills return errors after install:**

- Run `trabuco version` in a terminal. If the command is not found, the binary is missing from `PATH`. Re-run one of the CLI install methods above.
- Confirm you restarted your Claude Code session AFTER installing both the CLI and the plugin. MCP servers only spawn at session start.
- If you installed via `go install`, make sure `~/go/bin` (macOS/Linux) or `%USERPROFILE%\go\bin` (Windows) is on your shell `PATH`.
- If you installed via `npm install -g trabuco-mcp`, make sure your npm global bin directory is on `PATH` (`npm prefix -g` shows the location).
- If you installed via `curl | bash`, the install script places the binary in `~/.local/bin` by default — ensure that's on `PATH`.

**For Cursor, Copilot, and Codex users:** the plugin is Claude-Code-specific. Use the CLI's built-in MCP server directly — see [CLI MCP Server](#cli-mcp-server) below.

## Quick Start

**Interactive mode** — just run `trabuco init` and answer a few questions.

**Non-interactive mode** — pass all options as flags:

```bash
# Basic API with PostgreSQL
trabuco init --name=myapp --group-id=com.company.myapp \
  --modules=Model,SQLDatastore,Shared,API --database=postgresql

# Add background jobs with Worker
trabuco init --name=myapp --group-id=com.company.myapp \
  --modules=Model,SQLDatastore,Shared,API,Worker --database=postgresql

# Add event-driven messaging with Kafka
trabuco init --name=myapp --group-id=com.company.myapp \
  --modules=Model,API,EventConsumer --message-broker=kafka

# Full setup with CI, AI agents, and all modules
trabuco init --name=myapp --group-id=com.company.myapp \
  --modules=Model,SQLDatastore,Shared,API,Worker,EventConsumer \
  --database=postgresql --message-broker=kafka --ai-agents=claude,cursor --ci github
```

### Run your new project

```bash
cd myapp
mvn clean install
cd API && mvn spring-boot:run
```

Your API is now running at `http://localhost:8080`. Try the health endpoint:

```bash
curl http://localhost:8080/health
```

## Managing Existing Projects

Trabuco isn't just for creating new projects — it can also validate and extend existing ones. The `doctor` command checks project health, and the `add` command lets you add modules incrementally as your needs evolve.

### Project Health Check

The `trabuco doctor` command validates that your project is healthy and consistent:

```bash
cd myapp
trabuco doctor
```

**Example output:**

```
Trabuco Project Health Check
━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Project: myapp
Location: /Users/dev/myapp
Trabuco Version: 1.2.0

Running checks...

  ✓ Project structure valid
  ✓ Trabuco project detected
  ✓ Metadata file valid
  ✓ Parent POM valid
  ✓ Module POMs exist (4 modules)
  ✓ Java version consistent (21)
  ✓ Docker Compose in sync

━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status: HEALTHY
All 7 checks passed
```

**Doctor options:**

| Option | Description |
|--------|-------------|
| `--verbose` | Show all checks, not just failures |
| `--fix` | Auto-fix issues that can be fixed automatically |
| `--json` | Output as JSON (for CI/scripting) |

**Auto-fix capabilities:**

```bash
trabuco doctor --fix
```

This can automatically fix common issues like missing `.trabuco.json` metadata, out-of-sync module lists, and inconsistent Java versions across POMs.

### Adding Modules

Start with a minimal project and add modules as you need them:

```bash
# Create a simple API project
trabuco init \
  --name=myapp \
  --group-id=com.company.myapp \
  --modules=Model,Shared,API

# Later, add a database
cd myapp
trabuco add SQLDatastore --database=postgresql

# Even later, add background jobs
trabuco add Worker

# Or add event-driven messaging
trabuco add EventConsumer --message-broker=kafka
```

The `add` command automatically:
- Runs `doctor` to validate the project before making changes
- Creates the module directory structure
- Updates the parent POM with the new module
- Adds required properties and dependencies
- Updates `docker-compose.yml` with necessary services
- Regenerates CI workflow with new services (if CI is configured)
- Regenerates `AGENTS.md` with updated module list
- Updates `.trabuco.json` metadata
- Auto-includes dependent modules (e.g., `Worker` includes `Jobs`)
- Prompts to add CI if not already configured

**Add command options:**

| Option | Description |
|--------|-------------|
| `--database` | SQL database type (for SQLDatastore): `postgresql`, `mysql` |
| `--nosql-database` | NoSQL database type (for NoSQLDatastore): `mongodb`, `redis` |
| `--message-broker` | Message broker (for EventConsumer): `kafka`, `rabbitmq`, `sqs`, `pubsub` |
| `--dry-run` | Show what would change without making modifications |
| `--no-backup` | Skip creating backup before modifications |

**Interactive mode:**

```bash
trabuco add
```

If you don't specify a module, Trabuco prompts you to select one and asks for any required options.

**Dry-run example:**

```bash
trabuco add SQLDatastore --database=postgresql --dry-run
```

This shows exactly what files will be created and modified without changing anything.

**Backup and recovery:**

By default, `add` creates a backup in `.trabuco-backup/` before modifying files. If something goes wrong, you can restore from this backup. The backup is overwritten on each successful `add` operation.

**Module compatibility:**

| Adding | Conflicts With | Auto-Includes |
|--------|----------------|---------------|
| SQLDatastore | NoSQLDatastore | — |
| NoSQLDatastore | SQLDatastore | — |
| Worker | — | Jobs, Model |
| EventConsumer | — | Events, Model |

### Syncing AI Tooling

Trabuco's AI-tooling layer evolves across releases: new skills, new subagents, new task prompts, new review rules, new hooks. Projects generated on older CLIs keep their original files and miss anything the CLI added afterwards — the coding agents working on those projects run with a stale tool belt.

`trabuco sync` closes that gap. It's additive-only: given a project's `.trabuco.json` and the currently-installed CLI, it identifies every AI-tooling file the current CLI would generate that's missing from the project, and creates those files on `--apply`.

```bash
# Dry-run — shows what would be added, no writes
trabuco sync

# Actually create the missing files
trabuco sync --apply

# Machine-readable plan (for CI, scripting, or agent consumption)
trabuco sync --json
```

**What sync touches (in-jurisdiction):**

- `.ai/**` — prompts, task guides, quality specs
- `.claude/**` — settings, skills, subagents, hooks, rules
- `.cursor/**`, `.codex/**`, `.agents/**` — per-agent files
- `.github/instructions/**`, `.github/skills/**`, `.github/scripts/review-checks.sh`, `.github/workflows/copilot-setup-steps.yml`, `.github/copilot-instructions.md` — Copilot and cross-tool files
- `.trabuco/review.config.json` — review runtime config
- `CLAUDE.md`, `AGENTS.md` — top-level agent context files

**What sync NEVER touches (out of jurisdiction):**

- Java source (`**/src/**`), Flyway migrations, Maven POMs
- `application.yml`, `application.properties`, profile configs
- `docker-compose.yml`, `.env*`, `.dockerignore`
- CI workflow (`.github/workflows/ci.yml`) — only `copilot-setup-steps.yml` is in scope
- `.run/` IntelliJ configs, `README.md`, `.trabuco.json`
- `.ai/checkpoint.json` — live session state, explicitly excluded
- Any Git internals

Sync uses a whitelist of path prefixes, validated at both planning and write time. Business code is physically unreachable from this command — even if a future change added a registry entry for a Java source path, write-time validation would refuse.

**Operational guarantees:**

- **Additive only.** Existing files are never modified or deleted, regardless of how stale their content is. To refresh a file like `CLAUDE.md` with newer content, delete the file and re-run `trabuco sync --apply`.
- **Idempotent.** Running `trabuco sync --apply` twice in a row always succeeds; the second run is a no-op.
- **Atomic per-file.** Each file is written via tmpfile + rename, so a crash during sync cannot leave a partial file in the project.
- **Zero drift from init.** The expected state is produced by running the current generator against a temporary directory, so what sync considers "missing" is exactly what `trabuco init` would emit for the same project configuration.

**Typical flow after upgrading the CLI:**

```bash
# Upgrade the CLI (npm, curl, go install — your choice)
npm install -g trabuco-mcp@latest

# In your existing project
cd my-project
trabuco sync                # see what's new
trabuco sync --apply        # bring the AI layer up to date
trabuco doctor              # sanity-check the project after
```

## CLI MCP Server

Trabuco includes a built-in [Model Context Protocol](https://modelcontextprotocol.io) server that exposes all CLI functionality as structured tools. Instead of running shell commands and parsing terminal output, AI coding agents get proper JSON schemas for inputs and structured JSON results — no string parsing, no color codes, no guessing.

```bash
trabuco mcp
```

This starts the MCP server over stdio. You don't run this command directly — your AI agent launches it automatically based on its configuration.

### Configuration

Add Trabuco as an MCP server in your AI agent's configuration file. The recommended approach uses `npx` so there's nothing to install first:

**Claude Code**

```bash
claude mcp add trabuco -- npx -y trabuco-mcp
```

Or add to `.mcp.json` in your project root (shared with your team):

```json
{
  "mcpServers": {
    "trabuco": {
      "command": "npx",
      "args": ["-y", "trabuco-mcp"]
    }
  }
}
```

**Cursor** — add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "trabuco": {
      "command": "npx",
      "args": ["-y", "trabuco-mcp"]
    }
  }
}
```

**VS Code / GitHub Copilot** — add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "trabuco": {
      "command": "npx",
      "args": ["-y", "trabuco-mcp"]
    }
  }
}
```

**Codex** — add to `.codex/config.toml`:

```toml
[mcp_servers.trabuco]
command = "npx"
args = ["-y", "trabuco-mcp"]
```

<details>
<summary>Using the CLI binary directly (if already installed)</summary>

If you installed the Trabuco CLI via `curl | bash` or `go install`, you can reference the binary directly:

**Claude Code**

```bash
claude mcp add --transport stdio trabuco -- trabuco mcp
```

**All agents** — use `"command": "trabuco"` with `"args": ["mcp"]` in the config files shown above.

</details>

### Available Tools

Once configured, your AI agent can use these tools:

| Tool | Description |
|------|-------------|
| `suggest_architecture` | Analyze requirements and recommend modules, database, and architecture pattern |
| `design_system` | Decompose requirements into a multi-service system design (review-only) |
| `generate_workspace` | Generate a multi-service workspace with shared Docker Compose |
| `init_project` | Generate a new Java project with specified modules, database, and options |
| `add_module` | Add a module to an existing Trabuco project (with dry-run support) |
| `run_doctor` | Run health checks on a project and optionally auto-fix issues |
| `get_project_info` | Read project metadata and available actions |
| `check_docker` | Check if Docker is installed and running |
| `get_version` | Get the Trabuco CLI version |
| `auth_status` | Check which AI providers have credentials configured |
| `list_providers` | List supported AI providers with pricing and model info |
| `list_modules` | List all available modules with descriptions and dependency info |

### Prompts

Prompts provide expert knowledge for complex decisions:

| Prompt | Description |
|--------|-------------|
| `trabuco_expert` | General guidance for any Trabuco task — module decisions, common pitfalls, post-generation steps |
| `design_microservices` | Step-by-step guide for decomposing requirements into multiple services |
| `extend_project` | Instructions for adding features to an existing Trabuco project |
| `trabuco_ai_agent_expert` | Expert guidance for building and customizing AI agents with the AIAgent module |

### Resources

Resources provide stable reference data:

| Resource | Description |
|----------|-------------|
| `trabuco://modules` | Full module catalog with use cases, boundaries, dependencies, and conflicts |
| `trabuco://patterns` | Pre-built architecture patterns with module combinations and recommendations |
| `trabuco://limitations` | What Trabuco does NOT generate — check before suggesting Trabuco for a requirement |

**What this looks like in practice:** Describe your business to your AI agent — "I need an intelligent assistant that can answer customer questions, check order status, and schedule deliveries" — and it calls `suggest_architecture` to match the `ai-agent` pattern, then `init_project` with `Model,Shared,AIAgent` to generate a complete AI agent with tools, guardrails, and MCP server.

## Generated Project Structure

```
myapp/
├── pom.xml                          # Parent POM (module aggregator)
├── Model/                           # DTOs, Entities, Enums
│   └── src/main/java/.../model/
│       ├── dto/                     # Request/Response DTOs
│       ├── entities/                # Domain entities + DB records/documents
│       └── ImmutableStyle.java      # Immutables configuration
├── SQLDatastore/                    # SQL database layer (if selected)
│   └── src/
│       ├── main/
│       │   ├── java/.../sqldatastore/
│       │   │   ├── config/          # Database configuration
│       │   │   └── repository/      # Spring Data JDBC repositories
│       │   └── resources/
│       │       └── db/migration/    # Flyway SQL migrations
│       └── test/                    # Testcontainers integration tests
├── NoSQLDatastore/                  # NoSQL database layer (if selected)
│   └── src/
│       ├── main/java/.../nosqldatastore/
│       │   ├── config/              # NoSQL configuration
│       │   └── repository/          # Spring Data repositories
│       └── test/                    # Testcontainers integration tests
├── Shared/                          # Business logic & services
│   └── src/main/java/.../shared/
│       ├── config/                  # Circuit breaker config
│       └── service/                 # Service classes
├── API/                             # REST API (Spring Boot app)
│   └── src/main/
│       ├── java/.../api/
│       │   ├── controller/          # REST controllers
│       │   └── config/              # Web configuration
│       └── resources/
│           └── application.yml      # App configuration
├── Jobs/                            # Job services (auto-included with Worker)
│   └── src/main/java/.../jobs/
│       └── PlaceholderJobService.java  # Service for enqueueing jobs
├── Worker/                          # Background jobs (Spring Boot app)
│   └── src/main/
│       ├── java/.../worker/
│       │   ├── config/              # JobRunr configuration
│       │   └── handler/             # JobRequestHandler implementations
│       └── resources/
│           └── application.yml      # Worker configuration
├── Events/                          # Event publisher (auto-included with EventConsumer)
│   └── src/main/java/.../events/
│       └── EventPublisher.java      # Service for publishing events
├── EventConsumer/                   # Event listeners (Spring Boot app)
│   └── src/main/
│       ├── java/.../eventconsumer/
│       │   ├── config/              # Message broker configuration
│       │   └── listener/            # Event listener implementations
│       └── resources/
│           └── application.yml      # Consumer configuration
├── AIAgent/                         # AI Agent (Spring Boot app, if selected)
│   └── src/main/
│       ├── java/.../aiagent/
│       │   ├── config/              # ChatClient, MCP server, WebConfig
│       │   ├── security/            # Auth filter, scopes, guardrails, rate limiter
│       │   ├── tool/                # @Tool-annotated domain tools
│       │   ├── agent/               # Primary + Specialist agents
│       │   ├── brain/               # Scratchpad, reflection service
│       │   ├── knowledge/           # Knowledge base, token-free Q&A
│       │   ├── protocol/            # REST, A2A, discovery, SSE, webhooks
│       │   ├── task/                # Async task manager
│       │   └── event/               # Webhook dispatch
│       └── resources/
│           ├── application.yml      # Spring AI + Anthropic config
│           └── .well-known/
│               └── agent.json       # A2A discovery agent card
├── .ai/                             # AI context directory
│   ├── prompts/                     # Task guides and quality specs
│   │   ├── JAVA_CODE_QUALITY.md     # Code quality specification
│   │   ├── code-review.md           # Review checklist
│   │   ├── add-entity.md            # How to add an entity
│   │   ├── add-endpoint.md          # How to add a REST endpoint
│   │   ├── add-job.md               # How to add a background job
│   │   └── add-event.md             # How to add an event type
│   └── checkpoint.json              # Session state for AI continuity
├── .github/workflows/ci.yml         # GitHub Actions CI (if --ci github)
├── docker-compose.yml               # Local dev stack (database, message broker)
├── .run/                            # IntelliJ run configurations
├── .cursor/                         # Cursor IDE configuration
│   ├── rules/java.mdc               # Java coding rules
│   └── hooks.json                   # Auto-formatting hooks
├── CLAUDE.md                        # AI assistant context (Claude Code)
├── AGENTS.md                        # Cross-tool AI agent baseline
└── README.md                        # Project documentation
```

**Note:** SQLDatastore and NoSQLDatastore are mutually exclusive — choose one based on your data storage needs.

## Modules

### Model

The foundation. Contains all your data structures.

| What | Description |
|------|-------------|
| **DTOs** | Request/Response objects using Immutables |
| **Entities** | Domain objects with `ImmutableX.builder()` pattern |
| **Records** | Simple Java records for database persistence |
| **Enums** | Domain enumerations |

**Key principle:** Always use `ImmutableX` types and builders. Never instantiate interfaces directly.

```java
// ✅ Correct
ImmutablePlaceholder entity = ImmutablePlaceholder.builder()
    .name("example")
    .description("description")
    .build();

// ❌ Avoid
Placeholder entity = Placeholder.fromRecord(record);
```

### SQLDatastore

Database access layer using Spring Data JDBC.

| What | Description |
|------|-------------|
| **Repositories** | Spring Data JDBC repositories (not JPA!) |
| **Migrations** | Flyway SQL migrations in `db/migration/` |
| **Config** | HikariCP connection pool configuration |
| **Tests** | Testcontainers-based integration tests |

**Why JDBC over JPA?** No lazy loading gotchas, no proxy magic, no `@Transactional` surprises. What you write is what runs.

### NoSQLDatastore

NoSQL database access layer using Spring Data.

| What | Description |
|------|-------------|
| **Repositories** | Spring Data MongoDB or Redis repositories |
| **Config** | Database connection configuration |
| **Tests** | Testcontainers-based integration tests |

**Supported databases:**
- **MongoDB** — Document store with flexible schemas
- **Redis** — Key-value store for high-performance caching and data

### Shared

Business logic and cross-cutting concerns.

| What | Description |
|------|-------------|
| **Services** | Business logic with `@CircuitBreaker` support |
| **Config** | Resilience4j circuit breaker configuration |

Services convert between `PlaceholderRecord` (database) and `ImmutablePlaceholder` (domain) at the boundary.

### API

The REST API module — a runnable Spring Boot application.

| What | Description |
|------|-------------|
| **Controllers** | REST endpoints with validation |
| **Config** | CORS, Jackson, web configuration |
| **Health** | `/health` endpoint for load balancers |

Endpoints use `ImmutablePlaceholderRequest` for input and `ImmutablePlaceholderResponse` for output.

### Jobs

Job service module — contains services for enqueueing background jobs.

| What | Description |
|------|-------------|
| **PlaceholderJobService** | Service for enqueueing placeholder jobs |

The Jobs module is **auto-included** when Worker is selected. Job request schemas (sealed interfaces and records) live in the Model module for decoupled access.

**Enqueueing jobs via service:**
```java
@Autowired
private PlaceholderJobService jobService;

// Fire-and-forget (immediate)
jobService.processAsync("data");

// Delayed (at specific time)
jobService.processAt("data", Instant.now().plusHours(1));

// Batch (multiple items)
jobService.processBatch(List.of("item1", "item2", "item3"));
```

### Worker

Background job processing module — a runnable Spring Boot application using JobRunr.

| What | Description |
|------|-------------|
| **Handlers** | `JobRequestHandler` implementations that process job requests |
| **Config** | JobRunr configuration and recurring job registration |
| **Dashboard** | JobRunr dashboard at `http://localhost:8000` |
| **Health** | Actuator health endpoints at port 8082 |

**Architecture:** Jobs module contains request contracts, Worker module contains handlers. This allows any module to enqueue jobs without circular dependencies.

**Supported job types:**

| Type | Description | How to Enqueue |
|------|-------------|----------------|
| Fire-and-forget | Execute immediately in background | `BackgroundJobRequest.enqueue(request)` |
| Delayed | Execute at a specific time | `BackgroundJobRequest.schedule(instant, request)` |
| Recurring | Execute on a CRON schedule | Register in `RecurringJobsConfig` |
| Batch | Process multiple items efficiently | `BackgroundJobRequest.enqueue(requests.stream())` |

**Storage notes:**
- Worker uses your selected datastore (SQL or MongoDB) for job persistence
- If you select Redis, Worker uses PostgreSQL for job storage (Redis is deprecated in JobRunr 8+)
- If no datastore is selected, Worker defaults to PostgreSQL
- Jobs module is auto-included when Worker is selected

### Events

Event publisher module — contains services for publishing events to message brokers.

| What | Description |
|------|-------------|
| **EventPublisher** | Service for publishing events to your chosen message broker |
| **Config** | Message serialization configuration (broker-specific) |

The Events module is **auto-included** when EventConsumer is selected. Event schemas (sealed interfaces and records) live in the Model module for decoupled access.

**Publishing events:**
```java
@Autowired
private EventPublisher eventPublisher;

// Publish an event
eventPublisher.publish(new PlaceholderCreatedEvent("id-123", "Example", Instant.now()));
```

### EventConsumer

Event consumer module — a runnable Spring Boot application that listens for events.

| What | Description |
|------|-------------|
| **Listeners** | Event listener implementations for your chosen broker |
| **Config** | Message broker consumer configuration |
| **Dead Letter Queues** | Automatic DLQ/DLT setup for failed messages |

**Supported message brokers:**

| Broker | Description | Use Case |
|--------|-------------|----------|
| **Kafka** | High-throughput distributed streaming | Large-scale event streaming, log aggregation |
| **RabbitMQ** | Feature-rich message broker | Task queues, pub/sub, routing patterns |
| **AWS SQS** | Managed queue service | Serverless, AWS-native applications |
| **GCP Pub/Sub** | Google Cloud messaging | GCP-native applications, global distribution |

**Architecture:** Events module contains the publisher service, EventConsumer module contains listeners. This allows any module to publish events without circular dependencies. Event schemas live in the Model module.

**Listener examples:**

```java
// Kafka
@KafkaListener(topics = "${app.kafka.topics.placeholder-events}")
public void handleEvent(PlaceholderEvent event) { ... }

// RabbitMQ
@RabbitListener(queues = "${app.rabbitmq.queues.placeholder-events}")
public void handleEvent(PlaceholderEvent event) { ... }

// AWS SQS
@SqsListener("${app.sqs.queue.placeholder-events}")
public void handleEvent(PlaceholderEvent event, Acknowledgement ack) { ... }

// GCP Pub/Sub (uses Spring Integration)
@ServiceActivator(inputChannel = "placeholderInputChannel")
public void handleEvent(PlaceholderEvent event, BasicAcknowledgeablePubsubMessage msg) { ... }
```

### AI Agent

Production AI agent module — a runnable Spring Boot application powered by Spring AI with Anthropic Claude.

| What | Description |
|------|-------------|
| **Tools** | `@Tool`-annotated methods the LLM can call — replace placeholders with your domain logic |
| **Primary Agent** | Customer-facing agent with system prompt, tools, and circuit breaker |
| **Specialist Agent** | Worker agent for delegated sub-tasks (orchestrator-worker pattern) |
| **Input Guardrail** | LLM-based classifier blocks prompt injection and off-topic requests before processing |
| **Output Guardrail** | Regex PII detection (email, SSN, credit card, phone) on all responses |
| **Knowledge Base** | Token-free Q&A via keyword matching — saves LLM tokens for FAQ-style questions |
| **MCP Server** | Exposes all `@Tool` beans to Claude Code, Cursor, and other MCP clients |
| **A2A Protocol** | JSON-RPC 2.0 agent-to-agent communication with task lifecycle |
| **SSE Streaming** | Real-time task status updates via Server-Sent Events |
| **Webhooks** | HMAC-signed outbound event notifications for order/status changes |
| **Scratchpad** | Agent reasoning trace for debugging and auditability |
| **Reflection** | LLM-driven error recovery: retry, suggest alternative, escalate, or give up |

**Architecture:** The AI Agent follows Anthropic's production recommendations. Deterministic operations (listing items, checking stock) use plain code — no LLM tokens spent. The LLM is only invoked when natural language understanding or judgment is required (routing, multi-step reasoning, error recovery). The circuit breaker (`@CircuitBreaker(name = "llm")`) protects against LLM API outages with automatic fallback responses.

**Multi-agent pattern:**

```
User message → Primary Agent (orchestrator)
                    │
                    ├── calls domain tools directly (simple queries)
                    │
                    └── calls askSpecialist tool (complex queries)
                              │
                              └── Specialist Agent (worker)
                                      └── calls specialist tools
```

The Primary Agent decides when to delegate — same mechanism as any other tool call. The Specialist Agent never calls back (acyclic flow prevents infinite loops).

**Security pipeline (every request):**

```
Request → ApiKeyAuthFilter → ScopeInterceptor → RateLimiter
        → InputGuardrailAdvisor (LLM) → Agent → OutputGuardrailAdvisor (regex) → Response
```

**Graceful degradation:** All LLM-dependent beans use `@ConditionalOnBean(ChatModel.class)`. Without an `ANTHROPIC_API_KEY`, the application starts and serves non-AI endpoints (capabilities, health). The `/chat` endpoint returns a helpful message instead of failing.

**Connecting AI assistants:**

```bash
# Claude Code
claude mcp add --transport http my-agent http://localhost:8080/mcp

# Or via .mcp.json in your repo
{"mcpServers": {"my-agent": {"type": "http", "url": "http://localhost:8080/mcp"}}}
```

**Endpoints:**

| Method | Path | Scope | Description |
|--------|------|-------|-------------|
| POST | `/chat` | public | Natural language conversation (agent brain) |
| POST | `/ask` | public | Knowledge base Q&A (no LLM tokens) |
| POST | `/a2a` | varies | A2A JSON-RPC 2.0 (tasks/send, tasks/chat, tasks/get) |
| GET | `/capabilities` | anonymous | Live tool registry and protocol status |
| GET | `/tasks/{id}/stream` | partner | SSE task status streaming |
| POST | `/webhooks` | partner | Register webhook URL |
| GET | `/webhooks` | partner | List registered webhooks |
| DELETE | `/webhooks/{id}` | partner | Deregister webhook |

**Getting started:**

```bash
# Generate project with AI Agent
trabuco init --name=my-agent --group-id=com.company.agent --modules=Model,Shared,AIAgent

# Start without API key (infrastructure endpoints only)
cd my-agent && mvn spring-boot:run -pl AIAgent

# Start with full AI capabilities
ANTHROPIC_API_KEY=your-key mvn spring-boot:run -pl AIAgent
```

**Customization:**
1. Replace `PlaceholderTools` methods with your domain tools (follow the `@Tool` + `@ToolParam` annotations)
2. Update the system prompt in `PrimaryAgent` for your domain's personality and rules
3. Add domain-specific ALLOW/BLOCK criteria in `InputGuardrailAdvisor`
4. Replace `KnowledgeBase` entries with your FAQ content
5. Configure API keys and rate limits in `application.yml`

**Immutables:** All DTOs (`ChatRequest`, `ChatResponse`, `AskRequest`, `AskResponse`, `JsonRpcRequest`, `JsonRpcResponse`, `WebhookRegisterRequest`) and value objects (`CallerIdentity`, `MemoryEntry`, `ReflectionDecision`, `TaskEvent`, `WebhookRegistration`) use the Immutables library with the project's `@ImmutableStyle` — consistent with all other modules.

**Observability:** Actuator + Prometheus metrics are pre-configured. Spring AI auto-instruments `ChatClient` calls, providing `gen_ai.client.token.usage` metrics for cost tracking. Correlation IDs propagate through the full agent pipeline (guardrail → agent → tools → reflection) via MDC.

## Code Quality & Architecture

Generated projects come with strict code quality enforcement out of the box. Every project includes Google Java Format via Spotless, Maven Enforcer for dependency rules, and ArchUnit for architecture tests. These run as part of the normal build — violations fail the build, not just a linter warning.

### Auto-Formatting

All generated code follows Google Java Format (2-space indentation, specific import ordering). Trabuco runs `spotless:apply` during project generation so your code is formatted from the first commit.

**Manual commands:**

```bash
mvn spotless:apply      # Auto-format all Java files
mvn spotless:check      # Check formatting without modifying (CI-friendly)
mvn enforcer:enforce    # Check dependency rules
```

**IDE integration:** When you select AI coding agents during setup, Trabuco generates auto-formatting hooks so code stays formatted as you work:

| Agent | Hook File | Behavior |
|-------|-----------|----------|
| Claude Code | `.claude/settings.json` | Runs `spotless:apply` after Write/Edit operations |
| Cursor | `.cursor/hooks.json` | Runs `spotless:apply` after file edits |

### Architecture Tests

The Shared module includes [ArchUnit](https://www.archunit.org/) tests that enforce architectural rules at build time:

| Rule | Description |
|------|-------------|
| No field injection | `@Autowired` on fields is forbidden — use constructor injection |
| Controller-service boundary | Controllers cannot access repositories directly |
| No cyclic dependencies | Cross-module cyclic dependencies are not allowed |

These tests run as part of `mvn test` and fail the build if violated. To add project-specific rules, edit `Shared/src/test/java/.../shared/ArchitectureTest.java`.

### AI Task Prompts

Every generated project includes an `.ai/` directory with task-specific guides for AI coding assistants. Instead of relying on the AI to guess your project's patterns, these prompts provide step-by-step instructions with file locations and code examples.

| Prompt | Description |
|--------|-------------|
| `JAVA_CODE_QUALITY.md` | Comprehensive code quality specification |
| `code-review.md` | Review checklist and process |
| `add-entity.md` | How to add a new entity with migrations |
| `add-endpoint.md` | How to add a REST endpoint (if API selected) |
| `add-job.md` | How to add a background job (if Worker selected) |
| `add-event.md` | How to add an event type (if EventConsumer selected) |

The `checkpoint.json` file tracks session state (current work, completed steps, test status) so AI assistants can resume context across sessions.

### Code Review Workflow

The generated project includes AI task prompts and quality specifications that support structured code review. Claude Code gets a `/review` skill that reviews code against the project's quality specification.

**How it works:**
1. The AI reads `.ai/prompts/JAVA_CODE_QUALITY.md` for the project's quality rules
2. Claude Code loads path-scoped rules from `.claude/rules/` when Java files are accessed
3. The AI reviews code against relevant rules and reports findings by severity

**Review categories:** code-quality, modern-java, architecture, security, testing

## CI/CD

### GitHub Actions

Trabuco can generate a GitHub Actions CI workflow that matches your project's module configuration. The workflow is opt-in — pass `--ci github` during `init` or answer the CI prompt in interactive mode.

```bash
# Generate project with CI
trabuco init --name=myapp --group-id=com.company.myapp \
  --modules=Model,SQLDatastore,Shared,API --database=postgresql --ci github
```

The generated `.github/workflows/ci.yml` runs on pushes and pull requests to `main` and includes:

| Step | Description |
|------|-------------|
| Java setup | Configures your selected Java version with Maven caching |
| Compile | `mvn clean compile` |
| Format check | `mvn spotless:check` — fails if code isn't formatted |
| Dependency rules | `mvn enforcer:enforce` — fails if dependency boundaries are violated |
| Tests | `mvn test` — runs all tests including ArchUnit and Testcontainers |

**Conditional services:** The workflow automatically includes Docker services based on your modules:

| Module | Service |
|--------|---------|
| SQLDatastore (PostgreSQL) | PostgreSQL container with health check |
| SQLDatastore (MySQL) | MySQL container with health check |
| NoSQLDatastore (MongoDB) | MongoDB container |
| NoSQLDatastore (Redis) | Redis container |
| EventConsumer (Kafka) | Kafka + Zookeeper containers |
| EventConsumer (RabbitMQ) | RabbitMQ container |
| EventConsumer (SQS) | LocalStack with auto-created queue |
| EventConsumer (Pub/Sub) | Pub/Sub emulator with topic/subscription |
| Worker (no datastore) | PostgreSQL container for JobRunr storage |

**Regeneration on module addition:** When you add a module with `trabuco add`, the CI workflow is automatically regenerated to include the new services. If CI wasn't configured during `init`, you'll be prompted to add it after a module addition.

## Observability

### Metrics

All runtime modules expose Prometheus metrics at `/actuator/prometheus`. These include:
- JVM metrics (memory, GC, threads)
- HTTP request metrics (latency, status codes)
- Database connection pool metrics
- Circuit breaker state

### API Documentation

The API module includes Swagger UI for interactive API exploration:
- Swagger UI: `http://localhost:8080/swagger-ui.html`
- OpenAPI spec: `http://localhost:8080/api-docs`

Disable in production by setting `SPRINGDOC_ENABLED=false`.

### Request Tracing

Every request is assigned a correlation ID for distributed tracing:
- Incoming `X-Correlation-ID` header is preserved
- If not present, a new UUID is generated
- Correlation ID is included in all log entries
- Correlation ID is returned in response headers

### Health Checks

Health endpoints for monitoring and orchestration:
- `/actuator/health` — Overall health
- `/actuator/health/readiness` — Kubernetes readiness probe
- `/actuator/health/liveness` — Kubernetes liveness probe

Database and message broker connectivity is automatically included.

### Test Coverage

JaCoCo is configured for test coverage reporting. After running tests, coverage reports are available at:
- `<module>/target/site/jacoco/index.html`

Run tests with coverage:
```bash
mvn test
```

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--name` | Project name (lowercase, hyphens allowed) | — |
| `--group-id` | Maven group ID (e.g., `com.company.project`) | — |
| `--modules` | Modules to include (comma-separated) | — |
| `--database` | SQL database type: `postgresql`, `mysql`, `none` | `postgresql` |
| `--nosql-database` | NoSQL database type: `mongodb`, `redis` | `mongodb` |
| `--message-broker` | Message broker: `kafka`, `rabbitmq`, `sqs`, `pubsub` | `kafka` |
| `--java-version` | Java version: `21`, `25`, or `26` | `21` |
| `--ai-agents` | AI coding agents (comma-separated): `claude`, `cursor`, `copilot`, `codex` | — |
| `--ci` | CI/CD provider: `github` | — |
| `--skip-build` | Skip running `mvn clean install` after generation | `false` |
| `--strict` | Fail if specified Java version is not detected | `false` |

### Available Modules

| Module | Description | Dependencies |
|--------|-------------|--------------|
| `Model` | DTOs, Entities, Enums, Event/Job schemas | None (always included) |
| `SQLDatastore` | SQL Repositories, Migrations | Model |
| `NoSQLDatastore` | NoSQL Repositories | Model |
| `Shared` | Services, Circuit breakers | Model |
| `API` | REST endpoints | Model |
| `Worker` | Background jobs (JobRunr) | Model, Jobs (auto) |
| `EventConsumer` | Event listeners (Kafka/RabbitMQ/SQS/Pub/Sub) | Model, Events (auto) |
| `AIAgent` | AI agent (Spring AI, tools, guardrails, MCP, A2A) | Model |

**Notes:**
- SQLDatastore and NoSQLDatastore are mutually exclusive
- Worker uses your datastore for job persistence (defaults to PostgreSQL if none selected)
- Jobs module is auto-included when Worker is selected (not shown in CLI)
- Events module is auto-included when EventConsumer is selected (not shown in CLI)

### Java Version Detection

Trabuco automatically detects installed Java versions on your system. In interactive mode, version options show detection status:

```
Java version:
> 21 (LTS until 2031 - Recommended) [detected]
  25 (Latest LTS) [not detected]
```

If you select an undetected version, you'll be asked to confirm. In non-interactive mode, a warning is shown but the project is still generated. Use `--strict` to fail instead:

```bash
# Warns but continues
trabuco init --name=myapp --group-id=com.example --modules=Model --java-version=25

# Fails if Java 25 not installed
trabuco init --name=myapp --group-id=com.example --modules=Model --java-version=25 --strict
```

### AI Coding Agents

Trabuco generates context files, coding rules, and quality hooks for popular AI coding assistants. These aren't generic instructions — they contain your project's actual module structure, dependency boundaries, and quality standards.

| Agent | Files Generated | Description |
|-------|----------------|-------------|
| Claude Code | `CLAUDE.md`, `.claude/settings.json`, `.claude/skills/review.md` | Project context, permissions, auto-formatting hooks, code review skill |
| Cursor | `.cursor/rules/java.mdc`, `.cursor/hooks.json` | Java coding rules with auto-formatting hooks |
| GitHub Copilot | `.github/instructions/java.instructions.md`, `.github/workflows/copilot-setup-steps.yml` | Java coding instructions and cloud agent setup |
| Codex | `AGENTS.md`, `.codex/hooks.json` | Full project context in AGENTS.md, auto-formatting hooks |

Every agent also gets `AGENTS.md` — a cross-tool baseline with the project's structure, build commands, module dependencies, and coding patterns. Codex uses `AGENTS.md` as its primary context file, so when selected it receives the full project architecture and coding standards.

In interactive mode, you'll be prompted to select which agents you want context files for. In non-interactive mode:

```bash
# Generate for specific agents
trabuco init --name=myapp --group-id=com.example --modules=Model,API --ai-agents=claude,cursor

# Generate for all agents
trabuco init --name=myapp --group-id=com.example --modules=Model,API --ai-agents=claude,cursor,copilot,codex
```

All agents also get the `.ai/` directory with task prompts and quality specifications. See [Code Quality & Architecture](#code-quality--architecture) for details.

### CI/CD Provider

Trabuco can generate a CI/CD workflow during project creation. In interactive mode, you'll be prompted to choose a CI provider. In non-interactive mode, use the `--ci` flag:

```bash
# Generate with GitHub Actions CI
trabuco init --name=myapp --group-id=com.example --modules=Model,SQLDatastore,Shared,API \
  --database=postgresql --ci github
```

Currently supported providers:

| Provider | Flag Value | What's Generated |
|----------|-----------|-----------------|
| GitHub Actions | `github` | `.github/workflows/ci.yml` |

See [CI/CD](#cicd) for details on what the workflow includes.

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Java | 21, 25, or 26 | Runtime |
| Spring Boot | 3.4.2 | Application framework |
| Spring Data JDBC | — | SQL database access |
| Spring Data MongoDB | — | MongoDB access |
| Spring Data Redis | — | Redis access |
| Spring Kafka | — | Kafka messaging |
| Spring AMQP | — | RabbitMQ messaging |
| Spring Cloud AWS | 3.2.0 | AWS SQS messaging |
| Spring Cloud GCP | 5.8.0 | GCP Pub/Sub messaging |
| Immutables | 2.10.1 | Immutable value objects |
| Flyway | — | SQL database migrations |
| JobRunr | 7.3.2 | Background job processing |
| Testcontainers | 2.0.3 | Integration testing |
| ArchUnit | — | Architecture enforcement tests |
| Spotless | — | Code formatting (Google Java Format) |
| Resilience4j | — | Circuit breakers |
| PostgreSQL / MySQL | — | SQL databases |
| MongoDB / Redis | — | NoSQL databases |
| Apache Kafka | — | Distributed streaming |
| RabbitMQ | — | Message broker |
| AWS SQS | — | Managed queue service (via LocalStack for local dev) |
| GCP Pub/Sub | — | Google Cloud messaging (via emulator for local dev) |
| HikariCP | — | Connection pooling (SQL) |
| Spring AI | 1.0.5 | AI/LLM integration framework |
| Anthropic Claude | — | LLM provider for AI Agent module |
| MCP Server | — | Model Context Protocol for tool exposure |

## Local Development

The generated project includes a `docker-compose.yml` for local development:

```bash
docker-compose up -d    # Start database (and message broker if EventConsumer selected)
mvn spring-boot:run -pl API
```

If you selected EventConsumer, the docker-compose includes the appropriate local service:
- **Kafka** — Kafka with Zookeeper
- **RabbitMQ** — RabbitMQ with management UI
- **AWS SQS** — LocalStack with auto-created queue
- **GCP Pub/Sub** — Pub/Sub emulator with auto-created topic/subscription

### Running Tests

```bash
mvn test                # All tests
mvn test -pl Model      # Single module
```

**Note:** SQLDatastore and NoSQLDatastore tests require Docker to be running (Testcontainers).

## Requirements

- **Java 21+** (21, 25, or 26 — Trabuco auto-detects installed versions)
- **Maven 3.8+**
- **Docker** (for Testcontainers and local development)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License — do whatever you want with it.

---

**Built with Trabuco.** Now go build something amazing.
