package channel

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/kusari/internal/ent"
)

func (m *ChannelModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	channel := ctx.Options["channel"].ChannelValue(ctx.Session)

	s, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	channelID := string(channel.ID)
	_, err = m.settings.Update(
		context.Background(),
		s.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetChannelID(channelID)
		},
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("I will run in <#%s> from now on.", channel.ID),
	}, true)
}
