# Trabuco Security Audit — AI Surface Domain

Documentation, prompts, skills, MCP tool descriptions, and agent guidance Trabuco emits into generated projects. The AI surface must not normalize insecure defaults or teach unsafe patterns.

This file is the **detail reference** for the
`trabuco-security-audit-ai-surface` specialist subagent. The orchestrator
loads this file, the master checklist (`./checklist.md`), and the
specialist's prompt, then dispatches the subagent against the project's
source tree.

**How to read each entry:**

- **`<F-...>` heading** — the stable check ID. Findings reference this ID.
- **Severity floor** — the orchestrator may not downgrade below this
  unless an explicit `[suppress: <reason>]` justification is recorded.
- **Taxonomy** — OWASP / API Security / LLM / ASVS / CWE / Trabuco-specific
  cross-references.
- **Where to look** — the file paths and line ranges in a Trabuco-generated
  project where this issue typically lands.
- **Evidence pattern** — the antipattern to grep for. Specialist subagents
  use this as the primary detection signal.
- **Why it matters** — concise explanation of the threat model.
- **Suggested fix** — the recommended remediation. Specialists include this
  in their finding records so operators don't have to think from scratch.

**Total checks in this domain: 17**
(0 Critical,
 5 High,
 6 Medium,
 6 Low,
 0 Informational)

---

## F-AI-01 — Generated `CLAUDE.md` claims project does NOT include auth, contradicting dormant-by-default scaffolding

**Severity floor:** High
**Taxonomy:** OWASP-A05, ASVS-V14-1, TRABUCO-001 + AI-SURFACE-01 (proposed)

### Where to look

`/tmp/trabuco-secaudit/aiagent-pgvector-rag/CLAUDE.md:220` (also `/tmp/trabuco-secaudit/full-fat-all-modules/CLAUDE.md:232`); template at `/Users/arianlc/Documents/Work/Trabuco/templates/docs/CLAUDE.md.tmpl` (Boundaries section)

### Why it matters

Since Trabuco 1.11, every project with API or AIAgent ships a complete OIDC Resource Server scaffold (dual `SecurityFilterChain`, `JwtAuthenticationConverter`, RFC 7807 handlers, `IdentityClaims`, `AuthorityScope`, etc.) — dormant but present. The flagship file Claude Code reads first asserts the opposite. A coding agent following this guidance will proactively bolt on a *new* `SecurityConfig`, duplicating beans and creating bean-conflict / activation surprises (a form of CWE-1188 — insecure default + misleading guidance). It also tells the user "no auth" so they may not realize the permit-all chain is still active in dev and ship without enabling JWT.

---

## F-AI-02 — Generated `extending-the-project.md` walks users through ADDING Spring Security from scratch, while project already ships a JWT chain

**Severity floor:** High
**Taxonomy:** OWASP-A05, OWASP-A07, ASVS-V14-1, TRABUCO-001 + AI-SURFACE-02 (proposed)

### Where to look

`/tmp/trabuco-secaudit/aiagent-pgvector-rag/.ai/prompts/extending-the-project.md:5-25`; templates under `/Users/arianlc/Documents/Work/Trabuco/templates/ai/prompts/extending-the-project.md.tmpl`

### Why it matters

The project already contains `api/config/security/SecurityConfig.java`, the OAuth2 resource-server starter, RFC 7807 ProblemDetail handlers, `JwtAuthenticationConverter`, and an integration test suite (`SecurityIntegrationTest`, `AuthEndToEndTest`). Following these instructions will introduce a *second* `SecurityConfig` bean, conflict with the dormant chain, and silently mask the dormant-by-default safety story with hand-rolled rules. The plugin-mode `trabuco-architect` agent explicitly says *"Don't tell users to add Spring Security manually — it's already wired"* — but the project-mode prompt does exactly that. This is the most dangerous primer in the AI surface.

---

## F-AI-03 — Generated `add-a2a-skill` / `add-knowledge-entry` / `add-guardrail-rule` curl examples hardcode the well-known seeded API keys

**Severity floor:** High
**Taxonomy:** OWASP-A07, ASVS-V2-2, CWE-798, LLM-EXT-02 + AI-SURFACE-03 (proposed)

### Where to look

- `/tmp/trabuco-secaudit/aiagent-pgvector-rag/.ai/prompts/add-a2a-skill.md:38-49, 70-80`

### Why it matters

