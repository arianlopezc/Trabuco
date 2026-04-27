# Acme Orders Events — Spring Boot 2.7 + Kafka fixture

Messaging path for the Trabuco 1.10 migration. Exercises:

- `@KafkaListener` consumer — Phase 7 EventConsumer
- `KafkaTemplate` publisher service — Phase 4 Shared / Phase 7 publisher
  catalog
- No persistence (no entities, no repositories) — Phases 2/3
  `not_applicable`
- No web layer (no `@RestController`) — Phase 5 `not_applicable`
- No scheduled jobs — Phase 6 `not_applicable`

Use as a target for `trabuco migrate run`:

```
trabuco migrate run /path/to/spring-boot-27-kafka
```
