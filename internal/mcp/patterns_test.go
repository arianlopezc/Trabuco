package mcp

import (
	"strconv"
	"strings"
	"testing"
)

// =============================================================================
// Pattern matching: realistic user requests → correct top pattern
// =============================================================================

func TestPatternMatching_RESTAPIWithPostgres(t *testing.T) {
	// A user says: "I need a REST API with PostgreSQL"
	// Expected: rest-api wins (hits: rest, api, postgresql)
	patterns := scorePatterns("I need a REST API with PostgreSQL")
	requireTopPattern(t, patterns, "rest-api")
	requireModules(t, patterns[0], "Model", "SQLDatastore", "Shared", "API")
}

func TestPatternMatching_CRUDBackendWithDatabase(t *testing.T) {
	// "Build me a CRUD backend with a database and REST endpoints"
	// Expected: rest-api (hits: crud, backend, database, rest, endpoint)
	patterns := scorePatterns("Build me a CRUD backend with a database and REST endpoints")
	requireTopPattern(t, patterns, "rest-api")
	if patterns[0].Score < 40 {
		t.Errorf("Expected score >= 40 for strong match, got %d", patterns[0].Score)
	}
}

func TestPatternMatching_MongoDBDocumentStore(t *testing.T) {
	// "I want a REST API with MongoDB for document storage"
	// Expected: rest-api-nosql (hits: rest, api, mongodb, document)
	patterns := scorePatterns("I want a REST API with MongoDB for document storage")
	requireTopPattern(t, patterns, "rest-api-nosql")
	requireModules(t, patterns[0], "Model", "NoSQLDatastore", "Shared", "API")
}

func TestPatternMatching_RedisCache(t *testing.T) {
	// "REST API with Redis cache"
	// Expected: rest-api-nosql (hits: rest, api, redis, cache)
	patterns := scorePatterns("REST API with Redis cache")
	requireTopPattern(t, patterns, "rest-api-nosql")
}

func TestPatternMatching_KafkaEventStreaming(t *testing.T) {
	// "I need Kafka event streaming with async message processing"
	// Expected: event-driven (hits: kafka, event, streaming, async, message)
	patterns := scorePatterns("I need Kafka event streaming with async message processing")
	requireTopPattern(t, patterns, "event-driven")
	requireModules(t, patterns[0], "Model", "SQLDatastore", "Shared", "API", "EventConsumer")
	if patterns[0].RecommendedBrkr != "kafka" {
		t.Errorf("Expected recommended broker 'kafka', got '%s'", patterns[0].RecommendedBrkr)
	}
}

func TestPatternMatching_RabbitMQConsumer(t *testing.T) {
	// "Event-driven architecture with RabbitMQ message consumers"
	// Expected: event-driven (hits: event-driven, rabbitmq, message, event)
	patterns := scorePatterns("Event-driven architecture with RabbitMQ message consumers")
	requireTopPattern(t, patterns, "event-driven")
}

func TestPatternMatching_SQSProcessing(t *testing.T) {
	// "Process messages from AWS SQS"
	// Expected: event-driven (hits: sqs, message)
	patterns := scorePatterns("Process messages from AWS SQS")
	requireTopPattern(t, patterns, "event-driven")
}

func TestPatternMatching_CQRSArchitecture(t *testing.T) {
	// "CQRS architecture with event streaming"
	// Expected: event-driven (hits: cqrs, event, streaming)
	patterns := scorePatterns("CQRS architecture with event streaming")
	requireTopPattern(t, patterns, "event-driven")
}

func TestPatternMatching_BackgroundJobs(t *testing.T) {
	// "REST API with background job processing and scheduled tasks"
	// Expected: background-processing (hits: background, job, scheduled)
	patterns := scorePatterns("REST API with background job processing and scheduled tasks")
	requireTopPattern(t, patterns, "background-processing")
	requireModules(t, patterns[0], "Model", "SQLDatastore", "Shared", "API", "Worker")
}

func TestPatternMatching_CronScheduledWorker(t *testing.T) {
	// "I need a worker with cron scheduled jobs and batch processing"
	// Expected: background-processing (hits: worker, cron, scheduled, batch, job)
	patterns := scorePatterns("I need a worker with cron scheduled jobs and batch processing")
	requireTopPattern(t, patterns, "background-processing")
}

func TestPatternMatching_FireAndForget(t *testing.T) {
	// "Fire-and-forget async job queue"
	// Expected: background-processing (hits: fire-and-forget, async, queue, job)
	patterns := scorePatterns("Fire-and-forget async job queue")
	requireTopPattern(t, patterns, "background-processing")
}

