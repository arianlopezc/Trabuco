# Trabuco

**Generate production-ready Java projects in seconds.**

Starting a new Java project shouldn't feel like a chore. Yet every time you begin, it's the same story — setting up Maven modules, configuring database connections, writing migration scripts, wiring up test infrastructure, and copying boilerplate from that one project that "mostly works." Hours pass before you write a single line of actual business logic. Trabuco exists because your time is better spent building features, not fighting configuration files.

Trabuco is a command-line tool that generates complete, well-structured Java projects with a single command. Run `trabuco init`, answer a few questions (or pass flags for automation), and you get a fully functional multi-module Maven project ready for development. No templates to download, no manual setup, no guessing how things should connect. The CLI handles the tedious work so you can focus on what matters — your application's unique value.

The generated projects come batteries-included with production-proven technologies: Spring Boot for the application framework, Spring Data JDBC for straightforward database access, Flyway for version-controlled migrations, Testcontainers for realistic integration tests, and Resilience4j for fault tolerance. Everything is pre-configured and working together out of the box. Need PostgreSQL instead of MySQL? Just pick it during setup. Want the latest Java 25 instead of 21? One flag changes everything. The architecture is designed to be solid by default yet flexible when you need it.

The real power lies in the modular structure. Instead of a monolithic source tree where everything depends on everything, Trabuco generates clean, separated modules: Model for your data structures, SQLDatastore for persistence, Shared for business logic and services, and API for your REST endpoints. Each module has a clear responsibility and well-defined dependencies. This isn't just organization for organization's sake — it enforces good architecture, makes testing straightforward, helps new team members understand the codebase faster, and scales gracefully as your project grows from prototype to production. This clear structure also makes your codebase ideal for AI coding assistants. Tools like Claude Code thrive when they can understand where things belong, and Trabuco's organized layout removes the guesswork. The CLI even generates a `CLAUDE.md` file with project-specific conventions, patterns, and commands — giving AI assistants the context they need to write code that fits naturally into your project.

## Features

- **Multi-module Maven structure** — Clean separation between Model, Data, Services, and API
- **Immutables everywhere** — Type-safe, immutable DTOs and entities with builder pattern
- **Spring Boot 3.4** — Latest LTS with Spring Data JDBC (not JPA — no magic, no surprises)
- **Database ready** — PostgreSQL/MySQL support with Flyway migrations out of the box
- **Testcontainers 2.x** — Real database tests that actually work with Docker Desktop
- **Circuit breakers** — Resilience4j configured and ready to use
- **Docker Compose** — Local development stack included
- **IntelliJ run configs** — Just open and run
- **AI-friendly** — Generates `CLAUDE.md` with project conventions for AI assistants

## Installation

### From GitHub (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/trabuco/trabuco/main/scripts/install.sh | bash
```

### Using Go

```bash
go install github.com/trabuco/trabuco/cmd/trabuco@latest
```

Make sure `$GOPATH/bin` (usually `~/go/bin`) is in your PATH.

## Quick Start

### Interactive mode

```bash
trabuco init
```

Answer a few questions, and you're done.

### Non-interactive mode

```bash
trabuco init \
  --name=myapp \
  --group-id=com.company.myapp \
  --modules=Model,SQLDatastore,Shared,API \
  --database=postgresql
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

## Generated Project Structure

```
myapp/
├── pom.xml                          # Parent POM (module aggregator)
├── Model/                           # DTOs, Entities, Enums
│   └── src/main/java/.../model/
│       ├── dto/                     # Request/Response DTOs
│       ├── entities/                # Domain entities + DB records
│       └── ImmutableStyle.java      # Immutables configuration
├── SQLDatastore/                    # Database layer
│   └── src/
│       ├── main/
│       │   ├── java/.../sqldatastore/
│       │   │   ├── config/          # Database configuration
│       │   │   └── repository/      # Spring Data repositories
│       │   └── resources/
│       │       └── db/migration/    # Flyway SQL migrations
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
├── docker-compose.yml               # Local dev stack (Postgres, etc.)
├── .run/                            # IntelliJ run configurations
├── CLAUDE.md                        # AI assistant context
└── README.md                        # Project documentation
```

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

## Configuration Options

| Option | Description | Default |
|--------|-------------|---------|
| `--name` | Project name (lowercase, hyphens allowed) | — |
| `--group-id` | Maven group ID (e.g., `com.company.project`) | — |
| `--modules` | Modules to include (comma-separated) | — |
| `--database` | Database type: `postgresql`, `mysql`, `none` | `postgresql` |
| `--java-version` | Java version: `21` or `25` | `21` |
| `--include-claude` | Generate `CLAUDE.md` for AI assistants | `true` |

### Available Modules

| Module | Description | Dependencies |
|--------|-------------|--------------|
| `Model` | DTOs, Entities, Enums | None (always included) |
| `SQLDatastore` | Repositories, Migrations | Model |
| `Shared` | Services, Circuit breakers | Model, SQLDatastore |
| `API` | REST endpoints | Model, SQLDatastore, Shared |

## Tech Stack

| Technology | Version | Purpose |
|------------|---------|---------|
| Java | 21 or 25 | Runtime |
| Spring Boot | 3.4.2 | Application framework |
| Spring Data JDBC | — | Database access |
| Immutables | 2.10.1 | Immutable value objects |
| Flyway | — | Database migrations |
| Testcontainers | 2.0.3 | Integration testing |
| Resilience4j | — | Circuit breakers |
| PostgreSQL / MySQL | — | Database |
| HikariCP | — | Connection pooling |

## Local Development

The generated project includes a `docker-compose.yml` for local development:

```bash
docker-compose up -d    # Start Postgres
mvn spring-boot:run -pl API
```

### Running Tests

```bash
mvn test                # All tests
mvn test -pl Model      # Single module
```

**Note:** SQLDatastore tests require Docker to be running (Testcontainers).

## Requirements

- **Java 21** (or 25 if specified during init)
- **Maven 3.8+**
- **Docker** (for Testcontainers and local development)

## Contributing

Contributions are welcome! Please feel free to submit issues and pull requests.

## License

MIT License — do whatever you want with it.

---

**Built with Trabuco.** Now go build something amazing.
