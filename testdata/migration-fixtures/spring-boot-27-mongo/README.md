# Acme Blog — Spring Boot 2.7 + MongoDB fixture

NoSQL persistence path for the Trabuco 1.10 migration. Exercises:

- `@Document` entities (Post, User) — Phase 0 NoSQLDatastore detection
- `MongoRepository` with derived query methods + `Pageable` offset
  pagination — Phase 3 NoSQLDatastore + `OFFSET_PAGINATION_INCOMPATIBLE`
- `@Indexed` and unique-index annotations
- Field injection (`@Autowired` on fields) — Phase 4 shared specialist
- `@RestController` with bespoke (or absent) error handling — Phase 5 API
- No CI/CD, no scheduled jobs, no listeners — Phases 6/7/10
  `not_applicable`

Use as a target for `trabuco migrate run`:

```
trabuco migrate run /path/to/spring-boot-27-mongo
```