func TestPatternMatching_EnterpriseFullStack(t *testing.T) {
	// "Complete backend enterprise application with everything"
	// Expected: full-stack-backend (hits: complete backend, enterprise, everything)
	patterns := scorePatterns("Complete backend enterprise application with everything")
	requireTopPattern(t, patterns, "full-stack-backend")
	requireModules(t, patterns[0], "Model", "SQLDatastore", "Shared", "API", "Worker", "EventConsumer")
}

func TestPatternMatching_StatelessMicroservice(t *testing.T) {
	// "Lightweight stateless microservice proxy"
	// Expected: microservice-light (hits: lightweight, stateless, microservice, proxy)
	patterns := scorePatterns("Lightweight stateless microservice proxy")
	requireTopPattern(t, patterns, "microservice-light")
	requireModules(t, patterns[0], "Model", "Shared", "API")
}

func TestPatternMatching_APIGatewayAggregator(t *testing.T) {
	// "API gateway aggregation layer with no database"
	// Expected: microservice-light (hits: gateway, aggregat, no database)
	patterns := scorePatterns("API gateway aggregation layer with no database")
	requireTopPattern(t, patterns, "microservice-light")
}

func TestPatternMatching_ETLPipeline(t *testing.T) {
	// "Headless ETL pipeline for batch data processing"
	// Expected: worker-only (hits: headless, etl, pipeline, batch, data processing)
	patterns := scorePatterns("Headless ETL pipeline for batch data processing")
	requireTopPattern(t, patterns, "worker-only")
	requireModules(t, patterns[0], "Model", "SQLDatastore", "Shared", "Worker")
}

func TestPatternMatching_HeadlessProcessor(t *testing.T) {
	// "Headless processor with no API"
	// Expected: worker-only (hits: headless, processor, no api)
	patterns := scorePatterns("Headless processor with no API")
	requireTopPattern(t, patterns, "worker-only")
}

func TestPatternMatching_MCPServer(t *testing.T) {
	// "MCP server for AI tool integration"
	// Expected: mcp-server (hits: mcp, ai tool, ai integration)
	patterns := scorePatterns("MCP server for AI tool integration")
	requireTopPattern(t, patterns, "mcp-server")
	requireModules(t, patterns[0], "Model", "Shared", "MCP")
}

func TestPatternMatching_CodingAssistantTools(t *testing.T) {
	// "Build a coding assistant with MCP tools"
	// Expected: mcp-server (hits: coding assistant, mcp)
	patterns := scorePatterns("Build a coding assistant with MCP tools")
	requireTopPattern(t, patterns, "mcp-server")
}

// =============================================================================
// Score ordering: more specific requests should score higher
// =============================================================================

func TestScoreOrdering_SpecificBeatsVague(t *testing.T) {
	// "REST API" is vague — low score for rest-api
	// "REST API with PostgreSQL database for CRUD backend" is specific — high score
	vague := scorePatterns("REST API")
	specific := scorePatterns("REST API with PostgreSQL database for CRUD backend")

	if len(vague) == 0 || len(specific) == 0 {
		t.Fatal("Expected both to return results")
	}

	if specific[0].Score <= vague[0].Score {
		t.Errorf("Specific request should score higher: specific=%d, vague=%d", specific[0].Score, vague[0].Score)
	}
}

func TestScoreOrdering_ResultsAreSortedDescending(t *testing.T) {
	// "async message processing" hits both event-driven and background-processing
	patterns := scorePatterns("async message processing")
	for i := 1; i < len(patterns); i++ {
		if patterns[i].Score > patterns[i-1].Score {
			t.Errorf("Patterns not sorted: index %d (%s, score %d) > index %d (%s, score %d)",
				i, patterns[i].Name, patterns[i].Score,
				i-1, patterns[i-1].Name, patterns[i-1].Score)
		}
	}
}

func TestScoreOrdering_MaxThreeResults(t *testing.T) {
	// A broad request like "REST API backend with async queue and batch job worker"
	// should match many patterns but return at most 3
	patterns := scorePatterns("REST API backend with async queue and batch job worker")
	if len(patterns) > 3 {
		t.Errorf("Expected at most 3 patterns, got %d", len(patterns))
	}
}

// =============================================================================
// matchScore unit tests
// =============================================================================

