package server

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"

	"jurien.dev/yugen/koto/internal/ent"
	localStatic "jurien.dev/yugen/koto/internal/static"
	"jurien.dev/yugen/shared/config"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ServerModule) buildGameLines(
	bg context.Context,
	guildID string,
	settings *ent.Settings,
) []string {
	var gameLines []string

	currentGame, _ := m.gameSvc.GetCurrentGame(bg, guildID)
	lastSolved, _ := m.gameSvc.GetLastSolvedGame(bg, guildID)

	if currentGame == nil {
		nextStart, _ := m.gameSvc.GetNextGameStart(bg, guildID, settings)
		if nextStart == nil || !nextStart.After(time.Now()) {
			gameLines = append(gameLines, "Next game: **starting soon**")
		} else {
			gameLines = append(
				gameLines,
				fmt.Sprintf("Next game: <t:%d:R>", nextStart.Unix()),
			)
		}
	} else if settings.ChannelID != nil && *settings.ChannelID != "" {
		gameLines = append(
			gameLines,
			fmt.Sprintf("Ongoing game: **at <#%s>**", *settings.ChannelID),
		)
	}

	if lastSolved != nil {
		gameLines = append(gameLines,
			fmt.Sprintf("Last solved word: **%s**", lastSolved.Game.Word),
		)

		if lastSolved.Solver != "" {
			gameLines = append(gameLines,
				fmt.Sprintf("Last solver: <@%s>", lastSolved.Solver),
			)
		}
	}

	return gameLines
}

func sendServerError(e *handler.CommandEvent) error {
	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Sorry, couldn't retrieve the server information.",
		Flags:   discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("server: send followup: %w", sendErr)
	}

	return nil
}

func (m *ServerModule) server(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("server: defer: %w", err)
	}

	guildID := e.GuildID().String()
	bg := context.Background()

	settings, err := m.settingsSvc.GetByGuildID(bg, guildID)
	if err != nil || settings == nil {
		return sendServerError(e)
	}

	guildSnowflake, err := snowflake.Parse(guildID)
	if err != nil {
		return sendServerError(e)
	}

	guild, err := m.bot.Client().Rest.GetGuild(guildSnowflake, false)
	if err != nil {
		return sendServerError(e)
	}

	gameLines := m.buildGameLines(bg, guildID, settings)

	hintLines := fmt.Sprintf(
		"Guild hints: **%s/%s**\nHints used: **%s**",
		strconv.FormatFloat(settings.Hints, 'f', -1, 64),
		strconv.FormatFloat(settings.MaxHints, 'f', -1, 64),
		strconv.FormatFloat(settings.HintsUsed, 'f', -1, 64),
	)

	description := hintLines
	if len(gameLines) > 0 {
		description = strings.Join(gameLines, "\n") + "\n\n" + hintLines
	}

	cfg := m.container.Get(sharedStatic.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	embed := discord.NewEmbed().
		WithColor(localStatic.EmbedColor).
		WithTitle(guild.Name).
		WithDescription(description).
		WithEmbedFooter(footer)

	if iconURL := guild.IconURL(); iconURL != nil {
		embed = embed.WithThumbnail(*iconURL)
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("server: send followup: %w", sendErr)
	}

	return nil
}
