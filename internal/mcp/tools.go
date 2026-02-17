package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/auth"
	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/doctor"
	"github.com/arianlopezc/Trabuco/internal/generator"
	"github.com/arianlopezc/Trabuco/internal/migrate"
	"github.com/arianlopezc/Trabuco/internal/utils"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var (
	projectNameRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)
	groupIDRegex     = regexp.MustCompile(`^[a-z][a-z0-9]*(\.[a-z][a-z0-9]*)+$`)
)

func registerAllTools(s *server.MCPServer, version string) {
	registerInitProject(s, version)
	registerAddModule(s, version)
	registerRunDoctor(s, version)
	registerGetProjectInfo(s)
	registerListModules(s)
	registerCheckDocker(s)
	registerGetVersion(s, version)
	registerScanProject(s)
	registerMigrateProject(s, version)
	registerAuthStatus(s)
	registerListProviders(s)
}

// ---------- Project Management Tools ----------

func registerInitProject(s *server.MCPServer, version string) {
	tool := mcp.NewTool("init_project",
		mcp.WithDescription("Generate a new production-ready Java multi-module Maven project"),
		mcp.WithString("name",
			mcp.Description("Project name (lowercase, hyphens allowed, e.g. 'my-platform')"),
			mcp.Required(),
		),
		mcp.WithString("group_id",
			mcp.Description("Maven group ID (e.g. 'com.company.project')"),
			mcp.Required(),
		),
		mcp.WithString("modules",
			mcp.Description("Comma-separated modules: Model, SQLDatastore, NoSQLDatastore, Shared, API, Worker, Events, EventConsumer, MCP, Jobs"),
			mcp.Required(),
		),
		mcp.WithString("database",
			mcp.Description("SQL database type: postgresql, mysql, generic (required if SQLDatastore selected)"),
		),
		mcp.WithString("nosql_database",
			mcp.Description("NoSQL database type: mongodb, redis (required if NoSQLDatastore selected)"),
		),
		mcp.WithString("message_broker",
			mcp.Description("Message broker: kafka, rabbitmq, sqs, pubsub (required if EventConsumer selected)"),
		),
		mcp.WithString("java_version",
			mcp.Description("Java version: 17, 21, or 25 (default: 21)"),
		),
		mcp.WithString("ai_agents",
			mcp.Description("Comma-separated AI agent configs to include: claude, cursor, windsurf, copilot"),
		),
		mcp.WithString("output_dir",
			mcp.Description("Directory to create the project in (default: current directory)"),
		),
		mcp.WithBoolean("skip_build",
			mcp.Description("Skip running Maven build after generation (default: true)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		name := req.GetString("name", "")
		groupID := req.GetString("group_id", "")
		modulesStr := req.GetString("modules", "")
		database := req.GetString("database", "")
		nosqlDatabase := req.GetString("nosql_database", "")
		messageBroker := req.GetString("message_broker", "")
		javaVersion := req.GetString("java_version", "21")
		aiAgentsStr := req.GetString("ai_agents", "")
		outputDir := req.GetString("output_dir", "")
		skipBuild := req.GetBool("skip_build", true)

		// Validate name
		if !projectNameRegex.MatchString(name) {
			return toolError(fmt.Sprintf("Invalid project name '%s'. Must be lowercase, alphanumeric, hyphens allowed (not at start/end).", name)), nil
		}

		// Validate group ID
		if !groupIDRegex.MatchString(groupID) {
			return toolError(fmt.Sprintf("Invalid group ID '%s'. Must be valid Java package format (e.g., com.company.project).", groupID)), nil
		}

		// Validate Java version
		if javaVersion != "17" && javaVersion != "21" && javaVersion != "25" {
			return toolError(fmt.Sprintf("Invalid Java version '%s'. Must be 17, 21, or 25.", javaVersion)), nil
		}

		// Parse modules
		modules := strings.Split(modulesStr, ",")
		for i := range modules {
			modules[i] = strings.TrimSpace(modules[i])
		}

		// Validate module selection
		if validationErr := config.ValidateModuleSelection(modules); validationErr != "" {
			return toolError(fmt.Sprintf("Invalid module selection: %s", validationErr)), nil
		}

		// Resolve dependencies
		resolvedModules := config.ResolveDependencies(modules)

		// Parse AI agents
		var aiAgents []string
		if aiAgentsStr != "" {
			validAgents := make(map[string]bool)
			for _, id := range config.GetAIAgentIDs() {
				validAgents[id] = true
			}
			for _, agent := range strings.Split(aiAgentsStr, ",") {
				agent = strings.TrimSpace(strings.ToLower(agent))
				if agent == "" {
					continue
				}
				if !validAgents[agent] {
					return toolError(fmt.Sprintf("Invalid AI agent '%s'. Valid: %s", agent, strings.Join(config.GetAIAgentIDs(), ", "))), nil
				}
				aiAgents = append(aiAgents, agent)
			}
		}

		cfg := &config.ProjectConfig{
			ProjectName:   name,
			GroupID:       groupID,
			ArtifactID:    name,
			JavaVersion:   javaVersion,
			Modules:       resolvedModules,
			Database:      database,
			NoSQLDatabase: nosqlDatabase,
			MessageBroker: messageBroker,
			AIAgents:      aiAgents,
		}

		// Change to output dir if specified
		if outputDir != "" {
			absDir, err := filepath.Abs(outputDir)
			if err != nil {
				return toolError(fmt.Sprintf("Invalid output directory: %v", err)), nil
			}
			// Save current dir so generator creates project there
			origDir, _ := filepath.Abs(".")
			if err := changeDir(absDir); err != nil {
				return toolError(fmt.Sprintf("Cannot access output directory: %v", err)), nil
			}
			defer changeDir(origDir)
		}

		gen, err := generator.NewWithVersion(cfg, version)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to create generator: %v", err)), nil
		}

		if err := gen.Generate(); err != nil {
			return toolError(fmt.Sprintf("Failed to generate project: %v", err)), nil
		}

		var warnings []string
		if cfg.ShowRedisWorkerWarning() {
			warnings = append(warnings, "Redis support is deprecated in JobRunr 8+. Worker uses PostgreSQL for job storage.")
		}

		projectPath := name
		if outputDir != "" {
			projectPath = filepath.Join(outputDir, name)
		}
		absPath, _ := filepath.Abs(projectPath)

		// Run Maven build if not skipped
		buildStatus := "skipped"
		if !skipBuild {
			if err := utils.RunMavenBuild(absPath); err != nil {
				warnings = append(warnings, fmt.Sprintf("Maven build failed: %v", err))
				buildStatus = "failed"
			} else {
				buildStatus = "success"
			}
		}

		return toolJSON(map[string]any{
			"status":       "success",
			"path":         absPath,
			"modules":      resolvedModules,
			"database":     database,
			"java_version": javaVersion,
			"build":        buildStatus,
			"warnings":     warnings,
		})
	})
}

