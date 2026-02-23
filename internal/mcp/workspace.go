package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/generator"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// systemDesign is the output of design_system — a structured decomposition of requirements.
type systemDesign struct {
	Services           []serviceDesign `json:"services"`
	SharedInfra        []string        `json:"shared_infrastructure"`
	CommunicationNotes []string        `json:"communication_notes"`
	Warnings           []string        `json:"warnings"`
}

// serviceDesign describes a single service in a multi-service system.
type serviceDesign struct {
	Name          string   `json:"name"`
	Purpose       string   `json:"purpose"`
	Pattern       string   `json:"pattern"`
	Modules       string   `json:"modules"`
	Database      string   `json:"database,omitempty"`
	NoSQLDatabase string   `json:"nosql_database,omitempty"`
	MessageBroker string   `json:"message_broker,omitempty"`
	GroupID       string   `json:"group_id"`
	Boundaries    []string `json:"boundaries"`
}

func registerDesignSystem(s *server.MCPServer) {
	tool := mcp.NewTool("design_system",
		mcp.WithDescription(
			"Decompose natural language requirements into a multi-service system design. "+
				"Returns a structured design with service definitions, patterns, and infrastructure notes. "+
				"Does NOT generate any code — returns a design document for review before calling generate_workspace. "+
				"Use this when the user describes a system that needs multiple independent services.",
		),
		mcp.WithString("requirements",
			mcp.Description("Natural language description of the multi-service system requirements"),
			mcp.Required(),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		requirements := req.GetString("requirements", "")
		if requirements == "" {
			return toolError("requirements parameter is required"), nil
		}

		design := buildSystemDesign(requirements)
		return toolJSON(design)
	})
}

