# Acme Shop — Spring Boot 2.7 monolith fixture

This is a test fixture for the Trabuco 1.10 migration. It is intentionally
shaped like a typical small Spring Boot 2.7 monolith with patterns that
exercise specific blocker codes in the Trabuco migration:

- **`javax.*` imports** (JPA, validation) — exercises the Spring Boot 3
  jakarta migration that the model/datastore specialists handle.
- **FK constraints** (Order → Customer, OrderLine → Order) — exercises
  the `FK_REQUIRED` blocker in the datastore specialist.
- **OFFSET pagination** (`PageRequest.of`) — exercises
  `OFFSET_PAGINATION_INCOMPATIBLE` in the datastore specialist.
- **Field injection** (`@Autowired` on fields in `CustomerService`,
  `OrderService`, `DailyOrderReportJob`) — exercises the constructor-
  injection refactor in the shared specialist.
- **Bespoke `ErrorResponse`** — exercises `LEGACY_ERROR_FORMAT_REQUIRED`
  in the API specialist.
- **`spring.jpa.hibernate.ddl-auto=update`** — exercises the
  `SCHEMA_DRIFT_RISK` blocker in the datastore specialist.
- **Hardcoded password in `application-prod.properties`** —
  exercises `SECRET_IN_SOURCE` in the assessor.
- **`@Scheduled` job** (`DailyOrderReportJob`) — exercises the worker
  specialist's `@Scheduled → JobRunr` migration.
- **`@SpringBootTest`-only tests** — exercises the test specialist's
  ADAPT decisions.
- **GitHub Actions CI workflow** — exercises the deployment specialist.

The fixture is **not** a real product; it's intentionally minimal yet
complete enough to exercise the migration end-to-end. Run via:

```
trabuco migrate run /Users/arianlc/Documents/Work/Trabuco/testdata/migration-fixtures/spring-boot-27-monolith
```

(Or, more typically, copy it elsewhere first since the migration mutates
the working tree.)
