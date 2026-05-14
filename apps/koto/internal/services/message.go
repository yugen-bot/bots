package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/koto/prisma/db"
	"jurien.dev/yugen/shared/config"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type MessageService struct {
	database *db.PrismaClient
	settings *SettingsService
	cfg      *config.Config
	bot      *discordgoplus.Bot
}

func CreateMessageService(container *di.Container) *MessageService {
	utils.Logger.Info("Creating Message Service")
	return &MessageService{
		database: container.Get(sharedStatic.DiDatabase).(*db.PrismaClient),
		settings: container.Get(sharedStatic.DiSettings).(*SettingsService),
		cfg:      container.Get(sharedStatic.DiConfig).(*config.Config),
		bot:      container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

// Create sends or recreates the game embed message.
// isNew=true pings the role (if configured), false skips ping unless not pingOnlyNew.
func (m *MessageService) Create(ctx context.Context, game *db.GameModel, guesses []db.GuessModel, isNew bool) error {
	settings, err := m.settings.GetByGuildID(ctx, game.GuildID)
	if err != nil || settings == nil {
		return err
	}

	channelID, ok := settings.ChannelID()
	if !ok || channelID == "" {
		return nil
	}

	// Delete previous message if exists
	if lastMsgID, ok := game.LastMessageID(); ok && lastMsgID != "" {
		go func() {
			m.bot.ChannelMessageDelete(channelID, lastMsgID) //nolint:errcheck
		}()
	}

	embed, err := m.buildEmbed(game, guesses)
	if err != nil {
		return fmt.Errorf("message: create: build embed: %w", err)
	}

	// Build content (ping role)
	content := ""
	pingRoleID, hasPingRole := settings.PingRoleID()
	if hasPingRole && pingRoleID != "" && (isNew || !settings.PingOnlyNew) {
		content = fmt.Sprintf("<@&%s>", pingRoleID)
	}

	allowedRoles := []string{}
	if hasPingRole && pingRoleID != "" {
		allowedRoles = []string{pingRoleID}
	}

	sentMsg, err := m.bot.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: content,
		Embeds:  []*discordgo.MessageEmbed{embed},
		AllowedMentions: &discordgo.MessageAllowedMentions{
			Users: []string{},
			Roles: allowedRoles,
		},
	})
	if err != nil {
		utils.Logger.Warnf("message: create: send failed for guild %s: %v", game.GuildID, err)
		return nil
	}

	// Update game with new message ID
	_, err = m.database.Game.FindUnique(
		db.Game.ID.Equals(game.ID),
	).Update(
		db.Game.LastMessageID.Set(sentMsg.ID),
	).Exec(ctx)

	return err
}

func (m *MessageService) buildEmbed(game *db.GameModel, guesses []db.GuessModel) (*discordgo.MessageEmbed, error) {
	meta, err := localUtils.ParseGameMeta(json.RawMessage(game.Meta))
	if err != nil {
		return nil, fmt.Errorf("message: build embed: parse meta: %w", err)
	}

	rows := m.buildRows(guesses, game.Status)
	keyboard := m.buildKeyboard(meta)
	info := m.buildGameInfo(game, meta, len(guesses))

	color := localStatic.EmbedColorInProgress
	switch game.Status {
	case db.GameStatusCompleted:
		color = localStatic.EmbedColorSuccess
	case db.GameStatusFailed, db.GameStatusOutOfTime:
		color = localStatic.EmbedColorFail
	}

	footer := utils.CreateEmbedFooter(
		m.bot,
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		m.cfg.OwnerID,
	)

	return &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("Koto #%d", game.Number),
		Color:       color,
		Description: rows + "\n" + keyboard + "\n" + info,
		Footer:      footer,
	}, nil
}

