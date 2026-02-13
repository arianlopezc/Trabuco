package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	anthropicModelsURL = "https://api.anthropic.com/v1/models"
	anthropicVersion   = "2023-06-01"
)

// APIModel represents a model returned by the Anthropic Models API
type APIModel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	CreatedAt   string `json:"created_at"`
	Type        string `json:"type"`
}

// ModelsResponse represents the response from the Models API
type ModelsResponse struct {
	Data    []APIModel `json:"data"`
	HasMore bool       `json:"has_more"`
	FirstID string     `json:"first_id"`
	LastID  string     `json:"last_id"`
}

// ModelsClient provides access to the Anthropic Models API
type ModelsClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewModelsClient creates a new Models API client
func NewModelsClient(apiKey string) *ModelsClient {
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	return &ModelsClient{
		apiKey:  apiKey,
		baseURL: anthropicModelsURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListModels fetches all available models from the Anthropic API
// Models are returned sorted by release date (newest first)
func (c *ModelsClient) ListModels(ctx context.Context) ([]APIModel, error) {
	if c.apiKey == "" {
		return nil, ErrNoAPIKey
	}

	var allModels []APIModel
	afterID := ""

	for {
		url := c.baseURL
		if afterID != "" {
			url = fmt.Sprintf("%s?after_id=%s", c.baseURL, afterID)
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("anthropic-version", anthropicVersion)

		resp, err := c.client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrProviderError, err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == 401 {
			return nil, ErrInvalidAPIKey
		}
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("%w: %s", ErrProviderError, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		var modelsResp ModelsResponse
		if err := json.Unmarshal(body, &modelsResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		allModels = append(allModels, modelsResp.Data...)

		if !modelsResp.HasMore {
			break
		}
		afterID = modelsResp.LastID
	}

	return allModels, nil
}

// ValidateModel checks if a model ID exists in the available models
func (c *ModelsClient) ValidateModel(ctx context.Context, modelID string) (bool, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return false, err
	}

	for _, m := range models {
		if m.ID == modelID {
			return true, nil
		}
	}

	return false, nil
}

// GetLatestModel returns the latest model for a given tier (sonnet, haiku, opus)
// Models are returned sorted by release date, so the first match is the latest
func (c *ModelsClient) GetLatestModel(ctx context.Context, tier string) (*APIModel, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	tier = strings.ToLower(tier)
	for _, m := range models {
		if strings.Contains(strings.ToLower(m.ID), tier) {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("%w: no %s model found", ErrModelNotFound, tier)
}

// GetModelInfo returns detailed information about a specific model
func (c *ModelsClient) GetModelInfo(ctx context.Context, modelID string) (*APIModel, error) {
	models, err := c.ListModels(ctx)
	if err != nil {
		return nil, err
	}

	for _, m := range models {
		if m.ID == modelID {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrModelNotFound, modelID)
}

// ResolveModelAlias resolves a model alias to the actual model ID
// If the modelID is already a full version (contains date), it's returned as-is
// If it's an alias, the API is queried to find the matching model
func (c *ModelsClient) ResolveModelAlias(ctx context.Context, modelID string) (string, bool, error) {
	// Check if it's already a pinned version (contains a date pattern like -20250929)
	if isPinnedVersion(modelID) {
		return modelID, false, nil // Not an alias, already pinned
	}

	// It's an alias - find the latest matching model
	models, err := c.ListModels(ctx)
	if err != nil {
		// If we can't reach the API, just use the alias directly
		// The API will resolve it server-side
		return modelID, true, nil
	}

	// Find the first (latest) model that matches the alias pattern
	for _, m := range models {
		if matchesAlias(m.ID, modelID) {
			return m.ID, true, nil
		}
	}

	// Alias not found in API, but it might still work server-side
	return modelID, true, nil
}

// isPinnedVersion checks if a model ID is a pinned version (contains date)
func isPinnedVersion(modelID string) bool {
	// Pinned versions have a date suffix like -20250929
	parts := strings.Split(modelID, "-")
	if len(parts) < 2 {
		return false
	}
	lastPart := parts[len(parts)-1]
	// Check if the last part is a date (8 digits)
	if len(lastPart) == 8 {
		for _, c := range lastPart {
			if c < '0' || c > '9' {
				return false
			}
		}
		return true
	}
	return false
}

// matchesAlias checks if a full model ID matches an alias
// e.g., "claude-sonnet-4-5-20250929" matches alias "claude-sonnet-4-5"
func matchesAlias(fullID, alias string) bool {
	return strings.HasPrefix(fullID, alias+"-") || fullID == alias
}

// FormatModelList returns a formatted string of available models
func FormatModelList(models []APIModel) string {
	var sb strings.Builder
	sb.WriteString("Available Models (newest first):\n")
	sb.WriteString("─────────────────────────────────────────\n")

	for _, m := range models {
		sb.WriteString(fmt.Sprintf("  • %s\n", m.ID))
		if m.DisplayName != "" && m.DisplayName != m.ID {
			sb.WriteString(fmt.Sprintf("    Display Name: %s\n", m.DisplayName))
		}
	}

	return sb.String()
}
