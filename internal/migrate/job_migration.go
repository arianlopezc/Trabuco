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

// Job migration prompts
const jobSystemPrompt = `You are an expert Java developer specializing in Spring Boot migrations.
You convert legacy scheduled jobs (Spring @Scheduled, Quartz, etc.) to Trabuco's Worker module format using JobRunr.

TRABUCO PATTERNS:
- Uses JobRunr for background job processing
- Jobs are split into JobRequest (data class) and JobHandler (processing logic)
- JobRequest implements org.jobrunr.jobs.lambdas.JobRequest
- JobHandler implements org.jobrunr.jobs.lambdas.JobRequestHandler
- Uses constructor injection
- Scheduling is done via configuration, not annotations`

const jobPromptTemplate = `Convert this scheduled job to Trabuco Worker module format using JobRunr.

SOURCE JOB:
%s

TARGET PACKAGE: %s.worker

OUTPUT FORMAT (JSON):
{
  "name": "JobName",
  "job_request_code": "// Full Java code for the JobRequest class",
  "job_handler_code": "// Full Java code for the JobHandler class",
  "notes": ["Any migration notes"],
  "requires_review": false,
  "review_reason": ""
}`

// JobMigrationResult contains the AI's job conversion output
type JobMigrationResult struct {
	Name           string   `json:"name"`
	JobRequestCode string   `json:"job_request_code"`
	JobHandlerCode string   `json:"job_handler_code"`
	Notes          []string `json:"notes"`
	RequiresReview bool     `json:"requires_review"`
	ReviewReason   string   `json:"review_reason"`
}

// convertJob uses AI to convert a scheduled job to Trabuco format
func (m *Migrator) convertJob(ctx context.Context, job *JavaClass) (*ConvertedJob, error) {
	prompt := fmt.Sprintf(jobPromptTemplate, job.Content, m.projectInfo.GroupID)

	response, err := m.provider.Analyze(ctx, &ai.AnalysisRequest{
		SystemPrompt: jobSystemPrompt,
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
		Stage:           StageJobs,
		File:            job.FilePath,
		Action:          "convert",
		PromptSummary:   fmt.Sprintf("Convert job %s", job.Name),
		ResponseSummary: truncate(response.Content, 200),
		TokensUsed: TokenUsage{
			InputTokens:  response.InputTokens,
			OutputTokens: response.OutputTokens,
		},
	})

	// Update cost in checkpoint (for backward compatibility)
	cost := m.provider.EstimateCost(response.InputTokens, response.OutputTokens)
	m.checkpoint.UpdateCost(cost)

	result, err := parseJobResult(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &ConvertedJob{
		Name:           result.Name,
		JobRequestCode: result.JobRequestCode,
		JobHandlerCode: result.JobHandlerCode,
	}, nil
}

// writeJobFiles writes the converted job files
func (m *Migrator) writeJobFiles(job *ConvertedJob) error {
	packagePath := strings.ReplaceAll(m.projectInfo.GroupID, ".", string(filepath.Separator))

	// Write job request
	requestDir := filepath.Join(m.config.OutputPath, config.ModuleWorker, "src", "main", "java",
		packagePath, "worker", "job")
	if err := os.MkdirAll(requestDir, 0755); err != nil {
		return err
	}

	requestPath := filepath.Join(requestDir, job.Name+"Request.java")
	if err := os.WriteFile(requestPath, []byte(job.JobRequestCode), 0644); err != nil {
		return err
	}

	// Write job handler
	handlerDir := filepath.Join(m.config.OutputPath, config.ModuleWorker, "src", "main", "java",
		packagePath, "worker", "handler")
	if err := os.MkdirAll(handlerDir, 0755); err != nil {
		return err
	}

	handlerPath := filepath.Join(handlerDir, job.Name+"Handler.java")
	return os.WriteFile(handlerPath, []byte(job.JobHandlerCode), 0644)
}

func parseJobResult(content string) (*JobMigrationResult, error) {
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var result JobMigrationResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}
