# Authentication

Trabuco generates production-ready OIDC Resource Server scaffolding
whenever the **API** or **AIAgent** module is selected. Every generated
project includes the JWT validation chain, the scope-mapping converter,
RFC 7807 ProblemDetail handlers, and test utilities. Which chain is
active is selected at runtime by `trabuco.auth.enabled`, which **must
be set explicitly** — there is no implicit default. The generated app
refuses to boot if the property is unset.

This document covers what gets generated, how to enable it, how to wire
each major IdP, and how to test code that depends on identity.

## Enabling auth

Auth is gated at runtime by a single property that **must be set
explicitly**:

```yaml
trabuco:
  auth:
    enabled: ${TRABUCO_AUTH_ENABLED:}   # no default — required
```

`SecurityConfig#validateAuthDecisionMade` (and the AIAgent equivalent)
reads this property at boot and refuses to start the application if it
is unset, blank, or anything other than `true`/`false`. The error message
points operators back to this document. The intent is that no service
ever ships with neither filter chain wired because someone forgot a
property.

To turn the JWT validation chain on:

1. Set `trabuco.auth.enabled=true` (or export `TRABUCO_AUTH_ENABLED=true`).
2. Set `OIDC_ISSUER_URI` to your IdP's discovery endpoint (see provider
   recipes below).

To run locally without an IdP, or — for AIAgent — to operate on the
legacy API-key path only, set `trabuco.auth.enabled=false` explicitly.

When set to `false`, the `permitAllFilterChain` bean is the active
filter chain and no JWT enforcement runs at the HTTP layer. When set
to `true`, the `oauth2FilterChain` bean takes over and validates
incoming JWTs against the configured issuer. The two chains are mutually
exclusive (`@ConditionalOnProperty` pair) so exactly one is wired at
any time.

Both `SecurityConfig` (API) and `AgentSecurityConfig` (AIAgent) follow
the same dual-chain pattern.

## Why an explicit decision is required

A scaffolded service is one rushed deploy away from production. If the
runtime gate had an implicit default, a developer who forgot to set the
property would ship — silently — with no auth on the API module (which
has no fallback) or with only the seeded API-key path on AIAgent. The
explicit-decision guardrail makes that mistake impossible: the app
won't boot until the operator chooses.

The auth code is always present in source so it's discoverable — you
can read it, customize the claim extractor, drop in scope checks on
controllers — without re-running the generator or installing a CLI
flag. The choice between `enabled=true` and `enabled=false` is a
runtime concern, not a code-generation concern.

When `enabled=false`, the JWT chain bean isn't created, no JwtDecoder
is wired, the issuer URI is never read — so failure modes like
"silently accepting unsigned tokens because the issuer was misconfigured"
are physically impossible.

## What gets generated

Every project that includes API or AIAgent gets the full file set below:

**Model module — universal data types**
- `model/auth/IdentityClaims.java` — Immutables value object: `subject`,
  `email`, `tenantId`, `scopes`, `rawClaims`. Read by API filters,
  Worker handlers, EventConsumer listeners, and AIAgent tools.
- `model/auth/AuthorityScope.java` — centralized scope constants
  (`SCOPE_*` raw form + `AUTHORITY_*` Spring Security-prefixed form).
- `model/auth/AuthenticatedRequest.java` — generic wrapper record for
  carrying identity alongside any payload across async boundaries (job
  parameters, broker bodies, webhook payloads).

**Shared module — cross-module logic**
- `shared/auth/RequestContextHolder.java` — per-request `ThreadLocal`
  for the current `IdentityClaims`.
- `shared/auth/JwtClaimsExtractor.java` — interface; replace the
  default bean for non-RFC-conformant providers.
- `shared/auth/DefaultJwtClaimsExtractor.java` — `@Component` reading
  the standard `scope` / `scp` claim.
- `shared/auth/AuthContextPropagator.java` — interface for serializing
  claims into broker headers / job metadata.
- `shared/auth/DefaultAuthContextPropagator.java` — `@Component` impl
  using `X-Identity-*` header convention.
