package ai

import (
	"context"
	"errors"
)

// Provider represents an AI provider that can analyze and transform code
type Provider interface {
	// Name returns the provider name (e.g., "anthropic", "openrouter")
	Name() string

	// Analyze sends a prompt with code context and returns the AI response
	Analyze(ctx context.Context, request *AnalysisRequest) (*AnalysisResponse, error)

	// Stream sends a prompt and streams the response
	Stream(ctx context.Context, request *AnalysisRequest) (<-chan StreamChunk, error)

	// ValidateAPIKey checks if the API key is valid
	ValidateAPIKey(ctx context.Context) error

	// EstimateTokens estimates the number of tokens for the given content
	EstimateTokens(content string) int

	// EstimateCost estimates the cost in USD for the given token counts
	EstimateCost(inputTokens, outputTokens int) float64
}

// Model represents an AI model configuration
type Model struct {
	ID              string  // Model identifier (e.g., "claude-sonnet-4-5" or "claude-sonnet-4-5-20250929")
	Alias           string  // Model alias (e.g., "claude-sonnet-4-5") - empty if ID is already an alias
	Name            string  // Human-readable name
	InputCostPer1M  float64 // Cost per 1M input tokens in USD
	OutputCostPer1M float64 // Cost per 1M output tokens in USD
	ContextWindow   int     // Maximum context window size
	MaxOutputTokens int     // Maximum output tokens
	IsAlias         bool    // True if ID is an alias that auto-updates to latest version
}

// Common models using aliases (auto-update to latest versions)
// For pinned versions, use the full model ID like "claude-sonnet-4-5-20250929"
var (
	// ModelClaudeSonnet uses the alias for automatic updates
	ModelClaudeSonnet = Model{
		ID:              "claude-sonnet-4-5",
		Alias:           "claude-sonnet-4-5",
		Name:            "Claude Sonnet 4.5",
		InputCostPer1M:  3.00,
		OutputCostPer1M: 15.00,
		ContextWindow:   200000,
		MaxOutputTokens: 64000,
		IsAlias:         true,
	}

	// ModelClaudeHaiku uses the alias for automatic updates
	ModelClaudeHaiku = Model{
		ID:              "claude-haiku-4-5",
		Alias:           "claude-haiku-4-5",
		Name:            "Claude Haiku 4.5",
		InputCostPer1M:  1.00,
		OutputCostPer1M: 5.00,
		ContextWindow:   200000,
		MaxOutputTokens: 64000,
		IsAlias:         true,
	}

	// ModelClaudeOpus uses the alias for automatic updates
	ModelClaudeOpus = Model{
		ID:              "claude-opus-4-6",
		Alias:           "claude-opus-4-6",
		Name:            "Claude Opus 4.6",
		InputCostPer1M:  5.00,
		OutputCostPer1M: 25.00,
		ContextWindow:   200000,
		MaxOutputTokens: 128000,
		IsAlias:         true,
	}

	// modelsByName maps CLI-friendly names to models
	modelsByName = map[string]Model{
		// Aliases (recommended - auto-update)
		"sonnet":           ModelClaudeSonnet,
		"claude-sonnet":    ModelClaudeSonnet,
		"claude-sonnet-4":  ModelClaudeSonnet,
		"claude-sonnet-4-5": ModelClaudeSonnet,
		"haiku":            ModelClaudeHaiku,
		"claude-haiku":     ModelClaudeHaiku,
		"claude-haiku-4":   ModelClaudeHaiku,
		"claude-haiku-4-5": ModelClaudeHaiku,
		"opus":             ModelClaudeOpus,
		"claude-opus":      ModelClaudeOpus,
		"claude-opus-4":    ModelClaudeOpus,
		"claude-opus-4-6":  ModelClaudeOpus,
	}
)

