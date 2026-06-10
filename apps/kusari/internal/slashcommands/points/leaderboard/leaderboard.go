// Package leaderboard contains the kusari /leaderboard slash command.
package leaderboard

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kusari/internal/services"
	local "jurien.dev/yugen/kusari/internal/static"
	"jurien.dev/yugen/shared/utils"
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

func (m *LeaderboardModule) Commands() []discord.ApplicationCommandCreate {
	return utils.GetLeaderboardCommands()
}

func (m *LeaderboardModule) Register(r handler.Router) {
	r.SlashCommand("/leaderboard", m.command)
	r.Component("/LEADERBOARD/{page}", m.messageComponent)
}

func (m *LeaderboardModule) getItems(
	guildID snowflake.ID,
	page int,
) ([]any, int, error) {
	items, total, err := m.points.GetLeaderboardByGuildID(
		context.Background(),
		guildID.String(),
		page,
	)

	return utils.UnpackArray(items), total, err
}
