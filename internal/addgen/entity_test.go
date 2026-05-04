package addgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGenerateEntity_SQL_Postgres exercises every supported field
// type against a Postgres SQLDatastore project, asserting both Java
// and SQL outputs contain the right type tokens. This is the primary
// data-type-matrix regression suite.
func TestGenerateEntity_SQL_Postgres(t *testing.T) {
	cases := []struct {
		name           string
		fields         string
		entityContains []string // substrings expected in {Name}.java + {Name}Record.java
		migContains    []string // substrings expected in V{N}__create_{table}.sql
	}{
		{
			name:           "string field",
			fields:         "name:string",
			entityContains: []string{"String name();", "String name", "@Table(\"orders\")"},
			migContains:    []string{"name VARCHAR(255) NOT NULL", "id BIGSERIAL PRIMARY KEY"},
		},
		{
			name:           "text field",
			fields:         "body:text",
			entityContains: []string{"String body();", "String body"},
			migContains:    []string{"body TEXT NOT NULL"},
		},
		{
			name:           "integer field",
			fields:         "count:integer",
			entityContains: []string{"Integer count();"},
			migContains:    []string{"count INTEGER NOT NULL"},
		},
		{
			name:           "int alias",
			fields:         "count:int",
			entityContains: []string{"Integer count();"},
			migContains:    []string{"count INTEGER NOT NULL"},
		},
		{
			name:           "long field",
			fields:         "size:long",
			entityContains: []string{"Long size();"},
			migContains:    []string{"size BIGINT NOT NULL"},
		},
		{
			name:           "decimal field",
			fields:         "price:decimal",
			entityContains: []string{"BigDecimal price();", "import java.math.BigDecimal;"},
			migContains:    []string{"price NUMERIC(19,4) NOT NULL"},
		},
		{
			name:           "boolean field",
			fields:         "active:boolean",
			entityContains: []string{"Boolean active();"},
			migContains:    []string{"active BOOLEAN NOT NULL"},
		},
		{
			name:           "instant field",
			fields:         "createdAt:instant",
			entityContains: []string{"Instant createdAt();", "import java.time.Instant;"},
			migContains:    []string{"created_at TIMESTAMP WITH TIME ZONE NOT NULL"},
		},
		{
			name:           "localdate field",
			fields:         "birthDate:localdate",
			entityContains: []string{"LocalDate birthDate();", "import java.time.LocalDate;"},
			migContains:    []string{"birth_date DATE NOT NULL"},
		},
		{
			name:           "uuid field",
			fields:         "externalRef:uuid",
			entityContains: []string{"UUID externalRef();", "import java.util.UUID;"},
			migContains:    []string{"external_ref UUID NOT NULL"},
		},
		{
			name:           "json field",
			fields:         "metadata:json",
			entityContains: []string{"JsonNode metadata();", "import com.fasterxml.jackson.databind.JsonNode;"},
			migContains:    []string{"metadata JSONB NOT NULL"},
		},
		{
			name:           "bytes field",
			fields:         "payload:bytes",
			entityContains: []string{"byte[] payload();"},
			migContains:    []string{"payload BYTEA NOT NULL"},
		},
		{
			name:           "enum field",
			fields:         "status:enum:Status",
			entityContains: []string{"Status status();"}, // no import (same package)
			migContains:    []string{"status VARCHAR(64) NOT NULL"},
		},
		{
			name:           "nullable string",
			fields:         "tag:string?",
			entityContains: []string{"@Nullable\n  String tag();"},
			migContains:    []string{"tag VARCHAR(255)\n"}, // no NOT NULL
		},
		{
			name:           "nullable enum",
			fields:         "kind:enum:Kind?",
			entityContains: []string{"@Nullable\n  Kind kind();"},
			migContains:    []string{"kind VARCHAR(64)\n"},
		},
		{
			name:   "multi-field combo",
			fields: "customerId:string,total:decimal,placedAt:instant,notes:text?",
			entityContains: []string{
				"String customerId();",
				"BigDecimal total();",
				"Instant placedAt();",
				"@Nullable\n  String notes();",
			},
			migContains: []string{
				"customer_id VARCHAR(255) NOT NULL",
				"total NUMERIC(19,4) NOT NULL",
				"placed_at TIMESTAMP WITH TIME ZONE NOT NULL",
				"notes TEXT\n", // nullable
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := setupProject(t, apiSqlPgFixture())
			ctx := mustCtx(t, project)
			result, err := GenerateEntity(ctx, EntityOpts{
				Name:   "Order",
				Fields: tc.fields,
			})
			if err != nil {
				t.Fatalf("GenerateEntity: %v", err)
			}
			entityFile := readPath(t, project, "Model/src/main/java/com/example/demo/model/entities/Order.java")
			recordFile := readPath(t, project, "Model/src/main/java/com/example/demo/model/entities/OrderRecord.java")
			migrationFile := readPath(t, project, findMigration(t, project))
			combined := entityFile + "\n" + recordFile

			for _, want := range tc.entityContains {
				if !strings.Contains(combined, want) {
					t.Errorf("entity files missing %q\nentity:\n%s\nrecord:\n%s", want, entityFile, recordFile)
				}
			}
			for _, want := range tc.migContains {
				if !strings.Contains(migrationFile, want) {
					t.Errorf("migration missing %q\ngot:\n%s", want, migrationFile)
				}
			}

			// Always-present invariants
			invariants := []string{
				"package com.example.demo.model.entities;",
				"@Value.Immutable",
				"@ImmutableStyle",
				"@JsonSerialize(as = ImmutableOrder.class)",
				"@JsonDeserialize(as = ImmutableOrder.class)",
				"@Nullable\n  Long id();",
				"@Table(\"orders\")",
				"public record OrderRecord(",
				"@Id @Nullable Long id",
			}
			for _, want := range invariants {
				if !strings.Contains(combined, want) {
					t.Errorf("entity invariant missing: %q", want)
				}
			}
			_ = result
		})
	}
}

