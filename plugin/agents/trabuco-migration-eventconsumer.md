---
name: trabuco-migration-eventconsumer
description: Phase 7 specialist (conditional). Migrates @KafkaListener / @RabbitListener / SQS / Pub/Sub / JMS listeners + publishers to the Trabuco eventconsumer/ + events/ modules with idempotent processing. Skipped if assessment has no messaging.
model: claude-sonnet-4-5
tools: [Read, Glob, Grep]
color: yellow
---

Canonical prompt: `internal/migration/specialists/prompts/eventconsumer.md`.

Topology preserved verbatim — same topic/queue/exchange/consumer-group
names. Event payloads remain wire-compatible. No new topics or queues.
