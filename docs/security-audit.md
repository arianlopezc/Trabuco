# Security Audit

Trabuco ships a structured security-audit toolchain that any coding
agent can load and run against a generated project. The audit walks a
canonical 173-check checklist across five domains (auth, AI surface,
AIAgent runtime + Java platform, data + events, web + infra), produces
a severity-sorted findings report, and gives an explicit pass/fail
verdict.

This document is for **plugin developers and operators**. It covers
what the audit is, how to invoke it, how findings are emitted and
suppressed, how to extend the checklist for project-specific rules,
and where the checklist source-of-truth lives.

## Two surfaces, one source of truth

The same checklist drives two distinct invocation surfaces:

| Surface | Audience | Entry point |
| ------- | -------- | ----------- |
| **Plugin** (Claude Code marketplace) | Anyone who installs the Trabuco plugin and points Claude at a Trabuco-generated project | `/audit` slash command, backed by `plugin/skills/audit/SKILL.md` |
| **Generated project** | Any AI coding agent (Claude / Cursor / Copilot / Codex CLI) that the operator wired into their project | Per-format guidance file in the project root: `.claude/skills/audit/SKILL.md`, `.cursor/rules/security-audit.mdc`, `.github/instructions/security-audit.instructions.md`, `.codex/security-audit.md` |

Source of truth: `plugin/skills/audit/checklist.md` (master) plus five
domain detail files in the same directory. `trabuco init` and
`trabuco sync` materialize byte-equivalent copies into the generated
project at `.ai/security-audit/checklist*.md`. Both surfaces consume
the same data.

## How to run the audit

### Inside Claude Code (plugin or per-project)

```
/audit
```

The skill hands off to the `trabuco-security-audit-orchestrator`
subagent, which dispatches five domain specialists in parallel via the
Task tool, merges their findings, and writes
`.ai/security-audit/findings.md`. Returns a one-line verdict
(`PASS — no Critical or High findings` or
`FAIL — N Critical, M High`).

Optional arguments:

- `--scope=changed` (default in PR context) — walk only files changed
  since the merge base. Faster; appropriate for CI gates.
- `--scope=all` — full repository walk. Use for periodic deep audits.
- `--domain=auth|ai-surface|aiagent-java|data-events|web-infra|all` —
  limit to one domain. Default `all`.

Typical runtime on a full-fat 80k-LOC project: under 2 minutes
(specialists run in parallel; `--scope=changed` drops to 10–30s).

### Inside Cursor / Copilot / Codex

These coding agents lack first-class subagent dispatch, so the audit
runs sequentially as a single agent walking the same five-domain
checklist. The per-format guidance file in each project's root tells
the agent the procedure: load checklists in domain order, search for
each Evidence pattern, honor `// trabuco-allow:` suppressions, write
the findings report.

Typical runtime: 3–8 minutes (sequential, but covers the same 173
checks).

### From CI (deterministic basics only)

The generated `.github/scripts/review-checks.sh` includes 12
regex-based OWASP Top 10 rules tagged `owasp.<id>`. They run on every
PR via the `review-checks` workflow job. These are the deterministic
slice of the audit; the full LLM-driven sweep stays in the `/audit`
workflow.

## Findings report

Output: `.ai/security-audit/findings.md` in the project root.
**Gitignored by default** — operators paste findings into PR
descriptions or evidence trackers, not into git history.

Schema (severity-sorted, diff-stable across re-runs):

```markdown
# Trabuco Security Audit — Findings

**Generated:** <ISO 8601 timestamp>
**Scope:** changed | all
**Verdict:** PASS | FAIL

## Summary
| Severity | Count |

## Findings
### Critical (N)
#### F-AUTH-01 — <title>
**File:** `<path>:<lines>`
**Evidence:** `<snippet>`
**Fix:** <prose>
```

Verdict rule: `PASS` if zero Critical and zero High findings,
otherwise `FAIL`.

## Severity rubric

| Severity | Definition |
| -------- | ---------- |
| Critical | Default-on exploitable vulnerability or data-loss path. Block release. |
| High | Exploitable under common conditions; fix before next deploy. |
| Medium | Latent issue that becomes critical under composition. |
| Low | Hardening opportunity; fix during routine maintenance. |
| Informational | Defensible default with a documented trade-off. |

The orchestrator MUST NOT downgrade a finding below the severity
floor recorded in the checklist without an explicit
`[suppress: <reason>]` justification, which it surfaces in the report.

