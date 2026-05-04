package addgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInferTestSubpackage(t *testing.T) {
	cases := []struct {
		target, testType, want string
	}{
		{"OrderService", "unit", "service"},
		{"OrderController", "unit", "controller"},
		{"OrderRepository", "unit", "repository"},
		{"OrderHandler", "unit", "handler"},
		{"OrderListener", "unit", "listener"},
		{"AppConfig", "unit", "config"},
		{"DatabaseConfiguration", "unit", "config"},
		{"NoSuffix", "unit", ""},
		{"NoSuffix", "repository", "repository"}, // repository tests force "repository" subpkg
		{"OrderService", "repository", "service"}, // explicit suffix wins
	}
	for _, tc := range cases {
		got := inferTestSubpackage(tc.target, tc.testType)
		if got != tc.want {
			t.Errorf("inferTestSubpackage(%q, %q) = %q, want %q", tc.target, tc.testType, got, tc.want)
		}
	}
}

func TestIsValidJavaIdentifier(t *testing.T) {
	good := []string{"Foo", "foo", "_x", "Foo123", "OrderService", "ABC"}
	bad := []string{"", "1Foo", "foo-bar", "foo bar", "foo.Bar", "foo$bar", "中文"}
	for _, s := range good {
		if !isValidJavaIdentifier(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
	for _, s := range bad {
		if isValidJavaIdentifier(s) {
			t.Errorf("expected %q to be invalid", s)
		}
	}
}

func TestGenerateTest(t *testing.T) {
	cases := []struct {
		name            string
		seedFiles       map[string]string
		opts            TestOpts
		wantPathSuffix  string
		wantContains    []string
		wantNotContains []string
		wantErrContains string
	}{
		{
			name:           "unit test for service",
			seedFiles:      apiSqlPgFixture(),
			opts:           TestOpts{Target: "OrderService", Module: "Shared"},
			wantPathSuffix: "Shared/src/test/java/com/example/demo/shared/service/OrderServiceTest.java",
			wantContains: []string{
				"package com.example.demo.shared.service;",
				"import org.mockito.junit.jupiter.MockitoExtension;",
				"@ExtendWith(MockitoExtension.class)",
				"class OrderServiceTest {",
				"fail(\"not implemented\");",
			},
			wantNotContains: []string{"@SpringBootTest", "@DataJdbcTest", "Testcontainers"},
		},
		{
			name:           "integration test for controller",
			seedFiles:      apiSqlPgFixture(),
			opts:           TestOpts{Target: "OrderController", Module: "API", Type: "integration"},
			wantPathSuffix: "API/src/test/java/com/example/demo/api/controller/OrderControllerIT.java",
			wantContains: []string{
				"package com.example.demo.api.controller;",
				"import org.springframework.boot.test.context.SpringBootTest;",
				"@SpringBootTest",
				"class OrderControllerIT {",
			},
			wantNotContains: []string{"MockitoExtension", "DataJdbcTest"},
		},
		{
			name:           "repository test on postgres uses PostgreSQLContainer",
			seedFiles:      apiSqlPgFixture(),
			opts:           TestOpts{Target: "OrderRepository", Module: "SQLDatastore", Type: "repository"},
			wantPathSuffix: "SQLDatastore/src/test/java/com/example/demo/sqldatastore/repository/OrderRepositoryTest.java",
			wantContains: []string{
				"@DataJdbcTest",
				"@AutoConfigureTestDatabase(replace = AutoConfigureTestDatabase.Replace.NONE)",
				"@Testcontainers(disabledWithoutDocker = true)",
				"PostgreSQLContainer",
				"postgres:15-alpine",
			},
			wantNotContains: []string{"MySQLContainer", "MongoDBContainer"},
		},
		{
			name: "repository test on mysql uses MySQLContainer",
			seedFiles: map[string]string{
				".trabuco.json": `{
  "version": "1.13.2",
  "projectName": "demo",
  "groupId": "com.example.demo",
  "artifactId": "demo",
  "javaVersion": "21",
  "modules": ["Model", "SQLDatastore", "Shared", "API"],
  "database": "mysql"
}`,
			},
			opts:           TestOpts{Target: "OrderRepository", Module: "SQLDatastore", Type: "repository"},
			wantPathSuffix: "SQLDatastore/src/test/java/com/example/demo/sqldatastore/repository/OrderRepositoryTest.java",
			wantContains: []string{
				"@DataJdbcTest",
				"MySQLContainer",
				"mysql:8.0",
			},
			wantNotContains: []string{"PostgreSQLContainer"},
		},
		{
			name: "repository test on mongo uses MongoDBContainer + DataMongoTest",
			seedFiles: map[string]string{
				".trabuco.json": `{
  "version": "1.13.2",
  "projectName": "demo",
  "groupId": "com.example.demo",
  "artifactId": "demo",
  "javaVersion": "21",
  "modules": ["Model", "NoSQLDatastore", "Shared", "API"],
  "noSqlDatabase": "mongodb"
}`,
			},
			opts:           TestOpts{Target: "OrderRepository", Module: "NoSQLDatastore", Type: "repository"},
			wantPathSuffix: "NoSQLDatastore/src/test/java/com/example/demo/nosqldatastore/repository/OrderRepositoryTest.java",
			wantContains: []string{
				"@DataMongoTest",
				"MongoDBContainer",
				"mongo:7.0",
			},
			wantNotContains: []string{"DataJdbcTest", "PostgreSQL"},
		},
		{
			name:           "no suffix Target lands in module root package",
			seedFiles:      apiSqlPgFixture(),
			opts:           TestOpts{Target: "MyHelper", Module: "Shared"},
			wantPathSuffix: "Shared/src/test/java/com/example/demo/shared/MyHelperTest.java",
			wantContains:   []string{"package com.example.demo.shared;"},
		},
		{
			name:            "rejects empty target",
			seedFiles:       apiSqlPgFixture(),
			opts:            TestOpts{Target: "", Module: "Shared"},
			wantErrContains: "target class name is required",
		},
		{
			name:            "rejects invalid Java identifier",
			seedFiles:       apiSqlPgFixture(),
			opts:            TestOpts{Target: "1Bad", Module: "Shared"},
			wantErrContains: "not a valid Java class name",
		},
		{
			name:            "rejects unknown type",
			seedFiles:       apiSqlPgFixture(),
			opts:            TestOpts{Target: "Foo", Module: "Shared", Type: "smoke"},
			wantErrContains: "--type must be one of",
		},
		{
			name:            "rejects missing module",
			seedFiles:       apiSqlPgFixture(),
			opts:            TestOpts{Target: "Foo"},
			wantErrContains: "--module is required",
		},
		{
			name:            "rejects module not in project",
			seedFiles:       apiSqlPgFixture(),
			opts:            TestOpts{Target: "Foo", Module: "Worker"},
			wantErrContains: "project does not have module Worker",
		},
		{
			name:            "rejects repository test against non-datastore module",
			seedFiles:       apiSqlPgFixture(),
			opts:            TestOpts{Target: "Foo", Module: "Shared", Type: "repository"},
			wantErrContains: "requires --module=SQLDatastore or NoSQLDatastore",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			projectPath := setupProject(t, tc.seedFiles)
			ctx, err := LoadContext(projectPath)
			if err != nil {
				t.Fatalf("LoadContext: %v", err)
			}

			result, err := GenerateTest(ctx, tc.opts)

			if tc.wantErrContains != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (result %+v)", tc.wantErrContains, result)
				}
				if !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErrContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.wantPathSuffix != "" {
				if len(result.Created) != 1 || !strings.HasSuffix(result.Created[0], tc.wantPathSuffix) {
					t.Fatalf("Created = %v, want one file ending in %q", result.Created, tc.wantPathSuffix)
				}
			}

			body, err := os.ReadFile(filepath.Join(projectPath, result.Created[0]))
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tc.wantContains {
				if !strings.Contains(string(body), want) {
					t.Errorf("file missing %q\ngot:\n%s", want, body)
				}
			}
			for _, ban := range tc.wantNotContains {
				if strings.Contains(string(body), ban) {
					t.Errorf("file contains forbidden %q\ngot:\n%s", ban, body)
				}
			}
		})
	}
}
