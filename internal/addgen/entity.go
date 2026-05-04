package addgen

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/utils"
)

// EntityOpts is the input contract for `trabuco add entity`.
type EntityOpts struct {
	// Name is the PascalCase entity class name (e.g. "Order").
	Name string

	// Fields is the raw --fields="..." string. Parsed into typed
	// Fields by ParseFields.
	Fields string

	// Module forces a specific datastore module when both
	// SQLDatastore and NoSQLDatastore are present. Empty = auto-detect
	// (errors when both are present without an explicit choice).
	Module string

	// TableName overrides the auto-generated plural snake_case table
	// name. Useful for irregular plurals (Person → people) and for
	// matching legacy schemas (LegacyOrder → orders_v1).
	TableName string
}

// GenerateEntity dispatches to the SQL or Mongo flavor based on the
// project's datastore modules. Validates the entity name and field
// spec up front, then delegates to the dedicated generator.
//
// Files emitted (SQL flavor):
//   - Model/.../entities/{Name}.java         (Immutables interface)
//   - Model/.../entities/{Name}Record.java   (Spring Data JDBC record)
//   - SQLDatastore/.../repository/{Name}Repository.java
//   - SQLDatastore/.../db/migration/V{N}__create_{table}.sql
//   - Model/.../entities/{Enum}.java         (one per distinct enum field; skipped if exists)
//
// Files emitted (Mongo flavor):
//   - Model/.../entities/{Name}.java         (Immutables interface)
//   - Model/.../entities/{Name}Document.java (Spring Data MongoDB document)
//   - NoSQLDatastore/.../repository/{Name}DocumentRepository.java
//   - Model/.../entities/{Enum}.java         (one per distinct enum field)
func GenerateEntity(ctx *Context, opts EntityOpts) (*Result, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	name := strings.TrimSpace(opts.Name)
	if name == "" {
		return nil, fmt.Errorf("entity name is required (positional argument)")
	}
	if !isValidJavaIdentifier(name) || !isUpperFirst(name) {
		return nil, fmt.Errorf("entity name %q must be a PascalCase Java identifier", name)
	}
	if reservedJavaWords[strings.ToLower(name)] {
		return nil, fmt.Errorf("entity name %q is a reserved Java keyword", name)
	}

	fields, err := ParseFields(opts.Fields)
	if err != nil {
		return nil, err
	}
	if !ctx.HasModule(config.ModuleModel) {
		return nil, fmt.Errorf("project does not have the Model module — entities live there")
	}

	module, err := pickEntityModule(ctx, opts.Module)
	if err != nil {
		return nil, err
	}

	switch module {
	case config.ModuleSQLDatastore:
		return generateSQLEntity(ctx, opts, fields)
	case config.ModuleNoSQLDatastore:
		return generateMongoEntity(ctx, opts, fields)
	}
	return nil, fmt.Errorf("unsupported entity module %s", module)
}

// pickEntityModule resolves which datastore to target. Honors an
// explicit --module if provided; otherwise auto-detects when exactly
// one datastore module is present in the project.
func pickEntityModule(ctx *Context, explicit string) (string, error) {
	hasSQL := ctx.HasModule(config.ModuleSQLDatastore)
	hasMongo := ctx.HasModule(config.ModuleNoSQLDatastore)
	if explicit != "" {
		switch explicit {
		case config.ModuleSQLDatastore:
			if !hasSQL {
				return "", fmt.Errorf("--module=SQLDatastore but the project doesn't have it")
			}
			return explicit, nil
		case config.ModuleNoSQLDatastore:
			if !hasMongo {
				return "", fmt.Errorf("--module=NoSQLDatastore but the project doesn't have it")
			}
			return explicit, nil
		default:
			return "", fmt.Errorf("--module must be SQLDatastore or NoSQLDatastore (got %s)", explicit)
		}
	}
	switch {
	case hasSQL && hasMongo:
		return "", fmt.Errorf("project has both SQLDatastore and NoSQLDatastore — pass --module=SQLDatastore or --module=NoSQLDatastore explicitly")
	case hasSQL:
		return config.ModuleSQLDatastore, nil
	case hasMongo:
		return config.ModuleNoSQLDatastore, nil
	}
	return "", fmt.Errorf("project has no datastore module — add SQLDatastore or NoSQLDatastore first")
}

// resolveTableName returns the user-supplied table name or the
// PluralLowerSnake form of the entity name.
func resolveTableName(name, override string) string {
	if override != "" {
		return override
	}
	return utils.PluralLowerSnake(name)
}

// emitEnumIfMissing creates an enum class file under
// Model/.../entities/{Enum}.java unless the file already exists.
// Existing enums are noted but never overwritten — multiple
// add-entity calls referencing the same enum are idempotent.
func (c *Context) emitEnumIfMissing(enumName string, result *Result) error {
	rel := filepath.Join(c.JavaSrcMain(config.ModuleModel, "entities"), enumName+".java")
	if _, err := os.Stat(filepath.Join(c.ProjectPath, rel)); err == nil {
		result.Notes = append(result.Notes, fmt.Sprintf("Enum %s already exists at %s — left untouched.", enumName, rel))
		return nil
	}
	return c.emitFile(rel, renderEnumStub(c, enumName), result)
}

// renderEnumStub returns a placeholder enum class. The agent edits
// the values after generation; the CLI deliberately picks PLACEHOLDER
// constants to make the unfinished state obvious in code review.
func renderEnumStub(ctx *Context, enumName string) string {
	pkg := ctx.JavaPackage(config.ModuleModel, "entities")
	var b strings.Builder
	fmt.Fprintf(&b, "package %s;\n\n", pkg)
	b.WriteString("/**\n")
	fmt.Fprintf(&b, " * Enum stub generated by `trabuco add entity`. Replace the placeholder\n")
	b.WriteString(" * values with your actual states.\n")
	b.WriteString(" */\n")
	fmt.Fprintf(&b, "public enum %s {\n", enumName)
	b.WriteString("  PLACEHOLDER_VALUE\n")
	b.WriteString("}\n")
	return b.String()
}
