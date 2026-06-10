package setautostart

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/koto/internal/ent"
	localUtils "jurien.dev/yugen/koto/internal/utils"
)

func (m *SetAutoStartModule) set(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	guildID := (*e.GuildID()).String()
	enabled := data.Bool("value")

	existing, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || existing == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	if _, err := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) { u.SetAutoStart(enabled) },
	); err != nil {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		return err
	}

	if enabled {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Koto will now automatically start a new game after one ends!",
			Flags:   discord.MessageFlagEphemeral,
		})
	} else {
		_, err = e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Koto will no longer automatically start a new game after one ends.",
			Flags:   discord.MessageFlagEphemeral,
		})
	}
	return err
}
