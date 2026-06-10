package authorstarring

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/hoshi/internal/ent"
)

func (m *AuthorStarringModule) set(data discord.SlashCommandInteractionData, e *handler.CommandEvent) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return err
	}

	allowed := data.Bool("allowed")

	err := m.settings.Set(
		context.Background(),
		(*e.GuildID()).String(),
		func(u *ent.SettingsUpdateOne) { u.SetSelf(allowed) },
	)
	if err != nil {
		_, ferr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Something went wrong.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if ferr != nil {
			return ferr
		}
		return err
	}

	state := "disallowed"
	if allowed {
		state = "allowed"
	}

	_, err = e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Message authors are now **" + state + "** to star their own message.",
		Flags:   discord.MessageFlagEphemeral,
	})
	return err
}
