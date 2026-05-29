package settings

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/ent"
	"jurien.dev/yugen/koto/internal/services"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/static"
)

type SetChannelModule struct {
	container *di.Container
	settings  *services.SettingsService
}

func GetSetChannelModule(container *di.Container) *SetChannelModule {
	return &SetChannelModule{
		container: container,
		settings:  container.Get(static.DiSettings).(*services.SettingsService),
	}
}

func (m *SetChannelModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	guildID := ctx.Interaction.GuildID
	channel := ctx.Options["channel"].ChannelValue(ctx.Session)

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		localUtils.ReplyNoSettings(ctx)
		return
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetChannelID(channel.ID) },
	); err != nil {
		discordgoplus.InteractionError(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf(
			"Koto will listen for guesses in <#%s>!",
			channel.ID,
		),
	}, true)
}

func (m *SetChannelModule) Commands() []*discordgoplus.Command {
	return []*discordgoplus.Command{
		{
			Name:        "channel",
			Description: "Set the channel where Koto listens for guesses",
			Handler:     discordgoplus.HandlerFunc(m.set),
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionChannel,
					Name:        "channel",
					Description: "The channel to listen in.",
					Required:    true,
				},
			},
		},
	}
}
