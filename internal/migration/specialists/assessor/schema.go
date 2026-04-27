// Package assessor implements the Phase 0 specialist. It runs first,
// scans the source repo, and produces .trabuco-migration/assessment.json —
// the no-out-of-scope contract that every later specialist reads to know
// what artifacts exist.
package assessor

import (
	"encoding/json"
	"os"
	"strings"
)

// Assessment is the structured catalog the assessor writes to
// .trabuco-migration/assessment.json. Every later specialist reads this;
// no specialist may touch source artifacts not listed here.
type Assessment struct {
	SchemaVersion int    `json:"schemaVersion"`
	GeneratedAt   string `json:"generatedAt"`

	// Source-level context
	BuildSystem       string   `json:"buildSystem"`     // maven | gradle | other
	Framework         string   `json:"framework"`       // spring-boot-2.x | spring-boot-3.x | quarkus | micronaut | jaxrs | servlet | non-spring | mixed
	JavaVersion       string   `json:"javaVersion"`
	IsMultiModule     bool     `json:"isMultiModule"`
	ModulePaths       []string `json:"modulePaths,omitempty"`
	HasFrontendInRepo bool     `json:"hasFrontendInRepo"`
	HasNonJVMCode     bool     `json:"hasNonJvmCode"`

	// Persistence
	Persistence  string         `json:"persistence"` // jpa | spring-data-jdbc | jdbc-template | mybatis | mongodb | redis | none | mixed
	Entities     []EntityInfo   `json:"entities,omitempty"`
	Repositories []RepoInfo     `json:"repositories,omitempty"`
	MigrationsDir string        `json:"migrationsDir,omitempty"` // Flyway/Liquibase location

	// Web layer
	WebLayer    string           `json:"webLayer"` // spring-mvc | webflux | jaxrs | none
	Controllers []ControllerInfo `json:"controllers,omitempty"`

	// Service layer
	Services []ServiceInfo `json:"services,omitempty"`

	// Async / scheduled
	AsyncFramework string     `json:"asyncFramework"` // scheduled-annotation | quartz | jobrunr | other | none
	Jobs           []JobInfo  `json:"jobs,omitempty"`

	// Messaging
	Messaging  string          `json:"messaging"` // kafka | rabbitmq | sqs | pubsub | jms | none | mixed
	Listeners  []ListenerInfo  `json:"listeners,omitempty"`
	Publishers []PublisherInfo `json:"publishers,omitempty"`

	// AI / LLM integration
	HasAIIntegration bool   `json:"hasAiIntegration"`
	AIFramework      string `json:"aiFramework,omitempty"` // spring-ai | langchain4j | bespoke | none

	// CI / CD
	CISystems       []CIInfo         `json:"ciSystems,omitempty"` // empty if no CI/CD detected
	DeploymentFiles []DeploymentFile `json:"deploymentFiles,omitempty"`

	// Tests
	TestFramework string     `json:"testFramework"` // junit-4 | junit-5 | spock | testng | mixed
	Tests         []TestInfo `json:"tests,omitempty"`

	// Configuration
	ConfigFormat   string   `json:"configFormat"` // yaml | properties | mixed
	ConfigFiles    []string `json:"configFiles,omitempty"`
	SpringProfiles []string `json:"springProfiles,omitempty"`

	// Secrets / sensitive findings
	SecretsInSource []string `json:"secretsInSource,omitempty"` // file:line of suspected hardcoded credentials

	// Recommended target config (assessor's suggestion based on findings)
	RecommendedTarget RecommendedTarget `json:"recommendedTarget"`

	// Feasibility verdict
	Feasibility   string   `json:"feasibility"`           // green | yellow | red
	BlockerCodes  []string `json:"blockerCodes,omitempty"` // top-level blockers requiring user decision
	Notes         []string `json:"notes,omitempty"`
}

// EntityInfo catalogs one persistent entity in the source.
type EntityInfo struct {
	File             string   `json:"file"`
	ClassName        string   `json:"className"`
	TableName        string   `json:"tableName,omitempty"`
	IsJPA            bool     `json:"isJpa"`
	IsDocument       bool     `json:"isDocument"`
	HasFK            bool     `json:"hasFk"`
	HasCompositePK   bool     `json:"hasCompositePk"`
	UsesEntityGraph  bool     `json:"usesEntityGraph"`
	Aggregate        string   `json:"aggregate,omitempty"` // grouping for vertical-slice migration
}

// RepoInfo catalogs one repository / DAO.
type RepoInfo struct {
	File           string `json:"file"`
	ClassName      string `json:"className"`
	Style          string `json:"style"` // spring-data | jdbc-template | mongo-template | mybatis | bespoke
	UsesPagination bool   `json:"usesPagination"`
	PaginationKind string `json:"paginationKind,omitempty"` // offset | keyset | none
}

// ControllerInfo catalogs one REST controller.
type ControllerInfo struct {
	File           string         `json:"file"`
	ClassName      string         `json:"className"`
	BasePath       string         `json:"basePath,omitempty"`
	Endpoints      []EndpointInfo `json:"endpoints,omitempty"`
	UsesValidation bool           `json:"usesValidation"`
	ErrorEnvelope  string         `json:"errorEnvelope,omitempty"` // bespoke | rfc7807 | none
}

// EndpointInfo catalogs one HTTP endpoint.
type EndpointInfo struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

