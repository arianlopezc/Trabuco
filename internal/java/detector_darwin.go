//go:build darwin

package java

import (
	"os/exec"
	"regexp"
	"strings"
)

// detectPlatformSpecific finds Java installations on macOS
func detectPlatformSpecific() []JavaInstallation {
	var installations []JavaInstallation
	seen := make(map[int]bool)

	// Use /usr/libexec/java_home -V to list all installed JVMs
	cmd := exec.Command("/usr/libexec/java_home", "-V")
	output, err := cmd.CombinedOutput() // -V outputs to stderr
	if err == nil {
		installations = append(installations, parseJavaHomeOutput(string(output), seen)...)
	}

	return installations
}

// parseJavaHomeOutput parses the output of java_home -V
// Output format example:
// Matching Java Virtual Machines (3):
//
//	21.0.1 (x86_64) "Oracle Corporation" - "Java SE 21.0.1" /Library/Java/JavaVirtualMachines/jdk-21.jdk/Contents/Home
//	17.0.9 (x86_64) "Eclipse Adoptium" - "OpenJDK 17.0.9" /Library/Java/JavaVirtualMachines/temurin-17.jdk/Contents/Home
//	11.0.21 (x86_64) "Eclipse Adoptium" - "OpenJDK 11.0.21" /Library/Java/JavaVirtualMachines/temurin-11.jdk/Contents/Home
func parseJavaHomeOutput(output string, seen map[int]bool) []JavaInstallation {
	var installations []JavaInstallation

	// Pattern to match version lines
	// Example: "    21.0.1 (x86_64) \"Oracle Corporation\" - \"Java SE 21.0.1\" /Library/..."
	linePattern := regexp.MustCompile(`^\s*(\d+[\.\d]*)\s+\([^)]+\)\s+"[^"]+"\s+-\s+"[^"]+"\s+(.+)$`)

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		matches := linePattern.FindStringSubmatch(line)
		if len(matches) >= 3 {
			versionStr := matches[1]
			path := strings.TrimSpace(matches[2])

			major, full, err := parseVersionNumber(versionStr)
			if err != nil || major == 0 {
				continue
			}

			if !seen[major] && IsVersionCompatible(major) {
				installations = append(installations, JavaInstallation{
					Version:     major,
					VersionFull: full,
					Path:        path,
					Source:      "java_home",
				})
				seen[major] = true
			}
		}
	}

	return installations
}
