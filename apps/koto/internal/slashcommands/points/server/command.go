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

	localStatic "jurien.dev/yugen/koto/internal/static"
	"jurien.dev/yugen/shared/config"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ServerModule) server(
	_ discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	guildID := (*e.GuildID()).String()
	bg := context.Background()

	settings, err := m.settingsSvc.GetByGuildID(bg, guildID)
	if err != nil || settings == nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry, couldn't retrieve the server information.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	guildSnowflake, err := snowflake.Parse(guildID)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry, couldn't retrieve the server information.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	guild, err := m.bot.Client().Rest.GetGuild(guildSnowflake, false)
	if err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Sorry, couldn't retrieve the server information.",
			Flags:   discord.MessageFlagEphemeral,
		})

		return err
	}

	currentGame, _ := m.gameSvc.GetCurrentGame(bg, guildID)
	lastSolved, _ := m.gameSvc.GetLastSolvedGame(bg, guildID)

	var gameLines []string

	if currentGame != nil {
		if settings.ChannelID != nil && *settings.ChannelID != "" {
			gameLines = append(
				gameLines,
				fmt.Sprintf("Ongoing game: **at <#%s>**", *settings.ChannelID),
			)
		}
	} else {
		nextStart, _ := m.gameSvc.GetNextGameStart(bg, guildID, settings)
		if nextStart == nil || !nextStart.After(time.Now()) {
			gameLines = append(gameLines, "Next game: **starting soon**")
		} else {
			gameLines = append(
				gameLines,
				fmt.Sprintf("Next game: <t:%d:R>", nextStart.Unix()),
			)
		}
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

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Flags:  discord.MessageFlagEphemeral,
	})

	return err
}
