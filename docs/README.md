# Trabuco documentation

Browse by topic. Most users will start at [`manual.md`](manual.md) (full
CLI reference) or [`migration-guide.md`](migration-guide.md) (migrating
an existing Spring Boot project).

## Reference

- **[`manual.md`](manual.md)** — full CLI reference: installation, the
  Claude Code plugin, quick start, managing existing projects (doctor /
  add / sync), MCP server configuration for every supported agent, the
  generated project anatomy, all 8 modules in detail, code-quality
  enforcement, CI/CD, observability, every flag, the tech stack, local
  development.
- **[`migration-guide.md`](migration-guide.md)** — migrating an existing
  Spring Boot 2.x or 3.x repo into a Trabuco-shaped multi-module project
  via the orchestrated 14-phase flow. Phases, gates, decisions, rollback,
  troubleshooting, design principles.

## Design context (loaded by the Claude Code plugin, useful for everyone)

These five files live under [`../plugin/docs/`](../plugin/docs/) because
the Trabuco plugin loads them into the assistant's context to ground
recommendations. They're written for the AI but read perfectly well as
human documentation — read them when you want to understand *why*
Trabuco makes the choices it does, or when you're deciding whether
Trabuco is the right tool for your problem.

- **[Philosophy](../plugin/docs/philosophy.md)** — design principles
  behind Trabuco's hard conventions (Maven-only, constructor injection,
  no FK constraints, Immutables, virtual threads, etc.).
- **[Module catalog interpretation](../plugin/docs/module-catalog.md)**
  — which of the 8 modules to reach for, idiomatic combinations, code
  smells in module selection.
- **[Pattern recipes](../plugin/docs/pattern-recipes.md)** — playbooks
  for common requirements (CRUD REST, event-driven, AI agent, batch
  processor, gateway, multi-service workspace).
- **[When NOT to use Trabuco](../plugin/docs/when-not-to-use.md)** —
  hard stops, strong warnings, and red flags. Honest framing for when
  Trabuco is the wrong tool.
- **[Limitations](../plugin/docs/limitations.md)** — what Trabuco
  doesn't generate (frontend UI, identity-provider side, vector RAG, K8s)
  and how to plan around the gaps. Auth scaffolding (resource-server side)
  ships built-in for API/AIAgent — see [`auth.md`](auth.md).
- **[Authentication](auth.md)** — OIDC Resource Server scaffolding
  generated for API/AIAgent: dual SecurityFilterChain pattern, dormant
  by default, activated via `trabuco.auth.enabled=true`. Per-provider
  recipes for Keycloak, Auth0, Okta, AWS Cognito, generic OIDC.

## Other entry points

- **[Repository README](../README.md)** — pitch, install one-liner,
  quickstart.
- **[Plugin README](../plugin/README.md)** — Trabuco as a Claude Code
  plugin (skills, subagents, hooks, MCP server).
- **[npm package README](../npm/README.md)** — the `trabuco-mcp` npm
  wrapper for MCP-only setups (no plugin).
