package config

import "github.com/arianlopezc/Trabuco/internal/utils"

// ProjectConfig holds all configuration for a generated project
type ProjectConfig struct {
	// Basic info
	ProjectName string // e.g., "my-platform"
	GroupID     string // e.g., "com.company.project"
	ArtifactID  string // e.g., "my-platform" (usually same as ProjectName)

	// Java
	JavaVersion         string // "21", "25", or "26"
	JavaVersionDetected bool   // Whether the selected Java version was detected on the system

	// Modules
	Modules []string // e.g., ["Model", "SQLDatastore", "NoSQLDatastore", "Shared", "API"]

	// SQL Database (only if SQLDatastore selected)
	Database string // "postgresql", "mysql", or "generic"

	// NoSQL Database (only if NoSQLDatastore selected)
	NoSQLDatabase string // "mongodb" or "redis"

	// Message Broker (only if EventConsumer selected)
	MessageBroker string // "kafka" or "rabbitmq"

	// AI Coding Agents
	AIAgents []string // Selected agents: "claude", "cursor", "copilot", "codex"

	// CI/CD Provider
	CIProvider string // "github" or "" (empty = none)

	// VectorStore: vector-similarity backend for the AIAgent module's
	// RAG (Retrieval-Augmented Generation) layer. Empty / "none" =
	// keyword-only knowledge retrieval (the default). When set, the
	// AIAgent module ships embedding + vector-store dependencies and a
	// VectorKnowledgeRetriever that supplants the keyword fallback via
	// @ConditionalOnMissingBean(VectorStore.class).
	//   - "pgvector"  — PGVector inside the existing Postgres datastore
	//                   (separate `vector` schema, separate Flyway bean)
	//   - "qdrant"    — standalone Qdrant container (best raw perf)
	//   - "mongodb"   — MongoDB Atlas Vector Search (cloud-only;
	//                   community Mongo cannot serve $vectorSearch)
	//   - "" / "none" — no RAG; keyword retrieval only
	VectorStore string

	// Review: on-turn code review automation (subagents + hooks + skills)
	Review ReviewConfig

	// Deprecated: Use AIAgents instead
	IncludeCLAUDEMD bool // Legacy field for backwards compatibility
}

// ReviewMode selects how much review scaffolding to emit.
const (
	ReviewModeFull    = "full"    // subagents + skills + hooks (default)
	ReviewModeMinimal = "minimal" // subagents + skills, no Stop hook guard
	ReviewModeOff     = "off"     // no review artifacts at all
)

// ReviewConfig controls what review artifacts are generated.
//
// Semantics:
//   - Mode == "off"      → skip all review emission
//   - Mode == "minimal"  → emit subagents + skills + CLAUDE.md directive; no Stop hook
//   - Mode == "full"     → emit everything including the Stop hook enforcement guard
type ReviewConfig struct {
	Mode        string // "full" | "minimal" | "off"
	GeneratedAt string // RFC3339 timestamp of generation
}

// Derived helper methods

// PackagePath returns the group ID as a file path (e.g., "com/company/project")
func (c *ProjectConfig) PackagePath() string {
	path := ""
	for _, ch := range c.GroupID {
		if ch == '.' {
			path += "/"
		} else {
			path += string(ch)
		}
	}
	return path
}

// ProjectNamePascal returns the project name in PascalCase (e.g., "MyPlatform")
func (c *ProjectConfig) ProjectNamePascal() string {
	return utils.ToPascalCase(c.ProjectName)
}

// ProjectNameCamel returns the project name in camelCase (e.g., "myPlatform")
func (c *ProjectConfig) ProjectNameCamel() string {
	return utils.ToCamelCase(c.ProjectName)
}

// ProjectNameSnake returns the project name in snake_case (e.g., "my_platform")
func (c *ProjectConfig) ProjectNameSnake() string {
	result := ""
	for _, ch := range c.ProjectName {
		if ch == '-' {
			result += "_"
		} else {
			result += string(ch)
		}
	}
	return result
}

// HasModule checks if a specific module is included
func (c *ProjectConfig) HasModule(name string) bool {
	for _, m := range c.Modules {
		if m == name {
			return true
		}
	}
	return false
}

// HasAllModules checks if all specified modules are included
func (c *ProjectConfig) HasAllModules(names ...string) bool {
	for _, name := range names {
		if !c.HasModule(name) {
			return false
		}
	}
	return true
}

