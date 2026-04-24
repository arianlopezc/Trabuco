package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	syncpkg "github.com/arianlopezc/Trabuco/internal/sync"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerSyncProject exposes `trabuco sync` as an MCP tool. Agents call this
// to detect and apply drift between a project's AI-tooling files and what
// the installed CLI would generate today. The tool is safe to call with
// apply=false (the default) at any time; it never writes files on a dry-run.
func registerSyncProject(s *server.MCPServer, version string) {
	tool := mcp.NewTool("sync_project",
		mcp.WithDescription(
			"Identify AI-tooling files the current Trabuco CLI would generate for this project but that are missing on disk, and optionally create them. "+
				"WHEN TO USE: An existing Trabuco project was generated with an older CLI, and the user wants the skills, subagents, prompts, hooks, and review scaffolding to match the current CLI. "+
				"Also use after `trabuco add <module>` to pick up new module-specific AI files (e.g., add-tool.md when AIAgent is added). "+
				"ADDITIVE ONLY: Existing files are never modified or deleted. To refresh a file like CLAUDE.md, the user must delete it before syncing. "+
				"JURISDICTION: Only AI-tooling files (.ai/, .claude/, .cursor/, .codex/, .agents/, .github/instructions/, .github/skills/, review-checks.sh, CLAUDE.md, AGENTS.md, and a few specific .github files) are in scope. "+
				"Java source, POMs, migrations, application.yml, docker-compose.yml, CI workflows (except copilot-setup-steps.yml), and all other business/infrastructure files are NEVER touched. "+
				"USAGE: Call with apply=false first to preview the plan; if the user confirms, call again with apply=true. The tool is idempotent — running twice with apply=true is equivalent to running once.",
		),
		mcp.WithString("project_path",
			mcp.Description("Absolute path to the Trabuco project directory. Must contain .trabuco.json."),
			mcp.Required(),
		),
		mcp.WithBoolean("apply",
			mcp.Description("If true, create the missing files. Defaults to false (dry-run)."),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := req.Params.Arguments.(map[string]any)
		if !ok {
			return mcp.NewToolResultError("invalid arguments"), nil
		}

		projectPath, _ := args["project_path"].(string)
		if projectPath == "" {
			return mcp.NewToolResultError("project_path is required"), nil
		}
		apply, _ := args["apply"].(bool)

		plan, err := syncpkg.Run(projectPath, version, apply)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("sync failed: %v", err)), nil
		}

		data, err := json.MarshalIndent(plan, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marshal plan: %v", err)), nil
		}
		return mcp.NewToolResultText(string(data)), nil
	})
}
