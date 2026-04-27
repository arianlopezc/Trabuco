// Package llm provides a shared specialist implementation that drives
// LLM-only specialists (M2-M10). Each specialist supplies its own system
// prompt and an optional input builder; the runner handles the LLM call,
// JSON parsing, and validation common to all specialists.
//
// Design: the system prompt does the reasoning work. The Go side packages
// the assessment + state into a user message, asks for a JSON response
// matching specialists.Output, and validates the result against the no-out-
// of-scope contract before returning to the orchestrator.
package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/ai"
	"github.com/arianlopezc/Trabuco/internal/migration/specialists"
	"github.com/arianlopezc/Trabuco/internal/migration/state"
	"github.com/arianlopezc/Trabuco/internal/migration/types"
)

// Spec configures one LLM-driven specialist. All M2-M10 specialists are
// built by passing one of these to New().
type Spec struct {
	// Phase is the phase this specialist owns.
	Phase types.Phase

	// Name is the canonical specialist name (e.g., "assessor", "model").
	Name string

	// SystemPrompt is the full system prompt for the specialist. In
	// practice this is the body of plugin/agents/trabuco-migration-{name}.md
	// loaded into the Go binary at compile time via embed.
	SystemPrompt string

	// MaxTokens caps the LLM response length. Defaults to a per-phase
	// reasonable value if zero.
	MaxTokens int

	// BuildUserPrompt is an optional override that lets a specialist
	// customize what goes into the user message. Default: a JSON dump of
	// the input augmented with the assessment artifact.
	BuildUserPrompt func(in *specialists.Input) (string, error)
}

// Specialist is the LLM-driven implementation of specialists.Specialist.
type Specialist struct {
	spec     Spec
	provider ai.Provider
}

// New constructs a Specialist from a Spec. The provider is resolved lazily
// on first Run() so importers don't need to wire up auth at registration
// time.
func New(spec Spec) *Specialist {
	return &Specialist{spec: spec}
}

// Phase implements specialists.Specialist.
func (s *Specialist) Phase() types.Phase { return s.spec.Phase }

// Name implements specialists.Specialist.
func (s *Specialist) Name() string { return s.spec.Name }

// Run implements specialists.Specialist. Builds the prompt, calls the LLM,
// parses the JSON output, and returns it.
func (s *Specialist) Run(ctx context.Context, in *specialists.Input) (*specialists.Output, error) {
	if s.provider == nil {
		p, err := defaultProvider()
		if err != nil {
			return nil, fmt.Errorf("no LLM provider available (run 'trabuco auth login' first): %w", err)
		}
		s.provider = p
	}

	user, err := s.buildUserPrompt(in)
	if err != nil {
		return nil, err
	}

	maxTokens := s.spec.MaxTokens
	if maxTokens == 0 {
		maxTokens = 8000
	}

	req := &ai.AnalysisRequest{
		SystemPrompt: s.spec.SystemPrompt + "\n\n" + outputContract,
		UserPrompt:   user,
		MaxTokens:    maxTokens,
		Temperature:  0.2, // mostly-deterministic; prompts demand JSON
	}
	resp, err := s.provider.Analyze(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("LLM call: %w", err)
	}

	// Persist raw LLM output for debugging. Best-effort; failure here
	// must not mask the real result.
	_ = state.WriteRawLLM(in.RepoRoot, s.spec.Phase, s.spec.Name, resp.Content)

	out, err := parseOutput(resp.Content, s.spec.Phase)
	if err != nil {
		return nil, fmt.Errorf("parse LLM output: %w (content: %s)", err, truncate(resp.Content, 1000))
	}
	return out, nil
}

// buildUserPrompt is the default implementation; specialists can override
// via Spec.BuildUserPrompt.
func (s *Specialist) buildUserPrompt(in *specialists.Input) (string, error) {
	if s.spec.BuildUserPrompt != nil {
		return s.spec.BuildUserPrompt(in)
	}
	return DefaultUserPrompt(in)
}

