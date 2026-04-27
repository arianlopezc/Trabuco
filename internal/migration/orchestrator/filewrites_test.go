package orchestrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

func TestApplyFileWrites_CreateReplaceDelete(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := &specialists.Output{
		Items: []types.OutputItem{
			{
				ID:    "create-new",
				State: types.ItemApplied,
				FileWrites: []types.FileWrite{
					{Path: "new.txt", Operation: types.OpCreate, Content: "fresh"},
				},
			},
			{
				ID:    "replace-existing",
				State: types.ItemApplied,
				FileWrites: []types.FileWrite{
					{Path: "existing.txt", Operation: types.OpReplace, Content: "new"},
				},
			},
			{
				ID:    "delete-and-recreate",
				State: types.ItemApplied,
				FileWrites: []types.FileWrite{
					{Path: "existing.txt", Operation: types.OpDelete},
				},
			},
		},
	}
	if err := applyFileWrites(dir, out); err != nil {
		t.Fatalf("applyFileWrites: %v", err)
	}
	if data, _ := os.ReadFile(filepath.Join(dir, "new.txt")); string(data) != "fresh" {
		t.Errorf("new.txt = %q, want fresh", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "existing.txt")); !os.IsNotExist(err) {
		t.Errorf("existing.txt should be deleted")
	}
}

func TestApplyFileWrites_NestedDirCreate(t *testing.T) {
	dir := t.TempDir()
	out := &specialists.Output{
		Items: []types.OutputItem{
			{
				ID:    "deep-create",
				State: types.ItemApplied,
				FileWrites: []types.FileWrite{
					{Path: "model/src/main/java/com/x/User.java", Operation: types.OpCreate, Content: "package com.x;\nclass User {}\n"},
				},
			},
		},
	}
	if err := applyFileWrites(dir, out); err != nil {
		t.Fatalf("applyFileWrites: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "model/src/main/java/com/x/User.java")); err != nil {
		t.Errorf("nested file not created: %v", err)
	}
}

func TestApplyFileWrites_RollbackOnLaterFailure(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "existing.txt"), []byte("original"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := &specialists.Output{
		Items: []types.OutputItem{
			{
				ID:    "atomic-batch",
				State: types.ItemApplied,
				FileWrites: []types.FileWrite{
					{Path: "existing.txt", Operation: types.OpReplace, Content: "modified"},
					{Path: "../escape.txt", Operation: types.OpCreate, Content: "fails — path traversal forbidden"},
				},
			},
		},
	}
	if err := applyFileWrites(dir, out); err == nil {
		t.Error("expected error on second write (path traversal)")
	}
	// First write should be rolled back.
	if data, _ := os.ReadFile(filepath.Join(dir, "existing.txt")); string(data) != "original" {
		t.Errorf("existing.txt = %q after rollback, want original", data)
	}
}

func TestApplyFileWrites_NotApplicable_Skipped(t *testing.T) {
	dir := t.TempDir()
	out := &specialists.Output{
		Items: []types.OutputItem{
			{ID: "skip", State: types.ItemNotApplicable, Reason: "no work"},
		},
	}
	if err := applyFileWrites(dir, out); err != nil {
		t.Errorf("not_applicable items should be no-ops: %v", err)
	}
}

func TestSafeJoin_RejectsTraversal(t *testing.T) {
	dir := "/tmp/repo"
	cases := []struct {
		path    string
		wantErr string
	}{
		{"../escape", "traversal"},
		{"foo/../bar", "traversal"},
		{"/abs/path", "absolute"},
		{"", "empty"},
	}
	for _, c := range cases {
		_, err := safeJoin(dir, c.path)
		if err == nil {
			t.Errorf("safeJoin(%q): expected error", c.path)
			continue
		}
		if !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("safeJoin(%q): err = %v, want substring %q", c.path, err, c.wantErr)
		}
	}
}

func TestSafeJoin_AllowsCleanRelativePath(t *testing.T) {
	dir := t.TempDir()
	got, err := safeJoin(dir, "src/main/java/Foo.java")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "src/main/java/Foo.java")
	if got != want {
		t.Errorf("safeJoin = %q, want %q", got, want)
	}
}