func registerAddModule(s *server.MCPServer, version string) {
	tool := mcp.NewTool("add_module",
		mcp.WithDescription("Add a module to an existing Trabuco project"),
		mcp.WithString("path",
			mcp.Description("Path to the Trabuco project root"),
			mcp.Required(),
		),
		mcp.WithString("module",
			mcp.Description("Module to add: SQLDatastore, NoSQLDatastore, Shared, API, Worker, EventConsumer, MCP"),
			mcp.Required(),
		),
		mcp.WithString("database",
			mcp.Description("SQL database type: postgresql, mysql, generic (for SQLDatastore)"),
		),
		mcp.WithString("nosql_database",
			mcp.Description("NoSQL database type: mongodb, redis (for NoSQLDatastore)"),
		),
		mcp.WithString("message_broker",
			mcp.Description("Message broker: kafka, rabbitmq, sqs, pubsub (for EventConsumer)"),
		),
		mcp.WithBoolean("dry_run",
			mcp.Description("Preview changes without applying them"),
		),
		mcp.WithBoolean("skip_build",
			mcp.Description("Skip Maven build after adding module (default: true)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := req.GetString("path", "")
		module := req.GetString("module", "")
		database := req.GetString("database", "")
		nosqlDatabase := req.GetString("nosql_database", "")
		messageBroker := req.GetString("message_broker", "")
		dryRun := req.GetBool("dry_run", false)
		skipBuild := req.GetBool("skip_build", true)

		absPath, err := resolvePath(path)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to resolve path: %v", err)), nil
		}

		meta, err := doctor.GetProjectMetadata(absPath)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to read project metadata: %v", err)), nil
		}

		adder := generator.NewModuleAdder(absPath, meta, version, true)

		if dryRun {
			result := adder.DryRun(module)
			return toolJSON(map[string]any{
				"status":         "dry_run",
				"module":         result.Module,
				"dependencies":   result.Dependencies,
				"files_created":  result.FilesCreated,
				"files_modified": result.FilesModified,
			})
		}

		if err := adder.Add(module, database, nosqlDatabase, messageBroker); err != nil {
			return toolError(fmt.Sprintf("Failed to add module: %v", err)), nil
		}

		// Run Maven build if not skipped
		buildStatus := "skipped"
		if !skipBuild {
			if err := utils.RunMavenBuild(absPath); err != nil {
				buildStatus = "failed"
			} else {
				buildStatus = "success"
			}
		}

		// Gather info about what was done
		dryResult := adder.DryRun(module) // safe to call for info even after add
		return toolJSON(map[string]any{
			"status":         "success",
			"module":         module,
			"dependencies":   dryResult.Dependencies,
			"files_created":  dryResult.FilesCreated,
			"files_modified": dryResult.FilesModified,
			"build":          buildStatus,
		})
	})
}

