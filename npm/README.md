# trabuco-mcp

MCP (Model Context Protocol) server for [Trabuco](https://github.com/arianlopezc/Trabuco) — generate and manage production-ready Java multi-module Maven projects via AI coding agents.

## Quick Setup

### Claude Code

```bash
claude mcp add trabuco -- npx -y trabuco-mcp
```

Or add to `.mcp.json` in your project root:

```json
{
  "mcpServers": {
    "trabuco": {
      "command": "npx",
      "args": ["-y", "trabuco-mcp"]
    }
  }
}
```

### Cursor

Add to `.cursor/mcp.json`:

```json
{
  "mcpServers": {
    "trabuco": {
      "command": "npx",
      "args": ["-y", "trabuco-mcp"]
    }
  }
}
```

### VS Code / GitHub Copilot

Add to `.vscode/mcp.json`:

```json
{
  "servers": {
    "trabuco": {
      "command": "npx",
      "args": ["-y", "trabuco-mcp"]
    }
  }
}
```

### Windsurf

Add to `~/.codeium/windsurf/mcp_config.json`:

```json
{
  "mcpServers": {
    "trabuco": {
      "command": "npx",
      "args": ["-y", "trabuco-mcp"]
    }
  }
}
```

## Available Tools

| Tool | Description |
|------|-------------|
| `init_project` | Generate a new Java project with specified modules, database, and options |
| `add_module` | Add a module to an existing Trabuco project (with dry-run support) |
| `run_doctor` | Run health checks on a project and optionally auto-fix issues |
| `get_project_info` | Read project metadata from `.trabuco.json` or inferred from POM |
| `list_modules` | List all available modules with descriptions and dependency info |
| `check_docker` | Check if Docker is installed and running |
| `get_version` | Get the Trabuco CLI version |
| `scan_project` | Analyze a legacy Java project's structure and dependencies |
| `migrate_project` | Full AI-powered migration of a legacy project |
| `auth_status` | Check which AI providers have credentials configured |
| `list_providers` | List supported AI providers with pricing and model info |

## How It Works

This package wraps the [Trabuco CLI](https://github.com/arianlopezc/Trabuco). On install, it downloads the correct Go binary for your platform. The `trabuco-mcp` command starts the MCP server over stdio — your AI agent handles the rest.

## Requirements

- **Node.js 16+** (for npx/npm)
- **Java 17+** (for generated projects)
- **Maven 3.8+** (for generated projects)
- **Docker** (for Testcontainers and local development)

## License

MIT