The strings `public-read-key` and `partner-secret-key` are not placeholders — they are the literal seed values hardcoded in `ApiKeyAuthFilter.java.tmpl` and bound in `application.yml.tmpl`. The AI surface uses them in 5+ curl examples without any "rotate before deploy" / "do not commit" warning. A coding agent reading these will (a) tell the user to copy these as-is, (b) carry them into committed integration tests, (c) reuse them in deployment scripts. Anyone on the public internet who reads the Trabuco docs can authenticate as `partner` to a freshly-deployed agent service while `app.aiagent.api-key.enabled=true` (default).

---

## F-AI-04 — Plugin docs and skills repeatedly tell users "every endpoint is open … fine for local dev, do not deploy to prod that way" without enforcement

**Severity floor:** High
**Taxonomy:** OWASP-A05, ASVS-V4-1, CWE-1188, TRABUCO-001

### Where to look

- `/Users/arianlc/Documents/Work/Trabuco/plugin/skills/new-project/SKILL.md:38`

### Why it matters

This is the recurring framing across the AI surface. The phrase "fine for local dev" is a known anti-pattern primer — it tells coding agents (and Claude downstream) that the permit-all default is OK as long as the user remembers to flip a flag before prod. There is no machine-checkable guard, no `/trabuco:doctor` rule, no SessionStart warning when a generated project is cloned into a prod-shaped directory. The same primer appears in the `trabuco-architect` agent ("auth comes free with API/AIAgent … the user activates it at runtime"), the `trabuco-ai-agent-expert` agent, the `add-module` skill, the `design-system` skill, and the MCP `init_project` tool boundaries field — each reinforcing that production deployment with auth-off is something the *user* must remember.

---

## F-AI-05 — `app.aiagent.api-key.enabled=true` (default) framed as "backward compatibility" without warning seeded keys are well-known

**Severity floor:** High
**Taxonomy:** OWASP-A07, ASVS-V2-2, CWE-798

### Where to look

- `/Users/arianlc/Documents/Work/Trabuco/plugin/skills/add-module/SKILL.md:38`

### Why it matters

Every AI-surface mention of `app.aiagent.api-key.enabled` describes "default on" / "default true" as if backward compatibility is a neutral or positive property. None of the mentions warn that the seeded `public-read-key` / `partner-secret-key` are committed to the template, end up in every generated project, and are documented verbatim in the project's own `.ai/prompts/`. A coding agent will follow the AI surface's framing — "this is the legacy path; it's on for compat" — and not surface the seed-key-rotation requirement. The auth.md matrix even shows `(false, false) → "Don't run this in prod"` but does NOT call out `(true, anything) with default seeds → instant compromise`.

---

## F-AI-06 — Stop-hook block messages teach the agent the kill switches and the suppression directive

**Severity floor:** Medium
**Taxonomy:** OWASP-A05, ASVS-V14-1, AI-SURFACE-04 (proposed)

### Where to look

- `/Users/arianlc/Documents/Work/Trabuco/templates/claude/hooks/require-review.sh.tmpl:114, 146, 151` (and emitted to `/tmp/trabuco-secaudit/aiagent-pgvector-rag/.claude/hooks/require-review.sh:114, 146, 151`)

### Why it matters

When the deterministic review check (the only enforcement layer that catches `findAll()`, offset pagination, hardcoded secrets, SQL string concatenation) fails, the *block reason itself* shipped to the agent contains the kill-switch commands. A frustrated user or an over-helpful coding agent reading the block message will be primed to disable the hook rather than fix the violation. The same applies to the `// trabuco-allow:` inline-suppression directive — its existence is documented inside the block reason, before the user has even read the rule that fired. This is a teach-the-bypass anti-pattern: the safety net advertises its own kill switch as the first remediation option.

---

## F-AI-07 — Migration `tests` specialist allows `@MockBean JwtDecoder` and `MockJwtFactory`-driven SecurityContext as documented test patterns

**Severity floor:** Medium
**Taxonomy:** OWASP-A07, ASVS-V2-1, ASVS-V14-1

### Where to look

`/Users/arianlc/Documents/Work/Trabuco/docs/auth.md:286-298`

### Why it matters

`MockJwtFactory.jwt("user-42", "admin")` produces a token that bypasses signature verification (the decoder is mocked). The doc presents this as "Pattern 2" for any `@SpringBootTest`. There is no warning that copy-pasting Pattern 2 into a non-test source set or a misconfigured `@TestConfiguration` that bleeds into production scope re-enables permit-all by accident. The doc also doesn't recommend Pattern 3 only inside `@WithSecurityContext`-scoped helpers — it shows raw `SecurityContextHolder.setAuthentication`, which is dangerous if accidentally invoked from `main`. A downstream coding agent will see this as the canonical "how to authenticate" pattern.

---