// HasAnyDatastore checks if any datastore module is included
func (c *ProjectConfig) HasAnyDatastore() bool {
	return c.HasModule(ModuleSQLDatastore) || c.HasModule(ModuleNoSQLDatastore)
}

// HasBothDatastores checks if both datastore modules are included
func (c *ProjectConfig) HasBothDatastores() bool {
	return c.HasModule(ModuleSQLDatastore) && c.HasModule(ModuleNoSQLDatastore)
}

// HasAIAgentModule returns true if the AI Agent module is selected
func (c *ProjectConfig) HasAIAgentModule() bool {
	return c.HasModule(ModuleAIAgent)
}

// JobRunr Storage Configuration Helpers
// These determine what storage backend JobRunr should use for job persistence.
// The storage is separate from the main application datastore to allow for
// independent scaling and configuration in production.

// JobRunrStorageType returns the storage type for JobRunr:
// - "sql" if SQLDatastore is selected (PostgreSQL or MySQL)
// - "mongodb" if NoSQLDatastore with MongoDB is selected
// - "sql" (PostgreSQL fallback) if NoSQLDatastore with Redis is selected (Redis deprecated in JobRunr 8)
// - "sql" (PostgreSQL fallback) if no datastore is selected but Worker is
// - "" if Worker is not selected (no storage needed)
func (c *ProjectConfig) JobRunrStorageType() string {
	// Only relevant when Worker module is selected
	if !c.HasModule(ModuleWorker) {
		return ""
	}

	if c.HasModule(ModuleSQLDatastore) {
		return "sql"
	}
	if c.HasModule(ModuleNoSQLDatastore) {
		if c.NoSQLDatabase == DatabaseMongoDB {
			return DatabaseMongoDB
		}
		// Redis is deprecated in JobRunr 8, fallback to PostgreSQL
		return "sql"
	}
	// No datastore selected but Worker module is, fallback to PostgreSQL
	return "sql"
}

// JobRunrUsesSql returns true if JobRunr should use SQL storage
func (c *ProjectConfig) JobRunrUsesSql() bool {
	return c.JobRunrStorageType() == "sql"
}

// JobRunrUsesMongoDB returns true if JobRunr should use MongoDB storage
func (c *ProjectConfig) JobRunrUsesMongoDB() bool {
	return c.JobRunrStorageType() == DatabaseMongoDB
}

// JobRunrSqlDatabase returns the SQL database type for JobRunr storage:
// - If SQLDatastore is selected, uses the same database type
// - Otherwise, defaults to "postgresql" (for Redis fallback or no datastore)
func (c *ProjectConfig) JobRunrSqlDatabase() string {
	if c.HasModule(ModuleSQLDatastore) {
		return c.Database
	}
	return DatabasePostgreSQL
}

// WorkerUsesPostgresFallback returns true if Worker is using PostgreSQL
// as a fallback because the user selected Redis (which is deprecated in JobRunr 8)
// or no datastore at all
func (c *ProjectConfig) WorkerUsesPostgresFallback() bool {
	if !c.HasModule(ModuleWorker) {
		return false
	}
	// Redis fallback - Redis is deprecated in JobRunr 8+
	if c.HasModule(ModuleNoSQLDatastore) && c.NoSQLDatabase == DatabaseRedis {
		return true
	}
	// No datastore selected - Worker uses PostgreSQL fallback for JobRunr storage
	if !c.HasModule(ModuleSQLDatastore) && !c.HasModule(ModuleNoSQLDatastore) {
		return true
	}
	return false
}

// WorkerNeedsOwnPostgres returns true if Worker needs its own PostgreSQL
// instance because no SQL datastore is available for JobRunr storage.
// This happens when:
// - NoSQLDatastore with Redis is selected (Redis deprecated in JobRunr 8+)
// - No datastore is selected at all
func (c *ProjectConfig) WorkerNeedsOwnPostgres() bool {
	if !c.HasModule(ModuleWorker) {
		return false
	}
	// If NoSQLDatastore with Redis is selected, Worker needs PostgreSQL for JobRunr
	if c.HasModule(ModuleNoSQLDatastore) && c.NoSQLDatabase == DatabaseRedis {
		return true
	}
	// If no datastore is selected, Worker needs PostgreSQL for JobRunr
	if !c.HasModule(ModuleSQLDatastore) && !c.HasModule(ModuleNoSQLDatastore) {
		return true
	}
	return false
}

