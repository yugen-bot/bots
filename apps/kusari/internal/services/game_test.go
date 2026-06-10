package services

import (
	"errors"
	"testing"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
)

// newKusariMsg builds a minimal discord.Message for testing ParseWord
// without requiring any DI or network access.
func newKusariMsg(content string, isBot bool) discord.Message {
	return discord.Message{
		Content: content,
		Author: discord.User{
			ID:  snowflake.ID(456),
			Bot: isBot,
		},
	}
}

// ----------------------------------------------------------------------------
// Regex tests – pure, no struct needed
// ----------------------------------------------------------------------------

func TestFirstLetterRegex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"a", true},
		{"Z", true},
		{"!", true},
		{"!a", true}, // leading exclamation mark is allowed by the regex
		{"1", false}, // digit not in [A-Za-z!]
		{"", false},
		{" ", false},
		{"#word", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			// Only test the first character as the service does: string(word[0])
			first := string(tc.input)

			got := firstLetterRegex.MatchString(first)
			if got != tc.want {
				t.Errorf(
					"firstLetterRegex.MatchString(%q) = %v, want %v",
					first,
					got,
					tc.want,
				)
			}
		})
	}
}

func TestLastLetterRegex(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"a", true},
		{"Z", true},
		{"!", false}, // exclamation is NOT in [A-Za-z]
		{"1", false},
		{"", false},
		{" ", false},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			last := tc.input

			got := lastLetterRegex.MatchString(last)
			if got != tc.want {
				t.Errorf(
					"lastLetterRegex.MatchString(%q) = %v, want %v",
					last,
					got,
					tc.want,
				)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// ParseWord tests – exercises the method through a zero-value GameService.
// ParseWord only accesses message.Author.Bot and message.Content, so no DI
// is required for these paths.
// ----------------------------------------------------------------------------

func TestParseWord(t *testing.T) {
	svc := &GameService{} // all fields nil – safe for the paths under test

	tests := []struct {
		name     string
		content  string
		isBot    bool
		wantWord string
		wantErr  error
	}{
		{
			name:     "bot author is rejected",
			content:  "apple",
			isBot:    true,
			wantWord: "",
			wantErr:  ErrAuthorIsBot,
		},
		{
			name:     "valid single word is returned lowercased",
			content:  "Apple",
			isBot:    false,
			wantWord: "apple",
			wantErr:  nil,
		},
		{
			name:     "multiple words returns empty",
			content:  "two words",
			isBot:    false,
			wantWord: "",
			wantErr:  nil,
		},
		{
			name:     "empty content returns empty",
			content:  "",
			isBot:    false,
			wantWord: "",
			wantErr:  nil,
		},
		{
			name:     "word starting with exclamation is stripped",
			content:  "!apple",
			isBot:    false,
			wantWord: "apple",
			wantErr:  nil,
		},
		{
			name:     "word starting with digit is rejected (first letter regex)",
			content:  "1apple",
			isBot:    false,
			wantWord: "",
			wantErr:  nil,
		},
		{
			name:     "word ending with digit is rejected (last letter regex)",
			content:  "apple1",
			isBot:    false,
			wantWord: "",
			wantErr:  nil,
		},
		{
			name:     "all uppercase word is returned lowercased",
			content:  "BANANA",
			isBot:    false,
			wantWord: "banana",
			wantErr:  nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			msg := newKusariMsg(tc.content, tc.isBot)

			// Act
			gotWord, gotErr := svc.ParseWord(msg)

			// Assert
			if !errors.Is(gotErr, tc.wantErr) {
				t.Errorf("want error %v, got %v", tc.wantErr, gotErr)
			}

			if gotWord != tc.wantWord {
				t.Errorf("want word %q, got %q", tc.wantWord, gotWord)
			}
		})
	}
}

// TestGetRandomLetter verifies getRandomLetter returns a non-empty string and
// always returns a value from the expected letter set.
func TestGetRandomLetter(t *testing.T) {
	svc := &GameService{}
	validLetters := map[string]bool{
		"a": true, "b": true, "c": true, "d": true, "e": true,
		"f": true, "g": true, "h": true, "i": true, "j": true,
		"k": true, "l": true, "m": true, "n": true, "o": true,
		"p": true, "q": true, "r": true, "s": true, "t": true,
		"u": true, "v": true, "w": true, "y": true, "z": true,
	}

	for i := 0; i < 50; i++ {
		letter := svc.getRandomLetter()
		if letter == "" {
			t.Fatal("getRandomLetter returned empty string")
		}

		if !validLetters[letter] {
			t.Errorf("getRandomLetter returned unexpected letter %q", letter)
		}
	}
}