// TestGenerateEntity_SQL_MySQL verifies MySQL-specific column types.
func TestGenerateEntity_SQL_MySQL(t *testing.T) {
	project := setupProject(t, map[string]string{
		".trabuco.json": `{
  "version": "1.13.2", "projectName": "demo", "groupId": "com.example.demo",
  "artifactId": "demo", "javaVersion": "21",
  "modules": ["Model", "SQLDatastore", "Shared", "API"], "database": "mysql"
}`,
		"SQLDatastore/src/main/resources/db/migration/V1__baseline.sql": "-- baseline\n",
	})
	ctx := mustCtx(t, project)
	if _, err := GenerateEntity(ctx, EntityOpts{
		Name:   "Order",
		Fields: "name:string,total:decimal,placedAt:instant,kind:uuid,blob:bytes,meta:json",
	}); err != nil {
		t.Fatal(err)
	}
	mig := readPath(t, project, findMigration(t, project))
	wants := []string{
		"id BIGINT AUTO_INCREMENT PRIMARY KEY",
		"name VARCHAR(255) NOT NULL",
		"total DECIMAL(19,4) NOT NULL", // DECIMAL not NUMERIC
		"placed_at TIMESTAMP(6) NOT NULL",
		"kind BINARY(16) NOT NULL",
		"blob LONGBLOB NOT NULL",
		"meta JSON NOT NULL", // JSON not JSONB
	}
	for _, w := range wants {
		if !strings.Contains(mig, w) {
			t.Errorf("MySQL migration missing %q\ngot:\n%s", w, mig)
		}
	}
}

// TestGenerateEntity_Mongo exercises the MongoDB three-file output.
func TestGenerateEntity_Mongo(t *testing.T) {
	project := setupProject(t, map[string]string{
		".trabuco.json": `{
  "version": "1.13.2", "projectName": "demo", "groupId": "com.example.demo",
  "artifactId": "demo", "javaVersion": "21",
  "modules": ["Model", "NoSQLDatastore", "Shared", "API"], "noSqlDatabase": "mongodb"
}`,
	})
	ctx := mustCtx(t, project)
	result, err := GenerateEntity(ctx, EntityOpts{
		Name:   "Order",
		Fields: "customerId:string,total:decimal,status:enum:OrderStatus,notes:text?",
	})
	if err != nil {
		t.Fatal(err)
	}

	expectedFiles := []string{
		"Model/src/main/java/com/example/demo/model/entities/Order.java",
		"Model/src/main/java/com/example/demo/model/entities/OrderDocument.java",
		"NoSQLDatastore/src/main/java/com/example/demo/nosqldatastore/repository/OrderDocumentRepository.java",
		"Model/src/main/java/com/example/demo/model/entities/OrderStatus.java",
	}
	for _, f := range expectedFiles {
		if !contains(result.Created, f) {
			t.Errorf("Mongo result missing %s, got %v", f, result.Created)
		}
	}

	// No SQL migration emitted for Mongo.
	for _, f := range result.Created {
		if strings.Contains(f, "db/migration/") {
			t.Errorf("Mongo flow should not emit a migration, got %s", f)
		}
	}

	entity := readPath(t, project, expectedFiles[0])
	doc := readPath(t, project, expectedFiles[1])
	repo := readPath(t, project, expectedFiles[2])

	wants := []string{
		"String documentId();",     // entity has documentId, not id
		"@Document(collection = \"orders\")",
		"@Id @Nullable String documentId",
		"MongoRepository<OrderDocument, String>",
	}
	for _, w := range wants {
		if !strings.Contains(entity+doc+repo, w) {
			t.Errorf("Mongo files missing %q", w)
		}
	}
}