// ServiceInfo catalogs one service / business-logic class.
type ServiceInfo struct {
	File             string `json:"file"`
	ClassName        string `json:"className"`
	UsesFieldInject  bool   `json:"usesFieldInjection"`
	HasStaticState   bool   `json:"hasStaticState"`
	UsesAppContext   bool   `json:"usesAppContextLookup"`
	UsesServiceLoader bool  `json:"usesServiceLoader"`
}

// JobInfo catalogs one scheduled / async job.
type JobInfo struct {
	File      string `json:"file"`
	ClassName string `json:"className"`
	Kind      string `json:"kind"` // scheduled | async | quartz | other
	Cron      string `json:"cron,omitempty"`
}

// ListenerInfo catalogs one message listener.
type ListenerInfo struct {
	File      string `json:"file"`
	ClassName string `json:"className"`
	Broker    string `json:"broker"` // kafka | rabbitmq | sqs | pubsub | jms
	Topic     string `json:"topic,omitempty"`
}

// PublisherInfo catalogs one message publisher.
type PublisherInfo struct {
	File      string `json:"file"`
	ClassName string `json:"className"`
	Broker    string `json:"broker"`
	Topic     string `json:"topic,omitempty"`
}

// CIInfo catalogs one CI/CD system file.
type CIInfo struct {
	System string   `json:"system"` // github-actions | gitlab-ci | jenkins | circleci | azure-pipelines | travis | argo | flux | helm | k8s | terraform
	Files  []string `json:"files"`  // file paths relative to repo root
}

// DeploymentFile catalogs deployment infrastructure files (Dockerfiles,
// compose files, etc.) separate from CI workflow files.
type DeploymentFile struct {
	File string `json:"file"`
	Kind string `json:"kind"` // dockerfile | compose-prod | helm-chart | k8s-manifest | terraform | sam | cdk
}

// TestInfo catalogs one test class. The test specialist (Phase 11) will
// later annotate each with KEEP / ADAPT / DISCARD / CHARACTERIZE-FIRST.
type TestInfo struct {
	File           string `json:"file"`
	ClassName      string `json:"className"`
	Style          string `json:"style"`           // springboot-test | webmvc-test | datajdbc-test | unit | spock | other
	UsesPowerMock  bool   `json:"usesPowerMock"`
	UsesH2         bool   `json:"usesH2"`
	UsesTestcontainers bool `json:"usesTestcontainers"`
}

// RecommendedTarget captures the assessor's suggested Trabuco target config.
// The user can override before Phase 1 runs.
type RecommendedTarget struct {
	Modules       []string `json:"modules"`
	Database      string   `json:"database,omitempty"`
	NoSQLDatabase string   `json:"nosqlDatabase,omitempty"`
	MessageBroker string   `json:"messageBroker,omitempty"`
	AIAgents      []string `json:"aiAgents,omitempty"`
	CIProvider    string   `json:"ciProvider,omitempty"`
	JavaVersion   string   `json:"javaVersion"`
	Notes         []string `json:"notes,omitempty"`
}

// Save writes the assessment to .trabuco-migration/assessment.json.
func Save(path string, a *Assessment) error {
	a.SchemaVersion = 1
	data, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// PrefixSourcePaths rewrites every file path inside the assessment that
// points into the original source tree (i.e. starts with "src/") so it
// is prefixed with `prefix` (typically "legacy/"). Called after Phase 1
// moves the user's `src/` into `legacy/src/` so downstream specialists'
// source_evidence resolves to the new locations.
//
// Paths that don't start with "src/" — CI workflows under .github/,
// deployment files at repo root, etc. — are left untouched because
// Phase 1 doesn't move them.
func (a *Assessment) PrefixSourcePaths(prefix string) {
	prefixIfSrc := func(p string) string {
		if strings.HasPrefix(p, "src/") {
			return prefix + p
		}
		return p
	}
	for i := range a.Entities {
		a.Entities[i].File = prefixIfSrc(a.Entities[i].File)
	}
	for i := range a.Repositories {
		a.Repositories[i].File = prefixIfSrc(a.Repositories[i].File)
	}
	for i := range a.Controllers {
		a.Controllers[i].File = prefixIfSrc(a.Controllers[i].File)
	}
	for i := range a.Services {
		a.Services[i].File = prefixIfSrc(a.Services[i].File)
	}
	for i := range a.Jobs {
		a.Jobs[i].File = prefixIfSrc(a.Jobs[i].File)
	}
	for i := range a.Listeners {
		a.Listeners[i].File = prefixIfSrc(a.Listeners[i].File)
	}
	for i := range a.Publishers {
		a.Publishers[i].File = prefixIfSrc(a.Publishers[i].File)
	}
	for i := range a.Tests {
		a.Tests[i].File = prefixIfSrc(a.Tests[i].File)
	}
	for i, p := range a.ConfigFiles {
		a.ConfigFiles[i] = prefixIfSrc(p)
	}
	// secretsInSource entries are "file:line" — split, prefix, rejoin.
	for i, sec := range a.SecretsInSource {
		if idx := strings.IndexByte(sec, ':'); idx > 0 {
			a.SecretsInSource[i] = prefixIfSrc(sec[:idx]) + sec[idx:]
		} else {
			a.SecretsInSource[i] = prefixIfSrc(sec)
		}
	}
}

// Load reads an assessment from disk. Used by other specialists to
// understand what's in scope.
func Load(path string) (*Assessment, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var a Assessment
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	return &a, nil
}
