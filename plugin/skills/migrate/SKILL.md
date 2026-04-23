---
name: migrate
description: Migrate a legacy Java project into Trabuco's multi-module structure. Scans first to assess feasibility, then runs AI-powered code transformation. Use when the user has an existing Spring Boot / Java project and wants to adopt Trabuco's structure + AI-native integrations.
user-invocable: true
allowed-tools: [mcp__trabuco__scan_project, mcp__trabuco__migrate_project, mcp__trabuco__auth_status, mcp__trabuco__list_providers, mcp__trabuco__run_doctor]
argument-hint: "[path to legacy project]"
---

# Migrate a legacy Java project to Trabuco

Migration is expensive (LLM API calls, destructive if run on source). This skill enforces the safe sequence.

## Flow

1. **Ask for the path** if not provided. It must point at a legacy Java project directory, not a Trabuco-generated one.

2. **Scan first, always**: call `mcp__trabuco__scan_project`. It returns:
   - Project shape (Maven / Gradle, Spring Boot version, entity counts, etc.)
   - `migration_summary` with compatible/replaceable/unsupported dependencies
   - `has_blockers` boolean

   Read carefully. This is a READ-ONLY analysis.

3. **Report findings honestly**:
   - If `has_blockers: true`, list the blockers. Migration cannot proceed until the user addresses unsupported dependencies (usually by removing them or finding Trabuco-compatible alternatives).
   - If `has_blockers: false` but many dependencies are "replaceable," warn the user that the migration will replace them with Trabuco equivalents — breakage risk depends on how the app uses those libs.
   - If the project is tiny (<5 entities) or stateless, warn that migration may be more effort than rewriting from scratch via `/trabuco:new-project`.

4. **Check AI credentials**: call `mcp__trabuco__auth_status`. Migration uses an LLM. If no provider has creds, call `mcp__trabuco__list_providers` and instruct the user to configure one (typically Anthropic for best results with Java code). Do NOT proceed without a provider.

5. **Dry-run first**: call `mcp__trabuco__migrate_project` with `dry_run: true`. This produces a migration plan without modifying anything. Show the user the plan (modules proposed, files to transform).

6. **Confirm explicitly**: "Should I run the actual migration?" The answer must be yes, not just "sounds good." Migration creates a new directory but consumes LLM credits.

7. **Execute**: call `mcp__trabuco__migrate_project` without `dry_run`. Report progress and the output directory.

8. **Validate output**: immediately call `mcp__trabuco__run_doctor` on the migrated project. If there are critical findings, walk the user through fixes.

9. **Next steps**: tell the user the migration produced a NEW directory. Their original project is untouched. They should:
   - `cd` into the new directory
   - Run `mvn clean install` to check build
   - Review the generated structure against their expectations
   - Commit to a new branch, not directly to main

## Rules

- **Never skip the scan**. It's the only way to know if migration is feasible.
- **Never run without dry-run first**. LLM-powered migration is expensive to re-run from scratch.
- **Never touch the source directory**. Output always goes to a new path.
