// Package validation implements the 6-step funnel from §10 of the plan:
// lex/parse → source-evidence → compile → ArchUnit (deferred) → unit
// tests → integration tests. It is the verification mechanism that
// backstops LLM-only transformation.
//
// During migration phases (1-11), enforcer/spotless/archunit are skipped
// because they are intentionally off in the migration-mode parent POM.
// The activation phase (12) is the only place the full funnel runs with
// enforcement on; the activator specialist invokes RunActivation directly.
package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// Step identifies which step of the funnel is running.
type Step string

const (
	StepLexParse        Step = "lex_parse"
	StepSourceEvidence  Step = "source_evidence"
	StepCompile         Step = "compile"
	StepArchUnit        Step = "archunit"
	StepUnitTests       Step = "unit_tests"
	StepIntegrationTest Step = "integration_tests"
)

// Result is the outcome of running the funnel against a candidate change.
type Result struct {
	Passed       bool          `json:"passed"`
	FailedStep   Step          `json:"failedStep,omitempty"`
	BlockerCode  types.BlockerCode `json:"blockerCode,omitempty"`
	FailureLog   string        `json:"failureLog,omitempty"`
	Duration     time.Duration `json:"duration"`
}

// Mode picks which steps run. Migration phases use ModeMigration (skips
// enforcer/spotless/archunit, since they're deliberately deferred).
// Activation phase uses ModeActivation (runs everything).
type Mode int

const (
	ModeMigration Mode = iota
	ModeActivation
)

// Run executes the full funnel against the given repo state. The orchestrator
// calls this after a specialist produces output, before showing the user a
// gate. Failures are auto-fed back to the specialist for retry.
//
// affectedModules is the list of Maven module paths the change touched
// (e.g., ["model", "api"]); the funnel narrows compile/test runs to those
// modules where possible.
func Run(repoRoot string, mode Mode, affectedModules []string) Result {
	start := time.Now()

	// Step 1: lex/parse. We rely on the compile step to surface parse
	// errors with full context; this is a no-op placeholder for now and
	// the compile step's failures are classified as COMPILE_FAILED.
	// (A future tree-sitter pass could fail faster with cheaper feedback.)

	// Step 2: source-evidence is enforced by the orchestrator at the
	// patch-output boundary (it parses each OutputItem.SourceEvidence
	// against the source file). This step is included here for parity
	// with the funnel diagram in the plan but does not run again.

	// Step 3: compile.
	if log, ok := runMavenCompile(repoRoot, affectedModules); !ok {
		return Result{
			Passed:      false,
			FailedStep:  StepCompile,
			BlockerCode: types.BlockerCompileFailed,
			FailureLog:  log,
			Duration:    time.Since(start),
		}
	}

	// Step 4: ArchUnit. Deferred during migration (see §5 of plan); the
	// `trabuco-arch` JUnit tag is excluded from Surefire in migration-mode
	// parent POM. Activation removes the exclusion.
	if mode == ModeActivation {
		if log, ok := runArchUnit(repoRoot); !ok {
			return Result{
				Passed:      false,
				FailedStep:  StepArchUnit,
				BlockerCode: types.BlockerArchUnitFailed,
				FailureLog:  log,
				Duration:    time.Since(start),
			}
		}
	}

	// Step 5+6: tests. mvn test runs both unit + integration via Surefire.
	if log, ok := runMavenTests(repoRoot, affectedModules); !ok {
		// Distinguish step by inspecting log content; unit failures and
		// integration failures both manifest as test failures from
		// Surefire, so we lump them under TESTS_REGRESSED for now and
		// rely on FailureLog for diagnosis.
		return Result{
			Passed:      false,
			FailedStep:  StepUnitTests,
			BlockerCode: types.BlockerTestsRegressed,
			FailureLog:  log,
			Duration:    time.Since(start),
		}
	}

	return Result{Passed: true, Duration: time.Since(start)}
}

// runMavenCompile invokes mvn compile, scoped to affectedModules when
// possible.
func runMavenCompile(repoRoot string, modules []string) (string, bool) {
	args := []string{"compile", "-q", "-DskipTests"}
	if len(modules) > 0 {
		args = append(args, "-pl", joinModulePaths(modules), "-am")
	}
	return runMaven(repoRoot, args...)
}

// runMavenTests invokes mvn test (which on Trabuco projects includes
// Testcontainers integration tests via Surefire).
func runMavenTests(repoRoot string, modules []string) (string, bool) {
	args := []string{"test", "-q"}
	if len(modules) > 0 {
		args = append(args, "-pl", joinModulePaths(modules), "-am")
	}
	return runMaven(repoRoot, args...)
}

