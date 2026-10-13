package java

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestCIMatrixMatchesSupportedVersions pins the CI archetype matrix to
// SupportedVersions. The two lists have drifted before (CI kept testing a
// version the CLI no longer accepted, and vice versa); this makes the
// drift a build failure instead of a silent gap in coverage.
func TestCIMatrixMatchesSupportedVersions(t *testing.T) {
	path := filepath.Join("..", "..", ".github", "workflows", "ci.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("ci.yml not available: %v", err)
	}
	re := regexp.MustCompile(`(?m)^\s*java:\s*\[([^\]]+)\]`)
	m := re.FindStringSubmatch(string(data))
	if m == nil {
		t.Fatalf("no `java: [...]` matrix line found in %s", path)
	}
	var got []string
	for _, v := range strings.Split(m[1], ",") {
		got = append(got, strings.Trim(strings.TrimSpace(v), `'"`))
	}
	want := make([]string, 0, len(SupportedVersions))
	for _, v := range SupportedVersions {
		want = append(want, formatInt(v))
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("ci.yml archetype matrix java=%v but SupportedVersions=%v — keep them identical", got, want)
	}
}
