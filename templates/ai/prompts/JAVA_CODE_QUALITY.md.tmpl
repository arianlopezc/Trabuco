# Java Code Quality Specification

This document defines the code quality standards for testing. AI coding assistants MUST read this specification before generating code and self-review against it after generation.

---

## Self-Review Workflow

**CRITICAL**: After generating any Java code, you MUST:

1. **Read this entire specification** before writing code
2. **Generate the code** following these standards
3. **Self-review** against each section's checklist
4. **Refactor** any violations found
5. **Verify** the refactored code still compiles and passes tests

Do NOT submit code that violates these standards. Fix issues proactively.

---

## 1. Modern Java Idioms (Java 17+)

### 1.1 Streams Over Loops

**Use streams for filtering, mapping, and collecting operations.**

```java
// CORRECT: Declarative stream pipeline
List<String> activeUserNames = users.stream()
    .filter(User::isActive)
    .map(User::getName)
    .sorted()
    .toList();

// WRONG: Imperative loop
List<String> activeUserNames = new ArrayList<>();
for (User user : users) {
    if (user.isActive()) {
        activeUserNames.add(user.getName());
    }
}
Collections.sort(activeUserNames);
```

**When to use loops instead:**
- Complex control flow requiring `break` with conditions
- Performance-critical code on very small collections (< 10 elements)
- When mutable accumulation is significantly clearer

**Checklist:**
- [ ] No `for` loops that could be replaced by `stream().filter().map().collect()`
- [ ] Using primitive streams (`mapToInt`, `mapToLong`) to avoid boxing
- [ ] No side effects in stream operations (except terminal `forEach`)
- [ ] Using `.toList()` instead of `.collect(Collectors.toList())` for unmodifiable lists

### 1.2 Records for Data Classes

**Use records for DTOs, value objects, and simple data carriers.**

```java
// CORRECT: Record with validation
public record UserRequest(String name, String email) {
    public UserRequest {
        Objects.requireNonNull(name, "name must not be null");
        if (!email.contains("@")) {
            throw new IllegalArgumentException("Invalid email");
        }
    }
}

// WRONG: Verbose class with boilerplate
public class UserRequest {
    private final String name;
    private final String email;
    // constructor, getters, equals, hashCode, toString...
}
```

**Note**: This project uses Immutables for entities and DTOs. Use records for:
- Internal data transfer within a method/class
- Repository boundary objects (`*Record`, `*Document`)
- Simple local value objects

**Checklist:**
- [ ] Records used for simple data carriers without behavior
- [ ] Compact constructors used for validation when needed
- [ ] Mutable components (List, Map) defensively copied: `this.items = List.copyOf(items)`

### 1.3 Pattern Matching

**Use pattern matching to eliminate manual casting.**

```java
// CORRECT: Pattern matching for instanceof
if (event instanceof UserCreatedEvent e) {
    processUserCreated(e.userId(), e.name());
}

// CORRECT: Pattern matching in switch
String describe(Shape shape) {
    return switch (shape) {
        case Circle c -> "Circle with radius " + c.radius();
        case Rectangle r -> "Rectangle " + r.width() + "x" + r.height();
    };
}

// WRONG: Manual instanceof + cast
if (event instanceof UserCreatedEvent) {
    UserCreatedEvent e = (UserCreatedEvent) event;
    processUserCreated(e.userId(), e.name());
}
```

**Checklist:**
- [ ] No `instanceof` followed by explicit cast on next line
- [ ] Switch expressions used instead of switch statements where returning a value
- [ ] No unnecessary `default` case with sealed types (compiler checks exhaustiveness)

### 1.4 Optional Usage

**Use Optional only as return type for methods that may not have a result.**