// TestGenerateEntity_EnumDedup ensures multiple enum fields with the
// same name produce exactly one enum class file.
func TestGenerateEntity_EnumDedup(t *testing.T) {
	project := setupProject(t, apiSqlPgFixture())
	ctx := mustCtx(t, project)
	result, err := GenerateEntity(ctx, EntityOpts{
		Name:   "Order",
		Fields: "primary:enum:Status,secondary:enum:Status,tertiary:enum:Priority",
	})
	if err != nil {
		t.Fatal(err)
	}
	enumCount := 0
	for _, f := range result.Created {
		if strings.HasSuffix(f, "Status.java") || strings.HasSuffix(f, "Priority.java") {
			enumCount++
		}
	}
	if enumCount != 2 {
		t.Errorf("expected 2 enum files (Status + Priority), got %d in %v", enumCount, result.Created)
	}
}

// TestGenerateEntity_PreexistingEnumNotClobbered verifies that an
// existing enum file in the project is preserved and noted, not
// silently overwritten.
func TestGenerateEntity_PreexistingEnumNotClobbered(t *testing.T) {
	files := apiSqlPgFixture()
	files["Model/src/main/java/com/example/demo/model/entities/Status.java"] = "package com.example.demo.model.entities;\npublic enum Status { ACTIVE, INACTIVE }\n"
	project := setupProject(t, files)
	ctx := mustCtx(t, project)
	result, err := GenerateEntity(ctx, EntityOpts{
		Name:   "Order",
		Fields: "name:string,status:enum:Status",
	})
	if err != nil {
		t.Fatal(err)
	}
	enumPath := "Model/src/main/java/com/example/demo/model/entities/Status.java"
	if contains(result.Created, enumPath) {
		t.Errorf("existing enum should NOT be in Created, got %v", result.Created)
	}
	if len(result.Notes) == 0 || !strings.Contains(strings.Join(result.Notes, "\n"), "Status already exists") {
		t.Errorf("expected a Note about the existing enum, got %v", result.Notes)
	}
	body := readPath(t, project, enumPath)
	if !strings.Contains(body, "ACTIVE, INACTIVE") {
		t.Errorf("existing enum body was clobbered: %s", body)
	}
}

