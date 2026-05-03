package generator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// Skill emission is the single source of truth for every workflow surfaced
// as a slash-command across Claude Code, OpenAI Codex CLI, GitHub Copilot,
// and Cursor. Three of the four tools (Claude, Codex, Copilot) have
// converged on Anthropic's SKILL.md format; Cursor has no skills primitive
// so we degrade to a `.cursor/rules/*.mdc` with glob-based activation.
//
// Content discipline: the skill body is a short summary + pointer to the
// authoritative `.ai/prompts/*.md` reference. Source of truth stays in
// `.ai/prompts/` (tool-agnostic, already cited by 20+ citations including
// generated Java code) — the skill layer is a discoverable entry point,
// not a content duplicate.

// skillDef captures everything a single workflow skill needs to emit to
// every supported tool.
type skillDef struct {
	// Name must be lowercase-letters-and-hyphens (Anthropic constraint; also
	// required by Copilot's strict name-must-match-directory rule).
	Name string

	// Description is surfaced in the agent's skill listing at session start
	// and drives autonomous match-by-description. Front-load the use-case
	// keywords; keep under ~200 chars so 17 skills fit in the session
	// context without the listing block getting truncated.
	Description string

	// ArgumentHint is the `<hint>` shown after the slash-command name in
	// Claude/Codex autocomplete. Optional.
	ArgumentHint string

	// Paths restricts where this skill is considered relevant (Anthropic
	// frontmatter). Empty means always relevant.
	Paths []string

	// BodyTmpl is the relative path under `templates/skills/` for the
	// workflow body template. The body is rendered once and wrapped in
	// tool-specific frontmatter per emission target.
	BodyTmpl string

	// RequiredModule gates emission: empty means always, "SQLDatastore"
	// means only when SQLDatastore is selected, "AIAgent" when AIAgent, etc.
	// The special value "anyDatastore" means either SQL or NoSQL datastore.
	RequiredModule string

	// Invocable toggles `user-invocable: true`. Most skills are invocable;
	// the few that exist only as reference may flip this off.
	Invocable bool

	// CursorPort controls whether this skill has a Cursor rules equivalent.
	// Cursor's rules are passive/glob-activated — unsuitable for pure
	// workflow invocations like /commit, so those are Claude/Codex/Copilot
	// only. The add-* family does port cleanly (glob-activated by path).
	CursorPort bool

	// CursorGlobs is the `globs:` value for the Cursor rules file. Used
	// only when CursorPort is true. Typically mirrors Paths.
	CursorGlobs []string
}

