package ai

import (
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AnthropicProvider implements Provider using the Anthropic Claude API
type AnthropicProvider struct {
	client       anthropic.Client
	config       *ProviderConfig
	model        Model
	modelsClient *ModelsClient
}

// NewAnthropicProvider creates a new Anthropic provider
func NewAnthropicProvider(config *ProviderConfig) (*AnthropicProvider, error) {
	if config == nil {
		config = DefaultConfig("")
	}

	// Check for API key in config or environment
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}

	// Create client options
	opts := []option.RequestOption{
		option.WithAPIKey(apiKey),
	}

	if config.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(config.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	// Determine model using the unified GetModelByName function
	model := ModelClaudeSonnet // Default
	if config.Model != "" {
		if m, ok := GetModelByName(config.Model); ok {
			model = m
		}
	}

	return &AnthropicProvider{
		client:       client,
		config:       config,
		model:        model,
		modelsClient: NewModelsClient(apiKey),
	}, nil
}

// Name returns the provider name
func (p *AnthropicProvider) Name() string {
	return "anthropic"
}

// Analyze sends a prompt with code context and returns the AI response
func (p *AnthropicProvider) Analyze(ctx context.Context, request *AnalysisRequest) (*AnalysisResponse, error) {
	// Build the user message content
	userContent := request.UserPrompt
	if request.Code != "" {
		userContent += "\n\n```java\n" + request.Code + "\n```"
	}
	if request.Context != "" {
		userContent += "\n\nAdditional Context:\n" + request.Context
	}

	// Build messages
	messages := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userContent)),
	}

	// Determine model
	modelID := p.model.ID
	if request.Model != "" {
		modelID = request.Model
	}

	// Build request params
	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(modelID),
		MaxTokens: int64(request.MaxTokens),
		Messages:  messages,
	}

	// Add system prompt if provided
	if request.SystemPrompt != "" {
		params.System = []anthropic.TextBlockParam{
			{Text: request.SystemPrompt},
		}
	}

	// Set temperature if specified
	if request.Temperature > 0 {
		params.Temperature = anthropic.Float(request.Temperature)
	}

	// Make the API call
	message, err := p.client.Messages.New(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderError, err)
	}

	// Extract text content
	var content string
	for _, block := range message.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	return &AnalysisResponse{
		Content:      content,
		InputTokens:  int(message.Usage.InputTokens),
		OutputTokens: int(message.Usage.OutputTokens),
		Model:        string(message.Model),
		StopReason:   string(message.StopReason),
	}, nil
}

// Stream sends a prompt and streams the response
func (p *AnthropicProvider) Stream(ctx context.Context, request *AnalysisRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)

		// Build the user message content
		userContent := request.UserPrompt
		if request.Code != "" {
			userContent += "\n\n```java\n" + request.Code + "\n```"
		}
		if request.Context != "" {
			userContent += "\n\nAdditional Context:\n" + request.Context
		}

		// Build messages
		messages := []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userContent)),
		}

		// Determine model
		modelID := p.model.ID
		if request.Model != "" {
			modelID = request.Model
		}

		// Build request params
		params := anthropic.MessageNewParams{
			Model:     anthropic.Model(modelID),
			MaxTokens: int64(request.MaxTokens),
			Messages:  messages,
		}

		// Add system prompt if provided
		if request.SystemPrompt != "" {
			params.System = []anthropic.TextBlockParam{
				{Text: request.SystemPrompt},
			}
		}

		// Create streaming request
		stream := p.client.Messages.NewStreaming(ctx, params)

		var totalInputTokens, totalOutputTokens int

		for stream.Next() {
			event := stream.Current()

			switch eventVariant := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch deltaVariant := eventVariant.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					ch <- StreamChunk{Text: deltaVariant.Text}
				}
			case anthropic.MessageDeltaEvent:
				totalOutputTokens = int(eventVariant.Usage.OutputTokens)
			case anthropic.MessageStartEvent:
				totalInputTokens = int(eventVariant.Message.Usage.InputTokens)
			}
		}

		if err := stream.Err(); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("%w: %v", ErrProviderError, err)}
			return
		}

		// Send final chunk with usage
		ch <- StreamChunk{
			Done: true,
			Usage: &TokenUsage{
				InputTokens:  totalInputTokens,
				OutputTokens: totalOutputTokens,
			},
		}
	}()

	return ch, nil
}

// ValidateAPIKey checks if the API key is valid by making a minimal request
func (p *AnthropicProvider) ValidateAPIKey(ctx context.Context) error {
	// Make a minimal request to validate the key
	_, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(p.model.ID),
		MaxTokens: 10,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock("Hi")),
		},
	})

	if err != nil {
		// Check for authentication errors
		return fmt.Errorf("%w: %v", ErrInvalidAPIKey, err)
	}

	return nil
}

// EstimateTokens estimates the number of tokens for the given content
// Using a rough approximation: ~4 characters per token for English text,
// ~3 characters per token for code
func (p *AnthropicProvider) EstimateTokens(content string) int {
	// Rough estimation: code is denser than natural language
	// Average ~3.5 characters per token
	return len(content) / 3
}

// EstimateCost estimates the cost in USD for the given token counts
func (p *AnthropicProvider) EstimateCost(inputTokens, outputTokens int) float64 {
	inputCost := float64(inputTokens) * p.model.InputCostPer1M / 1_000_000
	outputCost := float64(outputTokens) * p.model.OutputCostPer1M / 1_000_000
	return inputCost + outputCost
}

// GetModel returns the current model configuration
func (p *AnthropicProvider) GetModel() Model {
	return p.model
}

// ListAvailableModels returns all models available via the Anthropic API
func (p *AnthropicProvider) ListAvailableModels(ctx context.Context) ([]APIModel, error) {
	return p.modelsClient.ListModels(ctx)
}

// ValidateModel checks if the specified model exists
func (p *AnthropicProvider) ValidateModel(ctx context.Context, modelID string) (bool, error) {
	return p.modelsClient.ValidateModel(ctx, modelID)
}

// ResolveModel resolves a model alias to its actual ID and returns whether it's an alias
func (p *AnthropicProvider) ResolveModel(ctx context.Context, modelID string) (resolvedID string, isAlias bool, err error) {
	return p.modelsClient.ResolveModelAlias(ctx, modelID)
}

// IsUsingAlias returns true if the provider is configured to use a model alias
func (p *AnthropicProvider) IsUsingAlias() bool {
	return p.model.IsAlias
}
