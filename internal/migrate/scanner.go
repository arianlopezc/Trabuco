package migrate

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// ProjectInfo contains information about the scanned source project
type ProjectInfo struct {
	// Name of the project
	Name string

	// ProjectType (e.g., "Spring Boot 2.7.x (Maven)")
	ProjectType string

	// JavaVersion detected from pom.xml
	JavaVersion string

	// BasePackage is the root package (e.g., "com.example.app")
	BasePackage string

	// GroupID from pom.xml
	GroupID string

	// ArtifactID from pom.xml
	ArtifactID string

	// SpringBootVersion detected
	SpringBootVersion string

	// IsMultiModule indicates if this is a multi-module Maven project
	IsMultiModule bool

	// Modules contains the list of submodule names (for multi-module projects)
	Modules []string

	// Dependencies from pom.xml (aggregated from all modules)
	Dependencies []Dependency

	// Entities are classes annotated with @Entity, @Document, @Table
	Entities []*JavaClass

	// Repositories are interfaces extending Repository, CrudRepository, etc.
	Repositories []*JavaClass

	// Services are classes annotated with @Service
	Services []*JavaClass

	// Controllers are classes annotated with @RestController, @Controller
	Controllers []*JavaClass

	// ScheduledJobs are classes with @Scheduled methods
	ScheduledJobs []*JavaClass

	// EventListeners are classes with @KafkaListener, @RabbitListener, etc.
	EventListeners []*JavaClass

	// ConfigClasses are classes annotated with @Configuration
	ConfigClasses []*JavaClass

	// HasScheduledJobs indicates if scheduled jobs were found
	HasScheduledJobs bool

	// HasEventListeners indicates if event listeners were found
	HasEventListeners bool

	// UsesNoSQL indicates if MongoDB is primary database
	UsesNoSQL bool

	// Database is the detected database type (postgresql, mysql, mongodb, etc.)
	Database string

	// UsesRabbitMQ indicates if RabbitMQ is the message broker
	UsesRabbitMQ bool

	// UsesRedis indicates if Redis is used for caching
	UsesRedis bool

	// MessageBroker detected (kafka, rabbitmq, etc.)
	MessageBroker string
}

// Dependency represents a Maven dependency
type Dependency struct {
	GroupID    string
	ArtifactID string
	Version    string
	Scope      string
}

// JavaClass represents a scanned Java class
type JavaClass struct {
	// Name of the class (e.g., "UserEntity")
	Name string

	// Package (e.g., "com.example.app.entity")
	Package string

	// FilePath is the absolute path to the source file
	FilePath string

	// Content is the raw source code
	Content string

	// Annotations found on the class
	Annotations []string

	// Imports in the file
	Imports []string

	// Methods found in the class
	Methods []JavaMethod

	// Fields found in the class
	Fields []JavaField

	// Implements interfaces
	Implements []string

	// Extends class
	Extends string
}

// JavaMethod represents a method in a Java class
type JavaMethod struct {
	Name        string
	ReturnType  string
	Parameters  []string
	Annotations []string
}

// JavaField represents a field in a Java class
type JavaField struct {
	Name        string
	Type        string
	Annotations []string
}

// ProjectScanner scans a Java project
type ProjectScanner struct {
	sourcePath string
}

// NewProjectScanner creates a new scanner
func NewProjectScanner(sourcePath string) *ProjectScanner {
	return &ProjectScanner{
		sourcePath: sourcePath,
	}
}

// Scan performs the project scan
func (s *ProjectScanner) Scan() (*ProjectInfo, error) {
	info := &ProjectInfo{}

	// Parse pom.xml
	if err := s.parsePOM(info); err != nil {
		return nil, fmt.Errorf("failed to parse pom.xml: %w", err)
	}

	// Detect project type
	info.ProjectType = s.detectProjectType(info)

	// Find Java source files (now aware of multi-module projects)
	javaFiles, err := s.findJavaFiles(info)
	if err != nil {
		return nil, fmt.Errorf("failed to scan Java files: %w", err)
	}

	// Categorize classes
	for _, filePath := range javaFiles {
		class, err := s.parseJavaFile(filePath)
		if err != nil {
			continue // Skip files we can't parse
		}

		s.categorizeClass(info, class)
	}

	// Detect base package
	info.BasePackage = s.detectBasePackage(info)

	// Detect message broker
	info.MessageBroker = s.detectMessageBroker(info)

	// Detect database type
	info.Database = s.detectDatabase(info)
	info.UsesNoSQL = info.Database == "mongodb"

	// Set flags
	info.HasScheduledJobs = len(info.ScheduledJobs) > 0
	info.HasEventListeners = len(info.EventListeners) > 0
	info.UsesRabbitMQ = info.MessageBroker == "rabbitmq"
	info.UsesRedis = s.detectRedis(info)

	return info, nil
}

