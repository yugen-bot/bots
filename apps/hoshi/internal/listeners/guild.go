package listeners

import (
	"context"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/config"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GuildListener struct {
	client      *bot.Client
	settingsSvc *services.SettingsService
	cfg         *config.Config
}

func GetGuildListener(container *di.Container) *GuildListener {
	return &GuildListener{
		client:      container.Get(sharedStatic.DiBot).(*disgoplus.Bot).Client(),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		cfg:         container.Get(sharedStatic.DiConfig).(*config.Config),
	}
}

func AddGuildListeners(container *di.Container) {
	l := GetGuildListener(container)
	disgoBot := container.Get(sharedStatic.DiBot).(*disgoplus.Bot)
	disgoBot.Client().EventManager.AddEventListeners(
		bot.NewListenerFunc(l.OnGuildJoin),
		bot.NewListenerFunc(l.OnGuildLeave),
	)
}

func (l *GuildListener) OnGuildJoin(e *events.GuildJoin) {
	ctx := context.Background()

	existing, err := l.settingsSvc.GetByGuildID(ctx, e.GuildID.String())
	if err != nil {
		return
	}

	if existing != nil {
		return
	}

	utils.Logger.Infof("Joined guild: %s", e.Guild.Name)

	for _, ch := range e.Guild.Channels {
		if ch.Type() != discord.ChannelTypeGuildText {
			continue
		}

		// Try sending; catch permission errors and continue to next channel.
		if err := localUtils.SendWelcomeMessage(ch, l.client, l.cfg.OwnerID); err == nil {
			break
		}
	}

	if _, err := l.settingsSvc.GetByGuildID(ctx, e.GuildID.String()); err != nil {
		utils.Logger.With("guildID", e.GuildID).
			Warn("Failed to seed settings on guild create")
	}
}

func (l *GuildListener) OnGuildLeave(e *events.GuildLeave) {
	utils.Logger.Infof("Left guild: %s", e.GuildID)
}
