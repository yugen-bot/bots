package server

import (
	"context"
	"fmt"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/zekroTJA/shinpuru/pkg/hammertime"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ServerModule) errResponse(e *handler.CommandEvent) error {
	_, followupErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Sorry couldn't retrieve the server information...",
		Flags:   discord.MessageFlagEphemeral,
	})
	if followupErr != nil {
		return fmt.Errorf(
			"server: create follow up message: %w",
			followupErr,
		)
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

	data, err := m.fetchServerData(e, guildID)
	if err != nil {
		return err
	}

	embed := m.buildServerEmbed(e, data)

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
	if err != nil {
		utils.Logger.Errorw(
			"server: follow up failed",
			"error", err, "guildID", guildID,
		)

		return fmt.Errorf("server: create follow up message: %w", err)
	}

	return nil
}

type serverData struct {
	settings      *ent.Settings
	gameExists    bool
	history       *ent.History
	historyExists bool
	guild         *discord.RestGuild
	guildID       string
	selfID        string
}

func (m *ServerModule) fetchServerData(
	e *handler.CommandEvent,
	guildID string,
) (*serverData, error) {
	settings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil {
		utils.Logger.Errorw(
			"server: get settings failed",
			"error", err, "guildID", guildID,
		)

		return nil, m.errResponse(e)
	}

	gameResult, gameExists, err := m.game.GetCurrentGame(
		context.Background(), guildID,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get current game failed",
			"error", err, "guildID", guildID,
		)

		return nil, m.errResponse(e)
	}

	history, historyExists, err := m.game.GetLastHistory(
		context.Background(), gameResult,
	)
	if err != nil {
		utils.Logger.Errorw(
			"server: get last history failed",
			"error", err, "guildID", guildID,
		)

		return nil, m.errResponse(e)
	}

	guildSnowflake, parseErr := snowflake.Parse(guildID)
	if parseErr != nil {
		return nil, m.errResponse(e)
	}

	restGuild, guildErr := e.Client().Rest.GetGuild(guildSnowflake, false)
	if guildErr != nil {
		utils.Logger.Errorw(
			"server: get guild failed",
			"error", guildErr, "guildID", guildID,
		)

		return nil, m.errResponse(e)
	}

	self, _ := e.Client().Caches.SelfUser()

	return &serverData{
		settings:      settings,
		gameExists:    gameExists,
		history:       history,
		historyExists: historyExists,
		guild:         restGuild,
		guildID:       guildID,
		selfID:        self.ID.String(),
	}, nil
}

func (m *ServerModule) buildServerEmbed(
	_ *handler.CommandEvent,
	d *serverData,
) discord.Embed {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	onGoingGameText := "None"
	if d.gameExists && d.settings.ChannelID != nil {
		onGoingGameText = fmt.Sprintf("at <#%s>", *d.settings.ChannelID)
	}

	highscoreDateText := ""
	if d.settings.HighscoreDate != nil {
		highscoreDateText = " - " + hammertime.Format(
			*d.settings.HighscoreDate, hammertime.Span,
		)
	}

	lastWordText := "-"
	if d.historyExists && d.history.UserID != d.selfID {
		lastWordText = fmt.Sprintf("<@%s>", d.history.UserID)
	}

	iconURL := ""
	if url := d.guild.IconURL(); url != nil {
		iconURL = *url
	}

	return discord.NewEmbed().
		WithColor(embedColor).
		WithTitle(d.guild.Name).
		WithThumbnail(iconURL).
		WithDescription(buildServerDescription(d, onGoingGameText, highscoreDateText, lastWordText)).
		WithEmbedFooter(footer)
}

func buildServerDescription(
	d *serverData,
	onGoingGameText string,
	highscoreDateText string,
	lastWordText string,
) string {
	lastWord := "-"
	if d.history != nil {
		lastWord = d.history.Word
	}

	return fmt.Sprintf(
		"Ongoing game: **%s**\n"+
			"High score: **%d%s**\n"+
			"Last word: **%s**\n"+
			"Last word by: **%s**\n\n"+
			"Guild saves: **%s/%s**\n"+
			"Saves used: **%s**",
		onGoingGameText,
		d.settings.Highscore,
		highscoreDateText,
		lastWord,
		lastWordText,
		strconv.FormatFloat(d.settings.Saves, 'f', -1, 64),
		strconv.FormatFloat(d.settings.MaxSaves, 'f', -1, 64),
		strconv.FormatFloat(d.settings.SavesUsed, 'f', -1, 64),
	)
}
