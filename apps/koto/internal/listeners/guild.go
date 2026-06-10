package listeners

import (
	"context"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/jurienhamaker/disgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

type GuildListener struct {
	client      *bot.Client
	settingsSvc *services.SettingsService
	pointsSvc   *services.PointsService
}

func GetGuildListener(container *di.Container) *GuildListener {
	return &GuildListener{
		client:      container.Get(sharedStatic.DiClient).(*disgoplus.Bot).Client(),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		pointsSvc:   container.Get(localStatic.DiPoints).(*services.PointsService),
	}
}

func AddGuildListeners(container *di.Container) {
	l := GetGuildListener(container)
	disgoBot := container.Get(sharedStatic.DiClient).(*disgoplus.Bot)
	disgoBot.Client().EventManager.AddEventListeners(
		bot.NewListenerFunc(l.OnGuildJoin),
		bot.NewListenerFunc(l.OnGuildLeave),
		bot.NewListenerFunc(l.OnGuildMemberLeave),
	)
}

func (l *GuildListener) OnGuildJoin(e *events.GuildJoin) {
	ctx := context.Background()

	guildSettings, err := l.settingsSvc.GetByGuildID(ctx, e.GuildID.String(), false)
	if err != nil {
		utils.Logger.Warnf("guild create: failed to retrieve settings for %s: %v", e.GuildID, err)
	}

	if guildSettings != nil {
		return
	}

	utils.Logger.Infof("Joined guild: %s", e.Guild.Name)

	if _, err := l.settingsSvc.GetByGuildID(ctx, e.GuildID.String(), true); err != nil {
		utils.Logger.Warnf("guild create: seed settings failed for %s: %v", e.GuildID, err)
	}

	go localUtils.SendWelcomeMessage(l.client, e.GuildID.String())
}

func (l *GuildListener) OnGuildLeave(e *events.GuildLeave) {
	utils.Logger.Infof("Left guild: %s", e.GuildID)
}

func (l *GuildListener) OnGuildMemberLeave(e *events.GuildMemberLeave) {
	go func() {
		if err := l.pointsSvc.RemovePlayerFromGuild(
			context.Background(),
			e.GuildID.String(),
			e.User.ID.String(),
		); err != nil {
			utils.Logger.Warnf("guild member remove: %s/%s: %v", e.GuildID, e.User.ID, err)
		}
	}()
}
