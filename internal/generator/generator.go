package generator

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/templates"
)

// Generator handles project generation
type Generator struct {
	config *config.ProjectConfig
	engine *templates.Engine
	outDir string
}

// New creates a new Generator
func New(cfg *config.ProjectConfig) (*Generator, error) {
	engine := templates.NewEngine()

	return &Generator{
		config: cfg,
		engine: engine,
		outDir: cfg.ProjectName,
	}, nil
}

// Generate creates the complete project structure
func (g *Generator) Generate() error {
	green := color.New(color.FgGreen)
	yellow := color.New(color.FgYellow)

	yellow.Println("\nGenerating project...")

	// Check if directory already exists
	if _, err := os.Stat(g.outDir); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", g.outDir)
	}

	// Create directory structure
	if err := g.createDirectories(); err != nil {
		g.cleanup()
		return fmt.Errorf("failed to create directories: %w", err)
	}
	green.Println("  ✓ Created directory structure")

	// Generate parent POM
	if err := g.generateParentPOM(); err != nil {
		g.cleanup()
		return fmt.Errorf("failed to generate parent pom.xml: %w", err)
	}
	green.Println("  ✓ Created parent pom.xml")

	// Generate modules
	for _, module := range g.config.Modules {
		if err := g.generateModule(module); err != nil {
			g.cleanup()
			return fmt.Errorf("failed to generate %s module: %w", module, err)
		}
		green.Printf("  ✓ Created %s module\n", module)
	}

	// Generate documentation files
	if err := g.generateDocs(); err != nil {
		g.cleanup()
		return fmt.Errorf("failed to generate documentation: %w", err)
	}
	green.Println("  ✓ Created documentation files")

	return nil
}

