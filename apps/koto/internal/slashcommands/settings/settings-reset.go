package settings

import (
	"context"
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	"jurien.dev/yugen/shared/static"
)

var settingsResetChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "Channel", Value: "channel"},
	{Name: "Role", Value: "role"},
	{Name: "Frequency", Value: "frequency"},
	{Name: "Time limit", Value: "time-limit"},
	{Name: "Cooldown", Value: "cooldown"},
	{Name: "Back-to-back cooldown", Value: "back-to-back-cooldown"},
	{Name: "Inform cooldown", Value: "inform-cooldown"},
	{Name: "Auto start", Value: "auto-start"},
	{Name: "Members privilege", Value: "members-privilege"},
	{Name: "Start after first guess", Value: "start-after-first-guess"},
	{Name: "Bot updates channel", Value: "bot-updates-channel"},
	{Name: "All settings", Value: "all"},
}

type SettingsResetModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSettingsResetModule(container *di.Container) *SettingsResetModule {
	return &SettingsResetModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SettingsResetModule) reset(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	setting := ctx.Options["setting"].StringValue()

	if _, err := m.settings.Reset(context.Background(), ctx.Interaction.GuildID, []string{setting}); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	idx := slices.IndexFunc(settingsResetChoices, func(c *discordgo.ApplicationCommandOptionChoice) bool {
		return c.Value == setting
	})

	name := setting
	if idx >= 0 {
		name = settingsResetChoices[idx].Name
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("**%s** has been reset to its default value.", name),
	}, true)
}

func (m *SettingsResetModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset a Koto setting to its default value",
			Handler:     discordgoplus.HandlerFunc(m.reset),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "setting",
					Description: "The setting to reset to its default value.",
					Required:    true,
					Choices:     settingsResetChoices,
				},
			},
		},
	}
}
