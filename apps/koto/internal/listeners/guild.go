package listeners

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/koto/internal/services"
	localStatic "jurien.dev/yugen/koto/internal/static"
	localUtils "jurien.dev/yugen/koto/internal/utils"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func AddGuildListeners(container *di.Container) {
	bot := container.Get(sharedStatic.DiBot).(*discordgoplus.Bot)
	settingsSvc := container.Get(sharedStatic.DiSettings).(*services.SettingsService)
	pointsSvc := container.Get(localStatic.DiPoints).(*services.PointsService)

	bot.AddHandler(func(s *discordgo.Session, event *discordgo.GuildCreate) {
		ctx := context.Background()

		settings, err := settingsSvc.GetByGuildID(
			ctx,
			event.ID,
			false,
		)
		if err != nil {
			utils.Logger.Warnf(
				"guild create: failed to retrieve settings for %s: %v",
				event.ID,
				err,
			)
		}

		if settings != nil {
			return
		}

		utils.Logger.Infof("Joined guild: %s", event.Name)

		if _, err := settingsSvc.GetByGuildID(
			ctx,
			event.ID,
			true,
		); err != nil {
			utils.Logger.Warnf(
				"guild create: seed settings failed for %s: %v",
				event.ID,
				err,
			)
		}

		go localUtils.SendWelcomeMessage(bot, event.ID)
	})

	bot.AddHandler(func(s *discordgo.Session, event *discordgo.GuildDelete) {
		utils.Logger.Infof("Left guild: %s", event.ID)

		if err := settingsSvc.Delete(
			context.Background(),
			event.ID,
		); err != nil {
			utils.Logger.Warnf(
				"guild delete: cleanup failed for %s: %v",
				event.ID,
				err,
			)
		}
	})

	bot.AddHandler(
		func(s *discordgo.Session, event *discordgo.GuildMemberRemove) {
			go func() {
				if err := pointsSvc.RemovePlayerFromGuild(
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
		},
	)
}