// NeedsDockerCompose returns true if docker-compose.yml should be generated.
// This is the case when a runtime module (API or Worker) needs a datastore,
// when Worker needs its own PostgreSQL for JobRunr storage,
// or when EventConsumer needs a message broker.
func (c *ProjectConfig) NeedsDockerCompose() bool {
	hasDatastore := (c.HasModule(ModuleSQLDatastore) && c.Database != "") ||
		(c.HasModule(ModuleNoSQLDatastore) && c.NoSQLDatabase != "")
	hasRuntime := c.HasModule(ModuleAPI) || c.HasModule(ModuleWorker)
	return (hasRuntime && hasDatastore) || c.WorkerNeedsOwnPostgres() || c.EventConsumerNeedsDockerCompose()
}

// ShowRedisWorkerWarning returns true if a warning should be shown about
// Redis + Worker combination (Redis is deprecated in JobRunr 8+)
func (c *ProjectConfig) ShowRedisWorkerWarning() bool {
	return c.HasModule(ModuleWorker) && c.HasModule(ModuleNoSQLDatastore) && c.NoSQLDatabase == DatabaseRedis
}

// Event Consumer Configuration Helpers

// UsesKafka returns true if Kafka is the selected message broker
func (c *ProjectConfig) UsesKafka() bool {
	return c.MessageBroker == BrokerKafka
}

// UsesRabbitMQ returns true if RabbitMQ is the selected message broker
func (c *ProjectConfig) UsesRabbitMQ() bool {
	return c.MessageBroker == BrokerRabbitMQ
}

// UsesSQS returns true if AWS SQS is the selected message broker
func (c *ProjectConfig) UsesSQS() bool {
	return c.MessageBroker == BrokerSQS
}

// UsesPubSub returns true if GCP Pub/Sub is the selected message broker
func (c *ProjectConfig) UsesPubSub() bool {
	return c.MessageBroker == BrokerPubSub
}

// EventConsumerNeedsDockerCompose returns true if EventConsumer needs docker-compose services
func (c *ProjectConfig) EventConsumerNeedsDockerCompose() bool {
	return c.HasModule(ModuleEventConsumer) && c.MessageBroker != ""
}

// AI Agent Configuration Helpers

// AIAgentInfo contains metadata about an AI coding agent
type AIAgentInfo struct {
	ID          string // Internal identifier (e.g., "claude", "cursor")
	Name        string // Display name (e.g., "Claude Code", "Cursor")
	FilePath    string // Output file path (e.g., "CLAUDE.md", ".cursor/rules/project.mdc")
	Description string // Short description for prompts
}

// GetAvailableAIAgents returns all supported AI coding agents
func GetAvailableAIAgents() []AIAgentInfo {
	return []AIAgentInfo{
		{ID: "claude", Name: "Claude Code", FilePath: "CLAUDE.md", Description: "Anthropic's Claude Code CLI"},
		{ID: "cursor", Name: "Cursor", FilePath: ".cursor/rules/project.mdc", Description: "AI-first code editor"},
		{ID: "copilot", Name: "GitHub Copilot", FilePath: ".github/copilot-instructions.md", Description: "GitHub's AI pair programmer"},
		{ID: "codex", Name: "Codex", FilePath: "AGENTS.md", Description: "OpenAI's software engineering agent"},
	}
}

// GetAIAgentIDs returns just the agent IDs for validation
func GetAIAgentIDs() []string {
	agents := GetAvailableAIAgents()
	ids := make([]string, len(agents))
	for i, a := range agents {
		ids[i] = a.ID
	}
	return ids
}

// GetAIAgentDisplayOptions returns formatted display strings for prompts
func GetAIAgentDisplayOptions() []string {
	agents := GetAvailableAIAgents()
	options := make([]string, len(agents))
	for i, a := range agents {
		options[i] = a.Name + " - " + a.Description
	}
	return options
}

// HasAIAgent checks if a specific AI agent is selected
func (c *ProjectConfig) HasAIAgent(id string) bool {
	for _, a := range c.AIAgents {
		if a == id {
			return true
		}
	}
	// Backwards compatibility: check legacy field for Claude
	if id == "claude" && c.IncludeCLAUDEMD {
		return true
	}
	return false
}

