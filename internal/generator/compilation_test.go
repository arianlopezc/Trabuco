//go:build integration

package generator

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// TestCompilation_* tests verify that generated projects compile successfully with Maven.
// These are integration tests that require Maven to be installed.
// Run with: go test -tags=integration ./internal/generator/...

// checkMavenInstalled skips the test if Maven is not available
func checkMavenInstalled(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("mvn"); err != nil {
		t.Skip("Maven not installed, skipping compilation test")
	}
}

// runMavenCompile runs 'mvn clean compile -q' in the given directory
func runMavenCompile(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("mvn", "clean", "compile", "-q", "-DskipTests")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Maven compilation failed in %s: %v", dir, err)
	}
}

// runMavenInstall runs 'mvn clean install' (including tests) in the given directory
func runMavenInstall(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("mvn", "clean", "install")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("Maven install failed in %s: %v", dir, err)
	}
}

// generateProject is a helper to generate a project with given config
func generateProject(t *testing.T, tempDir string, cfg *config.ProjectConfig) string {
	t.Helper()

	oldWd, _ := os.Getwd()
	os.Chdir(tempDir)
	defer os.Chdir(oldWd)

	gen, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create generator: %v", err)
	}

	if err := gen.Generate(); err != nil {
		t.Fatalf("Failed to generate project: %v", err)
	}

	return filepath.Join(tempDir, cfg.ProjectName)
}

func TestCompilation_ModelOnly(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "model-only",
		GroupID:     "com.test.modelonly",
		ArtifactID:  "model-only",
		JavaVersion: "21",
		Modules:     []string{"Model"},
		Database:    "",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Model-only project compiled successfully")
}

func TestCompilation_ModelAndSQLDatastore(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "model-sqldatastore",
		GroupID:     "com.test.modelsql",
		ArtifactID:  "model-sqldatastore",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Model + SQLDatastore project compiled successfully")
}

func TestCompilation_ModelAndShared(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "model-shared",
		GroupID:     "com.test.modelshared",
		ArtifactID:  "model-shared",
		JavaVersion: "21",
		Modules:     []string{"Model", "Shared"},
		Database:    "",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Model + Shared project compiled successfully")
}

func TestCompilation_ModelAndAPI(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "model-api",
		GroupID:     "com.test.modelapi",
		ArtifactID:  "model-api",
		JavaVersion: "21",
		Modules:     []string{"Model", "API"},
		Database:    "",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Model + API project compiled successfully")
}

func TestCompilation_ModelSQLDatastoreShared(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "model-sql-shared",
		GroupID:     "com.test.modelsqlshared",
		ArtifactID:  "model-sql-shared",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "Shared"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Model + SQLDatastore + Shared project compiled successfully")
}

func TestCompilation_ModelSQLDatastoreAPI(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "model-sql-api",
		GroupID:     "com.test.modelsqlapi",
		ArtifactID:  "model-sql-api",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "API"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Model + SQLDatastore + API project compiled successfully")
}

func TestCompilation_AllModules(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "all-modules",
		GroupID:     "com.test.allmodules",
		ArtifactID:  "all-modules",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "Shared", "API"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("All modules project compiled successfully")
}

func TestCompilation_MySQLDatabase(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "mysql-project",
		GroupID:     "com.test.mysql",
		ArtifactID:  "mysql-project",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore", "API"},
		Database:    "mysql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("MySQL project compiled successfully")
}

