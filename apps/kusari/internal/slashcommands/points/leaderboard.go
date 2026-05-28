package points

import (
	"context"
	"fmt"

	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/shared/utils"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/kusari/prisma/db"
)

type LeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
}

func GetLeaderboardModule(container *di.Container) *LeaderboardModule {
	return &LeaderboardModule{
		container: container,
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

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
	parsed := item.(db.PlayerStatsModel)
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

func (m *LeaderboardModule) Commands() []*discordgoplus.Command {
	return utils.GetLeaderboardCommands(discordgoplus.HandlerFunc(m.command))
}

func (m *LeaderboardModule) MessageComponents() []*discordgoplus.MessageComponent {
	return utils.GetLeaderboardMessageComponents(
		discordgoplus.HandlerFunc(m.messageComponent),
	)
}