func TestMatchScore_NoKeywordsMatch(t *testing.T) {
	p := &ArchitecturePattern{keywords: []string{"foo", "bar", "baz"}}
	score := p.matchScore("something completely different")
	if score != 0 {
		t.Errorf("Expected score 0, got %d", score)
	}
}

func TestMatchScore_OneKeywordMatch(t *testing.T) {
	p := &ArchitecturePattern{keywords: []string{"rest", "api", "crud", "database"}}
	score := p.matchScore("I need a REST service")
	// 1 match out of 4: base = 25, no bonus for single match
	if score != 25 {
		t.Errorf("Expected score 25, got %d", score)
	}
}

func TestMatchScore_MultipleKeywordsMatch(t *testing.T) {
	p := &ArchitecturePattern{keywords: []string{"rest", "api", "crud", "database"}}
	score := p.matchScore("REST API with CRUD database")
	// 4 matches out of 4: base = 100, bonus = 4*5 = 20, capped at 100
	if score != 100 {
		t.Errorf("Expected score 100, got %d", score)
	}
}

func TestMatchScore_EmptyKeywords(t *testing.T) {
	p := &ArchitecturePattern{keywords: []string{}}
	score := p.matchScore("anything")
	if score != 0 {
		t.Errorf("Expected score 0 for empty keywords, got %d", score)
	}
}

func TestMatchScore_CaseInsensitive(t *testing.T) {
	p := &ArchitecturePattern{keywords: []string{"kafka", "event"}}
	score := p.matchScore("KAFKA EVENT streaming")
	if score == 0 {
		t.Error("Expected non-zero score for case-insensitive match")
	}
}

func TestMatchScore_CappedAt100(t *testing.T) {
	// Pattern with only 2 keywords, both match
	p := &ArchitecturePattern{keywords: []string{"api", "rest"}}
	score := p.matchScore("REST API")
	// base = 100, bonus = 10 → capped at 100
	if score > 100 {
		t.Errorf("Score should be capped at 100, got %d", score)
	}
}

// =============================================================================
// scorePatterns edge cases
// =============================================================================

func TestScorePatterns_EmptyInput(t *testing.T) {
	patterns := scorePatterns("")
	if len(patterns) != 0 {
		t.Errorf("Expected no patterns for empty input, got %d", len(patterns))
	}
}

func TestScorePatterns_Gibberish(t *testing.T) {
	patterns := scorePatterns("xyzzy plugh twisty little passages")
	if len(patterns) != 0 {
		t.Errorf("Expected no patterns for gibberish, got %d", len(patterns))
	}
}

func TestScorePatterns_OnlyStopWords(t *testing.T) {
	patterns := scorePatterns("I need to build a system that does something")
	// None of these words are pattern keywords
	if len(patterns) != 0 {
		t.Errorf("Expected no patterns for generic words, got %d", len(patterns))
	}
}

// =============================================================================
// buildRecommendedConfig tests
// =============================================================================

func TestRecommendedConfig_HighConfidence(t *testing.T) {
	// Force a high-scoring pattern
	patterns := scorePatterns("REST API CRUD backend with PostgreSQL database and SQL endpoints")
	if len(patterns) == 0 {
		t.Fatal("Expected patterns")
	}
	rec := buildRecommendedConfig(patterns)
	if rec == nil {
		t.Fatal("Expected recommended config")
	}
	if rec.Database != "postgresql" {
		t.Errorf("Expected database 'postgresql', got '%s'", rec.Database)
	}
	if !strings.Contains(rec.Modules, "API") {
		t.Errorf("Expected modules to contain 'API', got '%s'", rec.Modules)
	}
	if !strings.Contains(rec.Modules, "SQLDatastore") {
		t.Errorf("Expected modules to contain 'SQLDatastore', got '%s'", rec.Modules)
	}
}

func TestRecommendedConfig_NilForEmptyPatterns(t *testing.T) {
	rec := buildRecommendedConfig(nil)
	if rec != nil {
		t.Error("Expected nil recommended config for empty patterns")
	}
	rec = buildRecommendedConfig([]scoredPattern{})
	if rec != nil {
		t.Error("Expected nil recommended config for empty slice")
	}
}

func TestRecommendedConfig_NilForVeryLowScore(t *testing.T) {
	patterns := []scoredPattern{
		{ArchitecturePattern: ArchitecturePattern{Name: "test"}, Score: 10},
	}
	rec := buildRecommendedConfig(patterns)
	if rec != nil {
		t.Errorf("Expected nil for score < 20, got config with confidence '%s'", rec.Confidence)
	}
}

