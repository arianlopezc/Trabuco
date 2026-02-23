package mcp

import (
	"fmt"
	"strings"
)

// ArchitecturePattern represents a pre-built combination of modules for common use cases.
type ArchitecturePattern struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	UseCases        []string `json:"use_cases"`
	Modules         []string `json:"modules"`
	RecommendedDB   string   `json:"recommended_database,omitempty"`
	RecommendedNoDB string   `json:"recommended_nosql_database,omitempty"`
	RecommendedBrkr string   `json:"recommended_broker,omitempty"`
	Constraints     []string `json:"constraints,omitempty"`
	// keywords are internal terms used for matching against requirements.
	keywords []string
}

// matchScore scores how well this pattern matches natural language requirements.
// Returns a value from 0 to 100.
func (p *ArchitecturePattern) matchScore(requirements string) int {
	lower := strings.ToLower(requirements)
	if len(p.keywords) == 0 {
		return 0
	}

	matched := 0
	for _, kw := range p.keywords {
		if strings.Contains(lower, kw) {
			matched++
		}
	}

	// Base score: percentage of keywords matched
	score := (matched * 100) / len(p.keywords)

	// Bonus: each additional keyword match beyond the first adds extra weight
	// to reward patterns with more overlap
	if matched > 1 {
		score += matched * 5
	}

	if score > 100 {
		score = 100
	}
	return score
}

// patternCatalog contains all pre-defined architectural patterns.
var patternCatalog = []ArchitecturePattern{
	{
		Name:          "rest-api",
		Description:   "Simple CRUD backend with a REST API and SQL database",
		UseCases:      []string{"CRUD applications", "REST backends", "simple web services", "admin panels"},
		Modules:       []string{"Model", "SQLDatastore", "Shared", "API"},
		RecommendedDB: "postgresql",
		Constraints:   []string{"No background processing — add Worker if needed later"},
		keywords:      []string{"rest", "api", "crud", "backend", "web service", "endpoint", "http", "database", "sql", "postgresql", "mysql"},
	},
	{
		Name:            "rest-api-nosql",
		Description:     "Document-based REST backend with NoSQL storage",
		UseCases:        []string{"Document stores", "content management", "flexible schemas", "caching layers"},
		Modules:         []string{"Model", "NoSQLDatastore", "Shared", "API"},
		RecommendedNoDB: "mongodb",
		Constraints:     []string{"No SQL migrations — uses NoSQL storage only"},
		keywords:        []string{"rest", "api", "nosql", "mongodb", "document", "redis", "cache", "flexible schema"},
	},
	{
		Name:            "event-driven",
		Description:     "Event-driven architecture with REST API, message broker, and SQL database",
		UseCases:        []string{"Event sourcing", "async workflows", "decoupled microservices", "CQRS"},
		Modules:         []string{"Model", "SQLDatastore", "Shared", "API", "EventConsumer"},
		RecommendedDB:   "postgresql",
		RecommendedBrkr: "kafka",
		Constraints:     []string{"Requires an external message broker (Kafka, RabbitMQ, SQS, or Pub/Sub)"},
		keywords:        []string{"event", "kafka", "rabbitmq", "sqs", "pubsub", "message", "async", "streaming", "event-driven", "cqrs"},
	},
	{
		Name:          "background-processing",
		Description:   "REST API with background job processing and SQL database",
		UseCases:      []string{"Job queues", "scheduled tasks", "batch processing", "async task execution"},
		Modules:       []string{"Model", "SQLDatastore", "Shared", "API", "Worker"},
		RecommendedDB: "postgresql",
		Constraints:   []string{"Worker uses the SQL database for job storage"},
		keywords:      []string{"background", "job", "worker", "scheduled", "cron", "batch", "queue", "async", "delayed", "fire-and-forget"},
	},
	{
		Name:            "full-stack-backend",
		Description:     "Complete backend with REST API, background jobs, event processing, and SQL database",
		UseCases:        []string{"Complex backends", "enterprise applications", "systems needing all capabilities"},
		Modules:         []string{"Model", "SQLDatastore", "Shared", "API", "Worker", "EventConsumer"},
		RecommendedDB:   "postgresql",
		RecommendedBrkr: "kafka",
		Constraints:     []string{"Requires an external message broker", "Most complex setup — start simpler if unsure"},
		keywords:        []string{"full stack", "full backend", "complete backend", "enterprise", "complex backend", "everything", "all modules", "all capabilities", "kitchen sink"},
	},
	{
		Name:        "microservice-light",
		Description: "Stateless microservice with REST API only — no database",
		UseCases:    []string{"API gateways", "proxy services", "aggregation layers", "stateless processors"},
		Modules:     []string{"Model", "Shared", "API"},
		Constraints: []string{"No persistence layer — add SQLDatastore or NoSQLDatastore if storage is needed later"},
		keywords:    []string{"microservice", "stateless", "gateway", "proxy", "lightweight", "no database", "aggregation", "aggregate", "routing layer"},
	},
	{
		Name:          "worker-only",
		Description:   "Headless background processor with SQL database — no REST API",
		UseCases:      []string{"Batch processors", "ETL pipelines", "scheduled data jobs", "headless workers"},
		Modules:       []string{"Model", "SQLDatastore", "Shared", "Worker"},
		RecommendedDB: "postgresql",
		Constraints:   []string{"No HTTP endpoints — add API module if REST access is needed"},
		keywords:      []string{"headless", "processor", "etl", "pipeline", "batch", "data processing", "worker only", "no api", "data import", "ingestion"},
	},
	{
		Name:        "mcp-server",
		Description: "AI tool server exposing MCP (Model Context Protocol) tools",
		UseCases:    []string{"AI tool integration", "coding assistant tooling", "AI-powered development tools"},
		Modules:     []string{"Model", "Shared", "MCP"},
		Constraints: []string{"Generates build/test/review MCP tools — custom tools must be added manually"},
		keywords:    []string{"mcp", "ai tool", "coding assistant", "model context protocol", "ai integration", "llm tool", "ai server", "tool server"},
	},
}

