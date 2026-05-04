package addgen

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
)

// TestOpts is the input contract for `trabuco add test`.
type TestOpts struct {
	// Target is the class under test, e.g. "OrderService". The
	// generated file is {Target}Test.java for unit/repository tests
	// and {Target}IT.java for integration tests (Surefire vs Failsafe
	// convention).
	Target string

	// Type selects the test shape: "unit" (default), "integration",
	// or "repository". Repository requires SQLDatastore or
	// NoSQLDatastore; the right Spring Boot test slice + Testcontainer
	// is wired automatically.
	Type string

	// Module the test belongs to. Drives the file path and (for
	// repository tests) the slice/container choice.
	Module string

	// Subpackage allows the caller to force a specific subpackage
	// under {moduleLower}. Empty means infer from Target suffix
	// (Service → service, Controller → controller, Repository →
	// repository, Handler → handler), or empty if no suffix matches.
	Subpackage string
}

const (
	TestTypeUnit        = "unit"
	TestTypeIntegration = "integration"
	TestTypeRepository  = "repository"
)

// GenerateTest emits a new {Target}{Test,IT}.java skeleton in the
// correct module test source root. The body imports the right JUnit
// 5 + Mockito / SpringBoot / Testcontainers bits and leaves a single
// stub @Test that fails until the agent fills in real assertions.
//
// Addition-only: refuses to clobber an existing file at the target
// path. The agent removes the file before regenerating if needed.
func GenerateTest(ctx *Context, opts TestOpts) (*Result, error) {
	if ctx == nil {
		return nil, fmt.Errorf("nil context")
	}
	if strings.TrimSpace(opts.Target) == "" {
		return nil, fmt.Errorf("target class name is required (positional argument)")
	}
	target := opts.Target
	if !isValidJavaIdentifier(target) {
		return nil, fmt.Errorf("target %q is not a valid Java class name", target)
	}

	if opts.Type == "" {
		opts.Type = TestTypeUnit
	}
	switch opts.Type {
	case TestTypeUnit, TestTypeIntegration, TestTypeRepository:
	default:
		return nil, fmt.Errorf("--type must be one of: unit, integration, repository (got %q)", opts.Type)
	}

	if opts.Module == "" {
		return nil, fmt.Errorf("--module is required (e.g. Shared, API, SQLDatastore, Worker)")
	}
	if !ctx.HasModule(opts.Module) {
		return nil, fmt.Errorf("project does not have module %s", opts.Module)
	}

	// Repository tests are a Spring Boot test-slice. They only make
	// sense for the datastore modules.
	if opts.Type == TestTypeRepository {
		if opts.Module != config.ModuleSQLDatastore && opts.Module != config.ModuleNoSQLDatastore {
			return nil, fmt.Errorf("--type=repository requires --module=SQLDatastore or NoSQLDatastore (got %s)", opts.Module)
		}
	}

	subpkg := opts.Subpackage
	if subpkg == "" {
		subpkg = inferTestSubpackage(target, opts.Type)
	}

	javaPkg := ctx.JavaPackage(opts.Module, subpkg)
	fileName := target + "Test.java"
	if opts.Type == TestTypeIntegration {
		fileName = target + "IT.java"
	}
	relPath := filepath.Join(ctx.JavaSrcTest(opts.Module, subpkg), fileName)

	content, err := buildTestContent(ctx, opts, target, javaPkg)
	if err != nil {
		return nil, err
	}

	result := &Result{}
	if err := ctx.emitFile(relPath, content, result); err != nil {
		return nil, err
	}

	result.NextSteps = []string{
		fmt.Sprintf("Replace the placeholder @Test in %s with real assertions for %s.", relPath, target),
		"Run `mvn -pl " + opts.Module + " test` (or `verify` for IT) to confirm the test wires up.",
	}
	if opts.Type == TestTypeRepository && opts.Module == config.ModuleSQLDatastore {
		result.NextSteps = append(result.NextSteps,
			"Repository tests use Testcontainers — make sure Docker is running before `mvn verify`.",
		)
	}
	return result, nil
}