func TestRecommendedConfig_LowConfidenceForWeakMatch(t *testing.T) {
	patterns := []scoredPattern{
		{ArchitecturePattern: ArchitecturePattern{
			Name:    "rest-api",
			Modules: []string{"Model", "SQLDatastore", "Shared", "API"},
		}, Score: 30},
	}
	rec := buildRecommendedConfig(patterns)
	if rec == nil {
		t.Fatal("Expected config for score 30")
	}
	if rec.Confidence != "low" {
		t.Errorf("Expected 'low' confidence for score 30, got '%s'", rec.Confidence)
	}
}

func TestRecommendedConfig_MediumConfidenceForModerateMatch(t *testing.T) {
	patterns := []scoredPattern{
		{ArchitecturePattern: ArchitecturePattern{
			Name:    "rest-api",
			Modules: []string{"Model", "SQLDatastore", "Shared", "API"},
		}, Score: 60},
	}
	rec := buildRecommendedConfig(patterns)
	if rec == nil {
		t.Fatal("Expected config for score 60")
	}
	if rec.Confidence != "medium" {
		t.Errorf("Expected 'medium' confidence for score 60, got '%s'", rec.Confidence)
	}
}

func TestRecommendedConfig_ConfidenceLoweredWhenTopTwoClose(t *testing.T) {
	// Two patterns within 10 points should lower confidence
	patterns := []scoredPattern{
		{ArchitecturePattern: ArchitecturePattern{
			Name:    "event-driven",
			Modules: []string{"Model", "SQLDatastore", "Shared", "API", "EventConsumer"},
		}, Score: 75},
		{ArchitecturePattern: ArchitecturePattern{
			Name:    "background-processing",
			Modules: []string{"Model", "SQLDatastore", "Shared", "API", "Worker"},
		}, Score: 70},
	}
	rec := buildRecommendedConfig(patterns)
	if rec == nil {
		t.Fatal("Expected config")
	}
	// Would be "high" (score 75) but lowered to "medium" because top 2 are within 10 pts
	if rec.Confidence != "medium" {
		t.Errorf("Expected 'medium' (lowered from high due to close scores), got '%s'", rec.Confidence)
	}
	if !strings.Contains(rec.Reasoning, "background-processing") {
		t.Error("Reasoning should mention the competing pattern")
	}
}

func TestRecommendedConfig_EventDrivenIncludesBroker(t *testing.T) {
	patterns := scorePatterns("Kafka event-driven message streaming with async consumers")
	rec := buildRecommendedConfig(patterns)
	if rec == nil {
		t.Fatal("Expected config")
	}
	if rec.MessageBroker != "kafka" {
		t.Errorf("Expected broker 'kafka', got '%s'", rec.MessageBroker)
	}
}

func TestRecommendedConfig_NoSQLIncludesNoSQLDB(t *testing.T) {
	patterns := scorePatterns("REST API with MongoDB document store for flexible NoSQL data")
	rec := buildRecommendedConfig(patterns)
	if rec == nil {
		t.Fatal("Expected config")
	}
	if rec.NoSQLDatabase != "mongodb" {
		t.Errorf("Expected nosql_database 'mongodb', got '%s'", rec.NoSQLDatabase)
	}
	if rec.Database != "" {
		t.Errorf("Expected empty database for NoSQL pattern, got '%s'", rec.Database)
	}
}

// =============================================================================
// buildAdvisory integration: verify patterns and config alongside existing fields
// =============================================================================

func TestBuildAdvisory_IncludesPatternsAndConfig(t *testing.T) {
	advisory := buildAdvisory("REST API with PostgreSQL database")
	if len(advisory.Patterns) == 0 {
		t.Error("Expected patterns in advisory")
	}
	if advisory.RecommendedConfig == nil {
		t.Error("Expected recommended_config in advisory")
	}
	// Existing fields should still be populated
	if len(advisory.Modules) == 0 {
		t.Error("Expected modules in advisory")
	}
	if len(advisory.DatabaseOptions) == 0 {
		t.Error("Expected database_options in advisory")
	}
	if len(advisory.BrokerOptions) == 0 {
		t.Error("Expected broker_options in advisory")
	}
	if len(advisory.Constraints) == 0 {
		t.Error("Expected constraints in advisory")
	}
}

