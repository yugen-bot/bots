package utils

import (
	"strings"
	"testing"
)

func TestSplitBySentence_Empty(t *testing.T) {
	if got := SplitBySentence("", 100); got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestSplitBySentence_ShortContent(t *testing.T) {
	content := "Hello world."
	chunks := SplitBySentence(content, 100)
	if len(chunks) != 1 || chunks[0] != content {
		t.Errorf("expected single chunk %q, got %v", content, chunks)
	}
}

func TestSplitBySentence_MultiSentence(t *testing.T) {
	content := "First sentence. Second sentence. Third sentence."
	chunks := SplitBySentence(content, 30)
	for _, c := range chunks {
		if len(c) > 30 {
			t.Errorf("chunk %q exceeds limit of 30", c)
		}
	}
	joined := strings.Join(chunks, " ")
	for _, word := range []string{"First", "Second", "Third"} {
		if !strings.Contains(joined, word) {
			t.Errorf("word %q missing from result", word)
		}
	}
}

func TestSplitBySentence_ExclamationAndQuestion(t *testing.T) {
	content := "Hello! Are you there? Yes I am."
	chunks := SplitBySentence(content, 20)
	for _, c := range chunks {
		if len(c) > 20 {
			t.Errorf("chunk %q exceeds limit of 20", c)
		}
	}
}

func TestSplitBySentence_NewlineBoundary(t *testing.T) {
	content := "Line one\nLine two\nLine three"
	chunks := SplitBySentence(content, 15)
	for _, c := range chunks {
		if len(c) > 15 {
			t.Errorf("chunk %q exceeds limit of 15", c)
		}
	}
}

func TestSplitBySentence_OversizedSentenceFallsBackToWord(t *testing.T) {
	word := strings.Repeat("x", 5)
	// Build sentence > 20 chars with word boundaries
	content := word + " " + word + " " + word + " " + word + " " + word
	chunks := SplitBySentence(content, 20)
	for _, c := range chunks {
		if len(c) > 20 {
			t.Errorf("chunk %q exceeds limit of 20", c)
		}
	}
}

func TestSplitBySentence_OversizedSentenceNoSpaceHardCut(t *testing.T) {
	content := strings.Repeat("a", 50)
	chunks := SplitBySentence(content, 20)
	for _, c := range chunks {
		if len(c) > 20 {
			t.Errorf("chunk %q exceeds limit of 20", c)
		}
	}
	// All content preserved
	if strings.Join(chunks, "") != content {
		t.Errorf("content not fully preserved after hard-cut")
	}
}
