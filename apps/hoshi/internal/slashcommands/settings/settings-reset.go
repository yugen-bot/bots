package slashcommands

import (
	"context"
	"fmt"
	"slices"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/lib/pq"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/ent"
	"jurien.dev/yugen/hoshi/internal/services"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
)

var resetChoices = []*discordgo.ApplicationCommandOptionChoice{
	{Name: "Treshold", Value: "treshold"},
	{Name: "Author starring", Value: "self"},
	{Name: "Ignored channels", Value: "ignoredChannelIds"},
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

	var (
		apply func(*ent.SettingsUpdateOne)
		value string
	)

	switch setting {
	case "treshold":
		apply = func(u *ent.SettingsUpdateOne) { u.SetTreshold(3) }
		value = "3"
	case "self":
		apply = func(u *ent.SettingsUpdateOne) { u.SetSelf(false) }
		value = "false"
	case "ignoredChannelIds":
		apply = func(u *ent.SettingsUpdateOne) { u.SetIgnoredChannelIds(pq.StringArray{}) }
		value = "[]"
	default:
		discordgoplus.InteractionError(ctx, true)
		return
	}

	if err := m.settings.Set(context.Background(), ctx.Interaction.GuildID, apply); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	idx := slices.IndexFunc(
		resetChoices,
		func(c *discordgo.ApplicationCommandOptionChoice) bool {
			return c.Value == setting
		},
	)
	name := resetChoices[idx].Name

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"%s has been reset to its default value of `%s`",
			name,
			value,
		),
	}, true)
}

func (m *SettingsResetModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "reset",
			Description: "Reset a Hoshi setting to its default value.",
			Middlewares: []discordgoplus.Handler{
				discordgoplus.HandlerFunc(middlewares.GuildAdminMiddleware),
			},
			Handler: discordgoplus.HandlerFunc(m.reset),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "setting",
					Description: "The setting to reset to its default value.",
					Required:    true,
					Choices:     resetChoices,
				},
			},
		},
	}
}
