package java

import (
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantMajor   int
		wantFull    string
		shouldParse bool
	}{
		{
			name:        "simple version 21",
			input:       "21.0.2",
			wantMajor:   21,
			wantFull:    "21.0.2",
			shouldParse: true,
		},
		{
			name:        "simple version 17",
			input:       "17.0.9",
			wantMajor:   17,
			wantFull:    "17.0.9",
			shouldParse: true,
		},
		{
			name:        "legacy 1.8 format",
			input:       "1.8.0_312",
			wantMajor:   8,
			wantFull:    "1.8.0_312",
			shouldParse: true,
		},
		{
			name:        "openjdk version quoted",
			input:       `openjdk version "21.0.1"`,
			wantMajor:   21,
			wantFull:    "21.0.1",
			shouldParse: true,
		},
		{
			name:        "java version quoted",
			input:       `java version "17.0.1"`,
			wantMajor:   17,
			wantFull:    "17.0.1",
			shouldParse: true,
		},
		{
			name:        "version 25",
			input:       "25.0.1",
			wantMajor:   25,
			wantFull:    "25.0.1",
			shouldParse: true,
		},
		{
			name:        "empty string",
			input:       "",
			wantMajor:   0,
			wantFull:    "",
			shouldParse: false,
		},
		{
			name:        "whitespace",
			input:       "   ",
			wantMajor:   0,
			wantFull:    "",
			shouldParse: false,
		},
		{
			name:        "just major version",
			input:       "21",
			wantMajor:   21,
			wantFull:    "21",
			shouldParse: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			major, full, err := ParseVersion(tt.input)
			if err != nil {
				t.Errorf("ParseVersion() unexpected error: %v", err)
				return
			}
			if major != tt.wantMajor {
				t.Errorf("ParseVersion() major = %d, want %d", major, tt.wantMajor)
			}
			if tt.shouldParse && full != tt.wantFull {
				t.Errorf("ParseVersion() full = %s, want %s", full, tt.wantFull)
			}
		})
	}
}

func TestIsVersionCompatible(t *testing.T) {
	tests := []struct {
		version    int
		compatible bool
	}{
		{8, false},
		{11, false},
		{14, false},
		{16, false},
		{17, true},
		{18, true},
		{21, true},
		{25, true},
	}

	for _, tt := range tests {
		t.Run(formatInt(tt.version), func(t *testing.T) {
			got := IsVersionCompatible(tt.version)
			if got != tt.compatible {
				t.Errorf("IsVersionCompatible(%d) = %v, want %v", tt.version, got, tt.compatible)
			}
		})
	}
}

func TestIsSupportedVersion(t *testing.T) {
	tests := []struct {
		version   int
		supported bool
	}{
		{17, true},
		{21, true},
		{25, true},
		{18, false},
		{19, false},
		{11, false},
	}

	for _, tt := range tests {
		t.Run(formatInt(tt.version), func(t *testing.T) {
			got := IsSupportedVersion(tt.version)
			if got != tt.supported {
				t.Errorf("IsSupportedVersion(%d) = %v, want %v", tt.version, got, tt.supported)
			}
		})
	}
}

func TestDetectionResult_IsVersionDetected(t *testing.T) {
	result := &DetectionResult{
		Installations: []JavaInstallation{
			{Version: 21, VersionFull: "21.0.1", Path: "/path/to/21", Source: "test"},
			{Version: 17, VersionFull: "17.0.9", Path: "/path/to/17", Source: "test"},
		},
		DefaultVersion: 21,
	}

	if !result.IsVersionDetected(21) {
		t.Error("IsVersionDetected(21) should return true")
	}
	if !result.IsVersionDetected(17) {
		t.Error("IsVersionDetected(17) should return true")
	}
	if result.IsVersionDetected(25) {
		t.Error("IsVersionDetected(25) should return false")
	}
	if result.IsVersionDetected(11) {
		t.Error("IsVersionDetected(11) should return false")
	}
}

func TestDetectionResult_GetDetectedVersions(t *testing.T) {
	result := &DetectionResult{
		Installations: []JavaInstallation{
			{Version: 21},
			{Version: 17},
		},
	}

	versions := result.GetDetectedVersions()
	if len(versions) != 2 {
		t.Errorf("GetDetectedVersions() returned %d versions, want 2", len(versions))
	}

	// Check both versions are present
	found21, found17 := false, false
	for _, v := range versions {
		if v == 21 {
			found21 = true
		}
		if v == 17 {
			found17 = true
		}
	}
	if !found21 || !found17 {
		t.Error("GetDetectedVersions() should contain 21 and 17")
	}
}

func TestDetectionResult_GetCompatibleVersions(t *testing.T) {
	result := &DetectionResult{
		Installations: []JavaInstallation{
			{Version: 21},
			{Version: 17},
			{Version: 11}, // Not compatible
		},
	}

	versions := result.GetCompatibleVersions()
	if len(versions) != 2 {
		t.Errorf("GetCompatibleVersions() returned %d versions, want 2", len(versions))
	}

	// Should only contain 21 and 17, not 11
	for _, v := range versions {
		if v == 11 {
			t.Error("GetCompatibleVersions() should not contain version 11")
		}
	}
}

func TestFormatDetectedVersions(t *testing.T) {
	tests := []struct {
		name     string
		versions []int
		want     string
	}{
		{"empty", []int{}, "none"},
		{"single", []int{21}, "21"},
		{"multiple", []int{21, 17}, "21, 17"},
		{"three", []int{25, 21, 17}, "25, 21, 17"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatDetectedVersions(tt.versions)
			if got != tt.want {
				t.Errorf("FormatDetectedVersions() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatInt(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0"},
		{1, "1"},
		{17, "17"},
		{21, "21"},
		{25, "25"},
		{100, "100"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatInt(tt.input)
			if got != tt.want {
				t.Errorf("formatInt(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
