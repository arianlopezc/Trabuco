---
name: design-system
description: Decompose a multi-service system into independent Trabuco services. Use when the user is describing a system of multiple cooperating services (marketplace, platform, microservices architecture) rather than a single service.
user-invocable: true
allowed-tools: [mcp__trabuco__design_system, mcp__trabuco__generate_workspace, mcp__trabuco__list_modules, mcp__trabuco__suggest_architecture]
argument-hint: "[system description]"
---

# Design a multi-service Trabuco system

Turn high-level system requirements into a workspace of independent Trabuco services, each with its own module selection.

## When to pick this skill over /trabuco:new-project

- User says "system", "platform", "marketplace", "multiple services", "event-driven architecture across teams", "CQRS with separate read/write services", or similar.
- The workload naturally splits — different services own different entities, communicate via events, can be deployed independently.

If only ONE service will own the domain, use `/trabuco:new-project` instead.

## Flow

1. **Delegate to the specialist**: for the decomposition itself, invoke the `trabuco-architect` subagent if it's available — it loads the `design_microservices` MCP prompt which encodes Trabuco's boundaries doctrine.

2. **If running skill directly**: call `mcp__trabuco__design_system` with the user's requirements. It returns a proposed decomposition (services, their modules, inter-service communication patterns).

3. **Review with the user**: present each service with:
   - Its purpose (one sentence)
   - Its module selection (from Trabuco's catalog)
   - How it communicates with siblings (REST? Events? A2A?)
   - What it owns (entities, events)

   Surface decomposition tensions honestly: if two services would share entities heavily, flag it and ask the user whether to merge them.

4. **Iterate if needed**: user may want to split a service further or merge two. Re-call `design_system` with adjusted requirements, or manually adjust the proposed config.

5. **Generate the workspace**: once locked, call `mcp__trabuco__generate_workspace` with the approved design. This creates a workspace root + per-service directories, each a full Trabuco project.

6. **Next steps**: tell the user how to build the workspace (typically `mvn clean install` at the workspace root, or per-service), and how the A2A agent cards discover each other if AIAgent modules are involved. **Auth across services**: every service whose module list includes API or AIAgent gets dormant OIDC Resource Server scaffolding. For consistent cross-service identity, point all services at the same `OIDC_ISSUER_URI` (one IdP, multiple resource servers) and flip `trabuco.auth.enabled=true` when ready. The same `IdentityClaims` shape lives in each service's Model module so token claims propagate cleanly via headers / event payloads. Full guide: `docs/auth.md`.

## Rules

- **Services are independent deployables, not packages in a monolith.** If the user is designing a modular monolith, talk them out of `generate_workspace` and toward `/trabuco:new-project` with multiple modules instead.
- **Event-driven communication by default.** REST-between-services is fine but creates tight coupling; warn the user when the design has >3 direct REST hops between services.
- **Don't fracture on anemic reasons.** "Each entity gets its own service" is an anti-pattern. Flag if the decomposition proposes tiny services with no independent lifecycle.
