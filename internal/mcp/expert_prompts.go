package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAllPrompts(s *server.MCPServer) {
	registerTrabucoExpert(s)
	registerDesignMicroservices(s)
	registerExtendProject(s)
}

func registerTrabucoExpert(s *server.MCPServer) {
	s.AddPrompt(mcp.NewPrompt("trabuco_expert",
		mcp.WithPromptDescription("Get expert instructions for using Trabuco effectively for a specific task"),
		mcp.WithArgument("task",
			mcp.ArgumentDescription("What you need help with (e.g., 'create a new microservice', 'add event processing', 'choose a database')"),
			mcp.RequiredArgument(),
		),
	), func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		task := req.Params.Arguments["task"]
		if task == "" {
			return nil, fmt.Errorf("task argument is required")
		}

		text := fmt.Sprintf(`You are an expert architect using Trabuco to help with: %s

TRABUCO MODULE DECISION TREE:
1. Does the user need HTTP endpoints? → Add API
2. Does the user need SQL persistence? → Add SQLDatastore (pick postgresql or mysql)
3. Does the user need NoSQL persistence? → Add NoSQLDatastore (pick mongodb or redis)
   - SQLDatastore and NoSQLDatastore are MUTUALLY EXCLUSIVE
4. Does the user need business logic orchestration? → Add Shared
5. Does the user need background jobs? → Add Worker (uses SQL database for job storage)
6. Does the user need message broker consumers? → Add EventConsumer (pick kafka, rabbitmq, sqs, or pubsub)
7. Does the user need AI tool integration? → Add MCP

COMMON PITFALLS TO AVOID:
- Do NOT select both SQLDatastore and NoSQLDatastore — they conflict
- Worker requires a SQL database for job storage — if you pick Worker without SQLDatastore, it defaults to PostgreSQL
- EventConsumer requires specifying a message_broker parameter
- "events" in user requirements may mean webhooks (use API) not message broker consumers (EventConsumer)
- "queue" may mean internal job queue (Worker) not external message queue (EventConsumer)
- "cache" may mean application-level caching (not generated) not Redis data store (NoSQLDatastore)

WHAT TRABUCO GENERATES:
- Multi-module Maven project with Spring Boot 3.4
- Spring Data JDBC (not JPA), Immutables for DTOs/entities
- Testcontainers for integration tests, Spotless for formatting
- Docker Compose for local development
- CI workflow, AI context files, code quality enforcement

WHAT TRABUCO DOES NOT GENERATE:
- Authentication/authorization (add Spring Security manually)
- Frontend/UI (backend only)
- GraphQL, gRPC, WebSockets (REST only)
- Kubernetes manifests, Terraform, cloud deployment
- Custom business logic or production database schemas
- Rate limiting, multi-tenancy, API versioning

POST-GENERATION STEPS:
1. Replace placeholder entities in Model/ with real domain objects
2. Update Flyway migrations with your actual schema (if using SQLDatastore)
3. Implement business logic in Shared/src/main/java/.../shared/service/
4. Read AGENTS.md for coding patterns and conventions
5. Read .ai/prompts/ for step-by-step guides (add-entity, add-endpoint, add-job, add-event)
6. Run 'mvn test' to verify everything compiles
7. Run 'mvn spotless:apply' after making changes

RECOMMENDED WORKFLOW:
1. Call suggest_architecture with the user's requirements to get pattern matches
2. Review the recommended_config — adjust if needed
3. Call init_project with the chosen configuration
4. Follow the next_steps and key_files in the response`, task)

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Expert guidance for: %s", task),
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.TextContent{Type: "text", Text: text},
				},
			},
		}, nil
	})
}

