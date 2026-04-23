---
name: extend
description: Propose what to add next to an existing Trabuco project based on its current state and what the user wants to achieve. Different from /trabuco:add-module — that adds a specific module; this one ADVISES on what to add. Use when the user has a project and asks "what's next?" or "how would you extend this?"
user-invocable: true
allowed-tools: [mcp__trabuco__get_project_info, mcp__trabuco__scan_project, mcp__trabuco__suggest_architecture, mcp__trabuco__list_modules]
argument-hint: "[goal description]"
---

# Advise on extending an existing Trabuco project

## Flow

1. **Load context**:
   - `mcp__trabuco__get_project_info` for the current project
   - If ambiguous, `mcp__trabuco__scan_project` for a deeper structural read
   - Read `trabuco://modules` resource to know the full catalog

2. **Load expertise**: request the `extend_project` MCP prompt (argument: project path). This loads Trabuco's extension philosophy — the reasoning framework the advice should follow.

3. **Understand the goal**: if the user's argument is vague, ask specifically what they want to enable. "Make it faster" is too vague; "handle 10k concurrent requests" is actionable.

4. **Match against patterns**: call `mcp__trabuco__suggest_architecture` with the extended requirements (current + goal). Compare the recommended config against what's currently installed. The delta is the extension proposal.

5. **Present as a plan, not a wall of text**: for each proposed addition, say:
   - What (specific module or pattern)
   - Why (how it addresses the goal)
   - Trade-off (what complexity it adds)
   - Effort (rough)

6. **Offer to execute**: if the user approves, delegate to `/trabuco:add-module` for each addition. If the proposal involves creating new entities/endpoints/events, mention that the project's own `.claude/skills/` already has `/add-entity`, `/add-endpoint`, etc. for the implementation work.

## Rules

- **Don't over-prescribe**. If the current project already does what the user wants and the bottleneck is elsewhere, say that. "You don't need to change the architecture — you need to fix the query in `FooRepository.findAll()`" is a valid answer.
- **Respect limitations**. Read `trabuco://limitations` before proposing something Trabuco can't help with (e.g., adding GraphQL, adding a React frontend).
- **Prefer conservative additions**. Adding a module is cheap to propose but expensive to maintain. Default to "fewer modules, clearer boundaries."
