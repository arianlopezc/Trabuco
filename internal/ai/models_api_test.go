package ai

import (
	"testing"
)

func TestIsPinnedVersion(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected bool
	}{
		{
			name:     "pinned version with date",
			modelID:  "claude-sonnet-4-5-20250929",
			expected: true,
		},
		{
			name:     "pinned version opus",
			modelID:  "claude-opus-4-6-20251015",
			expected: true,
		},
		{
			name:     "alias without date",
			modelID:  "claude-sonnet-4-5",
			expected: false,
		},
		{
			name:     "short alias",
			modelID:  "sonnet",
			expected: false,
		},
		{
			name:     "haiku alias",
			modelID:  "claude-haiku-4-5",
			expected: false,
		},
		{
			name:     "empty string",
			modelID:  "",
			expected: false,
		},
		{
			name:     "partial date (7 digits)",
			modelID:  "claude-sonnet-2025092",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPinnedVersion(tt.modelID)
			if result != tt.expected {
				t.Errorf("isPinnedVersion(%q) = %v, want %v", tt.modelID, result, tt.expected)
			}
		})
	}
}

func TestMatchesAlias(t *testing.T) {
	tests := []struct {
		name     string
		fullID   string
		alias    string
		expected bool
	}{
		{
			name:     "exact match",
			fullID:   "claude-sonnet-4-5",
			alias:    "claude-sonnet-4-5",
			expected: true,
		},
		{
			name:     "pinned matches alias",
			fullID:   "claude-sonnet-4-5-20250929",
			alias:    "claude-sonnet-4-5",
			expected: true,
		},
		{
			name:     "different models",
			fullID:   "claude-opus-4-6-20251015",
			alias:    "claude-sonnet-4-5",
			expected: false,
		},
		{
			name:     "partial alias matches prefix",
			fullID:   "claude-sonnet-4-5-20250929",
			alias:    "claude-sonnet-4",
			expected: true, // HasPrefix match: claude-sonnet-4-5-20250929 starts with claude-sonnet-4-
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesAlias(tt.fullID, tt.alias)
			if result != tt.expected {
				t.Errorf("matchesAlias(%q, %q) = %v, want %v", tt.fullID, tt.alias, result, tt.expected)
			}
		})
	}
}

func TestGetModelTier(t *testing.T) {
	tests := []struct {
		name     string
		modelID  string
		expected string
	}{
		{
			name:     "sonnet alias",
			modelID:  "claude-sonnet-4-5",
			expected: "sonnet",
		},
		{
			name:     "sonnet pinned",
			modelID:  "claude-sonnet-4-5-20250929",
			expected: "sonnet",
		},
		{
			name:     "haiku alias",
			modelID:  "claude-haiku-4-5",
			expected: "haiku",
		},
		{
			name:     "opus alias",
			modelID:  "claude-opus-4-6",
			expected: "opus",
		},
		{
			name:     "unknown model",
			modelID:  "gpt-4-turbo",
			expected: "unknown",
		},
		{
			name:     "empty string",
			modelID:  "",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetModelTier(tt.modelID)
			if result != tt.expected {
				t.Errorf("GetModelTier(%q) = %v, want %v", tt.modelID, result, tt.expected)
			}
		})
	}
}

func TestNewModelsClient(t *testing.T) {
	// Test with provided API key
	client := NewModelsClient("test-api-key")
	if client.apiKey != "test-api-key" {
		t.Errorf("apiKey = %v, want test-api-key", client.apiKey)
	}

	// Test with empty API key (should use env)
	client = NewModelsClient("")
	// Just verify it doesn't panic
	if client == nil {
		t.Error("NewModelsClient returned nil")
	}
}

func TestFormatModelList(t *testing.T) {
	models := []APIModel{
		{ID: "claude-sonnet-4-5-20250929", DisplayName: "Claude Sonnet 4.5"},
		{ID: "claude-haiku-4-5-20251001", DisplayName: "Claude Haiku 4.5"},
	}

	result := FormatModelList(models)

	if result == "" {
		t.Error("FormatModelList returned empty string")
	}

	// Check that model IDs are present
	if !containsString(result, "claude-sonnet-4-5-20250929") {
		t.Error("Result should contain claude-sonnet-4-5-20250929")
	}
	if !containsString(result, "claude-haiku-4-5-20251001") {
		t.Error("Result should contain claude-haiku-4-5-20251001")
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
