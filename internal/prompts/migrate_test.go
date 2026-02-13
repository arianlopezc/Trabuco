package prompts

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateSourcePath(t *testing.T) {
	// Create a temp directory with pom.xml for valid tests
	tempDir := t.TempDir()
	pomPath := filepath.Join(tempDir, "pom.xml")
	if err := os.WriteFile(pomPath, []byte("<project></project>"), 0644); err != nil {
		t.Fatalf("Failed to create test pom.xml: %v", err)
	}

	// Create a file (not directory) for test
	tempFile := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(tempFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create directory without pom.xml
	noPomDir := t.TempDir()

	tests := []struct {
		name      string
		input     interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid Maven project",
			input:     tempDir,
			wantError: false,
		},
		{
			name:      "empty path",
			input:     "",
			wantError: true,
			errorMsg:  "path is required",
		},
		{
			name:      "non-existent directory",
			input:     "/path/that/does/not/exist",
			wantError: true,
			errorMsg:  "directory does not exist",
		},
		{
			name:      "path is a file not directory",
			input:     tempFile,
			wantError: true,
			errorMsg:  "path is not a directory",
		},
		{
			name:      "directory without pom.xml",
			input:     noPomDir,
			wantError: true,
			errorMsg:  "no pom.xml found - is this a Maven project?",
		},
		{
			name:      "invalid input type",
			input:     123,
			wantError: true,
			errorMsg:  "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSourcePath(tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestValidateAPIKey(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid API key",
			input:     "sk-ant-api03-xxxxxxxxxxxxxxxxx",
			wantError: false,
		},
		{
			name:      "empty API key",
			input:     "",
			wantError: true,
			errorMsg:  "API key is required",
		},
		{
			name:      "too short API key",
			input:     "short",
			wantError: true,
			errorMsg:  "API key seems too short",
		},
		{
			name:      "minimum valid length",
			input:     "1234567890",
			wantError: false,
		},
		{
			name:      "invalid input type",
			input:     123,
			wantError: true,
			errorMsg:  "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIKey(tt.input)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				} else if tt.errorMsg != "" && err.Error() != tt.errorMsg {
					t.Errorf("Expected error '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestMaskAPIKey(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "normal key",
			key:      "sk-ant-api03-xxxxxxxxxxxxx",
			expected: "sk-a...xxxx",
		},
		{
			name:     "short key",
			key:      "short",
			expected: "****",
		},
		{
			name:     "exactly 8 chars",
			key:      "12345678",
			expected: "****",
		},
		{
			name:     "9 chars",
			key:      "123456789",
			expected: "1234...6789",
		},
		{
			name:     "empty key",
			key:      "",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskAPIKey(tt.key)
			if result != tt.expected {
				t.Errorf("maskAPIKey(%q) = %q, want %q", tt.key, result, tt.expected)
			}
		})
	}
}

func TestMigrateConfigFields(t *testing.T) {
	// Test that MigrateConfig has all required fields
	cfg := &MigrateConfig{
		SourcePath:   "/path/to/source",
		OutputPath:   "/path/to/output",
		Provider:     "anthropic",
		APIKey:       "test-key",
		Model:        "claude-sonnet-4-5",
		DryRun:       true,
		Interactive:  true,
		IncludeTests: false,
		SkipBuild:    false,
	}

	if cfg.SourcePath != "/path/to/source" {
		t.Error("SourcePath not set correctly")
	}
	if cfg.Provider != "anthropic" {
		t.Error("Provider not set correctly")
	}
	if cfg.Model != "claude-sonnet-4-5" {
		t.Error("Model not set correctly")
	}
	if !cfg.DryRun {
		t.Error("DryRun not set correctly")
	}
	if !cfg.Interactive {
		t.Error("Interactive not set correctly")
	}

	// Test ProjectInfo and DependencyReport can be set
	cfg.ProjectInfo = nil
	cfg.DependencyReport = nil
}