- `shared/auth/AuthScope.java` — try-with-resources helper:
  ```java
  try (var scope = AuthScope.set(claims)) { doWork(); }
  ```

**Shared test utilities**
- `shared/auth/MockJwtFactory.java` — mint fake but well-shaped JWTs,
  authentication tokens, and decoders for use in tests. Three usage
  patterns documented in the class javadoc.

**API module — HTTP filter chains**
- `api/config/security/SecurityConfig.java` — **dual `SecurityFilterChain`
  beans** gated on `trabuco.auth.enabled`. The JWT chain wires stateless
  validation, `@EnableMethodSecurity`, and permit-all for actuator +
  swagger. The permit-all chain is the default fallback.
- `api/config/security/JwtAuthenticationConverter.java` — `Jwt` →
  `JwtAuthenticationToken`, populates `RequestContextHolder`.
- `api/config/security/AuthProblemDetailHandler.java` — RFC 7807
  `application/problem+json` for 401/403.
- `api/config/security/OpenApiSecurityConfig.java` — adds `bearerAuth`
  scheme to Swagger UI's "Authorize" button.
- `api/test/security/SecurityIntegrationTest.java` — MockMvc test that
  sets `trabuco.auth.enabled=true` and verifies the 401 + ProblemDetail
  path with a mocked `JwtDecoder`.
- `api/test/security/AuthEndToEndTest.java` — `@SpringBootTest` with
  RANDOM_PORT, real RS256-signed tokens, real signature verification.
  Sets `trabuco.auth.enabled=true` for the e2e suite.
- `api/test/security/SignedJwtTestSupport.java` — generates a fixed RSA
  key pair, signs JWTs with the private half, exposes a `JwtDecoder`
  configured with the public half.

**AIAgent module — JWT path coexisting with API-key path**
- `aiagent/config/security/AgentSecurityConfig.java` — same dual-chain
  pattern as `SecurityConfig`, gated on `trabuco.auth.enabled`.
- `aiagent/config/security/JwtAuthenticationConverter.java`,
  `AuthProblemDetailHandler.java` — duplicated from API since AIAgent
  doesn't depend on API.
- `aiagent/security/ApiKeyAuthFilter.java` is governed by its **own**
  property `app.aiagent.api-key.enabled` (default `true`), independent
  of `trabuco.auth.enabled`. The legacy tier-based path stays on by
  default; turn it off with `app.aiagent.api-key.enabled=false`.

## Provider configuration

Auth is configured via the standard Spring Boot property
`spring.security.oauth2.resourceserver.jwt.issuer-uri`, which in the
generated `application.yml` is sourced from the `OIDC_ISSUER_URI`
environment variable. The yaml block is always present; an empty value
is safe because Spring's `IssuerUriCondition` uses `StringUtils.hasText`
to skip JwtDecoder creation when unset.

Below are the values for each major IdP.

### Keycloak

```bash
export TRABUCO_AUTH_ENABLED=true
export OIDC_ISSUER_URI=http://localhost:8180/realms/myrealm
# OIDC_AUDIENCE is required (validateAuthDecisionMade refuses to boot without it).
# Keycloak doesn't put aud=<client_id> by default — configure a token mapper
# to add the client_id (or any service identifier you choose) as the aud claim,
# then set OIDC_AUDIENCE to that value.
export OIDC_AUDIENCE=my-api-service
```

For local dev, run Keycloak in Docker:

```bash
docker run -d --name keycloak \
  -p 8180:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:25.0 start-dev
```

Then create a realm + client + user in the Keycloak admin UI at
`http://localhost:8180`. The default `DefaultJwtClaimsExtractor` reads
the standard `scope` claim that Keycloak emits — no customization
needed.

### Auth0

```bash
export TRABUCO_AUTH_ENABLED=true
export OIDC_ISSUER_URI=https://YOUR_TENANT.auth0.com/
export OIDC_AUDIENCE=https://your-api-identifier
```

