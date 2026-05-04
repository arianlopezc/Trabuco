package addgen

import (
	"errors"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnakeCase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"add_orders", "add_orders"},
		{"Add Orders", "add_orders"},
		{"Add Orders Table", "add_orders_table"},
		{"add-orders", "add_orders"},
		{"AddOrdersTable", "addorderstable"},     // no word-boundary detection in PascalCase — feature, not bug; keeps the helper minimal
		{"  spaced  out  ", "spaced_out"},
		{"weird@chars!", "weirdchars"},
		{"v3 migration", "v3_migration"},
		{"", ""},
		{"___", ""},
	}
	for _, tc := range cases {
		got := snakeCase(tc.in)
		if got != tc.want {
			t.Errorf("snakeCase(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNextMigrationVersion(t *testing.T) {
	cases := []struct {
		name  string
		files []string
		want  int
	}{
		{"missing dir", nil, 1},
		{"empty dir", []string{}, 1},
		{"only V1", []string{"V1__baseline.sql"}, 2},
		{"V1+V2+V3", []string{"V1__baseline.sql", "V2__add_orders.sql", "V3__add_users.sql"}, 4},
		{"gap", []string{"V1__baseline.sql", "V5__add_users.sql"}, 6},
		{"ignore non-Vfiles", []string{"V1__baseline.sql", "README.md", "rollback.sql"}, 2},
		{"ignore subdirs and broken names", []string{"V1__baseline.sql", "Vbad__x.sql"}, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			migrationDir := filepath.Join(dir, "migration")
			if tc.files != nil {
				if err := os.MkdirAll(migrationDir, 0o755); err != nil {
					t.Fatal(err)
				}
				for _, f := range tc.files {
					if err := os.WriteFile(filepath.Join(migrationDir, f), []byte(""), 0o644); err != nil {
						t.Fatal(err)
					}
				}
			}
			n, err := nextMigrationVersion(migrationDir)
			if err != nil {
				t.Fatalf("nextMigrationVersion: %v", err)
			}
			if n != tc.want {
				t.Errorf("nextMigrationVersion = %d, want %d", n, tc.want)
			}
		})
	}
}

func TestGenerateMigration(t *testing.T) {
	cases := []struct {
		name            string
		seedFiles       map[string]string
		opts            MigrationOpts
		dryRun          bool
		wantPathSuffix  string // path under project root
		wantContains    []string
		wantErrContains string
		wantWritten     bool // when false, file should NOT exist on disk after the call
	}{
		{
			name:           "first migration in fresh SQLDatastore",
			seedFiles:      apiSqlPgFixture(),
			opts:           MigrationOpts{Description: "add orders"},
			wantPathSuffix: "SQLDatastore/src/main/resources/db/migration/V2__add_orders.sql",
			wantContains:   []string{"-- V2__add_orders.sql", "-- TODO: add orders"},
			wantWritten:    true,
		},
		{
			name: "snake-cases multi-word description",
			seedFiles: mergeFiles(apiSqlPgFixture(), map[string]string{
				"SQLDatastore/src/main/resources/db/migration/V2__add_orders.sql": "-- existing\n",
			}),
			opts:           MigrationOpts{Description: "Add Customer Profiles Table"},
			wantPathSuffix: "SQLDatastore/src/main/resources/db/migration/V3__add_customer_profiles_table.sql",
			wantContains:   []string{"-- V3__add_customer_profiles_table.sql", "-- TODO: Add Customer Profiles Table"},
			wantWritten:    true,
		},
		{
			name: "skips gaps and uses max+1",
			seedFiles: mergeFiles(apiSqlPgFixture(), map[string]string{
				"SQLDatastore/src/main/resources/db/migration/V5__legacy.sql": "-- legacy\n",
			}),
			opts:           MigrationOpts{Description: "next thing"},
			wantPathSuffix: "SQLDatastore/src/main/resources/db/migration/V6__next_thing.sql",
			wantWritten:    true,
		},
		{
			name:           "explicit module=SQLDatastore is fine",
			seedFiles:      apiSqlPgFixture(),
			opts:           MigrationOpts{Description: "ok", Module: "SQLDatastore"},
			wantPathSuffix: "SQLDatastore/src/main/resources/db/migration/V2__ok.sql",
			wantWritten:    true,
		},
		{
			name:            "rejects empty description",
			seedFiles:       apiSqlPgFixture(),
			opts:            MigrationOpts{Description: ""},
			wantErrContains: "--description is required",
		},
		{
			name:            "rejects whitespace-only description",
			seedFiles:       apiSqlPgFixture(),
			opts:            MigrationOpts{Description: "    "},
			wantErrContains: "--description is required",
		},
		{
			name:            "rejects punctuation-only description",
			seedFiles:       apiSqlPgFixture(),
			opts:            MigrationOpts{Description: "!!!"},
			wantErrContains: "empty snake_case",
		},
		{
			name:            "rejects unsupported module",
			seedFiles:       apiSqlPgFixture(),
			opts:            MigrationOpts{Description: "x", Module: "Worker"},
			wantErrContains: "not supported for migrations",
		},
		{
			name:            "rejects project missing SQLDatastore",
			seedFiles:       apiOnlyFixture(),
			opts:            MigrationOpts{Description: "x"},
			wantErrContains: "does not have the SQLDatastore module",
		},
		// Note: refuse-clobber for `add migration` is unreachable through
		// the public API because nextMigrationVersion always picks an
		// unoccupied V{N}. The protection still matters for concurrent
		// invocations and for other add-commands; covered by
		// TestEmitFile_RefusesClobber in emit_test.go.
		{
			name:           "dry-run reports file but does not write",
			seedFiles:      apiSqlPgFixture(),
			opts:           MigrationOpts{Description: "ghost"},
			dryRun:         true,
			wantPathSuffix: "SQLDatastore/src/main/resources/db/migration/V2__ghost.sql",
			wantWritten:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			projectPath := setupProject(t, tc.seedFiles)

			ctx, err := LoadContext(projectPath)
			if err != nil {
				t.Fatalf("LoadContext: %v", err)
			}
			ctx.DryRun = tc.dryRun

			result, err := GenerateMigration(ctx, tc.opts)

			if tc.wantErrContains != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (result: %+v)", tc.wantErrContains, result)
				}
				if !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tc.wantPathSuffix != "" {
				if len(result.Created) != 1 {
					t.Fatalf("Created = %v, want exactly one file (%s)", result.Created, tc.wantPathSuffix)
				}
				if !strings.HasSuffix(result.Created[0], tc.wantPathSuffix) {
					t.Fatalf("Created[0] = %q, want suffix %q", result.Created[0], tc.wantPathSuffix)
				}
			}

			absFile := filepath.Join(projectPath, result.Created[0])
			_, statErr := os.Stat(absFile)
			if tc.wantWritten {
				if statErr != nil {
					t.Fatalf("expected file %s to exist on disk: %v", result.Created[0], statErr)
				}
				body, err := os.ReadFile(absFile)
				if err != nil {
					t.Fatal(err)
				}
				for _, want := range tc.wantContains {
					if !strings.Contains(string(body), want) {
						t.Errorf("file %s missing %q\ngot:\n%s", result.Created[0], want, body)
					}
				}
			} else {
				if !errors.Is(statErr, os.ErrNotExist) {
					t.Fatalf("expected file %s to NOT exist (dry-run), but stat err = %v", result.Created[0], statErr)
				}
			}

			if len(result.NextSteps) == 0 {
				t.Errorf("expected NextSteps to be populated, got empty")
			}
		})
	}
}

