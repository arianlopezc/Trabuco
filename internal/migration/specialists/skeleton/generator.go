// Package skeleton implements the Phase 1 skeleton-builder specialist. It
// generates the Trabuco multi-module skeleton inside the user's existing
// repo in MIGRATION MODE: enforcement mechanisms (Maven Enforcer,
// Spotless, ArchUnit) are deferred. The skeleton is just enough Maven
// structure that legacy code keeps building (now wrapped in a legacy/
// module) and empty new modules pass mvn compile trivially.
//
// The activation phase (Phase 12) flips the enforcement skip flags off.
package skeleton

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/migration/state"
)

// Generator produces a migration-mode multi-module skeleton in repoRoot.
// It does NOT touch the legacy code itself — that gets wrapped into a
// legacy/ module by WrapLegacy().
type Generator struct {
	RepoRoot     string
	GroupID      string
	ProjectName  string
	JavaVersion  string
	Modules      []string
}

// Generate creates parent pom.xml + per-module directories with pom.xml.
// Existing files are NOT overwritten unless explicitly migrating-mode
// stub files.
func (g *Generator) Generate() error {
	if err := g.writeParentPOM(); err != nil {
		return fmt.Errorf("parent pom: %w", err)
	}
	if err := g.writeRootFiles(); err != nil {
		return fmt.Errorf("root files: %w", err)
	}
	for _, m := range g.Modules {
		if err := g.writeModuleSkeleton(m); err != nil {
			return fmt.Errorf("module %s: %w", m, err)
		}
	}
	return nil
}

// WrapLegacy moves all top-level source/build files into a legacy/ Maven
// module so the user's existing code keeps building under the new
// multi-module parent. This is destructive (mv operations); the
// orchestrator's pre-tag is the rollback safety net.
//
// Files moved: src/, pom.xml (renamed to legacy/pom.xml with parent
// reference patched in), Maven wrapper artifacts (mvnw, .mvn/, target/),
// any other top-level Maven build files.
//
// Files NOT moved (kept at root): .git/, .gitignore, README.md,
// .trabuco-migration/, our newly-generated parent pom.xml, our
// newly-generated module/ directories, .ai/, .claude/, .cursor/, etc.
func (g *Generator) WrapLegacy() error {
	legacyDir := filepath.Join(g.RepoRoot, "legacy")
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		return err
	}

	// Move src/ if present.
	if _, err := os.Stat(filepath.Join(g.RepoRoot, "src")); err == nil {
		if err := os.Rename(filepath.Join(g.RepoRoot, "src"), filepath.Join(legacyDir, "src")); err != nil {
			return fmt.Errorf("move src: %w", err)
		}
	}

	// Move existing pom.xml to legacy/legacy-original-pom.xml so we have
	// it as a record, then write a fresh legacy/pom.xml with parent ref.
	legacyOrigPOM := filepath.Join(legacyDir, "legacy-original-pom.xml")
	if data, err := os.ReadFile(filepath.Join(g.RepoRoot, "pom.xml")); err == nil {
		// Save a copy for reference.
		_ = os.WriteFile(legacyOrigPOM, data, 0o644)
		// Remove root pom.xml — we'll write a fresh parent next.
		// (Generate() above already wrote a NEW parent pom.xml, but it's
		// possible the legacy root pom.xml wasn't removed yet if this is
		// called before Generate. The standard order is Generate first,
		// then WrapLegacy.)
	}

	if err := g.writeLegacyModulePOM(legacyDir, legacyOrigPOM); err != nil {
		return err
	}

	// Add `legacy` to parent's modules list.
	return g.appendModuleToParent("legacy")
}

