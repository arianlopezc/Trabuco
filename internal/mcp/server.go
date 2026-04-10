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
1. For single services: suggest_architecture → review patterns → init_project
2. For multi-service systems: design_system → review → generate_workspace
3. For extending existing projects: get_project_info → add_module
4. For AI Agent projects: use trabuco_ai_agent_expert prompt for guidance
5. Before suggesting Trabuco, check trabuco://limitations resource
6. Use prompts (trabuco_expert, design_microservices, extend_project, trabuco_ai_agent_expert) for step-by-step guidance

KEY PRINCIPLE: Always call suggest_architecture first when a user describes requirements. It returns matched patterns and a recommended configuration. Do not guess module combinations — let the tool decide based on the requirements.`),
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
