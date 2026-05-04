package mcp

import (
	"context"
	"fmt"

	"github.com/arianlopezc/Trabuco/internal/addgen"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// registerAddCommandTools registers the eight `add_*` tools that
// mirror `trabuco add <type>` CLI subcommands. They are addition-only:
// every tool produces NEW files in deterministic locations and refuses
// to clobber existing ones. Edits and deletes stay with the agent.
//
// Both the CLI commands and these MCP tools call into internal/addgen,
// so the two surfaces are guaranteed to produce byte-identical output.
func registerAddCommandTools(s *server.MCPServer) {
	registerAddMigration(s)
	registerAddTest(s)
	registerAddEntity(s)
	registerAddService(s)
	registerAddJob(s)
	registerAddEndpoint(s)
	registerAddStreamingEndpoint(s)
	registerAddEvent(s)
}

// loadAddCtx is the shared boilerplate for every add tool: resolve the
// path, load the .trabuco.json metadata, honor the dry_run flag.
// Returns a tool error if any step fails so handlers can early-return.
func loadAddCtx(path string, dryRun bool) (*addgen.Context, *mcp.CallToolResult) {
	absPath, err := resolvePath(path)
	if err != nil {
		return nil, toolError(fmt.Sprintf("Failed to resolve path: %v", err))
	}
	ctx, err := addgen.LoadContext(absPath)
	if err != nil {
		return nil, toolError(err.Error())
	}
	ctx.DryRun = dryRun
	return ctx, nil
}

// addResultJSON formats a Result for MCP, mirroring the human/JSON CLI
// output. dry_run is captured so callers can render UI hints.
func addResultJSON(result *addgen.Result, dryRun bool) (*mcp.CallToolResult, error) {
	out := map[string]any{
		"status":  "success",
		"dry_run": dryRun,
		"created": result.Created,
	}
	if len(result.NextSteps) > 0 {
		out["next_steps"] = result.NextSteps
	}
	if len(result.Notes) > 0 {
		out["notes"] = result.Notes
	}
	return toolJSON(out)
}

// --- migration ---

func registerAddMigration(s *server.MCPServer) {
	tool := mcp.NewTool("add_migration",
		mcp.WithDescription(
			"Generate an empty Flyway migration under SQLDatastore/.../db/migration/V{N}__{description}.sql. "+
				"V{N} is auto-picked. The body is a TODO header — agent fills the DDL. "+
				"Refuses to clobber existing files. Mirrors `trabuco add migration` CLI.",
		),
		mcp.WithString("path", mcp.Description("Project root path"), mcp.Required()),
		mcp.WithString("description", mcp.Description("Migration description, snake_cased into the filename"), mcp.Required()),
		mcp.WithString("module", mcp.Description("Always SQLDatastore (default)")),
		mcp.WithBoolean("dry_run", mcp.Description("Preview without writing to disk")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx, errRes := loadAddCtx(req.GetString("path", ""), req.GetBool("dry_run", false))
		if errRes != nil {
			return errRes, nil
		}
		result, err := addgen.GenerateMigration(ctx, addgen.MigrationOpts{
			Description: req.GetString("description", ""),
			Module:      req.GetString("module", ""),
		})
		if err != nil {
			return toolError(err.Error()), nil
		}
		return addResultJSON(result, ctx.DryRun)
	})
}

// --- test ---

func registerAddTest(s *server.MCPServer) {
	tool := mcp.NewTool("add_test",
		mcp.WithDescription(
			"Generate a JUnit 5 test class skeleton (unit, integration, or repository). "+
				"Repository tests wire the right Testcontainer (Postgres/MySQL/Mongo). "+
				"Mirrors `trabuco add test` CLI.",
		),
		mcp.WithString("path", mcp.Description("Project root path"), mcp.Required()),
		mcp.WithString("target", mcp.Description("Class under test, e.g. OrderService"), mcp.Required()),
		mcp.WithString("module", mcp.Description("Module the test belongs to"), mcp.Required()),
		mcp.WithString("type", mcp.Description("unit | integration | repository (default unit)")),
		mcp.WithString("subpackage", mcp.Description("Override the auto-inferred subpackage")),
		mcp.WithBoolean("dry_run", mcp.Description("Preview without writing to disk")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx, errRes := loadAddCtx(req.GetString("path", ""), req.GetBool("dry_run", false))
		if errRes != nil {
			return errRes, nil
		}
		result, err := addgen.GenerateTest(ctx, addgen.TestOpts{
			Target:     req.GetString("target", ""),
			Type:       req.GetString("type", ""),
			Module:     req.GetString("module", ""),
			Subpackage: req.GetString("subpackage", ""),
		})
		if err != nil {
			return toolError(err.Error()), nil
		}
		return addResultJSON(result, ctx.DryRun)
	})
}

// --- entity ---

func registerAddEntity(s *server.MCPServer) {
	tool := mcp.NewTool("add_entity",
		mcp.WithDescription(
			"Generate the full entity bundle: Immutables interface + JDBC record + repository + Flyway migration "+
				"(SQL projects), or Immutables interface + Mongo document + MongoRepository (NoSQL projects). "+
				"Auto-emits enum stubs for any enum:Name fields. "+
				"Field syntax: name:type[?] where type is one of "+
				"string|text|integer|long|decimal|boolean|instant|localdate|uuid|json|bytes|enum:EnumName. "+
				"Trailing ? marks the field nullable. Mirrors `trabuco add entity` CLI.",
		),
		mcp.WithString("path", mcp.Description("Project root path"), mcp.Required()),
		mcp.WithString("name", mcp.Description("PascalCase entity name, e.g. Order"), mcp.Required()),
		mcp.WithString("fields", mcp.Description(`Comma-separated field spec, e.g. "customerId:string,total:decimal,placedAt:instant,notes:text?"`), mcp.Required()),
		mcp.WithString("module", mcp.Description("Force SQLDatastore or NoSQLDatastore (default: auto-detect)")),
		mcp.WithString("table_name", mcp.Description("Override the auto-derived plural snake_case table/collection name")),
		mcp.WithBoolean("dry_run", mcp.Description("Preview without writing to disk")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx, errRes := loadAddCtx(req.GetString("path", ""), req.GetBool("dry_run", false))
		if errRes != nil {
			return errRes, nil
		}
		result, err := addgen.GenerateEntity(ctx, addgen.EntityOpts{
			Name:      req.GetString("name", ""),
			Fields:    req.GetString("fields", ""),
			Module:    req.GetString("module", ""),
			TableName: req.GetString("table_name", ""),
		})
		if err != nil {
			return toolError(err.Error()), nil
		}
		return addResultJSON(result, ctx.DryRun)
	})
}

// --- service ---

func registerAddService(s *server.MCPServer) {
	tool := mcp.NewTool("add_service",
		mcp.WithDescription(
			"Generate Shared/.../service/{Name}.java — a Spring @Service with constructor injection. "+
				"When `entity` is set, wires in the matching repository (SQL or Mongo). "+
				"Mirrors `trabuco add service` CLI.",
		),
		mcp.WithString("path", mcp.Description("Project root path"), mcp.Required()),
		mcp.WithString("name", mcp.Description("PascalCase service name, e.g. OrderService"), mcp.Required()),
		mcp.WithString("entity", mcp.Description("Optional entity name; when set, the constructor injects {Entity}Repository")),
		mcp.WithBoolean("dry_run", mcp.Description("Preview without writing to disk")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx, errRes := loadAddCtx(req.GetString("path", ""), req.GetBool("dry_run", false))
		if errRes != nil {
			return errRes, nil
		}
		result, err := addgen.GenerateService(ctx, addgen.ServiceOpts{
			Name:   req.GetString("name", ""),
			Entity: req.GetString("entity", ""),
		})
		if err != nil {
			return toolError(err.Error()), nil
		}
		return addResultJSON(result, ctx.DryRun)
	})
}

// --- job ---

func registerAddJob(s *server.MCPServer) {
	tool := mcp.NewTool("add_job",
		mcp.WithDescription(
			"Generate the three-file JobRunr job bundle: JobRequest record (Model), no-op base handler (Model), "+
				"and concrete @Component handler (Worker). Recurring schedule registration is left to the agent. "+
				"Mirrors `trabuco add job` CLI.",
		),
		mcp.WithString("path", mcp.Description("Project root path"), mcp.Required()),
		mcp.WithString("name", mcp.Description("PascalCase job name (verb-noun), e.g. ProcessShipment"), mcp.Required()),
		mcp.WithString("payload", mcp.Description(`JobRequest payload fields, e.g. "orderId:string,priority:integer"`), mcp.Required()),
		mcp.WithBoolean("dry_run", mcp.Description("Preview without writing to disk")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx, errRes := loadAddCtx(req.GetString("path", ""), req.GetBool("dry_run", false))
		if errRes != nil {
			return errRes, nil
		}
		result, err := addgen.GenerateJob(ctx, addgen.JobOpts{
			Name:    req.GetString("name", ""),
			Payload: req.GetString("payload", ""),
		})
		if err != nil {
			return toolError(err.Error()), nil
		}
		return addResultJSON(result, ctx.DryRun)
	})
}

// --- endpoint ---

func registerAddEndpoint(s *server.MCPServer) {
	tool := mcp.NewTool("add_endpoint",
		mcp.WithDescription(
			"Generate API/.../controller/{Name}Controller.java. type=plain (default) gives an empty controller; "+
				"type=crud gives five method stubs (POST, GET-by-id, GET-list, PUT, DELETE). "+
				"URL path auto-derived as /api/{plural} unless overridden. "+
				"Mirrors `trabuco add endpoint` CLI.",
		),
		mcp.WithString("path", mcp.Description("Project root path"), mcp.Required()),
		mcp.WithString("name", mcp.Description("PascalCase resource name, e.g. Order"), mcp.Required()),
		mcp.WithString("type", mcp.Description("plain | crud (default plain)")),
		mcp.WithString("url_path", mcp.Description("Override URL path, e.g. /healthz")),
		mcp.WithBoolean("dry_run", mcp.Description("Preview without writing to disk")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx, errRes := loadAddCtx(req.GetString("path", ""), req.GetBool("dry_run", false))
		if errRes != nil {
			return errRes, nil
		}
		result, err := addgen.GenerateEndpoint(ctx, addgen.EndpointOpts{
			Name: req.GetString("name", ""),
			Type: req.GetString("type", ""),
			Path: req.GetString("url_path", ""),
		})
		if err != nil {
			return toolError(err.Error()), nil
		}
		return addResultJSON(result, ctx.DryRun)
	})
}

// --- streaming endpoint ---

func registerAddStreamingEndpoint(s *server.MCPServer) {
	tool := mcp.NewTool("add_streaming_endpoint",
		mcp.WithDescription(
			"Generate AIAgent/.../protocol/{Name}StreamController.java — an SSE controller skeleton "+
				"suitable for LLM token streaming. Replace the heartbeat-only loop with real streaming. "+
				"Mirrors `trabuco add streaming-endpoint` CLI. AIAgent module required.",
		),
		mcp.WithString("path", mcp.Description("Project root path"), mcp.Required()),
		mcp.WithString("name", mcp.Description("PascalCase controller stem, e.g. Conversation"), mcp.Required()),
		mcp.WithString("url_path", mcp.Description("Override URL path (default /api/agent/stream/{name})")),
		mcp.WithBoolean("dry_run", mcp.Description("Preview without writing to disk")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx, errRes := loadAddCtx(req.GetString("path", ""), req.GetBool("dry_run", false))
		if errRes != nil {
			return errRes, nil
		}
		result, err := addgen.GenerateStreamingEndpoint(ctx, addgen.StreamingEndpointOpts{
			Name: req.GetString("name", ""),
			Path: req.GetString("url_path", ""),
		})
		if err != nil {
			return toolError(err.Error()), nil
		}
		return addResultJSON(result, ctx.DryRun)
	})
}

// --- event ---

func registerAddEvent(s *server.MCPServer) {
	tool := mcp.NewTool("add_event",
		mcp.WithDescription(
			"Generate Model/.../events/{Name}.java — a Java record for a domain event. "+
				"The CLI does NOT add the event to any sealed permits clause or listener switch — "+
				"those edits stay with the agent. Mirrors `trabuco add event` CLI.",
		),
		mcp.WithString("path", mcp.Description("Project root path"), mcp.Required()),
		mcp.WithString("name", mcp.Description("PascalCase event name, e.g. OrderShipped"), mcp.Required()),
		mcp.WithString("fields", mcp.Description(`Comma-separated field spec, e.g. "orderId:string,shippedAt:instant"`), mcp.Required()),
		mcp.WithBoolean("dry_run", mcp.Description("Preview without writing to disk")),
	)
	s.AddTool(tool, func(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		ctx, errRes := loadAddCtx(req.GetString("path", ""), req.GetBool("dry_run", false))
		if errRes != nil {
			return errRes, nil
		}
		result, err := addgen.GenerateEvent(ctx, addgen.EventOpts{
			Name:   req.GetString("name", ""),
			Fields: req.GetString("fields", ""),
		})
		if err != nil {
			return toolError(err.Error()), nil
		}
		return addResultJSON(result, ctx.DryRun)
	})
}
