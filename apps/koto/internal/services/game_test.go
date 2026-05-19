package services

import (
	"testing"
	"time"

	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/koto/prisma/db"
)

// Zero-value *GameService is safe for checkWord, createBaseState, checkCooldown,
// and roundToNearestMinute — none of those methods access any GameService fields.

// ─────────────────────────────────────────────────────────────────────────────
// createBaseState
// ─────────────────────────────────────────────────────────────────────────────

func TestCreateBaseState_Length(t *testing.T) {
	svc := &GameService{}
	word := "marble"
	state := svc.createBaseState(word)

	want := len([]rune(word))
	if len(state.Word) != want {
		t.Errorf("Word len = %d, want %d", len(state.Word), want)
	}
}

func TestCreateBaseState_AllWrong(t *testing.T) {
	svc := &GameService{}
	state := svc.createBaseState("marble")

	for i, ws := range state.Word {
		if ws.Type != localUtils.GameTypeWrong {
			t.Errorf("Word[%d].Type = %q, want WRONG", i, ws.Type)
		}
	}
}

func TestCreateBaseState_DiscoveryZero(t *testing.T) {
	svc := &GameService{}
	state := svc.createBaseState("aab")

	for letter, v := range state.Discovery.Almost {
		if v != 0 {
			t.Errorf("Discovery.Almost[%q] = %d, want 0", letter, v)
		}
	}

	for letter, v := range state.Discovery.Correct {
		if v != 0 {
			t.Errorf("Discovery.Correct[%q] = %d, want 0", letter, v)
		}
	}
}

func TestCreateBaseState_EmptyKeyboard(t *testing.T) {
	svc := &GameService{}
	state := svc.createBaseState("marble")

	if len(state.Keyboard) != 0 {
		t.Errorf("Keyboard len = %d, want 0", len(state.Keyboard))
	}
}

