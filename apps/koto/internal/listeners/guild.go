package listeners

import (
	"context"

	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
)

type GuildListener struct {
	bot         *discordgoplus.Bot
	settingsSvc *services.SettingsService
	pointsSvc   *services.PointsService
}

func GetGuildListener(container *di.Container) *GuildListener {
	return &GuildListener{
		bot:         container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		pointsSvc:   container.Get(localStatic.DiPoints).(*services.PointsService),
	}
}

func AddGuildListeners(container *di.Container) {
	l := GetGuildListener(container)
	l.bot.AddHandler(l.OnGuildCreate)
	l.bot.AddHandler(l.OnGuildDelete)
	l.bot.AddHandler(l.OnGuildMemberRemove)
}

func (l *GuildListener) OnGuildCreate(
	_ *discordgo.Session,
	event *discordgo.GuildCreate,
) {
	ctx := context.Background()

	guildSettings, err := l.settingsSvc.GetByGuildID(ctx, event.ID, false)
	if err != nil {
		utils.Logger.Warnf(
			"guild create: failed to retrieve settings for %s: %v",
			event.ID,
			err,
		)
	}

	if guildSettings != nil {
		return
	}

	utils.Logger.Infof("Joined guild: %s", event.Name)

	if _, err := l.settingsSvc.GetByGuildID(ctx, event.ID, true); err != nil {
		utils.Logger.Warnf(
			"guild create: seed settings failed for %s: %v",
			event.ID,
			err,
		)
	}

	go localUtils.SendWelcomeMessage(l.bot, event.ID)
}

func (l *GuildListener) OnGuildDelete(
	_ *discordgo.Session,
	event *discordgo.GuildDelete,
) {
	utils.Logger.Infof("Left guild: %s", event.ID)
}

func (l *GuildListener) OnGuildMemberRemove(
	_ *discordgo.Session,
	event *discordgo.GuildMemberRemove,
) {
	go func() {
		if err := l.pointsSvc.RemovePlayerFromGuild(
			context.Background(),
			event.GuildID,
			event.User.ID,
		); err != nil {
			utils.Logger.Warnf(
				"guild member remove: %s/%s: %v",
				event.GuildID,
				event.User.ID,
				err,
			)
		}
	}()
}