// skillCatalog returns the ordered list of skills to emit. Order is not
// functionally significant but kept stable for deterministic generation.
// Frontmatter descriptions are tight (≤200 chars) to keep the combined
// skill-listing block comfortably under the ~1,536-char per-skill cap
// across all 17 skills.
func skillCatalog() []skillDef {
	javaPaths := []string{"**/*.java"}
	aiPaths := []string{"AIAgent/**/*.java", "AIAgent/**/*.md"}
	sqlPaths := []string{"**/*.java", "**/db/migration/*.sql"}

	return []skillDef{
		// ─── Review (Claude-primary, ported to Codex/Copilot) ────────────
		{
			Name:         "review",
			Description:  "Review code changes for quality, modern Java idioms, architecture boundaries, security, and test coverage against project standards.",
			ArgumentHint: "[file-or-branch]",
			Paths:        javaPaths,
			BodyTmpl:     "skills/review.body.md.tmpl",
			Invocable:    true,
			CursorPort:   false, // review is workflow-invocation; Cursor rules can't auto-invoke
		},
		{
			Name:           "review-performance",
			Description:    "Review datastore access patterns for N+1 queries, unbounded scans, offset pagination, and missing indexes. Manual counterpart to the performance-reviewer subagent.",
			ArgumentHint:   "[file-or-branch]",
			Paths:          sqlPaths,
			BodyTmpl:       "skills/review-performance.body.md.tmpl",
			RequiredModule: "anyDatastore",
			Invocable:      true,
			CursorPort:     false,
		},
		{
			Name:           "review-prompts",
			Description:    "Review AI agent code — system prompts, guardrails, tools, knowledge entries — for injection resilience, role clarity, and default-deny guardrail coverage.",
			ArgumentHint:   "[file-or-branch]",
			Paths:          aiPaths,
			BodyTmpl:       "skills/review-prompts.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:      true,
			CursorPort:     false,
		},

		// ─── Git workflow (Claude/Codex/Copilot only; Cursor has no analogue) ─
		{
			Name:         "pr",
			Description:  "Generate a pull request title and description from branch changes: modules affected, breaking changes, test coverage notes.",
			ArgumentHint: "[base-branch]",
			BodyTmpl:     "skills/pr.body.md.tmpl",
			Invocable:    true,
			CursorPort:   false,
		},
		{
			Name:         "commit",
			Description:  "Generate a Conventional Commits message from staged changes: type, scope, summary, and body.",
			ArgumentHint: "[--amend]",
			BodyTmpl:     "skills/commit.body.md.tmpl",
			Invocable:    true,
			CursorPort:   false,
		},

		// ─── Add-production-code workflows (always available) ─────────────
		{
			Name:         "add-entity",
			Description:  "Add a new domain entity. Creates immutable entity, request DTO, datastore record, repository wiring, and tests following project conventions.",
			ArgumentHint: "[entity-name]",
			Paths:        javaPaths,
			BodyTmpl:     "skills/add-entity.body.md.tmpl",
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"Model/**/*.java", "SQLDatastore/**/*.java", "NoSQLDatastore/**/*.java"},
		},
		{
			Name:         "add-endpoint",
			Description:  "Add a new REST endpoint. Creates controller method, request/response DTOs, service logic, validation, and tests following API conventions.",
			ArgumentHint: "[endpoint-path]",
			Paths:        javaPaths,
			BodyTmpl:     "skills/add-endpoint.body.md.tmpl",
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"API/**/*.java"},
		},
		{
			Name:         "add-service",
			Description:  "Add a new business logic service. Creates the service class with circuit breaker, constructor injection, immutable request/response types, and unit tests.",
			ArgumentHint: "[service-name]",
			Paths:        javaPaths,
			BodyTmpl:     "skills/add-service.body.md.tmpl",
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"Shared/**/*.java"},
		},
		{
			Name:         "add-migration",
			Description:  "Add a new Flyway database migration. Generates a versioned SQL file with project conventions: no foreign keys, indexes for foreign-key columns, backward compatibility.",
			ArgumentHint: "[migration-summary]",
			Paths:        []string{"SQLDatastore/**/db/migration/*.sql"},
			BodyTmpl:     "skills/add-migration.body.md.tmpl",
			RequiredModule: "SQLDatastore",
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"SQLDatastore/**/*.sql"},
		},
		{
			Name:         "add-repository-method",
			Description:  "Add a custom repository query. Derived or @Query method with keyset pagination, bulk I/O (findByIds), and project performance rules (§5.5).",
			ArgumentHint: "[method-name]",
			Paths:        javaPaths,
			BodyTmpl:     "skills/add-repository-method.body.md.tmpl",
			RequiredModule: "anyDatastore",
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"SQLDatastore/**/*.java", "NoSQLDatastore/**/*.java"},
		},
		{
			Name:         "add-event",
			Description:  "Add a new event type for async processing. Creates event class in Events module, publisher wiring, and consumer/listener in EventConsumer.",
			ArgumentHint: "[event-name]",
			Paths:        javaPaths,
			BodyTmpl:     "skills/add-event.body.md.tmpl",
			RequiredModule: "Events",
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"Events/**/*.java", "EventConsumer/**/*.java"},
		},
		{
			Name:         "add-job",
			Description:  "Add a new background job. JobRunr JobRequest + handler in Worker module with retry config, testing, and observability.",
			ArgumentHint: "[job-name]",
			Paths:        javaPaths,
			BodyTmpl:     "skills/add-job.body.md.tmpl",
			RequiredModule: "Jobs",
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"Jobs/**/*.java", "Worker/**/*.java"},
		},

		// ─── Add-test (new skill, universal) ──────────────────────────────
		{
			Name:         "add-test",
			Description:  "Add tests for existing or new code. Picks unit vs Testcontainers integration based on dependencies; applies project conventions (MockitoExtension, no field injection, Java 25 gotchas).",
			ArgumentHint: "[target-file]",
			Paths:        []string{"**/src/**/*.java"},
			BodyTmpl:     "skills/add-test.body.md.tmpl",
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"**/src/test/**/*.java", "**/src/main/**/*.java"},
		},

		// ─── Add-AI-agent-code workflows (AIAgent-gated) ─────────────────
		{
			Name:         "add-tool",
			Description:  "Expose a new AI-agent @Tool method. Description clarity, parameter bounding, scope enforcement, and auth gating.",
			ArgumentHint: "[tool-name]",
			Paths:        aiPaths,
			BodyTmpl:     "skills/add-tool.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"AIAgent/**/*.java"},
		},
		{
			Name:         "add-guardrail-rule",
			Description:  "Add a classifier guardrail rule for AI-agent input. Separate-LLM classifier, ALLOW/BLOCK with examples, default-deny, testing.",
			ArgumentHint: "[rule-name]",
			Paths:        aiPaths,
			BodyTmpl:     "skills/add-guardrail-rule.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"AIAgent/**/guardrail/**/*.java"},
		},
		{
			Name:         "add-knowledge-entry",
			Description:  "Add a keyword-matched knowledge-base entry for the AI agent. Token-free FAQ answers before falling back to the LLM.",
			ArgumentHint: "[topic]",
			Paths:        aiPaths,
			BodyTmpl:     "skills/add-knowledge-entry.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"AIAgent/**/knowledge/**/*"},
		},
		{
			Name:         "add-retriever",
			Description:  "Add a custom DocumentRetriever to the AIAgent RAG path (re-ranking, hybrid, multi-source). Composition order is a security invariant — tenant filter innermost, fencing outermost.",
			ArgumentHint: "[retriever-name]",
			Paths:        aiPaths,
			BodyTmpl:     "skills/add-retriever.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"AIAgent/**/knowledge/**/*.java", "AIAgent/**/agent/PrimaryAgent.java"},
		},
		{
			Name:         "add-streaming-endpoint",
			Description:  "Add an SSE streaming endpoint with the canonical 5-part lifecycle (rate-limit, per-caller cap, BOLA-safe ownership, named listener with paired unsubscribe in all 3 emitter hooks).",
			ArgumentHint: "[endpoint-name]",
			Paths:        aiPaths,
			BodyTmpl:     "skills/add-streaming-endpoint.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"AIAgent/**/protocol/**/*.java", "AIAgent/**/task/**/*.java"},
		},
		{
			Name:         "add-agent-variant",
			Description:  "Add a new specialist agent (compliance, summarization, translation, etc.) alongside PrimaryAgent / SpecialistAgent. 4-file pattern: @Qualifier ChatModel + agent class + @Tool wrapper + PrimaryAgent injection. Avoids the @ConditionalOnBean(ChatModel.class) silent-miss trap.",
			ArgumentHint: "[agent-name]",
			Paths:        aiPaths,
			BodyTmpl:     "skills/add-agent-variant.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"AIAgent/**/agent/**/*.java", "AIAgent/**/config/ChatClientConfig.java"},
		},
		{
			Name:         "extend-rag-ingestion",
			Description:  "Extend the RAG ingestion pipeline: new HTTP endpoints, scheduled / event-driven paths, custom metadata fields with indexing. Preserves the 5 ingestion invariants (DocumentIngestionService routing, CallerContext on non-request threads, metadata allow-list / x-* namespace, indexing for filterable fields, rate-limit + scope gate).",
			ArgumentHint: "[ingestion-name]",
			Paths:        aiPaths,
			BodyTmpl:     "skills/extend-rag-ingestion.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"AIAgent/**/knowledge/**/*.java", "AIAgent/**/protocol/Ingestion*.java", "AIAgent/**/db/vector-migration/*.sql"},
		},
		{
			Name:         "add-a2a-skill",
			Description:  "Expose an Agent-to-Agent (A2A) skill endpoint: JSON-RPC handler, async task manager wiring, A2AController registration.",
			ArgumentHint: "[skill-name]",
			Paths:        aiPaths,
			BodyTmpl:     "skills/add-a2a-skill.body.md.tmpl",
			RequiredModule: config.ModuleAIAgent,
			Invocable:    true,
			CursorPort:   true,
			CursorGlobs:  []string{"AIAgent/**/protocol/**/*.java", "AIAgent/**/task/**/*.java"},
		},
	}
}

