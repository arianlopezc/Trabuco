package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// POMUpdater handles modifications to pom.xml files
type POMUpdater struct {
	path    string
	content string
}

// NewPOMUpdater creates a new POMUpdater
func NewPOMUpdater(pomPath string) (*POMUpdater, error) {
	data, err := os.ReadFile(pomPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read POM file: %w", err)
	}

	return &POMUpdater{
		path:    pomPath,
		content: string(data),
	}, nil
}

// Save writes the modified POM back to disk
func (p *POMUpdater) Save() error {
	return os.WriteFile(p.path, []byte(p.content), 0644)
}

// AddModule adds a module to the <modules> section
func (p *POMUpdater) AddModule(moduleName string) error {
	// Find the modules section
	modulesRegex := regexp.MustCompile(`(<modules>)([\s\S]*?)(</modules>)`)
	matches := modulesRegex.FindStringSubmatch(p.content)

	if len(matches) < 4 {
		return fmt.Errorf("could not find <modules> section in POM")
	}

	modulesContent := matches[2]

	// Check if module already exists
	if strings.Contains(modulesContent, fmt.Sprintf("<module>%s</module>", moduleName)) {
		return nil // Already exists
	}

	// Find the last module entry to determine indentation
	moduleEntryRegex := regexp.MustCompile(`(\s*)<module>[^<]+</module>`)
	moduleMatches := moduleEntryRegex.FindAllStringSubmatch(modulesContent, -1)

	indent := "        " // Default indent
	if len(moduleMatches) > 0 {
		lastMatch := moduleMatches[len(moduleMatches)-1]
		indent = lastMatch[1]
	}

	// Add new module entry before </modules>
	newModuleEntry := fmt.Sprintf("%s<module>%s</module>\n", indent, moduleName)
	newModulesSection := matches[1] + modulesContent + newModuleEntry + "    " + matches[3]

	p.content = modulesRegex.ReplaceAllString(p.content, newModulesSection)
	return nil
}

// AddProperty adds a property to the <properties> section
func (p *POMUpdater) AddProperty(name, value string) error {
	// Check if property already exists
	propRegex := regexp.MustCompile(fmt.Sprintf(`<%s>[^<]+</%s>`, regexp.QuoteMeta(name), regexp.QuoteMeta(name)))
	if propRegex.MatchString(p.content) {
		return nil // Already exists
	}

	// Find the properties section
	propertiesRegex := regexp.MustCompile(`(<properties>)([\s\S]*?)(</properties>)`)
	matches := propertiesRegex.FindStringSubmatch(p.content)

	if len(matches) < 4 {
		return fmt.Errorf("could not find <properties> section in POM")
	}

	propertiesContent := matches[2]

	// Find indentation from existing properties
	propEntryRegex := regexp.MustCompile(`(\s*)<[a-z]`)
	propMatches := propEntryRegex.FindStringSubmatch(propertiesContent)

	indent := "        " // Default indent
	if len(propMatches) > 0 {
		indent = propMatches[1]
	}

	// Add new property before </properties>
	newPropEntry := fmt.Sprintf("%s<%s>%s</%s>\n", indent, name, value, name)
	newPropertiesSection := matches[1] + propertiesContent + newPropEntry + "    " + matches[3]

	p.content = propertiesRegex.ReplaceAllString(p.content, newPropertiesSection)
	return nil
}

