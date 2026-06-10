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

func (m *ResetLeaderboardModule) request(ctx *disgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*disgoplus.Bot),
		&utils.CreateEmbedFooterParams{IsVote: false},
		cfg.OwnerID,
	)

	guild, err := ctx.Client.Rest.GetGuild(snowflake.MustParse(ctx.GuildID.String()), false)
	if err != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: get guild failed",
			"error", err,
			"guildID", ctx.GuildID.String(),
		)
		m.errResponse(ctx)
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
				discord.NewDangerButton("Reset leaderboard", fmt.Sprintf("RESET_LEADERBOARD/true/%s", userID)),
				discord.NewSecondaryButton("Cancel", fmt.Sprintf("RESET_LEADERBOARD/false/%s", userID)),
			),
		},
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: respond failed",
			"error", err,
			"guildID", ctx.GuildID.String(),
		)
	}
}

func (m *ResetLeaderboardModule) reset(ctx *disgoplus.Ctx) {
	reset := ctx.MessageComponentOptions["reset"] == "true"

	if !reset {
		contentText := "I have not reset the leaderboard"
		if ctx.MessageComponentOptions["userID"] != "none" {
			contentText = fmt.Sprintf(
				"%s for <@%s>",
				contentText,
				ctx.MessageComponentOptions["userID"],
			)
		}

		empty := []discord.LayoutComponent{}
		emptyEmbeds := []discord.Embed{}
		disgoplus.Update(ctx, discord.MessageUpdate{ //nolint:errcheck
			Content:    &contentText,
			Components: &empty,
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
		go m.points.ResetLeaderboardByGuildIDAndUserID( //nolint:errcheck
			context.Background(),
			ctx.GuildID.String(),
			ctx.MessageComponentOptions["userID"],
		)
	} else {
		go m.points.ResetLeaderboardByGuildID( //nolint:errcheck
			context.Background(),
			ctx.GuildID.String(),
		)
	}

	empty := []discord.LayoutComponent{}
	emptyEmbeds := []discord.Embed{}
	disgoplus.Update(ctx, discord.MessageUpdate{ //nolint:errcheck
		Content:    &contentText,
		Components: &empty,
		Embeds:     &emptyEmbeds,
	})
}
