package migrate

import (
	"strings"
)

// DependencyReport contains the analysis of project dependencies
type DependencyReport struct {
	// Compatible dependencies that work as-is
	Compatible []Dependency

	// Replaceable dependencies that have Trabuco alternatives
	Replaceable []*ReplaceableDependency

	// Unsupported dependencies that require manual handling
	Unsupported []string
}

// ReplaceableDependency represents a dependency that can be replaced
type ReplaceableDependency struct {
	// Source is the original dependency identifier
	Source string

	// SourceDependency is the full dependency info
	SourceDependency Dependency

	// TrabucoAlternative is what Trabuco uses instead
	TrabucoAlternative string

	// MigrationImpact describes what will change
	MigrationImpact string

	// MigrationComplexity: low, medium, high
	MigrationComplexity string

	// Accepted indicates if user accepted this replacement
	Accepted bool
}

// DependencyAnalyzer analyzes project dependencies
type DependencyAnalyzer struct {
	// Mapping of known dependencies to their Trabuco alternatives
	replacements map[string]*DependencyReplacement

	// Known compatible dependencies
	compatible map[string]bool

	// Known unsupported patterns
	unsupported []string
}

// DependencyReplacement defines how to replace a dependency
type DependencyReplacement struct {
	Alternative string
	Impact      string
	Complexity  string
}

// NewDependencyAnalyzer creates a new analyzer
func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{
		replacements: buildReplacementMap(),
		compatible:   buildCompatibleMap(),
		unsupported:  buildUnsupportedPatterns(),
	}
}

// Analyze analyzes dependencies and returns a report
func (a *DependencyAnalyzer) Analyze(dependencies []Dependency) *DependencyReport {
	report := &DependencyReport{}

	for _, dep := range dependencies {
		depKey := dep.GroupID + ":" + dep.ArtifactID

		// Check if it's a known replaceable dependency
		if replacement, ok := a.findReplacement(depKey); ok {
			report.Replaceable = append(report.Replaceable, &ReplaceableDependency{
				Source:              depKey,
				SourceDependency:    dep,
				TrabucoAlternative:  replacement.Alternative,
				MigrationImpact:     replacement.Impact,
				MigrationComplexity: replacement.Complexity,
			})
			continue
		}

		// Check if it's compatible
		if a.isCompatible(depKey) {
			report.Compatible = append(report.Compatible, dep)
			continue
		}

		// Check if it's unsupported
		if a.isUnsupported(depKey) {
			report.Unsupported = append(report.Unsupported, depKey)
			continue
		}

		// Default: assume compatible (most Spring Boot starters work)
		report.Compatible = append(report.Compatible, dep)
	}

	return report
}

func (a *DependencyAnalyzer) findReplacement(depKey string) (*DependencyReplacement, bool) {
	// Check exact match
	if r, ok := a.replacements[depKey]; ok {
		return r, true
	}

	// Check partial match
	for pattern, r := range a.replacements {
		if strings.Contains(depKey, pattern) {
			return r, true
		}
	}

	return nil, false
}

func (a *DependencyAnalyzer) isCompatible(depKey string) bool {
	if a.compatible[depKey] {
		return true
	}

	// Check patterns
	compatiblePatterns := []string{
		"spring-boot-starter",
		"jackson",
		"slf4j",
		"logback",
		"micrometer",
		"springdoc",
		"testcontainers",
		"junit",
		"mockito",
		"assertj",
	}

	for _, pattern := range compatiblePatterns {
		if strings.Contains(depKey, pattern) {
			return true
		}
	}

	return false
}

func (a *DependencyAnalyzer) isUnsupported(depKey string) bool {
	for _, pattern := range a.unsupported {
		if strings.Contains(depKey, pattern) {
			return true
		}
	}
	return false
}