func registerGenerateWorkspace(s *server.MCPServer, version string) {
	tool := mcp.NewTool("generate_workspace",
		mcp.WithDescription(
			"Generate a multi-service workspace with shared infrastructure. "+
				"Creates a workspace directory containing multiple Trabuco projects and a shared docker-compose.yml. "+
				"Use design_system first to plan the services, then call this with the service configurations. "+
				"Each service is generated using the same engine as init_project.",
		),
		mcp.WithString("services",
			mcp.Description("JSON array of service configs. Each object: {name, modules, group_id, database?, nosql_database?, message_broker?, java_version?}"),
			mcp.Required(),
		),
		mcp.WithString("workspace_dir",
			mcp.Description("Directory to create the workspace in (will be created if it doesn't exist)"),
			mcp.Required(),
		),
		mcp.WithString("group_id_prefix",
			mcp.Description("Common group ID prefix for all services (e.g., 'com.company.platform')"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		servicesJSON := req.GetString("services", "")
		workspaceDir := req.GetString("workspace_dir", "")
		groupIDPrefix := req.GetString("group_id_prefix", "")

		if servicesJSON == "" {
			return toolError("services parameter is required"), nil
		}
		if workspaceDir == "" {
			return toolError("workspace_dir parameter is required"), nil
		}

		// Parse service configs
		var services []serviceConfig
		if err := json.Unmarshal([]byte(servicesJSON), &services); err != nil {
			return toolError(fmt.Sprintf("Invalid services JSON: %v", err)), nil
		}
		if len(services) == 0 {
			return toolError("At least one service must be specified"), nil
		}

		// Resolve workspace path
		absWorkspace, err := resolvePath(workspaceDir)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to resolve workspace path: %v", err)), nil
		}

		// Create workspace directory
		if err := os.MkdirAll(absWorkspace, 0755); err != nil {
			return toolError(fmt.Sprintf("Failed to create workspace directory: %v", err)), nil
		}

		// Validate all services before generating any
		for i, svc := range services {
			if svc.Name == "" {
				return toolError(fmt.Sprintf("Service %d: name is required", i)), nil
			}
			if !projectNameRegex.MatchString(svc.Name) {
				return toolError(fmt.Sprintf("Service %d: invalid name '%s'. Must be lowercase, alphanumeric, hyphens allowed.", i, svc.Name)), nil
			}
			if svc.Modules == "" {
				return toolError(fmt.Sprintf("Service '%s': modules is required", svc.Name)), nil
			}

			// Apply group ID prefix if service doesn't specify its own
			if svc.GroupID == "" && groupIDPrefix != "" {
				services[i].GroupID = groupIDPrefix + "." + strings.ReplaceAll(svc.Name, "-", "")
			}
			if services[i].GroupID == "" {
				return toolError(fmt.Sprintf("Service '%s': group_id is required (provide per-service or use group_id_prefix)", svc.Name)), nil
			}
			if !groupIDRegex.MatchString(services[i].GroupID) {
				return toolError(fmt.Sprintf("Service '%s': invalid group_id '%s'", svc.Name, services[i].GroupID)), nil
			}

			// Validate modules
			modules := strings.Split(svc.Modules, ",")
			for j := range modules {
				modules[j] = strings.TrimSpace(modules[j])
			}
			if validationErr := config.ValidateModuleSelection(modules); validationErr != "" {
				return toolError(fmt.Sprintf("Service '%s': %s", svc.Name, validationErr)), nil
			}

			// Check for directory conflicts
			svcPath := filepath.Join(absWorkspace, svc.Name)
			if _, err := os.Stat(svcPath); !os.IsNotExist(err) {
				return toolError(fmt.Sprintf("Service '%s': directory already exists at %s", svc.Name, svcPath)), nil
			}
		}

		// Generate each service
		var generatedServices []map[string]any
		for _, svc := range services {
			modules := strings.Split(svc.Modules, ",")
			for j := range modules {
				modules[j] = strings.TrimSpace(modules[j])
			}
			resolvedModules := config.ResolveDependencies(modules)

			javaVersion := svc.JavaVersion
			if javaVersion == "" {
				javaVersion = "21"
			}

			cfg := &config.ProjectConfig{
				ProjectName:   svc.Name,
				GroupID:       svc.GroupID,
				ArtifactID:    svc.Name,
				JavaVersion:   javaVersion,
				Modules:       resolvedModules,
				Database:      svc.Database,
				NoSQLDatabase: svc.NoSQLDatabase,
				MessageBroker: svc.MessageBroker,
			}

			outDir := filepath.Join(absWorkspace, svc.Name)
			gen, err := generator.NewWithVersionAt(cfg, version, outDir)
			if err != nil {
				return toolError(fmt.Sprintf("Service '%s': failed to create generator: %v", svc.Name, err)), nil
			}

			if err := gen.Generate(); err != nil {
				return toolError(fmt.Sprintf("Service '%s': generation failed: %v", svc.Name, err)), nil
			}

			generatedServices = append(generatedServices, map[string]any{
				"name":    svc.Name,
				"path":    outDir,
				"modules": resolvedModules,
			})
		}

		// Generate shared docker-compose.yml
		composePath := filepath.Join(absWorkspace, "docker-compose.yml")
		composeContent := buildSharedDockerCompose(services)
		if err := os.WriteFile(composePath, []byte(composeContent), 0644); err != nil {
			return toolError(fmt.Sprintf("Failed to write shared docker-compose.yml: %v", err)), nil
		}

		return toolJSON(map[string]any{
			"status":         "success",
			"workspace":      absWorkspace,
			"services":       generatedServices,
			"docker_compose": composePath,
			"next_steps": []string{
				"Review each service's AGENTS.md for coding patterns",
				"Replace placeholder entities in each service's Model/",
				"Configure inter-service communication (REST calls or shared broker)",
				"Run 'docker compose up' from the workspace root to start shared infrastructure",
				"Run 'mvn test' in each service directory to verify compilation",
			},
		})
	})
}

// serviceConfig is the input format for generate_workspace.
type serviceConfig struct {
	Name          string `json:"name"`
	Modules       string `json:"modules"`
	GroupID       string `json:"group_id"`
	Database      string `json:"database,omitempty"`
	NoSQLDatabase string `json:"nosql_database,omitempty"`
	MessageBroker string `json:"message_broker,omitempty"`
	JavaVersion   string `json:"java_version,omitempty"`
}

