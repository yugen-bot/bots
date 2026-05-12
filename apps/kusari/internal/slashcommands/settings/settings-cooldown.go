package slashcommands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/kusari/internal/services"
	"jurien.dev/yugen/kusari/prisma/db"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type SettingsCooldownModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsCooldownModule(container *di.Container) *SettingsCooldownModule {
	return &SettingsCooldownModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsCooldownModule) set(ctx *discordgoplus.Ctx) {
	utils.Logger.With("Options", ctx.Options, "GuildID", ctx.Interaction.GuildID).Debug("Cooldown command used")
	discordgoplus.Defer(ctx, true)

	seconds := ctx.Options["seconds"].IntValue()

	settings, err := m.settings.GetByGuildId(ctx.Interaction.GuildID)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		settings.ID,
		db.Settings.Cooldown.Set(int(seconds)),
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	secondsText := "seconds"
	if seconds == 1 {
		secondsText = "second"
	}

	content := fmt.Sprintf("Members will now be able to provide a word every %d %s.", seconds, secondsText)
	if seconds == 0 {
		content = "Cooldown has been removed!"
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: content,
	}, true)
}

func (m *SettingsCooldownModule) Commands() []*discordgoplus.Command {
	minValue := 0.0
	maxValue := 3600.0

	return []*discordgoplus.Command{
		{
			Name:        "cooldown",
			Description: "Set the cooldown between answers.",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionInteger,
					Name:        "seconds",
					Description: "The amount of seconds between answers.",
					Required:    true,
					MinValue:    &minValue,
					MaxValue:    maxValue,
				},
			},
		},
	}
}
