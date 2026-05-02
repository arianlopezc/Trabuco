---
name: add-module
description: Add a Trabuco module to an existing project. Detects current project state, recommends compatible modules, and wires them in. Use when the user has an existing Trabuco project and wants to extend it (add AIAgent, add EventConsumer, add NoSQL, etc.).
user-invocable: true
allowed-tools: [mcp__trabuco__get_project_info, mcp__trabuco__list_modules, mcp__trabuco__add_module, mcp__trabuco__run_doctor]
argument-hint: "[module name]"
---

# Add a module to an existing Trabuco project

## Flow

1. **Detect project**: call `mcp__trabuco__get_project_info` with the current working directory. If it doesn't return a Trabuco project (no `.trabuco.json`), tell the user this skill requires a Trabuco project and suggest `/trabuco:new-project` instead.

2. **Confirm current state**: show the user what modules are currently installed and which provider/broker is in use. This is for their confirmation before additions.

3. **Recommend if no argument**: if the user didn't specify a module, call `mcp__trabuco__list_modules` and recommend based on gaps:
   - Has API but no SQLDatastore? → suggest SQLDatastore
   - Has EventConsumer but no Events contracts module? → suggest Events
   - Has Shared/API but no AIAgent? → suggest AIAgent if they're building user-facing services
   - Etc.

   Present the top 1–3 candidates with rationale.

4. **Check conflicts**: SQLDatastore and NoSQLDatastore are mutually exclusive — warn if the user is attempting a conflict.

5. **Execute**: call `mcp__trabuco__add_module` with the chosen module. Confirm any broker/database sub-choices (e.g., Kafka vs RabbitMQ for EventConsumer).

6. **Post-add verification**: call `mcp__trabuco__run_doctor` to verify the updated project is structurally sound. Report any warnings.

7. **Next steps**: tell the user what the new module brought in — new package structure, new POM dependencies, new generated files to customize, any new `/add-X` skills that became relevant inside the project.

## Rules

- **Never add a module without explicit confirmation** when the project is large (>5 modules already present).
- **Respect module dependencies**. Some modules require others (e.g., EventConsumer needs Events; API and AIAgent both have Shared as a hard dependency now because the auth runtime utilities live in Shared). If the user picks a module with missing prerequisites, offer to add the prerequisite first — `add_module` resolves them automatically but it's good to surface.
- **After adding API or AIAgent**, remind the user that OIDC Resource Server scaffolding ships dormant. The app refuses to boot when `trabuco.auth.enabled` is unset: operators must explicitly set it `=false` (local dev) or `=true` plus `OIDC_ISSUER_URI` and `OIDC_AUDIENCE` (deployed). The validator catches the missing-decision case at boot rather than letting "no auth" become silent factory state. See `docs/auth.md` for the per-provider matrix.
- **After adding AIAgent**, remind the user about: (1) A2A agent card at `/.well-known/agent.json` (served dynamically by `DiscoveryController`, advertises `mcp` only when MCP server is on); (2) MCP server is off by default — `MCP_SERVER_ENABLED=true` to opt in, scope-gated by `SCOPE_mcp:invoke` once enabled; (3) the legacy `ApiKeyAuthFilter` (`app.aiagent.api-key.enabled`, default on) which now requires operator-populated `agent.auth.keys.*` and refuses to boot with an empty key set. Demo keys live in `application-local-dev.yml` (`SPRING_PROFILES_ACTIVE=local-dev` to load).
