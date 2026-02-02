//go:build windows

package java

import (
	"os"
	"path/filepath"
	"regexp"
)

// detectPlatformSpecific finds Java installations on Windows
func detectPlatformSpecific() []JavaInstallation {
	var installations []JavaInstallation
	seen := make(map[int]bool)

	// Common Windows Java installation directories
	searchDirs := []string{
		`C:\Program Files\Java`,
		`C:\Program Files (x86)\Java`,
		`C:\Program Files\Eclipse Adoptium`,
		`C:\Program Files\Eclipse Foundation`,
		`C:\Program Files\Microsoft\jdk`,
		`C:\Program Files\Zulu`,
		`C:\Program Files\Amazon Corretto`,
		`C:\Program Files\BellSoft\LibericaJDK`,
	}

	// Also check LOCALAPPDATA for scoop installations
	if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
		searchDirs = append(searchDirs, filepath.Join(localAppData, "Programs", "Java"))
		searchDirs = append(searchDirs, filepath.Join(localAppData, "scoop", "apps"))
	}

	// Check each directory
	for _, dir := range searchDirs {
		if entries, err := os.ReadDir(dir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					javaPath := filepath.Join(dir, entry.Name())

					// For scoop apps, look for java subdirectories
					if filepath.Base(dir) == "apps" {
						if subEntries, err := os.ReadDir(javaPath); err == nil {
							for _, subEntry := range subEntries {
								if subEntry.IsDir() {
									subPath := filepath.Join(javaPath, subEntry.Name())
									if inst := checkJavaPathWindows(subPath, "scoop", seen); inst != nil {
										installations = append(installations, *inst)
									}
								}
							}
						}
						continue
					}

					if inst := checkJavaPathWindows(javaPath, "system", seen); inst != nil {
						installations = append(installations, *inst)
					}
				}
			}
		}
	}

	// Check for jabba installations
	if userProfile := os.Getenv("USERPROFILE"); userProfile != "" {
		jabbaDir := filepath.Join(userProfile, ".jabba", "jdk")
		if entries, err := os.ReadDir(jabbaDir); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					javaPath := filepath.Join(jabbaDir, entry.Name())
					if inst := checkJavaPathWindows(javaPath, "jabba", seen); inst != nil {
						installations = append(installations, *inst)
					}
				}
			}
		}
	}

	return installations
}

// checkJavaPathWindows checks a path for Java on Windows and adds to seen map
func checkJavaPathWindows(basePath, source string, seen map[int]bool) *JavaInstallation {
	// First try to detect version from directory name
	dirName := filepath.Base(basePath)

	// Common patterns in directory names
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`jdk-?(\d+)`),               // jdk-21, jdk21
		regexp.MustCompile(`java-?(\d+)`),              // java-17, java17
		regexp.MustCompile(`openjdk-?(\d+)`),           // openjdk-21
		regexp.MustCompile(`temurin-(\d+)`),            // temurin-21
		regexp.MustCompile(`zulu-?(\d+)`),              // zulu21, zulu-21
		regexp.MustCompile(`corretto-?(\d+)`),          // corretto-21
		regexp.MustCompile(`liberica-?jdk-?(\d+)`),     // liberica-jdk-21
		regexp.MustCompile(`microsoft-?(\d+)(?:-jdk)?`), // microsoft-21-jdk
	}

	for _, pattern := range patterns {
		if matches := pattern.FindStringSubmatch(dirName); len(matches) > 1 {
			major, _, _ := parseVersionNumber(matches[1])
			if major > 0 && !seen[major] && IsVersionCompatible(major) {
				// Verify the installation exists
				javaBin := filepath.Join(basePath, "bin", "java.exe")
				if _, err := os.Stat(javaBin); err == nil {
					seen[major] = true
					return &JavaInstallation{
						Version:     major,
						VersionFull: matches[1],
						Path:        basePath,
						Source:      source,
					}
				}
			}
		}
	}

	// Fallback: try running java -version
	if inst := checkJavaPath(basePath, source); inst != nil {
		if !seen[inst.Version] && IsVersionCompatible(inst.Version) {
			seen[inst.Version] = true
			return inst
		}
	}

	return nil
}
