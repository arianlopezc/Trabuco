package addgen

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// MigrationOpts is the input contract for `trabuco add migration`.
type MigrationOpts struct {
	// Description becomes the migration filename suffix after
	// snake-casing (e.g. "Add Orders Table" → "add_orders_table").
	// Required.
	Description string

	// Module gates which datastore the migration lands in. Only
	// SQLDatastore is supported today (Flyway is the SQL story);
	// callers who pass empty get SQLDatastore by default.
	Module string
}

// GenerateMigration emits a new empty Flyway migration file under
// SQLDatastore/src/main/resources/db/migration/V{N}__{snake_desc}.sql
// where {N} is the next available version. The file body is a
// minimal comment header — the agent fills in the DDL in a follow-up
// edit. Refuse-clobber if a file at that path already exists.
func GenerateMigration(ctx *Context, opts MigrationOpts) (*Result, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	if strings.TrimSpace(opts.Description) == "" {
		return nil, fmt.Errorf("--description is required")
	}
	if opts.Module == "" {
		opts.Module = config.ModuleSQLDatastore
	}
	if opts.Module != config.ModuleSQLDatastore {
		return nil, fmt.Errorf("--module=%s not supported for migrations (only SQLDatastore has Flyway)", opts.Module)
	}
	if !ctx.HasModule(config.ModuleSQLDatastore) {
		return nil, fmt.Errorf("project does not have the SQLDatastore module — run `trabuco add SQLDatastore` first")
	}

	desc := snakeCase(opts.Description)
	if desc == "" {
		return nil, fmt.Errorf("--description %q produced an empty snake_case form; pass alphanumeric text", opts.Description)
	}

	migrationDir := ctx.ResourcesMain(config.ModuleSQLDatastore, filepath.Join("db", "migration"))
	absMigrationDir := filepath.Join(ctx.ProjectPath, migrationDir)

	version, err := nextMigrationVersion(absMigrationDir)
	if err != nil {
		return nil, fmt.Errorf("failed to scan %s: %w", migrationDir, err)
	}

	relPath := filepath.Join(migrationDir, fmt.Sprintf("V%d__%s.sql", version, desc))
	content := buildMigrationContent(version, desc, opts.Description)

	result := &Result{}
	if err := ctx.emitFile(relPath, content, result); err != nil {
		return nil, err
	}
	result.NextSteps = []string{
		fmt.Sprintf("Edit %s and add the DDL.", relPath),
		"Run `mvn -pl SQLDatastore flyway:info` (or boot the API module) to verify Flyway picks the migration up.",
		"Add a corresponding Flyway repair entry only if you need to roll forward in a deployed environment.",
	}
	return result, nil
}

// buildMigrationContent renders the empty-migration body. Kept out of
// templates/ because there's no template-engine value here — it's a
// fixed two-line header that the agent immediately replaces.
func buildMigrationContent(version int, snakeDesc string, originalDesc string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "-- V%d__%s.sql\n", version, snakeDesc)
	fmt.Fprintf(&b, "-- TODO: %s\n", strings.TrimSpace(originalDesc))
	b.WriteString("-- Replace this comment with the migration DDL.\n")
	return b.String()
}

var migrationFilenameRE = regexp.MustCompile(`^V(\d+)__.*\.sql$`)

// nextMigrationVersion returns max(V{N}) + 1 across .sql files in the
// directory. Returns 1 when the directory is missing or empty so an
// agent can call this command on a project where SQLDatastore was
// just added (and V1 hasn't been generated yet, in some edge cases).
func nextMigrationVersion(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 1, nil
		}
		return 0, err
	}
	max := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		m := migrationFilenameRE.FindStringSubmatch(e.Name())
		if m == nil {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if n > max {
			max = n
		}
	}
	return max + 1, nil
}

// snakeCase converts arbitrary user input (with spaces, hyphens, mixed
// case) into a Flyway-friendly snake_case identifier. Drops anything
// that isn't a letter, digit, space, hyphen, or underscore. Collapses
// whitespace runs to a single underscore.
func snakeCase(s string) string {
	var b strings.Builder
	prevSep := true // suppress leading separator
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + 32)
			prevSep = false
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			prevSep = false
		case r == ' ' || r == '-' || r == '_' || r == '\t':
			if !prevSep {
				b.WriteByte('_')
				prevSep = true
			}
		default:
			// drop other characters
		}
	}
	return strings.TrimRight(b.String(), "_")
}