func registerRunDoctor(s *server.MCPServer, version string) {
	tool := mcp.NewTool("run_doctor",
		mcp.WithDescription("Run health checks on a Trabuco project to detect issues"),
		mcp.WithString("path",
			mcp.Description("Path to the Trabuco project root"),
			mcp.Required(),
		),
		mcp.WithBoolean("fix",
			mcp.Description("Attempt to auto-fix issues"),
		),
		mcp.WithString("category",
			mcp.Description("Run specific check category: structure, metadata, consistency"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := req.GetString("path", "")
		fix := req.GetBool("fix", false)
		category := req.GetString("category", "")

		absPath, err := resolvePath(path)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to resolve path: %v", err)), nil
		}

		doc := doctor.New(absPath, version)

		if fix {
			result, fixes, err := doc.RunAndFix()
			if err != nil {
				return toolError(fmt.Sprintf("Doctor failed: %v", err)), nil
			}
			jsonBytes, err := result.ToJSON()
			if err != nil {
				return toolError(fmt.Sprintf("Failed to serialize result: %v", err)), nil
			}
			// Include fix results
			fixSummary := make([]map[string]any, len(fixes))
			for i, f := range fixes {
				fixSummary[i] = map[string]any{
					"check":   f.CheckID,
					"name":    f.Name,
					"success": f.Success,
					"error":   f.Error,
				}
			}
			// Combine doctor JSON with fixes
			var raw map[string]any
			if err := json.Unmarshal(jsonBytes, &raw); err != nil {
				return mcp.NewToolResultText(string(jsonBytes)), nil
			}
			raw["fixes"] = fixSummary
			return toolJSON(raw)
		}

		if category != "" {
			result, err := doc.RunCategory(category)
			if err != nil {
				return toolError(fmt.Sprintf("Doctor failed: %v", err)), nil
			}
			jsonBytes, err := result.ToJSON()
			if err != nil {
				return toolError(fmt.Sprintf("Failed to serialize result: %v", err)), nil
			}
			return mcp.NewToolResultText(string(jsonBytes)), nil
		}

		result, err := doc.Run()
		if err != nil {
			return toolError(fmt.Sprintf("Doctor failed: %v", err)), nil
		}
		jsonBytes, err := result.ToJSON()
		if err != nil {
			return toolError(fmt.Sprintf("Failed to serialize result: %v", err)), nil
		}
		return mcp.NewToolResultText(string(jsonBytes)), nil
	})
}

func registerGetProjectInfo(s *server.MCPServer) {
	tool := mcp.NewTool("get_project_info",
		mcp.WithDescription("Read project metadata from a Trabuco project (.trabuco.json or inferred from POM)"),
		mcp.WithString("path",
			mcp.Description("Path to the Trabuco project root"),
			mcp.Required(),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := req.GetString("path", "")

		absPath, err := resolvePath(path)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to resolve path: %v", err)), nil
		}

		// Try .trabuco.json first
		meta, err := config.LoadMetadata(absPath)
		if err != nil {
			// Fall back to POM inference
			meta, err = doctor.GetProjectMetadata(absPath)
			if err != nil {
				return toolError(fmt.Sprintf("Failed to read project info: %v", err)), nil
			}
		}

		return toolJSON(meta)
	})
}

// ---------- Discovery Tools ----------

func registerListModules(s *server.MCPServer) {
	tool := mcp.NewTool("list_modules",
		mcp.WithDescription("List all available Trabuco modules with descriptions and dependency info"),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		type moduleInfo struct {
			Name         string   `json:"name"`
			Description  string   `json:"description"`
			Required     bool     `json:"required"`
			Internal     bool     `json:"internal"`
			Dependencies []string `json:"dependencies"`
		}

		modules := make([]moduleInfo, len(config.ModuleRegistry))
		for i, m := range config.ModuleRegistry {
			modules[i] = moduleInfo{
				Name:         m.Name,
				Description:  m.Description,
				Required:     m.Required,
				Internal:     m.Internal,
				Dependencies: m.Dependencies,
			}
		}

		return toolJSON(modules)
	})
}

