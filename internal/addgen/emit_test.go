package addgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmitFile_WritesContent(t *testing.T) {
	dir := t.TempDir()
	ctx := &Context{ProjectPath: dir}
	r := &Result{}

	if err := ctx.emitFile("a/b/c.txt", "hello", r); err != nil {
		t.Fatalf("emitFile: %v", err)
	}
	if got := []string{"a/b/c.txt"}; !sliceEq(r.Created, got) {
		t.Errorf("Created = %v, want %v", r.Created, got)
	}
	body, err := os.ReadFile(filepath.Join(dir, "a/b/c.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != "hello" {
		t.Errorf("file body = %q, want %q", body, "hello")
	}
}

func TestEmitFile_RefusesClobber(t *testing.T) {
	dir := t.TempDir()
	pre := filepath.Join(dir, "x.txt")
	if err := os.WriteFile(pre, []byte("pre-existing"), 0o644); err != nil {
		t.Fatal(err)
	}
	ctx := &Context{ProjectPath: dir}

	err := ctx.emitFile("x.txt", "new", &Result{})
	if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
		t.Fatalf("expected refusing-to-overwrite error, got %v", err)
	}

	// Verify pre-existing content is intact.
	body, _ := os.ReadFile(pre)
	if string(body) != "pre-existing" {
		t.Errorf("file was corrupted: %q", body)
	}
}

func TestEmitFile_DryRunDoesNotWrite(t *testing.T) {
	dir := t.TempDir()
	ctx := &Context{ProjectPath: dir, DryRun: true}
	r := &Result{}

	if err := ctx.emitFile("a/b/c.txt", "hello", r); err != nil {
		t.Fatalf("emitFile: %v", err)
	}
	if len(r.Created) != 1 || r.Created[0] != "a/b/c.txt" {
		t.Errorf("Created = %v, want [a/b/c.txt]", r.Created)
	}
	if _, err := os.Stat(filepath.Join(dir, "a/b/c.txt")); !os.IsNotExist(err) {
		t.Errorf("expected file NOT to exist in dry-run, stat err = %v", err)
	}
}

func sliceEq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