func buildReplacementMap() map[string]*DependencyReplacement {
	return map[string]*DependencyReplacement{
		// ORM replacements
		"org.hibernate:hibernate-core": {
			Alternative: "Spring Data JDBC",
			Impact:      "Entity annotations will change (@Entity → @Table, no lazy loading, no @OneToMany/@ManyToOne)",
			Complexity:  "medium",
		},
		"hibernate-core": {
			Alternative: "Spring Data JDBC",
			Impact:      "Entity annotations will change (@Entity → @Table, no lazy loading, no @OneToMany/@ManyToOne)",
			Complexity:  "medium",
		},
		"spring-boot-starter-data-jpa": {
			Alternative: "spring-boot-starter-data-jdbc",
			Impact:      "JPA repositories will be converted to Spring Data JDBC repositories",
			Complexity:  "medium",
		},
		"spring-data-jpa": {
			Alternative: "spring-data-jdbc",
			Impact:      "JPA repositories will be converted to Spring Data JDBC repositories",
			Complexity:  "medium",
		},

		// Scheduler replacements
		"org.quartz-scheduler:quartz": {
			Alternative: "JobRunr 8.4.0",
			Impact:      "Quartz jobs will be converted to JobRunr job requests and handlers",
			Complexity:  "medium",
		},
		"quartz": {
			Alternative: "JobRunr 8.4.0",
			Impact:      "Quartz jobs will be converted to JobRunr job requests and handlers",
			Complexity:  "medium",
		},

		// Migration tool replacements
		"org.liquibase:liquibase-core": {
			Alternative: "Flyway",
			Impact:      "Liquibase changesets will need manual conversion to Flyway SQL migrations",
			Complexity:  "low",
		},
		"liquibase": {
			Alternative: "Flyway",
			Impact:      "Liquibase changesets will need manual conversion to Flyway SQL migrations",
			Complexity:  "low",
		},

		// Code generation replacements
		"org.projectlombok:lombok": {
			Alternative: "Immutables (for DTOs) + standard getters/setters",
			Impact:      "Lombok annotations will be expanded to full implementations",
			Complexity:  "medium",
		},
		"lombok": {
			Alternative: "Immutables (for DTOs) + standard getters/setters",
			Impact:      "Lombok annotations will be expanded to full implementations",
			Complexity:  "medium",
		},

		// Documentation replacements
		"io.springfox:springfox-swagger2": {
			Alternative: "SpringDoc OpenAPI 2.7.0",
			Impact:      "Swagger annotations will be converted to OpenAPI 3 annotations",
			Complexity:  "low",
		},
		"springfox": {
			Alternative: "SpringDoc OpenAPI 2.7.0",
			Impact:      "Swagger annotations will be converted to OpenAPI 3 annotations",
			Complexity:  "low",
		},

		// Query DSL replacements
		"org.mybatis:mybatis": {
			Alternative: "Spring Data JDBC",
			Impact:      "MyBatis mappers will need manual conversion to repositories",
			Complexity:  "high",
		},
		"mybatis": {
			Alternative: "Spring Data JDBC",
			Impact:      "MyBatis mappers will need manual conversion to repositories",
			Complexity:  "high",
		},
		"org.jooq:jooq": {
			Alternative: "Spring Data JDBC",
			Impact:      "JOOQ queries will need manual conversion to repository methods",
			Complexity:  "high",
		},
		"jooq": {
			Alternative: "Spring Data JDBC",
			Impact:      "JOOQ queries will need manual conversion to repository methods",
			Complexity:  "high",
		},

		// Message broker replacements
		"org.apache.activemq:activemq": {
			Alternative: "RabbitMQ or Kafka",
			Impact:      "ActiveMQ listeners will be converted to RabbitMQ/Kafka listeners",
			Complexity:  "medium",
		},
		"activemq": {
			Alternative: "RabbitMQ or Kafka",
			Impact:      "ActiveMQ listeners will be converted to RabbitMQ/Kafka listeners",
			Complexity:  "medium",
		},

		// AWS SDK
		"com.amazonaws:aws-java-sdk": {
			Alternative: "AWS SDK v2 (via Spring Cloud AWS 3.2.0)",
			Impact:      "AWS SDK v1 calls will need conversion to v2 API",
			Complexity:  "medium",
		},
	}
}