// POM XML structures
type pomXML struct {
	XMLName    xml.Name `xml:"project"`
	GroupID    string   `xml:"groupId"`
	ArtifactID string   `xml:"artifactId"`
	Version    string   `xml:"version"`
	Parent     struct {
		GroupID    string `xml:"groupId"`
		ArtifactID string `xml:"artifactId"`
		Version    string `xml:"version"`
	} `xml:"parent"`
	Properties struct {
		JavaVersion string `xml:"java.version"`
	} `xml:"properties"`
	Modules struct {
		Module []string `xml:"module"`
	} `xml:"modules"`
	Dependencies struct {
		Dependency []struct {
			GroupID    string `xml:"groupId"`
			ArtifactID string `xml:"artifactId"`
			Version    string `xml:"version"`
			Scope      string `xml:"scope"`
		} `xml:"dependency"`
	} `xml:"dependencies"`
}

func (s *ProjectScanner) parsePOM(info *ProjectInfo) error {
	pomPath := filepath.Join(s.sourcePath, "pom.xml")

	data, err := os.ReadFile(pomPath)
	if err != nil {
		return fmt.Errorf("pom.xml not found: %w", err)
	}

	var pom pomXML
	if err := xml.Unmarshal(data, &pom); err != nil {
		return fmt.Errorf("failed to parse pom.xml: %w", err)
	}

	// Set project info
	info.GroupID = pom.GroupID
	if info.GroupID == "" {
		info.GroupID = pom.Parent.GroupID
	}

	info.ArtifactID = pom.ArtifactID
	info.Name = pom.ArtifactID

	info.JavaVersion = pom.Properties.JavaVersion
	if info.JavaVersion == "" {
		info.JavaVersion = "17" // Default
	}

	// Extract Spring Boot version from parent
	if strings.Contains(pom.Parent.ArtifactID, "spring-boot") {
		info.SpringBootVersion = pom.Parent.Version
	}

	// Check for multi-module project
	if len(pom.Modules.Module) > 0 {
		info.IsMultiModule = true
		info.Modules = pom.Modules.Module
	}

	// Extract dependencies from root POM
	for _, dep := range pom.Dependencies.Dependency {
		info.Dependencies = append(info.Dependencies, Dependency{
			GroupID:    dep.GroupID,
			ArtifactID: dep.ArtifactID,
			Version:    dep.Version,
			Scope:      dep.Scope,
		})
	}

	// For multi-module projects, also parse child module POMs for dependencies
	if info.IsMultiModule {
		for _, module := range info.Modules {
			childPomPath := filepath.Join(s.sourcePath, module, "pom.xml")
			childData, err := os.ReadFile(childPomPath)
			if err != nil {
				continue // Skip modules without pom.xml
			}

			var childPom pomXML
			if err := xml.Unmarshal(childData, &childPom); err != nil {
				continue // Skip malformed POMs
			}

			// Extract Spring Boot version from child if not found in parent
			if info.SpringBootVersion == "" && strings.Contains(childPom.Parent.ArtifactID, "spring-boot") {
				info.SpringBootVersion = childPom.Parent.Version
			}

			// Add child dependencies (avoiding duplicates)
			for _, dep := range childPom.Dependencies.Dependency {
				if !s.hasDependency(info.Dependencies, dep.GroupID, dep.ArtifactID) {
					info.Dependencies = append(info.Dependencies, Dependency{
						GroupID:    dep.GroupID,
						ArtifactID: dep.ArtifactID,
						Version:    dep.Version,
						Scope:      dep.Scope,
					})
				}
			}
		}
	}

	return nil
}

