package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

func TestParseLineRange(t *testing.T) {
	cases := []struct {
		spec       string
		start, end int
		wantErr    bool
	}{
		{"5-10", 5, 10, false},
		{"42", 42, 42, false},
		{"5 - 10", 5, 10, false},
		{"abc", 0, 0, true},
		{"5-abc", 0, 0, true},
		{"", 0, 0, true},
	}
	for _, c := range cases {
		s, e, err := parseLineRange(c.spec)
		if (err != nil) != c.wantErr {
			t.Errorf("parseLineRange(%q): err = %v, wantErr = %v", c.spec, err, c.wantErr)
			continue
		}
		if !c.wantErr && (s != c.start || e != c.end) {
			t.Errorf("parseLineRange(%q) = (%d,%d), want (%d,%d)", c.spec, s, e, c.start, c.end)
		}
	}
}

func TestVerifyEvidence_Success(t *testing.T) {
	dir := t.TempDir()
	content := "line1\nline2\nline3\nline4\n"
	if err := os.WriteFile(filepath.Join(dir, "file.java"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	excerpt := "line2\nline3"
	sum := sha256.Sum256([]byte(excerpt))
	ev := &types.SourceEvidence{
		File:        "file.java",
		Lines:       "2-3",
		ContentHash: hex.EncodeToString(sum[:]),
	}
	if err := VerifyEvidence(dir, ev); err != nil {
		t.Errorf("VerifyEvidence on valid evidence: %v", err)
	}
}

func TestVerifyEvidence_Failures(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "file.java"), []byte("line1\nline2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		ev   *types.SourceEvidence
		want string
	}{
		{"nil", nil, "is nil"},
		{"empty file", &types.SourceEvidence{Lines: "1"}, "file is empty"},
		{"missing file", &types.SourceEvidence{File: "missing.java", Lines: "1"}, "not readable"},
		{"out of range", &types.SourceEvidence{File: "file.java", Lines: "1-100", ContentHash: "x"}, "out of range"},
		{"bad hash", &types.SourceEvidence{File: "file.java", Lines: "1-2", ContentHash: "deadbeef"}, "mismatch"},
	}
	for _, c := range cases {
		err := VerifyEvidence(dir, c.ev)
		if err == nil {
			t.Errorf("%s: expected error, got nil", c.name)
			continue
		}
		if !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: error = %q, want substring %q", c.name, err.Error(), c.want)
		}
	}
}