// GetModelByName returns a model by its CLI-friendly name or model ID.
// It supports both aliases (e.g., "claude-sonnet-4-5") and pinned versions
// (e.g., "claude-sonnet-4-5-20250929").
func GetModelByName(name string) (Model, bool) {
	// First check our known models map
	if model, ok := modelsByName[name]; ok {
		return model, true
	}

	// If not found, treat it as a custom/pinned model ID
	// This allows users to specify exact versions like "claude-sonnet-4-5-20250929"
	if name != "" {
		return Model{
			ID:      name,
			Name:    name,
			IsAlias: false, // Pinned version
		}, true
	}

	return Model{}, false
}

// GetModelTier returns the model tier (sonnet, haiku, opus) from a model ID or alias
func GetModelTier(modelID string) string {
	switch {
	case contains(modelID, "sonnet"):
		return "sonnet"
	case contains(modelID, "haiku"):
		return "haiku"
	case contains(modelID, "opus"):
		return "opus"
	default:
		return "unknown"
	}
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsLower(s, substr))
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldSubstring(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

func equalFoldSubstring(a, b string) bool {
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

// AnalysisRequest represents a request to analyze code
type AnalysisRequest struct {
	// SystemPrompt sets the AI's behavior context
	SystemPrompt string

	// UserPrompt is the main prompt with instructions
	UserPrompt string

	// Code is the source code to analyze
	Code string

	// Context provides additional context (related files, etc.)
	Context string

	// MaxTokens limits the response length
	MaxTokens int

	// Temperature controls randomness (0.0 = deterministic, 1.0 = creative)
	Temperature float64

	// Model specifies which model to use (optional, uses default if empty)
	Model string
}

// AnalysisResponse represents the AI's response
type AnalysisResponse struct {
	// Content is the AI's response text
	Content string

	// InputTokens is the number of tokens in the request
	InputTokens int

	// OutputTokens is the number of tokens in the response
	OutputTokens int

	// Model is the model that was used
	Model string

	// StopReason indicates why the response ended
	StopReason string
}

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	// Text is the chunk of text
	Text string

	// Done indicates if this is the final chunk
	Done bool

	// Error contains any error that occurred
	Error error

	// Usage contains token usage (only in final chunk)
	Usage *TokenUsage
}

// TokenUsage tracks token consumption
type TokenUsage struct {
	InputTokens  int
	OutputTokens int
}

// Common errors
var (
	ErrInvalidAPIKey    = errors.New("invalid API key")
	ErrRateLimited      = errors.New("rate limited, please try again later")
	ErrContextTooLong   = errors.New("context exceeds maximum token limit")
	ErrProviderError    = errors.New("AI provider error")
	ErrNoAPIKey         = errors.New("no API key provided")
	ErrModelNotFound    = errors.New("model not found")
)

// ProviderConfig holds configuration for creating a provider
type ProviderConfig struct {
	// APIKey is the authentication key
	APIKey string

	// Model specifies the default model to use
	Model string

	// BaseURL overrides the default API endpoint (for proxies)
	BaseURL string

	// Timeout in seconds for API calls
	Timeout int

	// MaxRetries for failed requests
	MaxRetries int
}

// DefaultConfig returns a default provider configuration
func DefaultConfig(apiKey string) *ProviderConfig {
	return &ProviderConfig{
		APIKey:     apiKey,
		Model:      ModelClaudeSonnet.ID,
		Timeout:    120,
		MaxRetries: 3,
	}
}

// Validate checks if the analysis request is valid
func (r *AnalysisRequest) Validate() error {
	if r.UserPrompt == "" {
		return errors.New("user prompt is required")
	}
	return nil
}

// EstimateTokens provides a rough estimate of tokens for given content
// Uses ~4 characters per token as a simple heuristic
func EstimateTokens(content string) int {
	if len(content) == 0 {
		return 0
	}
	tokens := len(content) / 4
	if tokens == 0 {
		tokens = 1
	}
	return tokens
}
