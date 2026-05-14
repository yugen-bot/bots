package utils

import "unicode"

// IsPalindrome reports whether s reads the same forwards and backwards,
// ignoring case. Works correctly with multi-byte Unicode characters.
func IsPalindrome(s string) bool {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		if unicode.ToLower(runes[i]) != unicode.ToLower(runes[j]) {
			return false
		}
	}

	return true
}
