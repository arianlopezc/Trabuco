package java

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// JavaInstallation represents a detected Java installation
type JavaInstallation struct {
	Version     int    // Major version (21, 17, etc.)
	VersionFull string // Full version string
	Path        string // Installation path
	Source      string // Detection source (e.g., "JAVA_HOME", "sdkman", "system")
	IsDefault   bool   // Whether this is the default/current Java
}

// DetectionResult holds all detected Java installations
type DetectionResult struct {
	Installations  []JavaInstallation
	DefaultVersion int
}

// Detect discovers all Java installations on the system
func Detect() *DetectionResult {
	result := &DetectionResult{
		Installations: []JavaInstallation{},
	}

	seen := make(map[int]bool)

	// 1. Check JAVA_HOME environment variable
	if javaHome := os.Getenv("JAVA_HOME"); javaHome != "" {
		if inst := checkJavaPath(javaHome, "JAVA_HOME"); inst != nil {
			inst.IsDefault = true
			if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
				result.Installations = append(result.Installations, *inst)
				result.DefaultVersion = inst.Version
				seen[inst.Version] = true
			}
		}
	}

	// 2. Check java -version (current java command)
	if inst := detectFromJavaCommand(); inst != nil {
		if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
			if result.DefaultVersion == 0 {
				inst.IsDefault = true
				result.DefaultVersion = inst.Version
			}
			if !seen[inst.Version] {
				result.Installations = append(result.Installations, *inst)
				seen[inst.Version] = true
			}
		}
	}

	// 3. Check SDKMAN
	if home := os.Getenv("HOME"); home != "" {
		sdkmanDir := filepath.Join(home, ".sdkman", "candidates", "java")
		if entries, err := os.ReadDir(sdkmanDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() && entry.Name() != "current" {
					javaPath := filepath.Join(sdkmanDir, entry.Name())
					if inst := checkJavaPath(javaPath, "sdkman"); inst != nil {
						if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
							result.Installations = append(result.Installations, *inst)
							seen[inst.Version] = true
						}
					}
				}
			}
		}
	}

	// 4. Platform-specific detection
	platformInstallations := detectPlatformSpecific()
	for _, inst := range platformInstallations {
		if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
			result.Installations = append(result.Installations, inst)
			seen[inst.Version] = true
		}
	}

	// Sort by version descending
	sort.Slice(result.Installations, func(i, j int) bool {
		return result.Installations[i].Version > result.Installations[j].Version
	})

	return result
}

// IsVersionDetected checks if a specific major version is installed
func (r *DetectionResult) IsVersionDetected(version int) bool {
	for _, inst := range r.Installations {
		if inst.Version == version {
			return true
		}
	}
	return false
}

// GetDetectedVersions returns all detected major versions
func (r *DetectionResult) GetDetectedVersions() []int {
	versions := make([]int, 0, len(r.Installations))
	for _, inst := range r.Installations {
		versions = append(versions, inst.Version)
	}
	return versions
}

// GetCompatibleVersions returns only compatible versions (>= 17)
func (r *DetectionResult) GetCompatibleVersions() []int {
	versions := make([]int, 0, len(r.Installations))
	for _, inst := range r.Installations {
		if IsVersionCompatible(inst.Version) {
			versions = append(versions, inst.Version)
		}
	}
	return versions
}

// detectFromJavaCommand runs java -version and parses the output
func detectFromJavaCommand() *JavaInstallation {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return nil
	}

	major, full, err := ParseVersion(lines[0])
	if err != nil || major == 0 {
		return nil
	}

	// Try to get the java path
	path := ""
	if whichCmd := exec.Command("which", "java"); whichCmd != nil {
		if whichOutput, err := whichCmd.Output(); err == nil {
			path = strings.TrimSpace(string(whichOutput))
			// Resolve symlink if possible
			if resolved, err := filepath.EvalSymlinks(path); err == nil {
				path = resolved
			}
		}
	}

	return &JavaInstallation{
		Version:     major,
		VersionFull: full,
		Path:        path,
		Source:      "system",
	}
}

// checkJavaPath checks if a path contains a valid Java installation
func checkJavaPath(basePath, source string) *JavaInstallation {
	// Look for java executable
	javaBin := filepath.Join(basePath, "bin", "java")
	if _, err := os.Stat(javaBin); os.IsNotExist(err) {
		// Try macOS Contents/Home structure
		javaBin = filepath.Join(basePath, "Contents", "Home", "bin", "java")
		if _, err := os.Stat(javaBin); os.IsNotExist(err) {
			return nil
		}
	}

	// Run java -version
	cmd := exec.Command(javaBin, "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return nil
	}

	major, full, err := ParseVersion(lines[0])
	if err != nil || major == 0 {
		return nil
	}

	return &JavaInstallation{
		Version:     major,
		VersionFull: full,
		Path:        basePath,
		Source:      source,
	}
}

// FormatDetectedVersions formats detected versions for display
func FormatDetectedVersions(versions []int) string {
	if len(versions) == 0 {
		return "none"
	}
	strs := make([]string, len(versions))
	for i, v := range versions {
		strs[i] = string(rune('0'+v/10)) + string(rune('0'+v%10))
		if v < 10 {
			strs[i] = string(rune('0' + v))
		}
	}
	// Use proper string formatting
	result := ""
	for i, v := range versions {
		if i > 0 {
			result += ", "
		}
		result += formatInt(v)
	}
	return result
}

// formatInt converts an int to a string without using strconv
func formatInt(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	return result
}