## F-AI-08 — Migration assessor classifies hardcoded secrets as "blocker" but config specialist treats env-var conversion as a "backstop"

**Severity floor:** Medium
**Taxonomy:** CWE-798, OWASP-A02, ASVS-V14-1

### Where to look

- `/Users/arianlc/Documents/Work/Trabuco/internal/migration/specialists/assessor/prompt.md:101-104, 167-170`

### Why it matters

The two prompts have inconsistent semantics. Assessor declares `SECRET_IN_SOURCE` a mandatory blocker that the user must address before migration. Config specialist (Phase 9, runs much later) describes itself as a "backstop" — which means the migration may have proceeded for 8 phases with secrets present (in entities, services, controllers) before Phase 9 quietly converts them. Worse, in CLI mode the assessor blocker can be acknowledged via `migrate_decision` and the migration proceeds. The AI surface gives no clear instruction on what `migrate_decision` choice is safe for `SECRET_IN_SOURCE` — there is no enumerated alternative ("rotate now" vs "leave for backstop"), so an LLM-driven orchestrator will likely pick "proceed" and rely on Phase 9.

---

## F-AI-09 — `trabuco-ai-agent-expert` MCP prompt teaches that guardrails save tokens as the *primary* benefit, not security

**Severity floor:** Medium
**Taxonomy:** LLM-01, LLM-05, LLM-06, AI-SURFACE-05 (proposed)

### Where to look

`/Users/arianlc/Documents/Work/Trabuco/internal/mcp/expert_prompts.go:319-324`

### Why it matters

Step 4 frames the *primary* benefit of input guardrails as token-cost reduction. This subtly trains downstream agents to treat guardrails as a performance optimization rather than a security boundary. A coding agent that internalizes "guardrails save tokens" will be willing to disable them when token cost is not a concern (e.g., self-hosted models) or to short-circuit them on cache hits. The plugin-side `trabuco-ai-agent-expert` agent gets this right ("Default deny … Never fall through to ALLOW"), but the MCP-prompt copy of the same agent does not.

---

## F-AI-10 — Generated `add-tool.md` allows-and-trusts LLM-synthesized inputs without mandatory parameter type bounding

**Severity floor:** Medium
**Taxonomy:** LLM-01, LLM-05, LLM-06, ASVS-V5-1

### Where to look

`/tmp/trabuco-secaudit/aiagent-pgvector-rag/.claude/skills/add-tool/SKILL.md:28`

### Why it matters

The phrasing "don't trust" is a soft suggestion, not a hard rule. There is no enumerated minimum (e.g., "every numeric param requires `@Min`/`@Max`; every String requires `@Size`; enums must be Java enums, not String"). Compare with the structural enforcement that exists for other Trabuco patterns ("constructor injection only," "keyset pagination — banned offset"). For `@Tool` methods exposed to MCP and A2A, this should be the same level of enforcement. The prompt-reviewer agent does check for this, but it runs *after* code is generated, not as a hard generation rule.

---

## F-AI-11 — `trabuco-migration-deployment` specialist forbids security additions while migrating CI ("do not add security scanning")

**Severity floor:** Medium
**Taxonomy:** OWASP-A05, MAVEN-004, CI-008, AI-SURFACE-06 (proposed)

### Where to look

`/Users/arianlc/Documents/Work/Trabuco/internal/migration/specialists/prompts/deployment.md:46-56`

### Why it matters

"Don't add security scanning" is principled (no out-of-scope) but produces a hostile primer for downstream agents. After migration, a fresh Trabuco project that passed through Phase 10 will have an *adapted* legacy CI that has no SCA, no SAST, no secret scanning, no SBOM generation — exactly what the checklist (`MAVEN-004`, `CI-013`) requires for ASVS L1/L2. The AI surface tells specialists never to fix this, even when adapting CI is the natural moment. A user reading this prompt is primed to think "we explicitly chose not to add security checks" and not surface the gap to the user as a follow-up todo.

---

## F-AI-12 — `trabuco-migration-orchestrator` allows the user to type "approve" without per-file diff inspection

**Severity floor:** Low
**Taxonomy:** OWASP-A05, ASVS-V14-1

### Where to look

`/Users/arianlc/Documents/Work/Trabuco/plugin/agents/trabuco-migration-orchestrator.md:46-50`

### Why it matters

"per-module diff summary (not file-by-file)" — for a sensitive phase like Phase 4 (Shared, where auth utilities live), this means the user approves an entire module's worth of changes from a summary. An attacker scenario isn't required; even an honest LLM hallucination that adds a method on `JwtClaimsExtractor` would be approved at module granularity without surfacing. The `--per-aggregate` flag exists but is mentioned only in skill rules, not in the orchestrator's gate copy, so an LLM driving the orchestrator won't propose it.