**Custom extractor required.** Auth0 emits scopes in the standard
`scope` claim, but custom claims (roles, permissions, tenant) are
namespaced under your domain (e.g.,
`https://YOUR_TENANT/permissions`). Replace
`DefaultJwtClaimsExtractor` with a custom bean:

```java
@Component
@Primary  // overrides the default
public class Auth0JwtClaimsExtractor implements JwtClaimsExtractor {
    @Override
    public IdentityClaims extract(Jwt jwt) {
        return ImmutableIdentityClaims.builder()
            .subject(jwt.getSubject())
            .email(Optional.ofNullable(jwt.getClaimAsString("email")))
            .scopes(Set.copyOf(jwt.getClaimAsStringList(
                "https://YOUR_TENANT/permissions")))
            .rawClaims(jwt.getClaims())
            .build();
    }
}
```

### Okta

```bash
export TRABUCO_AUTH_ENABLED=true
export OIDC_ISSUER_URI=https://YOUR_DOMAIN.okta.com/oauth2/default
# Set OIDC_AUDIENCE to the API resource identifier configured in your
# Okta authorization server (Authorization Servers → Default → Settings → Audience).
export OIDC_AUDIENCE=api://your-api
```

Okta emits standard `scp` (array form). The default extractor handles
both `scope` and `scp` automatically — no customization needed.

### AWS Cognito

```bash
export TRABUCO_AUTH_ENABLED=true
export OIDC_ISSUER_URI=https://cognito-idp.{region}.amazonaws.com/{userPoolId}
# Cognito puts the app client_id in the aud claim by default for ID tokens.
# For access tokens, the aud claim is named client_id (not aud); see the
# custom JwtClaimsExtractor below if you need access-token validation.
export OIDC_AUDIENCE=YOUR_APP_CLIENT_ID
```

**Custom extractor required.** Cognito puts group membership in
`cognito:groups` (not in `scope`). Replace
`DefaultJwtClaimsExtractor`:

```java
@Component
@Primary
public class CognitoJwtClaimsExtractor implements JwtClaimsExtractor {
    @Override
    public IdentityClaims extract(Jwt jwt) {
        Set<String> scopes = new LinkedHashSet<>();
        // Cognito puts OAuth scopes in `scope` (space-delimited)
        String scopeClaim = jwt.getClaimAsString("scope");
        if (scopeClaim != null) {
            scopes.addAll(Arrays.asList(scopeClaim.split(" ")));
        }
        // ...AND group memberships in `cognito:groups` (array)
        List<String> groups = jwt.getClaimAsStringList("cognito:groups");
        if (groups != null) {
            scopes.addAll(groups);
        }
        return ImmutableIdentityClaims.builder()
            .subject(jwt.getSubject())
            .scopes(scopes)
            .rawClaims(jwt.getClaims())
            .build();
    }
}
```

### Generic OIDC

For any other RFC-conformant OIDC provider:

```bash
export TRABUCO_AUTH_ENABLED=true
export OIDC_ISSUER_URI=https://your-idp.example.com
# Pick whatever value your IdP mints into the aud claim — typically
# the API URL or a configured "audience" / "resource" identifier.
export OIDC_AUDIENCE=https://your-service-identifier
```

The default extractor reads the standard `scope` (or `scp`) claim. If
the provider emits a non-standard layout, register a custom
`JwtClaimsExtractor` bean with `@Primary`.

## Configuring the legacy API-key path (AIAgent)

The AIAgent module ships with two coexisting auth mechanisms — the
legacy tier-based `ApiKeyAuthFilter` (governed by
`app.aiagent.api-key.enabled`) and the new JWT chain (governed by
`trabuco.auth.enabled`). The matrix:

