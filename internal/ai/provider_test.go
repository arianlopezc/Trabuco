package ai

import (
	"testing"
)

func TestModelPricing(t *testing.T) {
	tests := []struct {
		name         string
		model        Model
		inputTokens  int
		outputTokens int
		wantMinCost  float64
		wantMaxCost  float64
	}{
		{
			name:         "Claude Sonnet 4.5 pricing",
			model:        ModelClaudeSonnet,
			inputTokens:  1000,
			outputTokens: 1000,
			wantMinCost:  0.01,
			wantMaxCost:  0.02,
		},
		{
			name:         "Claude Haiku 4.5 pricing (cheaper)",
			model:        ModelClaudeHaiku,
			inputTokens:  1000,
			outputTokens: 1000,
			wantMinCost:  0.001,
			wantMaxCost:  0.01,
		},
		{
			name:         "Claude Opus 4.6 pricing",
			model:        ModelClaudeOpus,
			inputTokens:  1000,
			outputTokens: 1000,
			wantMinCost:  0.01,
			wantMaxCost:  0.05,
		},
		{
			name:         "Zero tokens",
			model:        ModelClaudeSonnet,
			inputTokens:  0,
			outputTokens: 0,
			wantMinCost:  0,
			wantMaxCost:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cost := (float64(tt.inputTokens) * tt.model.InputCostPer1M / 1_000_000) +
				(float64(tt.outputTokens) * tt.model.OutputCostPer1M / 1_000_000)

			if cost < tt.wantMinCost || cost > tt.wantMaxCost {
				t.Errorf("cost = %v, want between %v and %v", cost, tt.wantMinCost, tt.wantMaxCost)
			}
		})
	}
}

func TestModelProperties(t *testing.T) {
	tests := []struct {
		name              string
		model             Model
		wantID            string
		wantMaxTokens     int
		wantContextWindow int
		wantIsAlias       bool
	}{
		{
			name:              "Claude Sonnet 4.5",
			model:             ModelClaudeSonnet,
			wantID:            "claude-sonnet-4-5",
			wantMaxTokens:     64000,
			wantContextWindow: 200000,
			wantIsAlias:       true,
		},
		{
			name:              "Claude Haiku 4.5",
			model:             ModelClaudeHaiku,
			wantID:            "claude-haiku-4-5",
			wantMaxTokens:     64000,
			wantContextWindow: 200000,
			wantIsAlias:       true,
		},
		{
			name:              "Claude Opus 4.6",
			model:             ModelClaudeOpus,
			wantID:            "claude-opus-4-6",
			wantMaxTokens:     128000,
			wantContextWindow: 200000,
			wantIsAlias:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.model.ID != tt.wantID {
				t.Errorf("ID = %v, want %v", tt.model.ID, tt.wantID)
			}
			if tt.model.MaxOutputTokens != tt.wantMaxTokens {
				t.Errorf("MaxOutputTokens = %v, want %v", tt.model.MaxOutputTokens, tt.wantMaxTokens)
			}
			if tt.model.ContextWindow != tt.wantContextWindow {
				t.Errorf("ContextWindow = %v, want %v", tt.model.ContextWindow, tt.wantContextWindow)
			}
			if tt.model.IsAlias != tt.wantIsAlias {
				t.Errorf("IsAlias = %v, want %v", tt.model.IsAlias, tt.wantIsAlias)
			}
		})
	}
}

