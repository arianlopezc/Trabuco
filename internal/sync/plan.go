package sync

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Plan describes the changeset sync would apply to a project. The plan is
// purely additive: WouldAdd lists files missing in the project that the
// current CLI would generate. Existing files are never modified.
type Plan struct {
	ProjectPath       string   `json:"project_path"`
	CLIVersion        string   `json:"cli_version"`
	StampedVersion    string   `json:"stamped_version"`
	Modules           []string `json:"modules"`
	AIAgents          []string `json:"ai_agents"`
	WouldAdd          []string `json:"would_add"`
	AlreadyPresent    []string `json:"already_present"`
	OutOfJurisdiction []string `json:"out_of_jurisdiction,omitempty"`
	Blockers          []string `json:"blockers,omitempty"`
}

// HasWork reports whether the plan has any additive operations.
func (p *Plan) HasWork() bool {
	return len(p.WouldAdd) > 0
}

// Blocked reports whether the plan cannot be applied.
func (p *Plan) Blocked() bool {
	return len(p.Blockers) > 0
}

// WriteJSON serializes the plan to JSON for machine consumption.
func (p *Plan) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(p)
}

// WritePretty renders a human-readable summary of the plan.
func (p *Plan) WritePretty(w io.Writer) error {
	if p.Blocked() {
		fmt.Fprintf(w, "Trabuco Sync — cannot proceed — %s\n\n", p.ProjectPath)
		for _, b := range p.Blockers {
			fmt.Fprintf(w, "  ✗ %s\n", b)
		}
		fmt.Fprintln(w)
		return nil
	}

	fmt.Fprintf(w, "Trabuco Sync — %s\n\n", p.ProjectPath)
	if p.StampedVersion != "" {
		fmt.Fprintf(w, "Project generated at CLI: %s\n", p.StampedVersion)
	}
	if p.CLIVersion != "" {
		fmt.Fprintf(w, "Current CLI version:      %s\n", p.CLIVersion)
	}
	if len(p.Modules) > 0 {
		fmt.Fprintf(w, "Modules:                  %s\n", strings.Join(p.Modules, ", "))
	}
	if len(p.AIAgents) > 0 {
		fmt.Fprintf(w, "AI agents:                %s\n", strings.Join(p.AIAgents, ", "))
	}
	fmt.Fprintln(w)

	if !p.HasWork() {
		fmt.Fprintln(w, "All AI-tooling files expected for this project's configuration are present.")
		fmt.Fprintln(w, "Nothing to sync.")
		return nil
	}

	adds := append([]string(nil), p.WouldAdd...)
	sort.Strings(adds)
	fmt.Fprintf(w, "Would add %d files:\n", len(adds))
	for _, path := range adds {
		fmt.Fprintf(w, "  + %s\n", path)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Existing files are NOT modified by sync.")
	return nil
}
