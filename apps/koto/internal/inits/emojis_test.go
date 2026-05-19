package inits

import (
	"testing"

	localStatic "jurien.dev/yugen/koto/internal/static"
)

func TestParseEmojiName(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantColor  localStatic.EmojiColor
		wantLetter string
		wantOk     bool
	}{
		{"letter A white", "letterAwhite", localStatic.EmojiColorWhite, "a", true},
		{"letter Z green", "letterZgreen", localStatic.EmojiColorGreen, "z", true},
		{"letter M yellow", "letterMyellow", localStatic.EmojiColorYellow, "m", true},
		{"blank gray", "blankgray", localStatic.EmojiColorGray, localStatic.EmojiLetterBlank, true},
		{"blank white", "blankwhite", localStatic.EmojiColorWhite, localStatic.EmojiLetterBlank, true},
		{"unknown color", "letterAblue", "", "", false},
		{"blank unknown color", "blankblue", "", "", false},
		{"random name", "random-name", "", "", false},
		{"empty", "", "", "", false},
		{"letter only", "letter", "", "", false},
		{"lowercase letter char", "letterawhite", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color, letter, ok := parseEmojiName(tt.input)
			if ok != tt.wantOk {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOk)
			}
			if !tt.wantOk {
				return
			}
			if color != tt.wantColor {
				t.Errorf("color = %q, want %q", color, tt.wantColor)
			}
			if letter != tt.wantLetter {
				t.Errorf("letter = %q, want %q", letter, tt.wantLetter)
			}
		})
	}
}

func TestValidateEmojiTable(t *testing.T) {
	t.Run("valid full table", func(t *testing.T) {
		table := buildFullTable()
		if err := validateEmojiTable(table); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("missing letter", func(t *testing.T) {
		table := buildFullTable()
		delete(table[localStatic.EmojiColorWhite], "a")
		if err := validateEmojiTable(table); err == nil {
			t.Error("expected error for missing entry")
		}
	})

	t.Run("missing blank", func(t *testing.T) {
		table := buildFullTable()
		delete(table[localStatic.EmojiColorGray], localStatic.EmojiLetterBlank)
		if err := validateEmojiTable(table); err == nil {
			t.Error("expected error for missing blank")
		}
	})

	t.Run("extra entry inflates count", func(t *testing.T) {
		table := buildFullTable()
		table[localStatic.EmojiColorGreen]["extra"] = localStatic.EmojiData{Name: "x", ID: "1"}
		if err := validateEmojiTable(table); err == nil {
			t.Error("expected error for inflated count")
		}
	})
}

// buildFullTable constructs a minimal valid 4×27 emoji table for testing.
func buildFullTable() map[localStatic.EmojiColor]map[string]localStatic.EmojiData {
	letters := []string{
		localStatic.EmojiLetterBlank,
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	}
	colors := []localStatic.EmojiColor{
		localStatic.EmojiColorGreen,
		localStatic.EmojiColorYellow,
		localStatic.EmojiColorGray,
		localStatic.EmojiColorWhite,
	}

	table := make(map[localStatic.EmojiColor]map[string]localStatic.EmojiData, len(colors))
	for _, c := range colors {
		table[c] = make(map[string]localStatic.EmojiData, len(letters))
		for _, l := range letters {
			table[c][l] = localStatic.EmojiData{Name: string(c) + l, ID: "1"}
		}
	}
	return table
}
