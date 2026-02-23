package mcp

import (
	"context"
	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Start creates the MCP server, registers all tools, and runs the stdio transport.
// Internal packages print colored output to os.Stdout. The MCP stdio transport also
// uses stdout for JSON-RPC. To avoid collisions, we save real stdout for MCP and
// redirect os.Stdout to os.Stderr so all internal output goes there instead.
func Start(version string) error {
	// Save real stdout for MCP transport, redirect os.Stdout -> os.Stderr
	realStdout := os.Stdout
	os.Stdout = os.Stderr

	s := server.NewMCPServer(
		"trabuco",
		version,
		server.WithToolCapabilities(false),
		server.WithPromptCapabilities(false),
		server.WithResourceCapabilities(false, false),
		server.WithInstructions(`Trabuco generates production-ready Java multi-module Maven projects with Spring Boot.

WORKFLOW:
1. For single services: suggest_architecture → review patterns and recommended_config → init_project
2. For multi-service systems: design_system → review with user → generate_workspace
3. For extending existing projects: get_project_info → add_module
4. Use MCP Prompts for expert guidance on architecture decisions
5. Use MCP Resources to read module catalog, patterns, and limitations

TOOLS:
- suggest_architecture: Returns matched patterns, recommended config, module catalog, and warnings
- design_system: Decomposes requirements into multiple service definitions (review-only, no generation)
- generate_workspace: Creates multiple services with shared infrastructure (docker-compose)
- init_project: Generate a single Trabuco project
- add_module: Add a module to an existing project
- get_project_info: Read project metadata
- list_modules: List all available modules

PROMPTS (request expert knowledge):
- trabuco_expert: General guidance for any Trabuco task
- design_microservices: Step-by-step microservice decomposition guide
- extend_project: Instructions for extending an existing project

RESOURCES (stable reference data):
- trabuco://modules: Full module catalog with use cases and boundaries
- trabuco://patterns: Pre-built architecture patterns with module combinations
- trabuco://limitations: Complete list of what Trabuco does NOT generate

WHAT TRABUCO GENERATES:
- Multi-module Maven project (Model, Datastore, Shared, API, Worker, EventConsumer, MCP)
- Spring Boot 3.4 with Spring Data JDBC (not JPA), Immutables, Testcontainers
- Docker Compose, CI workflow, AI context files, code quality enforcement
- Working placeholder code that serves as patterns for real implementation

WHAT TRABUCO DOES NOT GENERATE:
- Authentication/authorization (add Spring Security)
- Frontend/UI (backend only)
- GraphQL, gRPC, WebSockets (REST only)
- Kubernetes manifests, Terraform, cloud deployment
- Custom business logic or production database schemas

When the user's requirements include items Trabuco doesn't cover, inform them clearly and suggest what to add manually after generation.`),
	)

	registerAllTools(s, version)
	registerAllPrompts(s)
	registerAllResources(s)

	stdioServer := server.NewStdioServer(s)
	return stdioServer.Listen(context.Background(), os.Stdin, realStdout)
}

// toolError returns an MCP result with isError: true.
func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: msg,
			},
		},
		IsError: true,
	}
}

// toolJSON marshals v as indented JSON and returns it as a text result.
func toolJSON(v any) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultJSON(v)
}

// resolvePath resolves an empty or relative path to an absolute path.
func resolvePath(path string) (string, error) {
	if path == "" {
		return os.Getwd()
	}
	if path[0] == '/' {
		return path, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return wd + "/" + path, nil
}