func TestBuildAdvisory_EmptyPatternsForGibberish(t *testing.T) {
	advisory := buildAdvisory("xyzzy plugh nothing relevant")
	if len(advisory.Patterns) != 0 {
		t.Errorf("Expected no patterns for gibberish, got %d", len(advisory.Patterns))
	}
	if advisory.RecommendedConfig != nil {
		t.Error("Expected no recommended config for gibberish")
	}
}

func TestBuildAdvisory_DisambiguationWarnings(t *testing.T) {
	// "event" without broker keywords should trigger disambiguation
	advisory := buildAdvisory("I need to process events in my system")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'event'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'event'")
	}
}

func TestBuildAdvisory_QueueDisambiguation(t *testing.T) {
	advisory := buildAdvisory("I need a queue for tasks")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'queue'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'queue'")
	}
}

func TestBuildAdvisory_NoDisambiguationWhenBrokerSpecified(t *testing.T) {
	// "event" WITH a broker keyword should NOT trigger disambiguation
	advisory := buildAdvisory("I need Kafka event processing")
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'event'") {
			t.Error("Should NOT warn about 'event' when Kafka is mentioned")
		}
	}
}

func TestBuildAdvisory_UnsupportedRequirements(t *testing.T) {
	advisory := buildAdvisory("REST API with OAuth login and GraphQL")
	if len(advisory.NotCovered) == 0 {
		t.Error("Expected not_covered items")
	}
	foundAuth := false
	foundGraphQL := false
	for _, nc := range advisory.NotCovered {
		if strings.Contains(nc, "Authentication") {
			foundAuth = true
		}
		if strings.Contains(nc, "GraphQL") {
			foundGraphQL = true
		}
	}
	if !foundAuth {
		t.Error("Expected auth unsupported warning for 'OAuth login'")
	}
	if !foundGraphQL {
		t.Error("Expected GraphQL unsupported warning")
	}
}

func TestBuildAdvisory_CacheDisambiguation(t *testing.T) {
	advisory := buildAdvisory("I need caching in my application")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'cache'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'cache'")
	}
}

func TestBuildAdvisory_AsyncDisambiguation(t *testing.T) {
	advisory := buildAdvisory("I need asynchronous processing")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'async'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'async'")
	}
}

// =============================================================================
// Cross-pattern disambiguation: when "async" or "queue" hits multiple patterns
// =============================================================================

func TestCrossPattern_AsyncHitsBothEventAndBackground(t *testing.T) {
	// "async" is a keyword in both event-driven and background-processing
	// Both should appear in results, but the one with more keyword overlap wins
	patterns := scorePatterns("async processing")
	if len(patterns) < 2 {
		t.Fatalf("Expected at least 2 patterns for 'async processing', got %d", len(patterns))
	}
	// Both event-driven and background-processing should be present
	names := patternNames(patterns)
	if !containsName(names, "event-driven") && !containsName(names, "background-processing") {
		t.Errorf("Expected both event-driven and background-processing in results, got %v", names)
	}
}

func TestCrossPattern_QueueResolvesToBackground(t *testing.T) {
	// "queue" is only in background-processing, not event-driven
	// "job queue" should clearly be background-processing
	patterns := scorePatterns("job queue for background tasks")
	requireTopPattern(t, patterns, "background-processing")
}

func TestCrossPattern_KafkaResolvesToEventDriven(t *testing.T) {
	// "kafka" is unique to event-driven
	patterns := scorePatterns("kafka consumers")
	requireTopPattern(t, patterns, "event-driven")
}

func TestCrossPattern_BatchHitsBothWorkerPatterns(t *testing.T) {
	// "batch" appears in both background-processing and worker-only
	// "batch" alone should match both
	patterns := scorePatterns("batch processing")
	if len(patterns) < 2 {
		t.Fatalf("Expected at least 2 patterns for 'batch processing', got %d", len(patterns))
	}
	names := patternNames(patterns)
	if !containsName(names, "background-processing") {
		t.Errorf("Expected background-processing, got %v", names)
	}
	if !containsName(names, "worker-only") {
		t.Errorf("Expected worker-only, got %v", names)
	}
}

func TestCrossPattern_BatchWithJobAndWorkerResolvesToBackground(t *testing.T) {
	// Adding "job" and "worker" tilts the score toward background-processing
	// because those are strong background-processing keywords
	patterns := scorePatterns("API with background batch job worker")
	requireTopPattern(t, patterns, "background-processing")
}

func TestCrossPattern_BatchWithHeadlessResolvesToWorkerOnly(t *testing.T) {
	// "batch" + "headless" should favor worker-only
	patterns := scorePatterns("headless batch processor")
	requireTopPattern(t, patterns, "worker-only")
}