// AddDependency adds a dependency to the <dependencies> section
func (p *POMUpdater) AddDependency(groupID, artifactID, version string) error {
	// Find the dependencies section that is NOT inside dependencyManagement
	// We need to check for the dependency only within the main dependencies section
	depsRegex := regexp.MustCompile(`(?s)<dependencies>(.*?)</dependencies>`)

	// Find all dependencies sections
	allMatches := depsRegex.FindAllStringSubmatchIndex(p.content, -1)
	if len(allMatches) == 0 {
		return fmt.Errorf("could not find <dependencies> section in POM")
	}

	// Find the main dependencies section (not inside dependencyManagement)
	// The main dependencies section is typically after </dependencyManagement>
	depMgmtEnd := strings.Index(p.content, "</dependencyManagement>")
	var mainDepsStart, mainDepsEnd int
	found := false

	for _, match := range allMatches {
		start := match[0]
		end := match[1]
		// If dependencyManagement exists, main deps section should be after it
		// If not, take the first dependencies section
		if depMgmtEnd == -1 || start > depMgmtEnd {
			mainDepsStart = start
			mainDepsEnd = end
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("could not find main <dependencies> section in POM")
	}

	// Extract the main dependencies section content
	mainDepsContent := p.content[mainDepsStart:mainDepsEnd]

	// Check if dependency already exists in the main dependencies section
	if strings.Contains(mainDepsContent, fmt.Sprintf("<artifactId>%s</artifactId>", artifactID)) {
		return nil // Already exists in main dependencies
	}

	// Build the dependency entry
	var depBuilder strings.Builder
	depBuilder.WriteString("        <dependency>\n")
	depBuilder.WriteString(fmt.Sprintf("            <groupId>%s</groupId>\n", groupID))
	depBuilder.WriteString(fmt.Sprintf("            <artifactId>%s</artifactId>\n", artifactID))
	if version != "" {
		depBuilder.WriteString(fmt.Sprintf("            <version>%s</version>\n", version))
	}
	depBuilder.WriteString("        </dependency>\n")

	depEntry := depBuilder.String()

	// Find where to insert (before </dependencies> in the main section)
	insertPos := mainDepsEnd - len("</dependencies>")

	// Insert the new dependency
	p.content = p.content[:insertPos] + depEntry + "    " + p.content[insertPos:]

	return nil
}

// AddDependencyManagement adds a dependency to the dependencyManagement section
func (p *POMUpdater) AddDependencyManagement(groupID, artifactID, version, depType, scope string) error {
	// Build the dependency entry
	var depBuilder strings.Builder
	depBuilder.WriteString("            <dependency>\n")
	depBuilder.WriteString(fmt.Sprintf("                <groupId>%s</groupId>\n", groupID))
	depBuilder.WriteString(fmt.Sprintf("                <artifactId>%s</artifactId>\n", artifactID))
	depBuilder.WriteString(fmt.Sprintf("                <version>%s</version>\n", version))
	if depType != "" {
		depBuilder.WriteString(fmt.Sprintf("                <type>%s</type>\n", depType))
	}
	if scope != "" {
		depBuilder.WriteString(fmt.Sprintf("                <scope>%s</scope>\n", scope))
	}
	depBuilder.WriteString("            </dependency>\n")

	depEntry := depBuilder.String()

	// Check if dependency already exists
	if strings.Contains(p.content, fmt.Sprintf("<artifactId>%s</artifactId>", artifactID)) {
		return nil // Already exists
	}

	// Find the dependencyManagement/dependencies section
	depMgmtRegex := regexp.MustCompile(`(<dependencyManagement>\s*<dependencies>)([\s\S]*?)(</dependencies>\s*</dependencyManagement>)`)
	matches := depMgmtRegex.FindStringSubmatch(p.content)

	if len(matches) < 4 {
		return fmt.Errorf("could not find <dependencyManagement> section in POM")
	}

	// Add new dependency
	newContent := matches[1] + matches[2] + depEntry + "        " + matches[3]
	p.content = depMgmtRegex.ReplaceAllString(p.content, newContent)
	return nil
}

// DockerComposeUpdater handles modifications to docker-compose.yml
type DockerComposeUpdater struct {
	path     string
	content  map[string]interface{}
	services map[string]interface{}
	volumes  map[string]interface{}
}

// NewDockerComposeUpdater creates a new DockerComposeUpdater
func NewDockerComposeUpdater(composePath string) (*DockerComposeUpdater, error) {
	data, err := os.ReadFile(composePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create new docker-compose structure
			return &DockerComposeUpdater{
				path: composePath,
				content: map[string]interface{}{
					"services": map[string]interface{}{},
				},
				services: map[string]interface{}{},
				volumes:  map[string]interface{}{},
			}, nil
		}
		return nil, fmt.Errorf("failed to read docker-compose file: %w", err)
	}

	var content map[string]interface{}
	if err := yaml.Unmarshal(data, &content); err != nil {
		return nil, fmt.Errorf("failed to parse docker-compose file: %w", err)
	}

	services := map[string]interface{}{}
	if s, ok := content["services"].(map[string]interface{}); ok {
		services = s
	}

	volumes := map[string]interface{}{}
	if v, ok := content["volumes"].(map[string]interface{}); ok {
		volumes = v
	}

	return &DockerComposeUpdater{
		path:     composePath,
		content:  content,
		services: services,
		volumes:  volumes,
	}, nil
}

