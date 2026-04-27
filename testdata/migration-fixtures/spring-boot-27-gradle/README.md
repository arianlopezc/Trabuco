# Acme Inventory — Spring Boot 2.7 + Gradle fixture

Gradle build-system path for the Trabuco 1.10 migration. Trabuco's
skeleton-builder is Maven-only and the validation funnel invokes `mvn`
directly; this fixture exists to verify that the assessor cleanly blocks
Gradle source at Phase 0 with a `NON_MAVEN_BUILD_SYSTEM` blocker rather
than crashing or producing a half-converted project.

Expected outcome — `trabuco migrate assess` produces:

```
[blocked] source build system is "gradle"; Trabuco 1.10 supports Maven only
  blocker: NON_MAVEN_BUILD_SYSTEM — ...
```

with two alternatives offered: convert to Maven (`gradle init --type pom`)
or wait for 1.11+ Gradle support.