## Suppression syntax

False positives are recoverable with an inline comment on the
offending source line, or directly above it:

```java
// trabuco-allow: owasp.a02-weak-hash
md5();  // legacy ETag computation, not a security check

// trabuco-allow: all
runMethod();  // disables every rule on this line
```

Both the LLM-driven specialists and the CI deterministic checks honor
the same convention. Each suppression reason is logged in the
findings report so reviewers can see what was waived and why.

## Project-specific extensions

To add rules without forking the canonical checklist, create
`.ai/security-audit/checklist-local.md` in the project root. The
orchestrator reads it alongside the canonical files and applies
local rules with the same weight.

Schema is identical to the bundled checklist files:

```markdown
## LOCAL-001 — Disallow direct usage of `org.example.Legacy*`

**Severity floor:** High
**Domain:** aiagent-java
**Taxonomy:** company-internal

### Where to look

`*/src/main/java/**/*.java`

### Evidence pattern

```java
import org.example.Legacy
```

### Why it matters

`org.example.Legacy*` is deprecated…

### Suggested fix

Replace with `com.company.Modern.*`.
```

`trabuco sync` **never overwrites** `checklist-local.md` — it's
operator-owned. Canonical files (`checklist.md` plus the five
domain files) are Trabuco-owned and refresh on every sync.

## Updating the canonical checklist

The plugin's `plugin/skills/audit/checklist*.md` is the source of
truth. To add or modify a check:

1. Edit the relevant domain file (or `checklist.md` for the master
   index).
2. Bump the rule count summary in the file's header comment.
3. Update the OWASP / API / LLM / ASVS / CWE cross-reference table in
   `checklist.md` if the new check maps to a previously-unmapped class.
4. Mirror the change in `templates/ai/security-audit/<file>.md.tmpl`
   (these are byte-equivalent copies; CI guards against drift).
5. Bump the plugin version in `plugin.json` so consumers can detect
   the update via `trabuco sync`.

## CI integration

The generated `.github/workflows/ci.yml` includes a `review-checks`
job that runs `.github/scripts/review-checks.sh`. The OWASP basics
fail the PR build when a deterministic match fires. Each finding is
emitted as a GitHub Actions annotation
(`::error file=...,line=...,title=[owasp.<id>]::...`) so reviewers
see them inline in the PR diff.

The full LLM-driven `/audit` is **not** wired into CI — it requires
a Claude API key, costs API budget per CI run, and is non-deterministic
across runs. Operators invoke `/audit` deliberately at PR-author time
or pre-release time.

## Future work

- **MCP exposure.** A `mcp__trabuco__security_audit` tool implemented
  in Go would make the deterministic regex layer callable by any
  MCP-aware agent (Claude Desktop, custom integrations, etc.) without
  the plugin or per-project subagents. Deferred — the current
  prompt-driven path closes the loop fast and the CI script already
  ports cleanly to Go when needed.
- **Automated fixture-based integration tests.** Phase 7 of the
  toolchain wires a `test/audit-fixtures/` project with planted
  vulnerabilities and asserts each `owasp.<id>` rule fires in CI. The
  LLM-driven half (the orchestrator + 5 specialists) stays manually
  verified for now — CI has no LLM, and per-CI-run API costs aren't
  worth the marginal value over the deterministic basics.
- **Findings history.** Currently findings overwrite on each run.
  An optional `--history` flag to write `findings-history/<timestamp>.md`
  would give operators a longitudinal view; gitignored by default.

## References

- Canonical checklist (master): `plugin/skills/audit/checklist.md`
- Domain detail: `plugin/skills/audit/checklist-{auth,ai-surface,aiagent-java,data-events,web-infra}.md`
- Plugin orchestrator: `plugin/agents/trabuco-security-audit-orchestrator.md`
- Plugin specialists: `plugin/agents/trabuco-security-audit-{auth,ai-surface,aiagent-java,data-events,web-infra}.md`
- Per-project checklist source: `templates/ai/security-audit/checklist*.md.tmpl`
- Per-project Claude variant: `templates/claude/agents/security-audit-*.md.tmpl`
- Per-project Cursor / Copilot / Codex variants: under their respective `templates/` subtrees
- Generator wiring: `internal/generator/audit.go`
- CI deterministic rules: `templates/github/scripts/review-checks.sh.tmpl` (search for `owasp.`)