// setupProject seeds a fresh tmp dir with the given files and returns
// the absolute path. Each test gets its own isolated project tree.
func setupProject(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for relPath, content := range files {
		abs := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// mergeFiles overlays b onto a (b wins on key collision) and returns
// a new map. Used for parameterizing fixtures across test cases.
func mergeFiles(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	maps.Copy(out, a)
	maps.Copy(out, b)
	return out
}

// apiSqlPgFixture is the canonical "API + SQLDatastore (Postgres)"
// fixture: minimal .trabuco.json plus the V1__baseline migration.
// Mirrors what `trabuco init --modules=API,SQLDatastore --database=postgresql`
// produces.
func apiSqlPgFixture() map[string]string {
	return map[string]string{
		".trabuco.json": `{
  "version": "1.13.2",
  "generatedAt": "2026-01-01T00:00:00Z",
  "projectName": "demo",
  "groupId": "com.example.demo",
  "artifactId": "demo",
  "javaVersion": "21",
  "modules": ["Model", "SQLDatastore", "Shared", "API"],
  "database": "postgresql"
}`,
		"SQLDatastore/src/main/resources/db/migration/V1__baseline.sql": "-- baseline\nCREATE TABLE placeholders (id BIGSERIAL PRIMARY KEY, name VARCHAR(255));\n",
	}
}

// apiOnlyFixture is "API module only" — no SQLDatastore. Used to test
// the "missing SQLDatastore module" error path.
func apiOnlyFixture() map[string]string {
	return map[string]string{
		".trabuco.json": `{
  "version": "1.13.2",
  "generatedAt": "2026-01-01T00:00:00Z",
  "projectName": "demo",
  "groupId": "com.example.demo",
  "artifactId": "demo",
  "javaVersion": "21",
  "modules": ["Model", "Shared", "API"]
}`,
	}
}
