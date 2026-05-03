package generator

import (
	"fmt"
	"path/filepath"

	"github.com/arianlopezc/Trabuco/internal/config"
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
	if g.config.HasModule(config.ModuleSQLDatastore) {
		if err := g.writeTemplate(
			"java/model/entities/PlaceholderRecord.java.tmpl",
			g.javaPath("Model", filepath.Join("entities", "PlaceholderRecord.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PlaceholderRecord.java: %w", err)
		}
	}

	// PlaceholderDocument.java (NoSQL document) - only if NoSQLDatastore selected
	if g.config.HasModule(config.ModuleNoSQLDatastore) {
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
	if g.config.HasModule(config.ModuleEventConsumer) {
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

	// Auth model files (emitted whenever a consuming module — API or
	// AIAgent — is selected; the runtime gate is the
	// {@code trabuco.auth.enabled} property in the generated app's yaml,
	// not a generator-time switch). IdentityClaims and AuthorityScope are
	// universal data — read by API filters, Worker handlers, EventConsumer
	// listeners, and AIAgent tools — so they live in Model alongside the
	// rest of the project's schemas. AuthenticatedRequest is a generic
	// wrapper used at async boundaries (job parameters, broker bodies) when
	// callers want to carry identity inline rather than via headers.
	if g.config.AuthEnabled() {
		modelAuthFiles := []struct {
			tmpl string
			out  string
		}{
			{"java/model/auth/IdentityClaims.java.tmpl", "IdentityClaims.java"},
			{"java/model/auth/AuthorityScope.java.tmpl", "AuthorityScope.java"},
			{"java/model/auth/AuthenticatedRequest.java.tmpl", "AuthenticatedRequest.java"},
		}
		for _, f := range modelAuthFiles {
			if err := g.writeTemplate(f.tmpl, g.javaPath("Model", filepath.Join("auth", f.out))); err != nil {
				return fmt.Errorf("failed to generate %s: %w", f.out, err)
			}
		}
		// Jackson round-trip test for AuthenticatedRequest — verifies it
		// (de)serializes cleanly across module / process boundaries.
		if err := g.writeTemplate(
			"java/model/test/auth/AuthenticatedRequestTest.java.tmpl",
			g.testJavaPath("Model", filepath.Join("auth", "AuthenticatedRequestTest.java")),
		); err != nil {
			return fmt.Errorf("failed to generate AuthenticatedRequestTest.java: %w", err)
		}
	}

	// Job request classes (only if Worker is selected)
	if g.config.HasModule(config.ModuleWorker) {
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

	// ArchitectureTest.java (ArchUnit rules)
	if err := g.writeTemplate(
		"java/shared/test/ArchitectureTest.java.tmpl",
		g.testJavaPath("Shared", "ArchitectureTest.java"),
	); err != nil {
		return fmt.Errorf("failed to generate ArchitectureTest.java: %w", err)
	}

	// Auth utilities (emitted whenever a consuming module — API or
	// AIAgent — is selected; the runtime gate is
	// {@code trabuco.auth.enabled} in the generated app's yaml).
	// Logic — RequestContextHolder, JwtClaimsExtractor (interface +
	// default impl), AuthContextPropagator (interface + default impl),
	// AuthScope (try-with-resources helper) — lives in Shared. The
	// data types these reference (IdentityClaims, AuthenticatedRequest)
	// are in Model. MockJwtFactory is a test utility for minting fake
	// JWTs/Authentications/Decoders — also in Shared so all modules
	// that depend on Shared (test scope) can use it.
	if g.config.AuthEnabled() {
		authFiles := []struct {
			tmpl string
			out  string
		}{
			{"java/shared/auth/RequestContextHolder.java.tmpl", "RequestContextHolder.java"},
			{"java/shared/auth/JwtClaimsExtractor.java.tmpl", "JwtClaimsExtractor.java"},
			{"java/shared/auth/DefaultJwtClaimsExtractor.java.tmpl", "DefaultJwtClaimsExtractor.java"},
			{"java/shared/auth/AuthContextPropagator.java.tmpl", "AuthContextPropagator.java"},
			{"java/shared/auth/DefaultAuthContextPropagator.java.tmpl", "DefaultAuthContextPropagator.java"},
			{"java/shared/auth/AuthScope.java.tmpl", "AuthScope.java"},
		}
		for _, f := range authFiles {
			if err := g.writeTemplate(f.tmpl, g.javaPath("Shared", filepath.Join("auth", f.out))); err != nil {
				return fmt.Errorf("failed to generate %s: %w", f.out, err)
			}
		}
		// Test utility + unit tests for the auth utilities — same
		// package as production auth types so tests can use it
		// without import gymnastics. SignedJwtTestSupport is the
		// e2e helper used by API/AuthEndToEndTest; lives in Shared
		// so AIAgent or any other module can reuse it.
		sharedAuthTestFiles := []struct {
			tmpl string
			out  string
		}{
			{"java/shared/test/auth/MockJwtFactory.java.tmpl", "MockJwtFactory.java"},
			{"java/shared/test/auth/AuthScopeTest.java.tmpl", "AuthScopeTest.java"},
			{"java/shared/test/auth/DefaultAuthContextPropagatorTest.java.tmpl", "DefaultAuthContextPropagatorTest.java"},
			{"java/shared/test/auth/MockJwtFactoryTest.java.tmpl", "MockJwtFactoryTest.java"},
		}
		for _, f := range sharedAuthTestFiles {
			if err := g.writeTemplate(f.tmpl, g.testJavaPath("Shared", filepath.Join("auth", f.out))); err != nil {
				return fmt.Errorf("failed to generate %s: %w", f.out, err)
			}
		}
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
	if g.config.HasModule(config.ModuleWorker) {
		if err := g.writeTemplate(
			"java/api/controller/PlaceholderJobController.java.tmpl",
			g.javaPath("API", filepath.Join("controller", "PlaceholderJobController.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PlaceholderJobController.java: %w", err)
		}
	}

	// EventController.java (only when EventConsumer module is selected)
	if g.config.HasModule(config.ModuleEventConsumer) {
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

	// F-DATA-02 / F-DATA-12: WeakCredentialsWarning emits a loud
	// boot-time error when the datasource password matches a
	// well-known weak default outside dev/test profiles.
	if g.config.HasModule(config.ModuleSQLDatastore) {
		if err := g.writeTemplate(
			"java/api/config/WeakCredentialsWarning.java.tmpl",
			g.javaPath("API", filepath.Join("config", "WeakCredentialsWarning.java")),
		); err != nil {
			return fmt.Errorf("failed to generate WeakCredentialsWarning.java: %w", err)
		}
	}

	// CorrelationIdFilter.java
	if err := g.writeTemplate(
		"java/api/config/CorrelationIdFilter.java.tmpl",
		g.javaPath("API", filepath.Join("config", "CorrelationIdFilter.java")),
	); err != nil {
		return fmt.Errorf("failed to generate CorrelationIdFilter.java: %w", err)
	}

	// OpenAPIConfig.java
	if err := g.writeTemplate(
		"java/api/config/OpenAPIConfig.java.tmpl",
		g.javaPath("API", filepath.Join("config", "OpenAPIConfig.java")),
	); err != nil {
		return fmt.Errorf("failed to generate OpenAPIConfig.java: %w", err)
	}

	// API auth filter chain (emitted whenever API is selected — auth
	// scaffolding ships with the REST tier and stays dormant until
	// {@code trabuco.auth.enabled=true} flips it on). The HTTP-specific
	// concerns — Spring Security dual filter chains, JWT to Authentication
	// conversion, ProblemDetail-formatted 401/403, OpenAPI bearer scheme —
	// live in API. Cross-module identity utilities live in Shared and the
	// underlying data types live in Model.
	if g.config.AuthEnabled() {
		apiAuthFiles := []struct {
			tmpl string
			out  string
		}{
			{"java/api/config/security/SecurityConfig.java.tmpl", "SecurityConfig.java"},
			{"java/api/config/security/MethodSecurityConfig.java.tmpl", "MethodSecurityConfig.java"},
			{"java/api/config/security/JwtAuthenticationConverter.java.tmpl", "JwtAuthenticationConverter.java"},
			{"java/api/config/security/AuthProblemDetailHandler.java.tmpl", "AuthProblemDetailHandler.java"},
			{"java/api/config/security/OpenApiSecurityConfig.java.tmpl", "OpenApiSecurityConfig.java"},
			// F-AUTH-14: filter-level RequestContextHolder.clear() so
			// identity doesn't leak across virtual-thread carrier reuse.
			{"java/api/config/security/RequestContextClearingFilter.java.tmpl", "RequestContextClearingFilter.java"},
		}
		for _, f := range apiAuthFiles {
			out := filepath.Join("config", "security", f.out)
			if err := g.writeTemplate(f.tmpl, g.javaPath("API", out)); err != nil {
				return fmt.Errorf("failed to generate %s: %w", f.out, err)
			}
		}
		// Integration test for the resource-server filter chain
		// (MockMvc + @MockBean JwtDecoder — fast, no real HTTP).
		if err := g.writeTemplate(
			"java/api/test/security/SecurityIntegrationTest.java.tmpl",
			g.testJavaPath("API", filepath.Join("config", "security", "SecurityIntegrationTest.java")),
		); err != nil {
			return fmt.Errorf("failed to generate SecurityIntegrationTest.java: %w", err)
		}
		// End-to-end test (real Tomcat + real signed JWTs + real
		// signature verification via NimbusJwtDecoder) plus its
		// helper. Higher fidelity than SecurityIntegrationTest.
		apiE2EFiles := []struct {
			tmpl string
			out  string
		}{
			{"java/api/test/security/SignedJwtTestSupport.java.tmpl", "SignedJwtTestSupport.java"},
			{"java/api/test/security/AuthEndToEndTest.java.tmpl", "AuthEndToEndTest.java"},
			// Regression backstop for the dormant default — verifies
			// that when trabuco.auth.enabled is unset the permit-all
			// chain is the active SecurityFilterChain (no 401 leakage).
			{"java/api/test/security/AuthDormantTest.java.tmpl", "AuthDormantTest.java"},
		}
		for _, f := range apiE2EFiles {
			if err := g.writeTemplate(f.tmpl, g.testJavaPath("API", filepath.Join("config", "security", f.out))); err != nil {
				return fmt.Errorf("failed to generate %s: %w", f.out, err)
			}
		}
	}

	// F-WEB-01 ArchUnit guard — every controller endpoint must declare
	// an explicit authorization decision. Lives module-local because
	// the test classpath has to see API's own controller classes.
	if err := g.writeTemplate(
		"java/api/test/ApiArchitectureTest.java.tmpl",
		g.testJavaPath("API", "ApiArchitectureTest.java"),
	); err != nil {
		return fmt.Errorf("failed to generate ApiArchitectureTest.java: %w", err)
	}

	// GlobalExceptionHandler integration test — emitted only when both
	// the SQLDatastore module and Postgres database are selected, since
	// the test relies on a Postgres Testcontainer to surface real
	// DataIntegrityViolation behaviour (H2's compatibility mode emits a
	// different SQLState that translates to a different exception class).
	// AuthEnabled is also required because the existing API pom only
	// pulls Testcontainers test deps under the same gate.
	if g.config.HasModule(config.ModuleAPI) && g.config.HasModule(config.ModuleSQLDatastore) && g.config.Database == config.DatabasePostgreSQL && g.config.AuthEnabled() {
		if err := g.writeTemplate(
			"java/api/test/exception/GlobalExceptionHandlerIntegrationTest.java.tmpl",
			g.testJavaPath("API", filepath.Join("config", "GlobalExceptionHandlerIntegrationTest.java")),
		); err != nil {
			return fmt.Errorf("failed to generate GlobalExceptionHandlerIntegrationTest.java: %w", err)
		}
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

	// PubSubPublisherConfig.java (Pub/Sub JSON configuration) - only for Pub/Sub
	if g.config.UsesPubSub() {
		if err := g.writeTemplate(
			"java/events/config/PubSubPublisherConfig.java.tmpl",
			g.javaPath("Events", filepath.Join("config", "PubSubPublisherConfig.java")),
		); err != nil {
			return fmt.Errorf("failed to generate Events PubSubPublisherConfig.java: %w", err)
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

	// Config (Kafka, RabbitMQ, SQS, or Pub/Sub)
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
	} else if g.config.UsesSQS() {
		if err := g.writeTemplate(
			"java/eventconsumer/config/SqsConfig.java.tmpl",
			g.javaPath("EventConsumer", filepath.Join("config", "SqsConfig.java")),
		); err != nil {
			return fmt.Errorf("failed to generate SqsConfig.java: %w", err)
		}
	} else if g.config.UsesPubSub() {
		if err := g.writeTemplate(
			"java/eventconsumer/config/PubSubConfig.java.tmpl",
			g.javaPath("EventConsumer", filepath.Join("config", "PubSubConfig.java")),
		); err != nil {
			return fmt.Errorf("failed to generate PubSubConfig.java: %w", err)
		}
	}

	// PlaceholderEventListener.java
	if err := g.writeTemplate(
		"java/eventconsumer/listener/PlaceholderEventListener.java.tmpl",
		g.javaPath("EventConsumer", filepath.Join("listener", "PlaceholderEventListener.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderEventListener.java: %w", err)
	}

	// F-EVENTS-05: in-memory idempotency tracker — bounded LRU; doc
	// recommends DB/Redis-backed replacement for multi-instance
	// deployments.
	if err := g.writeTemplate(
		"java/eventconsumer/listener/IdempotencyTracker.java.tmpl",
		g.javaPath("EventConsumer", filepath.Join("listener", "IdempotencyTracker.java")),
	); err != nil {
		return fmt.Errorf("failed to generate IdempotencyTracker.java: %w", err)
	}

	// IdempotencyConfig: registers the default in-memory tracker as
	// @ConditionalOnMissingBean so users can override with a DB-backed
	// or Redis-backed implementation. Emits a startup WARN when the
	// default is in use, surfacing the multi-instance limitation.
	if err := g.writeTemplate(
		"java/eventconsumer/config/IdempotencyConfig.java.tmpl",
		g.javaPath("EventConsumer", filepath.Join("config", "IdempotencyConfig.java")),
	); err != nil {
		return fmt.Errorf("failed to generate IdempotencyConfig.java: %w", err)
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

// generateAIAgentModuleAuthFiles emits the OIDC-based security scaffolding
// for AIAgent: AgentSecurityConfig (dual filter chains gated on
// {@code trabuco.auth.enabled}), JwtAuthenticationConverter
// (Jwt → Authentication + RequestContextHolder population), and
// AuthProblemDetailHandler (RFC 7807 401/403 emission). Coexists with the
// legacy ApiKeyAuthFilter, which is governed by its own
// {@code app.aiagent.api-key.enabled} property and stays on by default.
func (g *Generator) generateAIAgentModuleAuthFiles() error {
	files := []struct {
		tmpl string
		out  string
	}{
		{"java/aiagent/config/security/AgentSecurityConfig.java.tmpl", "AgentSecurityConfig.java"},
		{"java/aiagent/config/security/MethodSecurityConfig.java.tmpl", "MethodSecurityConfig.java"},
		{"java/aiagent/config/security/JwtAuthenticationConverter.java.tmpl", "JwtAuthenticationConverter.java"},
		{"java/aiagent/config/security/AuthProblemDetailHandler.java.tmpl", "AuthProblemDetailHandler.java"},
		// RequestContextClearingFilter: mirrors the API module's
		// equivalent. AIAgent's JwtAuthenticationConverter populates
		// RequestContextHolder via convert(); without a finally-clear
		// filter, identity ThreadLocal leaks across virtual-thread
		// carrier reuse.
		{"java/aiagent/config/security/RequestContextClearingFilter.java.tmpl", "RequestContextClearingFilter.java"},
	}
	for _, f := range files {
		out := filepath.Join("config", "security", f.out)
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}
	return nil
}

// generateAIAgentModule generates all files for the AI Agent module.
// This includes: security (auth, scopes, guardrails), tools, agents (primary + specialist),
// brain (scratchpad, reflection), knowledge, protocols (REST, A2A, discovery, streaming, webhooks),
// task manager, webhook manager, config, and resources.
func (g *Generator) generateAIAgentModule() error {
	// Generate module POM
	if err := g.generateModulePOM("AIAgent"); err != nil {
		return err
	}

	// ─── Application Class ──────────────────────────────────────────────
	applicationFile := fmt.Sprintf("%sAIAgentApplication.java", g.config.ProjectNamePascal())
	if err := g.writeTemplate(
		"java/aiagent/AIAgentApplication.java.tmpl",
		g.javaPath("AIAgent", applicationFile),
	); err != nil {
		return fmt.Errorf("failed to generate AIAgentApplication.java: %w", err)
	}

	// ─── Config ─────────────────────────────────────────────────────────
	configFiles := []struct{ tmpl, out string }{
		{"java/aiagent/config/ChatClientConfig.java.tmpl", filepath.Join("config", "ChatClientConfig.java")},
		{"java/aiagent/config/McpServerConfig.java.tmpl", filepath.Join("config", "McpServerConfig.java")},
		{"java/aiagent/config/WebConfig.java.tmpl", filepath.Join("config", "WebConfig.java")},
		{"java/aiagent/config/AgentExceptionHandler.java.tmpl", filepath.Join("config", "AgentExceptionHandler.java")},
		// F-WEB-03: AIAgent module's response security headers — same set
		// the API module's SecurityHeadersFilter applies, plus HSTS.
		{"java/aiagent/config/AgentSecurityHeadersFilter.java.tmpl", filepath.Join("config", "AgentSecurityHeadersFilter.java")},
	}
	if g.config.VectorStoreIsPgVector() {
		// Second Flyway bean for the vector schema. Only emitted when
		// PGVector is selected — Qdrant and Mongo Atlas paths don't
		// share a Postgres datasource, so they don't need this.
		configFiles = append(configFiles,
			struct{ tmpl, out string }{"java/aiagent/config/VectorFlywayConfig.java.tmpl", filepath.Join("config", "VectorFlywayConfig.java")},
		)
	}
	if g.config.HasVectorStore() {
		// Wires VectorKnowledgeRetriever (@Primary) and
		// DocumentIngestionService as plain @Bean methods that take
		// VectorStore as a parameter. Generator-time conditional
		// emission means VectorStore is guaranteed to exist at runtime
		// when this file is generated — no @ConditionalOnBean
		// gymnastics needed (and they wouldn't work reliably anyway,
		// see the Javadoc on the configuration class).
		configFiles = append(configFiles,
			struct{ tmpl, out string }{"java/aiagent/config/KnowledgeBeansConfiguration.java.tmpl", filepath.Join("config", "KnowledgeBeansConfiguration.java")},
		)
	}
	for _, f := range configFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// ─── Vector schema migration (PGVector only) ────────────────────────
	if g.config.VectorStoreIsPgVector() {
		if err := g.writeTemplate(
			"java/aiagent/migration/V1__create_vector_schema.sql.tmpl",
			g.resourcePath("AIAgent", filepath.Join("db", "vector-migration", "V1__create_vector_schema.sql")),
		); err != nil {
			return fmt.Errorf("failed to generate vector-schema migration: %w", err)
		}
	}

	// ─── Security ───────────────────────────────────────────────────────
	securityFiles := []struct{ tmpl, out string }{
		{"java/aiagent/security/CallerIdentity.java.tmpl", filepath.Join("security", "CallerIdentity.java")},
		{"java/aiagent/security/CallerContext.java.tmpl", filepath.Join("security", "CallerContext.java")},
		{"java/aiagent/security/AgentAuthProperties.java.tmpl", filepath.Join("security", "AgentAuthProperties.java")},
		{"java/aiagent/security/ApiKeyAuthFilter.java.tmpl", filepath.Join("security", "ApiKeyAuthFilter.java")},
		{"java/aiagent/security/DemoKeyStartupWarning.java.tmpl", filepath.Join("security", "DemoKeyStartupWarning.java")},
		{"java/aiagent/security/ScopeEnforcer.java.tmpl", filepath.Join("security", "ScopeEnforcer.java")},
		{"java/aiagent/security/RequireScope.java.tmpl", filepath.Join("security", "RequireScope.java")},
		{"java/aiagent/security/ScopeInterceptor.java.tmpl", filepath.Join("security", "ScopeInterceptor.java")},
		{"java/aiagent/security/RateLimiter.java.tmpl", filepath.Join("security", "RateLimiter.java")},
		{"java/aiagent/security/InputGuardrailAdvisor.java.tmpl", filepath.Join("security", "InputGuardrailAdvisor.java")},
		{"java/aiagent/security/OutputGuardrailAdvisor.java.tmpl", filepath.Join("security", "OutputGuardrailAdvisor.java")},
		{"java/aiagent/security/CorrelationIdFilter.java.tmpl", filepath.Join("security", "CorrelationIdFilter.java")},
		// F-AIAGENT-06 / F-INFRA-15: gates /mcp/** behind JWT scope or
		// partner-tier API key. Spring conditional registration ties
		// the filter's existence to spring.ai.mcp.server.enabled=true.
		{"java/aiagent/security/AgentMcpAuthorizationFilter.java.tmpl", filepath.Join("security", "AgentMcpAuthorizationFilter.java")},
	}
	for _, f := range securityFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// OIDC-based security scaffolding (emitted whenever AIAgent is
	// selected; runtime-activated via {@code trabuco.auth.enabled}).
	// AgentSecurityConfig + JwtAuthenticationConverter +
	// AuthProblemDetailHandler — coexists with the legacy
	// ApiKeyAuthFilter above (governed by its own
	// {@code app.aiagent.api-key.enabled} property).
	if g.config.AuthEnabled() {
		if err := g.generateAIAgentModuleAuthFiles(); err != nil {
			return err
		}
	}

	// ─── Tools ──────────────────────────────────────────────────────────
	if err := g.writeTemplate(
		"java/aiagent/tool/PlaceholderTools.java.tmpl",
		g.javaPath("AIAgent", filepath.Join("tool", "PlaceholderTools.java")),
	); err != nil {
		return fmt.Errorf("failed to generate PlaceholderTools.java: %w", err)
	}

	// ─── Agents ─────────────────────────────────────────────────────────
	agentFiles := []struct{ tmpl, out string }{
		{"java/aiagent/agent/PrimaryAgent.java.tmpl", filepath.Join("agent", "PrimaryAgent.java")},
		{"java/aiagent/agent/SpecialistAgent.java.tmpl", filepath.Join("agent", "SpecialistAgent.java")},
		{"java/aiagent/agent/SpecialistAgentTool.java.tmpl", filepath.Join("agent", "SpecialistAgentTool.java")},
	}
	for _, f := range agentFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// ─── Brain ──────────────────────────────────────────────────────────
	brainFiles := []struct{ tmpl, out string }{
		{"java/aiagent/brain/MemoryEntry.java.tmpl", filepath.Join("brain", "MemoryEntry.java")},
		{"java/aiagent/brain/Scratchpad.java.tmpl", filepath.Join("brain", "Scratchpad.java")},
		{"java/aiagent/brain/ReflectionDecision.java.tmpl", filepath.Join("brain", "ReflectionDecision.java")},
		{"java/aiagent/brain/ReflectionService.java.tmpl", filepath.Join("brain", "ReflectionService.java")},
	}
	for _, f := range brainFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// ─── Knowledge ──────────────────────────────────────────────────────
	// KnowledgeRetriever is the strategy interface for fetching relevant
	// documents; KeywordKnowledgeRetriever is the default token-scoring
	// implementation (active when no VectorStore bean is wired). When
	// HasVectorStore is true, VectorKnowledgeRetriever supplants it via
	// @ConditionalOnMissingBean(VectorStore.class), and the Spring AI
	// VectorStore + EmbeddingModel beans come from the conditionally-
	// added starters in aiagent.xml.tmpl.
	knowledgeFiles := []struct{ tmpl, out string }{
		{"java/aiagent/knowledge/KnowledgeBase.java.tmpl", filepath.Join("knowledge", "KnowledgeBase.java")},
		{"java/aiagent/knowledge/KnowledgeRetriever.java.tmpl", filepath.Join("knowledge", "KnowledgeRetriever.java")},
		{"java/aiagent/knowledge/KeywordKnowledgeRetriever.java.tmpl", filepath.Join("knowledge", "KeywordKnowledgeRetriever.java")},
		{"java/aiagent/knowledge/KnowledgeTools.java.tmpl", filepath.Join("knowledge", "KnowledgeTools.java")},
		// F-AIAGENT-04: every retrieved chunk reaches the LLM through
		// FencingDocumentRetriever, which wraps text in
		// <untrusted_document> tags and strips role-token bleed. Used
		// by both the RAG advisor (PrimaryAgent) and the @Tool path
		// (KnowledgeTools), so emit unconditionally whenever AIAgent
		// is selected — regardless of whether a vector store is wired.
		{"java/aiagent/knowledge/FencingDocumentRetriever.java.tmpl", filepath.Join("knowledge", "FencingDocumentRetriever.java")},
		// D1: TenantFilteringDocumentRetriever is referenced from
		// PrimaryAgent.buildRagAdvisor unconditionally. The advisor
		// path itself only fires when an Optional<VectorStore> bean
		// resolves at runtime, but the source must compile in
		// no-vector-store builds too — its only deps are Spring AI core
		// types that ship with every AIAgent variant.
		{"java/aiagent/knowledge/TenantFilteringDocumentRetriever.java.tmpl", filepath.Join("knowledge", "TenantFilteringDocumentRetriever.java")},
	}
	if g.config.HasVectorStore() {
		knowledgeFiles = append(knowledgeFiles,
			struct{ tmpl, out string }{"java/aiagent/knowledge/VectorKnowledgeRetriever.java.tmpl", filepath.Join("knowledge", "VectorKnowledgeRetriever.java")},
			struct{ tmpl, out string }{"java/aiagent/knowledge/EmbeddingService.java.tmpl", filepath.Join("knowledge", "EmbeddingService.java")},
			struct{ tmpl, out string }{"java/aiagent/knowledge/DocumentIngestionService.java.tmpl", filepath.Join("knowledge", "DocumentIngestionService.java")},
		)
	}
	for _, f := range knowledgeFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// ─── Protocol (Controllers) ─────────────────────────────────────────
	protocolFiles := []struct{ tmpl, out string }{
		{"java/aiagent/protocol/AgentRestController.java.tmpl", filepath.Join("protocol", "AgentRestController.java")},
		{"java/aiagent/protocol/A2AController.java.tmpl", filepath.Join("protocol", "A2AController.java")},
		{"java/aiagent/protocol/DiscoveryController.java.tmpl", filepath.Join("protocol", "DiscoveryController.java")},
		{"java/aiagent/protocol/StreamingController.java.tmpl", filepath.Join("protocol", "StreamingController.java")},
		{"java/aiagent/protocol/WebhookController.java.tmpl", filepath.Join("protocol", "WebhookController.java")},
	}
	if g.config.HasVectorStore() {
		// Ingestion REST endpoint (POST /ingest, /ingest/batch). Conditional
		// because the underlying DocumentIngestionService only loads when a
		// VectorStore bean is wired.
		protocolFiles = append(protocolFiles,
			struct{ tmpl, out string }{"java/aiagent/protocol/IngestionController.java.tmpl", filepath.Join("protocol", "IngestionController.java")},
		)
	}
	for _, f := range protocolFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// ─── Task Manager ───────────────────────────────────────────────────
	taskFiles := []struct{ tmpl, out string }{
		{"java/aiagent/task/TaskRecord.java.tmpl", filepath.Join("task", "TaskRecord.java")},
		{"java/aiagent/task/TaskEvent.java.tmpl", filepath.Join("task", "TaskEvent.java")},
		{"java/aiagent/task/TaskManager.java.tmpl", filepath.Join("task", "TaskManager.java")},
	}
	for _, f := range taskFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// ─── Event (Webhooks) ───────────────────────────────────────────────
	eventFiles := []struct{ tmpl, out string }{
		{"java/aiagent/event/WebhookRegistration.java.tmpl", filepath.Join("event", "WebhookRegistration.java")},
		{"java/aiagent/event/WebhookManager.java.tmpl", filepath.Join("event", "WebhookManager.java")},
	}
	for _, f := range eventFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// ─── Model DTOs (in Model module) ───────────────────────────────────
	modelDtoFiles := []struct{ tmpl, out string }{
		{"java/aiagent/model/JsonRpcRequest.java.tmpl", filepath.Join("dto", "JsonRpcRequest.java")},
		{"java/aiagent/model/JsonRpcResponse.java.tmpl", filepath.Join("dto", "JsonRpcResponse.java")},
		{"java/aiagent/model/ChatRequest.java.tmpl", filepath.Join("dto", "ChatRequest.java")},
		{"java/aiagent/model/ChatResponse.java.tmpl", filepath.Join("dto", "ChatResponse.java")},
		{"java/aiagent/model/AskRequest.java.tmpl", filepath.Join("dto", "AskRequest.java")},
		{"java/aiagent/model/AskResponse.java.tmpl", filepath.Join("dto", "AskResponse.java")},
		{"java/aiagent/model/WebhookRegisterRequest.java.tmpl", filepath.Join("dto", "WebhookRegisterRequest.java")},
	}
	for _, f := range modelDtoFiles {
		if err := g.writeTemplate(f.tmpl, g.javaPath("Model", f.out)); err != nil {
			return fmt.Errorf("failed to generate %s: %w", f.out, err)
		}
	}

	// ─── Resources ──────────────────────────────────────────────────────
	if err := g.writeTemplate(
		"java/aiagent/resources/application.yml.tmpl",
		g.resourcePath("AIAgent", "application.yml"),
	); err != nil {
		return fmt.Errorf("failed to generate AIAgent application.yml: %w", err)
	}

	if err := g.writeTemplate(
		"java/aiagent/resources/application-local-dev.yml.tmpl",
		g.resourcePath("AIAgent", "application-local-dev.yml"),
	); err != nil {
		return fmt.Errorf("failed to generate AIAgent application-local-dev.yml: %w", err)
	}

	if err := g.writeTemplate(
		"java/aiagent/resources/logback-spring.xml.tmpl",
		g.resourcePath("AIAgent", "logback-spring.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate AIAgent logback-spring.xml: %w", err)
	}

	if err := g.writeTemplate(
		"java/aiagent/resources/agent.json.tmpl",
		filepath.Join("AIAgent", "src", "main", "resources", ".well-known", "agent.json"),
	); err != nil {
		return fmt.Errorf("failed to generate agent.json: %w", err)
	}

	// ─── Docker ─────────────────────────────────────────────────────────
	if err := g.writeTemplate(
		"docker/aiagent.Dockerfile.tmpl",
		filepath.Join("AIAgent", "Dockerfile"),
	); err != nil {
		return fmt.Errorf("failed to generate AIAgent Dockerfile: %w", err)
	}

	// ─── Tests ──────────────────────────────────────────────────────────
	testFiles := []struct{ tmpl, out string }{
		{"java/aiagent/test/CallerIdentityTest.java.tmpl", filepath.Join("security", "CallerIdentityTest.java")},
		{"java/aiagent/test/ScopeEnforcerTest.java.tmpl", filepath.Join("security", "ScopeEnforcerTest.java")},
		{"java/aiagent/test/RateLimiterTest.java.tmpl", filepath.Join("security", "RateLimiterTest.java")},
		{"java/aiagent/test/OutputGuardrailTest.java.tmpl", filepath.Join("security", "OutputGuardrailTest.java")},
		{"java/aiagent/test/CorrelationIdFilterTest.java.tmpl", filepath.Join("security", "CorrelationIdFilterTest.java")},
		// F-AIAGENT-06 / F-INFRA-15: regression cases for the MCP
		// authorization filter (anonymous → 403, public-tier → 403,
		// partner-tier → pass, JWT with SCOPE_mcp:invoke → pass).
		{"java/aiagent/test/AgentMcpAuthorizationFilterTest.java.tmpl", filepath.Join("security", "AgentMcpAuthorizationFilterTest.java")},
		{"java/aiagent/test/ScratchpadTest.java.tmpl", filepath.Join("brain", "ScratchpadTest.java")},
		{"java/aiagent/test/ReflectionDecisionTest.java.tmpl", filepath.Join("brain", "ReflectionDecisionTest.java")},
		{"java/aiagent/test/PlaceholderToolsTest.java.tmpl", filepath.Join("tool", "PlaceholderToolsTest.java")},
		{"java/aiagent/test/TaskManagerTest.java.tmpl", filepath.Join("task", "TaskManagerTest.java")},
		{"java/aiagent/test/KeywordKnowledgeRetrieverTest.java.tmpl", filepath.Join("knowledge", "KeywordKnowledgeRetrieverTest.java")},
		// F-AIAGENT-04: regression cases for the retrieved-chunk
		// fencing pipeline (close-tag bleed, role-token neutering,
		// source-attribute escaping).
		{"java/aiagent/test/FencingDocumentRetrieverTest.java.tmpl", filepath.Join("knowledge", "FencingDocumentRetrieverTest.java")},
		// F-WEB-01 ArchUnit guard for the AIAgent module — every
		// REST/protocol controller endpoint must declare an explicit
		// authorization decision (recognises both @PreAuthorize and
		// the legacy @RequireScope marker).
		{"java/aiagent/test/AgentArchitectureTest.java.tmpl", "AgentArchitectureTest.java"},
	}
	// PGVector + Postgres + SQLDatastore is the only combination where
	// the integration test can compile and boot — the test imports
	// PostgreSQLContainer and relies on the V1 vector migration. Other
	// vector-store flavors (qdrant, mongodb-atlas) get their own
	// integration scaffolding in a future phase.
	if g.config.VectorStoreIsPgVector() && g.config.HasModule(config.ModuleSQLDatastore) && g.config.Database == config.DatabasePostgreSQL {
		testFiles = append(testFiles,
			struct{ tmpl, out string }{"java/aiagent/test/VectorRagIntegrationTest.java.tmpl", filepath.Join("knowledge", "VectorRagIntegrationTest.java")},
			// AgentWiringIntegrationTest guards against the
			// @ConditionalOnBean(ChatModel.class) /
			// @ConditionalOnBean(VectorStore.class) anti-pattern that
			// previously left PrimaryAgent / SpecialistAgent /
			// InputGuardrailAdvisor / IngestionController unregistered.
			// Boots the full Spring context against a real Postgres
			// + pgvector container, asserts each bean is in the
			// context, and POSTs to /ingest to confirm the controller
			// actually serves traffic. Same gating as
			// VectorRagIntegrationTest because both rely on the
			// pgvector Testcontainer image.
			struct{ tmpl, out string }{"java/aiagent/test/AgentWiringIntegrationTest.java.tmpl", "AgentWiringIntegrationTest.java"},
			// VectorTenantIsolationTest guards F-AIAGENT-03 / F-DATA-08:
			// ingest as tenant-A, retrieve as tenant-B, assert B sees
			// nothing of A's content. Same gating as the other vector
			// integration tests — needs the pgvector Testcontainer
			// image and the V1 vector schema migration.
			struct{ tmpl, out string }{"java/aiagent/test/VectorTenantIsolationTest.java.tmpl", filepath.Join("knowledge", "VectorTenantIsolationTest.java")},
			// AuthChainIntegrationTest boots with trabuco.auth.enabled=true
			// so the OAuth2 resource-server filter chain is exercised
			// end-to-end (no-JWT → 401, mockJwt with public scope →
			// 200, mockJwt with insufficient scope on /ingest → 403).
			// Same gating as the other AIAgent integration tests so it
			// shares the pgvector Testcontainer image; the auth chain
			// itself does not require pgvector, but the AIAgent
			// application context boot does.
			struct{ tmpl, out string }{"java/aiagent/test/AuthChainIntegrationTest.java.tmpl", filepath.Join("config", "security", "AuthChainIntegrationTest.java")},
		)
	}
	for _, f := range testFiles {
		if err := g.writeTemplate(f.tmpl, g.testJavaPath("AIAgent", f.out)); err != nil {
			return fmt.Errorf("failed to generate test %s: %w", f.out, err)
		}
	}

	// ─── IntelliJ Run Configuration ─────────────────────────────────────
	if err := g.writeTemplate(
		"idea/run/AIAgent__Maven_.run.xml.tmpl",
		filepath.Join(".run", "AIAgent.run.xml"),
	); err != nil {
		return fmt.Errorf("failed to generate AIAgent run configuration: %w", err)
	}

	return nil
}