func registerCheckDocker(s *server.MCPServer) {
	tool := mcp.NewTool("check_docker",
		mcp.WithDescription("Check if Docker is installed and running (required for project generation and tests)"),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		status := utils.CheckDocker()
		return toolJSON(map[string]any{
			"installed": status.Installed,
			"running":   status.Running,
			"version":   status.Version,
			"error":     status.Error,
		})
	})
}

func registerGetVersion(s *server.MCPServer, version string) {
	tool := mcp.NewTool("get_version",
		mcp.WithDescription("Get the Trabuco CLI version"),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return toolJSON(map[string]string{
			"version": version,
		})
	})
}

// ---------- Migration Tools ----------

func registerScanProject(s *server.MCPServer) {
	tool := mcp.NewTool("scan_project",
		mcp.WithDescription("Analyze a legacy Java project structure and dependencies (fast, no AI required)"),
		mcp.WithString("path",
			mcp.Description("Path to the legacy Java project to scan"),
			mcp.Required(),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path := req.GetString("path", "")

		absPath, err := resolvePath(path)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to resolve path: %v", err)), nil
		}

		scanner := migrate.NewProjectScanner(absPath)
		projectInfo, err := scanner.Scan()
		if err != nil {
			return toolError(fmt.Sprintf("Scan failed: %v", err)), nil
		}

		analyzer := migrate.NewDependencyAnalyzer()
		depReport := analyzer.Analyze(projectInfo.Dependencies)

		return toolJSON(map[string]any{
			"project": map[string]any{
				"name":                projectInfo.Name,
				"project_type":       projectInfo.ProjectType,
				"java_version":       projectInfo.JavaVersion,
				"base_package":       projectInfo.BasePackage,
				"group_id":           projectInfo.GroupID,
				"artifact_id":        projectInfo.ArtifactID,
				"spring_boot_version": projectInfo.SpringBootVersion,
				"is_multi_module":    projectInfo.IsMultiModule,
				"modules":            projectInfo.Modules,
				"database":           projectInfo.Database,
				"message_broker":     projectInfo.MessageBroker,
				"has_scheduled_jobs": projectInfo.HasScheduledJobs,
				"has_event_listeners": projectInfo.HasEventListeners,
				"uses_nosql":         projectInfo.UsesNoSQL,
				"entities_count":     len(projectInfo.Entities),
				"repositories_count": len(projectInfo.Repositories),
				"services_count":     len(projectInfo.Services),
				"controllers_count":  len(projectInfo.Controllers),
			},
			"dependencies": map[string]any{
				"compatible_count":  len(depReport.Compatible),
				"replaceable":      depReport.Replaceable,
				"unsupported":      depReport.Unsupported,
				"has_blockers":     depReport.HasBlockers(),
				"migration_summary": depReport.GetMigrationSummary(),
			},
		})
	})
}

func registerMigrateProject(s *server.MCPServer, version string) {
	tool := mcp.NewTool("migrate_project",
		mcp.WithDescription("Full AI-powered migration of a legacy Java project to Trabuco structure (long-running)"),
		mcp.WithString("source_path",
			mcp.Description("Path to the legacy Java project to migrate"),
			mcp.Required(),
		),
		mcp.WithString("output_path",
			mcp.Description("Where to create the migrated project (default: <source>-trabuco)"),
		),
		mcp.WithString("provider",
			mcp.Description("AI provider: anthropic, openrouter (default: anthropic)"),
		),
		mcp.WithString("api_key",
			mcp.Description("API key for the AI provider (or use trabuco auth login)"),
		),
		mcp.WithString("model",
			mcp.Description("Model to use (default: claude-sonnet-4-5)"),
		),
		mcp.WithBoolean("dry_run",
			mcp.Description("Analyze without generating files"),
		),
		mcp.WithBoolean("include_tests",
			mcp.Description("Migrate test files"),
		),
		mcp.WithBoolean("skip_build",
			mcp.Description("Skip Maven build after migration (default: true)"),
		),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sourcePath := req.GetString("source_path", "")
		outputPath := req.GetString("output_path", "")
		providerName := req.GetString("provider", "anthropic")
		apiKey := req.GetString("api_key", "")
		model := req.GetString("model", "")
		dryRun := req.GetBool("dry_run", false)
		includeTests := req.GetBool("include_tests", false)
		skipBuild := req.GetBool("skip_build", true)

		absSource, err := resolvePath(sourcePath)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to resolve source path: %v", err)), nil
		}

		// Create AI provider
		provider, err := createProvider(providerName, apiKey, model)
		if err != nil {
			return toolError(fmt.Sprintf("Failed to create AI provider: %v", err)), nil
		}

		if outputPath == "" {
			outputPath = absSource + "-trabuco"
		}
		absOutput, err := filepath.Abs(outputPath)
		if err != nil {
			return toolError(fmt.Sprintf("Invalid output path: %v", err)), nil
		}

		migrateCfg := &migrate.Config{
			SourcePath:     absSource,
			OutputPath:     absOutput,
			DryRun:         dryRun,
			Interactive:    false, // MCP is non-interactive
			IncludeTests:   includeTests,
			SkipBuild:      skipBuild,
			TrabucoVersion: version,
		}

		migrator := migrate.NewMigrator(provider, migrateCfg)

		if err := migrator.Run(); err != nil {
			return toolError(fmt.Sprintf("Migration failed: %v", err)), nil
		}

		status := "success"
		if dryRun {
			status = "dry_run"
		}

		return toolJSON(map[string]any{
			"status":      status,
			"output_path": absOutput,
		})
	})
}

