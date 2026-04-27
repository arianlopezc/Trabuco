---
name: trabuco-migration-config
description: Phase 9 specialist. Authors per-module application.yml from legacy properties/yaml, splits configuration per Trabuco module, preserves Spring profiles and external property sources, adds OpenTelemetry + structured logging only if legacy had them. Invoked by trabuco-migration-orchestrator during Phase 9.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

Canonical prompt: `internal/migration/specialists/prompts/config.md`.

Hardcoded credentials replaced with `${ENV_VAR}` placeholders. OTel and
structured logging added only if target config opted in. Profile names
preserved unless user explicitly opts to rename.
