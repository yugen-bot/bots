package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ServerModule) err(ctx *discordgoplus.Ctx) {
	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: "Sorry couldn't retrieve the server information...",
	}, true)
}

func (m *ServerModule) server(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get settings failed",
			"error",
			err,
			"guildID",
			ctx.Interaction.GuildID,
		)
		m.err(ctx)

		return
	}

	game, gameExists, err := m.game.GetCurrentGame(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get current game failed",
			"error",
			err,
			"guildID",
			ctx.Interaction.GuildID,
		)
		m.err(ctx)

		return
	}

	history, historyExists, err := m.game.GetLastHistory(
		context.Background(),
		game,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get last history failed",
			"error",
			err,
			"guildID",
			ctx.Interaction.GuildID,
		)
		m.err(ctx)

		return
	}

	guild, err := m.bot.Guild(ctx.Interaction.GuildID)
	if err != nil {
		utils.Logger.Errorw(
			"server: get guild failed",
			"error",
			err,
			"guildID",
			ctx.Interaction.GuildID,
		)
		m.err(ctx)

		return
	}

	self := ctx.State.User

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	onGoingGameText := "None"

	channelID := settings.ChannelID
	if gameExists && channelID != nil {
		onGoingGameText = fmt.Sprintf("at <#%s>", *channelID)
	}

	highscoreDateText := ""

	highscoreDate := settings.HighscoreDate
	if highscoreDate != nil {
		highscoreDateText = " - " + hammertime.Format(
			*highscoreDate,
			hammertime.Span,
		)
	}

	lastNumber := 0
	lastCountedText := "-"
	if historyExists && history != nil {
		lastNumber = history.Number
		if history.UserID != self.ID {
			lastCountedText = fmt.Sprintf("<@%s>", history.UserID)
		}
	}

	lastShamedText := "\n"

	shameRoleID := settings.ShameRoleID
	if shameRoleID != nil {
		lastShameUserID := settings.LastShameUserID

		userText := "-"
		if lastShameUserID != nil {
			userText = fmt.Sprintf("<@%s>", *lastShameUserID)
		}

		lastShamedText = fmt.Sprintf("Last shamed user: **%s**\n", userText)
	}

	embed := &discordgo.MessageEmbed{
		Color: embedColor,
		Title: guild.Name,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: guild.IconURL("64"),
		},
		Description: fmt.Sprintf(
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
		),
		Footer: footer,
	}

	err = discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}, true)
	if err != nil {
		utils.Logger.Errorw(
			"server: follow up failed",
			"error",
			err,
			"guildID",
			ctx.Interaction.GuildID,
		)
	}
}
