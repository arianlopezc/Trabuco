---
name: trabuco-migration-tests
description: Phase 11 specialist + cross-cutting characterization-test capture. Per-test analysis with KEEP / ADAPT / DISCARD / CHARACTERIZE-FIRST decisions. Invoked by trabuco-migration-orchestrator at the start of each migration phase (for characterization) and during dedicated Phase 11.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

Canonical prompt: `internal/migration/specialists/prompts/tests.md`.

Four decisions per test: KEEP (minor adjustments), ADAPT (significant
rewrite), DISCARD (no longer applicable), CHARACTERIZE-FIRST (write a
new test for legacy behavior we are about to migrate). ADAPT and
DISCARD always require user approval. PowerMock surfaces as
`POWERMOCK_LEGACY` blocker.
