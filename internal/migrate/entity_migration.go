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

// Entity migration prompts
const entitySystemPrompt = `You are an expert Java developer specializing in Spring Boot migrations.
You convert legacy Java entities to Trabuco's Model module format.

TRABUCO PATTERNS:
- Uses Spring Data JDBC (NOT JPA)
- Entities are simple POJOs with @Table annotation
- No lazy loading, no @OneToMany/@ManyToOne
- Foreign keys are explicit fields (e.g., Long userId, not User user)
- DTOs use Immutables-style builders
- Records are preferred for simple value objects

IMPORTANT:
- Remove Lombok annotations and generate explicit getters/setters
- Convert @Entity to @Table
- Convert @Id to Spring Data JDBC @Id
- Remove JPA relationship annotations
- Convert @Column to simple fields (Spring Data JDBC infers column names)
- Generate corresponding DTO and Response classes`

const entityPromptTemplate = `Convert this Java entity to Trabuco's Model module format.

SOURCE ENTITY:
%s

PACKAGE INFO:
- Source package: %s
- Target package: %s.model

IMPORTANT: If this is NOT a database entity (e.g., it's an annotation, interface, abstract base class,
utility class, or configuration class), set "skip": true and explain why in "skip_reason".

OUTPUT FORMAT (JSON only, no markdown):
{
  "skip": false,
  "skip_reason": "",
  "entity": {
    "name": "ClassName",
    "code": "// Full Java code for the entity class"
  },
  "dto": {
    "name": "ClassNameRequest",
    "code": "// Full Java code for the request DTO"
  },
  "response": {
    "name": "ClassNameResponse",
    "code": "// Full Java code for the response DTO"
  },
  "flyway_migration": "-- SQL DDL for creating the table",
  "notes": ["Any migration notes or warnings"],
  "requires_review": false,
  "review_reason": ""
}

CRITICAL: Return ONLY valid JSON. Do NOT wrap the response in markdown code blocks.
Generate complete, compilable Java code. Include all necessary imports.`

// EntityMigrationResult contains the AI's entity conversion output
type EntityMigrationResult struct {
	Entity struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"entity"`
	DTO struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"dto"`
	Response struct {
		Name string `json:"name"`
		Code string `json:"code"`
	} `json:"response"`
	FlywayMigration string   `json:"flyway_migration"`
	Notes           []string `json:"notes"`
	RequiresReview  bool     `json:"requires_review"`
	ReviewReason    string   `json:"review_reason"`
	Skip            bool     `json:"skip"`           // True if this is not a real entity (annotation, etc.)
	SkipReason      string   `json:"skip_reason"`    // Reason for skipping
}

