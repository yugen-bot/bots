package leaderboard

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/disgoplus"

	"jurien.dev/yugen/kusari/internal/ent"
	"jurien.dev/yugen/shared/utils"
)

func (m *LeaderboardModule) getItems(
	ctx *disgoplus.Ctx,
	page int,
) ([]any, int, error) {
	items, total, err := m.points.GetLeaderboardByGuildID(
		context.Background(),
		ctx.GuildID.String(),
		page,
	)

	return utils.UnpackArray(items), total, err
}

func (m *LeaderboardModule) formatItem(_ *disgoplus.Ctx, item any) string {
	parsed := item.(*ent.PlayerStats)
	return fmt.Sprintf("<@%s>: **%d**", parsed.UserID, parsed.Points)
}

func (m *LeaderboardModule) command(ctx *disgoplus.Ctx) {
	utils.LeaderboardCommandHandler(
		ctx,
		m.container,
		m.getItems,
		m.formatItem,
	)
}

func (m *LeaderboardModule) messageComponent(ctx *disgoplus.Ctx) {
	utils.LeaderboardMessageComponentHandler(
		ctx,
		m.container,
		m.getItems,
		m.formatItem,
	)
}
