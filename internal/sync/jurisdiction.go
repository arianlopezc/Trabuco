package sync

import (
	"path/filepath"
	"strings"
)

// Jurisdiction defines which paths sync is allowed to touch. Paths outside
// this allow-list are filtered out silently during planning and blocked at
// write time as a defensive safeguard.
//
// The list is deliberately conservative: AI-tooling files only. Business
// code (Java sources, POMs, migrations, application.yml) and infrastructure
// (docker-compose.yml, CI workflows other than copilot-setup-steps.yml) are
// explicitly out of scope — users who want those refreshed must regenerate
// their project or edit manually.
//
// Any change to this list is a sync jurisdiction change and requires
// coordinated updates to the tests in sync_test.go and the README.

// allowedPrefixes lists path prefixes (relative to project root) that sync
// may read and write. A path is in-jurisdiction if it has one of these
// prefixes OR matches exactly in allowedExact.
var allowedPrefixes = []string{
	".ai/",
	".claude/",
	".cursor/",
	".codex/",
	".agents/",
	".github/instructions/",
	".github/skills/",
	".github/scripts/review-checks.sh",
	".github/workflows/copilot-setup-steps.yml",
	".github/copilot-instructions.md",
	".trabuco/review.config.json",
}

// allowedExact lists top-level files (no trailing slash) that sync handles.
var allowedExact = map[string]bool{
	"CLAUDE.md": true,
	"AGENTS.md": true,
}

// excludedPaths lists paths within otherwise-allowed prefixes that sync
// must still ignore. These are session state or live data, not templates.
var excludedPaths = map[string]bool{
	".ai/checkpoint.json": true,
}

// InJurisdiction reports whether sync is permitted to act on the given
// project-relative path. Returns false for any path outside the allow-list
// and for explicitly-excluded paths within allowed prefixes.
//
// The input must be a forward-slash, project-relative path (the form used
// by filepath.ToSlash). Absolute paths, ../ traversal attempts, and paths
// beginning with a separator all return false.
func InJurisdiction(relPath string) bool {
	if relPath == "" {
		return false
	}
	// Reject absolute paths and traversal attempts. filepath.Clean on a
	// relative path that escapes the root produces a string starting with
	// "..", which we reject defensively.
	if filepath.IsAbs(relPath) {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(relPath))
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." {
		return false
	}
	if strings.HasPrefix(clean, "/") {
		return false
	}
	if excludedPaths[clean] {
		return false
	}
	if allowedExact[clean] {
		return true
	}
	for _, prefix := range allowedPrefixes {
		// Exact-file entries like ".github/scripts/review-checks.sh" have
		// no trailing slash. Treat them as exact matches.
		if !strings.HasSuffix(prefix, "/") {
			if clean == prefix {
				return true
			}
			continue
		}
		if strings.HasPrefix(clean, prefix) {
			return true
		}
	}
	return false
}