func TestGetModelByName(t *testing.T) {
	tests := []struct {
		name        string
		modelName   string
		wantModelID string
		wantFound   bool
		wantIsAlias bool
	}{
		// Alias shortcuts
		{
			name:        "sonnet shortcut",
			modelName:   "sonnet",
			wantModelID: "claude-sonnet-4-5",
			wantFound:   true,
			wantIsAlias: true,
		},
		{
			name:        "haiku shortcut",
			modelName:   "haiku",
			wantModelID: "claude-haiku-4-5",
			wantFound:   true,
			wantIsAlias: true,
		},
		{
			name:        "opus shortcut",
			modelName:   "opus",
			wantModelID: "claude-opus-4-6",
			wantFound:   true,
			wantIsAlias: true,
		},
		// Full alias names
		{
			name:        "claude-sonnet-4-5",
			modelName:   "claude-sonnet-4-5",
			wantModelID: "claude-sonnet-4-5",
			wantFound:   true,
			wantIsAlias: true,
		},
		{
			name:        "claude-haiku-4-5",
			modelName:   "claude-haiku-4-5",
			wantModelID: "claude-haiku-4-5",
			wantFound:   true,
			wantIsAlias: true,
		},
		{
			name:        "claude-opus-4-6",
			modelName:   "claude-opus-4-6",
			wantModelID: "claude-opus-4-6",
			wantFound:   true,
			wantIsAlias: true,
		},
		// Pinned version (custom model)
		{
			name:        "pinned version",
			modelName:   "claude-sonnet-4-5-20250929",
			wantModelID: "claude-sonnet-4-5-20250929",
			wantFound:   true,
			wantIsAlias: false,
		},
		{
			name:      "empty string",
			modelName: "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model, found := GetModelByName(tt.modelName)
			if found != tt.wantFound {
				t.Errorf("found = %v, want %v", found, tt.wantFound)
			}
			if found {
				if model.ID != tt.wantModelID {
					t.Errorf("model.ID = %v, want %v", model.ID, tt.wantModelID)
				}
				if model.IsAlias != tt.wantIsAlias {
					t.Errorf("model.IsAlias = %v, want %v", model.IsAlias, tt.wantIsAlias)
				}
			}
		})
	}
}

func TestAnalysisRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		request *AnalysisRequest
		wantErr bool
	}{
		{
			name: "valid request",
			request: &AnalysisRequest{
				SystemPrompt: "You are a helpful assistant.",
				UserPrompt:   "Hello",
				MaxTokens:    1000,
				Temperature:  0.7,
			},
			wantErr: false,
		},
		{
			name: "empty user prompt",
			request: &AnalysisRequest{
				SystemPrompt: "You are a helpful assistant.",
				UserPrompt:   "",
				MaxTokens:    1000,
			},
			wantErr: true,
		},
		{
			name: "zero max tokens is valid (uses default)",
			request: &AnalysisRequest{
				SystemPrompt: "You are a helpful assistant.",
				UserPrompt:   "Hello",
				MaxTokens:    0,
			},
			wantErr: false,
		},
		{
			name: "no system prompt is valid",
			request: &AnalysisRequest{
				UserPrompt: "Hello",
				MaxTokens:  1000,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantMin int
		wantMax int
	}{
		{
			name:    "empty string",
			content: "",
			wantMin: 0,
			wantMax: 0,
		},
		{
			name:    "short text",
			content: "Hello world",
			wantMin: 1,
			wantMax: 5,
		},
		{
			name:    "longer text",
			content: "This is a longer piece of text that should result in more tokens being estimated.",
			wantMin: 10,
			wantMax: 30,
		},
		{
			name:    "code snippet",
			content: "public class HelloWorld { public static void main(String[] args) { System.out.println(\"Hello\"); } }",
			wantMin: 15,
			wantMax: 40,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			estimated := EstimateTokens(tt.content)

			if estimated < tt.wantMin || estimated > tt.wantMax {
				t.Errorf("EstimateTokens() = %v, want between %v and %v", estimated, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test-api-key")

	if config.APIKey != "test-api-key" {
		t.Errorf("APIKey = %v, want test-api-key", config.APIKey)
	}

	if config.Model != ModelClaudeSonnet.ID {
		t.Errorf("Model = %v, want %v", config.Model, ModelClaudeSonnet.ID)
	}

	if config.Timeout != 120 {
		t.Errorf("Timeout = %v, want 120", config.Timeout)
	}

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %v, want 3", config.MaxRetries)
	}
}