// DefaultUserPrompt is the baseline user-prompt structure: phase context +
// state.json + assessment (if present) + the actual source files relevant
// to this phase (with line numbers, so the LLM can emit accurate
// source_evidence ranges). Specialists with phase-specific inputs override
// BuildUserPrompt to customize further.
func DefaultUserPrompt(in *specialists.Input) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "Phase: %d (%s)\n", int(in.Phase), in.Phase)
	fmt.Fprintf(&b, "Repo root: %s\n\n", in.RepoRoot)
	if in.UserHint != "" {
		fmt.Fprintf(&b, "User guidance from previous gate (apply this in your output):\n%s\n\n", in.UserHint)
	}
	if in.Aggregate != "" {
		fmt.Fprintf(&b, "Restrict scope to aggregate: %s\n\n", in.Aggregate)
	}

	stateJSON, err := json.MarshalIndent(in.State, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal state: %w", err)
	}
	fmt.Fprintf(&b, "Current state.json:\n```json\n%s\n```\n\n", stateJSON)

	if in.Phase != types.PhaseAssessment {
		// Assessment artifact should exist by now; include for grounding.
		assessPath := state.AssessmentPath(in.RepoRoot)
		if data, err := readFileBest(assessPath); err == nil {
			fmt.Fprintf(&b, "Assessment artifact (.trabuco-migration/assessment.json):\n```json\n%s\n```\n\n", data)
		}

		// Embed the source files this phase needs, with line numbers, so
		// the LLM can emit accurate source_evidence and produce informed
		// migrations grounded in real code. relevantFilePaths returns
		// just the legacy/* paths the phase consumes.
		paths := relevantFilePaths(in)
		if len(paths) > 0 {
			fmt.Fprintf(&b, "## Source files in scope for this phase\n\n")
			fmt.Fprintf(&b, "Each file below is shown with line numbers. Emit source_evidence.lines ranges that fall within the actual line counts you see here. content_hash is optional — if omitted, the orchestrator validates only file+lines.\n\n")
			for _, p := range paths {
				if body, err := readFileBest(filepath.Join(in.RepoRoot, p)); err == nil {
					fmt.Fprintf(&b, "### %s\n```\n%s```\n\n", p, withLineNumbers(body))
				}
			}
		}

		// For phases that build on prior phases (shared, api, worker,
		// eventconsumer, aiagent, config, deployment, tests, activation,
		// finalization), include every Trabuco-module .java file produced
		// by earlier phases. Without this the LLM hallucinates the
		// packages of classes its code imports.
		if needsPriorTrabucoSources(in.Phase) {
			if files := listTrabucoModuleSources(in.RepoRoot); len(files) > 0 {
				fmt.Fprintf(&b, "## Trabuco modules built so far (for reference; do NOT modify unless the phase scope demands it)\n\n")
				for _, f := range files {
					if body, err := readFileBest(filepath.Join(in.RepoRoot, f)); err == nil {
						fmt.Fprintf(&b, "### %s\n```\n%s```\n\n", f, body)
					}
				}
			}
		}

		// Always include the parent pom.xml + every module pom.xml that
		// already exists. Without this, specialists that touch any
		// pom.xml hallucinate version strings, groupIds, etc. Costs a
		// few KB but eliminates whole categories of compile failures.
		fmt.Fprintf(&b, "## Current Maven POMs (parent + every module). Do NOT change groupId/artifactId/version when replacing these — copy the <parent> block character-for-character.\n\n")
		pomCandidates := []string{"pom.xml"}
		for _, m := range []string{"legacy", "model", "sqldatastore", "nosqldatastore", "shared", "api", "worker", "eventconsumer", "aiagent"} {
			pomCandidates = append(pomCandidates, filepath.Join(m, "pom.xml"))
		}
		for _, p := range pomCandidates {
			if body, err := readFileBest(filepath.Join(in.RepoRoot, p)); err == nil {
				fmt.Fprintf(&b, "### %s\n```xml\n%s```\n\n", p, body)
			}
		}
	}

	return b.String(), nil
}