// buildSystemDesign decomposes requirements into a multi-service design.
func buildSystemDesign(requirements string) *systemDesign {
	lower := strings.ToLower(requirements)

	var services []serviceDesign
	var sharedInfra []string
	var commNotes []string
	var warnings []string

	// Detect individual service mentions.
	// Ordered longest-first so "api gateway" matches before "gateway",
	// "notification service" before "notification", etc.
	type serviceKeyword struct {
		keyword string
		shorter string // if this matches, skip the shorter variant
		purpose string
		pattern string
	}
	serviceKeywordList := []serviceKeyword{
		// Routing / gateway
		{"api gateway", "gateway", "Route and aggregate requests to backend services", "microservice-light"},
		{"gateway", "", "Route and aggregate requests to backend services", "microservice-light"},
		// User / auth
		{"user service", "", "Manage user accounts and profiles", "rest-api"},
		{"auth service", "", "Handle authentication and authorization", "rest-api"},
		// Commerce
		{"order service", "", "Manage orders and order lifecycle", "rest-api"},
		{"payment service", "", "Process payments and billing", "rest-api"},
		{"billing service", "", "Handle billing, invoices, and subscriptions", "rest-api"},
		{"inventory service", "", "Track product inventory", "rest-api"},
		{"catalog service", "", "Manage product catalog", "rest-api"},
		// Notifications / email
		{"notification service", "notification", "Send notifications (email, push, SMS)", "background-processing"},
		{"notification", "", "Send notifications (email, push, SMS)", "background-processing"},
		{"email service", "", "Send and manage emails", "background-processing"},
		// Search / data
		{"search service", "", "Full-text search and indexing", "rest-api-nosql"},
		{"cache service", "", "Caching and session storage", "rest-api-nosql"},
		{"data service", "", "Data access and persistence layer", "rest-api"},
		// Background / scheduling
		{"scheduling service", "", "Manage scheduled tasks and cron jobs", "background-processing"},
		{"workflow service", "", "Orchestrate multi-step business workflows", "background-processing"},
		{"import service", "", "Handle data imports and batch ingestion", "worker-only"},
		// Event-driven / analytics
		{"analytics service", "analytics", "Collect and process analytics", "event-driven"},
		{"analytics", "", "Collect and process analytics data", "event-driven"},
		{"logging service", "", "Centralized logging and monitoring", "event-driven"},
		{"audit service", "", "Track and record audit events", "event-driven"},
		// Admin / integration
		{"admin service", "", "Internal administration and management tools", "rest-api"},
		{"integration service", "", "Connect with external systems and third-party APIs", "rest-api"},
		{"webhook service", "", "Receive and process incoming webhooks", "rest-api"},
		{"reporting service", "", "Generate reports and data aggregations", "background-processing"},
	}

	matchedKeywords := make(map[string]bool) // tracks which keywords were matched
	for _, entry := range serviceKeywordList {
		if !strings.Contains(lower, entry.keyword) {
			continue
		}
		// Skip if a longer variant already matched this keyword
		name := strings.ReplaceAll(entry.keyword, " ", "-")
		if matchedKeywords[name] {
			continue
		}

		pattern := findPattern(entry.pattern)
		if pattern == nil {
			continue
		}

		matchedKeywords[name] = true
		// Mark the shorter variant as consumed to prevent duplicates
		if entry.shorter != "" {
			shorterName := strings.ReplaceAll(entry.shorter, " ", "-")
			matchedKeywords[shorterName] = true
		}

		svc := serviceDesign{
			Name:          name,
			Purpose:       entry.purpose,
			Pattern:       pattern.Name,
			Modules:       strings.Join(pattern.Modules, ","),
			Database:      pattern.RecommendedDB,
			NoSQLDatabase: pattern.RecommendedNoDB,
			MessageBroker: pattern.RecommendedBrkr,
			GroupID:       "com.company." + strings.ReplaceAll(name, "-", ""),
			Boundaries:    []string{"Owns its own data", "Communicates via REST or events"},
		}
		services = append(services, svc)
	}

	// If no specific services detected, provide a generic decomposition
	if len(services) == 0 {
		// Score against patterns to at least suggest a single service
		scored := scorePatterns(requirements)
		if len(scored) > 0 {
			top := scored[0]
			services = append(services, serviceDesign{
				Name:          "main-service",
				Purpose:       "Primary service based on requirements",
				Pattern:       top.Name,
				Modules:       strings.Join(top.Modules, ","),
				Database:      top.RecommendedDB,
				NoSQLDatabase: top.RecommendedNoDB,
				MessageBroker: top.RecommendedBrkr,
				GroupID:       "com.company.mainservice",
				Boundaries:    []string{"Single service — consider splitting as the system grows"},
			})
		}
		warnings = append(warnings, "Could not identify distinct service boundaries. Consider describing specific services (e.g., 'user service', 'notification service').")
	}

	// Detect shared infrastructure needs
	needsDB := false
	needsBroker := false
	for _, svc := range services {
		if svc.Database != "" {
			needsDB = true
		}
		if svc.MessageBroker != "" {
			needsBroker = true
		}
	}

	if needsDB {
		sharedInfra = append(sharedInfra, "PostgreSQL (shared instance, separate databases per service)")
	}
	if needsBroker {
		sharedInfra = append(sharedInfra, "Message broker (shared instance for inter-service events)")
	}

	// Communication notes
	if len(services) > 1 {
		commNotes = append(commNotes,
			"Services should communicate via REST calls for synchronous operations",
			"Use events (via shared message broker) for asynchronous, decoupled communication",
			"Each service owns its data — avoid direct database access between services",
			"Consider API Gateway pattern if clients need a single entry point",
		)
	}

	// Detect unsupported requirements
	unsupported := detectUnsupported(lower)
	for _, u := range unsupported {
		warnings = append(warnings, u)
	}

	if warnings == nil {
		warnings = []string{}
	}
	if sharedInfra == nil {
		sharedInfra = []string{}
	}
	if commNotes == nil {
		commNotes = []string{}
	}

	return &systemDesign{
		Services:           services,
		SharedInfra:        sharedInfra,
		CommunicationNotes: commNotes,
		Warnings:           warnings,
	}
}

