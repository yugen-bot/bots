// Package leaderboard provides the leaderboard slash command for kazu.
package leaderboard

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/handler"
	"github.com/disgoorg/snowflake/v2"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/kazu/internal/ent"
	"jurien.dev/yugen/kazu/internal/services"
	local "jurien.dev/yugen/kazu/internal/static"
	"jurien.dev/yugen/shared/utils"
)

// LeaderboardModule provides the /leaderboard command and its pagination component.
type LeaderboardModule struct {
	container *di.Container
	points    *services.PointsService
}

// GetLeaderboardModule constructs a LeaderboardModule from the DI container.
func GetLeaderboardModule(container *di.Container) *LeaderboardModule {
	return &LeaderboardModule{
		container: container,
		points:    container.Get(local.DiPoints).(*services.PointsService),
	}
}

// Commands returns the leaderboard command definition.
func (m *LeaderboardModule) Commands() []disgoplus.CommandRegistration {
	return utils.GetLeaderboardCommands()
}

// Register wires the leaderboard command and component onto the router.
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