func registerDesignMicroservices(s *server.MCPServer) {
	s.AddPrompt(mcp.NewPrompt("design_microservices",
		mcp.WithPromptDescription("Get step-by-step instructions for decomposing requirements into multiple independent Trabuco services"),
		mcp.WithArgument("requirements",
			mcp.ArgumentDescription("Description of the system to decompose into microservices"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("service_count",
			mcp.ArgumentDescription("Approximate number of services to target (optional, e.g., '3')"),
		),
	), func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		requirements := req.Params.Arguments["requirements"]
		if requirements == "" {
			return nil, fmt.Errorf("requirements argument is required")
		}
		serviceCount := req.Params.Arguments["service_count"]

		countNote := ""
		if serviceCount != "" {
			countNote = fmt.Sprintf("\nTarget approximately %s services.", serviceCount)
		}

		text := fmt.Sprintf(`You are designing a microservices system using Trabuco. Requirements: %s%s

STEP-BY-STEP DECOMPOSITION:

1. IDENTIFY SERVICE BOUNDARIES
   - Each service should own a single business domain (e.g., users, orders, notifications)
   - Services should be independently deployable and scalable
   - Avoid shared databases between services — each service owns its data

2. FOR EACH SERVICE, DETERMINE:
   - Name: lowercase-with-hyphens (e.g., user-service, order-service)
   - Group ID: use a shared prefix (e.g., com.company.platform)
   - Which Trabuco pattern fits: rest-api, event-driven, worker-only, etc.
   - Which modules it needs
   - Which database/broker it requires

3. INTER-SERVICE COMMUNICATION
   - REST calls: direct HTTP between services (simple but creates coupling)
   - Event-driven: services publish events, others subscribe (decoupled, recommended)
   - If using events, all services sharing events should use the same broker type
   - Consider: which service is the source of truth for each entity?

4. SHARED INFRASTRUCTURE
   - All services can share a single Docker Compose for local development
   - Use generate_workspace to create a workspace with all services
   - Shared database instances (e.g., one PostgreSQL server, separate databases per service)
   - Shared message broker instance if using events

5. WORKSPACE LAYOUT
   workspace-dir/
   ├── docker-compose.yml        (shared infrastructure)
   ├── service-a/                (Trabuco project)
   │   ├── Model/
   │   ├── SQLDatastore/
   │   ├── Shared/
   │   └── API/
   ├── service-b/                (Trabuco project)
   │   ├── Model/
   │   ├── SQLDatastore/
   │   ├── Shared/
   │   ├── API/
   │   └── EventConsumer/
   └── service-c/                (Trabuco project)
       ├── Model/
       ├── Shared/
       └── Worker/

6. USE TRABUCO TOOLS:
   - Call design_system with the full requirements to get a structured design
   - Review the design with the user
   - Call generate_workspace to create all services at once
   - Or call init_project separately for each service if you prefer manual control

COMMON PATTERNS:
- API Gateway + Backend Services: One lightweight API service that routes to domain services
- Event Bus: Multiple services connected through a shared message broker
- CQRS: Separate read (API) and write (EventConsumer + Worker) services`, requirements, countNote)

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Microservices design guide for: %s", requirements),
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.TextContent{Type: "text", Text: text},
				},
			},
		}, nil
	})
}

func registerExtendProject(s *server.MCPServer) {
	s.AddPrompt(mcp.NewPrompt("extend_project",
		mcp.WithPromptDescription("Get instructions for extending an existing Trabuco project with a new feature"),
		mcp.WithArgument("project_path",
			mcp.ArgumentDescription("Path to the existing Trabuco project"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("feature",
			mcp.ArgumentDescription("The feature to add (e.g., 'background email sending', 'Kafka event processing')"),
			mcp.RequiredArgument(),
		),
	), func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		projectPath := req.Params.Arguments["project_path"]
		feature := req.Params.Arguments["feature"]
		if projectPath == "" {
			return nil, fmt.Errorf("project_path argument is required")
		}
		if feature == "" {
			return nil, fmt.Errorf("feature argument is required")
		}

		text := fmt.Sprintf(`You are extending an existing Trabuco project at: %s
Feature to add: %s

STEP-BY-STEP EXTENSION GUIDE:

1. ASSESS CURRENT STATE
   - Call get_project_info with path="%s" to see installed modules
   - Check which modules are already present
   - Determine what new modules are needed for the feature

2. DETERMINE WHICH MODULE TO ADD
   - Background jobs → Worker module (adds JobRunr)
   - Message processing → EventConsumer module (needs a broker: kafka, rabbitmq, sqs, pubsub)
   - SQL persistence → SQLDatastore module (needs database: postgresql or mysql)
   - NoSQL persistence → NoSQLDatastore module (needs nosql_database: mongodb or redis)
   - REST endpoints → API module
   - Business logic → Shared module
   - AI tooling → MCP module

3. ADD THE MODULE
   - First preview: call add_module with dry_run=true to see what will change
   - Then execute: call add_module with the required parameters
   - Dependencies are resolved automatically (e.g., adding Worker also adds Jobs)

4. IMPLEMENT THE FEATURE
   - Read the generated playbook files in .ai/prompts/:
     * add-entity.md — for new domain objects
     * add-endpoint.md — for new REST endpoints (if API module present)
     * add-job.md — for new background jobs (if Worker module present)
     * add-event.md — for new event consumers (if EventConsumer module present)
   - Follow AGENTS.md for coding conventions
   - Follow .ai/prompts/JAVA_CODE_QUALITY.md for quality standards

5. VERIFY
   - Run 'mvn test' to ensure everything compiles and tests pass
   - Run 'mvn spotless:apply' to format code
   - Check run_doctor to validate project health

IMPORTANT CONSTRAINTS:
- SQLDatastore and NoSQLDatastore cannot coexist in the same project
- If the project already has SQLDatastore, you cannot add NoSQLDatastore (and vice versa)
- Worker uses the SQL database for job storage
- Adding a module creates backup of modified files automatically`, projectPath, feature, projectPath)

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("Extension guide for adding '%s' to project at %s", feature, projectPath),
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.TextContent{Type: "text", Text: text},
				},
			},
		}, nil
	})
}
