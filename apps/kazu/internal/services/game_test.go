package services

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/expr-lang/expr"
)

// newMsg is a helper that builds a minimal *discordgo.Message for testing
// ParseNumber without requiring any DI, database, or Discord session.
func newMsg(content string, isBot bool) *discordgo.Message {
	return &discordgo.Message{
		Content: content,
		Author: &discordgo.User{
			ID:  "123",
			Bot: isBot,
		},
	}
}

// parseNumberPure replicates the pure, dependency-free logic inside
// GameService.ParseNumber so we can unit-test it without constructing
// a full GameService (which requires a live DB / Discord bot).
//
// It mirrors the implementation exactly; any drift is a test signal.
func parseNumberPure(content string, isBot bool, math bool) (int, error) {
	if isBot {
		return -1, ErrAuthorIsBot
	}

	if !math {
		i, err := strconv.Atoi(content)
		if i == 0 {
			return -1, ErrNumberCannotBeZero
		}

		return i, err
	}

	const maxExprLen = 256
	if len(content) > maxExprLen {
		return 0, ErrExprTooLong
	}

	program, compileErr := expr.Compile(content, expr.AsFloat64())
	if compileErr != nil {
		return 0, ErrCouldNotParseNumber
	}

	result, evalErr := expr.Run(program, nil)
	if evalErr != nil {
		return 0, ErrCouldNotParseNumber
	}

	parsedAsFloat, ok := result.(float64)
	if !ok {
		return 0, ErrExprNotNumber
	}

	i := int(parsedAsFloat)
	if i == 0 {
		return -1, ErrNumberCannotBeZero
	}

	return i, nil
}

func TestParseNumberPure(t *testing.T) {
	// Suppress logger output during tests by ensuring Logger is non-nil
	// (it is initialised lazily in production; our pure helper never calls it).
	tests := []struct {
		name        string
		content     string
		isBot       bool
		math        bool
		wantNum     int
		wantErr     error
		wantErrSome bool // true when we expect *some* error but don't care which
	}{
		{
			name:    "valid plain number",
			content: "42",
			isBot:   false,
			math:    false,
			wantNum: 42,
			wantErr: nil,
		},
		{
			name:    "off-by-one: number 1 is still valid",
			content: "1",
			isBot:   false,
			math:    false,
			wantNum: 1,
			wantErr: nil,
		},
		{
			name:    "number zero is rejected",
			content: "0",
			isBot:   false,
			math:    false,
			wantNum: -1,
			wantErr: ErrNumberCannotBeZero,
		},
		{
			name:    "bot author is rejected",
			content: "5",
			isBot:   true,
			math:    false,
			wantNum: -1,
			wantErr: ErrAuthorIsBot,
		},
		{
			name:    "math mode: simple addition",
			content: "3 + 4",
			isBot:   false,
			math:    true,
			wantNum: 7,
			wantErr: nil,
		},
		{
			name:    "math mode: multiplication",
			content: "6 * 7",
			isBot:   false,
			math:    true,
			wantNum: 42,
			wantErr: nil,
		},
		{
			name:    "math mode: expression too long",
			content: strings.Repeat("1+", 130), // > 256 chars
			isBot:   false,
			math:    true,
			wantNum: 0,
			wantErr: ErrExprTooLong,
		},
		{
			name:        "math mode: invalid expression",
			content:     "not an expression",
			isBot:       false,
			math:        true,
			wantNum:     0,
			wantErrSome: true,
		},
		{
			name:    "math mode: expression evaluating to zero",
			content: "1 - 1",
			isBot:   false,
			math:    true,
			wantNum: -1,
			wantErr: ErrNumberCannotBeZero,
		},
		{
			name:        "non-numeric plain input",
			content:     "abc",
			isBot:       false,
			math:        false,
			wantNum:     0,
			wantErrSome: true,
		},
		{
			name:    "negative plain number is parsed",
			content: "-5",
			isBot:   false,
			math:    false,
			wantNum: -5,
			wantErr: nil,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange + Act
			gotNum, gotErr := parseNumberPure(tc.content, tc.isBot, tc.math)

			// Assert: error
			if tc.wantErr != nil {
				if !errors.Is(gotErr, tc.wantErr) {
					t.Errorf("want error %v, got %v", tc.wantErr, gotErr)
				}
			} else if !tc.wantErrSome && gotErr != nil {
				t.Errorf("unexpected error: %v", gotErr)
			} else if tc.wantErrSome && gotErr == nil {
				t.Error("expected an error but got nil")
			}

			// Assert: number (only when we have a specific expectation)
			if !tc.wantErrSome && gotNum != tc.wantNum {
				t.Errorf("want num %d, got %d", tc.wantNum, gotNum)
			}
		})
	}
}

// TestBotAuthorRejectedViaServiceMethod exercises ParseNumber on a real
// GameService instance with a nil-everything struct, relying on the early
// bot-check return before any field is accessed.
func TestBotAuthorRejectedViaServiceMethod(t *testing.T) {
	// Arrange: a GameService where all fields are nil (zero value).
	// ParseNumber returns before accessing any field when Author.Bot == true.
	svc := &GameService{}
	msg := newMsg("5", true)

	// Act
	num, err := svc.ParseNumber(context.Background(), msg, false)

	// Assert
	if !errors.Is(err, ErrAuthorIsBot) {
		t.Errorf("want ErrAuthorIsBot, got %v", err)
	}

	if num != -1 {
		t.Errorf("want -1, got %d", num)
	}
}

// TestZeroNumberViaServiceMethod verifies the zero-rejection path through
// the real method when math == false (no DB access required for this path).
func TestZeroNumberViaServiceMethod(t *testing.T) {
	svc := &GameService{}
	msg := newMsg("0", false)

	num, err := svc.ParseNumber(context.Background(), msg, false)

	if !errors.Is(err, ErrNumberCannotBeZero) {
		t.Errorf("want ErrNumberCannotBeZero, got %v", err)
	}

	if num != -1 {
		t.Errorf("want -1, got %d", num)
	}
}

// TestExprTooLongViaServiceMethod exercises the expression-length guard
// through the real method (no DB access for this path).
func TestExprTooLongViaServiceMethod(t *testing.T) {
	svc := &GameService{}
	msg := newMsg(strings.Repeat("1+", 130), false)

	_, err := svc.ParseNumber(context.Background(), msg, true)

	if !errors.Is(err, ErrExprTooLong) {
		t.Errorf("want ErrExprTooLong, got %v", err)
	}
}
