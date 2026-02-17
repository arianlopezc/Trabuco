package cli

import (
	"fmt"
	"os"

	mcpserver "github.com/arianlopezc/Trabuco/internal/mcp"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP (Model Context Protocol) server for AI agent integration",
	Long: `Start an MCP server over stdio that exposes Trabuco functionality as
structured tools for AI coding agents (Claude Code, Cursor, etc.).

The server communicates via JSON-RPC over stdin/stdout. Configure your
AI agent to run "trabuco mcp" as an MCP server:

  {
    "mcpServers": {
      "trabuco": {
        "command": "trabuco",
        "args": ["mcp"]
      }
    }
  }

Available tools:
  init_project    Generate a new Java project
  add_module      Add a module to an existing project
  run_doctor      Run health checks on a project
  get_project_info Read project metadata
  list_modules    List available modules
  check_docker    Check Docker status
  get_version     Get Trabuco version
  scan_project    Analyze a legacy project (no AI)
  migrate_project AI-powered migration (long-running)
  auth_status     Check configured AI providers
  list_providers  List supported providers with pricing`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := mcpserver.Start(Version); err != nil {
			fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
			os.Exit(1)
		}
	},
}