// createDirectories creates all necessary directories for the project
func (g *Generator) createDirectories() error {
	packagePath := g.config.PackagePath()

	dirs := []string{
		// Root
		g.outDir,
	}

	// Model module directories (always required)
	if g.config.HasModule("Model") {
		modelBase := filepath.Join(g.outDir, "Model", "src", "main", "java", packagePath, "model")
		dirs = append(dirs,
			modelBase,
			filepath.Join(modelBase, "entities"),
			filepath.Join(modelBase, "dto"),
			filepath.Join(modelBase, "enums"),
			filepath.Join(modelBase, "exception"),
			filepath.Join(modelBase, "util"),
			filepath.Join(modelBase, "events"),
			filepath.Join(modelBase, "jobs"),
			filepath.Join(modelBase, "validation"),
		)
	}

	// SQLDatastore module directories
	if g.config.HasModule("SQLDatastore") {
		sqlBase := filepath.Join(g.outDir, "SQLDatastore", "src", "main", "java", packagePath, "sqldatastore")
		sqlTestBase := filepath.Join(g.outDir, "SQLDatastore", "src", "test", "java", packagePath, "sqldatastore")
		dirs = append(dirs,
			filepath.Join(sqlBase, "config"),
			filepath.Join(sqlBase, "repository"),
			filepath.Join(g.outDir, "SQLDatastore", "src", "main", "resources", "db", "migration"),
			filepath.Join(sqlTestBase, "repository"),
		)
	}

	// NoSQLDatastore module directories
	if g.config.HasModule("NoSQLDatastore") {
		nosqlBase := filepath.Join(g.outDir, "NoSQLDatastore", "src", "main", "java", packagePath, "nosqldatastore")
		nosqlTestBase := filepath.Join(g.outDir, "NoSQLDatastore", "src", "test", "java", packagePath, "nosqldatastore")
		dirs = append(dirs,
			filepath.Join(nosqlBase, "config"),
			filepath.Join(nosqlBase, "repository"),
			filepath.Join(g.outDir, "NoSQLDatastore", "src", "main", "resources"),
			filepath.Join(nosqlTestBase, "repository"),
		)
	}

	// Shared module directories
	if g.config.HasModule("Shared") {
		sharedBase := filepath.Join(g.outDir, "Shared", "src", "main", "java", packagePath, "shared")
		sharedTestBase := filepath.Join(g.outDir, "Shared", "src", "test", "java", packagePath, "shared")
		dirs = append(dirs,
			filepath.Join(sharedBase, "config"),
			filepath.Join(sharedBase, "service"),
			filepath.Join(g.outDir, "Shared", "src", "main", "resources"),
			filepath.Join(sharedTestBase, "service"),
		)
	}

	// API module directories
	if g.config.HasModule("API") {
		apiBase := filepath.Join(g.outDir, "API", "src", "main", "java", packagePath, "api")
		dirs = append(dirs,
			apiBase,
			filepath.Join(apiBase, "controller"),
			filepath.Join(apiBase, "config"),
			filepath.Join(g.outDir, "API", "src", "main", "resources"),
			filepath.Join(g.outDir, ".run"), // IntelliJ run configurations
		)
	}

	// Jobs module directories (auto-included with Worker)
	// NOTE: Job request schemas are in Model module; Jobs module contains job services
	if g.config.HasModule("Jobs") {
		jobsBase := filepath.Join(g.outDir, "Jobs", "src", "main", "java", packagePath, "jobs")
		dirs = append(dirs,
			jobsBase,
		)
	}

	// Worker module directories
	if g.config.HasModule("Worker") {
		workerBase := filepath.Join(g.outDir, "Worker", "src", "main", "java", packagePath, "worker")
		workerTestBase := filepath.Join(g.outDir, "Worker", "src", "test", "java", packagePath, "worker")
		dirs = append(dirs,
			workerBase,
			filepath.Join(workerBase, "config"),
			filepath.Join(workerBase, "handler"),
			filepath.Join(g.outDir, "Worker", "src", "main", "resources"),
			filepath.Join(workerTestBase, "handler"),
			filepath.Join(g.outDir, ".run"), // IntelliJ run configurations (if not already created by API)
		)
	}

	// Events module directories (auto-included with EventConsumer)
	// NOTE: Event schemas are now in Model module; Events module contains EventPublisher service
	if g.config.HasModule("Events") {
		eventsBase := filepath.Join(g.outDir, "Events", "src", "main", "java", packagePath, "events")
		dirs = append(dirs,
			eventsBase,
			filepath.Join(eventsBase, "config"),
		)
	}

	// EventConsumer module directories
	if g.config.HasModule("EventConsumer") {
		eventConsumerBase := filepath.Join(g.outDir, "EventConsumer", "src", "main", "java", packagePath, "eventconsumer")
		eventConsumerTestBase := filepath.Join(g.outDir, "EventConsumer", "src", "test", "java", packagePath, "eventconsumer")
		dirs = append(dirs,
			eventConsumerBase,
			filepath.Join(eventConsumerBase, "config"),
			filepath.Join(eventConsumerBase, "listener"),
			filepath.Join(g.outDir, "EventConsumer", "src", "main", "resources"),
			filepath.Join(eventConsumerTestBase, "listener"),
		)
	}

	// Create all directories
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// generateModule generates all files for a specific module
func (g *Generator) generateModule(module string) error {
	switch module {
	case "Model":
		return g.generateModelModule()
	case "Jobs":
		return g.generateJobsModule()
	case "SQLDatastore":
		return g.generateSQLDatastoreModule()
	case "NoSQLDatastore":
		return g.generateNoSQLDatastoreModule()
	case "Shared":
		return g.generateSharedModule()
	case "API":
		return g.generateAPIModule()
	case "Worker":
		return g.generateWorkerModule()
	case "Events":
		return g.generateEventsModule()
	case "EventConsumer":
		return g.generateEventConsumerModule()
	default:
		return fmt.Errorf("unknown module: %s", module)
	}
}

// cleanup removes the generated directory on error
func (g *Generator) cleanup() {
	if g.outDir != "" {
		if err := os.RemoveAll(g.outDir); err != nil {
			log.Printf("Warning: failed to cleanup directory %s: %v", g.outDir, err)
		}
	}
}

// writeFile writes content to a file, creating parent directories if needed
func (g *Generator) writeFile(path string, content string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// renderTemplate renders a template with the project config
func (g *Generator) renderTemplate(templatePath string) (string, error) {
	return g.engine.Execute(templatePath, g.config)
}

// writeTemplate renders a template and writes it to a file
func (g *Generator) writeTemplate(templatePath, outputPath string) error {
	content, err := g.renderTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", templatePath, err)
	}

	fullPath := filepath.Join(g.outDir, outputPath)
	return g.writeFile(fullPath, content)
}

// writeTemplateExecutable renders a template and writes it as an executable file
func (g *Generator) writeTemplateExecutable(templatePath, outputPath string) error {
	content, err := g.renderTemplate(templatePath)
	if err != nil {
		return fmt.Errorf("failed to render template %s: %w", templatePath, err)
	}

	fullPath := filepath.Join(g.outDir, outputPath)
	return g.writeFileExecutable(fullPath, content)
}

// writeFileExecutable writes content to a file with executable permissions (0755)
func (g *Generator) writeFileExecutable(path string, content string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Write file with executable permissions
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

// javaPath returns the Java source path for a module
func (g *Generator) javaPath(module, subpackage string) string {
	packagePath := g.config.PackagePath()
	moduleLower := strings.ToLower(module)
	return filepath.Join(module, "src", "main", "java", packagePath, moduleLower, subpackage)
}

// testJavaPath returns the Java test source path for a module
func (g *Generator) testJavaPath(module, subpackage string) string {
	packagePath := g.config.PackagePath()
	moduleLower := strings.ToLower(module)
	return filepath.Join(module, "src", "test", "java", packagePath, moduleLower, subpackage)
}

// resourcePath returns the resources path for a module
func (g *Generator) resourcePath(module, subpath string) string {
	return filepath.Join(module, "src", "main", "resources", subpath)
}
