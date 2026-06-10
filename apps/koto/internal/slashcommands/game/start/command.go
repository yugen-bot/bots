package start

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	localUtils "jurien.dev/yugen/koto/internal/utils"
	"jurien.dev/yugen/shared/utils"
)

func (m *StartModule) start(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	if err := e.DeferCreateMessage(true); err != nil {
		return fmt.Errorf("game start: defer: %w", err)
	}

	guildID := e.GuildID().String()

	guildSettings, err := m.settings.GetByGuildID(context.Background(), guildID)
	if err != nil || guildSettings == nil {
		return localUtils.ReplyNoSettings(e, true)
	}

	if guildSettings.ChannelID == nil || *guildSettings.ChannelID == "" {
		return localUtils.ReplyNoSettings(e, true)
	}

	isModerator := e.Member() != nil &&
		e.Member().Permissions.Has(discord.PermissionManageGuild)

	if !guildSettings.MembersCanStart && !isModerator {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "Only moderators can start games unless members privilege is enabled.",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("game start: send followup: %w", sendErr)
		}

		return nil
	}

	started, err := m.game.Start(context.Background(), guildID, true, false, "")
	if err != nil {
		utils.Logger.Warnw("game: start: start failed: %w", err)

		return localUtils.HandleChannelInaccessible(
			e,
			*guildSettings.ChannelID,
			err,
		)
	}

	if !started {
		_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
			Content: "There is already an active game!",
			Flags:   discord.MessageFlagEphemeral,
		})
		if sendErr != nil {
			return fmt.Errorf("game start: send followup: %w", sendErr)
		}

		return nil
	}

	_, sendErr := e.CreateFollowupMessage(discord.MessageCreate{
		Content: "Game started!",
		Flags:   discord.MessageFlagEphemeral,
	})
	if sendErr != nil {
		return fmt.Errorf("game start: send followup: %w", sendErr)
	}

	return nil
}
