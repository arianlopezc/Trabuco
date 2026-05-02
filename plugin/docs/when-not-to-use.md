# When NOT to use Trabuco

The worst failure mode for this plugin is generating a project the user later regrets. This doc is the fast filter. Before recommending `/trabuco:new-project`, check this list — if ANY apply, either steer the user elsewhere or force an explicit conversation about the tradeoff.

## Hard stops — do not recommend Trabuco

### The project's core is outside Trabuco's scope

- **Frontend-dominant projects** — a React/Next.js/Vue SPA with a thin backend. Trabuco won't help meaningfully. Recommend the frontend tool of choice + Trabuco only if there's real backend complexity.
- **GraphQL-first services** — Trabuco emits REST. Fighting it to produce a GraphQL schema is a losing battle.
- **gRPC-only services** — no Protobuf, no gRPC. Use `grpc-spring-boot-starter` directly without Trabuco.
- **WebSocket-centric services** — no SSE, no STOMP, no WebSocket endpoints. Use Spring WebFlux or Spring WebSocket directly.

### Not a JVM project

If the user is considering Go, Python, Node, Rust, .NET — Trabuco is the wrong answer. Say so clearly.

### Not Spring Boot

- Quarkus users: Trabuco doesn't target Quarkus. No help here.
- Micronaut, Dropwizard, Helidon, Ktor: same.
- Spring Boot 2.x: Trabuco is 3.4.2-only. Users stuck on 2.x will hit breaking changes.

## Strong warnings — explicit tradeoff conversation required

### Tiny projects (<5 entities, <10 endpoints)

Trabuco generates a full multi-module Maven layout with review tooling, skills, hooks, and scaffolding for 8+ modules. For a 2-file toy, this is overkill.

**Better move:** Recommend Spring Initializr or a single-file Spring Boot Java class. Reserve Trabuco for projects that benefit from the module structure.

### User wants "a quick prototype"

Trabuco projects are production-shaped — migrations, review checks, Testcontainers, keyset pagination, multi-module Maven. A prototype doesn't need that ceremony and the user will fight it.

**Better move:** Ask whether the prototype will become production. If yes, Trabuco now saves refactoring later. If no, use something lighter.

### User has a strongly-held preference against a Trabuco convention

Trabuco is opinionated. If the user insists on:

- Offset pagination
- FK constraints in migrations
- JPA/Hibernate instead of Spring Data JDBC
- Lombok instead of Immutables
- `@Autowired` field injection
- Mocking the database in integration tests

...then the generated project's review tooling will fight them on every commit. That's often what the user wants (enforcement) — but if they're opposed to the convention itself, Trabuco will feel like a prison. Ask. If they want the escape hatches, Trabuco isn't it.

### User is migrating a massive legacy codebase

Migration is supported as of Trabuco 1.10 — `/trabuco:migrate` drives a 14-phase orchestrated flow with per-phase approval gates and atomic rollback (see `docs/migration-guide.md`). Caveats to set expectations up front:

- **Maven only.** Gradle source is blocked at Phase 0 with `NON_MAVEN_BUILD_SYSTEM`; the user must convert (`gradle init --type pom`) first.
- **Spring Boot 2.x or 3.x.** Quarkus / Micronaut / non-Spring frameworks block as `NON_SPRING_FRAMEWORK`.
- **Each phase is one Anthropic API call.** A typical 14-phase migration costs real tokens. Warn cost-conscious users.
- **The build JDK on PATH must match the target Java version.** Mismatch is caught by a preflight check, but the user should set `JAVA_HOME` before they start.

For projects outside those bounds, manual extraction (entity-by-entity, with `/trabuco:new-project` producing the target shape) is still the right answer.

### User needs regulatory compliance (PCI, HIPAA, SOC2)

Trabuco generates standard Spring Boot code. It does NOT ship PCI-compliant logging, HIPAA-aware data handling, or SOC2 audit trails. The user will add those on top. That's fine — but they should know up front that compliance work is theirs.

What Trabuco *does* ship is a structured security audit (`/trabuco:audit` or `/audit` inside a generated project) that walks 173 checks across auth, AI surface, AIAgent runtime + Java platform, data + events, web + infra — covering OWASP Top 10 (2021), OWASP API Security Top 10 (2023), OWASP LLM Top 10 (2025), an ASVS L1/L2 subset, the CWE Top 25, and Trabuco-specific patterns. That's a strong baseline but not a substitute for compliance attestation work — frame it as "we've passed the OWASP-class audit, now we still need to prove the controls and evidence trail your auditor wants."

## Soft warnings — proceed, but note the tradeoff

### Single-team project that will never be split

Trabuco's multi-module structure is partly about enforcing bounded contexts. If a team of 1–2 engineers will own every line forever, the module boundaries are less valuable. They still give you good separation, but the ceremony cost is real.

### User wants a lot of customization

Trabuco's value is its opinions. Users who want to customize the generation deeply (custom templates, custom conventions) will find themselves forking. If they're heading that direction, help them understand: Trabuco's not the right base.

### User doesn't know Spring Boot

Trabuco assumes the user is comfortable with Spring Boot. If they're a Spring novice, the generated project will look like a lot. That's not a Trabuco failure — but it's a reason to budget learning time, not just generation time.

## Red flags in requirement gathering

These aren't disqualifying, but they should trigger "ask a clarifying question" behavior:

- "I want something like Netflix/Uber/big-tech" — usually overspec. Ask about team size and actual scale.
- "I want everything" — rarely correct. Every module has a cost. Push back.
- "I want to use Trabuco AND a completely different stack" — ask why. Sometimes this means two separate services, sometimes it means Trabuco isn't fitting.
- "Can we turn off X" where X is a Trabuco convention — investigate WHY they want it off. The convention usually has a reason; if they know and still want it off, Trabuco's not their tool.

## Honest framing to the user

When Trabuco isn't the right fit, the best service is saying so. Sample phrasings:

- "Trabuco's for Spring Boot 3 multi-module Java services. For a Next.js app with minimal backend, you'd spend more time fighting Trabuco's opinions than they'd save you. I'd recommend a different approach."
- "Your requirement is essentially 'a GraphQL API.' Trabuco emits REST only and isn't going to help with schema-first design. I'd reach for Netflix DGS or graphql-java directly."
- "You want to prototype this over the weekend? Trabuco will give you more scaffolding than a weekend prototype needs. Use Spring Initializr and come back to Trabuco when the prototype proves out."

Honesty here builds trust. Users remember "you told me not to use the tool" better than "you sold me the tool."
