package config

import (
	"strings"
	"testing"
)

func TestValidateVectorStoreFlag(t *testing.T) {
	cases := []struct {
		in        string
		wantError bool
	}{
		{"", false},
		{"none", false},
		{"pgvector", false},
		{"qdrant", false},
		{"mongodb", false},
		{"PGVECTOR", true},
		{"postgres", true},
		{"redis", true},
		{"chroma", true},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			got := ValidateVectorStoreFlag(tc.in)
			if (got != "") != tc.wantError {
				t.Fatalf("ValidateVectorStoreFlag(%q) = %q, wantError=%v", tc.in, got, tc.wantError)
			}
		})
	}
}

func TestResolveVectorStore_NoOpForEmptyOrNone(t *testing.T) {
	for _, vs := range []string{"", "none"} {
		t.Run(vs, func(t *testing.T) {
			cfg := &ProjectConfig{
				Modules:     []string{ModuleModel, ModuleAPI},
				Database:    DatabaseMySQL,
				VectorStore: vs,
			}
			if err := cfg.ResolveVectorStore(); err != "" {
				t.Fatalf("expected no-op for %q, got error: %s", vs, err)
			}
			if cfg.Database != DatabaseMySQL {
				t.Fatalf("Database mutated unexpectedly: %s", cfg.Database)
			}
		})
	}
}

func TestResolveVectorStore_RequiresAIAgent(t *testing.T) {
	cfg := &ProjectConfig{
		Modules:     []string{ModuleModel, ModuleAPI},
		Database:    DatabasePostgreSQL,
		VectorStore: VectorStorePgVector,
	}
	got := cfg.ResolveVectorStore()
	if got == "" {
		t.Fatal("expected error for vector store without AIAgent, got nil")
	}
	if !strings.Contains(got, "AIAgent") {
		t.Fatalf("error message should mention AIAgent: %q", got)
	}
}

func TestResolveVectorStore_PgVector_AutoAddsSQLDatastore(t *testing.T) {
	cfg := &ProjectConfig{
		Modules:     []string{ModuleModel, ModuleAPI, ModuleAIAgent},
		Database:    "",
		VectorStore: VectorStorePgVector,
	}
	if err := cfg.ResolveVectorStore(); err != "" {
		t.Fatalf("unexpected error: %s", err)
	}
	if !cfg.HasModule(ModuleSQLDatastore) {
		t.Fatalf("SQLDatastore should have been auto-added; modules=%v", cfg.Modules)
	}
	if cfg.Database != DatabasePostgreSQL {
		t.Fatalf("Database should default to postgresql, got %q", cfg.Database)
	}
}

func TestResolveVectorStore_PgVector_RejectsMySQL(t *testing.T) {
	cfg := &ProjectConfig{
		Modules:     []string{ModuleModel, ModuleSQLDatastore, ModuleAPI, ModuleAIAgent},
		Database:    DatabaseMySQL,
		VectorStore: VectorStorePgVector,
	}
	got := cfg.ResolveVectorStore()
	if got == "" {
		t.Fatal("expected error for pgvector + mysql, got nil")
	}
	if !strings.Contains(got, "postgresql") {
		t.Fatalf("error message should mention postgresql: %q", got)
	}
}

func TestResolveVectorStore_PgVector_RejectsNoSQLDatastore(t *testing.T) {
	cfg := &ProjectConfig{
		Modules:       []string{ModuleModel, ModuleNoSQLDatastore, ModuleAPI, ModuleAIAgent},
		NoSQLDatabase: DatabaseMongoDB,
		VectorStore:   VectorStorePgVector,
	}
	got := cfg.ResolveVectorStore()
	if got == "" {
		t.Fatal("expected error for pgvector + NoSQLDatastore, got nil")
	}
	if !strings.Contains(got, "NoSQLDatastore") {
		t.Fatalf("error message should mention NoSQLDatastore: %q", got)
	}
}

func TestResolveVectorStore_MongoDB_CoercesNoSQLDatabase(t *testing.T) {
	cfg := &ProjectConfig{
		Modules:       []string{ModuleModel, ModuleNoSQLDatastore, ModuleAPI, ModuleAIAgent},
		NoSQLDatabase: "",
		VectorStore:   VectorStoreMongoDB,
	}
	if err := cfg.ResolveVectorStore(); err != "" {
		t.Fatalf("unexpected error: %s", err)
	}
	if cfg.NoSQLDatabase != DatabaseMongoDB {
		t.Fatalf("NoSQLDatabase should be coerced to mongodb, got %q", cfg.NoSQLDatabase)
	}
}

func TestResolveVectorStore_MongoDB_RejectsRedis(t *testing.T) {
	cfg := &ProjectConfig{
		Modules:       []string{ModuleModel, ModuleNoSQLDatastore, ModuleAPI, ModuleAIAgent},
		NoSQLDatabase: DatabaseRedis,
		VectorStore:   VectorStoreMongoDB,
	}
	got := cfg.ResolveVectorStore()
	if got == "" {
		t.Fatal("expected error for mongodb vector store + redis nosql, got nil")
	}
	if !strings.Contains(got, "mongodb") {
		t.Fatalf("error message should mention mongodb: %q", got)
	}
}

func TestResolveVectorStore_MongoDB_StandaloneIsValid(t *testing.T) {
	cfg := &ProjectConfig{
		Modules:     []string{ModuleModel, ModuleAPI, ModuleAIAgent},
		VectorStore: VectorStoreMongoDB,
	}
	if err := cfg.ResolveVectorStore(); err != "" {
		t.Fatalf("standalone mongodb should be valid: %s", err)
	}
	if !cfg.VectorStoreNeedsStandaloneMongoConnection() {
		t.Fatal("standalone mongodb should need its own connection wiring")
	}
}

func TestResolveVectorStore_Qdrant_HasNoConstraints(t *testing.T) {
	cases := []ProjectConfig{
		{Modules: []string{ModuleModel, ModuleAPI, ModuleAIAgent}, VectorStore: VectorStoreQdrant},
		{Modules: []string{ModuleModel, ModuleSQLDatastore, ModuleAPI, ModuleAIAgent}, Database: DatabaseMySQL, VectorStore: VectorStoreQdrant},
		{Modules: []string{ModuleModel, ModuleNoSQLDatastore, ModuleAPI, ModuleAIAgent}, NoSQLDatabase: DatabaseRedis, VectorStore: VectorStoreQdrant},
	}
	for i := range cases {
		cfg := cases[i]
		if err := cfg.ResolveVectorStore(); err != "" {
			t.Fatalf("qdrant should have no constraints, case %d errored: %s", i, err)
		}
	}
}
