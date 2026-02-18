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

<h3 align="center">Generate production-ready Java projects in seconds — or migrate existing ones with AI.</h3>

Starting a new Java project shouldn't feel like a chore. Yet every time you begin, it's the same story — setting up Maven modules, configuring database connections, writing migration scripts, wiring up test infrastructure, and copying boilerplate from that one project that "mostly works." Hours pass before you write a single line of actual business logic. Trabuco exists because your time is better spent building features, not fighting configuration files.

Trabuco is a command-line tool that generates complete, well-structured Java projects with a single command. Run `trabuco init`, answer a few questions (or pass flags for automation), and you get a fully functional multi-module Maven project ready for development. No templates to download, no manual setup, no guessing how things should connect. The CLI handles the tedious work so you can focus on what matters — your application's unique value.

But Trabuco isn't just for greenfield projects. The `trabuco migrate` command uses AI (Claude) to analyze existing Spring Boot applications and intelligently transform them into Trabuco's clean multi-module architecture. It understands your entities, services, controllers, and repositories — then reorganizes them into the right modules, converts JPA to Spring Data JDBC, replaces Quartz with JobRunr, and generates all the configuration you need. Migration happens in stages with checkpoints, so you can resume if interrupted and roll back if needed.

The generated projects come batteries-included with production-proven technologies: Spring Boot for the application framework, Spring Data JDBC for straightforward database access, Flyway for version-controlled migrations, Testcontainers for realistic integration tests, and Resilience4j for fault tolerance. Everything is pre-configured and working together out of the box. Need PostgreSQL instead of MySQL? Just pick it during setup. Want the latest Java 25 instead of 21? One flag changes everything. The architecture is designed to be solid by default yet flexible when you need it.