// convertEntity uses AI to convert an entity to Trabuco format
func (m *Migrator) convertEntity(ctx context.Context, entity *JavaClass) (*ConvertedEntity, error) {
	// Build the prompt
	prompt := fmt.Sprintf(entityPromptTemplate,
		entity.Content,
		entity.Package,
		m.projectInfo.GroupID,
	)

	// Start with a reasonable token limit, retry with more if truncated
	maxTokens := 16384

	var response *ai.AnalysisResponse
	var err error

	for attempt := 0; attempt < 2; attempt++ {
		response, err = m.provider.Analyze(ctx, &ai.AnalysisRequest{
			SystemPrompt: entitySystemPrompt,
			UserPrompt:   prompt,
			MaxTokens:    maxTokens,
			Temperature:  0.1, // Low temperature for deterministic output
		})

		if err != nil {
			return nil, fmt.Errorf("AI analysis failed: %w", err)
		}

		// Check if response was truncated
		if response.StopReason == "max_tokens" && attempt == 0 {
			// Retry with higher token limit
			maxTokens = 32000
			if m.config.Verbose {
				m.yellow.Printf("    Response truncated, retrying with %d tokens...\n", maxTokens)
			}
			continue
		}
		break
	}

	// Final check for truncation
	if response.StopReason == "max_tokens" {
		return nil, fmt.Errorf("response truncated even with %d tokens - entity may be too complex", maxTokens)
	}

	// Record usage in cost tracker
	m.costTracker.RecordFromResponse(response)

	// Record the AI decision
	m.checkpoint.AddAIDecision(AIDecision{
		Stage:           StageEntities,
		File:            entity.FilePath,
		Action:          "convert",
		PromptSummary:   fmt.Sprintf("Convert entity %s to Trabuco format", entity.Name),
		ResponseSummary: truncate(response.Content, 200),
		TokensUsed: TokenUsage{
			InputTokens:  response.InputTokens,
			OutputTokens: response.OutputTokens,
		},
	})

	// Update cost in checkpoint (for backward compatibility)
	cost := m.provider.EstimateCost(response.InputTokens, response.OutputTokens)
	m.checkpoint.UpdateCost(cost)

	// Parse the JSON response
	result, err := parseEntityMigrationResult(response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Check if this should be skipped (annotation, interface, etc.)
	if result.Skip || result.Entity.Name == "" || result.Entity.Name == "N/A" ||
	   strings.HasPrefix(result.Entity.Name, "N/A") {
		reason := result.SkipReason
		if reason == "" {
			reason = "Not a valid entity"
		}
		return &ConvertedEntity{
			Name:    "",
			Skip:    true,
			Notes:   []string{reason},
		}, nil
	}

	// Check if review is needed
	if result.RequiresReview && m.config.Interactive {
		if err := m.promptForEntityReview(entity, result); err != nil {
			return nil, err
		}
	}

	return &ConvertedEntity{
		Name:           result.Entity.Name,
		EntityCode:     result.Entity.Code,
		DTOCode:        result.DTO.Code,
		ResponseCode:   result.Response.Code,
		FlywayMigration: result.FlywayMigration,
		Notes:          result.Notes,
	}, nil
}

// writeEntityFiles writes the converted entity files to the output directory
func (m *Migrator) writeEntityFiles(entity *ConvertedEntity) error {
	// Skip if this is not a valid entity
	if entity.Skip || entity.Name == "" {
		return nil
	}

	packagePath := strings.ReplaceAll(m.projectInfo.GroupID, ".", string(filepath.Separator))
	baseDir := filepath.Join(m.config.OutputPath, config.ModuleModel, "src", "main", "java", packagePath, "model")

	// Write entity
	entityDir := filepath.Join(baseDir, "entities")
	if err := os.MkdirAll(entityDir, 0755); err != nil {
		return err
	}
	entityPath := filepath.Join(entityDir, entity.Name+".java")
	if err := os.WriteFile(entityPath, []byte(entity.EntityCode), 0644); err != nil {
		return err
	}

	// Write request DTO
	dtoDir := filepath.Join(baseDir, "dto")
	if err := os.MkdirAll(dtoDir, 0755); err != nil {
		return err
	}

	// Extract DTO name from code (look for class name)
	dtoName := entity.Name + "Request"
	if entity.DTOCode != "" {
		dtoPath := filepath.Join(dtoDir, dtoName+".java")
		if err := os.WriteFile(dtoPath, []byte(entity.DTOCode), 0644); err != nil {
			return err
		}
	}

	// Write response DTO
	if entity.ResponseCode != "" {
		responseName := entity.Name + "Response"
		responsePath := filepath.Join(dtoDir, responseName+".java")
		if err := os.WriteFile(responsePath, []byte(entity.ResponseCode), 0644); err != nil {
			return err
		}
	}

	return nil
}

// generateFlywayMigrations generates Flyway migration files from entity conversions
func (m *Migrator) generateFlywayMigrations(ctx context.Context) error {
	// Collect all Flyway migrations from converted entities
	var migrations []string

	// Get migrations from checkpoint data
	if m.checkpoint.Current() != nil {
		if data, ok := m.checkpoint.Current().Data["entity_migrations"]; ok {
			if migrationList, ok := data.([]string); ok {
				migrations = migrationList
			}
		}
	}

	if len(migrations) == 0 {
		return nil
	}

	// Write migration file
	migrationDir := filepath.Join(m.config.OutputPath, config.ModuleSQLDatastore,
		"src", "main", "resources", "db", "migration")
	if err := os.MkdirAll(migrationDir, 0755); err != nil {
		return err
	}

	// Create V1__baseline.sql
	migrationPath := filepath.Join(migrationDir, "V1__baseline.sql")
	content := "-- Baseline migration generated by Trabuco Migrate\n\n" + strings.Join(migrations, "\n\n")

	return os.WriteFile(migrationPath, []byte(content), 0644)
}

// parseEntityMigrationResult parses the JSON response from AI
func parseEntityMigrationResult(content string) (*EntityMigrationResult, error) {
	// Find JSON in the response (it might be wrapped in markdown code blocks)
	jsonContent := extractJSON(content)
	if jsonContent == "" {
		return nil, fmt.Errorf("no JSON found in response")
	}

	var result EntityMigrationResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		// Log first 500 chars of extracted content for debugging
		preview := jsonContent
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		return nil, fmt.Errorf("failed to parse JSON: %w\nExtracted content preview: %s", err, preview)
	}

	return &result, nil
}