// scorePatterns scores all patterns against requirements and returns them sorted by score (descending).
// Only patterns with score > 0 are returned.
func scorePatterns(requirements string) []scoredPattern {
	var results []scoredPattern
	for _, p := range patternCatalog {
		score := p.matchScore(requirements)
		if score > 0 {
			results = append(results, scoredPattern{
				ArchitecturePattern: p,
				Score:               score,
				Reasoning:           buildPatternReasoning(p, requirements),
			})
		}
	}

	// Sort by score descending (simple insertion sort — small slice)
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Score > results[j-1].Score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}

	// Return top 3
	if len(results) > 3 {
		results = results[:3]
	}
	return results
}

// buildPatternReasoning explains why a pattern matches the requirements.
func buildPatternReasoning(p ArchitecturePattern, requirements string) string {
	lower := strings.ToLower(requirements)
	var matched []string
	for _, kw := range p.keywords {
		if strings.Contains(lower, kw) {
			matched = append(matched, kw)
		}
	}
	if len(matched) == 0 {
		return "Low relevance to the stated requirements"
	}
	return "Matches requirement keywords: " + strings.Join(matched, ", ")
}

// buildRecommendedConfig generates a concrete init_project configuration from the top-scoring pattern.
// Returns nil if no pattern scored high enough.
func buildRecommendedConfig(patterns []scoredPattern) *recommendedConfig {
	if len(patterns) == 0 {
		return nil
	}
	top := patterns[0]
	if top.Score < 20 {
		return nil
	}

	confidence := "high"
	reasoning := "Strong match with the '" + top.Name + "' pattern"

	if top.Score < 50 {
		confidence = "low"
		reasoning = "Weak match — review the pattern details and adjust modules as needed"
	} else if top.Score < 70 {
		confidence = "medium"
		reasoning = "Moderate match with '" + top.Name + "' — verify the module selection fits your needs"
	}

	// If top two patterns are close in score, lower confidence
	if len(patterns) > 1 && (top.Score-patterns[1].Score) <= 10 {
		if confidence == "high" {
			confidence = "medium"
		} else if confidence == "medium" {
			confidence = "low"
		}
		reasoning += fmt.Sprintf("; also consider '%s' pattern (score %d vs %d)", patterns[1].Name, patterns[1].Score, top.Score)
	}

	return &recommendedConfig{
		Modules:       strings.Join(top.Modules, ","),
		Database:      top.RecommendedDB,
		NoSQLDatabase: top.RecommendedNoDB,
		MessageBroker: top.RecommendedBrkr,
		Confidence:    confidence,
		Reasoning:     reasoning,
	}
}
