# Authentication

Trabuco generates production-ready OIDC Resource Server scaffolding by
default whenever the **API** or **AIAgent** module is selected. The code
ships **dormant**: every generated project includes the JWT validation
chain, the scope-mapping converter, RFC 7807 ProblemDetail handlers, and
test utilities — but a permit-all SecurityFilterChain is the active bean
until you flip the runtime gate.

This document covers what gets generated, how to enable it, how to wire
each major IdP, and how to test code that depends on identity.

## Enabling auth

Auth is gated at runtime by a single property:

```yaml
trabuco:
  auth:
    enabled: ${TRABUCO_AUTH_ENABLED:false}   # default
```

To turn the JWT validation chain on:

1. Set `trabuco.auth.enabled=true` (or export `TRABUCO_AUTH_ENABLED=true`).
2. Set `OIDC_ISSUER_URI` to your IdP's discovery endpoint (see provider
   recipes below).

When `trabuco.auth.enabled` is missing or `false`, the
`SecurityConfig.permitAllFilterChain` bean is the active filter chain
and no authentication is enforced. When set to `true`, the
`oauth2FilterChain` bean takes over and validates incoming JWTs against
the configured issuer. The two chains are mutually exclusive
(`@ConditionalOnProperty` pair) so exactly one is wired at any time.

Both `SecurityConfig` (API) and `AgentSecurityConfig` (AIAgent) follow
the same dual-chain pattern.

## Why dormant by default

The auth code is always present so it's discoverable in source — you can
read it, customize the claim extractor, drop in scope checks on
controllers — without re-running the generator or installing a CLI flag.
But until you have an IdP configured and have decided to enforce auth,
the runtime stays open. This avoids a class of mistakes where projects
boot with a half-configured resource server and either:

- fail mysteriously at startup because the issuer URI isn't set, or
- silently accept unsigned tokens because a misconfiguration disabled
  signature verification.

When the gate is off, both failure modes are physically impossible: the
JWT chain bean isn't created, no JwtDecoder is wired, the issuer URI is
never read.

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
```

Okta emits standard `scp` (array form). The default extractor handles
both `scope` and `scp` automatically — no customization needed.

### AWS Cognito

```bash
export TRABUCO_AUTH_ENABLED=true
export OIDC_ISSUER_URI=https://cognito-idp.{region}.amazonaws.com/{userPoolId}
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
```

The default extractor reads the standard `scope` (or `scp`) claim. If
the provider emits a non-standard layout, register a custom
`JwtClaimsExtractor` bean with `@Primary`.

## Migration from API-key auth (AIAgent)

The AIAgent module ships with two coexisting auth mechanisms — the
legacy tier-based `ApiKeyAuthFilter` (governed by
`app.aiagent.api-key.enabled`) and the new JWT chain (governed by
`trabuco.auth.enabled`). The matrix:

| `app.aiagent.api-key.enabled` | `trabuco.auth.enabled` | Behavior |
|-----------------------------|-----------------------|----------|
| `true` (default)            | `false` (default)     | Legacy API-key path active, JWT dormant. Pre-auth-scaffolding behavior preserved. |
| `true`                      | `true`                | Hybrid — both filters run, either credential type accepted. |
| `false`                     | `true`                | JWT-only. Recommended target state. |
| `false`                     | `false`               | No auth enforced. (Don't run this in prod.) |

Incremental migration path:

1. **Project is generated with both paths already wired** — no
   regeneration needed.
2. **Adopt JWT for new endpoints.** Use
   `@PreAuthorize("hasAuthority('SCOPE_agent:read')")` instead of
   `@RequireScope`. Set `trabuco.auth.enabled=true` and configure the
   issuer URI.
3. **Migrate existing endpoints incrementally.** Replace
   `@RequireScope("public")` with the appropriate
   `@PreAuthorize("hasAuthority('SCOPE_*')")` annotation.
4. **Disable the legacy filter when ready.** Set
   `app.aiagent.api-key.enabled=false` in `application.yml`. Remove the
   API-key handling code if you want to clean up.

The two systems have different semantics: tiers are hierarchical
(`anonymous < public < partner`), JWT scopes are flat presence checks.
Don't try to bridge them — pick one per endpoint.

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

Tests that exercise the JWT filter chain must explicitly opt into auth
via a property — the dormant default would otherwise route every
request through the permit-all chain and short-circuit the 401 path:

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
- **Why dual chains, not a single conditional bean.** Spring Security
  on the classpath would auto-configure HTTP Basic on every endpoint if
  no `SecurityFilterChain` bean is provided. The permit-all chain is
  the explicit "open" stance that takes its place when auth is
  dormant — making "no enforcement" a deliberate config rather than
  surprising default behavior.
