package points

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type PointsModule struct {
	container   *di.Container
	pointsSvc   *services.PointsService
	hintsSvc    *services.HintsService
	settingsSvc *services.SettingsService
}

func GetPointsModule(container *di.Container) *PointsModule {
	return &PointsModule{
		container:   container,
		pointsSvc:   container.Get(localStatic.DiPoints).(*services.PointsService),
		hintsSvc:    container.Get(localStatic.DiHints).(*services.HintsService),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
	}
}

func (m *PointsModule) points(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	settings, err := m.settingsSvc.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil || settings == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	channelID, ok := settings.ChannelID()
	if !ok || channelID == "" {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	player, err := m.pointsSvc.GetPlayer(
		context.Background(),
		ctx.Interaction.GuildID,
		ctx.Interaction.Member.User.ID,
		false,
	)
	if err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	playerHints, _ := m.hintsSvc.GetPlayerHintsByUserID(
		context.Background(),
		ctx.Interaction.Member.User.ID,
	)

	bot := m.container.Get(sharedStatic.DiBot).(*discordgoplus.Bot)
	footer := utils.CreateEmbedFooter(
		bot,
		&utils.CreateEmbedFooterParams{IsVote: false},
		"",
	)

	hintsValue := "0/3"
	if playerHints != nil {
		hintsValue = fmt.Sprintf(
			"%s/%s",
			strconv.FormatFloat(playerHints.Hints, 'f', -1, 64),
			strconv.FormatFloat(playerHints.MaxHints, 'f', -1, 64),
		)
	}

	embed := &discordgo.MessageEmbed{
		Title: "Your Koto Stats",
		Color: localStatic.EmbedColor,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Points",
				Value:  fmt.Sprintf("%d", player.Points),
				Inline: true,
			},
			{
				Name:   "Games Participated",
				Value:  fmt.Sprintf("%d", player.Participated),
				Inline: true,
			},
			{
				Name:   "Games Won",
				Value:  fmt.Sprintf("%d", player.Wins),
				Inline: true,
			},
			{
				Name:   "Hints",
				Value:  hintsValue,
				Inline: true,
			},
		},
		Footer: footer,
	}

	discordgoplus.FollowUp(
		ctx,
		&discordgo.WebhookParams{Embeds: []*discordgo.MessageEmbed{embed}},
		true,
	)
}

func (m *PointsModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "points",
			Description: "View your Koto points",
			Handler:     discordgoplus.HandlerFunc(m.points),
		},
	}
}
