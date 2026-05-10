---
name: plan
description: Produce a stratified implementation plan for a multi-module change in a Trabuco-shaped Java/Spring Boot project. Each plan stage is one module in dependency order (Foundation → Contracts → Persistence → BusinessLogic → Edge), with per-stage validate command and depends-on rationale. Use when the user wants an explicit plan-then-execute workflow before code changes — high-stakes features, team review processes, or just to make the plan reviewable as a separate artifact.
user-invocable: true
allowed-tools: [mcp__trabuco__get_project_info, Read, Grep, Glob, Bash, Write]
argument-hint: "[task description]"
---

# Stratified plan for a Trabuco-project change

This skill is the **explicit gate** version of stratified planning: you
invoke it deliberately when you want a plan as a reviewable artifact
before any code changes. The same shape is also produced ambiently by
the `trabuco-planner` subagent when CLAUDE.md's delegation rule fires —
this skill is for cases where you want to drive the gate yourself.

## What you'll get

A markdown plan in canonical Trabuco-stratified shape:

```
# Plan: <task>

## Stage 1 — Model
- Files added/modified
**Validate:** `mvn -pl Model install`

## Stage 2 — Jobs (or whichever Layer-1 contract module applies)
...

## Stage 3 — SQLDatastore (or NoSQL)
...

(continues through layers present in this project)

## Final verification
- mvn install across the reactor
- trabuco doctor reports clean
```

For non-trivial features (3+ stages), the plan is also saved at
`.trabuco/plans/<date>-<slug>.md` so the PR description can reference
it and reviewers can compare diff to plan.

## Flow

1. **Confirm the project shape.** Run `mcp__trabuco__get_project_info`
   (or `cat .trabuco.json | jq .modules` if MCP isn't available). If
   the user is not in a Trabuco project, instruct them to either `cd`
   in or run `trabuco init` first — stratified planning doesn't apply
   outside Trabuco's hierarchy.

2. **Delegate to the trabuco-planner subagent.** The persona owns the
   plan-production logic: archetype detection (additive feature / job
   feature / event feature / rename / bug fix / single-module /
   migration-only), per-stage validate commands, depends-on rationale,
   anti-pattern avoidance.

   Invoke as: ask the main agent to spawn `@trabuco-planner` with the
   user's task description. The planner reads `.trabuco.json`, picks
   the archetype, and produces the plan.

3. **Render the plan for review.** Show it in chat. Ask:

   ```
   📋 Plan ready. Approve to execute, request changes, or cancel?
   ```

4. **On approval, hand off to the main agent for execution.** The main
   agent works through the stages in order, running each `Validate:`
   command before proceeding to the next stage.

5. **On request-changes, iterate with the planner.** Ask the user
   what's wrong; refine the plan; re-render for approval. Don't
   execute until the plan is right.

## When to use this skill

- The change spans multiple modules and you want explicit review
  before code lands
- You're working on a high-stakes feature and your team requires
  plan-first workflow
- You want the plan saved as a versioned artifact for the PR
- You want to surface architectural decisions (which modules, which
  archetype, where the contract/consumer boundary falls) before any
  code is written

## When NOT to use this skill

- The change is single-module — the planner produces a one-stage plan
  and the ceremony is wasteful
- The task is trivial (typo fix, dependency bump, formatting)
- You're already mid-implementation — the planner is for the design
  phase, not for retroactive documentation

## Rules

- **Always delegate to `trabuco-planner`** — this skill is the entry
  point, not the implementation. The planner persona is where the
  archetype-detection and convention-checking logic lives.
- **Never execute before approval** — render the plan, await user
  approval, only then hand off to the main agent for code changes.
- **Persist non-trivial plans** to `.trabuco/plans/` so the PR has a
  versioned record of the design intent.
