package migrate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/templates"
)

// generateDockerCompose creates the docker-compose.yml file
func (m *Migrator) generateDockerCompose() error {
	services := []string{}
	volumes := []string{}

	// Database service
	if m.projectInfo.UsesNoSQL {
		services = append(services, `  mongodb:
    image: mongo:7
    container_name: {{name}}-mongodb
    ports:
      - "27017:27017"
    environment:
      MONGO_INITDB_ROOT_USERNAME: admin
      MONGO_INITDB_ROOT_PASSWORD: changeme
    volumes:
      - mongodb_data:/data/db
    healthcheck:
      test: ["CMD", "mongosh", "--eval", "db.adminCommand('ping')"]
      interval: 10s
      timeout: 5s
      retries: 5`)
		volumes = append(volumes, "  mongodb_data:")
	} else {
		services = append(services, `  postgres:
    image: postgres:16-alpine
    container_name: {{name}}-postgres
    ports:
      - "5432:5432"
    environment:
      POSTGRES_USER: {{name}}
      POSTGRES_PASSWORD: changeme
      POSTGRES_DB: {{name}}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U {{name}}"]
      interval: 10s
      timeout: 5s
      retries: 5`)
		volumes = append(volumes, "  postgres_data:")
	}

	// Message broker service
	if m.projectInfo.HasEventListeners {
		if m.projectInfo.UsesRabbitMQ {
			services = append(services, `  rabbitmq:
    image: rabbitmq:3-management-alpine
    container_name: {{name}}-rabbitmq
    ports:
      - "5672:5672"
      - "15672:15672"
    environment:
      RABBITMQ_DEFAULT_USER: guest
      RABBITMQ_DEFAULT_PASS: guest
    volumes:
      - rabbitmq_data:/var/lib/rabbitmq
    healthcheck:
      test: ["CMD", "rabbitmq-diagnostics", "check_running"]
      interval: 10s
      timeout: 5s
      retries: 5`)
			volumes = append(volumes, "  rabbitmq_data:")
		} else {
			services = append(services, `  kafka:
    image: confluentinc/cp-kafka:7.5.0
    container_name: {{name}}-kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
      KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,CONTROLLER://localhost:9093
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      CLUSTER_ID: MkU3OEVBNTcwNTJENDM2Qk
    volumes:
      - kafka_data:/var/lib/kafka/data
    healthcheck:
      test: ["CMD", "kafka-broker-api-versions", "--bootstrap-server", "localhost:9092"]
      interval: 10s
      timeout: 10s
      retries: 5`)
			volumes = append(volumes, "  kafka_data:")
		}
	}

	// Redis for caching (optional, include if detected)
	if m.projectInfo.UsesRedis {
		services = append(services, `  redis:
    image: redis:7-alpine
    container_name: {{name}}-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 5s
      retries: 5`)
		volumes = append(volumes, "  redis_data:")
	}

	// Build the complete docker-compose.yml
	content := fmt.Sprintf(`version: '3.8'

services:
%s

volumes:
%s
`, strings.Join(services, "\n\n"), strings.Join(volumes, "\n"))

	// Replace placeholders
	name := strings.ToLower(m.projectInfo.Name)
	name = strings.ReplaceAll(name, " ", "-")
	content = strings.ReplaceAll(content, "{{name}}", name)

	return os.WriteFile(filepath.Join(m.config.OutputPath, "docker-compose.yml"), []byte(content), 0644)
}