| `app.aiagent.api-key.enabled` | `trabuco.auth.enabled` | Behavior |
|-----------------------------|-----------------------|----------|
| `true` (default)            | `false` (explicit)    | Legacy API-key path is the only HTTP-layer auth. Pre-auth-scaffolding behavior preserved. |
| `true`                      | `true`                | Hybrid — both filters run, either credential type accepted. |
| `false`                     | `true`                | JWT-only. Recommended target state. |
| `false`                     | `false`               | No HTTP-layer auth enforced. (Don't run this in prod.) |
| any                         | *unset*               | App refuses to boot. `validateAuthDecisionMade` rejects the missing property. |

### Configuring API keys (no seeded defaults since 1.12)

Trabuco no longer ships seeded keys (`partner-secret-key`,
`public-read-key`) in source. The filter consumes
`@ConfigurationProperties("agent.auth")` and refuses to boot when it is
enabled but the keys map is empty. Operators must populate the keys via
config:

```yaml
# application-prod.yml (or any env-specific override)
agent:
  auth:
    keys:
      "${PARTNER_KEY_VALUE}":           # the literal bearer token clients send
        tier: partner                   # public | partner — drives ScopeEnforcer
        label: prod-partner-2026-q1     # rotation tag for logs / metrics
      "${INTERNAL_KEY_VALUE}":
        tier: partner
        label: internal-services
```

Inject the key values via env vars / Spring Cloud Config / K8s Secret —
do not commit literals. To rotate a key, add the new entry, deploy, then
remove the old one once clients have rolled forward.

For local development, the `local-dev` profile loads two demo keys
(`dev-public-key`, `dev-partner-key`) from `application-local-dev.yml`
and emits a noisy startup `WARN`:

```bash
SPRING_PROFILES_ACTIVE=local-dev mvn spring-boot:run -pl AIAgent

curl -H "Authorization: Bearer dev-partner-key" \
     http://localhost:8080/ingest -d '{...}'
```

Do **not** activate the `local-dev` profile in any deployed
environment. The `DemoKeyStartupWarning` bean only registers under that
profile, so CI / staging / prod never see the demo keys.

Incremental migration path:

1. **Project is generated with both paths already wired** — no
   regeneration needed.
2. **Turn the JWT chain on.** Set `trabuco.auth.enabled=true` and
   configure `OIDC_ISSUER_URI` + `OIDC_AUDIENCE`. Existing
   `@RequireScope` annotations keep working — `ScopeEnforcer` bridges
   the JWT scope claim into the same tier ladder.
3. **Optionally adopt `@PreAuthorize` for new endpoints.** Use
   `@PreAuthorize("hasAuthority('SCOPE_*')")` when you need finer
   scopes than the three-tier ladder offers. The two annotations
   coexist on the same controller without conflict.
4. **Disable the legacy filter when ready.** Set
   `app.aiagent.api-key.enabled=false` in `application.yml`. Existing
   `@RequireScope` annotations keep enforcing the tier ladder via the
   JWT bridge — you do not need to remove them.

The two systems are bridged in `ScopeEnforcer` (since 1.12). The JWT
authority → tier mapping:

| JWT authorities | Effective tier |
| --------------- | -------------- |
| `SCOPE_partner`, `SCOPE_agent:write`, `SCOPE_agent:admin` | `partner` |
| `SCOPE_public`, `SCOPE_agent:read` | `public` |
| Authenticated, no recognized scope | `public` (any valid token earns baseline trust) |
| Anonymous / no JWT | `anonymous` |

Hybrid mode (both filters on) takes the higher tier of (API-key
result, JWT result), so a request authenticated through either path
with sufficient scope is allowed.

## Testing

`MockJwtFactory` provides three patterns for testing code that depends
on identity. See its javadoc for full examples; the short version:

```java
// Pattern 1 — unit test of code that receives Jwt
Jwt jwt = MockJwtFactory.jwt("user-42", "agent:read", "agent:write");

// Pattern 2 — @SpringBootTest with mocked decoder
@MockBean JwtDecoder jwtDecoder;
@BeforeEach void setup() {
    when(jwtDecoder.decode(any()))
        .thenReturn(MockJwtFactory.jwt("user-42", "admin"));
}

// Pattern 3 — set SecurityContext manually
SecurityContextHolder.getContext().setAuthentication(
    MockJwtFactory.authentication("user-42", "agent:read"));
```

Tests that exercise the JWT filter chain must set
`trabuco.auth.enabled=true` explicitly. Tests that don't care about
auth (or that exercise the open chain) must set it to `false`. Either
way, the property must appear in the test's `@TestPropertySource` /
`@SpringBootTest(properties=...)` block — `validateAuthDecisionMade`
refuses to boot without it:

```java
@SpringBootTest
@TestPropertySource(properties = "trabuco.auth.enabled=true")
class MyAuthenticatedEndpointTest { ... }
```

The generated `SecurityIntegrationTest` (in `API/test/security/`) is a
working MockMvc example. The generated `AuthEndToEndTest` is the
highest-fidelity test: real Tomcat on a random port, real RS256-signed
tokens via `SignedJwtTestSupport`, real signature verification end-to-end.

## What Trabuco does NOT generate

By design, Trabuco is a **resource server**, not an identity provider:

- No login forms, password handling, or MFA enrollment
- No token issuance, refresh-token rotation, or PKCE flows
- No user management UI or admin console
- No session cookies or browser-flow CSRF
- No social login (Google/GitHub OAuth2 client flows)
- No role/permission persistence — roles come from the JWT, Trabuco
  doesn't own the identity store

Use a hosted IdP (Keycloak / Auth0 / Okta / Cognito) for the producer
side. Trabuco only validates the tokens they issue.

## Architecture notes

- **Module boundaries.** Universal data (`IdentityClaims`,
  `AuthorityScope`, `AuthenticatedRequest`) lives in Model. Logic
  (extractors, propagators, scope helpers) lives in Shared. HTTP-
  specific filter chains live in API and AIAgent. This keeps Worker
  and EventConsumer (which can't import from API) able to read
  identity through `RequestContextHolder` without a web dependency.
- **Identity propagation.** API populates `RequestContextHolder` after
  JWT validation. For async boundaries — JobRunr jobs, Kafka events,
  webhooks — use `AuthenticatedRequest<T>` to carry identity in the
  payload, or `DefaultAuthContextPropagator` to inject claims into
  broker headers. Identity must NOT be re-validated downstream; the
  message broker is a trusted internal channel.
- **Virtual-thread safety.** `RequestContextHolder` uses `ThreadLocal`
  for Java 21 compatibility. When the runtime is Java 25+, this can
  be replaced with `ScopedValue` for inherited propagation across
  structured concurrency.
- **Tenant isolation in the AIAgent vector store (since 1.12).**
  `CallerIdentity.tenantId()` defaults to `keyHash()` — one logical
  tenant per credential, the safe default for single-tenant or
  per-partner deployments. JWT-mode deployments override this with
  the `tenant_id` claim (read by `ScopeEnforcer.jwtDerivedCaller`).
  Every chunk written through `DocumentIngestionService.ingest(...)`
  is server-side stamped with `metadata.tenant_id = caller.tenantId()`
  — caller-supplied `tenant_id` in the request body is overwritten
  and a WARN is logged. `VectorKnowledgeRetriever.retrieve(...)`
  filters every `SearchRequest` with
  `FilterExpressionBuilder.eq("tenant_id", caller.tenantId())`. The
  PGVector schema ships a btree expression index on
  `(metadata->>'tenant_id')` so the planner prunes by tenant before
  HNSW similarity search runs. Cross-tenant retrieval is blocked
  by construction.
- **Why dual chains, not a single conditional bean.** Spring Security
  on the classpath would auto-configure HTTP Basic on every endpoint if
  no `SecurityFilterChain` bean is provided. The open chain
  (`permitAllFilterChain` / `agentPermitAllFilterChain`) is the
  explicit replacement when the operator opts into the open stance via
  `trabuco.auth.enabled=false` — making "no JWT enforcement" a
  deliberate config rather than surprising default behavior. The
  `validateAuthDecisionMade` boot guardrail ensures neither chain is
  silently selected by omission.
