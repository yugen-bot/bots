package leaderboard

import (
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *LeaderboardModule) formatItem(item any) string {
	parsed := item.(*ent.PlayerStats)
	return fmt.Sprintf("<@%s>: **%d**", parsed.UserID, parsed.Points)
}

func (m *LeaderboardModule) command(
	data discord.SlashCommandInteractionData,
	e *handler.CommandEvent,
) error {
	return utils.LeaderboardCommandHandler(
		data,
		e,
		m.container,
		m.getItems,
		m.formatItem,
	)
}

func (m *LeaderboardModule) messageComponent(e *handler.ComponentEvent) error {
	return utils.LeaderboardComponentHandler(
		e,
		m.container,
		m.getItems,
		m.formatItem,
	)
}
