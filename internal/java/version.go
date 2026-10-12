package java

import (
	"regexp"
	"strconv"
	"strings"
)

// MinSupportedVersion is the minimum Java version supported by Trabuco
const MinSupportedVersion = 21

// SupportedVersions lists all Java versions supported by this tool.
//
// Policy: every LTS release from MinSupportedVersion up, plus the feature
// releases shipped since the last LTS. The list is the single source of
// truth — the `--java-version` flag help, the MCP `java_version` parameter,
// the interactive menu and the CI archetype matrix are all derived from (or
// asserted against) it, so adding a version here is the only code change
// needed to accept it.
//
// Why 25+ is allowed again (1.16): the templates moved from Spring Boot
// 3.4.2 to 3.5.16 (Spring Framework 6.2.19). 3.4.x's repackaged ASM
// rejected class-file major version 69 (Java 25) — projects compiled but
// crashed at component scan with "Unsupported class file major version
// 69". Framework 6.2.13+ accepts newer class files leniently, and every
// other bytecode-reading tool in the templates (Mockito/Byte Buddy,
// JaCoCo, ArchUnit, JobRunr) is pinned to a release that knows Java 27.
// Spring Boot 3.5 officially supports Java 17–25; 26 and 27 work on a
// best-effort basis and are exercised by the CI archetype matrix.
var SupportedVersions = []int{21, 24, 25, 26, 27}

// ParseVersion extracts the major version and full version string from various formats.
// Handles formats like:
//   - "21.0.2"
//   - "1.8.0_312" (legacy format, returns 8)
//   - "openjdk version \"21.0.1\""
//   - "java version \"17.0.1\""
func ParseVersion(s string) (major int, full string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, "", nil
	}

	// Pattern for version strings in output like: openjdk version "21.0.1" or java version "17.0.1"
	quotedPattern := regexp.MustCompile(`(?:openjdk|java)\s+version\s+"([^"]+)"`)
	if matches := quotedPattern.FindStringSubmatch(s); len(matches) > 1 {
		return parseVersionNumber(matches[1])
	}

	// Try to parse as a direct version number
	return parseVersionNumber(s)
}

// parseVersionNumber parses a version number string like "21.0.2" or "1.8.0_312"
func parseVersionNumber(s string) (major int, full string, err error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, "", nil
	}

	full = s

	// Handle legacy 1.x format (e.g., "1.8.0_312" -> major 8)
	if strings.HasPrefix(s, "1.") {
		parts := strings.Split(s, ".")
		if len(parts) >= 2 {
			if v, err := strconv.Atoi(parts[1]); err == nil {
				return v, full, nil
			}
		}
	}

	// Extract first numeric segment for modern format (e.g., "21.0.2" -> 21)
	numPattern := regexp.MustCompile(`^(\d+)`)
	if matches := numPattern.FindStringSubmatch(s); len(matches) > 1 {
		if v, err := strconv.Atoi(matches[1]); err == nil {
			return v, full, nil
		}
	}

	return 0, full, nil
}

// IsVersionCompatible checks if a Java version is compatible with Spring Boot 3.x
func IsVersionCompatible(version int) bool {
	return version >= MinSupportedVersion
}

// IsSupportedVersion checks if a version is in our supported list
func IsSupportedVersion(version int) bool {
	for _, v := range SupportedVersions {
		if v == version {
			return true
		}
	}
	return false
}
