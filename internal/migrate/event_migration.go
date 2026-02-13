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

// Event migration prompts
const eventSystemPrompt = `You are an expert Java developer specializing in Spring Boot migrations.
You convert legacy event listeners to Trabuco's EventConsumer module format.

TRABUCO PATTERNS:
- Uses Kafka or RabbitMQ for message consumption (depending on project config)
- Event classes are simple POJOs for message payloads
- Listeners use @KafkaListener or @RabbitListener annotations
- Uses constructor injection
- Includes error handling and dead letter queue patterns
- Events should be placed in the Model module's events package`

const eventPromptTemplateKafka = `Convert this event listener to Trabuco EventConsumer module format using Kafka.

SOURCE LISTENER:
%s

TARGET PACKAGES:
- Event class: %s.model.events
- Listener class: %s.eventconsumer.listener

OUTPUT FORMAT (JSON):
{
  "name": "ListenerName",
  "event_code": "// Full Java code for the Event class (goes in Model module)",
  "listener_code": "// Full Java code for the Kafka Listener class",
  "notes": ["Any migration notes"],
  "requires_review": false,
  "review_reason": ""
}`

const eventPromptTemplateRabbitMQ = `Convert this event listener to Trabuco EventConsumer module format using RabbitMQ.

SOURCE LISTENER:
%s

TARGET PACKAGES:
- Event class: %s.model.events
- Listener class: %s.eventconsumer.listener

OUTPUT FORMAT (JSON):
{
  "name": "ListenerName",
  "event_code": "// Full Java code for the Event class (goes in Model module)",
  "listener_code": "// Full Java code for the RabbitMQ Listener class",
  "notes": ["Any migration notes"],
  "requires_review": false,
  "review_reason": ""
}`

// EventMigrationResult contains the AI's event conversion output
type EventMigrationResult struct {
	Name           string   `json:"name"`
	EventCode      string   `json:"event_code"`
	ListenerCode   string   `json:"listener_code"`
	Notes          []string `json:"notes"`
	RequiresReview bool     `json:"requires_review"`
	ReviewReason   string   `json:"review_reason"`
}

// convertEventListener uses AI to convert an event listener to Trabuco format
func (m *Migrator) convertEventListener(ctx context.Context, listener *JavaClass) (*ConvertedEventListener, error) {
	// Determine message broker type (default to Kafka)
	promptTemplate := eventPromptTemplateKafka
	if m.projectInfo.UsesRabbitMQ {
		promptTemplate = eventPromptTemplateRabbitMQ
	}

	prompt := fmt.Sprintf(promptTemplate, listener.Content, m.projectInfo.GroupID, m.projectInfo.GroupID)

	response, err := m.provider.Analyze(ctx, &ai.AnalysisRequest{
		SystemPrompt: eventSystemPrompt,
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
		Stage:           StageEvents,
		File:            listener.FilePath,
		Action:          "convert",
		PromptSummary:   fmt.Sprintf("Convert event listener %s", listener.Name),
		ResponseSummary: truncate(response.Content, 200),
		TokensUsed: TokenUsage{
			InputTokens:  response.InputTokens,
			OutputTokens: response.OutputTokens,
		},
	})

	// Update cost in checkpoint (for backward compatibility)
	cost := m.provider.EstimateCost(response.InputTokens, response.OutputTokens)
	m.checkpoint.UpdateCost(cost)

	result, err := parseEventResult(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &ConvertedEventListener{
		Name:         result.Name,
		EventCode:    result.EventCode,
		ListenerCode: result.ListenerCode,
	}, nil
}

// writeEventListenerFiles writes the converted event listener files
func (m *Migrator) writeEventListenerFiles(listener *ConvertedEventListener) error {
	packagePath := strings.ReplaceAll(m.projectInfo.GroupID, ".", string(filepath.Separator))

	// Write event class to Model module
	if listener.EventCode != "" {
		eventDir := filepath.Join(m.config.OutputPath, config.ModuleModel, "src", "main", "java",
			packagePath, "model", "events")
		if err := os.MkdirAll(eventDir, 0755); err != nil {
			return err
		}

		eventPath := filepath.Join(eventDir, listener.Name+"Event.java")
		if err := os.WriteFile(eventPath, []byte(listener.EventCode), 0644); err != nil {
			return err
		}
	}

	// Write listener class to EventConsumer module
	listenerDir := filepath.Join(m.config.OutputPath, config.ModuleEventConsumer, "src", "main", "java",
		packagePath, "eventconsumer", "listener")
	if err := os.MkdirAll(listenerDir, 0755); err != nil {
		return err
	}

	listenerPath := filepath.Join(listenerDir, listener.Name+"Listener.java")
	return os.WriteFile(listenerPath, []byte(listener.ListenerCode), 0644)
}

func parseEventResult(content string) (*EventMigrationResult, error) {
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var result EventMigrationResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &result, nil
}