```java
// CORRECT: Return Optional from finder methods
public Optional<User> findById(Long id) {
    return Optional.ofNullable(repository.get(id));
}

// CORRECT: Functional handling
String userName = findById(id)
    .map(User::getName)
    .orElse("Unknown");

// WRONG: isPresent + get pattern
if (userOpt.isPresent()) {
    User user = userOpt.get();  // Avoid this
}

// WRONG: Optional as parameter
public void process(Optional<Config> config) { }  // Never do this

// WRONG: Optional as field
private Optional<String> middleName;  // Never do this
```

**Checklist:**
- [ ] Optional used only as return types, never as parameters or fields
- [ ] No `isPresent()` + `get()` pattern - use `map`, `flatMap`, `orElse`, `orElseThrow`
- [ ] Collections never wrapped in Optional - return empty collection instead
- [ ] Using `orElseThrow()` with descriptive exception for required values

### 1.5 Immutability by Default

**Make classes immutable unless mutation is explicitly required.**

```java
// CORRECT: Immutable with defensive copy
public final class Team {
    private final String name;
    private final List<String> members;

    public Team(String name, List<String> members) {
        this.name = Objects.requireNonNull(name);
        this.members = List.copyOf(members);  // Defensive copy
    }

    public List<String> members() {
        return members;  // Already immutable
    }
}

// WRONG: Leaking mutable state
public List<String> getMembers() {
    return members;  // Caller can modify internal list
}
```

**Checklist:**
- [ ] All fields are `private final`
- [ ] No setter methods
- [ ] Mutable inputs defensively copied in constructor
- [ ] Using `List.of()`, `Set.of()`, `Map.of()` for immutable collections
- [ ] Using `java.time` classes instead of `Date`/`Calendar`

### 1.6 Collection Factory Methods

**Use immutable collection factory methods.**

```java
// CORRECT: Immutable collections
List<String> names = List.of("Alice", "Bob", "Charlie");
Set<Integer> numbers = Set.of(1, 2, 3);
Map<String, Integer> scores = Map.of("Alice", 100, "Bob", 95);

// CORRECT: When mutability needed
List<String> mutableList = new ArrayList<>(List.of("a", "b", "c"));

// WRONG: Verbose creation
List<String> names = new ArrayList<>();
names.add("Alice");
names.add("Bob");
```

**Checklist:**
- [ ] Using `List.of()`, `Set.of()`, `Map.of()` for constant collections
- [ ] Using `.toList()` at end of streams for unmodifiable result
- [ ] Not using `Arrays.asList()` (fixed-size but mutable)

### 1.7 Text Blocks

**Use text blocks for multi-line strings.**

```java
// CORRECT: Text block for SQL
String sql = """
    SELECT u.id, u.name, u.email
    FROM users u
    WHERE u.active = true
    ORDER BY u.name
    """;

// CORRECT: Text block for JSON
String json = """
    {
        "name": "%s",
        "email": "%s"
    }
    """.formatted(name, email);

// WRONG: String concatenation
String sql = "SELECT u.id, u.name, u.email\n" +
    "FROM users u\n" +
    "WHERE u.active = true";
```

**Checklist:**
- [ ] Text blocks used for SQL queries, JSON, HTML, and other multi-line strings
- [ ] No string concatenation with `\n` for multi-line strings
- [ ] Using `.formatted()` for string interpolation in text blocks

### 1.8 var Keyword

**Use var when the type is obvious from the right-hand side.**

```java
// CORRECT: Type obvious from constructor
var users = new ArrayList<User>();
var response = httpClient.send(request, BodyHandlers.ofString());

// CORRECT: Complex generic types
var entrySet = map.entrySet();

// WRONG: Type not obvious
var result = service.process();  // What type is result?

// WRONG: Numeric literals
var count = 0;      // Is this int, long, Integer?
var price = 19.99;  // Is this double, BigDecimal?
```

**Checklist:**
- [ ] var used only when type is clear from right-hand side
- [ ] Not using var with numeric literals
- [ ] Not using var when it hurts readability
- [ ] Choosing descriptive variable names when using var