// generateEnvExample creates the .env.example file
func (m *Migrator) generateEnvExample() error {
	lines := []string{
		"# Application Configuration",
		"SERVER_PORT=8080",
		"",
	}

	// Database configuration
	if m.projectInfo.UsesNoSQL {
		lines = append(lines,
			"# MongoDB Configuration",
			"SPRING_DATA_MONGODB_URI=mongodb://admin:changeme@localhost:27017/"+strings.ToLower(m.projectInfo.Name)+"?authSource=admin",
			"",
		)
	} else {
		lines = append(lines,
			"# PostgreSQL Configuration",
			"SPRING_DATASOURCE_URL=jdbc:postgresql://localhost:5432/"+strings.ToLower(m.projectInfo.Name),
			"SPRING_DATASOURCE_USERNAME="+strings.ToLower(m.projectInfo.Name),
			"SPRING_DATASOURCE_PASSWORD=changeme",
			"",
		)
	}

	// Message broker configuration
	if m.projectInfo.HasEventListeners {
		if m.projectInfo.UsesRabbitMQ {
			lines = append(lines,
				"# RabbitMQ Configuration",
				"SPRING_RABBITMQ_HOST=localhost",
				"SPRING_RABBITMQ_PORT=5672",
				"SPRING_RABBITMQ_USERNAME=guest",
				"SPRING_RABBITMQ_PASSWORD=guest",
				"",
			)
		} else {
			lines = append(lines,
				"# Kafka Configuration",
				"SPRING_KAFKA_BOOTSTRAP_SERVERS=localhost:9092",
				"",
			)
		}
	}

	// Redis configuration
	if m.projectInfo.UsesRedis {
		lines = append(lines,
			"# Redis Configuration",
			"SPRING_DATA_REDIS_HOST=localhost",
			"SPRING_DATA_REDIS_PORT=6379",
			"",
		)
	}

	// JobRunr configuration for Worker module
	if m.projectInfo.HasScheduledJobs {
		lines = append(lines,
			"# JobRunr Configuration",
			"ORG_JOBRUNR_BACKGROUND_JOB_SERVER_ENABLED=true",
			"ORG_JOBRUNR_DASHBOARD_ENABLED=true",
			"ORG_JOBRUNR_DASHBOARD_PORT=8000",
			"",
		)
	}

	// Logging configuration
	lines = append(lines,
		"# Logging Configuration",
		"LOGGING_LEVEL_ROOT=INFO",
		"LOGGING_LEVEL_"+strings.ToUpper(strings.ReplaceAll(m.projectInfo.GroupID, ".", "_"))+"=DEBUG",
		"",
	)

	content := strings.Join(lines, "\n")
	return os.WriteFile(filepath.Join(m.config.OutputPath, ".env.example"), []byte(content), 0644)
}

// generateAIAgentFiles creates AI agent context files
func (m *Migrator) generateAIAgentFiles() error {
	// Generate CLAUDE.md
	claudeContent := fmt.Sprintf(`# %s

This is a multi-module Spring Boot project generated by Trabuco Migrate.

## Project Structure

- **Model/** - Domain entities, DTOs, and value objects
- **%s/** - Data access layer with repositories
- **Shared/** - Business logic and shared services
- **API/** - REST controllers and HTTP endpoints
`, m.projectInfo.Name, m.getDatastoreModule())

	if m.projectInfo.HasScheduledJobs {
		claudeContent += "- **Worker/** - Background job processing with JobRunr\n"
	}
	if m.projectInfo.HasEventListeners {
		if m.projectInfo.UsesRabbitMQ {
			claudeContent += "- **EventConsumer/** - RabbitMQ message consumers\n"
		} else {
			claudeContent += "- **EventConsumer/** - Kafka message consumers\n"
		}
	}

	claudeContent += fmt.Sprintf(`
## Key Patterns

- **Spring Data JDBC** for database access (not JPA)
- **Constructor injection** for all dependencies
- **Immutables-style DTOs** for data transfer
- **Resilience4j** for circuit breakers

## Base Package

%s

## Build

` + "```bash" + `
mvn clean install
` + "```" + `

## Run

` + "```bash" + `
docker-compose up -d
cd API && mvn spring-boot:run
` + "```" + `
`, m.projectInfo.GroupID)

	if err := os.WriteFile(filepath.Join(m.config.OutputPath, "CLAUDE.md"), []byte(claudeContent), 0644); err != nil {
		return err
	}

	// Generate .cursorrules
	cursorContent := fmt.Sprintf(`# Cursor Rules for %s

This is a Trabuco-generated multi-module Spring Boot project.

## Architecture
- Multi-module Maven project
- Spring Boot 3.x with Spring Data JDBC
- No JPA/Hibernate - use Spring Data JDBC patterns

## Coding Standards
- Constructor injection only (no @Autowired on fields)
- Entities use @Table annotation (not @Entity)
- Foreign keys as explicit fields (Long userId, not User user)
- DTOs use builder pattern

## Module Responsibilities
- Model: Entities, DTOs, value objects only
- %s: Repositories only
- Shared: Services and business logic
- API: REST controllers only

## Testing
- Use @SpringBootTest for integration tests
- Use Mockito for unit tests
- Each service should have corresponding test class
`, m.projectInfo.Name, m.getDatastoreModule())

	return os.WriteFile(filepath.Join(m.config.OutputPath, ".cursorrules"), []byte(cursorContent), 0644)
}

