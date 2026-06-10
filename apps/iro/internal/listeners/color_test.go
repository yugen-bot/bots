package listeners

import (
	"testing"
)

const (
	testColorAAA = "#AAAAAA"
	testColorBBB = "#BBBBBB"
)

// ----------------------------------------------------------------------------
// rxColorHex regex tests
// ----------------------------------------------------------------------------

func TestRxColorHex(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Valid: 6-digit hex with leading '#'
		{name: "6-digit hex with hash", input: "#1A2B3C", want: true},
		// Valid: 6-digit hex without leading '#'
		{name: "6-digit hex without hash", input: "1A2B3C", want: true},
		// Valid: lowercase hex without hash
		{name: "6-digit lowercase hex", input: "aabbcc", want: true},
		// Valid: mixed case with hash
		{name: "mixed case with hash", input: "#aAbBcC", want: true},
		// Valid: 8-digit hex (RGBA) with hash
		{name: "8-digit hex with hash", input: "#1A2B3CFF", want: true},
		// Valid: 8-digit hex without hash
		{name: "8-digit hex without hash", input: "1A2B3CFF", want: true},
		// Invalid: 3-digit hex does not satisfy {6,8} quantifier
		{name: "3-digit hex with hash", input: "#FFF", want: false},
		// Invalid: 3-digit hex without hash
		{name: "3-digit hex without hash", input: "FFF", want: false},
		// Invalid: contains non-hex character
		{name: "invalid char G", input: "#GGGGGG", want: false},
		// Invalid: empty string
		{name: "empty string", input: "", want: false},
		// Invalid: only spaces
		{name: "spaces", input: "      ", want: false},
		// Invalid: 5-digit hex (below minimum 6)
		{name: "5-digit hex", input: "#FFFFF", want: false},
		// Invalid: 9-digit hex (above maximum 8)
		{name: "9-digit hex", input: "#FFFFFFFFF", want: false},
		// Valid edge: exactly 6 zeros
		{name: "6 zeros no hash", input: "000000", want: true},
		// Valid edge: exactly 6 nines
		{name: "6 nines no hash", input: "999999", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			got := rxColorHex.MatchString(tc.input)

			// Assert
			if got != tc.want {
				t.Errorf(
					"rxColorHex.MatchString(%q) = %v, want %v",
					tc.input,
					got,
					tc.want,
				)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// appendIfUnique tests
// ----------------------------------------------------------------------------

func TestAppendIfUnique(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		elem  string
		want  []string
	}{
		{
			name:  "append to empty slice",
			slice: []string{},
			elem:  "#FFFFFF",
			want:  []string{"#FFFFFF"},
		},
		{
			name:  "append unique element",
			slice: []string{testColorAAA},
			elem:  testColorBBB,
			want:  []string{testColorAAA, testColorBBB},
		},
		{
			name:  "duplicate is not appended",
			slice: []string{testColorAAA, testColorBBB},
			elem:  testColorAAA,
			want:  []string{testColorAAA, testColorBBB},
		},
		{
			name:  "case-sensitive: different cases are treated as distinct",
			slice: []string{"#aaaaaa"},
			elem:  testColorAAA,
			want:  []string{"#aaaaaa", testColorAAA},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange: copy slice to avoid mutation between sub-tests
			input := make([]string, len(tc.slice))
			copy(input, tc.slice)

			// Act
			got := appendIfUnique(input, tc.elem)

			// Assert
			if len(got) != len(tc.want) {
				t.Fatalf(
					"len(got) = %d, want %d; got = %v",
					len(got),
					len(tc.want),
					got,
				)
			}

			for i, v := range tc.want {
				if got[i] != v {
					t.Errorf("got[%d] = %q, want %q", i, got[i], v)
				}
			}
		})
	}
}
