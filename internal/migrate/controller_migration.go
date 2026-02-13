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

// Controller migration prompts
const controllerSystemPrompt = `You are an expert Java developer specializing in Spring Boot migrations.
You convert legacy REST controllers to Trabuco's API module format.

TRABUCO PATTERNS:
- Controllers are in the API module
- Uses @RestController with @RequestMapping
- Request validation with @Valid
- Consistent error handling via GlobalExceptionHandler
- OpenAPI annotations for documentation
- Uses constructor injection`

const controllerPromptTemplate = `Convert this Java controller to Trabuco format.

SOURCE CONTROLLER:
%s

TARGET PACKAGE: %s.api.controller

OUTPUT FORMAT (JSON):
{
  "name": "ControllerName",
  "code": "// Full Java code for the controller class",
  "notes": ["Any migration notes"],
  "requires_review": false,
  "review_reason": ""
}`

// ControllerMigrationResult contains the AI's controller conversion output
type ControllerMigrationResult struct {
	Name           string   `json:"name"`
	Code           string   `json:"code"`
	Notes          []string `json:"notes"`
	RequiresReview bool     `json:"requires_review"`
	ReviewReason   string   `json:"review_reason"`
}

// convertController uses AI to convert a controller to Trabuco format
func (m *Migrator) convertController(ctx context.Context, controller *JavaClass) (*ConvertedController, error) {
	prompt := fmt.Sprintf(controllerPromptTemplate, controller.Content, m.projectInfo.GroupID)

	response, err := m.provider.Analyze(ctx, &ai.AnalysisRequest{
		SystemPrompt: controllerSystemPrompt,
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
		Stage:           StageControllers,
		File:            controller.FilePath,
		Action:          "convert",
		PromptSummary:   fmt.Sprintf("Convert controller %s", controller.Name),
		ResponseSummary: truncate(response.Content, 200),
		TokensUsed: TokenUsage{
			InputTokens:  response.InputTokens,
			OutputTokens: response.OutputTokens,
		},
	})

	// Update cost in checkpoint (for backward compatibility)
	cost := m.provider.EstimateCost(response.InputTokens, response.OutputTokens)
	m.checkpoint.UpdateCost(cost)

	result, err := parseControllerResult(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &ConvertedController{
		Name:           result.Name,
		ControllerCode: result.Code,
	}, nil
}

// writeControllerFiles writes the converted controller files
func (m *Migrator) writeControllerFiles(controller *ConvertedController) error {
	packagePath := strings.ReplaceAll(m.projectInfo.GroupID, ".", string(filepath.Separator))

	controllerDir := filepath.Join(m.config.OutputPath, config.ModuleAPI, "src", "main", "java",
		packagePath, "api", "controller")

	if err := os.MkdirAll(controllerDir, 0755); err != nil {
		return err
	}

	controllerPath := filepath.Join(controllerDir, controller.Name+".java")
	return os.WriteFile(controllerPath, []byte(controller.ControllerCode), 0644)
}

func parseControllerResult(content string) (*ControllerMigrationResult, error) {
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var result ControllerMigrationResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}
