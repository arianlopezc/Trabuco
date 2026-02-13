package migrate

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProjectScanner_ParsePOM(t *testing.T) {
	// Create temp directory with pom.xml
	tempDir := t.TempDir()

	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <groupId>com.example</groupId>
    <artifactId>legacy-app</artifactId>
    <version>1.0.0</version>
    <properties>
        <java.version>17</java.version>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-data-jpa</artifactId>
        </dependency>
        <dependency>
            <groupId>org.postgresql</groupId>
            <artifactId>postgresql</artifactId>
            <scope>runtime</scope>
        </dependency>
    </dependencies>
</project>`

	if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
		t.Fatalf("failed to create pom.xml: %v", err)
	}

	// Create minimal src directory
	srcDir := filepath.Join(tempDir, "src", "main", "java")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	scanner := NewProjectScanner(tempDir)
	info := &ProjectInfo{}
	err := scanner.parsePOM(info)

	if err != nil {
		t.Fatalf("parsePOM() error = %v", err)
	}

	if info.GroupID != "com.example" {
		t.Errorf("GroupID = %v, want com.example", info.GroupID)
	}

	if info.ArtifactID != "legacy-app" {
		t.Errorf("ArtifactID = %v, want legacy-app", info.ArtifactID)
	}

	if info.JavaVersion != "17" {
		t.Errorf("JavaVersion = %v, want 17", info.JavaVersion)
	}

	if info.SpringBootVersion != "3.2.0" {
		t.Errorf("SpringBootVersion = %v, want 3.2.0", info.SpringBootVersion)
	}

	if len(info.Dependencies) != 2 {
		t.Errorf("len(Dependencies) = %v, want 2", len(info.Dependencies))
	}
}

func TestProjectScanner_ParsePOMInheritedGroupID(t *testing.T) {
	tempDir := t.TempDir()

	// POM with no groupId (inherits from parent)
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <parent>
        <groupId>com.company</groupId>
        <artifactId>parent-pom</artifactId>
        <version>1.0.0</version>
    </parent>
    <artifactId>child-app</artifactId>
    <properties>
        <java.version>21</java.version>
    </properties>
</project>`

	if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
		t.Fatalf("failed to create pom.xml: %v", err)
	}

	srcDir := filepath.Join(tempDir, "src", "main", "java")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	scanner := NewProjectScanner(tempDir)
	info := &ProjectInfo{}
	err := scanner.parsePOM(info)

	if err != nil {
		t.Fatalf("parsePOM() error = %v", err)
	}

	// Should inherit groupId from parent
	if info.GroupID != "com.company" {
		t.Errorf("GroupID = %v, want com.company (inherited from parent)", info.GroupID)
	}
}

func TestProjectScanner_ParseJavaFile(t *testing.T) {
	tempDir := t.TempDir()

	javaContent := `package com.example.entity;

import javax.persistence.Entity;
import javax.persistence.Id;
import javax.persistence.Table;
import lombok.Data;

@Entity
@Table(name = "users")
@Data
public class User {
    @Id
    private Long id;
    private String name;
    private String email;
}
`

	srcDir := filepath.Join(tempDir, "src", "main", "java", "com", "example", "entity")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	javaFile := filepath.Join(srcDir, "User.java")
	if err := os.WriteFile(javaFile, []byte(javaContent), 0644); err != nil {
		t.Fatalf("failed to create Java file: %v", err)
	}

	scanner := NewProjectScanner(tempDir)
	class, err := scanner.parseJavaFile(javaFile)

	if err != nil {
		t.Fatalf("parseJavaFile() error = %v", err)
	}

	if class.Name != "User" {
		t.Errorf("Name = %v, want User", class.Name)
	}

	if class.Package != "com.example.entity" {
		t.Errorf("Package = %v, want com.example.entity", class.Package)
	}

	// Check annotations
	hasEntity := false
	hasTable := false
	hasData := false
	for _, ann := range class.Annotations {
		switch ann {
		case "Entity":
			hasEntity = true
		case "Table":
			hasTable = true
		case "Data":
			hasData = true
		}
	}

	if !hasEntity {
		t.Error("missing @Entity annotation")
	}
	if !hasTable {
		t.Error("missing @Table annotation")
	}
	if !hasData {
		t.Error("missing @Data annotation")
	}
}

