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

// WrapLegacy moves the user's existing Maven source layout into a
// legacy/ module so the multi-module parent compiles. Standard call
// order is Generate() first (which wrote a fresh parent pom.xml,
// renaming the user's original to pom.xml.pre-trabuco-backup) then
// WrapLegacy() (which moves the backup + src/ into legacy/).
//
// Files moved into legacy/:
//   - src/ → legacy/src/
//   - pom.xml.pre-trabuco-backup → legacy/legacy-original-pom.xml
//
// A new legacy/pom.xml is written that inherits from the parent and
// preserves the user's plugin/dep declarations from the original POM
// (best-effort: the legacy original is also preserved verbatim for
// reference).
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

	// Move the user's original pom.xml (backed up by Generate as
	// pom.xml.pre-trabuco-backup) into the legacy module. If for some
	// reason the backup doesn't exist (e.g., Generate was skipped or the
	// repo started without a pom), proceed with a stub legacy pom.
	backup := filepath.Join(g.RepoRoot, "pom.xml.pre-trabuco-backup")
	legacyOrigPOM := filepath.Join(legacyDir, "legacy-original-pom.xml")
	if _, err := os.Stat(backup); err == nil {
		if err := os.Rename(backup, legacyOrigPOM); err != nil {
			return fmt.Errorf("move legacy original pom: %w", err)
		}
	}

	if err := g.writeLegacyModulePOM(legacyDir); err != nil {
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
  "version": "1.12.1",
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

// writeLegacyModulePOM creates legacy/pom.xml. If the user's original POM
// was preserved at legacy/legacy-original-pom.xml, we extract its
// <dependencies> and <build><plugins> and embed them in the new legacy
// module POM so the legacy source still has its build/runtime
// requirements satisfied. We also import the spring-boot-dependencies
// BOM to restore the version-management the original spring-boot-starter-
// parent provided (the legacy module's parent is now the Trabuco parent,
// which has no BOM). Otherwise we write a minimal stub.
func (g *Generator) writeLegacyModulePOM(legacyDir string) error {
	depsBlock, pluginsBlock, depMgmtBlock := "", "", ""
	origPOMPath := filepath.Join(legacyDir, "legacy-original-pom.xml")
	if data, err := os.ReadFile(origPOMPath); err == nil {
		orig := string(data)
		depsBlock = extractXMLBlock(orig, "dependencies")
		pluginsBlock = extractXMLBlock(orig, "plugins")
		if v := extractSpringBootParentVersion(orig); v != "" {
			depMgmtBlock = fmt.Sprintf(springBootBOMImport, v)
		}
	}
	body := fmt.Sprintf(legacyModulePOMTemplate,
		g.GroupID, g.ProjectName+"-parent",
		depMgmtBlock, depsBlock, pluginsBlock,
	)
	return os.WriteFile(filepath.Join(legacyDir, "pom.xml"), []byte(body), 0o644)
}

// extractSpringBootParentVersion finds <parent>...spring-boot-starter-parent
// ...<version>X</version> and returns X. Empty if not present.
func extractSpringBootParentVersion(xml string) string {
	parent := extractXMLBlock(xml, "parent")
	if parent == "" || !strings.Contains(parent, "spring-boot-starter-parent") {
		return ""
	}
	v := extractXMLBlock(parent, "version")
	v = strings.TrimPrefix(v, "<version>")
	v = strings.TrimSuffix(v, "</version>")
	return strings.TrimSpace(v)
}

// extractXMLBlock returns the first <tag>...</tag> block from xml, or
// empty if not found. Whitespace-tolerant; doesn't validate XML.
func extractXMLBlock(xml, tag string) string {
	open := "<" + tag + ">"
	close := "</" + tag + ">"
	i := strings.Index(xml, open)
	if i == -1 {
		return ""
	}
	j := strings.Index(xml[i:], close)
	if j == -1 {
		return ""
	}
	return xml[i : i+j+len(close)]
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
         at every phase boundary during migration. -->

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

    <!-- Original user pom.xml is preserved at legacy/legacy-original-pom.xml
         for full reference; the dependencies and build plugins below were
         extracted from it best-effort so the legacy source still compiles.
         The dependencyManagement BOM import (when present) restores the
         version-management spring-boot-starter-parent provided in the
         original POM. -->

    %s

    %s

    <build>
        %s
    </build>
</project>
`

// springBootBOMImport is the dependencyManagement block that imports
// spring-boot-dependencies BOM at the version we extracted from the
// user's original spring-boot-starter-parent.
const springBootBOMImport = `<dependencyManagement>
        <dependencies>
            <dependency>
                <groupId>org.springframework.boot</groupId>
                <artifactId>spring-boot-dependencies</artifactId>
                <version>%s</version>
                <type>pom</type>
                <scope>import</scope>
            </dependency>
        </dependencies>
    </dependencyManagement>`

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
