package listeners

import (
	"context"

	"jurien.dev/yugen/hoshi/internal/services"
	localUtils "jurien.dev/yugen/hoshi/internal/utils"
	"jurien.dev/yugen/shared/config"
	sharedStatic "jurien.dev/yugen/shared/static"
	"jurien.dev/yugen/shared/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
)

type GuildListener struct {
	bot         *discordgoplus.Bot
	settingsSvc *services.SettingsService
	cfg         *config.Config
}

func GetGuildListener(container *di.Container) *GuildListener {
	return &GuildListener{
		bot:         container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
		settingsSvc: container.Get(sharedStatic.DiSettings).(*services.SettingsService),
		cfg:         container.Get(sharedStatic.DiConfig).(*config.Config),
	}
}

func AddGuildListeners(container *di.Container) {
	l := GetGuildListener(container)
	l.bot.AddHandler(l.OnGuildCreate)
	l.bot.AddHandler(l.OnGuildDelete)
}

func (l *GuildListener) OnGuildCreate(
	s *discordgo.Session,
	event *discordgo.GuildCreate,
) {
	ctx := context.Background()

	existing, err := l.settingsSvc.GetByGuildID(ctx, event.ID)
	if err != nil {
		return
	}

	if existing != nil {
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
			localUtils.SendWelcomeMessage(ch, l.bot, l.cfg.OwnerID)
			break
		}
	}

	if _, err := l.settingsSvc.GetByGuildID(ctx, event.ID); err != nil {
		utils.Logger.With("guildID", event.ID).
			Warn("Failed to seed settings on guild create")
	}
}

func (l *GuildListener) OnGuildDelete(
	_ *discordgo.Session,
	event *discordgo.GuildDelete,
) {
	utils.Logger.Infof("Left guild: %s", event.ID)
}