// shouldEmit decides whether a given skill applies to the selected config.
func (s skillDef) shouldEmit(cfg *config.ProjectConfig) bool {
	switch s.RequiredModule {
	case "":
		return true
	case "anyDatastore":
		return cfg.HasAnyDatastore()
	default:
		return cfg.HasModule(s.RequiredModule)
	}
}

// generateSkills emits the full skill set to every selected AI agent's
// native skill location, plus Cursor rules where applicable. Called from
// the main Generate() flow after the review artifacts.
func (g *Generator) generateSkills() error {
	if !g.config.ReviewEnabled() {
		// Skills share kill-switch semantics with review artifacts. If the
		// user disabled review, they likely don't want discovery clutter.
		return nil
	}

	for _, s := range skillCatalog() {
		if !s.shouldEmit(g.config) {
			continue
		}

		// Render the canonical body once per skill. All three native-skill
		// tools consume the identical body; only the frontmatter wrapper
		// differs.
		body, err := g.renderTemplate(s.BodyTmpl)
		if err != nil {
			return fmt.Errorf("render skill body %s: %w", s.BodyTmpl, err)
		}
		body = strings.TrimSpace(body) + "\n"

		// Claude
		if g.config.HasAIAgent("claude") {
			out := claudeSkillFrontmatter(s) + body
			path := fmt.Sprintf(".claude/skills/%s/SKILL.md", s.Name)
			if err := g.writeFile(joinOut(g, path), out); err != nil {
				return fmt.Errorf("write claude skill %s: %w", s.Name, err)
			}
		}

		// Codex
		if g.config.HasAIAgent("codex") {
			out := codexSkillFrontmatter(s) + body
			path := fmt.Sprintf(".agents/skills/%s/SKILL.md", s.Name)
			if err := g.writeFile(joinOut(g, path), out); err != nil {
				return fmt.Errorf("write codex skill %s: %w", s.Name, err)
			}
		}

		// Copilot
		if g.config.HasAIAgent("copilot") {
			out := copilotSkillFrontmatter(s) + body
			path := fmt.Sprintf(".github/skills/%s/SKILL.md", s.Name)
			if err := g.writeFile(joinOut(g, path), out); err != nil {
				return fmt.Errorf("write copilot skill %s: %w", s.Name, err)
			}
		}

		// Cursor — passive glob-activated rules. Only port the skills that
		// have a natural glob scope; workflow-invocation skills like /commit
		// have no meaningful passive form.
		if g.config.HasAIAgent("cursor") && s.CursorPort {
			out := cursorRuleFrontmatter(s) + body
			path := fmt.Sprintf(".cursor/rules/%s.mdc", s.Name)
			if err := g.writeFile(joinOut(g, path), out); err != nil {
				return fmt.Errorf("write cursor rule %s: %w", s.Name, err)
			}
		}
	}
	return nil
}

