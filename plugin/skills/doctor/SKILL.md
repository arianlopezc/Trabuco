---
name: doctor
description: Validate a Trabuco project's structural health. Runs the Trabuco doctor + interprets findings. Use when a project isn't building correctly, after a manual edit, or before committing significant changes.
user-invocable: true
allowed-tools: [mcp__trabuco__run_doctor, mcp__trabuco__get_project_info, mcp__trabuco__check_docker]
argument-hint: "[project path, optional]"
---

# Run Trabuco's project health check

## Flow

1. **Target project**: if no argument, assume current working directory. If the argument is a path, use it. Call `mcp__trabuco__get_project_info` first to confirm it's a Trabuco project.

2. **Run doctor**: call `mcp__trabuco__run_doctor`. Capture the structured output — it includes per-module checks, dependency conflicts, missing artifacts, etc.

3. **Interpret**: don't dump the raw output. Group findings:
   - **Critical** (must fix — project won't build): missing POM entries, broken module dependencies, invalid configuration
   - **Warnings** (should fix — degraded behavior): outdated dep versions, missing CI workflows, un-gitignored secrets
   - **Hints** (nice-to-have): missing optional modules, suggested additions

4. **Fix proposals**: for each critical, offer a concrete fix. If it's something Trabuco itself would regenerate correctly, suggest `trabuco add` or regenerating the affected artifacts. If it's user code (broken entity, missing import), point at the file.

5. **Docker check** (if relevant): if the project uses Testcontainers and `mcp__trabuco__check_docker` reports it's not running, that's a critical finding for testing — flag it even if `run_doctor` didn't.

## Rules

- **Never auto-fix without confirmation**. Doctor findings are advisory; propose fixes, don't execute them.
- **Don't over-interpret**. If doctor returns empty, tell the user the project is healthy. Don't invent warnings.
