package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	openRouterBaseURL = "https://openrouter.ai/api/v1"
)

// OpenRouter models mapping - using aliases for automatic updates
var (
	OpenRouterModelClaudeSonnet = Model{
		ID:              "anthropic/claude-sonnet-4-5",
		Alias:           "anthropic/claude-sonnet-4-5",
		Name:            "Claude Sonnet 4.5 (via OpenRouter)",
		InputCostPer1M:  3.00,
		OutputCostPer1M: 15.00,
		ContextWindow:   200000,
		MaxOutputTokens: 64000,
		IsAlias:         true,
	}

	OpenRouterModelClaudeHaiku = Model{
		ID:              "anthropic/claude-haiku-4-5",
		Alias:           "anthropic/claude-haiku-4-5",
		Name:            "Claude Haiku 4.5 (via OpenRouter)",
		InputCostPer1M:  1.00,
		OutputCostPer1M: 5.00,
		ContextWindow:   200000,
		MaxOutputTokens: 64000,
		IsAlias:         true,
	}

	OpenRouterModelClaudeOpus = Model{
		ID:              "anthropic/claude-opus-4-6",
		Alias:           "anthropic/claude-opus-4-6",
		Name:            "Claude Opus 4.6 (via OpenRouter)",
		InputCostPer1M:  5.00,
		OutputCostPer1M: 25.00,
		ContextWindow:   200000,
		MaxOutputTokens: 128000,
		IsAlias:         true,
	}

	OpenRouterModelGPT4 = Model{
		ID:              "openai/gpt-4-turbo",
		Name:            "GPT-4 Turbo (via OpenRouter)",
		InputCostPer1M:  10.00,
		OutputCostPer1M: 30.00,
		ContextWindow:   128000,
		MaxOutputTokens: 4096,
		IsAlias:         false,
	}
)

// OpenRouterProvider implements Provider using the OpenRouter API
type OpenRouterProvider struct {
	client  *http.Client
	config  *ProviderConfig
	model   Model
	baseURL string
}

// NewOpenRouterProvider creates a new OpenRouter provider
func NewOpenRouterProvider(config *ProviderConfig) (*OpenRouterProvider, error) {
	if config == nil {
		config = DefaultConfig("")
	}

	// Check for API key in config or environment
	apiKey := config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("OPENROUTER_API_KEY")
	}
	if apiKey == "" {
		return nil, ErrNoAPIKey
	}

	config.APIKey = apiKey

	// Create HTTP client with timeout
	timeout := time.Duration(config.Timeout) * time.Second
	if timeout == 0 {
		timeout = 120 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	// Determine model
	model := OpenRouterModelClaudeSonnet
	if config.Model != "" {
		switch config.Model {
		case "haiku", "claude-haiku", "claude-haiku-4", "claude-haiku-4-5":
			model = OpenRouterModelClaudeHaiku
		case "sonnet", "claude-sonnet", "claude-sonnet-4", "claude-sonnet-4-5":
			model = OpenRouterModelClaudeSonnet
		case "opus", "claude-opus", "claude-opus-4", "claude-opus-4-6":
			model = OpenRouterModelClaudeOpus
		case "gpt-4", "gpt4":
			model = OpenRouterModelGPT4
		default:
			// Use as custom model ID (supports pinned versions)
			model = Model{ID: config.Model, Name: config.Model, IsAlias: !isPinnedVersion(config.Model)}
		}
	}

	baseURL := openRouterBaseURL
	if config.BaseURL != "" {
		baseURL = config.BaseURL
	}

	return &OpenRouterProvider{
		client:  client,
		config:  config,
		model:   model,
		baseURL: baseURL,
	}, nil
}

// Name returns the provider name
func (p *OpenRouterProvider) Name() string {
	return "openrouter"
}

// OpenRouter API request/response structures
type openRouterRequest struct {
	Model       string                   `json:"model"`
	Messages    []openRouterMessage      `json:"messages"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	Temperature float64                  `json:"temperature,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
}

type openRouterMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	ID      string `json:"id"`
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

