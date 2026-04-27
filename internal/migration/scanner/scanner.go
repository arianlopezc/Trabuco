// Package scanner does a structured pre-scan of the user's source
// repository so the assessor specialist's LLM has concrete data to
// reason over. The LLM cannot file-walk on its own (no tool-calling in
// our current architecture), so the Go side prepares everything the
// LLM needs and embeds it in the user prompt.
//
// The scan is intentionally conservative: it captures structural
// signals (file paths, build files, key annotations on Java classes)
// without trying to fully parse Java. The assessor LLM does the
// classification.
package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Scan walks repoRoot and returns a structured snapshot suitable for
// the assessor's user prompt. It does NOT modify any files.
func Scan(repoRoot string) (*Snapshot, error) {
	snap := &Snapshot{
		RepoRoot: repoRoot,
	}

	// Top-level build files.
	if data, err := os.ReadFile(filepath.Join(repoRoot, "pom.xml")); err == nil {
		snap.BuildSystem = "maven"
		snap.RootPOM = string(data)
	} else if data, err := os.ReadFile(filepath.Join(repoRoot, "build.gradle")); err == nil {
		snap.BuildSystem = "gradle"
		snap.RootBuild = string(data)
	} else if data, err := os.ReadFile(filepath.Join(repoRoot, "build.gradle.kts")); err == nil {
		snap.BuildSystem = "gradle-kotlin"
		snap.RootBuild = string(data)
	} else {
		snap.BuildSystem = "unknown"
	}

	// Walk for .java sources.
	_ = filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable
		}
		// Skip known build/output and tooling dirs. Do NOT blanket-skip
		// dot-prefixed dirs — `.github`, `.gitlab`, `.circleci`,
		// `.devcontainer`, etc. carry critical CI/deployment signal.
		if d.IsDir() {
			name := d.Name()
			if isSkipDir(name) && path != repoRoot {
				return fs.SkipDir
			}
			return nil
		}
		rel, _ := filepath.Rel(repoRoot, path)
		ext := strings.ToLower(filepath.Ext(path))

		switch ext {
		case ".java":
			snap.JavaFiles = append(snap.JavaFiles, scanJavaFile(repoRoot, rel))
		case ".kt", ".kts":
			snap.KotlinFiles = append(snap.KotlinFiles, rel)
		case ".scala":
			snap.ScalaFiles = append(snap.ScalaFiles, rel)
		case ".groovy":
			snap.GroovyFiles = append(snap.GroovyFiles, rel)
		case ".py", ".js", ".ts", ".go", ".rs":
			snap.NonJVMFiles = append(snap.NonJVMFiles, rel)
		}

		// Notable file types regardless of extension.
		base := strings.ToLower(d.Name())
		switch {
		case base == "dockerfile" || strings.HasPrefix(base, "dockerfile."):
			snap.Dockerfiles = append(snap.Dockerfiles, rel)
		case base == "jenkinsfile":
			snap.CIFiles = append(snap.CIFiles, ciHit{Provider: "jenkins", Path: rel})
		case base == ".gitlab-ci.yml" || base == "gitlab-ci.yml":
			snap.CIFiles = append(snap.CIFiles, ciHit{Provider: "gitlab", Path: rel})
		case base == ".travis.yml":
			snap.CIFiles = append(snap.CIFiles, ciHit{Provider: "travis", Path: rel})
		case base == "azure-pipelines.yml":
			snap.CIFiles = append(snap.CIFiles, ciHit{Provider: "azure", Path: rel})
		case base == "fly.toml":
			snap.DeploymentFiles = append(snap.DeploymentFiles, deployHit{Kind: "fly", Path: rel})
		case base == "render.yaml":
			snap.DeploymentFiles = append(snap.DeploymentFiles, deployHit{Kind: "render", Path: rel})
		}

		// Path-based detection.
		switch {
		case strings.HasPrefix(rel, ".github/workflows/") && (ext == ".yml" || ext == ".yaml"):
			snap.CIFiles = append(snap.CIFiles, ciHit{Provider: "github-actions", Path: rel})
		case strings.HasPrefix(rel, ".circleci/"):
			snap.CIFiles = append(snap.CIFiles, ciHit{Provider: "circleci", Path: rel})
		case strings.Contains(rel, "/argocd/") || strings.HasPrefix(rel, "argocd/"):
			snap.DeploymentFiles = append(snap.DeploymentFiles, deployHit{Kind: "argo", Path: rel})
		case strings.HasPrefix(rel, "helm/") || strings.HasPrefix(rel, "charts/"):
			snap.DeploymentFiles = append(snap.DeploymentFiles, deployHit{Kind: "helm", Path: rel})
		case strings.HasPrefix(rel, "k8s/") || strings.HasPrefix(rel, "kubernetes/") || strings.HasPrefix(rel, "manifests/"):
			snap.DeploymentFiles = append(snap.DeploymentFiles, deployHit{Kind: "k8s", Path: rel})
		case strings.HasPrefix(rel, "terraform/"):
			snap.DeploymentFiles = append(snap.DeploymentFiles, deployHit{Kind: "terraform", Path: rel})
		}

		// Configuration files. Capture path + content so the assessor LLM
		// can inspect them for hardcoded secrets without a separate tool
		// call. Cap size at 8KB per file to bound prompt growth.
		if base == "application.properties" || base == "application.yml" || base == "application.yaml" ||
			strings.HasPrefix(base, "application-") {
			snap.ConfigFiles = append(snap.ConfigFiles, rel)
			if data, err := os.ReadFile(path); err == nil {
				content := string(data)
				if len(content) > 8192 {
					content = content[:8192] + "\n... (truncated)"
				}
				snap.ConfigFileContents = append(snap.ConfigFileContents, configFile{Path: rel, Content: content})
			}
		}

		// Flyway / Liquibase migrations.
		if strings.Contains(rel, "/db/migration/") || strings.Contains(rel, "/db/changelog/") {
			snap.MigrationFiles = append(snap.MigrationFiles, rel)
		}

		return nil
	})

	return snap, nil
}

