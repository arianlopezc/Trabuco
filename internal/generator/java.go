package generator

import (
	"fmt"
	"path/filepath"
)

// generateModelModule generates all Model module files
func (g *Generator) generateModelModule() error {
	// Generate module POM
	if err := g.generateModulePOM("Model"); err != nil {
		return err
	}

	// ImmutableStyle.java
	if err := g.writeTemplate(
		"java/model/ImmutableStyle.java.tmpl",
		g.javaPath("Model", "ImmutableStyle.java"),
	); err != nil {
		return fmt.Errorf("failed to generate ImmutableStyle.java: %w", err)
	}

	// Placeholder.java (entity interface)
	if err := g.writeTemplate(
		"java/model/entities/Placeholder.java.tmpl",
		g.javaPath("Model", filepath.Join("entities", "Placeholder.java")),
	); err != nil {
		return fmt.Errorf("failed to generate Placeholder.java: %w", err)
	}

	// PlaceholderRecord.java (SQL database record) - only if SQLDatastore selected
	if g.config.HasModule("SQLDatastore") {
		if err := g.writeTemplate(
			"java/model/entities/PlaceholderRecord.java.tmpl",
			g.javaPath("Model", filepath.Join("entities", "PlaceholderRecord.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PlaceholderRecord.java: %w", err)
		}
	}

	// PlaceholderDocument.java (NoSQL document) - only if NoSQLDatastore selected
	if g.config.HasModule("NoSQLDatastore") {
		if err := g.writeTemplate(
			"java/model/entities/PlaceholderDocument.java.tmpl",
			g.javaPath("Model", filepath.Join("entities", "PlaceholderDocument.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PlaceholderDocument.java: %w", err)
		}
	}

	// PlaceholderRequest.java (DTO)
	if err := g.writeTemplate(
		"java/model/dto/PlaceholderRequest.java.tmpl",
		g.javaPath("Model", filepath.Join("dto", "PlaceholderRequest.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderRequest.java: %w", err)
	}

	// PlaceholderResponse.java (DTO)
	if err := g.writeTemplate(
		"java/model/dto/PlaceholderResponse.java.tmpl",
		g.javaPath("Model", filepath.Join("dto", "PlaceholderResponse.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderResponse.java: %w", err)
	}

	// Event classes (only if EventConsumer is selected)
	if g.config.HasModule("EventConsumer") {
		// PlaceholderEvent.java (sealed interface)
		if err := g.writeTemplate(
			"java/model/events/PlaceholderEvent.java.tmpl",
			g.javaPath("Model", filepath.Join("events", "PlaceholderEvent.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PlaceholderEvent.java: %w", err)
		}

		// PlaceholderCreatedEvent.java (concrete event)
		if err := g.writeTemplate(
			"java/model/events/PlaceholderCreatedEvent.java.tmpl",
			g.javaPath("Model", filepath.Join("events", "PlaceholderCreatedEvent.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PlaceholderCreatedEvent.java: %w", err)
		}
	}

	// Job request classes (only if Worker is selected)
	if g.config.HasModule("Worker") {
		// PlaceholderJobRequest.java (sealed interface for placeholder domain)
		if err := g.writeTemplate(
			"java/model/jobs/PlaceholderJobRequest.java.tmpl",
			g.javaPath("Model", filepath.Join("jobs", "PlaceholderJobRequest.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PlaceholderJobRequest.java: %w", err)
		}

		// ProcessPlaceholderJobRequest.java (concrete job request)
		if err := g.writeTemplate(
			"java/model/jobs/ProcessPlaceholderJobRequest.java.tmpl",
			g.javaPath("Model", filepath.Join("jobs", "ProcessPlaceholderJobRequest.java")),
		); err != nil {
			return fmt.Errorf("failed to generate ProcessPlaceholderJobRequest.java: %w", err)
		}

		// ProcessPlaceholderJobRequestHandler.java (base handler class)
		if err := g.writeTemplate(
			"java/model/jobs/ProcessPlaceholderJobRequestHandler.java.tmpl",
			g.javaPath("Model", filepath.Join("jobs", "ProcessPlaceholderJobRequestHandler.java")),
		); err != nil {
			return fmt.Errorf("failed to generate ProcessPlaceholderJobRequestHandler.java: %w", err)
		}
	}

	return nil
}

// generateSQLDatastoreModule generates all SQLDatastore module files
func (g *Generator) generateSQLDatastoreModule() error {
	// Generate module POM
	if err := g.generateModulePOM("SQLDatastore"); err != nil {
		return err
	}

	// DatabaseConfig.java
	if err := g.writeTemplate(
		"java/sqldatastore/config/DatabaseConfig.java.tmpl",
		g.javaPath("SQLDatastore", filepath.Join("config", "DatabaseConfig.java")),
	); err != nil {
		return fmt.Errorf("failed to generate DatabaseConfig.java: %w", err)
	}

	// PlaceholderRepository.java
	if err := g.writeTemplate(
		"java/sqldatastore/repository/PlaceholderRepository.java.tmpl",
		g.javaPath("SQLDatastore", filepath.Join("repository", "PlaceholderRepository.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderRepository.java: %w", err)
	}

	// V1__baseline.sql (Flyway migration)
	if err := g.writeTemplate(
		"java/sqldatastore/migration/V1__baseline.sql.tmpl",
		g.resourcePath("SQLDatastore", filepath.Join("db", "migration", "V1__baseline.sql")),
	); err != nil {
		return fmt.Errorf("failed to generate V1__baseline.sql: %w", err)
	}

	// application.yml (database configuration)
	if err := g.writeTemplate(
		"java/sqldatastore/resources/application.yml.tmpl",
		g.resourcePath("SQLDatastore", "application.yml"),
	); err != nil {
		return fmt.Errorf("failed to generate SQLDatastore application.yml: %w", err)
	}

	// TestConfig.java (test configuration for library module)
	if err := g.writeTemplate(
		"java/sqldatastore/test/TestConfig.java.tmpl",
		g.testJavaPath("SQLDatastore", "TestConfig.java"),
	); err != nil {
		return fmt.Errorf("failed to generate TestConfig.java: %w", err)
	}

	// PlaceholderRepositoryTest.java
	if err := g.writeTemplate(
		"java/sqldatastore/test/PlaceholderRepositoryTest.java.tmpl",
		g.testJavaPath("SQLDatastore", filepath.Join("repository", "PlaceholderRepositoryTest.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderRepositoryTest.java: %w", err)
	}

	return nil
}

// generateNoSQLDatastoreModule generates all NoSQLDatastore module files
func (g *Generator) generateNoSQLDatastoreModule() error {
	// Generate module POM
	if err := g.generateModulePOM("NoSQLDatastore"); err != nil {
		return err
	}

	// NoSQLConfig.java
	if err := g.writeTemplate(
		"java/nosqldatastore/config/NoSQLConfig.java.tmpl",
		g.javaPath("NoSQLDatastore", filepath.Join("config", "NoSQLConfig.java")),
	); err != nil {
		return fmt.Errorf("failed to generate NoSQLConfig.java: %w", err)
	}

	// PlaceholderDocumentRepository.java
	if err := g.writeTemplate(
		"java/nosqldatastore/repository/PlaceholderDocumentRepository.java.tmpl",
		g.javaPath("NoSQLDatastore", filepath.Join("repository", "PlaceholderDocumentRepository.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderDocumentRepository.java: %w", err)
	}

	// application.yml (NoSQL configuration)
	if err := g.writeTemplate(
		"java/nosqldatastore/resources/application.yml.tmpl",
		g.resourcePath("NoSQLDatastore", "application.yml"),
	); err != nil {
		return fmt.Errorf("failed to generate NoSQLDatastore application.yml: %w", err)
	}

	// TestConfig.java (test configuration for library module)
	if err := g.writeTemplate(
		"java/nosqldatastore/test/TestConfig.java.tmpl",
		g.testJavaPath("NoSQLDatastore", "TestConfig.java"),
	); err != nil {
		return fmt.Errorf("failed to generate NoSQLDatastore TestConfig.java: %w", err)
	}

	// PlaceholderDocumentRepositoryTest.java
	if err := g.writeTemplate(
		"java/nosqldatastore/test/PlaceholderDocumentRepositoryTest.java.tmpl",
		g.testJavaPath("NoSQLDatastore", filepath.Join("repository", "PlaceholderDocumentRepositoryTest.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderDocumentRepositoryTest.java: %w", err)
	}

	return nil
}

// generateSharedModule generates all Shared module files
func (g *Generator) generateSharedModule() error {
	// Generate module POM
	if err := g.generateModulePOM("Shared"); err != nil {
		return err
	}

	// SharedConfig.java
	if err := g.writeTemplate(
		"java/shared/config/SharedConfig.java.tmpl",
		g.javaPath("Shared", filepath.Join("config", "SharedConfig.java")),
	); err != nil {
		return fmt.Errorf("failed to generate SharedConfig.java: %w", err)
	}

	// CircuitBreakerConfiguration.java
	if err := g.writeTemplate(
		"java/shared/config/CircuitBreakerConfiguration.java.tmpl",
		g.javaPath("Shared", filepath.Join("config", "CircuitBreakerConfiguration.java")),
	); err != nil {
		return fmt.Errorf("failed to generate CircuitBreakerConfiguration.java: %w", err)
	}

	// application.yml (circuit breaker configuration)
	if err := g.writeTemplate(
		"java/shared/resources/application.yml.tmpl",
		g.resourcePath("Shared", "application.yml"),
	); err != nil {
		return fmt.Errorf("failed to generate Shared application.yml: %w", err)
	}

	// PlaceholderService.java
	if err := g.writeTemplate(
		"java/shared/service/PlaceholderService.java.tmpl",
		g.javaPath("Shared", filepath.Join("service", "PlaceholderService.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderService.java: %w", err)
	}

	// PlaceholderServiceTest.java
	if err := g.writeTemplate(
		"java/shared/test/PlaceholderServiceTest.java.tmpl",
		g.testJavaPath("Shared", filepath.Join("service", "PlaceholderServiceTest.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderServiceTest.java: %w", err)
	}

	return nil
}

// generateAPIModule generates all API module files
func (g *Generator) generateAPIModule() error {
	// Generate module POM
	if err := g.generateModulePOM("API"); err != nil {
		return err
	}

	// Application.java (main class)
	applicationFile := fmt.Sprintf("%sApiApplication.java", g.config.ProjectNamePascal())
	if err := g.writeTemplate(
		"java/api/Application.java.tmpl",
		g.javaPath("API", applicationFile),
	); err != nil {
		return fmt.Errorf("failed to generate Application.java: %w", err)
	}

	// HealthController.java
	if err := g.writeTemplate(
		"java/api/controller/HealthController.java.tmpl",
		g.javaPath("API", filepath.Join("controller", "HealthController.java")),
	); err != nil {
		return fmt.Errorf("failed to generate HealthController.java: %w", err)
	}

	// PlaceholderController.java
	if err := g.writeTemplate(
		"java/api/controller/PlaceholderController.java.tmpl",
		g.javaPath("API", filepath.Join("controller", "PlaceholderController.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderController.java: %w", err)
	}

	// PlaceholderJobController.java (only when Worker module is selected)
	if g.config.HasModule("Worker") {
		if err := g.writeTemplate(
			"java/api/controller/PlaceholderJobController.java.tmpl",
			g.javaPath("API", filepath.Join("controller", "PlaceholderJobController.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PlaceholderJobController.java: %w", err)
		}
	}

	// EventController.java (only when EventConsumer module is selected)
	if g.config.HasModule("EventConsumer") {
		if err := g.writeTemplate(
			"java/api/controller/EventController.java.tmpl",
			g.javaPath("API", filepath.Join("controller", "EventController.java")),
		); err != nil {
			return fmt.Errorf("failed to generate EventController.java: %w", err)
		}
	}

	// WebConfig.java
	if err := g.writeTemplate(
		"java/api/config/WebConfig.java.tmpl",
		g.javaPath("API", filepath.Join("config", "WebConfig.java")),
	); err != nil {
		return fmt.Errorf("failed to generate WebConfig.java: %w", err)
	}

	// GlobalExceptionHandler.java
	if err := g.writeTemplate(
		"java/api/config/GlobalExceptionHandler.java.tmpl",
		g.javaPath("API", filepath.Join("config", "GlobalExceptionHandler.java")),
	); err != nil {
		return fmt.Errorf("failed to generate GlobalExceptionHandler.java: %w", err)
	}

	// SecurityHeadersFilter.java
	if err := g.writeTemplate(
		"java/api/config/SecurityHeadersFilter.java.tmpl",
		g.javaPath("API", filepath.Join("config", "SecurityHeadersFilter.java")),
	); err != nil {
		return fmt.Errorf("failed to generate SecurityHeadersFilter.java: %w", err)
	}

	// application.yml
	if err := g.writeTemplate(
		"java/api/resources/application.yml.tmpl",
		g.resourcePath("API", "application.yml"),
	); err != nil {
		return fmt.Errorf("failed to generate application.yml: %w", err)
	}

	// logback-spring.xml (structured logging)
	if err := g.writeTemplate(
		"java/api/resources/logback-spring.xml.tmpl",
		g.resourcePath("API", "logback-spring.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate API logback-spring.xml: %w", err)
	}

	// Dockerfile
	if err := g.writeTemplate(
		"docker/api.Dockerfile.tmpl",
		filepath.Join("API", "Dockerfile"),
	); err != nil {
		return fmt.Errorf("failed to generate API Dockerfile: %w", err)
	}

	// IntelliJ IDEA Run Configuration (Maven)
	if err := g.writeTemplate(
		"idea/run/API__Maven_.run.xml.tmpl",
		filepath.Join(".run", "API.run.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate API run configuration: %w", err)
	}

	return nil
}

// generateJobsModule generates all Jobs module files
// NOTE: Job request schemas are generated in the Model module.
// The Jobs module contains job service classes for enqueueing jobs.
func (g *Generator) generateJobsModule() error {
	// Generate module POM
	if err := g.generateModulePOM("Jobs"); err != nil {
		return err
	}

	// PlaceholderJobService.java (service for enqueueing jobs)
	if err := g.writeTemplate(
		"java/jobs/PlaceholderJobService.java.tmpl",
		g.javaPath("Jobs", "PlaceholderJobService.java"),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderJobService.java: %w", err)
	}

	return nil
}

// generateWorkerModule generates all Worker module files
func (g *Generator) generateWorkerModule() error {
	// Generate module POM
	if err := g.generateModulePOM("Worker"); err != nil {
		return err
	}

	// WorkerApplication.java (main class)
	applicationFile := fmt.Sprintf("%sWorkerApplication.java", g.config.ProjectNamePascal())
	if err := g.writeTemplate(
		"java/worker/WorkerApplication.java.tmpl",
		g.javaPath("Worker", applicationFile),
	); err != nil {
		return fmt.Errorf("failed to generate WorkerApplication.java: %w", err)
	}

	// JobRunrConfig.java
	if err := g.writeTemplate(
		"java/worker/config/JobRunrConfig.java.tmpl",
		g.javaPath("Worker", filepath.Join("config", "JobRunrConfig.java")),
	); err != nil {
		return fmt.Errorf("failed to generate JobRunrConfig.java: %w", err)
	}

	// RecurringJobsConfig.java
	if err := g.writeTemplate(
		"java/worker/config/RecurringJobsConfig.java.tmpl",
		g.javaPath("Worker", filepath.Join("config", "RecurringJobsConfig.java")),
	); err != nil {
		return fmt.Errorf("failed to generate RecurringJobsConfig.java: %w", err)
	}

	// ProcessPlaceholderJobRequestHandler.java
	if err := g.writeTemplate(
		"java/worker/handler/ProcessPlaceholderJobRequestHandler.java.tmpl",
		g.javaPath("Worker", filepath.Join("handler", "ProcessPlaceholderJobRequestHandler.java")),
	); err != nil {
		return fmt.Errorf("failed to generate ProcessPlaceholderJobRequestHandler.java: %w", err)
	}

	// application.yml
	if err := g.writeTemplate(
		"java/worker/resources/application.yml.tmpl",
		g.resourcePath("Worker", "application.yml"),
	); err != nil {
		return fmt.Errorf("failed to generate Worker application.yml: %w", err)
	}

	// ProcessPlaceholderJobRequestHandlerTest.java
	if err := g.writeTemplate(
		"java/worker/test/ProcessPlaceholderJobRequestHandlerTest.java.tmpl",
		g.testJavaPath("Worker", filepath.Join("handler", "ProcessPlaceholderJobRequestHandlerTest.java")),
	); err != nil {
		return fmt.Errorf("failed to generate ProcessPlaceholderJobRequestHandlerTest.java: %w", err)
	}

	// logback-spring.xml (structured logging)
	if err := g.writeTemplate(
		"java/worker/resources/logback-spring.xml.tmpl",
		g.resourcePath("Worker", "logback-spring.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate Worker logback-spring.xml: %w", err)
	}

	// Dockerfile
	if err := g.writeTemplate(
		"docker/worker.Dockerfile.tmpl",
		filepath.Join("Worker", "Dockerfile"),
	); err != nil {
		return fmt.Errorf("failed to generate Worker Dockerfile: %w", err)
	}

	// IntelliJ IDEA Run Configuration (Maven)
	if err := g.writeTemplate(
		"idea/run/Worker__Maven_.run.xml.tmpl",
		filepath.Join(".run", "Worker.run.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate Worker run configuration: %w", err)
	}

	return nil
}

// generateEventsModule generates the Events module (EventPublisher service)
func (g *Generator) generateEventsModule() error {
	// pom.xml
	if err := g.writeTemplate(
		"pom/events.xml.tmpl",
		filepath.Join("Events", "pom.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate Events pom.xml: %w", err)
	}

	// EventPublisher.java (service for publishing events)
	if err := g.writeTemplate(
		"java/events/EventPublisher.java.tmpl",
		g.javaPath("Events", "EventPublisher.java"),
	); err != nil {
		return fmt.Errorf("failed to generate EventPublisher.java: %w", err)
	}

	// RabbitConfig.java (RabbitMQ JSON configuration) - only for RabbitMQ
	if g.config.UsesRabbitMQ() {
		if err := g.writeTemplate(
			"java/events/config/RabbitConfig.java.tmpl",
			g.javaPath("Events", filepath.Join("config", "RabbitConfig.java")),
		); err != nil {
			return fmt.Errorf("failed to generate Events RabbitConfig.java: %w", err)
		}
	}

	return nil
}

// generateEventConsumerModule generates the EventConsumer module
func (g *Generator) generateEventConsumerModule() error {
	// pom.xml
	if err := g.writeTemplate(
		"pom/eventconsumer.xml.tmpl",
		filepath.Join("EventConsumer", "pom.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate EventConsumer pom.xml: %w", err)
	}

	// Application.java
	if err := g.writeTemplate(
		"java/eventconsumer/EventConsumerApplication.java.tmpl",
		g.javaPath("EventConsumer", g.config.ProjectNamePascal()+"EventConsumerApplication.java"),
	); err != nil {
		return fmt.Errorf("failed to generate EventConsumerApplication.java: %w", err)
	}

	// Config (Kafka or RabbitMQ)
	if g.config.UsesKafka() {
		if err := g.writeTemplate(
			"java/eventconsumer/config/KafkaConfig.java.tmpl",
			g.javaPath("EventConsumer", filepath.Join("config", "KafkaConfig.java")),
		); err != nil {
			return fmt.Errorf("failed to generate KafkaConfig.java: %w", err)
		}
	} else if g.config.UsesRabbitMQ() {
		if err := g.writeTemplate(
			"java/eventconsumer/config/RabbitConfig.java.tmpl",
			g.javaPath("EventConsumer", filepath.Join("config", "RabbitConfig.java")),
		); err != nil {
			return fmt.Errorf("failed to generate RabbitConfig.java: %w", err)
		}
	}

	// PlaceholderEventListener.java
	if err := g.writeTemplate(
		"java/eventconsumer/listener/PlaceholderEventListener.java.tmpl",
		g.javaPath("EventConsumer", filepath.Join("listener", "PlaceholderEventListener.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderEventListener.java: %w", err)
	}

	// application.yml
	if err := g.writeTemplate(
		"java/eventconsumer/resources/application.yml.tmpl",
		g.resourcePath("EventConsumer", "application.yml"),
	); err != nil {
		return fmt.Errorf("failed to generate EventConsumer application.yml: %w", err)
	}

	// logback-spring.xml
	if err := g.writeTemplate(
		"java/eventconsumer/resources/logback-spring.xml.tmpl",
		g.resourcePath("EventConsumer", "logback-spring.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate EventConsumer logback-spring.xml: %w", err)
	}

	// Dockerfile
	if err := g.writeTemplate(
		"docker/eventconsumer.Dockerfile.tmpl",
		filepath.Join("EventConsumer", "Dockerfile"),
	); err != nil {
		return fmt.Errorf("failed to generate EventConsumer Dockerfile: %w", err)
	}

	// Test
	if err := g.writeTemplate(
		"java/eventconsumer/test/PlaceholderEventListenerTest.java.tmpl",
		g.testJavaPath("EventConsumer", filepath.Join("listener", "PlaceholderEventListenerTest.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderEventListenerTest.java: %w", err)
	}

	// IntelliJ run configuration
	if err := g.writeTemplate(
		"idea/run/EventConsumer__Maven_.run.xml.tmpl",
		filepath.Join(".run", "EventConsumer.run.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate EventConsumer run configuration: %w", err)
	}

	return nil
}
