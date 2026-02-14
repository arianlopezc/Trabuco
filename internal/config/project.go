package config

import "github.com/arianlopezc/Trabuco/internal/utils"

// ProjectConfig holds all configuration for a generated project
type ProjectConfig struct {
	// Basic info
	ProjectName string // e.g., "my-platform"
	GroupID     string // e.g., "com.company.project"
	ArtifactID  string // e.g., "my-platform" (usually same as ProjectName)

	// Java
	JavaVersion         string // "17", "21", or "25"
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
	AIAgents []string // Selected agents: "claude", "cursor", "copilot", "windsurf", "cline"

	// Deprecated: Use AIAgents instead
	IncludeCLAUDEMD bool // Legacy field for backwards compatibility
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
	FilePath    string // Output file path (e.g., "CLAUDE.md", ".cursorrules")
	Description string // Short description for prompts
}

// GetAvailableAIAgents returns all supported AI coding agents
func GetAvailableAIAgents() []AIAgentInfo {
	return []AIAgentInfo{
		{ID: "claude", Name: "Claude Code", FilePath: "CLAUDE.md", Description: "Anthropic's Claude Code CLI"},
		{ID: "cursor", Name: "Cursor", FilePath: ".cursorrules", Description: "AI-first code editor"},
		{ID: "copilot", Name: "GitHub Copilot", FilePath: ".github/copilot-instructions.md", Description: "GitHub's AI pair programmer"},
		{ID: "windsurf", Name: "Windsurf", FilePath: ".windsurf/rules/project.md", Description: "Codeium's agentic IDE"},
		{ID: "cline", Name: "Cline", FilePath: ".clinerules/project.md", Description: "VS Code autonomous agent"},
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