// inferTestSubpackage maps a Target's suffix to the conventional
// subpackage. Mirrors how the rest of the project lays things out
// (controller/, service/, repository/, handler/). Returns "" when
// nothing matches — the test then lands directly under the module's
// root package, which is fine for unit tests of utility classes.
func inferTestSubpackage(target, testType string) string {
	switch {
	case strings.HasSuffix(target, "Controller"):
		return "controller"
	case strings.HasSuffix(target, "Service"):
		return "service"
	case strings.HasSuffix(target, "Repository"):
		return "repository"
	case strings.HasSuffix(target, "Handler"):
		return "handler"
	case strings.HasSuffix(target, "Listener"):
		return "listener"
	case strings.HasSuffix(target, "Config") || strings.HasSuffix(target, "Configuration"):
		return "config"
	}
	if testType == TestTypeRepository {
		return "repository"
	}
	return ""
}

// buildTestContent renders the test class body. Three flavors:
// unit (Mockito), integration (SpringBootTest), repository
// (DataJdbcTest / DataMongoTest with Testcontainers).
func buildTestContent(ctx *Context, opts TestOpts, target, javaPkg string) (string, error) {
	switch opts.Type {
	case TestTypeUnit:
		return renderUnitTest(target, javaPkg), nil
	case TestTypeIntegration:
		return renderIntegrationTest(target, javaPkg), nil
	case TestTypeRepository:
		if opts.Module == config.ModuleSQLDatastore {
			return renderSQLRepositoryTest(target, javaPkg, ctx.Database), nil
		}
		return renderMongoRepositoryTest(target, javaPkg), nil
	}
	return "", fmt.Errorf("unsupported type %q", opts.Type)
}

