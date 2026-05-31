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
	sharedStatic "jurien.dev/yugen/shared/static"
)

type DonateHintModule struct {
	container   *di.Container
	settingsSvc *services.SettingsService
	hintsSvc    *services.HintsService
}

func GetDonateHintModule(container *di.Container) *DonateHintModule {
	return &DonateHintModule{
		container:   container,
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		hintsSvc:    container.Get(localStatic.DiHints).(*services.HintsService),
	}
}

func (m *DonateHintModule) donateHint(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	userID := ctx.Interaction.Member.User.ID

	player, err := m.hintsSvc.GetPlayerHintsByUserID(
		context.Background(),
		userID,
	)
	if err != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Sorry, couldn't retrieve your hint data.",
		}, true)
		return
	}

	settings, err := m.settingsSvc.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil || settings == nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Sorry, couldn't retrieve server settings.",
		}, true)
		return
	}

	if player.Hints < 1 {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"You don't have at least 1 hint to donate. You currently have **%s** hints.",
				strconv.FormatFloat(player.Hints, 'f', -1, 64),
			),
		}, true)
		return
	}

	if settings.Hints >= settings.MaxHints {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"The server already has **%s/%s** hints!",
				strconv.FormatFloat(settings.Hints, 'f', -1, 64),
				strconv.FormatFloat(settings.MaxHints, 'f', -1, 64),
			),
		}, true)
		return
	}

	go m.hintsSvc.DeductHintFromPlayer(context.Background(), userID, 1)

	hints, maxHints, err := m.hintsSvc.AddHintToGuild(
		context.Background(),
		ctx.Interaction.GuildID,
		settings,
		localStatic.DonationGuildValue,
	)
	if err != nil {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: "Sorry, failed to donate the hint.",
		}, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"**Hint donated!**\nThe server now has **%s/%s** hints!",
			strconv.FormatFloat(hints, 'f', -1, 64),
			strconv.FormatFloat(maxHints, 'f', -1, 64),
		),
	}, true)
}

func (m *DonateHintModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "donate-hint",
			Description: "Donate a personal hint to the server.",
			Handler:     discordgoplus.HandlerFunc(m.donateHint),
		},
	}
}