// Frontmatter builders. Each tool wants slightly different fields; extract
// into helpers so the emitter loop stays readable.

func claudeSkillFrontmatter(s skillDef) string {
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	if s.Invocable {
		b.WriteString("user-invocable: true\n")
	}
	b.WriteString("allowed-tools: Read Grep Glob Bash Edit Write\n")
	if s.ArgumentHint != "" {
		fmt.Fprintf(&b, "argument-hint: %q\n", s.ArgumentHint)
	}
	if len(s.Paths) > 0 {
		b.WriteString("paths:\n")
		for _, p := range s.Paths {
			fmt.Fprintf(&b, "  - %q\n", p)
		}
	}
	b.WriteString("---\n\n")
	return b.String()
}

func codexSkillFrontmatter(s skillDef) string {
	// Codex follows Anthropic's shape closely; `allow_implicit_invocation`
	// controls autonomous matching (analog to Claude's invocable flag).
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	if s.Invocable {
		b.WriteString("allow_implicit_invocation: true\n")
	}
	if s.ArgumentHint != "" {
		fmt.Fprintf(&b, "argument-hint: %q\n", s.ArgumentHint)
	}
	b.WriteString("---\n\n")
	return b.String()
}

func copilotSkillFrontmatter(s skillDef) string {
	// Copilot is strict: the `name` field MUST match the directory name.
	// We enforce that by construction — both use s.Name.
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "name: %s\n", s.Name)
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	b.WriteString("---\n\n")
	return b.String()
}

func cursorRuleFrontmatter(s skillDef) string {
	// Cursor rules with glob-based activation. `alwaysApply: false` means
	// the rule attaches only when a matching file is opened/referenced —
	// the closest analogue to skill auto-invocation Cursor offers.
	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "description: %s\n", s.Description)
	if len(s.CursorGlobs) > 0 {
		b.WriteString("globs:\n")
		for _, g := range s.CursorGlobs {
			fmt.Fprintf(&b, "  - %q\n", g)
		}
	}
	b.WriteString("alwaysApply: false\n")
	b.WriteString("---\n\n")
	return b.String()
}

// joinOut resolves a project-relative path against the generator's outDir.
// Kept private to this file so the skill emitter doesn't have to reach
// into the Generator type's unexported internals.
func joinOut(g *Generator, rel string) string {
	return filepath.Join(g.outDir, rel)
}