func TestCompilation_Java17(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "java17-project",
		GroupID:     "com.test.java17",
		ArtifactID:  "java17-project",
		JavaVersion: "17",
		Modules:     []string{"Model", "SQLDatastore", "Shared", "API"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Java 17 project compiled successfully")
}

func TestCompilation_ModelSQLDatastoreMySQL(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "model-sql-mysql",
		GroupID:     "com.test.mysqlonly",
		ArtifactID:  "model-sql-mysql",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore"},
		Database:    "mysql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenInstall(t, projectDir)
	t.Log("Model + SQLDatastore (MySQL) project installed successfully")
}

func TestCompilation_ModelSQLDatastorePostgres_Install(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "model-sql-postgres",
		GroupID:     "com.test.pgonly",
		ArtifactID:  "model-sql-postgres",
		JavaVersion: "21",
		Modules:     []string{"Model", "SQLDatastore"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenInstall(t, projectDir)
	t.Log("Model + SQLDatastore (PostgreSQL) project installed successfully")
}

func TestCompilation_WorkerWithSQLDatastore(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "worker-sql",
		GroupID:     "com.test.workersql",
		ArtifactID:  "worker-sql",
		JavaVersion: "21",
		Modules:     []string{"Model", "Jobs", "SQLDatastore", "Worker"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Worker + SQLDatastore project compiled successfully")
}

func TestCompilation_WorkerWithNoSQLDatastore_MongoDB(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "worker-mongodb",
		GroupID:     "com.test.workermongo",
		ArtifactID:  "worker-mongodb",
		JavaVersion: "21",
		Modules:     []string{"Model", "Jobs", "NoSQLDatastore", "Worker"},
		NoSQLDatabase: "mongodb",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Worker + NoSQLDatastore (MongoDB) project compiled successfully")
}

func TestCompilation_WorkerWithNoSQLDatastore_Redis(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "worker-redis",
		GroupID:     "com.test.workerredis",
		ArtifactID:  "worker-redis",
		JavaVersion: "21",
		Modules:     []string{"Model", "Jobs", "NoSQLDatastore", "Worker"},
		NoSQLDatabase: "redis",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Worker + NoSQLDatastore (Redis) project compiled successfully")
}

func TestCompilation_AllModulesWithWorker(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "all-with-worker",
		GroupID:     "com.test.allwithworker",
		ArtifactID:  "all-with-worker",
		JavaVersion: "21",
		Modules:     []string{"Model", "Jobs", "SQLDatastore", "Shared", "API", "Worker"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("All modules with Worker project compiled successfully")
}

func TestCompilation_WorkerWithSQLDatastore_Install(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	cfg := &config.ProjectConfig{
		ProjectName: "worker-sql-install",
		GroupID:     "com.test.workersqlinstall",
		ArtifactID:  "worker-sql-install",
		JavaVersion: "21",
		Modules:     []string{"Model", "Jobs", "SQLDatastore", "Worker"},
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenInstall(t, projectDir)
	t.Log("Worker + SQLDatastore project installed successfully (with tests)")
}

func TestCompilation_WorkerWithPostgresFallback(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Worker without any datastore - should use PostgreSQL fallback for JobRunr
	cfg := &config.ProjectConfig{
		ProjectName: "worker-fallback",
		GroupID:     "com.test.workerfallback",
		ArtifactID:  "worker-fallback",
		JavaVersion: "21",
		Modules:     []string{"Model", "Jobs", "Worker"},
		// No Database or NoSQLDatabase - Worker will use PostgreSQL fallback
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Worker with PostgreSQL fallback (no datastore) compiled successfully")
}

func TestCompilation_WorkerAutoResolvesJobs(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Worker without Jobs explicitly listed - Jobs should be auto-resolved
	modules := config.ResolveDependencies([]string{"Worker", "SQLDatastore"})

	// Verify Jobs was auto-included
	hasJobs := false
	for _, m := range modules {
		if m == "Jobs" {
			hasJobs = true
			break
		}
	}
	if !hasJobs {
		t.Fatal("Jobs module should be auto-resolved when Worker is selected")
	}

	cfg := &config.ProjectConfig{
		ProjectName: "worker-auto-jobs",
		GroupID:     "com.test.workerautojobs",
		ArtifactID:  "worker-auto-jobs",
		JavaVersion: "21",
		Modules:     modules, // Uses resolved modules (Jobs auto-included)
		Database:    "postgresql",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("Worker with auto-resolved Jobs module compiled successfully")
}

func TestCompilation_EventConsumerWithKafka(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve modules to auto-include Events
	modules := config.ResolveDependencies([]string{"Model", "API", "EventConsumer"})

	cfg := &config.ProjectConfig{
		ProjectName:   "kafka-events",
		GroupID:       "com.test.kafkaevents",
		ArtifactID:    "kafka-events",
		JavaVersion:   "21",
		Modules:       modules,
		MessageBroker: "kafka",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("EventConsumer with Kafka compiled successfully")
}

func TestCompilation_EventConsumerWithRabbitMQ(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve modules to auto-include Events
	modules := config.ResolveDependencies([]string{"Model", "API", "EventConsumer"})

	cfg := &config.ProjectConfig{
		ProjectName:   "rabbitmq-events",
		GroupID:       "com.test.rabbitmqevents",
		ArtifactID:    "rabbitmq-events",
		JavaVersion:   "21",
		Modules:       modules,
		MessageBroker: "rabbitmq",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("EventConsumer with RabbitMQ compiled successfully")
}

func TestCompilation_EventConsumerWithSQS(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve modules to auto-include Events
	modules := config.ResolveDependencies([]string{"Model", "API", "EventConsumer"})

	cfg := &config.ProjectConfig{
		ProjectName:   "sqs-events",
		GroupID:       "com.test.sqsevents",
		ArtifactID:    "sqs-events",
		JavaVersion:   "21",
		Modules:       modules,
		MessageBroker: "sqs",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("EventConsumer with AWS SQS compiled successfully")
}

func TestCompilation_EventConsumerWithPubSub(t *testing.T) {
	checkMavenInstalled(t)

	tempDir, err := os.MkdirTemp("", "trabuco-compile-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Resolve modules to auto-include Events
	modules := config.ResolveDependencies([]string{"Model", "API", "EventConsumer"})

	cfg := &config.ProjectConfig{
		ProjectName:   "pubsub-events",
		GroupID:       "com.test.pubsubevents",
		ArtifactID:    "pubsub-events",
		JavaVersion:   "21",
		Modules:       modules,
		MessageBroker: "pubsub",
	}

	projectDir := generateProject(t, tempDir, cfg)
	t.Logf("Generated project at: %s", projectDir)

	runMavenCompile(t, projectDir)
	t.Log("EventConsumer with GCP Pub/Sub compiled successfully")
}