// =============================================================================
// End-to-end scenario tests: realistic requests produce the right config
// =============================================================================

func TestE2E_EcommerceAPI(t *testing.T) {
	// "I'm building an e-commerce REST API with PostgreSQL for storing products and orders"
	advisory := buildAdvisory("I'm building an e-commerce REST API with PostgreSQL for storing products and orders")
	if advisory.RecommendedConfig == nil {
		t.Fatal("Expected recommended config")
	}
	rec := advisory.RecommendedConfig
	if !strings.Contains(rec.Modules, "API") {
		t.Error("E-commerce API should include API module")
	}
	if !strings.Contains(rec.Modules, "SQLDatastore") {
		t.Error("E-commerce with PostgreSQL should include SQLDatastore")
	}
	if rec.Database != "postgresql" {
		t.Errorf("Expected postgresql, got '%s'", rec.Database)
	}
}

func TestE2E_NotificationSystem(t *testing.T) {
	// "Build a notification system with background email sending and scheduled digests"
	advisory := buildAdvisory("Build a notification system with background email sending and scheduled digests")
	if advisory.RecommendedConfig == nil {
		t.Fatal("Expected recommended config")
	}
	rec := advisory.RecommendedConfig
	if !strings.Contains(rec.Modules, "Worker") {
		t.Error("Notification system with scheduled sending should include Worker")
	}
}

func TestE2E_KafkaOrderProcessing(t *testing.T) {
	// "Event-driven order processing with Kafka and PostgreSQL"
	advisory := buildAdvisory("Event-driven order processing with Kafka and PostgreSQL")
	requireTopPattern(t, advisory.Patterns, "event-driven")
	rec := advisory.RecommendedConfig
	if rec == nil {
		t.Fatal("Expected recommended config")
	}
	if !strings.Contains(rec.Modules, "EventConsumer") {
		t.Error("Kafka processing should include EventConsumer")
	}
	if rec.MessageBroker != "kafka" {
		t.Errorf("Expected kafka broker, got '%s'", rec.MessageBroker)
	}
	if rec.Database != "postgresql" {
		t.Errorf("Expected postgresql, got '%s'", rec.Database)
	}
}

func TestE2E_AIToolServer(t *testing.T) {
	// "I want to build an MCP server so coding assistants can use my AI tools"
	advisory := buildAdvisory("I want to build an MCP server so coding assistants can use my AI tools")
	requireTopPattern(t, advisory.Patterns, "mcp-server")
	rec := advisory.RecommendedConfig
	if rec == nil {
		t.Fatal("Expected recommended config")
	}
	if !strings.Contains(rec.Modules, "MCP") {
		t.Error("MCP server request should include MCP module")
	}
	// Should NOT include database or broker
	if rec.Database != "" {
		t.Errorf("MCP server shouldn't recommend database, got '%s'", rec.Database)
	}
	if rec.MessageBroker != "" {
		t.Errorf("MCP server shouldn't recommend broker, got '%s'", rec.MessageBroker)
	}
}

func TestE2E_DataIngestionPipeline(t *testing.T) {
	// "Headless ETL data processing pipeline for batch imports"
	advisory := buildAdvisory("Headless ETL data processing pipeline for batch imports")
	requireTopPattern(t, advisory.Patterns, "worker-only")
	rec := advisory.RecommendedConfig
	if rec == nil {
		t.Fatal("Expected recommended config")
	}
	if !strings.Contains(rec.Modules, "Worker") {
		t.Error("ETL pipeline should include Worker")
	}
	if strings.Contains(rec.Modules, "API") {
		t.Error("Headless ETL should NOT include API")
	}
}

func TestE2E_MixedUnsupported(t *testing.T) {
	// "REST API with JWT auth, React frontend, and WebSocket chat"
	advisory := buildAdvisory("REST API with JWT auth, React frontend, and WebSocket chat")
	if len(advisory.NotCovered) < 3 {
		t.Errorf("Expected at least 3 not_covered items (auth, frontend, websocket), got %d", len(advisory.NotCovered))
	}
	// But should still recommend a pattern (REST API is mentioned)
	if len(advisory.Patterns) == 0 {
		t.Error("Should still match rest-api pattern despite unsupported items")
	}
}

// =============================================================================
// findPattern tests
// =============================================================================

