package utils

import (
	"encoding/json"
	"testing"
)

func TestParseGameMeta_Valid(t *testing.T) {
	raw := json.RawMessage(`{
		"keyboard": {"a": "CORRECT", "b": "WRONG"},
		"word": [
			{"index": 0, "letter": "a", "type": "CORRECT"},
			{"index": 1, "letter": "b", "type": "WRONG"}
		],
		"discovery": {"almost": {"c": 1}, "correct": {"a": 2}}
	}`)

	meta, err := ParseGameMeta(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Keyboard["a"] != GameTypeCorrect {
		t.Errorf(
			"keyboard[a] = %q, want %q",
			meta.Keyboard["a"],
			GameTypeCorrect,
		)
	}

	if meta.Keyboard["b"] != GameTypeWrong {
		t.Errorf("keyboard[b] = %q, want %q", meta.Keyboard["b"], GameTypeWrong)
	}

	if len(meta.Word) != 2 {
		t.Fatalf("word len = %d, want 2", len(meta.Word))
	}

	if meta.Word[0].Letter != "a" || meta.Word[0].Type != GameTypeCorrect {
		t.Errorf("word[0] = %+v, unexpected", meta.Word[0])
	}

	if meta.Discovery.Correct["a"] != 2 {
		t.Errorf(
			"discovery.correct[a] = %d, want 2",
			meta.Discovery.Correct["a"],
		)
	}
}

func TestParseGameMeta_Empty(t *testing.T) {
	meta, err := ParseGameMeta(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta.Keyboard != nil && len(meta.Keyboard) != 0 {
		t.Errorf("expected empty keyboard, got %v", meta.Keyboard)
	}
}

func TestParseGameMeta_Invalid(t *testing.T) {
	_, err := ParseGameMeta(json.RawMessage(`not json`))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGameTypeToEmojiColor(t *testing.T) {
	tests := []struct {
		gameType GameType
		want     string
	}{
		{GameTypeCorrect, "GREEN"},
		{GameTypeAlmost, "YELLOW"},
		{GameTypeWrong, "GRAY"},
		{GameTypeDefault, "WHITE"},
	}

	for _, tc := range tests {
		got, ok := GameTypeToEmojiColor[tc.gameType]
		if !ok {
			t.Errorf("GameTypeToEmojiColor[%q] not found", tc.gameType)
			continue
		}

		if got != tc.want {
			t.Errorf(
				"GameTypeToEmojiColor[%q] = %q, want %q",
				tc.gameType,
				got,
				tc.want,
			)
		}
	}
}