func TestProjectScanner_CategorizeClass(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantEntity  bool
		wantRepo    bool
		wantService bool
		wantCtrl    bool
		wantJob     bool
		wantEvent   bool
	}{
		{
			name: "JPA Entity",
			content: `package com.example;
import javax.persistence.Entity;
@Entity
public class User {}`,
			wantEntity: true,
		},
		{
			name: "MongoDB Document",
			content: `package com.example;
import org.springframework.data.mongodb.core.mapping.Document;
@Document
public class UserDocument {}`,
			wantEntity: true,
		},
		{
			name: "JPA Repository",
			content: `package com.example;
import org.springframework.data.jpa.repository.JpaRepository;
public interface UserRepository extends JpaRepository<User, Long> {}`,
			wantRepo: true,
		},
		{
			name: "CrudRepository",
			content: `package com.example;
import org.springframework.data.repository.CrudRepository;
public interface UserRepository extends CrudRepository<User, Long> {}`,
			wantRepo: true,
		},
		{
			name: "Service class",
			content: `package com.example;
import org.springframework.stereotype.Service;
@Service
public class UserService {}`,
			wantService: true,
		},
		{
			name: "RestController",
			content: `package com.example;
import org.springframework.web.bind.annotation.RestController;
@RestController
public class UserController {}`,
			wantCtrl: true,
		},
		{
			name: "Scheduled job",
			content: `package com.example;
import org.springframework.scheduling.annotation.Scheduled;
public class DataCleanupJob {
    @Scheduled(cron = "0 0 * * * *")
    public void cleanup() {}
}`,
			wantJob: true,
		},
		{
			name: "Kafka listener",
			content: `package com.example;
import org.springframework.kafka.annotation.KafkaListener;
public class UserEventListener {
    @KafkaListener(topics = "users")
    public void handle(UserEvent event) {}
}`,
			wantEvent: true,
		},
		{
			name: "RabbitMQ listener",
			content: `package com.example;
import org.springframework.amqp.rabbit.annotation.RabbitListener;
public class OrderEventListener {
    @RabbitListener(queues = "orders")
    public void handle(OrderEvent event) {}
}`,
			wantEvent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			srcDir := filepath.Join(tempDir, "src", "main", "java", "com", "example")
			if err := os.MkdirAll(srcDir, 0755); err != nil {
				t.Fatalf("failed to create src dir: %v", err)
			}

			javaFile := filepath.Join(srcDir, "Test.java")
			if err := os.WriteFile(javaFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create Java file: %v", err)
			}

			scanner := NewProjectScanner(tempDir)
			class, err := scanner.parseJavaFile(javaFile)
			if err != nil {
				t.Fatalf("parseJavaFile() error = %v", err)
			}

			info := &ProjectInfo{}
			scanner.categorizeClass(info, class)

			gotEntity := len(info.Entities) > 0
			gotRepo := len(info.Repositories) > 0
			gotService := len(info.Services) > 0
			gotCtrl := len(info.Controllers) > 0
			gotJob := len(info.ScheduledJobs) > 0
			gotEvent := len(info.EventListeners) > 0

			if gotEntity != tt.wantEntity {
				t.Errorf("entity detection: got %v, want %v", gotEntity, tt.wantEntity)
			}
			if gotRepo != tt.wantRepo {
				t.Errorf("repository detection: got %v, want %v", gotRepo, tt.wantRepo)
			}
			if gotService != tt.wantService {
				t.Errorf("service detection: got %v, want %v", gotService, tt.wantService)
			}
			if gotCtrl != tt.wantCtrl {
				t.Errorf("controller detection: got %v, want %v", gotCtrl, tt.wantCtrl)
			}
			if gotJob != tt.wantJob {
				t.Errorf("job detection: got %v, want %v", gotJob, tt.wantJob)
			}
			if gotEvent != tt.wantEvent {
				t.Errorf("event listener detection: got %v, want %v", gotEvent, tt.wantEvent)
			}
		})
	}
}

func TestProjectScanner_DetectMessageBroker(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []Dependency
		wantBroker   string
	}{
		{
			name: "Kafka",
			dependencies: []Dependency{
				{GroupID: "org.springframework.kafka", ArtifactID: "spring-kafka"},
			},
			wantBroker: "kafka",
		},
		{
			name: "RabbitMQ",
			dependencies: []Dependency{
				{GroupID: "org.springframework.amqp", ArtifactID: "spring-amqp"},
			},
			wantBroker: "rabbitmq",
		},
		{
			name: "RabbitMQ via AMQP",
			dependencies: []Dependency{
				{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-amqp"},
			},
			wantBroker: "rabbitmq",
		},
		{
			name: "SQS",
			dependencies: []Dependency{
				{GroupID: "io.awspring.cloud", ArtifactID: "spring-cloud-aws-sqs"},
			},
			wantBroker: "sqs",
		},
		{
			name: "Pub/Sub",
			dependencies: []Dependency{
				{GroupID: "com.google.cloud", ArtifactID: "spring-cloud-gcp-pubsub"},
			},
			wantBroker: "pubsub",
		},
		{
			name:         "No broker",
			dependencies: []Dependency{},
			wantBroker:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			scanner := NewProjectScanner(tempDir)
			info := &ProjectInfo{Dependencies: tt.dependencies}

			broker := scanner.detectMessageBroker(info)

			if broker != tt.wantBroker {
				t.Errorf("detectMessageBroker() = %v, want %v", broker, tt.wantBroker)
			}
		})
	}
}