---

## F-AI-13 — Plugin SessionStart hook emits "MCP tools available" without warning the user MCP tools accept arbitrary file paths

**Severity floor:** Low
**Taxonomy:** LLM-EXT-02, OWASP-A01, CWE-22

### Where to look

`/Users/arianlc/Documents/Work/Trabuco/plugin/hooks/session-start.sh:58`; `/Users/arianlc/Documents/Work/Trabuco/internal/mcp/tools.go:308-310, 671-674` (the `path` parameter)

### Why it matters

Several MCP tools (`add_module`, `run_doctor`, `get_project_info`, `migrate_*`) accept an arbitrary `path` parameter and operate on it (read POMs, write files via `add_module`, run `git reset --hard` via `migrate_rollback`). The tool descriptions don't bound the path to the working directory or warn the agent. A coding agent following user prompts ("add a module to ../../some-other-project") will happily mutate sibling repos. The SessionStart hook tells the user "MCP tools are available" without summarizing the surface area each one mutates. This is also relevant to LLM-EXT-02 (MCP server insecure exposure).

---

## F-AI-14 — Migration-mode parent POM ships with enforcement deferred for 11 phases and the orchestrator does not prompt for spot-checks

**Severity floor:** Low
**Taxonomy:** ASVS-V14-1, MAVEN-001, CWE-1188

### Where to look

- `/Users/arianlc/Documents/Work/Trabuco/plugin/agents/trabuco-migration-orchestrator.md:54-55` (referenced)

### Why it matters

During phases 1-11 (potentially hours/days of agent-driven work), Maven Enforcer / Spotless / ArchUnit / Jacoco are all skipped. ArchUnit boundary tests are tagged-and-excluded so module dependencies are not checked. A specialist that hallucinates a cross-module import (e.g., `Worker` importing from `API`) will not be caught until Phase 12. The AI surface acknowledges this is intentional ("safe over fast") but does not prompt the orchestrator or user to spot-check the dependency graph at any intermediate phase boundary. "for now" / "until then" framing again primes a "we'll fix it later" mentality.

---

## F-AI-15 — `trabuco_expert` MCP prompt's POST-GENERATION STEPS omit auth activation

**Severity floor:** Low
**Taxonomy:** OWASP-A05, TRABUCO-001, AI-SURFACE-07 (proposed)

### Where to look

`/Users/arianlc/Documents/Work/Trabuco/internal/mcp/expert_prompts.go:79-86`

### Why it matters

The numbered list of post-generation actions does not contain "configure OIDC_ISSUER_URI" or "set trabuco.auth.enabled=true before deployment." The earlier section in the same prompt mentions auth scaffolding generated dormant, but the actionable checklist a coding agent will copy into a "next steps" message to the user does not. Compare with `init_project`'s `next_steps` (`tools.go:238-251`) which also omits auth activation from the JSON returned to the agent. Combined with F-AI-04, this is the place where auth activation falls between cracks.

---

## F-AI-16 — `init_project` MCP tool description claims auth scaffolding is dormant "by default" without surfacing the seeded API-key risk

**Severity floor:** Low
**Taxonomy:** OWASP-A07, CWE-798

### Where to look

`/Users/arianlc/Documents/Work/Trabuco/internal/mcp/tools.go:55-57`

### Why it matters

The tool description tells the calling agent that auth is "dormant by default" — a comforting framing — but does not mention that AIAgent simultaneously ships the legacy `ApiKeyAuthFilter` *active* by default with hardcoded seed keys. An agent will read this description, conclude "auth is off until I flip it on," and miss that an attacker can already authenticate via API key on a freshly-deployed AIAgent service.

---

## F-AI-17 — Generated test patterns recommend `@MockBean` of JwtDecoder with no warning that `MockJwtFactory` issues admin scopes trivially

**Severity floor:** Low
**Taxonomy:** ASVS-V14-1, OWASP-A07

### Where to look

`/Users/arianlc/Documents/Work/Trabuco/docs/auth.md:280-298`

### Why it matters

`MockJwtFactory.jwt("user-42", "admin")` accepts any scope string trivially. A test scaffold that mocks the `JwtDecoder` is appropriate, but the convenience of "give me an admin token in one line" can encourage agents to use this pattern in non-test code (e.g., dev endpoint shortcuts, "happy-path" demo controllers). There is no scope safelist or warning that scopes passed to the factory are inherently un-trusted.

---