// HasAnyAIAgent returns true if any AI agent is selected
func (c *ProjectConfig) HasAnyAIAgent() bool {
	return len(c.AIAgents) > 0 || c.IncludeCLAUDEMD
}

// GetSelectedAIAgents returns info about all selected AI agents
func (c *ProjectConfig) GetSelectedAIAgents() []AIAgentInfo {
	allAgents := GetAvailableAIAgents()
	var selected []AIAgentInfo
	for _, agent := range allAgents {
		if c.HasAIAgent(agent.ID) {
			selected = append(selected, agent)
		}
	}
	return selected
}

// CI/CD Provider Configuration Helpers

// CIProviderInfo contains metadata about a CI/CD provider
type CIProviderInfo struct {
	ID          string
	Name        string
	Description string
}

// GetAvailableCIProviders returns all supported CI/CD providers
func GetAvailableCIProviders() []CIProviderInfo {
	return []CIProviderInfo{
		{ID: "github", Name: "GitHub Actions", Description: "CI/CD for GitHub repositories"},
	}
}

// GetCIProviderDisplayOptions returns formatted display strings for prompts
func GetCIProviderDisplayOptions() []string {
	providers := GetAvailableCIProviders()
	options := make([]string, len(providers))
	for i, p := range providers {
		options[i] = p.Name + " - " + p.Description
	}
	return options
}

// HasCIProvider checks if a specific CI provider is configured
func (c *ProjectConfig) HasCIProvider(id string) bool {
	return c.CIProvider == id
}

// HasAnyCIProvider returns true if any CI provider is configured
func (c *ProjectConfig) HasAnyCIProvider() bool {
	return c.CIProvider != ""
}

// ReviewEnabled returns true when any review scaffolding should be emitted.
// Mirrors Mode for convenience in templates.
func (c *ProjectConfig) ReviewEnabled() bool {
	return c.Review.Mode != ReviewModeOff && c.Review.Mode != ""
}

// ReviewEmitsStopHook returns true only when Mode=full — i.e., we wire the
// Stop hook guard that ensures the code-reviewer subagent is invoked.
func (c *ProjectConfig) ReviewEmitsStopHook() bool {
	return c.Review.Mode == ReviewModeFull
}

// Vector store constants
const (
	VectorStoreNone     = "none"
	VectorStorePgVector = "pgvector"
	VectorStoreQdrant   = "qdrant"
	VectorStoreMongoDB  = "mongodb"
)

// HasVectorStore returns true when the project should ship vector-store
// scaffolding (embedding starter, VectorKnowledgeRetriever, ingestion).
// AIAgent must be selected for the vector path to be useful — RAG without
// an agent has no consumer in any pattern Trabuco generates.
func (c *ProjectConfig) HasVectorStore() bool {
	if !c.HasModule(ModuleAIAgent) {
		return false
	}
	return c.VectorStore != "" && c.VectorStore != VectorStoreNone
}

// VectorStoreIsPgVector returns true when the PGVector flavor is selected.
// Implies a Postgres SQLDatastore is also present (the CLI auto-resolves
// this in Phase E).
func (c *ProjectConfig) VectorStoreIsPgVector() bool {
	return c.VectorStore == VectorStorePgVector
}

// VectorStoreIsQdrant returns true for the standalone Qdrant container.
func (c *ProjectConfig) VectorStoreIsQdrant() bool {
	return c.VectorStore == VectorStoreQdrant
}

// VectorStoreIsMongoAtlas returns true when the project uses MongoDB
// Atlas Vector Search. Note: community Mongo cannot serve $vectorSearch
// — generated code still works but only against Atlas.
func (c *ProjectConfig) VectorStoreIsMongoAtlas() bool {
	return c.VectorStore == VectorStoreMongoDB
}

// VectorStoreNeedsStandaloneMongoConnection returns true when the project
// uses MongoDB Atlas as its vector store but does NOT also have a
// NoSQLDatastore=mongodb wiring a Mongo connection. In that case the
// AIAgent application.yml needs its own spring.data.mongodb.uri so the
// Atlas vector-store starter has somewhere to connect; otherwise the
// connection comes from the NoSQL block above and a second emission
// would be a duplicate YAML key. (Phase E auto-adds NoSQLDatastore so
// most users never hit the standalone branch.)
func (c *ProjectConfig) VectorStoreNeedsStandaloneMongoConnection() bool {
	if !c.VectorStoreIsMongoAtlas() {
		return false
	}
	return !(c.HasModule(ModuleNoSQLDatastore) && c.NoSQLDatabase == DatabaseMongoDB)
}

