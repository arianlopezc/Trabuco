package migrate

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/config"
)

// Repository migration prompts
const repositorySystemPrompt = `You are an expert Java developer specializing in Spring Boot migrations.
You convert legacy repository interfaces to Trabuco's datastore module format.

TRABUCO PATTERNS:
- Uses Spring Data JDBC (NOT JPA) for SQL databases
- Uses Spring Data MongoDB for NoSQL
- Repository interfaces extend CrudRepository or custom interfaces
- Custom queries use @Query annotation with native SQL or MongoDB queries
- No JPA-specific features (EntityManager, JPQL, etc.)`

const repositoryPromptTemplate = `Convert this Java repository to Trabuco format.

SOURCE REPOSITORY:
%s

DATABASE TYPE: %s
TARGET PACKAGE: %s

OUTPUT FORMAT (JSON):
{
  "name": "RepositoryName",
  "code": "// Full Java code for the repository interface",
  "notes": ["Any migration notes"],
  "requires_review": false,
  "review_reason": ""
}`

// Service migration prompts
const serviceSystemPrompt = `You are an expert Java developer specializing in Spring Boot migrations.
You convert legacy service classes to Trabuco's Shared module format.

TRABUCO PATTERNS:
- Services are in the Shared module
- Uses constructor injection (no @Autowired on fields)
- Circuit breaker patterns with Resilience4j
- Clear separation between business logic and data access`

const servicePromptTemplate = `Convert this Java service to Trabuco format.

SOURCE SERVICE:
%s

TARGET PACKAGE: %s.shared.service

OUTPUT FORMAT (JSON):
{
  "name": "ServiceName",
  "code": "// Full Java code for the service class",
  "test_code": "// Full Java code for the service test class",
  "notes": ["Any migration notes"],
  "requires_review": false,
  "review_reason": ""
}`

// RepositoryMigrationResult contains the AI's repository conversion output
type RepositoryMigrationResult struct {
	Name           string   `json:"name"`
	Code           string   `json:"code"`
	Notes          []string `json:"notes"`
	RequiresReview bool     `json:"requires_review"`
	ReviewReason   string   `json:"review_reason"`
}

// ServiceMigrationResult contains the AI's service conversion output
type ServiceMigrationResult struct {
	Name           string   `json:"name"`
	Code           string   `json:"code"`
	TestCode       string   `json:"test_code"`
	Notes          []string `json:"notes"`
	RequiresReview bool     `json:"requires_review"`
	ReviewReason   string   `json:"review_reason"`
}

// convertRepository uses AI to convert a repository to Trabuco format
func (m *Migrator) convertRepository(ctx context.Context, repo *JavaClass) (*ConvertedRepository, error) {
	dbType := "SQL (Spring Data JDBC)"
	if m.projectInfo.UsesNoSQL {
		dbType = "NoSQL (MongoDB)"
	}

	targetPackage := m.projectInfo.GroupID
	if m.projectInfo.UsesNoSQL {
		targetPackage += ".nosqldatastore.repository"
	} else {
		targetPackage += ".sqldatastore.repository"
	}

	prompt := fmt.Sprintf(repositoryPromptTemplate, repo.Content, dbType, targetPackage)

	response, err := m.provider.Analyze(ctx, &ai.AnalysisRequest{
		SystemPrompt: repositorySystemPrompt,
		UserPrompt:   prompt,
		MaxTokens:    8192,
		Temperature:  0.1,
	})

	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Record usage in cost tracker
	m.costTracker.RecordFromResponse(response)

	m.checkpoint.AddAIDecision(AIDecision{
		Stage:           StageRepositories,
		File:            repo.FilePath,
		Action:          "convert",
		PromptSummary:   fmt.Sprintf("Convert repository %s", repo.Name),
		ResponseSummary: truncate(response.Content, 200),
		TokensUsed: TokenUsage{
			InputTokens:  response.InputTokens,
			OutputTokens: response.OutputTokens,
		},
	})

	// Update cost in checkpoint (for backward compatibility)
	cost := m.provider.EstimateCost(response.InputTokens, response.OutputTokens)
	m.checkpoint.UpdateCost(cost)

	result, err := parseRepositoryResult(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &ConvertedRepository{
		Name:           result.Name,
		RepositoryCode: result.Code,
	}, nil
}

// writeRepositoryFiles writes the converted repository files
func (m *Migrator) writeRepositoryFiles(repo *ConvertedRepository, module string) error {
	packagePath := strings.ReplaceAll(m.projectInfo.GroupID, ".", string(filepath.Separator))

	var subPackage string
	if module == config.ModuleSQLDatastore {
		subPackage = "sqldatastore"
	} else {
		subPackage = "nosqldatastore"
	}

	repoDir := filepath.Join(m.config.OutputPath, module, "src", "main", "java",
		packagePath, subPackage, "repository")

	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return err
	}

	repoPath := filepath.Join(repoDir, repo.Name+".java")
	return os.WriteFile(repoPath, []byte(repo.RepositoryCode), 0644)
}

