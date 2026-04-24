// Package sync implements `trabuco sync` — an additive-only operation that
// brings a project's AI-tooling files up to date with what the installed
// CLI would generate for the same module and agent selection.
//
// Design principles:
//   - Additive only. Existing files are never modified or deleted.
//   - Jurisdiction is enforced. Only paths matching the AI-tooling allow-list
//     are considered; business code (Java, POMs, migrations, app config) is
//     physically unreachable from this package.
//   - No drift risk. The expected state is produced by running the current
//     generator against a temp directory — the generator is the single
//     source of truth, not a parallel registry.
//
// The package does NOT handle section-level merging (for files like
// CLAUDE.md that evolve over time). If a file exists in the project, sync
// leaves it alone entirely, stale content and all. Users who want to
// refresh CLAUDE.md delete it and re-run sync.
package sync

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/generator"
)

// ErrOutOfJurisdiction is returned when Apply is asked to write a path that
// is not inside the allow-list. This is a defensive safeguard on top of the
// planning-time filter — a Plan constructed through normal means will never
// contain out-of-jurisdiction entries, but Apply validates again at write
// time to catch bugs and tampering.
var ErrOutOfJurisdiction = errors.New("path is outside sync jurisdiction")

// Run executes the full sync flow: plan + optionally apply. It holds the
// generator-output directory alive between planning and writing so Apply
// can copy files from it directly (no double-generation).
//
// If apply is false, the function returns the plan and no writes happen.
// If apply is true and the plan is not blocked and has work, the missing
// files are copied into the project. The returned plan reflects what was
// actually attempted; errors during individual file writes are wrapped and
// returned.
func Run(projectPath, cliVersion string, apply bool) (*Plan, error) {
	absProject, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, fmt.Errorf("resolve project path: %w", err)
	}

	plan := &Plan{
		ProjectPath: absProject,
		CLIVersion:  cliVersion,
	}

	if !config.MetadataExists(absProject) {
		plan.Blockers = append(plan.Blockers,
			".trabuco.json not found — not a Trabuco project.")
		return plan, nil
	}

	meta, err := config.LoadMetadata(absProject)
	if err != nil {
		plan.Blockers = append(plan.Blockers,
			fmt.Sprintf("failed to read .trabuco.json: %v", err))
		return plan, nil
	}

	plan.StampedVersion = meta.Version
	plan.Modules = append([]string{}, meta.Modules...)
	plan.AIAgents = append([]string{}, meta.AIAgents...)

	tmpRoot, err := os.MkdirTemp("", "trabuco-sync-")
	if err != nil {
		return nil, fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpRoot)

	expectedDir := filepath.Join(tmpRoot, "expected")

	cfg := meta.ToProjectConfig()
	// `.trabuco.json` does not carry review configuration. That decision lives
	// in `.trabuco/review.config.json` alongside the generated review
	// artifacts, so we reconstruct the effective Review.Mode from that file
	// (or default to "full" — init's default — when absent) so the simulated
	// generation emits the same set of review subagents, hooks, and skills
	// the project started with.
	applyReviewConfig(cfg, absProject)

	gen, err := generator.NewWithVersionAt(cfg, cliVersion, expectedDir)
	if err != nil {
		return nil, fmt.Errorf("initialize generator: %w", err)
	}

	restoreStdout := silenceStdout()
	genErr := gen.Generate()
	restoreStdout()
	if genErr != nil {
		return nil, fmt.Errorf("simulate project generation: %w", genErr)
	}

	walkErr := filepath.WalkDir(expectedDir, func(absPath string, d fs.DirEntry, errArg error) error {
		if errArg != nil {
			return errArg
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(expectedDir, absPath)
		if relErr != nil {
			return relErr
		}
		rel = filepath.ToSlash(rel)

		if !InJurisdiction(rel) {
			plan.OutOfJurisdiction = append(plan.OutOfJurisdiction, rel)
			return nil
		}

		projectFile := filepath.Join(absProject, filepath.FromSlash(rel))
		if _, statErr := os.Stat(projectFile); errors.Is(statErr, fs.ErrNotExist) {
			plan.WouldAdd = append(plan.WouldAdd, rel)
		} else if statErr != nil {
			return fmt.Errorf("stat %s: %w", projectFile, statErr)
		} else {
			plan.AlreadyPresent = append(plan.AlreadyPresent, rel)
		}
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("walk expected state: %w", walkErr)
	}

	if !apply || plan.Blocked() || !plan.HasWork() {
		return plan, nil
	}

	// Apply: copy each WouldAdd file from the expected dir into the project.
	// Re-validate jurisdiction at write time as a defense-in-depth check.
	for _, rel := range plan.WouldAdd {
		if !InJurisdiction(rel) {
			return plan, fmt.Errorf("%w: %s", ErrOutOfJurisdiction, rel)
		}
		src := filepath.Join(expectedDir, filepath.FromSlash(rel))
		dst := filepath.Join(absProject, filepath.FromSlash(rel))
		if err := copyFile(src, dst); err != nil {
			return plan, fmt.Errorf("copy %s: %w", rel, err)
		}
	}

	return plan, nil
}

// copyFile copies src to dst, creating parent directories and preserving
// the source's permission bits (important for executable scripts like
// review-checks.sh and require-review.sh).
func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer srcFile.Close()

	// Write to a temp file in the same directory, then rename for atomicity.
	// This means a failed write never leaves a partial file in the project.
	tmpFile, err := os.CreateTemp(filepath.Dir(dst), ".trabuco-sync-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	// On any error after this point, clean up the temp file.
	cleanup := func() { _ = os.Remove(tmpPath) }

	if _, err := io.Copy(tmpFile, srcFile); err != nil {
		tmpFile.Close()
		cleanup()
		return fmt.Errorf("copy contents: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		cleanup()
		return fmt.Errorf("close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, srcInfo.Mode().Perm()); err != nil {
		cleanup()
		return fmt.Errorf("chmod: %w", err)
	}
	if err := os.Rename(tmpPath, dst); err != nil {
		cleanup()
		return fmt.Errorf("rename: %w", err)
	}
	return nil
}

// applyReviewConfig hydrates cfg.Review from the project's on-disk review
// config. If `.trabuco/review.config.json` is missing or unreadable, defaults
// to "full" — matching what `trabuco init` picks when the user doesn't
// override it. A project that was explicitly generated with review disabled
// has mode:"off" in the file, and sync honors that (so no review artifacts
// are added by sync either).
func applyReviewConfig(cfg *config.ProjectConfig, projectPath string) {
	type reviewConfigFile struct {
		Mode        string `json:"mode"`
		GeneratedAt string `json:"generatedAt"`
	}

	path := filepath.Join(projectPath, ".trabuco", "review.config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		cfg.Review.Mode = config.ReviewModeFull
		return
	}

	var rcf reviewConfigFile
	if err := json.Unmarshal(data, &rcf); err != nil || rcf.Mode == "" {
		cfg.Review.Mode = config.ReviewModeFull
		return
	}
	cfg.Review.Mode = rcf.Mode
	cfg.Review.GeneratedAt = rcf.GeneratedAt
}

// silenceStdout redirects os.Stdout AND the fatih/color package's cached
// writer to /dev/null for the duration of the returned closure's lifetime.
// The color package captures os.Stdout at init, so swapping os.Stdout alone
// leaves colorized output leaking. Both must be redirected to suppress the
// generator's "Generating project..." progress chatter during sync.
func silenceStdout() func() {
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		// If /dev/null can't be opened, the generator's output is merely
		// cosmetic — fall through rather than failing the sync.
		return func() {}
	}
	origStdout := os.Stdout
	origColor := color.Output
	os.Stdout = devNull
	color.Output = devNull
	return func() {
		os.Stdout = origStdout
		color.Output = origColor
		devNull.Close()
	}
}