// ---------- Authentication Tools ----------

func registerAuthStatus(s *server.MCPServer) {
	tool := mcp.NewTool("auth_status",
		mcp.WithDescription("Check which AI providers have credentials configured"),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		manager, err := auth.NewManager()
		if err != nil {
			return toolError(fmt.Sprintf("Failed to initialize credential manager: %v", err)), nil
		}

		statuses := manager.ListConfigured()

		providerList := make([]map[string]any, len(statuses))
		for i, s := range statuses {
			providerList[i] = map[string]any{
				"provider":     string(s.Provider),
				"name":         s.Info.Name,
				"configured":   s.Configured,
				"is_default":   s.IsDefault,
				"source":       s.Source,
				"model":        s.Model,
				"validated_at": s.ValidatedAt,
			}
		}

		return toolJSON(map[string]any{
			"providers":       providerList,
			"storage_backend": manager.StorageBackend(),
		})
	})
}

func registerListProviders(s *server.MCPServer) {
	tool := mcp.NewTool("list_providers",
		mcp.WithDescription("List supported AI providers with pricing and model information"),
	)

	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		var providers []map[string]any
		for id, info := range auth.SupportedProviders {
			providers = append(providers, map[string]any{
				"id":               string(id),
				"name":             info.Name,
				"env_var":          info.EnvVar,
				"models":           info.Models,
				"input_cost_per_1m":  info.InputCostPer1M,
				"output_cost_per_1m": info.OutputCostPer1M,
				"requires_key":     info.RequiresKey,
			})
		}

		return toolJSON(providers)
	})
}

// ---------- Helpers ----------

// createProvider creates an AI provider from the given parameters,
// falling back to the credential manager if no API key is provided.
func createProvider(providerName, apiKey, model string) (ai.Provider, error) {
	if apiKey != "" {
		cfg := &ai.ProviderConfig{
			APIKey: apiKey,
			Model:  model,
		}
		switch strings.ToLower(providerName) {
		case "openrouter":
			return ai.NewOpenRouterProvider(cfg)
		default:
			return ai.NewAnthropicProvider(cfg)
		}
	}

	// Use credential manager
	manager, err := auth.NewManager()
	if err != nil {
		return nil, fmt.Errorf("no API key provided and credential manager failed: %w", err)
	}

	var preferred auth.Provider
	switch strings.ToLower(providerName) {
	case "anthropic":
		preferred = auth.ProviderAnthropic
	case "openrouter":
		preferred = auth.ProviderOpenRouter
	case "openai":
		preferred = auth.ProviderOpenAI
	}

	cred, err := manager.GetCredentialWithFallback(preferred)
	if err != nil {
		return nil, fmt.Errorf("no API key provided and no stored credentials found. Run 'trabuco auth login' or pass api_key parameter: %w", err)
	}

	cfg := &ai.ProviderConfig{
		APIKey: cred.APIKey,
		Model:  model,
	}
	if cfg.Model == "" && cred.Model != "" {
		cfg.Model = cred.Model
	}

	switch cred.Provider {
	case auth.ProviderOpenRouter:
		return ai.NewOpenRouterProvider(cfg)
	default:
		return ai.NewAnthropicProvider(cfg)
	}
}

// changeDir changes the working directory.
func changeDir(dir string) error {
	return os.Chdir(dir)
}
