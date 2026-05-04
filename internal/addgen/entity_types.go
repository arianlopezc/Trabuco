package addgen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/arianlopezc/Trabuco/internal/config"
	"github.com/arianlopezc/Trabuco/internal/utils"
)

// FieldType is the user-facing type token from --fields. Each token
// has a single canonical Java type and per-database SQL type.
type FieldType string

const (
	FTString    FieldType = "string"
	FTText      FieldType = "text"
	FTInteger   FieldType = "integer"
	FTLong      FieldType = "long"
	FTDecimal   FieldType = "decimal"
	FTBoolean   FieldType = "boolean"
	FTInstant   FieldType = "instant"
	FTLocalDate FieldType = "localdate"
	FTUUID      FieldType = "uuid"
	FTJSON      FieldType = "json"
	FTEnum      FieldType = "enum"
	FTBytes     FieldType = "bytes"
)

// Field is one parsed entry from --fields="name:type[?]". For enum
// fields, EnumName holds the referenced enum class name (e.g.
// "Status" from "status:enum:Status").
type Field struct {
	Name     string
	Type     FieldType
	Nullable bool
	EnumName string
}

// JavaType returns the boxed Java type for the field. Boxed
// (Integer/Long/Boolean) regardless of nullability so the same
// declaration works as a record component, builder argument, and
// nullable @Nullable interface method. Spring Data JDBC unboxes for
// non-null columns automatically.
func (f Field) JavaType() string {
	switch f.Type {
	case FTString, FTText:
		return "String"
	case FTInteger:
		return "Integer"
	case FTLong:
		return "Long"
	case FTDecimal:
		return "BigDecimal"
	case FTBoolean:
		return "Boolean"
	case FTInstant:
		return "Instant"
	case FTLocalDate:
		return "LocalDate"
	case FTUUID:
		return "UUID"
	case FTJSON:
		return "JsonNode"
	case FTBytes:
		return "byte[]"
	case FTEnum:
		return f.EnumName
	}
	return "Object"
}

// JavaImport returns the FQN that needs to be imported for this
// field's Java type, or "" if no import is needed (primitives,
// java.lang types, same-package enum references).
func (f Field) JavaImport() string {
	switch f.Type {
	case FTDecimal:
		return "java.math.BigDecimal"
	case FTInstant:
		return "java.time.Instant"
	case FTLocalDate:
		return "java.time.LocalDate"
	case FTUUID:
		return "java.util.UUID"
	case FTJSON:
		return "com.fasterxml.jackson.databind.JsonNode"
	}
	return ""
}

// SQLType returns the column DDL type for a database flavor.
// PostgreSQL gets the richer types (NUMERIC, JSONB, TIMESTAMP WITH
// TIME ZONE); MySQL falls back to the closest standard type.
func (f Field) SQLType(database string) string {
	switch f.Type {
	case FTString:
		return "VARCHAR(255)"
	case FTText:
		return "TEXT"
	case FTInteger:
		if database == config.DatabaseMySQL {
			return "INT"
		}
		return "INTEGER"
	case FTLong:
		return "BIGINT"
	case FTDecimal:
		if database == config.DatabaseMySQL {
			return "DECIMAL(19,4)"
		}
		return "NUMERIC(19,4)"
	case FTBoolean:
		return "BOOLEAN"
	case FTInstant:
		if database == config.DatabaseMySQL {
			return "TIMESTAMP(6)"
		}
		return "TIMESTAMP WITH TIME ZONE"
	case FTLocalDate:
		return "DATE"
	case FTUUID:
		if database == config.DatabaseMySQL {
			return "BINARY(16)"
		}
		return "UUID"
	case FTJSON:
		if database == config.DatabaseMySQL {
			return "JSON"
		}
		return "JSONB"
	case FTBytes:
		if database == config.DatabaseMySQL {
			return "LONGBLOB"
		}
		return "BYTEA"
	case FTEnum:
		return "VARCHAR(64)"
	}
	return ""
}

// ColumnName returns the snake_case column name for this field.
// "customerId" → "customer_id", "URLConfig" → "url_config".
func (f Field) ColumnName() string {
	return utils.ToSnakeCase(f.Name)
}

