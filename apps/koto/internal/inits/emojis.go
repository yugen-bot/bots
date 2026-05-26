package inits

import (
	"fmt"
	"strings"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	localStatic "jurien.dev/yugen/koto/internal/static"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

const expectedEmojiCount = 4 * 27 // colors × (blank + 26 letters)

var knownColors = map[string]localStatic.EmojiColor{
	"green":  localStatic.EmojiColorGreen,
	"yellow": localStatic.EmojiColorYellow,
	"gray":   localStatic.EmojiColorGray,
	"white":  localStatic.EmojiColorWhite,
}

// InitEmojis fetches application emojis from Discord and populates the emoji table.
func InitEmojis(container *di.Container) error {
	bot := container.Get(static.DiBot).(*discordgoplus.Bot)
	cfg := container.Get(static.DiConfig).(*config.Config)

	emojis, err := bot.ApplicationEmojis(cfg.DiscordAppID)
	if err != nil {
		return fmt.Errorf("emojis: list application emojis: %w", err)
	}

	table := make(
		map[localStatic.EmojiColor]map[string]localStatic.EmojiData,
		len(knownColors),
	)
	totalEmojis := 0

	for _, c := range knownColors {
		table[c] = make(map[string]localStatic.EmojiData, 27)
	}

	for _, e := range emojis {
		color, letter, ok := parseEmojiName(e.Name)
		if !ok {
			continue
		}

		table[color][letter] = localStatic.EmojiData{Name: e.Name, ID: e.ID}
		totalEmojis++
	}

	if err := validateEmojiTable(table); err != nil {
		return fmt.Errorf("emojis: validate: %w", err)
	}

	localStatic.SetEmojiTable(table)

	utils.Logger.Infof("Loaded %d emojis", totalEmojis)

	return nil
}

func parseEmojiName(name string) (localStatic.EmojiColor, string, bool) {
	if rest, ok := strings.CutPrefix(name, "blank"); ok {
		color, known := knownColors[strings.ToLower(rest)]
		if !known {
			return "", "", false
		}

		return color, localStatic.EmojiLetterBlank, true
	}

	if rest, ok := strings.CutPrefix(name, "letter"); ok && len(rest) >= 2 {
		letterCh := rest[0]
		if letterCh < 'A' || letterCh > 'Z' {
			return "", "", false
		}

		color, known := knownColors[strings.ToLower(rest[1:])]
		if !known {
			return "", "", false
		}

		return color, strings.ToLower(string(letterCh)), true
	}

	return "", "", false
}

func validateEmojiTable(
	table map[localStatic.EmojiColor]map[string]localStatic.EmojiData,
) error {
	letters := []string{
		localStatic.EmojiLetterBlank,
		"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m",
		"n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z",
	}

	total := 0

	for color, sub := range table {
		for _, l := range letters {
			if _, ok := sub[l]; !ok {
				return fmt.Errorf("missing %s/%s", color, l)
			}
		}

		total += len(sub)
	}

	if total != expectedEmojiCount {
		return fmt.Errorf(
			"expected %d emojis, got %d",
			expectedEmojiCount,
			total,
		)
	}

	return nil
}
