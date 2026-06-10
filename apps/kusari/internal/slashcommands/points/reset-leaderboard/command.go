package resetleaderboard

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ResetLeaderboardModule) request(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiBot).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	guildSnowflake, parseErr := snowflake.Parse((*e.GuildID()).String())
	if parseErr != nil {
		return e.CreateMessage(discord.MessageCreate{
			Content: "Something wen't wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	guild, err := e.Client().Rest.GetGuild(guildSnowflake, false)
	if err != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: get guild failed",
			"error",
			err,
			"guildID",
			e.GuildID(),
		)

		return e.CreateMessage(discord.MessageCreate{
			Content: "Something wen't wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	userID := "none"
	confirmationTarget := guild.Name

	if u, ok := data.OptUser("member"); ok {
		userID = u.ID.String()
		confirmationTarget = fmt.Sprintf("<@%s>", userID)
	}

	embed := discord.NewEmbed().
		WithColor(embedColor).
		WithTitle("Reset leaderboard").
		WithDescription(fmt.Sprintf(
			`Are you sure you want to reset the leaderboard of **%s**
**This action is irreversible**`,
			confirmationTarget,
		)).
		WithEmbedFooter(footer)

	err = e.CreateMessage(discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Components: []discord.LayoutComponent{
			discord.NewActionRow(
				discord.NewDangerButton(
					"Reset leaderboard",
					fmt.Sprintf(customIDResetLeaderboardTrue, userID),
				),
				discord.NewSecondaryButton(
					"Cancel",
					fmt.Sprintf(customIDResetLeaderboardFalse, userID),
				),
			),
		},
	})
	if err != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: respond failed",
			"error",
			err,
			"guildID",
			e.GuildID(),
		)
	}

	return err
}

func (m *ResetLeaderboardModule) reset(e *handler.ComponentEvent) error {
	doReset := e.Vars["reset"] == "true"

	if !doReset {
		contentText := "I have not reset the leaderboard"
		if e.Vars["userID"] != "none" {
			contentText = fmt.Sprintf(
				"%s for <@%s>",
				contentText,
				e.Vars["userID"],
			)
		}

		emptyEmbeds := []discord.Embed{}
		emptyComponents := []discord.LayoutComponent{}

		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &contentText,
			Components: &emptyComponents,
			Embeds:     &emptyEmbeds,
		})
	}

	contentText := "The leaderboard points have been reset"
	if e.Vars["userID"] != "none" {
		contentText = fmt.Sprintf(
			"%s for <@%s>",
			contentText,
			e.Vars["userID"],
		)
		go m.points.ResetLeaderboardByGuildIDAndUserID(
			context.Background(),
			(*e.GuildID()).String(),
			e.Vars["userID"],
		)
	} else {
		go m.points.ResetLeaderboardByGuildID(
			context.Background(),
			(*e.GuildID()).String(),
		)
	}

	emptyEmbeds := []discord.Embed{}
	emptyComponents := []discord.LayoutComponent{}

	return e.UpdateMessage(discord.MessageUpdate{
		Content:    &contentText,
		Components: &emptyComponents,
		Embeds:     &emptyEmbeds,
	})
}
