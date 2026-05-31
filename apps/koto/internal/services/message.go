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

	"jurien.dev/yugen/koto/internal/ent"
	entgame "jurien.dev/yugen/koto/internal/ent/game"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/config"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type MessageService struct {
	database *ent.Client
	settings *SettingsService
	cfg      *config.Config
	bot      *discordgoplus.Bot
}

func CreateMessageService(container *di.Container) *MessageService {
	utils.Logger.Info("Creating Message Service")

	return &MessageService{
		database: container.Get(sharedStatic.DiDatabase).(*ent.Client),
		settings: container.Get(sharedStatic.DiSettings).(*SettingsService),
		cfg:      container.Get(sharedStatic.DiConfig).(*config.Config),
		bot:      container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
	}
}

// Create sends or recreates the game embed message.
// isNew=true pings the role (if configured), false skips ping unless not pingOnlyNew.
func (m *MessageService) Create(
	ctx context.Context,
	currentGame *ent.Game,
	guesses []*ent.Guess,
	isNew bool,
) error {
	guildSettings, err := m.settings.GetByGuildID(ctx, currentGame.GuildID)

	if err != nil || guildSettings == nil {
		return err
	}

	if guildSettings.ChannelID == nil || *guildSettings.ChannelID == "" {
		return nil
	}

	channelID := *guildSettings.ChannelID

	// Delete previous message if exists
	if currentGame.LastMessageID != nil && *currentGame.LastMessageID != "" {
		utils.LogIfErr(
			utils.Logger,
			"channel-message-delete",
			m.bot.ChannelMessageDelete(channelID, *currentGame.LastMessageID),
		)
	}

	embed, err := m.buildEmbed(currentGame, guesses, guildSettings)
	if err != nil {
		return fmt.Errorf("message: create: build embed: %w", err)
	}

	// Build content (ping role)
	content := ""

	if guildSettings.PingRoleID != nil && *guildSettings.PingRoleID != "" &&
		(isNew || !guildSettings.PingOnlyNew) {
		content = fmt.Sprintf("<@&%s>", *guildSettings.PingRoleID)
	}

	allowedRoles := []string{}
	if guildSettings.PingRoleID != nil && *guildSettings.PingRoleID != "" {
		allowedRoles = []string{*guildSettings.PingRoleID}
	}

	meta, _ := localUtils.ParseGameMeta(json.RawMessage(currentGame.Meta))

	var components []discordgo.MessageComponent
	if currentGame.Status == entgame.StatusIN_PROGRESS && meta != nil && meta.CanHint {
		components = []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: fmt.Sprintf("GAME_HINT/%d", currentGame.ID),
						Style:    discordgo.SecondaryButton,
						Label:    "Hint",
						Emoji:    &discordgo.ComponentEmoji{Name: "💡"},
					},
				},
			},
		}
	}

	sentMsg, err := m.bot.ChannelMessageSendComplex(
		channelID,
		&discordgo.MessageSend{
			Content:    content,
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
			AllowedMentions: &discordgo.MessageAllowedMentions{
				Users: []string{},
				Roles: allowedRoles,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("message: create: send: %w", err)
	}

	// Update game with new message ID
	_, err = m.database.Game.UpdateOneID(currentGame.ID).
		SetLastMessageID(sentMsg.ID).
		Save(ctx)

	return err
}

func (m *MessageService) buildEmbed(
	currentGame *ent.Game,
	guesses []*ent.Guess,
	guildSettings *ent.Settings,
) (*discordgo.MessageEmbed, error) {
	meta, err := localUtils.ParseGameMeta(json.RawMessage(currentGame.Meta))
	if err != nil {
		return nil, fmt.Errorf("message: build embed: parse meta: %w", err)
	}

	rows := m.buildRows(guesses, currentGame.Status)
	keyboard := m.buildKeyboard(meta)
	info := m.buildGameInfo(currentGame, meta, len(guesses), guildSettings)

	color := localStatic.EmbedColorInProgress

	switch currentGame.Status {
	case entgame.StatusCOMPLETED:
		color = localStatic.EmbedColorSuccess
	case entgame.StatusFAILED, entgame.StatusOUT_OF_TIME:
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
		Title:       fmt.Sprintf("Koto #%d", currentGame.Number),
		Color:       color,
		Description: rows + "\n" + keyboard + "\n" + info,
		Footer:      footer,
	}, nil
}

func (m *MessageService) buildRows(
	guesses []*ent.Guess,
	status entgame.Status,
) string {
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

	b, _ := m.bot.ShardByShardID(0)

	// Pad to MaxGuesses rows with blanks for IN_PROGRESS
	if len(rows) < localStatic.MaxGuesses && status == entgame.StatusIN_PROGRESS {
		for i := localStatic.MaxGuesses - len(rows); i > 0; i-- {
			blank := make(localUtils.GuessMetaSlice, localStatic.WordLength)

			for j := 0; j < localStatic.WordLength; j++ {
				blank[j] = localUtils.GuessMeta{
					Type:   localUtils.GameTypeDefault,
					Letter: "blank",
				}
			}

			rows = append(rows, rowData{
				meta:      blank,
				userID:    b.State.User.ID,
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

		if row.userID != b.State.User.ID {
			bonus := ""
			if status == entgame.StatusCOMPLETED && !receivedBonus[row.userID] {
				bonus = " (+2)"
			}

			fmt.Fprintf(&sb, " <@%s> **+%d%s**", row.userID, row.points, bonus)
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

func (m *MessageService) buildGameInfo(
	currentGame *ent.Game,
	_ *localUtils.GameMeta,
	guessCount int,
	guildSettings *ent.Settings,
) string {
	footer := "Don't know how to play? Use the /tutorial commands for detailed instructions."

	envSuffix := ""
	if os.Getenv("ENV") != "production" {
		envSuffix = fmt.Sprintf("\nDevelopment mode: **%s**", currentGame.Word)
	}

	nextKoto := ""
	nextAt := currentGame.CreatedAt.Add(
		time.Duration(guildSettings.Frequency) * time.Minute,
	)

	if !guildSettings.AutoStart {
		nextKoto = fmt.Sprintf("\nNext koto <t:%d:R>", nextAt.Unix())
	}

	switch currentGame.Status {
	case entgame.StatusCOMPLETED:
		return fmt.Sprintf(
			"\nGood job! Everyone who participated gets **+2** points!%s\n\n%s%s",
			nextKoto,
			footer,
			envSuffix,
		)
	case entgame.StatusFAILED:
		return fmt.Sprintf(
			"\nOut of guesses, The correct word was **%s**!%s\n\n%s%s",
			strings.ToUpper(currentGame.Word),
			nextKoto,
			footer,
			envSuffix,
		)
	case entgame.StatusOUT_OF_TIME:
		return fmt.Sprintf(
			"\nTime's up! The correct word was **%s**!%s\n\n%s%s",
			strings.ToUpper(currentGame.Word),
			nextKoto,
			footer,
			envSuffix,
		)
	default:
		remaining := localStatic.MaxGuesses - guessCount

		var timerLine string
		if currentGame.EndingAt.Year() == 3000 {
			timerLine = "Timer will start after first guess"
		} else {
			timerLine = fmt.Sprintf(
				"Time runs out <t:%d:R>",
				currentGame.EndingAt.Unix(),
			)
		}

		return fmt.Sprintf(
			"\n%d guesses remaining\n%s\n\n%s%s",
			remaining,
			timerLine,
			footer,
			envSuffix,
		)
	}
}