// Analyze sends a prompt with code context and returns the AI response
func (p *OpenRouterProvider) Analyze(ctx context.Context, request *AnalysisRequest) (*AnalysisResponse, error) {
	// Build messages
	var messages []openRouterMessage

	// Add system message if provided
	if request.SystemPrompt != "" {
		messages = append(messages, openRouterMessage{
			Role:    "system",
			Content: request.SystemPrompt,
		})
	}

	// Build user message content
	userContent := request.UserPrompt
	if request.Code != "" {
		userContent += "\n\n```java\n" + request.Code + "\n```"
	}
	if request.Context != "" {
		userContent += "\n\nAdditional Context:\n" + request.Context
	}

	messages = append(messages, openRouterMessage{
		Role:    "user",
		Content: userContent,
	})

	// Determine model
	modelID := p.model.ID
	if request.Model != "" {
		modelID = request.Model
	}

	// Build request
	reqBody := openRouterRequest{
		Model:       modelID,
		Messages:    messages,
		MaxTokens:   request.MaxTokens,
		Temperature: request.Temperature,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com/arianlopezc/Trabuco")
	httpReq.Header.Set("X-Title", "Trabuco CLI")

	// Make request
	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrProviderError, err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var orResp openRouterResponse
	if err := json.Unmarshal(body, &orResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if orResp.Error != nil {
		if orResp.Error.Type == "authentication_error" {
			return nil, ErrInvalidAPIKey
		}
		if orResp.Error.Type == "rate_limit_error" {
			return nil, ErrRateLimited
		}
		return nil, fmt.Errorf("%w: %s", ErrProviderError, orResp.Error.Message)
	}

	if len(orResp.Choices) == 0 {
		return nil, fmt.Errorf("%w: no response choices", ErrProviderError)
	}

	return &AnalysisResponse{
		Content:      orResp.Choices[0].Message.Content,
		InputTokens:  orResp.Usage.PromptTokens,
		OutputTokens: orResp.Usage.CompletionTokens,
		Model:        orResp.Model,
		StopReason:   orResp.Choices[0].FinishReason,
	}, nil
}

// Stream sends a prompt and streams the response
func (p *OpenRouterProvider) Stream(ctx context.Context, request *AnalysisRequest) (<-chan StreamChunk, error) {
	ch := make(chan StreamChunk, 100)

	go func() {
		defer close(ch)

		// Build messages
		var messages []openRouterMessage

		if request.SystemPrompt != "" {
			messages = append(messages, openRouterMessage{
				Role:    "system",
				Content: request.SystemPrompt,
			})
		}

		userContent := request.UserPrompt
		if request.Code != "" {
			userContent += "\n\n```java\n" + request.Code + "\n```"
		}
		if request.Context != "" {
			userContent += "\n\nAdditional Context:\n" + request.Context
		}

		messages = append(messages, openRouterMessage{
			Role:    "user",
			Content: userContent,
		})

		modelID := p.model.ID
		if request.Model != "" {
			modelID = request.Model
		}

		reqBody := openRouterRequest{
			Model:       modelID,
			Messages:    messages,
			MaxTokens:   request.MaxTokens,
			Temperature: request.Temperature,
			Stream:      true,
		}

		jsonBody, err := json.Marshal(reqBody)
		if err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("failed to marshal request: %w", err)}
			return
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
		if err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("failed to create request: %w", err)}
			return
		}

		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+p.config.APIKey)
		httpReq.Header.Set("HTTP-Referer", "https://github.com/arianlopezc/Trabuco")
		httpReq.Header.Set("X-Title", "Trabuco CLI")

		resp, err := p.client.Do(httpReq)
		if err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("%w: %v", ErrProviderError, err)}
			return
		}
		defer resp.Body.Close()

		// Read SSE stream
		decoder := json.NewDecoder(resp.Body)
		var totalInputTokens, totalOutputTokens int

		for {
			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
					FinishReason string `json:"finish_reason"`
				} `json:"choices"`
				Usage struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
				} `json:"usage"`
			}

			if err := decoder.Decode(&chunk); err != nil {
				if err == io.EOF {
					break
				}
				// Skip parsing errors for SSE format
				continue
			}

			if len(chunk.Choices) > 0 {
				if chunk.Choices[0].Delta.Content != "" {
					ch <- StreamChunk{Text: chunk.Choices[0].Delta.Content}
				}
				if chunk.Choices[0].FinishReason == "stop" {
					break
				}
			}

			if chunk.Usage.PromptTokens > 0 {
				totalInputTokens = chunk.Usage.PromptTokens
				totalOutputTokens = chunk.Usage.CompletionTokens
			}
		}

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

// ValidateAPIKey checks if the API key is valid
func (p *OpenRouterProvider) ValidateAPIKey(ctx context.Context) error {
	// Make a minimal request to validate the key
	_, err := p.Analyze(ctx, &AnalysisRequest{
		UserPrompt: "Hi",
		MaxTokens:  10,
	})

	if err != nil {
		return err
	}

	return nil
}

// EstimateTokens estimates the number of tokens for the given content
func (p *OpenRouterProvider) EstimateTokens(content string) int {
	return len(content) / 3
}

// EstimateCost estimates the cost in USD for the given token counts
func (p *OpenRouterProvider) EstimateCost(inputTokens, outputTokens int) float64 {
	inputCost := float64(inputTokens) * p.model.InputCostPer1M / 1_000_000
	outputCost := float64(outputTokens) * p.model.OutputCostPer1M / 1_000_000
	return inputCost + outputCost
}

// GetModel returns the current model configuration
func (p *OpenRouterProvider) GetModel() Model {
	return p.model
}
