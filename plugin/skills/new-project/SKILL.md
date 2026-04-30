---
name: new-project
description: Start a new Java/Spring Boot project from natural-language requirements. Recommends modules, confirms with the user, then generates via the Trabuco CLI. Use when the user wants to create a new service and describes what it should do.
user-invocable: true
allowed-tools: [mcp__trabuco__suggest_architecture, mcp__trabuco__list_modules, mcp__trabuco__list_providers, mcp__trabuco__check_docker, mcp__trabuco__init_project, mcp__trabuco__get_version]
argument-hint: "[short requirement description]"
---

# Start a new Trabuco project

Guide the user from a natural-language requirement to a generated, opinionated Java/Spring Boot project.

## Flow

1. **Preflight**: call `mcp__trabuco__get_version` to confirm the CLI is reachable. If it errors, tell the user to install trabuco (https://github.com/arianlopezc/Trabuco/releases) and stop. Call `mcp__trabuco__check_docker` — Trabuco needs Docker for Testcontainers. If Docker isn't running, warn but continue (user can start it later).

2. **Clarify intent**: if the argument is missing or vague ("make me a service"), ask 2–3 tight questions:
   - What does it do? (one sentence)
   - Relational or document store? If unsure, ask what entities they'll model.
   - Sync (REST-only) or async (events, jobs, both)?
   - AI agent involved? (enables the AIAgent module with A2A + MCP + guardrails)

3. **Recommend**: call `mcp__trabuco__suggest_architecture` with the requirements. It returns matched patterns + a `recommendedConfig`. Present the recommendation *with rationale grounded in the returned data* — never invent modules that weren't in the response. If confidence is low or the top two patterns score close, surface the ambiguity and ask the user to pick.

4. **Confirm parameters**: before generating, confirm with the user:
   - project name (kebab-case, e.g., `order-service`)
   - group ID (`com.company.project`)
   - Java version (21 default; 25 if they want latest LTS)
   - Modules (from the recommendation; let them add/drop)
   - Database / broker choices
   - AI agents to integrate (`claude`, `cursor`, `codex`, `copilot` — default all four since the plugin is Claude-focused but the user may pair-tool)

5. **Generate**: call `mcp__trabuco__init_project` with the confirmed parameters. Include `skip_build: true` on the first pass so the user sees the tree before Maven runs. Report the output directory when done.

6. **Next steps**: once generated, tell the user:
   - `cd <project>` and run `mvn clean install` to verify build
   - If AIAgent module was included, `mvn spring-boot:run -pl AIAgent` starts the agent with MCP + A2A surfaces
   - **If API or AIAgent was selected**: OIDC Resource Server scaffolding has been generated, dormant by default. To activate auth, set `trabuco.auth.enabled=true` and configure `OIDC_ISSUER_URI` to your IdP discovery endpoint (Keycloak / Auth0 / Okta / Cognito / generic). See `docs/auth.md` in the generated project for per-provider recipes. Until then, every endpoint is open via the permit-all `SecurityFilterChain` — fine for local dev, do not deploy to prod that way.
   - Skills like `/add-entity`, `/add-endpoint`, `/add-test` are now available *inside the generated project's .claude/skills/*
   - Ask if they want you to pre-add an entity or endpoint (delegate to `/add-module` or direct MCP tool calls)

## Rules

- **Never invent modules**. Only suggest what `suggest_architecture` returned or what `list_modules` confirms exists.
- **Never skip the confirmation step**. Project generation is not reversible from chat — confirm explicitly before calling `init_project`.
- **Respect limitations**. Read `trabuco://limitations` resource before suggesting something that might be out of scope (e.g., GraphQL, gRPC, frontend UI).
- **Default to opinions**. Trabuco is opinionated — keyset pagination, no FK constraints, constructor injection, Immutables. Don't suggest workarounds unless the user asks.

## When NOT to use this skill

- User already has a project and wants to extend it → use `/trabuco:add-module` or `/trabuco:extend`.
- User wants multi-service / microservices design → use `/trabuco:design-system` instead (calls `design_system`, not `init_project`).