// extractJSON extracts JSON from a response that might contain markdown
func extractJSON(content string) string {
	// First, try to strip any markdown code block wrapper
	trimmed := strings.TrimSpace(content)

	// Check for ```json wrapper
	if strings.HasPrefix(trimmed, "```json") {
		// Find the end of the opening line
		start := 7 // len("```json")
		if start < len(trimmed) && trimmed[start] == '\n' {
			start++
		}
		rest := trimmed[start:]
		// Find closing ``` in the rest of the content
		if end := strings.Index(rest, "```"); end != -1 {
			return strings.TrimSpace(rest[:end])
		}
		// No closing ```, try to extract JSON from what we have
		return extractJSONObject(rest)
	}

	// Check for plain ``` wrapper
	if strings.HasPrefix(trimmed, "```") {
		// Skip the ``` and any language identifier
		start := 3
		// Find end of first line
		if newlineIdx := strings.Index(trimmed[start:], "\n"); newlineIdx != -1 {
			start = start + newlineIdx + 1
		}
		rest := trimmed[start:]
		// Find closing ```
		if end := strings.Index(rest, "```"); end != -1 {
			return strings.TrimSpace(rest[:end])
		}
		// No closing ```, try to extract JSON from what we have
		return extractJSONObject(rest)
	}

	// Try to find raw JSON
	return extractJSONObject(content)
}

// extractJSONObject finds and extracts a complete JSON object from content
func extractJSONObject(content string) string {
	if idx := strings.Index(content, "{"); idx != -1 {
		// Find matching closing brace, accounting for strings
		depth := 0
		inString := false
		escaped := false
		for i := idx; i < len(content); i++ {
			c := content[i]
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' && inString {
				escaped = true
				continue
			}
			if c == '"' {
				inString = !inString
				continue
			}
			if inString {
				continue
			}
			switch c {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return content[idx : i+1]
				}
			}
		}
	}

	return content
}

// promptForEntityReview prompts the user to review an entity conversion
func (m *Migrator) promptForEntityReview(original *JavaClass, result *EntityMigrationResult) error {
	fmt.Println()
	m.yellow.Printf("  ⚠ Entity %s requires review:\n", original.Name)
	fmt.Printf("    Reason: %s\n", result.ReviewReason)

	if len(result.Notes) > 0 {
		fmt.Println("    Notes:")
		for _, note := range result.Notes {
			fmt.Printf("      - %s\n", note)
		}
	}

	fmt.Print("    [A]ccept / [S]kip / [V]iew code: ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "v", "view":
		fmt.Println("\n--- Generated Entity ---")
		fmt.Println(result.Entity.Code)
		fmt.Println("--- End ---")
		return m.promptForEntityReview(original, result) // Ask again
	case "s", "skip":
		return fmt.Errorf("skipped by user")
	default:
		return nil // Accept
	}
}

// truncate truncates a string to the given length
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ConvertedEntity adds ResponseCode field
func init() {
	// This is a workaround - the ConvertedEntity struct in migrator.go
	// needs to have ResponseCode field. Since we can't modify structs
	// across files easily, we'll work with what we have.
}
