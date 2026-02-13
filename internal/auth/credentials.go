package auth

import (
	"encoding/json"
	"errors"
	"time"
)

// Provider represents a supported LLM provider
type Provider string

const (
	ProviderAnthropic  Provider = "anthropic"
	ProviderOpenRouter Provider = "openrouter"
	ProviderOpenAI     Provider = "openai"
	ProviderOllama     Provider = "ollama"
)

// ProviderInfo contains metadata about an LLM provider
type ProviderInfo struct {
	Name           string   // Human-readable name
	EnvVar         string   // Environment variable name
	KeyPrefix      string   // API key prefix for validation (empty if no standard prefix)
	BaseURL        string   // Default API base URL
	DocumentURL    string   // URL to get API keys
	Models         []string // Available models
	InputCostPer1M float64  // Cost per 1M input tokens (for primary model)
	OutputCostPer1M float64 // Cost per 1M output tokens (for primary model)
	RequiresKey    bool     // Whether this provider requires an API key
}

// SupportedProviders contains info for all supported providers
var SupportedProviders = map[Provider]ProviderInfo{
	ProviderAnthropic: {
		Name:           "Anthropic (Claude)",
		EnvVar:         "ANTHROPIC_API_KEY",
		KeyPrefix:      "sk-ant-",
		BaseURL:        "https://api.anthropic.com",
		DocumentURL:    "https://console.anthropic.com/settings/keys",
		Models:         []string{"claude-sonnet-4-5", "claude-haiku-4-5", "claude-opus-4-6"},
		InputCostPer1M: 3.00,
		OutputCostPer1M: 15.00,
		RequiresKey:    true,
	},
	ProviderOpenRouter: {
		Name:           "OpenRouter",
		EnvVar:         "OPENROUTER_API_KEY",
		KeyPrefix:      "sk-or-",
		BaseURL:        "https://openrouter.ai/api",
		DocumentURL:    "https://openrouter.ai/keys",
		Models:         []string{"anthropic/claude-sonnet-4-5", "anthropic/claude-haiku-4-5", "openai/gpt-4o"},
		InputCostPer1M: 3.00,
		OutputCostPer1M: 15.00,
		RequiresKey:    true,
	},
	ProviderOpenAI: {
		Name:           "OpenAI",
		EnvVar:         "OPENAI_API_KEY",
		KeyPrefix:      "sk-",
		BaseURL:        "https://api.openai.com",
		DocumentURL:    "https://platform.openai.com/api-keys",
		Models:         []string{"gpt-4o", "gpt-4o-mini"},
		InputCostPer1M: 2.50,
		OutputCostPer1M: 10.00,
		RequiresKey:    true,
	},
	ProviderOllama: {
		Name:           "Ollama (Local)",
		EnvVar:         "",
		KeyPrefix:      "",
		BaseURL:        "http://localhost:11434",
		DocumentURL:    "https://ollama.ai/download",
		Models:         []string{"llama3.3", "codellama", "mistral"},
		InputCostPer1M: 0.00,
		OutputCostPer1M: 0.00,
		RequiresKey:    false,
	},
}

// Credential represents stored credentials for a provider
type Credential struct {
	Provider    Provider  `json:"provider"`
	APIKey      string    `json:"api_key,omitempty"`
	BaseURL     string    `json:"base_url,omitempty"`
	Model       string    `json:"model,omitempty"`
	ValidatedAt time.Time `json:"validated_at,omitempty"`
	IsDefault   bool      `json:"is_default,omitempty"`
}

// CredentialStore represents a collection of credentials
type CredentialStore struct {
	Credentials    map[Provider]*Credential `json:"credentials"`
	DefaultProvider Provider                 `json:"default_provider,omitempty"`
	UpdatedAt      time.Time                `json:"updated_at"`
}

// NewCredentialStore creates an empty credential store
func NewCredentialStore() *CredentialStore {
	return &CredentialStore{
		Credentials: make(map[Provider]*Credential),
		UpdatedAt:   time.Now(),
	}
}

// SetCredential stores a credential for a provider
func (s *CredentialStore) SetCredential(cred *Credential) {
	s.Credentials[cred.Provider] = cred
	s.UpdatedAt = time.Now()

	// Set as default if no default exists
	if s.DefaultProvider == "" {
		s.DefaultProvider = cred.Provider
	}
}

// GetCredential retrieves a credential for a provider
func (s *CredentialStore) GetCredential(provider Provider) (*Credential, bool) {
	cred, ok := s.Credentials[provider]
	return cred, ok
}

// GetDefaultCredential returns the default provider's credential
func (s *CredentialStore) GetDefaultCredential() (*Credential, bool) {
	if s.DefaultProvider == "" {
		return nil, false
	}
	return s.GetCredential(s.DefaultProvider)
}

// RemoveCredential removes a credential
func (s *CredentialStore) RemoveCredential(provider Provider) {
	delete(s.Credentials, provider)
	s.UpdatedAt = time.Now()

	// Clear default if we removed it
	if s.DefaultProvider == provider {
		s.DefaultProvider = ""
		// Set a new default if any credentials remain
		for p := range s.Credentials {
			s.DefaultProvider = p
			break
		}
	}
}

// SetDefault sets the default provider
func (s *CredentialStore) SetDefault(provider Provider) error {
	if _, ok := s.Credentials[provider]; !ok {
		return errors.New("provider not configured")
	}
	s.DefaultProvider = provider
	s.UpdatedAt = time.Now()
	return nil
}

// ListProviders returns all configured providers
func (s *CredentialStore) ListProviders() []Provider {
	providers := make([]Provider, 0, len(s.Credentials))
	for p := range s.Credentials {
		providers = append(providers, p)
	}
	return providers
}

// MarshalJSON implements custom JSON marshaling
func (s *CredentialStore) MarshalJSON() ([]byte, error) {
	type alias CredentialStore
	return json.Marshal((*alias)(s))
}

// UnmarshalJSON implements custom JSON unmarshaling
func (s *CredentialStore) UnmarshalJSON(data []byte) error {
	type alias CredentialStore
	aux := (*alias)(s)
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	if s.Credentials == nil {
		s.Credentials = make(map[Provider]*Credential)
	}
	return nil
}

// ValidateAPIKey checks if an API key looks valid for a provider
func ValidateAPIKey(provider Provider, apiKey string) error {
	info, ok := SupportedProviders[provider]
	if !ok {
		return errors.New("unsupported provider")
	}

	if !info.RequiresKey {
		return nil // No key required
	}

	if apiKey == "" {
		return errors.New("API key is required")
	}

	// Check prefix if provider has one
	if info.KeyPrefix != "" {
		if len(apiKey) < len(info.KeyPrefix) {
			return errors.New("API key appears invalid (too short)")
		}
		// Note: We don't enforce prefix strictly as some users may have legacy keys
	}

	return nil
}

// GetProviderFromKey attempts to detect the provider from an API key prefix
func GetProviderFromKey(apiKey string) (Provider, bool) {
	for provider, info := range SupportedProviders {
		if info.KeyPrefix != "" && len(apiKey) >= len(info.KeyPrefix) {
			if apiKey[:len(info.KeyPrefix)] == info.KeyPrefix {
				return provider, true
			}
		}
	}
	return "", false
}

// Common errors
var (
	ErrNoCredentials     = errors.New("no credentials configured")
	ErrProviderNotFound  = errors.New("provider not found")
	ErrInvalidCredential = errors.New("invalid credential")
	ErrStorageError      = errors.New("credential storage error")
)
