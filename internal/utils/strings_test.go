package utils

import "testing"

func TestToSnakeCase(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"order", "order"},
		{"Order", "order"},
		{"OrderItem", "order_item"},
		{"orderItem", "order_item"},
		{"URLConfig", "url_config"},
		{"HTTPRequest", "http_request"},
		{"already_snake", "already_snake"},
		{"customerId", "customer_id"},
		{"v2Migration", "v2_migration"},
		{"kebab-case", "kebab_case"},
	}
	for _, tc := range cases {
		got := ToSnakeCase(tc.in)
		if got != tc.want {
			t.Errorf("ToSnakeCase(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPluralLowerSnake(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"Order", "orders"},
		{"OrderItem", "order_items"},
		{"Customer", "customers"},
		{"Currency", "currencies"},
		{"Box", "boxes"},
		{"Branch", "branches"},
		{"Wish", "wishes"},
		{"Day", "days"}, // vowel before y → just +s
	}
	for _, tc := range cases {
		got := PluralLowerSnake(tc.in)
		if got != tc.want {
			t.Errorf("PluralLowerSnake(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-project", "MyProject"},
		{"simple", "Simple"},
		{"my-cool-app", "MyCoolApp"},
		{"already_snake", "AlreadySnake"},
		{"UPPER", "UPPER"},
		{"", ""},
		{"a", "A"},
		{"a-b-c", "ABC"},
	}

	for _, tt := range tests {
		result := ToPascalCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToCamelCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"my-project", "myProject"},
		{"simple", "simple"},
		{"my-cool-app", "myCoolApp"},
		{"already_snake", "alreadySnake"},
		{"", ""},
		{"A", "a"},
		{"ABC", "aBC"},
	}

	for _, tt := range tests {
		result := ToCamelCase(tt.input)
		if result != tt.expected {
			t.Errorf("ToCamelCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToTitle(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello world", "Hello World"},
		{"hello", "Hello"},
		{"HELLO", "HELLO"},
		{"", ""},
		{"a b c", "A B C"},
	}

	for _, tt := range tests {
		result := ToTitle(tt.input)
		if result != tt.expected {
			t.Errorf("ToTitle(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToUpper(t *testing.T) {
	tests := []struct {
		input    rune
		expected rune
	}{
		{'a', 'A'},
		{'z', 'Z'},
		{'A', 'A'},
		{'1', '1'},
	}

	for _, tt := range tests {
		result := ToUpper(tt.input)
		if result != tt.expected {
			t.Errorf("ToUpper(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    rune
		expected rune
	}{
		{'A', 'a'},
		{'Z', 'z'},
		{'a', 'a'},
		{'1', '1'},
	}

	for _, tt := range tests {
		result := ToLower(tt.input)
		if result != tt.expected {
			t.Errorf("ToLower(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
