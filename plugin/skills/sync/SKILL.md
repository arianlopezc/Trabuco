---
name: sync
description: Bring an existing Trabuco project's AI tooling up to date with the installed CLI. Detects missing skills, subagents, prompts, hooks, and review scaffolding, then creates them additively. Never modifies existing files.
user-invocable: true
allowed-tools: [mcp__trabuco__sync_project, mcp__trabuco__get_project_info]
argument-hint: "[project path]"
---

# Sync AI tooling in a Trabuco project

This skill updates the AI-tooling layer of an existing Trabuco project so its skills, subagents, prompts, hooks, and review scaffolding match what the currently-installed Trabuco CLI would generate today.

## Flow

1. **Identify the project.** If the user didn't provide a path, use the current working directory. Call `mcp__trabuco__get_project_info` to confirm it's a Trabuco project — if it returns an error, tell the user and stop.

2. **Dry-run first.** Call `mcp__trabuco__sync_project` with `apply: false` to produce a plan. The response is JSON with:
   - `would_add`: files the current CLI would emit that are missing
   - `already_present`: files that exist (untouched regardless of content)
   - `out_of_jurisdiction`: business / infra files sync never considers
   - `blockers`: reasons sync can't proceed (e.g., no `.trabuco.json`)

3. **Summarize to the user.** Present the `would_add` list grouped by directory (.claude/, .codex/, .cursor/, .github/, .ai/, top-level). Keep the explanation short — the plan output speaks for itself.

4. **Confirm.** If `would_add` is non-empty, ask the user "Apply these N additions?" Wait for explicit confirmation (not just acknowledgement — "yes" or "apply", not "sounds good").

5. **Apply.** On confirmation, call `mcp__trabuco__sync_project` with `apply: true`. Report the count of files added.

6. **Next steps.** Tell the user:
   - "Restart your Claude Code session if you want the new hooks and skills to load immediately."
   - "Run `trabuco doctor` to confirm project health."
   - "Review the changes with `git diff` before committing."

## Rules

- **Never skip the dry-run.** The user should always see what's going to happen before it happens.
- **If `would_add` is empty, report that clearly** ("AI tooling is already up to date") and stop — no need to confirm or apply.
- **If there are blockers**, surface them verbatim and suggest fixes (usually: make sure the directory is a Trabuco project with `.trabuco.json`).
- **Never claim sync modified existing files.** It doesn't. If the user asks about refreshing a specific existing file (like `CLAUDE.md`), explain that sync is additive-only and the workaround is to delete the file first, then run sync.
- **Don't imply sync touches business code.** Sync has strict jurisdiction boundaries: Java source, POMs, `application.yml`, `docker-compose.yml`, migrations, and CI workflows (other than `copilot-setup-steps.yml`) are never considered. If the user asks sync to "update my application.yml," redirect them.

## Example dialog

> User: my project feels stale, can you update the AI helpers?
> You: [call get_project_info, then sync_project with apply=false]
> You: Sync would add 4 files to /path/to/project:
>   - `.claude/skills/add-test/SKILL.md`
>   - `.claude/subagents/code-simplifier.md`
>   - `.ai/prompts/testing-guide.md`
>   - `.cursor/rules/add-endpoint.mdc`
>   Apply these?
> User: yes
> You: [call sync_project with apply=true]
> You: Added 4 files. Restart your Claude Code session to pick up the new skills and subagents.
