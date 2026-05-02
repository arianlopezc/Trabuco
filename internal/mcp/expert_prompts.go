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
	registerAIAgentExpert(s)
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
7. Does the user need AI/LLM capabilities? → Add AIAgent (tool calling, guardrails, multi-agent, MCP server, A2A)
8. Does the user need vector search / RAG? → Add AIAgent + pass --vector-store=pgvector|qdrant|mongodb
   - pgvector: same Postgres datastore (auto-adds SQLDatastore + forces postgresql)
   - qdrant: standalone gRPC server, best raw perf
   - mongodb: Atlas Vector Search ($vectorSearch is Atlas-only — community Mongo cannot serve queries)

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
- OIDC Resource Server scaffolding when API or AIAgent is selected
  (auto-generated and dormant; activated by trabuco.auth.enabled=true).
  IdentityClaims/AuthorityScope in Model, JwtClaimsExtractor +
  RequestContextHolder + AuthContextPropagator in Shared, dual
  SecurityFilterChain (JWT or permit-all) in API/AIAgent. Works
  provider-agnostic against Keycloak / Auth0 / Okta / Cognito /
  generic OIDC. Full guide: docs/auth.md.

WHAT TRABUCO DOES NOT GENERATE:
- Identity-provider side (login forms, token issuance, MFA, user
  management). Trabuco's auth is resource-server only — pair it
  with a hosted IdP.
- Frontend/UI (backend only)
- GraphQL, gRPC, WebSockets (REST only)
- Kubernetes manifests, Terraform, cloud deployment
- Custom business logic or production database schemas
- Rate limiting (available in AIAgent module; for API-only projects, add manually), multi-tenancy, API versioning

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

   IF PROJECT HAS AIAgent MODULE:
   - To add a new @Tool method: create a new method annotated with @Tool in the tools package, replacing or extending PlaceholderTools
   - To update the system prompt: modify the system prompt text in PrimaryAgent configuration
   - To add domain-specific guardrail rules: add ALLOW/BLOCK rules to the input guardrail configuration
   - To add a new A2A skill: register a new skill in the A2A agent card and implement the corresponding handler

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