// generateParentPOM creates the parent pom.xml file using the template
func (m *Migrator) generateParentPOM() error {
	cfg := m.buildProjectConfig()
	engine := templates.NewEngine()

	content, err := engine.Execute("pom/parent.xml.tmpl", cfg)
	if err != nil {
		return fmt.Errorf("failed to render parent POM template: %w", err)
	}

	return os.WriteFile(filepath.Join(m.config.OutputPath, "pom.xml"), []byte(content), 0644)
}

// generateREADME creates the README.md file
func (m *Migrator) generateREADME() error {
	content := fmt.Sprintf(`# %s

This project was migrated from a legacy Spring Boot application to Trabuco's multi-module architecture.

## Prerequisites

- Java %s+
- Maven 3.9+
- Docker & Docker Compose

## Quick Start

1. Start infrastructure:
`+"```bash"+`
docker-compose up -d
`+"```"+`

2. Build the project:
`+"```bash"+`
mvn clean install
`+"```"+`

3. Run the API:
`+"```bash"+`
cd API && mvn spring-boot:run
`+"```"+`

## Modules

| Module | Description |
|--------|-------------|
| Model | Domain entities, DTOs, and value objects |
| %s | Data access layer with repositories |
| Shared | Business logic and shared services |
| API | REST controllers and HTTP endpoints |
`, m.projectInfo.Name, m.projectInfo.JavaVersion, m.getDatastoreModule())

	if m.projectInfo.HasScheduledJobs {
		content += "| Worker | Background job processing with JobRunr |\n"
	}
	if m.projectInfo.HasEventListeners {
		broker := "Kafka"
		if m.projectInfo.UsesRabbitMQ {
			broker = "RabbitMQ"
		}
		content += fmt.Sprintf("| EventConsumer | %s message consumers |\n", broker)
	}

	content += "\n## Configuration\n\n"
	content += "Copy `.env.example` to `.env` and update the values:\n\n"
	content += "```bash\n"
	content += "cp .env.example .env\n"
	content += "```\n\n"
	content += "## API Documentation\n\n"
	content += "Once running, access Swagger UI at: http://localhost:8080/swagger-ui.html\n\n"
	content += "## Migration Notes\n\n"
	content += fmt.Sprintf("This project was migrated on %s.\n\n", time.Now().Format("2006-01-02"))
	content += "### Key Changes from Original\n\n"
	content += "- **JPA → Spring Data JDBC**: Entities no longer use JPA annotations\n"
	content += "- **Lombok → Explicit code**: Getters/setters are now explicit\n"
	content += "- **Single module → Multi-module**: Code is organized by architectural layer\n\n"
	content += "For more details, see the CLAUDE.md file.\n"

	return os.WriteFile(filepath.Join(m.config.OutputPath, "README.md"), []byte(content), 0644)
}

// generateGitignore creates the .gitignore file
func (m *Migrator) generateGitignore() error {
	content := `# Maven
target/
pom.xml.tag
pom.xml.releaseBackup
pom.xml.versionsBackup
pom.xml.next
release.properties

# IDE
.idea/
*.iml
*.ipr
*.iws
.project
.classpath
.settings/
.vscode/
*.swp
*.swo
*~

# Build
build/
out/
bin/

# Logs
logs/
*.log

# Environment
.env
.env.local
.env.*.local

# OS
.DS_Store
Thumbs.db

# Application
*.jar
*.war
*.ear

# Test coverage
jacoco.exec
*.exec

# Trabuco
.trabuco-migrate/
`

	return os.WriteFile(filepath.Join(m.config.OutputPath, ".gitignore"), []byte(content), 0644)
}