func buildCompatibleMap() map[string]bool {
	return map[string]bool{
		// Spring Boot starters
		"org.springframework.boot:spring-boot-starter":            true,
		"org.springframework.boot:spring-boot-starter-web":        true,
		"org.springframework.boot:spring-boot-starter-validation": true,
		"org.springframework.boot:spring-boot-starter-actuator":   true,
		"org.springframework.boot:spring-boot-starter-test":       true,
		"org.springframework.boot:spring-boot-starter-data-jdbc":  true,

		// Jackson
		"com.fasterxml.jackson.core:jackson-databind":              true,
		"com.fasterxml.jackson.datatype:jackson-datatype-jdk8":     true,
		"com.fasterxml.jackson.module:jackson-module-parameter-names": true,

		// Logging
		"ch.qos.logback:logback-classic":                  true,
		"net.logstash.logback:logstash-logback-encoder":   true,
		"org.slf4j:slf4j-api":                             true,

		// Testing
		"org.junit.jupiter:junit-jupiter":        true,
		"org.mockito:mockito-core":               true,
		"org.assertj:assertj-core":               true,
		"org.testcontainers:testcontainers":      true,
		"org.testcontainers:junit-jupiter":       true,
		"org.testcontainers:postgresql":          true,
		"org.testcontainers:mongodb":             true,

		// Database
		"org.postgresql:postgresql":              true,
		"com.mysql:mysql-connector-j":            true,
		"org.flywaydb:flyway-core":               true,
		"com.zaxxer:HikariCP":                    true,

		// Messaging (supported)
		"org.springframework.kafka:spring-kafka": true,
		"org.springframework.amqp:spring-rabbit": true,

		// Monitoring
		"io.micrometer:micrometer-registry-prometheus": true,
	}
}

func buildUnsupportedPatterns() []string {
	return []string{
		// Internal/proprietary libraries
		"-internal",
		"-proprietary",

		// Legacy/deprecated
		"javax.servlet",
		"javax.ejb",
		"javax.jms",

		// Application servers
		"jboss",
		"weblogic",
		"websphere",

		// Non-Spring frameworks
		"struts",
		"jsf",
		"wicket",
		"vaadin",

		// Legacy ORM
		"openjpa",
		"eclipselink",
		"toplink",
	}
}

// GetMigrationSummary returns a summary of what will be migrated
func (r *DependencyReport) GetMigrationSummary() string {
	var sb strings.Builder

	sb.WriteString("Migration Summary:\n")
	sb.WriteString("─────────────────\n")

	if len(r.Compatible) > 0 {
		sb.WriteString("✓ Compatible dependencies: ")
		sb.WriteString(string(rune(len(r.Compatible))))
		sb.WriteString("\n")
	}

	if len(r.Replaceable) > 0 {
		sb.WriteString("⚠ Dependencies to replace:\n")
		for _, dep := range r.Replaceable {
			sb.WriteString("  - ")
			sb.WriteString(dep.Source)
			sb.WriteString(" → ")
			sb.WriteString(dep.TrabucoAlternative)
			sb.WriteString("\n")
		}
	}

	if len(r.Unsupported) > 0 {
		sb.WriteString("✗ Unsupported (manual handling required):\n")
		for _, dep := range r.Unsupported {
			sb.WriteString("  - ")
			sb.WriteString(dep)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// HasBlockers returns true if there are issues that would block migration
func (r *DependencyReport) HasBlockers() bool {
	// Check for high-complexity replacements that weren't accepted
	for _, dep := range r.Replaceable {
		if dep.MigrationComplexity == "high" && !dep.Accepted {
			return true
		}
	}
	return false
}
