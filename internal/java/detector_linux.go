//go:build linux

package java

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// detectPlatformSpecific finds Java installations on Linux
func detectPlatformSpecific() []JavaInstallation {
	var installations []JavaInstallation
	seen := make(map[int]bool)

	// 1. Check /usr/lib/jvm/ directory (common location for system JVMs)
	jvmDir := "/usr/lib/jvm"
	if entries, err := os.ReadDir(jvmDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				javaPath := filepath.Join(jvmDir, entry.Name())
				if inst := checkJavaPath(javaPath, "system"); inst != nil {
					if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
						installations = append(installations, *inst)
						seen[inst.Version] = true
					}
				}
			}
		}
	}

	// 2. Check update-alternatives for java
	if inst := checkUpdateAlternatives(seen); inst != nil {
		if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
			installations = append(installations, *inst)
			seen[inst.Version] = true
		}
	}

	// 3. Check common installation directories
	commonDirs := []string{
		"/opt/java",
		"/opt/jdk",
		"/usr/java",
	}

	for _, dir := range commonDirs {
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					javaPath := filepath.Join(dir, entry.Name())
					if inst := checkJavaPath(javaPath, "system"); inst != nil {
						if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
							installations = append(installations, *inst)
							seen[inst.Version] = true
						}
					}
				}
			}
		}
	}

	// 4. Check ASDF installations
	if home := os.Getenv("HOME"); home != "" {
		asdfDir := filepath.Join(home, ".asdf", "installs", "java")
		if entries, err := os.ReadDir(asdfDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					javaPath := filepath.Join(asdfDir, entry.Name())
					if inst := checkJavaPath(javaPath, "asdf"); inst != nil {
						if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
							installations = append(installations, *inst)
							seen[inst.Version] = true
						}
					}
				}
			}
		}
	}

	return installations
}

// checkUpdateAlternatives uses update-alternatives to find system Java
func checkUpdateAlternatives(seen map[int]bool) *JavaInstallation {
	cmd := exec.Command("update-alternatives", "--display", "java")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	// Look for current best version line
	// Example: "link best version is /usr/lib/jvm/java-21-openjdk-amd64/bin/java"
	bestPattern := regexp.MustCompile(`link best version is ([^\s]+)`)
	matches := bestPattern.FindStringSubmatch(string(output))
	if len(matches) < 2 {
		return nil
	}

	javaPath := matches[1]
	// Get the JVM home directory (go up from bin/java)
	jvmHome := filepath.Dir(filepath.Dir(javaPath))

	// Check this path
	if inst := checkJavaPath(jvmHome, "update-alternatives"); inst != nil {
		return inst
	}

	// Try to parse version from path (e.g., java-21-openjdk)
	pathPattern := regexp.MustCompile(`java-(\d+)`)
	if matches := pathPattern.FindStringSubmatch(javaPath); len(matches) > 1 {
		major, _, _ := parseVersionNumber(matches[1])
		if major > 0 && !seen[major] {
			return &JavaInstallation{
				Version:     major,
				VersionFull: matches[1],
				Path:        jvmHome,
				Source:      "update-alternatives",
			}
		}
	}

	return nil
}
