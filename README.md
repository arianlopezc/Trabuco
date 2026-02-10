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

<h3 align="center">Generate production-ready Java projects in seconds.</h3>

Starting a new Java project shouldn't feel like a chore. Yet every time you begin, it's the same story — setting up Maven modules, configuring database connections, writing migration scripts, wiring up test infrastructure, and copying boilerplate from that one project that "mostly works." Hours pass before you write a single line of actual business logic. Trabuco exists because your time is better spent building features, not fighting configuration files.

Trabuco is a command-line tool that generates complete, well-structured Java projects with a single command. Run `trabuco init`, answer a few questions (or pass flags for automation), and you get a fully functional multi-module Maven project ready for development. No templates to download, no manual setup, no guessing how things should connect. The CLI handles the tedious work so you can focus on what matters — your application's unique value.

The generated projects come batteries-included with production-proven technologies: Spring Boot for the application framework, Spring Data JDBC for straightforward database access, Flyway for version-controlled migrations, Testcontainers for realistic integration tests, and Resilience4j for fault tolerance. Everything is pre-configured and working together out of the box. Need PostgreSQL instead of MySQL? Just pick it during setup. Want the latest Java 25 instead of 21? One flag changes everything. The architecture is designed to be solid by default yet flexible when you need it.

The real power lies in the modular structure. Instead of a monolithic source tree where everything depends on everything, Trabuco generates clean, separated modules: Model for your data structures, SQLDatastore or NoSQLDatastore for persistence, Shared for business logic and services, and API for your REST endpoints. Each module has a clear responsibility and well-defined dependencies. This isn't just organization for organization's sake — it enforces good architecture, makes testing straightforward, helps new team members understand the codebase faster, and scales gracefully as your project grows from prototype to production. This clear structure also makes your codebase ideal for AI coding assistants. Tools like Claude Code, Cursor, GitHub Copilot, Windsurf, and Cline thrive when they can understand where things belong, and Trabuco's organized layout removes the guesswork. The CLI can generate context files for your preferred AI coding agents with project-specific conventions, patterns, and commands — giving them the context they need to write code that fits naturally into your project.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Managing Existing Projects](#managing-existing-projects)
  - [Project Health Check](#project-health-check)
  - [Adding Modules](#adding-modules)
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

## Installation

### From GitHub (recommended)

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
├── docker-compose.yml               # Local dev stack (database, message broker)
├── .run/                            # IntelliJ run configurations
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

Trabuco can generate context files for popular AI coding assistants. These files contain project-specific conventions, commands, and patterns that help AI tools write code that fits naturally into your project.

| Agent | Context File | Description |
|-------|--------------|-------------|
| Claude Code | `CLAUDE.md` | Anthropic's CLI for Claude |
| Cursor | `.cursorrules` | AI-first code editor |
| GitHub Copilot | `.github/copilot-instructions.md` | GitHub's AI pair programmer |
| Windsurf | `.windsurfrules` | Codeium's agentic IDE |
| Cline | `.clinerules` | VS Code autonomous agent |

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