### 1.9 Method References

**Use method references when clearer than lambdas.**

```java
// CORRECT: Method reference
users.stream()
    .map(User::getName)
    .filter(Objects::nonNull)
    .forEach(System.out::println);

// CORRECT: Lambda when logic is complex
users.stream()
    .filter(u -> u.getAge() > 18 && u.isActive())
    .toList();

// WRONG: Lambda when method reference works
users.stream()
    .map(user -> user.getName())  // Use User::getName
    .toList();
```

**Checklist:**
- [ ] Method references used for simple method calls
- [ ] Lambdas used when logic involves multiple operations or external variables

### 1.10 Try-with-Resources

**Always use try-with-resources for AutoCloseable resources.**

```java
// CORRECT: Try-with-resources
try (var connection = dataSource.getConnection();
     var statement = connection.prepareStatement(sql);
     var resultSet = statement.executeQuery()) {
    while (resultSet.next()) {
        // process
    }
}

// WRONG: Manual resource management
Connection conn = null;
try {
    conn = dataSource.getConnection();
    // use connection
} finally {
    if (conn != null) conn.close();
}
```

**Checklist:**
- [ ] All `Connection`, `InputStream`, `OutputStream`, etc. in try-with-resources
- [ ] No manual close() calls in finally blocks

---

## 2. Method Complexity

### 2.1 Method Length

**Methods should be short and focused. Maximum 20-30 lines of logic.**

```java
// CORRECT: Short, focused method with helpers
public Order processOrder(OrderRequest request) {
    validateRequest(request);
    var items = resolveItems(request.itemIds());
    var pricing = calculatePricing(items, request.discountCode());
    var order = createOrder(request.customerId(), items, pricing);
    notifyCustomer(order);
    return order;
}

private void validateRequest(OrderRequest request) { /* ... */ }
private List<Item> resolveItems(List<Long> itemIds) { /* ... */ }
private Pricing calculatePricing(List<Item> items, String discountCode) { /* ... */ }
private Order createOrder(Long customerId, List<Item> items, Pricing pricing) { /* ... */ }
private void notifyCustomer(Order order) { /* ... */ }

// WRONG: Long method doing everything
public Order processOrder(OrderRequest request) {
    // 100+ lines of validation, item lookup, pricing calculation,
    // order creation, notification, logging, etc.
}
```

**Checklist:**
- [ ] No method exceeds 30 lines of logic (excluding blank lines and braces)
- [ ] Each method does ONE thing
- [ ] Complex logic extracted to private helper methods
- [ ] Method name describes what it does, not how

### 2.2 Cyclomatic Complexity

**Keep cyclomatic complexity low (ideally < 5, maximum 10).**

```java
// CORRECT: Low complexity with early returns
public String getStatus(User user) {
    if (user == null) return "UNKNOWN";
    if (!user.isActive()) return "INACTIVE";
    if (user.isAdmin()) return "ADMIN";
    return "ACTIVE";
}

// CORRECT: Extract conditions to methods
public boolean canAccessResource(User user, Resource resource) {
    return isAuthenticated(user)
        && hasPermission(user, resource)
        && isResourceAvailable(resource);
}

// WRONG: Nested conditionals
public String getStatus(User user) {
    if (user != null) {
        if (user.isActive()) {
            if (user.isAdmin()) {
                return "ADMIN";
            } else {
                return "ACTIVE";
            }
        } else {
            return "INACTIVE";
        }
    } else {
        return "UNKNOWN";
    }
}
```

**Checklist:**
- [ ] No deeply nested conditionals (max 2 levels)
- [ ] Using early returns to reduce nesting
- [ ] Complex boolean expressions extracted to descriptive methods
- [ ] Switch/case replaced with polymorphism or pattern matching where appropriate

### 2.3 Parameter Count

**Methods should have few parameters (ideally <= 3, maximum 5).**

