package ai

import (
	"fmt"
	"os"
)

// ProviderType represents the type of AI provider
type ProviderType string

const (
	ProviderTypeAnthropic  ProviderType = "anthropic"
	ProviderTypeOpenRouter ProviderType = "openrouter"
)

// NewProvider creates a new AI provider based on the specified type
func NewProvider(providerType ProviderType, config *ProviderConfig) (Provider, error) {
	switch providerType {
	case ProviderTypeAnthropic:
		return NewAnthropicProvider(config)
	case ProviderTypeOpenRouter:
		return NewOpenRouterProvider(config)
	default:
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
}

// AutoDetectProvider creates a provider by auto-detecting available API keys
// Priority: ANTHROPIC_API_KEY > OPENROUTER_API_KEY
func AutoDetectProvider(config *ProviderConfig) (Provider, error) {
	if config == nil {
		config = &ProviderConfig{}
	}

	// Check for explicit API key in config
	if config.APIKey != "" {
		// Try Anthropic first (if key looks like Anthropic key)
		if len(config.APIKey) > 7 && config.APIKey[:7] == "sk-ant-" {
			return NewAnthropicProvider(config)
		}
		// Try OpenRouter
		if len(config.APIKey) > 5 && config.APIKey[:5] == "sk-or" {
			return NewOpenRouterProvider(config)
		}
		// Default to Anthropic
		return NewAnthropicProvider(config)
	}

	// Check environment variables
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		return NewAnthropicProvider(config)
	}

	if os.Getenv("OPENROUTER_API_KEY") != "" {
		return NewOpenRouterProvider(config)
	}

	return nil, ErrNoAPIKey
}

// GetAvailableProviders returns a list of providers with available API keys
func GetAvailableProviders() []ProviderType {
	var available []ProviderType

	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		available = append(available, ProviderTypeAnthropic)
	}

	if os.Getenv("OPENROUTER_API_KEY") != "" {
		available = append(available, ProviderTypeOpenRouter)
	}

	return available
}

// GetProviderDescription returns a human-readable description of the provider
func GetProviderDescription(providerType ProviderType) string {
	switch providerType {
	case ProviderTypeAnthropic:
		return "Anthropic Claude API (direct)"
	case ProviderTypeOpenRouter:
		return "OpenRouter (multi-model gateway)"
	default:
		return string(providerType)
	}
}

// GetModelOptions returns available models for a provider
func GetModelOptions(providerType ProviderType) []Model {
	switch providerType {
	case ProviderTypeAnthropic:
		return []Model{
			ModelClaudeSonnet,
			ModelClaudeHaiku,
			ModelClaudeOpus,
		}
	case ProviderTypeOpenRouter:
		return []Model{
			OpenRouterModelClaudeSonnet,
			OpenRouterModelClaudeHaiku,
			OpenRouterModelGPT4,
		}
	default:
		return nil
	}
}
