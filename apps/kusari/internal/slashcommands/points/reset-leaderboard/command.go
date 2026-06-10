package resetleaderboard

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ResetLeaderboardModule) err(ctx *disgoplus.Ctx) {
	disgoplus.Respond(ctx, discord.MessageCreate{ //nolint:errcheck
		Content: "Something wen't wrong, try again later.",
		Flags:   discord.MessageFlagEphemeral,
	})
}

func (m *ResetLeaderboardModule) request(ctx *disgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	bot := m.container.Get(static.DiClient).(*disgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	guildSnowflake, parseErr := snowflake.Parse(ctx.GuildID.String())
	if parseErr != nil {
		m.err(ctx)
		return
	}

	guild, err := ctx.Client.Rest.GetGuild(guildSnowflake, false)
	if err != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: get guild failed",
			"error",
			err,
			"guildID",
			ctx.GuildID,
		)
		m.err(ctx)

		return
	}

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	userID := "none"
	confirmationTarget := guild.Name

	if u, ok := ctx.CommandData.OptUser("member"); ok {
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

	err = disgoplus.Respond(ctx, discord.MessageCreate{
		Embeds: []discord.Embed{embed},
		Components: []discord.LayoutComponent{
			discord.NewActionRow(
				discord.NewDangerButton(
					"Reset leaderboard",
					fmt.Sprintf("RESET_LEADERBOARD/true/%s", userID),
				),
				discord.NewSecondaryButton(
					"Cancel",
					fmt.Sprintf("RESET_LEADERBOARD/false/%s", userID),
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
			ctx.GuildID,
		)
	}
}

func (m *ResetLeaderboardModule) reset(ctx *disgoplus.Ctx) {
	doReset := ctx.MessageComponentOptions["reset"] == "true"

	if !doReset {
		contentText := "I have not reset the leaderboard"
		if ctx.MessageComponentOptions["userID"] != "none" {
			contentText = fmt.Sprintf(
				"%s for <@%s>",
				contentText,
				ctx.MessageComponentOptions["userID"],
			)
		}

		emptyEmbeds := []discord.Embed{}
		emptyComponents := []discord.LayoutComponent{}
		disgoplus.Update(ctx, discord.MessageUpdate{ //nolint:errcheck
			Content:    &contentText,
			Components: &emptyComponents,
			Embeds:     &emptyEmbeds,
		})

		return
	}

	contentText := "The leaderboard points have been reset"
	if ctx.MessageComponentOptions["userID"] != "none" {
		contentText = fmt.Sprintf(
			"%s for <@%s>",
			contentText,
			ctx.MessageComponentOptions["userID"],
		)
		go m.points.ResetLeaderboardByGuildIDAndUserID(
			context.Background(),
			ctx.GuildID.String(),
			ctx.MessageComponentOptions["userID"],
		)
	} else {
		go m.points.ResetLeaderboardByGuildID(
			context.Background(),
			ctx.GuildID.String(),
		)
	}

	emptyEmbeds := []discord.Embed{}
	emptyComponents := []discord.LayoutComponent{}
	disgoplus.Update(ctx, discord.MessageUpdate{ //nolint:errcheck
		Content:    &contentText,
		Components: &emptyComponents,
		Embeds:     &emptyEmbeds,
	})
}