// isSkipDir reports whether a directory should be skipped during the
// walk. Build outputs, version-control internals, and IDE caches go in;
// CI/deployment-relevant hidden dirs (.github, .gitlab, .circleci, etc.)
// stay walkable.
func isSkipDir(name string) bool {
	switch name {
	case ".git", ".idea", ".vscode", ".gradle", ".mvn", ".m2",
		"node_modules", "target", "build", "dist", "out",
		".trabuco-migration":
		return true
	}
	return false
}

// scanJavaFile reads a Java file and extracts its annotation/import
// signature without full parsing. The LLM does the semantic
// classification.
func scanJavaFile(repoRoot, rel string) JavaFile {
	jf := JavaFile{Path: rel}
	full := filepath.Join(repoRoot, rel)
	data, err := os.ReadFile(full)
	if err != nil {
		return jf
	}
	src := string(data)

	if m := pkgRegexp.FindStringSubmatch(src); len(m) == 2 {
		jf.Package = m[1]
	}
	if m := classRegexp.FindStringSubmatch(src); len(m) == 2 {
		jf.ClassName = m[1]
	}

	// Detect notable annotations cheaply.
	annotations := []string{
		"@Entity", "@Document", "@Table", "@Repository",
		"@RestController", "@Controller", "@Service", "@Component",
		"@Configuration", "@Scheduled", "@Async",
		"@KafkaListener", "@RabbitListener", "@SqsListener",
		"@SpringBootApplication", "@SpringBootTest", "@WebMvcTest",
		"@DataJdbcTest", "@DataJpaTest", "@Test", "@RunWith",
		"@Autowired", "@Inject",
		"@ManyToOne", "@OneToMany", "@OneToOne", "@ManyToMany",
		"@JoinColumn", "@JoinTable", "@ForeignKey",
	}
	for _, ann := range annotations {
		if strings.Contains(src, ann) {
			jf.Annotations = append(jf.Annotations, ann)
		}
	}

	// Detect a few other signal patterns.
	if strings.Contains(src, "ApplicationContext") && strings.Contains(src, "getBean") {
		jf.Signals = append(jf.Signals, "appcontext-getbean")
	}
	if strings.Contains(src, "static final") && strings.Contains(src, "= new HashMap") {
		jf.Signals = append(jf.Signals, "static-mutable-state-suspect")
	}
	if strings.Contains(src, "ServiceLoader") {
		jf.Signals = append(jf.Signals, "serviceloader")
	}
	if strings.Contains(src, "PowerMock") {
		jf.Signals = append(jf.Signals, "powermock")
	}
	if strings.Contains(src, "import org.junit.Test;") {
		jf.Signals = append(jf.Signals, "junit-4")
	}
	if strings.Contains(src, "import org.junit.jupiter.api.Test;") {
		jf.Signals = append(jf.Signals, "junit-5")
	}
	if strings.Contains(src, "import org.testcontainers") {
		jf.Signals = append(jf.Signals, "testcontainers")
	}
	if strings.Contains(src, "@Autowired") &&
		!strings.Contains(src, "@Autowired\n  public") &&
		!strings.Contains(src, "@Autowired\npublic") {
		jf.Signals = append(jf.Signals, "field-injection-suspect")
	}

	// Detect hardcoded credential-looking strings (very rough).
	if credPattern.MatchString(src) {
		jf.Signals = append(jf.Signals, "hardcoded-credential-suspect")
	}

	// Pagination signals — Spring Data Pageable / PageRequest indicate
	// OFFSET pagination, which Trabuco rejects in favor of keyset.
	if strings.Contains(src, "Pageable") || strings.Contains(src, "PageRequest") {
		jf.Signals = append(jf.Signals, "uses-pageable-offset")
	}

	// FK relationship signals (separate from annotation list because the
	// JSON consumer treats annotations as a flat list; aggregating into a
	// single signal makes the assessor's logic clearer).
	if strings.Contains(src, "@ManyToOne") || strings.Contains(src, "@OneToMany") ||
		strings.Contains(src, "@OneToOne") || strings.Contains(src, "@ManyToMany") {
		jf.Signals = append(jf.Signals, "has-jpa-relationship")
	}

	return jf
}

