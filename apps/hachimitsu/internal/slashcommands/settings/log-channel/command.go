package logchannel

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/hachimitsu/internal/ent"
)

func (m *LogChannelModule) set(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("settings log-channel: defer: %w", err)
	}

	guildID := e.GuildID().String()
	channel := data.Channel("channel")

	existing, err := m.settings.GetByGuildID(
		context.Background(), guildID, true,
	)
	if err != nil || existing == nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("settings log-channel: send followup: %w", fErr)
		}

		return nil
	}

	if _, updateErr := m.settings.Update(
		context.Background(),
		existing.ID,
		func(u *ent.SettingsUpdateOne) {
			u.SetLogChannelID(channel.ID.String())
		},
	); updateErr != nil {
		_, fErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong, try again later.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if fErr != nil {
			return fmt.Errorf("settings log-channel: send followup: %w", fErr)
		}

		return nil
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: fmt.Sprintf(
			"Ban logs will now be posted in <#%s>!",
			channel.ID.String(),
		),
		Flags: discord.MessageFlagEphemeral,
	})
	if err != nil {
		return fmt.Errorf("settings log-channel: send followup: %w", err)
	}

	return nil
}
