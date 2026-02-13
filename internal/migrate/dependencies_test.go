package migrate

import (
	"testing"
)

func TestDependencyAnalyzer_Analyze(t *testing.T) {
	analyzer := NewDependencyAnalyzer()

	tests := []struct {
		name               string
		dependencies       []Dependency
		wantCompatible     int
		wantReplaceable    int
		wantUnsupportedMin int
	}{
		{
			name: "Spring Boot starter web is compatible",
			dependencies: []Dependency{
				{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-web"},
			},
			wantCompatible: 1,
		},
		{
			name: "Hibernate/JPA needs replacement",
			dependencies: []Dependency{
				{GroupID: "org.hibernate", ArtifactID: "hibernate-core"},
			},
			wantReplaceable: 1,
		},
		{
			name: "Spring Data JPA needs replacement",
			dependencies: []Dependency{
				{GroupID: "org.springframework.data", ArtifactID: "spring-data-jpa"},
			},
			wantReplaceable: 1,
		},
		{
			name: "Quartz needs replacement",
			dependencies: []Dependency{
				{GroupID: "org.quartz-scheduler", ArtifactID: "quartz"},
			},
			wantReplaceable: 1,
		},
		{
			name: "Lombok needs replacement",
			dependencies: []Dependency{
				{GroupID: "org.projectlombok", ArtifactID: "lombok"},
			},
			wantReplaceable: 1,
		},
		{
			name: "Mixed dependencies",
			dependencies: []Dependency{
				{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-web"},
				{GroupID: "org.hibernate", ArtifactID: "hibernate-core"},
				{GroupID: "org.projectlombok", ArtifactID: "lombok"},
			},
			wantCompatible:  1,
			wantReplaceable: 2,
		},
		{
			name:           "Empty dependencies",
			dependencies:   []Dependency{},
			wantCompatible: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := analyzer.Analyze(tt.dependencies)

			if len(report.Compatible) < tt.wantCompatible {
				t.Errorf("len(Compatible) = %v, want >= %v", len(report.Compatible), tt.wantCompatible)
			}

			if len(report.Replaceable) < tt.wantReplaceable {
				t.Errorf("len(Replaceable) = %v, want >= %v", len(report.Replaceable), tt.wantReplaceable)
			}

			if tt.wantUnsupportedMin > 0 && len(report.Unsupported) < tt.wantUnsupportedMin {
				t.Errorf("len(Unsupported) = %v, want >= %v", len(report.Unsupported), tt.wantUnsupportedMin)
			}
		})
	}
}

func TestDependencyAnalyzer_ReplaceableDetails(t *testing.T) {
	analyzer := NewDependencyAnalyzer()

	tests := []struct {
		name              string
		dependency        Dependency
		wantSource        string
		wantReplacePrefix string // Check prefix since versions may be included
	}{
		{
			name:              "Hibernate replacement",
			dependency:        Dependency{GroupID: "org.hibernate", ArtifactID: "hibernate-core"},
			wantSource:        "org.hibernate:hibernate-core",
			wantReplacePrefix: "Spring Data JDBC",
		},
		{
			name:              "Quartz replacement",
			dependency:        Dependency{GroupID: "org.quartz-scheduler", ArtifactID: "quartz"},
			wantSource:        "org.quartz-scheduler:quartz",
			wantReplacePrefix: "JobRunr",
		},
		{
			name:              "Lombok replacement",
			dependency:        Dependency{GroupID: "org.projectlombok", ArtifactID: "lombok"},
			wantSource:        "org.projectlombok:lombok",
			wantReplacePrefix: "Immutables",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report := analyzer.Analyze([]Dependency{tt.dependency})

			if len(report.Replaceable) == 0 {
				t.Fatal("expected at least one replaceable dependency")
			}

			replacement := report.Replaceable[0]

			if replacement.Source != tt.wantSource {
				t.Errorf("Source = %v, want %v", replacement.Source, tt.wantSource)
			}

			// Check that the alternative contains the expected prefix
			if len(replacement.TrabucoAlternative) < len(tt.wantReplacePrefix) ||
				replacement.TrabucoAlternative[:len(tt.wantReplacePrefix)] != tt.wantReplacePrefix {
				t.Errorf("TrabucoAlternative = %v, want prefix %v", replacement.TrabucoAlternative, tt.wantReplacePrefix)
			}

			if replacement.MigrationImpact == "" {
				t.Error("MigrationImpact should not be empty")
			}
		})
	}
}

func TestDependencyAnalyzer_TestScopeDependencies(t *testing.T) {
	analyzer := NewDependencyAnalyzer()

	// Test scope dependencies should be compatible
	deps := []Dependency{
		{GroupID: "org.junit.jupiter", ArtifactID: "junit-jupiter", Scope: "test"},
		{GroupID: "org.mockito", ArtifactID: "mockito-core", Scope: "test"},
	}

	report := analyzer.Analyze(deps)

	// Test dependencies should generally be compatible
	if len(report.Compatible) < 2 {
		t.Errorf("len(Compatible) = %v, want >= 2 for test dependencies", len(report.Compatible))
	}
}

func TestDependencyAnalyzer_ProvidedScopeDependencies(t *testing.T) {
	analyzer := NewDependencyAnalyzer()

	deps := []Dependency{
		{GroupID: "jakarta.servlet", ArtifactID: "jakarta.servlet-api", Scope: "provided"},
	}

	report := analyzer.Analyze(deps)

	// Provided dependencies should be handled
	total := len(report.Compatible) + len(report.Replaceable) + len(report.Unsupported)
	if total != 1 {
		t.Errorf("total categorized = %v, want 1", total)
	}
}

func TestDependencyAnalyzer_SpringBootStarters(t *testing.T) {
	analyzer := NewDependencyAnalyzer()

	starters := []Dependency{
		{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-web"},
		{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-actuator"},
		{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-validation"},
		{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-test", Scope: "test"},
	}

	report := analyzer.Analyze(starters)

	// All Spring Boot starters should be compatible
	if len(report.Compatible) != 4 {
		t.Errorf("len(Compatible) = %v, want 4 for Spring Boot starters", len(report.Compatible))
	}
}

func TestDependencyAnalyzer_DatabaseDrivers(t *testing.T) {
	analyzer := NewDependencyAnalyzer()

	drivers := []Dependency{
		{GroupID: "org.postgresql", ArtifactID: "postgresql", Scope: "runtime"},
		{GroupID: "com.mysql", ArtifactID: "mysql-connector-j", Scope: "runtime"},
	}

	report := analyzer.Analyze(drivers)

	// Database drivers should be compatible
	if len(report.Compatible) != 2 {
		t.Errorf("len(Compatible) = %v, want 2 for database drivers", len(report.Compatible))
	}
}

func TestDependencyAnalyzer_OldJavaxNamespace(t *testing.T) {
	analyzer := NewDependencyAnalyzer()

	// Old javax.* dependencies - the analyzer categorizes these
	deps := []Dependency{
		{GroupID: "javax.validation", ArtifactID: "validation-api"},
		{GroupID: "javax.persistence", ArtifactID: "javax.persistence-api"},
	}

	report := analyzer.Analyze(deps)

	// The analyzer should process these dependencies (either compatible, replaceable, or unsupported)
	total := len(report.Compatible) + len(report.Replaceable) + len(report.Unsupported)
	if total != 2 {
		t.Errorf("expected 2 total categorized dependencies, got %d", total)
	}
}