// needsPriorTrabucoSources reports whether this phase needs to see what
// earlier phases already produced in the Trabuco target modules.
func needsPriorTrabucoSources(p types.Phase) bool {
	switch p {
	case types.PhaseShared, types.PhaseAPI, types.PhaseWorker,
		types.PhaseEventConsumer, types.PhaseAIAgent,
		types.PhaseConfiguration, types.PhaseDeployment, types.PhaseTests,
		types.PhaseActivation, types.PhaseFinalization:
		return true
	}
	return false
}

// listTrabucoModuleSources returns every .java file under known Trabuco
// module directories at repoRoot. Excludes legacy/, build outputs, and
// the migration state dir.
func listTrabucoModuleSources(repoRoot string) []string {
	modules := []string{"model", "sqldatastore", "nosqldatastore", "shared", "api", "worker", "eventconsumer", "aiagent"}
	var out []string
	for _, m := range modules {
		root := filepath.Join(repoRoot, m, "src", "main", "java")
		_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info == nil || info.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".java") {
				return nil
			}
			if rel, err := filepath.Rel(repoRoot, path); err == nil {
				out = append(out, rel)
			}
			return nil
		})
	}
	return out
}

// relevantFilePaths returns the set of source files the given phase
// should reason over, derived from the assessment. Every path is
// repo-relative.
func relevantFilePaths(in *specialists.Input) []string {
	a, err := loadAssessmentMap(state.AssessmentPath(in.RepoRoot))
	if err != nil {
		return nil
	}
	var paths []string
	// Every module-shaped phase wants the root parent POM (so the LLM
	// uses the right groupId/artifactId for module <parent>) and the
	// module's own scaffolded POM (so it knows the stub exists and uses
	// "replace" instead of "create"). modulePOMHint maps phase → module
	// directory name.
	modulePOMHint := map[types.Phase]string{
		types.PhaseModel:         "model",
		types.PhaseDatastore:     "sqldatastore",
		types.PhaseShared:        "shared",
		types.PhaseAPI:           "api",
		types.PhaseWorker:        "worker",
		types.PhaseEventConsumer: "eventconsumer",
		types.PhaseAIAgent:       "aiagent",
	}
	if mod, ok := modulePOMHint[in.Phase]; ok {
		paths = append(paths, "pom.xml", mod+"/pom.xml")
	}

	switch in.Phase {
	case types.PhaseModel:
		paths = append(paths, collectFileField(a, "entities")...)
	case types.PhaseDatastore:
		paths = append(paths, collectFileField(a, "repositories")...)
		paths = append(paths, collectFileField(a, "entities")...)
	case types.PhaseShared:
		paths = append(paths, collectFileField(a, "services")...)
	case types.PhaseAPI:
		paths = append(paths, collectFileField(a, "controllers")...)
	case types.PhaseWorker:
		paths = append(paths, collectFileField(a, "jobs")...)
	case types.PhaseEventConsumer:
		paths = append(paths, collectFileField(a, "listeners")...)
		paths = append(paths, collectFileField(a, "publishers")...)
	case types.PhaseAIAgent:
		// AI files aren't yet a separate Assessment field; rely on the
		// LLM reading the assessment to decide which services/controllers
		// are AI-related.
	case types.PhaseTests:
		paths = collectFileField(a, "tests")
	case types.PhaseConfiguration:
		if cf, ok := a["configFiles"].([]any); ok {
			for _, p := range cf {
				if s, ok := p.(string); ok {
					paths = append(paths, s)
				}
			}
		}
	case types.PhaseDeployment:
		// Deployment files paths come from ciSystems[].files and
		// deploymentFiles[].file.
		if ci, ok := a["ciSystems"].([]any); ok {
			for _, sys := range ci {
				if m, ok := sys.(map[string]any); ok {
					if files, ok := m["files"].([]any); ok {
						for _, f := range files {
							if s, ok := f.(string); ok {
								paths = append(paths, s)
							}
						}
					}
				}
			}
		}
		paths = append(paths, collectFileField(a, "deploymentFiles")...)
	}
	if in.Aggregate != "" {
		paths = filterByAggregate(paths, in.Aggregate)
	}
	return paths
}

