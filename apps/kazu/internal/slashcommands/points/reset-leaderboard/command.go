package resetleaderboard

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/shared/config"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func (m *ResetLeaderboardModule) request(ctx *discordgoplus.Ctx) {
	cfg := m.container.Get(static.DiConfig).(*config.Config)
	footer := utils.CreateEmbedFooter(
		m.container.Get(static.DiBot).(*discordgoplus.Bot),
		&utils.CreateEmbedFooterParams{
			IsVote: false,
		},
		cfg.OwnerID,
	)

	guild, err := m.bot.Guild(ctx.Interaction.GuildID)
	if err != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: get guild failed",
			"error",
			err,
			"guildID",
			ctx.Interaction.GuildID,
		)
		m.errResponse(ctx)

		return
	}

	embedColor := m.container.Get(static.DiEmbedColor).(int)

	memberOption := ctx.Options["member"]
	userID := "none"
	confirmationTarget := guild.Name

	if memberOption != nil {
		userID = memberOption.Value.(string)
		confirmationTarget = fmt.Sprintf("<@%s>", userID)
	}

	embed := &discordgo.MessageEmbed{
		Color: embedColor,
		Title: "Reset leaderboard",
		Description: fmt.Sprintf(
			`Are you sure you want to reset the leaderboard of **%s**
**This action is irreversible**`,
			confirmationTarget,
		),
		Footer: footer,
	}

	err = discordgoplus.Respond(ctx, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						CustomID: fmt.Sprintf(
							"RESET_LEADERBOARD/true/%s",
							userID,
						),
						Style: discordgo.DangerButton,
						Label: "Reset leaderboard",
					},
					discordgo.Button{
						CustomID: fmt.Sprintf(
							"RESET_LEADERBOARD/false/%s",
							userID,
						),
						Style: discordgo.SecondaryButton,
						Label: "Cancel",
					},
				},
			},
		},
	}, true)
	if err != nil {
		utils.Logger.Errorw(
			"reset-leaderboard: respond failed",
			"error",
			err,
			"guildID",
			ctx.Interaction.GuildID,
		)
	}
}

func (m *ResetLeaderboardModule) reset(ctx *discordgoplus.Ctx) {
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

		discordgoplus.Update(ctx, &discordgo.InteractionResponseData{
			Content:    contentText,
			Components: []discordgo.MessageComponent{},
			Embeds:     []*discordgo.MessageEmbed{},
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
			ctx.Interaction.GuildID,
			ctx.MessageComponentOptions["userID"],
		)
	} else {
		go m.points.ResetLeaderboardByGuildID(
			context.Background(),
			ctx.Interaction.GuildID,
		)
	}

	discordgoplus.Update(ctx, &discordgo.InteractionResponseData{
		Content:    contentText,
		Components: []discordgo.MessageComponent{},
		Embeds:     []*discordgo.MessageEmbed{},
	})
}
