# Trabuco Implementation Plan

> **Status:** Planning Phase
> **Location:** `/Users/arianlc/Documents/Work/Trabuco`
> **Last Updated:** 2026-01-29

---

## Table of Contents

1. [Research & Technical Decisions](#research--technical-decisions)
2. [Stage 1: Project Initialization](#stage-1-project-initialization)
3. [Stage 2: CLI Framework Setup](#stage-2-cli-framework-setup)
4. [Stage 3: Configuration & Module Definitions](#stage-3-configuration--module-definitions)
5. [Stage 4: Interactive Prompts](#stage-4-interactive-prompts)
6. [Stage 5: Template System](#stage-5-template-system)
7. [Stage 6: POM Templates](#stage-6-pom-templates)
8. [Stage 7: Java Source Templates](#stage-7-java-source-templates)
9. [Stage 8: Infrastructure Templates](#stage-8-infrastructure-templates)
10. [Stage 9: Generator Engine](#stage-9-generator-engine)
11. [Stage 10: Testing & Validation](#stage-10-testing--validation)
12. [Stage 11: Distribution](#stage-11-distribution)

---

## Research & Technical Decisions

> **Research Date:** 2026-01-29
> **Sources:** Oracle Java SE Roadmap, Spring Boot Documentation, Baeldung, Official Library Documentation

### 1. Java Version

| Version | Type | Premier Support | Extended Support | Recommendation |
|---------|------|-----------------|------------------|----------------|
| Java 21 | LTS | Sept 2028 | Sept 2031 | **Default choice** |
| Java 25 | LTS | Sept 2030 | Sept 2033 | Latest LTS (Sept 2025) |

**Decision:** Use **Java 21** as default. Wamefy uses Java 21. It's stable, widely adopted, and has long support.

**Sources:**
- [Oracle Java SE Support Roadmap](https://www.oracle.com/java/technologies/java-se-support-roadmap.html)
- [JDK Releases](https://www.java.com/releases/)

---

### 2. Module Naming

| Original Name | New Name | Reason |
|---------------|----------|--------|
| Data | **SQLDatastore** | More specific; clarifies it's for SQL databases |

**Module Registry (Updated):**

| Module | Required | Dependencies |
|--------|----------|--------------|
| Model | Yes | None |
| SQLDatastore | No | Model |
| Shared | No | Model, SQLDatastore |
| API | No | Model, SQLDatastore, Shared |

---

### 3. Database Dependencies (Production-Verified)

#### PostgreSQL Stack

| Dependency | Version | Purpose | Source |
|------------|---------|---------|--------|
| `org.postgresql:postgresql` | 42.7.8 | JDBC Driver | [PostgreSQL JDBC](https://jdbc.postgresql.org/) |
| `com.zaxxer:HikariCP` | 5.1.0 | Connection Pool | [HikariCP GitHub](https://github.com/brettwooldridge/HikariCP) |
| `org.springframework.boot:spring-boot-starter-data-jdbc` | 3.4.x | Spring Data JDBC | [Spring Docs](https://docs.spring.io/spring-boot/reference/data/sql.html) |
| `org.flywaydb:flyway-core` | 10.22.0 | Migrations | [Flyway](https://flywaydb.org/) |
| `org.flywaydb:flyway-database-postgresql` | 10.22.0 | Postgres support | [Flyway](https://flywaydb.org/) |

#### MySQL Stack

| Dependency | Version | Purpose | Source |
|------------|---------|---------|--------|
| `com.mysql:mysql-connector-j` | 9.1.0 | JDBC Driver | [MySQL Connector/J](https://dev.mysql.com/doc/connector-j/en/) |
| `com.zaxxer:HikariCP` | 5.1.0 | Connection Pool | [HikariCP GitHub](https://github.com/brettwooldridge/HikariCP) |
| `org.springframework.boot:spring-boot-starter-data-jdbc` | 3.4.x | Spring Data JDBC | [Spring Docs](https://docs.spring.io/spring-boot/reference/data/sql.html) |
| `org.flywaydb:flyway-core` | 10.22.0 | Migrations | [Flyway](https://flywaydb.org/) |
| `org.flywaydb:flyway-mysql` | 10.22.0 | MySQL support | [Flyway](https://flywaydb.org/) |

#### Generic SQL Stack (Other)

| Dependency | Version | Purpose | Source |
|------------|---------|---------|--------|
| `com.zaxxer:HikariCP` | 5.1.0 | Connection Pool | [HikariCP GitHub](https://github.com/brettwooldridge/HikariCP) |
| `org.springframework.boot:spring-boot-starter-data-jdbc` | 3.4.x | Spring Data JDBC | [Spring Docs](https://docs.spring.io/spring-boot/reference/data/sql.html) |
| `org.flywaydb:flyway-core` | 10.22.0 | Migrations | [Flyway](https://flywaydb.org/) |

> **Note:** Generic requires user to add their specific JDBC driver manually.

**Why Spring Data JDBC over JPA?**
- 4x faster for bulk operations ([Source](https://medium.com/@mesfandiari77/spring-data-jdbc-vs-jpa-why-simplicity-is-winning-over-hibernate-complexity-in-2025-59eba2661cb0))
- No proxies, no lazy loading complexity
- Simpler mental model (what you load is what you get)
- Better suited for microservices and bounded contexts
- Wamefy uses Spring Data JDBC - proven in production

**Why HikariCP?**
- Default in Spring Boot
- "Zero-overhead" production-ready pool
- Consistently outperforms alternatives (DBCP, C3P0, Tomcat JDBC)
- 165KB footprint

**Why Flyway over Liquibase?**
- Simpler (SQL-first approach)
- Lower learning curve
- Wamefy uses Flyway - proven in production
- Sufficient for most projects

---

### 4. Validation Dependencies (API Module)

| Dependency | Purpose | Annotations |
|------------|---------|-------------|
| `spring-boot-starter-validation` | Bean validation | All Jakarta validation |
| `jakarta.validation:jakarta.validation-api` | API (in Model) | Annotation definitions |

**Production-Approved Annotations:**

| Annotation | Package | Usage |
|------------|---------|-------|
| `@NotNull` | jakarta.validation.constraints | Non-null values |
| `@NotBlank` | jakarta.validation.constraints | Non-blank strings |
| `@NotEmpty` | jakarta.validation.constraints | Non-empty collections |
| `@Size(min, max)` | jakarta.validation.constraints | String/collection size |
| `@Min`, `@Max` | jakarta.validation.constraints | Numeric range |
| `@Pattern` | jakarta.validation.constraints | Regex validation |
| `@Email` | jakarta.validation.constraints | Email format |
| `@Valid` | jakarta.validation | Cascade validation |
| `@Validated` | org.springframework.validation.annotation | Class-level validation |

**Sources:**
- [Spring Boot Validation Docs](https://docs.spring.io/spring-boot/reference/io/validation.html)
- [Baeldung - Validation in Spring Boot](https://www.baeldung.com/spring-boot-bean-validation)

---

### 5. Circuit Breaker (Shared Module)

| Dependency | Version | Purpose |
|------------|---------|---------|
| `io.github.resilience4j:resilience4j-spring-boot3` | 2.2.0 | Circuit breaker, retry, bulkhead |
| `org.springframework.boot:spring-boot-starter-actuator` | 3.4.x | Health indicators |
| `org.springframework.boot:spring-boot-starter-aop` | 3.4.x | Required for annotations |

**Recommended Configuration:**

```yaml
resilience4j:
  circuitbreaker:
    instances:
      default:
        registerHealthIndicator: true
        slidingWindowSize: 10
        minimumNumberOfCalls: 5
        permittedNumberOfCallsInHalfOpenState: 3
        waitDurationInOpenState: 30s
        failureRateThreshold: 50
        automaticTransitionFromOpenToHalfOpenEnabled: true
```

**Why Resilience4j?**
- Designed for Java 8+ functional programming
- Lightweight (no Hystrix baggage)
- Spring Boot 3 native support
- Actuator integration for monitoring
- Wamefy uses Resilience4j - proven in production

**Sources:**
- [Resilience4j Docs](https://resilience4j.readme.io/docs/getting-started-3)
- [Baeldung - Resilience4j with Spring Boot](https://www.baeldung.com/spring-boot-resilience4j)

---

### 6. Immutables Configuration

| Dependency | Version | Scope |
|------------|---------|-------|
| `org.immutables:value` | 2.10.1 | provided |

**ImmutableStyle Configuration:**

```java
@Target({ElementType.PACKAGE, ElementType.TYPE})
@Retention(RetentionPolicy.CLASS)
@Value.Style(
    passAnnotations = {Id.class, Table.class},  // Spring Data JDBC annotations
    visibility = Value.Style.ImplementationVisibility.PUBLIC
)
public @interface ImmutableStyle {}
```

**Usage Pattern (with Jackson serialization):**

```java
@Value.Immutable
@ImmutableStyle
@JsonSerialize(as = ImmutablePlaceholder.class)
@JsonDeserialize(as = ImmutablePlaceholder.class)
public interface Placeholder {
    @NotBlank
    String name();
}
```

**Rules:**
- All DTOs, entities, value objects use `@Value.Immutable` + `@ImmutableStyle`
- All use `@JsonSerialize` and `@JsonDeserialize` for Jackson
- **Enums do NOT use Immutables** (they're already immutable by nature)

**Sources:**
- [Immutables Documentation](https://immutables.github.io/)
- [Immutables JSON Serialization](https://immutables.github.io/json.html)

---

### 7. Model Module Package Structure

```
Model/src/main/java/{{package}}/model/
├── ImmutableStyle.java          # Immutables style configuration
├── entities/                    # Database entities (with @Id, @Table)
│   └── Placeholder.java         # Example entity
├── dto/                         # Request/Response DTOs
│   ├── PlaceholderRequest.java  # Example request DTO
│   └── PlaceholderResponse.java # Example response DTO
├── enums/                       # Domain enumerations (NO Immutables)
├── exception/                   # Custom exceptions
├── util/                        # Utility classes
├── events/                      # Domain events (optional)
└── validation/                  # Custom validators (optional)
```

**Package Purposes:**

| Package | Purpose | Uses Immutables? |
|---------|---------|------------------|
| `entities/` | Database-mapped objects | Yes |
| `dto/` | API request/response objects | Yes |
| `enums/` | Domain enumerations | **No** |
| `exception/` | Custom exception classes | No |
| `util/` | Helper/utility classes | No |
| `events/` | Domain event definitions | Yes |
| `validation/` | Custom validation logic | No |

---

### 8. Placeholder Entity - CRUD Example

The generated project includes a complete CRUD example using a `Placeholder` entity.

#### Model Module

**Placeholder.java** (Entity):
```java
@Value.Immutable
@ImmutableStyle
@JsonSerialize(as = ImmutablePlaceholder.class)
@JsonDeserialize(as = ImmutablePlaceholder.class)
@Table("placeholders")
public interface Placeholder {
    @Id
    @Nullable
    Long id();

    @NotBlank
    String name();
}
```

**PlaceholderRequest.java** (DTO):
```java
@Value.Immutable
@ImmutableStyle
@JsonSerialize(as = ImmutablePlaceholderRequest.class)
@JsonDeserialize(as = ImmutablePlaceholderRequest.class)
public interface PlaceholderRequest {
    @NotBlank
    @Size(min = 1, max = 255)
    String name();
}
```

#### SQLDatastore Module

**PlaceholderRepository.java**:
```java
@Repository
public interface PlaceholderRepository extends CrudRepository<ImmutablePlaceholder, Long> {
    // Spring Data JDBC provides: save, findById, findAll, deleteById
}
```

**V1__baseline.sql**:
```sql
CREATE TABLE placeholders (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);
```

**Integration Tests** (4 tests):
```java
@SpringBootTest
@Transactional
class PlaceholderRepositoryTest {
    @Test void shouldCreatePlaceholder() { ... }
    @Test void shouldReadPlaceholder() { ... }
    @Test void shouldUpdatePlaceholder() { ... }
    @Test void shouldDeletePlaceholder() { ... }
}
```

#### Shared Module

**PlaceholderService.java**:
```java
@Service
public class PlaceholderService {
    private final PlaceholderRepository repository;

    public ImmutablePlaceholder create(PlaceholderRequest request) { ... }
    public Optional<ImmutablePlaceholder> findById(Long id) { ... }
    public ImmutablePlaceholder update(Long id, PlaceholderRequest request) { ... }
    public void delete(Long id) { ... }
}
```

**Unit Tests** (4 tests - mocked repository):
```java
@ExtendWith(MockitoExtension.class)
class PlaceholderServiceTest {
    @Test void shouldCreatePlaceholder() { ... }
    @Test void shouldFindPlaceholderById() { ... }
    @Test void shouldUpdatePlaceholder() { ... }
    @Test void shouldDeletePlaceholder() { ... }
}
```

#### API Module

**PlaceholderController.java**:
```java
@RestController
@RequestMapping("/api/placeholders")
@Validated
public class PlaceholderController {
    private final PlaceholderService service;

    @PostMapping
    public ResponseEntity<ImmutablePlaceholder> create(@Valid @RequestBody ImmutablePlaceholderRequest request) { ... }

    @GetMapping("/{id}")
    public ResponseEntity<ImmutablePlaceholder> findById(@PathVariable Long id) { ... }

    @PutMapping("/{id}")
    public ResponseEntity<ImmutablePlaceholder> update(@PathVariable Long id, @Valid @RequestBody ImmutablePlaceholderRequest request) { ... }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> delete(@PathVariable Long id) { ... }
}
```

**Test Summary:**

| Module | Test Type | Count | Purpose |
|--------|-----------|-------|---------|
| SQLDatastore | Integration | 4 | Repository CRUD with real DB |
| Shared | Unit | 4 | Service logic with mocked repo |
| API | None | 0 | No additional tests (as specified) |

---

### 9. Dependency Summary by Module

#### Parent POM (Dependency Management)

```xml
<properties>
    <java.version>21</java.version>
    <spring-boot.version>3.4.2</spring-boot.version>
    <immutables.version>2.10.1</immutables.version>
    <resilience4j.version>2.2.0</resilience4j.version>
    <flyway.version>10.22.0</flyway.version>
    <hikaricp.version>5.1.0</hikaricp.version>
</properties>
```

#### Model Module

| Dependency | Purpose |
|------------|---------|
| `org.immutables:value` (provided) | Immutable generation |
| `jakarta.validation:jakarta.validation-api` | Validation annotations |
| `com.fasterxml.jackson.core:jackson-databind` | JSON serialization |
| `org.springframework.data:spring-data-commons` | @Id, @Table annotations |

#### SQLDatastore Module

| Dependency | Purpose |
|------------|---------|
| Model (internal) | Entity definitions |
| `spring-boot-starter-data-jdbc` | Repository support |
| `HikariCP` | Connection pooling |
| `postgresql` OR `mysql-connector-j` | JDBC driver |
| `flyway-core` + DB-specific | Migrations |

#### Shared Module

| Dependency | Purpose |
|------------|---------|
| Model (internal) | Entity definitions |
| SQLDatastore (internal) | Repository access |
| `spring-boot-starter` | Core Spring |
| `resilience4j-spring-boot3` | Circuit breaker |
| `spring-boot-starter-actuator` | Health/metrics |
| `spring-boot-starter-aop` | Resilience4j annotations |

#### API Module

| Dependency | Purpose |
|------------|---------|
| Model (internal) | DTO definitions |
| SQLDatastore (internal) | Repository access |
| Shared (internal) | Services |
| `spring-boot-starter-web` | REST endpoints |
| `spring-boot-starter-validation` | Request validation |

> **Note:** API module has NO direct database dependencies, NO circuit breaker. Only web + validation.

---

## Stage 1: Project Initialization

**Goal:** Set up the Go module and basic project structure.

### Tasks

- [ ] 1.1 Initialize Go module
  ```bash
  go mod init github.com/your-org/trabuco
  ```

- [ ] 1.2 Create directory structure
  ```
  trabuco/
  ├── cmd/trabuco/
  ├── internal/
  │   ├── cli/
  │   ├── config/
  │   ├── generator/
  │   ├── prompts/
  │   └── templates/
  └── templates/
      ├── pom/
      ├── java/
      ├── config/
      ├── docker/
      └── github/
  ```

- [ ] 1.3 Add dependencies to `go.mod`
  - `github.com/spf13/cobra` - CLI framework
  - `github.com/AlecAivazis/survey/v2` - Interactive prompts
  - `github.com/fatih/color` - Colored output

### Files to Create

| File | Description |
|------|-------------|
| `go.mod` | Go module definition |
| `go.sum` | Dependency checksums (auto-generated) |
| `Makefile` | Build automation |
| `.gitignore` | Git ignore patterns |

### Acceptance Criteria

- [ ] `go mod tidy` runs without errors
- [ ] Directory structure matches plan
- [ ] Dependencies are resolved

---

## Stage 2: CLI Framework Setup

**Goal:** Implement the cobra CLI structure with root, init, and version commands.

### Tasks

- [ ] 2.1 Create entry point (`cmd/trabuco/main.go`)
- [ ] 2.2 Implement root command (`internal/cli/root.go`)
- [ ] 2.3 Implement version command (`internal/cli/version.go`)
- [ ] 2.4 Create init command stub (`internal/cli/init.go`)

### Files to Create

| File | Description |
|------|-------------|
| `cmd/trabuco/main.go` | Entry point, calls root command |
| `internal/cli/root.go` | Root command with description |
| `internal/cli/version.go` | `trabuco version` command |
| `internal/cli/init.go` | `trabuco init` command (stub) |

### Command Structure

```
trabuco                    # Shows help
trabuco init               # Interactive project generation
trabuco version            # Shows version info
trabuco help               # Shows help (built-in)
```

### Acceptance Criteria

- [ ] `go build ./cmd/trabuco` succeeds
- [ ] `./trabuco --help` shows usage
- [ ] `./trabuco version` shows version
- [ ] `./trabuco init` runs (even if incomplete)

---

## Stage 3: Configuration & Module Definitions

**Goal:** Define the data structures for project configuration and module metadata.

### Tasks

- [ ] 3.1 Define `ProjectConfig` struct (`internal/config/project.go`)
- [ ] 3.2 Define `Module` struct and module registry (`internal/config/modules.go`)
- [ ] 3.3 Implement dependency resolution logic

### Data Structures

**ProjectConfig:**
```go
type ProjectConfig struct {
    ProjectName     string
    GroupID         string
    ArtifactID      string
    JavaVersion     string
    Modules         []string
    Database        string   // postgresql, mysql, none
    Messaging       string   // sqs, none
    IncludeDocker   bool
    IncludeGitHub   bool
}
```

**Module:**
```go
type Module struct {
    Name         string
    Description  string
    Required     bool
    Dependencies []string
}
```

### Module Registry

| Module | Required | Dependencies |
|--------|----------|--------------|
| Model | Yes | None |
| SQLDatastore | No | Model |
| Shared | No | Model, SQLDatastore |
| API | No | Model, SQLDatastore, Shared |

### Files to Create

| File | Description |
|------|-------------|
| `internal/config/project.go` | ProjectConfig struct |
| `internal/config/modules.go` | Module definitions & dependency resolution |

### Acceptance Criteria

- [ ] ProjectConfig holds all necessary fields
- [ ] Module dependencies are correctly defined
- [ ] `ResolveDependencies(selected []string)` returns full list with deps
- [ ] Unit tests pass for dependency resolution

---

## Stage 4: Interactive Prompts

**Goal:** Implement the interactive user experience using survey library.

### Tasks

- [ ] 4.1 Implement prompt functions (`internal/prompts/prompts.go`)
- [ ] 4.2 Integrate prompts into init command
- [ ] 4.3 Add validation for inputs

### Prompts Flow

```
1. Project name        (text input, validate: lowercase, no spaces)
2. Group ID            (text input, default: com.company.project)
3. Module selection    (multi-select with dependency hints)
   - Model (required, always selected)
   - SQLDatastore (requires Model)
   - Shared (requires Model, SQLDatastore)
   - API (requires Model, SQLDatastore, Shared)
4. Java version        (select: 21 recommended, 25 latest LTS)
5. Database            (select: PostgreSQL recommended, MySQL, Other/Generic)
   - Only shown if SQLDatastore module is selected
6. Messaging           (select: AWS SQS, None)
   - Only shown if Shared module is selected
7. Include Docker?     (confirm: yes/no)
8. Include GitHub?     (confirm: yes/no)
```

### Validation Rules

| Field | Rules |
|-------|-------|
| Project name | lowercase, alphanumeric + hyphens, no leading/trailing hyphen |
| Group ID | valid Java package (e.g., `com.company.project`) |
| Modules | At least Model must be selected |

### Files to Create

| File | Description |
|------|-------------|
| `internal/prompts/prompts.go` | All interactive prompts |

### Acceptance Criteria

- [ ] All prompts display correctly
- [ ] Validation prevents invalid inputs
- [ ] Module selection auto-includes dependencies
- [ ] `RunPrompts()` returns populated `ProjectConfig`

---

## Stage 5: Template System

**Goal:** Set up the embedded template system using go:embed.

### Tasks

- [ ] 5.1 Create template loading utilities (`internal/templates/templates.go`)
- [ ] 5.2 Define custom template functions (pascalCase, camelCase, hasModule, etc.)
- [ ] 5.3 Set up go:embed for all template directories

### Template Functions

| Function | Example Input | Output |
|----------|---------------|--------|
| `pascalCase` | `my-project` | `MyProject` |
| `camelCase` | `my-project` | `myProject` |
| `packagePath` | `com.company.project` | `com/company/project` |
| `hasModule` | `"Data"` | `true/false` |
| `ifDB` | `"postgresql"` | `true/false` |

### Files to Create

| File | Description |
|------|-------------|
| `internal/templates/templates.go` | Template loading, functions, embed directives |

### Acceptance Criteria

- [ ] Templates load from embedded filesystem
- [ ] Custom functions work in templates
- [ ] Template execution returns rendered content

---

## Stage 6: POM Templates

**Goal:** Create all Maven POM templates based on Wamefy patterns.

### Tasks

- [ ] 6.1 Create parent POM template with dependency management
- [ ] 6.2 Create Model module POM template
- [ ] 6.3 Create SQLDatastore module POM template (database-conditional)
- [ ] 6.4 Create Shared module POM template (with Resilience4j)
- [ ] 6.5 Create API module POM template (web + validation only)

### Reference Files from Wamefy

| Template | Reference |
|----------|-----------|
| Parent POM | `/Users/arianlc/IdeaProjects/Wamefy/pom.xml` |
| Model POM | `/Users/arianlc/IdeaProjects/Wamefy/Model/pom.xml` |
| SQLDatastore POM | `/Users/arianlc/IdeaProjects/Wamefy/Data/pom.xml` |
| Shared POM | `/Users/arianlc/IdeaProjects/Wamefy/Shared/pom.xml` |
| API POM | `/Users/arianlc/IdeaProjects/Wamefy/API/pom.xml` |

### Template Variables

```
{{.ProjectName}}       - my-platform
{{.ProjectNamePascal}} - MyPlatform
{{.GroupID}}           - com.company.project
{{.ArtifactID}}        - my-platform
{{.JavaVersion}}       - 21
{{.Modules}}           - []string{"Model", "SQLDatastore", "Shared", "API"}
{{.Database}}          - postgresql | mysql | generic
{{.Messaging}}         - sqs | none
```

### Database-Conditional Dependencies (SQLDatastore POM)

```xml
{{- if eq .Database "postgresql"}}
<dependency>
    <groupId>org.postgresql</groupId>
    <artifactId>postgresql</artifactId>
    <version>42.7.8</version>
</dependency>
<dependency>
    <groupId>org.flywaydb</groupId>
    <artifactId>flyway-database-postgresql</artifactId>
    <version>${flyway.version}</version>
</dependency>
{{- else if eq .Database "mysql"}}
<dependency>
    <groupId>com.mysql</groupId>
    <artifactId>mysql-connector-j</artifactId>
    <version>9.1.0</version>
</dependency>
<dependency>
    <groupId>org.flywaydb</groupId>
    <artifactId>flyway-mysql</artifactId>
    <version>${flyway.version}</version>
</dependency>
{{- end}}
```

### Files to Create

| File | Description |
|------|-------------|
| `templates/pom/parent.xml.tmpl` | Parent POM with dependency management |
| `templates/pom/model.xml.tmpl` | Model module POM (Immutables, validation-api) |
| `templates/pom/sqldatastore.xml.tmpl` | SQLDatastore POM (DB-conditional) |
| `templates/pom/shared.xml.tmpl` | Shared module POM (Resilience4j) |
| `templates/pom/api.xml.tmpl` | API module POM (web + validation only) |

### Acceptance Criteria

- [ ] Each POM renders with correct group/artifact IDs
- [ ] SQLDatastore POM includes correct driver based on Database choice
- [ ] SQLDatastore POM includes correct Flyway module based on Database choice
- [ ] Shared POM includes Resilience4j dependencies
- [ ] API POM includes ONLY web and validation starters (no DB deps)
- [ ] Parent POM only includes selected modules
- [ ] API POM includes Spring Boot plugin

---

## Stage 7: Java Source Templates

**Goal:** Create Java source file templates for each module, including Placeholder CRUD example.

### Tasks

- [ ] 7.1 Create Model module Java templates (with Placeholder entity/DTOs)
- [ ] 7.2 Create SQLDatastore module Java templates (with PlaceholderRepository + tests)
- [ ] 7.3 Create Shared module Java templates (with PlaceholderService + tests)
- [ ] 7.4 Create API module Java templates (with PlaceholderController)

### Model Module Templates

| File | Purpose |
|------|---------|
| `ImmutableStyle.java.tmpl` | Immutables annotation config |
| `entities/Placeholder.java.tmpl` | Example entity with @Id, @Table |
| `dto/PlaceholderRequest.java.tmpl` | Example request DTO |
| `dto/PlaceholderResponse.java.tmpl` | Example response DTO |

**Generated Package Structure:**
```
Model/src/main/java/{{package}}/model/
├── ImmutableStyle.java
├── entities/
│   └── Placeholder.java
├── dto/
│   ├── PlaceholderRequest.java
│   └── PlaceholderResponse.java
├── enums/                    # Empty, for user's enums
├── exception/                # Empty, for user's exceptions
├── util/                     # Empty, for user's utilities
├── events/                   # Empty, for domain events
└── validation/               # Empty, for custom validators
```

### SQLDatastore Module Templates

| File | Purpose |
|------|---------|
| `config/DatabaseConfig.java.tmpl` | HikariCP DataSource configuration |
| `repository/PlaceholderRepository.java.tmpl` | Spring Data JDBC repository |
| `V1__baseline.sql.tmpl` | Initial Flyway migration with placeholders table |

**Test Templates:**
| File | Purpose |
|------|---------|
| `repository/PlaceholderRepositoryTest.java.tmpl` | 4 integration tests for CRUD |

**Generated Package Structure:**
```
SQLDatastore/
├── src/main/java/{{package}}/sqldatastore/
│   ├── config/
│   │   └── DatabaseConfig.java
│   └── repository/
│       └── PlaceholderRepository.java
├── src/main/resources/
│   └── db/migration/
│       └── V1__baseline.sql
└── src/test/java/{{package}}/sqldatastore/
    └── repository/
        └── PlaceholderRepositoryTest.java   # 4 integration tests
```

### Shared Module Templates

| File | Purpose |
|------|---------|
| `config/SharedConfig.java.tmpl` | Shared module Spring config |
| `config/CircuitBreakerConfig.java.tmpl` | Resilience4j configuration |
| `service/PlaceholderService.java.tmpl` | CRUD service with circuit breaker |

**Test Templates:**
| File | Purpose |
|------|---------|
| `service/PlaceholderServiceTest.java.tmpl` | 4 unit tests (mocked repository) |

**Generated Package Structure:**
```
Shared/
├── src/main/java/{{package}}/shared/
│   ├── config/
│   │   ├── SharedConfig.java
│   │   └── CircuitBreakerConfig.java
│   └── service/
│       └── PlaceholderService.java
└── src/test/java/{{package}}/shared/
    └── service/
        └── PlaceholderServiceTest.java      # 4 unit tests
```

### API Module Templates

| File | Purpose |
|------|---------|
| `Application.java.tmpl` | Spring Boot main class |
| `controller/PlaceholderController.java.tmpl` | CRUD endpoints |
| `controller/HealthController.java.tmpl` | Health check endpoint |
| `config/WebConfig.java.tmpl` | Web/CORS configuration |
| `application.yml.tmpl` | Application properties |

**Generated Package Structure:**
```
API/
├── src/main/java/{{package}}/api/
│   ├── {{ProjectNamePascal}}ApiApplication.java
│   ├── controller/
│   │   ├── PlaceholderController.java
│   │   └── HealthController.java
│   └── config/
│       └── WebConfig.java
└── src/main/resources/
    └── application.yml
```

### Reference Files from Wamefy

| Template | Reference |
|----------|-----------|
| ImmutableStyle | `/Users/arianlc/IdeaProjects/Wamefy/Model/src/main/java/com/wamefy/model/ImmutableStyle.java` |
| DatabaseConfig | `/Users/arianlc/IdeaProjects/Wamefy/Data/src/main/java/com/wamefy/data/config/` |
| CircuitBreaker | `/Users/arianlc/IdeaProjects/Wamefy/Data/src/main/java/com/wamefy/data/circuitbreaker/` |
| Application | `/Users/arianlc/IdeaProjects/Wamefy/API/src/main/java/com/wamefy/api/WamefyApiApplication.java` |
| application.yml | `/Users/arianlc/IdeaProjects/Wamefy/API/src/main/resources/application.yml` |

### Files to Create

```
templates/java/
├── model/
│   ├── ImmutableStyle.java.tmpl
│   ├── entities/Placeholder.java.tmpl
│   ├── dto/PlaceholderRequest.java.tmpl
│   └── dto/PlaceholderResponse.java.tmpl
├── sqldatastore/
│   ├── config/DatabaseConfig.java.tmpl
│   ├── repository/PlaceholderRepository.java.tmpl
│   ├── migration/V1__baseline.sql.tmpl
│   └── test/PlaceholderRepositoryTest.java.tmpl
├── shared/
│   ├── config/SharedConfig.java.tmpl
│   ├── config/CircuitBreakerConfig.java.tmpl
│   ├── service/PlaceholderService.java.tmpl
│   └── test/PlaceholderServiceTest.java.tmpl
└── api/
    ├── Application.java.tmpl
    ├── controller/PlaceholderController.java.tmpl
    ├── controller/HealthController.java.tmpl
    ├── config/WebConfig.java.tmpl
    └── resources/application.yml.tmpl
```

### application.yml Template (Database-Conditional)

```yaml
spring:
  application:
    name: {{.ProjectName}}-api

{{- if eq .Database "postgresql"}}
  datasource:
    url: jdbc:postgresql://${DB_HOST:localhost}:${DB_PORT:5432}/${DB_NAME:{{.ProjectName}}}
    username: ${DB_USER:postgres}
    password: ${DB_PASSWORD:postgres}
    driver-class-name: org.postgresql.Driver
{{- else if eq .Database "mysql"}}
  datasource:
    url: jdbc:mysql://${DB_HOST:localhost}:${DB_PORT:3306}/${DB_NAME:{{.ProjectName}}}
    username: ${DB_USER:root}
    password: ${DB_PASSWORD:root}
    driver-class-name: com.mysql.cj.jdbc.Driver
{{- else}}
  datasource:
    url: ${DB_URL:jdbc:h2:mem:testdb}
    username: ${DB_USER:sa}
    password: ${DB_PASSWORD:}
{{- end}}

  flyway:
    enabled: true
    locations: classpath:db/migration
    baseline-on-migrate: true

resilience4j:
  circuitbreaker:
    instances:
      default:
        registerHealthIndicator: true
        slidingWindowSize: 10
        minimumNumberOfCalls: 5
        failureRateThreshold: 50
        waitDurationInOpenState: 30s

management:
  endpoints:
    web:
      exposure:
        include: health,info,metrics
  endpoint:
    health:
      show-details: always
```

### Test Summary

| Module | Test Type | Count | What's Tested |
|--------|-----------|-------|---------------|
| SQLDatastore | Integration | 4 | Repository CRUD with real database |
| Shared | Unit | 4 | Service methods with mocked repository |
| API | — | 0 | No tests (as specified) |
| **Total** | — | **8** | — |

### Acceptance Criteria

- [ ] All Java files compile after generation
- [ ] Package names use correct GroupID
- [ ] Immutables generates ImmutablePlaceholder, ImmutablePlaceholderRequest, etc.
- [ ] Enums folder is empty (no Immutables applied)
- [ ] SQLDatastore tests pass with H2 or Testcontainers
- [ ] Shared tests pass with Mockito
- [ ] PlaceholderController exposes 4 CRUD endpoints
- [ ] Validation annotations work (@NotBlank, @Size)

---

## Stage 8: Infrastructure Templates

**Goal:** Create Docker, GitHub Actions, IntelliJ run configurations, and project documentation templates.

### Tasks

- [ ] 8.1 Create API Dockerfile template (multi-stage)
- [ ] 8.2 Create docker-compose.yml template
- [ ] 8.3 Create GitHub Actions workflow template
- [ ] 8.4 Create IntelliJ run configuration templates
- [ ] 8.5 Create project documentation templates

### Docker Templates

| File | Purpose |
|------|---------|
| `api/Dockerfile.tmpl` | Multi-stage build for API |
| `docker-compose.yml.tmpl` | Local dev (database + LocalStack if SQS) |

**Database-Conditional docker-compose.yml:**

```yaml
services:
{{- if eq .Database "postgresql"}}
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: {{.ProjectName}}
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
{{- else if eq .Database "mysql"}}
  mysql:
    image: mysql:8.0
    environment:
      MYSQL_DATABASE: {{.ProjectName}}
      MYSQL_ROOT_PASSWORD: root
    ports:
      - "3306:3306"
{{- end}}

{{- if eq .Messaging "sqs"}}
  localstack:
    image: localstack/localstack:latest
    environment:
      SERVICES: sqs
    ports:
      - "4566:4566"
{{- end}}
```

### GitHub Actions Template

| File | Purpose |
|------|---------|
| `build-and-push.yml.tmpl` | CI/CD pipeline |

### IntelliJ Run Configurations

Run configurations are stored in `.run/` directory (shareable via Git).

| File | Type | Purpose |
|------|------|---------|
| `API.run.xml` | Spring Boot | Run the API application |
| `Maven Install.run.xml` | Maven | Run `mvn clean install` on parent |
| `Maven Compile.run.xml` | Maven | Run `mvn compile` on parent |
| `Docker Compose Up.run.xml` | Shell Script | Start local dev services |
| `Docker Compose Down.run.xml` | Shell Script | Stop local dev services |

**Spring Boot Run Configuration (API):**
```xml
<component name="ProjectRunConfigurationManager">
  <configuration name="API" type="SpringBootApplicationConfigurationType">
    <module name="API" />
    <option name="SPRING_BOOT_MAIN_CLASS" value="{{.GroupID}}.api.{{.ProjectNamePascal}}ApiApplication" />
    <option name="ACTIVE_PROFILES" value="local" />
    <option name="WORKING_DIRECTORY" value="$PROJECT_DIR$/API" />
  </configuration>
</component>
```

**Maven Run Configuration:**
```xml
<component name="ProjectRunConfigurationManager">
  <configuration name="Maven Install" type="MavenRunConfiguration">
    <MavenSettings>
      <option name="myRunnerParameters">
        <MavenRunnerParameters>
          <option name="goals">
            <list>
              <option value="clean" />
              <option value="install" />
            </list>
          </option>
          <option name="workingDirPath" value="$PROJECT_DIR$" />
        </MavenRunnerParameters>
      </option>
    </MavenSettings>
  </configuration>
</component>
```

### Documentation Templates

| File | Purpose |
|------|---------|
| `.gitignore.tmpl` | Git ignore patterns |
| `.env.example.tmpl` | Environment variables template |
| `README.md.tmpl` | Project README |
| `CLAUDE.md.tmpl` | AI assistant context |

### Reference Files from Wamefy

| Template | Reference |
|----------|-----------|
| Dockerfile | `/Users/arianlc/IdeaProjects/Wamefy/.infrastructure/docker/api/Dockerfile` |
| docker-compose | `/Users/arianlc/IdeaProjects/Wamefy/.infrastructure/docker/docker-compose.yml` |
| Run Configs | `/Users/arianlc/IdeaProjects/Wamefy/.idea/runConfigurations/*.xml` |

### Files to Create

```
templates/
├── docker/
│   ├── Dockerfile.tmpl
│   └── docker-compose.yml.tmpl
├── github/
│   └── build-and-push.yml.tmpl
├── intellij/
│   ├── API.run.xml.tmpl
│   ├── Maven_Install.run.xml.tmpl
│   ├── Maven_Compile.run.xml.tmpl
│   ├── Docker_Compose_Up.run.xml.tmpl
│   └── Docker_Compose_Down.run.xml.tmpl
└── docs/
    ├── gitignore.tmpl
    ├── env.example.tmpl
    ├── README.md.tmpl
    └── CLAUDE.md.tmpl
```

### Generated Project Structure (Run Configs)

```
my-platform/
├── .run/
│   ├── API.run.xml                    # Spring Boot run config
│   ├── Maven Install.run.xml          # mvn clean install
│   ├── Maven Compile.run.xml          # mvn compile
│   ├── Docker Compose Up.run.xml      # Start local services
│   └── Docker Compose Down.run.xml    # Stop local services
├── ...
```

### Acceptance Criteria

- [ ] Dockerfile builds successfully
- [ ] docker-compose starts all services
- [ ] GitHub Actions workflow is valid YAML
- [ ] README includes correct project info
- [ ] IntelliJ recognizes run configurations on project import
- [ ] API run config starts the Spring Boot application
- [ ] Maven run configs execute correctly

---

## Stage 9: Generator Engine

**Goal:** Implement the main generation logic that orchestrates file creation.

### Tasks

- [ ] 9.1 Implement generator main logic (`internal/generator/generator.go`)
- [ ] 9.2 Implement POM generation (`internal/generator/pom.go`)
- [ ] 9.3 Implement Java source generation (`internal/generator/java.go`)
- [ ] 9.4 Implement infrastructure generation (`internal/generator/infra.go`)
- [ ] 9.5 Add progress output and error handling

### Generator Flow

```
1. Create project root directory
2. Generate parent pom.xml
3. For each selected module:
   a. Create module directory structure
   b. Generate module pom.xml
   c. Generate Java source files
   d. Generate resources (if any)
4. If Docker enabled:
   a. Create .infrastructure/docker/
   b. Generate Dockerfile and docker-compose.yml
5. If GitHub enabled:
   a. Create .github/workflows/
   b. Generate build-and-push.yml
6. Generate IntelliJ run configurations:
   a. Create .run/ directory
   b. Generate API run config (if API module selected)
   c. Generate Maven run configs
   d. Generate Docker Compose run configs (if Docker enabled)
7. Generate documentation files
8. Print success message and next steps
```

### Directory Creation Order

```go
[]string{
    // Root
    "{{.ProjectName}}",

    // Model module
    "{{.ProjectName}}/Model/src/main/java/{{.PackagePath}}/model",
    "{{.ProjectName}}/Model/src/main/java/{{.PackagePath}}/model/entities",
    "{{.ProjectName}}/Model/src/main/java/{{.PackagePath}}/model/dto",
    "{{.ProjectName}}/Model/src/main/java/{{.PackagePath}}/model/enums",
    "{{.ProjectName}}/Model/src/main/java/{{.PackagePath}}/model/exception",
    "{{.ProjectName}}/Model/src/main/java/{{.PackagePath}}/model/util",
    "{{.ProjectName}}/Model/src/main/java/{{.PackagePath}}/model/events",
    "{{.ProjectName}}/Model/src/main/java/{{.PackagePath}}/model/validation",

    // SQLDatastore module (if selected)
    "{{.ProjectName}}/SQLDatastore/src/main/java/{{.PackagePath}}/sqldatastore/config",
    "{{.ProjectName}}/SQLDatastore/src/main/java/{{.PackagePath}}/sqldatastore/repository",
    "{{.ProjectName}}/SQLDatastore/src/main/resources/db/migration",
    "{{.ProjectName}}/SQLDatastore/src/test/java/{{.PackagePath}}/sqldatastore/repository",

    // Shared module (if selected)
    "{{.ProjectName}}/Shared/src/main/java/{{.PackagePath}}/shared/config",
    "{{.ProjectName}}/Shared/src/main/java/{{.PackagePath}}/shared/service",
    "{{.ProjectName}}/Shared/src/test/java/{{.PackagePath}}/shared/service",

    // API module (if selected)
    "{{.ProjectName}}/API/src/main/java/{{.PackagePath}}/api",
    "{{.ProjectName}}/API/src/main/java/{{.PackagePath}}/api/controller",
    "{{.ProjectName}}/API/src/main/java/{{.PackagePath}}/api/config",
    "{{.ProjectName}}/API/src/main/resources",

    // Infrastructure (if enabled)
    "{{.ProjectName}}/.infrastructure/docker/api",
    "{{.ProjectName}}/.github/workflows",
    "{{.ProjectName}}/.run",
}
```

### Files to Create

| File | Description |
|------|-------------|
| `internal/generator/generator.go` | Main Generate() function |
| `internal/generator/pom.go` | POM file generation |
| `internal/generator/java.go` | Java source generation |
| `internal/generator/infra.go` | Docker/GitHub/docs generation |

### Acceptance Criteria

- [ ] Full project generates in target directory
- [ ] All files have correct content
- [ ] Progress output shows each step
- [ ] Errors are handled gracefully
- [ ] No partial generation on failure (cleanup)

---

## Stage 10: Testing & Validation

**Goal:** Ensure the generated projects are valid and functional.

### Tasks

- [ ] 10.1 Write unit tests for config/modules
- [ ] 10.2 Write unit tests for template rendering
- [ ] 10.3 Write integration test for full generation
- [ ] 10.4 Write E2E test: generate + maven compile
- [ ] 10.5 Test all module combinations

### Test Matrix

| Modules | Database | Messaging | Docker | GitHub | IntelliJ |
|---------|----------|-----------|--------|--------|----------|
| Model only | N/A | None | No | No | Yes |
| Model + SQLDatastore | PostgreSQL | N/A | No | No | Yes |
| Model + SQLDatastore | MySQL | N/A | No | No | Yes |
| Model + SQLDatastore + Shared | PostgreSQL | SQS | No | No | Yes |
| All modules | PostgreSQL | SQS | Yes | Yes | Yes |
| All modules | MySQL | None | Yes | Yes | Yes |
| All modules | Generic | None | Yes | Yes | Yes |

### Test Commands

```bash
# Unit tests (Go CLI)
go test ./...

# Build CLI
go build -o trabuco ./cmd/trabuco

# Generate test project (non-interactive)
./trabuco init --non-interactive \
  --name test-project \
  --group-id com.test \
  --modules Model,SQLDatastore,Shared,API \
  --database postgresql \
  --messaging sqs

# Maven compile test
cd test-project && mvn clean compile

# Run generated tests (SQLDatastore + Shared = 8 tests)
mvn test

# API startup test
cd API && mvn spring-boot:run &
sleep 10
curl http://localhost:8080/actuator/health

# Test Placeholder CRUD endpoints
curl -X POST http://localhost:8080/api/placeholders \
  -H "Content-Type: application/json" \
  -d '{"name": "test"}'

curl http://localhost:8080/api/placeholders/1
```

### Files to Create

```
internal/
├── config/
│   └── modules_test.go
├── templates/
│   └── templates_test.go
└── generator/
    └── generator_test.go
```

### Acceptance Criteria

- [ ] All unit tests pass (Go CLI)
- [ ] Generated project compiles with Maven
- [ ] Generated project tests pass (8 tests: 4 integration + 4 unit)
- [ ] API starts successfully
- [ ] Health endpoint responds 200
- [ ] Placeholder CRUD endpoints work (POST, GET, PUT, DELETE)
- [ ] `.run/` directory contains valid XML run configurations
- [ ] IntelliJ imports project and shows run configurations

---

## Stage 11: Distribution

**Goal:** Set up release automation and installation methods.

### Tasks

- [ ] 11.1 Create Makefile with build targets
- [ ] 11.2 Create GitHub Actions release workflow
- [ ] 11.3 Create install.sh script
- [ ] 11.4 Document installation methods
- [ ] 11.5 Create Homebrew formula (optional)

### Build Targets

```makefile
PLATFORMS := darwin-amd64 darwin-arm64 linux-amd64 linux-arm64 windows-amd64

build:
	go build -o trabuco ./cmd/trabuco

build-all:
	$(foreach platform,$(PLATFORMS),\
		GOOS=$(word 1,$(subst -, ,$(platform))) \
		GOARCH=$(word 2,$(subst -, ,$(platform))) \
		go build -o dist/trabuco-$(platform) ./cmd/trabuco;)
```

### Install Script (`install.sh`)

```bash
#!/bin/bash
set -e

REPO="your-org/trabuco"
VERSION="latest"
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture
case $ARCH in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac

# Download and install
curl -sSL "https://github.com/$REPO/releases/$VERSION/download/trabuco-$OS-$ARCH" \
  -o /usr/local/bin/trabuco
chmod +x /usr/local/bin/trabuco

echo "Trabuco installed successfully!"
trabuco version
```

### GitHub Actions Release Workflow

```yaml
name: Release
on:
  push:
    tags: ['v*']
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: Build all platforms
        run: make build-all
      - name: Create Release
        uses: softprops/action-gh-release@v1
        with:
          files: dist/*
```

### Files to Create

| File | Description |
|------|-------------|
| `Makefile` | Build automation |
| `scripts/install.sh` | Curl install script |
| `.github/workflows/release.yml` | Release automation |
| `docs/INSTALLATION.md` | Installation guide |

### Installation Methods

1. **curl (Primary)**
   ```bash
   curl -sSL https://github.com/your-org/trabuco/releases/latest/download/install.sh | bash
   ```

2. **Homebrew**
   ```bash
   brew tap your-org/trabuco && brew install trabuco
   ```

3. **Go install**
   ```bash
   go install github.com/your-org/trabuco/cmd/trabuco@latest
   ```

4. **Manual download**
   - Download from GitHub releases
   - Move to PATH

### Acceptance Criteria

- [ ] `make build-all` produces binaries for all platforms
- [ ] GitHub release creates automatically on tag push
- [ ] Install script works on Mac and Linux
- [ ] `trabuco version` shows correct version after install

---

## Summary Checklist

### Stage Completion

| Stage | Description | Status |
|-------|-------------|--------|
| 0 | Research & Technical Decisions | [x] |
| 1 | Project Initialization | [x] |
| 2 | CLI Framework Setup | [x] |
| 3 | Configuration & Module Definitions | [ ] |
| 4 | Interactive Prompts | [ ] |
| 5 | Template System | [ ] |
| 6 | POM Templates | [ ] |
| 7 | Java Source Templates | [ ] |
| 8 | Infrastructure Templates | [ ] |
| 9 | Generator Engine | [ ] |
| 10 | Testing & Validation | [ ] |
| 11 | Distribution | [ ] |

### Key Milestones

| Milestone | Criteria | Status |
|-----------|----------|--------|
| CLI runs | `trabuco version` works | [x] |
| Prompts work | Interactive flow completes | [ ] |
| Basic generation | Model module generates | [ ] |
| Full generation | All 4 modules generate | [ ] |
| Valid project | `mvn clean compile` succeeds | [ ] |
| API runs | Health endpoint responds | [ ] |
| IntelliJ ready | Run configs appear in IDE | [ ] |
| Release ready | Install script works | [ ] |

---

## Notes

### Implementation Guidelines
- Reference Wamefy project extensively during template creation
- Test each module independently before combining
- Keep templates simple; avoid over-engineering
- Version templates with semantic versioning
- Consider backward compatibility for future module additions

### Key Technical Decisions (Verified)

| Decision | Choice | Reason |
|----------|--------|--------|
| Java Version | 21 (LTS) | Stable, widely adopted, long support |
| ORM | Spring Data JDBC | 4x faster, simpler than JPA |
| Connection Pool | HikariCP 5.1.0 | Zero-overhead, Spring Boot default |
| Migrations | Flyway 10.22.0 | SQL-first, simple, proven |
| Immutables | 2.10.1 | Thread-safe value objects |
| Circuit Breaker | Resilience4j 2.2.0 | Lightweight, Spring Boot 3 native |
| Validation | Jakarta Validation | Standard, well-documented |

### Sources Used for Research
- [Oracle Java SE Support Roadmap](https://www.oracle.com/java/technologies/java-se-support-roadmap.html)
- [Spring Boot Documentation](https://docs.spring.io/spring-boot/reference/)
- [Baeldung Tutorials](https://www.baeldung.com/)
- [HikariCP GitHub](https://github.com/brettwooldridge/HikariCP)
- [Flyway Documentation](https://flywaydb.org/documentation/)
- [Resilience4j Documentation](https://resilience4j.readme.io/)
- [Immutables Documentation](https://immutables.github.io/)
