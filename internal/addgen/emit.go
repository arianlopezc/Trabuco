package addgen

import (
	"fmt"
	"os"
	"path/filepath"
)

// Result describes what an add-command produced (or would produce in
// dry-run). JSON-serialized when callers ask for --json output.
type Result struct {
	// Created lists the files this command wrote, relative to the
	// project root. In dry-run mode these are the files that *would*
	// be written.
	Created []string `json:"created"`

	// NextSteps tells the agent (and human) what to do after the
	// generator finishes — typically the edits the CLI deliberately
	// did NOT make, since this CLI is addition-only. E.g. for `add
	// event`, the next step is "add OrderShipped to the sealed
	// permits clause in OrderEvent.java".
	NextSteps []string `json:"next_steps,omitempty"`

	// Notes captures non-blocking commentary — e.g. "skipped existing
	// migration directory; using V3 as the next version".
	Notes []string `json:"notes,omitempty"`
}

// emitFile writes content to relPath under the project root, refusing
// to overwrite if the file already exists. In dry-run mode, content
// is not written but the path is still appended to result.Created so
// callers can preview the change set.
//
// Refuse-to-clobber is intentional and mirrors `git add`'s policy of
// never silently overwriting working-tree edits. There is no --force
// flag; the agent must `rm` first if it really wants to regenerate.
func (c *Context) emitFile(relPath string, content string, result *Result) error {
	abs := filepath.Join(c.ProjectPath, relPath)

	if _, err := os.Stat(abs); err == nil {
		return fmt.Errorf("refusing to overwrite existing file: %s (delete it first if you want to regenerate)", relPath)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("failed to stat %s: %w", relPath, err)
	}

	if !c.DryRun {
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", relPath, err)
		}
		if err := os.WriteFile(abs, []byte(content), 0o644); err != nil {
			return fmt.Errorf("failed to write %s: %w", relPath, err)
		}
	}

	result.Created = append(result.Created, relPath)
	return nil
}
