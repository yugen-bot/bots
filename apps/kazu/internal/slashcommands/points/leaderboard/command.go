package leaderboard

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/discordgoplus"

	"jurien.dev/yugen/shared/utils"

	"jurien.dev/yugen/kazu/internal/ent"
)

func (m *LeaderboardModule) getItems(
	ctx *discordgoplus.Ctx,
	page int,
) ([]any, int, error) {
	items, total, err := m.points.GetLeaderboardByGuildID(
		context.Background(),
		ctx.Interaction.GuildID,
		page,
	)

	return utils.UnpackArray(items), total, err
}

func (m *LeaderboardModule) formatItem(_ *discordgoplus.Ctx, item any) string {
	parsed := item.(*ent.PlayerStats)
	return fmt.Sprintf("<@%s>: **%d**", parsed.UserID, parsed.Points)
}

func (m *LeaderboardModule) command(ctx *discordgoplus.Ctx) {
	utils.LeaderboardCommandHandler(
		ctx,
		m.container,
		m.getItems,
		m.formatItem,
	)
}

func (m *LeaderboardModule) messageComponent(ctx *discordgoplus.Ctx) {
	utils.LeaderboardMessageComponentHandler(
		ctx,
		m.container,
		m.getItems,
		m.formatItem,
	)
}