func registerAIAgentExpert(s *server.MCPServer) {
	s.AddPrompt(mcp.NewPrompt("trabuco_ai_agent_expert",
		mcp.WithPromptDescription("Get expert instructions for building and customizing AI agents with Trabuco's AIAgent module"),
		mcp.WithArgument("task",
			mcp.ArgumentDescription("What you need help with (e.g., 'add a custom tool', 'configure guardrails', 'set up multi-agent', 'connect MCP client')"),
			mcp.RequiredArgument(),
		),
	), func(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		task := req.Params.Arguments["task"]
		if task == "" {
			return nil, fmt.Errorf("task argument is required")
		}

		text := fmt.Sprintf(`You are an expert on Trabuco's AIAgent module. The user needs help with: %s

AIAGENT MODULE STRUCTURE:
- aiagent/src/main/java/.../aiagent/
  - agent/         — PrimaryAgent (main agent), SpecialistAgent (sub-domain delegation)
  - config/        — Spring AI configuration, model settings, advisor chains
  - controller/    — REST endpoints: /chat, /ask, /a2a
  - guardrail/     — Input/output guardrails with ALLOW/BLOCK domain rules
  - mcp/           — MCP server configuration for Claude Code / Cursor integration
  - tools/         — @Tool methods (replace PlaceholderTools with real tools)
  - a2a/           — Agent-to-Agent protocol: agent card, skill handlers

HOW TO ADD CUSTOM @Tool METHODS:
1. Create a new class in the tools/ package (or add methods to an existing class)
2. Annotate each method with @Tool(description = "...") — the description is what the LLM sees
3. Use @ToolParam for parameter descriptions
4. Replace PlaceholderTools with your domain-specific tools
5. Register the tool bean in the Spring context (use @Component or @Bean)
6. The agent automatically discovers all @Tool beans via Spring AI's ToolCallbackProvider

HOW TO CUSTOMIZE THE SYSTEM PROMPT:
1. Find the PrimaryAgent configuration in agent/ package
2. Modify the system prompt text to match your domain
3. Include: what the agent does, what tools it has, how it should respond
4. Keep the prompt focused — long prompts increase token cost per request

HOW TO CONFIGURE INPUT GUARDRAILS:
1. Find the guardrail configuration in guardrail/ package
2. Add domain-specific ALLOW rules (topics the agent should handle)
3. Add BLOCK rules (topics the agent must refuse)
4. Guardrails run BEFORE the LLM call — they reject bad inputs early to save tokens
5. Output guardrails run AFTER the LLM response to filter sensitive content

HOW TO SET UP THE SPECIALIST AGENT:
1. The SpecialistAgent handles sub-domain delegation from the PrimaryAgent
2. Configure it with a focused system prompt for its specific domain
3. The PrimaryAgent delegates to specialists via tool calls
4. Each specialist has its own tools and guardrails

HOW TO CONNECT CLAUDE CODE / CURSOR VIA MCP:
1. The AIAgent module exposes an MCP server endpoint
2. In Claude Code: add the server URL to your MCP configuration
3. In Cursor: add the server to .cursor/mcp.json
4. The MCP server exposes your agent's tools as MCP tools

HOW TO USE THE A2A PROTOCOL:
1. The A2A (Agent-to-Agent) protocol enables inter-agent communication
2. Your agent publishes an agent card at /.well-known/agent.json
3. Other agents discover your agent's skills via the card
4. Implement skill handlers in the a2a/ package
5. Use A2A to connect multiple Trabuco agents or third-party agents

HOW TO ENABLE VECTOR RAG (RETRIEVAL-AUGMENTED GENERATION):
1. Generate the project with --vector-store=<flavor>:
   - pgvector: same Postgres datastore (auto-adds SQLDatastore + forces postgresql)
   - qdrant:   standalone gRPC server (best raw performance; runs in Docker)
   - mongodb:  Atlas Vector Search (Atlas-only — community Mongo cannot serve $vectorSearch)
2. Default embedding model is local ONNX (intfloat/e5-small-v2, 384-dim).
   No API key required. To swap: replace spring-ai-starter-model-transformers
   with spring-ai-starter-model-{openai,bedrock,vertex-ai-embedding} and
   update the dimensionality (vector(N) for pgvector V1 migration, Atlas
   index numDimensions for mongodb).
3. Two retrieval surfaces are wired automatically:
   - KnowledgeTools.@Tool askKnowledge — LLM-controlled lookups
   - RetrievalAugmentationAdvisor on PrimaryAgent's ChatClient —
     ambient RAG: every prompt triggers a similarity search,
     top-K (default 4) prepends the prompt as context.
4. Ingest documents via DocumentIngestionService.ingest(text, metadata)
   programmatically, or POST /ingest (gated by @RequireScope("partner")
   AND the agent.ingest.enabled property — default false, so the
   endpoint does not register until the operator opts in). Per-document
   bounds: 1 MB text, 32 metadata keys; rate-limited per caller; audit
   logged at INFO. Per-tenant quotas + content scanning are
   operator-policy decisions — wire them into DocumentIngestionService.
5. PGVector + Postgres: an integration test (VectorRagIntegrationTest)
   boots a real pgvector/pgvector:pg16 Testcontainer and asserts the
   ingest → similarity-search round-trip.
6. MongoDB Atlas: the vector index MUST be created out-of-band in
   Atlas (UI/CLI/API) — Spring AI cannot create it.
   Full guide with index DDL: docs/vector-rag.md.
7. The retriever pair (Vector when wired, Keyword as fallback) is
   wired via @Bean methods in KnowledgeBeansConfiguration with
   @Primary on the vector path. NOT via @Component +
   @ConditionalOnBean — that pattern is unreliable on user-scanned
   components because Spring evaluates the conditional before the
   vector-store auto-config registers VectorStore.

WHEN TO USE LLM vs DETERMINISTIC CODE:
- Use LLM: natural language understanding, classification, summarization, creative content
- Use deterministic code: CRUD operations, calculations, validations, data transformations
- Token cost guidance: each LLM call costs tokens — avoid LLM for tasks that can be coded
- Use @Tool methods to let the LLM call deterministic code when needed

SECURITY PIPELINE:
auth → scope → rate limit → guardrail → agent → output guardrail
1. Authentication: AIAgent ships TWO coexisting paths.
   - OIDC JWT validation via AgentSecurityConfig (dual chain — gated
     on trabuco.auth.enabled property). Validates tokens from any
     RFC-conformant OIDC issuer (Keycloak / Auth0 / Okta / Cognito).
     RFC 7807 ProblemDetail for 401/403.
   - Legacy ApiKeyAuthFilter (gated on app.aiagent.api-key.enabled,
     default true). Tier-based API keys feed CallerContext for the
     existing @RequireScope annotation.
   Both run side-by-side; flip either property to migrate. Default
   state matches pre-1.11 AIAgent behavior (API-key on, JWT dormant).
2. Scope: @PreAuthorize("hasAuthority('SCOPE_*')") and @RequireScope("public"|"partner")
   work interchangeably. Since 1.12, ScopeEnforcer bridges the two so
   either annotation resolves through the same tier ladder under either
   auth mode (API-key, JWT-only, or hybrid).
3. Rate limit: throttle requests per user/API key
4. Input guardrail: validate the prompt against domain rules
5. Agent: process the request with the LLM
6. Output guardrail: filter the response for sensitive content

OBSERVABILITY:
- Token metrics: track input/output tokens per request via Spring AI metrics
- Correlation IDs: trace requests across agent calls and tool invocations
- Circuit breaker: protects against LLM provider outages with fallback responses

TESTING:
- All tests work WITHOUT an API key — LLM calls are mocked in tests
- Integration tests use Testcontainers (no external services needed)
- Test guardrails independently from the agent
- Test tools independently from the LLM

IMMUTABLES PATTERN:
- Use Immutables for all DTOs and value objects (consistent with other Trabuco modules)
- Define abstract classes with @Value.Immutable annotation
- Generated immutable implementations provide builders, equals, hashCode, toString`, task)

		return &mcp.GetPromptResult{
			Description: fmt.Sprintf("AIAgent expert guidance for: %s", task),
			Messages: []mcp.PromptMessage{
				{
					Role:    mcp.RoleUser,
					Content: mcp.TextContent{Type: "text", Text: text},
				},
			},
		}, nil
	})
}
