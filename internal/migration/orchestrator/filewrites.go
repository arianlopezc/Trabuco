package orchestrator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// applyFileWrites materializes every applied item's FileWrites onto disk.
// Operations are applied in order, atomically per-item: if any write in
// an item fails, the whole item is rolled back to its pre-state.
//
// The orchestrator calls this AFTER the specialist returns and BEFORE
// the validation funnel runs.
func applyFileWrites(repoRoot string, out *specialists.Output) error {
	for i := range out.Items {
		item := &out.Items[i]
		if item.State != types.ItemApplied {
			continue
		}
		if len(item.FileWrites) == 0 {
			// applied with no file_writes is allowed if SourceEvidence
			// records something the specialist did via state alone (rare),
			// but generally indicates a malformed specialist output. The
			// orchestrator's source_evidence verification will catch it.
			continue
		}
		applied, err := applyOneItem(repoRoot, item.FileWrites)
		if err != nil {
			rollbackApplied(repoRoot, applied)
			return fmt.Errorf("item %s: %w", item.ID, err)
		}
	}
	return nil
}

// applyOneItem applies all file writes for a single item and returns the
// list of paths it touched (for rollback if a later op fails).
func applyOneItem(repoRoot string, writes []types.FileWrite) ([]appliedWrite, error) {
	applied := make([]appliedWrite, 0, len(writes))
	for _, w := range writes {
		full, err := safeJoin(repoRoot, w.Path)
		if err != nil {
			return applied, fmt.Errorf("%s: %w", w.Path, err)
		}
		switch w.Operation {
		case types.OpCreate:
			if _, err := os.Stat(full); err == nil {
				return applied, fmt.Errorf("%s: create requested but file already exists", w.Path)
			}
			prev := appliedWrite{path: full, op: types.OpCreate}
			if err := writeFile(full, w.Content); err != nil {
				return applied, fmt.Errorf("%s: %w", w.Path, err)
			}
			applied = append(applied, prev)

		case types.OpReplace:
			old, err := os.ReadFile(full)
			if err != nil && !os.IsNotExist(err) {
				return applied, fmt.Errorf("%s: %w", w.Path, err)
			}
			prev := appliedWrite{path: full, op: types.OpReplace, oldContent: old, oldExisted: err == nil}
			if err := writeFile(full, w.Content); err != nil {
				return applied, fmt.Errorf("%s: %w", w.Path, err)
			}
			applied = append(applied, prev)

		case types.OpDelete:
			old, err := os.ReadFile(full)
			if err != nil {
				if os.IsNotExist(err) {
					// idempotent delete
					applied = append(applied, appliedWrite{path: full, op: types.OpDelete})
					continue
				}
				return applied, fmt.Errorf("%s: %w", w.Path, err)
			}
			prev := appliedWrite{path: full, op: types.OpDelete, oldContent: old, oldExisted: true}
			if err := os.Remove(full); err != nil {
				return applied, fmt.Errorf("%s: %w", w.Path, err)
			}
			applied = append(applied, prev)

		default:
			return applied, fmt.Errorf("%s: unknown operation %q", w.Path, w.Operation)
		}
	}
	return applied, nil
}

// rollbackApplied undoes every successfully-applied write for an item
// when a later write in the same item fails.
func rollbackApplied(repoRoot string, applied []appliedWrite) {
	// Reverse order so a create-then-replace inside the same item is
	// rolled back to original-create-state correctly.
	for i := len(applied) - 1; i >= 0; i-- {
		a := applied[i]
		switch a.op {
		case types.OpCreate:
			_ = os.Remove(a.path)
		case types.OpReplace:
			if a.oldExisted {
				_ = writeFile(a.path, string(a.oldContent))
			} else {
				_ = os.Remove(a.path)
			}
		case types.OpDelete:
			if a.oldExisted {
				_ = writeFile(a.path, string(a.oldContent))
			}
		}
	}
}

type appliedWrite struct {
	path       string
	op         types.FileOperation
	oldContent []byte
	oldExisted bool
}

// safeJoin joins repoRoot and rel, refusing absolute paths, paths with
// any '..' segment (even ones that resolve inside the repo — they're a
// code smell), and paths that escape repoRoot after resolution.
func safeJoin(repoRoot, rel string) (string, error) {
	if rel == "" {
		return "", fmt.Errorf("empty path")
	}
	if filepath.IsAbs(rel) {
		return "", fmt.Errorf("absolute path forbidden: %s", rel)
	}
	// Reject any '..' segment in the raw path. A specialist that emits
	// "foo/../bar" is confused; the orchestrator demands a clean path.
	for _, seg := range strings.Split(filepath.ToSlash(rel), "/") {
		if seg == ".." {
			return "", fmt.Errorf("path traversal forbidden: %s", rel)
		}
	}
	clean := filepath.Clean(rel)
	full := filepath.Join(repoRoot, clean)
	rootAbs, err := filepath.Abs(repoRoot)
	if err != nil {
		return "", err
	}
	fullAbs, err := filepath.Abs(full)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(fullAbs, rootAbs+string(filepath.Separator)) && fullAbs != rootAbs {
		return "", fmt.Errorf("path escapes repo root: %s", rel)
	}
	return full, nil
}

// writeFile is a small helper that creates parent dirs as needed.
func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}
