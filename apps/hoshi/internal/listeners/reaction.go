package listeners

import (
	"context"
	"sync"
	"time"

	"jurien.dev/yugen/hoshi/internal/services"
	localStatic "jurien.dev/yugen/hoshi/internal/static"
	sharedStatic "jurien.dev/yugen/shared/static"

	"github.com/bwmarrin/discordgo"
	"github.com/jurienhamaker/discordgoplus"
	"github.com/sarulabs/di/v2"
)

type ReactionListener struct {
	bot    *discordgoplus.Bot
	svc    *services.StarboardService
	timers sync.Map
}

func GetReactionListener(container *di.Container) *ReactionListener {
	return &ReactionListener{
		bot: container.Get(sharedStatic.DiBot).(*discordgoplus.Bot),
		svc: container.Get(localStatic.DiStarboard).(*services.StarboardService),
	}
}

func AddReactionListeners(container *di.Container) {
	l := GetReactionListener(container)
	l.bot.AddHandler(l.OnMessageReactionAdd)
	l.bot.AddHandler(l.OnMessageReactionRemove)
}

func (l *ReactionListener) OnMessageReactionAdd(
	s *discordgo.Session,
	r *discordgo.MessageReactionAdd,
) {
	if r.UserID == s.State.User.ID {
		return
	}

	l.debounce(r.ChannelID, r.MessageID, r.GuildID, r.Emoji.ID, r.Emoji.Name)
}

func (l *ReactionListener) OnMessageReactionRemove(
	_ *discordgo.Session,
	r *discordgo.MessageReactionRemove,
) {
	l.debounce(r.ChannelID, r.MessageID, r.GuildID, r.Emoji.ID, r.Emoji.Name)
}

func (l *ReactionListener) debounce(
	channelID, messageID, guildID, emojiID, emojiName string,
) {
	if existing, ok := l.timers.LoadAndDelete(messageID); ok {
		existing.(*time.Timer).Stop()
	}

	t := time.AfterFunc(250*time.Millisecond, func() {
		l.timers.Delete(messageID)
		l.svc.CheckReaction(
			context.Background(),
			channelID,
			messageID,
			guildID,
			emojiID,
			emojiName,
		)
	})

	l.timers.Store(messageID, t)
}