```java
// CORRECT: Parameter object
public Order createOrder(OrderRequest request) {
    // request contains customerId, items, shippingAddress, paymentMethod, discountCode
}

// CORRECT: Builder for complex construction
var order = Order.builder()
    .customerId(customerId)
    .items(items)
    .shippingAddress(address)
    .paymentMethod(payment)
    .build();

// WRONG: Too many parameters
public Order createOrder(Long customerId, List<Item> items,
    Address shippingAddress, PaymentMethod payment, String discountCode,
    boolean expressShipping, String giftMessage) { }
```

**Checklist:**
- [ ] No method has more than 5 parameters
- [ ] Related parameters grouped into objects
- [ ] Using builders for complex object construction

---

## 3. Naming Conventions

### 3.1 Clear, Descriptive Names

```java
// CORRECT: Descriptive names
public List<User> findActiveUsersByDepartment(String departmentId) { }
private boolean isEligibleForDiscount(Order order) { }
private void sendWelcomeEmail(User user) { }

// WRONG: Abbreviated or unclear names
public List<User> getUsrs(String dId) { }
private boolean check(Order o) { }
private void send(User u) { }
```

### 3.2 Naming Patterns

| Element | Pattern | Example |
|---------|---------|---------|
| Boolean methods | `is*`, `has*`, `can*`, `should*` | `isActive()`, `hasPermission()` |
| Finder methods | `find*By*`, `get*` | `findUserById()`, `getActiveUsers()` |
| Predicates | Describe the condition | `isValidEmail`, `hasEnoughStock` |
| Collections | Plural nouns | `users`, `orderItems`, `activeAccounts` |
| Counts | `*Count` or `numberOf*` | `orderCount`, `numberOfItems` |

**Checklist:**
- [ ] Method names are verbs or verb phrases
- [ ] Variable names are nouns or noun phrases
- [ ] Boolean names read naturally in `if` statements
- [ ] No abbreviations except universally known ones (id, url, http)
- [ ] No single-letter names except in tiny scopes (lambdas, loops)

---

## 4. Error Handling

### 4.1 Exception Strategy

```java
// CORRECT: Specific exception with context
public User findUserOrThrow(Long id) {
    return userRepository.findById(id)
        .orElseThrow(() -> new UserNotFoundException("User not found: " + id));
}

// CORRECT: Let framework handle common exceptions
@GetMapping("/{id}")
public ImmutableUserResponse getUser(@PathVariable Long id) {
    return userService.findById(id);  // GlobalExceptionHandler handles 404
}

// WRONG: Catching generic Exception
try {
    process();
} catch (Exception e) {  // Too broad
    log.error("Error", e);
}

// WRONG: Empty catch block
try {
    process();
} catch (IOException e) {
    // Silently swallowed
}
```

**Checklist:**
- [ ] No `catch (Exception e)` unless re-throwing or at top level
- [ ] No empty catch blocks
- [ ] Exception messages include relevant context (IDs, parameters)
- [ ] Not catching exceptions handled by GlobalExceptionHandler

### 4.2 Validation

```java
// CORRECT: Fail fast with Objects.requireNonNull
public UserService(UserRepository repository, EmailService emailService) {
    this.repository = Objects.requireNonNull(repository, "repository");
    this.emailService = Objects.requireNonNull(emailService, "emailService");
}

// CORRECT: Bean validation on DTOs
public record CreateUserRequest(
    @NotBlank String name,
    @Email String email,
    @Min(18) int age
) { }

// WRONG: Null checks scattered throughout code
public void process(Data data) {
    if (data != null) {
        if (data.getValue() != null) {
            // ...
        }
    }
}
```

**Checklist:**
- [ ] Constructor parameters validated with `Objects.requireNonNull`
- [ ] DTOs use Bean Validation annotations (`@NotNull`, `@NotBlank`, etc.)
- [ ] No defensive null checks for values that should never be null
- [ ] Validation happens at system boundaries, not throughout codebase