// AuthEnabled returns true when auth scaffolding should be generated for
// the project. Auth code is emitted whenever a consuming module (API or
// AIAgent) is selected — the resulting files live alongside the rest of
// the project source so users can opt in at runtime by setting
// {@code trabuco.auth.enabled=true} (and configuring an OIDC issuer URI),
// without re-running the generator. When neither API nor AIAgent is
// present nothing in the generated project would consume the auth
// scaffolding, so we skip the emission entirely.
func (c *ProjectConfig) AuthEnabled() bool {
	return c.HasModule(ModuleAPI) || c.HasModule(ModuleAIAgent)
}

// ValidateVectorStoreFlag returns "" when the value is one of the
// recognized vector-store flavors (or empty), and an error message
// otherwise. Pure value validation — does not look at modules or
// other flags. Use ResolveVectorStore for the cross-flag rules.
func ValidateVectorStoreFlag(vectorStore string) string {
	switch vectorStore {
	case "", VectorStoreNone, VectorStorePgVector, VectorStoreQdrant, VectorStoreMongoDB:
		return ""
	}
	return "Invalid --vector-store value '" + vectorStore + "'. Valid options: pgvector, qdrant, mongodb, none"
}

// ResolveVectorStore enforces the cross-flag rules for the vector
// store and adjusts the config in-place when a default is unambiguous.
// Returns "" on success or a human-readable error message.
//
// Rules per backend:
//
//   - "" / "none": no-op.
//
//   - "pgvector": pgvector lives inside the application's Postgres
//     datastore. Forces Database=postgresql; auto-adds SQLDatastore
//     when missing. Errors when Database is already set to a different
//     SQL flavor, or when NoSQLDatastore is selected (mutually
//     exclusive with SQLDatastore).
//
//   - "qdrant": standalone server, no module/database constraints.
//
//   - "mongodb": MongoDB Atlas Vector Search. Coexists with
//     NoSQLDatastore=mongodb (auto-coerces NoSQLDatabase) or runs
//     standalone (Phase C VectorStoreNeedsStandaloneMongoConnection
//     wires its own connection). Errors when NoSQLDatastore is
//     present with a non-mongo NoSQLDatabase (e.g. redis).
//
// All non-empty backends require AIAgent — RAG without an agent has
// no consumer.
func (c *ProjectConfig) ResolveVectorStore() string {
	if c.VectorStore == "" || c.VectorStore == VectorStoreNone {
		return ""
	}

	if !c.HasModule(ModuleAIAgent) {
		return "--vector-store=" + c.VectorStore + " requires the AIAgent module — vector RAG has no consumer otherwise. Add AIAgent to --modules or drop --vector-store."
	}

	switch c.VectorStore {
	case VectorStorePgVector:
		if c.HasModule(ModuleNoSQLDatastore) {
			return "--vector-store=pgvector conflicts with the NoSQLDatastore module. Pgvector lives inside the application's Postgres SQLDatastore; pick one or the other."
		}
		// CLI default for --database is "postgresql", so an unset
		// value also lands here. Treat anything-other-than-postgresql
		// as an explicit user request that conflicts with pgvector.
		if c.Database != "" && c.Database != DatabasePostgreSQL {
			return "--vector-store=pgvector requires --database=postgresql, got '" + c.Database + "'. Pgvector lives in the application's Postgres datastore; switch to postgresql or pick a different vector store."
		}
		c.Database = DatabasePostgreSQL
		if !c.HasModule(ModuleSQLDatastore) {
			c.Modules = ResolveDependencies(append(c.Modules, ModuleSQLDatastore))
		}

	case VectorStoreMongoDB:
		if c.HasModule(ModuleNoSQLDatastore) && c.NoSQLDatabase != "" && c.NoSQLDatabase != DatabaseMongoDB {
			return "--vector-store=mongodb requires --nosql-database=mongodb, got '" + c.NoSQLDatabase + "'. Atlas Vector Search is only available on MongoDB; switch to mongodb or pick a different vector store."
		}
		if c.HasModule(ModuleNoSQLDatastore) {
			c.NoSQLDatabase = DatabaseMongoDB
		}

	case VectorStoreQdrant:
		// No constraints. Qdrant is its own server.
	}

	return ""
}
