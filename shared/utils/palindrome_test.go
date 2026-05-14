package utils

import "testing"

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "empty string", input: "", want: true},
		{name: "single character", input: "a", want: true},
		{name: "simple palindrome", input: "racecar", want: true},
		{name: "not a palindrome", input: "hello", want: false},
		{name: "case-insensitive palindrome", input: "Aba", want: true},
		{name: "numeric palindrome", input: "121", want: true},
		{name: "numeric non-palindrome", input: "123", want: false},
		{name: "even-length palindrome", input: "abba", want: true},
		{name: "unicode palindrome", input: "åbå", want: true},
		{name: "two same chars", input: "aa", want: true},
		{name: "two different chars", input: "ab", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := IsPalindrome(tc.input)
			if got != tc.want {
				t.Errorf(
					"IsPalindrome(%q) = %v, want %v",
					tc.input,
					got,
					tc.want,
				)
			}
		})
	}
}