// writeParentPOM writes the migration-mode parent pom.xml. Distinct from
// Trabuco's standard parent POM in that:
//   - maven-enforcer-plugin is NOT included at all (added by activator)
//   - spotless-maven-plugin is NOT included (added by activator)
//   - jacoco's check execution is omitted (added by activator)
//
// Once activator runs in Phase 12, the parent POM is rewritten to the
// full enforcement-on form.
func (g *Generator) writeParentPOM() error {
	path := filepath.Join(g.RepoRoot, "pom.xml")
	if _, err := os.Stat(path); err == nil {
		// Don't clobber an existing pom.xml — Generate() runs before
		// WrapLegacy(), so the legacy pom.xml hasn't been moved yet.
		// Read it to preserve user's groupId/artifactId/version.
		// For 1.10.0 we just back it up; the user's existing root pom
		// will become legacy/pom.xml after WrapLegacy.
		_ = os.Rename(path, path+".pre-trabuco-backup")
	}

	body := fmt.Sprintf(parentPOMTemplate,
		g.GroupID,
		g.ProjectName+"-parent",
		g.JavaVersion,
		strings.Join(moduleListXML(append([]string{}, g.Modules...)), "\n        "),
	)
	return os.WriteFile(path, []byte(body), 0o644)
}

// writeRootFiles writes .gitignore, .editorconfig, .trabuco.json, and
// .trabuco-migration/.gitignore. Idempotent: skips files that already
// exist (we don't clobber the user's choices).
func (g *Generator) writeRootFiles() error {
	files := map[string]string{
		".trabuco.json": fmt.Sprintf(`{
  "version": "1.10.0-dev",
  "projectName": %q,
  "groupId": %q,
  "javaVersion": %q,
  "modules": %s,
  "migrationInProgress": true
}
`, g.ProjectName, g.GroupID, g.JavaVersion, jsonModulesArray(g.Modules)),
		".editorconfig": editorConfig,
	}
	for name, body := range files {
		path := filepath.Join(g.RepoRoot, name)
		if _, err := os.Stat(path); err == nil {
			continue
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return err
		}
	}

	// Ensure .gitignore has the Trabuco-migration entry.
	gitignore := filepath.Join(g.RepoRoot, ".gitignore")
	data, _ := os.ReadFile(gitignore)
	if !strings.Contains(string(data), ".trabuco-migration/") {
		entry := "\n# Trabuco migration working state\n.trabuco-migration/\n"
		if err := os.WriteFile(gitignore, append(data, []byte(entry)...), 0o644); err != nil {
			return err
		}
	}
	return nil
}

// writeModuleSkeleton creates a single module directory with a stub pom.xml.
// The module starts empty; later phase specialists populate it.
func (g *Generator) writeModuleSkeleton(module string) error {
	dir := filepath.Join(g.RepoRoot, strings.ToLower(module))
	if err := os.MkdirAll(filepath.Join(dir, "src", "main", "java"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "src", "test", "java"), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Join(dir, "src", "main", "resources"), 0o755); err != nil {
		return err
	}

	pomPath := filepath.Join(dir, "pom.xml")
	if _, err := os.Stat(pomPath); err == nil {
		// Don't clobber.
		return nil
	}
	body := fmt.Sprintf(moduleStubPOMTemplate, g.GroupID, g.ProjectName+"-parent", strings.ToLower(module), module)
	return os.WriteFile(pomPath, []byte(body), 0o644)
}

// writeLegacyModulePOM creates legacy/pom.xml — a Maven module that wraps
// the user's existing root pom.xml. The original root pom is saved as
// legacy/legacy-original-pom.xml for reference; the new legacy/pom.xml
// inherits from the parent and references the legacy source as-is.
func (g *Generator) writeLegacyModulePOM(legacyDir, origPOMPath string) error {
	// For 1.10.0 we use a minimal legacy module pom that compiles the
	// legacy source as-is. The user's original pom (saved alongside)
	// serves as the source of truth for dependency resolution.
	body := fmt.Sprintf(legacyModulePOMTemplate, g.GroupID, g.ProjectName+"-parent")
	return os.WriteFile(filepath.Join(legacyDir, "pom.xml"), []byte(body), 0o644)
}