// runArchUnit triggers ArchUnit-specific tests by name. Trabuco generates
// these in the Shared module under the `trabuco-arch` JUnit tag;
// activation re-enables them, and the funnel runs them as a separate
// step.
//
// Scope to modules that actually contain a boundary test file. Without
// scoping, `mvn test -Dgroups=trabuco-arch` runs on the entire reactor
// and surefire fails on any module that lacks a JUnit engine on its test
// classpath ("groups/excludedGroups require ... a specific engine
// required on classpath"). Empty modules like model/ are common, so
// always-full-reactor breaks fixtures with sparse tests.
func runArchUnit(repoRoot string) (string, bool) {
	modules := findArchUnitModules(repoRoot)
	if len(modules) == 0 {
		// No boundary tests in the project — skip the step entirely.
		// Running surefire with -Dgroups against the whole reactor would
		// fail on every module that lacks a JUnit engine on classpath
		// (typical for empty model/ modules), so a no-op is correct here.
		return "no ArchitectureTest files found — step skipped", true
	}
	args := []string{"test", "-q", "-Dgroups=trabuco-arch", "-DfailIfNoTests=false",
		"-pl", joinModulePaths(modules), "-am"}
	return runMaven(repoRoot, args...)
}

// findArchUnitModules walks repoRoot looking for *ArchitectureTest*.java
// files under any first-level module's src/test/. Returns the list of
// module names (top-level dirs that contain such a file).
func findArchUnitModules(repoRoot string) []string {
	entries, err := os.ReadDir(repoRoot)
	if err != nil {
		return nil
	}
	seen := make(map[string]struct{})
	var modules []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") || name == "target" || name == "legacy" {
			continue
		}
		testRoot := filepath.Join(repoRoot, name, "src", "test")
		_ = filepath.Walk(testRoot, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".java") {
				return nil
			}
			if !strings.Contains(filepath.Base(path), "ArchitectureTest") {
				return nil
			}
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				modules = append(modules, name)
			}
			return nil
		})
	}
	return modules
}

// runMaven runs mvn with the given args, using the repo's mvnw if present.
// Returns combined output and a success boolean.
func runMaven(repoRoot string, args ...string) (string, bool) {
	mvn := "mvn"
	if _, err := os.Stat(filepath.Join(repoRoot, "mvnw")); err == nil {
		mvn = "./mvnw"
	}
	cmd := exec.Command(mvn, args...)
	cmd.Dir = repoRoot
	out, err := cmd.CombinedOutput()
	return string(out), err == nil
}

func joinModulePaths(modules []string) string {
	prefixed := make([]string, len(modules))
	for i, m := range modules {
		prefixed[i] = ":" + m
	}
	return strings.Join(prefixed, ",")
}

// VerifyEvidence checks a SourceEvidence struct against the actual source
// file. Returns nil if the evidence is valid, or an error indicating which
// of the four checks failed (file exists, lines parseable, lines exist,
// content hash matches). The orchestrator runs this on every applied patch
// before accepting it.
func VerifyEvidence(repoRoot string, ev *types.SourceEvidence) error {
	if ev == nil {
		return fmt.Errorf("source_evidence is nil")
	}
	if ev.File == "" {
		return fmt.Errorf("source_evidence.file is empty")
	}
	full := filepath.Join(repoRoot, ev.File)
	data, err := os.ReadFile(full)
	if err != nil {
		return fmt.Errorf("source_evidence.file not readable: %w", err)
	}
	startLine, endLine, err := parseLineRange(ev.Lines)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	if startLine < 1 || startLine > len(lines) || startLine > endLine {
		return fmt.Errorf("source_evidence.lines %q out of range (file has %d lines)", ev.Lines, len(lines))
	}
	// Tolerate small endLine overshoots (LLM off-by-one is common; the
	// startLine being valid already proves the file was read). Clamp.
	if endLine > len(lines) {
		endLine = len(lines)
	}
	// content_hash is optional. Specialists that emit it get verified;
	// specialists that omit it are trusted on file+lines anchoring alone.
	// LLMs cannot reliably compute sha256, so requiring it would force
	// every prompt to attach a precomputed hash table — too brittle for
	// the marginal out-of-scope guard it provides.
	if ev.ContentHash != "" {
		excerpt := strings.Join(lines[startLine-1:endLine], "\n")
		got := sha256Hex([]byte(excerpt))
		want := strings.TrimPrefix(ev.ContentHash, "sha256:")
		if got != want {
			return fmt.Errorf("source_evidence.content_hash mismatch (got %s, want %s) — source has changed since assessment", got, want)
		}
	}
	return nil
}

// parseLineRange parses a "start-end" or "start" line spec.
func parseLineRange(spec string) (int, int, error) {
	parts := strings.SplitN(spec, "-", 2)
	start, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("source_evidence.lines start: %w", err)
	}
	if len(parts) == 1 {
		return start, start, nil
	}
	end, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("source_evidence.lines end: %w", err)
	}
	return start, end, nil
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