// Save writes the modified docker-compose.yml back to disk
func (d *DockerComposeUpdater) Save() error {
	d.content["services"] = d.services
	if len(d.volumes) > 0 {
		d.content["volumes"] = d.volumes
	}

	data, err := yaml.Marshal(d.content)
	if err != nil {
		return fmt.Errorf("failed to marshal docker-compose: %w", err)
	}

	return os.WriteFile(d.path, data, 0644)
}

// HasService checks if a service exists
func (d *DockerComposeUpdater) HasService(name string) bool {
	_, ok := d.services[name]
	return ok
}

// AddService adds a new service
func (d *DockerComposeUpdater) AddService(name string, config map[string]interface{}) {
	d.services[name] = config
}

// AddVolume adds a named volume (with empty config, which is standard for named volumes)
func (d *DockerComposeUpdater) AddVolume(name string) {
	d.volumes[name] = nil
}

// RemoveService removes a service
func (d *DockerComposeUpdater) RemoveService(name string) {
	delete(d.services, name)
}

// GetPostgresService returns a PostgreSQL service configuration
// hostPort allows customization to avoid conflicts (use 5433 for main db, 5434 for jobrunr)
func GetPostgresService(serviceName, database, user, password string, hostPort int) map[string]interface{} {
	return map[string]interface{}{
		"image": "postgres:16-alpine",
		"ports": []string{fmt.Sprintf("%d:5432", hostPort)},
		"environment": map[string]string{
			"POSTGRES_DB":       database,
			"POSTGRES_USER":     user,
			"POSTGRES_PASSWORD": password,
		},
		"volumes": []string{serviceName + "-data:/var/lib/postgresql/data"},
		"healthcheck": map[string]interface{}{
			"test":     []string{"CMD-SHELL", "pg_isready -U " + user + " -d " + database},
			"interval": "5s",
			"timeout":  "5s",
			"retries":  5,
		},
	}
}

// GetMySQLService returns a MySQL service configuration
// Uses port 3307 on host to avoid conflicts with local MySQL installations
// Only root user is created to match application.yml template defaults (username: root, password: root)
func GetMySQLService(serviceName, database, rootPassword string) map[string]interface{} {
	return map[string]interface{}{
		"image": "mysql:8.0",
		"ports": []string{"3307:3306"},
		"environment": map[string]string{
			"MYSQL_ROOT_PASSWORD": rootPassword,
			"MYSQL_DATABASE":      database,
		},
		"volumes": []string{serviceName + "-data:/var/lib/mysql"},
		"healthcheck": map[string]interface{}{
			"test":     []string{"CMD", "mysqladmin", "ping", "-h", "localhost"},
			"interval": "5s",
			"timeout":  "5s",
			"retries":  5,
		},
		// Use mysql_native_password for Java driver compatibility
		"command": "--default-authentication-plugin=mysql_native_password",
	}
}

// GetMongoDBService returns a MongoDB service configuration
// No authentication for local development (matches docker-compose template)
func GetMongoDBService(serviceName, database string) map[string]interface{} {
	return map[string]interface{}{
		"image": "mongo:7",
		"ports": []string{"27017:27017"},
		"environment": map[string]string{
			"MONGO_INITDB_DATABASE": database,
		},
		"volumes": []string{serviceName + "-data:/data/db"},
	}
}

// GetRedisService returns a Redis service configuration
func GetRedisService(serviceName string) map[string]interface{} {
	return map[string]interface{}{
		"image":   "redis:7-alpine",
		"ports":   []string{"6379:6379"},
		"volumes": []string{serviceName + "-data:/data"},
	}
}