func TestCreateBaseState_LetterIndices(t *testing.T) {
	svc := &GameService{}
	state := svc.createBaseState("abc")

	for i, ws := range state.Word {
		if ws.Index != i {
			t.Errorf("Word[%d].Index = %d, want %d", i, ws.Index, i)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// roundToNearestMinute
// ─────────────────────────────────────────────────────────────────────────────

func TestRoundToNearestMinute(t *testing.T) {
	base := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name  string
		input time.Time
		want  time.Time
	}{
		{
			name:  "already on minute",
			input: base,
			want:  base,
		},
		{
			name:  "29s past rounds down",
			input: base.Add(29 * time.Second),
			want:  base,
		},
		{
			name:  "30s past rounds up",
			input: base.Add(30 * time.Second),
			want:  base.Add(time.Minute),
		},
		{
			name:  "59s past rounds up",
			input: base.Add(59 * time.Second),
			want:  base.Add(time.Minute),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := roundToNearestMinute(tc.input)
			if !got.Equal(tc.want) {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// checkWord helpers
// ─────────────────────────────────────────────────────────────────────────────

// freshState returns a clean game state for word, using createBaseState.
func freshState(word string) *localUtils.GameMeta {
	return (&GameService{}).createBaseState(word)
}

// ─────────────────────────────────────────────────────────────────────────────
// checkWord
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckWord_PerfectMatch(t *testing.T) {
	svc := &GameService{}
	word := "marble"
	meta, guessed, points, _ := svc.checkWord(word, word, freshState(word))

	if !guessed {
		t.Error("guessed = false, want true for exact match")
	}

	if points <= 0 {
		t.Errorf("points = %d, want > 0 for exact match", points)
	}

	for i, m := range meta {
		if m.Type != localUtils.GameTypeCorrect {
			t.Errorf("meta[%d].Type = %q, want CORRECT", i, m.Type)
		}
	}
}

func TestCheckWord_NoOverlap(t *testing.T) {
	svc := &GameService{}
	word := "abcdef"
	guess := "ghijkl"
	meta, guessed, points, _ := svc.checkWord(word, guess, freshState(word))

	if guessed {
		t.Error("guessed = true, want false for no-overlap guess")
	}

	if points != 0 {
		t.Errorf("points = %d, want 0", points)
	}

	for i, m := range meta {
		if m.Type != localUtils.GameTypeWrong {
			t.Errorf("meta[%d].Type = %q, want WRONG", i, m.Type)
		}
	}
}

func TestCheckWord_AllAlmost(t *testing.T) {
	svc := &GameService{}
	// "ab" → "ba": both letters present but swapped positions
	meta, guessed, points, _ := svc.checkWord("ab", "ba", freshState("ab"))

	if guessed {
		t.Error("guessed = true, want false")
	}

	if points <= 0 {
		t.Errorf("points = %d, want > 0 for ALMOST letters", points)
	}

	for i, m := range meta {
		if m.Type != localUtils.GameTypeAlmost {
			t.Errorf("meta[%d].Type = %q, want ALMOST", i, m.Type)
		}
	}
}

func TestCheckWord_MixedCorrectAndAlmost(t *testing.T) {
	svc := &GameService{}
	// word="abc" guess="bac":
	//   pos 0: 'b' is in word but wrong position → ALMOST
	//   pos 1: 'a' is in word but wrong position → ALMOST
	//   pos 2: 'c' matches exactly → CORRECT
	meta, guessed, _, _ := svc.checkWord("abc", "bac", freshState("abc"))

	if guessed {
		t.Error("guessed = true, want false")
	}

	want := []localUtils.GameType{
		localUtils.GameTypeAlmost,
		localUtils.GameTypeAlmost,
		localUtils.GameTypeCorrect,
	}

	for i, wantType := range want {
		if meta[i].Type != wantType {
			t.Errorf("meta[%d].Type = %q, want %q", i, meta[i].Type, wantType)
		}
	}
}

func TestCheckWord_DuplicateLetterInGuess_OnlyOneCredited(t *testing.T) {
	svc := &GameService{}
	// word has a single 'a' at position 0; guess repeats 'a' everywhere
	word := "abcdef"
	guess := "aaaaaa"
	meta, _, _, _ := svc.checkWord(word, guess, freshState(word))

	// Position 0: exact match → CORRECT
	if meta[0].Type != localUtils.GameTypeCorrect {
		t.Errorf("meta[0].Type = %q, want CORRECT", meta[0].Type)
	}

	// Positions 1-5: no unmatched 'a' remaining → WRONG
	for i := 1; i < len(meta); i++ {
		if meta[i].Type != localUtils.GameTypeWrong {
			t.Errorf("meta[%d].Type = %q, want WRONG (extra duplicate)", i, meta[i].Type)
		}
	}
}

func TestCheckWord_DiscoveryPoints_NewVsRepeat(t *testing.T) {
	svc := &GameService{}
	word := "abc"
	// guess "axc": 'a' and 'c' are correct, 'x' is wrong
	state := freshState(word)
	meta1, _, _, updatedState := svc.checkWord(word, "axc", state)

	// First discovery of 'a' in correct position → 2 pts
	if meta1[0].Points != 2 {
		t.Errorf("first CORRECT discovery: points = %d, want 2", meta1[0].Points)
	}

	// Second guess on the same (already updated) state: 'a' already known → 0 pts
	meta2, _, _, _ := svc.checkWord(word, "axc", updatedState)
	if meta2[0].Points != 0 {
		t.Errorf("repeat CORRECT: points = %d, want 0", meta2[0].Points)
	}
}

func TestCheckWord_KeyboardUpdated(t *testing.T) {
	svc := &GameService{}
	word := "abc"
	state := freshState(word)
	_, _, _, updatedState := svc.checkWord(word, "axc", state)

	// 'a' in correct position → keyboard CORRECT
	if updatedState.Keyboard["a"] != localUtils.GameTypeCorrect {
		t.Errorf("keyboard[a] = %q, want CORRECT", updatedState.Keyboard["a"])
	}

	// 'x' not in word → keyboard WRONG
	if updatedState.Keyboard["x"] != localUtils.GameTypeWrong {
		t.Errorf("keyboard[x] = %q, want WRONG", updatedState.Keyboard["x"])
	}
}

func TestCheckWord_AlmostDoesNotOverwriteCorrectInKeyboard(t *testing.T) {
	svc := &GameService{}
	// word="aab" guess1="axb": 'a' correct at pos 0, pos 1 wrong
	// guess2 on updated state="abx": pos 0 'a' correct, pos 1 'b' almost → keyboard['a'] must stay CORRECT
	word := "aab"
	state := freshState(word)
	_, _, _, state = svc.checkWord(word, "axb", state)

	// Now 'a' is CORRECT in keyboard; next guess has 'a' as ALMOST at pos 1
	_, _, _, state = svc.checkWord(word, "bax", state)

	if state.Keyboard["a"] != localUtils.GameTypeCorrect {
		t.Errorf(
			"keyboard[a] = %q after ALMOST, want CORRECT (must not downgrade)",
			state.Keyboard["a"],
		)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// checkCooldown helpers
// ─────────────────────────────────────────────────────────────────────────────

// makeGuess builds a minimal GuessModel with only the fields checkCooldown reads.
func makeGuess(userID string, age time.Duration) db.GuessModel {
	return db.GuessModel{
		InnerGuess: db.InnerGuess{
			UserID:    userID,
			CreatedAt: time.Now().Add(-age),
		},
	}
}

// makeSettings builds a SettingsModel with only cooldown-related fields set.
func makeSettings(cooldown int, b2bEnabled bool, b2bCooldown int) *db.SettingsModel {
	return &db.SettingsModel{
		InnerSettings: db.InnerSettings{
			Cooldown:                 cooldown,
			EnableBackToBackCooldown: b2bEnabled,
			BackToBackCooldown:       b2bCooldown,
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// checkCooldown
// ─────────────────────────────────────────────────────────────────────────────

func TestCheckCooldown_NoGuesses(t *testing.T) {
	svc := &GameService{}
	result := svc.checkCooldown("user1", nil, makeSettings(60, false, 0))
	if result.Hit || result.RepeatHit {
		t.Error("want no cooldown for empty guesses slice")
	}
}

func TestCheckCooldown_UserNotInGuesses(t *testing.T) {
	svc := &GameService{}
	guesses := []db.GuessModel{makeGuess("other", 10*time.Second)}
	result := svc.checkCooldown("user1", guesses, makeSettings(60, false, 0))
	if result.Hit || result.RepeatHit {
		t.Error("want no cooldown when user has made no guesses")
	}
}

func TestCheckCooldown_WithinCooldown(t *testing.T) {
	svc := &GameService{}
	// user1 guessed 30s ago; cooldown = 60s → still active
	guesses := []db.GuessModel{makeGuess("user1", 30*time.Second)}
	result := svc.checkCooldown("user1", guesses, makeSettings(60, false, 0))
	if !result.Hit {
		t.Error("want Hit = true when last guess is within cooldown window")
	}
}

func TestCheckCooldown_OutsideCooldown(t *testing.T) {
	svc := &GameService{}
	// user1 guessed 90s ago; cooldown = 60s → expired
	guesses := []db.GuessModel{makeGuess("user1", 90*time.Second)}
	result := svc.checkCooldown("user1", guesses, makeSettings(60, false, 0))
	if result.Hit || result.RepeatHit {
		t.Error("want no cooldown when last guess is outside cooldown window")
	}
}

func TestCheckCooldown_BackToBack_Hit(t *testing.T) {
	svc := &GameService{}
	// user1 is the only (and therefore last) guesser; within both windows
	g := makeGuess("user1", 10*time.Second)
	guesses := []db.GuessModel{g}
	result := svc.checkCooldown("user1", guesses, makeSettings(120, true, 60))
	if !result.RepeatHit {
		t.Error("want RepeatHit = true when user was last guesser within b2b window")
	}
}

func TestCheckCooldown_BackToBack_Disabled(t *testing.T) {
	svc := &GameService{}
	g := makeGuess("user1", 10*time.Second)
	guesses := []db.GuessModel{g}
	// b2b is disabled even though user was last guesser within the window
	result := svc.checkCooldown("user1", guesses, makeSettings(120, false, 60))
	if result.RepeatHit {
		t.Error("want RepeatHit = false when b2b cooldown is disabled")
	}
}

func TestCheckCooldown_BackToBack_NotLastGuesser(t *testing.T) {
	svc := &GameService{}
	// other-user was the most recent guesser; user1 guessed before them
	guesses := []db.GuessModel{
		makeGuess("other", 5*time.Second),
		makeGuess("user1", 15*time.Second),
	}
	result := svc.checkCooldown("user1", guesses, makeSettings(120, true, 60))
	if result.RepeatHit {
		t.Error("want RepeatHit = false when another user was the last guesser")
	}
}

func TestCheckCooldown_BothHit(t *testing.T) {
	svc := &GameService{}
	// user1 guessed 10s ago and was last; cooldown=120s, b2b=60s → both active
	g := makeGuess("user1", 10*time.Second)
	guesses := []db.GuessModel{g}
	result := svc.checkCooldown("user1", guesses, makeSettings(120, true, 60))
	if !result.Hit {
		t.Error("want Hit = true")
	}
	if !result.RepeatHit {
		t.Error("want RepeatHit = true")
	}
}

func TestCheckCooldown_RepeatHitOnly(t *testing.T) {
	svc := &GameService{}
	// user1 guessed 70s ago; regular cooldown=60s (expired), b2b=120s (still active)
	// user1 was the last guesser → RepeatHit only
	g := makeGuess("user1", 70*time.Second)
	guesses := []db.GuessModel{g}
	result := svc.checkCooldown("user1", guesses, makeSettings(60, true, 120))
	if result.Hit {
		t.Error("want Hit = false (regular cooldown expired)")
	}
	if !result.RepeatHit {
		t.Error("want RepeatHit = true (b2b cooldown still active)")
	}
}
