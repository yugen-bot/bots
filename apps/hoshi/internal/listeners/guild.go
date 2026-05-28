package listeners

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
	"jurien.dev/yugen/hoshi/internal/services"
	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/config"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"
)

func AddGuildListeners(container *di.Container) {
	bot := container.Get(sharedStatic.DiBot).(*discordgoplus.Bot)
	settingsSvc := container.Get(sharedStatic.DiSettings).(*services.SettingsService)
	cfg := container.Get(sharedStatic.DiConfig).(*config.Config)

	bot.AddHandler(func(s *discordgo.Session, event *discordgo.GuildCreate) {
		ctx := context.Background()

		settings, err := settingsSvc.GetByGuildID(ctx, event.ID)
		if err != nil {
			return
		}

		if settings != nil {
			return
		}

		utils.Logger.Infof("Joined guild: %s", event.Name)

		for _, ch := range event.Channels {
			if ch.Type != discordgo.ChannelTypeGuildText {
				continue
			}

			perms, err := s.UserChannelPermissions(s.State.User.ID, ch.ID)
			if err != nil {
				continue
			}

			if perms&discordgo.PermissionSendMessages != 0 {
				localUtils.SendWelcomeMessage(ch, bot, cfg.OwnerID)
				break
			}
		}

		if _, err := settingsSvc.GetByGuildID(ctx, event.ID); err != nil {
			utils.Logger.With("guildID", event.ID).
				Warn("Failed to seed settings on guild create")
		}
	})

	bot.AddHandler(func(s *discordgo.Session, event *discordgo.GuildDelete) {
		utils.Logger.Infof("Left guild: %s", event.ID)
	})
}