// loadAssessmentMap reads assessment.json into a generic map so we can
// pluck file paths without importing the assessor package (which would
// create a circular import).
func loadAssessmentMap(path string) (map[string]any, error) {
	data, err := readFileBest(path)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// collectFileField pulls "file" entries from an array-of-objects field.
func collectFileField(m map[string]any, field string) []string {
	arr, ok := m[field].([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, item := range arr {
		obj, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if f, ok := obj["file"].(string); ok && f != "" {
			out = append(out, f)
		}
	}
	return out
}

// filterByAggregate returns only paths whose corresponding entry in the
// assessment matches the given aggregate. For now we approximate by
// checking whether the path contains the aggregate substring — the
// assessor sets aggregate to a package-fragment-like string.
func filterByAggregate(paths []string, aggregate string) []string {
	var out []string
	for _, p := range paths {
		if strings.Contains(p, "/"+aggregate+"/") || strings.Contains(p, "/"+strings.ToLower(aggregate)+"/") {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		// If filtering empties the list we likely got the aggregate name
		// wrong; fall back to all paths so the LLM at least has context.
		return paths
	}
	return out
}

// withLineNumbers prefixes each line with "%4d  " so the LLM can pick
// exact line ranges.
func withLineNumbers(content string) string {
	lines := strings.Split(content, "\n")
	var b strings.Builder
	for i, line := range lines {
		fmt.Fprintf(&b, "%4d  %s\n", i+1, line)
	}
	return b.String()
}

// outputContract is appended to every system prompt; it tells the LLM
// exactly what JSON shape to return so the Go side can parse without LLM-
// specific tooling. The shape mirrors specialists.Output verbatim.
const outputContract = `

# Output contract

You MUST respond with a single JSON object matching this Go struct exactly.
No surrounding prose, no Markdown fences, no commentary — just the JSON.

` + "```json" + `
{
  "phase": <number 0-13>,
  "summary": "<one-paragraph summary of what you did or why nothing was done>",
  "items": [
    {
      "id": "<unique id within this phase>",
      "state": "applied" | "blocked" | "requires_decision" | "not_applicable" | "retained_legacy",
      "description": "<what this item represents>",
      "source_evidence": {
        "file": "<path relative to repo root>",
        "lines": "<start-end, e.g. 12-58>",
        "content_hash": "<sha256 of the byte range, hex>"
      },
      "file_writes": [
        {
          "path": "<path relative to repo root>",
          "operation": "create" | "replace" | "delete",
          "content": "<full file content; required for create/replace, omit for delete>"
        }
      ],
      "blocker_code": "<one of the fixed BlockerCode enum values, only when state=blocked>",
      "blocker_note": "<concrete file:line context, only when state=blocked>",
      "alternatives": ["<alternative 1>", "<alternative 2>"],
      "question": "<the decision question, only when state=requires_decision>",
      "choices": ["<choice 1>", "<choice 2>"],
      "reason": "<why this is not_applicable, only when state=not_applicable>"
    }
  ],
  "decisions": [
    { "id": "<...>", "question": "<...>", "choices": ["<...>"], "context": "<...>" }
  ]
}
` + "```" + `

# Assessor exception (Phase 0 ONLY)

The Phase 0 assessor is the ONLY specialist that uses ` + "`patch`" + ` instead of
` + "`file_writes`" + `. It emits exactly ONE OutputItem with state="applied",
description="initial assessment", and ` + "`patch`" + ` set to the JSON-stringified
Assessment struct. The assessor MUST NOT use file_writes — its prompt
overrides this section. All other phases (1-13) MUST use file_writes
exclusively and leave ` + "`patch`" + ` empty.

# How to express changes via file_writes (phases 1-13)

For every item with state=applied, your patch is expressed as one or more
file_writes. Each file_write has:
- path: relative to repo root (no leading slash, no traversal like '..')
- operation: "create", "replace", or "delete"
- content: the FULL file content for create/replace (do NOT use diff
  syntax; do NOT abbreviate with '...'); omit for delete

Examples:

Adding a new entity:
  { "path": "model/src/main/java/com/x/model/User.java",
    "operation": "create",
    "content": "package com.x.model;\n\npublic record User(...) {}\n" }

Marking the legacy class @Deprecated:
  { "path": "legacy/src/main/java/com/x/User.java",
    "operation": "replace",
    "content": "<full new content with @Deprecated added>" }

Deleting an obsolete file:
  { "path": "legacy/src/main/java/com/x/Old.java",
    "operation": "delete" }

# Module catalog is fixed

The set of Trabuco modules is FIXED by the target config in state.json
and was generated by the skeleton-builder in Phase 1. Read
state.targetConfig.modules to see the authoritative list. You MUST NOT:
- Create a new top-level Maven module directory (e.g. jobs/, lib/,
  common/) that isn't already in target.modules.
- Add a new <module>X</module> entry to the parent pom.xml's <modules>
  section.

If your phase's logical artifacts don't fit any existing module, place
them in the most appropriate existing module using a sub-package, NOT
a new module.

# Module POM dependency hygiene

If your phase introduces code into a Trabuco module (model/, sqldatastore/,
shared/, api/, worker/, eventconsumer/, aiagent/), you MUST also keep that
module's pom.xml in sync with the dependencies your code uses. The
skeleton-builder left every module/pom.xml as a minimal stub (just the
parent declaration and artifactId). When you introduce, e.g., spring-data-
jdbc imports in a new entity, you must replace the module pom.xml to add
the matching dependencies block. Every dependency version should be
inherited from the parent's BOM where possible (spring-boot-dependencies);
add an explicit version only if it isn't BOM-managed.

CRITICAL: when replacing a pom.xml, copy the existing <parent> block
verbatim. In particular, the <version> inside <parent> AND the project's
own <version> must match what's currently in the parent pom.xml — DO NOT
"normalize" 1.0-SNAPSHOT to 1.0.0-SNAPSHOT, do NOT change groupId or
artifactId. The user's prompt includes the current pom files; copy them
character-for-character for the parent block, then add/modify only the
<dependencies> and <build> sections you actually need.

The compile step in the validation funnel covers the module you write to,
so a missing dependency will surface as COMPILE_FAILED and roll your
phase back. Add the dependency in the same item that introduces the code
that requires it.

# Constraints

- Every item with state=applied MUST include source_evidence pointing at
  REAL file paths and line ranges in the assessment artifact. content_hash
  is optional — if you can't compute sha256 reliably, omit it; the
  orchestrator will validate file+lines only.
- file_writes paths must be inside the repo (no '..' traversal, no leading
  '/'). The orchestrator rejects unsafe paths.
- For replace operations, you MUST provide the FULL new content. Do not
  abbreviate, do not use '...' or 'rest of file unchanged'. The new content
  REPLACES the entire file.
- For create operations, the file must NOT already exist. Use replace
  instead if it does.
- BlockerCode MUST be one of the fixed enum values from the plan. Inventing
  new codes will cause the orchestrator to reject your output.
- If your phase has no work to do (no source artifacts match its scope),
  return ONE item with state=not_applicable and an explanation in 'reason'.
  Do NOT include file_writes when state=not_applicable.
- Do NOT propose changes that aren't grounded in source evidence. The
  orchestrator's diff-inspection layer will reject scope creep.
`

func parseOutput(raw string, expectedPhase types.Phase) (*specialists.Output, error) {
	// Some models wrap responses in markdown fences despite instructions.
	clean := stripMarkdownFences(raw)

	var out specialists.Output
	if err := json.Unmarshal([]byte(clean), &out); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	if out.Phase != expectedPhase {
		return nil, fmt.Errorf("phase mismatch: got %d, want %d", out.Phase, expectedPhase)
	}

	// Validate every blocker code is in the fixed enum.
	for _, item := range out.Items {
		if item.State == types.ItemBlocked && !item.BlockerCode.IsKnown() {
			return nil, fmt.Errorf("item %s: blocker_code %q not in fixed enum", item.ID, item.BlockerCode)
		}
	}

	return &out, nil
}

func stripMarkdownFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
		if i := strings.LastIndex(s, "```"); i != -1 {
			s = s[:i]
		}
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
		if i := strings.LastIndex(s, "```"); i != -1 {
			s = s[:i]
		}
	}
	return strings.TrimSpace(s)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...[truncated]"
}
