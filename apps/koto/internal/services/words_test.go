package services

import (
	"testing"

	"jurien.dev/yugen/shared/utils"
)

var testWordsSvc *WordsService

func getWordsService(t *testing.T) *WordsService {
	t.Helper()

	if testWordsSvc == nil {
		utils.CreateLogger("test")

		testWordsSvc = CreateWordsService()
	}

	return testWordsSvc
}

func TestCreateWordsService_HasWords(t *testing.T) {
	svc := getWordsService(t)

	if svc.Amount == 0 {
		t.Error("WordsService has no words")
	}
}

func TestExists_ValidWord(t *testing.T) {
	svc := getWordsService(t)

	// Pick a word from the service's own map so the test is data-independent
	var validWord string

	for _, words := range svc.wordsByLetter {
		if len(words) > 0 {
			validWord = words[0]
			break
		}
	}

	if validWord == "" {
		t.Skip("no words available")
	}

	if !svc.Exists(validWord) {
		t.Errorf("Exists(%q) = false, want true", validWord)
	}
}

func TestExists_InvalidWord(t *testing.T) {
	svc := getWordsService(t)

	if svc.Exists("zzzzzzzzzzz") {
		t.Error("Exists(zzzzzzzzzzz) should be false")
	}
}

func TestExists_EmptyString(t *testing.T) {
	svc := getWordsService(t)

	if svc.Exists("") {
		t.Error("Exists(\"\") should be false")
	}
}

func TestGetRandom_ReturnsSixLetterWord(t *testing.T) {
	svc := getWordsService(t)

	word := svc.GetRandom(nil, false)
	if len([]rune(word)) != 6 {
		t.Errorf("GetRandom returned %q (len %d), want 6 letters", word, len([]rune(word)))
	}
}

func TestGetRandom_RespectsIgnoreList(t *testing.T) {
	svc := getWordsService(t)

	// Collect all words from one letter bucket to ignore them, then verify
	// we still get a result (from another bucket)
	var ignored []string

	for _, words := range svc.wordsByLetter {
		ignored = append(ignored, words...)
		break
	}

	word := svc.GetRandom(ignored, false)
	if word == "" {
		t.Error("GetRandom returned empty string")
	}
}

func TestGetRandom_HardMode(t *testing.T) {
	svc := getWordsService(t)

	word := svc.GetRandom(nil, true)
	if word == "" {
		t.Error("GetRandom(hard=true) returned empty string")
	}
}