func TestProjectScanner_DetectNoSQL(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []Dependency
		entities     []*JavaClass
		wantNoSQL    bool
	}{
		{
			name: "MongoDB dependency",
			dependencies: []Dependency{
				{GroupID: "org.springframework.data", ArtifactID: "spring-data-mongodb"},
			},
			wantNoSQL: true,
		},
		{
			name: "Document annotation without MongoDB dependency",
			dependencies: []Dependency{},
			entities: []*JavaClass{
				{Annotations: []string{"Document", "Id"}},
			},
			wantNoSQL: false, // Document annotation alone is not enough - need explicit MongoDB dependency
		},
		{
			name: "Document annotation with MongoDB dependency",
			dependencies: []Dependency{
				{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-data-mongodb"},
			},
			entities: []*JavaClass{
				{Annotations: []string{"Document", "Id"}},
			},
			wantNoSQL: true,
		},
		{
			name: "JPA only",
			dependencies: []Dependency{
				{GroupID: "org.springframework.data", ArtifactID: "spring-data-jpa"},
			},
			wantNoSQL: false,
		},
		{
			name: "PostgreSQL takes priority over MongoDB when both present",
			dependencies: []Dependency{
				{GroupID: "org.postgresql", ArtifactID: "postgresql"},
				{GroupID: "org.springframework.data", ArtifactID: "spring-data-mongodb"},
			},
			wantNoSQL: false, // SQL database takes priority
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			scanner := NewProjectScanner(tempDir)
			info := &ProjectInfo{
				Dependencies: tt.dependencies,
				Entities:     tt.entities,
			}

			gotNoSQL := scanner.detectNoSQL(info)

			if gotNoSQL != tt.wantNoSQL {
				t.Errorf("detectNoSQL() = %v, want %v", gotNoSQL, tt.wantNoSQL)
			}
		})
	}
}

func TestProjectScanner_DetectDatabase(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []Dependency
		wantDatabase string
	}{
		{
			name: "PostgreSQL dependency",
			dependencies: []Dependency{
				{GroupID: "org.postgresql", ArtifactID: "postgresql"},
			},
			wantDatabase: "postgresql",
		},
		{
			name: "MySQL dependency",
			dependencies: []Dependency{
				{GroupID: "mysql", ArtifactID: "mysql-connector-java"},
			},
			wantDatabase: "mysql",
		},
		{
			name: "MongoDB dependency",
			dependencies: []Dependency{
				{GroupID: "org.springframework.data", ArtifactID: "spring-data-mongodb"},
			},
			wantDatabase: "mongodb",
		},
		{
			name: "Oracle dependency",
			dependencies: []Dependency{
				{GroupID: "com.oracle.database.jdbc", ArtifactID: "ojdbc8"},
			},
			wantDatabase: "oracle",
		},
		{
			name: "SQL Server dependency",
			dependencies: []Dependency{
				{GroupID: "com.microsoft.sqlserver", ArtifactID: "mssql-jdbc"},
			},
			wantDatabase: "sqlserver",
		},
		{
			name: "No database dependency defaults to postgresql",
			dependencies: []Dependency{
				{GroupID: "org.springframework.boot", ArtifactID: "spring-boot-starter-web"},
			},
			wantDatabase: "postgresql",
		},
		{
			name: "PostgreSQL takes priority when multiple present",
			dependencies: []Dependency{
				{GroupID: "org.postgresql", ArtifactID: "postgresql"},
				{GroupID: "org.springframework.data", ArtifactID: "spring-data-mongodb"},
			},
			wantDatabase: "postgresql",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			scanner := NewProjectScanner(tempDir)
			info := &ProjectInfo{
				Dependencies: tt.dependencies,
			}

			gotDatabase := scanner.detectDatabase(info)

			if gotDatabase != tt.wantDatabase {
				t.Errorf("detectDatabase() = %v, want %v", gotDatabase, tt.wantDatabase)
			}
		})
	}
}

func TestProjectScanner_DetectRedis(t *testing.T) {
	tests := []struct {
		name         string
		dependencies []Dependency
		wantRedis    bool
	}{
		{
			name: "Redis dependency",
			dependencies: []Dependency{
				{GroupID: "org.springframework.data", ArtifactID: "spring-data-redis"},
			},
			wantRedis: true,
		},
		{
			name:         "No Redis",
			dependencies: []Dependency{},
			wantRedis:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			scanner := NewProjectScanner(tempDir)
			info := &ProjectInfo{Dependencies: tt.dependencies}

			gotRedis := scanner.detectRedis(info)

			if gotRedis != tt.wantRedis {
				t.Errorf("detectRedis() = %v, want %v", gotRedis, tt.wantRedis)
			}
		})
	}
}