func renderUnitTest(target, pkg string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s;\n\n", pkg)
	b.WriteString("import org.junit.jupiter.api.Test;\n")
	b.WriteString("import org.junit.jupiter.api.extension.ExtendWith;\n")
	b.WriteString("import org.mockito.junit.jupiter.MockitoExtension;\n\n")
	b.WriteString("import static org.junit.jupiter.api.Assertions.*;\n\n")
	b.WriteString("@ExtendWith(MockitoExtension.class)\n")
	fmt.Fprintf(&b, "class %sTest {\n\n", target)
	b.WriteString("    @Test\n")
	b.WriteString("    void TODO_addAssertions() {\n")
	fmt.Fprintf(&b, "        // TODO: replace this stub with real tests for %s\n", target)
	b.WriteString("        fail(\"not implemented\");\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

func renderIntegrationTest(target, pkg string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s;\n\n", pkg)
	b.WriteString("import org.junit.jupiter.api.Test;\n")
	b.WriteString("import org.springframework.boot.test.context.SpringBootTest;\n\n")
	b.WriteString("import static org.junit.jupiter.api.Assertions.*;\n\n")
	b.WriteString("@SpringBootTest\n")
	fmt.Fprintf(&b, "class %sIT {\n\n", target)
	b.WriteString("    @Test\n")
	b.WriteString("    void TODO_addAssertions() {\n")
	fmt.Fprintf(&b, "        // TODO: integration test for %s. Wire fixtures, exercise public APIs, assert on outputs.\n", target)
	b.WriteString("        fail(\"not implemented\");\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

func renderSQLRepositoryTest(target, pkg, database string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s;\n\n", pkg)
	b.WriteString("import org.junit.jupiter.api.Test;\n")
	b.WriteString("import org.springframework.boot.test.autoconfigure.data.jdbc.DataJdbcTest;\n")
	b.WriteString("import org.springframework.boot.test.autoconfigure.jdbc.AutoConfigureTestDatabase;\n")
	useContainer := database == config.DatabasePostgreSQL || database == config.DatabaseMySQL
	if useContainer {
		b.WriteString("import org.springframework.boot.testcontainers.service.connection.ServiceConnection;\n")
		b.WriteString("import org.testcontainers.junit.jupiter.Container;\n")
		b.WriteString("import org.testcontainers.junit.jupiter.Testcontainers;\n")
		switch database {
		case config.DatabasePostgreSQL:
			b.WriteString("import org.testcontainers.containers.PostgreSQLContainer;\n")
		case config.DatabaseMySQL:
			b.WriteString("import org.testcontainers.containers.MySQLContainer;\n")
		}
	}
	b.WriteString("\n")
	b.WriteString("import static org.junit.jupiter.api.Assertions.*;\n\n")
	b.WriteString("@DataJdbcTest\n")
	b.WriteString("@AutoConfigureTestDatabase(replace = AutoConfigureTestDatabase.Replace.NONE)\n")
	if useContainer {
		b.WriteString("@Testcontainers(disabledWithoutDocker = true)\n")
	}
	fmt.Fprintf(&b, "class %sTest {\n\n", target)
	if useContainer {
		switch database {
		case config.DatabasePostgreSQL:
			b.WriteString("    @Container @ServiceConnection\n")
			b.WriteString("    static PostgreSQLContainer<?> postgres = new PostgreSQLContainer<>(\"postgres:15-alpine\");\n\n")
		case config.DatabaseMySQL:
			b.WriteString("    @Container @ServiceConnection\n")
			b.WriteString("    static MySQLContainer<?> mysql = new MySQLContainer<>(\"mysql:8.0\");\n\n")
		}
	}
	b.WriteString("    @Test\n")
	b.WriteString("    void TODO_addAssertions() {\n")
	fmt.Fprintf(&b, "        // TODO: inject %s and exercise its query methods.\n", target)
	b.WriteString("        fail(\"not implemented\");\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

func renderMongoRepositoryTest(target, pkg string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "package %s;\n\n", pkg)
	b.WriteString("import org.junit.jupiter.api.Test;\n")
	b.WriteString("import org.springframework.boot.test.autoconfigure.data.mongo.DataMongoTest;\n")
	b.WriteString("import org.springframework.boot.testcontainers.service.connection.ServiceConnection;\n")
	b.WriteString("import org.testcontainers.containers.MongoDBContainer;\n")
	b.WriteString("import org.testcontainers.junit.jupiter.Container;\n")
	b.WriteString("import org.testcontainers.junit.jupiter.Testcontainers;\n\n")
	b.WriteString("import static org.junit.jupiter.api.Assertions.*;\n\n")
	b.WriteString("@DataMongoTest\n")
	b.WriteString("@Testcontainers(disabledWithoutDocker = true)\n")
	fmt.Fprintf(&b, "class %sTest {\n\n", target)
	b.WriteString("    @Container @ServiceConnection\n")
	b.WriteString("    static MongoDBContainer mongo = new MongoDBContainer(\"mongo:7.0\");\n\n")
	b.WriteString("    @Test\n")
	b.WriteString("    void TODO_addAssertions() {\n")
	fmt.Fprintf(&b, "        // TODO: inject %s and exercise its document operations.\n", target)
	b.WriteString("        fail(\"not implemented\");\n")
	b.WriteString("    }\n")
	b.WriteString("}\n")
	return b.String()
}

// isValidJavaIdentifier accepts conventional Java class names: starts
// with [A-Za-z_], rest is [A-Za-z0-9_]. Stricter than the JLS but
// matches every realistic name a user would pass.
func isValidJavaIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') && r != '_' {
				return false
			}
			continue
		}
		if !(r >= 'A' && r <= 'Z') && !(r >= 'a' && r <= 'z') && !(r >= '0' && r <= '9') && r != '_' {
			return false
		}
	}
	return true
}
