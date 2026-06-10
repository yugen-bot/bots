package resetleaderboard

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ResetLeaderboardModule) request(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	guild, err := e.Client().Rest.GetGuild(*e.GuildID(), false)
	if err != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: get guild failed",
			"error", err,
			"guildID", (*e.GuildID()).String(),
		)

		return e.CreateMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
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

	return e.CreateMessage(discord.MessageCreate{
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
		Flags: discord.MessageFlagEphemeral,
	})
}

func (m *ResetLeaderboardModule) reset(e *handler.ComponentEvent) error {
	reset := e.Vars["reset"] == "true"
	userID := e.Vars["userID"]

	if !reset {
		contentText := "I have not reset the leaderboard"
		if userID != "none" {
			contentText = fmt.Sprintf(
				"%s for <@%s>",
				contentText,
				userID,
			)
		}

		empty := []discord.LayoutComponent{}
		emptyEmbeds := []discord.Embed{}

		return e.UpdateMessage(discord.MessageUpdate{
			Content:    &contentText,
			Components: &empty,
			Embeds:     &emptyEmbeds,
		})
	}

	contentText := "The leaderboard points have been reset"
	if userID != "none" {
		contentText = fmt.Sprintf(
			"%s for <@%s>",
			contentText,
			userID,
		)
		go m.points.ResetLeaderboardByGuildIDAndUserID( //nolint:errcheck
			context.Background(),
			(*e.GuildID()).String(),
			userID,
		)
	} else {
		go m.points.ResetLeaderboardByGuildID( //nolint:errcheck
			context.Background(),
			(*e.GuildID()).String(),
		)
	}

	empty := []discord.LayoutComponent{}
	emptyEmbeds := []discord.Embed{}

	return e.UpdateMessage(discord.MessageUpdate{
		Content:    &contentText,
		Components: &empty,
		Embeds:     &emptyEmbeds,
	})
}
