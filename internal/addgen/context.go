// Package addgen implements the deterministic file generators behind
// `trabuco add <type>` (migration, entity, service, job, endpoint,
// streaming-endpoint, event, test).
//
// These commands only ever CREATE new files. They never edit or delete.
// Edits and deletes stay with the coding agent — the CLI's contract is
// "ask me for additions, I produce byte-deterministic output every time."
//
// Each generator is a pure function over a Context (loaded from
// .trabuco.json) plus typed Opts; it returns a Result describing what
// was (or would be, in dry-run) written, plus next-step hints for the
// agent. CLI wrappers in internal/cli/add_*.go are thin — they parse
// flags, call into here, and print the Result. MCP tools call the same
// functions directly so both surfaces stay in lockstep.
package addgen

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// Context bundles the project state every add-command needs: the
// resolved ProjectConfig (so all the helpers like HasModule,
// PackagePath, ProjectNamePascal are available) plus the absolute
// project root path and the dry-run flag.
type Context struct {
	*config.ProjectConfig
	ProjectPath string
	DryRun      bool
}

// LoadContext reads .trabuco.json from projectPath and returns a
// Context ready for any add-command. Returns a clear error when the
// file is missing — every add-command requires a Trabuco-managed
// project to operate on.
func LoadContext(projectPath string) (*Context, error) {
	abs, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project path: %w", err)
	}
	meta, err := config.LoadMetadata(abs)
	if err != nil {
		return nil, fmt.Errorf("not a Trabuco project (no %s in %s): %w", config.MetadataFileName, abs, err)
	}
	return &Context{
		ProjectConfig: meta.ToProjectConfig(),
		ProjectPath:   abs,
	}, nil
}

// modulePackageSegment returns the lowercase package fragment for a
// given module — e.g. "SQLDatastore" → "sqldatastore". Mirrors the
// internal generator's javaPath() convention so add-commands emit
// files into the same locations as the original init scaffolding.
func modulePackageSegment(module string) string {
	return strings.ToLower(module)
}

// JavaSrcMain returns the relative path to the main Java source root
// for a module + subpackage:
//
//	SQLDatastore/src/main/java/com/example/demo/sqldatastore/repository
//
// Module directory is PascalCase, package segment is lowercase. The
// subpackage may be empty (returns just the module Java root).
func (c *Context) JavaSrcMain(module, subpackage string) string {
	parts := []string{module, "src", "main", "java", c.PackagePath(), modulePackageSegment(module)}
	if subpackage != "" {
		parts = append(parts, subpackage)
	}
	return filepath.Join(parts...)
}

// JavaSrcTest returns the relative path to the test Java source root
// for a module + subpackage. Same shape as JavaSrcMain but under
// src/test/java.
func (c *Context) JavaSrcTest(module, subpackage string) string {
	parts := []string{module, "src", "test", "java", c.PackagePath(), modulePackageSegment(module)}
	if subpackage != "" {
		parts = append(parts, subpackage)
	}
	return filepath.Join(parts...)
}

// ResourcesMain returns the relative path under src/main/resources
// for a module. Used for application.yml, db/migration/, etc.
func (c *Context) ResourcesMain(module, subdir string) string {
	parts := []string{module, "src", "main", "resources"}
	if subdir != "" {
		parts = append(parts, subdir)
	}
	return filepath.Join(parts...)
}

// JavaPackage returns the dotted Java package name for a module +
// subpackage — e.g. "com.example.demo.sqldatastore.repository".
// Module segment is lowercased to match javaPath() output.
func (c *Context) JavaPackage(module, subpackage string) string {
	pkg := c.GroupID + "." + modulePackageSegment(module)
	if subpackage != "" {
		pkg += "." + subpackage
	}
	return pkg
}
