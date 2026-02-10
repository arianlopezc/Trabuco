package doctor

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
	"gopkg.in/yaml.v3"
)

// POMProject represents relevant parts of a Maven POM file
type POMProject struct {
	XMLName    xml.Name      `xml:"project"`
	GroupID    string        `xml:"groupId"`
	ArtifactID string        `xml:"artifactId"`
	Modules    []string      `xml:"modules>module"`
	Properties POMProperties `xml:"properties"`
}

// POMProperties holds relevant POM properties
type POMProperties struct {
	JavaSource string `xml:"maven.compiler.source"`
	JavaTarget string `xml:"maven.compiler.target"`
}

// AppConfig represents relevant parts of application.yml
type AppConfig struct {
	Spring SpringConfig `yaml:"spring"`
}

// SpringConfig holds Spring configuration
type SpringConfig struct {
	Datasource  DatasourceConfig  `yaml:"datasource"`
	Data        DataConfig        `yaml:"data"`
	Kafka       KafkaConfig       `yaml:"kafka"`
	RabbitMQ    RabbitMQConfig    `yaml:"rabbitmq"`
}

// DatasourceConfig holds database configuration
type DatasourceConfig struct {
	URL string `yaml:"url"`
}

// DataConfig holds Spring Data configuration
type DataConfig struct {
	MongoDB MongoDBConfig `yaml:"mongodb"`
	Redis   RedisConfig   `yaml:"redis"`
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI string `yaml:"uri"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host string `yaml:"host"`
}

// KafkaConfig holds Kafka configuration
type KafkaConfig struct {
	BootstrapServers string `yaml:"bootstrap-servers"`
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	Host string `yaml:"host"`
}

// DetectProject detects if a directory is a Trabuco project and returns its metadata
// It first tries to load .trabuco.json, falling back to POM inference if not found
func DetectProject(projectPath string) (*config.ProjectMetadata, error) {
	// First, check if pom.xml exists (required for any Maven project)
	pomPath := filepath.Join(projectPath, "pom.xml")
	if _, err := os.Stat(pomPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("not a Maven project (pom.xml not found)")
	}

	// Try to load existing .trabuco.json
	if config.MetadataExists(projectPath) {
		metadata, err := config.LoadMetadata(projectPath)
		if err == nil {
			return metadata, nil
		}
		// If metadata exists but is invalid, we'll try to infer from POM
	}

	// Infer from POM
	return InferFromPOM(projectPath)
}

// InferFromPOM infers project metadata from the parent POM file
func InferFromPOM(projectPath string) (*config.ProjectMetadata, error) {
	pomPath := filepath.Join(projectPath, "pom.xml")
	pom, err := ParseParentPOM(pomPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse parent POM: %w", err)
	}

	// Check if this looks like a Trabuco project
	if !isTrabucoProject(projectPath, pom) {
		return nil, fmt.Errorf("not a Trabuco project (structure doesn't match)")
	}

	metadata := &config.ProjectMetadata{
		ProjectName: extractProjectName(pom.ArtifactID),
		GroupID:     pom.GroupID,
		ArtifactID:  extractProjectName(pom.ArtifactID),
		JavaVersion: pom.Properties.JavaSource,
		Modules:     pom.Modules,
	}

	// If Java version is empty, try JavaTarget
	if metadata.JavaVersion == "" {
		metadata.JavaVersion = pom.Properties.JavaTarget
	}

	// Infer database configuration
	metadata.Database, metadata.NoSQLDatabase = inferDatabaseConfig(projectPath, pom.Modules)

	// Infer message broker configuration
	metadata.MessageBroker = inferMessageBrokerConfig(projectPath, pom.Modules)

	return metadata, nil
}

// ParseParentPOM parses the parent pom.xml file
func ParseParentPOM(pomPath string) (*POMProject, error) {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read POM file: %w", err)
	}

	var pom POMProject
	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil, fmt.Errorf("failed to parse POM XML: %w", err)
	}

	return &pom, nil
}

// ParseApplicationYAML parses an application.yml file
func ParseApplicationYAML(yamlPath string) (*AppConfig, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read YAML file: %w", err)
	}

	var appConfig AppConfig
	if err := yaml.Unmarshal(data, &appConfig); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &appConfig, nil
}

// isTrabucoProject checks if a project has the structure of a Trabuco-generated project
func isTrabucoProject(projectPath string, pom *POMProject) bool {
	// Must have at least Model module (all Trabuco projects have Model)
	hasModel := false
	for _, module := range pom.Modules {
		if module == config.ModuleModel {
			hasModel = true
			break
		}
	}

	if !hasModel {
		return false
	}

	// Check for Model module directory with expected structure
	modelPath := filepath.Join(projectPath, "Model", "src", "main", "java")
	if _, err := os.Stat(modelPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// extractProjectName removes the "-parent" suffix from artifact ID if present
func extractProjectName(artifactID string) string {
	return strings.TrimSuffix(artifactID, "-parent")
}

// inferDatabaseConfig infers database configuration from module structure
func inferDatabaseConfig(projectPath string, modules []string) (database, nosqlDatabase string) {
	for _, module := range modules {
		switch module {
		case config.ModuleSQLDatastore:
			// Try to detect database type from application.yml
			yamlPath := filepath.Join(projectPath, config.ModuleSQLDatastore, "src", "main", "resources", "application.yml")
			appConfig, err := ParseApplicationYAML(yamlPath)
			if err == nil && appConfig.Spring.Datasource.URL != "" {
				database = detectDatabaseFromURL(appConfig.Spring.Datasource.URL)
			}
			if database == "" {
				database = "postgresql" // Default
			}
		case config.ModuleNoSQLDatastore:
			// Try to detect NoSQL database type from application.yml
			yamlPath := filepath.Join(projectPath, config.ModuleNoSQLDatastore, "src", "main", "resources", "application.yml")
			appConfig, err := ParseApplicationYAML(yamlPath)
			if err == nil {
				if appConfig.Spring.Data.MongoDB.URI != "" {
					nosqlDatabase = "mongodb"
				} else if appConfig.Spring.Data.Redis.Host != "" {
					nosqlDatabase = "redis"
				}
			}
			if nosqlDatabase == "" {
				nosqlDatabase = "mongodb" // Default
			}
		}
	}
	return database, nosqlDatabase
}

// detectDatabaseFromURL detects database type from JDBC URL
func detectDatabaseFromURL(url string) string {
	url = strings.ToLower(url)
	if strings.Contains(url, "postgresql") || strings.Contains(url, "postgres") {
		return "postgresql"
	}
	if strings.Contains(url, "mysql") {
		return "mysql"
	}
	return "generic"
}

// inferMessageBrokerConfig infers message broker configuration from module structure
func inferMessageBrokerConfig(projectPath string, modules []string) string {
	for _, module := range modules {
		if module == config.ModuleEventConsumer {
			// Try to detect message broker from application.yml
			yamlPath := filepath.Join(projectPath, config.ModuleEventConsumer, "src", "main", "resources", "application.yml")
			appConfig, err := ParseApplicationYAML(yamlPath)
			if err == nil {
				if appConfig.Spring.Kafka.BootstrapServers != "" {
					return "kafka"
				}
				if appConfig.Spring.RabbitMQ.Host != "" {
					return "rabbitmq"
				}
			}

			// Check for config files that indicate broker type
			configPath := filepath.Join(projectPath, config.ModuleEventConsumer, "src", "main", "java")
			if containsFile(configPath, "KafkaConfig.java") {
				return "kafka"
			}
			if containsFile(configPath, "RabbitConfig.java") {
				return "rabbitmq"
			}
			if containsFile(configPath, "SqsConfig.java") {
				return "sqs"
			}
			if containsFile(configPath, "PubSubConfig.java") {
				return "pubsub"
			}

			return "kafka" // Default
		}
	}
	return ""
}

// containsFile checks if a directory (recursively) contains a file with the given name
func containsFile(dir, filename string) bool {
	found := false
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() && info.Name() == filename {
			found = true
			return filepath.SkipAll
		}
		return nil
	})
	return found
}

// GetModulesFromPOM extracts module names from a parent POM
func GetModulesFromPOM(projectPath string) ([]string, error) {
	pomPath := filepath.Join(projectPath, "pom.xml")
	pom, err := ParseParentPOM(pomPath)
	if err != nil {
		return nil, err
	}
	return pom.Modules, nil
}

// GetJavaVersionFromPOM extracts Java version from a POM
func GetJavaVersionFromPOM(projectPath string) (string, error) {
	pomPath := filepath.Join(projectPath, "pom.xml")
	pom, err := ParseParentPOM(pomPath)
	if err != nil {
		return "", err
	}
	if pom.Properties.JavaSource != "" {
		return pom.Properties.JavaSource, nil
	}
	return pom.Properties.JavaTarget, nil
}

// GetGroupIDFromPOM extracts group ID from a POM
func GetGroupIDFromPOM(projectPath string) (string, error) {
	pomPath := filepath.Join(projectPath, "pom.xml")
	pom, err := ParseParentPOM(pomPath)
	if err != nil {
		return "", err
	}
	return pom.GroupID, nil
}

// IsValidPOM checks if a POM file is valid XML
func IsValidPOM(pomPath string) bool {
	_, err := ParseParentPOM(pomPath)
	return err == nil
}

// HasRequiredPOMSections checks if a parent POM has the required sections
func HasRequiredPOMSections(projectPath string) (hasModules, hasProperties bool, err error) {
	pomPath := filepath.Join(projectPath, "pom.xml")
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return false, false, err
	}

	content := string(data)
	hasModules = strings.Contains(content, "<modules>")
	hasProperties = strings.Contains(content, "<properties>")
	return hasModules, hasProperties, nil
}

// ParseModulePOM parses a module's pom.xml and returns relevant info
type ModulePOMInfo struct {
	GroupID    string
	ArtifactID string
	Parent     struct {
		GroupID    string
		ArtifactID string
	}
}

// ParseModulePOM parses a module's pom.xml
func ParseModulePOM(pomPath string) (*ModulePOMInfo, error) {
	type ModulePOM struct {
		XMLName    xml.Name `xml:"project"`
		GroupID    string   `xml:"groupId"`
		ArtifactID string   `xml:"artifactId"`
		Parent     struct {
			GroupID    string `xml:"groupId"`
			ArtifactID string `xml:"artifactId"`
		} `xml:"parent"`
	}

	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, err
	}

	var pom ModulePOM
	if err := xml.Unmarshal(data, &pom); err != nil {
		return nil, err
	}

	info := &ModulePOMInfo{
		GroupID:    pom.GroupID,
		ArtifactID: pom.ArtifactID,
	}
	info.Parent.GroupID = pom.Parent.GroupID
	info.Parent.ArtifactID = pom.Parent.ArtifactID

	// If module doesn't have its own groupId, inherit from parent
	if info.GroupID == "" {
		info.GroupID = info.Parent.GroupID
	}

	return info, nil
}

// DockerComposeService represents a service in docker-compose.yml
type DockerComposeService struct {
	Image string `yaml:"image"`
}

// DockerCompose represents a docker-compose.yml file
type DockerCompose struct {
	Services map[string]DockerComposeService `yaml:"services"`
}

// ParseDockerCompose parses a docker-compose.yml file
func ParseDockerCompose(composePath string) (*DockerCompose, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		return nil, err
	}

	var dc DockerCompose
	if err := yaml.Unmarshal(data, &dc); err != nil {
		return nil, err
	}

	return &dc, nil
}

// GetRequiredDockerServices returns the services required based on metadata
func GetRequiredDockerServices(meta *config.ProjectMetadata) []string {
	var required []string

	// SQL database service
	if meta.HasModule(config.ModuleSQLDatastore) {
		switch meta.Database {
		case config.DatabasePostgreSQL:
			required = append(required, "postgres")
		case config.DatabaseMySQL:
			required = append(required, "mysql")
		}
	}

	// NoSQL database service
	if meta.HasModule(config.ModuleNoSQLDatastore) {
		switch meta.NoSQLDatabase {
		case config.DatabaseMongoDB:
			required = append(required, "mongodb")
		case config.DatabaseRedis:
			required = append(required, "redis")
		}
	}

	// Message broker service
	if meta.HasModule(config.ModuleEventConsumer) {
		switch meta.MessageBroker {
		case config.BrokerKafka:
			required = append(required, "kafka")
		case config.BrokerRabbitMQ:
			required = append(required, "rabbitmq")
		case config.BrokerSQS:
			required = append(required, "localstack")
		case config.BrokerPubSub:
			required = append(required, "pubsub-emulator")
		}
	}

	// Worker with PostgreSQL fallback
	if meta.HasModule(config.ModuleWorker) {
		cfg := meta.ToProjectConfig()
		if cfg.WorkerNeedsOwnPostgres() {
			required = append(required, "postgres-jobrunr")
		}
	}

	return required
}

// ExtractVersionFromPOMProperty extracts a version from POM using regex
func ExtractVersionFromPOMProperty(pomPath, propertyName string) (string, error) {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return "", err
	}

	pattern := fmt.Sprintf(`<%s>([^<]+)</%s>`, regexp.QuoteMeta(propertyName), regexp.QuoteMeta(propertyName))
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return "", fmt.Errorf("property %s not found", propertyName)
	}

	return matches[1], nil
}