// findPattern returns a pattern by name from the catalog.
func findPattern(name string) *ArchitecturePattern {
	for i := range patternCatalog {
		if patternCatalog[i].Name == name {
			return &patternCatalog[i]
		}
	}
	return nil
}

// buildSharedDockerCompose generates a docker-compose.yml for shared infrastructure.
func buildSharedDockerCompose(services []serviceConfig) string {
	var b strings.Builder
	b.WriteString("# Shared infrastructure for multi-service workspace\n")
	b.WriteString("# Generated by Trabuco — customize as needed\n")
	b.WriteString("services:\n")

	needsPostgres := false
	needsMySQL := false
	needsMongo := false
	needsRedis := false
	needsKafka := false
	needsRabbitMQ := false

	for _, svc := range services {
		switch svc.Database {
		case "postgresql":
			needsPostgres = true
		case "mysql":
			needsMySQL = true
		}
		switch svc.NoSQLDatabase {
		case "mongodb":
			needsMongo = true
		case "redis":
			needsRedis = true
		}
		switch svc.MessageBroker {
		case "kafka":
			needsKafka = true
		case "rabbitmq":
			needsRabbitMQ = true
		}
	}

	if needsPostgres {
		b.WriteString(`
  postgres:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: trabuco
      POSTGRES_PASSWORD: trabuco
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./init-databases.sh:/docker-entrypoint-initdb.d/init-databases.sh
`)
	}

	if needsMySQL {
		b.WriteString(`
  mysql:
    image: mysql:8.4
    environment:
      MYSQL_ROOT_PASSWORD: trabuco
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
`)
	}

	if needsMongo {
		b.WriteString(`
  mongodb:
    image: mongo:7
    ports:
      - "27017:27017"
    volumes:
      - mongo_data:/data/db
`)
	}

	if needsRedis {
		b.WriteString(`
  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
`)
	}

	if needsKafka {
		b.WriteString(`
  kafka:
    image: confluentinc/cp-kafka:7.7.0
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://localhost:9092
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@localhost:9093
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      CLUSTER_ID: trabuco-workspace
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
    ports:
      - "9092:9092"
`)
	}

	if needsRabbitMQ {
		b.WriteString(`
  rabbitmq:
    image: rabbitmq:4-management-alpine
    ports:
      - "5672:5672"
      - "15672:15672"
`)
	}

	// Volumes section
	var volumes []string
	if needsPostgres {
		volumes = append(volumes, "  postgres_data:")
	}
	if needsMySQL {
		volumes = append(volumes, "  mysql_data:")
	}
	if needsMongo {
		volumes = append(volumes, "  mongo_data:")
	}
	if needsRedis {
		volumes = append(volumes, "  redis_data:")
	}

	if len(volumes) > 0 {
		b.WriteString("\nvolumes:\n")
		for _, v := range volumes {
			b.WriteString(v + "\n")
		}
	}

	return b.String()
}