func (m *MessageService) buildRows(guesses []db.GuessModel, status db.GameStatus) string {
	type rowData struct {
		meta      localUtils.GuessMetaSlice
		userID    string
		points    int
		createdAt time.Time
	}

	rows := make([]rowData, 0, len(guesses))
	for _, g := range guesses {
		var guessMeta localUtils.GuessMetaSlice
		json.Unmarshal(g.Meta, &guessMeta) //nolint:errcheck
		rows = append(rows, rowData{
			meta:      guessMeta,
			userID:    g.UserID,
			points:    g.Points,
			createdAt: g.CreatedAt,
		})
	}

	// Pad to MaxGuesses rows with blanks for IN_PROGRESS
	if len(rows) < localStatic.MaxGuesses && status == db.GameStatusInProgress {
		for i := localStatic.MaxGuesses - len(rows); i > 0; i-- {
			blank := make(localUtils.GuessMetaSlice, localStatic.WordLength)
			for j := 0; j < localStatic.WordLength; j++ {
				blank[j] = localUtils.GuessMeta{Type: localUtils.GameTypeDefault, Letter: "blank"}
			}
			rows = append(rows, rowData{
				meta:      blank,
				userID:    m.bot.State.User.ID,
				points:    0,
				createdAt: time.Now(),
			})
		}
	}

	// Sort by createdAt ascending
	sort.Slice(rows, func(i, j int) bool {
		return rows[i].createdAt.Before(rows[j].createdAt)
	})

	receivedBonus := map[string]bool{}
	var sb strings.Builder
	for i, row := range rows {
		sb.WriteString(localStatic.AsciiNumbers[i+1])
		for _, letterMeta := range row.meta {
			emojiColor := localUtils.GameTypeToEmojiColor[letterMeta.Type]
			sb.WriteString(localUtils.GetEmoji(emojiColor, letterMeta.Letter))
		}

		if row.userID != m.bot.State.User.ID {
			bonus := ""
			if status == db.GameStatusCompleted && !receivedBonus[row.userID] {
				bonus = " (+2)"
			}
			sb.WriteString(fmt.Sprintf(" <@%s> **+%d%s**", row.userID, row.points, bonus))
		}
		sb.WriteString("\n")
		receivedBonus[row.userID] = true
	}

	return sb.String()
}

func (m *MessageService) buildKeyboard(meta *localUtils.GameMeta) string {
	rows := [][]interface{}{
		{"q", "w", "e", "r", "t", "y", "u", "i", "o", "p"},
		{"a", "s", "d", "f", "g", "h", "j", "k", "l", nil},
		{nil, "z", "x", "c", "v", "b", "n", "m", nil},
	}

	var sb strings.Builder
	for _, row := range rows {
		for _, item := range row {
			if item == nil {
				sb.WriteString(localUtils.GetEmoji("GRAY", "blank"))
			} else {
				letter := item.(string)
				gameType, exists := meta.Keyboard[letter]
				if !exists {
					gameType = localUtils.GameTypeDefault
				}
				color := localUtils.GameTypeToEmojiColor[gameType]
				sb.WriteString(localUtils.GetEmoji(color, letter))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func (m *MessageService) buildGameInfo(game *db.GameModel, _ *localUtils.GameMeta, guessCount int) string {
	footer := "Don't know how to play? Use the /tutorial commands for detailed instructions."

	envSuffix := ""
	if os.Getenv("ENV") != "production" {
		envSuffix = fmt.Sprintf("\nDevelopment mode: **%s**", game.Word)
	}

	switch game.Status {
	case db.GameStatusCompleted:
		nextKoto := fmt.Sprintf("Next koto <t:%d:R>", game.EndingAt.Unix())
		return fmt.Sprintf("\nGood job! Everyone who participated gets **+2** points!\n%s\n\n%s%s", nextKoto, footer, envSuffix)
	case db.GameStatusFailed:
		nextKoto := fmt.Sprintf("Next koto <t:%d:R>", game.EndingAt.Unix())
		return fmt.Sprintf("\nOut of guesses, The correct word was **%s**!\n%s\n\n%s%s", strings.ToUpper(game.Word), nextKoto, footer, envSuffix)
	case db.GameStatusOutOfTime:
		return fmt.Sprintf("\nTime's up! The correct word was **%s**!\n\n%s%s", strings.ToUpper(game.Word), footer, envSuffix)
	default:
		remaining := localStatic.MaxGuesses - guessCount
		var timerLine string
		if game.EndingAt.Year() == 3000 {
			timerLine = "Timer will start after first guess"
		} else {
			timerLine = fmt.Sprintf("Time runs out <t:%d:R>", game.EndingAt.Unix())
		}
		return fmt.Sprintf("\n%d guesses remaining\n%s\n\n%s%s", remaining, timerLine, footer, envSuffix)
	}
}
