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
		server.WithInstructions(`Trabuco generates production-ready Java multi-module Maven projects with Spring Boot.

WORKFLOW:
1. Call suggest_architecture with the user's requirements
2. Read the module catalog — match user needs to module use cases and boundaries
3. Check warnings for ambiguous terms and not_covered for unsupported requirements
4. YOU decide which modules, database, and message broker to use based on the catalog
5. Call init_project with your chosen configuration
6. The response includes next_steps and key_files to guide post-generation work
7. Read AGENTS.md in the generated project for coding patterns and conventions
8. Use add_module to incrementally add capabilities as needs evolve

WHAT TRABUCO GENERATES:
- Multi-module Maven project (Model, Datastore, Shared, API, Worker, EventConsumer)
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