func TestFindPattern_AllPatternsExist(t *testing.T) {
	names := []string{"rest-api", "rest-api-nosql", "event-driven", "background-processing",
		"full-stack-backend", "microservice-light", "worker-only", "mcp-server"}
	for _, name := range names {
		p := findPattern(name)
		if p == nil {
			t.Errorf("Pattern '%s' not found in catalog", name)
		}
	}
}

func TestFindPattern_NonExistent(t *testing.T) {
	p := findPattern("nonexistent")
	if p != nil {
		t.Error("Expected nil for non-existent pattern")
	}
}

func TestFindPattern_AllHaveModules(t *testing.T) {
	for _, p := range patternCatalog {
		if len(p.Modules) == 0 {
			t.Errorf("Pattern '%s' has no modules", p.Name)
		}
		// All patterns should include Model
		if !containsStr(p.Modules, "Model") {
			t.Errorf("Pattern '%s' doesn't include Model", p.Name)
		}
	}
}

func TestFindPattern_AllHaveKeywords(t *testing.T) {
	for _, p := range patternCatalog {
		if len(p.keywords) == 0 {
			t.Errorf("Pattern '%s' has no keywords", p.Name)
		}
	}
}

// =============================================================================
// buildPatternReasoning tests
// =============================================================================

func TestBuildPatternReasoning_IncludesMatchedKeywords(t *testing.T) {
	p := patternCatalog[0] // rest-api
	reasoning := buildPatternReasoning(p, "REST API with database")
	if !strings.Contains(reasoning, "rest") {
		t.Error("Reasoning should mention matched keyword 'rest'")
	}
	if !strings.Contains(reasoning, "api") {
		t.Error("Reasoning should mention matched keyword 'api'")
	}
	if !strings.Contains(reasoning, "database") {
		t.Error("Reasoning should mention matched keyword 'database'")
	}
}

func TestBuildPatternReasoning_LowRelevanceForNoMatch(t *testing.T) {
	p := patternCatalog[0] // rest-api
	reasoning := buildPatternReasoning(p, "xyzzy nothing relevant")
	if !strings.Contains(reasoning, "Low relevance") {
		t.Errorf("Expected 'Low relevance' for no match, got '%s'", reasoning)
	}
}

// =============================================================================
// Strengthened keyword tests: new keywords added in polish pass
// =============================================================================

func TestPatternMatching_FullStackViaExplicit(t *testing.T) {
	// Explicit "full stack backend" should match
	patterns := scorePatterns("I need a full stack backend with all capabilities")
	requireTopPattern(t, patterns, "full-stack-backend")
}

func TestPatternMatching_FullStackViaKitchenSink(t *testing.T) {
	// "kitchen sink" + "enterprise" should hit full-stack-backend
	patterns := scorePatterns("Give me the enterprise kitchen sink")
	requireTopPattern(t, patterns, "full-stack-backend")
}

func TestPatternMatching_ImplicitComboShowsMultiplePatterns(t *testing.T) {
	// When a user says "API with workers and Kafka events", they don't say "full stack"
	// explicitly. The system should return MULTIPLE patterns (event-driven, background-
	// processing) — the agent then combines them. This is the advisory working as designed.
	patterns := scorePatterns("REST API with background workers and Kafka event consumers")
	if len(patterns) < 2 {
		t.Fatalf("Expected multiple patterns for combo request, got %d", len(patterns))
	}
	// Both event-driven and background-processing should appear
	names := patternNames(patterns)
	hasEvent := containsName(names, "event-driven")
	hasBackground := containsName(names, "background-processing")
	if !hasEvent || !hasBackground {
		t.Errorf("Expected both event-driven and background-processing, got %v", names)
	}
}

func TestPatternMatching_MCPViaLLMTool(t *testing.T) {
	// "build an LLM tool server" should now hit mcp-server
	patterns := scorePatterns("Build an LLM tool server")
	requireTopPattern(t, patterns, "mcp-server")
}

func TestPatternMatching_MCPViaAIServer(t *testing.T) {
	// "AI server for tools" should match mcp-server
	patterns := scorePatterns("AI server for exposing tools")
	requireTopPattern(t, patterns, "mcp-server")
}

func TestPatternMatching_MCPViaToolServer(t *testing.T) {
	patterns := scorePatterns("Build a tool server for coding")
	requireTopPattern(t, patterns, "mcp-server")
}

func TestPatternMatching_AggregationLayer(t *testing.T) {
	// "aggregation" keyword matches microservice-light
	patterns := scorePatterns("Stateless aggregation microservice")
	requireTopPattern(t, patterns, "microservice-light")
}