func TestProjectScanner_DetectProjectType(t *testing.T) {
	tests := []struct {
		name              string
		dependencies      []Dependency
		springBootVersion string
		wantType          string
	}{
		{
			name:              "Spring Boot 3.2.0",
			springBootVersion: "3.2.0",
			wantType:          "Spring Boot 3.2.0 (Maven)",
		},
		{
			name: "Plain Spring",
			dependencies: []Dependency{
				{GroupID: "org.springframework", ArtifactID: "spring-core"},
			},
			wantType: "Spring",
		},
		{
			name:         "Plain Java Maven",
			dependencies: []Dependency{},
			wantType:     "Java Maven",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			scanner := NewProjectScanner(tempDir)
			info := &ProjectInfo{
				Dependencies:      tt.dependencies,
				SpringBootVersion: tt.springBootVersion,
			}

			gotType := scanner.detectProjectType(info)

			if gotType != tt.wantType {
				t.Errorf("detectProjectType() = %v, want %v", gotType, tt.wantType)
			}
		})
	}
}

func TestProjectScanner_FullScan(t *testing.T) {
	tempDir := t.TempDir()

	// Create pom.xml
	pomContent := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0">
    <parent>
        <groupId>org.springframework.boot</groupId>
        <artifactId>spring-boot-starter-parent</artifactId>
        <version>3.2.0</version>
    </parent>
    <groupId>com.example</groupId>
    <artifactId>test-app</artifactId>
    <properties>
        <java.version>21</java.version>
    </properties>
    <dependencies>
        <dependency>
            <groupId>org.springframework.boot</groupId>
            <artifactId>spring-boot-starter-data-jpa</artifactId>
        </dependency>
    </dependencies>
</project>`
	if err := os.WriteFile(filepath.Join(tempDir, "pom.xml"), []byte(pomContent), 0644); err != nil {
		t.Fatalf("failed to create pom.xml: %v", err)
	}

	// Create entity
	entityDir := filepath.Join(tempDir, "src", "main", "java", "com", "example", "entity")
	if err := os.MkdirAll(entityDir, 0755); err != nil {
		t.Fatalf("failed to create entity dir: %v", err)
	}
	entityContent := `package com.example.entity;
import javax.persistence.Entity;
@Entity
public class User {}`
	if err := os.WriteFile(filepath.Join(entityDir, "User.java"), []byte(entityContent), 0644); err != nil {
		t.Fatalf("failed to create entity: %v", err)
	}

	// Create service
	serviceDir := filepath.Join(tempDir, "src", "main", "java", "com", "example", "service")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		t.Fatalf("failed to create service dir: %v", err)
	}
	serviceContent := `package com.example.service;
import org.springframework.stereotype.Service;
@Service
public class UserService {}`
	if err := os.WriteFile(filepath.Join(serviceDir, "UserService.java"), []byte(serviceContent), 0644); err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	// Create controller
	controllerDir := filepath.Join(tempDir, "src", "main", "java", "com", "example", "controller")
	if err := os.MkdirAll(controllerDir, 0755); err != nil {
		t.Fatalf("failed to create controller dir: %v", err)
	}
	controllerContent := `package com.example.controller;
import org.springframework.web.bind.annotation.RestController;
@RestController
public class UserController {}`
	if err := os.WriteFile(filepath.Join(controllerDir, "UserController.java"), []byte(controllerContent), 0644); err != nil {
		t.Fatalf("failed to create controller: %v", err)
	}

	// Run scan
	scanner := NewProjectScanner(tempDir)
	info, err := scanner.Scan()

	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	// Verify results
	if info.Name != "test-app" {
		t.Errorf("Name = %v, want test-app", info.Name)
	}

	if info.GroupID != "com.example" {
		t.Errorf("GroupID = %v, want com.example", info.GroupID)
	}

	if info.JavaVersion != "21" {
		t.Errorf("JavaVersion = %v, want 21", info.JavaVersion)
	}

	if len(info.Entities) != 1 {
		t.Errorf("len(Entities) = %v, want 1", len(info.Entities))
	}

	if len(info.Services) != 1 {
		t.Errorf("len(Services) = %v, want 1", len(info.Services))
	}

	if len(info.Controllers) != 1 {
		t.Errorf("len(Controllers) = %v, want 1", len(info.Controllers))
	}

	if info.HasScheduledJobs {
		t.Error("HasScheduledJobs should be false")
	}

	if info.HasEventListeners {
		t.Error("HasEventListeners should be false")
	}
}