// hasDependency checks if a dependency already exists in the list
func (s *ProjectScanner) hasDependency(deps []Dependency, groupID, artifactID string) bool {
	for _, d := range deps {
		if d.GroupID == groupID && d.ArtifactID == artifactID {
			return true
		}
	}
	return false
}

func (s *ProjectScanner) detectProjectType(info *ProjectInfo) string {
	projectType := "Java Maven"

	// Check for Spring Boot
	for _, dep := range info.Dependencies {
		if strings.Contains(dep.GroupID, "springframework") {
			projectType = "Spring"
			break
		}
	}

	if info.SpringBootVersion != "" {
		projectType = fmt.Sprintf("Spring Boot %s (Maven)", info.SpringBootVersion)
	}

	return projectType
}

func (s *ProjectScanner) findJavaFiles(info *ProjectInfo) ([]string, error) {
	var files []string

	// Collect all source directories to scan
	var srcDirs []string

	// Always check root src/main/java
	rootSrcPath := filepath.Join(s.sourcePath, "src", "main", "java")
	if _, err := os.Stat(rootSrcPath); err == nil {
		srcDirs = append(srcDirs, rootSrcPath)
	}

	// For multi-module projects, also scan each module's src/main/java
	if info.IsMultiModule {
		for _, module := range info.Modules {
			moduleSrcPath := filepath.Join(s.sourcePath, module, "src", "main", "java")
			if _, err := os.Stat(moduleSrcPath); err == nil {
				srcDirs = append(srcDirs, moduleSrcPath)
			}
		}
	}

	// If no src directories found, try to find any Java files recursively
	// (handles non-standard project structures)
	if len(srcDirs) == 0 {
		srcDirs = append(srcDirs, s.sourcePath)
	}

	// Scan all source directories
	for _, srcPath := range srcDirs {
		err := filepath.Walk(srcPath, func(path string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}

			// Skip common non-source directories
			if fileInfo.IsDir() {
				dirName := fileInfo.Name()
				if dirName == "target" || dirName == "build" || dirName == ".git" ||
					dirName == ".idea" || dirName == "node_modules" || dirName == "test" {
					return filepath.SkipDir
				}
			}

			if !fileInfo.IsDir() && strings.HasSuffix(path, ".java") {
				// Skip test files unless scanning root (for non-standard structures)
				if !strings.Contains(path, "/test/") {
					files = append(files, path)
				}
			}

			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

func (s *ProjectScanner) parseJavaFile(filePath string) (*JavaClass, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	code := string(content)

	class := &JavaClass{
		FilePath: filePath,
		Content:  code,
	}

	// Extract package
	packageRegex := regexp.MustCompile(`package\s+([\w.]+);`)
	if match := packageRegex.FindStringSubmatch(code); len(match) > 1 {
		class.Package = match[1]
	}

	// Extract class name
	classRegex := regexp.MustCompile(`(?:public\s+)?(?:abstract\s+)?(?:class|interface|enum)\s+(\w+)`)
	if match := classRegex.FindStringSubmatch(code); len(match) > 1 {
		class.Name = match[1]
	}

	// Extract imports
	importRegex := regexp.MustCompile(`import\s+([\w.]+);`)
	for _, match := range importRegex.FindAllStringSubmatch(code, -1) {
		if len(match) > 1 {
			class.Imports = append(class.Imports, match[1])
		}
	}

	// Extract class-level annotations
	annotationRegex := regexp.MustCompile(`@(\w+)(?:\([^)]*\))?`)
	for _, match := range annotationRegex.FindAllStringSubmatch(code, -1) {
		if len(match) > 1 {
			class.Annotations = append(class.Annotations, match[1])
		}
	}

	// Extract extends
	extendsRegex := regexp.MustCompile(`(?:class|interface)\s+\w+\s+extends\s+(\w+)`)
	if match := extendsRegex.FindStringSubmatch(code); len(match) > 1 {
		class.Extends = match[1]
	}

	// Extract implements
	implementsRegex := regexp.MustCompile(`(?:class)\s+\w+(?:\s+extends\s+\w+)?\s+implements\s+([\w,\s<>]+)`)
	if match := implementsRegex.FindStringSubmatch(code); len(match) > 1 {
		interfaces := strings.Split(match[1], ",")
		for _, iface := range interfaces {
			iface = strings.TrimSpace(iface)
			// Remove generics
			if idx := strings.Index(iface, "<"); idx != -1 {
				iface = iface[:idx]
			}
			class.Implements = append(class.Implements, iface)
		}
	}

	return class, nil
}

func (s *ProjectScanner) categorizeClass(info *ProjectInfo, class *JavaClass) {
	annotations := strings.Join(class.Annotations, " ")

	// Check for entities
	if strings.Contains(annotations, "Entity") ||
		strings.Contains(annotations, "Document") ||
		strings.Contains(annotations, "Table") {
		info.Entities = append(info.Entities, class)
		return
	}

	// Check for repositories
	for _, impl := range class.Implements {
		if strings.Contains(impl, "Repository") ||
			strings.Contains(impl, "CrudRepository") ||
			strings.Contains(impl, "JpaRepository") ||
			strings.Contains(impl, "MongoRepository") {
			info.Repositories = append(info.Repositories, class)
			return
		}
	}
	if strings.Contains(class.Extends, "Repository") {
		info.Repositories = append(info.Repositories, class)
		return
	}

	// Check for controllers
	if strings.Contains(annotations, "RestController") ||
		strings.Contains(annotations, "Controller") {
		info.Controllers = append(info.Controllers, class)
		return
	}

	// Check for services
	if strings.Contains(annotations, "Service") {
		info.Services = append(info.Services, class)
		return
	}

	// Check for scheduled jobs
	if strings.Contains(class.Content, "@Scheduled") {
		info.ScheduledJobs = append(info.ScheduledJobs, class)
		return
	}

	// Check for event listeners
	if strings.Contains(class.Content, "@KafkaListener") ||
		strings.Contains(class.Content, "@RabbitListener") ||
		strings.Contains(class.Content, "@SqsListener") ||
		strings.Contains(class.Content, "@JmsListener") {
		info.EventListeners = append(info.EventListeners, class)
		return
	}

	// Check for configuration
	if strings.Contains(annotations, "Configuration") {
		info.ConfigClasses = append(info.ConfigClasses, class)
	}
}

func (s *ProjectScanner) detectBasePackage(info *ProjectInfo) string {
	// Start with GroupID as base
	basePackage := info.GroupID

	// Find the most common root package among all classes
	packageCounts := make(map[string]int)

	allClasses := append(info.Entities, info.Repositories...)
	allClasses = append(allClasses, info.Services...)
	allClasses = append(allClasses, info.Controllers...)

	for _, class := range allClasses {
		parts := strings.Split(class.Package, ".")
		for i := 1; i <= len(parts); i++ {
			prefix := strings.Join(parts[:i], ".")
			packageCounts[prefix]++
		}
	}

	// Find the longest common prefix
	maxCount := 0
	for pkg, count := range packageCounts {
		if count >= len(allClasses)/2 && len(pkg) > len(basePackage) {
			if count > maxCount {
				basePackage = pkg
				maxCount = count
			}
		}
	}

	return basePackage
}

func (s *ProjectScanner) detectMessageBroker(info *ProjectInfo) string {
	for _, dep := range info.Dependencies {
		if strings.Contains(dep.ArtifactID, "kafka") {
			return "kafka"
		}
		if strings.Contains(dep.ArtifactID, "rabbitmq") || strings.Contains(dep.ArtifactID, "amqp") {
			return "rabbitmq"
		}
		if strings.Contains(dep.ArtifactID, "sqs") {
			return "sqs"
		}
		if strings.Contains(dep.ArtifactID, "pubsub") {
			return "pubsub"
		}
	}
	return ""
}

func (s *ProjectScanner) detectDatabase(info *ProjectInfo) string {
	// Check dependencies for database drivers/starters
	for _, dep := range info.Dependencies {
		artifactLower := strings.ToLower(dep.ArtifactID)
		groupLower := strings.ToLower(dep.GroupID)

		// PostgreSQL
		if strings.Contains(artifactLower, "postgresql") ||
			strings.Contains(artifactLower, "postgres") ||
			strings.Contains(groupLower, "postgresql") {
			return "postgresql"
		}

		// MySQL
		if strings.Contains(artifactLower, "mysql") {
			return "mysql"
		}

		// MariaDB
		if strings.Contains(artifactLower, "mariadb") {
			return "mysql" // Compatible with MySQL
		}

		// MongoDB
		if strings.Contains(artifactLower, "mongodb") ||
			strings.Contains(artifactLower, "mongo") ||
			strings.Contains(groupLower, "mongodb") {
			return "mongodb"
		}

		// Oracle
		if strings.Contains(artifactLower, "oracle") ||
			strings.Contains(artifactLower, "ojdbc") {
			return "oracle"
		}

		// SQL Server
		if strings.Contains(artifactLower, "sqlserver") ||
			strings.Contains(artifactLower, "mssql") {
			return "sqlserver"
		}
	}

	// Check for H2 (usually test/dev)
	for _, dep := range info.Dependencies {
		if strings.Contains(strings.ToLower(dep.ArtifactID), "h2") {
			// H2 is often used with PostgreSQL dialect, default to postgresql
			return "postgresql"
		}
	}

	// Check for generic JPA/JDBC - default to postgresql
	for _, dep := range info.Dependencies {
		artifactLower := strings.ToLower(dep.ArtifactID)
		if strings.Contains(artifactLower, "spring-data-jpa") ||
			strings.Contains(artifactLower, "hibernate") ||
			strings.Contains(artifactLower, "jdbc") {
			return "postgresql" // Default to PostgreSQL for generic JPA
		}
	}

	// Check entity annotations for hints
	for _, entity := range info.Entities {
		for _, imp := range entity.Imports {
			if strings.Contains(imp, "mongodb") {
				return "mongodb"
			}
			if strings.Contains(imp, "javax.persistence") ||
				strings.Contains(imp, "jakarta.persistence") {
				return "postgresql" // JPA annotations suggest SQL database
			}
		}
	}

	// Default to postgresql
	return "postgresql"
}

func (s *ProjectScanner) detectNoSQL(info *ProjectInfo) bool {
	// First, check if we have explicit SQL database dependencies
	// If we have PostgreSQL, MySQL, H2, etc., we should NOT use NoSQL
	hasSQLDatabase := false
	hasMongoDBDependency := false

	for _, dep := range info.Dependencies {
		artifactLower := strings.ToLower(dep.ArtifactID)
		groupLower := strings.ToLower(dep.GroupID)

		// Check for SQL databases
		if strings.Contains(artifactLower, "postgresql") ||
			strings.Contains(artifactLower, "postgres") ||
			strings.Contains(artifactLower, "mysql") ||
			strings.Contains(artifactLower, "mariadb") ||
			strings.Contains(artifactLower, "h2") ||
			strings.Contains(artifactLower, "oracle") ||
			strings.Contains(artifactLower, "sqlserver") ||
			strings.Contains(artifactLower, "jdbc") ||
			strings.Contains(artifactLower, "spring-data-jpa") ||
			strings.Contains(artifactLower, "hibernate") {
			hasSQLDatabase = true
		}

		// Check for MongoDB
		if strings.Contains(artifactLower, "mongodb") ||
			strings.Contains(artifactLower, "mongo") ||
			strings.Contains(groupLower, "mongodb") {
			hasMongoDBDependency = true
		}
	}

	// If we have SQL dependencies, prefer SQL over NoSQL
	if hasSQLDatabase {
		return false
	}

	// If we have explicit MongoDB dependency, use NoSQL
	if hasMongoDBDependency {
		return true
	}

	// Check entity annotations for MongoDB @Document
	// Be careful: only match org.springframework.data.mongodb.core.mapping.Document
	for _, entity := range info.Entities {
		for _, imp := range entity.Imports {
			if strings.Contains(imp, "mongodb") && strings.Contains(imp, "Document") {
				return true
			}
		}
	}

	return false
}

func (s *ProjectScanner) detectRedis(info *ProjectInfo) bool {
	for _, dep := range info.Dependencies {
		if strings.Contains(dep.ArtifactID, "redis") {
			return true
		}
	}
	return false
}
