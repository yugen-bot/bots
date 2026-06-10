package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ServerModule) err(ctx *disgoplus.Ctx) {
	disgoplus.FollowUp(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: "Sorry couldn't retrieve the server information...",
		Flags:   discord.MessageFlagEphemeral,
	})
}

func (m *ServerModule) server(ctx *disgoplus.Ctx) {
	disgoplus.Defer(ctx, true) //nolint:errcheck

	guildID := ctx.GuildID.String()

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		guildID,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get settings failed",
			"error",
			err,
			"guildID",
			guildID,
		)
		m.err(ctx)

		return
	}

	gameResult, gameExists, err := m.game.GetCurrentGame(
		context.Background(),
		guildID,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get current game failed",
			"error",
			err,
			"guildID",
			guildID,
		)
		m.err(ctx)

		return
	}

	history, historyExists, err := m.game.GetLastHistory(
		context.Background(),
		gameResult,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get last history failed",
			"error",
			err,
			"guildID",
			guildID,
		)
		m.err(ctx)

		return
	}

	guildSnowflake, parseErr := snowflake.Parse(guildID)
	if parseErr != nil {
		m.err(ctx)
		return
	}

	guild, err := ctx.Client.Rest.GetGuild(guildSnowflake, false)
	if err != nil {
		utils.Logger.Errorw(
			"server: get guild failed",
			"error",
			err,
			"guildID",
			guildID,
		)
		m.err(ctx)

		return
	}

	self, _ := ctx.Client.Caches.SelfUser()

	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	onGoingGameText := "None"

	if gameExists && settings.ChannelID != nil {
		onGoingGameText = fmt.Sprintf("at <#%s>", *settings.ChannelID)
	}

	highscoreDateText := ""

	if settings.HighscoreDate != nil {
		highscoreDateText = " - " + hammertime.Format(
			*settings.HighscoreDate,
			hammertime.Span,
		)
	}

	lastWordText := "-"
	if historyExists && history.UserID != self.ID.String() {
		lastWordText = fmt.Sprintf("<@%s>", history.UserID)
	}

	lastWord := "-"
	if history != nil {
		lastWord = history.Word
	}

	iconURL := ""
	if url := guild.IconURL(); url != nil {
		iconURL = *url
	}

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(guild.Name).
		WithThumbnail(iconURL).
		WithDescription(fmt.Sprintf(
			`Ongoing game: **%s**
High score: **%d%s**
Last word: **%s**
Last word by: **%s**

Guild saves: **%s/%s**
Saves used: **%s**
				`,
			onGoingGameText,
			settings.Highscore,
			highscoreDateText,
			lastWord,
			lastWordText,
			strconv.FormatFloat(settings.Saves, 'f', -1, 64),
			strconv.FormatFloat(settings.MaxSaves, 'f', -1, 64),
			strconv.FormatFloat(settings.SavesUsed, 'f', -1, 64),
		)).
		WithEmbedFooter(footer)

	_, err = disgoplus.FollowUp(ctx, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
	if err != nil {
		utils.Logger.Errorw(
			"server: follow up failed",
			"error",
			err,
			"guildID",
			guildID,
		)
	}
}