// GetKafkaService returns Kafka service configurations (Kafka + Zookeeper)
func GetKafkaService() (kafka, zookeeper map[string]interface{}) {
	zookeeper = map[string]interface{}{
		"image": "confluentinc/cp-zookeeper:" + ConfluentKafkaVersion,
		"environment": map[string]string{
			"ZOOKEEPER_CLIENT_PORT": "2181",
			"ZOOKEEPER_TICK_TIME":   "2000",
		},
	}

	kafka = map[string]interface{}{
		"image":      "confluentinc/cp-kafka:" + ConfluentKafkaVersion,
		"depends_on": []string{"zookeeper"},
		"ports":      []string{"9092:9092"},
		"environment": map[string]string{
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_ZOOKEEPER_CONNECT":                "zookeeper:2181",
			"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT://localhost:9092",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
		},
	}

	return kafka, zookeeper
}

// GetRabbitMQService returns a RabbitMQ service configuration
func GetRabbitMQService(user, password string) map[string]interface{} {
	return map[string]interface{}{
		"image": "rabbitmq:3-management-alpine",
		"ports": []string{"5672:5672", "15672:15672"},
		"environment": map[string]string{
			"RABBITMQ_DEFAULT_USER": user,
			"RABBITMQ_DEFAULT_PASS": password,
		},
		"volumes": []string{"rabbitmq-data:/var/lib/rabbitmq"},
	}
}

// GetLocalStackService returns a LocalStack service configuration for SQS
func GetLocalStackService() map[string]interface{} {
	return map[string]interface{}{
		"image": "localstack/localstack:" + LocalStackImageVersion,
		"ports": []string{"4566:4566"},
		"environment": map[string]string{
			"SERVICES":       "sqs",
			"DEFAULT_REGION": "us-east-1",
		},
		"volumes": []string{"./localstack-init:/etc/localstack/init"},
	}
}

// GetPubSubEmulatorService returns a Pub/Sub emulator service configuration
func GetPubSubEmulatorService() map[string]interface{} {
	return map[string]interface{}{
		"image": "gcr.io/google.com/cloudsdktool/google-cloud-cli:emulators",
		"ports": []string{"8085:8085"},
		"command": []string{
			"gcloud", "beta", "emulators", "pubsub", "start",
			"--host-port=0.0.0.0:8085",
		},
	}
}

// EnvUpdater handles modifications to .env.example files
type EnvUpdater struct {
	path    string
	entries map[string]string
	order   []string
}

// NewEnvUpdater creates a new EnvUpdater
func NewEnvUpdater(envPath string) (*EnvUpdater, error) {
	updater := &EnvUpdater{
		path:    envPath,
		entries: make(map[string]string),
		order:   []string{},
	}

	data, err := os.ReadFile(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			return updater, nil // Empty file
		}
		return nil, err
	}

	// Parse existing entries
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			updater.entries[key] = value
			updater.order = append(updater.order, key)
		}
	}

	return updater, nil
}

// Save writes the modified .env.example back to disk
func (e *EnvUpdater) Save() error {
	var lines []string
	for _, key := range e.order {
		if value, ok := e.entries[key]; ok {
			lines = append(lines, fmt.Sprintf("%s=%s", key, value))
		}
	}

	content := strings.Join(lines, "\n") + "\n"
	return os.WriteFile(e.path, []byte(content), 0644)
}

// Set adds or updates an environment variable
func (e *EnvUpdater) Set(key, value string) {
	if _, exists := e.entries[key]; !exists {
		e.order = append(e.order, key)
	}
	e.entries[key] = value
}

// AddModuleToPOM adds a module to the parent POM
func AddModuleToPOM(projectPath, moduleName string) error {
	pomPath := filepath.Join(projectPath, "pom.xml")
	updater, err := NewPOMUpdater(pomPath)
	if err != nil {
		return err
	}

	if err := updater.AddModule(moduleName); err != nil {
		return err
	}

	return updater.Save()
}

// AddPropertyToPOM adds a property to the parent POM
func AddPropertyToPOM(projectPath, name, value string) error {
	pomPath := filepath.Join(projectPath, "pom.xml")
	updater, err := NewPOMUpdater(pomPath)
	if err != nil {
		return err
	}

	if err := updater.AddProperty(name, value); err != nil {
		return err
	}

	return updater.Save()
}
