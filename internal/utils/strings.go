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