func TestPatternMatching_DataIngestionWorker(t *testing.T) {
	// New keyword "data import" and "ingestion" for worker-only
	patterns := scorePatterns("Data import ingestion pipeline")
	requireTopPattern(t, patterns, "worker-only")
}

func TestPatternMatching_RoutingLayer(t *testing.T) {
	// New keyword "routing layer" for microservice-light
	patterns := scorePatterns("Lightweight routing layer microservice")
	requireTopPattern(t, patterns, "microservice-light")
}

// =============================================================================
// Disambiguation: new warnings for batch, notification, streaming, webhook
// =============================================================================

func TestBuildAdvisory_BatchDisambiguation(t *testing.T) {
	// "batch" without worker/background context should warn
	advisory := buildAdvisory("I need batch processing for my data")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'batch'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'batch'")
	}
}

func TestBuildAdvisory_BatchNoWarningWithWorker(t *testing.T) {
	// "batch" WITH "worker" should NOT warn (disambiguated)
	advisory := buildAdvisory("I need a worker for batch processing")
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'batch'") {
			t.Error("Should NOT warn about 'batch' when 'worker' is mentioned")
		}
	}
}

func TestBuildAdvisory_NotificationDisambiguation(t *testing.T) {
	// "notification" without worker/broker context should warn
	advisory := buildAdvisory("I need to send notifications to users")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'notification'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'notification'")
	}
}

func TestBuildAdvisory_NotificationNoWarningWithWorker(t *testing.T) {
	advisory := buildAdvisory("background worker for sending notifications")
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'notification'") {
			t.Error("Should NOT warn about 'notification' when 'worker' is mentioned")
		}
	}
}

func TestBuildAdvisory_StreamingDisambiguation(t *testing.T) {
	// "streaming" without broker keywords should warn
	advisory := buildAdvisory("I need data streaming for my application")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'streaming'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'streaming'")
	}
}

func TestBuildAdvisory_StreamingNoWarningWithKafka(t *testing.T) {
	advisory := buildAdvisory("Kafka streaming for event processing")
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'streaming'") {
			t.Error("Should NOT warn about 'streaming' when 'Kafka' is mentioned")
		}
	}
}

func TestBuildAdvisory_WebhookDisambiguation(t *testing.T) {
	advisory := buildAdvisory("I need to handle webhooks from Stripe")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'webhook'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'webhook'")
	}
}

func TestBuildAdvisory_CallbackDisambiguation(t *testing.T) {
	advisory := buildAdvisory("Process callback URLs from payment providers")
	found := false
	for _, w := range advisory.Warnings {
		if strings.Contains(w, "Ambiguous term 'webhook'") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected disambiguation warning for 'callback'")
	}
}

// =============================================================================
// Unsupported detection: synced entries
// =============================================================================

func TestBuildAdvisory_PulumiUnsupported(t *testing.T) {
	advisory := buildAdvisory("Deploy with Pulumi infrastructure as code")
	found := false
	for _, nc := range advisory.NotCovered {
		if strings.Contains(nc, "Infrastructure as code") {
			found = true
		}
	}
	if !found {
		t.Error("Expected unsupported warning for 'Pulumi'")
	}
}

// =============================================================================
// Test helpers
// =============================================================================

func requireTopPattern(t *testing.T, patterns []scoredPattern, expectedName string) {
	t.Helper()
	if len(patterns) == 0 {
		t.Fatalf("Expected patterns, got none")
	}
	if patterns[0].Name != expectedName {
		names := patternNames(patterns)
		t.Errorf("Expected top pattern '%s', got '%s' (score %d). All: %v",
			expectedName, patterns[0].Name, patterns[0].Score, names)
	}
}

func requireModules(t *testing.T, pattern scoredPattern, expected ...string) {
	t.Helper()
	for _, mod := range expected {
		if !containsStr(pattern.Modules, mod) {
			t.Errorf("Pattern '%s' missing expected module '%s'. Has: %v",
				pattern.Name, mod, pattern.Modules)
		}
	}
}

func patternNames(patterns []scoredPattern) []string {
	names := make([]string, len(patterns))
	for i, p := range patterns {
		names[i] = p.Name + ":" + strconv.Itoa(p.Score)
	}
	return names
}

func containsName(names []string, prefix string) bool {
	for _, n := range names {
		if strings.HasPrefix(n, prefix+":") {
			return true
		}
	}
	return false
}

func containsStr(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