---

## 5. Architecture Compliance

### 5.1 Module Boundaries

| Module | Depends On | Never Depends On |
|--------|------------|------------------|
| Model | (none) | Everything else |
| SQLDatastore | Model | Shared, API, Worker, EventConsumer |
| NoSQLDatastore | Model | Shared, API, Worker, EventConsumer |
| Shared | Model, SQLDatastore, NoSQLDatastore | API, Worker, EventConsumer |
| API | Model, Shared | Worker, EventConsumer |
| Worker | Model, Shared, Jobs | API, EventConsumer |
| EventConsumer | Model, Shared, Events | API, Worker |

**Checklist:**
- [ ] No imports from disallowed modules
- [ ] Services in Shared, not in API/Worker/EventConsumer
- [ ] Repository interfaces in Datastore modules only
- [ ] DTOs in Model module only

### 5.2 Persistence Boundaries

```java
// CORRECT: Convert at repository boundary
public Optional<ImmutableUser> findById(Long id) {
    return repository.findById(id)
        .map(this::toImmutable);
}

private ImmutableUser toImmutable(UserRecord record) {
    return ImmutableUser.builder()
        .id(record.id())
        .name(record.name())
        .build();
}

// WRONG: Exposing record outside service
public UserRecord findById(Long id) {  // Don't expose Record
    return repository.findById(id).orElseThrow();
}
```

**Checklist:**
- [ ] `*Record` and `*Document` types never exposed outside service layer
- [ ] Conversion to `Immutable*` happens immediately after repository call
- [ ] Repository methods return records, service methods return Immutables

---

## 6. Spring Best Practices

### 6.1 Dependency Injection

```java
// CORRECT: Constructor injection with final fields
@Service
@RequiredArgsConstructor
public class UserService {
    private final UserRepository userRepository;
    private final EmailService emailService;
}

// WRONG: Field injection
@Service
public class UserService {
    @Autowired
    private UserRepository userRepository;  // Not final, not testable
}
```

### 6.2 Transaction Management

```java
// CORRECT: @Transactional on service methods
@Transactional
public void transferFunds(Long fromId, Long toId, BigDecimal amount) {
    // multiple repository operations
}

// CORRECT: Read-only for queries
@Transactional(readOnly = true)
public List<ImmutableUser> findActiveUsers() {
    return repository.findByActive(true).stream()
        .map(this::toImmutable)
        .toList();
}

// WRONG: @Transactional on private methods (doesn't work)
@Transactional  // Ignored!
private void updateInternal() { }
```

**Checklist:**
- [ ] Constructor injection used (no `@Autowired` on fields)
- [ ] All fields in services are `private final`
- [ ] `@Transactional` on public methods only
- [ ] `@Transactional(readOnly = true)` for read-only operations
- [ ] `@CircuitBreaker` on methods calling external services

---

## 7. Final Review Checklist

Before submitting any code, verify:

### Code Quality
- [ ] All methods under 30 lines
- [ ] No nested conditionals deeper than 2 levels
- [ ] No method with more than 5 parameters
- [ ] All names are clear and descriptive

### Modern Java
- [ ] Streams used instead of loops where appropriate
- [ ] Records used for simple data classes
- [ ] Pattern matching used instead of instanceof + cast
- [ ] Optional used correctly (return types only)
- [ ] Immutable collections where appropriate
- [ ] Text blocks for multi-line strings
- [ ] Try-with-resources for all AutoCloseable

### Architecture
- [ ] Module dependencies respected
- [ ] Repository records converted at service boundary
- [ ] Services use constructor injection
- [ ] DTOs are Immutables in Model module

### Testing
- [ ] Unit tests for business logic
- [ ] Integration tests for repository operations
- [ ] Tests follow Arrange-Act-Assert pattern

---

_This specification is loaded by AI coding assistants. Violations should be fixed before code submission._
