# Authentication

Trabuco generates production-ready OIDC Resource Server scaffolding when
you pass `--auth=oidc` to `trabuco init`. This document covers what gets
generated, how to wire each major IdP, and how to test code that depends
on identity.

## What gets generated

When `--auth=oidc` is set, Trabuco emits the following files (all gated
by the same flag — default `--auth=none` produces zero auth code):

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

**API module — HTTP filter chain**
- `api/config/security/SecurityConfig.java` — stateless JWT filter
  chain, `@EnableMethodSecurity`, permitAll for actuator + swagger.
- `api/config/security/JwtAuthenticationConverter.java` — `Jwt` →
  `JwtAuthenticationToken`, populates `RequestContextHolder`.
- `api/config/security/AuthProblemDetailHandler.java` — RFC 7807
  `application/problem+json` for 401/403.
- `api/config/security/OpenApiSecurityConfig.java` — adds `bearerAuth`
  scheme to Swagger UI's "Authorize" button.
- `api/test/security/SecurityIntegrationTest.java` — boot the full
  context, verify 401 + ProblemDetail without a token.

**AIAgent module — JWT path coexisting with API-key path**
- `aiagent/config/security/AgentSecurityConfig.java` — JWT filter chain
  for the AIAgent Spring Boot app (mirrors API's).
- `aiagent/config/security/JwtAuthenticationConverter.java`,
  `AuthProblemDetailHandler.java` — duplicated from API since AIAgent
  doesn't depend on API.
- `aiagent/security/ApiKeyAuthFilter.java` becomes
  `@ConditionalOnProperty(matchIfMissing=true)` so the legacy
  tier-based path stays enabled by default but can be turned off
  with `app.aiagent.api-key.enabled=false`.

## Provider configuration

Auth is configured via the standard Spring Boot property
`spring.security.oauth2.resourceserver.jwt.issuer-uri`, which in the
generated `application.yml` is sourced from the `OIDC_ISSUER_URI`
environment variable. Below are the values for each major IdP.

### Keycloak

```bash
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
export OIDC_ISSUER_URI=https://YOUR_DOMAIN.okta.com/oauth2/default
```

Okta emits standard `scp` (array form). The default extractor handles
both `scope` and `scp` automatically — no customization needed.

### AWS Cognito

```bash
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
export OIDC_ISSUER_URI=https://your-idp.example.com
```

The default extractor reads the standard `scope` (or `scp`) claim. If
the provider emits a non-standard layout, register a custom
`JwtClaimsExtractor` bean with `@Primary`.

## Migration from API-key auth (AIAgent)

When `--auth=oidc` is set on a project that already uses AIAgent's
legacy tier-based API-key auth (`@RequireScope("public")` /
`CallerContext`), both paths run side-by-side. The migration is
incremental:

1. **Generate or regenerate with `--auth=oidc`.** Existing
   `@RequireScope` annotations continue working — `ApiKeyAuthFilter`
   is `@ConditionalOnProperty(matchIfMissing=true)` so it stays on
   by default.
2. **Adopt JWT for new endpoints.** Use
   `@PreAuthorize("hasAuthority('SCOPE_agent:read')")` instead of
   `@RequireScope`.
3. **Migrate existing endpoints incrementally.** Replace
   `@RequireScope("public")` with the appropriate
   `@PreAuthorize("hasAuthority('SCOPE_*')")` annotation.
4. **Disable the legacy filter when ready.** Set
   `app.aiagent.api-key.enabled=false` in `application.yml`.
   Remove the API-key handling code if you want to clean up.

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

The generated `SecurityIntegrationTest` (in `API/test/security/`) is a
working example that uses `@MockBean JwtDecoder` to satisfy the
resource-server starter without making real network calls.

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
