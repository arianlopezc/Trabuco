package utils

// ToPascalCase converts kebab-case or snake_case to PascalCase
func ToPascalCase(s string) string {
	result := ""
	capitalizeNext := true
	for _, ch := range s {
		if ch == '-' || ch == '_' {
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result += string(ToUpper(ch))
			capitalizeNext = false
		} else {
			result += string(ch)
		}
	}
	return result
}

// ToCamelCase converts kebab-case or snake_case to camelCase
func ToCamelCase(s string) string {
	pascal := ToPascalCase(s)
	if len(pascal) == 0 {
		return ""
	}
	first := pascal[0]
	if first >= 'A' && first <= 'Z' {
		return string(first+32) + pascal[1:]
	}
	return pascal
}

// ToTitle capitalizes the first letter of each word
func ToTitle(s string) string {
	result := ""
	capitalizeNext := true
	for _, ch := range s {
		if ch == ' ' || ch == '\t' || ch == '\n' {
			result += string(ch)
			capitalizeNext = true
			continue
		}
		if capitalizeNext {
			result += string(ToUpper(ch))
			capitalizeNext = false
		} else {
			result += string(ch)
		}
	}
	return result
}

// ToUpper converts a lowercase ASCII letter to uppercase
func ToUpper(r rune) rune {
	if r >= 'a' && r <= 'z' {
		return r - 32
	}
	return r
}

// ToLower converts an uppercase ASCII letter to lowercase
func ToLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + 32
	}
	return r
}

// ToSnakeCase converts PascalCase or camelCase to snake_case.
// Inserts an underscore before each uppercase ASCII letter that
// follows a lowercase letter, digit, or another uppercase letter
// followed by a lowercase letter (the "ABc" → "a_bc" rule that
// keeps "OrderItem" → "order_item" and "URLConfig" → "url_config").
// Hyphens and existing underscores are preserved as-is.
func ToSnakeCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	var out []rune
	for i, r := range runes {
		isUpper := r >= 'A' && r <= 'Z'
		if isUpper && i > 0 {
			prev := runes[i-1]
			prevLowerOrDigit := (prev >= 'a' && prev <= 'z') || (prev >= '0' && prev <= '9')
			next := rune(0)
			if i+1 < len(runes) {
				next = runes[i+1]
			}
			nextLower := next >= 'a' && next <= 'z'
			prevUpper := prev >= 'A' && prev <= 'Z'

			// Insert underscore at the boundary between a lowercase/digit
			// and an uppercase letter ("orderItem" → "order_Item"), and
			// between two uppercase letters when the second is followed
			// by a lowercase ("URLConfig" → "URL_Config").
			if prevLowerOrDigit || (prevUpper && nextLower) {
				out = append(out, '_')
			}
		}
		if r == '-' {
			out = append(out, '_')
			continue
		}
		if isUpper {
			out = append(out, r+32)
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

// PluralLowerSnake converts a Java type name (PascalCase) into a
// plural snake_case table name. Applies the simple English rules:
//   - words ending in "y" after a consonant → "ies"  (Currency → currencies)
//   - words ending in s/x/z/ch/sh             → "es"  (Box → boxes)
//   - everything else                          → "s"   (Order → orders)
//
// Override with --table-name= for irregular plurals (mouse → mice,
// child → children) — the helper deliberately keeps the rule set
// small and predictable.
func PluralLowerSnake(s string) string {
	snake := ToSnakeCase(s)
	if snake == "" {
		return ""
	}
	last := snake[len(snake)-1]
	switch {
	case last == 'y' && len(snake) >= 2 && !isVowel(snake[len(snake)-2]):
		return snake[:len(snake)-1] + "ies"
	case last == 's' || last == 'x' || last == 'z':
		return snake + "es"
	}
	if len(snake) >= 2 {
		suffix := snake[len(snake)-2:]
		if suffix == "ch" || suffix == "sh" {
			return snake + "es"
		}
	}
	return snake + "s"
}

func isVowel(c byte) bool {
	switch c {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}
