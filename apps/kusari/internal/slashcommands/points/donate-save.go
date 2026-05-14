package slashcommands

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/static"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
)

type DonateSaveModule struct {
	container *di.Container
	settings  *services.SettingsService
	saves     *services.SavesService
}

func GetDonateSaveModule(container *di.Container) *DonateSaveModule {
	return &DonateSaveModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
		saves:     container.Get(local.DiSaves).(*services.SavesService),
	}
}

func (m *DonateSaveModule) donateSave(ctx *discordgoplus.Ctx) {
	err := discordgoplus.Defer(ctx, true)
	if err != nil {
		return
	}

	player, err := m.saves.GetPlayerSavesByUserID(
		context.Background(),
		ctx.Interaction.Member.User.ID,
	)
	if err != nil {
		return
	}

	settings, err := m.settings.GetByGuildId(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		return
	}

	if player.Saves < 1 {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"You currently don't have atleast 1 save to donate, you currently have **%d** saves!",
				int(player.Saves),
			),
		}, true)
		return
	}

	if settings.Saves >= settings.MaxSaves {
		discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
			Content: fmt.Sprintf(
				"The server already has **%s/%s** saves!",
				strconv.FormatFloat(settings.Saves, 'f', -1, 64),
				strconv.FormatFloat(settings.MaxSaves, 'f', -1, 64),
			),
		}, true)
		return
	}

	go m.saves.DeductSaveFromPlayer(
		context.Background(),
		ctx.Interaction.Member.User.ID,
		1,
	)
	saves, maxSaves, err := m.saves.AddSaveToGuild(
		context.Background(),
		settings.GuildID,
		settings,
		0.2,
	)
	if err != nil {
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			`**Save donated!**
The server now has **%s/%s** saves!`,
			strconv.FormatFloat(saves, 'f', -1, 64),
			strconv.FormatFloat(maxSaves, 'f', -1, 64),
		),
	}, true)
}

func (m *DonateSaveModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "donate-save",
			Description: "Donate a personal save to the server.",
			Handler:     discordgoplus.HandlerFunc(m.donateSave),
		},
	}
}
