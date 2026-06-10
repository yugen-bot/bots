package setbacktobackcooldown

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetBackToBackCooldownModule) set(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	guildID := (*e.GuildID()).String()
	enable := data.Bool("enabled")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetEnableBackToBackCooldown(enable)
			if v, ok := data.OptInt("seconds"); ok {
				u.SetBackToBackCooldown(v)
			}
		},
	); err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	if enable {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Back-to-back cooldown has been **enabled**!",
			Flags:   discord.MessageFlagEphemeral,
		})
	} else {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Back-to-back cooldown has been **disabled**!",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	return err
}
