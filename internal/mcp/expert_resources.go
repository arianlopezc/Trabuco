package mcp

import (
	"context"
	"encoding/json"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerAllResources(s *server.MCPServer) {
	registerModulesResource(s)
	registerPatternsResource(s)
	registerLimitationsResource(s)
}

func registerModulesResource(s *server.MCPServer) {
	s.AddResource(
		mcp.NewResource(
			"trabuco://modules",
			"Module Catalog",
			mcp.WithResourceDescription("Full Trabuco module catalog with names, descriptions, use cases, boundaries, dependencies, and conflicts"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			type moduleEntry struct {
				Name           string   `json:"name"`
				Description    string   `json:"description"`
				UseCase        string   `json:"use_case"`
				WhenToUse      string   `json:"when_to_use"`
				DoesNotInclude string   `json:"does_not_include"`
				Required       bool     `json:"required"`
				Internal       bool     `json:"internal"`
				Dependencies   []string `json:"dependencies"`
				ConflictsWith  []string `json:"conflicts_with"`
			}

			entries := make([]moduleEntry, len(config.ModuleRegistry))
			for i, m := range config.ModuleRegistry {
				deps := m.Dependencies
				if deps == nil {
					deps = []string{}
				}
				conflicts := m.ConflictsWith
				if conflicts == nil {
					conflicts = []string{}
				}
				entries[i] = moduleEntry{
					Name:           m.Name,
					Description:    m.Description,
					UseCase:        m.UseCase,
					WhenToUse:      m.WhenToUse,
					DoesNotInclude: m.DoesNotInclude,
					Required:       m.Required,
					Internal:       m.Internal,
					Dependencies:   deps,
					ConflictsWith:  conflicts,
				}
			}

			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return nil, err
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "trabuco://modules",
					MIMEType: "application/json",
					Text:     string(data),
				},
			}, nil
		},
	)
}

func registerPatternsResource(s *server.MCPServer) {
	s.AddResource(
		mcp.NewResource(
			"trabuco://patterns",
			"Architecture Patterns",
			mcp.WithResourceDescription("Pre-built architectural patterns with module combinations, recommended infrastructure, and use cases"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			type patternEntry struct {
				Name            string   `json:"name"`
				Description     string   `json:"description"`
				UseCases        []string `json:"use_cases"`
				Modules         []string `json:"modules"`
				RecommendedDB   string   `json:"recommended_database,omitempty"`
				RecommendedNoDB string   `json:"recommended_nosql_database,omitempty"`
				RecommendedBrkr string   `json:"recommended_broker,omitempty"`
				Constraints     []string `json:"constraints,omitempty"`
			}

			entries := make([]patternEntry, len(patternCatalog))
			for i, p := range patternCatalog {
				entries[i] = patternEntry{
					Name:            p.Name,
					Description:     p.Description,
					UseCases:        p.UseCases,
					Modules:         p.Modules,
					RecommendedDB:   p.RecommendedDB,
					RecommendedNoDB: p.RecommendedNoDB,
					RecommendedBrkr: p.RecommendedBrkr,
					Constraints:     p.Constraints,
				}
			}

			data, err := json.MarshalIndent(entries, "", "  ")
			if err != nil {
				return nil, err
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "trabuco://patterns",
					MIMEType: "application/json",
					Text:     string(data),
				},
			}, nil
		},
	)
}

func registerLimitationsResource(s *server.MCPServer) {
	s.AddResource(
		mcp.NewResource(
			"trabuco://limitations",
			"Trabuco Limitations",
			mcp.WithResourceDescription("Complete list of what Trabuco does NOT generate — check this before suggesting Trabuco for a requirement"),
			mcp.WithMIMEType("application/json"),
		),
		func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			type limitation struct {
				Category    string   `json:"category"`
				Description string   `json:"description"`
				Workaround  string   `json:"workaround"`
				Keywords    []string `json:"detection_keywords"`
			}

			limitations := []limitation{
				{
					Category:    "Authentication & Authorization",
					Description: "No login, OAuth, JWT, session management, RBAC, or permission systems",
					Workaround:  "Add Spring Security manually after generation",
					Keywords:    []string{"auth", "login", "oauth", "jwt", "session", "permission", "rbac", "role"},
				},
				{
					Category:    "Frontend & UI",
					Description: "No frontend code, HTML templates, React, Angular, Vue, or any UI components",
					Workaround:  "Use a separate frontend framework; Trabuco generates backend only",
					Keywords:    []string{"frontend", "react", "angular", "vue", "ui", "html", "css", "template"},
				},
				{
					Category:    "GraphQL",
					Description: "No GraphQL schema or resolvers — REST APIs only",
					Workaround:  "Add graphql-java or Netflix DGS manually",
					Keywords:    []string{"graphql"},
				},
				{
					Category:    "gRPC & Protobuf",
					Description: "No gRPC services or Protocol Buffer definitions",
					Workaround:  "Add grpc-spring-boot-starter manually",
					Keywords:    []string{"grpc", "protobuf"},
				},
				{
					Category:    "WebSockets & Server-Sent Events",
					Description: "No WebSocket endpoints or SSE streaming",
					Workaround:  "Add Spring WebSocket or Spring WebFlux manually",
					Keywords:    []string{"websocket", "socket", "real-time", "sse", "server-sent"},
				},
				{
					Category:    "Kubernetes & Container Orchestration",
					Description: "No K8s manifests, Helm charts, or container orchestration configs. Docker Compose is for local dev only",
					Workaround:  "Add Kubernetes manifests, Helm charts, or use a deployment platform",
					Keywords:    []string{"kubernetes", "k8s", "helm", "docker deploy", "container orchestration"},
				},
				{
					Category:    "Infrastructure as Code",
					Description: "No Terraform, CloudFormation, Pulumi, or cloud deployment configs",
					Workaround:  "Add IaC configurations manually for your cloud provider",
					Keywords:    []string{"terraform", "cloudformation", "infrastructure as code", "pulumi"},
				},
				{
					Category:    "Rate Limiting",
					Description: "No request rate limiting or throttling",
					Workaround:  "Add Spring Cloud Gateway or Bucket4j manually",
					Keywords:    []string{"rate limit", "throttl"},
				},
				{
					Category:    "Multi-tenancy",
					Description: "No tenant isolation, per-tenant databases, or SaaS multi-tenancy",
					Workaround:  "Implement tenant isolation manually in the persistence layer",
					Keywords:    []string{"multi-tenant", "tenant", "saas"},
				},
				{
					Category:    "API Versioning",
					Description: "No built-in API versioning strategy",
					Workaround:  "Implement URL path or header-based versioning in the API module",
					Keywords:    []string{"api version"},
				},
				{
					Category:    "Database Schema Design",
					Description: "Only placeholder entities and migrations — no production schema",
					Workaround:  "Replace placeholder entities with real domain objects and update Flyway migrations",
					Keywords:    []string{},
				},
				{
					Category:    "Business Logic",
					Description: "No custom business logic — only service structure and patterns",
					Workaround:  "Implement business logic in the Shared module's service classes",
					Keywords:    []string{},
				},
			}

			data, err := json.MarshalIndent(limitations, "", "  ")
			if err != nil {
				return nil, err
			}

			return []mcp.ResourceContents{
				mcp.TextResourceContents{
					URI:      "trabuco://limitations",
					MIMEType: "application/json",
					Text:     string(data),
				},
			}, nil
		},
	)
}
