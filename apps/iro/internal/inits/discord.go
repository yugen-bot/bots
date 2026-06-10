package inits

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/iro/internal/listeners"
	sharedListeners "jurien.dev/yugen/shared/listeners"
	"jurien.dev/yugen/shared/middlewares"
	"jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func InitDiscordBot(ctx context.Context, container *di.Container) error {
	disgoBot := container.Get(static.DiBot).(*disgoplus.Bot)

	disgoBot.Client().EventManager.AddEventListeners(
		bot.NewListenerFunc(func(e *events.Ready) {
			self, _ := e.Client().Caches.SelfUser()
			utils.Logger.Infof(
				"Logged in as: %v (shard %d)",
				self.Username,
				e.ShardID(),
			)
		}),
	)

	middlewares.InitMiddlewares(container)

	sharedListeners.AddLogListeners(container)
	sharedListeners.AddMetricsListeners(container)
	sharedListeners.AddVoteListeners(container)

	listeners.AddColorListeners(container)

	if err := InitCommands(container); err != nil {
		return fmt.Errorf("discord: init commands: %w", err)
	}

	if err := disgoBot.Open(ctx); err != nil {
		return fmt.Errorf("discord: open: %w", err)
	}

	return nil
}
