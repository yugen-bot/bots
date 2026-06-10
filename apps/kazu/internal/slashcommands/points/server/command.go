package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"

	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ServerModule) errFollowup(e *handler.CommandEvent) error {
	if _, err := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Sorry couldn't retrieve the server information...",
		Flags:   discord.MessageFlagEphemeral,
	}); err != nil {
		return fmt.Errorf("server: create followup message: %w", err)
	}

	return nil
}

func (m *ServerModule) server(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("server: defer create message: %w", err)
	}

	guildID := e.GuildID().String()

	embed, buildErr := m.buildServerEmbed(e, guildID)
	if buildErr != nil {
		return buildErr
	}

	if _, err := e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	}); err != nil {
		utils.Logger.Errorw(
			"server: follow up failed",
			"error", err,
			"guildID", guildID,
		)

		return fmt.Errorf("server: create followup message: %w", err)
	}

	return nil
}

func (m *ServerModule) buildServerEmbed(
	e *handler.CommandEvent,
	guildID string,
) (discord.Embed, error) {
	settings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil {
		utils.Logger.Errorw(
			"server: get settings failed",
			"error", err,
			"guildID", guildID,
		)

		return discord.Embed{}, m.errFollowup(e)
	}

	game, gameExists, err := m.game.GetCurrentGame(
		context.Background(),
		guildID,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get current game failed",
			"error", err,
			"guildID", guildID,
		)

		return discord.Embed{}, m.errFollowup(e)
	}

	history, historyExists, err := m.game.GetLastHistory(
		context.Background(),
		game,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get last history failed",
			"error", err,
			"guildID", guildID,
		)

		return discord.Embed{}, m.errFollowup(e)
	}

	guild, err := e.Client().Rest.GetGuild(*e.GuildID(), false)
	if err != nil {
		utils.Logger.Errorw(
			"server: get guild failed",
			"error", err,
			"guildID", guildID,
		)

		return discord.Embed{}, m.errFollowup(e)
	}

	self, _ := e.Client().Caches.SelfUser()

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	desc := m.buildServerDescription(
		settings,
		history,
		gameExists,
		historyExists,
		self.ID.String(),
	)

	iconURL := ""
	if url := guild.IconURL(); url != nil {
		iconURL = *url
	}

	return discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(guild.Name).
		WithThumbnail(iconURL).
		WithDescription(desc).
		WithEmbedFooter(footer), nil
}

func (m *ServerModule) buildServerDescription(
	settings *ent.Settings,
	history *ent.History,
	gameExists bool,
	historyExists bool,
	selfID string,
) string {
	onGoingGameText := "None"

	channelID := settings.ChannelID
	if gameExists && channelID != nil {
		onGoingGameText = fmt.Sprintf("at <#%s>", *channelID)
	}

	highscoreDateText := ""
	if highscoreDate := settings.HighscoreDate; highscoreDate != nil {
		highscoreDateText = " - " + hammertime.Format(
			*highscoreDate,
			hammertime.Span,
		)
	}

	lastNumber := 0
	lastCountedText := "-"

	if historyExists && history != nil {
		lastNumber = history.Number
		if history.UserID != selfID {
			lastCountedText = fmt.Sprintf("<@%s>", history.UserID)
		}
	}

	lastShamedText := "\n"

	if shameRoleID := settings.ShameRoleID; shameRoleID != nil {
		lastShameUserID := settings.LastShameUserID

		userText := "-"
		if lastShameUserID != nil {
			userText = fmt.Sprintf("<@%s>", *lastShameUserID)
		}

		lastShamedText = fmt.Sprintf("Last shamed user: **%s**\n", userText)
	}

	return fmt.Sprintf(
		`Ongoing game: **%s**
High score: **%d%s**
Last number: **%d**
Last count by: **%s**
%s
Guild saves: **%s/%s**
Saves used: **%s**
			`,
		onGoingGameText,
		settings.Highscore,
		highscoreDateText,
		lastNumber,
		lastCountedText,
		lastShamedText,
		strconv.FormatFloat(settings.Saves, 'f', -1, 64),
		strconv.FormatFloat(settings.MaxSaves, 'f', -1, 64),
		strconv.FormatFloat(settings.SavesUsed, 'f', -1, 64),
	)
}