// TestGenerateEntity_TableNamePluralization checks the auto-table-name
// rules and the --table-name override.
func TestGenerateEntity_TableNamePluralization(t *testing.T) {
	cases := []struct {
		entity, override, wantTable string
	}{
		{"Order", "", "orders"},
		{"OrderItem", "", "order_items"},
		{"Currency", "", "currencies"},
		{"Box", "", "boxes"},
		{"Person", "people", "people"}, // override for irregular plural
	}
	for _, tc := range cases {
		t.Run(tc.entity, func(t *testing.T) {
			project := setupProject(t, apiSqlPgFixture())
			ctx := mustCtx(t, project)
			_, err := GenerateEntity(ctx, EntityOpts{
				Name:      tc.entity,
				Fields:    "x:string",
				TableName: tc.override,
			})
			if err != nil {
				t.Fatal(err)
			}
			migs, _ := filepath.Glob(filepath.Join(project, "SQLDatastore/src/main/resources/db/migration/V*__create_*.sql"))
			found := false
			for _, m := range migs {
				if strings.HasSuffix(m, "create_"+tc.wantTable+".sql") {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected migration with table %q, got files: %v", tc.wantTable, migs)
			}
		})
	}
}

// TestGenerateEntity_ModuleDisambiguation covers the --module flag
// behavior: explicit, auto-detect single, and conflict on dual.
func TestGenerateEntity_ModuleDisambiguation(t *testing.T) {
	dual := map[string]string{
		".trabuco.json": `{
  "version": "1.13.2", "projectName": "demo", "groupId": "com.example.demo",
  "artifactId": "demo", "javaVersion": "21",
  "modules": ["Model", "SQLDatastore", "NoSQLDatastore", "Shared", "API"],
  "database": "postgresql", "noSqlDatabase": "mongodb"
}`,
	}
	t.Run("dual datastores require explicit module", func(t *testing.T) {
		project := setupProject(t, dual)
		ctx := mustCtx(t, project)
		_, err := GenerateEntity(ctx, EntityOpts{Name: "Order", Fields: "name:string"})
		if err == nil || !strings.Contains(err.Error(), "pass --module") {
			t.Fatalf("expected dual-datastore disambiguation error, got %v", err)
		}
	})
	t.Run("explicit module=NoSQLDatastore in dual project", func(t *testing.T) {
		project := setupProject(t, dual)
		ctx := mustCtx(t, project)
		result, err := GenerateEntity(ctx, EntityOpts{Name: "Order", Fields: "name:string", Module: "NoSQLDatastore"})
		if err != nil {
			t.Fatal(err)
		}
		hasDoc := false
		for _, f := range result.Created {
			if strings.HasSuffix(f, "OrderDocument.java") {
				hasDoc = true
				break
			}
		}
		if !hasDoc {
			t.Errorf("expected Mongo document in Created, got %v", result.Created)
		}
	})
}

// TestGenerateEntity_Errors covers validation paths.
func TestGenerateEntity_Errors(t *testing.T) {
	cases := []struct {
		name      string
		seedFiles map[string]string
		opts      EntityOpts
		wantErr   string
	}{
		{
			name:      "empty name",
			seedFiles: apiSqlPgFixture(),
			opts:      EntityOpts{Name: "", Fields: "x:string"},
			wantErr:   "entity name is required",
		},
		{
			name:      "lowercase name",
			seedFiles: apiSqlPgFixture(),
			opts:      EntityOpts{Name: "order", Fields: "x:string"},
			wantErr:   "PascalCase",
		},
		{
			name:      "invalid name",
			seedFiles: apiSqlPgFixture(),
			opts:      EntityOpts{Name: "Order-1", Fields: "x:string"},
			wantErr:   "PascalCase",
		},
		{
			name:      "empty fields",
			seedFiles: apiSqlPgFixture(),
			opts:      EntityOpts{Name: "Order", Fields: ""},
			wantErr:   "must not be empty",
		},
		{
			name: "no Model module",
			seedFiles: map[string]string{
				".trabuco.json": `{
  "version": "1.13.2", "projectName": "demo", "groupId": "com.example.demo",
  "artifactId": "demo", "javaVersion": "21",
  "modules": ["SQLDatastore", "Shared", "API"], "database": "postgresql"
}`,
			},
			opts:    EntityOpts{Name: "Order", Fields: "x:string"},
			wantErr: "Model module",
		},
		{
			name: "no datastore module",
			seedFiles: map[string]string{
				".trabuco.json": `{
  "version": "1.13.2", "projectName": "demo", "groupId": "com.example.demo",
  "artifactId": "demo", "javaVersion": "21",
  "modules": ["Model", "Shared"]
}`,
			},
			opts:    EntityOpts{Name: "Order", Fields: "x:string"},
			wantErr: "no datastore module",
		},
		{
			name:      "explicit module not in project",
			seedFiles: apiSqlPgFixture(),
			opts:      EntityOpts{Name: "Order", Fields: "x:string", Module: "NoSQLDatastore"},
			wantErr:   "doesn't have it",
		},
		{
			name:      "explicit unknown module",
			seedFiles: apiSqlPgFixture(),
			opts:      EntityOpts{Name: "Order", Fields: "x:string", Module: "Worker"},
			wantErr:   "must be SQLDatastore or NoSQLDatastore",
		},
		{
			name:      "refuse-clobber on second run",
			seedFiles: apiSqlPgFixture(),
			opts:      EntityOpts{Name: "Order", Fields: "x:string"},
			// Tested via two-call sequence below
			wantErr: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			project := setupProject(t, tc.seedFiles)
			ctx := mustCtx(t, project)
			_, err := GenerateEntity(ctx, tc.opts)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErr)
			}
		})
	}

	t.Run("second run on same entity refuses to clobber", func(t *testing.T) {
		project := setupProject(t, apiSqlPgFixture())
		ctx := mustCtx(t, project)
		if _, err := GenerateEntity(ctx, EntityOpts{Name: "Order", Fields: "x:string"}); err != nil {
			t.Fatalf("first call: %v", err)
		}
		_, err := GenerateEntity(ctx, EntityOpts{Name: "Order", Fields: "x:string"})
		if err == nil || !strings.Contains(err.Error(), "refusing to overwrite") {
			t.Fatalf("expected refuse-to-overwrite, got %v", err)
		}
	})
}

// --- helpers ---

func mustCtx(t *testing.T, project string) *Context {
	t.Helper()
	ctx, err := LoadContext(project)
	if err != nil {
		t.Fatalf("LoadContext: %v", err)
	}
	return ctx
}

func readPath(t *testing.T, project, rel string) string {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(project, rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(body)
}

func findMigration(t *testing.T, project string) string {
	t.Helper()
	migs, err := filepath.Glob(filepath.Join(project, "SQLDatastore/src/main/resources/db/migration/V*__create_*.sql"))
	if err != nil || len(migs) == 0 {
		t.Fatalf("no create_* migration found in project")
	}
	rel, err := filepath.Rel(project, migs[0])
	if err != nil {
		t.Fatal(err)
	}
	return rel
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}