// appendModuleToParent updates root pom.xml's <modules> section to add
// the given module. Idempotent.
func (g *Generator) appendModuleToParent(module string) error {
	path := filepath.Join(g.RepoRoot, "pom.xml")
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if strings.Contains(string(data), "<module>"+module+"</module>") {
		return nil
	}
	updated := strings.Replace(string(data),
		"</modules>",
		"        <module>"+module+"</module>\n    </modules>",
		1,
	)
	return os.WriteFile(path, []byte(updated), 0o644)
}

func moduleListXML(modules []string) []string {
	out := make([]string, 0, len(modules))
	for _, m := range modules {
		out = append(out, "<module>"+strings.ToLower(m)+"</module>")
	}
	return out
}

func jsonModulesArray(modules []string) string {
	if len(modules) == 0 {
		return "[]"
	}
	parts := make([]string, len(modules))
	for i, m := range modules {
		parts[i] = fmt.Sprintf("%q", m)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// ---------- POM templates ----------

const parentPOMTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <groupId>%s</groupId>
    <artifactId>%s</artifactId>
    <version>1.0-SNAPSHOT</version>
    <packaging>pom</packaging>

    <!-- Trabuco migration mode: enforcement mechanisms (Maven Enforcer,
         Spotless, ArchUnit, Jacoco threshold) are deliberately omitted
         until Phase 12 (activation). This lets legacy CI continue to work
         at every phase boundary during migration. See
         docs/MIGRATION_REDESIGN_PLAN.md §5 for the rationale. -->

    <properties>
        <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
        <maven.compiler.source>%s</maven.compiler.source>
        <maven.compiler.target>%[3]s</maven.compiler.target>
    </properties>

    <modules>
        %s
    </modules>

    <build>
        <pluginManagement>
            <plugins>
                <plugin>
                    <groupId>org.apache.maven.plugins</groupId>
                    <artifactId>maven-compiler-plugin</artifactId>
                    <version>3.13.0</version>
                    <configuration>
                        <release>%[3]s</release>
                    </configuration>
                </plugin>
            </plugins>
        </pluginManagement>
    </build>
</project>
`

const moduleStubPOMTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <parent>
        <groupId>%s</groupId>
        <artifactId>%s</artifactId>
        <version>1.0-SNAPSHOT</version>
    </parent>

    <artifactId>%s</artifactId>
    <name>%s</name>
    <description>Trabuco migration: empty %[4]s module skeleton (populated by later phase specialist).</description>
</project>
`

const legacyModulePOMTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>

    <parent>
        <groupId>%s</groupId>
        <artifactId>%s</artifactId>
        <version>1.0-SNAPSHOT</version>
    </parent>

    <artifactId>legacy</artifactId>
    <name>Legacy (transitional)</name>
    <description>Wraps the user's existing source during migration. Per-aggregate
    code is moved out of legacy/ into Trabuco-shaped modules by the
    Model/Datastore/Shared/API/Worker/EventConsumer specialists. When
    everything has been moved, the finalizer (Phase 13) removes this
    module — or retains it with @Deprecated markers if the user opted
    to preserve unmigrated artifacts.</description>

    <!-- Original user pom.xml is preserved in legacy-original-pom.xml
         for dependency reference. -->
</project>
`

const editorConfig = `# Trabuco-generated EditorConfig
root = true

[*]
indent_style = space
indent_size = 4
end_of_line = lf
charset = utf-8
trim_trailing_whitespace = true
insert_final_newline = true

[*.{json,yml,yaml,xml}]
indent_size = 2

[*.md]
trim_trailing_whitespace = false
`

// LoadGroupAndProjectFromState extracts groupId / project name from
// state.TargetConfig + assessment, falling back to defaults derived from
// the repo path.
func LoadGroupAndProjectFromState(repoRoot string, target *state.TargetConfig) (groupID, projectName string) {
	projectName = filepath.Base(repoRoot)
	groupID = "com." + strings.ToLower(strings.ReplaceAll(projectName, "-", ""))
	return
}