// ParseFields splits a --fields="..." value into typed Field structs.
// Format:
//
//	"name:type[?]"            simple type, optional nullable suffix
//	"name:enum:EnumName[?]"   enum referencing an enum class
//
// Multiple entries are comma-separated. Whitespace around tokens is
// trimmed. Returns an error pinpointing which token failed if any.
func ParseFields(spec string) ([]Field, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, fmt.Errorf("--fields must not be empty (e.g. --fields=\"name:string,total:decimal\")")
	}
	rawTokens := strings.Split(spec, ",")
	fields := make([]Field, 0, len(rawTokens))
	seen := map[string]bool{}
	for _, raw := range rawTokens {
		token := strings.TrimSpace(raw)
		if token == "" {
			return nil, fmt.Errorf("--fields contains empty entry — check for trailing or duplicate commas")
		}
		f, err := parseField(token)
		if err != nil {
			return nil, fmt.Errorf("invalid field %q: %w", token, err)
		}
		if seen[f.Name] {
			return nil, fmt.Errorf("--fields has duplicate field name %q", f.Name)
		}
		seen[f.Name] = true
		fields = append(fields, f)
	}
	return fields, nil
}

func parseField(token string) (Field, error) {
	parts := strings.Split(token, ":")
	if len(parts) < 2 {
		return Field{}, fmt.Errorf("expected name:type or name:enum:EnumName")
	}
	name := strings.TrimSpace(parts[0])
	if !isValidJavaIdentifier(name) {
		return Field{}, fmt.Errorf("name %q is not a valid Java identifier", name)
	}
	if reservedJavaWords[name] {
		return Field{}, fmt.Errorf("name %q is a reserved Java keyword", name)
	}

	rawType := strings.TrimSpace(parts[1])
	nullable := false
	// `?` may live on the trailing token. For "name:string?" → token[1] = "string?".
	// For "name:enum:Status?" → token[2] = "Status?".
	last := len(parts) - 1
	tail := strings.TrimSpace(parts[last])
	if strings.HasSuffix(tail, "?") {
		nullable = true
		tail = strings.TrimSuffix(tail, "?")
		parts[last] = tail
		if last == 1 {
			rawType = tail
		}
	}

	switch FieldType(strings.ToLower(rawType)) {
	case FTString, FTText, FTInteger, FTLong, FTDecimal, FTBoolean,
		FTInstant, FTLocalDate, FTUUID, FTJSON, FTBytes:
		if len(parts) != 2 {
			return Field{}, fmt.Errorf("type %q does not take parameters", rawType)
		}
		// "int" is a friendly alias for "integer"
		t := FieldType(strings.ToLower(rawType))
		if t == "int" {
			t = FTInteger
		}
		return Field{Name: name, Type: t, Nullable: nullable}, nil
	case FTEnum:
		if len(parts) != 3 {
			return Field{}, fmt.Errorf("enum type requires explicit name: name:enum:EnumName")
		}
		enumName := strings.TrimSpace(parts[2])
		if !isValidJavaIdentifier(enumName) || !isUpperFirst(enumName) {
			return Field{}, fmt.Errorf("enum name %q must be a PascalCase Java identifier", enumName)
		}
		return Field{Name: name, Type: FTEnum, Nullable: nullable, EnumName: enumName}, nil
	}
	if rawType == "int" { // fallthrough alias outside the FieldType switch
		return Field{Name: name, Type: FTInteger, Nullable: nullable}, nil
	}
	return Field{}, fmt.Errorf("unknown type %q (supported: string, text, integer, long, decimal, boolean, instant, localdate, uuid, json, bytes, enum:Name)", rawType)
}

func isUpperFirst(s string) bool {
	if s == "" {
		return false
	}
	r := s[0]
	return r >= 'A' && r <= 'Z'
}

// reservedJavaWords blocks the most likely accidental field names.
// Not exhaustive — Java's full list has 50+ entries — but covers the
// ones an agent might naturally pick from a domain prompt.
var reservedJavaWords = map[string]bool{
	"abstract": true, "assert": true, "boolean": true, "break": true,
	"byte": true, "case": true, "catch": true, "char": true,
	"class": true, "const": true, "continue": true, "default": true,
	"do": true, "double": true, "else": true, "enum": true,
	"extends": true, "final": true, "finally": true, "float": true,
	"for": true, "goto": true, "if": true, "implements": true,
	"import": true, "instanceof": true, "int": true, "interface": true,
	"long": true, "native": true, "new": true, "null": true,
	"package": true, "private": true, "protected": true, "public": true,
	"return": true, "short": true, "static": true, "strictfp": true,
	"super": true, "switch": true, "synchronized": true, "this": true,
	"throw": true, "throws": true, "transient": true, "try": true,
	"void": true, "volatile": true, "while": true, "true": true,
	"false": true,
}

// uniqueImports collects every JavaImport from a slice of fields,
// dedupes, and sorts alphabetically. The output is ready to render
// directly into the import block of an entity / record / repository
// file.
func uniqueImports(fields []Field, extras ...string) []string {
	set := map[string]bool{}
	for _, f := range fields {
		if imp := f.JavaImport(); imp != "" {
			set[imp] = true
		}
	}
	for _, e := range extras {
		if e != "" {
			set[e] = true
		}
	}
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