var (
	pkgRegexp   = regexp.MustCompile(`(?m)^\s*package\s+([\w.]+)\s*;`)
	classRegexp = regexp.MustCompile(`(?m)^\s*(?:public\s+|abstract\s+|final\s+)*(?:class|interface|record|enum)\s+(\w+)`)
	credPattern = regexp.MustCompile(`(?i)(password|passwd|secret|api[_-]?key|access[_-]?token)\s*=\s*"[^${}]+"`)
)

// Snapshot is the structured pre-scan result.
type Snapshot struct {
	RepoRoot     string
	BuildSystem  string
	RootPOM      string
	RootBuild    string

	JavaFiles    []JavaFile
	KotlinFiles  []string
	ScalaFiles   []string
	GroovyFiles  []string
	NonJVMFiles  []string

	ConfigFiles        []string
	ConfigFileContents []configFile
	MigrationFiles     []string
	Dockerfiles        []string
	CIFiles            []ciHit
	DeploymentFiles    []deployHit
}

type configFile struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// JavaFile captures one .java file's coarse signature.
type JavaFile struct {
	Path        string   `json:"path"`
	Package     string   `json:"package,omitempty"`
	ClassName   string   `json:"className,omitempty"`
	Annotations []string `json:"annotations,omitempty"`
	Signals     []string `json:"signals,omitempty"`
}

type ciHit struct {
	Provider string `json:"provider"`
	Path     string `json:"path"`
}

type deployHit struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
}
