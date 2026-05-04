package addgen

import (
	"strings"
	"testing"
)

func TestParseFields(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		want     []Field
		wantErr  string
	}{
		{
			name: "single string field",
			in:   "name:string",
			want: []Field{{Name: "name", Type: FTString}},
		},
		{
			name: "all primitives",
			in:   "a:string,b:text,c:integer,d:long,e:decimal,f:boolean,g:instant,h:localdate,i:uuid,j:json,k:bytes",
			want: []Field{
				{Name: "a", Type: FTString},
				{Name: "b", Type: FTText},
				{Name: "c", Type: FTInteger},
				{Name: "d", Type: FTLong},
				{Name: "e", Type: FTDecimal},
				{Name: "f", Type: FTBoolean},
				{Name: "g", Type: FTInstant},
				{Name: "h", Type: FTLocalDate},
				{Name: "i", Type: FTUUID},
				{Name: "j", Type: FTJSON},
				{Name: "k", Type: FTBytes},
			},
		},
		{
			name: "int alias for integer",
			in:   "count:int",
			want: []Field{{Name: "count", Type: FTInteger}},
		},
		{
			name: "nullable suffix",
			in:   "notes:text?,total:decimal",
			want: []Field{
				{Name: "notes", Type: FTText, Nullable: true},
				{Name: "total", Type: FTDecimal},
			},
		},
		{
			name: "enum field",
			in:   "status:enum:Status",
			want: []Field{{Name: "status", Type: FTEnum, EnumName: "Status"}},
		},
		{
			name: "nullable enum",
			in:   "kind:enum:Kind?",
			want: []Field{{Name: "kind", Type: FTEnum, EnumName: "Kind", Nullable: true}},
		},
		{
			name: "whitespace tolerated",
			in:   "  customerId : string  ,  total : decimal  ",
			want: []Field{
				{Name: "customerId", Type: FTString},
				{Name: "total", Type: FTDecimal},
			},
		},
		{
			name: "case-insensitive type",
			in:   "x:STRING,y:Integer,z:DeCiMaL",
			want: []Field{
				{Name: "x", Type: FTString},
				{Name: "y", Type: FTInteger},
				{Name: "z", Type: FTDecimal},
			},
		},
		{name: "empty spec", in: "", wantErr: "--fields must not be empty"},
		{name: "trailing comma", in: "x:string,", wantErr: "empty entry"},
		{name: "missing type", in: "x", wantErr: "name:type"},
		{name: "unknown type", in: "x:weird", wantErr: "unknown type"},
		{name: "invalid name", in: "1bad:string", wantErr: "valid Java identifier"},
		{name: "reserved keyword", in: "class:string", wantErr: "reserved Java keyword"},
		{name: "duplicate field", in: "x:string,x:long", wantErr: "duplicate field name"},
		{name: "enum without name", in: "status:enum", wantErr: "enum type requires"},
		{name: "lowercase enum name rejected", in: "status:enum:status", wantErr: "PascalCase"},
		{name: "extra colons on simple type", in: "x:string:Foo", wantErr: "does not take parameters"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseFields(tc.in)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got fields %+v", tc.wantErr, got)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tc.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %d fields, want %d (%v vs %v)", len(got), len(tc.want), got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("field[%d] = %+v, want %+v", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestField_JavaType(t *testing.T) {
	cases := []struct {
		f    Field
		want string
	}{
		{Field{Type: FTString}, "String"},
		{Field{Type: FTText}, "String"},
		{Field{Type: FTInteger}, "Integer"},
		{Field{Type: FTLong}, "Long"},
		{Field{Type: FTDecimal}, "BigDecimal"},
		{Field{Type: FTBoolean}, "Boolean"},
		{Field{Type: FTInstant}, "Instant"},
		{Field{Type: FTLocalDate}, "LocalDate"},
		{Field{Type: FTUUID}, "UUID"},
		{Field{Type: FTJSON}, "JsonNode"},
		{Field{Type: FTBytes}, "byte[]"},
		{Field{Type: FTEnum, EnumName: "Status"}, "Status"},
	}
	for _, tc := range cases {
		got := tc.f.JavaType()
		if got != tc.want {
			t.Errorf("JavaType(%v) = %q, want %q", tc.f, got, tc.want)
		}
	}
}

func TestField_JavaImport(t *testing.T) {
	imports := map[FieldType]string{
		FTString:    "",
		FTText:      "",
		FTInteger:   "",
		FTLong:      "",
		FTBoolean:   "",
		FTBytes:     "",
		FTDecimal:   "java.math.BigDecimal",
		FTInstant:   "java.time.Instant",
		FTLocalDate: "java.time.LocalDate",
		FTUUID:      "java.util.UUID",
		FTJSON:      "com.fasterxml.jackson.databind.JsonNode",
		FTEnum:      "",
	}
	for ft, want := range imports {
		got := Field{Type: ft}.JavaImport()
		if got != want {
			t.Errorf("JavaImport(%s) = %q, want %q", ft, got, want)
		}
	}
}

func TestField_SQLType(t *testing.T) {
	cases := []struct {
		ft       FieldType
		pgsql    string
		mysql    string
	}{
		{FTString, "VARCHAR(255)", "VARCHAR(255)"},
		{FTText, "TEXT", "TEXT"},
		{FTInteger, "INTEGER", "INT"},
		{FTLong, "BIGINT", "BIGINT"},
		{FTDecimal, "NUMERIC(19,4)", "DECIMAL(19,4)"},
		{FTBoolean, "BOOLEAN", "BOOLEAN"},
		{FTInstant, "TIMESTAMP WITH TIME ZONE", "TIMESTAMP(6)"},
		{FTLocalDate, "DATE", "DATE"},
		{FTUUID, "UUID", "BINARY(16)"},
		{FTJSON, "JSONB", "JSON"},
		{FTBytes, "BYTEA", "LONGBLOB"},
		{FTEnum, "VARCHAR(64)", "VARCHAR(64)"},
	}
	for _, tc := range cases {
		f := Field{Type: tc.ft}
		if got := f.SQLType("postgresql"); got != tc.pgsql {
			t.Errorf("SQLType(%s, postgresql) = %q, want %q", tc.ft, got, tc.pgsql)
		}
		if got := f.SQLType("mysql"); got != tc.mysql {
			t.Errorf("SQLType(%s, mysql) = %q, want %q", tc.ft, got, tc.mysql)
		}
	}
}

func TestField_ColumnName(t *testing.T) {
	cases := []struct{ name, want string }{
		{"customerId", "customer_id"},
		{"name", "name"},
		{"createdAt", "created_at"},
		{"URLConfig", "url_config"},
	}
	for _, tc := range cases {
		got := Field{Name: tc.name}.ColumnName()
		if got != tc.want {
			t.Errorf("ColumnName(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestUniqueImports(t *testing.T) {
	fields := []Field{
		{Type: FTString},                    // no import
		{Type: FTDecimal},                   // java.math.BigDecimal
		{Type: FTInstant},                   // java.time.Instant
		{Type: FTLocalDate},                 // java.time.LocalDate
		{Type: FTDecimal},                   // duplicate — must dedupe
		{Type: FTEnum, EnumName: "Status"},  // no import (same package)
	}
	got := uniqueImports(fields, "jakarta.annotation.Nullable")
	want := []string{
		"jakarta.annotation.Nullable",
		"java.math.BigDecimal",
		"java.time.Instant",
		"java.time.LocalDate",
	}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("uniqueImports = %v, want %v", got, want)
	}
}