// convertService uses AI to convert a service to Trabuco format
func (m *Migrator) convertService(ctx context.Context, service *JavaClass) (*ConvertedService, error) {
	prompt := fmt.Sprintf(servicePromptTemplate, service.Content, m.projectInfo.GroupID)

	response, err := m.provider.Analyze(ctx, &ai.AnalysisRequest{
		SystemPrompt: serviceSystemPrompt,
		UserPrompt:   prompt,
		MaxTokens:    16384,
		Temperature:  0.1,
	})

	if err != nil {
		return nil, fmt.Errorf("AI analysis failed: %w", err)
	}

	// Record usage in cost tracker
	m.costTracker.RecordFromResponse(response)

	m.checkpoint.AddAIDecision(AIDecision{
		Stage:           StageServices,
		File:            service.FilePath,
		Action:          "convert",
		PromptSummary:   fmt.Sprintf("Convert service %s", service.Name),
		ResponseSummary: truncate(response.Content, 200),
		TokensUsed: TokenUsage{
			InputTokens:  response.InputTokens,
			OutputTokens: response.OutputTokens,
		},
	})

	// Update cost in checkpoint (for backward compatibility)
	cost := m.provider.EstimateCost(response.InputTokens, response.OutputTokens)
	m.checkpoint.UpdateCost(cost)

	result, err := parseServiceResult(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &ConvertedService{
		Name:        result.Name,
		ServiceCode: result.Code,
		TestCode:    result.TestCode,
	}, nil
}

// writeServiceFiles writes the converted service files
func (m *Migrator) writeServiceFiles(service *ConvertedService) error {
	packagePath := strings.ReplaceAll(m.projectInfo.GroupID, ".", string(filepath.Separator))

	// Write service
	serviceDir := filepath.Join(m.config.OutputPath, config.ModuleShared, "src", "main", "java",
		packagePath, "shared", "service")
	if err := os.MkdirAll(serviceDir, 0755); err != nil {
		return err
	}

	servicePath := filepath.Join(serviceDir, service.Name+".java")
	if err := os.WriteFile(servicePath, []byte(service.ServiceCode), 0644); err != nil {
		return err
	}

	// Write test if provided
	if service.TestCode != "" {
		testDir := filepath.Join(m.config.OutputPath, config.ModuleShared, "src", "test", "java",
			packagePath, "shared", "service")
		if err := os.MkdirAll(testDir, 0755); err != nil {
			return err
		}

		testPath := filepath.Join(testDir, service.Name+"Test.java")
		if err := os.WriteFile(testPath, []byte(service.TestCode), 0644); err != nil {
			return err
		}
	}

	return nil
}

func parseRepositoryResult(content string) (*RepositoryMigrationResult, error) {
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var result RepositoryMigrationResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}

func parseServiceResult(content string) (*ServiceMigrationResult, error) {
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var result ServiceMigrationResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}
