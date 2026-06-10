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

	guildSnowflake, parseErr := snowflake.Parse(e.GuildID().String())
	if parseErr != nil {
		if err := e.CreateMessage(discord.MessageCreate{
			Content: "Something wen't wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); err != nil {
			return fmt.Errorf("reset leaderboard: create message: %w", err)
		}

		return nil
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

		if createErr := e.CreateMessage(discord.MessageCreate{
			Content: "Something wen't wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		}); createErr != nil {
			return fmt.Errorf(
				"reset leaderboard: create message: %w",
				createErr,
			)
		}

		return nil
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

	if createErr := e.CreateMessage(discord.MessageCreate{
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
	}); createErr != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: respond failed",
			"error",
			createErr,
			"guildID",
			e.GuildID(),
		)

		return fmt.Errorf("reset leaderboard: create message: %w", createErr)
	}

	return nil
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

		if err := e.UpdateMessage(discord.MessageUpdate{
			Content:    &contentText,
			Components: &emptyComponents,
			Embeds:     &emptyEmbeds,
		}); err != nil {
			return fmt.Errorf("reset leaderboard: update message: %w", err)
		}

		return nil
	}

	contentText := "The leaderboard points have been reset"
	if e.Vars["userID"] != "none" {
		contentText = fmt.Sprintf(
			"%s for <@%s>",
			contentText,
			e.Vars["userID"],
		)

		go func() {
			if resetErr := m.points.ResetLeaderboardByGuildIDAndUserID(
				context.Background(),
				e.GuildID().String(),
				e.Vars["userID"],
			); resetErr != nil {
				utils.Logger.Errorw(
					"reset leaderboard: reset by guild and user failed",
					"error", resetErr,
				)
			}
		}()
	} else {
		go func() {
			if resetErr := m.points.ResetLeaderboardByGuildID(
				context.Background(),
				e.GuildID().String(),
			); resetErr != nil {
				utils.Logger.Errorw(
					"reset leaderboard: reset by guild failed",
					"error", resetErr,
				)
			}
		}()
	}

	emptyEmbeds := []discord.Embed{}
	emptyComponents := []discord.LayoutComponent{}

	if err := e.UpdateMessage(discord.MessageUpdate{
		Content:    &contentText,
		Components: &emptyComponents,
		Embeds:     &emptyEmbeds,
	}); err != nil {
		return fmt.Errorf("reset leaderboard: update message: %w", err)
	}

	return nil
}