// generateMetadata creates the .trabuco.json metadata file
func (m *Migrator) generateMetadata() error {
	modules := []string{config.ModuleModel, m.getDatastoreModule(), config.ModuleShared, config.ModuleAPI}

	if m.projectInfo.HasScheduledJobs {
		modules = append(modules, config.ModuleWorker)
	}
	if m.projectInfo.HasEventListeners {
		modules = append(modules, config.ModuleEventConsumer)
	}

	database := "postgresql"
	if m.projectInfo.UsesNoSQL {
		database = "mongodb"
	}

	messageBroker := ""
	if m.projectInfo.HasEventListeners {
		messageBroker = "kafka"
		if m.projectInfo.UsesRabbitMQ {
			messageBroker = "rabbitmq"
		}
	}

	metadata := map[string]interface{}{
		"version":       "1.0.0",
		"trabuco":       "1.4",
		"migrated_from": m.config.SourcePath,
		"migrated_at":   time.Now().Format(time.RFC3339),
		"project": map[string]interface{}{
			"name":     m.projectInfo.Name,
			"group_id": m.projectInfo.GroupID,
			"modules":  modules,
		},
		"infrastructure": map[string]interface{}{
			"database":       database,
			"message_broker": messageBroker,
			"uses_redis":     m.projectInfo.UsesRedis,
		},
		"statistics": map[string]interface{}{
			"entities":        len(m.projectInfo.Entities),
			"repositories":    len(m.projectInfo.Repositories),
			"services":        len(m.projectInfo.Services),
			"controllers":     len(m.projectInfo.Controllers),
			"jobs":            len(m.projectInfo.ScheduledJobs),
			"event_listeners": len(m.projectInfo.EventListeners),
		},
	}

	// Add AI cost if tracked
	if m.checkpoint.Current() != nil {
		metadata["ai_cost"] = m.checkpoint.Current().EstimatedCost
	}

	content, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(m.config.OutputPath, ".trabuco.json"), []byte(content), 0644)
}

// getDatastoreModule returns the appropriate datastore module name
func (m *Migrator) getDatastoreModule() string {
	if m.projectInfo.UsesNoSQL {
		return config.ModuleNoSQLDatastore
	}
	return config.ModuleSQLDatastore
}

// buildProjectConfig creates a ProjectConfig from the migrator's ProjectInfo
func (m *Migrator) buildProjectConfig() *config.ProjectConfig {
	modules := []string{config.ModuleModel, m.getDatastoreModule(), config.ModuleShared, config.ModuleAPI}

	if m.projectInfo.HasScheduledJobs {
		modules = append(modules, config.ModuleWorker)
	}
	if m.projectInfo.HasEventListeners {
		modules = append(modules, config.ModuleEventConsumer)
	}

	database := config.DatabasePostgreSQL
	noSQLDatabase := ""
	messageBroker := ""

	if m.projectInfo.UsesNoSQL {
		noSQLDatabase = config.DatabaseMongoDB
	}

	if m.projectInfo.HasEventListeners {
		if m.projectInfo.UsesRabbitMQ {
			messageBroker = config.BrokerRabbitMQ
		} else {
			messageBroker = config.BrokerKafka
		}
	}

	javaVersion := m.projectInfo.JavaVersion
	if javaVersion == "" {
		javaVersion = "21"
	}

	// Strip "-parent" suffix if present since templates add it
	projectName := strings.ToLower(m.projectInfo.Name)
	artifactID := projectName
	if strings.HasSuffix(artifactID, "-parent") {
		artifactID = strings.TrimSuffix(artifactID, "-parent")
	}

	return &config.ProjectConfig{
		ProjectName:   artifactID,
		GroupID:       m.projectInfo.GroupID,
		ArtifactID:    artifactID,
		JavaVersion:   javaVersion,
		Modules:       modules,
		Database:      database,
		NoSQLDatabase: noSQLDatabase,
		MessageBroker: messageBroker,
		AIAgents:      []string{"claude", "cursor"},
	}
}

// generateModulePOMs creates pom.xml files for each module
func (m *Migrator) generateModulePOMs() error {
	cfg := m.buildProjectConfig()
	engine := templates.NewEngine()

	for _, module := range cfg.Modules {
		var templateName string
		switch module {
		case config.ModuleModel:
			templateName = "pom/model.xml.tmpl"
		case config.ModuleSQLDatastore:
			templateName = "pom/sqldatastore.xml.tmpl"
		case config.ModuleNoSQLDatastore:
			templateName = "pom/nosqldatastore.xml.tmpl"
		case config.ModuleShared:
			templateName = "pom/shared.xml.tmpl"
		case config.ModuleAPI:
			templateName = "pom/api.xml.tmpl"
		case config.ModuleWorker:
			templateName = "pom/worker.xml.tmpl"
		case config.ModuleEventConsumer:
			templateName = "pom/eventconsumer.xml.tmpl"
		default:
			continue
		}

		content, err := engine.Execute(templateName, cfg)
		if err != nil {
			return fmt.Errorf("failed to render template for %s: %w", module, err)
		}

		outputPath := filepath.Join(m.config.OutputPath, module, "pom.xml")
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", module, err)
		}

		if err := os.WriteFile(outputPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write pom.xml for %s: %w", module, err)
		}
	}

	return nil
}