The real power lies in the modular structure. Instead of a monolithic source tree where everything depends on everything, Trabuco generates clean, separated modules: Model for your data structures, SQLDatastore or NoSQLDatastore for persistence, Shared for business logic and services, and API for your REST endpoints. Each module has a clear responsibility and well-defined dependencies. This isn't just organization for organization's sake — it enforces good architecture, makes testing straightforward, helps new team members understand the codebase faster, and scales gracefully as your project grows from prototype to production. This clear structure also makes your codebase ideal for AI coding assistants. Tools like Claude Code, Cursor, GitHub Copilot, Windsurf, and Cline thrive when they can understand where things belong, and Trabuco's organized layout removes the guesswork. The CLI can generate context files for your preferred AI coding agents with project-specific conventions, patterns, and commands — giving them the context they need to write code that fits naturally into your project.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Migrating Legacy Projects](#migrating-legacy-projects)
  - [How It Works](#how-it-works)
  - [Authentication](#authentication)
  - [Migration Options](#migration-options)
  - [Checkpoints and Recovery](#checkpoints-and-recovery)
  - [What Gets Migrated](#what-gets-migrated)
- [Managing Existing Projects](#managing-existing-projects)
  - [Project Health Check](#project-health-check)
  - [Adding Modules](#adding-modules)
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
  - [MCP](#mcp)
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
- [Tech Stack](#tech-stack)
- [Local Development](#local-development)
- [Requirements](#requirements)
- [Contributing](#contributing)
- [License](#license)

## Features

- **AI-powered migration** *(experimental)* — Transform existing Spring Boot projects into Trabuco's architecture with `trabuco migrate`
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
- **AI-friendly** — Generates context files for Claude, Cursor, GitHub Copilot, Windsurf, and Cline
- **MCP server** — Optional Model Context Protocol server for AI tool integration in generated projects
- **CLI MCP server** — `trabuco mcp` exposes all CLI functionality as structured tools for AI coding agents

## Installation

### npx (recommended for MCP use)

No installation needed — just reference `trabuco-mcp` in your AI agent config and npx handles everything:

```bash
claude mcp add trabuco -- npx -y trabuco-mcp
```

See [CLI MCP Server > Configuration](#configuration) for all agent configs.

### From GitHub

```bash
curl -sSL https://raw.githubusercontent.com/arianlopezc/Trabuco/main/scripts/install.sh | bash
```

### Using Go

```bash
go install github.com/arianlopezc/Trabuco/cmd/trabuco@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your PATH.

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

# Add MCP server for AI tool integration
trabuco init --name=myapp --group-id=com.company.myapp \
  --modules=Model,SQLDatastore,Shared,API,MCP --database=postgresql
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

## Migrating Legacy Projects

> **⚠️ Experimental Feature**
>
> The `trabuco migrate` command is under active development. While functional, results should be reviewed and may require manual adjustments. We recommend running with `--dry-run` first and testing thoroughly before relying on migrated code in production.
>
> Set `TRABUCO_ACKNOWLEDGE_EXPERIMENTAL=true` to suppress the experimental warning in CI/CD pipelines.

The `trabuco migrate` command uses AI to analyze existing Spring Boot projects and transform them into Trabuco's multi-module architecture. It's designed for projects that have grown organically and need restructuring — or for teams adopting Trabuco's patterns for existing codebases.

```bash
# Basic migration
trabuco migrate /path/to/legacy-app

# Specify output directory
trabuco migrate /path/to/legacy-app -o ./migrated-app

# Dry run — analyze without generating files
trabuco migrate --dry-run /path/to/legacy-app
```

### How It Works

Migration happens in 10 stages, each with checkpoints for recovery:

| Stage | Description |
|-------|-------------|
| 1. Discovery | Scans your project structure, parses `pom.xml`, categorizes Java classes |
| 2. Dependency Analysis | Identifies compatible, replaceable, and unsupported dependencies |
| 3. Entity Extraction | Converts JPA entities to Spring Data JDBC format with Flyway migrations |
| 4. Repository Migration | Transforms repositories to Spring Data JDBC/MongoDB patterns |
| 5. Service Extraction | Moves services to Shared module with constructor injection |
| 6. Controller Migration | Restructures REST controllers for the API module |
| 7. Jobs Migration | Converts `@Scheduled` and Quartz jobs to JobRunr format |
| 8. Events Migration | Transforms event listeners to Kafka/RabbitMQ patterns |
| 9. Configuration | Generates `docker-compose.yml`, `.env.example`, AI agent files |
| 10. Final Assembly | Creates parent POM, README, and project metadata |

Before making changes, the AI shows you what it found and estimates the cost. In interactive mode (default), you confirm each major decision.

### Authentication

Trabuco uses Claude (Anthropic's AI) for intelligent code transformation. You need an API key:

**Option 1: Anthropic (recommended)**
```bash
export ANTHROPIC_API_KEY=sk-ant-...
trabuco migrate /path/to/legacy-app
```

**Option 2: OpenRouter (alternative)**

[OpenRouter](https://openrouter.ai) provides access to Claude models without an Anthropic account:

```bash
export OPENROUTER_API_KEY=sk-or-...
trabuco migrate --provider=openrouter /path/to/legacy-app
```

You can also pass the key directly:
```bash
trabuco migrate --api-key=sk-ant-... /path/to/legacy-app
```

### Migration Options

| Option | Description | Default |
|--------|-------------|---------|
| `-o, --output` | Output directory | `./<project-name>-trabuco` |
| `--dry-run` | Analyze only, don't generate files | `false` |
| `--interactive` | Confirm each major decision | `true` |
| `--resume` | Resume from last checkpoint | `false` |
| `--rollback` | Rollback migration completely | `false` |
| `--rollback-to` | Rollback to specific stage | — |
| `--provider` | AI provider: `anthropic`, `openrouter` | `anthropic` |
| `--model` | Model to use | `claude-sonnet-4-5` |
| `--include-tests` | Migrate test files | `false` |
| `--skip-build` | Skip Maven build after migration | `false` |
| `-v, --verbose` | Verbose output | `false` |
| `--debug` | Save all AI interactions | `false` |

### Checkpoints and Recovery

Migration creates checkpoints in `.trabuco-migrate/` within your source project. If something goes wrong:

```bash
# Resume from where you left off
trabuco migrate --resume /path/to/legacy-app

# Rollback to a specific stage
trabuco migrate --rollback-to=entities /path/to/legacy-app

# Rollback completely (removes output directory)
trabuco migrate --rollback /path/to/legacy-app
```

Checkpoints track:
- Completed stages and their outputs
- AI decisions made (for transparency)
- Token usage and estimated cost
- Errors encountered

### What Gets Migrated

**Automatically converted:**

| From | To |
|------|----|
| `@Entity` (JPA) | `@Table` (Spring Data JDBC) |
| `@OneToMany`, `@ManyToOne` | Explicit foreign key fields |
| Lombok `@Data`, `@Getter`, `@Setter` | Explicit getters/setters |
| `JpaRepository` | `CrudRepository` (JDBC) |
| `@Scheduled` methods | JobRunr `JobRequest` + `JobRequestHandler` |
| Quartz jobs | JobRunr jobs |
| `@KafkaListener`, `@RabbitListener` | Trabuco EventConsumer patterns |

**Dependency replacements:**

| Legacy | Trabuco Alternative |
|--------|---------------------|
| Hibernate / JPA | Spring Data JDBC |
| Quartz Scheduler | JobRunr |
| Lombok | Explicit code or Immutables |
| Spring Data JPA | Spring Data JDBC |
| javax.* | jakarta.* (Spring Boot 3.x) |

**Requires manual attention:**
- Complex JPA relationships (inheritance, embedded collections)
- Native SQL queries with JPA-specific syntax
- Custom Hibernate types
- AspectJ weaving

The AI flags these during migration and explains what needs manual review.

### Cost Transparency

Before migration begins, Trabuco estimates the cost based on:
- Number of files to process
- Estimated input/output tokens
- Current model pricing

Example output:
```
Estimated Migration Cost:
  Files to process: 47
  Est. tokens:      ~70,500
  Est. cost:        $0.35 - $0.65
```

Actual cost is tracked during migration and saved in the checkpoint.

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
- Updates `.trabuco.json` metadata
- Auto-includes dependent modules (e.g., `Worker` includes `Jobs`)

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

**Windsurf** — add to `~/.codeium/windsurf/mcp_config.json`:

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
| `init_project` | Generate a new Java project with specified modules, database, and options |
| `add_module` | Add a module to an existing Trabuco project (with dry-run support) |
| `run_doctor` | Run health checks on a project and optionally auto-fix issues |
| `get_project_info` | Read project metadata from `.trabuco.json` or inferred from POM |
| `list_modules` | List all available modules with descriptions and dependency info |
| `check_docker` | Check if Docker is installed and running |
| `get_version` | Get the Trabuco CLI version |
| `scan_project` | Analyze a legacy Java project's structure and dependencies (no AI required) |
| `migrate_project` | Full AI-powered migration of a legacy project (long-running) |
| `auth_status` | Check which AI providers have credentials configured |
| `list_providers` | List supported AI providers with pricing and model info |

**What this looks like in practice:** Ask your AI agent "create a new Java project called order-service with PostgreSQL and Kafka" and it calls `init_project` with the right parameters, gets back structured JSON with the project path, resolved modules, and any warnings — no terminal output to parse.

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
├── MCP/                             # MCP server (if selected)
│   └── src/main/
│       ├── java/.../mcp/
│       │   ├── McpServerApplication.java  # MCP server entry point
│       │   └── tools/               # Build, test, and project tools
│       └── resources/
│           └── logback.xml          # MCP logging configuration
├── docker-compose.yml               # Local dev stack (database, message broker)
├── .run/                            # IntelliJ run configurations
├── .mcp.json                        # MCP config for Claude Code (if MCP selected)
├── .cursor/mcp.json                 # MCP config for Cursor (if MCP selected)
├── .vscode/mcp.json                 # MCP config for VS Code/Copilot (if MCP selected)
├── CLAUDE.md                        # AI assistant context
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

### MCP

Model Context Protocol server — provides AI coding assistants with tools for building, testing, and introspecting the project.

| What | Description |
|------|-------------|
| **Build tools** | `build`, `package`, `clean` — Maven build operations |
| **Test tools** | `test`, `test-single` — Run tests or specific test classes |
| **Project tools** | `list-modules`, `list-entities`, `get-config`, `project-info` — Project introspection |

The MCP server is a standalone Java application that communicates via STDIO with AI coding assistants that support the Model Context Protocol.

**Building the MCP server:**

```bash
mvn package -pl MCP -am -DskipTests
```

The executable JAR is created at `MCP/target/MCP-1.0-SNAPSHOT.jar`.

**Auto-configured for multiple agents:**

When you select the MCP module, Trabuco generates configuration files for all major AI coding assistants:

| Agent | Config File | Auto-Detected |
|-------|-------------|---------------|
| Claude Code | `.mcp.json` | ✅ Yes — approve on first use |
| Cursor | `.cursor/mcp.json` | ✅ Yes — approve on first use |
| VS Code / GitHub Copilot | `.vscode/mcp.json` | ✅ Yes — click "Start" on first use |
| Windsurf | See `MCP/README.md` | ❌ Manual setup required |
| Cline | See `MCP/README.md` | ❌ Manual setup required |

For agents with project-local configs (Claude Code, Cursor, VS Code), just open the project and approve the server when prompted. For detailed setup instructions for all agents, see `MCP/README.md`.

**Available MCP tools:**

| Tool | Description |
|------|-------------|
| `build` | Compile the project or a specific module |
| `package` | Package all modules into JAR files |
| `clean` | Clean all build artifacts |
| `test` | Run all tests or tests for a specific module |
| `test-single` | Run a specific test class or method |
| `list-modules` | List all Maven modules in the project |
| `list-entities` | List all entity classes in the Model module |
| `get-config` | Get the application.yml for a module |
| `project-info` | Get project metadata from .trabuco.json |

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
| `--java-version` | Java version: `17`, `21`, or `25` | `21` |
| `--ai-agents` | AI coding agents (comma-separated): `claude`, `cursor`, `copilot`, `windsurf`, `cline` | — |
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
| `MCP` | Model Context Protocol server for AI tools | Model |

**Notes:**
- SQLDatastore and NoSQLDatastore are mutually exclusive
- Worker uses your datastore for job persistence (defaults to PostgreSQL if none selected)
- Jobs module is auto-included when Worker is selected (not shown in CLI)
- Events module is auto-included when EventConsumer is selected (not shown in CLI)
- MCP generates configuration files for Claude Code, Cursor, VS Code, Windsurf, and Cline

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

Trabuco can generate context files for popular AI coding assistants. These files contain project-specific conventions, commands, and patterns that help AI tools write code that fits naturally into your project.

| Agent | Context File | Description |
|-------|--------------|-------------|
| Claude Code | `CLAUDE.md` | Anthropic's CLI for Claude |
| Cursor | `.cursorrules` | AI-first code editor |
| GitHub Copilot | `.github/copilot-instructions.md` | GitHub's AI pair programmer |
| Windsurf | `.windsurf/rules/project.md` | Codeium's agentic IDE |
| Cline | `.clinerules/project.md` | VS Code autonomous agent |

In interactive mode, you'll be prompted to select which agents you want context files for. In non-interactive mode:

```bash
# Generate for specific agents
trabuco init --name=myapp --group-id=com.example --modules=Model,API --ai-agents=claude,cursor

# Generate for all agents
trabuco init --name=myapp --group-id=com.example --modules=Model,API --ai-agents=claude,cursor,copilot,windsurf,cline
```

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Java | 17, 21, or 25 | Runtime |
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
| Resilience4j | — | Circuit breakers |
| PostgreSQL / MySQL | — | SQL databases |
| MongoDB / Redis | — | NoSQL databases |
| Apache Kafka | — | Distributed streaming |
| RabbitMQ | — | Message broker |
| AWS SQS | — | Managed queue service (via LocalStack for local dev) |
| GCP Pub/Sub | — | Google Cloud messaging (via emulator for local dev) |
| HikariCP | — | Connection pooling (SQL) |

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

- **Java 17+** (17, 21, or 25 — Trabuco auto-detects installed versions)
- **Maven 3.8+**
- **Docker** (for Testcontainers and local development)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License — do whatever you want with it.

---

**Built with Trabuco.** Now go build something amazing.
