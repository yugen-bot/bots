package channel

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *ChannelModule) set(ctx *discordgoplus.Ctx) {
	discordgoplus.Defer(ctx, true)

	ch := ctx.Options["channel"].ChannelValue(ctx.Session)

	settings, err := m.settings.GetByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	_, err = m.settings.Update(
		context.Background(),
		settings.ID,
		func(u *ent.SettingsUpdateOne) { u.SetChannelID(ch.ID) },
	)
	if err != nil {
		discordgoplus.ErrorResponse(ctx, true)
		return
	}

	discordgoplus.FollowUp(ctx, &discordgo.WebhookParams{
		Content: fmt.Sprintf("I will run in <#%s> from now on.", ch.ID),
	}, true)
}
