package listeners

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"
)

type reactionDebouncer struct {
	timers sync.Map
	svc    *services.StarboardService
	bot    *discordgoplus.Bot
}

func (d *reactionDebouncer) onReaction(channelID, messageID, guildID, emojiID, emojiName string) {
	if existing, ok := d.timers.LoadAndDelete(messageID); ok {
		existing.(*time.Timer).Stop()
	}

	t := time.AfterFunc(500*time.Millisecond, func() {
		d.timers.Delete(messageID)
		d.svc.CheckReaction(channelID, messageID, guildID, emojiID, emojiName)
	})
	d.timers.Store(messageID, t)
}

func AddReactionListeners(container *di.Container) {
	bot := container.Get(sharedStatic.DiBot).(*discordgoplus.Bot)
	starboardSvc := container.Get(localStatic.DiStarboard).(*services.StarboardService)

	d := &reactionDebouncer{svc: starboardSvc, bot: bot}

	bot.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionAdd) {
		if r.UserID == s.State.User.ID {
			return
		}

		d.onReaction(r.ChannelID, r.MessageID, r.GuildID, r.Emoji.ID, r.Emoji.Name)
	})

	bot.AddHandler(func(s *discordgo.Session, r *discordgo.MessageReactionRemove) {
		d.onReaction(r.ChannelID, r.MessageID, r.GuildID, r.Emoji.ID, r.Emoji.Name)
	})
}
